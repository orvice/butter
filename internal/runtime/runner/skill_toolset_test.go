package runner

import (
	"context"
	"testing"

	skillmemory "go.orx.me/apps/butter/internal/repo/skill/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func skillAgentProto(workspaceID string, skills []string) *agentsv1.Agent {
	return &agentsv1.Agent{
		Name:        "helper",
		WorkspaceId: workspaceID,
		Config:      &agentsv1.AgentConfig{Skills: skills},
	}
}

func TestToolsetFactoryAttachesSkillToolset(t *testing.T) {
	ctx := context.Background()
	factory := newToolsetFactory(nil, 0, skillmemory.New())
	if factory == nil {
		t.Fatal("factory should exist when a skill repository is configured")
	}

	toolsets, err := factory(ctx, skillAgentProto("ws-1", []string{"alpha"}))
	if err != nil {
		t.Fatalf("factory: %v", err)
	}
	if len(toolsets) != 1 {
		t.Fatalf("toolsets = %d, want 1 skill toolset", len(toolsets))
	}
	tools, err := toolsets[0].Tools(nil)
	if err != nil {
		t.Fatalf("Tools: %v", err)
	}
	got := make(map[string]bool, len(tools))
	for _, tl := range tools {
		got[tl.Name()] = true
	}
	for _, name := range []string{"list_skills", "load_skill", "load_skill_resource"} {
		if !got[name] {
			t.Errorf("skill toolset is missing tool %q (has %v)", name, got)
		}
	}
}

func TestToolsetFactorySkipsSkillToolset(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name string
		pb   *agentsv1.Agent
	}{
		{"no workspace binding", skillAgentProto("", []string{"alpha"})},
		{"empty skills list", skillAgentProto("ws-1", nil)},
		{"no config block", &agentsv1.Agent{Name: "helper", WorkspaceId: "ws-1"}},
	}
	factory := newToolsetFactory(nil, 0, skillmemory.New())
	for _, tc := range cases {
		toolsets, err := factory(ctx, tc.pb)
		if err != nil {
			t.Fatalf("%s: factory: %v", tc.name, err)
		}
		if len(toolsets) != 0 {
			t.Errorf("%s: toolsets = %d, want none", tc.name, len(toolsets))
		}
	}

	// No skill repository at all (e.g. bare NewService): no factory-built
	// skill toolset, and with no agent-file repo either, no factory at all.
	if factory := newToolsetFactory(nil, 0, nil); factory != nil {
		t.Error("factory should be nil when no built-in toolset backend is configured")
	}
}
