package github

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"io"
	"net/url"
	"sort"
	"strconv"
	"time"

	ws "go.orx.me/apps/butter/internal/workstation"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

var _ ws.BatchProvider = (*Provider)(nil)

const maxBatchCursor = 16384

// The cursor contains only scope metadata and compact loop fingerprints.
// Fixed scan start and scope bindings prevent replay into another traversal.
type batchCursor struct {
	Profile string    `json:"profile"`
	Scope   string    `json:"scope"`
	Started time.Time `json:"started"`
	Name    string    `json:"name"`
	Phase   string    `json:"phase"`
	Page    int       `json:"page"`
	Pages   []byte    `json:"pages,omitempty"`
}

func decodeBatchCursor(raw string, scope ws.Scope, now time.Time) (batchCursor, error) {
	var cursor batchCursor
	if len(raw) > maxBatchCursor {
		return cursor, ws.ErrInvalid
	}
	decoder := json.NewDecoder(bytes.NewBufferString(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&cursor) != nil || decoder.Decode(new(any)) != io.EOF || cursor.Profile != "github-full-v1" || cursor.Scope != scope.ID || !cursor.Started.Equal(now) || cursor.Page < 1 || cursor.Page > maxPages || len(cursor.Name) > 1024 || len(cursor.Pages)%8 != 0 || len(cursor.Pages) > maxPages*8 {
		return batchCursor{}, ws.ErrInvalid
	}
	if cursor.Phase != "issues" && cursor.Phase != "pulls" && cursor.Phase != "verify" {
		return batchCursor{}, ws.ErrInvalid
	}
	if _, err := repositoryPath(cursor.Name); err != nil {
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

// ScanBatch advances at most one upstream request. All pages are staged by the
// runtime, and only a separate final access/identity check marks completion.
func (p *Provider) ScanBatch(ctx context.Context, token string, scope ws.Scope, now time.Time, raw string) (ws.ScanBatch, error) {
	if !decimal(scope.ID) || now.IsZero() {
		return ws.ScanBatch{}, ws.ErrInvalid
	}
	// Mongo persists the fixed start as a BSON Date with millisecond precision.
	// Use that same precision before first-page retention and cursor binding.
	now = now.UTC().Truncate(time.Millisecond)
	if raw == "" {
		repo, err := p.repository(ctx, token, scope.ID)
		if err != nil {
			return ws.ScanBatch{}, err
		}
		if repo.HasIssues == nil {
			return ws.ScanBatch{}, failure("SCHEMA_ERROR")
		}
		if !*repo.HasIssues {
			return ws.ScanBatch{}, failure("FEATURE_UNAVAILABLE")
		}
		if _, err := repositoryPath(repo.FullName); err != nil {
			return ws.ScanBatch{}, err
		}
		next, err := encodeBatchCursor(batchCursor{Profile: "github-full-v1", Scope: scope.ID, Started: now, Name: repo.FullName, Phase: "issues", Page: 1})
		if err != nil {
			return ws.ScanBatch{}, err
		}
		return ws.ScanBatch{Next: next, Coverage: "COMPLETE"}, nil
	}
	cursor, err := decodeBatchCursor(raw, scope, now)
	if err != nil {
		return ws.ScanBatch{}, err
	}
	if cursor.Phase == "verify" {
		repo, err := p.repository(ctx, token, scope.ID)
		if err != nil {
			return ws.ScanBatch{}, err
		}
		if repo.FullName != cursor.Name || repo.HasIssues == nil {
			return ws.ScanBatch{}, failure("SCHEMA_ERROR")
		}
		if !*repo.HasIssues {
			return ws.ScanBatch{}, failure("FEATURE_UNAVAILABLE")
		}
		return ws.ScanBatch{Complete: true, Coverage: "COMPLETE"}, nil
	}
	base, err := repositoryPath(cursor.Name)
	if err != nil {
		return ws.ScanBatch{}, err
	}
	path := base + "/" + cursor.Phase
	query := url.Values{"state": {"all"}, "sort": {"updated"}, "direction": {"desc"}, "per_page": {"100"}, "page": {strconv.Itoa(cursor.Page)}}
	var rows []remoteItem
	headers, err := p.get(ctx, token, path, query, &rows)
	if err != nil {
		return ws.ScanBatch{}, err
	}
	if rows == nil {
		return ws.ScanBatch{}, failure("SCHEMA_ERROR")
	}
	signature := ""
	for _, row := range rows {
		if !decimal(row.Number.String()) {
			return ws.ScanBatch{}, failure("SCHEMA_ERROR")
		}
		signature += row.Number.String() + ","
	}
	if len(rows) > 0 {
		hash := fingerprint(signature)
		if containsFingerprint(cursor.Pages, hash) {
			return ws.ScanBatch{}, failure("SCHEMA_ERROR")
		}
		cursor.Pages = append(cursor.Pages, hash...)
	}
	next, err := nextPage(headers, path, query, cursor.Page)
	if err != nil {
		return ws.ScanBatch{}, err
	}
	if next != "" && len(rows) == 0 {
		return ws.ScanBatch{}, failure("SCHEMA_ERROR")
	}
	batch := ws.ScanBatch{Coverage: "COMPLETE"}
	items := map[string]*v1.WorkStationItem{}
	kind := "ISSUE"
	if cursor.Phase == "pulls" {
		kind = "PULL_REQUEST"
	}
	for _, row := range rows {
		if kind == "ISSUE" && len(row.PullRequest) > 0 && string(row.PullRequest) != "null" {
			continue
		}
		item, err := normalize(row, ws.Scope{ID: scope.ID, Name: cursor.Name}, kind)
		if err != nil {
			return ws.ScanBatch{}, err
		}
		if old := items[item.RemoteId]; old == nil || item.SourceUpdatedAt.AsTime().After(old.SourceUpdatedAt.AsTime()) {
			items[item.RemoteId] = item
		}
	}
	for _, item := range items {
		if item.Terminal && item.EndedAt == nil {
			batch.Unresolved++
		}
		// Preserve markers and expired terminal rows until the runtime chooses
		// the latest observation across all pages. Filtering earlier can revive
		// an older open row from another page.
		batch.Items = append(batch.Items, item)
	}
	if batch.Unresolved > 0 {
		batch.Coverage = "PARTIAL"
	}
	sort.Slice(batch.Items, func(i, j int) bool { return batch.Items[i].RemoteId < batch.Items[j].RemoteId })
	if next != "" {
		cursor.Page++
	} else {
		cursor.Page = 1
		cursor.Pages = nil
		if cursor.Phase == "issues" {
			cursor.Phase = "pulls"
		} else {
			cursor.Phase = "verify"
		}
	}
	batch.Next, err = encodeBatchCursor(cursor)
	if err != nil {
		return ws.ScanBatch{}, err
	}
	return batch, nil
}
