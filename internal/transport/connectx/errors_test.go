package connectx

import (
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/twitchtv/twirp"
)

func TestTwirpErrorToConnect_CodeMapping(t *testing.T) {
	cases := []struct {
		twirpCode twirp.ErrorCode
		want      connect.Code
	}{
		{twirp.Canceled, connect.CodeCanceled},
		{twirp.InvalidArgument, connect.CodeInvalidArgument},
		{twirp.Malformed, connect.CodeInvalidArgument},
		{twirp.DeadlineExceeded, connect.CodeDeadlineExceeded},
		{twirp.NotFound, connect.CodeNotFound},
		{twirp.BadRoute, connect.CodeUnimplemented},
		{twirp.AlreadyExists, connect.CodeAlreadyExists},
		{twirp.PermissionDenied, connect.CodePermissionDenied},
		{twirp.Unauthenticated, connect.CodeUnauthenticated},
		{twirp.ResourceExhausted, connect.CodeResourceExhausted},
		{twirp.FailedPrecondition, connect.CodeFailedPrecondition},
		{twirp.Aborted, connect.CodeAborted},
		{twirp.OutOfRange, connect.CodeOutOfRange},
		{twirp.Unimplemented, connect.CodeUnimplemented},
		{twirp.Internal, connect.CodeInternal},
		{twirp.Unavailable, connect.CodeUnavailable},
		{twirp.DataLoss, connect.CodeDataLoss},
		{twirp.Unknown, connect.CodeUnknown},
	}
	for _, tc := range cases {
		t.Run(string(tc.twirpCode), func(t *testing.T) {
			in := twirp.NewError(tc.twirpCode, "boom")
			got := TwirpErrorToConnect(in)
			if got == nil {
				t.Fatalf("nil result for %v", tc.twirpCode)
			}
			if got.Code() != tc.want {
				t.Fatalf("code: got %v want %v", got.Code(), tc.want)
			}
			if got.Message() != "boom" {
				t.Fatalf("msg: got %q want %q", got.Message(), "boom")
			}
		})
	}
}

func TestTwirpErrorToConnect_NilInput(t *testing.T) {
	if got := TwirpErrorToConnect(nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestTwirpErrorToConnect_NonTwirpFallsBackToInternal(t *testing.T) {
	got := TwirpErrorToConnect(errors.New("plain error"))
	if got == nil {
		t.Fatal("expected non-nil connect error")
	}
	if got.Code() != connect.CodeInternal {
		t.Fatalf("code: got %v want CodeInternal", got.Code())
	}
	if got.Message() != "plain error" {
		t.Fatalf("msg: got %q", got.Message())
	}
}

func TestTwirpErrorToConnect_PreservesMeta(t *testing.T) {
	in := twirp.NewError(twirp.InvalidArgument, "missing").WithMeta("argument", "username")
	got := TwirpErrorToConnect(in)
	if got == nil {
		t.Fatal("expected non-nil")
	}
	if v := got.Meta().Get("argument"); v != "username" {
		t.Fatalf("meta argument: got %q want username", v)
	}
}
