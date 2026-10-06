package app

import (
	"context"
	"io"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/maintenance/webchatcleanup"
	"go.orx.me/apps/butter/internal/redislease"
)

// TEMPORARY (#411): the startup web-chat cleanup below deletes the old
// dashboard Chat's leftover data. The project owner decided that this release
// deletes it automatically on startup, which is their confirmation for the
// production deletion, and that a later release removes this startup logic
// again. Remove this file, its call in StartChannels and the
// maintenance.delete_web_chat flag in a later release, once it has run in
// production.

const (
	// webChatCleanupLeaseKey elects the one process that runs the startup
	// web-chat cleanup.
	webChatCleanupLeaseKey = "butter:maintenance:web-chat-cleanup"
	// webChatCleanupLeaseTTL is how long a process that died while cleaning
	// keeps the others from starting the cleanup. The holder renews the
	// lease every TTL/3.
	webChatCleanupLeaseTTL = 30 * time.Second
	// webChatCleanupTimeout bounds the startup cleanup, taking the lease
	// included. What it leaves is deleted at a later startup.
	webChatCleanupTimeout = 10 * time.Minute
)

// leaseGuard elects the one holder of a key across processes and keeps the
// lease renewed while the holder works. It cancels the context Acquire
// returns if the holder loses the lease. *redislease.Guard implements it.
type leaseGuard interface {
	Acquire(ctx context.Context, key string) (context.Context, func(), bool, error)
}

// webChatCleanup deletes the old dashboard Chat's leftover web-chat data
// (#411) once per startup, in the background. It runs on every startup, with
// no marker of having run: once nothing is left it is a quick no-op, and it
// removes what a Pod still running the old release created meanwhile.
// TEMPORARY: see the note at the top of this file.
type webChatCleanup struct {
	// enabled is maintenance.delete_web_chat.
	enabled bool
	// guard elects the one process that runs the cleanup. It is nil without
	// Redis, where this process is the only one.
	guard leaseGuard
	// clean deletes the data and reports what it found and deleted.
	clean func(context.Context) (*webchatcleanup.Report, error)
	// timeout bounds one cleanup, taking the lease included.
	timeout time.Duration
	logger  *slog.Logger
}

// newWebChatCleanup wires the startup cleanup to db, under a Redis lease held
// as holder when rdb is set.
func newWebChatCleanup(cfg config.MaintenanceConfig, db *mongo.Database, rdb *redis.Client, holder string, logger *slog.Logger) webChatCleanup {
	c := webChatCleanup{
		enabled: cfg.EffectiveDeleteWebChat(),
		clean: func(ctx context.Context) (*webchatcleanup.Report, error) {
			return webchatcleanup.Run(ctx, db, true, io.Discard)
		},
		timeout: webChatCleanupTimeout,
		logger:  logger,
	}
	if rdb != nil {
		c.guard = redislease.NewGuard(rdb, holder, webChatCleanupLeaseTTL)
	}
	return c
}

// start runs the cleanup in a background goroutine and returns at once, so
// it never holds up startup. The channel it returns is closed once the
// cleanup has ended or was skipped.
func (c webChatCleanup) start(ctx context.Context) <-chan struct{} {
	done := make(chan struct{})
	if !c.enabled {
		c.logger.Info("web-chat cleanup: skipped; maintenance.delete_web_chat is false")
		close(done)
		return done
	}
	go func() {
		defer close(done)
		// Nothing the cleanup does may take the process down.
		defer func() {
			if r := recover(); r != nil {
				c.logger.Warn("web-chat cleanup: panicked; the next startup runs it again",
					"panic", r, "stack", string(debug.Stack()))
			}
		}()
		c.run(ctx)
	}()
	return done
}

// run takes the lease, when there is one, and cleans under it. A process
// that does not get the lease skips the cleanup. Without Redis this process
// is the only one, and the cleanup is idempotent, so it runs without a lease.
func (c webChatCleanup) run(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	if c.guard != nil {
		leaseCtx, release, ok, err := c.guard.Acquire(ctx, webChatCleanupLeaseKey)
		if err != nil {
			c.logger.Warn("web-chat cleanup: skipped; could not take the lease, the next startup tries again",
				"lease", webChatCleanupLeaseKey, "err", err)
			return
		}
		if !ok {
			c.logger.Info("web-chat cleanup: skipped; another process holds the lease and runs it",
				"lease", webChatCleanupLeaseKey)
			return
		}
		defer release()
		ctx = leaseCtx
	}
	c.logger.Info("web-chat cleanup: started", "timeout", c.timeout)
	began := time.Now()
	rep, err := c.clean(ctx)
	c.report(rep, err, time.Since(began))
}

// report logs what the cleanup found per workspace and what it deleted per
// collection, at Info, and an error at Warn.
func (c webChatCleanup) report(rep *webchatcleanup.Report, err error, took time.Duration) {
	if rep != nil && rep.Found != nil {
		for _, w := range rep.Found.Workspaces {
			c.logger.Info("web-chat cleanup: found in workspace",
				append([]any{"workspace", w.Workspace}, countAttrs(w.Counts)...)...)
		}
	}
	if err != nil {
		attrs := []any{"err", err}
		if rep != nil {
			// What it deleted before the error.
			attrs = append(attrs, slog.Group("deleted", countAttrs(rep.Deleted)...))
		}
		c.logger.Warn("web-chat cleanup: failed; the next startup finishes it", append(attrs, "took", took)...)
		return
	}
	if rep == nil || rep.Found == nil || rep.Found.Total() == (webchatcleanup.Counts{}) {
		c.logger.Info("web-chat cleanup: nothing to delete", "took", took)
		return
	}
	attrs := append(countAttrs(rep.Deleted), "kept_other_app_invocations", rep.Found.OtherApps)
	if !rep.Found.FirstStart.IsZero() {
		// Past Activity counts drop over this span.
		attrs = append(attrs, "first_started_at", rep.Found.FirstStart, "last_started_at", rep.Found.LastStart)
	}
	c.logger.Info("web-chat cleanup: deleted", append(attrs, "took", took)...)
}

// countAttrs are counts as log attributes, keyed by collection.
func countAttrs(c webchatcleanup.Counts) []any {
	return []any{
		webchatcleanup.SessionsCollection, c.Sessions,
		webchatcleanup.EventsCollection, c.Events,
		webchatcleanup.InvocationsCollection, c.Invocations,
		webchatcleanup.InputPartsCollection, c.InputParts,
	}
}
