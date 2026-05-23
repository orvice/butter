package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGoogleProvider_AuthorizeURL(t *testing.T) {
	p := NewGoogle(GoogleConfig{
		ClientID:     "g-client",
		ClientSecret: "shh",
		RedirectURL:  "https://app.example.com/cb",
	})
	if p == nil {
		t.Fatal("NewGoogle returned nil for valid config")
	}
	url := p.AuthorizeURL("state-xyz")
	for _, want := range []string{
		"https://accounts.google.com/o/oauth2/v2/auth",
		"client_id=g-client",
		"state=state-xyz",
		"response_type=code",
		"scope=openid+email+profile",
		"redirect_uri=https%3A%2F%2Fapp.example.com%2Fcb",
	} {
		if !strings.Contains(url, want) {
			t.Errorf("authorize URL missing %q: %s", want, url)
		}
	}
}

func TestGoogleProvider_Exchange(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		if got := r.FormValue("grant_type"); got != "authorization_code" {
			t.Errorf("token endpoint grant_type = %q", got)
		}
		if got := r.FormValue("code"); got != "abc" {
			t.Errorf("token endpoint code = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"tok","token_type":"Bearer","expires_in":3600,"scope":"openid email profile"}`))
	})
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer tok" {
			t.Errorf("userinfo endpoint auth = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sub":"105050506","email":"bob@example.com","email_verified":true,"name":"Bob Smith","picture":"https://lh3.googleusercontent.com/a/bob"}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	p := NewGoogle(GoogleConfig{
		ClientID:     "cid",
		ClientSecret: "secret",
		RedirectURL:  "http://callback",
		TokenURL:     srv.URL + "/token",
		UserURL:      srv.URL + "/userinfo",
	})
	claims, err := p.Exchange(context.Background(), "abc")
	if err != nil {
		t.Fatalf("Exchange failed: %v", err)
	}
	if claims.ExternalID != "105050506" {
		t.Errorf("ExternalID = %q, want 105050506", claims.ExternalID)
	}
	if claims.Login != "bob" {
		t.Errorf("Login = %q, want bob (local-part of email)", claims.Login)
	}
	if claims.Email != "bob@example.com" {
		t.Errorf("Email = %q", claims.Email)
	}
	if claims.Provider != "google" {
		t.Errorf("Provider = %q, want google", claims.Provider)
	}
}

func TestGoogleProvider_NewGoogle_NilWithoutCreds(t *testing.T) {
	if NewGoogle(GoogleConfig{}) != nil {
		t.Error("expected nil provider when credentials are missing")
	}
	if NewGoogle(GoogleConfig{ClientID: "x"}) != nil {
		t.Error("expected nil provider when secret is missing")
	}
}
