<p align="center">
  <img src="cmd/server/assets/static/img/login.png" alt="Go DDD Hexagonal Starter logo" width="300">
</p>

# Go DDD Hexagonal Starter

A production-ready Go template demonstrating Domain-Driven Design (DDD) and Hexagonal Architecture (Ports & Adapters) patterns for building cloud-native applications.

[![Go Reference](https://pkg.go.dev/badge/github.com/andygeiss/go-ddd-hex-starter.svg)](https://pkg.go.dev/github.com/andygeiss/go-ddd-hex-starter)
[![Go Report Card](https://goreportcard.com/badge/github.com/andygeiss/go-ddd-hex-starter)](https://goreportcard.com/report/github.com/andygeiss/go-ddd-hex-starter)
[![License](https://img.shields.io/github/license/andygeiss/go-ddd-hex-starter.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/andygeiss/go-ddd-hex-starter.svg)](https://github.com/andygeiss/go-ddd-hex-starter/releases)

---

## Overview

This repository serves as a **reference implementation** for AI coding agents and developers to spin up new Go projects following established architectural patterns and conventions. It provides:

- Clean separation of domain logic from infrastructure concerns
- Event-driven architecture with pub/sub messaging
- OIDC authentication via Keycloak
- HTTP server with HTMX-powered UI
- Profile-Guided Optimization (PGO) for performance
- Docker/Podman containerization with multi-stage builds

---

## Key Features

| Feature | Description |
|---------|-------------|
| **Hexagonal Architecture** | Domain at the center with ports and adapters for flexibility and testability |
| **Domain-Driven Design** | Aggregates, entities, value objects, and domain events |
| **Event-Driven** | Internal pub/sub messaging with optional Kafka support |
| **OIDC Authentication** | Secure authentication via Keycloak |
| **HTMX UI** | Modern, hypermedia-driven web interface |
| **PGO Optimization** | Profile-Guided Optimization for production builds |
| **Multi-stage Docker** | Minimal (~5-10MB) scratch-based production images |
| **Just Task Runner** | Simple commands for build, test, and deployment |

---

## Architecture Overview

This template follows **Hexagonal Architecture** (Ports & Adapters) with **Domain-Driven Design** principles:

```
┌─────────────────────────────────────────────────────────────────┐
│                        cmd/ (Entry Points)                      │
│   ┌─────────────┐                       ┌─────────────┐         │
│   │   cli/      │                       │   server/   │         │
│   │  main.go    │                       │   main.go   │         │
│   └──────┬──────┘                       └──────┬──────┘         │
└──────────┼─────────────────────────────────────┼────────────────┘
           │                                     │
           ▼                                     ▼
┌─────────────────────────────────────────────────────────────────┐
│               internal/adapters/ (Infrastructure)               │
│   ┌────────────────────────┐    ┌────────────────────────┐      │
│   │       inbound/         │    │       outbound/        │      │
│   │  - HTTP handlers       │    │  - Repositories        │      │
│   │  - File readers        │    │  - Event publishers    │      │
│   │  - Event subscribers   │    │  - External services   │      │
│   │  - Router, Middleware  │    │                        │      │
│   └───────────┬────────────┘    └────────────┬───────────┘      │
└───────────────┼──────────────────────────────┼──────────────────┘
                │         Ports (Interfaces)   │
                ▼                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    internal/domain/ (Core)                      │
│   ┌────────────────────────┐    ┌────────────────────────┐      │
│   │        event/          │    │       indexing/        │      │
│   │  - Event interface     │    │  - Aggregate (Index)   │      │
│   │  - Publisher port      │    │  - Entities (FileInfo) │      │
│   │  - Subscriber port     │    │  - Value Objects       │      │
│   │  - Factory/Handler     │    │  - Domain Events       │      │
│   │                        │    │  - Ports (interfaces)  │      │
│   │                        │    │  - Service             │      │
│   └────────────────────────┘    └────────────────────────┘      │
└─────────────────────────────────────────────────────────────────┘
```

### Architectural Layers

| Layer | Location | Responsibility |
|-------|----------|----------------|
| **Domain** | `internal/domain/` | Pure business logic, aggregates, entities, value objects, domain events, port interfaces |
| **Adapters** | `internal/adapters/` | Infrastructure implementations (HTTP, filesystem, messaging, persistence) |
| **Application** | `cmd/` | Entry points, dependency wiring, lifecycle management |

### Data Flow

1. **Inbound adapters** receive external requests (HTTP, events, CLI)
2. **Domain services** orchestrate business workflows using domain objects
3. **Outbound adapters** persist state or publish events to external systems
4. **Domain events** decouple bounded contexts via pub/sub

---

## Project Structure

```
.
├── cmd/
│   ├── cli/                 # CLI application entry point
│   │   ├── main.go
│   │   ├── main_test.go     # Benchmarks for PGO
│   │   └── assets/          # Embedded CLI assets
│   └── server/              # HTTP server entry point
│       ├── main.go
│       ├── main_test.go     # Benchmarks for PGO
│       └── assets/          # Embedded web assets (templates, static)
├── internal/
│   ├── adapters/
│   │   ├── inbound/         # Driving adapters (HTTP, filesystem, events)
│   │   │   ├── router.go           # HTTP route registration
│   │   │   ├── middleware.go       # HTTP middleware (logging, security)
│   │   │   ├── http_*.go           # HTTP handlers
│   │   │   ├── file_reader.go      # Filesystem adapter
│   │   │   └── event_subscriber.go # Event subscription adapter
│   │   └── outbound/        # Driven adapters (repositories, publishers)
│   │       ├── file_index_repository.go  # Persistence adapter
│   │       └── event_publisher.go        # Event publishing adapter
│   └── domain/
│       ├── event/           # Cross-cutting event abstractions
│       │   ├── event.go            # Event interface
│       │   ├── event_factory.go    # EventFactoryFn type
│       │   ├── event_handler.go    # EventHandlerFn type
│       │   ├── event_publisher.go  # EventPublisher port
│       │   └── event_subscriber.go # EventSubscriber port
│       └── indexing/        # Example bounded context
│           ├── aggregate.go        # Index aggregate root
│           ├── entities.go         # FileInfo entity
│           ├── value_objects.go    # IndexID value object
│           ├── events.go           # Domain events
│           ├── ports_inbound.go    # FileReader port
│           ├── ports_outbound.go   # IndexRepository port
│           └── service.go          # IndexingService
├── tools/                   # Python utilities for development
│   ├── change_me_local_secret.py   # Keycloak secret rotation
│   └── create_pgo.py               # PGO profile generation
├── .justfile                # Task runner configuration
├── docker-compose.yml       # Development stack (Keycloak, Kafka, App)
├── Dockerfile               # Multi-stage production build
├── go.mod                   # Go module definition
├── CONTEXT.md               # AI agent context documentation
├── README.md                # This file
└── VENDOR.md                # Vendor library documentation
```

---

## Conventions & Standards

### Naming Conventions

| Element | Convention | Example |
|---------|------------|---------|
| **Packages** | Lowercase, single word | `indexing`, `inbound`, `outbound` |
| **Files** | Snake_case | `file_reader.go`, `event_publisher.go` |
| **Aggregates** | PascalCase noun | `Index`, `Order`, `User` |
| **Entities** | PascalCase noun | `FileInfo`, `LineItem` |
| **Value Objects** | PascalCase with `ID` suffix for identifiers | `IndexID`, `OrderID` |
| **Interfaces (ports)** | PascalCase noun describing capability | `FileReader`, `IndexRepository` |
| **Services** | PascalCase with `Service` suffix | `IndexingService` |
| **Events** | PascalCase with `Event` prefix | `EventFileIndexCreated` |
| **HTTP handlers** | `HttpView<Name>` or `Http<Action><Resource>` | `HttpViewIndex`, `HttpViewLogin` |
| **Test functions** | `Test_<Struct>_<Method>_With_<Condition>_Should_<Result>` | `Test_IndexingService_CreateIndex_With_Mockup_Should_Return_Two_Entries` |

### Coding Style Disclaimer

> The coding style in this repository reflects a combination of widely used practices, prior experience, and personal preference, and is influenced by the Go projects on github.com/andygeiss. There is no single "best" project setup; you are encouraged to adapt this structure, evolve your own style, and use this repository as a starting point for your own projects.

---

## Using This Repository as a Template

### What Must Be Preserved

| Element | Reason |
|---------|--------|
| Directory structure (`cmd/`, `internal/adapters/`, `internal/domain/`) | Enforces hexagonal architecture |
| Port/adapter pattern | Maintains testability and flexibility |
| Event-driven patterns | Enables loose coupling |
| `cloud-native-utils` dependency | Provides cross-cutting utilities |
| Testing conventions | Ensures consistent test quality |
| PGO workflow | Maintains performance optimization capability |

### What Should Be Customized

| Element | Customization |
|---------|---------------|
| `internal/domain/indexing/` | Replace with your bounded contexts |
| `cmd/cli/` and `cmd/server/` | Adapt entry points for your use case |
| `.env.example` | Update environment variables for your app |
| `APP_NAME`, `APP_SHORTNAME`, `APP_DESCRIPTION` | Your application identity |
| Templates in `assets/templates/` | Your UI design |
| Keycloak realm configuration | Your OIDC setup |

### Steps to Create a New Project

1. **Clone/copy this template**
   ```bash
   git clone https://github.com/andygeiss/go-ddd-hex-starter my-project
   cd my-project
   rm -rf .git && git init
   ```

2. **Update project metadata**
   - Rename module in `go.mod`
   - Update `APP_NAME`, `APP_SHORTNAME`, `APP_DESCRIPTION` in `.env.example`
   - Update `README.md` with your project description
   - Update `LICENSE` if needed

3. **Configure local development**
   ```bash
   cp .env.example .env
   cp .keycloak.json.example .keycloak.json
   ```

4. **Replace example domain**
   - Remove or rename `internal/domain/indexing/`
   - Create your bounded contexts under `internal/domain/`
   - Implement aggregates, entities, value objects, events, ports, and services

5. **Implement adapters**
   - Create inbound adapters for your entry points
   - Create outbound adapters for your persistence and messaging needs

6. **Wire dependencies**
   - Update `cmd/*/main.go` to inject your services and adapters

7. **Add tests and benchmarks**
   - Write tests following naming conventions
   - Add benchmarks in `cmd/*/main_test.go` for PGO

8. **Run the stack**
   ```bash
   just up
   ```

---

## Getting Started

### Prerequisites

- Go 1.25+
- Docker & Docker Compose
- Podman (for image builds)
- [just](https://github.com/casey/just) command runner
- Python 3 (for tooling scripts)

### Installation

```bash
# Install dependencies via Homebrew (macOS/Linux)
just setup

# Copy environment templates
cp .env.example .env
cp .keycloak.json.example .keycloak.json
```

---

## Running, Scripts, and Workflows

### Key Commands

| Command | Alias | Description |
|---------|-------|-------------|
| `just setup` | | Install dependencies (docker-compose, just) via Homebrew |
| `just build` | `just b` | Build Docker image with PGO optimization |
| `just up` | `just u` | Generate secrets, build image, start all services |
| `just down` | `just d` | Stop and remove all Docker Compose services |
| `just test` | `just t` | Run all tests with coverage |
| `just serve` | | Run HTTP server locally (outside Docker) |
| `just run` | | Run CLI application locally |
| `just profile` | | Generate CPU profiles for PGO |

### Environment Selection

- **Local development (without Docker):** Use `just serve` or `just run`
- **Containerized development:** Use `just up`

### Service URLs (after `just up`)

| Service | URL |
|---------|-----|
| Application | http://localhost:8080 |
| Keycloak Admin | http://localhost:8180/admin (admin:admin) |
| Kafka | localhost:9092 |

---

## Usage Examples

### CLI Application

The CLI demonstrates the event-driven architecture with file indexing:

```bash
just run
```

Output:
```
❯ main: creating index for path: /path/to/project
❯ event: received EventFileIndexCreated - IndexID: /path/to/project, FileCount: 42
❯ main: index created at 2026-01-05T12:00:00Z with 42 files
❯ main: index hash: abc123...
❯ main: first 5 files in index:
  - /path/to/file1.go (1234 bytes)
  - /path/to/file2.go (5678 bytes)
  ...
```

### HTTP Server

Start the server locally:

```bash
just serve
```

Or with Docker Compose (includes Keycloak for authentication):

```bash
just up
```

Access the application at http://localhost:8080.

---

## Testing & Quality

### Running Tests

```bash
# Run all tests with coverage
just test

# Output includes coverage percentage
test coverage: 85.2%
```

### Test Conventions

- **Framework:** Go stdlib `testing` package with `cloud-native-utils/assert`
- **Pattern:** Arrange-Act-Assert (AAA)
- **Naming:** `Test_<Struct>_<Method>_With_<Condition>_Should_<Result>`
- **Location:** `*_test.go` files alongside source files

### Example Test

```go
func Test_IndexingService_CreateIndex_With_Mockup_Should_Return_Two_Entries(t *testing.T) {
    // Arrange
    sut, _ := setupIndexingService()
    path := "testdata/index.json"
    ctx := context.Background()

    // Act
    err := sut.CreateIndex(ctx, path)
    files, err2 := sut.IndexFiles(ctx, path)

    // Assert
    assert.That(t, "err must be nil", err == nil, true)
    assert.That(t, "err2 must be nil", err2 == nil, true)
    assert.That(t, "index must have two entries", len(files) == 2, true)
}
```

### Profile-Guided Optimization (PGO)

Generate CPU profiles for optimized production builds:

```bash
just profile
```

This runs benchmarks and generates `cpuprofile.pprof`, which is used during `just build` for PGO.

---

## CI/CD

The project uses `just` commands for build automation. The Dockerfile implements a multi-stage build:

1. **Builder stage:** Compiles Go with PGO optimization
2. **Runtime stage:** Minimal scratch image (~5-10MB)

```bash
# Build production image
just build

# Deploy with Docker Compose
just up
```

---

## Limitations and Roadmap

### Current Limitations

- Platform assumes macOS or Linux (uses Homebrew for setup)
- Podman required for image builds; Docker Compose for orchestration
- PGO profile file must exist for Docker builds (or remove `-pgo` flag)

### Roadmap

- [ ] Additional bounded context examples
- [ ] External Kafka integration examples
- [ ] Kubernetes deployment manifests
- [ ] GitHub Actions CI/CD workflows

---

## Technology Stack

| Category | Technology |
|----------|------------|
| **Language** | Go 1.25+ |
| **Architecture** | Domain-Driven Design, Hexagonal Architecture |
| **HTTP Framework** | `net/http` (stdlib) with `cloud-native-utils/security` |
| **Frontend** | HTMX, HTML templates |
| **Authentication** | OpenID Connect (OIDC) via Keycloak |
| **Messaging** | Internal pub/sub or Apache Kafka via `cloud-native-utils/messaging` |
| **Build/Task Runner** | `just` (justfile) |
| **Containerization** | Docker/Podman with multi-stage builds |
| **Orchestration** | Docker Compose |
| **Profiling** | Go PGO with benchmark-driven profiles |
| **Vendor Library** | [`cloud-native-utils`](https://github.com/andygeiss/cloud-native-utils) v0.4.8 |

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

Copyright (c) 2025 Andreas Geiß
