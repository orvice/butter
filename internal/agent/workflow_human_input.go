package agent

import (
	"iter"

	"github.com/google/uuid"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/workflow"
)

// humanInputNode is the butter-owned Human Input node: it pauses the
// workflow (an Interrupt) by emitting a request-input event carrying the
// configured question, then exits. Handoff resume semantics (the engine
// default): the human's reply flows to the node's successor as its input.
// The Interrupt ID is unique per activation; the runner resolves which
// Interrupt a reply answers by scanning session events (ADR 0002), so the
// ID needs no stable form beyond uniqueness.
type humanInputNode struct {
	workflow.BaseNode
	question string
}

func newHumanInputNode(name, question string, cfg workflow.NodeConfig) *humanInputNode {
	return &humanInputNode{
		BaseNode: workflow.NewBaseNode(name, "", cfg),
		question: question,
	}
}

func (n *humanInputNode) Run(ctx agent.Context, input any) iter.Seq2[*session.Event, error] {
	return func(yield func(*session.Event, error) bool) {
		yield(workflow.NewRequestInputEvent(ctx, session.RequestInput{
			InterruptID: n.Name() + "-" + uuid.NewString(),
			Message:     n.question,
			Payload:     input,
		}), nil)
	}
}
