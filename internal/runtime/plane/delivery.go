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
	DeliveryClaimer
	DeliveryCreator
	Get(context.Context, string) (*DeliveryAttempt, error)
	Put(context.Context, *DeliveryAttempt) error
	ListRecoverable(context.Context) ([]*DeliveryAttempt, error)
	// CompareAndSwap changes a record only while its observed status,
	// attempt count and claim token still match.
	CompareAndSwap(context.Context, *DeliveryAttempt, *DeliveryAttempt) error
}

// DeliveryClaimer fences delivery attempts across processes.
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
	if row != nil && row.Status == DeliveryPermanentFailure {
		p.mu.Unlock()
		return ErrPermanentDelivery
	}
	if row != nil && row.Status == DeliveryPending {
		if row.LeaseUntil != nil && row.LeaseUntil.After(time.Now().UTC()) {
			p.mu.Unlock()
			return ErrDeliveryInProgress
		}
		before := *row
		row.Status = DeliveryUnknown
		row.LastError = "delivery lease expired; result is ambiguous; owner confirmation required"
		row.UpdatedAt = time.Now().UTC()
		row.ClaimToken = ""
		row.LeaseUntil = nil
		err := p.store.CompareAndSwap(ctx, &before, row)
		p.mu.Unlock()
		if err != nil {
			return err
		}
		return ErrPermanentDelivery
	}
	if row != nil && row.Attempts >= maxAutomaticDeliveryAttempts && row.Status == DeliveryRetryable {
		before := *row
		row.Status = DeliveryUnknown
		row.LastError = "automatic retry budget exhausted; owner confirmation required"
		err := p.store.CompareAndSwap(ctx, &before, row)
		p.mu.Unlock()
		if err != nil {
			return err
		}
		return ErrPermanentDelivery
	}
	if row == nil {
		row = &DeliveryAttempt{Key: key, Event: e, Reply: reply, WorkspaceID: e.WorkspaceID, AppID: e.AppID, InstallationID: e.InstallationID}
		row.Status = DeliveryRetryable
		if err = p.store.Create(ctx, row); err != nil && !errors.Is(err, ErrDeliveryAlreadyExists) {
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
	row, err = p.store.Claim(ctx, key, token, time.Now().UTC(), deliveryLeaseDuration)
	if err != nil {
		p.mu.Unlock()
		return err
	}
	p.mu.Unlock()
	err = p.target.PublishReply(ctx, row.Event, row.Reply)
	p.mu.Lock()
	defer p.mu.Unlock()
	row.UpdatedAt = time.Now().UTC()
	if err == nil {
		row.Status = DeliveryDelivered
		row.LastError = ""
		return p.putOwned(ctx, row, token)
	}
	row.LastError = safeDeliveryError(err)
	var classifier DeliveryClassifier
	var retryable RetryableError
	if errors.As(err, &classifier) {
		switch classifier.DeliveryStatus() {
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
	} else if errors.As(err, &retryable) && retryable.Retryable() {
		if row.Attempts >= maxAutomaticDeliveryAttempts {
			row.Status = DeliveryUnknown
		} else {
			row.Status = DeliveryRetryable
		}
	} else if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		row.Status = DeliveryUnknown
	} else {
		row.Status = DeliveryPermanentFailure
	}
	if saveErr := p.putOwned(ctx, row, token); saveErr != nil {
		return errors.Join(fmt.Errorf("publish reply: %w", err), saveErr)
	}
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
	if row.Status == DeliveryPending {
		return ErrDeliveryInProgress
	}
	before := *row
	row.OwnerConfirmed = true
	row.Status = DeliveryRetryable
	row.LastError = ""
	row.UpdatedAt = time.Now().UTC()
	if err := p.store.CompareAndSwap(ctx, &before, row); err != nil {
		return err
	}
	return p.publishLocked(ctx, row)
}

func (p *DurablePublisher) publishLocked(ctx context.Context, row *DeliveryAttempt) error {
	token := deliveryClaimToken()
	claimed, err := p.store.Claim(ctx, row.Key, token, time.Now().UTC(), deliveryLeaseDuration)
	if err != nil {
		return err
	}
	row = claimed
	err = p.target.PublishReply(ctx, row.Event, row.Reply)
	row.UpdatedAt = time.Now().UTC()
	if err == nil {
		row.Status = DeliveryDelivered
		row.LastError = ""
	} else {
		row.LastError = safeDeliveryError(err)
		row.Status = DeliveryUnknown
	}
	if saveErr := p.putOwned(ctx, row, token); saveErr != nil {
		return errors.Join(err, saveErr)
	}
	return err
}

func (p *DurablePublisher) putOwned(ctx context.Context, row *DeliveryAttempt, token string) error {
	saveCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	clean := *row
	clean.ClaimToken = ""
	clean.LeaseUntil = nil
	return p.store.PutOwned(saveCtx, &clean, token)
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
	return cloneDelivery(r), nil
}
func (s *MemoryDeliveryStore) Put(_ context.Context, r *DeliveryAttempt) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rows[r.Key] = cloneDelivery(r)
	return nil
}
func (s *MemoryDeliveryStore) Create(_ context.Context, r *DeliveryAttempt) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.rows[r.Key]; ok {
		return ErrDeliveryAlreadyExists
	}
	s.rows[r.Key] = cloneDelivery(r)
	return nil
}
func (s *MemoryDeliveryStore) Claim(_ context.Context, key, token string, now time.Time, lease time.Duration) (*DeliveryAttempt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.rows[key]
	if !ok {
		return nil, ErrDeliveryNotFound
	}
	if r.Status != DeliveryRetryable {
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
	c.ClaimedAt = &now
	s.rows[key] = &c
	return cloneDelivery(&c), nil
}
func (s *MemoryDeliveryStore) PutOwned(_ context.Context, r *DeliveryAttempt, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.rows[r.Key]
	if !ok || current.ClaimToken != token {
		return ErrDeliveryLeaseLost
	}
	s.rows[r.Key] = cloneDelivery(r)
	return nil
}
func (s *MemoryDeliveryStore) CompareAndSwap(_ context.Context, before, after *DeliveryAttempt) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.rows[before.Key]
	if !ok || current.Status != before.Status || current.Attempts != before.Attempts || current.ClaimToken != before.ClaimToken {
		return ErrDeliveryInProgress
	}
	s.rows[before.Key] = cloneDelivery(after)
	return nil
}

func cloneDelivery(r *DeliveryAttempt) *DeliveryAttempt {
	c := *r
	if r.ClaimedAt != nil {
		v := *r.ClaimedAt
		c.ClaimedAt = &v
	}
	if r.LeaseUntil != nil {
		v := *r.LeaseUntil
		c.LeaseUntil = &v
	}
	if r.Event != nil {
		event := *r.Event
		event.Raw = append([]byte(nil), r.Event.Raw...)
		c.Event = &event
	}
	return &c
}
func (s *MemoryDeliveryStore) ListAll(_ context.Context) ([]*DeliveryAttempt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*DeliveryAttempt, 0, len(s.rows))
	for _, r := range s.rows {
		out = append(out, cloneDelivery(r))
	}
	return out, nil
}
func (s *MemoryDeliveryStore) ListRecoverable(_ context.Context) ([]*DeliveryAttempt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []*DeliveryAttempt{}
	for _, r := range s.rows {
		if r.Status == DeliveryPending || r.Status == DeliveryRetryable {
			out = append(out, cloneDelivery(r))
		}
	}
	return out, nil
}
