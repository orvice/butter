package application

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func (s *PlaneAppServiceServer) ListPlaneInstallations(context.Context, *connect.Request[agentsv1.ListPlaneInstallationsRequest]) (*connect.Response[agentsv1.ListPlaneInstallationsResponse], error) {
	return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("plane installation storage is not configured"))
}
func (s *PlaneAppServiceServer) GetPlaneInstallation(context.Context, *connect.Request[agentsv1.GetPlaneInstallationRequest]) (*connect.Response[agentsv1.GetPlaneInstallationResponse], error) {
	return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("plane installation storage is not configured"))
}
