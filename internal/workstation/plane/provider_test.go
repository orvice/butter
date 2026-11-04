package plane

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
	code         int
	body         string
	path, method string
	headers      http.Header
}

func (f *fixture) RoundTrip(r *http.Request) (*http.Response, error) {
	f.path = r.URL.Path
	f.method = r.Method
	return &http.Response{StatusCode: f.code, Header: f.headers, Body: io.NopCloser(strings.NewReader(f.body)), Request: r}, nil
}
func TestProfilesAndReadOnlySemantics(t *testing.T) {
	if _, e := NewSelfHosted("http://plane.local", nil); !errors.Is(e, ws.ErrInvalid) {
		t.Fatal("insecure self-host must be rejected")
	}
	f := &fixture{code: 200, body: `{"results":[{"id":"p1","name":"Core","archived":false}],"next":"https://plane.example/api/projects/?page=2"}`}
	p, e := NewSelfHosted("https://plane.example", f)
	if e != nil {
		t.Fatal(e)
	}
	page, e := p.DiscoverScopes(context.Background(), "token", "")
	if e != nil || len(page.Scopes) != 1 || page.Next == "" {
		t.Fatalf("page=%+v err=%v", page, e)
	}
	if f.method != http.MethodGet {
		t.Fatalf("method=%s", f.method)
	}
}
func TestVersionAndRateErrors(t *testing.T) {
	for _, x := range []struct {
		code int
		want string
	}{{409, "API_VERSION_UNSUPPORTED"}, {401, "AUTH_INVALID"}, {429, "RATE_LIMITED"}} {
		t.Run(x.want, func(t *testing.T) {
			f := &fixture{code: x.code, body: `{}`}
			p, _ := New(f)
			_, e := p.Probe(context.Background(), "x")
			var got *ws.Failure
			if !errors.As(e, &got) || got.Category != x.want {
				t.Fatalf("err=%v", e)
			}
		})
	}
}
