package connectx

import (
	"errors"
	"testing"

	"connectrpc.com/connect"
)

func TestRequiredArgument(t *testing.T) {
	got := RequiredArgument("username")
	if got.Code() != connect.CodeInvalidArgument {
		t.Fatalf("code: got %v want CodeInvalidArgument", got.Code())
	}
	if got.Message() != "username is required" {
		t.Fatalf("message: got %q", got.Message())
	}
	if v := got.Meta().Get("argument"); v != "username" {
		t.Fatalf("argument meta: got %q want username", v)
	}
}

func TestInvalidArgument(t *testing.T) {
	got := InvalidArgument("page_size", "must be positive")
	if got.Code() != connect.CodeInvalidArgument {
		t.Fatalf("code: got %v", got.Code())
	}
	if got.Message() != "page_size must be positive" {
		t.Fatalf("message: got %q", got.Message())
	}
}

func TestNotFound(t *testing.T) {
	got := NotFound("agent not found")
	if got.Code() != connect.CodeNotFound {
		t.Fatalf("code: got %v", got.Code())
	}
	if got.Message() != "agent not found" {
		t.Fatalf("message: got %q", got.Message())
	}
}

func TestInternal(t *testing.T) {
	got := Internal("boom")
	if got.Code() != connect.CodeInternal {
		t.Fatalf("code: got %v", got.Code())
	}
}

func TestInternalWith_PreservesUnderlying(t *testing.T) {
	underlying := errors.New("disk full")
	got := InternalWith(underlying)
	if got.Code() != connect.CodeInternal {
		t.Fatalf("code: got %v", got.Code())
	}
	if got.Message() != "disk full" {
		t.Fatalf("message: got %q", got.Message())
	}
	if !errors.Is(got, underlying) {
		t.Fatal("expected wrapped error to be reachable via errors.Is")
	}
}
