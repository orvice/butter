package workstation

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const cleanupAttempts = 3
const LeaseDuration = 2 * time.Minute

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
			if sc.Pending != nil && !sc.Pending.Started.After(now) && now.Sub(sc.Pending.Started) < 24*time.Hour {
				sc.ScanStarted = sc.Pending.Started
				sc.ScannedItems = sc.Pending.Count
			} else {
				sc.Pending = nil
				sc.ScanStarted = now
				sc.ScannedItems = 0
			}
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
		next.Scopes = CloneScopes(old.Scopes)
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
				sc.SyncPhase = "READY"
				sc.Scanned = uint32(len(result.Items))
				sc.CheckpointLag = false
				sc.CheckpointError = ""
				sc.Coverage = result.Coverage
				sc.Unresolved = result.Unresolved
				sc.LastSuccess = now
				sc.Pending = nil
				sc.ScannedItems = uint32(len(result.Items))
				if result.Coverage == "COMPLETE" {
					sc.Checkpoint = a.Scope.ScanStarted
					if result.Mode != "INCREMENTAL" {
						sc.LastFull = a.Scope.ScanStarted
					}
				}
				sc.Hidden = false
			} else {
				f := Classify(scanErr)
				sc.Health = f.Category
				sc.SyncPhase = "ERROR"
				if f.Category == "SCHEMA_ERROR" || f.Category == "UNSUPPORTED" || errors.Is(scanErr, ErrInvalid) {
					sc.Pending = nil
				}
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
			if scanErr != nil && Classify(scanErr).Category == "RATE_LIMITED" {
				delay := max(Classify(scanErr).RetryAfter, time.Minute)
				i, k := BudgetKeys(old)
				if err := s.Repo.DeferBudget(ctx, i, delay); err != nil {
					return err
				}
				if err := s.Repo.DeferBudget(ctx, k, delay); err != nil {
					return err
				}
			}
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
			retry, e := s.Budget.Reserve(attemptCtx, BudgetKey(a.Record.SourceIdentity()), now)
			err = e
			if err == nil && retry > 0 {
				err = &Failure{Category: "RATE_LIMITED", RetryAfter: retry}
			}
		}
		if err == nil {
			var result Scan
			resolved, e := s.resolve(attemptCtx, a.Record)
			if e != nil {
				err = e
			} else if incremental, ok := resolved.Adapter.(IncrementalProvider); ok && a.Scope.Pending == nil && a.Scope.Snapshot != "" && a.Scope.Coverage == "COMPLETE" && a.Scope.Unresolved == 0 && !a.Scope.Checkpoint.IsZero() && now.Sub(a.Scope.LastFull) < 24*time.Hour {
				previous, e := s.Repo.Query(attemptCtx, a.Record.WorkspaceID, Query{AllowedConnections: []string{a.Record.ID}, ConnectionID: a.Record.ID, ScopeID: a.Scope.ID, Status: "all", Now: now})
				if e != nil {
					err = e
				} else {
					result, err = incremental.ScanIncremental(attemptCtx, token, a.Scope, now, previous)
					if result.Mode == "" {
						result.Mode = "INCREMENTAL"
					}
				}
			} else {
				if batched, ok := resolved.Adapter.(BatchProvider); ok {
					result, err = s.scanBatches(attemptCtx, a, batched, token)
					if errors.Is(err, errScanPending) {
						return nil
					}
				} else {
					result, err = resolved.Adapter.Scan(attemptCtx, token, a.Scope, now)
				}
				result.Mode = "FULL"
			}
			if err == nil {
				err = s.Publish(attemptCtx, a, result)
				if errors.Is(err, ErrCleanupPending) {
					// The pointer is committed; never record a scan failure.
					return err
				}
				if err == nil {
					if checkpointErr := s.advanceCheckpoint(ctx, a, now); checkpointErr != nil {
						_ = s.markCheckpointLag(ctx, a, checkpointErr)
						return checkpointErr
					}
					return nil
				}
			}
		}
	}
	// Use the parent context for recording a scan timeout; the fence still wins
	// over every late error, config change, revocation and lease takeover.
	return s.Fail(ctx, a, err)
}

// RunOnce is shared by the runtime and integration tests. Connections/scopes
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

func (s *Service) advanceCheckpoint(ctx context.Context, a Attempt, now time.Time) error {
	if s.Checkpoint == nil {
		return nil
	}
	currentRecord, guardErr := s.Repo.Get(ctx, a.Record.WorkspaceID, a.Record.ID)
	if guardErr != nil {
		return guardErr
	}
	if currentRecord.Epoch != a.Record.Epoch || !slices.ContainsFunc(currentRecord.Scopes, func(sc Scope) bool {
		return sc.ID == a.Scope.ID && sc.Generation == a.Scope.Generation && sc.Snapshot == a.Token
	}) {
		return ErrFenced
	}
	key := CheckpointKey(a.Record.WorkspaceID, a.Record.ID, a.Scope.ID)
	current, exists, err := s.Checkpoint.Load(ctx, key)
	if err != nil {
		return err
	}
	expected := uint64(0)
	next := Checkpoint{Generation: a.Scope.Generation, UpdatedAt: now}
	if exists {
		expected = current.Revision
		if current.Generation == next.Generation {
			next.Sequence = current.Sequence + 1
		}
	} else {
		next.Sequence = 1
	}
	_, err = s.Checkpoint.CompareAndSwap(ctx, key, expected, next)
	return err
}

func (s *Service) markCheckpointLag(ctx context.Context, a Attempt, err error) error {
	category := ReportCategory(err)
	_, updateErr := s.mutate(ctx, a.Record.WorkspaceID, a.Record.ID, func(r *Record, _ time.Time) error {
		if r.Epoch != a.Record.Epoch {
			return ErrFenced
		}
		for i := range r.Scopes {
			if r.Scopes[i].ID == a.Scope.ID && r.Scopes[i].Generation == a.Scope.Generation && r.Scopes[i].Snapshot == a.Token {
				r.Scopes[i].CheckpointLag = true
				r.Scopes[i].CheckpointError = category
				return nil
			}
		}
		return ErrNotFound
	})
	return updateErr
}

func (s *Service) Cleanup(ctx context.Context) error {
	var last error
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(30 * time.Second)
	}
	task := CleanupTask{Key: "observations", Deadline: deadline}
	for attempt := 0; attempt < cleanupAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		last = s.Repo.Collect(ctx)
		if last == nil {
			return nil
		}
		if attempt == cleanupAttempts-1 {
			return last
		}
		var result CleanupResult
		task, result = PlanCleanupRetry(task, time.Now(), last, CleanupPolicy{BaseDelay: 100 * time.Millisecond, MaxDelay: time.Second})
		if result != CleanupRetry {
			return last
		}
		timer := time.NewTimer(max(time.Until(task.NextAt), time.Duration(0)))
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		}
	}
	return last
}
