# Hotel Booking — Documentation Index

> **Primary AI retrieval entry point.** This index is the canonical starting place for anyone (human or AI) approaching this codebase.

**Generated:** 2026-04-16
**Scan mode:** initial_scan
**Scan level:** exhaustive

---

## Project Overview

- **Type:** Monolith (single Go module, single binary)
- **Primary language:** Go 1.27.1 (`go.mod` declares `go 1.27`)
- **Architecture:** Hexagonal (Ports & Adapters) + Domain-Driven Design
- **Topology:** Three bounded contexts inside one process (Reservation, Payment, Orchestration) communicating over Kafka
- **UI:** Go `html/template` SSR + htmx 2.0.10 + PWA manifest (no service worker is registered)
- **Auth:** Keycloak OIDC (dual client: session for UI, client-credentials for `/mcp`)
- **Persistence:** Two isolated PostgreSQL databases, key/value schema
- **External library:** `github.com/andygeiss/cloud-native-utils v0.5.6`

## Quick Reference

| Aspect | Value |
|--------|-------|
| Entry point | `cmd/server/main.go` |
| Router | `internal/adapters/inbound/router.go` (`Route(RouterConfig)`) |
| HTTP port | `8080` |
| Reservation DB | `localhost:5432` (`reservation_db`, `kv_store` schema) |
| Payment DB | `localhost:5433` (`payment_db`, `kv_store` schema) |
| Keycloak | `localhost:8180` (realm `local`) |
| Kafka | `localhost:9092` |
| MCP endpoint | `POST /mcp` (Bearer auth, client `hotel-booking-mcp`) |
| Command runner | `make` (see `Makefile`); no CI server |
| Ops listener | `127.0.0.1:6060` — `/healthz` and `/debug/pprof`, never proxied |
| Container runtime stage | `FROM scratch` (~5–10 MB) |
| Build optimization | Profile-Guided Optimization via the committed `cmd/server/default.pgo` |

## Generated Documentation

- [Project Overview](./project-overview.md)
- [Architecture (ARCHITECTURE.md)](./ARCHITECTURE.md) — hand-written deep dive (canonical)
- [Source Tree Analysis](./source-tree-analysis.md)
- [Component Inventory](./component-inventory.md)
- [API Contracts](./api-contracts.md)
- [Data Models](./data-models.md)
- [Development Guide](./development-guide.md)
- [Deployment Guide](./deployment-guide.md)

## Project-Level References

- [../README.md](../README.md) — user-facing quick start
- [../CLAUDE.md](../CLAUDE.md) — conventions, ubiquitous language, state machines, gotchas
- [../.env.example](../.env.example) — environment variable catalog
- [../SPEC.md](../SPEC.md) — the project brief: job, why, guardrails, done means
- [../Makefile](../Makefile) — the only command surface, and the gates
- [../docker-compose.yml](../docker-compose.yml) — dev stack definition
- [../Dockerfile](../Dockerfile) — multi-stage production image

## Bounded Contexts

| Context | Folder | Aggregate | Key Events |
|---------|--------|-----------|------------|
| Reservation | `internal/domain/reservation/` | `Reservation` (Pending → Confirmed → Active → Completed; → Cancelled) | `reservation.created/confirmed/activated/completed/cancelled` |
| Payment | `internal/domain/payment/` | `Payment` (Pending → Authorized → Captured; → Failed; → Refunded) | `payment.authorized/captured/failed/refunded` |
| Orchestration | `internal/domain/orchestration/` | (no aggregate — Saga coordinator) | Subscribes to `reservation.created`, `payment.authorized/captured/failed` |

## Getting Started

1. **Install tooling**
   ```bash
   brew install go docker-compose podman graphviz
   ```
2. **Configure environment**
   ```bash
   cp .env.example .env
   cp .keycloak.json.example .keycloak.json
   ```
3. **Start the stack**
   ```bash
   podman build -t "$USER/hotel-booking:latest" -f Dockerfile .
   docker-compose --env-file .env up -d
   ```
4. **Verify**
   - App: http://localhost:8080/ui
   - Keycloak admin: http://localhost:8180/admin (admin / admin)
5. **Iterate**
   - `make check` — every gate, before every commit
   - `make ci` — the same gates against the commit, before every push
   - `make test` / `make fmt` — the inner loop
   - `make profile` — regenerate the PGO profile

See [development-guide.md](./development-guide.md) for the full daily-driver workflow and [deployment-guide.md](./deployment-guide.md) for everything container-related.

## Using these docs as PRD input

When running a brownfield PRD or feature-planning workflow, point it at this `index.md`. The index links to every downstream doc (architecture, data models, API contracts, source tree) at a level of detail suited to AI-assisted planning.

- **UI-only feature?** Start with `component-inventory.md` (inbound section) + `api-contracts.md`.
- **Payment/Reservation flow change?** Start with `data-models.md` + `ARCHITECTURE.md` §Saga.
- **New bounded context?** Start with `ARCHITECTURE.md` §Bounded Contexts and mirror the existing layout under `internal/domain/`.
- **Infrastructure change?** Start with `deployment-guide.md`.

## Workflow State

The workflow state file is at [`./project-scan-report.json`](./project-scan-report.json). Re-running the `bmad-document-project` skill will offer to resume from this file or start fresh.
