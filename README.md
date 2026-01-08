<p align="center">
<img src="https://github.com/andygeiss/go-ddd-hex-starter/blob/main/cmd/server/assets/static/img/icon-192.png?raw=true" width="100"/>
</p>

# Go DDD Hexagonal Starter

[![Go Reference](https://pkg.go.dev/badge/github.com/andygeiss/go-ddd-hex-starter.svg)](https://pkg.go.dev/github.com/andygeiss/go-ddd-hex-starter)
[![License](https://img.shields.io/github/license/andygeiss/go-ddd-hex-starter)](https://github.com/andygeiss/go-ddd-hex-starter/blob/master/LICENSE)
[![Releases](https://img.shields.io/github/v/release/andygeiss/go-ddd-hex-starter)](https://github.com/andygeiss/go-ddd-hex-starter/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/andygeiss/go-ddd-hex-starter)](https://goreportcard.com/report/github.com/andygeiss/go-ddd-hex-starter)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/f9f01632dff14c448dbd4688abbd04e8)](https://app.codacy.com/gh/andygeiss/go-ddd-hex-starter/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade)
[![Codacy Badge](https://app.codacy.com/project/badge/Coverage/f9f01632dff14c448dbd4688abbd04e8)](https://app.codacy.com/gh/andygeiss/go-ddd-hex-starter/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_coverage)

A production-ready Go starter template demonstrating Domain-Driven Design (DDD) and Hexagonal Architecture (Ports & Adapters) patterns.

<p align="center">
<img src="https://github.com/andygeiss/go-ddd-hex-starter/blob/main/cmd/server/assets/static/img/login.png?raw=true" width="300"/>
</p>

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [Usage](#usage)
- [Testing](#testing)
- [Configuration](#configuration)
- [Using as a Template](#using-as-a-template)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

This repository provides a reference implementation for structuring Go applications with clean architecture principles. It demonstrates how to:

- Organize code using **Hexagonal Architecture** (Ports & Adapters)
- Apply **Domain-Driven Design** tactical patterns (aggregates, entities, value objects, domain events)
- Integrate authentication via **OIDC/Keycloak**
- Implement event-driven communication with **Apache Kafka**

The template includes an `indexing` bounded context, an HTTP server with OIDC authentication, and a CLI demonstrating event-driven file indexing.

---

## Key Features

- **Developer Experience** — `just` task runner, golangci-lint, comprehensive test coverage
- **Domain-Driven Design** — Aggregates, entities, value objects, services, and domain events
- **Event Streaming** — Kafka-based pub/sub for domain events
- **File Indexing & Search** — Index workspace files and search by filename with relevance scoring
- **Hexagonal Architecture** — Clear separation between domain logic and infrastructure
- **OIDC Authentication** — Keycloak integration with session management
- **Production-Ready Docker** — Multi-stage build with PGO optimization (~5-10MB images)
- **Progressive Web App** — Service worker, manifest, and offline support for installable web apps

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Entry Points (cmd/)                      │
│                   cli/main.go, server/main.go               │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                  Inbound Adapters                           │
│   HTTP handlers, file readers, event subscribers            │
│              internal/adapters/inbound/                     │
└──────────────────────────┬──────────────────────────────────┘
                           │ implements ports
┌──────────────────────────▼──────────────────────────────────┐
│                     Domain Layer                            │
│   Bounded contexts: indexing/, event/                       │
│   Aggregates, entities, value objects, services, ports      │
│                   internal/domain/                          │
└──────────────────────────┬──────────────────────────────────┘
                           │ defines ports
┌──────────────────────────▼──────────────────────────────────┐
│                  Outbound Adapters                          │
│   Event publisher, repositories                             │
│              internal/adapters/outbound/                    │
└─────────────────────────────────────────────────────────────┘
```

### Bounded Contexts

| Context | Purpose |
|---------|---------|
| `event` | Domain event contracts and infrastructure |
| `indexing` | File indexing, search, and repository management |

For detailed architectural documentation, see [CONTEXT.md](CONTEXT.md).

---

## Project Structure

```
go-ddd-hex-starter/
├── cmd/                          # Application entry points
│   ├── cli/                      # CLI application (indexing demo)
│   └── server/                   # HTTP server (OIDC-protected UI)
├── internal/
│   ├── adapters/
│   │   ├── inbound/              # HTTP handlers, file readers, subscribers
│   │   └── outbound/             # Repositories, publishers
│   └── domain/
│       ├── event/                # Shared event infrastructure
│       └── indexing/             # Indexing bounded context
├── tools/                        # Build tooling (Python scripts)
├── .justfile                     # Task runner commands
├── docker-compose.yml            # Dev stack (Keycloak, Kafka, app)
└── Dockerfile                    # Multi-stage production build
```

---

## Getting Started

### Prerequisites

- **Docker** and **Docker Compose** (or Podman)
- **Go 1.25+**
- **golangci-lint** (for linting/formatting)
- **just** task runner

### Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/andygeiss/go-ddd-hex-starter.git
   cd go-ddd-hex-starter
   ```

2. **Install development tools:**
   ```bash
   just setup
   ```
   This installs `docker-compose`, `golangci-lint`, `just`, and `podman` via Homebrew.

3. **Configure environment:**
   ```bash
   cp .env.example .env
   cp .keycloak.json.example .keycloak.json
   ```

4. **Start the development stack:**
   ```bash
   just up
   ```
   This generates secrets, builds the Docker image, and starts Keycloak, Kafka, and the application.

5. **Access the application:**
   - **App:** http://localhost:8080/ui
   - **Keycloak Admin:** http://localhost:8180/admin (admin:admin)

---

## Usage

### Commands

| Command | Description |
|---------|-------------|
| `just build` | Build Docker image |
| `just down` | Stop all services |
| `just fmt` | Format code |
| `just lint` | Run linter |
| `just profile` | Generate CPU profile for PGO |
| `just run` | Run CLI application locally |
| `just serve` | Run HTTP server locally |
| `just setup` | Install development dependencies |
| `just test` | Run unit tests with coverage |
| `just test-integration` | Run integration tests |
| `just up` | Start full development stack |

### Running Locally (without Docker)

To run the server locally (requires Kafka running on localhost:9092):

```bash
# Ensure KAFKA_BROKERS is set to localhost:9092 in .env
just serve
```

### Running the CLI

The CLI demonstrates event-driven file indexing with Kafka.

**Basic usage:**
```bash
just run
```

**Example output:**
```
❯ main: creating index for path: /path/to/project
❯ main: waiting for event processing to complete...
❯ event: received EventFileIndexCreated - IndexID: /path/to/project, FileCount: 42
❯ main: event processing completed
❯ main: index created at 2026-01-07T10:30:00Z with 42 files
❯ main: index hash: abc123...
❯ main: first 5 files in index:
  - /path/to/project/main.go (1234 bytes)
  - /path/to/project/go.mod (567 bytes)
  ... and 37 more files
```

The CLI:
1. Indexes the current directory
2. Publishes an `EventFileIndexCreated` event to Kafka
3. Receives the event via subscription
4. Displays a summary of the indexed files

---

## Testing

### Unit Tests

Run all unit tests with coverage:

```bash
just test
```

This runs both Go and Python tests, generating `coverage.pprof`.

### Integration Tests

Integration tests require external services (Kafka, Keycloak):

```bash
just test-integration
```

### Test Organization

- Unit tests are colocated with source files (`*_test.go`)
- Integration tests are tagged with `//go:build integration`
- Test fixtures live in `testdata/` directories

---

## Configuration

Configuration is managed via environment variables. Copy `.env.example` to `.env` and customize:

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_NAME` | Display name for UI and PWA | `Template` |
| `APP_SHORTNAME` | Docker image/container name | `template` |
| `APP_VERSION` | Version for PWA cache busting | `1.0.0` |
| `KAFKA_BROKERS` | Kafka broker addresses | `localhost:9092` |
| `OIDC_CLIENT_ID` | OIDC client ID | `template` |
| `OIDC_CLIENT_SECRET` | OIDC client secret | Auto-generated |
| `OIDC_ISSUER` | Keycloak realm URL | `http://localhost:8180/realms/local` |
| `PORT` | HTTP server port | `8080` |

See `.env.example` for the complete list with documentation.

---

## Using as a Template

### Quick Start

1. **Clone and reinitialize:**
   ```bash
   git clone https://github.com/andygeiss/go-ddd-hex-starter my-project
   cd my-project
   rm -rf .git && git init
   ```

2. **Update module path:**
   ```bash
   go mod edit -module github.com/yourorg/my-project
   # Update import paths in all .go files
   ```

3. **Configure project identity:**
   ```bash
   cp .env.example .env
   # Edit APP_NAME, APP_SHORTNAME, APP_DESCRIPTION, APP_VERSION
   ```

4. **Add your domain logic:**
   - Create bounded contexts in `internal/domain/`
   - Implement adapters in `internal/adapters/`
   - Wire up entry points in `cmd/`

### What to Keep

- Directory structure (`cmd/`, `internal/adapters/`, `internal/domain/`)
- Hexagonal architecture pattern
- `cloud-native-utils` as infrastructure library
- `context.Context` threading through all layers

### What to Customize

- Bounded contexts and domain logic
- Static assets and templates in `cmd/*/assets/`
- Environment configuration in `.env`
- Docker Compose services as needed

For detailed conventions and rules, see [CONTEXT.md](CONTEXT.md).

---

## Contributing

1. Follow the coding conventions documented in [CONTEXT.md](CONTEXT.md)
2. Ensure all tests pass: `just test`
3. Ensure code is formatted and linted: `just fmt && just lint`
4. Update documentation if architecture changes

---

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
