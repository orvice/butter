package application

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/auth/provider"
	"go.orx.me/apps/butter/internal/repo/auth"
	"go.orx.me/apps/butter/internal/repo/oauthstate"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// oauthStateTTL bounds how long an unredeemed BeginOAuthFlow state token is
// valid. Provider authorization flows are interactive; 10 minutes is plenty
// for the user to complete the consent screen and short enough that an
// abandoned flow cannot be replayed.
const oauthStateTTL = 10 * time.Minute

// SetProviderRegistry wires the OAuth provider registry. Calling it with nil
// disables the OAuth endpoints (they return FailedPrecondition).
func (s *AuthServiceServer) SetProviderRegistry(r *provider.Registry) {
	s.providers = r
}

// SetOAuthStateRepo wires the OAuth state store. Required for BeginOAuthFlow
// / CompleteOAuthFlow to function.
func (s *AuthServiceServer) SetOAuthStateRepo(r oauthstate.Repository) {
	s.stateRepo = r
}

func (s *AuthServiceServer) ListOAuthProviders(_ context.Context, _ *connect.Request[agentsv1.ListOAuthProvidersRequest]) (*connect.Response[agentsv1.ListOAuthProvidersResponse], error) {
	if s.providers == nil {
		return connect.NewResponse(&agentsv1.ListOAuthProvidersResponse{}), nil
	}
	list := s.providers.List()
	out := make([]*agentsv1.OAuthProvider, 0, len(list))
	for _, p := range list {
		out = append(out, &agentsv1.OAuthProvider{
			Name:        p.Name(),
			DisplayName: p.DisplayName(),
		})
	}
	return connect.NewResponse(&agentsv1.ListOAuthProvidersResponse{Providers: out}), nil
}

func (s *AuthServiceServer) BeginOAuthFlow(ctx context.Context, req *connect.Request[agentsv1.BeginOAuthFlowRequest]) (*connect.Response[agentsv1.BeginOAuthFlowResponse], error) {
	if s.providers == nil || s.stateRepo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("oauth login not available"))
	}
	name := strings.TrimSpace(req.Msg.GetProvider())
	if name == "" {
		return nil, connectx.RequiredArgument("provider")
	}
	p, err := s.providers.Get(name)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("unknown oauth provider"))
	}
	state, err := generateOAuthState()
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	now := time.Now().UTC()
	entry := &oauthstate.Entry{
		State:       state,
		Provider:    name,
		RedirectURI: strings.TrimSpace(req.Msg.GetRedirectUri()),
		CreatedAt:   now,
		ExpiresAt:   now.Add(oauthStateTTL),
	}
	if err := s.stateRepo.Create(ctx, entry); err != nil {
		return nil, connectx.InternalWith(err)
	}
	return connect.NewResponse(&agentsv1.BeginOAuthFlowResponse{
		AuthorizeUrl: p.AuthorizeURL(state),
		State:        state,
	}), nil
}

func (s *AuthServiceServer) CompleteOAuthFlow(ctx context.Context, req *connect.Request[agentsv1.CompleteOAuthFlowRequest]) (*connect.Response[agentsv1.CompleteOAuthFlowResponse], error) {
	logger := log.FromContext(ctx)
	if s.repo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("auth store not available"))
	}
	if s.providers == nil || s.stateRepo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("oauth login not available"))
	}
	name := strings.TrimSpace(req.Msg.GetProvider())
	if name == "" {
		return nil, connectx.RequiredArgument("provider")
	}
	code := strings.TrimSpace(req.Msg.GetCode())
	state := strings.TrimSpace(req.Msg.GetState())
	if code == "" {
		return nil, connectx.RequiredArgument("code")
	}
	if state == "" {
		return nil, connectx.RequiredArgument("state")
	}

	entry, err := s.stateRepo.Consume(ctx, state, time.Now().UTC())
	if err != nil {
		if errors.Is(err, oauthstate.ErrNotFound) {
			return nil, connect.NewError(connect.CodePermissionDenied, errors.New("invalid or expired state"))
		}
		return nil, connectx.InternalWith(err)
	}
	if entry.Provider != name {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("state provider mismatch"))
	}

	p, err := s.providers.Get(name)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("unknown oauth provider"))
	}

	claims, err := p.Exchange(ctx, code)
	if err != nil {
		logger.Error("oauth exchange failed", "provider", name, "err", err)
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("oauth code exchange failed"))
	}
	if claims == nil || claims.ExternalID == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("oauth provider returned no identity"))
	}

	user, err := s.upsertOAuthUser(ctx, name, claims)
	if err != nil {
		return nil, err
	}
	if user.GetDisabled() {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("user disabled"))
	}

	secret, err := generateSessionSecret()
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	now := time.Now().UTC()
	expiresAt := now.Add(s.sessionTTL)
	session := &auth.Session{
		ID:        uuid.NewString(),
		UserID:    user.GetId(),
		TokenHash: HashAuthSessionToken(secret),
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, connectx.InternalWith(err)
	}

	workspaces := s.userWorkspaces(ctx, user)
	logger.Info("oauth login succeeded",
		"provider", name,
		"user_id", user.GetId(),
		"username", user.GetUsername(),
		"session_id", session.ID,
		"workspace_count", len(workspaces),
	)
	return connect.NewResponse(&agentsv1.CompleteOAuthFlowResponse{
		Token:      secret,
		User:       user,
		ExpiresAt:  timestamppb.New(expiresAt),
		Workspaces: workspaces,
	}), nil
}

// upsertOAuthUser returns the existing user for (provider, externalID) or
// creates a new one. Newly created users receive role "user", are enabled,
// and get a username derived from the provider login to avoid collisions
// with existing password-based accounts.
func (s *AuthServiceServer) upsertOAuthUser(ctx context.Context, providerName string, claims *provider.Claims) (*agentsv1.User, error) {
	logger := log.FromContext(ctx)
	existing, err := s.repo.FindUserByExternalID(ctx, providerName, claims.ExternalID)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, auth.ErrUserNotFound) {
		return nil, connectx.InternalWith(err)
	}

	username := oauthUsername(providerName, claims)
	displayName := claims.Name
	if displayName == "" {
		displayName = claims.Login
	}
	if displayName == "" {
		displayName = username
	}
	now := time.Now().UTC()
	user := &agentsv1.User{
		Id:          uuid.NewString(),
		Username:    username,
		DisplayName: displayName,
		Email:       claims.Email,
		AvatarUrl:   claims.AvatarURL,
		Provider:    providerName,
		ExternalId:  claims.ExternalID,
		Role:        "user",
		CreatedAt:   timestamppb.New(now),
		UpdatedAt:   timestamppb.New(now),
	}
	if err := s.repo.CreateUser(ctx, user, ""); err != nil {
		if errors.Is(err, auth.ErrUserAlreadyExists) {
			// Username collision — fall back to a UUID-suffixed username and retry once.
			user.Username = username + "_" + shortID(user.GetId())
			if err2 := s.repo.CreateUser(ctx, user, ""); err2 != nil {
				logger.Error("oauth user create retry failed", "provider", providerName, "external_id", claims.ExternalID, "err", err2)
				return nil, connectx.InternalWith(err2)
			}
		} else {
			logger.Error("oauth user create failed", "provider", providerName, "external_id", claims.ExternalID, "err", err)
			return nil, connectx.InternalWith(err)
		}
	}
	logger.Info("oauth user created", "provider", providerName, "external_id", claims.ExternalID, "user_id", user.GetId(), "username", user.GetUsername())
	return user, nil
}

func oauthUsername(providerName string, claims *provider.Claims) string {
	base := strings.TrimSpace(claims.Login)
	if base == "" {
		base = "user"
	}
	prefix := strings.ToLower(providerName)
	if len(prefix) > 4 {
		prefix = prefix[:4]
	}
	return prefix + "_" + base
}

func shortID(id string) string {
	id = strings.ReplaceAll(id, "-", "")
	if len(id) > 6 {
		return id[:6]
	}
	return id
}

func generateOAuthState() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
