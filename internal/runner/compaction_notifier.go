package runner

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/plugin"
	adkrunner "google.golang.org/adk/runner"
)

const stateKeyPrefixContentsAtCompaction = "__context_guard_contents_at_compaction_"

// CompactionCallback is called when context compaction is detected for an agent.
type CompactionCallback func(agentName string)

type compactionCallbackKey struct{}

// WithCompactionCallback stores a CompactionCallback in the context so the
// compaction notifier plugin can retrieve it during model callbacks.
func WithCompactionCallback(ctx context.Context, cb CompactionCallback) context.Context {
	return context.WithValue(ctx, compactionCallbackKey{}, cb)
}

// newCompactionNotifierPlugin creates an ADK plugin that detects when the
// contextguard plugin performs compaction. It must be registered AFTER the
// contextguard plugin so its BeforeModelCallback runs after contextguard's.
func newCompactionNotifierPlugin() adkrunner.PluginConfig {
	n := &compactionNotifier{
		lastSeen: make(map[string]int),
	}

	p, _ := plugin.New(plugin.Config{
		Name:                "compaction_notifier",
		BeforeModelCallback: llmagent.BeforeModelCallback(n.beforeModel),
	})

	return adkrunner.PluginConfig{
		Plugins: []*plugin.Plugin{p},
	}
}

type compactionNotifier struct {
	mu       sync.Mutex
	lastSeen map[string]int // "sessionID:agentName" -> last known compaction count
}

func (n *compactionNotifier) beforeModel(ctx agent.CallbackContext, _ *model.LLMRequest) (*model.LLMResponse, error) {
	agentName := ctx.AgentName()
	key := stateKeyPrefixContentsAtCompaction + agentName

	val, err := ctx.State().Get(key)
	if err != nil || val == nil {
		return nil, nil
	}

	var current int
	switch v := val.(type) {
	case int:
		current = v
	case float64:
		current = int(v)
	default:
		return nil, nil
	}

	if current == 0 {
		return nil, nil
	}

	sessionKey := fmt.Sprintf("%s:%s", ctx.SessionID(), agentName)

	n.mu.Lock()
	prev, seen := n.lastSeen[sessionKey]
	changed := current != prev
	n.lastSeen[sessionKey] = current
	n.mu.Unlock()

	if changed && seen {
		if cb, ok := ctx.Value(compactionCallbackKey{}).(CompactionCallback); ok && cb != nil {
			cb(agentName)
		}
	}

	return nil, nil
}
