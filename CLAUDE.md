# Hotel Booking System

## Documentation Policy

> **Where do I document X?** See [`docs/README.md`](./docs/README.md) for the canonical ownership map across `CLAUDE.md`, `docs/ARCHITECTURE.md`, and the reference catalogs in `docs/`.

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
- [ ] New terms → CLAUDE.md § Ubiquitous Language?
- [ ] Roadmap item completed → Mark as [x] in CLAUDE.md § Roadmap?
- [ ] New pattern/gotcha → CLAUDE.md § Gotchas?
- [ ] New HTTP handler → `docs/api-contracts.md`?
- [ ] New MCP tool → `docs/api-contracts.md`?
- [ ] New domain error → CLAUDE.md § Domain Errors?
- [ ] New state transition → `docs/ARCHITECTURE.md` aggregate state machine?
- [ ] New environment variable → `docs/deployment-guide.md`?

### Documentation Update Matrix

| Change Type | Canonical Doc | Also touch CLAUDE.md? | README.md |
|-------------|---------------|-----------------------|-----------|
| New HTTP handler | `docs/api-contracts.md` | no | - |
| New MCP tool | `docs/api-contracts.md` | no | - |
| New domain error | CLAUDE.md § Domain Errors | yes | - |
| New state transition | `docs/ARCHITECTURE.md` § aggregate | no | - |
| Feature complete | CLAUDE.md § Roadmap | yes (mark `[x]`) | Features section |
| New gotcha | CLAUDE.md § Gotchas | yes | - |
| Architectural decision | CLAUDE.md § Decisions | yes | - |
| New environment variable | `docs/deployment-guide.md` | no | Setup |
| New event topic | `docs/api-contracts.md` § Domain Event Topics | no | - |

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
| PaymentID | `pay-{uuid}` (stripping `res-` prefix from ReservationID) | `pay-abc123` |
| GuestID | Email address | `john@example.com` |
| RoomID | `room-{number}` | `room-101` |

---

## State Machines

See [ARCHITECTURE.md § Reservation Aggregate](./docs/ARCHITECTURE.md#reservation-aggregate) and [§ Payment Aggregate](./docs/ARCHITECTURE.md#payment-aggregate) for the state diagrams and transition rules.

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

See [API Contracts § Domain Event Topics](./docs/api-contracts.md#domain-event-topics) for the canonical table with publisher, subscribers, and payload schemas.

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
      middleware.go    secureHeaders, the CSP constant, WithSecurity chain
      router.go        Central HTTP routing
      http_ops.go      /healthz + pprof for the loopback ops listener
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
Makefile               The only command surface (baseline stack/makefile.md)
SPEC.md                The project brief: job, why, guardrails, done means
```

---

## Commands

`make` is the only command surface, copied from the baseline's `stack/makefile.md`.
There is no CI server: `make check` before every commit, `make ci` before every push.

```bash
# The two that matter
make check           # Every gate against the working tree (the default target)
make ci              # The same gates against the commit, in a clean copy

# Inner loop
make run             # Run the server locally (loads .env if present)
make test            # go test -race -shuffle=on ./...
make fmt             # goimports + go fix — read the diff before committing it
make build           # Release-shaped binary into bin/
make clean           # Remove bin/
make profile         # CPU profile for the -pgo Docker build
```

The local container stack is not a Make target (the baseline bans Docker targets).
Use `docker-compose --env-file .env up -d` — see
[Development Guide](./docs/development-guide.md).

**The eight gates `check` runs, in order:** format, vet, fix, staticcheck, govulncheck,
tidy, test, static build. A red `govulncheck` means bump the dependency, never silence
the check.

---

## Environment Variables

See [Deployment Guide § Environment Variables](./docs/deployment-guide.md#environment-variables) for the complete reference (application identity, HTTP server, OIDC/Keycloak, databases, Kafka, resilience knobs).

**Gotchas (not in the guide):**
- **Kafka broker config** — use `localhost:9092` for local dev, `kafka:9092` inside Docker compose.
- **OIDC issuer** — `http://localhost:8180/realms/local` must resolve from both the browser and the app container; `docker-compose.yml` uses `extra_hosts: ["localhost:host-gateway"]` for this.

---

## MCP Tools

See [API Contracts § MCP Tools](./docs/api-contracts.md#mcp-tools) for the canonical tool catalog (parameters and descriptions) and [§ MCP Authentication Flow](./docs/api-contracts.md#mcp-authentication-flow) for the Bearer token flow.

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
| Make, no CI server | One gate list, run on the developer's machine; `make ci` proves it on the commit |
| Security chain in `Route` | A handler cannot be registered outside CSP, CSRF and the body cap |
| Loopback ops listener | `/healthz` and pprof are unreachable from the proxy, which is their only access control |

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
- [x] Engineering baseline: Go 1.27 pin, Makefile, security headers, CSRF, ops listener
- [ ] Email notifications
- [ ] Calendar integration
- [ ] Admin dashboard

### Open baseline work

Boxes in the baseline's `checklists/web-application.md` that this project does not check
yet. These are open tasks, not waivers — what is genuinely waived is in
[README § Baseline deviations](./README.md#baseline-deviations).

- [ ] **Delete `GET /sw.js`.** Nothing registers a service worker any more. The route
      stays only long enough to hand an unregistering worker to browsers that still
      hold the old one.
- [ ] **Move config into `main`.** Eight files under `internal/adapters/inbound` call
      `os.Getenv` for `APP_NAME` and `APP_DESCRIPTION`, which is why every handler test
      needs `t.Setenv`. The baseline wants one `Config` struct parsed in `main`
      (`patterns/go-config.md`).
- [ ] **Give shutdown a deadline.** `srv.Shutdown(context.Background())` waits forever
      on an in-flight request. The baseline calls for a fresh ~10 s budget.
- [ ] **Ship the PGO profile, or drop `-pgo`.** `.cpuprofile.pprof` is gitignored, so
      the Dockerfile's `-pgo` build fails on a clean checkout until `make profile` runs.
- [ ] **Dual-mode htmx responses.** `Vary: HX-Request`, `hx-push-url`, and the
      fragment-or-full-page test are not in place yet
      (`patterns/htmx-server-rendering.md`).
- [ ] **`DESIGN.md`.** The three stylesheets under `cmd/server/assets/static/css` have
      no design file to stay in lockstep with (`patterns/design-system.md`).

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

10. **PaymentID convention** - Derive from ReservationID by stripping the `res-` prefix: `fmt.Sprintf("pay-%s", strings.TrimPrefix(string(reservationID), "res-"))`. Both IDs share the same UUID → grep-friendly traceability without a double prefix.

11. **Database per context** - Reservation and Payment use separate PostgreSQL instances. Never cross-query.

12. **Kafka broker config** - Use `localhost:9092` for local dev, `kafka:9092` inside Docker compose.

13. **No inline JavaScript, ever** - The CSP carries no `'unsafe-inline'`, so an inline
    `<script>` or `<style>` is dead on arrival in the browser. Put behaviour in htmx
    attributes and rules in a stylesheet. Adding `'unsafe-inline'` to make one thing
    work is banned: it is a tier-1 rule in the baseline.

14. **Security wraps the router, not `main`** - `Route` returns its mux already wrapped
    in `WithSecurity`. Never register a handler outside it, and never hand `main` an
    unwrapped mux.

15. **Ops endpoints stay off the app mux** - `/healthz` and `/debug/pprof` belong to
    `OpsHandler` on `127.0.0.1:6060`. Being unreachable from the proxy is their only
    access control, so putting either on the application listener exposes them.

16. **The gates are yours to run** - There is no CI workflow. `make check` before a
    commit, `make ci` before a push, and read `make ci`'s `go version` line: it is the
    only record of which toolchain ran.
