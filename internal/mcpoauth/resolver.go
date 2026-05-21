package mcpoauth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	repo "go.orx.me/apps/butter/internal/repo/mcpoauth"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"golang.org/x/oauth2"
)

const refreshSkew = 60 * time.Second

type Resolver struct {
	repo   repo.Repository
	config ConfigProvider
	client *http.Client
	locks  sync.Map
}

func NewResolver(connectionRepo repo.Repository, config ConfigProvider) *Resolver {
	return &Resolver{
		repo:   connectionRepo,
		config: config,
		client: &http.Client{Timeout: defaultDiscoveryTimeout},
	}
}

func (r *Resolver) HTTPClientForMCP(ctx context.Context, srv *agentsv1.MCPServer) (*http.Client, error) {
	authType := AuthType(srv)
	switch authType {
	case agentsv1.MCPServerAuthType_MCP_SERVER_AUTH_TYPE_NONE:
		if len(srv.GetHeaders()) == 0 {
			return nil, nil
		}
		return staticHeaderClient(srv.GetHeaders()), nil
	case agentsv1.MCPServerAuthType_MCP_SERVER_AUTH_TYPE_STATIC_HEADERS:
		return staticHeaderClient(srv.GetHeaders()), nil
	case agentsv1.MCPServerAuthType_MCP_SERVER_AUTH_TYPE_OAUTH2:
		if r == nil || r.repo == nil {
			return nil, fmt.Errorf("mcp oauth resolver is not configured")
		}
		if srv.GetWorkspaceId() == "" || srv.GetId() == "" {
			return nil, fmt.Errorf("oauth2 mcp server requires workspace_id and id")
		}
		return &http.Client{
			Transport: &oauthTransport{
				base:    http.DefaultTransport,
				headers: srv.GetHeaders(),
				source:  r.tokenSource(ctx, srv.GetWorkspaceId(), srv.GetId()),
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported mcp auth type %v", authType)
	}
}

func (r *Resolver) tokenSource(ctx context.Context, workspaceID, serverID string) *persistentTokenSource {
	return &persistentTokenSource{
		repo:        r.repo,
		config:      r.config,
		httpClient:  r.client,
		locks:       &r.locks,
		workspaceID: workspaceID,
		serverID:    serverID,
		now:         func() time.Time { return time.Now().UTC() },
		baseCtx:     ctx,
	}
}

type persistentTokenSource struct {
	repo        repo.Repository
	config      ConfigProvider
	httpClient  *http.Client
	locks       *sync.Map
	workspaceID string
	serverID    string
	now         func() time.Time
	baseCtx     context.Context
}

func (s *persistentTokenSource) Bearer(ctx context.Context) (string, error) {
	if ctx == nil {
		ctx = s.baseCtx
	}
	if ctx == nil {
		ctx = context.Background()
	}
	key := s.workspaceID + ":" + s.serverID
	lockAny, _ := s.locks.LoadOrStore(key, &sync.Mutex{})
	lock := lockAny.(*sync.Mutex)
	lock.Lock()
	defer lock.Unlock()

	conn, err := s.repo.Get(ctx, s.workspaceID, s.serverID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return "", fmt.Errorf("mcp oauth connection requires authorization")
		}
		return "", err
	}
	if conn.State != agentsv1.MCPOAuthConnectionState_MCP_OAUTH_CONNECTION_STATE_CONNECTED {
		return "", fmt.Errorf("mcp oauth connection requires authorization")
	}
	cfg := Config{}
	if s.config != nil {
		cfg = s.config()
	}
	cipher, err := NewCipher(cfg.EncryptionKey)
	if err != nil {
		return "", err
	}
	payload, err := decryptTokenPayload(cipher, conn.EncryptedToken)
	if err != nil {
		_ = s.markReauth(ctx, "stored token could not be decrypted")
		return "", fmt.Errorf("mcp oauth token is not usable")
	}
	now := s.now()
	if payload.AccessToken != "" && (payload.Expiry.IsZero() || now.Add(refreshSkew).Before(payload.Expiry)) {
		return payload.AccessToken, nil
	}
	if payload.RefreshToken == "" {
		_ = s.markReauth(ctx, "refresh token is missing")
		return "", fmt.Errorf("mcp oauth connection requires reauthorization")
	}

	clientSecret := ""
	if conn.EncryptedClientSecret != "" {
		raw, err := cipher.Decrypt(conn.EncryptedClientSecret)
		if err != nil {
			_ = s.markReauth(ctx, "stored client secret could not be decrypted")
			return "", fmt.Errorf("mcp oauth client secret is not usable")
		}
		clientSecret = string(raw)
	}
	oauthCfg := oauth2.Config{
		ClientID:     conn.ClientID,
		ClientSecret: clientSecret,
		Scopes:       conn.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  conn.AuthorizationURL,
			TokenURL: conn.TokenURL,
		},
	}
	oauthCtx := context.WithValue(ctx, oauth2.HTTPClient, s.httpClient)
	refreshed, err := oauthCfg.TokenSource(oauthCtx, oauthTokenFromPayload(payload)).Token()
	if err != nil {
		_ = s.markReauth(ctx, "refresh token failed")
		return "", fmt.Errorf("mcp oauth token refresh failed")
	}
	if refreshed.AccessToken == "" {
		_ = s.markReauth(ctx, "refresh response did not include an access token")
		return "", fmt.Errorf("mcp oauth token refresh failed")
	}
	if refreshed.RefreshToken == "" {
		refreshed.RefreshToken = payload.RefreshToken
	}
	nextPayload := tokenPayloadFromOAuth(refreshed, conn.Scopes)
	encryptedToken, err := encryptTokenPayload(cipher, nextPayload)
	if err != nil {
		return "", fmt.Errorf("encrypt refreshed mcp oauth token: %w", err)
	}
	conn.EncryptedToken = encryptedToken
	conn.ExpiresAt = refreshed.Expiry
	conn.LastError = ""
	conn.LastCheckedAt = now
	conn.UpdatedAt = now
	conn.State = agentsv1.MCPOAuthConnectionState_MCP_OAUTH_CONNECTION_STATE_CONNECTED
	conn.ReauthorizationRequired = false
	if err := s.repo.Save(ctx, conn); err != nil {
		return "", err
	}
	return refreshed.AccessToken, nil
}

func (s *persistentTokenSource) markReauth(ctx context.Context, detail string) error {
	return s.repo.MarkState(ctx, s.workspaceID, s.serverID, agentsv1.MCPOAuthConnectionState_MCP_OAUTH_CONNECTION_STATE_REAUTHORIZATION_REQUIRED, detail, s.now())
}

type oauthTransport struct {
	base    http.RoundTripper
	headers map[string]string
	source  *persistentTokenSource
}

func (t *oauthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	token, err := t.source.Bearer(req.Context())
	if err != nil {
		return nil, err
	}
	clone := req.Clone(req.Context())
	for k, v := range t.headers {
		if strings.EqualFold(k, "Authorization") {
			continue
		}
		clone.Header.Set(k, v)
	}
	clone.Header.Set("Authorization", "Bearer "+token)
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(clone)
}

func staticHeaderClient(headers map[string]string) *http.Client {
	if len(headers) == 0 {
		return nil
	}
	return &http.Client{
		Transport: &staticHeaderTransport{
			base:    http.DefaultTransport,
			headers: headers,
		},
	}
}

type staticHeaderTransport struct {
	base    http.RoundTripper
	headers map[string]string
}

func (t *staticHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	for k, v := range t.headers {
		clone.Header.Set(k, v)
	}
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(clone)
}
