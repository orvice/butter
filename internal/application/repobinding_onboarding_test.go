package application

// Service-level tests for workspace onboarding (EXPORT_CURRENT /
// IMPORT_REPOSITORY) and safe detachment (issue #219): exporting DB-managed
// content to Git, importing only matching Agent IDs, materializing the Active
// Revision back into DB fields on unbind, and refusing to unbind without a
// valid snapshot unless a recovery path is chosen.

import (
	"errors"
	"testing"

	"connectrpc.com/connect"

	agentcontentmemory "go.orx.me/apps/butter/internal/repo/agentcontent/memory"
	repobindingrepo "go.orx.me/apps/butter/internal/repo/repobinding"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestOnboardExportCurrent(t *testing.T) {
	fx, rt := newPublicationFixture(t)
	fx.fake.materialize = true
	ctx := ownerCtx()

	// Give the DB agent content to export.
	ag, err := fx.agentRepo.GetAgent(ctx, "ws-a", "my-agent")
	if err != nil {
		t.Fatalf("GetAgent: %v", err)
	}
	ag.Description = "Database description."
	ag.Config = &agentsv1.AgentConfig{Instruction: "Database prompt instruction."}
	if _, err := fx.agentRepo.UpdateAgent(ctx, "ws-a", ag); err != nil {
		t.Fatalf("UpdateAgent: %v", err)
	}

	reloadsBefore := rt.reloadCount
	resp, err := fx.svc.OnboardWorkspaceRepository(ctx, connect.NewRequest(&agentsv1.OnboardWorkspaceRepositoryRequest{
		Mode: agentsv1.RepoBindingOnboardingMode_REPO_BINDING_ONBOARDING_MODE_EXPORT_CURRENT,
	}))
	if err != nil {
		t.Fatalf("OnboardWorkspaceRepository export: %v", err)
	}
	if len(resp.Msg.GetValidationErrors()) != 0 {
		t.Fatalf("unexpected validation errors: %v", resp.Msg.GetValidationErrors())
	}
	if resp.Msg.GetCommitSha() == "" {
		t.Fatal("expected a non-empty export commit SHA")
	}
	if !resp.Msg.GetPublished() {
		t.Fatal("export should have published the Active Revision")
	}
	if resp.Msg.GetAgentsExported() != 1 {
		t.Fatalf("agents_exported = %d, want 1", resp.Msg.GetAgentsExported())
	}
	if resp.Msg.GetBinding().GetActiveCommitSha() != resp.Msg.GetCommitSha() {
		t.Fatalf("active_commit_sha %q != export commit %q",
			resp.Msg.GetBinding().GetActiveCommitSha(), resp.Msg.GetCommitSha())
	}
	if rt.reloadCount <= reloadsBefore {
		t.Fatal("export should have reloaded the runner after publication")
	}

	// The published snapshot must carry the exported DB content.
	snap, err := fx.svc.contentRepo.GetSnapshot(ctx, "ws-a", resp.Msg.GetBinding().GetActiveCommitSha())
	if err != nil {
		t.Fatalf("GetSnapshot: %v", err)
	}
	got := snap.Entries["my-agent"]
	if got.Instruction != "Database prompt instruction." {
		t.Fatalf("published instruction = %q, want exported DB value", got.Instruction)
	}
	if got.Description != "Database description." {
		t.Fatalf("published description = %q, want exported DB value", got.Description)
	}

	// The export must have written the DB content into Git.
	fileResp, err := fx.svc.GetRepositoryFile(ctx, connect.NewRequest(&agentsv1.GetRepositoryFileRequest{
		Path: "agents/my-agent/prompt.md",
	}))
	if err != nil {
		t.Fatalf("GetRepositoryFile: %v", err)
	}
	if fileResp.Msg.GetContent() != "Database prompt instruction." {
		t.Fatalf("git prompt.md = %q, want exported DB value", fileResp.Msg.GetContent())
	}
}

// TestOnboardExportGatesContentEditingViaAPI proves the ownership switch through
// the real Agent API: a bound-but-not-yet-onboarded workspace still accepts
// content edits (DB-owned), and after EXPORT_CURRENT publishes an Active
// Revision the same edit is ignored (Git-owned, content preserved). This also
// exercises the failed-export recovery path — content stays editable via the API
// until onboarding actually succeeds.
func TestOnboardExportGatesContentEditingViaAPI(t *testing.T) {
	fx, rt := newPublicationFixture(t)
	fx.fake.materialize = true
	ctx := ownerCtx()

	agentSvc := NewAgentServiceServer(fx.agentRepo)
	agentSvc.SetWorkspaceRepo(fx.wsRepo)
	agentSvc.SetRuntime(rt)
	agentSvc.SetContentCoordinator(fx.svc)

	// Before onboarding: bound but not published → DB-owned, edits accepted.
	if owned, _ := fx.svc.IsContentGitOwned(ctx, "ws-a"); owned {
		t.Fatal("workspace must not be Git-owned before onboarding")
	}
	if _, err := agentSvc.UpdateAgent(ctx, connect.NewRequest(&agentsv1.UpdateAgentRequest{
		Agent: &agentsv1.Agent{AgentId: "my-agent", Config: &agentsv1.AgentConfig{Instruction: "V1 prompt."}},
	})); err != nil {
		t.Fatalf("pre-onboard UpdateAgent: %v", err)
	}
	got, _ := fx.agentRepo.GetAgent(ctx, "ws-a", "my-agent")
	if got.GetConfig().GetInstruction() != "V1 prompt." {
		t.Fatalf("pre-onboard edit not applied: %q", got.GetConfig().GetInstruction())
	}

	// Export → workspace becomes Git-owned.
	resp, err := fx.svc.OnboardWorkspaceRepository(ctx, connect.NewRequest(&agentsv1.OnboardWorkspaceRepositoryRequest{
		Mode: agentsv1.RepoBindingOnboardingMode_REPO_BINDING_ONBOARDING_MODE_EXPORT_CURRENT,
	}))
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if !resp.Msg.GetPublished() {
		t.Fatal("export should publish")
	}
	if owned, _ := fx.svc.IsContentGitOwned(ctx, "ws-a"); !owned {
		t.Fatal("workspace must be Git-owned after export")
	}

	// After onboarding: content edits through the Agent API are ignored; the DB
	// field is preserved because content is now Git-owned.
	if _, err := agentSvc.UpdateAgent(ctx, connect.NewRequest(&agentsv1.UpdateAgentRequest{
		Agent: &agentsv1.Agent{AgentId: "my-agent", Config: &agentsv1.AgentConfig{Instruction: "V2 via API"}},
	})); err != nil {
		t.Fatalf("post-onboard UpdateAgent: %v", err)
	}
	got, _ = fx.agentRepo.GetAgent(ctx, "ws-a", "my-agent")
	if got.GetConfig().GetInstruction() != "V1 prompt." {
		t.Fatalf("Git-owned content must be preserved, got %q", got.GetConfig().GetInstruction())
	}
}

// TestGetAgentOverlaysGitContent proves that GetAgent and ListAgents return
// the Effective Agent with Git-owned content (prompt, description) overlaid
// onto the DB agent after sync publishes an Active Revision.
func TestGetAgentOverlaysGitContent(t *testing.T) {
	fx, rt := newPublicationFixture(t)
	ctx := ownerCtx()

	agentSvc := NewAgentServiceServer(fx.agentRepo)
	agentSvc.SetWorkspaceRepo(fx.wsRepo)
	agentSvc.SetRuntime(rt)
	agentSvc.SetContentCoordinator(fx.svc)

	// Before sync: GetAgent returns DB content (empty).
	resp, err := agentSvc.GetAgent(ctx, connect.NewRequest(&agentsv1.GetAgentRequest{AgentId: "my-agent"}))
	if err != nil {
		t.Fatalf("GetAgent before sync: %v", err)
	}
	if resp.Msg.GetAgent().GetConfig().GetInstruction() != "" {
		t.Fatalf("expected empty DB prompt before sync, got %q", resp.Msg.GetAgent().GetConfig().GetInstruction())
	}

	// Sync → publishes Active Revision from Git content.
	syncResp, err := fx.svc.SyncWorkspaceRepository(ctx, connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{}))
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	if !syncResp.Msg.GetPublished() {
		t.Fatal("sync should publish")
	}

	// After sync: GetAgent must return the Git-cached prompt.
	resp, err = agentSvc.GetAgent(ctx, connect.NewRequest(&agentsv1.GetAgentRequest{AgentId: "my-agent"}))
	if err != nil {
		t.Fatalf("GetAgent after sync: %v", err)
	}
	got := resp.Msg.GetAgent()
	if got.GetConfig().GetInstruction() != "You are a helpful agent." {
		t.Fatalf("GetAgent prompt = %q, want Git-cached value", got.GetConfig().GetInstruction())
	}
	if got.GetDescription() != "My agent description." {
		t.Fatalf("GetAgent description = %q, want Git-cached value", got.GetDescription())
	}

	// ListAgents must also overlay.
	listResp, err := agentSvc.ListAgents(ctx, connect.NewRequest(&agentsv1.ListAgentsRequest{}))
	if err != nil {
		t.Fatalf("ListAgents: %v", err)
	}
	found := false
	for _, a := range listResp.Msg.GetAgents() {
		if a.GetAgentId() == "my-agent" {
			found = true
			if a.GetConfig().GetInstruction() != "You are a helpful agent." {
				t.Fatalf("ListAgents prompt = %q, want Git-cached value", a.GetConfig().GetInstruction())
			}
		}
	}
	if !found {
		t.Fatal("agent my-agent not found in ListAgents")
	}
}

func TestSyncRepairsMissingActiveContentSnapshot(t *testing.T) {
	fx, rt := newPublicationFixture(t)
	ctx := ownerCtx()

	if _, err := fx.svc.SyncWorkspaceRepository(ctx, connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{})); err != nil {
		t.Fatalf("initial sync: %v", err)
	}

	// Older deployments stored Active Agent Content only in process memory.
	// Simulate a restart while keeping the persisted binding and repository
	// cache, then sync the unchanged revision to repair the missing snapshot.
	fx.svc.SetContentRepo(agentcontentmemory.New())
	if _, err := fx.svc.SyncWorkspaceRepository(ctx, connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{})); err != nil {
		t.Fatalf("repair sync: %v", err)
	}

	agentSvc := NewAgentServiceServer(fx.agentRepo)
	agentSvc.SetRuntime(rt)
	agentSvc.SetContentCoordinator(fx.svc)
	resp, err := agentSvc.GetAgent(ctx, connect.NewRequest(&agentsv1.GetAgentRequest{AgentId: "my-agent"}))
	if err != nil {
		t.Fatalf("GetAgent after repair sync: %v", err)
	}
	if got := resp.Msg.GetAgent().GetConfig().GetInstruction(); got != "You are a helpful agent." {
		t.Fatalf("GetAgent prompt after repair sync = %q, want Git Active Revision content", got)
	}
}

func TestGetAgentRejectsMissingActiveContentSnapshot(t *testing.T) {
	fx, rt := newPublicationFixture(t)
	ctx := ownerCtx()

	if _, err := fx.svc.SyncWorkspaceRepository(ctx, connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{})); err != nil {
		t.Fatalf("initial sync: %v", err)
	}
	fx.svc.SetContentRepo(agentcontentmemory.New())

	agentSvc := NewAgentServiceServer(fx.agentRepo)
	agentSvc.SetRuntime(rt)
	agentSvc.SetContentCoordinator(fx.svc)
	_, err := agentSvc.GetAgent(ctx, connect.NewRequest(&agentsv1.GetAgentRequest{AgentId: "my-agent"}))
	wantCode(t, err, connect.CodeInternal)
}

func TestOnboardExportRequiresOwnerOrAdmin(t *testing.T) {
	fx, _ := newPublicationFixture(t)
	_, err := fx.svc.OnboardWorkspaceRepository(memberCtx(), connect.NewRequest(&agentsv1.OnboardWorkspaceRepositoryRequest{
		Mode: agentsv1.RepoBindingOnboardingMode_REPO_BINDING_ONBOARDING_MODE_EXPORT_CURRENT,
	}))
	wantCode(t, err, connect.CodePermissionDenied)
}

func TestOnboardInvalidMode(t *testing.T) {
	fx, _ := newPublicationFixture(t)
	_, err := fx.svc.OnboardWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.OnboardWorkspaceRepositoryRequest{}))
	wantCode(t, err, connect.CodeInvalidArgument)
}

func TestOnboardImportRepository(t *testing.T) {
	fx, rt := newPublicationFixture(t)
	ctx := ownerCtx()

	resp, err := fx.svc.OnboardWorkspaceRepository(ctx, connect.NewRequest(&agentsv1.OnboardWorkspaceRepositoryRequest{
		Mode: agentsv1.RepoBindingOnboardingMode_REPO_BINDING_ONBOARDING_MODE_IMPORT_REPOSITORY,
	}))
	if err != nil {
		t.Fatalf("OnboardWorkspaceRepository import: %v", err)
	}
	if !resp.Msg.GetPublished() {
		t.Fatal("import should have published the Active Revision")
	}
	if resp.Msg.GetCommitSha() != "" {
		t.Fatalf("import should not create a commit, got %q", resp.Msg.GetCommitSha())
	}
	// The seeded repo contains agents/my-agent (matching) and agents/unclaimed-dir
	// (no matching Agent ID). Only the matching directory is imported.
	if resp.Msg.GetAgentsImported() != 1 {
		t.Fatalf("agents_imported = %d, want 1 (unknown dirs must stay unclaimed)", resp.Msg.GetAgentsImported())
	}
	if rt.reloadCount == 0 {
		t.Fatal("import should have reloaded the runner after publication")
	}

	snap, err := fx.svc.contentRepo.GetSnapshot(ctx, "ws-a", resp.Msg.GetBinding().GetActiveCommitSha())
	if err != nil {
		t.Fatalf("GetSnapshot: %v", err)
	}
	if _, ok := snap.Entries["my-agent"]; !ok {
		t.Fatal("matching agent my-agent should be imported")
	}
	if _, ok := snap.Entries["unclaimed-dir"]; ok {
		t.Fatal("unknown directory unclaimed-dir must not be imported")
	}
}

func TestDetachMaterializesActiveContent(t *testing.T) {
	fx, rt := newContentEditFixture(t) // syncs + publishes: active snapshot exists
	ctx := ownerCtx()
	binding, err := fx.bindingRepo.Get(ctx, "ws-a")
	if err != nil {
		t.Fatalf("Get binding: %v", err)
	}

	// The published snapshot carries git content the DB agent does not yet have.
	snap, err := fx.svc.contentRepo.GetSnapshot(ctx, "ws-a", binding.GetActiveCommitSha())
	if err != nil {
		t.Fatalf("GetSnapshot: %v", err)
	}
	wantInstruction := snap.Entries["my-agent"].Instruction
	if wantInstruction == "" {
		t.Fatal("precondition: active snapshot should carry a non-empty instruction")
	}

	commitsBefore := len(fx.fake.commits)
	reloadsBefore := rt.reloadCount

	resp, err := fx.svc.DeleteWorkspaceRepoBinding(ctx, connect.NewRequest(&agentsv1.DeleteWorkspaceRepoBindingRequest{}))
	if err != nil {
		t.Fatalf("DeleteWorkspaceRepoBinding: %v", err)
	}
	if resp.Msg.GetAgentsMaterialized() != 1 {
		t.Fatalf("agents_materialized = %d, want 1", resp.Msg.GetAgentsMaterialized())
	}

	// DB content now carries the previously Git-owned instruction.
	ag, err := fx.agentRepo.GetAgent(ctx, "ws-a", "my-agent")
	if err != nil {
		t.Fatalf("GetAgent: %v", err)
	}
	if ag.GetConfig().GetInstruction() != wantInstruction {
		t.Fatalf("materialized instruction = %q, want %q", ag.GetConfig().GetInstruction(), wantInstruction)
	}

	// Detachment never modifies remote Git content.
	if len(fx.fake.commits) != commitsBefore {
		t.Fatalf("detach created %d commit(s); it must never modify remote", len(fx.fake.commits)-commitsBefore)
	}
	// Runtime reloaded before removal.
	if rt.reloadCount <= reloadsBefore {
		t.Fatal("detach should have reloaded the runner")
	}
	// Binding, credential, cache, and snapshot are gone.
	if _, err := fx.bindingRepo.Get(ctx, "ws-a"); !errors.Is(err, repobindingrepo.ErrNotFound) {
		t.Fatalf("binding still present after detach: %v", err)
	}
	if _, err := fx.svc.contentRepo.GetSnapshot(ctx, "ws-a", binding.GetActiveCommitSha()); err == nil {
		t.Fatal("content snapshot should be removed after detach")
	}
}

func TestDetachRefusesWithoutSnapshotThenRecovers(t *testing.T) {
	// Bound with a credential but never synced/published: no active snapshot.
	fx, _ := newPublicationFixture(t)
	ctx := ownerCtx()

	// Seed a DB agent with content that must survive a KEEP_DATABASE detach.
	ag, _ := fx.agentRepo.GetAgent(ctx, "ws-a", "my-agent")
	ag.Config = &agentsv1.AgentConfig{Instruction: "Untouched DB instruction."}
	if _, err := fx.agentRepo.UpdateAgent(ctx, "ws-a", ag); err != nil {
		t.Fatalf("UpdateAgent: %v", err)
	}

	// Default recovery refuses.
	_, err := fx.svc.DeleteWorkspaceRepoBinding(ctx, connect.NewRequest(&agentsv1.DeleteWorkspaceRepoBindingRequest{}))
	wantCode(t, err, connect.CodeFailedPrecondition)
	if _, err := fx.bindingRepo.Get(ctx, "ws-a"); err != nil {
		t.Fatalf("binding must remain after a refused detach: %v", err)
	}

	// KEEP_DATABASE proceeds, keeping DB content as-is.
	resp, err := fx.svc.DeleteWorkspaceRepoBinding(ctx, connect.NewRequest(&agentsv1.DeleteWorkspaceRepoBindingRequest{
		Recovery: agentsv1.RepoBindingDetachRecovery_REPO_BINDING_DETACH_RECOVERY_KEEP_DATABASE,
	}))
	if err != nil {
		t.Fatalf("DeleteWorkspaceRepoBinding KEEP_DATABASE: %v", err)
	}
	if resp.Msg.GetAgentsMaterialized() != 0 {
		t.Fatalf("agents_materialized = %d, want 0 (no snapshot to materialize)", resp.Msg.GetAgentsMaterialized())
	}
	if _, err := fx.bindingRepo.Get(ctx, "ws-a"); !errors.Is(err, repobindingrepo.ErrNotFound) {
		t.Fatalf("binding should be removed after KEEP_DATABASE detach: %v", err)
	}
	got, _ := fx.agentRepo.GetAgent(ctx, "ws-a", "my-agent")
	if got.GetConfig().GetInstruction() != "Untouched DB instruction." {
		t.Fatalf("KEEP_DATABASE must not alter DB content, got %q", got.GetConfig().GetInstruction())
	}
}

func TestDetachRequiresOwnerOrAdmin(t *testing.T) {
	fx, _ := newContentEditFixture(t)
	_, err := fx.svc.DeleteWorkspaceRepoBinding(memberCtx(), connect.NewRequest(&agentsv1.DeleteWorkspaceRepoBindingRequest{}))
	wantCode(t, err, connect.CodePermissionDenied)
}

// TestRepositoryMigrationEndToEnd walks the full onboarding/offboarding
// lifecycle: an export that fails validation leaves the workspace DB-owned,
// the recovered export publishes an Active Revision, and detachment
// materializes every agent's content back into the database without touching
// the remote.
func TestRepositoryMigrationEndToEnd(t *testing.T) {
	fx, rt := newPublicationFixture(t)
	fx.fake.materialize = true
	ctx := ownerCtx()
	exportReq := func() *connect.Request[agentsv1.OnboardWorkspaceRepositoryRequest] {
		return connect.NewRequest(&agentsv1.OnboardWorkspaceRepositoryRequest{
			Mode: agentsv1.RepoBindingOnboardingMode_REPO_BINDING_ONBOARDING_MODE_EXPORT_CURRENT,
		})
	}

	// Root agent has a prompt; a second LLM agent has a description but no
	// prompt — an invalid Effective Agent that must block the export.
	root, _ := fx.agentRepo.GetAgent(ctx, "ws-a", "my-agent")
	root.Config = &agentsv1.AgentConfig{Instruction: "Root prompt."}
	if _, err := fx.agentRepo.UpdateAgent(ctx, "ws-a", root); err != nil {
		t.Fatalf("update root: %v", err)
	}
	if _, err := fx.agentRepo.CreateAgent(ctx, "ws-a", &agentsv1.Agent{
		Name: "Planner", AgentId: "planner", Type: agentsv1.AgentType_AGENT_TYPE_LLM,
		Description: "Plans multi-step work.",
	}); err != nil {
		t.Fatalf("create planner: %v", err)
	}

	// 1) Export fails validation → nothing committed, workspace stays DB-owned.
	commitsBefore := len(fx.fake.commits)
	resp, err := fx.svc.OnboardWorkspaceRepository(ctx, exportReq())
	if err != nil {
		t.Fatalf("export (invalid): %v", err)
	}
	if len(resp.Msg.GetValidationErrors()) == 0 {
		t.Fatal("expected validation errors for the prompt-less LLM agent")
	}
	if resp.Msg.GetPublished() {
		t.Fatal("invalid content must not publish")
	}
	if len(fx.fake.commits) != commitsBefore {
		t.Fatal("failed export must not create a commit")
	}
	if b, _ := fx.bindingRepo.Get(ctx, "ws-a"); b.GetActiveCommitSha() != "" {
		t.Fatal("workspace must remain DB-owned after a failed export")
	}

	// 2) Recover: give the planner a prompt, re-export succeeds and publishes.
	planner, _ := fx.agentRepo.GetAgent(ctx, "ws-a", "planner")
	planner.Config = &agentsv1.AgentConfig{Instruction: "Plan carefully."}
	if _, err := fx.agentRepo.UpdateAgent(ctx, "ws-a", planner); err != nil {
		t.Fatalf("fix planner: %v", err)
	}
	resp2, err := fx.svc.OnboardWorkspaceRepository(ctx, exportReq())
	if err != nil {
		t.Fatalf("export (recovered): %v", err)
	}
	if len(resp2.Msg.GetValidationErrors()) != 0 {
		t.Fatalf("unexpected validation errors after recovery: %v", resp2.Msg.GetValidationErrors())
	}
	if !resp2.Msg.GetPublished() {
		t.Fatal("recovered export should publish")
	}
	if resp2.Msg.GetAgentsExported() != 2 {
		t.Fatalf("agents_exported = %d, want 2", resp2.Msg.GetAgentsExported())
	}

	// 3) Detach materializes both agents back and never modifies remote.
	commitsBefore = len(fx.fake.commits)
	reloadsBefore := rt.reloadCount
	del, err := fx.svc.DeleteWorkspaceRepoBinding(ctx, connect.NewRequest(&agentsv1.DeleteWorkspaceRepoBindingRequest{}))
	if err != nil {
		t.Fatalf("detach: %v", err)
	}
	if del.Msg.GetAgentsMaterialized() != 2 {
		t.Fatalf("agents_materialized = %d, want 2", del.Msg.GetAgentsMaterialized())
	}
	if len(fx.fake.commits) != commitsBefore {
		t.Fatal("detach must never modify remote content")
	}
	if rt.reloadCount <= reloadsBefore {
		t.Fatal("detach should reload the runner")
	}
	if _, err := fx.bindingRepo.Get(ctx, "ws-a"); !errors.Is(err, repobindingrepo.ErrNotFound) {
		t.Fatalf("binding should be removed after detach: %v", err)
	}
}
