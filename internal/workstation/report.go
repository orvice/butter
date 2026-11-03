package workstation

import (
	"context"
	"errors"
	"fmt"
)

func ReportCategory(err error) string {
	switch {
	case errors.Is(err, ErrCleanupPending):
		return "CLEANUP_PENDING"
	case errors.Is(err, context.DeadlineExceeded):
		return "TIMEOUT"
	case errors.Is(err, context.Canceled):
		return "CANCELED"
	case errors.Is(err, ErrConflict):
		return "CONFLICT"
	case errors.Is(err, ErrFenced):
		return "FENCED"
	default:
		return "INTERNAL"
	}
}

// RuntimeError names the failed boundary without embedding unsafe driver text.
type RuntimeError struct {
	Operation string
	Cause     error
}

func (e *RuntimeError) Error() string { return e.Operation + " failed" }
func (e *RuntimeError) Unwrap() error { return e.Cause }
func CauseType(err error) string {
	for err != nil {
		if joined, ok := err.(interface{ Unwrap() []error }); ok && len(joined.Unwrap()) > 0 {
			err = joined.Unwrap()[0]
			continue
		}
		if next := errors.Unwrap(err); next != nil {
			err = next
			continue
		}
		return fmt.Sprintf("%T", err)
	}
	return "none"
}
