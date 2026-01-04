# AGENT

This document specifies how a README maintainer agent should understand and work with the `go-ddd-hex-starter` repository. It complements `CONTEXT.md` (architectural constraints) and `README.md` (human-facing documentation).

***

## 1. Purpose of this agent

The maintainer agent is responsible for:

- Keeping `README.md` aligned with the actual codebase, directory structure, and workflows.
- Reflecting architectural and coding conventions defined in `CONTEXT.md` without divergence.
- Presenting the repository as a reusable Go template for humans and coding agents.

The agent must never invent commands, files, or features that are not present in the repository.

***

## 2. Source files to treat as ground truth

When updating documentation, the agent must treat the following files as authoritative:

- `CONTEXT.md`: project architecture, conventions, and agent rules.
- `README.md`: current public-facing description, features, and usage.
- `VENDOR.md`: documentation for `cloud-native-utils` and external patterns.
- `go.mod`, `.justfile`, and `cmd/`, `internal/` packages: actual implementation and workflows.

`CONTEXT.md` always has higher priority than `README.md` when describing architecture or conventions.

***

## 3. README responsibilities and structure

When generating or updating `README.md`, the agent must:

- Preserve the main title: `Go DDD Hexagonal Starter`.
- Keep the badge block at the top with:
  - Go Reference (pkg.go.dev).
  - Go Report Card (goreportcard.com).
  - License.
  - Release.
  - Test coverage.
- Follow this section structure (unless the repository itself changes direction):
  1. Project title and one-line value proposition.
  2. Overview / Motivation.
  3. Key Features.
  4. Architecture overview.
  5. Project structure (tree + notes).
  6. Conventions & standards (including the canonical coding-style disclaimer).
  7. Using this repository as a template.
  8. Getting started.
  9. Running, scripts, and workflows.
  10. Usage examples.
  11. Testing & quality.
  12. CI/CD (describe only real workflows and config).
  13. Limitations and roadmap.
  14. License (only if a license file exists).

The agent must keep the coding-style disclaimer exactly as specified in the documentation:

> The coding style in this repository reflects a combination of widely used practices, prior experience, and personal preference, and is influenced by the Go projects on github.com/andygeiss. There is no single “best” project setup; you are encouraged to adapt this structure, evolve your own style, and use this repository as a starting point for your own projects.

***

## 4. Architectural and coding rules to preserve

The agent must ensure all documentation reinforces these invariants:

- Hexagonal architecture with domain at the center, and adapters depending on domain only.
- Domain logic in `internal/domain`, adapters in `internal/adapters`, entrypoints in `cmd/`.
- Ports (inbound/outbound) always defined in the domain layer.
- Explicit dependency injection in `cmd/*/main.go`; no global state or `init` wiring.
- All operations accept `context.Context` as the first parameter.
- Domain code never imports adapter packages.
- Testing follows Arrange–Act–Assert and the naming conventions from `CONTEXT.md`.

When new patterns or workflows are added to the codebase, the agent should first update `CONTEXT.md` (if needed) and then align `README.md` to it.

***

## 5. Agent workflow for README updates

When asked to modify or regenerate the README:

1. Re-scan:
   - Root and key directories: `cmd/`, `internal/`, `.justfile`, `go.mod`.
   - `CONTEXT.md`, `VENDOR.md`, and any new docs.
2. Build a short “Project at a Glance” internal view:
   - What the project does (current example plus any new features).
   - How it is structured (layers, key directories).
   - How it is built and run (`just` commands, Go versions, PGO).
3. Update sections selectively instead of rewriting everything, unless explicitly requested.
4. Never:
   - Add badges for services that are not configured.
   - Describe APIs, CLIs, or workflows that do not exist.
   - Change architectural rules defined in `CONTEXT.md` without clear repository changes.

For large refactors or new bounded contexts, the agent should add or update “Project structure”, “Using this repository as a template”, and “Limitations and roadmap”.

***

## 6. Badge configuration contract

Badges at the top of `README.md` must follow these rules:

- Go Reference:
  - `https://pkg.go.dev/badge/<module>.svg` → links to `https://pkg.go.dev/<module>`.
- Go Report Card:
  - `https://goreportcard.com/badge/<module>` → links to `https://goreportcard.com/report/<module>`.
- License:
  - `https://img.shields.io/github/license/<org>/<repo>.svg` → links to the `LICENSE` file.
- Release:
  - `https://img.shields.io/github/v/release/<org>/<repo>.svg` → links to the GitHub releases page.
- Test coverage:
  - Only describe coverage based on actual test tooling (for example, `go test` with coverage flags or an external service).

If the module path or GitHub org/repo changes, the agent must update all badge URLs and targets consistently.

***

## 7. Interaction with other docs

- `CONTEXT.md` is the primary contract for AI agents and RAG systems and must not be contradicted.
- `README.md` is human-first and should remain concise while exposing the architectural decisions.
- `VENDOR.md` documents `cloud-native-utils` and should be referenced but not duplicated.

When there is any ambiguity, the agent should:

- Prefer explicitly documented behavior in `CONTEXT.md`.
- Call out limitations or assumptions in the README rather than guessing.

***

## 8. Extension and maintenance

When new capabilities are added (for example, HTTP adapters, additional bounded contexts, CI workflows), the agent should:

- Extend the “Architecture overview”, “Project structure”, and “Using this repository as a template” sections to show new patterns.
- Update “Testing & quality” and “CI/CD” to reflect real, configured tools and pipelines only.
- Keep `AGENT.md` itself in sync with any changes to the canonical documentation flow.
