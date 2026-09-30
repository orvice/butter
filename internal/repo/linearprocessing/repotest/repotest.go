// Package repotest is a conformance suite every linearprocessing.Repository
// implementation must pass.
package repotest

import (
	"errors"
	"testing"
	"time"

	"go.orx.me/apps/butter/internal/repo/linearprocessing"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Factory builds a fresh, empty repository per test.
type Factory func(t *testing.T) linearprocessing.Repository

func received(deliveryID string) *agentsv1.LinearProcessingRecord {
	return &agentsv1.LinearProcessingRecord{
		WorkspaceId: "ws-a", AppId: "app-1", DeliveryId: deliveryID, AgentSessionId: "s-1",
		Status: agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_RECEIVED,
	}
}

// Run exercises the conformance suite.
func Run(t *testing.T, factory Factory) {
	now := time.Now().UTC().Truncate(time.Millisecond)

	t.Run("OneRecordPerDeliveryWithALiveLease", func(t *testing.T) {
		repo := factory(t)
		ctx := t.Context()
		first, action, err := repo.Claim(ctx, received("d-1"), "lease-a", now, now.Add(time.Minute))
		if err != nil || action != linearprocessing.ClaimRunAgent {
			t.Fatalf("first Claim = %v, %v; want ClaimRunAgent", action, err)
		}
		if first.GetId() == "" || first.GetInvocationId() == "" || first.GetAttempts() != 1 {
			t.Fatalf("first = %+v; want an ID, an invocation ID and one attempt", first)
		}
		if !first.GetExpiresAt().AsTime().Equal(now.Add(linearprocessing.RetentionPeriod)) {
			t.Fatalf("expires_at = %v, want 30 days on", first.GetExpiresAt().AsTime())
		}
		if _, _, err := repo.Claim(ctx, received("d-1"), "lease-b", now, now.Add(time.Minute)); !errors.Is(err, linearprocessing.ErrInProgress) {
			t.Fatalf("second Claim during a live lease = %v, want ErrInProgress", err)
		}
		// Once the lease expires a redelivery re-claims the same record.
		again, action, err := repo.Claim(ctx, received("d-1"), "lease-b", now.Add(2*time.Minute), now.Add(3*time.Minute))
		if err != nil || action != linearprocessing.ClaimRunAgent || again.GetId() != first.GetId() || again.GetAttempts() != 2 {
			t.Fatalf("re-claim = %+v, %v, %v; want the same record, attempt 2", again, action, err)
		}
		if again.GetInvocationId() != first.GetInvocationId() {
			t.Fatal("re-claim changed the invocation ID")
		}
	})

	t.Run("RecoveryIsDerivedFromTheStoredState", func(t *testing.T) {
		for _, tc := range []struct {
			name   string
			status agentsv1.LinearProcessingStatus
			output string
			want   linearprocessing.ClaimAction
		}{
			{"processing is reported as interrupted", agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_PROCESSING, "", linearprocessing.ClaimReportInterrupted},
			{"ready to deliver resumes delivery", agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_READY_TO_DELIVER, "reply", linearprocessing.ClaimResumeDelivery},
			{"failed with output resumes delivery", agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_FAILED, "reply", linearprocessing.ClaimResumeDelivery},
			{"failed before the agent retries", agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_FAILED, "", linearprocessing.ClaimRunAgent},
			{"succeeded is acknowledged", agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_SUCCEEDED, "reply", linearprocessing.ClaimAcknowledge},
			{"uncertain is acknowledged", agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_FAILED_UNCERTAIN, "", linearprocessing.ClaimAcknowledge},
			{"cancelled is acknowledged", agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_CANCELLED, "", linearprocessing.ClaimAcknowledge},
		} {
			t.Run(tc.name, func(t *testing.T) {
				repo := factory(t)
				ctx := t.Context()
				record, _, err := repo.Claim(ctx, received("d-1"), "lease-a", now, now.Add(time.Minute))
				if err != nil {
					t.Fatalf("Claim: %v", err)
				}
				record.Status = tc.status
				record.Output = tc.output
				if _, err := repo.UpdateClaimed(ctx, record, "lease-a"); err != nil {
					t.Fatalf("UpdateClaimed: %v", err)
				}
				if err := repo.ReleaseClaim(ctx, "ws-a", record.GetId(), "lease-a"); err != nil {
					t.Fatalf("ReleaseClaim: %v", err)
				}
				got, action, err := repo.Claim(ctx, received("d-1"), "lease-b", now.Add(time.Second), now.Add(time.Minute))
				if err != nil || action != tc.want {
					t.Fatalf("re-claim = %v, %v; want %v", action, err, tc.want)
				}
				if tc.want == linearprocessing.ClaimReportInterrupted {
					if got.GetStatus() != agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_FAILED_UNCERTAIN || !got.GetDeadLettered() {
						t.Fatalf("interrupted record = %+v; want FAILED_UNCERTAIN, dead-lettered", got)
					}
					stored, _ := repo.Get(ctx, "ws-a", got.GetId())
					if stored.GetStatus() != agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_FAILED_UNCERTAIN {
						t.Fatal("the interrupted mark was not persisted")
					}
				}
			})
		}
	})

	t.Run("UpdatesAreFencedOnTheLease", func(t *testing.T) {
		repo := factory(t)
		ctx := t.Context()
		record, _, err := repo.Claim(ctx, received("d-1"), "lease-a", now, now.Add(time.Minute))
		if err != nil {
			t.Fatalf("Claim: %v", err)
		}
		record.Status = agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_PROCESSING
		if _, err := repo.UpdateClaimed(ctx, record, "stale"); !errors.Is(err, linearprocessing.ErrLeaseLost) {
			t.Fatalf("stale UpdateClaimed = %v, want ErrLeaseLost", err)
		}
		if err := repo.RenewClaim(ctx, "ws-a", record.GetId(), "stale", now.Add(time.Hour)); !errors.Is(err, linearprocessing.ErrLeaseLost) {
			t.Fatalf("stale RenewClaim = %v, want ErrLeaseLost", err)
		}
		if err := repo.RenewClaim(ctx, "ws-a", record.GetId(), "lease-a", now.Add(time.Hour)); err != nil {
			t.Fatalf("RenewClaim: %v", err)
		}
		if _, err := repo.ClaimForResend(ctx, "ws-a", record.GetId(), "lease-b", now.Add(2*time.Minute), now.Add(3*time.Minute)); !errors.Is(err, linearprocessing.ErrInProgress) {
			t.Fatalf("ClaimForResend during a renewed lease = %v, want ErrInProgress", err)
		}
		if err := repo.ReleaseClaim(ctx, "ws-a", record.GetId(), "lease-a"); err != nil {
			t.Fatalf("ReleaseClaim: %v", err)
		}
		if _, err := repo.ClaimForResend(ctx, "ws-a", record.GetId(), "lease-b", now, now.Add(time.Minute)); err != nil {
			t.Fatalf("ClaimForResend after release: %v", err)
		}
	})

	t.Run("ListIsWorkspaceScopedNewestFirstAndFiltered", func(t *testing.T) {
		repo := factory(t)
		ctx := t.Context()
		for i, id := range []string{"d-1", "d-2", "d-3"} {
			record := received(id)
			if id == "d-3" {
				record.AppId = "app-2"
			}
			if _, _, err := repo.Claim(ctx, record, "", now.Add(time.Duration(i)*time.Second), now); err != nil {
				t.Fatalf("Claim(%s): %v", id, err)
			}
		}
		other := received("d-x")
		other.WorkspaceId = "ws-b"
		if _, _, err := repo.Claim(ctx, other, "", now, now); err != nil {
			t.Fatalf("Claim(ws-b): %v", err)
		}
		all, err := repo.List(ctx, linearprocessing.Filter{WorkspaceID: "ws-a"})
		if err != nil || len(all) != 3 || all[0].GetDeliveryId() != "d-3" {
			t.Fatalf("List = %d records, %v; want 3 newest first", len(all), err)
		}
		byApp, _ := repo.List(ctx, linearprocessing.Filter{WorkspaceID: "ws-a", AppID: "app-1"})
		if len(byApp) != 2 {
			t.Fatalf("List by app = %d, want 2", len(byApp))
		}
		limited, _ := repo.List(ctx, linearprocessing.Filter{WorkspaceID: "ws-a", Limit: 1})
		if len(limited) != 1 {
			t.Fatalf("List with limit = %d, want 1", len(limited))
		}
		byStatus, _ := repo.List(ctx, linearprocessing.Filter{WorkspaceID: "ws-a", Status: agentsv1.LinearProcessingStatus_LINEAR_PROCESSING_STATUS_SUCCEEDED})
		if len(byStatus) != 0 {
			t.Fatalf("List by status = %d, want 0", len(byStatus))
		}
		if _, err := repo.Get(ctx, "ws-b", all[0].GetId()); !errors.Is(err, linearprocessing.ErrNotFound) {
			t.Fatalf("cross-workspace Get = %v, want ErrNotFound", err)
		}
	})
}
