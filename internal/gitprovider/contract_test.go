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
