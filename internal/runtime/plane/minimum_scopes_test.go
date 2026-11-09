package plane

import (
	"errors"
	"strings"
	"testing"

	installationrepo "go.orx.me/apps/butter/internal/repo/planeinstallation"
)

func TestReceiverDoesNotRequireRunWrite(t *testing.T) {
	f := newReceiverFixture(t)
	// The integration consumes Plane-created Runs; it never creates/updates one.
	f.installation.GrantedScopes = []string{"agents.runs:read", "agents.run_activities:read", "agents.run_activities:write"}
	if _, err := f.installations.Upsert(t.Context(), "ws", "app", f.installation, installationrepo.Tokens{}); err != nil {
		t.Fatal(err)
	}
	decision, err := f.deliver(t, officialEnvelope(t))
	if err != nil || decision != DecisionAccepted || len(f.queue.events) != 1 {
		t.Fatalf("unneeded Run write permission blocked mention: decision=%s err=%v", decision, err)
	}
}

func TestReceiverReportsAllMissingScopesWithoutGrantValues(t *testing.T) {
	f := newReceiverFixture(t)
	f.installation.GrantedScopes = []string{"private-unknown-grant"}
	if _, err := f.installations.Upsert(t.Context(), "ws", "app", f.installation, installationrepo.Tokens{}); err != nil {
		t.Fatal(err)
	}
	_, err := f.deliver(t, officialEnvelope(t))
	if !errors.Is(err, ErrIdentity) || len(f.queue.events) != 0 {
		t.Fatalf("missing permissions admitted: %v", err)
	}
	if !strings.Contains(err.Error(), "missing_scopes=agents.runs:read,agents.run_activities:read,agents.run_activities:write") {
		t.Fatalf("missing grant checklist is incomplete: %v", err)
	}
	if strings.Contains(err.Error(), "private-unknown-grant") {
		t.Fatal("diagnostic exposed unrecognized grant content")
	}
}

func TestReceiverIdentityDiagnosticBusinessClassification(t *testing.T) {
	for _, kind := range []string{"agent_run_activity", "issue", "private-type"} {
		t.Run(kind, func(t *testing.T) {
			f, payload := newReceiverFixture(t), officialEnvelope(t)
			business := payload["data"].(map[string]any)
			business["type"] = kind
			if kind == "private-type" {
				business["action"] = "private-action"
				payload["event"] = "private-event"
			}
			delete(business, "workspace_id") // Preserve rejection; never fill from wrapper.
			_, err := f.deliver(t, payload)
			if !errors.Is(err, ErrIdentity) || len(f.queue.events) != 0 {
				t.Fatalf("missing workspace admitted: %v", err)
			}
			wantKind, wantAction, wantEvent := kind, "prompted", "agent_run_user_prompt"
			if kind == "private-type" {
				wantKind, wantAction, wantEvent = "other", "other", "other"
			}
			if !strings.Contains(err.Error(), "payload_type="+wantKind+" payload_action="+wantAction+" event_kind="+wantEvent) {
				t.Fatalf("missing safe business classification: %v", err)
			}
			if strings.Contains(err.Error(), "private-") {
				t.Fatal("diagnostic exposed unknown type/action values")
			}
		})
	}
}
