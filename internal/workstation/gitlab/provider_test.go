package gitlab

import (
	"context"
	"errors"
	ws "go.orx.me/apps/butter/internal/workstation"
	"io"
	"net/http"
	"strings"
	"testing"
)

type fixture struct {
	status       int
	body         string
	header       http.Header
	method, path string
}

func (f *fixture) RoundTrip(r *http.Request) (*http.Response, error) {
	f.method = r.Method
	f.path = r.URL.Path
	if r.URL.String() == "" {
		return nil, errors.New("bad request")
	}
	return &http.Response{StatusCode: f.status, Header: f.header, Body: io.NopCloser(strings.NewReader(f.body)), Request: r}, nil
}
func TestReadOnlyFixtureAndPagination(t *testing.T) {
	f := &fixture{status: 200, body: `[{"id":7,"name":"Core","path_with_namespace":"acme/core"}]`, header: http.Header{"X-Next-Page": {"2"}}}
	p := New(f)
	page, err := p.DiscoverScopes(context.Background(), "fixture", "")
	if err != nil || len(page.Scopes) != 1 || page.Next != "2" {
		t.Fatalf("page=%+v err=%v", page, err)
	}
	if f.method != http.MethodGet || f.path != "/api/v4/projects" {
		t.Fatalf("request=%s %s", f.method, f.path)
	}
}
func TestErrorClassification(t *testing.T) {
	for _, x := range []struct {
		code int
		want string
	}{{401, "AUTH_INVALID"}, {403, "ACCESS_DENIED"}, {429, "RATE_LIMITED"}} {
		t.Run(x.want, func(t *testing.T) {
			_, e := New(&fixture{status: x.code, body: `{}`}).Probe(context.Background(), "x")
			var f *ws.Failure
			if !errors.As(e, &f) || f.Category != x.want {
				t.Fatalf("err=%v", e)
			}
		})
	}
}
