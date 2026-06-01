package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"go.orx.me/apps/butter/internal/authn"
	"go.orx.me/apps/butter/internal/config"
)

// WorkspaceRepoProvider is re-exported for callers that wire the middleware.
type WorkspaceRepoProvider = authn.WorkspaceRepoProvider

// APITokenRepoProvider is re-exported for callers that wire the middleware.
type APITokenRepoProvider = authn.APITokenRepoProvider

// AuthRepoProvider is re-exported for callers that wire the middleware.
type AuthRepoProvider = authn.AuthRepoProvider

// AuthMiddleware validates incoming Gin requests using the shared
// authn.Resolver. Public-path bypass is HTTP-specific; the gRPC transport
// applies its own method-level bypass list at the interceptor.
//
// Callers that already hold a resolver should use AuthMiddlewareWithResolver
// to avoid constructing duplicate instances and warning logs.
func AuthMiddleware(cfg *config.AppConfig, authProvider AuthRepoProvider, apiTokenProvider APITokenRepoProvider, workspaceProvider WorkspaceRepoProvider) gin.HandlerFunc {
	return AuthMiddlewareWithResolver(authn.New(cfg, authProvider, apiTokenProvider, workspaceProvider))
}

// AuthMiddlewareWithResolver wires an already-constructed resolver into a
// Gin middleware. This is the form used by the production routes so that
// the gRPC interceptor and the Gin middleware share state (one fallback
// warning, one set of providers).
func AuthMiddlewareWithResolver(resolver *authn.Resolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		if isPublicPath(c.Request.URL.Path) {
			c.Next()
			return
		}
		res := resolver.Resolve(c.Request.Context(), ginHeaderSource{c})
		if res.Outcome != authn.OutcomeAuthenticated {
			unauthorized(c)
			return
		}
		c.Request = c.Request.WithContext(res.Ctx)
		c.Next()
	}
}

// APITokenAuthMiddleware is kept for tests/backward compatibility.
func APITokenAuthMiddleware(cfg *config.AppConfig, provider APITokenRepoProvider) gin.HandlerFunc {
	return AuthMiddleware(cfg, nil, provider, nil)
}

// ginHeaderSource adapts a *gin.Context to authn.HeaderSource.
type ginHeaderSource struct{ c *gin.Context }

func (g ginHeaderSource) Get(name string) string { return g.c.GetHeader(name) }

func isPublicPath(path string) bool {
	switch path {
	case "/ping",
		"/api/agents.v1.AuthService/Login",
		"/api/agents.v1.AuthService/ListOAuthProviders",
		"/api/agents.v1.AuthService/BeginOAuthFlow",
		"/api/agents.v1.AuthService/CompleteOAuthFlow",
		"/api/mcp/oauth/callback":
		return true
	}
	return false
}

func unauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
}
