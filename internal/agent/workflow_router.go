package agent

import (
	"fmt"
	"iter"
	"strings"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/workflow"
)

// routerNode is the butter-owned Router: it matches its input text against
// the route labels of its outgoing edges (trimmed, case-insensitive exact
// match) and stamps the winning label on the event's route tags, steering
// the branch taken. Unmatched input carries no route tag, so the default
// edge fires. ADK has no built-in way for config-driven graphs to produce
// route tags (ADR 0001).
type routerNode struct {
	workflow.BaseNode
	labels []string
}

func newRouterNode(name string, labels []string, cfg workflow.NodeConfig) *routerNode {
	return &routerNode{
		BaseNode: workflow.NewBaseNode(name, "", cfg),
		labels:   labels,
	}
}

func (n *routerNode) Run(ctx agent.Context, input any) iter.Seq2[*session.Event, error] {
	return func(yield func(*session.Event, error) bool) {
		text, ok := input.(string)
		if !ok {
			yield(nil, fmt.Errorf("router %q: input must be text, got %T", n.Name(), input))
			return
		}

		event := session.NewEvent(ctx, ctx.InvocationID())
		// The input passes through unchanged so the chosen branch receives
		// the same value the router saw.
		event.Output = input
		if label, ok := matchRouteLabel(text, n.labels); ok {
			// Stamp the label as configured on the edge: the engine
			// compares route tags to edge labels verbatim.
			event.Routes = []string{label}
		}
		yield(event, nil)
	}
}

// matchRouteLabel returns the first label matching the input text under the
// Router's matching rule — trimmed, case-insensitive exact match — and
// whether any matched. The returned label keeps its configured form.
func matchRouteLabel(text string, labels []string) (string, bool) {
	for _, label := range labels {
		if strings.EqualFold(strings.TrimSpace(text), strings.TrimSpace(label)) {
			return label, true
		}
	}
	return "", false
}
