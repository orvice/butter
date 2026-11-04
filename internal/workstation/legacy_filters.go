package workstation

import (
	"context"
	"slices"
)

type FilterOptionSet struct {
	Statuses  []string
	Assignees []string
	Labels    []string
}

func (s *Service) FilterOptions(ctx context.Context, w string, q Query) (*FilterOptionSet, error) {
	if q.Kind == "" {
		q.Kind = "ISSUE"
	}
	if q.Kind != "ISSUE" && q.Kind != "PULL_REQUEST" && q.Kind != "MERGE_REQUEST" {
		return nil, ErrInvalid
	}
	q.Status = "all"
	q.Limit = 0
	var err error
	q.Now, err = s.Repo.Now(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.allowedSources(ctx, w, &q); err != nil {
		return nil, err
	}
	items, err := s.Repo.Query(ctx, w, q)
	if err != nil {
		return nil, err
	}
	sets := map[string]map[string]struct{}{
		"status": {}, "assignee": {}, "label": {},
	}
	for _, item := range items {
		if item.Status != "" {
			sets["status"][item.Status] = struct{}{}
		}
		for _, assignee := range item.Assignees {
			if assignee != "" {
				sets["assignee"][assignee] = struct{}{}
			}
		}
		for _, label := range item.Labels {
			if label != "" {
				sets["label"][label] = struct{}{}
			}
		}
	}
	options := func(values map[string]struct{}) []string {
		out := make([]string, 0, len(values))
		for value := range values {
			out = append(out, value)
		}
		slices.Sort(out)
		return out
	}
	return &FilterOptionSet{Statuses: options(sets["status"]), Assignees: options(sets["assignee"]), Labels: options(sets["label"])}, nil
}
