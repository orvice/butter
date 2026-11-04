package application

import (
	"context"
	githost "go.orx.me/apps/butter/internal/repo/githost"
	ws "go.orx.me/apps/butter/internal/workstation"
	"go.orx.me/apps/butter/internal/workstation/github"
	"go.orx.me/apps/butter/internal/workstation/gitlab"
	"go.orx.me/apps/butter/internal/workstation/linear"
	"go.orx.me/apps/butter/internal/workstation/outbound"
	"go.orx.me/apps/butter/internal/workstation/plane"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"net/http"
	"strings"
	"sync"
)

// WorkStationProviders resolves only approved profiles and rechecks the host
// registry on every read/attempt. Adapter caches are bounded and epoch scoped.
func WorkStationProviders(gitHosts githost.Repository, repo ws.Repository) func(context.Context, ws.Record) (ws.ResolvedProvider, error) {
	var mu sync.Mutex
	cache := map[string]ws.ResolvedProvider{}
	return func(ctx context.Context, r ws.Record) (ws.ResolvedProvider, error) {
		provider := r.SourceIdentity().Provider
		api, web := "", ""
		switch provider {
		case "linear":
			if r.HostID != "" && r.HostID != "linear" {
				return ws.ResolvedProvider{}, ws.ErrInvalid
			}
			if r.ExternalWorkspaceSlug != "" {
				return ws.ResolvedProvider{}, ws.ErrInvalid
			}
			api, web = "https://api.linear.app", "https://linear.app"
		case "plane":
			h, err := repo.GetHost(ctx, r.HostID)
			if err != nil || h.Profile != "plane-v2" {
				return ws.ResolvedProvider{}, ws.ErrUnsupported
			}
			if err := ws.ValidateHost(&h); err != nil {
				return ws.ResolvedProvider{}, err
			}
			api, web = h.APIBaseURL, h.WebBaseURL
		case "github", "gitlab":
			if gitHosts == nil {
				return ws.ResolvedProvider{}, ws.ErrUnsupported
			}
			h, err := gitHosts.Get(ctx, r.HostID)
			if err != nil {
				return ws.ResolvedProvider{}, ws.ErrUnsupported
			}
			if r.ExternalWorkspaceSlug != "" {
				return ws.ResolvedProvider{}, ws.ErrInvalid
			}
			api, err = ws.ApprovedURL(h.ApiBaseUrl)
			if err != nil {
				return ws.ResolvedProvider{}, err
			}
			web, err = ws.ApprovedURL(h.WebBaseUrl)
			if err != nil {
				return ws.ResolvedProvider{}, err
			}
			if provider == "github" && (h.Kind != v1.GitHostKind_GIT_HOST_KIND_GITHUB || api != "https://api.github.com" || web != "https://github.com") {
				return ws.ResolvedProvider{}, ws.ErrUnsupported
			}
			if provider == "gitlab" && h.Kind != v1.GitHostKind_GIT_HOST_KIND_GITLAB {
				return ws.ResolvedProvider{}, ws.ErrUnsupported
			}
		default:
			return ws.ResolvedProvider{}, ws.ErrUnsupported
		}
		// Even a cached adapter is discarded on endpoint mutation. resolve in the
		// domain also compares this source with the connection's pinned identity.
		key := strings.Join([]string{r.WorkspaceID, r.ID, r.Epoch, r.PrincipalID, provider, api, web, r.ExternalWorkspaceSlug}, "\x00")
		mu.Lock()
		cached, ok := cache[key]
		mu.Unlock()
		if ok {
			return cached, nil
		}
		client, err := outbound.NewClient(api, nil)
		if err != nil {
			return ws.ResolvedProvider{}, ws.ErrUnsupported
		}
		budgetRecord := r
		budgetRecord.Source = ws.SourceIdentity{Provider: provider, Origin: api}
		var transport http.RoundTripper = ws.BudgetTransport(repo, budgetRecord, client.Transport)
		resolved := ws.ResolvedProvider{Source: budgetRecord.Source, WebBaseURL: web}
		switch provider {
		case "linear":
			resolved.Adapter = linear.New(transport)
		case "github":
			resolved.Adapter = github.New(transport)
		case "gitlab":
			resolved.Adapter, err = gitlab.New(api, web, transport)
		case "plane":
			resolved.Adapter, err = plane.New(api, web, r.ExternalWorkspaceSlug, transport)
		}
		if err != nil {
			return ws.ResolvedProvider{}, ws.ErrUnsupported
		}
		mu.Lock()
		if len(cache) >= 256 {
			clear(cache)
		}
		cache[key] = resolved
		mu.Unlock()
		return resolved, nil
	}
}
