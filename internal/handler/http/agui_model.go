package http

import (
	"context"
	"errors"
	"net/http"
	"strings"

	internalagent "go.orx.me/apps/butter/internal/agent"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// modelOverrideProp is a per-run extension, never a write to agent config or
// shared state. Absent or empty means the agent's configured model.
func modelOverrideProp(forwardedProps any) (string, error) {
	props, _ := forwardedProps.(map[string]any)
	raw, exists := props["butterModel"]
	if !exists {
		return "", nil
	}
	model, ok := raw.(map[string]any)
	if !ok {
		return "", errors.New("forwardedProps.butterModel must be an object")
	}
	for key := range model {
		if key != "override" {
			return "", errors.New("unsupported forwardedProps.butterModel field")
		}
	}
	rawOverride, exists := model["override"]
	if !exists {
		return "", nil
	}
	override, ok := rawOverride.(string)
	if !ok {
		return "", errors.New("forwardedProps.butterModel.override must be a string")
	}
	return strings.TrimSpace(override), nil
}

func (h *AGUIHandler) prepareModelOverride(ctx context.Context, target aguiTarget, props any) (string, int, error) {
	override, err := modelOverrideProp(props)
	if err != nil {
		return "", http.StatusBadRequest, err
	}
	if override == "" {
		return "", 0, nil
	}
	if target.agent.GetType() != agentsv1.AgentType_AGENT_TYPE_LLM && target.agent.GetType() != agentsv1.AgentType_AGENT_TYPE_UNSPECIFIED {
		return "", http.StatusBadRequest, errors.New("model override is supported only for LLM agents")
	}
	if h.modelRepo == nil {
		return "", http.StatusServiceUnavailable, errors.New("model catalog unavailable")
	}
	providers, err := h.modelRepo.ListModelProviders(ctx, target.workspaceID)
	if err != nil {
		return "", http.StatusServiceUnavailable, errors.New("model catalog unavailable")
	}
	if _, found := internalagent.ResolveModelAliasPtr(override, providers); !found {
		return "", http.StatusBadRequest, errors.New("model override is not configured in this workspace")
	}
	return override, 0, nil
}
