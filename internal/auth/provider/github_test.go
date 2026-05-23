package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGitHubProvider_AuthorizeURL(t *testing.T) {
	p := NewGitHub(GitHubConfig{
		ClientID:     "client-id",
		ClientSecret: "shh",
		RedirectURL:  "https://app.example.com/cb",
		Scopes:       []string{"read:user"},
	})
	if p == nil {
		t.Fatal("NewGitHub returned nil for valid config")
	}
	url := p.AuthorizeURL("state-xyz")
	for _, want := range []string{
		"https://github.com/login/oauth/authorize",
		"client_id=client-id",
		"state=state-xyz",
		"scope=read%3Auser",
		"redirect_uri=https%3A%2F%2Fapp.example.com%2Fcb",
	} {
		if !strings.Contains(url, want) {
			t.Errorf("authorize URL missing %q: %s", want, url)
		}
	}
}

func TestGitHubProvider_Exchange(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/login/oauth/access_token", func(w http.ResponseWriter, r *http.Request) {
		if got := r.FormValue("code"); got != "abc" {
			t.Errorf("token endpoint got code %q, want abc", got)
		}
		if got := r.FormValue("client_id"); got != "cid" {
			t.Errorf("token endpoint got client_id %q, want cid", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"tok-xyz","token_type":"bearer","scope":"read:user"}`))
	})
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer tok-xyz" {
			t.Errorf("user endpoint got auth %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":12345,"login":"alice","name":"Alice","email":"","avatar_url":"https://avatars.githubusercontent.com/u/12345"}`))
	})
	mux.HandleFunc("/user/emails", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"email":"alice@example.com","primary":true,"verified":true}]`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	p := NewGitHub(GitHubConfig{
		ClientID:     "cid",
		ClientSecret: "secret",
		RedirectURL:  "http://callback",
		AuthorizeURL: srv.URL + "/login/oauth/authorize",
		TokenURL:     srv.URL + "/login/oauth/access_token",
		UserURL:      srv.URL + "/user",
		EmailsURL:    srv.URL + "/user/emails",
	})
	claims, err := p.Exchange(context.Background(), "abc")
	if err != nil {
		t.Fatalf("Exchange failed: %v", err)
	}
	if claims.ExternalID != "12345" {
		t.Errorf("ExternalID = %q, want 12345", claims.ExternalID)
	}
	if claims.Login != "alice" {
		t.Errorf("Login = %q, want alice", claims.Login)
	}
	if claims.Email != "alice@example.com" {
		t.Errorf("Email = %q, want alice@example.com (fallback to /user/emails)", claims.Email)
	}
	if claims.Provider != "github" {
		t.Errorf("Provider = %q, want github", claims.Provider)
	}
}

func TestGitHubProvider_NewGitHub_NilWithoutCreds(t *testing.T) {
	if NewGitHub(GitHubConfig{}) != nil {
		t.Error("expected nil provider when credentials are missing")
	}
	if NewGitHub(GitHubConfig{ClientID: "x"}) != nil {
		t.Error("expected nil provider when secret is missing")
	}
}
