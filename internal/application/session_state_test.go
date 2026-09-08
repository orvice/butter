package application

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestSessionToInfoPreservesStateWithBSONDocument(t *testing.T) {
	sess := &fakeSession{
		id: "chat-1",
		state: &fakeState{data: map[string]any{
			"agent_id":   "pi-agent",
			"agent_name": "PiAgent",
			"pibox:pi-agent": bson.D{
				{Key: "pi_session_id", Value: "pi-session-1"},
				{Key: "butterbox_id", Value: "box-1"},
				{Key: "working_dir", Value: "/workspace"},
			},
		}},
	}

	state := sessionToInfo(sess).GetState()
	if state == nil {
		t.Fatal("session state was dropped")
	}
	got := state.AsMap()
	if got["agent_id"] != "pi-agent" {
		t.Fatalf("agent_id = %v, want pi-agent", got["agent_id"])
	}
	binding, ok := got["pibox:pi-agent"].(map[string]any)
	if !ok {
		t.Fatalf("Pi binding type = %T, want map[string]any", got["pibox:pi-agent"])
	}
	if binding["pi_session_id"] != "pi-session-1" {
		t.Fatalf("pi_session_id = %v, want pi-session-1", binding["pi_session_id"])
	}
}
