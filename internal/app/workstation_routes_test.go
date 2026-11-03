package app

import (
	"connectrpc.com/connect"
	"context"
	"github.com/gin-gonic/gin"
	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/runtime/daemon"
	wsmemory "go.orx.me/apps/butter/internal/workstation/memory"
	"go.orx.me/apps/butter/internal/workstation/repotest"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWorkStationConnectRouteAndNoStore(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, _ := SetupRoutes(&config.AppConfig{Auth: config.AuthConfig{AllowUnauthenticated: true}}, daemon.NewRegistry())
	engine := gin.New()
	router(engine)
	req := httptest.NewRequest(http.MethodPost, "/api/agents.v1.WorkStationService/ListConnections", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Workspace-ID", "A")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("route not wired: %d %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("private projection cacheable")
	}
}

func TestWorkStationShutdownCancelsAndJoins(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	h := &Handlers{stopWorkStation: cancel, workStationDone: done}
	go func() { <-ctx.Done(); close(done) }()
	deadline, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()
	if err := h.ShutdownWorkStation(deadline); err != nil {
		t.Fatal(err)
	}
	if err := h.ShutdownWorkStation(deadline); err != nil {
		t.Fatal("shutdown not idempotent", err)
	}
}

func TestWorkStationConnectHTTPHappyPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, h := SetupRoutes(&config.AppConfig{Auth: config.AuthConfig{AllowUnauthenticated: true}}, daemon.NewRegistry())
	domain := repotest.NewService(wsmemory.New(), &repotest.Provider{Items: []*v1.WorkStationItem{repotest.Item("1", "HTTP smoke", time.Now().UTC())}})
	h.workStationSvcServer.SetService(domain, nil)
	engine := gin.New()
	router(engine)
	server := httptest.NewServer(engine)
	defer server.Close()
	client := agentsv1connect.NewWorkStationServiceClient(server.Client(), server.URL+"/api", connect.WithInterceptors(connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
			request.Header().Set("X-Workspace-ID", "A")
			return next(ctx, request)
		}
	})))
	ctx := context.Background()
	created, err := client.CreateConnection(ctx, connect.NewRequest(&v1.WorkStationServiceCreateConnectionRequest{Name: "HTTP connection", HostId: "approved-host", ApiKey: "synthetic-test-key"}))
	if err != nil {
		t.Fatal(err)
	}
	c := created.Msg.Connection
	if !c.CredentialSet || len(c.Scopes) != 0 || created.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("create contract", created.Msg)
	}
	_, err = client.SetSelectedScopes(ctx, connect.NewRequest(&v1.WorkStationServiceSetSelectedScopesRequest{Id: c.Id, ExpectedRevision: c.Revision, ScopeIds: []string{"42"}}))
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Refresh(ctx, connect.NewRequest(&v1.WorkStationServiceRefreshRequest{Id: c.Id}))
	if err != nil {
		t.Fatal(err)
	}
	before, err := client.ListItems(ctx, connect.NewRequest(&v1.WorkStationServiceListItemsRequest{}))
	if err != nil || len(before.Msg.Items) != 0 {
		t.Fatal("refresh exposed staging", err)
	}
	if err := domain.RunOnce(ctx); err != nil {
		t.Fatal(err)
	}
	after, err := client.ListItems(ctx, connect.NewRequest(&v1.WorkStationServiceListItemsRequest{}))
	if err != nil || len(after.Msg.Items) != 1 || after.Msg.Items[0].Title != "HTTP smoke" {
		t.Fatal("HTTP projection", after, err)
	}
	detail, err := client.GetItem(ctx, connect.NewRequest(&v1.WorkStationServiceGetItemRequest{Id: after.Msg.Items[0].Id}))
	if err != nil || detail.Msg.Item.Title != "HTTP smoke" {
		t.Fatal("HTTP detail", err)
	}
}
