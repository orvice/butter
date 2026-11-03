package application

import (
	"context"
	"strings"
	"testing"

	"connectrpc.com/connect"
	authrepo "go.orx.me/apps/butter/internal/repo/auth"
	githostmemory "go.orx.me/apps/butter/internal/repo/githost/memory"
	workspacerepo "go.orx.me/apps/butter/internal/repo/workspace"
	workspacememory "go.orx.me/apps/butter/internal/repo/workspace/memory"
	workspacecontext "go.orx.me/apps/butter/internal/workspace"
	ws "go.orx.me/apps/butter/internal/workstation"
	wsmemory "go.orx.me/apps/butter/internal/workstation/memory"
	"go.orx.me/apps/butter/internal/workstation/repotest"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

func stationFixture(t *testing.T) (*WorkStationServiceServer, *ws.Service, workspacerepo.Repository) {
	t.Helper()
	repo := wsmemory.New()
	domain := repotest.NewService(repo, &repotest.Provider{})
	workspaces := workspacememory.New()
	ctx := context.Background()
	_, err := workspaces.CreateWorkspace(ctx, &v1.Workspace{Id: "A", Name: "A", Slug: "a"})
	if err != nil {
		t.Fatal(err)
	}
	for _, role := range []string{"owner", "admin", "member"} {
		_, err := workspaces.AddMember(ctx, &v1.WorkspaceMember{WorkspaceId: "A", UserId: role, Role: role})
		if err != nil {
			t.Fatal(err)
		}
	}
	s := NewWorkStationServiceServer()
	s.SetService(domain, workspaces)
	return s, domain, workspaces
}
func stationContext(role string) context.Context {
	return workspacecontext.WithID(authrepo.WithAuthenticated(context.Background(), &v1.User{Id: role, Role: "user"}, nil), "A")
}
func TestWorkStationPermissionsAndWriteOnlyCredentials(t *testing.T) {
	s, _, _ := stationFixture(t)
	for _, role := range []string{"admin", "member"} {
		ctx := stationContext(role)
		if _, err := s.ListConnections(ctx, connect.NewRequest(&v1.WorkStationServiceListConnectionsRequest{})); err != nil {
			t.Fatal(err)
		}
		calls := []func() error{
			func() error {
				_, e := s.CreateConnection(ctx, connect.NewRequest(&v1.WorkStationServiceCreateConnectionRequest{}))
				return e
			},
			func() error {
				_, e := s.UpdateConnection(ctx, connect.NewRequest(&v1.WorkStationServiceUpdateConnectionRequest{}))
				return e
			},
			func() error {
				_, e := s.DeleteConnection(ctx, connect.NewRequest(&v1.WorkStationServiceDeleteConnectionRequest{}))
				return e
			},
			func() error {
				_, e := s.TestConnection(ctx, connect.NewRequest(&v1.WorkStationServiceTestConnectionRequest{}))
				return e
			},
			func() error {
				_, e := s.DiscoverScopes(ctx, connect.NewRequest(&v1.WorkStationServiceDiscoverScopesRequest{}))
				return e
			},
			func() error {
				_, e := s.SetSelectedScopes(ctx, connect.NewRequest(&v1.WorkStationServiceSetSelectedScopesRequest{}))
				return e
			},
		}
		for _, call := range calls {
			if code := connect.CodeOf(call()); code != connect.CodePermissionDenied {
				t.Fatalf("%s: got %s", role, code)
			}
		}
	}
	owner := stationContext("owner")
	res, err := s.CreateConnection(owner, connect.NewRequest(&v1.WorkStationServiceCreateConnectionRequest{Name: "one", HostId: "host", ApiKey: "super-private-token"}))
	if err != nil {
		t.Fatal(err)
	}
	id := res.Msg.Connection.Id
	read, err := s.ListConnections(stationContext("member"), connect.NewRequest(&v1.WorkStationServiceListConnectionsRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	data, err := protojson.Marshal(read.Msg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetConnection(stationContext("member"), connect.NewRequest(&v1.WorkStationServiceGetConnectionRequest{Id: id})); err != nil {
		t.Fatal("member cannot read one connection", err)
	}
	if strings.Contains(string(data), "super-private-token") || strings.Contains(string(data), "ciphertext") || strings.Contains(string(data), "key_id") {
		t.Fatal("credential leaked")
	}
	outsider := stationContext("outsider")
	if _, err := s.GetConnection(outsider, connect.NewRequest(&v1.WorkStationServiceGetConnectionRequest{Id: id})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatal("nonmember read one connection", err)
	}
	if _, err := s.Refresh(stationContext("member"), connect.NewRequest(&v1.WorkStationServiceRefreshRequest{Id: id})); err != nil {
		t.Fatal(err)
	}
	ctx := workspacecontext.WithID(authrepo.WithAdmin(context.Background()), "A")
	if _, err := s.CreateConnection(ctx, connect.NewRequest(&v1.WorkStationServiceCreateConnectionRequest{Name: "global", HostId: "host", ApiKey: "global-token"})); err != nil {
		t.Fatal(err)
	}
}
func TestWorkStationRechecksMembershipAndWorkspaceEveryRead(t *testing.T) {
	s, domain, workspaces := stationFixture(t)
	r, err := domain.Create(context.Background(), "A", "main", "host", "key")
	if err != nil {
		t.Fatal(err)
	}
	if err := workspaces.RemoveMember(context.Background(), "A", "member"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ListItems(stationContext("member"), connect.NewRequest(&v1.WorkStationServiceListItemsRequest{})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatal("ex-member still reads", err)
	}
	b := workspacecontext.WithID(authrepo.WithAdmin(context.Background()), "B")
	if _, err := s.Refresh(b, connect.NewRequest(&v1.WorkStationServiceRefreshRequest{Id: r.ID})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatal("cross-workspace refresh", err)
	}
}
func TestWorkStationHostProfileAndReferenceGuard(t *testing.T) {
	ctx := authrepo.WithAdmin(context.Background())
	hosts := githostmemory.New()
	h, err := hosts.Create(ctx, &v1.GitHost{Id: "official", Name: "GitHub", Kind: v1.GitHostKind_GIT_HOST_KIND_GITHUB, ApiBaseUrl: "https://api.github.com", WebBaseUrl: "https://github.com"})
	if err != nil {
		t.Fatal(err)
	}
	if err := WorkStationHostCheck(hosts)(ctx, h.Id); err != nil {
		t.Fatal(err)
	}
	repo := wsmemory.New()
	service := repotest.NewService(repo, &repotest.Provider{})
	if _, err := service.Create(ctx, "A", "one", h.Id, "secret"); err != nil {
		t.Fatal(err)
	}
	s := NewGitHostServiceServer(hosts)
	s.SetWorkStationHostGuard(repo.HostInUse)
	h.Name = "renamed"
	if _, err := s.UpdateGitHost(ctx, connect.NewRequest(&v1.UpdateGitHostRequest{Host: h})); err != nil {
		t.Fatal("rename blocked", err)
	}
	h.ApiBaseUrl = "https://evil.invalid"
	if _, err := s.UpdateGitHost(ctx, connect.NewRequest(&v1.UpdateGitHostRequest{Host: h})); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatal("endpoint changed", err)
	}
	if _, err := s.DeleteGitHost(ctx, connect.NewRequest(&v1.DeleteGitHostRequest{Id: h.Id})); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatal("referenced host deleted", err)
	}
	for _, endpoint := range []string{"https://api.github.com@evil.invalid", "http://api.github.com", "https://api.github.com/evil", "https://api.github.com?token=secret"} {
		other, err := hosts.Create(ctx, &v1.GitHost{Id: endpoint, Name: "bad", Kind: v1.GitHostKind_GIT_HOST_KIND_GITHUB, ApiBaseUrl: endpoint, WebBaseUrl: "https://github.com"})
		if err != nil {
			t.Fatal(err)
		}
		if WorkStationHostCheck(hosts)(ctx, other.Id) == nil {
			t.Fatalf("accepted %s", endpoint)
		}
	}
}

func TestWorkStationMissingIDIsInvalid(t *testing.T) {
	s, _, _ := stationFixture(t)
	ctx := stationContext("owner")
	calls := []func() error{
		func() error {
			_, e := s.DeleteConnection(ctx, connect.NewRequest(&v1.WorkStationServiceDeleteConnectionRequest{ExpectedRevision: 1}))
			return e
		},
		func() error {
			_, e := s.Refresh(ctx, connect.NewRequest(&v1.WorkStationServiceRefreshRequest{}))
			return e
		},
		func() error {
			_, e := s.GetConnection(ctx, connect.NewRequest(&v1.WorkStationServiceGetConnectionRequest{}))
			return e
		},
		func() error {
			_, e := s.GetItem(ctx, connect.NewRequest(&v1.WorkStationServiceGetItemRequest{}))
			return e
		},
	}
	for _, call := range calls {
		if code := connect.CodeOf(call()); code != connect.CodeInvalidArgument {
			t.Fatalf("missing id: %s", code)
		}
	}
}
