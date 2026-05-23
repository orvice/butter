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
	"go.orx.me/apps/butter/internal/repo/workspace"
	wsctx "go.orx.me/apps/butter/internal/workspace"
)

// WorkspaceRepoProvider returns the active workspace repository, if wired.
type WorkspaceRepoProvider func() workspace.Repository

const (
	bearerPrefix        = "Bearer "
	defaultTouchTimeout = 2 * time.Second
)

type APITokenRepoProvider func() apitoken.Repository

type AuthRepoProvider func() auth.Repository

// AuthMiddleware validates incoming requests against MongoDB-backed user
// sessions. The legacy root API token and API-token repo are still accepted for
// integrations/migration, but dashboard login should use AuthService.Login.
//
// On success the middleware also resolves the workspace requested via
// X-Workspace-ID (or falls back to the API token's own workspace) and
// validates that the caller is a member. The resolved workspace id is then
// attached to the request context so Twirp services can scope their reads
// and writes.
func AuthMiddleware(cfg *config.AppConfig, authProvider AuthRepoProvider, apiTokenProvider APITokenRepoProvider, workspaceProvider WorkspaceRepoProvider) gin.HandlerFunc {
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
		var workspaceRepo workspace.Repository
		if workspaceProvider != nil {
			workspaceRepo = workspaceProvider()
		}

		// Legacy/dev behavior before bootstrap wires repositories.
		if rootToken == "" && authRepo == nil && apiTokenRepo == nil {
			ctx := auth.WithAdmin(c.Request.Context())
			c.Request = c.Request.WithContext(applyWorkspaceHeader(ctx, c, workspaceRepo, "", true))
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
				if user.GetRole() == "admin" {
					ctx = auth.WithAdmin(ctx)
				}
				ctx = applyWorkspaceHeader(ctx, c, workspaceRepo, user.GetId(), user.GetRole() == "admin")
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
			ctx := auth.WithAdmin(c.Request.Context())
			c.Request = c.Request.WithContext(applyWorkspaceHeader(ctx, c, workspaceRepo, "", true))
			c.Next()
			return
		}

		if apiTokenRepo != nil {
			stored, err := apiTokenRepo.Lookup(c.Request.Context(), hashSecret(token))
			if err == nil {
				// API tokens are workspace-scoped by design. A stored token
				// without a workspace_id signals data corruption or a
				// creation-time bug; previously such a token would silently
				// fall back to applyWorkspaceHeader with admin=true and let
				// the caller pick any workspace via the header. Reject it
				// instead so the invariant is enforced defensively.
				if strings.TrimSpace(stored.GetWorkspaceId()) == "" {
					log.FromContext(c.Request.Context()).Warn(
						"api token has no workspace binding; rejecting",
						"token_id", stored.GetId(),
					)
					unauthorized(c)
					return
				}
				go touchAPIToken(apiTokenRepo, stored.GetId())
				ctx := wsctx.WithID(c.Request.Context(), stored.GetWorkspaceId())
				c.Request = c.Request.WithContext(ctx)
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

// applyWorkspaceHeader resolves the X-Workspace-ID header (if set) and
// validates that the authenticated user is a member of that workspace.
// Returns a context with the workspace id attached, or the original context
// when no header is present (the workspace remains implicit until a
// downstream service requires it).
func applyWorkspaceHeader(ctx context.Context, c *gin.Context, repo workspace.Repository, userID string, isAdmin bool) context.Context {
	header := strings.TrimSpace(c.GetHeader(wsctx.HeaderName))
	if header == "" {
		return ctx
	}
	if repo == nil {
		// No repo wired yet: accept the header verbatim. This keeps
		// development/test setups working before bootstrap completes.
		return wsctx.WithID(ctx, header)
	}
	ws, err := repo.GetWorkspace(ctx, header)
	if err != nil {
		log.FromContext(ctx).Warn("workspace header rejected", "workspace_id", header, "err", err)
		return ctx
	}
	if !isAdmin && userID != "" {
		member, err := repo.IsMember(ctx, ws.GetId(), userID)
		if err != nil {
			log.FromContext(ctx).Warn("workspace membership check failed", "workspace_id", ws.GetId(), "user_id", userID, "err", err)
			return ctx
		}
		if !member {
			log.FromContext(ctx).Warn("workspace access denied", "workspace_id", ws.GetId(), "user_id", userID)
			return ctx
		}
	}
	return wsctx.WithID(ctx, ws.GetId())
}

// APITokenAuthMiddleware is kept for tests/backward compatibility.
func APITokenAuthMiddleware(cfg *config.AppConfig, provider APITokenRepoProvider) gin.HandlerFunc {
	return AuthMiddleware(cfg, nil, provider, nil)
}

func isPublicPath(path string) bool {
	return path == "/ping" || path == "/api/agents.v1.AuthService/Login" || path == "/api/mcp/oauth/callback"
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
