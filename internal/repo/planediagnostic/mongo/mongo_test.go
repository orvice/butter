package mongodiagnostic

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	plane "go.orx.me/apps/butter/internal/runtime/plane"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestDiagnosticBSONMatchesScopedQueries(t *testing.T) {
	record := &plane.DiagnosticRecord{Key: "receipt", WorkspaceID: "ws", AppID: "app", Class: plane.FailurePermanent, Summary: "safe", CreatedAt: time.Now()}
	raw, err := bson.Marshal(toDoc(record))
	if err != nil {
		t.Fatal(err)
	}
	var fields bson.M
	if err := bson.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	if fields["workspace_id"] != "ws" || fields["app_id"] != "app" || fields["created_at"] == nil {
		t.Fatalf("diagnostic cannot be found by scoped queries/indexes: %v", fields)
	}
	if _, exists := fields["expires_at"]; exists {
		t.Fatal("legacy execution diagnostics gained a destructive TTL")
	}
}

func TestMongoRejectedReceiptRoundTripAndIsolation(t *testing.T) {
	uri := os.Getenv("BUTTER_TEST_MONGO_URI")
	if uri == "" {
		t.Skip("BUTTER_TEST_MONGO_URI is required")
	}
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatal(err)
	}
	db := client.Database("butter_plane_receipt_" + strconv.FormatInt(time.Now().UnixNano(), 10))
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
	record := &plane.DiagnosticRecord{Key: "webhook:receipt", WorkspaceID: "ws", AppID: "app", Class: plane.FailurePermanent, Summary: "safe rejection", CreatedAt: time.Now()}
	for range 2 {
		if err := store.Put(t.Context(), record); err != nil {
			t.Fatal(err)
		}
	}
	row, err := store.Get(t.Context(), "ws", record.Key)
	if err != nil || row.AppID != "app" || row.Summary != record.Summary || row.Recoverable {
		t.Fatalf("receipt not readable after durable Put: %v %v", row, err)
	}
	rows, err := store.List(t.Context(), "ws", "app")
	if err != nil || len(rows) != 1 {
		t.Fatalf("receipt projection/retry: count=%d err=%v", len(rows), err)
	}
	if _, err := store.Get(t.Context(), "foreign", record.Key); err == nil {
		t.Fatal("foreign workspace read the receipt")
	}
	rows, err = store.List(t.Context(), "ws", "foreign-app")
	if err != nil || len(rows) != 0 {
		t.Fatal("foreign App listed the receipt")
	}
	_, err = store.c.InsertMany(t.Context(), []any{
		bson.M{"key": "legacy", "workspaceid": "ws", "appid": "app", "createdat": time.Now(), "summary": "old diagnostic", "class": "permanent"},
		bson.M{"key": "mixed-foreign", "workspace_id": "foreign", "workspaceid": "ws", "app_id": "app", "appid": "app"},
		bson.M{"key": "mixed-app", "workspace_id": "ws", "app_id": "foreign-app", "appid": "app"},
	})
	if err != nil {
		t.Fatal(err)
	}
	legacy, err := store.Get(t.Context(), "ws", "legacy")
	if err != nil || legacy.WorkspaceID != "ws" || legacy.AppID != "app" || legacy.CreatedAt.IsZero() {
		t.Fatalf("legacy read failed: %v %v", legacy, err)
	}
	if _, err := store.Get(t.Context(), "ws", "mixed-foreign"); err == nil {
		t.Fatal("legacy fallback bypassed canonical workspace")
	}
	rows, err = store.List(t.Context(), "ws", "app")
	if err != nil || len(rows) != 2 {
		t.Fatalf("legacy App filter bypassed canonical claims: count=%d err=%v", len(rows), err)
	}
	if err := store.Put(t.Context(), legacy); err != nil {
		t.Fatal(err)
	}
	rows, err = store.List(t.Context(), "ws", "app")
	if err != nil || len(rows) != 2 {
		t.Fatal("writing legacy record duplicated or lost history")
	}
	// Cross the same receipt method as the HTTP adapter, backed by real Mongo.
	receiver := plane.NewReceiver(nil, nil, nil, nil)
	receiver.SetDiagnosticStore(store)
	key, err := receiver.RecordWebhookRejection(t.Context(), &agentsv1.PlaneApp{WorkspaceId: "ws", Id: "app"}, []byte(`{"event":"unsupported"}`), plane.ErrIdentity)
	if err != nil {
		t.Fatal(err)
	}
	row, err = store.Get(t.Context(), "ws", key)
	if err != nil || row.AppID != "app" || row.Class != plane.FailurePermanent || row.ExpiresAt.IsZero() {
		t.Fatalf("HTTP receipt is not durably readable: %v %v", row, err)
	}
	var history []any
	for i := range 202 {
		item := bson.M{"key": fmt.Sprintf("history-%d", i)}
		if i%2 == 0 {
			item["workspace_id"], item["app_id"], item["created_at"] = "ws", "history-app", time.Now().Add(time.Duration(i)*time.Minute)
		} else {
			item["workspaceid"], item["appid"], item["createdat"] = "ws", "history-app", time.Now().Add(time.Duration(i)*time.Minute)
		}
		history = append(history, item)
	}
	if _, err := store.c.InsertMany(t.Context(), history); err != nil {
		t.Fatal(err)
	}
	rows, err = store.List(t.Context(), "ws", "history-app")
	if err != nil || len(rows) != plane.DiagnosticHistoryLimit || rows[0].Key != "history-201" || rows[len(rows)-1].Key != "history-2" {
		t.Fatalf("history not bounded/ordered across schemas: count=%d err=%v", len(rows), err)
	}
}
