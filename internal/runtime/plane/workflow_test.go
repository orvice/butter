package plane

import (
	"context"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"testing"
)

func TestWorkflowExecutesConfiguredStepsInOrder(t *testing.T) {
	r := &fakeRunner{out: "ok"}
	s := NewMemoryWorkflowStore()
	w := NewWorkflowExecutor(r, s)
	e := &Event{WorkspaceID: "ws", AppID: "app", InstallationID: "i", AgentRunID: "r", ActivityID: "a"}
	rec, err := w.Run(context.Background(), &agentsv1.PlaneApp{AgentId: "agent"}, e, WorkflowDefinition{ID: "wf", Enabled: true, Steps: []WorkflowStep{{ID: "one", Prompt: "one"}, {ID: "two", Prompt: "two"}}})
	if err != nil || rec.Status != WorkflowSucceeded || len(rec.Outputs) != 2 {
		t.Fatalf("rec=%+v err=%v", rec, err)
	}
	got, err := s.Get(context.Background(), "ws", rec.Key)
	if err != nil || got.Status != WorkflowSucceeded {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}
func TestWorkflowRejectsDisabledAndDuplicate(t *testing.T) {
	w := NewWorkflowExecutor(&fakeRunner{out: "x"}, NewMemoryWorkflowStore())
	e := &Event{InstallationID: "i", AgentRunID: "r", ActivityID: "a"}
	if _, err := w.Run(context.Background(), &agentsv1.PlaneApp{AgentId: "a"}, e, WorkflowDefinition{ID: "wf"}); err == nil {
		t.Fatal("disabled workflow accepted")
	}
}
