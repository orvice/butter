package plane

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"testing"
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
