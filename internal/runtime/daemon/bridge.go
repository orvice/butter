package daemon

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/session"
	"google.golang.org/genai"

	"github.com/google/uuid"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Bridge wraps daemon execution behind the ADK agent interface.
type Bridge struct {
	registry    *Registry
	workspaceID string
	runtimeID   string
	acpRuntime  string
	workDirRoot string
	metrics     *Metrics
}

// NewBridge creates a bridge for a daemon runtime and ACP runtime pair. It
// pulls the metrics collector off the registry so bridge invocations are recorded.
func NewBridge(registry *Registry, workspaceID, runtimeID, acpRuntime string) *Bridge {
	b := &Bridge{
		registry:    registry,
		workspaceID: workspaceID,
		runtimeID:   runtimeID,
		acpRuntime:  acpRuntime,
		workDirRoot: filepath.Join(os.TempDir(), "butter-daemon-workdirs"),
	}
	if registry != nil {
		b.metrics = registry.Metrics()
	}
	return b
}

// SetMetrics wires a metrics collector that will receive per-invocation
// latency observations.
func (b *Bridge) SetMetrics(m *Metrics) {
	b.metrics = m
}

// BuildAgent produces an agent.Agent via agent.New(). This is required because
// ADK v1.1.0's agent.Agent interface has an unexported internal() method.
func (b *Bridge) BuildAgent(name, description string) (agent.Agent, error) {
	return agent.New(agent.Config{
		Name:        name,
		Description: description,
		Run:         b.run,
	})
}

// run implements the ADK agent Run function signature.
func (b *Bridge) run(ctx agent.InvocationContext) iter.Seq2[*session.Event, error] {
	return func(yield func(*session.Event, error) bool) {
		started := time.Now()
		defer func() {
			if b.metrics != nil {
				b.metrics.RecordLatency(time.Since(started))
			}
		}()

		input := extractText(ctx.UserContent())

		conn := b.registry.Get(b.workspaceID, b.runtimeID)
		if conn == nil {
			yield(nil, fmt.Errorf("daemon runtime %q is not online", b.runtimeID))
			return
		}
		workDir, err := b.workDir(ctx.Session().ID())
		if err != nil {
			yield(nil, err)
			return
		}

		task := &agentsv1.DaemonTask{
			TaskId:          uuid.NewString(),
			AgentName:       ctx.Agent().Name(),
			Input:           input,
			SessionId:       ctx.Session().ID(),
			UserId:          ctx.Session().UserID(),
			WorkspaceId:     b.workspaceID,
			DaemonRuntimeId: b.runtimeID,
			AcpRuntime:      b.acpRuntime,
			WorkDir:         workDir,
		}

		resultCh, err := conn.SendTask(task)
		if err != nil {
			yield(nil, fmt.Errorf("daemon send task: %w", err))
			return
		}

		var lastOutput string
		for {
			select {
			case update, ok := <-resultCh:
				if !ok {
					// Channel closed without terminal — daemon disconnected.
					yield(nil, ErrDaemonDisconnected)
					return
				}
				switch update.Status {
				case agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_COMPLETED:
					event := session.NewEvent(ctx.InvocationID())
					event.Author = ctx.Agent().Name()
					event.Content = genai.NewContentFromText(update.Output, genai.RoleModel)
					yield(event, nil)
					return
				case agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_FAILED:
					yield(nil, fmt.Errorf("daemon task failed: %s", update.Error))
					return
				case agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_CANCELLED:
					yield(nil, fmt.Errorf("daemon task cancelled"))
					return
				case agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_RUNNING,
					agentsv1.DaemonTaskStatus_DAEMON_TASK_STATUS_ACCEPTED:
					lastOutput = update.Output
				}
			case <-ctx.Done():
				_ = conn.CancelTask(task.TaskId)
				if lastOutput != "" {
					event := session.NewEvent(ctx.InvocationID())
					event.Author = ctx.Agent().Name()
					event.LLMResponse.Content = genai.NewContentFromText(lastOutput, genai.RoleModel)
					yield(event, ctx.Err())
				} else {
					yield(nil, ctx.Err())
				}
				return
			}
		}
	}
}

func (b *Bridge) workDir(sessionID string) (string, error) {
	seed := b.workspaceID + ":" + sessionID
	if seed == ":" {
		seed = b.workspaceID + ":" + uuid.NewString()
	}
	sum := sha256.Sum256([]byte(seed))
	name := hex.EncodeToString(sum[:])[:32]
	dir := filepath.Join(b.workDirRoot, name)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("create daemon work_dir: %w", err)
	}
	return dir, nil
}

// extractText pulls text from a genai.Content, joining all text parts.
func extractText(c *genai.Content) string {
	if c == nil {
		return ""
	}
	var parts []string
	for _, p := range c.Parts {
		if p.Text != "" {
			parts = append(parts, p.Text)
		}
	}
	return strings.Join(parts, "\n")
}
