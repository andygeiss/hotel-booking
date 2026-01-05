# CONTEXT.md

## 1. Project Purpose

This repository is a **Go starter template** demonstrating Domain-Driven Design (DDD) and Hexagonal Architecture (Ports & Adapters) patterns. It provides a production-ready foundation for building cloud-native Go applications with clean separation of concerns, event-driven architecture, and infrastructure-as-code containerization.

The template includes a working example domain (`indexing`) that demonstrates file indexing with event publishing, OIDC authentication via Keycloak, and Kafka-based messaging. It serves as a reference implementation and scaffolding for teams adopting DDD/Hexagonal patterns in Go.

---

## 2. Technology Stack

| Category | Technology |
|----------|------------|
| **Language** | Go 1.25+ (module: `github.com/andygeiss/go-ddd-hex-starter`) |
| **Core Library** | `github.com/andygeiss/cloud-native-utils` (logging, security, messaging, templating, resource access) |
| **Authentication** | OpenID Connect via Keycloak (`github.com/coreos/go-oidc/v3`) |
| **Messaging** | Apache Kafka (`github.com/segmentio/kafka-go`) |
| **Frontend** | Server-side Go templates with HTMX, vanilla CSS |
| **Build System** | `just` (command runner), Podman/Docker |
| **Containerization** | Multi-stage Dockerfile (scratch-based runtime) |
| **Orchestration** | Docker Compose (Keycloak, Kafka, app) |

---

## 3. High-Level Architecture

This project implements **Hexagonal Architecture** (Ports & Adapters) with DDD tactical patterns:

```
┌─────────────────────────────────────────────────────────────────┐
│                        cmd/ (Entry Points)                      │
│         server/main.go (HTTP)    cli/main.go (CLI demo)         │
└───────────────────────────┬─────────────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────────┐
│              internal/adapters/inbound/ (Driving)               │
│    HTTP handlers, event subscribers, file readers               │
│    Implements domain ports; receives external input             │
└───────────────────────────┬─────────────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────────┐
│                   internal/domain/                              │
│    event/       - Event interfaces (Event, Publisher, Subscriber)
│    indexing/    - Example bounded context                       │
│      ├── aggregate.go      (Index aggregate root)               │
│      ├── entities.go       (FileInfo entity)                    │
│      ├── value_objects.go  (IndexID)                            │
│      ├── events.go         (EventFileIndexCreated)              │
│      ├── ports_inbound.go  (FileReader interface)               │
│      ├── ports_outbound.go (IndexRepository interface)          │
│      └── service.go        (IndexingService - use cases)        │
└───────────────────────────┬─────────────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────────┐
│              internal/adapters/outbound/ (Driven)               │
│    Event publisher (Kafka), file-based repository               │
│    Implements domain ports; interacts with external systems     │
└─────────────────────────────────────────────────────────────────┘
```

**Key Architectural Decisions:**

- **Domain layer** contains only pure Go code with no external dependencies (except cloud-native-utils for interface contracts)
- **Ports** are defined as interfaces in the domain layer
- **Adapters** implement ports and live in `internal/adapters/`
- **Services** orchestrate use cases and publish domain events
- **Events** are explicit structs serialized by infrastructure adapters
- **Dependency injection** via constructor functions (no DI framework)

---

## 4. Directory Structure (Contract)

```
go-ddd-hex-starter/
├── cmd/                          # Application entry points
│   ├── cli/                      # CLI demonstration application
│   │   ├── main.go               # CLI entry point with event-driven demo
│   │   └── assets/               # Embedded CLI assets
│   └── server/                   # HTTP server application
│       ├── main.go               # Server entry point with OIDC auth
│       └── assets/               # Embedded server assets
│           ├── static/           # CSS, JS (HTMX), images
│           └── templates/        # Go HTML templates (*.tmpl)
├── internal/                     # Private application code
│   ├── adapters/                 # Hexagonal adapters layer
│   │   ├── inbound/              # Driving adapters (HTTP, events, files)
│   │   │   ├── router.go         # HTTP route definitions
│   │   │   ├── http_*.go         # HTTP handlers
│   │   │   ├── middleware.go     # HTTP middleware
│   │   │   ├── event_subscriber.go # Kafka event subscription
│   │   │   └── file_reader.go    # Filesystem adapter
│   │   └── outbound/             # Driven adapters (persistence, messaging)
│   │       ├── event_publisher.go    # Kafka event publishing
│   │       └── file_index_repository.go # JSON file persistence
│   └── domain/                   # Domain layer (pure business logic)
│       ├── event/                # Event infrastructure interfaces
│       │   ├── event.go          # Event interface
│       │   ├── event_publisher.go
│       │   ├── event_subscriber.go
│       │   ├── event_factory.go
│       │   └── event_handler.go
│       └── indexing/             # Example bounded context
│           ├── aggregate.go      # Aggregate root
│           ├── entities.go       # Domain entities
│           ├── value_objects.go  # Value objects
│           ├── events.go         # Domain events
│           ├── ports_inbound.go  # Inbound port interfaces
│           ├── ports_outbound.go # Outbound port interfaces
│           └── service.go        # Application service
├── tools/                        # Development/build utilities
│   ├── change_me_local_secret.py # Secret rotation for local dev
│   └── create_pgo.py             # Profile-Guided Optimization script
├── bin/                          # Compiled binaries (gitignored)
├── .justfile                     # Command runner recipes
├── .env.example                  # Environment template (commit this)
├── .keycloak.json.example        # Keycloak realm template (commit this)
├── docker-compose.yml            # Local development stack
└── Dockerfile                    # Production container build
```

### Rules for New Code

| What | Where |
|------|-------|
| **New bounded context** | `internal/domain/<context-name>/` with aggregate, entities, value objects, events, ports, service |
| **Inbound adapters** (HTTP handlers, subscribers, readers) | `internal/adapters/inbound/` |
| **Outbound adapters** (repositories, publishers, external APIs) | `internal/adapters/outbound/` |
| **Domain event interfaces** | `internal/domain/event/` (shared across contexts) |
| **Unit tests for domain** | `internal/domain/<context>/*_test.go` (same package or `_test` suffix) |
| **Integration tests** | `cmd/<app>/*_test.go` |
| **HTTP routes** | Register in `internal/adapters/inbound/router.go` |
| **Static assets** | `cmd/server/assets/static/` (embedded via `//go:embed`) |
| **HTML templates** | `cmd/server/assets/templates/*.tmpl` |
| **New CLI tool** | `cmd/<tool-name>/main.go` |

---

## 5. Coding Conventions

### 5.1 General

- Small, focused modules with single responsibilities
- Pure functions in domain layer; side effects isolated to adapters
- Constructor functions (`NewXxx`) for all structs with dependencies
- Accept `context.Context` as the first parameter for cancellation/timeout propagation
- Domain services orchestrate use cases and publish events
- Adapters never call each other directly; communicate through domain ports

### 5.2 Naming

| Element | Convention | Example |
|---------|------------|---------|
| Files | `snake_case.go` | `event_publisher.go`, `http_index.go` |
| Packages | `lowercase` | `indexing`, `inbound`, `outbound` |
| Interfaces | `PascalCase`, noun or verb-noun | `FileReader`, `IndexRepository`, `EventPublisher` |
| Structs | `PascalCase` | `IndexingService`, `FileInfo`, `Index` |
| Constructors | `NewXxx` | `NewIndex()`, `NewFileReader()` |
| Methods | `PascalCase` | `CreateIndex()`, `ReadFileInfos()` |
| Value objects | Type alias or struct | `type IndexID string` |
| Events | `Event<Action>` | `EventFileIndexCreated` |
| Event topics | `<context>.<snake_case_action>` | `indexing.file_index_created` |
| HTTP handlers | `Http<Type><Resource>` | `HttpViewIndex`, `HttpViewLogin` |
| Test functions | `Test_<Struct>_<Method>_With_<Condition>_Should_<Result>` | `Test_Index_Hash_With_No_FileInfos_Should_Return_Valid_Hash` |

### 5.3 Error Handling & Logging

- **Error propagation:** Return errors up the call stack; let the entry point (cmd) decide how to handle
- **Error wrapping:** Use `fmt.Errorf("context: %w", err)` for additional context
- **Logging:** Use structured JSON logging via `cloud-native-utils/logging.NewJsonLogger()`
- **Log levels:**
  - `Info`: Normal operations (server started, request handled)
  - `Error`: Failures requiring attention (server failed, handler error)
- **Context propagation:** Pass `context.Context` to all operations for cancellation and deadlines
- **Panic:** Only in truly unrecoverable initialization scenarios; never in request handlers

### 5.4 Testing

- **Framework:** Standard `testing` package with `cloud-native-utils/assert` for assertions
- **Pattern:** Arrange-Act-Assert (AAA)
- **Naming:** `Test_<Struct>_<Method>_With_<Condition>_Should_<Result>`
- **Test files:** Same directory as production code, `*_test.go` suffix
- **Package:** Use `<package>_test` for black-box testing of exported API
- **Mocking:**
  - Create mock structs implementing domain interfaces
  - Use `cloud-native-utils/resource.NewMockAccess` for repository mocks
- **Integration tests:** Located in `cmd/<app>/*_test.go`, use `httptest.NewServer`
- **Coverage:** Run with `just test`, generates `coverage.pprof`

### 5.5 Formatting & Linting

- **Formatter:** `golangci-lint fmt` (via `just fmt`)
- **Linter:** `golangci-lint` with project-specific configuration (`.golangci.yml`)
- **Workflow:** Run `just fmt` then `just lint` after every code change; resolve all issues until `0 issues`
- **Import organization:** Standard library, blank line, external packages, blank line, internal packages
- **Line length:** No hard limit; prefer readability
- **Comments:** Document exported types and functions; use `//` style
- **Build flags:** Use `-ldflags "-s -w"` for production (strip debug symbols)

---

## 6. Cross-Cutting Concerns and Reusable Patterns

### Security & Authentication

- **OIDC:** Keycloak integration via `cloud-native-utils/security`
- **Session management:** Server-side sessions with `security.NewServeMux()`
- **Auth middleware:** `security.WithAuth()` protects authenticated routes
- **Security headers:** Applied via `WithSecurityHeaders()` middleware
- **Secrets:** Never commit secrets; use `.env` (local) and environment variables (production)
- **Secret rotation:** `tools/change_me_local_secret.py` generates per-machine secrets

### Logging & Observability

- **Structured logging:** JSON format via `logging.NewJsonLogger()`
- **Request logging:** `logging.WithLogging(logger, handler)` middleware
- **Health checks:** `/liveness` and `/readiness` endpoints (via cloud-native-utils)

### Configuration

- **Environment variables:** All configuration via environment (12-factor)
- **Template:** `.env.example` committed; `.env` is local-only (gitignored)
- **Required variables:** See `.env.example` for full list with documentation
- **Validation:** Read at startup; fail fast on missing required config

### Messaging & Events

- **Dispatcher:** `messaging.NewExternalDispatcher()` for Kafka
- **Publishing:** Domain services publish via `event.EventPublisher` port
- **Subscribing:** `event.EventSubscriber` with factory and handler functions
- **Serialization:** JSON encoding/decoding in adapter layer
- **Topics:** Named as `<context>.<action>` (e.g., `indexing.file_index_created`)

### Persistence

- **Repository pattern:** Domain defines `IndexRepository` interface (extends `resource.Access`)
- **File-based:** `resource.NewJsonFileAccess` for simple JSON persistence
- **Database:** Implement domain repository interface in `internal/adapters/outbound/`

### HTTP Patterns

- **Router:** Standard `http.ServeMux` with pattern matching
- **Handlers:** Return `http.HandlerFunc` for composability
- **Middleware:** Wrap handlers with logging, auth, security headers
- **Templates:** `cloud-native-utils/templating.Engine` with `embed.FS`
- **Static assets:** Served from `/static/` path via embedded filesystem

### Resilience (via cloud-native-utils)

| Pattern | Environment Variable | Purpose |
|---------|---------------------|---------|
| Circuit breaker | `SERVICE_BREAKER_THRESHOLD` | Fast-fail after N failures |
| Rate limiting | `SERVICE_DEBOUNCE_PER_SEC` | Limit events per second |
| Retry | `SERVICE_RETRY_DELAY`, `SERVICE_RETRY_MAX` | Retry transient failures |
| Timeout | `SERVICE_TIMEOUT` | Limit external call duration |

---

## 7. Using This Repo as a Template

### What Must Be Preserved

- **Directory structure:** `cmd/`, `internal/adapters/`, `internal/domain/` organization
- **Port/adapter separation:** Domain defines interfaces; adapters implement them
- **Event-driven patterns:** Services publish events; subscribers react
- **Dependency injection:** Constructor-based injection without frameworks
- **Context propagation:** Pass `context.Context` through all layers
- **Testing conventions:** AAA pattern, mock interfaces, black-box tests

### What Is Designed to Be Customized

- **Bounded contexts:** Replace/extend `internal/domain/indexing/` with your domains
- **Adapters:** Add new inbound (HTTP, gRPC, CLI) and outbound (DB, APIs) adapters
- **Events:** Define domain-specific events in each bounded context
- **Templates/UI:** Replace `cmd/server/assets/` with your frontend
- **Configuration:** Update `.env.example` with your application's settings

### Steps to Create a New Project

1. **Clone/copy** this repository
2. **Update module path** in `go.mod`:
   ```
   module github.com/your-org/your-project
   ```
3. **Update imports** across all Go files to match new module path
4. **Update application metadata** in `.env.example`:
   - `APP_NAME`, `APP_SHORTNAME`, `APP_DESCRIPTION`
5. **Copy configuration templates:**
   ```bash
   cp .env.example .env
   cp .keycloak.json.example .keycloak.json
   ```
6. **Remove example domain** (`internal/domain/indexing/`) or use as reference
7. **Create your bounded contexts** following the indexing example structure
8. **Implement adapters** for your infrastructure (databases, external APIs)
9. **Update routes** in `internal/adapters/inbound/router.go`
10. **Update tests** to cover your domain logic

---

## 8. Key Commands & Workflows

All commands are defined in `.justfile` and executed via `just`:

| Command | Description |
|---------|-------------|
| `just setup` | Install dependencies (docker-compose, just) via Homebrew |
| `just build` (or `just b`) | Build Docker image using Podman |
| `just up` (or `just u`) | Generate secrets, build image, start all services |
| `just down` (or `just d`) | Stop and remove all containers |
| `just fmt` | Format Go code with golangci-lint formatters |
| `just lint` | Run golangci-lint to check code quality (0 issues required) |
| `just test` (or `just t`) | Run unit tests with coverage |
| `just serve` | Run HTTP server locally (requires Kafka/Keycloak) |
| `just run` | Run CLI demo locally |
| `just profile` | Generate PGO profiles via benchmarks |

### Manual Go Commands

```bash
# Run tests with verbose output
go test -v ./internal/...

# Build binaries locally
go build -o bin/server ./cmd/server
go build -o bin/cli ./cmd/cli

# Run specific benchmarks
go test -bench=. -benchtime=5s ./internal/...
```

### Service URLs (when running `just up`)

| Service | URL |
|---------|-----|
| Application | http://localhost:8080 |
| Keycloak Admin | http://localhost:8180/admin (admin:admin) |
| Kafka Broker | localhost:9092 |

---

## 9. Important Notes & Constraints

### Security Constraints

- **Never commit** `.env` or `.keycloak.json` (local files with secrets)
- **Always commit** `.env.example` and `.keycloak.json.example` (templates)
- **Production secrets:** Use external secret management (Vault, K8s Secrets)
- **OIDC issuer URL:** Must match between app and Keycloak configuration exactly

### Platform Assumptions

- **macOS/Linux:** Developed and tested on Unix-like systems
- **Homebrew:** Used for dependency installation (`just setup`)
- **Docker/Podman:** Required for containerized deployment
- **Go 1.25+:** Uses modern Go features and toolchain

### Build & Performance

- **PGO:** Profile-Guided Optimization enabled; requires `cpuprofile.pprof`
- **Embedded assets:** Static files and templates compiled into binary via `//go:embed`
- **Scratch container:** Production image has no OS (requires static binary)
- **Multi-stage build:** Separates build environment from minimal runtime

### Known Limitations

- **Single OIDC provider:** Template assumes Keycloak; other providers need adapter changes
- **File-based persistence:** Example uses JSON files; production needs proper database
- **Local Kafka:** Development uses single-node Kafka; production needs cluster

### Deprecated/Do Not Touch

- `bin/` directory is gitignored and auto-generated
- Profiling artifacts (`*.pprof`, `*.svg`) are generated and gitignored
- `coverage.pprof` is test output; do not commit

---

## 10. How AI Tools and RAG Should Use This File

### Intended Consumption

- **Primary context:** Read `CONTEXT.md` first before any significant changes
- **Supplementary:** Combine with specific file reads for implementation details
- **Constraint source:** Treat architectural rules as binding unless explicitly updating them

### Instructions for AI Agents

1. **Before major changes:** Always read this file to understand project structure
2. **New code:** Follow directory placement rules and naming conventions exactly
3. **New bounded context:** Use `internal/domain/indexing/` as the reference implementation
4. **Testing:** Create tests following the `Test_<Struct>_<Method>_With_<Condition>_Should_<Result>` pattern
5. **Configuration:** Add new environment variables to `.env.example` with documentation
6. **Events:** Define event structs in domain, serialize in adapters
7. **Dependencies:** Prefer `cloud-native-utils` patterns; document new external libraries

### When Updating This File

Update `CONTEXT.md` when:
- Adding new bounded contexts or major features
- Changing architectural patterns or conventions
- Adding new cross-cutting concerns
- Modifying build/deployment processes
