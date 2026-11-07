package plane

import (
	"context"
	"github.com/redis/go-redis/v9"
	"os"
	"testing"
	"time"
)

// This is an opt-in real Redis contract test. It is skipped unless
// BUTTER_REDIS_TEST_ADDR is supplied; a skip is evidence that Redis was not run.
func TestRedisFencedTokenLeaseContract(t *testing.T) {
	addr := os.Getenv("BUTTER_REDIS_TEST_ADDR")
	if addr == "" {
		t.Skip("real Redis contract not run: set BUTTER_REDIS_TEST_ADDR")
	}
	r := redis.NewClient(&redis.Options{Addr: addr})
	defer r.Close()
	ctx := context.Background()
	if e := r.Ping(ctx).Err(); e != nil {
		t.Fatalf("redis unavailable: %v", e)
	}
	l := &RedisFencedTokenLease{Client: r, Prefix: "test:"}
	key := "contract-" + time.Now().Format("150405.000000000")
	a, e := l.AcquireToken(ctx, key, "one", 1, time.Now().Add(time.Minute))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = l.AcquireToken(ctx, key, "two", 1, time.Now().Add(time.Minute)); e != ErrTokenFenceConflict {
		t.Fatalf("owner condition=%v", e)
	}
	if e := l.ReleaseToken(ctx, key, "two", a); e != ErrTokenFenceConflict {
		t.Fatalf("release owner=%v", e)
	}
	if e := l.ReleaseToken(ctx, key, "one", a+1); e != ErrTokenFenceConflict {
		t.Fatalf("release fence=%v", e)
	}
	if e := l.ReleaseToken(ctx, key, "one", a); e != nil {
		t.Fatal(e)
	}
	b, e := l.AcquireToken(ctx, key, "two", 2, time.Now().Add(40*time.Millisecond))
	if e != nil {
		t.Fatal(e)
	}
	if b <= a {
		t.Fatalf("fence not monotonic: %d <= %d", b, a)
	}
	time.Sleep(60 * time.Millisecond)
	c, e := l.AcquireToken(ctx, key, "three", 3, time.Now().Add(time.Minute))
	if e != nil || c <= b {
		t.Fatalf("expiry takeover fence=%d err=%v", c, e)
	}
	defer l.ReleaseToken(ctx, key, "three", c)
}
