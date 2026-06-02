package application

import (
	"context"

	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

// AuthServiceConnectAdapter implements agentsv1connect.AuthServiceHandler by
// delegating to the existing Twirp-shaped AuthServiceServer. It lets the
// migration flip the wire protocol service-by-service without rewriting the
// 12 method bodies first.
type AuthServiceConnectAdapter struct {
	agentsv1connect.UnimplementedAuthServiceHandler
	svc *AuthServiceServer
}

func NewAuthServiceConnectAdapter(svc *AuthServiceServer) *AuthServiceConnectAdapter {
	return &AuthServiceConnectAdapter{svc: svc}
}

func (a *AuthServiceConnectAdapter) Login(ctx context.Context, req *connect.Request[agentsv1.LoginRequest]) (*connect.Response[agentsv1.LoginResponse], error) {
	return connectx.WrapUnary(a.svc.Login)(ctx, req)
}

func (a *AuthServiceConnectAdapter) Me(ctx context.Context, req *connect.Request[agentsv1.MeRequest]) (*connect.Response[agentsv1.MeResponse], error) {
	return connectx.WrapUnary(a.svc.Me)(ctx, req)
}

func (a *AuthServiceConnectAdapter) Logout(ctx context.Context, req *connect.Request[agentsv1.LogoutRequest]) (*connect.Response[agentsv1.LogoutResponse], error) {
	return connectx.WrapUnary(a.svc.Logout)(ctx, req)
}

func (a *AuthServiceConnectAdapter) ListUsers(ctx context.Context, req *connect.Request[agentsv1.ListUsersRequest]) (*connect.Response[agentsv1.ListUsersResponse], error) {
	return connectx.WrapUnary(a.svc.ListUsers)(ctx, req)
}

func (a *AuthServiceConnectAdapter) CreateUser(ctx context.Context, req *connect.Request[agentsv1.CreateUserRequest]) (*connect.Response[agentsv1.CreateUserResponse], error) {
	return connectx.WrapUnary(a.svc.CreateUser)(ctx, req)
}

func (a *AuthServiceConnectAdapter) UpdateUserPassword(ctx context.Context, req *connect.Request[agentsv1.UpdateUserPasswordRequest]) (*connect.Response[agentsv1.UpdateUserPasswordResponse], error) {
	return connectx.WrapUnary(a.svc.UpdateUserPassword)(ctx, req)
}

func (a *AuthServiceConnectAdapter) SetUserDisabled(ctx context.Context, req *connect.Request[agentsv1.SetUserDisabledRequest]) (*connect.Response[agentsv1.SetUserDisabledResponse], error) {
	return connectx.WrapUnary(a.svc.SetUserDisabled)(ctx, req)
}

func (a *AuthServiceConnectAdapter) UpdateProfile(ctx context.Context, req *connect.Request[agentsv1.UpdateProfileRequest]) (*connect.Response[agentsv1.UpdateProfileResponse], error) {
	return connectx.WrapUnary(a.svc.UpdateProfile)(ctx, req)
}

func (a *AuthServiceConnectAdapter) ChangePassword(ctx context.Context, req *connect.Request[agentsv1.ChangePasswordRequest]) (*connect.Response[agentsv1.ChangePasswordResponse], error) {
	return connectx.WrapUnary(a.svc.ChangePassword)(ctx, req)
}

func (a *AuthServiceConnectAdapter) ListOAuthProviders(ctx context.Context, req *connect.Request[agentsv1.ListOAuthProvidersRequest]) (*connect.Response[agentsv1.ListOAuthProvidersResponse], error) {
	return connectx.WrapUnary(a.svc.ListOAuthProviders)(ctx, req)
}

func (a *AuthServiceConnectAdapter) BeginOAuthFlow(ctx context.Context, req *connect.Request[agentsv1.BeginOAuthFlowRequest]) (*connect.Response[agentsv1.BeginOAuthFlowResponse], error) {
	return connectx.WrapUnary(a.svc.BeginOAuthFlow)(ctx, req)
}

func (a *AuthServiceConnectAdapter) CompleteOAuthFlow(ctx context.Context, req *connect.Request[agentsv1.CompleteOAuthFlowRequest]) (*connect.Response[agentsv1.CompleteOAuthFlowResponse], error) {
	return connectx.WrapUnary(a.svc.CompleteOAuthFlow)(ctx, req)
}
