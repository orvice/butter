package application

import (
	"context"

	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

type ChannelServiceConnectAdapter struct {
	agentsv1connect.UnimplementedChannelServiceHandler
	svc *ChannelServiceServer
}

func NewChannelServiceConnectAdapter(svc *ChannelServiceServer) *ChannelServiceConnectAdapter {
	return &ChannelServiceConnectAdapter{svc: svc}
}

func (a *ChannelServiceConnectAdapter) ListChannels(ctx context.Context, req *connect.Request[agentsv1.ListChannelsRequest]) (*connect.Response[agentsv1.ListChannelsResponse], error) {
	return connectx.WrapUnary(a.svc.ListChannels)(ctx, req)
}

func (a *ChannelServiceConnectAdapter) GetChannel(ctx context.Context, req *connect.Request[agentsv1.GetChannelRequest]) (*connect.Response[agentsv1.GetChannelResponse], error) {
	return connectx.WrapUnary(a.svc.GetChannel)(ctx, req)
}

func (a *ChannelServiceConnectAdapter) CreateChannel(ctx context.Context, req *connect.Request[agentsv1.CreateChannelRequest]) (*connect.Response[agentsv1.CreateChannelResponse], error) {
	return connectx.WrapUnary(a.svc.CreateChannel)(ctx, req)
}

func (a *ChannelServiceConnectAdapter) UpdateChannel(ctx context.Context, req *connect.Request[agentsv1.UpdateChannelRequest]) (*connect.Response[agentsv1.UpdateChannelResponse], error) {
	return connectx.WrapUnary(a.svc.UpdateChannel)(ctx, req)
}

func (a *ChannelServiceConnectAdapter) DeleteChannel(ctx context.Context, req *connect.Request[agentsv1.DeleteChannelRequest]) (*connect.Response[agentsv1.DeleteChannelResponse], error) {
	return connectx.WrapUnary(a.svc.DeleteChannel)(ctx, req)
}

func (a *ChannelServiceConnectAdapter) GetChannelStatus(ctx context.Context, req *connect.Request[agentsv1.GetChannelStatusRequest]) (*connect.Response[agentsv1.GetChannelStatusResponse], error) {
	return connectx.WrapUnary(a.svc.GetChannelStatus)(ctx, req)
}

func (a *ChannelServiceConnectAdapter) RestartChannel(ctx context.Context, req *connect.Request[agentsv1.RestartChannelRequest]) (*connect.Response[agentsv1.RestartChannelResponse], error) {
	return connectx.WrapUnary(a.svc.RestartChannel)(ctx, req)
}

func (a *ChannelServiceConnectAdapter) PauseChannel(ctx context.Context, req *connect.Request[agentsv1.PauseChannelRequest]) (*connect.Response[agentsv1.PauseChannelResponse], error) {
	return connectx.WrapUnary(a.svc.PauseChannel)(ctx, req)
}

func (a *ChannelServiceConnectAdapter) ResumeChannel(ctx context.Context, req *connect.Request[agentsv1.ResumeChannelRequest]) (*connect.Response[agentsv1.ResumeChannelResponse], error) {
	return connectx.WrapUnary(a.svc.ResumeChannel)(ctx, req)
}
