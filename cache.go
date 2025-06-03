package cache

import (
	"log"
	"sync"
	"time"
)

// Cache periodically reloads a map[string]interface{} via a loader function.
// It swaps in the entire map atomically on each reload.

type Cache struct {
	loader   func() (map[string]interface{}, error)
	interval time.Duration
	data     map[string]interface{}
	mu       sync.RWMutex
	ticker   *time.Ticker
	quit     chan struct{}
}

// NewCache constructs a Cache. interval is how often Load() runs.
func NewCache(loader func() (map[string]interface{}, error), interval time.Duration) *Cache {
	return &Cache{
		loader:   loader,
		interval: interval,
		data:     make(map[string]interface{}),
		quit:     make(chan struct{}),
	}
}

// Load invokes loader() and swaps in the resulting map on success.
func (c *Cache) Load() {
	result, err := c.loader()
	if err != nil {
		log.Println("Cache load error:", err)
		return
	}
	c.mu.Lock()
	c.data = result
	c.mu.Unlock()
	log.Printf("[%s] Cache reloaded (%d items)\n", time.Now().Format(time.RFC3339), len(result))
}

// StartAutoReload begins a ticker that calls Load() every interval.
func (c *Cache) StartAutoReload() {
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

// StopAutoReload stops the periodic reload.
func (c *Cache) StopAutoReload() {
	close(c.quit)
}

// Get returns the value for a key, if present.
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.data[key]
	return val, ok
}

// GetAll returns a copy of the entire data map.
func (c *Cache) GetAll() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	copyMap := make(map[string]interface{}, len(c.data))
	for k, v := range c.data {
		copyMap[k] = v
	}
	return copyMap
}

// Find returns all cache items whose values satisfy the given predicate.
func (c *Cache) Find(predicate func(interface{}) bool) []interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var results []interface{}
	for _, v := range c.data {
		if predicate(v) {
			results = append(results, v)
		}
	}
	return results
}

// FindOne returns the first cache item whose value satisfies the predicate.
func (c *Cache) FindOne(predicate func(interface{}) bool) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, v := range c.data {
		if predicate(v) {
			return v, true
		}
	}
	return nil, false
}

// SetInterval allows updating the reload interval at runtime.
func (c *Cache) SetInterval(interval time.Duration) {
	c.mu.Lock()
	c.interval = interval
	if c.ticker != nil {
		c.ticker.Stop()
		c.ticker = time.NewTicker(c.interval)
	}
	c.mu.Unlock()
}
