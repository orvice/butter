package http

// Protocol regression at the public webhook boundary (KON-77). The Stage 1
// evidence package is a sanitized shape fixture, not live evidence: these
// tests replay its request shapes so the documented field paths, dedupe and
// organization routing keep matching the accepted implementation.

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type linearProtocolFixture struct {
	FixtureVersion string `json:"fixture_version"`
	CaptureStatus  string `json:"capture_status"`
	Cases          []struct {
		Name    string `json:"name"`
		Request struct {
			Headers map[string]string `json:"headers"`
			Body    map[string]any    `json:"body"`
		} `json:"request"`
		ExpectedButter map[string]any `json:"expected_butter"`
	} `json:"cases"`
}

func loadLinearProtocolFixture(t *testing.T) linearProtocolFixture {
	t.Helper()
	path := filepath.Join("..", "..", "..",
		"docs", "research", "fixtures", "linear-agent-session", "agent-session-events.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read protocol fixture: %v", err)
	}
	var fixture linearProtocolFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode protocol fixture: %v", err)
	}
	if fixture.CaptureStatus != "synthetic_shape_only" {
		t.Fatalf("fixture capture status = %q; only sanitized shapes may drive this test", fixture.CaptureStatus)
	}
	return fixture
}

// TestStageOneFixturesAreAcceptedAtTheWebhook keeps the documented request
// shapes reachable through the public route. Each case is accepted once and a
// redelivery of the same signed body is a duplicate that runs nothing new.
func TestStageOneFixturesAreAcceptedAtTheWebhook(t *testing.T) {
	fixture := loadLinearProtocolFixture(t)

	want := map[string]struct {
		action string
		stop   bool
		prompt bool
		user   string
	}{
		"delegation_created": {action: "created", user: "user-redacted"},
		"mention_created":    {action: "created", prompt: true, user: "user-redacted"},
		"followup_prompted":  {action: "prompted", prompt: true, user: "user-redacted"},
		"stop_prompted":      {action: "prompted", stop: true, user: "user-redacted"},
	}

	seen := map[string]bool{}
	for _, tc := range fixture.Cases {
		if len(tc.Request.Body) == 0 {
			// Shape-only cases (success/failure/redelivery activities) do not
			// carry a request body.
			continue
		}
		t.Run(tc.Name, func(t *testing.T) {
			expect, ok := want[tc.Name]
			if !ok {
				t.Fatalf("fixture case %q has no expectation; update this test with the new shape", tc.Name)
			}
			seen[tc.Name] = true

			fx := newWebhookFixture(t, "org-redacted")
			// The fixture pins a capture timestamp; the replay window is
			// measured against the receiver clock, so move it to now.
			body := tc.Request.Body
			body["webhookTimestamp"] = fx.now.UnixMilli()
			raw, err := json.Marshal(body)
			if err != nil {
				t.Fatalf("encode fixture body: %v", err)
			}
			delivery := tc.Request.Headers["Linear-Delivery"]
			if delivery == "" {
				t.Fatalf("fixture case %q has no Linear-Delivery header", tc.Name)
			}

			if w := fx.post(t, "app-1", raw, signed(raw, delivery)); w.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200; body %s", w.Code, w.Body.String())
			}
			events := fx.queue.accepted()
			if len(events) != 1 {
				t.Fatalf("accepted = %d, want 1", len(events))
			}
			event := events[0]
			if event.OrganizationID != "org-redacted" || event.InstallationID != "inst-org-redacted" {
				t.Fatalf("organization routing = %q / %q, want the installed organization", event.OrganizationID, event.InstallationID)
			}
			if event.Action != expect.action {
				t.Fatalf("action = %q, want %q", event.Action, expect.action)
			}
			if event.Stop != expect.stop {
				t.Fatalf("stop = %v, want %v", event.Stop, expect.stop)
			}
			if expect.prompt && event.PromptText == "" {
				t.Fatal("prompt text is empty, want the fixture's message")
			}
			if event.PromptingUserID != expect.user {
				t.Fatalf("prompting user = %q, want %q", event.PromptingUserID, expect.user)
			}
			if event.DeliveryID != delivery {
				t.Fatalf("delivery = %q, want %q", event.DeliveryID, delivery)
			}

			// The same signed body and delivery ID is a duplicate: 200, no
			// second event.
			if w := fx.post(t, "app-1", raw, signed(raw, delivery)); w.Code != http.StatusOK {
				t.Fatalf("redelivery status = %d, want 200", w.Code)
			}
			if got := len(fx.queue.accepted()); got != 1 {
				t.Fatalf("accepted = %d after redelivery, want 1", got)
			}
		})
	}

	for name := range want {
		if !seen[name] {
			t.Fatalf("fixture no longer contains case %q", name)
		}
	}
}

// TestARejectedDeliveryDoesNotBurnTheDedupeKey keeps a bad signature or a
// stale timestamp from consuming the delivery ID: the genuine retry of that
// same delivery must still be accepted.
func TestARejectedDeliveryDoesNotBurnTheDedupeKey(t *testing.T) {
	for _, tc := range []struct {
		name    string
		headers func(*webhookFixture, []byte) map[string]string
	}{
		{
			name: "wrong signature",
			headers: func(_ *webhookFixture, body []byte) map[string]string {
				return map[string]string{"Linear-Signature": sign("another-secret", body), "Linear-Delivery": "d-1"}
			},
		},
		{
			name: "stale timestamp",
			headers: func(fx *webhookFixture, body []byte) map[string]string {
				return signed(body, "d-1")
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fx := newWebhookFixture(t)
			body := fx.payload(nil)
			if tc.name == "stale timestamp" {
				body = fx.payload(func(p map[string]any) {
					p["webhookTimestamp"] = fx.now.Add(-2 * time.Minute).UnixMilli()
				})
			}
			if w := fx.post(t, "app-1", body, tc.headers(fx, body)); w.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401", w.Code)
			}
			if got := len(fx.queue.accepted()); got != 0 {
				t.Fatalf("accepted = %d after a rejected delivery, want 0", got)
			}

			// The genuine delivery reuses the same ID and must be accepted.
			live := fx.payload(nil)
			if w := fx.post(t, "app-1", live, signed(live, "d-1")); w.Code != http.StatusOK {
				t.Fatalf("retry status = %d, want 200", w.Code)
			}
			if got := len(fx.queue.accepted()); got != 1 {
				t.Fatalf("accepted = %d after retry, want 1", got)
			}
		})
	}
}

// TestTheDeliveryHeaderDeduplicatesDifferentBodies covers the documented
// dedupe key: Linear-Delivery scopes dedupe, so a body rewritten by a proxy
// or a retry with changed bytes is still one delivery.
func TestTheDeliveryHeaderDeduplicatesDifferentBodies(t *testing.T) {
	fx := newWebhookFixture(t)
	first := fx.payload(nil)
	if w := fx.post(t, "app-1", first, signed(first, "d-shared")); w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	second := fx.payload(func(p map[string]any) { p["action"] = "prompted" })
	if w := fx.post(t, "app-1", second, signed(second, "d-shared")); w.Code != http.StatusOK {
		t.Fatalf("second status = %d, want 200", w.Code)
	}
	if got := len(fx.queue.accepted()); got != 1 {
		t.Fatalf("accepted = %d, want 1 for one delivery ID", got)
	}
}
