package mcpoauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	repo "go.orx.me/apps/butter/internal/repo/mcpoauth"
	"go.orx.me/apps/butter/internal/repo/mcpoauth/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const testKey = "0123456789abcdef0123456789abcdef"

func TestServiceStartAndCompleteStoresEncryptedConnection(t *testing.T) {
	var tokenRequests int
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/token" {
			http.NotFound(w, r)
			return
		}
		tokenRequests++
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if r.Form.Get("code") != "code-1" {
			t.Fatalf("unexpected code %q", r.Form.Get("code"))
		}
		if r.Form.Get("code_verifier") == "" {
			t.Fatalf("missing pkce verifier")
		}
		writeJSON(w, map[string]any{
			"access_token":  "access-1",
			"refresh_token": "refresh-1",
			"token_type":    "Bearer",
			"expires_in":    3600,
		})
	}))
	defer authServer.Close()

	store := memory.New()
	svc := NewService(store, NewMemoryFlowStore(), func() Config {
		return Config{CallbackBaseURL: "http://127.0.0.1:8080", EncryptionKey: testKey}
	})
	srv := oauthServer("srv", authServer.URL+"/authorize", authServer.URL+"/token")
	start, err := svc.Start(context.Background(), "ws-a", "user-a", srv, "http://dashboard.local/mcp-servers")
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	authURL, err := url.Parse(start.AuthorizationURL)
	if err != nil {
		t.Fatalf("parse auth url: %v", err)
	}
	q := authURL.Query()
	if q.Get("state") == "" || q.Get("code_challenge") == "" {
		t.Fatalf("authorization url missing state or pkce: %s", start.AuthorizationURL)
	}
	conn, err := svc.Complete(context.Background(), "ws-a", start.FlowID, "code-1", q.Get("state"))
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if tokenRequests != 1 {
		t.Fatalf("expected one token request, got %d", tokenRequests)
	}
	if conn.State != agentsv1.MCPOAuthConnectionState_MCP_OAUTH_CONNECTION_STATE_CONNECTED {
		t.Fatalf("unexpected state %v", conn.State)
	}
	stored, err := store.Get(context.Background(), "ws-a", "srv")
	if err != nil {
		t.Fatalf("stored connection: %v", err)
	}
	if strings.Contains(stored.EncryptedToken, "access-1") || strings.Contains(stored.EncryptedToken, "refresh-1") {
		t.Fatalf("token payload was not encrypted")
	}
	cipher, err := NewCipher(testKey)
	if err != nil {
		t.Fatalf("cipher: %v", err)
	}
	payload, err := decryptTokenPayload(cipher, stored.EncryptedToken)
	if err != nil {
		t.Fatalf("decrypt token: %v", err)
	}
	if payload.AccessToken != "access-1" || payload.RefreshToken != "refresh-1" {
		t.Fatalf("unexpected payload %#v", payload)
	}
}

func TestServiceRejectsInvalidStateAndCrossWorkspace(t *testing.T) {
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"access_token": "access-1", "refresh_token": "refresh-1"})
	}))
	defer authServer.Close()

	store := memory.New()
	svc := NewService(store, NewMemoryFlowStore(), func() Config {
		return Config{CallbackBaseURL: "http://127.0.0.1:8080", EncryptionKey: testKey}
	})
	srv := oauthServer("srv", authServer.URL+"/authorize", authServer.URL)
	start, err := svc.Start(context.Background(), "ws-a", "user-a", srv, "")
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if _, err := svc.Complete(context.Background(), "ws-a", start.FlowID, "code", "wrong-state"); err == nil {
		t.Fatalf("expected invalid state error")
	}

	start, err = svc.Start(context.Background(), "ws-a", "user-a", srv, "")
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	authURL, _ := url.Parse(start.AuthorizationURL)
	if _, err := svc.Complete(context.Background(), "ws-b", start.FlowID, "code", authURL.Query().Get("state")); err == nil {
		t.Fatalf("expected cross-workspace rejection")
	}
	if _, err := store.Get(context.Background(), "ws-b", "srv"); err == nil {
		t.Fatalf("cross-workspace callback stored credentials")
	}
}

func TestServiceMissingEncryptionKey(t *testing.T) {
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"access_token": "access-1", "refresh_token": "refresh-1"})
	}))
	defer authServer.Close()

	store := memory.New()
	svc := NewService(store, NewMemoryFlowStore(), func() Config {
		return Config{CallbackBaseURL: "http://127.0.0.1:8080"}
	})
	srv := oauthServer("srv", authServer.URL+"/authorize", authServer.URL)
	start, err := svc.Start(context.Background(), "ws-a", "user-a", srv, "")
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	authURL, _ := url.Parse(start.AuthorizationURL)
	if _, err := svc.Complete(context.Background(), "ws-a", start.FlowID, "code", authURL.Query().Get("state")); err == nil {
		t.Fatalf("expected missing encryption key error")
	}
}

func TestResolverInjectsBearerAndPersistsRotatedRefreshToken(t *testing.T) {
	var sawRefreshToken string
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		sawRefreshToken = r.Form.Get("refresh_token")
		writeJSON(w, map[string]any{
			"access_token":  "access-2",
			"refresh_token": "refresh-2",
			"token_type":    "Bearer",
			"expires_in":    3600,
		})
	}))
	defer authServer.Close()

	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer access-2" {
			t.Fatalf("unexpected authorization header %q", got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer target.Close()

	store := memory.New()
	cipher, err := NewCipher(testKey)
	if err != nil {
		t.Fatalf("cipher: %v", err)
	}
	encrypted, err := encryptTokenPayload(cipher, TokenPayload{
		AccessToken:  "access-1",
		RefreshToken: "refresh-1",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(-time.Minute),
	})
	if err != nil {
		t.Fatalf("encrypt token: %v", err)
	}
	if err := store.Save(context.Background(), &repo.Connection{
		WorkspaceID:    "ws",
		ServerID:       "srv",
		State:          agentsv1.MCPOAuthConnectionState_MCP_OAUTH_CONNECTION_STATE_CONNECTED,
		ClientID:       "client",
		TokenURL:       authServer.URL,
		EncryptedToken: encrypted,
	}); err != nil {
		t.Fatalf("save connection: %v", err)
	}
	resolver := NewResolver(store, func() Config { return Config{EncryptionKey: testKey} })
	client, err := resolver.HTTPClientForMCP(context.Background(), &agentsv1.MCPServer{
		Id:          "srv",
		WorkspaceId: "ws",
		Url:         target.URL,
		Auth: &agentsv1.MCPServerAuth{
			Type: agentsv1.MCPServerAuthType_MCP_SERVER_AUTH_TYPE_OAUTH2,
			Oauth2: &agentsv1.MCPServerOAuth2Config{
				ClientId: "client",
			},
		},
	})
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	req, _ := http.NewRequest(http.MethodGet, target.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	resp.Body.Close()
	if sawRefreshToken != "refresh-1" {
		t.Fatalf("refresh did not use stored refresh token: %q", sawRefreshToken)
	}
	stored, err := store.Get(context.Background(), "ws", "srv")
	if err != nil {
		t.Fatalf("get stored: %v", err)
	}
	payload, err := decryptTokenPayload(cipher, stored.EncryptedToken)
	if err != nil {
		t.Fatalf("decrypt stored: %v", err)
	}
	if payload.RefreshToken != "refresh-2" {
		t.Fatalf("rotated refresh token was not persisted: %#v", payload)
	}
}

func TestResolverMarksReauthorizationRequiredOnRefreshFailure(t *testing.T) {
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "invalid_grant", http.StatusBadRequest)
	}))
	defer authServer.Close()
	store := memory.New()
	cipher, err := NewCipher(testKey)
	if err != nil {
		t.Fatalf("cipher: %v", err)
	}
	encrypted, err := encryptTokenPayload(cipher, TokenPayload{
		AccessToken:  "access-1",
		RefreshToken: "refresh-1",
		Expiry:       time.Now().Add(-time.Minute),
	})
	if err != nil {
		t.Fatalf("encrypt token: %v", err)
	}
	if err := store.Save(context.Background(), &repo.Connection{
		WorkspaceID:    "ws",
		ServerID:       "srv",
		State:          agentsv1.MCPOAuthConnectionState_MCP_OAUTH_CONNECTION_STATE_CONNECTED,
		ClientID:       "client",
		TokenURL:       authServer.URL,
		EncryptedToken: encrypted,
	}); err != nil {
		t.Fatalf("save connection: %v", err)
	}
	resolver := NewResolver(store, func() Config { return Config{EncryptionKey: testKey} })
	client, err := resolver.HTTPClientForMCP(context.Background(), &agentsv1.MCPServer{
		Id:          "srv",
		WorkspaceId: "ws",
		Auth:        &agentsv1.MCPServerAuth{Type: agentsv1.MCPServerAuthType_MCP_SERVER_AUTH_TYPE_OAUTH2},
	})
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1/", nil)
	if _, err := client.Do(req); err == nil {
		t.Fatalf("expected refresh failure")
	}
	conn, err := store.Get(context.Background(), "ws", "srv")
	if err != nil {
		t.Fatalf("get connection: %v", err)
	}
	if conn.State != agentsv1.MCPOAuthConnectionState_MCP_OAUTH_CONNECTION_STATE_REAUTHORIZATION_REQUIRED {
		t.Fatalf("expected reauth required, got %v", conn.State)
	}
}

func TestResolverPreservesStaticHeaders(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Test"); got != "ok" {
			t.Fatalf("missing static header, got %q", got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer target.Close()
	resolver := NewResolver(nil, nil)
	client, err := resolver.HTTPClientForMCP(context.Background(), &agentsv1.MCPServer{
		Headers: map[string]string{"X-Test": "ok"},
		Auth:    &agentsv1.MCPServerAuth{Type: agentsv1.MCPServerAuthType_MCP_SERVER_AUTH_TYPE_STATIC_HEADERS},
	})
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	req, _ := http.NewRequest(http.MethodGet, target.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	resp.Body.Close()
}

func oauthServer(id, authURL, tokenURL string) *agentsv1.MCPServer {
	return &agentsv1.MCPServer{
		Id:        id,
		Name:      id,
		Transport: agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
		Url:       "https://mcp.example.com/mcp",
		Auth: &agentsv1.MCPServerAuth{
			Type: agentsv1.MCPServerAuthType_MCP_SERVER_AUTH_TYPE_OAUTH2,
			Oauth2: &agentsv1.MCPServerOAuth2Config{
				ClientId:         "client",
				AuthorizationUrl: authURL,
				TokenUrl:         tokenURL,
				Scopes:           []string{"tools.read", "offline_access"},
			},
		},
	}
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}
