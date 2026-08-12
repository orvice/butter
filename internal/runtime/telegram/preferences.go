package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// preferenceKeyPrefix namespaces stored selections.
const preferenceKeyPrefix = "butter:telegram:prefs:"

// preferenceTTL bounds how long an unused selection survives. It is long
// enough that a preference outlives restarts and quiet weekends, which is the
// stated requirement, without keeping state for chats that stopped existing.
const preferenceTTL = 90 * 24 * time.Hour

// Preferences is a controller's stored Agent/Model selection for one session
// subject at one Destination.
type Preferences struct {
	AgentID string `json:"agent_id,omitempty"`
	Model   string `json:"model,omitempty"`
	// Debug carries the per-session debug toggle.
	Debug *bool `json:"debug,omitempty"`
}

// Empty reports whether nothing is selected.
func (p Preferences) Empty() bool { return p.AgentID == "" && p.Model == "" && p.Debug == nil }

// PreferenceStore persists selections outside process memory so they survive
// restarts and are visible to every Pod — which matters because any Pod may
// pick up the next message for the same Destination.
type PreferenceStore interface {
	Get(ctx context.Context, key string) (Preferences, error)
	Put(ctx context.Context, key string, prefs Preferences) error
	Delete(ctx context.Context, key string) error
	// DeletePrefix removes every selection under a Destination, used when a
	// Destination is disabled or deleted.
	DeletePrefix(ctx context.Context, prefix string) error
}

// PreferenceKey identifies one stored selection. Selections are per session
// subject, not per Destination, so a USER-policy topic gives each user their
// own choice.
func PreferenceKey(destinationID, subject string) string {
	return preferenceKeyPrefix + destinationID + ":" + subject
}

// DestinationPreferencePrefix matches every selection under a Destination.
func DestinationPreferencePrefix(destinationID string) string {
	return preferenceKeyPrefix + destinationID + ":"
}

// --- Redis ------------------------------------------------------------------

// RedisPreferenceStore persists selections in Redis.
type RedisPreferenceStore struct {
	rdb *redis.Client
}

func NewRedisPreferenceStore(rdb *redis.Client) *RedisPreferenceStore {
	if rdb == nil {
		return nil
	}
	return &RedisPreferenceStore{rdb: rdb}
}

var _ PreferenceStore = (*RedisPreferenceStore)(nil)

func (s *RedisPreferenceStore) Get(ctx context.Context, key string) (Preferences, error) {
	raw, err := s.rdb.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return Preferences{}, nil
	}
	if err != nil {
		return Preferences{}, fmt.Errorf("read telegram preferences: %w", err)
	}
	var prefs Preferences
	if err := json.Unmarshal([]byte(raw), &prefs); err != nil {
		// A corrupt record is treated as "no selection": falling back to the
		// Destination default keeps the topic working.
		return Preferences{}, nil
	}
	return prefs, nil
}

func (s *RedisPreferenceStore) Put(ctx context.Context, key string, prefs Preferences) error {
	if prefs.Empty() {
		return s.Delete(ctx, key)
	}
	raw, err := json.Marshal(prefs)
	if err != nil {
		return fmt.Errorf("encode telegram preferences: %w", err)
	}
	if err := s.rdb.Set(ctx, key, raw, preferenceTTL).Err(); err != nil {
		return fmt.Errorf("store telegram preferences: %w", err)
	}
	return nil
}

func (s *RedisPreferenceStore) Delete(ctx context.Context, key string) error {
	if err := s.rdb.Del(ctx, key).Err(); err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("clear telegram preferences: %w", err)
	}
	return nil
}

func (s *RedisPreferenceStore) DeletePrefix(ctx context.Context, prefix string) error {
	// SCAN rather than KEYS: this runs on a live Redis serving the receive
	// path, and KEYS would block it.
	var cursor uint64
	for {
		keys, next, err := s.rdb.Scan(ctx, cursor, prefix+"*", 100).Result()
		if err != nil {
			return fmt.Errorf("scan telegram preferences: %w", err)
		}
		if len(keys) > 0 {
			if err := s.rdb.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("clear telegram preferences: %w", err)
			}
		}
		if next == 0 {
			return nil
		}
		cursor = next
	}
}

// --- Memory -----------------------------------------------------------------

// MemoryPreferenceStore is the in-process implementation used by tests and
// by deployments without Redis. Selections then do not survive a restart,
// which is why Redis is the production path.
type MemoryPreferenceStore struct {
	mu    sync.RWMutex
	items map[string]Preferences
}

func NewMemoryPreferenceStore() *MemoryPreferenceStore {
	return &MemoryPreferenceStore{items: make(map[string]Preferences)}
}

var _ PreferenceStore = (*MemoryPreferenceStore)(nil)

func (s *MemoryPreferenceStore) Get(_ context.Context, key string) (Preferences, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.items[key], nil
}

func (s *MemoryPreferenceStore) Put(_ context.Context, key string, prefs Preferences) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if prefs.Empty() {
		delete(s.items, key)
		return nil
	}
	s.items[key] = prefs
	return nil
}

func (s *MemoryPreferenceStore) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, key)
	return nil
}

func (s *MemoryPreferenceStore) DeletePrefix(_ context.Context, prefix string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for key := range s.items {
		if strings.HasPrefix(key, prefix) {
			delete(s.items, key)
		}
	}
	return nil
}

// --- Resolution -------------------------------------------------------------

// Effective is the routing a turn actually uses, after applying a stored
// selection to the Destination's configuration.
type Effective struct {
	AgentID string
	Model   string
	// StaleAgent and StaleModel report that a stored choice is no longer
	// allowed and was dropped, so the caller can clear it.
	StaleAgent bool
	StaleModel bool
}

// ResolveEffective applies a stored selection to the Destination policy.
//
// A stored choice that current configuration no longer allows is ignored
// rather than honored: an operator who removes an Agent from the selectable
// list expects the topic to stop using it, not to keep using it until someone
// happens to switch again.
func ResolveEffective(config *agentsv1.TelegramDestinationConfig, stored Preferences) Effective {
	effective := Effective{AgentID: config.GetAgentId(), Model: config.GetModel()}

	if stored.AgentID != "" {
		switch {
		case stored.AgentID == config.GetAgentId():
			// Already the default; nothing to apply.
		case agentSelectable(config, stored.AgentID):
			effective.AgentID = stored.AgentID
		default:
			effective.StaleAgent = true
		}
	}
	if stored.Model != "" {
		switch {
		case stored.Model == config.GetModel():
		case modelSelectable(config, stored.Model):
			effective.Model = stored.Model
		default:
			effective.StaleModel = true
		}
	}
	return effective
}

// agentSelectable reports whether a controller may switch to this Agent. An
// empty selectable list locks selection to the default, which is what keeps
// topic routing fixed unless an operator opts in.
func agentSelectable(config *agentsv1.TelegramDestinationConfig, agentID string) bool {
	return len(config.GetSelectableAgentIds()) > 0 &&
		slices.Contains(config.GetSelectableAgentIds(), agentID)
}

func modelSelectable(config *agentsv1.TelegramDestinationConfig, model string) bool {
	return len(config.GetSelectableModels()) > 0 &&
		slices.Contains(config.GetSelectableModels(), model)
}

// AgentSelectionEnabled reports whether `/agent` can do anything here.
func AgentSelectionEnabled(config *agentsv1.TelegramDestinationConfig) bool {
	return len(config.GetSelectableAgentIds()) > 0
}

// ModelSelectionEnabled reports whether `/model` can do anything here.
func ModelSelectionEnabled(config *agentsv1.TelegramDestinationConfig) bool {
	return len(config.GetSelectableModels()) > 0
}
