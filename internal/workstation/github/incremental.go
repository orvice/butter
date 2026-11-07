package github

import (
	"context"
	"errors"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	ws "go.orx.me/apps/butter/internal/workstation"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
)

var errItemMissing = errors.New("upstream item is no longer in this scope")
var _ ws.IncrementalProvider = (*Provider)(nil)

const overlap = 5 * time.Minute
const maxSuspectChecks = 100

func repositoryPath(name string) (string, error) {
	parts := strings.Split(name, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" || strings.ContainsAny(name, "?#%\\") || parts[0] == ".." || parts[1] == ".." || parts[0] == "." || parts[1] == "." {
		return "", failure("SCHEMA_ERROR")
	}
	return "/repos/" + url.PathEscape(parts[0]) + "/" + url.PathEscape(parts[1]), nil
}

func (p *Provider) itemPages(ctx context.Context, token, path string, query url.Values, stopBefore time.Time) ([]remoteItem, error) {
	var result []remoteItem
	seen := map[string]bool{}
	var lastUpdated time.Time
	for page := 1; ; page++ {
		query.Set("page", strconv.Itoa(page))
		var rows []remoteItem
		h, err := p.get(ctx, token, path, query, &rows)
		if err != nil {
			return nil, err
		}
		if rows == nil {
			return nil, failure("SCHEMA_ERROR")
		}
		signature := ""
		pastOverlap := false
		for _, row := range rows {
			if !decimal(row.Number.String()) {
				return nil, failure("SCHEMA_ERROR")
			}
			signature += row.Number.String() + ","
			if !stopBefore.IsZero() {
				if row.Updated.IsZero() || (!lastUpdated.IsZero() && row.Updated.After(lastUpdated)) {
					return nil, failure("SCHEMA_ERROR")
				}
				lastUpdated = row.Updated
				if row.Updated.Before(stopBefore) {
					pastOverlap = true
					continue
				}
			}
			result = append(result, row)
		}
		if len(rows) > 0 && seen[signature] {
			return nil, failure("SCHEMA_ERROR")
		}
		seen[signature] = true
		next, err := nextPage(h, path, query, page)
		if err != nil {
			return nil, err
		}
		if next == "" || pastOverlap {
			return result, nil
		}
		if len(rows) == 0 {
			return nil, failure("SCHEMA_ERROR")
		}
	}
}

// ScanIncremental merges changes into the previous source-qualified snapshot.
// The checkpoint is the previous successful scan's START time, with overlap.
// Every open item is reconciled separately; a first-page 304 is not proof that
// subsequent pages or another endpoint are unchanged.
func (p *Provider) ScanIncremental(ctx context.Context, token string, scope ws.Scope, now time.Time, previous []*v1.WorkStationItem) (ws.Scan, error) {
	if scope.Checkpoint.IsZero() || scope.Checkpoint.After(now) || now.Sub(scope.LastFull) >= 24*time.Hour || scope.Coverage != "COMPLETE" || scope.Unresolved > 0 {
		return p.Scan(ctx, token, scope, now)
	}
	repo, err := p.repository(ctx, token, scope.ID)
	if err != nil {
		return ws.Scan{}, err
	}
	if repo.HasIssues == nil {
		return ws.Scan{}, failure("SCHEMA_ERROR")
	}
	if !*repo.HasIssues {
		return ws.Scan{}, failure("FEATURE_UNAVAILABLE")
	}
	scope.Name = repo.FullName
	base, err := repositoryPath(scope.Name)
	if err != nil {
		return ws.Scan{}, err
	}
	items := map[string]*v1.WorkStationItem{}
	priorOpen := map[string]*v1.WorkStationItem{}
	for _, item := range previous {
		if item == nil || item.Provider != "github" || item.ScopeId != scope.ID || (item.Kind != "ISSUE" && item.Kind != "PULL_REQUEST") || !decimal(item.RemoteId) || item.SourceUpdatedAt == nil {
			return ws.Scan{}, ws.ErrInvalid
		}
		if !ws.Retained(item, now) {
			continue
		}
		copy := proto.Clone(item).(*v1.WorkStationItem)
		copy.ScopeName = scope.Name
		key := item.Kind + ":" + item.RemoteId
		items[key] = copy
		if !item.Terminal {
			priorOpen[key] = copy
		}
	}
	cutoff := scope.Checkpoint.UTC().Add(-overlap)
	changed := map[string]bool{}
	open := map[string]bool{}
	merge := func(rows []remoteItem, kind string, allOpen bool) error {
		for _, row := range rows {
			if kind == "ISSUE" && len(row.PullRequest) > 0 && string(row.PullRequest) != "null" {
				continue
			}
			item, err := normalize(row, scope, kind)
			if err != nil {
				return err
			}
			if allOpen && (row.State != "open" || item.Terminal) {
				return failure("SCHEMA_ERROR")
			}
			key := kind + ":" + item.RemoteId
			if old := items[key]; old == nil || !old.SourceUpdatedAt.AsTime().After(item.SourceUpdatedAt.AsTime()) {
				items[key] = item
			}
			changed[key] = true
			if allOpen {
				open[key] = true
			}
		}
		return nil
	}
	// Issues provide a native updated-since parameter. Pull requests require
	// ordered traversal until the overlapping updated-time boundary is crossed.
	issuesQuery := url.Values{"state": {"all"}, "sort": {"updated"}, "direction": {"desc"}, "per_page": {"100"}, "since": {cutoff.Format(time.RFC3339)}}
	rows, err := p.itemPages(ctx, token, base+"/issues", issuesQuery, time.Time{})
	if err != nil {
		return ws.Scan{}, err
	}
	if err = merge(rows, "ISSUE", false); err != nil {
		return ws.Scan{}, err
	}
	pullsQuery := url.Values{"state": {"all"}, "sort": {"updated"}, "direction": {"desc"}, "per_page": {"100"}}
	rows, err = p.itemPages(ctx, token, base+"/pulls", pullsQuery, cutoff)
	if err != nil {
		return ws.Scan{}, err
	}
	if err = merge(rows, "PULL_REQUEST", false); err != nil {
		return ws.Scan{}, err
	}
	for _, endpoint := range []struct{ path, kind string }{{"/issues", "ISSUE"}, {"/pulls", "PULL_REQUEST"}} {
		query := url.Values{"state": {"open"}, "sort": {"updated"}, "direction": {"desc"}, "per_page": {"100"}}
		rows, err = p.itemPages(ctx, token, base+endpoint.path, query, time.Time{})
		if err != nil {
			return ws.Scan{}, err
		}
		if err = merge(rows, endpoint.kind, true); err != nil {
			return ws.Scan{}, err
		}
	}
	// A previous open item absent from the complete open inventory is suspect.
	// Updates can cross page boundaries, so re-read it rather than preserving an
	// old open observation. Large discrepancy sets use a complete scan instead.
	var suspects []string
	for key := range priorOpen {
		if !open[key] && !changed[key] {
			suspects = append(suspects, key)
		}
	}
	if len(suspects) > maxSuspectChecks {
		return p.Scan(ctx, token, scope, now)
	}
	sort.Strings(suspects)
	for _, key := range suspects {
		old := priorOpen[key]
		endpoint := "/issues/"
		if old.Kind == "PULL_REQUEST" {
			endpoint = "/pulls/"
		}
		var remote remoteItem
		_, err = p.getResource(ctx, token, base+endpoint+old.RemoteId, nil, &remote, true)
		if errors.Is(err, errItemMissing) {
			delete(items, key)
			continue
		}
		if err != nil {
			return ws.Scan{}, err
		}
		if remote.Number.String() != old.RemoteId || (old.Kind == "ISSUE" && len(remote.PullRequest) > 0 && string(remote.PullRequest) != "null") {
			return ws.Scan{}, failure("SCHEMA_ERROR")
		}
		item, err := normalize(remote, scope, old.Kind)
		if err != nil {
			return ws.Scan{}, err
		}
		items[key] = item
	}
	current, err := p.repository(ctx, token, scope.ID)
	if err != nil {
		return ws.Scan{}, err
	}
	if current.FullName != scope.Name {
		return ws.Scan{}, failure("SCHEMA_ERROR")
	}
	result := ws.Scan{Coverage: "COMPLETE", Mode: "INCREMENTAL"}
	for _, item := range items {
		if item.Terminal && item.EndedAt == nil {
			result.Coverage = "PARTIAL"
			result.Unresolved++
			continue
		}
		if ws.Retained(item, now) {
			result.Items = append(result.Items, item)
		}
	}
	sort.Slice(result.Items, func(i, j int) bool {
		return result.Items[i].Kind+":"+result.Items[i].RemoteId < result.Items[j].Kind+":"+result.Items[j].RemoteId
	})
	return result, nil
}
