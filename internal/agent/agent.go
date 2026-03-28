package agent

import (
	"context"
	"fmt"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/agent/workflowagents/loopagent"
	"google.golang.org/adk/agent/workflowagents/parallelagent"
	"google.golang.org/adk/agent/workflowagents/sequentialagent"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// NewFromProto creates an ADK agent from an agentsv1.Agent proto config.
// providers is the list of model provider mappings used to resolve LLM backends.
func NewFromProto(ctx context.Context, pb *agentsv1.Agent, providers []agentsv1.ModelProvider) (agent.Agent, error) {
	if pb == nil {
		return nil, fmt.Errorf("agent config is nil")
	}

	// Recursively build sub-agents.
	subAgents := make([]agent.Agent, 0, len(pb.GetSubAgents()))
	for _, sub := range pb.GetSubAgents() {
		sa, err := NewFromProto(ctx, sub, providers)
		if err != nil {
			return nil, fmt.Errorf("building sub-agent %q: %w", sub.GetName(), err)
		}
		subAgents = append(subAgents, sa)
	}

	switch pb.GetType() {
	case agentsv1.AgentType_AGENT_TYPE_LLM, agentsv1.AgentType_AGENT_TYPE_UNSPECIFIED:
		return newLLMAgent(ctx, pb, subAgents, providers)
	case agentsv1.AgentType_AGENT_TYPE_LOOP:
		return newLoopAgent(pb, subAgents)
	case agentsv1.AgentType_AGENT_TYPE_SEQUENTIAL:
		return newSequentialAgent(pb, subAgents)
	case agentsv1.AgentType_AGENT_TYPE_PARALLEL:
		return newParallelAgent(pb, subAgents)
	default:
		return nil, fmt.Errorf("unsupported agent type: %v", pb.GetType())
	}
}

func newLLMAgent(ctx context.Context, pb *agentsv1.Agent, subAgents []agent.Agent, providers []agentsv1.ModelProvider) (agent.Agent, error) {
	llmCfg := pb.GetConfig().GetLlm()
	if llmCfg == nil {
		return nil, fmt.Errorf("agent %q: LLM agent requires llm config", pb.GetName())
	}

	m, err := resolveModel(ctx, llmCfg.GetModel(), providers)
	if err != nil {
		return nil, fmt.Errorf("agent %q: creating model %q: %w", pb.GetName(), llmCfg.GetModel(), err)
	}

	cfg := llmagent.Config{
		Name:                     pb.GetName(),
		Description:              pb.GetDescription(),
		SubAgents:                subAgents,
		Model:                    m,
		Instruction:              llmCfg.GetInstruction(),
		GlobalInstruction:        llmCfg.GetGlobalInstruction(),
		DisallowTransferToParent: llmCfg.GetDisallowTransferToParent(),
		DisallowTransferToPeers:  llmCfg.GetDisallowTransferToPeers(),
		OutputKey:                llmCfg.GetOutputKey(),
	}

	switch llmCfg.GetIncludeContents() {
	case agentsv1.LLMIncludeContents_LLM_INCLUDE_CONTENTS_NONE:
		cfg.IncludeContents = llmagent.IncludeContentsNone
	case agentsv1.LLMIncludeContents_LLM_INCLUDE_CONTENTS_DEFAULT:
		cfg.IncludeContents = llmagent.IncludeContentsDefault
	}

	return llmagent.New(cfg)
}

func newLoopAgent(pb *agentsv1.Agent, subAgents []agent.Agent) (agent.Agent, error) {
	loopCfg := pb.GetConfig().GetLoop()
	var maxIter uint
	if loopCfg != nil {
		maxIter = uint(loopCfg.GetMaxIterations())
	}

	return loopagent.New(loopagent.Config{
		AgentConfig: agent.Config{
			Name:        pb.GetName(),
			Description: pb.GetDescription(),
			SubAgents:   subAgents,
		},
		MaxIterations: maxIter,
	})
}

func newSequentialAgent(pb *agentsv1.Agent, subAgents []agent.Agent) (agent.Agent, error) {
	return sequentialagent.New(sequentialagent.Config{
		AgentConfig: agent.Config{
			Name:        pb.GetName(),
			Description: pb.GetDescription(),
			SubAgents:   subAgents,
		},
	})
}

func newParallelAgent(pb *agentsv1.Agent, subAgents []agent.Agent) (agent.Agent, error) {
	return parallelagent.New(parallelagent.Config{
		AgentConfig: agent.Config{
			Name:        pb.GetName(),
			Description: pb.GetDescription(),
			SubAgents:   subAgents,
		},
	})
}
