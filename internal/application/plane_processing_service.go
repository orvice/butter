package application

import (
	"connectrpc.com/connect"
	"context"
	"errors"
	planeprocessing "go.orx.me/apps/butter/internal/repo/planeprocessing"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func (s *PlaneAppServiceServer) ListPlaneProcessingRecords(ctx context.Context, req *connect.Request[agentsv1.ListPlaneProcessingRecordsRequest]) (*connect.Response[agentsv1.ListPlaneProcessingRecordsResponse], error) {
	if s.processingRepo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("plane processing storage is not configured"))
	}
	ws, e := requireWorkspace(ctx)
	if e != nil {
		return nil, e
	}
	if req.Msg.GetAppId() != "" {
		if _, e = s.repo.GetApp(ctx, ws, req.Msg.GetAppId()); e != nil {
			return nil, mapPlaneRepoErr(e)
		}
	}
	limit := int(req.Msg.GetPageSize())
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	rows, e := s.processingRepo.List(ctx, planeprocessing.Filter{WorkspaceID: ws, AppID: req.Msg.GetAppId(), Status: req.Msg.GetStatus(), Limit: limit})
	if e != nil {
		return nil, connectx.InternalWith(e)
	}
	return connect.NewResponse(&agentsv1.ListPlaneProcessingRecordsResponse{Records: rows}), nil
}
func (s *PlaneAppServiceServer) GetPlaneProcessingRecord(ctx context.Context, req *connect.Request[agentsv1.GetPlaneProcessingRecordRequest]) (*connect.Response[agentsv1.GetPlaneProcessingRecordResponse], error) {
	if s.processingRepo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("plane processing storage is not configured"))
	}
	ws, e := requireWorkspace(ctx)
	if e != nil {
		return nil, e
	}
	row, e := s.processingRepo.Get(ctx, ws, req.Msg.GetId())
	if errors.Is(e, planeprocessing.ErrNotFound) {
		return nil, connectx.NotFound("plane processing record not found")
	}
	if e != nil {
		return nil, connectx.InternalWith(e)
	}
	return connect.NewResponse(&agentsv1.GetPlaneProcessingRecordResponse{Record: row}), nil
}
