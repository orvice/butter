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
			Cursor: &agentsv1.CursorAgentConfig{
				ButterboxId: boxID,
				WorkingDir:  "projects/demo",
				Model:       "composer-2.5",
				Mode:        "agent",
			},
		},
	}
}

func TestCreateCursorAgent(t *testing.T) {
	store := memory.New()
	boxes := butterboxmemory.New()
	seedButterBox(t, boxes, "box-1", true)
	seedButterBox(t, boxes, "box-disabled", false)
	svc := NewAgentServiceServer(store)
	svc.SetButterBoxRepo(boxes)
	ctx := testCtx()

	// Unknown box is rejected with a pointed error.
	_, err := svc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: testCursorAgent("cur-a", "missing-box"),
	}))
	if err == nil || connect.CodeOf(err) != connect.CodeInvalidArgument || !strings.Contains(err.Error(), "missing-box") {
		t.Fatalf("expected InvalidArgument naming the box, got %v", err)
	}

	// A disabled box is not available for a new binding.
	_, err = svc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: testCursorAgent("cur-disabled", "box-disabled"),
	}))
	if err == nil || connect.CodeOf(err) != connect.CodeInvalidArgument || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("expected InvalidArgument for a disabled box, got %v", err)
	}

	// Box-owned behavior fields are rejected on write.
	bad := testCursorAgent("cur-b", "box-1")
	bad.Config.Instruction = "be helpful"
	_, err = svc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{Agent: bad}))
	if err == nil || connect.CodeOf(err) != connect.CodeInvalidArgument || !strings.Contains(err.Error(), "instruction") {
		t.Fatalf("expected InvalidArgument naming instruction, got %v", err)
	}

	// A cursor agent is a leaf.
	withChild := testCursorAgent("cur-c", "box-1")
	withChild.ChildAgentIds = []string{"other"}
	_, err = svc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{Agent: withChild}))
	if err == nil || connect.CodeOf(err) != connect.CodeInvalidArgument || !strings.Contains(err.Error(), "child_agent_ids") {
		t.Fatalf("expected InvalidArgument naming child_agent_ids, got %v", err)
	}

	// A valid cursor agent is accepted and the box reference is canonicalized.
	resp, err := svc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: testCursorAgent("cur-ok", " box-1 "),
	}))
	if err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}
	if got := resp.Msg.GetAgent().GetConfig().GetCursor().GetButterboxId(); got != "box-1" {
		t.Fatalf("stored butterbox_id: got %q", got)
	}
}

func TestUpdateCursorAgentValidates(t *testing.T) {
	store := memory.New()
	boxes := butterboxmemory.New()
	seedButterBox(t, boxes, "box-1", true)
	svc := NewAgentServiceServer(store)
	svc.SetButterBoxRepo(boxes)
	ctx := testCtx()

	if _, err := svc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: testCursorAgent("cur-1", "box-1"),
	})); err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}

	// Repointing to an unknown box is rejected.
	_, err := svc.UpdateAgent(ctx, connect.NewRequest(&agentsv1.UpdateAgentRequest{
		Agent: testCursorAgent("cur-1", "gone-box"),
	}))
	if err == nil || connect.CodeOf(err) != connect.CodeInvalidArgument || !strings.Contains(err.Error(), "gone-box") {
		t.Fatalf("expected InvalidArgument naming the box, got %v", err)
	}

	// Repointing the working directory or model is a legal update.
	update := testCursorAgent("cur-1", "box-1")
	update.Config.Cursor.WorkingDir = "projects/other"
	update.Config.Cursor.Model = "auto-smart"
	update.Config.Cursor.Mode = "plan"
	if _, err := svc.UpdateAgent(ctx, connect.NewRequest(&agentsv1.UpdateAgentRequest{Agent: update})); err != nil {
		t.Fatalf("UpdateAgent: %v", err)
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

	for _, id := range []string{"cur-one", "cur-two"} {
		if _, err := agentSvc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{
			Agent: testCursorAgent(id, "box-1"),
		})); err != nil {
			t.Fatalf("CreateAgent %s: %v", id, err)
		}
	}

	_, err := boxSvc.DeleteButterBox(ctx, connect.NewRequest(&agentsv1.DeleteButterBoxRequest{Id: "box-1"}))
	if err == nil || connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("expected FailedPrecondition, got %v", err)
	}
	if !strings.Contains(err.Error(), "cur-one") || !strings.Contains(err.Error(), "cur-two") {
		t.Fatalf("error must list the referencing agent ids, got %v", err)
	}

	// A soft delete leaves a tombstone that still holds the box reference —
	// the guard keeps refusing until the rows are gone.
	for _, id := range []string{"cur-one", "cur-two"} {
		if _, err := agentSvc.DeleteAgent(ctx, connect.NewRequest(&agentsv1.DeleteAgentRequest{AgentId: id})); err != nil {
			t.Fatalf("DeleteAgent %s: %v", id, err)
		}
	}
	if _, err := boxSvc.DeleteButterBox(ctx, connect.NewRequest(&agentsv1.DeleteButterBoxRequest{Id: "box-1"})); err == nil || connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("expected FailedPrecondition while tombstones reference the box, got %v", err)
	}

	for _, id := range []string{"cur-one", "cur-two"} {
		if err := agents.DeleteAgent(ctx, wsTest, id); err != nil {
			t.Fatalf("purge agent %s: %v", id, err)
		}
	}
	if _, err := boxSvc.DeleteButterBox(ctx, connect.NewRequest(&agentsv1.DeleteButterBoxRequest{Id: "box-1"})); err != nil {
		t.Fatalf("DeleteButterBox after purge: %v", err)
	}
}
