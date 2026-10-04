package http

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"butterfly.orx.me/core/log"
	aguievents "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/events"
	aguisse "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/encoding/sse"
	"github.com/gin-gonic/gin"

	"go.orx.me/apps/butter/internal/runtime/runlog"
)

// The Run Log (ADR-0016 decision 5, #404): a Detached Run's events go to a
// bounded, short-lived log, and every observer replays the log from
// RUN_STARTED and follows it to the run's end, on any Pod. The POST response
// is the first observer; GET …/threads/:thread_id/run adds more. A live view
// and a replay therefore see the same sequence.
//
// The run never waits on the log: its writer queues events, coalescing text
// deltas, and appends them in the background, bounded. An observer that
// cannot follow the run from RUN_STARTED to its end — the log stopped short,
// lapsed, or the run vanished without its end — ends its stream with the
// fallback marker, and the client reads the thread instead. The log is a
// replay buffer: the session stays the record of the conversation.

// AGUIRunLogKeyPrefix namespaces the Run Logs of Detached Runs. A log is keyed
// like the thread lease, by caller and thread, and by its run's Invocation
// ID: butter:agui:log:session:{user}:agui-{threadId}:{invocationId}.
const AGUIRunLogKeyPrefix = "butter:agui:log:session:"

// AGUIRunLogRetention is how long a Detached Run's Run Log, and its run state
// marked ended, are kept once the run ended: a late observer replays the
// whole run within it.
const AGUIRunLogRetention = 5 * time.Minute

// aguiRunLogKey is the key of a run's log: its thread's key and its
// Invocation ID, which no other run shares.
func aguiRunLogKey(thread, invocationID string) string {
	return thread + ":" + invocationID
}

// aguiRunLogLimits bound a Detached Run's Run Log, and pace its writer and
// its observers. Tests shrink them.
type aguiRunLogLimits struct {
	// maxEntries and maxBytes cap one run's log. A run whose events outgrow
	// them goes on; its log ends with a truncation marker.
	maxEntries int
	maxBytes   int
	// maxPending bounds the bytes a writer holds while the store catches up.
	// A writer that falls further behind takes the truncation path. It is
	// as large as maxBytes, so a store that keeps up never truncates a log
	// its cap holds.
	maxPending int
	// linger is how long a writer holds a text delta for the next one to
	// join it, so consecutive deltas become one entry.
	linger time.Duration
	// ttl keeps a log while its run lives. The writer renews it, so the log
	// of a run whose process died lapses within one ttl.
	ttl time.Duration
	// retention keeps a log, and its run's ended run state, once the run
	// ended.
	retention time.Duration
	// closeWait bounds how long a run that ended waits for its last events
	// to reach its log before it lets its thread go.
	closeWait time.Duration
	// check paces how often a waiting observer looks at the run state to
	// tell whether the run still runs.
	check time.Duration
	// endGrace is how long an observer that saw its run end, or vanish,
	// still waits for the run's end to reach the log.
	endGrace time.Duration
}

func defaultAGUIRunLogLimits() aguiRunLogLimits {
	return aguiRunLogLimits{
		maxEntries: 50000,
		maxBytes:   8 << 20,
		maxPending: 8 << 20,
		linger:     50 * time.Millisecond,
		ttl:        AGUISessionLeaseTTL,
		retention:  AGUIRunLogRetention,
		closeWait:  10 * time.Second,
		check:      2 * time.Second,
		endGrace:   15 * time.Second,
	}
}

const (
	// aguiRunLogAppendTimeout bounds one write to the log store.
	aguiRunLogAppendTimeout = 5 * time.Second
	// aguiRunLogMaxBackoff caps the pause before a failed append is retried.
	aguiRunLogMaxBackoff = time.Second
	// aguiRunLogReadBatch is how many entries an observer reads at once.
	aguiRunLogReadBatch = 256
	// aguiRunLogReadRetry paces an observer whose reads fail.
	aguiRunLogReadRetry = 250 * time.Millisecond
	// aguiRunLogTextOverhead stands for the encoding around a text delta,
	// which a writer holds unencoded until it appends it.
	aguiRunLogTextOverhead = 128
)

// aguiHeartbeatInterval is how long an observer's stream may stay quiet
// before it sends a heartbeat, so proxies keep the connection open while the
// run thinks or waits on a tool.
const aguiHeartbeatInterval = 15 * time.Second

// aguiHeartbeat is an SSE comment. AG-UI clients skip frames without data.
var aguiHeartbeat = []byte(": heartbeat\n\n")

// aguiFallbackEvent names the Butter CUSTOM event an observer ends its stream
// with when it cannot follow the run from RUN_STARTED to its end: the client
// reads the thread instead (ADR-0016 decision 5).
const aguiFallbackEvent = "butter.fallback"

// Why an observer fell back.
const (
	// aguiFallbackTruncated: the log stopped short of the run's end. The
	// run's events outgrew it, its writer fell behind the store, or the run
	// ended without its last events reaching it.
	aguiFallbackTruncated = "truncated"
	// aguiFallbackExpired: the log is gone. It lapsed, or its thread was
	// deleted.
	aguiFallbackExpired = "expired"
	// aguiFallbackLost: the run lost its thread without its end reaching the
	// log, as when its server died, or the log could not be read.
	aguiFallbackLost = "lost"
)

// aguiFallbackValue is the value of the fallback marker.
type aguiFallbackValue struct {
	ThreadID string `json:"threadId"`
	RunID    string `json:"runId"`
	Reason   string `json:"reason"`
}

// aguiSSEFrame encodes ev as one SSE frame, as a run in its request writes it.
// The encoder refuses a cancelled context, and a run's terminal event is
// often sent after its context ended, so it encodes on its own.
func aguiSSEFrame(enc *aguisse.SSEWriter, ev aguievents.Event) ([]byte, error) {
	var frame bytes.Buffer
	if err := enc.WriteEvent(context.Background(), &frame, ev); err != nil {
		return nil, err
	}
	return frame.Bytes(), nil
}

// aguiTerminal reports whether ev ends its run's stream.
func aguiTerminal(ev aguievents.Event) bool {
	switch ev.Type() {
	case aguievents.EventTypeRunFinished, aguievents.EventTypeRunError:
		return true
	default:
		return false
	}
}

// --- The writer ----------------------------------------------------------------

// aguiLogWriter appends a Detached Run's events to its Run Log. Its emit is
// the run's aguiEmitter: it encodes the event and queues it, and never waits
// on the store, so the store never stalls the model loop. A goroutine of its
// own appends what is queued in batches, coalescing consecutive
// TEXT_MESSAGE_CONTENT deltas. The log is capped by entries and bytes, and
// the writer holds a bounded queue: past either bound the log ends with a
// truncation marker and the run goes on.
type aguiLogWriter struct {
	store  runlog.Store
	key    string
	limits aguiRunLogLimits
	// enc encodes on the run's side; flushEnc the coalesced text, on the
	// writer's.
	enc      *aguisse.SSEWriter
	flushEnc *aguisse.SSEWriter

	mu      sync.Mutex
	queue   []aguiLogItem
	pending int
	// truncated reports that the log stops short: the writer fell behind or
	// the log hit its cap. Nothing more is queued, and the log ends with a
	// truncation marker.
	truncated bool
	closing   bool
	wake      chan struct{}

	// ctx is the appends'; cancelled once the run gives up waiting for them.
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}

	// The writer goroutine's own. Read by others once done is closed.
	appended int  // entries in the log
	bytes    int  // bytes of those entries
	final    bool // the log holds its last entry: the run's end or a marker
	ended    bool // that last entry is the run's end
	gone     bool // the log no longer exists, or no longer holds what was appended
}

// aguiLogItem is one queued event: encoded, or text whose deltas still
// coalesce.
type aguiLogItem struct {
	kind  runlog.Kind
	frame []byte
	text  *aguievents.TextMessageContentEvent
	delta *strings.Builder
	// since is when text was queued: the writer holds it back for the
	// deltas that follow until since plus the linger.
	since time.Time
	// size is what the item counts toward the writer's bound.
	size int
}

// newAGUILogWriter starts the writer of the log at key, which is open.
func newAGUILogWriter(store runlog.Store, key string, limits aguiRunLogLimits) *aguiLogWriter {
	ctx, cancel := context.WithCancel(context.Background())
	w := &aguiLogWriter{
		store: store, key: key, limits: limits,
		enc: aguisse.NewSSEWriter(), flushEnc: aguisse.NewSSEWriter(),
		wake: make(chan struct{}, 1), ctx: ctx, cancel: cancel, done: make(chan struct{}),
	}
	go w.run()
	return w
}

// emit queues ev for the log. It fails only for an event that cannot be
// encoded. Once the log stops short, or the run ended, it drops ev.
func (w *aguiLogWriter) emit(ev aguievents.Event) error {
	if text, ok := ev.(*aguievents.TextMessageContentEvent); ok {
		if err := text.Validate(); err != nil {
			return err
		}
		w.push(aguiLogItem{kind: runlog.KindEvent, text: text})
		return nil
	}
	frame, err := aguiSSEFrame(w.enc, ev)
	if err != nil {
		return err
	}
	kind := runlog.KindEvent
	if aguiTerminal(ev) {
		kind = runlog.KindEnd
	}
	w.push(aguiLogItem{kind: kind, frame: frame})
	return nil
}

func (w *aguiLogWriter) push(item aguiLogItem) {
	w.mu.Lock()
	if w.truncated || w.closing {
		w.mu.Unlock()
		return
	}
	switch {
	case item.text != nil && len(w.queue) > 0 && w.queue[len(w.queue)-1].text != nil &&
		w.queue[len(w.queue)-1].text.MessageID == item.text.MessageID:
		// A delta right after another of the same message joins it.
		tail := &w.queue[len(w.queue)-1]
		tail.delta.WriteString(item.text.Delta)
		tail.size += len(item.text.Delta)
		w.pending += len(item.text.Delta)
	case item.text != nil:
		item.delta = &strings.Builder{}
		item.delta.WriteString(item.text.Delta)
		item.since = time.Now()
		item.size = len(item.text.Delta) + aguiRunLogTextOverhead
		w.queue = append(w.queue, item)
		w.pending += item.size
	default:
		item.size = len(item.frame)
		w.queue = append(w.queue, item)
		w.pending += item.size
	}
	if w.pending > w.limits.maxPending {
		// The store is too far behind. The log stops here, and the run goes
		// on.
		w.truncated = true
		w.queue, w.pending = nil, 0
	}
	w.mu.Unlock()
	w.signal()
}

func (w *aguiLogWriter) signal() {
	select {
	case w.wake <- struct{}{}:
	default:
	}
}

// take hands the writer goroutine what is queued, whether the log stops
// short, and whether the run ended. Text the queue ends with stays queued
// for the deltas that follow until its linger is over, unless the run ended:
// holdUntil says until when.
func (w *aguiLogWriter) take(now time.Time) (items []aguiLogItem, truncated, closing bool, holdUntil time.Time) {
	w.mu.Lock()
	defer w.mu.Unlock()
	items, w.queue, w.pending = w.queue, nil, 0
	if n := len(items); n > 0 && !w.closing && !w.truncated {
		if last := items[n-1]; last.text != nil && now.Before(last.since.Add(w.limits.linger)) {
			items, holdUntil = items[:n-1], last.since.Add(w.limits.linger)
			w.queue, w.pending = []aguiLogItem{last}, last.size
		}
	}
	return items, w.truncated, w.closing, holdUntil
}

// stopShort makes the log end with a truncation marker: nothing more is
// queued.
func (w *aguiLogWriter) stopShort() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.truncated = true
	w.queue, w.pending = nil, 0
}

// aguiLogBatch is what the next append carries. An append that fails is
// retried with the batch as it is, so the store can tell a retry whose first
// attempt went through.
type aguiLogBatch struct {
	entries []runlog.Entry
	bytes   int
}

func (b *aguiLogBatch) final() bool {
	return len(b.entries) > 0 && b.entries[len(b.entries)-1].Kind.Final()
}

// extend adds the taken items to the batch, as log entries, within the log's
// caps. A log that would outgrow its caps, or whose writer fell behind, ends
// with a truncation marker instead. Nothing is added after a final entry.
func (w *aguiLogWriter) extend(b *aguiLogBatch, items []aguiLogItem, truncated bool) {
	if w.final || w.gone || b.final() {
		return
	}
	for _, item := range items {
		frame := item.frame
		if item.text != nil {
			item.text.Delta = item.delta.String()
			encoded, err := aguiSSEFrame(w.flushEnc, item.text)
			if err != nil {
				// The deltas were each valid; their sum always is.
				log.FromContext(w.ctx).Error("agui run log: text could not be encoded", "err", err)
				truncated = true
				break
			}
			frame = encoded
		}
		if w.appended+len(b.entries)+1 > w.limits.maxEntries || w.bytes+b.bytes+len(frame) > w.limits.maxBytes {
			log.FromContext(w.ctx).Warn("agui run log reached its cap; it ends here and the run goes on",
				"log", w.key, "entries", w.appended+len(b.entries), "bytes", w.bytes+b.bytes)
			w.stopShort()
			truncated = true
			break
		}
		b.entries = append(b.entries, runlog.Entry{Kind: item.kind, Frame: frame})
		b.bytes += len(frame)
		if item.kind.Final() {
			return
		}
	}
	if truncated {
		b.entries = append(b.entries, runlog.Entry{Kind: runlog.KindTruncated})
	}
}

// run appends what is queued until the run ended and its last events are in
// the log, or the run gave up waiting for them.
func (w *aguiLogWriter) run() {
	defer close(w.done)
	keepAlive := time.NewTicker(max(w.limits.ttl/3, time.Millisecond))
	defer keepAlive.Stop()
	var batch aguiLogBatch
	backoff := time.Duration(0)
	held := time.NewTimer(time.Hour)
	held.Stop()
	defer held.Stop()
	for {
		if len(batch.entries) == 0 {
			items, truncated, closing, holdUntil := w.take(time.Now())
			w.extend(&batch, items, truncated)
			if len(batch.entries) == 0 {
				if closing {
					w.keep(w.limits.retention)
					return
				}
				// Wait for more, or for the end of the held text's linger.
				if !holdUntil.IsZero() {
					held.Reset(time.Until(holdUntil))
				}
				select {
				case <-w.wake:
				case <-held.C:
				case <-keepAlive.C:
					w.keep(w.limits.ttl)
				case <-w.ctx.Done():
					return
				}
				held.Stop()
				continue
			}
		}
		err := w.append(batch)
		switch {
		case err == nil:
			backoff = 0
			w.appended += len(batch.entries)
			w.bytes += batch.bytes
			if batch.final() {
				w.final = true
				w.ended = batch.entries[len(batch.entries)-1].Kind == runlog.KindEnd
			}
			batch = aguiLogBatch{}
		case errors.Is(err, runlog.ErrGone), errors.Is(err, runlog.ErrConflict):
			// Nothing more reaches the log. Its observers find it gone, or
			// stop at its last entry and learn from the run state that the
			// run ended.
			log.FromContext(w.ctx).Warn("agui run log is gone; the run goes on without it", "log", w.key, "err", err)
			w.gone = true
			batch = aguiLogBatch{}
		default:
			if w.ctx.Err() != nil {
				return
			}
			backoff = min(max(2*backoff, 50*time.Millisecond), aguiRunLogMaxBackoff)
			log.FromContext(w.ctx).Warn("agui run log append failed; retrying", "log", w.key, "err", err, "retry_in", backoff.String())
			select {
			case <-w.ctx.Done():
				return
			case <-time.After(backoff):
			}
		}
	}
}

func (w *aguiLogWriter) append(b aguiLogBatch) error {
	ctx, cancel := context.WithTimeout(w.ctx, aguiRunLogAppendTimeout)
	defer cancel()
	return w.store.Append(ctx, w.key, w.appended, b.entries, w.limits.ttl)
}

// keep keeps the log for ttl from now: renewed while the run lives, and for
// its retention once the run ended.
func (w *aguiLogWriter) keep(ttl time.Duration) {
	if w.gone {
		return
	}
	ctx, cancel := context.WithTimeout(w.ctx, aguiRunLogAppendTimeout)
	defer cancel()
	err := w.store.Expire(ctx, w.key, ttl)
	switch {
	case errors.Is(err, runlog.ErrGone):
		w.gone = true
	case err != nil:
		log.FromContext(w.ctx).Warn("agui run log could not be kept; it lapses on its own", "log", w.key, "err", err)
	}
}

// close lets the writer append what is queued, the run's end with it, and
// waits for it at most limits.closeWait; then the log is kept for its
// retention. It reports whether the log holds the run's end.
func (w *aguiLogWriter) close() bool {
	w.mu.Lock()
	w.closing = true
	w.mu.Unlock()
	w.signal()
	timer := time.NewTimer(w.limits.closeWait)
	defer timer.Stop()
	select {
	case <-w.done:
	case <-timer.C:
		// The store does not keep up: what is left never reaches the log,
		// and its observers read the thread instead.
		w.cancel()
		<-w.done
	}
	w.cancel()
	return w.ended
}

// --- Observers -----------------------------------------------------------------

// aguiFollow is the Detached Run an observer follows: its log, and how to
// tell whether the run still runs.
type aguiFollow struct {
	// log is the key of the run's log.
	log string
	// thread is the thread's key, under which the run state is kept, and
	// invocationID names the run there.
	thread       string
	invocationID string
	threadID     string
	runID        string
	// local is closed once the run, in this process, let its thread go: its
	// log then holds all it ever will. Nil when the run may be elsewhere.
	local <-chan struct{}
}

// follow writes the run's log to the response from RUN_STARTED and follows
// it to the run's end, with heartbeats while the run is quiet. It reports
// whether the stream ended with the run's end; a client going away, or the
// fallback marker, ends it otherwise. A client going away detaches this
// observer only: the run goes on.
//
// When the observer cannot follow the run to its end it ends the stream with
// the fallback marker: the log stopped short or is gone, or the run is no
// longer running — its run state ended or vanished — and its end did not
// reach the log within a grace.
func (h *AGUIHandler) follow(c *gin.Context, f aguiFollow) (atEnd bool) {
	ctx := c.Request.Context()
	logs, states := h.getRunLogs(), h.getRunStateStore()
	limits, heartbeat := h.runLogLimits(), h.heartbeatInterval()
	// waitCtx also ends once a run of this process let its thread go, so
	// the observer reads what is left at once.
	waitCtx, stopWaiting := context.WithCancel(ctx)
	defer stopWaiting()
	if f.local != nil {
		go func() {
			select {
			case <-f.local:
				stopWaiting()
			case <-waitCtx.Done():
			}
		}()
	}
	o := aguiObserver{c: c, f: f}
	cursor := ""
	lastWrite := time.Now()
	nextCheck := lastWrite.Add(limits.check)
	// ending is when the run was first seen no longer running, its end not
	// read yet; why says how it was seen. readFailing is when the reads
	// started failing.
	var ending, readFailing time.Time
	why, localDone := "", false
	for {
		recs, err := logs.Read(ctx, f.log, cursor, aguiRunLogReadBatch)
		if ctx.Err() != nil {
			return false
		}
		switch {
		case errors.Is(err, runlog.ErrGone):
			o.fallBack(aguiFallbackExpired)
			return false
		case err != nil:
			if readFailing.IsZero() {
				readFailing = time.Now()
			}
			if time.Since(readFailing) > limits.endGrace {
				log.FromContext(ctx).Warn("agui run log unreadable", "log", f.log, "err", err)
				o.fallBack(aguiFallbackLost)
				return false
			}
		default:
			readFailing = time.Time{}
		}
		for _, rec := range recs {
			if rec.Kind == runlog.KindTruncated {
				o.fallBack(aguiFallbackTruncated)
				return false
			}
			if !o.write(rec.Frame) {
				return false
			}
			lastWrite, cursor = time.Now(), rec.ID
			if rec.Kind == runlog.KindEnd {
				return true
			}
		}
		if len(recs) == aguiRunLogReadBatch {
			continue
		}

		// Caught up with the log, and the run has not ended in it.
		if f.local != nil && waitCtx.Err() != nil {
			if localDone {
				// Read once more after the run let its thread go: its log
				// holds all it ever will.
				o.fallBack(aguiFallbackTruncated)
				return false
			}
			localDone = true
			continue
		}
		now := time.Now()
		if !ending.IsZero() && now.Sub(ending) >= limits.endGrace {
			o.fallBack(why)
			return false
		}
		wait := min(lastWrite.Add(heartbeat).Sub(now), nextCheck.Sub(now))
		if !ending.IsZero() {
			wait = min(wait, ending.Add(limits.endGrace).Sub(now))
		}
		if readFailing.IsZero() {
			logs.Wait(waitCtx, f.log, cursor, max(wait, time.Millisecond))
		} else {
			// The store may be down: pause before reading again.
			pause := time.NewTimer(max(min(wait, aguiRunLogReadRetry), time.Millisecond))
			select {
			case <-waitCtx.Done():
			case <-pause.C:
			}
			pause.Stop()
		}
		if ctx.Err() != nil {
			return false
		}
		now = time.Now()
		if !now.Before(lastWrite.Add(heartbeat)) {
			if !o.write(aguiHeartbeat) {
				return false
			}
			lastWrite = now
		}
		if !now.Before(nextCheck) && states != nil {
			nextCheck = now.Add(limits.check)
			st, held, err := states.Get(ctx, f.thread)
			switch {
			case err != nil:
				// Unknown for now; the next check tells.
			case !held || st.InvocationID != f.invocationID:
				if ending.IsZero() {
					ending, why = now, aguiFallbackLost
				}
			case st.Ended:
				if ending.IsZero() {
					ending, why = now, aguiFallbackTruncated
				}
			default:
				ending = time.Time{}
			}
		}
	}
}

// aguiObserver writes one observer's stream.
type aguiObserver struct {
	c *gin.Context
	f aguiFollow
	// started reports that the stream carries an event.
	started bool
}

func (o *aguiObserver) write(frame []byte) bool {
	if !writeAGUIFrame(o.c, frame) {
		return false
	}
	if !bytes.HasPrefix(frame, []byte(":")) {
		o.started = true
	}
	return true
}

// fallBack ends the stream with the fallback marker, which tells the client
// to read the thread instead. A stream that carries no event yet opens with
// RUN_STARTED first, so it stays a valid AG-UI stream.
func (o *aguiObserver) fallBack(reason string) {
	f := o.f
	log.FromContext(o.c.Request.Context()).Warn("agui observer cannot follow its run to its end; the client reads the thread instead",
		"thread_id", f.threadID, "run_id", f.runID, "invocation_id", f.invocationID, "reason", reason)
	enc := aguisse.NewSSEWriter()
	if !o.started {
		frame, err := aguiSSEFrame(enc, aguievents.NewRunStartedEvent(f.threadID, f.runID))
		if err != nil || !o.write(frame) {
			return
		}
	}
	frame, err := aguiSSEFrame(enc, aguievents.NewCustomEvent(aguiFallbackEvent,
		aguievents.WithValue(aguiFallbackValue{ThreadID: f.threadID, RunID: f.runID, Reason: reason})))
	if err != nil {
		return
	}
	o.write(frame)
}

// writeAGUIFrame writes one SSE frame and flushes it; false means the client
// is gone.
func writeAGUIFrame(c *gin.Context, frame []byte) bool {
	if _, err := c.Writer.Write(frame); err != nil {
		return false
	}
	c.Writer.Flush()
	return true
}

// aguiStreamHeaders opens an SSE response, and sends its headers at once: a
// client learns that the stream opened before its first event.
func aguiStreamHeaders(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)
	c.Writer.Flush()
}

// --- Attach --------------------------------------------------------------------

// AttachRun handles GET /api/agui/:agent_id/threads/:thread_id/run: it streams
// the thread's Detached Run from its log, on any Pod. It replays the run from
// RUN_STARTED, under its runId, then follows it to its end, with heartbeats;
// the stream is the one the run's POST response carries. It ends with the
// fallback marker when it cannot follow the run that far.
//
// It answers 204 when there is no log to follow: no Detached Run is in
// flight or recently ended, the thread's run did not opt in, or the caller
// does not hold the thread, which reveals nothing about whose it is. The
// checks are those of the thread reads.
func (h *AGUIHandler) AttachRun(c *gin.Context) {
	threadID, thread, ok := h.boundThread(c)
	if !ok {
		return
	}
	states, logs := h.getRunStateStore(), h.getRunLogs()
	if thread == "" || states == nil || logs == nil {
		c.Status(http.StatusNoContent)
		return
	}
	ctx := c.Request.Context()
	st, held, err := states.Get(ctx, thread)
	if err != nil {
		log.FromContext(ctx).Error("agui attach: run state unavailable", "thread_id", threadID, "err", err)
		c.JSON(http.StatusServiceUnavailable, aguiErrorResponse{Error: "run state unavailable, retry later"})
		return
	}
	if !held || !st.Detached {
		c.Status(http.StatusNoContent)
		return
	}
	key := aguiRunLogKey(thread, st.InvocationID)
	if _, err := logs.Read(ctx, key, "", 1); err != nil {
		if errors.Is(err, runlog.ErrGone) {
			c.Status(http.StatusNoContent)
			return
		}
		log.FromContext(ctx).Error("agui attach: run log unavailable", "thread_id", threadID, "err", err)
		c.JSON(http.StatusServiceUnavailable, aguiErrorResponse{Error: "run log unavailable, retry later"})
		return
	}
	log.FromContext(ctx).Info("agui run attached",
		"thread_id", threadID, "run_id", st.RunID, "invocation_id", st.InvocationID, "ended", st.Ended)
	aguiStreamHeaders(c)
	h.follow(c, aguiFollow{log: key, thread: thread, invocationID: st.InvocationID, threadID: threadID, runID: st.RunID})
}
