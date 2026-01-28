package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/PersonaForge/backend/internal/response"
	"github.com/gin-gonic/gin"
)

// RateLimiter implements a simple in-memory rate limiter
type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}

	// Cleanup old entries every minute
	go rl.cleanup()

	return rl
}

// Middleware returns a Gin middleware function
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get identifier (IP or session/user ID)
		identifier := c.ClientIP()
		if userID, exists := c.Get("user_id"); exists {
			identifier = fmt.Sprintf("user:%v", userID)
		} else if sessionID, exists := c.Get("session_id"); exists {
			identifier = sessionID.(string)
		}

		if !rl.Allow(identifier) {
			response.TooManyRequests(c, "Rate limit exceeded. Please try again later.")
			c.Abort()
			return
		}

		c.Next()
	}
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow(identifier string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Get existing requests
	requests := rl.requests[identifier]

	// Filter out old requests
	var validRequests []time.Time
	for _, t := range requests {
		if t.After(cutoff) {
			validRequests = append(validRequests, t)
		}
	}

	// Check if limit exceeded
	if len(validRequests) >= rl.limit {
		rl.requests[identifier] = validRequests
		return false
	}

	// Add current request
	validRequests = append(validRequests, now)
	rl.requests[identifier] = validRequests

	return true
}

// cleanup removes old entries periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-rl.window)

		for key, requests := range rl.requests {
			var validRequests []time.Time
			for _, t := range requests {
				if t.After(cutoff) {
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

// RateLimitConfig holds rate limit configuration for different user types
type RateLimitConfig struct {
	FreeLimit int
	AuthLimit int
	Window    time.Duration
}

// AdaptiveRateLimit creates a rate limiter that adjusts based on user type
func AdaptiveRateLimit(config RateLimitConfig) gin.HandlerFunc {
	freeLimiter := NewRateLimiter(config.FreeLimit, config.Window)
	authLimiter := NewRateLimiter(config.AuthLimit, config.Window)

	return func(c *gin.Context) {
		identifier := c.ClientIP()
		limiter := freeLimiter

		// Check if user is authenticated
		if userID, exists := c.Get("user_id"); exists {
			identifier = fmt.Sprintf("user:%v", userID)
			limiter = authLimiter
		} else if sessionID, exists := c.Get("session_id"); exists {
			identifier = sessionID.(string)
		}

		if !limiter.Allow(identifier) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"status":  "error",
				"message": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
