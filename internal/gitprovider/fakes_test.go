package gitprovider

// Deterministic fake GitHub / GitLab REST servers for the contract suite.
// Both serve the same fixture: private repository "acme/agents", default
// branch "main", branches develop (feedc0de) and release/v1 (0ddba11).
// writeToken has push/Developer access, readOnlyToken read-only, anything
// else is unauthorized. Routing works on the escaped path so encoded "/" in
// GitLab project IDs and branch names survives.
//
// Tree fixture (ref "main"):
//   agents/
//     my-agent/
//       prompt.md          → "You are a helpful agent."
//       description.md     → "My agent description."
//     unclaimed-dir/
//       notes.md           → "Some notes."

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
)

// fakeState is shared mutable state for the fake servers, tracking branch
// heads, created commits, branches, and change requests across write tests.
type fakeState struct {
	mu       sync.Mutex
	branches map[string]string
	// commits tracks parent→child for fast-forward verification
	commits map[string]string // sha → parent sha
	// createdPRs / createdMRs track change requests
	createdPRs int
	createdMRs int
	// lastCommitMessage captures the most recent commit message
	lastCommitMessage string
	// lastCommitActions captures paths of the most recent commit
	lastCommitActions []string
}

func newFakeState() *fakeState {
	return &fakeState{
		branches: map[string]string{
			"main":       "abc123",
			"develop":    "feedc0de",
			"release/v1": "0ddba11",
		},
		commits: map[string]string{},
	}
}

var fakeBranches = map[string]string{
	"main":       "abc123",
	"develop":    "feedc0de",
	"release/v1": "0ddba11",
}

// fakeComparisons returns "ahead" unless overridden.
var fakeComparisons = map[string]string{
	"abc123...feedc0de": "ahead",
	"feedc0de...abc123": "behind",
	"abc123...abc123":   "identical",
}

var fakeFiles = map[string]string{
	"agents/my-agent/prompt.md":      "You are a helpful agent.",
	"agents/my-agent/description.md": "My agent description.",
	"agents/unclaimed-dir/notes.md":  "Some notes.",
}

type fakeTreeEntry struct {
	path     string
	nodeType string // "blob", "tree", "commit"
	mode     string
	sha      string
	size     int64
}

var fakeTreeEntries = []fakeTreeEntry{
	{path: "agents", nodeType: "tree", mode: "040000", sha: "tree-agents"},
	{path: "agents/my-agent", nodeType: "tree", mode: "040000", sha: "tree-my-agent"},
	{path: "agents/my-agent/prompt.md", nodeType: "blob", mode: "100644", sha: "blob-prompt", size: 25},
	{path: "agents/my-agent/description.md", nodeType: "blob", mode: "100644", sha: "blob-desc", size: 22},
	{path: "agents/unclaimed-dir", nodeType: "tree", mode: "040000", sha: "tree-unclaimed"},
	{path: "agents/unclaimed-dir/notes.md", nodeType: "blob", mode: "100644", sha: "blob-notes", size: 11},
}

func newGitHubFake(t *testing.T, apiPrefix string) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlPath := strings.TrimPrefix(r.URL.EscapedPath(), apiPrefix)
		authHeader := r.Header.Get("Authorization")
		token := strings.TrimPrefix(strings.TrimPrefix(authHeader, "Bearer "), "token ")
		if token != writeToken && token != readOnlyToken {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"message":"Bad credentials"}`)
			return
		}
		switch {
		case urlPath == "/repos/acme/agents":
			push := token == writeToken
			fmt.Fprintf(w, `{"full_name":"acme/agents","private":true,"default_branch":"main",`+
				`"permissions":{"admin":false,"maintain":false,"push":%v,"triage":false,"pull":true}}`, push)
		case strings.HasPrefix(urlPath, "/repos/acme/agents/branches/"):
			raw := strings.TrimPrefix(urlPath, "/repos/acme/agents/branches/")
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

		case strings.HasPrefix(urlPath, "/repos/acme/agents/git/ref/heads/"):
			raw := strings.TrimPrefix(urlPath, "/repos/acme/agents/git/ref/heads/")
			branch, _ := url.PathUnescape(raw)
			sha, ok := fakeBranches[branch]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			fmt.Fprintf(w, `{"ref":"refs/heads/%s","object":{"sha":%q,"type":"commit"}}`, branch, sha)

		case strings.HasPrefix(urlPath, "/repos/acme/agents/git/commits/"):
			fmt.Fprint(w, `{"tree":{"sha":"root-tree-sha"}}`)

		case strings.HasPrefix(urlPath, "/repos/acme/agents/git/trees/"):
			rawTree := strings.TrimPrefix(urlPath, "/repos/acme/agents/git/trees/")
			treeSHA := strings.Split(rawTree, "?")[0]
			recursive := strings.Contains(r.URL.RawQuery, "recursive=1")
			serveGitHubTree(w, treeSHA, recursive)

		case strings.HasPrefix(urlPath, "/repos/acme/agents/contents/"):
			filePath := strings.TrimPrefix(urlPath, "/repos/acme/agents/contents/")
			filePath, _ = url.PathUnescape(filePath)
			content, ok := fakeFiles[filePath]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprint(w, `{"message":"Not Found"}`)
				return
			}
			accept := r.Header.Get("Accept")
			if strings.Contains(accept, "application/vnd.github.raw") {
				w.Header().Set("Content-Type", "application/octet-stream")
				fmt.Fprint(w, content)
			} else {
				encoded := base64.StdEncoding.EncodeToString([]byte(content))
				fmt.Fprintf(w, `{"content":%q,"encoding":"base64","size":%d}`, encoded, len(content))
			}

		case strings.HasPrefix(urlPath, "/repos/acme/agents/compare/"):
			raw := strings.TrimPrefix(urlPath, "/repos/acme/agents/compare/")
			basehead, _ := url.PathUnescape(raw)
			status, ok := fakeComparisons[basehead]
			if !ok {
				status = "diverged"
			}
			fmt.Fprintf(w, `{"status":%q,"ahead_by":1,"behind_by":0}`, status)

		case strings.HasPrefix(urlPath, "/repos/"):
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"message":"Not Found"}`)
		default:
			t.Errorf("github fake: unexpected path %q", urlPath)
			w.WriteHeader(http.StatusNotFound)
		}
	})
}

func serveGitHubTree(w http.ResponseWriter, treeSHA string, recursive bool) {
	var entries []fakeTreeEntry
	switch treeSHA {
	case "root-tree-sha":
		if recursive {
			entries = fakeTreeEntries
		} else {
			entries = []fakeTreeEntry{fakeTreeEntries[0]}
		}
	case "tree-agents":
		if recursive {
			for _, e := range fakeTreeEntries[1:] {
				rel := strings.TrimPrefix(e.path, "agents/")
				entries = append(entries, fakeTreeEntry{
					path: rel, nodeType: e.nodeType, mode: e.mode, sha: e.sha, size: e.size,
				})
			}
		} else {
			entries = []fakeTreeEntry{
				{path: "my-agent", nodeType: "tree", mode: "040000", sha: "tree-my-agent"},
				{path: "unclaimed-dir", nodeType: "tree", mode: "040000", sha: "tree-unclaimed"},
			}
		}
	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"sha":"`+treeSHA+`","tree":[`)
	for i, e := range entries {
		if i > 0 {
			fmt.Fprint(w, ",")
		}
		fmt.Fprintf(w, `{"path":%q,"mode":%q,"type":%q,"sha":%q,"size":%d}`,
			e.path, e.mode, e.nodeType, e.sha, e.size)
	}
	fmt.Fprint(w, `],"truncated":false}`)
}

func newGitHubFakeStateful(t *testing.T, apiPrefix string, state *fakeState) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlPath := strings.TrimPrefix(r.URL.EscapedPath(), apiPrefix)
		authHeader := r.Header.Get("Authorization")
		token := strings.TrimPrefix(strings.TrimPrefix(authHeader, "Bearer "), "token ")
		if token != writeToken && token != readOnlyToken {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"message":"Bad credentials"}`)
			return
		}

		// Write operations require push access.
		if r.Method == http.MethodPost || r.Method == http.MethodPatch {
			if token != writeToken {
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprint(w, `{"message":"Forbidden"}`)
				return
			}
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(urlPath, "/git/blobs"):
				// Create blob — return a fake SHA.
				var body struct {
					Content  string `json:"content"`
					Encoding string `json:"encoding"`
				}
				json.NewDecoder(r.Body).Decode(&body)
				sha := fmt.Sprintf("blob-%x", len(body.Content))
				w.WriteHeader(http.StatusCreated)
				fmt.Fprintf(w, `{"sha":%q}`, sha)
				return

			case r.Method == http.MethodPost && strings.HasSuffix(urlPath, "/git/trees"):
				// Create tree — return a fake SHA.
				w.WriteHeader(http.StatusCreated)
				fmt.Fprint(w, `{"sha":"new-tree-sha"}`)
				return

			case r.Method == http.MethodPost && strings.HasSuffix(urlPath, "/git/commits"):
				var body struct {
					Message string   `json:"message"`
					Tree    string   `json:"tree"`
					Parents []string `json:"parents"`
				}
				json.NewDecoder(r.Body).Decode(&body)
				state.mu.Lock()
				commitSHA := fmt.Sprintf("commit-%d", len(state.commits)+1)
				if len(body.Parents) > 0 {
					state.commits[commitSHA] = body.Parents[0]
				}
				state.lastCommitMessage = body.Message
				state.mu.Unlock()
				w.WriteHeader(http.StatusCreated)
				fmt.Fprintf(w, `{"sha":%q}`, commitSHA)
				return

			case r.Method == http.MethodPatch && strings.Contains(urlPath, "/git/refs/heads/"):
				var body struct {
					SHA   string `json:"sha"`
					Force bool   `json:"force"`
				}
				json.NewDecoder(r.Body).Decode(&body)
				branch := strings.TrimPrefix(urlPath, "/repos/acme/agents/git/refs/heads/")
				branch, _ = url.PathUnescape(branch)
				state.mu.Lock()
				current, ok := state.branches[branch]
				parent := state.commits[body.SHA]
				if ok && !body.Force && parent != current {
					state.mu.Unlock()
					w.WriteHeader(http.StatusUnprocessableEntity)
					fmt.Fprint(w, `{"message":"Update is not a fast forward"}`)
					return
				}
				state.branches[branch] = body.SHA
				state.mu.Unlock()
				fmt.Fprintf(w, `{"ref":"refs/heads/%s","object":{"sha":%q}}`, branch, body.SHA)
				return

			case r.Method == http.MethodPost && strings.HasSuffix(urlPath, "/git/refs"):
				var body struct {
					Ref string `json:"ref"`
					SHA string `json:"sha"`
				}
				json.NewDecoder(r.Body).Decode(&body)
				branch := strings.TrimPrefix(body.Ref, "refs/heads/")
				state.mu.Lock()
				state.branches[branch] = body.SHA
				state.mu.Unlock()
				w.WriteHeader(http.StatusCreated)
				fmt.Fprintf(w, `{"ref":%q,"object":{"sha":%q}}`, body.Ref, body.SHA)
				return

			case r.Method == http.MethodPost && strings.HasSuffix(urlPath, "/pulls"):
				var body struct {
					Title string `json:"title"`
					Body  string `json:"body"`
					Head  string `json:"head"`
					Base  string `json:"base"`
				}
				json.NewDecoder(r.Body).Decode(&body)
				state.mu.Lock()
				state.createdPRs++
				num := state.createdPRs
				state.mu.Unlock()
				w.WriteHeader(http.StatusCreated)
				fmt.Fprintf(w, `{"number":%d,"html_url":"https://github.com/acme/agents/pull/%d","title":%q}`,
					num, num, body.Title)
				return
			}
		}

		// Read operations — delegate to the stateful branch map.
		switch {
		case urlPath == "/repos/acme/agents":
			push := token == writeToken
			fmt.Fprintf(w, `{"full_name":"acme/agents","private":true,"default_branch":"main",`+
				`"permissions":{"admin":false,"maintain":false,"push":%v,"triage":false,"pull":true}}`, push)

		case strings.HasPrefix(urlPath, "/repos/acme/agents/branches/"):
			raw := strings.TrimPrefix(urlPath, "/repos/acme/agents/branches/")
			branch, _ := url.PathUnescape(raw)
			state.mu.Lock()
			sha, ok := state.branches[branch]
			state.mu.Unlock()
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			fmt.Fprintf(w, `{"name":%q,"commit":{"sha":%q}}`, branch, sha)

		case strings.HasPrefix(urlPath, "/repos/acme/agents/git/ref/heads/"):
			raw := strings.TrimPrefix(urlPath, "/repos/acme/agents/git/ref/heads/")
			branch, _ := url.PathUnescape(raw)
			state.mu.Lock()
			sha, ok := state.branches[branch]
			state.mu.Unlock()
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			fmt.Fprintf(w, `{"ref":"refs/heads/%s","object":{"sha":%q,"type":"commit"}}`, branch, sha)

		case strings.HasPrefix(urlPath, "/repos/acme/agents/git/commits/"):
			fmt.Fprint(w, `{"tree":{"sha":"root-tree-sha"}}`)

		case strings.HasPrefix(urlPath, "/repos/acme/agents/git/trees/"):
			rawTree := strings.TrimPrefix(urlPath, "/repos/acme/agents/git/trees/")
			treeSHA := strings.Split(rawTree, "?")[0]
			recursive := strings.Contains(r.URL.RawQuery, "recursive=1")
			serveGitHubTree(w, treeSHA, recursive)

		case strings.HasPrefix(urlPath, "/repos/acme/agents/contents/"):
			filePath := strings.TrimPrefix(urlPath, "/repos/acme/agents/contents/")
			filePath, _ = url.PathUnescape(filePath)
			content, ok := fakeFiles[filePath]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			fmt.Fprint(w, content)

		case strings.HasPrefix(urlPath, "/repos/acme/agents/compare/"):
			raw := strings.TrimPrefix(urlPath, "/repos/acme/agents/compare/")
			basehead, _ := url.PathUnescape(raw)
			status, ok := fakeComparisons[basehead]
			if !ok {
				status = "diverged"
			}
			fmt.Fprintf(w, `{"status":%q}`, status)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
}

func newGitLabFake(t *testing.T, apiPrefix string) http.Handler {
	t.Helper()
	project := url.PathEscape("acme/agents") // acme%2Fagents
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlPath := strings.TrimPrefix(r.URL.EscapedPath(), apiPrefix)
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
		case urlPath == "/projects/"+project:
			fmt.Fprintf(w, `{"path_with_namespace":"acme/agents","visibility":"private","default_branch":"main",`+
				`"permissions":{"project_access":{"access_level":%d},"group_access":null}}`, level)
		case strings.HasPrefix(urlPath, "/projects/"+project+"/repository/branches/"):
			raw := strings.TrimPrefix(urlPath, "/projects/"+project+"/repository/branches/")
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

		case strings.HasPrefix(urlPath, "/projects/"+project+"/repository/tree"):
			refPath := r.URL.Query().Get("path")
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, "[")
			first := true
			for _, e := range fakeTreeEntries {
				if refPath != "" && !strings.HasPrefix(e.path, refPath+"/") && e.path != refPath {
					continue
				}
				if !first {
					fmt.Fprint(w, ",")
				}
				first = false
				fmt.Fprintf(w, `{"id":%q,"name":%q,"type":%q,"path":%q,"mode":%q}`,
					e.sha, e.path[strings.LastIndex(e.path, "/")+1:], e.nodeType, e.path, e.mode)
			}
			fmt.Fprint(w, "]")

		case strings.HasPrefix(urlPath, "/projects/"+project+"/repository/files/"):
			raw := strings.TrimPrefix(urlPath, "/projects/"+project+"/repository/files/")
			filePath, _ := url.PathUnescape(raw)
			content, ok := fakeFiles[filePath]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprint(w, `{"message":"404 File Not Found"}`)
				return
			}
			encoded := base64.StdEncoding.EncodeToString([]byte(content))
			fmt.Fprintf(w, `{"file_name":%q,"file_path":%q,"size":%d,"encoding":"base64","content":%q}`,
				filePath, filePath, len(content), encoded)

		case strings.HasPrefix(urlPath, "/projects/"+project+"/repository/compare"):
			from := r.URL.Query().Get("from")
			to := r.URL.Query().Get("to")
			key := from + "..." + to
			status, ok := fakeComparisons[key]
			if !ok {
				status = "diverged"
			}
			switch status {
			case "identical":
				fmt.Fprint(w, `{"commits":[],"diffs":[],"compare_timeout":false,"compare_same_ref":true}`)
			case "ahead":
				fmt.Fprintf(w, `{"commits":[{"id":%q}],"diffs":[],"compare_timeout":false,"compare_same_ref":false}`, to)
			default:
				fmt.Fprint(w, `{"commits":[],"diffs":[],"compare_timeout":false,"compare_same_ref":false}`)
			}

		case strings.HasPrefix(urlPath, "/projects/"):
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"message":"404 Project Not Found"}`)
		default:
			t.Errorf("gitlab fake: unexpected path %q", urlPath)
			w.WriteHeader(http.StatusNotFound)
		}
	})
}

func newGitLabFakeStateful(t *testing.T, apiPrefix string, state *fakeState) http.Handler {
	t.Helper()
	project := url.PathEscape("acme/agents")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlPath := strings.TrimPrefix(r.URL.EscapedPath(), apiPrefix)
		token := r.Header.Get("PRIVATE-TOKEN")
		if token == "" {
			token = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		}
		if token != writeToken && token != readOnlyToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		level := 20
		if token == writeToken {
			level = 30
		}

		// Write operations.
		if r.Method == http.MethodPost {
			if token != writeToken {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			switch {
			case urlPath == "/projects/"+project+"/repository/commits":
				var body struct {
					Branch        string `json:"branch"`
					CommitMessage string `json:"commit_message"`
					StartSHA      string `json:"start_sha"`
					Actions       []struct {
						Action   string `json:"action"`
						FilePath string `json:"file_path"`
					} `json:"actions"`
				}
				json.NewDecoder(r.Body).Decode(&body)
				state.mu.Lock()
				commitSHA := fmt.Sprintf("gl-commit-%d", len(state.commits)+1)
				state.commits[commitSHA] = body.StartSHA
				state.lastCommitMessage = body.CommitMessage
				paths := make([]string, len(body.Actions))
				for i, a := range body.Actions {
					paths[i] = a.FilePath
				}
				state.lastCommitActions = paths
				state.branches[body.Branch] = commitSHA
				state.mu.Unlock()
				w.WriteHeader(http.StatusCreated)
				fmt.Fprintf(w, `{"id":%q,"short_id":"short","title":%q}`, commitSHA, body.CommitMessage)
				return

			case urlPath == "/projects/"+project+"/repository/branches":
				var body struct {
					Branch string `json:"branch"`
					Ref    string `json:"ref"`
				}
				json.NewDecoder(r.Body).Decode(&body)
				state.mu.Lock()
				state.branches[body.Branch] = body.Ref
				state.mu.Unlock()
				w.WriteHeader(http.StatusCreated)
				fmt.Fprintf(w, `{"name":%q,"commit":{"id":%q}}`, body.Branch, body.Ref)
				return

			case urlPath == "/projects/"+project+"/merge_requests":
				var body struct {
					SourceBranch string `json:"source_branch"`
					TargetBranch string `json:"target_branch"`
					Title        string `json:"title"`
				}
				json.NewDecoder(r.Body).Decode(&body)
				state.mu.Lock()
				state.createdMRs++
				num := state.createdMRs
				state.mu.Unlock()
				w.WriteHeader(http.StatusCreated)
				fmt.Fprintf(w, `{"iid":%d,"web_url":"https://gitlab.com/acme/agents/-/merge_requests/%d","title":%q}`,
					num, num, body.Title)
				return
			}
		}

		// Read operations with stateful branches.
		switch {
		case urlPath == "/projects/"+project:
			fmt.Fprintf(w, `{"path_with_namespace":"acme/agents","visibility":"private","default_branch":"main",`+
				`"permissions":{"project_access":{"access_level":%d},"group_access":null}}`, level)

		case strings.HasPrefix(urlPath, "/projects/"+project+"/repository/branches/"):
			raw := strings.TrimPrefix(urlPath, "/projects/"+project+"/repository/branches/")
			branch, _ := url.PathUnescape(raw)
			state.mu.Lock()
			sha, ok := state.branches[branch]
			state.mu.Unlock()
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			fmt.Fprintf(w, `{"name":%q,"commit":{"id":%q}}`, branch, sha)

		case strings.HasPrefix(urlPath, "/projects/"+project+"/repository/tree"):
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, "[]")

		case strings.HasPrefix(urlPath, "/projects/"+project+"/repository/compare"):
			from := r.URL.Query().Get("from")
			to := r.URL.Query().Get("to")
			if from == to {
				fmt.Fprint(w, `{"commits":[],"compare_same_ref":true}`)
			} else {
				fmt.Fprintf(w, `{"commits":[{"id":%q}],"compare_same_ref":false}`, to)
			}

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
}
