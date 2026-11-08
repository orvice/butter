package plane

import (
	"context"
	"crypto/rand"
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
	DeliveryUnknown          DeliveryStatus = "unknown"
)

var ErrDeliveryAlreadyDone = errors.New("plane reply already delivered")
var ErrDeliveryNotOwner = errors.New("plane delivery owner confirmation required")
var ErrDeliveryInProgress = errors.New("plane reply delivery is already in progress")
var ErrDeliveryLeaseLost = errors.New("plane reply delivery lease lost")
var ErrPermanentDelivery = errors.New("plane reply delivery permanently failed")

const maxAutomaticDeliveryAttempts = 3
const deliveryLeaseDuration = 2 * time.Minute

type DeliveryAttempt struct {
	Key            string
	Event          *Event
	Reply          string
	Status         DeliveryStatus
	Attempts       int
	LastError      string
	UpdatedAt      time.Time
	ClaimedAt      *time.Time
	ClaimToken     string     `bson:"claim_token,omitempty"`
	LeaseUntil     *time.Time `bson:"lease_until,omitempty"`
	OwnerConfirmed bool
	WorkspaceID    string
	AppID          string
	InstallationID string
}
type DeliveryStore interface {
	Get(context.Context, string) (*DeliveryAttempt, error)
	Put(context.Context, *DeliveryAttempt) error
	ListRecoverable(context.Context) ([]*DeliveryAttempt, error)
}

// DeliveryClaimer is implemented by durable stores that can fence delivery
// attempts across processes. The fallback remains useful for old/test stores,
// while Mongo uses an atomic find-and-update.
type DeliveryClaimer interface {
	Claim(context.Context, string, string, time.Time, time.Duration) (*DeliveryAttempt, error)
	PutOwned(context.Context, *DeliveryAttempt, string) error
}
type DeliveryCreator interface {
	Create(context.Context, *DeliveryAttempt) error
}
type DeliveryAuditStore interface {
	ListAll(context.Context) ([]*DeliveryAttempt, error)
}
type RetryableError interface{ Retryable() bool }
type DeliveryClassifier interface{ DeliveryStatus() DeliveryStatus }
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
	if row != nil && row.Status == DeliveryUnknown {
		p.mu.Unlock()
		return ErrPermanentDelivery
	}
	if row != nil && row.Status == DeliveryPending {
		if row.LeaseUntil != nil && row.LeaseUntil.After(time.Now().UTC()) {
			p.mu.Unlock()
			return ErrDeliveryInProgress
		}
		row.Status = DeliveryUnknown
		row.LastError = "delivery lease expired; result is ambiguous; owner confirmation required"
		row.UpdatedAt = time.Now().UTC()
		_ = p.store.Put(ctx, row)
		p.mu.Unlock()
		return ErrPermanentDelivery
	}
	if row != nil && row.Attempts >= maxAutomaticDeliveryAttempts && row.Status == DeliveryRetryable {
		row.Status = DeliveryUnknown
		row.LastError = "automatic retry budget exhausted; owner confirmation required"
		_ = p.store.Put(ctx, row)
		p.mu.Unlock()
		return ErrPermanentDelivery
	}
	if row == nil {
		row = &DeliveryAttempt{Key: key, Event: e, Reply: reply, WorkspaceID: e.WorkspaceID, AppID: e.AppID, InstallationID: e.InstallationID}
		row.Status = DeliveryRetryable
		if creator, ok := p.store.(DeliveryCreator); ok {
			if err = creator.Create(ctx, row); err != nil && !errors.Is(err, ErrDeliveryAlreadyExists) {
				p.mu.Unlock()
				return err
			}
			if err != nil {
				row, err = p.store.Get(ctx, key)
				if err != nil {
					p.mu.Unlock()
					return err
				}
			}
		} else if err = p.store.Put(ctx, row); err != nil {
			p.mu.Unlock()
			return err
		}
	}
	if row.Event == nil {
		row.Event = e
	}
	if row.WorkspaceID == "" {
		row.WorkspaceID = e.WorkspaceID
	}
	if row.AppID == "" {
		row.AppID = e.AppID
	}
	if row.InstallationID == "" {
		row.InstallationID = e.InstallationID
	}
	if row.Reply == "" {
		row.Reply = reply
	}

	token := deliveryClaimToken()
	if claimer, ok := p.store.(DeliveryClaimer); ok {
		row, err = claimer.Claim(ctx, key, token, time.Now().UTC(), deliveryLeaseDuration)
		if err != nil {
			p.mu.Unlock()
			return err
		}
	} else {
		row.Attempts++
		row.Status = DeliveryPending
		row.ClaimToken = token
		n := time.Now().UTC().Add(deliveryLeaseDuration)
		row.LeaseUntil = &n
		row.UpdatedAt = time.Now().UTC()
		if err = p.store.Put(ctx, row); err != nil {
			p.mu.Unlock()
			return err
		}
	}
	p.mu.Unlock()
	err = p.target.PublishReply(ctx, e, reply)
	p.mu.Lock()
	defer p.mu.Unlock()
	row.UpdatedAt = time.Now().UTC()
	if err == nil {
		row.Status = DeliveryDelivered
		row.LastError = ""
		return p.putOwned(ctx, row, token)
	}
	row.LastError = safeDeliveryError(err)
	if r, ok := err.(DeliveryClassifier); ok {
		switch r.DeliveryStatus() {
		case DeliveryRetryable:
			if row.Attempts >= maxAutomaticDeliveryAttempts {
				row.Status = DeliveryUnknown
			} else {
				row.Status = DeliveryRetryable
			}
		case DeliveryUnknown:
			row.Status = DeliveryUnknown
		default:
			row.Status = DeliveryPermanentFailure
		}
	} else if r, ok := err.(RetryableError); ok && r.Retryable() {
		if row.Attempts >= maxAutomaticDeliveryAttempts {
			row.Status = DeliveryUnknown
		} else {
			row.Status = DeliveryRetryable
		}
	} else if errors.Is(err, context.DeadlineExceeded) {
		row.Status = DeliveryUnknown
	} else {
		row.Status = DeliveryPermanentFailure
	}
	_ = p.putOwned(ctx, row, token)
	return fmt.Errorf("publish reply: %w", err)
}
func (p *DurablePublisher) ConfirmAndRetry(ctx context.Context, key, owner string) error {
	if p == nil || p.target == nil || p.store == nil {
		return ErrReplyUnavailable
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	row, err := p.store.Get(ctx, key)
	if err != nil {
		return err
	}
	if row.Event == nil || row.Event.BotUserID == "" || row.Event.BotUserID != owner {
		return ErrDeliveryNotOwner
	}
	if row.Status == DeliveryDelivered {
		return ErrDeliveryAlreadyDone
	}
	row.OwnerConfirmed = true
	row.Status = DeliveryRetryable
	row.LastError = ""
	row.UpdatedAt = time.Now().UTC()
	if err := p.store.Put(ctx, row); err != nil {
		return err
	}
	return p.publishLocked(ctx, row)
}

func (p *DurablePublisher) publishLocked(ctx context.Context, row *DeliveryAttempt) error {
	token := deliveryClaimToken()
	if claimer, ok := p.store.(DeliveryClaimer); ok {
		claimed, err := claimer.Claim(ctx, row.Key, token, time.Now().UTC(), deliveryLeaseDuration)
		if err != nil {
			return err
		}
		row = claimed
	} else {
		row.Attempts++
		row.Status = DeliveryPending
		row.ClaimToken = token
		n := time.Now().UTC().Add(deliveryLeaseDuration)
		row.LeaseUntil = &n
		row.UpdatedAt = time.Now().UTC()
		if err := p.store.Put(ctx, row); err != nil {
			return err
		}
	}
	err := p.target.PublishReply(ctx, row.Event, row.Reply)
	row.UpdatedAt = time.Now().UTC()
	if err == nil {
		row.Status = DeliveryDelivered
		row.LastError = ""
	} else {
		row.LastError = safeDeliveryError(err)
		row.Status = DeliveryUnknown
	}
	_ = p.putOwned(ctx, row, token)
	return err
}

func (p *DurablePublisher) putOwned(ctx context.Context, row *DeliveryAttempt, token string) error {
	clean := *row
	clean.ClaimToken = ""
	clean.LeaseUntil = nil
	if claimer, ok := p.store.(DeliveryClaimer); ok {
		return claimer.PutOwned(ctx, &clean, token)
	}
	return p.store.Put(ctx, &clean)
}

func deliveryClaimToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err == nil {
		return fmt.Sprintf("%x", b)
	}
	return fmt.Sprintf("%d", time.Now().UnixNano())
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
var ErrDeliveryAlreadyExists = errors.New("plane delivery already exists")

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
func (s *MemoryDeliveryStore) Create(_ context.Context, r *DeliveryAttempt) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.rows[r.Key]; ok {
		return ErrDeliveryAlreadyExists
	}
	c := *r
	s.rows[r.Key] = &c
	return nil
}
func (s *MemoryDeliveryStore) Claim(_ context.Context, key, token string, now time.Time, lease time.Duration) (*DeliveryAttempt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.rows[key]
	if !ok {
		return nil, ErrDeliveryNotFound
	}
	if r.Status == DeliveryDelivered || r.Status == DeliveryUnknown || r.Status == DeliveryPermanentFailure {
		return nil, ErrDeliveryInProgress
	}
	if r.LeaseUntil != nil && r.LeaseUntil.After(now) {
		return nil, ErrDeliveryInProgress
	}
	c := *r
	c.Attempts++
	c.Status = DeliveryPending
	c.ClaimToken = token
	n := now.Add(lease)
	c.LeaseUntil = &n
	c.UpdatedAt = now
	s.rows[key] = &c
	return &c, nil
}
func (s *MemoryDeliveryStore) PutOwned(_ context.Context, r *DeliveryAttempt, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.rows[r.Key]
	if !ok || current.ClaimToken != token {
		return ErrDeliveryLeaseLost
	}
	c := *r
	s.rows[r.Key] = &c
	return nil
}
func (s *MemoryDeliveryStore) ListAll(_ context.Context) ([]*DeliveryAttempt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*DeliveryAttempt, 0, len(s.rows))
	for _, r := range s.rows {
		c := *r
		out = append(out, &c)
	}
	return out, nil
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
