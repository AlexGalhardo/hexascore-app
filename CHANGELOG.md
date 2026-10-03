# Changelog

All notable changes to this project are documented here.
Format: [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/). Versioning: [SemVer](https://semver.org/).

## [Unreleased]

### Added

- `PLAN.md` with decisions and milestones; reference docs in `docs/` (architecture, data model, API, frontend, design system, payments, ingestion, security, testing, process).
- `AGENTS.md` and `CLAUDE.md` agent guides.
- `infra/docker-compose.yaml`: Postgres 18.6, Redis 8.10.2 (password-protected, not published to the host) and an Upstash-compatible REST proxy (`serverless-redis-http` 0.0.10, port 8079) with healthchecks; ports bound to 127.0.0.1 and credentials documented as dev-only; optional `app` profile running the web app from `infra/Dockerfile` (`oven/bun:1.4.2`, multi-stage, non-root, migrates on start).
- Idempotent local setup scripts in `setups/` (Unix/Windows × Docker/PGlite): check Bun/Docker, install, create/fill `apps/web/.env.local` with generated secrets, start containers, migrate, seed when available.
- CI workflow (`lint`, `typecheck`, `test`, `build`, `audit`) on PRs and pushes to `dev`/`main`, with a shared `setup-bun` composite action and Bun install cache.
- CodeQL workflow (JavaScript/TypeScript + Python), gated on `CODE_SCANNING_ENABLED` for private repos.
- release-please (`simple`, manifest at `0.0.0`, Keep a Changelog sections) on pushes to `main`.
- Root `README.md` with quickstart, commands and docs index.
- `.github/rulesets/protect-main-dev.json`: ruleset for `main`/`dev` (PR required, checks `lint`/`typecheck`/`test`/`build`, no force push/deletion), ready to apply once the GitHub plan allows it.

### Changed

- Dependabot now covers `bun`, `uv` (`webscrappings/`), `github-actions`, `docker` and `docker-compose` (`infra/`), grouped minor/patch, targeting `dev`.
- `docs/process.md` and `docs/testing.md` list the implemented workflows and exact commands.

### Removed

- Files inherited from the previous project (RespondeAê): old plan, docs, setup scripts, tests, REST client samples, graphify output, `vercel.json`.
- Inherited Prisma-era `ci.yml` and `e2e.yml` workflows (replaced by the new CI; E2E returns in a later milestone).
