package gitprovider

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
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
	// Bearer auth (supported by GitLab for PATs) rather than PRIVATE-TOKEN:
	// Go's http.Client strips Authorization on cross-origin redirects but
	// forwards custom headers, so PRIVATE-TOKEN could leak the credential.
	req.Header.Set("Authorization", "Bearer "+c.token)
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

func (c *gitlabClient) GetTree(ctx context.Context, ref, path string) ([]TreeEntry, error) {
	q := url.Values{}
	q.Set("ref", ref)
	q.Set("recursive", "true")
	q.Set("per_page", "100")
	if path != "" {
		q.Set("path", path)
	}
	var all []TreeEntry
	page := 1
	for {
		q.Set("page", fmt.Sprintf("%d", page))
		var items []struct {
			Path string `json:"path"`
			Type string `json:"type"`
			Mode string `json:"mode"`
			ID   string `json:"id"`
		}
		if err := c.get(ctx, c.projectPath()+"/repository/tree?"+q.Encode(), &items); err != nil {
			return nil, err
		}
		for _, item := range items {
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
			all = append(all, TreeEntry{
				Path: item.Path,
				Kind: kind,
				SHA:  item.ID,
			})
		}
		if len(items) < 100 {
			break
		}
		page++
	}
	return all, nil
}

func (c *gitlabClient) GetBlob(ctx context.Context, ref, path string) ([]byte, error) {
	encodedPath := url.PathEscape(path)
	q := url.Values{}
	q.Set("ref", ref)
	var body struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
		Size     int64  `json:"size"`
	}
	apiPath := c.projectPath() + "/repository/files/" + encodedPath + "?" + q.Encode()
	if err := c.get(ctx, apiPath, &body); err != nil {
		return nil, err
	}
	if body.Encoding == "base64" {
		return base64.StdEncoding.DecodeString(strings.ReplaceAll(body.Content, "\n", ""))
	}
	return []byte(body.Content), nil
}
