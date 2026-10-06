package a2uitool_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
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
	targets, reps := probeTargets(t)
	agent := probeAgent{instruction: probeInstruction}

	var turns []probeTurn
	for _, target := range targets {
		for _, sc := range probeScenarios {
			for rep := 1; rep <= reps; rep++ {
				got := runProbeScenario(t, target, agent, sc, rep)
				for _, turn := range got {
					t.Logf("%s %s#%d turn %d: %d call(s), card=%v, text=%v %s",
						target.name, sc.name, rep, turn.Turn, len(turn.Calls), turn.cardShown(), turn.Text != "", turn.Err)
				}
				turns = append(turns, got...)
			}
		}
	}

	t.Log("\n" + summarizeProbe(turns))
	writeProbeOut(t, turns)
}

// TestRenderUIPresentationProbe measures how often real models show a card
// for borderline prompts under the AUTO and the PREFERRED presentation of
// the Card Policy (#442). The instruction never mentions cards, so the only
// difference between the two runs is the hint in render_ui's description.
// Like TestRenderUIProbe it is a measurement, skipped unless
// BUTTER_A2UI_PROBE names the targets; -parallel bounds the turns in flight:
//
//	BUTTER_A2UI_PROBE=openai:gpt-5.6-luna \
//	go test ./internal/a2uitool/ -run TestRenderUIPresentationProbe -v -parallel 4 -timeout 60m
func TestRenderUIPresentationProbe(t *testing.T) {
	targets, reps := probeTargets(t)
	presentations := []agentsv1.ResultCardPresentation{
		agentsv1.ResultCardPresentation_RESULT_CARD_PRESENTATION_AUTO,
		agentsv1.ResultCardPresentation_RESULT_CARD_PRESENTATION_PREFERRED,
	}

	var mu sync.Mutex
	var turns []probeTurn
	// The group returns once every parallel turn in it has finished.
	t.Run("turns", func(t *testing.T) {
		for _, target := range targets {
			for _, pr := range presentations {
				agent := probeAgent{
					instruction: borderlineInstruction,
					cards:       &agentsv1.ResultCardConfig{Presentation: pr},
					variant:     strings.ToLower(strings.TrimPrefix(pr.String(), "RESULT_CARD_PRESENTATION_")),
				}
				for _, sc := range borderlineScenarios {
					for rep := 1; rep <= reps; rep++ {
						t.Run(fmt.Sprintf("%s/%s/%s#%d", target.name, agent.variant, sc.name, rep), func(t *testing.T) {
							t.Parallel()
							got := runProbeScenario(t, target, agent, sc, rep)
							for _, turn := range got {
								t.Logf("card=%v, text=%v %s", turn.cardShown(), turn.Text != "", turn.Err)
							}
							mu.Lock()
							turns = append(turns, got...)
							mu.Unlock()
						})
					}
				}
			}
		}
	})

	writeProbeOut(t, turns)
	t.Log("\n" + summarizePresentation(turns))
}

type probeTarget struct {
	name, model string
	providers   []agentsv1.ModelProvider
}

// probeTargets reads the targets of BUTTER_A2UI_PROBE and the repetitions of
// BUTTER_A2UI_PROBE_RUNS, skipping the test when no target is named.
func probeTargets(t *testing.T) ([]probeTarget, int) {
	t.Helper()
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
	var targets []probeTarget
	for _, target := range strings.Split(spec, ",") {
		target = strings.TrimSpace(target)
		kind, model, ok := strings.Cut(target, ":")
		if !ok || model == "" {
			t.Fatalf("target %q: want provider:model", target)
		}
		targets = append(targets, probeTarget{name: target, model: model, providers: probeProviders(t, kind, model)})
	}
	return targets, reps
}

func writeProbeOut(t *testing.T, turns []probeTurn) {
	t.Helper()
	out := os.Getenv("BUTTER_A2UI_PROBE_OUT")
	if out == "" {
		return
	}
	raw, err := json.MarshalIndent(turns, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(out, raw, 0o600); err != nil {
		t.Fatal(err)
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

// borderlineInstruction never mentions cards: in the presentation probe the
// hint in render_ui's description is the only nudge.
const borderlineInstruction = `You are an operations assistant in a chat app. Answer the user's questions briefly.`

// borderlineScenarios are prompts a text answer serves as well as a card:
// a little structure, nothing that asks for a view.
var borderlineScenarios = []probeScenario{
	{"sla-facts", []string{
		"Remind me of our SLA: 99.9% monthly uptime, 4-hour response for SEV-2 and 1-hour response for SEV-1.",
	}},
	{"build-times", []string{
		"Our last three builds took 4m12s, 6m03s and 5m40s. Is that normal?",
	}},
	{"regions", []string{
		"Which regions is checkout running in? It is in us-east-1, eu-west-1 and ap-southeast-2.",
	}},
	{"disk-trend", []string{
		"The disk on db-03 is at 87% and grows about 2% a day. Should I worry?",
	}},
	{"open-prs", []string{
		"We have three open PRs: #120 by Alice is ready, #121 by Bob is a draft, #124 by Carol needs review. What is left to do?",
	}},
	{"key-rotation", []string{
		"How do I rotate the billing service's API key? The old key expires on Friday.",
	}},
	{"probes", []string{
		"What is the difference between a liveness probe and a readiness probe?",
	}},
	{"p95", []string{
		"What does p95 latency mean?",
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
	Variant  string      `json:"variant,omitempty"`
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

// probeAgent is the configuration a probe runs its agent with.
type probeAgent struct {
	instruction string
	cards       *agentsv1.ResultCardConfig
	// variant names the configuration in the results.
	variant string
}

func runProbeScenario(t *testing.T, target probeTarget, cfg probeAgent, sc probeScenario, rep int) []probeTurn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	pb := &agentsv1.Agent{
		Name: "prober", AgentId: "prober", WorkspaceId: "ws-probe",
		Config: &agentsv1.AgentConfig{Model: target.model, Instruction: cfg.instruction, ResultCards: cfg.cards},
	}
	ag, err := internalagent.NewFromProto(ctx, pb, target.providers, nil, nil, nil)
	if err != nil {
		t.Fatalf("%s: build agent: %v", target.name, err)
	}
	r, err := adkrunner.New(adkrunner.Config{AppName: "probe", Agent: ag, SessionService: session.InMemoryService(), AutoCreateSession: true})
	if err != nil {
		t.Fatalf("%s: build runner: %v", target.name, err)
	}
	sessionID := fmt.Sprintf("%s-%d", sc.name, rep)

	var turns []probeTurn
	for i, prompt := range sc.turns {
		turn := probeTurn{Target: target.name, Variant: cfg.variant, Scenario: sc.name, Rep: rep, Turn: i + 1}
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

// summarizePresentation reports, per target, how many turns showed a card
// under each presentation, scenario by scenario, with the run errors that
// would void a comparison and the text answers and first-try validity that a
// hint must not cost.
func summarizePresentation(turns []probeTurn) string {
	type key struct{ target, variant, scenario string }
	type tally struct{ turns, errors, cards, text, calls, firstTry int }
	counts := map[key]*tally{}
	var targets, variants, scenarios []string
	for _, turn := range turns {
		targets = appendNew(targets, turn.Target)
		variants = appendNew(variants, turn.Variant)
		scenarios = appendNew(scenarios, turn.Scenario)
		// The empty scenario is the target's total for the variant.
		for _, k := range []key{{turn.Target, turn.Variant, turn.Scenario}, {turn.Target, turn.Variant, ""}} {
			c := counts[k]
			if c == nil {
				c = &tally{}
				counts[k] = c
			}
			c.turns++
			if turn.Err != "" {
				c.errors++
			}
			if turn.cardShown() {
				c.cards++
			}
			if turn.Text != "" {
				c.text++
			}
			if len(turn.Calls) > 0 {
				c.calls++
				if turn.Calls[0].OK {
					c.firstTry++
				}
			}
		}
	}
	sort.Strings(targets)
	sort.Strings(variants)
	sort.Strings(scenarios)

	var b strings.Builder
	for _, target := range targets {
		fmt.Fprintf(&b, "== %s: turns showing a card\n%-14s", target, "scenario")
		for _, v := range variants {
			fmt.Fprintf(&b, "%12s", v)
		}
		b.WriteString("\n")
		for _, sc := range append(scenarios, "") {
			name := sc
			if name == "" {
				name = "total"
			}
			fmt.Fprintf(&b, "%-14s", name)
			for _, v := range variants {
				c := counts[key{target, v, sc}]
				if c == nil {
					c = &tally{}
				}
				fmt.Fprintf(&b, "%12s", fmt.Sprintf("%d/%d", c.cards, c.turns))
			}
			b.WriteString("\n")
		}
		for _, v := range variants {
			if c := counts[key{target, v, ""}]; c != nil {
				fmt.Fprintf(&b, "%s: run errors %d/%d, text answers %d/%d, turns calling render_ui %d, first call valid %d\n",
					v, c.errors, c.turns, c.text, c.turns, c.calls, c.firstTry)
			}
		}
	}
	return b.String()
}

func appendNew(list []string, v string) []string {
	if slices.Contains(list, v) {
		return list
	}
	return append(list, v)
}
