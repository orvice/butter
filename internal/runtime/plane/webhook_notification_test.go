package plane

import (
	"errors"
	"testing"
)

// Synthetic ordinary issue_comment shape, based on the generic envelope and
// the production classification, not a captured body. It is not Agent input.
func issueCommentNotification() map[string]any {
	return map[string]any{
		"event": "issue_comment", "action": "created", "workspace_id": "plane-ws", "webhook_id": "notification-1",
		"data":     map[string]any{"id": "comment-1", "workspace": "plane-ws", "comment_html": "Ordinary comment"},
		"activity": map[string]any{"actor": map[string]string{"id": "audit-author"}, "agent_run": map[string]string{"id": "audit-run"}, "signal": "stop"},
	}
}

func TestReceiverIgnoresAuthenticatedIssueComment(t *testing.T) {
	f := newReceiverFixture(t)
	decision, err := f.deliver(t, issueCommentNotification())
	if err != nil || decision != DecisionIgnored || len(f.queue.events) != 0 {
		t.Fatalf("ordinary notification treated as Agent input: decision=%s err=%v", decision, err)
	}
}

func TestIssueCommentCannotHideAgentClaims(t *testing.T) {
	for _, mode := range []string{"complete Agent", "missing workspace", "missing type", "unknown type", "case-folded Agent field", "flat Agent"} {
		t.Run(mode, func(t *testing.T) {
			f, payload := newReceiverFixture(t), officialEnvelope(t)
			payload["event"] = "issue_comment"
			business := payload["data"].(map[string]any)
			switch mode {
			case "missing workspace":
				delete(business, "workspace_id")
			case "missing type":
				delete(business, "type")
			case "unknown type":
				business["type"] = "unknown-agent-type"
			case "case-folded Agent field":
				delete(business, "type")
				business["TYPE"] = "agent_run_activity"
				delete(business, "workspace_id")
			case "flat Agent":
				payload = officialPrompt(t)
				payload["event"] = "issue_comment"
			}
			decision, err := f.deliver(t, payload)
			if mode == "complete Agent" || mode == "flat Agent" {
				if err != nil || decision != DecisionAccepted || len(f.queue.events) != 1 {
					t.Fatalf("outer label suppressed canonical Agent input: %s %v", decision, err)
				}
			} else if err == nil || decision == DecisionIgnored || len(f.queue.events) != 0 {
				t.Fatalf("Agent claim silently ignored or admitted: %s %v", decision, err)
			}
		})
	}
}

func TestIssueCommentFilterBoundaries(t *testing.T) {
	for _, body := range []string{
		`{"event":"issue_comment","type":"agent_run","data":{}}`,
		`{"event":"issue_comment","data":{"type":"unknown-type"}}`,
		`{"event":"issue_comment","agent_run":{},"data":{}}`,
		`{"event":"issue_comment","AGENT_RUN_ACTIVITY":{},"data":{}}`,
		`{"event":"issue_comment","data":{"activity":{}}}`,
		`{"event":"issue_comment","data":{"run":{}}}`,
		`{"event":"issue_comment","data":null}`,
		`{"event":"issue_comment","data":[]}`,
		`{"event":"unknown_event","data":{}}`,
		`{"event":"agent_run_user_prompt","data":{}}`,
		`{"event":"issue_comment","data":{}} invalid`,
	} {
		if isIssueCommentNotification([]byte(body)) {
			t.Fatalf("ambiguous/malformed payload silently ignored: %s", body)
		}
	}
}

func TestIssueCommentStillRequiresSignature(t *testing.T) {
	f := newReceiverFixture(t)
	_, err := f.receiver.Deliver(t.Context(), f.app, nil, []byte(`{"event":"issue_comment","data":{}}`))
	if !errors.Is(err, ErrUnauthorized) || len(f.queue.events) != 0 {
		t.Fatalf("unsigned notification acknowledged: %v", err)
	}
}
