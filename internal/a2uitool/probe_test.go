package a2uitool_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"google.golang.org/adk/v2/agent"
	adkrunner "google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/a2ui"
	internalagent "go.orx.me/apps/butter/internal/agent"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// TestRenderUIProbe runs render_ui against real models through Butter's own
// agent construction and reports how their calls fare. It is a measurement,
// not a regression test, and is skipped unless BUTTER_A2UI_PROBE names the
// targets:
//
//	BUTTER_A2UI_PROBE=gemini:gemini-3.5-flash,openai:gpt-5.1 \
//	go test ./internal/a2uitool/ -run TestRenderUIProbe -v -timeout 60m
//
// Keys come from GEMINI_API_KEY (or GOOGLE_API_KEY) and OPENAI_API_KEY;
// GEMINI_BASE_URL and OPENAI_BASE_URL point a provider elsewhere.
// BUTTER_A2UI_PROBE_RUNS sets the repetitions per scenario (default 3) and
// BUTTER_A2UI_PROBE_OUT writes every turn and call as JSON.
func TestRenderUIProbe(t *testing.T) {
	spec := os.Getenv("BUTTER_A2UI_PROBE")
	if spec == "" {
		t.Skip("set BUTTER_A2UI_PROBE to run render_ui against real models")
	}
	reps := 3
	if v := os.Getenv("BUTTER_A2UI_PROBE_RUNS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			t.Fatalf("BUTTER_A2UI_PROBE_RUNS=%q: want a positive integer", v)
		}
		reps = n
	}

	var turns []probeTurn
	for _, target := range strings.Split(spec, ",") {
		target = strings.TrimSpace(target)
		kind, model, ok := strings.Cut(target, ":")
		if !ok || model == "" {
			t.Fatalf("target %q: want provider:model", target)
		}
		providers := probeProviders(t, kind, model)
		for _, sc := range probeScenarios {
			for rep := 1; rep <= reps; rep++ {
				got := runProbeScenario(t, providers, target, model, sc, rep)
				for _, turn := range got {
					t.Logf("%s %s#%d turn %d: %d call(s), card=%v, text=%v %s",
						target, sc.name, rep, turn.Turn, len(turn.Calls), turn.cardShown(), turn.Text != "", turn.Err)
				}
				turns = append(turns, got...)
			}
		}
	}

	t.Log("\n" + summarizeProbe(turns))
	if out := os.Getenv("BUTTER_A2UI_PROBE_OUT"); out != "" {
		raw, err := json.MarshalIndent(turns, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(out, raw, 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

const probeInstruction = `You are an operations assistant in a chat app that can show cards. ` +
	`When your answer is a structured result, such as a summary with key facts, a status, or a short comparison, ` +
	`show it with the render_ui tool and also answer in one or two sentences of text. ` +
	`When information you already showed in a card changes, update that card instead of creating a new one.`

type probeScenario struct {
	name  string
	turns []string
}

var probeScenarios = []probeScenario{
	{"deploy-summary", []string{
		"Summarize this deploy for me: service api-gateway, version 2.4.1, region us-east-1, 3 of 3 replicas healthy, finished 10 minutes ago.",
	}},
	{"job-status", []string{
		"The nightly ETL job finished at 02:14. It processed 1,204 rows with 3 warnings and no errors. What is its status?",
	}},
	{"plan-comparison", []string{
		"Compare our plans: Free has 1 seat and 1 GB; Pro has 5 seats, 50 GB and costs $12 a month; Team has 20 seats, 500 GB and costs $49 a month.",
	}},
	{"on-call", []string{
		"Who is on call this week? Primary is Alice Chen, secondary is Bob Diaz, the escalation channel is #ops-escalation, and the rotation ends Sunday 18:00.",
	}},
	{"incident-update", []string{
		"Open an incident card: checkout latency is elevated in eu-west-1, p95 is 2.3 s, severity SEV-2, owner Dana.",
		"Update: the incident is resolved and p95 is back to 180 ms. Update the card you showed.",
	}},
}

func probeProviders(t *testing.T, kind, model string) []agentsv1.ModelProvider {
	t.Helper()
	models := []*agentsv1.ModelConfig{{Name: model}}
	switch kind {
	case "gemini":
		key := os.Getenv("GEMINI_API_KEY")
		if key == "" {
			key = os.Getenv("GOOGLE_API_KEY")
		}
		if key == "" {
			t.Fatal("gemini target: set GEMINI_API_KEY or GOOGLE_API_KEY")
		}
		return []agentsv1.ModelProvider{{Name: "probe-gemini", Type: "gemini", ApiKey: key, BaseUrl: os.Getenv("GEMINI_BASE_URL"), Models: models}}
	case "openai":
		key := os.Getenv("OPENAI_API_KEY")
		if key == "" {
			t.Fatal("openai target: set OPENAI_API_KEY")
		}
		return []agentsv1.ModelProvider{{Name: "probe-openai", Type: "openai", ApiKey: key, BaseUrl: os.Getenv("OPENAI_BASE_URL"), Models: models}}
	default:
		t.Fatalf("provider %q: want gemini or openai", kind)
		return nil
	}
}

// probeTurn is one user message and every render_ui call the agent made
// answering it.
type probeTurn struct {
	Target   string      `json:"target"`
	Scenario string      `json:"scenario"`
	Rep      int         `json:"rep"`
	Turn     int         `json:"turn"`
	Calls    []probeCall `json:"calls"`
	Text     string      `json:"text"`
	Err      string      `json:"err,omitempty"`
	Duration string      `json:"duration"`
}

type probeCall struct {
	Args map[string]any `json:"args"`
	// EmptyMessages counts messages the model sent as {}.
	EmptyMessages int    `json:"empty_messages,omitempty"`
	MessagesType  string `json:"messages_type"`
	OK            bool   `json:"ok"`
	Status        string `json:"status,omitempty"`
	SurfaceID     string `json:"surface_id,omitempty"`
	Error         string `json:"error,omitempty"`
	Category      string `json:"category,omitempty"`
	responded     bool
}

func (t probeTurn) cardShown() bool {
	for _, c := range t.Calls {
		if c.OK && c.Status != "deleted" {
			return true
		}
	}
	return false
}

func runProbeScenario(t *testing.T, providers []agentsv1.ModelProvider, target, model string, sc probeScenario, rep int) []probeTurn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	pb := &agentsv1.Agent{
		Name: "prober", AgentId: "prober", WorkspaceId: "ws-probe",
		Config: &agentsv1.AgentConfig{Model: model, Instruction: probeInstruction},
	}
	ag, err := internalagent.NewFromProto(ctx, pb, providers, nil, nil, nil)
	if err != nil {
		t.Fatalf("%s: build agent: %v", target, err)
	}
	r, err := adkrunner.New(adkrunner.Config{AppName: "probe", Agent: ag, SessionService: session.InMemoryService(), AutoCreateSession: true})
	if err != nil {
		t.Fatalf("%s: build runner: %v", target, err)
	}
	sessionID := fmt.Sprintf("%s-%d", sc.name, rep)

	var turns []probeTurn
	for i, prompt := range sc.turns {
		turn := probeTurn{Target: target, Scenario: sc.name, Rep: rep, Turn: i + 1}
		runCtx := a2ui.WithRun(ctx, &a2ui.Run{ThreadID: sessionID, RunID: fmt.Sprintf("run-%d", i+1), MessageID: fmt.Sprintf("msg-%d", i+1)})
		calls := map[string]*probeCall{}
		var order []string
		start := time.Now()
		msg := genai.NewContentFromText(prompt, genai.RoleUser)
		for ev, err := range r.Run(runCtx, "probe-user", sessionID, msg, agent.RunConfig{StreamingMode: agent.StreamingModeSSE}) {
			if err != nil {
				turn.Err = err.Error()
				break
			}
			if ev == nil || ev.Partial || ev.Content == nil {
				continue
			}
			for _, p := range ev.Content.Parts {
				switch {
				case p.FunctionCall != nil && p.FunctionCall.Name == "render_ui":
					id := p.FunctionCall.ID
					if id == "" {
						id = fmt.Sprintf("seq-%d", len(order))
					}
					calls[id] = newProbeCall(p.FunctionCall.Args)
					order = append(order, id)
				case p.FunctionResponse != nil && p.FunctionResponse.Name == "render_ui":
					c := calls[p.FunctionResponse.ID]
					if c == nil {
						// No ID to match on: answer the oldest open call.
						for _, id := range order {
							if !calls[id].responded {
								c = calls[id]
								break
							}
						}
					}
					if c != nil {
						c.record(p.FunctionResponse.Response)
					}
				case p.Text != "" && !p.Thought:
					turn.Text += p.Text
				}
			}
		}
		turn.Duration = time.Since(start).Round(time.Millisecond).String()
		for _, id := range order {
			turn.Calls = append(turn.Calls, *calls[id])
		}
		turns = append(turns, turn)
	}
	return turns
}

func newProbeCall(args map[string]any) *probeCall {
	c := &probeCall{Args: args, MessagesType: fmt.Sprintf("%T", args["messages"])}
	if list, ok := args["messages"].([]any); ok {
		for _, item := range list {
			if m, ok := item.(map[string]any); ok && len(m) == 0 {
				c.EmptyMessages++
			}
		}
	}
	return c
}

func (c *probeCall) record(resp map[string]any) {
	c.responded = true
	if msg, ok := resp["error"].(string); ok {
		c.Error = msg
		c.Category = probeCategory(msg)
		return
	}
	c.OK = true
	c.Status, _ = resp["status"].(string)
	c.SurfaceID, _ = resp["surface_id"].(string)
}

// probeCategories maps a phrase of a render_ui or ADK error to a category,
// checked in order.
var probeCategories = []struct{ phrase, category string }{
	{"for the rest of this turn", "stopped"},
	{"a message must be one of", "empty_or_unknown_message"},
	{"unknown member", "empty_or_unknown_message"},
	{"which does not exist", "dangling_reference"},
	{`"root"`, "missing_root"},
	{"unknown component", "unknown_component"},
	{"is not available to render_ui", "input_component"},
	{"has no property", "unknown_property"},
	{"requires", "missing_property"},
	{"must be one of", "bad_enum"},
	{"must be plain text", "html"},
	{"must not contain a URL", "url"},
	{"fallback is required", "missing_fallback"},
	{"is not a card in this conversation", "unknown_surface"},
	{"createSurface is not accepted", "create_surface"},
	{"its own ancestor", "cycle"},
	{"templates are not supported", "template"},
}

func probeCategory(msg string) string {
	for _, c := range probeCategories {
		if strings.Contains(msg, c.phrase) {
			return c.category
		}
	}
	return "other"
}

func summarizeProbe(turns []probeTurn) string {
	type stats struct {
		turns, turnsWithCalls, cards, firstTry, noText, stopped, errors int
		calls, okCalls, emptyCalls, nonArray                            int
		updateTurns, updatedSame                                        int
		categories                                                      map[string]int
	}
	byTarget := map[string]*stats{}
	var targets []string
	created := map[string]string{} // target/scenario#rep -> surface created in turn 1
	for _, turn := range turns {
		s := byTarget[turn.Target]
		if s == nil {
			s = &stats{categories: map[string]int{}}
			byTarget[turn.Target] = s
			targets = append(targets, turn.Target)
		}
		key := fmt.Sprintf("%s/%s#%d", turn.Target, turn.Scenario, turn.Rep)
		s.turns++
		if turn.Err != "" {
			s.errors++
		}
		if turn.Text == "" {
			s.noText++
		}
		if len(turn.Calls) > 0 {
			s.turnsWithCalls++
			if turn.Calls[0].OK {
				s.firstTry++
			}
		}
		if turn.cardShown() {
			s.cards++
		}
		for _, c := range turn.Calls {
			s.calls++
			if c.OK {
				s.okCalls++
				if turn.Turn == 1 && c.Status == "created" && created[key] == "" {
					created[key] = c.SurfaceID
				}
			} else {
				s.categories[c.Category]++
			}
			if c.EmptyMessages > 0 {
				s.emptyCalls++
			}
			if c.MessagesType != "[]interface {}" {
				s.nonArray++
			}
			if c.Category == "stopped" {
				s.stopped++
			}
		}
		if turn.Scenario == "incident-update" && turn.Turn == 2 {
			s.updateTurns++
			for _, c := range turn.Calls {
				if c.OK && c.Status == "updated" && c.SurfaceID != "" && c.SurfaceID == created[key] {
					s.updatedSame++
					break
				}
			}
		}
	}

	var b strings.Builder
	for _, target := range targets {
		s := byTarget[target]
		fmt.Fprintf(&b, "== %s\n", target)
		fmt.Fprintf(&b, "turns %d (run errors %d, no text answer %d)\n", s.turns, s.errors, s.noText)
		fmt.Fprintf(&b, "turns calling render_ui %d, showing a card %d, first call valid %d\n", s.turnsWithCalls, s.cards, s.firstTry)
		fmt.Fprintf(&b, "calls %d, valid %d, with {} messages %d, messages not an array %d, refused after the limit %d\n", s.calls, s.okCalls, s.emptyCalls, s.nonArray, s.stopped)
		fmt.Fprintf(&b, "incident updates reusing the card %d/%d\n", s.updatedSame, s.updateTurns)
		var cats []string
		for cat, n := range s.categories {
			cats = append(cats, fmt.Sprintf("%s=%d", cat, n))
		}
		sort.Strings(cats)
		fmt.Fprintf(&b, "failures: %s\n", strings.Join(cats, ", "))
	}
	return b.String()
}
