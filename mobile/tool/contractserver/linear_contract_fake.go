package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"time"

	"go.orx.me/apps/butter/internal/linearapi"
)

// contractLinearFake is a long-lived in-process Linear OAuth/GraphQL double for
// the mobile contract fixture (lineartest.New requires testing.TB).
type contractLinearFake struct {
	srv *httptest.Server

	mu        sync.Mutex
	codes     map[string]linearapi.Token
	tokens    map[string]linearapi.Identity
	tokenBody []url.Values
}

func newContractLinearFake() *contractLinearFake {
	f := &contractLinearFake{
		codes:  map[string]linearapi.Token{},
		tokens: map[string]linearapi.Identity{},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /oauth/token", f.handleToken)
	mux.HandleFunc("POST /graphql", f.handleGraphQL)
	f.srv = httptest.NewServer(mux)
	return f
}

func (f *contractLinearFake) Close() { f.srv.Close() }

func (f *contractLinearFake) Client() *linearapi.Client {
	return linearapi.New(f.srv.Client(), linearapi.Endpoints{
		AuthorizeURL: "https://linear.test/oauth/authorize",
		TokenURL:     f.srv.URL + "/oauth/token",
		RevokeURL:    f.srv.URL + "/oauth/revoke",
		GraphQLURL:   f.srv.URL + "/graphql",
	})
}

func (f *contractLinearFake) AddCode(code string, token linearapi.Token, id linearapi.Identity) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.codes[code] = token
	f.tokens[token.AccessToken] = id
}

func (f *contractLinearFake) identityForToken(token string) linearapi.Identity {
	f.mu.Lock()
	defer f.mu.Unlock()
	if id, ok := f.tokens[token]; ok {
		return id
	}
	return linearapi.Identity{
		AppUserID: "app-user-1", OrganizationID: "org-1", OrganizationName: "Acme Corp",
	}
}

func (f *contractLinearFake) handleToken(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	f.mu.Lock()
	defer f.mu.Unlock()
	f.tokenBody = append(f.tokenBody, r.PostForm)
	code := r.PostForm.Get("code")
	if tok, ok := f.codes[code]; ok {
		delete(f.codes, code)
		f.writeToken(w, tok)
		return
	}
	if strings.HasPrefix(code, "contract-install-") {
		tok := linearapi.Token{
			AccessToken:  "access-" + code,
			RefreshToken: "refresh-" + code,
			ExpiresIn:    time.Hour,
		}
		f.tokens[tok.AccessToken] = linearapi.Identity{
			AppUserID:        "app-user-" + code,
			OrganizationID:   "org-" + code,
			OrganizationName: "Org " + code,
		}
		f.writeToken(w, tok)
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	_, _ = io.WriteString(w, `{"error":"invalid_grant"}`)
}

func (f *contractLinearFake) writeToken(w http.ResponseWriter, tok linearapi.Token) {
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access_token":  tok.AccessToken,
		"refresh_token": tok.RefreshToken,
		"expires_in":    int64(tok.ExpiresIn / time.Second),
		"scope":         "read,write",
		"token_type":    "Bearer",
	})
}

func (f *contractLinearFake) handleGraphQL(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	id := f.identityForToken(token)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": map[string]any{
			"viewer":       map[string]any{"id": id.AppUserID},
			"organization": map[string]any{"id": id.OrganizationID, "name": id.OrganizationName},
		},
	})
}
