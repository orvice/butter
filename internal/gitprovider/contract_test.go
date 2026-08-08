package gitprovider

// Contract suite: every provider adapter must expose identical behavior
// through the Client seam (issue #214). Each case builds a fake host serving
// that provider's REST dialect and runs the same assertions, including
// self-hosted API roots mounted under a subpath.

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	writeToken    = "tok-write-secret"
	readOnlyToken = "tok-read-secret"
)

// fixture semantics shared by both fakes: one repository "acme/agents" with
// default branch "main" and an extra branch "develop". writeToken has push
// (GitLab: Developer) access; readOnlyToken has read-only access; any other
// token is unauthorized.

type harness struct {
	kind Kind
	// newHandler returns an http.Handler serving the provider REST dialect
	// rooted at apiPrefix ("" mounts at server root).
	newHandler func(t *testing.T, apiPrefix string) http.Handler
}

func harnesses() []harness {
	return []harness{
		{kind: KindGitHub, newHandler: newGitHubFake},
		{kind: KindGitLab, newHandler: newGitLabFake},
	}
}

func newClientForTest(t *testing.T, kind Kind, baseURL, repo, token string) Client {
	t.Helper()
	c, err := New(Config{Kind: kind, APIBaseURL: baseURL, Repository: repo, Token: token})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func TestProviderContract(t *testing.T) {
	for _, h := range harnesses() {
		t.Run(string(h.kind), func(t *testing.T) {
			srv := httptest.NewServer(h.newHandler(t, ""))
			t.Cleanup(srv.Close)
			ctx := context.Background()

			t.Run("RepositoryWithWriteAccess", func(t *testing.T) {
				c := newClientForTest(t, h.kind, srv.URL, "acme/agents", writeToken)
				repo, err := c.GetRepository(ctx)
				if err != nil {
					t.Fatalf("GetRepository: %v", err)
				}
				if repo.FullName != "acme/agents" {
					t.Errorf("FullName = %q", repo.FullName)
				}
				if !repo.Private {
					t.Error("expected private repository")
				}
				if repo.DefaultBranch != "main" {
					t.Errorf("DefaultBranch = %q", repo.DefaultBranch)
				}
				if !repo.CanRead || !repo.CanWrite || !repo.CanOpenChangeRequests {
					t.Errorf("capabilities = read:%v write:%v cr:%v, want all true",
						repo.CanRead, repo.CanWrite, repo.CanOpenChangeRequests)
				}
			})

			t.Run("RepositoryWithReadOnlyAccess", func(t *testing.T) {
				c := newClientForTest(t, h.kind, srv.URL, "acme/agents", readOnlyToken)
				repo, err := c.GetRepository(ctx)
				if err != nil {
					t.Fatalf("GetRepository: %v", err)
				}
				if !repo.CanRead {
					t.Error("expected read capability")
				}
				if repo.CanWrite || repo.CanOpenChangeRequests {
					t.Errorf("read-only token got write:%v cr:%v", repo.CanWrite, repo.CanOpenChangeRequests)
				}
			})

			t.Run("BadTokenIsUnauthorized", func(t *testing.T) {
				c := newClientForTest(t, h.kind, srv.URL, "acme/agents", "tok-bogus")
				_, err := c.GetRepository(ctx)
				if !errors.Is(err, ErrUnauthorized) {
					t.Fatalf("err = %v, want ErrUnauthorized", err)
				}
			})

			t.Run("UnknownRepositoryIsNotFound", func(t *testing.T) {
				c := newClientForTest(t, h.kind, srv.URL, "acme/nope", writeToken)
				_, err := c.GetRepository(ctx)
				if !errors.Is(err, ErrNotFound) {
					t.Fatalf("err = %v, want ErrNotFound", err)
				}
			})

			t.Run("BranchHead", func(t *testing.T) {
				c := newClientForTest(t, h.kind, srv.URL, "acme/agents", writeToken)
				sha, err := c.GetBranchHead(ctx, "develop")
				if err != nil {
					t.Fatalf("GetBranchHead: %v", err)
				}
				if sha != "feedc0de" {
					t.Errorf("sha = %q", sha)
				}
			})

			t.Run("MissingBranchIsNotFound", func(t *testing.T) {
				c := newClientForTest(t, h.kind, srv.URL, "acme/agents", writeToken)
				_, err := c.GetBranchHead(ctx, "gone")
				if !errors.Is(err, ErrNotFound) {
					t.Fatalf("err = %v, want ErrNotFound", err)
				}
			})

			t.Run("BranchNameWithSlashIsEscaped", func(t *testing.T) {
				c := newClientForTest(t, h.kind, srv.URL, "acme/agents", writeToken)
				sha, err := c.GetBranchHead(ctx, "release/v1")
				if err != nil {
					t.Fatalf("GetBranchHead: %v", err)
				}
				if sha != "0ddba11" {
					t.Errorf("sha = %q", sha)
				}
			})

			t.Run("SelfHostedSubpathBaseURL", func(t *testing.T) {
				sub := httptest.NewServer(h.newHandler(t, "/custom/api"))
				t.Cleanup(sub.Close)
				c := newClientForTest(t, h.kind, sub.URL+"/custom/api/", "acme/agents", writeToken)
				repo, err := c.GetRepository(ctx)
				if err != nil {
					t.Fatalf("GetRepository via subpath: %v", err)
				}
				if repo.FullName != "acme/agents" {
					t.Errorf("FullName = %q", repo.FullName)
				}
			})

			t.Run("RedirectsAreNotFollowed", func(t *testing.T) {
				// A compromised or misconfigured host must not be able to
				// bounce the credential to another origin.
				var evilHits int
				evil := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					evilHits++
				}))
				t.Cleanup(evil.Close)
				redirector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Redirect(w, r, evil.URL+r.URL.Path, http.StatusFound)
				}))
				t.Cleanup(redirector.Close)
				c := newClientForTest(t, h.kind, redirector.URL, "acme/agents", writeToken)
				if _, err := c.GetRepository(ctx); err == nil {
					t.Fatal("expected error for redirecting host")
				}
				if evilHits != 0 {
					t.Fatalf("redirect was followed %d times", evilHits)
				}
			})

		t.Run("GetTree", func(t *testing.T) {
				c := newClientForTest(t, h.kind, srv.URL, "acme/agents", writeToken)
				entries, err := c.GetTree(ctx, "main", "agents")
				if err != nil {
					t.Fatalf("GetTree: %v", err)
				}
				if len(entries) == 0 {
					t.Fatal("expected non-empty tree")
				}
				hasFile := false
				hasDir := false
				for _, e := range entries {
					if e.Kind == TreeEntryFile {
						hasFile = true
					}
					if e.Kind == TreeEntryDirectory {
						hasDir = true
					}
				}
				if !hasFile {
					t.Error("expected at least one file entry")
				}
				if !hasDir {
					t.Error("expected at least one directory entry")
				}
			})

			t.Run("GetBlob", func(t *testing.T) {
				c := newClientForTest(t, h.kind, srv.URL, "acme/agents", writeToken)
				data, err := c.GetBlob(ctx, "main", "agents/my-agent/prompt.md")
				if err != nil {
					t.Fatalf("GetBlob: %v", err)
				}
				if string(data) != "You are a helpful agent." {
					t.Errorf("content = %q", string(data))
				}
			})

			t.Run("GetBlobMissingFile", func(t *testing.T) {
				c := newClientForTest(t, h.kind, srv.URL, "acme/agents", writeToken)
				_, err := c.GetBlob(ctx, "main", "agents/missing.md")
				if !errors.Is(err, ErrNotFound) {
					t.Fatalf("err = %v, want ErrNotFound", err)
				}
			})

			t.Run("CompareCommits/Ahead", func(t *testing.T) {
				c := newClientForTest(t, h.kind, srv.URL, "acme/agents", writeToken)
				cmp, err := c.CompareCommits(ctx, "abc123", "feedc0de")
				if err != nil {
					t.Fatalf("CompareCommits err = %v", err)
				}
				if cmp.Status != "ahead" {
					t.Errorf("status = %q, want ahead", cmp.Status)
				}
			})

			t.Run("CompareCommits/Identical", func(t *testing.T) {
				c := newClientForTest(t, h.kind, srv.URL, "acme/agents", writeToken)
				cmp, err := c.CompareCommits(ctx, "abc123", "abc123")
				if err != nil {
					t.Fatalf("CompareCommits err = %v", err)
				}
				if cmp.Status != "identical" {
					t.Errorf("status = %q, want identical", cmp.Status)
				}
			})

			t.Run("ErrorsNeverContainToken", func(t *testing.T) {
				c := newClientForTest(t, h.kind, srv.URL, "acme/nope", writeToken)
				_, err := c.GetRepository(ctx)
				if err == nil {
					t.Fatal("expected error")
				}
				if strings.Contains(err.Error(), writeToken) {
					t.Fatalf("error leaks token: %v", err)
				}
				cBad := newClientForTest(t, h.kind, srv.URL, "acme/agents", "tok-bogus")
				if _, err := cBad.GetRepository(ctx); err != nil && strings.Contains(err.Error(), "tok-bogus") {
					t.Fatalf("error leaks token: %v", err)
				}
			})
		})
	}
}

// ── Write operation contract tests ──────────────────────────────────────

type statefulHarness struct {
	kind       Kind
	newHandler func(t *testing.T, apiPrefix string, state *fakeState) http.Handler
}

func statefulHarnesses() []statefulHarness {
	return []statefulHarness{
		{kind: KindGitHub, newHandler: newGitHubFakeStateful},
		{kind: KindGitLab, newHandler: newGitLabFakeStateful},
	}
}

func TestProviderWriteContract(t *testing.T) {
	for _, h := range statefulHarnesses() {
		t.Run(string(h.kind), func(t *testing.T) {
			t.Run("CreateCommit", func(t *testing.T) {
				state := newFakeState()
				srv := httptest.NewServer(h.newHandler(t, "", state))
				t.Cleanup(srv.Close)
				ctx := context.Background()
				c := newClientForTest(t, h.kind, srv.URL, "acme/agents", writeToken)

				result, err := c.CreateCommit(ctx, "main", "abc123", "test commit", []FileAction{
					{Path: "agents/new/prompt.md", Content: []byte("Hello!")},
				})
				if err != nil {
					t.Fatalf("CreateCommit: %v", err)
				}
				if result.SHA == "" {
					t.Fatal("commit SHA is empty")
				}

				state.mu.Lock()
				if state.lastCommitMessage != "test commit" {
					t.Errorf("commit message = %q, want %q", state.lastCommitMessage, "test commit")
				}
				if state.branches["main"] != result.SHA {
					t.Errorf("branch main = %q, want %q", state.branches["main"], result.SHA)
				}
				state.mu.Unlock()
			})

			t.Run("CreateCommitWithDelete", func(t *testing.T) {
				state := newFakeState()
				srv := httptest.NewServer(h.newHandler(t, "", state))
				t.Cleanup(srv.Close)
				ctx := context.Background()
				c := newClientForTest(t, h.kind, srv.URL, "acme/agents", writeToken)

				result, err := c.CreateCommit(ctx, "main", "abc123", "delete file", []FileAction{
					{Path: "agents/old/prompt.md", Delete: true},
				})
				if err != nil {
					t.Fatalf("CreateCommit with delete: %v", err)
				}
				if result.SHA == "" {
					t.Fatal("commit SHA is empty")
				}
			})

			t.Run("CreateBranch", func(t *testing.T) {
				state := newFakeState()
				srv := httptest.NewServer(h.newHandler(t, "", state))
				t.Cleanup(srv.Close)
				ctx := context.Background()
				c := newClientForTest(t, h.kind, srv.URL, "acme/agents", writeToken)

				err := c.CreateBranch(ctx, "butter/test-branch", "abc123")
				if err != nil {
					t.Fatalf("CreateBranch: %v", err)
				}

				state.mu.Lock()
				sha, ok := state.branches["butter/test-branch"]
				state.mu.Unlock()
				if !ok || sha != "abc123" {
					t.Fatalf("branch not created: ok=%v sha=%q", ok, sha)
				}
			})

			t.Run("CreateChangeRequest", func(t *testing.T) {
				state := newFakeState()
				srv := httptest.NewServer(h.newHandler(t, "", state))
				t.Cleanup(srv.Close)
				ctx := context.Background()
				c := newClientForTest(t, h.kind, srv.URL, "acme/agents", writeToken)

				if err := c.CreateBranch(ctx, "butter/cr-test", "abc123"); err != nil {
					t.Fatalf("CreateBranch: %v", err)
				}

				cr, err := c.CreateChangeRequest(ctx, "butter/cr-test", "main",
					"Agent Content Update", "Automated content update from Butter")
				if err != nil {
					t.Fatalf("CreateChangeRequest: %v", err)
				}
				if cr.ID <= 0 {
					t.Fatalf("change request ID = %d", cr.ID)
				}
				if cr.URL == "" {
					t.Fatal("change request URL is empty")
				}
				if cr.Title != "Agent Content Update" {
					t.Fatalf("title = %q", cr.Title)
				}
			})

			t.Run("WriteRequiresWriteToken", func(t *testing.T) {
				state := newFakeState()
				srv := httptest.NewServer(h.newHandler(t, "", state))
				t.Cleanup(srv.Close)
				ctx := context.Background()
				c := newClientForTest(t, h.kind, srv.URL, "acme/agents", readOnlyToken)

				_, err := c.CreateCommit(ctx, "main", "abc123", "test", []FileAction{
					{Path: "test.md", Content: []byte("test")},
				})
				if err == nil {
					t.Fatal("expected error for read-only token")
				}
			})

			t.Run("ErrorsNeverContainToken", func(t *testing.T) {
				state := newFakeState()
				srv := httptest.NewServer(h.newHandler(t, "", state))
				t.Cleanup(srv.Close)
				ctx := context.Background()
				c := newClientForTest(t, h.kind, srv.URL, "acme/agents", readOnlyToken)

				_, err := c.CreateCommit(ctx, "main", "abc123", "test", []FileAction{
					{Path: "test.md", Content: []byte("test")},
				})
				if err != nil && strings.Contains(err.Error(), readOnlyToken) {
					t.Fatalf("error leaks token: %v", err)
				}
			})
		})
	}
}

func TestNewRejectsInvalidConfig(t *testing.T) {
	valid := Config{Kind: KindGitHub, APIBaseURL: "https://api.github.com", Repository: "a/b", Token: "t"}

	cfg := valid
	cfg.Kind = Kind("svn")
	if _, err := New(cfg); err == nil {
		t.Error("expected error for unknown kind")
	}

	cfg = valid
	cfg.APIBaseURL = "not a url"
	if _, err := New(cfg); err == nil {
		t.Error("expected error for invalid base URL")
	}

	cfg = valid
	cfg.Repository = "norepo"
	if _, err := New(cfg); err == nil {
		t.Error("expected error for repository without namespace")
	}

	cfg = valid
	cfg.Token = ""
	if _, err := New(cfg); err == nil {
		t.Error("expected error for empty token")
	}
}
