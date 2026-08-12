package telegram

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"butterfly.orx.me/core/log"

	"go.orx.me/apps/butter/internal/telegramapi"
	"go.orx.me/apps/butter/internal/telegramqueue"
	"go.orx.me/apps/butter/internal/telegramsend"
)

const (
	// workerReadCount bounds how much one poll claims.
	workerReadCount = 16
	// workerBlock is how long a poll waits for work before looping.
	workerBlock = 5 * time.Second
	// reclaimIdle is how long an entry may sit unacknowledged before another
	// Pod takes it over — the recovery path for a crashed worker.
	reclaimIdle = 2 * time.Minute
)

// EventHandler processes one claimed event. Returning nil acknowledges it.
//
// This is the seam later tickets extend: #267 handles `/where` and
// acknowledges everything else, #268 adds Agent invocation behind the same
// contract without changing how work is claimed or acknowledged.
type EventHandler interface {
	Handle(ctx context.Context, event *telegramqueue.Event) error
}

// Worker consumes accepted updates from the durable queue.
//
// Any Pod can claim any Channel's work: that is what makes the fleet
// horizontally scalable, and it is why the event carries a frozen routing
// snapshot rather than a pointer into one Pod's memory.
type Worker struct {
	queue    *telegramqueue.Queue
	handler  EventHandler
	consumer string
	stop     chan struct{}
	stopped  chan struct{}
	stopOnce sync.Once
}

func NewWorker(queue *telegramqueue.Queue, handler EventHandler, consumer string) *Worker {
	return &Worker{
		queue:    queue,
		handler:  handler,
		consumer: consumer,
		stop:     make(chan struct{}),
		stopped:  make(chan struct{}),
	}
}

// Start begins consuming. It first drains this consumer's own pending
// entries — the work it was holding when the process last died — before
// taking anything new.
func (w *Worker) Start(ctx context.Context) error {
	if !w.queue.Available() {
		return fmt.Errorf("telegram worker requires a configured queue")
	}
	if err := w.queue.EnsureGroup(ctx); err != nil {
		return err
	}
	go w.run(ctx)
	return nil
}

// Stop halts the loop.
func (w *Worker) Stop() {
	w.stopOnce.Do(func() { close(w.stop) })
	<-w.stopped
}

func (w *Worker) run(ctx context.Context) {
	defer close(w.stopped)
	logger := log.FromContext(ctx)

	if pending, err := w.queue.ReadPending(ctx, w.consumer, workerReadCount); err != nil {
		logger.Warn("could not recover pending telegram work", "err", err)
	} else {
		w.process(ctx, pending)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stop:
			return
		default:
		}

		// Take over anything a dead Pod abandoned before claiming new work,
		// so a crash delays a message rather than losing it.
		if reclaimed, err := w.queue.Claim(ctx, w.consumer, reclaimIdle, workerReadCount); err != nil {
			logger.Debug("telegram reclaim failed", "err", err)
		} else {
			w.process(ctx, reclaimed)
		}

		deliveries, err := w.queue.Read(ctx, w.consumer, workerReadCount, workerBlock)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			logger.Warn("telegram worker read failed", "err", err)
			continue
		}
		w.process(ctx, deliveries)
	}
}

func (w *Worker) process(ctx context.Context, deliveries []telegramqueue.Delivery) {
	logger := log.FromContext(ctx)
	for _, delivery := range deliveries {
		if err := w.handler.Handle(ctx, delivery.Event); err != nil {
			// Leave it unacknowledged: it stays in the pending list and is
			// reclaimed after reclaimIdle, by this Pod or another.
			logger.Error("telegram event handling failed",
				"stream_id", delivery.ID, "channel_id", delivery.Event.ChannelID,
				"update_id", delivery.Event.UpdateID, "err", err)
			continue
		}
		if err := w.queue.Ack(ctx, delivery.ID); err != nil {
			logger.Warn("could not acknowledge telegram event", "stream_id", delivery.ID, "err", err)
		}
	}
}

// WhereHandler answers the transport-level `/where` command and acknowledges
// everything else.
//
// It is a separate handler from Agent execution because `/where` deliberately
// has none of that machinery: no Destination, no admission check, no session,
// no Agent. Its whole purpose is to work where nothing is configured yet.
type WhereHandler struct {
	sender *telegramsend.Sender
	// next handles events this handler does not own. Nil acknowledges them.
	next EventHandler
}

func NewWhereHandler(sender *telegramsend.Sender, next EventHandler) *WhereHandler {
	return &WhereHandler{sender: sender, next: next}
}

func (h *WhereHandler) Handle(ctx context.Context, event *telegramqueue.Event) error {
	if event.Kind != telegramqueue.KindWhere {
		if h.next != nil {
			return h.next.Handle(ctx, event)
		}
		return nil
	}

	update, err := telegramapi.ParseUpdate(event.Update)
	if err != nil {
		// Nothing to answer, and retrying cannot help.
		return nil
	}
	msg, _ := update.RoutableMessage()
	userID := ""
	if msg != nil && msg.From != nil {
		userID = telegramapi.FormatID(msg.From.ID)
	}

	// Reply to the update's original chat and thread, which is exactly the
	// address the caller is asking about — and which by definition may not
	// have a Destination.
	if _, err := h.sender.SendRaw(ctx, event.WorkspaceID, event.ChannelID,
		event.Address.ChatID, event.Address.MessageThreadID,
		formatWhere(event, userID)); err != nil {
		return fmt.Errorf("answer /where: %w", err)
	}
	log.FromContext(ctx).Info("answered telegram /where",
		"channel_id", event.ChannelID, "chat_id", event.Address.ChatID,
		"message_thread_id", event.Address.MessageThreadID)
	return nil
}

// formatWhere renders the identifiers an administrator needs to create a
// Destination. It deliberately reveals nothing else: no workspace metadata,
// no credentials, no other Destination configuration.
func formatWhere(event *telegramqueue.Event, userID string) string {
	var b strings.Builder
	b.WriteString("Telegram identifiers for this chat:\n")
	fmt.Fprintf(&b, "channel_id: %s\n", event.ChannelID)
	fmt.Fprintf(&b, "chat_id: %s\n", event.Address.ChatID)
	if event.Address.MessageThreadID != "" {
		fmt.Fprintf(&b, "message_thread_id: %s\n", event.Address.MessageThreadID)
	} else {
		b.WriteString("message_thread_id: (none — not a forum topic)\n")
	}
	if userID != "" {
		fmt.Fprintf(&b, "user_id: %s", userID)
	}
	return strings.TrimRight(b.String(), "\n")
}
