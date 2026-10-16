package main

// Deterministic git provider double for repository binding contract tests.
// Copied from internal/application/repobinding_service_test.go (fakeProviderClient).

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"go.orx.me/apps/butter/internal/gitprovider"
)

type fakeGitClient struct {
	repo        *gitprovider.Repository
	repoErr     error
	branches    map[string]string
	branchErr   error
	trees       map[string][]gitprovider.TreeEntry
	blobs       map[string][]byte
	materialize bool
	commitErr   error
	mu          sync.Mutex
	commits     []fakeCommit
}

type fakeCommit struct {
	branch, parentSHA, message, sha string
	actions                         []gitprovider.FileAction
}

func newFakeGitClient() *fakeGitClient {
	return &fakeGitClient{
		repo: &gitprovider.Repository{
			FullName: "acme/agents", Private: true, DefaultBranch: "main",
			CanRead: true, CanWrite: true, CanOpenChangeRequests: true,
		},
		branches: map[string]string{"main": "abc123"},
		trees:    map[string][]gitprovider.TreeEntry{},
		blobs:    map[string][]byte{},
	}
}

func (f *fakeGitClient) GetRepository(context.Context) (*gitprovider.Repository, error) {
	if f.repoErr != nil {
		return nil, f.repoErr
	}
	return f.repo, nil
}

func (f *fakeGitClient) GetBranchHead(_ context.Context, branch string) (string, error) {
	if f.branchErr != nil {
		return "", f.branchErr
	}
	sha, ok := f.branches[branch]
	if !ok {
		return "", gitprovider.ErrNotFound
	}
	return sha, nil
}

func (f *fakeGitClient) GetTree(_ context.Context, ref, path string) ([]gitprovider.TreeEntry, error) {
	key := ref + ":" + path
	entries, ok := f.trees[key]
	if !ok {
		return nil, gitprovider.ErrNotFound
	}
	return entries, nil
}

func (f *fakeGitClient) GetBlob(_ context.Context, ref, path string) ([]byte, error) {
	key := ref + ":" + path
	data, ok := f.blobs[key]
	if !ok {
		return nil, gitprovider.ErrNotFound
	}
	return data, nil
}

func (f *fakeGitClient) CompareCommits(_ context.Context, base, head string) (*gitprovider.CommitComparison, error) {
	if base == head {
		return &gitprovider.CommitComparison{Status: "identical"}, nil
	}
	return &gitprovider.CommitComparison{Status: "ahead"}, nil
}

func (f *fakeGitClient) CreateCommit(_ context.Context, branch, parentSHA, message string, actions []gitprovider.FileAction) (*gitprovider.CommitResult, error) {
	if f.commitErr != nil {
		return nil, f.commitErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	sha := fmt.Sprintf("commit-%d", len(f.commits)+1)
	f.commits = append(f.commits, fakeCommit{branch: branch, parentSHA: parentSHA, message: message, actions: actions, sha: sha})
	f.branches[branch] = sha
	if f.materialize {
		f.materializeCommitLocked(parentSHA, sha, actions)
	}
	return &gitprovider.CommitResult{SHA: sha}, nil
}

func (f *fakeGitClient) materializeCommitLocked(parentSHA, newSHA string, actions []gitprovider.FileAction) {
	for k, v := range f.trees {
		if suffix, ok := strings.CutPrefix(k, parentSHA+":"); ok {
			f.trees[newSHA+":"+suffix] = append([]gitprovider.TreeEntry(nil), v...)
		}
	}
	for k, v := range f.blobs {
		if suffix, ok := strings.CutPrefix(k, parentSHA+":"); ok {
			f.blobs[newSHA+":"+suffix] = v
		}
	}
	rootKey := newSHA + ":"
	entries := f.trees[rootKey]
	for _, a := range actions {
		blobKey := newSHA + ":" + a.Path
		if a.Delete {
			delete(f.blobs, blobKey)
			continue
		}
		f.blobs[blobKey] = a.Content
		entries = upsertTreeEntry(entries, a.Path)
	}
	f.trees[rootKey] = entries
}

func upsertTreeEntry(entries []gitprovider.TreeEntry, path string) []gitprovider.TreeEntry {
	for _, e := range entries {
		if e.Path == path {
			return entries
		}
	}
	return append(entries, gitprovider.TreeEntry{Path: path, Kind: gitprovider.TreeEntryFile, Size: 1, SHA: "blob-" + path})
}

func (f *fakeGitClient) CreateBranch(context.Context, string, string) error { return nil }
func (f *fakeGitClient) DeleteBranch(context.Context, string) error         { return nil }
func (f *fakeGitClient) CreateChangeRequest(context.Context, string, string, string, string) (*gitprovider.ChangeRequestResult, error) {
	return &gitprovider.ChangeRequestResult{ID: 1, URL: "https://example.invalid/pr/1", Title: "t"}, nil
}
