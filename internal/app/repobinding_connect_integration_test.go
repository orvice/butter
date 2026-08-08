package app

// ConnectRPC integration test for GitHostService and
// WorkspaceRepoBindingService (issue #214): real HTTP routing, gin auth
// middleware, snake_case JSON codec, workspace header plumbing, memory
// repositories, real PAT encryption, and the real GitHub adapter talking to
// an in-process fake host. Runs without external infrastructure.

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/gin-gonic/gin"

	"go.orx.me/apps/butter/internal/config"
	githostmemory "go.orx.me/apps/butter/internal/repo/githost/memory"
	repobindingmemory "go.orx.me/apps/butter/internal/repo/repobinding/memory"
	workspacememory "go.orx.me/apps/butter/internal/repo/workspace/memory"
	"go.orx.me/apps/butter/internal/runtime/daemon"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

const integrationPAT = "ghp_integration_secret"

// fakeGitHubHost serves just enough of the GitHub REST dialect for binding
// validation: one private repository with push access and a main branch.
func fakeGitHubHost(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+integrationPAT {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch r.URL.Path {
		case "/repos/acme/agents":
			fmt.Fprint(w, `{"full_name":"acme/agents","private":true,"default_branch":"main",`+
				`"permissions":{"admin":false,"maintain":false,"push":true,"triage":false,"pull":true}}`)
		case "/repos/acme/agents/branches/main":
			fmt.Fprint(w, `{"name":"main","commit":{"sha":"abc123"}}`)
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"message":"Not Found"}`)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestRepoBindingServices_ConnectIntegration(t *testing.T) {
	gitHub := fakeGitHubHost(t)

	cfg := &config.AppConfig{
		StorageBackend: "memory",
		Auth:           config.AuthConfig{AllowUnauthenticated: true},
		Git:            config.GitConfig{EncryptionKey: "0123456789abcdef0123456789abcdef"},
	}
	wsRepo := workspacememory.New()
	for _, ws := range []string{"ws-test", "ws-other"} {
		if _, err := wsRepo.CreateWorkspace(t.Context(), &agentsv1.Workspace{Id: ws, Name: ws, Slug: ws}); err != nil {
			t.Fatalf("seed workspace %s: %v", ws, err)
		}
	}
	routerFn, handlers := SetupRoutes(cfg, daemon.NewRegistry())
	handlers.Wire(&BootstrapResult{
		GitHostRepo:     githostmemory.New(),
		RepoBindingRepo: repobindingmemory.New(),
		WorkspaceRepo:   wsRepo,
	})

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	routerFn(engine)
	server := httptest.NewServer(engine)
	t.Cleanup(server.Close)

	ctx := t.Context()
	opts := []connect.ClientOption{connect.WithInterceptors(workspaceHeaderInterceptor("ws-test"))}
	hostClient := agentsv1connect.NewGitHostServiceClient(server.Client(), server.URL+"/api", opts...)
	bindingClient := agentsv1connect.NewWorkspaceRepoBindingServiceClient(server.Client(), server.URL+"/api", opts...)

	// 1. Platform admin registers the (fake) GitHub host.
	hostResp, err := hostClient.CreateGitHost(ctx, connect.NewRequest(&agentsv1.CreateGitHostRequest{
		Host: &agentsv1.GitHost{
			Name:       "Fake GitHub",
			Kind:       agentsv1.GitHostKind_GIT_HOST_KIND_GITHUB,
			ApiBaseUrl: gitHub.URL,
		},
	}))
	if err != nil {
		t.Fatalf("CreateGitHost: %v", err)
	}
	hostID := hostResp.Msg.GetHost().GetId()

	// 2. Workspace binds a repository on it.
	putResp, err := bindingClient.PutWorkspaceRepoBinding(ctx, connect.NewRequest(&agentsv1.PutWorkspaceRepoBindingRequest{
		Binding: &agentsv1.WorkspaceRepoBinding{
			GitHostId:  hostID,
			Repository: "acme/agents",
			Branch:     "main",
			RootPath:   "butter",
		},
	}))
	if err != nil {
		t.Fatalf("PutWorkspaceRepoBinding: %v", err)
	}
	if got := putResp.Msg.GetBinding().GetWorkspaceId(); got != "ws-test" {
		t.Fatalf("workspace not derived from header: %q", got)
	}

	// 3. Owner sets the PAT (write-only) and validates the binding through
	// the real GitHub adapter against the fake host.
	if _, err := bindingClient.SetWorkspaceRepoBindingCredential(ctx, connect.NewRequest(&agentsv1.SetWorkspaceRepoBindingCredentialRequest{
		Pat: integrationPAT,
	})); err != nil {
		t.Fatalf("SetWorkspaceRepoBindingCredential: %v", err)
	}
	valResp, err := bindingClient.ValidateWorkspaceRepoBinding(ctx, connect.NewRequest(&agentsv1.ValidateWorkspaceRepoBindingRequest{}))
	if err != nil {
		t.Fatalf("ValidateWorkspaceRepoBinding: %v", err)
	}
	status := valResp.Msg.GetBinding().GetStatus()
	if status.GetState() != agentsv1.RepoBindingConnectionState_REPO_BINDING_CONNECTION_STATE_OK {
		t.Fatalf("state = %v, error = %q, checks = %v", status.GetState(), status.GetError(), status.GetChecks())
	}

	// 4. Status is visible over Get; the PAT never crosses the wire back.
	getResp, err := bindingClient.GetWorkspaceRepoBinding(ctx, connect.NewRequest(&agentsv1.GetWorkspaceRepoBindingRequest{}))
	if err != nil {
		t.Fatalf("GetWorkspaceRepoBinding: %v", err)
	}
	binding := getResp.Msg.GetBinding()
	if !binding.GetCredentialSet() {
		t.Fatal("credential_set not reported")
	}
	if binding.GetStatus().GetState() != agentsv1.RepoBindingConnectionState_REPO_BINDING_CONNECTION_STATE_OK {
		t.Fatalf("validation state not persisted: %v", binding.GetStatus())
	}
	if s := binding.String(); strings.Contains(s, integrationPAT) {
		t.Fatalf("binding leaks PAT: %s", s)
	}

	// 5. A wrong workspace header sees no binding (workspace isolation).
	otherOpts := []connect.ClientOption{connect.WithInterceptors(workspaceHeaderInterceptor("ws-other"))}
	otherClient := agentsv1connect.NewWorkspaceRepoBindingServiceClient(server.Client(), server.URL+"/api", otherOpts...)
	otherResp, err := otherClient.GetWorkspaceRepoBinding(ctx, connect.NewRequest(&agentsv1.GetWorkspaceRepoBindingRequest{}))
	if err != nil {
		t.Fatalf("Get (other workspace): %v", err)
	}
	if otherResp.Msg.GetBinding() != nil {
		t.Fatalf("binding visible across workspaces: %v", otherResp.Msg.GetBinding())
	}

	// 6. Delete removes binding and credential.
	if _, err := bindingClient.DeleteWorkspaceRepoBinding(ctx, connect.NewRequest(&agentsv1.DeleteWorkspaceRepoBindingRequest{})); err != nil {
		t.Fatalf("DeleteWorkspaceRepoBinding: %v", err)
	}
	getResp, err = bindingClient.GetWorkspaceRepoBinding(ctx, connect.NewRequest(&agentsv1.GetWorkspaceRepoBindingRequest{}))
	if err != nil {
		t.Fatalf("Get after delete: %v", err)
	}
	if getResp.Msg.GetBinding() != nil {
		t.Fatal("binding survived delete")
	}
}
