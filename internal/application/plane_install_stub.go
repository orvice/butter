package application

import (
	"connectrpc.com/connect"
	"context"
	"errors"
	"go.orx.me/apps/butter/internal/repo/planeinstallation"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func (s *PlaneAppServiceServer) ListPlaneInstallations(ctx context.Context, req *connect.Request[agentsv1.ListPlaneInstallationsRequest]) (*connect.Response[agentsv1.ListPlaneInstallationsResponse], error) {
	if s.installRepo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("plane installation storage is not configured"))
	}
	w, e := requireWorkspace(ctx)
	if e != nil {
		return nil, e
	}
	if _, e = s.repo.GetApp(ctx, w, req.Msg.GetAppId()); e != nil {
		return nil, mapPlaneRepoErr(e)
	}
	rows, e := s.installRepo.List(ctx, w, req.Msg.GetAppId())
	if e != nil {
		return nil, connectx.InternalWith(e)
	}
	return connect.NewResponse(&agentsv1.ListPlaneInstallationsResponse{Installations: rows}), nil
}
func (s *PlaneAppServiceServer) GetPlaneInstallation(ctx context.Context, req *connect.Request[agentsv1.GetPlaneInstallationRequest]) (*connect.Response[agentsv1.GetPlaneInstallationResponse], error) {
	if s.installRepo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("plane installation storage is not configured"))
	}
	w, e := requireWorkspace(ctx)
	if e != nil {
		return nil, e
	}
	row, e := s.installRepo.Get(ctx, w, req.Msg.GetId())
	if e != nil {
		if errors.Is(e, planeinstallation.ErrNotFound) {
			return nil, connectx.NotFound(e.Error())
		}
		return nil, connectx.InternalWith(e)
	}
	return connect.NewResponse(&agentsv1.GetPlaneInstallationResponse{Installation: row}), nil
}
