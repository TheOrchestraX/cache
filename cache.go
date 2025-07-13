package cache

import (
	"log"
	"sync"
	"time"
)

// Cache is a generic container that holds items of type T, periodically
// reloading them via a loader function, and supporting on-demand reloads,
// individual additions/removals, and flexible searches.
// It swaps in the entire map atomically on each reload.

type Cache[T any] struct {
	loader   func() (map[string]T, error)
	interval time.Duration
	mu       sync.RWMutex
	data     map[string]T
	ticker   *time.Ticker
	quit     chan struct{}
}

// NewCache constructs a Cache for type T. interval defines how often
// AutoReload triggers. The initial data map is empty.
func NewCache[T any](loader func() (map[string]T, error), interval time.Duration) *Cache[T] {
	return &Cache[T]{
		loader:   loader,
		interval: interval,
		data:     make(map[string]T),
		quit:     make(chan struct{}),
	}
}

// Load invokes the loader function and, on success, swaps in the new map.
func (c *Cache[T]) Load() {
	result, err := c.loader()
	if err != nil {
		log.Println("Cache load error:", err)
		return
	}
	c.mu.Lock()
	c.data = result
	c.mu.Unlock()
	log.Printf("[%s] Cache reloaded (%d items)", time.Now().Format(time.RFC3339), len(result))
}

// Reload is an alias for Load, to explicitly reload on demand.
func (c *Cache[T]) Reload() {
	c.Load()
}

// StartAutoReload spins up a ticker to call Load() every interval.
func (c *Cache[T]) StartAutoReload() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.ticker != nil {
		return // already running
	}
	c.ticker = time.NewTicker(c.interval)
	go func() {
		for {
			select {
			case <-c.ticker.C:
				c.Load()
			case <-c.quit:
				c.ticker.Stop()
				return
			}
		}
	}()
}

// StopAutoReload stops the periodic reload and cleans up resources.
func (c *Cache[T]) StopAutoReload() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.ticker != nil {
		c.ticker.Stop()
		c.ticker = nil
	}
	close(c.quit)
}

// SetInterval updates the reload interval at runtime.
func (c *Cache[T]) SetInterval(interval time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.interval = interval
	if c.ticker != nil {
		c.ticker.Stop()
		c.ticker = time.NewTicker(c.interval)
	}
}

// Add inserts or updates a single item in the cache under the given key.
func (c *Cache[T]) Add(key string, value T) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

// Delete removes the item with the given key from the cache.
func (c *Cache[T]) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

// Clear empties the entire cache.
func (c *Cache[T]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]T)
}

// Get returns the item for a key, and a boolean indicating presence.
func (c *Cache[T]) Get(key string) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.data[key]
	return val, ok
}

// GetAll returns a shallow copy of the entire cached map.
func (c *Cache[T]) GetAll() map[string]T {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]T, len(c.data))
	for k, v := range c.data {
		result[k] = v
	}
	return result
}

// Find returns all items satisfying the provided predicate.
func (c *Cache[T]) Find(predicate func(T) bool) []T {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var results []T
	for _, v := range c.data {
		if predicate(v) {
			results = append(results, v)
		}
	}
	return results
}

// FindOne returns the first item satisfying predicate, or false if none.
func (c *Cache[T]) FindOne(predicate func(T) bool) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, v := range c.data {
		if predicate(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}
