package http

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json/v2"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	cryptokeymem "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	planerepo "go.orx.me/apps/butter/internal/repo/plane"
	planemem "go.orx.me/apps/butter/internal/repo/plane/memory"
	installationrepo "go.orx.me/apps/butter/internal/repo/planeinstallation"
	installationmem "go.orx.me/apps/butter/internal/repo/planeinstallation/memory"
	planeruntime "go.orx.me/apps/butter/internal/runtime/plane"
	"go.orx.me/apps/butter/internal/secretbox"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type planeWebhookQueue struct{ calls int }

func (q *planeWebhookQueue) Accept(context.Context, *planeruntime.Event) (string, error) {
	q.calls++
	return "1-0", nil
}

func TestPlaneWebhookOfficialPayloadAndSafeDiagnostics(t *testing.T) {
	for _, mode := range []string{"accepted", "accepted envelope", "missing identity", "conflicting identity", "invalid signature", "malformed type", "malformed syntax"} {
		t.Run(mode, func(t *testing.T) {
			var logs bytes.Buffer
			previous := slog.Default()
			slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
			t.Cleanup(func() { slog.SetDefault(previous) })
			apps, installations := planemem.New(), installationmem.New()
			ring := secretbox.NewKeyring(cryptokeymem.New())
			const secret = "private-webhook-secret"
			ciphertext, keyID, err := ring.Encrypt(t.Context(), []byte(secret))
			if err != nil {
				t.Fatal(err)
			}
			app, err := apps.CreateApp(t.Context(), "ws", &agentsv1.PlaneApp{Id: "private-app-id", ClientId: "client", AgentId: "agent", InboundEnabled: true}, planerepo.AppCredentials{WebhookSecret: planerepo.Credential{Ciphertext: ciphertext, KeyID: keyID}})
			if err != nil {
				t.Fatal(err)
			}
			_, err = installations.Upsert(t.Context(), "ws", app.GetId(), &agentsv1.PlaneInstallation{
				AppInstallationId: "private-installation-id", PlaneWorkspaceId: "plane-ws", AppBotUserId: "bot", Status: "installed",
				GrantedScopes: []string{"agents.runs:read", "agents.run_activities:read", "agents.run_activities:write"},
			}, installationrepo.Tokens{})
			if err != nil {
				t.Fatal(err)
			}
			fixture := "../../runtime/plane/testdata/agent_run_user_prompt.json"
			if mode == "accepted envelope" {
				fixture = "../../runtime/plane/testdata/agent_run_user_prompt_envelope.json"
			}
			raw, err := os.ReadFile(fixture)
			if err != nil {
				t.Fatal(err)
			}
			var payload map[string]any
			if err := json.Unmarshal(raw, &payload); err != nil {
				t.Fatal(err)
			}
			switch mode {
			case "missing identity":
				delete(payload, "agent_user_id")
			case "conflicting identity":
				payload["agent_run_activity"].(map[string]any)["user_id"] = "private-conflicting-actor"
			case "malformed type":
				payload["agent_run_activity"].(map[string]any)["actor"] = map[string]string{"id": "private-expanded-actor"}
			}
			body, err := json.Marshal(payload)
			if err != nil {
				t.Fatal(err)
			}
			if mode == "malformed syntax" {
				body = append(body, []byte(" private-invalid-token")...)
			}
			mac := hmac.New(sha256.New, []byte(secret))
			_, _ = mac.Write(body)
			signature := hex.EncodeToString(mac.Sum(nil))
			if mode == "invalid signature" {
				signature = strings.Repeat("0", 64)
			}
			q := &planeWebhookQueue{}
			router := gin.New()
			NewPlaneWebhookHandler(planeruntime.NewReceiver(apps, installations, ring, q)).Register(router)
			req := httptest.NewRequest(http.MethodPost, "/api/plane/webhook/"+app.GetId(), bytes.NewReader(body))
			req.Header.Set(planeruntime.SignatureHeader, signature)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)
			if strings.HasPrefix(mode, "accepted") {
				if response.Code != http.StatusOK || q.calls != 1 || !strings.Contains(logs.String(), `"decision":"accepted"`) {
					t.Fatalf("status=%d queued=%d logs=%s", response.Code, q.calls, logs.String())
				}
			} else {
				wantStatus, wantResponse := http.StatusUnauthorized, `{"error":"unauthorized"}`
				if strings.HasPrefix(mode, "malformed ") {
					wantStatus, wantResponse = http.StatusBadRequest, `{"error":"malformed webhook"}`
				}
				if response.Code != wantStatus || q.calls != 0 || response.Body.String() != wantResponse {
					t.Fatalf("status=%d queued=%d response=%s", response.Code, q.calls, response.Body)
				}
				reasons := map[string]string{
					"missing identity":     "reason=missing field=agent_user_id",
					"conflicting identity": "reason=conflict field=activity.actor",
					"invalid signature":    "invalid plane webhook signature",
					"malformed type":       "reason=json_type field=agent_run_activity.actor expected_type=string actual_kind=object offset=",
					"malformed syntax":     "reason=json_syntax field=$ expected_type=unknown actual_kind=unknown offset=",
				}
				if !strings.Contains(logs.String(), reasons[mode]) {
					t.Fatalf("missing reject diagnosis: %s", logs.String())
				}
			}
			for _, sensitive := range []string{secret, signature, string(body), "private-app-id", "private-installation-id", "private-conflicting-actor", "private-expanded-actor", "private-invalid-token", "What is the weather", "audit-event-1", "audit-author", "audit-resource", "Audit metadata is not a prompt"} {
				if strings.Contains(logs.String(), sensitive) {
					t.Fatalf("webhook logs leak %q", sensitive)
				}
			}
		})
	}
}
