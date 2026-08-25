package agent

import (
	"context"
	"strings"
	"testing"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/protobuf/proto"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func validPiAgent() *agentsv1.Agent {
	return &agentsv1.Agent{
		Name:    "pi-coder",
		AgentId: "pi-coder",
		Type:    agentsv1.AgentType_AGENT_TYPE_PI,
		Config: &agentsv1.AgentConfig{
			Pi: &agentsv1.PiAgentConfig{
				ButterboxId:   "box-1",
				WorkingDir:    "projects/demo",
				Provider:      "anthropic",
				Model:         "claude-fable-5",
				ThinkingLevel: "high",
			},
		},
	}
}

func TestValidatePiAgent_Valid(t *testing.T) {
	if err := ValidatePiAgent(validPiAgent()); err != nil {
		t.Fatalf("valid pi agent rejected: %v", err)
	}
	// Minimal: only the box reference is required.
	minimal := &agentsv1.Agent{
		Name: "pi-min",
		Type: agentsv1.AgentType_AGENT_TYPE_PI,
		Config: &agentsv1.AgentConfig{
			Pi: &agentsv1.PiAgentConfig{ButterboxId: "box-1"},
		},
	}
	if err := ValidatePiAgent(minimal); err != nil {
		t.Fatalf("minimal pi agent rejected: %v", err)
	}
	// Non-PI agents pass through untouched.
	llm := &agentsv1.Agent{
		Name:   "llm",
		Type:   agentsv1.AgentType_AGENT_TYPE_LLM,
		Config: &agentsv1.AgentConfig{Instruction: "be nice"},
	}
	if err := ValidatePiAgent(llm); err != nil {
		t.Fatalf("non-pi agent rejected: %v", err)
	}
	if err := ValidatePiAgent(nil); err != nil {
		t.Fatalf("nil agent rejected: %v", err)
	}
}

func TestValidatePiAgent_RequiresPiConfig(t *testing.T) {
	pb := validPiAgent()
	pb.Config.Pi = nil
	if err := ValidatePiAgent(pb); err == nil || !strings.Contains(err.Error(), "config.pi") {
		t.Fatalf("expected missing config.pi error, got %v", err)
	}
	pb.Config = nil
	if err := ValidatePiAgent(pb); err == nil {
		t.Fatal("expected error for pi agent without config")
	}
}

func TestValidatePiAgent_RequiresButterBoxID(t *testing.T) {
	pb := validPiAgent()
	pb.Config.Pi.ButterboxId = "  "
	if err := ValidatePiAgent(pb); err == nil || !strings.Contains(err.Error(), "butterbox_id") {
		t.Fatalf("expected butterbox_id error, got %v", err)
	}
}

func TestValidatePiAgent_RejectsNegativeMaxRunSeconds(t *testing.T) {
	pb := validPiAgent()
	pb.Config.Pi.MaxRunSeconds = proto.Int32(-1)
	if err := ValidatePiAgent(pb); err == nil || !strings.Contains(err.Error(), "max_run_seconds") {
		t.Fatalf("expected max_run_seconds error, got %v", err)
	}
	pb.Config.Pi.MaxRunSeconds = proto.Int32(0) // explicit 0 = unlimited, allowed
	if err := ValidatePiAgent(pb); err != nil {
		t.Fatalf("explicit 0 max_run_seconds rejected: %v", err)
	}
}

// A pi agent is a leaf: declaring children is rejected with an error that
// says how to compose instead.
func TestValidatePiAgent_RejectsChildren(t *testing.T) {
	pb := validPiAgent()
	pb.ChildAgentIds = []string{"c1"}
	err := ValidatePiAgent(pb)
	if err == nil || !strings.Contains(err.Error(), "child_agent_ids") || !strings.Contains(err.Error(), "leaf") {
		t.Fatalf("expected leaf rejection naming child_agent_ids, got %v", err)
	}
}

// Every behavior-surface field owned by the box must be rejected with an
// error that names the field and points at box-side configuration.
func TestValidatePiAgent_RejectsBoxOwnedFields(t *testing.T) {
	cases := []struct {
		field  string
		mutate func(pb *agentsv1.Agent)
	}{
		{"instruction", func(pb *agentsv1.Agent) { pb.Config.Instruction = "do things" }},
		{"global_instruction", func(pb *agentsv1.Agent) { pb.Config.GlobalInstruction = "always" }},
		{"mcp_servers", func(pb *agentsv1.Agent) {
			pb.Config.McpServers = []*agentsv1.MCPServer{{Name: "srv"}}
		}},
		{"mcp_server_ids", func(pb *agentsv1.Agent) { pb.Config.McpServerIds = []string{"m1"} }},
		{"skills", func(pb *agentsv1.Agent) { pb.Config.Skills = []string{"sk"} }},
		{"file_mounts", func(pb *agentsv1.Agent) {
			pb.Config.FileMounts = []*agentsv1.AgentFileMount{{SpaceId: "sp"}}
		}},
		{"context_guard", func(pb *agentsv1.Agent) {
			pb.Config.ContextGuard = &agentsv1.ContextGuardConfig{
				Strategy: agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_THRESHOLD,
			}
		}},
		{"remote_agent_ids", func(pb *agentsv1.Agent) { pb.Config.RemoteAgentIds = []string{"r1"} }},
	}
	for _, tc := range cases {
		pb := validPiAgent()
		tc.mutate(pb)
		err := ValidatePiAgent(pb)
		if err == nil {
			t.Errorf("%s: expected rejection, got nil", tc.field)
			continue
		}
		if !strings.Contains(err.Error(), tc.field) {
			t.Errorf("%s: error does not name the field: %v", tc.field, err)
		}
		if !strings.Contains(err.Error(), "on the box") {
			t.Errorf("%s: error does not point at box-side configuration: %v", tc.field, err)
		}
	}
}

func TestNewFromProto_PiWithoutBuilderFails(t *testing.T) {
	_, err := NewFromProtoWithToolsetFactory(context.Background(), validPiAgent(), nil, nil, nil, nil, nil, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "pi") {
		t.Fatalf("expected pi-not-available error, got %v", err)
	}
}

func TestNewFromProto_PiUsesBuilder(t *testing.T) {
	var got *agentsv1.Agent
	builder := PiAgentBuilder(func(pb *agentsv1.Agent) (agent.Agent, error) {
		got = pb
		return agent.New(agent.Config{Name: pb.GetName(), Description: pb.GetDescription()})
	})
	a, err := NewFromProtoWithToolsetFactory(context.Background(), validPiAgent(), nil, nil, nil, nil, nil, nil, builder)
	if err != nil {
		t.Fatalf("NewFromProto: %v", err)
	}
	if a.Name() != "pi-coder" {
		t.Fatalf("agent name: got %q", a.Name())
	}
	if got.GetConfig().GetPi().GetButterboxId() != "box-1" {
		t.Fatalf("builder did not receive the pi config: %v", got)
	}
}

func TestNewFromProto_PiInvalidConfigFails(t *testing.T) {
	pb := validPiAgent()
	pb.Config.Instruction = "nope"
	builder := PiAgentBuilder(func(pb *agentsv1.Agent) (agent.Agent, error) {
		t.Fatal("builder must not be called for an invalid pi agent")
		return nil, nil
	})
	if _, err := NewFromProtoWithToolsetFactory(context.Background(), pb, nil, nil, nil, nil, nil, nil, builder); err == nil {
		t.Fatal("expected invalid pi config to fail construction")
	}
}
