package middlewares

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type limiter struct {
	mu      sync.Mutex
	count   int
	resetAt time.Time
}

type inMemoryStore struct {
	mu      sync.RWMutex
	clients map[string]*limiter
	limit   int
	window  time.Duration
}

func NewInMemoryStore(limit int, window time.Duration) *inMemoryStore {
	return &inMemoryStore{
		clients: make(map[string]*limiter),
		limit:   limit,
		window:  window,
	}
}

func (s *inMemoryStore) Allow(key string) bool {
	s.mu.RLock()
	l, exists := s.clients[key]
	s.mu.RUnlock()

	now := time.Now()
	if !exists {
		s.mu.Lock()
		s.clients[key] = &limiter{
			count:   1,
			resetAt: now.Add(s.window),
		}
		s.mu.Unlock()
		return true
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if now.After(l.resetAt) {
		l.count = 1
		l.resetAt = now.Add(s.window)
		return true
	}

	if l.count < s.limit {
		l.count++
		return true
	}
	return false
}

func RateLimit(store *inMemoryStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		if !store.Allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests, please slow down",
			})
			return
		}
		c.Next()
	}
}
