package gitprovider

// Deterministic fake GitHub / GitLab REST servers for the contract suite.
// Both serve the same fixture: private repository "acme/agents", default
// branch "main", branches develop (feedc0de) and release/v1 (0ddba11).
// writeToken has push/Developer access, readOnlyToken read-only, anything
// else is unauthorized. Routing works on the escaped path so encoded "/" in
// GitLab project IDs and branch names survives.

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

var fakeBranches = map[string]string{
	"main":       "abc123",
	"develop":    "feedc0de",
	"release/v1": "0ddba11",
}

func newGitHubFake(t *testing.T, apiPrefix string) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.EscapedPath(), apiPrefix)
		auth := r.Header.Get("Authorization")
		token := strings.TrimPrefix(strings.TrimPrefix(auth, "Bearer "), "token ")
		if token != writeToken && token != readOnlyToken {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"message":"Bad credentials"}`)
			return
		}
		switch {
		case path == "/repos/acme/agents":
			push := token == writeToken
			fmt.Fprintf(w, `{"full_name":"acme/agents","private":true,"default_branch":"main",`+
				`"permissions":{"admin":false,"maintain":false,"push":%v,"triage":false,"pull":true}}`, push)
		case strings.HasPrefix(path, "/repos/acme/agents/branches/"):
			raw := strings.TrimPrefix(path, "/repos/acme/agents/branches/")
			branch, err := url.PathUnescape(raw)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			sha, ok := fakeBranches[branch]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprint(w, `{"message":"Branch not found"}`)
				return
			}
			fmt.Fprintf(w, `{"name":%q,"commit":{"sha":%q}}`, branch, sha)
		case strings.HasPrefix(path, "/repos/"):
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"message":"Not Found"}`)
		default:
			t.Errorf("github fake: unexpected path %q", path)
			w.WriteHeader(http.StatusNotFound)
		}
	})
}

func newGitLabFake(t *testing.T, apiPrefix string) http.Handler {
	t.Helper()
	project := url.PathEscape("acme/agents") // acme%2Fagents
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.EscapedPath(), apiPrefix)
		token := r.Header.Get("PRIVATE-TOKEN")
		if token == "" {
			token = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		}
		if token != writeToken && token != readOnlyToken {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"message":"401 Unauthorized"}`)
			return
		}
		level := 20 // Reporter
		if token == writeToken {
			level = 30 // Developer
		}
		switch {
		case path == "/projects/"+project:
			fmt.Fprintf(w, `{"path_with_namespace":"acme/agents","visibility":"private","default_branch":"main",`+
				`"permissions":{"project_access":{"access_level":%d},"group_access":null}}`, level)
		case strings.HasPrefix(path, "/projects/"+project+"/repository/branches/"):
			raw := strings.TrimPrefix(path, "/projects/"+project+"/repository/branches/")
			branch, err := url.PathUnescape(raw)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			sha, ok := fakeBranches[branch]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprint(w, `{"message":"404 Branch Not Found"}`)
				return
			}
			fmt.Fprintf(w, `{"name":%q,"commit":{"id":%q}}`, branch, sha)
		case strings.HasPrefix(path, "/projects/"):
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"message":"404 Project Not Found"}`)
		default:
			t.Errorf("gitlab fake: unexpected path %q", path)
			w.WriteHeader(http.StatusNotFound)
		}
	})
}
