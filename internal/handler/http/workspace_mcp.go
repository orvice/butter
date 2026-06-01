package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"go.orx.me/apps/butter/internal/repo/auth"
	wsctx "go.orx.me/apps/butter/internal/workspace"
)

// RegisterWorkspaceMCP mounts a workspace-scoped MCP endpoint.
func RegisterWorkspaceMCP(r *gin.Engine, handler http.Handler) {
	if handler == nil {
		return
	}
	r.Any("/api/workspaces/:workspace_id/mcp", func(c *gin.Context) {
		workspaceID := strings.TrimSpace(c.Param("workspace_id"))
		if workspaceID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id required"})
			return
		}
		current, hasWorkspace := wsctx.FromContext(c.Request.Context())
		if !hasWorkspace && !auth.IsAdmin(c.Request.Context()) {
			c.JSON(http.StatusForbidden, gin.H{"error": "workspace required"})
			return
		}
		if hasWorkspace && current != workspaceID {
			c.JSON(http.StatusForbidden, gin.H{"error": "workspace mismatch"})
			return
		}
		c.Request = c.Request.WithContext(wsctx.WithID(c.Request.Context(), workspaceID))
		handler.ServeHTTP(c.Writer, c.Request)
	})
}
