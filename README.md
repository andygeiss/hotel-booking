# Go DDD Hexagonal Starter

[![Go Reference](https://pkg.go.dev/badge/github.com/andygeiss/go-ddd-hex-starter.svg)](https://pkg.go.dev/github.com/andygeiss/go-ddd-hex-starter)
[![License](https://img.shields.io/github/license/andygeiss/go-ddd-hex-starter)](https://github.com/andygeiss/go-ddd-hex-starter/blob/master/LICENSE)
[![Releases](https://img.shields.io/github/v/release/andygeiss/go-ddd-hex-starter)](https://github.com/andygeiss/go-ddd-hex-starter/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/andygeiss/go-ddd-hex-starter)](https://goreportcard.com/report/github.com/andygeiss/go-ddd-hex-starter)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/f9f01632dff14c448dbd4688abbd04e8)](https://app.codacy.com/gh/andygeiss/go-ddd-hex-starter/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade)
[![Codacy Badge](https://app.codacy.com/project/badge/Coverage/f9f01632dff14c448dbd4688abbd04e8)](https://app.codacy.com/gh/andygeiss/go-ddd-hex-starter/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_coverage)

A production-ready Go starter template demonstrating Domain-Driven Design (DDD) and Hexagonal Architecture (Ports & Adapters) patterns.

---

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Prerequisites](#prerequisites)
- [Getting Started](#getting-started)
- [Usage](#usage)
- [Testing](#testing)
- [Configuration](#configuration)
- [License](#license)

---

## Overview

This template provides a clean foundation for building cloud-native Go applications with proper separation of concerns. It includes a working example domain (`indexing`) that demonstrates:

- File indexing with domain events
- OIDC authentication via Keycloak
- Event streaming with Apache Kafka
- Server-side rendering with Go templates and HTMX

Use this as a starting point for your own projects or as a reference implementation for DDD/Hexagonal patterns in Go.

---

## Features

- **Hexagonal Architecture** — Clear separation between domain logic, inbound adapters (HTTP, events), and outbound adapters (persistence, messaging)
- **Domain-Driven Design** — Aggregates, entities, value objects, domain events, and application services
- **Event-Driven Architecture** — Kafka-based event publishing and subscribing
- **OIDC Authentication** — Keycloak integration with session management
- **Embedded Assets** — Static files and templates compiled into the binary
- **Production Container** — Multi-stage Dockerfile producing a minimal scratch-based image (~5-10MB)
- **Profile-Guided Optimization** — PGO support for optimized builds
- **Developer Experience** — `just` command runner for common tasks

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                   cmd/ (Entry Points)                       │
│           server/main.go       cli/main.go                  │
└─────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────┐
│           internal/adapters/inbound/ (Driving)              │
│      HTTP handlers, event subscribers, file readers         │
└─────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────┐
│                  internal/domain/                           │
│     Pure business logic: aggregates, entities, services     │
└─────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────┐
│           internal/adapters/outbound/ (Driven)              │
│       Event publisher (Kafka), file-based repository        │
└─────────────────────────────────────────────────────────────┘
```

For detailed architectural documentation, see [CONTEXT.md](CONTEXT.md).

---

## Project Structure

```
go-ddd-hex-starter/
├── cmd/
│   ├── cli/                  # CLI demo application
│   └── server/               # HTTP server with OIDC auth
│       └── assets/           # Embedded templates and static files
├── internal/
│   ├── adapters/
│   │   ├── inbound/          # HTTP handlers, event subscribers
│   │   └── outbound/         # Repositories, event publishers
│   └── domain/
│       ├── event/            # Event interfaces
│       └── indexing/         # Example bounded context
├── tools/                    # Development utilities
├── .justfile                 # Command runner recipes
├── docker-compose.yml        # Local development stack
└── Dockerfile                # Production container build
```

---

## Prerequisites

- **Go 1.25+**
- **Docker** or **Podman** (for container builds)
- **Docker Compose** (for local development stack)
- **just** (command runner) — install via `brew install just`

---

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/andygeiss/go-ddd-hex-starter.git
cd go-ddd-hex-starter
```

### 2. Install Dependencies

```bash
just setup
```

This installs `docker-compose`, `golangci-lint`, `just` and `podman` via Homebrew.

### 3. Configure Environment

```bash
cp .env.example .env
cp .keycloak.json.example .keycloak.json
```

### 4. Start Services

```bash
just up
```

This will:
1. Generate a random OIDC client secret
2. Build the Docker image
3. Start Keycloak, Kafka, and the application

### 5. Access the Application

| Service | URL |
|---------|-----|
| Application | http://localhost:8080 |
| Keycloak Admin | http://localhost:8180/admin |

Default Keycloak credentials: `admin` / `admin`

---

## Usage

### Commands

| Command | Alias | Description |
|---------|-------|-------------|
| `just build` | `just b` | Build Docker image |
| `just up` | `just u` | Start all services |
| `just down` | `just d` | Stop all services |
| `just fmt` | — | Format Go code with golangci-lint |
| `just lint` | — | Run golangci-lint checks |
| `just test` | `just t` | Run tests with coverage |
| `just serve` | — | Run HTTP server locally |
| `just run` | — | Run CLI demo locally |
| `just profile` | — | Generate PGO profiles |

### Running Locally (without Docker)

To run the server locally (requires Kafka and Keycloak running separately):

```bash
just serve
```

To run the CLI demo:

```bash
just run
```

---

## Testing

Run all tests with coverage:

```bash
just test
```

Or use Go directly:

```bash
go test -v ./internal/...
```

Tests follow the naming convention:
```
Test_<Struct>_<Method>_With_<Condition>_Should_<Result>
```

---

## Configuration

All configuration is via environment variables. See [.env.example](.env.example) for the complete list with documentation.

### Key Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port | `8080` |
| `KAFKA_BROKERS` | Kafka broker addresses | `localhost:9092` |
| `OIDC_ISSUER` | Keycloak realm URL | `http://localhost:8180/realms/local` |
| `OIDC_CLIENT_ID` | OIDC client identifier | `template` |
| `OIDC_CLIENT_SECRET` | OIDC client secret | (generated) |

---

## Using as a Template

1. Update `go.mod` with your module path
2. Update all imports to match
3. Configure `.env.example` with your application metadata
4. Replace or extend `internal/domain/indexing/` with your domains
5. Implement adapters for your infrastructure

See [CONTEXT.md](CONTEXT.md) for detailed conventions and guidelines.

---

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
