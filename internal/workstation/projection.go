package workstation

import (
	"cmp"
	"slices"
	"strings"
	"time"

	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
)

func Retained(item *v1.WorkStationItem, now time.Time) bool {
	return !item.Terminal || (item.EndedAt != nil && !item.EndedAt.AsTime().Before(now.Add(-90*24*time.Hour)))
}
func visible(r Record, o Observation) bool {
	if r.ID != o.ConnectionID || r.WorkspaceID != o.WorkspaceID || r.Paused || r.Credential.Ciphertext == "" {
		return false
	}
	for _, s := range r.Scopes {
		if s.ID == o.ScopeID && !s.Hidden && s.Snapshot != "" && s.Snapshot == o.Snapshot {
			return true
		}
	}
	return false
}

// Project is also the memory backend's reference implementation of the Mongo
// aggregation: restrict sources, choose one whole observation, filter, keyset.
func Project(records []Record, observations []Observation, q Query) []*v1.WorkStationItem {
	groups := map[string]*v1.WorkStationItem{}
	// Stable tie break: observed_at DESC, connection_id/scope_id DESC.
	slices.SortFunc(observations, func(a, b Observation) int {
		if c := b.Item.SourceUpdatedAt.AsTime().Compare(a.Item.SourceUpdatedAt.AsTime()); c != 0 {
			return c
		}
		if c := b.Item.ObservedAt.AsTime().Compare(a.Item.ObservedAt.AsTime()); c != 0 {
			return c
		}
		return cmp.Compare(b.ConnectionID+"/"+b.ScopeID, a.ConnectionID+"/"+a.ScopeID)
	})
	for _, o := range observations {
		if q.AllowedConnections != nil && !slices.Contains(q.AllowedConnections, o.ConnectionID) {
			continue
		}
		if (q.ConnectionID != "" && o.ConnectionID != q.ConnectionID) || (q.ScopeID != "" && o.ScopeID != q.ScopeID) {
			continue
		}
		if !slices.ContainsFunc(records, func(r Record) bool { return visible(r, o) }) {
			continue
		}
		// Window/content filters must follow source choice: an older eligible
		// version cannot resurrect an item whose latest valid version has ended.
		item := groups[o.Item.Id]
		if item == nil {
			item = proto.Clone(o.Item).(*v1.WorkStationItem)
			item.Sources = nil
			groups[item.Id] = item
		}
		item.Sources = append(item.Sources, &v1.WorkStationSource{ConnectionId: o.ConnectionID, ScopeId: o.ScopeID})
	}
	result := []*v1.WorkStationItem{}
	for _, item := range groups {
		if !Retained(item, q.Now) || (q.ID != "" && item.Id != q.ID) || (q.Kind != "" && item.Kind != q.Kind) {
			continue
		}
		if (q.Provider != "" && item.Provider != q.Provider) || (q.ProjectID != "" && item.ProjectId != q.ProjectID) {
			continue
		}
		if q.Status == "open" && item.Terminal {
			continue
		}
		if q.Status != "" && q.Status != "all" && q.Status != "open" && item.Status != q.Status {
			continue
		}
		if q.Assignee != "" && !slices.Contains(item.Assignees, q.Assignee) {
			continue
		}
		keyword := strings.ToLower(q.Keyword)
		if keyword != "" && !strings.Contains(strings.ToLower(item.Title), keyword) && !strings.Contains(strings.ToLower(item.Identifier), keyword) {
			continue
		}
		if q.AfterID != "" && (item.SourceUpdatedAt.AsTime().After(q.AfterTime) || (item.SourceUpdatedAt.AsTime().Equal(q.AfterTime) && item.Id >= q.AfterID)) {
			continue
		}
		result = append(result, item)
	}
	slices.SortFunc(result, func(a, b *v1.WorkStationItem) int {
		if c := b.SourceUpdatedAt.AsTime().Compare(a.SourceUpdatedAt.AsTime()); c != 0 {
			return c
		}
		return cmp.Compare(b.Id, a.Id)
	})
	if q.Limit > 0 && len(result) > q.Limit {
		result = result[:q.Limit]
	}
	return result
}
