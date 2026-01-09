# VENDOR.md

This file documents external libraries used in this project, their purposes, and recommended usage patterns. It serves as guidance for AI agents and developers to leverage existing vendor capabilities instead of reimplementing functionality.

---

## Overview

This project follows a **minimal dependency philosophy**. The primary external dependency is `cloud-native-utils`, which provides most infrastructure concerns. Other dependencies are indirect (pulled in by `cloud-native-utils`) or standard library extensions.

| Category | Primary Vendor |
|----------|----------------|
| Infrastructure utilities | `github.com/andygeiss/cloud-native-utils` |
| Authentication | `github.com/coreos/go-oidc/v3` (via cloud-native-utils) |
| Event streaming | `github.com/segmentio/kafka-go` (via cloud-native-utils) |

---

## Approved Vendor Libraries

### cloud-native-utils

- **Purpose**: Core infrastructure library providing HTTP server, authentication, messaging, logging, templating, event contracts, and resilience patterns for cloud-native Go applications.
- **Repository**: https://github.com/andygeiss/cloud-native-utils
- **Version**: v0.4.12 (see `go.mod`)

#### Key Packages

| Package | Import Path | Purpose |
|---------|-------------|---------|| `event` | `github.com/andygeiss/cloud-native-utils/event` | Domain event contracts (Event interface, EventPublisher, EventSubscriber) || `logging` | `github.com/andygeiss/cloud-native-utils/logging` | Structured JSON logging with `slog` |
| `messaging` | `github.com/andygeiss/cloud-native-utils/messaging` | Kafka pub/sub abstraction |
| `redirecting` | `github.com/andygeiss/cloud-native-utils/redirecting` | HTTP redirect helpers |
| `resource` | `github.com/andygeiss/cloud-native-utils/resource` | Generic repository patterns (JSON file access) |
| `security` | `github.com/andygeiss/cloud-native-utils/security` | HTTP server, OIDC auth, session management |
| `service` | `github.com/andygeiss/cloud-native-utils/service` | Context handling, graceful shutdown, resilience wrappers |
| `templating` | `github.com/andygeiss/cloud-native-utils/templating` | HTML template engine with `fs.FS` support |

#### When to Use

| Concern | Use This | Instead Of |
|---------|----------|------------|
| Domain event contracts | `event.Event`, `event.EventPublisher` | Custom event interfaces |
| HTTP server setup | `security.NewServer()` | Custom `http.Server` configuration |
| HTTP routing with sessions | `security.NewServeMux()` | Plain `http.NewServeMux()` |
| OIDC authentication | `security.WithAuth()` middleware | Custom OIDC implementation |
| Request logging | `logging.WithLogging()` | Custom logging middleware |
| JSON structured logging | `logging.NewJsonLogger()` | Custom `slog.Handler` |
| Kafka pub/sub | `messaging.NewKafkaDispatcher()` | Direct `kafka-go` usage |
| Message handling | `messaging.Dispatcher` interface | Custom message broker abstraction |
| HTML templates | `templating.NewEngine()` | Direct `html/template` usage |
| JSON file repository | `resource.JsonFileAccess` | Custom file I/O |
| Graceful shutdown | `service.Context()`, `service.RegisterOnContextDone()` | Manual signal handling |
| Resilience (retry, circuit breaker) | `service.Wrap()` | Custom resilience patterns |

#### Integration Patterns

**HTTP Server Setup** (in `cmd/server/main.go`):
```go
import (
    "github.com/andygeiss/cloud-native-utils/logging"
    "github.com/andygeiss/cloud-native-utils/security"
    "github.com/andygeiss/cloud-native-utils/service"
)

func main() {
    ctx, cancel := service.Context()
    defer cancel()

    logger := logging.NewJsonLogger()
    mux := inbound.Route(ctx, efs, logger)
    srv := security.NewServer(mux)
    
    service.RegisterOnContextDone(ctx, func() {
        _ = srv.Shutdown(context.Background())
    })
    
    _ = srv.ListenAndServe()
}
```

**Routing with Authentication** (in `internal/adapters/inbound/router.go`):
```go
import (
    "github.com/andygeiss/cloud-native-utils/logging"
    "github.com/andygeiss/cloud-native-utils/security"
    "github.com/andygeiss/cloud-native-utils/templating"
)

func Route(ctx context.Context, efs fs.FS, logger *slog.Logger) *http.ServeMux {
    mux, serverSessions := security.NewServeMux(ctx, efs)
    e := templating.NewEngine(efs)
    e.Parse("assets/templates/*.tmpl")
    
    // Protected route with logging and auth
    mux.HandleFunc("GET /ui/", 
        logging.WithLogging(logger, 
            security.WithAuth(serverSessions, HttpViewIndex(e))))
    
    return mux
}
```

**Event Publishing** (in `internal/adapters/outbound/event_publisher.go`):
```go
import "github.com/andygeiss/cloud-native-utils/messaging"

type EventPublisher struct {
    dispatcher messaging.Dispatcher
}

func (ep *EventPublisher) Publish(ctx context.Context, e event.Event) error {
    encoded, _ := json.Marshal(e)
    msg := messaging.NewMessage(e.Topic(), encoded)
    return ep.dispatcher.Publish(ctx, msg)
}
```

**Event Subscribing** (in `internal/adapters/inbound/event_subscriber.go`):
```go
import (
    "github.com/andygeiss/cloud-native-utils/messaging"
    "github.com/andygeiss/cloud-native-utils/service"
)

func (es *EventSubscriber) Subscribe(ctx context.Context, topic string, factory func() event.Event, handler func(e event.Event) error) error {
    messageFn := func(msg messaging.Message) (messaging.MessageState, error) {
        evt := factory()
        json.Unmarshal(msg.Data, evt)
        handler(evt)
        return messaging.MessageStateCompleted, nil
    }
    return es.dispatcher.Subscribe(ctx, topic, service.Wrap(messageFn))
}
```

**JSON File Repository** (in `internal/adapters/outbound/file_index_repository.go`):
```go
import "github.com/andygeiss/cloud-native-utils/resource"

func NewFileIndexRepository(filename string) indexing.IndexRepository {
    return resource.NewJsonFileAccess[indexing.IndexID, indexing.Index](filename)
}
```

The `resource.Access[K, V]` interface provides CRUD operations:
- `Create(ctx, key, value)` — Create new entry
- `Read(ctx, key)` — Read entry by key
- `Update(ctx, key, value)` — Update existing entry
- `Delete(ctx, key)` — Delete entry
- `ReadAll(ctx)` — Read all entries (useful for search operations)

#### Cautions

- **Version compatibility**: Always check release notes when upgrading; API changes may occur
- **OIDC configuration**: Requires proper environment variables (`OIDC_*`) and Keycloak setup
- **Kafka dispatcher**: Requires `KAFKA_BROKERS` environment variable
- **Server timeouts**: Configured via `SERVER_*` environment variables (see `.env.example`)

---

### go-oidc/v3 (Indirect)

- **Purpose**: OpenID Connect client library for Go, used by `cloud-native-utils/security` for OIDC authentication.
- **Repository**: https://github.com/coreos/go-oidc
- **Usage**: Indirect dependency; access through `cloud-native-utils/security`

#### When to Use

- **Do**: Use `security.WithAuth()` middleware for OIDC-protected routes
- **Don't**: Import `go-oidc` directly; let `cloud-native-utils` handle the integration

#### Context Values

After authentication, these context keys are available (from `security` package):

| Context Key | Type | Description |
|-------------|------|-------------|
| `security.ContextSessionID` | `string` | Session identifier |
| `security.ContextEmail` | `string` | User email |
| `security.ContextName` | `string` | User display name |
| `security.ContextSubject` | `string` | OIDC subject claim |
| `security.ContextIssuer` | `string` | OIDC issuer URL |
| `security.ContextVerified` | `bool` | Email verification status |

---

### kafka-go (Indirect)

- **Purpose**: Kafka client library for Go, used by `cloud-native-utils/messaging` for event streaming.
- **Repository**: https://github.com/segmentio/kafka-go
- **Usage**: Indirect dependency; access through `cloud-native-utils/messaging`

#### When to Use

- **Do**: Use `messaging.Dispatcher` interface for pub/sub operations
- **Don't**: Import `kafka-go` directly unless extending messaging functionality

#### Configuration

Set via environment variables:

| Variable | Description |
|----------|-------------|
| `KAFKA_BROKERS` | Comma-separated broker addresses |
| `KAFKA_CONSUMER_GROUP_ID` | Consumer group identifier |

---

## Cross-Cutting Concerns

### Logging

**Preferred approach**: Use `cloud-native-utils/logging`

```go
logger := logging.NewJsonLogger()
logger.Info("operation completed", "key", value, "duration", elapsed)
```

**Middleware pattern**:
```go
mux.HandleFunc("GET /path", logging.WithLogging(logger, handler))
```

### Error Handling

**Domain layer**: Return errors; do not log
```go
func (s *Service) DoWork(ctx context.Context) error {
    if err := s.repo.Save(ctx, data); err != nil {
        return fmt.Errorf("save failed: %w", err)
    }
    return nil
}
```

**Adapter layer**: Log and translate to appropriate responses
```go
func HttpHandler(w http.ResponseWriter, r *http.Request) {
    if err := service.DoWork(r.Context()); err != nil {
        logger.Error("operation failed", "error", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
}
```

### Resilience

**Use `service.Wrap()`** to add circuit breaker, retry, and debounce:

```go
import "github.com/andygeiss/cloud-native-utils/service"

wrappedFn := service.Wrap(messageFn)
dispatcher.Subscribe(ctx, topic, wrappedFn)
```

Configuration via environment:

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVICE_BREAKER_THRESHOLD` | Failures before circuit opens | `5` |
| `SERVICE_DEBOUNCE_PER_SEC` | Max events per second | `10` |
| `SERVICE_RETRY_DELAY` | Delay between retries | `5s` |
| `SERVICE_RETRY_MAX` | Maximum retry attempts | `3` |
| `SERVICE_TIMEOUT` | External call timeout | `5s` |

### Agent Tool Execution

**Pattern for implementing tool executors** (in `internal/adapters/outbound/`):

The `agent.ToolExecutor` interface enables LLM agents to invoke tools. Implement this pattern when adding new agent capabilities:

```go
// Define tool argument struct in agent/entities.go
type SearchIndexToolArgs struct {
    Query string `json:"query"`
    Limit int    `json:"limit,omitempty"`
}

// Implement ToolExecutor in adapters/outbound/
type IndexSearchToolExecutor struct {
    tools           map[string]toolFunc
    indexingService *indexing.IndexingService
    indexID         indexing.IndexID
}

func (e *IndexSearchToolExecutor) Execute(ctx context.Context, toolName string, arguments string) (string, error) {
    tool, ok := e.tools[toolName]
    if !ok {
        return "", fmt.Errorf("tool not found: %s", toolName)
    }
    return tool(ctx, arguments)
}

func (e *IndexSearchToolExecutor) GetAvailableTools() []string {
    // Return list of tool names
}

func (e *IndexSearchToolExecutor) HasTool(name string) bool {
    _, ok := e.tools[name]
    return ok
}

func (e *IndexSearchToolExecutor) GetToolDefinitions() []agent.ToolDefinition {
    // Return OpenAI-compatible tool definitions for LLM
}
```

**When to use**:
- Adding new capabilities the LLM agent can invoke (file search, API calls, etc.)
- Each tool executor handles a related set of tools
- Wire executors in entry points (`cmd/cli/main.go`)

**Integration pattern**:
```go
// In cmd/cli/main.go
toolExecutor := outbound.NewIndexSearchToolExecutor(indexingService, indexID)
taskService := agent.NewTaskService(llmClient, toolExecutor, eventPublisher)
```

---

## Vendors to Avoid

| Library | Reason | Use Instead |
|---------|--------|-------------|
| `gin`, `echo`, `chi` | Project uses stdlib `net/http` with `cloud-native-utils` | `security.NewServeMux()` |
| `logrus`, `zap` | Project standardized on `slog` | `logging.NewJsonLogger()` |
| `viper`, `envconfig` | Project uses direct `os.Getenv()` | Environment variables |
| `testify` (for mocking) | Prefer interface-based mocking | Implement test doubles |
| Direct `kafka-go` | Abstracted by `messaging` package | `messaging.Dispatcher` |
| Direct `go-oidc` | Abstracted by `security` package | `security.WithAuth()` |

---

## Adding New Vendors

Before adding a new vendor dependency:

1. **Check if `cloud-native-utils` already provides the capability**
2. **Verify the library is actively maintained** (recent commits, issue response)
3. **Prefer stdlib extensions** over full frameworks
4. **Document the addition** in this file with:
   - Purpose statement
   - Key packages/functions
   - Integration pattern
   - When to use / when not to use

Update this file via the `VENDOR-maintainer` agent when dependencies change.

---

## Version Notes

### cloud-native-utils v0.4.12

Current version. Includes the `event` package with domain event contracts (`Event` interface, `EventPublisher`, `EventSubscriber`, `EventFactoryFn`, `EventHandlerFn`).

Key capabilities:
- Domain event contracts and infrastructure
- OIDC integration with Keycloak
- Kafka messaging abstraction
- Generic repository pattern (`resource.Access[K, V]`)
- Structured logging with `slog`
- HTTP security middleware

For upgrade guidance, see: https://github.com/andygeiss/cloud-native-utils/releases
