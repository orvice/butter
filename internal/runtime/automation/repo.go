package automation

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

var (
	ErrAutomationNotFound      = errors.New("automation not found")
	ErrAutomationAlreadyExists = errors.New("automation already exists")
	ErrRunNotFound             = errors.New("automation run not found")
	ErrStepRunNotFound         = errors.New("automation step run not found")
)

// DefinitionRepo persists workspace-scoped automation definitions.
type DefinitionRepo interface {
	EnsureIndexes(ctx context.Context) error
	List(ctx context.Context, workspaceID string) ([]*agentsv1.Automation, error)
	ListAll(ctx context.Context) ([]*agentsv1.Automation, error)
	Get(ctx context.Context, workspaceID, name string) (*agentsv1.Automation, error)
	Create(ctx context.Context, automation *agentsv1.Automation) error
	Update(ctx context.Context, automation *agentsv1.Automation) error
	Delete(ctx context.Context, workspaceID, name string) (*agentsv1.Automation, error)
}

// RunRepo persists automation run records.
type RunRepo interface {
	EnsureIndexes(ctx context.Context) error
	Save(ctx context.Context, run *agentsv1.AutomationRun) error
	Get(ctx context.Context, workspaceID, id string) (*agentsv1.AutomationRun, error)
	List(ctx context.Context, workspaceID, automationName string, pageSize int32, pageToken string) ([]*agentsv1.AutomationRun, string, error)
	// ListWaitingBySession returns every WAITING_INPUT run recorded for the
	// given session coordinates, across workspaces. The resume path (a runner
	// turn or a session deletion) knows only the session coordinates — never a
	// workspace — so the lookup is workspace-agnostic. Automation sessions are
	// per-run (automation:<run-id>), so this normally returns at most one run.
	ListWaitingBySession(ctx context.Context, appName, userID, sessionID string) ([]*agentsv1.AutomationRun, error)
	// ListStaleRunning returns RUNNING runs started before the given time,
	// across workspaces. Reconciliation fails them: their process exited
	// without finalizing them, and nothing else ever will.
	ListStaleRunning(ctx context.Context, before time.Time) ([]*agentsv1.AutomationRun, error)
}

// Run listing paginates with a (started_at, id) cursor rather than an offset:
// deep offsets scan, and concurrent inserts shift offset pages. The token is
// opaque to clients; an unparseable token (including a legacy offset token)
// falls back to the first page.
const runPageTokenVersion = "v2"

// EncodeRunPageToken builds the cursor for the next page after the given run.
// started_at is millisecond-truncated to match BSON datetime precision, so
// the token compares identically against stored values.
func EncodeRunPageToken(startedAt time.Time, id string) string {
	raw := fmt.Sprintf("%s:%d:%s", runPageTokenVersion, startedAt.UnixMilli(), id)
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

// DecodeRunPageToken parses a cursor token. ok=false means first page.
func DecodeRunPageToken(token string) (startedAt time.Time, id string, ok bool) {
	if token == "" {
		return time.Time{}, "", false
	}
	raw, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return time.Time{}, "", false
	}
	parts := strings.SplitN(string(raw), ":", 3)
	if len(parts) != 3 || parts[0] != runPageTokenVersion {
		return time.Time{}, "", false
	}
	ms, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return time.Time{}, "", false
	}
	return time.UnixMilli(ms).UTC(), parts[2], true
}

// clampRunPageSize normalizes a requested page size into [1, 200].
func clampRunPageSize(pageSize int32) int32 {
	if pageSize <= 0 {
		return 20
	}
	if pageSize > 200 {
		return 200
	}
	return pageSize
}

// StepRunRepo persists step-level run records.
type StepRunRepo interface {
	EnsureIndexes(ctx context.Context) error
	Save(ctx context.Context, stepRun *agentsv1.AutomationStepRun) error
	Get(ctx context.Context, workspaceID, id string) (*agentsv1.AutomationStepRun, error)
	ListByRun(ctx context.Context, workspaceID, runID string) ([]*agentsv1.AutomationStepRun, error)
}
