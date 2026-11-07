package gitlab

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strconv"
	"time"

	ws "go.orx.me/apps/butter/internal/workstation"
)

// batchCursor contains only traversal metadata, never credentials or item bodies.
// Remembering the IIDs detects pagination that stopped making progress.
// At the 50,000-row endpoint cap it remains well below Mongo's document limit.
type batchCursor struct {
	Version int      `json:"v"`
	API     string   `json:"api"`
	Scope   string   `json:"scope"`
	Path    string   `json:"path"`
	Started int64    `json:"started"`
	Phase   int      `json:"phase"`
	Page    int      `json:"page"`
	Seen    []string `json:"seen,omitempty"`
}

func (p *Provider) decodeCursor(encoded string, scope ws.Scope, now time.Time) (batchCursor, error) {
	if !decimal(scope.ID) {
		return batchCursor{}, ws.ErrInvalid
	}
	if encoded == "" {
		return batchCursor{Version: 1, API: p.api.String(), Scope: scope.ID, Started: now.UnixNano(), Page: 1}, nil
	}
	// IDs are at most 30 bytes; allow bounded JSON/base64 overhead too.
	if len(encoded) > 4<<20 {
		return batchCursor{}, ws.ErrInvalid
	}
	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return batchCursor{}, ws.ErrInvalid
	}
	var cursor batchCursor
	if json.Unmarshal(data, &cursor) != nil || cursor.Version != 1 || cursor.API != p.api.String() || cursor.Scope != scope.ID || cursor.Started != now.UnixNano() || !safePath(cursor.Path, false) || cursor.Phase < 0 || cursor.Phase > 1 || cursor.Page < 1 || cursor.Page > maxPages || len(cursor.Seen) > pageSize*maxPages {
		return batchCursor{}, ws.ErrInvalid
	}
	seen := make(map[string]bool, len(cursor.Seen))
	for _, id := range cursor.Seen {
		if !decimal(id) || seen[id] {
			return batchCursor{}, ws.ErrInvalid
		}
		seen[id] = true
	}
	return cursor, nil
}

func encodeCursor(cursor batchCursor) string {
	data, _ := json.Marshal(cursor)
	return base64.RawURLEncoding.EncodeToString(data)
}

// ScanBatch advances exactly one Issues or MR list page. The shared runtime
// persists its cursor only together with a fenced, immutable staged batch. A
// request failure therefore retries this page, keeping the old projection live.
// now is the scan's original timestamp, including after a process restart.
func (p *Provider) ScanBatch(ctx context.Context, token string, scope ws.Scope, now time.Time, encoded string) (ws.ScanBatch, error) {
	cursor, err := p.decodeCursor(encoded, scope, now)
	if err != nil {
		return ws.ScanBatch{}, err
	}
	r, err := p.project(ctx, token, scope.ID)
	if err != nil {
		return ws.ScanBatch{}, err
	}
	if r.IssuesEnabled == nil || r.MergeRequestsEnabled == nil {
		return ws.ScanBatch{}, failure("SCHEMA_ERROR")
	}
	if !*r.IssuesEnabled || !*r.MergeRequestsEnabled {
		return ws.ScanBatch{}, failure("FEATURE_UNAVAILABLE")
	}
	if cursor.Path != "" && cursor.Path != r.Path {
		return ws.ScanBatch{}, failure("ACCESS_DENIED")
	}
	cursor.Path = r.Path
	scope.Name = r.Path
	endpoint, kind := "issues", "ISSUE"
	if cursor.Phase == 1 {
		endpoint, kind = "merge_requests", "MERGE_REQUEST"
	}
	path := "/projects/" + scope.ID + "/" + endpoint
	q := url.Values{"state": {"all"}, "scope": {"all"}, "order_by": {"updated_at"}, "sort": {"desc"}, "per_page": {strconv.Itoa(pageSize)}, "page": {strconv.Itoa(cursor.Page)}}
	var remote []remoteItem
	h, err := p.get(ctx, token, path, q, &remote)
	if err != nil {
		return ws.ScanBatch{}, err
	}
	if remote == nil {
		return ws.ScanBatch{}, failure("SCHEMA_ERROR")
	}
	next, err := p.nextPage(h, path, q, cursor.Page, len(remote))
	if err != nil {
		return ws.ScanBatch{}, err
	}
	seen := make(map[string]bool, len(cursor.Seen)+len(remote))
	for _, id := range cursor.Seen {
		seen[id] = true
	}
	result := ws.ScanBatch{Coverage: "COMPLETE"}
	newItems := 0
	pageSeen := make(map[string]bool, len(remote))
	for _, r := range remote {
		item, err := p.normalize(r, scope, kind)
		if err != nil {
			return ws.ScanBatch{}, err
		}
		if pageSeen[item.RemoteId] {
			return ws.ScanBatch{}, failure("SCHEMA_ERROR")
		}
		pageSeen[item.RemoteId] = true
		if !seen[item.RemoteId] {
			seen[item.RemoteId] = true
			cursor.Seen = append(cursor.Seen, item.RemoteId)
			newItems++
		}
		if item.Terminal && item.EndedAt == nil {
			result.Coverage, result.Unresolved = "PARTIAL", result.Unresolved+1
		}
		// Keep terminal markers and expired rows until the runtime selects the
		// newest observation across all pages. Omitting a newer closed version
		// here would let an older open version from another page reappear.
		result.Items = append(result.Items, item)
	}
	if len(remote) > 0 && newItems == 0 {
		return ws.ScanBatch{}, failure("SCHEMA_ERROR")
	}
	if next != "" {
		cursor.Page++
		result.Next = encodeCursor(cursor)
		return result, nil
	}
	if cursor.Phase == 0 {
		cursor.Phase, cursor.Page, cursor.Seen = 1, 1, nil
		result.Next = encodeCursor(cursor)
		return result, nil
	}
	// The final permission/name check is part of the last batch. Its failure
	// leaves all prior batches unpublished, just like a failed intermediate page.
	current, err := p.project(ctx, token, scope.ID)
	if err != nil {
		return ws.ScanBatch{}, err
	}
	if current.Path != cursor.Path || current.IssuesEnabled == nil || current.MergeRequestsEnabled == nil || !*current.IssuesEnabled || !*current.MergeRequestsEnabled {
		return ws.ScanBatch{}, failure("ACCESS_DENIED")
	}
	result.Complete = true
	return result, nil
}

var _ ws.BatchProvider = (*Provider)(nil)
