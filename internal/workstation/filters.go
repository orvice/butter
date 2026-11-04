package workstation

import (
	"context"
	"encoding/base64"
	"encoding/json"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"slices"
	"strings"
)

// Filter options use the same source choice and retention as List/Get. They
// never call discovery or expose a principal from a hidden observation.
func (s *Service) FilterOptions(ctx context.Context, w string, q Query, dimension, search, page string, limit int) ([]*v1.WorkStationFilterOption, string, error) {
	if !slices.Contains([]string{"provider", "scope", "project", "status", "assignee"}, dimension) || len(search) > 256 {
		return nil, "", ErrInvalid
	}
	if q.Kind == "" {
		q.Kind = "ISSUE"
	}
	if !slices.Contains([]string{"ISSUE", "PULL_REQUEST", "MERGE_REQUEST"}, q.Kind) {
		return nil, "", ErrInvalid
	}
	if limit == 0 {
		limit = 50
	}
	if limit < 1 || limit > 100 {
		return nil, "", ErrInvalid
	}
	q.Status = "all"
	q.Limit = limit
	fp := fingerprint(w, q) + "/" + dimension + "/" + search
	after := ""
	if page != "" {
		if len(page) > 4096 {
			return nil, "", ErrInvalid
		}
		b, e := base64.RawURLEncoding.DecodeString(page)
		var c struct{ Fingerprint, ID string }
		if e != nil || json.Unmarshal(b, &c) != nil || c.Fingerprint != fp {
			return nil, "", ErrInvalid
		}
		after = c.ID
	}
	var err error
	q.Now, err = s.Repo.Now(ctx)
	if err != nil {
		return nil, "", err
	}
	if err = s.allowedSources(ctx, w, &q); err != nil {
		return nil, "", err
	}
	options, err := s.Repo.Facets(ctx, w, q, dimension, search, after, limit+1)
	if err != nil {
		return nil, "", err
	}
	next := ""
	if len(options) > limit {
		options = options[:limit]
		b, _ := json.Marshal(struct{ Fingerprint, ID string }{fp, options[len(options)-1].Id})
		next = base64.RawURLEncoding.EncodeToString(b)
	}
	return options, next, nil
}

// ProjectOptions is the memory reference contract for the Mongo aggregation.
func ProjectOptions(items []*v1.WorkStationItem, dimension, search, after string, limit int) []*v1.WorkStationFilterOption {
	values := map[string]string{}
	add := func(id, name string) {
		if id != "" && id > after && (search == "" || strings.Contains(strings.ToLower(id+" "+name), strings.ToLower(search))) {
			if old, ok := values[id]; !ok || name < old {
				values[id] = name
			}
		}
	}
	for _, item := range items {
		switch dimension {
		case "provider":
			add(item.Provider, item.Provider)
		case "scope":
			add(item.ScopeId, item.ScopeName)
		case "project":
			add(item.ProjectId, item.ProjectName)
		case "status":
			add(item.Status, item.Status)
		case "assignee":
			for _, a := range item.Assignees {
				add(a, a)
			}
		}
	}
	out := []*v1.WorkStationFilterOption{}
	for id, name := range values {
		out = append(out, &v1.WorkStationFilterOption{Id: id, Name: name})
	}
	slices.SortFunc(out, func(a, b *v1.WorkStationFilterOption) int { return strings.Compare(a.Id, b.Id) })
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}
