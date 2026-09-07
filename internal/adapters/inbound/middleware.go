package inbound

import "net/http"

// maxRequestBody caps every request body at 1 MiB. Without a cap, ParseForm on
// a hostile body reads until the process runs out of memory.
const maxRequestBody = 1 << 20

// csp is the whole Content-Security-Policy, built once. It is a constant on
// purpose: a policy assembled per request is a policy that can differ per
// request, and nobody can review that.
//
// img-src needs data: because the CSS mask icons are data:image/svg+xml URLs.
// Drop it and every icon in the app disappears. frame-ancestors, base-uri and
// form-action have no default-src fallback, so each one is written out.
const csp = "default-src 'self'; " +
	"img-src 'self' data:; " +
	"frame-ancestors 'none'; " +
	"base-uri 'none'; " +
	"form-action 'self'"

// secureHeaders sends the four security headers on every response.
//
// The headers go out before the next handler runs. A header set after a
// handler has called WriteHeader is dropped without a word, because the status
// line is already on the wire.
func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", csp)
		h.Set("Strict-Transport-Security", "max-age=31536000")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "same-origin")
		next.ServeHTTP(w, r)
	})
}

// WithSecurity wraps a handler in the security chain: headers first, then the
// CSRF check, then the body cap.
//
// The CSRF check rejects unsafe cross-origin requests using Sec-Fetch-Site,
// falling back to Origin against Host. A client that sends neither header is
// not a browser, so it passes and the MCP endpoint keeps working.
//
// The body cap sits innermost. An outer cap cannot be raised further in: the
// body is already wrapped in the smaller reader by the time a handler runs, so
// a route that one day accepts uploads picks its own limit at the cap site.
func WithSecurity(next http.Handler) http.Handler {
	csrf := http.NewCrossOriginProtection()
	return secureHeaders(csrf.Handler(http.MaxBytesHandler(next, maxRequestBody)))
}
