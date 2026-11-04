package github

import (
	"crypto/sha256"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode"
)

const cacheEntries = 128
const cacheBytes = 32 << 20
const cacheEntryBytes = 2 << 20

type cachedResponse struct {
	etag    string
	body    []byte
	header  http.Header
	touched time.Time
}
type responseCache struct {
	mu      sync.Mutex
	entries map[[32]byte]cachedResponse
	bytes   int
}

// A credential fingerprint, method, complete URL (including every query
// argument), media representation and API version define one cache namespace.
// Neither a raw token nor an Authorization header is retained by the cache.
func requestKey(token string, req *http.Request) [32]byte {
	credential := sha256.Sum256([]byte(token))
	h := sha256.New()
	h.Write(credential[:])
	for _, part := range []string{req.Method, req.URL.String(), req.Header.Get("Accept"), req.Header.Get("X-GitHub-Api-Version")} {
		h.Write([]byte{0})
		h.Write([]byte(part))
	}
	var key [32]byte
	copy(key[:], h.Sum(nil))
	return key
}
func (c *responseCache) get(key [32]byte) (cachedResponse, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.entries[key]
	if !ok {
		return cachedResponse{}, false
	}
	if time.Since(entry.touched) > 30*time.Minute {
		delete(c.entries, key)
		c.bytes -= len(entry.body)
		return cachedResponse{}, false
	}
	entry.touched = time.Now()
	c.entries[key] = entry
	return entry, true
}
func (c *responseCache) remove(key [32]byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if old, ok := c.entries[key]; ok {
		c.bytes -= len(old.body)
		delete(c.entries, key)
	}
}
func (c *responseCache) put(key [32]byte, etag string, body []byte, headers http.Header) {
	if etag == "" || len(etag) > 512 || strings.ContainsFunc(etag, unicode.IsControl) || len(body) > cacheEntryBytes {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.entries == nil {
		c.entries = map[[32]byte]cachedResponse{}
	}
	if old, ok := c.entries[key]; ok {
		c.bytes -= len(old.body)
		delete(c.entries, key)
	}
	for len(c.entries) >= cacheEntries || c.bytes+len(body) > cacheBytes {
		var oldest [32]byte
		var at time.Time
		for key, entry := range c.entries {
			if at.IsZero() || entry.touched.Before(at) {
				oldest, at = key, entry.touched
			}
		}
		c.bytes -= len(c.entries[oldest].body)
		delete(c.entries, oldest)
	}
	// Only representation metadata is needed. Never retain cookies or an
	// upstream header that might accidentally contain sensitive values.
	metadata := http.Header{}
	for _, name := range []string{"ETag", "Link", "Content-Type"} {
		if value := headers.Get(name); value != "" {
			metadata.Set(name, value)
		}
	}
	c.entries[key] = cachedResponse{etag: etag, body: append([]byte(nil), body...), header: metadata, touched: time.Now()}
	c.bytes += len(body)
}
