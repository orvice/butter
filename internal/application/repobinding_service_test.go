package application

// Service-level tests for WorkspaceRepoBindingService (issue #214): role
// authorization (member read-only vs owner/admin manage), binding
// validation via a fake provider, and secret handling (plaintext PATs never
// escape through responses or persisted models).

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	agentcontentmemory "go.orx.me/apps/butter/internal/repo/agentcontent/memory"
	"go.orx.me/apps/butter/internal/gitprovider"
	"go.orx.me/apps/butter/internal/repo/auth"
	configmemory "go.orx.me/apps/butter/internal/repo/config/memory"
	githostmemory "go.orx.me/apps/butter/internal/repo/githost/memory"
	repobindingrepo "go.orx.me/apps/butter/internal/repo/repobinding"
	repobindingmemory "go.orx.me/apps/butter/internal/repo/repobinding/memory"
	"go.orx.me/apps/butter/internal/repo/repocache"
	repocachememory "go.orx.me/apps/butter/internal/repo/repocache/memory"
	workspacememory "go.orx.me/apps/butter/internal/repo/workspace/memory"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	testEncryptionKey = "0123456789abcdef0123456789abcdef"
	testPAT           = "ghp_plaintext_pat_do_not_leak"
)

// fakeProviderClient implements gitprovider.Client deterministically.
type fakeProviderClient struct {
	repo        *gitprovider.Repository
	repoErr     error
	branches    map[string]string
	branchErr   error // if set, GetBranchHead returns this error
	trees       map[string][]gitprovider.TreeEntry // keyed by "ref:path"
	blobs       map[string][]byte                  // keyed by "ref:path"
	treeErr     error
	blobErr     error
	comparisons map[string]string // keyed by "base...head" → status

	// Write operation state
	mu                sync.Mutex
	commits           []fakeCommit
	createdBranches   map[string]string // branch → sha
	deletedBranches   []string
	createdCRs        []fakeChangeRequest
	commitErr         error
	createBranchErr   error
	deleteBranchErr   error
	createCRErr       error
}

type fakeCommit struct {
	branch    string
	parentSHA string
	message   string
	actions   []gitprovider.FileAction
	sha       string
}

type fakeChangeRequest struct {
	source string
	target string
	title  string
}

func (f *fakeProviderClient) GetRepository(context.Context) (*gitprovider.Repository, error) {
	if f.repoErr != nil {
		return nil, f.repoErr
	}
	return f.repo, nil
}

func (f *fakeProviderClient) GetBranchHead(_ context.Context, branch string) (string, error) {
	if f.branchErr != nil {
		return "", f.branchErr
	}
	sha, ok := f.branches[branch]
	if !ok {
		return "", gitprovider.ErrNotFound
	}
	return sha, nil
}

func (f *fakeProviderClient) GetTree(_ context.Context, ref, path string) ([]gitprovider.TreeEntry, error) {
	if f.treeErr != nil {
		return nil, f.treeErr
	}
	key := ref + ":" + path
	entries, ok := f.trees[key]
	if !ok {
		return nil, gitprovider.ErrNotFound
	}
	return entries, nil
}

func (f *fakeProviderClient) GetBlob(_ context.Context, ref, path string) ([]byte, error) {
	if f.blobErr != nil {
		return nil, f.blobErr
	}
	key := ref + ":" + path
	data, ok := f.blobs[key]
	if !ok {
		return nil, gitprovider.ErrNotFound
	}
	return data, nil
}

func (f *fakeProviderClient) CompareCommits(_ context.Context, base, head string) (*gitprovider.CommitComparison, error) {
	if base == head {
		return &gitprovider.CommitComparison{Status: "identical"}, nil
	}
	if f.comparisons != nil {
		if status, ok := f.comparisons[base+"..."+head]; ok {
			return &gitprovider.CommitComparison{Status: status}, nil
		}
	}
	return &gitprovider.CommitComparison{Status: "ahead"}, nil
}

func (f *fakeProviderClient) CreateCommit(_ context.Context, branch, parentSHA, message string, actions []gitprovider.FileAction) (*gitprovider.CommitResult, error) {
	if f.commitErr != nil {
		return nil, f.commitErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	sha := fmt.Sprintf("commit-%d", len(f.commits)+1)
	fc := fakeCommit{
		branch:    branch,
		parentSHA: parentSHA,
		message:   message,
		actions:   actions,
		sha:       sha,
	}
	f.commits = append(f.commits, fc)
	f.branches[branch] = sha
	return &gitprovider.CommitResult{SHA: sha}, nil
}

func (f *fakeProviderClient) CreateBranch(_ context.Context, branch, sha string) error {
	if f.createBranchErr != nil {
		return f.createBranchErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.createdBranches == nil {
		f.createdBranches = make(map[string]string)
	}
	f.createdBranches[branch] = sha
	f.branches[branch] = sha
	return nil
}

func (f *fakeProviderClient) DeleteBranch(_ context.Context, branch string) error {
	if f.deleteBranchErr != nil {
		return f.deleteBranchErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deletedBranches = append(f.deletedBranches, branch)
	delete(f.createdBranches, branch)
	delete(f.branches, branch)
	return nil
}

func (f *fakeProviderClient) CreateChangeRequest(_ context.Context, source, target, title, description string) (*gitprovider.ChangeRequestResult, error) {
	if f.createCRErr != nil {
		return nil, f.createCRErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.createdCRs = append(f.createdCRs, fakeChangeRequest{source: source, target: target, title: title})
	id := len(f.createdCRs)
	return &gitprovider.ChangeRequestResult{
		ID:    id,
		URL:   fmt.Sprintf("https://github.com/acme/agents/pull/%d", id),
		Title: title,
	}, nil
}

type bindingFixture struct {
	svc         *RepoBindingServiceServer
	bindingRepo *repobindingmemory.Store
	hostRepo    *githostmemory.Store
	wsRepo      *workspacememory.Store
	agentRepo   *configmemory.Store
	fake        *fakeProviderClient
	// lastProviderCfg captures what the service handed the provider factory.
	lastProviderCfg *gitprovider.Config
}

func newBindingFixture(t *testing.T) *bindingFixture {
	t.Helper()
	fx := &bindingFixture{
		bindingRepo: repobindingmemory.New(),
		hostRepo:    githostmemory.New(),
		wsRepo:      workspacememory.New(),
		agentRepo:   configmemory.New(),
		fake: &fakeProviderClient{
			repo: &gitprovider.Repository{
				FullName: "acme/agents", Private: true, DefaultBranch: "main",
				CanRead: true, CanWrite: true, CanOpenChangeRequests: true,
			},
			branches: map[string]string{"main": "abc123"},
		},
	}
	ctx := context.Background()
	if _, err := fx.hostRepo.Create(ctx, &agentsv1.GitHost{
		Id: "gh-1", Name: "GitHub.com",
		Kind:       agentsv1.GitHostKind_GIT_HOST_KIND_GITHUB,
		ApiBaseUrl: "https://api.github.com",
	}); err != nil {
		t.Fatalf("seed git host: %v", err)
	}
	for _, ws := range []struct{ id, name string }{{"ws-a", "Alpha"}, {"ws-b", "Beta"}} {
		if _, err := fx.wsRepo.CreateWorkspace(ctx, &agentsv1.Workspace{Id: ws.id, Name: ws.name, Slug: ws.id}); err != nil {
			t.Fatalf("seed workspace: %v", err)
		}
	}
	for _, m := range []struct{ user, role string }{
		{"owner-user", "owner"}, {"admin-user", "admin"}, {"member-user", "member"},
	} {
		if _, err := fx.wsRepo.AddMember(ctx, &agentsv1.WorkspaceMember{
			WorkspaceId: "ws-a", UserId: m.user, Role: m.role,
		}); err != nil {
			t.Fatalf("seed member: %v", err)
		}
	}
	if _, err := fx.agentRepo.CreateAgent(ctx, "ws-a", &agentsv1.Agent{
		Name: "My Agent", AgentId: "my-agent",
	}); err != nil {
		t.Fatalf("seed agent: %v", err)
	}

	svc := NewRepoBindingServiceServer(fx.bindingRepo, fx.hostRepo)
	svc.SetWorkspaceRepo(fx.wsRepo)
	svc.SetAgentRepo(fx.agentRepo)
	svc.SetEncryptionKeyProvider(func() string { return testEncryptionKey })
	svc.SetProviderClientFactory(func(cfg gitprovider.Config) (gitprovider.Client, error) {
		fx.lastProviderCfg = &cfg
		return fx.fake, nil
	})
	fx.svc = svc
	return fx
}

func ctxAs(userID, role, workspaceID string) context.Context {
	ctx := workspace.WithID(context.Background(), workspaceID)
	return auth.WithAuthenticated(ctx, &agentsv1.User{Id: userID, Role: role}, nil)
}

func ownerCtx() context.Context   { return ctxAs("owner-user", "user", "ws-a") }
func wsAdminCtx() context.Context { return ctxAs("admin-user", "user", "ws-a") }
func memberCtx() context.Context  { return ctxAs("member-user", "user", "ws-a") }

func validBinding() *agentsv1.WorkspaceRepoBinding {
	return &agentsv1.WorkspaceRepoBinding{
		GitHostId:  "gh-1",
		Repository: "acme/agents",
		Branch:     "main",
	}
}

func putBinding(t *testing.T, fx *bindingFixture, ctx context.Context) *agentsv1.WorkspaceRepoBinding {
	t.Helper()
	resp, err := fx.svc.PutWorkspaceRepoBinding(ctx, connect.NewRequest(&agentsv1.PutWorkspaceRepoBindingRequest{
		Binding: validBinding(),
	}))
	if err != nil {
		t.Fatalf("PutWorkspaceRepoBinding: %v", err)
	}
	return resp.Msg.GetBinding()
}

func setCredential(t *testing.T, fx *bindingFixture, ctx context.Context) {
	t.Helper()
	if _, err := fx.svc.SetWorkspaceRepoBindingCredential(ctx, connect.NewRequest(&agentsv1.SetWorkspaceRepoBindingCredentialRequest{
		Pat: testPAT,
	})); err != nil {
		t.Fatalf("SetWorkspaceRepoBindingCredential: %v", err)
	}
}

func wantCode(t *testing.T, err error, code connect.Code) {
	t.Helper()
	if connect.CodeOf(err) != code {
		t.Fatalf("err = %v (code %v), want code %v", err, connect.CodeOf(err), code)
	}
}

func TestRepoBindingPutAndGet(t *testing.T) {
	fx := newBindingFixture(t)
	in := validBinding()
	in.RootPath = " butter/./content "
	// Server-owned fields on input must be ignored.
	in.CredentialSet = true
	in.Status = &agentsv1.RepoBindingStatus{State: agentsv1.RepoBindingConnectionState_REPO_BINDING_CONNECTION_STATE_OK}
	resp, err := fx.svc.PutWorkspaceRepoBinding(ownerCtx(), connect.NewRequest(&agentsv1.PutWorkspaceRepoBindingRequest{Binding: in}))
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	b := resp.Msg.GetBinding()
	if b.GetWorkspaceId() != "ws-a" || b.GetRootPath() != "butter/content" {
		t.Fatalf("unexpected binding: %v", b)
	}
	if b.GetWriteMode() != agentsv1.RepoBindingWriteMode_REPO_BINDING_WRITE_MODE_DIRECT_COMMIT {
		t.Fatalf("write mode not defaulted: %v", b.GetWriteMode())
	}
	if b.GetContentSchemaVersion() != 1 {
		t.Fatalf("schema version not defaulted: %v", b.GetContentSchemaVersion())
	}
	if b.GetCredentialSet() {
		t.Fatal("credential_set forged through Put")
	}
	if b.GetStatus().GetState() != agentsv1.RepoBindingConnectionState_REPO_BINDING_CONNECTION_STATE_UNVALIDATED {
		t.Fatalf("status not reset to UNVALIDATED: %v", b.GetStatus())
	}

	got, err := fx.svc.GetWorkspaceRepoBinding(memberCtx(), connect.NewRequest(&agentsv1.GetWorkspaceRepoBindingRequest{}))
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Msg.GetBinding().GetRepository() != "acme/agents" {
		t.Fatalf("member read failed: %v", got.Msg.GetBinding())
	}
}

func TestRepoBindingGetWithoutBindingIsEmpty(t *testing.T) {
	fx := newBindingFixture(t)
	resp, err := fx.svc.GetWorkspaceRepoBinding(memberCtx(), connect.NewRequest(&agentsv1.GetWorkspaceRepoBindingRequest{}))
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if resp.Msg.GetBinding() != nil {
		t.Fatalf("expected unset binding, got %v", resp.Msg.GetBinding())
	}
}

func TestRepoBindingAuthorization(t *testing.T) {
	fx := newBindingFixture(t)
	putBinding(t, fx, ownerCtx())

	// Members hold read-only access.
	_, err := fx.svc.PutWorkspaceRepoBinding(memberCtx(), connect.NewRequest(&agentsv1.PutWorkspaceRepoBindingRequest{Binding: validBinding()}))
	wantCode(t, err, connect.CodePermissionDenied)
	_, err = fx.svc.DeleteWorkspaceRepoBinding(memberCtx(), connect.NewRequest(&agentsv1.DeleteWorkspaceRepoBindingRequest{}))
	wantCode(t, err, connect.CodePermissionDenied)
	_, err = fx.svc.SetWorkspaceRepoBindingCredential(memberCtx(), connect.NewRequest(&agentsv1.SetWorkspaceRepoBindingCredentialRequest{Pat: testPAT}))
	wantCode(t, err, connect.CodePermissionDenied)
	_, err = fx.svc.ValidateWorkspaceRepoBinding(memberCtx(), connect.NewRequest(&agentsv1.ValidateWorkspaceRepoBindingRequest{}))
	wantCode(t, err, connect.CodePermissionDenied)

	// Workspace admins manage like owners.
	if _, err := fx.svc.PutWorkspaceRepoBinding(wsAdminCtx(), connect.NewRequest(&agentsv1.PutWorkspaceRepoBindingRequest{Binding: validBinding()})); err != nil {
		t.Fatalf("admin Put: %v", err)
	}
	// Global admins bypass workspace roles.
	globalAdmin := auth.WithAdmin(workspace.WithID(context.Background(), "ws-a"))
	if _, err := fx.svc.PutWorkspaceRepoBinding(globalAdmin, connect.NewRequest(&agentsv1.PutWorkspaceRepoBindingRequest{Binding: validBinding()})); err != nil {
		t.Fatalf("global admin Put: %v", err)
	}
	// Non-members are indistinguishable from missing workspaces.
	_, err = fx.svc.PutWorkspaceRepoBinding(ctxAs("stranger", "user", "ws-a"), connect.NewRequest(&agentsv1.PutWorkspaceRepoBindingRequest{Binding: validBinding()}))
	wantCode(t, err, connect.CodeNotFound)
}

func TestRepoBindingPutValidation(t *testing.T) {
	fx := newBindingFixture(t)
	cases := []struct {
		name   string
		mutate func(*agentsv1.WorkspaceRepoBinding)
	}{
		{"unknown host", func(b *agentsv1.WorkspaceRepoBinding) { b.GitHostId = "gh-missing" }},
		{"missing repository", func(b *agentsv1.WorkspaceRepoBinding) { b.Repository = "" }},
		{"repository without namespace", func(b *agentsv1.WorkspaceRepoBinding) { b.Repository = "solo" }},
		{"missing branch", func(b *agentsv1.WorkspaceRepoBinding) { b.Branch = "" }},
		{"absolute root path", func(b *agentsv1.WorkspaceRepoBinding) { b.RootPath = "/etc" }},
		{"traversal root path", func(b *agentsv1.WorkspaceRepoBinding) { b.RootPath = "../outside" }},
		{"embedded traversal is rejected not cleaned", func(b *agentsv1.WorkspaceRepoBinding) { b.RootPath = "a/../b" }},
		{"traversal repository", func(b *agentsv1.WorkspaceRepoBinding) { b.Repository = "acme/../etc" }},
		{"unsupported schema version", func(b *agentsv1.WorkspaceRepoBinding) { b.ContentSchemaVersion = 2 }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := validBinding()
			tc.mutate(b)
			_, err := fx.svc.PutWorkspaceRepoBinding(ownerCtx(), connect.NewRequest(&agentsv1.PutWorkspaceRepoBindingRequest{Binding: b}))
			wantCode(t, err, connect.CodeInvalidArgument)
		})
	}
}

func TestRepoBindingCredentialLifecycle(t *testing.T) {
	fx := newBindingFixture(t)

	// No binding yet: nothing to attach a credential to.
	_, err := fx.svc.SetWorkspaceRepoBindingCredential(ownerCtx(), connect.NewRequest(&agentsv1.SetWorkspaceRepoBindingCredentialRequest{Pat: testPAT}))
	wantCode(t, err, connect.CodeNotFound)

	putBinding(t, fx, ownerCtx())
	resp, err := fx.svc.SetWorkspaceRepoBindingCredential(ownerCtx(), connect.NewRequest(&agentsv1.SetWorkspaceRepoBindingCredentialRequest{Pat: testPAT}))
	if err != nil {
		t.Fatalf("SetCredential: %v", err)
	}
	b := resp.Msg.GetBinding()
	if !b.GetCredentialSet() || b.GetCredentialUpdatedAt() == nil {
		t.Fatalf("credential fields not reported: %v", b)
	}
	// Stored ciphertext must not be the plaintext.
	ct, err := fx.bindingRepo.GetCredential(context.Background(), "ws-a")
	if err != nil {
		t.Fatalf("GetCredential: %v", err)
	}
	if strings.Contains(ct, testPAT) {
		t.Fatal("credential stored unencrypted")
	}
}

func TestRepoBindingCredentialRequiresEncryptionKey(t *testing.T) {
	fx := newBindingFixture(t)
	putBinding(t, fx, ownerCtx())
	fx.svc.SetEncryptionKeyProvider(func() string { return "" })
	_, err := fx.svc.SetWorkspaceRepoBindingCredential(ownerCtx(), connect.NewRequest(&agentsv1.SetWorkspaceRepoBindingCredentialRequest{Pat: testPAT}))
	wantCode(t, err, connect.CodeFailedPrecondition)
}

func TestRepoBindingValidateHappyPath(t *testing.T) {
	fx := newBindingFixture(t)
	putBinding(t, fx, ownerCtx())
	setCredential(t, fx, ownerCtx())

	resp, err := fx.svc.ValidateWorkspaceRepoBinding(ownerCtx(), connect.NewRequest(&agentsv1.ValidateWorkspaceRepoBindingRequest{}))
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	status := resp.Msg.GetBinding().GetStatus()
	if status.GetState() != agentsv1.RepoBindingConnectionState_REPO_BINDING_CONNECTION_STATE_OK {
		t.Fatalf("state = %v, error = %q", status.GetState(), status.GetError())
	}
	if status.GetLastValidatedAt() == nil {
		t.Fatal("last_validated_at not set")
	}
	if len(status.GetChecks()) != 4 {
		t.Fatalf("expected 4 checks, got %v", status.GetChecks())
	}
	// change_request_capability is reported but not required in direct mode.
	for _, c := range status.GetChecks() {
		if c.GetName() == checkChangeRequest && c.GetRequired() {
			t.Fatal("change request check must not gate direct-commit mode")
		}
	}
	// The provider factory received the decrypted PAT and the host's API root.
	if fx.lastProviderCfg == nil || fx.lastProviderCfg.Token != testPAT {
		t.Fatalf("provider did not receive the decrypted PAT: %+v", fx.lastProviderCfg)
	}
	if fx.lastProviderCfg.APIBaseURL != "https://api.github.com" || fx.lastProviderCfg.Kind != gitprovider.KindGitHub {
		t.Fatalf("provider config mismatch: %+v", fx.lastProviderCfg)
	}
	// The outcome is persisted, visible to members via Get.
	got, err := fx.svc.GetWorkspaceRepoBinding(memberCtx(), connect.NewRequest(&agentsv1.GetWorkspaceRepoBindingRequest{}))
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Msg.GetBinding().GetStatus().GetState() != agentsv1.RepoBindingConnectionState_REPO_BINDING_CONNECTION_STATE_OK {
		t.Fatalf("validation outcome not persisted: %v", got.Msg.GetBinding().GetStatus())
	}
}

func TestRepoBindingValidateFailures(t *testing.T) {
	newValidated := func(t *testing.T, mutate func(*bindingFixture), bind func(*agentsv1.WorkspaceRepoBinding)) *agentsv1.RepoBindingStatus {
		t.Helper()
		fx := newBindingFixture(t)
		b := validBinding()
		if bind != nil {
			bind(b)
		}
		if _, err := fx.svc.PutWorkspaceRepoBinding(ownerCtx(), connect.NewRequest(&agentsv1.PutWorkspaceRepoBindingRequest{Binding: b})); err != nil {
			t.Fatalf("Put: %v", err)
		}
		setCredential(t, fx, ownerCtx())
		if mutate != nil {
			mutate(fx)
		}
		resp, err := fx.svc.ValidateWorkspaceRepoBinding(ownerCtx(), connect.NewRequest(&agentsv1.ValidateWorkspaceRepoBindingRequest{}))
		if err != nil {
			t.Fatalf("Validate: %v", err)
		}
		return resp.Msg.GetBinding().GetStatus()
	}

	failed := agentsv1.RepoBindingConnectionState_REPO_BINDING_CONNECTION_STATE_FAILED

	t.Run("unauthorized credential", func(t *testing.T) {
		status := newValidated(t, func(fx *bindingFixture) { fx.fake.repoErr = gitprovider.ErrUnauthorized }, nil)
		if status.GetState() != failed {
			t.Fatalf("state = %v", status.GetState())
		}
		if !strings.Contains(status.GetError(), checkRepositoryRead) {
			t.Fatalf("error = %q", status.GetError())
		}
		if strings.Contains(status.GetError(), testPAT) {
			t.Fatal("error leaks PAT")
		}
	})

	t.Run("missing branch", func(t *testing.T) {
		status := newValidated(t, nil, func(b *agentsv1.WorkspaceRepoBinding) { b.Branch = "gone" })
		if status.GetState() != failed || !strings.Contains(status.GetError(), checkBranchExists) {
			t.Fatalf("state = %v, error = %q", status.GetState(), status.GetError())
		}
	})

	t.Run("read-only credential fails direct commit", func(t *testing.T) {
		status := newValidated(t, func(fx *bindingFixture) {
			fx.fake.repo.CanWrite = false
			fx.fake.repo.CanOpenChangeRequests = false
		}, nil)
		if status.GetState() != failed || !strings.Contains(status.GetError(), checkWriteCapability) {
			t.Fatalf("state = %v, error = %q", status.GetState(), status.GetError())
		}
	})

	t.Run("change request mode requires cr capability", func(t *testing.T) {
		status := newValidated(t, func(fx *bindingFixture) {
			fx.fake.repo.CanOpenChangeRequests = false
		}, func(b *agentsv1.WorkspaceRepoBinding) {
			b.WriteMode = agentsv1.RepoBindingWriteMode_REPO_BINDING_WRITE_MODE_CHANGE_REQUEST
		})
		if status.GetState() != failed || !strings.Contains(status.GetError(), checkChangeRequest) {
			t.Fatalf("state = %v, error = %q", status.GetState(), status.GetError())
		}
	})

	t.Run("validate without credential", func(t *testing.T) {
		fx := newBindingFixture(t)
		putBinding(t, fx, ownerCtx())
		_, err := fx.svc.ValidateWorkspaceRepoBinding(ownerCtx(), connect.NewRequest(&agentsv1.ValidateWorkspaceRepoBindingRequest{}))
		wantCode(t, err, connect.CodeFailedPrecondition)
	})
}

func TestRepoBindingOverlaps(t *testing.T) {
	fx := newBindingFixture(t)
	putBinding(t, fx, ownerCtx())
	// ws-b binds the same location (global admin acts across workspaces).
	wsB := auth.WithAdmin(workspace.WithID(context.Background(), "ws-b"))
	if _, err := fx.svc.PutWorkspaceRepoBinding(wsB, connect.NewRequest(&agentsv1.PutWorkspaceRepoBindingRequest{Binding: validBinding()})); err != nil {
		t.Fatalf("Put ws-b: %v", err)
	}

	resp, err := fx.svc.GetWorkspaceRepoBinding(memberCtx(), connect.NewRequest(&agentsv1.GetWorkspaceRepoBindingRequest{}))
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	overlaps := resp.Msg.GetOverlaps()
	if len(overlaps) != 1 || overlaps[0].GetWorkspaceId() != "ws-b" || overlaps[0].GetWorkspaceName() != "Beta" {
		t.Fatalf("unexpected overlaps: %v", overlaps)
	}

	// A duplicate host record for the same endpoint and a case variant of
	// the repository still resolve to the same effective location.
	if _, err := fx.hostRepo.Create(context.Background(), &agentsv1.GitHost{
		Id: "gh-dup", Name: "GitHub duplicate", Kind: agentsv1.GitHostKind_GIT_HOST_KIND_GITHUB,
		ApiBaseUrl: "https://API.github.com/",
	}); err != nil {
		t.Fatalf("seed duplicate host: %v", err)
	}
	alias := validBinding()
	alias.GitHostId = "gh-dup"
	alias.Repository = "ACME/Agents"
	if _, err := fx.svc.PutWorkspaceRepoBinding(wsB, connect.NewRequest(&agentsv1.PutWorkspaceRepoBindingRequest{Binding: alias})); err != nil {
		t.Fatalf("Put ws-b alias: %v", err)
	}
	resp, err = fx.svc.GetWorkspaceRepoBinding(memberCtx(), connect.NewRequest(&agentsv1.GetWorkspaceRepoBindingRequest{}))
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(resp.Msg.GetOverlaps()) != 1 || resp.Msg.GetOverlaps()[0].GetWorkspaceId() != "ws-b" {
		t.Fatalf("host/case alias not detected as overlap: %v", resp.Msg.GetOverlaps())
	}

	// A different root path is a different effective location.
	diff := validBinding()
	diff.RootPath = "other"
	if _, err := fx.svc.PutWorkspaceRepoBinding(wsB, connect.NewRequest(&agentsv1.PutWorkspaceRepoBindingRequest{Binding: diff})); err != nil {
		t.Fatalf("Put ws-b: %v", err)
	}
	resp, err = fx.svc.GetWorkspaceRepoBinding(memberCtx(), connect.NewRequest(&agentsv1.GetWorkspaceRepoBindingRequest{}))
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(resp.Msg.GetOverlaps()) != 0 {
		t.Fatalf("expected no overlaps, got %v", resp.Msg.GetOverlaps())
	}
}

// TestRepoBindingResponsesNeverContainPAT serializes every RPC response
// produced during a full binding lifecycle and asserts the plaintext PAT
// never appears (issue #214 secret handling).
func TestRepoBindingResponsesNeverContainPAT(t *testing.T) {
	fx := newBindingFixture(t)

	var payloads []proto.Message
	put, err := fx.svc.PutWorkspaceRepoBinding(ownerCtx(), connect.NewRequest(&agentsv1.PutWorkspaceRepoBindingRequest{Binding: validBinding()}))
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	payloads = append(payloads, put.Msg)
	cred, err := fx.svc.SetWorkspaceRepoBindingCredential(ownerCtx(), connect.NewRequest(&agentsv1.SetWorkspaceRepoBindingCredentialRequest{Pat: testPAT}))
	if err != nil {
		t.Fatalf("SetCredential: %v", err)
	}
	payloads = append(payloads, cred.Msg)
	val, err := fx.svc.ValidateWorkspaceRepoBinding(ownerCtx(), connect.NewRequest(&agentsv1.ValidateWorkspaceRepoBindingRequest{}))
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	payloads = append(payloads, val.Msg)
	get, err := fx.svc.GetWorkspaceRepoBinding(memberCtx(), connect.NewRequest(&agentsv1.GetWorkspaceRepoBindingRequest{}))
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	payloads = append(payloads, get.Msg)

	for _, msg := range payloads {
		raw, err := protojson.Marshal(msg)
		if err != nil {
			t.Fatalf("marshal response: %v", err)
		}
		if strings.Contains(string(raw), testPAT) {
			t.Fatalf("response leaks PAT: %s", raw)
		}
	}

	// Failure paths must not leak either.
	fx.fake.repoErr = errors.New("boom with " + testPAT + " inside")
	val, err = fx.svc.ValidateWorkspaceRepoBinding(ownerCtx(), connect.NewRequest(&agentsv1.ValidateWorkspaceRepoBindingRequest{}))
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	raw, _ := protojson.Marshal(val.Msg)
	if strings.Contains(string(raw), testPAT) {
		t.Fatalf("validation status leaks provider error containing PAT: %s", raw)
	}
}

// failingCredentialRepo makes SetCredential fail after the fixture is set
// up, to exercise the partial-failure path of credential replacement.
type failingCredentialRepo struct {
	repobindingrepo.Repository
	fail bool
}

func (f *failingCredentialRepo) SetCredential(ctx context.Context, ws, ct string) error {
	if f.fail {
		return errors.New("storage unavailable")
	}
	return f.Repository.SetCredential(ctx, ws, ct)
}

// TestRepoBindingCredentialReplacementFailsClosed proves ADR-0005's
// replacement contract survives partial failure: when the credential write
// fails, the binding is left UNVALIDATED with the old credential — never a
// new credential wearing a stale OK status.
func TestRepoBindingCredentialReplacementFailsClosed(t *testing.T) {
	fx := newBindingFixture(t)
	failing := &failingCredentialRepo{Repository: fx.bindingRepo}
	fx.svc.SetRepos(failing, fx.hostRepo)

	putBinding(t, fx, ownerCtx())
	setCredential(t, fx, ownerCtx())
	if _, err := fx.svc.ValidateWorkspaceRepoBinding(ownerCtx(), connect.NewRequest(&agentsv1.ValidateWorkspaceRepoBindingRequest{})); err != nil {
		t.Fatalf("Validate: %v", err)
	}

	failing.fail = true
	_, err := fx.svc.SetWorkspaceRepoBindingCredential(ownerCtx(), connect.NewRequest(&agentsv1.SetWorkspaceRepoBindingCredentialRequest{Pat: "new-pat"}))
	if err == nil {
		t.Fatal("expected error from failing credential write")
	}
	got, err := fx.svc.GetWorkspaceRepoBinding(memberCtx(), connect.NewRequest(&agentsv1.GetWorkspaceRepoBindingRequest{}))
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	b := got.Msg.GetBinding()
	if b.GetStatus().GetState() != agentsv1.RepoBindingConnectionState_REPO_BINDING_CONNECTION_STATE_UNVALIDATED {
		t.Fatalf("stale status survived failed replacement: %v", b.GetStatus().GetState())
	}
	if !b.GetCredentialSet() {
		t.Fatal("old credential lost on failed replacement")
	}
	ct, err := fx.bindingRepo.GetCredential(context.Background(), "ws-a")
	if err != nil || ct == "" {
		t.Fatalf("old ciphertext gone: %q, %v", ct, err)
	}
}

// TestRepoBindingHostChangeClearsCredential proves a stored PAT is never
// forwarded to a different host: rebinding to another GitHost clears the
// credential and validation demands a fresh one.
func TestRepoBindingHostChangeClearsCredential(t *testing.T) {
	fx := newBindingFixture(t)
	if _, err := fx.hostRepo.Create(context.Background(), &agentsv1.GitHost{
		Id: "gl-1", Name: "GitLab", Kind: agentsv1.GitHostKind_GIT_HOST_KIND_GITLAB,
		ApiBaseUrl: "https://gitlab.example.com/api/v4",
	}); err != nil {
		t.Fatalf("seed second host: %v", err)
	}
	putBinding(t, fx, ownerCtx())
	setCredential(t, fx, ownerCtx())

	moved := validBinding()
	moved.GitHostId = "gl-1"
	resp, err := fx.svc.PutWorkspaceRepoBinding(ownerCtx(), connect.NewRequest(&agentsv1.PutWorkspaceRepoBindingRequest{Binding: moved}))
	if err != nil {
		t.Fatalf("Put with new host: %v", err)
	}
	if resp.Msg.GetBinding().GetCredentialSet() {
		t.Fatal("credential survived a host change")
	}
	_, err = fx.svc.ValidateWorkspaceRepoBinding(ownerCtx(), connect.NewRequest(&agentsv1.ValidateWorkspaceRepoBindingRequest{}))
	wantCode(t, err, connect.CodeFailedPrecondition)

	// Same-host re-put keeps the credential (regression guard).
	putBinding(t, fx, ownerCtx()) // back to gh-1, no credential now
	setCredential(t, fx, ownerCtx())
	again, err := fx.svc.PutWorkspaceRepoBinding(ownerCtx(), connect.NewRequest(&agentsv1.PutWorkspaceRepoBindingRequest{Binding: validBinding()}))
	if err != nil {
		t.Fatalf("same-host Put: %v", err)
	}
	if !again.Msg.GetBinding().GetCredentialSet() {
		t.Fatal("credential lost on same-host Put")
	}
}

func TestRepoBindingDelete(t *testing.T) {
	fx := newBindingFixture(t)
	putBinding(t, fx, ownerCtx())
	if _, err := fx.svc.DeleteWorkspaceRepoBinding(ownerCtx(), connect.NewRequest(&agentsv1.DeleteWorkspaceRepoBindingRequest{})); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	resp, err := fx.svc.GetWorkspaceRepoBinding(memberCtx(), connect.NewRequest(&agentsv1.GetWorkspaceRepoBindingRequest{}))
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if resp.Msg.GetBinding() != nil {
		t.Fatal("binding survived delete")
	}
	_, err = fx.svc.DeleteWorkspaceRepoBinding(ownerCtx(), connect.NewRequest(&agentsv1.DeleteWorkspaceRepoBindingRequest{}))
	wantCode(t, err, connect.CodeNotFound)
}

func TestRepoBindingRequiresWorkspaceHeader(t *testing.T) {
	fx := newBindingFixture(t)
	ctx := auth.WithAuthenticated(context.Background(), &agentsv1.User{Id: "owner-user"}, nil)
	_, err := fx.svc.GetWorkspaceRepoBinding(ctx, connect.NewRequest(&agentsv1.GetWorkspaceRepoBindingRequest{}))
	wantCode(t, err, connect.CodeFailedPrecondition)
}

// ── Sync and cache tests (issue #215) ───────────────────────────────────

func newSyncFixture(t *testing.T) *bindingFixture {
	return newSyncFixtureWithRoot(t, "")
}

func newSyncFixtureWithRoot(t *testing.T, rootPath string) *bindingFixture {
	t.Helper()
	fx := newBindingFixture(t)
	const sha = "abc123"

	prefix := rootPath
	if prefix != "" {
		prefix = strings.TrimRight(prefix, "/") + "/"
	}

	fx.fake.trees = map[string][]gitprovider.TreeEntry{
		sha + ":" + strings.TrimRight(rootPath, "/"): {
			{Path: prefix + "agents", Kind: gitprovider.TreeEntryDirectory, SHA: "tree-agents"},
			{Path: prefix + "agents/my-agent", Kind: gitprovider.TreeEntryDirectory, SHA: "tree-my-agent"},
			{Path: prefix + "agents/my-agent/prompt.md", Kind: gitprovider.TreeEntryFile, Size: 25, SHA: "blob-prompt"},
			{Path: prefix + "agents/my-agent/description.md", Kind: gitprovider.TreeEntryFile, Size: 22, SHA: "blob-desc"},
			{Path: prefix + "agents/unclaimed-dir", Kind: gitprovider.TreeEntryDirectory, SHA: "tree-unclaimed"},
			{Path: prefix + "agents/unclaimed-dir/notes.md", Kind: gitprovider.TreeEntryFile, Size: 11, SHA: "blob-notes"},
		},
	}
	fx.fake.blobs = map[string][]byte{
		sha + ":" + prefix + "agents/my-agent/prompt.md":      []byte("You are a helpful agent."),
		sha + ":" + prefix + "agents/my-agent/description.md": []byte("My agent description."),
		sha + ":" + prefix + "agents/unclaimed-dir/notes.md":  []byte("Some notes."),
	}
	fx.svc.SetCacheRepo(repocachememory.New())
	b := validBinding()
	b.RootPath = rootPath
	if _, err := fx.svc.PutWorkspaceRepoBinding(ownerCtx(), connect.NewRequest(&agentsv1.PutWorkspaceRepoBindingRequest{
		Binding: b,
	})); err != nil {
		t.Fatalf("PutWorkspaceRepoBinding: %v", err)
	}
	setCredential(t, fx, ownerCtx())
	return fx
}

func TestSyncWorkspaceRepository(t *testing.T) {
	fx := newSyncFixture(t)
	resp, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{}))
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if resp.Msg.GetEntriesSynced() == 0 {
		t.Fatal("expected entries synced > 0")
	}
	binding := resp.Msg.GetBinding()
	if binding.GetObservedCommitSha() != "abc123" {
		t.Fatalf("observed_commit_sha = %q, want abc123", binding.GetObservedCommitSha())
	}
	if binding.GetLastSyncedAt() == nil {
		t.Fatal("last_synced_at not set")
	}
	if binding.GetLastSyncError() != "" {
		t.Fatalf("last_sync_error = %q", binding.GetLastSyncError())
	}
}

func TestSyncIdempotent(t *testing.T) {
	fx := newSyncFixture(t)
	resp1, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{}))
	if err != nil {
		t.Fatalf("Sync 1: %v", err)
	}
	if resp1.Msg.GetEntriesSynced() == 0 {
		t.Fatal("first sync should populate cache")
	}

	resp2, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{}))
	if err != nil {
		t.Fatalf("Sync 2: %v", err)
	}
	if resp2.Msg.GetEntriesSynced() != 0 {
		t.Fatalf("idempotent sync should return 0 entries_synced, got %d", resp2.Msg.GetEntriesSynced())
	}
}

func TestSyncRequiresOwnerOrAdmin(t *testing.T) {
	fx := newSyncFixture(t)
	_, err := fx.svc.SyncWorkspaceRepository(memberCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{}))
	wantCode(t, err, connect.CodePermissionDenied)

	_, err = fx.svc.SyncWorkspaceRepository(wsAdminCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{}))
	if err != nil {
		t.Fatalf("admin sync: %v", err)
	}
}

func TestListRepositoryEntries(t *testing.T) {
	fx := newSyncFixture(t)
	if _, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{})); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	resp, err := fx.svc.ListRepositoryEntries(memberCtx(), connect.NewRequest(&agentsv1.ListRepositoryEntriesRequest{}))
	if err != nil {
		t.Fatalf("ListEntries root: %v", err)
	}
	if resp.Msg.GetCommitSha() != "abc123" {
		t.Fatalf("commit_sha = %q", resp.Msg.GetCommitSha())
	}
	if len(resp.Msg.GetEntries()) == 0 {
		t.Fatal("expected entries in root")
	}

	t.Run("subdirectory", func(t *testing.T) {
		resp, err := fx.svc.ListRepositoryEntries(memberCtx(), connect.NewRequest(&agentsv1.ListRepositoryEntriesRequest{
			Path: "agents",
		}))
		if err != nil {
			t.Fatalf("ListEntries agents: %v", err)
		}
		if len(resp.Msg.GetEntries()) == 0 {
			t.Fatal("expected entries under agents/")
		}
		claimed := map[string]bool{}
		for _, entry := range resp.Msg.GetEntries() {
			claimed[entry.GetPath()] = entry.GetClaimed()
		}
		if !claimed["agents/my-agent"] {
			t.Fatal("known Agent directory should be claimed")
		}
		if claimed["agents/unclaimed-dir"] {
			t.Fatal("unknown Agent directory should remain unclaimed")
		}
	})
}

func TestListRepositoryEntriesNoCache(t *testing.T) {
	fx := newSyncFixture(t)
	_, err := fx.svc.ListRepositoryEntries(memberCtx(), connect.NewRequest(&agentsv1.ListRepositoryEntriesRequest{}))
	wantCode(t, err, connect.CodeNotFound)
}

func TestGetRepositoryFile(t *testing.T) {
	fx := newSyncFixture(t)
	if _, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{})); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	resp, err := fx.svc.GetRepositoryFile(memberCtx(), connect.NewRequest(&agentsv1.GetRepositoryFileRequest{
		Path: "agents/my-agent/prompt.md",
	}))
	if err != nil {
		t.Fatalf("GetFile: %v", err)
	}
	if resp.Msg.GetCommitSha() != "abc123" {
		t.Fatalf("commit_sha = %q", resp.Msg.GetCommitSha())
	}
	if resp.Msg.GetContent() != "You are a helpful agent." {
		t.Fatalf("content = %q", resp.Msg.GetContent())
	}
	entry := resp.Msg.GetEntry()
	if entry == nil {
		t.Fatal("entry is nil")
	}
	if entry.GetPath() != "agents/my-agent/prompt.md" {
		t.Fatalf("entry path = %q", entry.GetPath())
	}
	if entry.GetKind() != agentsv1.RepoCacheEntryKind_REPO_CACHE_ENTRY_KIND_FILE {
		t.Fatalf("entry kind = %v", entry.GetKind())
	}
}

func TestGetRepositoryFileMissing(t *testing.T) {
	fx := newSyncFixture(t)
	if _, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{})); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	_, err := fx.svc.GetRepositoryFile(memberCtx(), connect.NewRequest(&agentsv1.GetRepositoryFileRequest{
		Path: "agents/nonexistent.md",
	}))
	wantCode(t, err, connect.CodeNotFound)
}

func TestGetRepositorySyncStatus(t *testing.T) {
	fx := newSyncFixture(t)
	if _, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{})); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	resp, err := fx.svc.GetRepositorySyncStatus(memberCtx(), connect.NewRequest(&agentsv1.GetRepositorySyncStatusRequest{}))
	if err != nil {
		t.Fatalf("GetSyncStatus: %v", err)
	}
	binding := resp.Msg.GetBinding()
	if binding.GetObservedCommitSha() != "abc123" {
		t.Fatalf("observed_commit_sha = %q", binding.GetObservedCommitSha())
	}
}

func TestSyncWorkspaceIsolation(t *testing.T) {
	fx := newSyncFixture(t)
	if _, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{})); err != nil {
		t.Fatalf("Sync ws-a: %v", err)
	}

	// ws-b has no binding and no cache — reads must fail.
	wsBCtx := ctxAs("owner-user", "user", "ws-b")
	for _, m := range []struct{ user, role string }{
		{"owner-user", "owner"},
	} {
		if _, err := fx.wsRepo.AddMember(context.Background(), &agentsv1.WorkspaceMember{
			WorkspaceId: "ws-b", UserId: m.user, Role: m.role,
		}); err != nil {
			t.Fatalf("seed ws-b member: %v", err)
		}
	}
	_, err := fx.svc.ListRepositoryEntries(wsBCtx, connect.NewRequest(&agentsv1.ListRepositoryEntriesRequest{}))
	wantCode(t, err, connect.CodeNotFound)
}

func TestSyncPathTraversalRejected(t *testing.T) {
	fx := newSyncFixture(t)
	if _, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{})); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	cases := []string{
		"../etc/passwd",
		"/absolute/path",
		"agents/../../../etc/passwd",
	}
	for _, p := range cases {
		t.Run(p, func(t *testing.T) {
			_, err := fx.svc.ListRepositoryEntries(memberCtx(), connect.NewRequest(&agentsv1.ListRepositoryEntriesRequest{Path: p}))
			wantCode(t, err, connect.CodeInvalidArgument)
		})
		t.Run("file/"+p, func(t *testing.T) {
			_, err := fx.svc.GetRepositoryFile(memberCtx(), connect.NewRequest(&agentsv1.GetRepositoryFileRequest{Path: p}))
			wantCode(t, err, connect.CodeInvalidArgument)
		})
	}
}

func TestSyncRecordsErrorOnBranchFailure(t *testing.T) {
	fx := newSyncFixture(t)
	fx.fake.branches = map[string]string{}
	_, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{}))
	if err == nil {
		t.Fatal("expected error when branch HEAD fails")
	}

	resp, getErr := fx.svc.GetRepositorySyncStatus(memberCtx(), connect.NewRequest(&agentsv1.GetRepositorySyncStatusRequest{}))
	if getErr != nil {
		t.Fatalf("GetSyncStatus: %v", getErr)
	}
	if resp.Msg.GetBinding().GetLastSyncError() == "" {
		t.Fatal("last_sync_error not recorded")
	}
}

func TestGetRepositoryFileRequiresPath(t *testing.T) {
	fx := newSyncFixture(t)
	if _, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{})); err != nil {
		t.Fatalf("Sync: %v", err)
	}
	_, err := fx.svc.GetRepositoryFile(memberCtx(), connect.NewRequest(&agentsv1.GetRepositoryFileRequest{}))
	wantCode(t, err, connect.CodeInvalidArgument)
}

func TestSyncWithRootPath(t *testing.T) {
	fx := newSyncFixtureWithRoot(t, "content")
	resp, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{}))
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if resp.Msg.GetEntriesSynced() == 0 {
		t.Fatal("expected entries synced > 0")
	}

	// Entries should be relative to root_path (no "content/" prefix).
	listResp, err := fx.svc.ListRepositoryEntries(memberCtx(), connect.NewRequest(&agentsv1.ListRepositoryEntriesRequest{}))
	if err != nil {
		t.Fatalf("ListEntries root: %v", err)
	}
	if len(listResp.Msg.GetEntries()) == 0 {
		t.Fatal("expected entries in root listing")
	}
	for _, e := range listResp.Msg.GetEntries() {
		if strings.HasPrefix(e.GetPath(), "content/") {
			t.Fatalf("entry path still contains root_path prefix: %q", e.GetPath())
		}
	}

	// File access uses root-relative paths.
	fileResp, err := fx.svc.GetRepositoryFile(memberCtx(), connect.NewRequest(&agentsv1.GetRepositoryFileRequest{
		Path: "agents/my-agent/prompt.md",
	}))
	if err != nil {
		t.Fatalf("GetFile: %v", err)
	}
	if fileResp.Msg.GetContent() != "You are a helpful agent." {
		t.Fatalf("content = %q", fileResp.Msg.GetContent())
	}
}

func TestSyncBranchMovement(t *testing.T) {
	fx := newSyncFixture(t)

	// First sync at abc123.
	resp1, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{}))
	if err != nil {
		t.Fatalf("Sync 1: %v", err)
	}
	if resp1.Msg.GetEntriesSynced() == 0 {
		t.Fatal("first sync should populate cache")
	}

	// Move branch to a new commit.
	newSHA := "def456"
	fx.fake.branches["main"] = newSHA
	fx.fake.trees[newSHA+":"] = []gitprovider.TreeEntry{
		{Path: "agents", Kind: gitprovider.TreeEntryDirectory, SHA: "tree-agents"},
		{Path: "agents/new-file.md", Kind: gitprovider.TreeEntryFile, Size: 8, SHA: "blob-new"},
	}
	fx.fake.blobs[newSHA+":agents/new-file.md"] = []byte("Updated!")

	resp2, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{}))
	if err != nil {
		t.Fatalf("Sync 2: %v", err)
	}
	if resp2.Msg.GetEntriesSynced() == 0 {
		t.Fatal("second sync with new SHA should populate cache")
	}
	if resp2.Msg.GetBinding().GetObservedCommitSha() != newSHA {
		t.Fatalf("observed_commit_sha = %q, want %q", resp2.Msg.GetBinding().GetObservedCommitSha(), newSHA)
	}

	// Old file should be gone, new file should be present.
	_, err = fx.svc.GetRepositoryFile(memberCtx(), connect.NewRequest(&agentsv1.GetRepositoryFileRequest{
		Path: "agents/my-agent/prompt.md",
	}))
	wantCode(t, err, connect.CodeNotFound)

	fileResp, err := fx.svc.GetRepositoryFile(memberCtx(), connect.NewRequest(&agentsv1.GetRepositoryFileRequest{
		Path: "agents/new-file.md",
	}))
	if err != nil {
		t.Fatalf("GetFile new: %v", err)
	}
	if fileResp.Msg.GetContent() != "Updated!" {
		t.Fatalf("content = %q", fileResp.Msg.GetContent())
	}
}

func TestSyncGitLabSizeEnforcement(t *testing.T) {
	fx := newBindingFixture(t)
	const sha = "abc123"

	// Simulate GitLab: tree entries have Size=0, actual blob is large.
	fx.fake.trees = map[string][]gitprovider.TreeEntry{
		sha + ":": {
			{Path: "big-file.md", Kind: gitprovider.TreeEntryFile, Size: 0, SHA: "blob-big"},
		},
	}
	bigContent := make([]byte, 2*1024*1024) // 2 MiB
	for i := range bigContent {
		bigContent[i] = 'A'
	}
	fx.fake.blobs = map[string][]byte{
		sha + ":big-file.md": bigContent,
	}
	fx.svc.SetCacheRepo(repocachememory.New())
	putBinding(t, fx, ownerCtx())
	setCredential(t, fx, ownerCtx())

	resp, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{}))
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	// The oversized file should be skipped but the sync should succeed.
	if resp.Msg.GetBinding().GetLastSyncError() != "" {
		t.Fatalf("unexpected sync error: %q", resp.Msg.GetBinding().GetLastSyncError())
	}

	// The large file's blob should NOT be cached.
	_, err = fx.svc.GetRepositoryFile(memberCtx(), connect.NewRequest(&agentsv1.GetRepositoryFileRequest{
		Path: "big-file.md",
	}))
	wantCode(t, err, connect.CodeNotFound)
}

func TestBindingInvalidationOnUpdate(t *testing.T) {
	fx := newSyncFixture(t)
	if _, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{})); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	// Confirm cache is populated.
	_, err := fx.svc.ListRepositoryEntries(memberCtx(), connect.NewRequest(&agentsv1.ListRepositoryEntriesRequest{}))
	if err != nil {
		t.Fatalf("ListEntries before update: %v", err)
	}

	// Update binding (changes repo config).
	b := validBinding()
	b.RootPath = "new-root"
	if _, err := fx.svc.PutWorkspaceRepoBinding(ownerCtx(), connect.NewRequest(&agentsv1.PutWorkspaceRepoBindingRequest{Binding: b})); err != nil {
		t.Fatalf("Put: %v", err)
	}

	// Cache should be invalidated.
	_, err = fx.svc.ListRepositoryEntries(memberCtx(), connect.NewRequest(&agentsv1.ListRepositoryEntriesRequest{}))
	wantCode(t, err, connect.CodeNotFound)
}

type deleteFailingCache struct {
	repocache.Repository
}

func (c *deleteFailingCache) Delete(context.Context, string) error {
	return errors.New("cache delete failed")
}

func TestBindingUpdateCannotReadOldCacheWhenInvalidationFails(t *testing.T) {
	fx := newSyncFixture(t)
	if _, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{})); err != nil {
		t.Fatalf("Sync: %v", err)
	}
	failing := &deleteFailingCache{Repository: fx.svc.cacheRepo}
	fx.svc.SetCacheRepo(failing)

	binding := validBinding()
	binding.RootPath = "new-root"
	if _, err := fx.svc.PutWorkspaceRepoBinding(ownerCtx(), connect.NewRequest(&agentsv1.PutWorkspaceRepoBindingRequest{Binding: binding})); err != nil {
		t.Fatalf("Put: %v", err)
	}

	_, err := fx.svc.ListRepositoryEntries(memberCtx(), connect.NewRequest(&agentsv1.ListRepositoryEntriesRequest{}))
	wantCode(t, err, connect.CodeNotFound)
}

func TestSyncWorkspaceLimitDoesNotPublishPartialSnapshot(t *testing.T) {
	fx := newSyncFixture(t)
	if _, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{})); err != nil {
		t.Fatalf("initial Sync: %v", err)
	}

	const newSHA = "limit456"
	fx.fake.branches["main"] = newSHA
	fx.fake.trees[newSHA+":"] = []gitprovider.TreeEntry{
		{Path: "agents", Kind: gitprovider.TreeEntryDirectory, SHA: "tree-agents"},
		{Path: "agents/my-agent", Kind: gitprovider.TreeEntryDirectory, SHA: "tree-agent"},
		{Path: "agents/my-agent/a.md", Kind: gitprovider.TreeEntryFile, Size: 8, SHA: "blob-a"},
		{Path: "agents/my-agent/b.md", Kind: gitprovider.TreeEntryFile, Size: 8, SHA: "blob-b"},
	}
	fx.fake.blobs[newSHA+":agents/my-agent/a.md"] = []byte("12345678")
	fx.fake.blobs[newSHA+":agents/my-agent/b.md"] = []byte("abcdefgh")
	fx.svc.SetCacheLimitsProvider(func() (int64, int64) { return 1024, 10 })

	_, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{}))
	wantCode(t, err, connect.CodeFailedPrecondition)

	oldFile, getErr := fx.svc.GetRepositoryFile(memberCtx(), connect.NewRequest(&agentsv1.GetRepositoryFileRequest{
		Path: "agents/my-agent/prompt.md",
	}))
	if getErr != nil {
		t.Fatalf("old complete snapshot should remain readable: %v", getErr)
	}
	if oldFile.Msg.GetCommitSha() != "abc123" {
		t.Fatalf("cache commit = %q, want abc123", oldFile.Msg.GetCommitSha())
	}
	if oldFile.Msg.GetObservedCommitSha() != newSHA {
		t.Fatalf("observed commit = %q, want %q", oldFile.Msg.GetObservedCommitSha(), newSHA)
	}
	_, getErr = fx.svc.GetRepositoryFile(memberCtx(), connect.NewRequest(&agentsv1.GetRepositoryFileRequest{
		Path: "agents/my-agent/a.md",
	}))
	wantCode(t, getErr, connect.CodeNotFound)
}

func TestBindingInvalidationOnDelete(t *testing.T) {
	fx := newSyncFixture(t)
	if _, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{})); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	if _, err := fx.svc.DeleteWorkspaceRepoBinding(ownerCtx(), connect.NewRequest(&agentsv1.DeleteWorkspaceRepoBindingRequest{})); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// List requires a binding, so it should fail with NotFound (no binding).
	_, err := fx.svc.ListRepositoryEntries(memberCtx(), connect.NewRequest(&agentsv1.ListRepositoryEntriesRequest{}))
	wantCode(t, err, connect.CodeNotFound)
}

func TestListRepositoryEntriesResponseSHAs(t *testing.T) {
	fx := newSyncFixture(t)
	if _, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{})); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	resp, err := fx.svc.ListRepositoryEntries(memberCtx(), connect.NewRequest(&agentsv1.ListRepositoryEntriesRequest{}))
	if err != nil {
		t.Fatalf("ListEntries: %v", err)
	}
	if resp.Msg.GetObservedCommitSha() != "abc123" {
		t.Fatalf("observed_commit_sha = %q, want abc123", resp.Msg.GetObservedCommitSha())
	}
	if resp.Msg.GetActiveCommitSha() != "" {
		t.Fatalf("active_commit_sha = %q, want empty before publication", resp.Msg.GetActiveCommitSha())
	}
}

func TestGetRepositoryFileResponseSHAs(t *testing.T) {
	fx := newSyncFixture(t)
	if _, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{})); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	resp, err := fx.svc.GetRepositoryFile(memberCtx(), connect.NewRequest(&agentsv1.GetRepositoryFileRequest{
		Path: "agents/my-agent/prompt.md",
	}))
	if err != nil {
		t.Fatalf("GetFile: %v", err)
	}
	if resp.Msg.GetObservedCommitSha() != "abc123" {
		t.Fatalf("observed_commit_sha = %q, want abc123", resp.Msg.GetObservedCommitSha())
	}
	if resp.Msg.GetActiveCommitSha() != "" {
		t.Fatalf("active_commit_sha = %q, want empty before publication", resp.Msg.GetActiveCommitSha())
	}
}

// ── Publication tests (issue #216) ──────────────────────────────────────

// fakeConfigRuntime records ReloadRunner calls.
type fakeConfigRuntime struct {
	reloadCount int
	reloadErr   error
}

func (f *fakeConfigRuntime) ReloadRunner(_ context.Context) error {
	f.reloadCount++
	return f.reloadErr
}
func (f *fakeConfigRuntime) ReloadChannels(context.Context) error { return nil }

func newPublicationFixture(t *testing.T) (*bindingFixture, *fakeConfigRuntime) {
	t.Helper()
	fx := newSyncFixture(t)
	rt := &fakeConfigRuntime{}
	fx.svc.SetContentRepo(agentcontentmemory.New())
	fx.svc.SetConfigRuntime(rt)
	return fx, rt
}

func TestPublishSyncPublishesActiveRevision(t *testing.T) {
	fx, rt := newPublicationFixture(t)

	resp, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{}))
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if !resp.Msg.GetPublished() {
		t.Fatal("sync should have published")
	}
	if len(resp.Msg.GetPublicationErrors()) != 0 {
		t.Fatalf("unexpected publication errors: %v", resp.Msg.GetPublicationErrors())
	}
	binding := resp.Msg.GetBinding()
	if binding.GetActiveCommitSha() != "abc123" {
		t.Fatalf("active_commit_sha = %q, want abc123", binding.GetActiveCommitSha())
	}
	if rt.reloadCount != 1 {
		t.Fatalf("reload count = %d, want 1", rt.reloadCount)
	}
}

func TestPublishExplicitRPC(t *testing.T) {
	fx, _ := newPublicationFixture(t)

	// Sync first to populate cache.
	if _, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{})); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	// Explicit publish should be idempotent.
	pubResp, err := fx.svc.PublishWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.PublishWorkspaceRepositoryRequest{}))
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if pubResp.Msg.GetBinding().GetActiveCommitSha() != "abc123" {
		t.Fatalf("active_commit_sha = %q", pubResp.Msg.GetBinding().GetActiveCommitSha())
	}
}

func TestPublishLLMAgentWithoutPromptFails(t *testing.T) {
	fx, _ := newPublicationFixture(t)

	// Register an LLM agent that has no instruction.
	if _, err := fx.agentRepo.CreateAgent(context.Background(), "ws-a", &agentsv1.Agent{
		Name:    "LLM Agent",
		AgentId: "llm-agent",
		Type:    agentsv1.AgentType_AGENT_TYPE_LLM,
		Config:  &agentsv1.AgentConfig{},
	}); err != nil {
		t.Fatalf("seed llm-agent: %v", err)
	}

	// Provide agents/llm-agent directory with only description.md (no prompt).
	fx.fake.trees["abc123:"] = append(fx.fake.trees["abc123:"],
		gitprovider.TreeEntry{Path: "agents/llm-agent", Kind: gitprovider.TreeEntryDirectory, SHA: "tree-llm"},
		gitprovider.TreeEntry{Path: "agents/llm-agent/description.md", Kind: gitprovider.TreeEntryFile, Size: 10, SHA: "blob-llm-desc"},
	)
	fx.fake.blobs["abc123:agents/llm-agent/description.md"] = []byte("LLM desc.")

	resp, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{}))
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if resp.Msg.GetPublished() {
		t.Fatal("should not publish with validation errors")
	}
	if len(resp.Msg.GetPublicationErrors()) == 0 {
		t.Fatal("expected validation errors for LLM agent without prompt")
	}
	if resp.Msg.GetBinding().GetActiveCommitSha() != "" {
		t.Fatalf("active_commit_sha should be empty, got %q", resp.Msg.GetBinding().GetActiveCommitSha())
	}
}

func TestPublishFileResponseShowsActiveSHA(t *testing.T) {
	fx, _ := newPublicationFixture(t)

	if _, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{})); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	resp, err := fx.svc.GetRepositoryFile(memberCtx(), connect.NewRequest(&agentsv1.GetRepositoryFileRequest{
		Path: "agents/my-agent/prompt.md",
	}))
	if err != nil {
		t.Fatalf("GetFile: %v", err)
	}
	if resp.Msg.GetActiveCommitSha() != "abc123" {
		t.Fatalf("active_commit_sha = %q, want abc123", resp.Msg.GetActiveCommitSha())
	}
}

func TestPublishInvalidKeepsLastGood(t *testing.T) {
	fx, _ := newPublicationFixture(t)

	// First sync + publish succeeds.
	if _, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{})); err != nil {
		t.Fatalf("Sync 1: %v", err)
	}

	// Add an LLM agent and advance branch.
	if _, err := fx.agentRepo.CreateAgent(context.Background(), "ws-a", &agentsv1.Agent{
		Name:    "Broken LLM",
		AgentId: "broken-llm",
		Type:    agentsv1.AgentType_AGENT_TYPE_LLM,
	}); err != nil {
		t.Fatalf("seed broken-llm: %v", err)
	}
	newSHA := "invalid456"
	fx.fake.branches["main"] = newSHA
	fx.fake.trees[newSHA+":"] = []gitprovider.TreeEntry{
		{Path: "agents", Kind: gitprovider.TreeEntryDirectory, SHA: "tree-agents"},
		{Path: "agents/broken-llm", Kind: gitprovider.TreeEntryDirectory, SHA: "tree-broken"},
		{Path: "agents/broken-llm/description.md", Kind: gitprovider.TreeEntryFile, Size: 5, SHA: "blob-broken"},
	}
	fx.fake.blobs[newSHA+":agents/broken-llm/description.md"] = []byte("Desc.")

	resp, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{}))
	if err != nil {
		t.Fatalf("Sync 2: %v", err)
	}
	if resp.Msg.GetPublished() {
		t.Fatal("should not publish invalid revision")
	}
	if resp.Msg.GetBinding().GetActiveCommitSha() != "abc123" {
		t.Fatalf("active should stay at abc123, got %q", resp.Msg.GetBinding().GetActiveCommitSha())
	}
	if resp.Msg.GetBinding().GetObservedCommitSha() != newSHA {
		t.Fatalf("observed should advance to %q, got %q", newSHA, resp.Msg.GetBinding().GetObservedCommitSha())
	}
}

// ── Webhook tests (issue #216) ──────────────────────────────────────────

func TestConfigureWebhookSecret(t *testing.T) {
	fx, _ := newPublicationFixture(t)
	fx.svc.SetWebhookBaseURL(func() string { return "https://butter.example.com" })
	putBinding(t, fx, ownerCtx())
	setCredential(t, fx, ownerCtx())

	resp, err := fx.svc.ConfigureWebhookSecret(ownerCtx(), connect.NewRequest(&agentsv1.ConfigureWebhookSecretRequest{}))
	if err != nil {
		t.Fatalf("ConfigureWebhookSecret: %v", err)
	}
	if resp.Msg.GetWebhookSecret() == "" {
		t.Fatal("webhook secret is empty")
	}
	if resp.Msg.GetCallbackUrl() != "https://butter.example.com/api/webhooks/repository/ws-a" {
		t.Fatalf("callback_url = %q", resp.Msg.GetCallbackUrl())
	}
	if !resp.Msg.GetBinding().GetWebhookSecretSet() {
		t.Fatal("webhook_secret_set not reported")
	}

	// Verify HMAC signature against configured secret.
	body := []byte(`{"ref":"refs/heads/main"}`)
	mac := hmac.New(sha256.New, []byte(resp.Msg.GetWebhookSecret()))
	mac.Write(body)
	sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if !fx.svc.VerifyWebhookSignature(context.Background(), "ws-a", body, sig, "") {
		t.Fatal("valid GitHub HMAC should verify")
	}
	if fx.svc.VerifyWebhookSignature(context.Background(), "ws-a", body, "sha256=bad", "") {
		t.Fatal("bad HMAC should fail")
	}

	// Verify GitLab-style token verification.
	if !fx.svc.VerifyWebhookSignature(context.Background(), "ws-a", body, "", resp.Msg.GetWebhookSecret()) {
		t.Fatal("valid GitLab token should verify")
	}
}

// ── DEGRADED / DIVERGED tests (issue #216) ──────────────────────────────

func TestSyncDegradedOnProviderFailure(t *testing.T) {
	fx := newBindingFixture(t)
	fx.svc.SetCacheRepo(repocachememory.New())
	fx.svc.SetContentRepo(agentcontentmemory.New())
	putBinding(t, fx, ownerCtx())
	setCredential(t, fx, ownerCtx())

	// Simulate a revoked PAT: GetBranchHead returns ErrUnauthorized.
	fx.fake.branchErr = gitprovider.ErrUnauthorized
	_, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{}))
	if err == nil {
		t.Fatal("expected error from unauthorized provider")
	}

	got, getErr := fx.svc.GetWorkspaceRepoBinding(memberCtx(), connect.NewRequest(&agentsv1.GetWorkspaceRepoBindingRequest{}))
	if getErr != nil {
		t.Fatalf("Get: %v", getErr)
	}
	state := got.Msg.GetBinding().GetStatus().GetState()
	if state != agentsv1.RepoBindingConnectionState_REPO_BINDING_CONNECTION_STATE_DEGRADED {
		t.Fatalf("state = %v, want DEGRADED", state)
	}
}

func TestSyncDivergedOnNonFastForward(t *testing.T) {
	fx, _ := newPublicationFixture(t)

	// First sync establishes observed SHA.
	resp, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{}))
	if err != nil {
		t.Fatalf("Sync 1: %v", err)
	}
	if resp.Msg.GetBinding().GetObservedCommitSha() != "abc123" {
		t.Fatalf("observed = %q", resp.Msg.GetBinding().GetObservedCommitSha())
	}

	// Branch moves with a force-push (non-fast-forward).
	fx.fake.branches["main"] = "force-pushed-sha"
	fx.fake.comparisons = map[string]string{
		"abc123...force-pushed-sha": "diverged",
	}
	fx.fake.trees["force-pushed-sha:"] = []gitprovider.TreeEntry{
		{Path: "agents", Kind: gitprovider.TreeEntryDirectory, SHA: "tree-agents"},
	}

	resp2, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{}))
	if err != nil {
		t.Fatalf("Sync 2: %v", err)
	}
	state := resp2.Msg.GetBinding().GetStatus().GetState()
	if state != agentsv1.RepoBindingConnectionState_REPO_BINDING_CONNECTION_STATE_DIVERGED {
		t.Fatalf("state = %v, want DIVERGED", state)
	}
}

func TestAcceptRepositoryBaseline(t *testing.T) {
	fx, _ := newPublicationFixture(t)

	// First sync establishes observed SHA.
	if _, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{})); err != nil {
		t.Fatalf("Sync 1: %v", err)
	}

	// Force divergence.
	fx.fake.branches["main"] = "force-pushed-sha"
	fx.fake.comparisons = map[string]string{
		"abc123...force-pushed-sha": "diverged",
	}
	fx.fake.trees["force-pushed-sha:"] = []gitprovider.TreeEntry{
		{Path: "agents", Kind: gitprovider.TreeEntryDirectory, SHA: "tree-agents"},
		{Path: "agents/my-agent", Kind: gitprovider.TreeEntryDirectory, SHA: "tree-agent"},
		{Path: "agents/my-agent/prompt.md", Kind: gitprovider.TreeEntryFile, Size: 12, SHA: "blob-new"},
	}
	fx.fake.blobs["force-pushed-sha:agents/my-agent/prompt.md"] = []byte("New prompt!!")

	if _, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{})); err != nil {
		t.Fatalf("Sync 2: %v", err)
	}

	// Accept baseline resets DIVERGED and re-syncs.
	fx.fake.comparisons = nil
	resp, err := fx.svc.AcceptRepositoryBaseline(ownerCtx(), connect.NewRequest(&agentsv1.AcceptRepositoryBaselineRequest{}))
	if err != nil {
		t.Fatalf("AcceptRepositoryBaseline: %v", err)
	}
	state := resp.Msg.GetBinding().GetStatus().GetState()
	if state == agentsv1.RepoBindingConnectionState_REPO_BINDING_CONNECTION_STATE_DIVERGED {
		t.Fatal("state should no longer be DIVERGED after acceptance")
	}
}

func TestAcceptRepositoryBaselineRequiresDivergedState(t *testing.T) {
	fx, _ := newPublicationFixture(t)
	putBinding(t, fx, ownerCtx())

	_, err := fx.svc.AcceptRepositoryBaseline(ownerCtx(), connect.NewRequest(&agentsv1.AcceptRepositoryBaselineRequest{}))
	wantCode(t, err, connect.CodeFailedPrecondition)
}

// ── TriggerSyncAndPublish tests (issue #216) ────────────────────────────

func TestTriggerSyncAndPublish(t *testing.T) {
	fx, rt := newPublicationFixture(t)

	err := fx.svc.TriggerSyncAndPublish(context.Background(), "ws-a")
	if err != nil {
		t.Fatalf("TriggerSyncAndPublish: %v", err)
	}
	if rt.reloadCount != 1 {
		t.Fatalf("reload count = %d, want 1", rt.reloadCount)
	}

	binding, err := fx.bindingRepo.Get(context.Background(), "ws-a")
	if err != nil {
		t.Fatalf("Get binding: %v", err)
	}
	if binding.GetActiveCommitSha() != "abc123" {
		t.Fatalf("active_commit_sha = %q, want abc123", binding.GetActiveCommitSha())
	}
}

// ── Agent Content editing tests (issue #217) ────────────────────────────

func newContentEditFixture(t *testing.T) (*bindingFixture, *fakeConfigRuntime) {
	t.Helper()
	fx, rt := newPublicationFixture(t)
	// Sync first so there's active content.
	if _, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{})); err != nil {
		t.Fatalf("initial sync: %v", err)
	}
	return fx, rt
}

func TestCommitAgentContentDirectCommit(t *testing.T) {
	fx, _ := newContentEditFixture(t)

	resp, err := fx.svc.CommitAgentContent(ownerCtx(), connect.NewRequest(&agentsv1.CommitAgentContentRequest{
		Actions: []*agentsv1.ContentFileAction{
			{
				Path:      "agents/my-agent/prompt.md",
				Operation: agentsv1.ContentFileOperation_CONTENT_FILE_OPERATION_PUT,
				Content:   "Updated instruction.",
			},
		},
		Message:      "update my-agent prompt",
		BaseRevision: "abc123",
	}))
	if err != nil {
		t.Fatalf("CommitAgentContent: %v", err)
	}
	if resp.Msg.GetCommitSha() == "" {
		t.Fatal("expected non-empty commit SHA")
	}
	if resp.Msg.GetChangeRequestUrl() != "" {
		t.Fatalf("unexpected change request URL for direct commit: %q", resp.Msg.GetChangeRequestUrl())
	}
	if len(resp.Msg.GetValidationErrors()) > 0 {
		t.Fatalf("unexpected validation errors: %v", resp.Msg.GetValidationErrors())
	}

	fx.fake.mu.Lock()
	defer fx.fake.mu.Unlock()
	if len(fx.fake.commits) == 0 {
		t.Fatal("no commit recorded")
	}
	lastCommit := fx.fake.commits[len(fx.fake.commits)-1]
	if !strings.Contains(lastCommit.message, "update my-agent prompt") {
		t.Errorf("commit message = %q", lastCommit.message)
	}
}

func TestCommitAgentContentAuditMetadata(t *testing.T) {
	fx, _ := newContentEditFixture(t)

	resp, err := fx.svc.CommitAgentContent(ownerCtx(), connect.NewRequest(&agentsv1.CommitAgentContentRequest{
		Actions: []*agentsv1.ContentFileAction{
			{
				Path:      "agents/my-agent/description.md",
				Operation: agentsv1.ContentFileOperation_CONTENT_FILE_OPERATION_PUT,
				Content:   "New description.",
			},
		},
		BaseRevision: "old-sha-123",
	}))
	if err != nil {
		t.Fatalf("CommitAgentContent: %v", err)
	}

	fx.fake.mu.Lock()
	defer fx.fake.mu.Unlock()
	lastCommit := fx.fake.commits[len(fx.fake.commits)-1]
	if !strings.Contains(lastCommit.message, "Butter-Actor: owner-user") {
		t.Errorf("commit message missing actor: %q", lastCommit.message)
	}
	if !strings.Contains(lastCommit.message, "Butter-Workspace: ws-a") {
		t.Errorf("commit message missing workspace: %q", lastCommit.message)
	}
	if !strings.Contains(lastCommit.message, "Butter-Operation: commit") {
		t.Errorf("commit message missing operation: %q", lastCommit.message)
	}
	if !strings.Contains(lastCommit.message, "Butter-Base-SHA: old-sha-123") {
		t.Errorf("commit message missing base SHA: %q", lastCommit.message)
	}
	_ = resp
}

func TestCommitAgentContentDeleteAction(t *testing.T) {
	fx, _ := newContentEditFixture(t)

	resp, err := fx.svc.CommitAgentContent(ownerCtx(), connect.NewRequest(&agentsv1.CommitAgentContentRequest{
		Actions: []*agentsv1.ContentFileAction{
			{
				Path:      "agents/my-agent/description.md",
				Operation: agentsv1.ContentFileOperation_CONTENT_FILE_OPERATION_DELETE,
			},
		},
	}))
	if err != nil {
		t.Fatalf("CommitAgentContent: %v", err)
	}
	if resp.Msg.GetCommitSha() == "" {
		t.Fatal("expected non-empty commit SHA")
	}

	fx.fake.mu.Lock()
	defer fx.fake.mu.Unlock()
	lastCommit := fx.fake.commits[len(fx.fake.commits)-1]
	foundDelete := false
	for _, a := range lastCommit.actions {
		if a.Delete && strings.HasSuffix(a.Path, "agents/my-agent/description.md") {
			foundDelete = true
		}
	}
	if !foundDelete {
		t.Fatal("expected a delete action for description.md")
	}
}

func TestCommitAgentContentMultipleActions(t *testing.T) {
	fx, _ := newContentEditFixture(t)

	resp, err := fx.svc.CommitAgentContent(ownerCtx(), connect.NewRequest(&agentsv1.CommitAgentContentRequest{
		Actions: []*agentsv1.ContentFileAction{
			{
				Path:      "agents/my-agent/prompt.md",
				Operation: agentsv1.ContentFileOperation_CONTENT_FILE_OPERATION_PUT,
				Content:   "New prompt.",
			},
			{
				Path:      "agents/my-agent/description.md",
				Operation: agentsv1.ContentFileOperation_CONTENT_FILE_OPERATION_PUT,
				Content:   "New description.",
			},
			{
				Path:      "agents/my-agent/global-prompt.md",
				Operation: agentsv1.ContentFileOperation_CONTENT_FILE_OPERATION_DELETE,
			},
		},
	}))
	if err != nil {
		t.Fatalf("CommitAgentContent: %v", err)
	}
	if resp.Msg.GetCommitSha() == "" {
		t.Fatal("expected commit SHA")
	}

	fx.fake.mu.Lock()
	defer fx.fake.mu.Unlock()
	lastCommit := fx.fake.commits[len(fx.fake.commits)-1]
	if len(lastCommit.actions) != 3 {
		t.Fatalf("expected 3 actions, got %d", len(lastCommit.actions))
	}
}

func TestCommitAgentContentPermissions(t *testing.T) {
	fx, _ := newContentEditFixture(t)
	actions := []*agentsv1.ContentFileAction{
		{
			Path:      "agents/my-agent/prompt.md",
			Operation: agentsv1.ContentFileOperation_CONTENT_FILE_OPERATION_PUT,
			Content:   "test.",
		},
	}

	// Members cannot commit.
	_, err := fx.svc.CommitAgentContent(memberCtx(), connect.NewRequest(&agentsv1.CommitAgentContentRequest{
		Actions: actions,
	}))
	wantCode(t, err, connect.CodePermissionDenied)

	// Admins can commit.
	_, err = fx.svc.CommitAgentContent(wsAdminCtx(), connect.NewRequest(&agentsv1.CommitAgentContentRequest{
		Actions: actions,
	}))
	if err != nil {
		t.Fatalf("admin commit: %v", err)
	}
}

func TestCommitAgentContentPathValidation(t *testing.T) {
	fx, _ := newContentEditFixture(t)

	cases := []struct {
		name string
		path string
	}{
		{"outside agents subtree", "config/settings.md"},
		{"not markdown", "agents/my-agent/config.yaml"},
		{"path traversal", "agents/../etc/passwd.md"},
		{"absolute path", "/agents/my-agent/prompt.md"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := fx.svc.CommitAgentContent(ownerCtx(), connect.NewRequest(&agentsv1.CommitAgentContentRequest{
				Actions: []*agentsv1.ContentFileAction{
					{
						Path:      tc.path,
						Operation: agentsv1.ContentFileOperation_CONTENT_FILE_OPERATION_PUT,
						Content:   "test",
					},
				},
			}))
			wantCode(t, err, connect.CodeInvalidArgument)
		})
	}
}

func TestCommitAgentContentValidationErrors(t *testing.T) {
	fx, _ := newContentEditFixture(t)

	// Deleting prompt.md from an LLM agent should cause validation failure.
	resp, err := fx.svc.CommitAgentContent(ownerCtx(), connect.NewRequest(&agentsv1.CommitAgentContentRequest{
		Actions: []*agentsv1.ContentFileAction{
			{
				Path:      "agents/my-agent/prompt.md",
				Operation: agentsv1.ContentFileOperation_CONTENT_FILE_OPERATION_DELETE,
			},
		},
	}))
	if err != nil {
		t.Fatalf("CommitAgentContent: %v", err)
	}
	if len(resp.Msg.GetValidationErrors()) == 0 {
		t.Fatal("expected validation errors for deleting LLM agent prompt")
	}
	if resp.Msg.GetCommitSha() != "" {
		t.Fatal("commit SHA should be empty when validation fails")
	}
}

func TestCommitAgentContentChangeRequestMode(t *testing.T) {
	fx, _ := newContentEditFixture(t)

	// Switch binding to CHANGE_REQUEST mode.
	b := validBinding()
	b.WriteMode = agentsv1.RepoBindingWriteMode_REPO_BINDING_WRITE_MODE_CHANGE_REQUEST
	if _, err := fx.svc.PutWorkspaceRepoBinding(ownerCtx(), connect.NewRequest(&agentsv1.PutWorkspaceRepoBindingRequest{
		Binding: b,
	})); err != nil {
		t.Fatalf("Put binding: %v", err)
	}
	setCredential(t, fx, ownerCtx())
	// Re-sync to populate the cache.
	if _, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{})); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	resp, err := fx.svc.CommitAgentContent(ownerCtx(), connect.NewRequest(&agentsv1.CommitAgentContentRequest{
		Actions: []*agentsv1.ContentFileAction{
			{
				Path:      "agents/my-agent/prompt.md",
				Operation: agentsv1.ContentFileOperation_CONTENT_FILE_OPERATION_PUT,
				Content:   "Updated prompt.",
			},
		},
	}))
	if err != nil {
		t.Fatalf("CommitAgentContent: %v", err)
	}
	if resp.Msg.GetChangeRequestUrl() == "" {
		t.Fatal("expected change request URL for CHANGE_REQUEST mode")
	}
	if resp.Msg.GetCommitSha() == "" {
		t.Fatal("expected commit SHA")
	}

	fx.fake.mu.Lock()
	defer fx.fake.mu.Unlock()
	if len(fx.fake.createdCRs) == 0 {
		t.Fatal("no change request was created")
	}
	cr := fx.fake.createdCRs[0]
	if !strings.HasPrefix(cr.source, "butter/content-") {
		t.Errorf("branch name = %q, expected butter/content- prefix", cr.source)
	}
	if cr.target != "main" {
		t.Errorf("target branch = %q, expected main", cr.target)
	}
}

// When the change request fails to open after the work branch and commit were
// created, the work branch must be cleaned up so retries do not accumulate
// orphaned branches.
func TestCommitAgentContentChangeRequestCleanupOnFailure(t *testing.T) {
	fx, _ := newContentEditFixture(t)

	b := validBinding()
	b.WriteMode = agentsv1.RepoBindingWriteMode_REPO_BINDING_WRITE_MODE_CHANGE_REQUEST
	if _, err := fx.svc.PutWorkspaceRepoBinding(ownerCtx(), connect.NewRequest(&agentsv1.PutWorkspaceRepoBindingRequest{
		Binding: b,
	})); err != nil {
		t.Fatalf("Put binding: %v", err)
	}
	setCredential(t, fx, ownerCtx())
	if _, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{})); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	fx.fake.createCRErr = errors.New("boom: change request rejected")

	_, err := fx.svc.CommitAgentContent(ownerCtx(), connect.NewRequest(&agentsv1.CommitAgentContentRequest{
		Actions: []*agentsv1.ContentFileAction{
			{
				Path:      "agents/my-agent/prompt.md",
				Operation: agentsv1.ContentFileOperation_CONTENT_FILE_OPERATION_PUT,
				Content:   "Updated prompt.",
			},
		},
	}))
	if err == nil {
		t.Fatal("expected error when change request fails to open")
	}
	if connect.CodeOf(err) != connect.CodeInternal {
		t.Errorf("code = %v, want Internal", connect.CodeOf(err))
	}

	fx.fake.mu.Lock()
	defer fx.fake.mu.Unlock()
	// The commit was created on a work branch; that branch must have been
	// deleted, and no branch should remain behind.
	if len(fx.fake.deletedBranches) != 1 {
		t.Fatalf("deleted branches = %v, want exactly one cleanup", fx.fake.deletedBranches)
	}
	deleted := fx.fake.deletedBranches[0]
	if !strings.HasPrefix(deleted, "butter/content-") {
		t.Errorf("deleted branch = %q, expected butter/content- prefix", deleted)
	}
	if _, stillThere := fx.fake.createdBranches[deleted]; stillThere {
		t.Errorf("work branch %q was not removed from created set", deleted)
	}
}

func TestCommitAgentContentWithRootPath(t *testing.T) {
	fx := newSyncFixtureWithRoot(t, "content")
	fx.svc.SetContentRepo(agentcontentmemory.New())
	fx.svc.SetConfigRuntime(&fakeConfigRuntime{})
	if _, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{})); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	resp, err := fx.svc.CommitAgentContent(ownerCtx(), connect.NewRequest(&agentsv1.CommitAgentContentRequest{
		Actions: []*agentsv1.ContentFileAction{
			{
				Path:      "agents/my-agent/prompt.md",
				Operation: agentsv1.ContentFileOperation_CONTENT_FILE_OPERATION_PUT,
				Content:   "Updated prompt.",
			},
		},
	}))
	if err != nil {
		t.Fatalf("CommitAgentContent: %v", err)
	}
	if resp.Msg.GetCommitSha() == "" {
		t.Fatal("expected commit SHA")
	}

	fx.fake.mu.Lock()
	defer fx.fake.mu.Unlock()
	lastCommit := fx.fake.commits[len(fx.fake.commits)-1]
	for _, a := range lastCommit.actions {
		if !strings.HasPrefix(a.Path, "content/") {
			t.Errorf("action path %q should be prefixed with root_path", a.Path)
		}
	}
}

func TestCommitAgentContentLastWriteWins(t *testing.T) {
	fx, _ := newContentEditFixture(t)

	// Move branch head to simulate concurrent change.
	fx.fake.branches["main"] = "concurrent-sha-999"

	resp, err := fx.svc.CommitAgentContent(ownerCtx(), connect.NewRequest(&agentsv1.CommitAgentContentRequest{
		Actions: []*agentsv1.ContentFileAction{
			{
				Path:      "agents/my-agent/prompt.md",
				Operation: agentsv1.ContentFileOperation_CONTENT_FILE_OPERATION_PUT,
				Content:   "Updated prompt.",
			},
		},
		BaseRevision: "abc123",
	}))
	if err != nil {
		t.Fatalf("CommitAgentContent: %v", err)
	}

	// The commit should have been made against the latest HEAD, not the
	// stale base revision.
	fx.fake.mu.Lock()
	defer fx.fake.mu.Unlock()
	lastCommit := fx.fake.commits[len(fx.fake.commits)-1]
	if lastCommit.parentSHA != "concurrent-sha-999" {
		t.Fatalf("parent SHA = %q, want concurrent-sha-999 (last-write-wins)", lastCommit.parentSHA)
	}
	_ = resp
}

func TestCommitAgentContentFileSizeLimit(t *testing.T) {
	fx, _ := newContentEditFixture(t)
	fx.svc.SetCacheLimitsProvider(func() (int64, int64) { return 100, 1024 * 1024 })

	bigContent := strings.Repeat("A", 200)
	_, err := fx.svc.CommitAgentContent(ownerCtx(), connect.NewRequest(&agentsv1.CommitAgentContentRequest{
		Actions: []*agentsv1.ContentFileAction{
			{
				Path:      "agents/my-agent/prompt.md",
				Operation: agentsv1.ContentFileOperation_CONTENT_FILE_OPERATION_PUT,
				Content:   bigContent,
			},
		},
	}))
	wantCode(t, err, connect.CodeInvalidArgument)
}

func TestCommitAgentContentEmptyActions(t *testing.T) {
	fx, _ := newContentEditFixture(t)

	_, err := fx.svc.CommitAgentContent(ownerCtx(), connect.NewRequest(&agentsv1.CommitAgentContentRequest{}))
	wantCode(t, err, connect.CodeInvalidArgument)
}

// ── Rollback tests (issue #217) ─────────────────────────────────────────

func TestRollbackAgentContent(t *testing.T) {
	fx, _ := newContentEditFixture(t)

	// Set up a target revision with content.
	targetSHA := "rollback-target-sha"
	fx.fake.trees[targetSHA+":"] = []gitprovider.TreeEntry{
		{Path: "agents", Kind: gitprovider.TreeEntryDirectory, SHA: "tree-agents"},
		{Path: "agents/my-agent", Kind: gitprovider.TreeEntryDirectory, SHA: "tree-my-agent"},
		{Path: "agents/my-agent/prompt.md", Kind: gitprovider.TreeEntryFile, Size: 20, SHA: "blob-old-prompt"},
		{Path: "agents/my-agent/description.md", Kind: gitprovider.TreeEntryFile, Size: 15, SHA: "blob-old-desc"},
	}
	fx.fake.blobs[targetSHA+":agents/my-agent/prompt.md"] = []byte("Old prompt content.")
	fx.fake.blobs[targetSHA+":agents/my-agent/description.md"] = []byte("Old description.")

	resp, err := fx.svc.RollbackAgentContent(ownerCtx(), connect.NewRequest(&agentsv1.RollbackAgentContentRequest{
		TargetCommitSha: targetSHA,
	}))
	if err != nil {
		t.Fatalf("RollbackAgentContent: %v", err)
	}
	if resp.Msg.GetCommitSha() == "" {
		t.Fatal("expected non-empty commit SHA")
	}

	fx.fake.mu.Lock()
	defer fx.fake.mu.Unlock()
	if len(fx.fake.commits) == 0 {
		t.Fatal("no commit recorded")
	}
	lastCommit := fx.fake.commits[len(fx.fake.commits)-1]
	if !strings.Contains(lastCommit.message, "Rollback") {
		t.Errorf("commit message = %q", lastCommit.message)
	}
	if !strings.Contains(lastCommit.message, "Butter-Operation: rollback") {
		t.Errorf("commit message missing rollback operation: %q", lastCommit.message)
	}
	if !strings.Contains(lastCommit.message, targetSHA[:12]) {
		t.Errorf("commit message missing target SHA: %q", lastCommit.message)
	}
}

func TestRollbackAgentContentPermissions(t *testing.T) {
	fx, _ := newContentEditFixture(t)

	_, err := fx.svc.RollbackAgentContent(memberCtx(), connect.NewRequest(&agentsv1.RollbackAgentContentRequest{
		TargetCommitSha: "any-sha",
	}))
	wantCode(t, err, connect.CodePermissionDenied)
}

func TestRollbackAgentContentMissingTarget(t *testing.T) {
	fx, _ := newContentEditFixture(t)

	_, err := fx.svc.RollbackAgentContent(ownerCtx(), connect.NewRequest(&agentsv1.RollbackAgentContentRequest{}))
	wantCode(t, err, connect.CodeInvalidArgument)
}

func TestRollbackAgentContentChangeRequestMode(t *testing.T) {
	fx, _ := newContentEditFixture(t)

	// Switch to CHANGE_REQUEST mode.
	b := validBinding()
	b.WriteMode = agentsv1.RepoBindingWriteMode_REPO_BINDING_WRITE_MODE_CHANGE_REQUEST
	if _, err := fx.svc.PutWorkspaceRepoBinding(ownerCtx(), connect.NewRequest(&agentsv1.PutWorkspaceRepoBindingRequest{
		Binding: b,
	})); err != nil {
		t.Fatalf("Put binding: %v", err)
	}
	setCredential(t, fx, ownerCtx())
	if _, err := fx.svc.SyncWorkspaceRepository(ownerCtx(), connect.NewRequest(&agentsv1.SyncWorkspaceRepositoryRequest{})); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	targetSHA := "rollback-target"
	fx.fake.trees[targetSHA+":"] = []gitprovider.TreeEntry{
		{Path: "agents/my-agent", Kind: gitprovider.TreeEntryDirectory, SHA: "tree-agent"},
		{Path: "agents/my-agent/prompt.md", Kind: gitprovider.TreeEntryFile, Size: 10, SHA: "blob-p"},
	}
	fx.fake.blobs[targetSHA+":agents/my-agent/prompt.md"] = []byte("Old prompt.")

	resp, err := fx.svc.RollbackAgentContent(ownerCtx(), connect.NewRequest(&agentsv1.RollbackAgentContentRequest{
		TargetCommitSha: targetSHA,
	}))
	if err != nil {
		t.Fatalf("RollbackAgentContent: %v", err)
	}
	if resp.Msg.GetChangeRequestUrl() == "" {
		t.Fatal("expected change request URL for rollback in CR mode")
	}

	fx.fake.mu.Lock()
	defer fx.fake.mu.Unlock()
	if len(fx.fake.createdCRs) == 0 {
		t.Fatal("no change request was created")
	}
	cr := fx.fake.createdCRs[0]
	if !strings.HasPrefix(cr.source, "butter/rollback-") {
		t.Errorf("branch name = %q, expected butter/rollback- prefix", cr.source)
	}
}

func TestRollbackNoContent(t *testing.T) {
	fx, _ := newContentEditFixture(t)

	targetSHA := "empty-target"
	fx.fake.trees[targetSHA+":"] = []gitprovider.TreeEntry{
		{Path: "readme.md", Kind: gitprovider.TreeEntryFile, Size: 10, SHA: "blob-readme"},
	}
	fx.fake.blobs[targetSHA+":readme.md"] = []byte("# README")

	_, err := fx.svc.RollbackAgentContent(ownerCtx(), connect.NewRequest(&agentsv1.RollbackAgentContentRequest{
		TargetCommitSha: targetSHA,
	}))
	wantCode(t, err, connect.CodeFailedPrecondition)
}
