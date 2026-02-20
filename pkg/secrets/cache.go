package secrets

import (
	"sync"
	"time"
)

// DefaultCacheTTL is the default time-to-live for cached secrets.
const DefaultCacheTTL = 5 * time.Minute

// cacheEntry holds a cached secret value with expiration.
type cacheEntry struct {
	value     string
	expiresAt time.Time
}

// Cache provides in-memory caching of resolved secrets.
// This reduces repeated calls to secret backends during a single command execution.
//
// Thread-safe: all operations are protected by a mutex.
//
// Security note: Cached values are stored in memory only and are cleared
// when Clear() is called or when the Cache is garbage collected.
type Cache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
	ttl     time.Duration
	now     func() time.Time // injectable for testing
}

// CacheOption is a functional option for configuring Cache.
type CacheOption func(*Cache)

// WithTTL sets the cache TTL.
func WithTTL(ttl time.Duration) CacheOption {
	return func(c *Cache) {
		c.ttl = ttl
	}
}

// withNowFunc sets a custom time function (for testing).
func withNowFunc(fn func() time.Time) CacheOption {
	return func(c *Cache) {
		c.now = fn
	}
}

// NewCache creates a new secret cache with the given TTL.
// If no TTL options are provided, DefaultCacheTTL is used.
func NewCache(opts ...CacheOption) *Cache {
	c := &Cache{
		entries: make(map[string]cacheEntry),
		ttl:     DefaultCacheTTL,
		now:     time.Now,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// cacheKey generates a unique cache key for a secret request.
func cacheKey(provider, name, key string) string {
	if key != "" {
		return provider + ":" + name + ":" + key
	}
	return provider + ":" + name
}

// Get retrieves a cached secret value.
// Returns the value and true if found and not expired.
// Returns "", false if not found or expired.
func (c *Cache) Get(provider, name, key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	k := cacheKey(provider, name, key)
	entry, ok := c.entries[k]
	if !ok {
		return "", false
	}

	if c.now().After(entry.expiresAt) {
		// Expired - caller will need to re-fetch
		return "", false
	}

	return entry.value, true
}

// Set stores a secret value in the cache.
func (c *Cache) Set(provider, name, key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	k := cacheKey(provider, name, key)
	c.entries[k] = cacheEntry{
		value:     value,
		expiresAt: c.now().Add(c.ttl),
	}
}

// Delete removes a specific entry from the cache.
func (c *Cache) Delete(provider, name, key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	k := cacheKey(provider, name, key)
	delete(c.entries, k)
}

// Clear removes all entries from the cache.
// This should be called after a command completes to ensure secrets
// don't persist longer than necessary.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]cacheEntry)
}

// Prune removes expired entries from the cache.
// This is optional cleanup - expired entries are also handled lazily by Get.
func (c *Cache) Prune() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.now()
	for k, entry := range c.entries {
		if now.After(entry.expiresAt) {
			delete(c.entries, k)
		}
	}
}

// Size returns the number of entries in the cache (including expired).
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}
