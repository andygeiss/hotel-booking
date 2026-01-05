set dotenv-load := true
set shell := ["bash", "-cu"]

# ======================================
# Variables
# ======================================
# User running the command (defaults to 'user')

app_user := env("USER", "user")

# Docker image name: {user}/{app}:latest

app_image := app_user + "/" + env("APP_SHORTNAME") + ":latest"

# ======================================
# Aliases - Quick shortcuts
# ======================================

alias b := build
alias u := up
alias d := down
alias t := test

# ======================================
# Build - Create Docker image
# ======================================
# Builds the application container image.
#
# Notes:
# - This template uses `podman build` for image builds.
# - `up`/`down` use `docker-compose` to run the dev stack.
# - If you prefer Docker for builds, replace `podman build` with `docker build`.

build:
    @echo "Building image: {{ app_image }}"
    @podman build \
      -t {{ app_image }} \
      -f Dockerfile .

# ======================================
# Down - Stop Docker Compose services
# ======================================
# Stops and removes all containers defined in docker-compose.yml
# Loads environment from .env file (required for variable interpolation)
# Note: `.env` is expected to be a local file (copy from `.env.example`).

down:
    @docker-compose --env-file .env down

# ======================================
# Fmt - Format Go code
# ======================================
# Formats Go source files using golangci-lint formatters
# Modifies files in place
#
# Requirements:
# - `golangci-lint` must be installed (brew install golangci-lint)

fmt:
    @golangci-lint fmt ./...

# ======================================
# Lint - Run golangci-lint
# ======================================
# Runs golangci-lint to check code quality and style (read-only)
# Uses default configuration or .golangci.yml if present
#
# Requirements:
# - `golangci-lint` must be installed (brew install golangci-lint)

lint:
    @golangci-lint run ./...

# ======================================
# Profile - CPU profiling and benchmarks
# ======================================
# Runs go test benchmarks for each package with CPU profiling
# Generates cpuprofile.pprof (merged) and cpuprofile.svg (visualization)
# Packages profiled: cmd/cli, internal/adapters/inbound, internal/adapters/outbound
#
# Requirements:
# - `go` must be on PATH
# - `go tool pprof -svg` requires Graphviz (`dot`) to be installed
#
# Output:
# - Writes cpuprofile.pprof / cpuprofile.svg into the repo root.
# - These are generated artifacts and are typically ignored by git.

profile:
    @python3 tools/create_pgo.py

# ======================================
# Run - Execute CLI application locally
# ======================================
# Builds the image then runs the CLI binary from cmd/cli/main.go

run:
    @go run ./cmd/cli/main.go

# ======================================
# Serve - Run HTTP server locally
# ======================================
# Starts the HTTP server from cmd/server/main.go
# Server listens on the configured port (see .env or config)

serve:
    @go run ./cmd/server/main.go

# ======================================
# Setup - Install dependencies
# ======================================
# Installs required development tools via Homebrew (macOS/Linux)
# Required: docker-compose (container orchestration), just (command runner),
#           golangci-lint (linting), podman (container runtime)

setup:
    @echo "Installing dependencies via Homebrew..."
    @brew install docker-compose golangci-lint just podman
    @echo "Setup complete! You can now use 'just' commands."

# ======================================
# Up - Start Docker Compose services
# ======================================
# Generates random Keycloak secret, builds image, and starts all services
# Steps:
#   1. Replace CHANGE_ME_LOCAL_SECRET placeholder with random secret
#   2. Build Docker image
#   3. Start all services defined in docker-compose.yml with .env variables
#   4. Wait briefly for Keycloak initialization (best-effort)
#
# Notes:
# - `.env` and `.keycloak.json` are local-only files (copy from *.example).
# - The secret rotation runs before containers start to ensure Keycloak and app match.

up: build
    @python3 tools/change_me_local_secret.py
    @docker-compose --env-file .env up -d
    @echo "Waiting for Keycloak to be ready..."
    @for i in {1..60}; do \
      if podman exec keycloak test -d /opt/keycloak/data/import 2>/dev/null; then \
        echo "✓ Keycloak is ready"; \
        exit 0; \
      fi; \
      sleep 1; \
    done

# ======================================
# Test - Run unit tests with coverage
# ======================================
# Runs all tests in internal/ with coverage profiling
# Outputs coverage percentage and generates coverage.pprof

test:
    @go test -v -coverprofile=coverage.pprof ./internal/...
    @echo "total coverage: $(go tool cover -func=coverage.pprof | grep total | awk '{print $3}')"
