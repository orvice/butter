package plane

import (
	"context"
	"errors"
	planeprocessing "go.orx.me/apps/butter/internal/repo/planeprocessing"
	"sync"
	"time"
)

var ErrRuntimeDependency = errors.New("plane runtime production dependency is missing")

type FencedTokenLease interface {
	AcquireToken(context.Context, string, string, int64, time.Time) (uint64, error)
	ReleaseToken(context.Context, string, string, uint64) error
}
type TokenLeaseAdapter struct{ Lease FencedTokenLease }

func (a TokenLeaseAdapter) Acquire(ctx context.Context, key string, revision int64, expiry time.Time) (uint64, error) {
	if a.Lease == nil {
		return 0, ErrRuntimeDependency
	}
	return a.Lease.AcquireToken(ctx, key, "", revision, expiry)
}

type PlaneRuntimeDependencies struct {
	Queue       *StreamQueue
	Resolver    PlaneAppResolver
	Runner      Runner
	Activity    ActivityClient
	Token       ActivityTokenSource
	Delivery    DeliveryStore
	Processing  planeprocessing.ClaimRepository
	Progress    ProgressSink
	Diagnostics DiagnosticStore
	TokenLease  FencedTokenLease
	Consumer    string
}
type PlaneRuntime struct {
	Client    ActivityClient
	Publisher *DurablePublisher
	Processor *Processor
	Worker    *Worker
}
type PlaneRuntimeFactory struct{}

func (PlaneRuntimeFactory) Build(d PlaneRuntimeDependencies) (*PlaneRuntime, error) {
	if d.Queue == nil || d.Resolver == nil || d.Runner == nil || d.Activity == nil || d.Token == nil || d.Delivery == nil || d.Processing == nil || d.Progress == nil || d.Diagnostics == nil || d.TokenLease == nil {
		return nil, ErrRuntimeDependency
	}
	p := NewProcessor(d.Runner, NewDurablePublisher(ActivityReplyPublisher{Client: d.Activity}, d.Delivery))
	p.SetProgressSink(d.Progress)
	p.SetDiagnosticStore(d.Diagnostics)
	p.SetClaimRepository(d.Processing, d.Consumer)
	w := NewWorker(d.Queue, p, d.Resolver, d.Consumer)
	return &PlaneRuntime{Client: d.Activity, Publisher: NewDurablePublisher(ActivityReplyPublisher{Client: d.Activity}, d.Delivery), Processor: p, Worker: w}, nil
}

type RuntimeCoordinator struct {
	mu      sync.Mutex
	runtime *PlaneRuntime
	started bool
	stopped bool
}

func NewRuntimeCoordinator(r *PlaneRuntime) *RuntimeCoordinator {
	return &RuntimeCoordinator{runtime: r}
}
func (c *RuntimeCoordinator) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.runtime == nil || c.runtime.Worker == nil {
		return ErrRuntimeDependency
	}
	if c.started && !c.stopped {
		return nil
	}
	if c.stopped {
		return errors.New("plane runtime coordinator is stopped")
	}
	if e := c.runtime.Worker.Start(ctx); e != nil {
		return e
	}
	c.started = true
	return nil
}
func (c *RuntimeCoordinator) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.stopped {
		return
	}
	if c.runtime != nil && c.runtime.Worker != nil {
		c.runtime.Worker.Stop()
	}
	c.stopped = true
}
