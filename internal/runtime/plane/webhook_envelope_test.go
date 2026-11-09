package plane

import (
	"encoding/json/v2"
	"errors"
	"os"
	"strings"
	"testing"

	installationrepo "go.orx.me/apps/butter/internal/repo/planeinstallation"
)

// Synthetic shape: official prd-agent WebhookEvent.data/optional activity,
// plus the object kinds observed in production diagnostics. Not a body capture.
func officialEnvelope(t *testing.T) map[string]any {
	t.Helper()
	raw, err := os.ReadFile("testdata/agent_run_user_prompt_envelope.json")
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	return payload
}

func TestReceiverSeparatesEnvelopeMetadata(t *testing.T) {
	for _, shape := range []string{"expanded actor", "expanded run", "no metadata", "arbitrary metadata"} {
		t.Run(shape, func(t *testing.T) {
			f, payload := newReceiverFixture(t), officialEnvelope(t)
			switch shape {
			case "expanded run":
				payload["activity"].(map[string]any)["actor"] = "audit-author"
			case "no metadata":
				delete(payload, "activity")
			case "arbitrary metadata":
				payload["activity"] = map[string]any{"actor": []any{"not", "an", "identity"}, "agent_run": true, "content": map[string]any{"body": map[string]any{"private": "metadata"}}}
			}
			decision, err := f.deliver(t, payload)
			if err != nil || decision != DecisionAccepted || len(f.queue.events) != 1 {
				t.Fatalf("envelope rejected before queue: decision=%s err=%v", decision, err)
			}
			e := f.queue.events[0]
			if e.ActivityID != "activity-1" || e.ActorUserID != "user-1" || e.AgentRunID != "run-1" || e.Prompt != "What is the weather like?" || e.Signal != "continue" || e.Stop {
				t.Fatalf("transport metadata polluted the admitted prompt: %+v", e)
			}
		})
	}
}

func TestReceiverEnvelopeCannotSupplyMissingBusinessIdentity(t *testing.T) {
	for _, field := range []string{"actor", "workspace", "client", "bot", "run", "activity", "type", "action"} {
		t.Run(field, func(t *testing.T) {
			f, payload := newReceiverFixture(t), officialEnvelope(t)
			business := payload["data"].(map[string]any)
			activity := business["agent_run_activity"].(map[string]any)
			switch field {
			case "actor":
				delete(activity, "actor")
			case "workspace":
				delete(business, "workspace_id")
			case "client":
				delete(business, "app_client_id")
				payload["app_client_id"] = "client"
			case "bot":
				delete(business, "agent_user_id")
				payload["agent_user_id"] = "bot"
			case "run":
				delete(business, "agent_run")
			case "activity":
				delete(business, "agent_run_activity")
			case "type":
				delete(business, "type")
				payload["type"] = "agent_run_activity"
			case "action":
				delete(business, "action")
				payload["action"] = "prompted"
			}
			_, err := f.deliver(t, payload)
			if err == nil || len(f.queue.events) != 0 {
				t.Fatalf("envelope manufactured missing %s: %v", field, err)
			}
		})
	}
}

func TestReceiverEnvelopeBusinessPolicies(t *testing.T) {
	for _, tc := range []struct {
		name     string
		decision Decision
		wantErr  error
	}{
		{"create", DecisionIgnored, nil},
		{"stop while disabled", DecisionAccepted, nil},
		{"prompt while disabled", DecisionIgnored, nil},
		{"disallowed actor", "", ErrIdentity},
		{"bot actor", "", ErrIdentity},
		{"missing prompt", "", ErrMalformed},
		{"business actor object", "", ErrMalformed},
		{"business run object", "", ErrMalformed},
		{"outer run conflict", "", ErrIdentity},
		{"outer actor conflict", "", ErrIdentity},
		{"inner alias conflict", "", ErrIdentity},
		{"stale outer installation", "", ErrIdentity},
		{"valid outer installation", DecisionAccepted, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f, payload := newReceiverFixture(t), officialEnvelope(t)
			business := payload["data"].(map[string]any)
			activity := business["agent_run_activity"].(map[string]any)
			switch tc.name {
			case "create":
				payload["event"] = "agent_run_create"
				business["type"], business["action"] = "agent_run", "created"
				delete(business, "agent_run_activity")
			case "stop while disabled":
				f.app.InboundEnabled = false
				activity["signal"] = "stop"
				delete(activity, "content")
			case "prompt while disabled":
				f.app.InboundEnabled = false
			case "disallowed actor":
				f.app.AllowedUserIds = []string{"not-the-prompt-actor"}
			case "bot actor":
				activity["actor"] = "bot"
			case "missing prompt":
				delete(activity, "content")
			case "business actor object":
				activity["actor"] = map[string]string{"id": "user-1"}
			case "business run object":
				activity["agent_run"] = map[string]string{"id": "run-1"}
			case "outer run conflict":
				payload["agent_run"] = map[string]string{"id": "conflicting-run"}
			case "outer actor conflict":
				payload["agent_run_activity"] = map[string]string{"actor": "conflicting-actor"}
			case "inner alias conflict":
				business["activity"] = map[string]string{"id": "conflicting-activity"}
			case "stale outer installation":
				payload["app_installation_id"] = "stale-installation"
			case "valid outer installation":
				payload["app_installation_id"] = "external-installation"
			}
			decision, err := f.deliver(t, payload)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) || len(f.queue.events) != 0 {
					t.Fatalf("decision=%s err=%v queued=%d", decision, err, len(f.queue.events))
				}
				return
			}
			if err != nil || decision != tc.decision {
				t.Fatalf("decision=%s err=%v", decision, err)
			}
			if tc.decision == DecisionAccepted {
				if len(f.queue.events) != 1 || f.queue.events[0].Stop != (tc.name == "stop while disabled") {
					t.Fatal("wrong input/control admission")
				}
			} else if len(f.queue.events) != 0 {
				t.Fatal("ignored notification entered the queue")
			}
		})
	}
}

func TestReceiverEnvelopeRequiresDocumentedScopes(t *testing.T) {
	// Deliberately independent of requiredAgentScopes so misspelled production
	// requirements cannot silently produce equally misspelled test grants.
	scopes := []string{"agents.runs:read", "agents.runs:write", "agents.run_activities:read", "agents.run_activities:write"}
	for _, missing := range append(scopes, "undocumented aliases") {
		t.Run(missing, func(t *testing.T) {
			f := newReceiverFixture(t)
			f.installation.GrantedScopes = nil
			for _, scope := range scopes {
				if scope != missing {
					f.installation.GrantedScopes = append(f.installation.GrantedScopes, scope)
				}
			}
			if missing == "undocumented aliases" {
				f.installation.GrantedScopes = []string{"agents.runs:read", "agents.runs:write", "agents.activities:read", "agents.activities:write"}
			}
			if _, err := f.installations.Upsert(t.Context(), "ws", "app", f.installation, installationrepo.Tokens{}); err != nil {
				t.Fatal(err)
			}
			_, err := f.deliver(t, officialEnvelope(t))
			if !errors.Is(err, ErrIdentity) || !strings.Contains(err.Error(), "reason=scope_incomplete") || len(f.queue.events) != 0 {
				t.Fatalf("incomplete real grant admitted: %v", err)
			}
		})
	}
}

func TestReceiverEnvelopeRejectsIdentityConflicts(t *testing.T) {
	for _, field := range []string{"workspace_id", "app_client_id", "agent_user_id", "app_installation_id"} {
		t.Run(field, func(t *testing.T) {
			f, payload := newReceiverFixture(t), officialEnvelope(t)
			payload[field] = "conflicting-id"
			// Optional installation IDs must also agree when declared at both levels.
			if field == "app_installation_id" {
				payload["data"].(map[string]any)[field] = "external-installation"
			}
			_, err := f.deliver(t, payload)
			if !errors.Is(err, ErrIdentity) || len(f.queue.events) != 0 {
				t.Fatalf("conflicting %s was not rejected: %v", field, err)
			}
		})
	}
}
