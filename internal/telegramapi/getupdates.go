package telegramapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

// GetUpdatesParams describes one long-poll fetch.
type GetUpdatesParams struct {
	// Offset is the first update ID to return. Telegram treats a request
	// with offset N as confirmation that everything below N was handled, so
	// this doubles as the commit — which is why the caller must only ever
	// advance it after an update is durably accepted.
	Offset int64
	// Limit caps how many updates one fetch returns.
	Limit int
	// TimeoutSeconds is Telegram's own long-poll hold time. A non-zero value
	// is what makes this a long poll rather than a busy loop.
	TimeoutSeconds int
	// AllowedUpdates narrows what Telegram sends. Empty means its default.
	AllowedUpdates []string
}

// ErrWebhookActive means Telegram refuses getUpdates because a webhook is
// registered for this Bot. Telegram reports it as HTTP 409, and it is the
// single most likely reason a freshly switched Channel receives nothing.
var ErrWebhookActive = errWebhookActive{}

type errWebhookActive struct{}

func (errWebhookActive) Error() string {
	return "telegram refuses getUpdates while a webhook is registered"
}

// isWebhookConflict recognizes Telegram's 409 for "a webhook is set".
func (e *APIError) isWebhookConflict() bool { return e.Code == http.StatusConflict }

// errorsAsAPIError unwraps a Telegram API error, if the failure was one.
func errorsAsAPIError(err error) (*APIError, bool) {
	return errors.AsType[*APIError](err)
}

// PollingClient is the long-poll half of the Bot API. It is separate from
// Client so that Pods which never poll do not depend on it.
type PollingClient interface {
	// GetUpdates fetches the next batch. The raw JSON of each update is
	// returned so the caller can route on exactly the same bytes a webhook
	// callback would have delivered.
	GetUpdates(ctx context.Context, params GetUpdatesParams) ([]RawUpdate, error)
}

// RawUpdate pairs an update ID with the bytes it came in, so the receive path
// can dedupe on the ID while routing the original payload.
type RawUpdate struct {
	UpdateID int64
	Raw      json.RawMessage
}

func (c *HTTPClient) GetUpdates(ctx context.Context, params GetUpdatesParams) ([]RawUpdate, error) {
	payload := map[string]any{}
	if params.Offset > 0 {
		payload["offset"] = params.Offset
	}
	if params.Limit > 0 {
		payload["limit"] = params.Limit
	}
	if params.TimeoutSeconds > 0 {
		payload["timeout"] = params.TimeoutSeconds
	}
	if len(params.AllowedUpdates) > 0 {
		payload["allowed_updates"] = params.AllowedUpdates
	}

	// Decode into raw messages rather than a typed slice: the router works
	// on the original bytes so long polling and webhooks share one code path
	// down to the payload.
	var results []json.RawMessage
	if err := c.call(ctx, "getUpdates", payload, &results); err != nil {
		if apiErr, ok := errorsAsAPIError(err); ok && apiErr.isWebhookConflict() {
			return nil, ErrWebhookActive
		}
		return nil, err
	}

	updates := make([]RawUpdate, 0, len(results))
	for _, raw := range results {
		var envelope struct {
			UpdateID int64 `json:"update_id"`
		}
		if err := json.Unmarshal(raw, &envelope); err != nil {
			// An update we cannot even read the ID of cannot be deduped or
			// committed; skipping it is the only safe option, and Telegram
			// will not resend it once the batch is confirmed.
			continue
		}
		updates = append(updates, RawUpdate{UpdateID: envelope.UpdateID, Raw: raw})
	}
	return updates, nil
}
