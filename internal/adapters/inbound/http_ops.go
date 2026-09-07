package inbound

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"time"
)

// OpsHandler serves /healthz and the pprof endpoints for the operator.
//
// It is a separate handler because it belongs on a separate, loopback-only
// listener. Being unreachable from anywhere but the machine itself is the only
// access control these endpoints have, so they must never join the application
// mux, where the proxy would expose them to the internet.
//
// ping reports whether the databases answer. version names the build, so
// whoever is looking at a sick process can tell which one it is.
func OpsHandler(version string, ping func(ctx context.Context) error) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		// A one-second budget of its own. A database that hangs must show up
		// as 503, not as a health poll that never answers.
		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()

		status, state := http.StatusOK, "ok"
		if err := ping(ctx); err != nil {
			status, state = http.StatusServiceUnavailable, "unavailable"
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = fmt.Fprintf(w, `{"status":%q,"version":%q}`, state, version)
	})

	// Registered by hand: the blank net/http/pprof import registers on
	// http.DefaultServeMux, which nothing here ever serves. Index dispatches
	// the named runtime profiles itself, so only these four need a route of
	// their own. They name no method because Symbol answers GET and POST.
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return mux
}
