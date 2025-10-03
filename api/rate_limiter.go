package api

import (
	"net/http"
	"sync"
	"time"

	"github.com/Digital-Data-Co/forge/api/helpers"
)

// RateLimiter implements a simple in-memory rate limiter
type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
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

	// Clean up old entries periodically
	go rl.cleanup()
	return rl
}

// Allow checks if a request from the given key should be allowed
func (rl *RateLimiter) Allow(key string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Get existing requests for this key
	requests, exists := rl.requests[key]
	if !exists {
		requests = make([]time.Time, 0)
	}

	// Remove old requests outside the window
	var validRequests []time.Time
	for _, reqTime := range requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Check if we're under the limit
	if len(validRequests) >= rl.limit {
		return false
	}

	// Add current request
	validRequests = append(validRequests, now)
	rl.requests[key] = validRequests

	return true
}

// cleanup removes old entries to prevent memory leaks
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		cutoff := time.Now().Add(-rl.window * 2) // Keep some buffer

		for key, requests := range rl.requests {
			var validRequests []time.Time
			for _, reqTime := range requests {
				if reqTime.After(cutoff) {
					validRequests = append(validRequests, reqTime)
				}
			}

			if len(validRequests) == 0 {
				delete(rl.requests, key)
			} else {
				rl.requests[key] = validRequests
			}
		}
		rl.mutex.Unlock()
	}
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client IP
			clientIP := getClientIP(r)

			// Check rate limit
			if !rl.Allow(clientIP) {
				helpers.WriteJSON(w, http.StatusTooManyRequests, map[string]string{
					"error": "Rate limit exceeded. Please try again later.",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the real client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for reverse proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := len(xff); idx > 0 {
			for i, char := range xff {
				if char == ',' {
					idx = i
					break
				}
			}
			return xff[:idx]
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := len(ip); idx > 0 {
		for i, char := range ip {
			if char == ':' {
				idx = i
				break
			}
		}
		return ip[:idx]
	}

	return "unknown"
}

// Global rate limiters for different endpoints
var (
	// General API rate limiter: 100 requests per minute
	GeneralRateLimiter = NewRateLimiter(100, time.Minute)

	// Auth endpoints rate limiter: 10 requests per minute (more restrictive)
	AuthRateLimiter = NewRateLimiter(10, time.Minute)

	// File upload rate limiter: 5 uploads per minute
	UploadRateLimiter = NewRateLimiter(5, time.Minute)
)
