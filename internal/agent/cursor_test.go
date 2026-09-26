package agent

import (
	"context"
	"strings"
	"testing"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/protobuf/proto"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func validCursorAgent() *agentsv1.Agent {
	return &agentsv1.Agent{
		Name:    "cursor-coder",
		AgentId: "cursor-coder",
		Type:    agentsv1.AgentType_AGENT_TYPE_CURSOR,
		Config: &agentsv1.AgentConfig{
			Cursor: &agentsv1.CursorAgentConfig{
				ButterboxId: "box-1",
				WorkingDir:  "projects/demo",
				Model:       "composer-2.5",
				Mode:        CursorModePlan,
			},
		},
	}
}

func TestValidateCursorAgent_Valid(t *testing.T) {
	if err := ValidateCursorAgent(validCursorAgent()); err != nil {
		t.Fatalf("valid cursor agent rejected: %v", err)
	}
	minimal := &agentsv1.Agent{
		Name:   "cursor-min",
		Type:   agentsv1.AgentType_AGENT_TYPE_CURSOR,
		Config: &agentsv1.AgentConfig{Cursor: &agentsv1.CursorAgentConfig{ButterboxId: "box-1"}},
	}
	if err := ValidateCursorAgent(minimal); err != nil {
		t.Fatalf("minimal cursor agent rejected: %v", err)
	}
	llm := &agentsv1.Agent{Name: "llm", Type: agentsv1.AgentType_AGENT_TYPE_LLM, Config: &agentsv1.AgentConfig{Instruction: "hi"}}
	if err := ValidateCursorAgent(llm); err != nil {
		t.Fatalf("non-cursor agent rejected: %v", err)
	}
	if err := ValidateCursorAgent(nil); err != nil {
		t.Fatalf("nil agent rejected: %v", err)
	}
}

func TestValidateCursorAgent_ConfigRules(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(pb *agentsv1.Agent)
		want   string
	}{
		{"missing config", func(pb *agentsv1.Agent) { pb.Config.Cursor = nil }, "config.cursor"},
		{"missing box", func(pb *agentsv1.Agent) { pb.Config.Cursor.ButterboxId = " " }, "butterbox_id"},
		{"bad mode", func(pb *agentsv1.Agent) { pb.Config.Cursor.Mode = "yolo" }, "mode"},
		{"negative max run", func(pb *agentsv1.Agent) { pb.Config.Cursor.MaxRunSeconds = proto.Int32(-1) }, "max_run_seconds"},
		{"children", func(pb *agentsv1.Agent) { pb.ChildAgentIds = []string{"c1"} }, "leaf"},
	}
	for _, tc := range cases {
		pb := validCursorAgent()
		tc.mutate(pb)
		if err := ValidateCursorAgent(pb); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: expected error mentioning %q, got %v", tc.name, tc.want, err)
		}
	}

	pb := validCursorAgent()
	pb.Config.Cursor.MaxRunSeconds = proto.Int32(0) // explicit 0 = unlimited
	if err := ValidateCursorAgent(pb); err != nil {
		t.Fatalf("explicit 0 max_run_seconds rejected: %v", err)
	}
}

func TestValidateCursorAgent_RejectsBoxOwnedFields(t *testing.T) {
	cases := []struct {
		field  string
		mutate func(pb *agentsv1.Agent)
	}{
		{"instruction", func(pb *agentsv1.Agent) { pb.Config.Instruction = "do things" }},
		{"global_instruction", func(pb *agentsv1.Agent) { pb.Config.GlobalInstruction = "always" }},
		{"mcp_servers", func(pb *agentsv1.Agent) { pb.Config.McpServers = []*agentsv1.MCPServer{{Name: "srv"}} }},
		{"mcp_server_ids", func(pb *agentsv1.Agent) { pb.Config.McpServerIds = []string{"m1"} }},
		{"skills", func(pb *agentsv1.Agent) { pb.Config.Skills = []string{"sk"} }},
		{"file_mounts", func(pb *agentsv1.Agent) { pb.Config.FileMounts = []*agentsv1.AgentFileMount{{SpaceId: "sp"}} }},
		{"context_guard", func(pb *agentsv1.Agent) {
			pb.Config.ContextGuard = &agentsv1.ContextGuardConfig{Strategy: agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_THRESHOLD}
		}},
		{"remote_agent_ids", func(pb *agentsv1.Agent) { pb.Config.RemoteAgentIds = []string{"r1"} }},
		{"memory", func(pb *agentsv1.Agent) { pb.Config.Memory = &agentsv1.MemoryConfig{Enabled: true} }},
	}
	for _, tc := range cases {
		pb := validCursorAgent()
		tc.mutate(pb)
		err := ValidateCursorAgent(pb)
		if err == nil || !strings.Contains(err.Error(), tc.field) || !strings.Contains(err.Error(), ".cursor/rules") {
			t.Errorf("%s: expected rejection naming the field and .cursor/rules, got %v", tc.field, err)
		}
	}
}

func TestNewFromProto_CursorWithoutBuilderFails(t *testing.T) {
	_, err := NewFromProtoWithToolsetFactory(context.Background(), validCursorAgent(), nil, nil, nil, nil, nil, nil, &BoxAgentBuilders{})
	if err == nil || !strings.Contains(err.Error(), "cursor agents are not available") {
		t.Fatalf("expected cursor-not-available error, got %v", err)
	}
}

func TestNewFromProto_CursorUsesBuilder(t *testing.T) {
	var got *agentsv1.Agent
	builder := CursorAgentBuilder(func(pb *agentsv1.Agent) (agent.Agent, error) {
		got = pb
		return agent.New(agent.Config{Name: pb.GetName()})
	})
	a, err := NewFromProtoWithToolsetFactory(context.Background(), validCursorAgent(), nil, nil, nil, nil, nil, nil, &BoxAgentBuilders{Cursor: builder})
	if err != nil {
		t.Fatalf("NewFromProto: %v", err)
	}
	if a.Name() != "cursor-coder" || got.GetConfig().GetCursor().GetButterboxId() != "box-1" {
		t.Fatalf("builder did not receive the cursor config: %v", got)
	}
}

func TestNewFromProto_CursorInvalidConfigFails(t *testing.T) {
	pb := validCursorAgent()
	pb.Config.Skills = []string{"nope"}
	builder := CursorAgentBuilder(func(*agentsv1.Agent) (agent.Agent, error) {
		t.Fatal("builder must not be called for an invalid cursor agent")
		return nil, nil
	})
	if _, err := NewFromProtoWithToolsetFactory(context.Background(), pb, nil, nil, nil, nil, nil, nil, &BoxAgentBuilders{Cursor: builder}); err == nil {
		t.Fatal("expected invalid cursor config to fail construction")
	}
}
