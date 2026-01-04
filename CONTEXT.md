# CONTEXT.md

> **Project context** for AI coding agents, retrieval systems, and developers.  
> This file is the authoritative reference for architecture, conventions, and contracts in this repository.

---

## 1. Project Purpose

**go-ddd-hex-starter** is a production-ready Go template demonstrating Domain-Driven Design (DDD) and Hexagonal Architecture (Ports and Adapters).

It provides:

- A **reusable blueprint** for building maintainable Go applications with well-defined boundaries.
- A **reference implementation** of the Ports and Adapters pattern with working examples.
- A **template** for developers and AI coding agents that need a consistent, well-documented structure to extend.

The included examples demonstrate:

- A **CLI tool** that indexes files in a directory and persists the result to JSON.
- An **HTTP server** with OIDC authentication, templating, and session management.

---

## 2. Technology Stack

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

**Primary vendor library**: `github.com/andygeiss/cloud-native-utils` v0.4.8

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

- `adapters` depends on `domain`
- `domain` depends on **nothing** (no infrastructure imports)

---

## 4. Directory Structure (Contract)

```
go-ddd-hex-starter/
├── .justfile                 # Task runner commands (build, test, profile, run, serve)
├── go.mod                    # Go module definition (Go 1.25+)
├── CONTEXT.md                # AI/developer context documentation (this file)
├── VENDOR.md                 # Vendor library documentation (cloud-native-utils)
├── README.md                 # User-facing documentation
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

### Rules for New Code

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

## 5. Coding Conventions

### 5.1 General

- Keep modules small and focused on a single responsibility.
- Prefer pure functions where possible.
- Respect layer boundaries: domain code must not import infrastructure packages.
- Dependency injection via constructors: wire adapters to domain services in `cmd/*/main.go`.
- Embedded assets via `embed.FS` for portability.

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
| HTTP Handlers | `Http<View/Action><Name>` | `HttpViewIndex`, `HttpViewLogin` |
| Middleware | `With<Capability>` | `WithLogging`, `WithSecurityHeaders` |
| Tests | `Test_<Struct>_<Method>_With_<Condition>_Should_<Result>` | `Test_Index_Hash_With_No_FileInfos_Should_Return_Valid_Hash` |
| Benchmarks | `Benchmark_<Struct>_<Method>_With_<Condition>_Should_<Result>` | `Benchmark_Main_With_Inbound_And_Outbound_Adapters_Should_Run_Efficiently` |

### 5.3 Error Handling & Logging

**Error Handling**:

- Return errors from functions; do not panic except for unrecoverable situations.
- Errors propagate upward through the call stack to the application layer.
- Domain layer returns plain errors; adapters may wrap errors with context.

**Logging**:

- Use `log/slog` with `cloud-native-utils/logging.NewJsonLogger()` for structured JSON logs.
- Log at the adapter/application layer, not in domain code.
- Standard log levels: `Info`, `Error`, `Debug`.
- Include contextual fields: `"method"`, `"path"`, `"duration"`, `"reason"`.

### 5.4 Testing

**Framework**: Standard `testing` package + `cloud-native-utils/assert`.

**Pattern**: Arrange-Act-Assert (AAA).

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

**Test file location**: Same directory as source, suffix `_test.go`.

**Mocking**: Use `resource.NewMockAccess[K, V]()` for repository mocks. Create simple structs implementing interfaces for other mocks.

**Coverage**: Run `just test` or `go test -coverprofile=coverage.pprof ./internal/...`.

### 5.5 Formatting & Linting

- Use `gofmt` or `goimports` for formatting (default Go toolchain).
- No external linter configuration files in this template—relies on Go defaults.
- Consistent import grouping: standard library, external packages, internal packages.

### 5.6 Context Propagation

- Always pass `context.Context` as the first parameter.
- Create contexts in inbound adapters or application layer (`context.Background()`).
- Never create contexts in domain code.
- Respect context cancellation in long-running operations.

---

## 6. Agent-Specific Patterns

This repository does not contain AI agents, but serves as a **template for agent-based projects**. The patterns below apply to extending this template for agent use cases.

### Domain Structure for Agents

If adding agent capabilities:

1. Create a new bounded context: `internal/domain/agents/`
2. Define agent entities in `entities.go` (agent state, configuration).
3. Define inbound ports in `ports_inbound.go` (interfaces for invoking agents).
4. Define outbound ports in `ports_outbound.go` (interfaces for tools, LLM calls, memory).
5. Implement service orchestration in `service.go`.

### Tool Registration Pattern

Tools (external capabilities for agents) follow the adapter pattern:

- Define tool interface in domain layer (`ports_outbound.go`).
- Implement tool adapter in `internal/adapters/outbound/`.
- Wire tool to service via constructor injection in `cmd/*/main.go`.

### Prompt and Workflow Organization

- Store prompt templates in `cmd/<app>/assets/templates/` (embedded via `embed.FS`).
- Implement workflow orchestration in domain services.
- Keep prompts as data (templates), not hardcoded strings in Go code.

### Checklist: Adding a New Agent

1. Create bounded context in `internal/domain/<agent_name>/`.
2. Define `entities.go`, `value_objects.go`, `ports_inbound.go`, `ports_outbound.go`.
3. Implement `service.go` for orchestration.
4. Create adapters in `internal/adapters/inbound/` (trigger) and `outbound/` (tools).
5. Wire in `cmd/<app>/main.go`.
6. Add tests following naming conventions.

### Checklist: Adding a New Tool

1. Define interface in `internal/domain/<context>/ports_outbound.go`.
2. Implement adapter in `internal/adapters/outbound/<tool_name>.go`.
3. Add constructor `New<ToolName>()` returning the interface type.
4. Inject via domain service constructor.
5. Add unit tests with mocks.

---

## 7. Using This Repo as a Template

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

---

## 8. Key Commands & Workflows

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

### Environment Variables

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

## 9. Important Notes & Constraints

### Security

- OIDC credentials (`OIDC_CLIENT_SECRET`) must be kept secret; use environment variables or secret managers.
- The `.keycloak.json` realm file is for local development only; generate unique secrets via `tools/change_me_local_secret.py` before use.
- Security headers are applied via `WithSecurityHeaders` middleware.

### Performance

- Use Profile-Guided Optimization (PGO) for production builds: `just profile` then `just build`.
- The `cpuprofile.pprof` file is generated from benchmarks and improves binary performance.
- Avoid blocking operations in HTTP handlers; use context cancellation.

### Platform Assumptions

- Developed and tested on macOS; compatible with Linux.
- Docker/Podman required for containerized workflows.
- `just` task runner required for standardized commands.
- Go 1.25+ required (uses `b.Loop()` in benchmarks).

### Do Not Modify

- Do not import infrastructure packages in `internal/domain/`.
- Do not break the hexagonal dependency rule.
- Do not commit sensitive credentials to version control.
- Generated files (`cpuprofile.pprof`, `coverage.pprof`, `bin/`) are typically gitignored.

### Deprecated / Experimental

- None currently documented.

---

## 10. How AI Tools and RAG Should Use This File

### Consumption Instructions

1. **Read `CONTEXT.md` first** before making any architectural changes or large refactors.
2. Treat the rules and contracts in this file as **constraints** unless explicitly updated.
3. Use `VENDOR.md` as the reference for `cloud-native-utils` library patterns.
4. Use `README.md` for user-facing documentation and quick-start instructions.

### Priority Order for Context

1. `CONTEXT.md` — Architecture, conventions, contracts.
2. `VENDOR.md` — External library documentation.
3. `README.md` — User-facing overview and commands.
4. Source code — Implementation details.

### Key Constraints for AI Agents

- Respect the hexagonal architecture: domain code must not import adapters.
- Follow naming conventions exactly (test names, file names, constructor patterns).
- Always pass `context.Context` as the first parameter.
- Use `cloud-native-utils` patterns before implementing custom utilities.
- Add tests for any new or changed code following the AAA pattern.
- Maintain the dependency injection pattern via constructors.

### When Generating Code

- Place new domain logic in `internal/domain/<context>/`.
- Place new adapters in `internal/adapters/inbound/` or `outbound/`.
- Wire new components in `cmd/<app>/main.go`.
- Return interfaces from constructors, not concrete types.
- Create corresponding `_test.go` files with proper test naming.
