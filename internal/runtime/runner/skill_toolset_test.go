package runner

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/model"
	adkskill "google.golang.org/adk/v2/tool/skilltoolset/skill"
	"google.golang.org/adk/v2/tool/toolconfirmation"

	skillmemory "go.orx.me/apps/butter/internal/repo/skill/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// requestProcessor is the ADK toolset hook that injects the skill catalog into
// each LLM request (SkillToolset.ProcessRequest). It is unexported in ADK, so
// we redeclare the method set to type-assert against it.
type requestProcessor interface {
	ProcessRequest(ctx agent.Context, req *model.LLMRequest) error
}

func seedSkillMD(t *testing.T, repo *skillmemory.Store, workspaceID, name, description string) {
	t.Helper()
	md := "---\nname: " + name + "\ndescription: " + description + "\n---\nBody.\n"
	if _, err := repo.Create(context.Background(), workspaceID, &agentsv1.Skill{
		Name:        name,
		Description: description,
	}, md); err != nil {
		t.Fatalf("seed skill %q: %v", name, err)
	}
}

func skillAgentProto(workspaceID string, skills []string) *agentsv1.Agent {
	return &agentsv1.Agent{
		Name:        "helper",
		WorkspaceId: workspaceID,
		Config:      &agentsv1.AgentConfig{Skills: skills},
	}
}

func TestToolsetFactoryAttachesSkillToolset(t *testing.T) {
	ctx := context.Background()
	factory := newToolsetFactory(toolsetDeps{skillRepo: skillmemory.New()})
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

func TestToolsetFactoryInjectsSkillCatalog(t *testing.T) {
	ctx := context.Background()
	repo := skillmemory.New()
	seedSkillMD(t, repo, "ws-1", "alpha", "Summarise a document")
	seedSkillMD(t, repo, "ws-1", "hidden", "Not attached to this agent")

	factory := newToolsetFactory(toolsetDeps{skillRepo: repo})
	toolsets, err := factory(ctx, skillAgentProto("ws-1", []string{"alpha"}))
	if err != nil {
		t.Fatalf("factory: %v", err)
	}
	if len(toolsets) != 1 {
		t.Fatalf("toolsets = %d, want 1", len(toolsets))
	}
	rp, ok := toolsets[0].(requestProcessor)
	if !ok {
		t.Fatalf("skill toolset %T does not inject requests", toolsets[0])
	}

	mock := agent.NewStrictContextMock(ctx)
	req := &model.LLMRequest{}
	if err := rp.ProcessRequest(&mock, req); err != nil {
		t.Fatalf("ProcessRequest: %v", err)
	}

	instruction := systemInstructionText(req)
	if !strings.Contains(instruction, "alpha") || !strings.Contains(instruction, "Summarise a document") {
		t.Errorf("LLM request instruction missing allowlisted skill catalog entry; got:\n%s", instruction)
	}
	if strings.Contains(instruction, "hidden") {
		t.Errorf("LLM request instruction leaked a non-allowlisted skill; got:\n%s", instruction)
	}
}

// systemInstructionText concatenates the text parts of a request's system
// instruction, where ADK injects the skill catalog.
func systemInstructionText(req *model.LLMRequest) string {
	if req.Config == nil || req.Config.SystemInstruction == nil {
		return ""
	}
	var sb strings.Builder
	for _, p := range req.Config.SystemInstruction.Parts {
		sb.WriteString(p.Text)
	}
	return sb.String()
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
	factory := newToolsetFactory(toolsetDeps{skillRepo: skillmemory.New()})
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
	if factory := newToolsetFactory(toolsetDeps{}); factory != nil {
		t.Error("factory should be nil when no built-in toolset backend is configured")
	}
}

// runnableTool redeclares ADK's internal FunctionTool run surface (the
// method itself is exported) so the end-to-end test can invoke tools the way
// the ADK flow does.
type runnableTool interface {
	Run(ctx agent.Context, args any) (map[string]any, error)
}

// toolContextFake overrides the one context method the functiontool run
// path touches (ToolConfirmation); everything else stays strict.
type toolContextFake struct {
	agent.StrictContextMock
}

func (f *toolContextFake) ToolConfirmation() *toolconfirmation.ToolConfirmation { return nil }

func runTool(t *testing.T, ctx context.Context, tools map[string]runnableTool, name string, args map[string]any) map[string]any {
	t.Helper()
	tl, ok := tools[name]
	if !ok {
		t.Fatalf("tool %q not found", name)
	}
	result, err := tl.Run(&toolContextFake{agent.NewStrictContextMock(ctx)}, args)
	if err != nil {
		t.Fatalf("%s(%v): %v", name, args, err)
	}
	return result
}

// Issue #154 acceptance: an uploaded resource is discoverable and readable by
// the agent at runtime through list_skills → load_skill →
// load_skill_resource; a missing resource surfaces ADK's sentinel error.
func TestSkillResourceRuntimeEndToEnd(t *testing.T) {
	ctx := context.Background()
	repo := skillmemory.New()
	seedSkillMD(t, repo, "ws-1", "alpha", "Summarise a document")
	if _, err := repo.PutResource(ctx, "ws-1", "alpha", &agentsv1.SkillResource{
		Path:        "references/guide.md",
		ContentType: "text/markdown",
	}, []byte("Follow the guide.")); err != nil {
		t.Fatalf("PutResource: %v", err)
	}

	factory := newToolsetFactory(toolsetDeps{skillRepo: repo})
	toolsets, err := factory(ctx, skillAgentProto("ws-1", []string{"alpha"}))
	if err != nil {
		t.Fatalf("factory: %v", err)
	}
	if len(toolsets) != 1 {
		t.Fatalf("toolsets = %d, want 1", len(toolsets))
	}
	adkTools, err := toolsets[0].Tools(nil)
	if err != nil {
		t.Fatalf("Tools: %v", err)
	}
	tools := make(map[string]runnableTool, len(adkTools))
	for _, tl := range adkTools {
		rt, ok := tl.(runnableTool)
		if !ok {
			t.Fatalf("tool %q is not runnable", tl.Name())
		}
		tools[tl.Name()] = rt
	}

	listed := runTool(t, ctx, tools, "list_skills", map[string]any{})
	if s := fmt.Sprintf("%v", listed); !strings.Contains(s, "alpha") {
		t.Fatalf("list_skills does not expose the skill: %v", listed)
	}

	loaded := runTool(t, ctx, tools, "load_skill", map[string]any{"name": "alpha"})
	if s := fmt.Sprintf("%v", loaded); !strings.Contains(s, "Body.") {
		t.Fatalf("load_skill missing instructions: %v", loaded)
	}

	resource := runTool(t, ctx, tools, "load_skill_resource", map[string]any{
		"skill_name":    "alpha",
		"resource_path": "references/guide.md",
	})
	if resource["content"] != "Follow the guide." {
		t.Fatalf("load_skill_resource content = %v, want %q", resource["content"], "Follow the guide.")
	}

	// Missing resources yield ADK's sentinel through the tool error path.
	_, err = tools["load_skill_resource"].Run(&toolContextFake{agent.NewStrictContextMock(ctx)}, map[string]any{
		"skill_name":    "alpha",
		"resource_path": "references/absent.md",
	})
	if err == nil || !errors.Is(err, adkskill.ErrResourceNotFound) {
		t.Fatalf("expected ErrResourceNotFound for missing resource, got %v", err)
	}
}
