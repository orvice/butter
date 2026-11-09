package plane

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"go.orx.me/apps/butter/internal/repo/invocation"
	invocationmem "go.orx.me/apps/butter/internal/repo/invocation/memory"
	installationrepo "go.orx.me/apps/butter/internal/repo/planeinstallation"
	processingmem "go.orx.me/apps/butter/internal/repo/planeprocessing/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/adk/v2/agent"
)

// Real Redis admission, worker loops, fencing/batching and the real ADK Runner.
// Only storage, the deterministic agent leaf and Plane HTTP are test doubles.
func TestRedisEnvelopeMentionReachesReplyAndDeduplicates(t *testing.T) {
	for _, global := range []bool{false, true} {
		t.Run(map[bool]string{false: "granular grants", true: "global read write"}[global], func(t *testing.T) {
			testRedisEnvelopeMentionReachesReplyAndDeduplicates(t, global)
		})
	}
}

func testRedisEnvelopeMentionReachesReplyAndDeduplicates(t *testing.T, global bool) {
	t.Helper()
	client := planeTestRedis(t) // Dedicated test Redis, database 12.
	if err := client.Del(t.Context(), StreamKey, ControlStreamKey).Err(); err != nil {
		t.Fatal(err)
	}
	f := newReceiverFixture(t)
	if global {
		f.installation.GrantedScopes = []string{"read", "write"}
		if _, err := f.installations.Upsert(t.Context(), "ws", "app", f.installation, installationrepo.Tokens{}); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		keys, _ := client.Keys(ctx, controlPrefix+"ws:app:"+f.installation.GetId()+":*").Result()
		seen, _ := client.Keys(ctx, "butter:plane:seen:"+f.installation.GetId()+":*").Result()
		keys = append(keys, seen...)
		keys = append(keys, StreamKey, ControlStreamKey)
		_ = client.Del(ctx, keys...).Err()
	})
	q := NewQueue(client)
	f.receiver.queue = q
	records, invocations := processingmem.New(), invocationmem.New()
	var executions, responses, posts atomic.Int32
	started, release := make(chan struct{}), make(chan struct{})
	caller, cancelCaller := context.WithCancel(t.Context())
	defer cancelCaller()
	base, _, app := invocationProcessor(t, func(ctx agent.InvocationContext) (string, error) {
		if executions.Add(1) == 1 {
			close(started)
			select {
			case <-release:
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}
		return "Weather report", nil
	}, invocations)
	f.app.AgentId = app.GetAgentId()
	p := batchProcessor(q, base.runner, records, "envelope-worker")
	activityClient := NewActivityHTTPClient("https://plane.invalid", &http.Client{Transport: activityRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		n := posts.Add(1)
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/workspaces/plane-workspace/runs/run-1/activities/" {
			t.Error("wrong Plane activity endpoint")
		}
		if r.Header.Get("Authorization") != "Bearer fixture-bot-token" {
			t.Error("wrong Plane installation token")
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		if strings.Contains(string(body), `"type":"response"`) {
			responses.Add(1)
			if !strings.Contains(string(body), "Weather report") {
				t.Error("missing final response body")
			}
		} else if !strings.Contains(string(body), `"type":"thought"`) {
			t.Error("unexpected Plane activity type")
		}
		return &http.Response{StatusCode: 201, Body: io.NopCloser(strings.NewReader(fmt.Sprintf(`{"id":"created-%d"}`, n)))}, nil
	})}, func(_ context.Context, ws, app, installation string) (string, error) {
		if ws != "ws" || app != "app" || installation != "external-installation" {
			t.Error("token scope did not use the verified installation")
		}
		return "fixture-bot-token", nil
	})
	publisher := NewDurablePublisher(ActivityReplyPublisher{Client: activityClient, Installations: f.installations}, NewMemoryDeliveryStore())
	publisher.observe = p.deliveryProjection
	p.publisher = publisher
	worker := NewWorker(q, p, &planeTestResolver{app: f.app}, "envelope-worker")
	if err := worker.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(worker.Stop)
	payload := officialEnvelope(t)
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	if decision, err := f.deliverBodyContext(t, caller, body); err != nil || decision != DecisionAccepted {
		t.Fatalf("mention admission=%s err=%v", decision, err)
	}
	awaitPlane(t, started, "model still running after durable ACK")
	if responses.Load() != 0 {
		t.Fatal("reply finished before the deliberately blocked model")
	}
	// Disconnecting the caller after ACK must not cancel already-durable work.
	cancelCaller()
	close(release)
	awaitRecord := func(activity, actor string) *agentsv1.PlaneProcessingRecord {
		t.Helper()
		var row *agentsv1.PlaneProcessingRecord
		waitBatch(t, func() bool {
			var err error
			row, err = records.FindByEvent(t.Context(), &agentsv1.PlaneProcessingRecord{WorkspaceId: "ws", AppId: "app", InstallationId: f.installation.GetId(), RunId: "run-1", ActivityId: activity, EventAction: "prompted"})
			return err == nil && row.GetStatus() == agentsv1.PlaneProcessingStatus_PLANE_PROCESSING_STATUS_SUCCEEDED && row.GetReplyDeliveryStatus() == "delivered"
		}, "accepted envelope did not reach a delivered reply")
		if row.GetActorUserId() != actor || row.GetInvocationId() == "" || row.GetExecutionFence() == 0 {
			t.Fatalf("missing business actor or fenced Invocation: %v", row)
		}
		return row
	}
	first := awaitRecord("activity-1", "user-1")
	payload["webhook_id"] = "retry-webhook"
	payload["activity"].(map[string]any)["id"] = "different-audit-event"
	if decision, err := f.deliver(t, payload); err != nil || decision != DecisionDuplicate {
		t.Fatalf("audit metadata changed business deduplication: decision=%s err=%v", decision, err)
	}
	business := payload["data"].(map[string]any)
	prompt := business["agent_run_activity"].(map[string]any)
	prompt["id"], prompt["actor"] = "activity-2", "user-2"
	prompt["content"].(map[string]any)["body"] = "Follow-up question"
	if decision, err := f.deliver(t, payload); err != nil || decision != DecisionAccepted {
		t.Fatalf("follow-up admission=%s err=%v", decision, err)
	}
	second := awaitRecord("activity-2", "user-2")
	rows, _, _, err := invocations.List(t.Context(), invocation.ListFilter{WorkspaceID: "ws"}, 10, "")
	if err != nil || len(rows) != 2 || executions.Load() != 2 || responses.Load() != 2 {
		t.Fatalf("wrong executions/replies: invocations=%d runs=%d replies=%d err=%v", len(rows), executions.Load(), responses.Load(), err)
	}
	ids := map[string]bool{}
	for _, row := range rows {
		ids[row.GetId()] = true
		if row.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED {
			t.Fatalf("Invocation failed: %v", row)
		}
	}
	if first.GetInvocationId() == second.GetInvocationId() || !ids[first.GetInvocationId()] || !ids[second.GetInvocationId()] {
		t.Fatal("processing records did not retain their actual Invocation IDs")
	}
}
