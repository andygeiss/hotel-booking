# CONTEXT.md

> **Authoritative project context** for AI coding agents, retrieval systems, and developers.  
> Read this file first before making major changes or large refactors.

---

## 1. Project Purpose

**go-ddd-hex-starter** is a production-ready Go template demonstrating Domain-Driven Design (DDD) and Hexagonal Architecture (Ports and Adapters). It provides a clean, minimal foundation for Go projects requiring clear separation between business logic and infrastructure.

This repository serves as:

- A **reusable blueprint** for building maintainable, scalable Go applications with well-defined boundaries.
- A **reference implementation** of the Ports and Adapters pattern with working examples.
- A **template** for AI coding agents that need a consistent, well-documented structure to extend.

The included examples demonstrate:

- A **CLI tool** that indexes files in a directory and persists the result to JSON.
- An **HTTP server** with OIDC authentication, templating, and session management.

---

## 2. Technology Stack

| Component | Technology |
|-----------|------------|
| **Language** | Go 1.25+ |
| **Architecture** | Hexagonal (Ports and Adapters) + DDD |
| **Primary Library** | `github.com/andygeiss/cloud-native-utils` v0.4.8 |
| **Build System** | `go build` with Profile-Guided Optimization (PGO) |
| **Task Runner** | `just` (justfile) |
| **Testing** | Go standard `testing` package + `cloud-native-utils/assert` |
| **HTTP Server** | Standard library `net/http` + `cloud-native-utils/security` |
| **Templating** | `cloud-native-utils/templating` with `embed.FS` |
| **Authentication** | OIDC via `cloud-native-utils/security` |

### Key Packages from cloud-native-utils

| Package | Purpose |
|---------|---------|
| `assert` | Test assertions (`assert.That(t, msg, got, want)`) |
| `resource` | Generic CRUD access (`resource.Access[K, V]`, `resource.JsonFileAccess`) |
| `messaging` | Event dispatcher for pub/sub (`messaging.Dispatcher`) |
| `security` | HTTP server, OIDC, sessions, secure defaults |
| `templating` | Template engine with `embed.FS` support |
| `logging` | Structured JSON logger (`logging.NewJsonLogger()`) |
| `service` | Context management, graceful shutdown |

See [VENDOR.md](VENDOR.md) for complete library documentation.

---

## 3. High-Level Architecture

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

- `adapters` → depends on → `domain`
- `domain` → depends on → **nothing** (no infrastructure imports)

---

## 4. Directory Structure (Contract)

```
go-ddd-hex-starter/
├── .github/
│   └── agents/               # GitHub Copilot agent instructions
├── .justfile                 # Task runner commands (build, test, profile, run, serve)
├── go.mod                    # Go module definition (Go 1.25+)
├── README.md                 # Project documentation
├── CONTEXT.md                # This file (AI/developer context)
├── VENDOR.md                 # Vendor library documentation
├── cpuprofile.pprof          # PGO profile data (auto-generated)
├── bin/                      # Compiled binaries (gitignored)
├── cmd/                      # Application entry points
│   ├── cli/                  # CLI application
│   │   ├── main.go           # Wires adapters, runs file indexing
│   │   ├── main_test.go      # Benchmarks for PGO
│   │   └── assets/           # Embedded assets (embed.FS)
│   └── server/               # HTTP server application
│       ├── main.go           # Wires adapters, starts server with OIDC
│       └── assets/           # Embedded assets
│           ├── static/       # CSS, JS files (base.css, htmx.min.js, theme.css)
│           └── templates/    # HTML templates (*.tmpl)
└── internal/
    ├── adapters/             # Infrastructure implementations
    │   ├── inbound/          # Driving adapters
    │   │   ├── file_reader.go      # Filesystem access (implements FileReader port)
    │   │   ├── router.go           # HTTP routing and middleware wiring
    │   │   ├── http_index.go       # Index view handler
    │   │   ├── http_login.go       # Login view handler
    │   │   ├── http_view.go        # Generic view renderer
    │   │   └── middleware.go       # HTTP middleware (logging, security headers)
    │   └── outbound/         # Driven adapters
    │       ├── file_index_repository.go  # JSON file persistence (implements IndexRepository port)
    │       └── event_publisher.go        # Event publishing (implements EventPublisher port)
    └── domain/               # Pure business logic
        ├── event/            # Domain event interface
        │   └── event.go      # Event interface with Topic() method
        └── indexing/         # Bounded Context: Indexing
            ├── aggregate.go      # Aggregate Root (Index with Hash())
            ├── entities.go       # Entities (FileInfo)
            ├── value_objects.go  # Value Objects (IndexID) + Domain Events (EventFileIndexCreated)
            ├── service.go        # Domain Service (IndexingService)
            ├── ports_inbound.go  # Interfaces for driving adapters (FileReader)
            └── ports_outbound.go # Interfaces for driven adapters (IndexRepository, EventPublisher)
```

### Rules for New Code

| What to Add | Where It Goes |
|-------------|---------------|
| New bounded context | `internal/domain/<context_name>/` |
| New aggregate or entity | `internal/domain/<context>/aggregate.go` or `entities.go` |
| New value object | `internal/domain/<context>/value_objects.go` |
| New domain event | `internal/domain/<context>/value_objects.go` (implement `Topic()` method) |
| New inbound port | `internal/domain/<context>/ports_inbound.go` |
| New outbound port | `internal/domain/<context>/ports_outbound.go` |
| New driving adapter (HTTP, CLI, filesystem) | `internal/adapters/inbound/<adapter_name>.go` |
| New driven adapter (DB, queue, external API) | `internal/adapters/outbound/<adapter_name>.go` |
| New application entry point | `cmd/<app_name>/main.go` |
| Tests for any file | `<filename>_test.go` (same directory) |
| Static assets | `cmd/<app>/assets/static/` |
| Templates | `cmd/<app>/assets/templates/` |

---

## 5. Coding Conventions

### 5.1 General

- Prefer small, focused modules with single responsibilities.
- Use pure functions where possible; minimize side effects.
- All dependencies are injected explicitly—no global state.
- Use `context.Context` for cancellation and timeout propagation.
- Use `embed.FS` for bundling static files into binaries.
- Domain code must **never** import adapter packages.

### 5.2 Naming

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
| HTTP Handlers | `HttpView<Name>` | `HttpViewIndex`, `HttpViewLogin` |
| HTTP Response Structs | `HttpView<Name>Response` | `HttpViewIndexResponse` |
| Middleware | `With<Capability>` | `WithLogging`, `WithSecurityHeaders` |

#### Test Naming

```
Test_<Struct>_<Method>_With_<Condition>_Should_<Result>
Benchmark_<Struct>_<Method>_With_<Condition>_Should_<Result>
```

Examples:

- `Test_Index_Hash_With_No_FileInfos_Should_Return_Valid_Hash`
- `Benchmark_FileReader_ReadFileInfos_With_1000_Entries_Should_Be_Fast`

### 5.3 Error Handling & Logging

**Error Handling:**

- Return errors from functions; do not panic except for unrecoverable situations.
- Errors propagate upward through the call stack to the application layer.
- Domain layer returns plain errors; adapters may wrap errors with context.

**Logging:**

- Use `cloud-native-utils/logging.NewJsonLogger()` for structured JSON logs.
- Logging is handled at the application layer (`cmd/`), not within domain or adapters.
- Use `slog.Logger` with structured fields: `logger.Info("message", "key", value)`.
- Use `fmt.Printf` with emoji prefixes for CLI output (e.g., `❯ main: ...`).

### 5.4 Testing

| Aspect | Standard |
|--------|----------|
| Framework | Go standard `testing` package |
| Assertions | `github.com/andygeiss/cloud-native-utils/assert` |
| Pattern | Arrange-Act-Assert (AAA) |
| Location | `*_test.go` files next to source |
| Mocking | Use interfaces + simple mock structs; `resource.NewMockAccess[K, V]()` for repositories |
| Benchmarks | Include for performance-critical code; feeds into PGO |

**Test Template:**

```go
func Test_<Struct>_<Method>_With_<Condition>_Should_<Result>(t *testing.T) {
    // Arrange
    // ... setup

    // Act
    // ... execute

    // Assert
    assert.That(t, "description", got, want)
}
```

### 5.5 Formatting & Linting

- Use `gofmt` or `goimports` for formatting.
- Follow standard Go style guidelines.
- Keep imports organized: stdlib, then external, then internal.
- Use doc comments on all exported types, functions, and methods.

---

## 6. Agent-Specific Patterns

This section documents patterns for building services within this architecture.

### Domain Event Pattern

Events implement the `event.Event` interface defined in `internal/domain/event/event.go`:

```go
type Event interface {
    Topic() string
}
```

Create domain events as structs in `value_objects.go`:

```go
type EventFileIndexCreated struct {
    IndexID   IndexID `json:"index_id"`
    FileCount int     `json:"file_count"`
}

func (e EventFileIndexCreated) Topic() string {
    return "file_index_created"
}
```

Events are published by services via the `EventPublisher` port; serialization happens in the outbound adapter.

### Port Interfaces

**Inbound ports** (what the domain exposes) in `ports_inbound.go`:

```go
type FileReader interface {
    ReadFileInfos(ctx context.Context, path string) ([]FileInfo, error)
}
```

**Outbound ports** (what the domain requires) in `ports_outbound.go`:

```go
type IndexRepository resource.Access[IndexID, Index]

type EventPublisher interface {
    Publish(ctx context.Context, e event.Event) error
}
```

### Repository Pattern

Repositories use the generic `resource.Access[K, V]` interface from cloud-native-utils. Implementations like `resource.JsonFileAccess` are wrapped in outbound adapters:

```go
func NewFileIndexRepository(filename string) indexing.IndexRepository {
    return resource.NewJsonFileAccess[indexing.IndexID, indexing.Index](filename)
}
```

### Service Pattern

Services orchestrate use cases and coordinate between aggregates, repositories, and publishers:

```go
type IndexingService struct {
    fileReader      FileReader      // inbound port
    indexRepository IndexRepository // outbound port
    publisher       EventPublisher  // outbound port
}

func NewIndexingService(fr FileReader, ir IndexRepository, ep EventPublisher) *IndexingService {
    return &IndexingService{fileReader: fr, indexRepository: ir, publisher: ep}
}
```

### HTTP Handler Pattern

HTTP handlers use the `cloud-native-utils` templating engine:

```go
func HttpViewIndex(e *templating.Engine) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        // Authentication check, data gathering...
        data := HttpViewIndexResponse{/* ... */}
        HttpView(e, "index", data)(w, r)
    }
}
```

Routes are defined in `router.go` using `mux.HandleFunc()` with middleware chains.

### Adding a New Bounded Context

1. Create directory: `internal/domain/<context_name>/`
2. Define aggregate root: `aggregate.go`
3. Define entities: `entities.go`
4. Define value objects and domain events: `value_objects.go`
5. Define inbound ports: `ports_inbound.go`
6. Define outbound ports: `ports_outbound.go`
7. Implement domain service: `service.go`
8. Write tests: `*_test.go` for each file
9. Implement adapters in `internal/adapters/`
10. Wire in `cmd/<app>/main.go`

### Adding a New Adapter

**Inbound (driving) adapter:**

1. Create `internal/adapters/inbound/<name>.go`
2. Implement the inbound port interface from domain
3. Constructor returns the domain interface type: `func New<Name>() <domain>.<Interface>`
4. Write tests in `<name>_test.go`

**Outbound (driven) adapter:**

1. Create `internal/adapters/outbound/<name>.go`
2. Implement the outbound port interface from domain
3. Constructor returns the domain interface type
4. Prefer reusing `cloud-native-utils` types (e.g., `resource.JsonFileAccess`)
5. Write tests in `<name>_test.go`

### Adding a New HTTP Endpoint

1. Create handler in `internal/adapters/inbound/http_<name>.go`
2. Define response struct: `HttpView<Name>Response`
3. Implement handler function: `HttpView<Name>(e *templating.Engine) http.HandlerFunc`
4. Create template in `cmd/server/assets/templates/<name>.tmpl`
5. Register route in `router.go` with appropriate middleware

---

## 7. Using This Repo as a Template

### What Must Be Preserved (Invariants)

- Hexagonal architecture: domain → adapters → application layering
- Dependency rule: dependencies point inward only
- Port interface pattern in domain layer
- Constructor pattern: `New<Type>()` returning interface types
- Test naming convention and AAA pattern
- Context propagation through all layers
- Benchmark-driven PGO workflow

### What Is Designed for Customization

- **Domain logic**: Replace/extend the `indexing` bounded context
- **Adapters**: Add database, HTTP, queue, or API adapters
- **Entry points**: Add CLIs, servers, workers under `cmd/`
- **Embedded assets**: Replace contents of `cmd/*/assets/`
- **Configuration**: Add environment variables or config files
- **Templates**: Customize UI in `assets/templates/`

### Steps to Create a New Project

1. Clone or copy this template repository.
2. Update module name in `go.mod`.
3. Rename `go-ddd-hex-starter` references throughout.
4. Clear or replace the example `indexing` bounded context.
5. Create your bounded contexts in `internal/domain/`.
6. Implement inbound and outbound adapters.
7. Wire adapters to domain services in `cmd/<app>/main.go`.
8. Write tests following conventions.
9. Run `just profile` to generate PGO baseline.
10. Build with `just build` for optimized binary.

---

## 8. Key Commands & Workflows

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

### Environment Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `PORT` | HTTP server port | `8080` |
| `OIDC_CLIENT_ID` | OIDC client identifier | — |
| `OIDC_CLIENT_SECRET` | OIDC client secret | — |
| `OIDC_ISSUER` | OIDC issuer URL | — |
| `OIDC_REDIRECT_URL` | OIDC callback URL | — |

---

## 9. Important Notes & Constraints

### Security Considerations

- Use `cloud-native-utils/security` for HTTP servers (secure defaults).
- OIDC authentication is built into the server entry point.
- Security headers applied via `WithSecurityHeaders` middleware:
  - `Referrer-Policy: strict-origin-when-cross-origin`
  - `Cache-Control: no-store`
  - `X-Content-Type-Options: nosniff`
  - `X-Frame-Options: DENY`
- Sessions are managed via `security.ServerSessions`.

### Performance Considerations

- Profile-Guided Optimization (PGO) is enabled via benchmarks.
- Run `just profile` before production builds.
- Benchmarks should cover hot paths (see `*_test.go` files).
- `cpuprofile.pprof` is checked into the repo and used during builds.

### Context Propagation

- Always pass `context.Context` as the first parameter.
- Create contexts in inbound adapters or application layer (`context.Background()`).
- Never create contexts in domain code.
- Respect context cancellation in long-running operations.
- Use `service.Context()` for graceful shutdown with signal handling.

### Embedded Files

- Use `//go:embed assets` directive for bundling assets.
- Embedded files are read-only at runtime.
- Store templates in `cmd/<app>/assets/templates/`.
- Store static files in `cmd/<app>/assets/static/`.

### Current Limitations

- Single bounded context example (`indexing`).
- File-based persistence only (no database adapters included).
- HTTP server uses OIDC—requires provider configuration for authentication.

### Do Not Modify Directly

- `cpuprofile.pprof`: Auto-generated by `just profile`.
- `coverage.pprof`: Test coverage data.
- `bin/`: Build output directory.

---

## 10. How AI Tools and RAG Should Use This File

### Consumption Guidelines

1. **Read `CONTEXT.md` first** before making major changes or refactors.
2. Use in combination with [README.md](README.md) for user-facing documentation.
3. Consult [VENDOR.md](VENDOR.md) for `cloud-native-utils` API details.

### Constraints for AI Agents

- Treat directory structure and naming conventions as **constraints**.
- Always place new code in the correct layer per Section 4.
- Follow test naming patterns exactly.
- Prefer `cloud-native-utils` utilities over reimplementing (see VENDOR.md).
- Do not break the inward dependency rule.
- Always propagate `context.Context` through function calls.
- Do not invent new architectural patterns—reuse existing ones.
- Do not place domain logic in adapters or infrastructure code in domain.

### When to Update This File

Update `CONTEXT.md` when:

- Adding a new bounded context.
- Introducing a new architectural pattern.
- Changing naming conventions.
- Adding significant new dependencies.
- Modifying the directory structure.

Do **not** update for:

- Bug fixes.
- Adding individual adapters following existing patterns.
- Test additions.

---

*Last updated: 2026-01-04*