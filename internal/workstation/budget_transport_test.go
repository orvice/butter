package workstation_test

import (
	"context"
	"errors"
	ws "go.orx.me/apps/butter/internal/workstation"
	"go.orx.me/apps/butter/internal/workstation/github"
	"go.orx.me/apps/butter/internal/workstation/memory"
	"net/http"
	"testing"
	"time"
)

type budgetFixtureTransport func(*http.Request) (*http.Response, error)

func (f budgetFixtureTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestSharedBackoffReachesProviderWithoutSendingOrWaitingForTimeout(t *testing.T) {
	repo := memory.New()
	record := ws.Record{Source: ws.GitHubSource(), PrincipalID: "account"}
	instance, _ := ws.BudgetKeys(record)
	if err := repo.DeferBudget(context.Background(), instance, 5*time.Minute); err != nil {
		t.Fatal(err)
	}
	called := false
	next := budgetFixtureTransport(func(*http.Request) (*http.Response, error) { called = true; return nil, errors.New("private secret") })
	p := github.New(ws.BudgetTransport(repo, record, next))
	_, err := p.Probe(context.Background(), "synthetic-key")
	var failure *ws.Failure
	if called || !errors.As(err, &failure) || failure.Category != "RATE_LIMITED" || failure.RetryAfter < 4*time.Minute {
		t.Fatal("shared retry lost", called, err)
	}
}
