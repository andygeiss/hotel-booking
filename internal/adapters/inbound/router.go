package inbound

import (
	"context"
	"embed"
	"log/slog"
	"net/http"

	"github.com/andygeiss/cloud-native-utils/logging"
	"github.com/andygeiss/cloud-native-utils/security"
	"github.com/andygeiss/cloud-native-utils/templating"
)

// Route creates a new mux with the liveness and readiness probe (/liveness, /readiness),
// the static assets endpoint (/) and the ui endpoints (/ui).
func Route(ctx context.Context, efs embed.FS, logger *slog.Logger) *http.ServeMux {
	// Create a new mux with liveness and readyness endpoint.
	// Embed the assets into the mux.
	mux, serverSessions := security.NewServeMux(ctx, efs)

	// Create a new templating engine.
	// We use the embed.FS to load the templates from the file system.
	// We use the templating.Engine from cloud-native-utils and reuse it for all views.
	e := templating.NewEngine(efs)

	// Parse the templates under the assets/templates directory.
	// Every template must have a .tmpl extension.
	e.Parse("assets/templates/*.tmpl")

	// The static assets are served from the embed.FS under the /static path directly.
	// This is defined in the security.NewServeMux function from cloud-native-utils.
	// There is a /static/keepalive.txt with a keepalive message (OK).
	// We can use this to check if the server is alive.
	//
	// curl http://localhost:8080/static/keepalive.txt

	// Add the index endpoint for the UI.
	// The HttpViewIndex is handling unauthenticated and authenticated requests.
	// The unauthenticated requests are redirected to the login page /ui/login.
	// The authenticated requests are rendered with the index template.
	mux.HandleFunc("GET /ui/", logging.WithLogging(logger, security.WithAuth(serverSessions, HttpViewIndex(e))))

	// Add session-aware index endpoints (for OIDC callback redirects).
	// We support both /ui/{session_id}/ and /ui/{session_id} endpoints.
	// This is important because some OIDC providers redirect to the endpoint without the trailing slash.
	mux.HandleFunc("GET /ui/{session_id}/", logging.WithLogging(logger, security.WithAuth(serverSessions, HttpViewIndex(e))))
	mux.HandleFunc("GET /ui/{session_id}", logging.WithLogging(logger, security.WithAuth(serverSessions, HttpViewIndex(e))))

	// Add the login endpoint for the UI.
	// This endpoint is used to forward the user to the login page of the OIDC provider.
	mux.HandleFunc("GET /ui/login", logging.WithLogging(logger, HttpViewLogin(e)))

	return mux
}
