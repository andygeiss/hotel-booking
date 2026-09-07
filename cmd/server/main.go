package main

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/andygeiss/cloud-native-utils/logging"
	"github.com/andygeiss/cloud-native-utils/mcp"
	"github.com/andygeiss/cloud-native-utils/messaging"
	"github.com/andygeiss/cloud-native-utils/resource"
	"github.com/andygeiss/cloud-native-utils/service"
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

// shutdownTimeout is how long a server gets to finish in-flight requests once
// a stop has been asked for. It is a fresh budget, never the signal context:
// that one is already cancelled by the time shutdown starts, so passing it
// would make Shutdown return at once and kill the requests it is meant to
// drain. Without any deadline, a single stuck request holds the process open
// forever.
const shutdownTimeout = 10 * time.Second

// oidcTimeout bounds every call to the identity provider: the discovery
// document on the first /mcp request, and the key fetches behind each token
// verification after that. It sits under the app listener's 30 s write
// timeout, so the handler answers before the server gives up on it.
const oidcTimeout = 5 * time.Second

// shutdownOnSignal stops srv once ctx is done, giving it shutdownTimeout to
// drain. RegisterOnContextDone waits five seconds first, so the readiness
// probe fails and the load balancer stops sending traffic before the listener
// closes.
func shutdownOnSignal(ctx context.Context, srv *http.Server) {
	service.RegisterOnContextDone(ctx, func() {
		stop, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		_ = srv.Shutdown(stop)
	})
}

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
	cfg Config,
	reservationService *reservation.Service,
	availabilityChecker reservation.AvailabilityChecker,
	paymentService *payment.Service,
) *mcp.Server {
	server := mcp.NewServer(cfg.AppShortname, cfg.AppVersion)

	// Register tools from each bounded context.
	reservation.RegisterTools(server, reservationService, availabilityChecker)
	payment.RegisterTools(server, paymentService)

	return server
}

func main() {
	// Configuration first: a bad port or a missing credential must surface as
	// one line and exit 2, before anything opens a database or binds a socket.
	cfg, err := parseConfig(os.Args[1:], os.Stderr)
	switch {
	case errors.Is(err, flag.ErrHelp):
		return // -h: the usage text is the answer
	case errors.Is(err, errUsage):
		os.Exit(2) // the FlagSet already printed what was wrong
	case err != nil:
		fmt.Fprintf(os.Stderr, "server: %v\n", err)
		os.Exit(2)
	}

	// Create a new context with a cancel function.
	ctx, cancel := service.Context()
	defer cancel()

	// Create a new logger.
	// We use the logging.NewJsonLogger function from the cloud-native-utils/logging package.
	logger := logging.NewJsonLogger()
	logger.Info("configuration loaded", slog.Any("config", cfg))
	for _, db := range []struct {
		name string
		cfg  DatabaseConfig
	}{{"reservation", cfg.ReservationDB}, {"payment", cfg.PaymentDB}} {
		if db.cfg.Password == "" {
			logger.Warn("database password is empty",
				"database", db.name,
				"fix", "put one file per secret in $CREDENTIALS_DIRECTORY, or set the matching _PASSWORD variable for local development")
		}
	}

	// Initialize Reservation Database connection.
	reservationDB, err := sql.Open("pgx", cfg.ReservationDB.DSN())
	if err != nil {
		logger.Error("failed to connect to reservation database", "error", err)
		os.Exit(1)
	}
	defer reservationDB.Close()

	// Initialize Payment Database connection.
	paymentDB, err := sql.Open("pgx", cfg.PaymentDB.DSN())
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

	// How the MCP endpoint verifies bearer tokens. The provider is built on the
	// first /mcp request, not here: fetching the issuer's discovery document is
	// a call to somebody else's process, and boot must depend on local facts
	// only. Doing it here meant Keycloak being down stopped this app starting.
	//
	// ctx is the process context rather than a request's, because go-oidc keeps
	// it for the key set it refetches on every verification. oidcClient bounds
	// both that discovery call and every later key fetch; it is injected
	// because http.DefaultClient has no timeout at all.
	oidcClient := &http.Client{Timeout: oidcTimeout}
	verifierCtx := oidc.ClientContext(ctx, oidcClient)
	newVerifier := func() (*oidc.IDTokenVerifier, error) {
		provider, err := oidc.NewProvider(verifierCtx, cfg.OIDCIssuer)
		if err != nil {
			return nil, fmt.Errorf("discovering OIDC issuer %q: %w", cfg.OIDCIssuer, err)
		}
		// A separate client ID: machine-to-machine, not the browser session.
		return provider.Verifier(&oidc.Config{ClientID: cfg.MCPClientID}), nil
	}

	// Build the MCP server with all tools registered.
	mcpServer := buildMCPServer(cfg, reservationService, availabilityChecker, paymentService)

	// Create router with all dependencies via RouterConfig.
	mux := inbound.Route(inbound.RouterConfig{
		App: inbound.AppInfo{
			Description: cfg.AppDescription,
			Name:        cfg.AppName,
			Version:     cfg.AppVersion,
		},
		Ctx:                ctx,
		EFS:                efs,
		Logger:             logger,
		ReservationService: reservationService,
		MCPServer:          mcpServer,
		NewVerifier:        newVerifier,
	})

	// Start the ops listener. /healthz and the pprof endpoints answer here and
	// nowhere else: the address is loopback, so the proxy cannot reach them and
	// being unreachable is their whole access control.
	opsSrv := &http.Server{
		Addr: "127.0.0.1:6060",
		Handler: inbound.OpsHandler(cfg.AppVersion, func(ctx context.Context) error {
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

	// How this goroutine stops: the shutdown closes the listener when the
	// context is done, and ListenAndServe then returns http.ErrServerClosed.
	shutdownOnSignal(ctx, opsSrv)
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

	// The application listener. Built here rather than by a helper so that
	// cfg.Host and cfg.Port actually decide the address, and so the four
	// timeouts are visible where somebody debugging a hung request will look.
	// An empty host binds every interface, which is what the container needs;
	// a bare-metal deployment behind a proxy sets HOST=127.0.0.1.
	srv := &http.Server{
		Addr:              net.JoinHostPort(cfg.Host, cfg.Port),
		Handler:           mux,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
		IdleTimeout:       2 * time.Minute,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
	}
	defer func() { _ = srv.Close() }()

	shutdownOnSignal(ctx, srv)

	logger.Info("server initialized", "addr", srv.Addr, "version", cfg.AppVersion)

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
