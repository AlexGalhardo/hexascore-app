# HexaScore

[hexascore.app](https://hexascore.app): community scores for games, movies, TV series and books, six categories per rating, a published formula, a read-only public API and an Android app. Go on the [vukem](https://vukem.dev) framework, Inertia + React with SSR, PostgreSQL, deployed on Railway.

This repository is the rewrite of the Next.js app kept on branch `legacy`. `v1.0.0` is parity with that app plus the cutover of `hexascore.app`; the plan is in [ROADMAP.md](ROADMAP.md).

## Run it locally

```sh
setups/setup-windows.ps1   # or setups/setup-unix.sh: tools, .env, APP_KEY, compose stack, bun install, migrate
vukem serve                # http://localhost:3000 with Vite HMR
```

Full container with the demo catalogue (what the QA suites and the owner test):

```sh
E2E_IMAGE_TAG=local docker compose --profile app build app-init
E2E_IMAGE_TAG=local docker compose --profile app up -d --wait   # http://127.0.0.1:3000
```

Demo admin: `admin@example.test` / `Demo-password-1` (seeder `DemoCatalog`).

## Checks

```sh
go build ./... && golangci-lint run --build-tags integration ./...
bun run lint && bun run typecheck && bun run test
TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable go test -tags integration ./...
bun run e2e                 # Playwright + axe against the container
```

## Where to read next

| Document | What |
| --- | --- |
| [AGENTS.md](AGENTS.md) | rules, commands and layout for every coding agent (humans too) |
| [ROADMAP.md](ROADMAP.md) | phases, microtasks and acceptance criteria up to `v1.0.0` and beyond |
| [PROMPT-WORKERS.md](PROMPT-WORKERS.md) | system prompt of the agent team that executes the roadmap |
| [docs/sdd/](docs/sdd/README.md) | one design document per domain |
| [docs/adrs/](docs/adrs/README.md) | architecture decisions |
| [docs/PROGRESS.md](docs/PROGRESS.md) | programme log, newest at the bottom |
| [CHANGELOG.md](CHANGELOG.md) | releases (git-cliff) |
| [vukem-feedback.md](vukem-feedback.md) | framework bugs and requests found while building HexaScore |

## Project standards

| File | What it does |
| --- | --- |
| `.editorconfig` | tabs of width 4, spaces where the format needs them (YAML, Markdown, JavaScript) |
| `.gitattributes` | LF in the repository and in every working copy, Windows included |
| `.githooks/commit-msg` | Conventional Commits check (`git config core.hooksPath .githooks`, done by the setup scripts) |
