package linear

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	ws "go.orx.me/apps/butter/internal/workstation"
)

type fixtureTransport struct {
	status int
	body   string
	head   http.Header
	query  string
}

func (t *fixtureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.String() != endpoint || req.Method != http.MethodPost {
		return nil, errors.New("unexpected request target")
	}
	data, _ := io.ReadAll(req.Body)
	t.query = string(data)
	return &http.Response{StatusCode: t.status, Header: t.head, Body: io.NopCloser(strings.NewReader(t.body)), Request: req}, nil
}

func TestReadOnlyProbeAndDiscoveryFixture(t *testing.T) {
	transport := &fixtureTransport{status: 200, body: `{"data":{"viewer":{"id":"user-1"},"teams":{"nodes":[{"id":"team-1","name":"Core","key":"COR"}],"pageInfo":{"hasNextPage":true,"endCursor":"cursor-1"}}}}`}
	p := New(transport)
	identity, err := p.Probe(context.Background(), "Bearer fixture")
	if err != nil || identity.Source.Provider != "linear" || identity.Source.TenantID != "user-1" {
		t.Fatalf("probe = %+v, %v", identity, err)
	}
	page, err := p.DiscoverScopes(context.Background(), "Bearer fixture", "")
	if err != nil || len(page.Scopes) != 1 || page.Next != "cursor-1" {
		t.Fatalf("page = %+v, %v", page, err)
	}
	if strings.Contains(strings.ToLower(transport.query), "mutation") {
		t.Fatal("fixture sent a mutation query")
	}
}

func TestGraphQLErrorsAndRateLimitAreClassified(t *testing.T) {
	for _, test := range []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{"partial graphql", 200, `{"data":{"viewer":null},"errors":[{"message":"denied"}]}`, "SCHEMA_ERROR"},
		{"rate limit", 429, `{}`, "RATE_LIMITED"},
		{"unauthorized", 401, `{}`, "AUTH_INVALID"},
	} {
		t.Run(test.name, func(t *testing.T) {
			tr := &fixtureTransport{status: test.status, body: test.body, head: http.Header{"Retry-After": {"3"}}}
			_, err := New(tr).Probe(context.Background(), "fixture")
			var failure *ws.Failure
			if !errors.As(err, &failure) || failure.Category != test.want {
				t.Fatalf("error = %v, want %s", err, test.want)
			}
		})
	}
}
