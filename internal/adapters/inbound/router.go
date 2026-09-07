package inbound

import (
	"context"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/andygeiss/cloud-native-utils/logging"
	"github.com/andygeiss/cloud-native-utils/mcp"
	"github.com/andygeiss/cloud-native-utils/web"
	"github.com/andygeiss/hotel-booking/internal/domain/reservation"
)

// RouterConfig holds all dependencies for HTTP routing.
type RouterConfig struct {
	App                AppInfo // Name, description and version every page renders
	Ctx                context.Context
	EFS                fs.FS
	Logger             *slog.Logger
	MCPServer          *mcp.Server  // Optional: nil disables MCP endpoint
	NewVerifier        VerifierFunc // Required if MCPServer is set; called on the first /mcp request, never here
	ReservationService *reservation.Service
}

// Route creates a new mux with the liveness and readiness probe (/liveness, /readiness),
// the static assets endpoint (/) and the ui endpoints (/ui).
// The EFS field in config accepts any fs.FS implementation (embed.FS, fs.Sub result, etc.).
// Every route it returns is wrapped in WithSecurity, so no handler can be
// registered outside the security headers, the CSRF check and the body cap.
func Route(config RouterConfig) *http.ServeMux {
	// Create a new mux with liveness and readyness endpoint.
	// Embed the assets into the mux.
	mux, serverSessions := web.NewServeMux(config.Ctx, config.EFS)

	// Parse every page at boot: the layout shell plus one file per page, one
	// template set each. A template that does not parse is a programming error,
	// so it stops the process here rather than the first request that needs it.
	v, err := NewView(config.EFS, config.Logger)
	if err != nil {
		panic("inbound: " + err.Error())
	}

	// The static assets are served from the embed.FS under the /static path directly.
	// This is defined in the web.NewServeMux function from cloud-native-utils.

	// Add the index endpoint for the UI.
	// The HttpViewIndex is handling unauthenticated and authenticated requests.
	// The unauthenticated requests are redirected to the login page /ui/login.
	// The authenticated requests are rendered with the index template.
	mux.HandleFunc("GET /ui/", logging.WithLogging(config.Logger, web.WithAuth(serverSessions, HttpViewIndex(v, config.App))))

	// Add the login endpoint for the UI.
	// This endpoint is used to forward the user to the login page of the OIDC provider.
	mux.HandleFunc("GET /ui/login", logging.WithLogging(config.Logger, HttpViewLogin(v, config.App)))

	// Add the error endpoint for displaying user-friendly error pages.
	// This endpoint accepts query parameters: title, message, and details.
	mux.HandleFunc("GET /ui/error", logging.WithLogging(config.Logger, HttpViewError(v, config.App)))

	// Add the manifest endpoint for the PWA.
	// This endpoint serves the manifest.json file for Progressive Web App support.
	mux.HandleFunc("GET /manifest.json", logging.WithLogging(config.Logger, HttpViewManifest(config.App)))

	// Add the reservations list endpoint.
	mux.HandleFunc("GET /ui/reservations", logging.WithLogging(config.Logger, web.WithAuth(serverSessions, HttpViewReservations(v, config.App, config.ReservationService))))

	// Add the new reservation form endpoint.
	mux.HandleFunc("GET /ui/reservations/new", logging.WithLogging(config.Logger, web.WithAuth(serverSessions, HttpViewReservationForm(v, config.App))))

	// Add the create reservation endpoint.
	mux.HandleFunc("POST /ui/reservations", logging.WithLogging(config.Logger, web.WithAuth(serverSessions, HttpCreateReservation(v, config.App, config.ReservationService))))

	// Add the reservation detail endpoint.
	mux.HandleFunc("GET /ui/reservations/{id}", logging.WithLogging(config.Logger, web.WithAuth(serverSessions, HttpViewReservationDetail(v, config.App, config.ReservationService))))

	// Add the cancel reservation endpoint.
	mux.HandleFunc("POST /ui/reservations/{id}/cancel", logging.WithLogging(config.Logger, web.WithAuth(serverSessions, HttpCancelReservation(v, config.App, config.ReservationService))))

	// Add MCP endpoint if configured.
	if config.MCPServer != nil {
		mcpHandler := web.NewMCPHandler(config.MCPServer)
		if config.NewVerifier != nil {
			mux.Handle("POST /mcp", logging.WithLogging(config.Logger, WithBearerAuth(config.NewVerifier, config.Logger, mcpHandler.Handler())))
		} else {
			mux.Handle("POST /mcp", logging.WithLogging(config.Logger, mcpHandler.Handler()))
		}
	}

	// Hand back a root mux whose single route is the application behind the
	// security chain. Wrapping here rather than in main is what makes the
	// chain impossible to forget: there is no way to reach a handler that
	// skips it.
	root := http.NewServeMux()
	root.Handle("/", WithSecurity(mux))

	return root
}
