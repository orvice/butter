package notify

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type blockingTransport struct{}

func (blockingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	<-req.Context().Done()
	return nil, req.Context().Err()
}

type captureTransport struct {
	reqBodies [][]byte
	reqs      []*http.Request
}

func (t *captureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	t.reqs = append(t.reqs, req)
	t.reqBodies = append(t.reqBodies, body)
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       http.NoBody,
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func TestNewSenderUsesDefaultTimeout(t *testing.T) {
	sender := NewSender(nil)
	if sender.client.Timeout != DefaultHTTPTimeout {
		t.Fatalf("expected default timeout %s, got %s", DefaultHTTPTimeout, sender.client.Timeout)
	}
}

func TestSendLarkPayload(t *testing.T) {
	var payload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	err := NewSender(server.Client()).Send(context.Background(), "ws-test", &agentsv1.NotifyTarget{
		Enabled: true,
		Type:    agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_LARK_WEBHOOK,
		Lark:    &agentsv1.LarkNotifyTarget{WebhookUrl: server.URL},
	}, Message{Title: "title", Text: "body"})
	if err != nil {
		t.Fatalf("send lark: %v", err)
	}
	if payload["msg_type"] != "text" {
		t.Fatalf("unexpected msg_type %#v", payload["msg_type"])
	}
	content, ok := payload["content"].(map[string]any)
	if !ok {
		t.Fatalf("unexpected content %#v", payload["content"])
	}
	if content["text"] != "title\nbody" {
		t.Fatalf("unexpected text %#v", content["text"])
	}
}

func TestSendDiscordPayload(t *testing.T) {
	var threadID string
	var payload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		threadID = r.URL.Query().Get("thread_id")
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	err := NewSender(server.Client()).Send(context.Background(), "ws-test", &agentsv1.NotifyTarget{
		Enabled: true,
		Type:    agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_DISCORD_WEBHOOK,
		Discord: &agentsv1.DiscordNotifyTarget{
			WebhookUrl: server.URL + "?wait=true",
			Username:   "Butter",
			AvatarUrl:  "https://example.com/avatar.png",
			ThreadId:   "thread-1",
		},
	}, Message{Text: "body"})
	if err != nil {
		t.Fatalf("send discord: %v", err)
	}
	if threadID != "thread-1" {
		t.Fatalf("unexpected thread_id %q", threadID)
	}
	if payload["content"] != "body" {
		t.Fatalf("unexpected content %#v", payload["content"])
	}
	if payload["username"] != "Butter" {
		t.Fatalf("unexpected username %#v", payload["username"])
	}
	if payload["avatar_url"] != "https://example.com/avatar.png" {
		t.Fatalf("unexpected avatar_url %#v", payload["avatar_url"])
	}
}

func TestPostJSONIncludesResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"ok":false,"description":"Bad Request: can't parse entities"}`))
	}))
	defer server.Close()

	sender := NewSender(server.Client())
	err := sender.Send(context.Background(), "ws-test", &agentsv1.NotifyTarget{
		Enabled: true,
		Type:    agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_LARK_WEBHOOK,
		Lark:    &agentsv1.LarkNotifyTarget{WebhookUrl: server.URL},
	}, Message{Text: "body"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Fatalf("expected status 400 in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "can't parse entities") {
		t.Fatalf("expected response body in error, got: %v", err)
	}
}

func TestSendHonorsContextTimeout(t *testing.T) {
	sender := NewSender(&http.Client{Transport: blockingTransport{}})
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := sender.Send(ctx, "ws-test", &agentsv1.NotifyTarget{
		Enabled: true,
		Type:    agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_LARK_WEBHOOK,
		Lark:    &agentsv1.LarkNotifyTarget{WebhookUrl: "https://example.invalid/hook"},
	}, Message{Text: "body"})
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}

// --- Telegram delegation (issue #264) --------------------------------------

// fakeTelegramDelivery records what the notifier handed to the unified sender.
type fakeTelegramDelivery struct {
	calls []struct{ workspaceID, destinationID, text string }
	err   error
}

func (f *fakeTelegramDelivery) SendToDestination(_ context.Context, workspaceID, destinationID, text string) error {
	f.calls = append(f.calls, struct{ workspaceID, destinationID, text string }{workspaceID, destinationID, text})
	return f.err
}

// A Telegram target now carries only a Destination ID; the notifier resolves
// nothing itself and issues no HTTP request of its own.
func TestSendTelegramDelegatesToTheUnifiedSender(t *testing.T) {
	transport := &captureTransport{}
	delivery := &fakeTelegramDelivery{}
	sender := NewSender(&http.Client{Transport: transport})
	sender.SetTelegramDelivery(delivery)

	err := sender.Send(context.Background(), "ws-test", &agentsv1.NotifyTarget{
		Enabled:  true,
		Type:     agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_TELEGRAM,
		Telegram: &agentsv1.TelegramNotifyTarget{DestinationId: "dest-1"},
	}, Message{Title: "Cron job job1: success", Text: "done"})
	if err != nil {
		t.Fatalf("send telegram: %v", err)
	}
	if len(transport.reqs) != 0 {
		t.Fatalf("notifier issued %d telegram HTTP requests itself", len(transport.reqs))
	}
	if len(delivery.calls) != 1 {
		t.Fatalf("expected one delivery, got %d", len(delivery.calls))
	}
	call := delivery.calls[0]
	if call.workspaceID != "ws-test" || call.destinationID != "dest-1" {
		t.Errorf("delivered to workspace %q destination %q", call.workspaceID, call.destinationID)
	}
	if call.text != "Cron job job1: success\ndone" {
		t.Errorf("text = %q", call.text)
	}
}

func TestSendTelegramRequiresADestination(t *testing.T) {
	sender := NewSender(&http.Client{Transport: &captureTransport{}})
	sender.SetTelegramDelivery(&fakeTelegramDelivery{})

	err := sender.Send(context.Background(), "ws-test", &agentsv1.NotifyTarget{
		Enabled:  true,
		Type:     agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_TELEGRAM,
		Telegram: &agentsv1.TelegramNotifyTarget{},
	}, Message{Text: "hi"})
	if err == nil {
		t.Fatal("expected a target without destination_id to fail")
	}
}

// An unconfigured delivery fails loudly. Skipping would turn a misconfigured
// deployment into alerts that silently go nowhere.
func TestSendTelegramFailsWhenDeliveryIsUnconfigured(t *testing.T) {
	sender := NewSender(&http.Client{Transport: &captureTransport{}})

	err := sender.Send(context.Background(), "ws-test", &agentsv1.NotifyTarget{
		Enabled:  true,
		Type:     agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_TELEGRAM,
		Telegram: &agentsv1.TelegramNotifyTarget{DestinationId: "dest-1"},
	}, Message{Text: "hi"})
	if err == nil {
		t.Fatal("expected an unconfigured telegram delivery to fail")
	}
}

func TestSendTelegramSurfacesDeliveryFailure(t *testing.T) {
	delivery := &fakeTelegramDelivery{err: errors.New("destination is disabled")}
	sender := NewSender(&http.Client{Transport: &captureTransport{}})
	sender.SetTelegramDelivery(delivery)

	err := sender.Send(context.Background(), "ws-test", &agentsv1.NotifyTarget{
		Enabled:  true,
		Type:     agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_TELEGRAM,
		Telegram: &agentsv1.TelegramNotifyTarget{DestinationId: "dest-1"},
	}, Message{Text: "hi"})
	if err == nil || !strings.Contains(err.Error(), "destination is disabled") {
		t.Fatalf("err = %v, want the delivery failure surfaced", err)
	}
}
