package plane

import (
	"context"
	"errors"
	"io"
	"sync"
)

type CloseFunc func() error

func (f CloseFunc) Close() error { return f() }

// PlaneResourceOwner is the application boundary for Plane runtime resources.
// It owns lifecycle only; construction remains fail-closed in PlaneRuntimeFactory.
type PlaneResourceOwner struct {
	mu      sync.Mutex
	runtime *RuntimeCoordinator
	worker  *Worker
	closers []io.Closer
	started bool
	closed  bool
}

func NewPlaneResourceOwner(runtime *RuntimeCoordinator, closers ...io.Closer) *PlaneResourceOwner {
	return &PlaneResourceOwner{runtime: runtime, closers: closers}
}
func (o *PlaneResourceOwner) Start(ctx context.Context) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.closed {
		return errors.New("plane resources are closed")
	}
	if o.started {
		return nil
	}
	if o.runtime == nil {
		o.rollbackLocked()
		return ErrRuntimeDependency
	}
	if e := o.runtime.Start(ctx); e != nil {
		o.rollbackLocked()
		return e
	}
	o.worker = o.runtime.runtime.Worker
	o.started = true
	return nil
}
func (o *PlaneResourceOwner) rollbackLocked() {
	if o.runtime != nil {
		o.runtime.Stop()
	}
	for i := len(o.closers) - 1; i >= 0; i-- {
		_ = o.closers[i].Close()
	}
}
func (o *PlaneResourceOwner) Stop() {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.closed {
		return
	}
	if o.runtime != nil {
		o.runtime.Stop()
	}
	for i := len(o.closers) - 1; i >= 0; i-- {
		_ = o.closers[i].Close()
	}
	o.closed = true
}
