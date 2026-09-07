<p align="center">
<img src="https://github.com/andygeiss/hotel-booking/blob/main/cmd/server/assets/static/img/icon-192.png?raw=true" width="100"/>
</p>

# Hotel Booking

[![Go Reference](https://pkg.go.dev/badge/github.com/andygeiss/hotel-booking.svg)](https://pkg.go.dev/github.com/andygeiss/hotel-booking)
[![License](https://img.shields.io/github/license/andygeiss/hotel-booking)](https://github.com/andygeiss/hotel-booking/blob/master/LICENSE)
[![Releases](https://img.shields.io/github/v/release/andygeiss/hotel-booking)](https://github.com/andygeiss/hotel-booking/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/andygeiss/hotel-booking)](https://goreportcard.com/report/github.com/andygeiss/hotel-booking)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/f9f01632dff14c448dbd4688abbd04e8)](https://app.codacy.com/gh/andygeiss/hotel-booking/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade)

A hotel reservation and payment management system built with Go, demonstrating Domain-Driven Design (DDD) and Hexagonal Architecture (Ports & Adapters) patterns.

<p align="center">
<img src="https://github.com/andygeiss/hotel-booking/blob/main/cmd/server/assets/static/img/login.png?raw=true" width="300"/>
</p>

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Architecture](#architecture)
- [Bounded Contexts](#bounded-contexts)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [Usage](#usage)
- [Testing](#testing)
- [Configuration](#configuration)
- [Using as a Template](#using-as-a-template)
- [Baseline deviations](#baseline-deviations)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

This repository provides a reference implementation for structuring Go applications with clean architecture principles. It demonstrates how to:

- Organize code using **Hexagonal Architecture** (Ports & Adapters)
- Apply **Domain-Driven Design** tactical patterns (aggregates, entities, value objects, domain events)
- Structure code into **Bounded Contexts** with clear boundaries
- Implement **Event-Driven Communication** between contexts via Kafka
- Use the **Saga Pattern** for cross-context workflow orchestration
- Integrate authentication via **OIDC/Keycloak**
- Persist data with **PostgreSQL** using key/value storage (separate databases per bounded context)

---

## Key Features

- **Bounded Context Architecture** — Separate reservation, payment, and orchestration contexts
- **Developer Experience** — one `make` command surface, staticcheck and govulncheck as gates, comprehensive test coverage
- **Domain-Driven Design** — Aggregates, entities, value objects, domain services, and domain events
- **Event-Driven Communication** — Kafka-based pub/sub for inter-context messaging
- **Hexagonal Architecture** — Clear separation between domain logic and infrastructure
- **OIDC Authentication** — Keycloak integration with session management
- **PostgreSQL Persistence** — Key/value storage with separate databases per bounded context
- **Production-Ready Docker** — Multi-stage build with PGO optimization
- **Progressive Web App** — Service worker, manifest, and offline support
- **Saga Pattern** — Event-driven booking workflow with compensation on failure
- **MCP Integration** — Model Context Protocol endpoint for AI tool integration

---

## Architecture

```
                    ┌─────────────────────────────────────────┐
                    │            Entry Point                  │
                    │         cmd/server/main.go              │
                    │      (DI wiring, bootstrap)             │
                    └─────────────────┬───────────────────────┘
                                      │
         ┌────────────────────────────┼────────────────────────────┐
         │                            │                            │
         ▼                            ▼                            ▼
┌─────────────────┐          ┌─────────────────┐          ┌──────────────────┐
│ Inbound Adapter │          │  Domain Layer   │          │Outbound Adapter  │
│  (HTTP, Events) │─────────▶│ (Bounded Ctxs)  │◀─────────│ (Repos, Gateways)│
│                 │          │                 │          │                  │
│ implements      │          │   defines       │          │ implements       │
│ domain ports    │          │   ports         │          │ domain ports     │
└─────────────────┘          └─────────────────┘          └──────────────────┘
                                      │
                    ┌─────────────────┴──────────────┐
                    │                                │
         ┌──────────┴──────────┐                     │
         │                     │                     │
         ▼                     ▼                     ▼
┌─────────────────┐   ┌─────────────────┐   ┌─────────────────┐
│   Reservation   │   │     Payment     │   │  Orchestration  │
│    Context      │   │     Context     │   │     Layer       │
│                 │   │                 │   │                 │
│ aggregate.go    │   │ aggregate.go    │   │ booking_svc.go  │
│ service.go      │   │ service.go      │   │ event_handlers  │
│ events.go       │   │ events.go       │   │                 │
└─────────────────┘   └─────────────────┘   └─────────────────┘
         │                     │                     │
         └─────────────────────┴─────────────────────┘
                               │
                    ┌──────────┴──────────┐
                    │    Shared Kernel    │
                    │  (Money, IDs)       │
                    └─────────────────────┘
```

### Event-Driven Communication

Bounded contexts communicate via domain events through Kafka:

```
┌─────────────────┐     reservation.created      ┌─────────────────┐
│   Reservation   │ ─────────────────────────▶   │     Payment     │
│    Context      │                              │     Context     │
└─────────────────┘                              └─────────────────┘
                                                          │
                        payment.authorized                │
         ┌────────────────────────────────────────────────┘
         │
         ▼
┌─────────────────┐     payment.captured         ┌─────────────────┐
│  Orchestration  │ ─────────────────────────▶   │   Reservation   │
│     Layer       │                              │     Context     │
└─────────────────┘                              └─────────────────┘
```

**Event Topics:**
- `reservation.created` — Payment context subscribes to authorize payment
- `reservation.confirmed` — Notification context subscribes
- `reservation.cancelled` — Notification context subscribes
- `payment.authorized` — Orchestration subscribes to capture payment
- `payment.captured` — Reservation context subscribes to confirm reservation
- `payment.failed` — Orchestration subscribes for compensation

---

## Bounded Contexts

The domain is split into three bounded contexts with clear responsibilities:

| Context | Purpose | Key Aggregates | Database |
|---------|---------|----------------|----------|
| **Reservation** | Room booking lifecycle | `Reservation` | `reservation_db` |
| **Payment** | Payment processing | `Payment` | `payment_db` |
| **Orchestration** | Cross-context coordination | Saga coordination | — |

### Reservation Context

The Reservation aggregate manages the complete booking lifecycle:

```
Reservation (Aggregate Root)
├── ReservationID (Value Object)
├── GuestID (Value Object)
├── RoomID (Value Object)
├── DateRange (Value Object)
│   ├── CheckIn
│   └── CheckOut
├── TotalAmount (Money - Shared Kernel)
├── Guests (Entity Collection)
│   └── GuestInfo
│       ├── Name
│       ├── Email
│       └── PhoneNumber
└── ReservationStatus (Value Object)
    States: pending → confirmed → active → completed
                  ↘ cancelled
```

**Business Rules:**
- Minimum 1 night stay required
- Check-in must be in the future
- Cannot cancel within 24 hours of check-in
- Same-day checkout/check-in allowed (no overlap)
- Cancelled reservations don't block availability

### Payment Context

The Payment aggregate handles payment processing with retry support:

```
Payment (Aggregate Root)
├── PaymentID (Value Object)
├── ReservationID (Shared Kernel)
├── Amount (Money - Shared Kernel)
├── PaymentMethod
├── TransactionID
├── PaymentStatus (Value Object)
│   States: pending → authorized → captured
│                  ↘ failed      ↘ refunded
└── Attempts (Entity Collection)
    └── PaymentAttempt
        ├── Status
        ├── ErrorCode
        └── AttemptedAt
```

**Business Rules:**
- Authorization-Capture pattern (Authorize → Capture)
- Failed payments can be retried
- Only captured payments can be refunded

### Orchestration Layer (Saga Pattern)

Event-driven workflow coordination with compensation:

```
Booking Workflow:
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ 1. Create       │───▶│ 2. Authorize    │───▶│ 3. Capture      │
│    Reservation  │    │    Payment      │    │    Payment      │
│    (pending)    │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │                      │
                              ▼ (on failure)         ▼
                       Cancel Reservation     Refund + Cancel
                                                     │
                                                     ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │ 5. Send         │◀───│ 4. Confirm      │
                       │    Notification │    │    Reservation  │
                       └─────────────────┘    └─────────────────┘
```

---

## Project Structure

```
hotel-booking/
├── cmd/server/                   # HTTP server entry point
│   ├── main.go                   # DI wiring, bootstrap, lifecycle
│   └── assets/
│       ├── static/               # CSS, JS, images (embedded)
│       └── templates/            # HTML templates (*.tmpl, embedded)
│           └── error.tmpl        # User-friendly error page
├── docker-compose.yml            # Dev stack (PostgreSQL x2, Keycloak, Kafka, app)
├── Dockerfile                    # Multi-stage production build
├── Makefile                      # The only command surface; the gates live in `check`
├── SPEC.md                       # Project brief: job, why, guardrails, done means
├── migrations/
│   ├── reservation/
│   │   └── init.sql              # Reservation database schema (key/value)
│   └── payment/
│       └── init.sql              # Payment database schema (key/value)
├── internal/
│   ├── adapters/
│   │   ├── inbound/              # HTTP handlers, event subscribers
│   │   │   ├── router.go         # HTTP routing; returns the mux behind WithSecurity
│   │   │   ├── middleware.go     # secureHeaders, the CSP, the WithSecurity chain
│   │   │   ├── http_ops.go       # /healthz + pprof for the loopback ops listener
│   │   │   ├── http_{feature}.go # HTTP handlers
│   │   │   ├── http_error.go     # Error page handler
│   │   │   └── event_subscriber.go
│   │   └── outbound/             # Repositories, gateways, publishers
│   │       ├── postgres_connection.go
│   │       ├── postgres_reservation_repository.go
│   │       ├── postgres_payment_repository.go
│   │       ├── repository_{checker}.go
│   │       ├── mock_{service}.go
│   │       └── event_publisher.go
│   └── domain/
│       ├── shared/               # Shared kernel
│       │   └── types.go          # Cross-context types (Money, ReservationID)
│       ├── reservation/          # Reservation bounded context
│       │   ├── aggregate.go      # Reservation aggregate + value objects
│       │   ├── entities.go       # DateRange, GuestInfo
│       │   ├── events.go         # Domain events
│       │   ├── ports.go          # Interface definitions
│       │   ├── service.go        # ReservationService
│       │   └── tools.go          # MCP tools
│       ├── payment/              # Payment bounded context
│       │   ├── aggregate.go      # Payment aggregate + status
│       │   ├── entities.go       # PaymentAttempt
│       │   ├── events.go         # Domain events
│       │   ├── ports.go          # Interface definitions
│       │   ├── service.go        # PaymentService
│       │   └── tools.go          # MCP tools
│       └── orchestration/        # Cross-context coordination
│           ├── booking_service.go    # Saga coordinator
│           ├── event_handlers.go     # Event subscriptions
│           └── ports.go              # NotificationService interface
└── docs/
    └── ARCHITECTURE.md           # Detailed architecture documentation
```

---

## Getting Started

### Prerequisites

- **Go 1.27.1** — the version pinned by the [engineering baseline](https://github.com/andygeiss/baseline/blob/main/VERSIONS.md)
- **Make** — the only command surface; already on every machine
- **Docker** and **Docker Compose** (or Podman) — for the local stack only

`make check` runs staticcheck and govulncheck through `go run`, so it needs the network
but nothing installed.

### Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/andygeiss/hotel-booking.git
   cd hotel-booking
   ```

2. **Install the local stack tools:**
   ```bash
   brew install docker-compose podman graphviz
   ```
   `graphviz` is only needed by `make profile`.

3. **Configure environment:**
   ```bash
   cp .env.example .env
   cp .keycloak.json.example .keycloak.json
   ```

4. **Start the development stack:**
   ```bash
   sed -i '' "s/CHANGE_ME_LOCAL_SECRET/$(LC_ALL=C tr -dc 'A-Za-z0-9' < /dev/urandom | head -c 32)/g" .env .keycloak.json
   podman build -t "$USER/hotel-booking:latest" -f Dockerfile .
   docker-compose --env-file .env up -d
   ```
   The first line gives the app and the Keycloak realm the same random secret; it does
   nothing once the placeholder is gone. The rest starts two PostgreSQL databases
   (reservation_db, payment_db), Keycloak, Kafka, and the application. Full walkthrough:
   [docs/development-guide.md](docs/development-guide.md).

5. **Access the application:**
   - **App:** http://localhost:8080/ui
   - **Keycloak Admin:** http://localhost:8180/admin (admin:admin)

---

## Usage

### Commands

`make` is the whole command surface. `make check` before every commit, `make ci` before
every push — there is no CI server, so nothing runs the gates for you.

| Command | Description |
|---------|-------------|
| `make build` | Build the release-shaped binary into `bin/` |
| `make check` | Every gate against the working tree — the default target |
| `make ci` | The same gates against the commit, in a clean copy |
| `make clean` | Remove `bin/` |
| `make fmt` | Apply goimports and `go fix` |
| `make profile` | Generate the CPU profile the `-pgo` Docker build reads |
| `make run` | Run the server, loading `.env` if it is there |
| `make test` | The inner loop: `go test -race -shuffle=on ./...` |

The local container stack is not a Make target — see
[docs/development-guide.md](docs/development-guide.md) for `docker-compose` commands.

### Run Single Test

```bash
go test -v -run TestFunctionName ./internal/domain/reservation/...
```

### Booking Workflow

Once the application is running:

1. **Login** at http://localhost:8080/ui/login via Keycloak
2. **View Reservations** at `/ui/reservations` to see your bookings
3. **Create Reservation** at `/ui/reservations/new`:
   - Select a room and dates
   - Total is calculated automatically (nights x room price)
   - Submit to create a pending reservation
4. **View Details** at `/ui/reservations/{id}` to see reservation status
5. **Cancel Reservation** from the detail page (if >24 hours before check-in)

### API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/ui/` | GET | Dashboard (authenticated) |
| `/ui/login` | GET | Login page |
| `/ui/reservations` | GET | List user's reservations |
| `/ui/reservations/new` | GET | Reservation form |
| `/ui/reservations` | POST | Create reservation |
| `/ui/reservations/{id}` | GET | Reservation detail |
| `/ui/reservations/{id}/cancel` | POST | Cancel reservation |
| `/ui/error` | GET | Error page (query params: title, message, details) |
| `/mcp` | POST | MCP JSON-RPC endpoint for AI tools |

### MCP Endpoint

The application exposes an MCP (Model Context Protocol) endpoint for AI tool integration:

```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
```

See [ARCHITECTURE.md](docs/ARCHITECTURE.md#7-mcp-integration) for details on adding custom tools.

---

## Testing

### Unit Tests

```bash
make test
```

Every run uses `-race -shuffle=on`, the same flags `make check` uses. For coverage:

```bash
go test -coverprofile=.coverage.pprof ./internal/...
go tool cover -func=.coverage.pprof | grep total
```

### Integration Tests

Integration tests require external services (PostgreSQL databases, Kafka, Keycloak):

```bash
go test -tags=integration -v ./internal/...
```

### Test Organization

- Unit tests are colocated with source files (`*_test.go`)
- Integration tests are tagged with `//go:build integration`
- Test fixtures live in `testdata/` directories

### Test Naming Convention

Tests follow the pattern: `Test_{Component}_{Scenario}_Should_{ExpectedResult}`

```go
// Domain unit tests
func Test_Reservation_Confirm_From_Pending_Should_Change_Status(t *testing.T)

// Service tests
func Test_ReservationService_CreateReservation_Should_Succeed(t *testing.T)

// HTTP handler tests
func Test_Route_Liveness_Endpoint_Should_Return_200(t *testing.T)
```

---

## Configuration

**`./server -h` is the contract.** Every knob is one field of a `Config` struct parsed and
validated in `main` before a database opens or a listener binds. **Flags beat environment
variables beat built-in defaults**, and each flag's help text names its variable. A bad
setting is one line and exit 2:

```
$ ./server -port=http
server: port "http": want a number from 0 to 65535
```

Copy `.env.example` to `.env` for local development — `make run` loads it. The common
settings:

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_NAME` | Display name for UI and PWA | `Hotel Booking` |
| `APP_DESCRIPTION` | Application description | `Hotel reservation and payment management system` |
| `APP_SHORTNAME` | Docker image/container name | `hotel-booking` |
| `KAFKA_BROKERS` | Kafka broker addresses | `localhost:9092` |
| `OIDC_CLIENT_ID` | OIDC client ID | `hotel-booking` |
| `OIDC_CLIENT_SECRET` | OIDC client secret | Auto-generated |
| `OIDC_ISSUER` | Keycloak realm URL | `http://localhost:8180/realms/local` |
| `HOST` | Bind address; empty means every interface | empty |
| `PORT` | HTTP server port | `8080` |
| `RESERVATION_DB_HOST` | Reservation database host | `localhost` |
| `RESERVATION_DB_PORT` | Reservation database port | `5432` |
| `RESERVATION_DB_USER` | Reservation database user | `reservation` |
| `RESERVATION_DB_PASSWORD` | Reservation database password (dev fallback — see below) | none |
| `RESERVATION_DB_NAME` | Reservation database name | `reservation_db` |
| `RESERVATION_DB_SSLMODE` | SSL mode | `disable` |
| `PAYMENT_DB_HOST` | Payment database host | `localhost` |
| `PAYMENT_DB_PORT` | Payment database port | `5433` |
| `PAYMENT_DB_USER` | Payment database user | `payment` |
| `PAYMENT_DB_PASSWORD` | Payment database password (dev fallback — see below) | none |
| `PAYMENT_DB_NAME` | Payment database name | `payment_db` |
| `PAYMENT_DB_SSLMODE` | SSL mode | `disable` |

**Secrets are files.** The two database passwords are read from `CREDENTIALS_DIRECTORY`,
one file per secret (`reservation-db-password`, `payment-db-password`). A file is not
inherited by every child process and does not appear in a process listing, which is true
of neither a flag value nor an environment variable. The `*_PASSWORD` variables above are
a development fallback, and a recorded deviation — compose does not supply files yet.

`Config.LogValue` is an allowlist, so the boot log prints the safe fields and a secret
added to the struct later is not logged by accident.

See `.env.example` for the complete list with documentation, and
[docs/deployment-guide.md](docs/deployment-guide.md#configuration) for the rest.

---

## Using as a Template

### Quick Start

1. **Clone and reinitialize:**
   ```bash
   git clone https://github.com/andygeiss/hotel-booking my-project
   cd my-project
   rm -rf .git && git init
   ```

2. **Update module path:**
   ```bash
   go mod edit -module github.com/yourorg/my-project
   # Update import paths in all .go files
   ```

3. **Configure project identity:**
   ```bash
   cp .env.example .env
   # Edit APP_NAME, APP_SHORTNAME, APP_DESCRIPTION
   ```

4. **Add your domain logic:**
   - Create bounded contexts in `internal/domain/`
   - Add shared types to `internal/domain/shared/`
   - Implement adapters in `internal/adapters/`
   - Wire up in `cmd/server/main.go`

### What to Keep

- Directory structure (`cmd/`, `internal/adapters/`, `internal/domain/`)
- Hexagonal architecture pattern
- Bounded context organization
- Event-driven communication pattern
- `cloud-native-utils` as infrastructure library
- `context.Context` threading through all layers

### What to Customize

- Bounded contexts (replace `reservation/`, `payment/`, `orchestration/` with your domains)
- Shared kernel types in `internal/domain/shared/`
- Static assets and templates in `cmd/server/assets/`
- PostgreSQL schemas in `migrations/` (uses simple key/value pattern)
- Environment configuration in `.env`
- Docker Compose services as needed
- Swap mock adapters for real implementations

---

## Baseline deviations

This project follows the [engineering baseline](https://github.com/andygeiss/baseline).
Where it does not, the reason is here. A reader hunting for gaps can count every bullet
in this section as one.

### Waived rules

Each entry states the rule, the document, the date, who decided, why, and what contains
the damage.

- **Persistence: SQLite first**
  ([project-types/web-application.md](https://github.com/andygeiss/baseline/blob/main/project-types/web-application.md))
  — waived 2026-09-07 by Andy. One database per bounded context is the thing this
  repository demonstrates, and a single SQLite file cannot show that isolation.
  Contained: `pgx/v5` is the only driver, both schemas are the same key/value shape in
  `migrations/`, and no query ever crosses a context.

- **Sessions via `alexedwards/scs/v2`**
  ([project-types/web-application.md](https://github.com/andygeiss/baseline/blob/main/project-types/web-application.md))
  — waived 2026-09-07 by Andy. The project exists partly to show OAuth 2.1 against a
  real identity provider, including two clients: a browser session for the web UI and a
  Bearer token for MCP. Contained: sessions stay server-side in
  `cloud-native-utils/web`, the cookie carries a token and nothing else, and Keycloak is
  reached only from `cmd/server/main.go` and the OIDC middleware.

- **Project layout**
  ([patterns/go-project-layout.md](https://github.com/andygeiss/baseline/blob/main/patterns/go-project-layout.md))
  — waived 2026-09-07 by Andy. The baseline's `internal/{app,domain,store}` has one
  slot for one context; this repository teaches several bounded contexts, so it uses
  `internal/domain/<context>` with `internal/adapters/{inbound,outbound}`. Contained:
  the dependency direction is unchanged — adapters import domain, domain imports no
  adapter — and routes still live in one file.

- **Approved dependency list**
  ([stack/go.md](https://github.com/andygeiss/baseline/blob/main/stack/go.md))
  — waived 2026-09-07 by Andy. Three dependencies sit outside the list, each because it
  *is* the mechanism a chapter of this reference implementation demonstrates:
  `cloud-native-utils` (server, sessions, MCP, templating, Kafka dispatcher),
  `go-oidc/v3` (token verification), and `kafka-go` (event streaming, pulled in
  indirectly). Contained: every one is reached through a port in `internal/domain`, so
  the domain layer compiles without any of them.

- **No service worker**
  ([patterns/pwa.md](https://github.com/andygeiss/baseline/blob/main/patterns/pwa.md))
  — waived 2026-09-07 by Andy, temporarily. Nothing registers a service worker any
  more, and `GET /sw.js` now serves a tombstone: it takes over from the old worker,
  deletes every cache the old one filled, unregisters itself, and reloads the pages it
  controlled. It handles no fetch events, so nothing is served from the client.
  Contained: no template references it, a test pins that no fetch handler comes back,
  and the waiver ends when the route is deleted after the deprecation window
  (tracked in [CLAUDE.md § Roadmap](./CLAUDE.md#roadmap)).

- **Secrets arrive as files, never as environment variables**
  ([patterns/go-config.md](https://github.com/andygeiss/baseline/blob/main/patterns/go-config.md)
  *Secrets* — **tier 1, so this is an open task and not a real waiver**) — recorded
  2026-09-07 by Andy. The binary already prefers one file per secret in
  `$CREDENTIALS_DIRECTORY`, and `Config.LogValue` keeps every password out of the logs.
  What is missing is the other half: `docker-compose.yml` still passes the two database
  passwords as environment variables, so the fallback is still the path in use.
  Contained: the fallback is two fields, read in one function, and finishing it is a
  deployment change rather than a code change.

### Rules met by a different route

- **Health checks.** The baseline puts `/healthz` on a loopback-only ops listener, and
  this project does that (`127.0.0.1:6060`, together with `/debug/pprof`). The app
  listener additionally answers `/health`, `/liveness`, and `/readiness` for the
  container orchestrator. Those three return a status code and nothing else, so the
  reason `/healthz` stays private — it names the build and reaches the databases — does
  not apply to them.

- **Deployment shape.** The baseline ships one static binary behind a TLS proxy. This
  one is built `CGO_ENABLED=0` and runs alone in a `scratch` image, so it is that same
  binary; the container is how it is delivered, and belongs to the operations
  repository rather than here.

---

## Contributing

1. Ensure `make check` is green, and `make ci` before you push
2. Fix formatting with `make fmt` and read the `go fix` diff before committing it
3. Follow hexagonal architecture patterns (ports in domain, adapters in adapters/)
4. Maintain bounded context boundaries (communicate via events, not direct calls)
5. Update documentation if architecture changes

---

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
