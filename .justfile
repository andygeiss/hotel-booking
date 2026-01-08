set dotenv-load := true
set shell := ["bash", "-cu"]

# ======================================
# Variables
# ======================================
# User running the command (defaults to 'user')

app_user := env("USER", "user")

# Docker image name: {user}/{app}:latest

app_image := app_user + "/" + env("APP_SHORTNAME", "app") + ":latest"

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
# Profile - CPU profiling for PGO
# ======================================
# Runs go test benchmarks with CPU profiling for Profile-Guided Optimization
# Generates cpuprofile.pprof and cpuprofile.svg in the repo root
#
# Requirements:
# - `go` must be on PATH
# - Graphviz (`dot`) must be installed for SVG generation (brew install graphviz)
#
# Usage:
#   just profile              # Run benchmarks and generate profile
#   go build -pgo=cpuprofile.pprof ./cmd/cli  # Build with PGO
#
# Output:
# - cpuprofile.pprof: CPU profile for PGO builds
# - cpuprofile.svg: Visual flame graph of CPU usage

profile:
    @echo "Running benchmarks with CPU profiling..."
    @go test -bench=. -benchtime=10s -cpuprofile=cpuprofile.pprof ./cmd/server/...
    @echo "Generating SVG visualization..."
    @go tool pprof -svg cpuprofile.pprof > cpuprofile.svg
    @echo "Profile written to cpuprofile.pprof"
    @echo "SVG written to cpuprofile.svg"
 
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
    #!/usr/bin/env bash
    set -eu
    echo "Starting local secret rotation..."
    SECRET=$(LC_ALL=C tr -dc 'A-Za-z0-9' < /dev/urandom | head -c 32)
    PLACEHOLDER="CHANGE_ME_LOCAL_SECRET"
    for file in .env .keycloak.json; do
        if [[ -f "$file" ]] && grep -q "$PLACEHOLDER" "$file"; then
            sed -i '' "s/$PLACEHOLDER/$SECRET/g" "$file"
            echo "Updated secret in: $file"
        fi
    done
    echo "Secret rotation complete."
    docker-compose --env-file .env up -d
    echo "Waiting for Keycloak to be ready..."
    for i in {1..60}; do
        if podman exec keycloak test -d /opt/keycloak/data/import 2>/dev/null; then
            echo "✓ Keycloak is ready"
            exit 0
        fi
        sleep 1
    done

# ======================================
# Test - Run unit tests with coverage
# ======================================
# Runs all tests in internal/ with coverage profiling
# Outputs coverage percentage and generates coverage.pprof
# Also runs Python tests for tools/

test:
    @echo "Running Go tests..."
    @go test -v -coverprofile=coverage.pprof ./internal/...
    @echo "total coverage: $(go tool cover -func=coverage.pprof | grep total | awk '{print $3}')"
    @echo ""
    @echo "Running Python tests..."
    @cd tools && python3 -m unittest create_pgo_test -v

# ======================================
# Test Integration - Run integration tests
# ======================================
# Runs integration tests that require external services (e.g., LM Studio)
# These tests are tagged with //go:build integration and are skipped by default
#
# Requirements:
# - LM Studio must be running locally (default: http://localhost:1234)
# - Set LM_STUDIO_URL and LM_STUDIO_MODEL in .env or environment
#
# Usage:
#   just test-integration                    # Run all integration tests
#   just test-integration ./internal/...     # Run integration tests in specific path

test-integration *ARGS='./internal/...':
    @echo "Running integration tests..."
    @echo "LM_STUDIO_URL=${LM_STUDIO_URL:-http://localhost:1234}"
    @echo "LM_STUDIO_MODEL=${LM_STUDIO_MODEL:-default}"
    @echo ""
    @go test -tags=integration -v {{ ARGS }}
