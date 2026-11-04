package app

import (
	"connectrpc.com/connect"
	"context"
	"github.com/gin-gonic/gin"
	"go.orx.me/apps/butter/internal/config"
	cryptokeymemory "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	"go.orx.me/apps/butter/internal/runtime/daemon"
	"go.orx.me/apps/butter/internal/secretbox"
	workstationmemory "go.orx.me/apps/butter/internal/workstation/memory"
	"go.orx.me/apps/butter/internal/workstation/providerregistry"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type workstationFixtureTransport struct{}

func (workstationFixtureTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.Method != http.MethodGet || r.URL.String() != "https://api.github.com/user" {
		return nil, io.ErrUnexpectedEOF
	}
	return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"id":7}`)), Request: r}, nil
}
func TestConfiguredRegistryConnectCreateSmoke(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, h := SetupRoutes(&config.AppConfig{Auth: config.AuthConfig{AllowUnauthenticated: true}}, daemon.NewRegistry())
	if err := h.workStationSvcServer.SetConfiguredService(providerregistry.Config{Kind: providerregistry.GitHub}, providerregistry.ServiceDeps{Repo: workstationmemory.New(), Keyring: secretbox.NewKeyring(cryptokeymemory.New()), Transport: workstationFixtureTransport{}, CheckHost: func(context.Context, string) error { return nil }}, nil); err != nil {
		t.Fatal(err)
	}
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
	resp, err := client.CreateConnection(ctx, connect.NewRequest(&v1.WorkStationServiceCreateConnectionRequest{Name: "registry", HostId: "approved", ApiKey: "fixture"}))
	if err != nil {
		t.Fatal(err)
	}
	if resp.Msg.Connection.Provider != "github" || !resp.Msg.Connection.CredentialSet {
		t.Fatalf("connection=%+v", resp.Msg.Connection)
	}
}
