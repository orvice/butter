package application

import (
	"context"

	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

type ForumServiceConnectAdapter struct {
	agentsv1connect.UnimplementedForumServiceHandler
	svc *ForumServiceServer
}

func NewForumServiceConnectAdapter(svc *ForumServiceServer) *ForumServiceConnectAdapter {
	return &ForumServiceConnectAdapter{svc: svc}
}

func (a *ForumServiceConnectAdapter) ListThreads(ctx context.Context, req *connect.Request[agentsv1.ListThreadsRequest]) (*connect.Response[agentsv1.ListThreadsResponse], error) {
	return connectx.WrapUnary(a.svc.ListThreads)(ctx, req)
}

func (a *ForumServiceConnectAdapter) ListThreadLabels(ctx context.Context, req *connect.Request[agentsv1.ListThreadLabelsRequest]) (*connect.Response[agentsv1.ListThreadLabelsResponse], error) {
	return connectx.WrapUnary(a.svc.ListThreadLabels)(ctx, req)
}

func (a *ForumServiceConnectAdapter) GetThread(ctx context.Context, req *connect.Request[agentsv1.GetThreadRequest]) (*connect.Response[agentsv1.GetThreadResponse], error) {
	return connectx.WrapUnary(a.svc.GetThread)(ctx, req)
}

func (a *ForumServiceConnectAdapter) CreateThread(ctx context.Context, req *connect.Request[agentsv1.CreateThreadRequest]) (*connect.Response[agentsv1.CreateThreadResponse], error) {
	return connectx.WrapUnary(a.svc.CreateThread)(ctx, req)
}

func (a *ForumServiceConnectAdapter) UpdateThread(ctx context.Context, req *connect.Request[agentsv1.UpdateThreadRequest]) (*connect.Response[agentsv1.UpdateThreadResponse], error) {
	return connectx.WrapUnary(a.svc.UpdateThread)(ctx, req)
}

func (a *ForumServiceConnectAdapter) DeleteThread(ctx context.Context, req *connect.Request[agentsv1.DeleteThreadRequest]) (*connect.Response[agentsv1.DeleteThreadResponse], error) {
	return connectx.WrapUnary(a.svc.DeleteThread)(ctx, req)
}

func (a *ForumServiceConnectAdapter) CreatePost(ctx context.Context, req *connect.Request[agentsv1.CreatePostRequest]) (*connect.Response[agentsv1.CreatePostResponse], error) {
	return connectx.WrapUnary(a.svc.CreatePost)(ctx, req)
}

func (a *ForumServiceConnectAdapter) DeletePost(ctx context.Context, req *connect.Request[agentsv1.DeletePostRequest]) (*connect.Response[agentsv1.DeletePostResponse], error) {
	return connectx.WrapUnary(a.svc.DeletePost)(ctx, req)
}

func (a *ForumServiceConnectAdapter) InvokeAgentInThread(ctx context.Context, req *connect.Request[agentsv1.InvokeAgentInThreadRequest]) (*connect.Response[agentsv1.InvokeAgentInThreadResponse], error) {
	return connectx.WrapUnary(a.svc.InvokeAgentInThread)(ctx, req)
}
