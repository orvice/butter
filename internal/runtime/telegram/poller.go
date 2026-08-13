package telegram

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"butterfly.orx.me/core/log"

	telegramrepo "go.orx.me/apps/butter/internal/repo/telegram"
	"go.orx.me/apps/butter/internal/secretbox"
	"go.orx.me/apps/butter/internal/telegramapi"
	"go.orx.me/apps/butter/internal/telegramqueue"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	// pollTimeoutSeconds is Telegram's own hold time. Long enough that an
	// idle Channel costs one request every half minute rather than a busy
	// loop; short enough that losing the lease is noticed promptly.
	pollTimeoutSeconds = 25
	// pollBatchLimit caps one fetch.
	pollBatchLimit = 50
	// pollLeaseTTL must exceed a full poll cycle, or a leader would lose its
	// own lease while legitimately waiting on Telegram.
	pollLeaseTTL = 90 * time.Second
	// pollSupervisorInterval is how often the supervisor re-reads Channels
	// and re-evaluates leadership.
	pollSupervisorInterval = 15 * time.Second
	// pollErrorBackoff throttles a Channel that keeps failing so one broken
	// credential cannot spin the CPU.
	pollErrorBackoff = 10 * time.Second
)

// PollingStatus is what an operator sees about one Channel's poller. It is
// observed, never persisted as configuration.
type PollingStatus struct {
	// Leader reports whether this Pod currently holds the Channel's lease.
	Leader bool
	// Offset is the last confirmed update ID boundary.
	Offset int64
	// LastFetchedUpdateID and LastAcceptedUpdateID separate "Telegram gave us
	// this" from "we durably kept it", which is exactly the gap an operator
	// needs to see when updates arrive but nothing happens.
	LastFetchedUpdateID  int64
	LastAcceptedUpdateID int64
	LastPolledAt         time.Time
	LastError            string
}

// Poller consumes Telegram updates by long polling.
//
// It exists as an alternative *transport*, not an alternative pipeline: a
// fetched update goes through the same Router, the same policy snapshot, the
// same Redis Stream, and the same workers as a webhook callback. Anything
// else would mean two receive paths to keep in agreement, and they would
// drift.
type Poller struct {
	repo      telegramrepo.Repository
	keyring   *secretbox.Keyring
	router    *Router
	newClient func(token string) telegramapi.PollingClient
	offsets   OffsetStore
	newLease  func(channelID string) Leader
	instance  string

	mu       sync.RWMutex
	statuses map[string]PollingStatus

	stop     chan struct{}
	stopped  chan struct{}
	stopOnce sync.Once

	// running tracks the per-Channel loops this Pod owns.
	runningMu sync.Mutex
	running   map[string]context.CancelFunc
}

// NewPoller builds the long-poll supervisor. `newClient` and `newLease` are
// injectable so tests can drive leadership and Telegram without either.
func NewPoller(
	repo telegramrepo.Repository,
	keyring *secretbox.Keyring,
	router *Router,
	offsets OffsetStore,
	newClient func(token string) telegramapi.PollingClient,
	newLease func(channelID string) Leader,
	instance string,
) *Poller {
	if newClient == nil {
		newClient = func(token string) telegramapi.PollingClient { return telegramapi.New(token) }
	}
	return &Poller{
		repo:      repo,
		keyring:   keyring,
		router:    router,
		newClient: newClient,
		offsets:   offsets,
		newLease:  newLease,
		instance:  instance,
		statuses:  make(map[string]PollingStatus),
		stop:      make(chan struct{}),
		stopped:   make(chan struct{}),
		running:   make(map[string]context.CancelFunc),
	}
}

// RedisPollingLeaseFactory builds per-Channel leases from a Redis client.
//
// The lease holder is the instance ID alone — not instance+channel — because
// the offset store fences commits by comparing the lease value, and it must
// see exactly what this Pod would write.
func RedisPollingLeaseFactory(rdb *redis.Client, instance string) func(string) Leader {
	return func(channelID string) Leader {
		return telegramqueue.NewLease(rdb, telegramqueue.PollingLeaseKey(channelID),
			instance, pollLeaseTTL)
	}
}

// Start runs the supervisor until Stop.
func (p *Poller) Start(ctx context.Context) {
	go func() {
		defer close(p.stopped)
		ticker := time.NewTicker(pollSupervisorInterval)
		defer ticker.Stop()
		p.supervise(ctx)
		for {
			select {
			case <-ctx.Done():
				p.stopAll()
				return
			case <-p.stop:
				p.stopAll()
				return
			case <-ticker.C:
				p.supervise(ctx)
			}
		}
	}()
}

// Stop halts every Channel loop.
func (p *Poller) Stop() {
	p.stopOnce.Do(func() { close(p.stop) })
	<-p.stopped
}

// Status returns the observed poller state for a Channel.
func (p *Poller) Status(channelID string) (PollingStatus, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	status, ok := p.statuses[channelID]
	return status, ok
}

// supervise starts loops for Channels that should be polled and stops the
// rest, so a mode switch or a disable takes effect within one interval.
func (p *Poller) supervise(ctx context.Context) {
	logger := log.FromContext(ctx)
	channels, err := p.repo.ListChannelsAcrossWorkspaces(ctx)
	if err != nil {
		logger.Error("telegram poller could not list channels", "err", err)
		return
	}

	wanted := make(map[string]struct{})
	for _, channel := range channels {
		if !shouldPoll(channel) {
			continue
		}
		wanted[channel.GetId()] = struct{}{}
		p.ensureRunning(ctx, channel)
	}

	p.runningMu.Lock()
	defer p.runningMu.Unlock()
	for channelID, cancel := range p.running {
		if _, ok := wanted[channelID]; !ok {
			// A Channel that switched to Webhook or was disabled stops here,
			// which is also what releases its lease.
			cancel()
			delete(p.running, channelID)
		}
	}
}

// shouldPoll reports whether a Channel wants long polling right now. A
// Webhook Channel never starts a poller: the two modes are exclusive, and
// running both would double-process every message.
func shouldPoll(channel *agentsv1.TelegramChannel) bool {
	return channel.GetInboundEnabled() &&
		channel.GetReceiveMode() == agentsv1.TelegramReceiveMode_TELEGRAM_RECEIVE_MODE_LONG_POLLING
}

func (p *Poller) ensureRunning(ctx context.Context, channel *agentsv1.TelegramChannel) {
	p.runningMu.Lock()
	defer p.runningMu.Unlock()
	if _, ok := p.running[channel.GetId()]; ok {
		return
	}
	loopCtx, cancel := context.WithCancel(ctx)
	p.running[channel.GetId()] = cancel
	go p.pollChannel(loopCtx, channel.GetId())
}

func (p *Poller) stopAll() {
	p.runningMu.Lock()
	defer p.runningMu.Unlock()
	for channelID, cancel := range p.running {
		cancel()
		delete(p.running, channelID)
	}
}

// pollChannel is one Channel's loop: hold the lease, fetch, route, commit.
func (p *Poller) pollChannel(ctx context.Context, channelID string) {
	logger := log.FromContext(ctx).With("channel_id", channelID)
	lease := p.newLease(channelID)
	defer func() {
		releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = lease.Release(releaseCtx)
		p.setLeader(channelID, false)
	}()

	for {
		if ctx.Err() != nil {
			return
		}
		held, err := lease.Acquire(ctx)
		if err != nil {
			p.recordError(channelID, err)
			if !sleepCtx(ctx, pollErrorBackoff) {
				return
			}
			continue
		}
		p.setLeader(channelID, held)
		if !held {
			// Another Pod is polling this Channel. Exactly one consumer per
			// Bot: two would each receive a different slice of updates.
			if !sleepCtx(ctx, pollSupervisorInterval) {
				return
			}
			continue
		}

		if err := p.pollOnce(ctx, channelID); err != nil {
			logger.Warn("telegram poll cycle failed", "err", err)
			p.recordError(channelID, err)
			if !sleepCtx(ctx, pollErrorBackoff) {
				return
			}
		}
	}
}

// pollOnce fetches one batch and advances the offset only as far as it is
// safe to.
func (p *Poller) pollOnce(ctx context.Context, channelID string) error {
	channel, err := p.repo.FindChannel(ctx, channelID)
	if err != nil {
		return fmt.Errorf("reload channel: %w", err)
	}
	if !shouldPoll(channel) {
		return nil
	}
	client, err := p.clientFor(ctx, channel)
	if err != nil {
		return err
	}

	offset, err := p.offsets.Get(ctx, channelID)
	if err != nil {
		return err
	}

	updates, err := client.GetUpdates(ctx, telegramapi.GetUpdatesParams{
		Offset:         offset,
		Limit:          pollBatchLimit,
		TimeoutSeconds: pollTimeoutSeconds,
	})
	if err != nil {
		if errors.Is(err, telegramapi.ErrWebhookActive) {
			// Telegram will not long-poll while a webhook is registered. The
			// reconciler removes it; say so rather than looking like a
			// credential problem.
			return fmt.Errorf("%w — the webhook reconciler will remove the stale registration", err)
		}
		return err
	}
	if len(updates) == 0 {
		p.recordPoll(ctx, channelID, 0, 0)
		return nil
	}

	logger := log.FromContext(ctx)
	var lastFetched, lastAccepted, commitTo int64
	for _, update := range updates {
		lastFetched = update.UpdateID

		decision, routeErr := p.router.Route(ctx, channel, update.Raw)
		switch {
		case routeErr != nil && errors.Is(routeErr, ErrMalformedUpdate):
			// Unusable on any retry, so confirming it is correct.
		case routeErr != nil:
			// We could not durably keep this update. Stop the batch and
			// leave the offset where it is: Telegram resends from there, and
			// (channel_id, update_id) dedupe suppresses anything we did
			// already accept.
			logger.Warn("telegram update was not accepted; leaving the offset unadvanced",
				"update_id", update.UpdateID, "err", routeErr)
			p.recordPoll(ctx, channelID, lastFetched, lastAccepted)
			if commitTo > 0 {
				return p.commit(ctx, channelID, commitTo)
			}
			return routeErr
		case decision == DecisionAccepted || decision == DecisionDuplicate:
			lastAccepted = update.UpdateID
		}
		// Accepted, duplicate, ignored, or permanently malformed: all four
		// are settled, so the offset may pass this update.
		commitTo = update.UpdateID + 1
	}

	p.recordPoll(ctx, channelID, lastFetched, lastAccepted)
	return p.commit(ctx, channelID, commitTo)
}

// commit advances the confirmed offset, tolerating the case where this Pod
// lost the lease mid-batch.
func (p *Poller) commit(ctx context.Context, channelID string, offset int64) error {
	if offset <= 0 {
		return nil
	}
	if err := p.offsets.Commit(ctx, channelID, p.instance, offset); err != nil {
		if errors.Is(err, ErrNotOffsetOwner) {
			// A stale owner must not confirm updates the new leader may still
			// be processing.
			log.FromContext(ctx).Info("telegram offset commit refused: lease moved",
				"channel_id", channelID)
			p.setLeader(channelID, false)
			return nil
		}
		return err
	}
	return nil
}

func (p *Poller) clientFor(ctx context.Context, channel *agentsv1.TelegramChannel) (telegramapi.PollingClient, error) {
	cred, err := p.repo.GetChannelCredential(ctx, channel.GetWorkspaceId(), channel.GetId())
	if err != nil {
		return nil, fmt.Errorf("read channel credential: %w", err)
	}
	token, err := p.keyring.Decrypt(ctx, cred.Ciphertext, cred.KeyID)
	if err != nil {
		return nil, fmt.Errorf("decrypt channel credential: %w", err)
	}
	return p.newClient(string(token)), nil
}

func (p *Poller) setLeader(channelID string, leader bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	status := p.statuses[channelID]
	status.Leader = leader
	p.statuses[channelID] = status
}

func (p *Poller) recordPoll(ctx context.Context, channelID string, fetched, accepted int64) {
	offset, _ := p.offsets.Get(ctx, channelID)
	p.mu.Lock()
	defer p.mu.Unlock()
	status := p.statuses[channelID]
	status.LastPolledAt = time.Now().UTC()
	status.Offset = offset
	status.LastError = ""
	if fetched > 0 {
		status.LastFetchedUpdateID = fetched
	}
	if accepted > 0 {
		status.LastAcceptedUpdateID = accepted
	}
	p.statuses[channelID] = status
}

func (p *Poller) recordError(channelID string, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	status := p.statuses[channelID]
	status.LastError = sanitizeProcessingError(err)
	p.statuses[channelID] = status
}

// sleepCtx waits, reporting false when the context ended first.
func sleepCtx(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
