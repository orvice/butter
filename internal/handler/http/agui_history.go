package http

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	aguitypes "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/types"
	"github.com/gin-gonic/gin"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/workflow"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/a2ui"
	"go.orx.me/apps/butter/internal/a2uitool"
	"go.orx.me/apps/butter/internal/runtime/interrupt"
	"go.orx.me/apps/butter/internal/runtime/runner"
)

// aguiThreadHistory is the body of GET
// /api/agui/:agent_id/threads/:thread_id/messages.
type aguiThreadHistory struct {
	ThreadID string `json:"threadId"`
	// Messages is the conversation as AG-UI messages, oldest first.
	Messages []aguitypes.Message `json:"messages"`
	// Interrupts are the Interrupts still open, as the last run's
	// RUN_FINISHED reported them.
	Interrupts []aguitypes.Interrupt `json:"interrupts"`
	// Surfaces places each restorable A2UI surface in the assistant message
	// that produced it. The surfaces themselves come from the UI snapshot.
	Surfaces []aguiHistorySurface `json:"surfaces"`
}

type aguiHistorySurface struct {
	SurfaceID string `json:"surfaceId"`
	MessageID string `json:"messageId"`
}

// ThreadMessages handles GET /api/agui/:agent_id/threads/:thread_id/messages:
// the thread's conversation rebuilt from its persisted session. Like the UI
// snapshot it never starts a run and reads under the thread's session lease;
// a thread without a session, or bound to another caller, workspace or agent,
// answers an empty history.
func (h *AGUIHandler) ThreadMessages(c *gin.Context) {
	threadID, sess, release, ok := h.readThread(c)
	if !ok {
		return
	}
	defer release()
	c.JSON(http.StatusOK, aguiHistory(threadID, sess))
}

// aguiHistory rebuilds a thread's conversation from its session events in the
// shapes its runs streamed (aguiSink):
//   - A user turn is a user message: its text, or, once it carried images,
//     AG-UI content parts with the images inline, so an image-only turn is
//     kept too. So is an answer to a Human Input node, typed or implicit
//     (ADR-0002); a form's answer reads as the dashboard showed the
//     submission.
//   - Everything the agent produced between two user turns is one assistant
//     message, as each run streamed one: its text, the question of every
//     Interrupt it raised that is answered by now, and its tool calls, each
//     result following as a tool message. Without text, a workflow's Output
//     stands in, as it does for the live turn.
//   - What the stream hides stays hidden: thoughts, the request-input
//     handshake, and render_ui calls, whose card is placed through Surfaces.
//     An open Interrupt is reported in Interrupts, as RUN_FINISHED reported it.
//   - A tool call is kept only with its result or while the session still
//     awaits one from the client. Any other unanswered call is dropped, so a
//     client never answers a call the server would reject.
func aguiHistory(threadID string, sess session.Session) aguiThreadHistory {
	out := aguiThreadHistory{
		ThreadID:   threadID,
		Messages:   []aguitypes.Message{},
		Interrupts: []aguitypes.Interrupt{},
		Surfaces:   []aguiHistorySurface{},
	}
	if sess == nil {
		return out
	}
	b := newAGUIHistoryBuilder(sess)
	events := sess.Events()
	for i := 0; i < events.Len(); i++ {
		b.add(events.At(i))
	}
	b.flush()
	out.Messages = append(out.Messages, b.messages...)
	out.Surfaces = append(out.Surfaces, b.surfaces...)
	for _, p := range interrupt.Pending(sess) {
		out.Interrupts = append(out.Interrupts, aguitypes.Interrupt{ID: p.InterruptID, Reason: aguiInterruptReason, Message: p.Question})
	}
	return out
}

type aguiHistoryBuilder struct {
	// answered holds every FunctionResponse ID in the session.
	answered map[string]bool
	// awaited holds the client tool calls the session still waits on.
	awaited map[string]bool
	// open holds the Interrupts still unanswered.
	open map[string]bool
	// forms are the Human Input forms, by Interrupt ID.
	forms map[string]a2ui.Form
	// liveCards are the cards the UI snapshot restores.
	liveCards map[string]bool

	kept     map[string]bool
	placed   map[string]bool
	messages []aguitypes.Message
	surfaces []aguiHistorySurface
	turn     *aguiHistoryTurn
}

// aguiHistoryTurn is the assistant message being built: what one run showed.
type aguiHistoryTurn struct {
	id      string
	text    []string
	output  any
	calls   []aguitypes.ToolCall
	results []aguitypes.Message
	// hosts keeps a turn without text or calls: it raised an Interrupt or
	// produced a surface, which a client shows on it.
	hosts bool
}

func newAGUIHistoryBuilder(sess session.Session) *aguiHistoryBuilder {
	b := &aguiHistoryBuilder{
		answered:  map[string]bool{},
		awaited:   map[string]bool{},
		open:      interrupt.PendingIDs(sess),
		forms:     map[string]a2ui.Form{},
		liveCards: map[string]bool{},
		kept:      map[string]bool{},
		placed:    map[string]bool{},
	}
	events := sess.Events()
	for i := 0; i < events.Len(); i++ {
		ev := events.At(i)
		if ev == nil {
			continue
		}
		if form, ok := a2ui.FormOf(ev); ok {
			b.forms[form.InterruptID] = form
		}
		if ev.Content == nil {
			continue
		}
		for _, part := range ev.Content.Parts {
			if part != nil && part.FunctionResponse != nil && part.FunctionResponse.ID != "" {
				b.answered[part.FunctionResponse.ID] = true
			}
		}
	}
	for _, call := range interrupt.PendingToolCalls(sess) {
		b.awaited[call.ID] = true
	}
	for _, card := range a2ui.LiveCards(sess.State()) {
		b.liveCards[card.ID] = true
	}
	return b
}

func (b *aguiHistoryBuilder) add(ev *session.Event) {
	if ev == nil || ev.Partial {
		return
	}
	if ev.Author == "user" {
		b.addUser(ev)
		return
	}
	b.addAgent(ev)
}

// addUser closes the run before it: the user's text and images, a Human
// Input answer, or a client tool's result all start the next one.
func (b *aguiHistoryBuilder) addUser(ev *session.Event) {
	if ev.Content == nil {
		return
	}
	var content []aguitypes.InputContent
	var results []aguitypes.Message
	for _, part := range ev.Content.Parts {
		switch {
		case part == nil || part.Thought:
		case part.FunctionResponse != nil && part.FunctionResponse.Name == workflow.WorkflowInputFunctionCallName:
			if answer := b.answerText(part.FunctionResponse); answer != "" {
				content = append(content, aguiTextContent(answer))
			}
		case part.FunctionResponse != nil:
			if b.kept[part.FunctionResponse.ID] {
				results = append(results, aguiToolResultMessage(part.FunctionResponse))
			}
		case part.Text != "":
			content = append(content, aguiTextContent(part.Text))
		case aguiIsImage(part.InlineData):
			content = append(content, aguiImageContent(part.InlineData))
		}
	}
	b.flush()
	b.messages = append(b.messages, results...)
	if len(content) > 0 {
		b.messages = append(b.messages, aguitypes.Message{ID: ev.ID, Role: aguitypes.RoleUser, Content: aguiUserMessageContent(content)})
	}
}

// aguiUserMessageContent is a user message's content: plain text, its parts
// joined, while the turn carried only text, and AG-UI content parts in the
// order sent once it carried an image.
func aguiUserMessageContent(content []aguitypes.InputContent) any {
	text := make([]string, 0, len(content))
	for _, c := range content {
		if c.Type != aguitypes.InputContentTypeText {
			return content
		}
		text = append(text, c.Text)
	}
	return strings.Join(text, "\n\n")
}

func aguiTextContent(text string) aguitypes.InputContent {
	return aguitypes.InputContent{Type: aguitypes.InputContentTypeText, Text: text}
}

// aguiIsImage reports whether inline data is an image a user sent.
func aguiIsImage(blob *genai.Blob) bool {
	return blob != nil && len(blob.Data) > 0 && strings.HasPrefix(blob.MIMEType, "image/")
}

// aguiImageContent is an image as a client sends one: inline, as a base64
// data source.
func aguiImageContent(blob *genai.Blob) aguitypes.InputContent {
	return aguitypes.InputContent{
		Type: aguitypes.InputContentTypeImage,
		Source: &aguitypes.InputContentSource{
			Type:     aguitypes.InputContentSourceTypeData,
			Value:    base64.StdEncoding.EncodeToString(blob.Data),
			MimeType: blob.MIMEType,
		},
	}
}

// answerText is how a Human Input answer reads: a form's fields as the
// dashboard showed them, any other answer as its text.
func (b *aguiHistoryBuilder) answerText(fr *genai.FunctionResponse) string {
	var answer string
	switch payload := fr.Response[aguiRequestInputPayloadKey].(type) {
	case nil:
		return ""
	case string:
		answer = payload
	default:
		encoded, err := json.Marshal(payload)
		if err != nil {
			return ""
		}
		answer = string(encoded)
	}
	if form, ok := b.forms[fr.ID]; ok {
		return form.ReadableAnswer(answer)
	}
	return answer
}

func (b *aguiHistoryBuilder) addAgent(ev *session.Event) {
	if b.turn == nil {
		b.turn = &aguiHistoryTurn{id: ev.ID}
	}
	t := b.turn
	if ev.Output != nil {
		t.output = ev.Output
	}
	if ev.Content != nil {
		for _, part := range ev.Content.Parts {
			switch {
			case part == nil || part.Thought:
			case part.FunctionCall != nil:
				b.addCall(t, part.FunctionCall)
			case part.FunctionResponse != nil:
				if b.kept[part.FunctionResponse.ID] {
					t.results = append(t.results, aguiToolResultMessage(part.FunctionResponse))
				}
			case part.Text != "":
				t.text = append(t.text, part.Text)
			}
		}
	}
	b.placeCards(t, ev.Actions.StateDelta)
}

func (b *aguiHistoryBuilder) addCall(t *aguiHistoryTurn, fc *genai.FunctionCall) {
	switch {
	case fc.ID == "":
	case fc.Name == workflow.WorkflowInputFunctionCallName:
		t.hosts = true
		form, hasForm := b.forms[fc.ID]
		if b.open[fc.ID] {
			if hasForm {
				b.place(t, form.SurfaceID)
			}
			return
		}
		question := interrupt.QuestionOf(fc)
		if hasForm && form.Question != "" {
			question = form.Question
		}
		if question != "" {
			t.text = append(t.text, question)
		}
	case fc.Name == a2uitool.ToolName:
	case b.answered[fc.ID] || b.awaited[fc.ID]:
		args, err := json.Marshal(fc.Args)
		if err != nil || fc.Args == nil {
			args = []byte("{}")
		}
		t.calls = append(t.calls, aguitypes.ToolCall{
			ID:       fc.ID,
			Type:     aguitypes.ToolCallTypeFunction,
			Function: aguitypes.FunctionCall{Name: fc.Name, Arguments: string(args)},
		})
		b.kept[fc.ID] = true
	}
}

// placeCards places each live card on the turn whose state delta first
// carries it: the render_ui call that created it.
func (b *aguiHistoryBuilder) placeCards(t *aguiHistoryTurn, delta map[string]any) {
	var ids []string
	for key := range delta {
		if id, ok := a2ui.CardIDFromKey(key); ok && b.liveCards[id] && !b.placed[id] {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	for _, id := range ids {
		b.place(t, id)
	}
}

func (b *aguiHistoryBuilder) place(t *aguiHistoryTurn, surfaceID string) {
	if surfaceID == "" || b.placed[surfaceID] {
		return
	}
	b.placed[surfaceID] = true
	b.surfaces = append(b.surfaces, aguiHistorySurface{SurfaceID: surfaceID, MessageID: t.id})
	t.hosts = true
}

func (b *aguiHistoryBuilder) flush() {
	t := b.turn
	if t == nil {
		return
	}
	b.turn = nil
	text := strings.Join(t.text, "\n\n")
	if text == "" && t.output != nil {
		text = runner.RenderEventOutput(t.output)
	}
	if text == "" && len(t.calls) == 0 && !t.hosts {
		return
	}
	msg := aguitypes.Message{ID: t.id, Role: aguitypes.RoleAssistant, ToolCalls: t.calls}
	if text != "" {
		msg.Content = text
	}
	b.messages = append(b.messages, msg)
	b.messages = append(b.messages, t.results...)
}

// aguiToolResultMessage is a tool's result as the live stream sent it: the
// JSON of the FunctionResponse.
func aguiToolResultMessage(fr *genai.FunctionResponse) aguitypes.Message {
	content, err := json.Marshal(fr.Response)
	if err != nil {
		content = []byte("{}")
	}
	return aguitypes.Message{ID: "result:" + fr.ID, Role: aguitypes.RoleTool, ToolCallID: fr.ID, Content: string(content)}
}
