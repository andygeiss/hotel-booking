# Architecture & Development Guide

> **Purpose**: Agent-friendly documentation capturing all architectural decisions, patterns, conventions, and recipes for this Go hexagonal architecture project with Domain-Driven Design.

---

## Quick Reference

### Commands

| Command | Purpose |
|---------|---------|
| `just setup` | Install dev tools (docker-compose, golangci-lint, just, podman) |
| `just up` | Build image + start Keycloak, Kafka, and app |
| `just down` | Stop and remove all containers |
| `just test` | Run unit tests with coverage (outputs `.coverage.pprof`) |
| `just test-integration` | Run integration tests (requires external services) |
| `just lint` | Run golangci-lint (read-only check) |
| `just fmt` | Format code with golangci-lint |
| `just profile` | Generate CPU profile for PGO optimization |

### Run Single Test

```bash
go test -v -run TestFunctionName ./internal/domain/booking/...
```

### Critical Paths

| Layer | Path | Purpose |
|-------|------|---------|
| Entry | `cmd/server/main.go` | DI wiring, bootstrap |
| Domain | `internal/domain/booking/` | Business logic, aggregates, ports |
| Inbound | `internal/adapters/inbound/` | HTTP handlers, event subscribers |
| Outbound | `internal/adapters/outbound/` | Repositories, gateways, publishers |

---

## 1. Architecture Overview

### Hexagonal Architecture (Ports & Adapters)

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
┌─────────────────┐          ┌─────────────────┐          ┌─────────────────┐
│ Inbound Adapter │          │  Domain Layer   │          │Outbound Adapter │
│  (HTTP, Events) │─────────▶│ (Pure Business) │◀─────────│ (Repos, Gateways)│
│                 │          │                 │          │                 │
│ implements      │          │   defines       │          │ implements      │
│ domain ports    │          │   ports         │          │ domain ports    │
└─────────────────┘          └─────────────────┘          └─────────────────┘
```

### Dependency Rule

- **Domain layer has ZERO infrastructure imports**
- Adapters import domain; domain never imports adapters
- All external dependencies injected via interfaces (ports)
- DI wiring happens only in `main.go`

### File Organization

```
internal/domain/{context}/
├── {aggregate}.go              # Aggregate root + value objects + errors
├── {aggregate}_test.go         # Aggregate unit tests
├── {aggregate}_service.go      # Domain service + events
├── {aggregate}_service_test.go # Service tests
├── ports.go                    # Interface definitions (ports)
└── orchestration_service.go    # Saga/workflow coordination

internal/adapters/inbound/
├── router.go                   # HTTP routing + middleware composition
├── http_{feature}.go           # HTTP handler for feature
├── http_{feature}_test.go      # Handler tests
└── event_subscriber.go         # Event subscription handler

internal/adapters/outbound/
├── file_{entity}_repository.go     # File-based repository
├── file_{entity}_repository_test.go
├── repository_{checker}.go         # Composite adapter
├── mock_{service}.go               # Mock adapter (for prod/test)
├── mock_{service}_test.go
└── event_publisher.go              # Event publishing adapter
```

---

## 2. Domain Layer Patterns

### 2.1 Strong-Typed IDs

**Pattern**: Use type aliases for all entity identifiers.

**Location**: `internal/domain/booking/reservation.go:10-13`

```go
type ReservationID string
type GuestID string
type RoomID string
```

**Rationale**: Prevents accidental parameter swapping, improves readability, enables type-safe method signatures.

---

### 2.2 Value Objects

**Pattern**: Immutable structs with factory functions.

**Location**: `internal/domain/booking/reservation.go:16-25`

```go
type DateRange struct {
    CheckIn  time.Time
    CheckOut time.Time
}

type Money struct {
    Currency string // ISO 4217 (e.g., "USD")
    Amount   int64  // Cents/smallest unit
}

// Factory function with normalization
func NewMoney(amount int64, currency string) Money {
    return Money{
        Amount:   amount,
        Currency: strings.ToUpper(currency),
    }
}

func NewDateRange(checkIn, checkOut time.Time) DateRange {
    return DateRange{CheckIn: checkIn, CheckOut: checkOut}
}

func NewGuestInfo(name, email, phone string) GuestInfo {
    return GuestInfo{Name: name, Email: email, Phone: phone}
}
```

---

### 2.3 Aggregate Roots

**Pattern**: Entity with identity, embedded entities, state machine, and business methods.

**Location**: `internal/domain/booking/reservation.go:46-57`

```go
type Reservation struct {
    ID                 ReservationID
    GuestID            GuestID
    RoomID             RoomID
    DateRange          DateRange
    Status             ReservationStatus
    TotalAmount        Money
    CancellationReason string
    CreatedAt          time.Time
    UpdatedAt          time.Time
    Guests             []GuestInfo  // Embedded entity collection
}
```

---

### 2.4 State Machine Pattern

**Pattern**: Status as typed constant, transitions via methods with validation.

**Location**: `internal/domain/booking/reservation.go:27-36, 93-152`

```go
// Status constants
type ReservationStatus string

const (
    StatusPending   ReservationStatus = "pending"
    StatusConfirmed ReservationStatus = "confirmed"
    StatusActive    ReservationStatus = "active"
    StatusCompleted ReservationStatus = "completed"
    StatusCancelled ReservationStatus = "cancelled"
)

// State transition method with validation
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

**Payment State Diagram** (`internal/domain/booking/payment.go:15-21`):
```
pending ──▶ authorized ──▶ captured ──▶ refunded
   │            │
   └────────────┴──▶ failed
```

---

### 2.5 Sentinel Errors

**Pattern**: Package-level error variables for type checking with `errors.Is()`.

**Location**: `internal/domain/booking/reservation.go:59-70`

```go
var (
    ErrInvalidDateRange        = errors.New("check-out must be after check-in")
    ErrCheckInPast             = errors.New("check-in date must be in the future")
    ErrMinimumStay             = errors.New("minimum stay is 1 night")
    ErrInvalidStateTransition  = errors.New("invalid state transition")
    ErrCannotCancelNearCheckIn = errors.New("cannot cancel within 24 hours of check-in")
    ErrCannotCancelActive      = errors.New("cannot cancel active reservation")
    ErrCannotCancelCompleted   = errors.New("cannot cancel completed reservation")
    ErrAlreadyCancelled        = errors.New("reservation already cancelled")
    ErrNoGuests                = errors.New("at least one guest required")
)
```

**Usage**:
```go
// In tests
if !errors.Is(err, booking.ErrInvalidDateRange) {
    t.Errorf("expected ErrInvalidDateRange, got %v", err)
}

// In production code - wrap with context
return fmt.Errorf("%w: cannot confirm from %s", ErrInvalidStateTransition, r.Status)
```

---

### 2.6 Ports (Interfaces)

**Pattern**: Interfaces defined in domain layer, implemented by adapters.

**Location**: `internal/domain/booking/ports.go`

```go
// Repository port (uses generics from cloud-native-utils)
type ReservationRepository resource.Access[ReservationID, Reservation]
type PaymentRepository resource.Access[PaymentID, Payment]

// External service ports
type PaymentGateway interface {
    Authorize(ctx context.Context, payment *Payment) (transactionID string, err error)
    Capture(ctx context.Context, transactionID string, amount Money) error
    Refund(ctx context.Context, transactionID string, amount Money) error
}

type AvailabilityChecker interface {
    IsRoomAvailable(ctx context.Context, roomID RoomID, dateRange DateRange) (bool, error)
    GetOverlappingReservations(ctx context.Context, roomID RoomID, dateRange DateRange) ([]*Reservation, error)
}

type NotificationService interface {
    SendReservationConfirmation(ctx context.Context, reservation *Reservation) error
    SendCancellationNotice(ctx context.Context, reservation *Reservation, reason string) error
    SendPaymentReceipt(ctx context.Context, payment *Payment) error
}

// Event publishing port (alias to external interface)
type EventPublisher event.EventPublisher
```

---

### 2.7 Domain Events

**Pattern**: Events with topic method and fluent builder pattern.

**Location**: `internal/domain/booking/reservation_service.go:11-75`

```go
// Event topic constant (Kafka-style naming)
const EventTopicReservationCreated = "booking.reservation_created"

// Event struct with JSON tags
type EventReservationCreated struct {
    ReservationID ReservationID `json:"reservation_id"`
    GuestID       GuestID       `json:"guest_id"`
    RoomID        RoomID        `json:"room_id"`
    CheckIn       time.Time     `json:"check_in"`
    CheckOut      time.Time     `json:"check_out"`
    TotalAmount   Money         `json:"total_amount"`
}

// Constructor
func NewEventReservationCreated() *EventReservationCreated {
    return &EventReservationCreated{}
}

// Topic method (implements event.Event interface)
func (e *EventReservationCreated) Topic() string {
    return EventTopicReservationCreated
}

// Fluent builder methods
func (e *EventReservationCreated) WithReservationID(id ReservationID) *EventReservationCreated {
    e.ReservationID = id
    return e
}

func (e *EventReservationCreated) WithGuestID(id GuestID) *EventReservationCreated {
    e.GuestID = id
    return e
}
// ... more With* methods
```

**Publishing pattern**:
```go
evt := NewEventReservationCreated().
    WithReservationID(id).
    WithGuestID(guestID).
    WithRoomID(roomID).
    WithCheckIn(dateRange.CheckIn).
    WithCheckOut(dateRange.CheckOut).
    WithTotalAmount(amount)

if err := s.publisher.Publish(ctx, evt); err != nil {
    return nil, fmt.Errorf("failed to publish event: %w", err)
}
```

---

### 2.8 Domain Services

**Pattern**: Service struct with injected ports, orchestrates business operations.

**Location**: `internal/domain/booking/reservation_service.go:139-157`

```go
type ReservationService struct {
    reservationRepo     ReservationRepository
    availabilityChecker AvailabilityChecker
    publisher           event.EventPublisher
}

func NewReservationService(
    repo ReservationRepository,
    checker AvailabilityChecker,
    pub event.EventPublisher,
) *ReservationService {
    return &ReservationService{
        reservationRepo:     repo,
        availabilityChecker: checker,
        publisher:           pub,
    }
}

func (s *ReservationService) CreateReservation(
    ctx context.Context,
    id ReservationID,
    guestID GuestID,
    roomID RoomID,
    dateRange DateRange,
    amount Money,
    guests []GuestInfo,
) (*Reservation, error) {
    // 1. Check availability
    available, err := s.availabilityChecker.IsRoomAvailable(ctx, roomID, dateRange)
    if err != nil {
        return nil, fmt.Errorf("failed to check availability: %w", err)
    }
    if !available {
        return nil, fmt.Errorf("room %s is not available for the selected dates", roomID)
    }

    // 2. Create aggregate
    reservation, err := NewReservation(id, guestID, roomID, dateRange, amount, guests)
    if err != nil {
        return nil, err
    }

    // 3. Persist
    if err := s.reservationRepo.Create(ctx, id, *reservation); err != nil {
        return nil, fmt.Errorf("failed to persist reservation: %w", err)
    }

    // 4. Publish event
    evt := NewEventReservationCreated().
        WithReservationID(id).
        WithGuestID(guestID)
    // ...
    if err := s.publisher.Publish(ctx, evt); err != nil {
        return nil, fmt.Errorf("failed to publish event: %w", err)
    }

    return reservation, nil
}
```

---

### 2.9 Saga/Orchestration Pattern

**Pattern**: Multi-step workflow with explicit compensation on failure.

**Location**: `internal/domain/booking/orchestration_service.go`

```go
type BookingOrchestrationService struct {
    reservationService  *ReservationService
    paymentService      *PaymentService
    notificationService NotificationService
}

func (s *BookingOrchestrationService) CompleteBooking(ctx context.Context, ...) (*Reservation, error) {
    // Step 1: Create reservation
    reservation, err := s.createReservationStep(ctx, ...)
    if err != nil {
        return nil, err  // No compensation needed
    }

    // Step 2: Authorize payment
    payment, err := s.authorizePaymentStep(ctx, reservationID, ...)
    if err != nil {
        // Compensation: Cancel reservation
        _ = s.reservationService.CancelReservation(ctx, reservationID, "payment_auth_failed")
        return nil, err
    }

    // Step 3: Capture payment
    if err := s.capturePaymentStep(ctx, payment.ID, reservationID); err != nil {
        // Compensation: Cancel reservation (refund implicit in gateway)
        return nil, err
    }

    // Step 4: Confirm reservation
    if err := s.confirmReservationStep(ctx, reservationID, payment.ID); err != nil {
        // Compensation: Refund + Cancel
        _ = s.paymentService.RefundPayment(ctx, payment.ID)
        _ = s.reservationService.CancelReservation(ctx, reservationID, "confirmation_failed")
        return nil, err
    }

    // Step 5: Send notification (fire-and-forget)
    _ = s.notificationService.SendReservationConfirmation(ctx, reservation)

    return s.reservationService.GetReservation(ctx, reservationID)
}
```

---

## 3. Adapter Layer Patterns

### 3.1 HTTP Handler Factory Pattern

**Pattern**: Factory function returns `http.HandlerFunc` closure with captured dependencies.

**Location**: `internal/adapters/inbound/http_booking_reservations.go:33-81`

```go
// Response DTO (view model)
type ReservationListItem struct {
    ID          string
    RoomID      string
    CheckIn     string
    CheckOut    string
    Status      string
    StatusClass string  // CSS class for styling
    TotalAmount string
    CanCancel   bool
}

type HttpViewReservationsResponse struct {
    AppName      string
    Title        string
    SessionID    string
    Reservations []ReservationListItem
}

// Handler factory
func HttpViewReservations(e *templating.Engine, reservationService *booking.ReservationService) http.HandlerFunc {
    appName := os.Getenv("APP_NAME")  // Captured at creation time
    title := appName + " - Reservations"

    return func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()

        // 1. Extract auth context (set by middleware)
        sessionID, _ := ctx.Value(security.ContextSessionID).(string)
        email, _ := ctx.Value(security.ContextEmail).(string)

        // 2. Auth check
        if sessionID == "" || email == "" {
            redirecting.Redirect(w, r, "/ui/login")
            return
        }

        // 3. Call domain service
        guestID := booking.GuestID(email)
        reservations, err := reservationService.ListReservationsByGuest(ctx, guestID)
        if err != nil {
            http.Error(w, "Failed to load reservations", http.StatusInternalServerError)
            return
        }

        // 4. Map domain to view model
        items := make([]ReservationListItem, 0, len(reservations))
        for _, res := range reservations {
            items = append(items, ReservationListItem{
                ID:          string(res.ID),
                CheckIn:     res.DateRange.CheckIn.Format("2006-01-02"),
                CheckOut:    res.DateRange.CheckOut.Format("2006-01-02"),
                Status:      string(res.Status),
                StatusClass: getStatusClass(res.Status),
                TotalAmount: res.TotalAmount.FormatAmount(),
                CanCancel:   res.CanBeCancelled(),
            })
        }

        // 5. Render template
        data := HttpViewReservationsResponse{
            AppName:      appName,
            Title:        title,
            SessionID:    sessionID,
            Reservations: items,
        }
        HttpView(e, "reservations", data)(w, r)
    }
}
```

---

### 3.2 Router Middleware Composition

**Pattern**: Functional middleware chain, explicit composition per route.

**Location**: `internal/adapters/inbound/router.go`

```go
func Route(ctx context.Context, efs fs.FS, logger *slog.Logger,
           reservationService *booking.ReservationService) *http.ServeMux {

    mux, serverSessions := security.NewServeMux(ctx, efs)
    e := templating.NewEngine(efs)
    e.Parse("assets/templates/*.tmpl")

    // Public endpoints (no auth middleware)
    mux.HandleFunc("GET /ui/login",
        logging.WithLogging(logger, HttpViewLogin(e)))

    // Protected endpoints (with auth middleware)
    mux.HandleFunc("GET /ui/",
        logging.WithLogging(logger,
            security.WithAuth(serverSessions, HttpViewIndex(e))))

    mux.HandleFunc("GET /ui/reservations",
        logging.WithLogging(logger,
            security.WithAuth(serverSessions,
                HttpViewReservations(e, reservationService))))

    mux.HandleFunc("POST /ui/reservations",
        logging.WithLogging(logger,
            security.WithAuth(serverSessions,
                HttpCreateReservation(e, reservationService))))

    return mux
}
```

**Middleware order**: `logging` (outermost) → `security.WithAuth` → `handler` (innermost)

---

### 3.3 Repository Implementation (Generics)

**Pattern**: Embed generic file access from cloud-native-utils, return interface type.

**Location**: `internal/adapters/outbound/file_reservation_repository.go`

```go
type FileReservationRepository struct {
    resource.JsonFileAccess[booking.ReservationID, booking.Reservation]
}

func NewFileReservationRepository(filename string) booking.ReservationRepository {
    return resource.NewJsonFileAccess[booking.ReservationID, booking.Reservation](filename)
}
```

**Key characteristics**:
- Composition over inheritance (embedded generic type)
- Constructor returns interface type (port), not concrete type
- Minimal code - logic delegated to library

---

### 3.4 Mock Adapter Pattern

**Pattern**: Configurable mock for production/testing with error injection and state tracking.

**Location**: `internal/adapters/outbound/mock_payment_gateway.go`

```go
type MockPaymentGateway struct {
    transactions map[string]booking.Money
    FailureRate  float64  // 0.0 to 1.0 for chaos testing
    ShouldFail   bool     // Force immediate failure
}

func NewMockPaymentGateway() *MockPaymentGateway {
    return &MockPaymentGateway{
        ShouldFail:   false,
        FailureRate:  0.0,
        transactions: make(map[string]booking.Money),
    }
}

func (g *MockPaymentGateway) Authorize(ctx context.Context, payment *booking.Payment) (string, error) {
    // Error injection for chaos engineering
    if g.ShouldFail || (g.FailureRate > 0 && cryptoRandFloat64() < g.FailureRate) {
        return "", errors.New("payment authorization failed: insufficient funds")
    }

    // Track transaction for validation
    transactionID := fmt.Sprintf("txn_%s_%d", payment.ID, payment.Amount.Amount)
    g.transactions[transactionID] = payment.Amount

    return transactionID, nil
}

func (g *MockPaymentGateway) Capture(ctx context.Context, transactionID string, amount booking.Money) error {
    // Validate against authorized amount
    authorizedAmount, exists := g.transactions[transactionID]
    if !exists {
        return fmt.Errorf("transaction %s not found", transactionID)
    }
    if authorizedAmount.Amount != amount.Amount {
        return fmt.Errorf("capture amount mismatch")
    }
    return nil
}

// Test helper methods
func (g *MockPaymentGateway) SetShouldFail(shouldFail bool) { g.ShouldFail = shouldFail }
func (g *MockPaymentGateway) SetFailureRate(rate float64)   { g.FailureRate = rate }
func (g *MockPaymentGateway) Reset() {
    g.transactions = make(map[string]booking.Money)
    g.ShouldFail = false
    g.FailureRate = 0.0
}
```

---

### 3.5 Event Publisher/Subscriber

**Location**: `internal/adapters/outbound/event_publisher.go`, `internal/adapters/inbound/event_subscriber.go`

```go
// Publisher: domain event → JSON → messaging
type EventPublisher struct {
    dispatcher messaging.Dispatcher
}

func NewEventPublisher(dispatcher messaging.Dispatcher) *EventPublisher {
    return &EventPublisher{dispatcher: dispatcher}
}

func (ep *EventPublisher) Publish(ctx context.Context, e event.Event) error {
    encoded, err := json.Marshal(e)
    if err != nil {
        return err
    }
    msg := messaging.NewMessage(e.Topic(), encoded)
    return ep.dispatcher.Publish(ctx, msg)
}

// Subscriber: messaging → JSON → domain handler (with factory pattern)
func (es *EventSubscriber) Subscribe(ctx context.Context, topic string,
    factory func() event.Event,
    handler func(e event.Event) error) error {

    messageFn := func(msg messaging.Message) (messaging.MessageState, error) {
        evt := factory()  // Type-safe instantiation
        if err := json.Unmarshal(msg.Data, evt); err != nil {
            return messaging.MessageStateFailed, err
        }
        if err := handler(evt); err != nil {
            return messaging.MessageStateFailed, err
        }
        return messaging.MessageStateCompleted, nil
    }
    return es.dispatcher.Subscribe(ctx, topic, service.Wrap(messageFn))
}
```

---

## 4. Testing Patterns

### 4.1 Test Naming Convention

**Pattern**: `Test_{Component}_{Scenario}_Should_{ExpectedResult}`

```go
// Domain unit tests
func Test_Reservation_Confirm_From_Pending_Should_Change_Status(t *testing.T)
func Test_NewReservation_With_CheckOut_Before_CheckIn_Should_Return_Error(t *testing.T)
func Test_Reservation_Cancel_Within_24_Hours_Should_Return_Error(t *testing.T)

// Service tests
func Test_ReservationService_CreateReservation_Should_Succeed(t *testing.T)
func Test_ReservationService_CreateReservation_Room_Unavailable_Should_Fail(t *testing.T)

// HTTP handler tests
func Test_Route_Liveness_Endpoint_Should_Return_200(t *testing.T)
func Test_Route_UI_Endpoint_Without_Session_Should_Redirect_To_Login(t *testing.T)

// Adapter tests
func Test_MockPaymentGateway_Authorize_With_ShouldFail_Should_Return_Error(t *testing.T)
```

---

### 4.2 AAA Pattern (Arrange-Act-Assert)

**Pattern**: Clear phase separation with optional comments.

**Location**: `internal/domain/booking/reservation_service_test.go:17-51`

```go
func Test_ReservationService_CreateReservation_Should_Succeed(t *testing.T) {
    // Arrange
    resRepo := newMockReservationRepository()
    availChecker := &mockAvailabilityChecker{available: true}
    publisher := &mockEventPublisher{}
    svc := booking.NewReservationService(resRepo, availChecker, publisher)
    ctx := context.Background()

    checkIn := time.Now().AddDate(0, 0, 7)
    checkOut := time.Now().AddDate(0, 0, 10)
    dateRange := booking.NewDateRange(checkIn, checkOut)
    guests := []booking.GuestInfo{
        booking.NewGuestInfo("John Doe", "john@example.com", "+1234567890"),
    }
    amount := booking.NewMoney(30000, "USD")

    // Act
    reservation, err := svc.CreateReservation(
        ctx, "res-001", "guest-001", "room-101",
        dateRange, amount, guests,
    )

    // Assert
    assert.That(t, "error must be nil", err == nil, true)
    assert.That(t, "reservation must not be nil", reservation != nil, true)
    assert.That(t, "status must be pending", reservation.Status, booking.StatusPending)
}
```

---

### 4.3 Assert Helper

**Pattern**: Use `assert.That()` from cloud-native-utils for consistent assertions.

```go
import "github.com/andygeiss/cloud-native-utils/assert"

// Syntax: assert.That(t, description, actual, expected)
assert.That(t, "error must be nil", err == nil, true)
assert.That(t, "status must be pending", reservation.Status, booking.StatusPending)
assert.That(t, "must return 2 reservations", len(reservations), 2)
assert.That(t, "transaction ID must not be empty", txnID != "", true)
```

---

### 4.4 Mock Implementation Pattern (In-Test)

**Pattern**: Simple structs with error injection fields, defined in test file.

**Location**: `internal/domain/booking/orchestration_service_test.go:18-70`

```go
type mockReservationRepository struct {
    reservations map[booking.ReservationID]booking.Reservation
    createErr    error  // Inject error for Create()
    readErr      error  // Inject error for Read()
    updateErr    error  // Inject error for Update()
}

func newMockReservationRepository() *mockReservationRepository {
    return &mockReservationRepository{
        reservations: make(map[booking.ReservationID]booking.Reservation),
    }
}

func (m *mockReservationRepository) Create(ctx context.Context, id booking.ReservationID, res booking.Reservation) error {
    if m.createErr != nil {
        return m.createErr
    }
    m.reservations[id] = res
    return nil
}

func (m *mockReservationRepository) Read(ctx context.Context, id booking.ReservationID) (booking.Reservation, error) {
    if m.readErr != nil {
        return booking.Reservation{}, m.readErr
    }
    res, ok := m.reservations[id]
    if !ok {
        return booking.Reservation{}, errors.New("not found")
    }
    return res, nil
}

// Event publisher mock with recording
type mockEventPublisher struct {
    publishErr error
    events     []event.Event  // Records for assertion
}

func (m *mockEventPublisher) Publish(ctx context.Context, e event.Event) error {
    if m.publishErr != nil {
        return m.publishErr
    }
    m.events = append(m.events, e)
    return nil
}
```

---

### 4.5 Test Helper Functions

**Pattern**: Use `t.Helper()` for reusable setup functions.

**Location**: `internal/domain/booking/reservation_test.go:447-468`

```go
func createValidReservation(t *testing.T) *booking.Reservation {
    t.Helper()  // Marks as helper for accurate error line reporting

    checkIn := time.Now().AddDate(0, 0, 7)
    checkOut := time.Now().AddDate(0, 0, 10)
    dateRange := booking.NewDateRange(checkIn, checkOut)
    guests := []booking.GuestInfo{
        booking.NewGuestInfo("John Doe", "john@example.com", "+1234567890"),
    }

    reservation, err := booking.NewReservation(
        "res-001", "guest-001", "room-101",
        dateRange, booking.NewMoney(30000, "USD"), guests,
    )
    if err != nil {
        t.Fatalf("failed to create test reservation: %v", err)
    }
    return reservation
}

func createTestDateRange() booking.DateRange {
    return booking.NewDateRange(
        time.Now().AddDate(0, 0, 7),
        time.Now().AddDate(0, 0, 10),
    )
}
```

---

### 4.6 HTTP Handler Testing

**Pattern**: `httptest.NewRequest` + `httptest.NewRecorder` for handler tests.

**Location**: `internal/adapters/inbound/router_test.go:67-85`

```go
func Test_Route_Liveness_Endpoint_Should_Return_200(t *testing.T) {
    // Arrange
    t.Setenv("APP_NAME", "TestApp")
    ctx := context.Background()
    logger := slog.Default()
    reservationService := createTestReservationService(t)
    mux := inbound.Route(ctx, getRouterTestFS(t), logger, reservationService)

    req := httptest.NewRequest(http.MethodGet, "/liveness", nil)
    rec := httptest.NewRecorder()

    // Act
    mux.ServeHTTP(rec, req)

    // Assert
    assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
}
```

---

### 4.7 Embedded Test Assets

**Pattern**: `//go:embed` with `fs.Sub` for template testing.

**Location**: `internal/adapters/inbound/router_test.go:26-37`

```go
//go:embed testdata/assets
var routerTestAssetsRaw embed.FS

func getRouterTestFS(t *testing.T) fs.FS {
    t.Helper()
    sub, err := fs.Sub(routerTestAssetsRaw, "testdata")
    if err != nil {
        t.Fatalf("failed to create sub filesystem: %v", err)
    }
    return sub
}
```

---

## 5. Error Handling Patterns

### 5.1 Error Wrapping

**Pattern**: Wrap errors with context using `fmt.Errorf("%w", err)`.

```go
if err := s.reservationRepo.Create(ctx, id, *reservation); err != nil {
    return nil, fmt.Errorf("failed to persist reservation: %w", err)
}

if err := s.publisher.Publish(ctx, evt); err != nil {
    return nil, fmt.Errorf("failed to publish event: %w", err)
}
```

### 5.2 Sentinel Error Checking

**Pattern**: Use `errors.Is()` for type checking in tests and production code.

```go
// In tests
if !errors.Is(err, booking.ErrInvalidDateRange) {
    t.Errorf("expected ErrInvalidDateRange, got %v", err)
}

// In production
if errors.Is(err, booking.ErrAlreadyCancelled) {
    // Handle specific case
}
```

### 5.3 State Transition Error Context

**Pattern**: Include current state in error message for debugging.

```go
if r.Status != StatusPending {
    return fmt.Errorf("%w: cannot confirm from %s", ErrInvalidStateTransition, r.Status)
}

if p.Status != PaymentPending && p.Status != PaymentFailed {
    return fmt.Errorf("%w: cannot authorize from %s", ErrInvalidPaymentTransition, p.Status)
}
```

---

## 6. Naming Conventions

### 6.1 File Naming

| Pattern | Example | Purpose |
|---------|---------|---------|
| `{aggregate}.go` | `reservation.go` | Aggregate root + value objects + errors |
| `{aggregate}_service.go` | `reservation_service.go` | Domain service + events |
| `{aggregate}_test.go` | `reservation_test.go` | Unit tests (colocated) |
| `http_{feature}.go` | `http_booking_reservations.go` | HTTP handler |
| `http_{feature}_test.go` | `http_booking_reservations_test.go` | Handler tests |
| `file_{entity}_repository.go` | `file_reservation_repository.go` | File-based repository |
| `repository_{checker}.go` | `repository_availability_checker.go` | Composite adapter |
| `mock_{service}.go` | `mock_payment_gateway.go` | Mock adapter |
| `event_publisher.go` | - | Event publishing adapter |
| `event_subscriber.go` | - | Event subscription adapter |
| `router.go` | - | HTTP routing |
| `ports.go` | - | Interface definitions |
| `orchestration_service.go` | - | Saga coordination |

### 6.2 Type Naming

| Category | Pattern | Examples |
|----------|---------|----------|
| IDs | `{Entity}ID` | `ReservationID`, `PaymentID`, `GuestID`, `RoomID` |
| Status | `{Entity}Status` | `ReservationStatus`, `PaymentStatus` |
| Value Objects | CamelCase noun | `DateRange`, `Money`, `GuestInfo` |
| Aggregates | CamelCase noun | `Reservation`, `Payment` |
| Events | `Event{Aggregate}{Action}` | `EventReservationCreated`, `EventPaymentCaptured` |
| Services | `{Domain}Service` | `ReservationService`, `BookingOrchestrationService` |
| HTTP Response | `Http{Action}{Subject}Response` | `HttpViewReservationsResponse` |
| Ports | Verb-noun | `PaymentGateway`, `AvailabilityChecker`, `NotificationService` |

### 6.3 Function Naming

| Category | Pattern | Examples |
|----------|---------|----------|
| Constructors | `New{Type}` | `NewReservation()`, `NewMoney()`, `NewPaymentService()` |
| HTTP handlers | `Http{Action}{Subject}` | `HttpViewReservations()`, `HttpCreateReservation()` |
| State transitions | Verb | `Confirm()`, `Activate()`, `Complete()`, `Cancel()`, `Authorize()`, `Capture()` |
| Query methods | `Is{Condition}` / `Can{Action}` / `{Noun}` | `IsOverlapping()`, `CanBeCancelled()`, `Nights()` |
| Event builders | `With{Field}` | `WithReservationID()`, `WithAmount()`, `WithGuestID()` |
| Test helpers | `create{Subject}` / `new{Mock}` | `createValidReservation()`, `newMockReservationRepository()` |

### 6.4 Constant Naming

```go
// Status constants: Type prefix + CamelCase
const (
    StatusPending   ReservationStatus = "pending"
    StatusConfirmed ReservationStatus = "confirmed"
    PaymentPending  PaymentStatus = "pending"
    PaymentCaptured PaymentStatus = "captured"
)

// Event topics: dot-separated namespace
const (
    EventTopicReservationCreated   = "booking.reservation_created"
    EventTopicReservationConfirmed = "booking.reservation_confirmed"
    EventTopicPaymentAuthorized    = "payment.payment_authorized"
)

// Error variables: Err prefix + CamelCase description
var (
    ErrInvalidDateRange       = errors.New("check-out must be after check-in")
    ErrInvalidStateTransition = errors.New("invalid state transition")
)
```

---

## 7. Recipes (Cookbooks)

### Recipe 1: Add a New Aggregate

**Steps**:

1. **Create aggregate file**: `internal/domain/{context}/{aggregate}.go`

```go
package {context}

import (
    "errors"
    "fmt"
    "time"
)

// 1. Strong-typed ID
type {Aggregate}ID string

// 2. Status type and constants
type {Aggregate}Status string

const (
    {Aggregate}StatusPending   {Aggregate}Status = "pending"
    {Aggregate}StatusCompleted {Aggregate}Status = "completed"
)

// 3. Value objects (if needed)
type {ValueObject} struct {
    Field1 string
    Field2 int64
}

func New{ValueObject}(field1 string, field2 int64) {ValueObject} {
    return {ValueObject}{Field1: field1, Field2: field2}
}

// 4. Sentinel errors
var (
    Err{Aggregate}InvalidInput = errors.New("invalid input for {aggregate}")
    Err{Aggregate}InvalidTransition = errors.New("invalid state transition")
)

// 5. Aggregate struct
type {Aggregate} struct {
    ID        {Aggregate}ID
    Status    {Aggregate}Status
    CreatedAt time.Time
    UpdatedAt time.Time
}

// 6. Constructor with validation
func New{Aggregate}(id {Aggregate}ID, /* params */) (*{Aggregate}, error) {
    // Validate inputs
    if id == "" {
        return nil, Err{Aggregate}InvalidInput
    }

    return &{Aggregate}{
        ID:        id,
        Status:    {Aggregate}StatusPending,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }, nil
}

// 7. State transition methods
func (a *{Aggregate}) Complete() error {
    if a.Status != {Aggregate}StatusPending {
        return fmt.Errorf("%w: cannot complete from %s", Err{Aggregate}InvalidTransition, a.Status)
    }
    a.Status = {Aggregate}StatusCompleted
    a.UpdatedAt = time.Now()
    return nil
}
```

2. **Add port to ports.go**:

```go
type {Aggregate}Repository resource.Access[{Aggregate}ID, {Aggregate}]
```

3. **Create service**: `internal/domain/{context}/{aggregate}_service.go`

4. **Create tests**: `internal/domain/{context}/{aggregate}_test.go`

5. **Create repository adapter**: `internal/adapters/outbound/file_{aggregate}_repository.go`

---

### Recipe 2: Add a New HTTP Endpoint

**Steps**:

1. **Create handler**: `internal/adapters/inbound/http_{feature}.go`

```go
package inbound

import (
    "net/http"
    "os"

    "your/domain/booking"
    "github.com/andygeiss/cloud-native-utils/redirecting"
    "github.com/andygeiss/cloud-native-utils/security"
    "github.com/andygeiss/cloud-native-utils/templating"
)

// Response DTO
type Http{Feature}Response struct {
    AppName string
    Title   string
    // ... view model fields
}

// Handler factory
func Http{Feature}(e *templating.Engine, svc *booking.Service) http.HandlerFunc {
    appName := os.Getenv("APP_NAME")

    return func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()

        // 1. Auth check
        sessionID, _ := ctx.Value(security.ContextSessionID).(string)
        if sessionID == "" {
            redirecting.Redirect(w, r, "/ui/login")
            return
        }

        // 2. Parse input (for POST/PUT)
        // 3. Call domain service
        // 4. Map to view model
        // 5. Render template

        data := Http{Feature}Response{
            AppName: appName,
            Title:   appName + " - {Feature}",
        }
        HttpView(e, "{template_name}", data)(w, r)
    }
}
```

2. **Add route to router.go**:

```go
mux.HandleFunc("GET /ui/{path}",
    logging.WithLogging(logger,
        security.WithAuth(serverSessions, Http{Feature}(e, svc))))
```

3. **Create template**: `assets/templates/{template_name}.tmpl`

4. **Add tests**: `internal/adapters/inbound/http_{feature}_test.go`

---

### Recipe 3: Add a New Domain Event

**Steps**:

1. **Add event topic constant** (in `{aggregate}_service.go`):

```go
const EventTopic{Aggregate}{Action} = "{context}.{aggregate}_{action}"
```

2. **Create event struct**:

```go
type Event{Aggregate}{Action} struct {
    {Aggregate}ID {Aggregate}ID `json:"{aggregate}_id"`
    Timestamp     time.Time     `json:"timestamp"`
    // ... additional fields
}

func NewEvent{Aggregate}{Action}() *Event{Aggregate}{Action} {
    return &Event{Aggregate}{Action}{
        Timestamp: time.Now(),
    }
}

func (e *Event{Aggregate}{Action}) Topic() string {
    return EventTopic{Aggregate}{Action}
}
```

3. **Add fluent builder methods**:

```go
func (e *Event{Aggregate}{Action}) With{Aggregate}ID(id {Aggregate}ID) *Event{Aggregate}{Action} {
    e.{Aggregate}ID = id
    return e
}
```

4. **Publish in service method**:

```go
evt := NewEvent{Aggregate}{Action}().
    With{Aggregate}ID(id)
if err := s.publisher.Publish(ctx, evt); err != nil {
    return fmt.Errorf("failed to publish event: %w", err)
}
```

---

### Recipe 4: Write a Test

**Template**:

```go
func Test_{Component}_{Scenario}_Should_{ExpectedResult}(t *testing.T) {
    // Arrange
    mock := newMock{Dependency}()
    // mock.{field}Err = errors.New("injected error")  // For error scenarios
    svc := New{Service}(mock)
    ctx := context.Background()
    // ... setup test data

    // Act
    result, err := svc.{Method}(ctx, args)

    // Assert
    assert.That(t, "error must be nil", err == nil, true)
    assert.That(t, "result must not be nil", result != nil, true)
    assert.That(t, "{field} must equal expected", result.{Field}, expectedValue)
}
```

**For error scenarios**:

```go
func Test_{Component}_{ErrorScenario}_Should_Return_Error(t *testing.T) {
    // Arrange
    mock := newMock{Dependency}()
    mock.{operation}Err = errors.New("simulated failure")
    svc := New{Service}(mock)
    ctx := context.Background()

    // Act
    _, err := svc.{Method}(ctx, args)

    // Assert
    assert.That(t, "error must not be nil", err != nil, true)
    // Or check for specific error type:
    if !errors.Is(err, Expected{Error}) {
        t.Errorf("expected %v, got %v", Expected{Error}, err)
    }
}
```

---

## 8. Dependencies

### External Libraries

| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/andygeiss/cloud-native-utils` | v0.4.18 | Logging, messaging, security, templating, service patterns |
| `github.com/google/uuid` | v1.6.0 | UUID generation |

### cloud-native-utils Packages Used

| Package | Import | Usage |
|---------|--------|-------|
| `assert` | `github.com/andygeiss/cloud-native-utils/assert` | Test assertions |
| `event` | `github.com/andygeiss/cloud-native-utils/event` | Event interface |
| `logging` | `github.com/andygeiss/cloud-native-utils/logging` | JSON logger, HTTP middleware |
| `messaging` | `github.com/andygeiss/cloud-native-utils/messaging` | Dispatcher, Message types |
| `redirecting` | `github.com/andygeiss/cloud-native-utils/redirecting` | HTTP redirects |
| `resource` | `github.com/andygeiss/cloud-native-utils/resource` | Generic repository access |
| `security` | `github.com/andygeiss/cloud-native-utils/security` | Auth middleware, sessions, OIDC |
| `service` | `github.com/andygeiss/cloud-native-utils/service` | Context handling, lifecycle |
| `templating` | `github.com/andygeiss/cloud-native-utils/templating` | HTML template engine |

---

## 9. Configuration

### Environment Variables

| Variable | Purpose | Default/Example |
|----------|---------|-----------------|
| `PORT` | HTTP server port | `8080` |
| `APP_NAME` | Display name for UI | `Template` |
| `APP_SHORTNAME` | Docker image name | `template` |
| `APP_VERSION` | Application version | `1.0.0` |
| `KAFKA_BROKERS` | Kafka broker addresses | `localhost:9092` |
| `KAFKA_CONSUMER_GROUP_ID` | Consumer group ID | `test-group` |
| `OIDC_CLIENT_ID` | Keycloak client ID | `template` |
| `OIDC_CLIENT_SECRET` | Keycloak client secret | `CHANGE_ME_LOCAL_SECRET` |
| `OIDC_ISSUER` | Keycloak issuer URL | `http://localhost:8180/realms/local` |
| `OIDC_REDIRECT_URL` | OAuth callback URL | `http://localhost:8080/auth/callback` |
| `SERVICE_TIMEOUT` | Service call timeout | `5s` |
| `SERVICE_RETRY_MAX` | Max retry attempts | `3` |
| `SERVICE_RETRY_DELAY` | Delay between retries | `5s` |

### Configuration Files

| File | Purpose |
|------|---------|
| `.env` | Environment configuration (copy from `.env.example`) |
| `.env.example` | Template with all variables documented |
| `.keycloak.json` | Keycloak realm config (copy from `.keycloak.json.example`) |
| `.golangci.yml` | Linter configuration |
| `justfile` | Task runner commands |
| `go.mod` | Go module definition (Go 1.25.5) |

---

## 10. Linting Rules

### Configuration

**File**: `.golangci.yml`

### Intentionally Disabled Linters

| Linter | Reason |
|--------|--------|
| `exhaustruct` | Too verbose for tests; not all struct fields need explicit initialization |
| `ireturn` | Conflicts with DDD port pattern (returning interfaces is intentional) |
| `varnamelen` | Too strict; short names like `id`, `ctx`, `err` are idiomatic |
| `wrapcheck` | Too strict for simple internal calls |
| `err113` | Too strict; sentinel errors are preferred pattern here |
| `mnd` | Magic number detection too noisy for business logic constants |
| `paralleltest` | Not forcing `t.Parallel()` on all tests |
| `tagliatelle` | Project uses snake_case JSON tags intentionally |

### Formatting Rules

```yaml
formatters:
  settings:
    gofmt:
      rewrite-rules:
        - pattern: "interface{}"
          replacement: "any"
        - pattern: "a[b:len(a)]"
          replacement: "a[b:]"
```

---

## 11. DI Wiring (main.go)

**Location**: `cmd/server/main.go`

```go
func main() {
    // 1. Application context with graceful shutdown
    ctx, cancel := service.Context()
    defer cancel()

    // 2. Logger
    logger := logging.NewJsonLogger()

    // 3. Outbound adapters (repositories, gateways, publishers)
    reservationRepo := outbound.NewFileReservationRepository("reservations.json")
    availabilityChecker := outbound.NewRepositoryAvailabilityChecker(reservationRepo)
    eventPublisher := outbound.NewEventPublisher(messaging.NewInternalDispatcher())

    // 4. Domain services (wired with outbound adapters)
    reservationService := booking.NewReservationService(
        reservationRepo,
        availabilityChecker,
        eventPublisher,
    )

    // 5. Inbound adapters (HTTP handlers wired with domain services)
    mux := inbound.Route(ctx, efs, logger, reservationService)

    // 6. HTTP server
    srv := security.NewServer(mux)
    defer func() { _ = srv.Close() }()

    // 7. Graceful shutdown
    service.RegisterOnContextDone(ctx, func() {
        _ = srv.Shutdown(context.Background())
    })

    // 8. Start
    logger.Info("server initialized", "port", os.Getenv("PORT"))
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        logger.Error("server failed", "error", err)
    }
}
```

**Key characteristics**:
- Manual wiring (no DI container)
- Explicit dependency order
- Deferred cleanup in reverse order
- All infrastructure created in main, injected into domain
