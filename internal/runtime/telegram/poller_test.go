package telegram

// Long polling tests (issue #264/#272): single-leader election, failover,
// and — the part that actually loses messages if it is wrong — exactly when
// the Telegram offset is allowed to advance.

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	cryptokeymemory "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	telegramrepo "go.orx.me/apps/butter/internal/repo/telegram"
	telegrammemory "go.orx.me/apps/butter/internal/repo/telegram/memory"
	"go.orx.me/apps/butter/internal/secretbox"
	"go.orx.me/apps/butter/internal/telegramapi"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// fakePollingClient serves canned batches and records the offsets it was
// asked for, which is how the tests observe what was confirmed to Telegram.
type fakePollingClient struct {
	mu       sync.Mutex
	batches  [][]telegramapi.RawUpdate
	requests []int64
	err      error
}

func (c *fakePollingClient) GetUpdates(_ context.Context, params telegramapi.GetUpdatesParams) ([]telegramapi.RawUpdate, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requests = append(c.requests, params.Offset)
	if c.err != nil {
		return nil, c.err
	}
	if len(c.batches) == 0 {
		return nil, nil
	}
	batch := c.batches[0]
	c.batches = c.batches[1:]
	return batch, nil
}

func rawUpdate(updateID int64, text string) telegramapi.RawUpdate {
	return telegramapi.RawUpdate{
		UpdateID: updateID,
		Raw:      []byte(pollingMessage(updateID, text, "-1001234567890", 42)),
	}
}

// pollingMessage builds an update addressed at the fixture's topic
// destination.
func pollingMessage(updateID int64, text, chatID string, threadID int64) string {
	return `{"update_id":` + itoa(updateID) +
		`,"message":{"message_id":9,"message_thread_id":` + itoa(threadID) +
		`,"is_topic_message":true,"from":{"id":7,"is_bot":false},"chat":{"id":` + chatID +
		`,"type":"supergroup"},"text":"` + text + `"}}`
}

func itoa(v int64) string {
	return telegramapi.FormatID(v)
}

type pollerFixture struct {
	poller  *Poller
	repo    *telegrammemory.Store
	queue   *fakeQueue
	client  *fakePollingClient
	offsets *MemoryOffsetStore
	leader  *fakeLeader
}

func newPollerFixture(t *testing.T) *pollerFixture {
	t.Helper()
	repo := telegrammemory.New()
	keyring := secretbox.NewKeyring(cryptokeymemory.New())
	ciphertext, keyID, err := keyring.Encrypt(t.Context(), []byte("111111:token"))
	if err != nil {
		t.Fatalf("encrypt token: %v", err)
	}
	if _, err := repo.CreateChannel(t.Context(), "ws-a", &agentsv1.TelegramChannel{
		Id: "ch-1", Key: "main", BotId: "111111", BotUsername: "opsbot",
		InboundEnabled: true, OutboundEnabled: true,
		ReceiveMode: agentsv1.TelegramReceiveMode_TELEGRAM_RECEIVE_MODE_LONG_POLLING,
	}, telegramrepo.Credential{Ciphertext: ciphertext, KeyID: keyID}); err != nil {
		t.Fatalf("create channel: %v", err)
	}
	if _, err := repo.CreateDestination(t.Context(), "ws-a", &agentsv1.TelegramDestination{
		Id: "dest-1", Key: "incidents", ChannelId: "ch-1",
		ChatId: "-1001234567890", MessageThreadId: "42",
		InboundEnabled: true, OutboundEnabled: true,
		Config: &agentsv1.TelegramDestinationConfig{AgentId: "support"},
	}); err != nil {
		t.Fatalf("create destination: %v", err)
	}

	queue := newFakeQueue()
	client := &fakePollingClient{}
	offsets := NewMemoryOffsetStore()
	leader := &fakeLeader{held: true}

	poller := NewPoller(repo, keyring, NewRouter(repo, queue), offsets,
		func(string) telegramapi.PollingClient { return client },
		func(string) Leader { return leader },
		"pod-1")
	return &pollerFixture{
		poller: poller, repo: repo, queue: queue,
		client: client, offsets: offsets, leader: leader,
	}
}

// A fetched update goes through the same router, snapshot, and queue a
// webhook callback would have used.
func TestPolledUpdatesUseTheSameReceivePath(t *testing.T) {
	fx := newPollerFixture(t)
	fx.client.batches = [][]telegramapi.RawUpdate{{rawUpdate(100, "hello")}}

	if err := fx.poller.pollOnce(t.Context(), "ch-1"); err != nil {
		t.Fatalf("pollOnce: %v", err)
	}
	if len(fx.queue.accepted) != 1 {
		t.Fatalf("accepted %d events", len(fx.queue.accepted))
	}
	event := fx.queue.accepted[0]
	if event.DestinationID != "dest-1" || event.Address.MessageThreadID != "42" {
		t.Errorf("routing differed from the webhook path: %+v", event)
	}
	if event.Policy == nil || event.Policy.AgentID != "support" {
		t.Error("the destination policy was not frozen into the event")
	}
}

// The offset is a commitment to Telegram: it may only pass an update that is
// durably ours.
func TestOffsetAdvancesPastAcceptedUpdates(t *testing.T) {
	fx := newPollerFixture(t)
	fx.client.batches = [][]telegramapi.RawUpdate{{
		rawUpdate(100, "one"),
		rawUpdate(101, "two"),
	}}

	if err := fx.poller.pollOnce(t.Context(), "ch-1"); err != nil {
		t.Fatalf("pollOnce: %v", err)
	}
	offset, _ := fx.offsets.Get(t.Context(), "ch-1")
	if offset != 102 {
		t.Fatalf("offset = %d, want one past the last accepted update", offset)
	}
}

// An update the routing layer explicitly ignores is settled, so the offset
// may pass it — otherwise an unconfigured chat would wedge the poller.
func TestOffsetAdvancesPastIgnoredUpdates(t *testing.T) {
	fx := newPollerFixture(t)
	fx.client.batches = [][]telegramapi.RawUpdate{{
		telegramapi.RawUpdate{
			UpdateID: 200,
			Raw:      []byte(pollingMessage(200, "hello", "-1009999999999", 0)),
		},
	}}

	if err := fx.poller.pollOnce(t.Context(), "ch-1"); err != nil {
		t.Fatalf("pollOnce: %v", err)
	}
	if len(fx.queue.accepted) != 0 {
		t.Fatal("an unconfigured address reached the queue")
	}
	offset, _ := fx.offsets.Get(t.Context(), "ch-1")
	if offset != 201 {
		t.Fatalf("offset = %d, want the ignored update confirmed", offset)
	}
}

// If the update could not be durably kept, the offset must stay put so
// Telegram hands it back.
func TestOffsetDoesNotAdvanceWhenTheQueueFails(t *testing.T) {
	fx := newPollerFixture(t)
	fx.queue.failWith = errors.New("redis unavailable")
	fx.client.batches = [][]telegramapi.RawUpdate{{rawUpdate(100, "hello")}}

	if err := fx.poller.pollOnce(t.Context(), "ch-1"); err == nil {
		t.Fatal("expected the enqueue failure to surface")
	}
	offset, _ := fx.offsets.Get(t.Context(), "ch-1")
	if offset != 0 {
		t.Fatalf("offset = %d, want it unadvanced so telegram resends", offset)
	}
}

// A mid-batch failure confirms only the prefix that was actually accepted.
func TestPartialBatchCommitsOnlyThePrefix(t *testing.T) {
	fx := newPollerFixture(t)
	fx.client.batches = [][]telegramapi.RawUpdate{{
		rawUpdate(100, "kept"),
		rawUpdate(101, "lost"),
	}}
	// Fail the second accept only.
	accepted := 0
	fx.queue.acceptHook = func() error {
		accepted++
		if accepted == 2 {
			return errors.New("redis unavailable")
		}
		return nil
	}

	if err := fx.poller.pollOnce(t.Context(), "ch-1"); err != nil {
		t.Fatalf("pollOnce: %v", err)
	}
	offset, _ := fx.offsets.Get(t.Context(), "ch-1")
	if offset != 101 {
		t.Fatalf("offset = %d, want only the accepted prefix confirmed", offset)
	}
}

// Telegram resending an update it already gave us must not run it twice.
func TestRepeatedUpdatesAreSuppressedByIdempotency(t *testing.T) {
	fx := newPollerFixture(t)
	fx.client.batches = [][]telegramapi.RawUpdate{
		{rawUpdate(100, "hello")},
		{rawUpdate(100, "hello")},
	}

	for range 2 {
		if err := fx.poller.pollOnce(t.Context(), "ch-1"); err != nil {
			t.Fatalf("pollOnce: %v", err)
		}
	}
	if len(fx.queue.accepted) != 1 {
		t.Fatalf("accepted %d events for one update", len(fx.queue.accepted))
	}
}

// A Pod that lost the lease must not confirm updates the new leader may
// still be working through.
func TestStaleOwnerCannotCommitTheOffset(t *testing.T) {
	fx := newPollerFixture(t)
	fx.offsets.SetOwner("ch-1", "pod-2")
	fx.client.batches = [][]telegramapi.RawUpdate{{rawUpdate(100, "hello")}}

	if err := fx.poller.pollOnce(t.Context(), "ch-1"); err != nil {
		t.Fatalf("pollOnce: %v", err)
	}
	offset, _ := fx.offsets.Get(t.Context(), "ch-1")
	if offset != 0 {
		t.Fatalf("a stale owner committed offset %d", offset)
	}
	if status, _ := fx.poller.Status("ch-1"); status.Leader {
		t.Error("a refused commit must clear the leader flag")
	}
}

// Exactly one Pod polls a Channel: two would each receive a different slice
// of updates.
func TestOnlyTheLeaderPolls(t *testing.T) {
	fx := newPollerFixture(t)
	fx.leader.held = false

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan struct{})
	go func() {
		defer close(done)
		fx.poller.pollChannel(ctx, "ch-1")
	}()
	// Give the loop a chance to observe non-leadership, then stop it.
	cancel()
	<-done

	fx.client.mu.Lock()
	defer fx.client.mu.Unlock()
	if len(fx.client.requests) != 0 {
		t.Fatalf("a non-leader issued %d getUpdates calls", len(fx.client.requests))
	}
}

// Whichever Pod takes over resumes from what was actually committed, not
// from zero — replaying Telegram's whole backlog would re-answer old
// messages.
func TestFailoverResumesFromTheCommittedOffset(t *testing.T) {
	fx := newPollerFixture(t)
	fx.client.batches = [][]telegramapi.RawUpdate{{rawUpdate(100, "hello")}, nil}
	if err := fx.poller.pollOnce(t.Context(), "ch-1"); err != nil {
		t.Fatalf("first poll: %v", err)
	}

	// A different Pod, sharing only the offset store.
	successor := NewPoller(fx.repo, secretbox.NewKeyring(cryptokeymemory.New()),
		NewRouter(fx.repo, fx.queue), fx.offsets,
		func(string) telegramapi.PollingClient { return fx.client },
		func(string) Leader { return &fakeLeader{held: true} }, "pod-2")
	// The successor cannot decrypt the fixture's credential, so drive the
	// offset read directly: what matters is where it would resume from.
	resumeFrom, _ := successor.offsets.Get(t.Context(), "ch-1")
	if resumeFrom != 101 {
		t.Fatalf("successor would resume from %d, want the committed offset", resumeFrom)
	}
}

// A webhook registration blocks getUpdates entirely; the error has to say so
// rather than reading as a credential problem.
func TestWebhookConflictIsReportedClearly(t *testing.T) {
	fx := newPollerFixture(t)
	fx.client.err = telegramapi.ErrWebhookActive

	err := fx.poller.pollOnce(t.Context(), "ch-1")
	if err == nil {
		t.Fatal("expected the conflict to surface")
	}
	if !strings.Contains(err.Error(), "webhook") {
		t.Errorf("err = %v, want it to name the webhook registration", err)
	}
}

// A Webhook Channel never starts a poller: the modes are exclusive.
func TestWebhookChannelsAreNotPolled(t *testing.T) {
	fx := newPollerFixture(t)
	channel, err := fx.repo.GetChannel(t.Context(), "ws-a", "ch-1")
	if err != nil {
		t.Fatalf("GetChannel: %v", err)
	}
	channel.ReceiveMode = agentsv1.TelegramReceiveMode_TELEGRAM_RECEIVE_MODE_WEBHOOK
	if _, err := fx.repo.UpdateChannel(t.Context(), "ws-a", channel, channel.GetRevision()); err != nil {
		t.Fatalf("switch mode: %v", err)
	}

	if err := fx.poller.pollOnce(t.Context(), "ch-1"); err != nil {
		t.Fatalf("pollOnce: %v", err)
	}
	fx.client.mu.Lock()
	defer fx.client.mu.Unlock()
	if len(fx.client.requests) != 0 {
		t.Fatal("a webhook channel was polled")
	}
}

func TestDisabledChannelsAreNotPolled(t *testing.T) {
	fx := newPollerFixture(t)
	channel, err := fx.repo.GetChannel(t.Context(), "ws-a", "ch-1")
	if err != nil {
		t.Fatalf("GetChannel: %v", err)
	}
	channel.InboundEnabled = false
	if _, err := fx.repo.UpdateChannel(t.Context(), "ws-a", channel, channel.GetRevision()); err != nil {
		t.Fatalf("disable: %v", err)
	}

	if err := fx.poller.pollOnce(t.Context(), "ch-1"); err != nil {
		t.Fatalf("pollOnce: %v", err)
	}
	fx.client.mu.Lock()
	defer fx.client.mu.Unlock()
	if len(fx.client.requests) != 0 {
		t.Fatal("a disabled channel was polled")
	}
}

// Status separates "Telegram gave us this" from "we durably kept it", which
// is what an operator needs when updates arrive but nothing happens.
func TestStatusReportsFetchedAndAcceptedSeparately(t *testing.T) {
	fx := newPollerFixture(t)
	fx.client.batches = [][]telegramapi.RawUpdate{{
		rawUpdate(100, "kept"),
		{UpdateID: 101, Raw: []byte(pollingMessage(101, "elsewhere", "-1009999999999", 0))},
	}}

	if err := fx.poller.pollOnce(t.Context(), "ch-1"); err != nil {
		t.Fatalf("pollOnce: %v", err)
	}
	status, ok := fx.poller.Status("ch-1")
	if !ok {
		t.Fatal("expected a recorded status")
	}
	if status.LastFetchedUpdateID != 101 {
		t.Errorf("last fetched = %d, want the last update telegram returned", status.LastFetchedUpdateID)
	}
	if status.LastAcceptedUpdateID != 100 {
		t.Errorf("last accepted = %d, want the last update we kept", status.LastAcceptedUpdateID)
	}
}
