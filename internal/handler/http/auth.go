package http

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"butterfly.orx.me/core/log"
	"github.com/gin-gonic/gin"

	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/repo/apitoken"
)

const (
	bearerPrefix        = "Bearer "
	defaultTouchTimeout = 2 * time.Second
)

// APITokenRepoProvider resolves the current apitoken repository. It is invoked
// per-request so the repository can be wired in after route setup.
type APITokenRepoProvider func() apitoken.Repository

// APITokenAuthMiddleware validates incoming requests against either:
//   - The single root token from AppConfig.APIToken (preserved for ops/CLI).
//   - A token stored in the apitoken.Repository (added at runtime via
//     APITokenService). When provider is nil or returns nil, only the root
//     token is checked.
//
// The /ping endpoint is always public.
func APITokenAuthMiddleware(cfg *config.AppConfig, provider APITokenRepoProvider) gin.HandlerFunc {
	rootToken := strings.TrimSpace(cfg.APIToken)

	return func(c *gin.Context) {
		if c.Request.URL.Path == "/ping" {
			c.Next()
			return
		}

		var repo apitoken.Repository
		if provider != nil {
			repo = provider()
		}

		// If no auth is configured at all, allow through (legacy behavior).
		if rootToken == "" && repo == nil {
			c.Next()
			return
		}

		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, bearerPrefix) {
			unauthorized(c)
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(auth, bearerPrefix))
		if token == "" {
			unauthorized(c)
			return
		}

		// Try root token first (constant-time compare).
		if rootToken != "" && subtle.ConstantTimeCompare([]byte(token), []byte(rootToken)) == 1 {
			c.Next()
			return
		}

		if repo != nil {
			hash := hashSecret(token)
			stored, err := repo.Lookup(c.Request.Context(), hash)
			if err == nil {
				go touchToken(repo, stored.GetId())
				c.Next()
				return
			}
			if !errors.Is(err, apitoken.ErrNotFound) {
				log.FromContext(c.Request.Context()).Warn("api token lookup failed", "err", err)
			}
		}

		unauthorized(c)
	}
}

func hashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

func touchToken(repo apitoken.Repository, id string) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTouchTimeout)
	defer cancel()
	_ = repo.TouchLastUsed(ctx, id)
}

func unauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
}
