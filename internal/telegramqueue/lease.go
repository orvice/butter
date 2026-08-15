package telegramqueue

import (
	"time"

	"github.com/redis/go-redis/v9"

	"go.orx.me/apps/butter/internal/redislease"
)

// Lease is a Redis-backed leadership token.
//
// Webhook reconciliation and Long Polling both need exactly one Pod acting at
// a time: several Pods calling setWebhook would race each other's
// registrations, and several Pods long-polling the same Bot would each
// receive a different slice of updates. The mechanics live in
// internal/redislease; this alias keeps the Telegram runtime's vocabulary.
type Lease = redislease.Lease

// NewLease builds a lease for the given logical role. `holder` must be unique
// per process (an instance ID).
func NewLease(rdb *redis.Client, key, holder string, ttl time.Duration) *Lease {
	return redislease.New(rdb, key, holder, ttl)
}

// Lease keys used by the Telegram runtime.
const (
	// WebhookReconcilerLeaseKey elects the single Pod that registers and
	// repairs Telegram webhooks.
	WebhookReconcilerLeaseKey = "butter:telegram:lease:webhook-reconciler"
	// PollingLeaseKeyPrefix elects the single Pod long-polling one Channel.
	PollingLeaseKeyPrefix = "butter:telegram:lease:polling:"
	// SessionLeaseKeyPrefix serializes work within one derived session.
	SessionLeaseKeyPrefix = "butter:telegram:lease:session:"
)

// SessionLeaseKey serializes turns inside one derived session. Two updates
// for the same conversation running concurrently would interleave their
// history writes; unrelated sessions are deliberately unaffected, which is
// what keeps the fleet parallel.
func SessionLeaseKey(sessionID string) string { return SessionLeaseKeyPrefix + sessionID }

// PollingLeaseKey returns the Long Polling lease key for one Channel.
func PollingLeaseKey(channelID string) string { return PollingLeaseKeyPrefix + channelID }
