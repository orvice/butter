package plane

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"strings"
	"time"

	ws "go.orx.me/apps/butter/internal/workstation"
)

// batchCursor contains only traversal identity and position. Source bodies and
// credentials never enter it. The worker persists it behind the epoch/selection
// fence; an adapter cursor alone does not authorize publication.
type batchCursor struct {
	Version  int       `json:"v"`
	Instance string    `json:"instance"`
	Scope    string    `json:"scope"`
	NameHash string    `json:"name_hash"`
	Started  time.Time `json:"started"`
	Next     string    `json:"next"`
	Pages    int       `json:"pages"`
	LastID   string    `json:"last_id"`
	Seen     []string  `json:"seen"`
}

func digest(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}
func (p *Provider) cursorInstance() string { return digest(p.api + "\x00" + p.workspace) }
func validDigest(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size && strings.ToLower(value) == value
}
func (p *Provider) decodeCursor(raw string, scope ws.Scope, started time.Time) (batchCursor, error) {
	if raw == "" {
		return batchCursor{Version: 1, Instance: p.cursorInstance(), Scope: scope.ID, NameHash: digest(scope.Name), Started: started}, nil
	}
	if len(raw) > 128<<10 {
		return batchCursor{}, ws.ErrInvalid
	}
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return batchCursor{}, ws.ErrInvalid
	}
	var cursor batchCursor
	decoder := json.NewDecoder(strings.NewReader(string(decoded)))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&cursor) != nil || decoder.Decode(&struct{}{}) != io.EOF {
		return batchCursor{}, ws.ErrInvalid
	}
	if cursor.Version != 1 || cursor.Instance != p.cursorInstance() || cursor.Scope != scope.ID || cursor.NameHash != digest(scope.Name) || !cursor.Started.Equal(started) || cursor.Next == "" || len(cursor.Next) > 4096 || cursor.Pages < 1 || cursor.Pages >= maxPages || len(cursor.Seen) != cursor.Pages || !pathAtom.MatchString(cursor.LastID) {
		return batchCursor{}, ws.ErrInvalid
	}
	seen := map[string]bool{}
	for _, hash := range cursor.Seen {
		if !validDigest(hash) || seen[hash] {
			return batchCursor{}, ws.ErrInvalid
		}
		seen[hash] = true
	}
	if !seen[digest(cursor.Next)] {
		return batchCursor{}, ws.ErrInvalid
	}
	return cursor, nil
}

// compareID matches the explicit id ordering without floating-point conversion.
// Plane UUIDs sort lexically; integer fixture/profile IDs sort numerically.
func compareID(left, right string) int {
	decimal := func(value string) bool {
		if value == "" {
			return false
		}
		for _, c := range value {
			if c < '0' || c > '9' {
				return false
			}
		}
		return true
	}
	if decimal(left) && decimal(right) {
		l, r := strings.TrimLeft(left, "0"), strings.TrimLeft(right, "0")
		if len(l) < len(r) {
			return -1
		}
		if len(l) > len(r) {
			return 1
		}
		if cmp := strings.Compare(l, r); cmp != 0 {
			return cmp
		}
	}
	return strings.Compare(left, right)
}

// ScanBatch reads exactly one work-item page, with access/metadata checks. The
// scan-start timestamp is fixed for the whole traversal, including restarts.
// The worker may stage the result but must not expose it until Complete is true.
func (p *Provider) ScanBatch(ctx context.Context, token string, selected ws.Scope, started time.Time, rawCursor string) (ws.ScanBatch, error) {
	if started.IsZero() {
		return ws.ScanBatch{}, ws.ErrInvalid
	}
	scope, err := p.ResolveScope(ctx, token, selected.ID)
	if err != nil {
		return ws.ScanBatch{}, err
	}
	cursor, err := p.decodeCursor(rawCursor, scope, started)
	if err != nil {
		return ws.ScanBatch{}, err
	}
	q, _ := paging(cursor.Next)
	q.Set("archived", "true")
	q.Set("is_draft", "false")
	q.Set("fields", "id,project_id,identifier,name,description_html,is_draft,created_at,updated_at,completed_at,created_by_id,state_id,assignee_ids,label_ids")
	q.Set("expand", "state,assignees,labels")
	var page envelope[remoteItem]
	if err := p.get(ctx, token, p.base()+"/projects/"+scope.ID+"/work-items/", q, &page); err != nil {
		return ws.ScanBatch{}, err
	}
	nextCursor, err := next(page, cursor.Next)
	if err != nil {
		return ws.ScanBatch{}, err
	}
	result := ws.ScanBatch{Coverage: "COMPLETE"}
	for _, remote := range page.Data {
		id := string(remote.ID)
		if id == "" || (cursor.LastID != "" && compareID(id, cursor.LastID) <= 0) {
			return ws.ScanBatch{}, failure("SCHEMA_ERROR")
		}
		cursor.LastID = id
		if remote.Draft != nil && *remote.Draft {
			continue
		}
		item, err := p.normalize(remote, scope)
		if err != nil {
			return ws.ScanBatch{}, err
		}
		if item.Terminal && item.EndedAt == nil {
			result.Coverage = "PARTIAL"
			result.Unresolved++
			continue
		}
		if ws.Retained(item, started) {
			result.Items = append(result.Items, item)
		}
	}
	if nextCursor == "" {
		current, err := p.ResolveScope(ctx, token, scope.ID)
		if err != nil {
			return ws.ScanBatch{}, err
		}
		if digest(current.Name) != cursor.NameHash {
			return ws.ScanBatch{}, failure("SCHEMA_ERROR")
		}
		result.Complete = true
		return result, nil
	}
	if cursor.Pages+1 >= maxPages {
		return ws.ScanBatch{}, failure("UNSUPPORTED")
	}
	hash := digest(nextCursor)
	for _, seen := range cursor.Seen {
		if seen == hash {
			return ws.ScanBatch{}, failure("SCHEMA_ERROR")
		}
	}
	cursor.Pages++
	cursor.Seen = append(cursor.Seen, hash)
	cursor.Next = nextCursor
	encoded, err := json.Marshal(cursor)
	if err != nil {
		return ws.ScanBatch{}, ws.ErrInvalid
	}
	result.Next = base64.RawURLEncoding.EncodeToString(encoded)
	return result, nil
}

var _ ws.BatchProvider = (*Provider)(nil)
