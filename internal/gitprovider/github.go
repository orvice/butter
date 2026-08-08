package gitprovider

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// githubClient speaks the GitHub REST v3 dialect. base is the full API root:
// "https://api.github.com" or "https://ghe.example.com/api/v3".
type githubClient struct {
	base  string
	repo  string
	token string
	http  *http.Client
}

func (c *githubClient) get(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+path, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("github request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return statusError(resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode github response: %w", err)
	}
	return nil
}

func (c *githubClient) GetRepository(ctx context.Context) (*Repository, error) {
	var body struct {
		FullName      string `json:"full_name"`
		Private       bool   `json:"private"`
		DefaultBranch string `json:"default_branch"`
		Permissions   struct {
			Admin    bool `json:"admin"`
			Maintain bool `json:"maintain"`
			Push     bool `json:"push"`
			Pull     bool `json:"pull"`
		} `json:"permissions"`
	}
	if err := c.get(ctx, "/repos/"+c.repo, &body); err != nil {
		return nil, err
	}
	canWrite := body.Permissions.Push || body.Permissions.Maintain || body.Permissions.Admin
	return &Repository{
		FullName:      body.FullName,
		Private:       body.Private,
		DefaultBranch: body.DefaultBranch,
		CanRead:       true,
		CanWrite:      canWrite,
		// Butter opens PRs from branches pushed to the same repository, so
		// change-request capability requires push access too.
		CanOpenChangeRequests: canWrite,
	}, nil
}

func (c *githubClient) GetBranchHead(ctx context.Context, branch string) (string, error) {
	var body struct {
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}
	if err := c.get(ctx, "/repos/"+c.repo+"/branches/"+url.PathEscape(branch), &body); err != nil {
		return "", err
	}
	return body.Commit.SHA, nil
}

func (c *githubClient) GetTree(ctx context.Context, ref, path string) ([]TreeEntry, error) {
	treeSHA, err := c.resolveTreeSHA(ctx, ref, path)
	if err != nil {
		return nil, err
	}
	var body struct {
		Tree []struct {
			Path string `json:"path"`
			Mode string `json:"mode"`
			Type string `json:"type"`
			Size int64  `json:"size"`
			SHA  string `json:"sha"`
		} `json:"tree"`
		Truncated bool `json:"truncated"`
	}
	if err := c.get(ctx, "/repos/"+c.repo+"/git/trees/"+treeSHA+"?recursive=1", &body); err != nil {
		return nil, err
	}
	prefix := strings.TrimRight(path, "/")
	entries := make([]TreeEntry, 0, len(body.Tree))
	for _, item := range body.Tree {
		kind := TreeEntryFile
		switch item.Type {
		case "tree":
			kind = TreeEntryDirectory
		case "commit":
			kind = TreeEntrySubmodule
		case "blob":
			if item.Mode == "120000" {
				kind = TreeEntrySymlink
			}
		}
		entryPath := item.Path
		if prefix != "" {
			entryPath = prefix + "/" + item.Path
		}
		entries = append(entries, TreeEntry{
			Path: entryPath,
			Kind: kind,
			Size: item.Size,
			SHA:  item.SHA,
		})
	}
	return entries, nil
}

// resolveTreeSHA finds the tree SHA for the given ref and optional subdirectory.
func (c *githubClient) resolveTreeSHA(ctx context.Context, ref, path string) (string, error) {
	var commit struct {
		SHA    string `json:"sha"`
		Object struct {
			SHA string `json:"sha"`
		} `json:"object"`
	}
	if err := c.get(ctx, "/repos/"+c.repo+"/git/ref/heads/"+url.PathEscape(ref), &commit); err != nil {
		// Try as a raw commit SHA.
		var commitObj struct {
			Tree struct {
				SHA string `json:"sha"`
			} `json:"tree"`
		}
		if err2 := c.get(ctx, "/repos/"+c.repo+"/git/commits/"+url.PathEscape(ref), &commitObj); err2 != nil {
			return "", err
		}
		if path == "" {
			return commitObj.Tree.SHA, nil
		}
		return c.walkTree(ctx, commitObj.Tree.SHA, path)
	}
	var commitObj struct {
		Tree struct {
			SHA string `json:"sha"`
		} `json:"tree"`
	}
	if err := c.get(ctx, "/repos/"+c.repo+"/git/commits/"+commit.Object.SHA, &commitObj); err != nil {
		return "", err
	}
	if path == "" {
		return commitObj.Tree.SHA, nil
	}
	return c.walkTree(ctx, commitObj.Tree.SHA, path)
}

func (c *githubClient) walkTree(ctx context.Context, rootTreeSHA, path string) (string, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	current := rootTreeSHA
	for _, part := range parts {
		var tree struct {
			Tree []struct {
				Path string `json:"path"`
				Type string `json:"type"`
				SHA  string `json:"sha"`
			} `json:"tree"`
		}
		if err := c.get(ctx, "/repos/"+c.repo+"/git/trees/"+current, &tree); err != nil {
			return "", err
		}
		found := false
		for _, entry := range tree.Tree {
			if entry.Path == part && entry.Type == "tree" {
				current = entry.SHA
				found = true
				break
			}
		}
		if !found {
			return "", ErrNotFound
		}
	}
	return current, nil
}

func (c *githubClient) CompareCommits(ctx context.Context, base, head string) (*CommitComparison, error) {
	var body struct {
		Status string `json:"status"`
	}
	if err := c.get(ctx, "/repos/"+c.repo+"/compare/"+url.PathEscape(base)+"..."+url.PathEscape(head), &body); err != nil {
		return nil, err
	}
	return &CommitComparison{Status: body.Status}, nil
}

func (c *githubClient) post(ctx context.Context, path string, body any, out any) (int, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return 0, fmt.Errorf("marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+path, bytes.NewReader(payload))
	if err != nil {
		return 0, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, fmt.Errorf("github request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.StatusCode, statusError(resp.StatusCode)
	}
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return resp.StatusCode, fmt.Errorf("decode github response: %w", err)
		}
	}
	return resp.StatusCode, nil
}

func (c *githubClient) patch(ctx context.Context, path string, body any) (int, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return 0, fmt.Errorf("marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.base+path, bytes.NewReader(payload))
	if err != nil {
		return 0, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, fmt.Errorf("github request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.StatusCode, statusError(resp.StatusCode)
	}
	return resp.StatusCode, nil
}

func (c *githubClient) CreateCommit(ctx context.Context, branch, parentSHA, message string, actions []FileAction) (*CommitResult, error) {
	// Build tree entries. For PUTs, create blobs first; for DELETEs, set sha to nil.
	type treeEntry struct {
		Path string      `json:"path"`
		Mode string      `json:"mode"`
		Type string      `json:"type"`
		SHA  interface{} `json:"sha"` // string for blobs, nil for deletes
	}
	var entries []treeEntry
	for _, a := range actions {
		if a.Delete {
			entries = append(entries, treeEntry{
				Path: a.Path, Mode: "100644", Type: "blob", SHA: nil,
			})
			continue
		}
		var blob struct {
			SHA string `json:"sha"`
		}
		if _, err := c.post(ctx, "/repos/"+c.repo+"/git/blobs", map[string]string{
			"content":  base64.StdEncoding.EncodeToString(a.Content),
			"encoding": "base64",
		}, &blob); err != nil {
			return nil, fmt.Errorf("create blob for %s: %w", a.Path, err)
		}
		entries = append(entries, treeEntry{
			Path: a.Path, Mode: "100644", Type: "blob", SHA: blob.SHA,
		})
	}

	// Get the tree SHA of the parent commit.
	var parentCommit struct {
		Tree struct {
			SHA string `json:"sha"`
		} `json:"tree"`
	}
	if err := c.get(ctx, "/repos/"+c.repo+"/git/commits/"+parentSHA, &parentCommit); err != nil {
		return nil, fmt.Errorf("get parent commit: %w", err)
	}

	// Create the new tree with base_tree for incremental update.
	var newTree struct {
		SHA string `json:"sha"`
	}
	if _, err := c.post(ctx, "/repos/"+c.repo+"/git/trees", map[string]interface{}{
		"base_tree": parentCommit.Tree.SHA,
		"tree":      entries,
	}, &newTree); err != nil {
		return nil, fmt.Errorf("create tree: %w", err)
	}

	// Create the commit.
	var commitObj struct {
		SHA string `json:"sha"`
	}
	if _, err := c.post(ctx, "/repos/"+c.repo+"/git/commits", map[string]interface{}{
		"message": message,
		"tree":    newTree.SHA,
		"parents": []string{parentSHA},
	}, &commitObj); err != nil {
		return nil, fmt.Errorf("create commit: %w", err)
	}

	// Update the branch ref. GitHub returns 422 if the update is not fast-forward.
	status, err := c.patch(ctx, "/repos/"+c.repo+"/git/refs/heads/"+url.PathEscape(branch), map[string]interface{}{
		"sha":   commitObj.SHA,
		"force": false,
	})
	if err != nil {
		if status == http.StatusUnprocessableEntity {
			return nil, ErrConflict
		}
		return nil, fmt.Errorf("update ref: %w", err)
	}
	return &CommitResult{SHA: commitObj.SHA}, nil
}

func (c *githubClient) CreateBranch(ctx context.Context, branch, sha string) error {
	_, err := c.post(ctx, "/repos/"+c.repo+"/git/refs", map[string]string{
		"ref": "refs/heads/" + branch,
		"sha": sha,
	}, nil)
	if err != nil {
		return fmt.Errorf("create branch: %w", err)
	}
	return nil
}

func (c *githubClient) CreateChangeRequest(ctx context.Context, source, target, title, description string) (*ChangeRequestResult, error) {
	var pr struct {
		Number  int    `json:"number"`
		HTMLURL string `json:"html_url"`
		Title   string `json:"title"`
	}
	if _, err := c.post(ctx, "/repos/"+c.repo+"/pulls", map[string]string{
		"title": title,
		"body":  description,
		"head":  source,
		"base":  target,
	}, &pr); err != nil {
		return nil, fmt.Errorf("create pull request: %w", err)
	}
	return &ChangeRequestResult{ID: pr.Number, URL: pr.HTMLURL, Title: pr.Title}, nil
}

func (c *githubClient) GetBlob(ctx context.Context, ref, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.base+"/repos/"+c.repo+"/contents/"+path+"?ref="+url.QueryEscape(ref), nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github.raw+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, statusError(resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if strings.Contains(ct, "application/json") {
		var body struct {
			Content  string `json:"content"`
			Encoding string `json:"encoding"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			return nil, fmt.Errorf("decode github blob: %w", err)
		}
		if body.Encoding == "base64" {
			return base64.StdEncoding.DecodeString(strings.ReplaceAll(body.Content, "\n", ""))
		}
		return []byte(body.Content), nil
	}
	return io.ReadAll(resp.Body)
}
