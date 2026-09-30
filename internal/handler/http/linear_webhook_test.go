package http

// Webhook tests at the public HTTP boundary (ADR-0015, #364): 200 means the
// delivery was durably accepted, everything else tells Linear whether to
// retry.

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"go.orx.me/apps/butter/internal/config"
	cryptokeymemory "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	linearrepo "go.orx.me/apps/butter/internal/repo/linear"
	linearmemory "go.orx.me/apps/butter/internal/repo/linear/memory"
	linearruntime "go.orx.me/apps/butter/internal/runtime/linear"
	"go.orx.me/apps/butter/internal/secretbox"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const webhookSecret = "linear-webhook-signing-secret"

// memoryQueue stands in for the Redis Stream: it deduplicates by App and
// delivery ID and records what was accepted.
type memoryQueue struct {
	mu     sync.Mutex
	seen   map[string]bool
	events []*linearruntime.Event
	err    error
}

func (q *memoryQueue) Accept(_ context.Context, event *linearruntime.Event) (string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.err != nil {
		return "", q.err
	}
	key := event.AppID + ":" + event.DeliveryID
	if q.seen[key] {
		return "", linearruntime.ErrDuplicate
	}
	q.seen[key] = true
	q.events = append(q.events, event)
	return "1-0", nil
}

func (q *memoryQueue) accepted() []*linearruntime.Event {
	q.mu.Lock()
	defer q.mu.Unlock()
	return append([]*linearruntime.Event(nil), q.events...)
}

type webhookFixture struct {
	repo   *linearmemory.Store
	queue  *memoryQueue
	router *gin.Engine
	app    *agentsv1.LinearApp
	now    time.Time
}

func newWebhookFixture(t *testing.T, orgs ...string) *webhookFixture {
	t.Helper()
	ctx := t.Context()
	repo := linearmemory.New()
	keyring := secretbox.NewKeyring(cryptokeymemory.New())
	ciphertext, keyID, err := keyring.Encrypt(ctx, []byte(webhookSecret))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	app, err := repo.CreateApp(ctx, "ws-a", &agentsv1.LinearApp{
		Id: "app-1", ClientId: "client-1", AgentId: "support", DisplayName: "Support", InboundEnabled: true,
	}, linearrepo.AppCredentials{WebhookSecret: linearrepo.Credential{Ciphertext: ciphertext, KeyID: keyID}})
	if err != nil {
		t.Fatalf("CreateApp: %v", err)
	}
	if len(orgs) == 0 {
		orgs = []string{"org-1"}
	}
	for _, org := range orgs {
		if _, err := repo.UpsertInstallation(ctx, "ws-a", &agentsv1.LinearInstallation{
			Id: "inst-" + org, AppId: "app-1", OrganizationId: org, OrganizationName: org,
		}, linearrepo.InstallationTokens{}); err != nil {
			t.Fatalf("UpsertInstallation: %v", err)
		}
	}
	fx := &webhookFixture{repo: repo, queue: &memoryQueue{seen: map[string]bool{}}, app: app, now: time.Now()}
	receiver := linearruntime.NewReceiver(repo, keyring, fx.queue)
	receiver.SetClock(func() time.Time { return fx.now })

	gin.SetMode(gin.TestMode)
	fx.router = gin.New()
	fx.router.Use(AuthMiddleware(&config.AppConfig{}, nil, nil, nil))
	NewLinearWebhookHandler(receiver).Register(fx.router)
	return fx
}

func (fx *webhookFixture) payload(mutate func(map[string]any)) []byte {
	body := map[string]any{
		"type":             "AgentSessionEvent",
		"action":           "created",
		"organizationId":   "org-1",
		"webhookTimestamp": fx.now.UnixMilli(),
		"promptContext":    "<issue identifier=\"ENG-1\">Fix login</issue>",
		"agentSession": map[string]any{
			"id":        "session-1",
			"creatorId": "user-1",
			"issue":     map[string]any{"id": "issue-1", "identifier": "ENG-1", "title": "Fix login", "url": "https://linear.app/acme/issue/ENG-1"},
		},
	}
	if mutate != nil {
		mutate(body)
	}
	raw, _ := json.Marshal(body)
	return raw
}

func sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func (fx *webhookFixture) post(t *testing.T, appID string, body []byte, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/linear/webhook/"+appID, bytes.NewReader(body))
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	fx.router.ServeHTTP(w, req)
	return w
}

func signed(body []byte, delivery string) map[string]string {
	h := map[string]string{"Linear-Signature": sign(webhookSecret, body)}
	if delivery != "" {
		h["Linear-Delivery"] = delivery
	}
	return h
}

func TestLinearWebhookStatuses(t *testing.T) {
	for _, tc := range []struct {
		name     string
		appID    string
		body     func(fx *webhookFixture) []byte
		headers  func(body []byte) map[string]string
		setup    func(fx *webhookFixture)
		status   int
		accepted int
	}{
		{
			name: "accepted", status: http.StatusOK, accepted: 1,
		},
		{
			name: "unknown app", appID: "missing", status: http.StatusNotFound,
		},
		{
			name:   "body over 1 MiB",
			body:   func(*webhookFixture) []byte { return bytes.Repeat([]byte("x"), 1<<20+1) },
			status: http.StatusRequestEntityTooLarge,
		},
		{
			name:    "missing signature",
			headers: func([]byte) map[string]string { return map[string]string{"Linear-Delivery": "d-1"} },
			status:  http.StatusUnauthorized,
		},
		{
			name: "wrong signature",
			headers: func(body []byte) map[string]string {
				return map[string]string{"Linear-Signature": sign("another-secret", body), "Linear-Delivery": "d-1"}
			},
			status: http.StatusUnauthorized,
		},
		{
			name:   "malformed JSON",
			body:   func(*webhookFixture) []byte { return []byte("{not json") },
			status: http.StatusBadRequest,
		},
		{
			name: "stale timestamp",
			body: func(fx *webhookFixture) []byte {
				return fx.payload(func(p map[string]any) { p["webhookTimestamp"] = fx.now.Add(-2 * time.Minute).UnixMilli() })
			},
			status: http.StatusUnauthorized,
		},
		{
			name: "not an agent session event",
			body: func(fx *webhookFixture) []byte {
				return fx.payload(func(p map[string]any) { p["type"] = "Issue" })
			},
			status: http.StatusOK,
		},
		{
			name: "app not receiving",
			setup: func(fx *webhookFixture) {
				app := fx.app
				app.InboundEnabled = false
				if _, err := fx.repo.UpdateApp(context.Background(), "ws-a", app, app.GetRevision()); err != nil {
					panic(err)
				}
			},
			status: http.StatusOK,
		},
		{
			name: "organization the app is not installed in",
			body: func(fx *webhookFixture) []byte {
				return fx.payload(func(p map[string]any) { p["organizationId"] = "org-other" })
			},
			status: http.StatusOK,
		},
		{
			name:   "queue failure",
			setup:  func(fx *webhookFixture) { fx.queue.err = errors.New("redis down") },
			status: http.StatusServiceUnavailable,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fx := newWebhookFixture(t)
			if tc.setup != nil {
				tc.setup(fx)
			}
			appID := "app-1"
			if tc.appID != "" {
				appID = tc.appID
			}
			body := fx.payload(nil)
			if tc.body != nil {
				body = tc.body(fx)
			}
			headers := signed(body, "d-1")
			if tc.headers != nil {
				headers = tc.headers(body)
			}
			w := fx.post(t, appID, body, headers)
			if w.Code != tc.status {
				t.Fatalf("status = %d, want %d; body %s", w.Code, tc.status, w.Body.String())
			}
			if got := len(fx.queue.accepted()); got != tc.accepted {
				t.Fatalf("accepted = %d, want %d", got, tc.accepted)
			}
		})
	}
}

func TestAcceptedDeliveryIsAFrozenSnapshot(t *testing.T) {
	fx := newWebhookFixture(t)
	body := fx.payload(func(p map[string]any) {
		p["action"] = "prompted"
		p["agentActivity"] = map[string]any{"id": "act-1", "userId": "user-2", "content": map[string]any{"body": "  please add a test  "}}
	})
	if w := fx.post(t, "app-1", body, signed(body, "d-1")); w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	events := fx.queue.accepted()
	if len(events) != 1 {
		t.Fatalf("accepted = %d", len(events))
	}
	ev := events[0]
	want := linearruntime.Event{
		WorkspaceID: "ws-a", AppID: "app-1", AppRevision: fx.app.GetRevision(),
		InstallationID: "inst-org-1", OrganizationID: "org-1",
		AgentSessionID: "session-1", Action: "prompted", PromptText: "please add a test",
		PromptingUserID: "user-2", PromptContext: "<issue identifier=\"ENG-1\">Fix login</issue>",
		Issue:      linearruntime.Issue{ID: "issue-1", Identifier: "ENG-1", Title: "Fix login", URL: "https://linear.app/acme/issue/ENG-1"},
		DeliveryID: "d-1", ReceivedAtUnixMs: fx.now.UnixMilli(),
	}
	if *ev != want {
		t.Fatalf("event = %+v\nwant    %+v", *ev, want)
	}
}

func TestStopSignalIsCarried(t *testing.T) {
	fx := newWebhookFixture(t)
	body := fx.payload(func(p map[string]any) {
		p["action"] = "prompted"
		p["agentActivity"] = map[string]any{"signal": "stop", "userId": "user-1", "content": map[string]any{"body": ""}}
	})
	fx.post(t, "app-1", body, signed(body, "d-1"))
	if events := fx.queue.accepted(); len(events) != 1 || !events[0].Stop {
		t.Fatalf("events = %+v; want one stop event", events)
	}
}

func TestRedeliveriesAreDeduplicated(t *testing.T) {
	fx := newWebhookFixture(t)
	body := fx.payload(nil)
	for i := 0; i < 2; i++ {
		if w := fx.post(t, "app-1", body, signed(body, "d-1")); w.Code != http.StatusOK {
			t.Fatalf("delivery %d status = %d", i, w.Code)
		}
	}
	// Without Linear-Delivery the body hash is the key: the same body
	// redelivered is a duplicate, a different one is not.
	for i := 0; i < 2; i++ {
		fx.post(t, "app-1", body, signed(body, ""))
	}
	other := fx.payload(func(p map[string]any) { p["action"] = "prompted" })
	fx.post(t, "app-1", other, signed(other, ""))
	if got := len(fx.queue.accepted()); got != 3 {
		t.Fatalf("accepted = %d, want 3 (one per distinct delivery)", got)
	}
}

func TestAnEventWithoutAnOrganizationUsesTheOnlyInstallation(t *testing.T) {
	noOrg := func(fx *webhookFixture) []byte {
		return fx.payload(func(p map[string]any) { delete(p, "organizationId") })
	}
	fx := newWebhookFixture(t)
	body := noOrg(fx)
	fx.post(t, "app-1", body, signed(body, "d-1"))
	if events := fx.queue.accepted(); len(events) != 1 || events[0].InstallationID != "inst-org-1" {
		t.Fatalf("events = %+v; want the only installation", events)
	}

	two := newWebhookFixture(t, "org-1", "org-2")
	body = noOrg(two)
	if w := two.post(t, "app-1", body, signed(body, "d-1")); w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (ignored)", w.Code)
	}
	if got := len(two.queue.accepted()); got != 0 {
		t.Fatalf("accepted = %d with two installations, want 0", got)
	}
}

func TestLinearWebhookIsPublic(t *testing.T) {
	if !isPublicPath("/api/linear/webhook/app-1") {
		t.Fatal("the Linear webhook must bypass authentication: Linear signs it instead")
	}
	if strings.HasPrefix("/api/linear/webhooks", "/api/linear/webhook/") {
		t.Fatal("prefix check is too loose")
	}
}
