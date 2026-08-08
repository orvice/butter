package gitprovider

import (
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
