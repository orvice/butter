package http

// The MCP OAuth callback at its public HTTP boundary. The browser arrives from
// the authorization server with no Butter session; the callback finishes the
// exchange on the server and only then redirects. A native client therefore
// never needs to be the redirect target (mobile #454): it reads the server's
// OAuth status instead, so these tests pin that the exchange is independent of
// where the redirect points and that a failure or a provider error never
// reports success.

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"

	"go.orx.me/apps/butter/internal/mcpoauth"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type fakeMCPOAuthCompleter struct {
	calls     int
	state     string
	code      string
	returnURL string
	status    *agentsv1.MCPOAuthConnectionStatus
	err       error
}

func (f *fakeMCPOAuthCompleter) CompleteMCPServerOAuthCallback(_ context.Context, state, code string) (string, *agentsv1.MCPOAuthConnectionStatus, error) {
	f.calls++
	f.state, f.code = state, code
	return f.returnURL, f.status, f.err
}

func callMCPOAuthCallback(t *testing.T, completer MCPOAuthCompleter, dashboard, query string) (*url.URL, int) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	NewMCPOAuthCallbackHandler(completer, func() string { return dashboard }).Register(r)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, mcpoauth.CallbackPath+"?"+query, nil))
	loc, err := url.Parse(rec.Header().Get("Location"))
	if err != nil {
		t.Fatalf("bad Location %q: %v", rec.Header().Get("Location"), err)
	}
	return loc, rec.Code
}

func TestMCPOAuthCallbackReportsSuccessWithServerID(t *testing.T) {
	completer := &fakeMCPOAuthCompleter{
		returnURL: "https://dash.example/mcp-servers?tab=1",
		status:    &agentsv1.MCPOAuthConnectionStatus{ServerId: "srv-1"},
	}
	loc, code := callMCPOAuthCallback(t, completer, "https://dash.example", "state=s1&code=c1")
	if code != http.StatusFound {
		t.Fatalf("status = %d", code)
	}
	if completer.calls != 1 || completer.state != "s1" || completer.code != "c1" {
		t.Fatalf("completer = %+v", completer)
	}
	q := loc.Query()
	if loc.Host != "dash.example" || q.Get("mcp_oauth") != "success" || q.Get("server_id") != "srv-1" || q.Get("tab") != "1" {
		t.Fatalf("Location = %s", loc)
	}
}

func TestMCPOAuthCallbackWithoutReturnURLUsesTheDashboardFallback(t *testing.T) {
	completer := &fakeMCPOAuthCompleter{status: &agentsv1.MCPOAuthConnectionStatus{ServerId: "srv-1"}}
	loc, _ := callMCPOAuthCallback(t, completer, "https://dash.example/", "state=s&code=c")
	if loc.String() != "https://dash.example/mcp-servers?mcp_oauth=success&server_id=srv-1" {
		t.Fatalf("Location = %s", loc)
	}
	loc, _ = callMCPOAuthCallback(t, completer, "", "state=s&code=c")
	if loc.Path != "/mcp-servers" || loc.Host != "" || loc.Query().Get("mcp_oauth") != "success" {
		t.Fatalf("relative fallback Location = %s", loc)
	}
}

func TestMCPOAuthCallbackFailedExchangeNeverReportsSuccess(t *testing.T) {
	completer := &fakeMCPOAuthCompleter{err: errors.New("invalid or expired oauth state")}
	loc, code := callMCPOAuthCallback(t, completer, "", "state=stale&code=c")
	if code != http.StatusFound || loc.Query().Get("mcp_oauth") != "error" || loc.Query().Get("server_id") != "" {
		t.Fatalf("Location = %s (%d)", loc, code)
	}
}

func TestMCPOAuthCallbackProviderErrorSkipsTheExchange(t *testing.T) {
	completer := &fakeMCPOAuthCompleter{status: &agentsv1.MCPOAuthConnectionStatus{ServerId: "srv-1"}}
	loc, _ := callMCPOAuthCallback(t, completer, "", "error=access_denied&state=s&code=c")
	if completer.calls != 0 {
		t.Fatal("a provider error must not consume the flow or exchange a code")
	}
	if loc.Query().Get("mcp_oauth") != "error" {
		t.Fatalf("Location = %s", loc)
	}
}
