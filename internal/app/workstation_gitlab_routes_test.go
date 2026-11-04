package app

import (
	"connectrpc.com/connect"
	"context"
	"github.com/gin-gonic/gin"
	"go.orx.me/apps/butter/internal/config"
	cryptomemory "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	"go.orx.me/apps/butter/internal/runtime/daemon"
	"go.orx.me/apps/butter/internal/secretbox"
	wsmemory "go.orx.me/apps/butter/internal/workstation/memory"
	"go.orx.me/apps/butter/internal/workstation/providerregistry"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type gitlabRouteFixture func(*http.Request) (*http.Response, error)

func (f gitlabRouteFixture) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestGitLabMergeRequestThroughConnectAndProjection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	route, h := SetupRoutes(&config.AppConfig{Auth: config.AuthConfig{AllowUnauthenticated: true}}, daemon.NewRegistry())
	calls := 0
	transport := gitlabRouteFixture(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Method != http.MethodGet || r.URL.Host != "gitlab.com" || !strings.HasPrefix(r.URL.Path, "/api/v4/") || r.Header.Get("PRIVATE-TOKEN") != "synthetic-key" {
			t.Fatal("escaped readonly approved profile")
		}
		body := ""
		switch r.URL.Path {
		case "/api/v4/user":
			body = `{"id":7}`
		case "/api/v4/version":
			body = `{"version":"18.0.0-synthetic","revision":"fixture"}`
		case "/api/v4/projects/42":
			body = `{"id":42,"path_with_namespace":"group/project","issues_enabled":true,"merge_requests_enabled":true}`
		case "/api/v4/projects/42/issues":
			body = `[]`
		case "/api/v4/projects/42/merge_requests":
			body = `[{"iid":7,"project_id":42,"title":"Synthetic draft MR","description":"local detail","state":"opened","draft":true,"created_at":"2026-10-01T00:00:00Z","updated_at":"2026-11-01T00:00:00Z","author":{"id":9},"assignees":[{"id":10,"username":"owner"}],"labels":["review"]}]`
		default:
			t.Fatal("unexpected endpoint", r.URL.Path)
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
	})
	station, err := providerregistry.NewService(providerregistry.Config{Kind: providerregistry.GitLab}, providerregistry.ServiceDeps{Repo: wsmemory.New(), Keyring: secretbox.NewKeyring(cryptomemory.New()), Transport: transport, CheckHost: func(context.Context, string) error { return nil }})
	if err != nil {
		t.Fatal(err)
	}
	h.workStationSvcServer.SetService(station, nil)
	engine := gin.New()
	route(engine)
	server := httptest.NewServer(engine)
	defer server.Close()
	client := agentsv1connect.NewWorkStationServiceClient(server.Client(), server.URL+"/api", connect.WithInterceptors(connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, r connect.AnyRequest) (connect.AnyResponse, error) {
			r.Header().Set("X-Workspace-ID", "A")
			return next(ctx, r)
		}
	})))
	ctx := context.Background()
	created, err := client.CreateConnection(ctx, connect.NewRequest(&v1.WorkStationServiceCreateConnectionRequest{Name: "GitLab", HostId: "approved", ApiKey: "synthetic-key", Provider: "gitlab"}))
	if err != nil {
		t.Fatal(err)
	}
	c := created.Msg.Connection
	if c.Provider != "gitlab" || !c.CredentialSet {
		t.Fatal(c)
	}
	if _, err := client.SetSelectedScopes(ctx, connect.NewRequest(&v1.WorkStationServiceSetSelectedScopesRequest{Id: c.Id, ExpectedRevision: c.Revision, ScopeIds: []string{"42"}})); err != nil {
		t.Fatal(err)
	}
	if err := station.RunOnce(ctx); err != nil {
		t.Fatal(err)
	}
	rows, err := client.ListItems(ctx, connect.NewRequest(&v1.WorkStationServiceListItemsRequest{Kind: "MERGE_REQUEST", Provider: "gitlab", Status: "all"}))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows.Msg.Items) != 1 || !rows.Msg.Items[0].Draft || rows.Msg.Items[0].Status != "OPEN" || rows.Msg.Items[0].Url != "https://gitlab.com/group/project/-/merge_requests/7" {
		t.Fatal(rows.Msg)
	}
	detail, err := client.GetItem(ctx, connect.NewRequest(&v1.WorkStationServiceGetItemRequest{Id: rows.Msg.Items[0].Id}))
	if err != nil || detail.Msg.Item.Description != "local detail" {
		t.Fatal(detail, err)
	}
	options, err := client.ListFilterOptions(ctx, connect.NewRequest(&v1.WorkStationServiceListFilterOptionsRequest{Kind: "MERGE_REQUEST", Provider: "gitlab", Dimension: "assignee"}))
	if err != nil || len(options.Msg.Options) != 1 || options.Msg.Options[0].Id != "10:owner" {
		t.Fatal(options, err)
	}
	if calls < 7 {
		t.Fatal("adapter traversal bypassed", calls)
	}
}
