package gitprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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
