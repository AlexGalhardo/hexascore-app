# HexaScore (vukem rewrite): guide for AI agents

HexaScore (hexascore.app) is a catalogue, rating and recommendation platform for games, movies, series and books, with a paid read-only public API and an Android app. This repository is the Go rewrite on the [vukem](https://vukem.dev) framework; it replaces the Next.js app kept on branch `legacy` of `github.com/AlexGalhardo/hexascore-app`.

This file is the single source of truth for every coding agent. `CLAUDE.md` only points here.

## Session start (mandatory)

1. Read this file, [docs/PROGRESS.md](docs/PROGRESS.md), and [ROADMAP.md](ROADMAP.md).
2. Read [.claude/skills/use-railway/SKILL.md](.claude/skills/use-railway/SKILL.md) and follow it for every Railway command. The skill is vendored from [railwayapp/railway-skills](https://github.com/railwayapp/railway-skills) commit `526e95524e964ed326689cbac4464d97a58ab81c` (`use-railway` 1.7.0). Also read [Grok Build + Railway](https://railway.com/agents/grok): Railway MCP is `https://mcp.railway.com`, and a deploy is a plan the owner approves before `railway up`.
3. If you are one of the team workers (senior, intern, specialist), your system prompt is [PROMPT-WORKERS.md](PROMPT-WORKERS.md) and your brief is `.orchestration/briefs/<role>.md`; claim work only through `.orchestration/queue/TASKS.md`. Otherwise pick the first unchecked microtask of the current phase unless the owner names another.
4. `graphify` is a CLI (`graphify query "<question>"`) when `graphify-out/graph.json` exists. It is not required to start.

Railway rules that override the skill's deploy-from-cwd shortcut: do not run `railway up`, `railway up -y`, or `railway up --new` unless the owner explicitly asks to deploy. Do not `git push origin main` (that push deploys production). Run the CLI from this repo, pass the linked project, environment, and service explicitly, and never print variable values. Do not create a staging environment, an extra region, or resize a volume without an explicit yes.

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
| `vukem-feedback.md` | open framework bugs, gaps and DX feedback (VF-xxx); remove an entry once it is resolved on vukem `main` (owner, 2026-10-31; git keeps history) |
| `CHANGELOG.md` | Keep a Changelog; generated per release, `[Unreleased]` edited by hand only before release |
| `legacy/` (sibling folder, not in this repo) | read-only Next.js reference; history on branch `legacy` |
| `vukem/` (sibling folder, own git) | framework checkout for local `go.work`/`replace` only, never committed |
| `PROMPT-WORKERS.md` | system prompt of the execution team (orchestrator, 2 seniors, intern, specialist): roles, queue protocol, report formats |
| `.grok/rules/` | language rules, loaded without a vendor compatibility mode |
| `.claude/` | the same rules and skills, kept so Claude-compatible tools still see them |
| `.orchestration/` | `queue/TASKS.md` (single task board), `briefs/`, `reports/`, `handoff/`; versioned in git |
| `docs/orchestration/` | `BOARD.md` (readable mirror of the queue) and `QUESTIONS.md` (owner gates and answers) |
| `apps/mobile` | Android Expo app (SDK 58, NativeWind). Do not commit `android/` or `ios/` |

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
- Release gate (owner rule 2026-10-31): before opening any release PR, run every job of `.github/workflows/ci.yml` locally, in parallel, on the exact release commit (`scripts/ci/local.sh` does this in one command), and open the PR only when all of them pass: lint (golangci-lint run and `fmt --diff`, Biome, `tsc`, markdownlint, `scripts/openapi-check.sh`, `bun run openapi:lint`, commitlint on the range since the last tag), tests (`go test -tags integration ./...` on Postgres, `bun test scripts/ci/migration-policy.test.ts`, SDK typecheck and tests), migrations round trip, deps (osv-scanner, `bun audit`), govulncheck, Docker build plus `scripts/infra/check-image.sh` and the `/up` smoke test, OpenAPI (`--image` diff, Redocly, Schemathesis), Lighthouse, and mobile lint/typecheck/test. Playwright e2e is excluded (owner, 2026-10-31): the CI `e2e` job runs only when the repo variable `RUN_E2E` is `true` (unset, so skipped) and is not part of the release gate. The `main` branch on GitHub requires a PR, green status checks and linear history; a red check on the release PR is a process failure.
- Documentation: update the SDD, `CHANGELOG.md` and this file in the same change when behaviour, commands or conventions change. Never claim a file was updated without editing it.
- Team (owner, 2026-10-31): one orchestrator plus four persistent workers (two seniors, one intern, one specialist) as defined in `PROMPT-WORKERS.md`; the specialist reviews every delivery and merges to local `main`. Parallel workers need the owner's explicit yes in each session. Do not `git push origin main`: that deploys Railway production.
- vukem work (bugs and features the app exposes) is a senior task in the sibling `vukem/` checkout, reviewed by the specialist, merged by PR with green CI; vukem releases only on the owner's request. Never commit a `replace` directive or `go.work` in this repository.

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
| `bun run --cwd apps/mobile lint` / `typecheck` / `test` | Android app (Biome, `tsc`, Jest). `catalogues` recopies the bundled i18n namespaces |
| `bunx eas-cli@27.2.0 build` / `submit` / `update` (in `apps/mobile`) | mobile release with EAS: follow `docs/runbooks/mobile-release.md` (skill `hexascore-mobile-release`); owner confirms version, every submit and production OTA; CI: manual `mobile-release.yml` |
| `vukem openapi:export -prefix /api/v1` (or `/app/v1`) | export an OpenAPI document; `scripts/openapi-check.sh` diffs both against the goldens (`--image` exports from a built image; CI `openapi`), `--write` refreshes them |
| `scripts/ci/local.sh [--only a,b] [--skip a,b] [--base REV] [--tree] [--keep]` | release gate in one command: every `ci.yml` job except e2e, in parallel, on a clean checkout of HEAD, with its own databases, ports and compose project; one log per job in `.ci-local/logs/`, summary table, exit 1 if any job fails (needs Docker, the compose Postgres on 5433, go, bun, golangci-lint; `setup-windows` runs only on Windows) |
| `qa/schemathesis/ci.sh [image]` | migrate the CI Postgres, insert a business key and an app token, run Schemathesis 4.29.0 against that image (Linux `--network host`) |
| `bun run openapi:lint` | Redocly lint of both goldens (`redocly.yaml`; CI `lint`) |
| `bun run sdk:generate` / `bun run sdk:generate:python` | regenerate `sdks/ts/src/gen` and `sdks/python/hexascore` from `openapi/api-v1.json` (TS tests: `cd sdks/ts && bun test`; Python needs `uv`: `cd sdks/python && uv run pytest -q`) |
| `bun run e2e` | Playwright + axe against a built container (CI `e2e`: `scripts/ci/qa-up.sh` loads `image.tar`, seeds DemoCatalog, uploads traces on failure) |
| `bunx --bun @lhci/cli@0.15.1 autorun` | Lighthouse budgets in `lighthouserc.cjs` (base URL from `E2E_BASE_URL`) on the five DemoCatalog URLs (CI `lighthouse`; `@lhci/cli` is not a package.json dependency) |
| `docker compose up -d --wait` | local Postgres 18.6 (127.0.0.1:5433) and RustFS (127.0.0.1:9000/9001), both healthy |
| `setups/setup-windows.ps1 [-NoServices]` | idempotent Windows setup: tools, `core.hooksPath .githooks` (commit-msg Conventional Commits check), `.env`, `APP_KEY`, compose stack, `bun install`, `vukem migrate`; `-NoServices` skips compose and migrate (CI Windows job) |
| push to `main` | Railway service `app` (source `AlexGalhardo/hexascore-app`, branch `main`) builds and deploys automatically; `.github/workflows/ci.yml` reports regressions after the fact |
| `railway up --service app --detach` | manual deploy of the working tree to Railway project `hexascore`, environment `production` (<https://hexascore.app>, also <https://www.hexascore.app>, the paid API at <https://api.hexascore.app/v1> and <https://app-production-785f.up.railway.app>; DNS on Vercel nameservers since the 2026-10-31 cutover; healthcheck `/up`, pre-deploy `/app/app migrate`) |
| `./app hexascore:import-legacy --stage all --dry-run` | legacy import plan and pre-flight over schema `legacy_import`; without `--dry-run` it imports stage by stage from checkpoints (SDD 11 §7.6); the `twofactor` stage needs `--stdin` fed by `scripts/legacy/twofactor-export.ts` |
| `./app hexascore:cutover-verify [--counts-only] [--sample-pct N]` | read-only cutover check after the import: counts, id-sample md5, cover keys on the `S3_*` bucket, orphan FKs, SQL vs Go scores (SDD 11 §7.6 step 6); exit 1 names the failed tables |
| `./app seed:mark --email-suffix @example.test --legacy-prefix seed-user- [--dry-run]` / `./app seed:purge [--dry-run] [--batch 500] --confirm` | flag the fake legacy seed users (`users.is_seed`, never able to sign in), later delete them in batches and queue the score recompute and search refresh of the items they touched (SDD 11 §7.6) |
| `./app backup:restore <key> --target-url <url>` / `./app backup:drill` | restore an encrypted dump and replay newer account deletions; monthly restore drill into `hexascore_drill` (runbook `docs/runbooks/restore.md`; tests need `pg_dump`/`pg_restore`, skipped when absent) |
| `./app backup:offsite` | mail `ADMIN_EMAILS` the 7-day signed link to this month's encrypted monthly dump (day 1, 07:00; runbook `docs/runbooks/restore.md` "Off-site copy") |
| `./app hexascore:predeploy` | pre-deploy entry point: drop `INVALID` index leftovers, then `migrate` (runbook `docs/runbooks/migrations.md`) |
| `./app admin:grant <email>` / `admin:revoke <email>` / `user:reset-2fa <email> --reason "<text>"` | role changes and manual 2FA reset (account_events with actor NULL, `data.cli=true`); admins are users with `role = admin`, not `ADMIN_EMAILS` |
| `scripts/release.sh vX.Y.Z` | changelog section via git-cliff (`.config/cliff.toml`), `chore(release)` commit and tag; needs git-cliff installed (`scoop install git-cliff` / `brew install git-cliff`) |
| `setups/setup-unix.sh` | same setup for Linux and macOS |

## Optional integrations (vukem v0.36.0)

All off while their variables are unset; names in `.env.example`, owner steps in [docs/analytics-acess.md](docs/analytics-acess.md).

- Analytics and verification: `UMAMI_WEBSITE_ID`, `UMAMI_DOMAINS`, `GA_MEASUREMENT_ID`, `GOOGLE_SITE_VERIFICATION`, `BING_SITE_VERIFICATION` render through `{{analytics}}` in `app.html`. GA4 waits for the cookie banner's Accept (ADR 0023); Umami is cookieless.
- SEO: `INDEXNOW_KEY` (key file `/{key}.txt`, job `seo.indexnow_changes` every minute), `BING_WEBMASTER_API_KEY` (client config only). `/llms.txt` and `/llms-full.txt` are always served.
- Turnstile: `TURNSTILE_SITE_KEY` + `TURNSTILE_SECRET_KEY` (+ `TURNSTILE_HOSTNAME`) protect `POST /signup`, `/login`, `/contact` and enable the scrape-guard step-up at `POST /turnstile` (ADR 0016 amendment). Local and tests use Cloudflare's test keys; `TURNSTILE_VERIFY_URL` points siteverify at a fake and refuses to boot in production.

## Agents

- Roles `orchestrator`, `senior`, `estagiario` (intern) and `especialista` (specialist) are defined in `PROMPT-WORKERS.md` and `.orchestration/briefs/`. The names are not tied to a vendor; the model per role is a recommendation from the owner (Opus for orchestrator and specialist, Sonnet for seniors, Haiku for the intern).
- Roadmap lines that say Opus or Sonnet are effort hints, not a required runtime.
- Local QA: `E2E_IMAGE_TAG=<tag> docker compose --profile app build app-init && docker compose --profile app up -d --wait` builds the current tree and serves it on `127.0.0.1:3000` with the DemoCatalog seed (`admin@example.test` / `Demo-password-1`); test in a browser before marking UI work done. Parallel stacks (one per worker): own project, host port and database, never `down` on project `hexascore`: `docker exec hexascore-postgres-1 createdb -U hexascore hexascore_qa_<tag>`, export `E2E_IMAGE_TAG=<tag> QA_PORT=<3100+> QA_DATABASE_URL=postgres://hexascore:hexascore@host.docker.internal:5433/hexascore_qa_<tag>?sslmode=disable`, then `docker compose --env-file <main checkout>/.env -p hexascore-qa-<tag> --profile app` with `build app-init`, `run --rm --no-deps app-init`, `up -d --wait --no-deps --no-build app`; run e2e and Lighthouse with `E2E_BASE_URL=http://localhost:<port> E2E_APP_CONTAINER=hexascore-qa-<tag>-app-1 E2E_DB_NAME=hexascore_qa_<tag>`; stop with `stop app` and `dropdb`. The image runs with `APP_ENV=production`, so CanonicalHost answers 301 to the compose `APP_URL` (`http://localhost:<port>`) for any other Host: use `localhost`, not `127.0.0.1`, as the base URL.
