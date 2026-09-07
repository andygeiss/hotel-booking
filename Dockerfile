# ======================================
# Dockerfile - Multi-stage Build
# ======================================
# Production-ready container for the Go DDD/Hexagonal application
# Strategy: Multi-stage build with minimal runtime scratch image
#   - Builder stage: golang:1.27.1-alpine3.23 (toolchain used inside the image)
#   - Runtime stage: scratch (runs minimal binary only)
# Result: ~5-10MB container image with no OS/libc dependencies
#

############################
# Build Stage
############################
# Compiles the Go application with optimizations
FROM golang:1.27.1-alpine3.23 AS builder

# Build environment setup
# CGO_ENABLED=0: Static compilation (no C dependencies)
# GO111MODULE=on: Enable go modules (required for dependency management)
ENV CGO_ENABLED=0 \
    GO111MODULE=on

WORKDIR /app

# Cache go module downloads separately
# Docker reuses this layer if go.mod/go.sum haven't changed
COPY go.mod go.sum ./
RUN go mod download

# Copy remaining source code (invalidates cache only if source changes)
COPY . .

# Build server binary with optimizations
# Flags:
#   -ldflags "-s -w": Strip debug symbols (smaller binary)
#   -trimpath: Reproducible paths, the same flag `make build` uses
#   -o server: Output binary name
#
# Profile-Guided Optimization needs no flag: go build reads cmd/server/default.pgo
# on its own because it sits next to the main package. That file is committed, so
# a clean checkout builds without running the benchmarks first. Refresh it with
# `make profile`.
RUN go build \
    -ldflags "-s -w" \
    -trimpath \
    -o server ./cmd/server

############################
# Runtime Stage
############################
# Minimal production image containing only the compiled binary
# Uses 'scratch' (empty image) because:
#   - Go binary is statically compiled (no libc needed)
#   - Assets (templates, CSS, JS) are embedded in binary
#   - No external dependencies required
# Result: Extremely small, fast, and secure image
FROM scratch

# Copy compiled server binary from builder stage
COPY --from=builder /app/server /server

# Server listens on this port (see cmd/server/main.go for actual binding)
EXPOSE 8080

# Start the server
ENTRYPOINT ["/server"]
