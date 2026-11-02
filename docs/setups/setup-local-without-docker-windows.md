# Local setup without Docker on Windows 11

This guide runs HexaScore on Windows 11 with a native PostgreSQL install and no Docker. It is enough to develop, run the unit and integration tests and lint. The Docker-based checks are not available; see [What does not work without Docker](#what-does-not-work-without-docker). The Docker guide is [setup-local-with-docker-windows.md](setup-local-with-docker-windows.md).

## Prerequisites

Install Go 1.27.3, bun 1.4.3, the vukem CLI v0.36.0 and git as described in the [Docker guide's prerequisites](setup-local-with-docker-windows.md#prerequisites), and enable long paths ([Troubleshooting](setup-local-with-docker-windows.md#paths-longer-than-260-characters)). Docker Desktop is not needed.

## Install PostgreSQL 18.6

The repository pins Postgres 18.6 (`postgres:18.6-alpine3.24` in `docker-compose.yml`, `postgresql18-client=18.6-r0` in the `Dockerfile`). Use the same version natively. Pick one of:

- **Official installer** (EnterpriseDB, linked from <https://www.postgresql.org/download/windows/>): download the 18.6 installer from <https://www.enterprisedb.com/downloads/postgres-postgresql-downloads>. When it asks for the port, enter `5433` so the URLs in `.env.example` work unchanged. It installs a Windows service and asks for a password for the `postgres` superuser.
- **Scoop** (main bucket, package `postgresql`, version `18.6-5`):

  ```powershell
  scoop install postgresql@18.6-5
  ```

  Follow the notes Scoop prints: start the server with `pg_ctl start`, or register a service with `pg_ctl register`. The superuser is `postgres` with a blank password, and the port is 5432 unless you change `port` in `postgresql.conf` (in Scoop's persisted data directory) to `5433`.

Either way the install puts `psql`, `pg_dump` and `pg_restore` on `PATH`; the backup tests use the last two and are skipped when they are missing.

## Create the role and databases

`.env.example` expects this:

| Variable | Value |
| --- | --- |
| `DATABASE_URL` | `postgres://hexascore:hexascore@127.0.0.1:5433/hexascore?sslmode=disable` |
| `TEST_DATABASE_URL` | `postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable` |

So create a login role `hexascore` with password `hexascore` that owns a database `hexascore`, and that may connect to the `postgres` database. The integration tests connect to `postgres` and create and drop their own databases, so the role also needs `CREATEDB`:

```powershell
psql -h 127.0.0.1 -p 5433 -U postgres -c "CREATE ROLE hexascore LOGIN PASSWORD 'hexascore' CREATEDB;"
psql -h 127.0.0.1 -p 5433 -U postgres -c "CREATE DATABASE hexascore OWNER hexascore;"
```

Use `-p 5432` if your server listens there. These credentials are for local development only. If you use another port, user or password, change `DATABASE_URL` and `TEST_DATABASE_URL` in `.env` to match.

## Object storage

The app does not need a bucket for local development. With `S3_BUCKET` empty (the default):

- covers come from the embedded demo SVGs;
- the daily ID exports go to a local directory (`ID_EXPORTS_DIR`, default `storage/id-exports`);
- cover ingestion and `backup:run` report that no storage disk is configured.

If you need real object storage, run any S3-compatible server natively (RustFS, <https://docs.rustfs.com/>, the version pinned in `docker-compose.yml` is 1.0.4; or MinIO), create a bucket and fill the `S3_*` variables in `.env`. See the [vukem storage docs](https://vukem.dev/docs/digging-deeper/storage).

## Run the setup script without services

`setups/setup-windows.ps1 -NoServices` does everything except the Docker stack and the migrations: it checks `go`, `bun` and `vukem`, creates `.env` from `.env.example`, generates `APP_KEY`, sets `core.hooksPath .githooks`, installs golangci-lint 2.15.0 and runs `bun install --frozen-lockfile`.

```powershell
powershell -ExecutionPolicy Bypass -File setups\setup-windows.ps1 -NoServices
```

Adjust `.env` now if your database URLs differ. Never commit `.env`.

## Migrate, seed and serve

```powershell
vukem migrate
vukem db:seed DemoCatalog
vukem serve
```

Open <http://localhost:3000>. The demo admin signs in as `admin@example.test` with password `Demo-password-1`.

## Tests and lint

Everything in the [Docker guide's table](setup-local-with-docker-windows.md#tests-and-lint) works except `scripts/ci/local.sh`: `go test ./...`, `go test -tags integration ./...` (against your native Postgres through `TEST_DATABASE_URL`), `golangci-lint run --build-tags integration ./...`, `bun run lint`, `bun run typecheck`, `bun run test`, the mobile checks and markdownlint.

## What does not work without Docker

| Check | Why |
| --- | --- |
| The full app container (AGENTS.md "Local QA", `docker compose --profile app`) | builds and runs the production image |
| `bun run e2e` (Playwright) | runs against the built container (`scripts/ci/qa-up.sh`) |
| Schemathesis (`qa/schemathesis/ci.sh`) | runs in a container against the built image |
| Lighthouse in the `lhci-runner:local` container | Linux container; host Lighthouse numbers are invalid with Kaspersky |
| `scripts/ci/local.sh` | needs Docker and the Compose Postgres on 5433 |
| Docker image build, `scripts/infra/check-image.sh`, `scripts/openapi-check.sh --image` | need Docker |

CI runs all of them on every push.

## Troubleshooting

The Windows notes in the Docker guide also apply: [260-character paths](setup-local-with-docker-windows.md#paths-longer-than-260-characters), [`node_modules` junctions in worktrees](setup-local-with-docker-windows.md#node_modules-junctions-in-git-worktrees) and [low memory](setup-local-with-docker-windows.md#low-memory).

- `password authentication failed for user "hexascore"`: the role or password differs from `.env`; fix either side.
- `connection refused` on 5433: the server listens on another port, or the service is not running (`pg_ctl status`, or the PostgreSQL service in `services.msc`).
- Integration tests fail with `permission denied to create database`: give the role `CREATEDB` (`ALTER ROLE hexascore CREATEDB;`).
