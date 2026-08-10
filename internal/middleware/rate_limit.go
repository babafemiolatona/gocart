package middleware

import (
	"net/http"
	"sync"
	"time"

	apperrors "gocart/internal/errors"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	hits   map[string][]time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limit:  limit,
		window: window,
		hits:   make(map[string][]time.Time),
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	now := time.Now()
	cutoff := now.Add(-rl.window)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	times := rl.hits[key]
	kept := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	rl.hits[key] = kept

	if len(kept) >= rl.limit {
		return false
	}

	rl.hits[key] = append(kept, now)
	return true
}

func RateLimitMiddleware(rl *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.Allow(c.ClientIP()) {
			c.Error(apperrors.New(
				http.StatusTooManyRequests,
				apperrors.CodeTooManyRequests,
				"too many requests, please try again later",
				nil,
			))
			c.Abort()
			return
		}
		c.Next()
	}
}
