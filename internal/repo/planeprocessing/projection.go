package planeprocessing

import (
	"crypto/sha256"
	"fmt"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func BusinessKey(r *agentsv1.PlaneProcessingRecord) string {
	return "record:" + r.GetWorkspaceId() + ":" + r.GetAppId() + ":" + LegacyKey(r)
}

func LegacyKey(r *agentsv1.PlaneProcessingRecord) string {
	return r.GetInstallationId() + ":" + r.GetRunId() + ":" + r.GetActivityId() + ":" + r.GetEventAction()
}

func Normalize(r *agentsv1.PlaneProcessingRecord) *agentsv1.PlaneProcessingRecord {
	c := proto.Clone(r).(*agentsv1.PlaneProcessingRecord)
	if c.Id == "" {
		c.Id = fmt.Sprintf("plane-%x", sha256.Sum256([]byte(BusinessKey(c))))
	}
	if c.ReceivedAt == nil {
		c.ReceivedAt = timestamppb.Now()
	}
	c.UpdatedAt = timestamppb.Now()
	return c
}

func Terminal(s agentsv1.PlaneProcessingStatus) bool {
	return s == agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_SUCCEEDED ||
		s == agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_FAILED ||
		s == agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_CANCELLED ||
		s == agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_FAILED_UNCERTAIN
}

// MergeProjection never lets an admission/queued observer overwrite execution,
// or a late holder overwrite another invocation. Confirmation delivery remains
// separate from the execution status, including after a terminal projection.
func MergeProjection(old, next *agentsv1.PlaneProcessingRecord) *agentsv1.PlaneProcessingRecord {
	if old == nil {
		return Normalize(next)
	}
	out := proto.Clone(old).(*agentsv1.PlaneProcessingRecord)
	if next.GetExecutionFence() > 0 && next.GetExecutionFence() < old.GetExecutionFence() {
		return out
	}
	if next.GetStatus() == agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_RECEIVED && old.GetStatus() != agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_RECEIVED {
		return out
	}
	if next.GetQueueConfirmationStatus() != "" && old.GetQueueConfirmationStatus() != "delivered" {
		out.QueueConfirmationStatus = next.GetQueueConfirmationStatus()
	}
	if next.GetReplyDeliveryStatus() != "" && old.GetReplyDeliveryStatus() != "delivered" && (old.GetInvocationId() == "" || old.GetInvocationId() == next.GetInvocationId()) {
		out.ReplyDeliveryStatus = next.GetReplyDeliveryStatus()
	}
	low := next.GetStatus() == agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_RECEIVED ||
		next.GetStatus() == agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_QUEUED
	if next.GetStatus() == agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_UNSPECIFIED || Terminal(old.GetStatus()) || (low && old.GetStatus() == agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_PROCESSING) ||
		(Terminal(next.GetStatus()) && next.GetInvocationId() != "" && old.GetInvocationId() != "" && next.GetInvocationId() != old.GetInvocationId()) {
		out.UpdatedAt = timestamppb.Now()
		return out
	}
	if old.GetInvocationId() != "" && old.GetInvocationId() == next.GetInvocationId() {
		next = proto.Clone(next).(*agentsv1.PlaneProcessingRecord)
		next.BatchActivityIds = append([]string(nil), old.GetBatchActivityIds()...)
		next.MemoryCaptureDisabledReason = old.GetMemoryCaptureDisabledReason()
		if next.ExecutionFence == 0 {
			next.ExecutionFence = old.GetExecutionFence()
		}
	}
	id, received := out.Id, out.ReceivedAt
	queueStatus, replyStatus := out.QueueConfirmationStatus, out.ReplyDeliveryStatus
	out = Normalize(next)
	out.Id, out.ReceivedAt = id, received
	if out.QueueConfirmationStatus == "" {
		out.QueueConfirmationStatus = queueStatus
	}
	if out.ReplyDeliveryStatus == "" {
		out.ReplyDeliveryStatus = replyStatus
	}
	return out
}
