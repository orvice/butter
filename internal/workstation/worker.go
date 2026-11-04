package workstation

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"sync"
	"time"

	"github.com/google/uuid"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const LeaseDuration = 2 * time.Minute
const cleanupAttempts = 3

// OrphanGrace exceeds a whole attempt/lease, so a GC candidate cannot
// later become a valid published snapshot. Normal predecessors skip this grace.
const OrphanGrace = 2 * LeaseDuration

// Attempt captures both configuration and publisher identity. Every success
// and error write rechecks these; ctx cancellation alone never authorizes it.
type Attempt struct {
	Record Record
	Scope  Scope
	Token  string
	Until  time.Time
}

func (s *Service) Claim(ctx context.Context, w, id, scopeID string) (Attempt, error) {
	var a Attempt
	_, err := s.mutate(ctx, w, id, func(r *Record, now time.Time) error {
		if r.Paused || r.Credential.Ciphertext == "" || r.RetryAfter.After(now) {
			return ErrPaused
		}
		for i := range r.Scopes {
			sc := &r.Scopes[i]
			if sc.ID != scopeID {
				continue
			}
			if sc.Hidden || sc.Due.After(now) || sc.LeaseUntil.After(now) {
				return ErrFenced
			}
			sc.Claim = uuid.NewString()
			sc.LeaseUntil = now.Add(LeaseDuration)
			sc.Health = "RUNNING"
			sc.SyncPhase = "SCANNING"
			a = Attempt{Record: *r, Scope: *sc, Token: sc.Claim, Until: sc.LeaseUntil}
			return nil
		}
		return ErrNotFound
	})
	return a, err
}
func (s *Service) finish(ctx context.Context, a Attempt, result Scan, scanErr error) error {
	for range 16 {
		old, err := s.Repo.Get(ctx, a.Record.WorkspaceID, a.Record.ID)
		if err != nil {
			return err
		}
		now, err := s.Repo.Now(ctx)
		if err != nil {
			return err
		}
		if old.Epoch != a.Record.Epoch || old.Paused || !now.Before(a.Until) {
			return ErrFenced
		}
		next := old
		next.Scopes = append([]Scope(nil), old.Scopes...)
		found := false
		previousSnapshot := ""
		for i := range next.Scopes {
			sc := &next.Scopes[i]
			if sc.ID != a.Scope.ID {
				continue
			}
			if sc.Generation != a.Scope.Generation || sc.Claim != a.Token {
				return ErrFenced
			}
			found = true
			sc.Claim = ""
			sc.LeaseUntil = time.Time{}
			sc.Due = maxTime(now.Add(5*time.Minute+jitter(old.ID+sc.ID)), old.RetryAfter)
			if scanErr == nil {
				previousSnapshot = sc.Snapshot
				sc.Snapshot = a.Token
				sc.Health = "HEALTHY"
				sc.Coverage = result.Coverage
				sc.Unresolved = result.Unresolved
				sc.Scanned = uint32(len(result.Items))
				sc.SyncPhase = "READY"
				sc.LastSuccess = now
				sc.Hidden = false
			} else {
				f := Classify(scanErr)
				sc.Health = f.Category
				sc.SyncPhase = "ERROR"
				if f.Category == "RATE_LIMITED" {
					sc.SyncPhase = "WAITING_RETRY"
				}
				switch f.Category {
				case "AUTH_INVALID", "CONFIG_ERROR":
					next.Paused = true
					next.Epoch = uuid.NewString()
					for j := range next.Scopes {
						next.Scopes[j].Health = f.Category
						next.Scopes[j].Hidden = true
					}
				case "ACCESS_DENIED", "FEATURE_UNAVAILABLE":
					sc.Hidden = true
				case "RATE_LIMITED":
					delay := max(f.RetryAfter, time.Minute)
					next.RetryAfter = now.Add(delay)
					for j := range next.Scopes {
						next.Scopes[j].Due = maxTime(next.Scopes[j].Due, next.RetryAfter)
					}
				}
			}
			break
		}
		if !found {
			return ErrFenced
		}
		next.Version++
		ok, err := s.Repo.CAS(ctx, old, &next, &Fence{Token: a.Token, Until: a.Until})
		if err != nil {
			return err
		}
		if ok {
			if scanErr == nil && previousSnapshot != "" {
				if err := s.Repo.PruneSnapshot(ctx, old.WorkspaceID, old.ID, a.Scope.ID, previousSnapshot); err != nil {
					return fmt.Errorf("%w: %w", ErrCleanupPending, err)
				}
			}
			return nil
		}
	}
	return ErrConflict
}
func jitter(id string) time.Duration {
	h := fnv.New32a()
	_, _ = h.Write([]byte(id))
	return time.Duration(h.Sum32()%30) * time.Second
}

// Publish stages the entire completed traversal and then atomically replaces
// one pointer. A staging failure never changes the currently visible snapshot.
func (s *Service) Publish(ctx context.Context, a Attempt, result Scan) error {
	if result.Coverage != "COMPLETE" && result.Coverage != "PARTIAL" {
		return ErrInvalid
	}
	if result.Unresolved > 0 && result.Coverage != "PARTIAL" {
		return ErrInvalid
	}
	now, err := s.Repo.Now(ctx)
	if err != nil {
		return err
	}
	if !now.Before(a.Until) {
		return ErrFenced
	}
	observations := []Observation{}
	seen := map[string]bool{}
	for _, sourceItem := range result.Items {
		if sourceItem == nil {
			return ErrInvalid
		}
		item := proto.Clone(sourceItem).(*v1.WorkStationItem)
		if item.SourceUpdatedAt == nil || item.RemoteId == "" || item.ScopeId != a.Scope.ID {
			return ErrInvalid
		}
		if item.Terminal && item.EndedAt == nil {
			result.Coverage = "PARTIAL"
			result.Unresolved++
			continue
		}
		item.Provider = a.Record.SourceIdentity().Provider
		item.Id = ItemID(a.Record.WorkspaceID, a.Record.SourceIdentity(), a.Scope.ID, item.Kind, item.RemoteId)
		if seen[item.Id] {
			return ErrInvalid
		}
		seen[item.Id] = true
		item.ObservedAt = timestamppb.New(now)
		item.Sources = nil
		observations = append(observations, Observation{WorkspaceID: a.Record.WorkspaceID, ConnectionID: a.Record.ID, ScopeID: a.Scope.ID, Snapshot: a.Token, Item: item})
	}
	if err := s.Repo.Stage(ctx, observations); err != nil {
		return err
	}
	return s.finish(ctx, a, result, nil)
}
func (s *Service) Fail(ctx context.Context, a Attempt, err error) error {
	return s.finish(ctx, a, Scan{}, err)
}
func (s *Service) syncScope(ctx context.Context, r Record, sc Scope) error {
	a, err := s.Claim(ctx, r.WorkspaceID, r.ID, sc.ID)
	if errors.Is(err, ErrFenced) || errors.Is(err, ErrPaused) || errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	// The whole attempt, including staging, is bounded below the lease TTL.
	attemptCtx, cancel := context.WithTimeout(ctx, 100*time.Second)
	defer cancel()
	token, err := s.token(attemptCtx, a.Record)
	if err == nil {
		now, e := s.Repo.Now(attemptCtx)
		err = e
		if err == nil && s.Budget != nil {
			var retry time.Duration
			retry, err = s.Budget.Reserve(attemptCtx, BudgetKey(a.Record.SourceIdentity()), now)
			if err == nil && retry > 0 {
				err = &Failure{Category: "RATE_LIMITED", RetryAfter: retry}
			}
		}
		if err == nil {
			var result Scan
			result, err = s.Provider.Scan(attemptCtx, token, a.Scope, now)
			if err == nil {
				err = s.Publish(attemptCtx, a, result)
				if errors.Is(err, ErrCleanupPending) {
					// The pointer is committed; never record a scan failure.
					return err
				}
				if err == nil {
					return nil
				}
			}
		}
	}
	// Use the parent context for recording a scan timeout; the fence still wins
	// over every late error, config change, revocation and lease takeover.
	return s.Fail(ctx, a, err)
}

// Cleanup retries idempotent orphan collection so a transient storage error
// does not postpone cleanup until the next lifecycle pass. A bounded retry
// keeps shutdown responsive and leaves the next scheduled pass as a fallback.
func (s *Service) Cleanup(ctx context.Context) error {
	var err error
	for attempt := 0; attempt < cleanupAttempts; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err = s.Repo.Collect(ctx)
		if err == nil {
			return nil
		}
	}
	return err
}

// are isolated and bounded to four in-flight scans per process.
func (s *Service) RunOnce(ctx context.Context) error {
	records, err := s.Repo.Due(ctx)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	sem := make(chan struct{}, 4)
	var mu sync.Mutex
	var errs []error
	for _, r := range records {
		for _, sc := range r.Scopes {
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				wg.Wait()
				return ctx.Err()
			}
			wg.Add(1)
			go func(r Record, sc Scope) {
				defer wg.Done()
				defer func() { <-sem }()
				if err := s.syncScope(ctx, r, sc); err != nil && !errors.Is(err, ErrFenced) && !errors.Is(err, ErrNotFound) {
					mu.Lock()
					errs = append(errs, err)
					mu.Unlock()
				}
			}(r, sc)
		}
	}
	wg.Wait()
	return errors.Join(errs...)
}

// Run belongs to the application lifecycle. Intents/claims remain in Mongo
// across restarts; stopping this goroutine loses no accepted refresh request.
func (s *Service) Run(ctx context.Context, report func(error)) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	nextGC := time.Now()
	for {
		if err := s.RunOnce(ctx); err != nil && ctx.Err() == nil {
			report(&RuntimeError{Operation: "sync", Cause: err})
		}
		if time.Now().After(nextGC) {
			gcCtx, cancelGC := context.WithTimeout(ctx, 30*time.Second)
			if err := s.Cleanup(gcCtx); err != nil && ctx.Err() == nil {
				report(&RuntimeError{Operation: "observation_gc", Cause: err})
			}
			cancelGC()
			nextGC = time.Now().Add(time.Minute)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
