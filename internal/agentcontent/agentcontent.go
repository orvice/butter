// Package agentcontent parses, validates, and manages Agent Content derived
// from the workspace repository cache (issue #216). The fixed file mapping:
//
//	agents/{agent-id}/description.md   → Agent.description
//	agents/{agent-id}/prompt.md        → AgentConfig.instruction
//	agents/{agent-id}/global-prompt.md → AgentConfig.global_instruction
//
// Missing or empty optional files clear their mapped values. LLM Agents
// require a non-empty prompt.md; other agent types ignore it.
package agentcontent

import (
	"fmt"
	"path"
	"strings"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	agentsDir        = "agents"
	descriptionFile  = "description.md"
	promptFile       = "prompt.md"
	globalPromptFile = "global-prompt.md"
)

// DescriptionPath returns the repo-relative (root_path-stripped) path of an
// agent's description.md file.
func DescriptionPath(agentID string) string { return path.Join(agentsDir, agentID, descriptionFile) }

// PromptPath returns the repo-relative path of an agent's prompt.md file.
func PromptPath(agentID string) string { return path.Join(agentsDir, agentID, promptFile) }

// GlobalPromptPath returns the repo-relative path of an agent's
// global-prompt.md file.
func GlobalPromptPath(agentID string) string {
	return path.Join(agentsDir, agentID, globalPromptFile)
}

// ManagedPaths returns every reserved managed file path for an agent, in a
// stable order (description, prompt, global-prompt).
func ManagedPaths(agentID string) []string {
	return []string{DescriptionPath(agentID), PromptPath(agentID), GlobalPromptPath(agentID)}
}

// AgentContent holds the parsed content for a single agent.
type AgentContent struct {
	AgentID           string
	Description       string
	Instruction       string
	GlobalInstruction string
}

// Snapshot stores validated agent content for one published revision.
type Snapshot struct {
	CommitSHA string
	Entries   map[string]AgentContent // keyed by agent_id
}

// ValidationError reports a content validation failure for one agent.
type ValidationError struct {
	AgentID string
	Path    string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("agent %q (%s): %s", e.AgentID, e.Path, e.Message)
	}
	return fmt.Sprintf("agent %q: %s", e.AgentID, e.Message)
}

// ParseResult holds the combined output of parsing the repo cache.
type ParseResult struct {
	Content map[string]AgentContent // keyed by agent_id
	Errors  []ValidationError
}

// CacheBlobReader returns the cached content for a cache-relative path.
// Returns empty data and false when the blob is not cached.
type CacheBlobReader func(path string) ([]byte, bool)

// Parse extracts agent content from cached blobs for every agent whose
// agent_id appears in knownAgentIDs. Agents without a directory in the cache
// are skipped silently (they keep their DB-only content).
func Parse(blobs CacheBlobReader, knownAgentIDs []string) *ParseResult {
	result := &ParseResult{
		Content: make(map[string]AgentContent, len(knownAgentIDs)),
	}

	for _, agentID := range knownAgentIDs {
		dirPath := path.Join(agentsDir, agentID)

		descData, _ := blobs(path.Join(dirPath, descriptionFile))
		promptData, _ := blobs(path.Join(dirPath, promptFile))
		globalData, _ := blobs(path.Join(dirPath, globalPromptFile))

		hasAnyFile := len(descData) > 0 || len(promptData) > 0 || len(globalData) > 0

		if !hasAnyFile {
			continue
		}

		result.Content[agentID] = AgentContent{
			AgentID:           agentID,
			Description:       strings.TrimSpace(string(descData)),
			Instruction:       strings.TrimSpace(string(promptData)),
			GlobalInstruction: strings.TrimSpace(string(globalData)),
		}
	}
	return result
}

// Validate checks parsed content against the agent proto definitions.
// Returns validation errors for invalid content. An empty error slice means
// the revision is publishable.
func Validate(parsed map[string]AgentContent, agents []*agentsv1.Agent) []ValidationError {
	requiresPrompt := make(map[string]bool, len(agents))
	for _, a := range agents {
		if id := a.GetAgentId(); id != "" {
			isLLM := a.GetType() == agentsv1.AgentType_AGENT_TYPE_LLM ||
				a.GetType() == agentsv1.AgentType_AGENT_TYPE_UNSPECIFIED
			requiresPrompt[id] = requiresPrompt[id] || isLLM
		}
	}

	var errs []ValidationError

	for agentID, content := range parsed {
		isLLM, ok := requiresPrompt[agentID]
		if !ok {
			continue
		}

		if isLLM && content.Instruction == "" {
			errs = append(errs, ValidationError{
				AgentID: agentID,
				Path:    path.Join(agentsDir, agentID, promptFile),
				Message: "LLM agent requires a non-empty prompt.md",
			})
		}
	}
	return errs
}

// ApplyToProto overlays agent content onto a slice of agent protos in place.
// Only agents with entries in the content map are modified. This produces
// the Effective Agent configuration used by the runner.
func ApplyToProto(agents []agentsv1.Agent, content map[string]AgentContent) {
	for i := range agents {
		agentID := agents[i].GetAgentId()
		if agentID == "" {
			continue
		}
		c, ok := content[agentID]
		if !ok {
			continue
		}
		agents[i].Description = c.Description
		if agents[i].Config == nil {
			agents[i].Config = &agentsv1.AgentConfig{}
		}
		agents[i].Config.Instruction = c.Instruction
		agents[i].Config.GlobalInstruction = c.GlobalInstruction
	}
}
