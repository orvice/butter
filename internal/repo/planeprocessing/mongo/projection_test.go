package mongo

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	planeprocessing "go.orx.me/apps/butter/internal/repo/planeprocessing"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
)

func projectionStore(t *testing.T) *Store {
	t.Helper()
	uri := os.Getenv("BUTTER_TEST_MONGO_URI")
	if uri == "" {
		t.Skip("BUTTER_TEST_MONGO_URI is required")
	}
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatal(err)
	}
	db := client.Database("butter_test_plane_proj_" + uuid.NewString())
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = db.Drop(ctx)
		_ = client.Disconnect(ctx)
	})
	store := New(db)
	if err := store.EnsureIndexes(t.Context()); err != nil {
		t.Fatal(err)
	}
	return store
}

func TestProjectionCreatesEveryMissingRecordAndPreservesFailure(t *testing.T) {
	store := projectionStore(t)
	for _, activity := range []string{"first", "second", "third"} {
		if err := store.Project(t.Context(), &agentsv1.PlaneProcessingRecord{
			WorkspaceId: "ws", AppId: "app", InstallationId: "inst", RunId: "run", ActivityId: activity,
			Status: agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_FAILED_UNCERTAIN, FailureReason: "holder disappeared",
		}); err != nil {
			t.Fatal(err)
		}
	}
	rows, err := store.List(t.Context(), planeprocessing.Filter{WorkspaceID: "ws", Status: agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_FAILED_UNCERTAIN})
	if err != nil || len(rows) != 3 {
		t.Fatalf("rows=%v err=%v", rows, err)
	}
	ids := map[string]bool{}
	for _, row := range rows {
		if row.GetId() == "" || ids[row.GetId()] || row.GetFailureReason() != "holder disappeared" || row.GetReceivedAt() == nil {
			t.Fatalf("lost projection=%v", row)
		}
		ids[row.GetId()] = true
	}
	other, err := store.List(t.Context(), planeprocessing.Filter{WorkspaceID: "other"})
	if err != nil || len(other) != 0 {
		t.Fatalf("cross-workspace records=%v err=%v", other, err)
	}
}

func TestProjectionCASPreventsLateAdmissionAndExpiredHolderWrites(t *testing.T) {
	store := projectionStore(t)
	record := &agentsv1.PlaneProcessingRecord{
		WorkspaceId: "ws", AppId: "app", InstallationId: "inst", RunId: "run", ActivityId: "prompt", EventAction: "prompted",
		Status: agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_PROCESSING, InvocationId: "new-holder", ExecutionFence: 2,
		ActorUserId: "alice", AgentId: "agent", BatchActivityIds: []string{"prompt", "followup"}, MemoryCaptureDisabledReason: "mixed actors",
	}
	if err := store.Project(t.Context(), record); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for range 12 {
		wg.Go(func() {
			admission := proto.Clone(record).(*agentsv1.PlaneProcessingRecord)
			admission.Status, admission.InvocationId, admission.ExecutionFence = agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_RECEIVED, "", 0
			if err := store.Project(t.Context(), admission); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	stale := proto.Clone(record).(*agentsv1.PlaneProcessingRecord)
	stale.ExecutionFence, stale.InvocationId = 1, "expired-holder"
	stale.Status = agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_PROCESSING
	if err := store.Project(t.Context(), stale); err != nil {
		t.Fatal(err)
	}
	// Recovery knows the original invocation, but reconstructs only this
	// Stream member. Its batch provenance and fence must survive the repair.
	recovered := proto.Clone(record).(*agentsv1.PlaneProcessingRecord)
	recovered.Status = agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_SUCCEEDED
	recovered.ExecutionFence, recovered.BatchActivityIds, recovered.MemoryCaptureDisabledReason = 0, []string{"prompt"}, ""
	if err := store.Project(t.Context(), recovered); err != nil {
		t.Fatal(err)
	}
	delivered := proto.Clone(recovered).(*agentsv1.PlaneProcessingRecord)
	delivered.Status, delivered.ReplyDeliveryStatus = agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_UNSPECIFIED, "delivered"
	if err := store.Project(t.Context(), delivered); err != nil {
		t.Fatal(err)
	}
	delivered.ReplyDeliveryStatus = "unconfirmed"
	if err := store.Project(t.Context(), delivered); err != nil {
		t.Fatal(err)
	}
	rows, err := store.List(t.Context(), planeprocessing.Filter{WorkspaceID: "ws"})
	if err != nil || len(rows) != 1 {
		t.Fatalf("rows=%v err=%v", rows, err)
	}
	row := rows[0]
	if row.GetStatus() != agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_SUCCEEDED || row.GetInvocationId() != "new-holder" ||
		row.GetExecutionFence() != 2 || len(row.GetBatchActivityIds()) != 2 || row.GetMemoryCaptureDisabledReason() != "mixed actors" ||
		row.GetReplyDeliveryStatus() != "delivered" {
		t.Fatalf("projection regressed: %v", row)
	}
}

func TestProjectionReadsAndUpgradesLegacyStatusOverlay(t *testing.T) {
	store := projectionStore(t)
	record := &agentsv1.PlaneProcessingRecord{Id: "legacy", WorkspaceId: "ws", AppId: "app", InstallationId: "inst", RunId: "run", ActivityId: "a",
		Status: agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_RECEIVED}
	d, err := encode(record)
	if err != nil {
		t.Fatal(err)
	}
	d.Key = planeprocessing.LegacyKey(record)
	status := agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_FAILED
	d.Status, d.Failure, d.UpdatedAt = &status, "legacy failure", time.Now()
	if _, err := store.col.InsertOne(t.Context(), d); err != nil {
		t.Fatal(err)
	}
	got, err := store.Get(t.Context(), "ws", "legacy")
	if err != nil || got.GetStatus() != status || got.GetFailureReason() != "legacy failure" || got.GetUpdatedAt() == nil {
		t.Fatalf("legacy overlay=%v err=%v", got, err)
	}
	record.Status = agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_QUEUED
	if err := store.Project(t.Context(), record); err != nil {
		t.Fatal(err)
	}
	var upgraded doc
	if err := store.col.FindOne(t.Context(), bson.M{"_id": "legacy"}).Decode(&upgraded); err != nil {
		t.Fatal(err)
	}
	if upgraded.Key != planeprocessing.BusinessKey(record) || upgraded.ProjectionRevision != 1 {
		t.Fatalf("legacy record not upgraded in place: %+v", upgraded)
	}
	got, _ = decode(upgraded)
	if got.GetStatus() != status || got.GetFailureReason() != "legacy failure" {
		t.Fatalf("late queue projection rewrote a terminal legacy record: %v", got)
	}
}
