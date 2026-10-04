package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/adk/v2/session"

	"go.orx.me/apps/butter/internal/a2ui"
)

// A2UI over AG-UI: the handler's share. It binds a new session to the
// caller's full context, resolves form submissions inside the session lease,
// and serves the read-only UI snapshot, which takes no lease (agui_read.go).
// See internal/a2ui and docs/api.md.

// aguiFormEntry is one resume entry that submits a form. Its FunctionResponse
// payload is filled in only after the submission is validated against the
// session, inside the lease.
type aguiFormEntry struct {
	partIndex   int
	interruptID string
	submission  a2ui.Submission
}

// aguiUIContext is what the run knows about the session's UI once the lease
// is held: the session as it stood before the run and whether its binding
// matches this caller.
type aguiUIContext struct {
	sess  session.Session
	bound bool
}

// aguiBinding is the authoritative UI binding for a request: the
// authenticated caller, the workspace, the agent, and the client's thread.
func aguiBinding(ctx context.Context, workspaceID, agentID, threadID string) a2ui.Binding {
	return a2ui.Binding{
		Principal:   aguiUserID(ctx),
		WorkspaceID: workspaceID,
		AgentID:     agentID,
		ThreadID:    threadID,
	}
}

// errThreadUnavailable refuses a run on a threadId the caller may not use:
// another user's, or the caller's own from another workspace.
var errThreadUnavailable = errors.New("threadId is not available; start a new thread")

// errThreadOfAnotherAgent refuses a run on the caller's own thread when the
// thread is bound to another agent, so one thread's history is one agent's.
var errThreadOfAnotherAgent = errors.New("threadId belongs to another agent; start a new thread")

// existingThread admits a run on a thread whose session already exists, as
// long as that session belongs to the request's workspace and, when it
// carries a binding, the binding names the route's agent.
func existingThread(sess session.Session, rc *aguiRunContext, binding a2ui.Binding) (*aguiUIContext, *aguiRefusal) {
	if ws, ok := sess.(interface{ WorkspaceID() string }); ok {
		if id := ws.WorkspaceID(); id != "" && id != rc.ctxInfo.GetWorkspaceId() {
			return nil, refuseAGUI(http.StatusForbidden, errThreadUnavailable)
		}
	}
	if held, ok := a2ui.BindingOf(sess.State()); ok {
		// An agent_id is unique only within its workspace, so the binding's
		// workspace is checked too, for a store that does not report one.
		if held.WorkspaceID != binding.WorkspaceID {
			return nil, refuseAGUI(http.StatusForbidden, errThreadUnavailable)
		}
		if held.AgentID != binding.AgentID {
			return nil, refuseAGUI(http.StatusForbidden, errThreadOfAnotherAgent)
		}
	}
	return &aguiUIContext{sess: sess, bound: a2ui.Bound(sess, binding)}, nil
}

// Form rejection codes, by submit error kind.
var aguiFormErrorCodes = map[a2ui.SubmitErrorKind]struct {
	status int
	code   string
}{
	a2ui.SubmitUnknown:  {http.StatusBadRequest, "form_unknown"},
	a2ui.SubmitAnswered: {http.StatusConflict, "form_answered"},
	a2ui.SubmitStale:    {http.StatusConflict, "form_stale"},
	a2ui.SubmitInvalid:  {http.StatusUnprocessableEntity, "form_invalid"},
}

// resolveFormSubmissions validates every form submission in the request
// against the session and fills in the answer each one delivers. A rejected
// submission fails the whole request before the stream opens: nothing is
// appended and the agent does not run. The client keeps the draft and shows
// the message and per-field errors; the code says what happened.
func resolveFormSubmissions(rc *aguiRunContext, ui *aguiUIContext) *aguiRefusal {
	if len(rc.formEntries) == 0 {
		return nil
	}
	if !ui.bound {
		// A thread without a binding (created before A2UI) has no forms.
		unknown := aguiFormErrorCodes[a2ui.SubmitUnknown]
		return &aguiRefusal{status: unknown.status, body: aguiCodedError{Error: "unknown or expired form", Code: unknown.code}}
	}
	for _, entry := range rc.formEntries {
		answer, err := a2ui.Resolve(ui.sess, entry.interruptID, entry.submission)
		if err != nil {
			var submitErr *a2ui.SubmitError
			if !errors.As(err, &submitErr) {
				return &aguiRefusal{status: http.StatusInternalServerError, body: aguiCodedError{Error: err.Error()}}
			}
			kind := aguiFormErrorCodes[submitErr.Kind]
			return &aguiRefusal{status: kind.status, body: aguiCodedError{
				Error: submitErr.Message, Code: kind.code, FieldErrors: submitErr.FieldErrors,
			}}
		}
		fr := rc.parts[entry.partIndex].FunctionResponse
		fr.Response = map[string]any{aguiRequestInputPayloadKey: answer}
	}
	return nil
}

// aguiUISnapshot is the body of GET /api/agui/:agent_id/threads/:thread_id/ui.
type aguiUISnapshot struct {
	Version   string             `json:"version"`
	CatalogID string             `json:"catalogId"`
	ThreadID  string             `json:"threadId"`
	Surfaces  []aguiSnapshotItem `json:"surfaces"`
	// Running names the run in flight. The surfaces are then those the run
	// found, less the forms its first turn answered.
	Running *aguiRunning `json:"running,omitempty"`
}

// aguiSnapshotItem is one surface in its current, authoritative form: the
// envelopes that rebuild it, and where it came from.
type aguiSnapshotItem struct {
	SurfaceID string          `json:"surfaceId"`
	Kind      a2ui.Kind       `json:"kind"`
	Revision  int             `json:"revision"`
	RunID     string          `json:"runId,omitempty"`
	MessageID string          `json:"messageId,omitempty"`
	Fallback  string          `json:"fallback,omitempty"`
	Form      *a2ui.FormView  `json:"form,omitempty"`
	Envelopes []a2ui.Envelope `json:"envelopes"`
}

// UISnapshot handles GET /api/agui/:agent_id/threads/:thread_id/ui: the
// current read-only cards and unanswered forms of one thread, rebuilt from
// the persisted session. It never starts a run, and never waits for one:
// during a run it answers with the surfaces as the run found them, and the
// run (readThread). A thread without a session, without a binding, or bound
// to another caller, workspace or agent answers with an empty snapshot, so
// the endpoint reveals nothing about threads the caller does not own.
func (h *AGUIHandler) UISnapshot(c *gin.Context) {
	read, ok := h.readThread(c)
	if !ok {
		return
	}

	snap := aguiUISnapshot{
		Version:   a2ui.Version,
		CatalogID: a2ui.CatalogID,
		ThreadID:  read.threadID,
		Surfaces:  []aguiSnapshotItem{},
		Running:   read.running,
	}
	sess := read.sess
	if sess == nil {
		c.JSON(http.StatusOK, snap)
		return
	}
	for _, card := range a2ui.LiveCards(sess.State()) {
		snap.Surfaces = append(snap.Surfaces, aguiSnapshotItem{
			SurfaceID: card.ID,
			Kind:      a2ui.KindCard,
			Revision:  card.Revision,
			RunID:     card.Created.RunID,
			MessageID: card.Created.MessageID,
			Fallback:  card.Fallback,
			Envelopes: card.Envelopes(),
		})
	}
	for _, form := range a2ui.PendingForms(sess) {
		snap.Surfaces = append(snap.Surfaces, aguiSnapshotItem{
			SurfaceID: form.SurfaceID,
			Kind:      a2ui.KindForm,
			Revision:  form.Revision,
			Fallback:  form.Fallback(),
			Form:      form.View(),
			Envelopes: form.Envelopes(),
		})
	}
	c.JSON(http.StatusOK, snap)
}
