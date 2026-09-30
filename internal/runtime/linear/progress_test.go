package linear

// Progress tests (ADR-0015 §9, #368): while the Agent works, the Linear
// Agent Session shows readable, rate-limited tool progress, and the run is
// never slowed by it.

import (
	"context"
	"strings"
	"testing"
	"time"

	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/linearapi"
	"go.orx.me/apps/butter/internal/linearapi/lineartest"
	"go.orx.me/apps/butter/internal/runtime/runner"
)

func callEvent(name string, args map[string]any) *session.Event {
	ev := session.NewEvent(context.Background(), "inv")
	ev.Content = &genai.Content{Role: genai.RoleModel, Parts: []*genai.Part{{FunctionCall: &genai.FunctionCall{Name: name, Args: args}}}}
	return ev
}

func actions(acts []lineartest.Activity) []lineartest.Activity {
	var out []lineartest.Activity
	for _, a := range acts {
		if a.Type == linearapi.ActivityAction {
			out = append(out, a)
		}
	}
	return out
}

func newProgressFixture(t *testing.T, interval time.Duration) *orchestratorFixture {
	t.Helper()
	fx := newOrchestratorFixture(t, nil)
	fx.orch.progressInterval = interval
	return fx
}

func TestABurstOfToolCallsPostsAtMostOneActionPerIntervalLatestWins(t *testing.T) {
	fx := newProgressFixture(t, 100*time.Millisecond)
	fx.runner.emit = func(onEvent runner.EventCallback, _ runner.CompactionCallback) {
		for i := 1; i <= 5; i++ {
			onEvent(callEvent("bash", map[string]any{"command": "step " + string(rune('0'+i))}))
		}
		time.Sleep(150 * time.Millisecond)
	}
	fx.handle(t, fx.event(ActionCreated))
	got := actions(fx.linear.Activities())
	if len(got) == 0 || len(got) > 2 {
		t.Fatalf("actions = %+v; want the first and then the latest", got)
	}
	if last := got[len(got)-1]; last.Parameter != "step 5" {
		t.Fatalf("last action = %+v; want the latest call", last)
	}
}

func TestConsecutiveDuplicateActionsAreDropped(t *testing.T) {
	fx := newProgressFixture(t, 10*time.Millisecond)
	fx.runner.emit = func(onEvent runner.EventCallback, _ runner.CompactionCallback) {
		for range 3 {
			onEvent(callEvent("read_file", map[string]any{"path": "main.go"}))
			time.Sleep(30 * time.Millisecond)
		}
	}
	fx.handle(t, fx.event(ActionCreated))
	if got := actions(fx.linear.Activities()); len(got) != 1 {
		t.Fatalf("actions = %+v; want one", got)
	}
}

func TestActionsNameTheToolAndSummarizeItsArguments(t *testing.T) {
	fx := newProgressFixture(t, 10*time.Millisecond)
	fx.runner.emit = func(onEvent runner.EventCallback, _ runner.CompactionCallback) {
		onEvent(callEvent("bash", map[string]any{"command": "go test\n  ./..."}))
		time.Sleep(30 * time.Millisecond)
		onEvent(callEvent("jira_lookup", map[string]any{"key": "OPS-7"}))
		time.Sleep(30 * time.Millisecond)
		onEvent(callEvent("bash", map[string]any{"command": "curl -H 'Authorization: Bearer abcdefghijklmnop1234567890' " + strings.Repeat("x", 400)}))
		time.Sleep(30 * time.Millisecond)
	}
	fx.handle(t, fx.event(ActionCreated))
	got := actions(fx.linear.Activities())
	if len(got) != 3 {
		t.Fatalf("actions = %+v; want three", got)
	}
	if got[0].Action != "Running" || got[0].Parameter != "go test ./..." {
		t.Errorf("bash action = %+v; want Running, one-line command", got[0])
	}
	if got[1].Action != "Using jira_lookup" || got[1].Parameter != `{"key":"OPS-7"}` {
		t.Errorf("unknown tool action = %+v", got[1])
	}
	if strings.Contains(got[2].Parameter, "abcdefghijklmnop") {
		t.Errorf("parameter leaked a credential: %q", got[2].Parameter)
	}
	if n := len([]rune(got[2].Parameter)); n > 200 {
		t.Errorf("parameter is %d runes, want at most 200", n)
	}
}

func TestReasoningAndTextDeltasNeverReachLinear(t *testing.T) {
	fx := newProgressFixture(t, 10*time.Millisecond)
	fx.runner.emit = func(onEvent runner.EventCallback, _ runner.CompactionCallback) {
		thought := session.NewEvent(context.Background(), "inv")
		thought.Content = &genai.Content{Role: genai.RoleModel, Parts: []*genai.Part{{Text: "secret plan", Thought: true}}}
		onEvent(thought)
		delta := session.NewEvent(context.Background(), "inv")
		delta.Partial = true
		delta.Content = genai.NewContentFromText("partial words", genai.RoleModel)
		onEvent(delta)
		time.Sleep(30 * time.Millisecond)
	}
	fx.handle(t, fx.event(ActionCreated))
	for _, a := range fx.linear.Activities() {
		if strings.Contains(a.Body, "secret plan") || strings.Contains(a.Body, "partial words") {
			t.Fatalf("activity leaked model text: %+v", a)
		}
	}
	if got := strings.Join(activityTypes(fx.linear.Activities()), ","); got != "thought,response" {
		t.Fatalf("activities = %s, want only the acknowledgement and the response", got)
	}
}

func TestCompactionIsAnEphemeralThought(t *testing.T) {
	fx := newProgressFixture(t, 10*time.Millisecond)
	fx.runner.emit = func(_ runner.EventCallback, onCompaction runner.CompactionCallback) {
		onCompaction("Support Agent")
		time.Sleep(30 * time.Millisecond)
	}
	fx.handle(t, fx.event(ActionCreated))
	acts := fx.linear.Activities()
	if len(acts) != 3 || acts[1].Type != linearapi.ActivityThought || !acts[1].Ephemeral || !strings.Contains(acts[1].Body, "Compacting") {
		t.Fatalf("activities = %+v; want an ephemeral compaction thought between ack and response", acts)
	}
}

func TestNoProgressArrivesAfterTheFinalActivity(t *testing.T) {
	fx := newProgressFixture(t, time.Second)
	fx.runner.emit = func(onEvent runner.EventCallback, _ runner.CompactionCallback) {
		onEvent(callEvent("bash", map[string]any{"command": "first"}))
		time.Sleep(20 * time.Millisecond)
		// Offered while the interval is still running: must be dropped.
		onEvent(callEvent("bash", map[string]any{"command": "second"}))
	}
	fx.handle(t, fx.event(ActionCreated))
	time.Sleep(50 * time.Millisecond)
	acts := fx.linear.Activities()
	if last := acts[len(acts)-1]; last.Type != linearapi.ActivityResponse {
		t.Fatalf("last activity = %+v; want the response", last)
	}
	for _, a := range actions(acts) {
		if a.Parameter == "second" {
			t.Fatal("an action posted after the response")
		}
	}
}

func TestASlowLinearAPIDoesNotSlowTheTurn(t *testing.T) {
	fx := newProgressFixture(t, 10*time.Millisecond)
	fx.linear.OnActivity(func(a lineartest.Activity) *lineartest.Failure {
		if a.Type == linearapi.ActivityAction {
			time.Sleep(300 * time.Millisecond)
		}
		return nil
	})
	fx.runner.emit = func(onEvent runner.EventCallback, _ runner.CompactionCallback) {
		for i := range 3 {
			onEvent(callEvent("bash", map[string]any{"command": strings.Repeat("x", i+1)}))
		}
	}
	fx.handle(t, fx.event(ActionCreated))
	fx.runner.mu.Lock()
	ran := fx.runner.ran
	fx.runner.mu.Unlock()
	if ran > 100*time.Millisecond {
		t.Fatalf("the turn took %v while Linear was slow; progress must not block the run", ran)
	}
}
