package linear

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"io"
	"sort"
	"time"

	ws "go.orx.me/apps/butter/internal/workstation"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

var _ ws.BatchProvider = (*Provider)(nil)

const maxBatchCursor = 16384

type batchCursor struct {
	Profile string    `json:"profile"`
	Scope   string    `json:"scope"`
	Started time.Time `json:"started"`
	Name    string    `json:"name"`
	Phase   string    `json:"phase"`
	Page    int       `json:"page"`
	After   string    `json:"after,omitempty"`
	Cursors []byte    `json:"cursors,omitempty"`
}

func decodeBatchCursor(raw string, scope ws.Scope, now time.Time) (batchCursor, error) {
	var cursor batchCursor
	if len(raw) > maxBatchCursor {
		return cursor, ws.ErrInvalid
	}
	decoder := json.NewDecoder(bytes.NewBufferString(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&cursor) != nil || decoder.Decode(new(any)) != io.EOF || cursor.Profile != "linear-full-v1" || cursor.Scope != scope.ID || !cursor.Started.Equal(now) || cursor.Name == "" || len(cursor.Name) > 1024 || cursor.Page < 1 || cursor.Page > maxPages || len(cursor.After) > 4096 || len(cursor.Cursors)%8 != 0 || len(cursor.Cursors) > maxPages*8 {
		return batchCursor{}, ws.ErrInvalid
	}
	if cursor.Phase != "issues" && cursor.Phase != "verify" {
		return batchCursor{}, ws.ErrInvalid
	}
	if cursor.Phase == "issues" && cursor.After == "" {
		return batchCursor{}, ws.ErrInvalid
	}
	return cursor, nil
}
func encodeBatchCursor(cursor batchCursor) (string, error) {
	raw, err := json.Marshal(cursor)
	if err != nil || len(raw) > maxBatchCursor {
		return "", failure("UNSUPPORTED")
	}
	return string(raw), nil
}
func fingerprint(value string) []byte { sum := sha256.Sum256([]byte(value)); return sum[:8] }
func containsFingerprint(values, needle []byte) bool {
	for n := 0; n < len(values); n += 8 {
		if bytes.Equal(values[n:n+8], needle) {
			return true
		}
	}
	return false
}

// ScanBatch returns one complete issue page, including its label metadata. A
// separate final Team query authorizes completion; the runtime owns staging and
// all fencing. No earlier page is a visible projection until that final step.
func (p *Provider) ScanBatch(ctx context.Context, token string, scope ws.Scope, now time.Time, raw string) (ws.ScanBatch, error) {
	if !validID(scope.ID) || now.IsZero() {
		return ws.ScanBatch{}, ws.ErrInvalid
	}
	// Keep cursor identity and filtering stable after BSON Date persistence.
	now = now.UTC().Truncate(time.Millisecond)
	cursor := batchCursor{Profile: "linear-full-v1", Scope: scope.ID, Started: now, Phase: "issues", Page: 1}
	var err error
	if raw != "" {
		cursor, err = decodeBatchCursor(raw, scope, now)
		if err != nil {
			return ws.ScanBatch{}, err
		}
	}
	if cursor.Phase == "verify" {
		current, err := p.ResolveScope(ctx, token, scope.ID)
		if err != nil {
			return ws.ScanBatch{}, err
		}
		if current.Name != cursor.Name {
			return ws.ScanBatch{}, failure("SCHEMA_ERROR")
		}
		return ws.ScanBatch{Complete: true, Coverage: "COMPLETE"}, nil
	}
	var data struct {
		Team *struct {
			team
			Issues *connection[issue] `json:"issues"`
		} `json:"team"`
	}
	if err = p.query(ctx, token, issuesQuery, map[string]any{"id": scope.ID, "after": nullable(cursor.After), "filter": retentionFilter(now)}, &data); err != nil {
		return ws.ScanBatch{}, err
	}
	if data.Team == nil {
		return ws.ScanBatch{}, failure("ACCESS_DENIED")
	}
	if data.Team.ID != scope.ID || data.Team.Issues == nil || data.Team.Issues.Nodes == nil {
		return ws.ScanBatch{}, failure("SCHEMA_ERROR")
	}
	fresh, err := teamScope(data.Team.team)
	if err != nil {
		return ws.ScanBatch{}, err
	}
	if cursor.Name != "" && fresh.Name != cursor.Name {
		return ws.ScanBatch{}, failure("SCHEMA_ERROR")
	}
	cursor.Name = fresh.Name
	next, err := data.Team.Issues.Info.next(cursor.After, len(data.Team.Issues.Nodes))
	if err != nil {
		return ws.ScanBatch{}, err
	}
	if next != "" {
		hash := fingerprint(next)
		if containsFingerprint(cursor.Cursors, hash) {
			return ws.ScanBatch{}, failure("SCHEMA_ERROR")
		}
		if cursor.Page >= maxPages {
			return ws.ScanBatch{}, failure("UNSUPPORTED")
		}
		cursor.Cursors = append(cursor.Cursors, hash...)
	}
	items := map[string]*v1.WorkStationItem{}
	for _, remote := range data.Team.Issues.Nodes {
		if remote.Team == nil || !validID(remote.Team.ID) {
			return ws.ScanBatch{}, failure("SCHEMA_ERROR")
		}
		if remote.Team.ID != scope.ID {
			continue
		}
		if err = p.completeLabels(ctx, token, &remote); err != nil {
			return ws.ScanBatch{}, err
		}
		item, err := normalize(remote, ws.Scope{ID: scope.ID, Name: cursor.Name})
		if err != nil {
			return ws.ScanBatch{}, err
		}
		if old := items[item.RemoteId]; old == nil || item.SourceUpdatedAt.AsTime().After(old.SourceUpdatedAt.AsTime()) {
			items[item.RemoteId] = item
		}
	}
	batch := ws.ScanBatch{Coverage: "COMPLETE"}
	for _, item := range items {
		if item.Terminal && item.EndedAt == nil {
			batch.Unresolved++
		}
		// The runtime deduplicates all pages before retention or unresolved
		// filtering, so a newer terminal version cannot revive an older row.
		batch.Items = append(batch.Items, item)
	}
	if batch.Unresolved > 0 {
		batch.Coverage = "PARTIAL"
	}
	sort.Slice(batch.Items, func(i, j int) bool { return batch.Items[i].RemoteId < batch.Items[j].RemoteId })
	if next == "" {
		cursor.Phase = "verify"
		cursor.After = ""
	} else {
		cursor.Page++
		cursor.After = next
	}
	batch.Next, err = encodeBatchCursor(cursor)
	if err != nil {
		return ws.ScanBatch{}, err
	}
	return batch, nil
}
