package notify

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
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
	reqBody []byte
	req     *http.Request
}

func (t *captureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	t.req = req
	t.reqBody = body
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

func TestSendTelegramPayload(t *testing.T) {
	var payload map[string]any
	transport := &captureTransport{}
	sender := NewSender(&http.Client{Transport: transport})
	err := sender.Send(context.Background(), &agentsv1.NotifyTarget{
		Enabled: true,
		Type:    agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_TELEGRAM,
		Telegram: &agentsv1.TelegramNotifyTarget{
			BotToken:        "secret-token",
			ChatId:          "chat-1",
			ParseMode:       "Markdown",
			MessageThreadId: 7,
		},
	}, Message{Title: "Cron job job1: success", Text: "done"})
	if err != nil {
		t.Fatalf("send telegram: %v", err)
	}
	if transport.req.Method != http.MethodPost {
		t.Fatalf("expected POST, got %s", transport.req.Method)
	}
	if transport.req.URL.String() != "https://api.telegram.org/botsecret-token/sendMessage" {
		t.Fatalf("unexpected telegram URL %s", transport.req.URL.String())
	}
	if got := transport.req.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected JSON content type, got %q", got)
	}
	if err := json.Unmarshal(transport.reqBody, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload["chat_id"] != "chat-1" {
		t.Fatalf("unexpected chat_id %#v", payload["chat_id"])
	}
	if payload["text"] != "Cron job job1: success\ndone" {
		t.Fatalf("unexpected text %#v", payload["text"])
	}
	if payload["parse_mode"] != "Markdown" {
		t.Fatalf("unexpected parse_mode %#v", payload["parse_mode"])
	}
	if payload["message_thread_id"] != float64(7) {
		t.Fatalf("unexpected message_thread_id %#v", payload["message_thread_id"])
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

	err := NewSender(server.Client()).Send(context.Background(), &agentsv1.NotifyTarget{
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

	err := NewSender(server.Client()).Send(context.Background(), &agentsv1.NotifyTarget{
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

func TestSendHonorsContextTimeout(t *testing.T) {
	sender := NewSender(&http.Client{Transport: blockingTransport{}})
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := sender.Send(ctx, &agentsv1.NotifyTarget{
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
