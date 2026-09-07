package inbound

import (
	"net/http"

	"github.com/andygeiss/cloud-native-utils/templating"
)

// HttpViewServiceWorker serves the tombstone service worker at /sw.js.
//
// Nothing registers a service worker any more. The route stays so that a
// browser still holding the old worker fetches this one on its next update
// check, which unregisters it and drops its caches. Delete the route, this
// handler and sw.tmpl once that deprecation window is over.
func HttpViewServiceWorker(e *templating.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		// Never cache a service worker: a cached tombstone would keep the old
		// worker alive on exactly the browsers this route exists to reach.
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		HttpView(e, "sw", nil)(w, r)
	}
}
