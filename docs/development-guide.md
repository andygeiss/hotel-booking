# Development Guide

Local development setup, build, test, and common workflows for the Hotel Booking backend.

---

## Prerequisites

| Tool | Why | Install (macOS) |
|------|-----|-----------------|
| Go 1.27.1 | The pinned toolchain; `go.mod` declares `go 1.27` and the Docker builder uses `golang:1.27.1-alpine3.23` | `brew install go` |
| `make` | The only command surface | already installed |
| `docker-compose` | Orchestrates the dev stack | `brew install docker-compose` |
| `podman` | Builds the image | `brew install podman` |
| `graphviz` | For `make profile` PGO SVG output | `brew install graphviz` |

```bash
brew install go docker-compose podman graphviz
```

`make check` needs no installed linter: it runs staticcheck and govulncheck through
`go run ...@latest`, so it needs the network instead.

If `make ci` prints a `go version` line that is not `go1.27.1`, put `GOTOOLCHAIN=go1.27.1`
in front of every `go` and `make` command. Never add a `toolchain` line to `go.mod`.

## First-Time Setup

1. **Clone and enter**
   ```bash
   git clone https://github.com/andygeiss/hotel-booking.git
   cd hotel-booking
   ```

2. **Install tooling** (Homebrew-based)
   ```bash
   brew install go docker-compose podman graphviz
   ```

3. **Provision local config files**
   ```bash
   cp .env.example .env
   cp .keycloak.json.example .keycloak.json
   ```
   Both files contain the placeholder `CHANGE_ME_LOCAL_SECRET`. Rotate it once, so the
   app and the Keycloak realm agree on the same random secret:
   ```bash
   sed -i '' "s/CHANGE_ME_LOCAL_SECRET/$(LC_ALL=C tr -dc 'A-Za-z0-9' < /dev/urandom | head -c 32)/g" .env .keycloak.json
   ```
   Running it again does nothing once the placeholder is gone.

4. **Start the stack**
   ```bash
   podman build -t "$USER/hotel-booking:latest" -f Dockerfile .
   docker-compose --env-file .env up -d
   ```
   Brings up `app`, `keycloak`, `kafka`, `postgres-reservation`, `postgres-payment`.
   Keycloak needs about a minute to import its realm on a cold start.

5. **Verify it's alive**
   - App: http://localhost:8080/ui
   - Keycloak admin: http://localhost:8180/admin (admin / admin)

## Daily Commands (`Makefile`)

`make` is the whole command surface, copied from the baseline's `stack/makefile.md`.
There is no CI server, so the two commands that matter are yours to run:
`make check` before every commit, `make ci` before every push.

| Command | What it does |
|---------|--------------|
| `make build` | `CGO_ENABLED=0 go build -trimpath -o bin/ ./cmd/server` |
| `make check` | The eight gates against the working tree — the default target |
| `make ci` | The same gates against `git archive HEAD`, in a clean copy |
| `make clean` | `rm -rf bin/` |
| `make fmt` | `goimports -w .` then `go fix ./...` — read the diff before committing it |
| `make profile` | Rewrites `cmd/server/default.pgo` from the benchmarks, plus `.cpuprofile.svg` |
| `make run` | `go run ./cmd/server`, loading `.env` if it is there |
| `make test` | `go test -race -shuffle=on ./...` |

**The eight gates, in order:** format, vet, fix, staticcheck, govulncheck, tidy, test,
static build. A red `govulncheck` means bump the dependency — never silence the check.

The container stack has no Make target; the baseline bans Docker targets. Use
`docker-compose` directly:

| Command | What it does |
|---------|--------------|
| `podman build -t "$USER/hotel-booking:latest" -f Dockerfile .` | Build the image |
| `docker-compose --env-file .env up -d` | Start the stack |
| `docker-compose --env-file .env down` | Stop it (volumes survive) |
| `docker-compose logs -f app` | Follow the app log |

## Running the Server Locally (without the stack)

When you want to iterate on Go code without the Docker wrapper (still need Postgres/Kafka/Keycloak running):

```bash
docker-compose --env-file .env up -d postgres-reservation postgres-payment kafka keycloak
make run
```

`make run` sources `.env` for you and starts the server with `go run`. It is the only
target that reads that file: `check` and `test` must never depend on a developer's
machine, or `make ci` would go red on a commit that is fine. Without `.env`, the
defaults inside `main.go` kick in.

## Configuration

`./server -h` is the whole contract. Every knob sits in one `Config` struct in
`cmd/server/config.go`, parsed and validated in `main` before a database opens or a
listener binds; nothing under `internal/` reads the environment. Flags beat environment
variables beat built-in defaults, and each flag's help text names its variable.

```bash
make run                          # loads .env, then starts
go run ./cmd/server -port 9090    # a flag beats PORT from .env
go run ./cmd/server -h            # the flags, plus the variables that are not flags
```

A bad setting is one line and exit 2, before anything starts:

```
$ go run ./cmd/server -port=http
server: port "http": want a number from 0 to 65535
```

Handler tests no longer need `t.Setenv("APP_NAME", ...)`: the identity arrives as an
`inbound.AppInfo` parameter, and `testApp()` in `router_test.go` supplies it.

## Single-Test Invocation

```bash
go test -v -run Test_Reservation_Confirm_From_Pending_Should_Change_Status \
    ./internal/domain/reservation/...
```

For table-driven tests, the inner name is usually passed with a slash: `-run 'Parent/SubTest'`.

## Test Conventions

- Pattern: `Test_{Component}_{Scenario}_Should_{ExpectedResult}` (e.g., `Test_Route_MCP_Endpoint_Without_MCPServer_Should_Return_404`).
- Unit tests live next to the code (`*_test.go`).
- Integration tests use the build tag `//go:build integration` (skipped by `make test`, run via `go test -tags=integration -v ./internal/...`).
- Handler tests use `httptest.NewRecorder()` and the testdata templates under `internal/adapters/inbound/testdata/assets/templates/`.
- Handlers take an `inbound.AppInfo`, so no test sets `APP_NAME` any more; use `testApp()` from `router_test.go`.
- For MCP route tests, pass `Verifier: nil` in `RouterConfig` to skip bearer auth (covered by `main_test.go:Benchmark_Server_Integration_MCP_Tools_List_Should_Be_Fast`).

## Domain Iteration Loop

1. Edit an aggregate or service under `internal/domain/...`.
2. Update or add a `*_test.go` beside it.
3. `make test` to verify — aim to keep domain tests fast and infrastructure-free.
4. If you changed an exported port or event shape, update:
   - the adapter(s) under `internal/adapters/...`
   - `cmd/server/main.go` DI wiring (if a new dependency is required)
   - `docs/api-contracts.md` / `docs/data-models.md` if behavior is externally visible

## Linter Notes

There is no `golangci-lint` and no `.golangci.yml` any more. The baseline allows exactly
one third-party lint tool, `staticcheck`, and `make check` runs it through
`go run honnef.co/go/tools/cmd/staticcheck@latest ./...`.

`@latest` is deliberate for staticcheck and govulncheck: both are analysis gates, never
build inputs, so a new check can redden a run but can never change the shipped binary.
If a staticcheck release breaks the gate on an unrelated morning, pin that one line to
the previous version in the commit that says why, and remove the pin once the findings
are fixed.

Formatting is `gofmt` through `goimports`; `make fmt` applies it, and `make check` fails
while a rewrite is pending.

## Debugging Tips

- **Port conflicts:** the stack uses `8080` (app), `8180` (keycloak), `9092` (kafka), `5432` (reservation DB), `5433` (payment DB). If any of those are in use, edit `docker-compose.yml` mappings.
- **OIDC issuer mismatch:** the compose file uses `extra_hosts: ["localhost:host-gateway"]` so that `http://localhost:8180/realms/local` resolves both from the browser and from inside the `app` container. Do not change `OIDC_ISSUER` without understanding this.
- **MCP 401s:** fetch a fresh bearer via client credentials (see `docs/api-contracts.md#mcp-authentication-flow`).
- **State machine errors** (e.g. "cannot confirm from cancelled"): the aggregate enforces transitions. Check the state machine in `docs/data-models.md` before reaching for the service layer.
- **Same-day checkout/check-in**: allowed (a new reservation with CheckIn == previous CheckOut does not overlap, per `Reservation.IsOverlapping`).

## PGO Profiling Loop

```bash
make profile   # rewrites cmd/server/default.pgo, and .cpuprofile.svg to look at
```

`cmd/server/default.pgo` is committed, and `go build` finds it on its own because it sits
next to the main package — no `-pgo` flag anywhere, and a clean checkout builds without
running the benchmarks first. `go install` of a tagged version gets the optimization too.

The profile currently comes from the benchmarks in `cmd/server/main_test.go`. The baseline
would rather it came from a 30-second CPU profile of real production traffic
(`/debug/pprof/profile?seconds=30` on the ops listener); refresh it that way once there is
traffic worth profiling.

## Documentation Touch Points (before commit)

From `CLAUDE.md`:

- New HTTP handler → update **Quick Reference** in CLAUDE.md
- New MCP tool → update **MCP Tools** in CLAUDE.md
- New domain error → update **Domain Errors** in CLAUDE.md
- New state transition → update **State Machines** in CLAUDE.md
- New env var → update **Environment Variables** in CLAUDE.md and README.md

The BMAD-generated docs (this folder) are regenerated by running the `bmad-document-project` workflow again.
