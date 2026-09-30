package agent

import (
	"iter"

	"github.com/google/uuid"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/workflow"

	"go.orx.me/apps/butter/internal/a2ui"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// humanInputNode is the butter-owned Human Input node: it pauses the
// workflow (an Interrupt) by emitting a request-input event carrying the
// configured question, then exits. Handoff resume semantics (the engine
// default): the human's reply flows to the node's successor as its input.
// The Interrupt ID is unique per activation; the runner resolves which
// Interrupt a reply answers by scanning session events (ADR 0002), so the
// ID needs no stable form beyond uniqueness.
//
// A node with a form also freezes the form's binding into the request-input
// event (a2ui.Form.Attach) and appends field instructions to the question,
// so entry points that cannot render the form still say what to answer.
// The reply the successor receives is still one string either way.
type humanInputNode struct {
	workflow.BaseNode
	question string
	form     *agentsv1.HumanInputForm
}

func newHumanInputNode(name, question string, form *agentsv1.HumanInputForm, cfg workflow.NodeConfig) *humanInputNode {
	n := &humanInputNode{
		BaseNode: workflow.NewBaseNode(name, "", cfg),
		question: question,
	}
	if a2ui.HasForm(form) {
		n.form = form
	}
	return n
}

func (n *humanInputNode) Run(ctx agent.Context, input any) iter.Seq2[*session.Event, error] {
	return func(yield func(*session.Event, error) bool) {
		interruptID := n.Name() + "-" + uuid.NewString()
		message := n.question
		var form *a2ui.Form
		if n.form != nil {
			f := a2ui.NewForm(interruptID, n.question, n.form)
			form = &f
			message = n.question + "\n\n" + a2ui.Instructions(f.Fields)
		}
		ev := workflow.NewRequestInputEvent(ctx, session.RequestInput{
			InterruptID: interruptID,
			Message:     message,
			Payload:     input,
		})
		if form != nil {
			form.Attach(ev)
		}
		yield(ev, nil)
	}
}
