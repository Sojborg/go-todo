package cache

import (
	"sync"
	"time"
)

type TokenInfo struct {
	UserID    string
	Email     string
	ExpiresAt time.Time
}

type TokenCache struct {
	cache map[string]TokenInfo
	mu    sync.RWMutex
}

func NewTokenCache() *TokenCache {
	return &TokenCache{
		cache: make(map[string]TokenInfo),
	}
}

func (tc *TokenCache) Set(token string, info TokenInfo) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.cache[token] = info
}

func (tc *TokenCache) Get(token string) (TokenInfo, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	info, exists := tc.cache[token]
	if !exists {
		return TokenInfo{}, false
	}

	if time.Now().After(info.ExpiresAt) {
		tc.mu.Lock()
		delete(tc.cache, token)
		tc.mu.Unlock()
		return TokenInfo{}, false
	}

	return info, true
}
