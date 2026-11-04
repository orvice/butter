package workstation

import (
	"errors"
	"testing"
	"time"
)

func TestPlanCleanupRetryBackoffAndCap(t *testing.T) {
	now := time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC)
	policy := CleanupPolicy{BaseDelay: time.Second, MaxDelay: 4 * time.Second}
	task := CleanupTask{Key: "k", Revision: 3}
	for attempt, want := range []time.Duration{time.Second, 2 * time.Second, 4 * time.Second, 4 * time.Second} {
		var result CleanupResult
		task, result = PlanCleanupRetry(task, now, errors.New("temporary storage error"), policy)
		if result != CleanupRetry || task.Attempts != attempt+1 || !task.NextAt.Equal(now.Add(want)) {
			t.Fatalf("attempt %d: task=%+v result=%d", attempt+1, task, result)
		}
	}
}

func TestPlanCleanupRetryDeadlineAndConflictAreTerminal(t *testing.T) {
	now := time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC)
	policy := CleanupPolicy{BaseDelay: time.Minute, MaxDelay: time.Hour}
	task := CleanupTask{Key: "k", Revision: 2, Deadline: now.Add(30 * time.Second)}
	if _, result := PlanCleanupRetry(task, now, errors.New("temporary"), policy); result != CleanupExpired {
		t.Fatal("retry past deadline was scheduled")
	}
	if _, result := PlanCleanupRetry(task, now, ErrCheckpointConflict, policy); result != CleanupDone {
		t.Fatal("revision conflict was retried")
	}
	if _, result := PlanCleanupRetry(task, now, nil, policy); result != CleanupDone {
		t.Fatal("successful cleanup was not terminal")
	}
}
