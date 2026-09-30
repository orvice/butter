package http

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// LinearOAuthCallbackPath is the public OAuth callback every Linear App
// shares; the single-use state names the App (ADR-0015).
const LinearOAuthCallbackPath = "/api/linear/oauth/callback"

// LinearInstallCompleter finishes a Linear App install and returns where to
// send the browser.
type LinearInstallCompleter interface {
	CompleteLinearInstall(ctx context.Context, state, code, linearError string) string
}

// LinearOAuthHandler serves the Linear OAuth callback. The browser arrives
// from Linear without a Butter session or workspace header, so the route is
// public and authenticated only by the single-use state.
type LinearOAuthHandler struct {
	completer LinearInstallCompleter
}

func NewLinearOAuthHandler(completer LinearInstallCompleter) *LinearOAuthHandler {
	return &LinearOAuthHandler{completer: completer}
}

func (h *LinearOAuthHandler) Register(r *gin.Engine) {
	r.GET(LinearOAuthCallbackPath, h.Handle)
}

func (h *LinearOAuthHandler) Handle(c *gin.Context) {
	target := h.completer.CompleteLinearInstall(c.Request.Context(),
		c.Query("state"), c.Query("code"), c.Query("error"))
	c.Redirect(http.StatusFound, target)
}
