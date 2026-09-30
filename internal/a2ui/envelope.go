package a2ui

// Envelope is one complete A2UI v0.9.1 server-to-client message. Exactly one
// of the operation members is set.
type Envelope struct {
	Version          string            `json:"version"`
	CreateSurface    *CreateSurface    `json:"createSurface,omitempty"`
	UpdateComponents *UpdateComponents `json:"updateComponents,omitempty"`
	UpdateDataModel  *UpdateDataModel  `json:"updateDataModel,omitempty"`
	DeleteSurface    *DeleteSurface    `json:"deleteSurface,omitempty"`
}

// CreateSurface opens a surface rendered with one catalog.
type CreateSurface struct {
	SurfaceID string `json:"surfaceId"`
	CatalogID string `json:"catalogId"`
}

// UpdateComponents upserts components into a surface by ID.
type UpdateComponents struct {
	SurfaceID  string      `json:"surfaceId"`
	Components []Component `json:"components"`
}

// UpdateDataModel replaces the value at Path; butter always sends a value.
type UpdateDataModel struct {
	SurfaceID string `json:"surfaceId"`
	Path      string `json:"path"`
	Value     any    `json:"value"`
}

// DeleteSurface removes a surface.
type DeleteSurface struct {
	SurfaceID string `json:"surfaceId"`
}

// Component is one catalog component instance: "id", "component", and the
// component's properties, exactly as it travels on the wire.
type Component map[string]any

// ID returns the component's ID.
func (c Component) ID() string {
	id, _ := c["id"].(string)
	return id
}

// Kind is what a surface is for the client.
type Kind string

const (
	// KindCard is a read-only result card a model rendered.
	KindCard Kind = "card"
	// KindForm is a Human Input form the server built for an Interrupt.
	KindForm Kind = "form"
)

// EventValue is the value of one butter.a2ui CUSTOM event: one envelope plus
// what a client needs to order it and place it.
//
// Revision is server-assigned per surface and grows with every persisted
// change; Seq orders the envelopes of one revision. A client applies an
// event only when (Revision, Seq) is past what it already applied for that
// surface, so a replayed or stale message never overwrites a newer one.
type EventValue struct {
	Version   string   `json:"version"`
	SurfaceID string   `json:"surfaceId"`
	Kind      Kind     `json:"kind"`
	Revision  int      `json:"revision"`
	Seq       int      `json:"seq"`
	ThreadID  string   `json:"threadId,omitempty"`
	RunID     string   `json:"runId,omitempty"`
	MessageID string   `json:"messageId,omitempty"`
	Envelope  Envelope `json:"envelope"`
	// Fallback is the readable text a client shows when it cannot render
	// the surface.
	Fallback string `json:"fallback,omitempty"`
	// Form binds a form surface to the Interrupt it answers.
	Form *FormView `json:"form,omitempty"`
}

func createEnvelope(surfaceID string) Envelope {
	return Envelope{Version: Version, CreateSurface: &CreateSurface{SurfaceID: surfaceID, CatalogID: CatalogID}}
}

func componentsEnvelope(surfaceID string, components []Component) Envelope {
	return Envelope{Version: Version, UpdateComponents: &UpdateComponents{SurfaceID: surfaceID, Components: components}}
}

func dataEnvelope(surfaceID, path string, value any) Envelope {
	return Envelope{Version: Version, UpdateDataModel: &UpdateDataModel{SurfaceID: surfaceID, Path: path, Value: value}}
}

func deleteEnvelope(surfaceID string) Envelope {
	return Envelope{Version: Version, DeleteSurface: &DeleteSurface{SurfaceID: surfaceID}}
}
