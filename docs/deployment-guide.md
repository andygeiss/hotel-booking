# Deployment Guide

Container image build, docker-compose dev stack, the gates, and environment configuration.

---

## Dockerfile (Multi-Stage)

Two stages, defined in `Dockerfile`:

1. **Builder**: `golang:1.27.1-alpine3.23` (the pinned Go toolchain — never an RC)
   - `CGO_ENABLED=0`, `GO111MODULE=on`
   - `go mod download` as a separate layer (cached while `go.mod`/`go.sum` are unchanged)
   - `go build -ldflags "-s -w" -trimpath -o server ./cmd/server`
   - Profile-Guided Optimization needs no flag: `go build` reads `cmd/server/default.pgo` on its own, because the file sits next to the main package. It is committed, so a clean checkout builds. Refresh it with `make profile`.
2. **Runtime**: `FROM scratch`
   - Copies `/server` only. Static binary, embedded assets, no libc.
   - `EXPOSE 8080`, `ENTRYPOINT ["/server"]`.

Final image is ~5-10 MB. Build it with `podman build -t "$USER/hotel-booking:latest" -f Dockerfile .`; `docker-compose.yml` expects the tag `${USER}/${APP_SHORTNAME}:latest` (e.g. `andygeiss/hotel-booking:latest`).

## `.dockerignore`

Present at repo root. Notable exclusion: `.git` is ignored (enforced explicitly — see commit `2b4c1b2 fix: prevent .git copy in docker`).

## Local Dev Stack (`docker-compose.yml`)

Five services:

| Service | Image | Port(s) | Notes |
|---------|-------|---------|-------|
| `app` | `${USER}/${APP_SHORTNAME}:latest` | `8080 → 8080` | Reads `.env`; `depends_on` keycloak, kafka, postgres-reservation, postgres-payment; `extra_hosts: ["localhost:host-gateway"]` so the container can reach host `localhost:8180` |
| `keycloak` | `quay.io/keycloak/keycloak:latest` | `8180 → 8080` | `start-dev --import-realm`, mounts `.keycloak.json` at `/opt/keycloak/data/import/local-realm.json` |
| `kafka` | `apache/kafka:latest` | `9092 → 9092` | Auto-creates topics |
| `postgres-reservation` | `postgres:16-alpine` | `5432 → 5432` | Mounts `migrations/reservation/init.sql` as init script; healthcheck via `pg_isready` |
| `postgres-payment` | `postgres:16-alpine` | `5433 → 5432` | Mounts `migrations/payment/init.sql`; healthcheck via `pg_isready` |

Persistent volumes: `postgres_reservation_data`, `postgres_payment_data`. Both DBs bootstrap with the same `kv_store` schema.

Bring the stack up and down with `docker-compose` directly. Rotate the
`CHANGE_ME_LOCAL_SECRET` placeholder in `.env` and `.keycloak.json` once before the first
start (see [Development Guide](./development-guide.md)); Keycloak needs about a minute to
import its realm on a cold start.

```bash
docker-compose --env-file .env up -d      # start
docker-compose --env-file .env down       # stop; volumes survive
```

## The gates (there is no CI server)

Nothing runs the tests for you. The gates live in the `Makefile` and run on the
developer's machine:

- **`make check` before every commit** — the eight gates against the working tree:
  format, vet, fix, staticcheck, govulncheck, tidy, test, static build.
- **`make ci` before every push** — the same list against `git archive HEAD`, copied to
  an empty directory. A file you forgot to `git add`, a stray `.env`, or a local edit
  cannot make it green. Read its `go version` line: it is the only record of which
  toolchain ran.

A red `govulncheck` means bump the dependency, never silence the check. Dependency
updates are done by hand on a 90-day cycle, under the pin:

```bash
export GOTOOLCHAIN=go1.27.1
go get -u ./... && go mod tidy && make check     # one chore(deps) commit
```

There is deliberately nothing under `.github/workflows/` and no dependency bot. Building
a container image is not a gate either: it would need a runtime up on every `make check`.

The image build does **not** push anywhere — deployment is out-of-band.

## Configuration

**`./server -h` is the contract, not this table.** Every knob the binary reads sits in one
`Config` struct in `cmd/server/config.go`, parsed and validated before a database opens or
a listener binds. Nothing under `internal/` reads the environment.

**Flags beat environment variables beat built-in defaults**, and each flag names its
variable in its own help text. The variables that are not flags — the database settings,
`CREDENTIALS_DIRECTORY`, and the ones cloud-native-utils reads for itself — are printed
under the flag list, so `-h` stays complete.

A configuration error is one line on stderr and exit 2, before anything starts:

```
$ ./server -port=http
server: port "http": want a number from 0 to 65535
```

Validated at boot: both listener and database ports, a non-empty app name, an `http`/`https`
issuer URL, and — because neither half is wrong on its own — that the two bounded contexts
do not point at the same database.

### Secrets

The two database passwords are read from **files**, one per secret, in the directory named
by `CREDENTIALS_DIRECTORY`: `reservation-db-password` and `payment-db-password`. A file is
not inherited by every child process and does not appear in a process listing, which is
true of neither a flag value nor an environment variable.

When `CREDENTIALS_DIRECTORY` is unset, the passwords fall back to `RESERVATION_DB_PASSWORD`
and `PAYMENT_DB_PASSWORD`. That fallback is a development convenience and a recorded
deviation — see [README § Baseline deviations](../README.md#baseline-deviations). Compose
does not supply credential files yet.

`Config.LogValue` is an allowlist, so the boot log prints the safe fields and no password
can be added to the struct and logged by accident.

### The variables

See `.env.example` for the annotated list. Summary:

### Application identity

- `APP_NAME` (default `Hotel Booking`) — UI + logs display name
- `APP_DESCRIPTION` — used in page titles and PWA manifest
- `APP_SHORTNAME` (default `hotel-booking`) — Docker image + container name
- `APP_VERSION` (default: the build version from `debug.ReadBuildInfo`) — reported by the MCP server and by `/healthz`

### HTTP server

- `HOST` (default empty — every interface, which is what the container needs; a bare-metal deployment behind a proxy sets `127.0.0.1`)
- `PORT` (default `8080`)
- `REDIRECT_URL` (default `http://localhost:8080/ui`)

The four server timeouts are no longer configurable. They are constants in
`cmd/server/main.go` at the baseline's values — read header 5s, read 10s, write 30s,
idle 2m — because the old `SERVER_*_TIMEOUT` defaults of 5s put the write timeout below
what a slow response needs.

### OIDC / Keycloak

- `OIDC_ISSUER` (default `http://localhost:8180/realms/local`) — same URL must resolve from both the browser and the app container
- `OIDC_CLIENT_ID` (default `hotel-booking`) — session-based web client
- `OIDC_CLIENT_SECRET` — starts as `CHANGE_ME_LOCAL_SECRET`; rotate it in place once, before the first start (see [Development Guide](./development-guide.md))
- `OIDC_REDIRECT_URL` (default `http://localhost:8080/auth/callback`)
- `MCP_CLIENT_ID` (default `hotel-booking-mcp`) — client-credentials flow for `/mcp`

### Databases

- Reservation: `RESERVATION_DB_{HOST,PORT,USER,NAME,SSLMODE}`, plus the password above
- Payment: `PAYMENT_DB_{HOST,PORT,USER,NAME,SSLMODE}`, plus the password above

The two must not name the same database; the binary refuses to start if they do.

In the compose network, DB hosts are `postgres-reservation` and `postgres-payment`; from the host, use `localhost` with ports `5432`/`5433`.

### Kafka

- `KAFKA_BROKERS` (default `localhost:9092`) — use `kafka:9092` when running inside compose
- `KAFKA_CONSUMER_GROUP_ID` (default `test-group`)

### Resilience knobs (consumed by `cloud-native-utils`)

- `SERVICE_BREAKER_THRESHOLD=5`, `SERVICE_DEBOUNCE_PER_SEC=10`, `SERVICE_RETRY_DELAY=5s`, `SERVICE_RETRY_MAX=3`, `SERVICE_TIMEOUT=5s`.

## Production Notes

This repo ships only development infrastructure. Production adaptations to consider:

- Replace `MockPaymentGateway` with a real provider (its `PaymentGateway` port is the only seam).
- Replace `MockNotificationService` with an email/push provider.
- Replace `apache/kafka:latest` with your managed Kafka / Redpanda / MSK endpoint via `KAFKA_BROKERS`.
- Point Keycloak at a durable realm and rotate `OIDC_CLIENT_SECRET` / `MCP_CLIENT_ID` secrets via your secret store (not the placeholder-rotation script).
- Build + push the image from CI (not currently wired).
- Run migrations through your own orchestrator — the compose healthchecks are local-dev only. The init SQL is idempotent (`CREATE TABLE IF NOT EXISTS`), so it's safe to re-run.
- Put a TLS-terminating reverse proxy in front of port 8080.

## Rollback & Observability

- The binary is stateless (all mutable state is in the two Postgres instances + Kafka). Rolling back reduces to redeploying a prior image tag.
- Logs are structured JSON via `logging.NewJsonLogger()` (from `cloud-native-utils`).
- No metrics/tracing wired by default; consider adding via `cloud-native-utils` helpers or the standard `runtime/metrics` + OpenTelemetry stack.
