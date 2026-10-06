package app

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/maintenance/webchatcleanup"
	"go.orx.me/apps/butter/internal/redislease"
)

// steps records what happened, in order, across goroutines.
type steps struct {
	mu   sync.Mutex
	list []string
}

func (s *steps) add(step string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.list = append(s.list, step)
}

func (s *steps) all() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.list)
}

// leaseKey marks the context a fakeLease hands its holder.
type leaseKey struct{}

// fakeLease is one lease that the processes of a test share. It has one
// holder at a time, as the Redis lease does.
type fakeLease struct {
	steps *steps
	// err, when set, fails every Acquire.
	err error

	mu   sync.Mutex
	held bool
	keys []string
}

func (l *fakeLease) Acquire(ctx context.Context, key string) (context.Context, func(), bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.keys = append(l.keys, key)
	if l.err != nil {
		return ctx, func() {}, false, l.err
	}
	if l.held {
		return ctx, func() {}, false, nil
	}
	l.held = true
	l.steps.add("acquire")
	var once sync.Once
	release := func() {
		once.Do(func() {
			l.mu.Lock()
			defer l.mu.Unlock()
			l.held = false
			l.steps.add("release")
		})
	}
	return context.WithValue(ctx, leaseKey{}, key), release, true, nil
}

// acquired lists the key of every Acquire.
func (l *fakeLease) acquired() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return slices.Clone(l.keys)
}

// fakeClean is a cleanup that records its calls and answers rep and err.
type fakeClean struct {
	steps *steps
	rep   *webchatcleanup.Report
	err   error
	// panics makes a call panic.
	panics bool
	// begun, when set, is closed once a call has begun.
	begun chan struct{}
	// proceed, when set, holds every call until it is closed.
	proceed chan struct{}

	once sync.Once
	mu   sync.Mutex
	ctxs []context.Context
}

func (f *fakeClean) clean(ctx context.Context) (*webchatcleanup.Report, error) {
	f.mu.Lock()
	f.ctxs = append(f.ctxs, ctx)
	f.mu.Unlock()
	f.steps.add("clean")
	if f.begun != nil {
		f.once.Do(func() { close(f.begun) })
	}
	if f.proceed != nil {
		<-f.proceed
	}
	if f.panics {
		panic("boom")
	}
	return f.rep, f.err
}

// calls lists the context of every call.
func (f *fakeClean) calls() []context.Context {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.ctxs)
}

// capturedLogs is a slog.Handler that keeps each record as one line:
// "LEVEL message key=value ...".
type capturedLogs struct {
	mu    sync.Mutex
	lines []string
}

func (c *capturedLogs) Enabled(context.Context, slog.Level) bool { return true }

func (c *capturedLogs) Handle(_ context.Context, r slog.Record) error {
	line := r.Level.String() + " " + r.Message
	r.Attrs(func(a slog.Attr) bool {
		line += " " + a.String()
		return true
	})
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lines = append(c.lines, line)
	return nil
}

func (c *capturedLogs) WithAttrs([]slog.Attr) slog.Handler { return c }
func (c *capturedLogs) WithGroup(string) slog.Handler      { return c }

func (c *capturedLogs) all() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return slices.Clone(c.lines)
}

// want checks that lines starting with these prefixes were logged, in this
// order.
func (c *capturedLogs) want(t *testing.T, prefixes ...string) {
	t.Helper()
	lines := c.all()
	next := 0
	for _, line := range lines {
		if next < len(prefixes) && strings.HasPrefix(line, prefixes[next]) {
			next++
		}
	}
	if next < len(prefixes) {
		t.Errorf("no log line starts with %q (in order); logged:\n%s", prefixes[next], strings.Join(lines, "\n"))
	}
}

// none checks that no logged line starts with prefix.
func (c *capturedLogs) none(t *testing.T, prefix string) {
	t.Helper()
	for _, line := range c.all() {
		if strings.HasPrefix(line, prefix) {
			t.Errorf("logged %q", line)
		}
	}
}

// waitFor waits for ch to close.
func waitFor(t *testing.T, ch <-chan struct{}, what string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for %s", what)
	}
}

// emptyReport is a cleanup that found nothing.
func emptyReport() *webchatcleanup.Report {
	return &webchatcleanup.Report{Found: &webchatcleanup.Found{}}
}

// foundReport is a cleanup that found web-chat data in a workspace and
// without one, and deleted it.
func foundReport() *webchatcleanup.Report {
	return &webchatcleanup.Report{
		Found: &webchatcleanup.Found{
			Workspaces: []webchatcleanup.WorkspaceCounts{
				{Workspace: "ws-a", Counts: webchatcleanup.Counts{Sessions: 3, Events: 4, Invocations: 4, InputParts: 4}},
				{Workspace: webchatcleanup.NoWorkspace, Counts: webchatcleanup.Counts{Sessions: 1, Invocations: 1, InputParts: 1}},
			},
			OtherApps:  1,
			FirstStart: time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC),
			LastStart:  time.Date(2026, 5, 1, 15, 0, 0, 0, time.UTC),
		},
		Deleted: webchatcleanup.Counts{Sessions: 4, Events: 4, Invocations: 5, InputParts: 5},
	}
}

// The cleanup runs once, between taking and releasing the lease, on the
// lease's context and within its timeout. It logs what it found per
// workspace and what it deleted per collection.
func TestWebChatCleanupRunsOnceUnderTheLease(t *testing.T) {
	order := &steps{}
	lease := &fakeLease{steps: order}
	cleanup := &fakeClean{steps: order, rep: foundReport()}
	logs := &capturedLogs{}
	c := webChatCleanup{enabled: true, guard: lease, clean: cleanup.clean, timeout: time.Minute, logger: slog.New(logs)}

	waitFor(t, c.start(t.Context()), "the cleanup")

	if got := order.all(); !slices.Equal(got, []string{"acquire", "clean", "release"}) {
		t.Fatalf("steps = %v, want the cleanup once, between taking and releasing the lease", got)
	}
	if got := lease.acquired(); !slices.Equal(got, []string{webChatCleanupLeaseKey}) {
		t.Errorf("lease keys = %v, want [%s]", got, webChatCleanupLeaseKey)
	}
	ctx := cleanup.calls()[0]
	if ctx.Value(leaseKey{}) != webChatCleanupLeaseKey {
		t.Error("the cleanup did not run on the lease's context")
	}
	if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) > time.Minute {
		t.Errorf("the cleanup's deadline = %v (set: %v), want one within its timeout", deadline, ok)
	}
	logs.want(t,
		"INFO web-chat cleanup: started",
		"INFO web-chat cleanup: found in workspace workspace=ws-a adk_sessions=3 adk_events=4 invocations=4 invocation_input_parts=4",
		"INFO web-chat cleanup: found in workspace workspace=(none) adk_sessions=1 adk_events=0 invocations=1 invocation_input_parts=1",
		"INFO web-chat cleanup: deleted adk_sessions=4 adk_events=4 invocations=5 invocation_input_parts=5 kept_other_app_invocations=1 "+
			"first_started_at=2026-05-01 09:00:00 +0000 UTC last_started_at=2026-05-01 15:00:00 +0000 UTC took=",
	)
	logs.none(t, "WARN")
}

// Of two processes starting together, the one that does not get the lease
// skips the cleanup, and the holder runs it.
func TestWebChatCleanupSkipsWhenAnotherProcessHoldsTheLease(t *testing.T) {
	order := &steps{}
	lease := &fakeLease{steps: order}
	holder := &fakeClean{steps: order, rep: emptyReport(), begun: make(chan struct{}), proceed: make(chan struct{})}
	other := &fakeClean{steps: order, rep: emptyReport()}
	otherLogs := &capturedLogs{}

	holderDone := webChatCleanup{enabled: true, guard: lease, clean: holder.clean, timeout: time.Minute,
		logger: slog.New(&capturedLogs{})}.start(t.Context())
	waitFor(t, holder.begun, "the holder's cleanup to begin")
	waitFor(t, webChatCleanup{enabled: true, guard: lease, clean: other.clean, timeout: time.Minute,
		logger: slog.New(otherLogs)}.start(t.Context()), "the other process")

	if n := len(other.calls()); n != 0 {
		t.Fatalf("the process without the lease cleaned %d time(s)", n)
	}
	otherLogs.want(t, "INFO web-chat cleanup: skipped; another process holds the lease and runs it lease=butter:maintenance:web-chat-cleanup")

	close(holder.proceed)
	waitFor(t, holderDone, "the holder's cleanup")
	if n := len(holder.calls()); n != 1 {
		t.Fatalf("the holder cleaned %d time(s), want 1", n)
	}
	if got := order.all(); !slices.Equal(got, []string{"acquire", "clean", "release"}) {
		t.Errorf("steps = %v, want one cleanup under the lease", got)
	}
	if got := lease.acquired(); !slices.Equal(got, []string{webChatCleanupLeaseKey, webChatCleanupLeaseKey}) {
		t.Errorf("lease keys = %v, want both processes to ask for %s", got, webChatCleanupLeaseKey)
	}
}

// maintenance.delete_web_chat: false skips the cleanup: no lease, no
// cleanup, and a line that says so.
func TestWebChatCleanupKillSwitchSkipsIt(t *testing.T) {
	off := false
	logs := &capturedLogs{}
	c := newWebChatCleanup(config.MaintenanceConfig{DeleteWebChat: &off}, nil, nil, "this-process", slog.New(logs))
	order := &steps{}
	lease := &fakeLease{steps: order}
	cleanup := &fakeClean{steps: order, rep: emptyReport()}
	c.guard, c.clean = lease, cleanup.clean

	done := c.start(t.Context())
	select {
	case <-done:
	default:
		t.Fatal("a skipped cleanup left its channel open")
	}
	if got := order.all(); len(got) != 0 {
		t.Fatalf("steps = %v, want none", got)
	}
	if got := lease.acquired(); len(got) != 0 {
		t.Fatalf("asked for the lease %v, want never", got)
	}
	logs.want(t, "INFO web-chat cleanup: skipped; maintenance.delete_web_chat is false")
}

// An error is logged at Warn and goes no further: the cleanup's goroutine
// ends, nothing panics past it, and the next startup tries again.
func TestWebChatCleanupLogsAnErrorAtWarnAndCarriesOn(t *testing.T) {
	partial := emptyReport()
	partial.Deleted = webchatcleanup.Counts{InputParts: 2, Invocations: 1}
	for _, tc := range []struct {
		name     string
		leaseErr error
		cleanup  *fakeClean
		want     string
	}{
		{
			name:    "finding fails",
			cleanup: &fakeClean{err: errors.New("mongo unreachable")},
			want:    "WARN web-chat cleanup: failed; the next startup finishes it err=mongo unreachable took=",
		},
		{
			name:    "deleting fails partway",
			cleanup: &fakeClean{rep: partial, err: errors.New("delete from adk_events: timeout")},
			want: "WARN web-chat cleanup: failed; the next startup finishes it err=delete from adk_events: timeout " +
				"deleted=[adk_sessions=0 adk_events=0 invocations=1 invocation_input_parts=2] took=",
		},
		{
			name:    "the cleanup panics",
			cleanup: &fakeClean{panics: true},
			want:    "WARN web-chat cleanup: panicked; the next startup runs it again panic=boom",
		},
		{
			name:     "the lease cannot be taken",
			leaseErr: errors.New("redis down"),
			cleanup:  &fakeClean{rep: emptyReport()},
			want:     "WARN web-chat cleanup: skipped; could not take the lease, the next startup tries again lease=butter:maintenance:web-chat-cleanup err=redis down",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			order := &steps{}
			tc.cleanup.steps = order
			logs := &capturedLogs{}
			c := webChatCleanup{enabled: true, guard: &fakeLease{steps: order, err: tc.leaseErr},
				clean: tc.cleanup.clean, timeout: time.Minute, logger: slog.New(logs)}

			waitFor(t, c.start(t.Context()), "the cleanup")

			logs.want(t, tc.want)
			logs.none(t, "INFO web-chat cleanup: deleted")
			logs.none(t, "ERROR")
			if tc.leaseErr != nil && len(tc.cleanup.calls()) != 0 {
				t.Error("the cleanup ran without the lease")
			}
		})
	}
}

// start returns while the cleanup still runs: startup never waits for it.
func TestWebChatCleanupDoesNotBlockStartup(t *testing.T) {
	cleanup := &fakeClean{steps: &steps{}, rep: emptyReport(), begun: make(chan struct{}), proceed: make(chan struct{})}
	c := webChatCleanup{enabled: true, guard: &fakeLease{steps: &steps{}}, clean: cleanup.clean,
		timeout: time.Minute, logger: slog.New(&capturedLogs{})}

	started := make(chan (<-chan struct{}), 1)
	go func() { started <- c.start(t.Context()) }()
	var done <-chan struct{}
	select {
	case done = <-started:
	case <-time.After(5 * time.Second):
		close(cleanup.proceed)
		t.Fatal("start waited for the cleanup")
	}
	waitFor(t, cleanup.begun, "the cleanup to begin")
	select {
	case <-done:
		t.Fatal("the cleanup ended before it was let go")
	default:
	}
	close(cleanup.proceed)
	waitFor(t, done, "the cleanup")
}

// Without Redis this process is the only one, and the cleanup runs without a
// lease. With Redis it takes the Redis lease. It is on by default.
func TestWebChatCleanupTakesALeaseOnlyWithRedis(t *testing.T) {
	logs := &capturedLogs{}
	without := newWebChatCleanup(config.MaintenanceConfig{}, nil, nil, "this-process", slog.New(logs))
	if !without.enabled {
		t.Fatal("the cleanup is off without maintenance.delete_web_chat")
	}
	if without.timeout != webChatCleanupTimeout {
		t.Errorf("timeout = %v, want %v", without.timeout, webChatCleanupTimeout)
	}
	if without.guard != nil {
		t.Fatalf("without Redis the cleanup takes a lease: %T", without.guard)
	}
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"}) // never dialed
	t.Cleanup(func() { _ = rdb.Close() })
	with := newWebChatCleanup(config.MaintenanceConfig{}, nil, rdb, "this-process", slog.New(logs))
	if _, ok := with.guard.(*redislease.Guard); !ok {
		t.Fatalf("with Redis the lease is %T, want *redislease.Guard", with.guard)
	}

	cleanup := &fakeClean{steps: &steps{}, rep: emptyReport()}
	without.clean = cleanup.clean
	waitFor(t, without.start(t.Context()), "the cleanup")
	if n := len(cleanup.calls()); n != 1 {
		t.Fatalf("cleaned %d time(s) without Redis, want 1", n)
	}
	logs.want(t, "INFO web-chat cleanup: started", "INFO web-chat cleanup: nothing to delete took=")
	logs.none(t, "WARN")
}
