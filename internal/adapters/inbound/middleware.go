package inbound

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/andygeiss/cloud-native-utils/web"
	"github.com/coreos/go-oidc/v3/oidc"
)

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

// VerifierFunc builds the OIDC verifier that the MCP endpoint checks bearer
// tokens against. It is a function rather than a verifier because building one
// calls the identity provider, and that must not happen while the app starts.
//
// It takes no context on purpose. go-oidc keeps the context it was built with
// for the key set it refetches on every token verification, so the caller must
// hand it one that lives as long as the process — a request's context would
// stop working the moment that request ended, and every later verification
// would fail. Bound the call with the HTTP client's own timeout instead.
type VerifierFunc func() (*oidc.IDTokenVerifier, error)

// WithBearerAuth checks the MCP endpoint's bearer token, building the verifier
// on the first request rather than at boot.
//
// Building it fetches the issuer's discovery document, which is a call to
// somebody else's process. Doing that in main made the app refuse to start
// whenever the identity provider was down — their outage became ours — and it
// meant the binary could not start with an empty environment at all.
//
// Only success is cached. A provider that was down when the first request
// arrived is asked again on the next one, so the endpoint recovers by itself.
func WithBearerAuth(newVerifier VerifierFunc, logger *slog.Logger, next http.HandlerFunc) http.HandlerFunc {
	var (
		mu       sync.Mutex
		verifier *oidc.IDTokenVerifier
	)

	// The mutex is held across the call, so a burst of first requests asks the
	// provider once instead of all at once.
	resolve := func() (*oidc.IDTokenVerifier, error) {
		mu.Lock()
		defer mu.Unlock()
		if verifier != nil {
			return verifier, nil
		}
		v, err := newVerifier()
		if err != nil {
			return nil, err
		}
		verifier = v
		return v, nil
	}

	return func(w http.ResponseWriter, r *http.Request) {
		v, err := resolve()
		if err != nil {
			logger.Error("cannot verify bearer tokens: the identity provider did not answer",
				"error", err,
				"fix", "check that the OIDC issuer is reachable from this process")
			// 503 rather than 401: the token was never read, so the client
			// should retry instead of going to look for new credentials.
			http.Error(w, "identity provider unavailable", http.StatusServiceUnavailable)
			return
		}
		web.WithBearerAuth(v, next)(w, r)
	}
}
