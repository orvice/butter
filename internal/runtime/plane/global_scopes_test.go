package plane

import (
	"errors"
	"slices"
	"testing"

	installationrepo "go.orx.me/apps/butter/internal/repo/planeinstallation"
)

func TestGlobalScopesNeverBypassIdentity(t *testing.T) {
	for _, field := range []string{"client", "workspace", "bot", "installation", "missing actor", "bot actor", "allowlist", "run reference"} {
		t.Run(field, func(t *testing.T) {
			f, payload := newReceiverFixture(t), officialEnvelope(t)
			f.installation.GrantedScopes = []string{"read", "write"}
			if _, err := f.installations.Upsert(t.Context(), "ws", "app", f.installation, installationrepo.Tokens{}); err != nil {
				t.Fatal(err)
			}
			data := payload["data"].(map[string]any)
			activity := data["agent_run_activity"].(map[string]any)
			switch field {
			case "client":
				data["app_client_id"] = "foreign-client"
			case "workspace":
				payload["workspace_id"], data["workspace_id"] = "foreign-workspace", "foreign-workspace"
				data["agent_run"].(map[string]any)["workspace"] = "foreign-workspace"
				activity["workspace"] = "foreign-workspace"
			case "bot":
				data["agent_user_id"] = "foreign-bot"
				data["agent_run"].(map[string]any)["agent_user"] = "foreign-bot"
			case "installation":
				data["app_installation_id"] = "foreign-installation"
			case "missing actor":
				delete(activity, "actor")
			case "bot actor":
				activity["actor"] = "bot"
			case "allowlist":
				f.app.AllowedUserIds = []string{"audit-author"}
			case "run reference":
				activity["agent_run"] = "foreign-run"
			}
			_, err := f.deliver(t, payload)
			if !errors.Is(err, ErrIdentity) || len(f.queue.events) != 0 {
				t.Fatalf("global grant bypassed %s: %v", field, err)
			}
		})
	}
	f, payload := newReceiverFixture(t), officialEnvelope(t)
	f.installation.GrantedScopes = nil
	if _, err := f.installations.Upsert(t.Context(), "ws", "app", f.installation, installationrepo.Tokens{}); err != nil {
		t.Fatal(err)
	}
	payload["data"].(map[string]any)["granted_scopes"] = []string{"read", "write"}
	_, err := f.deliver(t, payload)
	if !errors.Is(err, ErrIdentity) || len(f.queue.events) != 0 {
		t.Fatalf("payload-supplied grants authorized execution: %v", err)
	}
}

func TestReceiverGlobalAndGranularScopes(t *testing.T) {
	for _, tc := range []struct {
		name     string
		scopes   []string
		accepted bool
	}{
		{"global read write", []string{"read", "write"}, true},
		{"granular minimum", []string{"agents.runs:read", "agents.run_activities:read", "agents.run_activities:write"}, true},
		{"global read and granular write", []string{"read", "agents.run_activities:write"}, true},
		{"granular reads and global write", []string{"agents.runs:read", "agents.run_activities:read", "write"}, true},
		{"global read only", []string{"read"}, false},
		{"global write only", []string{"write"}, false},
		{"unrelated granular permissions", []string{"projects:read", "projects:write"}, false},
		{"lookalikes", []string{"bread", "overwrite", "READ", "WRITE", "read write", "*"}, false},
		{"missing grants", nil, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newReceiverFixture(t)
			f.installation.GrantedScopes = tc.scopes
			if _, err := f.installations.Upsert(t.Context(), "ws", "app", f.installation, installationrepo.Tokens{}); err != nil {
				t.Fatal(err)
			}
			decision, err := f.deliver(t, officialEnvelope(t))
			if tc.accepted {
				if err != nil || decision != DecisionAccepted || len(f.queue.events) != 1 {
					t.Fatalf("valid grants rejected: decision=%s err=%v", decision, err)
				}
			} else if !errors.Is(err, ErrIdentity) || len(f.queue.events) != 0 {
				t.Fatalf("insufficient grants admitted: decision=%s err=%v", decision, err)
			}
			stored, err := f.installations.GetScoped(t.Context(), "ws", "app", "external-installation")
			if err != nil || !slices.Equal(stored.GetGrantedScopes(), tc.scopes) {
				t.Fatal("coverage check rewrote provider grant metadata")
			}
		})
	}
}
