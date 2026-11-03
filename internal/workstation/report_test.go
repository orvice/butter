package workstation

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestReportCategoryDoesNotExposeRawError(t *testing.T) {
	for _, tc := range []struct {
		err  error
		want string
	}{{fmt.Errorf("%w: unsafe secret body", ErrCleanupPending), "CLEANUP_PENDING"}, {context.DeadlineExceeded, "TIMEOUT"}, {context.Canceled, "CANCELED"}, {ErrConflict, "CONFLICT"}, {ErrFenced, "FENCED"}, {errors.New("Authorization: secret-token, private body"), "INTERNAL"}} {
		if got := ReportCategory(tc.err); got != tc.want {
			t.Fatalf("category %q != %q", got, tc.want)
		}
	}
}

func TestRuntimeErrorPreservesCategoryAndSafeCauseType(t *testing.T) {
	e := &RuntimeError{Operation: "observation_gc", Cause: errors.Join(fmt.Errorf("wrapped secret: %w", context.DeadlineExceeded))}
	if ReportCategory(e) != "TIMEOUT" || CauseType(e) != "context.deadlineExceededError" || e.Error() != "observation_gc failed" {
		t.Fatal("lost safe diagnosis", ReportCategory(e), CauseType(e))
	}
}
