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
- Connect to local LLMs via **LM Studio** for AI agent capabilities

The template includes two bounded contexts (`agent` and `indexing`), an HTTP server with OIDC authentication, and a CLI demonstrating the agent loop pattern.

---

## Key Features

- **Hexagonal Architecture** — Clear separation between domain logic and infrastructure
- **Domain-Driven Design** — Aggregates, entities, value objects, services, and domain events
- **OIDC Authentication** — Keycloak integration with session management
- **Event Streaming** — Kafka-based pub/sub for domain events
- **LLM Agent with Tool Execution** — Observe → decide → act → update pattern with tool calling (file search, extensible)
- **File Indexing & Search** — Index workspace files and search by filename with relevance scoring
- **Progressive Web App** — Service worker, manifest, and offline support for installable web apps
- **Production-Ready Docker** — Multi-stage build with PGO optimization (~5-10MB images)
- **Developer Experience** — `just` task runner, golangci-lint, comprehensive test coverage

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
│   Bounded contexts: agent/, indexing/, event/               │
│   Aggregates, entities, value objects, services, ports      │
│                   internal/domain/                          │
└──────────────────────────┬──────────────────────────────────┘
                           │ defines ports
┌──────────────────────────▼──────────────────────────────────┐
│                  Outbound Adapters                          │
│   Event publisher, repositories, LLM client                 │
│              internal/adapters/outbound/                    │
└─────────────────────────────────────────────────────────────┘
```

### Bounded Contexts

| Context | Purpose |
|---------|---------|
| `agent` | LLM-based agent with observe→decide→act→update loop, tool execution |
| `indexing` | File indexing, search, and repository management |
| `event` | Domain event contracts and infrastructure |

For detailed architectural documentation, see [CONTEXT.md](CONTEXT.md).

---

## Project Structure

```
go-ddd-hex-starter/
├── cmd/                          # Application entry points
│   ├── cli/                      # CLI application (agent demo)
│   └── server/                   # HTTP server (OIDC-protected UI)
├── internal/
│   ├── adapters/
│   │   ├── inbound/              # HTTP handlers, file readers, subscribers
│   │   └── outbound/             # Repositories, publishers, LLM client
│   └── domain/
│       ├── agent/                # Agent bounded context
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

- **Go 1.25+**
- **Docker** and **Docker Compose** (or Podman)
- **just** task runner
- **golangci-lint** (for linting/formatting)

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
| `just setup` | Install development dependencies |
| `just build` | Build Docker image |
| `just up` | Start full development stack |
| `just down` | Stop all services |
| `just serve` | Run HTTP server locally |
| `just run` | Run CLI application locally |
| `just test` | Run unit tests with coverage |
| `just test-integration` | Run integration tests |
| `just lint` | Run linter |
| `just fmt` | Format code |
| `just profile` | Generate CPU profile for PGO |

### Running Locally (without Docker)

To run the server locally (requires Kafka running on localhost:9092):

```bash
# Ensure KAFKA_BROKERS is set to localhost:9092 in .env
just serve
```

To run the CLI application (demonstrates agent with file search):

```bash
just run
```

The CLI agent can search indexed files using the `search_index` tool when processing queries.

---

## Testing

### Unit Tests

Run all unit tests with coverage:

```bash
just test
```

This runs both Go and Python tests, generating `coverage.pprof`.

### Integration Tests

Integration tests require external services (LM Studio, Kafka, Keycloak):

```bash
# Ensure LM_STUDIO_URL and LM_STUDIO_MODEL are set in .env
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
| `PORT` | HTTP server port | `8080` |
| `KAFKA_BROKERS` | Kafka broker addresses | `localhost:9092` |
| `OIDC_ISSUER` | Keycloak realm URL | `http://localhost:8180/realms/local` |
| `OIDC_CLIENT_ID` | OIDC client ID | `template` |
| `OIDC_CLIENT_SECRET` | OIDC client secret | Auto-generated |
| `LM_STUDIO_URL` | LM Studio API URL | `http://localhost:1234` |
| `LM_STUDIO_MODEL` | LLM model name | `qwen/qwen3-coder-30b` |

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
