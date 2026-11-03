package workstation

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestMemoryBudgetSharesWindowByProviderIdentity(t *testing.T) {
	start := time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC)
	b := NewMemoryBudget(2, time.Minute)
	key := BudgetKey(SourceIdentity{Provider: "github", Origin: "https://api.github.com", TenantID: "org"})
	for i := 0; i < 2; i++ {
		if retry, err := b.Reserve(context.Background(), key, start.Add(time.Duration(i)*time.Second)); err != nil || retry != 0 {
			t.Fatalf("reservation %d: retry=%s err=%v", i, retry, err)
		}
	}
	retry, err := b.Reserve(context.Background(), key, start.Add(10*time.Second))
	if err != nil || retry != 50*time.Second {
		t.Fatalf("expected shared window rejection, retry=%s err=%v", retry, err)
	}
	if retry, err := b.Reserve(context.Background(), key, start.Add(time.Minute)); err != nil || retry != 0 {
		t.Fatalf("window did not reset: retry=%s err=%v", retry, err)
	}
}

func TestMemoryBudgetConcurrentReservationsAreBounded(t *testing.T) {
	b := NewMemoryBudget(1, time.Minute)
	key := BudgetKey(GitHubSource())
	start := time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC)
	var wg sync.WaitGroup
	var mu sync.Mutex
	allowed := 0
	rejected := 0
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			retry, err := b.Reserve(context.Background(), key, start)
			if err != nil {
				t.Errorf("reserve: %v", err)
				return
			}
			mu.Lock()
			defer mu.Unlock()
			if retry == 0 {
				allowed++
			} else {
				rejected++
			}
		}()
	}
	wg.Wait()
	if allowed != 1 || rejected != 15 {
		t.Fatalf("allowed=%d rejected=%d", allowed, rejected)
	}
}
