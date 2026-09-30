package a2ui

import (
	"encoding/json"
	"iter"
	"sort"
	"strings"

	"google.golang.org/adk/v2/session"
)

// UI records live in one namespace of ADK session state. Values are JSON
// strings: every session backend stores a string as-is, whereas nested
// documents come back from Mongo as bson.D rather than maps.
const (
	// StatePrefix namespaces every UI key. The AG-UI shared-state mirror
	// never includes it: clients cannot read or overwrite UI state through
	// STATE_* events.
	StatePrefix = "butter:a2ui:"
	// BindingKey holds the session's Binding.
	BindingKey = StatePrefix + "binding"
	// cardKeyPrefix + surface ID holds one Result Card's record.
	cardKeyPrefix = StatePrefix + "card:"
)

// CardKey is the session-state key of one card.
func CardKey(surfaceID string) string { return cardKeyPrefix + surfaceID }

// Binding is the authoritative context a session's UI belongs to. The AG-UI
// session key is (caller, thread) only — the same caller reusing a threadId
// under another workspace or agent lands on the same session — so UI state
// is exposed and accepted only when the request's full context matches the
// binding recorded when the session was created. A session without a
// binding predates A2UI and exposes no UI at all.
type Binding struct {
	Principal   string `json:"principal"`
	WorkspaceID string `json:"workspace_id"`
	AgentID     string `json:"agent_id"`
	ThreadID    string `json:"thread_id"`
}

// StateValue encodes b for session state.
func (b Binding) StateValue() string {
	raw, _ := json.Marshal(b)
	return string(raw)
}

// BindingOf reads the session's binding; ok is false for a session that
// has none (created before A2UI existed).
func BindingOf(st session.State) (Binding, bool) {
	if st == nil {
		return Binding{}, false
	}
	v, err := st.Get(BindingKey)
	if err != nil {
		return Binding{}, false
	}
	s, _ := v.(string)
	var b Binding
	if s == "" || json.Unmarshal([]byte(s), &b) != nil || b.Principal == "" {
		return Binding{}, false
	}
	return b, true
}

// Bound reports whether sess carries exactly this binding.
func Bound(sess session.Session, want Binding) bool {
	if sess == nil {
		return false
	}
	got, ok := BindingOf(sess.State())
	return ok && got == want
}

// StateValue encodes s for session state.
func (s *Card) StateValue() string {
	raw, _ := json.Marshal(s)
	return string(raw)
}

// CardIDFromKey returns the surface ID of the card a session-state key holds,
// and whether it holds one.
func CardIDFromKey(key string) (string, bool) {
	id, ok := strings.CutPrefix(key, cardKeyPrefix)
	return id, ok && id != ""
}

// DecodeCard reads one card record from a session-state value.
func DecodeCard(v any) (*Card, bool) {
	s, _ := v.(string)
	if s == "" {
		return nil, false
	}
	var out Card
	if err := json.Unmarshal([]byte(s), &out); err != nil || out.ID == "" {
		return nil, false
	}
	if out.Data == nil {
		out.Data = map[string]any{}
	}
	return &out, true
}

// Cards reads every card record, tombstones included, from session state.
func Cards(st session.State) map[string]*Card {
	cards := map[string]*Card{}
	if st == nil {
		return cards
	}
	collectCards(st.All(), cards)
	return cards
}

func collectCards(all iter.Seq2[string, any], into map[string]*Card) {
	for key, v := range all {
		id, ok := CardIDFromKey(key)
		if !ok {
			continue
		}
		if card, ok := DecodeCard(v); ok && card.ID == id {
			into[id] = card
		}
	}
}

// LiveCards returns the cards that still exist, oldest first.
func LiveCards(st session.State) []*Card {
	var out []*Card
	for _, c := range Cards(st) {
		if c.Live() {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].Created.At.Equal(out[j].Created.At) {
			return out[i].Created.At.Before(out[j].Created.At)
		}
		return out[i].ID < out[j].ID
	})
	return out
}
