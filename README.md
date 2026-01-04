<p align="center">
  <img src="cmd/server/assets/static/img/icon-192.png" alt="Go DDD Hexagonal Starter logo" width="96" height="96">
</p>

# Go DDD Hexagonal Starter

[![Go Reference](https://pkg.go.dev/badge/go-ddd-hex-starter.svg)](https://pkg.go.dev/go-ddd-hex-starter)
[![Go Report Card](https://goreportcard.com/badge/github.com/andygeiss/go-ddd-hex-starter)](https://goreportcard.com/report/github.com/andygeiss/go-ddd-hex-starter)
[![Release](https://img.shields.io/github/v/release/andygeiss/go-ddd-hex-starter.svg)](https://github.com/andygeiss/go-ddd-hex-starter/releases)

A production-ready Go template demonstrating Domain-Driven Design (DDD) and Hexagonal Architecture (Ports and Adapters) for building maintainable, testable, and cloud-native applications.

---

## Overview / Motivation

**go-ddd-hex-starter** provides a reusable blueprint for building maintainable Go applications with well-defined boundaries. It serves as both a reference implementation and a starting point for developers and AI coding agents.

The template demonstrates:

- **Hexagonal Architecture** with clear separation between domain logic, adapters, and application entry points.
- **Domain-Driven Design** patterns including aggregates, entities, value objects, and domain events.
- **Profile-Guided Optimization (PGO)** for optimized production builds.
- **OIDC Authentication** via Keycloak for secure HTTP endpoints.

---

## Key Features

- ✅ **Hexagonal Architecture** — Domain at the center, adapters depend on domain only
- ✅ **Domain-Driven Design** — Aggregates, entities, value objects, ports, and domain events
- ✅ **Profile-Guided Optimization** — Benchmark-driven PGO for optimized binaries
- ✅ **OIDC Authentication** — Keycloak integration with session management
- ✅ **Event-Driven Messaging** — Pub/sub dispatcher for domain events
- ✅ **Generic CRUD Repositories** — JSON file–backed storage with type-safe access
- ✅ **Structured Logging** — JSON logs via `log/slog`
- ✅ **Docker/Podman Support** — Multi-stage builds, scratch runtime images
- ✅ **Just Task Runner** — Standardized build, test, and deploy workflows
- ✅ **Embedded Assets** — Static files and templates via `embed.FS`

---

## Architecture Overview

The project implements **Hexagonal Architecture** with three distinct layers:

```
┌──────────────────────────────────────────────────────────────┐
│                    Application Layer                         │
│                  cmd/cli/main.go                             │
│                  cmd/server/main.go                          │
│        (Entry points, wire adapters to domain)               │
└─────────────────────────┬────────────────────────────────────┘
                          │
┌─────────────────────────▼────────────────────────────────────┐
│                       Domain Layer                           │
│                   internal/domain/                           │
│    (Pure business logic, defines Ports as interfaces)        │
└─────────────────────────▲────────────────────────────────────┘
                          │
┌─────────────────────────┴────────────────────────────────────┐
│  ┌────────────────┐           ┌──────────────────┐           │
│  │ Inbound        │───────────│ Outbound         │           │
│  │ Adapters       │           │ Adapters         │           │
│  │ (Driving)      │           │ (Driven)         │           │
│  └────────────────┘           └──────────────────┘           │
│                     Adapters Layer                           │
│                 internal/adapters/                           │
└──────────────────────────────────────────────────────────────┘
```

### Layer Responsibilities

| Layer | Location | Responsibility |
|-------|----------|----------------|
| **Domain** | `internal/domain/` | Pure business logic. Defines Ports (interfaces). Zero infrastructure knowledge. |
| **Adapters** | `internal/adapters/` | Infrastructure implementations. Inbound = driving; Outbound = driven. |
| **Application** | `cmd/` | Entry points that wire adapters to domain services via dependency injection. |

### Dependency Rule

Source code dependencies point **inward** only:

- `adapters` depends on `domain`
- `domain` depends on **nothing** (no infrastructure imports)

---

## Project Structure

```
go-ddd-hex-starter/
├── .justfile                 # Task runner commands (build, test, profile, run, serve)
├── go.mod                    # Go module definition (Go 1.25+)
├── CONTEXT.md                # AI/developer context documentation
├── VENDOR.md                 # Vendor library documentation (cloud-native-utils)
├── README.md                 # User-facing documentation (this file)
├── Dockerfile                # Multi-stage container build
├── docker-compose.yml        # Dev stack (Keycloak, Kafka, app)
├── cpuprofile.pprof          # PGO profile data (auto-generated)
├── coverage.pprof            # Test coverage data (auto-generated)
├── bin/                      # Compiled binaries (gitignored)
├── tools/                    # Build/dev scripts (Python helpers)
├── cmd/                      # Application entry points
│   ├── cli/                  # CLI application (file indexing example)
│   │   ├── main.go
│   │   ├── main_test.go      # Benchmarks for PGO
│   │   └── assets/
│   └── server/               # HTTP server application
│       ├── main.go
│       └── assets/
│           ├── static/       # CSS, JS, images
│           └── templates/    # HTML templates
└── internal/
    ├── adapters/             # Infrastructure implementations
    │   ├── inbound/          # Driving adapters (filesystem, HTTP handlers, routing)
    │   └── outbound/         # Driven adapters (persistence, messaging)
    └── domain/               # Pure business logic
        ├── event/            # Domain event interface
        └── indexing/         # Bounded Context: Indexing
            ├── aggregate.go
            ├── entities.go
            ├── value_objects.go
            ├── service.go
            ├── ports_inbound.go
            └── ports_outbound.go
```

### Included Examples

- **CLI tool** — Indexes files in a directory and persists the result to JSON.
- **HTTP server** — OIDC authentication, templating, and session management.

---

## Conventions & Standards

> The coding style in this repository reflects a combination of widely used practices, prior experience, and personal preference, and is influenced by the Go projects on github.com/andygeiss. There is no single "best" project setup; you are encouraged to adapt this structure, evolve your own style, and use this repository as a starting point for your own projects.

### Key Conventions

| Element | Convention | Example |
|---------|------------|---------|
| Files | `snake_case.go` | `file_reader.go`, `http_index.go` |
| Packages | lowercase, singular | `indexing`, `event`, `inbound` |
| Interfaces | Noun describing capability | `FileReader`, `IndexRepository` |
| Constructors | `New<Type>()` | `NewIndex()`, `NewFileReader()` |
| Tests | `Test_<Struct>_<Method>_With_<Condition>_Should_<Result>` | `Test_Index_Hash_With_No_FileInfos_Should_Return_Valid_Hash` |
| HTTP Handlers | `Http<View/Action><Name>` | `HttpViewIndex`, `HttpViewLogin` |
| Middleware | `With<Capability>` | `WithLogging`, `WithSecurityHeaders` |

### Architectural Rules

- Domain code never imports adapter packages.
- All operations accept `context.Context` as the first parameter.
- Dependency injection via constructors in `cmd/*/main.go`; no global state.
- Tests follow Arrange–Act–Assert pattern.

---

## Using This Repository as a Template

### Steps to Create a New Project

1. Clone or copy this template repository.
2. Update module name in `go.mod`.
3. Rename `go-ddd-hex-starter` references throughout.
4. Clear or replace the example `indexing` bounded context.
5. Create your bounded contexts in `internal/domain/`.
6. Implement inbound and outbound adapters in `internal/adapters/`.
7. Wire adapters to domain services in `cmd/<app>/main.go`.
8. Write tests following the established conventions.
9. Run `just profile` to generate a PGO baseline.
10. Build with `just build` for an optimized binary.

### Where to Add New Code

| What to Add | Where It Goes |
|-------------|---------------|
| New bounded context | `internal/domain/<context_name>/` |
| New aggregate or entity | `internal/domain/<context>/aggregate.go` or `entities.go` |
| New domain event | `internal/domain/<context>/value_objects.go` |
| New inbound port | `internal/domain/<context>/ports_inbound.go` |
| New outbound port | `internal/domain/<context>/ports_outbound.go` |
| New driving adapter (HTTP, CLI) | `internal/adapters/inbound/<adapter_name>.go` |
| New driven adapter (DB, queue) | `internal/adapters/outbound/<adapter_name>.go` |
| New application entry point | `cmd/<app_name>/main.go` |

---

## Getting Started

### Prerequisites

- **Go 1.25+** — Required for `b.Loop()` benchmark syntax and latest features.
- **Docker/Podman** — For containerized workflows.
- **Just** — Task runner for standardized commands.

### Installation

```bash
# Clone the repository
git clone https://github.com/andygeiss/go-ddd-hex-starter.git
cd go-ddd-hex-starter

# Install dependencies (macOS/Linux)
just setup
```

---

## Running, Scripts, and Workflows

### Task Runner Commands (`just`)

| Command | Description |
|---------|-------------|
| `just build` | Build optimized binary with PGO to `bin/` |
| `just run` | Build and run the CLI application |
| `just serve` | Run the HTTP server (development mode) |
| `just test` | Run all unit tests with coverage |
| `just profile` | Run benchmarks and generate PGO profile |
| `just up` | Start Docker Compose dev stack (Keycloak, Kafka, app) |
| `just down` | Stop Docker Compose services |
| `just setup` | Install dependencies via Homebrew |

### Aliases

| Alias | Command |
|-------|---------|
| `b` | `build` |
| `u` | `up` |
| `d` | `down` |
| `t` | `test` |

### Manual Commands

```bash
# Run all tests
go test ./...

# Run with verbose output and coverage
go test -v -coverprofile=coverage.pprof ./internal/...

# Run specific test
go test -v -run Test_Index_Hash ./internal/domain/indexing/

# Build without PGO
go build -o ./bin/app ./cmd/cli/main.go

# Build with PGO
go build -ldflags "-s -w" -pgo cpuprofile.pprof -o ./bin/app ./cmd/cli/main.go

# Run HTTP server
PORT=8080 go run ./cmd/server/main.go
```

---

## Usage Examples

### CLI: File Indexing

```bash
# Run the CLI to index files in the current directory
just run

# Or manually:
go run ./cmd/cli/main.go
```

Output:

```
❯ main: index has N files
```

### HTTP Server

```bash
# Start the server
just serve

# Or with a specific port:
PORT=3000 go run ./cmd/server/main.go
```

The server listens on the configured `PORT` (default: `8080`) and provides:

- OIDC-protected routes
- Session management
- HTML templating with HTMX

---

## Testing & Quality

### Running Tests

```bash
# Run all tests with coverage
just test

# Verbose output
go test -v ./internal/...

# Coverage report
go tool cover -html=coverage.pprof
```

### Test Patterns

Tests follow the **Arrange–Act–Assert (AAA)** pattern:

```go
func Test_Index_Hash_With_No_FileInfos_Should_Return_Valid_Hash(t *testing.T) {
    // Arrange
    index := indexing.Index{ID: "empty-index", FileInfos: []indexing.FileInfo{}}

    // Act
    hash := index.Hash()

    // Assert
    assert.That(t, "empty index must have a valid hash (size of 64 bytes)", len(hash), 64)
}
```

### Profile-Guided Optimization

```bash
# Generate PGO profile from benchmarks
just profile

# Build with PGO
just build
```

---

## CI/CD

### Docker Build

The project includes a multi-stage Dockerfile for production builds:

```bash
# Build container image
just build

# Start dev stack with Keycloak and Kafka
just up

# Stop services
just down
```

### Environment Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `PORT` | HTTP server port | `8080` |
| `OIDC_CLIENT_ID` | OIDC client identifier | — |
| `OIDC_CLIENT_SECRET` | OIDC client secret | — |
| `OIDC_ISSUER` | OIDC issuer URL | — |
| `OIDC_REDIRECT_URL` | OIDC callback URL | — |
| `APP_NAME` | Application display name | — |
| `APP_DESCRIPTION` | Application description | — |
| `APP_SHORTNAME` | Short name for Docker image | — |

---

## Limitations and Roadmap

### Current Limitations

- Single bounded context (`indexing`) provided as example.
- OIDC requires external Keycloak instance for authentication.
- JSON file storage is suitable for development; production may require database adapters.

### Potential Extensions

- Database adapters (PostgreSQL, SQLite).
- Additional bounded contexts and domain services.
- Kubernetes deployment manifests.
- OpenTelemetry observability integration.

---

## Technology Stack

| Category | Technology |
|----------|------------|
| **Language** | Go 1.25+ |
| **Architecture** | Hexagonal (Ports and Adapters), Domain-Driven Design |
| **HTTP Framework** | `net/http` (standard library) |
| **Authentication** | OIDC via `cloud-native-utils/security` |
| **Templating** | `cloud-native-utils/templating` with `html/template` |
| **Logging** | `log/slog` + `cloud-native-utils/logging` (JSON structured logs) |
| **Messaging** | `cloud-native-utils/messaging` (pub/sub dispatcher) |
| **Persistence** | `cloud-native-utils/resource` (generic CRUD, JSON file storage) |
| **Testing** | Standard `testing` + `cloud-native-utils/assert` |
| **Build** | `go build` with Profile-Guided Optimization (PGO) |
| **Task Runner** | `just` (justfile) |
| **Containers** | Docker/Podman, multi-stage builds, scratch runtime |
| **External Services** | Keycloak (OIDC provider), Kafka (event streaming) — optional |

**Primary vendor library**: [`github.com/andygeiss/cloud-native-utils`](https://github.com/andygeiss/cloud-native-utils) v0.4.8

---

## Related Documentation

- [CONTEXT.md](CONTEXT.md) — Architecture, conventions, and AI agent context.
- [VENDOR.md](VENDOR.md) — Vendor library documentation for `cloud-native-utils`.
