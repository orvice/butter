package notify

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type Message struct {
	Title string
	Text  string
}

const DefaultHTTPTimeout = 10 * time.Second

// TelegramDelivery delivers a message to one Telegram Destination.
//
// notify holds this as an interface rather than reaching for the Telegram
// repository directly: after issue #264 a notify target carries only a
// Destination ID, and resolving it — credential, chat, Forum Topic, Markdown
// handling, retry_after — is the unified sender's job, not this package's.
type TelegramDelivery interface {
	SendToDestination(ctx context.Context, workspaceID, destinationID, text string) error
}

type Sender struct {
	client   *http.Client
	telegram TelegramDelivery
}

func NewSender(client *http.Client) *Sender {
	if client == nil {
		client = &http.Client{Timeout: DefaultHTTPTimeout}
	}
	return &Sender{client: client}
}

// SetTelegramDelivery wires the unified Telegram sender. Without it, Telegram
// targets fail explicitly rather than being skipped: an alert that silently
// goes nowhere is worse than one that reports a failure.
func (s *Sender) SetTelegramDelivery(delivery TelegramDelivery) {
	s.telegram = delivery
}

func (s *Sender) Send(ctx context.Context, workspaceID string, target *agentsv1.NotifyTarget, msg Message) error {
	if target == nil {
		return fmt.Errorf("notify target is nil")
	}
	if !target.GetEnabled() {
		return nil
	}
	switch target.GetType() {
	case agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_TELEGRAM:
		return s.sendTelegram(ctx, workspaceID, target.GetTelegram(), msg)
	case agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_LARK_WEBHOOK:
		return s.sendLark(ctx, target.GetLark(), msg)
	case agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_DISCORD_WEBHOOK:
		return s.sendDiscord(ctx, target.GetDiscord(), msg)
	default:
		return fmt.Errorf("unsupported notify target type %s", target.GetType().String())
	}
}

func (s *Sender) sendTelegram(ctx context.Context, workspaceID string, target *agentsv1.TelegramNotifyTarget, msg Message) error {
	destinationID := strings.TrimSpace(target.GetDestinationId())
	if destinationID == "" {
		return fmt.Errorf("telegram target requires destination_id")
	}
	if s.telegram == nil {
		return fmt.Errorf("telegram destination delivery is not configured")
	}
	if workspaceID == "" {
		return fmt.Errorf("telegram delivery requires a workspace")
	}
	return s.telegram.SendToDestination(ctx, workspaceID, destinationID, formatMessage(msg))
}

func (s *Sender) sendLark(ctx context.Context, target *agentsv1.LarkNotifyTarget, msg Message) error {
	if target.GetWebhookUrl() == "" {
		return fmt.Errorf("lark target requires webhook_url")
	}
	payload := map[string]any{
		"msg_type": "text",
		"content":  map[string]string{"text": formatMessage(msg)},
	}
	if target.GetSecret() != "" {
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		payload["timestamp"] = timestamp
		payload["sign"] = larkSign(timestamp, target.GetSecret())
	}
	return s.postJSON(ctx, target.GetWebhookUrl(), payload)
}

func (s *Sender) sendDiscord(ctx context.Context, target *agentsv1.DiscordNotifyTarget, msg Message) error {
	if target.GetWebhookUrl() == "" {
		return fmt.Errorf("discord target requires webhook_url")
	}
	endpoint := target.GetWebhookUrl()
	if target.GetThreadId() != "" {
		u, err := url.Parse(endpoint)
		if err != nil {
			return fmt.Errorf("parse discord webhook_url: %w", err)
		}
		q := u.Query()
		q.Set("thread_id", target.GetThreadId())
		u.RawQuery = q.Encode()
		endpoint = u.String()
	}
	payload := map[string]any{"content": formatMessage(msg)}
	if target.GetUsername() != "" {
		payload["username"] = target.GetUsername()
	}
	if target.GetAvatarUrl() != "" {
		payload["avatar_url"] = target.GetAvatarUrl()
	}
	return s.postJSON(ctx, endpoint, payload)
}

func (s *Sender) postJSON(ctx context.Context, endpoint string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal notify payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create notify request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send notify request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		const maxBodyRead = 512
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, maxBodyRead))
		if len(respBody) > 0 {
			return fmt.Errorf("notify request returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
		}
		return fmt.Errorf("notify request returned status %d", resp.StatusCode)
	}
	return nil
}

func formatMessage(msg Message) string {
	if strings.TrimSpace(msg.Title) == "" {
		return msg.Text
	}
	if strings.TrimSpace(msg.Text) == "" {
		return msg.Title
	}
	return msg.Title + "\n" + msg.Text
}

func larkSign(timestamp, secret string) string {
	mac := hmac.New(sha256.New, []byte(timestamp+"\n"+secret))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
