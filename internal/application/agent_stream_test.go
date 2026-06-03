package application

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
)

func TestStreamAgentError(t *testing.T) {
	t.Run("preserves connect errors", func(t *testing.T) {
		want := connect.NewError(connect.CodeInvalidArgument, errors.New("bad request"))

		got := streamAgentError(want)
		if got != want {
			t.Fatalf("expected original connect error, got %v", got)
		}
	})

	t.Run("maps context cancellation", func(t *testing.T) {
		got := streamAgentError(context.Canceled)

		twerr, ok := got.(*connect.Error)
		if !ok {
			t.Fatalf("expected connect error, got %T", got)
		}
		if twerr.Code() != connect.CodeCanceled {
			t.Fatalf("expected CodeCanceled, got %s", twerr.Code())
		}
	})

	t.Run("defaults to internal", func(t *testing.T) {
		got := streamAgentError(errors.New("runner failed"))

		twerr, ok := got.(*connect.Error)
		if !ok {
			t.Fatalf("expected connect error, got %T", got)
		}
		if twerr.Code() != connect.CodeInternal {
			t.Fatalf("expected CodeInternal, got %s", twerr.Code())
		}
	})
}
