package workstation

import (
	"errors"
	"time"
)

type CleanupResult uint8

const (
	CleanupDone CleanupResult = iota
	CleanupRetry
	CleanupExpired
)

type CleanupTask struct {
	Key      string
	Revision uint64
	Attempts int
	NextAt   time.Time
	Deadline time.Time
}

type CleanupPolicy struct {
	BaseDelay time.Duration
	MaxDelay  time.Duration
}

// PlanCleanupRetry turns one fenced delete result into a durable retry plan.
// A revision conflict is terminal: another writer owns a newer checkpoint, so
// retrying the old cleanup could only be unsafe. Other errors retry with capped
// exponential backoff until the task deadline.
func PlanCleanupRetry(task CleanupTask, now time.Time, err error, policy CleanupPolicy) (CleanupTask, CleanupResult) {
	if err == nil || errors.Is(err, ErrCheckpointConflict) {
		task.NextAt = time.Time{}
		return task, CleanupDone
	}
	if !task.Deadline.IsZero() && !now.Before(task.Deadline) {
		return task, CleanupExpired
	}
	if policy.BaseDelay <= 0 {
		policy.BaseDelay = time.Second
	}
	if policy.MaxDelay < policy.BaseDelay {
		policy.MaxDelay = policy.BaseDelay
	}
	attempt := task.Attempts + 1
	delay := policy.BaseDelay
	for i := 1; i < attempt && delay < policy.MaxDelay; i++ {
		if delay > policy.MaxDelay/2 {
			delay = policy.MaxDelay
			break
		}
		delay *= 2
	}
	next := now.Add(delay)
	if !task.Deadline.IsZero() && next.After(task.Deadline) {
		return task, CleanupExpired
	}
	task.Attempts = attempt
	task.NextAt = next
	return task, CleanupRetry
}
