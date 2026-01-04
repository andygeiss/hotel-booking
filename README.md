<p align="center">
  <img src="cmd/server/assets/static/img/login.png" alt="Go DDD Hexagonal Starter logo" width="300">
</p>

# Go DDD Hexagonal Starter

[![Go Reference](https://pkg.go.dev/badge/go-ddd-hex-starter.svg)](https://pkg.go.dev/go-ddd-hex-starter)
[![Go Report Card](https://goreportcard.com/badge/go-ddd-hex-starter)](https://goreportcard.com/report/go-ddd-hex-starter)
[![License](https://img.shields.io/github/license/andygeiss/go-ddd-hex-starter.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/andygeiss/go-ddd-hex-starter.svg)](https://github.com/andygeiss/go-ddd-hex-starter/releases)
[![Test coverage](https://img.shields.io/badge/test%20coverage-go%20test%20-coverprofile)](#testing--quality)

Production-ready Go template demonstrating Domain-Driven Design (DDD) and Hexagonal Architecture (Ports and Adapters).

## Overview / Motivation

This repository is a starter template intended to be reused for real services:

- Clear boundaries between **domain**, **adapters**, and **entrypoints**.
- Dependency direction enforced (everything depends inward).
- Working examples for both a **CLI** (file indexing) and an **HTTP server** (OIDC login + templating).

## Key Features

- Hexagonal architecture with ports defined in the domain layer
- Pure domain logic in `internal/domain/` (no infrastructure imports)
- Inbound and outbound adapters in `internal/adapters/`
- Explicit dependency injection in `cmd/*/main.go`
- HTTP server based on `net/http` + `cloud-native-utils/security`
- OIDC authentication (Keycloak-friendly dev setup via Docker Compose)
- Embedded templates and static assets via `embed.FS`
- Task runner workflows via `just` (build, run, serve, test, profile, compose up/down)
- PGO support (CPU profiling artifacts used by the Docker build)

## Architecture Overview

This project follows Hexagonal Architecture (Ports and Adapters):

- **Domain layer**: pure business logic and ports (interfaces)
- **Adapters layer**: infrastructure implementations
  - inbound (driving): HTTP handlers, filesystem readers
  - outbound (driven): repositories, event publishers
- **Application layer**: entrypoints wiring everything together via dependency injection

Key invariants (from [CONTEXT.md](CONTEXT.md)):

- Ports (inbound/outbound) are defined in `internal/domain/**/ports_*.go`
- Domain code never imports adapter packages
- No global wiring: dependency injection happens in `cmd/*/main.go`
- All operations accept `context.Context` as the first parameter

## Project Structure

```
.
├── .justfile
├── CONTEXT.md
├── VENDOR.md
├── Dockerfile
├── docker-compose.yml
├── tools/
├── cmd/
│   ├── cli/
│   │   ├── main.go
│   │   └── assets/
│   └── server/
│       ├── main.go
│       └── assets/
│           ├── static/
│           └── templates/
└── internal/
    ├── adapters/
    │   ├── inbound/
    │   └── outbound/
    └── domain/
        ├── event/
        └── indexing/
```

Notes:

- `cmd/cli`: CLI example that indexes files in the current directory and writes an `index.json`.
- `cmd/server`: HTTP server example that serves UI templates and uses OIDC authentication.
- `internal/domain/indexing`: bounded context implementing the indexing domain.
- `internal/adapters/*`: adapter implementations (HTTP, filesystem, JSON file persistence, messaging).
- `tools/`: helper scripts used by `just` tasks.

## Conventions & Standards

- Architectural contracts and naming conventions live in [CONTEXT.md](CONTEXT.md) and are authoritative.
- Vendor library guidance (especially cross-cutting concerns) lives in [VENDOR.md](VENDOR.md).

Coding-style disclaimer (must remain unchanged):

> The coding style in this repository reflects a combination of widely used practices, prior experience, and personal preference, and is influenced by the Go projects on github.com/andygeiss. There is no single “best” project setup; you are encouraged to adapt this structure, evolve your own style, and use this repository as a starting point for your own projects.

## Using This Repository as a Template

- Click **Use this template** on GitHub to create a new repository from this one.
- Update the module name in [go.mod](go.mod) to your desired module path.
- Search/replace import paths in `cmd/` and `internal/` to match the new module path.
- Update application metadata in your `.env` (see below).

## Getting Started

### Prerequisites

- Go 1.25+ (see [go.mod](go.mod))
- `just` (task runner)
- `docker-compose`
- A container build tool for `just build` (this repo uses `podman build` by default)
- Python 3 (for `tools/*.py` tasks)

Optional (for `just profile` SVG output):

- Graphviz (`dot`) installed

### Local Configuration

This repository expects local-only config files:

- Copy `.env.example` to `.env`
- Copy `.keycloak.json.example` to `.keycloak.json`

`just up` will replace the placeholder `CHANGE_ME_LOCAL_SECRET` in both files with a generated secret.

## Running, Scripts, and Workflows

All workflows are defined in [.justfile](.justfile):

- `just setup`: installs `just` and `docker-compose` via Homebrew
- `just serve`: runs the HTTP server locally (`go run ./cmd/server/main.go`)
- `just run`: runs the CLI locally (`go run ./cmd/cli/main.go`)
- `just test`: runs unit tests under `./internal/...` and writes `coverage.pprof`
- `just profile`: generates `cpuprofile.pprof` and `cpuprofile.svg` via `tools/create_pgo.py`
- `just build`: builds the container image with `podman build`
- `just up`: builds the image and starts Keycloak + Kafka + app via Docker Compose
- `just down`: stops Docker Compose services

Container notes:

- `just build` uses **Podman** by default (`podman build`). If you prefer Docker, adjust the build recipe accordingly.
- `docker-compose.yml` runs Keycloak (OIDC) and Kafka (messaging) as optional dev dependencies.

## Usage Examples

### CLI: Index files

Run:

```bash
just run
```

Behavior (from [cmd/cli/main.go](cmd/cli/main.go)):

- Reads file infos from the current directory (`.`)
- Builds an index keyed by your working directory
- Writes a JSON file at `./index.json` (then removes it at process exit)

### Server: UI + OIDC Login

Run locally (without Docker Compose):

```bash
just serve
```

Important endpoints (from [internal/adapters/inbound/router.go](internal/adapters/inbound/router.go)):

- `GET /liveness`
- `GET /readiness`
- `GET /ui/` (requires authentication; redirects to `/ui/login` if unauthenticated)
- `GET /ui/login` (renders login page)

Auth callback:

- The OIDC redirect URI is configured via `OIDC_REDIRECT_URL` (see `.env.example`).
- The callback handling is provided by `cloud-native-utils/security` through the server mux initialization.

## Testing & Quality

Run tests and collect coverage:

```bash
just test
```

This will:

- run `go test -v -coverprofile=coverage.pprof ./internal/...`
- print the total coverage summary

Testing conventions (Arrange–Act–Assert, naming) are defined in [CONTEXT.md](CONTEXT.md).

## CI/CD

No GitHub Actions workflows are included in this repository at the moment.

If you add CI, keep it aligned with the existing workflows:

- `go test ./...` (or `./internal/...`)
- optional coverage reporting based on `coverage.pprof`
- optional PGO profiling via `just profile`

## Limitations and Roadmap

- The module name in [go.mod](go.mod) is currently `go-ddd-hex-starter`; template users typically change this to their own module path.
- Docker build uses `-pgo cpuprofile.pprof` (see [Dockerfile](Dockerfile)); if the file is missing, image build will fail until you generate it via `just profile`.
- No CI workflows are present yet (see the CI/CD section).
