package plane

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/genai"
)

type controlledTestRunner struct {
	started, canceled, settle chan struct{}
	calls                     atomic.Int32
}

func (r *controlledTestRunner) Run(ctx context.Context, _ string, _ []*genai.Part, _ string, _ *agentsv1.ContextInfo, _ runner.EventCallback, _ runner.CompactionCallback) (string, error) {
	r.calls.Add(1)
	close(r.started)
	<-ctx.Done()
	close(r.canceled)
	<-r.settle
	// Even a runner that returns cleanly after cancellation must lose to Stop.
	return "late successful result", nil
}

type planeTestResolver struct {
	app  *agentsv1.PlaneApp
	fail atomic.Bool
}

func (r *planeTestResolver) FindScopedApp(context.Context, string, string, string) (*agentsv1.PlaneApp, error) {
	if r.fail.Load() {
		return nil, errors.New("temporary lookup failure")
	}
	return r.app, nil
}
func planeQueueFixture(t *testing.T) (*StreamQueue, *Event) {
	t.Helper()
	client := planeTestRedis(t)
	if err := client.Del(t.Context(), StreamKey, ControlStreamKey).Err(); err != nil {
		t.Fatal(err)
	}
	q := NewQueue(client)
	if err := q.EnsureGroup(t.Context()); err != nil {
		t.Fatal(err)
	}
	e := &Event{WorkspaceID: "queue-test-" + uuid.NewString(), AppID: "app", InstallationID: uuid.NewString(),
		AgentRunID: "run", ActivityID: "prompt", Action: "prompted", Prompt: "work", ActorUserID: "user"}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		keys, _ := client.Keys(ctx, controlPrefix+e.WorkspaceID+":*").Result()
		dedupes, _ := client.Keys(ctx, "butter:plane:seen:"+e.InstallationID+":*").Result()
		keys = append(keys, dedupes...)
		keys = append(keys, StreamKey, ControlStreamKey)
		_ = client.Del(ctx, keys...).Err()
	})
	return q, e
}
func awaitPlane(t *testing.T, ch <-chan struct{}, what string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out: " + what)
	}
}

func TestAdmissionStopCancelsAcrossPodsBeforeAnyControlWorkerReads(t *testing.T) {
	q, e := planeQueueFixture(t)
	if _, err := q.Accept(t.Context(), e); err != nil {
		t.Fatal(err)
	}
	entries, err := q.Read(t.Context(), "holder", 1, time.Millisecond)
	if err != nil || len(entries) != 1 {
		t.Fatalf("read=%v err=%v", entries, err)
	}
	accepted, _ := DecodeEvent(entries[0].Payload)
	r := &controlledTestRunner{started: make(chan struct{}), canceled: make(chan struct{}), settle: make(chan struct{})}
	var settleOnce sync.Once
	settle := func() { settleOnce.Do(func() { close(r.settle) }) }
	defer settle()
	holderPub := &fakePublisher{}
	holder := NewProcessor(r, holderPub)
	holder.SetRunControl(NewRedisRunControl(q.client))
	holder.controlPoll = 10 * time.Millisecond
	runCtx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := holder.Process(runCtx, &agentsv1.PlaneApp{AgentId: "agent"}, accepted)
		done <- err
	}()
	awaitPlane(t, r.started, "model start")
	old := *e
	old.ActivityID = "unread-before-stop"
	if _, err := q.Accept(t.Context(), &old); err != nil {
		t.Fatal(err)
	}
	stop := *e
	stop.ActivityID, stop.Prompt, stop.Stop, stop.Signal = "stop", "", true, "stop"
	// This is the other Pod's webhook admission. No Worker has read Stop.
	if _, err := NewQueue(q.client).Accept(t.Context(), &stop); err != nil {
		t.Fatal(err)
	}
	awaitPlane(t, r.canceled, "cross-Pod persistent Stop marker")
	next := *e
	next.ActivityID = "new-after-stop"
	if _, err := q.Accept(t.Context(), &next); err != nil {
		t.Fatal(err)
	}
	controls, err := q.controls.Read(t.Context(), "other-pod", 1, time.Millisecond)
	if err != nil || len(controls) != 1 {
		t.Fatalf("control read=%v err=%v", controls, err)
	}
	control, _ := DecodeEvent(controls[0].Payload)
	if control.StopTarget == "" {
		t.Fatal("Stop admission did not freeze the holder token")
	}
	otherPub := &fakePublisher{}
	otherRunner := &fakeRunner{out: "new result"}
	other := NewProcessor(otherRunner, otherPub)
	other.SetRunControl(NewRedisRunControl(q.client))
	w := NewWorker(q, other, &planeTestResolver{app: &agentsv1.PlaneApp{AllowedUserIds: []string{"user"}}}, "other-pod")
	confirmed := make(chan error, 1)
	go func() { confirmed <- w.handle(t.Context(), controls[0]) }()
	// A canceled context alone is not a completed Stop.
	select {
	case err := <-confirmed:
		t.Fatalf("Stop confirmed before Runner converged: %v", err)
	case <-time.After(30 * time.Millisecond):
	}
	settle()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Stop lost to a clean runner return: %v", err)
	}
	if err := <-confirmed; err != nil {
		t.Fatal(err)
	}
	if holderPub.calls != 0 || otherPub.reply != "Stopped — send a message to continue." {
		t.Fatalf("late reply=%d confirmation=%q", holderPub.calls, otherPub.reply)
	}
	pending, err := q.Read(t.Context(), "other-pod", 2, time.Millisecond)
	if err != nil || len(pending) != 2 {
		t.Fatalf("pending=%v err=%v", pending, err)
	}
	older, _ := DecodeEvent(pending[0].Payload)
	newer, _ := DecodeEvent(pending[1].Payload)
	if !(accepted.AcceptedOrder < older.AcceptedOrder && older.AcceptedOrder < control.AcceptedOrder && control.AcceptedOrder < newer.AcceptedOrder) {
		t.Fatal("acceptance order was not durably assigned across streams")
	}
	if _, err := other.Process(t.Context(), &agentsv1.PlaneApp{AgentId: "agent"}, older); !errors.Is(err, ErrPromptDiscarded) {
		t.Fatalf("unread old Prompt executed: %v", err)
	}
	if _, err := other.Process(t.Context(), &agentsv1.PlaneApp{AgentId: "agent"}, newer); err != nil {
		t.Fatalf("new Prompt after Stop: %v", err)
	}
	if otherRunner.calls != 1 {
		t.Fatalf("model calls=%d, want only the new Prompt", otherRunner.calls)
	}
}

func TestFailedStopAdmissionDoesNotCancelOrSetCutoff(t *testing.T) {
	q, e := planeQueueFixture(t)
	c := NewRedisRunControl(q.client)
	if err := c.Begin(t.Context(), e, "holder"); err != nil {
		t.Fatal(err)
	}
	// An invalid destination must fail before changing the running turn.
	if err := q.client.Del(t.Context(), ControlStreamKey).Err(); err != nil {
		t.Fatal(err)
	}
	if err := q.client.Set(t.Context(), ControlStreamKey, "wrong type", 0).Err(); err != nil {
		t.Fatal(err)
	}
	stop := *e
	stop.ActivityID, stop.Stop = "stop", true
	if _, err := q.Accept(t.Context(), &stop); err == nil {
		t.Fatal("broken control stream accepted Stop")
	}
	if stopped, err := c.Renew(t.Context(), e, "holder"); err != nil || stopped {
		t.Fatalf("unaccepted Stop reached holder: stopped=%v err=%v", stopped, err)
	}
	if n := q.client.Exists(t.Context(), controlKey(e)+":cutoff").Val(); n != 0 {
		t.Fatal("unaccepted Stop discarded pending prompts")
	}
}

func TestDeniedStopAppendCannotLeaveControlSideEffects(t *testing.T) {
	q, e := planeQueueFixture(t)
	c := NewRedisRunControl(q.client)
	if err := c.Begin(t.Context(), e, "holder"); err != nil {
		t.Fatal(err)
	}
	user := "plane-test-" + uuid.NewString()
	password := uuid.NewString()
	if err := q.client.ACLSetUser(t.Context(), user, "on", ">"+password, "~*", "+@all", "-xadd").Err(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = q.client.ACLDelUser(context.Background(), user).Err() })
	client := redis.NewClient(&redis.Options{Addr: q.client.Options().Addr, DB: 12, Username: user, Password: password})
	t.Cleanup(func() { _ = client.Close() })
	stop := *e
	stop.ActivityID, stop.Stop = "stop", true
	if _, err := NewQueue(client).Accept(t.Context(), &stop); err == nil {
		t.Fatal("Stop accepted without XADD permission")
	}
	if stopped, err := c.Renew(t.Context(), e, "holder"); err != nil || stopped {
		t.Fatalf("unaccepted Stop reached holder: stopped=%v err=%v", stopped, err)
	}
	if q.client.Exists(t.Context(), controlKey(e)+":cutoff").Val() != 0 {
		t.Fatal("unaccepted Stop discarded pending prompts")
	}
}

func TestWorkerControlPoolBypassesFullModelPoolAndHeartbeats(t *testing.T) {
	q, e := planeQueueFixture(t)
	r := &controlledTestRunner{started: make(chan struct{}), canceled: make(chan struct{}), settle: make(chan struct{})}
	var once sync.Once
	settle := func() { once.Do(func() { close(r.settle) }) }
	defer settle()
	pub := &fakePublisher{}
	p := NewProcessor(r, pub)
	p.SetRunControl(NewRedisRunControl(q.client))
	p.controlPoll = 10 * time.Millisecond
	w := NewWorker(q, p, &planeTestResolver{app: &agentsv1.PlaneApp{AgentId: "agent", AllowedUserIds: []string{"user"}}}, "worker")
	w.concurrency, w.block, w.heartbeat, w.reclaimIdle = 1, 10*time.Millisecond, 10*time.Millisecond, 100*time.Millisecond
	if err := w.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer w.Stop()
	// Unblock the deliberately slow cancellation before waiting for Stop.
	defer settle()
	if _, err := q.Accept(t.Context(), e); err != nil {
		t.Fatal(err)
	}
	awaitPlane(t, r.started, "full model slot")
	time.Sleep(150 * time.Millisecond)
	stolen, err := q.inner.Claim(t.Context(), "thief", w.reclaimIdle, 1)
	if err != nil || len(stolen) != 0 {
		t.Fatalf("healthy model delivery was reclaimed: %v %v", stolen, err)
	}
	stop := *e
	stop.ActivityID, stop.Prompt, stop.Stop = "stop", "", true
	if _, err := q.Accept(t.Context(), &stop); err != nil {
		t.Fatal(err)
	}
	awaitPlane(t, r.canceled, "Stop while every model slot is full")
	settle()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		pub.mu.Lock()
		reply := pub.reply
		pub.mu.Unlock()
		if reply == "Stopped — send a message to continue." {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("control confirmation was blocked behind model work")
}

func TestWorkerRecoversUnstartedPendingAndDoesNotAckLookupFailure(t *testing.T) {
	q, e := planeQueueFixture(t)
	if _, err := q.Accept(t.Context(), e); err != nil {
		t.Fatal(err)
	}
	if _, err := q.Read(t.Context(), "dead-pod", 1, time.Millisecond); err != nil {
		t.Fatal(err)
	}
	resolver := &planeTestResolver{app: &agentsv1.PlaneApp{AgentId: "agent"}}
	resolver.fail.Store(true)
	r := &fakeRunner{out: "recovered"}
	p := NewProcessor(r, &fakePublisher{})
	p.SetRunControl(NewRedisRunControl(q.client))
	w := NewWorker(q, p, resolver, "replacement")
	w.block, w.reclaimIdle = 10*time.Millisecond, 30*time.Millisecond
	if err := w.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer w.Stop()
	time.Sleep(150 * time.Millisecond)
	pending, err := q.client.XPending(t.Context(), StreamKey, ConsumerGroup).Result()
	if err != nil || pending.Count != 1 {
		t.Fatalf("pre-execution failure was acknowledged: %+v %v", pending, err)
	}
	resolver.fail.Store(false)
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		pending, err = q.client.XPending(t.Context(), StreamKey, ConsumerGroup).Result()
		if err == nil && pending.Count == 0 {
			r.mu.Lock()
			calls := r.calls
			r.mu.Unlock()
			if calls != 1 {
				t.Fatalf("recovered model calls=%d", calls)
			}
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("unstarted pending event was not recovered")
}
