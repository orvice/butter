package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/gin-gonic/gin"
	adkrunner "google.golang.org/adk/v2/runner"
	adksession "google.golang.org/adk/v2/session"
	"google.golang.org/genai"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"

	"go.orx.me/apps/butter/internal/application"
	apitokenmemory "go.orx.me/apps/butter/internal/repo/apitoken/memory"
	configmemory "go.orx.me/apps/butter/internal/repo/config/memory"
	"go.orx.me/apps/butter/internal/repo/workspace"
	"go.orx.me/apps/butter/internal/runtime/daemon"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

// The Daemon fixture (#455): the production DaemonService (the control plane
// the App drives), the production poll transport of DaemonConnectorService
// (what cmd/butter-daemon speaks), the daemon registry, the API token
// repository and the production runner with its daemon Bridge, over in-memory
// storage. Only the worker is a double: a deterministic fake worker that
// speaks the real Register / Poll / ReportTaskUpdate / Unregister protocol
// with a worker token, so a dispatched task travels the same path as a real
// one. No worker is started on the phone and none of this is a real
// opencode or codex run.
//
// The fake worker and the dispatcher are driven over a small loopback control
// API under /fake-daemon/ (never reachable through the production routes):
//
//	POST worker/start    register a worker with a token the App minted
//	POST worker/stop     unregister it gracefully
//	POST worker/drop     lose it without Unregister (a crashed worker)
//	GET  worker/state    what it received, which cancels it saw
//	POST dispatch        run a saved Remote Agent through the production runner
//	GET  run             the run's outcome
//	POST task/progress   report a step and percentage
//	POST task/complete   report a result
//	POST task/fail       report a failure
//
// Worker modes steer cancel handling: "cooperative" reports CANCELLED once a
// cancel is delivered, "stubborn" records the cancel and keeps working.
const (
	daemonWorkspace     = "daemon"
	daemonPeerWorkspace = "daemon-peer"
)

type daemonFixture struct {
	ctx      context.Context
	agents   *configmemory.Store
	tokens   *apitokenmemory.Store
	registry *daemon.Registry
	svc      *application.DaemonServiceServer
	handler  *daemon.GRPCHandler
	runner   *runner.Service

	mu      sync.Mutex // guards workers, runs, nextRun and runner reloads
	workers map[string]*fakeWorker
	runs    map[string]*dispatchRun
	nextRun int
}

func newDaemonFixture(ctx context.Context, agents *configmemory.Store) (*daemonFixture, error) {
	f := &daemonFixture{
		ctx: ctx, agents: agents, tokens: apitokenmemory.New(),
		registry: daemon.NewRegistry(),
		workers:  map[string]*fakeWorker{}, runs: map[string]*dispatchRun{},
	}
	f.svc = application.NewDaemonServiceServer(agents, f.registry)
	f.svc.SetAPITokenRepo(f.tokens)
	f.handler = daemon.NewGRPCHandler(f.registry, f.tokens, agents)
	svc, err := runner.NewService(ctx, nil, nil, nil, nil, f.registry,
		adksession.InMemoryService(), nil, nil, adkrunner.PluginConfig{})
	if err != nil {
		return nil, err
	}
	f.runner = svc
	return f, nil
}

// register mounts the production procedures the App uses and the poll
// transport the worker uses, plus the control API.
func (f *daemonFixture) register(r *gin.Engine, mux *http.ServeMux, opts []connect.HandlerOption, workspaces workspace.Repository) {
	_, svcHandler := agentsv1connect.NewDaemonServiceHandler(f.svc, opts...)
	for _, procedure := range []string{
		agentsv1connect.DaemonServiceListDaemonRuntimesProcedure,
		agentsv1connect.DaemonServiceGetDaemonRuntimeProcedure,
		agentsv1connect.DaemonServiceCreateDaemonRuntimeProcedure,
		agentsv1connect.DaemonServiceUpdateDaemonRuntimeProcedure,
		agentsv1connect.DaemonServiceDeleteDaemonRuntimeProcedure,
		agentsv1connect.DaemonServiceCreateDaemonRuntimeTokenProcedure,
		agentsv1connect.DaemonServiceListDaemonsProcedure,
		agentsv1connect.DaemonServiceGetDaemonProcedure,
		agentsv1connect.DaemonServiceCancelDaemonTaskProcedure,
		agentsv1connect.DaemonServiceListDaemonTasksProcedure,
		agentsv1connect.DaemonServiceGetBridgeDiagnosticsProcedure,
	} {
		r.POST("/api"+procedure, gin.WrapH(http.StripPrefix("/api", svcHandler)))
	}
	// Token inventory: only listing and revoking. Creating an ordinary API
	// token is not offered, so a worker never gets one.
	tokenSvc := application.NewAPITokenServiceServer(f.tokens)
	tokenSvc.SetWorkspaceRepo(workspaces)
	_, tokenHandler := agentsv1connect.NewAPITokenServiceHandler(tokenSvc, opts...)
	for _, procedure := range []string{
		agentsv1connect.APITokenServiceListAPITokensProcedure,
		agentsv1connect.APITokenServiceRevokeAPITokenProcedure,
	} {
		r.POST("/api"+procedure, gin.WrapH(http.StripPrefix("/api", tokenHandler)))
	}
	// The worker's unary poll transport. Connect (the bidirectional stream)
	// is not mounted: the fake worker, like a worker behind a proxy, polls.
	_, connectorHandler := agentsv1connect.NewDaemonConnectorServiceHandler(f.handler, opts...)
	for _, procedure := range []string{
		agentsv1connect.DaemonConnectorServiceRegisterProcedure,
		agentsv1connect.DaemonConnectorServicePollProcedure,
		agentsv1connect.DaemonConnectorServiceReportTaskUpdateProcedure,
		agentsv1connect.DaemonConnectorServiceUnregisterProcedure,
	} {
		r.POST("/api"+procedure, gin.WrapH(http.StripPrefix("/api", connectorHandler)))
	}
	mux.Handle("/fake-daemon/", http.StripPrefix("/fake-daemon", f.control()))
}

func workerKey(workspaceID, runtimeID string) string { return workspaceID + "/" + runtimeID }

// ---- fake worker -------------------------------------------------------

type workerTask struct {
	TaskID      string `json:"task_id"`
	AgentName   string `json:"agent_name"`
	AgentID     string `json:"agent_id"`
	Input       string `json:"input"`
	WorkspaceID string `json:"workspace_id"`
	RuntimeID   string `json:"daemon_runtime_id"`
	AcpRuntime  string `json:"acp_runtime"`
	SessionID   string `json:"session_id"`
}

type fakeWorker struct {
	client    agentsv1connect.DaemonConnectorServiceClient
	token     string
	mode      string
	workspace string
	runtime   string
	cancel    context.CancelFunc
	done      chan struct{}

	mu      sync.Mutex
	tasks   []workerTask
	cancels []string
	stopped string // why the poll loop ended, "" while it runs
}

func authorized[T any](w *fakeWorker, msg *T) *connect.Request[T] {
	req := connect.NewRequest(msg)
	req.Header().Set("Authorization", "Bearer "+w.token)
	return req
}

func (w *fakeWorker) report(ctx context.Context, update *agentsv1.DaemonTaskUpdate) error {
	_, err := w.client.ReportTaskUpdate(ctx, authorized(w,
		&agentsv1.DaemonConnectorServiceReportTaskUpdateRequest{Update: update}))
	return err
}

func (w *fakeWorker) loop(ctx context.Context) {
	defer close(w.done)
	for ctx.Err() == nil {
		resp, err := w.client.Poll(ctx, authorized(w, &agentsv1.DaemonConnectorServicePollRequest{
			WaitTimeout: durationpb.New(200 * time.Millisecond),
		}))
		if err != nil {
			if ctx.Err() == nil {
				w.mu.Lock()
				w.stopped = err.Error()
				w.mu.Unlock()
			}
			return
		}
		for _, msg := range resp.Msg.GetMessages() {
			if task := msg.GetTask(); task != nil {
				w.mu.Lock()
				w.tasks = append(w.tasks, workerTask{
					TaskID: task.GetTaskId(), AgentName: task.GetAgentName(), AgentID: task.GetAgentId(),
					Input: task.GetInput(), WorkspaceID: task.GetWorkspaceId(),
					RuntimeID: task.GetDaemonRuntimeId(), AcpRuntime: task.GetAcpRuntime(),
					SessionID: task.GetSessionId(),
				})
				w.mu.Unlock()
				_ = w.report(ctx, &agentsv1.DaemonTaskUpdate{
					TaskId: task.GetTaskId(), Status: agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_RUNNING,
					CurrentStep: "Starting",
				})
			}
			if cancel := msg.GetCancel(); cancel != nil {
				w.mu.Lock()
				w.cancels = append(w.cancels, cancel.GetTaskId())
				w.mu.Unlock()
				if w.mode != "stubborn" {
					_ = w.report(ctx, &agentsv1.DaemonTaskUpdate{
						TaskId: cancel.GetTaskId(), Status: agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_CANCELLED,
					})
				}
			}
		}
	}
}

func (f *daemonFixture) startWorker(base, workspaceID, runtimeID, token, mode string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	ctx, cancel := context.WithCancel(f.ctx)
	w := &fakeWorker{
		client: agentsv1connect.NewDaemonConnectorServiceClient(http.DefaultClient, base+"/api"),
		token:  token, mode: mode, workspace: workspaceID, runtime: runtimeID,
		cancel: cancel, done: make(chan struct{}),
	}
	_, err := w.client.Register(ctx, authorized(w, &agentsv1.DaemonConnectorServiceRegisterRequest{
		Daemon: &agentsv1.DaemonInfo{
			Name: "fake-worker", AcpRuntimes: []string{"opencode", "codex"},
			Version: "v0.0.0-contract", Os: "linux-amd64", Executors: []string{"fake"},
			Labels: map[string]string{"role": "contract"},
		},
	}))
	if err != nil {
		cancel()
		return err
	}
	f.workers[workerKey(workspaceID, runtimeID)] = w
	go w.loop(ctx)
	return nil
}

func (f *daemonFixture) worker(workspaceID, runtimeID string) *fakeWorker {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.workers[workerKey(workspaceID, runtimeID)]
}

// ---- dispatch ----------------------------------------------------------

type dispatchRun struct {
	done   chan struct{}
	mu     sync.Mutex
	output string
	err    string
}

// dispatch runs a saved Remote Agent through the production runner and its
// daemon Bridge: a SEQUENTIAL host Agent whose remote_agent_ids names it.
func (f *daemonFixture) dispatch(workspaceID, remoteAgentID, input string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var hosts []agentsv1.Agent
	var remotes []agentsv1.RemoteAgent
	for _, ws := range []string{daemonWorkspace, daemonPeerWorkspace} {
		list, err := f.agents.ListRemoteAgents(f.ctx, ws)
		if err != nil {
			return "", err
		}
		for _, stored := range list {
			remotes = append(remotes, agentsv1.RemoteAgent{})
			proto.Merge(&remotes[len(remotes)-1], stored)
			remotes[len(remotes)-1].WorkspaceId = ws
			hosts = append(hosts, agentsv1.Agent{
				AgentId: "host-" + ws + "-" + stored.GetId(), Name: "host-" + ws + "-" + stored.GetId(),
				WorkspaceId: ws, Type: agentsv1.AgentType_AGENT_TYPE_SEQUENTIAL,
				Config: &agentsv1.AgentConfig{RemoteAgentIds: []string{stored.GetId()}},
			})
		}
	}
	if err := f.runner.ReloadProtoAgents(f.ctx, hosts, nil, nil, remotes); err != nil {
		return "", err
	}
	f.nextRun++
	id := fmt.Sprintf("run-%d", f.nextRun)
	run := &dispatchRun{done: make(chan struct{})}
	f.runs[id] = run
	host := "host-" + workspaceID + "-" + remoteAgentID
	go func() {
		defer close(run.done)
		turn, err := f.runner.RunTurn(f.ctx, host, []*genai.Part{{Text: input}}, "", &agentsv1.ContextInfo{
			Uuid: "uuid-" + id, SessionId: "session-" + id, UserId: "mobile-test",
			ChannelName: "contract", Source: agentsv1.ContextSource_CONTEXT_SOURCE_API,
			WorkspaceId: workspaceID,
		}, nil, nil)
		run.mu.Lock()
		defer run.mu.Unlock()
		if turn != nil {
			run.output = turn.Output
		}
		if err != nil {
			run.err = err.Error()
		}
	}()
	return id, nil
}

// ---- control API -------------------------------------------------------

func (f *daemonFixture) control() http.Handler {
	mux := http.NewServeMux()
	handle := func(path string, fn func(w http.ResponseWriter, r *http.Request) (any, int, error)) {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			out, code, err := fn(w, r)
			w.Header().Set("Content-Type", "application/json")
			if err != nil {
				w.WriteHeader(code)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			_ = json.NewEncoder(w).Encode(out)
		})
	}
	body := func(r *http.Request, into any) error { return json.NewDecoder(r.Body).Decode(into) }

	handle("/worker/start", func(_ http.ResponseWriter, r *http.Request) (any, int, error) {
		var in struct {
			Workspace string `json:"workspace"`
			Runtime   string `json:"runtime_id"`
			Token     string `json:"token"`
			Mode      string `json:"mode"`
		}
		if err := body(r, &in); err != nil {
			return nil, http.StatusBadRequest, err
		}
		if err := f.startWorker("http://"+r.Host, in.Workspace, in.Runtime, in.Token, in.Mode); err != nil {
			var ce *connect.Error
			if errors.As(err, &ce) {
				return nil, http.StatusUnauthorized, fmt.Errorf("%s: %s", ce.Code(), ce.Message())
			}
			return nil, http.StatusBadGateway, err
		}
		return map[string]bool{"registered": true}, 0, nil
	})
	stop := func(graceful bool) func(http.ResponseWriter, *http.Request) (any, int, error) {
		return func(_ http.ResponseWriter, r *http.Request) (any, int, error) {
			var in struct {
				Workspace string `json:"workspace"`
				Runtime   string `json:"runtime_id"`
			}
			if err := body(r, &in); err != nil {
				return nil, http.StatusBadRequest, err
			}
			w := f.worker(in.Workspace, in.Runtime)
			if w == nil {
				return nil, http.StatusNotFound, errors.New("no such fake worker")
			}
			if graceful {
				_, _ = w.client.Unregister(r.Context(), authorized(w, &agentsv1.DaemonConnectorServiceUnregisterRequest{}))
			}
			w.cancel()
			<-w.done
			return map[string]bool{"stopped": true}, 0, nil
		}
	}
	handle("/worker/stop", stop(true))
	handle("/worker/drop", stop(false))
	handle("/worker/state", func(_ http.ResponseWriter, r *http.Request) (any, int, error) {
		w := f.worker(r.URL.Query().Get("workspace"), r.URL.Query().Get("runtime_id"))
		if w == nil {
			return nil, http.StatusNotFound, errors.New("no such fake worker")
		}
		w.mu.Lock()
		defer w.mu.Unlock()
		return map[string]any{
			"tasks": append([]workerTask{}, w.tasks...), "cancels": append([]string{}, w.cancels...),
			"stopped": w.stopped,
		}, 0, nil
	})
	handle("/dispatch", func(_ http.ResponseWriter, r *http.Request) (any, int, error) {
		var in struct {
			Workspace   string `json:"workspace"`
			RemoteAgent string `json:"remote_agent_id"`
			Input       string `json:"input"`
		}
		if err := body(r, &in); err != nil {
			return nil, http.StatusBadRequest, err
		}
		id, err := f.dispatch(in.Workspace, in.RemoteAgent, in.Input)
		if err != nil {
			return nil, http.StatusBadGateway, err
		}
		return map[string]string{"run_id": id}, 0, nil
	})
	handle("/run", func(_ http.ResponseWriter, r *http.Request) (any, int, error) {
		f.mu.Lock()
		run := f.runs[r.URL.Query().Get("id")]
		f.mu.Unlock()
		if run == nil {
			return nil, http.StatusNotFound, errors.New("no such run")
		}
		select {
		case <-run.done:
		default:
			return map[string]any{"done": false}, 0, nil
		}
		run.mu.Lock()
		defer run.mu.Unlock()
		return map[string]any{"done": true, "output": run.output, "error": run.err}, 0, nil
	})
	report := func(status agentsv1.DaemonTaskStatus) func(http.ResponseWriter, *http.Request) (any, int, error) {
		return func(_ http.ResponseWriter, r *http.Request) (any, int, error) {
			var in struct {
				Workspace string `json:"workspace"`
				Runtime   string `json:"runtime_id"`
				TaskID    string `json:"task_id"`
				Step      string `json:"step"`
				Progress  int32  `json:"progress"`
				Output    string `json:"output"`
				Error     string `json:"error"`
			}
			if err := body(r, &in); err != nil {
				return nil, http.StatusBadRequest, err
			}
			w := f.worker(in.Workspace, in.Runtime)
			if w == nil {
				return nil, http.StatusNotFound, errors.New("no such fake worker")
			}
			if err := w.report(r.Context(), &agentsv1.DaemonTaskUpdate{
				TaskId: in.TaskID, Status: status, CurrentStep: in.Step, Progress: in.Progress,
				Output: in.Output, Error: in.Error,
			}); err != nil {
				return nil, http.StatusBadGateway, err
			}
			return map[string]bool{"reported": true}, 0, nil
		}
	}
	handle("/task/progress", report(agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_RUNNING))
	handle("/task/complete", report(agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_COMPLETED))
	handle("/task/fail", report(agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_FAILED))
	return mux
}

func init() {
	registerFixtureWorkspaces(
		fixtureWorkspace{ID: daemonWorkspace, Owner: true, NoSeedAgent: true},
		fixtureWorkspace{ID: daemonPeerWorkspace, NoSeedAgent: true},
	)
}
