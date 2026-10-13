package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/url"
	"strings"
	"time"

	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/auth/provider"
	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/repo/auth"
	"go.orx.me/apps/butter/internal/repo/oauthhandoff"
	"go.orx.me/apps/butter/internal/repo/oauthstate"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// oauthStateTTL bounds how long an unredeemed BeginOAuthFlow state token is
// valid. Provider authorization flows are interactive; 10 minutes is plenty
// for the user to complete the consent screen and short enough that an
// abandoned flow cannot be replayed.
const oauthStateTTL = 10 * time.Minute

// oauthHandoffTTL bounds how long a native login's handoff code may wait for
// the App. The code is delivered by a link that opens the App within seconds;
// a minute or two is ample and keeps an intercepted link short-lived.
const oauthHandoffTTL = 2 * time.Minute

// pkceChallengeLen is the length of an unpadded base64url SHA-256 digest.
const pkceChallengeLen = 43

// clock is the OAuth flows' notion of now.
func (s *AuthServiceServer) clock() time.Time {
	if s.now != nil {
		return s.now().UTC()
	}
	return time.Now().UTC()
}

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

// SetOAuthHandoffRepo wires the store of native-login handoff codes.
func (s *AuthServiceServer) SetOAuthHandoffRepo(r oauthhandoff.Repository) {
	s.handoffRepo = r
}

// SetNativeReturnURIs sets the exact return URIs native Apps may use
// (auth.native_return_uris). Empty disables native sign-in.
//
// An entry that is not a plain absolute URI (config.ValidNativeReturnURI) is
// dropped with a warning rather than trusted.
func (s *AuthServiceServer) SetNativeReturnURIs(uris []string) {
	s.nativeReturnURIs = nil
	for _, uri := range uris {
		if err := config.ValidNativeReturnURI(uri); err != nil {
			log.FromContext(context.Background()).Warn("ignoring invalid auth.native_return_uris entry", "uri", uri, "err", err)
			continue
		}
		s.nativeReturnURIs = append(s.nativeReturnURIs, uri)
	}
}

// RedeemOAuthHandoff completes a native login. The handoff code arrived in the
// App's return link and is worthless alone: it is single-use, short-lived,
// bound to the provider and state of the flow, and redeemable only with the
// PKCE code_verifier whose challenge the App registered at BeginOAuthFlow.
func (s *AuthServiceServer) RedeemOAuthHandoff(ctx context.Context, req *connect.Request[agentsv1.RedeemOAuthHandoffRequest]) (*connect.Response[agentsv1.RedeemOAuthHandoffResponse], error) {
	logger := log.FromContext(ctx)
	if s.repo == nil || s.handoffRepo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("native oauth login not available"))
	}
	name := strings.TrimSpace(req.Msg.GetProvider())
	state := strings.TrimSpace(req.Msg.GetState())
	handoff := strings.TrimSpace(req.Msg.GetHandoffCode())
	verifier := req.Msg.GetCodeVerifier()
	if name == "" {
		return nil, connectx.RequiredArgument("provider")
	}
	if state == "" {
		return nil, connectx.RequiredArgument("state")
	}
	if handoff == "" {
		return nil, connectx.RequiredArgument("handoff_code")
	}
	// Reject a malformed verifier before touching the handoff, so a client bug
	// cannot burn a good code. A well-formed wrong verifier does burn it.
	if !validPKCEVerifier(verifier) {
		return nil, connectx.InvalidArgument("code_verifier", "must be 43-128 unreserved characters")
	}

	entry, err := s.handoffRepo.Consume(ctx, hashHandoffCode(handoff), s.clock())
	if err != nil {
		if errors.Is(err, oauthhandoff.ErrNotFound) {
			return nil, connect.NewError(connect.CodePermissionDenied, errors.New("invalid or expired handoff"))
		}
		return nil, connectx.InternalWith(err)
	}
	if entry.Provider != name || !constantTimeEqual(entry.State, state) {
		logger.Warn("oauth handoff rejected: flow mismatch", "provider", name)
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("invalid or expired handoff"))
	}
	if !constantTimeEqual(pkceChallenge(verifier), entry.CodeChallenge) {
		logger.Warn("oauth handoff rejected: verifier mismatch", "provider", name)
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("invalid or expired handoff"))
	}

	user, err := s.repo.GetUser(ctx, entry.UserID)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, connect.NewError(connect.CodePermissionDenied, errors.New("invalid or expired handoff"))
		}
		return nil, connectx.InternalWith(err)
	}
	if user.GetDisabled() {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("user disabled"))
	}
	secret, expiresAt, session, err := s.issueSession(ctx, user)
	if err != nil {
		return nil, err
	}
	workspaces := s.userWorkspaces(ctx, user)
	logger.Info("native oauth login succeeded",
		"provider", name,
		"user_id", user.GetId(),
		"session_id", session.ID,
		"workspace_count", len(workspaces),
	)
	return connect.NewResponse(&agentsv1.RedeemOAuthHandoffResponse{
		Token:      secret,
		User:       user,
		ExpiresAt:  timestamppb.New(expiresAt),
		Workspaces: workspaces,
	}), nil
}

// issueSession creates the user's session and returns its one-time secret.
func (s *AuthServiceServer) issueSession(ctx context.Context, user *agentsv1.User) (string, time.Time, *auth.Session, error) {
	secret, err := generateSessionSecret()
	if err != nil {
		return "", time.Time{}, nil, connectx.InternalWith(err)
	}
	now := s.clock()
	expiresAt := now.Add(s.sessionTTL)
	session := &auth.Session{
		ID:        uuid.NewString(),
		UserID:    user.GetId(),
		TokenHash: HashAuthSessionToken(secret),
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return "", time.Time{}, nil, connectx.InternalWith(err)
	}
	return secret, expiresAt, session, nil
}

// checkNativeClient validates a native App's request against the server's
// allow-list. The return URI must match an entry exactly: no prefixes, no
// extra query, no other schemes.
func (s *AuthServiceServer) checkNativeClient(n *agentsv1.NativeOAuthClient) error {
	if s.handoffRepo == nil || len(s.nativeReturnURIs) == 0 {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("native oauth login is not enabled on this server"))
	}
	if !s.nativeReturnURIAllowed(n.GetReturnUri()) {
		return connectx.InvalidArgument("native.return_uri", "is not an allowed return URI")
	}
	if !validPKCEChallenge(n.GetCodeChallenge()) {
		return connectx.InvalidArgument("native.code_challenge", "must be a base64url SHA-256 digest")
	}
	return nil
}

func (s *AuthServiceServer) nativeReturnURIAllowed(uri string) bool {
	for _, allowed := range s.nativeReturnURIs {
		if allowed != "" && allowed == uri {
			return true
		}
	}
	return false
}

// nativeHandoff turns a finished native flow into the link that returns the
// browser to the App: the allow-listed URI plus a fresh one-time handoff code
// and the flow's state. The session itself is only minted at redemption.
func (s *AuthServiceServer) nativeHandoff(ctx context.Context, entry *oauthstate.Entry, state string, user *agentsv1.User) (*agentsv1.NativeOAuthHandoff, error) {
	if s.handoffRepo == nil || !s.nativeReturnURIAllowed(entry.NativeReturnURI) {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("native oauth login is not enabled on this server"))
	}
	link, err := url.Parse(entry.NativeReturnURI)
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	code, err := generateOAuthState()
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	now := s.clock()
	if err := s.handoffRepo.Create(ctx, &oauthhandoff.Entry{
		CodeHash:      hashHandoffCode(code),
		Provider:      entry.Provider,
		State:         state,
		UserID:        user.GetId(),
		CodeChallenge: entry.CodeChallenge,
		CreatedAt:     now,
		ExpiresAt:     now.Add(oauthHandoffTTL),
	}); err != nil {
		return nil, connectx.InternalWith(err)
	}
	q := link.Query()
	q.Set("handoff", code)
	q.Set("state", state)
	link.RawQuery = q.Encode()
	return &agentsv1.NativeOAuthHandoff{ReturnUrl: link.String()}, nil
}

func hashHandoffCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

func pkceChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func constantTimeEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// validPKCEChallenge accepts the shape of an S256 challenge: 43 base64url
// characters.
func validPKCEChallenge(c string) bool {
	if len(c) != pkceChallengeLen {
		return false
	}
	for i := 0; i < len(c); i++ {
		if !isBase64URL(c[i]) {
			return false
		}
	}
	return true
}

// validPKCEVerifier is RFC 7636's code-verifier: 43-128 unreserved characters.
func validPKCEVerifier(v string) bool {
	if len(v) < 43 || len(v) > 128 {
		return false
	}
	for i := 0; i < len(v); i++ {
		c := v[i]
		if !isBase64URL(c) && c != '.' && c != '~' {
			return false
		}
	}
	return true
}

func isBase64URL(c byte) bool {
	return c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-' || c == '_'
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
	native := req.Msg.GetNative()
	if native != nil {
		if err := s.checkNativeClient(native); err != nil {
			return nil, err
		}
	}
	state, err := generateOAuthState()
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	now := s.clock()
	entry := &oauthstate.Entry{
		State:       state,
		Provider:    name,
		RedirectURI: strings.TrimSpace(req.Msg.GetRedirectUri()),
		CreatedAt:   now,
		ExpiresAt:   now.Add(oauthStateTTL),
	}
	if native != nil {
		entry.NativeReturnURI = native.GetReturnUri()
		entry.CodeChallenge = native.GetCodeChallenge()
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

	entry, err := s.stateRepo.Consume(ctx, state, s.clock())
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

	if entry.NativeReturnURI != "" {
		// A native App began this flow: the browser gets a way back to the App,
		// never a session.
		handoff, err := s.nativeHandoff(ctx, entry, state, user)
		if err != nil {
			return nil, err
		}
		logger.Info("native oauth handoff issued", "provider", name, "user_id", user.GetId())
		return connect.NewResponse(&agentsv1.CompleteOAuthFlowResponse{NativeHandoff: handoff}), nil
	}

	secret, expiresAt, session, err := s.issueSession(ctx, user)
	if err != nil {
		return nil, err
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
