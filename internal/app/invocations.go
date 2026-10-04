package app

import (
	"context"
	"time"

	"butterfly.orx.me/core/log"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"go.orx.me/apps/butter/internal/repo/invocation"
	"go.orx.me/apps/butter/internal/runtime/liveness"
)

const (
	// invocationSweepInterval paces the periodic stale-invocation sweep. A
	// run whose process died is failed within about one liveness TTL plus
	// this.
	invocationSweepInterval = time.Minute
	// invocationSweepTimeout bounds one sweep, so an unresponsive Redis or
	// MongoDB cannot hold up startup on it.
	invocationSweepTimeout = 30 * time.Second
)

// processLiveness is this process's identity as an owner of Invocation
// records (#390, ADR-0016 decision 3).
type processLiveness struct {
	// instanceID is stamped on every Invocation record this process creates.
	instanceID string
	registry   liveness.Registry
	// shared reports whether other processes publish to the same registry.
	// Without Redis there is only this process.
	shared bool
}

// startProcessLiveness gives this process its instance ID and publishes it.
// It runs before the process can write an Invocation record, so no record
// carries an owner that other processes cannot see yet.
func startProcessLiveness(ctx context.Context, rdb *redis.Client) processLiveness {
	p := processLiveness{instanceID: uuid.NewString(), shared: rdb != nil}
	if p.shared {
		p.registry = liveness.NewRedis(rdb, liveness.DefaultTTL)
	} else {
		p.registry = liveness.NewMemory(liveness.DefaultTTL)
	}
	logger := log.FromContext(ctx)
	// Renew for as long as the process runs, not just until the bootstrap
	// context ends: teardown cancels that context first, while in-flight
	// runs are still writing their terminal state.
	if err := liveness.Keep(context.WithoutCancel(ctx), p.registry, p.instanceID); err != nil {
		logger.Warn("instance liveness not published yet; retrying in the background. Until it is, "+
			"other processes may fail this process's runs as stale",
			"instance_id", p.instanceID, "err", err)
	} else {
		logger.Info("instance liveness published",
			"instance_id", p.instanceID, "shared", p.shared, "ttl", p.registry.TTL())
	}
	return p
}

// startInvocationSweep fails the QUEUED and RUNNING invocations whose owning
// process is gone. It sweeps once now and, when other processes can own
// records too, every invocationSweepInterval until ctx ends.
//
// With Redis, a record is failed only once its owner's liveness key has
// lapsed, so another Pod starting never fails a run still going. A record
// without an owner, written before owners were stamped, is failed once it is
// older than legacyAge. Without Redis this process is the only one: the
// startup sweep fails every record it does not own, as before, and there is
// nothing to sweep afterwards.
func startInvocationSweep(ctx context.Context, repo invocation.Repository, p processLiveness, legacyAge time.Duration) {
	sweeper := &invocation.StaleSweeper{
		Repo:      repo,
		Liveness:  p.registry,
		Self:      p.instanceID,
		LegacyAge: legacyAge,
	}
	if !p.shared {
		sweeper.LegacyAge = 0
	}
	sweepInvocations(ctx, sweeper)
	if !p.shared {
		return
	}
	go func() {
		ticker := time.NewTicker(invocationSweepInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				sweepInvocations(ctx, sweeper)
			}
		}
	}()
}

func sweepInvocations(ctx context.Context, sweeper *invocation.StaleSweeper) {
	logger := log.FromContext(ctx)
	sweepCtx, cancel := context.WithTimeout(ctx, invocationSweepTimeout)
	defer cancel()
	result, err := sweeper.Sweep(sweepCtx)
	if err != nil {
		if ctx.Err() == nil {
			logger.Warn("stale invocation sweep failed; records whose process is gone stay QUEUED/RUNNING until a later sweep",
				"err", err)
		}
		return
	}
	if result.Failed > 0 {
		logger.Info("failed invocations whose owning process is gone",
			"count", result.Failed,
			"lost_owners", result.LostOwners,
			"unowned_older_than", sweeper.LegacyAge,
		)
	}
}
