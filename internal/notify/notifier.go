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
	"regexp"
	"strconv"
	"strings"
	"time"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// telegramBotTokenPattern matches the canonical Telegram Bot API token
// format `<bot_id>:<secret>`. We validate before interpolating the token
// into the API URL so a malformed value cannot inject path segments,
// query strings, or CRLF.
var telegramBotTokenPattern = regexp.MustCompile(`^[0-9]+:[A-Za-z0-9_-]+$`)

// telegramMaxTextBytes is the Telegram Bot API hard limit for the `text` field.
const telegramMaxTextBytes = 4096

// truncateForTelegram shortens s to at most telegramMaxTextBytes bytes,
// appending an ellipsis when truncation occurs.
func truncateForTelegram(s string) string {
	const ellipsis = "…[truncated]"
	if len(s) <= telegramMaxTextBytes {
		return s
	}
	// Cut at a rune boundary.
	cut := []rune(s)
	maxRunes := telegramMaxTextBytes - len(ellipsis)
	if maxRunes < 0 {
		maxRunes = 0
	}
	if len(cut) > maxRunes {
		cut = cut[:maxRunes]
	}
	return string(cut) + ellipsis
}

type Message struct {
	Title string
	Text  string
}

const DefaultHTTPTimeout = 10 * time.Second

type Sender struct {
	client *http.Client
}

func NewSender(client *http.Client) *Sender {
	if client == nil {
		client = &http.Client{Timeout: DefaultHTTPTimeout}
	}
	return &Sender{client: client}
}

func (s *Sender) Send(ctx context.Context, target *agentsv1.NotifyTarget, msg Message) error {
	if target == nil {
		return fmt.Errorf("notify target is nil")
	}
	if !target.GetEnabled() {
		return nil
	}
	switch target.GetType() {
	case agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_TELEGRAM:
		return s.sendTelegram(ctx, target.GetTelegram(), msg)
	case agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_LARK_WEBHOOK:
		return s.sendLark(ctx, target.GetLark(), msg)
	case agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_DISCORD_WEBHOOK:
		return s.sendDiscord(ctx, target.GetDiscord(), msg)
	default:
		return fmt.Errorf("unsupported notify target type %s", target.GetType().String())
	}
}

func (s *Sender) sendTelegram(ctx context.Context, target *agentsv1.TelegramNotifyTarget, msg Message) error {
	if target.GetBotToken() == "" || target.GetChatId() == "" {
		return fmt.Errorf("telegram target requires bot_token and chat_id")
	}
	if !telegramBotTokenPattern.MatchString(target.GetBotToken()) {
		return fmt.Errorf("telegram bot_token does not match expected format <bot_id>:<secret>")
	}
	typingPayload := map[string]any{
		"chat_id": target.GetChatId(),
		"action":  "typing",
	}
	if target.GetMessageThreadId() != 0 {
		typingPayload["message_thread_id"] = target.GetMessageThreadId()
	}
	typingEndpoint := "https://api.telegram.org/bot" + target.GetBotToken() + "/sendChatAction"
	if err := s.postJSON(ctx, typingEndpoint, typingPayload); err != nil {
		return err
	}

	text := formatMessage(msg)
	if isTelegramMarkdownV2(target.GetParseMode()) {
		text = escapeTelegramMarkdownV2(text)
	}
	payload := map[string]any{
		"chat_id": target.GetChatId(),
		"text":    truncateForTelegram(text),
	}
	if target.GetParseMode() != "" {
		payload["parse_mode"] = target.GetParseMode()
	}
	if target.GetMessageThreadId() != 0 {
		payload["message_thread_id"] = target.GetMessageThreadId()
	}
	endpoint := "https://api.telegram.org/bot" + target.GetBotToken() + "/sendMessage"
	return s.postJSON(ctx, endpoint, payload)
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

func isTelegramMarkdownV2(parseMode string) bool {
	return strings.EqualFold(strings.TrimSpace(parseMode), "MarkdownV2")
}

func escapeTelegramMarkdownV2(text string) string {
	var b strings.Builder
	b.Grow(len(text))
	for _, r := range text {
		switch r {
		case '_', '*', '[', ']', '(', ')', '~', '`', '>', '#', '+', '-', '=', '|', '{', '}', '.', '!', '\\':
			b.WriteRune('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}

func larkSign(timestamp, secret string) string {
	mac := hmac.New(sha256.New, []byte(timestamp+"\n"+secret))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
