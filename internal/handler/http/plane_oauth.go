package http

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
)

const PlaneOAuthCallbackPath = "/api/plane/oauth/callback"

type PlaneInstallCompleter interface {
	CompletePlaneInstall(context.Context, string, string, string) string
}
type PlaneOAuthHandler struct{ completer PlaneInstallCompleter }

func NewPlaneOAuthHandler(c PlaneInstallCompleter) *PlaneOAuthHandler {
	return &PlaneOAuthHandler{completer: c}
}
func (h *PlaneOAuthHandler) Register(r *gin.Engine) { r.GET(PlaneOAuthCallbackPath, h.Handle) }
func (h *PlaneOAuthHandler) Handle(c *gin.Context) {
	target := h.completer.CompletePlaneInstall(c.Request.Context(), c.Query("state"), c.Query("app_installation_id"), c.Query("error"))
	c.Redirect(http.StatusFound, target)
}
