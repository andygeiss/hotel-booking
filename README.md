# Go DDD Hexagonal Starter

[![Go Reference](https://pkg.go.dev/badge/go-ddd-hex-starter.svg)](https://pkg.go.dev/go-ddd-hex-starter)
[![Go Report Card](https://goreportcard.com/badge/go-ddd-hex-starter)](https://goreportcard.com/report/go-ddd-hex-starter)
[![Release](https://img.shields.io/github/v/release/andygeiss/go-ddd-hex-starter.svg)](https://github.com/andygeiss/go-ddd-hex-starter/releases)

A production-ready Go template demonstrating Domain-Driven Design (DDD) and Hexagonal Architecture (Ports and Adapters) for building maintainable, scalable applications.

---

## Overview

**go-ddd-hex-starter** provides a clean, minimal foundation for Go projects requiring clear separation between business logic and infrastructure. It serves as:

- A **reusable blueprint** for building maintainable Go applications with well-defined boundaries.
- A **reference implementation** of the Ports and Adapters pattern with working examples.
- A **template** for developers and AI coding agents that need a consistent, well-documented structure to extend.

The included examples demonstrate:

- A **CLI tool** that indexes files in a directory and persists the result to JSON.
- An **HTTP server** with OIDC authentication, templating, and session management.

---

## Key Features

- **Hexagonal Architecture**: Clean separation between domain logic and infrastructure through ports and adapters.
- **Domain-Driven Design**: Bounded contexts, aggregates, entities, value objects, and domain events.
- **Profile-Guided Optimization (PGO)**: Benchmark-driven build process for optimized binaries.
- **OIDC Authentication**: Built-in authentication using `cloud-native-utils/security`.
- **Embedded Assets**: Static files and templates bundled into binaries via `embed.FS`.
- **Event-Driven Architecture**: Domain events with pub/sub messaging support.
- **Generic Repository Pattern**: Type-safe CRUD operations using `resource.Access[K, V]`.
- **Structured Logging**: JSON logging via `cloud-native-utils/logging`.
- **Task Runner**: Standardized workflows with `just` (justfile).

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
├── CONTEXT.md                # Detailed AI/developer context documentation
├── VENDOR.md                 # Vendor library documentation (cloud-native-utils)
├── cpuprofile.pprof          # PGO profile data (auto-generated)
├── bin/                      # Compiled binaries (gitignored)
├── cmd/                      # Application entry points
│   ├── cli/                  # CLI application
│   │   ├── main.go           # Wires adapters, runs file indexing
│   │   ├── main_test.go      # Benchmarks for PGO
│   │   └── assets/           # Embedded assets
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

## Conventions and Standards

### Naming Conventions

| Element | Convention | Example |
|---------|------------|---------|
| Files | `snake_case.go` | `file_reader.go`, `http_index.go` |
| Packages | lowercase, singular | `indexing`, `event`, `inbound` |
| Interfaces | Noun describing capability | `FileReader`, `IndexRepository` |
| Structs | PascalCase noun | `Index`, `FileInfo`, `EventPublisher` |
| Methods | PascalCase verb phrase | `ReadFileInfos`, `CreateIndex` |
| Constructors | `New<Type>()` | `NewIndex()`, `NewFileReader()` |
| Value Objects | PascalCase, `ID` suffix for identifiers | `IndexID` |
| Domain Events | `Event<Action>` | `EventFileIndexCreated` |
| HTTP Handlers | `Http<View/Action><Name>` | `HttpViewIndex`, `HttpViewLogin` |
| Middleware | `With<Capability>` | `WithLogging`, `WithSecurityHeaders` |

### Test Naming

```
Test_<Struct>_<Method>_With_<Condition>_Should_<Result>
Benchmark_<Struct>_<Method>_With_<Condition>_Should_<Result>
```

### Error Handling

- Return errors from functions; do not panic except for unrecoverable situations.
- Errors propagate upward through the call stack to the application layer.
- Domain layer returns plain errors; adapters may wrap errors with context.

### Context Propagation

- Always pass `context.Context` as the first parameter.
- Create contexts in inbound adapters or application layer.
- Never create contexts in domain code.
- Respect context cancellation in long-running operations.

### Coding Style Disclaimer

> The coding style in this repository reflects a combination of widely used practices, prior experience, and personal preference, and is influenced by the Go projects on github.com/andygeiss. There is no single "best" project setup; you are encouraged to adapt this structure, evolve your own style, and use this repository as a starting point for your own projects.

---

## Using This Repository as a Template

### Invariants (What Must Be Preserved)

- Hexagonal architecture: domain → adapters → application layering.
- Dependency rule: dependencies point inward only.
- Port interface pattern in domain layer.
- Constructor pattern: `New<Type>()` returning interface types.
- Test naming convention and AAA (Arrange-Act-Assert) pattern.
- Context propagation through all layers.
- Benchmark-driven PGO workflow.

### Designed for Customization

- **Domain logic**: Replace or extend the `indexing` bounded context.
- **Adapters**: Add database, HTTP, queue, or API adapters.
- **Entry points**: Add CLIs, servers, workers under `cmd/`.
- **Embedded assets**: Replace contents of `cmd/*/assets/`.
- **Configuration**: Add environment variables or config files.
- **Templates**: Customize UI in `assets/templates/`.

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
| New value object | `internal/domain/<context>/value_objects.go` |
| New domain event | `internal/domain/<context>/value_objects.go` (with `Topic()` method) |
| New inbound port | `internal/domain/<context>/ports_inbound.go` |
| New outbound port | `internal/domain/<context>/ports_outbound.go` |
| New driving adapter (HTTP, CLI) | `internal/adapters/inbound/<adapter_name>.go` |
| New driven adapter (DB, queue) | `internal/adapters/outbound/<adapter_name>.go` |
| New application entry point | `cmd/<app_name>/main.go` |
| Tests | `<filename>_test.go` (same directory) |
| Static assets | `cmd/<app>/assets/static/` |
| Templates | `cmd/<app>/assets/templates/` |

---

## Getting Started

### Prerequisites

- **Go 1.25+** (see `go.mod`)
- **just** task runner ([installation guide](https://github.com/casey/just))
- **OIDC provider** (e.g., Keycloak) for HTTP server authentication (optional)

### Installation

```bash
# Clone the repository
git clone https://github.com/andygeiss/go-ddd-hex-starter.git
cd go-ddd-hex-starter

# Fetch dependencies
go mod tidy
```

### Configuration

Environment variables for the HTTP server:

| Variable | Purpose | Default |
|----------|---------|---------|
| `PORT` | HTTP server port | `8080` |
| `OIDC_CLIENT_ID` | OIDC client identifier | — |
| `OIDC_CLIENT_SECRET` | OIDC client secret | — |
| `OIDC_ISSUER` | OIDC issuer URL | — |
| `OIDC_REDIRECT_URL` | OIDC callback URL | — |

---

## Running, Scripts, and Workflows

### Task Runner Commands

| Command | Description |
|---------|-------------|
| `just build` | Build optimized binary with PGO to `bin/` |
| `just run` | Build and run the CLI application |
| `just serve` | Run the HTTP server (development mode) |
| `just test` | Run all unit tests with coverage |
| `just profile` | Run benchmarks and generate PGO profile |

### Manual Commands

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific test
go test -v -run Test_Index_Hash ./internal/domain/indexing/

# Build without PGO
go build -o ./bin/app ./cmd/cli/main.go

# Run HTTP server
PORT=8080 go run ./cmd/server/main.go
```

---

## Usage Examples

### CLI Tool

The CLI indexes files in a directory and persists results to JSON:

```bash
# Build and run the CLI
just run

# Or manually
go run ./cmd/cli/main.go
```

### HTTP Server

The HTTP server provides OIDC-protected views:

```bash
# Run the server
just serve

# Or manually with environment variables
PORT=8080 go run ./cmd/server/main.go
```

Access the server at `http://localhost:8080`.

---

## Testing and Quality

### Running Tests

```bash
# Run all tests with coverage
just test

# Run tests manually
go test -v -coverprofile=./coverage.pprof ./internal/...

# View coverage report
go tool cover -func=coverage.pprof
```

### Test Patterns

Tests follow the Arrange-Act-Assert pattern with consistent naming:

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

### Benchmarks and PGO

Generate a PGO profile for optimized builds:

```bash
# Run benchmarks and generate profile
just profile

# Build with PGO
just build
```

The `cpuprofile.pprof` file is used during builds for Profile-Guided Optimization.

---

## Limitations and Roadmap

### Current Limitations

- Single bounded context example (`indexing`).
- File-based persistence only (no database adapters included).
- HTTP server uses OIDC—requires provider configuration for authentication.

### Potential Extensions

- Database adapters (PostgreSQL, SQLite).
- Additional bounded contexts.
- gRPC adapters.
- OpenTelemetry integration.

---

## Additional Documentation

- [CONTEXT.md](CONTEXT.md) — Detailed architectural context for AI agents and developers.
- [VENDOR.md](VENDOR.md) — Documentation for `cloud-native-utils` and external patterns.
