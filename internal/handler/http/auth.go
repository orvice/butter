package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"go.orx.me/apps/butter/internal/config"
)

const bearerPrefix = "Bearer "

func APITokenAuthMiddleware(cfg *config.AppConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/ping" {
			c.Next()
			return
		}

		expected := strings.TrimSpace(cfg.APIToken)
		if expected == "" {
			c.Next()
			return
		}

		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, bearerPrefix) {
			unauthorized(c)
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(auth, bearerPrefix))
		if token == "" || token != expected {
			unauthorized(c)
			return
		}

		c.Next()
	}
}

func unauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
}
