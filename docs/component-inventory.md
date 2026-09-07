# Component Inventory

Catalog of every handler, service, adapter, and domain type shipped by the server, grouped by hexagonal role.

---

Architecture narrative lives in [ARCHITECTURE.md](./ARCHITECTURE.md); this inventory is keyed to files and types.

---

## Inbound Adapters (`internal/adapters/inbound/`)

### Router

| Component | File | Role |
|-----------|------|------|
| `RouterConfig` struct | `router.go` | Consolidates all routing dependencies (App, Ctx, EFS, Logger, MCPServer, NewVerifier, ReservationService) |
| `Route(RouterConfig) *http.ServeMux` | `router.go` | Builds the full mux; wraps every handler with `logging.WithLogging` and auth where needed, then returns a root mux with the whole application behind `WithSecurity` |

### Middleware and ops

| Component | File | Role |
|-----------|------|------|
| `WithSecurity(http.Handler) http.Handler` | `middleware.go` | The security chain: `secureHeaders`, then `http.CrossOriginProtection` (CSRF), then a 1 MiB `http.MaxBytesHandler`. Applied by `Route`, so no handler can skip it. |
| `secureHeaders` + the `csp` constant | `middleware.go` | CSP, HSTS, `nosniff`, `Referrer-Policy`, set before the next handler runs |
| `OpsHandler(version string, ping func(context.Context) error) http.Handler` | `http_ops.go` | `/healthz` and the pprof routes for the loopback-only ops listener on `127.0.0.1:6060`. Never registered on the application mux. |
| `WithBearerAuth(VerifierFunc, *slog.Logger, http.HandlerFunc) http.HandlerFunc` | `middleware.go` | Guards `/mcp`. Builds the OIDC verifier on the first request rather than at boot, caches only success, and answers 503 while the identity provider is unreachable. |
| `AppInfo` | `app_info.go` | The name, description and version every page renders. Passed in from `main`'s `Config`, so no handler reads the environment. |

### HTTP View Handlers (SSR via `html/template`)

See [API Contracts § HTTP Endpoints](./api-contracts.md#http-endpoints) for method/path/auth/handler details. All files live under `internal/adapters/inbound/http_*.go`.

All handlers are created by a **factory function** that closes over its dependencies (`inbound.View`, `reservation.Service`). Identity arrives as an `AppInfo` parameter, not from the environment.

### Event Subscription

| Component | File | Role |
|-----------|------|------|
| `EventSubscriber` | `event_subscriber.go` | Generic topic → handler bridge using `messaging.Dispatcher`. Decodes JSON payload into a factory-created event type, invokes the domain handler, returns `messaging.MessageState`. |

### View DTOs

- `HttpViewIndexResponse` — session identity for dashboard
- `HttpViewLoginResponse`, `HttpViewErrorResponse`, `HttpViewManifestResponse`
- `ReservationListItem`, `HttpViewReservationsResponse`
- `GuestInfoView`, `ReservationDetailView`, `HttpViewReservationDetailResponse`
- `RoomOption`, `HttpViewReservationFormResponse`
- Helpers: `reservationStatusClass(status)` (maps `ReservationStatus` → bootstrap-style CSS class)
- Form parsing: `reservationFormInput`, `parseReservationForm(r)`

### Embedded Templates (`cmd/server/assets/templates/`)

- `layout.tmpl` — the shell. Invokes `title` and `main`, and renders the nav from
  `Layout.SessionID`, so one shell serves signed-in and signed-out pages alike.
- `pages/index.tmpl`, `pages/login.tmpl`, `pages/error.tmpl`
- `pages/reservations.tmpl` — also defines the `reservation-list` fragment
- `pages/reservation_form.tmpl` — also defines the `reservation-form` fragment
- `pages/reservation_detail.tmpl`

Each page is parsed into its own set with the layout: they all define `title` and `main`,
so one shared set would keep only the last page parsed. The PWA manifest is no longer a
template — it is marshalled by `encoding/json`, because `html/template` escapes for an HTML
context and would corrupt a JSON string.

Test fixtures mirror the structure under `internal/adapters/inbound/testdata/assets/templates/`.

---

## Outbound Adapters (`internal/adapters/outbound/`)

| Component | File | Role | Implements |
|-----------|------|------|------------|
| `EventPublisher` | `event_publisher.go` | JSON-encodes domain events and publishes to Kafka via `messaging.Dispatcher` | `reservation.EventPublisher`, `payment.EventPublisher` (both are aliases for `event.EventPublisher`) |
| `RepositoryAvailabilityChecker` | `repository_availability_checker.go` | Implements room availability by scanning all reservations from the repository | `reservation.AvailabilityChecker` |
| `MockNotificationService` | `mock_notification_service.go` | Logs notification events via `slog` | `orchestration.NotificationService` |
| `MockPaymentGateway` | `mock_payment_gateway.go` | In-memory simulated gateway with `ShouldFail` + `FailureRate` knobs | `payment.PaymentGateway` |

### Repositories (no local adapter file — wired directly in `main.go`)

- Reservation: `resource.NewPostgresAccess[reservation.ReservationID, reservation.Reservation](reservationDB)`
- Payment: `resource.NewPostgresAccess[payment.PaymentID, payment.Payment](paymentDB)`

Both rely on the `resource.Access[K, V]` interface from `cloud-native-utils` (`Create`, `Read`, `Update`, `Delete`, `ReadAll`).

---

## Domain — Reservation (`internal/domain/reservation/`)

### Aggregate & Value Objects

- `Reservation` (aggregate root, `aggregate.go`)
- `DateRange` (value object, `entities.go`)
- `GuestInfo` (entity inside aggregate, `entities.go`)
- Local ID types: `GuestID`, `RoomID`
- Type aliases: `ReservationID = shared.ReservationID`, `Money = shared.Money`

### State Enum

`ReservationStatus`: `StatusPending`, `StatusConfirmed`, `StatusActive`, `StatusCompleted`, `StatusCancelled`

### Errors

`ErrInvalidDateRange`, `ErrCheckInPast`, `ErrMinimumStay`, `ErrInvalidStateTransition`, `ErrCannotCancelNearCheckIn`, `ErrCannotCancelActive`, `ErrCannotCancelCompleted`, `ErrAlreadyCancelled`, `ErrNoGuests`

### Service (`service.go`)

`Service` with dependencies (`reservationRepo`, `availabilityChecker`, `publisher`). Public methods:

- `CreateReservation` — availability check → aggregate → persist → publish `reservation.created`
- `ConfirmReservation` — load → `Confirm()` → update → publish `reservation.confirmed`
- `CancelReservation` — load → `Cancel(reason)` → update → publish `reservation.cancelled`
- `ActivateReservation` / `CompleteReservation` — similar, publish `reservation.activated` / `reservation.completed`
- `GetReservation`, `ListReservationsByGuest`
- `ConfirmReservationOnPaymentCaptured`, `CancelReservationOnPaymentFailed` — event-driven entry points
- Helper: `NewEventCreatedFromValues(...)`

### Events (`events.go`)

Types: `EventCreated`, `EventConfirmed`, `EventActivated`, `EventCompleted`, `EventCancelled`. Topic constants: `EventTopicCreated`, `EventTopicConfirmed`, `EventTopicActivated`, `EventTopicCompleted`, `EventTopicCancelled` (all `reservation.*`).

### Ports (`ports.go`)

- `ReservationRepository = resource.Access[ReservationID, Reservation]`
- `AvailabilityChecker` (`IsRoomAvailable`, `GetOverlappingReservations`)
- `EventPublisher = event.EventPublisher`

### MCP Tools (`tools.go`)

`RegisterTools(server, service, checker)` registers:

- `get_reservation`
- `list_reservations` (by guest email)
- `cancel_reservation` (requires reason)
- `check_availability` (room + RFC3339 dates)

---

## Domain — Payment (`internal/domain/payment/`)

### Aggregate & Entity

- `Payment` (aggregate root, `aggregate.go`)
- `PaymentAttempt` (entity collection inside aggregate, `entities.go`)
- Local ID: `PaymentID`
- Type aliases: `ReservationID = shared.ReservationID`, `Money = shared.Money`

### State Enum

`PaymentStatus`: `StatusPending`, `StatusAuthorized`, `StatusCaptured`, `StatusFailed`, `StatusRefunded`

### Errors

`ErrInvalidPaymentTransition`, `ErrAlreadyAuthorized`, `ErrNotAuthorized`, `ErrAlreadyCaptured`, `ErrNotCaptured`, `ErrAlreadyRefunded`, `ErrCannotRefund`

### Service (`service.go`)

- `AuthorizePayment` — creates aggregate → gateway authorize → persist → publish `payment.authorized` (or `payment.failed` on gateway error)
- `CapturePayment` — load → gateway capture → update → publish `payment.captured` (or `payment.failed`)
- `RefundPayment` — load → gateway refund → update → publish `payment.refunded`
- `GetPayment`
- Event-driven entries: `AuthorizePaymentForReservation`, `CapturePaymentOnAuthorization`
- Convenience: `NewMoney(amount, currency)` (re-export of `shared.NewMoney`)

### Events (`events.go`)

`EventAuthorized`, `EventCaptured`, `EventFailed`, `EventRefunded`. Topics: `payment.authorized`, `payment.captured`, `payment.failed`, `payment.refunded`.

### Ports (`ports.go`)

- `PaymentRepository = resource.Access[PaymentID, Payment]`
- `PaymentGateway` (`Authorize`, `Capture`, `Refund`)
- `EventPublisher = event.EventPublisher`

### MCP Tools (`tools.go`)

- `get_payment`
- `capture_payment`
- `refund_payment`

---

## Domain — Orchestration (`internal/domain/orchestration/`)

### Services

| Component | File | Role |
|-----------|------|------|
| `BookingService` | `booking_service.go` | Saga coordinator. `InitiateBooking` starts the event-driven flow; `CompleteBooking` runs the saga synchronously (useful for tests). Event callbacks: `OnPaymentAuthorized`, `OnPaymentCaptured`, `OnPaymentFailed`. `CancelBookingWithRefund` handles manual cancellation + notification. |
| `EventHandlers` | `event_handlers.go` | Subscribes to `reservation.created`, `payment.authorized`, `payment.captured`, `payment.failed` via `messaging.Dispatcher`. Maps events to `BookingService` methods. |

### Port

- `NotificationService` interface (`ports.go`): `SendReservationConfirmation`, `SendCancellationNotice`, `SendPaymentReceipt`

---

## Shared Kernel (`internal/domain/shared/`)

| Type | Kind | Purpose |
|------|------|---------|
| `ReservationID` | named string | Cross-context identity |
| `Money` | value object | `{Currency, Amount}` with `NewMoney`, `FormatAmount` |

---

## Entry Point (`cmd/server/`)

| Component | File | Role |
|-----------|------|------|
| `main` | `main.go` | DI wiring, Postgres pools, Kafka dispatcher, OIDC provider + verifier, MCP server, router, HTTP server + graceful shutdown |
| `buildMCPServer` | `main.go` | Constructs `*mcp.Server` and registers reservation + payment tools |
| `efs` | `main.go` (`//go:embed assets`) | Embedded filesystem for static assets and templates |
| Benchmarks | `main_test.go` | PGO profiling coverage: `/liveness`, `/static/css/base.css`, `/ui/login`, `/mcp tools/list`, domain lifecycles |

---

## Third-Party Building Blocks (`github.com/andygeiss/cloud-native-utils v0.5.6`)

Used throughout the codebase:

- `env.Get(key, default)` — typed env lookups (used for every config value)
- `logging.NewJsonLogger` — structured JSON logger
- `logging.WithLogging(logger, handler)` — request logging middleware
- `messaging.Dispatcher`, `messaging.NewExternalDispatcher`, `messaging.NewInternalDispatcher`, `messaging.Message`, `messaging.MessageState`
- `resource.Access[K, V]`, `resource.NewPostgresAccess[K, V](db)` — key/value aggregate repository
- `event.Event`, `event.EventPublisher` — domain event contract
- `mcp.Server`, `mcp.NewTool`, `mcp.NewObjectSchema`, `mcp.NewStringProperty`, `mcp.ContentBlock`, `mcp.NewTextContent`
- `html/template` (standard library) — SSR, via `inbound.View`
- `web.NewServeMux`, `web.NewServer`, `web.WithAuth`, `web.WithBearerAuth`, `web.NewMCPHandler`, `web.ContextSessionID/Email/Name/Issuer/Subject/Verified`
- `service.Context`, `service.RegisterOnContextDone`, `service.Wrap`
- `security.GenerateID()` — random ID generator used for reservation IDs
