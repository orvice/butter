package executor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/acp-go-sdk"
	"google.golang.org/protobuf/proto"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestACPExecutorSuccess(t *testing.T) {
	exec := testACPExecutor(t, "success")
	task := &agentsv1.DaemonTask{TaskId: "task-1", AcpRuntime: "test-acp", Input: "hello"}
	updates, onUpdate := collectUpdates()

	if err := exec.Execute(context.Background(), task, onUpdate); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	got := updates()
	if len(got) == 0 {
		t.Fatalf("expected updates")
	}
	last := got[len(got)-1]
	if last.Status != agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_COMPLETED {
		t.Fatalf("expected completed, got %s", last.Status)
	}
	if last.Output != "hello world" {
		t.Fatalf("expected final output, got %q", last.Output)
	}
	if !hasStep(got, "Read files (in_progress)") || !hasStep(got, "Plan: inspect repo") {
		t.Fatalf("expected tool and plan steps, got %+v", got)
	}
}

func TestACPExecutorFailures(t *testing.T) {
	for _, mode := range []string{"init-error", "session-error", "prompt-error"} {
		t.Run(mode, func(t *testing.T) {
			exec := testACPExecutor(t, mode)
			task := &agentsv1.DaemonTask{TaskId: "task-1", AcpRuntime: "test-acp", Input: "hello"}
			updates, onUpdate := collectUpdates()

			if err := exec.Execute(context.Background(), task, onUpdate); err == nil {
				t.Fatalf("expected error")
			}

			got := updates()
			last := got[len(got)-1]
			if last.Status != agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_FAILED {
				t.Fatalf("expected failed, got %s", last.Status)
			}
			if !strings.Contains(last.Error, "acp") {
				t.Fatalf("expected acp error, got %q", last.Error)
			}
		})
	}
}

func TestACPExecutorCancellation(t *testing.T) {
	exec := testACPExecutor(t, "block")
	task := &agentsv1.DaemonTask{TaskId: "task-1", AcpRuntime: "test-acp", Input: "hello"}
	updates, onUpdate := collectUpdates()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- exec.Execute(ctx, task, onUpdate)
	}()

	waitForUpdate(t, updates, func(update *agentsv1.DaemonTaskUpdate) bool {
		return strings.Contains(update.Output, "started")
	})
	cancel()

	select {
	case err := <-errCh:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context canceled, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for cancellation")
	}

	got := updates()
	last := got[len(got)-1]
	if last.Status != agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_CANCELLED {
		t.Fatalf("expected cancelled, got %s", last.Status)
	}
}

func TestACPExecutorStopReasonCancelled(t *testing.T) {
	exec := testACPExecutor(t, "cancel-stop")
	task := &agentsv1.DaemonTask{TaskId: "task-1", AcpRuntime: "test-acp", Input: "hello"}
	updates, onUpdate := collectUpdates()

	if err := exec.Execute(context.Background(), task, onUpdate); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	got := updates()
	last := got[len(got)-1]
	if last.Status != agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_CANCELLED {
		t.Fatalf("expected cancelled, got %s", last.Status)
	}
}

func TestACPTaskClientPermissionPolicy(t *testing.T) {
	task := &agentsv1.DaemonTask{TaskId: "task-1"}
	_, onUpdate := collectUpdates()
	req := acp.RequestPermissionRequest{
		Options: []acp.PermissionOption{
			{OptionId: "reject", Kind: acp.PermissionOptionKindRejectOnce},
			{OptionId: "allow", Kind: acp.PermissionOptionKindAllowOnce},
		},
	}

	client := newACPTaskClient(context.Background(), task, ACPConfig{PermissionPolicy: PermissionPolicyAllow}, "", onUpdate)
	resp, err := client.RequestPermission(context.Background(), req)
	if err != nil {
		t.Fatalf("RequestPermission allow: %v", err)
	}
	if resp.Outcome.Selected == nil || resp.Outcome.Selected.OptionId != "allow" {
		t.Fatalf("expected allow selection, got %+v", resp.Outcome)
	}

	client = newACPTaskClient(context.Background(), task, ACPConfig{PermissionPolicy: PermissionPolicyDeny}, "", onUpdate)
	resp, err = client.RequestPermission(context.Background(), req)
	if err != nil {
		t.Fatalf("RequestPermission deny: %v", err)
	}
	if resp.Outcome.Cancelled == nil {
		t.Fatalf("expected cancelled outcome, got %+v", resp.Outcome)
	}
}

func TestACPTaskClientFileCallbacks(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "notes.txt"), []byte("a\nb\nc\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	client := newACPTaskClient(context.Background(), &agentsv1.DaemonTask{TaskId: "task-1"}, ACPConfig{
		WorkDir: root,
		FS:      ACPFSConfig{Read: true, Write: true},
	}, root, func(*agentsv1.DaemonTaskUpdate) {})

	read, err := client.ReadTextFile(context.Background(), acp.ReadTextFileRequest{
		Path:  "notes.txt",
		Line:  acp.Ptr(2),
		Limit: acp.Ptr(1),
	})
	if err != nil {
		t.Fatalf("ReadTextFile: %v", err)
	}
	if read.Content != "b" {
		t.Fatalf("expected line-limited content, got %q", read.Content)
	}

	if _, err := client.WriteTextFile(context.Background(), acp.WriteTextFileRequest{
		Path:    "nested/out.txt",
		Content: "ok",
	}); err != nil {
		t.Fatalf("WriteTextFile: %v", err)
	}
	if got, err := os.ReadFile(filepath.Join(root, "nested", "out.txt")); err != nil || string(got) != "ok" {
		t.Fatalf("expected written content, got %q err=%v", string(got), err)
	}

	if _, err := client.ReadTextFile(context.Background(), acp.ReadTextFileRequest{Path: "../escape.txt"}); err == nil {
		t.Fatalf("expected path escape error")
	}
	if _, err := client.ReadTextFile(context.Background(), acp.ReadTextFileRequest{Path: filepath.Join(root, "notes.txt")}); err == nil {
		t.Fatalf("expected absolute path rejection")
	}
}

func TestACPTaskClientTerminalCallbacks(t *testing.T) {
	root := t.TempDir()
	client := newACPTaskClient(context.Background(), &agentsv1.DaemonTask{TaskId: "task-1"}, ACPConfig{
		WorkDir:  root,
		Terminal: true,
	}, root, func(*agentsv1.DaemonTaskUpdate) {})
	resp, err := client.CreateTerminal(context.Background(), acp.CreateTerminalRequest{
		Command: "sh",
		Args:    []string{"-c", "printf hello"},
	})
	if err != nil {
		t.Fatalf("CreateTerminal: %v", err)
	}
	if _, err := client.WaitForTerminalExit(context.Background(), acp.WaitForTerminalExitRequest{TerminalId: resp.TerminalId}); err != nil {
		t.Fatalf("WaitForTerminalExit: %v", err)
	}
	out, err := client.TerminalOutput(context.Background(), acp.TerminalOutputRequest{TerminalId: resp.TerminalId})
	if err != nil {
		t.Fatalf("TerminalOutput: %v", err)
	}
	if out.Output != "hello" {
		t.Fatalf("expected terminal output, got %q", out.Output)
	}
	if _, err := client.ReleaseTerminal(context.Background(), acp.ReleaseTerminalRequest{TerminalId: resp.TerminalId}); err != nil {
		t.Fatalf("ReleaseTerminal: %v", err)
	}
}

func TestHelperProcessACPAgent(t *testing.T) {
	if os.Getenv("BUTTER_ACP_HELPER") != "1" {
		return
	}
	mode := os.Getenv("BUTTER_ACP_HELPER_MODE")
	agent := &testAgent{mode: mode, sessions: make(map[string]context.CancelFunc)}
	conn := acp.NewAgentSideConnection(agent, os.Stdout, os.Stdin)
	agent.conn = conn
	<-conn.Done()
	os.Exit(0)
}

type testAgent struct {
	mode     string
	conn     *acp.AgentSideConnection
	mu       sync.Mutex
	sessions map[string]context.CancelFunc
}

func (a *testAgent) Authenticate(context.Context, acp.AuthenticateRequest) (acp.AuthenticateResponse, error) {
	return acp.AuthenticateResponse{}, nil
}

func (a *testAgent) Initialize(context.Context, acp.InitializeRequest) (acp.InitializeResponse, error) {
	if a.mode == "init-error" {
		return acp.InitializeResponse{}, fmt.Errorf("init boom")
	}
	return acp.InitializeResponse{ProtocolVersion: acp.ProtocolVersionNumber, AuthMethods: []acp.AuthMethod{}}, nil
}

func (a *testAgent) Logout(context.Context, acp.LogoutRequest) (acp.LogoutResponse, error) {
	return acp.LogoutResponse{}, nil
}

func (a *testAgent) Cancel(_ context.Context, params acp.CancelNotification) error {
	a.mu.Lock()
	cancel := a.sessions[string(params.SessionId)]
	a.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	return nil
}

func (a *testAgent) CloseSession(context.Context, acp.CloseSessionRequest) (acp.CloseSessionResponse, error) {
	return acp.CloseSessionResponse{}, nil
}

func (a *testAgent) ListSessions(context.Context, acp.ListSessionsRequest) (acp.ListSessionsResponse, error) {
	return acp.ListSessionsResponse{}, nil
}

func (a *testAgent) NewSession(context.Context, acp.NewSessionRequest) (acp.NewSessionResponse, error) {
	if a.mode == "session-error" {
		return acp.NewSessionResponse{}, fmt.Errorf("session boom")
	}
	return acp.NewSessionResponse{SessionId: "session-1"}, nil
}

func (a *testAgent) Prompt(ctx context.Context, params acp.PromptRequest) (acp.PromptResponse, error) {
	if a.mode == "prompt-error" {
		return acp.PromptResponse{}, fmt.Errorf("prompt boom")
	}
	promptCtx, cancel := context.WithCancel(ctx)
	a.mu.Lock()
	a.sessions[string(params.SessionId)] = cancel
	a.mu.Unlock()
	defer func() {
		a.mu.Lock()
		delete(a.sessions, string(params.SessionId))
		a.mu.Unlock()
		cancel()
	}()

	if a.mode == "block" {
		_ = a.conn.SessionUpdate(promptCtx, acp.SessionNotification{
			SessionId: params.SessionId,
			Update:    acp.UpdateAgentMessageText("started"),
		})
		<-promptCtx.Done()
		return acp.PromptResponse{StopReason: acp.StopReasonCancelled}, promptCtx.Err()
	}
	if a.mode == "cancel-stop" {
		return acp.PromptResponse{StopReason: acp.StopReasonCancelled}, nil
	}

	if err := a.conn.SessionUpdate(promptCtx, acp.SessionNotification{
		SessionId: params.SessionId,
		Update: acp.UpdatePlan(acp.PlanEntry{
			Content: "inspect repo",
			Status:  acp.PlanEntryStatusInProgress,
		}),
	}); err != nil {
		return acp.PromptResponse{}, err
	}
	if err := a.conn.SessionUpdate(promptCtx, acp.SessionNotification{
		SessionId: params.SessionId,
		Update: acp.StartToolCall(
			"tool-1",
			"Read files",
			acp.WithStartStatus(acp.ToolCallStatusInProgress),
		),
	}); err != nil {
		return acp.PromptResponse{}, err
	}
	if err := a.conn.SessionUpdate(promptCtx, acp.SessionNotification{
		SessionId: params.SessionId,
		Update:    acp.UpdateAgentThoughtText("hidden thought"),
	}); err != nil {
		return acp.PromptResponse{}, err
	}
	if err := a.conn.SessionUpdate(promptCtx, acp.SessionNotification{
		SessionId: params.SessionId,
		Update:    acp.UpdateAgentMessageText("hello "),
	}); err != nil {
		return acp.PromptResponse{}, err
	}
	if err := a.conn.SessionUpdate(promptCtx, acp.SessionNotification{
		SessionId: params.SessionId,
		Update:    acp.UpdateAgentMessageText("world"),
	}); err != nil {
		return acp.PromptResponse{}, err
	}
	return acp.PromptResponse{StopReason: acp.StopReasonEndTurn}, nil
}

func (a *testAgent) ResumeSession(context.Context, acp.ResumeSessionRequest) (acp.ResumeSessionResponse, error) {
	return acp.ResumeSessionResponse{}, nil
}

func (a *testAgent) SetSessionConfigOption(context.Context, acp.SetSessionConfigOptionRequest) (acp.SetSessionConfigOptionResponse, error) {
	return acp.SetSessionConfigOptionResponse{}, nil
}

func (a *testAgent) SetSessionMode(context.Context, acp.SetSessionModeRequest) (acp.SetSessionModeResponse, error) {
	return acp.SetSessionModeResponse{}, nil
}

func testACPExecutor(t *testing.T, mode string) *ACPExecutor {
	t.Helper()
	exec, err := NewACPExecutor(ACPConfig{
		Runtime:          "test-acp",
		Command:          os.Args[0],
		Args:             []string{"-test.run=TestHelperProcessACPAgent"},
		Env:              map[string]string{"BUTTER_ACP_HELPER": "1", "BUTTER_ACP_HELPER_MODE": mode},
		WorkDir:          t.TempDir(),
		PermissionPolicy: PermissionPolicyDeny,
		FS:               ACPFSConfig{Read: true, Write: true},
		Terminal:         true,
	})
	if err != nil {
		t.Fatalf("NewACPExecutor: %v", err)
	}
	return exec
}

func collectUpdates() (func() []*agentsv1.DaemonTaskUpdate, func(*agentsv1.DaemonTaskUpdate)) {
	var mu sync.Mutex
	var updates []*agentsv1.DaemonTaskUpdate
	return func() []*agentsv1.DaemonTaskUpdate {
			mu.Lock()
			defer mu.Unlock()
			out := make([]*agentsv1.DaemonTaskUpdate, len(updates))
			copy(out, updates)
			return out
		}, func(update *agentsv1.DaemonTaskUpdate) {
			mu.Lock()
			defer mu.Unlock()
			updates = append(updates, proto.Clone(update).(*agentsv1.DaemonTaskUpdate))
		}
}

func hasStep(updates []*agentsv1.DaemonTaskUpdate, step string) bool {
	for _, update := range updates {
		if update.CurrentStep == step {
			return true
		}
	}
	return false
}

func waitForUpdate(t *testing.T, updates func() []*agentsv1.DaemonTaskUpdate, match func(*agentsv1.DaemonTaskUpdate) bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		for _, update := range updates() {
			if match(update) {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for matching update; got %+v", updates())
}

func TestNoHelperProcessLeak(t *testing.T) {
	if os.Getenv("BUTTER_ACP_HELPER") == "1" {
		return
	}
	if _, err := exec.LookPath(os.Args[0]); err != nil {
		t.Fatalf("test binary is not executable: %v", err)
	}
}
