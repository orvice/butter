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
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// OnboardWorkspaceRepository performs the one-time reconciliation of a freshly
// bound workspace's Agent Content with the repository. EXPORT_CURRENT commits
// existing database-managed content; IMPORT_REPOSITORY adopts repository content
// for matching Agent IDs. In both modes the workspace only switches to
// Git-owned content once the commit/import, read-back, validation, Effective
// Agent construction, and Active Revision publication all succeed — the shared
// publication pipeline advances active_commit_sha only on success, so a failed
// validation leaves the workspace on its previous (database-owned) state.
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

// onboardExport writes the workspace's existing database-managed Agent Content
// to the repository as a single validated commit, then reads it back and
// publishes the Active Revision. It always commits directly (DIRECT_COMMIT) so
// onboarding is synchronous and can publish immediately; a CHANGE_REQUEST mode
// binding would otherwise leave content unpublished behind an open PR/MR and the
// workspace never switching to Git-owned content.
func (s *RepoBindingServiceServer) onboardExport(ctx context.Context, ws string, binding *agentsv1.WorkspaceRepoBinding) (*connect.Response[agentsv1.OnboardWorkspaceRepositoryResponse], error) {
	if s.agentRepo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("agent repository not configured"))
	}
	agents, err := s.agentRepo.ListAgents(ctx, ws)
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	actions, exported := exportContentActions(agents)

	logger := log.FromContext(ctx)
	if len(actions) == 0 {
		// No database-managed content to export. Adopt the (possibly empty)
		// repository state so the workspace still switches to Git-owned content.
		if err := s.TriggerSyncAndPublish(ctx, ws); err != nil {
			logger.Error("onboard export sync/publish failed", "workspace_id", ws, "err", err)
		}
		final, getErr := s.repo.Get(ctx, ws)
		if getErr != nil {
			final = binding
		}
		return connect.NewResponse(&agentsv1.OnboardWorkspaceRepositoryResponse{
			Binding:   final,
			Published: final.GetActiveCommitSha() != "",
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
		// unchanged, so the workspace stays on database-owned content.
		return connect.NewResponse(&agentsv1.OnboardWorkspaceRepositoryResponse{
			Binding:          res.binding,
			ValidationErrors: res.validationErrors,
		}), nil
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
// bidirectional merge is performed.
func (s *RepoBindingServiceServer) onboardImport(ctx context.Context, ws string, binding *agentsv1.WorkspaceRepoBinding) (*connect.Response[agentsv1.OnboardWorkspaceRepositoryResponse], error) {
	syncResp, err := s.SyncWorkspaceRepository(ctx, connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{}))
	if err != nil {
		return nil, err
	}
	imported := 0
	if syncResp.Msg.GetPublished() && s.contentRepo != nil {
		if snap, snapErr := s.contentRepo.GetSnapshot(ctx, ws); snapErr == nil {
			imported = len(snap.Entries)
		}
	}
	out := syncResp.Msg.GetBinding()
	if out == nil {
		out = binding
	}
	log.FromContext(ctx).Info("workspace repository onboarded via import", "workspace_id", ws,
		"published", syncResp.Msg.GetPublished(), "agents_imported", imported)
	return connect.NewResponse(&agentsv1.OnboardWorkspaceRepositoryResponse{
		Binding:          out,
		Published:        syncResp.Msg.GetPublished(),
		ValidationErrors: syncResp.Msg.GetPublicationErrors(),
		AgentsImported:   int32(imported),
	}), nil
}

// exportContentActions builds a PUT changeset from the database-managed content
// of every non-deleted agent with an assigned Agent ID. Empty fields are
// omitted rather than written as empty files; on read-back the publication
// pipeline treats a missing optional file as an empty (cleared) value, so the
// round trip is lossless. Returns the actions and the number of agents that
// contributed at least one file.
func exportContentActions(agents []*agentsv1.Agent) ([]*agentsv1.ContentFileAction, int) {
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
		contributed := false
		put := func(path, content string) {
			content = strings.TrimSpace(content)
			if content == "" {
				return
			}
			actions = append(actions, &agentsv1.ContentFileAction{
				Path:      path,
				Operation: agentsv1.ContentFileOperation_CONTENT_FILE_OPERATION_PUT,
				Content:   content,
			})
			contributed = true
		}
		put(agentcontent.DescriptionPath(id), a.GetDescription())
		put(agentcontent.PromptPath(id), a.GetConfig().GetInstruction())
		put(agentcontent.GlobalPromptPath(id), a.GetConfig().GetGlobalInstruction())
		if contributed {
			exported++
		}
	}
	return actions, exported
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
