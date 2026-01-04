# VENDOR.md

> **Vendor documentation** for AI agents, retrieval systems, and developers.  
> Consult this file when working with external libraries, especially for cross-cutting concerns.

---

## 1. Vendor Philosophy

This template **prefers reuse over reinvention**. Before writing new utility code for cross-cutting concerns, check if an approved vendor library already provides the functionality.

**Rule**: Use `cloud-native-utils` first. Only implement custom primitives when the library clearly does not cover the use case or when there is a justified, documented reason.

---

## 2. Primary Vendor: `cloud-native-utils`

| Property       | Value                                                                 |
|----------------|-----------------------------------------------------------------------|
| **Repository** | [github.com/andygeiss/cloud-native-utils](https://github.com/andygeiss/cloud-native-utils) |
| **Go Docs**    | [pkg.go.dev/github.com/andygeiss/cloud-native-utils](https://pkg.go.dev/github.com/andygeiss/cloud-native-utils) |
| **Version**    | v0.4.8 (see `go.mod`)                                                 |
| **License**    | MIT                                                                   |

### 2.1 Purpose

A collection of modular, high-performance utilities for building cloud-native Go applications. Each package is independent—no monolithic framework. The library covers:

- Testing assertions
- Transactional consistency
- Concurrency and channel helpers
- Generic resource access (CRUD)
- Security (encryption, hashing, OIDC)
- Service lifecycle and orchestration
- Stability patterns (circuit breaker, retry, throttle)
- Templating and i18n
- Messaging (pub/sub)
- Scheduling primitives

---

## 3. Package Reference

### 3.1 `assert` — Test Assertions

**Import**: `github.com/andygeiss/cloud-native-utils/assert`

| Function        | Description                                      |
|-----------------|--------------------------------------------------|
| `That(t, msg, got, want)` | Assert equality with a clear error message |

**When to use**:  
- All unit tests in this template use `assert.That` for readable, consistent assertions.

**Pattern** (from this template):
```go
import "github.com/andygeiss/cloud-native-utils/assert"

func Test_Index_Hash_With_No_FileInfos_Should_Return_Valid_Hash(t *testing.T) {
    index := indexing.Index{ID: "empty-index", FileInfos: []indexing.FileInfo{}}
    hash := index.Hash()
    assert.That(t, "empty index must have a valid hash (size of 64 bytes)", len(hash), 64)
}
```

---

### 3.2 `resource` — Generic CRUD Access

**Import**: `github.com/andygeiss/cloud-native-utils/resource`

| Type / Function               | Description                                       |
|-------------------------------|---------------------------------------------------|
| `Access[K, V]`                | Generic interface for CRUD on key-value pairs     |
| `NewInMemoryAccess[K, V]()`   | In-memory backend (useful for tests)              |
| `NewJsonFileAccess[K, V](f)`  | JSON file–backed storage                          |
| `NewYamlFileAccess[K, V](f)`  | YAML file–backed storage                          |
| `NewSqliteAccess[K, V](db)`   | SQLite-backed storage                             |
| `NewIndexedAccess[K, V](...)`  | Add secondary indexing to any `Access`           |
| `NewMockAccess[K, V]()`       | Mock for unit tests                               |

**When to use**:  
- Implement `IndexRepository` or any repository port where a key-value shape fits.
- Prefer this over ad-hoc repository interfaces.

**Pattern** (from this template):
```go
// Domain defines the port as a type alias
type IndexRepository resource.Access[IndexID, Index]

// Outbound adapter uses the JSON file implementation
func NewFileIndexRepository(filename string) indexing.IndexRepository {
    return resource.NewJsonFileAccess[indexing.IndexID, indexing.Index](filename)
}
```

**Where it lives**: Outbound adapters in `internal/adapters/outbound/`.

---

### 3.3 `messaging` — Publish/Subscribe

**Import**: `github.com/andygeiss/cloud-native-utils/messaging`

| Type / Function                 | Description                                       |
|---------------------------------|---------------------------------------------------|
| `Dispatcher`                    | Interface for pub/sub messaging                   |
| `NewInternalDispatcher()`       | In-memory dispatcher for local events             |
| `NewExternalDispatcher()`       | Kafka-backed dispatcher (requires `KAFKA_BROKERS`)|
| `NewMessage(topic, payload)`    | Create a new message                              |
| `Message`                       | Message struct with topic and payload             |
| `MessageState`                  | Enum for message handling results                 |

**When to use**:  
- Publish domain events from services.
- Decouple bounded contexts via event-driven patterns.

**Pattern** (from this template):
```go
type EventPublisher struct {
    dispatcher messaging.Dispatcher
}

func (ep *EventPublisher) Publish(ctx context.Context, e event.Event) error {
    encoded, _ := json.Marshal(e)
    msg := messaging.NewMessage(e.Topic(), encoded)
    return ep.dispatcher.Publish(ctx, msg)
}
```

**Where it lives**: Outbound adapters in `internal/adapters/outbound/`.

---

### 3.4 `stability` — Resilience Patterns

**Import**: `github.com/andygeiss/cloud-native-utils/stability`

| Function                          | Description                                      |
|-----------------------------------|--------------------------------------------------|
| `Breaker(fn, threshold)`          | Circuit breaker—opens after N failures           |
| `Retry(fn, attempts, delay)`      | Retry with configurable attempts and delay       |
| `Throttle(fn, maxConcurrent)`     | Limit concurrent executions                      |
| `Debounce(fn, duration)`          | Delay execution until quiet period               |
| `Timeout(fn, duration)`           | Enforce maximum execution time                   |

**When to use**:  
- Wrap calls to external services (HTTP APIs, databases, third-party systems).
- Prefer these wrappers over custom retry/circuit-breaker logic.

**Pattern**:
```go
import "github.com/andygeiss/cloud-native-utils/stability"

// Wrap an external call with retry and circuit breaker
fn := stability.Retry(stability.Breaker(externalCall, 3), 5, time.Second)
result, err := fn(ctx, input)
```

**Where it lives**: Outbound adapters wrapping external calls.

---

### 3.5 `service` — Context & Lifecycle

**Import**: `github.com/andygeiss/cloud-native-utils/service`

| Function / Type                        | Description                                  |
|----------------------------------------|----------------------------------------------|
| `Context()`                            | Signal-aware context (SIGINT/SIGTERM)        |
| `Wrap(fn)`                             | Wrap a function for context-aware execution  |
| `RegisterOnContextDone(ctx, cleanup)`  | Register cleanup on context cancellation     |
| `Function[In, Out]`                    | Generic function type used by stability pkg  |

**When to use**:  
- Create root context in CLI/HTTP entry points.
- Register graceful shutdown handlers.

**Pattern**:
```go
import "github.com/andygeiss/cloud-native-utils/service"

func main() {
    ctx, cancel := service.Context()
    defer cancel()

    service.RegisterOnContextDone(ctx, func() {
        // cleanup resources
    })
}
```

**Where it lives**: Application layer in `cmd/*/main.go`.

---

### 3.6 `security` — Encryption, Hashing, HTTP Server

**Import**: `github.com/andygeiss/cloud-native-utils/security`

| Function / Type                  | Description                                         |
|----------------------------------|-----------------------------------------------------|
| `GenerateKey()`                  | Generate AES-256 key                                |
| `Encrypt(plaintext, key)`        | AES-GCM encryption                                  |
| `Decrypt(ciphertext, key)`       | AES-GCM decryption                                  |
| `Password(pw)`                   | Hash password with bcrypt                           |
| `IsPasswordValid(hash, pw)`      | Verify bcrypt password                              |
| `GenerateID()`                   | Generate secure random ID                           |
| `GeneratePKCE()`                 | Generate PKCE verifier/challenge for OAuth2         |
| `NewServer(handler)`             | Preconfigured secure HTTP server                    |
| `IdentityProvider`               | OIDC helpers (Login, Callback, Logout)              |

**When to use**:  
- Encrypt sensitive data at rest.
- Hash user passwords.
- Set up production HTTP servers with proper timeouts.

**Pattern**:
```go
import "github.com/andygeiss/cloud-native-utils/security"

mux := http.NewServeMux()
// ... register handlers
server := security.NewServer(mux)  // reads PORT, SERVER_*_TIMEOUT env vars
```

**Where it lives**: Adapters (security utilities), application layer (HTTP server setup).

---

### 3.7 `efficiency` — Concurrency Helpers

**Import**: `github.com/andygeiss/cloud-native-utils/efficiency`

| Function                           | Description                                      |
|------------------------------------|--------------------------------------------------|
| `Generate(values...)`              | Create read-only channel from values             |
| `Merge(channels...)`               | Merge multiple channels into one                 |
| `Split(ch, n)`                     | Fan-out to N worker channels                     |
| `Process(ch, fn)`                  | Concurrent processing (NumCPU workers)           |
| `WithCompression(handler)`         | HTTP gzip middleware                             |

**When to use**:  
- Stream processing pipelines.
- Fan-out/fan-in patterns.
- HTTP response compression.

**Pattern**:
```go
import "github.com/andygeiss/cloud-native-utils/efficiency"

ch := efficiency.Generate(items...)
results, errs := efficiency.Process(ch, processFn)
```

---

### 3.8 `consistency` — Transactional Event Log

**Import**: `github.com/andygeiss/cloud-native-utils/consistency`

| Type / Function                          | Description                                  |
|------------------------------------------|----------------------------------------------|
| `NewJsonFileLogger[K, V](path)`          | File-backed transactional event log          |
| `WritePut(key, value)`                   | Log a put event                              |
| `WriteDelete(key)`                       | Log a delete event                           |
| `ReadEvents()`                           | Replay events from log                       |

**When to use**:  
- Event sourcing or audit trails.
- Durable operation logs.

---

### 3.9 `templating` — HTML Templating

**Import**: `github.com/andygeiss/cloud-native-utils/templating`

| Type / Function                | Description                                      |
|--------------------------------|--------------------------------------------------|
| `NewEngine(fs)`                | Create engine from `embed.FS`                    |
| `Parse(pattern)`               | Parse templates matching glob                    |
| `Render(w, name, data)`        | Render template to writer                        |

**When to use**:  
- HTML rendering in HTTP handlers.
- Use instead of rolling a custom template loader.

**Pattern**:
```go
//go:embed templates/*.html
var templatesFS embed.FS

engine := templating.NewEngine(templatesFS)
engine.Parse("templates/*.html")
engine.Render(w, "page.html", data)
```

---

### 3.10 `slices` — Generic Slice Utilities

**Import**: `github.com/andygeiss/cloud-native-utils/slices`

| Function              | Description                          |
|-----------------------|--------------------------------------|
| `Map(slice, fn)`      | Transform each element               |
| `Filter(slice, fn)`   | Filter elements by predicate         |
| `Unique(slice)`       | Remove duplicates                    |
| `Contains(slice, v)`  | Check membership                     |

**When to use**:  
- Functional-style slice transformations.
- Prefer over hand-rolled loops for common operations.

---

### 3.11 `scheduling` — Time/Day Primitives

**Import**: `github.com/andygeiss/cloud-native-utils/scheduling`

| Type / Function                        | Description                              |
|----------------------------------------|------------------------------------------|
| `TimeOfDay`, `MustTimeOfDay(h, m)`     | Time-of-day value object                 |
| `DayHours`, `NewDayHours(day, open, close)` | Opening hours for a weekday         |
| Slot/gap utilities                     | Booking system helpers                   |

**When to use**:  
- Booking or scheduling features.
- Business hours and slot management.

---

### 3.12 Additional Packages

| Package         | Purpose                                                      | When to use                                  |
|-----------------|--------------------------------------------------------------|----------------------------------------------|
| `logging`       | Structured JSON logging, HTTP middleware                     | Production logging, request tracing          |
| `redirecting`   | PRG pattern, HTMX-compatible redirects                       | Web apps with POST/Redirect/GET              |
| `i18n`          | YAML translations, date/money formatting                     | Internationalized applications               |
| `imaging`       | QR code generation                                           | QR features                                  |
| `extensibility` | Dynamic Go plugin loading                                    | Plugin architectures (platform-limited)      |

---

## 4. Integration Guidelines

### 4.1 Where Vendor Code Lives

| Integration Type                | Location                              |
|---------------------------------|---------------------------------------|
| Repository implementations      | `internal/adapters/outbound/`         |
| Event publishers                | `internal/adapters/outbound/`         |
| HTTP server setup               | `cmd/*/main.go`                       |
| Stability wrappers              | `internal/adapters/outbound/` (wrap external calls) |
| Test assertions                 | `*_test.go` (same directory as code)  |
| Templating engine               | `cmd/*/main.go` or dedicated handler adapter |

### 4.2 Domain Layer Rules

- **Never import `cloud-native-utils` directly in domain code** except:
  - Type aliases for ports (e.g., `type IndexRepository resource.Access[K, V]`).
- Domain logic remains infrastructure-agnostic.

### 4.3 Recommended Defaults

| Concern                   | Default Pattern                                              |
|---------------------------|--------------------------------------------------------------|
| Repository port           | `type XRepository resource.Access[K, V]`                     |
| JSON persistence          | `resource.NewJsonFileAccess[K, V](filename)`                 |
| In-memory (tests)         | `resource.NewInMemoryAccess[K, V]()`                         |
| Mock (tests)              | `resource.NewMockAccess[K, V]()`                             |
| Test assertions           | `assert.That(t, message, got, want)`                         |
| External call resilience  | `stability.Retry(stability.Breaker(fn, 3), 5, time.Second)`  |
| Signal-aware context      | `ctx, cancel := service.Context()`                           |
| Secure HTTP server        | `security.NewServer(mux)`                                    |

---

## 5. Checklist: Before Adding New Utilities

When implementing cross-cutting functionality:

1. **Search `cloud-native-utils`** for existing coverage.
2. **Consult this document** for recommended patterns.
3. **If covered**: Use the vendor package; add integration in adapters.
4. **If not covered**: Document the gap and implement minimally in `internal/`.
5. **Update `VENDOR.md`** if introducing a new vendor or new usage pattern.

---

## 6. Version Notes

| Version | Notable Changes                                              |
|---------|--------------------------------------------------------------|
| v0.4.8  | `resource.Access` made public; current template version      |

When upgrading `cloud-native-utils`:
- Review the [releases page](https://github.com/andygeiss/cloud-native-utils/releases).
- Check for breaking changes in packages this template uses.
- Update this document if APIs change significantly.

---

*Last updated: 2026-01-04*
