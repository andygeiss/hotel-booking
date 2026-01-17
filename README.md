<p align="center">
<img src="https://github.com/andygeiss/hotel-booking/blob/main/cmd/server/assets/static/img/icon-192.png?raw=true" width="100"/>
</p>

# Hotel Booking

[![Go Reference](https://pkg.go.dev/badge/github.com/andygeiss/hotel-booking.svg)](https://pkg.go.dev/github.com/andygeiss/hotel-booking)
[![License](https://img.shields.io/github/license/andygeiss/hotel-booking)](https://github.com/andygeiss/hotel-booking/blob/master/LICENSE)
[![Releases](https://img.shields.io/github/v/release/andygeiss/hotel-booking)](https://github.com/andygeiss/hotel-booking/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/andygeiss/hotel-booking)](https://goreportcard.com/report/github.com/andygeiss/hotel-booking)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/f9f01632dff14c448dbd4688abbd04e8)](https://app.codacy.com/gh/andygeiss/hotel-booking/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade)
[![Codacy Badge](https://app.codacy.com/project/badge/Coverage/f9f01632dff14c448dbd4688abbd04e8)](https://app.codacy.com/gh/andygeiss/hotel-booking/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_coverage)

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
- Persist data with **PostgreSQL**

---

## Key Features

- **Bounded Context Architecture** — Separate reservation, payment, and orchestration contexts
- **Developer Experience** — `just` task runner, golangci-lint, comprehensive test coverage
- **Domain-Driven Design** — Aggregates, entities, value objects, domain services, and domain events
- **Event-Driven Communication** — Kafka-based pub/sub for inter-context messaging
- **Hexagonal Architecture** — Clear separation between domain logic and infrastructure
- **OIDC Authentication** — Keycloak integration with session management
- **PostgreSQL Persistence** — Production-ready database with proper schema
- **Production-Ready Docker** — Multi-stage build with PGO optimization
- **Progressive Web App** — Service worker, manifest, and offline support
- **Saga Pattern** — Event-driven booking workflow with compensation on failure

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

| Context | Purpose | Key Aggregates |
|---------|---------|----------------|
| **Reservation** | Room booking lifecycle | `Reservation` |
| **Payment** | Payment processing | `Payment` |
| **Orchestration** | Cross-context coordination | Saga coordination |

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
├── .justfile                     # Task runner commands
├── cmd/server/                   # HTTP server entry point
│   ├── main.go                   # DI wiring, bootstrap, lifecycle
│   └── assets/
│       ├── static/               # CSS, JS, images (embedded)
│       └── templates/            # HTML templates (*.tmpl, embedded)
│           └── error.tmpl        # User-friendly error page
├── docker-compose.yml            # Dev stack (PostgreSQL, Keycloak, Kafka, app)
├── Dockerfile                    # Multi-stage production build
├── migrations/
│   └── init.sql                  # PostgreSQL schema
├── internal/
│   ├── adapters/
│   │   ├── inbound/              # HTTP handlers, event subscribers
│   │   │   ├── router.go         # HTTP routing & middleware
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
│       │   └── service.go        # ReservationService
│       ├── payment/              # Payment bounded context
│       │   ├── aggregate.go      # Payment aggregate + status
│       │   ├── entities.go       # PaymentAttempt
│       │   ├── events.go         # Domain events
│       │   ├── ports.go          # Interface definitions
│       │   └── service.go        # PaymentService
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

- **Docker** and **Docker Compose** (or Podman)
- **Go 1.24+**
- **golangci-lint** (for linting/formatting)
- **just** task runner

### Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/andygeiss/hotel-booking.git
   cd hotel-booking
   ```

2. **Install development tools:**
   ```bash
   just setup
   ```
   This installs `docker-compose`, `golangci-lint`, `just`, and `podman` via Homebrew.

3. **Configure environment:**
   ```bash
   cp .env.example .env
   cp .keycloak.json.example .keycloak.json
   ```

4. **Start the development stack:**
   ```bash
   just up
   ```
   This builds the Docker image and starts PostgreSQL, Keycloak, Kafka, and the application.

5. **Access the application:**
   - **App:** http://localhost:8080/ui
   - **Keycloak Admin:** http://localhost:8180/admin (admin:admin)

---

## Usage

### Commands

| Command | Description |
|---------|-------------|
| `just build` | Build Docker image |
| `just down` | Stop all services |
| `just fmt` | Format code |
| `just lint` | Run linter |
| `just profile` | Generate CPU profile for PGO |
| `just setup` | Install development dependencies |
| `just test` | Run unit tests with coverage |
| `just test-integration` | Run integration tests |
| `just up` | Start full development stack |

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

---

## Testing

### Unit Tests

Run all unit tests with coverage:

```bash
just test
```

This generates `.coverage.pprof` with coverage metrics.

### Integration Tests

Integration tests require external services (PostgreSQL, Kafka, Keycloak):

```bash
just test-integration
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

Configuration is managed via environment variables. Copy `.env.example` to `.env` and customize:

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_NAME` | Display name for UI and PWA | `Hotel Booking` |
| `APP_DESCRIPTION` | Application description | `Hotel reservation and payment management system` |
| `APP_SHORTNAME` | Docker image/container name | `hotel-booking` |
| `KAFKA_BROKERS` | Kafka broker addresses | `localhost:9092` |
| `OIDC_CLIENT_ID` | OIDC client ID | `hotel-booking` |
| `OIDC_CLIENT_SECRET` | OIDC client secret | Auto-generated |
| `OIDC_ISSUER` | Keycloak realm URL | `http://localhost:8180/realms/local` |
| `PORT` | HTTP server port | `8080` |
| `POSTGRES_HOST` | PostgreSQL host | `localhost` |
| `POSTGRES_PORT` | PostgreSQL port | `5432` |
| `POSTGRES_USER` | PostgreSQL user | `booking` |
| `POSTGRES_PASSWORD` | PostgreSQL password | `booking_secret` |
| `POSTGRES_DB` | PostgreSQL database | `booking_db` |
| `POSTGRES_SSLMODE` | SSL mode | `disable` |

See `.env.example` for the complete list with documentation.

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
- PostgreSQL schema in `migrations/`
- Environment configuration in `.env`
- Docker Compose services as needed
- Swap mock adapters for real implementations

---

## Contributing

1. Ensure all tests pass: `just test`
2. Ensure code is formatted and linted: `just fmt && just lint`
3. Follow hexagonal architecture patterns (ports in domain, adapters in adapters/)
4. Maintain bounded context boundaries (communicate via events, not direct calls)
5. Update documentation if architecture changes

---

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
