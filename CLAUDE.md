# Hotel Booking System

## Documentation Policy

### Update CLAUDE.md when changes affect:
1. **Architecture** - New patterns, packages, dependencies
2. **API surface** - New handlers, routes, MCP tools
3. **Domain model** - New entities, events, errors, state transitions
4. **Conventions** - New naming rules, anti-patterns, gotchas
5. **Decisions** - Architectural trade-offs or technology choices

### Update README.md when changes affect:
1. **User-facing behavior** - New features, commands, endpoints
2. **Setup instructions** - New prerequisites, environment variables

### Documentation Checklist (before commit):
- [ ] New terms → Ubiquitous Language?
- [ ] Roadmap item completed → Mark as [x]?
- [ ] New pattern/gotcha → Add to relevant section?
- [ ] New HTTP handler → Add to Quick Reference?
- [ ] New MCP tool → Add to MCP Tools section?
- [ ] New environment variable → Add to Environment Variables?

### Documentation Update Matrix

| Change Type | CLAUDE.md Section | README.md |
|-------------|-------------------|-----------|
| New HTTP handler | Quick Reference | - |
| New MCP tool | MCP Tools | - |
| New domain error | Domain Errors | - |
| New state transition | State Machines | - |
| Feature complete | Roadmap [x] | Features section |
| New gotcha | Gotchas | - |
| Architectural decision | Decisions | - |
| New environment variable | Environment Variables | Setup |

---

## Ubiquitous Language

| Term | Definition |
|------|------------|
| Reservation | A held room before payment confirmation |
| Booking | A confirmed and paid reservation |
| Payment | A financial transaction tied to a reservation |
| Guest | A person associated with a reservation |
| GuestInfo | Value object containing guest name, email, phone |
| DateRange | Check-in to check-out period |
| Money | Value object with amount and currency |
| Saga | Cross-context workflow with automatic compensation |
| Authorization | Pre-approval for payment capture |
| Capture | Final collection of authorized payment |
| Refund | Return of captured payment |
| Compensation | Rollback action when saga fails |

### Identifiers

| Type | Format | Example |
|------|--------|---------|
| ReservationID | `res-{uuid}` | `res-abc123` |
| PaymentID | `pay-{reservationID}` | `pay-res-abc123` |
| GuestID | Email address | `john@example.com` |
| RoomID | `room-{number}` | `room-101` |

---

## State Machines

### Reservation States

```
[Pending] ──→ [Confirmed] ──→ [Active] ──→ [Completed]
    │             │              │
    ▼             ▼              ▼
[Cancelled]  [Cancelled]   [Cancelled]
```

| Transition | Trigger | Validation |
|------------|---------|------------|
| Pending → Confirmed | Payment captured | - |
| Confirmed → Active | Check-in | - |
| Active → Completed | Check-out | - |
| * → Cancelled | User request / Payment failed | 24h before check-in (user), anytime (payment failure) |

### Payment States

```
[Pending] ──→ [Authorized] ──→ [Captured]
    │              │               │
    ▼              ▼               ▼
[Failed]      [Failed]       [Refunded]
```

| Transition | Trigger | External Call |
|------------|---------|---------------|
| Pending → Authorized | Gateway approval | PaymentGateway.Authorize |
| Authorized → Captured | Orchestration | PaymentGateway.Capture |
| Captured → Refunded | Admin action | PaymentGateway.Refund |
| * → Failed | Gateway rejection | - |

---

## Event Flow (Saga Pattern)

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Happy Path                                   │
└─────────────────────────────────────────────────────────────────────┘

ReservationCreated ──→ PaymentAuthorized ──→ PaymentCaptured ──→ ReservationConfirmed
        │                     │                    │
        │              (orchestration)      (orchestration)
        ▼                     ▼                    ▼
   payment ctx          booking svc           booking svc


┌─────────────────────────────────────────────────────────────────────┐
│                     Compensation Path                               │
└─────────────────────────────────────────────────────────────────────┘

ReservationCreated ──→ PaymentFailed ──→ ReservationCancelled
        │                    │
        │              (compensation)
        ▼                    ▼
   payment ctx          booking svc
```

### Event Topics

| Topic | Publisher | Subscribers |
|-------|-----------|-------------|
| `reservation.created` | Reservation Service | Payment Service |
| `payment.authorized` | Payment Service | Orchestration |
| `payment.captured` | Payment Service | Orchestration |
| `payment.failed` | Payment Service | Orchestration (compensation) |
| `reservation.confirmed` | Reservation Service | - |
| `reservation.cancelled` | Reservation Service | - |

---

## Project Structure

```
assets/
  static/              CSS, JS, images
  templates/           HTML templates (Go templates)
cmd/
  server/              HTTP server entry point
    main.go            Wiring, DI, server startup
    main_test.go       Integration benchmarks (PGO)
docs/
  ARCHITECTURE.md      Detailed architecture docs
internal/
  adapters/
    inbound/           HTTP handlers, router, RouterConfig
      router.go        Central HTTP routing
      http_*.go        One handler per file
    outbound/          Repository, event publisher, gateway mocks
      event_publisher.go
      mock_*.go
  domain/
    orchestration/     Saga coordination
      booking_service.go
      event_handlers.go
    payment/           Payment bounded context
      aggregate.go     Payment state machine
      service.go       Application service
      tools.go         MCP tool definitions
      events.go        Event types and topics
    reservation/       Reservation bounded context
      aggregate.go     Reservation state machine
      service.go       Application service
      tools.go         MCP tool definitions
      events.go        Event types and topics
      value_objects.go DateRange, GuestInfo
    shared/            Shared kernel
      identifiers.go   ReservationID type
      money.go         Money value object
      events.go        Base event types
migrations/
  payment/             Payment DB schema
  reservation/         Reservation DB schema
```

---

## Commands

```bash
# Build and run
just build           # Build server binary to bin/
just run             # Run server locally (uses .env)
just up              # Start Docker services (Postgres, Keycloak, Kafka)
just down            # Stop Docker services

# Quality
just lint            # Run golangci-lint
just test            # Run all tests with coverage
just profile         # Run benchmarks for PGO profiling

# Development
just serve           # Build and run in one step
```

---

## Environment Variables

### Application

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_NAME` | Display name for UI | `Hotel Booking` |
| `APP_SHORTNAME` | Docker tags, container names | `hotel-booking` |
| `APP_VERSION` | Version for PWA cache busting | `1.0.0` |
| `PORT` | HTTP server port | `8080` |

### OIDC / Keycloak

| Variable | Description | Default |
|----------|-------------|---------|
| `OIDC_ISSUER` | Keycloak realm URL | `http://localhost:8180/realms/local` |
| `OIDC_CLIENT_ID` | OAuth client for sessions | `hotel-booking` |
| `OIDC_CLIENT_SECRET` | Client secret (use placeholder) | `CHANGE_ME_LOCAL_SECRET` |
| `OIDC_REDIRECT_URL` | Callback after auth | `http://localhost:8080/auth/callback` |
| `MCP_CLIENT_ID` | OAuth client for MCP | `hotel-booking-mcp` |

### Reservation Database

| Variable | Description | Default |
|----------|-------------|---------|
| `RESERVATION_DB_HOST` | PostgreSQL host | `localhost` |
| `RESERVATION_DB_PORT` | PostgreSQL port | `5432` |
| `RESERVATION_DB_USER` | Database user | `reservation` |
| `RESERVATION_DB_PASSWORD` | Database password | `reservation_secret` |
| `RESERVATION_DB_NAME` | Database name | `reservation_db` |

### Payment Database

| Variable | Description | Default |
|----------|-------------|---------|
| `PAYMENT_DB_HOST` | PostgreSQL host | `localhost` |
| `PAYMENT_DB_PORT` | PostgreSQL port | `5433` |
| `PAYMENT_DB_USER` | Database user | `payment` |
| `PAYMENT_DB_PASSWORD` | Database password | `payment_secret` |
| `PAYMENT_DB_NAME` | Database name | `payment_db` |

### Kafka

| Variable | Description | Default |
|----------|-------------|---------|
| `KAFKA_BROKERS` | Broker addresses | `localhost:9092` |
| `KAFKA_CONSUMER_GROUP_ID` | Consumer group | `test-group` |

### Server Timeouts

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_READ_TIMEOUT` | Request read timeout | `5s` |
| `SERVER_WRITE_TIMEOUT` | Response write timeout | `5s` |
| `SERVER_IDLE_TIMEOUT` | Idle connection timeout | `5s` |
| `SERVER_READ_HEADER_TIMEOUT` | Header read timeout | `5s` |

---

## MCP Tools

### Reservation Tools

| Tool | Description | Parameters |
|------|-------------|------------|
| `get_reservation` | Get reservation by ID | `id` |
| `list_reservations` | List reservations by guest email | `guest_email` |
| `cancel_reservation` | Cancel a reservation | `id`, `reason` |
| `check_availability` | Check room availability | `room_id`, `check_in`, `check_out` |

### Payment Tools

| Tool | Description | Parameters |
|------|-------------|------------|
| `get_payment` | Get payment by ID | `id` |
| `capture_payment` | Capture authorized payment | `id` |
| `refund_payment` | Refund captured payment | `id` |

### MCP Authentication

```bash
# Get access token
TOKEN=$(curl -s -X POST "http://localhost:8180/realms/local/protocol/openid-connect/token" \
  -d "client_id=hotel-booking-mcp" \
  -d "grant_type=client_credentials" \
  -d "client_secret=<secret>" | jq -r '.access_token')

# Call MCP endpoint
curl -X POST http://localhost:8080/mcp \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
```

---

## Domain Errors

### Reservation Errors

| Error | When |
|-------|------|
| `ErrInvalidDateRange` | Check-out not after check-in |
| `ErrCheckInPast` | Check-in date in the past |
| `ErrMinimumStay` | Less than 1 night |
| `ErrInvalidStateTransition` | Invalid state change |
| `ErrCannotCancelNearCheckIn` | Cancel within 24h of check-in |
| `ErrCannotCancelActive` | Cancel active reservation |
| `ErrCannotCancelCompleted` | Cancel completed reservation |
| `ErrAlreadyCancelled` | Already cancelled |
| `ErrNoGuests` | No guests provided |

### Payment Errors

| Error | When |
|-------|------|
| `ErrInvalidPaymentTransition` | Invalid state change |
| `ErrAlreadyAuthorized` | Already authorized |
| `ErrNotAuthorized` | Capture without authorization |
| `ErrAlreadyCaptured` | Already captured |
| `ErrNotCaptured` | Refund without capture |
| `ErrAlreadyRefunded` | Already refunded |
| `ErrCannotRefund` | Refund non-captured payment |

---

## Patterns Reference

### Handler Factory Pattern

Handlers are created via factory functions that close over dependencies:

```go
func HttpViewReservations(e *templating.Engine, svc *reservation.Service) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        reservations, _ := svc.ListReservations(r.Context())
        e.Render(w, "reservations.tmpl", reservations)
    }
}
```

### RouterConfig Pattern

All HTTP routing dependencies consolidated in a struct:

```go
mux := inbound.Route(inbound.RouterConfig{
    Ctx:                ctx,
    EFS:                efs,
    Logger:             logger,
    ReservationService: reservationService,
    MCPServer:          mcpServer,  // nil disables /mcp endpoint
    Verifier:           verifier,   // Required if MCPServer is set
})
```

### Event Builder Pattern

Domain events use option functions for flexible construction:

```go
event := reservation.NewEventCreated(
    reservation.WithReservationID(id),
    reservation.WithGuestID(guestID),
    reservation.WithTotalAmount(amount),
)
```

### Key/Value Storage Pattern

Aggregates stored via `resource.PostgresAccess[K, V]` from cloud-native-utils:

```go
repo := resource.NewPostgresAccess[reservation.ReservationID, reservation.Reservation](db)
```

---

## Testing Conventions

### Naming

```
Test_{Component}_{Scenario}_Should_{Result}
```

Examples:
- `Test_Reservation_Cancel_Should_Fail_When_Already_Cancelled`
- `Test_Route_MCP_Endpoint_Without_MCPServer_Should_Return_404`

### Patterns

```go
func Test_Something(t *testing.T) {
    // Arrange
    t.Setenv("APP_NAME", "TestApp")
    svc := createTestService(t)

    // Act
    result, err := svc.DoSomething(ctx, input)

    // Assert
    assert.That(t, "error must be nil", err, nil)
    assert.That(t, "result must match", result.Status, "confirmed")
}
```

### Test Helpers

- Use `t.Helper()` in helper functions
- Use `httptest.NewRecorder()` for HTTP tests
- Mock repositories implement full interface
- Use `t.Setenv()` for environment variables (auto-cleanup)

---

## Decisions

| Decision | Rationale |
|----------|-----------|
| Hexagonal architecture | Testable domain, clean separation of concerns |
| Event-driven Saga | Cross-context consistency without distributed transactions |
| Key/Value storage | Aggregate-friendly, schema-less persistence |
| HTMX + SSR | Simpler code, progressive enhancement, no JS build step |
| RouterConfig struct | Consolidates routing dependencies, optional MCP |
| Dual OAuth clients | Session-based for web, Bearer for MCP (machine-to-machine) |
| Handler factory | Closure-based DI, testable handlers |
| Separate databases | Bounded context isolation, independent scaling |
| Kafka for events | Durable event streaming, replay capability |

---

## Roadmap

- [x] Domain model (Reservation, Payment)
- [x] Event-driven Saga orchestration
- [x] HTTP handlers (HTMX/SSR)
- [x] MCP tools integration
- [x] OAuth 2.1 authentication (Keycloak)
- [x] RouterConfig refactoring
- [x] Separate databases per bounded context
- [x] Kafka event streaming
- [ ] Email notifications
- [ ] Calendar integration
- [ ] Admin dashboard

---

## Project-Specific Gotchas

1. **State machine validation** - Aggregates validate transitions; service layer orchestrates. Don't bypass aggregate methods.

2. **Event topic constants** - Always use constants from `events.go` (e.g., `reservation.EventTopicCreated`). Never hardcode topic strings.

3. **Saga compensation** - `payment.failed` triggers automatic `ReservationCancelled`. Don't manually cancel after payment failure.

4. **RouterConfig nil checks** - `MCPServer: nil` disables `/mcp` endpoint. No auth needed if MCP disabled.

5. **DateRange validation** - Check-out must be after check-in. Minimum 1 night. Check-in cannot be in the past.

6. **Money immutability** - `shared.Money` is a value object. Create new instances instead of mutating.

7. **Test environment variables** - Always use `t.Setenv()` for `APP_NAME`, `APP_DESCRIPTION`. Tests fail without them.

8. **MCP auth in tests** - Pass `nil` Verifier for unit tests. Only integration tests need real auth.

9. **Template paths** - Must match `assets/templates/*.tmpl` pattern. Embedded via `//go:embed assets`.

10. **PaymentID convention** - Always derive from ReservationID: `pay-{reservationID}`. Enables correlation.

11. **Database per context** - Reservation and Payment use separate PostgreSQL instances. Never cross-query.

12. **Kafka broker config** - Use `localhost:9092` for local dev, `kafka:9092` inside Docker compose.
