# HexaScore (vukem rewrite): guide for AI agents

HexaScore (hexascore.app) is a catalogue, rating and recommendation platform for games, movies, series and books, with a paid read-only public API and an Android app. This repository is the Go rewrite on the [vukem](https://vukem.dev) framework; it replaces the Next.js app kept on branch `legacy` of `github.com/AlexGalhardo/hexascore-app`.

This file is the single source of truth for agent behaviour. `CLAUDE.md` only points here.

## Session start (mandatory)

1. Load these skills before any work, in this order: `karpathy-guidelines`, `graphify`, `caveman`. Availability check on 2026-10-28: `karpathy-guidelines` and `caveman` exist as skills; `graphify` exists only as a CLI (`graphify update .` / `graphify query "<question>"`), its skill folder has no `SKILL.md`. If the skill cannot be loaded, run the CLI instead and note it in the session; never silently skip.
2. Read [docs/PROGRESS.md](docs/PROGRESS.md) (state, decisions, pending items) and [ROADMAP.md](ROADMAP.md) (next task).
3. Pick the first unchecked microtask of the current roadmap phase unless the owner names another.

## Where things live

| Path | What |
| --- | --- |
| `docs/PROGRESS.md` | programme log per phase; update at the end of every session |
| `docs/brainstorming.md` | owner decisions D-xx and architect decisions A-xx; final unless edited first |
| `docs/web-research.md` | competitor research with sources |
| `docs/sdd/` | one design document per domain; `open-issues.md` is the triage of their open points |
| `docs/adrs/` | architecture decision records (MADR) |
| `architect-feedback-and-roadmap.md` | architect's assessment and executive summary |
| `ROADMAP.md` | phases, tasks, subtasks, microtasks with acceptance criteria |
| `vukem-feedback.md` | framework bugs, gaps and DX feedback (VF-xxx); append, never delete |
| `CHANGELOG.md` | Keep a Changelog; generated per release, `[Unreleased]` edited by hand only before release |
| `legacy/` (sibling folder, not in this repo) | read-only Next.js reference; history on branch `legacy` |
| `vukem/` (sibling folder, own git) | framework checkout for local `go.work`/`replace` only, never committed |
| `.claude/` | project skills and rules for Claude Code (adapted from vukem-marketplace-reactjs) |

## Rules

- Language: everything in this repository is English (code, identifiers, comments, docs, commits). Chat with the owner in pt-BR. UI strings in `en`, `pt-BR`, `es` only through the i18n catalogues.
- Framework facts come from <https://vukem.dev/docs> (cite the page); the local `vukem/docs` folder is not a reference. Never invent a vukem API: write `[NOT IN VUKEM DOCS]` and add an entry to `vukem-feedback.md`.
- Native vukem subsystems first; a third-party Go or npm dependency needs the owner's yes and an ADR. Pinned exact versions only (no `^`, no `latest`, no pre-releases); Docker images with fixed tags or digests.
- Decisions in `docs/brainstorming.md` and `docs/adrs/` are final; to change one, edit the decision first, then the SDD, then the code.
- Test first. A microtask is done only when its acceptance commands in `ROADMAP.md` pass. Prefer the specific test, lint and typecheck over the whole suite; the full gate runs before a PR.
- Security defaults are never weakened; every security behaviour has a `TestSecurity_<Axx>_<Behaviour>` test (vukem convention).
- Surgical changes: touch only what the task needs; no speculative abstractions; comment only the why.
- Git: Conventional Commits; feature branches `agent/<role>/<task-id>` merged straight into `main` by the orchestrator after review (owner rule 2026-10-29: no `dev`/staging step, do not wait for CI to merge or start the next task); tags `vX.Y.Z` on `main`; never rewrite history or force-push; branch `legacy` is immutable. Commit and push only when the owner asks.
- Releases follow the vukem pattern (`vX.Y.Z` tags one at a time, generated changelog, GitHub release); versions start at `v0.1.0`; `v1.0.0` = parity with the legacy app and successful cutover.
- Documentation: update the SDD, `CHANGELOG.md` and this file in the same change when behaviour, commands or conventions change. Never claim a file was updated without editing it.
- Parallel agents and worktrees only with the owner's explicit yes in the session; sequential research or review subagents are fine.

## Commands (fill in as the codebase appears; keep in sync with ROADMAP tasks)

| Command | Purpose |
| --- | --- |
| `vukem serve` | development server with Vite HMR |
| `vukem build` + `bun run build:ssr` | production build (Go binary, Vite assets, SSR bundle) |
| `go test ./...` | unit and integration tests (Postgres via `docker compose`) |
| `golangci-lint run` | Go lint and format check |
| `golangci-lint run --build-tags integration ./...` | Go lint incl. integration tests (CI pins v2.15.0) |
| `bunx commitlint --from <base-sha> --to <head-sha>` | Conventional Commits check on a commit range (CI `lint`; PR title via stdin) |
| `bun run build:docs` | re-render `resources/docs/*.md` into the committed `resources/docs/docs.json` (a vitest test fails when stale) |
| `bun run lint` / `bun run typecheck` / `bun run test` | TypeScript lint, types, unit tests |
| `vukem openapi:export -prefix /api/v1` (or `/app/v1`) | export an OpenAPI document; `scripts/openapi-check.sh` diffs both against the goldens in `openapi/` (CI `lint`), `--write` refreshes them |
| `bun run sdk:generate` / `bun run sdk:generate:python` | regenerate `sdks/ts/src/gen` and `sdks/python/hexascore` from `openapi/api-v1.json` (TS tests: `cd sdks/ts && bun test`; Python needs `uv`: `cd sdks/python && uv run pytest -q`) |
| `bun run e2e` | Playwright + axe against a built container |
| `docker compose up -d --wait` | local Postgres 18.6 (127.0.0.1:5433) and RustFS (127.0.0.1:9000/9001), both healthy |
| `setups/setup-windows.ps1 [-NoServices]` | idempotent Windows setup: tools, `core.hooksPath .githooks` (commit-msg Conventional Commits check), `.env`, `APP_KEY`, compose stack, `bun install`, `vukem migrate`; `-NoServices` skips compose and migrate (CI Windows job) |
| push to `main` | Railway service `app` (source `AlexGalhardo/hexascore-app`, branch `main`) builds and deploys automatically; `.github/workflows/ci.yml` reports regressions after the fact |
| `railway up --service app --detach` | manual deploy of the working tree to Railway project `hexascore`, environment `production` (<https://app-production-785f.up.railway.app>; healthcheck `/up`, pre-deploy `/app/app migrate`) |
| `./app hexascore:predeploy` | pre-deploy entry point: drop `INVALID` index leftovers, then `migrate` (runbook `docs/runbooks/migrations.md`) |
| `./app admin:grant <email>` / `admin:revoke <email>` / `user:reset-2fa <email> --reason "<text>"` | role changes and manual 2FA reset (account_events with actor NULL, `data.cli=true`); admins are users with `role = admin`, not `ADMIN_EMAILS` |
| `scripts/release.sh vX.Y.Z` | changelog section via git-cliff (`.config/cliff.toml`), `chore(release)` commit and tag; needs git-cliff installed (`scoop install git-cliff` / `brew install git-cliff`) |
| `setups/setup-unix.sh` | same setup for Linux and macOS |

## Models

- Development in the main session with the model the owner chose; roadmap tasks state a suggested model (Opus 5.5 for architectural or risky work, Sonnet 5.5 for well-bounded work).
- External documentation research: `researcher` subagent. Text and i18n edits: `text-editor` subagent.
