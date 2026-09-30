// Package a2ui is butter's A2UI v0.9.1 layer for AG-UI Chat: the
// butter-basic-v1 catalog and its validation, the read-only result cards a
// model renders through the render_ui tool, and the forms a Workflow Agent's
// Human Input node presents.
//
// AG-UI stays the transport. Butter carries each A2UI message in a CUSTOM
// event named EventName (a Butter extension, not an A2UI-standard AG-UI
// binding) and only for clients that declared the capability in
// forwardedProps. Everything the client sees is derived from persisted
// session data — cards from a dedicated namespace of ADK session state,
// forms from the request-input events that paused the workflow — so a Pod
// switch or restart neither loses UI nor creates a second Interrupt store
// (ADR-0002).
package a2ui

import (
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/google/uuid"
)

const (
	// Version is the A2UI protocol version butter speaks, and the "version"
	// member of every envelope it sends.
	Version = "v0.9.1"
	// CatalogID identifies the component catalog butter's server and
	// dashboard version together. It is an identifier, never a download.
	CatalogID = "butter-basic-v1"
	// EventName is the AG-UI CUSTOM event name carrying one A2UI envelope.
	EventName = "butter.a2ui"
	// CapabilityKey is the forwardedProps member a client declares A2UI
	// support with: {"version": "v0.9.1", "catalogs": ["butter-basic-v1"]}.
	CapabilityKey = "butterA2UI"
)

// Negotiate reads a client's A2UI declaration from AG-UI forwardedProps. The
// declaration only selects what the server already ships; it can never
// upload components or schemas. ok is false — plain chat, no error — when
// the client declares nothing or asks for a version or catalog butter does
// not serve. A declaration that is not shaped like one is an error, so a
// client is never left believing a malformed request was understood.
func Negotiate(forwardedProps any) (ok bool, err error) {
	props, isMap := forwardedProps.(map[string]any)
	if !isMap {
		return false, nil
	}
	raw, present := props[CapabilityKey]
	if !present || raw == nil {
		return false, nil
	}
	decl, isMap := raw.(map[string]any)
	if !isMap {
		return false, errors.New("forwardedProps.butterA2UI must be an object")
	}
	version, isString := decl["version"].(string)
	if !isString {
		return false, errors.New("forwardedProps.butterA2UI.version must be a string")
	}
	rawCatalogs, isList := decl["catalogs"].([]any)
	if !isList {
		return false, errors.New("forwardedProps.butterA2UI.catalogs must be an array of catalog IDs")
	}
	catalogs := make([]string, 0, len(rawCatalogs))
	for _, c := range rawCatalogs {
		id, isString := c.(string)
		if !isString {
			return false, errors.New("forwardedProps.butterA2UI.catalogs must be an array of catalog IDs")
		}
		catalogs = append(catalogs, id)
	}
	return version == Version && slices.Contains(catalogs, CatalogID), nil
}

// newSurfaceID mints a server-assigned surface ID: prefix plus 16 random hex
// digits, safe as a session-state key segment.
func newSurfaceID(prefix string) string {
	return prefix + "-" + strings.ReplaceAll(uuid.NewString(), "-", "")[:16]
}

// Run is one AG-UI run for which A2UI is live: the client negotiated it and
// the session's binding matches the caller. It rides the run's context so
// the render_ui tool is offered, and knows which message it belongs to, for
// exactly this run.
type Run struct {
	ThreadID  string
	RunID     string
	MessageID string
}

type runKey struct{}

// WithRun returns a context carrying run.
func WithRun(ctx context.Context, run *Run) context.Context {
	return context.WithValue(ctx, runKey{}, run)
}

// RunFrom returns the A2UI run carried by ctx, if any.
func RunFrom(ctx context.Context) (*Run, bool) {
	run, ok := ctx.Value(runKey{}).(*Run)
	return run, ok && run != nil
}
