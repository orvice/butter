package github

import "sync"

// ETagCache stores response bodies only for the exact request/authentication
// key supplied by the provider. Implementations must return independent byte
// slices so callers cannot mutate cached data while another request decodes it.
type ETagCache interface {
	Get(key string) (etag string, body []byte, ok bool)
	Put(key, etag string, body []byte)
}

type etagEntry struct {
	etag string
	body []byte
}

// MemoryETagCache is a process-local cache for tests and single-process
// deployments. It has no TTL or cross-Pod coordination and must not be treated
// as the production cache implementation.
type MemoryETagCache struct {
	mu      sync.RWMutex
	entries map[string]etagEntry
}

func NewMemoryETagCache() *MemoryETagCache {
	return &MemoryETagCache{entries: map[string]etagEntry{}}
}

func (c *MemoryETagCache) Get(key string) (string, []byte, bool) {
	if c == nil {
		return "", nil, false
	}
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok {
		return "", nil, false
	}
	return entry.etag, append([]byte(nil), entry.body...), true
}

func (c *MemoryETagCache) Put(key, etag string, body []byte) {
	if c == nil || key == "" || etag == "" || len(body) == 0 {
		return
	}
	c.mu.Lock()
	c.entries[key] = etagEntry{etag: etag, body: append([]byte(nil), body...)}
	c.mu.Unlock()
}
