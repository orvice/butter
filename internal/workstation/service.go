package workstation

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.orx.me/apps/butter/internal/secretbox"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	Repo     Repository
	Keyring  *secretbox.Keyring
	Provider Provider
	// CheckHost validates the existing Git Host on every operation/attempt.
	CheckHost func(context.Context, string) error
}

func (s *Service) ready() error {
	if s == nil || s.Repo == nil || s.Keyring == nil || s.Provider == nil || s.CheckHost == nil {
		return ErrUnsupported
	}
	return nil
}
func validName(name string) bool { return strings.TrimSpace(name) != "" && len(name) <= 128 }
func (s *Service) mutate(ctx context.Context, w, id string, change func(*Record, time.Time) error) (Record, error) {
	for range 16 {
		old, err := s.Repo.Get(ctx, w, id)
		if err != nil {
			return Record{}, err
		}
		now, err := s.Repo.Now(ctx)
		if err != nil {
			return Record{}, err
		}
		next := old
		next.Scopes = append([]Scope(nil), old.Scopes...)
		if err := change(&next, now); err != nil {
			return Record{}, err
		}
		next.Version++
		ok, err := s.Repo.CAS(ctx, old, &next, nil)
		if err != nil {
			return Record{}, err
		}
		if ok {
			return next, nil
		}
	}
	return Record{}, ErrConflict
}
func (s *Service) Create(ctx context.Context, w, name, host, key string) (Record, error) {
	if err := s.ready(); err != nil {
		return Record{}, err
	}
	if !validName(name) || host == "" || strings.TrimSpace(key) == "" || len(key) > 16384 {
		return Record{}, ErrInvalid
	}
	if err := s.CheckHost(ctx, host); err != nil {
		return Record{}, err
	}
	identity, err := s.Provider.Probe(ctx, key)
	if err != nil {
		return Record{}, err
	}
	cipher, kid, err := s.Keyring.Encrypt(ctx, []byte(key))
	if err != nil {
		return Record{}, err
	}
	r := Record{WorkspaceID: w, ID: uuid.NewString(), Name: strings.TrimSpace(name), HostID: host, PrincipalID: identity.PrincipalID, Revision: 1, Version: 1, Epoch: uuid.NewString(), Credential: Credential{Ciphertext: cipher, KeyID: kid}, Scopes: []Scope{}}
	return r, s.Repo.Create(ctx, r)
}
func (s *Service) token(ctx context.Context, r Record) (string, error) {
	if err := s.ready(); err != nil {
		return "", err
	}
	if err := s.CheckHost(ctx, r.HostID); err != nil {
		return "", err
	}
	if r.Credential.Ciphertext == "" {
		return "", ErrPaused
	}
	b, err := s.Keyring.Decrypt(ctx, r.Credential.Ciphertext, r.Credential.KeyID)
	if err != nil {
		return "", ErrPaused
	}
	return string(b), nil
}
func reset(r *Record, now time.Time) {
	r.Epoch = uuid.NewString()
	r.RetryAfter = time.Time{}
	for i := range r.Scopes {
		sc := &r.Scopes[i]
		sc.Generation = uuid.NewString()
		sc.Snapshot = ""
		sc.Claim = ""
		sc.LeaseUntil = time.Time{}
		sc.Hidden = false
		sc.Coverage = ""
		sc.Unresolved = 0
		sc.LastSuccess = time.Time{}
		sc.Due = now
		sc.Health = "QUEUED"
	}
}
func (s *Service) Update(ctx context.Context, w, id string, revision int64, name string, key *string) (Record, error) {
	if id == "" || revision < 1 {
		return Record{}, ErrInvalid
	}
	if !validName(name) {
		return Record{}, ErrInvalid
	}
	old, err := s.Repo.Get(ctx, w, id)
	if err != nil {
		return Record{}, err
	}
	if old.Revision != revision {
		return Record{}, ErrConflict
	}
	var cred Credential
	if key != nil && *key != "" {
		if len(*key) > 16384 {
			return Record{}, ErrInvalid
		}
		if err := s.CheckHost(ctx, old.HostID); err != nil {
			return Record{}, err
		}
		identity, err := s.Provider.Probe(ctx, *key)
		if err != nil {
			return Record{}, err
		}
		if identity.PrincipalID != old.PrincipalID {
			return Record{}, fmt.Errorf("%w: create a new connection for a different GitHub identity", ErrInvalid)
		}
		cred.Ciphertext, cred.KeyID, err = s.Keyring.Encrypt(ctx, []byte(*key))
		if err != nil {
			return Record{}, err
		}
	}
	return s.mutate(ctx, w, id, func(r *Record, now time.Time) error {
		if r.Revision != revision {
			return ErrConflict
		}
		r.Name = strings.TrimSpace(name)
		r.Revision++
		if key != nil {
			r.Credential = cred
			r.Paused = cred.Ciphertext == ""
			reset(r, now)
		}
		return nil
	})
}
func (s *Service) Delete(ctx context.Context, w, id string, revision int64) error {
	if id == "" || revision < 1 {
		return ErrInvalid
	}
	for range 16 {
		r, err := s.Repo.Get(ctx, w, id)
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		if r.Revision != revision {
			return ErrConflict
		}
		ok, err := s.Repo.CAS(ctx, r, nil, nil)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
	}
	return ErrConflict
}
func (s *Service) Test(ctx context.Context, w, id string) (Record, error) {
	if id == "" {
		return Record{}, ErrInvalid
	}
	r, err := s.Repo.Get(ctx, w, id)
	if err != nil {
		return Record{}, err
	}
	token, err := s.token(ctx, r)
	if err != nil {
		return Record{}, err
	}
	identity, probeErr := s.Provider.Probe(ctx, token)
	if probeErr != nil && Classify(probeErr).Category != "AUTH_INVALID" {
		return Record{}, probeErr
	}
	return s.mutate(ctx, w, id, func(next *Record, now time.Time) error {
		if next.Revision != r.Revision || next.Epoch != r.Epoch {
			return ErrConflict
		}
		if probeErr == nil && identity.PrincipalID != r.PrincipalID {
			return ErrInvalid
		}
		next.Revision++
		next.Paused = probeErr != nil
		reset(next, now)
		if probeErr != nil {
			for i := range next.Scopes {
				next.Scopes[i].Health = "AUTH_INVALID"
			}
		}
		return nil
	})
}
func (s *Service) Discover(ctx context.Context, w, id, page string) (Page, error) {
	if id == "" {
		return Page{}, ErrInvalid
	}
	r, err := s.Repo.Get(ctx, w, id)
	if err != nil {
		return Page{}, err
	}
	if r.Paused {
		return Page{}, ErrPaused
	}
	token, err := s.token(ctx, r)
	if err != nil {
		return Page{}, err
	}
	return s.Provider.DiscoverScopes(ctx, token, page)
}
func (s *Service) Select(ctx context.Context, w, id string, revision int64, ids []string) (Record, error) {
	if id == "" || revision < 1 {
		return Record{}, ErrInvalid
	}
	if len(ids) > 100 {
		return Record{}, ErrInvalid
	}
	r, err := s.Repo.Get(ctx, w, id)
	if err != nil {
		return Record{}, err
	}
	if r.Revision != revision {
		return Record{}, ErrConflict
	}
	resolved := map[string]Scope{}
	// An empty explicit replacement can always revoke ranges, even with no key.
	if len(ids) > 0 {
		if r.Paused {
			return Record{}, ErrPaused
		}
		token, err := s.token(ctx, r)
		if err != nil {
			return Record{}, err
		}
		for _, id := range ids {
			if _, ok := resolved[id]; ok {
				return Record{}, ErrInvalid
			}
			sc, err := s.Provider.ResolveScope(ctx, token, id)
			if err != nil {
				return Record{}, err
			}
			resolved[id] = sc
		}
	}
	return s.mutate(ctx, w, r.ID, func(next *Record, now time.Time) error {
		if next.Revision != revision {
			return ErrConflict
		}
		scopes := []Scope{}
		for _, id := range ids {
			sc := resolved[id]
			index := slices.IndexFunc(next.Scopes, func(sc Scope) bool { return sc.ID == id })
			if index >= 0 {
				sc = next.Scopes[index]
			} else {
				sc.Generation = uuid.NewString()
				sc.Due = now
				sc.Health = "QUEUED"
			}
			scopes = append(scopes, sc)
		}
		next.Scopes = scopes
		next.Revision++
		return nil
	})
}
func (s *Service) Refresh(ctx context.Context, w, id string) (string, time.Time, error) {
	if id == "" {
		return "", time.Time{}, ErrInvalid
	}
	status := "QUEUED"
	var due time.Time
	_, err := s.mutate(ctx, w, id, func(r *Record, now time.Time) error {
		status = "QUEUED"
		due = now
		if r.Paused || r.Credential.Ciphertext == "" {
			status = "PAUSED"
			return nil
		}
		if len(r.Scopes) == 0 {
			status = "ALREADY_PENDING"
			return nil
		}
		if r.ManualAfter.After(now) || r.RetryAfter.After(now) {
			status = "DEFERRED"
			due = maxTime(r.ManualAfter, r.RetryAfter)
			return nil
		}
		pending := true
		running := false
		for _, sc := range r.Scopes {
			if sc.LeaseUntil.After(now) {
				running = true
			}
			if sc.Due.After(now) {
				pending = false
			}
		}
		if running {
			status = "RUNNING"
			return nil
		}
		if pending {
			status = "ALREADY_PENDING"
			return nil
		}
		r.ManualAfter = now.Add(time.Minute)
		for i := range r.Scopes {
			if !r.Scopes[i].Hidden {
				r.Scopes[i].Due = now
				r.Scopes[i].Health = "QUEUED"
			}
		}
		return nil
	})
	return status, due, err
}
func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
func Public(r Record) *v1.WorkStationConnection {
	c := &v1.WorkStationConnection{Id: r.ID, Name: r.Name, Provider: "github", HostId: r.HostID, PrincipalId: r.PrincipalID, Revision: r.Revision, CredentialSet: r.Credential.Ciphertext != "", Paused: r.Paused}
	for _, sc := range r.Scopes {
		p := &v1.WorkStationScope{Id: sc.ID, Name: sc.Name, Health: sc.Health, Coverage: sc.Coverage, UnresolvedTerminalCount: sc.Unresolved, Initialized: sc.Snapshot != "" && !sc.Hidden && !r.Paused, NextAttemptAt: stamp(sc.Due), LastSuccessAt: stamp(sc.LastSuccess)}
		c.Scopes = append(c.Scopes, p)
	}
	return c
}
func stamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}
func (s *Service) Connections(ctx context.Context, w string) ([]*v1.WorkStationConnection, error) {
	rs, err := s.Repo.List(ctx, w)
	if err != nil {
		return nil, err
	}
	slices.SortFunc(rs, func(a, b Record) int { return strings.Compare(a.ID, b.ID) })
	out := []*v1.WorkStationConnection{}
	for _, r := range rs {
		p := Public(r)
		if err := s.CheckHost(ctx, r.HostID); err != nil {
			p.Paused = true
			for _, sc := range p.Scopes {
				sc.Health = "UNSUPPORTED"
				sc.Initialized = false
			}
		}
		out = append(out, p)
	}
	return out, nil
}

type pageCursor struct {
	Fingerprint string
	ID          string
	Updated     time.Time
}

func fingerprint(w string, q Query) string {
	b, _ := json.Marshal([]string{w, q.Kind, q.ConnectionID, q.ScopeID, q.Status, q.Assignee, q.Keyword})
	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:])
}
func (s *Service) List(ctx context.Context, w string, q Query, page string) ([]*v1.WorkStationItem, string, error) {
	if q.Kind == "" {
		q.Kind = "ISSUE"
	}
	if q.Kind != "ISSUE" && q.Kind != "PULL_REQUEST" {
		return nil, "", ErrInvalid
	}
	if q.Status == "" {
		q.Status = "open"
	}
	if !slices.Contains([]string{"open", "all", "NOT_STARTED", "DONE", "CANCELED", "OPEN", "MERGED", "CLOSED", "OTHER"}, q.Status) || len(q.Keyword) > 256 || len(q.Assignee) > 256 {
		return nil, "", ErrInvalid
	}
	if q.Limit == 0 {
		q.Limit = 50
	}
	if q.Limit < 1 || q.Limit > 100 {
		return nil, "", ErrInvalid
	}
	fp := fingerprint(w, q)
	if page != "" {
		if len(page) > 2048 {
			return nil, "", ErrInvalid
		}
		b, err := base64.RawURLEncoding.DecodeString(page)
		var c pageCursor
		if err != nil || json.Unmarshal(b, &c) != nil || c.Fingerprint != fp || len(c.ID) != 64 || c.Updated.IsZero() {
			return nil, "", ErrInvalid
		}
		q.AfterID = c.ID
		q.AfterTime = c.Updated
	}
	now, err := s.Repo.Now(ctx)
	if err != nil {
		return nil, "", err
	}
	q.Now = now
	if err := s.allowedSources(ctx, w, &q); err != nil {
		return nil, "", err
	}
	limit := q.Limit
	q.Limit++
	rows, err := s.Repo.Query(ctx, w, q)
	if err != nil {
		return nil, "", err
	}
	next := ""
	if len(rows) > limit {
		rows = rows[:limit]
		last := rows[len(rows)-1]
		b, _ := json.Marshal(pageCursor{Fingerprint: fp, ID: last.Id, Updated: last.SourceUpdatedAt.AsTime()})
		next = base64.RawURLEncoding.EncodeToString(b)
	}
	return rows, next, nil
}
func (s *Service) allowedSources(ctx context.Context, w string, q *Query) error {
	rs, err := s.Repo.List(ctx, w)
	if err != nil {
		return err
	}
	q.AllowedConnections = []string{}
	for _, r := range rs {
		if !r.Paused && r.Credential.Ciphertext != "" && s.CheckHost(ctx, r.HostID) == nil {
			q.AllowedConnections = append(q.AllowedConnections, r.ID)
		}
	}
	return nil
}
func (s *Service) Get(ctx context.Context, w, id string) (*v1.WorkStationItem, error) {
	if id == "" {
		return nil, ErrInvalid
	}
	q := Query{ID: id, Limit: 1}
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
	if len(items) == 0 {
		return nil, ErrNotFound
	}
	return items[0], nil
}
