package services

import (
	"fmt"
	"sync"
	"time"
)

// CacheService provides caching functionality
type CacheService struct {
	cache map[string]*CacheEntry
	mutex sync.RWMutex
}

// CacheEntry represents a cached item
type CacheEntry struct {
	Value     interface{}
	ExpiresAt time.Time
}

// NewCacheService creates a new cache service
func NewCacheService() *CacheService {
	return &CacheService{
		cache: make(map[string]*CacheEntry),
	}
}

// Get retrieves a value from cache
func (cs *CacheService) Get(key string) (interface{}, error) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	entry, exists := cs.cache[key]
	if !exists {
		return nil, ErrCacheNotFound
	}

	if time.Now().After(entry.ExpiresAt) {
		delete(cs.cache, key)
		return nil, ErrCacheExpired
	}

	return entry.Value, nil
}

// Set stores a value in cache
func (cs *CacheService) Set(key string, value interface{}, ttl time.Duration) error {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	cs.cache[key] = &CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}

	return nil
}

// Delete removes a value from cache
func (cs *CacheService) Delete(key string) error {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	delete(cs.cache, key)
	return nil
}

// Clear clears all cache entries
func (cs *CacheService) Clear() error {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	cs.cache = make(map[string]*CacheEntry)
	return nil
}

// Custom errors
var (
	ErrCacheNotFound = fmt.Errorf("cache entry not found")
	ErrCacheExpired  = fmt.Errorf("cache entry expired")
)
