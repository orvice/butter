// Package gitprovider is the provider-neutral seam Butter uses to talk to
// Git hosting APIs (issue #214). Adapters speak the host's REST dialect over
// plain HTTP — no durable local clone and no git CLI. Phase 1 covers the
// operations binding validation needs (repository metadata with effective
// capabilities, branch heads); later phases extend this contract with tree
// and blob reads, multi-file commits, change requests, and ancestry checks.
//
// Error values returned by adapters never contain credential material.
package gitprovider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Kind selects the provider adapter dialect.
type Kind string

const (
	// KindGitHub covers github.com and GitHub Enterprise (REST v3).
	KindGitHub Kind = "github"
	// KindGitLab covers gitlab.com and self-hosted GitLab (REST v4).
	KindGitLab Kind = "gitlab"
)

// Sentinel errors adapters map provider HTTP failures onto.
var (
	// ErrUnauthorized means the token was rejected (expired/revoked/invalid).
	ErrUnauthorized = errors.New("git provider: unauthorized")
	// ErrForbidden means the token is valid but lacks access.
	ErrForbidden = errors.New("git provider: forbidden")
	// ErrNotFound means the repository or ref does not exist or is invisible
	// to the token (providers deliberately blur those cases).
	ErrNotFound = errors.New("git provider: not found")
)

// Repository is the provider-neutral view of a repository plus the effective
// capabilities of the authenticating token.
type Repository struct {
	// FullName is the host-relative path, e.g. "owner/repo".
	FullName string
	// Private is true unless the repository is publicly visible.
	Private bool
	// DefaultBranch is the repository's default branch name.
	DefaultBranch string
	// CanRead reports read access (always true when the fetch succeeded).
	CanRead bool
	// CanWrite reports push access to the repository.
	CanWrite bool
	// CanOpenChangeRequests reports the ability to open a Pull Request
	// (GitHub) or Merge Request (GitLab) backed by a branch in this
	// repository.
	CanOpenChangeRequests bool
}

// TreeEntry is one node in a repository tree listing.
type TreeEntry struct {
	Path string
	Kind TreeEntryKind
	Size int64
	SHA  string
}

// TreeEntryKind classifies a tree entry.
type TreeEntryKind int

const (
	TreeEntryFile      TreeEntryKind = iota
	TreeEntryDirectory TreeEntryKind = iota
	TreeEntrySymlink   TreeEntryKind = iota
	TreeEntrySubmodule TreeEntryKind = iota
)

// Client is the per-binding handle onto one repository at one host.
type Client interface {
	// GetRepository fetches repository metadata and effective capabilities.
	GetRepository(ctx context.Context) (*Repository, error)
	// GetBranchHead returns the head commit SHA of the named branch.
	GetBranchHead(ctx context.Context, branch string) (string, error)
	// GetTree returns a recursive listing of the tree at the given ref and
	// path. Entries are relative to the requested path. An empty path means
	// the repository root.
	GetTree(ctx context.Context, ref, path string) ([]TreeEntry, error)
	// GetBlob returns the raw content of the file at the given ref and path.
	GetBlob(ctx context.Context, ref, path string) ([]byte, error)
}

// Config assembles a Client. Token is held in memory only for the lifetime
// of the client and never appears in errors.
type Config struct {
	Kind       Kind
	APIBaseURL string
	// Repository is the host-relative path, e.g. "owner/repo" (GitHub) or
	// "group[/subgroup]/project" (GitLab).
	Repository string
	Token      string
	// HTTPClient overrides the default client (10s timeout) when set.
	HTTPClient *http.Client
}

// New builds the adapter for the configured provider kind.
func New(cfg Config) (Client, error) {
	base := strings.TrimRight(strings.TrimSpace(cfg.APIBaseURL), "/")
	u, err := url.Parse(base)
	if err != nil || u.Scheme != "http" && u.Scheme != "https" || u.Host == "" {
		return nil, fmt.Errorf("invalid api base url %q", cfg.APIBaseURL)
	}
	repo := strings.Trim(strings.TrimSpace(cfg.Repository), "/")
	if !strings.Contains(repo, "/") {
		return nil, fmt.Errorf("repository %q must include its namespace (owner/repo)", repo)
	}
	if strings.TrimSpace(cfg.Token) == "" {
		return nil, errors.New("token is required")
	}
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 10 * time.Second,
			// API roots never redirect legitimately. Refusing to follow
			// keeps the credential from ever being replayed to another
			// origin (Go strips Authorization on cross-origin redirects,
			// but a redirect out of a configured host is wrong regardless).
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}
	switch cfg.Kind {
	case KindGitHub:
		return &githubClient{base: base, repo: repo, token: cfg.Token, http: httpClient}, nil
	case KindGitLab:
		return &gitlabClient{base: base, repo: repo, token: cfg.Token, http: httpClient}, nil
	default:
		return nil, fmt.Errorf("unsupported git provider kind %q", cfg.Kind)
	}
}

// statusError maps a provider HTTP status onto the sentinel taxonomy. The
// response body is intentionally dropped: provider messages are not needed
// for the validation UX and must never risk echoing credentials.
func statusError(status int) error {
	switch status {
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusNotFound:
		return ErrNotFound
	default:
		return fmt.Errorf("git provider: unexpected status %d", status)
	}
}
