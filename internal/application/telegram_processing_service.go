package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"butterfly.orx.me/core/log"
	"go.orx.me/apps/butter/internal/repo/telegramprocessing"
	workspacerepo "go.orx.me/apps/butter/internal/repo/workspace"
	telegramruntime "go.orx.me/apps/butter/internal/runtime/telegram"
	"go.orx.me/apps/butter/internal/telegramsend"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// defaultProcessingPageSize and maxProcessingPageSize bound a listing.
const (
	defaultProcessingPageSize = 50
	maxProcessingPageSize     = 200
)

// TelegramProcessingServiceServer implements
// agentsv1connect.TelegramProcessingServiceHandler (issue #264).
//
// It exposes exactly one recovery action: resending a reply that was already
// produced. There is deliberately no Agent rerun. Once Agent work may have
// produced external side effects, repeating it is a judgement call that needs
// context this API does not have — so the record says the work is uncertain
// and stops, rather than guessing on the operator's behalf.
type TelegramProcessingServiceServer struct {
	repo          telegramprocessing.Repository
	workspaceRepo workspacerepo.Repository
	sender        *telegramsend.Sender
}

func NewTelegramProcessingServiceServer(repo telegramprocessing.Repository) *TelegramProcessingServiceServer {
	return &TelegramProcessingServiceServer{repo: repo}
}

func (s *TelegramProcessingServiceServer) SetRepo(repo telegramprocessing.Repository) {
	s.repo = repo
}

func (s *TelegramProcessingServiceServer) SetWorkspaceRepo(repo workspacerepo.Repository) {
	s.workspaceRepo = repo
}

func (s *TelegramProcessingServiceServer) SetSender(sender *telegramsend.Sender) {
	s.sender = sender
}

func (s *TelegramProcessingServiceServer) requireReady() error {
	if s.repo == nil {
		return connect.NewError(connect.CodeFailedPrecondition,
			errors.New("telegram processing repository not configured"))
	}
	return nil
}

func (s *TelegramProcessingServiceServer) ListTelegramProcessingRecords(ctx context.Context, req *connect.Request[agentsv1.ListTelegramProcessingRecordsRequest]) (*connect.Response[agentsv1.ListTelegramProcessingRecordsResponse], error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	workspaceID, err := telegramWorkspace(ctx)
	if err != nil {
		return nil, err
	}

	limit := int(req.Msg.GetPageSize())
	if limit <= 0 {
		limit = defaultProcessingPageSize
	}
	limit = min(limit, maxProcessingPageSize)

	records, err := s.repo.List(ctx, telegramprocessing.Filter{
		WorkspaceID:   workspaceID,
		ChannelID:     strings.TrimSpace(req.Msg.GetChannelId()),
		DestinationID: strings.TrimSpace(req.Msg.GetDestinationId()),
		Status:        req.Msg.GetStatus(),
		Limit:         limit,
	})
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	return connect.NewResponse(&agentsv1.ListTelegramProcessingRecordsResponse{Records: records}), nil
}

func (s *TelegramProcessingServiceServer) GetTelegramProcessingRecord(ctx context.Context, req *connect.Request[agentsv1.GetTelegramProcessingRecordRequest]) (*connect.Response[agentsv1.GetTelegramProcessingRecordResponse], error) {
	record, err := s.load(ctx, req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&agentsv1.GetTelegramProcessingRecordResponse{Record: record}), nil
}

func (s *TelegramProcessingServiceServer) load(ctx context.Context, id string) (*agentsv1.TelegramProcessingRecord, error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(id) == "" {
		return nil, connectx.RequiredArgument("id")
	}
	workspaceID, err := telegramWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	record, err := s.repo.Get(ctx, workspaceID, id)
	if err != nil {
		if errors.Is(err, telegramprocessing.ErrNotFound) {
			return nil, connectx.NotFound(err.Error())
		}
		return nil, connectx.InternalWith(err)
	}
	return record, nil
}

// ResendTelegramReply re-delivers the segments that never landed.
//
// It is offered only when a complete response exists, and it continues from
// unsent and failed segments rather than restarting — so an operator
// recovering a partial delivery does not double-post the half that worked.
func (s *TelegramProcessingServiceServer) ResendTelegramReply(ctx context.Context, req *connect.Request[agentsv1.ResendTelegramReplyRequest]) (*connect.Response[agentsv1.ResendTelegramReplyResponse], error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	workspaceID, err := telegramWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireTelegramManageRole(ctx, s.workspaceRepo, workspaceID, "resend_telegram_reply"); err != nil {
		return nil, err
	}
	if s.sender == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			errors.New("telegram sender is not configured"))
	}
	if strings.TrimSpace(req.Msg.GetId()) == "" {
		return nil, connectx.RequiredArgument("id")
	}
	now := time.Now().UTC()
	leaseToken := uuid.NewString()
	record, err := s.repo.ClaimDelivery(ctx, workspaceID, req.Msg.GetId(), leaseToken, now, now.Add(5*time.Minute))
	if err != nil {
		switch {
		case errors.Is(err, telegramprocessing.ErrNotFound):
			return nil, connectx.NotFound(err.Error())
		case errors.Is(err, telegramprocessing.ErrInProgress):
			return nil, connect.NewError(connect.CodeAborted, errors.New("this reply is already being delivered"))
		default:
			return nil, connectx.InternalWith(err)
		}
	}
	defer func() {
		releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = s.repo.ReleaseClaim(releaseCtx, workspaceID, record.GetId(), leaseToken)
	}()

	if len(record.GetSegments()) == 0 || strings.TrimSpace(record.GetOutput()) == "" {
		// Without persisted output there is nothing to resend, and producing
		// it would mean re-running the Agent — which this action never does.
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			errors.New("this record has no complete response to resend"))
	}

	delivery := telegramruntime.DeliveryFromRecord(record)
	if !delivery.Pending() {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			errors.New("every segment of this response was already delivered"))
	}

	deliverErr := s.sender.DeliverSegments(ctx, workspaceID, record.GetDestinationId(), delivery,
		func(current *telegramsend.Delivery) error {
			record.Segments = telegramruntime.SegmentsToProto(current)
			_, progressErr := s.repo.UpdateClaimed(ctx, record, leaseToken)
			return progressErr
		})
	record.Segments = telegramruntime.SegmentsToProto(delivery)
	if deliverErr == nil {
		record.Status = agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_SUCCEEDED
		record.Error = ""
		record.DeadLettered = false
	} else if errors.Is(deliverErr, telegramsend.ErrDeliveryUncertain) || hasUncertainTelegramSegment(delivery) {
		record.Status = agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_FAILED_UNCERTAIN
		record.Error = deliverErr.Error()
		record.DeadLettered = true
	} else {
		record.Status = agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_FAILED
		record.Error = deliverErr.Error()
	}
	updated, updateErr := s.repo.UpdateClaimed(ctx, record, leaseToken)
	if updateErr != nil {
		return nil, connectx.InternalWith(updateErr)
	}
	if deliverErr != nil {
		return nil, mapTelegramSendErr(deliverErr)
	}

	log.FromContext(ctx).Info("telegram reply resent",
		"workspace_id", workspaceID, "record_id", record.GetId(),
		"destination_id", record.GetDestinationId())
	return connect.NewResponse(&agentsv1.ResendTelegramReplyResponse{
		Record:     updated,
		MessageIds: delivery.MessageIDs(),
	}), nil
}

func hasUncertainTelegramSegment(delivery *telegramsend.Delivery) bool {
	for _, segment := range delivery.Segments {
		if segment.Status == telegramsend.SegmentSending {
			return true
		}
	}
	return false
}
