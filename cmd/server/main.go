package main

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"os"

	"github.com/andygeiss/cloud-native-utils/logging"
	"github.com/andygeiss/cloud-native-utils/messaging"
	"github.com/andygeiss/cloud-native-utils/security"
	"github.com/andygeiss/cloud-native-utils/service"
	"github.com/andygeiss/hotel-booking/internal/adapters/inbound"
	"github.com/andygeiss/hotel-booking/internal/adapters/outbound"
	"github.com/andygeiss/hotel-booking/internal/domain/orchestration"
	"github.com/andygeiss/hotel-booking/internal/domain/payment"
	"github.com/andygeiss/hotel-booking/internal/domain/reservation"
)

//go:embed assets
var efs embed.FS

func main() {
	// Create a new context with a cancel function.
	ctx, cancel := service.Context()
	defer cancel()

	// Create a new logger.
	// We use the logging.NewJsonLogger function from the cloud-native-utils/logging package.
	logger := logging.NewJsonLogger()

	// Initialize PostgreSQL connection.
	db, err := outbound.NewPostgresConnection()
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Shared event dispatcher using Kafka for distributed event messaging.
	dispatcher := messaging.NewExternalDispatcher()

	// Initialize reservation bounded context.
	reservationRepo := outbound.NewPostgresReservationRepository(db)
	availabilityChecker := outbound.NewRepositoryAvailabilityChecker(reservationRepo)
	reservationPublisher := outbound.NewEventPublisher(dispatcher)
	reservationService := reservation.NewService(reservationRepo, availabilityChecker, reservationPublisher)

	// Initialize payment bounded context.
	paymentRepo := outbound.NewPostgresPaymentRepository(db)
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

	// Create a new service with the configuration.
	mux := inbound.Route(ctx, efs, logger, reservationService)
	srv := security.NewServer(mux)
	defer func() { _ = srv.Close() }()

	// Register the server shutdown function on the context done function.
	// We use the RegisterOnContextDone function from the cloud-native-utils/service package.
	// The server.Shutdown function waits for 5 seconds before shutting down the server.
	service.RegisterOnContextDone(ctx, func() {
		_ = srv.Shutdown(context.Background())
	})

	// The server implementation from the cloud-native-utils/security package uses
	// It uses the PORT environment variable to determine the port to listen on.
	// If the PORT environment variable is not set, it defaults to port 8080.
	logger.Info("server initialized", "port", os.Getenv("PORT"))

	// Start the HTTP server in the main goroutine.
	if err := srv.ListenAndServe(); err != nil {
		// Check if the server was closed intentionally.
		if err == http.ErrServerClosed {
			logger.Error("server closed", "reason", "server closed intentionally")
			return
		}

		// Log the error and terminate the program.
		logger.Error("server failed", "reason", fmt.Sprintf("listening failed: %v", err))
	}
}
