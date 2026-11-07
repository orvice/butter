package application

import (
	"connectrpc.com/connect"
	"context"
	"errors"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func (s *PlaneAppServiceServer) deliveryReady(ctx context.Context) (string, error) {
	ws, err := requireWorkspace(ctx)
	if err != nil {
		return "", err
	}
	if err := requireWorkspaceManageRole(ctx, s.workspaceRepo, ws, "plane"); err != nil {
		return "", err
	}
	if s.delivery == nil {
		return "", connect.NewError(connect.CodeFailedPrecondition, errors.New("plane delivery management is not configured"))
	}
	return ws, nil
}
func (s *PlaneAppServiceServer) ListPlaneDeliveries(ctx context.Context, _ *connect.Request[agentsv1.ListPlaneDeliveriesRequest]) (*connect.Response[agentsv1.ListPlaneDeliveriesResponse], error) {
	if _, err := s.deliveryReady(ctx); err != nil {
		return nil, err
	}
	rows, err := s.delivery.List(ctx)
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	out := make([]*agentsv1.PlaneDelivery, 0, len(rows))
	for _, r := range rows {
		out = append(out, &agentsv1.PlaneDelivery{Key: r.Key, Status: string(r.Status), Attempts: int32(r.Attempts), LastError: r.LastError, UpdatedAt: r.UpdatedAt, OwnerConfirmed: r.OwnerConfirmed})
	}
	return connect.NewResponse(&agentsv1.ListPlaneDeliveriesResponse{Deliveries: out}), nil
}
func (s *PlaneAppServiceServer) GetPlaneDelivery(ctx context.Context, req *connect.Request[agentsv1.GetPlaneDeliveryRequest]) (*connect.Response[agentsv1.GetPlaneDeliveryResponse], error) {
	if _, err := s.deliveryReady(ctx); err != nil {
		return nil, err
	}
	r, err := s.delivery.Get(ctx, req.Msg.GetKey())
	if err != nil {
		return nil, connectx.NotFound(err.Error())
	}
	return connect.NewResponse(&agentsv1.GetPlaneDeliveryResponse{Delivery: &agentsv1.PlaneDelivery{Key: r.Key, Status: string(r.Status), Attempts: int32(r.Attempts), LastError: r.LastError, UpdatedAt: r.UpdatedAt, OwnerConfirmed: r.OwnerConfirmed}}), nil
}
func (s *PlaneAppServiceServer) RetryPlaneDelivery(ctx context.Context, req *connect.Request[agentsv1.RetryPlaneDeliveryRequest]) (*connect.Response[agentsv1.RetryPlaneDeliveryResponse], error) {
	if _, err := s.deliveryReady(ctx); err != nil {
		return nil, err
	}
	if err := s.delivery.ConfirmRetry(ctx, req.Msg.GetKey(), req.Msg.GetOwner()); err != nil {
		return nil, connect.NewError(connect.CodePermissionDenied, err)
	}
	r, err := s.delivery.Get(ctx, req.Msg.GetKey())
	if err != nil {
		return nil, connectx.NotFound(err.Error())
	}
	return connect.NewResponse(&agentsv1.RetryPlaneDeliveryResponse{Delivery: &agentsv1.PlaneDelivery{Key: r.Key, Status: string(r.Status), Attempts: int32(r.Attempts), LastError: r.LastError, UpdatedAt: r.UpdatedAt, OwnerConfirmed: r.OwnerConfirmed}}), nil
}
