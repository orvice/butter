package application

// Workspace onboarding and safe detachment for repository-owned Agent Content
// (issue #219). Onboarding reconciles a freshly bound workspace's existing
// database-managed content with the repository — either by exporting the DB
// content to Git (EXPORT_CURRENT) or adopting the repository's content for
// matching Agent IDs (IMPORT_REPOSITORY). Detachment (DeleteWorkspaceRepoBinding,
// in repobinding_service.go) is the inverse: it materializes the Active Revision
// back into database-managed fields before the binding is removed.

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"

	"butterfly.orx.me/core/log"
	"go.orx.me/apps/butter/internal/agentcontent"
	"go.orx.me/apps/butter/internal/gitprovider"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// OnboardWorkspaceRepository performs the one-time reconciliation of a freshly
// bound workspace's Agent Content with the repository. EXPORT_CURRENT commits
// existing database-managed content; IMPORT_REPOSITORY adopts repository content
// for matching Agent IDs. In both modes the workspace only switches to
// Git-owned content — `active_commit_sha` advances — once the commit/import,
// read-back, content validation, and Active Revision publication all succeed; a
// failure leaves the workspace on its previous database-owned state.
func (s *RepoBindingServiceServer) OnboardWorkspaceRepository(ctx context.Context, req *connect.Request[agentsv1.OnboardWorkspaceRepositoryRequest]) (*connect.Response[agentsv1.OnboardWorkspaceRepositoryResponse], error) {
	if err := s.requireRepos(); err != nil {
		return nil, err
	}
	ws, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.requireBindingRole(ctx, ws); err != nil {
		return nil, err
	}
	binding, err := s.repo.Get(ctx, ws)
	if err != nil {
		return nil, mapRepoBindingErr(err)
	}

	switch req.Msg.GetMode() {
	case agentsv1.RepoBindingOnboardingMode_REPO_BINDING_ONBOARDING_MODE_EXPORT_CURRENT:
		return s.onboardExport(ctx, ws, binding)
	case agentsv1.RepoBindingOnboardingMode_REPO_BINDING_ONBOARDING_MODE_IMPORT_REPOSITORY:
		return s.onboardImport(ctx, ws, binding)
	default:
		return nil, connectx.InvalidArgument("mode", "must be EXPORT_CURRENT or IMPORT_REPOSITORY")
	}
}

// onboardExport writes the workspace's database-managed Agent Content to the
// repository as a single validated commit, then reads it back, validates, and
// publishes the Active Revision. The changeset is a full snapshot of the DB, not
// a merge: for each managed field it emits a PUT when the DB value is non-empty
// and a DELETE when the value is empty but a stale managed file exists remotely,
// so after the export the repository reflects the database exactly. It always
// commits directly (DIRECT_COMMIT) even for CHANGE_REQUEST bindings — onboarding
// must publish an Active Revision to switch to Git-owned content, and a PR/MR
// would leave content unpublished behind an open review (mirrors the lifecycle
// Saga, which forces DIRECT_COMMIT for the same reason, ADR-0006).
func (s *RepoBindingServiceServer) onboardExport(ctx context.Context, ws string, binding *agentsv1.WorkspaceRepoBinding) (*connect.Response[agentsv1.OnboardWorkspaceRepositoryResponse], error) {
	if s.agentRepo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("agent repository not configured"))
	}
	agents, err := s.agentRepo.ListAgents(ctx, ws)
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	existing, err := s.existingManagedContentFiles(ctx, ws, binding, agents)
	if err != nil {
		return nil, err
	}
	actions, exported := exportContentActions(agents, existing)

	logger := log.FromContext(ctx)
	if len(actions) == 0 {
		// Nothing in the database to export and nothing stale to clear. Do not
		// silently adopt repository state — the caller chose EXPORT, not IMPORT
		// — so leave the workspace database-owned.
		logger.Info("workspace repository export: nothing to export", "workspace_id", ws)
		return connect.NewResponse(&agentsv1.OnboardWorkspaceRepositoryResponse{
			Binding:   binding,
			Published: binding.GetActiveCommitSha() != "",
		}), nil
	}

	res, err := s.commitContent(ctx, ws, actions,
		agentsv1.RepoBindingWriteMode_REPO_BINDING_WRITE_MODE_DIRECT_COMMIT,
		"export", "Export Agent Content from Butter", "")
	if err != nil {
		return nil, err
	}
	if len(res.validationErrors) > 0 {
		// Content invalid: no commit was created and the Active Revision is
		// unchanged, so the workspace stays database-owned.
		return connect.NewResponse(&agentsv1.OnboardWorkspaceRepositoryResponse{
			Binding:          res.binding,
			ValidationErrors: res.validationErrors,
		}), nil
	}
	if res.publishErr != nil {
		// The commit landed but read-back/validation/publication or the runner
		// reload failed. The workspace has NOT switched to Git-owned content;
		// surface the failure so the operator retries rather than believing
		// onboarding succeeded.
		return nil, connectx.InternalWith(res.publishErr)
	}

	published := res.commitSHA != "" && res.binding.GetActiveCommitSha() == res.commitSHA
	logger.Info("workspace repository onboarded via export", "workspace_id", ws,
		"commit_sha", res.commitSHA, "agents_exported", exported, "published", published)
	return connect.NewResponse(&agentsv1.OnboardWorkspaceRepositoryResponse{
		Binding:        res.binding,
		CommitSha:      res.commitSHA,
		Published:      published,
		AgentsExported: int32(exported),
	}), nil
}

// onboardImport syncs the repository and publishes the Active Revision. The
// publication pipeline parses content only for known Agent IDs, so directories
// that do not match an existing workspace Agent stay unclaimed and no
// bidirectional merge is performed. `published` is derived from whether the
// Active Revision actually advanced to the observed revision — a repository with
// no matching content advances nothing and is reported as not published.
func (s *RepoBindingServiceServer) onboardImport(ctx context.Context, ws string, binding *agentsv1.WorkspaceRepoBinding) (*connect.Response[agentsv1.OnboardWorkspaceRepositoryResponse], error) {
	syncResp, err := s.SyncWorkspaceRepository(ctx, connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{}))
	if err != nil {
		return nil, err
	}
	final := syncResp.Msg.GetBinding()
	if final == nil {
		final = binding
	}
	published := final.GetActiveCommitSha() != "" && final.GetActiveCommitSha() == final.GetObservedCommitSha()
	imported := 0
	if published && s.contentRepo != nil {
		if snap, snapErr := s.contentRepo.GetSnapshot(ctx, ws); snapErr == nil {
			imported = len(snap.Entries)
		}
	}
	log.FromContext(ctx).Info("workspace repository onboarded via import", "workspace_id", ws,
		"published", published, "agents_imported", imported)
	return connect.NewResponse(&agentsv1.OnboardWorkspaceRepositoryResponse{
		Binding:          final,
		Published:        published,
		ValidationErrors: syncResp.Msg.GetPublicationErrors(),
		AgentsImported:   int32(imported),
	}), nil
}

// exportContentActions builds a full-snapshot changeset from the database-managed
// content of every non-deleted agent with an assigned Agent ID. For each managed
// file it emits a PUT when the DB value is non-empty, or a DELETE when the value
// is empty and a stale managed file exists remotely (from `existing`), so the
// repository ends up reflecting the database rather than merging with it.
// Returns the actions and the number of agents that exported at least one value.
func exportContentActions(agents []*agentsv1.Agent, existing map[string]bool) ([]*agentsv1.ContentFileAction, int) {
	var actions []*agentsv1.ContentFileAction
	exported := 0
	for _, a := range agents {
		id := a.GetAgentId()
		if id == "" {
			continue
		}
		if a.GetLifecycleStatus() == agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_DELETED {
			continue
		}
		wrote := false
		put := func(path, content string) {
			content = strings.TrimSpace(content)
			switch {
			case content != "":
				actions = append(actions, &agentsv1.ContentFileAction{
					Path:      path,
					Operation: agentsv1.ContentFileOperation_CONTENT_FILE_OPERATION_PUT,
					Content:   content,
				})
				wrote = true
			case existing[path]:
				// Empty DB value but a stale managed file exists remotely: clear
				// it so Git reflects the (now empty) database field.
				actions = append(actions, &agentsv1.ContentFileAction{
					Path:      path,
					Operation: agentsv1.ContentFileOperation_CONTENT_FILE_OPERATION_DELETE,
				})
			}
		}
		put(agentcontent.DescriptionPath(id), a.GetDescription())
		put(agentcontent.PromptPath(id), a.GetConfig().GetInstruction())
		put(agentcontent.GlobalPromptPath(id), a.GetConfig().GetGlobalInstruction())
		if wrote {
			exported++
		}
	}
	return actions, exported
}

// existingManagedContentFiles returns the set of managed content file paths
// (root_path-relative, e.g. "agents/{id}/prompt.md") that currently exist at the
// branch HEAD for agents known to this workspace. It is used so an export can
// delete stale managed files whose database field is now empty. A missing or
// unreadable managed tree (e.g. a fresh repository) yields an empty set rather
// than an error, so a first export is not blocked by an empty repository.
func (s *RepoBindingServiceServer) existingManagedContentFiles(ctx context.Context, ws string, binding *agentsv1.WorkspaceRepoBinding, agents []*agentsv1.Agent) (map[string]bool, error) {
	managed := make(map[string]struct{})
	for _, a := range agents {
		id := a.GetAgentId()
		if id == "" || a.GetLifecycleStatus() == agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_DELETED {
			continue
		}
		for _, p := range agentcontent.ManagedPaths(id) {
			managed[p] = struct{}{}
		}
	}
	out := make(map[string]bool)
	if len(managed) == 0 {
		return out, nil
	}
	client, err := s.resolveProviderClient(ctx, ws, binding)
	if err != nil {
		return nil, err
	}
	headSHA, err := client.GetBranchHead(ctx, binding.GetBranch())
	if err != nil {
		return out, nil
	}
	entries, err := client.GetTree(ctx, headSHA, binding.GetRootPath())
	if err != nil {
		return out, nil
	}
	rootPrefix := strings.TrimRight(binding.GetRootPath(), "/")
	if rootPrefix != "" {
		rootPrefix += "/"
	}
	for _, te := range entries {
		if te.Kind != gitprovider.TreeEntryFile {
			continue
		}
		p := te.Path
		if rootPrefix != "" {
			p = strings.TrimPrefix(p, rootPrefix)
		}
		if _, ok := managed[p]; ok {
			out[p] = true
		}
	}
	return out, nil
}

// materializeActiveContent writes each entry of the active content snapshot back
// onto its database-managed Agent fields (description / instruction /
// global_instruction). It is the inverse of the publication overlay and runs
// during safe detachment so runtime behavior is preserved once Git is no longer
// the source of truth. Returns the number of agents updated.
func (s *RepoBindingServiceServer) materializeActiveContent(ctx context.Context, ws string, snapshot agentcontent.Snapshot) (int, error) {
	if s.agentRepo == nil {
		return 0, errors.New("agent repository not configured")
	}
	count := 0
	for agentID, content := range snapshot.Entries {
		agent, err := s.agentRepo.GetAgentByID(ctx, ws, agentID)
		if err != nil {
			if errors.Is(err, configrepo.ErrNotFound) {
				// The snapshot references an agent that no longer exists; there
				// is nothing to materialize onto.
				continue
			}
			return count, err
		}
		agent.Description = content.Description
		if agent.Config == nil {
			agent.Config = &agentsv1.AgentConfig{}
		}
		agent.Config.Instruction = content.Instruction
		agent.Config.GlobalInstruction = content.GlobalInstruction
		if _, err := s.agentRepo.UpdateAgent(ctx, ws, agent); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}
