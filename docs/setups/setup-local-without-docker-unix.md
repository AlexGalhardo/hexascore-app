# Local setup without Docker on Linux and macOS

This guide runs HexaScore on Linux or macOS with a native PostgreSQL 18 install and no Docker. It is enough to develop, run the unit and integration tests and lint. The Docker-based checks are not available; see [What does not work without Docker](#what-does-not-work-without-docker). The Docker guide is [setup-local-with-docker-unix.md](setup-local-with-docker-unix.md).

## Prerequisites

Install Go 1.27.3, bun 1.4.3, the vukem CLI v0.36.0 and git as described in the [Docker guide's prerequisites](setup-local-with-docker-unix.md#prerequisites). Docker is not needed.

## Install PostgreSQL 18.6

The repository pins Postgres 18.6 (`postgres:18.6-alpine3.24` in `docker-compose.yml`, `postgresql18-client=18.6-r0` in the `Dockerfile`). Use the same version natively.

### Debian and Ubuntu (apt.postgresql.org)

Add the PGDG repository as documented on <https://www.postgresql.org/download/linux/ubuntu/>, then install the exact 18.6 package. The version string depends on your release; list it with `apt-cache policy` (on Ubuntu 24.04 it is `18.6-1.pgdg24.04+2`):

```sh
sudo apt install -y postgresql-common
sudo /usr/share/postgresql-common/pgdg/apt.postgresql.org.sh
apt-cache policy postgresql-18
sudo apt install -y postgresql-18=18.6-1.pgdg24.04+2 postgresql-client-18=18.6-1.pgdg24.04+2
```

The cluster listens on 5432. Either set `port = 5433` in `/etc/postgresql/18/main/postgresql.conf` and run `sudo systemctl restart postgresql`, or change the port in `DATABASE_URL` and `TEST_DATABASE_URL` later. Run `psql` as the `postgres` system user: `sudo -u postgres psql`.

### macOS (Homebrew)

```sh
brew install postgresql@18
brew services start postgresql@18
```

Homebrew installs the current `postgresql@18` formula (18.6 at the time of writing; check with `brew info postgresql@18`) and cannot pin the patch version. The formula is keg-only: add `$(brew --prefix postgresql@18)/bin` to `PATH`. The superuser is your macOS user, and the port is 5432.

Either way you also get `pg_dump` and `pg_restore`; the backup tests use them and are skipped when they are missing.

## Create the role and databases

`.env.example` expects this:

| Variable | Value |
| --- | --- |
| `DATABASE_URL` | `postgres://hexascore:hexascore@127.0.0.1:5433/hexascore?sslmode=disable` |
| `TEST_DATABASE_URL` | `postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable` |

So create a login role `hexascore` with password `hexascore` that owns a database `hexascore`, and that may connect to the `postgres` database. The integration tests connect to `postgres` and create and drop their own databases, so the role also needs `CREATEDB`. On Linux prefix `psql` with `sudo -u postgres`; on macOS run it as yourself:

```sh
psql -p 5433 -d postgres -c "CREATE ROLE hexascore LOGIN PASSWORD 'hexascore' CREATEDB;"
psql -p 5433 -d postgres -c "CREATE DATABASE hexascore OWNER hexascore;"
```

Use `-p 5432` if the server still listens there. The URLs connect over TCP to `127.0.0.1`, which the default `pg_hba.conf` of both packages accepts with a password. These credentials are for local development only. If you use another port, user or password, change `DATABASE_URL` and `TEST_DATABASE_URL` in `.env` to match.

## Object storage

The app does not need a bucket for local development. With `S3_BUCKET` empty (the default):

- covers come from the embedded demo SVGs;
- the daily ID exports go to a local directory (`ID_EXPORTS_DIR`, default `storage/id-exports`);
- cover ingestion and `backup:run` report that no storage disk is configured.

If you need real object storage, run any S3-compatible server natively (RustFS, <https://docs.rustfs.com/>, the version pinned in `docker-compose.yml` is 1.0.4; or MinIO), create a bucket and fill the `S3_*` variables in `.env`. See the [vukem storage docs](https://vukem.dev/docs/digging-deeper/storage).

## Setup steps

`setups/setup-unix.sh` requires Docker, so run its steps by hand (the same commands it runs, minus `docker compose`):

```sh
cp -n .env.example .env
grep -q '^APP_KEY=.' .env || vukem key:generate
git config core.hooksPath .githooks
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0
bun install --frozen-lockfile
```

Adjust `.env` now if your database URLs differ. Never commit `.env`.

## Migrate, seed and serve

```sh
vukem migrate
vukem db:seed DemoCatalog
vukem serve
```

Open <http://localhost:3000>. The demo admin signs in as `admin@example.test` with password `Demo-password-1`.

## Tests and lint

Everything in the [Docker guide's table](setup-local-with-docker-unix.md#tests-and-lint) works except `scripts/ci/local.sh`: `go test ./...`, `go test -tags integration ./...` (against your native Postgres through `TEST_DATABASE_URL`), `golangci-lint run --build-tags integration ./...`, `bun run lint`, `bun run typecheck`, `bun run test`, the mobile checks and markdownlint.

## What does not work without Docker

| Check | Why |
| --- | --- |
| The full app container (AGENTS.md "Local QA", `docker compose --profile app`) | builds and runs the production image |
| `bun run e2e` (Playwright) | runs against the built container (`scripts/ci/qa-up.sh`) |
| Schemathesis (`qa/schemathesis/ci.sh`) | runs in a container against the built image |
| Lighthouse in the `lhci-runner:local` container | Linux container used for the release gate |
| `scripts/ci/local.sh` | needs Docker and the Compose Postgres on 5433 |
| Docker image build, `scripts/infra/check-image.sh`, `scripts/openapi-check.sh --image` | need Docker |

CI runs all of them on every push.

## Troubleshooting

- `password authentication failed for user "hexascore"`: the role or password differs from `.env`; fix either side.
- `connection refused` on 5433: the server listens on another port or is not running (`systemctl status postgresql`, `brew services list`).
- Integration tests fail with `permission denied to create database`: give the role `CREATEDB` (`ALTER ROLE hexascore CREATEDB;`).
- Low memory: run golangci-lint, the integration tests and the Vite build one at a time, and limit the linter with `-j 2 --allow-parallel-runners`.
