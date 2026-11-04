package workstation

import "context"

// ResolvedProvider comes from the global approved host registry. A connection
// pins the probed canonical source; host edits cannot silently redirect its key.
type ResolvedProvider struct {
	Adapter    Provider
	Source     SourceIdentity
	WebBaseURL string
}

func (s *Service) resolve(ctx context.Context, r Record) (ResolvedProvider, error) {
	if s.ResolveProvider != nil {
		resolved, err := s.ResolveProvider(ctx, r)
		if err != nil {
			return ResolvedProvider{}, err
		}
		canonical, err := resolved.Source.Canonical()
		if err != nil || resolved.Adapter == nil {
			return ResolvedProvider{}, ErrUnsupported
		}
		resolved.Source = canonical
		if r.Source.Origin != "" && (r.Source.Provider != canonical.Provider || r.Source.Origin != canonical.Origin || (r.WebBaseURL != "" && r.WebBaseURL != resolved.WebBaseURL)) {
			return ResolvedProvider{}, ErrUnsupported
		}
		return resolved, nil
	}
	if s.Provider == nil || s.CheckHost == nil {
		return ResolvedProvider{}, ErrUnsupported
	}
	if err := s.CheckHost(ctx, r.HostID); err != nil {
		return ResolvedProvider{}, err
	}
	return ResolvedProvider{Adapter: s.Provider, Source: GitHubSource(), WebBaseURL: "https://github.com"}, nil
}
func (s *Service) hostValid(ctx context.Context, r Record) bool {
	_, err := s.resolve(ctx, r)
	return err == nil
}
