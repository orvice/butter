package pibox

import (
	"strings"
	"testing"

	adkrunner "google.golang.org/adk/v2/runner"

	"go.orx.me/apps/butter/internal/a2ui"
	internalagent "go.orx.me/apps/butter/internal/agent"
	"go.orx.me/apps/butter/internal/runtime/eventcommit"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestPiCardsThroughCompositeFactory(t *testing.T) {
	for _, kind := range []agentsv1.AgentType{agentsv1.AgentType_AGENT_TYPE_PARALLEL, agentsv1.AgentType_AGENT_TYPE_WORKFLOW} {
		t.Run(kind.String(), func(t *testing.T) {
			box := newHostPi()
			bridge, h := hostHarness(t, box)
			child := piAgentProto("box-1", "")
			root := &agentsv1.Agent{AgentId: "root", Name: "root", WorkspaceId: "ws-1", Type: kind, ChildAgentIds: []string{child.AgentId}, Config: &agentsv1.AgentConfig{}}
			if kind == agentsv1.AgentType_AGENT_TYPE_WORKFLOW {
				root.Config.Workflow = &agentsv1.WorkflowConfig{
					Nodes: []*agentsv1.WorkflowNode{{Name: "work", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, AgentId: child.AgentId}},
					Edges: []*agentsv1.WorkflowEdge{{From: "START", To: "work"}},
				}
			}
			ag, err := internalagent.NewFromProtoWithToolsetFactory(t.Context(), root, nil, nil, nil, nil, nil, nil, &internalagent.BoxAgentBuilders{Pi: AgentBuilder(bridge.factory)}, internalagent.AgentPool{child.AgentId: child})
			if err != nil {
				t.Fatal(err)
			}
			r, err := adkrunner.New(adkrunner.Config{AppName: "test-app", Agent: ag, SessionService: eventcommit.Service(a2ui.PersistCards(h.sessions))})
			if err != nil {
				t.Fatal(err)
			}
			h.runner = r
			if output, err := h.turn(uiContext(t), textContent("show status")); err != nil || !strings.Contains(output, "hello from pi") {
				t.Fatalf("turn=(%q,%v)", output, err)
			}
			if len(a2ui.LiveCards(storedSession(t, h).State())) != 1 {
				t.Fatal("composite did not persist the Pi card")
			}
		})
	}
}
