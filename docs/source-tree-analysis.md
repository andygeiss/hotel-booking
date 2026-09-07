# Source Tree Analysis

Annotated directory map of the Hotel Booking repository. Every critical folder is labeled with its purpose and the hexagonal/DDD role it plays.

**Repository type:** monolith
**Primary entry point:** `cmd/server/main.go`
**Embedded assets:** `cmd/server/assets/` via `//go:embed assets`

---

## Top-Level Layout

```
hotel-booking/
├── .env.example                      # Template for local env vars (copy to .env)
├── .keycloak.json.example            # Template for Keycloak realm config
├── .gitignore                        # Includes .claude, _bmad, _bmad-output (BMAD workspace)
├── CLAUDE.md                         # Project conventions, ubiquitous language, state machines, gotchas
├── Dockerfile                        # Multi-stage: golang:1.27.1-alpine3.23 → scratch, PGO enabled
├── LICENSE                           # MIT
├── Makefile                          # The only command surface; the eight gates live in `check`
├── README.md                         # User-facing quick start + architecture overview
├── SPEC.md                           # The project brief: job, why, guardrails, done means
├── docker-compose.yml                # Dev stack: app + keycloak + kafka + 2× postgres
├── go.mod / go.sum                   # Module: github.com/andygeiss/hotel-booking, `go 1.27`

There is deliberately no `.github/workflows/`, no dependency bot, and no
`.golangci.yml`: the gates run from the Makefile on the developer's machine, and
staticcheck is the only third-party lint tool.
│
├── cmd/server/                       # ┌─ Binary entry point
│   ├── config.go                     # │  Config struct + parseConfig: flags over env over
│   │                                 # │  defaults, validated before anything opens
│   ├── default.pgo                   # │  Committed CPU profile; go build finds it by name
│   ├── main.go                       # │  DI wiring, context, Postgres pools, Kafka dispatcher,
│   │                                 # │  OIDC provider, MCP server, HTTP server lifecycle
│   ├── main_test.go                  # │  Integration benchmarks → cmd/server/default.pgo
│   └── assets/                       # │  Embedded via //go:embed assets
│       ├── static/                   # │  CSS, JS, images (served at /static)
│       └── templates/                # │  Go html/template .tmpl files
│           ├── error.tmpl            # │  /ui/error — user-friendly error page
│           ├── index.tmpl            # │  /ui/ — authenticated dashboard
│           ├── login.tmpl            # │  /ui/login — OIDC login handoff
│           ├── manifest.tmpl         # │  /manifest.json — PWA manifest
│           ├── reservation_detail.tmpl
│           ├── reservation_form.tmpl
│           └── reservations.tmpl
│                                     # └─
│
├── docs/                             # Architecture and generated documentation
│   ├── ARCHITECTURE.md               # Hand-written deep-dive architecture doc (canonical)
│   ├── project-overview.md           # Generated
│   ├── source-tree-analysis.md       # Generated — this file
│   ├── api-contracts.md              # Generated — HTTP + MCP endpoint catalog
│   ├── data-models.md                # Generated — aggregates + KV schema
│   ├── development-guide.md          # Generated
│   ├── deployment-guide.md           # Generated
│   ├── component-inventory.md        # Generated
│   ├── index.md                      # Generated — master navigation
│   └── project-scan-report.json      # Workflow state file (resume marker)
│
├── migrations/                       # Per-database schema init (runs on first container start)
│   ├── payment/init.sql              # payment_db — creates kv_store table + key index
│   └── reservation/init.sql          # reservation_db — creates kv_store table + key index
│
└── internal/                         # ┌─ Hexagonal application code (non-public)
    ├── adapters/
    │   ├── inbound/                  # │  Driving adapters — things that call INTO the domain
    │   │   ├── router.go             # │  RouterConfig struct + Route() builds all /ui/* +
    │   │   │                         # │  /manifest.json + optional /mcp endpoint
    │   │   ├── render.go             # │  View: html/template sets, isFragment, Vary
    │   │   ├── http_index.go         # │  GET /ui/ (authenticated dashboard)
    │   │   ├── http_login.go         # │  GET /ui/login (OIDC handoff)
    │   │   ├── http_error.go         # │  GET /ui/error (error page with query params)
    │   │   ├── http_manifest.go      # │  GET /manifest.json (PWA)
    │   │   ├── http_booking_reservations.go     # │  GET /ui/reservations (list by guest email)
    │   │   ├── http_booking_reservation_form.go # │  GET /ui/reservations/new + POST /ui/reservations
    │   │   ├── http_booking_reservation_detail.go # │  GET/POST /ui/reservations/{id}[/cancel]
    │   │   ├── event_subscriber.go   # │  Generic Kafka topic→handler subscriber
    │   │   ├── *_test.go             # │  Handler + subscriber unit tests
    │   │   └── testdata/             # │  Template fixtures for handler tests
    │   │
    │   └── outbound/                 # │  Driven adapters — things the domain calls OUT to
    │       ├── event_publisher.go    # │  Publishes domain events to Kafka via messaging.Dispatcher
    │       ├── repository_availability_checker.go # │  AvailabilityChecker → reservation repo scan
    │       ├── mock_notification_service.go # │  Logs notifications (stand-in for email)
    │       ├── mock_payment_gateway.go      # │  Simulated gateway with configurable failure rate
    │       └── *_test.go
    │
    └── domain/                       # │  Business logic, no infrastructure concerns
        ├── shared/                   # │  ┌─ Shared kernel (cross-context types)
        │   └── types.go              # │  │  ReservationID, Money (amount + currency)
        │                             # │  └─
        │
        ├── reservation/              # │  ┌─ Reservation bounded context
        │   ├── aggregate.go          # │  │  Reservation aggregate + state machine + validation
        │   ├── entities.go           # │  │  DateRange + GuestInfo value objects
        │   ├── events.go             # │  │  EventCreated/Confirmed/Activated/Completed/Cancelled +
        │   │                         # │  │  topic constants (reservation.* )
        │   ├── ports.go              # │  │  ReservationRepository, AvailabilityChecker, EventPublisher
        │   ├── service.go            # │  │  Reservation service (use cases)
        │   ├── tools.go              # │  │  MCP tool registration (get_/list_/cancel_/check_availability)
        │   └── *_test.go             # │  └─
        │
        ├── payment/                  # │  ┌─ Payment bounded context
        │   ├── aggregate.go          # │  │  Payment aggregate + Authorize/Capture/Refund/Fail + retry
        │   ├── entities.go           # │  │  PaymentAttempt entity (attempt history)
        │   ├── events.go             # │  │  EventAuthorized/Captured/Failed/Refunded + topics (payment.*)
        │   ├── ports.go              # │  │  PaymentRepository, PaymentGateway, EventPublisher
        │   ├── service.go            # │  │  Payment service with gateway + failure publishing
        │   ├── tools.go              # │  │  MCP tool registration (get_/capture_/refund_payment)
        │   └── *_test.go             # │  └─
        │
        └── orchestration/            # │  ┌─ Saga / cross-context coordinator
            ├── booking_service.go    # │  │  BookingService: InitiateBooking (event-driven) +
            │                         # │  │  CompleteBooking (sync) + OnPayment* callbacks
            ├── event_handlers.go     # │  │  Subscribes to reservation.created + payment.*,
            │                         # │  │  wires up the full saga with compensation
            ├── ports.go              # │  │  NotificationService interface
            └── *_test.go             # │  └─
```

---

## Critical Folders

| Folder | Purpose | Entry Points |
|--------|---------|--------------|
| `cmd/server/` | Binary entry + DI wiring | `main.go` |
| `cmd/server/assets/templates/` | HTML templates (SSR via Go `html/template`) | embedded into binary |
| `internal/adapters/inbound/` | HTTP handlers, Kafka subscriber, router | `router.go` (`Route()`) |
| `internal/adapters/outbound/` | DB repos, event publisher, gateway mocks | — |
| `internal/domain/reservation/` | Reservation bounded context | `service.go`, `aggregate.go` |
| `internal/domain/payment/` | Payment bounded context | `service.go`, `aggregate.go` |
| `internal/domain/orchestration/` | Saga coordinator + event handlers | `booking_service.go`, `event_handlers.go` |
| `internal/domain/shared/` | Shared kernel (Money, ReservationID) | `types.go` |
| `migrations/` | Per-context DB init scripts | `*/init.sql` |

## Test Organization

- Unit tests colocated with source files (`*_test.go`).
- Integration tests are tagged with `//go:build integration` (run via `go test -tags=integration -v ./internal/...`).
- Benchmarks (`cmd/server/main_test.go`) drive PGO profile generation (`make profile`).
- Test fixtures live under `internal/adapters/inbound/testdata/assets/templates/` (mirrors embedded templates).

## Dependency Direction

```
cmd/server/main.go
        ↓
internal/adapters/{inbound, outbound}
        ↓
internal/domain/{reservation, payment, orchestration}
        ↓
internal/domain/shared  +  github.com/andygeiss/cloud-native-utils
```

Domain packages import only `shared` and port types from `cloud-native-utils` (`event.EventPublisher`, `resource.Access[K,V]`, `mcp.Server` for tool registration). They do **not** import adapters.
