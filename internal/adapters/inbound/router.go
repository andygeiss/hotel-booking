package inbound

import (
	"context"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/andygeiss/cloud-native-utils/logging"
	"github.com/andygeiss/cloud-native-utils/mcp"
	"github.com/andygeiss/cloud-native-utils/templating"
	"github.com/andygeiss/cloud-native-utils/web"
	"github.com/andygeiss/hotel-booking/internal/domain/reservation"
	"github.com/coreos/go-oidc/v3/oidc"
)

// RouterConfig holds all dependencies for HTTP routing.
type RouterConfig struct {
	Ctx                context.Context
	EFS                fs.FS
	Logger             *slog.Logger
	MCPServer          *mcp.Server // Optional: nil disables MCP endpoint
	ReservationService *reservation.Service
	Verifier           *oidc.IDTokenVerifier // Required if MCPServer is set
}

// Route creates a new mux with the liveness and readiness probe (/liveness, /readiness),
// the static assets endpoint (/) and the ui endpoints (/ui).
// The EFS field in config accepts any fs.FS implementation (embed.FS, fs.Sub result, etc.).
func Route(config RouterConfig) *http.ServeMux {
	// Create a new mux with liveness and readyness endpoint.
	// Embed the assets into the mux.
	mux, serverSessions := web.NewServeMux(config.Ctx, config.EFS)

	// Create a new templating engine.
	// We use the fs.FS to load the templates from the file system.
	// We use the templating.Engine from cloud-native-utils and reuse it for all views.
	e := templating.NewEngine(config.EFS)

	// Parse the templates under the assets/templates directory.
	// Every template must have a .tmpl extension.
	e.Parse("assets/templates/*.tmpl")

	// The static assets are served from the embed.FS under the /static path directly.
	// This is defined in the web.NewServeMux function from cloud-native-utils.

	// Add the index endpoint for the UI.
	// The HttpViewIndex is handling unauthenticated and authenticated requests.
	// The unauthenticated requests are redirected to the login page /ui/login.
	// The authenticated requests are rendered with the index template.
	mux.HandleFunc("GET /ui/", logging.WithLogging(config.Logger, web.WithAuth(serverSessions, HttpViewIndex(e))))

	// Add the login endpoint for the UI.
	// This endpoint is used to forward the user to the login page of the OIDC provider.
	mux.HandleFunc("GET /ui/login", logging.WithLogging(config.Logger, HttpViewLogin(e)))

	// Add the error endpoint for displaying user-friendly error pages.
	// This endpoint accepts query parameters: title, message, and details.
	mux.HandleFunc("GET /ui/error", logging.WithLogging(config.Logger, HttpViewError(e)))

	// Add the manifest endpoint for the PWA.
	// This endpoint serves the manifest.json file for Progressive Web App support.
	mux.HandleFunc("GET /manifest.json", logging.WithLogging(config.Logger, HttpViewManifest(e)))

	// Add the service worker endpoint for the PWA.
	// This endpoint serves the sw.js file for offline caching and installability.
	mux.HandleFunc("GET /sw.js", logging.WithLogging(config.Logger, HttpViewServiceWorker(e)))

	// Add the reservations list endpoint.
	mux.HandleFunc("GET /ui/reservations", logging.WithLogging(config.Logger, web.WithAuth(serverSessions, HttpViewReservations(e, config.ReservationService))))

	// Add the new reservation form endpoint.
	mux.HandleFunc("GET /ui/reservations/new", logging.WithLogging(config.Logger, web.WithAuth(serverSessions, HttpViewReservationForm(e))))

	// Add the create reservation endpoint.
	mux.HandleFunc("POST /ui/reservations", logging.WithLogging(config.Logger, web.WithAuth(serverSessions, HttpCreateReservation(e, config.ReservationService))))

	// Add the reservation detail endpoint.
	mux.HandleFunc("GET /ui/reservations/{id}", logging.WithLogging(config.Logger, web.WithAuth(serverSessions, HttpViewReservationDetail(e, config.ReservationService))))

	// Add the cancel reservation endpoint.
	mux.HandleFunc("POST /ui/reservations/{id}/cancel", logging.WithLogging(config.Logger, web.WithAuth(serverSessions, HttpCancelReservation(config.ReservationService))))

	// Add MCP endpoint if configured.
	if config.MCPServer != nil {
		mcpHandler := web.NewMCPHandler(config.MCPServer)
		if config.Verifier != nil {
			mux.Handle("POST /mcp", logging.WithLogging(config.Logger, web.WithBearerAuth(config.Verifier, mcpHandler.Handler())))
		} else {
			mux.Handle("POST /mcp", logging.WithLogging(config.Logger, mcpHandler.Handler()))
		}
	}

	return mux
}
