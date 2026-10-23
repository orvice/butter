// Package eventcommit lets an asynchronous leaf wait for the session service
// to acknowledge its event, rather than mistaking an iterator's yield for a
// durable write. Composite agents may resume a leaf before forwarding its event.
package eventcommit

import (
	"context"
	"errors"
	"sync"

	"google.golang.org/adk/v2/session"
)

type key struct{}
type tracker struct {
	mu      sync.Mutex
	pending map[string]*receipt
}
type receipt struct {
	done chan struct{}
	err  error
}

// Track scopes commit receipts to one runner invocation. cancel releases leaves
// still waiting when the runner exits without persisting their events.
func Track(ctx context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	return context.WithValue(ctx, key{}, &tracker{pending: make(map[string]*receipt)}), cancel
}

// Available reports whether the caller is running under the commit contract.
func Available(ctx context.Context) bool {
	_, ok := ctx.Value(key{}).(*tracker)
	return ok
}

// Expect registers before yielding an event. The returned wait cleans up its
// registration whether persistence succeeds, fails, or the run is cancelled.
func Expect(ctx context.Context, id string) (wait func(context.Context) error, discard func()) {
	t, ok := ctx.Value(key{}).(*tracker)
	if !ok {
		return func(context.Context) error { return errors.New("event commit tracking is unavailable") }, func() {}
	}
	r := &receipt{done: make(chan struct{})}
	t.mu.Lock()
	t.pending[id] = r
	t.mu.Unlock()
	discard = func() {
		t.mu.Lock()
		delete(t.pending, id)
		t.mu.Unlock()
	}
	return func(waitCtx context.Context) error {
		defer discard()
		select {
		case <-r.done:
			return r.err
		case <-waitCtx.Done():
			return waitCtx.Err()
		case <-ctx.Done():
			return ctx.Err()
		}
	}, discard
}

// Service wraps the actual persistence seam; it never acknowledges an event
// before both the underlying event and state writes have returned successfully.
func Service(s session.Service) session.Service { return &service{Service: s} }

type service struct{ session.Service }

func (s *service) AppendEvent(ctx context.Context, sess session.Session, event *session.Event) error {
	err := s.Service.AppendEvent(ctx, sess, event)
	if t, ok := ctx.Value(key{}).(*tracker); ok {
		t.mu.Lock()
		if r, found := t.pending[event.ID]; found {
			r.err = err
			delete(t.pending, event.ID)
			close(r.done)
		}
		t.mu.Unlock()
	}
	return err
}
