# Hotel Booking - Architecture & Development Guide

> **Purpose**: Agent-friendly documentation capturing all architectural decisions, patterns, conventions, and recipes for this hotel reservation and payment management system built with Go, Hexagonal Architecture, and Domain-Driven Design.

---

## Quick Reference

### Commands

| Command | Purpose |
|---------|---------|
| `just setup` | Install dev tools (docker-compose, golangci-lint, just, podman) |
| `just up` | Build image + start PostgreSQL, Keycloak, Kafka, and app |
| `just down` | Stop and remove all containers |
| `just test` | Run unit tests with coverage (outputs `.coverage.pprof`) |
| `just test-integration` | Run integration tests (requires external services) |
| `just lint` | Run golangci-lint (read-only check) |
| `just fmt` | Format code with golangci-lint |
| `just profile` | Generate CPU profile for PGO optimization |

### Run Single Test

```bash
go test -v -run TestFunctionName ./internal/domain/reservation/...
```

### Critical Paths

| Layer | Path | Purpose |
|-------|------|---------|
| Entry | `cmd/server/main.go` | DI wiring, bootstrap |
| Shared Kernel | `internal/domain/shared/` | Cross-context types (Money, ReservationID) |
| Reservation Context | `internal/domain/reservation/` | Reservation aggregate, service, events |
| Payment Context | `internal/domain/payment/` | Payment aggregate, service, events |
| Orchestration | `internal/domain/orchestration/` | Saga coordination, event handlers |
| Inbound | `internal/adapters/inbound/` | HTTP handlers, event subscribers |
| Outbound | `internal/adapters/outbound/` | PostgreSQL repositories, gateways, publishers |

---

## 1. Architecture Overview

### Hexagonal Architecture (Ports & Adapters) with Bounded Contexts

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

### Bounded Contexts

The domain is split into three bounded contexts with clear responsibilities:

| Context | Purpose | Key Aggregates |
|---------|---------|----------------|
| **Reservation** | Room booking lifecycle | `Reservation` |
| **Payment** | Payment processing | `Payment` |
| **Orchestration** | Cross-context coordination | Saga coordination |

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
- `reservation.created` - Payment context subscribes to authorize payment
- `reservation.confirmed` - Notification context subscribes
- `reservation.cancelled` - Notification context subscribes
- `payment.authorized` - Orchestration subscribes to capture payment
- `payment.captured` - Reservation context subscribes to confirm reservation
- `payment.failed` - Orchestration subscribes for compensation

### Dependency Rule

- **Domain layer has ZERO infrastructure imports**
- Adapters import domain; domain never imports adapters
- All external dependencies injected via interfaces (ports)
- DI wiring happens only in `main.go`
- Bounded contexts communicate only via events (no direct calls)

### File Organization

```
internal/domain/
├── shared/
│   └── types.go                    # Cross-context types (Money, ReservationID)
├── reservation/
│   ├── aggregate.go                # Reservation aggregate + value objects + errors
│   ├── entities.go                 # DateRange, GuestInfo value objects
│   ├── ports.go                    # Repository, AvailabilityChecker interfaces
│   ├── events.go                   # EventCreated, EventConfirmed, EventCancelled
│   └── service.go                  # ReservationService
├── payment/
│   ├── aggregate.go                # Payment aggregate + status + errors
│   ├── entities.go                 # PaymentAttempt entity
│   ├── ports.go                    # Repository, PaymentGateway interfaces
│   ├── events.go                   # EventAuthorized, EventCaptured, EventFailed
│   └── service.go                  # PaymentService
└── orchestration/
    ├── booking_service.go          # Saga coordinator
    ├── event_handlers.go           # Cross-context event subscriptions
    └── ports.go                    # NotificationService interface

internal/adapters/inbound/
├── router.go                       # HTTP routing + middleware composition
├── http_{feature}.go               # HTTP handler for feature
├── http_{feature}_test.go          # Handler tests
├── http_error.go                   # Error page handler
└── event_subscriber.go             # Event subscription handler

internal/adapters/outbound/
├── postgres_connection.go          # PostgreSQL connection helper
├── postgres_reservation_repository.go  # PostgreSQL reservation repository
├── postgres_payment_repository.go      # PostgreSQL payment repository
├── repository_{checker}.go         # Composite adapter
├── mock_{service}.go               # Mock adapter (for prod/test)
├── mock_{service}_test.go
└── event_publisher.go              # Event publishing adapter

migrations/
└── init.sql                        # PostgreSQL schema
```

---

## 2. Domain Layer Patterns

### 2.1 Shared Kernel

**Pattern**: Common types shared across bounded contexts.

**Location**: `internal/domain/shared/types.go`

```go
package shared

// ReservationID is shared across reservation and payment contexts
type ReservationID string

// Money is a shared value object for currency amounts
type Money struct {
    Amount   int64  // Cents/smallest unit
    Currency string // ISO 4217 (e.g., "USD")
}

func NewMoney(amount int64, currency string) Money {
    return Money{
        Amount:   amount,
        Currency: strings.ToUpper(currency),
    }
}

func (m Money) FormatAmount() string {
    dollars := float64(m.Amount) / 100
    return fmt.Sprintf("%s %.2f", m.Currency, dollars)
}
```

**Rationale**: Provides a minimal shared vocabulary between bounded contexts without tight coupling.

---

### 2.2 Strong-Typed IDs

**Pattern**: Use type aliases for all entity identifiers within each context.

**Location**: `internal/domain/reservation/aggregate.go:10-13`

```go
// Reservation context IDs
type ReservationID = shared.ReservationID  // Alias to shared type
type GuestID string
type RoomID string
```

**Location**: `internal/domain/payment/aggregate.go:10-12`

```go
// Payment context IDs
type PaymentID string
type ReservationID = shared.ReservationID  // Alias to shared type
```

**Rationale**: Prevents accidental parameter swapping, improves readability, enables type-safe method signatures. ReservationID is shared because it's referenced across contexts.

---

### 2.3 Value Objects

**Pattern**: Immutable structs with factory functions, defined per context.

**Location**: `internal/domain/reservation/entities.go`

```go
type DateRange struct {
    CheckIn  time.Time
    CheckOut time.Time
}

func NewDateRange(checkIn, checkOut time.Time) DateRange {
    return DateRange{CheckIn: checkIn, CheckOut: checkOut}
}

type GuestInfo struct {
    Name        string
    Email       string
    PhoneNumber string
}

func NewGuestInfo(name, email, phone string) GuestInfo {
    return GuestInfo{Name: name, Email: email, PhoneNumber: phone}
}
```

---

### 2.4 Aggregate Roots

**Pattern**: Entity with identity, embedded entities, state machine, and business methods.

**Location**: `internal/domain/reservation/aggregate.go`

```go
type Reservation struct {
    ID                 ReservationID
    GuestID            GuestID
    RoomID             RoomID
    DateRange          DateRange
    Status             ReservationStatus
    TotalAmount        shared.Money
    CancellationReason string
    CreatedAt          time.Time
    UpdatedAt          time.Time
    Guests             []GuestInfo  // Embedded entity collection
}
```

**Location**: `internal/domain/payment/aggregate.go`

```go
type Payment struct {
    ID            PaymentID
    ReservationID ReservationID
    Amount        shared.Money
    PaymentMethod string
    Status        PaymentStatus
    TransactionID string
    CreatedAt     time.Time
    UpdatedAt     time.Time
    Attempts      []PaymentAttempt
}
```

---

### 2.5 State Machine Pattern

**Pattern**: Status as typed constant, transitions via methods with validation.

**Location**: `internal/domain/reservation/aggregate.go`

```go
type ReservationStatus string

const (
    StatusPending   ReservationStatus = "pending"
    StatusConfirmed ReservationStatus = "confirmed"
    StatusActive    ReservationStatus = "active"
    StatusCompleted ReservationStatus = "completed"
    StatusCancelled ReservationStatus = "cancelled"
)

func (r *Reservation) Confirm() error {
    if r.Status != StatusPending {
        return fmt.Errorf("%w: cannot confirm from %s", ErrInvalidStateTransition, r.Status)
    }
    r.Status = StatusConfirmed
    r.UpdatedAt = time.Now()
    return nil
}
```

**Reservation State Diagram**:
```
pending ──▶ confirmed ──▶ active ──▶ completed
   │            │
   └────────────┴──▶ cancelled
```

**Payment State Diagram** (`internal/domain/payment/aggregate.go`):
```
pending ──▶ authorized ──▶ captured ──▶ refunded
   │            │
   └────────────┴──▶ failed
```

---

### 2.6 Sentinel Errors

**Pattern**: Package-level error variables for type checking with `errors.Is()`.

**Location**: `internal/domain/reservation/aggregate.go`

```go
var (
    ErrInvalidDateRange        = errors.New("check-out must be after check-in")
    ErrCheckInPast             = errors.New("check-in date must be in the future")
    ErrMinimumStay             = errors.New("minimum stay is 1 night")
    ErrInvalidStateTransition  = errors.New("invalid state transition")
    ErrCannotCancelNearCheckIn = errors.New("cannot cancel within 24 hours of check-in")
    ErrNoGuests                = errors.New("at least one guest required")
)
```

**Location**: `internal/domain/payment/aggregate.go`

```go
var (
    ErrInvalidPaymentTransition = errors.New("invalid payment state transition")
    ErrPaymentNotFound          = errors.New("payment not found")
    ErrPaymentAlreadyAuthorized = errors.New("payment already authorized")
)
```

---

### 2.7 Ports (Interfaces)

**Pattern**: Interfaces defined in each bounded context, implemented by adapters.

**Location**: `internal/domain/reservation/ports.go`

```go
// Repository port (uses generics from cloud-native-utils)
type ReservationRepository resource.Access[ReservationID, Reservation]

// External service ports
type AvailabilityChecker interface {
    IsRoomAvailable(ctx context.Context, roomID RoomID, dateRange DateRange) (bool, error)
    GetOverlappingReservations(ctx context.Context, roomID RoomID, dateRange DateRange) ([]*Reservation, error)
}

// Event publishing port
type EventPublisher event.EventPublisher
```

**Location**: `internal/domain/payment/ports.go`

```go
type PaymentRepository resource.Access[PaymentID, Payment]

type PaymentGateway interface {
    Authorize(ctx context.Context, payment *Payment) (transactionID string, err error)
    Capture(ctx context.Context, transactionID string, amount shared.Money) error
    Refund(ctx context.Context, transactionID string, amount shared.Money) error
}

type EventPublisher event.EventPublisher
```

**Location**: `internal/domain/orchestration/ports.go`

```go
type NotificationService interface {
    SendReservationConfirmation(ctx context.Context, reservation *reservation.Reservation) error
    SendCancellationNotice(ctx context.Context, reservation *reservation.Reservation, reason string) error
    SendPaymentReceipt(ctx context.Context, payment *payment.Payment) error
}
```

---

### 2.8 Domain Events

**Pattern**: Events with topic method, defined per bounded context.

**Location**: `internal/domain/reservation/events.go`

```go
// Event topic constants (Kafka-style naming)
const (
    EventTopicCreated   = "reservation.created"
    EventTopicConfirmed = "reservation.confirmed"
    EventTopicCancelled = "reservation.cancelled"
)

type EventCreated struct {
    ReservationID string      `json:"reservation_id"`
    GuestID       string      `json:"guest_id"`
    RoomID        string      `json:"room_id"`
    CheckIn       time.Time   `json:"check_in"`
    CheckOut      time.Time   `json:"check_out"`
    TotalAmount   shared.Money `json:"total_amount"`
}

func (e *EventCreated) Topic() string {
    return EventTopicCreated
}
```

**Location**: `internal/domain/payment/events.go`

```go
const (
    EventTopicAuthorized = "payment.authorized"
    EventTopicCaptured   = "payment.captured"
    EventTopicFailed     = "payment.failed"
    EventTopicRefunded   = "payment.refunded"
)

type EventAuthorized struct {
    PaymentID     string `json:"payment_id"`
    ReservationID string `json:"reservation_id"`
    TransactionID string `json:"transaction_id"`
}

func (e *EventAuthorized) Topic() string {
    return EventTopicAuthorized
}
```

---

### 2.9 Domain Services

**Pattern**: Service struct with injected ports, orchestrates business operations within a context.

**Location**: `internal/domain/reservation/service.go`

```go
type Service struct {
    reservationRepo     ReservationRepository
    availabilityChecker AvailabilityChecker
    publisher           event.EventPublisher
}

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

func (s *Service) CreateReservation(ctx context.Context, ...) (*Reservation, error) {
    // 1. Check availability
    // 2. Create aggregate
    // 3. Persist
    // 4. Publish event
}
```

**Location**: `internal/domain/payment/service.go`

```go
type Service struct {
    paymentRepo    PaymentRepository
    paymentGateway PaymentGateway
    publisher      event.EventPublisher
}

func (s *Service) AuthorizePaymentForReservation(ctx context.Context, ...) (*Payment, error) {
    // 1. Create payment aggregate
    // 2. Call payment gateway
    // 3. Update payment status
    // 4. Persist
    // 5. Publish event
}
```

---

### 2.10 Saga/Orchestration Pattern (Event-Driven)

**Pattern**: Event-driven saga coordination with compensation on failure.

**Location**: `internal/domain/orchestration/booking_service.go`

```go
type BookingService struct {
    reservationService  *reservation.Service
    paymentService      *payment.Service
    notificationService NotificationService
}

func (s *BookingService) OnPaymentAuthorized(ctx context.Context, paymentID payment.PaymentID, reservationID shared.ReservationID) error {
    // Capture the payment
    if err := s.paymentService.CapturePayment(ctx, paymentID); err != nil {
        return fmt.Errorf("failed to capture payment: %w", err)
    }
    return nil
}

func (s *BookingService) OnPaymentCaptured(ctx context.Context, reservationID shared.ReservationID) error {
    // Confirm the reservation
    if err := s.reservationService.ConfirmReservation(ctx, reservationID); err != nil {
        return fmt.Errorf("failed to confirm reservation: %w", err)
    }
    return nil
}

func (s *BookingService) OnPaymentFailed(ctx context.Context, reservationID shared.ReservationID, reason string) error {
    // Compensation: Cancel the reservation
    if err := s.reservationService.CancelReservation(ctx, reservationID, reason); err != nil {
        return fmt.Errorf("failed to cancel reservation: %w", err)
    }
    return nil
}
```

**Location**: `internal/domain/orchestration/event_handlers.go`

```go
type EventHandlers struct {
    bookingService     *BookingService
    reservationService *reservation.Service
    paymentService     *payment.Service
}

func (h *EventHandlers) RegisterHandlers(ctx context.Context, dispatcher messaging.Dispatcher) error {
    // Payment context subscribes to reservation.created
    if err := dispatcher.Subscribe(ctx, reservation.EventTopicCreated, service.Wrap(h.handleReservationCreated)); err != nil {
        return err
    }

    // Orchestration subscribes to payment.authorized
    if err := dispatcher.Subscribe(ctx, payment.EventTopicAuthorized, service.Wrap(h.handlePaymentAuthorized)); err != nil {
        return err
    }

    // Reservation context subscribes to payment.captured
    if err := dispatcher.Subscribe(ctx, payment.EventTopicCaptured, service.Wrap(h.handlePaymentCaptured)); err != nil {
        return err
    }

    // Orchestration subscribes to payment.failed for compensation
    if err := dispatcher.Subscribe(ctx, payment.EventTopicFailed, service.Wrap(h.handlePaymentFailed)); err != nil {
        return err
    }

    return nil
}

func (h *EventHandlers) handleReservationCreated(msg messaging.Message) (messaging.MessageState, error) {
    var evt reservation.EventCreated
    if err := json.Unmarshal(msg.Data, &evt); err != nil {
        return messaging.MessageStateFailed, err
    }

    // Trigger payment authorization
    _, err := h.paymentService.AuthorizePaymentForReservation(ctx, ...)
    if err != nil {
        return messaging.MessageStateFailed, err
    }

    return messaging.MessageStateCompleted, nil
}
```

---

## 3. Adapter Layer Patterns

### 3.1 PostgreSQL Repository Implementation

**Pattern**: SQL repository implementing domain port, using pgx driver.

**Location**: `internal/adapters/outbound/postgres_reservation_repository.go`

```go
type PostgresReservationRepository struct {
    db *sql.DB
}

func NewPostgresReservationRepository(db *sql.DB) *PostgresReservationRepository {
    return &PostgresReservationRepository{db: db}
}

func (r *PostgresReservationRepository) Create(ctx context.Context, id reservation.ReservationID, res reservation.Reservation) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()

    // Insert reservation
    _, err = tx.ExecContext(ctx, `
        INSERT INTO reservations (id, guest_id, room_id, check_in, check_out, status, total_amount, currency, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    `, id, res.GuestID, res.RoomID, res.DateRange.CheckIn, res.DateRange.CheckOut, res.Status, res.TotalAmount.Amount, res.TotalAmount.Currency, res.CreatedAt, res.UpdatedAt)
    if err != nil {
        return fmt.Errorf("failed to insert reservation: %w", err)
    }

    // Insert guests
    for _, guest := range res.Guests {
        _, err = tx.ExecContext(ctx, `
            INSERT INTO reservation_guests (reservation_id, name, email, phone_number)
            VALUES ($1, $2, $3, $4)
        `, id, guest.Name, guest.Email, guest.PhoneNumber)
        if err != nil {
            return fmt.Errorf("failed to insert guest: %w", err)
        }
    }

    return tx.Commit()
}

func (r *PostgresReservationRepository) Read(ctx context.Context, id reservation.ReservationID) (*reservation.Reservation, error) {
    // Query reservation
    // Query guests
    // Assemble aggregate
    return &res, nil
}
```

**Location**: `internal/adapters/outbound/postgres_connection.go`

```go
func NewPostgresConnection() (*sql.DB, error) {
    dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        os.Getenv("POSTGRES_HOST"),
        os.Getenv("POSTGRES_PORT"),
        os.Getenv("POSTGRES_USER"),
        os.Getenv("POSTGRES_PASSWORD"),
        os.Getenv("POSTGRES_DB"),
        os.Getenv("POSTGRES_SSLMODE"),
    )

    db, err := sql.Open("pgx", dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }

    // Configure connection pool
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(5 * time.Minute)

    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    return db, nil
}
```

---

### 3.2 Database Schema

**Location**: `migrations/init.sql`

```sql
CREATE TABLE reservations (
    id VARCHAR(255) PRIMARY KEY,
    guest_id VARCHAR(255) NOT NULL,
    room_id VARCHAR(255) NOT NULL,
    check_in TIMESTAMP NOT NULL,
    check_out TIMESTAMP NOT NULL,
    status VARCHAR(50) NOT NULL,
    total_amount BIGINT NOT NULL,
    currency VARCHAR(10) NOT NULL,
    cancellation_reason TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE reservation_guests (
    id SERIAL PRIMARY KEY,
    reservation_id VARCHAR(255) NOT NULL REFERENCES reservations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone_number VARCHAR(50)
);

CREATE TABLE payments (
    id VARCHAR(255) PRIMARY KEY,
    reservation_id VARCHAR(255) NOT NULL,
    amount BIGINT NOT NULL,
    currency VARCHAR(10) NOT NULL,
    payment_method VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    transaction_id VARCHAR(255),
    error_code VARCHAR(50),
    error_message TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE payment_attempts (
    id SERIAL PRIMARY KEY,
    payment_id VARCHAR(255) NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL,
    error_code VARCHAR(50),
    error_message TEXT,
    attempted_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_reservations_guest_id ON reservations(guest_id);
CREATE INDEX idx_reservations_room_id ON reservations(room_id);
CREATE INDEX idx_reservations_status ON reservations(status);
CREATE INDEX idx_reservations_check_in ON reservations(check_in);
CREATE INDEX idx_payments_reservation_id ON payments(reservation_id);
```

---

### 3.3 HTTP Handler Factory Pattern

**Pattern**: Factory function returns `http.HandlerFunc` closure with captured dependencies.

**Location**: `internal/adapters/inbound/http_booking_reservations.go`

```go
func HttpViewReservations(e *templating.Engine, reservationService *reservation.Service) http.HandlerFunc {
    appName := os.Getenv("APP_NAME")
    title := appName + " - Reservations"

    return func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()

        sessionID, _ := ctx.Value(security.ContextSessionID).(string)
        email, _ := ctx.Value(security.ContextEmail).(string)
        if sessionID == "" || email == "" {
            redirecting.Redirect(w, r, "/ui/login")
            return
        }

        guestID := reservation.GuestID(email)
        reservations, err := reservationService.ListReservationsByGuest(ctx, guestID)
        // ... map to view model and render
    }
}
```

---

### 3.4 Error Page Handler Pattern

**Pattern**: Dedicated error page handler that accepts error details via query parameters.

**Location**: `internal/adapters/inbound/http_error.go`

```go
type HttpViewErrorResponse struct {
    AppName      string
    Title        string
    ErrorTitle   string
    ErrorMessage string
    ErrorDetails string
}

func HttpViewError(e *templating.Engine) http.HandlerFunc {
    appName := os.Getenv("APP_NAME")
    pageTitle := appName + " - Error"

    return func(w http.ResponseWriter, r *http.Request) {
        errorTitle := r.URL.Query().Get("title")
        errorMessage := r.URL.Query().Get("message")
        errorDetails := r.URL.Query().Get("details")

        // Set defaults if not provided
        if errorTitle == "" {
            errorTitle = "An Error Occurred"
        }
        if errorMessage == "" {
            errorMessage = "Something went wrong. Please try again."
        }

        data := HttpViewErrorResponse{
            AppName:      appName,
            Title:        pageTitle,
            ErrorTitle:   errorTitle,
            ErrorMessage: errorMessage,
            ErrorDetails: errorDetails,
        }

        HttpView(e, "error", data)(w, r)
    }
}
```

**Usage**: Redirect to error page with URL-encoded parameters:
```
/ui/error?title=Authentication%20Failed&message=Invalid%20credentials&details=oauth2%3A%20unauthorized_client
```

**Template Location**: `cmd/server/assets/templates/error.tmpl`

---

### 3.5 Router Middleware Composition

**Pattern**: Functional middleware chain, explicit composition per route.

**Location**: `internal/adapters/inbound/router.go`

```go
func Route(ctx context.Context, efs fs.FS, logger *slog.Logger,
           reservationService *reservation.Service) *http.ServeMux {

    mux, serverSessions := security.NewServeMux(ctx, efs)
    e := templating.NewEngine(efs)
    e.Parse("assets/templates/*.tmpl")

    // Public endpoints
    mux.HandleFunc("GET /ui/login",
        logging.WithLogging(logger, HttpViewLogin(e)))

    // Protected endpoints
    mux.HandleFunc("GET /ui/reservations",
        logging.WithLogging(logger,
            security.WithAuth(serverSessions,
                HttpViewReservations(e, reservationService))))

    return mux
}
```

---

### 3.6 Mock Adapter Pattern

**Pattern**: Configurable mock for production/testing with error injection and state tracking.

**Location**: `internal/adapters/outbound/mock_payment_gateway.go`

```go
type MockPaymentGateway struct {
    transactions map[string]shared.Money
    FailureRate  float64
    ShouldFail   bool
}

func NewMockPaymentGateway() *MockPaymentGateway {
    return &MockPaymentGateway{
        transactions: make(map[string]shared.Money),
    }
}

func (g *MockPaymentGateway) Authorize(ctx context.Context, pay *payment.Payment) (string, error) {
    if g.ShouldFail {
        return "", errors.New("payment authorization failed")
    }

    transactionID := fmt.Sprintf("txn_%s_%d", pay.ID, pay.Amount.Amount)
    g.transactions[transactionID] = pay.Amount
    return transactionID, nil
}
```

---

## 4. Testing Patterns

### 4.1 Test Naming Convention

**Pattern**: `Test_{Component}_{Scenario}_Should_{ExpectedResult}`

```go
// Domain unit tests
func Test_Reservation_Confirm_From_Pending_Should_Change_Status(t *testing.T)

// Service tests
func Test_ReservationService_CreateReservation_Should_Succeed(t *testing.T)

// HTTP handler tests
func Test_Route_Liveness_Endpoint_Should_Return_200(t *testing.T)

// Adapter tests
func Test_MockPaymentGateway_Authorize_With_ShouldFail_Should_Return_Error(t *testing.T)
```

---

### 4.2 Mock Repository Pattern (In-Test)

**Pattern**: Simple in-memory mock implementing repository interface.

```go
type mockReservationRepository struct {
    reservations map[reservation.ReservationID]reservation.Reservation
    readAllErr   error
}

func newMockReservationRepository() *mockReservationRepository {
    return &mockReservationRepository{
        reservations: make(map[reservation.ReservationID]reservation.Reservation),
    }
}

func (m *mockReservationRepository) Create(ctx context.Context, id reservation.ReservationID, res reservation.Reservation) error {
    m.reservations[id] = res
    return nil
}

func (m *mockReservationRepository) Read(ctx context.Context, id reservation.ReservationID) (*reservation.Reservation, error) {
    res, ok := m.reservations[id]
    if !ok {
        return nil, errors.New("not found")
    }
    return &res, nil
}
```

---

## 5. Configuration

### Environment Variables

| Variable | Purpose | Default/Example |
|----------|---------|-----------------|
| `PORT` | HTTP server port | `8080` |
| `APP_NAME` | Display name for UI | `Hotel Booking` |
| `APP_DESCRIPTION` | Application description | `Hotel reservation and payment management system` |
| `APP_SHORTNAME` | Short name for Docker/containers | `hotel-booking` |
| `POSTGRES_HOST` | PostgreSQL host | `localhost` |
| `POSTGRES_PORT` | PostgreSQL port | `5432` |
| `POSTGRES_USER` | PostgreSQL user | `booking` |
| `POSTGRES_PASSWORD` | PostgreSQL password | `booking_secret` |
| `POSTGRES_DB` | PostgreSQL database | `booking_db` |
| `POSTGRES_SSLMODE` | SSL mode | `disable` |
| `KAFKA_BROKERS` | Kafka broker addresses | `localhost:9092` |
| `OIDC_CLIENT_ID` | Keycloak client ID | `hotel-booking` |
| `OIDC_ISSUER` | Keycloak issuer URL | `http://localhost:8180/realms/local` |

---

## 6. DI Wiring (main.go)

**Location**: `cmd/server/main.go`

```go
func main() {
    ctx, cancel := service.Context()
    defer cancel()

    logger := logging.NewJsonLogger()

    // PostgreSQL connection
    db, err := outbound.NewPostgresConnection()
    if err != nil {
        logger.Error("failed to connect to database", "error", err)
        return
    }
    defer db.Close()

    // Event dispatcher (Kafka)
    dispatcher := messaging.NewKafkaDispatcher(os.Getenv("KAFKA_BROKERS"))

    // Reservation context
    reservationRepo := outbound.NewPostgresReservationRepository(db)
    availabilityChecker := outbound.NewRepositoryAvailabilityChecker(reservationRepo)
    reservationPublisher := outbound.NewEventPublisher(dispatcher)
    reservationService := reservation.NewService(reservationRepo, availabilityChecker, reservationPublisher)

    // Payment context
    paymentRepo := outbound.NewPostgresPaymentRepository(db)
    paymentGateway := outbound.NewMockPaymentGateway()
    paymentPublisher := outbound.NewEventPublisher(dispatcher)
    paymentService := payment.NewService(paymentRepo, paymentGateway, paymentPublisher)

    // Orchestration layer
    notificationService := outbound.NewMockNotificationService(logger)
    bookingService := orchestration.NewBookingService(reservationService, paymentService, notificationService)

    // Register event handlers
    eventHandlers := orchestration.NewEventHandlers(bookingService, reservationService, paymentService)
    if err := eventHandlers.RegisterHandlers(ctx, dispatcher); err != nil {
        logger.Error("failed to register event handlers", "error", err)
        return
    }

    // HTTP routing
    mux := inbound.Route(ctx, efs, logger, reservationService)

    // Start server
    srv := security.NewServer(mux)
    defer srv.Close()

    logger.Info("server initialized", "port", os.Getenv("PORT"))
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        logger.Error("server failed", "error", err)
    }
}
```

---

## 7. Dependencies

### External Libraries

| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/andygeiss/cloud-native-utils` | v0.4.18 | Logging, messaging, security, templating |
| `github.com/jackc/pgx/v5` | v5.x | PostgreSQL driver |
| `github.com/google/uuid` | v1.6.0 | UUID generation |

### Infrastructure

| Service | Purpose | Port |
|---------|---------|------|
| PostgreSQL | Primary database | 5432 |
| Kafka | Event messaging | 9092 |
| Keycloak | OIDC authentication | 8180 |
