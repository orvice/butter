package agent

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/workflowagent"
	"google.golang.org/adk/v2/workflow"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// WorkflowStartNodeName is the reserved sentinel used in edge `from` fields
// to mark the graph's entry point. It maps onto workflow.Start.
const WorkflowStartNodeName = "START"

// ValidateWorkflowAgent checks the workflow graphs in an agent config tree.
// It is pure proto validation — no models or toolsets are constructed — so
// the service layer can reject a bad graph at save time. Non-workflow agents
// pass through; sub-agents are validated recursively.
func ValidateWorkflowAgent(pb *agentsv1.Agent) error {
	if pb == nil {
		return nil
	}
	if pb.GetType() == agentsv1.AgentType_AGENT_TYPE_WORKFLOW {
		if err := validateWorkflowGraph(pb); err != nil {
			return fmt.Errorf("agent %q: %w", pb.GetName(), err)
		}
	}
	for _, sub := range pb.GetSubAgents() {
		if err := ValidateWorkflowAgent(sub); err != nil {
			return err
		}
	}
	return nil
}

// validateWorkflowGraph checks a single WORKFLOW-type agent's graph: node
// names are unique and not reserved, kinds are known, AGENT nodes reference
// declared sub-agents, edges reference declared nodes, and the graph has an
// entry edge from START.
func validateWorkflowGraph(pb *agentsv1.Agent) error {
	wf := pb.GetConfig().GetWorkflow()
	if len(wf.GetNodes()) == 0 {
		return fmt.Errorf("a workflow agent requires a workflow config with at least one node")
	}

	subAgentNames := make(map[string]struct{}, len(pb.GetSubAgents()))
	for _, sub := range pb.GetSubAgents() {
		subAgentNames[sub.GetName()] = struct{}{}
	}

	nodeNames := make(map[string]struct{}, len(wf.GetNodes()))
	for _, n := range wf.GetNodes() {
		name := n.GetName()
		if name == "" {
			return fmt.Errorf("workflow node has no name")
		}
		if name == WorkflowStartNodeName {
			return fmt.Errorf("workflow node name %q is reserved for the entry sentinel", WorkflowStartNodeName)
		}
		if _, exists := nodeNames[name]; exists {
			return fmt.Errorf("duplicate node name %q", name)
		}
		nodeNames[name] = struct{}{}

		switch n.GetKind() {
		case agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT:
			if n.GetAgent() == "" {
				return fmt.Errorf("workflow node %q: an AGENT node requires an agent reference", name)
			}
			if _, ok := subAgentNames[n.GetAgent()]; !ok {
				return fmt.Errorf("workflow node %q references sub-agent %q, which is not declared in sub_agents", name, n.GetAgent())
			}
		case agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_HUMAN_INPUT,
			agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_ROUTER,
			agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_JOIN:
			// Schema-valid kinds; execution support arrives in later slices.
		default:
			return fmt.Errorf("workflow node %q: kind must be one of AGENT, HUMAN_INPUT, ROUTER, JOIN", name)
		}

		if n.GetParallelWorker() && n.GetKind() != agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT {
			return fmt.Errorf("workflow node %q: parallel_worker is only supported on AGENT nodes", name)
		}
	}

	hasEntry := false
	hasDefaultEdge := make(map[string]bool)
	for _, e := range wf.GetEdges() {
		if e.GetFrom() == WorkflowStartNodeName {
			hasEntry = true
		} else if _, ok := nodeNames[e.GetFrom()]; !ok {
			return fmt.Errorf("edge %q -> %q references unknown node %q", e.GetFrom(), e.GetTo(), e.GetFrom())
		}
		if _, ok := nodeNames[e.GetTo()]; !ok {
			return fmt.Errorf("edge %q -> %q references unknown node %q", e.GetFrom(), e.GetTo(), e.GetTo())
		}
		if e.GetIsDefault() && e.GetRoute() != "" {
			return fmt.Errorf("edge %q -> %q: route and is_default are mutually exclusive", e.GetFrom(), e.GetTo())
		}
		if e.GetIsDefault() {
			hasDefaultEdge[e.GetFrom()] = true
		}
	}
	if !hasEntry {
		return fmt.Errorf("a workflow graph requires at least one entry edge from %q", WorkflowStartNodeName)
	}

	// An unmatched Router with no default edge dead-ends silently in the ADK
	// engine, so every Router must have a default outgoing edge.
	for _, n := range wf.GetNodes() {
		if n.GetKind() == agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_ROUTER && !hasDefaultEdge[n.GetName()] {
			return fmt.Errorf("router node %q requires a default outgoing edge (is_default: true) to catch unmatched input", n.GetName())
		}
	}

	return nil
}

// newWorkflowAgent builds an ADK workflow agent from a WORKFLOW-type proto
// config. subAgents are the already-built sub-agents of pb (including
// resolved remote agents); AGENT nodes reference them by name.
func newWorkflowAgent(pb *agentsv1.Agent, subAgents []agent.Agent) (agent.Agent, error) {
	if err := validateWorkflowGraph(pb); err != nil {
		return nil, err
	}
	wf := pb.GetConfig().GetWorkflow()

	built := make(map[string]agent.Agent, len(subAgents))
	for _, sa := range subAgents {
		built[sa.Name()] = sa
	}

	// Routers match against the route labels of their outgoing edges, so
	// collect labels per source node before building nodes.
	outgoingLabels := make(map[string][]string)
	for _, e := range wf.GetEdges() {
		if e.GetRoute() != "" {
			outgoingLabels[e.GetFrom()] = append(outgoingLabels[e.GetFrom()], e.GetRoute())
		}
	}

	nodes := make(map[string]workflow.Node, len(wf.GetNodes()))
	for _, n := range wf.GetNodes() {
		switch n.GetKind() {
		case agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT:
			sa, ok := built[n.GetAgent()]
			if !ok {
				return nil, fmt.Errorf("workflow node %q: sub-agent %q not found", n.GetName(), n.GetAgent())
			}
			if n.GetParallelWorker() {
				// The engine's parallel-worker mechanism is a wrapper node;
				// retry/timeout options live on the wrapper (the wrapped node
				// must not carry a retry policy).
				inner, err := workflow.NewAgentNode(sa, workflow.NodeConfig{})
				if err != nil {
					return nil, fmt.Errorf("workflow node %q: %w", n.GetName(), err)
				}
				node, err := newParallelWorkerNode(n.GetName(), inner, workflowNodeConfig(n))
				if err != nil {
					return nil, fmt.Errorf("workflow node %q: %w", n.GetName(), err)
				}
				nodes[n.GetName()] = node
				continue
			}
			node, err := workflow.NewAgentNode(sa, workflowNodeConfig(n))
			if err != nil {
				return nil, fmt.Errorf("workflow node %q: %w", n.GetName(), err)
			}
			nodes[n.GetName()] = node
		case agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_ROUTER:
			nodes[n.GetName()] = newRouterNode(n.GetName(), outgoingLabels[n.GetName()], workflowNodeConfig(n))
		case agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_JOIN:
			nodes[n.GetName()] = workflow.NewJoinNode(n.GetName())
		default:
			return nil, fmt.Errorf("workflow node %q: kind %s is not supported yet", n.GetName(), n.GetKind())
		}
	}

	edges := make([]workflow.Edge, 0, len(wf.GetEdges()))
	for _, e := range wf.GetEdges() {
		var from workflow.Node
		if e.GetFrom() == WorkflowStartNodeName {
			from = workflow.Start
		} else {
			from = nodes[e.GetFrom()]
		}
		to := nodes[e.GetTo()]
		if from == nil || to == nil {
			return nil, fmt.Errorf("workflow edge %q -> %q: references an undeclared node", e.GetFrom(), e.GetTo())
		}
		var route workflow.Route
		switch {
		case e.GetIsDefault():
			route = workflow.Default
		case e.GetRoute() != "":
			route = workflow.StringRoute(e.GetRoute())
		}
		edges = append(edges, workflow.Edge{From: from, To: to, Route: route})
	}

	return workflowagent.New(workflowagent.Config{
		Name:        pb.GetName(),
		Description: pb.GetDescription(),
		SubAgents:   subAgents,
		Edges:       edges,
	})
}

// parallelWorkerNode wraps the engine's ParallelWorker to coerce text input:
// in a config-driven graph the upstream node is typically an agent whose
// list-shaped answer arrives as JSON text, while the engine requires a real
// slice. A string input that parses as a JSON array becomes one; anything
// else falls through to the engine's own validation.
type parallelWorkerNode struct {
	*workflow.ParallelWorker
}

func newParallelWorkerNode(name string, wrapped workflow.Node, cfg workflow.NodeConfig) (*parallelWorkerNode, error) {
	pw, err := workflow.NewParallelWorker(name, wrapped, 0, cfg)
	if err != nil {
		return nil, err
	}
	return &parallelWorkerNode{ParallelWorker: pw}, nil
}

func (n *parallelWorkerNode) ValidateInput(input any) (any, error) {
	if s, ok := input.(string); ok {
		var items []any
		if err := json.Unmarshal([]byte(strings.TrimSpace(s)), &items); err == nil {
			return items, nil
		}
	}
	return n.ParallelWorker.ValidateInput(input)
}

// workflowNodeConfig translates the serializable node options onto the ADK
// NodeConfig. The parallel_worker option is deliberately not mapped to
// NodeConfig.ParallelWorker (inert in the engine) — it selects the
// ParallelWorker wrapper node instead.
func workflowNodeConfig(n *agentsv1.WorkflowNode) workflow.NodeConfig {
	cfg := workflow.NodeConfig{}
	if n.GetTimeoutSeconds() > 0 {
		cfg.Timeout = time.Duration(n.GetTimeoutSeconds()) * time.Second
	}
	if r := n.GetRetry(); r != nil {
		rc := workflow.DefaultRetryConfig()
		if r.GetMaxAttempts() > 0 {
			rc.MaxAttempts = int(r.GetMaxAttempts())
		}
		if r.GetInitialDelaySeconds() > 0 {
			rc.InitialDelay = time.Duration(r.GetInitialDelaySeconds()) * time.Second
		}
		if r.GetMaxDelaySeconds() > 0 {
			rc.MaxDelay = time.Duration(r.GetMaxDelaySeconds()) * time.Second
		}
		if r.GetBackoffFactor() > 0 {
			rc.BackoffFactor = r.GetBackoffFactor()
		}
		cfg.RetryConfig = rc
	}
	return cfg
}
