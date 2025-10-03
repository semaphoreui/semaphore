package api

import (
	"net/http"
)

// SecurityHeadersMiddleware adds security headers to all responses
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Strict Transport Security (HSTS)
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// Content Security Policy
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " + // Allow inline scripts for Vue.js
			"style-src 'self' 'unsafe-inline'; " + // Allow inline styles
			"img-src 'self' data: https:; " + // Allow images from self, data URLs, and HTTPS
			"font-src 'self' data:; " + // Allow fonts from self and data URLs
			"connect-src 'self' ws: wss:; " + // Allow WebSocket connections
			"frame-ancestors 'none'; " + // Prevent clickjacking
			"base-uri 'self'; " + // Restrict base tag
			"form-action 'self'; " + // Restrict form submissions
			"upgrade-insecure-requests" // Upgrade HTTP to HTTPS

		w.Header().Set("Content-Security-Policy", csp)

		// X-Frame-Options (redundant with CSP frame-ancestors, but good for older browsers)
		w.Header().Set("X-Frame-Options", "DENY")

		// X-Content-Type-Options
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// X-XSS-Protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Referrer Policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions Policy (formerly Feature Policy)
		permissionsPolicy := "geolocation=(), " +
			"microphone=(), " +
			"camera=(), " +
			"payment=(), " +
			"usb=(), " +
			"magnetometer=(), " +
			"gyroscope=(), " +
			"speaker=(), " +
			"vibrate=(), " +
			"fullscreen=(self), " +
			"sync-xhr=()"

		w.Header().Set("Permissions-Policy", permissionsPolicy)

		// Cache Control for sensitive endpoints
		if isSensitiveEndpoint(r.URL.Path) {
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		}

		next.ServeHTTP(w, r)
	})
}

// isSensitiveEndpoint checks if the endpoint should have no-cache headers
func isSensitiveEndpoint(path string) bool {
	sensitivePaths := []string{
		"/api/auth/",
		"/api/user/",
		"/api/users",
		"/api/projects/",
		"/api/compliance/",
	}

	for _, sensitivePath := range sensitivePaths {
		if len(path) >= len(sensitivePath) && path[:len(sensitivePath)] == sensitivePath {
			return true
		}
	}

	return false
}

// CORSHeadersMiddleware adds proper CORS headers
func CORSHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Allow specific origins or same origin
		if origin == "" || isAllowedOrigin(origin, r.Host) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			// For same-origin requests, allow the origin
			w.Header().Set("Access-Control-Allow-Origin", "null")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, HEAD")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-Runner-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isAllowedOrigin checks if an origin is allowed
func isAllowedOrigin(origin, host string) bool {
	// Allow same origin
	if origin == "https://"+host || origin == "http://"+host {
		return true
	}

	// Allow localhost for development
	if origin == "http://localhost:3000" || origin == "https://localhost:3000" {
		return true
	}

	// Add your production domains here
	allowedOrigins := []string{
		// Add your production domains here
		// "https://yourdomain.com",
		// "https://app.yourdomain.com",
	}

	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return true
		}
	}

	return false
}
