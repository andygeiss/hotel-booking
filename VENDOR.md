# VENDOR.md

## Overview

This document describes the external vendor libraries used in this project and provides guidance on when and how to use them. The project primarily depends on `cloud-native-utils` as its core utility library, with transitive dependencies for OIDC authentication and Kafka messaging.

**Philosophy:** Prefer reusing vendor functionality over re-implementing similar utilities. This keeps the codebase lean and benefits from battle-tested, maintained libraries.

---

## Approved Vendor Libraries

### cloud-native-utils

- **Purpose:** Core utility library providing modular, cloud-native building blocks for Go applications. This is the primary vendor dependency and should be the first choice for cross-cutting concerns.
- **Repository:** [github.com/andygeiss/cloud-native-utils](https://github.com/andygeiss/cloud-native-utils)
- **Version:** v0.4.8
- **Documentation:** [pkg.go.dev](https://pkg.go.dev/github.com/andygeiss/cloud-native-utils)

#### Key Packages

| Package | Purpose | Used In This Project |
|---------|---------|---------------------|
| `assert` | Test assertions (`assert.That`) | `internal/domain/*_test.go`, `cmd/*_test.go` |
| `logging` | Structured JSON logging | `cmd/server/main.go` via `logging.NewJsonLogger()` |
| `messaging` | Pub/sub dispatcher (in-memory or Kafka) | `cmd/cli/main.go`, `internal/adapters/outbound/event_publisher.go` |
| `redirecting` | HTMX-compatible redirects | `internal/adapters/inbound/http_index.go` |
| `resource` | Generic CRUD interface with backends | `internal/adapters/outbound/file_index_repository.go` |
| `security` | OIDC, encryption, HTTP server | `cmd/server/main.go`, `internal/adapters/inbound/router.go` |
| `service` | Context helpers, lifecycle hooks | `cmd/server/main.go`, `cmd/cli/main.go` |
| `templating` | HTML template engine with `embed.FS` | `internal/adapters/inbound/router.go`, `http_view.go` |

#### When to Use

| Concern | Package | Pattern |
|---------|---------|---------|
| **Testing assertions** | `assert` | `assert.That(t, "description", actual, expected)` |
| **Structured logging** | `logging` | `logging.NewJsonLogger()` at startup, pass to handlers |
| **HTTP request logging** | `logging` | `logging.WithLogging(logger, handler)` middleware |
| **Event publishing/subscribing** | `messaging` | `messaging.NewExternalDispatcher()` for Kafka |
| **HTTP redirects** | `redirecting` | `redirecting.Redirect(w, r, "/path")` |
| **CRUD repositories** | `resource` | `resource.NewJsonFileAccess[K, V](filename)` |
| **Mock repositories** | `resource` | `resource.NewMockAccess[K, V]()` for tests |
| **HTTP server setup** | `security` | `security.NewServer(mux)` with env-based config |
| **OIDC authentication** | `security` | `security.NewServeMux(ctx, efs)` with session management |
| **Auth middleware** | `security` | `security.WithAuth(sessions, handler)` |
| **Context with signals** | `service` | `service.Context()` for graceful shutdown |
| **Shutdown hooks** | `service` | `service.RegisterOnContextDone(ctx, fn)` |
| **Function wrapping** | `service` | `service.Wrap(fn)` for context-aware functions |
| **HTML templating** | `templating` | `templating.NewEngine(efs)` with `embed.FS` |

#### Integration Patterns

**Logging Setup:**
```go
import "github.com/andygeiss/cloud-native-utils/logging"

logger := logging.NewJsonLogger()
mux.HandleFunc("GET /path", logging.WithLogging(logger, handler))
```

**Messaging Setup:**
```go
import "github.com/andygeiss/cloud-native-utils/messaging"

// For Kafka (requires KAFKA_BROKERS env var)
dispatcher := messaging.NewExternalDispatcher()

// For in-memory (testing/development)
dispatcher := messaging.NewInternalDispatcher()
```

**Repository Setup:**
```go
import "github.com/andygeiss/cloud-native-utils/resource"

// JSON file persistence
repo := resource.NewJsonFileAccess[KeyType, ValueType](filename)

// In-memory (testing)
repo := resource.NewInMemoryAccess[KeyType, ValueType]()
```

**HTTP Server Setup:**
```go
import "github.com/andygeiss/cloud-native-utils/security"

mux, sessions := security.NewServeMux(ctx, efs)
srv := security.NewServer(mux)
```

#### Cautions

- **Environment variables:** Many packages read configuration from environment variables at initialization. Ensure variables are set before creating instances.
- **OIDC configuration:** Requires `OIDC_ISSUER`, `OIDC_CLIENT_ID`, `OIDC_CLIENT_SECRET`, `OIDC_REDIRECT_URL` environment variables.
- **Kafka configuration:** Requires `KAFKA_BROKERS` and optionally `KAFKA_CONSUMER_GROUP_ID`.
- **Server timeouts:** Configured via `SERVER_*_TIMEOUT` environment variables.

---

### coreos/go-oidc (Transitive)

- **Purpose:** OpenID Connect client library for Go
- **Repository:** [github.com/coreos/go-oidc/v3](https://github.com/coreos/go-oidc)
- **Version:** v3.17.0
- **Status:** Transitive dependency via `cloud-native-utils/security`

#### When to Use

**Do not use directly.** Use `cloud-native-utils/security` which wraps this library with:
- Automatic provider discovery
- Session management integration
- Environment-based configuration

The only case to use directly is if you need advanced OIDC features not exposed by cloud-native-utils.

---

### segmentio/kafka-go (Transitive)

- **Purpose:** Kafka client library for Go
- **Repository:** [github.com/segmentio/kafka-go](https://github.com/segmentio/kafka-go)
- **Version:** v0.4.49
- **Status:** Transitive dependency via `cloud-native-utils/messaging`

#### When to Use

**Do not use directly.** Use `cloud-native-utils/messaging` which provides:
- Unified `Dispatcher` interface
- In-memory and Kafka backends with the same API
- Simplified publish/subscribe patterns

The only case to use directly is for advanced Kafka features (consumer groups, partitioning, transactions) not exposed by the dispatcher abstraction.

---

### golang.org/x/oauth2 (Transitive)

- **Purpose:** OAuth 2.0 client library
- **Repository:** [golang.org/x/oauth2](https://pkg.go.dev/golang.org/x/oauth2)
- **Version:** v0.34.0
- **Status:** Transitive dependency via `cloud-native-utils/security`

#### When to Use

**Do not use directly.** Use `cloud-native-utils/security` which handles:
- PKCE code generation
- Token exchange
- Session management

---

## Cross-cutting Concerns and Recommended Patterns

### Testing

| Concern | Vendor | Pattern |
|---------|--------|---------|
| Assertions | `cloud-native-utils/assert` | `assert.That(t, msg, actual, expected)` |
| Mock repositories | `cloud-native-utils/resource` | `resource.NewMockAccess[K, V]()` |
| Mock functions | Standard library | Create interface + mock struct |

### Logging

| Concern | Vendor | Pattern |
|---------|--------|---------|
| Structured logging | `cloud-native-utils/logging` | `logging.NewJsonLogger()` |
| Request logging | `cloud-native-utils/logging` | `logging.WithLogging(logger, handler)` |

### HTTP

| Concern | Vendor | Pattern |
|---------|--------|---------|
| Server creation | `cloud-native-utils/security` | `security.NewServer(mux)` |
| Routing + sessions | `cloud-native-utils/security` | `security.NewServeMux(ctx, efs)` |
| Authentication | `cloud-native-utils/security` | `security.WithAuth(sessions, handler)` |
| Templating | `cloud-native-utils/templating` | `templating.NewEngine(efs)` |
| Redirects | `cloud-native-utils/redirecting` | `redirecting.Redirect(w, r, path)` |

### Messaging

| Concern | Vendor | Pattern |
|---------|--------|---------|
| Event dispatcher | `cloud-native-utils/messaging` | `messaging.NewExternalDispatcher()` |
| Publishing | `messaging.Dispatcher` | `dispatcher.Publish(ctx, msg)` |
| Subscribing | `messaging.Dispatcher` | `dispatcher.Subscribe(ctx, topic, handler)` |

### Persistence

| Concern | Vendor | Pattern |
|---------|--------|---------|
| CRUD interface | `cloud-native-utils/resource` | `resource.Access[K, V]` interface |
| JSON file storage | `cloud-native-utils/resource` | `resource.NewJsonFileAccess[K, V](file)` |
| In-memory storage | `cloud-native-utils/resource` | `resource.NewInMemoryAccess[K, V]()` |

### Resilience

| Concern | Vendor | Pattern |
|---------|--------|---------|
| Circuit breaker | `cloud-native-utils/stability` | `stability.Breaker(fn, threshold)` |
| Retry | `cloud-native-utils/stability` | `stability.Retry(fn, attempts, delay)` |
| Throttle | `cloud-native-utils/stability` | `stability.Throttle(fn, limit)` |
| Timeout | `cloud-native-utils/stability` | `stability.Timeout(fn, duration)` |

---

## Vendors to Avoid

### Testing Frameworks

**Avoid:** testify, gomega, ginkgo, goconvey

**Reason:** `cloud-native-utils/assert` provides sufficient assertion capabilities with a minimal API. The standard `testing` package plus `assert.That` covers all testing needs without additional complexity.

### Logging Libraries

**Avoid:** logrus, zap, zerolog

**Reason:** `cloud-native-utils/logging` wraps the standard `log/slog` package with JSON formatting. Using a different logger would create inconsistency and duplicate functionality.

### HTTP Routers

**Avoid:** gorilla/mux, chi, gin, echo

**Reason:** Go 1.22+ `http.ServeMux` supports pattern matching (e.g., `GET /path/{id}`). Combined with `cloud-native-utils/security.NewServeMux`, there's no need for third-party routers.

### Configuration Libraries

**Avoid:** viper, envconfig, godotenv

**Reason:** This project uses environment variables directly via `os.Getenv`. The `.env` file is loaded by Docker Compose or shell, not by the application. Adding configuration libraries adds complexity without benefit.

### Dependency Injection Frameworks

**Avoid:** wire, dig, fx

**Reason:** This project uses constructor-based dependency injection. Functions like `NewIndexingService(reader, repo, publisher)` provide explicit, traceable dependencies without magic.

---

## Adding New Vendor Libraries

Before adding a new vendor library:

1. **Check cloud-native-utils first:** It may already have the functionality you need.
2. **Evaluate necessity:** Can the functionality be implemented in 50-100 lines of Go?
3. **Consider maintenance:** Is the library actively maintained? Does it have a stable API?
4. **Check compatibility:** Does it work with Go 1.25+ and the existing stack?

If adding a new library:

1. Add to `go.mod` with `go get`
2. Document in this file with:
   - Purpose and repository link
   - Key packages/functions used
   - Integration patterns
   - Cautions and constraints
3. Update `CONTEXT.md` if it affects architecture or conventions
