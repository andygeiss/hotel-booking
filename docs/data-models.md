# Data Models

Full inventory of domain aggregates, value objects, events, and persistence schemas.

---

## Persistence Schema

Both bounded contexts use a single, uniform key/value schema via `cloud-native-utils/resource.PostgresAccess[K, V]`. The same DDL runs on first container start for each database (`migrations/{reservation,payment}/init.sql`):

```sql
CREATE TABLE IF NOT EXISTS kv_store (
    key   TEXT PRIMARY KEY,
    value TEXT
);

CREATE INDEX IF NOT EXISTS idx_kv_store_key ON kv_store (key);
```

The adapter is typed in Go as:

- Reservations: `resource.NewPostgresAccess[reservation.ReservationID, reservation.Reservation](reservationDB)` → `reservation_db`
- Payments: `resource.NewPostgresAccess[payment.PaymentID, payment.Payment](paymentDB)` → `payment_db`

Aggregates are serialized to JSON into `kv_store.value`. **Never cross-query between the two databases** — each bounded context owns its own DB.

---

## Shared Kernel (`internal/domain/shared/types.go`)

### `ReservationID`

```go
type ReservationID string
```

Shared because `payment.Payment` needs to reference it. Format: `res-{uuid}`, enforced by the `shared.NewReservationID()` constructor.

### `Money` (value object)

```go
type Money struct {
    Currency string // ISO 4217 (e.g. "USD")
    Amount   int64  // smallest currency unit (cents)
}
```

- `NewMoney(amount int64, currency string) Money` — uppercases currency.
- `FormatAmount() string` — returns `"99.00 USD"` style strings.
- Immutable by convention — create a new instance instead of mutating.

---

## Reservation Bounded Context (`internal/domain/reservation`)

### Aggregate: `Reservation` (aggregate root)

```go
type Reservation struct {
    ID                 ReservationID   // shared kernel
    GuestID            GuestID         // == OIDC email in current adapters
    RoomID             RoomID
    DateRange          DateRange       // value object (CheckIn/CheckOut)
    Status             ReservationStatus
    TotalAmount        Money           // shared kernel
    CancellationReason string
    CreatedAt          time.Time
    UpdatedAt          time.Time
    Guests             []GuestInfo     // entity collection inside the aggregate
}
```

### Local Types

- `GuestID string` — guest identity within this context (the HTTP adapter maps the OIDC email to this).
- `RoomID string` — hardcoded catalog in `http_booking_reservation_form.go` (`room-101`, `room-102`, `room-201`, `room-202`, `room-301`).
- `DateRange { CheckIn, CheckOut time.Time }` — value object.
- `GuestInfo { Name, Email, PhoneNumber string }` — entity inside the Reservation aggregate.

### Status State Machine

See [ARCHITECTURE.md § Reservation Aggregate](./ARCHITECTURE.md#reservation-aggregate) for the state diagram and transition guards. Status enum: `StatusPending`, `StatusConfirmed`, `StatusActive`, `StatusCompleted`, `StatusCancelled`.

### Validation (in `NewReservation` → `validate()`)

- `ErrNoGuests` — `len(Guests) == 0`.
- `ErrMinimumStay` — `CheckOut == CheckIn` (same-day).
- `ErrInvalidDateRange` — `CheckOut ≤ CheckIn` (other cases).
- `ErrCheckInPast` — `CheckIn < today (truncated to 24h)`.

### Cancel-specific Errors

- `ErrAlreadyCancelled`
- `ErrCannotCancelCompleted`
- `ErrCannotCancelActive`
- `ErrCannotCancelNearCheckIn` (< 24h before check-in)

### Domain Events (`events.go`)

| Type | Topic Constant | Fields |
|------|----------------|--------|
| `EventCreated` | `EventTopicCreated` (`reservation.created`) | `reservation_id, guest_id, room_id, check_in, check_out, total_amount` |
| `EventConfirmed` | `EventTopicConfirmed` | `reservation_id, guest_id` |
| `EventActivated` | `EventTopicActivated` | `reservation_id` |
| `EventCompleted` | `EventTopicCompleted` | `reservation_id` |
| `EventCancelled` | `EventTopicCancelled` | `reservation_id, guest_id, reason` |

Each event uses the builder/options pattern (`NewEvent*().WithX(...)`) and exposes `Topic() string`.

---

## Payment Bounded Context (`internal/domain/payment`)

### Aggregate: `Payment` (aggregate root)

```go
type Payment struct {
    ID            PaymentID
    ReservationID ReservationID       // shared kernel
    Amount        Money               // shared kernel
    Status        PaymentStatus
    PaymentMethod string
    TransactionID string               // external gateway's id
    CreatedAt     time.Time
    UpdatedAt     time.Time
    Attempts      []PaymentAttempt     // entity collection (attempt history)
}
```

### Local Types

- `PaymentID string` — payment identity. Format: `pay-{uuid}`, derived by stripping the `res-` prefix from `ReservationID` in `orchestration.handleReservationCreated`. Both IDs share the same underlying UUID.
- `PaymentAttempt { AttemptedAt, Status, ErrorCode, ErrorMsg }` — appended to `Attempts` on every state change and failure.

### Status State Machine

See [ARCHITECTURE.md § Payment Aggregate](./ARCHITECTURE.md#payment-aggregate) for the state diagram and transition guards. Status enum: `StatusPending`, `StatusAuthorized`, `StatusCaptured`, `StatusFailed`, `StatusRefunded`.

### Payment-specific Errors

- `ErrInvalidPaymentTransition`
- `ErrAlreadyAuthorized`, `ErrAlreadyCaptured`, `ErrAlreadyRefunded`
- `ErrNotAuthorized`, `ErrNotCaptured`, `ErrCannotRefund`

### Retry Policy

`Payment.CanBeRetried()` returns `true` if status ∈ {Pending, Failed} **and** there are fewer than 3 entries in `Attempts` with `Status == Failed`.

### Domain Events (`events.go`)

| Type | Topic Constant | Fields |
|------|----------------|--------|
| `EventAuthorized` | `EventTopicAuthorized` (`payment.authorized`) | `payment_id, reservation_id, transaction_id, amount` |
| `EventCaptured` | `EventTopicCaptured` | `payment_id, reservation_id, amount` |
| `EventFailed` | `EventTopicFailed` | `payment_id, reservation_id, error_code, error_msg` |
| `EventRefunded` | `EventTopicRefunded` | `payment_id, reservation_id, amount` |

---

## Orchestration (`internal/domain/orchestration`)

No aggregate of its own. Coordinates the Saga via `BookingService` + `EventHandlers`. Uses `NotificationService` (outbound port; mock-backed) for confirmation/cancellation notifications.

### `NotificationService` port

```go
type NotificationService interface {
    SendReservationConfirmation(ctx, r *reservation.Reservation) error
    SendCancellationNotice(ctx, r *reservation.Reservation, reason string) error
    SendPaymentReceipt(ctx, p *payment.Payment) error
}
```

Currently backed by `outbound.MockNotificationService`, which writes structured `slog` log lines rather than sending email.

---

## Identifier Conventions

| Type | Format | Notes |
|------|--------|-------|
| `ReservationID` | `res-{uuid}` | Enforced by `shared.NewReservationID()` constructor |
| `PaymentID` | `pay-{uuid}` | Derived by stripping `res-` prefix in `orchestration.handleReservationCreated`; shares the same underlying UUID as the reservation |
| `GuestID` | OIDC email | Set in `http_booking_reservation_form.go` from session |
| `RoomID` | `room-{number}` | Hardcoded in form adapter |

---

## Serialization

All aggregates serialize to JSON with snake_case field names (see `events.go` JSON tags). `PostgresAccess` stores JSON blobs in `kv_store.value`. MCP tool outputs use `json.MarshalIndent` for human-readable responses.
