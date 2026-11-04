package workstation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"time"
)

// BudgetRepository is shared by all workspaces and Pods. Reservation uses
// database time; increasing connections/keys cannot multiply account throughput.
type BudgetRepository interface {
	ReserveBudget(context.Context, string, time.Duration) (time.Duration, error)
	DeferBudget(context.Context, string, time.Duration) error
}

func BudgetKeys(r Record) (string, string) {
	source := r.SourceIdentity()
	hash := func(raw string) string { sum := sha256.Sum256([]byte(raw)); return hex.EncodeToString(sum[:]) }
	instance := source.Provider + "\x00" + source.Origin
	return hash("instance\x00" + instance), hash("account\x00" + instance + "\x00" + r.PrincipalID)
}

type budgetTransport struct {
	repo              BudgetRepository
	next              http.RoundTripper
	instance, account string
}

func BudgetTransport(repo BudgetRepository, r Record, next http.RoundTripper) http.RoundTripper {
	i, a := BudgetKeys(r)
	return &budgetTransport{repo: repo, next: next, instance: i, account: a}
}
func (t *budgetTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for _, b := range []struct {
		key     string
		spacing time.Duration
	}{{t.instance, 100 * time.Millisecond}, {t.account, 500 * time.Millisecond}} {
		for {
			wait, err := t.repo.ReserveBudget(req.Context(), b.key, b.spacing)
			if err != nil {
				return nil, &Failure{Category: "DEGRADED"}
			}
			if wait > time.Second {
				return nil, &Failure{Category: "RATE_LIMITED", RetryAfter: wait}
			}
			if wait <= 0 {
				break
			}
			timer := time.NewTimer(wait)
			select {
			case <-timer.C:
			case <-req.Context().Done():
				timer.Stop()
				return nil, req.Context().Err()
			}
		}
	}
	return t.next.RoundTrip(req)
}

// RequestFailure preserves only the local coordinator's safe retry category.
// Raw transport errors, URLs, headers and credential strings never escape.
func RequestFailure(err error) error {
	var f *Failure
	if errors.As(err, &f) && f.Category == "RATE_LIMITED" {
		return &Failure{Category: f.Category, RetryAfter: f.RetryAfter}
	}
	return &Failure{Category: "DEGRADED"}
}
