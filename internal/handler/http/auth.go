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
	"go.orx.me/apps/butter/internal/repo/auth"
)

const (
	bearerPrefix        = "Bearer "
	defaultTouchTimeout = 2 * time.Second
)

type APITokenRepoProvider func() apitoken.Repository

type AuthRepoProvider func() auth.Repository

// AuthMiddleware validates incoming requests against MongoDB-backed user
// sessions. The legacy root API token and API-token repo are still accepted for
// integrations/migration, but dashboard login should use AuthService.Login.
func AuthMiddleware(cfg *config.AppConfig, authProvider AuthRepoProvider, apiTokenProvider APITokenRepoProvider) gin.HandlerFunc {
	rootToken := strings.TrimSpace(cfg.APIToken)

	return func(c *gin.Context) {
		if isPublicPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		var authRepo auth.Repository
		if authProvider != nil {
			authRepo = authProvider()
		}
		var apiTokenRepo apitoken.Repository
		if apiTokenProvider != nil {
			apiTokenRepo = apiTokenProvider()
		}

		// Legacy/dev behavior before bootstrap wires repositories.
		if rootToken == "" && authRepo == nil && apiTokenRepo == nil {
			c.Next()
			return
		}

		token, ok := bearerToken(c)
		if !ok {
			unauthorized(c)
			return
		}

		if authRepo != nil {
			session, user, err := authRepo.LookupSession(c.Request.Context(), hashSecret(token), time.Now().UTC())
			if err == nil {
				go touchSession(authRepo, session.ID)
				ctx := auth.WithAuthenticated(c.Request.Context(), user, session)
				c.Request = c.Request.WithContext(ctx)
				c.Next()
				return
			}
			if !errors.Is(err, auth.ErrSessionNotFound) && !errors.Is(err, auth.ErrUserNotFound) && !errors.Is(err, auth.ErrUserDisabled) {
				log.FromContext(c.Request.Context()).Warn("auth session lookup failed", "err", err)
			}
		}

		// Try root token (constant-time compare) for ops/daemon/API compatibility.
		if rootToken != "" && subtle.ConstantTimeCompare([]byte(token), []byte(rootToken)) == 1 {
			c.Next()
			return
		}

		if apiTokenRepo != nil {
			stored, err := apiTokenRepo.Lookup(c.Request.Context(), hashSecret(token))
			if err == nil {
				go touchAPIToken(apiTokenRepo, stored.GetId())
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

// APITokenAuthMiddleware is kept for tests/backward compatibility.
func APITokenAuthMiddleware(cfg *config.AppConfig, provider APITokenRepoProvider) gin.HandlerFunc {
	return AuthMiddleware(cfg, nil, provider)
}

func isPublicPath(path string) bool {
	return path == "/ping" || path == "/api/agents.v1.AuthService/Login"
}

func bearerToken(c *gin.Context) (string, bool) {
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix))
	return token, token != ""
}

func hashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

func touchAPIToken(repo apitoken.Repository, id string) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTouchTimeout)
	defer cancel()
	_ = repo.TouchLastUsed(ctx, id)
}

func touchSession(repo auth.Repository, id string) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTouchTimeout)
	defer cancel()
	_ = repo.TouchSession(ctx, id, time.Now().UTC())
}

func unauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
}
