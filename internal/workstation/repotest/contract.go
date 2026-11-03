// Package repotest runs the same behavioral contract against memory and Mongo.
package repotest

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	cryptomemory "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	"go.orx.me/apps/butter/internal/secretbox"
	ws "go.orx.me/apps/butter/internal/workstation"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Provider struct {
	Failure error
	Items   []*v1.WorkStationItem
}

func (p *Provider) Probe(context.Context, string) (ws.Identity, error) {
	return ws.Identity{PrincipalID: "123"}, p.Failure
}
func (p *Provider) DiscoverScopes(context.Context, string, string) (ws.Page, error) {
	return ws.Page{Scopes: []ws.Scope{{ID: "42", Name: "org/repo"}}}, p.Failure
}
func (p *Provider) ResolveScope(_ context.Context, _ string, id string) (ws.Scope, error) {
	if id != "42" {
		return ws.Scope{}, &ws.Failure{Category: "ACCESS_DENIED"}
	}
	return ws.Scope{ID: id, Name: "org/repo"}, p.Failure
}
func (p *Provider) Scan(context.Context, string, ws.Scope, time.Time) (ws.Scan, error) {
	return ws.Scan{Items: p.Items, Coverage: "COMPLETE"}, p.Failure
}
func NewService(repo ws.Repository, p *Provider) *ws.Service {
	return &ws.Service{Repo: repo, Provider: p, Keyring: secretbox.NewKeyring(cryptomemory.New()), CheckHost: func(context.Context, string) error { return nil }}
}
func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
func Item(number, title string, updated time.Time) *v1.WorkStationItem {
	return &v1.WorkStationItem{Provider: "github", Kind: "ISSUE", RemoteId: number, Identifier: "#" + number, ScopeId: "42", ScopeName: "org/repo", Title: title, Description: title, DescriptionAvailable: true, Status: "NOT_STARTED", SourceUpdatedAt: timestamppb.New(updated), SourceCreatedAt: timestamppb.New(updated.Add(-365 * 24 * time.Hour))}
}
func create(t *testing.T, s *ws.Service, w, name string) ws.Record {
	t.Helper()
	r, err := s.Create(context.Background(), w, name, "host", "secret-value")
	must(t, err)
	return r
}
func selectScope(t *testing.T, s *ws.Service, r ws.Record) ws.Record {
	t.Helper()
	r, err := s.Select(context.Background(), r.WorkspaceID, r.ID, r.Revision, []string{"42"})
	must(t, err)
	return r
}
func claim(t *testing.T, s *ws.Service, r ws.Record) ws.Attempt {
	t.Helper()
	a, err := s.Claim(context.Background(), r.WorkspaceID, r.ID, "42")
	must(t, err)
	return a
}
func publish(t *testing.T, s *ws.Service, a ws.Attempt, items ...*v1.WorkStationItem) {
	t.Helper()
	must(t, s.Publish(context.Background(), a, ws.Scan{Coverage: "COMPLETE", Items: items}))
}
func list(t *testing.T, s *ws.Service, w string, q ws.Query) []*v1.WorkStationItem {
	t.Helper()
	items, _, err := s.List(context.Background(), w, q, "")
	must(t, err)
	return items
}
func Run(t *testing.T, newRepo func(*testing.T) ws.Repository) {
	t.Run("staging_visibility_and_workspace_isolation", func(t *testing.T) {
		repo := newRepo(t)
		s := NewService(repo, &Provider{})
		r := create(t, s, "A", "main")
		if len(r.Scopes) != 0 {
			t.Fatal("create selected a remote resource")
		}
		_, err := s.Discover(context.Background(), "A", r.ID, "")
		must(t, err)
		if len(list(t, s, "A", ws.Query{})) != 0 {
			t.Fatal("discovery imported data")
		}
		r = selectScope(t, s, r)
		a := claim(t, s, r)
		now, _ := repo.Now(context.Background())
		item := Item("1", "staging", now)
		item.Id = ws.ItemID("A", "42", "ISSUE", "1")
		item.ObservedAt = timestamppb.New(now)
		must(t, repo.Stage(context.Background(), []ws.Observation{{WorkspaceID: "A", ConnectionID: r.ID, ScopeID: "42", Snapshot: "orphan", Item: item}, {WorkspaceID: "A", ConnectionID: r.ID, ScopeID: "42", Snapshot: "", Item: item}}))
		if len(list(t, s, "A", ws.Query{})) != 0 {
			t.Fatal("staging leaked")
		}
		publish(t, s, a, Item("1", "published", now))
		if len(list(t, s, "A", ws.Query{})) != 1 || len(list(t, s, "B", ws.Query{})) != 0 {
			t.Fatal("workspace isolation")
		}
		_, err = s.Get(context.Background(), "B", item.Id)
		if !errors.Is(err, ws.ErrNotFound) {
			t.Fatalf("cross-workspace detail: %v", err)
		}
		_, _, err = s.Refresh(context.Background(), "B", r.ID)
		if !errors.Is(err, ws.ErrNotFound) {
			t.Fatal(err)
		}
	})
	t.Run("source_choice_filters_and_immediate_fallback", func(t *testing.T) {
		repo := newRepo(t)
		s := NewService(repo, &Provider{})
		r1 := selectScope(t, s, create(t, s, "A", "one"))
		r2 := selectScope(t, s, create(t, s, "A", "two"))
		now, _ := repo.Now(context.Background())
		publish(t, s, claim(t, s, r1), Item("1", "older safe body", now.Add(-time.Hour)))
		publish(t, s, claim(t, s, r2), Item("1", "new private body", now))
		got := list(t, s, "A", ws.Query{})
		if len(got) != 1 || got[0].Title != "new private body" || len(got[0].Sources) != 2 {
			t.Fatalf("dedup: %v", got)
		}
		got = list(t, s, "A", ws.Query{ConnectionID: r1.ID})
		if len(got) != 1 || got[0].Title != "older safe body" {
			t.Fatal("connection filter chose another source")
		}
		if len(list(t, s, "A", ws.Query{ConnectionID: r1.ID, Keyword: "new private"})) != 0 {
			t.Fatal("content filter applied before source restriction")
		}
		_, err := s.Select(context.Background(), "A", r2.ID, r2.Revision, nil)
		must(t, err)
		got = list(t, s, "A", ws.Query{})
		if len(got) != 1 || got[0].Title != "older safe body" || len(got[0].Sources) != 1 {
			t.Fatal("revoked source survived")
		}
		must(t, s.Delete(context.Background(), "A", r1.ID, r1.Revision))
		if len(list(t, s, "A", ws.Query{})) != 0 {
			t.Fatal("deleted source survived")
		}
		must(t, s.Delete(context.Background(), "A", r1.ID, r1.Revision))
	})
	t.Run("late_publish_and_errors_cannot_restore_removed_scope_or_key", func(t *testing.T) {
		for _, operation := range []string{"remove", "clear", "rotate", "delete"} {
			t.Run(operation, func(t *testing.T) {
				repo := newRepo(t)
				s := NewService(repo, &Provider{})
				r := selectScope(t, s, create(t, s, "A", "main"))
				a := claim(t, s, r)
				now, _ := repo.Now(context.Background())
				var err error
				switch operation {
				case "remove":
					_, err = s.Select(context.Background(), "A", r.ID, r.Revision, nil)
				case "clear":
					key := ""
					_, err = s.Update(context.Background(), "A", r.ID, r.Revision, r.Name, &key)
				case "rotate":
					key := "replacement"
					_, err = s.Update(context.Background(), "A", r.ID, r.Revision, r.Name, &key)
				case "delete":
					err = s.Delete(context.Background(), "A", r.ID, r.Revision)
				}
				must(t, err)
				err = s.Publish(context.Background(), a, ws.Scan{Coverage: "COMPLETE", Items: []*v1.WorkStationItem{Item("1", "late private body", now)}})
				if err == nil {
					t.Fatal("late publisher won")
				}
				if s.Fail(context.Background(), a, &ws.Failure{Category: "AUTH_INVALID"}) == nil {
					t.Fatal("late error won")
				}
				if len(list(t, s, "A", ws.Query{})) != 0 {
					t.Fatal("old generation visible")
				}
			})
		}
	})
	t.Run("takeover_fences_previous_publisher", func(t *testing.T) {
		repo := newRepo(t)
		s := NewService(repo, &Provider{})
		r := selectScope(t, s, create(t, s, "A", "main"))
		a := claim(t, s, r)
		_, err := s.Claim(context.Background(), "A", r.ID, "42")
		if !errors.Is(err, ws.ErrFenced) {
			t.Fatal("second holder acquired live lease")
		}
		old, err := repo.Get(context.Background(), "A", r.ID)
		must(t, err)
		next := old
		next.Scopes = append([]ws.Scope(nil), old.Scopes...)
		next.Scopes[0].LeaseUntil = time.Now().Add(-time.Second)
		next.Version++
		ok, err := repo.CAS(context.Background(), old, &next, nil)
		must(t, err)
		if !ok {
			t.Fatal("expire fixture")
		}
		b := claim(t, s, r)
		now, _ := repo.Now(context.Background())
		publish(t, s, b, Item("1", "new holder", now))
		if s.Publish(context.Background(), a, ws.Scan{Coverage: "COMPLETE"}) == nil {
			t.Fatal("old holder published")
		}
		current, err := repo.Get(context.Background(), "A", r.ID)
		must(t, err)
		copy := current
		copy.Version++
		ok, err = repo.CAS(context.Background(), current, &copy, &ws.Fence{Token: b.Token, Until: now.Add(-time.Second)})
		must(t, err)
		if ok {
			t.Fatal("expired fence committed")
		}
		if got := list(t, s, "A", ws.Query{}); len(got) != 1 || got[0].Title != "new holder" {
			t.Fatal("late write changed chosen body")
		}
	})
	t.Run("failure_preserves_snapshot_but_revocation_hides", func(t *testing.T) {
		repo := newRepo(t)
		s := NewService(repo, &Provider{})
		r := selectScope(t, s, create(t, s, "A", "main"))
		now, _ := repo.Now(context.Background())
		publish(t, s, claim(t, s, r), Item("1", "safe", now))
		_, _, err := s.Refresh(context.Background(), "A", r.ID)
		must(t, err)
		a := claim(t, s, r)
		must(t, s.Fail(context.Background(), a, &ws.Failure{Category: "RATE_LIMITED", RetryAfter: 10 * time.Minute}))
		if len(list(t, s, "A", ws.Query{})) != 1 {
			t.Fatal("rate limit removed snapshot")
		}
		status, due, err := s.Refresh(context.Background(), "A", r.ID)
		must(t, err)
		if status != "DEFERRED" || due.Before(now.Add(9*time.Minute)) {
			t.Fatal("refresh bypassed Retry-After")
		}
		// An explicit owner test resets the backoff, but hides the old key epoch.
		r, err = s.Test(context.Background(), "A", r.ID)
		must(t, err)
		a = claim(t, s, r)
		publish(t, s, a, Item("1", "new safe", now))
		r, err = s.Test(context.Background(), "A", r.ID)
		must(t, err)
		a = claim(t, s, r)
		must(t, s.Fail(context.Background(), a, &ws.Failure{Category: "AUTH_INVALID"}))
		if len(list(t, s, "A", ws.Query{})) != 0 {
			t.Fatal("invalid credential visible")
		}
	})
	t.Run("window_keyset_and_query_bound_tokens", func(t *testing.T) {
		repo := newRepo(t)
		s := NewService(repo, &Provider{})
		r := selectScope(t, s, create(t, s, "A", "main"))
		now, _ := repo.Now(context.Background())
		old := Item("4", "edited old closed", now)
		old.Terminal = true
		old.EndedAt = timestamppb.New(now.Add(-91 * 24 * time.Hour))
		unknown := Item("5", "unknown end", now)
		unknown.Terminal = true
		recent := Item("6", "recent close", now)
		recent.Terminal = true
		recent.EndedAt = timestamppb.New(now.Add(-89 * 24 * time.Hour))
		publish(t, s, claim(t, s, r), Item("1", "one [literal]", now), Item("2", "two", now), Item("3", "ancient open", now.Add(-400*24*time.Hour)), old, unknown, recent)
		all := list(t, s, "A", ws.Query{Status: "all"})
		if len(all) != 4 {
			t.Fatalf("window: got %d", len(all))
		}
		if len(list(t, s, "A", ws.Query{Keyword: "[literal]"})) != 1 {
			t.Fatal("keyword was not literal")
		}
		seen := map[string]bool{}
		page := ""
		for {
			rows, next, err := s.List(context.Background(), "A", ws.Query{Limit: 1, Status: "all"}, page)
			must(t, err)
			for _, item := range rows {
				if seen[item.Id] {
					t.Fatal("duplicate keyset row")
				}
				seen[item.Id] = true
			}
			if next == "" {
				break
			}
			page = next
		}
		if len(seen) != 4 {
			t.Fatal("keyset omitted rows")
		}
		_, token, err := s.List(context.Background(), "A", ws.Query{Limit: 1}, "")
		must(t, err)
		if token == "" {
			t.Fatal("expected token")
		}
		for _, w := range []string{"A", "B"} {
			q := ws.Query{Limit: 1}
			if w == "A" {
				q.Keyword = "different"
			}
			_, _, err = s.List(context.Background(), w, q, token)
			if !errors.Is(err, ws.ErrInvalid) {
				t.Fatal("reused foreign query token")
			}
		}
	})
	t.Run("config_CAS_and_secret_seam", func(t *testing.T) {
		repo := newRepo(t)
		s := NewService(repo, &Provider{})
		r := create(t, s, "A", "main")
		if strings.Contains(r.Credential.Ciphertext, "secret-value") {
			t.Fatal("plaintext stored")
		}
		before := r.Credential
		next, err := s.Update(context.Background(), "A", r.ID, r.Revision, "renamed", nil)
		must(t, err)
		if next.Credential != before {
			t.Fatal("absent key replaced it")
		}
		_, err = s.Update(context.Background(), "A", r.ID, r.Revision, "stale", nil)
		if !errors.Is(err, ws.ErrConflict) {
			t.Fatal("stale config accepted")
		}
		key := ""
		next, err = s.Update(context.Background(), "A", r.ID, next.Revision, next.Name, &key)
		must(t, err)
		if next.Credential.Ciphertext != "" || !next.Paused {
			t.Fatal("clear key did not revoke")
		}
		used, err := repo.HostInUse(context.Background(), "host")
		must(t, err)
		if !used {
			t.Fatal("host reference missing")
		}
	})
	t.Run("gc_keeps_live_and_other_connections", func(t *testing.T) {
		repo := newRepo(t)
		s := NewService(repo, &Provider{})
		r := selectScope(t, s, create(t, s, "A", "main"))
		now, _ := repo.Now(context.Background())
		a := claim(t, s, r)
		publish(t, s, a, Item("1", "keep", now))
		must(t, repo.Collect(context.Background()))
		if len(list(t, s, "A", ws.Query{})) != 1 {
			t.Fatal("GC removed committed data")
		}
	})
}
