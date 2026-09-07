# Spec

The project brief. Every task brief is a delta against this file.

## Job

A guest holds a room, pays for it, and ends up with a confirmed booking — from the web
UI, or from an AI agent through the MCP endpoint. A developer can read this repository
as the worked example of Domain-Driven Design and hexagonal architecture in Go.

## Why

Reservations and payments live in separate bounded contexts with separate databases, so
nothing can wrap them in one transaction. A failed payment has to roll the reservation
back on its own, and that gap is where booking systems leak reservations that are held
forever and paid never. The saga in `internal/domain/orchestration` closes it.

The second reader is a developer who wants DDD in Go and finds blog posts instead of a
repository that builds, tests, and deploys.

## Guardrails

The [engineering baseline](https://github.com/andygeiss/baseline) is the standing
guardrail; this file does not restate it.

- What this project waives, and what it meets by a different route, is in
  [README.md § Baseline deviations](./README.md#baseline-deviations). Nothing gets
  waived without an entry there.
- Architectural decisions and their reasons are in
  [CLAUDE.md § Decisions](./CLAUDE.md#decisions).
- Where each kind of documentation belongs is in [docs/README.md](./docs/README.md).
  Add a table, not a second copy of one.

## Done means

- `make ci` is green on the commit being pushed. Nothing else runs the gates.
- The boxes in the baseline's `checklists/web-application.md` are walked: every one is
  either checked, or waived on the record in the README.
- The documentation checklist in [CLAUDE.md](./CLAUDE.md) is walked.
