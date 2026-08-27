package agent

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func cursorProto(butterboxID string) *agentsv1.Agent {
	return &agentsv1.Agent{
		Name:    "cursor-coder",
		AgentId: "cursor-coder",
		Type:    agentsv1.AgentType_AGENT_TYPE_CURSOR,
		Config: &agentsv1.AgentConfig{
			Cursor: &agentsv1.CursorAgentConfig{
				ButterboxId: butterboxID,
				WorkingDir:  "/project",
				Model:       "composer-2.5",
				Mode:        "agent",
			},
		},
	}
}

func TestValidateCursorAgent_Valid(t *testing.T) {
	if err := ValidateCursorAgent(cursorProto("box-1")); err != nil {
		t.Fatalf("expected valid, got: %v", err)
	}
}

func TestValidateCursorAgent_MissingConfig(t *testing.T) {
	pb := &agentsv1.Agent{
		Name: "c", Type: agentsv1.AgentType_AGENT_TYPE_CURSOR,
		Config: &agentsv1.AgentConfig{},
	}
	if err := ValidateCursorAgent(pb); err == nil || !strings.Contains(err.Error(), "config.cursor") {
		t.Fatalf("expected config.cursor error, got %v", err)
	}
}

func TestValidateCursorAgent_MissingButterboxID(t *testing.T) {
	if err := ValidateCursorAgent(cursorProto("")); err == nil || !strings.Contains(err.Error(), "butterbox_id") {
		t.Fatalf("expected butterbox_id error, got %v", err)
	}
}

func TestValidateCursorAgent_NegativeMaxRunSeconds(t *testing.T) {
	pb := cursorProto("box-1")
	pb.Config.Cursor.MaxRunSeconds = proto.Int32(-1)
	if err := ValidateCursorAgent(pb); err == nil || !strings.Contains(err.Error(), "max_run_seconds") {
		t.Fatalf("expected max_run_seconds error, got %v", err)
	}
}

func TestValidateCursorAgent_InvalidMode(t *testing.T) {
	pb := cursorProto("box-1")
	pb.Config.Cursor.Mode = "execute"
	if err := ValidateCursorAgent(pb); err == nil || !strings.Contains(err.Error(), "mode") {
		t.Fatalf("expected mode error, got %v", err)
	}
}

func TestValidateCursorAgent_ChildAgentsRejected(t *testing.T) {
	pb := cursorProto("box-1")
	pb.ChildAgentIds = []string{"child-1"}
	if err := ValidateCursorAgent(pb); err == nil || !strings.Contains(err.Error(), "child_agent_ids") {
		t.Fatalf("expected child_agent_ids error, got %v", err)
	}
}

func TestValidateCursorAgent_BoxOwnedFieldsRejected(t *testing.T) {
	tests := []struct {
		name  string
		setup func(pb *agentsv1.Agent)
		field string
	}{
		{"instruction", func(pb *agentsv1.Agent) { pb.Config.Instruction = "do something" }, "instruction"},
		{"global_instruction", func(pb *agentsv1.Agent) { pb.Config.GlobalInstruction = "global" }, "global_instruction"},
		{"mcp_servers", func(pb *agentsv1.Agent) { pb.Config.McpServers = []*agentsv1.MCPServer{{}} }, "mcp_servers"},
		{"mcp_server_ids", func(pb *agentsv1.Agent) { pb.Config.McpServerIds = []string{"x"} }, "mcp_server_ids"},
		{"skills", func(pb *agentsv1.Agent) { pb.Config.Skills = []string{"my-skill"} }, "skills"},
		{"file_mounts", func(pb *agentsv1.Agent) { pb.Config.FileMounts = []*agentsv1.AgentFileMount{{}} }, "file_mounts"},
		{"context_guard", func(pb *agentsv1.Agent) { pb.Config.ContextGuard = &agentsv1.ContextGuardConfig{} }, "context_guard"},
		{"remote_agent_ids", func(pb *agentsv1.Agent) { pb.Config.RemoteAgentIds = []string{"r"} }, "remote_agent_ids"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pb := cursorProto("box-1")
			tc.setup(pb)
			err := ValidateCursorAgent(pb)
			if err == nil || !strings.Contains(err.Error(), tc.field) {
				t.Fatalf("expected %s rejection, got %v", tc.field, err)
			}
			if !strings.Contains(err.Error(), ".cursor/rules") {
				t.Fatalf("expected box-config pointer, got %v", err)
			}
		})
	}
}

func TestValidateCursorAgent_NonCursorTypePasses(t *testing.T) {
	pb := &agentsv1.Agent{Name: "llm", Type: agentsv1.AgentType_AGENT_TYPE_LLM}
	if err := ValidateCursorAgent(pb); err != nil {
		t.Fatalf("non-cursor type should pass: %v", err)
	}
}

func TestNewCursorAgent_NilBuilderReturnsError(t *testing.T) {
	pb := cursorProto("box-1")
	_, err := newCursorAgent(pb, nil)
	if err == nil || !strings.Contains(err.Error(), "not available") {
		t.Fatalf("expected not-available error, got %v", err)
	}
}
