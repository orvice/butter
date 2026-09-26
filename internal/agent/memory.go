package agent

import (
	"fmt"
	"math"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// MaxMemoryTopK bounds MemoryConfig.top_k: recalled memories are injected
// into every model call of the turn, so the cap protects the context window.
const MaxMemoryTopK = 50

// ValidateMemoryConfig checks an agent's MemoryConfig bounds (ADR-0013).
// Type restrictions live elsewhere: PI and CURSOR agents reject the field
// through rejectBoxOwnedFields, and composite agents accept it because a
// composite root still drives recall and capture for its LLM sub-agents.
func ValidateMemoryConfig(pb *agentsv1.Agent) error {
	mc := pb.GetConfig().GetMemory()
	if mc == nil {
		return nil
	}
	if mc.TopK != nil && (mc.GetTopK() < 1 || mc.GetTopK() > MaxMemoryTopK) {
		return fmt.Errorf("agent %q: config.memory.top_k must be between 1 and %d (unset means 5)", pb.GetName(), MaxMemoryTopK)
	}
	if mc.Threshold != nil {
		th := float64(mc.GetThreshold())
		if math.IsNaN(th) || th < 0 || th > 1 {
			return fmt.Errorf("agent %q: config.memory.threshold must be between 0 and 1 (unset means 0.3)", pb.GetName())
		}
	}
	return nil
}
