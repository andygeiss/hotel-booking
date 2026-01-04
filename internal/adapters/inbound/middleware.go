package inbound

import (
	"log/slog"
	"net/http"
	"time"
)

// WithLogging logs the request with method, path and duration.
func WithLogging(logger *slog.Logger, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next(w, r)

		// Log the request with method, path and duration.
		logger.Info(
			"http request handled",
			"method", r.Method,
			"path", r.RequestURI,
			"duration", time.Since(start),
		)
	}
}

// WithSecurityHeaders adds security headers to authenticated responses.
func WithSecurityHeaders(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Prevent sensitive information leakage via Referer header.
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Prevent caching of authenticated responses.
		w.Header().Set("Cache-Control", "no-store")

		// Prevent content type sniffing.
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking.
		w.Header().Set("X-Frame-Options", "DENY")

		next(w, r)
	}
}
