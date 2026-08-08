package gitprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// gitlabClient speaks the GitLab REST v4 dialect. base is the full API root:
// "https://gitlab.com/api/v4" or "https://gitlab.example.com/api/v4".
type gitlabClient struct {
	base  string
	repo  string
	token string
	http  *http.Client
}

// GitLab access levels: 10 Guest, 20 Reporter, 30 Developer, 40 Maintainer,
// 50 Owner. Developer is the minimum that can push branches and open MRs.
const gitlabDeveloperAccess = 30

func (c *gitlabClient) get(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+path, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("PRIVATE-TOKEN", c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("gitlab request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return statusError(resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode gitlab response: %w", err)
	}
	return nil
}

func (c *gitlabClient) projectPath() string {
	return "/projects/" + url.PathEscape(c.repo)
}

func (c *gitlabClient) GetRepository(ctx context.Context) (*Repository, error) {
	var body struct {
		PathWithNamespace string `json:"path_with_namespace"`
		Visibility        string `json:"visibility"`
		DefaultBranch     string `json:"default_branch"`
		Permissions       struct {
			ProjectAccess *struct {
				AccessLevel int `json:"access_level"`
			} `json:"project_access"`
			GroupAccess *struct {
				AccessLevel int `json:"access_level"`
			} `json:"group_access"`
		} `json:"permissions"`
	}
	if err := c.get(ctx, c.projectPath(), &body); err != nil {
		return nil, err
	}
	level := 0
	if pa := body.Permissions.ProjectAccess; pa != nil && pa.AccessLevel > level {
		level = pa.AccessLevel
	}
	if ga := body.Permissions.GroupAccess; ga != nil && ga.AccessLevel > level {
		level = ga.AccessLevel
	}
	canWrite := level >= gitlabDeveloperAccess
	return &Repository{
		FullName:      body.PathWithNamespace,
		Private:       body.Visibility != "public",
		DefaultBranch: body.DefaultBranch,
		CanRead:       true,
		CanWrite:      canWrite,
		// MRs are opened from branches pushed to the same project, which
		// needs Developer access — the same threshold as writes.
		CanOpenChangeRequests: canWrite,
	}, nil
}

func (c *gitlabClient) GetBranchHead(ctx context.Context, branch string) (string, error) {
	var body struct {
		Commit struct {
			ID string `json:"id"`
		} `json:"commit"`
	}
	if err := c.get(ctx, c.projectPath()+"/repository/branches/"+url.PathEscape(branch), &body); err != nil {
		return "", err
	}
	return body.Commit.ID, nil
}
