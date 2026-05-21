package mcpoauth

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	repo "go.orx.me/apps/butter/internal/repo/mcpoauth"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"golang.org/x/oauth2"
)

const (
	CallbackPath   = "/api/mcp/oauth/callback"
	defaultFlowTTL = 10 * time.Minute
)

type Config struct {
	CallbackBaseURL   string
	DashboardBaseURL  string
	EncryptionKey     string
	AllowInsecureHTTP bool
}

type ConfigProvider func() Config

type Service struct {
	repo    repo.Repository
	flows   FlowStore
	config  ConfigProvider
	client  *http.Client
	flowTTL time.Duration
	now     func() time.Time
}

type StartResult struct {
	AuthorizationURL string
	FlowID           string
}

type CallbackResult struct {
	Connection *repo.Connection
	ReturnURL  string
}

func NewService(connectionRepo repo.Repository, flows FlowStore, config ConfigProvider) *Service {
	return &Service{
		repo:    connectionRepo,
		flows:   flows,
		config:  config,
		client:  &http.Client{Timeout: defaultDiscoveryTimeout},
		flowTTL: defaultFlowTTL,
		now:     func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Start(ctx context.Context, workspaceID, userID string, srv *agentsv1.MCPServer, returnURL string) (*StartResult, error) {
	if s == nil || s.repo == nil || s.flows == nil {
		return nil, fmt.Errorf("mcp oauth service is not configured")
	}
	if workspaceID == "" {
		return nil, fmt.Errorf("workspace id is required")
	}
	if srv == nil {
		return nil, fmt.Errorf("mcp server is required")
	}
	if AuthType(srv) != agentsv1.MCPServerAuthType_MCP_SERVER_AUTH_TYPE_OAUTH2 {
		return nil, fmt.Errorf("mcp server is not configured for oauth2")
	}
	cfg := s.currentConfig()
	redirectURI, err := callbackURL(cfg)
	if err != nil {
		return nil, err
	}
	endpoints, err := Discover(ctx, s.client, srv, cfg.AllowInsecureHTTP)
	if err != nil {
		return nil, err
	}

	clientID := strings.TrimSpace(srv.GetAuth().GetOauth2().GetClientId())
	clientSecret := strings.TrimSpace(srv.GetAuth().GetOauth2().GetClientSecret())
	if clientID == "" {
		reg, err := s.dynamicRegister(ctx, endpoints.RegistrationURL, redirectURI, cfg.AllowInsecureHTTP)
		if err != nil {
			return nil, err
		}
		clientID = reg.ClientID
		clientSecret = reg.ClientSecret
	}
	if clientID == "" {
		return nil, fmt.Errorf("oauth2 client_id is required")
	}

	flowID := uuid.NewString()
	state, err := randomURLToken(32)
	if err != nil {
		return nil, err
	}
	verifier, err := randomURLToken(48)
	if err != nil {
		return nil, err
	}
	createdAt := s.now()
	flow := &Flow{
		ID:           flowID,
		State:        state,
		CodeVerifier: verifier,
		WorkspaceID:  workspaceID,
		UserID:       userID,
		ServerID:     srv.GetId(),
		ReturnURL:    sanitizeReturnURL(returnURL, cfg),
		RedirectURI:  redirectURI,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AuthURL:      endpoints.AuthorizationURL,
		TokenURL:     endpoints.TokenURL,
		Resource:     endpoints.Resource,
		Scopes:       uniqueStrings(endpoints.Scopes),
		CreatedAt:    createdAt,
		ExpiresAt:    createdAt.Add(s.flowTTL),
	}
	if err := s.flows.Save(ctx, flow); err != nil {
		return nil, err
	}

	oauthCfg := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Scopes:       flow.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  endpoints.AuthorizationURL,
			TokenURL: endpoints.TokenURL,
		},
	}
	authURL := oauthCfg.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge", pkceChallenge(verifier)),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("resource", endpoints.Resource),
	)
	return &StartResult{AuthorizationURL: authURL, FlowID: flowID}, nil
}

func (s *Service) Complete(ctx context.Context, workspaceID, flowID, code, state string) (*repo.Connection, error) {
	if s == nil || s.repo == nil || s.flows == nil {
		return nil, fmt.Errorf("mcp oauth service is not configured")
	}
	if workspaceID == "" {
		return nil, fmt.Errorf("workspace id is required")
	}
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("authorization code is required")
	}
	flow, err := s.flows.Consume(ctx, flowID, state, workspaceID, s.now())
	if err != nil {
		return nil, fmt.Errorf("invalid or expired oauth state")
	}
	conn, err := s.exchangeAndSave(ctx, flow, code)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (s *Service) CompleteByState(ctx context.Context, state, code string) (*CallbackResult, error) {
	if s == nil || s.repo == nil || s.flows == nil {
		return nil, fmt.Errorf("mcp oauth service is not configured")
	}
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("authorization code is required")
	}
	flow, err := s.flows.ConsumeByState(ctx, state, s.now())
	if err != nil {
		return nil, fmt.Errorf("invalid or expired oauth state")
	}
	conn, err := s.exchangeAndSave(ctx, flow, code)
	if err != nil {
		return nil, err
	}
	return &CallbackResult{Connection: conn, ReturnURL: flow.ReturnURL}, nil
}

func (s *Service) Status(ctx context.Context, workspaceID, serverID string) (*repo.Connection, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("mcp oauth service is not configured")
	}
	conn, err := s.repo.Get(ctx, workspaceID, serverID)
	if err != nil {
		return nil, err
	}
	conn.LastCheckedAt = s.now()
	return conn, nil
}

func (s *Service) Disconnect(ctx context.Context, workspaceID, serverID string) error {
	if s == nil || s.repo == nil {
		return fmt.Errorf("mcp oauth service is not configured")
	}
	return s.repo.Delete(ctx, workspaceID, serverID)
}

func (s *Service) exchangeAndSave(ctx context.Context, flow *Flow, code string) (*repo.Connection, error) {
	cfg := s.currentConfig()
	cipher, err := NewCipher(cfg.EncryptionKey)
	if err != nil {
		return nil, err
	}
	oauthCfg := oauth2.Config{
		ClientID:     flow.ClientID,
		ClientSecret: flow.ClientSecret,
		RedirectURL:  flow.RedirectURI,
		Scopes:       flow.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  flow.AuthURL,
			TokenURL: flow.TokenURL,
		},
	}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, s.client)
	tok, err := oauthCfg.Exchange(ctx, code,
		oauth2.SetAuthURLParam("code_verifier", flow.CodeVerifier),
		oauth2.SetAuthURLParam("resource", flow.Resource),
	)
	if err != nil {
		return nil, fmt.Errorf("exchange oauth code: %w", err)
	}
	if tok.AccessToken == "" {
		return nil, fmt.Errorf("oauth token response did not include an access token")
	}
	payload := tokenPayloadFromOAuth(tok, flow.Scopes)
	encryptedToken, err := encryptTokenPayload(cipher, payload)
	if err != nil {
		return nil, fmt.Errorf("encrypt token payload: %w", err)
	}
	encryptedClientSecret := ""
	if flow.ClientSecret != "" {
		encryptedClientSecret, err = cipher.Encrypt([]byte(flow.ClientSecret))
		if err != nil {
			return nil, fmt.Errorf("encrypt client secret: %w", err)
		}
	}
	now := s.now()
	conn := &repo.Connection{
		WorkspaceID:           flow.WorkspaceID,
		ServerID:              flow.ServerID,
		UserID:                flow.UserID,
		State:                 agentsv1.MCPOAuthConnectionState_MCP_OAUTH_CONNECTION_STATE_CONNECTED,
		ClientID:              flow.ClientID,
		EncryptedClientSecret: encryptedClientSecret,
		AuthorizationURL:      flow.AuthURL,
		TokenURL:              flow.TokenURL,
		Resource:              flow.Resource,
		Scopes:                flow.Scopes,
		EncryptedToken:        encryptedToken,
		ConnectedAt:           now,
		ExpiresAt:             tok.Expiry,
		LastCheckedAt:         now,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	if err := s.repo.Save(ctx, conn); err != nil {
		return nil, err
	}
	return conn.Clone(), nil
}

func (s *Service) dynamicRegister(ctx context.Context, registrationURL, redirectURI string, allowInsecureHTTP bool) (*clientRegistration, error) {
	if registrationURL == "" {
		return nil, fmt.Errorf("oauth2 client_id is required because dynamic client registration is not advertised")
	}
	if err := validateEndpointURL(registrationURL, allowInsecureHTTP); err != nil {
		return nil, err
	}
	body := map[string]any{
		"client_name":                "Butter MCP",
		"redirect_uris":              []string{redirectURI},
		"grant_types":                []string{"authorization_code", "refresh_token"},
		"response_types":             []string{"code"},
		"token_endpoint_auth_method": "none",
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, registrationURL, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("dynamic client registration returned HTTP %d", resp.StatusCode)
	}
	var reg clientRegistration
	if err := json.NewDecoder(resp.Body).Decode(&reg); err != nil {
		return nil, fmt.Errorf("decode dynamic client registration: %w", err)
	}
	if reg.ClientID == "" {
		return nil, fmt.Errorf("dynamic client registration did not return client_id")
	}
	return &reg, nil
}

type clientRegistration struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func (s *Service) currentConfig() Config {
	if s.config == nil {
		return Config{}
	}
	return s.config()
}

func callbackURL(cfg Config) (string, error) {
	base := strings.TrimRight(strings.TrimSpace(cfg.CallbackBaseURL), "/")
	if base == "" {
		return "", fmt.Errorf("mcp oauth callback_base_url is required")
	}
	raw := base + CallbackPath
	if err := validateEndpointURL(raw, cfg.AllowInsecureHTTP); err != nil {
		return "", fmt.Errorf("callback_base_url: %w", err)
	}
	return raw, nil
}

func sanitizeReturnURL(returnURL string, cfg Config) string {
	returnURL = strings.TrimSpace(returnURL)
	if returnURL != "" {
		if u, err := url.Parse(returnURL); err == nil && u.Scheme != "" && u.Host != "" {
			return returnURL
		}
	}
	dashboard := strings.TrimSpace(cfg.DashboardBaseURL)
	if dashboard != "" {
		return strings.TrimRight(dashboard, "/") + "/mcp-servers"
	}
	return "/mcp-servers"
}

func randomURLToken(size int) (string, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func pkceChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
