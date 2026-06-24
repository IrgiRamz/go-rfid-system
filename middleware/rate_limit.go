package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/yimm/rfid-api/helpers"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	requests := rl.requests[key]
	var validRequests []time.Time
	for _, t := range requests {
		if t.After(windowStart) {
			validRequests = append(validRequests, t)
		}
	}

	if len(validRequests) >= rl.limit {
		rl.requests[key] = validRequests
		return false
	}

	validRequests = append(validRequests, now)
	rl.requests[key] = validRequests
	return true
}

func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		windowStart := now.Add(-rl.window)
		for key, requests := range rl.requests {
			var validRequests []time.Time
			for _, t := range requests {
				if t.After(windowStart) {
					validRequests = append(validRequests, t)
				}
			}
			if len(validRequests) == 0 {
				delete(rl.requests, key)
			} else {
				rl.requests[key] = validRequests
			}
		}
		rl.mu.Unlock()
	}
}

var (
	loginLimiter     = newRateLimiter(7, time.Minute)
	protectedLimiter = newRateLimiter(200, time.Minute)
)

func RateLimitLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !loginLimiter.allow(ip) {
			helpers.ErrorResponse(c, http.StatusTooManyRequests, "Too many requests. Please try again later.")
			c.Abort()
			return
		}
		c.Next()
	}
}

func RateLimitProtected() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !protectedLimiter.allow(ip) {
			helpers.ErrorResponse(c, http.StatusTooManyRequests, "Too many requests. Please try again later.")
			c.Abort()
			return
		}
		c.Next()
	}
}
