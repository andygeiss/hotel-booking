package main

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/andygeiss/cloud-native-utils/env"
	"github.com/andygeiss/cloud-native-utils/logging"
	"github.com/andygeiss/cloud-native-utils/mcp"
	"github.com/andygeiss/cloud-native-utils/messaging"
	"github.com/andygeiss/cloud-native-utils/resource"
	"github.com/andygeiss/cloud-native-utils/service"
	"github.com/andygeiss/cloud-native-utils/web"
	"github.com/andygeiss/hotel-booking/internal/adapters/inbound"
	"github.com/andygeiss/hotel-booking/internal/adapters/outbound"
	"github.com/andygeiss/hotel-booking/internal/domain/orchestration"
	"github.com/andygeiss/hotel-booking/internal/domain/payment"
	"github.com/andygeiss/hotel-booking/internal/domain/reservation"
	"github.com/coreos/go-oidc/v3/oidc"
	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed assets
var efs embed.FS

// buildVersion reports the module version this binary was built from.
// debug.ReadBuildInfo is the only source: a version baked in with -ldflags
// drifts from the tag that produced it, and nobody notices until an incident.
func buildVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" {
		return "unknown"
	}
	return info.Main.Version
}

// buildMCPServer creates the MCP server with all tools registered.
func buildMCPServer(
	reservationService *reservation.Service,
	availabilityChecker reservation.AvailabilityChecker,
	paymentService *payment.Service,
) *mcp.Server {
	server := mcp.NewServer(
		env.Get("APP_SHORTNAME", "mcp-server"),
		env.Get("APP_VERSION", "1.0.0"),
	)

	// Register tools from each bounded context.
	reservation.RegisterTools(server, reservationService, availabilityChecker)
	payment.RegisterTools(server, paymentService)

	return server
}

func main() {
	// Create a new context with a cancel function.
	ctx, cancel := service.Context()
	defer cancel()

	// Create a new logger.
	// We use the logging.NewJsonLogger function from the cloud-native-utils/logging package.
	logger := logging.NewJsonLogger()

	// Initialize Reservation Database connection.
	reservationDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		env.Get("RESERVATION_DB_HOST", "localhost"),
		env.Get("RESERVATION_DB_PORT", "5432"),
		env.Get("RESERVATION_DB_USER", "reservation"),
		env.Get("RESERVATION_DB_PASSWORD", "reservation_secret"),
		env.Get("RESERVATION_DB_NAME", "reservation_db"),
		env.Get("RESERVATION_DB_SSLMODE", "disable"),
	)
	reservationDB, err := sql.Open("pgx", reservationDSN)
	if err != nil {
		logger.Error("failed to connect to reservation database", "error", err)
		os.Exit(1)
	}
	defer reservationDB.Close()

	// Initialize Payment Database connection.
	paymentDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		env.Get("PAYMENT_DB_HOST", "localhost"),
		env.Get("PAYMENT_DB_PORT", "5433"),
		env.Get("PAYMENT_DB_USER", "payment"),
		env.Get("PAYMENT_DB_PASSWORD", "payment_secret"),
		env.Get("PAYMENT_DB_NAME", "payment_db"),
		env.Get("PAYMENT_DB_SSLMODE", "disable"),
	)
	paymentDB, err := sql.Open("pgx", paymentDSN)
	if err != nil {
		logger.Error("failed to connect to payment database", "error", err)
		os.Exit(1)
	}
	defer paymentDB.Close()

	// Shared event dispatcher using Kafka for distributed event messaging.
	dispatcher := messaging.NewExternalDispatcher()

	// Initialize reservation bounded context using PostgresAccess from cloud-native-utils.
	// Schema is created by Docker init scripts (migrations/reservation/init.sql).
	reservationRepo := resource.NewPostgresAccess[reservation.ReservationID, reservation.Reservation](reservationDB)
	availabilityChecker := outbound.NewRepositoryAvailabilityChecker(reservationRepo)
	reservationPublisher := outbound.NewEventPublisher(dispatcher)
	reservationService := reservation.NewService(reservationRepo, availabilityChecker, reservationPublisher)

	// Initialize payment bounded context using PostgresAccess from cloud-native-utils.
	paymentRepo := resource.NewPostgresAccess[payment.PaymentID, payment.Payment](paymentDB)
	paymentGateway := outbound.NewMockPaymentGateway()
	paymentPublisher := outbound.NewEventPublisher(dispatcher)
	paymentService := payment.NewService(paymentRepo, paymentGateway, paymentPublisher)

	// Initialize orchestration layer.
	notificationService := outbound.NewMockNotificationService(logger)
	bookingService := orchestration.NewBookingService(reservationService, paymentService, notificationService)

	// Register cross-context event handlers.
	eventHandlers := orchestration.NewEventHandlers(bookingService, reservationService, paymentService)
	if err := eventHandlers.RegisterHandlers(ctx, dispatcher); err != nil {
		logger.Error("failed to register event handlers", "error", err)
		os.Exit(1)
	}

	// Initialize OIDC provider for MCP token verification.
	// This connects to Keycloak to validate Bearer tokens for the MCP endpoint.
	// Reuses the existing OIDC_ISSUER environment variable for consistency.
	oidcIssuer := env.Get("OIDC_ISSUER", "http://localhost:8180/realms/local")
	provider, err := oidc.NewProvider(ctx, oidcIssuer)
	if err != nil {
		logger.Error("failed to initialize OIDC provider", "error", err)
		os.Exit(1)
	}

	// Configure token verifier for MCP client.
	// Uses a separate client ID for machine-to-machine MCP authentication.
	mcpClientID := env.Get("MCP_CLIENT_ID", "hotel-booking-mcp")
	verifier := provider.Verifier(&oidc.Config{ClientID: mcpClientID})

	// Build the MCP server with all tools registered.
	mcpServer := buildMCPServer(reservationService, availabilityChecker, paymentService)

	// Create router with all dependencies via RouterConfig.
	mux := inbound.Route(inbound.RouterConfig{
		Ctx:                ctx,
		EFS:                efs,
		Logger:             logger,
		ReservationService: reservationService,
		MCPServer:          mcpServer,
		Verifier:           verifier,
	})

	// Start the ops listener. /healthz and the pprof endpoints answer here and
	// nowhere else: the address is loopback, so the proxy cannot reach them and
	// being unreachable is their whole access control.
	opsSrv := &http.Server{
		Addr: "127.0.0.1:6060",
		Handler: inbound.OpsHandler(buildVersion(), func(ctx context.Context) error {
			if err := reservationDB.PingContext(ctx); err != nil {
				return fmt.Errorf("pinging reservation database: %w", err)
			}
			if err := paymentDB.PingContext(ctx); err != nil {
				return fmt.Errorf("pinging payment database: %w", err)
			}
			return nil
		}),
		IdleTimeout:       2 * time.Minute,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		// 30 seconds, well above the app listener, because a pprof profile
		// writes nothing until it finishes sampling.
		WriteTimeout: 30 * time.Second,
	}

	// How this goroutine stops: the shutdown below closes the listener when the
	// context is done, and ListenAndServe then returns http.ErrServerClosed.
	service.RegisterOnContextDone(ctx, func() {
		_ = opsSrv.Shutdown(context.Background())
	})
	go func() {
		if err := opsSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			// The app keeps serving without /healthz and pprof, so say plainly
			// what is gone and what frees the port. 6060 is the conventional
			// Go pprof port, so a second local service is the usual cause.
			logger.Error("ops listener failed: /healthz and /debug/pprof are unavailable",
				"addr", opsSrv.Addr,
				"error", err,
				"fix", "stop whatever else listens on 127.0.0.1:6060 (lsof -nP -iTCP:6060 -sTCP:LISTEN)")
		}
	}()

	srv := web.NewServer(mux)
	defer func() { _ = srv.Close() }()

	// Register the server shutdown function on the context done function.
	// We use the RegisterOnContextDone function from the cloud-native-utils/service package.
	// The server.Shutdown function waits for 5 seconds before shutting down the server.
	service.RegisterOnContextDone(ctx, func() {
		_ = srv.Shutdown(context.Background())
	})

	// The server implementation from the cloud-native-utils/web package uses
	// It uses the PORT environment variable to determine the port to listen on.
	// If the PORT environment variable is not set, it defaults to port 8080.
	logger.Info("server initialized", "port", env.Get("PORT", "8080"))

	// Start the HTTP server in the main goroutine.
	if err := srv.ListenAndServe(); err != nil {
		// Check if the server was closed intentionally.
		if errors.Is(err, http.ErrServerClosed) {
			logger.Error("server closed", "reason", "server closed intentionally")
			return
		}

		// Log the error and terminate the program.
		logger.Error("server failed", "reason", fmt.Sprintf("listening failed: %v", err))
	}
}
