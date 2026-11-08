package planeapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestAuthorizeURLUsesPlaneAppConsent(t *testing.T) {
	c := NewClient(nil, Profile{OAuthURL: "https://plane.example"})
	u, err := url.Parse(c.AuthorizeURL("client", "https://butter.example/api/plane/oauth/callback", "nonce +/&"))
	if err != nil {
		t.Fatal(err)
	}
	if u.Path != "/auth/o/authorize-app/" || u.Query().Get("state") != "nonce +/&" ||
		u.Query().Get("client_id") != "client" || u.Query().Get("response_type") != "code" {
		t.Fatalf("unexpected consent URL: %s", u)
	}
}

func TestBotTokenUsesInstallationGrantForInstallAndRenewal(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		client, secret, ok := r.BasicAuth()
		if !ok || client != "client" || secret != "secret+&=" || r.Method != http.MethodPost ||
			r.URL.Path != "/auth/o/token/" || r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("incorrect token request")
		}
		if err := r.ParseForm(); err != nil {
			t.Error(err)
		}
		if r.Form.Get("grant_type") != "client_credentials" || r.Form.Get("app_installation_id") != "inst +&=" ||
			len(r.Form) != 2 {
			t.Errorf("incorrect grant: %v", r.Form)
		}
		writeJSON(t, w, map[string]any{"access_token": "bot-token", "expires_in": 86400, "scope": "read write"})
	}))
	defer srv.Close()
	c := NewClient(srv.Client(), Profile{OAuthURL: srv.URL})
	for range 2 {
		tok, err := c.BotToken(t.Context(), "client", "secret+&=", "inst +&=")
		if err != nil || tok.AccessToken != "bot-token" || tok.InstallationID != "inst +&=" ||
			tok.ExpiresIn != 24*time.Hour || strings.Join(tok.Scopes, " ") != "read write" {
			t.Fatalf("token=%+v err=%v", tok, err)
		}
	}
	if _, err := c.BotToken(t.Context(), "client", "secret", ""); err == nil || calls != 2 {
		t.Fatalf("missing installation must not request token: calls=%d err=%v", calls, err)
	}
}

func TestOAuthRequestsDoNotFollowRedirectsOrExposeResponseSecrets(t *testing.T) {
	for _, status := range []int{http.StatusFound, http.StatusUnauthorized} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/auth/o/token/" {
					t.Error("followed token redirect")
				}
				w.Header().Set("Location", "/leak")
				w.WriteHeader(status)
				writeJSON(t, w, map[string]any{"access_token": "secret-token", "error": "secret-client"})
			}))
			defer srv.Close()
			c := NewClient(srv.Client(), Profile{OAuthURL: srv.URL})
			_, err := c.BotToken(t.Context(), "client", "secret", "inst")
			if err == nil || strings.Contains(err.Error(), "secret-") {
				t.Fatalf("unexpected token error: %v", err)
			}
		})
	}
}

func TestGetInstallationUsesOAuthOriginAndBindsIdentity(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
		ok   bool
	}{
		{"installed", `[{"id":"inst","workspace":"ws","workspace_detail":{"slug":"team","name":"Team"},"app_bot":"bot","status":"installed"}]`, true},
		{"wrong ID", `[{"id":"other","workspace":"ws","workspace_detail":{"slug":"team"},"app_bot":"bot","status":"installed"}]`, false},
		{"uninstalled", `[{"id":"inst","workspace":"ws","workspace_detail":{"slug":"team"},"app_bot":"bot","status":"uninstalled"}]`, false},
		{"missing slug", `[{"id":"inst","workspace":"ws","app_bot":"bot","status":"installed"}]`, false},
		{"ambiguous", `[{"id":"inst","workspace":"ws","workspace_detail":{"slug":"team"},"app_bot":"bot","status":"installed"},{"id":"inst","workspace":"ws","workspace_detail":{"slug":"team"},"app_bot":"bot","status":"installed"}]`, false},
		{"empty", `[]`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/auth/o/app-installation/" ||
					r.URL.Query().Get("id") != "inst" || r.Header.Get("Authorization") != "Bearer bot-token" {
					t.Errorf("unexpected installation request: %s %s", r.Method, r.URL)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()
			c := NewClient(srv.Client(), Profile{OAuthURL: srv.URL, APIBaseURL: srv.URL + "/not-oauth"})
			inst, err := c.GetInstallation(t.Context(), "bot-token", "inst")
			if (err == nil) != tc.ok {
				t.Fatalf("installation=%+v err=%v, want success=%v", inst, err, tc.ok)
			}
			if tc.ok && (inst.ID != "inst" || inst.WorkspaceID != "ws" || inst.WorkspaceSlug != "team" || inst.BotUserID != "bot") {
				t.Fatalf("incorrect identity: %+v", inst)
			}
		})
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Error(err)
	}
}
