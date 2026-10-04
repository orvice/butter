package sessionguard

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type fakeLease struct {
	mu        sync.Mutex
	acquireOK bool
	acquireEr error
	renewOK   bool
	// renewErrs are returned by the first renewals, in order; later renewals
	// answer renewOK.
	renewErrs []error
	renewed   int
	released  int
}

func (l *fakeLease) Acquire(context.Context) (bool, error) { return l.acquireOK, l.acquireEr }
func (l *fakeLease) Renew(context.Context) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.renewed++
	if len(l.renewErrs) > 0 {
		err := l.renewErrs[0]
		l.renewErrs = l.renewErrs[1:]
		if err != nil {
			return false, err
		}
	}
	return l.renewOK, nil
}
func (l *fakeLease) Release(context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.released++
	return nil
}

func (l *fakeLease) counts() (renewed, released int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.renewed, l.released
}

func redisGuardWith(lease *fakeLease, ttl time.Duration) *Leased {
	return NewLeased("pod", ttl, func(string, string) Lease { return lease })
}

// A lost lease (expiry or takeover by another Pod) must cancel the turn
// context so the fenced-out holder stops acting instead of racing the new one.
func TestRedisLeaseLossCancelsTurnContext(t *testing.T) {
	lease := &fakeLease{acquireOK: true, renewOK: false}
	guard := redisGuardWith(lease, 15*time.Millisecond)

	leaseCtx, release, ok, err := guard.Acquire(t.Context(), "session-a")
	if err != nil || !ok {
		t.Fatalf("Acquire: ok=%v err=%v", ok, err)
	}
	select {
	case <-leaseCtx.Done():
	case <-time.After(time.Second):
		t.Fatal("turn context was not cancelled after session lease loss")
	}
	if cause := context.Cause(leaseCtx); !errors.Is(cause, ErrLeaseLost) {
		t.Fatalf("cause = %v, want ErrLeaseLost", cause)
	}
	release()

	if renewed, released := lease.counts(); renewed == 0 || released != 1 {
		t.Fatalf("renewed=%d released=%d", renewed, released)
	}
}

// A renewal that errors is not a lost lease: until one TTL has passed since
// the last renewal that got through, nobody else can hold the session, so the
// turn keeps going and the guard tries again.
func TestRedisTransientRenewErrorKeepsTheTurn(t *testing.T) {
	const ttl = 60 * time.Millisecond
	lease := &fakeLease{acquireOK: true, renewOK: true, renewErrs: []error{errors.New("redis: connection reset")}}
	guard := redisGuardWith(lease, ttl)

	leaseCtx, release, ok, err := guard.Acquire(t.Context(), "session-a")
	if err != nil || !ok {
		t.Fatalf("Acquire: ok=%v err=%v", ok, err)
	}
	defer release()
	// Many TTLs: the failed renewal was retried and later ones kept the lease.
	select {
	case <-leaseCtx.Done():
		t.Fatalf("turn cancelled after one transient renew error: %v", context.Cause(leaseCtx))
	case <-time.After(6 * ttl):
	}
	if renewed, _ := lease.counts(); renewed < 3 {
		t.Fatalf("renewed %d times, want the failed renewal retried and the lease kept renewed", renewed)
	}
}

// Renewals that keep failing end the turn once the lease would really have
// lapsed, and not before.
func TestRedisRenewErrorsEndTheTurnWhenTheLeaseLapses(t *testing.T) {
	const ttl = 90 * time.Millisecond
	errs := make([]error, 1000)
	for i := range errs {
		errs[i] = errors.New("redis: i/o timeout")
	}
	lease := &fakeLease{acquireOK: true, renewOK: true, renewErrs: errs}
	guard := redisGuardWith(lease, ttl)

	start := time.Now()
	leaseCtx, release, ok, err := guard.Acquire(t.Context(), "session-a")
	if err != nil || !ok {
		t.Fatalf("Acquire: ok=%v err=%v", ok, err)
	}
	defer release()
	select {
	case <-leaseCtx.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("turn kept going long after its lease lapsed")
	}
	if elapsed := time.Since(start); elapsed < ttl-5*time.Millisecond {
		t.Fatalf("turn ended after %v, before its lease could lapse (ttl %v)", elapsed, ttl)
	}
	if cause := context.Cause(leaseCtx); !errors.Is(cause, ErrLeaseLost) {
		t.Fatalf("cause = %v, want ErrLeaseLost", cause)
	}
	if renewed, _ := lease.counts(); renewed < 2 {
		t.Fatalf("renewed %d times, want the failing renewal retried before giving up", renewed)
	}
}

// A held lease means busy, not an error: the caller distinguishes "someone
// else is running this session" from "Redis is down".
func TestRedisBusyWhenLeaseHeldElsewhere(t *testing.T) {
	guard := redisGuardWith(&fakeLease{acquireOK: false}, time.Minute)

	_, release, ok, err := guard.Acquire(t.Context(), "session-a")
	if err != nil {
		t.Fatalf("Acquire err = %v", err)
	}
	if ok {
		t.Fatal("Acquire reported acquired for a lease held elsewhere")
	}
	release() // must be a safe no-op
}

func TestRedisAcquireErrorSurfaces(t *testing.T) {
	guard := redisGuardWith(&fakeLease{acquireEr: errors.New("redis down")}, time.Minute)

	_, _, ok, err := guard.Acquire(t.Context(), "session-a")
	if ok || err == nil {
		t.Fatalf("Acquire: ok=%v err=%v, want busy with error", ok, err)
	}
}

// Release is idempotent and releases the underlying lease exactly once even
// when the turn context is already cancelled (client disconnect).
func TestRedisReleaseOnceAfterDisconnect(t *testing.T) {
	lease := &fakeLease{acquireOK: true, renewOK: true}
	guard := redisGuardWith(lease, time.Minute)

	ctx, cancel := context.WithCancel(t.Context())
	leaseCtx, release, ok, err := guard.Acquire(ctx, "session-a")
	if err != nil || !ok {
		t.Fatalf("Acquire: ok=%v err=%v", ok, err)
	}
	cancel() // client disconnected mid-turn
	release()
	release()

	if _, released := lease.counts(); released != 1 {
		t.Fatalf("released=%d, want exactly 1", released)
	}
	// The turn ended with its request, not because the lease was lost.
	if errors.Is(context.Cause(leaseCtx), ErrLeaseLost) {
		t.Fatal("a disconnect was reported as a lost lease")
	}
}

// The turn context names its acquisition: it carries the token the lease is
// held under, a new one for every acquisition (#402), so a Stop bound to one
// turn's token never reaches the next turn on the session.
func TestLeasedTurnContextCarriesTheAcquisitionToken(t *testing.T) {
	var holders []string
	guard := NewLeased("pod-a", time.Minute, func(_, leaseHolder string) Lease {
		holders = append(holders, leaseHolder)
		return &fakeLease{acquireOK: true, renewOK: true}
	})

	var tokens []string
	for range 2 {
		leaseCtx, release, ok, err := guard.Acquire(t.Context(), "session-a")
		if err != nil || !ok {
			t.Fatalf("Acquire: ok=%v err=%v", ok, err)
		}
		token, ok := Token(leaseCtx)
		if !ok {
			t.Fatal("the turn context carries no token")
		}
		tokens = append(tokens, token)
		release()
	}
	if len(holders) != 2 || tokens[0] != holders[0] || tokens[1] != holders[1] {
		t.Fatalf("tokens %v, leases held under %v: the token is not the lease's holder value", tokens, holders)
	}
	if tokens[0] == tokens[1] {
		t.Fatalf("two acquisitions share the token %q", tokens[0])
	}
	// A context derived from the turn context still names it.
	leaseCtx, release, _, _ := guard.Acquire(t.Context(), "session-a")
	defer release()
	derived, cancel := context.WithCancel(leaseCtx)
	defer cancel()
	if token, _ := Token(derived); token != holders[2] {
		t.Fatalf("derived context token = %q, want %q", token, holders[2])
	}
}

// A busy session, or a context no acquisition made, has no token.
func TestTokenIsAbsentWithoutAnAcquisition(t *testing.T) {
	if token, ok := Token(t.Context()); ok || token != "" {
		t.Fatalf("Token(plain context) = %q, %v", token, ok)
	}
	guard := redisGuardWith(&fakeLease{acquireOK: false}, time.Minute)
	busyCtx, _, ok, err := guard.Acquire(t.Context(), "session-a")
	if err != nil || ok {
		t.Fatalf("Acquire: ok=%v err=%v, want busy", ok, err)
	}
	if _, ok := Token(busyCtx); ok {
		t.Fatal("a busy acquisition handed out a token")
	}
}

// The in-process guard hands out a token per acquisition too.
func TestMemoryGuardTurnContextCarriesAToken(t *testing.T) {
	guard := NewMemory()
	firstCtx, release, ok, _ := guard.Acquire(t.Context(), "session-a")
	if !ok {
		t.Fatal("Acquire failed")
	}
	first, ok := Token(firstCtx)
	if !ok {
		t.Fatal("the turn context carries no token")
	}
	release()
	secondCtx, release, ok, _ := guard.Acquire(t.Context(), "session-a")
	if !ok {
		t.Fatal("Acquire failed")
	}
	defer release()
	if second, _ := Token(secondCtx); second == "" || second == first {
		t.Fatalf("tokens %q and %q: want a new one per acquisition", first, second)
	}
}

func TestMemoryGuardSerializesOneSessionOnly(t *testing.T) {
	guard := NewMemory()

	_, releaseA, ok, err := guard.Acquire(t.Context(), "session-a")
	if err != nil || !ok {
		t.Fatalf("first Acquire: ok=%v err=%v", ok, err)
	}

	// Same session is busy; an unrelated session proceeds.
	if _, _, ok, _ := guard.Acquire(t.Context(), "session-a"); ok {
		t.Fatal("second Acquire on a held session succeeded")
	}
	_, releaseB, ok, _ := guard.Acquire(t.Context(), "session-b")
	if !ok {
		t.Fatal("unrelated session was blocked")
	}
	releaseB()

	releaseA()
	releaseA() // idempotent
	if _, release, ok, _ := guard.Acquire(t.Context(), "session-a"); !ok {
		t.Fatal("session stayed busy after release")
	} else {
		release()
	}
}
