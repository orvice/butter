package telegram

import (
	"context"
	"sync"
	"testing"
	"time"
)

type fakeRenewableLease struct {
	mu       sync.Mutex
	renewed  int
	released int
}

func (l *fakeRenewableLease) Acquire(context.Context) (bool, error) { return true, nil }
func (l *fakeRenewableLease) Renew(context.Context) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.renewed++
	return false, nil
}
func (l *fakeRenewableLease) Release(context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.released++
	return nil
}

func TestSessionLeaseLossCancelsTurnContext(t *testing.T) {
	lease := &fakeRenewableLease{}
	guard := &RedisSessionGuard{
		ttl:   15 * time.Millisecond,
		lease: func(string, string) renewableLease { return lease },
	}

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
