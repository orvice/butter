// Package acceptance_test provides opt-in, GET/query-only live qualification.
package acceptance_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	ws "go.orx.me/apps/butter/internal/workstation"
	"go.orx.me/apps/butter/internal/workstation/github"
	"go.orx.me/apps/butter/internal/workstation/gitlab"
	"go.orx.me/apps/butter/internal/workstation/linear"
	"go.orx.me/apps/butter/internal/workstation/outbound"
	"go.orx.me/apps/butter/internal/workstation/plane"
)

type target struct {
	Name             string        `json:"name"`
	Provider         string        `json:"provider"`
	APIBase          string        `json:"api_base"`
	WebBase          string        `json:"web_base"`
	WorkspaceSlug    string        `json:"workspace_slug"`
	ScopeID          string        `json:"scope_id"`
	CredentialEnv    string        `json:"credential_env"`
	Version          string        `json:"version"`
	Profile          string        `json:"profile"`
	TenantID         string        `json:"tenant_id"`
	MinimumItems     *int          `json:"minimum_items"`
	ExpectedCoverage string        `json:"expected_coverage"`
	Expected         []expectation `json:"expected"`
}
type expectation struct {
	RemoteID             string `json:"remote_id"`
	Kind                 string `json:"kind"`
	Status               string `json:"status"`
	Absent               bool   `json:"absent"`
	Draft                *bool  `json:"draft"`
	DescriptionAvailable *bool  `json:"description_available"`
}
type readAudit struct {
	next          http.RoundTripper
	mu            sync.Mutex
	gets, queries int
}

func (a *readAudit) RoundTrip(req *http.Request) (*http.Response, error) {
	query := false
	if req.Method == http.MethodPost && req.URL.Host == "api.linear.app" && req.URL.Path == "/graphql" && req.GetBody != nil {
		body, err := req.GetBody()
		if err != nil {
			return nil, ws.ErrInvalid
		}
		var payload struct {
			Query string `json:"query"`
		}
		err = json.NewDecoder(io.LimitReader(body, 1<<20)).Decode(&payload)
		_ = body.Close()
		query = err == nil && strings.HasPrefix(strings.TrimSpace(payload.Query), "query ")
	}
	if req.Method != http.MethodGet && !query {
		return nil, ws.ErrInvalid
	}
	a.mu.Lock()
	if query {
		a.queries++
	} else {
		a.gets++
	}
	a.mu.Unlock()
	return a.next.RoundTrip(req)
}

func TestRealProviderProfiles(t *testing.T) {
	path := os.Getenv("BUTTER_WORK_STATION_ACCEPTANCE_MANIFEST")
	if path == "" {
		t.Skip("real platform acceptance not run: no explicitly approved target manifest")
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal("cannot read acceptance manifest")
	}
	var targets []target
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&targets) != nil || len(targets) == 0 {
		t.Fatal("invalid or empty acceptance manifest")
	}
	for _, target := range targets {
		t.Run(target.Name, func(t *testing.T) {
			if target.Name == "" || target.Version == "" || target.Profile == "" || target.ScopeID == "" || target.CredentialEnv == "" {
				t.Fatal("manifest lacks target/version/profile/scope/credential environment reference")
			}
			key := os.Getenv(target.CredentialEnv)
			if key == "" {
				t.Fatal("credential environment reference is not provisioned")
			}
			origin := target.APIBase
			switch target.Provider {
			case "linear":
				origin = "https://api.linear.app"
			case "github":
				origin = "https://api.github.com"
			}
			confined, err := outbound.NewClient(origin, nil)
			if err != nil {
				t.Fatal("manifest origin does not satisfy approved HTTPS policy")
			}
			audit := &readAudit{next: confined.Transport}
			var adapter ws.Provider
			switch target.Provider {
			case "linear":
				adapter = linear.New(audit)
			case "github":
				adapter = github.New(audit)
			case "gitlab":
				adapter, err = gitlab.New(target.APIBase, target.WebBase, audit)
			case "plane":
				adapter, err = plane.New(target.APIBase, target.WebBase, target.WorkspaceSlug, audit)
			default:
				t.Fatal("unknown provider")
			}
			if err != nil {
				t.Fatal("manifest target does not match the fixed provider profile")
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			identity, err := adapter.Probe(ctx, key)
			if err != nil {
				t.Fatalf("identity probe category=%s", ws.Classify(err).Category)
			}
			if identity.PrincipalID == "" || identity.Source.Provider != target.Provider {
				t.Fatal("identity probe returned incomplete identity")
			}
			if target.TenantID != "" && identity.Source.TenantID != target.TenantID {
				t.Fatal("external tenant identity did not match manifest")
			}
			scope, err := adapter.ResolveScope(ctx, key, target.ScopeID)
			if err != nil {
				t.Fatalf("scope check category=%s", ws.Classify(err).Category)
			}
			result, err := adapter.Scan(ctx, key, scope, time.Now().UTC())
			if err != nil {
				t.Fatalf("scan category=%s", ws.Classify(err).Category)
			}
			coverage := target.ExpectedCoverage
			if coverage == "" {
				coverage = "COMPLETE"
			}
			if result.Coverage != coverage {
				t.Fatalf("coverage=%s expected=%s unresolved=%d", result.Coverage, coverage, result.Unresolved)
			}
			minimum := 1
			if target.MinimumItems != nil {
				minimum = *target.MinimumItems
			}
			if minimum < 0 || len(result.Items) < minimum {
				t.Fatalf("controlled dataset missing: count=%d minimum=%d", len(result.Items), minimum)
			}
			for index, expected := range target.Expected {
				found := false
				for _, item := range result.Items {
					if item.RemoteId != expected.RemoteID || item.Kind != expected.Kind {
						continue
					}
					found = true
					if expected.Absent || (expected.Status != "" && item.Status != expected.Status) || (expected.Draft != nil && item.Draft != *expected.Draft) || (expected.DescriptionAvailable != nil && item.DescriptionAvailable != *expected.DescriptionAvailable) {
						t.Fatalf("controlled item expectation %d failed", index)
					}
				}
				if !found && !expected.Absent {
					t.Fatalf("controlled item expectation %d not found", index)
				}
			}
			audit.mu.Lock()
			gets, queries := audit.gets, audit.queries
			audit.mu.Unlock()
			t.Logf("profile=%q version_recorded=true readonly_gets=%d readonly_queries=%d items=%d coverage=%s unresolved=%d", target.Profile, gets, queries, len(result.Items), result.Coverage, result.Unresolved)
		})
	}
}
