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
	s := &inMemoryStore{
		clients: make(map[string]*limiter),
		limit:   limit,
		window:  window,
	}

	go s.cleanup()

	return s
}

func (s *inMemoryStore) cleanup() {
	ticker := time.NewTicker(s.window)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()

		s.mu.Lock()
		for key, l := range s.clients {
			l.mu.Lock()
			expired := now.After(l.resetAt)
			l.mu.Unlock()

			if expired {
				delete(s.clients, key)
			}
		}
		s.mu.Unlock()
	}
}

func (s *inMemoryStore) Allow(key string) bool {
	now := time.Now()

	s.mu.RLock()
	l := s.clients[key]
	s.mu.RUnlock()

	if l == nil {
		s.mu.Lock()

		if l = s.clients[key]; l == nil {
			s.clients[key] = &limiter{
				count:   1,
				resetAt: now.Add(s.window),
			}
			s.mu.Unlock()
			return true
		}

		s.mu.Unlock()
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if now.After(l.resetAt) {
		l.count = 1
		l.resetAt = now.Add(s.window)
		return true
	}

	if l.count >= s.limit {
		return false
	}

	l.count++
	return true
}

func RateLimit(store *inMemoryStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !store.Allow(c.ClientIP()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests",
			})
			return
		}

		c.Next()
	}
}
