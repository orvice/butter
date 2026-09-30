package application

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	"butterfly.orx.me/core/log"
	"go.orx.me/apps/butter/internal/repo/linearprocessing"
	workspacerepo "go.orx.me/apps/butter/internal/repo/workspace"
	linearruntime "go.orx.me/apps/butter/internal/runtime/linear"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// LinearReplyResender posts a persisted reply again. The Linear
// orchestrator implements it, so a resend goes through the same delivery
// path as the original.
type LinearReplyResender interface {
	Resend(ctx context.Context, workspaceID, recordID string) (*agentsv1.LinearProcessingRecord, error)
}

// LinearProcessingServiceServer implements
// agentsv1connect.LinearProcessingServiceHandler (ADR-0015 §8). Members read
// records; owners and admins resend a reply that was produced but not
// posted. There is deliberately no rerun.
type LinearProcessingServiceServer struct {
	repo          linearprocessing.Repository
	workspaceRepo workspacerepo.Repository
	resender      LinearReplyResender
}

func NewLinearProcessingServiceServer(repo linearprocessing.Repository) *LinearProcessingServiceServer {
	return &LinearProcessingServiceServer{repo: repo}
}

func (s *LinearProcessingServiceServer) SetRepo(repo linearprocessing.Repository) { s.repo = repo }

func (s *LinearProcessingServiceServer) SetWorkspaceRepo(repo workspacerepo.Repository) {
	s.workspaceRepo = repo
}

// SetResender wires the delivery path resends go through.
func (s *LinearProcessingServiceServer) SetResender(resender LinearReplyResender) {
	s.resender = resender
}

func (s *LinearProcessingServiceServer) requireReady() error {
	if s.repo == nil {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("linear processing records are not configured"))
	}
	return nil
}

func mapLinearProcessingErr(err error) *connect.Error {
	switch {
	case errors.Is(err, linearprocessing.ErrNotFound):
		return connectx.NotFound(err.Error())
	case errors.Is(err, linearprocessing.ErrInProgress):
		return connect.NewError(connect.CodeAborted, err)
	case errors.Is(err, linearruntime.ErrNotResendable):
		return connect.NewError(connect.CodeFailedPrecondition, err)
	default:
		return connectx.InternalWith(err)
	}
}

func (s *LinearProcessingServiceServer) ListLinearProcessingRecords(ctx context.Context, req *connect.Request[agentsv1.ListLinearProcessingRecordsRequest]) (*connect.Response[agentsv1.ListLinearProcessingRecordsResponse], error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	workspaceID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	limit := int(req.Msg.GetPageSize())
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	records, err := s.repo.List(ctx, linearprocessing.Filter{
		WorkspaceID: workspaceID,
		AppID:       req.Msg.GetAppId(),
		Status:      req.Msg.GetStatus(),
		Limit:       limit,
	})
	if err != nil {
		return nil, mapLinearProcessingErr(err)
	}
	return connect.NewResponse(&agentsv1.ListLinearProcessingRecordsResponse{Records: records}), nil
}

func (s *LinearProcessingServiceServer) GetLinearProcessingRecord(ctx context.Context, req *connect.Request[agentsv1.GetLinearProcessingRecordRequest]) (*connect.Response[agentsv1.GetLinearProcessingRecordResponse], error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	workspaceID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	record, err := s.repo.Get(ctx, workspaceID, req.Msg.GetId())
	if err != nil {
		return nil, mapLinearProcessingErr(err)
	}
	return connect.NewResponse(&agentsv1.GetLinearProcessingRecordResponse{Record: record}), nil
}

func (s *LinearProcessingServiceServer) ResendLinearReply(ctx context.Context, req *connect.Request[agentsv1.ResendLinearReplyRequest]) (*connect.Response[agentsv1.ResendLinearReplyResponse], error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	workspaceID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireWorkspaceManageRole(ctx, s.workspaceRepo, workspaceID, "linear"); err != nil {
		return nil, err
	}
	if s.resender == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("the Linear runtime is not running"))
	}
	record, err := s.resender.Resend(ctx, workspaceID, req.Msg.GetId())
	if err != nil {
		if record != nil && !errors.Is(err, linearruntime.ErrNotResendable) &&
			!errors.Is(err, linearprocessing.ErrNotFound) && !errors.Is(err, linearprocessing.ErrInProgress) {
			// The post failed again: the record says why, and stays
			// resendable.
			return nil, connect.NewError(connect.CodeUnavailable, err)
		}
		return nil, mapLinearProcessingErr(err)
	}
	log.FromContext(ctx).Info("linear reply resent", "workspace_id", workspaceID, "record_id", record.GetId())
	return connect.NewResponse(&agentsv1.ResendLinearReplyResponse{Record: record}), nil
}
