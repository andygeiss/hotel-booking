# Documentation Map

This folder holds the project's hand-maintained documentation. Read this file first if you are about to add or update docs — it tells you where each topic belongs.

## The three layers

| Layer | File(s) | Owns |
|-------|---------|------|
| **Conventions & decisions** (project root) | [`CLAUDE.md`](../CLAUDE.md) | Ubiquitous language, identifier formats, domain errors, decisions, roadmap, project-specific gotchas |
| **Narrative architecture** | [`ARCHITECTURE.md`](./ARCHITECTURE.md) | Hexagonal/DDD principles, aggregate state machines (with diagrams), saga walkthrough, recipes (add event/handler/context), router/MCP/security narrative |
| **Reference catalogs** | the BMAD-bootstrapped files in `docs/` | Exhaustive lookup tables (HTTP routes, MCP tools, event topics, env vars, source tree, component inventory, dev/deployment guides) |

All three layers are **hand-maintained**. The BMAD `bmad-document-project` skill was used once to bootstrap the reference catalogs; from now on they are edited like any other doc. Re-running the skill is exceptional (e.g. after a major refactor) — when you do, treat it as a regenerate-and-merge, not a sync.

## Where do I document X?

| If you are adding / changing… | Update this file |
|-------------------------------|------------------|
| HTTP handler, route, form contract | [`api-contracts.md`](./api-contracts.md) |
| MCP tool (new, removed, parameter change) | [`api-contracts.md`](./api-contracts.md) § MCP Tools |
| Kafka event topic or payload schema | [`api-contracts.md`](./api-contracts.md) § Domain Event Topics |
| Aggregate, value object, state transition | [`ARCHITECTURE.md`](./ARCHITECTURE.md) § Domain Layer (also `data-models.md` for the type signature) |
| Saga / compensation flow | [`ARCHITECTURE.md`](./ARCHITECTURE.md) § Saga Pattern Implementation |
| Environment variable | [`deployment-guide.md`](./deployment-guide.md) (canonical); also `.env.example` |
| Env-var **gotcha** (e.g. compose-vs-host hostnames) | [`CLAUDE.md`](../CLAUDE.md) § Environment Variables (gotchas only) |
| Container, compose, Dockerfile, CI step | [`deployment-guide.md`](./deployment-guide.md) |
| Source tree change (new package, new directory) | [`source-tree-analysis.md`](./source-tree-analysis.md) |
| Component inventory (handler, adapter, service file) | [`component-inventory.md`](./component-inventory.md) |
| Build/test/run command, debugging tip | [`development-guide.md`](./development-guide.md) |
| Domain error sentinel | [`CLAUDE.md`](../CLAUDE.md) § Domain Errors |
| Architectural decision | [`CLAUDE.md`](../CLAUDE.md) § Decisions |
| Project-wide gotcha | [`CLAUDE.md`](../CLAUDE.md) § Project-Specific Gotchas |
| Ubiquitous-language term | [`CLAUDE.md`](../CLAUDE.md) § Ubiquitous Language |
| Identifier convention (format, prefix) | [`CLAUDE.md`](../CLAUDE.md) § Identifiers |
| User-facing feature, setup step | [`README.md`](../README.md) (project root) |

If a change spans more than one layer, prefer to put the full content in **one** canonical file and add a short cross-link from the others. Do not duplicate tables.

## Cross-link rules

- Use relative paths and lowercased GitHub-style anchors (e.g. `./api-contracts.md#mcp-tools`).
- When you add a new `## Heading` to a canonical file, the anchor is the heading text lowercased with spaces → `-` and non-alphanumerics dropped.
- Do not introduce a new redundant table without a real reason. If you find one slipping in during review, push back: pick a single owner.

## Generated artifact

[`project-scan-report.json`](./project-scan-report.json) is the BMAD workflow's local state file. It is gitignored — do not commit. Delete it freely if you want a fresh skill run.

## File listing

| File | Status |
|------|--------|
| `README.md` (this file) | hand-written ownership map |
| `ARCHITECTURE.md` | hand-written; canonical for narrative + state machines + recipes |
| `api-contracts.md` | hand-maintained reference; canonical for HTTP routes, MCP tools, event topics |
| `data-models.md` | hand-maintained reference; canonical for aggregate type signatures |
| `component-inventory.md` | hand-maintained reference; canonical for adapter/service inventory |
| `source-tree-analysis.md` | hand-maintained reference; canonical for directory tree |
| `development-guide.md` | hand-maintained reference; canonical for dev workflow |
| `deployment-guide.md` | hand-maintained reference; canonical for env vars + Docker + CI |
| `index.md` | hand-maintained navigation hub |
| `project-overview.md` | hand-maintained executive summary |
| `project-scan-report.json` | generated; gitignored |
