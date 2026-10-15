package main

// The mobile ButterBox screens and the Pi/Cursor leaf editor (#453) talk only
// to the Butter server. These tests drive the production ButterBoxService,
// AgentService leaf validation, runner and the real pibox and cursorbox
// bridges through the fixture's HTTP handlers, with an in-process Box test
// double (boxes.go) standing in for the box. They show what the editor's
// configuration does on a real run; they are not a production Box or a device
// run.

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

type boxHarness struct {
	*providerFixture
	boxes  agentsv1connect.ButterBoxServiceClient
	agents agentsv1connect.AgentServiceClient
}

func newBoxHarness(t *testing.T) *boxHarness {
	t.Helper()
	f := newProviderFixture(t)
	interceptor := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+f.token)
			req.Header().Set("X-Workspace-ID", boxWorkspace)
			return next(ctx, req)
		}
	})
	return &boxHarness{
		providerFixture: f,
		boxes:           agentsv1connect.NewButterBoxServiceClient(http.DefaultClient, f.url+"/api", connect.WithInterceptors(interceptor)),
		agents:          f.agentClient(boxWorkspace),
	}
}

func (h *boxHarness) register(t *testing.T, name, path, token string, enabled bool) *agentsv1.ButterBox {
	t.Helper()
	resp, err := h.boxes.CreateButterBox(context.Background(), connect.NewRequest(&agentsv1.CreateButterBoxRequest{
		Name: name, BaseUrl: h.url + path, Enabled: enabled, Token: token,
	}))
	if err != nil {
		t.Fatalf("register %s: %v", name, err)
	}
	return resp.Msg.GetButterBox()
}

func (h *boxHarness) saveAgent(t *testing.T, agent *agentsv1.Agent) (*agentsv1.Agent, error) {
	t.Helper()
	resp, err := h.agents.CreateAgent(context.Background(), connect.NewRequest(&agentsv1.CreateAgentRequest{Agent: agent}))
	if err != nil {
		return nil, err
	}
	return resp.Msg.GetAgent(), nil
}

func piLeaf(id, boxID string, mutate func(*agentsv1.PiAgentConfig)) *agentsv1.Agent {
	pi := &agentsv1.PiAgentConfig{ButterboxId: boxID}
	if mutate != nil {
		mutate(pi)
	}
	return &agentsv1.Agent{
		AgentId: id, Name: id, Type: agentsv1.AgentType_AGENT_TYPE_PI,
		Config: &agentsv1.AgentConfig{Pi: pi},
	}
}

func cursorLeaf(id, boxID string, mutate func(*agentsv1.CursorAgentConfig)) *agentsv1.Agent {
	cursor := &agentsv1.CursorAgentConfig{ButterboxId: boxID}
	if mutate != nil {
		mutate(cursor)
	}
	return &agentsv1.Agent{
		AgentId: id, Name: id, Type: agentsv1.AgentType_AGENT_TYPE_CURSOR,
		Config: &agentsv1.AgentConfig{Cursor: cursor},
	}
}

// run sends one message through the production AG-UI handler and returns the
// assistant text and the run error, if any.
func (h *boxHarness) run(t *testing.T, agentID, thread, text string) (reply, runError string) {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"threadId": thread, "runId": "run-" + thread + "-" + fmt.Sprint(len(text)),
		"messages": []map[string]any{{"id": "m1", "role": "user", "content": text}},
		"tools":    []any{}, "context": []any{}, "state": nil,
	})
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, h.url+"/api/agui/"+agentID, bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.token)
	req.Header.Set("X-Workspace-ID", boxWorkspace)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("run %s: status %d", agentID, resp.StatusCode)
	}
	var out strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	for scanner.Scan() {
		data, ok := strings.CutPrefix(scanner.Text(), "data: ")
		if !ok {
			continue
		}
		var ev map[string]any
		if json.Unmarshal([]byte(data), &ev) != nil {
			continue
		}
		switch ev["type"] {
		case "TEXT_MESSAGE_CONTENT":
			out.WriteString(fmt.Sprint(ev["delta"]))
		case "RUN_ERROR":
			runError = fmt.Sprint(ev["message"])
		}
	}
	return out.String(), runError
}

func TestBoxTokenIsWriteOnlyAndKeepReplaceClearAreDistinct(t *testing.T) {
	h := newBoxHarness(t)
	ctx := context.Background()
	box := h.register(t, "laptop", "/fake-box/ok", boxToken, true)
	if !box.GetCredentialSet() {
		t.Fatal("a registered token must show as set")
	}
	if bytes.Contains(mustMarshal(t, box), []byte(boxToken)) {
		t.Fatal("the token appeared in the box message")
	}

	status, err := h.boxes.GetButterBoxStatus(ctx, connect.NewRequest(&agentsv1.GetButterBoxStatusRequest{Id: box.GetId()}))
	if err != nil || !status.Msg.GetReachable() {
		t.Fatalf("status with the token = %v, %v", status, err)
	}

	// An edit that carries no token keeps it: the box still answers.
	updated, err := h.boxes.UpdateButterBox(ctx, connect.NewRequest(&agentsv1.UpdateButterBoxRequest{
		Id: box.GetId(), Name: "laptop 2", BaseUrl: box.GetBaseUrl(), Enabled: true,
	}))
	if err != nil || !updated.Msg.GetButterBox().GetCredentialSet() {
		t.Fatalf("update dropped the token: %v, %v", updated, err)
	}
	status, _ = h.boxes.GetButterBoxStatus(ctx, connect.NewRequest(&agentsv1.GetButterBoxStatusRequest{Id: box.GetId()}))
	if !status.Msg.GetReachable() {
		t.Fatalf("the kept token no longer works: %s", status.Msg.GetError())
	}

	// A wrong replacement is a replacement: the box now refuses.
	if _, err := h.boxes.SetButterBoxToken(ctx, connect.NewRequest(&agentsv1.SetButterBoxTokenRequest{Id: box.GetId(), Token: "wrong"})); err != nil {
		t.Fatal(err)
	}
	status, _ = h.boxes.GetButterBoxStatus(ctx, connect.NewRequest(&agentsv1.GetButterBoxStatusRequest{Id: box.GetId()}))
	if status.Msg.GetReachable() || status.Msg.GetError() == "" {
		t.Fatalf("a wrong token must read as unreachable with a reason: %+v", status.Msg)
	}

	// Clearing is its own explicit write (an empty token).
	cleared, err := h.boxes.SetButterBoxToken(ctx, connect.NewRequest(&agentsv1.SetButterBoxTokenRequest{Id: box.GetId()}))
	if err != nil || cleared.Msg.GetButterBox().GetCredentialSet() {
		t.Fatalf("clear = %v, %v", cleared, err)
	}
}

func TestBoxUnreachableIsDataAndCatalogErrorsAreActionable(t *testing.T) {
	h := newBoxHarness(t)
	ctx := context.Background()

	dead := h.register(t, "dead", "", "", true)
	// Port 1 is closed on loopback: nothing answers.
	if _, err := h.boxes.UpdateButterBox(ctx, connect.NewRequest(&agentsv1.UpdateButterBoxRequest{
		Id: dead.GetId(), Name: "dead", BaseUrl: "http://127.0.0.1:1", Enabled: true,
	})); err != nil {
		t.Fatal(err)
	}
	status, err := h.boxes.GetButterBoxStatus(ctx, connect.NewRequest(&agentsv1.GetButterBoxStatusRequest{Id: dead.GetId()}))
	if err != nil || status.Msg.GetReachable() || status.Msg.GetError() == "" {
		t.Fatalf("an unreachable box is an answer: %+v, %v", status, err)
	}

	ok := h.register(t, "ok", "/fake-box/ok", boxToken, true)
	pi, err := h.boxes.ListButterBoxModels(ctx, connect.NewRequest(&agentsv1.ListButterBoxModelsRequest{Id: ok.GetId()}))
	if err != nil || len(pi.Msg.GetModels()) != 2 || pi.Msg.GetModels()[0].GetId() != "claude-fable-5" {
		t.Fatalf("pi catalog = %v, %v", pi, err)
	}
	cursor, err := h.boxes.ListButterBoxCursorModels(ctx, connect.NewRequest(&agentsv1.ListButterBoxCursorModelsRequest{Id: ok.GetId()}))
	if err != nil || len(cursor.Msg.GetModels()) != 2 || cursor.Msg.GetModels()[0].GetId() != "composer-2.5" {
		t.Fatalf("cursor catalog = %v, %v", cursor, err)
	}

	full := h.register(t, "full", "/fake-box/capacity", "", true)
	_, err = h.boxes.ListButterBoxModels(ctx, connect.NewRequest(&agentsv1.ListButterBoxModelsRequest{Id: full.GetId()}))
	if code := connect.CodeOf(err); code != connect.CodeResourceExhausted {
		t.Fatalf("a box at capacity = %v (%v)", code, err)
	}

	nokey := h.register(t, "nokey", "/fake-box/nokey", "", true)
	_, err = h.boxes.ListButterBoxCursorModels(ctx, connect.NewRequest(&agentsv1.ListButterBoxCursorModelsRequest{Id: nokey.GetId()}))
	if connect.CodeOf(err) != connect.CodeFailedPrecondition || !strings.Contains(err.Error(), "CURSOR_API_KEY") {
		t.Fatalf("a missing Cursor key must say so: %v", err)
	}
}

func TestLeafWritesFollowTheServerContract(t *testing.T) {
	h := newBoxHarness(t)
	ctx := context.Background()
	live := h.register(t, "live", "/fake-box/ok", boxToken, true)
	retired := h.register(t, "retired", "/fake-box/ok", boxToken, false)

	// Butter-side behavior fields are refused on a leaf; the editor never
	// sends them, this proves it would be told if it did.
	withInstruction := piLeaf("pi-bad", live.GetId(), nil)
	withInstruction.Config.Instruction = "be nice"
	_, err := h.saveAgent(t, withInstruction)
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("instruction on a leaf = %v", err)
	}

	// The editor's own shape saves, and the stored leaf round-trips with the
	// three run-limit forms kept apart.
	for _, tc := range []struct {
		id      string
		limit   *int32
		present bool
	}{
		{"pi-default", nil, false},
		{"pi-unlimited", proto.Int32(0), true},
		{"pi-timed", proto.Int32(90), true},
	} {
		saved, err := h.saveAgent(t, piLeaf(tc.id, live.GetId(), func(c *agentsv1.PiAgentConfig) {
			c.WorkingDir, c.Provider, c.Model, c.ThinkingLevel = "/srv/repo", "anthropic", "claude-fable-5", "high"
			c.MaxRunSeconds = tc.limit
		}))
		if err != nil {
			t.Fatalf("save %s: %v", tc.id, err)
		}
		pi := saved.GetConfig().GetPi()
		if pi.MaxRunSeconds != nil != tc.present || (tc.limit != nil && pi.GetMaxRunSeconds() != *tc.limit) {
			t.Fatalf("%s: limit present=%v value=%d, want present=%v %v", tc.id, pi.MaxRunSeconds != nil, pi.GetMaxRunSeconds(), tc.present, tc.limit)
		}
		if pi.GetProvider() != "anthropic" || pi.GetModel() != "claude-fable-5" || pi.GetThinkingLevel() != "high" || pi.GetWorkingDir() != "/srv/repo" {
			t.Fatalf("%s lost a field: %v", tc.id, pi)
		}
	}

	// A box that does not exist, and one that is disabled, take no new binding.
	_, err = h.saveAgent(t, piLeaf("pi-missing", "no-such-box", nil))
	if err == nil {
		t.Fatal("a leaf pointing at an unknown box was accepted")
	}
	_, err = h.saveAgent(t, cursorLeaf("cursor-retired", retired.GetId(), nil))
	if err == nil {
		t.Fatal("a new binding to a disabled box was accepted")
	}

	// The box is now referenced: deleting it names the Agents.
	_, err = h.boxes.DeleteButterBox(ctx, connect.NewRequest(&agentsv1.DeleteButterBoxRequest{Id: live.GetId()}))
	if connect.CodeOf(err) != connect.CodeFailedPrecondition || !strings.Contains(err.Error(), "pi-timed") {
		t.Fatalf("delete of a referenced box = %v", err)
	}
	// An unreferenced one goes.
	if _, err := h.boxes.DeleteButterBox(ctx, connect.NewRequest(&agentsv1.DeleteButterBoxRequest{Id: retired.GetId()})); err != nil {
		t.Fatalf("delete of an unreferenced box: %v", err)
	}
}

func TestPiLeafRunsKeepsASessionAndRepointsFresh(t *testing.T) {
	h := newBoxHarness(t)
	box := h.register(t, "live", "/fake-box/ok", boxToken, true)
	if _, err := h.saveAgent(t, piLeaf("pi-runner", box.GetId(), func(c *agentsv1.PiAgentConfig) {
		c.WorkingDir, c.Provider, c.Model, c.ThinkingLevel = "/srv/a", "anthropic", "claude-fable-5", "high"
	})); err != nil {
		t.Fatal(err)
	}

	first, runErr := h.run(t, "pi-runner", "thread-1", "hello")
	if runErr != "" || !strings.Contains(first, "cwd=/srv/a") || !strings.Contains(first, "model=anthropic/claude-fable-5") ||
		!strings.Contains(first, "thinking=high") || !strings.HasSuffix(first, "(hello)") {
		t.Fatalf("first turn = %q, %q", first, runErr)
	}
	session := between(first, "session=", ",")
	second, runErr := h.run(t, "pi-runner", "thread-1", "again")
	if runErr != "" || !strings.Contains(second, "session="+session+",") {
		t.Fatalf("a conversation keeps its box session: %q, %q", second, runErr)
	}
	other, _ := h.run(t, "pi-runner", "thread-2", "elsewhere")
	if strings.Contains(other, "session="+session+",") {
		t.Fatalf("another conversation shares a session: %q", other)
	}

	// Repointing the working directory leaves the old session behind and
	// starts a fresh one on the next message, as the editor's note says.
	stored, err := h.agents.GetAgent(context.Background(), connect.NewRequest(&agentsv1.GetAgentRequest{AgentId: "pi-runner"}))
	if err != nil {
		t.Fatal(err)
	}
	edit := proto.Clone(stored.Msg.GetAgent()).(*agentsv1.Agent)
	edit.Config.Pi.WorkingDir = "/srv/b"
	if _, err := h.agents.UpdateAgent(context.Background(), connect.NewRequest(&agentsv1.UpdateAgentRequest{Agent: edit})); err != nil {
		t.Fatalf("repoint: %v", err)
	}
	moved, runErr := h.run(t, "pi-runner", "thread-1", "after repoint")
	if runErr != "" || !strings.Contains(moved, "cwd=/srv/b") || strings.Contains(moved, "session="+session+",") {
		t.Fatalf("a repointed leaf must start a fresh session: %q, %q", moved, runErr)
	}
}

func TestCursorLeafRunsAndFailuresAreActionable(t *testing.T) {
	h := newBoxHarness(t)
	live := h.register(t, "live", "/fake-box/ok", boxToken, true)
	full := h.register(t, "full", "/fake-box/capacity", "", true)
	nokey := h.register(t, "nokey", "/fake-box/nokey", "", true)
	dead := h.register(t, "dead", "/fake-box/ok", "", true)
	// The saved box now points nowhere; the leaf is bound while it was valid.
	if _, err := h.saveAgent(t, cursorLeaf("cursor-dead", dead.GetId(), nil)); err != nil {
		t.Fatal(err)
	}
	if _, err := h.boxes.UpdateButterBox(context.Background(), connect.NewRequest(&agentsv1.UpdateButterBoxRequest{
		Id: dead.GetId(), Name: "dead", BaseUrl: "http://127.0.0.1:1", Enabled: true,
	})); err != nil {
		t.Fatal(err)
	}

	for _, leaf := range []*agentsv1.Agent{
		cursorLeaf("cursor-live", live.GetId(), func(c *agentsv1.CursorAgentConfig) {
			c.WorkingDir, c.Model, c.Mode = "/srv/c", "composer-2.5", "plan"
		}),
		cursorLeaf("cursor-full", full.GetId(), nil),
		cursorLeaf("cursor-nokey", nokey.GetId(), nil),
	} {
		if _, err := h.saveAgent(t, leaf); err != nil {
			t.Fatalf("save %s: %v", leaf.GetAgentId(), err)
		}
	}

	reply, runErr := h.run(t, "cursor-live", "c1", "build it")
	if runErr != "" || !strings.Contains(reply, "cwd=/srv/c") || !strings.Contains(reply, "model=composer-2.5") ||
		!strings.Contains(reply, "mode=plan") || !strings.HasSuffix(reply, "(build it)") {
		t.Fatalf("cursor turn = %q, %q", reply, runErr)
	}
	session := between(reply, "session=", ",")
	again, _ := h.run(t, "cursor-live", "c1", "more")
	if !strings.Contains(again, "session="+session+",") {
		t.Fatalf("cursor conversation lost its session: %q", again)
	}

	for _, tc := range []struct {
		agent string
		want  string
	}{
		{"cursor-full", "capacity"},
		{"cursor-nokey", "CURSOR_API_KEY"},
		{"cursor-dead", "unreachable"},
	} {
		got, runErr := h.run(t, tc.agent, "fail-"+tc.agent, "go")
		if runErr == "" || !strings.Contains(runErr, tc.want) || got != "" {
			t.Fatalf("%s: reply %q, error %q, want an error mentioning %q", tc.agent, got, runErr, tc.want)
		}
	}
}

func TestPiLeafFailuresAreActionable(t *testing.T) {
	h := newBoxHarness(t)
	full := h.register(t, "full", "/fake-box/capacity", "", true)
	if _, err := h.saveAgent(t, piLeaf("pi-full", full.GetId(), nil)); err != nil {
		t.Fatal(err)
	}
	got, runErr := h.run(t, "pi-full", "p1", "go")
	if runErr == "" || !strings.Contains(runErr, "capacity") || got != "" {
		t.Fatalf("reply %q, error %q, want a capacity error", got, runErr)
	}
}

func between(s, from, to string) string {
	_, rest, ok := strings.Cut(s, from)
	if !ok {
		return ""
	}
	value, _, _ := strings.Cut(rest, to)
	return value
}

func mustMarshal(t *testing.T, m proto.Message) []byte {
	t.Helper()
	b, err := proto.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
