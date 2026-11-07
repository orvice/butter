package telegramapi

import "context"

// ReactionParams addresses an existing message within a chat. Message IDs are
// chat-scoped, so the Bot API takes no message_thread_id for reactions.
type ReactionParams struct {
	ChatID    string
	MessageID string
	Emoji     string
}

func (c *HTTPClient) SetMessageReaction(ctx context.Context, params ReactionParams) error {
	return c.call(ctx, "setMessageReaction", map[string]any{
		"chat_id":    params.ChatID,
		"message_id": params.MessageID,
		"reaction": []map[string]string{
			{"type": "emoji", "emoji": params.Emoji},
		},
	}, nil)
}
