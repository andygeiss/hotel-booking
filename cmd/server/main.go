package main

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"os"

	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/inbound"

	"github.com/andygeiss/cloud-native-utils/logging"
	"github.com/andygeiss/cloud-native-utils/security"
	"github.com/andygeiss/cloud-native-utils/service"
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

	// Create a new service with the configuration.
	mux := inbound.Route(ctx, efs, logger)
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
