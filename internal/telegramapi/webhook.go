package telegramapi

import "context"

// WebhookInfo is the subset of Telegram `getWebhookInfo` the reconciler
// compares against desired state.
type WebhookInfo struct {
	// URL is empty when no webhook is registered.
	URL string
	// PendingUpdateCount is how many updates Telegram is holding.
	PendingUpdateCount int
	// LastErrorMessage is Telegram's own report of why delivery failed.
	LastErrorMessage string
	// MaxConnections and AllowedUpdates round out the observed registration.
	MaxConnections int
	AllowedUpdates []string
}

// SetWebhookParams describes a callback registration.
type SetWebhookParams struct {
	URL string
	// SecretToken is echoed back by Telegram in the
	// X-Telegram-Bot-Api-Secret-Token header on every callback.
	SecretToken string
	// AllowedUpdates narrows what Telegram sends. Empty means Telegram's
	// default set.
	AllowedUpdates []string
	// DropPendingUpdates discards Telegram's backlog at registration time.
	DropPendingUpdates bool
}

// WebhookClient is the receive-mode half of the Bot API. It is separate from
// Client so that callers which only send (Cron delivery, Notify Groups) do
// not depend on registration methods they never use.
type WebhookClient interface {
	GetWebhookInfo(ctx context.Context) (WebhookInfo, error)
	SetWebhook(ctx context.Context, params SetWebhookParams) error
	DeleteWebhook(ctx context.Context, dropPendingUpdates bool) error
}

type webhookInfoResult struct {
	URL                string   `json:"url"`
	PendingUpdateCount int      `json:"pending_update_count"`
	LastErrorMessage   string   `json:"last_error_message"`
	MaxConnections     int      `json:"max_connections"`
	AllowedUpdates     []string `json:"allowed_updates"`
}

func (c *HTTPClient) GetWebhookInfo(ctx context.Context) (WebhookInfo, error) {
	var result webhookInfoResult
	if err := c.call(ctx, "getWebhookInfo", nil, &result); err != nil {
		return WebhookInfo{}, err
	}
	return WebhookInfo{
		URL:                result.URL,
		PendingUpdateCount: result.PendingUpdateCount,
		LastErrorMessage:   result.LastErrorMessage,
		MaxConnections:     result.MaxConnections,
		AllowedUpdates:     result.AllowedUpdates,
	}, nil
}

func (c *HTTPClient) SetWebhook(ctx context.Context, params SetWebhookParams) error {
	payload := map[string]any{"url": params.URL}
	if params.SecretToken != "" {
		payload["secret_token"] = params.SecretToken
	}
	if len(params.AllowedUpdates) > 0 {
		payload["allowed_updates"] = params.AllowedUpdates
	}
	if params.DropPendingUpdates {
		payload["drop_pending_updates"] = true
	}
	return c.call(ctx, "setWebhook", payload, nil)
}

func (c *HTTPClient) DeleteWebhook(ctx context.Context, dropPendingUpdates bool) error {
	payload := map[string]any{}
	if dropPendingUpdates {
		payload["drop_pending_updates"] = true
	}
	return c.call(ctx, "deleteWebhook", payload, nil)
}
