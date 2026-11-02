# Local setup with Docker on Windows 11

This guide runs HexaScore on Windows 11 with Docker Desktop: Postgres and RustFS in containers, the Go server and Vite on the host. For Linux and macOS see [setup-local-with-docker-unix.md](setup-local-with-docker-unix.md); without Docker see [setup-local-without-docker-windows.md](setup-local-without-docker-windows.md).

## Prerequisites

| Tool | Version | Where the version comes from | Install |
| --- | --- | --- | --- |
| Go | 1.27.3 | `go` directive in `go.mod` | <https://go.dev/dl/> |
| bun | 1.4.3 | `packageManager` in `package.json` | PowerShell: `iex "& {$(irm https://bun.sh/install.ps1)} -Version 1.4.3"` ([bun docs](https://bun.sh/docs/installation)) |
| vukem CLI | v0.36.0 | `github.com/AlexGalhardo/vukem` in `go.mod` | `go install github.com/AlexGalhardo/vukem/cmd/vukem@v0.36.0` |
| golangci-lint | 2.15.0 | `GOLANGCI_LINT_VERSION` in `.github/workflows/ci.yml` | installed by the setup script |
| git | not pinned by the repo | | Git for Windows, <https://git-scm.com/download/win> |
| Docker Desktop | not pinned by the repo (WSL 2 backend) | | <https://docs.docker.com/desktop/setup/install/windows-install/> |

The vukem CLI version must match the vukem module in `go.mod`; CI installs it the same way. Make sure `$(go env GOPATH)\bin` is on `PATH`, or `vukem` and `golangci-lint` will not be found.

Before cloning, enable long paths (see [Troubleshooting](#paths-longer-than-260-characters)).

## Clone and run the setup script

```powershell
git clone https://github.com/AlexGalhardo/hexascore-app.git
cd hexascore-app
powershell -ExecutionPolicy Bypass -File setups\setup-windows.ps1
```

`setups/setup-windows.ps1` is idempotent; run it again whenever something is missing. It:

1. checks that `go`, `bun`, `vukem` and `docker` are on `PATH`;
2. copies `.env.example` to `.env` if `.env` does not exist;
3. runs `vukem key:generate` when `APP_KEY` is empty;
4. sets `git config core.hooksPath .githooks` (the commit-msg hook checks Conventional Commits);
5. installs golangci-lint 2.15.0 with `go install` when the installed version differs;
6. runs `docker compose up -d --wait`;
7. runs `bun install --frozen-lockfile`;
8. runs `vukem migrate`.

`-NoServices` skips steps 6 and 8 (the CI Windows job uses it because its runners have no Linux containers).

## The Docker Compose stack

`docker compose up -d --wait` starts two services from `docker-compose.yml` and waits until both are healthy:

| Service | Image | Host address | Credentials (dev only) |
| --- | --- | --- | --- |
| `postgres` | `postgres:18.6-alpine3.24` | `127.0.0.1:5433` | user `hexascore`, password `hexascore`, database `hexascore` |
| `rustfs` | `rustfs/rustfs:1.0.4` | S3 API `127.0.0.1:9000`, console `127.0.0.1:9001` | `hexascore` / `hexascore-local-secret` |

Postgres is published on 5433 because 5432 is often taken by another container. Both ports bind to `127.0.0.1` only. The local Postgres runs with `fsync=off`: it is for development and tests, never for real data.

## The `.env` file

The setup script creates `.env` from `.env.example`, which documents every variable. The defaults already point at the Compose stack:

- `DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/hexascore?sslmode=disable`
- `TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable`
- `APP_URL=http://localhost:3000`, `PORT=3000`
- `MAIL_MAILER=log`: mail is written to the log instead of being sent.

Never commit `.env`. Optional integrations (analytics, Turnstile, Telegram, IndexNow, Stripe) stay off while their variables are empty.

### Covers and object storage (optional)

With `S3_BUCKET` empty the app serves the embedded demo cover SVGs and writes the daily ID exports to a local directory (`ID_EXPORTS_DIR`, default `storage/id-exports`). Jobs that need a bucket (cover ingestion, backups) report that no storage disk is configured.

To use RustFS instead, open the console at <http://127.0.0.1:9001>, sign in with the credentials above, create a bucket, and fill `S3_BUCKET`, `S3_ACCESS_KEY_ID=hexascore` and `S3_SECRET_ACCESS_KEY=hexascore-local-secret` in `.env` (the other `S3_*` defaults already match). See the [vukem storage docs](https://vukem.dev/docs/digging-deeper/storage).

## Migrate and seed

The setup script already ran the migrations. To run them again, or after pulling new migrations:

```powershell
vukem migrate
```

Seed the demo catalogue (1 admin, 50 users, 300 items in four types, 2,000 ratings; idempotent):

```powershell
vukem db:seed DemoCatalog
```

The demo admin signs in as `admin@example.test` with password `Demo-password-1`.

## Run the app

```powershell
vukem serve
```

This starts the Go server with Vite HMR. Open <http://localhost:3000>. Use `localhost`, which matches `APP_URL`.

## Tests and lint

| Command | What it checks |
| --- | --- |
| `go test ./...` | Go unit tests |
| `go test -tags integration ./...` | Go integration tests on Postgres (`TEST_DATABASE_URL`) |
| `golangci-lint run --build-tags integration ./...` | Go lint and formatting |
| `bun run lint`, `bun run typecheck`, `bun run test` | web lint (Biome), types, Vitest |
| `bun run --cwd apps/mobile lint`, `typecheck`, `test` | Android app |
| `bunx markdownlint-cli2@0.24.0 "**/*.md"` | Markdown |

`scripts/ci/local.sh` runs every CI job except e2e locally (it needs Git Bash, Docker and the Compose Postgres on 5433). On a low-memory machine, read [Low memory](#low-memory) first.

## Optional: the full app in a container

This is the "Local QA" recipe of [AGENTS.md](../../AGENTS.md#agents): it builds the production image from the current tree, migrates, seeds DemoCatalog and serves it.

```powershell
$env:E2E_IMAGE_TAG = "local"
docker compose --profile app build app-init
docker compose --profile app up -d --wait
```

Open <http://localhost:3000>. The image runs with `APP_ENV=production`, so it answers `301` to `APP_URL` (`http://localhost:3000`) for any other host: use `localhost`, not `127.0.0.1`. Stop `vukem serve` first, or set `QA_PORT` to another port, because both use 3000. Running several stacks in parallel (own project, port and database) is described in the same AGENTS.md section.

## Troubleshooting

### Paths longer than 260 characters

Windows limits paths to 260 characters by default, and bun's isolated `node_modules` layout produces longer ones. Enable long paths once (administrator PowerShell), then tell git to use them:

```powershell
New-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Control\FileSystem" -Name "LongPathsEnabled" -Value 1 -PropertyType DWORD -Force
git config --global core.longpaths true
```

Some tools ignore the setting. The Android SDK's ninja 1.10.2 fails with `manifest 'build.ninja' still dirty after 100 tries`; the fix (ninja 1.13 or later through a Gradle init script) is in the [mobile release runbook](../runbooks/mobile-release.md). Cloning into a short path such as `C:\dev\hexascore-app` also helps.

### Lighthouse numbers are wrong with Kaspersky

Kaspersky (and similar antivirus products that inspect traffic) changes what the browser loads, so Lighthouse runs on the Windows host give invalid numbers. Measure Lighthouse in a Linux container (`lhci-runner:local`) against `localhost`, as described in the release gate rules of [AGENTS.md](../../AGENTS.md#rules).

### Docker Desktop has no `--network host`

Commands written for Linux CI that use `--network host` (Schemathesis in `qa/schemathesis/ci.sh`, k6) do not work on Docker Desktop. Run the tool container in the app container's network namespace (`--network container:<app container>`), or target `host.docker.internal` and send the original `Host` header.

### `node_modules` junctions in git worktrees

If you share one `node_modules` between worktrees through a junction, never run `bun install` inside the worktree (it writes into the main checkout), and remove the junction before `git worktree remove`, because a recursive delete can follow it into the main checkout.

### Low memory

The Go linter, the Docker image build, the integration tests and the Vite build each use a lot of memory. On a machine with little RAM, run them one at a time, limit golangci-lint with `golangci-lint run -j 2 --allow-parallel-runners --build-tags integration ./...`, and avoid several QA stacks at once. Give Docker Desktop a memory limit in its settings so that WSL 2 does not take the whole host.
