package plane

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type DeliveryStatus string

const (
	DeliveryPending          DeliveryStatus = "pending"
	DeliveryDelivered        DeliveryStatus = "delivered"
	DeliveryRetryable        DeliveryStatus = "retryable"
	DeliveryPermanentFailure DeliveryStatus = "permanent_failure"
)

var ErrDeliveryAlreadyDone = errors.New("plane reply already delivered")
var ErrPermanentDelivery = errors.New("plane reply delivery permanently failed")

type DeliveryAttempt struct {
	Key       string
	Event     *Event
	Reply     string
	Status    DeliveryStatus
	Attempts  int
	LastError string
	UpdatedAt time.Time
}
type DeliveryStore interface {
	Get(context.Context, string) (*DeliveryAttempt, error)
	Put(context.Context, *DeliveryAttempt) error
	ListRecoverable(context.Context) ([]*DeliveryAttempt, error)
}
type RetryableError interface{ Retryable() bool }
type DurablePublisher struct {
	target ReplyPublisher
	store  DeliveryStore
	mu     sync.Mutex
}

func NewDurablePublisher(target ReplyPublisher, store DeliveryStore) *DurablePublisher {
	return &DurablePublisher{target: target, store: store}
}
func DeliveryKey(e *Event) string {
	return e.InstallationID + ":" + e.AgentRunID + ":" + e.ActivityID + ":reply"
}
func (p *DurablePublisher) PublishReply(ctx context.Context, e *Event, reply string) error {
	if p == nil || p.target == nil || p.store == nil {
		return ErrReplyUnavailable
	}
	key := DeliveryKey(e)
	p.mu.Lock()
	row, err := p.store.Get(ctx, key)
	if err != nil && !errors.Is(err, ErrDeliveryNotFound) {
		p.mu.Unlock()
		return err
	}
	if row != nil && row.Status == DeliveryDelivered {
		p.mu.Unlock()
		return ErrDeliveryAlreadyDone
	}
	if row == nil {
		row = &DeliveryAttempt{Key: key, Event: e, Reply: reply}
	}
	row.Attempts++
	row.Status = DeliveryPending
	row.UpdatedAt = time.Now().UTC()
	if err = p.store.Put(ctx, row); err != nil {
		p.mu.Unlock()
		return err
	}
	p.mu.Unlock()
	err = p.target.PublishReply(ctx, e, reply)
	p.mu.Lock()
	defer p.mu.Unlock()
	row.UpdatedAt = time.Now().UTC()
	if err == nil {
		row.Status = DeliveryDelivered
		row.LastError = ""
		return p.store.Put(ctx, row)
	}
	row.LastError = safeDeliveryError(err)
	if r, ok := err.(RetryableError); ok && r.Retryable() {
		row.Status = DeliveryRetryable
	} else {
		row.Status = DeliveryPermanentFailure
	}
	_ = p.store.Put(ctx, row)
	return fmt.Errorf("publish reply: %w", err)
}
func (p *DurablePublisher) Recover(ctx context.Context) error {
	rows, err := p.store.ListRecoverable(ctx)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row.Status != DeliveryPending && row.Status != DeliveryRetryable {
			continue
		}
		if err := p.PublishReply(ctx, row.Event, row.Reply); err != nil && row.Status == DeliveryPermanentFailure {
			continue
		}
	}
	return nil
}
func safeDeliveryError(err error) string {
	if err == nil {
		return ""
	}
	s := err.Error()
	if len(s) > 256 {
		s = s[:256]
	}
	return s
}

var ErrDeliveryNotFound = errors.New("plane delivery not found")

type MemoryDeliveryStore struct {
	mu   sync.Mutex
	rows map[string]*DeliveryAttempt
}

func NewMemoryDeliveryStore() *MemoryDeliveryStore {
	return &MemoryDeliveryStore{rows: map[string]*DeliveryAttempt{}}
}
func (s *MemoryDeliveryStore) Get(_ context.Context, k string) (*DeliveryAttempt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.rows[k]
	if !ok {
		return nil, ErrDeliveryNotFound
	}
	c := *r
	return &c, nil
}
func (s *MemoryDeliveryStore) Put(_ context.Context, r *DeliveryAttempt) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	c := *r
	s.rows[r.Key] = &c
	return nil
}
func (s *MemoryDeliveryStore) ListRecoverable(_ context.Context) ([]*DeliveryAttempt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []*DeliveryAttempt{}
	for _, r := range s.rows {
		if r.Status == DeliveryPending || r.Status == DeliveryRetryable {
			c := *r
			out = append(out, &c)
		}
	}
	return out, nil
}
