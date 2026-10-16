package main

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"connectrpc.com/connect"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

func memoryConfigClient(t *testing.T) agentsv1connect.WorkspaceMemoryConfigServiceClient {
	t.Helper()
	f := newProviderFixture(t)
	interceptor := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+f.token)
			req.Header().Set("X-Workspace-ID", memoryWorkspace)
			return next(ctx, req)
		}
	})
	return agentsv1connect.NewWorkspaceMemoryConfigServiceClient(
		http.DefaultClient, f.url+"/api", connect.WithInterceptors(interceptor),
	)
}

func TestMemoryPutRejectsBadCredentials(t *testing.T) {
	ctx := context.Background()
	client := memoryConfigClient(t)
	got, err := client.GetWorkspaceMemoryConfig(ctx, connect.NewRequest(&agentsv1.GetWorkspaceMemoryConfigRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	badKey := "wrong-key"
	_, err = client.PutWorkspaceMemoryConfig(ctx, connect.NewRequest(&agentsv1.PutWorkspaceMemoryConfigRequest{
		BaseUrl: got.Msg.GetConfig().GetBaseUrl(), Enabled: true, ApiKey: &badKey,
	}))
	wantCode(t, err, connect.CodeFailedPrecondition)
}

func TestMemoryPutUnreachableSavesWithWarning(t *testing.T) {
	ctx := context.Background()
	client := memoryConfigClient(t)
	key := contractMem0Key
	got, err := client.PutWorkspaceMemoryConfig(ctx, connect.NewRequest(&agentsv1.PutWorkspaceMemoryConfigRequest{
		BaseUrl: "http://127.0.0.1:1/", Enabled: true, ApiKey: &key,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if got.Msg.GetWarning() == "" {
		t.Fatal("expected probe warning for unreachable mem0")
	}
	if !strings.Contains(got.Msg.GetWarning(), "mem0") {
		t.Fatalf("warning = %q", got.Msg.GetWarning())
	}
}
