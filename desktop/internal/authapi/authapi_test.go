package authapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type fakeAuth struct {
	agentsv1connect.UnimplementedAuthServiceHandler
	login  func(context.Context, *agentsv1.LoginRequest) (*agentsv1.LoginResponse, error)
	me     func(context.Context, string) (*agentsv1.MeResponse, error)
	logout func(context.Context, string) error
}

func (f fakeAuth) Login(ctx context.Context, req *connect.Request[agentsv1.LoginRequest]) (*connect.Response[agentsv1.LoginResponse], error) {
	res, err := f.login(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(res), nil
}

func (f fakeAuth) Me(ctx context.Context, req *connect.Request[agentsv1.MeRequest]) (*connect.Response[agentsv1.MeResponse], error) {
	res, err := f.me(ctx, req.Header().Get(TokenHeader))
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(res), nil
}

func (f fakeAuth) Logout(ctx context.Context, req *connect.Request[agentsv1.LogoutRequest]) (*connect.Response[agentsv1.LogoutResponse], error) {
	if f.logout != nil {
		if err := f.logout(ctx, req.Header().Get(TokenHeader)); err != nil {
			return nil, err
		}
	}
	return connect.NewResponse(&agentsv1.LogoutResponse{}), nil
}

func serve(t *testing.T, h agentsv1connect.AuthServiceHandler) string {
	t.Helper()
	_, handler := agentsv1connect.NewAuthServiceHandler(h)
	mux := http.NewServeMux()
	mux.Handle("/agents.v1.AuthService/", handler)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv.URL
}

func TestLoginSuccess(t *testing.T) {
	expires := time.Now().Add(7 * 24 * time.Hour).UTC().Truncate(time.Second)
	base := serve(t, fakeAuth{
		login: func(_ context.Context, req *agentsv1.LoginRequest) (*agentsv1.LoginResponse, error) {
			if req.GetUsername() != "lion" || req.GetPassword() != "secret" {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid username or password"))
			}
			return &agentsv1.LoginResponse{
				Token:     "tok-1",
				User:      &agentsv1.User{Id: "u1", Username: "lion", DisplayName: "Lion", Role: "user"},
				ExpiresAt: timestamppb.New(expires),
				Workspaces: []*agentsv1.Workspace{
					{Id: "w1", Name: "Default", Slug: "default"},
				},
			}, nil
		},
	})
	c := New(base, nil)
	s, err := c.Login(context.Background(), "lion", "secret")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if s.Token != "tok-1" || s.User.Username != "lion" || s.User.DisplayName != "Lion" {
		t.Fatalf("unexpected session: %+v", s)
	}
	if len(s.Workspaces) != 1 || s.Workspaces[0].ID != "w1" {
		t.Fatalf("unexpected workspaces: %+v", s.Workspaces)
	}
	if !s.ExpiresAt.Equal(expires) {
		t.Fatalf("expires = %v, want %v", s.ExpiresAt, expires)
	}
}

func TestLoginFailures(t *testing.T) {
	cases := []struct {
		name string
		err  error
		kind Kind
		msg  string
	}{
		{
			name: "wrong password",
			err:  connect.NewError(connect.CodeUnauthenticated, errors.New("invalid username or password")),
			kind: KindInvalidCredentials,
			msg:  "用户名或密码错误",
		},
		{
			name: "disabled user",
			err:  connect.NewError(connect.CodePermissionDenied, errors.New("user disabled")),
			kind: KindUserDisabled,
			msg:  "该账号已被禁用，请联系管理员",
		},
		{
			name: "missing field",
			err:  connect.NewError(connect.CodeInvalidArgument, errors.New("username: required")),
			kind: KindInvalidInput,
		},
		{
			name: "no auth store",
			err:  connect.NewError(connect.CodeFailedPrecondition, errors.New("auth store not available")),
			kind: KindStoreUnavailable,
		},
		{
			name: "server bug",
			err:  connect.NewError(connect.CodeInternal, errors.New("boom")),
			kind: KindServer,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			base := serve(t, fakeAuth{
				login: func(context.Context, *agentsv1.LoginRequest) (*agentsv1.LoginResponse, error) {
					return nil, tc.err
				},
			})
			_, err := New(base, nil).Login(context.Background(), "lion", "x")
			if KindOf(err) != tc.kind {
				t.Fatalf("kind = %v, want %v (err %v)", KindOf(err), tc.kind, err)
			}
			var apiErr *Error
			if !errors.As(err, &apiErr) {
				t.Fatalf("error is not *Error: %T", err)
			}
			if apiErr.Message() == "" {
				t.Fatal("empty user message")
			}
			if tc.msg != "" && apiErr.Message() != tc.msg {
				t.Fatalf("message = %q, want %q", apiErr.Message(), tc.msg)
			}
		})
	}
}

func TestMeExpiredSession(t *testing.T) {
	base := serve(t, fakeAuth{
		login: func(context.Context, *agentsv1.LoginRequest) (*agentsv1.LoginResponse, error) {
			return &agentsv1.LoginResponse{Token: "t"}, nil
		},
		me: func(context.Context, string) (*agentsv1.MeResponse, error) {
			return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid session"))
		},
	})
	_, err := New(base, nil).Me(context.Background(), "stale")
	if !Is(err, KindSessionExpired) {
		t.Fatalf("kind = %v, want session expired", KindOf(err))
	}
	if KindOf(err) == KindInvalidCredentials {
		t.Fatal("an expired session is not a wrong password")
	}
}

func TestMeSendsBearerToken(t *testing.T) {
	var got string
	base := serve(t, fakeAuth{
		login: func(context.Context, *agentsv1.LoginRequest) (*agentsv1.LoginResponse, error) {
			return &agentsv1.LoginResponse{}, nil
		},
		me: func(_ context.Context, h string) (*agentsv1.MeResponse, error) {
			got = h
			return &agentsv1.MeResponse{User: &agentsv1.User{Id: "u1", Username: "lion"}}, nil
		},
	})
	u, err := New(base, nil).Me(context.Background(), "abc")
	if err != nil {
		t.Fatalf("Me: %v", err)
	}
	if got != "Bearer abc" {
		t.Fatalf("Authorization = %q", got)
	}
	if u.Username != "lion" {
		t.Fatalf("user = %+v", u)
	}
}

func TestNetworkFailure(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	base := srv.URL
	srv.Close()
	_, err := New(base, nil).Me(context.Background(), "abc")
	if !Is(err, KindNetwork) {
		t.Fatalf("kind = %v, want network (err %v)", KindOf(err), err)
	}
}

func TestLogoutToleratesExpiredToken(t *testing.T) {
	base := serve(t, fakeAuth{
		login: func(context.Context, *agentsv1.LoginRequest) (*agentsv1.LoginResponse, error) {
			return &agentsv1.LoginResponse{}, nil
		},
		logout: func(context.Context, string) error {
			return connect.NewError(connect.CodeUnauthenticated, errors.New("invalid session"))
		},
	})
	if err := New(base, nil).Logout(context.Background(), "gone"); err != nil {
		t.Fatalf("Logout: %v", err)
	}
}
