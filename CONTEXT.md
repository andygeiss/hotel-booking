# CONTEXT.md

This file is the authoritative project context for AI coding agents, retrieval systems, and developers working on this codebase.

---

## 1. Project Purpose

This repository is a **production-ready Go starter template** demonstrating **Domain-Driven Design (DDD)** and **Hexagonal Architecture (Ports & Adapters)** patterns.

It serves as:
- A reference implementation for structuring Go applications with clean architecture
- A template for spinning up new projects with established conventions
- A demonstration of integrating LLM-based agents with event-driven patterns

The template includes two bounded contexts (`agent` and `indexing`), an HTTP server with OIDC authentication, Kafka-based event streaming, and LM Studio integration for local LLM inference.

---

## 2. Technology Stack

| Category | Technology |
|----------|------------|
| Language | Go 1.25+ |
| Core Library | [`github.com/andygeiss/cloud-native-utils`](https://github.com/andygeiss/cloud-native-utils) |
| Authentication | Keycloak (OIDC), `coreos/go-oidc/v3` |
| Event Streaming | Apache Kafka, `segmentio/kafka-go` |
| LLM Integration | LM Studio (OpenAI-compatible API) |
| Container Runtime | Podman / Docker |
| Orchestration | Docker Compose |
| Task Runner | [`just`](https://github.com/casey/just) |
| Linting | `golangci-lint` |
| CI | GitHub Actions |
| Profiling | PGO (Profile-Guided Optimization) |
| PWA | Service Worker, Web App Manifest |

---

## 3. High-Level Architecture

### Architectural Style

**Hexagonal Architecture (Ports & Adapters)** with **Domain-Driven Design** tactical patterns.

```
┌─────────────────────────────────────────────────────────────┐
│                    Entry Points (cmd/)                      │
│                   cli/main.go, server/main.go               │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                  Inbound Adapters                           │
│   HTTP handlers, file readers, event subscribers            │
│              internal/adapters/inbound/                     │
└──────────────────────────┬──────────────────────────────────┘
                           │ implements ports
┌──────────────────────────▼──────────────────────────────────┐
│                     Domain Layer                            │
│   Bounded contexts: agent/, indexing/, event/               │
│   Aggregates, entities, value objects, services, ports      │
│                   internal/domain/                          │
└──────────────────────────┬──────────────────────────────────┘
                           │ defines ports
┌──────────────────────────▼──────────────────────────────────┐
│                  Outbound Adapters                          │
│   Event publisher, repositories, LLM client                 │
│              internal/adapters/outbound/                    │
└─────────────────────────────────────────────────────────────┘
```

### Bounded Contexts

| Context | Purpose | Location |
|---------|---------|----------|
| `agent` | LLM-based agent with observe→decide→act→update loop, tool execution | `internal/domain/agent/` |
| `indexing` | File indexing, search, and repository management | `internal/domain/indexing/` |
| `event` | Domain event contracts and infrastructure | `internal/domain/event/` |

### Key Patterns

- **Ports (interfaces)** are defined in domain packages (`ports_inbound.go`, `ports_outbound.go`)
- **Adapters (implementations)** live in `internal/adapters/{inbound,outbound}/`
- **Services** orchestrate use cases and coordinate between ports
- **Aggregates** enforce consistency and business rules
- **Domain Events** enable loose coupling between bounded contexts

---

## 4. Directory Structure (Contract)

```
go-ddd-hex-starter/
├── cmd/                          # Application entry points
│   ├── cli/                      # CLI application (demonstrates agent loop)
│   │   ├── main.go
│   │   └── assets/               # Embedded assets for CLI
│   └── server/                   # HTTP server (OIDC-protected web UI)
│       ├── main.go
│       └── assets/               # Embedded static files & templates
│           ├── static/           # CSS, JS, images (served at /static)
│           └── templates/        # Go templates (*.tmpl)
│               ├── index.tmpl    # Main authenticated view
│               ├── login.tmpl    # Login page
│               ├── manifest.tmpl # PWA web app manifest
│               └── sw.tmpl       # PWA service worker
│
├── internal/                     # Private application code
│   ├── adapters/                 # Hexagonal adapters
│   │   ├── inbound/              # Driving adapters (HTTP, filesystem, events)
│   │   │   ├── router.go         # HTTP route definitions
│   │   │   ├── http_*.go         # HTTP handlers
│   │   │   │   ├── http_index.go        # Authenticated index view
│   │   │   │   ├── http_login.go        # Login view
│   │   │   │   ├── http_manifest.go     # PWA manifest.json endpoint
│   │   │   │   ├── http_service_worker.go # PWA service worker endpoint
│   │   │   │   └── http_view.go         # Generic template view handler
│   │   │   ├── file_reader.go    # Filesystem adapter
│   │   │   ├── event_subscriber.go
│   │   │   ├── middleware.go
│   │   │   └── testdata/         # Test fixtures
│   │   └── outbound/             # Driven adapters (repos, publishers, clients)
│   │       ├── event_publisher.go
│   │       ├── file_index_repository.go
│   │       ├── index_search_tool_executor.go  # Agent tool executor for file search
│   │       └── lmstudio_client.go
│   │
│   └── domain/                   # Domain layer (business logic)
│       ├── agent/                # Agent bounded context
│       │   ├── aggregate.go      # Agent aggregate root
│       │   ├── entities.go       # Task, ToolCall, SearchIndexToolArgs, SearchResult
│       │   ├── events.go         # Domain events
│       │   ├── ports_outbound.go # LLMClient, ToolExecutor interfaces
│       │   ├── service.go        # TaskService (orchestrates agent loop)
│       │   └── value_objects.go  # IDs, statuses, Message, Role
│       ├── event/                # Shared event infrastructure
│       │   ├── event.go          # Event interface
│       │   ├── event_publisher.go
│       │   ├── event_subscriber.go
│       │   ├── event_factory.go
│       │   └── event_handler.go
│       └── indexing/             # Indexing bounded context
│           ├── aggregate.go      # Index aggregate root (with Search capability)
│           ├── entities.go       # FileInfo entity
│           ├── events.go         # EventFileIndexCreated
│           ├── ports_inbound.go  # FileReader interface
│           ├── ports_outbound.go # IndexRepository interface
│           ├── service.go        # IndexingService (CreateIndex, SearchFiles)
│           └── value_objects.go  # IndexID, SearchResult
│
├── tools/                        # Build and dev tooling (Python scripts)
│   ├── change_me_local_secret.py # Generates per-machine OIDC secrets
│   └── create_pgo.py             # Profile-Guided Optimization generation
│
├── bin/                          # Compiled binaries (git-ignored)
├── .github/                      # GitHub config
│   ├── agents/                   # AI agent definitions
│   └── workflows/ci.yml          # CI pipeline
│
├── .justfile                     # Task runner commands
├── .golangci.yml                 # Linter configuration
├── .env.example                  # Environment template
├── .keycloak.json.example        # Keycloak realm template
├── docker-compose.yml            # Dev stack (Keycloak, Kafka, app)
├── Dockerfile                    # Multi-stage production build
└── go.mod                        # Go module definition
```

### Rules for New Code

| What | Where | Notes |
|------|-------|-------|
| New bounded context | `internal/domain/{context}/` | Include aggregate, entities, value objects, service, ports, events |
| Inbound adapter | `internal/adapters/inbound/` | Implement domain port interfaces; prefix HTTP handlers with `http_` |
| Outbound adapter | `internal/adapters/outbound/` | Implement domain port interfaces |
| Domain ports | `internal/domain/{context}/ports_*.go` | `ports_inbound.go` for driving ports, `ports_outbound.go` for driven ports |
| Domain events | `internal/domain/{context}/events.go` | Use builder pattern with `With*` methods |
| Test fixtures | `internal/adapters/inbound/testdata/` | Or colocated `*_test.go` files |
| Entry point | `cmd/{app}/main.go` | Wire adapters and start services |
| Embedded assets | `cmd/{app}/assets/` | Use `//go:embed assets` directive |

---

## 5. Coding Conventions

### 5.1 General

- Keep modules small and focused; one responsibility per file
- Prefer pure functions where possible
- Domain layer has **zero external dependencies** (only stdlib + value types)
- Adapters depend on domain; domain never depends on adapters
- Use `context.Context` for cancellation and timeouts across all layers
- Services do not create contexts; they receive and forward them

### 5.2 Naming

| Element | Convention | Example |
|---------|------------|---------|
| Files | `snake_case.go` | `event_publisher.go` |
| Packages | lowercase, no underscores | `indexing`, `outbound` |
| Interfaces | Descriptive noun | `FileReader`, `LLMClient` |
| Structs | PascalCase | `EventPublisher`, `TaskService` |
| Methods | PascalCase | `CreateIndex`, `RunTask` |
| Variables | camelCase | `fileInfos`, `serverSessions` |
| Constants | PascalCase | `TaskStatusPending`, `RoleUser` |
| Test files | `*_test.go` | `aggregate_test.go` |
| HTTP handlers | `Http*` or `HttpView*` | `HttpViewIndex`, `HttpViewLogin` |
| Domain events | `Event*` prefix | `EventTaskStarted`, `EventFileIndexCreated` |
| Event topics | `{context}.{action}` | `agent.task_started`, `indexing.file_index_created` |

### 5.3 Error Handling & Logging

**Error Handling:**
- Return errors; do not panic (except truly unrecoverable situations)
- Wrap errors with context using `fmt.Errorf("...: %w", err)`
- Domain layer returns domain-specific errors; adapters translate them
- HTTP handlers return appropriate status codes (400/401/404/500)

**Logging:**
- Use `log/slog` via `cloud-native-utils/logging`
- Create logger with `logging.NewJsonLogger()` at entry point
- Pass logger to components that need it (middleware, services)
- Use structured logging: `logger.Info("message", "key", value)`
- Log levels: `Info` for operations, `Error` for failures, `Debug` for tracing

### 5.4 Testing

**Framework:** Standard `testing` package

**Organization:**
- Unit tests colocated with source: `aggregate.go` → `aggregate_test.go`
- Integration tests tagged: `//go:build integration`
- Test fixtures in `testdata/` directories

**Expectations:**
- All public functions must have tests
- Use table-driven tests for multiple scenarios
- Mock external dependencies via interfaces
- Integration tests require `just test-integration` (external services)

**Running tests:**
```bash
just test              # Unit tests + coverage
just test-integration  # Integration tests (requires LM Studio, etc.)
```

### 5.5 Formatting & Linting

**Tools:**
- `golangci-lint` for linting and formatting
- `golangci-lint fmt ./...` to format
- `golangci-lint run ./...` to lint

**Key rules (from `.golangci.yml`):**
- `interface{}` → `any` (auto-rewritten)
- `a[b:len(a)]` → `a[b:]` (auto-rewritten)
- Disabled linters: `exhaustruct`, `ireturn`, `mnd`, `wrapcheck` (see config for rationale)

**CI enforcement:**
- GitHub Actions runs `just test` on every push
- Coverage uploaded to Codacy

---

## 6. Cross-Cutting Concerns and Reusable Patterns

### Vendor Library: `cloud-native-utils`

The primary external dependency. Use its utilities instead of rolling custom implementations:

| Concern | Package | Usage |
|---------|---------|-------|
| HTTP server & routing | `security.NewServer`, `security.NewServeMux` | Includes session management, OIDC |
| Templating | `templating.NewEngine` | HTML templates with `fs.FS` support |
| Logging middleware | `logging.WithLogging` | Request logging with duration |
| Messaging/Events | `messaging.Dispatcher`, `messaging.NewKafkaDispatcher` | Pub/sub abstraction |
| Resilience | `service.Wrap` | Circuit breaker, retry, debounce |
| JSON file access | `resource.JsonFileAccess` | Simple file-based repository |
| Graceful shutdown | `service.Context`, `service.RegisterOnContextDone` | Signal handling |

### Security

- OIDC authentication via Keycloak
- Session-based auth with `security.WithAuth` middleware
- Security headers via `WithSecurityHeaders` middleware
- Secrets managed via environment variables (never committed)
- `CHANGE_ME_LOCAL_SECRET` placeholder replaced at `just up` time

### Configuration

- Environment variables (see `.env.example`)
- Loaded via `dotenv-load` in `.justfile`
- Docker Compose uses `--env-file .env`
- Key variables: `PORT`, `KAFKA_BROKERS`, `OIDC_*`, `LM_STUDIO_URL`, `LM_STUDIO_MODEL`, `VERBOSE`, `APP_VERSION`

### Dependency Injection

- Constructor injection: `NewService(dependency1, dependency2)`
- Interfaces defined in domain; implementations in adapters
- Wiring happens in `cmd/*/main.go`

### Event-Driven Communication

- Domain events are plain Go structs implementing `event.Event`
- Events serialized to JSON for Kafka
- Builder pattern: `NewEventTaskStarted().WithAgentID(id).WithTaskID(taskID)`
- Topics follow pattern: `{bounded_context}.{event_name}`

### Progressive Web App (PWA)

- **Manifest:** `/manifest.json` served by `HttpViewManifest` (uses `APP_NAME` env var)
- **Service Worker:** `/sw.js` served by `HttpViewServiceWorker` (uses `APP_NAME`, `APP_VERSION` env vars)
- **Cache Strategy:** Service worker uses versioned cache (`{APP_NAME}-v{APP_VERSION}`) for offline support
- **Installability:** Meta tags in templates enable "Add to Home Screen" on mobile devices
- **Environment Variables:** Handlers read env vars at startup (not per-request) for efficiency

### HTTP/API Patterns

- Standard library `net/http` with `http.ServeMux`
- Handler functions: `func(w http.ResponseWriter, r *http.Request)`
- Middleware chains: `logging.WithLogging(logger, security.WithAuth(sessions, handler))`
- Static assets embedded via `//go:embed` and served at `/static`
- Templates rendered via `templating.Engine.View()`

### Agent Tool Execution Pattern

The agent can invoke tools during its observe→decide→act loop. Tool execution follows this pattern:

```
User Task → Agent Loop:
  1. OBSERVE: Gather messages + tool definitions
  2. DECIDE: LLM returns response (text or tool_calls)
  3. ACT: If tool_calls → execute via ToolExecutor → add results to messages
  4. UPDATE: Loop until task complete or max iterations
```

**Key interfaces:**

```go
// LLMClient sends messages and tool definitions to the LLM
type LLMClient interface {
    Run(ctx context.Context, messages []Message, tools []ToolDefinition) (LLMResponse, error)
}

// ToolExecutor executes tools requested by the LLM
type ToolExecutor interface {
    Execute(ctx context.Context, toolName string, arguments string) (string, error)
    GetAvailableTools() []string
    GetToolDefinitions() []ToolDefinition  // Returns OpenAI-compatible tool schemas
    HasTool(toolName string) bool
}
```

**Tool definitions use OpenAI-compatible JSON Schema:**
```json
{
  "type": "function",
  "function": {
    "name": "search_index",
    "description": "Search indexed files",
    "parameters": {
      "type": "object",
      "properties": {
        "query": {"type": "string", "description": "Search query"}
      },
      "required": ["query"],
      "additionalProperties": false
    }
  }
}
```

**Key components:**
- **`LLMClient` interface** (`agent/ports_outbound.go`): Sends messages + tools to LLM
- **`ToolExecutor` interface** (`agent/ports_outbound.go`): Defines contract for tool execution
- **`LMStudioClient`** (`adapters/outbound/lmstudio_client.go`): OpenAI-compatible LLM adapter
- **`IndexSearchToolExecutor`** (`adapters/outbound/`): Implements `search_index` tool for file search
- **`ToolDefinition`** (`agent/ports_outbound.go`): Domain representation of tool schema

**Adding new tools:**
1. Add tool function to executor's `initTools()` method
2. Add `ToolDefinition` to `GetToolDefinitions()` return value
3. Implement the tool logic (receives JSON arguments string, returns JSON result string)
4. Wire executor in entry point (`cmd/cli/main.go`)

---

## 7. Using This Repo as a Template

### Invariants (Must Preserve)

- Directory structure: `cmd/`, `internal/adapters/`, `internal/domain/`
- Hexagonal architecture: ports in domain, adapters separate
- `cloud-native-utils` as the core infrastructure library
- `context.Context` threading through all layers
- Domain event pattern for cross-context communication
- Testing colocated with source files

### Customization Points

- Add new bounded contexts under `internal/domain/`
- Add new entry points under `cmd/`
- Add new adapters under `internal/adapters/{inbound,outbound}/`
- Modify `.env.example` for project-specific configuration
- Update `APP_NAME`, `APP_SHORTNAME`, `APP_DESCRIPTION`, `APP_VERSION` in `.env`
- Replace/extend `assets/` with project-specific static files and templates
- Customize PWA icons by replacing files in `assets/static/img/`

### Steps to Create a New Project

1. **Clone the template:**
   ```bash
   git clone https://github.com/andygeiss/go-ddd-hex-starter my-project
   cd my-project
   rm -rf .git && git init
   ```

2. **Update module path:**
   ```bash
   # In go.mod, replace module path
   go mod edit -module github.com/yourorg/my-project
   # Find and replace import paths in all .go files
   ```

3. **Configure project identity:**
   ```bash
   cp .env.example .env
   cp .keycloak.json.example .keycloak.json
   # Edit .env: APP_NAME, APP_SHORTNAME, APP_DESCRIPTION
   ```

4. **Add domain logic:**
   - Create new bounded context: `internal/domain/mycontext/`
   - Define aggregates, entities, value objects, services, ports, events
   - Implement adapters in `internal/adapters/`

5. **Wire up entry point:**
   - Create or modify `cmd/myapp/main.go`
   - Inject dependencies and start services

6. **Verify:**
   ```bash
   just setup    # Install dev tools
   just test     # Run tests
   just serve    # Run locally
   just up       # Run with Docker Compose stack
   ```

---

## 8. Key Commands & Workflows

| Command | Description |
|---------|-------------|
| `just setup` | Install dev dependencies (brew: docker-compose, golangci-lint, just, podman) |
| `just build` | Build Docker image with Podman |
| `just up` | Generate secrets, build image, start Keycloak + Kafka + app |
| `just down` | Stop and remove all Docker Compose services |
| `just serve` | Run HTTP server locally (requires KAFKA_BROKERS=localhost:9092) |
| `just run` | Run CLI application locally |
| `just test` | Run unit tests with coverage (Go + Python) |
| `just test-integration` | Run integration tests (requires external services) |
| `just lint` | Run golangci-lint |
| `just fmt` | Format code with golangci-lint formatters |
| `just profile` | Generate CPU profile for PGO |

### Environment Selection

- **Local development:** `just serve` or `just run` (uses `.env` with `localhost` addresses)
- **Docker stack:** `just up` (services communicate via Docker network)
- **Integration tests:** Set `LM_STUDIO_URL`, `LM_STUDIO_MODEL` in `.env`, run `just test-integration`
- **Verbose tool logging:** Set `VERBOSE=true` to see tool calls and results in CLI output

---

## 9. Important Notes & Constraints

### Security

- Never commit secrets (`.env`, `.keycloak.json` are git-ignored)
- OIDC secrets are generated per-machine by `tools/change_me_local_secret.py`
- Production deployments must use proper secret management

### Performance

- Profile-Guided Optimization (PGO) enabled in Dockerfile
- Generate profiles with `just profile` before production builds
- If `cpuprofile.pprof` is missing, remove `-pgo` flag from Dockerfile

### Platform Assumptions

- macOS or Linux development (Homebrew for tooling)
- Podman preferred for builds; Docker Compose for orchestration
- LM Studio for local LLM inference (optional, for agent features)

### Deprecated/Experimental Areas

- None currently marked

### Known Limitations

- File-based repository (`JsonFileAccess`) is for demo only; replace with database for production
- `IndexSearchToolExecutor` performs substring matching on file paths; extend for content search
- Index search scores are heuristic (exact filename match > partial match)
- Integration tests require external services (LM Studio, Kafka, Keycloak)
- LM Studio requires models that support OpenAI-compatible tool calling (Qwen3, Llama 3.1+, Mistral, Hermes-2-Pro)
- The `LM_STUDIO_MODEL` env var must match the exact model ID loaded in LM Studio (use `curl http://localhost:1234/v1/models` to discover)

---

## 10. How AI Tools and RAG Should Use This File

### Priority

This file (`CONTEXT.md`) is the **top-priority context** for repository-wide work. Load it first when:
- Implementing new features
- Refactoring existing code
- Understanding architecture decisions
- Adding new bounded contexts or adapters

### Companion Documents

| Document | Purpose |
|----------|---------|
| `README.md` | Human-first introduction, badges, quick start |
| `AGENTS.md` | AI agent definitions and collaboration patterns |
| `.github/agents/*.md` | Individual agent role definitions |

### Instructions for AI Agents

1. **Always read `CONTEXT.md` first** before significant changes or large refactors
2. **Treat rules as constraints** — do not violate conventions unless explicitly updating them
3. **Reference this file** when documenting architectural decisions
4. **Follow naming conventions** exactly as specified
5. **Use `cloud-native-utils`** utilities instead of custom implementations
6. **Preserve hexagonal architecture** — ports in domain, adapters separate
7. **Update `CONTEXT.md`** via `CONTEXT-maintainer` agent when architecture evolves
