package main

// The mobile Daemon screens (#455) talk only to the Butter server. These tests
// drive the production DaemonService, the production poll transport that real
// workers use, the production RemoteAgentService and the production runner with
// its daemon Bridge, through the fixture's HTTP handlers. The only double is
// the worker (daemon.go), which speaks the real worker protocol with a worker
// token. They show what the App's operations do on the real server; they are
// not an opencode or codex run and not a device run.

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/durationpb"

	"go.orx.me/apps/butter/internal/application"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

type daemonHarness struct {
	t     *testing.T
	fx    *daemonFixture
	token string
	url   string
}

func newDaemonHarness(t *testing.T) *daemonHarness {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	handler, fx, err := newFixture(ctx)
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	t.Cleanup(func() {
		cancel()
		server.Close()
	})
	login, err := agentsv1connect.NewAuthServiceClient(server.Client(), server.URL+"/api").Login(ctx,
		connect.NewRequest(&agentsv1.LoginRequest{Username: "mobile-test", Password: "contract-test-password"}))
	if err != nil {
		t.Fatal(err)
	}
	return &daemonHarness{t: t, fx: fx, token: login.Msg.GetToken(), url: server.URL}
}

func (h *daemonHarness) opts(workspaceID string) connect.ClientOption {
	return connect.WithInterceptors(connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+h.token)
			if workspaceID != "" {
				req.Header().Set("X-Workspace-ID", workspaceID)
			}
			return next(ctx, req)
		}
	}))
}

func (h *daemonHarness) daemons(workspaceID string) agentsv1connect.DaemonServiceClient {
	return agentsv1connect.NewDaemonServiceClient(http.DefaultClient, h.url+"/api", h.opts(workspaceID))
}

func (h *daemonHarness) remotes(workspaceID string) agentsv1connect.RemoteAgentServiceClient {
	return agentsv1connect.NewRemoteAgentServiceClient(http.DefaultClient, h.url+"/api", h.opts(workspaceID))
}

func (h *daemonHarness) createRuntime(workspaceID, id string) {
	h.t.Helper()
	if _, err := h.daemons(workspaceID).CreateDaemonRuntime(context.Background(), connect.NewRequest(
		&agentsv1.CreateDaemonRuntimeRequest{Runtime: &agentsv1.DaemonRuntime{Id: id, Name: id}})); err != nil {
		h.t.Fatalf("create runtime %s: %v", id, err)
	}
}

// mint returns a worker token the way the App does: one RPC, secret shown once.
func (h *daemonHarness) mint(workspaceID, runtimeID string) *agentsv1.CreateDaemonRuntimeTokenResponse {
	h.t.Helper()
	resp, err := h.daemons(workspaceID).CreateDaemonRuntimeToken(context.Background(), connect.NewRequest(
		&agentsv1.CreateDaemonRuntimeTokenRequest{DaemonRuntimeId: runtimeID, Name: "worker"}))
	if err != nil {
		h.t.Fatalf("mint token for %s: %v", runtimeID, err)
	}
	return resp.Msg
}

// control calls the fixture's loopback control API.
func (h *daemonHarness) control(method, path string, in, out any) int {
	h.t.Helper()
	var body bytes.Buffer
	if in != nil {
		if err := json.NewEncoder(&body).Encode(in); err != nil {
			h.t.Fatal(err)
		}
	}
	req, err := http.NewRequest(method, h.url+"/fake-daemon"+path, &body)
	if err != nil {
		h.t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		h.t.Fatal(err)
	}
	defer resp.Body.Close()
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			h.t.Fatalf("decode %s: %v", path, err)
		}
	}
	return resp.StatusCode
}

func (h *daemonHarness) startWorker(workspaceID, runtimeID, secret, mode string) int {
	h.t.Helper()
	return h.control("POST", "/worker/start", map[string]string{
		"workspace": workspaceID, "runtime_id": runtimeID, "token": secret, "mode": mode,
	}, nil)
}

// onlineWorker creates a runtime, mints its token and brings a worker online.
func (h *daemonHarness) onlineWorker(workspaceID, runtimeID, mode string) {
	h.t.Helper()
	h.createRuntime(workspaceID, runtimeID)
	if code := h.startWorker(workspaceID, runtimeID, h.mint(workspaceID, runtimeID).GetSecret(), mode); code != http.StatusOK {
		h.t.Fatalf("worker %s did not register: HTTP %d", runtimeID, code)
	}
	h.t.Cleanup(func() {
		h.control("POST", "/worker/drop", map[string]string{"workspace": workspaceID, "runtime_id": runtimeID}, nil)
	})
}

func (h *daemonHarness) createDaemonAgent(workspaceID, id, runtimeID string) {
	h.t.Helper()
	if _, err := h.remotes(workspaceID).CreateRemoteAgent(context.Background(), connect.NewRequest(
		&agentsv1.CreateRemoteAgentRequest{RemoteAgent: &agentsv1.RemoteAgent{
			Id: id, Name: id, Protocol: agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_DAEMON,
			DaemonRuntimeId: runtimeID, AcpRuntime: "codex",
		}})); err != nil {
		h.t.Fatalf("create daemon agent %s: %v", id, err)
	}
}

func (h *daemonHarness) dispatch(workspaceID, remoteAgentID, input string) string {
	h.t.Helper()
	var out struct {
		RunID string `json:"run_id"`
	}
	if code := h.control("POST", "/dispatch", map[string]string{
		"workspace": workspaceID, "remote_agent_id": remoteAgentID, "input": input,
	}, &out); code != http.StatusOK {
		h.t.Fatalf("dispatch: HTTP %d", code)
	}
	return out.RunID
}

type runOutcome struct {
	Done   bool   `json:"done"`
	Output string `json:"output"`
	Error  string `json:"error"`
}

func (h *daemonHarness) runOutcome(runID string) runOutcome {
	h.t.Helper()
	var out runOutcome
	h.control("GET", "/run?id="+runID, nil, &out)
	return out
}

type workerState struct {
	Tasks   []workerTask `json:"tasks"`
	Cancels []string     `json:"cancels"`
	Stopped string       `json:"stopped"`
}

func (h *daemonHarness) workerState(workspaceID, runtimeID string) workerState {
	h.t.Helper()
	var out workerState
	h.control("GET", "/worker/state?workspace="+workspaceID+"&runtime_id="+runtimeID, nil, &out)
	return out
}

func (h *daemonHarness) tasks(workspaceID string) []*agentsv1.DaemonTaskInFlight {
	h.t.Helper()
	resp, err := h.daemons(workspaceID).ListDaemonTasks(context.Background(), connect.NewRequest(&agentsv1.ListDaemonTasksRequest{}))
	if err != nil {
		h.t.Fatal(err)
	}
	return resp.Msg.GetTasks()
}

func eventually(t *testing.T, what string, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if ok() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

func (h *daemonHarness) waitForTask(workspaceID string) *agentsv1.DaemonTaskInFlight {
	h.t.Helper()
	var task *agentsv1.DaemonTaskInFlight
	eventually(h.t, "a task in flight", func() bool {
		if list := h.tasks(workspaceID); len(list) > 0 {
			task = list[0]
			return true
		}
		return false
	})
	return task
}

func TestDaemonRuntimeCRUDAndTenantIsolation(t *testing.T) {
	h := newDaemonHarness(t)
	ctx := context.Background()
	d := h.daemons(daemonWorkspace)

	created, err := d.CreateDaemonRuntime(ctx, connect.NewRequest(&agentsv1.CreateDaemonRuntimeRequest{
		Runtime: &agentsv1.DaemonRuntime{
			Id: "rt-1", Name: "  Build box  ", Description: " Linux builder ",
			Labels: map[string]string{"role": "ci"},
		},
	}))
	if err != nil {
		t.Fatal(err)
	}
	got := created.Msg.GetRuntime()
	if got.GetName() != "Build box" || got.GetDescription() != "Linux builder" ||
		got.GetWorkspaceId() != daemonWorkspace || got.GetCreatedBy() != "mobile-test" ||
		got.GetCreatedAt() == nil || got.GetLabels()["role"] != "ci" {
		t.Fatalf("created runtime = %v", got)
	}

	_, err = d.CreateDaemonRuntime(ctx, connect.NewRequest(&agentsv1.CreateDaemonRuntimeRequest{
		Runtime: &agentsv1.DaemonRuntime{Id: "rt-1", Name: "again"}}))
	wantCode(t, err, connect.CodeAlreadyExists)
	_, err = d.CreateDaemonRuntime(ctx, connect.NewRequest(&agentsv1.CreateDaemonRuntimeRequest{
		Runtime: &agentsv1.DaemonRuntime{Id: "rt-2", Name: "  "}}))
	wantCode(t, err, connect.CodeInvalidArgument)

	updated, err := d.UpdateDaemonRuntime(ctx, connect.NewRequest(&agentsv1.UpdateDaemonRuntimeRequest{
		Runtime: &agentsv1.DaemonRuntime{Id: "rt-1", Name: "Renamed"}}))
	if err != nil {
		t.Fatal(err)
	}
	// The update is a full replace of the editable fields; created_* stay.
	if u := updated.Msg.GetRuntime(); u.GetName() != "Renamed" || u.GetDescription() != "" ||
		len(u.GetLabels()) != 0 || u.GetCreatedBy() != "mobile-test" {
		t.Fatalf("updated runtime = %v", u)
	}

	// Another Workspace neither sees, reads, edits, mints for nor deletes it.
	peer := h.daemons(daemonPeerWorkspace)
	list, err := peer.ListDaemonRuntimes(ctx, connect.NewRequest(&agentsv1.ListDaemonRuntimesRequest{}))
	if err != nil || len(list.Msg.GetRuntimes()) != 0 {
		t.Fatalf("peer list = %v, %v", list, err)
	}
	_, err = peer.GetDaemonRuntime(ctx, connect.NewRequest(&agentsv1.GetDaemonRuntimeRequest{Id: "rt-1"}))
	wantCode(t, err, connect.CodeNotFound)
	_, err = peer.UpdateDaemonRuntime(ctx, connect.NewRequest(&agentsv1.UpdateDaemonRuntimeRequest{
		Runtime: &agentsv1.DaemonRuntime{Id: "rt-1", Name: "stolen"}}))
	wantCode(t, err, connect.CodeNotFound)
	_, err = peer.CreateDaemonRuntimeToken(ctx, connect.NewRequest(&agentsv1.CreateDaemonRuntimeTokenRequest{DaemonRuntimeId: "rt-1"}))
	wantCode(t, err, connect.CodeNotFound)
	_, err = peer.DeleteDaemonRuntime(ctx, connect.NewRequest(&agentsv1.DeleteDaemonRuntimeRequest{Id: "rt-1"}))
	wantCode(t, err, connect.CodeNotFound)

	// Without a Workspace header and in a Workspace the account is not in.
	_, err = h.daemons("").ListDaemonRuntimes(ctx, connect.NewRequest(&agentsv1.ListDaemonRuntimesRequest{}))
	wantCode(t, err, connect.CodeFailedPrecondition)
	_, err = h.daemons(mcpDeniedWorkspace).ListDaemonRuntimes(ctx, connect.NewRequest(&agentsv1.ListDaemonRuntimesRequest{}))
	wantCode(t, err, connect.CodePermissionDenied)

	if _, err := d.DeleteDaemonRuntime(ctx, connect.NewRequest(&agentsv1.DeleteDaemonRuntimeRequest{Id: "rt-1"})); err != nil {
		t.Fatal(err)
	}
	_, err = d.GetDaemonRuntime(ctx, connect.NewRequest(&agentsv1.GetDaemonRuntimeRequest{Id: "rt-1"}))
	wantCode(t, err, connect.CodeNotFound)
}

func TestWorkerTokenIsMintedOnceBoundToItsRuntimeAndNeverAnAPIToken(t *testing.T) {
	h := newDaemonHarness(t)
	ctx := context.Background()
	h.createRuntime(daemonWorkspace, "rt-1")
	h.createRuntime(daemonWorkspace, "rt-2")

	minted := h.mint(daemonWorkspace, "rt-1")
	tok := minted.GetToken()
	if minted.GetSecret() == "" || tok.GetKind() != agentsv1.APITokenKind_API_TOKEN_KIND_DAEMON ||
		tok.GetDaemonRuntimeId() != "rt-1" || tok.GetWorkspaceId() != daemonWorkspace ||
		len(tok.GetScopes()) != 1 || tok.GetScopes()[0] != "daemon:connect" {
		t.Fatalf("minted = %v", minted)
	}
	if strings.Contains(tok.String(), minted.GetSecret()) {
		t.Fatal("the token record carries its own secret; only the response's secret field may")
	}

	// Not a runtime of this Workspace: no token.
	_, err := h.daemons(daemonWorkspace).CreateDaemonRuntimeToken(ctx, connect.NewRequest(
		&agentsv1.CreateDaemonRuntimeTokenRequest{DaemonRuntimeId: "ghost"}))
	wantCode(t, err, connect.CodeNotFound)

	// The worker token never opens the HTTP API.
	req, _ := http.NewRequest("POST", h.url+"/api"+agentsv1connect.DaemonServiceListDaemonRuntimesProcedure, strings.NewReader("{}"))
	req.Header.Set("Authorization", "Bearer "+minted.GetSecret())
	req.Header.Set("X-Workspace-ID", daemonWorkspace)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("worker token on the HTTP API = %d, want 401", resp.StatusCode)
	}

	// A signed-in session and an ordinary API token are not worker tokens.
	if code := h.startWorker(daemonWorkspace, "rt-1", h.token, ""); code == http.StatusOK {
		t.Fatal("a user session registered as a worker")
	}
	userSecret := "user-api-token-secret"
	if err := h.fx.tokens.Create(ctx, &agentsv1.APIToken{
		Id: "user-token", Name: "ordinary", WorkspaceId: daemonWorkspace,
		Kind: agentsv1.APITokenKind_API_TOKEN_KIND_USER, Scopes: []string{"api:*"},
	}, application.HashAPITokenSecret(userSecret)); err != nil {
		t.Fatal(err)
	}
	if code := h.startWorker(daemonWorkspace, "rt-1", userSecret, ""); code == http.StatusOK {
		t.Fatal("an ordinary API token registered as a worker")
	}
	if got := len(h.fx.registry.ListConnections(daemonWorkspace)); got != 0 {
		t.Fatalf("%d workers registered with the wrong credentials", got)
	}

	// The runtime's own token brings it online, and only as that runtime.
	if code := h.startWorker(daemonWorkspace, "rt-1", minted.GetSecret(), ""); code != http.StatusOK {
		t.Fatalf("worker with its own token: HTTP %d", code)
	}
	t.Cleanup(func() {
		h.control("POST", "/worker/drop", map[string]string{"workspace": daemonWorkspace, "runtime_id": "rt-1"}, nil)
	})
	list, err := h.daemons(daemonWorkspace).ListDaemons(ctx, connect.NewRequest(&agentsv1.ListDaemonsRequest{}))
	if err != nil || len(list.Msg.GetDaemons()) != 1 {
		t.Fatalf("daemons = %v, %v", list, err)
	}
	online := list.Msg.GetDaemons()[0]
	if online.GetDaemonRuntimeId() != "rt-1" || online.GetState() != agentsv1.DaemonStatus_STATE_IDLE ||
		online.GetVersion() != "v0.0.0-contract" || online.GetOs() != "linux-amd64" ||
		len(online.GetAcpRuntimes()) != 2 || online.GetLabels()["role"] != "contract" {
		t.Fatalf("online daemon = %v", online)
	}
}

func TestDispatchedTaskShowsProgressAndCancelIsOnlyDelivery(t *testing.T) {
	h := newDaemonHarness(t)
	ctx := context.Background()
	h.onlineWorker(daemonWorkspace, "rt-1", "stubborn")
	h.createDaemonAgent(daemonWorkspace, "ra-1", "rt-1")

	runID := h.dispatch(daemonWorkspace, "ra-1", "build it")
	task := h.waitForTask(daemonWorkspace)
	if task.GetDaemonRuntimeId() != "rt-1" || task.GetAcpRuntime() != "codex" ||
		task.GetWorkspaceId() != daemonWorkspace || task.GetAgentName() != "ra-1" ||
		task.GetAgentId() != "host-daemon-ra-1" || task.GetStartedAt() == nil {
		t.Fatalf("task = %v", task)
	}
	eventually(t, "the worker received the task", func() bool { return len(h.workerState(daemonWorkspace, "rt-1").Tasks) == 1 })
	received := h.workerState(daemonWorkspace, "rt-1").Tasks[0]
	if received.Input != "build it" || received.WorkspaceID != daemonWorkspace ||
		received.RuntimeID != "rt-1" || received.TaskID != task.GetTaskId() {
		t.Fatalf("worker received %+v", received)
	}

	// Progress the worker reports is what the list shows.
	h.control("POST", "/task/progress", map[string]any{
		"workspace": daemonWorkspace, "runtime_id": "rt-1", "task_id": task.GetTaskId(),
		"step": "Compiling", "progress": 40,
	}, nil)
	eventually(t, "progress in the task list", func() bool {
		list := h.tasks(daemonWorkspace)
		return len(list) == 1 && list[0].GetCurrentStep() == "Compiling" && list[0].GetProgress() == 40
	})
	if got := h.tasks(daemonWorkspace)[0]; got.GetElapsed().AsDuration() <= 0 {
		t.Fatalf("elapsed = %v", got.GetElapsed())
	}
	statuses, err := h.daemons(daemonWorkspace).ListDaemons(ctx, connect.NewRequest(&agentsv1.ListDaemonsRequest{}))
	if err != nil || statuses.Msg.GetDaemons()[0].GetActiveTasks() != 1 ||
		statuses.Msg.GetDaemons()[0].GetState() != agentsv1.DaemonStatus_STATE_ONLINE {
		t.Fatalf("daemons = %v, %v", statuses, err)
	}

	// Cancel: the server accepts and delivers it, naming the runtime that got
	// it. A stubborn worker keeps going, so the task is still in flight and
	// the run has not ended: success is delivery, not termination.
	resp, err := h.daemons(daemonWorkspace).CancelDaemonTask(ctx, connect.NewRequest(
		&agentsv1.CancelDaemonTaskRequest{TaskId: task.GetTaskId()}))
	if err != nil || resp.Msg.GetDaemonRuntimeId() != "rt-1" {
		t.Fatalf("cancel = %v, %v", resp, err)
	}
	eventually(t, "the worker saw the cancel", func() bool {
		c := h.workerState(daemonWorkspace, "rt-1").Cancels
		return len(c) == 1 && c[0] == task.GetTaskId()
	})
	time.Sleep(300 * time.Millisecond)
	if len(h.tasks(daemonWorkspace)) != 1 {
		t.Fatal("the task left the list although the worker never reported it cancelled")
	}
	if h.runOutcome(runID).Done {
		t.Fatal("the run ended on a delivered cancel")
	}

	// The worker's own terminal report ends the task and the run.
	h.control("POST", "/task/fail", map[string]any{
		"workspace": daemonWorkspace, "runtime_id": "rt-1", "task_id": task.GetTaskId(), "error": "stopped by worker",
	}, nil)
	eventually(t, "the task to leave the list", func() bool { return len(h.tasks(daemonWorkspace)) == 0 })
	eventually(t, "the run to end", func() bool { return h.runOutcome(runID).Done })
	if out := h.runOutcome(runID); !strings.Contains(out.Error, "stopped by worker") {
		t.Fatalf("run outcome = %+v", out)
	}
}

func TestCooperativeWorkerReportsCancelledAndSuccessCarriesTheResult(t *testing.T) {
	h := newDaemonHarness(t)
	ctx := context.Background()
	h.onlineWorker(daemonWorkspace, "rt-1", "cooperative")
	h.createDaemonAgent(daemonWorkspace, "ra-1", "rt-1")

	runID := h.dispatch(daemonWorkspace, "ra-1", "first")
	task := h.waitForTask(daemonWorkspace)
	if _, err := h.daemons(daemonWorkspace).CancelDaemonTask(ctx, connect.NewRequest(
		&agentsv1.CancelDaemonTaskRequest{TaskId: task.GetTaskId(), DaemonRuntimeId: "rt-1"})); err != nil {
		t.Fatal(err)
	}
	eventually(t, "the run to end", func() bool { return h.runOutcome(runID).Done })
	if out := h.runOutcome(runID); !strings.Contains(out.Error, "cancelled") {
		t.Fatalf("run outcome = %+v", out)
	}
	if len(h.tasks(daemonWorkspace)) != 0 {
		t.Fatal("a cancelled task is still listed")
	}

	second := h.dispatch(daemonWorkspace, "ra-1", "second")
	task = h.waitForTask(daemonWorkspace)
	h.control("POST", "/task/complete", map[string]any{
		"workspace": daemonWorkspace, "runtime_id": "rt-1", "task_id": task.GetTaskId(), "output": "all done",
	}, nil)
	eventually(t, "the run to end", func() bool { return h.runOutcome(second).Done })
	if out := h.runOutcome(second); out.Output != "all done" || out.Error != "" {
		t.Fatalf("run outcome = %+v", out)
	}
}

func TestCancelAndListNeverCrossTenantsAndFailWhenTheWorkerIsGone(t *testing.T) {
	h := newDaemonHarness(t)
	ctx := context.Background()
	// The same runtime ID in two Workspaces is two independent runtimes.
	h.onlineWorker(daemonWorkspace, "shared-rt", "stubborn")
	h.onlineWorker(daemonPeerWorkspace, "shared-rt", "stubborn")
	// Remote Agent IDs resolve in one flat registry at run time, so the two
	// Workspaces use distinct IDs; the App never chooses an ID.
	h.createDaemonAgent(daemonWorkspace, "ra-1", "shared-rt")
	h.createDaemonAgent(daemonPeerWorkspace, "peer-ra-1", "shared-rt")

	h.dispatch(daemonWorkspace, "ra-1", "mine")
	mine := h.waitForTask(daemonWorkspace)
	if got := h.tasks(daemonPeerWorkspace); len(got) != 0 {
		t.Fatalf("peer Workspace sees %d foreign tasks", len(got))
	}
	if peerWorker := h.workerState(daemonPeerWorkspace, "shared-rt"); len(peerWorker.Tasks) != 0 {
		t.Fatalf("peer worker received a foreign task: %+v", peerWorker.Tasks)
	}

	// Cancelling someone else's task, with or without a runtime hint, finds
	// nothing and delivers nothing.
	for _, hint := range []string{"", "shared-rt"} {
		_, err := h.daemons(daemonPeerWorkspace).CancelDaemonTask(ctx, connect.NewRequest(
			&agentsv1.CancelDaemonTaskRequest{TaskId: mine.GetTaskId(), DaemonRuntimeId: hint}))
		wantCode(t, err, connect.CodeNotFound)
	}
	time.Sleep(200 * time.Millisecond)
	if c := h.workerState(daemonWorkspace, "shared-rt").Cancels; len(c) != 0 {
		t.Fatalf("a cross-tenant cancel reached the worker: %v", c)
	}
	// A peer's daemon directory is its own.
	if _, err := h.daemons(daemonPeerWorkspace).GetDaemon(ctx, connect.NewRequest(&agentsv1.GetDaemonRequest{DaemonRuntimeId: "shared-rt"})); err != nil {
		t.Fatal(err)
	}
	_, err := h.daemons("alpha").GetDaemon(ctx, connect.NewRequest(&agentsv1.GetDaemonRequest{DaemonRuntimeId: "shared-rt"}))
	wantCode(t, err, connect.CodeNotFound)

	// The worker disappears without a word: the task is gone with it, and a
	// cancel for it is not found rather than silently "delivered".
	h.control("POST", "/worker/stop", map[string]string{"workspace": daemonWorkspace, "runtime_id": "shared-rt"}, nil)
	eventually(t, "the worker to leave the directory", func() bool {
		resp, err := h.daemons(daemonWorkspace).ListDaemons(ctx, connect.NewRequest(&agentsv1.ListDaemonsRequest{}))
		return err == nil && len(resp.Msg.GetDaemons()) == 0
	})
	_, err = h.daemons(daemonWorkspace).CancelDaemonTask(ctx, connect.NewRequest(
		&agentsv1.CancelDaemonTaskRequest{TaskId: mine.GetTaskId()}))
	wantCode(t, err, connect.CodeNotFound)
	if got := h.tasks(daemonWorkspace); len(got) != 0 {
		t.Fatalf("tasks of an offline worker = %v", got)
	}
}

func TestRemoteAgentStatusFollowsTheWorkerAndRuntimeDeleteIsUnguarded(t *testing.T) {
	h := newDaemonHarness(t)
	ctx := context.Background()
	h.createRuntime(daemonWorkspace, "rt-1")
	h.createDaemonAgent(daemonWorkspace, "ra-1", "rt-1")
	status := func() *agentsv1.RemoteAgentStatus {
		resp, err := h.remotes(daemonWorkspace).GetRemoteAgentStatus(ctx, connect.NewRequest(
			&agentsv1.GetRemoteAgentStatusRequest{Id: "ra-1"}))
		if err != nil {
			t.Fatal(err)
		}
		return resp.Msg.GetStatus()
	}

	if s := status(); s.GetState() != agentsv1.RemoteAgentStatus_STATE_UNREACHABLE {
		t.Fatalf("offline runtime status = %v", s)
	}
	if code := h.startWorker(daemonWorkspace, "rt-1", h.mint(daemonWorkspace, "rt-1").GetSecret(), "stubborn"); code != http.StatusOK {
		t.Fatalf("worker: HTTP %d", code)
	}
	t.Cleanup(func() {
		h.control("POST", "/worker/drop", map[string]string{"workspace": daemonWorkspace, "runtime_id": "rt-1"}, nil)
	})
	if s := status(); s.GetState() != agentsv1.RemoteAgentStatus_STATE_IDLE || s.GetServingDaemonRuntimeId() != "rt-1" {
		t.Fatalf("online runtime status = %v", s)
	}
	h.dispatch(daemonWorkspace, "ra-1", "busy")
	h.waitForTask(daemonWorkspace)
	if s := status(); s.GetState() != agentsv1.RemoteAgentStatus_STATE_ACTIVE {
		t.Fatalf("busy runtime status = %v", s)
	}

	// A Remote Agent can only bind a runtime of its own Workspace.
	_, err := h.remotes(daemonPeerWorkspace).CreateRemoteAgent(ctx, connect.NewRequest(
		&agentsv1.CreateRemoteAgentRequest{RemoteAgent: &agentsv1.RemoteAgent{
			Id: "ra-x", Name: "ra-x", Protocol: agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_DAEMON,
			DaemonRuntimeId: "rt-1", AcpRuntime: "codex",
		}}))
	wantCode(t, err, connect.CodeNotFound)
	_, err = h.remotes(daemonWorkspace).CreateRemoteAgent(ctx, connect.NewRequest(
		&agentsv1.CreateRemoteAgentRequest{RemoteAgent: &agentsv1.RemoteAgent{
			Id: "ra-y", Name: "ra-y", Protocol: agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_DAEMON,
			DaemonRuntimeId: "rt-1", AcpRuntime: "claude",
		}}))
	wantCode(t, err, connect.CodeInvalidArgument)

	// Characterization of a server gap the App must disclose: deleting a
	// runtime a Remote Agent references is allowed, the Remote Agent is not
	// told, and its status cannot tell a deleted runtime from an offline one.
	if _, err := h.daemons(daemonWorkspace).DeleteDaemonRuntime(ctx, connect.NewRequest(&agentsv1.DeleteDaemonRuntimeRequest{Id: "rt-1"})); err != nil {
		t.Fatalf("delete of a referenced runtime was refused: %v", err)
	}
	if _, err := h.remotes(daemonWorkspace).GetRemoteAgent(ctx, connect.NewRequest(&agentsv1.GetRemoteAgentRequest{Id: "ra-1"})); err != nil {
		t.Fatal(err)
	}
	// The deleted runtime's token no longer authenticates a worker.
	eventually(t, "the deleted runtime's worker to be refused", func() bool {
		return h.workerState(daemonWorkspace, "rt-1").Stopped != ""
	})
}

// Two Workspaces may each own a Remote Agent with the same ID. A run in one
// must reach only that Workspace's runtime, never a same-ID agent of another.
func TestSameRemoteAgentIDInTwoWorkspacesDispatchesToItsOwnWorker(t *testing.T) {
	h := newDaemonHarness(t)
	h.onlineWorker(daemonWorkspace, "rt-a", "stubborn")
	h.onlineWorker(daemonPeerWorkspace, "rt-b", "stubborn")
	h.createDaemonAgent(daemonWorkspace, "build-box", "rt-a")
	h.createDaemonAgent(daemonPeerWorkspace, "build-box", "rt-b")

	h.dispatch(daemonWorkspace, "build-box", "mine")
	h.waitForTask(daemonWorkspace)
	eventually(t, "the worker received the task", func() bool { return len(h.workerState(daemonWorkspace, "rt-a").Tasks) == 1 })
	time.Sleep(200 * time.Millisecond)
	if got := h.workerState(daemonPeerWorkspace, "rt-b").Tasks; len(got) != 0 {
		t.Fatalf("the peer Workspace's worker received a foreign task: %+v", got)
	}
	if got := h.tasks(daemonPeerWorkspace); len(got) != 0 {
		t.Fatalf("the peer Workspace lists %d foreign tasks", len(got))
	}

	h.dispatch(daemonPeerWorkspace, "build-box", "theirs")
	eventually(t, "the peer worker received its task", func() bool { return len(h.workerState(daemonPeerWorkspace, "rt-b").Tasks) == 1 })
	if got := h.workerState(daemonWorkspace, "rt-a").Tasks; len(got) != 1 || got[0].Input != "mine" {
		t.Fatalf("worker rt-a tasks = %+v", got)
	}
}

func TestWorkerTokensAreListedWithoutSecretsAndRevocationNeedsAnOwnerOrAdmin(t *testing.T) {
	h := newDaemonHarness(t)
	ctx := context.Background()
	h.createRuntime(daemonWorkspace, "rt-1")
	h.createRuntime(daemonPeerWorkspace, "rt-p")
	mine := h.mint(daemonWorkspace, "rt-1")
	theirs := h.mint(daemonPeerWorkspace, "rt-p")

	tokens := agentsv1connect.NewAPITokenServiceClient(http.DefaultClient, h.url+"/api", h.opts(daemonWorkspace))
	list, err := tokens.ListAPITokens(ctx, connect.NewRequest(&agentsv1.ListAPITokensRequest{}))
	if err != nil || len(list.Msg.GetTokens()) != 1 {
		t.Fatalf("tokens = %v, %v", list, err)
	}
	got := list.Msg.GetTokens()[0]
	if got.GetId() != mine.GetToken().GetId() || got.GetKind() != agentsv1.APITokenKind_API_TOKEN_KIND_DAEMON ||
		got.GetDaemonRuntimeId() != "rt-1" || got.GetPrefix() == "" || got.GetRevoked() {
		t.Fatalf("listed token = %v", got)
	}
	if strings.Contains(got.String(), mine.GetSecret()) {
		t.Fatal("the listing carries the token's secret")
	}

	// A member of another Workspace cannot revoke; the owner cannot reach a
	// token of a Workspace it does not name.
	member := agentsv1connect.NewAPITokenServiceClient(http.DefaultClient, h.url+"/api", h.opts(daemonPeerWorkspace))
	_, err = member.RevokeAPIToken(ctx, connect.NewRequest(&agentsv1.RevokeAPITokenRequest{Id: theirs.GetToken().GetId()}))
	wantCode(t, err, connect.CodePermissionDenied)
	_, err = tokens.RevokeAPIToken(ctx, connect.NewRequest(&agentsv1.RevokeAPITokenRequest{Id: theirs.GetToken().GetId()}))
	wantCode(t, err, connect.CodeNotFound)

	if _, err := tokens.RevokeAPIToken(ctx, connect.NewRequest(&agentsv1.RevokeAPITokenRequest{Id: mine.GetToken().GetId()})); err != nil {
		t.Fatal(err)
	}
	if code := h.startWorker(daemonWorkspace, "rt-1", mine.GetSecret(), ""); code == http.StatusOK {
		t.Fatal("a revoked worker token registered")
	}
	list, err = tokens.ListAPITokens(ctx, connect.NewRequest(&agentsv1.ListAPITokensRequest{}))
	if err != nil || !list.Msg.GetTokens()[0].GetRevoked() {
		t.Fatalf("tokens after revoke = %v, %v", list, err)
	}
}

// The App asks for a worker token's lifetime in hours, so one hour, twelve
// hours and half an hour must each reach the production service as exactly
// that, and a request without a ttl must stay a token that never expires.
func TestWorkerTokenLifetimeIsHonoredBelowOneDay(t *testing.T) {
	h := newDaemonHarness(t)
	ctx := context.Background()
	h.createRuntime(daemonWorkspace, "rt-ttl")

	mint := func(ttl *durationpb.Duration) *agentsv1.APIToken {
		t.Helper()
		resp, err := h.daemons(daemonWorkspace).CreateDaemonRuntimeToken(ctx, connect.NewRequest(
			&agentsv1.CreateDaemonRuntimeTokenRequest{DaemonRuntimeId: "rt-ttl", Name: "ttl", Ttl: ttl}))
		if err != nil {
			t.Fatal(err)
		}
		return resp.Msg.GetToken()
	}
	for _, tc := range []struct {
		name string
		ttl  time.Duration
	}{
		{"one hour", time.Hour},
		{"twelve hours", 12 * time.Hour},
		{"half an hour", 30 * time.Minute},
		{"a tenth of a second", 100 * time.Millisecond},
	} {
		before := time.Now()
		tok := mint(durationpb.New(tc.ttl))
		if tok.GetExpiresAt() == nil {
			t.Fatalf("%s: the token does not expire", tc.name)
		}
		left := tok.GetExpiresAt().AsTime().Sub(before)
		if left < tc.ttl-time.Second || left > tc.ttl+time.Second {
			t.Fatalf("%s: expires in %s, want about %s", tc.name, left, tc.ttl)
		}
	}
	if tok := mint(nil); tok.GetExpiresAt() != nil {
		t.Fatalf("no ttl minted a token that expires at %v", tok.GetExpiresAt())
	}
}
