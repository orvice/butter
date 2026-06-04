package executor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/coder/acp-go-sdk"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	// PermissionPolicyDeny denies ACP permission requests without local input.
	PermissionPolicyDeny = "deny"
	// PermissionPolicyAllow selects an allow option when an ACP agent asks for permission.
	PermissionPolicyAllow = "allow"

	acpCancelTimeout = 2 * time.Second
	acpKillTimeout   = 2 * time.Second
)

// ACPConfig configures one ACP-backed daemon executor.
type ACPConfig struct {
	Capability       string            `yaml:"capability"`
	Command          string            `yaml:"command"`
	Args             []string          `yaml:"args"`
	Env              map[string]string `yaml:"env"`
	WorkDir          string            `yaml:"work_dir"`
	PermissionPolicy string            `yaml:"permission_policy"`
	FS               ACPFSConfig       `yaml:"fs"`
	Terminal         bool              `yaml:"terminal"`
}

// ACPFSConfig controls file-system callbacks exposed to ACP agents.
type ACPFSConfig struct {
	Read               bool `yaml:"read"`
	Write              bool `yaml:"write"`
	AllowAbsolutePaths bool `yaml:"allow_absolute_paths"`
}

// ACPExecutor runs ACP-compatible coding agents over stdio.
type ACPExecutor struct {
	cfg ACPConfig
}

// NewACPExecutor validates config and creates an ACP executor.
func NewACPExecutor(cfg ACPConfig) (*ACPExecutor, error) {
	if cfg.Capability == "" {
		return nil, errors.New("acp capability is required")
	}
	if cfg.Command == "" {
		return nil, fmt.Errorf("acp executor %q: command is required", cfg.Capability)
	}
	if cfg.WorkDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("resolve work dir: %w", err)
		}
		cfg.WorkDir = wd
	}
	if !filepath.IsAbs(cfg.WorkDir) {
		abs, err := filepath.Abs(cfg.WorkDir)
		if err != nil {
			return nil, fmt.Errorf("resolve work dir %q: %w", cfg.WorkDir, err)
		}
		cfg.WorkDir = abs
	}
	cfg.WorkDir = filepath.Clean(cfg.WorkDir)
	switch cfg.PermissionPolicy {
	case "", PermissionPolicyDeny:
		cfg.PermissionPolicy = PermissionPolicyDeny
	case PermissionPolicyAllow:
	default:
		return nil, fmt.Errorf("acp executor %q: unsupported permission_policy %q", cfg.Capability, cfg.PermissionPolicy)
	}
	return &ACPExecutor{cfg: cfg}, nil
}

func (e *ACPExecutor) Capability() string { return e.cfg.Capability }

func (e *ACPExecutor) Execute(ctx context.Context, task *agentsv1.DaemonTask, onUpdate func(*agentsv1.DaemonTaskUpdate)) error {
	onUpdate(&agentsv1.DaemonTaskUpdate{
		TaskId: task.TaskId,
		Status: agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_ACCEPTED,
	})

	cmd := exec.CommandContext(ctx, e.cfg.Command, e.cfg.Args...)
	cmd.Dir = e.cfg.WorkDir
	cmd.Env = mergeEnv(e.cfg.Env)
	cmd.Stderr = newLogWriter("acp stderr", task.TaskId)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return sendFailed(task.TaskId, fmt.Sprintf("acp stdin pipe: %v", err), onUpdate)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return sendFailed(task.TaskId, fmt.Sprintf("acp stdout pipe: %v", err), onUpdate)
	}
	if err := cmd.Start(); err != nil {
		return sendFailed(task.TaskId, fmt.Sprintf("acp start: %v", err), onUpdate)
	}
	defer terminateProcess(cmd, acpKillTimeout)

	taskClient := newACPTaskClient(ctx, task, e.cfg, onUpdate)
	defer taskClient.cleanupTerminals(acpKillTimeout)
	conn := acp.NewClientSideConnection(taskClient, stdin, stdout)
	conn.SetLogger(slog.Default())

	onUpdate(&agentsv1.DaemonTaskUpdate{
		TaskId: task.TaskId,
		Status: agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_RUNNING,
	})

	initResp, err := conn.Initialize(ctx, acp.InitializeRequest{
		ProtocolVersion: acp.ProtocolVersionNumber,
		ClientInfo: &acp.Implementation{
			Name:    "butter-daemon",
			Version: "v0.1.0",
		},
		ClientCapabilities: acp.ClientCapabilities{
			Fs: acp.FileSystemCapabilities{
				ReadTextFile:  e.cfg.FS.Read,
				WriteTextFile: e.cfg.FS.Write,
			},
			Terminal: e.cfg.Terminal,
		},
	})
	if err != nil {
		return sendACPFailed(task.TaskId, "initialize", err, onUpdate)
	}
	slog.Info("acp agent initialized",
		"task_id", task.TaskId,
		"capability", e.cfg.Capability,
		"protocol_version", initResp.ProtocolVersion,
	)

	session, err := conn.NewSession(ctx, acp.NewSessionRequest{
		Cwd:        e.cfg.WorkDir,
		McpServers: []acp.McpServer{},
	})
	if err != nil {
		return sendACPFailed(task.TaskId, "new session", err, onUpdate)
	}
	taskClient.setSession(conn, session.SessionId)

	promptDone := make(chan acpPromptResult, 1)
	go func() {
		resp, err := conn.Prompt(ctx, acp.PromptRequest{
			SessionId: session.SessionId,
			Prompt:    []acp.ContentBlock{acp.TextBlock(task.Input)},
		})
		promptDone <- acpPromptResult{resp: resp, err: err}
	}()

	select {
	case result := <-promptDone:
		if result.err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				taskClient.cancelSession(acpCancelTimeout)
				onUpdate(&agentsv1.DaemonTaskUpdate{
					TaskId: task.TaskId,
					Status: agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_CANCELLED,
					Output: taskClient.output(),
				})
				return ctx.Err()
			}
			return sendACPFailed(task.TaskId, "prompt", result.err, onUpdate)
		}
		if result.resp.StopReason == acp.StopReasonCancelled {
			onUpdate(&agentsv1.DaemonTaskUpdate{
				TaskId: task.TaskId,
				Status: agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_CANCELLED,
				Output: taskClient.output(),
			})
			return nil
		}
	case <-ctx.Done():
		taskClient.cancelSession(acpCancelTimeout)
		onUpdate(&agentsv1.DaemonTaskUpdate{
			TaskId: task.TaskId,
			Status: agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_CANCELLED,
			Output: taskClient.output(),
		})
		return ctx.Err()
	}

	onUpdate(&agentsv1.DaemonTaskUpdate{
		TaskId: task.TaskId,
		Status: agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_COMPLETED,
		Output: taskClient.output(),
	})
	return nil
}

type acpPromptResult struct {
	resp acp.PromptResponse
	err  error
}

func sendACPFailed(taskID, phase string, err error, onUpdate func(*agentsv1.DaemonTaskUpdate)) error {
	return sendFailed(taskID, fmt.Sprintf("acp %s: %s", phase, formatACPError(err)), onUpdate)
}

func formatACPError(err error) string {
	var reqErr *acp.RequestError
	if errors.As(err, &reqErr) {
		if reqErr.Data != nil {
			return fmt.Sprintf("json-rpc %d: %s (%v)", reqErr.Code, reqErr.Message, reqErr.Data)
		}
		return fmt.Sprintf("json-rpc %d: %s", reqErr.Code, reqErr.Message)
	}
	return err.Error()
}

type acpTaskClient struct {
	ctx      context.Context
	task     *agentsv1.DaemonTask
	cfg      ACPConfig
	onUpdate func(*agentsv1.DaemonTaskUpdate)

	mu        sync.Mutex
	buf       strings.Builder
	step      string
	conn      *acp.ClientSideConnection
	sessionID acp.SessionId
	terms     map[string]*terminalState
	nextTerm  atomic.Uint64
}

func newACPTaskClient(ctx context.Context, task *agentsv1.DaemonTask, cfg ACPConfig, onUpdate func(*agentsv1.DaemonTaskUpdate)) *acpTaskClient {
	return &acpTaskClient{
		ctx:      ctx,
		task:     task,
		cfg:      cfg,
		onUpdate: onUpdate,
		terms:    make(map[string]*terminalState),
	}
}

func (c *acpTaskClient) setSession(conn *acp.ClientSideConnection, sessionID acp.SessionId) {
	c.mu.Lock()
	c.conn = conn
	c.sessionID = sessionID
	c.mu.Unlock()
}

func (c *acpTaskClient) cancelSession(timeout time.Duration) {
	c.mu.Lock()
	conn := c.conn
	sessionID := c.sessionID
	c.mu.Unlock()
	if conn == nil || sessionID == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := conn.Cancel(ctx, acp.CancelNotification{SessionId: sessionID}); err != nil {
		slog.Warn("failed to send acp session cancel", "task_id", c.task.TaskId, "err", err)
	}
}

func (c *acpTaskClient) output() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.buf.String()
}

func (c *acpTaskClient) RequestPermission(ctx context.Context, params acp.RequestPermissionRequest) (acp.RequestPermissionResponse, error) {
	selected := acp.PermissionOptionId("")
	if c.cfg.PermissionPolicy == PermissionPolicyAllow {
		// Only ever auto-select an explicit allow option; never fall back to an
		// arbitrary option, which could be a reject choice.
		for _, opt := range params.Options {
			if opt.Kind == acp.PermissionOptionKindAllowOnce || opt.Kind == acp.PermissionOptionKindAllowAlways {
				selected = opt.OptionId
				break
			}
		}
	}

	if selected == "" {
		c.emitRunning("Permission denied")
		return acp.RequestPermissionResponse{Outcome: acp.NewRequestPermissionOutcomeCancelled()}, nil
	}

	c.emitRunning("Permission approved")
	return acp.RequestPermissionResponse{Outcome: acp.NewRequestPermissionOutcomeSelected(selected)}, nil
}

func (c *acpTaskClient) SessionUpdate(ctx context.Context, params acp.SessionNotification) error {
	u := params.Update
	switch {
	case u.AgentMessageChunk != nil:
		text := contentText(u.AgentMessageChunk.Content)
		if text != "" {
			c.mu.Lock()
			c.buf.WriteString(text)
			output := c.buf.String()
			step := c.step
			c.mu.Unlock()
			c.onUpdate(&agentsv1.DaemonTaskUpdate{
				TaskId:      c.task.TaskId,
				Status:      agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_RUNNING,
				Output:      output,
				CurrentStep: step,
			})
		}
	case u.AgentThoughtChunk != nil:
		c.emitRunning("Thinking")
	case u.ToolCall != nil:
		step := u.ToolCall.Title
		if u.ToolCall.Status != "" {
			step = fmt.Sprintf("%s (%s)", step, u.ToolCall.Status)
		}
		c.emitRunning(step)
	case u.ToolCallUpdate != nil:
		step := fmt.Sprintf("Tool %s", u.ToolCallUpdate.ToolCallId)
		if u.ToolCallUpdate.Title != nil && *u.ToolCallUpdate.Title != "" {
			step = *u.ToolCallUpdate.Title
		}
		if u.ToolCallUpdate.Status != nil && *u.ToolCallUpdate.Status != "" {
			step = fmt.Sprintf("%s (%s)", step, *u.ToolCallUpdate.Status)
		}
		c.emitRunning(step)
	case u.Plan != nil:
		c.emitRunning(summarizePlan(u.Plan.Entries))
	case u.PlanUpdate != nil:
		c.emitRunning("Plan updated")
	case u.PlanRemoved != nil:
		c.emitRunning("Plan removed")
	}
	return nil
}

func (c *acpTaskClient) emitRunning(step string) {
	c.mu.Lock()
	c.step = step
	output := c.buf.String()
	c.mu.Unlock()
	c.onUpdate(&agentsv1.DaemonTaskUpdate{
		TaskId:      c.task.TaskId,
		Status:      agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_RUNNING,
		Output:      output,
		CurrentStep: step,
	})
}

func (c *acpTaskClient) ReadTextFile(ctx context.Context, params acp.ReadTextFileRequest) (acp.ReadTextFileResponse, error) {
	if !c.cfg.FS.Read {
		return acp.ReadTextFileResponse{}, errors.New("file read is disabled")
	}
	path, err := c.resolvePath(params.Path)
	if err != nil {
		return acp.ReadTextFileResponse{}, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return acp.ReadTextFileResponse{}, err
	}
	content := applyLineLimit(string(b), params.Line, params.Limit)
	return acp.ReadTextFileResponse{Content: content}, nil
}

func (c *acpTaskClient) WriteTextFile(ctx context.Context, params acp.WriteTextFileRequest) (acp.WriteTextFileResponse, error) {
	if !c.cfg.FS.Write {
		return acp.WriteTextFileResponse{}, errors.New("file write is disabled")
	}
	path, err := c.resolvePath(params.Path)
	if err != nil {
		return acp.WriteTextFileResponse{}, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return acp.WriteTextFileResponse{}, err
	}
	if err := os.WriteFile(path, []byte(params.Content), 0o644); err != nil {
		return acp.WriteTextFileResponse{}, err
	}
	return acp.WriteTextFileResponse{}, nil
}

func (c *acpTaskClient) resolvePath(path string) (string, error) {
	if path == "" {
		return "", errors.New("path is required")
	}
	var resolved string
	if filepath.IsAbs(path) {
		if !c.cfg.FS.AllowAbsolutePaths {
			return "", fmt.Errorf("absolute paths are disabled: %s", path)
		}
		resolved = filepath.Clean(path)
	} else {
		resolved = filepath.Clean(filepath.Join(c.cfg.WorkDir, path))
	}
	if !c.cfg.FS.AllowAbsolutePaths && !pathWithin(c.cfg.WorkDir, resolved) {
		return "", fmt.Errorf("path escapes work_dir: %s", path)
	}
	if !c.cfg.FS.AllowAbsolutePaths {
		// Resolve symlinks on the deepest existing ancestor so a symlink inside
		// work_dir cannot point the real target outside it.
		if real, err := evalSymlinkContainment(c.cfg.WorkDir, resolved); err != nil {
			return "", err
		} else {
			resolved = real
		}
	}
	return resolved, nil
}

// evalSymlinkContainment resolves symlinks on the longest existing prefix of
// path and verifies the real location stays within root. The path may not exist
// yet (e.g. a file about to be written), so non-existent tail segments are
// re-appended after resolving the existing ancestor.
func evalSymlinkContainment(root, path string) (string, error) {
	existing := path
	var tail []string
	for {
		if _, err := os.Lstat(existing); err == nil {
			break
		}
		parent := filepath.Dir(existing)
		if parent == existing {
			break
		}
		tail = append([]string{filepath.Base(existing)}, tail...)
		existing = parent
	}
	realExisting, err := filepath.EvalSymlinks(existing)
	if err != nil {
		return "", err
	}
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	resolved := filepath.Join(append([]string{realExisting}, tail...)...)
	if !pathWithin(realRoot, resolved) {
		return "", fmt.Errorf("path escapes work_dir via symlink: %s", path)
	}
	return resolved, nil
}

func (c *acpTaskClient) CreateTerminal(ctx context.Context, params acp.CreateTerminalRequest) (acp.CreateTerminalResponse, error) {
	if !c.cfg.Terminal {
		return acp.CreateTerminalResponse{}, errors.New("terminal support is disabled")
	}
	cwd := c.cfg.WorkDir
	if params.Cwd != nil && *params.Cwd != "" {
		resolved, err := c.resolvePath(*params.Cwd)
		if err != nil {
			return acp.CreateTerminalResponse{}, err
		}
		cwd = resolved
	}
	cmd := exec.CommandContext(c.ctx, params.Command, params.Args...)
	cmd.Dir = cwd
	cmd.Env = mergeACPEnv(params.Env)

	term := &terminalState{limit: params.OutputByteLimit}
	var out terminalBuffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	term.output = &out

	if err := cmd.Start(); err != nil {
		return acp.CreateTerminalResponse{}, err
	}
	term.cmd = cmd
	term.done = make(chan struct{})
	go func() {
		err := cmd.Wait()
		term.mu.Lock()
		term.err = err
		term.exit = exitStatus(err)
		term.mu.Unlock()
		close(term.done)
	}()

	id := fmt.Sprintf("term-%d", c.nextTerm.Add(1))
	c.mu.Lock()
	c.terms[id] = term
	c.mu.Unlock()
	return acp.CreateTerminalResponse{TerminalId: id}, nil
}

func (c *acpTaskClient) TerminalOutput(ctx context.Context, params acp.TerminalOutputRequest) (acp.TerminalOutputResponse, error) {
	term, err := c.terminal(params.TerminalId)
	if err != nil {
		return acp.TerminalOutputResponse{}, err
	}
	output, truncated := term.snapshot()
	term.mu.Lock()
	exit := term.exit
	term.mu.Unlock()
	return acp.TerminalOutputResponse{Output: output, Truncated: truncated, ExitStatus: exit}, nil
}

func (c *acpTaskClient) WaitForTerminalExit(ctx context.Context, params acp.WaitForTerminalExitRequest) (acp.WaitForTerminalExitResponse, error) {
	term, err := c.terminal(params.TerminalId)
	if err != nil {
		return acp.WaitForTerminalExitResponse{}, err
	}
	select {
	case <-term.done:
	case <-ctx.Done():
		return acp.WaitForTerminalExitResponse{}, ctx.Err()
	}
	term.mu.Lock()
	defer term.mu.Unlock()
	if term.exit == nil {
		return acp.WaitForTerminalExitResponse{}, nil
	}
	return acp.WaitForTerminalExitResponse{
		ExitCode: term.exit.ExitCode,
		Signal:   term.exit.Signal,
	}, nil
}

func (c *acpTaskClient) KillTerminal(ctx context.Context, params acp.KillTerminalRequest) (acp.KillTerminalResponse, error) {
	term, err := c.terminal(params.TerminalId)
	if err != nil {
		return acp.KillTerminalResponse{}, err
	}
	terminateTerminal(term, acpKillTimeout)
	return acp.KillTerminalResponse{}, nil
}

func (c *acpTaskClient) ReleaseTerminal(ctx context.Context, params acp.ReleaseTerminalRequest) (acp.ReleaseTerminalResponse, error) {
	c.mu.Lock()
	term, ok := c.terms[params.TerminalId]
	if ok {
		delete(c.terms, params.TerminalId)
	}
	c.mu.Unlock()
	if !ok {
		return acp.ReleaseTerminalResponse{}, fmt.Errorf("unknown terminal: %s", params.TerminalId)
	}
	terminateTerminal(term, acpKillTimeout)
	return acp.ReleaseTerminalResponse{}, nil
}

func (c *acpTaskClient) cleanupTerminals(timeout time.Duration) {
	c.mu.Lock()
	terms := make([]*terminalState, 0, len(c.terms))
	for id, term := range c.terms {
		terms = append(terms, term)
		delete(c.terms, id)
	}
	c.mu.Unlock()
	for _, term := range terms {
		terminateTerminal(term, timeout)
	}
}

func (c *acpTaskClient) terminal(id string) (*terminalState, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	term, ok := c.terms[id]
	if !ok {
		return nil, fmt.Errorf("unknown terminal: %s", id)
	}
	return term, nil
}

type terminalState struct {
	cmd    *exec.Cmd
	done   chan struct{}
	output *terminalBuffer
	limit  *int

	mu   sync.Mutex
	err  error
	exit *acp.TerminalExitStatus
}

func (t *terminalState) snapshot() (string, bool) {
	output := t.output.String()
	if t.limit == nil || *t.limit <= 0 || len(output) <= *t.limit {
		return output, false
	}
	return trimUTF8Suffix(output, *t.limit), true
}

type terminalBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *terminalBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *terminalBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func contentText(block acp.ContentBlock) string {
	switch {
	case block.Text != nil:
		return block.Text.Text
	case block.Resource != nil && block.Resource.Resource.TextResourceContents != nil:
		return block.Resource.Resource.TextResourceContents.Text
	default:
		return ""
	}
}

func summarizePlan(entries []acp.PlanEntry) string {
	for _, e := range entries {
		if e.Status == acp.PlanEntryStatusInProgress && e.Content != "" {
			return "Plan: " + e.Content
		}
	}
	for _, e := range entries {
		if e.Status == acp.PlanEntryStatusPending && e.Content != "" {
			return "Plan: " + e.Content
		}
	}
	if len(entries) > 0 && entries[0].Content != "" {
		return "Plan: " + entries[0].Content
	}
	return "Plan updated"
}

func applyLineLimit(content string, line, limit *int) string {
	if line == nil && limit == nil {
		return content
	}
	lines := strings.Split(content, "\n")
	start := 0
	if line != nil && *line > 0 {
		start = min(*line-1, len(lines))
	}
	end := len(lines)
	if limit != nil && *limit > 0 {
		end = min(start+*limit, end)
	}
	return strings.Join(lines[start:end], "\n")
}

func pathWithin(root, path string) bool {
	root = filepath.Clean(root)
	path = filepath.Clean(path)
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)))
}

func mergeEnv(extra map[string]string) []string {
	env := os.Environ()
	if len(extra) == 0 {
		return env
	}
	keys := make([]string, 0, len(extra))
	for k := range extra {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		env = append(env, k+"="+extra[k])
	}
	return env
}

func mergeACPEnv(extra []acp.EnvVariable) []string {
	env := os.Environ()
	for _, item := range extra {
		if item.Name != "" {
			env = append(env, item.Name+"="+item.Value)
		}
	}
	return env
}

func terminateProcess(cmd *exec.Cmd, timeout time.Duration) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	done := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(done)
	}()
	select {
	case <-done:
		return
	default:
	}
	_ = cmd.Process.Signal(os.Interrupt)
	select {
	case <-done:
	case <-time.After(timeout):
		_ = cmd.Process.Kill()
		<-done
	}
}

func terminateTerminal(term *terminalState, timeout time.Duration) {
	if term == nil || term.cmd == nil || term.cmd.Process == nil {
		return
	}
	select {
	case <-term.done:
		return
	default:
	}
	_ = term.cmd.Process.Signal(os.Interrupt)
	select {
	case <-term.done:
	case <-time.After(timeout):
		_ = term.cmd.Process.Kill()
		<-term.done
	}
}

func exitStatus(err error) *acp.TerminalExitStatus {
	if err == nil {
		code := 0
		return &acp.TerminalExitStatus{ExitCode: &code}
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			if status.Signaled() {
				sig := status.Signal().String()
				return &acp.TerminalExitStatus{Signal: &sig}
			}
			code := status.ExitStatus()
			return &acp.TerminalExitStatus{ExitCode: &code}
		}
	}
	code := 1
	return &acp.TerminalExitStatus{ExitCode: &code}
}

func trimUTF8Suffix(s string, maxBytes int) string {
	if maxBytes <= 0 || len(s) <= maxBytes {
		return s
	}
	out := s[len(s)-maxBytes:]
	for len(out) > 0 && !utf8.ValidString(out) {
		out = out[1:]
	}
	return out
}

type logWriter struct {
	prefix string
	taskID string
}

func newLogWriter(prefix, taskID string) io.Writer {
	return &logWriter{prefix: prefix, taskID: taskID}
}

func (w *logWriter) Write(p []byte) (int, error) {
	text := strings.TrimSpace(string(p))
	if text != "" {
		slog.Debug(w.prefix, "task_id", w.taskID, "output", text)
	}
	return len(p), nil
}
