package http

import (
	"net/http"
	"strings"

	"butterfly.orx.me/core/log"
	"github.com/gin-gonic/gin"

	"go.orx.me/apps/butter/internal/repo/auth"
	wsctx "go.orx.me/apps/butter/internal/workspace"
)

// RegisterWorkspaceMCP mounts a workspace-scoped MCP endpoint.
func RegisterWorkspaceMCP(r *gin.Engine, handler http.Handler, workspaceProvider WorkspaceRepoProvider) {
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
			user, hasUser := auth.UserFromContext(c.Request.Context())
			if !hasUser {
				c.JSON(http.StatusForbidden, gin.H{"error": "workspace required"})
				return
			}
			repo := workspaceProvider()
			if repo == nil {
				c.JSON(http.StatusForbidden, gin.H{"error": "workspace store not available"})
				return
			}
			ok, err := repo.IsMember(c.Request.Context(), workspaceID, user.GetId())
			if err != nil {
				log.FromContext(c.Request.Context()).Warn("workspace mcp membership check failed", "workspace_id", workspaceID, "user_id", user.GetId(), "err", err)
				c.JSON(http.StatusForbidden, gin.H{"error": "workspace access denied"})
				return
			}
			if !ok {
				c.JSON(http.StatusForbidden, gin.H{"error": "workspace access denied"})
				return
			}
		}
		if hasWorkspace && current != workspaceID {
			c.JSON(http.StatusForbidden, gin.H{"error": "workspace mismatch"})
			return
		}
		c.Request = c.Request.WithContext(wsctx.WithID(c.Request.Context(), workspaceID))
		handler.ServeHTTP(c.Writer, c.Request)
	})
}
