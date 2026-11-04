package handlers

import (
	"sync"
	"time"
)

// ttlCache memoizes a loader's result per key for a fixed time. It exists for
// whole-table aggregates (tag totals, popular broadcasters, trending tags)
// that every page load repeats but that change slowly. A failed load is not
// cached, so an outage does not pin an error for the whole window.
type ttlCache[K comparable, V any] struct {
	ttl     time.Duration
	now     func() time.Time
	mu      sync.Mutex
	entries map[K]ttlEntry[V]
}

type ttlEntry[V any] struct {
	value   V
	expires time.Time
}

func newTTLCache[K comparable, V any](ttl time.Duration) *ttlCache[K, V] {
	return &ttlCache[K, V]{ttl: ttl, now: time.Now, entries: map[K]ttlEntry[V]{}}
}

// get returns the cached value for key, or runs load and caches its result.
// Concurrent callers for a cold key each run load; the last writer wins.
func (c *ttlCache[K, V]) get(key K, load func() (V, error)) (V, error) {
	c.mu.Lock()
	entry, ok := c.entries[key]
	now := c.now()
	c.mu.Unlock()
	if ok && now.Before(entry.expires) {
		return entry.value, nil
	}
	value, err := load()
	if err != nil {
		return value, err
	}
	c.mu.Lock()
	c.entries[key] = ttlEntry[V]{value: value, expires: now.Add(c.ttl)}
	c.mu.Unlock()
	return value, nil
}
