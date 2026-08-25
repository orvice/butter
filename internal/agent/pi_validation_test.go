package agent

import (
	"strings"
	"testing"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func validPiAgent() *agentsv1.Agent {
	return &agentsv1.Agent{
		Type: agentsv1.AgentType_AGENT_TYPE_PI,
		Config: &agentsv1.AgentConfig{
			Pi: &agentsv1.PiAgentConfig{
				ButterboxId: "box-1",
			},
		},
	}
}

func TestValidatePiAgent_NonPiSkipped(t *testing.T) {
	a := &agentsv1.Agent{
		Type: agentsv1.AgentType_AGENT_TYPE_LLM,
		Config: &agentsv1.AgentConfig{
			Instruction:   "do stuff",
			McpServers:    []*agentsv1.MCPServer{{Name: "mcp"}},
			McpServerIds:  []string{"id1"},
			Skills:        []string{"skill-a"},
			FileMounts:    []*agentsv1.AgentFileMount{{SpaceId: "space-1"}},
			RemoteAgentIds: []string{"remote-1"},
			ContextGuard: &agentsv1.ContextGuardConfig{
				Strategy: agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_THRESHOLD,
			},
		},
		ChildAgentIds: []string{"child"},
	}
	if err := ValidatePiAgent(a); err != nil {
		t.Fatalf("expected nil for non-PI agent, got %v", err)
	}
}

func TestValidatePiAgent_MissingConfig(t *testing.T) {
	a := &agentsv1.Agent{
		Type:   agentsv1.AgentType_AGENT_TYPE_PI,
		Config: &agentsv1.AgentConfig{},
	}
	err := ValidatePiAgent(a)
	if err == nil {
		t.Fatal("expected error for missing config.pi")
	}
	if !strings.Contains(err.Error(), "config.pi") {
		t.Fatalf("expected error about config.pi, got %v", err)
	}
}

func TestValidatePiAgent_EmptyButterBoxID(t *testing.T) {
	tests := []struct {
		name         string
		butterboxID  string
	}{
		{"empty", ""},
		{"whitespace", "   "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := validPiAgent()
			a.Config.Pi.ButterboxId = tt.butterboxID
			err := ValidatePiAgent(a)
			if err == nil {
				t.Fatal("expected error for empty butterbox_id")
			}
			if !strings.Contains(err.Error(), "butterbox_id") {
				t.Fatalf("expected error about butterbox_id, got %v", err)
			}
		})
	}
}

func TestValidatePiAgent_ChildAgentsRejected(t *testing.T) {
	a := validPiAgent()
	a.ChildAgentIds = []string{"child-1"}
	err := ValidatePiAgent(a)
	if err == nil {
		t.Fatal("expected error for child_agent_ids")
	}
	if !strings.Contains(err.Error(), "child_agent_ids") {
		t.Fatalf("expected error about child_agent_ids, got %v", err)
	}
	if !strings.Contains(err.Error(), "leaves") {
		t.Fatalf("expected error about leaves, got %v", err)
	}
}

func TestValidatePiAgent_InstructionRejected(t *testing.T) {
	a := validPiAgent()
	a.Config.Instruction = "be helpful"
	err := ValidatePiAgent(a)
	if err == nil {
		t.Fatal("expected error for instruction")
	}
	if !strings.Contains(err.Error(), "instruction") {
		t.Fatalf("expected error listing instruction, got %v", err)
	}
}

func TestValidatePiAgent_McpServersRejected(t *testing.T) {
	a := validPiAgent()
	a.Config.McpServers = []*agentsv1.MCPServer{{Name: "mcp"}}
	err := ValidatePiAgent(a)
	if err == nil {
		t.Fatal("expected error for mcp_servers")
	}
	if !strings.Contains(err.Error(), "mcp_servers") {
		t.Fatalf("expected error listing mcp_servers, got %v", err)
	}
}

func TestValidatePiAgent_McpServerIdsRejected(t *testing.T) {
	a := validPiAgent()
	a.Config.McpServerIds = []string{"mcp-id-1"}
	err := ValidatePiAgent(a)
	if err == nil {
		t.Fatal("expected error for mcp_server_ids")
	}
	if !strings.Contains(err.Error(), "mcp_server_ids") {
		t.Fatalf("expected error listing mcp_server_ids, got %v", err)
	}
}

func TestValidatePiAgent_SkillsRejected(t *testing.T) {
	a := validPiAgent()
	a.Config.Skills = []string{"skill-a"}
	err := ValidatePiAgent(a)
	if err == nil {
		t.Fatal("expected error for skills")
	}
	if !strings.Contains(err.Error(), "skills") {
		t.Fatalf("expected error listing skills, got %v", err)
	}
}

func TestValidatePiAgent_FileMountsRejected(t *testing.T) {
	a := validPiAgent()
	a.Config.FileMounts = []*agentsv1.AgentFileMount{{SpaceId: "space-1"}}
	err := ValidatePiAgent(a)
	if err == nil {
		t.Fatal("expected error for file_mounts")
	}
	if !strings.Contains(err.Error(), "file_mounts") {
		t.Fatalf("expected error listing file_mounts, got %v", err)
	}
}

func TestValidatePiAgent_ContextGuardRejected(t *testing.T) {
	a := validPiAgent()
	a.Config.ContextGuard = &agentsv1.ContextGuardConfig{
		Strategy: agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_THRESHOLD,
	}
	err := ValidatePiAgent(a)
	if err == nil {
		t.Fatal("expected error for context_guard")
	}
	if !strings.Contains(err.Error(), "context_guard") {
		t.Fatalf("expected error listing context_guard, got %v", err)
	}
}

func TestValidatePiAgent_RemoteAgentIdsRejected(t *testing.T) {
	a := validPiAgent()
	a.Config.RemoteAgentIds = []string{"remote-1"}
	err := ValidatePiAgent(a)
	if err == nil {
		t.Fatal("expected error for remote_agent_ids")
	}
	if !strings.Contains(err.Error(), "remote_agent_ids") {
		t.Fatalf("expected error listing remote_agent_ids, got %v", err)
	}
}

func TestValidatePiAgent_MultipleRejected(t *testing.T) {
	a := validPiAgent()
	a.Config.Instruction = "be helpful"
	a.Config.McpServers = []*agentsv1.MCPServer{{Name: "mcp"}}
	a.Config.Skills = []string{"skill-a"}
	a.Config.RemoteAgentIds = []string{"remote-1"}

	err := ValidatePiAgent(a)
	if err == nil {
		t.Fatal("expected error for multiple disallowed fields")
	}
	msg := err.Error()
	for _, field := range []string{"instruction", "mcp_servers", "skills", "remote_agent_ids"} {
		if !strings.Contains(msg, field) {
			t.Fatalf("expected error to list %q, got %v", field, err)
		}
	}
}

func TestValidatePiAgent_ValidMinimal(t *testing.T) {
	a := validPiAgent()
	if err := ValidatePiAgent(a); err != nil {
		t.Fatalf("expected nil for valid minimal PI agent, got %v", err)
	}
}

func TestValidatePiAgent_ValidFull(t *testing.T) {
	a := validPiAgent()
	a.Config.Pi = &agentsv1.PiAgentConfig{
		ButterboxId:    "box-full",
		WorkingDir:     "/workspace",
		Provider:       "anthropic",
		Model:          "claude-sonnet",
		ThinkingLevel:  "high",
		MaxRunSeconds:  3600,
	}
	if err := ValidatePiAgent(a); err != nil {
		t.Fatalf("expected nil for valid full PI agent, got %v", err)
	}
}
