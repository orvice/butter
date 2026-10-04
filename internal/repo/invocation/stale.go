package invocation

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// NoReplaySuffix closes every operational failure reason recorded on an
// Invocation: work is never replayed automatically, and the way back is an
// explicit resubmission. Reasons are shown verbatim beside the turn.
const NoReplaySuffix = "; no work was replayed automatically. Review your message and resubmit to retry."

// LegacyStaleAge is how old a QUEUED or RUNNING record without an owner stamp
// must be before a sweep fails it. Such records were written before records
// carried an owner, so no liveness can vouch for them. Like automation's
// StaleRunAge, the cutoff sits well above any run a live process could still
// be doing.
const LegacyStaleAge = 24 * time.Hour

// Liveness reports which processes are alive. liveness.Registry satisfies it.
type Liveness interface {
	Alive(ctx context.Context, ids []string) (map[string]bool, error)
}

// StaleSweeper fails the QUEUED and RUNNING Invocations that no live process
// will finish, whatever entry point recorded them (#390, ADR-0016 decision 3).
//
// Each process stamps the records it creates with its instance ID and keeps
// its liveness published. A stamped record is stale only once its owner's
// liveness has lapsed, never because another process started. A record
// without a stamp is stale once it is older than LegacyAge.
//
// A sweep only marks records FAILED. It never re-invokes an Agent or repeats
// a tool's side effects.
type StaleSweeper struct {
	Repo     Repository
	Liveness Liveness
	// Self is the sweeping process's instance ID. Its own records are never
	// stale to it: the process asking is alive.
	Self string
	// LegacyAge is how old a record without an owner stamp must be to be
	// failed. Zero fails every such record, which is right only when this is
	// the one process: then nothing else can be running them.
	LegacyAge time.Duration
	// Now is the clock; nil means time.Now.
	Now func() time.Time
}

// SweepResult reports what one sweep did.
type SweepResult struct {
	// Failed is how many records the sweep failed.
	Failed int64
	// LostOwners are the instance IDs that still owned QUEUED or RUNNING
	// records after their liveness had lapsed.
	LostOwners []string
}

// Sweep fails the records whose owner is gone, and the records without an
// owner that are older than LegacyAge.
func (s *StaleSweeper) Sweep(ctx context.Context) (SweepResult, error) {
	if s.Repo == nil || s.Liveness == nil {
		return SweepResult{}, errors.New("stale invocation sweep needs a repository and a liveness registry")
	}
	owners, err := s.Repo.ActiveOwners(ctx)
	if err != nil {
		return SweepResult{}, fmt.Errorf("list the owners of active invocations: %w", err)
	}
	others := make([]string, 0, len(owners))
	for _, owner := range owners {
		if owner != s.Self {
			others = append(others, owner)
		}
	}
	var lost []string
	if len(others) > 0 {
		alive, err := s.Liveness.Alive(ctx, others)
		if err != nil {
			// No answer is not an answer of "gone": fail nothing this time.
			return SweepResult{}, fmt.Errorf("check the liveness of invocation owners: %w", err)
		}
		for _, owner := range others {
			if !alive[owner] {
				lost = append(lost, owner)
			}
		}
	}
	now := time.Now
	if s.Now != nil {
		now = s.Now
	}
	failed, err := s.Repo.MarkStaleRunning(ctx, StaleSelection{
		LostOwners:   lost,
		LegacyBefore: now().Add(-s.LegacyAge),
		Reason:       staleReason,
	})
	return SweepResult{Failed: failed, LostOwners: lost}, err
}

// staleReason is the error recorded on a record a sweep fails. It names the
// lost owner when the record has one.
func staleReason(owner string) string {
	if owner == "" {
		return "interrupted because the process running it stopped before it could finish" + NoReplaySuffix
	}
	return "interrupted because the process running it (instance " + owner + ") stopped before it could finish" + NoReplaySuffix
}
