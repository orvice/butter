package workstation

import (
	"context"
	"errors"
	"github.com/google/uuid"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"slices"
	"time"
)

var errScanPending = errors.New("scan saved for continuation")

func CloneScopes(in []Scope) []Scope {
	out := append([]Scope(nil), in...)
	for i := range out {
		if p := out[i].Pending; p != nil {
			copy := *p
			copy.Batches = append([]string(nil), p.Batches...)
			out[i].Pending = &copy
		}
	}
	return out
}
func (s *Service) advanceBatch(ctx context.Context, a Attempt, token string, b ScanBatch, yield bool) (PendingScan, error) {
	var pending PendingScan
	for range 16 {
		old, err := s.Repo.Get(ctx, a.Record.WorkspaceID, a.Record.ID)
		if err != nil {
			return pending, err
		}
		now, err := s.Repo.Now(ctx)
		if err != nil {
			return pending, err
		}
		if old.Epoch != a.Record.Epoch || old.Paused || !now.Before(a.Until) {
			return pending, ErrFenced
		}
		next := old
		next.Scopes = CloneScopes(old.Scopes)
		i := slices.IndexFunc(next.Scopes, func(sc Scope) bool { return sc.ID == a.Scope.ID })
		if i < 0 {
			return pending, ErrFenced
		}
		sc := &next.Scopes[i]
		if sc.Generation != a.Scope.Generation || sc.Claim != a.Token {
			return pending, ErrFenced
		}
		if sc.Pending != nil {
			pending = *sc.Pending
		} else {
			pending = PendingScan{Started: a.Scope.ScanStarted}
		}
		if len(pending.Batches) >= 10000 || len(b.Next) > 16384 || (!b.Complete && (b.Next == "" || b.Next == pending.Cursor)) {
			return pending, ErrInvalid
		}
		pending.Batches = append(append([]string(nil), pending.Batches...), token)
		pending.Cursor = b.Next
		pending.Count += uint32(len(b.Items))
		pending.Unresolved += b.Unresolved
		if pending.Count > 1000000 {
			return pending, ErrUnsupported
		}
		sc.Pending = &pending
		sc.ScannedItems = pending.Count
		if yield {
			sc.Claim = ""
			sc.LeaseUntil = time.Time{}
			sc.Due = now.Add(time.Second)
			sc.Health = "LAGGING"
		}
		next.Version++
		ok, err := s.Repo.CAS(ctx, old, &next, &Fence{Token: a.Token, Until: a.Until})
		if err != nil {
			return pending, err
		}
		if ok {
			return pending, nil
		}
	}
	return pending, ErrConflict
}

// scanBatches saves only complete pages under unique immutable tokens. At most
// ten pages occupy a lease; a later Pod resumes the fenced cursor and never
// treats its staging as a visible or complete projection.
func (s *Service) scanBatches(ctx context.Context, a Attempt, p BatchProvider, key string) (Scan, error) {
	pending := PendingScan{Started: a.Scope.ScanStarted}
	if a.Scope.Pending != nil {
		pending = *a.Scope.Pending
	}
	for page := 0; page < 10; page++ {
		batch, err := p.ScanBatch(ctx, key, a.Scope, pending.Started, pending.Cursor)
		if err != nil {
			return Scan{}, err
		}
		if batch.Coverage != "COMPLETE" && batch.Coverage != "PARTIAL" {
			return Scan{}, ErrInvalid
		}
		if batch.Coverage == "PARTIAL" && batch.Unresolved == 0 {
			return Scan{}, ErrInvalid
		}
		now, err := s.Repo.Now(ctx)
		if err != nil {
			return Scan{}, err
		}
		if !now.Before(a.Until) {
			return Scan{}, ErrFenced
		}
		token := uuid.NewString()
		obs := []Observation{}
		seen := map[string]bool{}
		for _, source := range batch.Items {
			if source == nil || source.RemoteId == "" || source.ScopeId != a.Scope.ID || source.SourceUpdatedAt == nil {
				return Scan{}, ErrInvalid
			}
			item := proto.Clone(source).(*v1.WorkStationItem)
			if item.Terminal && item.EndedAt == nil {
				batch.Unresolved++
				continue
			}
			item.Id = ItemID(a.Record.WorkspaceID, a.Record.SourceIdentity(), a.Scope.ID, item.Kind, item.RemoteId)
			if seen[item.Id] {
				return Scan{}, ErrInvalid
			}
			seen[item.Id] = true
			item.Provider = a.Record.SourceIdentity().Provider
			item.ObservedAt = timestamppb.New(now)
			item.Sources = nil
			obs = append(obs, Observation{WorkspaceID: a.Record.WorkspaceID, ConnectionID: a.Record.ID, ScopeID: a.Scope.ID, Snapshot: token, Item: item})
		}
		if err := s.Repo.Stage(ctx, obs); err != nil {
			return Scan{}, err
		}
		yield := !batch.Complete && page == 9
		pending, err = s.advanceBatch(ctx, a, token, batch, yield)
		if err != nil {
			return Scan{}, err
		}
		if yield {
			return Scan{}, errScanPending
		}
		if batch.Complete {
			items, err := s.Repo.Staged(ctx, a.Record.WorkspaceID, a.Record.ID, a.Scope.ID, pending.Batches)
			if err != nil {
				return Scan{}, err
			}
			// Moving pages can overlap: choose the latest complete source row.
			unique := map[string]*v1.WorkStationItem{}
			for _, item := range items {
				old := unique[item.Id]
				if old == nil || item.SourceUpdatedAt.AsTime().After(old.SourceUpdatedAt.AsTime()) || (item.SourceUpdatedAt.AsTime().Equal(old.SourceUpdatedAt.AsTime()) && item.ObservedAt.AsTime().After(old.ObservedAt.AsTime())) {
					unique[item.Id] = item
				}
			}
			result := Scan{Mode: "FULL", Coverage: "COMPLETE", Unresolved: pending.Unresolved}
			if result.Unresolved > 0 {
				result.Coverage = "PARTIAL"
			}
			for _, item := range unique {
				result.Items = append(result.Items, item)
			}
			return result, nil
		}
	}
	return Scan{}, ErrInvalid
}
