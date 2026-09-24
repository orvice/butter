package application

import (
	"strings"
	"testing"

	"connectrpc.com/connect"

	butterboxmemory "go.orx.me/apps/butter/internal/repo/butterbox/memory"
	"go.orx.me/apps/butter/internal/repo/config/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func testCursorAgent(agentID, boxID string) *agentsv1.Agent {
	return &agentsv1.Agent{
		Name:    agentID,
		AgentId: agentID,
		Type:    agentsv1.AgentType_AGENT_TYPE_CURSOR,
		Config: &agentsv1.AgentConfig{
			Cursor: &agentsv1.CursorAgentConfig{ButterboxId: boxID, WorkingDir: "projects/demo", Mode: "agent"},
		},
	}
}

func TestCreateCursorAgent(t *testing.T) {
	boxes := butterboxmemory.New()
	seedButterBox(t, boxes, "box-1", true)
	seedButterBox(t, boxes, "box-disabled", false)
	svc := NewAgentServiceServer(memory.New())
	svc.SetButterBoxRepo(boxes)
	ctx := testCtx()

	create := func(a *agentsv1.Agent) (*connect.Response[agentsv1.CreateAgentResponse], error) {
		return svc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{Agent: a}))
	}

	_, err := create(testCursorAgent("cursor-a", "missing-box"))
	if connect.CodeOf(err) != connect.CodeInvalidArgument || !strings.Contains(err.Error(), "config.cursor.butterbox_id") || !strings.Contains(err.Error(), "missing-box") {
		t.Fatalf("expected InvalidArgument naming the box, got %v", err)
	}

	_, err = create(testCursorAgent("cursor-disabled", "box-disabled"))
	if connect.CodeOf(err) != connect.CodeInvalidArgument || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("expected InvalidArgument for disabled box, got %v", err)
	}

	bad := testCursorAgent("cursor-b", "box-1")
	bad.Config.McpServerIds = []string{"m1"}
	_, err = create(bad)
	if connect.CodeOf(err) != connect.CodeInvalidArgument || !strings.Contains(err.Error(), "mcp_server_ids") || !strings.Contains(err.Error(), ".cursor/rules") {
		t.Fatalf("expected InvalidArgument naming mcp_server_ids, got %v", err)
	}

	resp, err := create(testCursorAgent("cursor-ok", " box-1 "))
	if err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}
	if got := resp.Msg.GetAgent().GetConfig().GetCursor().GetButterboxId(); got != "box-1" {
		t.Fatalf("stored butterbox_id: got %q", got)
	}

	// Disabling a box does not invalidate an agent already bound to it.
	box, err := boxes.Get(ctx, wsTest, "box-1")
	if err != nil {
		t.Fatalf("get butterbox: %v", err)
	}
	box.Enabled = false
	if _, err := boxes.Update(ctx, wsTest, box); err != nil {
		t.Fatalf("disable butterbox: %v", err)
	}
	update := testCursorAgent("cursor-ok", "box-1")
	update.Config.Cursor.WorkingDir = "projects/other"
	if _, err := svc.UpdateAgent(ctx, connect.NewRequest(&agentsv1.UpdateAgentRequest{Agent: update})); err != nil {
		t.Fatalf("UpdateAgent existing disabled binding: %v", err)
	}
}

func TestDeleteButterBoxRefusedWhileCursorAgentsReferenceIt(t *testing.T) {
	agents := memory.New()
	boxes := butterboxmemory.New()
	seedButterBox(t, boxes, "box-1", true)

	agentSvc := NewAgentServiceServer(agents)
	agentSvc.SetButterBoxRepo(boxes)
	boxSvc := NewButterBoxServiceServer(boxes)
	boxSvc.SetAgentRepo(agents)
	ctx := testCtx()

	for _, a := range []*agentsv1.Agent{testCursorAgent("cursor-one", "box-1"), testPiAgent("pi-one", "box-1")} {
		if _, err := agentSvc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{Agent: a})); err != nil {
			t.Fatalf("CreateAgent %s: %v", a.GetAgentId(), err)
		}
	}

	_, err := boxSvc.DeleteButterBox(ctx, connect.NewRequest(&agentsv1.DeleteButterBoxRequest{Id: "box-1"}))
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("expected FailedPrecondition, got %v", err)
	}
	if !strings.Contains(err.Error(), "cursor-one") || !strings.Contains(err.Error(), "pi-one") {
		t.Fatalf("error must list every referencing agent, got %v", err)
	}
}
