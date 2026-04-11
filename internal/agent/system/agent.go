package system

import (
	"context"
	"fmt"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/tool"

	internalagent "go.orx.me/apps/butter/internal/agent"
	"go.orx.me/apps/butter/internal/cron"
	"go.orx.me/apps/butter/internal/repo/configstore"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	// AgentName is the reserved name for the built-in system agent.
	AgentName = "system"

	systemInstruction = `You are the Butter system agent, a built-in administrative assistant.

Your capabilities:
- **Agent Management**: List all registered agents and view their detailed configurations.
- **Cron Job Management**: List, create, update, and delete scheduled cron jobs. View execution history.

When users ask about agents, use the list_agents or get_agent tools.
When users ask about scheduled tasks or cron jobs, use the appropriate cron tools.

Be concise and helpful. Format results clearly for the user.`
)

// NewAgent creates the built-in system agent with administrative tools.
// It requires a config store (for agent queries), a cron scheduler, and an execution repo.
// model is the LLM model name to use; providers are used to resolve it.
func NewAgent(ctx context.Context, store *configstore.Store, scheduler *cron.Scheduler, execRepo cron.ExecutionRepo, model string, providers []agentsv1.ModelProvider) (agent.Agent, error) {
	tools, err := buildTools(store, scheduler, execRepo)
	if err != nil {
		return nil, fmt.Errorf("building system agent tools: %w", err)
	}

	m, err := internalagent.ResolveModel(ctx, model, providers)
	if err != nil {
		return nil, fmt.Errorf("resolving model for system agent: %w", err)
	}

	return llmagent.New(llmagent.Config{
		Name:        AgentName,
		Description: "Built-in system agent for managing cron jobs and querying agents.",
		Model:       m,
		Instruction: systemInstruction,
		Tools:       tools,
	})
}

func buildTools(store *configstore.Store, scheduler *cron.Scheduler, execRepo cron.ExecutionRepo) ([]tool.Tool, error) {
	builders := []func() (tool.Tool, error){
		func() (tool.Tool, error) { return newListAgentsTool(store) },
		func() (tool.Tool, error) { return newGetAgentTool(store) },
		func() (tool.Tool, error) { return newListCronJobsTool(scheduler) },
		func() (tool.Tool, error) { return newCreateCronJobTool(scheduler) },
		func() (tool.Tool, error) { return newUpdateCronJobTool(scheduler) },
		func() (tool.Tool, error) { return newDeleteCronJobTool(scheduler) },
		func() (tool.Tool, error) { return newListCronExecutionsTool(execRepo) },
	}

	tools := make([]tool.Tool, 0, len(builders))
	for _, build := range builders {
		t, err := build()
		if err != nil {
			return nil, err
		}
		tools = append(tools, t)
	}
	return tools, nil
}
