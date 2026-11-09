package plane

import (
	"cmp"
	"context"
	"errors"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"
)

type FailureClass string

const (
	FailureRetryable FailureClass = "retryable"
	FailurePermanent FailureClass = "permanent"
	FailureCancelled FailureClass = "cancelled"
	FailureUnknown   FailureClass = "unknown"
)

type DiagnosticRecord struct {
	Key            string
	WorkspaceID    string
	AppID          string
	InstallationID string
	RunID          string
	ActivityID     string
	WorkflowID     string
	Class          FailureClass
	Summary        string
	Recoverable    bool
	CreatedAt      time.Time
	// ExpiresAt is set only for terminal webhook receipts; older execution
	// diagnostics keep their existing retention semantics.
	ExpiresAt time.Time
}

const DiagnosticHistoryLimit = 200

type DiagnosticStore interface {
	Put(context.Context, *DiagnosticRecord) error
	Get(context.Context, string, string) (*DiagnosticRecord, error)
	List(context.Context, string, string) ([]*DiagnosticRecord, error)
}
type MemoryDiagnosticStore struct {
	mu   sync.Mutex
	rows map[string]*DiagnosticRecord
}

func NewMemoryDiagnosticStore() *MemoryDiagnosticStore {
	return &MemoryDiagnosticStore{rows: map[string]*DiagnosticRecord{}}
}
func (s *MemoryDiagnosticStore) Put(_ context.Context, r *DiagnosticRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	c := *r
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now().UTC()
	}
	s.rows[r.Key] = &c
	return nil
}
func (s *MemoryDiagnosticStore) Get(_ context.Context, ws, k string) (*DiagnosticRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.rows[k]
	if !ok || r.WorkspaceID != ws {
		return nil, errors.New("diagnostic record not found")
	}
	c := *r
	return &c, nil
}
func (s *MemoryDiagnosticStore) List(_ context.Context, ws, app string) ([]*DiagnosticRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []*DiagnosticRecord{}
	for _, r := range s.rows {
		if r.WorkspaceID == ws && (app == "" || r.AppID == app) {
			c := *r
			out = append(out, &c)
		}
	}
	slices.SortFunc(out, func(a, b *DiagnosticRecord) int {
		return cmp.Or(b.CreatedAt.Compare(a.CreatedAt), strings.Compare(a.Key, b.Key))
	})
	return out[:min(len(out), DiagnosticHistoryLimit)], nil
}
func ClassifyFailure(err error) FailureClass {
	if err == nil {
		return FailureUnknown
	}
	if errors.Is(err, context.Canceled) {
		return FailureCancelled
	}
	if r, ok := err.(RetryableError); ok && r.Retryable() {
		return FailureRetryable
	}
	return FailurePermanent
}

var secretPattern = regexp.MustCompile(`(?i)(token|secret|authorization|signature|password|raw_payload|access_token|client_secret)[=:][^\s,;]+`)

func SafeDiagnosticSummary(err error) string {
	if err == nil {
		return ""
	}
	s := secretPattern.ReplaceAllString(err.Error(), "$1=[redacted]")
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > 256 {
		s = s[:256]
	}
	return s
}
func BuildDiagnostic(e *Event, workflowID string, err error) *DiagnosticRecord {
	class := ClassifyFailure(err)
	return &DiagnosticRecord{Key: e.InstallationID + ":" + e.AgentRunID + ":" + e.ActivityID + ":" + workflowID, WorkspaceID: e.WorkspaceID, AppID: e.AppID, InstallationID: e.InstallationID, RunID: e.AgentRunID, ActivityID: e.ActivityID, WorkflowID: workflowID, Class: class, Summary: SafeDiagnosticSummary(err), Recoverable: class == FailureRetryable, CreatedAt: time.Now().UTC()}
}
