# Development Guide

Local development setup, build, test, and common workflows for the Hotel Booking backend.

---

## Prerequisites

| Tool | Why | Install (macOS) |
|------|-----|-----------------|
| Go 1.25+ | Language toolchain (Docker image uses `golang:1.26rc1-alpine3.23`, but `go.mod` declares 1.25.5) | `brew install go` |
| `just` | Task runner for every script below | `brew install just` |
| `golangci-lint` v2+ | Lint + format (see `.golangci.yml`) | `brew install golangci-lint` |
| `docker-compose` | Orchestrates the dev stack | `brew install docker-compose` |
| `podman` | Default image builder used by `just build` | `brew install podman` |
| `graphviz` | For `just profile` PGO SVG output | `brew install graphviz` |

One-shot: `just setup` runs the Homebrew installs listed above.

## First-Time Setup

1. **Clone and enter**
   ```bash
   git clone https://github.com/andygeiss/hotel-booking.git
   cd hotel-booking
   ```

2. **Install tooling** (Homebrew-based)
   ```bash
   just setup
   ```

3. **Provision local config files**
   ```bash
   cp .env.example .env
   cp .keycloak.json.example .keycloak.json
   ```
   Both files contain the placeholder `CHANGE_ME_LOCAL_SECRET`. On first `just up`, the justfile rotates the placeholder in-place to a random 32-char secret so the app and Keycloak realm agree.

4. **Start the stack**
   ```bash
   just up
   ```
   Brings up `app`, `keycloak`, `kafka`, `postgres-reservation`, `postgres-payment`.

5. **Verify it's alive**
   - App: http://localhost:8080/ui
   - Keycloak admin: http://localhost:8180/admin (admin / admin)

## Daily Commands (`.justfile`)

| Command | What it does |
|---------|--------------|
| `just up` / `just u` | `podman build` the image, rotate the local secret if present, then `docker-compose up -d` and poll Keycloak until its import directory is ready |
| `just down` / `just d` | `docker-compose --env-file .env down` |
| `just build` / `just b` | `podman build -t ${USER}/${APP_SHORTNAME}:latest -f Dockerfile .` |
| `just fmt` | `golangci-lint fmt ./...` (formats in place) |
| `just lint` | `golangci-lint run ./...` |
| `just test` / `just t` | `go test -v -coverprofile=.coverage.pprof ./internal/...` and print total coverage |
| `just test-integration [path]` | `go test -tags=integration -v <path>` (default `./internal/...`) |
| `just profile` | Run benchmarks with `-cpuprofile=.cpuprofile.pprof` and generate `.cpuprofile.svg` (for the PGO build) |
| `just setup` | Install the tools listed above via Homebrew |

## Running the Server Locally (without the stack)

When you want to iterate on Go code without the Docker wrapper (still need Postgres/Kafka/Keycloak running):

```bash
just up                    # start infra only is not directly supported — see note
docker-compose --env-file .env up -d postgres-reservation postgres-payment kafka keycloak
go run ./cmd/server
```

`go run` uses the `.env` file only if you source it into the shell (`set -a; source .env; set +a`) or run through a tool like `direnv`. Otherwise the defaults inside `main.go` kick in.

## Single-Test Invocation

```bash
go test -v -run Test_Reservation_Confirm_From_Pending_Should_Change_Status \
    ./internal/domain/reservation/...
```

For table-driven tests, the inner name is usually passed with a slash: `-run 'Parent/SubTest'`.

## Test Conventions

- Pattern: `Test_{Component}_{Scenario}_Should_{ExpectedResult}` (e.g., `Test_Route_MCP_Endpoint_Without_MCPServer_Should_Return_404`).
- Unit tests live next to the code (`*_test.go`).
- Integration tests use the build tag `//go:build integration` (skipped in `just test`, run via `just test-integration`).
- Handler tests use `httptest.NewRecorder()` and the testdata templates under `internal/adapters/inbound/testdata/assets/templates/`.
- Use `t.Setenv("APP_NAME", ...)` and `t.Setenv("APP_DESCRIPTION", ...)` — a number of handlers read these at closure construction time.
- For MCP route tests, pass `Verifier: nil` in `RouterConfig` to skip bearer auth (covered by `main_test.go:Benchmark_Server_Integration_MCP_Tools_List_Should_Be_Fast`).

## Domain Iteration Loop

1. Edit an aggregate or service under `internal/domain/...`.
2. Update or add a `*_test.go` beside it.
3. `just test` to verify — aim to keep domain tests fast and infrastructure-free.
4. If you changed an exported port or event shape, update:
   - the adapter(s) under `internal/adapters/...`
   - `cmd/server/main.go` DI wiring (if a new dependency is required)
   - `docs/api-contracts.md` / `docs/data-models.md` if behavior is externally visible

## Linter Notes (`.golangci.yml`)

`default: all` with a deliberate list of disables: `depguard, exhaustruct, paralleltest, wsl, varnamelen, ireturn, noctx, tagliatelle, mnd, nlreturn, wrapcheck, noinlineerr, forbidigo, forcetypeassert, err113, lll, wsl_v5`. `gofmt` is configured with rewrite rules that replace `interface{}` → `any` and `a[b:len(a)]` → `a[b:]` when you run `just fmt`.

`revive` disables `package-comments`, `exported`, `var-naming`, `unused-parameter`. `gosec` excludes `G301` (dir perms). If a rule fires in CI but not locally, make sure you are on golangci-lint v2+.

## Debugging Tips

- **Port conflicts:** the stack uses `8080` (app), `8180` (keycloak), `9092` (kafka), `5432` (reservation DB), `5433` (payment DB). If any of those are in use, edit `docker-compose.yml` mappings.
- **OIDC issuer mismatch:** the compose file uses `extra_hosts: ["localhost:host-gateway"]` so that `http://localhost:8180/realms/local` resolves both from the browser and from inside the `app` container. Do not change `OIDC_ISSUER` without understanding this.
- **MCP 401s:** fetch a fresh bearer via client credentials (see `docs/api-contracts.md#mcp-authentication-flow`).
- **State machine errors** (e.g. "cannot confirm from cancelled"): the aggregate enforces transitions. Check the state machine in `docs/data-models.md` before reaching for the service layer.
- **Same-day checkout/check-in**: allowed (a new reservation with CheckIn == previous CheckOut does not overlap, per `Reservation.IsOverlapping`).

## PGO Profiling Loop

```bash
just profile                               # produces .cpuprofile.pprof and .cpuprofile.svg
just build                                 # Dockerfile builds with `-pgo=.cpuprofile.pprof`
```

The Dockerfile requires `.cpuprofile.pprof` to be present at build time. If it is missing, either run `just profile` first or remove the `-pgo` flag from the `go build` invocation in `Dockerfile`.

## Documentation Touch Points (before commit)

From `CLAUDE.md`:

- New HTTP handler → update **Quick Reference** in CLAUDE.md
- New MCP tool → update **MCP Tools** in CLAUDE.md
- New domain error → update **Domain Errors** in CLAUDE.md
- New state transition → update **State Machines** in CLAUDE.md
- New env var → update **Environment Variables** in CLAUDE.md and README.md

The BMAD-generated docs (this folder) are regenerated by running the `bmad-document-project` workflow again.
