<p align="center">
  <img src="cmd/server/assets/static/img/login.png" alt="Go DDD Hexagonal Starter logo" width="300">
</p>

# Go DDD Hexagonal Starter

A production-ready Go template demonstrating Domain-Driven Design (DDD) and Hexagonal Architecture (Ports and Adapters).

[![Go Reference](https://pkg.go.dev/badge/github.com/andygeiss/go-ddd-hex-starter.svg)](https://pkg.go.dev/github.com/andygeiss/go-ddd-hex-starter)
[![Go Report Card](https://goreportcard.com/badge/github.com/andygeiss/go-ddd-hex-starter)](https://goreportcard.com/report/github.com/andygeiss/go-ddd-hex-starter)
[![License](https://img.shields.io/github/license/andygeiss/go-ddd-hex-starter.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/andygeiss/go-ddd-hex-starter.svg)](https://github.com/andygeiss/go-ddd-hex-starter/releases)

---

## Overview

**go-ddd-hex-starter** provides a reusable blueprint for building maintainable Go applications with well-defined boundaries. It serves as both a reference implementation of the Ports and Adapters pattern and a template for developers and AI coding agents that need a consistent, well-documented structure to extend.

The included examples demonstrate:

- A **CLI tool** that indexes files in a directory and persists the result to JSON.
- An **HTTP server** with OIDC authentication, templating, and session management.

---

## Key Features

- **Hexagonal Architecture** with domain at the center and adapters depending on domain only.
- **Domain-Driven Design** patterns including aggregates, entities, value objects, and domain events.
- **Port Interface Pattern** with inbound (driving) and outbound (driven) ports defined in the domain layer.
- **Dependency Injection** via constructors—no global state or `init` wiring.
- **Profile-Guided Optimization (PGO)** for production builds.
- **OIDC Authentication** via Keycloak integration.
- **Structured JSON Logging** using `log/slog`.
- **Event-Driven Messaging** with pub/sub support.
- **Embedded Assets** via `embed.FS` for portable binaries.
- **Multi-stage Docker Builds** targeting scratch runtime.

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
│   ├── cli/                  # CLI application
│   │   ├── main.go           # Wires adapters, runs file indexing
│   │   ├── main_test.go      # Benchmarks for PGO
│   │   └── assets/           # Embedded assets (embed.FS)
│   └── server/               # HTTP server application
│       ├── main.go           # Wires adapters, starts server with OIDC
│       └── assets/
│           ├── static/       # CSS, JS (base.css, htmx.min.js, theme.css)
│           └── templates/    # HTML templates (index.tmpl, login.tmpl)
└── internal/
    ├── adapters/             # Infrastructure implementations
    │   ├── inbound/          # Driving adapters (filesystem, HTTP handlers, routing)
    │   │   ├── file_reader.go
    │   │   ├── router.go
    │   │   ├── http_index.go
    │   │   ├── http_login.go
    │   │   ├── http_view.go
    │   │   └── middleware.go
    │   └── outbound/         # Driven adapters (persistence, messaging)
    │       ├── file_index_repository.go
    │       └── event_publisher.go
    └── domain/               # Pure business logic
        ├── event/            # Domain event interface
        │   └── event.go
        └── indexing/         # Bounded Context: Indexing
            ├── aggregate.go      # Aggregate Root (Index)
            ├── entities.go       # Entities (FileInfo)
            ├── value_objects.go  # Value Objects (IndexID) + Domain Events
            ├── service.go        # Domain Service (IndexingService)
            ├── ports_inbound.go  # Interfaces for driving adapters
            └── ports_outbound.go # Interfaces for driven adapters
```

---

## Conventions & Standards

> The coding style in this repository reflects a combination of widely used practices, prior experience, and personal preference, and is influenced by the Go projects on github.com/andygeiss. There is no single "best" project setup; you are encouraged to adapt this structure, evolve your own style, and use this repository as a starting point for your own projects.

### Naming Conventions

| Element | Convention | Example |
|---------|------------|---------|
| Files | `snake_case.go` | `file_reader.go`, `http_index.go` |
| Packages | lowercase, singular | `indexing`, `event`, `inbound` |
| Interfaces | Noun describing capability | `FileReader`, `IndexRepository` |
| Constructors | `New<Type>()` | `NewIndex()`, `NewFileReader()` |
| HTTP Handlers | `Http<View/Action><Name>` | `HttpViewIndex`, `HttpViewLogin` |
| Tests | `Test_<Struct>_<Method>_With_<Condition>_Should_<Result>` | See test files |

### Key Rules

- Domain code never imports adapter packages.
- All operations accept `context.Context` as the first parameter.
- Explicit dependency injection in `cmd/*/main.go`.
- Testing follows Arrange–Act–Assert pattern.

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

### Invariants to Preserve

- Hexagonal architecture: domain → adapters → application layering.
- Dependency rule: dependencies point inward only.
- Port interface pattern in domain layer.
- Constructor pattern: `New<Type>()` returning interface types.
- Test naming convention and AAA pattern.
- Context propagation through all layers.

---

## Getting Started

### Prerequisites

- **Go 1.25+** (required for `b.Loop()` in benchmarks)
- **just** task runner
- **Docker/Podman** for containerized workflows
- **Homebrew** (macOS/Linux) for dependency installation

### Installation

```bash
# Clone the repository
git clone https://github.com/andygeiss/go-ddd-hex-starter.git
cd go-ddd-hex-starter

# Install dependencies (macOS/Linux)
just setup

# Copy environment files
cp .env.example .env
cp .keycloak.json.example .keycloak.json
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
# Build and run the CLI
just run

# Or run directly
go run ./cmd/cli/main.go
```

The CLI indexes files in a directory and persists the result to JSON using the outbound file repository adapter.

### HTTP Server

```bash
# Start the development server
just serve

# Or with environment variables
PORT=8080 go run ./cmd/server/main.go
```

The server provides:
- OIDC-protected routes
- HTML templating with HTMX
- Session management
- Security headers middleware

### Docker Compose Stack

```bash
# Start full dev environment (Keycloak, Kafka, app)
just up

# Stop services
just down
```

---

## Testing & Quality

### Running Tests

```bash
# Run all tests with coverage
just test

# Output includes coverage percentage
# Generates coverage.pprof for analysis
```

### Test Conventions

- **Framework**: Standard `testing` package + `cloud-native-utils/assert`
- **Pattern**: Arrange-Act-Assert (AAA)
- **Naming**: `Test_<Struct>_<Method>_With_<Condition>_Should_<Result>`
- **Location**: Same directory as source, suffix `_test.go`

### Example Test

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

---

## CI/CD

The repository includes GitHub-hosted CI configuration. Tests are run automatically on push and pull requests.

---

## Environment Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `APP_NAME` | Application display name | — |
| `APP_DESCRIPTION` | Application description | — |
| `APP_SHORTNAME` | Short name for Docker image | — |
| `PORT` | HTTP server port | `8080` |
| `OIDC_CLIENT_ID` | OIDC client identifier | — |
| `OIDC_CLIENT_SECRET` | OIDC client secret | — |
| `OIDC_ISSUER` | OIDC issuer URL | — |
| `OIDC_REDIRECT_URL` | OIDC callback URL | — |

---

## Limitations and Roadmap

### Current Limitations

- The example uses JSON file storage; production deployments may need database adapters.
- Keycloak configuration is for local development; generate unique secrets for production.
- External services (Kafka, Keycloak) are optional but required for full feature demonstration.

### Designed for Extension

- Add database adapters (PostgreSQL, SQLite) in `internal/adapters/outbound/`.
- Create additional bounded contexts in `internal/domain/`.
- Implement new HTTP handlers in `internal/adapters/inbound/`.
- Add CLI commands in `cmd/cli/` or create new entry points.

---

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
