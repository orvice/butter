package application

import (
	"strings"
	"testing"

	"connectrpc.com/connect"
	agentfilememory "go.orx.me/apps/butter/internal/repo/agentfile/memory"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestAgentFileServiceWriteEnforcesMaxFileBytes(t *testing.T) {
	ctx := workspace.WithID(t.Context(), "ws-files")
	repo := agentfilememory.New()
	svc := NewAgentFileServiceServer(repo)
	svc.SetMaxFileBytes(4)

	space, err := repo.CreateSpace(ctx, "ws-files", &agentsv1.AgentFileSpace{Name: "Notes"})
	if err != nil {
		t.Fatalf("CreateSpace: %v", err)
	}

	if _, err := svc.WriteAgentFile(ctx, &agentsv1.WriteAgentFileRequest{
		SpaceId: space.GetId(),
		Path:    "/ok.txt",
		Content: "1234",
	}); err != nil {
		t.Fatalf("WriteAgentFile within limit: %v", err)
	}

	_, err = svc.WriteAgentFile(ctx, &agentsv1.WriteAgentFileRequest{
		SpaceId: space.GetId(),
		Path:    "/too-large.txt",
		Content: "12345",
	})
	if err == nil {
		t.Fatal("expected max size error")
	}
	twerr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if twerr.Code() != connect.CodeInvalidArgument {
		t.Fatalf("expected InvalidArgument, got %s", twerr.Code())
	}
	if !strings.Contains(twerr.Message(), "max file size") {
		t.Fatalf("unexpected error message: %q", twerr.Message())
	}
}
