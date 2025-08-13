package middleware

import (
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements a simple rate limiter
type RateLimiter struct {
	clients map[string]*ClientInfo
	mutex   sync.RWMutex
	limit   int
	window  time.Duration
}

// ClientInfo tracks client request information
type ClientInfo struct {
	requests  int
	resetTime time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*ClientInfo),
		limit:   limit,
		window:  window,
	}
}

// RateLimit returns a middleware that implements rate limiting
func RateLimit() gin.HandlerFunc {
	// Check if rate limiting should be disabled for load testing
	if os.Getenv("DISABLE_RATE_LIMITING") == "true" {
		return gin.HandlerFunc(func(c *gin.Context) {
			c.Next()
		})
	}

	limiter := NewRateLimiter(10000, time.Minute) // 10,000 requests per minute for high-concurrency testing

	return gin.HandlerFunc(func(c *gin.Context) {
		clientIP := c.ClientIP()

		if !limiter.Allow(clientIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": limiter.window.Seconds(),
			})
			c.Abort()
			return
		}

		c.Next()
	})
}

// Allow checks if a client is allowed to make a request
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	client, exists := rl.clients[clientID]

	if !exists {
		rl.clients[clientID] = &ClientInfo{
			requests:  1,
			resetTime: now.Add(rl.window),
		}
		return true
	}

	// Reset if window has passed
	if now.After(client.resetTime) {
		client.requests = 1
		client.resetTime = now.Add(rl.window)
		return true
	}

	// Check if under limit
	if client.requests < rl.limit {
		client.requests++
		return true
	}

	return false
}
