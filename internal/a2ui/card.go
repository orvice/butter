package a2ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"
)

// Limits bound what one session's read-only cards may hold. Exceeding any of
// them rejects the whole batch; nothing is partially written.
type Limits struct {
	// MaxComponents per surface, after the batch is applied.
	MaxComponents int
	// MaxBatchBytes is the encoded size of one render_ui call's messages.
	MaxBatchBytes int
	// MaxSurfaces is how many read-only surfaces one session may hold.
	MaxSurfaces int
	// MaxSurfaceBytes bounds one surface's stored components and data, so
	// repeated updates cannot grow a card without limit.
	MaxSurfaceBytes int
	// MaxFallbackRunes bounds the readable fallback text.
	MaxFallbackRunes int
}

// DefaultLimits are the first-release limits (issue #350).
var DefaultLimits = Limits{
	MaxComponents:    100,
	MaxBatchBytes:    64 << 10,
	MaxSurfaces:      20,
	MaxSurfaceBytes:  256 << 10,
	MaxFallbackRunes: 2000,
}

// Anchor records where in the conversation a surface write happened.
type Anchor struct {
	ThreadID     string    `json:"thread_id,omitempty"`
	RunID        string    `json:"run_id,omitempty"`
	MessageID    string    `json:"message_id,omitempty"`
	InvocationID string    `json:"invocation_id,omitempty"`
	At           time.Time `json:"at"`
}

// Card is the authoritative state of one Result Card, as persisted in
// session state. Components keep first-appearance order.
type Card struct {
	ID         string         `json:"surface_id"`
	Revision   int            `json:"revision"`
	Components []Component    `json:"components"`
	Data       map[string]any `json:"data"`
	Fallback   string         `json:"fallback,omitempty"`
	Created    Anchor         `json:"created"`
	Updated    Anchor         `json:"updated"`
	// Deleted marks a tombstone: the card is gone, but its revision is kept
	// so the delete stays ordered after every earlier write.
	Deleted bool `json:"deleted,omitempty"`
}

// Live reports whether s is a card that still exists.
func (s *Card) Live() bool { return s != nil && !s.Deleted }

// Batch is one render_ui call: A2UI messages for one surface, applied in
// order. An empty SurfaceID creates a new surface whose ID the server
// assigns; otherwise the batch updates or deletes an existing one.
type Batch struct {
	SurfaceID string
	Messages  []map[string]any
	Fallback  string
}

// Result is a batch's validated outcome.
type Result struct {
	// Card is the new state to persist; a tombstone when the batch
	// deleted the card.
	Card    *Card
	Created bool
}

// Apply validates b against the session's current cards and computes the
// resulting surface. It writes nothing: callers persist Result and only then
// let a client see it. Every message is checked — catalog, properties,
// references, lifecycle, and limits — before any of them takes effect.
func Apply(cards map[string]*Card, b Batch, anchor Anchor, limits Limits) (Result, error) {
	if len(b.Messages) == 0 {
		return Result{}, errors.New("messages must contain at least one A2UI message")
	}
	encoded, err := json.Marshal(b.Messages)
	if err != nil {
		return Result{}, fmt.Errorf("messages are not valid JSON: %w", err)
	}
	if len(encoded) > limits.MaxBatchBytes {
		return Result{}, fmt.Errorf("messages are %d bytes; one call may send at most %d", len(encoded), limits.MaxBatchBytes)
	}
	if utf8.RuneCountInString(b.Fallback) > limits.MaxFallbackRunes {
		return Result{}, fmt.Errorf("fallback is longer than %d characters", limits.MaxFallbackRunes)
	}
	if err := validateLiteral(b.Fallback); err != nil {
		return Result{}, fmt.Errorf("fallback %w", err)
	}

	var next *Card
	create := b.SurfaceID == ""
	if create {
		if countLive(cards) >= limits.MaxSurfaces {
			return Result{}, fmt.Errorf("this conversation already has %d cards, the maximum; update or delete one instead", limits.MaxSurfaces)
		}
		if strings.TrimSpace(b.Fallback) == "" {
			return Result{}, errors.New("fallback is required when creating a card: a plain-text version for clients that cannot render it")
		}
		next = &Card{ID: newSurfaceID("card"), Data: map[string]any{}, Fallback: b.Fallback, Created: anchor}
	} else {
		prev, ok := cards[b.SurfaceID]
		if !ok || !prev.Live() {
			return Result{}, fmt.Errorf("surface_id %q is not a card in this conversation; omit surface_id to create a new card", b.SurfaceID)
		}
		next = prev.clone()
		if b.Fallback != "" {
			next.Fallback = b.Fallback
		}
	}

	deleted := false
	sawComponents := false
	for i, msg := range b.Messages {
		if deleted {
			return Result{}, fmt.Errorf("message %d: nothing may follow deleteSurface", i)
		}
		op, body, err := splitMessage(msg, next.ID)
		if err != nil {
			return Result{}, fmt.Errorf("message %d: %w", i, err)
		}
		switch op {
		case "createSurface":
			return Result{}, fmt.Errorf("message %d: createSurface is not accepted; omit surface_id and the server creates the surface", i)
		case "updateComponents":
			if err := next.upsertComponents(body); err != nil {
				return Result{}, fmt.Errorf("message %d: %w", i, err)
			}
			sawComponents = true
		case "updateDataModel":
			if err := next.updateData(body); err != nil {
				return Result{}, fmt.Errorf("message %d: %w", i, err)
			}
		case "deleteSurface":
			if create {
				return Result{}, fmt.Errorf("message %d: cannot delete a card that is being created", i)
			}
			deleted = true
		}
	}
	if create && !sawComponents {
		return Result{}, errors.New("a new card needs an updateComponents message with a \"root\" component")
	}

	prevRevision := 0
	if !create {
		prevRevision = cards[b.SurfaceID].Revision
	}
	if deleted {
		tomb := &Card{ID: next.ID, Revision: prevRevision + 1, Created: next.Created, Updated: anchor, Deleted: true}
		return Result{Card: tomb}, nil
	}
	if err := next.validateTree(limits); err != nil {
		return Result{}, err
	}
	next.Revision = prevRevision + 1
	next.Updated = anchor
	return Result{Card: next, Created: create}, nil
}

func countLive(cards map[string]*Card) int {
	n := 0
	for _, s := range cards {
		if s.Live() {
			n++
		}
	}
	return n
}

// splitMessage returns a message's single operation and its body, checking
// the optional version and surfaceId members against this batch.
func splitMessage(msg map[string]any, surfaceID string) (string, map[string]any, error) {
	op := ""
	var body map[string]any
	for key, v := range msg {
		switch key {
		case "version":
			version, _ := v.(string)
			if version != Version && version != "v0.9" {
				return "", nil, fmt.Errorf("version must be %q", Version)
			}
		case "createSurface", "updateComponents", "updateDataModel", "deleteSurface":
			if op != "" {
				return "", nil, errors.New("each message carries exactly one operation")
			}
			op = key
			m, ok := v.(map[string]any)
			if !ok {
				return "", nil, fmt.Errorf("%s must be an object", key)
			}
			body = m
		default:
			return "", nil, fmt.Errorf("unknown member %q; a message is one of updateComponents, updateDataModel, deleteSurface", key)
		}
	}
	if op == "" {
		return "", nil, errors.New("a message must be one of updateComponents, updateDataModel, deleteSurface")
	}
	if op == "createSurface" {
		return op, body, nil
	}
	if sid, present := body["surfaceId"]; present {
		if s, _ := sid.(string); s != surfaceID || surfaceID == "" {
			return "", nil, errors.New("surfaceId inside a message must be omitted or match surface_id")
		}
	}
	return op, body, nil
}

func (s *Card) upsertComponents(body map[string]any) error {
	for key := range body {
		if key != "components" && key != "surfaceId" {
			return fmt.Errorf("updateComponents has no member %q", key)
		}
	}
	list, ok := body["components"].([]any)
	if !ok || len(list) == 0 {
		return errors.New("updateComponents.components must be a non-empty array")
	}
	index := make(map[string]int, len(s.Components))
	for i, c := range s.Components {
		index[c.ID()] = i
	}
	seen := make(map[string]bool, len(list))
	for _, item := range list {
		raw, ok := item.(map[string]any)
		if !ok {
			return errors.New("every component must be an object")
		}
		c := Component(raw)
		if err := validateModelComponent(c); err != nil {
			return err
		}
		id := c.ID()
		if seen[id] {
			return fmt.Errorf("component id %q appears twice in one message", id)
		}
		seen[id] = true
		if i, exists := index[id]; exists {
			s.Components[i] = c
			continue
		}
		index[id] = len(s.Components)
		s.Components = append(s.Components, c)
	}
	return nil
}

func (s *Card) updateData(body map[string]any) error {
	for key := range body {
		if key != "path" && key != "value" && key != "surfaceId" {
			return fmt.Errorf("updateDataModel has no member %q", key)
		}
	}
	path := "/"
	if p, present := body["path"]; present {
		ps, ok := p.(string)
		if !ok {
			return errors.New("updateDataModel.path must be a string")
		}
		path = ps
	}
	tokens, err := parsePointer(path)
	if err != nil {
		return err
	}
	value, hasValue := body["value"]
	if hasValue {
		if err := validateDataValue(value); err != nil {
			return err
		}
	}
	if len(tokens) == 0 {
		if !hasValue {
			s.Data = map[string]any{}
			return nil
		}
		root, ok := value.(map[string]any)
		if !ok {
			return errors.New("the data model root must be an object")
		}
		s.Data = root
		return nil
	}
	parent := s.Data
	for _, tok := range tokens[:len(tokens)-1] {
		child, exists := parent[tok]
		if !exists || child == nil {
			m := map[string]any{}
			parent[tok] = m
			parent = m
			continue
		}
		m, ok := child.(map[string]any)
		if !ok {
			return fmt.Errorf("path %q goes through a non-object value; data paths may only traverse objects", path)
		}
		parent = m
	}
	last := tokens[len(tokens)-1]
	if hasValue {
		parent[last] = value
	} else {
		delete(parent, last)
	}
	return nil
}

// validateDataValue checks every string in a data value with the same rules
// as literal component text, since bound text renders the same way.
func validateDataValue(v any) error {
	switch value := v.(type) {
	case string:
		if err := validateLiteral(value); err != nil {
			return fmt.Errorf("data value %w", err)
		}
	case map[string]any:
		for _, item := range value {
			if err := validateDataValue(item); err != nil {
				return err
			}
		}
	case []any:
		for _, item := range value {
			if err := validateDataValue(item); err != nil {
				return err
			}
		}
	}
	return nil
}

// parsePointer splits an RFC 6901 JSON Pointer; "" and "/" address the root.
func parsePointer(path string) ([]string, error) {
	if path == "" || path == "/" {
		return nil, nil
	}
	if !strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("path %q must be a JSON Pointer starting with \"/\"", path)
	}
	parts := strings.Split(path[1:], "/")
	for i, p := range parts {
		if p == "" {
			return nil, fmt.Errorf("path %q has an empty segment", path)
		}
		parts[i] = strings.ReplaceAll(strings.ReplaceAll(p, "~1", "/"), "~0", "~")
	}
	return parts, nil
}

// validateTree checks the surface a batch produced: a root, resolvable
// child references, no cycles, and the size limits.
func (s *Card) validateTree(limits Limits) error {
	if len(s.Components) > limits.MaxComponents {
		return fmt.Errorf("a card may have at most %d components, this one would have %d", limits.MaxComponents, len(s.Components))
	}
	byID := make(map[string]Component, len(s.Components))
	for _, c := range s.Components {
		byID[c.ID()] = c
	}
	if _, ok := byID["root"]; !ok {
		return errors.New("a card needs a component with id \"root\"")
	}
	for _, c := range s.Components {
		for _, ref := range childRefs(c) {
			if _, ok := byID[ref]; !ok {
				return fmt.Errorf("component %q references %q, which does not exist", c.ID(), ref)
			}
		}
	}
	const (
		unvisited = iota
		visiting
		done
	)
	state := make(map[string]int, len(byID))
	var visit func(id string) error
	visit = func(id string) error {
		switch state[id] {
		case visiting:
			return fmt.Errorf("component %q is its own ancestor; components must form a tree", id)
		case done:
			return nil
		}
		state[id] = visiting
		for _, ref := range childRefs(byID[id]) {
			if err := visit(ref); err != nil {
				return err
			}
		}
		state[id] = done
		return nil
	}
	if err := visit("root"); err != nil {
		return err
	}
	encoded, err := json.Marshal(struct {
		C []Component    `json:"c"`
		D map[string]any `json:"d"`
	}{s.Components, s.Data})
	if err != nil {
		return fmt.Errorf("card is not valid JSON: %w", err)
	}
	if len(encoded) > limits.MaxSurfaceBytes {
		return fmt.Errorf("the card would hold %d bytes of components and data; the maximum is %d", len(encoded), limits.MaxSurfaceBytes)
	}
	return nil
}

func (s *Card) clone() *Card {
	raw, _ := json.Marshal(s)
	var out Card
	_ = json.Unmarshal(raw, &out)
	if out.Data == nil {
		out.Data = map[string]any{}
	}
	return &out
}

// Envelopes returns the messages that build s from nothing.
func (s *Card) Envelopes() []Envelope {
	return []Envelope{
		createEnvelope(s.ID),
		componentsEnvelope(s.ID, s.Components),
		dataEnvelope(s.ID, "/", dataOrEmpty(s.Data)),
	}
}

// Transition returns the messages that move a client from prev to next.
// A nil or deleted state means the card does not exist. The result depends
// only on the two persisted states, so it is the same wherever it is
// computed.
func Transition(id string, prev, next *Card) []Envelope {
	switch {
	case !prev.Live() && !next.Live():
		return nil
	case !prev.Live():
		return next.Envelopes()
	case !next.Live():
		return []Envelope{deleteEnvelope(id)}
	}
	var out []Envelope
	before := make(map[string]Component, len(prev.Components))
	for _, c := range prev.Components {
		before[c.ID()] = c
	}
	var changed []Component
	for _, c := range next.Components {
		if old, ok := before[c.ID()]; !ok || !jsonEqual(old, c) {
			changed = append(changed, c)
		}
	}
	if len(changed) > 0 {
		out = append(out, componentsEnvelope(id, changed))
	}
	if !jsonEqual(dataOrEmpty(prev.Data), dataOrEmpty(next.Data)) {
		out = append(out, dataEnvelope(id, "/", dataOrEmpty(next.Data)))
	}
	return out
}

func dataOrEmpty(d map[string]any) map[string]any {
	if d == nil {
		return map[string]any{}
	}
	return d
}

func jsonEqual(a, b any) bool {
	ra, errA := json.Marshal(a)
	rb, errB := json.Marshal(b)
	if errA != nil || errB != nil {
		return reflect.DeepEqual(a, b)
	}
	return string(ra) == string(rb)
}
