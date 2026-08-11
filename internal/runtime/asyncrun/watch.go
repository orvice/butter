package asyncrun

import (
	"sync"

	"google.golang.org/adk/v2/session"
	"google.golang.org/protobuf/proto"

	"go.orx.me/apps/butter/internal/runtime/streamorch"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// FrameKind discriminates the observer frame variants.
type FrameKind int

const (
	// FrameState carries an authoritative Invocation snapshot (RUNNING
	// transition and the single terminal state).
	FrameState FrameKind = iota + 1
	// FrameRunEvent carries one ADK session event from the live run.
	FrameRunEvent
	// FrameTextDelta carries one partial assistant-text chunk.
	FrameTextDelta
)

// Frame is one observer fan-out message for a watched invocation.
type Frame struct {
	Kind     FrameKind
	Identity streamorch.RunIdentity
	// Text is set for FrameTextDelta.
	Text string
	// Event is set for FrameRunEvent.
	Event *session.Event
	// Invocation is set for FrameState; always a defensive clone.
	Invocation *agentsv1.Invocation
}

// Terminal reports whether this is a state frame carrying a terminal status.
func (f Frame) Terminal() bool {
	return f.Kind == FrameState && f.Invocation != nil && IsTerminal(f.Invocation.GetStatus())
}

// IsTerminal reports whether the status is a final invocation state.
func IsTerminal(status agentsv1.InvocationStatus) bool {
	switch status {
	case agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED,
		agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED,
		agentsv1.InvocationStatus_INVOCATION_STATUS_CANCELLED:
		return true
	}
	return false
}

// observerBuffer is each subscriber's channel capacity. A subscriber whose
// buffer is full when a frame is published is considered lagged: it is
// dropped and its channel is closed so the runner is never backpressured.
const observerBuffer = 256

// watchHub fans live frames out to per-invocation observers. Publishing
// never blocks: slow observers are disconnected rather than slowing the run.
type watchHub struct {
	mu     sync.Mutex
	subs   map[string]map[int]chan Frame // invocationID → subID → channel
	nextID int
}

func newWatchHub() *watchHub {
	return &watchHub{subs: make(map[string]map[int]chan Frame)}
}

// subscribe registers an observer for the invocation. The returned cancel is
// idempotent and safe to call after the hub has already dropped or closed
// the subscription. A closed channel without a preceding terminal state
// frame means the observer lagged and must reconcile from persisted state.
func (h *watchHub) subscribe(invocationID string) (<-chan Frame, func()) {
	ch := make(chan Frame, observerBuffer)
	h.mu.Lock()
	h.nextID++
	id := h.nextID
	if h.subs[invocationID] == nil {
		h.subs[invocationID] = make(map[int]chan Frame)
	}
	h.subs[invocationID][id] = ch
	h.mu.Unlock()

	cancel := func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		if m, ok := h.subs[invocationID]; ok {
			if _, live := m[id]; live {
				delete(m, id)
				close(ch)
				if len(m) == 0 {
					delete(h.subs, invocationID)
				}
			}
		}
	}
	return ch, cancel
}

// publish delivers the frame to every observer of the invocation without
// blocking. Observers whose buffers are full are dropped and closed.
func (h *watchHub) publish(invocationID string, f Frame) {
	h.mu.Lock()
	defer h.mu.Unlock()
	m := h.subs[invocationID]
	for id, ch := range m {
		select {
		case ch <- f:
		default:
			delete(m, id)
			close(ch)
		}
	}
	if len(m) == 0 {
		delete(h.subs, invocationID)
	}
}

// closeAll closes every remaining observer channel for the invocation. Used
// after the terminal state frame so watchers observe frame-then-close.
func (h *watchHub) closeAll(invocationID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, ch := range h.subs[invocationID] {
		close(ch)
	}
	delete(h.subs, invocationID)
}

// Watch subscribes an observer to the invocation's live frames. Watching is
// passive: it never starts, cancels, or slows execution. The cancel function
// must be called when the observer detaches.
func (c *Coordinator) Watch(invocationID string) (<-chan Frame, func()) {
	return c.hub.subscribe(invocationID)
}

// publishState publishes an authoritative snapshot of inv to observers.
func (c *Coordinator) publishState(inv *agentsv1.Invocation) {
	c.hub.publish(inv.GetId(), Frame{
		Kind:       FrameState,
		Invocation: proto.Clone(inv).(*agentsv1.Invocation),
	})
}

// hubSink adapts streamorch frames to hub publishes. It also captures the
// final response text so the coordinator can persist it. Frame publishing is
// non-blocking by construction (see watchHub.publish), so the runner's event
// loop is never delayed by observers.
type hubSink struct {
	hub          *watchHub
	invocationID string
	response     string
}

func (s *hubSink) Started(streamorch.RunIdentity) error { return nil }

func (s *hubSink) TextDelta(id streamorch.RunIdentity, text string) error {
	s.hub.publish(s.invocationID, Frame{Kind: FrameTextDelta, Identity: id, Text: text})
	return nil
}

func (s *hubSink) RunEvent(id streamorch.RunIdentity, evt *session.Event) error {
	s.hub.publish(s.invocationID, Frame{Kind: FrameRunEvent, Identity: id, Event: evt})
	return nil
}

func (s *hubSink) Final(_ streamorch.RunIdentity, response string) error {
	s.response = response
	return nil
}
