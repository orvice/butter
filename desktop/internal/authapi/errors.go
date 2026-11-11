package authapi

import (
	"context"
	"errors"
	"fmt"
	"net"

	"connectrpc.com/connect"
)

// Kind is what went wrong, in the terms the interface acts on.
type Kind uint8

const (
	// KindUnknown is a failure the client does not classify.
	KindUnknown Kind = iota
	// KindInvalidCredentials is a rejected username or password.
	KindInvalidCredentials
	// KindUserDisabled is an account the server disabled.
	KindUserDisabled
	// KindSessionExpired is a token the server no longer accepts.
	KindSessionExpired
	// KindInvalidInput is a request the server refused as incomplete.
	KindInvalidInput
	// KindStoreUnavailable is a server without a usable auth store.
	KindStoreUnavailable
	// KindForbidden is a signed-in account without access to the resource.
	KindForbidden
	// KindNetwork is a server that could not be reached.
	KindNetwork
	// KindServer is a server-side failure.
	KindServer
	// KindCanceled is a request the caller abandoned.
	KindCanceled
)

// Error is a classified failure of one call.
type Error struct {
	Kind Kind
	Op   string
	Err  error
}

func (e *Error) Error() string {
	if e.Op == "" {
		return e.Err.Error()
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *Error) Unwrap() error { return e.Err }

// Message is the text to show the user: no server internals, no token.
func (e *Error) Message() string {
	switch e.Kind {
	case KindInvalidCredentials:
		return "用户名或密码错误"
	case KindUserDisabled:
		return "该账号已被禁用，请联系管理员"
	case KindSessionExpired:
		return "登录已过期，请重新登录"
	case KindInvalidInput:
		return "请填写服务器地址、用户名和密码"
	case KindStoreUnavailable:
		return "服务端认证存储不可用，请稍后再试"
	case KindForbidden:
		return "该账号无权访问这个资源"
	case KindNetwork:
		return "无法连接服务器，请检查地址与网络"
	case KindServer:
		return "服务器返回错误，请稍后再试"
	case KindCanceled:
		return "已取消"
	default:
		return "操作失败，请重试"
	}
}

// KindOf returns the kind of err, or KindUnknown.
func KindOf(err error) Kind {
	var e *Error
	if errors.As(err, &e) {
		return e.Kind
	}
	return KindUnknown
}

// Is reports whether err was classified as kind.
func Is(err error, kind Kind) bool { return KindOf(err) == kind }

type authMode uint8

const (
	// callUnauthenticated reads CodeUnauthenticated as a rejected password:
	// Login is the call that has no token yet.
	callUnauthenticated authMode = iota
	// callExpired reads CodeUnauthenticated as a session the server no
	// longer accepts.
	callExpired
)

func classify(op string, err error, mode authMode) error {
	if err == nil {
		return nil
	}
	kind := KindUnknown
	switch connect.CodeOf(err) {
	case connect.CodeUnauthenticated:
		if mode == callUnauthenticated {
			kind = KindInvalidCredentials
		} else {
			kind = KindSessionExpired
		}
	case connect.CodePermissionDenied:
		if mode == callUnauthenticated {
			kind = KindUserDisabled
		} else {
			kind = KindForbidden
		}
	case connect.CodeInvalidArgument:
		kind = KindInvalidInput
	case connect.CodeFailedPrecondition:
		kind = KindStoreUnavailable
	case connect.CodeUnavailable, connect.CodeDeadlineExceeded:
		kind = KindNetwork
	case connect.CodeCanceled:
		kind = KindCanceled
	case connect.CodeInternal, connect.CodeResourceExhausted, connect.CodeAborted,
		connect.CodeNotFound, connect.CodeAlreadyExists, connect.CodeOutOfRange,
		connect.CodeUnimplemented, connect.CodeDataLoss:
		kind = KindServer
	}
	// A context that ended is the caller's, never the server's, even when
	// the transport reports it as a plain error.
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return &Error{Kind: KindCanceled, Op: op, Err: err}
	}
	// A transport failure can surface as CodeUnknown; only call it network
	// trouble when it really is one.
	if kind == KindNetwork && !isTransport(err) {
		kind = KindServer
	}
	return &Error{Kind: kind, Op: op, Err: err}
}

func isTransport(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}
	var dnsErr *net.DNSError
	return errors.As(err, &dnsErr)
}
