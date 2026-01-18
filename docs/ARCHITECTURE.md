# Hotel Booking System - Architecture Documentation

A hotel reservation and payment management system built with Go, Hexagonal Architecture, and Domain-Driven Design (DDD).

## Table of Contents

1. [Overview](#overview)
2. [Technology Stack](#technology-stack)
3. [Architecture Principles](#architecture-principles)
4. [Project Structure](#project-structure)
5. [Bounded Contexts](#bounded-contexts)
6. [Domain Layer](#domain-layer)
7. [Application Layer (Orchestration)](#application-layer-orchestration)
8. [Adapter Layer](#adapter-layer)
9. [Event-Driven Communication](#event-driven-communication)
10. [Saga Pattern Implementation](#saga-pattern-implementation)
11. [Database Design](#database-design)
12. [API Design](#api-design)
13. [Testing Strategy](#testing-strategy)
14. [Conventions and Patterns](#conventions-and-patterns)
15. [Recipes](#recipes)
16. [Infrastructure](#infrastructure)
17. [Security](#security)
18. [Deployment](#deployment)

---

## Overview

This system manages hotel room reservations and associated payments. It demonstrates:

- **Hexagonal Architecture** (Ports & Adapters) for clean separation of concerns
- **Domain-Driven Design** with bounded contexts and aggregates
- **Event-Driven Architecture** using Kafka for cross-context communication
- **Saga Pattern** for distributed transaction coordination
- **Database-per-Context** for true microservice isolation

### System Capabilities

- Create, view, and cancel room reservations
- Authorization-Capture payment processing workflow
- Automatic payment processing triggered by domain events
- Compensation logic for handling failures
- PWA support for mobile-first experience
- MCP (Model Context Protocol) endpoint for AI tool integration

---

## Technology Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| Language | Go 1.25.5 | Application runtime |
| Architecture | Hexagonal + DDD | Clean separation, domain focus |
| Web Framework | `net/http` stdlib | HTTP server and routing |
| Frontend | HTML + HTMX | Interactive UI without JavaScript frameworks |
| Authentication | Keycloak (OIDC) | OAuth2/OpenID Connect provider |
| Database | PostgreSQL 16 | Persistent storage (one per context) |
| Message Broker | Apache Kafka | Event streaming between contexts |
| Containerization | Docker/Podman | Deployment and local development |
| Task Runner | Just | Build automation |

### Core Dependencies

```go
require (
    github.com/andygeiss/cloud-native-utils v0.5.5  // Logging, messaging, web, templating, MCP
    github.com/jackc/pgx/v5 v5.8.0                  // PostgreSQL driver
)
```

### Indirect Dependencies

- `coreos/go-oidc/v3` - OIDC client for Keycloak integration
- `segmentio/kafka-go` - Kafka client for event messaging
- `klauspost/compress`, `pierrec/lz4` - Compression for Kafka

---

## Architecture Principles

### Hexagonal Architecture (Ports & Adapters)

The architecture enforces strict layering with dependency inversion:

```
┌─────────────────────────────────────────────────────────────┐
│                     Inbound Adapters                        │
│           (HTTP Handlers, Event Subscribers)                │
└─────────────────────────────┬───────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      Inbound Ports                          │
│              (Interfaces defined in Domain)                 │
└─────────────────────────────┬───────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      Domain Layer                           │
│    (Aggregates, Entities, Value Objects, Domain Events)     │
│                    ZERO INFRASTRUCTURE IMPORTS              │
└─────────────────────────────┬───────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     Outbound Ports                          │
│              (Interfaces defined in Domain)                 │
└─────────────────────────────┬───────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Outbound Adapters                        │
│      (Repository Implementations, Payment Gateway,          │
│       Event Publishers, Notification Services)              │
└─────────────────────────────────────────────────────────────┘
```

### Key Architectural Rules

1. **Domain has no infrastructure dependencies** - The domain layer imports only the standard library and the shared kernel
2. **Dependencies point inward** - Adapters depend on domain, never the reverse
3. **Interfaces are defined by consumers** - Ports are defined in the domain, implemented in adapters
4. **One aggregate per transaction** - Each repository operation affects only one aggregate

---

## Project Structure

```
hotel-booking/
├── cmd/
│   └── server/
│       ├── main.go                 # Application entry point, DI wiring
│       └── assets/
│           ├── static/             # CSS, JS (HTMX), images
│           └── templates/          # HTML templates (*.tmpl)
├── internal/
│   ├── adapters/
│   │   ├── inbound/                # HTTP handlers, event subscribers
│   │   │   ├── router.go           # HTTP route definitions
│   │   │   ├── http_*.go           # HTTP handler implementations
│   │   │   └── event_subscriber.go # Event subscription adapter
│   │   └── outbound/               # Repository implementations, gateways
│   │       ├── event_publisher.go
│   │       ├── repository_availability_checker.go
│   │       ├── mock_payment_gateway.go
│   │       └── mock_notification_service.go
│   └── domain/
│       ├── shared/                 # Shared Kernel
│       │   └── types.go            # ReservationID, Money
│       ├── reservation/            # Reservation Bounded Context
│       │   ├── aggregate.go        # Reservation aggregate root
│       │   ├── entities.go         # DateRange, GuestInfo
│       │   ├── ports.go            # Repository, AvailabilityChecker interfaces
│       │   ├── events.go           # Domain events
│       │   └── service.go          # Application service
│       ├── payment/                # Payment Bounded Context
│       │   ├── aggregate.go        # Payment aggregate root
│       │   ├── entities.go         # PaymentAttempt
│       │   ├── ports.go            # Repository, PaymentGateway interfaces
│       │   ├── events.go           # Domain events
│       │   └── service.go          # Application service
│       └── orchestration/          # Saga Coordination Layer
│           ├── ports.go            # NotificationService interface
│           ├── booking_service.go  # Booking workflow orchestration
│           └── event_handlers.go   # Cross-context event handlers
├── migrations/
│   ├── reservation/init.sql        # Reservation database schema
│   └── payment/init.sql            # Payment database schema
├── docker-compose.yml              # Development stack
├── Dockerfile                      # Production build
├── go.mod                          # Go module definition
└── .justfile                       # Task runner commands
```

---

## Bounded Contexts

The system is divided into three bounded contexts, each with clear responsibilities:

### 1. Reservation Context

**Purpose:** Manages the complete reservation lifecycle

**Aggregate Root:** `Reservation`

**Responsibilities:**
- Room availability checking
- Reservation creation with validation
- State transitions (pending → confirmed → active → completed)
- Cancellation with business rules
- Guest information management

**Database:** `reservation_db` (port 5432)

### 2. Payment Context

**Purpose:** Handles all payment processing

**Aggregate Root:** `Payment`

**Responsibilities:**
- Payment authorization (two-phase commit)
- Payment capture
- Refund processing
- Payment attempt tracking
- Transaction ID management

**Database:** `payment_db` (port 5433)

### 3. Orchestration Context

**Purpose:** Coordinates workflows across bounded contexts

**Key Components:** `BookingService`, `EventHandlers`

**Responsibilities:**
- Booking saga coordination
- Event subscription and routing
- Compensation logic on failures
- Notification triggering

**Database:** None (stateless coordinator)

### Shared Kernel

Types shared across contexts without violating context boundaries:

```go
// internal/domain/shared/types.go

// ReservationID - shared because Payment needs to reference reservations
type ReservationID string

// Money - shared because both contexts deal with monetary values
type Money struct {
    Currency string // ISO 4217 (e.g., "USD")
    Amount   int64  // Amount in cents
}
```

---

## Domain Layer

### Aggregates

#### Reservation Aggregate

```go
// internal/domain/reservation/aggregate.go

type Reservation struct {
    ID                 ReservationID
    GuestID            GuestID
    RoomID             RoomID
    DateRange          DateRange          // Value Object
    Status             ReservationStatus  // State machine
    TotalAmount        Money              // Shared Kernel
    CancellationReason string
    CreatedAt          time.Time
    UpdatedAt          time.Time
    Guests             []GuestInfo        // Embedded entities
}
```

**State Machine:**

```
┌─────────┐     Confirm()      ┌───────────┐     Activate()     ┌────────┐     Complete()    ┌───────────┐
│ pending │──────────────────► │ confirmed │───────────────────►│ active │──────────────────►│ completed │
└────┬────┘                    └─────┬─────┘                    └────────┘                   └───────────┘
     │                               │
     │ Cancel()                      │ Cancel()
     ▼                               ▼
┌───────────┐                  ┌───────────┐
│ cancelled │◄─────────────────│ cancelled │
└───────────┘                  └───────────┘
```

**Business Rules:**
- Minimum 1 night stay required
- Check-in must be in the future
- Cannot cancel within 24 hours of check-in
- At least one guest required
- Cancelled reservations do not block availability

#### Payment Aggregate

```go
// internal/domain/payment/aggregate.go

type Payment struct {
    ID            PaymentID
    ReservationID ReservationID      // Cross-context reference (not FK)
    Amount        Money
    Status        PaymentStatus
    PaymentMethod string
    TransactionID string             // External gateway reference
    CreatedAt     time.Time
    UpdatedAt     time.Time
    Attempts      []PaymentAttempt   // Embedded entities
}
```

**State Machine (Authorization-Capture Pattern):**

```
┌─────────┐    Authorize()    ┌────────────┐    Capture()    ┌──────────┐
│ pending │──────────────────►│ authorized │────────────────►│ captured │
└────┬────┘                   └─────┬──────┘                 └─────┬────┘
     │                              │                              │
     │ Fail()                       │ Fail()                       │ Refund()
     ▼                              ▼                              ▼
┌────────┐                    ┌────────┐                     ┌──────────┐
│ failed │◄───────────────────│ failed │                     │ refunded │
└────────┘                    └────────┘                     └──────────┘
```

**Business Rules:**
- Authorization required before capture
- Only captured payments can be refunded
- Maximum 3 retry attempts for failed payments

### Value Objects

```go
// DateRange - immutable date period
type DateRange struct {
    CheckIn  time.Time
    CheckOut time.Time
}

// GuestInfo - guest details within reservation
type GuestInfo struct {
    Name        string
    Email       string
    PhoneNumber string
}

// PaymentAttempt - payment processing history
type PaymentAttempt struct {
    AttemptedAt time.Time
    Status      PaymentStatus
    ErrorCode   string
    ErrorMsg    string
}
```

### Strongly-Typed Identifiers

All entity identifiers use type aliases to prevent accidental mixing:

```go
type ReservationID = shared.ReservationID  // Shared
type GuestID string                         // Local to reservation context
type RoomID string                          // Local to reservation context
type PaymentID string                       // Local to payment context
```

### Sentinel Errors

Each context defines package-level error variables for type checking:

```go
// Reservation errors
var (
    ErrInvalidDateRange        = errors.New("check-out must be after check-in")
    ErrCheckInPast             = errors.New("check-in date must be in the future")
    ErrMinimumStay             = errors.New("minimum stay is 1 night")
    ErrInvalidStateTransition  = errors.New("invalid state transition")
    ErrCannotCancelNearCheckIn = errors.New("cannot cancel within 24 hours of check-in")
    ErrNoGuests                = errors.New("at least one guest required")
)

// Payment errors
var (
    ErrInvalidPaymentTransition = errors.New("invalid payment state transition")
    ErrNotAuthorized            = errors.New("payment not authorized")
    ErrAlreadyCaptured          = errors.New("payment already captured")
    ErrCannotRefund             = errors.New("can only refund captured payments")
)
```

---

## Application Layer (Orchestration)

### Domain Services

Each bounded context has a service that orchestrates domain operations:

```go
// internal/domain/reservation/service.go

type Service struct {
    reservationRepo     ReservationRepository
    availabilityChecker AvailabilityChecker
    publisher           event.EventPublisher
}

func (s *Service) CreateReservation(ctx context.Context, ...) (*Reservation, error) {
    // 1. Check room availability
    // 2. Create aggregate with validation
    // 3. Persist to repository
    // 4. Publish domain event
}
```

### Booking Service (Saga Coordinator)

The `BookingService` orchestrates the complete booking workflow:

```go
// internal/domain/orchestration/booking_service.go

type BookingService struct {
    reservationService  *reservation.Service
    paymentService      *payment.Service
    notificationService NotificationService
}
```

**Key Methods:**

| Method | Purpose |
|--------|---------|
| `InitiateBooking` | Creates reservation, triggers event-driven payment flow |
| `CompleteBooking` | Synchronous booking with all steps |
| `OnPaymentAuthorized` | Handles payment.authorized event |
| `OnPaymentCaptured` | Confirms reservation on successful payment |
| `OnPaymentFailed` | Cancels reservation as compensation |
| `CancelBookingWithRefund` | Cancels reservation and processes refund |

---

## Adapter Layer

### Inbound Adapters

#### HTTP Handlers

Handler factory pattern with dependency injection:

```go
// internal/adapters/inbound/http_booking_reservations.go

func HttpViewReservations(e *templating.Engine, reservationService *reservation.Service) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()

        // Extract authenticated user from context
        sessionID, _ := ctx.Value(web.ContextSessionID).(string)
        email, _ := ctx.Value(web.ContextEmail).(string)

        // Use domain service
        reservations, err := reservationService.ListReservationsByGuest(ctx, reservation.GuestID(email))

        // Render response
        HttpView(e, "reservations", data)(w, r)
    }
}
```

#### Event Subscriber

Subscribes to Kafka topics and routes to domain handlers:

```go
// internal/adapters/inbound/event_subscriber.go

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

### Outbound Adapters

#### Event Publisher

Publishes domain events to Kafka:

```go
// internal/adapters/outbound/event_publisher.go

func (ep *EventPublisher) Publish(ctx context.Context, e event.Event) error {
    encoded, _ := json.Marshal(e)
    msg := messaging.NewMessage(e.Topic(), encoded)
    return ep.dispatcher.Publish(ctx, msg)
}
```

#### Repository Availability Checker

Implements `AvailabilityChecker` port using the repository:

```go
// internal/adapters/outbound/repository_availability_checker.go

func (c *RepositoryAvailabilityChecker) IsRoomAvailable(ctx context.Context, roomID RoomID, dateRange DateRange) (bool, error) {
    overlapping, _ := c.GetOverlappingReservations(ctx, roomID, dateRange)
    return len(overlapping) == 0, nil
}
```

#### Mock Payment Gateway

Simulates external payment gateway for testing:

```go
// internal/adapters/outbound/mock_payment_gateway.go

type MockPaymentGateway struct {
    transactions map[string]shared.Money
    FailureRate  float64  // 0.0 to 1.0
    ShouldFail   bool     // Force failures for testing
}

func (g *MockPaymentGateway) Authorize(ctx context.Context, pay *Payment) (string, error) {
    if g.ShouldFail || cryptoRandFloat64() < g.FailureRate {
        return "", errors.New("payment authorization failed")
    }
    transactionID := fmt.Sprintf("txn_%s_%d", pay.ID, pay.Amount.Amount)
    g.transactions[transactionID] = pay.Amount
    return transactionID, nil
}
```

---

## Event-Driven Communication

### Domain Events

Events use a fluent builder pattern:

```go
// internal/domain/reservation/events.go

type EventCreated struct {
    ReservationID ReservationID `json:"reservation_id"`
    GuestID       GuestID       `json:"guest_id"`
    RoomID        RoomID        `json:"room_id"`
    CheckIn       time.Time     `json:"check_in"`
    CheckOut      time.Time     `json:"check_out"`
    TotalAmount   Money         `json:"total_amount"`
}

func (e *EventCreated) Topic() string { return EventTopicCreated }

// Fluent builder
func NewEventCreated() *EventCreated { return &EventCreated{} }

func (e *EventCreated) WithReservationID(id ReservationID) *EventCreated {
    e.ReservationID = id
    return e
}
```

### Event Topics

| Context | Topic | Trigger |
|---------|-------|---------|
| Reservation | `reservation.created` | New reservation created |
| Reservation | `reservation.confirmed` | Payment captured |
| Reservation | `reservation.activated` | Guest checked in |
| Reservation | `reservation.completed` | Guest checked out |
| Reservation | `reservation.cancelled` | Reservation cancelled |
| Payment | `payment.authorized` | Payment authorization succeeded |
| Payment | `payment.captured` | Payment finalized |
| Payment | `payment.failed` | Payment processing failed |
| Payment | `payment.refunded` | Payment refunded |

### Event Flow

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                              BOOKING WORKFLOW                                           │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                         │
│  User                Reservation              Payment               Orchestration       │
│   │                     │                        │                       │              │
│   │  Create Booking     │                        │                       │              │
│   │────────────────────►│                        │                       │              │
│   │                     │                        │                       │              │
│   │                     │  reservation.created   │                       │              │
│   │                     │───────────────────────────────────────────────►│              │
│   │                     │                        │                       │              │
│   │                     │                        │◄──────────────────────│              │
│   │                     │                        │   AuthorizePayment    │              │
│   │                     │                        │                       │              │
│   │                     │                        │  payment.authorized   │              │
│   │                     │                        │──────────────────────►│              │
│   │                     │                        │                       │              │
│   │                     │                        │◄──────────────────────│              │
│   │                     │                        │    CapturePayment     │              │
│   │                     │                        │                       │              │
│   │                     │                        │  payment.captured     │              │
│   │                     │                        │──────────────────────►│              │
│   │                     │                        │                       │              │
│   │                     │◄───────────────────────────────────────────────│              │
│   │                     │        ConfirmReservation                      │              │
│   │                     │                        │                       │              │
│   │                     │  reservation.confirmed │                       │              │
│   │                     │───────────────────────────────────────────────►│              │
│   │                     │                        │                       │              │
│   │◄────────────────────│                        │    SendNotification   │              │
│   │   Booking Complete  │                        │                       │              │
│                                                                                         │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Saga Pattern Implementation

### Event Handlers Registration

```go
// internal/domain/orchestration/event_handlers.go

func (h *EventHandlers) RegisterHandlers(ctx context.Context, dispatcher messaging.Dispatcher) error {
    // Payment context subscribes to reservation.created
    dispatcher.Subscribe(ctx, reservation.EventTopicCreated, service.Wrap(h.handleReservationCreated))

    // Orchestration subscribes to payment.authorized
    dispatcher.Subscribe(ctx, payment.EventTopicAuthorized, service.Wrap(h.handlePaymentAuthorized))

    // Reservation context subscribes to payment.captured
    dispatcher.Subscribe(ctx, payment.EventTopicCaptured, service.Wrap(h.handlePaymentCaptured))

    // Orchestration subscribes to payment.failed for compensation
    dispatcher.Subscribe(ctx, payment.EventTopicFailed, service.Wrap(h.handlePaymentFailed))

    return nil
}
```

### Compensation Logic

When payment fails, the reservation is automatically cancelled:

```go
func (h *EventHandlers) handlePaymentFailed(msg messaging.Message) (messaging.MessageState, error) {
    var evt payment.EventFailed
    json.Unmarshal(msg.Data, &evt)

    // Compensation: cancel the reservation
    reason := fmt.Sprintf("payment_failed: %s - %s", evt.ErrorCode, evt.ErrorMsg)
    h.bookingService.OnPaymentFailed(ctx, evt.ReservationID, reason)

    return messaging.MessageStateCompleted, nil
}
```

### Saga Steps with Compensation

| Step | Action | Compensation on Failure |
|------|--------|------------------------|
| 1 | Create Reservation | N/A (first step) |
| 2 | Authorize Payment | Cancel Reservation |
| 3 | Capture Payment | Cancel Reservation |
| 4 | Confirm Reservation | Refund Payment, Cancel Reservation |
| 5 | Send Notification | Best effort (no compensation) |

---

## Database Design

### Database-per-Context Pattern

Each bounded context has its own PostgreSQL database:

| Context | Database | Port | Container |
|---------|----------|------|-----------|
| Reservation | `reservation_db` | 5432 | `postgres-reservation` |
| Payment | `payment_db` | 5433 | `postgres-payment` |

### Key/Value Storage Pattern

Both contexts use a simple key/value storage pattern via `PostgresAccess` from `cloud-native-utils`. Aggregates are serialized as JSON and stored in a generic `kv_store` table:

```sql
-- migrations/reservation/init.sql and migrations/payment/init.sql

CREATE TABLE IF NOT EXISTS kv_store (
    key TEXT PRIMARY KEY,
    value TEXT
);

CREATE INDEX IF NOT EXISTS idx_kv_store_key ON kv_store (key);
```

**How it works:**
- **Key:** The aggregate ID (e.g., `ReservationID` or `PaymentID`)
- **Value:** The entire aggregate serialized as JSON

This approach:
1. Simplifies schema management (no migrations needed for domain changes)
2. Aligns with DDD aggregate boundaries (one row = one aggregate)
3. Enables schema-less evolution of domain models

### Cross-Context References

The `Payment` aggregate contains a `ReservationID` field but this is **not** a database foreign key because:

1. Reservations and payments are in different databases
2. Referential integrity is maintained via domain events
3. Each context can scale independently

---

## API Design

### HTTP Routes

| Method | Path | Handler | Auth | Description |
|--------|------|---------|------|-------------|
| GET | `/ui/` | `HttpViewIndex` | Yes | Dashboard |
| GET | `/ui/login` | `HttpViewLogin` | No | OIDC login redirect |
| GET | `/ui/error` | `HttpViewError` | No | Error page |
| GET | `/ui/reservations` | `HttpViewReservations` | Yes | List reservations |
| GET | `/ui/reservations/new` | `HttpViewReservationForm` | Yes | New reservation form |
| POST | `/ui/reservations` | `HttpCreateReservation` | Yes | Create reservation |
| GET | `/ui/reservations/{id}` | `HttpViewReservationDetail` | Yes | Reservation detail |
| POST | `/ui/reservations/{id}/cancel` | `HttpCancelReservation` | Yes | Cancel reservation |
| GET | `/manifest.json` | `HttpViewManifest` | No | PWA manifest |
| GET | `/sw.js` | `HttpViewServiceWorker` | No | Service worker |
| POST | `/mcp` | `mcpHandler.Handler()` | No | MCP JSON-RPC endpoint |
| GET | `/liveness` | (built-in) | No | Health check |
| GET | `/readiness` | (built-in) | No | Readiness check |

### View Response Pattern

Each view handler defines its own response struct:

```go
type HttpViewReservationsResponse struct {
    AppName      string
    Title        string
    SessionID    string
    Reservations []ReservationListItem
}

type ReservationListItem struct {
    ID          string
    RoomID      string
    CheckIn     string
    CheckOut    string
    Status      string
    StatusClass string  // CSS class for status badge
    TotalAmount string
    CanCancel   bool
}
```

### HTMX Integration

Cancel operations support HTMX for partial page updates:

```go
if r.Header.Get("HX-Request") == "true" {
    w.Header().Set("HX-Redirect", "/ui/reservations")
    w.WriteHeader(http.StatusOK)
    return
}
http.Redirect(w, r, "/ui/reservations", http.StatusSeeOther)
```

### MCP Tools

The MCP (Model Context Protocol) endpoint exposes domain functionality to AI models. Tools are defined in each bounded context following the same pattern as `events.go` and `ports.go`.

**Architecture:**
```
internal/domain/reservation/
├── tools.go          # MCP tools for reservation operations

internal/domain/payment/
├── tools.go          # MCP tools for payment operations
```

**Tool Registration in `main.go`:**
```go
func buildMCPServer(
    reservationService *reservation.Service,
    availabilityChecker reservation.AvailabilityChecker,
    paymentService *payment.Service,
) *mcp.Server {
    server := mcp.NewServer(
        env.Get("APP_SHORTNAME", "mcp-server"),
        env.Get("APP_VERSION", "1.0.0"),
    )

    reservation.RegisterTools(server, reservationService, availabilityChecker)
    payment.RegisterTools(server, paymentService)

    return server
}
```

**Available Tools:**

| Tool | Context | Description |
|------|---------|-------------|
| `get_reservation` | Reservation | Get reservation details by ID |
| `list_reservations` | Reservation | List all reservations for a guest |
| `cancel_reservation` | Reservation | Cancel a reservation with reason |
| `check_availability` | Reservation | Check room availability for date range |
| `get_payment` | Payment | Get payment details by ID |
| `capture_payment` | Payment | Capture an authorized payment |
| `refund_payment` | Payment | Refund a captured payment |

**Tool Implementation Pattern:**
```go
func newGetReservationTool(service *Service) mcp.Tool {
    return mcp.NewTool(
        "get_reservation",
        "Get reservation details by ID.",
        mcp.NewObjectSchema(
            map[string]mcp.Property{
                "id": mcp.NewStringProperty("The reservation ID"),
            },
            []string{"id"},
        ),
        func(ctx context.Context, params mcp.ToolsCallParams) (mcp.ToolsCallResult, error) {
            id, _ := params.Arguments["id"].(string)
            reservation, err := service.GetReservation(ctx, ReservationID(id))
            if err != nil {
                return mcp.ToolsCallResult{}, err
            }
            data, _ := json.MarshalIndent(reservation, "", "  ")
            return mcp.ToolsCallResult{
                Content: []mcp.ContentBlock{mcp.NewTextContent(string(data))},
            }, nil
        },
    )
}
```

**Testing MCP Tools:**
```bash
# Initialize MCP session
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}'

# List available tools
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/list"}'

# Call a tool
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"check_availability","arguments":{"room_id":"room-101","check_in":"2024-06-01T14:00:00Z","check_out":"2024-06-05T11:00:00Z"}}}'
```

---

## Testing Strategy

### Test File Organization

```
*_test.go files located alongside source files:

internal/domain/reservation/
├── aggregate.go
├── aggregate_test.go      # Aggregate unit tests
├── service.go
├── service_test.go        # Service integration tests
├── tools.go
└── tools_test.go          # MCP tools tests

internal/domain/payment/
├── aggregate.go
├── aggregate_test.go      # Aggregate unit tests
├── service.go
├── service_test.go        # Service integration tests
├── tools.go
└── tools_test.go          # MCP tools tests

internal/adapters/inbound/
├── router.go
├── router_test.go         # Route registration tests
├── http_booking_*.go
└── http_booking_*_test.go # Handler tests
```

### Test Naming Convention

```go
// Pattern: Test_{Component}_{Scenario}_Should_{ExpectedResult}

func Test_Reservation_Confirm_From_Pending_Should_Succeed(t *testing.T)
func Test_Service_CreateReservation_When_Room_Unavailable_Should_Return_Error(t *testing.T)
func Test_Route_Liveness_Endpoint_Should_Return_200(t *testing.T)
```

### Mock Implementations

Tests use in-test mock implementations:

```go
// Arrange-Act-Assert pattern

type mockReservationRepository struct {
    reservations map[reservation.ReservationID]reservation.Reservation
    createErr    error
    readErr      error
}

func (m *mockReservationRepository) Create(ctx context.Context, id ReservationID, res Reservation) error {
    if m.createErr != nil {
        return m.createErr
    }
    m.reservations[id] = res
    return nil
}

type mockEventPublisher struct {
    published []event.Event
    err       error
}
```

### Test Categories

| Category | Location | Purpose |
|----------|----------|---------|
| Aggregate Tests | `domain/*/aggregate_test.go` | Business rule validation |
| Service Tests | `domain/*/service_test.go` | Workflow orchestration |
| Handler Tests | `adapters/inbound/*_test.go` | HTTP request/response |
| Adapter Tests | `adapters/outbound/*_test.go` | Infrastructure integration |
| Integration Tests | Separate test suite | End-to-end flows |

### Running Tests

```bash
just test                    # Run all unit tests with coverage
just test-integration        # Run integration tests (requires Docker)
go test -v -run TestName ./internal/domain/reservation/...  # Single test
```

---

## Conventions and Patterns

### Package Naming

- Domain packages use singular nouns: `reservation`, `payment`, `orchestration`
- Adapter packages use direction: `inbound`, `outbound`
- Shared code uses descriptive names: `shared`

### File Naming

| Pattern | Content |
|---------|---------|
| `aggregate.go` | Aggregate root definition |
| `entities.go` | Value objects and embedded entities |
| `ports.go` | Interface definitions (ports) |
| `events.go` | Domain event definitions |
| `service.go` | Application service |
| `*_test.go` | Test files |
| `http_*.go` | HTTP handlers |

### Error Handling

1. **Wrap errors with context:**
```go
if err := s.reservationRepo.Create(ctx, id, *reservation); err != nil {
    return nil, fmt.Errorf("failed to persist reservation: %w", err)
}
```

2. **Use sentinel errors for domain rules:**
```go
if r.Status == StatusCancelled {
    return ErrAlreadyCancelled
}
```

3. **State transition errors include current state:**
```go
return fmt.Errorf("%w: cannot confirm from %s", ErrInvalidStateTransition, r.Status)
```

### Context Propagation

All methods accept `context.Context` as the first parameter:

```go
func (s *Service) CreateReservation(ctx context.Context, ...) (*Reservation, error)
func (r ReservationRepository) Create(ctx context.Context, id ReservationID, res Reservation) error
```

### Dependency Injection

Dependencies are injected via constructors:

```go
func NewService(
    repo ReservationRepository,
    checker AvailabilityChecker,
    pub event.EventPublisher,
) *Service {
    return &Service{
        reservationRepo:     repo,
        availabilityChecker: checker,
        publisher:           pub,
    }
}
```

### Event Builder Pattern

Events use method chaining for readable construction:

```go
evt := NewEventCreated().
    WithReservationID(id).
    WithGuestID(guestID).
    WithRoomID(roomID).
    WithCheckIn(dateRange.CheckIn).
    WithCheckOut(dateRange.CheckOut).
    WithTotalAmount(amount)
```

---

## Recipes

### Adding a New Domain Event

1. Define the event struct in `domain/{context}/events.go`:

```go
type EventNewThing struct {
    ReservationID ReservationID `json:"reservation_id"`
    NewField      string        `json:"new_field"`
}

func NewEventNewThing() *EventNewThing {
    return &EventNewThing{}
}

func (e *EventNewThing) Topic() string { return "reservation.new_thing" }

func (e *EventNewThing) WithReservationID(id ReservationID) *EventNewThing {
    e.ReservationID = id
    return e
}
```

2. Publish from the service:

```go
evt := NewEventNewThing().WithReservationID(id).WithNewField("value")
s.publisher.Publish(ctx, evt)
```

3. Subscribe in orchestration if cross-context:

```go
dispatcher.Subscribe(ctx, "reservation.new_thing", service.Wrap(h.handleNewThing))
```

### Adding a New HTTP Endpoint

1. Create handler in `adapters/inbound/http_{feature}.go`:

```go
func HttpViewNewFeature(e *templating.Engine, service *reservation.Service) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Implementation
    }
}
```

2. Register route in `router.go`:

```go
mux.HandleFunc("GET /ui/new-feature", logging.WithLogging(logger, web.WithAuth(serverSessions, HttpViewNewFeature(e, reservationService))))
```

### Adding a New Bounded Context

1. Create domain package structure:

```
internal/domain/newcontext/
├── aggregate.go
├── entities.go
├── ports.go
├── events.go
├── service.go
└── tools.go          # MCP tools (optional)
```

2. Create database migration with key/value schema:

```sql
-- migrations/newcontext/init.sql
CREATE TABLE IF NOT EXISTS kv_store (
    key TEXT PRIMARY KEY,
    value TEXT
);

CREATE INDEX IF NOT EXISTS idx_kv_store_key ON kv_store (key);
```

3. Add to `docker-compose.yml`:

```yaml
postgres-newcontext:
  image: postgres:16-alpine
  environment:
    POSTGRES_DB: newcontext_db
  volumes:
    - ./migrations/newcontext/init.sql:/docker-entrypoint-initdb.d/init.sql:ro
```

4. Wire in `main.go`:

```go
newcontextDB, _ := sql.Open("pgx", newcontextDSN)
newcontextRepo := resource.NewPostgresAccess[newcontext.ID, newcontext.Aggregate](newcontextDB)
newcontextService := newcontext.NewService(newcontextRepo, publisher)
```

5. (Optional) Register MCP tools:

```go
// In tools.go
func RegisterTools(server *mcp.Server, service *Service) {
    server.RegisterTool(newGetTool(service))
    // ... more tools
}

// In main.go buildMCPServer function
newcontext.RegisterTools(server, newcontextService)
```

### Implementing a New Outbound Adapter

1. Identify the port interface in domain:

```go
// domain/payment/ports.go
type PaymentGateway interface {
    Authorize(ctx context.Context, payment *Payment) (transactionID string, err error)
    Capture(ctx context.Context, transactionID string, amount Money) error
    Refund(ctx context.Context, transactionID string, amount Money) error
}
```

2. Implement in `adapters/outbound/`:

```go
// adapters/outbound/stripe_payment_gateway.go
type StripePaymentGateway struct {
    client *stripe.Client
}

func NewStripePaymentGateway(apiKey string) *StripePaymentGateway {
    return &StripePaymentGateway{
        client: stripe.NewClient(apiKey),
    }
}

func (g *StripePaymentGateway) Authorize(ctx context.Context, pay *payment.Payment) (string, error) {
    // Stripe API call
}
```

3. Inject in `main.go`:

```go
paymentGateway := outbound.NewStripePaymentGateway(os.Getenv("STRIPE_API_KEY"))
paymentService := payment.NewService(paymentRepo, paymentGateway, paymentPublisher)
```

---

## Infrastructure

### Docker Compose Services

```yaml
services:
  app:
    image: "${USER}/${APP_SHORTNAME}:latest"
    depends_on: [keycloak, kafka, postgres-reservation, postgres-payment]
    ports: ["8080:8080"]

  keycloak:
    image: quay.io/keycloak/keycloak:latest
    ports: ["8180:8080"]
    # OIDC provider

  kafka:
    image: apache/kafka:latest
    ports: ["9092:9092"]
    # Event streaming

  postgres-reservation:
    image: postgres:16-alpine
    ports: ["5432:5432"]
    # Reservation database

  postgres-payment:
    image: postgres:16-alpine
    ports: ["5433:5432"]
    # Payment database
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `APP_NAME` | - | Application display name |
| `APP_SHORTNAME` | - | Short name for Docker image |
| `RESERVATION_DB_HOST` | `localhost` | Reservation DB host |
| `RESERVATION_DB_PORT` | `5432` | Reservation DB port |
| `RESERVATION_DB_USER` | `reservation` | Reservation DB user |
| `RESERVATION_DB_PASSWORD` | `reservation_secret` | Reservation DB password |
| `RESERVATION_DB_NAME` | `reservation_db` | Reservation DB name |
| `PAYMENT_DB_HOST` | `localhost` | Payment DB host |
| `PAYMENT_DB_PORT` | `5433` | Payment DB port |
| `PAYMENT_DB_USER` | `payment` | Payment DB user |
| `PAYMENT_DB_PASSWORD` | `payment_secret` | Payment DB password |
| `PAYMENT_DB_NAME` | `payment_db` | Payment DB name |

### Embedded Filesystem

Static assets are embedded at compile time:

```go
//go:embed assets
var efs embed.FS
```

---

## Security

### Authentication

- **Keycloak** provides OIDC/OAuth2 authentication
- Sessions managed via `cloud-native-utils/web` package
- Protected routes use `web.WithAuth` middleware

### Authorization

- Guests can only view/modify their own reservations:

```go
if string(res.GuestID) != email {
    http.Error(w, "Access denied", http.StatusForbidden)
    return
}
```

### Cross-Context Security

- Databases are isolated with separate credentials
- Event messaging uses internal network only
- No direct database access between contexts

---

## Deployment

### Development

```bash
just up       # Start all services
just down     # Stop all services
just logs     # View application logs
```

### Production Dockerfile

```dockerfile
# Multi-stage build
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server ./cmd/server

FROM alpine:latest
COPY --from=builder /app/server /server
ENTRYPOINT ["/server"]
```

### Health Checks

- `/liveness` - Application is running
- `/readiness` - Application can serve requests

---

## Glossary

| Term | Definition |
|------|------------|
| **Aggregate** | Cluster of domain objects treated as a single unit |
| **Bounded Context** | Logical boundary within which a domain model is defined |
| **Domain Event** | Record of something significant that happened in the domain |
| **Port** | Interface defining how the domain interacts with the outside world |
| **Adapter** | Implementation of a port that connects to external systems |
| **Saga** | Pattern for managing distributed transactions via compensation |
| **Shared Kernel** | Common code shared between bounded contexts |
| **Value Object** | Immutable object defined by its attributes, not identity |
