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
	renewed   int
	released  int
}

func (l *fakeLease) Acquire(context.Context) (bool, error) { return l.acquireOK, l.acquireEr }
func (l *fakeLease) Renew(context.Context) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.renewed++
	return l.renewOK, nil
}
func (l *fakeLease) Release(context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.released++
	return nil
}

func redisGuardWith(lease *fakeLease, ttl time.Duration) *Redis {
	return &Redis{
		ttl:   ttl,
		lease: func(string, string) renewableLease { return lease },
	}
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
	release()

	lease.mu.Lock()
	defer lease.mu.Unlock()
	if lease.renewed == 0 || lease.released != 1 {
		t.Fatalf("renewed=%d released=%d", lease.renewed, lease.released)
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
	_, release, ok, err := guard.Acquire(ctx, "session-a")
	if err != nil || !ok {
		t.Fatalf("Acquire: ok=%v err=%v", ok, err)
	}
	cancel() // client disconnected mid-turn
	release()
	release()

	lease.mu.Lock()
	defer lease.mu.Unlock()
	if lease.released != 1 {
		t.Fatalf("released=%d, want exactly 1", lease.released)
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
