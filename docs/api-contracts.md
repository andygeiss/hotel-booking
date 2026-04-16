# API Contracts

Full catalog of HTTP endpoints and MCP tools exposed by the server. All routes are registered in `internal/adapters/inbound/router.go` via `Route(RouterConfig)`.

---

## HTTP Endpoints

### Health / Liveness

Mounted by `web.NewServeMux` from `cloud-native-utils` (see `cmd/server/main_test.go` benchmarks for `/liveness`).

| Method | Path | Purpose | Auth | Handler |
|--------|------|---------|------|---------|
| GET | `/liveness` | Liveness probe | none | `cloud-native-utils/web` default |
| GET | `/readiness` | Readiness probe | none | `cloud-native-utils/web` default |
| GET | `/static/*` | Serves embedded assets | none | default from `web.NewServeMux` |

### PWA

| Method | Path | Purpose | Auth | Handler |
|--------|------|---------|------|---------|
| GET | `/manifest.json` | PWA manifest (rendered from `manifest.tmpl`) | none | `HttpViewManifest` |
| GET | `/sw.js` | Service worker (no-cache headers) | none | `HttpViewServiceWorker` |

### UI (HTMX / SSR)

All `/ui/*` routes except `/ui/login` and `/ui/error` are wrapped in `web.WithAuth(serverSessions, ...)` — unauthenticated sessions are redirected to `/ui/login`.

| Method | Path | Purpose | Auth | Handler |
|--------|------|---------|------|---------|
| GET | `/ui/` | Dashboard | session | `HttpViewIndex` |
| GET | `/ui/login` | OIDC login handoff | none | `HttpViewLogin` |
| GET | `/ui/error` | Error page (query params: `title`, `message`, `details`) | none | `HttpViewError` |
| GET | `/ui/reservations` | List current user's reservations | session | `HttpViewReservations` |
| GET | `/ui/reservations/new` | New reservation form | session | `HttpViewReservationForm` |
| POST | `/ui/reservations` | Create reservation (form submit) | session | `HttpCreateReservation` |
| GET | `/ui/reservations/{id}` | Reservation detail | session | `HttpViewReservationDetail` |
| POST | `/ui/reservations/{id}/cancel` | Cancel reservation | session | `HttpCancelReservation` |

Session identity is read from `r.Context()` via `web.ContextSessionID`, `web.ContextEmail`, `web.ContextName`, `web.ContextIssuer`, `web.ContextSubject`, `web.ContextVerified`. Guest identity in domain terms is the OIDC email.

### MCP (Model Context Protocol)

Mounted only when `RouterConfig.MCPServer != nil`. Bearer authentication is applied only when `RouterConfig.Verifier != nil` (required in production; nil-Verifier path is used for unit tests and benchmarks).

| Method | Path | Purpose | Auth | Handler |
|--------|------|---------|------|---------|
| POST | `/mcp` | JSON-RPC 2.0 MCP endpoint | Bearer (OIDC `MCP_CLIENT_ID`) | `web.NewMCPHandler(MCPServer).Handler()` |

---

## Form Contracts

### POST `/ui/reservations`

Parsed in `http_booking_reservation_form.go:parseReservationForm`.

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `room_id` | string | yes | Must match a known room (`room-101`, `room-102`, `room-201`, `room-202`, `room-301`) |
| `check_in` | `YYYY-MM-DD` | yes | Parsed with `2006-01-02` layout |
| `check_out` | `YYYY-MM-DD` | yes | Parsed with `2006-01-02` layout |
| `guest_name` | string | yes | — |
| `guest_email` | string | yes | — |
| `guest_phone` | string | no | — |

**Pricing (hardcoded in `getRoomPrices`):**

| Room ID | Nightly price (cents) |
|---------|-----------------------|
| `room-101`, `room-102` | 9900 |
| `room-201`, `room-202` | 14900 |
| `room-301` | 24900 |

Total amount = `nightly_price × nights`, currency `USD`.

Reservation IDs are generated via `cloud-native-utils/security.GenerateID()`; the domain type `ReservationID` does not enforce the `res-{uuid}` prefix from CLAUDE.md — that is a documentation convention.

### POST `/ui/reservations/{id}/cancel`

No body. Uses session email to verify ownership against `Reservation.GuestID`. HTMX-aware: returns `HX-Redirect: /ui/reservations` when the `HX-Request: true` header is present.

---

## MCP Tools

Registered in `cmd/server/main.go:buildMCPServer` via `reservation.RegisterTools` and `payment.RegisterTools`.

### Reservation Tools

| Tool | Description | Parameters |
|------|-------------|------------|
| `get_reservation` | Get reservation by ID | `id: string` |
| `list_reservations` | List reservations for a guest email | `guest_email: string` |
| `cancel_reservation` | Cancel reservation with reason | `id: string`, `reason: string` |
| `check_availability` | Check if a room is free for a date range | `room_id: string`, `check_in: RFC3339`, `check_out: RFC3339` |

### Payment Tools

| Tool | Description | Parameters |
|------|-------------|------------|
| `get_payment` | Get payment by ID | `id: string` |
| `capture_payment` | Capture an authorized payment | `id: string` |
| `refund_payment` | Refund a captured payment | `id: string` |

All tool results return a single text content block with either a JSON-marshaled aggregate (`get_*`, `list_*`) or a short status string (`cancel_*`, `capture_*`, `refund_*`).

### MCP Authentication Flow

```bash
TOKEN=$(curl -s -X POST "http://localhost:8180/realms/local/protocol/openid-connect/token" \
  -d "client_id=hotel-booking-mcp" \
  -d "grant_type=client_credentials" \
  -d "client_secret=<secret>" | jq -r '.access_token')

curl -X POST http://localhost:8080/mcp \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
```

---

## Domain Event Topics

Events are published to Kafka via `outbound.EventPublisher` using `event.Topic()` from each event type. All events are JSON-marshaled with snake_case field names.

| Topic | Publisher | Subscribers | Payload |
|-------|-----------|-------------|---------|
| `reservation.created` | `reservation.Service.CreateReservation` | `orchestration.handleReservationCreated` (authorizes payment) | `EventCreated {reservation_id, guest_id, room_id, check_in, check_out, total_amount}` |
| `reservation.confirmed` | `reservation.Service.ConfirmReservation` | — | `EventConfirmed {reservation_id, guest_id}` |
| `reservation.activated` | `reservation.Service.ActivateReservation` | — | `EventActivated {reservation_id}` |
| `reservation.completed` | `reservation.Service.CompleteReservation` | — | `EventCompleted {reservation_id}` |
| `reservation.cancelled` | `reservation.Service.CancelReservation` | — | `EventCancelled {reservation_id, guest_id, reason}` |
| `payment.authorized` | `payment.Service.AuthorizePayment` | `orchestration.handlePaymentAuthorized` (captures payment) | `EventAuthorized {payment_id, reservation_id, transaction_id, amount}` |
| `payment.captured` | `payment.Service.CapturePayment` | `orchestration.handlePaymentCaptured` (confirms reservation) | `EventCaptured {payment_id, reservation_id, amount}` |
| `payment.failed` | `payment.Service.AuthorizePayment` / `CapturePayment` on gateway failure | `orchestration.handlePaymentFailed` (cancels reservation — compensation) | `EventFailed {payment_id, reservation_id, error_code, error_msg}` |
| `payment.refunded` | `payment.Service.RefundPayment` | — | `EventRefunded {payment_id, reservation_id, amount}` |

Topic constants live in `internal/domain/{reservation,payment}/events.go`. Subscription happens in `internal/domain/orchestration/event_handlers.go:RegisterHandlers`.
