# Project Overview

**Project:** Hotel Booking
**Type:** Go backend service (monolith)
**Architecture:** Hexagonal (Ports & Adapters) + Domain-Driven Design
**Primary Language:** Go 1.25.5
**Scan Date:** 2026-04-16
**Scan Level:** Exhaustive

---

## Purpose

A reference implementation of a hotel reservation and payment management system. It demonstrates clean Go backend structure built around Hexagonal Architecture, Domain-Driven Design tactical patterns (aggregates, value objects, domain events), and the Saga pattern for cross-context consistency via Kafka events.

The codebase is designed to be forked as a template — the directory layout, dependency wiring, and test conventions are production-ready while the domain logic is intentionally illustrative.

## Executive Summary

- **Single Go module** (`github.com/andygeiss/hotel-booking`) with a single `cmd/server` entry point.
- **Three bounded contexts** live under `internal/domain/`: `reservation`, `payment`, and `orchestration`. Each owns its aggregate, events, ports, and service.
- **Infrastructure lives under `internal/adapters/`** with inbound (HTTP handlers, event subscriber) and outbound (Postgres KV stores, mock payment gateway, mock notification, Kafka event publisher) adapters.
- **Shared kernel** (`internal/domain/shared`) contains `Money` and `ReservationID` — cross-context types without owning any context.
- **Event-driven Saga** orchestrates the booking workflow (`reservation.created` → `payment.authorized` → `payment.captured` → `reservation.confirmed`) with compensation on `payment.failed`.
- **Two isolated PostgreSQL databases** (reservation_db on 5432, payment_db on 5433) use a simple key/value schema (`kv_store` table) via `cloud-native-utils/resource.PostgresAccess`.
- **Dual OAuth clients** via Keycloak: one session-based for the HTMX UI, one client-credentials flow for the `/mcp` MCP (Model Context Protocol) endpoint.
- **Progressive Web App** support: manifest + service worker served from Go templates.
- **Profile-Guided Optimization** in the Docker build via benchmarks in `cmd/server/main_test.go`.

## Quick Reference

| Aspect | Value |
|--------|-------|
| Entry point | `cmd/server/main.go` |
| HTTP port | `8080` |
| Reservation DB | PostgreSQL on port `5432` (KV schema) |
| Payment DB | PostgreSQL on port `5433` (KV schema) |
| Auth provider | Keycloak on port `8180` (realm `local`) |
| Event broker | Kafka on port `9092` |
| Infrastructure library | `github.com/andygeiss/cloud-native-utils v0.5.6` |
| DB driver | `github.com/jackc/pgx/v5` |
| OIDC client | `github.com/coreos/go-oidc/v3` |
| Kafka client | `github.com/segmentio/kafka-go` (indirect via cloud-native-utils) |
| Task runner | `just` |
| Lint | `golangci-lint` (v2, `default: all` with noise disabled) |
| Container | Multi-stage Dockerfile → `scratch` runtime (~5-10 MB) |
| CI | GitHub Actions (`.github/workflows/ci.yml`) + Codacy coverage |
| Frontend | Go `html/template` SSR + HTMX + PWA (service worker) |

## Repository Structure

Monolith, single Go module. Two top-level groupings:

- **`cmd/server/`** — binary entry point, DI wiring, embedded static assets/templates
- **`internal/`** — hexagonal application code
  - `internal/adapters/inbound/` — HTTP handlers, event subscriber, router
  - `internal/adapters/outbound/` — Postgres KV repos (via `PostgresAccess`), mock payment gateway, mock notification service, Kafka event publisher
  - `internal/domain/reservation/` — Reservation aggregate + service + events + MCP tools
  - `internal/domain/payment/` — Payment aggregate + service + events + MCP tools
  - `internal/domain/orchestration/` — Saga (`BookingService`) + event handlers
  - `internal/domain/shared/` — shared kernel (`Money`, `ReservationID`)
- **`migrations/`** — init SQL per database (both use a simple `kv_store` schema)
- **`docs/`** — generated and hand-written architecture docs

## Architecture Type

- **Monolith** deployed as a single binary.
- Internally organized as **three bounded contexts** communicating via domain events through Kafka.
- Clean direction of dependencies: adapters depend on domain ports; the domain layer is free of infrastructure imports (only imports from `cloud-native-utils/event`, `cloud-native-utils/resource`, `cloud-native-utils/mcp` for port types).

## Links

- [Architecture (ARCHITECTURE.md)](./ARCHITECTURE.md) — full hand-written architecture document (kept as the canonical architecture doc; the BMAD workflow did not overwrite it)
- [Source Tree Analysis](./source-tree-analysis.md) — annotated directory map
- [API Contracts](./api-contracts.md) — HTTP + MCP endpoint catalog
- [Data Models](./data-models.md) — aggregates + KV schema
- [Development Guide](./development-guide.md) — local setup, build, test
- [Deployment Guide](./deployment-guide.md) — Docker, CI/CD, env vars
- [Component Inventory](./component-inventory.md) — handler/service/adapter inventory
- [CLAUDE.md](../CLAUDE.md) — project conventions, gotchas, state machines
- [README.md](../README.md) — user-facing overview and quick start
