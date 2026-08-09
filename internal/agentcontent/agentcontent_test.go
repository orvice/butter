package agentcontent

import (
	"testing"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func blobMap(m map[string]string) CacheBlobReader {
	return func(p string) ([]byte, bool) {
		v, ok := m[p]
		return []byte(v), ok
	}
}

func TestParse_FullAgent(t *testing.T) {
	blobs := blobMap(map[string]string{
		"agents/my-agent/description.md":   "Agent description",
		"agents/my-agent/prompt.md":        "You are a helpful assistant.",
		"agents/my-agent/global-prompt.md": "Global rules here.",
	})

	result := Parse(blobs, []string{"my-agent"})
	if len(result.Content) != 1 {
		t.Fatalf("got %d entries, want 1", len(result.Content))
	}
	c := result.Content["my-agent"]
	if c.Description != "Agent description" {
		t.Errorf("description = %q", c.Description)
	}
	if c.Instruction != "You are a helpful assistant." {
		t.Errorf("instruction = %q", c.Instruction)
	}
	if c.GlobalInstruction != "Global rules here." {
		t.Errorf("global_instruction = %q", c.GlobalInstruction)
	}
}

func TestParse_PromptOnly(t *testing.T) {
	blobs := blobMap(map[string]string{
		"agents/my-agent/prompt.md": "Just a prompt.",
	})

	result := Parse(blobs, []string{"my-agent"})
	c := result.Content["my-agent"]
	if c.Description != "" {
		t.Errorf("description should be empty, got %q", c.Description)
	}
	if c.Instruction != "Just a prompt." {
		t.Errorf("instruction = %q", c.Instruction)
	}
	if c.GlobalInstruction != "" {
		t.Errorf("global_instruction should be empty, got %q", c.GlobalInstruction)
	}
}

func TestParse_NoFilesSkipsAgent(t *testing.T) {
	blobs := blobMap(map[string]string{})

	result := Parse(blobs, []string{"no-files"})
	if len(result.Content) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(result.Content))
	}
}

func TestParse_WhitespaceTrimmed(t *testing.T) {
	blobs := blobMap(map[string]string{
		"agents/ws-agent/prompt.md": "  spaced out  \n",
	})

	result := Parse(blobs, []string{"ws-agent"})
	c := result.Content["ws-agent"]
	if c.Instruction != "spaced out" {
		t.Errorf("instruction = %q, want trimmed", c.Instruction)
	}
}

func TestParse_MultipleAgents(t *testing.T) {
	blobs := blobMap(map[string]string{
		"agents/alpha/prompt.md": "alpha prompt",
		"agents/beta/prompt.md":  "beta prompt",
	})

	result := Parse(blobs, []string{"alpha", "beta", "gamma"})
	if len(result.Content) != 2 {
		t.Fatalf("got %d entries, want 2", len(result.Content))
	}
	if result.Content["alpha"].Instruction != "alpha prompt" {
		t.Errorf("alpha instruction = %q", result.Content["alpha"].Instruction)
	}
	if result.Content["beta"].Instruction != "beta prompt" {
		t.Errorf("beta instruction = %q", result.Content["beta"].Instruction)
	}
	if _, ok := result.Content["gamma"]; ok {
		t.Error("gamma should not be in content (no files)")
	}
}

func TestValidate_LLMRequiresPrompt(t *testing.T) {
	content := map[string]AgentContent{
		"llm-agent": {AgentID: "llm-agent", Description: "has desc", Instruction: ""},
	}
	agents := []*agentsv1.Agent{
		{AgentId: "llm-agent", Type: agentsv1.AgentType_AGENT_TYPE_LLM},
	}

	errs := Validate(content, agents)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if errs[0].AgentID != "llm-agent" {
		t.Errorf("error agent_id = %q", errs[0].AgentID)
	}
}

func TestValidate_UnspecifiedTypeTreatedAsLLM(t *testing.T) {
	content := map[string]AgentContent{
		"unspec": {AgentID: "unspec", Instruction: ""},
	}
	agents := []*agentsv1.Agent{
		{AgentId: "unspec", Type: agentsv1.AgentType_AGENT_TYPE_UNSPECIFIED},
	}

	errs := Validate(content, agents)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error for unspecified type, got %d", len(errs))
	}
}

func TestValidate_NonLLMIgnoresEmptyPrompt(t *testing.T) {
	content := map[string]AgentContent{
		"seq-agent": {AgentID: "seq-agent", Instruction: ""},
	}
	agents := []*agentsv1.Agent{
		{AgentId: "seq-agent", Type: agentsv1.AgentType_AGENT_TYPE_SEQUENTIAL},
	}

	errs := Validate(content, agents)
	if len(errs) != 0 {
		t.Fatalf("expected 0 errors for non-LLM agent, got %d", len(errs))
	}
}

func TestValidate_DuplicateAgentIDUsesStrictestRequirement(t *testing.T) {
	content := map[string]AgentContent{
		"shared": {AgentID: "shared", Instruction: ""},
	}
	agents := []*agentsv1.Agent{
		{AgentId: "shared", Type: agentsv1.AgentType_AGENT_TYPE_LLM},
		{AgentId: "shared", Type: agentsv1.AgentType_AGENT_TYPE_SEQUENTIAL},
	}

	errs := Validate(content, agents)
	if len(errs) != 1 {
		t.Fatalf("expected shared LLM requirement to be retained, got %d errors", len(errs))
	}
}

func TestValidate_LLMWithPromptPasses(t *testing.T) {
	content := map[string]AgentContent{
		"good": {AgentID: "good", Instruction: "You are helpful."},
	}
	agents := []*agentsv1.Agent{
		{AgentId: "good", Type: agentsv1.AgentType_AGENT_TYPE_LLM},
	}

	errs := Validate(content, agents)
	if len(errs) != 0 {
		t.Fatalf("expected 0 errors, got %d", len(errs))
	}
}

func TestValidate_UnknownAgentIDIgnored(t *testing.T) {
	content := map[string]AgentContent{
		"unknown": {AgentID: "unknown", Instruction: ""},
	}
	errs := Validate(content, nil)
	if len(errs) != 0 {
		t.Fatalf("expected 0 errors for unknown agent, got %d", len(errs))
	}
}

func TestApplyToProto(t *testing.T) {
	agents := []agentsv1.Agent{
		{
			AgentId: "a1",
			Config:  &agentsv1.AgentConfig{Instruction: "old"},
		},
		{
			AgentId: "a2",
			Config:  &agentsv1.AgentConfig{Instruction: "keep"},
		},
		{
			Name: "no-id-agent",
		},
	}
	content := map[string]AgentContent{
		"a1": {Description: "new desc", Instruction: "new prompt", GlobalInstruction: "new global"},
	}

	ApplyToProto(agents, content)

	if agents[0].Description != "new desc" {
		t.Errorf("a1 description = %q", agents[0].Description)
	}
	if agents[0].Config.GetInstruction() != "new prompt" {
		t.Errorf("a1 instruction = %q", agents[0].Config.GetInstruction())
	}
	if agents[0].Config.GetGlobalInstruction() != "new global" {
		t.Errorf("a1 global_instruction = %q", agents[0].Config.GetGlobalInstruction())
	}
	if agents[1].Config.GetInstruction() != "keep" {
		t.Errorf("a2 instruction should be unchanged, got %q", agents[1].Config.GetInstruction())
	}
}

func TestApplyToProto_NilConfig(t *testing.T) {
	agents := []agentsv1.Agent{
		{AgentId: "x"},
	}
	content := map[string]AgentContent{
		"x": {Instruction: "hello"},
	}

	ApplyToProto(agents, content)

	if agents[0].Config == nil {
		t.Fatal("Config should have been created")
	}
	if agents[0].Config.GetInstruction() != "hello" {
		t.Errorf("instruction = %q", agents[0].Config.GetInstruction())
	}
}
