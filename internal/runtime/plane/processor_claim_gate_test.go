package plane

import (
	"context"
	"errors"
	planeprocessing "go.orx.me/apps/butter/internal/repo/planeprocessing"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/genai"
	"strings"
	"testing"
	"time"
)

type completionConflictClaims struct{}

func (completionConflictClaims) Claim(context.Context, string, string, string, string, string, time.Time) (planeprocessing.Claim, error) {
	return planeprocessing.Claim{Owner: "o", Revision: 1, Fence: 1, ExpiresAt: time.Now().Add(time.Minute)}, nil
}
func (completionConflictClaims) Renew(context.Context, string, planeprocessing.Claim, time.Time) (planeprocessing.Claim, error) {
	return planeprocessing.Claim{}, planeprocessing.ErrClaimExpired
}
func (completionConflictClaims) Complete(context.Context, string, planeprocessing.Claim) error {
	return planeprocessing.ErrClaimExpired
}
func (completionConflictClaims) ReclaimExpired(context.Context, string, string, string, string, string, time.Time) (planeprocessing.Claim, error) {
	return planeprocessing.Claim{}, planeprocessing.ErrClaimConflict
}
func (completionConflictClaims) RecoverExpired(context.Context, time.Time) error { return nil }

type spyRunner struct{}

func (spyRunner) Run(context.Context, string, []*genai.Part, string, *agentsv1.ContextInfo, runner.EventCallback, runner.CompactionCallback) (string, error) {
	return "token=secret payload", nil
}

type spyPublisher struct{ n int }

func (s *spyPublisher) PublishReply(context.Context, *Event, string) error { s.n++; return nil }

type spyProgress struct{ n int }

func (s *spyProgress) EmitProgress(context.Context, *Event, ProgressEvent) error { s.n++; return nil }

type spyDiagnostics struct{ n int }

func (s *spyDiagnostics) Put(context.Context, *DiagnosticRecord) error { s.n++; return nil }
func (s *spyDiagnostics) Get(context.Context, string, string) (*DiagnosticRecord, error) {
	return nil, errors.New("missing")
}
func (s *spyDiagnostics) List(context.Context, string, string) ([]*DiagnosticRecord, error) {
	return nil, nil
}
func TestProcessorDiscardsResultWhenClaimCompletionFails(t *testing.T) {
	pub := &spyPublisher{}
	prog := &spyProgress{}
	diag := &spyDiagnostics{}
	p := NewProcessor(spyRunner{}, pub)
	p.SetProgressSink(prog)
	p.SetDiagnosticStore(diag)
	p.SetClaimRepository(completionConflictClaims{}, "o")
	e := &Event{WorkspaceID: "w", AppID: "a", InstallationID: "i", AgentRunID: "r", ActivityID: "x", BotUserID: "u", Prompt: "token=secret"}
	_, err := p.Process(context.Background(), &agentsv1.PlaneApp{AgentId: "agent"}, e)
	if err == nil || pub.n != 0 || prog.n != 0 || diag.n != 0 {
		t.Fatalf("err=%v publisher=%d progress=%d diagnostics=%d", err, pub.n, prog.n, diag.n)
	}
	if strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), "payload") || strings.Contains(err.Error(), "token=") {
		t.Fatalf("unsafe error: %v", err)
	}
}
