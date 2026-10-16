package http

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"go.orx.me/apps/butter/internal/mcpoauth"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// MCPOAuthCompleter finishes an MCP OAuth browser callback: it consumes the
// single-use state, exchanges the code and stores the connection, and returns
// where the flow asked the browser to go next.
type MCPOAuthCompleter interface {
	CompleteMCPServerOAuthCallback(ctx context.Context, state, code string) (string, *agentsv1.MCPOAuthConnectionStatus, error)
}

// MCPOAuthCallbackHandler serves the public MCP OAuth callback
// (`mcpoauth.CallbackPath`). The browser arrives from the authorization
// server without a Butter session or workspace header, so the route is public
// and authenticated only by the single-use state. The code exchange happens
// here, on the server, whatever client started the flow; the redirect that
// follows is only a courtesy to whoever is looking at the browser, so a
// client that did not open a dashboard page learns the outcome by reading the
// server's OAuth status, never from this redirect.
type MCPOAuthCallbackHandler struct {
	completer MCPOAuthCompleter
	// dashboardBaseURL reads `mcp_oauth.dashboard_base_url` on every request,
	// as the config may be reloaded. The page it names is where the browser
	// goes when the flow named no return URL (or the provider reported an
	// error before the exchange).
	dashboardBaseURL func() string
}

// NewMCPOAuthCallbackHandler returns the handler. dashboardBaseURL yields
// `mcp_oauth.dashboard_base_url`; empty falls back to the relative
// `/mcp-servers` page.
func NewMCPOAuthCallbackHandler(completer MCPOAuthCompleter, dashboardBaseURL func() string) *MCPOAuthCallbackHandler {
	return &MCPOAuthCallbackHandler{completer: completer, dashboardBaseURL: dashboardBaseURL}
}

func (h *MCPOAuthCallbackHandler) Register(r *gin.Engine) {
	r.GET(mcpoauth.CallbackPath, h.Handle)
}

func (h *MCPOAuthCallbackHandler) Handle(c *gin.Context) {
	status := "error"
	serverID := ""
	target := mcpOAuthFallback(h.dashboardBaseURL())
	if oauthErr := strings.TrimSpace(c.Query("error")); oauthErr != "" {
		c.Redirect(http.StatusFound, appendMCPOAuthParams(target, status, serverID))
		return
	}
	returnURL, oauthStatus, err := h.completer.CompleteMCPServerOAuthCallback(c.Request.Context(), c.Query("state"), c.Query("code"))
	if returnURL != "" {
		target = returnURL
	}
	if err == nil {
		status = "success"
		if oauthStatus != nil {
			serverID = oauthStatus.GetServerId()
		}
	}
	c.Redirect(http.StatusFound, appendMCPOAuthParams(target, status, serverID))
}

func mcpOAuthFallback(dashboardBaseURL string) string {
	base := strings.TrimSpace(dashboardBaseURL)
	if base == "" {
		return "/mcp-servers"
	}
	return strings.TrimRight(base, "/") + "/mcp-servers"
}

func appendMCPOAuthParams(raw, status, serverID string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	q := u.Query()
	q.Set("mcp_oauth", status)
	if serverID != "" {
		q.Set("server_id", serverID)
	}
	u.RawQuery = q.Encode()
	return u.String()
}
