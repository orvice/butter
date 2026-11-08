package plane

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.orx.me/apps/butter/internal/redact"
	planeprocessing "go.orx.me/apps/butter/internal/repo/planeprocessing"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func processingState(state RunState) agentsv1.PlaneProcessingStatus {
	switch state {
	case RunStateSucceeded:
		return agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_SUCCEEDED
	case RunStateStopped:
		return agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_CANCELLED
	case "uncertain":
		return agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_FAILED_UNCERTAIN
	default:
		return agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_FAILED
	}
}

func (p *Processor) project(ctx context.Context, e *Event, status agentsv1.PlaneProcessingStatus, reason, confirmation, delivery string) error {
	if p.processing == nil {
		return nil
	}
	ids := []string{}
	for _, member := range membersOf(e) {
		ids = append(ids, member.ActivityID)
	}
	_, _, metadata, _ := batchInput(membersOf(e))
	for _, member := range membersOf(e) {
		row := &agentsv1.PlaneProcessingRecord{
			WorkspaceId: member.WorkspaceID, AppId: member.AppID, InstallationId: member.InstallationID,
			PlaneWorkspaceId: member.PlaneWorkspaceID, BotUserId: member.BotUserID, RunId: member.AgentRunID,
			ActivityId: member.ActivityID, EventType: member.Type, EventAction: member.Action,
			DeliveryId: member.DeliveryID, Status: status, FailureReason: redact.Text(reason),
			ActorUserId: member.ActorUserID, AgentId: member.AgentID, InvocationId: e.InvocationID,
			ExecutionFence: e.ExecutionFence,
			AcceptedOrder:  member.AcceptedOrder, BatchActivityIds: ids,
			MemoryCaptureDisabledReason: metadata["memory_capture_disabled_reason"],
			QueueConfirmationStatus:     confirmation, ReplyDeliveryStatus: delivery,
		}
		if member.ReceivedAt > 0 {
			row.ReceivedAt = timestamppb.New(time.UnixMilli(member.ReceivedAt))
		}
		if err := p.processing.Project(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func (w *Worker) queued(ctx context.Context, e *Event) error {
	singleton := *e
	singleton.Members = nil
	e = &singleton
	if err := w.processor.project(ctx, e, agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_QUEUED, "", "", ""); err != nil {
		return err
	}
	notification := *e
	notification.Members = nil
	notification.ActivityID += ":queued"
	notification.ActivityType = ActivityThought
	err := w.processor.publisher.PublishReply(ctx, &notification, "Queued — your message will be included in a following turn.")
	status := "delivered"
	if err != nil && !errors.Is(err, ErrDeliveryAlreadyDone) {
		status = "unconfirmed"
	}
	if saveErr := w.processor.project(ctx, e, agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_QUEUED, "", w.processor.deliveryStatus(ctx, &notification, status), ""); saveErr != nil {
		return saveErr
	}
	return ErrRunBusy
}

func (p *Processor) deliveryProjection(ctx context.Context, row *DeliveryAttempt) error {
	if row.Event == nil {
		return nil
	}
	e := *row.Event
	confirmation, delivery := "", string(row.Status)
	if e.ActivityType == ActivityThought && strings.HasSuffix(e.ActivityID, ":queued") {
		e.ActivityID = strings.TrimSuffix(e.ActivityID, ":queued")
		e.Members = nil
		confirmation, delivery = delivery, ""
	}
	return p.project(ctx, &e, agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_UNSPECIFIED, "", confirmation, delivery)
}

func (p *Processor) recoverDeliveryProjection(ctx context.Context, e *Event, state RunState) error {
	durable, ok := p.publisher.(*DurablePublisher)
	if !ok {
		return nil
	}
	primary := *e
	if p.processing != nil {
		record, err := p.processing.FindByEvent(ctx, &agentsv1.PlaneProcessingRecord{WorkspaceId: e.WorkspaceID, AppId: e.AppID, InstallationId: e.InstallationID, RunId: e.AgentRunID, ActivityId: e.ActivityID, EventAction: e.Action})
		if err != nil && !errors.Is(err, planeprocessing.ErrNotFound) {
			return err
		}
		if len(record.GetBatchActivityIds()) > 0 {
			primary.ActivityID = record.GetBatchActivityIds()[0]
		}
	}
	row, err := durable.store.Get(ctx, DeliveryKey(&primary))
	if errors.Is(err, ErrDeliveryNotFound) {
		return p.project(ctx, e, agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_UNSPECIFIED, "", "", "not_recorded")
	}
	if err != nil {
		return err
	}
	return p.deliveryProjection(ctx, row)
}

func (p *Processor) deliveryStatus(ctx context.Context, e *Event, fallback string) string {
	if durable, ok := p.publisher.(*DurablePublisher); ok {
		if row, err := durable.store.Get(ctx, DeliveryKey(e)); err == nil {
			return string(row.Status)
		}
	}
	return fallback
}
