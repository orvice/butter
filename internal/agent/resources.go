package agent

import (
	"errors"
	"io"
	"sync"
)

// Resources owns the live connections an agent tree opened while it was built:
// today the MCP toolsets of every LLM agent in the tree, each holding
// streamable HTTP / SSE sessions.
//
// A constructor returns one Resources for the whole tree, so whoever owns the
// agent — the runner registry, a reload, a test — closes exactly one handle
// instead of walking the tree. Runs Acquire the resources they use; Close
// retires them and the last release closes the connections, so replacing an
// agent mid-run never pulls a connection out from under a turn that is still
// running. A nil Resources is usable and closes nothing.
type Resources struct {
	mu      sync.Mutex
	closers []io.Closer
	// refs counts the runs currently using the resources.
	refs int
	// retired is set by Close: no new run may Acquire afterwards.
	retired bool
	// closed is set once the closers ran; further Add/Close calls only report.
	closed bool
	err    error
}

// Add registers closers, which are closed when the resources Close. Closers
// added after retirement are closed at once so a late registration cannot
// leak.
func (r *Resources) Add(closers ...io.Closer) {
	if r == nil || len(closers) == 0 {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.retired {
		for _, c := range closers {
			r.err = errors.Join(r.err, c.Close())
		}
		return
	}
	r.closers = append(r.closers, closers...)
}

// Acquire marks a run as using these resources. The returned release must run
// when the run is done. ok is false once the resources are retired; the caller
// must then resolve the agent again instead of using the closed one.
func (r *Resources) Acquire() (release func(), ok bool) {
	if r == nil {
		return func() {}, true
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.retired {
		return nil, false
	}
	r.refs++
	return r.release, true
}

func (r *Resources) release() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.refs--
	if r.refs == 0 && r.retired {
		r.closeLocked()
	}
}

// Close retires the resources and closes their connections unless a run still
// holds them, in which case the last release does. It is safe to call more
// than once and returns the close errors observed so far.
func (r *Resources) Close() error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.retired = true
	if r.refs == 0 {
		r.closeLocked()
	}
	return r.err
}

// Err returns the errors closing produced. It is complete only after the last
// acquired run released.
func (r *Resources) Err() error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.err
}

func (r *Resources) closeLocked() {
	if r.closed {
		return
	}
	r.closed = true
	closers := r.closers
	r.closers = nil
	for _, c := range closers {
		r.err = errors.Join(r.err, c.Close())
	}
}
