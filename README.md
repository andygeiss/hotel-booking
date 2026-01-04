<p align="center">
  <img src="cmd/server/assets/static/img/icon-192.png" alt="Go DDD Hexagonal Starter logo" width="96" height="96">
</p>

# Go DDD Hexagonal Starter

[![Go Reference](https://pkg.go.dev/badge/github.com/andygeiss/go-ddd-hex-starter.svg)](https://pkg.go.dev/github.com/andygeiss/go-ddd-hex-starter)
[![Go Report Card](https://goreportcard.com/badge/github.com/andygeiss/go-ddd-hex-starter)](https://goreportcard.com/report/github.com/andygeiss/go-ddd-hex-starter)
[![Release](https://img.shields.io/github/v/release/andygeiss/go-ddd-hex-starter.svg)](https://github.com/andygeiss/go-ddd-hex-starter/releases)

**A production-ready Go template demonstrating Domain-Driven Design (DDD) and Hexagonal Architecture (Ports and Adapters).**

## Overview

**go-ddd-hex-starter** provides a reusable blueprint for building maintainable Go applications with well-defined boundaries. It serves as a reference implementation of the Ports and Adapters pattern, designed to be easily extended by developers and AI coding agents alike.

The project includes working examples of:
- A **CLI tool** for indexing files.
- An **HTTP server** featuring OIDC authentication, templating, and session management.

## Key Features

- **Hexagonal Architecture:** Clear separation between Domain, Adapters, and Application layers.
- **Domain-Driven Design (DDD):** Pure business logic in the core, free of infrastructure dependencies.
- **Dependency Injection:** Explicit wiring in `main.go`, avoiding global state.
- **Standard Library First:** Built on `net/http` and standard Go patterns.
- **OIDC Authentication:** Secure login integration using `cloud-native-utils/security`.
- **Structured Logging:** JSON structured logs with `log/slog`.
- **Task Automation:** Integrated `just` task runner for build and test workflows.

## Architecture Overview

The project implements **Hexagonal Architecture** with three distinct layers:

```
┌──────────────────────────────────────────────────────────────┐
│                    Application Layer                         │
│                  cmd/cli/main.go                             │
│                  cmd/server/main.go                          │
│        (Entry points, wire adapters to domain)               │
└─────────────────────────┬────────────────────────────────────┘
                          │
┌─────────────────────────▼────────────────────────────────────┐
│                       Domain Layer                           │
│                   internal/domain/                           │
│    (Pure business logic, defines Ports as interfaces)        │
└─────────────────────────▲────────────────────────────────────┘
                          │
┌─────────────────────────┴────────────────────────────────────┐
│  ┌────────────────┐           ┌──────────────────┐           │
│  │ Inbound        │───────────│ Outbound         │           │
│  │ Adapters       │           │ Adapters         │           │
│  │ (Driving)      │           │ (Driven)         │           │
│  └────────────────┘           └──────────────────┘           │
│                     Adapters Layer                           │
│                 internal/adapters/                           │
└──────────────────────────────────────────────────────────────┘
```

### Layer Responsibilities

- **Domain** (`internal/domain/`): Pure business logic. Defines Ports (interfaces). Zero infrastructure knowledge.
- **Adapters** (`internal/adapters/`): Infrastructure implementations. Inbound (driving) and Outbound (driven).
- **Application** (`cmd/`): Entry points that wire adapters to domain services via dependency injection.

## Project Structure

```
go-ddd-hex-starter/
├── .justfile                 # Task runner commands
├── go.mod                    # Go module definition
├── CONTEXT.md                # Architecture and conventions documentation
├── VENDOR.md                 # Vendor library documentation
├── cmd/                      # Application entry points
│   ├── cli/                  # CLI application
│   └── server/               # HTTP server application
├── internal/
│   ├── adapters/             # Infrastructure implementations
│   │   ├── inbound/          # Driving adapters (HTTP, CLI)
│   │   └── outbound/         # Driven adapters (Persistence, Events)
│   └── domain/               # Core business logic
│       ├── event/            # Domain events
│       └── indexing/         # Bounded context: Indexing
└── tools/                    # Utility scripts
```

## Conventions & Standards

> The coding style in this repository reflects a combination of widely used practices, prior experience, and personal preference, and is influenced by the Go projects on github.com/andygeiss. There is no single “best” project setup; you are encouraged to adapt this structure, evolve your own style, and use this repository as a starting point for your own projects.

- **Dependency Rule:** Source code dependencies point **inward** only. Domain depends on nothing.
- **Context:** All operations accept `context.Context` as the first parameter.
- **Testing:** Follows Arrange–Act–Assert.

## Using this repository as a template

1.  **Clone or Use Template:** Start by cloning this repository.
2.  **Rename Module:** Update `go.mod` with your module path.
3.  **Define Domain:** Replace `internal/domain/indexing` with your own bounded contexts.
4.  **Implement Adapters:** Create new adapters in `internal/adapters` to satisfy your domain ports.
5.  **Wire Application:** Update `cmd/` to inject your new adapters.

## Getting Started

### Prerequisites

- **Go 1.25+**
- **Just** (Task runner)
- **Docker** or **Podman** (for container builds)

### Installation

```bash
git clone https://github.com/andygeiss/go-ddd-hex-starter.git
cd go-ddd-hex-starter
```

## Running, Scripts, and Workflows

This project uses `just` to manage workflows.

- **Build Container:**
  ```bash
  just build
  ```
- **Run Development Stack:**
  ```bash
  just up
  ```
- **Stop Development Stack:**
  ```bash
  just down
  ```
- **Run Tests:**
  ```bash
  just test
  ```
- **Run Profiling:**
  ```bash
  just profile
  ```

## Usage Examples

### CLI

The CLI tool indexes files in a directory.

```bash
go run cmd/cli/main.go -path /path/to/index
```

### Server

The HTTP server provides a web interface.

```bash
go run cmd/server/main.go
```
Access the server at `http://localhost:8080` (default).

## Testing & Quality

Tests are located alongside the code they test.

- Run all tests: `just test`
- Generate coverage: `go test ./... -cover`

## Limitations and Roadmap

- **CI/CD:** Currently, no automated CI/CD pipelines are configured in `.github/workflows`.
- **Persistence:** The default implementation uses a simple file-based JSON repository. For production, replace this with a database adapter (e.g., SQL, NoSQL).
