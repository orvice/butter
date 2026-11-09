package plane

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json/v2"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"testing"

	"go.orx.me/apps/butter/internal/eventqueue"
	cryptokeymem "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	"go.orx.me/apps/butter/internal/repo/invocation"
	invocationmem "go.orx.me/apps/butter/internal/repo/invocation/memory"
	planerepo "go.orx.me/apps/butter/internal/repo/plane"
	planemem "go.orx.me/apps/butter/internal/repo/plane/memory"
	installationrepo "go.orx.me/apps/butter/internal/repo/planeinstallation"
	installationmem "go.orx.me/apps/butter/internal/repo/planeinstallation/memory"
	processingmem "go.orx.me/apps/butter/internal/repo/planeprocessing/memory"
	"go.orx.me/apps/butter/internal/secretbox"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/adk/v2/agent"
)

// Adapted from https://developers.plane.so/dev-tools/agents/building-an-agent
// with synthetic IDs. In particular, the official example has no installation ID.
func officialPrompt(t *testing.T) map[string]any {
	t.Helper()
	body, err := os.ReadFile("testdata/agent_run_user_prompt.json")
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}
	return payload
}

type receiverFixture struct {
	receiver      *Receiver
	app           *agentsv1.PlaneApp
	installations *installationmem.Store
	installation  *agentsv1.PlaneInstallation
	queue         *admissionQueue
}

func newReceiverFixture(t *testing.T) *receiverFixture {
	t.Helper()
	apps, installations := planemem.New(), installationmem.New()
	ring := secretbox.NewKeyring(cryptokeymem.New())
	ciphertext, keyID, err := ring.Encrypt(t.Context(), []byte("fixture-webhook-secret"))
	if err != nil {
		t.Fatal(err)
	}
	app, err := apps.CreateApp(t.Context(), "ws", &agentsv1.PlaneApp{
		Id: "app", ClientId: "client", AgentId: "agent", InboundEnabled: true,
	}, planerepo.AppCredentials{WebhookSecret: planerepo.Credential{Ciphertext: ciphertext, KeyID: keyID}})
	if err != nil {
		t.Fatal(err)
	}
	inst, err := installations.Upsert(t.Context(), "ws", "app", &agentsv1.PlaneInstallation{
		AppInstallationId: "external-installation", PlaneWorkspaceId: "plane-ws", WorkspaceSlug: "plane-workspace",
		AppBotUserId: "bot", GrantedScopes: requiredAgentScopes, Status: "installed",
	}, installationrepo.Tokens{})
	if err != nil {
		t.Fatal(err)
	}
	q := &admissionQueue{}
	return &receiverFixture{NewReceiver(apps, installations, ring, q), app, installations, inst, q}
}

func (f *receiverFixture) deliver(t *testing.T, payload map[string]any) (Decision, error) {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return f.deliverBody(t, body)
}

func (f *receiverFixture) deliverBody(t *testing.T, body []byte) (Decision, error) {
	t.Helper()
	mac := hmac.New(sha256.New, []byte("fixture-webhook-secret"))
	_, _ = mac.Write(body)
	headers := http.Header{}
	headers.Set(SignatureHeader, hex.EncodeToString(mac.Sum(nil)))
	return f.receiver.Deliver(t.Context(), f.app, headers, body)
}

func TestReceiverRejectsUntrustedPlaneIdentity(t *testing.T) {
	cases := []struct {
		name, reason string
		change       func(map[string]any)
	}{
		{"missing client", "missing", func(p map[string]any) { delete(p, "app_client_id") }},
		{"wrong client", "app_client_mismatch", func(p map[string]any) { p["app_client_id"] = "foreign-client" }},
		{"missing workspace", "missing", func(p map[string]any) { delete(p, "workspace_id") }},
		{"missing bot", "missing", func(p map[string]any) { delete(p, "agent_user_id") }},
		{"wrong installation", "installation_not_found", func(p map[string]any) { p["app_installation_id"] = "foreign-installation" }},
		{"wrong workspace", "workspace_mismatch", func(p map[string]any) {
			p["workspace_id"] = "foreign-workspace"
			p["agent_run"].(map[string]any)["workspace"] = "foreign-workspace"
			p["agent_run_activity"].(map[string]any)["workspace"] = "foreign-workspace"
		}},
		{"wrong bot", "bot_user_mismatch", func(p map[string]any) {
			p["agent_user_id"] = "foreign-bot"
			p["agent_run"].(map[string]any)["agent_user"] = "foreign-bot"
		}},
		{"run workspace", "conflict", func(p map[string]any) { p["agent_run"].(map[string]any)["workspace"] = "foreign-workspace" }},
		{"run workspace alias", "conflict", func(p map[string]any) { p["agent_run"].(map[string]any)["workspace_id"] = "foreign-workspace" }},
		{"run bot", "conflict", func(p map[string]any) { p["agent_run"].(map[string]any)["agent_user"] = "foreign-bot" }},
		{"activity workspace", "conflict", func(p map[string]any) { p["agent_run_activity"].(map[string]any)["workspace"] = "foreign-workspace" }},
		{"activity run", "conflict", func(p map[string]any) { p["agent_run_activity"].(map[string]any)["agent_run"] = "foreign-run" }},
		{"missing actor", "missing", func(p map[string]any) { delete(p["agent_run_activity"].(map[string]any), "actor") }},
		{"bot actor", "actor_is_bot", func(p map[string]any) { p["agent_run_activity"].(map[string]any)["actor"] = "bot" }},
		{"conflicting actor", "conflict", func(p map[string]any) { p["agent_run_activity"].(map[string]any)["user_id"] = "foreign-actor" }},
		{"conflicting envelope client", "conflict", func(p map[string]any) { p["data"] = map[string]any{"app_client_id": "foreign-client"} }},
		{"conflicting run aliases", "conflict", func(p map[string]any) { p["run"] = map[string]any{"id": "foreign-run"} }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, payload := newReceiverFixture(t), officialPrompt(t)
			tc.change(payload)
			_, err := f.deliver(t, payload)
			if !errors.Is(err, ErrIdentity) || !strings.Contains(err.Error(), "reason="+tc.reason) {
				t.Fatalf("want identity reject %s, got %v", tc.reason, err)
			}
			if len(f.queue.events) != 0 {
				t.Fatal("untrusted event reached queue")
			}
			for _, sensitive := range []string{"foreign-", "What is the weather", "fixture-webhook-secret"} {
				if strings.Contains(err.Error(), sensitive) {
					t.Fatalf("diagnostic leaks data: %v", err)
				}
			}
		})
	}
}

func TestOfficialPlanePromptReachesInvocationAndActivity(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })
	f := newReceiverFixture(t)
	records, invocations := processingmem.New(), invocationmem.New()
	f.receiver.SetProcessingRepo(records)
	p, _, app := invocationProcessor(t, func(agent.InvocationContext) (string, error) { return "Weather report", nil }, invocations)
	f.app.AgentId = app.GetAgentId()
	posted := 0
	client := NewActivityHTTPClient("https://plane.invalid", &http.Client{Transport: activityRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		posted++
		if r.URL.Path != "/api/v1/workspaces/plane-workspace/runs/run-1/activities/" {
			t.Errorf("wrong activity URL: %s", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), `"type":"response"`) || !strings.Contains(string(body), "Weather report") {
			t.Errorf("incorrect Plane response: %s", body)
		}
		return &http.Response{StatusCode: 201, Body: io.NopCloser(strings.NewReader(`{"id":"response-1"}`))}, nil
	})}, func(_ context.Context, ws, app, inst string) (string, error) {
		if ws != "ws" || app != "app" || inst != "external-installation" {
			t.Errorf("wrong token scope: %s/%s/%s", ws, app, inst)
		}
		return "fixture-bot-token", nil
	})
	p.publisher = NewDurablePublisher(ActivityReplyPublisher{Client: client, Installations: f.installations}, NewMemoryDeliveryStore())
	p.SetProcessingRepository(records)
	if decision, err := f.deliver(t, officialPrompt(t)); err != nil || decision != DecisionAccepted {
		t.Fatalf("admission=%s err=%v", decision, err)
	}
	e := f.queue.events[0]
	lookup := &agentsv1.PlaneProcessingRecord{WorkspaceId: e.WorkspaceID, AppId: e.AppID, InstallationId: e.InstallationID, RunId: e.AgentRunID, ActivityId: e.ActivityID, EventAction: e.Action}
	row, err := records.FindByEvent(t.Context(), lookup)
	if err != nil || row.GetStatus() != agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_RECEIVED {
		t.Fatalf("missing receive projection: row=%v err=%v", row, err)
	}
	payload, err := e.Encode()
	if err != nil {
		t.Fatal(err)
	}
	w := &Worker{processor: p, resolver: &planeTestResolver{app: f.app}}
	if err := w.handle(t.Context(), eventqueue.Entry{ID: "1-0", Payload: payload}); err != nil {
		t.Fatal(err)
	}
	row, err = records.FindByEvent(t.Context(), lookup)
	if err != nil || row.GetStatus() != agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_SUCCEEDED || row.GetInvocationId() == "" || row.GetReplyDeliveryStatus() != "delivered" || posted != 1 {
		t.Fatalf("pipeline did not finish: row=%v posts=%d err=%v", row, posted, err)
	}
	invs, _, _, err := invocations.List(t.Context(), invocation.ListFilter{WorkspaceID: "ws"}, 10, "")
	if err != nil || len(invs) != 1 || invs[0].GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED {
		t.Fatalf("invocations=%v err=%v", invs, err)
	}
	// Only the new Plane diagnostics are covered here; the shared runner has
	// its own pre-existing logging contract for session/invocation identifiers.
	var planeLogs strings.Builder
	for line := range strings.SplitSeq(logs.String(), "\n") {
		var entry struct {
			Message string `json:"msg"`
		}
		if json.Unmarshal([]byte(line), &entry) == nil && strings.HasPrefix(entry.Message, "plane ") {
			planeLogs.WriteString(line)
		}
	}
	for _, message := range []string{"plane webhook queued", "plane event processing", "plane invocation starting", "plane invocation settled", "plane reply delivered"} {
		if !strings.Contains(planeLogs.String(), message) {
			t.Errorf("missing lifecycle log %q", message)
		}
	}
	for _, sensitive := range []string{"external-installation", "run-1", "activity-1", "user-1", "What is the weather", "Weather report", "fixture-bot-token", "fixture-webhook-secret"} {
		if strings.Contains(planeLogs.String(), sensitive) {
			t.Errorf("logs contain sensitive value %q", sensitive)
		}
	}
}

func TestReceiverInstallationBindingAndScopes(t *testing.T) {
	for _, mode := range []string{"explicit", "reinstalled", "stale explicit", "missing scopes", "uninstalled", "disallowed actor"} {
		t.Run(mode, func(t *testing.T) {
			f, payload := newReceiverFixture(t), officialPrompt(t)
			switch mode {
			case "explicit":
				payload["app_installation_id"] = f.installation.GetAppInstallationId()
			case "reinstalled", "stale explicit":
				f.installation.AppInstallationId = "new-installation"
				f.installation.AppBotUserId = "new-bot"
				payload["agent_user_id"] = "new-bot"
				payload["agent_run"].(map[string]any)["agent_user"] = "new-bot"
				if mode == "stale explicit" {
					payload["app_installation_id"] = "external-installation"
				}
			case "missing scopes":
				f.installation.GrantedScopes = []string{"agents.runs:read"}
			case "uninstalled":
				f.installation.Status = "uninstalled"
			case "disallowed actor":
				f.app.AllowedUserIds = []string{"other-user"}
			}
			if _, err := f.installations.Upsert(t.Context(), "ws", "app", f.installation, installationrepo.Tokens{}); err != nil {
				t.Fatal(err)
			}
			decision, err := f.deliver(t, payload)
			switch mode {
			case "explicit", "reinstalled":
				if err != nil || decision != DecisionAccepted {
					t.Fatalf("admission=%s err=%v", decision, err)
				}
				if f.queue.events[0].InstallationID != f.installation.GetId() || f.queue.events[0].PlaneInstallationID != f.installation.GetAppInstallationId() {
					t.Fatal("reinstallation binding was not refreshed")
				}
			default:
				reasons := map[string]string{"stale explicit": "installation_not_found", "missing scopes": "scope_incomplete", "uninstalled": "installation_inactive", "disallowed actor": "actor_not_allowed"}
				if !errors.Is(err, ErrIdentity) || !strings.Contains(err.Error(), "reason="+reasons[mode]) || len(f.queue.events) != 0 {
					t.Fatalf("admission=%s err=%v", decision, err)
				}
			}
		})
	}
}

func TestReceiverOfficialCreateAndStop(t *testing.T) {
	for _, kind := range []string{"create", "stop"} {
		t.Run(kind, func(t *testing.T) {
			f, payload := newReceiverFixture(t), officialPrompt(t)
			if kind == "create" {
				delete(payload, "agent_run_activity")
				payload["type"], payload["action"] = "agent_run", "created"
			} else {
				activity := payload["agent_run_activity"].(map[string]any)
				activity["signal"] = "stop"
				delete(activity, "content")
				f.app.InboundEnabled = false // Stop remains available.
			}
			decision, err := f.deliver(t, payload)
			if kind == "create" {
				if err != nil || decision != DecisionIgnored || len(f.queue.events) != 0 {
					t.Fatalf("create admission=%s err=%v", decision, err)
				}
			} else if err != nil || decision != DecisionAccepted || len(f.queue.events) != 1 || !f.queue.events[0].Stop || f.queue.events[0].ActorUserID != "user-1" {
				t.Fatalf("stop admission=%s err=%v", decision, err)
			}
		})
	}
}

func TestWebhookInstallationResolutionRejectsAmbiguity(t *testing.T) {
	f := newReceiverFixture(t)
	p := &webhookPayload{WorkspaceID: "plane-ws", AgentUserID: "bot"}
	_, err := resolveWebhookInstallation(p, []*agentsv1.PlaneInstallation{f.installation, f.installation})
	if !errors.Is(err, ErrIdentity) || !strings.Contains(err.Error(), "reason=installation_ambiguous") {
		t.Fatalf("ambiguous installation resolved: %v", err)
	}
	_, err = resolveWebhookInstallation(p, nil)
	if !errors.Is(err, ErrIdentity) || !strings.Contains(err.Error(), "reason=installation_not_found") {
		t.Fatalf("missing installation resolved: %v", err)
	}
}

func TestWebhookIdentityDiagnosticsContainOnlyFingerprints(t *testing.T) {
	expected, actual := "expected-sensitive-id", "received-sensitive-id"
	err := rejectWebhookIdentity("conflict", "workspace_id", expected, actual)
	if strings.Contains(err.Error(), expected) || strings.Contains(err.Error(), actual) {
		t.Fatal("diagnostic contains raw ID")
	}
	for _, value := range []string{expected, actual} {
		digest := sha256.Sum256([]byte(value))
		if !strings.Contains(err.Error(), hex.EncodeToString(digest[:6])) {
			t.Fatal("diagnostic missing identity fingerprint")
		}
	}
	if !strings.Contains(err.Error(), "expected_present=true") || !strings.Contains(err.Error(), "actual_present=true") {
		t.Fatal("diagnostic missing presence indicators")
	}
}

func TestReceiverOfficialPlanePayload(t *testing.T) {
	for _, envelope := range []bool{false, true} {
		t.Run(map[bool]string{false: "direct", true: "data envelope"}[envelope], func(t *testing.T) {
			f := newReceiverFixture(t)
			payload := officialPrompt(t)
			if envelope {
				payload = map[string]any{"event": "agent_run_user_prompt", "data": payload}
			}
			decision, err := f.deliver(t, payload)
			if err != nil || decision != DecisionAccepted {
				t.Fatalf("official prompt rejected before queue: decision=%s err=%v", decision, err)
			}
			if len(f.queue.events) != 1 {
				t.Fatalf("queued events=%d", len(f.queue.events))
			}
			e := f.queue.events[0]
			if e.AgentRunID != "run-1" || e.ActivityID != "activity-1" || e.ActorUserID != "user-1" || e.Prompt != "What is the weather like?" || e.PlaneInstallationID != "external-installation" {
				t.Fatalf("incorrect admitted identity: %+v", e)
			}
		})
	}
}
