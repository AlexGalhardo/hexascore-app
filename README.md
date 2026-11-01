# HexaScore.app

[![CI](https://github.com/AlexGalhardo/hexascore-app/actions/workflows/ci.yml/badge.svg)](https://github.com/AlexGalhardo/hexascore-app/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go 1.27.3](https://img.shields.io/badge/go-1.27.3-00ADD8?logo=go&logoColor=white)](go.mod)

Movies, TV Shows, Books and Games Open Source Almanac

## Introduction

HexaScore is an open source almanac and community score platform for games, movies, TV shows and books. Each item is rated by the community in six categories per type, and the ratings combine into a published score. The application is built on the [vukem](https://vukem.dev) Go framework.

This repository is the Go rewrite. The previous Next.js application is kept on the `legacy` branch.

## Summary

- [Introduction](#introduction)
- [Features](#features)
- [Tech stack](#tech-stack)
- [Live site](#live-site)
- [Public API](#public-api)
- [Getting started](#getting-started)
- [Project structure](#project-structure)
- [Testing and CI](#testing-and-ci)
- [Mobile app](#mobile-app)
- [Stability](#stability)
- [Contributing](#contributing)
- [Code of conduct](#code-of-conduct)
- [Security vulnerabilities](#security-vulnerabilities)
- [License](#license)
- [Acknowledgements](#acknowledgements)

## Features

- **Community scores** for games, movies, TV shows and books: six category ratings per item, entered in half-star steps, with a published Bayesian aggregate ([ADR 0005](docs/adrs/0005-scoring-model.md)).
- **Moderated edits**: community edit suggestions with an audit trail ([ADR 0010](docs/adrs/0010-community-edit-suggestions.md)).
- **Search** on PostgreSQL full-text search with `pg_trgm` ([ADR 0011](docs/adrs/0011-search-postgres.md)).
- **Books** modelled as works and editions ([ADR 0012](docs/adrs/0012-books-work-and-editions.md)).
- **Lists and statuses** per user.
- **Server-rendered public pages** with React and Inertia ([ADR 0004](docs/adrs/0004-inertia-react-with-ssr.md)).
- **Paid read-only public API** with Stripe billing ([ADR 0015](docs/adrs/0015-monetisation-v1.md)).
- **Android app** in `apps/mobile`.
- **Accounts** with two-factor authentication, data export and an admin panel restricted by IP allowlist.
- **Data ingestion** from external sources, and a social layer (follows and activity feed), both behind feature flags that are off by default.

## Tech stack

| Layer | Technology |
| --- | --- |
| Backend | Go 1.27.3 on the vukem framework v0.36.0 |
| Web UI | Inertia 3.9.1 with React 19.3.1 and server-side rendering for public pages |
| Styling | Tailwind CSS 4.3.3 |
| Build | Vite 8.3.5 |
| Database | PostgreSQL 18.6 |
| Object storage | S3-compatible storage (RustFS locally) |
| Billing | Stripe (vukem `contrib/stripe`) |
| Mobile | Expo SDK 58 with NativeWind, Android |
| Hosting | Railway ([ADR 0002](docs/adrs/0002-railway-over-vercel.md)) |
| CI | GitHub Actions |

## Live site

The production site is <https://hexascore.app>, also served at <https://www.hexascore.app>.

## Public API

The paid, read-only JSON API is served at `https://api.hexascore.app/v1`. Interactive documentation is at <https://api.hexascore.app/docs>, and the OpenAPI document is in `openapi/api-v1.json`.

The API serves only fields with redistributable provenance: Wikidata, Wikipedia, Open Library and HexaScore's own data. Fields derived from TMDB and IGDB are not part of the API ([ADR 0006](docs/adrs/0006-third-party-data-licensing.md)), so its coverage is smaller than the site's.

SDKs generated from the same OpenAPI document:

- TypeScript: `sdks/ts`
- Python: `sdks/python` (package `hexascore`)

Regenerate them with `bun run sdk:generate` and `bun run sdk:generate:python`.

## Getting started

### Requirements

- Go 1.27.3 (see `go.mod`)
- Bun 1.4.3 (see `packageManager` in `package.json`)
- Docker with Docker Compose
- golangci-lint v2.15.0 (the version CI pins)
- Git
- The `vukem` CLI, used by the commands below
- uv, only for the Python SDK tests

### Setup

The setup script installs the tools, sets the Git hooks path (`core.hooksPath .githooks`), creates `.env` and `APP_KEY`, starts the Docker Compose stack, and runs `bun install` and `vukem migrate`.

On Linux and macOS:

```bash
setups/setup-unix.sh
```

On Windows (PowerShell):

```powershell
setups/setup-windows.ps1
```

Add `-NoServices` to skip Docker Compose and migrations: `setups/setup-windows.ps1 -NoServices`.

Start the development server:

```bash
vukem serve
```

To start only the local services:

```bash
docker compose up -d --wait
```

This starts PostgreSQL 18.6 on `127.0.0.1:5433` and RustFS on `127.0.0.1:9000` (console on `9001`).

### Environment variables

The setup script creates `.env` from `.env.example`. To configure it by hand, copy `.env.example` to `.env`. Use local values only and never put production keys in `.env`.

| Variable | Purpose |
| --- | --- |
| `APP_KEY` | Application key, set by the setup script |
| `DATABASE_URL` | Local PostgreSQL from `docker-compose.yml` |
| `TEST_DATABASE_URL` | PostgreSQL database used by the test suite |
| `S3_*` | S3-compatible storage; fill with the `RUSTFS_*` values of `docker-compose.yml` for local runs |
| `FLAG_INGEST_ENABLED`, `FLAG_INGEST_<SOURCE>` | Data ingestion jobs, all off by default |
| `FLAG_SOCIAL` | Follows and activity feed, off by default |
| Integrations (Stripe, Turnstile, analytics, search verification, mail) | Off while their variables are empty |

### Common commands

| Command | Purpose |
| --- | --- |
| `vukem serve` | Development server with Vite HMR |
| `vukem build` and `bun run build:ssr` | Production build: Go binary, Vite assets and SSR bundle |
| `go test ./...` | Go unit and integration tests (PostgreSQL via Docker Compose) |
| `golangci-lint run` | Go lint and format check |
| `bun run lint`, `bun run typecheck`, `bun run test` | Biome, TypeScript and Vitest |
| `bun run openapi:lint` | Redocly lint of the OpenAPI documents |
| `bun run build:docs` | Re-render `resources/docs/*.md` into `resources/docs/docs.json` |

## Project structure

| Path | Contents |
| --- | --- |
| `resources/js/` | Inertia and React pages; SSR entry at `resources/js/ssr.tsx` |
| `packages/tokens`, `packages/app-client` | Bun workspace packages used by the web app |
| `apps/mobile/` | Android app (Expo) |
| `sdks/ts`, `sdks/python` | API clients generated from the OpenAPI document |
| `openapi/` | OpenAPI document and goldens |
| `config/` | Go configuration, such as `config/billing.go` |
| `scripts/` | CI, infrastructure, legacy import and release scripts |
| `setups/` | Local setup scripts for Windows and Unix |
| `qa/` | Playwright end-to-end tests and Schemathesis API tests |
| `docs/` | Design documents (`docs/sdd/`), decisions (`docs/adrs/`), runbooks (`docs/runbooks/`) and legal notes (`docs/legal/`) |
| `.github/workflows/` | CI workflow (`ci.yml`) |

## Testing and CI

The CI workflow in `.github/workflows/ci.yml` runs on pushes to `main`, on pull requests, weekly, and on demand. Its jobs cover lint (golangci-lint, gofmt, Biome, `tsc`, markdownlint and commit messages), tests, migrations, dependency audits, govulncheck, the Docker image, OpenAPI, Lighthouse and the mobile app. The end-to-end job runs only when the repository variable `RUN_E2E` is `true`.

To run the CI jobs locally in one command:

```bash
scripts/ci/local.sh
```

It runs every job except end-to-end in parallel, and needs Docker, PostgreSQL on port `5433`, Go, Bun and golangci-lint. Use `--only` or `--skip` to select jobs.

Other checks:

```bash
go test -tags integration ./...
bun run e2e
bunx markdownlint-cli2@0.24.0 "**/*.md"
```

## Mobile app

The Android app in `apps/mobile` is built with Expo SDK 58 and NativeWind ([ADR 0018](docs/adrs/0018-nativewind.md)). The `android/` and `ios/` folders are generated and not committed.

```bash
bun run --cwd apps/mobile lint
bun run --cwd apps/mobile typecheck
bun run --cwd apps/mobile test
```

Release steps are in [docs/runbooks/mobile-release.md](docs/runbooks/mobile-release.md).

## Stability

HexaScore follows [Semantic Versioning](https://semver.org) and tags releases as `vX.Y.Z` on `main`. The first release is `v0.1.0`. `v1.0.0` marks parity with the legacy application and a successful cutover ([ADR 0013](docs/adrs/0013-versioning-releases-cutover.md)).

Changes are listed in [CHANGELOG.md](CHANGELOG.md), which follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## Contributing

Thank you for considering a contribution to HexaScore.

- Read the design documents in `docs/` (SDDs in `docs/sdd/`, decisions in `docs/adrs/`) before changing behaviour.
- Write the test first. Run the specific test, lint and typecheck that cover your change.
- Security-relevant behaviour needs a test named `TestSecurity_<Axx>_<Behaviour>`.
- Update the relevant SDD, `CHANGELOG.md` (under `[Unreleased]`) and the documentation in the same change when behaviour, commands or conventions change.

### Commits

Use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) in English: `type(scope): description`.

- Types: `feat`, `fix`, `perf`, `refactor`, `docs`, `test`, `build`, `ci`, `chore`, `style`, `revert`.
- Scopes: `auth`, `catalog`, `ratings`, `lists`, `billing`, `api`, `admin`, `ui`, `infra`, `deps`, `repo`, `docs`.

The commit-msg hook enabled by the setup script checks the header locally, and commitlint checks the commits and the PR title in CI.

### Pull requests

1. Create a branch from `main`.
2. Make one logical change per commit.
3. Open a pull request against `main`.

`main` accepts changes only through a pull request with green required checks and linear history.

## Code of conduct

Keep discussions respectful and constructive. This repository does not publish a separate code of conduct file yet.

## Security vulnerabilities

Please do not open a public issue for a security vulnerability. Report it privately to the maintainer, [@AlexGalhardo](https://github.com/AlexGalhardo). This repository does not publish a separate security policy file yet.

## License

HexaScore is open source software released under the [MIT license](LICENSE).

## Acknowledgements

HexaScore builds on the following data sources and projects. Attribution follows the terms recorded in the repository documents.

- **Wikidata**: structured data released under CC0. It is the backbone of the catalogue and of the paid API.
- **Wikipedia**: extracts released under CC BY-SA. Attribution and the source URL are stored with each field.
- **TMDB (The Movie Database)**: movie and TV metadata, shown on the website and the mobile app but not in the paid API. Where TMDB fields are used, the TMDB logo and this notice apply: *This product uses the TMDB API but is not endorsed or certified by TMDB.*
- **IGDB (IGDB.com)**: game metadata. Attribution to IGDB.com is required on item pages and the about page.
- **Open Library**: bibliographic data from [openlibrary.org](https://openlibrary.org). See [docs/legal/open-library.md](docs/legal/open-library.md) for the licence check.
- **vukem**: the Go framework HexaScore is built on, at [vukem.dev](https://vukem.dev).
