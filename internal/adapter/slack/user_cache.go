package slack

import "sync"

// UserCache provides thread-safe caching of user display names
type UserCache struct {
	cache map[string]string
	mu    sync.RWMutex
}

// NewUserCache creates a new UserCache
func NewUserCache() *UserCache {
	return &UserCache{
		cache: make(map[string]string),
	}
}

// Get retrieves a user display name from the cache
func (c *UserCache) Get(userID string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	displayName, ok := c.cache[userID]
	return displayName, ok
}

// Set stores a user display name in the cache
func (c *UserCache) Set(userID, displayName string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[userID] = displayName
}
