// Package lineartest is an in-process fake of Linear's OAuth and GraphQL
// endpoints for tests. It records what Butter sent (token grants, revokes,
// agent activities, session updates) and can inject failures.
package lineartest

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"go.orx.me/apps/butter/internal/linearapi"
)

// Grant is what a code or refresh token is exchanged for.
type Grant struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	Scopes       []string
	Identity     linearapi.Identity
}

// Activity is one recorded agentActivityCreate call.
type Activity struct {
	Token          string
	AgentSessionID string
	Type           string
	Body           string
	Action         string
	Parameter      string
	Result         string
	Ephemeral      bool
	Signal         string
	SignalMetadata map[string]any
}

// SessionUpdate is one recorded agentSessionUpdate call.
type SessionUpdate struct {
	Token          string
	AgentSessionID string
	ExternalURLs   []map[string]any
}

// Failure makes one call fail. Status is the HTTP status; OAuthError and
// GraphQLCode shape the error body.
type Failure struct {
	Status      int
	OAuthError  string
	GraphQLCode string
	Message     string
}

// Fake is a running fake Linear.
type Fake struct {
	srv *httptest.Server

	mu            sync.Mutex
	codes         map[string]Grant
	refreshes     map[string]Grant
	tokens        map[string]linearapi.Identity
	tokenRequests []url.Values
	revoked       []string
	activities    []Activity
	updates       []SessionUpdate
	tokenFailures []Failure
	revokeFails   int
	activityFail  func(Activity) *Failure
	activityHook  func(Activity)
}

// New starts a fake Linear, stopped when t ends.
func New(t testing.TB) *Fake {
	f := &Fake{
		codes:     map[string]Grant{},
		refreshes: map[string]Grant{},
		tokens:    map[string]linearapi.Identity{},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /oauth/token", f.handleToken)
	mux.HandleFunc("POST /oauth/revoke", f.handleRevoke)
	mux.HandleFunc("POST /graphql", f.handleGraphQL)
	f.srv = httptest.NewServer(mux)
	t.Cleanup(f.srv.Close)
	return f
}

// Endpoints points a client at the fake.
func (f *Fake) Endpoints() linearapi.Endpoints {
	return linearapi.Endpoints{
		AuthorizeURL: "https://linear.test/oauth/authorize",
		TokenURL:     f.srv.URL + "/oauth/token",
		RevokeURL:    f.srv.URL + "/oauth/revoke",
		GraphQLURL:   f.srv.URL + "/graphql",
	}
}

// Client is a linearapi client wired to the fake.
func (f *Fake) Client() *linearapi.Client { return linearapi.New(f.srv.Client(), f.Endpoints()) }

// AddCode makes an authorization code exchangeable for g.
func (f *Fake) AddCode(code string, g Grant) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.codes[code] = g
}

// AddRefresh makes a refresh token exchangeable for g, once.
func (f *Fake) AddRefresh(refreshToken string, g Grant) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.refreshes[refreshToken] = g
}

// AddToken makes an access token valid for GraphQL calls.
func (f *Fake) AddToken(accessToken string, id linearapi.Identity) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.tokens[accessToken] = id
}

// FailNextToken makes the next token grant fail.
func (f *Fake) FailNextToken(failure Failure) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.tokenFailures = append(f.tokenFailures, failure)
}

// FailNextRevoke makes the next revoke answer HTTP 500.
func (f *Fake) FailNextRevoke() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.revokeFails++
}

// OnActivity sets a hook deciding whether an activity post fails. A nil
// return accepts it.
func (f *Fake) OnActivity(fail func(Activity) *Failure) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.activityFail = fail
}

// WhenActivity runs hook, outside the fake's lock, after an activity is
// recorded. Tests use it to act at a precise point in a run.
func (f *Fake) WhenActivity(hook func(Activity)) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.activityHook = hook
}

// TokenRequests returns every token grant request's form.
func (f *Fake) TokenRequests() []url.Values {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]url.Values(nil), f.tokenRequests...)
}

// Revoked returns every revoked access token.
func (f *Fake) Revoked() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.revoked...)
}

// Activities returns every accepted activity, in order.
func (f *Fake) Activities() []Activity {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]Activity(nil), f.activities...)
}

// SessionUpdates returns every accepted session update, in order.
func (f *Fake) SessionUpdates() []SessionUpdate {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]SessionUpdate(nil), f.updates...)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (f *Fake) handleToken(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}
	f.mu.Lock()
	f.tokenRequests = append(f.tokenRequests, r.PostForm)
	if len(f.tokenFailures) > 0 {
		failure := f.tokenFailures[0]
		f.tokenFailures = f.tokenFailures[1:]
		f.mu.Unlock()
		writeJSON(w, failure.Status, map[string]string{"error": failure.OAuthError, "error_description": failure.Message})
		return
	}
	var grant Grant
	var ok bool
	switch r.PostForm.Get("grant_type") {
	case "authorization_code":
		grant, ok = f.codes[r.PostForm.Get("code")]
		delete(f.codes, r.PostForm.Get("code"))
	case "refresh_token":
		grant, ok = f.refreshes[r.PostForm.Get("refresh_token")]
		delete(f.refreshes, r.PostForm.Get("refresh_token"))
	}
	if ok {
		f.tokens[grant.AccessToken] = grant.Identity
	}
	f.mu.Unlock()
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_grant", "error_description": "unknown grant"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"access_token":  grant.AccessToken,
		"refresh_token": grant.RefreshToken,
		"expires_in":    grant.ExpiresIn,
		"scope":         strings.Join(grant.Scopes, ","),
		"token_type":    "Bearer",
	})
}

func (f *Fake) handleRevoke(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	f.mu.Lock()
	if f.revokeFails > 0 {
		f.revokeFails--
		f.mu.Unlock()
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	f.revoked = append(f.revoked, token)
	delete(f.tokens, token)
	f.mu.Unlock()
	w.WriteHeader(http.StatusOK)
}

func graphQLFailure(w http.ResponseWriter, failure Failure) {
	status := failure.Status
	if status == 0 {
		status = http.StatusOK
	}
	message := failure.Message
	if message == "" {
		message = "injected failure"
	}
	writeJSON(w, status, map[string]any{
		"errors": []map[string]any{{"message": message, "extensions": map[string]string{"code": failure.GraphQLCode}}},
	})
}

func (f *Fake) handleGraphQL(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	raw, _ := io.ReadAll(r.Body)
	var req struct {
		Query     string         `json:"query"`
		Variables map[string]any `json:"variables"`
	}
	_ = json.Unmarshal(raw, &req)

	f.mu.Lock()
	identity, known := f.tokens[token]
	f.mu.Unlock()
	if !known {
		graphQLFailure(w, Failure{GraphQLCode: "AUTHENTICATION_ERROR", Message: "invalid token"})
		return
	}

	input, _ := req.Variables["input"].(map[string]any)
	switch {
	case strings.Contains(req.Query, "agentActivityCreate"):
		activity := decodeActivity(token, input)
		f.mu.Lock()
		fail := f.activityFail
		f.mu.Unlock()
		if fail != nil {
			if failure := fail(activity); failure != nil {
				graphQLFailure(w, *failure)
				return
			}
		}
		f.mu.Lock()
		f.activities = append(f.activities, activity)
		hook := f.activityHook
		f.mu.Unlock()
		if hook != nil {
			hook(activity)
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"agentActivityCreate": map[string]any{"success": true}}})
	case strings.Contains(req.Query, "agentSessionUpdate"):
		id, _ := req.Variables["id"].(string)
		update := SessionUpdate{Token: token, AgentSessionID: id}
		if urls, ok := input["externalUrls"].([]any); ok {
			for _, u := range urls {
				if m, ok := u.(map[string]any); ok {
					update.ExternalURLs = append(update.ExternalURLs, m)
				}
			}
		}
		f.mu.Lock()
		f.updates = append(f.updates, update)
		f.mu.Unlock()
		writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"agentSessionUpdate": map[string]any{"success": true}}})
	case strings.Contains(req.Query, "viewer"):
		writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{
			"viewer":       map[string]any{"id": identity.AppUserID},
			"organization": map[string]any{"id": identity.OrganizationID, "name": identity.OrganizationName},
		}})
	default:
		graphQLFailure(w, Failure{Message: "operation not supported by the fake"})
	}
}

func decodeActivity(token string, input map[string]any) Activity {
	str := func(m map[string]any, key string) string {
		v, _ := m[key].(string)
		return v
	}
	content, _ := input["content"].(map[string]any)
	activity := Activity{
		Token:          token,
		AgentSessionID: str(input, "agentSessionId"),
		Type:           str(content, "type"),
		Body:           str(content, "body"),
		Action:         str(content, "action"),
		Parameter:      str(content, "parameter"),
		Result:         str(content, "result"),
		Signal:         str(input, "signal"),
	}
	activity.Ephemeral, _ = input["ephemeral"].(bool)
	activity.SignalMetadata, _ = input["signalMetadata"].(map[string]any)
	return activity
}
