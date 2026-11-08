package plane

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	cryptokeymem "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	planerepo "go.orx.me/apps/butter/internal/repo/plane"
	planemem "go.orx.me/apps/butter/internal/repo/plane/memory"
	installationrepo "go.orx.me/apps/butter/internal/repo/planeinstallation"
	installationmem "go.orx.me/apps/butter/internal/repo/planeinstallation/memory"
	"go.orx.me/apps/butter/internal/secretbox"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestWebhookPayloadAcceptsVerifiedDiscriminators(t *testing.T) {
	var p webhookPayload
	if err := p.UnmarshalJSON([]byte(`{"type":"agent_run_activity","action":"prompted","app_client_id":"client","app_installation_id":"install","agent_user_id":"bot","workspace_id":"ws","run":{"id":"run-1","workspace_id":"ws"},"activity":{"id":"activity-1"}}`)); err != nil {
		t.Fatal(err)
	}
	if p.Run.ID != "run-1" || p.Activity.ID != "activity-1" {
		t.Fatalf("payload identities = %+v", p)
	}
}

func TestWebhookPayloadAcceptsPlaneEnvelope(t *testing.T) {
	var p webhookPayload
	body := `{"event":"agent_run_user_prompt","data":{"action":"prompted","type":"agent_run_activity","agent_user_id":"bot","app_client_id":"client","app_installation_id":"install","workspace_id":"ws","agent_run":{"id":"run-1","workspace_id":"ws"},"agent_run_activity":{"id":"activity-1","content":{"body":"<p>Hello Plane</p>"}}}}`
	if err := p.UnmarshalJSON([]byte(body)); err != nil {
		t.Fatal(err)
	}
	if p.Type != "agent_run_activity" || p.Action != "prompted" || p.Run.ID != "run-1" || p.Activity.ID != "activity-1" || p.Prompt != "<p>Hello Plane</p>" {
		t.Fatalf("decoded envelope = %+v", p)
	}
}

func TestWebhookPayloadPreservesStopSignalAndActor(t *testing.T) {
	var p webhookPayload
	body := `{"event":"agent_run_user_prompt","data":{"action":"prompted","type":"agent_run_activity","agent_user_id":"bot","app_client_id":"client","app_installation_id":"install","workspace_id":"ws","agent_run":{"id":"run-1","workspace_id":"ws"},"agent_run_activity":{"id":"activity-stop","signal":"stop","user_id":"user-1","content":{"body":"ignored"}}}}`
	if err := p.UnmarshalJSON([]byte(body)); err != nil {
		t.Fatal(err)
	}
	if p.Activity.Signal != "stop" || p.Activity.UserID != "user-1" {
		t.Fatalf("activity = %+v, want stop from user-1", p.Activity)
	}
	if !allowedStopUser([]string{"user-1"}, p.Activity.UserID) {
		t.Fatal("authorized stop user rejected")
	}
	if allowedStopUser([]string{"user-2"}, p.Activity.UserID) {
		t.Fatal("unauthorized stop user accepted")
	}
}
func TestValidSignatureUsesRawBodyAndConstantTimeDigest(t *testing.T) {
	body := []byte(`{"type":"agent_run","action":"created"}`)
	secret := []byte("secret")
	mac := hmac.New(sha256.New, secret)
	mac.Write(body)
	sig := hex.EncodeToString(mac.Sum(nil))
	if !validSignature(secret, sig, body) {
		t.Fatal("valid signature rejected")
	}
	if validSignature(secret, sig, []byte(`{"action":"created","type":"agent_run"}`)) {
		t.Fatal("re-serialized body accepted")
	}
	if validSignature(secret, "bogus", body) {
		t.Fatal("invalid signature accepted")
	}
}
func TestHeaderNames(t *testing.T) {
	h := http.Header{}
	h.Set(SignatureHeader, "x")
	if h.Get(SignatureHeader) != "x" {
		t.Fatal("signature header missing")
	}
}

func TestWebhookActorConflictsAreRejectedWithoutCreatorFallback(t *testing.T) {
	for _, body := range []string{
		`{"activity":{"id":"p","user_id":"alice"},"data":{"agent_run_activity":{"id":"p","user_id":"bob"}}}`,
		`{"agent_activity":{"id":"p","user_id":"alice"},"activity":{"id":"p","user_id":"bob"}}`,
		`{"activity":{"id":"p","signal":"stop"},"data":{"activity":{"id":"p","signal":"complete"}}}`,
	} {
		var payload webhookPayload
		if err := payload.UnmarshalJSON([]byte(body)); err == nil {
			t.Fatalf("conflicting activity identity accepted: %s", body)
		}
	}
	var payload webhookPayload
	if err := payload.UnmarshalJSON([]byte(`{"run":{"id":"run","user_id":"creator"},"activity":{"id":"p"}}`)); err != nil {
		t.Fatal(err)
	}
	if payload.Activity.UserID != "" || allowedStopUser(nil, payload.Activity.UserID) {
		t.Fatal("Run creator was used as the current actor")
	}
}

type admissionQueue struct{ events []*Event }

func (q *admissionQueue) Accept(_ context.Context, e *Event) (string, error) {
	copy := *e
	copy.AcceptedOrder = int64(len(q.events) + 1)
	e.AcceptedOrder = copy.AcceptedOrder
	q.events = append(q.events, &copy)
	return e.ActivityID, nil
}

func TestAdmissionFreezesActorAndAgentAndIgnoresRepeatedCreate(t *testing.T) {
	apps, installations := planemem.New(), installationmem.New()
	ring := secretbox.NewKeyring(cryptokeymem.New())
	ciphertext, keyID, err := ring.Encrypt(t.Context(), []byte("webhook-secret"))
	if err != nil {
		t.Fatal(err)
	}
	app, err := apps.CreateApp(t.Context(), "ws", &agentsv1.PlaneApp{
		Id: "app", ClientId: "client", AgentId: "agent-at-admission", InboundEnabled: true,
	}, planerepo.AppCredentials{WebhookSecret: planerepo.Credential{Ciphertext: ciphertext, KeyID: keyID}})
	if err != nil {
		t.Fatal(err)
	}
	inst, err := installations.Upsert(t.Context(), "ws", "app", &agentsv1.PlaneInstallation{
		AppId: "app", AppInstallationId: "external-installation", PlaneWorkspaceId: "plane-ws",
		AppBotUserId: "bot", GrantedScopes: requiredAgentScopes,
	}, installationrepo.Tokens{})
	if err != nil {
		t.Fatal(err)
	}
	q := &admissionQueue{}
	receiver := NewReceiver(apps, installations, ring, q)
	deliver := func(kind, action, actor, activity string) (Decision, error) {
		body, _ := json.Marshal(map[string]any{
			"type": kind, "action": action, "app_client_id": "client", "app_installation_id": "external-installation",
			"agent_user_id": "bot", "workspace_id": "plane-ws", "prompt": "exact input\n🙂",
			"run":      map[string]string{"id": "run", "workspace_id": "plane-ws", "user_id": "creator"},
			"activity": map[string]string{"id": activity, "user_id": actor},
		})
		mac := hmac.New(sha256.New, []byte("webhook-secret"))
		_, _ = mac.Write(body)
		header := http.Header{}
		header.Set(SignatureHeader, hex.EncodeToString(mac.Sum(nil)))
		return receiver.Deliver(t.Context(), app, header, body)
	}
	if decision, err := deliver("agent_run_activity", "prompted", "alice", "prompt"); err != nil || decision != DecisionAccepted {
		t.Fatalf("admission=%s err=%v", decision, err)
	}
	for range 2 {
		if decision, err := deliver("agent_run", "created", "", ""); err != nil || decision != DecisionIgnored {
			t.Fatalf("Create invoked or reset the Run: %s %v", decision, err)
		}
	}
	for _, actor := range []string{"", "bot"} {
		if _, err := deliver("agent_run_activity", "prompted", actor, "invalid"); !errors.Is(err, ErrIdentity) {
			t.Fatalf("missing or bot actor admitted: actor=%q err=%v", actor, err)
		}
	}
	if len(q.events) != 1 {
		t.Fatalf("Create or untrusted actor manufactured %d inputs", len(q.events))
	}
	event := q.events[0]
	if event.AgentID != "agent-at-admission" || event.ActorUserID != "alice" || event.InstallationID != inst.GetId() ||
		event.PlaneInstallationID != "external-installation" || event.Prompt != "exact input\n🙂" || event.AcceptedOrder != 1 {
		t.Fatalf("admission snapshot=%+v", event)
	}
}
