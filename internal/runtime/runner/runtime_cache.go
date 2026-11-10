package runner

import (
	"google.golang.org/adk/v2/agent"
	adkrunner "google.golang.org/adk/v2/runner"

	internalagent "go.orx.me/apps/butter/internal/agent"
)

type runtimeCacheStatus uint8

const (
	runtimeCacheStale runtimeCacheStatus = iota
	runtimeCacheMiss
	runtimeCacheHit
)

// runtimeCache owns all reload-sensitive cached values. Callers hold Service.mu
// while looking up, publishing, or resetting it; expensive builds stay outside
// that lock and publish only into the generation they started from.
type runtimeCache struct {
	generation       uint64
	runners          map[string]*adkrunner.Runner
	overriddenAgents map[string]agent.Agent
	// overriddenResources owns the MCP connections of every cached
	// model-overridden agent, by the same key.
	overriddenResources map[string]*internalagent.Resources
}

func newRuntimeCache() runtimeCache {
	return runtimeCache{
		generation:          1,
		runners:             make(map[string]*adkrunner.Runner),
		overriddenAgents:    make(map[string]agent.Agent),
		overriddenResources: make(map[string]*internalagent.Resources),
	}
}

// reset starts a new generation and returns the resources of the cached
// model-overridden agents it dropped, which the caller closes.
func (c *runtimeCache) reset() map[string]*internalagent.Resources {
	c.generation++
	c.runners = make(map[string]*adkrunner.Runner)
	c.overriddenAgents = make(map[string]agent.Agent)
	retired := c.overriddenResources
	c.overriddenResources = make(map[string]*internalagent.Resources)
	return retired
}

func (c *runtimeCache) current(expectedGeneration uint64) bool {
	return c.generation == expectedGeneration
}

func (c *runtimeCache) runner(expectedGeneration uint64, key string) (*adkrunner.Runner, runtimeCacheStatus) {
	return runtimeCacheLookup(c.generation, expectedGeneration, c.runners, key)
}

func (c *runtimeCache) publishRunner(expectedGeneration uint64, key string, value *adkrunner.Runner) (*adkrunner.Runner, runtimeCacheStatus) {
	return runtimeCachePublish(c.generation, expectedGeneration, c.runners, key, value)
}

func (c *runtimeCache) overriddenAgent(expectedGeneration uint64, key string) (agent.Agent, *internalagent.Resources, runtimeCacheStatus) {
	value, status := runtimeCacheLookup(c.generation, expectedGeneration, c.overriddenAgents, key)
	return value, c.overriddenResources[key], status
}

// publishOverriddenAgent stores built and its resources unless another build
// won the same key; on a hit the caller owns and closes the resources it built.
func (c *runtimeCache) publishOverriddenAgent(expectedGeneration uint64, key string, value agent.Agent, resources *internalagent.Resources) (agent.Agent, *internalagent.Resources, runtimeCacheStatus) {
	published, status := runtimeCachePublish(c.generation, expectedGeneration, c.overriddenAgents, key, value)
	if status == runtimeCacheMiss {
		c.overriddenResources[key] = resources
	}
	return published, c.overriddenResources[key], status
}

func runtimeCacheLookup[T any](currentGeneration, expectedGeneration uint64, values map[string]T, key string) (T, runtimeCacheStatus) {
	var zero T
	if currentGeneration != expectedGeneration {
		return zero, runtimeCacheStale
	}
	if value, ok := values[key]; ok {
		return value, runtimeCacheHit
	}
	return zero, runtimeCacheMiss
}

func runtimeCachePublish[T any](currentGeneration, expectedGeneration uint64, values map[string]T, key string, built T) (T, runtimeCacheStatus) {
	var zero T
	if currentGeneration != expectedGeneration {
		return zero, runtimeCacheStale
	}
	if value, ok := values[key]; ok {
		return value, runtimeCacheHit
	}
	values[key] = built
	return built, runtimeCacheMiss
}
