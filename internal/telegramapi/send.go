package telegramapi

import (
	"context"
	"strconv"
)

// ParseMode values accepted by the Bot API.
const (
	ParseModeMarkdownV2 = "MarkdownV2"
	// ParseModeNone sends the text verbatim. It is the fallback used when
	// Telegram rejects formatted text.
	ParseModeNone = ""
)

// MaxMessageRunes is the Bot API limit for the `text` field. Telegram counts
// UTF-16 code units, but runes are a safe under-approximation for every
// character that fits in the BMP and never over-estimates for the rest.
const MaxMessageRunes = 4096

// SendMessageParams describes one outbound Telegram message.
//
// MessageThreadID is a *string rather than an int64 so "no topic" is a
// distinct value from thread 0: including `message_thread_id: 0` in the
// payload makes Telegram reject the send, while omitting it targets the
// group's general conversation.
type SendMessageParams struct {
	ChatID          string
	MessageThreadID string
	Text            string
	ParseMode       string
	// ReplyToMessageID quotes an existing message. Empty sends standalone.
	ReplyToMessageID string
	// DisableNotification sends silently.
	DisableNotification bool
}

// Message is the subset of a sent Telegram message Butter records.
type Message struct {
	ID              int64
	ChatID          int64
	MessageThreadID int64
}

// EditMessageParams describes an in-place edit of a previously sent message.
type EditMessageParams struct {
	ChatID    string
	MessageID string
	Text      string
	ParseMode string
}

func (c *HTTPClient) SendMessage(ctx context.Context, params SendMessageParams) (Message, error) {
	payload := map[string]any{
		"chat_id": params.ChatID,
		"text":    params.Text,
	}
	if params.ParseMode != ParseModeNone {
		payload["parse_mode"] = params.ParseMode
	}
	if params.MessageThreadID != "" {
		payload["message_thread_id"] = params.MessageThreadID
	}
	if params.ReplyToMessageID != "" {
		payload["reply_parameters"] = map[string]any{
			"message_id": params.ReplyToMessageID,
			// The reply target may have been deleted; delivering the response
			// without the quote beats dropping it.
			"allow_sending_without_reply": true,
		}
	}
	if params.DisableNotification {
		payload["disable_notification"] = true
	}
	return c.sendMessagePayload(ctx, "sendMessage", payload)
}

func (c *HTTPClient) EditMessageText(ctx context.Context, params EditMessageParams) (Message, error) {
	payload := map[string]any{
		"chat_id":    params.ChatID,
		"message_id": params.MessageID,
		"text":       params.Text,
	}
	if params.ParseMode != ParseModeNone {
		payload["parse_mode"] = params.ParseMode
	}
	return c.sendMessagePayload(ctx, "editMessageText", payload)
}

type sentMessage struct {
	MessageID       int64 `json:"message_id"`
	MessageThreadID int64 `json:"message_thread_id"`
	Chat            struct {
		ID int64 `json:"id"`
	} `json:"chat"`
}

func (c *HTTPClient) sendMessagePayload(ctx context.Context, method string, payload map[string]any) (Message, error) {
	var result sentMessage
	if err := c.call(ctx, method, payload, &result); err != nil {
		return Message{}, err
	}
	return Message{
		ID:              result.MessageID,
		ChatID:          result.Chat.ID,
		MessageThreadID: result.MessageThreadID,
	}, nil
}

// FormatMessageID renders a Telegram message ID in the decimal-string form
// Butter records alongside chat and thread IDs.
func FormatMessageID(id int64) string { return strconv.FormatInt(id, 10) }
