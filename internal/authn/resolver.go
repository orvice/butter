// Package authn centralises request authentication and workspace resolution
// so HTTP/Gin (Twirp) and gRPC/grpc-web transports can share one
// implementation. Callers supply a HeaderSource and receive a context with
// auth.* and workspace ids attached.
package authn

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"

	"butterfly.orx.me/core/log"

	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/repo/apitoken"
	"go.orx.me/apps/butter/internal/repo/auth"
	"go.orx.me/apps/butter/internal/repo/workspace"
	wsctx "go.orx.me/apps/butter/internal/workspace"
)

const (
	bearerPrefix        = "Bearer "
	defaultTouchTimeout = 2 * time.Second

	// AuthorizationHeader and WorkspaceHeader are the canonical names the
	// resolver reads. HeaderSource implementations must accept these names
	// and adapt to their underlying convention (e.g. lowercase for gRPC
	// metadata, canonical MIME form for net/http).
	AuthorizationHeader = "Authorization"
	WorkspaceHeader     = wsctx.HeaderName
)

// HeaderSource returns a single header value by canonical name.
type HeaderSource interface {
	Get(name string) string
}

// HeaderSourceFunc adapts a plain func into a HeaderSource.
type HeaderSourceFunc func(name string) string

// Get implements HeaderSource.
func (f HeaderSourceFunc) Get(name string) string { return f(name) }

// AuthRepoProvider returns the user-session repository or nil.
type AuthRepoProvider func() auth.Repository

// APITokenRepoProvider returns the API-token repository or nil.
type APITokenRepoProvider func() apitoken.Repository

// WorkspaceRepoProvider returns the workspace repository or nil.
type WorkspaceRepoProvider func() workspace.Repository

// Outcome describes a successful authentication.
type Outcome int

const (
	// OutcomeAuthenticated indicates the request was accepted; the returned
	// context carries the resolved identity and workspace.
	OutcomeAuthenticated Outcome = iota
	// OutcomeRejected indicates the request must be denied (HTTP 401 or
	// gRPC Unauthenticated).
	OutcomeRejected
)

// Result is what Resolve returns.
type Result struct {
	Ctx     context.Context
	Outcome Outcome
}

// Resolver wires the configuration and repository providers needed to
// authenticate a request. Repository providers are read lazily on every
// call so they pick up post-bootstrap wiring.
type Resolver struct {
	cfg               *config.AppConfig
	authProvider      AuthRepoProvider
	apiTokenProvider  APITokenRepoProvider
	workspaceProvider WorkspaceRepoProvider

	fallbackWarn sync.Once
}

// New constructs a Resolver.
func New(cfg *config.AppConfig, authP AuthRepoProvider, apiTokenP APITokenRepoProvider, wsP WorkspaceRepoProvider) *Resolver {
	return &Resolver{
		cfg:               cfg,
		authProvider:      authP,
		apiTokenProvider:  apiTokenP,
		workspaceProvider: wsP,
	}
}

// Resolve attempts to authenticate the given headers and produce an
// authenticated context. Public-path bypass is handled by callers — the
// resolver always tries to authenticate.
func (r *Resolver) Resolve(ctx context.Context, h HeaderSource) Result {
	authRepo := r.lookupAuthRepo()
	apiTokenRepo := r.lookupAPITokenRepo()
	workspaceRepo := r.lookupWorkspaceRepo()
	rootToken := r.rootToken()

	// Dev/legacy bootstrap path: no auth wired at all.
	if rootToken == "" && authRepo == nil && apiTokenRepo == nil {
		if !r.allowUnauthenticated() {
			log.FromContext(ctx).Warn(
				"auth not configured and allow_unauthenticated=false; rejecting request",
			)
			return Result{Outcome: OutcomeRejected}
		}
		r.fallbackWarn.Do(func() {
			log.FromContext(ctx).Warn(
				"AUTH DISABLED: auth.allow_unauthenticated is true and no auth is wired; every request is granted admin. Do not run this configuration in production.",
			)
		})
		newCtx := auth.WithAdmin(ctx)
		newCtx = ApplyWorkspaceHeader(newCtx, h, workspaceRepo, "", true)
		return Result{Ctx: newCtx, Outcome: OutcomeAuthenticated}
	}

	token, ok := bearerToken(h)
	if !ok {
		return Result{Outcome: OutcomeRejected}
	}

	if authRepo != nil {
		session, user, err := authRepo.LookupSession(ctx, hashSecret(token), time.Now().UTC())
		if err == nil {
			go touchSession(authRepo, session.ID)
			newCtx := auth.WithAuthenticated(ctx, user, session)
			if user.GetRole() == "admin" {
				newCtx = auth.WithAdmin(newCtx)
			}
			newCtx = ApplyWorkspaceHeader(newCtx, h, workspaceRepo, user.GetId(), user.GetRole() == "admin")
			return Result{Ctx: newCtx, Outcome: OutcomeAuthenticated}
		}
		if !errors.Is(err, auth.ErrSessionNotFound) && !errors.Is(err, auth.ErrUserNotFound) && !errors.Is(err, auth.ErrUserDisabled) {
			log.FromContext(ctx).Warn("auth session lookup failed", "err", err)
		}
	}

	// Root token constant-time compare for ops/daemon/API compatibility.
	if rootToken != "" && subtle.ConstantTimeCompare([]byte(token), []byte(rootToken)) == 1 {
		newCtx := auth.WithAdmin(ctx)
		newCtx = ApplyWorkspaceHeader(newCtx, h, workspaceRepo, "", true)
		return Result{Ctx: newCtx, Outcome: OutcomeAuthenticated}
	}

	if apiTokenRepo != nil {
		stored, err := apiTokenRepo.Lookup(ctx, hashSecret(token))
		if err == nil {
			// API tokens are workspace-scoped by design. A stored token
			// without a workspace_id signals data corruption or a
			// creation-time bug; reject rather than fall back to letting
			// the caller pick any workspace via the header.
			if strings.TrimSpace(stored.GetWorkspaceId()) == "" {
				log.FromContext(ctx).Warn(
					"api token has no workspace binding; rejecting",
					"token_id", stored.GetId(),
				)
				return Result{Outcome: OutcomeRejected}
			}
			go touchAPIToken(apiTokenRepo, stored.GetId())
			newCtx := wsctx.WithID(ctx, stored.GetWorkspaceId())
			return Result{Ctx: newCtx, Outcome: OutcomeAuthenticated}
		}
		if !errors.Is(err, apitoken.ErrNotFound) {
			log.FromContext(ctx).Warn("api token lookup failed", "err", err)
		}
	}

	return Result{Outcome: OutcomeRejected}
}

func (r *Resolver) lookupAuthRepo() auth.Repository {
	if r.authProvider == nil {
		return nil
	}
	return r.authProvider()
}

func (r *Resolver) rootToken() string {
	if r.cfg == nil {
		return ""
	}
	return strings.TrimSpace(r.cfg.APIToken)
}

func (r *Resolver) allowUnauthenticated() bool {
	return r.cfg != nil && r.cfg.Auth.AllowUnauthenticated
}

func (r *Resolver) lookupAPITokenRepo() apitoken.Repository {
	if r.apiTokenProvider == nil {
		return nil
	}
	return r.apiTokenProvider()
}

func (r *Resolver) lookupWorkspaceRepo() workspace.Repository {
	if r.workspaceProvider == nil {
		return nil
	}
	return r.workspaceProvider()
}

// ApplyWorkspaceHeader resolves the workspace header and validates that the
// authenticated user is a member. Returns a context with the workspace id
// attached, or the original context when no header is present (the
// workspace remains implicit until a downstream service requires it).
func ApplyWorkspaceHeader(ctx context.Context, h HeaderSource, repo workspace.Repository, userID string, isAdmin bool) context.Context {
	header := strings.TrimSpace(h.Get(WorkspaceHeader))
	if header == "" {
		return ctx
	}
	if repo == nil {
		// No repo wired yet: accept the header verbatim. Keeps
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

func bearerToken(h HeaderSource) (string, bool) {
	v := h.Get(AuthorizationHeader)
	if !strings.HasPrefix(v, bearerPrefix) {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(v, bearerPrefix))
	return token, token != ""
}

// HashSecret returns the hex-encoded SHA-256 of the secret. Exported so
// tests and callers outside the resolver can compute lookup keys.
func HashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

func hashSecret(secret string) string { return HashSecret(secret) }

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
