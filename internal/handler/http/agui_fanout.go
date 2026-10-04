package http

import (
	"bytes"
	"context"
	"sync"
	"time"

	"butterfly.orx.me/core/log"
	aguievents "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/events"
	aguisse "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/encoding/sse"
	"github.com/gin-gonic/gin"
)

// aguiHeartbeatInterval is how long an observer's stream may stay quiet
// before it sends a heartbeat, so proxies keep the connection open while the
// run thinks or waits on a tool.
const aguiHeartbeatInterval = 15 * time.Second

// aguiHeartbeat is an SSE comment. AG-UI clients skip frames without data.
var aguiHeartbeat = []byte(": heartbeat\n\n")

// aguiObserverMaxQueue bounds the bytes waiting for one observer. An observer
// that falls this far behind is cut off rather than buffered without end; the
// run goes on.
const aguiObserverMaxQueue = 16 << 20 // 16 MiB

// aguiFanout hands one Detached Run's AG-UI events to the observers following
// it in this process (ADR-0016 decision 5). The POST response is the first
// observer; a disconnect detaches only that observer, never the run.
//
// emit is the run's aguiEmitter. It never waits on an observer and never
// fails because of one: each observer has its own queue, drained by the
// goroutine serving it. The Run Log (#404) replaces the fan-out behind the
// same emitter, adding replay for observers that arrive later.
type aguiFanout struct {
	writer *aguisse.SSEWriter

	mu        sync.Mutex
	observers map[*aguiObserver]struct{}
	closed    bool
}

func newAGUIFanout() *aguiFanout {
	return &aguiFanout{writer: aguisse.NewSSEWriter(), observers: map[*aguiObserver]struct{}{}}
}

// emit encodes ev once, as an SSE frame, and queues the frame for every
// observer.
func (f *aguiFanout) emit(ev aguievents.Event) error {
	var frame bytes.Buffer
	// The encoder refuses a cancelled context, and a run's terminal event is
	// often sent after its context ended.
	if err := f.writer.WriteEvent(context.Background(), &frame, ev); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	for o := range f.observers {
		o.push(frame.Bytes())
	}
	return nil
}

// attach adds an observer that receives every event emitted from now on.
func (f *aguiFanout) attach() *aguiObserver {
	o := &aguiObserver{fanout: f, wake: make(chan struct{}, 1)}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		o.ended = true
		return o
	}
	f.observers[o] = struct{}{}
	return o
}

// close ends the run's stream: every observer ends once it has written what
// is queued for it.
func (f *aguiFanout) close() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	for o := range f.observers {
		o.end()
	}
	clear(f.observers)
}

func (f *aguiFanout) detach(o *aguiObserver) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.observers, o)
}

// aguiObserver is one follower of a run: a queue of encoded frames and a
// wake-up for the goroutine writing them out.
type aguiObserver struct {
	fanout *aguiFanout
	wake   chan struct{}

	mu     sync.Mutex
	frames [][]byte
	queued int
	// ended means no frame follows the queued ones.
	ended bool
	// cutOff means the observer fell too far behind and was dropped.
	cutOff bool
}

func (o *aguiObserver) push(frame []byte) {
	o.mu.Lock()
	switch {
	case o.cutOff:
	case o.queued+len(frame) > aguiObserverMaxQueue:
		o.cutOff = true
		o.frames, o.queued = nil, 0
	default:
		o.frames = append(o.frames, frame)
		o.queued += len(frame)
	}
	o.mu.Unlock()
	o.signal()
}

func (o *aguiObserver) end() {
	o.mu.Lock()
	o.ended = true
	o.mu.Unlock()
	o.signal()
}

func (o *aguiObserver) signal() {
	select {
	case o.wake <- struct{}{}:
	default:
	}
}

// take returns the frames queued so far, whether the stream ends after them,
// and whether the observer was cut off.
func (o *aguiObserver) take() (frames [][]byte, ended, cutOff bool) {
	o.mu.Lock()
	defer o.mu.Unlock()
	frames, o.frames, o.queued = o.frames, nil, 0
	return frames, o.ended, o.cutOff
}

func (o *aguiObserver) detach() { o.fanout.detach(o) }

// observe writes a Detached Run's events to the response as they come, with
// heartbeats while the run is quiet, until the run's stream ends or the
// client goes away. A client going away detaches this observer only: the run
// goes on.
func (h *AGUIHandler) observe(c *gin.Context, o *aguiObserver) {
	defer o.detach()
	ctx := c.Request.Context()
	interval := h.heartbeatInterval()
	idle := time.NewTimer(interval)
	defer idle.Stop()
	for {
		frames, ended, cutOff := o.take()
		if cutOff {
			log.FromContext(ctx).Warn("agui observer fell behind its run and was cut off; the run goes on")
			return
		}
		for _, frame := range frames {
			if !writeAGUIFrame(c, frame) {
				return
			}
		}
		if ended {
			return
		}
		if len(frames) > 0 {
			idle.Reset(interval)
		}
		select {
		case <-ctx.Done():
			return
		case <-o.wake:
		case <-idle.C:
			if !writeAGUIFrame(c, aguiHeartbeat) {
				return
			}
			idle.Reset(interval)
		}
	}
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
