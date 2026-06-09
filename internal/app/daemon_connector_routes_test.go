package app

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/application"
	"go.orx.me/apps/butter/internal/config"
	tokenmem "go.orx.me/apps/butter/internal/repo/apitoken/memory"
	"go.orx.me/apps/butter/internal/runtime/daemon"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

func TestDaemonConnectorService_ConnectsViaAPIPrefix(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const (
		workspaceID = "ws-daemon-route"
		runtimeID   = "daemon-route"
		secret      = "bt_daemon_route_secret"
	)

	cfg := &config.AppConfig{}
	registry := daemon.NewRegistry()
	router, handlers := SetupRoutes(cfg, registry)

	if _, err := handlers.ConfigStore().CreateDaemonRuntime(context.Background(), workspaceID, &agentsv1.DaemonRuntime{
		Id:          runtimeID,
		Name:        "Route daemon",
		WorkspaceId: workspaceID,
	}); err != nil {
		t.Fatalf("seed daemon runtime: %v", err)
	}

	tokenRepo := tokenmem.New()
	if err := tokenRepo.Create(context.Background(), &agentsv1.APIToken{
		Id:              "tok-daemon-route",
		Name:            "daemon route",
		WorkspaceId:     workspaceID,
		Kind:            agentsv1.APITokenKind_API_TOKEN_KIND_DAEMON,
		Scopes:          []string{"daemon:connect"},
		DaemonRuntimeId: runtimeID,
		CreatedAt:       timestamppb.New(time.Now().UTC()),
	}, application.HashAPITokenSecret(secret)); err != nil {
		t.Fatalf("seed daemon token: %v", err)
	}
	handlers.Wire(&BootstrapResult{APITokenRepo: tokenRepo})

	engine := gin.New()
	router(engine)
	server := httptest.NewServer(h2c.NewHandler(engine, &http2.Server{}))
	defer server.Close()

	client := agentsv1connect.NewDaemonConnectorServiceClient(h2cClient(), server.URL+"/api")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stream := client.Connect(ctx)
	stream.RequestHeader().Set("Authorization", "Bearer "+secret)
	if err := stream.Send(&agentsv1.ConnectRequest{
		Message: &agentsv1.ConnectRequest_Register{
			Register: &agentsv1.DaemonInfo{
				DaemonRuntimeId: runtimeID,
				Name:            "route-test",
				AcpRuntimes:     []string{"codex"},
			},
		},
	}); err != nil {
		t.Fatalf("send register via /api route: %v", err)
	}

	if conn := waitForRouteConn(registry, workspaceID, runtimeID, 2*time.Second); conn == nil {
		t.Fatal("daemon did not register through /api-prefixed connector route")
	}

	if err := stream.CloseRequest(); err != nil && connect.CodeOf(err) != connect.CodeCanceled {
		t.Fatalf("close request: %v", err)
	}
}

func h2cClient() *http.Client {
	return &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, network, addr)
			},
		},
	}
}

func waitForRouteConn(registry *daemon.Registry, workspaceID, runtimeID string, timeout time.Duration) *daemon.Connection {
	deadline := time.After(timeout)
	for {
		if conn := registry.Get(workspaceID, runtimeID); conn != nil {
			return conn
		}
		select {
		case <-deadline:
			return nil
		case <-time.After(10 * time.Millisecond):
		}
	}
}
