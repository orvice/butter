package telegramsend

import (
	"context"
	"errors"
	"time"

	"go.orx.me/apps/butter/internal/telegramapi"
)

// React acknowledges an inbound message without posting a notification-producing
// placeholder. It uses the Destination's current address and credential, just
// like a send. Acknowledgement is optional: bound it so a slow API or rate limit
// cannot hold up the Agent. Callers should log errors and continue the turn.
func (s *Sender) React(ctx context.Context, workspaceID, destinationID, messageID, emoji string) error {
	if messageID == "" {
		return errors.New("reaction message ID is empty")
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	resolved, err := s.Resolve(ctx, workspaceID, destinationID)
	if err != nil {
		return err
	}
	params := telegramapi.ReactionParams{
		ChatID:    resolved.Destination.GetChatId(),
		MessageID: messageID,
		Emoji:     emoji,
	}
	for attempt := 1; ; attempt++ {
		err := resolved.Client.SetMessageReaction(ctx, params)
		if err == nil {
			return nil
		}
		wait, ok := telegramapi.RetryAfter(err)
		if !ok || attempt == s.maxAttempts {
			return err
		}
		if err := s.sleep(ctx, wait); err != nil {
			return err
		}
	}
}
