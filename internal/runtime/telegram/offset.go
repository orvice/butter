package telegram

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/redis/go-redis/v9"

	"go.orx.me/apps/butter/internal/telegramqueue"
)

// offsetKeyPrefix namespaces per-Channel long-poll offsets.
const offsetKeyPrefix = "butter:telegram:offset:"

// OffsetKey addresses one Channel's committed offset.
func OffsetKey(channelID string) string { return offsetKeyPrefix + channelID }

// OffsetStore persists how far a Channel's long poll has been confirmed.
//
// The offset lives outside the process because leadership moves: when a
// leader dies, whichever Pod takes the lease has to resume from what was
// actually committed, not from zero (which would replay Telegram's whole
// backlog) and not from its own memory (which is empty).
type OffsetStore interface {
	Get(ctx context.Context, channelID string) (int64, error)
	// Commit advances the offset. `owner` fences the write: a Pod that lost
	// the lease must not be able to commit, or it would confirm updates the
	// new leader is still processing.
	Commit(ctx context.Context, channelID, owner string, offset int64) error
}

// ErrNotOffsetOwner means the caller no longer holds the Channel's lease and
// its commit was refused.
var ErrNotOffsetOwner = errors.New("offset owner changed")

// RedisOffsetStore persists offsets in Redis, fenced by the polling lease.
type RedisOffsetStore struct {
	rdb *redis.Client
}

func NewRedisOffsetStore(rdb *redis.Client) *RedisOffsetStore {
	if rdb == nil {
		return nil
	}
	return &RedisOffsetStore{rdb: rdb}
}

var _ OffsetStore = (*RedisOffsetStore)(nil)

// commitScript writes the offset only while the caller still holds the lease,
// and only ever forward.
//
// Both conditions matter. The ownership check fences a paused leader that
// wakes up after another Pod took over. The monotonic check stops an
// out-of-order commit from rewinding the offset and replaying updates that
// were already handled.
var commitScript = redis.NewScript(`
if redis.call('GET', KEYS[1]) ~= ARGV[1] then
  return -1
end
local current = tonumber(redis.call('GET', KEYS[2]) or '0')
local next_offset = tonumber(ARGV[2])
if next_offset > current then
  redis.call('SET', KEYS[2], ARGV[2])
  return next_offset
end
return current
`)

func (s *RedisOffsetStore) Get(ctx context.Context, channelID string) (int64, error) {
	raw, err := s.rdb.Get(ctx, OffsetKey(channelID)).Result()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("read telegram offset: %w", err)
	}
	offset, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		// A corrupt offset is treated as "start from whatever Telegram still
		// has": resetting to zero is safe because dedupe suppresses replays.
		return 0, nil
	}
	return offset, nil
}

func (s *RedisOffsetStore) Commit(ctx context.Context, channelID, owner string, offset int64) error {
	result, err := commitScript.Run(ctx, s.rdb,
		[]string{telegramqueue.PollingLeaseKey(channelID), OffsetKey(channelID)},
		owner, offset,
	).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("commit telegram offset: %w", err)
	}
	if applied, _ := result.(int64); applied == -1 {
		return ErrNotOffsetOwner
	}
	return nil
}

// MemoryOffsetStore is the single-process implementation used by tests.
type MemoryOffsetStore struct {
	mu      sync.Mutex
	offsets map[string]int64
	owners  map[string]string
}

func NewMemoryOffsetStore() *MemoryOffsetStore {
	return &MemoryOffsetStore{offsets: make(map[string]int64), owners: make(map[string]string)}
}

var _ OffsetStore = (*MemoryOffsetStore)(nil)

// SetOwner records who currently holds a Channel's lease, so tests can
// exercise the fencing behaviour the Redis store gets from the lease key.
func (s *MemoryOffsetStore) SetOwner(channelID, owner string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.owners[channelID] = owner
}

func (s *MemoryOffsetStore) Get(_ context.Context, channelID string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.offsets[channelID], nil
}

func (s *MemoryOffsetStore) Commit(_ context.Context, channelID, owner string, offset int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if current, ok := s.owners[channelID]; ok && current != owner {
		return ErrNotOffsetOwner
	}
	if offset > s.offsets[channelID] {
		s.offsets[channelID] = offset
	}
	return nil
}
