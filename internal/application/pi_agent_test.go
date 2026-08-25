package application

import (
	"strings"
	"testing"

	"connectrpc.com/connect"

	butterboxrepo "go.orx.me/apps/butter/internal/repo/butterbox"
	butterboxmemory "go.orx.me/apps/butter/internal/repo/butterbox/memory"
	"go.orx.me/apps/butter/internal/repo/config/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func testPiAgent(agentID, boxID string) *agentsv1.Agent {
	return &agentsv1.Agent{
		Name:    agentID,
		AgentId: agentID,
		Type:    agentsv1.AgentType_AGENT_TYPE_PI,
		Config: &agentsv1.AgentConfig{
			Pi: &agentsv1.PiAgentConfig{ButterboxId: boxID, WorkingDir: "projects/demo"},
		},
	}
}

func seedButterBox(t *testing.T, repo butterboxrepo.Repository, id string, enabled bool) {
	t.Helper()
	if _, err := repo.Create(testCtx(), wsTest, &agentsv1.ButterBox{
		Id: id, Name: id, BaseUrl: "https://box.example.com", Enabled: enabled,
	}, butterboxrepo.Credential{}); err != nil {
		t.Fatalf("seed butterbox: %v", err)
	}
}

func TestCreatePiAgent(t *testing.T) {
	store := memory.New()
	boxes := butterboxmemory.New()
	seedButterBox(t, boxes, "box-1", true)
	seedButterBox(t, boxes, "box-disabled", false)
	svc := NewAgentServiceServer(store)
	svc.SetButterBoxRepo(boxes)
	ctx := testCtx()

	// Unknown box is rejected with a pointed error.
	_, err := svc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: testPiAgent("pi-a", "missing-box"),
	}))
	if err == nil || connect.CodeOf(err) != connect.CodeInvalidArgument || !strings.Contains(err.Error(), "missing-box") {
		t.Fatalf("expected InvalidArgument naming the box, got %v", err)
	}

	// A disabled box is not available for a new binding.
	_, err = svc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: testPiAgent("pi-disabled", "box-disabled"),
	}))
	if err == nil || connect.CodeOf(err) != connect.CodeInvalidArgument || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("expected InvalidArgument for disabled box, got %v", err)
	}

	// Box-owned behavior fields are rejected on write.
	bad := testPiAgent("pi-b", "box-1")
	bad.Config.Instruction = "be helpful"
	_, err = svc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{Agent: bad}))
	if err == nil || connect.CodeOf(err) != connect.CodeInvalidArgument || !strings.Contains(err.Error(), "instruction") {
		t.Fatalf("expected InvalidArgument naming instruction, got %v", err)
	}

	// A pi agent is a leaf.
	withChild := testPiAgent("pi-c", "box-1")
	withChild.ChildAgentIds = []string{"other"}
	_, err = svc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{Agent: withChild}))
	if err == nil || connect.CodeOf(err) != connect.CodeInvalidArgument || !strings.Contains(err.Error(), "child_agent_ids") {
		t.Fatalf("expected InvalidArgument naming child_agent_ids, got %v", err)
	}

	// A valid pi agent is accepted.
	resp, err := svc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: testPiAgent("pi-ok", " box-1 "),
	}))
	if err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}
	if got := resp.Msg.GetAgent().GetConfig().GetPi().GetButterboxId(); got != "box-1" {
		t.Fatalf("stored butterbox_id: got %q", got)
	}
}

func TestUpdatePiAgentValidates(t *testing.T) {
	store := memory.New()
	boxes := butterboxmemory.New()
	seedButterBox(t, boxes, "box-1", true)
	seedButterBox(t, boxes, "box-disabled", false)
	svc := NewAgentServiceServer(store)
	svc.SetButterBoxRepo(boxes)
	ctx := testCtx()

	if _, err := svc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: testPiAgent("pi-1", "box-1"),
	})); err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}

	// Repointing to an unknown box is rejected.
	update := testPiAgent("pi-1", "gone-box")
	_, err := svc.UpdateAgent(ctx, connect.NewRequest(&agentsv1.UpdateAgentRequest{Agent: update}))
	if err == nil || connect.CodeOf(err) != connect.CodeInvalidArgument || !strings.Contains(err.Error(), "gone-box") {
		t.Fatalf("expected InvalidArgument naming the box, got %v", err)
	}

	// Repointing to a disabled box is rejected.
	update = testPiAgent("pi-1", "box-disabled")
	_, err = svc.UpdateAgent(ctx, connect.NewRequest(&agentsv1.UpdateAgentRequest{Agent: update}))
	if err == nil || connect.CodeOf(err) != connect.CodeInvalidArgument || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("expected InvalidArgument for disabled box, got %v", err)
	}

	// Repointing the working directory is a legal update.
	update = testPiAgent("pi-1", "box-1")
	update.Config.Pi.WorkingDir = "projects/other"
	if _, err := svc.UpdateAgent(ctx, connect.NewRequest(&agentsv1.UpdateAgentRequest{Agent: update})); err != nil {
		t.Fatalf("UpdateAgent: %v", err)
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
	update = testPiAgent("pi-1", " box-1 ")
	update.Config.Pi.WorkingDir = "projects/after-disable"
	resp, err := svc.UpdateAgent(ctx, connect.NewRequest(&agentsv1.UpdateAgentRequest{Agent: update}))
	if err != nil {
		t.Fatalf("UpdateAgent existing disabled binding: %v", err)
	}
	if got := resp.Msg.GetAgent().GetConfig().GetPi().GetButterboxId(); got != "box-1" {
		t.Fatalf("stored butterbox_id: got %q", got)
	}
}

func TestDeleteButterBoxRefusedWhilePiAgentsReferenceIt(t *testing.T) {
	agents := memory.New()
	boxes := butterboxmemory.New()
	seedButterBox(t, boxes, "box-1", true)

	agentSvc := NewAgentServiceServer(agents)
	agentSvc.SetButterBoxRepo(boxes)
	boxSvc := NewButterBoxServiceServer(boxes)
	boxSvc.SetAgentRepo(agents)
	ctx := testCtx()

	for _, id := range []string{"pi-one", "pi-two"} {
		boxID := "box-1"
		if id == "pi-one" {
			boxID = " box-1 "
		}
		if _, err := agentSvc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{
			Agent: testPiAgent(id, boxID),
		})); err != nil {
			t.Fatalf("CreateAgent %s: %v", id, err)
		}
	}

	_, err := boxSvc.DeleteButterBox(ctx, connect.NewRequest(&agentsv1.DeleteButterBoxRequest{Id: "box-1"}))
	if err == nil || connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("expected FailedPrecondition, got %v", err)
	}
	if !strings.Contains(err.Error(), "pi-one") || !strings.Contains(err.Error(), "pi-two") {
		t.Fatalf("error must list the referencing agent ids, got %v", err)
	}

	// A soft delete leaves a tombstone that still holds the box reference —
	// restoring it must not resurrect a dangling binding, so the guard keeps
	// refusing.
	for _, id := range []string{"pi-one", "pi-two"} {
		if _, err := agentSvc.DeleteAgent(ctx, connect.NewRequest(&agentsv1.DeleteAgentRequest{AgentId: id})); err != nil {
			t.Fatalf("DeleteAgent %s: %v", id, err)
		}
	}
	if _, err := boxSvc.DeleteButterBox(ctx, connect.NewRequest(&agentsv1.DeleteButterBoxRequest{Id: "box-1"})); err == nil || connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("expected FailedPrecondition while tombstones reference the box, got %v", err)
	}

	// Only when the rows are gone does the delete go through.
	for _, id := range []string{"pi-one", "pi-two"} {
		if err := agents.DeleteAgent(ctx, wsTest, id); err != nil {
			t.Fatalf("purge agent %s: %v", id, err)
		}
	}
	if _, err := boxSvc.DeleteButterBox(ctx, connect.NewRequest(&agentsv1.DeleteButterBoxRequest{Id: "box-1"})); err != nil {
		t.Fatalf("DeleteButterBox after purge: %v", err)
	}
}
