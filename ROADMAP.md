# HexaScore roadmap (vukem rewrite)

Execution plan for any coding agent. Read [AGENTS.md](AGENTS.md) first, then [docs/PROGRESS.md](docs/PROGRESS.md). The team that executes it (orchestrator, two seniors, one intern, one specialist) and its queue protocol are defined in [PROMPT-WORKERS.md](PROMPT-WORKERS.md); state lives in `.orchestration/`. A model name on a microtask is an effort hint, not a required runtime. Tick boxes as work lands; update `docs/PROGRESS.md` at the end of each session.

Ordering principle (owner requirement): (1) foundation and parity with the Next.js app, (2) data migration and infra, (3) public API, (4) community edits, (5) mobile, (6) continuous ingestion and recommendations. Phase 0 is a short framework spike that de-risks everything after it.

Conventions:

- IDs: `P<phase>.<group>.<task>.<microtask>` (example `P1.B.2.3`). Reference them in commit bodies (`Refs: P1.B.2.3`).
- Model: `Opus` = architectural, cross-cutting or risky; `Sonnet` = well-bounded, pattern-following.
- Acceptance: commands with the expected decisive output; "green" means exit code 0 and no new failures. Every microtask also requires: lint clean, tests of the touched package green, the SDD updated if behaviour changed.
- Releases: one minor tag per task group (ADR 0013): `v0.N.0` before cutover, `v1.N.0` after. Before staging exists (P2.A) a release is a tag, a GitHub release and a container smoke test on the developer machine. `v1.0.0` closes Phase 2 (P2.G).
- Resolved conflicts between phase files: the trending API (P3.A.1.4) follows SDD 08 (top 10 per type, 30-day window, no `window=` parameter) and depends on P6.C; `/items/{id}/related` and `/me/for-you` app routes live in P6.D; the Google sign-in library for React Native (P5.A.3.4) is an owner decision with an ADR (architect default: Expo first-party `expo-auth-session`); `contrib/engine` is `[NOT IN VUKEM DOCS]` on vukem.dev, fallback `exec.CommandContext` with timeout (P6.B.2.1); Steam/GOG terms are recorded in `docs/legal/steam-gog.md` before their images are used.

## Status

| Field | Value |
| --- | --- |
| Current phase | P1 done (`v0.9.0`). P2 in progress: code-only items of P2.A to P2.F merged; 39 boxes wait for owner gates (Railway, DNS, Stripe live, real data, cutover). P3.A/P3.C and P4.A were built ahead of the cutover (`v0.11.0`). P2.H (launch readiness, audit 2026-10-31) is the next code-only group |
| Last release | app `v0.11.0` (2026-10-31, vukem `v0.34.0`, deployed by `release.yml`); passkeys stay off until `adotar VF-001` |
| Next microtask | P2.H.1.1 (home and catalogue covers), then the rest of P2.H; owner-gated P2 items as the owner releases them (`liberar <ID>` in `developer-manual-todos.md`) |
| v1.0.0 definition | parity with the legacy app and a successful cutover (ADR 0013), plus P2.H done: the audit of 2026-10-31 found the UI at wireframe level, which is not a public launch |
| Updated | 2026-10-31 |

## Resuming work (read first in a new session)

The work runs through [PROMPT-WORKERS.md](PROMPT-WORKERS.md) (system prompt of the team) and [`.orchestration/`](.orchestration/) (queue, reports, handoffs; versioned in git). Parallel workers need the owner's explicit yes in each new session (global CLAUDE.md rule).

1. Read [`.orchestration/handoff/orchestrator.md`](.orchestration/handoff/orchestrator.md) (state at pause), then [docs/orchestration/QUESTIONS.md](docs/orchestration/QUESTIONS.md) (owner gates and answers) and `C:\dev\hexascore-app\developer-manual-todos.md` (owner checklist, outside the repo).
2. `git pull`, `docker compose up -d --wait`, and run the gate once: `go build ./...`, `golangci-lint run --allow-parallel-runners --build-tags integration ./...`, `bun install && bun run lint && bun run typecheck && bun run test`, `TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable go test -tags integration -parallel 4 -timeout 30m ./...`.
3. Fill `.orchestration/queue/TASKS.md` from the first unchecked microtasks that have no open owner gate (P2.H first), then launch the four workers with the briefs in `.orchestration/briefs/`. Workers stay alive and take the next task from the queue; relaunch only after a context-limit stop.
4. Orchestrator loop: when the specialist marks a task `PRONTO_PARA_MERGE`, merge the branch with `git merge --no-ff`, run the gate, tick the microtask, renumber `VF-NEW-*` entries (next free: VF-063). `main` is pushed only when the owner asks (the push deploys Railway production).
5. Releases: `scripts/release.sh vX.Y.Z` on a clean tree (git-cliff), local container smoke test (`hexascore:predeploy` on a fresh database, then `/up?checks=1`), `git push origin main && git push origin vX.Y.Z`; the Release workflow publishes the GitHub release.

Open owner items: real Stripe values on Railway service `app` (Q-H31); the 39 owner-gated P2 boxes (Railway, buckets, DNS, Telegram, secrets, spend, real data, cutover) in `developer-manual-todos.md`; `adotar VF-001` (passkeys); the D-06 amendment for the social layer (Phase 7, question Q-A1 in QUESTIONS.md).

## Phase map

| Phase | Goal | Exit criteria | Release |
| --- | --- | --- | --- |
| P0 Framework spike | Prove the ten riskiest vukem capabilities on a throwaway app and close or confirm the related `vukem-feedback.md` entries | Spike report in `docs/spike-report.md`; each capability marked works / workaround / blocked | none |
| P1 Foundation and parity (done 2026-10-30) | The Go app serves every legacy page and flow kept by D-06 on a developer machine (staging exists only after P2.A), with the legacy QA suites ported | Parity checklist 118/118; Playwright 660 passed, k6, Schemathesis, ZAP, Lighthouse green on a built container | `v0.1.0` … `v0.9.0` (one per group P1.A–P1.I) |
| P2 Data migration and infra | Railway production, CI/CD, observability, security hardening, data migrated, launch readiness (P2.H), cutover done | DNS on Railway, rollback window started, backup drill passed, P2.H exit checks green | `v0.10.0` … `v0.16.0`, then `v1.0.0` at P2.G |
| P3 Public API | DX items beyond parity: changes feed, ID exports, webhooks, entity endpoints, provenance metadata, SDK publishing | Schemathesis green on the new document; SDKs published | `v1.1.0` (P3.A), `v1.2.0` (P3.B), `v1.3.0` (P3.C) |
| P4 Community edits | Suggestions, moderation queue, audit, revert, notifications | SLA dashboard live; first 100 suggestions processed in staging rehearsal | `v1.4.0` (P4.A), `v1.5.0` (P4.B) |
| P5 Mobile | Android app on Google Play internal → production track | Store listing approved; crash-free sessions ≥ 99 % in internal testing | `v1.6.0`–`v1.8.0`; app `mobile-v1.0.0` |
| P6 Ingestion and recommendations | Ingestion in Go with asserts; trending, similar items, personal recommendations | Daily jobs green for 30 days; trending and similar live | `v1.9.0`–`v1.12.0` |
| P7 Social layer (post-1.0; D-06 amended 2026-10-31) | Follows, activity feed, reviews with comments and reactions, richer profiles and shelves (Skoob, Goodreads, Letterboxd references) | Feed p95 ≤ 300 ms at 30 rps; moderation queue covers reviews and comments; Playwright + axe green | `v1.13.0`–`v1.15.0` |

## Task groups (summary)

Microtasks for each group are in the phase sections below.

| Group | Scope | Main SDD/ADR | Depends on |
| --- | --- | --- | --- |
| P0.F | Framework spike: auth scaffold + Better Auth hash check, Inertia SSR in one container, cursor pagination, host routing, migrations + concurrent index script, queue/scheduler inside `serve`, OpenAPI export + client generation, Stripe scaffold, storage bucket, Lighthouse CI sample | all SDDs; ADR 0001, 0004 | — |
| P1.A | Repository bootstrap: `vukem new`, Go module pinned to vukem, Vite/React/TypeScript, Biome, Vitest, tokens package, i18n package, `docker compose`, Taskfile/scripts, CHANGELOG, commitlint | SDD 09, 11; ADR 0013 | P0.F |
| P1.B | Catalogue model and read paths: items, editions, entities, relations, translations, alt titles, links, visibility, provenance, slugs, redirects, merge, sitemap data | SDD 01; ADR 0006, 0012 | P1.A |
| P1.C | Auth and accounts: vukem auth + social + 2FA, roles, bans, deletion grace, tokens with abilities, bot protection, mail templates | SDD 03, 13; ADR 0008, 0016 | P1.A |
| P1.D | Scoring: star input, storage, aggregate formula v1, confidence, anti-manipulation, status, time-to-beat, nightly recompute, public formula page | SDD 02; ADR 0005 | P1.B, P1.C |
| P1.E | Search: text search configurations, trigger-maintained tsvectors, trigram indexes, suggest table, ranking, golden queries | SDD 07; ADR 0011 | P1.B |
| P1.F | User library: lists, library views, profiles, reports, contact, notifications inbox, data export | SDD 14 | P1.C, P1.D |
| P1.G | Web app parity: Inertia pages, SSR, i18n routing, design tokens, star widget, CSP, SEO, docs pages, legal pages, admin panel wiring | SDD 09; ADR 0004 | P1.B–P1.F |
| P1.H | API parity: `/api/v1` read endpoints as legacy, keys, plans, quota metering in Postgres, Stripe subscriptions, OpenAPI document, Scalar UI, legacy SDK regeneration | SDD 05; ADR 0015, 0003 | P1.B, P1.C, P1.D |
| P1.I | QA parity gate: Playwright + axe specs ported, k6 scenarios, Schemathesis, ZAP baseline, Lighthouse CI budgets, parity checklist | SDD 09, 12, 13 | P1.G, P1.H |
| P2.A | Railway: project, environments, Postgres, bucket, domains, container, health checks, config-as-code, secrets | SDD 11; ADR 0002 | P1.A |
| P2.B | CI/CD and releases: GitHub Actions pipeline, deploy to staging/production, git-cliff releases, version in `/up` | SDD 11; ADR 0013 | P2.A |
| P2.C | Observability: logs, OTel to Grafana Cloud, metrics catalogue, alerts incl. SSR fallback and projected spend, Telegram channel, status page, runbooks | SDD 12 | P2.A |
| P2.D | Security hardening: headers, rate limits table, proof-of-work challenge, LGPD endpoints, admin 2FA, dependency scanning, ZAP active scan | SDD 13; ADR 0016 | P1.C, P2.B |
| P2.E | Backups and migrations policy: dump job, encryption, retention, restore drill, advisory lock, concurrent index script | SDD 11; ADR 0014 | P2.A |
| P2.F | Data migration: export from Neon, purge synthetic users, transform and import (users, catalogue, editions, ratings, statuses, lists, keys, subscriptions), covers copy, verification queries | SDD 11, 06, 03; A-04 | P1.I, P2.E |
| P2.G | Cutover: rehearsal on staging, read-only window, DNS switch, monitoring checklist, rollback runbook, 30-day retention, decommission list | SDD 11; ADR 0013 | P2.B–P2.F |
| P2.H | Launch readiness (audit 2026-10-31): covers and visual catalogue, item page layout, CSR skeletons, hydration feedback, API DX, flaky test, debt from the audit findings | SDD 09, 05, 14 | P1.I |
| P3.A | API DX: entity endpoints, trending endpoint, editions, provenance metadata and coverage table, `changes` feed, daily ID exports, webhooks, deprecation headers | SDD 05; ADR 0006 | P2.G |
| P3.B | SDKs and docs: TypeScript (npm, JSR) and Python (PyPI) publishing, docs pages, changelog, pricing page coverage statement | SDD 05 | P3.A |
| P3.C | First-party app API `/app/v1`: tokens, abilities, endpoints for mobile, generated TypeScript client package, client-errors/metrics intake | SDD 05, 03, 14; ADR 0009 | P2.G |
| P4.A | Suggestions model and submission: tables, state machine, quotas, anti-abuse, submission UI, status pages | SDD 04; ADR 0010 | P2.G |
| P4.B | Moderation and application: admin queue pages, apply/revert transactional writes, audit log, notifications, metrics and SLA alerts | SDD 04, 12 | P4.A |
| P5.A | Mobile foundation: Expo app, navigation, tokens, i18n, auth with tokens, generated client, offline cache | SDD 10; ADR 0009 | P3.C |
| P5.B | Mobile features: browse, item, search, rating widget, status, lists, profile, notifications | SDD 10 | P5.A |
| P5.C | Mobile release: EAS/Gradle pipeline, Play listing, account deletion page, data safety, internal → production | SDD 10, 11 | P5.B |
| P6.A | Ingestion in Go: connector interface, source_records, Wikidata/Wikipedia/Open Library connectors, normaliser, parity test vs Python | SDD 06; ADR 0007 | P2.G, legal notes (D-36, A-05) |
| P6.B | Ingestion extras: TMDB/IGDB/Steam/GOG display-only enrichment with provenance and expiry, covers pipeline (`cwebp`), trailers, edition matching, entity resolution queue, data-quality asserts, admin status | SDD 06 | P6.A |
| P6.C | Trending: activity primitives, flush, snapshots, home sections, API endpoint | SDD 08 | P2.G |
| P6.D | Similar items and personal recommendations: nightly similarity, cross-media links, "For you" with opt-out, evaluation protocol | SDD 08 | P6.C, P6.A |
| P7.A | Follows and activity feed: `follows` table, follower counts, activity events, `/feed` page, privacy rules | SDD 14 (new §12), SDD 08 | P2.G, D-06 amendment |
| P7.B | Reviews, comments and reactions: text review per rating, threaded comments, reactions, reports and moderation for all three | SDD 14, SDD 04, SDD 13 | P7.A, P4.B |
| P7.C | Social discovery: profile shelves with covers, "friends rated", list collaboration, share cards | SDD 09, SDD 08 | P7.A, P6.C |

## Phase 0: framework spike

Goal: prove the ten riskiest vukem capabilities on a throwaway app before the real repository exists, and turn every related `vukem-feedback.md` entry into "works", "workaround" or "blocked". The spike app lives in the sibling folder `C:\dev\hexascore-app\hexascore-spike` (never committed, deleted after P0.F.11.1); only `docs/spike-report.md` and `vukem-feedback.md` changes land in this repository. Each microtask appends its findings (commands run, decisive output, verdict) to a scratch section of `docs/spike-report.md` so P0.F.11.1 only consolidates.

vukem facts cited below come from <https://vukem.dev/docs/cli/reference-new>, `/docs/cli/reference-install`, `/docs/cli/reference-runtime`, `/docs/starters`, `/docs/frontend/inertia`, `/docs/security/authentication`, `/docs/digging-deeper/billing`, `/docs/rest-apis/openapi`, `/docs/rest-apis/client-generation`, `/docs/database/migrations` and `/docs/testing/overview` (read 2026-10-28).

### P0.F Framework spike

#### P0.F.1 Auth scaffold and Better Auth hash compatibility

- [x] **P0.F.1.1** Create the throwaway Inertia React app on Postgres and run the auth scaffold with Google and 2FA, recording the generated tables, columns, pages, routes and session env names.
  - Files/modules: `hexascore-spike/` (via `vukem new hexascore-spike --inertia react --pm bun -framework ../vukem`), `bootstrap/app.go` (Postgres driver), `docker-compose.yml` with `postgres:18.6-alpine3.24`, `docs/spike-report.md` (section "Auth scaffold")
  - Refs: SDD 03 §5, §6.1, §11 issues 7 and 10; SDD 11 §7.7; ADR 0008; <https://vukem.dev/docs/starters> "Inertia (React, Vue, Svelte)"; <https://vukem.dev/docs/cli/reference-install> "`vukem install:auth`"
  - Depends on: —
  - Model: Sonnet
  - Acceptance:
    - `vukem install:auth -social google -2fa -locales en,pt-BR,es` → exit 0 and `app/accounts/accounts.go` exists
    - `vukem migrate && vukem migrate:status --json` → `create_users_table` and `create_auth_tables` listed as ran on Postgres 18
    - `go test ./tests/...` → `ok` (scaffold tests on Postgres through `pgtest`, `TEST_DATABASE_URL` set)
  - Notes: closes open issues 03-7 (session timeout env names) and 03-10 (column and page names) by writing the exact names into the report; SDD 03 is updated later in P1.C.1.1, not here. `-framework` is the documented local-checkout flag (reference-new error table); never commit a `replace`. `install:auth` on an Inertia app writes React `Auth/*` pages (starters page): record whether they are TypeScript and whether they reuse the starter layout.

- [x] **P0.F.1.2** Pin Better Auth 1.7.7's password hash format from its own source and produce fixture hashes from the legacy seed script.
  - Files/modules: `legacy/` (read-only; run `bun install` in a temporary copy if writing `node_modules` into `legacy/` is not wanted), `legacy/apps/web/scripts/seed-admin.ts` or `admin-grant.ts`, spike `testdata/betterauth_hashes.json`
  - Refs: SDD 03 §7.2, §11 issue 1; open-issues 03-1, 03-2; ADR 0008
  - Depends on: —
  - Model: Opus
  - Acceptance:
    - `rg -n "scrypt|NFKC" <copy>/node_modules/better-auth/dist` → the N, r, p, dkLen and salt length constants quoted in the report with file and line
    - `jq length testdata/betterauth_hashes.json` → ≥ 3 (ASCII password, accented password to exercise NFKC, 32-character password)
  - Notes: open-issues 03-1 is a `roadmap` row: this is that task. Never use real user hashes; generate fixtures from throwaway passwords. 2FA secrets (03-2, decided) are exported later by a Bun script over stdin; here only record which Better Auth function decrypts them.

- [x] **P0.F.1.3** Prove a compatibility hasher that verifies Better Auth scrypt hashes and rehashes to argon2id on login.
  - Files/modules: spike `app/accounts/compat_hasher.go`, `app/accounts/compat_hasher_test.go`, user model wiring for `NeedsRehash` and `PasswordUpdater`
  - Refs: SDD 03 §7.2, §9; ADR 0008; <https://vukem.dev/docs/security/authentication> "The idea" (hasher, optional interfaces)
  - Depends on: P0.F.1.1, P0.F.1.2
  - Model: Opus
  - Acceptance:
    - `go test ./app/accounts -run TestSecurity_A07_LegacyScryptVerifiesAndRehashes` → Better Auth 1.7.7 fixture hash verifies, stored value then starts `$argon2id$`, wrong password fails
  - Notes: `golang.org/x/crypto/scrypt` becomes a direct dependency: ADR 0008 names it, but AGENTS.md still requires the owner's yes before `go get`. Constant-time compare only. If vukem exposes no hook to wrap `hash.Default()`, mark the capability "workaround" and file a VF entry.

#### P0.F.2 Inertia SSR in one container

- [x] **P0.F.2.1** Enable Inertia SSR with a Vite SSR bundle, a failure counter and an `ssr` health check.
  - Files/modules: spike `bootstrap/app.go` (`inertia.SSRConfig{Bundle: "bootstrap/ssr/ssr.js", OnFailure: ...}`), `resources/js/ssr.tsx`, `package.json` script `build:ssr` (`vite build --ssr`)
  - Refs: SDD 09 §6.1, §9; ADR 0004; VF-002, VF-027; <https://vukem.dev/docs/frontend/inertia> "Server-side rendering"; <https://vukem.dev/docs/cli/reference-runtime> "`vukem ssr:start` and `vukem ssr:check`"
  - Depends on: P0.F.1.1
  - Model: Opus
  - Acceptance:
    - `bun run build:ssr && vukem ssr:start` then `vukem ssr:check` → exit 0
    - `curl -s localhost:3000/ | grep -c 'data-page'` → 1 and the body contains server-rendered markup (not only the root `div`)
    - `curl -s localhost:3000/up?checks=1` → `ssr: ok`; after killing the SSR process → `ssr: fail` and the fallback counter > 0
  - Notes: `OnFailure` defaults to a warning log (inertia page): the spike must prove a counter increments per failure (VF-002 workaround). The health-check wiring (`vukem.WithHealthCheck`, `app.Inertia.SSRHealthy`) is cited by SDD 09 from `/docs/deployment/docker-and-railway`; if the API differs, record it.

- [x] **P0.F.2.2** Run Go and the Node SSR server in one container, with the Go process starting and restarting Node.
  - Files/modules: spike `Dockerfile` (starter Dockerfile plus a pinned Node runtime), `app/ssr/supervisor.go`
  - Refs: SDD 11 §6.1, §6.3; ADR 0004; D-07, D-08; VF-027
  - Depends on: P0.F.2.1
  - Model: Opus
  - Acceptance:
    - `docker build -t hexascore-spike .` → exit 0; image size recorded in the report
    - `docker run -p 3000:3000 hexascore-spike` then `curl -s localhost:3000/up?checks=1` → `ssr: ok`
    - `docker exec <id> pkill node` then within 5 s `curl -s localhost:3000/up?checks=1` → `ssr: ok` again (restarted)
  - Notes: `vukem build` does not cover the SSR bundle (VF-027): the Dockerfile runs `vite build --ssr` itself. Pin the Node image tag exactly. Use `vukem ssr:start -restart-delay` as a child process if that is enough, otherwise a small supervisor; record which and why.

- [x] **P0.F.2.3** Prove URL-prefix locales, the i18n `Once` prop and the theme cookie under SSR.
  - Files/modules: spike `app/http/middleware/locale_prefix.go`, `resources/views/app.html` (theme view function), `lang/en.json`, `lang/pt-BR.json`
  - Refs: SDD 09 §6.1, §7.1; open-issues 09-1, 09-5, 09-6; VF-024, VF-028; <https://vukem.dev/docs/frontend/inertia-i18n> "Testing"
  - Depends on: P0.F.2.1
  - Model: Sonnet
  - Acceptance:
    - `curl -s localhost:3000/pt-br/ | grep -o '<html lang="[^"]*"'` → `<html lang="pt-BR"`
    - `curl -sI -H 'Cookie: locale=pt-BR' localhost:3000/` → 307, `Location: /pt-br/`
    - `curl -s -H 'Cookie: theme=dark' localhost:3000/ | grep -c 'data-theme="dark"'` → 1
  - Notes: VF-024 workaround (path-prefix middleware before vukem's i18n middleware) and VF-028 workaround (view function). A second Inertia visit with `X-Inertia-Except-Once-Props: i18n:pt-BR` must omit `messages`.

#### P0.F.3 Cursor pagination

- [x] **P0.F.3.1** Implement a signed keyset cursor over `(sort value, id)` with `WhereRaw` and feed it to `inertia.Scroll`.
  - Files/modules: spike `app/catalog/cursor.go`, `cursor_test.go`, a 300-row seeded `items` table, one Inertia page using `Scroll`
  - Refs: SDD 01 §7.2, §9; SDD 05 §6.2 "Cursors"; open-issues 01-4, 05-8, 07-4, 09-4; VF-005
  - Depends on: P0.F.1.1
  - Model: Opus
  - Acceptance:
    - `go test ./app/catalog/... -run TestCursor` → PASS (stable order on ties, tampered cursor → `invalid-cursor`, cursor from another sort refused)
    - an Inertia partial visit for `items` with the next cursor → the next 24 rows, no duplicates across 13 pages
  - Notes: vukem `req.CursorPaginate` pages by primary key and rejects `sort` with a cursor (SDD 05 quotes `/docs/rest-apis/pagination` "Query conventions"), hence the app-signed cursor (HMAC with `APP_KEY`). Record the `ScrollMeta` shape accepted with an opaque cursor; this closes 07-4/09-4 or selects the fallback (plain props + `WhenVisible`).

#### P0.F.4 Host routing

- [x] **P0.F.4.1** Serve `api.<host>/v1/*` from the `/api/v1` route group with a host-matching middleware and set the OpenAPI `servers` entry.
  - Files/modules: spike `app/http/middleware/api_host.go`, `api_host_test.go`, `routes/api.go`
  - Refs: SDD 05 §6.1; open-issues 05-1; VF-015; <https://vukem.dev/docs/digging-deeper/billing> "Receiving webhooks yourself" (`r.UseServer` example)
  - Depends on: P0.F.1.1
  - Model: Sonnet
  - Acceptance:
    - `curl -s -H 'Host: api.localhost' localhost:3000/v1/ping` → same body as `curl -s localhost:3000/api/v1/ping`
    - `curl -s -o /dev/null -w '%{http_code}' -H 'Host: api.localhost' localhost:3000/` → `404`
    - `vukem openapi:export -prefix /api/v1 | jq -r '.servers[0].url'` → the API host URL, or the report records `[NOT IN VUKEM DOCS]` and the post-processing step used
  - Notes: in Git Bash write `-prefix //api/v1` (reference-runtime error table). The middleware runs before routing so route names and CSRF exemptions still match.

#### P0.F.5 Migrations and the concurrent index script

- [x] **P0.F.5.1** Prove raw `jsonb`, `text` + `CHECK` enums and a generated column through migrations and the ORM.
  - Files/modules: spike `database/migrations/*_create_items_table.go`, `*_create_ratings_table.go`, `tests/orm_generated_test.go`
  - Refs: SDD 01 §5.2; SDD 02 §5; open-issues 01-5, 02-5; VF-006, VF-009; <https://vukem.dev/docs/database/migrations> "Creating migrations", "Running migrations"
  - Depends on: P0.F.1.1
  - Model: Sonnet
  - Acceptance:
    - `vukem migrate && vukem migrate:rollback && vukem migrate` → exit 0 each
    - `go test ./tests -run TestGeneratedOverallColumn` → insert and update through the ORM succeed and `overall` equals `round(sum/6)`; if the ORM writes the column, the test documents the raw-SQL fallback
  - Notes: `s.Raw` for `jsonb` and CHECK constraints (VF-006 workaround). Record whether `t.JSON` maps to `jsonb` on Postgres.

- [x] **P0.F.5.2** Prove an advisory lock around the migration runner and a one-off `CREATE INDEX CONCURRENTLY` script outside a transaction.
  - Files/modules: spike custom command `migrate:locked` (via `vukem make:command`), `scripts/concurrent-index.sh`
  - Refs: SDD 11 §7.2; SDD 07 §7.10; open-issues 07-3, 11-3, 11-4; VF-017, VF-026; ADR 0014
  - Depends on: P0.F.5.1
  - Model: Opus
  - Acceptance:
    - two parallel runs of the locked command → one prints `Migrated:` lines, the other waits and prints `Nothing to migrate.`; both exit 0
    - `scripts/concurrent-index.sh items_title_trgm` → `pg_index.indisvalid` is `t` and no `AccessExclusiveLock` on `items` was observed in `pg_locks` during the build
  - Notes: if the runner exposes no hook to take the lock, run migrations as a separate pre-deploy command (VF-026 alternative) and record it.

#### P0.F.6 Queue and scheduler inside `serve`

- [x] **P0.F.6.1** Run database-driver queue workers and the scheduler inside `serve`, with per-family semaphores, `WithoutOverlapping` and jobs that survive a restart mid-run.
  - Files/modules: spike `app/jobs/recompute.go`, `app/jobs/semaphore.go`, `app/jobs/counter.go`, `routes/schedule.go`, `.env` (`QUEUE_CONNECTION=database`, `QUEUE_NAMES`, `QUEUE_CONCURRENCY`, `SCHEDULE_TIMEZONE=America/Sao_Paulo`)
  - Refs: SDD 02 §6 (jobs); SDD 11 §6.1; ADR 0007; open-issues 06-5, 08-10, 11-7; VF-018, VF-021; <https://vukem.dev/docs/digging-deeper/queue> "Inside serve"; <https://vukem.dev/docs/digging-deeper/scheduler> "Inside serve (default)"
  - Depends on: P0.F.5.1
  - Model: Opus
  - Acceptance:
    - `vukem serve` log contains `scheduler running inside serve`
    - dispatch 200 jobs of two families, `kill -TERM` the process mid-run, start again → every job's effect is applied exactly once (idempotency key) and the `heavy` family never exceeds its semaphore (asserted by a counter table)
    - `go test ./app/jobs/... -run TestBufferedCounterFlush` → PASS
  - Notes: queues are strict priority (queue page), so lower queues starve; the semaphore is the VF-018 workaround. `HourlyAt` is not documented (VF-021): use cron expressions. Tasks must honour `ctx` on shutdown (`SHUTDOWN_TIMEOUT`).

#### P0.F.7 OpenAPI export and client generation

- [x] **P0.F.7.1** Export two OpenAPI 3.1 documents by prefix, with RFC 9457 problems carrying kebab-case `code` extension members.
  - Files/modules: spike `routes/api.go`, `routes/app_api.go` (typed handlers via `api.Register`), `app/http/problems/problem.go`, `openapi/api-v1.json`, `openapi/app-v1.json`
  - Refs: SDD 05 §6.2, §6.5, §9; SDD README "Problem codes", "JSON field names"; open-issues 05-6, 05-7; VF-014, VF-016; <https://vukem.dev/docs/rest-apis/openapi>; <https://vukem.dev/docs/cli/reference-runtime> "`vukem openapi:export`"
  - Depends on: P0.F.4.1
  - Model: Opus
  - Acceptance:
    - `vukem openapi:export -prefix /api/v1 > openapi/api-v1.json` → exit 0; same with `-prefix /app/v1 > openapi/app-v1.json`
    - `curl -s -H 'Host: api.localhost' localhost:3000/v1/items/missing` → `application/problem+json` with `"code":"not-found"` and `request_id`
    - `go test ./tests -run TestAPIContract` → every response passes `AssertMatchesOpenAPI()`
  - Notes: record whether envelope keys can be camelCase (VF-014) and how extension members are emitted (VF-016). Page routes stay out of the documents through `-prefix` (duplicate `operationId` error in reference-runtime).

- [x] **P0.F.7.2** Generate the first-party TypeScript client and typed routes, and run the legacy SDK generators against the new document.
  - Files/modules: spike `packages/app-client/src/api.ts`, `resources/js/routes.ts`, temporary copies of `legacy/sdks/ts` and `legacy/sdks/python`
  - Refs: SDD 05 §6.5; SDD 09 §6.1; ADR 0009; <https://vukem.dev/docs/cli/reference-runtime> "`vukem codegen:client`", "`vukem codegen:routes`"; <https://vukem.dev/docs/rest-apis/client-generation>
  - Depends on: P0.F.7.1
  - Model: Sonnet
  - Acceptance:
    - `vukem codegen:client -prefix /app/v1 -o packages/app-client/src/api.ts` → exit 0 and `bunx tsc --noEmit -p packages/app-client` → 0 errors
    - `vukem codegen:routes -o resources/js/routes.ts` → exit 0
    - `cd sdks/ts && bun run sdk:generate && bun test`; `cd sdks/python && uv run pytest -q` → pass
  - Notes: `vukem codegen:types` (cited by SDD 09 §6.1) is not on the reference-runtime page read on 2026-10-28: write `[NOT IN VUKEM DOCS]` in the report and check whether `codegen:client` output types are enough for Inertia props. SDK generators keep the legacy pinned versions (`@hey-api/openapi-ts` 0.99.0, openapi-python-client 0.29.1).

#### P0.F.8 Stripe scaffold

- [x] **P0.F.8.1** Run the billing scaffold, rename plans to `pro`/`business` with per-currency prices, and pass the fake-Stripe tests.
  - Files/modules: spike `config/billing.go`, `routes/billing.go`, `tests/billing_test.go`, `.env.example` (`STRIPE_PRICE_PRO`, `STRIPE_PRICE_BUSINESS`)
  - Refs: SDD 05 §6.3, §7.5, §9; ADR 0015; D-14; <https://vukem.dev/docs/cli/reference-install> "`vukem install:billing`"; <https://vukem.dev/docs/digging-deeper/billing> "Configuration", "Testing"
  - Depends on: P0.F.1.1
  - Model: Sonnet
  - Acceptance:
    - `vukem install:billing` → exit 0; `vukem route:list` shows `GET /billing`, `POST /billing/checkout/{plan}`, `POST /billing/portal`, `POST /stripe/webhook`
    - `go test ./tests/... -run 'TestBilling|TestStripeWebhook'` → duplicate processed once, forged 400, `past_due` → free
  - Notes: the scaffold writes `STRIPE_PRICE_BASIC`/`STRIPE_PRICE_PRO`; record how BRL/USD/EUR prices per plan (SDD 05 §6.3) fit `config/billing.go`, or mark `[NOT IN VUKEM DOCS]` and file a VF entry. `contrib/stripe` is required at its latest stable release (install page). Test keys only.

#### P0.F.9 Storage bucket

- [x] **P0.F.9.1** Write, read, presign and serve objects on an S3-compatible disk (RustFS locally) and write a queued export ZIP into the bucket.
  - Files/modules: spike `config/storage.go`, `app/covers/serve.go`, `app/exports/zip_job.go`, `docker-compose.yml` (`rustfs/rustfs:1.0.4`)
  - Refs: SDD 11 §5 (bucket prefixes), §7.7; SDD 14 §7.4; SDD 05 §6.2 (`getExport` presign); open-issues 06-4, 11-1, 11-2, 14-6; VF-029; <https://vukem.dev/docs/digging-deeper/storage>
  - Depends on: P0.F.6.1
  - Model: Sonnet
  - Acceptance:
    - `curl -sI localhost:3000/covers/<id>-<hash16>.webp` → 200, `Cache-Control: public, max-age=31536000, immutable`
    - a queued export job writes `exports/<user>/<id>.zip`; the presigned URL downloads it before expiry and answers 403 after
  - Notes: export helpers only target `http.ResponseWriter` (VF-029): prove the in-memory `ResponseRecorder` workaround or the re-implementation, whichever is smaller. Legacy key format `covers/<id>-<hash16>.webp` is kept (06-4).

#### P0.F.10 Lighthouse CI sample

- [x] **P0.F.10.1** Run Lighthouse CI against one SSR page of the spike container with the SDD 09 budgets as assertions.
  - Files/modules: spike `lighthouserc.json`, `package.json` (`@lhci/cli` exact version)
  - Refs: SDD 09 §4, §9; ADR 0004 (budgets); D-22
  - Depends on: P0.F.2.2
  - Model: Sonnet
  - Acceptance:
    - `bunx lhci autorun` → assertions pass for `/` on mobile emulation (LCP ≤ 2.0 s, CLS ≤ 0.05, TBT ≤ 200 ms, total JS ≤ 150 kB gzip)
  - Notes: Lighthouse CI is approved by D-22; pin the latest stable version on the day. Record run time so P1.I.5.1 can size the step.

#### P0.F.11 Spike report and closure

- [x] **P0.F.11.1** Consolidate the spike into `docs/spike-report.md` and update every exercised `vukem-feedback.md` entry.
  - Files/modules: `docs/spike-report.md`, `vukem-feedback.md`, `docs/PROGRESS.md` (Phase 0 entry), `ROADMAP.md` status table
  - Refs: ROADMAP phase map (P0 exit criteria); VF-002, VF-005, VF-006, VF-009, VF-014 to VF-018, VF-021, VF-024, VF-026 to VF-029
  - Depends on: P0.F.1.3, P0.F.2.2, P0.F.2.3, P0.F.3.1, P0.F.4.1, P0.F.5.2, P0.F.6.1, P0.F.7.2, P0.F.8.1, P0.F.9.1, P0.F.10.1
  - Model: Opus
  - Acceptance:
    - `rg -c '^\| P0\.F\.([1-9]|10) ' docs/spike-report.md` → 10 (one row per capability: verdict works / workaround / blocked, commands, decisive output)
    - `rg -n '\*\*Status\*\*' vukem-feedback.md` → every VF entry exercised by the spike carries a dated note (`open` with the proven workaround, `reported` with an issue link, or `withdrawn` when the capability already works)
  - Notes: append, never delete VF entries (AGENTS.md). A "blocked" verdict stops the affected Phase 1 group until the owner decides (edit `docs/brainstorming.md` first). Delete the spike folder only after the owner has read the report.

## Phase 1: foundation and parity

Parity baseline (checked by P1.I.6.1):

- Routes: `legacy/apps/web/app/[locale]/` (home, `[section]`, `[section]/[slug]`, `search`, `u/[nickname]`, `u/[nickname]/lists/[id]`, `pricing`, `contact`, `policy`, `terms`, `login`, `signup`, `verify-email`, `recover-password`, `reset-password`, `onboarding`, `profile`, `profile/api`, `profile/api/checkout`, `admin` with `contact`, `reports`, `items`, `items/[id]`, `ingestion`, `docs` with `authentication`, `errors`, `pagination`, `rate-limits`, `changelog`, `reference`, `[...rest]`).
- The 33 Server Actions in `legacy/apps/web/actions/*.ts`: `account.ts` (`updateName`, `setNickname`, `setPassword`, `setProfilePrivate`, `scheduleDeletion`, `cancelDeletion`); `admin.ts` (`setContactHandled`, `setReportStatus`, `setListHidden`, `banUser`, `saveAffiliateLink`, `deleteAffiliateLink`, `mergeItems`); `api-keys.ts` (`createApiKeyAction`, `revokeApiKeyAction`); `billing.ts` (`startStripeCheckout`, `openBillingPortal`, `createPixCharge`, `simulatePixPaymentAction`); `contact.ts` (`sendContactMessage`); `follows.ts` (`toggleFollow`); `item-status.ts` (`saveItemStatus`); `lists.ts` (`createList`, `updateList`, `deleteList`, `addToList`, `removeFromList`, `moveInList`); `ratings.ts` (`saveRating`, `deleteRating`); `reports.ts` (`report`); `time-to-beat.ts` (`saveTimeToBeat`, `deleteTimeToBeat`). `docs/PROGRESS.md` gives the count; the names come from the files.
- The 16 specs in `legacy/qa/e2e/specs/`: `admin`, `api-keys`, `auth`, `catalog`, `contact`, `docs`, `headers`, `item`, `not-found`, `pricing`, `profile`, `public-pages`, `ratings`, `search`, `social`, `time-to-beat`.
- Deferred by D-06 (out of Phase 1): `createPixCharge`, `simulatePixPaymentAction`, `toggleFollow`, `/feed`, the follow/feed tests of `social.spec.ts`, credit packs, per-locale OpenAPI documents. Critic scores are dropped (D-09). Each microtask below names the legacy actions and routes it replaces.

Releases: ADR 0013 makes each group that ships to staging a `v0.N.0` release. Staging is provisioned by P2.A; until it exists the release microtask tags, publishes the GitHub release and smoke-tests the built image locally, and the staging deploy of that tag is done once P2.A lands. Procedure used by every release microtask (SDD 11 §7.1): `main` green; `git-cliff --config .config/cliff.toml --unreleased --tag vX.Y.Z` preview; `task release -- vX.Y.Z` (writes the changelog section, commits `chore(release): vX.Y.Z`, tags); `git push origin main`; `git push origin vX.Y.Z` (one tag per push); `gh run list --workflow release.yml`; release body not empty. Commit, tag and push only when the owner asks in that session (AGENTS.md).

### P1.A Repository bootstrap

#### P1.A.1 Project scaffold

- [x] **P1.A.1.1** Scaffold the real application from the Inertia React starter with bun and the pinned vukem module, on Postgres.
  - Files/modules: `go.mod` (`github.com/AlexGalhardo/vukem` exact version, no `replace`), `bootstrap/app.go`, `cmd/app/`, `resources/js/`, `resources/views/app.html`, `package.json` (`packageManager` bun), `.env.example`, `Dockerfile`, `.gitignore` (`go.work`, `.env`)
  - Refs: SDD 09 §7.10; SDD 11 §6.3, §7.7; ADR 0001, 0004; <https://vukem.dev/docs/cli/reference-new> "`vukem new`"; <https://vukem.dev/docs/starters> "Inertia (React, Vue, Svelte)", "What every starter shares"
  - Depends on: P0.F.11.1
  - Model: Sonnet
  - Acceptance:
    - `vukem new hexascore --inertia react --pm bun --module github.com/AlexGalhardo/hexascore-app` (in a temp folder, files moved into this repo) → exit 0
    - `go test ./...` → `ok` without Node; `bun run typecheck` → 0 errors
    - `go list -m github.com/AlexGalhardo/vukem` → an exact `vX.Y.Z`, not a pseudo-version
  - Notes: keep the files that already exist (`docs/`, `AGENTS.md`, `ROADMAP.md`, ...). Delete the starter's example posts CRUD here. Postgres driver wiring as proven in the spike. `install:auth` is P1.C.1.1, not here.

- [x] **P1.A.1.2** Add the local stack (Postgres 18.6, RustFS) and the Windows/Unix setup scripts.
  - Files/modules: `docker-compose.yml`, `setups/setup-windows.ps1`, `setups/setup-unix.sh`, `.env.example` (`DATABASE_URL`, `TEST_DATABASE_URL`, storage keys), `AGENTS.md` commands table
  - Refs: SDD 11 §5, §7.7; `legacy/setups/`; <https://vukem.dev/docs/testing/overview> "Testing against PostgreSQL"
  - Depends on: P1.A.1.1
  - Model: Sonnet
  - Acceptance:
    - `docker compose up -d && docker compose ps` → `postgres` (`postgres:18.6-alpine3.24`) and `rustfs` (`rustfs/rustfs:1.0.4`) healthy
    - `pwsh setups/setup-windows.ps1` on Windows 11 → exit 0, then `vukem migrate` → exit 0
  - Notes: bind Postgres to `127.0.0.1` only. Image tags exactly as SDD 11.

#### P1.A.2 Tooling and conventions

- [x] **P1.A.2.1** Add Biome, Vitest, Testing Library, strict TypeScript, `.editorconfig` and `golangci-lint` with pinned versions.
  - Files/modules: `biome.json`, `vitest.config.ts`, `tsconfig.json` (`strict`), `.editorconfig` (tabs, width 4; YAML with spaces), `.golangci.yml`, `package.json` scripts `lint`, `typecheck`, `test`
  - Refs: SDD 09 §7.10, §9; D-22; AGENTS.md "Rules"
  - Depends on: P1.A.1.1
  - Model: Sonnet
  - Acceptance:
    - `bunx biome ci resources/js packages && bun run typecheck` → 0 errors
    - `bun run test` → pass; `golangci-lint run` → 0 issues
  - Notes: exact versions in `package.json` (no `^`). `golangci-lint` is a dev binary, pinned in the setup scripts. markdownlint is not in D-22: propose it to the owner, do not add it without a yes.

- [x] **P1.A.2.2** Add Conventional Commits enforcement, the git-cliff config, the Keep a Changelog file and the release task.
  - Files/modules: `.config/cliff.toml` (adapted from vukem), `CHANGELOG.md` (`[Unreleased]` with Added/Changed/Fixed/Removed), `Taskfile.yml` (`release` task), commit-msg hook, `.github/workflows/release.yml` skeleton
  - Refs: ADR 0013; SDD 11 §7.1; D-20
  - Depends on: P1.A.2.1
  - Model: Sonnet
  - Acceptance:
    - `git-cliff --config .config/cliff.toml --unreleased --tag v0.1.0` → a section grouped by commit type
    - the commit-msg hook rejects `update stuff` and accepts `feat(catalog): add slugs`
  - Notes: owner input needed: commitlint, Task and git-cliff are not in D-22 (git-cliff is named by ADR 0013 and SDD 11). Ask one yes/no for the set; the zero-dependency fallback is a POSIX `scripts/release.sh` and a shell commit-msg hook, and SDD 11 §7.1 is then updated.

#### P1.A.3 Shared packages

- [x] **P1.A.3.1** Create `packages/tokens` with design tokens emitted as CSS variables and a TypeScript module, light and dark.
  - Files/modules: `packages/tokens/tokens.json`, `packages/tokens/build.ts`, `resources/css/app.css`, `packages/tokens/tokens.test.ts`
  - Refs: SDD 09 §7.5; ADR 0009; SDD 10 (consumer)
  - Depends on: P1.A.2.1
  - Model: Sonnet
  - Acceptance:
    - `bun run --cwd packages/tokens build && bunx vitest run packages/tokens` → pass, including a WCAG AA contrast assertion for every foreground/background pair
  - Notes: values start from the legacy Tailwind theme. Base UI enters with the components in P1.G.1.1.

- [x] **P1.A.3.2** Create `packages/i18n`, port the legacy catalogues to flat JSON, and serve them as the Inertia `Once` prop with the catalogue hash in the asset version.
  - Files/modules: `packages/i18n/src/flatten.ts`, `lang/en.json`, `lang/pt-BR.json`, `lang/es.json` (from `legacy/apps/web/messages/{en,pt-BR,es}/`), `bootstrap/app.go` (`inertia.Config{Version: ...}`), `app/i18n/plural_test.go`
  - Refs: SDD 09 §6.1, §9; open-issues 02-6, 09-5; ADR 0009; <https://vukem.dev/docs/frontend/inertia-i18n>
  - Depends on: P1.A.1.1
  - Model: Opus
  - Acceptance:
    - `go test ./app/i18n/... -run TestPluralGolden` → Go and TS outputs identical
    - `bunx vitest run packages/i18n` → every key in `en` exists in `pt-BR` and `es`
  - Notes: port only namespaces used by kept pages (no feed, follows or PIX strings, D-06). Copy work goes to the `text-editor` subagent.

#### P1.A.4 Platform middleware

- [x] **P1.A.4.1** Productionise the spike's locale-prefix middleware, prefixed route helpers and theme view function.
  - Files/modules: `app/http/middleware/locale_prefix.go`, `app/http/middleware/locale_prefix_test.go`, `resources/views/app.html`, `resources/js/lib/route.ts` (wraps `vukem codegen:routes` output with the prefix)
  - Refs: SDD 09 §7.1, §9; D-17; VF-024, VF-028
  - Depends on: P1.A.3.2
  - Model: Sonnet
  - Acceptance:
    - `curl -sI -H 'Cookie: locale=pt-BR' $WEB/` → 307, `Location: /pt-br/`
    - `go test ./app/http/middleware/... -run TestLocalePrefix` → PASS (`/pt-br`, `/es`, no prefix = `en`, unknown prefix = 404)
  - Notes: `en` has no prefix; a request without a locale cookie is never redirected.

- [x] **P1.A.4.2** Wire SSR supervision, the `ssr` health check, the fallback counter and the `build:ssr` step into the real app and its Dockerfile.
  - Files/modules: `app/ssr/supervisor.go`, `app/ssr/supervisor_test.go`, `bootstrap/app.go` (`SSRConfig` with `Except`), `Dockerfile`, `package.json` `build:ssr`
  - Refs: SDD 09 §6.1, §9; SDD 11 §6.3; ADR 0004; VF-002, VF-027
  - Depends on: P1.A.1.1
  - Model: Opus
  - Acceptance:
    - `curl -s $WEB/up?checks=1` → `ssr: ok`; after killing the SSR process → `ssr: fail` and `inertia_ssr_fallback_total` > 0.
    - `docker build -t hexascore:dev . && docker run --rm -p 3000:3000 hexascore:dev` → `curl -s localhost:3000/up` 200
  - Notes: `SSRConfig.Except`: `/admin`, `/profile`, `/me`, `/billing`, `/suggest`, `/suggestions` (SDD 09 §6.1). The OTLP exporter is P2.C; here the counter only needs to be observable in tests.

- [x] **P1.A.4.3** Add the advisory-locked migration command and the concurrent index script to the real app.
  - Files/modules: `app/console/migrate_locked.go`, `app/console/migrate_locked_test.go`, `scripts/concurrent-index.sh`, `docs/runbooks/migrations.md`
  - Refs: SDD 11 §7.2; SDD 07 §7.10; ADR 0014; VF-017, VF-026
  - Depends on: P1.A.1.2
  - Model: Sonnet
  - Acceptance:
    - `go test ./app/console -run TestMigrateLockedSerialises` → PASS
    - two parallel `./bin/app migrate:locked` runs → one migrates, the other prints `Nothing to migrate.`; both exit 0
  - Notes: same design as P0.F.5.2; if the spike chose a pre-deploy command, implement that and update SDD 11 §7.2.

#### P1.A.5 Release

- [x] **P1.A.5.1** Release `v0.1.0` (bootstrap) with the vukem-style process.
  - Files/modules: `CHANGELOG.md`, tag `v0.1.0`, GitHub release
  - Refs: ADR 0013; SDD 11 §7.1; D-20
  - Depends on: P1.A.1.2, P1.A.2.2, P1.A.3.1, P1.A.3.2, P1.A.4.1, P1.A.4.2, P1.A.4.3
  - Model: Sonnet
  - Acceptance:
    - `gh release view v0.1.0` → non-empty body
    - the tagged image → `curl -s localhost:3000/up?checks=1` shows `ssr: ok`
  - Notes: owner input needed: the first push of this repository to `main` of `github.com/AlexGalhardo/hexascore-app` replaces the Next.js tree (PROGRESS.md workspace notes); the owner runs or explicitly approves it. Branch `legacy` is never touched.

### P1.B Catalogue model and read paths

#### P1.B.1 Schema

- [x] **P1.B.1.1** Migrate `data_sources`, `items` (visibility, release precision and status, provenance), `item_translations` and `item_alt_titles`.
  - Files/modules: `database/migrations/*_create_data_sources_table.go`, `*_create_items_table.go`, `*_create_item_translations_table.go`, `*_create_item_alt_titles_table.go`, `app/models/item.go`
  - Refs: SDD 01 §5.2, §5.3, §9; ADR 0012; open-issues 02-1, 07-6, 08-3; VF-006
  - Depends on: P1.A.5.1
  - Model: Opus
  - Acceptance:
    - `vukem migrate` on empty Postgres 18 → exit 0; `vukem migrate:status --json` lists every catalog migration as ran.
    - `go test -tags=integration ./app/catalog/... -run TestItemConstraints` → PASS (type CHECK, unique `(type, slug)`, unique `wikidata_qid`, `visibility` in `public|hidden|takedown`)
  - Notes: UUID v7 keys (SDD README conventions). Search columns and indexes belong to P1.E.1.1. `external_scores` is not created (D-09).

- [x] **P1.B.1.2** Migrate entities and relations (`entities`, `entity_translations`, `entity_links`, `item_entities`, `item_links`, `entity_stats`) and schedule the entity stats refresh job.
  - Files/modules: `database/migrations/*_create_entities_table.go` and the five siblings, `app/models/entity.go`, `app/catalog/roles.go`, `app/catalog/entity_stats.go`, `app/jobs/refresh_entity_stats.go`, `routes/schedule.go`
  - Refs: SDD 01 §5.2, §5.3, §7.4; D-01, D-30; ADR 0007; open-issues 08-1; VF-018
  - Depends on: P1.B.1.1
  - Model: Sonnet
  - Acceptance:
    - `go test ./app/catalog/... -run TestRoleCompatibility` → PASS
    - `go test -tags=integration ./app/catalog/... -run 'TestEntityConstraints|TestEntityStats'` → PASS; stats counts equal a recount on the fixture
  - Notes: engines, edition contributors and episodes are deferred (D-30); add no role values for them. `avg_score` stays null until P1.D.3.2 provides scores.

- [x] **P1.B.1.3** Migrate `editions` with `isbn13_is_valid`, plus `affiliate_links` and `catalog_redirects`.
  - Files/modules: `database/migrations/*_create_editions_table.go`, `*_create_isbn13_function.go`, `*_create_affiliate_links_table.go`, `*_create_catalog_redirects_table.go`
  - Refs: SDD 01 §5.2, §9; D-10; ADR 0012; open-issues 01-1
  - Depends on: P1.B.1.1
  - Model: Sonnet
  - Acceptance:
    - `psql -c "select isbn13_is_valid('9780306406157'), isbn13_is_valid('9780306406158')"` → `t | f`.
    - `go test ./app/catalog/... -run TestISBN` → PASS
  - Notes: language `und` and null publisher are valid (01-1, DQ-10).

#### P1.B.2 Read models

- [x] **P1.B.2.1** Build the read-model assembler with audience filtering (`site`/`api`) driven by `CATALOG_API_SOURCES` and per-field provenance.
  - Files/modules: `app/catalog/readmodel.go`, `app/catalog/audience.go`, `app/catalog/audience_test.go`, `config/catalog.go`
  - Refs: SDD 01 §6.1, §6.3, §7.3, §9; ADR 0006; D-03; A-07; open-issues 01-2, 07-7
  - Depends on: P1.B.1.2, P1.B.1.3
  - Model: Opus
  - Acceptance:
    - `go test ./app/catalog/... -run TestVisibility` → PASS (TMDB/IGDB fixture fields omitted for `api`, present for `site`; omitted, not `null`; `meta.omittedFields` set)
  - Notes: default `CATALOG_API_SOURCES=wikidata,wikipedia,openlibrary,hexascore`; until `docs/legal/open-library.md` exists, `openlibrary` is allowed only for identifiers and titles (the licence check itself is a P6 blocking task).

- [x] **P1.B.2.2** Implement listing queries with the legacy filters and sorts over the signed keyset cursor.
  - Files/modules: `app/catalog/listing.go`, `app/catalog/cursor.go`, `app/catalog/listing_test.go`
  - Refs: SDD 01 §6.2, §7.2, §9; SDD 05 §6.2 "Cursors"; VF-005; `legacy/apps/web/lib/sections.ts`, `legacy/apps/web/lib/categories.ts`
  - Depends on: P1.B.2.1, P0.F.3.1
  - Model: Opus
  - Acceptance:
    - `go test ./app/catalog/... -run 'TestMerge|TestRedirect|TestVisibility|TestCursor'` → PASS (this microtask makes `TestCursor` pass; the others belong to their own microtasks)
    - `go test -tags=integration ./app/catalog/... -run TestListingFilters` → `genre`, `platform`, `year` (1800–2100), `min_score`/`max_score` (0–10 step 0.5), `min_<category>` and every legacy `sort` return the expected fixture slugs
  - Notes: `public` items only (FR-01-14). Score columns land in P1.D.1.1; until then score filters run against fixture values and P1.D.1.1 re-runs this test.

#### P1.B.3 Slugs, redirects and merge

- [x] **P1.B.3.1** Implement slug generation and `catalog_redirects`, answering 301 for old slugs and merged ids.
  - Files/modules: `app/catalog/slugs.go`, `app/catalog/redirects.go`, `app/http/middleware/catalog_redirect.go`, tests
  - Refs: SDD 01 §7.1, §9; D-17
  - Depends on: P1.B.1.3
  - Model: Sonnet
  - Acceptance:
    - `go test ./app/catalog/... -run TestRedirect` → PASS
    - `curl -sI $WEB/pt-br/games/$OLD_SLUG` → `301`, `Location: /pt-br/games/$NEW_SLUG`.
  - Notes: slugs identical across locales (D-17); redirect chains collapse to one hop.

- [x] **P1.B.3.2** Implement the item merge service with a jsonb snapshot of both items in the audit log and a `catalog:merge` console command (service half of `mergeItems`).
  - Files/modules: `app/catalog/merge.go`, `app/catalog/merge_test.go`, `app/console/catalog_merge.go`
  - Refs: SDD 01 §7.5, §9; open-issues 04-9; legacy `actions/admin.ts` (`mergeItems`)
  - Depends on: P1.B.3.1
  - Model: Opus
  - Acceptance:
    - `go test ./app/catalog/... -run TestMerge` → PASS (relations, translations and redirects moved; loser redirects to winner; snapshot row present; transaction rolls back on any failure)
    - `./bin/app catalog:merge --from <id> --into <id>` on the seed → exit 0 and `curl -sI $WEB/games/<loser-slug>` → 301
  - Notes: emits `catalog.merged` with `q.DispatchTx`; ratings, statuses and list items of the losing item are moved by subscribers in P1.D.3.2 and P1.F.1.2. The admin route and form (`/admin/items/{id}`) are wired in P1.G.5.1 once the admin gate exists. Merge is not reversible (04-9).

#### P1.B.4 Data quality, sitemap and seed data

- [x] **P1.B.4.1** Add the `catalog:check` command with the SDD 01 data-quality asserts and a JSON report.
  - Files/modules: `app/console/catalog_check.go`, `app/catalog/dq.go`, `app/catalog/dq_test.go`
  - Refs: SDD 01 §7.6, §9
  - Depends on: P1.B.2.1
  - Model: Sonnet
  - Acceptance:
    - `go test -tags=integration ./app/catalog/... -run TestDQ` → each assert fires on a crafted bad fixture
    - `vukem catalog:check --json` on the seeded database → `{"errors":0}`
  - Notes: the run on the migrated staging copy belongs to P2.F.

- [x] **P1.B.4.2** Stream the sitemap index and per-type child sitemaps with locale alternates.
  - Files/modules: `app/http/controllers/sitemap_controller.go`, `app/catalog/sitemap.go`, `routes/web.go`, tests
  - Refs: SDD 01 §9; SDD 09 §7.7, §9; open-issues 01-8, 09-8; VF-007
  - Depends on: P1.B.2.2
  - Model: Sonnet
  - Acceptance:
    - `curl $WEB/sitemap.xml` → index whose child files each hold ≤ 50,000 URLs with `xhtml:link` alternates for `en`, `pt-br`, `es`, `x-default`.
    - `curl -s $WEB/sitemap.xml` → index; every child ≤ 40,000 `<url>`; no entity with < 3 items.
  - Notes: a 40,000 cap satisfies both SDD criteria. Entities with < 3 linked items stay out and are `noindex,follow` (01-8).

- [x] **P1.B.4.3** Write the `DemoCatalog` seeder and factories used by tests and local development.
  - Files/modules: `database/seeders/demo_catalog.go`, `database/factories/*.go`
  - Refs: SDD 11 §7.7; SDD 01 §10; <https://vukem.dev/docs/database/factories-seeders>
  - Depends on: P1.B.1.3
  - Model: Sonnet
  - Acceptance:
    - `vukem db:seed DemoCatalog` → 300 items across four types with translations in three locales, entities and editions; two runs with `Seed(42)` produce identical slugs
  - Notes: P1.C.1.1 adds 1 admin and 50 users and P1.D.2.1 adds 2,000 ratings to the same seeder (SDD 11 §7.7).

#### P1.B.5 Release

- [x] **P1.B.5.1** Release `v0.2.0` (catalogue model and read paths).
  - Files/modules: `CHANGELOG.md`, tag `v0.2.0`
  - Refs: ADR 0013; SDD 11 §7.1
  - Depends on: P1.B.2.2, P1.B.3.2, P1.B.4.1, P1.B.4.2, P1.B.4.3
  - Model: Sonnet
  - Acceptance:
    - `go test -tags=integration ./app/catalog/...` with `DATABASE_URL` set → PASS (FR-01-1..13).
    - `gh release view v0.2.0` → non-empty body
  - Notes: procedure at the top of Phase 1.

### P1.C Auth and accounts

#### P1.C.1 Scaffold and hashing

- [x] **P1.C.1.1** Install vukem auth with Google and 2FA in the real app and extend `users` with the SDD 03 columns plus `account_events`.
  - Files/modules: output of `vukem install:auth -social google -2fa -locales en,pt-BR,es`, `database/migrations/*_add_profile_columns_to_users_table.go`, `*_create_account_events_table.go`, `app/models/user.go`, `routes/auth.go`, `database/seeders/demo_catalog.go`, `docs/sdd/03-users-auth.md` (names recorded in P0.F.1.1)
  - Refs: SDD 03 §5, §6.1, §9; ADR 0008; D-05; open-issues 03-7, 03-10; <https://vukem.dev/docs/cli/reference-install> "`vukem install:auth`"
  - Depends on: P1.A.5.1
  - Model: Opus
  - Acceptance:
    - `go test ./tests/... -run 'TestSecurity_A0[147]_'` → pass, including vukem-generated `TestSecurity_A07_SignupIsEnumerationSafe`, `TestSecurity_A07_ProvingTheEmailEndsSquatterAccess`
    - `vukem migrate:status --json` → the auth and profile migrations ran
  - Notes: columns per SDD 03 §5 (`username`, `role`, ban fields, `locale`, `profile_private`, `personalization_opt_out`, `password_set_at`, deletion fields). Rename scaffold paths to the legacy paths in the user-owned `routes/auth.go` (SDD 03 §6.1). Seeder gains 1 admin and 50 users.

- [x] **P1.C.1.2** Port the compatibility hasher and enforce the password rules (legacy minimum, maximum 128).
  - Files/modules: `app/accounts/compat_hasher.go`, `app/accounts/compat_hasher_test.go`, `app/accounts/password_rules.go`, `testdata/betterauth_hashes.json`
  - Refs: SDD 03 §7.2, §9; D-24; ADR 0008; legacy `lib/password-rules.ts`
  - Depends on: P1.C.1.1, P0.F.1.3
  - Model: Sonnet
  - Acceptance:
    - `go test ./app/accounts -run TestSecurity_A07_LegacyScryptVerifiesAndRehashes` → Better Auth 1.7.7 fixture hash verifies, stored value then starts `$argon2id$`, wrong password fails
    - `go test ./app/accounts -run TestPasswordRules` → 129 characters rejected, 128 accepted
  - Notes: fixtures from P0.F.1.2. The 2027-12-31 scrypt sunset is a P2 job, not here.

#### P1.C.2 Auth and profile pages

- [x] **P1.C.2.1** Build the localised `Auth/*` Inertia pages on the legacy paths, including onboarding with a nickname (replaces `setNickname`).
  - Files/modules: `resources/js/pages/Auth/{Login,Signup,CheckEmail,VerifyNotice,Recover,Reset,TwoFactorChallenge,Onboarding}.tsx`, `app/http/controllers/auth/onboarding_controller.go`, `tests/auth_pages_test.go`
  - Refs: SDD 03 §6.1, §7.1; SDD 09 §6.2 (`Auth/*` SSR, `noindex`); legacy `app/[locale]/{login,signup,verify-email,recover-password,reset-password,onboarding}/`, `actions/account.ts`
  - Depends on: P1.C.1.2, P1.A.4.1
  - Model: Sonnet
  - Acceptance:
    - `go test ./app/http/... -run 'TestInertiaPages/Auth'` → each page asserts its component with `AssertInertia` in `en`, `pt-BR`, `es`
    - `go test ./tests -run TestOnboardingNickname` → taken nickname → 422 `nickname-taken`; unverified user → redirect to `/verify-email`
  - Notes: `/two-factor-challenge` and `/user/two-factor*` keep the scaffold paths (SDD 03 §6.1). Google buttons use `/auth/google/redirect`.

- [x] **P1.C.2.2** Build the `/profile` sections (details, password, 2FA, sessions, privacy) replacing `updateName`, `setPassword` and `setProfilePrivate`.
  - Files/modules: `resources/js/pages/Profile/Index.tsx`, `app/http/controllers/profile/*.go`, `routes/profile.go`, `tests/profile_test.go`
  - Refs: SDD 03 §6.1, §9; D-25; legacy `app/[locale]/profile/page.tsx`, `actions/account.ts`
  - Depends on: P1.C.2.1
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests -run 'TestProfileDetails|TestProfilePassword|TestProfilePrivacy|TestSessionsRevoke'` → PASS
    - `go test ./tests -run TestSecurity_A07_GoogleOnlyMustSetPasswordBefore2FA` → 2FA routes refuse until `password_set_at` is set; the explanation key exists in three locales
  - Notes: `currentPassword` is required only when `password_set_at` is not NULL. `personalizationOptOut` is stored now and consumed in P6.D. Export and deletion sections link to P1.F.5.1 and P1.C.3.2.

#### P1.C.3 Admin, bans and deletion

- [x] **P1.C.3.1** Install the admin panel, define the admin gate, add the `users` resource with ban/unban (replaces `banUser`) and the admin CLI commands.
  - Files/modules: `vukem install:admin` output, `app/admin/admin.go` (gate), `app/admin/users_resource.go`, `app/console/{admin_grant,admin_revoke,user_reset_2fa}.go`, `app/accounts/ban.go`, `tests/admin_users_test.go`
  - Refs: SDD 03 §6.3, §9; open-issues 02-8, 03-6; VF-010, VF-012; legacy `actions/admin.ts` (`banUser`), `app/[locale]/admin/page.tsx`; <https://vukem.dev/docs/starters> "Add-ons" (`install:admin`)
  - Depends on: P1.C.1.1
  - Model: Opus
  - Acceptance:
    - `go test ./tests -run 'TestSecurity_A07_Ban'` → requests after a ban answer 403 `account-banned`; one `user.banned` event dispatched.
    - `go test ./tests -run TestSecurity_A01_AdminRoleNotEditableInPanel` → crafted `role=admin` POST changes nothing.
  - Notes: `BeforeSave` lacks previous values (VF-010): re-read the row inside the transaction. No impersonation in v1 (VF-012). `/admin` answers 404 to non-admins (legacy `admin.spec.ts`).

- [x] **P1.C.3.2** Implement deletion scheduling and cancellation, the purge job and the public `/help/delete-account` page (replaces `scheduleDeletion`, `cancelDeletion`).
  - Files/modules: `app/accounts/deletion.go`, `jobs/deletion_purge.go`, `routes/schedule.go`, `resources/js/pages/Help/DeleteAccount.tsx`, `tests/deletion_test.go`
  - Refs: SDD 03 §6.1, §7.5, §9; open-issues 10-7, 14-7; legacy `lib/account-deletion.ts`, `lib/account-purge.ts`, `app/api/cron/purge-accounts/`
  - Depends on: P1.C.3.1
  - Model: Opus
  - Acceptance:
    - `go test ./jobs -run TestDeletionPurge` → ratings unchanged, no PII in `users`, Stripe failure keeps the user scheduled.
    - `curl -s $WEB/help/delete-account | grep -c '<h1'` → 1
  - Notes: owner input needed: open-issues 14-7 (public lists of a deleted user) is still `owner`; implement the architect default (anonymise public lists, delete private ones) in one function so the answer is a one-line change. The Stripe cancellation hook is an interface here, implemented in P1.H.3.1.

#### P1.C.4 Access tokens

- [x] **P1.C.4.1** Implement `/app/v1/tokens` (password and Google ID token), rotation, listing and revocation with the six abilities.
  - Files/modules: `app/http/controllers/appapi/tokens_controller.go`, `app/accounts/app_tokens.go`, `routes/app_api.go`, `tests/app_tokens_test.go`
  - Refs: SDD 03 §6.2, §7.4, §9; SDD 05 §6.4; SDD README "Access-token abilities"; A-01; ADR 0009; <https://vukem.dev/docs/security/api-tokens>
  - Depends on: P1.C.3.1
  - Model: Opus
  - Acceptance:
    - `go test ./tests -run TestSecurity_A07_GoogleIDTokenAudience` → token for another client id → 401
    - `go test ./tests -run 'TestAppTokens|TestSecurity_A01_TokenAbilities'` → `hx_app_` prefix, all six abilities, never `*`; a paid `hx_live_` key is refused on `/app/v1`
  - Notes: `/app/v1/me` and `/me/deletion` reuse the P1.C.2.2 and P1.C.3.2 handlers; `/me/export` arrives in P1.F.5.1. Mobile catalogue endpoints are P3.C.

#### P1.C.5 Mail and bot protection

- [x] **P1.C.5.1** Add the localised account mail templates and send queued mail with the locale carried in the job payload.
  - Files/modules: `resources/views/mail/{password_changed,two_factor_changed,deletion_scheduled,account_anonymized}.html`, `app/mail/locale.go`, `lang/*.json` (mail keys), `tests/mail_test.go`
  - Refs: SDD 03 §6.5; open-issues 03-8; VF-013; <https://vukem.dev/docs/digging-deeper/mail>
  - Depends on: P1.C.2.2, P1.C.3.2
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests -run TestMailLocaleFromPayload` → a `pt-BR` user's queued mail renders the Portuguese subject outside a request
  - Notes: `passkeys_retired` belongs to the cutover (P2.G). Copy through the `text-editor` subagent.

- [x] **P1.C.5.2** Protect login, signup, recover, contact and token routes with the proof-of-work challenge, honeypot fields and rate limits.
  - Files/modules: `app/security/pow.go`, `app/security/pow_test.go`, `app/http/middleware/challenge.go`, `resources/js/lib/pow-worker.ts`
  - Refs: SDD 03 §7.3, §9; ADR 0016; open-issues 03-4; VF-003; legacy `lib/bot.ts`, `lib/rate-limit.ts`; <https://vukem.dev/docs/security/rate-limiting>
  - Depends on: P1.C.2.1
  - Model: Opus
  - Acceptance:
    - `go test ./tests -run 'TestSecurity_A07_Pow'` → 428 without a solution past the threshold; reused solution refused.
  - Notes: no third-party challenge (Turnstile) without an owner yes and an ADR amendment.

#### P1.C.6 Release

- [x] **P1.C.6.1** Release `v0.3.0` (auth and accounts).
  - Files/modules: `CHANGELOG.md`, tag `v0.3.0`
  - Refs: ADR 0013; SDD 11 §7.1
  - Depends on: P1.C.4.1, P1.C.5.1, P1.C.5.2
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests/... -run 'TestSecurity_A0[147]_'` → pass
    - `gh release view v0.3.0` → non-empty body
  - Notes: the Better Auth importer (`database/legacyimport`, SDD 03 §9) is P2.F.

#### P1.C.7 Passkeys (D-06 amended 2026-10-29, ADR 0017)

- [x] **P1.C.7.1** Adopt the vukem release with VF-001 and wire passkey registration, sign-in and removal on the server.
  - Files/modules: `go.mod` (vukem version with VF-001), `database/migrations/*_create_passkeys_table.go` (or the scaffold's), `routes/auth.go`, `app/accounts/passkeys.go`, `app/http/middleware/security_headers.go` (`publickey-credentials-get=(self), publickey-credentials-create=(self)`), `app/jobs/deletion_purge.go`, `tests/passkeys_test.go`
  - Refs: SDD 03 FR-03-20 to FR-03-24, §5, §7.8; SDD 13 (Permissions-Policy row); ADR 0017; VF-001; <https://vukem.dev/docs/security/two-factor> (passkey section once published)
  - Depends on: vukem release with VF-001 (owner releases it), P1.C.2.2
  - Model: Opus
  - Acceptance:
    - `go test ./tests -run 'TestSecurity_A07_Passkey|TestSecurity_A01_PasskeyRemoveOnlyOwn'` → PASS: foreign credential refused, replayed or expired challenge refused, counter regression refused, another user's passkey cannot be removed
    - `go test ./tests -run 'TestPasskeyRegisterLimit|TestPasskeyLoginSkipsTOTP|TestPasskeyChangeAuditedAndMailed'` → PASS: 11th passkey 409 `passkey-limit`, passkey login with TOTP enabled lands signed in, `account_events` rows and one `passkey_changed` mail per change
    - `curl -sI http://127.0.0.1:3000/ | grep -i permissions-policy` → contains `publickey-credentials-create=(self)`
  - Notes: no app-local WebAuthn verifier or app-level dependency (ADR 0017); if the vukem API differs from SDD 03 §7.8, amend the SDD first. RP ID `hexascore.app` in production so imported legacy credentials verify. Independent especialista review before merge (auth).

- [x] **P1.C.7.2** Build the passkey UI: profile section and login button in three locales, plus the `passkey_changed` mail.
  - Files/modules: `resources/js/pages/Profile/Index.tsx` (section `passkeys`), `resources/js/pages/Auth/Login.tsx`, `resources/js/lib/webauthn.ts`, `lang/{en,pt-BR,es}.json`, `resources/views/mail/passkey_changed.html`
  - Refs: SDD 03 FR-03-20 to FR-03-24, §6.1, §6.5, §7.8; SDD 09; legacy `app/[locale]/profile/passkeys-section.tsx`, `app/[locale]/login/login-form.tsx`, `messages/*/account.json` `profile.passkeys`, `messages/*/auth.json` (`passkey`, `errors.passkeyFailed`)
  - Depends on: P1.C.7.1
  - Model: Sonnet
  - Acceptance:
    - `bun run lint && bun run typecheck && bun run test` → pass; Vitest covers a cancelled browser prompt showing the "cancelled or failed" message
    - `go test ./tests -run TestPasskeyProfileProps` → `Profile/Index` carries the user's passkeys (id, name, createdAt, lastUsedAt) and never public keys or credential ids
  - Notes: legacy copy reused through the `text-editor` subagent; the remove button keeps the legacy accessible label "Remove passkey {name}".

### P1.D Scoring

#### P1.D.1 Schema and formula

- [x] **P1.D.1.1** Migrate ratings, events, flags, bursts, actions, priors, formula versions, `item_status`, `time_to_beat` and the item score columns.
  - Files/modules: `database/migrations/*_create_ratings_table.go` and siblings for `rating_events`, `rating_flags`, `score_bursts`, `score_actions`, `score_priors`, `score_formula_versions`, `rating_migration_audit`, `item_status`, `time_to_beat`; `*_add_score_columns_to_items_table.go`; `app/models/rating.go`
  - Refs: SDD 02 §5; D-00, D-13; ADR 0005; open-issues 02-5; VF-009
  - Depends on: P1.B.5.1, P1.C.6.1
  - Model: Opus
  - Acceptance:
    - `vukem migrate && vukem migrate:rollback && vukem migrate` → exit 0
    - `go test -tags=integration ./tests/scoring -run TestRatingConstraints` → category values only in steps of 5 within 5–100; generated `overall` correct; `(user_id, item_id)` unique
    - `go test -tags=integration ./app/catalog/... -run TestListingFilters` → still PASS with real score columns
  - Notes: generated column inserts follow the P0.F.5.1 verdict (raw SQL if the ORM writes it). `rating_migration_audit` stays empty until P2.F.

- [x] **P1.D.1.2** Implement formula v1 (Bayesian per category, weights, confidence) as pure Go plus the equivalent SQL recompute.
  - Files/modules: `app/scoring/formula.go`, `app/scoring/formula_test.go`, `app/scoring/recompute.sql.go`
  - Refs: SDD 02 §7.1, §9, §10; ADR 0005; D-13; open-issues 02-2
  - Depends on: P1.D.1.1
  - Model: Opus
  - Acceptance:
    - `go test ./app/scoring/... -run TestFormulaGolden` → passes; the 7.1 example yields `B = 73.4805 ± 1e-4`, stars 7.3, confidence low.
    - `go test ./app/scoring/... -run TestFormulaProperties` (stdlib `testing/quick`, 10,000 cases) → passes: raising any `c_{u,k}` never lowers `B_k` or `B`; adding a rating above `B` raises `B`, below lowers it; `B` lies between `C_t` and the raw weighted mean; raising `w_u` moves `B` toward `overall_u`; `B` in [5, 100].
    - `go test -tags=integration ./tests/scoring -run TestSQLMatchesGo` → SQL recompute equals the Go function to 1e-6 on random fixtures
  - Notes: default `m` per type and priors are measured on the first staging copy (open-issues 02-2, `roadmap`): seed provisional constants in `score_priors` and add a P2.F reminder in the SDD; do not tune here.

#### P1.D.2 Write paths

- [x] **P1.D.2.1** Implement the rating write path on web and `/app/v1` with row locking, the release gate and the 5-rating threshold (replaces `saveRating`, `deleteRating`).
  - Files/modules: `app/scoring/write.go`, `app/http/controllers/ratings_controller.go`, `routes/web.go`, `routes/app_api.go`, `tests/scoring/*_test.go`, `database/seeders/demo_catalog.go` (2,000 ratings)
  - Refs: SDD 02 §6, §7.2, §7.3, §7.4, §9; SDD 09 §6.2 (web mutation routes); A-01; legacy `actions/ratings.ts`, `lib/ratings.ts`
  - Depends on: P1.D.1.2
  - Model: Opus
  - Acceptance:
    - `go test ./tests/scoring -run TestThreshold` → 4 ratings: API `score: null`; 5th: non-null.
    - `go test ./tests/scoring -run TestReleaseGate` → future date 409; year-precision date blocked until 31 Dec; null date allowed; announced item 409.
    - `go test ./tests -run TestSecurity_A01_RatingNeedsAbility` → token without `ratings:write` → 403
  - Notes: per-user limit 30/min via `ratelimit.ByUser` (SDD 02 §6). Events dispatched from `tx.AfterCommit`. Web validation failures answer 303 with `props.errors`.

- [x] **P1.D.2.2** Prove concurrent writers and the nightly recompute never diverge, and measure hot-item recompute latency.
  - Files/modules: `tests/scoring/concurrent_test.go`, `qa/k6/rating-hot-item.js`
  - Refs: SDD 02 §4 (NFR-02-1), §9, §10; open-issues 02-3; <https://vukem.dev/docs/testing/overview> "Testing against PostgreSQL"
  - Depends on: P1.D.2.1
  - Model: Opus
  - Acceptance:
    - `TEST_DATABASE_URL=... go test ./tests/scoring -run TestConcurrentWriters -count=100` (vukem `pgtest.Database`, `https://vukem.dev/docs/testing/overview`) → 50 goroutines upserting/deleting on one item plus a nightly recompute; stored columns equal a from-scratch recompute every run; no deadlock errors.
    - `k6 run qa/k6/rating-hot-item.js` → 50 concurrent raters on one item with 10,000 ratings; recompute p95 recorded; if > 60 ms a covering index migration is added in this microtask
  - Notes: open-issues 02-3 (`roadmap`) is closed here with the measured number written into SDD 02. k6 is an existing legacy QA tool (`legacy/qa/k6`); pin its version.

#### P1.D.3 Anti-manipulation and jobs

- [x] **P1.D.3.1** Implement the burst detector, per-user daily caps and the admin approve/exclude/restore routes.
  - Files/modules: `app/scoring/burst.go`, `app/scoring/caps.go`, `app/http/controllers/admin/scoring_controller.go`, `routes/admin.go`, admin read-only resources for `score_bursts`, `rating_flags`, `score_actions`, `score_formula_versions`
  - Refs: SDD 02 §6 (admin), §7.5, §7.6, §9; VF-008; open-issues 02-4, 02-7
  - Depends on: P1.D.2.1, P1.C.3.1
  - Model: Opus
  - Acceptance:
    - `go test ./tests/scoring -run TestBurst` → seeded 40 inserts in 1 h on an item with `b = 2` opens one burst, holds new-account ratings at 0.25, returns 429 to a 3-day account, admin approve restores the aggregate to the value with all flags removed, exactly.
    - `go test ./tests/scoring -run TestDailyCap` → 21st insert of a 2-day account returns 429 with `Retry-After`.
  - Notes: custom admin routes `/admin/scoring/bursts/{id}/approve|exclude` and `/admin/scoring/flags/{user}/{item}/restore` because row actions are not in vukem (VF-008). Thresholds are untuned until the 90-day review (02-7).

- [x] **P1.D.3.2** Schedule the nightly recompute, burst release and event prune, subscribe to ban/unban and merge events, and add the scoring DQ asserts.
  - Files/modules: `app/jobs/recompute_scores.go`, `app/jobs/recompute_item_score.go`, `app/jobs/release_bursts.go`, `app/jobs/prune_rating_events.go`, `app/scoring/listeners.go`, `routes/schedule.go`, `app/scoring/dq.go`
  - Refs: SDD 02 §6 (jobs, subscriptions), §10; open-issues 02-8; SDD 01 §7.5 (merge); VF-018, VF-021
  - Depends on: P1.D.3.1, P1.B.3.2
  - Model: Sonnet
  - Acceptance:
    - `go test -tags=integration ./tests/scoring -run 'TestBanRecompute|TestMergeMovesRatings|TestNightlyRepairsDrift'` → PASS (batches of 500 item ids ascending; drift count 0 after the nightly job)
    - `vukem schedule:list` → `RecomputeScores` daily 04:00 `America/Sao_Paulo`, `ReleaseBursts` hourly, `PruneRatingEvents` daily 04:30
  - Notes: `schedule:list` per <https://vukem.dev/docs/cli/reference-runtime> "`vukem schedule:run`, `schedule:work`, `schedule:list`" (prints `TASK EXPRESSION NEXT RUN`). Entity `avg_score` (P1.B.1.2 job) now reads real scores.

#### P1.D.4 Status, time-to-beat, widget and help page

- [x] **P1.D.4.1** Implement status and time-to-beat writes on web and `/app/v1` (replaces `saveItemStatus`, `saveTimeToBeat`, `deleteTimeToBeat`).
  - Files/modules: `app/scoring/status.go`, `app/scoring/time_to_beat.go`, `app/http/controllers/status_controller.go`, `routes/web.go`, `routes/app_api.go`, tests
  - Refs: SDD 02 §6, §7.7; legacy `actions/item-status.ts`, `actions/time-to-beat.ts`, `lib/time-to-beat.ts`
  - Depends on: P1.D.2.1
  - Model: Sonnet
  - Acceptance:
    - `go test -tags=integration ./tests/scoring -run 'TestStatus|TestTimeToBeat'` → 0.5 h steps within 0.5–10,000 h; 409 `status-not-completed`; 422 `not-a-game`; `StatusChanged` and `TimeToBeatReported` events dispatched after commit
  - Notes: `GET /app/v1/items/{id}/me` returns `{rating, status, timeToBeat}` (SDD 02 §6).

- [x] **P1.D.4.2** Build the keyboard-operable half-star widget component and the localised `/help/score` page with category keys.
  - Files/modules: `resources/js/components/rating/StarRating.tsx`, `resources/js/components/rating/StarRating.test.tsx`, `resources/js/pages/Help/Score.tsx`, `lang/*.json` (`categories.<type>.<key>`, `rating.valuetext`), `tests/i18n/category_keys_test.go`
  - Refs: SDD 02 §10 (help outline, i18n); SDD 09 §7.4, §9; D-00; A-06; `legacy/apps/web/lib/categories.ts`
  - Depends on: P1.D.1.2, P1.A.3.2
  - Model: Sonnet
  - Acceptance:
    - `bunx vitest run resources/js/components/rating` → passes keyboard, pointer, ARIA cases.
    - `go test ./tests/i18n -run TestCategoryKeys` → every `categories.<type>.<key>` exists in en, pt-BR, es.
  - Notes: `role="slider"`, arrow keys in 0.5 steps, `aria-valuetext` "7.5 of 10 stars". The widget is wired into the item page in P1.G.2.3.

#### P1.D.5 Release

- [x] **P1.D.5.1** Release `v0.4.0` (scoring).
  - Files/modules: `CHANGELOG.md`, tag `v0.4.0`
  - Refs: ADR 0013; SDD 11 §7.1
  - Depends on: P1.D.2.2, P1.D.3.2, P1.D.4.1, P1.D.4.2
  - Model: Sonnet
  - Acceptance:
    - `go test ./app/scoring/... ./tests/scoring/...` → PASS
    - `gh release view v0.4.0` → non-empty body
  - Notes: `go test ./tests/migration -run TestLegacyRounding` (SDD 02 §9) belongs to the importer in P2.F.

### P1.E Search

#### P1.E.1 Search schema

- [x] **P1.E.1.1** Add the extensions, normalisation functions, per-locale text search configurations, the trigger-maintained `search_tsv`, trigram indexes and `search_dirty`.
  - Files/modules: `database/migrations/*_create_search_extensions.go`, `*_create_hs_norm_function.go`, `*_create_text_search_configurations.go`, `*_add_search_columns_to_item_translations_table.go`, `*_create_search_indexes.go`, `*_create_search_dirty_table.go`
  - Refs: SDD 07 §5, §7.9, §9; ADR 0011; D-16; open-issues 07-1, 07-3, 07-9; VF-017
  - Depends on: P1.B.5.1
  - Model: Opus
  - Acceptance:
    - `psql "$DATABASE_URL" -c "select to_tsvector('hs_pt','Ação') = to_tsvector('hs_pt','acao')"` → `t`.
    - `psql ... -c "select hs_norm('Pokémon: Edição Ouro!')"` → `pokemon edicao ouro`.
    - `go test -tags=integration ./internal/search/... -run TestMigrationsUpDown` → PASS
  - Notes: trigger, not a generated column (07-1). On the empty table indexes are created inside the migration; the concurrent script (P1.A.4.3) is for later rebuilds (VF-017 workaround). Minimum Postgres 16.

- [x] **P1.E.1.2** Add the `search_suggest` table, its prefix and trigram indexes and the incremental sync runner driven by `search_dirty`.
  - Files/modules: `database/migrations/*_create_search_suggest_table.go`, `internal/search/suggest_sync.go`, `app/jobs/sync_search_suggest.go`, `routes/schedule.go`
  - Refs: SDD 07 §5, §7.5, §7.10, §10
  - Depends on: P1.E.1.1
  - Model: Sonnet
  - Acceptance:
    - `go test -tags=integration ./internal/search/... -run TestSuggestSyncConcurrentRunners` → two runners leave no duplicate or missing rows
    - `EXPLAIN` of `label_norm ^@ 'zel'` → `Index Scan` on `search_suggest_prefix`
  - Notes: runner honours `ctx` and the VF-018 semaphore.

#### P1.E.2 Searcher and relevance

- [x] **P1.E.2.1** Implement the `Searcher` interface (search, suggest, entities) with tiered ranking, query parsing and `site`/`api` audiences.
  - Files/modules: `internal/search/searcher.go`, `internal/search/query.go`, `internal/search/rank.go`, `internal/search/*_test.go`
  - Refs: SDD 07 §6, §7.1–§7.4, §7.7, §9; open-issues 07-7; ADR 0006
  - Depends on: P1.E.1.2, P1.B.2.1
  - Model: Opus
  - Acceptance:
    - `EXPLAIN` of `search_tsv @@ websearch_to_tsquery('hs_en','zelda')`, `hs_norm('zleda') <% title_norm` and `label_norm ^@ 'zel'` shows `Bitmap Index Scan` / `Index Scan` on `item_translations_tsv_gin`, `item_translations_title_trgm`, `search_suggest_prefix`.
    - `go test ./internal/search/... -run 'TestParse|TestAudience'` → control chars stripped, length limits, ISBN detection, exclusion-only queries handled; `api` audience ranks only allowlisted-source text
  - Notes: engine swap stays behind this interface (D-16). Page beyond 200 results → 400 `invalid-query`.

- [x] **P1.E.2.2** Build the 60-query golden relevance set and make it pass, deciding the `unaccent` question with evidence.
  - Files/modules: `internal/search/testdata/golden.yaml`, `internal/search/testdata/fixture.sql` (3k items), `internal/search/golden_test.go`, `docs/sdd/07-search.md` (decision on 07-2)
  - Refs: SDD 07 §9, §10; open-issues 07-2
  - Depends on: P1.E.2.1
  - Model: Opus
  - Acceptance:
    - `go test -tags=integration ./internal/search/... -run 'TestGolden|TestSearchSQLConfined'` passes.
  - Notes: seeds from SDD 07 §10 (`zleda`, `acao`, `ações`, `o senhor dos aneis`, `senor de los anillos`, `"the last of us" -part`, `gojira`, `dom casmurro`, `machado de assis`, `harry potter`, one ISBN-13). open-issues 07-2 (`roadmap`) closes here: record whether both accented and unaccented forms are indexed. The 3k fixture is drawn from seed data until the migrated catalogue exists (P2.F re-runs it).

#### P1.E.3 Search routes and load

- [x] **P1.E.3.1** Expose `/search`, `/search/suggest` and `POST /search/events` with caching, ETag/304, the legacy suggest redirect and 30-day event purge.
  - Files/modules: `app/http/controllers/search_controller.go`, `routes/web.go`, `internal/search/events.go`, `app/console/search_purge.go`, `tests/search_routes_test.go`
  - Refs: SDD 07 §6, §7.6, §7.8, §9; SDD 09 §6.2 (`Search/Index` SSR `noindex, follow`); open-issues 08-5; legacy `app/api/search/suggest/route.ts`
  - Depends on: P1.E.2.1, P1.A.4.1
  - Model: Sonnet
  - Acceptance:
    - `curl -s 'localhost:8080/search/suggest?q=zel' | jq '.data|length'` ≤ 8; first title contains "Zelda".
    - Anonymous `/search?q=zelda` repeated with `If-None-Match` → 304.
    - An event at `now() - 31 days` is gone after `search:purge`.
  - Notes: the page UI is P1.G.3.1; here the controller renders `Search/Index` props and the tests assert them. `/api/v1/search` is P1.H.2.1. `localhost:8080` in the SDD criterion means `$WEB`.

- [x] **P1.E.3.2** Add the search k6 scenario and record suggest p95 at 2-character prefixes and index sizes.
  - Files/modules: `qa/k6/search.js`, `docs/sdd/07-search.md` (measured numbers)
  - Refs: SDD 07 §4, §9, §10; open-issues 07-8, 07-9; `legacy/qa/k6/api.js` line 58
  - Depends on: P1.E.3.1
  - Model: Sonnet
  - Acceptance:
    - `SEARCH_CACHE=off k6 run qa/k6/search.js` → `p(95)<300` (search, 20 rps) and `p(95)<50` (suggest, 50 rps) pass.
    - `select pg_size_pretty(sum(pg_relation_size(indexrelid))) from pg_index where indexrelid::regclass::text ~ '(tsv|trgm|suggest)'` → recorded for the seed; the ≤ 600 MB check on the full catalogue runs in P2.F
  - Notes: open-issues 07-8 closes here (fallback: 3-character minimum if p95 fails at 2).

#### P1.E.4 Release

- [x] **P1.E.4.1** Release `v0.5.0` (search).
  - Files/modules: `CHANGELOG.md`, tag `v0.5.0`
  - Refs: ADR 0013; SDD 11 §7.1
  - Depends on: P1.E.2.2, P1.E.3.2
  - Model: Sonnet
  - Acceptance:
    - `go test -tags=integration ./internal/search/...` → PASS
    - `gh release view v0.5.0` → non-empty body
  - Notes: procedure at the top of Phase 1.

### P1.F User library

#### P1.F.1 Lists

- [x] **P1.F.1.1** Migrate `lists` and `list_items`, add `ListPolicy` and the list limits, and implement create, update and delete on web and `/app/v1` (replaces `createList`, `updateList`, `deleteList`).
  - Files/modules: `database/migrations/*_create_lists_table.go`, `*_create_list_items_table.go`, `app/policies/list_policy.go` (`vukem make:policy List`), `app/library/lists.go`, `app/http/controllers/lists_controller.go`, `routes/web.go`, `routes/app_api.go`
  - Refs: SDD 14 §5, §6.1, §6.3, §7.1, §9; open-issues 08-2; legacy `actions/lists.ts`, `lib/db/schema/social.ts`; <https://vukem.dev/docs/security/authorization>
  - Depends on: P1.D.5.1
  - Model: Opus
  - Acceptance:
    - `go test -tags=integration ./tests -run 'TestSecurity_A01_ListOwnerOnly|TestListLimits|TestListTypeMismatch'` → pass; non-owner PATCH → 404.
  - Notes: one item type per list (legacy). `View` denies as not-found for private or hidden lists (SDD 14 §6.1). Writes accept `Idempotency-Key` on `/app/v1`.

- [x] **P1.F.1.2** Implement list items add, remove and versioned reorder, the list events and the merge subscriber (replaces `addToList`, `removeFromList`, `moveInList`).
  - Files/modules: `app/library/list_items.go`, `app/library/reorder.go`, `app/library/reorder_test.go`, `app/library/listeners.go`, controllers and routes for `PUT|DELETE /lists/{id}/items/{itemId}`, `PUT /lists/{id}/order`
  - Refs: SDD 14 §6.2, §6.3, §6.5, §7.1, §9, §10; SDD 01 §7.5
  - Depends on: P1.F.1.1, P1.B.3.2
  - Model: Opus
  - Acceptance:
    - `go test ./app/library -run TestReorderProperty` → 10,000 random permutations: positions dense, set unchanged, stale version → 409.
    - `go test -tags=integration ./tests -run 'TestListEventsAfterCommit|TestMergeMovesListItems'` → `list.item_added` exists only when the transaction commits; merged items collapse without duplicates
  - Notes: events are dispatched with `q.DispatchTx` for SDD 08 (P6.C) consumers; no consumer in Phase 1.

#### P1.F.2 Library and profiles

- [x] **P1.F.2.1** Implement the library view with edition selection and the public profile with privacy rules.
  - Files/modules: `database/migrations/*_create_user_item_editions_table.go`, `app/library/library.go`, `app/library/profile.go`, `app/http/controllers/{library,profile_public}_controller.go`, routes for `/library`, `/u/{nickname}`, `PUT|DELETE /items/{id}/edition`, `/app/v1/me/library`, `/app/v1/users/{nickname}`
  - Refs: SDD 14 §5, §6.2, §6.3, §7.2, §9; open-issues 14-5; legacy `app/[locale]/u/[nickname]/page.tsx`, `u/[nickname]/lists/[id]/page.tsx`
  - Depends on: P1.F.1.1
  - Model: Sonnet
  - Acceptance:
    - `curl -s /u/<private>` → 200, body has no counts, `<meta name="robots" content="noindex">`.
    - `go test -tags=integration ./tests -run 'TestLibraryGroups|TestEditionNotOfWork'` → grouped by type with `Scroll` 30/type; foreign edition → 422 `edition-not-of-work`
  - Notes: private profile = 200 with nickname only (14-5 decision, legacy parity). Page components are built in P1.G.3.2; this microtask asserts props with `AssertInertia`.

#### P1.F.3 Reports, contact and notifications

- [x] **P1.F.3.1** Implement the `user_notifications` inbox, `notifications.Send` with dedupe, the unread shared prop and the read routes.
  - Files/modules: `database/migrations/*_create_user_notifications_table.go`, `app/notifications/send.go`, `app/notifications/send_test.go`, `app/http/middleware/unread_notifications.go`, routes `/me/notifications`, `POST /me/notifications/{id}/read`, `POST /me/notifications/read-all`, `/app/v1/me/notifications*`
  - Refs: SDD 14 §5, §6.2, §6.3, §6.5, §7.5, §10; open-issues 04-6, 10-1, 14-1; VF-011; <https://vukem.dev/docs/frontend/inertia> "Shared props"
  - Depends on: P1.F.1.1
  - Model: Sonnet
  - Acceptance:
    - `go test -tags=integration ./tests -run 'TestNotificationDedupeOnRetry|TestSecurity_A01_NotificationsOwnerOnly'` → a retried job writes one row; another user's id → 404
  - Notes: `notifications.Send` is the only writer of `user_notifications` (SDD 14 §6.5). No push in v1.

- [x] **P1.F.3.2** Implement reports and the contact form, their admin resources and resolve routes (replaces `report`, `sendContactMessage`, `setReportStatus`, `setContactHandled`, `setListHidden`).
  - Files/modules: `database/migrations/*_create_reports_table.go`, `*_create_contact_messages_table.go`, `app/library/reports.go`, `app/library/contact.go`, `app/admin/{report,contact_message}_resource.go` (`vukem make:admin-resource`), `app/http/controllers/admin/reports_controller.go`, `routes/admin.go`
  - Refs: SDD 14 §5, §6.3, §6.4, §7.3, §9; SDD 03 §7.3 (honeypot); VF-008; legacy `actions/reports.ts`, `actions/contact.ts`, `actions/admin.ts`, `lib/contact.ts`, `app/[locale]/admin/{reports,contact}/`
  - Depends on: P1.F.3.1, P1.C.5.2
  - Model: Sonnet
  - Acceptance:
    - `go test -tags=integration ./tests -run TestReportResolveNotifies` → one notification per reporter, duplicate open report → 200 and 1 row.
    - `go test ./tests -run 'TestContactHoneypot|TestSecurity_A01_ReportSelf'` → honeypot submission dropped silently; reporting yourself → 403 `report-self`
  - Notes: the resolve notice goes through `notifications.Send` (P1.F.3.1). Contact keeps the legacy topics (`lib/contact.ts`).

#### P1.F.4 Data export

- [x] **P1.F.4.1** Implement the background ZIP export with a signed 24 h link, the 7-day cooldown and the CSV formula guard.
  - Files/modules: `database/migrations/*_create_data_exports_table.go`, `app/library/export.go`, `app/jobs/library_export.go`, routes `POST /profile/export`, `POST /app/v1/me/export`, `resources/views/mail/export_ready.html`
  - Refs: SDD 14 §5, §6.3, §6.5, §7.4, §9; SDD 03 §6.4; open-issues 14-4, 14-6; VF-029
  - Depends on: P1.F.3.1, P1.C.5.1, P0.F.9.1
  - Model: Opus
  - Acceptance:
    - `go test -tags=integration ./tests -run TestExportZip` → ZIP has the 12 files, CSV cells starting `=` are prefixed `'`, second request in 7 days → 429 `export-cooldown`, link after 24 h → 403.
  - Notes: repeatable-read transaction (SDD 03 §6.4); secrets and other users' data excluded. VF-029 workaround as chosen in P0.F.9.1. Storage is RustFS locally, the Railway bucket from P2.A.

#### P1.F.5 Release

- [x] **P1.F.5.1** Release `v0.6.0` (user library).
  - Files/modules: `CHANGELOG.md`, tag `v0.6.0`
  - Refs: ADR 0013; SDD 11 §7.1
  - Depends on: P1.F.1.2, P1.F.2.1, P1.F.3.2, P1.F.4.1
  - Model: Sonnet
  - Acceptance:
    - `vukem openapi:export -prefix /app/v1 | diff - openapi/app-v1.json` → empty after the routes are added.
    - `gh release view v0.6.0` → non-empty body
  - Notes: `k6 run qa/k6/public-lists.js` (SDD 14 §9) runs in P1.I.3.1. 2026-10-29: released without the OpenAPI diff, because the `/app/v1` handlers are not typed yet and `openapi/` does not exist; the `/app/v1` document and its diff check move to P1.H.4.1.

### P1.G Web app parity

#### P1.G.1 Shell and security headers

- [x] **P1.G.1.1** Build the app shell: root view, shared props, layout and navigation, SEO head (canonical, hreflang, Open Graph, JSON-LD), flash, the `Error` page and Base UI primitives.
  - Files/modules: `resources/views/app.html`, `app/http/middleware/share_props.go`, `app/seo/seo.go`, `resources/js/layouts/AppLayout.tsx`, `resources/js/components/{Seo,Nav,Footer,CoverThumb,Dialog}.tsx`, `resources/js/pages/Error.tsx`, `resources/js/ssr.tsx`, `package.json` (Base UI exact version)
  - Refs: SDD 09 §6.1, §7.5, §7.6, §7.7, §9; D-17, D-22; open-issues 09-2; legacy `lib/seo.ts`, `components/`
  - Depends on: P1.A.5.1, P1.A.3.1, P1.C.6.1
  - Model: Opus
  - Acceptance:
    - `curl -s $WEB/games/<slug> | grep -o 'hreflang="[^"]*"' | sort -u` → `en`, `es`, `pt-BR`, `x-default`.
    - `bunx biome ci resources/js packages && bun run typecheck` → 0 errors.
    - `curl -s $WEB/no-such-page` → 404 rendered by `Error` with `<meta name="robots" content="noindex">`
  - Notes: pages are lazy on the client and eager in `ssr.tsx`. No ETag on HTML documents (09-2). Icon set and ICU message helper as named by SDD 09.

- [x] **P1.G.1.2** Add the nonce-based CSP with `strict-dynamic` and the security headers, with `TestSecurity_*` tests and a CSP violation check.
  - Files/modules: `app/http/middleware/security_headers.go`, `app/http/middleware/security_headers_test.go`, `resources/views/app.html` (nonce meta)
  - Refs: SDD 09 §7.8, §9; legacy `lib/security-headers.ts`, `qa/e2e/specs/headers.spec.ts`; <https://vukem.dev/docs/security/security-headers>
  - Depends on: P1.G.1.1
  - Model: Opus
  - Acceptance:
    - `curl -sI $WEB/` → CSP with `'nonce-` and `'strict-dynamic'`, no `unsafe-inline`; `X-Frame-Options: DENY`.
    - `go test ./app/http/middleware -run 'TestSecurity_A05_'` → CSP, HSTS, `Referrer-Policy`, `Permissions-Policy` and `X-Content-Type-Options` asserted on HTML and JSON responses
  - Notes: SDD 13 (security) does not exist yet: name the tests after the OWASP category the behaviour covers (A05 Security Misconfiguration) and record them for SDD 13. The SSR output must carry the same nonce (VF-025 caching is not used in v1).

#### P1.G.2 Catalogue pages

- [x] **P1.G.2.1** Build `Home/Index` with the legacy per-type section highlights.
  - Files/modules: `app/http/controllers/home_controller.go`, `resources/js/pages/Home/Index.tsx`, `resources/js/components/catalog/SectionHighlights.tsx`, `tests/inertia_pages_test.go`
  - Refs: SDD 09 §6.2 (`Home/Index`); D-02; open-issues 08-8; legacy `app/[locale]/page.tsx` (`SectionHighlights` per item type)
  - Depends on: P1.G.1.1, P1.D.5.1
  - Model: Sonnet
  - Acceptance:
    - `go test ./app/http/... -run 'TestInertiaPages/Home'` → component `Home/Index`, `sections` has four entries, anonymous and logged in
  - Notes: the `trending` props of SDD 09 depend on SDD 08 trending (P6.C). Phase 1 keeps legacy parity with section highlights computed by the catalogue read model; the trending prop is added in P6.C. Public SEO props are eager, not `Defer` (08-8).

- [x] **P1.G.2.2** Build `Catalog/Index` listings with filters, facets and `Scroll` infinite pages.
  - Files/modules: `app/http/controllers/catalog_controller.go`, `resources/js/pages/Catalog/Index.tsx`, `resources/js/components/catalog/{Filters,ItemCard}.tsx`
  - Refs: SDD 09 §6.2, §7.3; SDD 01 §6.2; legacy `app/[locale]/[section]/page.tsx`, `lib/page-window.ts`
  - Depends on: P1.G.2.1, P1.B.2.2
  - Model: Sonnet
  - Acceptance:
    - `go test ./app/http/... -run 'TestInertiaPages/Catalog/Index'` → `items` is a `Scroll` prop of 24 with an opaque cursor; `facets` optional
    - `curl -s $WEB/movies?genre=<slug> | grep -c '<h1'` → 1
  - Notes: scroll meta shape as decided in P0.F.3.1.

- [x] **P1.G.2.3** Build `Catalog/Show` and `Catalog/Edition` with the rating widget, status, time-to-beat, add-to-list, report and editions.
  - Files/modules: `resources/js/pages/Catalog/{Show,Edition}.tsx`, `resources/js/components/item/{RatingPanel,StatusPicker,TimeToBeat,AddToList,ReportButton,Editions}.tsx`, `app/http/controllers/item_controller.go`
  - Refs: SDD 09 §6.2, §7.4, §9; SDD 02 §6; SDD 14 §6.2; SDD 01 §6.1 (attribution); legacy `app/[locale]/[section]/[slug]/page.tsx`
  - Depends on: P1.G.2.2, P1.D.4.2, P1.F.1.2, P1.F.3.2
  - Model: Opus
  - Acceptance:
    - `curl -s $WEB/games/<slug> | grep -c '<h1'` → 1; same for `/pt-br/…`, `/es/…`.
    - `go test ./app/http/... -run 'TestInertiaPages/Catalog/Show'` → `item`, `ratingSummary`, `entities`, `editions` (books, first 10 + `editionCount`), `timeToBeat` (games), `me` only when logged in
  - Notes: mutations use the session routes from P1.D and P1.F; optimistic UI must reconcile with the returned aggregate. TMDB/IGDB attribution visible where those fields render (ADR 0006).

- [x] **P1.G.2.4** Build `Entity/Show` for people, companies, franchises, platforms, genres, game modes, countries and languages.
  - Files/modules: `app/http/controllers/entity_controller.go`, `resources/js/pages/Entity/Show.tsx`
  - Refs: SDD 09 §6.2; SDD 01 §6.2, §7.4; D-01; open-issues 01-8
  - Depends on: P1.G.2.2
  - Model: Sonnet
  - Acceptance:
    - `go test ./app/http/... -run 'TestInertiaPages/Entity/Show'` → `entity`, `roles`, `items` (`Scroll`, grouped by type), `counts`, `avgScore`
    - `curl -s $WEB/people/<slug-with-2-items> | grep -c 'noindex,follow'` → 1
  - Notes: new pages without a legacy equivalent (D-01); still part of the Phase 1 scope because the catalogue SDD ships them.

#### P1.G.3 Search, library and account pages

- [x] **P1.G.3.1** Build the `Search/Index` page and the header suggest combobox.
  - Files/modules: `resources/js/pages/Search/Index.tsx`, `resources/js/components/search/{SuggestBox,ResultGroup}.tsx`
  - Refs: SDD 07 §6, §7.7; SDD 09 §6.2, §7.6; legacy `app/[locale]/search/page.tsx`, `qa/e2e/specs/search.spec.ts`
  - Depends on: P1.G.1.1, P1.E.3.1
  - Model: Sonnet
  - Acceptance:
    - `bunx vitest run resources/js/components/search` → combobox keyboard navigation and ARIA (`role="combobox"`, `aria-activedescendant`) pass
    - `go test ./app/http/... -run 'TestInertiaPages/Search'` → grouped results without `type`, `Scroll` results with `type`
  - Notes: partial reload `only: ['results']` when a type is set (SDD 07 §6). Click events go to `POST /search/events`.

- [x] **P1.G.3.2** Build the profile, list, library, lists-management and notifications pages.
  - Files/modules: `resources/js/pages/User/{Profile,List}.tsx`, `resources/js/pages/Library/Index.tsx`, `resources/js/pages/Lists/{Index,Edit}.tsx`, `resources/js/pages/Account/Notifications.tsx`
  - Refs: SDD 14 §6.2; SDD 09 §6.2; legacy `app/[locale]/u/`
  - Depends on: P1.G.1.1, P1.F.2.1, P1.F.3.1
  - Model: Sonnet
  - Acceptance:
    - `go test ./app/http/... -run 'TestInertiaPages/(User|Library|Lists|Account)'` → every row of SDD 14 §6.2 asserts its component and props (owner, other, anonymous)
    - `bun run typecheck` → 0 errors
  - Notes: drag-and-drop reorder must have a keyboard alternative (move up/down buttons) for axe and SDD 09 §7.6.

#### P1.G.4 Content pages and admin

- [x] **P1.G.4.1** Build `Docs/Page` (Markdown rendered and sanitised at build), `Legal/Page`, `Contact/Index`, robots and the legacy catch-all.
  - Files/modules: `resources/docs/*.md`, `scripts/build-docs.ts`, `resources/js/pages/{Docs/Page,Legal/Page,Contact/Index}.tsx`, `app/http/controllers/{docs,legal,contact}_controller.go`, `app/http/controllers/robots_controller.go`
  - Refs: SDD 09 §6.2, §7.7; SDD 05 §6.5 (docs list); legacy `app/[locale]/{docs,policy,terms,contact,[...rest]}/`, `lib/legal.ts`
  - Depends on: P1.G.1.1, P1.F.3.2
  - Model: Sonnet
  - Acceptance:
    - `curl -s $WEB/docs/errors | grep -c '<h1'` → 1 and `curl -s $WEB/terms | grep -c '<h1'` → 1 in three locales
    - `curl -s $WEB/robots.txt` → disallows `/admin`, `/profile`, `/me`, lists the sitemap
  - Notes: docs bodies are English with translated chrome (open-issues 05-7); coverage, webhooks, exports and changes docs pages arrive with P3. Markdown rendering library: none new; use the build step already present or ask the owner. The privacy policy text change for personalisation (D-29) needs the owner's legal review before cutover.

- [x] **P1.G.4.2** Wire the admin area: items list and detail with the merge form and affiliate links, contact and reports queues, and the ingestion status page (replaces `saveAffiliateLink`, `deleteAffiliateLink` and the UI of `mergeItems`).
  - Files/modules: `app/admin/items_resource.go`, `app/admin/affiliate_links_resource.go`, `app/http/controllers/admin/{items,merge,ingestion}_controller.go`, `resources/views/admin/*.html`, `routes/admin.go`
  - Refs: SDD 09 §6.2, open-issues 09-12 (vukem views inside `/admin`, not Inertia); SDD 01 §7.5; VF-008; legacy `app/[locale]/admin/{items,items/[id],ingestion,contact,reports}/`, `actions/admin.ts`
  - Depends on: P1.B.3.2, P1.C.3.1, P1.F.3.2
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests -run 'TestSecurity_A01_AdminPagesNotFoundForMembers|TestAdminMerge|TestAffiliateLinks'` → `/admin*` answers 404 to visitors and members; merge form calls the P1.B.3.2 service; affiliate link save/delete audited
  - Notes: affiliate links render on the site only when the flag is on and never in the API (SDD 01 §6.1); use `vukem install:flags` only with the owner's yes, otherwise an env flag. The ingestion page shows `source_records` counts and last run, which stay empty until P6.A.

#### P1.G.5 Release

- [x] **P1.G.5.1** Release `v0.7.0` (web app parity).
  - Files/modules: `CHANGELOG.md`, tag `v0.7.0`
  - Refs: ADR 0013; SDD 11 §7.1
  - Depends on: P1.G.1.2, P1.G.2.3, P1.G.2.4, P1.G.3.1, P1.G.3.2, P1.G.4.1, P1.G.4.2
  - Model: Sonnet
  - Acceptance:
    - `go test ./app/http/... -run TestInertiaPages` → every row of §6.2 asserts its component and props (anonymous and logged in)
    - `gh release view v0.7.0` → non-empty body
  - Notes: `Developer/ApiKeys`, `Billing/Index` and `Pricing/Index` rows of SDD 09 §6.2 are covered by P1.H; exclude them from this run and include them in P1.H.5.1.

### P1.H API parity

#### P1.H.1 Keys and quota

- [x] **P1.H.1.1** Migrate `api_keys`, implement the fail-closed `hx_live_` key guard and the key management page (replaces `createApiKeyAction`, `revokeApiKeyAction`).
  - Files/modules: `database/migrations/*_create_api_keys_table.go`, `app/api/keys.go`, `app/api/guard.go`, `app/http/controllers/developer/api_keys_controller.go`, `resources/js/pages/Developer/ApiKeys.tsx`, `routes/web.go`
  - Refs: SDD 05 §5, §6.1, §7.1, §9; SDD 09 §6.2 (`/profile/api`); legacy `actions/api-keys.ts`, `app/[locale]/profile/api/page.tsx`; <https://vukem.dev/docs/security/api-tokens>
  - Depends on: P1.G.1.1
  - Model: Opus
  - Acceptance:
    - `go test ./app/api/... -run TestKeys` → pass (sha256 at rest, shown once via flash, revoked key → 401 `unauthorized`)
    - `go test ./tests -run 'TestSecurity_A07_ApiKeyHashedAtRest|TestSecurity_A01_CrossSurfaceCredentials'` → app tokens refused on `/api/v1`, API keys refused on `/app/v1`
  - Notes: legacy keys migrate as hashes in P2.F, so keep the legacy hash format (sha256 hex) exactly.

- [x] **P1.H.1.2** Implement plans, quota metering in Postgres with legacy credit balances, quota and rate-limit headers, and 5xx refunds.
  - Files/modules: `database/migrations/*_create_api_quota_counters_table.go`, `*_create_api_usage_daily_table.go`, `app/api/plans.go`, `app/api/quota.go`, `app/api/quota_test.go`, `app/jobs/flush_api_usage.go`
  - Refs: SDD 05 §5, §6.1, §7.2, §7.4, §9; ADR 0003, 0015; D-23, D-33; open-issues 05-4, 11-9; legacy `app/api/cron/flush-usage/`
  - Depends on: P1.H.1.1
  - Model: Opus
  - Acceptance:
    - `go test ./app/api/... -run 'TestQuota|TestKeys|TestPlans'` → pass; free key requests 1–100 → 200, 101 → 402.
    - `go test -tags=integration ./tests -run 'TestCreditBalanceConsumedFirst|TestRefundOn5xx|TestQuotaFlushCrash'` → PASS
  - Notes: no Redis (ADR 0003): counters in Postgres. Legacy prices and limits unchanged (D-33).

#### P1.H.2 Read endpoints

- [x] **P1.H.2.1** Implement the `/api/v1` read endpoints at legacy parity plus the SDD 01/07 additions, with the licence filter and ETags.
  - Files/modules: `app/http/controllers/api/{items,entities,search,lookup,usage}_controller.go`, `routes/api.go` (typed handlers via `api.Register`), `tests/api_*_test.go`
  - Refs: SDD 05 §6.2, §9; SDD 01 §6.2; SDD 07 §6; open-issues 05-3; ADR 0006; legacy `app/api/v1/`, `lib/api/schemas.ts`
  - Depends on: P1.H.1.2, P1.E.4.1, P1.D.5.1
  - Model: Opus
  - Acceptance:
    - `go test ./tests -run TestAPILicenceFilter` → TMDB/IGDB fixture fields absent from `/api/v1`, present in `/app/v1`.
    - `go test ./tests -run TestAPIContract` → every response passes `AssertMatchesOpenAPI()`.
    - `curl -si -H "Authorization: Bearer $K" -H "If-None-Match: $ETAG" $API/v1/games/<id>` → 304, `X-Quota-Remaining` down by 1
  - Notes: operations: `listItems`, `getItem`, `listGenres` (four per-type operations, see SDD 05 §6.2), `listEditions`, `getEntity`, `listEntityItems`, `getPerson`, `getOrganization`, `getFranchise`, `getPlatform`, `searchItems`, `searchEntities`, `lookupItem`, `getUsage`. `listTrending`, `listChanges`, exports and webhooks are P3.A. The `/app/v1` side of `TestAPILicenceFilter` needs the app read endpoints of P3.C: until then assert against the `site` read model directly.

- [x] **P1.H.2.2** Finish the API surface: host routing in production config, the problem catalogue with `/docs/errors` anchors, CORS, IP limits and the OpenAPI `servers` entry.
  - Files/modules: `app/http/middleware/api_host.go`, `app/http/problems/catalog.go`, `routes/api.go`, `resources/docs/errors.md`
  - Refs: SDD 05 §6.1, §6.2 (problems); SDD README "Problem codes", "Route prefixes"; VF-014, VF-015, VF-016
  - Depends on: P1.H.2.1, P0.F.4.1
  - Model: Sonnet
  - Acceptance:
    - `curl -s -H 'Host: api.hexascore.app' $WEB/v1/games -H "Authorization: Bearer $K" | jq '.data|length'` → > 0
    - `go test ./tests -run TestProblemCatalog` → every code in SDD 05 §6.2 has a type URL that redirects to `/docs/errors#<code>`
  - Notes: `APP_PROBLEM_BASE_URL=https://hexascore.app/problems/`. CORS `*` without credentials.

#### P1.H.3 Billing and pricing

- [x] **P1.H.3.1** Install billing in the real app with `pro`/`business` plans in BRL/USD/EUR, the webhook events, the plan state machine and the checkout pages (replaces `startStripeCheckout`, `openBillingPortal`).
  - Files/modules: `vukem install:billing` output, `config/billing.go`, `app/billing/plan_state.go`, `app/billing/listeners.go`, `resources/js/pages/Billing/Index.tsx`, routes `/billing`, `/profile/api/checkout` (legacy path redirect), `app/accounts/deletion.go` (Stripe cancel hook)
  - Refs: SDD 05 §6.3, §7.5, §9; ADR 0015; D-14; <https://vukem.dev/docs/digging-deeper/billing>; legacy `actions/billing.ts`, `lib/billing/catalog.ts`, `lib/billing/stripe-webhook.ts`, `app/[locale]/profile/api/checkout/page.tsx`
  - Depends on: P1.H.1.2, P0.F.8.1
  - Model: Opus
  - Acceptance:
    - `go test ./tests -run 'TestStripeWebhook'` → duplicate processed once, forged 400, `past_due` → free
    - `go test ./jobs -run TestDeletionPurge` → still passes with the real Stripe hook (Stripe failure keeps the user scheduled)
  - Notes: `CACHE_STORE=database` for processed event ids (billing page). PIX and credit packs stay out (D-06, D-14); the legacy `billing-reminders` cron is not ported. `TestWebhookDelivery` and `TestAppMinVersion` from the same SDD criterion belong to P3.A and P3.C.

- [x] **P1.H.3.2** Build `Pricing/Index` and the API docs pages (quickstart, authentication, errors, pagination, rate limits, coverage, changelog) and serve the Scalar reference.
  - Files/modules: `resources/js/pages/Pricing/Index.tsx`, `resources/docs/{index,authentication,errors,pagination,rate-limits,coverage,changelog}.md`, `routes/docs.go` (`apidocs.Serve` at `/docs/reference`, `vukem.ServeOpenAPIWith` at `/api/v1/openapi.json`)
  - Refs: SDD 05 §6.5; SDD 09 §6.2; ADR 0006; D-03, D-33; open-issues 05-7, 05-11; legacy `app/[locale]/{pricing,docs}/`
  - Depends on: P1.H.3.1, P1.G.4.1
  - Model: Sonnet
  - Acceptance:
    - `curl -s $WEB/pricing | grep -c 'Data coverage'` → 1 (coverage statement per D-03, D-33)
    - `curl -s -o /dev/null -w '%{http_code}' $WEB/docs/reference` → `200` and `curl -s $WEB/api/v1/openapi.json | jq -r .openapi` → `3.1.x`
  - Notes: coverage percentages are `[NOT VERIFIED]` until `qa/sql/api_coverage.sql` runs on the first staging copy (open-issues 05-11, `roadmap`); write the query here and run it in P2.F. Scalar UI stays English (05-7).

#### P1.H.4 Contract and SDKs

- [x] **P1.H.4.1** Commit the golden OpenAPI documents with a diff check, lint them, and regenerate the TypeScript and Python SDKs.
  - Files/modules: `openapi/api-v1.json`, `openapi/app-v1.json`, `scripts/openapi-check.sh`, `sdks/ts/` (from `legacy/sdks/ts`), `sdks/python/` (from `legacy/sdks/python`), `packages/app-client/src/api.ts`
  - Refs: SDD 05 §6.5, §9, §10; ADR 0009; D-06 (one English document); <https://vukem.dev/docs/rest-apis/openapi>; <https://vukem.dev/docs/cli/reference-runtime> "`vukem openapi:export`", "`vukem codegen:client`"
  - Depends on: P1.H.2.2, P1.H.3.1
  - Model: Sonnet
  - Acceptance:
    - `vukem openapi:export -prefix /api/v1 | diff - openapi/api-v1.json` → empty (same for `/app/v1`); Redocly lint (pinned) → 0 errors.
    - `cd sdks/ts && bun run sdk:generate && bun test`; `cd sdks/python && uv run pytest -q` → pass
  - Notes: owner input needed: Redocly CLI is not in D-22; ask, and if refused rely on Schemathesis plus the golden diff. SDK publishing is P3.B; here only regeneration and tests.

#### P1.H.5 Release

- [x] **P1.H.5.1** Release `v0.8.0` (API parity).
  - Files/modules: `CHANGELOG.md`, tag `v0.8.0`
  - Refs: ADR 0013; SDD 11 §7.1
  - Depends on: P1.H.3.2, P1.H.4.1
  - Model: Sonnet
  - Acceptance:
    - `go test ./app/http/... -run 'TestInertiaPages/(Developer|Billing|Pricing)'` → PASS
    - `gh release view v0.8.0` → non-empty body
  - Notes: procedure at the top of Phase 1.

### P1.I QA parity gate

The gate runs against the built container (`docker compose --profile app up`, seeded with `DemoCatalog`) until staging exists (P2.A); the same commands are re-run against staging before cutover (D-21). Legacy configuration is the starting point: `legacy/qa/e2e/playwright.config.ts`, `global-setup.ts`, `helpers/`, `legacy/qa/k6/`, `legacy/qa/schemathesis/`, `legacy/qa/zap/`.

#### P1.I.1 Playwright harness

- [x] **P1.I.1.1** Set up Playwright with axe against the built container, three locales, mobile and desktop projects, and seeded fixtures.
  - Files/modules: `qa/e2e/playwright.config.ts`, `qa/e2e/global-setup.ts`, `qa/e2e/helpers/{auth,axe,db}.ts`, `package.json` (`e2e` script, exact `@playwright/test` and `@axe-core/playwright` versions), `docker-compose.yml` profile `app`
  - Refs: SDD 09 §9, §10; D-21, D-22; legacy `qa/e2e/`
  - Depends on: P1.H.5.1
  - Model: Sonnet
  - Acceptance:
    - `bun run e2e -- --list` → projects `desktop` and `mobile` × locales `en`, `pt-BR`, `es`
    - a smoke spec opening `/` in each project → pass with axe 0 serious/critical
  - Notes: test users are created through the app (signup + mail log) or the seeder, never by writing hashes by hand. 2026-10-29: pulled forward before P1.H.5.1 (billing waits on owner question Q-H31); billing and pricing pages excluded until then.

#### P1.I.2 Ported and new specs

- [x] **P1.I.2.1** Port `auth`, `profile`, `api-keys` and `pricing` specs.
  - Files/modules: `qa/e2e/specs/{auth,profile,api-keys,pricing,passkeys}.spec.ts`
  - Refs: SDD 03 §7.8, §9; SDD 05 §10 (E2E); SDD 09 §10; ADR 0017; legacy `qa/e2e/specs/{auth,profile,api-keys,pricing}.spec.ts`
  - Depends on: P1.I.1.1
  - Model: Sonnet
  - Acceptance:
    - `bunx playwright test qa/e2e/specs/auth.spec.ts qa/e2e/specs/profile.spec.ts` → signup, verify, TOTP login, reset, delete and cancel; axe clean in three locales.
    - `bunx playwright test qa/e2e/specs/api-keys.spec.ts qa/e2e/specs/pricing.spec.ts` → key shown once, checkout redirect (fake Stripe), axe 0 serious
    - `bunx playwright test qa/e2e/specs/passkeys.spec.ts` (after P1.C.7.2) → with a CDP virtual authenticator (`WebAuthn.enable`, `WebAuthn.addVirtualAuthenticator`, ctap2, internal, resident keys, user verification): add a passkey in the profile, sign out, sign in with it, remove it; axe 0 serious
  - Notes: passkeys are back in scope (D-06 amended 2026-10-29, ADR 0017). The legacy suite has no passkey spec (checked 2026-10-29: no `passkey`/`webauthn` in `legacy/qa/e2e/specs`), so `passkeys.spec.ts` is new, not a port; it lands with or after P1.C.7.2 and does not block the other four specs.

- [x] **P1.I.2.2** Port `catalog`, `item`, `ratings`, `time-to-beat` and `search` specs.
  - Files/modules: `qa/e2e/specs/{catalog,item,ratings,time-to-beat,search}.spec.ts`
  - Refs: SDD 02 §10; SDD 07 §10; SDD 09 §10; SDD 14 §10 ("status shelves", "curated list"); legacy specs of the same names
  - Depends on: P1.I.1.1
  - Model: Sonnet
  - Acceptance:
    - `bunx playwright test qa/e2e/specs/{catalog,item,ratings,time-to-beat,search}.spec.ts` → pass in three locales; axe 0 serious/critical
  - Notes: legacy ratings used a 0–100 slider; the port drives the star widget by keyboard (D-00). Critic-score assertions are dropped (D-09).

- [x] **P1.I.2.3** Port `social`, `contact`, `admin`, `docs`, `public-pages`, `headers` and `not-found` specs.
  - Files/modules: `qa/e2e/specs/{social,contact,admin,docs,public-pages,headers,not-found}.spec.ts`
  - Refs: SDD 14 §9, §10; SDD 09 §10; legacy specs of the same names
  - Depends on: P1.I.1.1
  - Model: Sonnet
  - Acceptance:
    - `bunx playwright test specs/social.spec.ts specs/lists.spec.ts specs/contact.spec.ts specs/notifications.spec.ts` → pass, axe 0 serious. (`lists` and `notifications` come from P1.I.2.4; run the full line after it.)
    - `bunx playwright test qa/e2e/specs/{admin,docs,public-pages,headers,not-found}.spec.ts` → pass
  - Notes: `social.spec.ts` keeps "public profile lists ratings; a private one hides them from everyone but the owner" and drops "follow a member and see their ratings in /feed" and "/feed requires login" (D-06). `admin.spec.ts` keeps "/admin answers 404 to visitors and regular members" and "a member reports a user; an admin resolves it in /admin/reports".

- [x] **P1.I.2.4** Write the new specs `i18n`, `seo`, `csp`, `keyboard`, `ssr-nojs`, `lists` and `notifications`.
  - Files/modules: `qa/e2e/specs/{i18n,seo,csp,keyboard,ssr-nojs,lists,notifications}.spec.ts`
  - Refs: SDD 09 §9, §10; SDD 14 §9
  - Depends on: P1.I.2.3
  - Model: Sonnet
  - Acceptance:
    - `bunx playwright test` (qa/e2e) → all specs pass in 3 locales; axe 0 serious/critical; 0 CSP violations.
  - Notes: `ssr-nojs` runs with JavaScript disabled and asserts content on every public page type. `suggestions.spec.ts` (SDD 09 §10) belongs to P4.A.

#### P1.I.3 Load

- [x] **P1.I.3.1** Port and extend the k6 scenarios (API, search, public lists, rating hot item, page TTFB) and record results.
  - Files/modules: `qa/k6/{api-v1,search,public-lists,rating-hot-item,pages-ttfb}.js`, `docs/qa/k6-baseline.md`
  - Refs: SDD 05 §9; SDD 07 §9; SDD 14 §9; SDD 09 §4 (TTFB p75 ≤ 400 ms); open-issues 09-3; VF-025; legacy `qa/k6/`
  - Depends on: P1.I.1.1
  - Model: Sonnet
  - Acceptance:
    - `k6 run qa/k6/api-v1.js` → list `p(95)<300`, detail `p(95)<150` at 50 rps, `http_req_failed<0.001`.
    - `k6 run qa/k6/public-lists.js` → p95 ≤ 400 ms at 30 rps, errors < 0.5 %.
    - `k6 run qa/k6/pages-ttfb.js` → TTFB p75 ≤ 400 ms on home, listing, item and entity pages
  - Notes: open-issues 09-3 closes here: if TTFB fails, the VF-025 workaround (anonymous HTML cache with nonce re-injection) becomes a new microtask, not a silent change.

#### P1.I.4 Contract and security scans

- [x] **P1.I.4.1** Run Schemathesis on both OpenAPI documents and the OWASP ZAP baseline scan, with zero failures and zero High alerts.
  - Files/modules: `qa/schemathesis/run.sh`, `qa/zap/baseline.sh`, `qa/zap/rules.tsv`, `docs/qa/scans.md`
  - Refs: SDD 05 §9, §10; SDD 01 §9; SDD 03 §10; SDD 14 §10; SDD 12; legacy `qa/schemathesis/`, `qa/zap/`
  - Depends on: P1.I.1.1
  - Model: Sonnet
  - Acceptance:
    - `schemathesis run openapi/api-v1.json --url http://localhost:3000 -H "Authorization: Bearer $HX_TEST_KEY" --checks all` (pinned) → 0 failures
    - Schemathesis on `listItems`, `getItem`, `getEntity`, `listEntityItems`, `listEditions` → no failures
    - `qa/zap/baseline.sh $WEB` (ZAP image pinned by digest) → 0 High; Medium findings triaged in `docs/qa/scans.md`
  - Notes: Schemathesis and ZAP are legacy QA tools already in use (pinned there); keep their versions. `rules.tsv` may only lower noise with a written reason, never ignore a High.

#### P1.I.5 Lighthouse

- [x] **P1.I.5.1** Run Lighthouse CI with the SDD 09 budgets on the five reference pages.
  - Files/modules: `lighthouserc.json`, `qa/lighthouse/urls.txt`
  - Refs: SDD 09 §4, §9; ADR 0004; D-22; P0.F.10.1 findings
  - Depends on: P1.I.1.1
  - Model: Sonnet
  - Acceptance:
    - `bunx lhci autorun` → assertions of §4 pass for home, `/games`, a game, a book, a person (mobile).
  - Notes: a failing LCP on item pages triggers the 300 px cover decision (open-issues 09-9), recorded as a follow-up, not fixed here unless trivial.

#### P1.I.6 Parity checklist

- [x] **P1.I.6.1** Write `docs/parity-checklist.md` mapping every legacy route, Server Action and e2e test to its replacement, test and status.
  - Files/modules: `docs/parity-checklist.md`, `scripts/parity-inventory.sh` (lists legacy routes, actions and test titles)
  - Refs: ROADMAP phase map (P1 exit criteria); D-06, D-09, D-21; `docs/PROGRESS.md` (33 Server Actions); legacy `apps/web/app/[locale]/`, `apps/web/actions/`, `qa/e2e/specs/`
  - Depends on: P1.I.2.4, P1.I.3.1, P1.I.4.1, P1.I.5.1
  - Model: Opus
  - Acceptance:
    - `scripts/parity-inventory.sh | wc -l` equals the number of rows in the checklist tables, and `rg -c '\| (done|deferred D-0[69]) \|' docs/parity-checklist.md` equals the same number (100 %)
    - every `deferred` row names its decision (D-06 or D-09) and every `done` row names a test (`Test…` or a spec title)
  - Notes: the checklist is the D-21 cutover gate; it is re-signed on staging in P2.G. Any row that cannot be `done` or `deferred` blocks the release and goes to the owner.

#### P1.I.7 Release

- [x] **P1.I.7.1** Release `v0.9.0` (QA parity gate green on the local container). Done 2026-10-30: `gh release view v0.9.0` published; isolated Playwright run `660 passed (5.9m)` (evidence in `developer-manual-todos.md`).
  - Files/modules: `CHANGELOG.md`, tag `v0.9.0`, `docs/PROGRESS.md`, `ROADMAP.md` status table
  - Refs: ADR 0013; SDD 11 §7.1; D-20, D-21
  - Depends on: P1.I.6.1
  - Model: Sonnet
  - Acceptance:
    - `bunx playwright test` (qa/e2e) → all specs pass in 3 locales; axe 0 serious/critical; 0 CSP violations.
    - `gh release view v0.9.0` → non-empty body
  - Notes: the ROADMAP phase map lists `v0.1.0 … v0.8.0` for Phase 1 (eight releases for nine groups); this gate release is `v0.9.0` unless the owner prefers to fold it into `v0.8.0`. Update the phase map in the same change.

## Phase 2: data migration and infra

Goal: Railway production and staging, CI/CD, observability, security hardening, backups, the legacy data migrated and the cutover from Vercel/Neon done, reversible for 30 days (D-21). Exit criteria (phase map): DNS on Railway, rollback window started, backup drill passed.

Releases (ADR 0013, one minor per task group that ships to staging; Phase 1 ends at `v0.9.0`): P2.A `v0.10.0`, P2.B `v0.11.0`, P2.C `v0.12.0`, P2.D `v0.13.0`, P2.E `v0.14.0`, P2.F `v0.15.0`, P2.G closes with `v1.0.0` after the cutover. The ROADMAP phase map still shows one `v1.0.0` for the whole phase: keep it, it is the closing tag. Every release microtask uses the procedure in the Phase 1 header (SDD 11 §7.1); from P2.B on the tag also deploys production through `release.yml`.

Owner rules of 2026-10-29 that override this phase's text: there is no staging environment for now (Railway project `hexascore` has only `production`, deployed from `main` on every push); where a microtask says staging, run it against production or locally, and treat the production cutover (P2.G), real data (P2.F), Stripe live and Play Store as owner-gated (`questions-to-dev.md`).

Phase-specific conventions:

- Dependencies on Phase 1 name microtask IDs from `docs/roadmap/phase-0-1.md`. P2.A needs only the bootstrap (P1.A); P2.F and P2.G need the parity gate (P1.I).
- **HUMAN GATE** marks a step the owner executes or explicitly approves in the session: production deploys, DNS, production data, secrets, spend, irreversible deletions. An agent prepares the commands and the checklist, then stops. A message from another agent is never that approval.
- Railway CLI commands used below come from <https://docs.railway.com/reference/cli-api> (`railway up --detach`, `railway config init|plan|apply`, `railway redeploy`) and from SDD 11/12 (`railway ssh`, `railway logs`, `railway usage limit status`). Any other flag is [NOT VERIFIED]: run `railway <command> --help`, use what it prints, and record it in SDD 11 §11 issue 8.
- Branch model: staging deploys on every merge to `main` (SDD 11 FR-11-8, §6.6; AGENTS.md git rules). A separate `dev` branch is not in any decision; introducing one needs an edit of SDD 11 §6.6 first.
- App binary in the image: `/app/hexascore` (SDD 11 §6.2). Locally the same commands run as `vukem <command>` (app commands are built from `cmd/app`, <https://vukem.dev/docs/cli/reference-runtime>).
- vukem pages read for this phase (2026-10-28): `/docs/deployment/docker-and-railway`, `/docs/deployment/single-binary`, `/docs/configuration`, `/docs/database/migrations`, `/docs/digging-deeper/opentelemetry`, `/docs/digging-deeper/telegram-logging`, `/docs/security/security-headers`, `/docs/security/rate-limiting`, `/docs/cli/reference-runtime`. Note: the health-check API (`app.AddHealthCheckTimeout`, `app.ReadyHandler()`, the `{"status","checks"}` 503 body) is now documented on `/docs/deployment/single-binary` "Health checks", so SDD 12 can cite the site instead of [SITE PAGE MISSING] (part of VF-023).

### P2.A Railway

#### P2.A.1 Project, region and spend controls

- [ ] **P2.A.1.1** Measure latency from Brazil to the candidate Railway regions and record the region choice.
  - Files/modules: `docs/adrs/0002-railway-over-vercel.md` (addendum "Region measurement"), `scripts/infra/region-latency.sh`
  - Refs: SDD 11 §6.1 (Region, Measurement); ADR 0002; A-03; <https://docs.railway.com/reference/deployment-regions>
  - Depends on: P1.A.5.1
  - Model: Sonnet
  - Acceptance:
    - `scripts/infra/region-latency.sh https://<east>.up.railway.app/up https://<west>.up.railway.app/up` → 50 samples per target of `curl -w '%{time_connect} %{time_starttransfer}'`, p50/p95 printed per region
    - ADR 0002 addendum lists p50/p95 from the owner's line and 3 Brazilian probes, and names the chosen region (`us-east4-eqdc4a` unless the numbers say otherwise)
  - Notes: throwaway services run the starter image only (`/up`), in a scratch Railway project, and are deleted the same day (cost). HUMAN GATE: creating the scratch project uses the owner's Railway account.

- [ ] **P2.A.1.2** Create the Railway project on the Pro plan with environments `production` and `staging`, the `postgres` service per environment, and the spend controls.
  - Files/modules: `docs/sdd/11-infra-deploy.md` (§6.1 plan line, §11 issues 6 and 13 notes), `docs/sdd/12-observability.md` (§11 issue 4 closed: Pro, 30-day logs, monitors available)
  - Refs: SDD 11 §5, §6.1, §6.5, FR-11-1, NFR-11-1; SDD 12 §7.5 A15, §11 issues 4 and 9; ADR 0002, 0003; D-08, D-27; <https://docs.railway.com/reference/pricing/plans>; <https://docs.railway.com/guides/postgresql>; <https://docs.railway.com/pricing/cost-control>
  - Depends on: P2.A.1.1
  - Model: Sonnet
  - Acceptance:
    - `railway status` (linked to the project) → project `hexascore`, environments `production` and `staging`, service `postgres` in each, region as recorded in P2.A.1.1
    - `railway usage limit status` → hard limit US$30; the Railway usage page shows the "Custom email alert" at US$24 (SDD 12 §9 last item)
    - the `postgres` service has no TCP proxy and a volume sized for §5 (6 GB)
  - Notes: HUMAN GATE (spend): the owner upgrades to Pro, enters payment, and sets the alert and the hard limit. Postgres tuning (`shared_buffers=256MB`, `work_mem=8MB`, `maintenance_work_mem=256MB`, `max_connections=60`) is applied in P2.A.4.3 once `ALTER SYSTEM` is verified. Do not add Redis or a worker service (ADR 0003, D-08).

#### P2.A.2 Buckets and domains

- [ ] **P2.A.2.1** Create one private bucket per environment, store its credentials as service variables and upload the storage health sentinel.
  - Files/modules: `.env.example` (`S3_*`), `docs/sdd/11-infra-deploy.md` §6.4 (exact credential names from the Credentials tab), `app/storage/health.go` (sentinel key `health/sentinel.txt`)
  - Refs: SDD 11 §5 (bucket prefixes), §6.1, §6.4; SDD 12 §7.1 (`storage` check); ADR 0014 (amended); <https://docs.railway.com/guides/storage-buckets>; <https://vukem.dev/docs/digging-deeper/storage>
  - Depends on: P2.A.1.2
  - Model: Sonnet
  - Acceptance:
    - Railway shows buckets `hexascore-staging` and `hexascore-production` in the app region, private
    - `go test ./app/storage -run TestStorageHealthSentinel` against RustFS (`docker compose`) → PASS; the same check reports `storage: ok` on staging in P2.A.4.3
  - Notes: buckets are private and have no lifecycle rules, versioning or server-side encryption (11-1, decided): never rely on any of them. HUMAN GATE (secrets): the owner copies the credentials into Railway variables; they never enter git or chat logs.

- [ ] **P2.A.2.2** Confirm or move the `hexascore.app` zone to Cloudflare and point the staging hosts at Railway.
  - Files/modules: `docs/runbooks/dns.md` (records, TTLs, who changes them), `docs/runbooks/dns-legacy.txt` (export of the current Vercel-facing records, used by rollback R1)
  - Refs: SDD 11 §6.1 (Domains), §7.6 (R1), §11 issue 10; SDD 12 §7.7; D-28; ADR 0002
  - Depends on: P2.A.1.2
  - Model: Sonnet
  - Acceptance:
    - `dig +short NS hexascore.app` → Cloudflare name servers
    - `dig +short CNAME staging.hexascore.app` and `dig +short CNAME api.staging.hexascore.app` → the Railway domains of the staging `app` service; Cloudflare proxy off (grey cloud)
    - `docs/runbooks/dns-legacy.txt` lists every current record of apex, `www`, `api`, `dev` with values and TTLs
  - Notes: HUMAN GATE (DNS): a zone move changes name servers at the registrar and must not touch the live apex records; copy them first and compare after. Production hosts (apex with CNAME flattening, `www`, `api`) are switched only in P2.G.3.3. `status.hexascore.app` is added in P2.C.4.1.

#### P2.A.3 Container

- [x] **P2.A.3.1** Turn the bootstrap Dockerfile into the production image: pinned digests, `build:ssr` stage, Node runtime for SSR, `postgresql18-client` and `cwebp`, non-root user.
  - Files/modules: `Dockerfile`, `.dockerignore`, `.buildtags`, `scripts/infra/check-image.sh`, `docs/sdd/11-infra-deploy.md` §6.3 (apk versions found)
  - Refs: SDD 11 §6.3, §9 items 1–2, FR-11-2, NFR-11-3, NFR-11-5; ADR 0004; VF-027; <https://vukem.dev/docs/deployment/docker-and-railway> "What the Dockerfile does"; <https://vukem.dev/docs/deployment/single-binary>
  - Depends on: P1.A.4.2
  - Model: Opus
  - Acceptance:
    - `docker build --build-arg VERSION=v0.0.0-test -t hx .` → exit 0; `docker image inspect hx --format '{{.Config.User}}'` → `app`
    - `grep -nE ':latest|^FROM [^@]+$' Dockerfile` → no output (every `FROM` has a tag and `@sha256:`)
    - `scripts/infra/check-image.sh hx` → compressed size ≤ 400 MB, `pg_dump --version` 18.x and `cwebp -version` run as `app`, `/app/storage` and `/tmp` the only writable paths
  - Notes: the starter uses `oven/bun:1.4.3-alpine` for the assets stage with `--pm bun`; SDD 11 §6.3 specifies `node:24.22.0-alpine3.24` plus a pinned Bun binary. Pick one, record it in §6.3, keep `bun install --frozen-lockfile`. `build:ssr` is a permanent step (VF-027, 09-7). `postgresql18-client` and `libwebp-tools` availability on Alpine 3.24 is [NOT VERIFIED]: if a pinned apk version is missing, record the alternative in §6.3 instead of unpinning. Extra `-X` flags go through plain `go build`, not `vukem build` ([NOT IN VUKEM DOCS]).

- [x] **P2.A.3.2** Harden `serve` for production: signal forwarding to the SSR child, shutdown budget, memory limits and the four readiness checks.
  - Files/modules: `app/ssr/supervisor.go`, `bootstrap/app.go` (health checks), `app/scheduler/heartbeat.go`, `tests/ready_test.go`, `.env.example` (`SHUTDOWN_TIMEOUT=30s`, `GOMEMLIMIT=900MiB`, `NODE_OPTIONS=--max-old-space-size=384`, `QUEUE_NAMES=default,ingest`, `QUEUE_CONCURRENCY=3`, `SCHEDULE_IN_SERVE=true`)
  - Refs: SDD 11 §6.3 (Process, Health, Shutdown), §6.4; SDD 12 FR-12-5, §7.1; VF-002, VF-018, VF-027; <https://vukem.dev/docs/deployment/single-binary> "Health checks"; <https://vukem.dev/docs/cli/reference-runtime> "`vukem ssr:start` and `vukem ssr:check`"
  - Depends on: P2.A.3.1, P2.A.2.1
  - Model: Opus
  - Acceptance:
    - `go test ./tests -run TestReady_ChecksPresent` → PASS: `/up?checks=1` returns `db`, `ssr`, `storage`, `scheduler`, each with a 2 s timeout
    - `docker run` the image, `docker stop -t 40` it during a 10 s request → request completes, logs show the SSR child stopped after the HTTP server, exit 0
    - killing the Node child inside the container → restarted after 1 s (`ssr:start` behaviour), `ssr` check back to `ok` within 5 s
  - Notes: `app.AddHealthCheckTimeout` and the `ok`/`fail` body are documented on `/docs/deployment/single-binary`. Queue priority across `QUEUE_NAMES` is undocumented (VF-018, 11-7): add a test that records the observed order, do not depend on it. The scheduler check uses `s.Func("heartbeat").EveryMinute()`, age < 180 s.

#### P2.A.4 Config as code, secrets and first deploy

- [ ] **P2.A.4.1** Commit `.railway/railway.ts` for both environments and the full variable table, and set the service settings IaC cannot express.
  - Files/modules: `.railway/railway.ts`, `.env.example` (every key of SDD 11 §6.4, empty secrets), `docs/sdd/11-infra-deploy.md` §6.2, §6.4, §11 issue 8
  - Refs: SDD 11 §6.2, §6.4, FR-11-3; VF-026; <https://vukem.dev/docs/deployment/docker-and-railway> "Config as Code", "Start commands replace the entrypoint", "Trusted proxies on Railway"; <https://docs.railway.com/reference/config-as-code>; <https://docs.railway.com/reference/cli-api> "Infrastructure as Code"
  - Depends on: P2.A.3.2, P2.A.2.1
  - Model: Sonnet
  - Acceptance:
    - `railway config plan` → only the expected services (`app`, `postgres`) and variables; no secret value printed (secrets are `preserve()`)
    - `railway config apply` → exit 0; the `app` service shows healthcheck `/up?checks=1` and pre-deploy `/app/hexascore hexascore:predeploy`
    - `grep -c 'preserve()' .railway/railway.ts` ≥ the number of secret rows in §6.4
  - Notes: `railway.json`/`railway.toml` stop being read on 2026-12-01: never add them. Overlap 20 s, draining 35 s, restart `ALWAYS` max 5 are set in service settings until their IaC keys are verified ([NOT VERIFIED], 11-8). Commands are the binary path plus arguments, no shell. Until P2.E.1.1 lands, `hexascore:predeploy` is the `migrate:locked` command from P1.A.4.3 under its final name. `OTEL_METRICS_EXPORTER=none` stays until P2.C.2.1 points metrics at Grafana.

- [ ] **P2.A.4.2** Generate and store the per-environment secrets and keep the owner's offline copy.
  - Files/modules: `docs/runbooks/secrets.md` (inventory: name, environment, provider, rotation, where the offline copy lives; never values)
  - Refs: SDD 11 §5 (Railway variables not backed up), §6.4, NFR-11-5; SDD 13 §7.4; SDD 12 §7.8 R9; <https://vukem.dev/docs/configuration> "APP_KEY"; <https://vukem.dev/docs/deployment/docker-and-railway> "Quick start" (`docker run --rm blog key:generate`)
  - Depends on: P2.A.4.1
  - Model: Sonnet
  - Acceptance:
    - `railway ssh` into staging `app`, then `/app/hexascore about` → `APP_ENV=production`, trusted proxies listed, no secret value printed
    - `go test ./... -run TestSecurity_A02_DockerIgnoresSecrets` → PASS; `git grep -nE 'base64:[A-Za-z0-9+/]{40}|sk_live_|re_[A-Za-z0-9]{20}'` → no output
  - Notes: HUMAN GATE (secrets): the owner generates `APP_KEY` per environment (`docker run --rm hx key:generate`), `BACKUP_KEY` (32 bytes, separate from `APP_KEY`), Stripe test keys for staging, Resend, and keeps an encrypted export of the variables offline (SDD 11 §5). Telegram and OTLP secrets come in P2.C. Production Stripe live keys are set only in P2.G.3.1.

- [ ] **P2.A.4.3** Deploy `main` to staging through Railway's GitHub autodeploy, then verify the Postgres facts and apply the tuning.
  - Files/modules: `docs/sdd/11-infra-deploy.md` (§5 version and extensions, §9 item "SHOW server_version", §11 issue 6 closed), `docs/sdd/07-search.md` (§11 issue 9), `docs/sdd/open-issues.md` (rows 07-9, 11-6 closed)
  - Refs: SDD 11 §5, §6.1, §9, §11 issue 6; SDD 07 §11 issue 9; <https://vukem.dev/docs/deployment/docker-and-railway> "Deploy on Railway", "Trusted proxies on Railway"; <https://docs.railway.com/reference/healthchecks>; <https://docs.railway.com/reference/private-networking>
  - Depends on: P2.A.4.2, P2.A.2.2
  - Model: Sonnet
  - Acceptance:
    - `curl -fsS https://staging.hexascore.app/up?checks=1` → 200 with `db`, `ssr`, `storage`, `scheduler` `ok`
    - `railway ssh` then `psql "$DATABASE_URL" -c 'SHOW server_version' -c 'SELECT extname FROM pg_extension'` → version ≥ 16 (18 expected), `pg_trgm` and `unaccent` present after migrations
    - `railway logs` → request lines show `peer` in `100.64.0.x` and `ip` with the real client address
  - Notes: if Railway rejects a healthcheck path with a query string ([NOT VERIFIED]), mount `app.ReadyHandler()` at `/ready` (single-binary page) and change §6.2. If `ALTER SYSTEM` is refused, record the parameters that could not be set and their defaults. The private hostname is recorded, never exposed.

- [ ] **P2.A.4.4** Prove a staging deploy drops no requests and a failing migration aborts the deploy.
  - Files/modules: `qa/k6/deploy-switch.js`, `docs/qa/deploy-switch.md`, `docs/sdd/11-infra-deploy.md` §11 issue 8 (overlap/draining values observed)
  - Refs: SDD 11 §7.2, §9 items 4–5, FR-11-4, NFR-11-2; <https://docs.railway.com/guides/pre-deploy-command> ("the deployment will not proceed")
  - Depends on: P2.A.4.3
  - Model: Sonnet
  - Acceptance:
    - `k6 run -e BASE=https://staging.hexascore.app qa/k6/deploy-switch.js` during `railway redeploy` → 0 non-2xx at 20 rps
    - a test branch with a migration that raises an error, deployed to staging → deployment failed, previous deployment still serving, `migrate:status --json` unchanged
  - Notes: revert the failing migration with a new commit; never force-push. If requests drop, tune overlap/draining before anything else depends on zero-downtime deploys.

#### P2.A.5 Release

- [x] **P2.A.5.1** Release `v0.10.0` (Railway staging) and deploy every earlier `v0.x` tag's content as the staging baseline. Done 2026-10-30 as `v0.10.0` on GitHub; staging does not exist (owner, 2026-10-29), production is the single environment, so the staging URL check is void.
  - Files/modules: `CHANGELOG.md`, tag `v0.10.0`, `docs/PROGRESS.md`, `ROADMAP.md` status table, `AGENTS.md` commands table (`railway config plan|apply`, `scripts/infra/check-image.sh`)
  - Refs: ADR 0013; SDD 11 §7.1; D-20; Phase 1 header (staging deploys of Phase 1 tags wait for P2.A)
  - Depends on: P2.A.4.4, P2.A.1.2
  - Model: Sonnet
  - Acceptance:
    - `gh release view v0.10.0` → non-empty body
    - `curl -fsS https://staging.hexascore.app/up?checks=1` → 200, all checks `ok`
  - Notes: production has a service and an empty database but no traffic and no custom domain until P2.G. The production deploy pipeline arrives in P2.B; this tag deploys staging only.

### P2.B CI/CD and releases

#### P2.B.1 Pull request pipeline

- [x] **P2.B.1.1** Add the `lint` and `vuln` jobs: golangci-lint with format diff, `go vet`, Biome, `tsc`, markdownlint, commitlint, PR title, and pinned `govulncheck`.
  - Files/modules: `.github/workflows/ci.yml`, `.golangci.yml`, `.markdownlint-cli2.jsonc`, `commitlint.config.ts`, `go.mod` (`tool golang.org/x/vuln/cmd/govulncheck`), `docs/sdd/11-infra-deploy.md` §6.6 (`ESLint` → Biome), `docs/sdd/13-security.md` §11 issue 3 closed
  - Refs: SDD 11 §6.6 (`lint`, `vuln`); SDD 13 §7.7, §7.8, NFR-13-5, §11 issue 3; D-22, D-26, D-37; ADR 0013
  - Depends on: P2.A.5.1
  - Model: Sonnet
  - Acceptance:
    - `gh workflow run ci.yml` on a branch → `lint` and `vuln` green; a PR titled `update stuff` → `lint` red on the title check
    - `go tool govulncheck ./...` → exit 0 locally and in CI; `vuln` also runs on a weekly `schedule`
    - `grep -nE 'uses: [^@]+@v?[0-9]' .github/workflows/*.yml` → no output (every action pinned by full SHA with a version comment)
  - Notes: D-37 approves markdownlint-cli2, commitlint and Redocly; versions exact. The lefthook `pre-push` mirror in SDD 11 §6.6 is not in D-22/D-37: keep `scripts/ci-local.sh` as a plain script called from a git hook and ask the owner before adding lefthook.

- [x] **P2.B.1.2** Add the `test` job: Go unit and integration tests on a Postgres 18.6 service, TypeScript unit tests, and the Windows setup check.
  - Files/modules: `.github/workflows/ci.yml` (`test`, `setup-windows`), `setups/setup-windows.ps1`, `setups/setup-unix.sh`
  - Refs: SDD 11 §6.6 (`test`), §7.7, FR-11-9, §9 last item; <https://vukem.dev/docs/testing/overview> "Testing against PostgreSQL"
  - Depends on: P2.B.1.1
  - Model: Sonnet
  - Acceptance:
    - CI `test` on `ubuntu-24.04` with service `postgres:18.6-alpine3.24` → `go test ./...` and `go test -tags=integration ./...` green, `bun run test` green
    - CI `setup-windows` on `windows-2025` → `setups\setup-windows.ps1` exit 0 and `go test ./...` (unit, no Postgres) green
  - Notes: GitHub service containers run only on Linux runners; the Windows job proves setup and unit tests, not integration. Image tag exactly as SDD 11 (digest pinned).

- [x] **P2.B.1.3** Add the `build` job: image build with the version argument, size and pin checks, and the image artifact reused by later jobs.
  - Files/modules: `.github/workflows/ci.yml` (`build`), `scripts/infra/check-image.sh`
  - Refs: SDD 11 §6.3, §6.6 (`build`), §9 items 1–2, NFR-11-3
  - Depends on: P2.B.1.1
  - Model: Sonnet
  - Acceptance:
    - CI `build` → `docker build --build-arg VERSION=$GITHUB_SHA` exit 0, `scripts/infra/check-image.sh` passes (≤ 400 MB, user `app`, no unpinned `FROM`), artifact `image.tar` uploaded
    - the Vite build (`bun x vite build`) and `bun run build:ssr` run inside the image build, never on the runner
  - Notes: use the Docker layer cache of the Actions cache backend only if it needs no new action outside the pinned set; otherwise accept slower builds.

- [x] **P2.B.1.4** Add the `openapi` job: export diff, Redocly lint, and Schemathesis against the built container.
  - Files/modules: `.github/workflows/ci.yml` (`openapi`), `redocly.yaml`, `qa/schemathesis/run.sh` (from P1.I.4.1)
  - Refs: SDD 11 §6.6 (`openapi`); SDD 05 §6.5, §10; D-37; <https://vukem.dev/docs/cli/reference-runtime> (`openapi:export`)
  - Depends on: P2.B.1.3
  - Model: Sonnet
  - Acceptance:
    - CI: `vukem openapi:export && git diff --exit-code openapi/` → exit 0; a PR that changes a handler without regenerating → red
    - CI: `bunx redocly lint openapi/api-v1.json openapi/app-v1.json` → 0 errors; Schemathesis `--checks all` against the job's container → 0 failures
  - Notes: reuse the image artifact from `build`; never rebuild inside this job. Schemathesis keeps the version pinned in legacy QA.

- [x] **P2.B.1.5** Add the `e2e` and `lighthouse` jobs against the built container.
  - Files/modules: `.github/workflows/ci.yml` (`e2e`, `lighthouse`), `docker-compose.yml` profile `app`, `qa/e2e/playwright.config.ts`, `lighthouserc.json`
  - Refs: SDD 11 §6.6 (`e2e`, `lighthouse`); SDD 09 §4, §9; ADR 0004; P1.I.1.1, P1.I.5.1
  - Depends on: P2.B.1.3
  - Model: Sonnet
  - Acceptance:
    - CI `e2e` → `bun run e2e` green in 3 locales × desktop/mobile, axe 0 serious/critical; traces uploaded on failure
    - CI `lighthouse` → `bunx lhci autorun` assertions of SDD 09 §4 pass on the five reference pages
  - Notes: the container is loaded from `image.tar` and seeded with `DemoCatalog`; tests never hit staging from PRs.

#### P2.B.2 Deploys and version

- [ ] **P2.B.2.1** Gate staging on CI, add the post-deploy `smoke` job and make the PR checks required on `main`.
  - Files/modules: `.github/workflows/staging-smoke.yml`, `qa/k6/smoke.js`, `docs/runbooks/release.md` (section "Staging")
  - Refs: SDD 11 §6.6 (`smoke`, "Staging: Railway GitHub autodeploy of `main`"), §7.1 step 1, FR-11-8; <https://vukem.dev/docs/deployment/docker-and-railway> "Deploy on Railway"
  - Depends on: P2.B.1.2, P2.B.1.4, P2.B.1.5
  - Model: Sonnet
  - Acceptance:
    - a merge to `main` → Railway deploys staging after CI is green, then `staging-smoke` runs `k6 run qa/k6/smoke.js` → `p(95)<500`, 0 errors
    - `gh api repos/AlexGalhardo/hexascore-app/branches/main/protection` → required checks `lint`, `test`, `vuln`, `openapi`, `build`, `e2e`, `lighthouse`
  - Notes: Railway's "wait for CI" option is [NOT VERIFIED] (11-8); if it does not exist, turn autodeploy off and deploy staging from the workflow with `railway up --detach` after the checks. HUMAN GATE: branch protection is a repository setting the owner applies or approves.

- [ ] **P2.B.2.2** Write `release.yml`: tag `v*` and `workflow_dispatch(version)` deploy production with notes taken from `CHANGELOG.md`, behind a GitHub environment approval.
  - Files/modules: `.github/workflows/release.yml`, `Taskfile.yml` (`release`), `docs/runbooks/release.md` (release, missed run, rollback by `-f version=vPREV`)
  - Refs: SDD 11 §6.6 (`release`), §7.1, §8 ("Tag pushed without a run"); ADR 0013; D-20, D-37; <https://docs.railway.com/reference/cli-api> "Deployment" (`railway up --detach`)
  - Depends on: P2.B.2.1
  - Model: Opus
  - Acceptance:
    - `gh workflow run release.yml -f version=v0.10.0` → job waits for the `production` environment reviewer, then checks out the tag and runs `railway up --detach`; `gh run list --workflow release.yml` → success
    - a tag whose `CHANGELOG.md` section is missing → the notes step (literal `awk` heading match) fails before deploying
    - `git-cliff --config .config/cliff.toml --unreleased --tag v0.11.0` → preview grouped by type; `task release -- v0.11.0 --dry-run` (or the task's documented preview) writes nothing
  - Notes: HUMAN GATE (production deploy): the GitHub `production` environment has the owner as required reviewer. The project token secret name and `railway up` flags beyond `--detach` are [NOT VERIFIED]: check `railway up --help`. Never move a pushed tag; one tag per push.

- [x] **P2.B.2.3** Inject the release version into the binary and expose it on every response, in `about` and on `/up`.
  - Files/modules: `cmd/app/main.go` (`version`, `commit`), `app/http/middleware/version.go`, `app/http/middleware/version_test.go`, `Dockerfile` (`VERSION`, `COMMIT` build args)
  - Refs: SDD 11 §6.6 (Version), §7.1 step 5, §9 item 8; SDD 12 §7.8; VF-027; <https://vukem.dev/docs/deployment/single-binary>
  - Depends on: P2.B.1.3
  - Model: Sonnet
  - Acceptance:
    - `go test ./app/http/middleware -run TestVersionHeader` → PASS, header `X-Hexascore-Version` on HTML, JSON and `/up` responses
    - after a staging deploy: `curl -sI https://staging.hexascore.app/up | grep -i x-hexascore-version` → the deployed tag or commit
  - Notes: a `version` field inside the `/up` body is [NOT IN VUKEM DOCS] (VF-027): use the header on `/up` plus `/app/hexascore about`; add the body field only if vukem documents a hook.

#### P2.B.3 Release

- [x] **P2.B.3.1** Release `v0.11.0` (CI/CD) and run the first production deploy through `release.yml`. Done 2026-10-31: `gh run list --workflow release.yml` → `Release v0.11.0 success`; GitHub release published from PR #56.
  - Files/modules: `CHANGELOG.md`, tag `v0.11.0`, `docs/PROGRESS.md`, `AGENTS.md` (commands: `task release`, `gh workflow run release.yml`)
  - Refs: ADR 0013; SDD 11 §7.1; D-20
  - Depends on: P2.B.2.2, P2.B.2.3
  - Model: Sonnet
  - Acceptance:
    - `gh release view v0.11.0` → non-empty body; `gh run list --workflow release.yml` → success for `v0.11.0`
    - `railway ssh` into production `app`, `curl -sI localhost:$PORT/up` → 200 with `X-Hexascore-Version: v0.11.0`
  - Notes: HUMAN GATE (production deploy). Production still has no custom domain and an empty database; it proves the pipeline only.

### P2.C Observability

#### P2.C.1 Logs and Telegram

- [x] **P2.C.1.1** Replace the built-in request logger with the privacy-preserving JSON logger: hashed user id, truncated IP, trace id, sampling, redaction.
  - Files/modules: `app/http/middleware/request_log.go`, `app/http/middleware/request_log_test.go`, `app/logging/redact.go`, `bootstrap/app.go` (`WithLogger`), `docs/sdd/12-observability.md` §6.1
  - Refs: SDD 12 FR-12-1, FR-12-8, §6.1, §11 issue 2; SDD 13 NFR-13-4; VF-022; <https://vukem.dev/docs/digging-deeper/opentelemetry> (`otel.LogHandler`); <https://vukem.dev/docs/configuration> (`LOG_FORMAT`, `LOG_DEDUP_WINDOW`)
  - Depends on: P2.A.5.1
  - Model: Opus
  - Acceptance:
    - `go test ./... -run 'TestLog_RequestLineShape|TestSecurity_A09_HexaLogsRedacted'` → PASS (password, bearer token, API key, email and search text absent; IPv4 /24, IPv6 /48)
    - `railway logs` on staging piped through `scripts/ops/check-logs.sh` → every line parses as JSON with `level`, `request_id`, `route`; 0 email matches
  - Notes: VF-022 workaround. SDD 12 derives the hash key as `app.DeriveKey("log", 32)` and SDD 13 NFR-13-4 as `app.DeriveKey("logid", 32)`: pick one, fix the other SDD in the same change. Responses also carry `X-Trace-ID`; making request id equal trace id stays open (issue 2). Output to stdout only (Railway maps stderr to error).

- [ ] **P2.C.1.2** Wire the Telegram channel (page chat and silent error chat), dedup, and the daily error-fingerprint digest.
  - Files/modules: `bootstrap/app.go` (handler chain), `app/errors/fingerprint.go`, `app/errors/digest.go`, migration `create_error_digest_daily`, `app/errors/digest_test.go`, `.env.example` (`TELEGRAM_*`)
  - Refs: SDD 12 FR-12-6, §5, §7.2, §7.5 (Channels); <https://vukem.dev/docs/digging-deeper/telegram-logging> "Quick start"; <https://vukem.dev/docs/configuration> (`LOG_DEDUP_WINDOW`)
  - Depends on: P2.C.1.1
  - Model: Sonnet
  - Acceptance:
    - `go test ./app/errors -run 'TestFingerprint|TestDigestFakeClock'` → PASS: top 20 fingerprints at 08:00 America/Sao_Paulo, one request id each, prune > 90 days
    - on staging, a forced test error → one message in the error chat (silent); a `LevelFatal` test record → the page chat rings
  - Notes: HUMAN GATE (secrets): the owner creates the bot and the two chats and sets `TELEGRAM_BOT_TOKEN` and the chat ids. `TELEGRAM_MIN_LEVEL=error`, `TELEGRAM_RATE_LIMIT=20`. Digest uses `app.Telegram.SendSync`. Buffered upserts every 60 s (VF-021 pattern), never per error.

#### P2.C.2 Traces and metrics

- [ ] **P2.C.2.1** Create the Grafana Cloud stack and send traces and metrics there over OTLP, with Railway's receiver as the trace-only fallback.
  - Files/modules: `bootstrap/app.go` (`otel.New` first in `WithProviders`, `Config.Global`), `.railway/railway.ts` (`OTEL_*`), `docs/sdd/12-observability.md` §6.3, §11 issue 8 (OTLP URL, Telegram contact point, cert checks verified)
  - Refs: SDD 12 §6.2, §6.3, §7.3, §11 issues 1 and 8; SDD 11 §6.4 (`OTEL_*`); D-35; <https://vukem.dev/docs/digging-deeper/opentelemetry> "Configuration", "Outbound HTTP calls", "Security"; <https://vukem.dev/docs/deployment/docker-and-railway> "OpenTelemetry on Railway"; <https://grafana.com/pricing/>
  - Depends on: P2.C.1.1
  - Model: Sonnet
  - Acceptance:
    - `curl -sI https://staging.hexascore.app/` → `X-Request-ID` and `X-Trace-ID`; that trace id opens in Grafana Tempo
    - Grafana shows `http.server.request.duration` from staging within 5 min of a deploy
  - Notes: HUMAN GATE (account, secrets): the owner creates the free stack and sets `OTEL_EXPORTER_OTLP_ENDPOINT` and `OTEL_EXPORTER_OTLP_HEADERS` (`config.Secret`). Once metrics go to Grafana, `OTEL_METRICS_EXPORTER=none` is removed; keep it whenever the endpoint is Railway's receiver (any value other than `otlp` or `none` stops the boot). Sampler `parentbased_traceidratio`, 0.1 production, 1.0 staging. No `/metrics` endpoint (D-35).

- [x] **P2.C.2.2** Register the platform metrics of the SDD 12 catalogue through the global meter provider, with label caps.
  - Files/modules: `app/metrics/metrics.go`, `app/metrics/labels.go` (caps → `other`), `app/metrics/sql_gauges.go`, `app/scheduler/instrument.go` (`hexa.schedule.runs` wrapper), `tests/metrics_test.go`
  - Refs: SDD 12 FR-12-2, FR-12-3, §6.2, NFR-12-2, §11 issues 1, 7, 10; SDD 02, 05, 07 metric rows (verify registered); VF-019, VF-021; <https://vukem.dev/docs/digging-deeper/opentelemetry> "Metrics", "Configuration"
  - Depends on: P2.C.2.1
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests -run TestMetrics_EveryRouteGroupEmitsLatency` → PASS (manual reader, one histogram point per route group)
    - `go test ./app/metrics -run 'TestLabelCap|TestSQLGauges'` → PASS; `hexa.errors` caps at 50 `error.type` values then `other`
  - Notes: VF-019 workaround: `otel.GetMeterProvider().Meter("hexascore")` after `Config.Global`. Platform metrics only (`hexa.*`, `backup.*` placeholders, `biz.*`); domain metrics owned by Phase 1 groups are checked for presence, not re-implemented; P4/P6 metrics are not added here. Names exactly as §6.2 (no underscore variants).

#### P2.C.3 Alerts and dashboards

- [ ] **P2.C.3.1** Write alert rules A3–A16 with `promtool` fixtures and load them into Grafana alerting.
  - Files/modules: `ops/alerts/rules.yaml`, `ops/alerts/tests.yaml`, `.github/workflows/ci.yml` (`alerts` step with a digest-pinned `prom/prometheus` image), `docs/sdd/12-observability.md` §10 (ruler import result)
  - Refs: SDD 12 FR-12-4, §7.5, §9 item 2, §10; SDD 02 (A13), 04 (A9), 06 (A7), 08 (A14) as inactive rules until their metrics exist
  - Depends on: P2.C.2.2
  - Model: Sonnet
  - Acceptance:
    - `docker run --rm -v "$PWD/ops/alerts:/a" --entrypoint promtool prom/prometheus@sha256:<pinned> test rules /a/tests.yaml` → `SUCCESS` for A3–A16, each firing and resolving
    - Grafana alerting lists the A3–A16 rules with severity labels `page`/`warn`
  - Notes: Grafana ruler import of Prometheus-format rules is [NOT VERIFIED]; if it fails, recreate the rules in Grafana from the same YAML and keep the YAML as the tested source. A15 lives in Railway (P2.A.1.2); A1/A2 are synthetic (P2.C.3.2); A17 is P2.C.3.4.

- [ ] **P2.C.3.2** Configure the synthetic probes, the Telegram and e-mail contact points and the quiet-hours policy.
  - Files/modules: `ops/grafana/synthetics.md` (checks, probes, intervals), `ops/grafana/notification-policy.md`, `docs/sdd/12-observability.md` §7.4, §11 issue 8
  - Refs: SDD 12 §7.4, §7.5 (A1, A2, Channels, quiet hours 23:00–08:00), NFR-12-6
  - Depends on: P2.C.3.1
  - Model: Sonnet
  - Acceptance:
    - Grafana Synthetic Monitoring shows `/up` every 60 s (probe A), `/up` every 120 s (probe B), `/up?checks=1` every 300 s (probe A) against staging; monthly executions ≈ 73,440
    - A1 drill: stopping staging `app` for 3 min → page in Telegram within 2 min; restart → resolved
  - Notes: Telegram as a Grafana contact point and certificate-expiry checks are [NOT VERIFIED]: record the result; fall back to e-mail if Telegram is not offered. Production probes are switched on in P2.G.3.3.

- [ ] **P2.C.3.3** Implement the app-side alerts: SSR fallback page (A5), job failed twice (A6), quota flush errors (A10).
  - Files/modules: `bootstrap/app.go` (`SSRConfig.OnFailure`), `app/alerts/jobs.go` (`q.AddObserver`), `app/api/quota/flush.go`, `tests/alerts_test.go`
  - Refs: SDD 12 FR-12-3, §7.2, §7.5 (A5, A6, A10), §8, §9 item 4; D-07; VF-002; <https://vukem.dev/docs/frontend/inertia> "Server-side rendering"; <https://vukem.dev/docs/digging-deeper/telegram-logging>
  - Depends on: P2.C.1.2, P2.C.2.2
  - Model: Sonnet
  - Acceptance:
    - `go test -tags=integration ./tests -run 'TestAlert_SSRFallbackPublic|TestAlert_JobFailedTwice|TestAlert_QuotaFlush'` → PASS; a logged-in route fallback is counted with `public="false"` and does not page
    - on staging, `railway ssh` + `pkill node` → `/up?checks=1` 503 with `ssr: fail` and a Telegram page within 5 min
  - Notes: the page uses `telegram.LevelFatal` routed to the page chat; dedup keeps a crash loop to one message per window.

- [ ] **P2.C.3.4** Build the A17 projected-spend alert as a daily GitHub Actions job with a calendar fallback.
  - Files/modules: `.github/workflows/projected-spend.yml`, `scripts/ops/projected-spend.ts`, `scripts/ops/projected-spend.test.ts`, `docs/runbooks/cost.md`
  - Refs: SDD 12 FR-12-10, §7.5 A17, §7.9; SDD 11 NFR-11-1, §6.5; D-27; <https://docs.railway.com/pricing/cost-control>
  - Depends on: P2.C.1.2
  - Model: Sonnet
  - Acceptance:
    - `bun test scripts/ops/projected-spend.test.ts` → PASS: day 20 with US$19 month-to-date in a 30-day cycle → projected US$28.5 → alert; US$17 → no alert
    - `gh workflow run projected-spend.yml` → reads usage and posts to Telegram only above US$27
  - Notes: programmatic access to Railway usage is [NOT VERIFIED]. If no documented API exists, the job only posts the day-20 reminder and the owner reads the usage page; record which path was taken in SDD 12 §7.5. Bun runs the script with no new dependency.

- [ ] **P2.C.3.5** Commit the dashboards and check the series budget and telemetry overhead on staging.
  - Files/modules: `ops/grafana/dashboards/{service-health,api-customers,scoring-integrity}.json`, `docs/qa/telemetry-overhead.md`
  - Refs: SDD 12 §7.6, NFR-12-2, NFR-12-3, §9 item 6, §11 issue 7; VF-019
  - Depends on: P2.C.3.1
  - Model: Sonnet
  - Acceptance:
    - after `k6 run qa/k6/api-v1.js` against staging → Grafana shows ≤ 6,000 active series
    - k6 with tracing on vs off → ≤ 3 % CPU and ≤ 40 MB RSS difference, recorded in `docs/qa/telemetry-overhead.md`
  - Notes: Ingestion and Moderation dashboards are added by P6.A and P4.B. Above 6,000 series the trigger of issue 7 (metric views, [NOT IN VUKEM DOCS]) fires: drop labels in app code and file the VF update.

#### P2.C.4 Status page and runbooks

- [ ] **P2.C.4.1** Publish the hand-edited status page on GitHub Pages at `status.hexascore.app`.
  - Files/modules: new repository `AlexGalhardo/hexascore-status` (`index.html`, incident template), `ops/status/README.md` in this repo (how to post an incident), `docs/runbooks/dns.md`
  - Refs: SDD 12 §7.7, §11 issue 5; SDD 11 §7.5 step 1; ADR 0014 (24 h RPO statement); D-28
  - Depends on: P2.A.2.2
  - Model: Sonnet
  - Acceptance:
    - `curl -fsS https://status.hexascore.app/ | grep -c 'RPO'` → ≥ 1
    - `dig +short CNAME status.hexascore.app` → the GitHub Pages host; HTTPS certificate valid
  - Notes: HUMAN GATE (DNS, new repository): the owner creates the repository and the Cloudflare record. Static HTML only, no build step, no third-party script.

- [ ] **P2.C.4.2** Commit runbooks R1–R9 and the monthly cost checklist under `docs/runbooks/`, and rehearse R1 and R7 on staging.
  - Files/modules: `docs/runbooks/{ssr-crash-loop,pg-connections,ingestion-stuck,quota-drift,suggestion-backlog,disk-full,rollback,restore,leaked-secret}.md`, `docs/runbooks/cost.md`, `docs/runbooks/README.md` (index with alert → runbook map)
  - Refs: SDD 12 §7.8, §7.9; SDD 11 §7.5; SDD 13 §7.9; <https://vukem.dev/docs/cli/reference-runtime> "`vukem down` and `vukem up`"; <https://docs.railway.com/guides/roll-back-bad-deploy>
  - Depends on: P2.C.3.3
  - Model: Sonnet
  - Acceptance:
    - every alert A1–A17 in `ops/alerts/rules.yaml` and SDD 12 §7.5 links a runbook: `scripts/ops/check-runbook-links.sh` → 0 missing
    - R7 on staging: Railway rollback to the previous deployment → `/up?checks=1` 200 within 5 min, timing recorded in `rollback.md`
  - Notes: R3, R5 refer to systems built in P6/P4; write them now with a "applies from P4/P6" line. R8 links the restore command delivered by P2.E.3.1. Git Bash on Windows needs `MSYS_NO_PATHCONV=1` for `railway ssh` paths ([SITE PAGE MISSING], SDD 11 §7.5).

#### P2.C.5 Release

- [ ] **P2.C.5.1** Release `v0.12.0` (observability).
  - Files/modules: `CHANGELOG.md`, tag `v0.12.0`, `docs/PROGRESS.md`, `docs/sdd/open-issues.md` (rows 12-4, 12-8 closed), `vukem-feedback.md` (VF-019, VF-022, VF-023 notes)
  - Refs: ADR 0013; SDD 12 §9; SDD 11 §7.1
  - Depends on: P2.C.3.2, P2.C.3.4, P2.C.3.5, P2.C.4.1, P2.C.4.2
  - Model: Sonnet
  - Acceptance:
    - `go test ./... -run 'TestMetrics_|TestLog_|TestReady_|TestSecurity_A09_Hexa'` → PASS
    - `gh release view v0.12.0` → non-empty body; production deploy through `release.yml` green
  - Notes: HUMAN GATE (production deploy).

### P2.D Security hardening

#### P2.D.1 Headers and CSP

- [x] **P2.D.1.1** Configure the web and API header sets of SDD 13 §6.2 and the CSP violation endpoint, with tests and staging checks.
  - Files/modules: `bootstrap/app.go` (`MiddlewareOptions.SecureHeaders`, `APIMiddlewareOptions.SecureHeaders`), `app/http/csp_reports.go`, `tests/security_headers_test.go`, `qa/e2e/specs/{headers,csp}.spec.ts`
  - Refs: SDD 13 §6.1, §6.2, §7.1 (`/csp-reports` 30/min, 16 KiB), §9 items 1–2; SDD 09 §7.8; <https://vukem.dev/docs/security/security-headers> "Configuration", "What CSPSources refuses", "Collecting violation reports", "Common mistakes"
  - Depends on: P2.C.5.1, P1.G.1.2
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests -run 'TestSecurity_A02_HexaHeadersWeb|TestSecurity_A02_HexaHeadersAPI|TestSecurity_A02_HexaCSPHasStrictDynamic'` → PASS
    - `curl -sI https://staging.hexascore.app/` → CSP with `'nonce-` and `'strict-dynamic'`, no `unsafe-inline`; HSTS `max-age=63072000; includeSubDomains` without `preload`; `Permissions-Policy` as §6.2
    - `curl -sI https://api.staging.hexascore.app/v1/items -H "Authorization: Bearer $HX_STAGING_KEY"` → API CSP, `Cross-Origin-Resource-Policy: cross-origin`, `access-control-allow-origin: *`, no credentials header; `bunx playwright test headers csp` → pass, 0 violations
  - Notes: `SecureHeaders` is configured, never added twice. HSTS comes from `TRUSTED_PROXIES` + `TRUST_FORWARDED_PROTO` (SDD 11), not `HSTS_FORCE`. `/csp-reports` is in `CSRFBypass` and `SessionBypass` and nowhere else. No `preload` in v1 (issue 10).

#### P2.D.2 Limits, inventory and challenge

- [x] **P2.D.2.1** Implement the full rate-limit table on the shared cache store with fail-closed wiring.
  - Files/modules: `app/http/limits/limits.go` (one named limiter per row of §7.1), `routes/*.go`, `tests/security_limits_test.go`
  - Refs: SDD 13 FR-13-2, NFR-13-2, §7.1, §9 item 4; SDD 03 NFR-03-3; ADR 0016; <https://vukem.dev/docs/security/rate-limiting> "Failure behaviour", "Common mistakes"
  - Depends on: P2.D.1.1
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests -run 'TestSecurity_A06_HexaLimits|TestSecurity_A06_HexaLimitsFailClosed'` → PASS (table-driven over every row; store error → 503)
    - `go test -tags=integration ./tests -run TestLimitsSurviveRestart` → PASS
    - sixth `POST /login` in 1 min from one IP on staging → 429 with `Retry-After` and `RateLimit-Remaining: 0`
  - Notes: `RATE_LIMIT_STORE=cache` over `CACHE_STORE=database`; `memory` only in tests. Each limit has its own name (shared names share buckets). Keys use `ratelimit.ByIP`, never `RemoteAddr`.

- [x] **P2.D.2.2** Add the route stack inventory and the per-route body limits.
  - Files/modules: `tests/security_inventory_test.go`, `app/http/limits/body.go`, `routes/*.go`
  - Refs: SDD 13 FR-13-1, FR-13-4, §6.1, §7.3 (body limits table), §7.0.1; <https://vukem.dev/docs/security/overview>
  - Depends on: P2.D.2.1
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests -run 'TestSecurity_A01_HexaRouteStackInventory|TestSecurity_A06_HexaBodyLimits'` → PASS; adding an unclassified route makes the inventory test fail
  - Notes: limits exactly as §7.3 (forms 64 KiB, Stripe webhook 512 KiB, `client-errors` 16 KiB, `client-metrics` 4 KiB, other 1 MiB). Cover upload and suggestion limits ship with P4.A routes.

- [x] **P2.D.2.3** Complete the proof-of-work challenge and form integrity: global trip, bit fallback, store-down behaviour, minimum fill time, load check.
  - Files/modules: `app/challenge/pow.go`, `app/challenge/pow_test.go`, `app/forms/token.go` (render time in the signed form token), `resources/js/challenge/worker.ts`, `qa/k6/login-pow.js`
  - Refs: SDD 13 FR-13-3, §7.2, §10 (k6 50 VUs); SDD 03 §7.3; ADR 0016; VF-003; P1.C.5.2
  - Depends on: P2.D.2.1
  - Model: Opus
  - Acceptance:
    - `go test ./app/challenge ./tests -run 'TestSecurity_A07_Pow|TestSecurity_A10_HexaPowStoreDown503|TestSecurity_A07_HexaHoneypotAndFillTime'` → PASS (replay → 428, wrong route, expiry, global trip → 21 bits, 1 h hold, `bits-2` once per hour per IP)
    - `k6 run qa/k6/login-pow.js` (50 VUs, PoW active) on staging → app RSS ≤ 512 MiB
  - Notes: extends P1.C.5.2; do not re-implement what it shipped, add the missing rules. The no-JS path offers "Continue with Google" and is never challenged. Honeypot violations get the normal success answer and log `security.honeypot`. Turnstile stays off (needs an ADR 0016 update).

#### P2.D.3 Admin hardening

- [x] **P2.D.3.1** Add the admin middleware: mandatory confirmed 2FA, idle 30 min and absolute 12 h sessions, optional IP allowlist.
  - Files/modules: `app/http/middleware/admin_guard.go`, `app/http/middleware/admin_guard_test.go`, `routes/admin.go`, `.env.example` (`ADMIN_IP_ALLOWLIST`), `vukem-feedback.md` (new entry: role-scoped 2FA, session lifetime, IP allowlist)
  - Refs: SDD 13 FR-13-5, §7.5 (Admin), §9 last item, §11 issue 7; SDD 03 §7.1; <https://vukem.dev/docs/security/rate-limiting> (`c.IP()` behind `TRUSTED_PROXIES`)
  - Depends on: P2.D.2.2, P1.C.3.1
  - Model: Opus
  - Acceptance:
    - `go test ./app/http/middleware -run 'TestSecurity_A01_HexaAdminNeeds2FA|TestSecurity_A07_HexaAdminSessionExpiry|TestSecurity_A01_HexaAdminIPAllowlist'` → PASS
    - staging: an admin without 2FA → redirected to setup; idle 31 min → login
  - Notes: [NOT IN VUKEM DOCS] (issue 7): app middleware, plus a VF entry if none exists. `Regenerate()` on 2FA confirmation and role change. Lost TOTP → recovery codes or an owner CLI reset that writes an audit event. Allowlist empty means off.

#### P2.D.4 LGPD

- [x] **P2.D.4.1** Verify the data-subject rights end to end, add the retention jobs, and update the privacy policy.
  - Files/modules: `app/privacy/retention.go` (session IP ≤ 30 days, `client_reports` 7 days, export ZIPs 24 h), `tests/lgpd_test.go`, `lang/{en,pt-BR,es}.json` (`legal.privacy`), `resources/js/pages/Legal/Page.tsx` content source, `docs/legal/privacy-review.md`
  - Refs: SDD 13 FR-13-6, §5 (inventory), §7.6; SDD 14 FR-14-12; SDD 03 FR-03-15; D-29; P1.F.4.1, P1.C.3.2
  - Depends on: P2.D.3.1
  - Model: Sonnet
  - Acceptance:
    - `go test -tags=integration ./tests -run 'TestLGPD_ExportDeleteOptOut|TestRetentionJobs'` → PASS
    - an export link with an altered `signature`, or used after 24 h, on staging → 403
    - the policy page in three locales lists processors (Railway, Grafana Cloud, Resend, Stripe, Google, Telegram), cookies, retention, minimum age 13 and the encarregado contact; no Vercel, Neon, Upstash, passkey or AbacatePay mention
  - Notes: policy text edits go to the `text-editor` subagent. HUMAN GATE: the D-29/D-21 legal review and the encarregado identity (issues 4–6) are the owner's; record the outcome in `docs/legal/privacy-review.md`. LGPD article text was not fetched (issue 5): do not quote it.

#### P2.D.5 Supply chain, scans and incidents

- [x] **P2.D.5.1** Add Dependabot, `osv-scanner` and `bun audit` to the pipeline and document the fix SLA.
  - Files/modules: `.github/dependabot.yml` (`gomod`, `bun`, `docker`, `github-actions`, weekly), `.github/workflows/ci.yml` (`deps` job), `docs/runbooks/dependencies.md`
  - Refs: SDD 13 NFR-13-5, §7.7, §7.8; SDD 11 §6.3 (Dependabot `docker`)
  - Depends on: P2.B.1.1
  - Model: Sonnet
  - Acceptance:
    - CI `deps` → `osv-scanner` (pinned) on `go.mod` and `bun.lock` and `bun audit --audit-level=high` exit 0
    - `gh api repos/AlexGalhardo/hexascore-app/dependabot/alerts` → reachable; Dependabot opens PRs for the four ecosystems
  - Notes: `osv-scanner` is cited from `vukem/docs/maintenance.md` [SITE PAGE MISSING]; it is a CI binary, not an app dependency, but still pinned by version and checksum. SLA: critical 72 h, high 7 days, medium 30 days.

- [ ] **P2.D.5.2** Run the ZAP baseline and active scans weekly against staging and gate releases on zero High.
  - Files/modules: `qa/zap/plan.yaml` (port, `errorLevel: High`), `qa/zap/run.sh`, `.github/workflows/zap.yml`, `docs/qa/scans.md`
  - Refs: SDD 13 NFR-13-3, §7.8, §9 item 7; SDD 11 §6.6 (`zap`); P1.I.4.1
  - Depends on: P2.D.1.1
  - Model: Sonnet
  - Acceptance:
    - `bash qa/zap/run.sh staging` (ZAP image pinned by digest) → exit 0, 0 High
    - `zap.yml` weekly → opens a GitHub issue on failure (tested once with a deliberately weakened header on a throwaway branch deployed to a local container, never staging)
  - Notes: the active scan writes data: run it with a dedicated staging test account and the seed catalogue; never against production. Medium findings are triaged in `docs/qa/scans.md`.

- [x] **P2.D.5.3** Write the incident response procedure, the review template and the user notification templates.
  - Files/modules: `docs/runbooks/incident-response.md`, `docs/incidents/TEMPLATE.md`, `lang/{en,pt-BR,es}.json` (`mail.incident_notice`) or `docs/runbooks/incident-notice.{en,pt-BR,es}.md`
  - Refs: SDD 13 FR-13-7, NFR-13-6, §7.9; SDD 12 §7.8 R9
  - Depends on: P2.C.4.2
  - Model: Sonnet
  - Acceptance:
    - `bunx markdownlint-cli2 docs/runbooks/incident-response.md docs/incidents/TEMPLATE.md` → 0 errors
    - the procedure covers SEV1–SEV3, the five steps, Railway log export before expiry, the 72 h target and a review within 5 working days
  - Notes: the ANPD deadline is [NOT VERIFIED] (SDD 13 issue 5): write "per ANPD rule, verify before use". Template text goes through the `text-editor` subagent.

#### P2.D.6 Release

- [ ] **P2.D.6.1** Release `v0.13.0` (security hardening).
  - Files/modules: `CHANGELOG.md`, tag `v0.13.0`, `docs/PROGRESS.md`
  - Refs: ADR 0013; SDD 13 §9; SDD 11 §7.1
  - Depends on: P2.D.2.3, P2.D.4.1, P2.D.5.1, P2.D.5.2, P2.D.5.3
  - Model: Sonnet
  - Acceptance:
    - `go test ./... -run 'TestSecurity_'` → PASS; `bash qa/zap/run.sh staging` → 0 High
    - `gh release view v0.13.0` → non-empty body
  - Notes: HUMAN GATE (production deploy).

### P2.E Backups and migrations policy

#### P2.E.1 Migrations

- [x] **P2.E.1.1** Build `hexascore:predeploy`: advisory lock around the migration runner, then the concurrent index step, as the Railway pre-deploy command.
  - Files/modules: `app/console/predeploy.go`, `app/console/index_concurrently.go` (`db:index-concurrently`), `database/concurrent/` (`NNN_name.sql`), `tests/predeploy_test.go`, `docs/runbooks/migrations.md`
  - Refs: SDD 11 §7.2, FR-11-4, §8 (long index build), §11 issues 3–4; SDD 07 §5; ADR 0014; VF-017, VF-026; <https://vukem.dev/docs/database/migrations> "Running migrations"; <https://docs.railway.com/guides/pre-deploy-command>
  - Depends on: P2.D.6.1, P1.A.4.3
  - Model: Opus
  - Acceptance:
    - `go test -tags=integration ./tests -run 'TestPredeploy_ConcurrentRunsSerialise|TestIndexConcurrently_DropsInvalid'` → PASS: two parallel runs, one migrates, the other waits on `pg_advisory_lock(hashtext('hexascore:migrate'))` and finds nothing to do; an `INVALID` index is dropped and rebuilt
    - staging deploy of a branch adding one `database/concurrent/` file → pre-deploy log shows `migrate` then `CREATE INDEX CONCURRENTLY IF NOT EXISTS`; `migrate:status --json` records the matching `CREATE INDEX IF NOT EXISTS` migration
  - Notes: reuses P1.A.4.3's `migrate:locked`; this task only adds the concurrent step and the single entry point. Concurrent statements run outside a transaction on a connection that holds no other lock. Big indexes ship alone (pre-deploy timeout, SDD 11 §8). Retire `database/concurrent/` when VF-017 is fixed.

- [x] **P2.E.1.2** Add the CI expand/contract policy check and a down-migration round trip.
  - Files/modules: `scripts/ci/migration-policy.ts`, `scripts/ci/migration-policy.test.ts`, `.github/workflows/ci.yml` (`migrations` job), `docs/runbooks/migrations.md` (policy, contract marker)
  - Refs: SDD 11 §7.2 (forward-only; expand → migrate → contract two releases later; `down` tested, never run in production); ADR 0014; SDD 06 §7.8 (backfill job); <https://vukem.dev/docs/database/migrations> "Running migrations"
  - Depends on: P2.E.1.1
  - Model: Sonnet
  - Acceptance:
    - `bun test scripts/ci/migration-policy.test.ts` → PASS: `DROP COLUMN`, `DROP TABLE`, `RENAME`, `SET NOT NULL` and a non-concurrent `CREATE INDEX` on `items`, `ratings` or `item_translations` are rejected unless the migration carries a `contract:` marker naming a release at least two minors older
    - CI `migrations` on Postgres 18.6 → `vukem migrate`, `vukem migrate:rollback --step 1`, `vukem migrate` exit 0 for every new migration in the PR
  - Notes: the check reads migration source text; it is a guard, not a SQL parser: keep the rules explicit and listed in the runbook. `migrate:fresh`/`reset` refuse production without `--force`: never add `--force` to scripts.

#### P2.E.2 Daily dump

- [x] **P2.E.2.1** Implement `backup:run`: streamed `pg_dump`, client-side frame encryption, SHA-256 and multipart upload, recorded in `backup_runs`.
  - Files/modules: migration `create_backup_runs`, `app/backup/dump.go`, `app/backup/crypt.go` (4 MiB frames, `crypt.New(BACKUP_KEY)`, `Seal(frame, "backup:<key>:<n>")`), `app/backup/crypt_test.go`, `app/console/backup_run.go`
  - Refs: SDD 11 §5 (`backup_runs`, prefixes), §7.3 steps 2–3, FR-11-5; ADR 0014 (amended); <https://vukem.dev/docs/security/encryption>; <https://vukem.dev/docs/digging-deeper/storage>
  - Depends on: P2.E.1.1
  - Model: Opus
  - Acceptance:
    - `go test ./app/backup -run 'TestFrameEncryptRoundTrip|TestFrameReorderRejected|TestFrameTruncationRejected'` → PASS
    - `vukem backup:run --kind daily` against the compose Postgres and RustFS → object under `backups/postgres/daily/YYYY/MM/DD/`, `backup_runs.status='uploaded'`, `sha256` matches the object
  - Notes: `pg_dump -Fc -Z 6 --no-owner` through `contrib/engine` `RunStream` ([SITE PAGE MISSING], `vukem/docs/subsystems.md`); the child environment is not inherited, so `PG*` values are passed explicitly. The plaintext dump never touches disk. Each frame's additional data carries its index and the last frame is marked, so frames cannot be swapped, dropped or truncated unnoticed.

- [x] **P2.E.2.2** Verify each dump, schedule the daily run, prune by retention, and copy the monthly snapshot.
  - Files/modules: `app/backup/verify.go`, `app/backup/prune.go`, `app/backup/prune_test.go`, `routes/schedule.go` (`backup.daily` 06:00 America/Sao_Paulo, `WithoutOverlapping()`), `app/metrics/backup.go`
  - Refs: SDD 11 §7.3 steps 1, 4, 5, 7, §9 item 6; SDD 12 §6.2 (`backup.*`), §7.5 A8; ADR 0014; <https://vukem.dev/docs/digging-deeper/scheduler>
  - Depends on: P2.E.2.1
  - Model: Sonnet
  - Acceptance:
    - `go test ./app/backup -run 'TestPruneSelection|TestVerifyRejectsSmallDump'` → PASS (30 daily + 12 monthly kept; size ≤ 50 % of the 7-day median fails)
    - on staging, `railway ssh` then `/app/hexascore backup:run --kind daily` → `status='verified'` after decrypt → `pg_restore --list` contains `items`, `users`, `ratings`
    - `backup.last_success_age_seconds` visible in Grafana; the A8 fixture in `ops/alerts/tests.yaml` still passes
  - Notes: buckets have no lifecycle rules: pruning is app code. The task will run after the SDD 06 asserts once they exist (P6); until then 06:00 alone.
  - Status (2026-10-29): verify, monthly copy, prune and the 06:00 schedule delivered; `app/metrics/backup.go` moves to P2.C.2.2 (no meter provider before P2.C.2.1); staging and Grafana checks deferred (owner gate).

- [ ] **P2.E.2.3** Mail the owner a 7-day signed link to the encrypted monthly dump (off-site copy).
  - Files/modules: `app/backup/offsite.go`, `resources/views/mail/backup_offsite.*`, `tests/backup_offsite_test.go`, `docs/runbooks/restore.md` (section "Off-site copy")
  - Refs: SDD 11 §7.3 step 6, §11 issue 11; D-32; <https://vukem.dev/docs/security/signed-urls>
  - Depends on: P2.E.2.2
  - Model: Sonnet
  - Acceptance:
    - `go test -tags=integration ./tests -run TestBackupOffsiteLinkMailed` → PASS: `storage.TemporaryURL` valid 7 days, `backup_runs(kind='offsite')` written when mailed, mail sent to `ADMIN_EMAILS` only
  - Notes: HUMAN GATE (production data): the owner downloads the file each month to personal storage and keeps `BACKUP_KEY` offline separately; the link is never posted to Telegram. Replaced by an external bucket at the first budget increase.

#### P2.E.3 Drills

- [x] **P2.E.3.1** Implement `backup:restore` and the monthly `backup:drill` on staging, including the deletion replay.
  - Files/modules: `app/console/backup_restore.go` (`--target-url`), `app/console/backup_drill.go`, `routes/schedule.go` (first Sunday 07:00, staging only), `tests/backup_drill_test.go`, `docs/runbooks/restore.md`
  - Refs: SDD 11 §7.4, §7.5 step 3, FR-11-6, §9 item 6, §10; SDD 13 §8 (restore re-applies deletions), §11 issue 11; ADR 0014; D-18
  - Depends on: P2.E.2.2
  - Model: Opus
  - Acceptance:
    - `go test -tags=integration ./tests -run 'TestBackupDrill|TestRestoreReappliesDeletions'` → PASS
    - on staging with `BACKUP_SOURCE_S3_*` set: `/app/hexascore backup:drill` → exit 0, `backup_runs(kind='drill')` with `duration_s < 3600`, row counts equal to the dump manifest
  - Notes: the drill creates and drops `hexascore_drill`; it never restores over the staging database. `dq:asserts` runs the catalogue asserts that exist (SDD 01 §7.6); SDD 06 asserts join in P6. HUMAN GATE (production data): granting staging read access to the production bucket is the owner's call.
  - Status (2026-10-29): local acceptance green (`TestBackupDrill`, `TestRestoreReappliesDeletions` PASS in `golang:1.27.3-alpine3.24` + `postgresql18-client`; they skip on hosts without `pg_dump`/`pg_restore`). Deferred (owner gate): the staging run with `BACKUP_SOURCE_S3_*`. Manifest = per-table row counts in the `pg_dump --snapshot` snapshot; `pg_restore` single job on stdin (SDD 11 §7.4 amended). Found and fixed: `hs_norm` unqualified call made every dump unrestorable.

- [ ] **P2.E.3.2** Run a timed RTO game day on staging and add the migration rehearsal script.
  - Files/modules: `scripts/ops/migration-rehearsal.sh` (restore latest dump into staging, deploy candidate, record duration and lock waits), `docs/runbooks/restore.md` (timings), `docs/runbooks/migrations.md` (rehearsal section)
  - Refs: SDD 11 §7.2 (rehearsal; > 5 s `ACCESS EXCLUSIVE` blocks the release), §7.5 (RTO steps 1–6), NFR-11-2; SDD 12 §7.8 R8; ADR 0014; <https://vukem.dev/docs/cli/reference-runtime> "`vukem down` and `vukem up`"
  - Depends on: P2.E.3.1
  - Model: Sonnet
  - Acceptance:
    - game day following `docs/runbooks/restore.md` steps 1–6 on staging → new Postgres service serving within 60 min; time per step recorded
    - `scripts/ops/migration-rehearsal.sh <candidate-commit>` → prints migration duration and max `ACCESS EXCLUSIVE` wait; exit 1 when > 5 s
  - Notes: until P2.F the dump is staging's own (seed catalogue); the rehearsal re-runs on real data in P2.F.5.2. Delete the scratch Postgres service the same day (cost).

#### P2.E.4 Release

- [ ] **P2.E.4.1** Release `v0.14.0` (backups and migrations policy).
  - Files/modules: `CHANGELOG.md`, tag `v0.14.0`, `docs/PROGRESS.md`, `AGENTS.md` (commands: `backup:run`, `backup:drill`, `backup:restore`, `hexascore:predeploy`, `db:index-concurrently`)
  - Refs: ADR 0013, 0014; SDD 11 §7.1, §9
  - Depends on: P2.E.1.2, P2.E.2.3, P2.E.3.2
  - Model: Sonnet
  - Acceptance:
    - production: first scheduled `backup.daily` after the deploy → `backup_runs.status='verified'`
    - `gh release view v0.14.0` → non-empty body
  - Notes: HUMAN GATE (production deploy). Production holds no user data yet; the drill on real data is part of the P2.G exit.

### P2.F Data migration

The importer reads the legacy dump restored into schema `legacy_import` of the target database and writes the new schema; it never connects to Neon. Ids are preserved (SDD 01 §5.4, SDD 03 §5). Stages are idempotent and resumable. Legacy facts: 27 tables in `legacy/apps/web/lib/db/schema/{auth,billing,catalog,social}.ts`; ~10k synthetic users with ~2.1M ratings, marked by id `seed-user-NNNNNN` and email `<id>@example.test` (`legacy/apps/web/lib/db/seed-data.ts`); 78,904 items with covers on Vercel Blob plus orphans of replaced covers (`legacy/docs/backup-local.md`; 20,360 orphans in the 2026-10-10 run, `docs/PROGRESS.md`). The seed script lives at `legacy/apps/web/lib/db/seed.ts` (run as `bun run db:seed` in `apps/web`), not `scripts/seed.ts`.

#### P2.F.1 Export and purge

- [ ] **P2.F.1.1** Write the legacy inventory and export scripts and restore a Neon copy into `legacy_import` on staging.
  - Files/modules: `scripts/legacy/inventory.sql` (per-table counts, seed-user counts by each marker, sizes), `scripts/legacy/export.sh` (`pg_dump -Fc --no-owner --no-acl` with the SDD 11 §7.6 table list), `scripts/legacy/restore-import-schema.sh`, `docs/migration/legacy-inventory.md`
  - Refs: SDD 11 §7.6 (window steps 2–3, excluded tables), FR-11-7; SDD 03 §5; A-04; `legacy/docs/backup-local.md`
  - Depends on: P1.I.7.1, P2.E.4.1
  - Model: Opus
  - Acceptance:
    - `psql "$NEON_RO_URL" -f scripts/legacy/inventory.sql` → counts for every exported and excluded table, written to `docs/migration/legacy-inventory.md` (counts only, no rows)
    - `scripts/legacy/restore-import-schema.sh <dump> "$STAGING_DATABASE_URL"` → schema `legacy_import` holds the exported tables; `public` untouched
  - Notes: HUMAN GATE (production data): the owner creates a read-only Neon role or a Neon branch and passes its URL through the environment, never in a file. `pg_restore` cannot rename a schema: restore into a scratch database, `ALTER SCHEMA public RENAME TO legacy_import`, dump that schema, restore it into the target. The dump is encrypted with `BACKUP_KEY` before upload to `backups/legacy/`; plaintext stays in a temp directory deleted at the end.

- [ ] **P2.F.1.2** Purge the synthetic users from the legacy production database with the legacy seed script, after a verified full Neon dump, and add the importer guard.
  - Files/modules: `docs/runbooks/legacy-purge.md`, `database/legacyimport/guard.go` (deletes leftovers in `legacy_import`, aborts if markers remain), `database/legacyimport/guard_test.go`, `docs/sdd/11-infra-deploy.md` §7.6 step 4 (purge on Neon first, importer guard after)
  - Refs: SDD 11 §7.6 step 4; SDD 03 §5 ("the 10k synthetic users are purged first"); SDD 02 §5 (copy ratings after the purge); A-04; `legacy/apps/web/lib/db/seed.ts` (`--purge`, `purgeSeedUsers`), `legacy/apps/web/lib/db/seed-guard.ts`
  - Depends on: P2.F.1.1
  - Model: Opus
  - Acceptance:
    - rehearsal on a Neon branch, from a temp copy of `legacy/apps/web`: `SEED_TARGET=sandbox bun run db:seed --purge` → `[seed] purged <N> seed users`, N equal to the inventory count
    - production: `SEED_TARGET=production bun run db:seed --purge --i-know-this-is-production` → same N; then `SELECT count(*) FROM "user" WHERE id LIKE 'seed-user-%' OR email LIKE '%@example.test'` → 0
    - `go test ./database/legacyimport -run TestGuardRejectsSeedMarkers` → PASS
  - Notes: HUMAN GATE (production data, irreversible): the owner runs the production command only after a full `pg_dump` of Neon is uploaded to `backups/legacy/` and verified with `pg_restore --list`. The legacy purge matches id prefix AND email domain while SDD 11 says OR: the inventory reports both counts; rows matching only one marker are reviewed by the owner before the guard deletes them in `legacy_import`. The purge recomputes legacy stats, so the legacy site stays consistent for the rollback window. Never edit `legacy/` files.

#### P2.F.2 Import stages

- [x] **P2.F.2.1** Build the `hexascore:import-legacy` runner: stage ordering, checkpoints, dry run, pre-flight asserts.
  - Files/modules: `app/console/import_legacy.go`, `database/legacyimport/runner.go`, `database/legacyimport/preflight.go`, migration `create_legacy_import_runs`, `database/legacyimport/testdata/fixture.sql` (synthetic Better Auth-shaped data), `database/legacyimport/runner_test.go`
  - Refs: SDD 11 §7.6 step 5, §10; SDD 03 §5 (pre-flight asserts); SDD 06 §7.9 (idempotent, resumable; registration [NOT VERIFIED]); A-04
  - Depends on: P2.F.1.1
  - Model: Opus
  - Acceptance:
    - `go test -tags=integration ./database/legacyimport -run 'TestRunnerResumes|TestPreflightAborts'` → PASS: a run killed mid-stage resumes from its checkpoint; duplicate `lower(email)`, `two_factor_enabled` without a row and accounts with no login method abort with counts
    - `vukem hexascore:import-legacy --stage all --dry-run` on the fixture → planned counts per stage, 0 writes
  - Notes: stages `users`, `twofactor`, `catalog`, `entities`, `social`, `library`, `billing`, `finalise` (P2.F.2.2–P2.F.4.1). Batches inside transactions, last key recorded per stage. The fixture holds no real data.

- [ ] **P2.F.2.2** Implement the `users` stage with Better Auth credential and Google mapping, and prove legacy scrypt logins on the imported copy.
  - Files/modules: `database/legacyimport/users.go`, `database/legacyimport/users_test.go`, `database/legacyimport/passkeys.go` (Better Auth `passkey` rows → `passkeys`), `docs/migration/passkey-notice.md` (fallback only: selection query, never addresses)
  - Refs: SDD 03 §5 (mapping table), §7.2, §7.7, §7.8, FR-03-18, §9; ADR 0008, ADR 0017; D-25, D-37; P1.C.1.2
  - Depends on: P2.F.2.1
  - Model: Opus
  - Acceptance:
    - `go test -tags=integration ./database/legacyimport -run TestUsersStage` → PASS: email lowercased, `email_verified` → `email_verified_at`, `role` containing `admin` → `admin`, expired bans dropped, credential hash copied verbatim, Google → `social_accounts(provider='google', provider_id=account_id)`, no credential → random argon2id with `password_set_at` NULL, deletion grace preserved
    - `go test ./app/accounts -run TestSecurity_A07_LegacyScryptVerifiesAndRehashes` → PASS; on staging a fixture account imported with a Better Auth 1.7.7 hash logs in and its stored hash then starts `$argon2id$`
    - `go test -tags=integration ./database/legacyimport -run TestPasskeysStage` → PASS: a Better Auth `passkey` fixture row (credential id, public key, counter, transports, AAGUID, backup flags, name, created_at) imports into `passkeys` and a signed assertion from the fixture authenticator verifies after import
  - Notes: `session`, `verification` and `rate_limit` are not imported. Legacy `passkey` rows are imported (ADR 0017, depends on P1.C.7.1); the Better Auth storage encoding (base64 vs base64url, COSE key layout) is [NOT VERIFIED] until proven on a Neon copy. Only if it proves incompatible: passkey-only users are counted and mailed the `passkeys_retired` reset notice 14 days before the window (P2.G.1.2). OAuth tokens dropped. Never log hashes or key material.
  - Status (2026-10-29): `users` stage and the scrypt login proof delivered; `passkeys.go` and `TestPasskeysStage` wait for P1.C.7.1 (no `passkeys` table before the vukem release with VF-001), tracked as P2.F.2.2b.

- [x] **P2.F.2.3** Export the 2FA secrets with a Bun script using Better Auth's own helpers and import them over stdin.
  - Files/modules: `scripts/legacy/twofactor-export.ts` (run from a temp copy of `legacy/apps/web` with its `node_modules`), `database/legacyimport/twofactor.go`, `database/legacyimport/twofactor_test.go`
  - Refs: SDD 03 §7.6, §11 issue 2 (03-2 decided); SDD 13 §5 (`orm.Encrypted`); P0.F.1.2 (which Better Auth function decrypts)
  - Depends on: P2.F.2.2
  - Model: Opus
  - Acceptance:
    - `bun scripts/legacy/twofactor-export.ts | vukem hexascore:import-legacy --stage twofactor --stdin` on the fixture → secrets sealed, unused backup codes SHA-256-hashed, `two_factor_confirmed_at` set for verified rows; no secret file left in the temp directory
    - `go test -tags=integration ./database/legacyimport -run TestTwoFactorUsersPassChallenge` → PASS: a migrated user passes TOTP with a code from the fixture secret
  - Notes: HUMAN GATE (secrets): `BETTER_AUTH_SECRET` is provided by the owner as an environment variable for the run only. NDJSON `{userId, secret, backupCodes[]}` never touches disk or logs. The encryption format is [NOT VERIFIED]: the script calls the library's decrypt, never a reimplementation. Lockout counters are dropped.

- [x] **P2.F.2.4** Implement the `catalog` stage: `source_records`, `items`, `item_translations`, external ids, affiliate links, cover keys.
  - Files/modules: `database/legacyimport/catalog.go`, `database/legacyimport/catalog_test.go`
  - Refs: SDD 01 §5.3, §5.4 (mapping rows); SDD 06 §7.9; SDD 02 §5 (score columns filled by recompute); D-09; A-04
  - Depends on: P2.F.2.1
  - Model: Sonnet
  - Acceptance:
    - `go test -tags=integration ./database/legacyimport -run TestCatalogStage` → PASS: ids preserved, `popularity` → `popularity_prior`, `release_precision`/`release_status` derived as §5.4, `visibility='public'`, `synopsis_license` → `synopsis_licence`, sources inferred from URL host, `meta.coverUrl` → `cover_key` `covers/<id>-<hash16>.webp`
    - `external_scores` in the new schema → none (dropped, D-09; kept only in the legacy dump)
  - Notes: typed columns (`runtimeMinutes`, `pages`, `seasons`, `episodes`, `trailers`) copied 1:1; `affiliate_links` copied if any. Entities and editions are the next stage.

- [x] **P2.F.2.5** Implement the `entities` stage: genres, platforms, creators and countries from `meta`, and book editions from `isbn13`.
  - Files/modules: `database/legacyimport/entities.go`, `database/legacyimport/entities_test.go`, `docs/migration/entity-resolution.md` (counts: resolved by QID, native, roleless)
  - Refs: SDD 01 §5.4 (`meta.*` and book rows), §7.6 (DQ-09, DQ-10), §11 issue 1; SDD 06 §7.4, §7.9; ADR 0012
  - Depends on: P2.F.2.4
  - Model: Opus
  - Acceptance:
    - `go test -tags=integration ./database/legacyimport -run TestEntitiesStage` → PASS: strings resolved through QIDs in `source_records` payloads, unresolved strings become native entities flagged by DQ-09, roleless creators get `creator`/`author`/`developer` by type, one edition per `isbn13` with `language='und'` when unknown, flagged by DQ-10
    - a book without edition data never fails the stage (01-1)
  - Notes: this is the minimal mapping of SDD 01 §5.4. The full Go normaliser over all records named in SDD 06 §7.9 belongs to P6.A and re-runs later; do not build it here.

- [x] **P2.F.2.6** Implement the `social` and `library` stages: ratings with the legacy rounding, statuses, time-to-beat, lists, list items, reports, contact messages.
  - Files/modules: `database/legacyimport/social.go`, `database/legacyimport/library.go`, `tests/migration/legacy_rounding_test.go`, `database/legacyimport/social_test.go`
  - Refs: SDD 02 FR-02-13, §5 (`rating_migration_audit`, `score_actions` `legacy_rounding`), §9; SDD 14 §5 ("Migration" paragraph, invariants); A-04; D-06
  - Depends on: P2.F.2.2, P2.F.2.4
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests/migration -run TestLegacyRounding` → 0,1,2,3,12,13,97,98,100 map to 5,5,5,5,10,15,95,100,100; audit rows = changed rows; one `legacy_rounding` summary row
    - `go test -tags=integration ./database/legacyimport -run 'TestSocialStage|TestLibraryStage'` → PASS: `list_items.added_at := lists.created_at`, positions dense, legacy mixed-type lists keep their rows, `follows` not imported (archived in the dump)
  - Notes: ratings of purged users are already gone (P2.F.1.2 guard). Notifications have no legacy table.

- [x] **P2.F.2.7** Implement the `billing` stage: API keys unchanged, subscriptions, Stripe customers, usage, and the D-23 credit balances.
  - Files/modules: `database/legacyimport/billing.go`, `database/legacyimport/billing_test.go`, `docs/sdd/05-public-api.md` §5 (remaining-credit formula)
  - Refs: SDD 05 §5 (legacy source, `api_credit_balances`), §7.2; SDD 11 §7.6 step 5, §11 issue 9; D-23; ADR 0015
  - Depends on: P2.F.2.2
  - Model: Opus
  - Acceptance:
    - `go test -tags=integration ./database/legacyimport -run TestBillingStage` → PASS: `api_keys.hash` copied verbatim (a fixture key authenticates on `/api/v1` after import), subscriptions mapped to `plan`/`provider`/`state` with at most one active per user, `api_credit_balances.remaining` = legacy remaining credits per user (≥ 0), `credit_ledger` and `pix_charges` copied verbatim
    - `go test ./... -run TestSecurity_A08_StripeReplayIgnored` → PASS: a replayed `customer.subscription.updated` / `invoice.paid` event id is a no-op (vukem `contrib/stripe` cache-backed dedup) and a new event about the same subscription converges on one row; `processed_events` is not imported (SDD 05 §5: production `CACHE_STORE=database`, handlers idempotent by Stripe object id)
  - Notes: derive the remaining-credit formula from `legacy/apps/web/lib/db/schema/billing.ts` and the legacy ledger code, and write it into SDD 05 §5 before coding. No refund flow (D-23).
  - Status (2026-10-29): done. `TestBillingStage` PASS; `processed_events` dropped by orchestrator decision (QUESTIONS.md), SDD 05, SDD 11 §7.6 and SDD 13 A08 amended (DOC-processed-events); `TestSecurity_A08_StripeReplayIgnored` PASS.

#### P2.F.3 Covers

- [x] **P2.F.3.1** Copy covers referenced by the database from Vercel Blob to the bucket with a manifest, skipping orphans, with a delta mode for the window.
  - Files/modules: `scripts/legacy/blob-export.ts` (Bun, `fetch` + `Bun.s3`, no new dependency), `docs/migration/covers.md` (counts: referenced, copied, orphans skipped, failures)
  - Refs: SDD 11 §7.6 (T-7 d, window step 5 covers delta); SDD 01 §5.4 (cover row); SDD 06 §7.5 (keys); `legacy/docs/backup-local.md`
  - Depends on: P2.F.2.4
  - Model: Sonnet
  - Acceptance:
    - `bun scripts/legacy/blob-export.ts --db "$NEON_RO_URL" --bucket staging --manifest manifest.ndjson` → one `{legacyUrl,key,sha256}` line per DB-referenced cover; a re-run copies 0 (idempotent by key and sha256)
    - `bun scripts/legacy/blob-export.ts --delta --since <ts> ...` → copies only covers referenced after `<ts>`
  - Notes: SDD 11 places the script in `legacy/scripts/`; `legacy/` is read-only and branch `legacy` immutable, so it lives here and reads the Blob URLs over HTTPS. Orphans are never copied; their count feeds the decommission (P2.G.4.3). `Bun.s3` against the bucket's path-style setting is [NOT VERIFIED]: test on staging first. The flags above are this script's own.

#### P2.F.4 Verification

- [x] **P2.F.4.1** Implement `hexascore:cutover-verify` and the `finalise` stage (search rebuild, score recompute).
  - Files/modules: `app/console/cutover_verify.go`, `database/legacyimport/finalise.go`, `tests/cutover_verify_test.go`
  - Refs: SDD 11 §7.6 steps 5–6, §9 item 7, §10; SDD 02 §5, FR-02-13; SDD 07 (`search:rebuild`); ADR 0013
  - Depends on: P2.F.2.3, P2.F.2.5, P2.F.2.6, P2.F.2.7, P2.F.3.1
  - Model: Opus
  - Acceptance:
    - `go test -tags=integration ./tests -run TestCutoverVerify_DetectsSeededMismatch` → exit 1 naming the table for a changed row, an orphan FK and a missing cover key
    - `vukem hexascore:cutover-verify` on the fixture import → exit 0: per-table counts equal (after documented exclusions), 1 % id-sample `md5(string_agg(<cols>, '|' ORDER BY id))` equal, every cover key in the bucket, 0 orphan FKs; `--counts-only` finishes in < 60 s
    - scores: the Go formula and the nightly SQL recompute give equal `items.score` for every item; legacy `items.score` vs the mean of the imported (rounded) ratings within ±1 on 0–100, or the item is listed in the report
  - Notes: the new formula (SDD 02) differs from the legacy aggregate by design: equality is checked between the two new code paths; the legacy comparison is a sanity check on the copy. If the ±1 tolerance fails broadly because of the rounding to multiples of 5, record the measured spread and ask the owner to amend SDD 11 §7.6 step 6.

#### P2.F.5 Dry runs

- [ ] **P2.F.5.1** Dry run 1 on staging from a fresh Neon copy: full export, import, verify; record timings and defects.
  - Files/modules: `scripts/legacy/dry-run.sh` (export → restore → guard → stages → covers delta → finalise → verify, timestamped), `docs/migration/dry-run-1.md`
  - Refs: SDD 11 §7.6, NFR-11-6; SDD 02 §11 issue 2 (median ratings per type on migrated data); SDD 05 §11 issue 11 (`qa/sql/api_coverage.sql` on the first staging copy); ADR 0013
  - Depends on: P2.F.4.1, P2.F.1.2
  - Model: Opus
  - Acceptance:
    - `scripts/legacy/dry-run.sh "$NEON_BRANCH_URL" "$STAGING_DATABASE_URL"` → each step timed; `hexascore:cutover-verify` exit 0
    - `docs/migration/dry-run-1.md` lists step timings, defects with fix commits, median ratings per type (02-2) and API coverage numbers (05-11)
  - Notes: HUMAN GATE (production data): the source is a Neon branch made by the owner, never the primary. Staging then holds real personal data: confirm the staging mail allowlist, `noindex` and Stripe test keys (SDD 11 §6.4) before the restore.

- [ ] **P2.F.5.2** Dry run 2 from a new copy within the time budget, then run the full QA gate on staging with real data.
  - Files/modules: `docs/migration/dry-run-2.md`, `docs/qa/staging-gate.md`
  - Refs: SDD 11 §7.2 (rehearsal), §7.6, NFR-11-6; D-21; P1.I.2.4, P1.I.3.1, P1.I.4.1, P1.I.5.1; P2.E.3.2
  - Depends on: P2.F.5.1
  - Model: Sonnet
  - Acceptance:
    - `scripts/legacy/dry-run.sh` on a fresh branch → steps 2–6 ≤ 40 min (20 of the 60 left for freeze, DNS and checks); verify exit 0
    - against staging: `bunx playwright test` (3 locales, axe 0 serious/critical), `k6 run qa/k6/api-v1.js`, Schemathesis `--checks all`, `bash qa/zap/run.sh staging` (0 High), `bunx lhci autorun` → green; results in `docs/qa/staging-gate.md`
    - `scripts/ops/migration-rehearsal.sh` on the restored real data → max `ACCESS EXCLUSIVE` wait ≤ 5 s
  - Notes: failures become fix microtasks before P2.F.6.1; a gate is waived only by an owner decision recorded in `docs/qa/staging-gate.md`.

#### P2.F.6 Release

- [ ] **P2.F.6.1** Release `v0.15.0` (data migration tooling, two dry runs green).
  - Files/modules: `CHANGELOG.md`, tag `v0.15.0`, `docs/PROGRESS.md`, `AGENTS.md` (commands: `hexascore:import-legacy`, `hexascore:cutover-verify`, `scripts/legacy/*`)
  - Refs: ADR 0013; SDD 11 §7.1, §7.6
  - Depends on: P2.F.5.2
  - Model: Sonnet
  - Acceptance:
    - `gh release view v0.15.0` → non-empty body; production deploy green, production database still without user data
  - Notes: HUMAN GATE (production deploy).

### P2.G Cutover

Timeline (SDD 11 §7.6): T-14 d rehearsal, T-7 d bulk covers, T-24 h TTL 60 s, window ≤ 60 min (target 45), watch 2 h then daily for 7 days, rollback until T+7 d, decommission at T+30 d.

#### P2.G.1 Preparation

- [x] **P2.G.1.1** Implement `hexascore:export-since` and the replay into legacy for rollback R3.
  - Files/modules: `app/console/export_since.go`, `scripts/legacy/replay.sh`, `tests/export_since_test.go`, `docs/runbooks/cutover-rollback.md`
  - Refs: SDD 11 §7.6 (Rollback R1–R3), §10 ("export-since replay"); ADR 0013; D-21
  - Depends on: P2.F.6.1
  - Model: Opus
  - Acceptance:
    - `go test -tags=integration ./tests -run TestExportSince_ReplayIntoLegacyShape` → PASS: ratings, statuses, TTB, lists upserted by natural key into a legacy-shaped schema; new API keys carried; new accounts emitted with a reset-link flag (argon2id is unreadable by Better Auth)
    - ratings created after the cutover replay as legacy 0–100 values unchanged (multiples of 5 are valid legacy values)
  - Notes: accepted losses (sessions, suggestions, flag changes) are listed in the runbook. Replay targets Neon only after R2 (read-only off) and never runs automatically.

- [ ] **P2.G.1.2** Write the cutover runbook with the go/no-go checklist, and send the pre-cutover notices.
  - Files/modules: `docs/runbooks/cutover.md` (T-14 d … T+30 d, commands, owners, timings from dry run 2, abort points), `docs/runbooks/cutover-rollback.md`, `docs/runbooks/dns-legacy.txt` (refreshed)
  - Refs: SDD 11 §7.6, NFR-11-6; SDD 03 §7.7, §8 (passkey-only users mailed); SDD 13 NFR-13-3; D-21, D-29; ADR 0013
  - Depends on: P2.G.1.1
  - Model: Opus
  - Acceptance:
    - `bunx markdownlint-cli2 docs/runbooks/cutover*.md` → 0 errors; every window step has a command, an expected output, a time budget and an abort action
    - the Google Cloud console lists the new callback URI beside the legacy one (SDD 03 §7.7); passkey-only users mailed ≥ 14 days before the window only if the passkey import fell back (ADR 0017)
  - Notes: HUMAN GATE (external consoles, user mail): the owner adds the Google callback and approves the notice. The legacy banner (T-24 h) needs a change to the legacy deployment while branch `legacy` is immutable; the mechanism is [NOT VERIFIED] and the owner decides (for example a setting the legacy app already reads, or a status-page notice only).

#### P2.G.2 Rehearsal and go/no-go

- [ ] **P2.G.2.1** Run the full timed rehearsal at T-14 d on staging, including rollback R1–R3.
  - Files/modules: `docs/migration/rehearsal.md` (actual time per step, deviations)
  - Refs: SDD 11 §7.6 (T-14 d), §9 item 7, NFR-11-6; SDD 12 §10 (staging probe is a D-21 gate); ADR 0013; D-21
  - Depends on: P2.G.1.2
  - Model: Opus
  - Acceptance:
    - on a fresh Neon branch: freeze (`ALTER DATABASE <db> SET default_transaction_read_only = on;`), steps 2–6, a DNS switch of a test hostname on Cloudflare, step 8 checks → total ≤ 60 min, `hexascore:cutover-verify` exit 0
    - rollback: test hostname back to the Vercel target (R1), read-only off on the branch (R2), `hexascore:export-since` replay of staging writes into the branch (R3) → rows visible in the legacy shape
  - Notes: HUMAN GATE (production data, DNS): the Neon branch and the test record are the owner's. The Stripe webhook switch is rehearsed with test-mode endpoints only.

- [ ] **P2.G.2.2** Hold the go/no-go: parity checklist re-signed on staging, gates green, drill passed, legal review done.
  - Files/modules: `docs/parity-checklist.md` (re-signed on staging), `docs/migration/go-no-go.md`
  - Refs: ROADMAP phase map (P2 exit); D-21, D-29; SDD 13 NFR-13-3; SDD 11 §9; P1.I.6.1, P2.F.5.2, P2.E.3.1
  - Depends on: P2.G.2.1
  - Model: Opus
  - Acceptance:
    - `docs/migration/go-no-go.md` records, with dates: parity checklist 100 % on staging, staging gate green (P2.F.5.2), ZAP 0 High, latest `backup:drill` pass, rehearsal ≤ 60 min, legal review (D-29) done, Stripe live keys and webhook prepared, owner's go
  - Notes: HUMAN GATE: only the owner signs "go" and picks the window (low traffic, owner awake for 2 h after). Any "no" moves the date.

#### P2.G.3 Window

- [ ] **P2.G.3.1** Prepare production: bulk covers copy at T-7 d, live secrets, custom domains staged, TTL 60 s at T-24 h.
  - Files/modules: `docs/runbooks/cutover.md` (checked boxes), `docs/migration/covers.md` (production counts)
  - Refs: SDD 11 §6.1, §6.4, §7.6 (T-7 d, T-24 h), NFR-11-6; SDD 05 (Stripe live); D-28
  - Depends on: P2.G.2.2
  - Model: Sonnet
  - Acceptance:
    - `bun scripts/legacy/blob-export.ts --bucket production ...` → manifest count equals the referenced-cover count of the inventory
    - `dig +noall +answer hexascore.app` (and `www`, `api`) → TTL 60 from T-24 h
    - production `/up?checks=1` on the Railway domain → 200, all checks `ok`, release tag in `X-Hexascore-Version`
  - Notes: HUMAN GATE (DNS, secrets): the owner sets Stripe live keys and lowers the TTLs. Whether Railway accepts and certifies custom domains before DNS points at it, and how long TLS issuance takes, are [NOT VERIFIED]: check the service networking settings and put the measured delay into the runbook.

- [ ] **P2.G.3.2** Run the read-only window and the final migration into production.
  - Files/modules: `docs/migration/cutover-log.md` (UTC timestamps per step, counts, verify output)
  - Refs: SDD 11 §7.6 window steps 1–6, FR-11-7, NFR-11-6; SDD 03 §7.7; D-21; A-04
  - Depends on: P2.G.3.1
  - Model: Opus
  - Acceptance:
    - Neon: `SHOW default_transaction_read_only` in a new session → `on`; legacy Stripe webhook disabled
    - `scripts/legacy/dry-run.sh "$NEON_URL" "$PRODUCTION_DATABASE_URL"` (window mode) → guard, stages, covers delta, `search:rebuild`, recompute done; `hexascore:cutover-verify` exit 0
    - freeze to verify ≤ 45 min target, 60 max
  - Notes: HUMAN GATE (production data): the owner runs or approves every command in the session. Abort rule: verify fails or 60 min pass → R2 (read-only off on Neon), the window ends with legacy serving, nothing else changes. `vukem down` keeps the new production closed to visitors until DNS switches.

- [ ] **P2.G.3.3** Switch DNS and the Stripe live webhook, and turn on production monitoring.
  - Files/modules: `docs/migration/cutover-log.md`, `ops/grafana/synthetics.md` (production checks)
  - Refs: SDD 11 §6.1 (apex CNAME flattening, proxy off), §7.6 step 7, §9 item 8; SDD 12 §7.4, §7.7; D-28; ADR 0002
  - Depends on: P2.G.3.2
  - Model: Sonnet
  - Acceptance:
    - `dig +short hexascore.app`, `www.hexascore.app`, `api.hexascore.app` → Railway targets; `curl -sI https://www.hexascore.app/` → 301 to the apex
    - `curl -fsS https://hexascore.app/up?checks=1` → 200, all checks `ok`; `curl -sI https://hexascore.app/up | grep -i x-hexascore-version` → the release tag
    - Stripe: live webhook at the new route with the new secret, a test event processed; synthetic probes A1/A2 green on production
  - Notes: HUMAN GATE (DNS, secrets): the owner changes the records and the Stripe endpoint. `vukem up` once DNS resolves; post the status-page notice. On failure: R1 from `docs/runbooks/dns-legacy.txt`.

#### P2.G.4 After the cutover

- [ ] **P2.G.4.1** Watch production for 2 h and then daily for 7 days; close the rollback window at T+7 d.
  - Files/modules: `docs/migration/post-cutover.md` (daily rows: 5xx %, SSR fallbacks, login success vs legacy baseline, Stripe events, job failures, backup status)
  - Refs: SDD 11 §7.6 step 8 and Rollback ("until T+7 d, then forward-fix"); SDD 12 §7.5; ADR 0014; D-21
  - Depends on: P2.G.3.3
  - Model: Sonnet
  - Acceptance:
    - each day: 5xx < 0.5 %, `hexa.ssr.fallbacks{public="true"}` = 0, login success ≥ 95 % of the legacy baseline, every Stripe event processed, no job kind failing twice
    - first production `backup.daily` → `verified`; the next `backup:drill` on staging from the production dump → pass
  - Notes: a breach before T+7 d → the owner decides between forward fix and R1–R3. At T+7 d record "rollback window closed, forward-fix only" in `docs/PROGRESS.md`. Neon stays read-only until T+30 d.

- [ ] **P2.G.4.2** Release `v1.0.0` (parity and cutover).
  - Files/modules: `CHANGELOG.md`, tag `v1.0.0`, `docs/PROGRESS.md`, `ROADMAP.md` (status table, phase map: P2 done), `AGENTS.md`
  - Refs: ADR 0013 ("`v1.0.0` = parity with the legacy app and successful cutover"); SDD 11 §7.1; D-20, D-21; ROADMAP phase map (P2 exit criteria)
  - Depends on: P2.G.3.3
  - Model: Sonnet
  - Acceptance:
    - exit criteria met: DNS on Railway (P2.G.3.3), rollback window started, `backup:drill` pass on a production dump (P2.G.4.1)
    - `gh release view v1.0.0` → non-empty body; `curl -sI https://hexascore.app/up | grep -i x-hexascore-version` → `v1.0.0`
  - Notes: HUMAN GATE (release, production deploy): the owner chooses the moment, at the earliest after the first production drill passes, and pushes the tag (one tag per push).

- [ ] **P2.G.4.3** At T+30 d take the final legacy dump, decommission Vercel, Neon, Upstash and Blob, and re-baseline the cost sheet.
  - Files/modules: `docs/migration/decommission.md` (checklist with dates), `docs/sdd/11-infra-deploy.md` (§6.5 measured averages, §11 issue 13 closed), `docs/sdd/open-issues.md` (row 11-13)
  - Refs: SDD 11 §7.6 (T+30 d), §6.5, §11 issue 13; SDD 12 §7.9; ADR 0002, 0013; D-21
  - Depends on: P2.G.4.1
  - Model: Sonnet
  - Acceptance:
    - final Neon dump encrypted under `backups/legacy/` (kept 12 months), verified with `pg_restore --list`
    - checklist done: Vercel project (including `dev.hexascore.app`), Neon project, Upstash database, Blob store (orphans included), legacy Stripe webhook, old Google callback and `BETTER_AUTH_SECRET` removed; `legacy_import` dropped by a migration
    - SDD 11 §6.5 updated with 30 days of Railway metrics; `railway usage limit status` still US$30
  - Notes: HUMAN GATE (irreversible deletion, credentials): each deletion is a separate owner action, only after the final dump is verified. Branch `legacy` on GitHub stays; nothing is deleted from git.

### P2.H Launch readiness (audit 2026-10-31)

Goal: close the gap between "parity" and "a product people can use in public". Source: the architecture audit of 2026-10-31 (browser test of `v0.11.0`, findings A1–A22 in `docs/PROGRESS.md`). Every item is code-only, needs no owner gate, and ships before `v1.0.0`. Fixed in the audit itself (no box): mobile header, visible category labels on the star widget, cookie banner overlap, brand fallback, onboarding after a late verification, list indexing rule, profile and list 404 with the shell, API `unknown type` message, demo seeder distribution, report dialog on profiles and lists, dashboard page.

#### P2.H.1 Visual catalogue

- [x] **P2.H.1.1** Serve item covers in every listing: home highlights, section listings, search results, library, lists and public profiles.
  - Files/modules: `app/catalog/listing_view.go`, `app/catalog/item_page.go` (cover URL builder on the `S3_*` bucket, placeholder when absent), `resources/js/components/CoverThumb.tsx`, `resources/js/components/catalog/ItemCard.tsx`, `resources/js/components/catalog/SectionHighlights.tsx`, `resources/js/components/library/ItemLink.tsx`
  - Refs: SDD 09 §7.5 (`CoverThumb` 32×48 / 44×64, `srcSet` 96w/600w, lazy; item cover eager with `fetchpriority="high"`); SDD 06 covers pipeline (P6.B.2); ADR 0004 budgets
  - Depends on: —
  - Model: Sonnet
  - Acceptance:
    - `go test ./app/catalog -run TestCoverURL` → placeholder for items without a cover key, bucket URL otherwise
    - `bun run test -- CoverThumb` → renders `srcSet`, `loading="lazy"` and the placeholder
    - `bunx --bun @lhci/cli@0.15.1 autorun` → LCP ≤ 2.0 s and CLS ≤ 0.05 still green with covers on `/games`
  - Notes: covers come from the legacy copy (P2.F.3) or the demo seeder (add 300 placeholder keys to DemoCatalog so screenshots and e2e see real images).

- [x] **P2.H.1.2** Rebuild the item page layout: cover and score badge in the header, details next to the synopsis, rating widget as the primary action above the fold on mobile, trailers and editions below.
  - Files/modules: `resources/js/pages/Catalog/Show.tsx`, `resources/js/components/item/RatingPanel.tsx`, `resources/js/components/rating/StarRating.tsx` (SVG stars, 44 px targets)
  - Refs: SDD 09 §6.2, §7.4, §7.5; audit A2, A19
  - Depends on: P2.H.1.1
  - Model: Opus
  - Acceptance:
    - `bun run e2e -- item.spec.ts` → rating save, update and delete pass in the three locales, desktop and mobile; axe 0 serious
    - a saved rating shows a success toast (`Flash`) and the community score re-renders without a full page load
  - Notes: keep the six-category input; a star widget drawn with inline SVG replaces the text glyphs so half stars render the same on every platform.

- [x] **P2.H.1.3** Give the home page a first impression: hero with the value proposition, the four sections with covers, a "recently rated by members" strip, and the `Explore` call to action.
  - Files/modules: `resources/js/pages/Home/Index.tsx`, `app/controllers/home.go` (recent public ratings, 12 rows, cached 60 s)
  - Refs: SDD 09 §6.2 (`Home/Index`), SDD 08 §6 (trending arrives with P6.C; until then the strip is "recently rated"); `home.title`, `home.explore` catalogue keys
  - Depends on: P2.H.1.1
  - Model: Sonnet
  - Acceptance:
    - `curl -s localhost:3000/ | grep -c '<img'` ≥ 12
    - Lighthouse `/` mobile: performance ≥ 0.9, LCP ≤ 2.0 s (hero text or first cover)

- [x] **P2.H.1.4** Library and lists as shelves: cards with cover, status and own score, grouped by status, with counts in the filters; public profile tabs reuse the same cards.
  - Files/modules: `resources/js/pages/Library/Index.tsx`, `resources/js/pages/Lists/Index.tsx`, `resources/js/pages/Lists/Edit.tsx`, `resources/js/pages/User/Profile.tsx`, `resources/js/pages/User/List.tsx`, `resources/js/components/library/ShelfCard.tsx`
  - Refs: SDD 14 FR-14-7 (library grouping), FR-14-6; legacy `docs/frontend.md` "Padrão de listagens"
  - Depends on: P2.H.1.1
  - Model: Sonnet
  - Acceptance:
    - `bun run e2e -- library.spec.ts lists.spec.ts social.spec.ts` → green in the three locales, desktop and mobile
    - `bun run test -- ShelfCard` → cover, status label and score render; missing score renders the dash

- [x] **P2.H.1.5** Let a list owner add items from the list page: search box with suggest inside `Lists/Edit`, then `PUT /lists/{id}/items/{itemId}`.
  - Files/modules: `resources/js/pages/Lists/Edit.tsx`, `resources/js/components/search/SuggestBox.tsx` (reusable `onPick`), `app/controllers/lists_controller.go`
  - Refs: SDD 14 §6.2; audit (list edit page has no way to add items)
  - Depends on: —
  - Model: Sonnet
  - Acceptance:
    - `bun run e2e -- lists.spec.ts` → a new spec adds an item from the edit page and sees it in the public list

#### P2.H.2 Interaction states

- [x] **P2.H.2.1** Render a server-side skeleton for the CSR pages (`/profile`, `/library`, `/lists`, `/me/notifications`, `/profile/api`, `/billing`) so the document is never blank before hydration; set the `<title>` server-side for them.
  - Files/modules: `resources/js/components/Skeleton.tsx`, the six pages, `app/controllers/*` (`inertia.Props` title), SSR of a shell-only render
  - Refs: SDD 09 §6.2 (CSR pages), §7.5 ("every Defer group has a skeleton"); audit A4
  - Depends on: —
  - Model: Opus
  - Acceptance:
    - `curl -s -b "$SESSION" localhost:3000/profile | grep -c 'aria-busy="true"'` → 1 and `<title>Profile` present in the HTML
    - Playwright `profile.spec.ts` → no blank frame between navigation and first paint (screenshot assertion)

- [ ] **P2.H.2.2** Make the hydration gap visible: a disabled state with a spinner on submit buttons until `inert` is lifted, and a no-JS fallback for login and signup (plain form post works without the challenge when `CHALLENGE_*` allows it).
  - Files/modules: `resources/js/app.tsx`, `resources/js/components/Field.tsx`, `resources/js/pages/Auth/Login.tsx`, `resources/js/pages/Auth/Signup.tsx`, `app/http/middleware/challenge.go`
  - Refs: audit A5; SDD 13 A07 (challenge); SDD 09 §7.6
  - Depends on: —
  - Model: Opus
  - Acceptance:
    - `go test -tags integration ./tests -run TestSecurity_A07` → green (no weakening)
    - Playwright: a click on "Log in" 100 ms after `DOMContentLoaded` either submits or shows the pending state; never a silent no-op

- [x] **P2.H.2.3** Success feedback after rating, status, time-to-beat and list changes (toast via `Flash`), and a message on `/user/two-factor` when the admin area redirected there.
  - Files/modules: `app/controllers/ratings_controller.go`, `app/controllers/library_controller.go`, `routes/admin.go` (flash before the redirect), `resources/js/components/Flash.tsx`
  - Refs: SDD 09 §7.9; audit A16, A19
  - Depends on: —
  - Model: Sonnet
  - Acceptance:
    - `go test -tags integration ./tests -run 'TestRatings|TestLibrary'` → every mutation answers 303 with a `status` flash
    - Playwright `item.spec.ts` → the toast text is visible after a save

#### P2.H.3 API DX and quality debt

- [x] **P2.H.3.1** Declare `type` as an enum (`games`, `movies`, `series`, `books`) in the OpenAPI documents and regenerate the SDKs; allow a burst on the free plan (`rate_limit.burst = 5`) while keeping 1 request/second sustained.
  - Files/modules: `app/controllers/publicapi/items.go` (`doc:"..."` enum), `openapi/api-v1.json`, `openapi/app-v1.json`, `config/billing.go`, `app/api/quota_middleware.go`, `sdks/ts`, `sdks/python`, `resources/docs/rate-limits.md`
  - Refs: SDD 05 §6.2, §6.5; D-33 (limits reviewed 90 days after cutover: this is the first review); audit A10, A11
  - Depends on: —
  - Model: Sonnet
  - Acceptance:
    - `bash scripts/openapi-check.sh --write && bun run openapi:lint && bun run sdk:generate` → green; `jq '.paths["/api/v1/{type}"].get.parameters[0].schema.enum' openapi/api-v1.json` → the four values
    - `go test -tags integration ./tests -run TestAPIRateLimit` → 5 immediate requests pass, the 6th within the same second is 429 with `Retry-After`

- [x] **P2.H.3.2** Fix the flaky `TestSecurity_A07_HexaHoneypotAndFillTime` under `-race` (CI run 41169281164 failed, passes locally): isolate the challenge counters per test or stop sharing the limiter store between parallel tests.
  - Files/modules: `tests/pow_hardening_test.go`, `app/http/middleware/challenge.go`, `app/challenge/*`
  - Refs: audit A21; CI job "Tests (Go unit and integration on Postgres, TypeScript)"
  - Depends on: —
  - Model: Opus
  - Acceptance:
    - `go test -race -count=20 -tags integration ./tests -run TestSecurity_A07_HexaHoneypotAndFillTime` → 20/20 green

- [x] **P2.H.3.3** One admin layout: the custom `admin/ingestion` page uses the vukem admin navigation (same 11 entries) or is registered as a vukem admin page.
  - Files/modules: `app/admin/*`, `resources/views/admin/ingestion.html` (or the vukem admin registration)
  - Refs: <https://vukem.dev/docs/digging-deeper/admin>; audit A17
  - Depends on: —
  - Model: Sonnet
  - Acceptance:
    - `curl -s -b "$ADMIN_SESSION" localhost:3000/admin/ingestion | grep -c 'href="/admin/score-bursts"'` → 1

- [ ] **P2.H.3.4** Define a `primary` design token (buttons, links, focus) in `packages/tokens` and replace the scaffold's hardcoded `indigo-*` classes in the 21 files that use them.
  - Files/modules: `packages/tokens/tokens.json`, `packages/tokens/build.ts`, `packages/tokens/tokens.css`, `resources/js/**/*.tsx`
  - Refs: SDD 09 §7.5; ADR 0009 (tokens shared with mobile); audit A20
  - Depends on: —
  - Model: Sonnet
  - Acceptance:
    - `grep -rn "indigo-" resources/js | wc -l` → 0; `bun run test -- tokens` → round-trip test green; `bun run --cwd apps/mobile typecheck` → green

- [ ] **P2.H.3.5** Cache anonymous HTML for the public catalogue pages (SDD 09 §7.2) and prove it with the k6 TTFB scenario.
  - Files/modules: `app/http/middleware/render_cache.go`, `routes/web.go`, `qa/k6/page-ttfb.js`
  - Refs: SDD 09 §7.2; VF-025; audit A22
  - Depends on: —
  - Model: Opus
  - Acceptance:
    - `curl -sI localhost:3000/games | grep -i cache-control` → `public, max-age=60, stale-while-revalidate=300` for anonymous requests; `private, no-store` with a session cookie
    - `k6 run qa/k6/page-ttfb.js` → p75 ≤ 400 ms

- [ ] **P2.H.3.6** Release `v0.16.0` (launch readiness).
  - Files/modules: `CHANGELOG.md`, tag `v0.16.0`, `docs/PROGRESS.md`, `ROADMAP.md` status table
  - Refs: ADR 0013
  - Depends on: P2.H.1.1–P2.H.3.5
  - Model: Sonnet
  - Acceptance:
    - `gh release view v0.16.0` → non-empty body; `bun run e2e` → green; `bunx --bun @lhci/cli@0.15.1 autorun` → green
  - Notes: the owner pushes the tag.

## Phase 3: public API

Dependencies on Phases 1–2 are written as group IDs (`P1.H`, `P2.G`) until those phases are expanded; read them as "the last microtask of that group is done". Release numbering follows ADR 0013 (one minor per task group that ships to staging): P3.A `v1.1.0`, P3.B `v1.2.0`, P3.C `v1.3.0`, P4.A `v1.4.0`, P4.B `v1.5.0`. The phase map in `ROADMAP.md` still shows one minor per phase; renumber it (and P5/P6) when this file is merged. `API_VERSION` (public API SemVer, SDD 05 §6.5) moves separately: additive minor per release that changes `openapi/api-v1.json`.

### P3.A API DX

#### P3.A.1 Entity, editions and trending endpoints

- [x] **P3.A.1.1** Serve `getEntity` and the kind-specific `getPerson`, `getOrganization`, `getFranchise`, `getPlatform` on `/api/v1` with audience `api`.
  - Files/modules: `app/api/v1/entities.go`, `app/catalog/readmodel/entity.go`, `routes/api.go`, `tests/api_entities_test.go`
  - Refs: SDD 05 §6.2, §7.4, FR-05-6, FR-05-8; SDD 01 §6.1, §6.2; ADR 0006
  - Depends on: P2.G
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests -run 'TestAPIEntities|TestAPIContract'` → `ok`; `/v1/people/{id}` with an organisation id → 404 `not-found`
    - `curl -si -H "Authorization: Bearer $K" -H "If-None-Match: $ETAG" http://localhost:3000/api/v1/entities/$ID` → `304`, `X-Quota-Remaining` down by 1 (FR-05-4)
  - Notes: typed routes via `api.Register` with `Op.ValidateResponse: true` and `Op.Name` in the `api.v1.` group (<https://vukem.dev/docs/rest-apis/openapi> "OpenAPI"). Entity ids are UUID v7 (README convention). Relations are emitted only for rows whose `source` is allowed (SDD 01 §6.1). Do not add `/app/v1` twins here (P3.C.2.1).

- [x] **P3.A.1.2** Serve `listEntityItems` (`/v1/entities/{id}/items?type=&role=`) with the signed keyset cursor.
  - Files/modules: `app/api/v1/entities.go`, `app/api/cursor/cursor.go` (reuse the P1.H cursor), `tests/api_entities_test.go`
  - Refs: SDD 05 §6.2 "Cursors"; SDD 01 §6.2; VF-005; VF-014
  - Depends on: P3.A.1.1
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests -run 'TestAPIEntityItems|TestAPICursorTamper'` → `ok`; tampered cursor → 400 `invalid-cursor`; `limit` above plan cap → 400 naming the cap
  - Notes: VF-005 workaround: predicate built with `WhereRaw` over `(sort column, id)` and an HMAC-signed opaque cursor; envelope `{data, meta:{next_cursor}, links:{next}}` exactly as vukem emits it (<https://vukem.dev/docs/rest-apis/pagination> "Envelopes"; VF-014 casing accepted). Never pass `sort` to `req.CursorPaginate` (it rejects it, "Query conventions").

- [x] **P3.A.1.3** Serve `listEditions` (`/v1/books/{id}/editions`) for book works.
  - Files/modules: `app/api/v1/editions.go`, `app/catalog/readmodel/edition.go`, `tests/api_editions_test.go`
  - Refs: SDD 05 §6.2; SDD 01 §6.2, §6.3; ADR 0012; D-10
  - Depends on: P3.A.1.2
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests -run 'TestAPIEditions|TestAPIContract'` → `ok`; a non-book id → 404; merged work id → 301 to the target (SDD 01)
  - Notes: edition covers only when `cover_source` is allowed (Open Library pending A-05, SDD 01 issue 2: identifiers and titles only until `docs/legal/open-library.md` exists).

- [ ] **P3.A.1.4** Serve `listTrending` (`/v1/{type}/trending?window=day|week`, top 100) and the business-only `getItemActivity`.
  - Files/modules: `app/api/v1/trending.go`, `app/recommend/trending/read.go`, `tests/api_trending_test.go`
  - Refs: SDD 05 §6.2, FR-05-17; SDD 08 §6 (`trending_snapshots`, `item_activity_daily`), §9
  - Depends on: P6.C (snapshot job and tables; see Notes)
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests -run 'TestAPITrending|TestAPIActivityPlan'` → `ok`; pro key on `/activity` → 403 `plan-required`; business key → 200, counted as 1 request
    - `curl -sI http://localhost:3000/api/v1/games/trending -H "Authorization: Bearer $KEY"` → `ETag` present; same request with `If-None-Match` → 304 (SDD 08 §9)
  - Notes: P6.C ships after P3 in the phase order, but it only depends on P2.G. Owner input: pull P6.C forward before this microtask, or leave this box unchecked and release P3.A without trending. Never return a fabricated or empty list as if it were real.

#### P3.A.2 Provenance metadata and coverage

- [x] **P3.A.2.1** Make the audience `api` serialiser emit `provenance` per field and `meta.omittedFields` on every `/api/v1` payload, and keep `lookupItem` from echoing restricted ids.
  - Files/modules: `app/catalog/readmodel/audience.go`, `app/api/v1/serialize.go`, `tests/api_licence_test.go`, fixtures `tests/fixtures/tmdb_igdb_item.json`
  - Refs: SDD 05 FR-05-6, FR-05-16, §6.2 example; SDD 01 §6.1; ADR 0006 decisions 1–2; open issues 05-3
  - Depends on: P3.A.1.3
  - Model: Opus
  - Acceptance:
    - `go test ./tests -run TestAPILicenceFilter` → TMDB/IGDB fixture fields absent from `/api/v1`, present in `/app/v1` (SDD 05 §9)
    - `go test ./tests -run TestAPILookupNoRestrictedEcho` → lookup by TMDB id returns no `tmdb` key in the body
  - Notes: SDD 05 §6.2 shows `meta.omittedFields` as an array of field names, SDD 01 §6.3 as a count; use the array (SDD 05 owns the API contract) and fix SDD 01 §6.3 in the same change. Excluded fields are omitted, never `null`. `CATALOG_API_SOURCES` stays configuration, no code allowlist.

- [ ] **P3.A.2.2** Write `qa/sql/api_coverage.sql` (per-field share of public items with an allowed source, per type) and record the staging numbers.
  - Files/modules: `qa/sql/api_coverage.sql`, `docs/sdd/05-public-api.md` §6.2 coverage table, `resources/data/api_coverage.json` (read by P3.B.3.2)
  - Refs: SDD 05 §6.2 "Coverage", open issue 05-11; ADR 0006 decision 6; D-33
  - Depends on: P3.A.2.1
  - Model: Sonnet
  - Acceptance:
    - `psql "$STAGING_DATABASE_URL" -f qa/sql/api_coverage.sql` → one row per (type, field) with `pct` 0–100 and no error
    - `git diff --stat docs/sdd/05-public-api.md` → coverage table updated, `[NOT VERIFIED]` removed
  - Notes: run on the staging copy only (read-only role). The SQL reads `provenance` jsonb and `*_source` columns using the same allowlist as the serialiser; a mismatch is a bug in one of them. Closes open issue 05-11.

#### P3.A.3 Changes feed

- [x] **P3.A.3.1** Add `api_changes` and a single writer called inside the cause's transaction (merge, ingestion upsert, `ItemScoreChanged`, and later SDD 04 apply/revert).
  - Files/modules: `database/migrations/<ts>_create_api_changes.go`, `app/api/changes/record.go`, call sites in `app/catalog/merge.go`, `app/scoring/aggregate.go`, `tests/api_changes_record_test.go`
  - Refs: SDD 05 §5 (`api_changes`), §7.3; SDD 02 §6 (score fields); SDD 01 merge
  - Depends on: P3.A.2.1
  - Model: Opus
  - Acceptance:
    - `vukem migrate && vukem migrate:status` → `create_api_changes` Ran
    - `go test ./app/api/changes/... -run 'TestRecord|TestHiddenOnlyChangeWritesNothing|TestScoreChangeOnlyOnDecimal'` → `ok`
  - Notes: SDD 05 §5 types `item_id` as `bigint`, but the README convention says UUID v7 primary keys; use `uuid` and fix SDD 05 §5 (and `merged_into`) in the same change. `fields` holds api-visible names only (serialiser allowlist), so a TMDB-only refresh writes nothing. Score rows only when the one-decimal score, confidence or `ratingCount` changes.

- [x] **P3.A.3.2** Serve `listChanges` (`/v1/changes?since=&type=&cursor=&limit≤1000`) and prune rows older than 30 days daily.
  - Files/modules: `app/api/v1/changes.go`, `app/jobs/prune_api_changes.go`, `routes/schedule.go`, `tests/api_changes_test.go`
  - Refs: SDD 05 FR-05-9, §6.2, §8 (future `since`, exactly 14 days)
  - Depends on: P3.A.3.1
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests -run 'TestAPIChanges'` → `ok`; `since` 14 days + 1 s back → 400 `changes-window-exceeded`; exactly 14 days → 200; future → 400
    - merged item appears as `deleted` with `mergedInto`
  - Notes: cursor on `(changed_at, id)` using the P3.A.1.2 signed cursor. Schedule with the vukem scheduler (<https://vukem.dev/docs/digging-deeper/scheduler>); VF-021: use a cron expression if a typed helper is not documented.

#### P3.A.4 Daily ID exports

- [x] **P3.A.4.1** Add the `jobs.ExportIDs` daily job (07:00 UTC) writing gzip NDJSON per type to the bucket and deleting objects older than 90 days.
  - Files/modules: `app/jobs/export_ids.go`, `routes/schedule.go`, `tests/export_ids_test.go`
  - Refs: SDD 05 FR-05-10, §7.3, NFR-05-6; VF-029; <https://vukem.dev/docs/digging-deeper/storage> "Reading and writing"
  - Depends on: P3.A.3.2
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests -run TestExportIDs` (fake clock + S3 fake) → `ok`: key `exports/games_ids_2026_11_02.json.gz`, lines `{"id","slug","wikidata","popularity","updatedAt"}`, object of day −91 deleted
    - staging: total size of one day's objects ≤ 50 MB
  - Notes: VF-029 workaround: vukem `export.NDJSON(w, r, …)` needs a request, so the job streams rows with `Query.Each` into `gzip.Writer` → `io.Pipe` → `d.Put` (no CSV, so the formula guard is not needed). Never load all rows in memory. `popularity` is `popularity_prior` (sort-only field, SDD 01 §6.1): confirm with the owner before shipping it in a bulk file (A-07 "no bulk redistribution").

- [ ] **P3.A.4.2** Serve `listExports` and `getExport` (`/v1/exports/{type}/{day}` → 302 to a 5-minute presigned URL) for pro and business.
  - Files/modules: `app/api/v1/exports.go`, `tests/api_exports_test.go`
  - Refs: SDD 05 §6.2; <https://vukem.dev/docs/digging-deeper/storage> (`PresignGet(p, ttl)`)
  - Depends on: P3.A.4.1
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests -run TestAPIExports` → free key 403 `plan-required`; pro key 302 with `Location` expiring in 300 s; missing day 404
  - Notes: list from the bucket prefix (no table, SDD 05 §5). The presigned URL must not be cached (`Cache-Control: no-store` on the 302).

#### P3.A.5 Webhooks

- [x] **P3.A.5.1** Add `webhook_endpoints`/`webhook_cursor` and serve `listWebhooks`, `createWebhook` (secret shown once), `deleteWebhook` with plan caps.
  - Files/modules: `database/migrations/<ts>_create_webhook_endpoints.go`, `app/api/v1/webhooks.go`, `app/webhooks/endpoints.go`, `tests/api_webhooks_test.go`, `tests/security_webhooks_test.go`
  - Refs: SDD 05 FR-05-11, §5, §6.2, §8, NFR-05-4; <https://vukem.dev/docs/digging-deeper/webhooks> "Sending webhooks"; <https://vukem.dev/docs/security/building-blocks> "Outgoing webhooks"
  - Depends on: P3.A.3.2
  - Model: Opus
  - Acceptance:
    - `go test ./tests -run 'TestAPIWebhooks|TestSecurity_A10_WebhookSSRF|TestSecurity_A02_WebhookSecretEncrypted'` → `ok`: pro third endpoint → 422; URL resolving to a private IP → 422; secret column never equals the plaintext; webhook calls not counted in quota
  - Notes: secret from `webhook.NewSecret()` (`whsec_…`), encrypted with `app.Crypt`, returned once in the 201 body. `https://` only (DB check). Validation-time SSRF check is a convenience; delivery-time `httpclient.Safe` is the real guard.

- [x] **P3.A.5.2** Add the per-minute fan-out job: read ≤ 10,000 `api_changes` after `webhook_cursor`, coalesce per item, send `item.batch` (≤ 100 events) through `webhook.Sender` with `webhook.SQLDeliveryLog`.
  - Files/modules: `app/webhooks/fanout.go`, `app/webhooks/sender.go`, `routes/schedule.go`, `tests/webhook_delivery_test.go`
  - Refs: SDD 05 §6.2 "Webhook batch", §7.3, FR-05-11; <https://vukem.dev/docs/digging-deeper/webhooks> "The delivery log"; SDD 12 §6 (`api.webhook.deliveries`)
  - Depends on: P3.A.5.1
  - Model: Opus
  - Acceptance:
    - `go test ./tests -run TestWebhookDelivery` → signed batch received and verified with `webhook.Standard`; 410 disables the endpoint; three exhausted deliveries disable it and send one mail (`MAIL_MAILER=array`)
    - `psql "$DATABASE_URL" -c "\d webhook_deliveries"` → table with index on `(endpoint, attempted_at)`
  - Notes: retries per SDD 05 (8 from 30 s) are a sender option; if `SenderOptions` lacks that knob, write `[NOT IN VUKEM DOCS]` and add a `vukem-feedback.md` entry instead of a custom retry loop. Cursor advances in the same transaction that enqueues the jobs. Payload carries ids and field names only.

- [x] **P3.A.5.3** Serve `testWebhook`, prune `webhook_deliveries` older than 30 days, and emit `api.webhook.deliveries{outcome}`.
  - Files/modules: `app/api/v1/webhooks.go`, `app/jobs/prune_webhook_deliveries.go`, `app/observability/metrics.go`, `tests/api_webhooks_test.go`
  - Refs: SDD 05 §6.2; SDD 12 §6 metrics table, §7.6 "API customers"; <https://vukem.dev/docs/digging-deeper/webhooks> "The delivery log" (table grows forever)
  - Depends on: P3.A.5.2
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests -run 'TestAPIWebhookTest|TestPruneDeliveries|TestMetrics_WebhookOutcome'` → `ok`
  - Notes: 30-day delivery retention is an assumption (not in SDD 05); record it in SDD 05 §5 in the same change. A customer-facing delivery log route is not in SDD 05: do not add one without an SDD edit.

#### P3.A.6 Deprecation and contract gate

- [x] **P3.A.6.1** Wire `api.Deprecate` for deprecated operations and document the policy (≥ 6 months before `Sunset`, changelog entry, mail to key owners).
  - Files/modules: `app/api/v1/deprecation.go`, `tests/api_deprecation_test.go`, `docs/runbooks/api-deprecation.md`
  - Refs: SDD 05 §6.1 headers, §6.5 "Versioning"; <https://vukem.dev/docs/rest-apis/overview> "Versioning" (`Deprecation: @<unix>`, `Sunset`, `Link`)
  - Depends on: P3.A.5.3
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests -run TestAPIDeprecationHeaders` → fixture route answers `Deprecation`, `Sunset`, `Link`; a helper call with `sunset - since < 183 days` fails the test
  - Notes: no route is deprecated in v1; the test uses a test-only route. The owner mail is a runbook step, not an automatic job.

- [x] **P3.A.6.2** Run the contract gate on the extended document: export diff, Redocly, Schemathesis, k6 with the new routes, ZAP baseline.
  - Files/modules: `openapi/api-v1.json`, `qa/k6/api-v1.js`, `.github/workflows/ci.yml`
  - Refs: SDD 05 §9, §10, FR-05-15, NFR-05-1
  - Depends on: P3.A.6.1
  - Model: Sonnet
  - Acceptance:
    - `vukem openapi:export -prefix /api/v1 | diff - openapi/api-v1.json` → empty; Redocly lint (pinned) → 0 errors
    - `schemathesis run openapi/api-v1.json --url http://localhost:3000 -H "Authorization: Bearer $HX_TEST_KEY" --checks all` (pinned) → 0 failures
    - `k6 run qa/k6/api-v1.js` → list `p(95)<300`, detail `p(95)<150` at 50 rps, `http_req_failed<0.001`
  - Notes: OpenAPI `servers` = `https://api.hexascore.app/v1` (VF-015 workaround, SDD 05 §6.1). Problem extension members (VF-016) must appear in the schema or Schemathesis will flag them.

#### P3.A.7 Release

- [ ] **P3.A.7.1** Release `v1.1.0` (API `1.1.0`, additive) to staging, then production.
  - Files/modules: `CHANGELOG.md`, `cliff.toml`, `.env` `API_VERSION`, docs changelog entry
  - Refs: ADR 0013; SDD 05 §6.5 "Versioning"; SDD 11 releases
  - Depends on: P3.A.6.2
  - Model: Sonnet
  - Acceptance:
    - `curl -s https://staging.hexascore.app/up` → version `v1.1.0`; `curl -s https://staging-api.hexascore.app/v1/openapi.json | jq -r .info.version` → `1.1.0`
    - staging smoke: `/v1/changes`, `/v1/exports`, `/v1/webhooks` answer per SDD 05 §9 with a test key
  - Notes: tag pushed by the owner (commit/push only when asked). Changelog generated, `[Unreleased]` not hand-edited after release.

### P3.B SDKs and docs

#### P3.B.1 SDK generation

- [ ] **P3.B.1.1** Port the TypeScript SDK (`sdks/ts`, `@hexascore/sdk`) from legacy and regenerate it from `openapi/api-v1.json`.
  - Files/modules: `sdks/ts/package.json`, `sdks/ts/jsr.json`, `sdks/ts/openapi-ts.config.ts`, `sdks/ts/src/`, `sdks/ts/tsconfig.json`, `sdks/ts/README.md`
  - Refs: SDD 05 §6.5, §9; `legacy/sdks/ts/package.json`; `legacy/docs/sdk-publish-npm.md` §2; `legacy/docs/sdk-publish-jsr.md` "Regras do JSR"
  - Depends on: P3.A.7.1
  - Model: Sonnet
  - Acceptance:
    - `cd sdks/ts && bun run sdk:generate && bun run typecheck && bun test` → pass
    - `cd sdks/ts && npm pack --dry-run` → only `dist/**`, `README.md`, `LICENSE`, `package.json`
  - Notes: keep exact pins (`@hey-api/openapi-ts` 0.99.0, `typescript` 6.0.3) and `importFileExtension: ".ts"` for JSR. Version = `API_VERSION` in `package.json` and `jsr.json`. Rewrite the README for the new headers (`X-Quota-*`, problem codes); legacy `RateLimit-*` carried the quota and no longer does.

- [ ] **P3.B.1.2** Port the Python SDK (`sdks/python`, `hexascore`) and regenerate it with openapi-python-client 0.29.1 `--meta none`.
  - Files/modules: `sdks/python/pyproject.toml`, `sdks/python/config.yml`, `sdks/python/hexascore/` (generated), `sdks/python/tests/`
  - Refs: SDD 05 §6.5, §9; `legacy/sdks/python/pyproject.toml`; `legacy/docs/sdk-publish-pypi.md` §2
  - Depends on: P3.B.1.1
  - Model: Sonnet
  - Acceptance:
    - `cd sdks/python && uv run pytest -q` → pass; `uv build && uvx twine@6.2.0 check dist/*` → `PASSED`
  - Notes: `hexascore/` is generated; never edit it by hand. Same version as the TS SDK.

- [ ] **P3.B.1.3** Add the CI job that regenerates both SDKs and fails on diff, then runs typecheck, tests and dry-run packaging.
  - Files/modules: `.github/workflows/ci.yml` (job `sdk`), root `package.json` scripts `sdk:generate`, `sdk:generate:python`
  - Refs: SDD 05 FR-05-15, §10 "SDK smoke"
  - Depends on: P3.B.1.2
  - Model: Sonnet
  - Acceptance:
    - CI job `sdk` green on a clean branch; an edited route without regenerated SDKs → job red with a diff
    - `cd sdks/ts && bunx jsr@0.13.5 publish --dry-run` → `Success Dry run complete`
  - Notes: no network calls in unit tests (mocked fetch / transport); staging smoke is P3.B.2.2.

#### P3.B.2 Publishing

- [ ] **P3.B.2.1** Add the SDK release workflow with trusted publishing (npm with provenance, JSR via OIDC, PyPI via `uv publish`), rehearsed on TestPyPI.
  - Files/modules: `.github/workflows/sdk-release.yml`, `docs/runbooks/sdk-release.md` (English port of the three legacy guides)
  - Refs: SDD 05 §6.5; `legacy/docs/sdk-publish-npm.md` "Trusted publishing", `sdk-publish-jsr.md` §1.4, `sdk-publish-pypi.md` (Trusted Publishing)
  - Depends on: P3.B.1.3
  - Model: Sonnet
  - Acceptance:
    - workflow run against TestPyPI → package visible on test.pypi.org; `npm publish --dry-run` and `jsr publish --dry-run` steps green
  - Notes: owner input required once: npm org `hexascore` with 2FA `auth-and-writes`, JSR scope `@hexascore` linked to the repo, PyPI/TestPyPI accounts and trusted-publisher entries (environment `pypi`). Job needs `permissions: id-token: write`; no long-lived tokens in secrets. Actions pinned by commit SHA.

- [ ] **P3.B.2.2** Publish SDK `1.1.0` to npm, JSR and PyPI and run install smoke tests against staging.
  - Files/modules: `qa/sdk-smoke/ts.ts`, `qa/sdk-smoke/py.py`
  - Refs: SDD 05 §9 (smoke `listItems`, `getItem`, `getUsage`)
  - Depends on: P3.B.2.1
  - Model: Sonnet
  - Acceptance:
    - `bun add @hexascore/sdk@1.1.0` in a scratch project + `bun qa/sdk-smoke/ts.ts` → three calls 200
    - `uv run --no-project --isolated --with hexascore==1.1.0 python qa/sdk-smoke/py.py` → three calls 200
  - Notes: published versions are immutable on all three registries; a mistake means a patch bump. Smoke key is a staging free key.

#### P3.B.3 Docs pages and pricing

- [ ] **P3.B.3.1** Add the developer guides for changes, exports, webhooks (with Standard Webhooks verification in TS and Python) and the errors page anchors for every problem code.
  - Files/modules: `resources/js/pages/Docs/{Changes,Exports,Webhooks,Errors}.tsx`, `resources/lang/{en,pt-BR,es}.json` (chrome only), `routes/web.go` (`/problems/{code}` → `/docs/errors#<code>`)
  - Refs: SDD 05 §6.2 "Problems", §6.5 "Docs"; SDD 09 docs pages; open issue 05-7 (English Scalar inside translated shell)
  - Depends on: P3.B.2.2
  - Model: Sonnet
  - Acceptance:
    - `bun run e2e -- docs.spec.ts` → pages render SSR, axe 0 serious/critical in en, pt-BR, es
    - `curl -sI https://staging.hexascore.app/problems/quota-exhausted` → 301 to `/docs/errors#quota-exhausted`
  - Notes: guides in English (D-06), surrounding chrome translated via the `text-editor` subagent. Code samples use the published SDKs, not raw fetch.

- [ ] **P3.B.3.2** Add `/docs/coverage`, the pricing page coverage statement (D-33) and the API changelog page.
  - Files/modules: `resources/js/pages/Docs/Coverage.tsx`, `resources/js/pages/Pricing/Index.tsx`, `resources/js/pages/Docs/Changelog.tsx`, `app/web/pricing.go`
  - Refs: SDD 05 §6.2 coverage table, open issue 05-10, 05-11; ADR 0006 decision 6; D-33; ADR 0015
  - Depends on: P3.B.3.1, P3.A.2.2
  - Model: Sonnet
  - Acceptance:
    - `bun run e2e -- pricing.spec.ts` → coverage statement and link to `/docs/coverage` visible in three locales; numbers equal `resources/data/api_coverage.json`
  - Notes: prices stay the legacy table (D-33); do not change amounts. The statement must say covers for films, series and games are never in the API (SDD 05 §6.2).

- [ ] **P3.B.3.3** Release `v1.2.0` (SDKs published, docs live).
  - Files/modules: `CHANGELOG.md`
  - Refs: ADR 0013; SDD 05 §6.5
  - Depends on: P3.B.3.2
  - Model: Sonnet
  - Acceptance:
    - `curl -s https://staging.hexascore.app/up` → `v1.2.0`; `npm view @hexascore/sdk version` → `1.1.0`; `pip index versions hexascore` lists `1.1.0`
  - Notes: no API change in this release, so `API_VERSION` stays `1.1.0`.

### P3.C First-party app API `/app/v1`

#### P3.C.1 Route group and tokens

- [x] **P3.C.1.1** Build the `/app/v1` stack: `hx_app_` token guard, `NotBanned`, `auth.Verified`, `auth.Ability`, per-token limits, `Idempotency-Key` on writes, `X-App-Version`/426.
  - Files/modules: `routes/app_v1.go`, `app/appapi/middleware.go`, `app/appapi/version.go`, `tests/app_stack_test.go`, `tests/security_app_test.go`
  - Refs: SDD 05 §6.4, FR-05-13, FR-05-14; SDD 03 §6.2; README abilities convention; <https://vukem.dev/docs/security/api-tokens> "Routing"; <https://vukem.dev/docs/rest-apis/idempotency>
  - Depends on: P2.G
  - Model: Opus
  - Acceptance:
    - `go test ./tests -run 'TestAppMinVersion|TestSecurity_A01_CrossSurfaceCredentials|TestSecurity_A01_AbilityRequired'` → 426 below `APP_MIN_SUPPORTED_VERSION` (logout exempt); `hx_live_` key on `/app/v1` → 401; `hx_app_` token on `/api/v1` → 401; missing ability → 403
    - `go test ./tests -run TestAppIdempotency` → replay on same key, 409 in flight, 422 for a different body
  - Notes: `api.Idempotency` goes after authentication (keys scoped to the caller, vukem idempotency page) on the database cache store. Limits: `ratelimit.ByToken` 120/min reads, 30/min writes, behind the IP limit. Not plan-metered.

- [x] **P3.C.1.2** Serve the token routes: `createAppToken`, `createAppTokenFromGoogle`, `rotateAppToken` (≤ 10/h), `listAppTokens`, `DELETE /tokens/current`, `DELETE /tokens/{id}`.
  - Files/modules: `app/appapi/tokens.go`, `tests/app_tokens_test.go`
  - Refs: SDD 03 §6.2, §7.4; SDD 10 §6.2; ADR 0016 (challenge)
  - Depends on: P3.C.1.1, P1.C
  - Model: Opus
  - Acceptance:
    - `go test ./tests -run 'TestAppTokens'` → 201 with all six abilities; 422 `two-factor-required`; 428 `challenge-required`; 403 `account-banned` with `reason`, `expiresAt`; another user's token id → 404
  - Notes: reuse the P1.C token service; never grant `*`. Tokens never logged (`TestSecurity_A09_TokenNotLogged`).

#### P3.C.2 Catalogue and scoring routes

- [x] **P3.C.2.1** Mount the catalogue reads on `/app/v1` with audience `site`: `/{type}`, `/{type}/{id}`, `/{type}/genres`, `/{type}/trending`, `/entities/{id}`, `/entities/{id}/items`, `/books/{id}/editions`.
  - Files/modules: `app/appapi/catalog.go`, `tests/app_catalog_test.go`
  - Refs: SDD 05 §6.4; SDD 01 §6.1 (`site` audience), §6.2; SDD 10 §6.1
  - Depends on: P3.C.1.2, P3.A.1.3 (trending route only after P3.A.1.4)
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests -run 'TestAPILicenceFilter|TestAppCatalog'` → TMDB/IGDB attributed fields present on `/app/v1`
  - Notes: share handlers with P3.A, switching only the audience; operation names in the `app.v1.` group. `/items/{id}/related` and `/me/for-you` arrive with P6.D, not here.

- [x] **P3.C.2.2** Serve `/search`, `/search/suggest` and `POST /search/events` on `/app/v1`.
  - Files/modules: `app/appapi/search.go`, `tests/app_search_test.go`
  - Refs: SDD 10 §6.1 (Search row, auth optional); SDD 07 §6; SDD 05 §6.4
  - Depends on: P3.C.2.1
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests -run TestAppSearch` → `ok` with and without a token; `q` > 100 chars → 400
  - Notes: SDD 05 §6.4 lists only `/search`; add `/search/suggest` and `/search/events` to that table in the same change (SDD 10 needs them).

- [x] **P3.C.2.3** Serve `/items/{id}/me` and PUT/DELETE `/items/{id}/rating`, `/items/{id}/status`, `/items/{id}/time-to-beat`.
  - Files/modules: `app/appapi/scoring.go`, `tests/app_scoring_test.go`
  - Refs: SDD 05 §6.4; SDD 02 §6 (problems, e.g. `rating-before-release`); SDD 10 §6.1
  - Depends on: P3.C.2.1
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests -run TestAppScoring` → rating PUT twice with the same `Idempotency-Key` stores once; token without `ratings:write` → 403
  - Notes: reuse the SDD 02 services used by the web; no app-only business rules.

#### P3.C.3 Library, profile and notifications

- [x] **P3.C.3.1** Serve `GET/PATCH /me`, `POST /me/export` (202), `POST/DELETE /me/deletion`.
  - Files/modules: `app/appapi/me.go`, `tests/app_me_test.go`
  - Refs: SDD 03 §6.2; SDD 14 §6.3; SDD 05 §6.4; open issue 14-2 (export is POST)
  - Depends on: P3.C.1.2
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests -run TestAppMe` → `nickname-taken` 422; second export within 7 days → 429 `export-cooldown` with `Retry-After`
  - Notes: SDD 10 §6.1 still says `GET /me/export`; fix SDD 10 to POST in the same change.

- [x] **P3.C.3.2** Serve the lists routes and the edition choice: `/me/lists`, `/lists/{id}`, POST `/lists`, PATCH/DELETE `/lists/{id}`, PUT/DELETE `/lists/{id}/items/{itemId}`, PUT `/lists/{id}/order`, PUT/DELETE `/items/{id}/edition`.
  - Files/modules: `app/appapi/lists.go`, `tests/app_lists_test.go`
  - Refs: SDD 14 §6.1, §6.3, §6.5; SDD 05 §6.4
  - Depends on: P3.C.3.1
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests -run TestAppLists` → `list-version-conflict` 409, `order-mismatch` 422, `edition-not-of-work` 422; another user's private list → 404 (`DenyAsNotFound`)
  - Notes: ability conflict: SDD 14 §6.3 puts `/items/{id}/edition` under `status:write`, SDD 05 §6.4 under `lists:write`. Owner/architect input; default to SDD 05 (API owner) and align SDD 14. Events dispatched with `q.DispatchTx`.

- [x] **P3.C.3.3** Serve `/me/library`, `/users/{nickname}`, `/me/notifications`, `POST /me/notifications/read` and `POST /reports`.
  - Files/modules: `app/appapi/library.go`, `app/appapi/notifications.go`, `app/appapi/reports.go`, `tests/app_library_test.go`
  - Refs: SDD 14 §6.3, §6.5; SDD 10 §6.1; open issue 10-1
  - Depends on: P3.C.3.2
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests -run 'TestAppLibrary|TestAppNotifications|TestAppReports'` → private profile returns nickname only; `{ids}` > 100 → 400; `report-self` 403
  - Notes: notifications read from SDD 14 `user_notifications`; `meta.unread` beside `next_cursor`.

#### P3.C.4 Client telemetry intake

- [x] **P3.C.4.1** Serve `POST /app/v1/client-errors` and `POST /app/v1/client-metrics` (anonymous, opt-in, CSRF-exempt), stored 7 days.
  - Files/modules: `database/migrations/<ts>_create_client_reports.go`, `app/appapi/telemetry.go`, `app/jobs/prune_client_reports.go`, `tests/app_telemetry_test.go`
  - Refs: SDD 05 FR-05-18, §6.4 "Client telemetry intake"; SDD 12 FR-12-7, §6.4; open issues 10-2, 09-11
  - Depends on: P3.C.1.1
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests -run TestAppTelemetry` → 204; unknown field 400; 11th error in a minute 429; body > 16 KiB (errors) / 4 KiB (metrics) → 413; prune removes rows at day 8 (fake clock)
  - Notes: SDD 12 §6.4 says 202 and a different body shape and keeps errors 30 days in logs; SDD 05 says 204 and 7 days. Use SDD 05 (contract owner) and fix SDD 12 in the same change. Never store user id, token id or e-mail.

#### P3.C.5 Generated client and release

- [x] **P3.C.5.1** Export `openapi/app-v1.json`, generate the private `@hexascore/app-client` package, and gate it in CI with Schemathesis and k6.
  - Files/modules: `openapi/app-v1.json`, `packages/app-client/package.json`, `packages/app-client/src/api.ts` (generated), `packages/app-client/src/index.test.ts`, `qa/k6/app-v1.js`, `.github/workflows/ci.yml`
  - Refs: SDD 05 §6.5, FR-05-15, §9; ADR 0009 decision 1; <https://vukem.dev/docs/rest-apis/client-generation> "The command"
  - Depends on: P3.C.4.1, P3.C.3.3, P3.C.2.3
  - Model: Sonnet
  - Acceptance:
    - `vukem openapi:export -prefix /app/v1 | diff - openapi/app-v1.json` → empty
    - `vukem codegen:client -prefix /app/v1 -o packages/app-client/src/api.ts && git diff --exit-code packages/app-client` → no diff; `bun run --filter @hexascore/app-client typecheck test` → pass
    - `schemathesis run openapi/app-v1.json --url http://localhost:3000 -H "Authorization: Bearer $HX_APP_TEST_TOKEN" --checks all` → 0 failures; `k6 run qa/k6/app-v1.js` → `p(95)<300`
  - Notes: the app document is not served in production (SDD 05 §6.5). Package is `"private": true`, workspace only. `codegen:client` runs in warmup mode, no secrets needed in CI.

- [ ] **P3.C.5.2** Release `v1.3.0` (first-party API ready for P5).
  - Files/modules: `CHANGELOG.md`
  - Refs: ADR 0013; ADR 0009
  - Depends on: P3.C.5.1
  - Model: Sonnet
  - Acceptance:
    - `curl -s https://staging.hexascore.app/up` → `v1.3.0`; `curl -s -X POST https://staging.hexascore.app/app/v1/tokens -d '{}' -H 'Content-Type: application/json'` → 422 problem+json
  - Notes: `APP_MIN_SUPPORTED_VERSION` starts at `0.0.0` until the first store build (P5).

## Phase 4: community edits

### P4.A Suggestions model and submission

#### P4.A.1 Tables, fields and state machine

- [x] **P4.A.1.1** Create the SDD 04 schema: enums, `suggestions`, `suggestion_events`, `catalog_audit` (append-only trigger, INSERT/SELECT-only grant), `contributor_stats`, `catalog_locks`.
  - Files/modules: `database/migrations/<ts>_create_community_*.go`, `app/models/suggestion.go`, `app/models/catalog_audit.go`, `tests/community_schema_test.go`
  - Refs: SDD 04 §5, FR-04-10, NFR-04-3; SDD 11 (app role grants); SDD 14 §5 (`user_notifications`, not created here); ADR 0010
  - Depends on: P2.G
  - Model: Opus
  - Acceptance:
    - `vukem migrate` then `vukem migrate:status` → the community migrations Ran
    - `psql "$DATABASE_URL" -c "update catalog_audit set after='null' where id=1"` → `ERROR: catalog_audit is append-only`
    - `go test -tags=integration ./app/community/... -run TestAuditAppendOnly` → SQLSTATE `P0001`
  - Notes: SDD 04 §9 says "six tables"; §5 defines five plus `user_notifications`, which SDD 14 owns (open issue 14-1). Do not create a second notifications table. `author_id` and friends are `bigint` here while SDD 14 uses a UUID user id: check `users.id` type from P1.C and align the FK types before writing the migration (VF-006: jsonb and enums as raw DDL).

- [x] **P4.A.1.2** Implement the per-target field allowlist and validators for `changes`, note, locale and source URLs.
  - Files/modules: `app/community/fields.go`, `app/community/validate.go`, `app/community/validate_test.go`
  - Refs: SDD 04 §7.2 (field table, source URL rules), FR-04-2; SDD 01 (allowlist per target)
  - Depends on: P4.A.1.1
  - Model: Sonnet
  - Acceptance:
    - `go test ./app/community/... -run 'TestValidate'` → table tests for every row of §7.2: ISBN-13 checksum, Wikidata `Q\d+`, synopsis 50–2,000, release date 1870 to now + 10 y, ≤ 20 relation ops, ≤ 3 URLs, hexascore.app and shortener URLs refused, empty effective change → 422
  - Notes: validation with vukem `z` schemas (shared with the client as JSON Schema per ADR 0009). Text is NFC-normalised before comparison.

- [x] **P4.A.1.3** Implement the state machine service: allowed transitions, one `suggestion_events` row per transition, `version` optimistic lock.
  - Files/modules: `app/community/state.go`, `app/community/state_test.go`
  - Refs: SDD 04 §7.1, FR-04-4, FR-04-5; ADR 0010
  - Depends on: P4.A.1.2
  - Model: Opus
  - Acceptance:
    - `go test -tags=integration ./app/community/... -run TestTransitions` → every pair outside §7.1 returns 409 `invalid-transition` with `from`, `to` (81 pairs)
    - `-run TestTransitionWritesEvent` → one event per transition in the same transaction; `version` mismatch → 409 `stale-before`
  - Notes: ADR 0010 also names a `stale → expired` path; SDD 04 §7.1 models staleness as a badge plus `expired` on `target_gone`. Follow the SDD. Approve without deferred work records `approved` and `applied` in one transaction.

#### P4.A.2 Quotas and anti-abuse

- [x] **P4.A.2.1** Implement submit: eligibility (verified, ≥ 7 days, not banned or blocked), `contributor_stats` row lock, 3 per UTC day counting withdrawn/expired, `Retry-After` to 00:00 UTC.
  - Files/modules: `app/community/submit.go`, `app/community/quota.go`, `tests/community_quota_test.go`
  - Refs: SDD 04 FR-04-1, FR-04-3, §7.3, §8 (concurrent 3rd/4th, 23:59:59.900); D-31; open issue 04-7
  - Depends on: P4.A.1.3
  - Model: Opus
  - Acceptance:
    - `go test -tags=integration ./app/community/... -run TestQuota` → 3 submits 303, 4th 429 with `type=…/suggestion-quota` and `Retry-After` > 0; after the fake clock reaches 00:00 UTC, 303
    - `-run TestWithdrawStillCounts` → submit, withdraw, 2 submits, 4th 429
    - `-run 'TestEligibility|TestConcurrentSubmit'` → 403 `suggestion-not-eligible` with `eligibleAt`; 20 concurrent submits admit exactly 3
  - Notes: the quota is a DB count, not the vukem limiter (calendar alignment undocumented, open issue 04-7). Refused submits consume nothing.

- [x] **P4.A.2.2** Add duplicate detection under `pg_advisory_xact_lock`, competing flags, field/record locks (423) and the `stale-before` check at submit.
  - Files/modules: `app/community/duplicates.go`, `app/community/locks.go`, `tests/community_duplicates_test.go`
  - Refs: SDD 04 FR-04-6, FR-04-7, §7.2 (stale `before`), §7.3 "Duplicates", "Locks"
  - Depends on: P4.A.2.1
  - Model: Opus
  - Acceptance:
    - `go test -tags=integration ./app/community/... -run 'TestDuplicate|TestFieldLock|TestStaleBeforeAtSubmit'` → 409 `suggestion-duplicate` (with `existingId` only for the author's own), 423 `field-locked` with `lockReason`, 409 `stale-before` with `field`, `current`
  - Notes: expired lock rows are ignored, not deleted. Different `after` values from different authors are allowed and flagged `competing` in `meta`.

- [x] **P4.A.2.3** Implement tiers, cooldowns and the hourly maintenance job (expire drafts at 14 days, refresh tiers, release claims idle > 2 h), plus `community:stats:rebuild`.
  - Files/modules: `app/community/tiers.go`, `app/jobs/community_hourly.go`, `app/commands/community_stats_rebuild.go`, `routes/schedule.go`, `tests/community_tiers_test.go`
  - Refs: SDD 04 §7.3 tier table and cooldown rules, §5 invariants (stats rebuildable); D-15; D-31
  - Depends on: P4.A.2.2
  - Model: Sonnet
  - Acceptance:
    - `go test ./app/community/... -run 'TestTier|TestAccuracy|TestCooldown'` → Laplace accuracy, `blocked` at decided ≥ 20 and accuracy < 0.20, 3 rejections in 24 h → 48 h cooldown, `spam` → 7 days, 429 `suggestion-cooldown`
    - `./app community:stats:rebuild` on a seeded DB → no diff against live `contributor_stats`
  - Notes: command scaffolded with `vukem make:command`. `trusted_eligible` is computed but read by nothing in v1. Expired draft deletes its uploaded cover (§8).

- [x] **P4.A.2.4** Add the honeypot, burst limits, `POST /suggestions/uploads/cover`, and the `community.check_sources` / `community.cover_fetch` jobs over `httpclient.Safe`.
  - Files/modules: `app/community/abuse.go`, `app/community/uploads.go`, `app/jobs/check_sources.go`, `app/jobs/cover_fetch.go`, `tests/security_community_test.go`
  - Refs: SDD 04 §6 jobs, §7.2 covers, §7.3 honeypot and burst, NFR-04-4, §10 security tests; <https://vukem.dev/docs/digging-deeper/uploads>; <https://vukem.dev/docs/security/rate-limiting>
  - Depends on: P4.A.2.3
  - Model: Opus
  - Acceptance:
    - `go test ./... -run 'TestSecurity_Community_A01_SourceURLSSRF|TestSecurity_Community_A04_HoneypotLogged'` → `ok`
    - `go test ./tests -run TestCoverUpload` → AVIF refused, > 5 MiB → 413, short side < 300 px → 422
  - Notes: source checks never block a submission; they store `source_status` in `meta`. Limits: drafts 30/h, submit/withdraw 10/min, uploads 10/h, `RATE_LIMIT_STORE=cache`. Third honeypot hit in 7 days sets tier `blocked`.

#### P4.A.3 Submission UI

- [x] **P4.A.3.1** Build `Community/SuggestEdit` with draft save, the "Suggest an edit" entry points on item, edition and entity pages, and the `community.*` strings in three locales.
  - Files/modules: `resources/js/pages/Community/SuggestEdit.tsx`, `resources/js/components/community/FieldEditor.tsx`, `app/web/community.go`, `resources/lang/{en,pt-BR,es}.json`
  - Refs: SDD 04 §6 route table (GET `/suggest/{target_type}/{id}`, POST/PATCH `/suggestions`), NFR-04-6, NFR-04-7; SDD 09 pages; D-17
  - Depends on: P4.A.2.4
  - Model: Sonnet
  - Acceptance:
    - `bun run typecheck && bun run test -- community` → pass
    - `bun run e2e -- community-draft.spec.ts` → draft saved and restored; quota left shown; locked field disabled with reason; axe 0 serious/critical in en, pt-BR, es
    - i18n CI check → no missing `community.*` key in any locale
  - Notes: Inertia validation errors come back as props (SDD 04 §6). Translations via the `text-editor` subagent. Hidden `website` honeypot field rendered off-screen, not `display:none` only.

- [ ] **P4.A.3.2** Build `Community/SuggestCreate` for create and merge requests (`/suggest/new/{domain}`).
  - Files/modules: `resources/js/pages/Community/SuggestCreate.tsx`, `app/web/community.go`
  - Refs: SDD 04 §2 (create requests, merge reports), §7.2 last row, §7.6
  - Depends on: P4.A.3.1
  - Model: Sonnet
  - Acceptance:
    - `bun run e2e -- community-create.spec.ts` → create draft with required fields of the kind; merge request limited to same type and domain; axe clean
  - Notes: merge here only files a request; execution is P4.B.2.3.

- [ ] **P4.A.3.3** Build `Community/SuggestionStatus` and `Community/MySuggestions` with submit and withdraw actions.
  - Files/modules: `resources/js/pages/Community/{SuggestionStatus,MySuggestions}.tsx`, `app/web/community.go`, `app/policies/suggestion.go`
  - Refs: SDD 04 FR-04-12, §6 (POST `/suggestions/{id}/submit`, `/withdraw`, GET `/suggestions/{id}`, `/me/suggestions`); <https://vukem.dev/docs/security/authorization> (`authz.DenyAsNotFound`)
  - Depends on: P4.A.3.2
  - Model: Sonnet
  - Acceptance:
    - `go test -tags=integration ./app/community/... -run TestStatusPagePolicy` → author and admin 200, others 404
    - `bun run e2e -- community-status.spec.ts` → timeline, reason and text diff visible; 429 shows reset time; axe clean in three locales
  - Notes: policy via `vukem make:policy Suggestion`; fail closed. Diff uses text markers, not colour alone (NFR-04-6).

#### P4.A.4 Release

- [ ] **P4.A.4.1** Release `v1.4.0` to staging with submission routes behind a feature flag that stays off in production until P4.B ships.
  - Files/modules: `CHANGELOG.md`, `app/flags/flags.go`
  - Refs: ADR 0013; <https://vukem.dev/docs/digging-deeper/admin> "Feature flags"
  - Depends on: P4.A.3.3
  - Model: Sonnet
  - Acceptance:
    - `curl -s https://staging.hexascore.app/up` → `v1.4.0`; production `GET /suggest/item/$ID` → 404 while the flag is off
  - Notes: the flag (`community.submissions`) is not named in SDD 04; owner input: confirm or ship P4.A and P4.B to production together. Without it, users could submit with nobody able to moderate.

### P4.B Moderation and application

#### P4.B.1 Admin moderation pages

- [ ] **P4.B.1.1** Add the gate abilities `moderate`, `viewCatalogAudit`, `revertCatalog`, `lockCatalog`, the `community.trusted_moderators` flag, and an app-rendered page shell that reuses the vukem admin layout.
  - Files/modules: `app/admin/admin.go`, `app/admin/abilities.go`, `resources/views/admin/layouts/admin.html` (override), `resources/views/moderation/layout.html`, `routes/admin_moderation.go`, `tests/admin_moderation_gate_test.go`
  - Refs: SDD 04 §6 (pages as `html/template` views in the admin layout, abilities, flag); VF-008; open issue 04-5; <https://vukem.dev/docs/digging-deeper/admin> "Customizing the look", "Feature flags"; <https://vukem.dev/docs/basics/views>
  - Depends on: P4.A.4.1
  - Model: Opus
  - Acceptance:
    - `go test ./tests -run 'TestModerationGate|TestSecurity_Community_A01_PolicyFailsClosed'` → admin 200; user 403; moderator 403 with the flag off and 200 with it on; revert and locks admin-only either way
  - Notes: VF-008 workaround: the panel has no custom pages, so app routes under `/admin/moderation` use the same guard and gate and render templates that wrap the layout copied from `admin.Templates()` via `Config.Views`. Mounting these pages in the panel navigation is [NOT IN VUKEM DOCS]: add a hand-written link in the overridden layout. No inline `<script>` or `style` (strict CSP).

- [ ] **P4.B.1.2** Build the queue view `/admin/moderation` with filters, ageing badges, competing/stale/unreachable badges and claim/release.
  - Files/modules: `app/moderation/queue.go`, `resources/views/moderation/queue.html`, `tests/moderation_queue_test.go`
  - Refs: SDD 04 FR-04-8, §7.4 "Queue", NFR-04-1, NFR-04-2 (green < 12 h, amber 12–24 h, red > 24 h)
  - Depends on: P4.B.1.1
  - Model: Sonnet
  - Acceptance:
    - `go test -tags=integration ./app/moderation/... -run 'TestQueueFilters|TestClaimRelease'` → each filter narrows the result; two admins claiming → second gets 409 `stale-before`
  - Notes: oldest first, keyset pagination on `(submitted_at, id)` (VF-005 workaround). Forms post with CSRF tokens and answer 303.

- [ ] **P4.B.1.3** Build the detail view `/admin/moderation/{id}` with the per-field diff, approve (`confirm_stale`), reject (reason, note), "clone as new draft", and bulk actions (≤ 50, per-id results).
  - Files/modules: `app/moderation/detail.go`, `app/moderation/diff.go`, `app/moderation/bulk.go`, `resources/views/moderation/{detail,bulk_result}.html`, `tests/moderation_detail_test.go`
  - Refs: SDD 04 FR-04-8, §6 routes, §7.1 (clone), §7.4 (diff, bulk), §8 (bulk competing); VF-008
  - Depends on: P4.B.1.2
  - Model: Sonnet
  - Acceptance:
    - `go test ./... -run TestSecurity_Community_A03_DiffEscapesHTML` → `<script>` in an `after` value rendered as text
    - `go test -tags=integration ./app/moderation/... -run 'TestBulk|TestReject'` → 51 ids → 422; bulk approve of two competing suggestions → second `stale-before`; reject without reason → 422
  - Notes: approve calls the P4.B.2.1 service; this microtask can stub it behind the interface and must not write the catalogue itself. Word diff for text, chips for relations, covers side by side. Ban outcome links to SDD 03 `POST /admin/users/{id}/ban`.

#### P4.B.2 Apply, revert and audit

- [ ] **P4.B.2.1** Implement the apply transaction (§7.4 steps 1–6): locks, `confirm_stale`, catalogue write with provenance `community`, one audit row per field, `api_changes`, `DispatchTx` of `catalog.changed` and `community.notify`.
  - Files/modules: `app/community/apply.go`, `app/catalog/write.go`, `tests/community_apply_test.go`
  - Refs: SDD 04 FR-04-9, §7.4, §8 (deadlock retried once); SDD 05 §7.3; P3.A.3.1 writer; <https://vukem.dev/docs/digging-deeper/queue> "Jobs and transactions"
  - Depends on: P4.B.1.3, P3.A.3.1
  - Model: Opus
  - Acceptance:
    - `go test -tags=integration ./app/community/... -run TestApplyAtomic` → injected failure leaves item, suggestion, audit unchanged and no job
    - `-run 'TestApplyProvenance|TestApplyWritesChanges|TestApplyDeadlockRetry'` → `ok`
  - Notes: lock target rows ordered by id (legacy `mergeItems`). Audit stores the real previous value and provenance, not the suggestion's `before`. Cover changes go to `approved` and finish when the SDD 06 cover pipeline runs.

- [ ] **P4.B.2.2** Implement the `catalog.changed` handler: cache purge, search reindex, sitemap `lastmod`, idempotent by `audit_ids`.
  - Files/modules: `app/jobs/catalog_changed.go`, `tests/catalog_changed_test.go`
  - Refs: SDD 04 §6 jobs; SDD 01 (cache), SDD 07 (reindex), SDD 05 §7.3
  - Depends on: P4.B.2.1
  - Model: Sonnet
  - Acceptance:
    - `go test -tags=integration ./app/jobs/... -run TestCatalogChangedIdempotent` → running the job twice purges and reindexes once per audit id
  - Notes: `api_changes` is written in the apply transaction (P4.B.2.1), not here, so a failed job never loses a feed row.

- [ ] **P4.B.2.3** Apply `create` and `merge` suggestions: insert with `source=create` audit rows and unblock `depends_on`; merge through the SDD 01 service with a `*merge*` audit row holding both snapshots.
  - Files/modules: `app/community/apply_create.go`, `app/community/apply_merge.go`, `tests/community_create_merge_test.go`
  - Refs: SDD 04 §7.6, §8 (target merged → open suggestions `expired`); open issue 04-9 (snapshot jsonb); SDD 01 merge
  - Depends on: P4.B.2.2
  - Model: Opus
  - Acceptance:
    - `go test -tags=integration ./app/community/... -run 'TestApplyCreate|TestApplyMerge'` → merge expires open suggestions on the source with `target_gone`; merge row not revertible; typed-title confirmation required
  - Notes: never implement an automatic merge revert (out of scope v1). Merge emits `deleted` with `mergedInto` in `api_changes`.

- [ ] **P4.B.2.4** Implement revert `single`, `cascade`, `force` and the `/admin/catalog-audit` list with the revert form.
  - Files/modules: `app/community/revert.go`, `app/moderation/audit.go`, `resources/views/moderation/audit.html`, `tests/community_revert_test.go`
  - Refs: SDD 04 FR-04-11, §7.5, §6 (`POST /admin/catalog-audit/{id}/revert`)
  - Depends on: P4.B.2.3
  - Model: Opus
  - Acceptance:
    - `go test -tags=integration ./app/community/... -run TestRevert` → `single` with a later change → 409 `revert-conflict` with `laterAuditIds`; `cascade` reverts newest first in one transaction; `force` leaves later suggestions `applied`; reverting a revert → rejected by the unique index
  - Notes: only reason `incorrect` counts against the author. Revert goes through the same write path as apply (provenance restore, `api_changes`, `catalog.changed`).

- [ ] **P4.B.2.5** Make admin-panel catalogue edits write `catalog_audit` (`source=admin`) inside the panel transaction.
  - Files/modules: `app/admin/resources/items.go`, `app/admin/resources/entities.go`, `app/admin/resources/editions.go`, `tests/admin_catalog_audit_test.go`
  - Refs: SDD 04 §7.4 last paragraph, open issue 04-4; VF-010; <https://vukem.dev/docs/digging-deeper/admin> "Resources" (`BeforeSave`)
  - Depends on: P4.B.2.4
  - Model: Sonnet
  - Acceptance:
    - `go test -tags=integration ./app/admin/... -run TestAdminEditWritesAudit` → one audit row per changed field with the correct `before`
  - Notes: VF-010 workaround: `BeforeSave` re-reads the row `FOR UPDATE` in the same transaction to get previous values. `admin.AuditLog` keeps logging panel CRUD separately.

- [ ] **P4.B.2.6** Add the locks pages (`POST/DELETE /admin/locks`) and the `catalog_audit` export (`export.csv`, `export.ndjson`).
  - Files/modules: `app/moderation/locks.go`, `app/moderation/audit_export.go`, `resources/views/moderation/locks.html`, `tests/moderation_locks_test.go`
  - Refs: SDD 04 FR-04-15 (export), §6 routes; <https://vukem.dev/docs/digging-deeper/exports> "Exports", "Streaming and failures"
  - Depends on: P4.B.2.5
  - Model: Sonnet
  - Acceptance:
    - `go test -tags=integration ./app/moderation/... -run 'TestLocks|TestAuditExportStreams'` → 100k-row export streams with flat memory; CSV formula guard on
  - Notes: export runs in the request (helpers need `w, r`), so VF-029 does not apply here.

- [ ] **P4.B.2.7** Add the public SSR "Recent changes" tab (`Catalog/RecentChanges`, 20 per page) and contributor counts on profiles.
  - Files/modules: `resources/js/pages/Catalog/RecentChanges.tsx`, `app/web/recent_changes.go`, `resources/js/pages/User/Profile.tsx`, `tests/recent_changes_test.go`
  - Refs: SDD 04 FR-04-13, FR-04-15, §7.6, NFR-04-5; SDD 14 §6.2 (`User/Profile`)
  - Depends on: P4.B.2.4
  - Model: Sonnet
  - Acceptance:
    - `bun run e2e -- recent-changes.spec.ts` → SSR HTML contains the entries; "HexaScore team" for admins, "deleted user" after deletion; axe clean in three locales
    - k6 on staging → Recent changes p95 < 150 ms server time
  - Notes: tier and accuracy visible only to the user and admins; profiles show counts and approval rate after ≥ 1 decision.

#### P4.B.3 Notifications and system transitions

- [ ] **P4.B.3.1** Implement `community.notify`: `notifications.Send` with dedupe key `suggestion:<id>:<state>` and a localized mail via `SendLaterTx` (opt-out respected).
  - Files/modules: `app/jobs/community_notify.go`, `resources/views/mail/suggestion_{decided,reverted,expired}.html`, `tests/community_notify_test.go`
  - Refs: SDD 04 FR-04-14, §6 jobs, open issue 04-6; SDD 14 §5, §6.5, §7.5; VF-011; <https://vukem.dev/docs/digging-deeper/mail>
  - Depends on: P4.B.2.1
  - Model: Sonnet
  - Acceptance:
    - `go test -tags=integration ./app/jobs/... -run TestCommunityNotify` (`MAIL_MAILER=array`) → one inbox row per state on retry; mail subject in the author's locale; opted-out user gets the inbox row only
  - Notes: VF-011 workaround is SDD 14's `user_notifications`; this job only calls `notifications.Send`. Locale on a non-request context is [NOT VERIFIED] (SDD 03 §6.5); build it from `users.locale`.

- [ ] **P4.B.3.2** Add system transitions: withdraw open suggestions of banned or deleted authors, expire them when the target is deleted or merged, null actor ids on deletion.
  - Files/modules: `app/community/listeners.go`, `tests/community_system_test.go`
  - Refs: SDD 04 §7.1 (system rows), §8, NFR-04-5; SDD 03 (`user.banned`, deletion)
  - Depends on: P4.B.3.1
  - Model: Sonnet
  - Acceptance:
    - `go test -tags=integration ./app/community/... -run 'TestBanWithdraws|TestDeletionNullsActors|TestTargetGoneExpires'` → `ok`; `target_gone` leaves `contributor_stats` unchanged
  - Notes: listeners on the SDD 03 events; audit rows are never deleted.

#### P4.B.4 Metrics and SLA

- [ ] **P4.B.4.1** Build `/admin/moderation/metrics`: queue size, oldest open age, p50/p95 time-to-decision, approval rate over 7 and 30 days.
  - Files/modules: `app/moderation/metrics.go`, `resources/views/moderation/metrics.html`, `tests/moderation_metrics_test.go`
  - Refs: SDD 04 FR-04-16, NFR-04-2; SDD 12 runbook R5
  - Depends on: P4.B.2.4
  - Model: Sonnet
  - Acceptance:
    - `go test -tags=integration ./app/moderation/... -run TestMetricsQuery` → seeded fixture gives the expected percentiles and rates
  - Notes: one SQL query with `percentile_cont`; sortable by age (R5 step 1).

- [ ] **P4.B.4.2** Emit `suggest.queue.*`, `suggest.decisions`, `suggest.sla.breaches`, add alert A9, the Moderation dashboard and the nightly SDD 04 asserts.
  - Files/modules: `app/observability/community_metrics.go`, `ops/grafana/dashboards/moderation.json`, `ops/grafana/alerts/a9_moderation.yaml`, `app/jobs/community_asserts.go`, `tests/metrics_community_test.go`
  - Refs: SDD 12 §6 metrics table, §7.5 A9, §7.6 "Moderation"; SDD 04 NFR-04-2, §10 nightly asserts; VF-019
  - Depends on: P4.B.4.1
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests -run TestMetrics_CommunityGauges` (manual reader) → gauges and counters present with the `decision` attribute
    - `go test -tags=integration ./app/jobs/... -run TestCommunityAsserts` → fails on an `applied` suggestion without audit rows
  - Notes: VF-019 workaround: meter from `Config.Global` plus the upstream OTel API. A9 fires when the oldest open suggestion passes 20 h or p95 age > 24 h, severity warn.

#### P4.B.5 Gate, rehearsal and release

- [ ] **P4.B.5.1** Run the community quality gate: `TestSecurity_Community_*`, ZAP, Playwright `community.spec.ts` end to end with axe in three locales, two-admin claim conflict, and k6 `moderation.js` with 10,000 open suggestions.
  - Files/modules: `qa/e2e/community.spec.ts`, `qa/k6/moderation.js`, `database/seeders/community_load_seeder.go`
  - Refs: SDD 04 §9, §10, NFR-04-1
  - Depends on: P4.B.3.2, P4.B.4.2, P4.B.2.7, P4.B.2.6
  - Model: Sonnet
  - Acceptance:
    - `go test ./... -run TestSecurity_Community` → PASS
    - Playwright `community.spec.ts`: suggest → submit → approve → Recent changes → revert; axe clean in en, pt-BR, es
    - k6 `moderation.js` with 10,000 open suggestions: queue p95 < 500 ms, submit p95 < 300 ms
  - Notes: seeder via `vukem make:seeder`; never run it against production.

- [ ] **P4.B.5.2** Rehearse on staging: process 100 synthetic suggestions (mixed kinds, approvals, rejections, a revert, a lock, a competing pair) and confirm the SLA dashboard and notifications.
  - Files/modules: `app/commands/community_rehearsal.go`, `docs/runbooks/community-rehearsal.md`
  - Refs: ROADMAP P4 exit criteria; SDD 04 NFR-04-2, §10 nightly asserts; SDD 12 §7.6 "Moderation", R5
  - Depends on: P4.B.5.1
  - Model: Opus
  - Acceptance:
    - `./app community:rehearsal --count 100` on staging creates 100 suggestions from synthetic accounts marked as test users
    - after moderation: `psql "$STAGING_DATABASE_URL" -c "select status, count(*) from suggestions where meta->>'rehearsal' = 'true' group by 1"` → no `submitted`/`in_review` rows; nightly asserts green; Moderation dashboard shows decisions and p95 time-to-decision < 24 h
  - Notes: owner input: who moderates (owner alone or with a helper) and when. Rehearsal rows are reverted or deleted afterwards only through revert, never by editing `catalog_audit` (append-only). The command flag names are app code (`vukem make:command`), not vukem flags.

- [ ] **P4.B.5.3** Release `v1.5.0` and turn the community submissions flag on in production.
  - Files/modules: `CHANGELOG.md`
  - Refs: ADR 0013; ADR 0010; P4.A.4.1 flag
  - Depends on: P4.B.5.2
  - Model: Sonnet
  - Acceptance:
    - `curl -s https://hexascore.app/up` → `v1.5.0`; `/admin/_flags` shows `community.submissions` on; first production suggestion reaches the queue
  - Notes: alert A9 must be active before the flag goes on.

## Phase 5: mobile

Dependencies on earlier phases are written as microtask IDs where those phases are expanded (`P3.C.5.2`) and as group IDs otherwise (`P1.A`, `P2.F`); read a group ID as "the last microtask of that group is done". Release numbering follows ADR 0013 (one minor per task group), continuing after Phase 4's `v1.5.0`: P5.A `v1.6.0`, P5.B `v1.7.0`, P5.C `v1.8.0`, P6.A `v1.9.0`, P6.B `v1.10.0`, P6.C `v1.11.0`, P6.D `v1.12.0`. The Android app has its own SemVer, tagged `mobile-vX.Y.Z` (SDD 10 §7.4); the first store build is `1.0.0`. The phase map in `ROADMAP.md` still shows the old numbers; renumber it when this file is merged.

Mobile conventions for every P5 microtask: Expo managed (CNG), no `android/` folder committed, TypeScript strict, Bun workspaces, exact pinned versions. Screens call the server only through `@hexascore/app-client` (P3.C.5.1) wrapped in `apps/mobile/src/api/` (FR-10-1). An Expo SDK module the SDD names is approved. Any other npm package needs the owner's yes and an ADR (AGENTS.md). Expo and React Native facts come from the pages SDD 10 cites as [E1]–[E18].

### P5.A Mobile foundation

#### P5.A.1 Expo project and navigation

- [x] **P5.A.1.1** Scaffold `apps/mobile` as an Expo managed app in the Bun workspace: TypeScript strict, Hermes, Biome lint with a rule that bans `fetch(` outside `src/api/`, Jest with the `jest-expo` preset and React Native Testing Library.
  - Files/modules: `apps/mobile/package.json`, `apps/mobile/app.config.ts`, `apps/mobile/tsconfig.json`, `apps/mobile/biome.json` (extends the root config), `apps/mobile/jest.config.ts`, `apps/mobile/src/hermes.test.ts`, root `package.json` workspaces, `.github/workflows/ci.yml` (mobile job: lint, typecheck, test)
  - Refs: SDD 10 §7.1, §10, FR-10-1, NFR-10-4; ADR 0009; D-12; [E3], [E4]
  - Depends on: P3.C.5.2, P1.A
  - Model: Opus
  - Acceptance:
    - `bun run --cwd apps/mobile typecheck && bun run --cwd apps/mobile lint` → 0 errors; a scratch file with `fetch(` in `src/screens/` → lint error (then removed)
    - `bun run --cwd apps/mobile test` → green, including the test that asserts `global.HermesInternal` is defined
  - Notes: pick the latest stable Expo SDK at implementation time that targets Android API 36 (Play rule since 2026-08-31, [E14]). Write the SDK version and the reason in the PR. `android.package = app.hexascore`. Vitest on React Native is [NOT VERIFIED] (SDD 10 §10), so keep Jest. Use no Expo template that adds analytics or other extra packages.

- [x] **P5.A.1.2** Lay out the Expo Router tree from SDD 10 §6.1 with typed routes, the tab layout, placeholder screens and the `src/platform/` boundary (network state, store link).
  - Files/modules: `apps/mobile/app/_layout.tsx`, `app/(auth)/login.tsx`, `app/(tabs)/_layout.tsx`, `app/(tabs)/index.tsx`, `app/(tabs)/[type].tsx`, `app/(tabs)/search.tsx`, `app/(tabs)/lists.tsx`, `app/(tabs)/profile.tsx`, `app/[type]/[slug].tsx`, `app/entities/[id].tsx`, `app/lists/[id].tsx`, `app/update-required.tsx`, `src/platform/network.ts`, `src/platform/store.ts`, `src/navigation.test.tsx`
  - Refs: SDD 10 §6.1, §7.1 (Expo Router, paths mirror the web, D-17), §2 (Android-only APIs only behind `src/platform/`); [E5]
  - Depends on: P5.A.1.1
  - Model: Sonnet
  - Acceptance:
    - `bun run --cwd apps/mobile test -- src/navigation` → each §6.1 route renders its placeholder; `[type]` rejects anything outside `games|movies|series|books`
    - `bun run --cwd apps/mobile typecheck` → typed `href` errors on an unknown route
  - Notes: route segments use the web plural slugs (`games`, `movies`, `series`, `books`) so App Links need no mapping (P5.B.5.1). Do not add screens beyond §6.1 (no billing, keys or admin, D-12).

#### P5.A.2 Shared tokens and i18n

- [x] **P5.A.2.1** Consume `packages/tokens` in a theme module (light/dark colours, spacing, type scale, category accents, score bands) and `packages/i18n` catalogues (build-time copy of the bundled namespaces, locale matcher, ICU formatting), with an in-app override persisted locally.
  - Files/modules: `apps/mobile/src/theme/index.ts`, `src/theme/scoreBand.ts`, `src/theme/scoreBand.test.ts`, `apps/mobile/scripts/copy-catalogues.ts`, `src/i18n/index.ts`, `src/i18n/locale.ts`, `src/i18n/locale.test.ts`, `src/i18n/keys.test.ts`
  - Refs: SDD 10 §7.2, FR-10-6, §8 (locale cases), §4 (AA contrast); ADR 0009; SDD 09 §7.1; open issues 09-5, 10-4; [E10]
  - Depends on: P5.A.1.2
  - Model: Sonnet
  - Acceptance:
    - `bun run --cwd apps/mobile test -- src/theme src/i18n` → bands `<5.0` low, `5.0–7.4` mid, `≥7.5` high, `null` none; `pt-PT → pt-BR`, `es-MX → es`, `fr → en`; every key in the `en` catalogue exists in `pt-BR` and `es` for the bundled namespaces
    - `bun run --cwd apps/mobile test -- src/theme/contrast` → every text/background token pair ≥ 4.5:1 in both themes
  - Notes: bundled namespaces are `common`, `auth`, `account`, `catalog`, `categories`, `item`, `lists`, `ratings`, `search`, `errors`. Never bundle `api`, `billing`, `docs`, `openapi`, `contact`, `legal` or `social`. Use the ICU library SDD 09 picked; do not choose another one. Tokens are plain objects fed to `StyleSheet.create`, with no shared components (ADR 0009). The server-side `PATCH /me {locale}` call is wired in P5.B.4.2.

#### P5.A.3 API client and authentication

- [x] **P5.A.3.1** Wrap `@hexascore/app-client` in `src/api/`: base URL from `EXPO_PUBLIC_API_BASE_URL`, headers (`Authorization`, `X-App-Version`, `Accept-Language`, UUID v7 `Idempotency-Key` on writes), mapping of RFC 9457 `code` to `errors.<code>`, and detection of `X-App-Min-Version` and 426.
  - Files/modules: `apps/mobile/src/api/client.ts`, `src/api/problems.ts`, `src/api/version.ts`, `src/api/client.test.ts`
  - Refs: SDD 10 §6.1 "Headers", §7.4, FR-10-1, FR-10-5; SDD 05 §6.4; <https://vukem.dev/docs/rest-apis/client-generation> "The command"; <https://vukem.dev/docs/rest-apis/idempotency>
  - Depends on: P5.A.2.1
  - Model: Opus
  - Acceptance:
    - `bun run --cwd apps/mobile test -- src/api` → unknown code maps to `errors.unexpected` with `request_id`; a response whose `X-App-Min-Version` is above the running version sets `upgradeRequired`; a write without an idempotency key is impossible at the type level
    - `bun run --cwd apps/mobile typecheck` → green against the generated `packages/app-client/src/api.ts` (no hand-written request types)
  - Notes: never edit the generated client by hand. Regenerate it with `vukem codegen:client -prefix /app/v1 -o packages/app-client/src/api.ts`, which runs in warmup mode. `EXPO_PUBLIC_*` holds only public values: the API base URL and the Google client id (SDD 10 §4).

- [ ] **P5.A.3.2** Store the token in SecureStore and run the session state machine: 401 wipes the token and cache, 403 `account-banned` shows the reason and end date, 403 `email-not-verified` shows a banner, and `deletionScheduledAt` routes to the cancel screen.
  - Files/modules: `apps/mobile/src/auth/tokenStore.ts`, `src/auth/session.ts`, `src/auth/session.test.ts`, `src/auth/noTokenInCache.test.ts`
  - Refs: SDD 10 §5 (SecureStore row), §6.2, §8 (token expired in background), FR-10-2; SDD 03 §7.4; [E1]
  - Depends on: P5.A.3.1
  - Model: Opus
  - Acceptance:
    - `bun run --cwd apps/mobile test -- src/auth` → green with ≥ 80 % coverage on `src/auth`; a test serialises the persisted query cache and asserts no `hx_app_` substring
    - `bun run --cwd apps/mobile test -- src/auth/session -t "401 wipes"` → token, cache and queue for that user are cleared
  - Notes: store only `token`, `tokenId` and `expiresAt` (≈ 100 B, below the ~2048 B SecureStore warning). Tokens must never reach logs, the cache or crash reports.

- [ ] **P5.A.3.3** Build the login screen for email/password with the `two-factor-required` branch (TOTP or recovery code) and the 428 `challenge-required` branch, which opens the browser. Signup and recovery open the web pages. `deviceName` comes from the Expo Device API.
  - Files/modules: `apps/mobile/app/(auth)/login.tsx`, `src/auth/login.ts`, `src/auth/deviceName.ts`, `src/auth/login.test.tsx`, `.maestro/smoke/login.yaml`
  - Refs: SDD 10 §2, §6.2 (Issue, Device name rows), FR-10-2; SDD 03 §2, §6.1; ADR 0016 (challenge); open issue 10-10
  - Depends on: P5.A.3.2, P3.C.1.2
  - Model: Sonnet
  - Acceptance:
    - `bun run --cwd apps/mobile test -- src/auth/login` → 422 `two-factor-required` shows the code field; a recovery code is accepted; 428 opens the web challenge; default device name is `HexaScore · <model>`, editable, 1–64 chars
    - `maestro test .maestro/smoke/login.yaml` (staging, test account) → PASSED
  - Notes: open issue 10-10 is closed here. Cite the Expo Device docs page in SDD 10 §6.2 and mark the issue decided in `open-issues.md`. Opening the browser needs an Expo SDK module: check SDD 10 or ask the owner before you add `expo-web-browser`. `Linking.openURL` is the fallback that adds no dependency.

- [ ] **P5.A.3.4** Add Google sign-in that posts the ID token to `/tokens/google`, with the same 2FA branch, a 409 `email-unverified-at-provider` message and silent cancel.
  - Files/modules: `apps/mobile/src/auth/google.ts`, `src/auth/google.test.ts`, `app.config.ts` (Google client id), `docs/adrs/00xx-mobile-google-signin.md`
  - Refs: SDD 10 §6.2 (Issue Google row), §8 (cancelled), FR-10-2; SDD 03 §6.2
  - Depends on: P5.A.3.3
  - Model: Opus
  - Acceptance:
    - `bun run --cwd apps/mobile test -- src/auth/google` → cancel leaves the screen without an error; 409 shows `errors.email-unverified-at-provider`
    - Manual on the reference device (preview build): Google account sign-in → token stored, home loads
  - Notes: **owner input needed.** SDD 10 does not name the Google sign-in library, and every candidate is a new dependency. Write the ADR (options, licence, native config plugin under CNG) and get the owner's yes before you code. Only the ID token goes to the server. No access token is stored.

- [ ] **P5.A.3.5** Rotate the token on foreground (`expiresAt − now < 15 d`, at most once per 24 h), log out with `DELETE /tokens/current` (wiping locally even when the call fails), and show the blocking update screen on 426 or a higher min version. Logout stays available there.
  - Files/modules: `apps/mobile/src/auth/rotate.ts`, `src/auth/rotate.test.ts`, `src/auth/logout.ts`, `app/update-required.tsx`, `src/platform/store.ts` (`market://details?id=app.hexascore`)
  - Refs: SDD 10 §6.2 (Rotate, Logout rows), §7.4, FR-10-5, §8 (rotate response lost; 426 with a pending queue)
  - Depends on: P5.A.3.2
  - Model: Sonnet
  - Acceptance:
    - `bun run --cwd apps/mobile test -- src/auth/rotate src/auth/logout` → no rotation at 16 d; one rotation at 14 d; a second foreground within 24 h does nothing; logout with the network down still wipes the token
    - `curl -si -H "X-App-Version: 0.0.1" -H "Authorization: Bearer $T" https://staging.hexascore.app/app/v1/me` → 426 `app-upgrade-required` (SDD 10 §9), and the app shows the update screen with logout
  - Notes: a lost rotate response leads to a 401 and a new login. This is accepted (SDD 10 §8); do not add retry logic that reuses the old token.

#### P5.A.4 Query cache and offline queue

- [ ] **P5.A.4.1** Configure TanStack Query with the legacy defaults (`staleTime` 30 s, `retry` 1), `focusManager` and `onlineManager` through `src/platform/`, and `persistQueryClient` on Expo SQLite (10 MB, `maxAge` 24 h, `buster` = app version + `formulaVersion`, read queries only).
  - Files/modules: `apps/mobile/src/query/client.ts`, `src/query/persister.ts`, `src/query/persister.test.ts`, `src/platform/network.ts`
  - Refs: SDD 10 §5 (persisted cache row), §7.1, §7.3 "Reads", NFR-10-6; D-22; open issue 10-6; [E2]
  - Depends on: P5.A.3.2
  - Model: Opus
  - Acceptance:
    - `bun run --cwd apps/mobile test -- src/query` → a mutation result or a problem body is never persisted; data older than 24 h is discarded; a `formulaVersion` change busts the cache; the persisted size stays ≤ 10 MB with a 2,000-item fixture
  - Notes: Expo SQLite is approved (D-22); no other storage library. Keep the dehydrate filter an allowlist of query keys, not a denylist.

- [ ] **P5.A.4.2** Implement the offline mutation queue in `src/queue/`: persisted entries `{id, op, itemId, payload, idempotencyKey, createdAt, attempts}` (≤ 50, 72 h), replacement per `(itemId, op)` with a fresh key, FIFO drain on reconnect/foreground with 2 s/10 s/60 s backoff, the §7.3 response table, user binding, and pause on 426.
  - Files/modules: `apps/mobile/src/queue/store.ts`, `src/queue/drain.ts`, `src/queue/policy.ts`, `src/queue/queue.test.ts`
  - Refs: SDD 10 §5 (mutation queue row), §7.3, §8, FR-10-4; <https://vukem.dev/docs/rest-apis/idempotency> (409 in flight, 422 reused key)
  - Depends on: P5.A.4.1
  - Model: Opus
  - Acceptance:
    - `bun run --cwd apps/mobile test -- src/queue --coverage` → ≥ 80 % on `src/queue`; 409 in flight retries; 409 `rating-not-released` and 422 roll back; 429 `item-ratings-limited` rolls back and shows `Retry-After`; 404 drops the entry; 5xx after 5 attempts or 72 h drops it with a toast
    - `bun run --cwd apps/mobile test -- src/queue -t "replace"` → a newer write for the same `(itemId, op)` replaces the queued one with a new key
  - Notes: the server wins on conflict. Never reuse a key with a new body (that returns 422). The queue replays only for the user who created it. Lists create, rename, delete and reorder are not queued (SDD 10 §7.3).

- [ ] **P5.A.4.3** Add the offline banner and the crash-report opt-in (first launch, default off, revocable): a JS handler and an error boundary post `{appVersion, versionCode, androidApi, screen, message, stack ≤ 8 KB}` to `POST /app/v1/client-errors`.
  - Files/modules: `apps/mobile/src/platform/OfflineBanner.tsx`, `src/telemetry/consent.ts`, `src/telemetry/crash.ts`, `src/telemetry/crash.test.ts`, `app/_layout.tsx`
  - Refs: SDD 10 §7.7, §5 (preferences row); SDD 05 §6.4; SDD 12 §6.4; open issue 10-2
  - Depends on: P5.A.4.1, P3.C.4.1
  - Model: Sonnet
  - Acceptance:
    - `bun run --cwd apps/mobile test -- src/telemetry` → nothing is sent without consent; the payload has no user id, token or e-mail; the stack is truncated to 8 KB
  - Notes: P3.C.4.1 settled the response at 204 (SDD 05 owns the contract). Native crash capture (Sentry or similar) needs an ADR and is out of scope here. No third-party analytics.

#### P5.A.5 Release

- [ ] **P5.A.5.1** Release `v1.6.0`: mobile foundation merged, with a `development` and a `preview` build installable from EAS on the reference device.
  - Files/modules: `CHANGELOG.md`, `apps/mobile/app.config.ts` (`version` `0.1.0`), `docs/PROGRESS.md`
  - Refs: ADR 0013; SDD 10 §7.4, §7.8
  - Depends on: P5.A.4.3, P5.A.3.5, P5.A.3.4
  - Model: Sonnet
  - Acceptance:
    - `curl -s https://staging.hexascore.app/up` → version `v1.6.0`
    - preview APK on the reference device → login (password and Google) against staging, kill and reopen offline → cached home placeholder renders, offline banner shown
  - Notes: no store submission yet. The owner pushes the tag. Note in the release that the app version (`mobile-v0.1.0`) is separate from the API tag.

### P5.B Mobile features

#### P5.B.1 Browse

- [ ] **P5.B.1.1** Build the home screen with four "Top 10 trending (30 days)" sections from `GET /{type}/trending`, plus the shared listing row (category label with icon, 44 × 64 dp `coverThumbUrl` via `expo-image` `memory-disk`, tuned `FlatList`).
  - Files/modules: `apps/mobile/app/(tabs)/index.tsx`, `src/components/ItemRow.tsx`, `src/components/CategorySection.tsx`, `src/components/ItemRow.test.tsx`, `src/screens/home.test.tsx`
  - Refs: SDD 10 §6.1 (Home row), §7.1 (FlatList), §7.3 "Images", FR-10-10, NFR-10-2; SDD 08 §6 (title rule, "Popular" chip for `basis:"prior"`), FR-08-5, FR-08-11; D-02; [E6], [E7]
  - Depends on: P5.A.5.1, P3.C.2.1
  - Model: Sonnet
  - Acceptance:
    - `bun run --cwd apps/mobile test -- src/components src/screens/home` → sections in `ITEM_TYPES` order; empty type hidden; a prior entry shows the "Popular" chip; all-prior changes the title to "Popular {type}"; the score is always shown next to its band colour
  - Notes: the trending route on `/app/v1` exists only after P3.A.1.4, which waits for P6.C. Until then, show the empty state, never fabricated data. Raise the ordering question with the owner if P6.C is not pulled forward. FlashList only through an ADR if NFR-10-2 fails.

- [ ] **P5.B.1.2** Build the section listing `(tabs)/[type]` with sort, genre filter (`/{type}/genres`) and cursor-based infinite scroll.
  - Files/modules: `apps/mobile/app/(tabs)/[type].tsx`, `src/screens/listing.test.tsx`, `src/query/keys.ts`
  - Refs: SDD 10 §6.1 (Section row), FR-10-10, NFR-10-2; SDD 05 §6.2 "Cursors"
  - Depends on: P5.B.1.1
  - Model: Sonnet
  - Acceptance:
    - `bun run --cwd apps/mobile test -- src/screens/listing` → `next_cursor` drives the next page; changing sort or genre resets the cursor; the end of the list stops fetching
  - Notes: treat the cursor as opaque (signed server-side, VF-005). Never build one on the client.

#### P5.B.2 Item, rating and entity

- [ ] **P5.B.2.1** Build the item screen `[type]/[slug]`: header with cover (600 px scaled), score with band, `ratingCount`, confidence or "–", categories, entities grouped by role, book editions, TMDB/IGDB attribution, and `/items/{id}/me` state.
  - Files/modules: `apps/mobile/app/[type]/[slug].tsx`, `src/screens/item/ItemHeader.tsx`, `src/screens/item/Attribution.tsx`, `src/screens/item/item.test.tsx`
  - Refs: SDD 10 §6.1 (Item row), FR-10-9, §8 (merged item, empty groups), NFR-10-3; SDD 02 §6 (public read fields, FR-02-6); SDD 01 §6.1 (`site` audience); open issue 01-6
  - Depends on: P5.B.1.2
  - Model: Sonnet
  - Acceptance:
    - `bun run --cwd apps/mobile test -- src/screens/item` → a merged id follows the 301 and rewrites the cache key; empty edition and entity groups are hidden; the attribution block is shown only when a TMDB/IGDB field is present
  - Notes: the "similar" and "cross-media" blocks need `/items/{id}/related` (P6.D.1.3). Leave them out, not as empty placeholders. Adding them later is a new feature, so it needs a store release, not OTA (SDD 10 §7.8).

- [ ] **P5.B.2.2** Build the star widget `src/rating/StarInput`: ten stars with half-star taps (0.5–10.0), re-tap is a no-op, selection haptic, one `adjustable` accessibility element with doubled integer `now` and ±0.5 increments, 32 dp stars with 6 dp hit slop, wrapping below 360 dp.
  - Files/modules: `apps/mobile/src/rating/StarInput.tsx`, `src/rating/StarInput.test.tsx`
  - Refs: SDD 10 §7.5, FR-10-3, §4 (targets ≥ 44 dp, 200 % font scale); ADR 0005; SDD 09 (web widget); [E9], [E11]
  - Depends on: P5.A.5.1
  - Model: Sonnet
  - Acceptance:
    - `bun run --cwd apps/mobile test -- src/rating --coverage` → ≥ 80 % on `src/rating`; a tap on the left half of star 7 gives 6.5; the accessibility increment from 6.5 gives 7.0; `accessibilityValue.now` is 13 for 6.5; `Haptics.selectionAsync` is called once per change
  - Notes: no drag gesture in v1. "Remove rating" is a separate control (P5.B.2.3).

- [ ] **P5.B.2.3** Build the rate sheet: six category widgets for the item's type, submit disabled until all are set, a single PUT with all keys through the offline queue, a remove-rating action, and rollback toasts.
  - Files/modules: `apps/mobile/src/rating/RateSheet.tsx`, `src/rating/useRateMutation.ts`, `src/rating/RateSheet.test.tsx`, `.maestro/smoke/rate-item.yaml`
  - Refs: SDD 10 §6.1 (Rate sheet row), §7.3, §7.5 "Rate sheet", FR-10-4; SDD 02 §6 (PUT body, 409 `rating-not-released`, 429 limits); A-06
  - Depends on: P5.B.2.1, P5.B.2.2, P5.A.4.2
  - Model: Sonnet
  - Acceptance:
    - `bun run --cwd apps/mobile test -- src/rating/RateSheet` → the optimistic aggregate updates at once and is replaced by the server body on 2xx; 409 before release rolls back
    - `maestro test .maestro/smoke/rate-item.yaml` (staging) → PASSED
  - Notes: input is in stars (the server multiplies by 10). The category keys are the stable A-06 keys from the generated client; never hard-code labels.

- [ ] **P5.B.2.4** Build the status and time-to-beat sheet: PUT/DELETE status, and PUT/DELETE time-to-beat in 0.5 h steps for games only, both queued offline.
  - Files/modules: `apps/mobile/src/status/StatusSheet.tsx`, `src/status/TimeToBeatSheet.tsx`, `src/status/status.test.tsx`
  - Refs: SDD 10 §6.1 (Status/time-to-beat row), §7.3; SDD 02 §6, §7.7 (409 `status-not-completed`, 422 `not-a-game`)
  - Depends on: P5.B.2.3
  - Model: Sonnet
  - Acceptance:
    - `bun run --cwd apps/mobile test -- src/status` → time-to-beat is hidden for non-games; values outside 0.5–10,000 h are rejected client-side; 409 `status-not-completed` rolls back
  - Notes: keep the validation messages from the `errors` catalogue. Do not duplicate server rules beyond input bounds.

- [ ] **P5.B.2.5** Build the entity screen `entities/[id]` with items grouped by type and role (`/entities/{id}/items?type=&role=`, cursor).
  - Files/modules: `apps/mobile/app/entities/[id].tsx`, `src/screens/entity.test.tsx`
  - Refs: SDD 10 §6.1 (Entity row), §8 (empty groups hidden), FR-10-10; SDD 01 §6.2
  - Depends on: P5.B.2.1
  - Model: Sonnet
  - Acceptance:
    - `bun run --cwd apps/mobile test -- src/screens/entity` → groups in `ITEM_TYPES` order; empty groups hidden; tapping a row opens the item route

#### P5.B.3 Search

- [ ] **P5.B.3.1** Build the search tab: debounced `/search/suggest`, full `/search` results grouped by type, and `POST /search/events` for `search.performed` and `search.clicked`. Auth is optional.
  - Files/modules: `apps/mobile/app/(tabs)/search.tsx`, `src/search/useSuggest.ts`, `src/search/events.ts`, `src/search/search.test.tsx`
  - Refs: SDD 10 §6.1 (Search row), §7.7 (only these two client events); SDD 07 §6, §7.8; SDD 08 §6 (`SearchResultClicked` feeds `search_clicks`)
  - Depends on: P5.B.1.2, P3.C.2.2
  - Model: Sonnet
  - Acceptance:
    - `bun run --cwd apps/mobile test -- src/search` → one suggest call per debounce window; `q` > 100 chars is cut before sending; the click event carries item id and position, not the raw query (the server hashes it)
  - Notes: search history is "in-app search history (30 days)" in the data safety answers (P5.C.2.1). Keep it consistent with what the server stores.

#### P5.B.4 Lists, profile and notifications

- [ ] **P5.B.4.1** Build the lists tab, the list detail and the add-to-list sheet: create, rename, delete and reorder only while online; single add/remove queued offline; handle `list-version-conflict` and `order-mismatch`.
  - Files/modules: `apps/mobile/app/(tabs)/lists.tsx`, `app/lists/[id].tsx`, `src/lists/AddToListSheet.tsx`, `src/lists/lists.test.tsx`, `.maestro/smoke/add-to-list.yaml`
  - Refs: SDD 10 §6.1 (Lists row), §7.3 "Lists"; SDD 14 §6.1, §6.3; P3.C.3.2
  - Depends on: P5.B.2.1, P5.A.4.2
  - Model: Sonnet
  - Acceptance:
    - `bun run --cwd apps/mobile test -- src/lists` → offline create/rename/delete/reorder buttons are disabled; offline add is queued and drains on reconnect; 409 `list-version-conflict` refetches and shows a toast
    - `maestro test .maestro/smoke/add-to-list.yaml` (staging) → PASSED

- [ ] **P5.B.4.2** Build the profile tab: `GET/PATCH /me` (nickname, locale override, `personalizationOptOut`), devices (`GET /tokens`, `DELETE /tokens/{id}`), crash-report consent toggle, theme, and logout.
  - Files/modules: `apps/mobile/app/(tabs)/profile.tsx`, `src/profile/Settings.tsx`, `src/profile/Devices.tsx`, `src/profile/profile.test.tsx`
  - Refs: SDD 10 §6.1 (Profile row), §6.2 (Revoke a device), FR-10-6; SDD 03 §5 (`personalization_opt_out`), FR-03-19; SDD 08 FR-08-10; D-29
  - Depends on: P5.A.5.1, P3.C.3.1
  - Model: Sonnet
  - Acceptance:
    - `bun run --cwd apps/mobile test -- src/profile` → 422 `nickname-taken` shown inline; revoking the current device logs out; changing the locale calls `PATCH /me {locale}` and re-renders without a restart
  - Notes: the opt-out toggle ships now even though "For you" arrives in P6.D. The server field already exists (SDD 03), and the toggle must be present before personalisation turns on (D-29).

- [ ] **P5.B.4.3** Add account actions and notifications: export (`POST /me/export`, 202, cooldown 429), delete account after a 10 s countdown (`POST /me/deletion`), the deletion-scheduled screen linking to web `/login`, the About screen with TMDB/IGDB attribution, and the notifications list with mark-as-read and an unread badge.
  - Files/modules: `apps/mobile/src/profile/DeleteAccount.tsx`, `src/profile/Export.tsx`, `app/about.tsx`, `app/notifications.tsx`, `src/notifications/useUnread.ts`, `src/profile/account.test.tsx`, `.maestro/flows/delete-account.yaml`
  - Refs: SDD 10 §6.1 (Profile, Notifications rows), §6.2 (Deletion scheduled), FR-10-8, FR-10-9; SDD 03 FR-03-15, §8; SDD 14 §6.3, §7.4; open issues 10-1, 14-2
  - Depends on: P5.B.4.2, P3.C.3.3
  - Model: Sonnet
  - Acceptance:
    - `bun run --cwd apps/mobile test -- src/profile/account src/notifications` → the delete button is enabled only after 10 s; second export within 7 days shows `Retry-After`; mark-read sends ≤ 100 ids and clears the badge
    - `maestro test .maestro/flows/delete-account.yaml` (staging throwaway account) → PASSED, and the account shows `deletionScheduledAt`
  - Notes: no push notifications (out of scope, SDD 10 §2). Read `meta.unread` from the list response; do not poll.

#### P5.B.5 Deep links and end-to-end flows

- [ ] **P5.B.5.1** Serve `GET /.well-known/assetlinks.json` from the Go app and declare verified App Links in the Expo config (`autoVerify: true`, host `hexascore.app`, prefixes `/games`, `/movies`, `/series`, `/books`, `/entities`, `/lists`, `/pt-br`, `/es`).
  - Files/modules: `routes/web.go` (well-known route), `app/wellknown/assetlinks.go`, `tests/assetlinks_test.go`, `.env.example` (`ANDROID_PACKAGE`, `ANDROID_CERT_SHA256`), `apps/mobile/app.config.ts` (`android.intentFilters`), `apps/mobile/src/linking.test.ts`, `.maestro/smoke/deep-link.yaml`
  - Refs: SDD 10 §7.6, FR-10-7, §9; <https://vukem.dev/docs/basics/routing> "Defining routes"; [E12]
  - Depends on: P5.B.2.5, P5.B.4.1
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests -run TestAssetLinks` → 200 `application/json`, `Cache-Control: public, max-age=3600`, no auth, no redirect, package `app.hexascore`, every fingerprint in `ANDROID_CERT_SHA256`
    - `adb shell pm get-app-links app.hexascore` (preview build, staging host override) → `verified`; `maestro test .maestro/smoke/deep-link.yaml` → PASSED for `/pt-br/games/<slug>` and `/entities/<id>`
  - Notes: `ANDROID_CERT_SHA256` holds a comma-separated list (upload key now, Play signing key added in P5.C.2.1, from `eas credentials -p android`). Locale-prefixed paths open the same screen with that locale applied for the session only.

- [ ] **P5.B.5.2** Complete the Maestro smoke suite (`login`, `rate-item`, `add-to-list`, `offline-rating`, `deep-link`) on staging and run the accessibility checks (role/label queries, TalkBack, 200 % font scale, Accessibility Scanner on 8 screens).
  - Files/modules: `.maestro/smoke/offline-rating.yaml`, `.maestro/smoke/config.yaml`, `docs/release/mobile-a11y-checklist.md`
  - Refs: SDD 10 §9, §10, FR-10-4; [E13]
  - Depends on: P5.B.5.1, P5.B.4.3, P5.B.3.1
  - Model: Opus
  - Acceptance:
    - `maestro test .maestro/smoke/` (staging) → all five flows PASSED
    - Accessibility Scanner on home, listing, item, rate sheet, search, lists, profile, login → 0 findings recorded in the checklist
  - Notes: the airplane-mode toggle in `offline-rating.yaml` depends on what the emulator allows. If Maestro cannot toggle the network, record it here and use an `adb` network toggle step in CI. Do not mark the flow passed without it.

#### P5.B.6 Release

- [ ] **P5.B.6.1** Release `v1.7.0` (assetlinks route in production) and build `mobile-v0.9.0` as a preview APK for owner testing.
  - Files/modules: `CHANGELOG.md`, `apps/mobile/app.config.ts` (`version`), `docs/PROGRESS.md`
  - Refs: ADR 0013; SDD 10 §7.4, §9
  - Depends on: P5.B.5.2
  - Model: Sonnet
  - Acceptance:
    - `curl -si https://hexascore.app/.well-known/assetlinks.json` → 200 `application/json`, package `app.hexascore`
    - `bun run --cwd apps/mobile typecheck && bun run --cwd apps/mobile lint && bun run --cwd apps/mobile test` → green
  - Notes: the owner pushes the tag.

### P5.C Mobile release

#### P5.C.1 Build pipeline

- [ ] **P5.C.1.1** Configure EAS: profiles `development`, `preview` (APK), `e2e-test`, `production` (AAB); remote `versionCode` with `autoIncrement` on production; `runtimeVersion` policy `fingerprint`; one update channel per profile; an EAS Workflow on PRs that builds `e2e-test` and runs Maestro.
  - Files/modules: `apps/mobile/eas.json`, `apps/mobile/app.config.ts`, `apps/mobile/.eas/workflows/pr.yml`, `docs/release/mobile-ota-rule.md`
  - Refs: SDD 10 §7.4, §7.8 (pipeline, OTA), §10; D-34; [E8], [E13], [E17], [E18]
  - Depends on: P5.B.6.1
  - Model: Opus
  - Acceptance:
    - `bunx eas-cli build --profile preview --platform android --non-interactive` → build finished, APK artifact URL
    - PR workflow run → `e2e-test` build plus Maestro smoke → passed
  - Notes: EAS free tier (D-34). Record the monthly build quota used in `docs/PROGRESS.md`. The OTA rule document lists what may ship over the air (JS/asset, copy, catalogue fixes) and what forces a store release (fingerprint change, permission, target API, new screen or feature, data collection, min-version raise).

- [ ] **P5.C.1.2** Prepare the Gradle-on-GitHub-Actions fallback: `expo prebuild` output, `bundleRelease` signed with the upload key from Actions secrets, AAB uploaded as an artifact. Run it manually.
  - Files/modules: `.github/workflows/mobile-gradle.yml`, `docs/runbooks/mobile-build-fallback.md`
  - Refs: SDD 10 §7.8 "Build service"; D-34; D-26; [E3]
  - Depends on: P5.C.1.1
  - Model: Opus
  - Acceptance:
    - `gh workflow run mobile-gradle.yml -f ref=main` then `gh run watch` → success, artifact `app-release.aab` present
    - `unzip -l app-release.aab | rg BundleConfig.pb` → one match
  - Notes: never commit the generated `android/` folder or the keystore. The upload key lives only in Actions secrets and with EAS credentials. Pin the JDK and Gradle action versions. The workflow stays `workflow_dispatch` until EAS stops being free.

- [ ] **P5.C.1.3** Add the tag workflow for `mobile-v*`: `eas build --profile production` and `eas submit` to the internal testing track, gated on lint, typecheck, Jest and the contract check.
  - Files/modules: `apps/mobile/.eas/workflows/release.yml` (or `.github/workflows/mobile-release.yml`), `docs/release/mobile-release-checklist.md`
  - Refs: SDD 10 §7.8 "Pipeline", §10 "Release checklist"; ADR 0013
  - Depends on: P5.C.1.1
  - Model: Sonnet
  - Acceptance:
    - a dry run on tag `mobile-v0.9.1` → AAB submitted to internal testing, visible in Play Console
  - Notes: the release checklist covers target API, data safety, assetlinks, deletion URL, min version, OTA rule and staged rollout (SDD 10 §10). The Play service account JSON is an EAS secret, never in the repository.

#### P5.C.2 Play Store compliance

- [ ] **P5.C.2.1** Close the Play policy items: the `/help/delete-account` page works without the app in three locales; the Play signing fingerprint is added to `ANDROID_CERT_SHA256`; the data safety answers are recorded; target API 36 is confirmed.
  - Files/modules: `docs/release/play-data-safety.md`, `docs/release/mobile-release-checklist.md`, production env `ANDROID_CERT_SHA256`
  - Refs: SDD 10 §7.8 (account deletion, data safety, target API), FR-10-8; SDD 09 §6.2; SDD 03 FR-03-15, §7.5; open issue 10-7; [E14], [E15], [E16]
  - Depends on: P5.B.6.1, P1.G
  - Model: Sonnet
  - Acceptance:
    - `curl -s https://hexascore.app/help/delete-account | rg -c "30"` → ≥ 1 (grace period stated); same for `/pt-br/` and `/es/` paths
    - `curl -s https://hexascore.app/.well-known/assetlinks.json | jq '.[0].target.sha256_cert_fingerprints | length'` → 2
  - Notes: **owner input needed:** the Play Console account and the data safety form submission are the owner's. The answers follow SDD 10 §7.8 exactly: email, name, interactions, user content, search history (30 days) and optional crash logs are collected, not shared, encrypted in transit and deletable. Location, contacts, device IDs and financial data are not collected. Verify that no bundled SDK contradicts this before submitting.

- [ ] **P5.C.2.2** Produce the store listing: texts in `en`, `pt-BR`, `es`, the 512 px icon, the 1024 × 500 graphic, ≥ 4 screenshots per locale from Maestro `takeScreenshot`, and the privacy URL `/legal/privacy`.
  - Files/modules: `apps/mobile/store/listing.<locale>.md`, `.maestro/screenshots/*.yaml`, `apps/mobile/store/assets/`
  - Refs: SDD 10 §7.8 "Listing assets"
  - Depends on: P5.C.2.1
  - Model: Sonnet
  - Acceptance:
    - `maestro test .maestro/screenshots/` → ≥ 4 PNGs per locale in `apps/mobile/store/assets/screenshots/<locale>/`
  - Notes: copy edits go to the `text-editor` subagent. Listing texts state that billing and API keys stay on the web (D-12).

#### P5.C.3 Tracks

- [ ] **P5.C.3.1** Ship `mobile-v1.0.0` to the internal testing track and measure the NFR budgets on the reference device: cold start, scroll jank, item page timing, memory, bundle and download size.
  - Files/modules: `docs/release/mobile-perf-1.0.0.md`, `apps/mobile/app.config.ts` (`version` `1.0.0`)
  - Refs: SDD 10 §4 (NFR-10-1…7), §9; ROADMAP P5 exit criteria
  - Depends on: P5.C.1.3, P5.C.2.2
  - Model: Sonnet
  - Acceptance:
    - `adb shell am start -W app.hexascore/.MainActivity` ×10 → `TotalTime` p90 ≤ 2500 ms cached, ≤ 3500 ms uncached
    - `adb shell dumpsys gfxinfo app.hexascore` after a 200-row fling → janky frames ≤ 5 %; `grep -rE "hx_app_|sk_live|APP_KEY" apps/mobile/dist` → no match
  - Notes: a budget miss blocks promotion. Record the numbers, open a task, and only then consider FlashList (ADR) or the 2× thumb (open issue 10-3).

- [ ] **P5.C.3.2** Promote to closed testing for ≥ 7 days and track crash-free sessions and `client.errors` by `app_version`.
  - Files/modules: `docs/release/mobile-release-checklist.md`, `docs/PROGRESS.md`
  - Refs: SDD 10 §7.8 "Pipeline" (closed ≥ 7 days, crashes ≤ 1 % of sessions); SDD 12 §6.2 (`client.errors`)
  - Depends on: P5.C.3.1
  - Model: Sonnet
  - Acceptance:
    - Play Console vitals after 7 days → crash-free sessions ≥ 99 %, recorded with the date in `docs/PROGRESS.md`
  - Notes: the owner manages the testers list. A crash fix that changes only JS may ship over the air under the OTA rule; anything else needs a new store build.

- [ ] **P5.C.3.3** Release to production with a staged rollout of 20 % → 100 % over ≥ 3 days, raise `APP_MIN_SUPPORTED_VERSION` to `1.0.0`, and release `v1.8.0`.
  - Files/modules: `CHANGELOG.md`, production env `APP_MIN_SUPPORTED_VERSION`, `docs/PROGRESS.md`
  - Refs: SDD 10 §7.4, §7.8; SDD 05 §6.4 (kill switch); ADR 0013
  - Depends on: P5.C.3.2
  - Model: Sonnet
  - Acceptance:
    - `curl -s https://hexascore.app/up` → `v1.8.0`; `curl -si -H "X-App-Version: 0.9.0" -H "Authorization: Bearer $T" https://hexascore.app/app/v1/me` → 426
    - Play Console → production track at 100 %, listing approved (phase exit)
  - Notes: raise the min version only after 100 % rollout, or users stuck at 20 % would be blocked. The owner pushes both tags (`v1.8.0`, `mobile-v1.0.0`).

## Phase 6: ingestion and recommendations

Ingestion conventions for every P6.A/P6.B microtask: Go code lives under `internal/ingest/` with integration tests in `tests/ingest/`. Each port names the Python function in `legacy/webscrappings/` and the pytest it mirrors; fixtures are copied from `legacy/webscrappings/tests/fixtures/`. The Python package is the executable specification (D-11), and Python stays reference-only. Outbound HTTP goes only through `httpclient.Safe` (NFR-06-6). Every source is behind a fail-closed kill switch. Jobs run inside `serve` (ADR 0007, D-08).

P6.C depends only on P2.G and can start before P6.A. P3.A.1.4 (API trending) stays open until P6.C.3.2 closes it.

### P6.A Ingestion in Go

#### P6.A.1 Legal gate and foundations

- [ ] **P6.A.1.1** Read the Open Library licensing and covers pages and record quotes, date and the A-05 decision in `docs/legal/open-library.md`, then set `CATALOG_API_SOURCES` and the `openlibrary` licence row accordingly.
  - Files/modules: `docs/legal/open-library.md`, `database/seeders/data_sources.go` (licence row), `.env.example` (`CATALOG_API_SOURCES`), `docs/sdd/06-ingestion.md` §7.1, `docs/sdd/open-issues.md` (01-2, 06-1)
  - Refs: SDD 06 §7.1 "A-05 blocking task", §11 issue 1; SDD 01 open issue 2; A-05; D-03
  - Depends on: P2.G
  - Model: Opus
  - Acceptance:
    - `rg -c "openlibrary.org/developers/licensing|openlibrary.org/dev/docs/api/covers" docs/legal/open-library.md` → 2, with an access date and quoted text
    - `rg "01 \| 2|06 \| 1" docs/sdd/open-issues.md` → both rows `decided`
  - Notes: apply the rule as written. CC0 or public domain → ingest, API-eligible. Attribution-only → API-eligible with `data_sources.attribution`. Silent, non-commercial or no-redistribution → do not ingest, purge, and drop it from `CATALOG_API_SOURCES`. Covers are API-eligible only with an explicit grant. Architect task; the owner confirms the file before P6.A.2.4 enables the connector.

- [ ] **P6.A.1.2** Create the ingestion tables (`source_records` extensions, `source_record_history`, `job_runs`, `ingest_failures`, `entity_match_reviews`, `dq_results`, `catalog_counts`, `field_coverage`) and the record store: canonical SHA-256, upsert that writes history in the same CTE, and `checked_at`-only on an unchanged hash.
  - Files/modules: `database/migrations/<ts>_create_ingestion_tables.go`, `internal/ingest/store/hash.go`, `internal/ingest/store/records.go`, `internal/ingest/store/records_test.go`
  - Refs: SDD 06 §5, FR-06-1; <https://vukem.dev/docs/database/migrations> "Migrations" (`s.Raw`); VF-006; legacy `common/db.py` `canonical_hash`, `upsert_source_record`, `store_records`
  - Depends on: P2.F, P6.A.1.1
  - Model: Opus
  - Acceptance:
    - `vukem migrate && vukem migrate:status` → `create_ingestion_tables` Ran
    - `go test ./internal/ingest/store/... -run 'TestCanonicalHashIgnoresKeyOrder|TestUpsertIsNoopWhenPayloadUnchanged|TestStoreRecordsCountsInsertsAndChanges|TestHistoryWrittenOnHashChange'` → `ok` (ports `test_canonical_hash_ignores_key_order`, `test_upsert_is_noop_when_payload_unchanged`, `test_store_records_counts_inserts_and_changes`)
  - Notes: P2.F already imported `source_records` 1:1. Add only the missing columns (`checked_at`, `last_run_id`) and the `source_records_item_idx` index, and backfill `checked_at = fetched_at`. jsonb and text + CHECK are raw DDL (VF-006). The hash must equal Python's for the same payload; test against three fixture payloads hashed by `canonical_hash`.

- [ ] **P6.A.1.3** Build the run framework: the `job_runs` lifecycle, heartbeat (stale > 15 min → `failed`, resumable), chunked continuation (≤ 500 targets or 5 min, cursor saved, successor via `q.DispatchTx` in the same transaction), `queue.ErrPermanent`, closing the run on the last attempt, per-target `ingest_failures` (parked after 5 runs), and fail-closed kill switches.
  - Files/modules: `internal/ingest/run/run.go`, `internal/ingest/run/continuation.go`, `internal/ingest/run/flags.go`, `internal/ingest/run/failures.go`, `tests/ingest/resume_test.go`, `tests/ingest/killswitch_test.go`
  - Refs: SDD 06 §6 (jobs, kill switches), §7.6 "Dead letter", FR-06-6, FR-06-7; <https://vukem.dev/docs/digging-deeper/queue> "Jobs and transactions", "Retries, backoff and timeouts"; <https://vukem.dev/docs/digging-deeper/feature-flags> "Feature flags"; VF-018
  - Depends on: P6.A.1.2
  - Model: Opus
  - Acceptance:
    - `TEST_DATABASE_URL=… go test ./tests/ingest -run TestResumeAfterCancel -count=20` → tables equal an uninterrupted run (SDD 06 §9)
    - `go test ./tests/ingest -run 'TestKillSwitchSkips|TestOneActiveRunPerJob'` → disabled source ends `skipped` with zero `httptest` hits; a second queued run → 409 `run-active` (unique index)
  - Notes: queue `ingest`, `queue.Timeout(10*time.Minute)`, `queue.Retries(3)` (that is four attempts, per the queue page). VF-018 workaround: an app-level semaphore caps concurrent ingestion jobs at 2 (NFR-06-2) whatever `QUEUE_CONCURRENCY` is. The "close the run on the last attempt" helper SDD 06 §6 names (`queue.AttemptFrom(ctx).Last()`) is not on the fetched queue page; confirm it on <https://vukem.dev/docs/digging-deeper/queue> before use. If it is missing, write `[NOT IN VUKEM DOCS]`, add a `vukem-feedback.md` entry, and close the run from a `q.AddObserver` failure hook instead.

- [ ] **P6.A.1.4** Build the outbound client on `httpclient.Safe`: per-host throttle with slot reservation, retries for transport errors and 429/5xx (5 attempts, `2^n` s ≤ 60 s), `Retry-After` ≤ 300 s that pauses the whole host, rate halving after 20 hits, a per-host breaker, the bot User-Agent, a host allowlist and size caps.
  - Files/modules: `internal/ingest/httpx/client.go`, `internal/ingest/httpx/throttle.go`, `internal/ingest/httpx/breaker.go`, `internal/ingest/httpx/client_test.go`, `tests/ingest/security_ssrf_test.go`
  - Refs: SDD 06 §7.2, NFR-06-5, NFR-06-6; <https://vukem.dev/docs/security/overview> "A01"; VF-020; legacy `common/http.py` (`_throttle`, `retry_after_seconds`, `user_agent`), `tests/test_http.py`
  - Depends on: P6.A.1.3
  - Model: Opus
  - Acceptance:
    - `go test ./internal/ingest/httpx/... -run 'TestUserAgentIdentifiesBot|TestRetryAfterParsing|TestRetries429HonoringRetryAfterThenSucceeds|TestGivesUpAfterMaxAttempts|TestDoesNotRetryClientErrors|TestRateLimitSpacesRequestsPerHost|TestRateLimitedResponsePausesTheWholeHost|TestBreakerOpensAndProbes'` → `ok` on a fake clock
    - `go test ./tests/ingest -run TestSecurity_A01_IngestSSRF` → a private IP or a non-allowlisted host is refused before connecting
  - Notes: this is the VF-020 workaround (app code, no new dependency). Rates come from the §7.2 table as configuration. The User-Agent is `HexaScoreBot/1.0 (+$INGEST_CONTACT_URL)`. Limits: JSON ≤ 20 MB, images ≤ 15 MB. Credentials are `config.Secret`.

#### P6.A.2 Open connectors

- [ ] **P6.A.2.1** Define the `Connector` interface and the Wikidata selection: SPARQL by `P31` classes and sitelinks (bootstrap top-N per type, incremental `dateModified ≥ since` with ≥ 15 sitelinks).
  - Files/modules: `internal/ingest/connector.go`, `internal/ingest/wikidata/select.go`, `internal/ingest/wikidata/select_test.go`, `internal/ingest/testdata/sparql_top_movies.json`
  - Refs: SDD 06 §7.3 (interface, selection), §7.2 (query.wikidata.org 1/s); legacy `connectors/wikidata.py` `select_qids`, `top_query`, `modified_query`, `_values`; tests `test_select_qids_parses_sparql_bindings`, `test_queries_target_type_classes`
  - Depends on: P6.A.1.4
  - Model: Opus
  - Acceptance:
    - `go test ./internal/ingest/wikidata/... -run 'TestSelectQidsParsesSparqlBindings|TestQueriesTargetTypeClasses'` → `ok`
  - Notes: the interface is exactly SDD 06 §7.3 (`Source`, `Targets` as `iter.Seq2`, `Fetch`). Do not add methods "for later".

- [ ] **P6.A.2.2** Fetch and trim Wikidata entities (`wbgetentities`, 50 ids, kept props incl. aliases and P144) and fetch referenced QIDs as `wikidata:entity` records.
  - Files/modules: `internal/ingest/wikidata/entities.go`, `internal/ingest/wikidata/trim.go`, `internal/ingest/wikidata/entities_test.go`
  - Refs: SDD 06 §7.3 ("trimming is load-bearing"), §7.4 (role properties), FR-06-11; legacy `connectors/wikidata.py` `simplify_value`, `trim_entity`, `referenced_qids`, `get_entities`, `fetch_entities`; tests `test_trim_entity_keeps_only_used_data`, `test_score_statements_keep_their_qualifiers`, `test_missing_entities_are_skipped`, `test_invalid_entity_payload_is_rejected`
  - Depends on: P6.A.2.1
  - Model: Opus
  - Acceptance:
    - `go test ./internal/ingest/wikidata/... -run 'TestTrimEntityKeepsOnlyUsedData|TestMissingEntitiesAreSkipped|TestInvalidEntityPayloadIsRejected'` → `ok`
    - `go test ./internal/ingest/wikidata/... -run TestTrimHashMatchesPython` → hash of the trimmed `wikidata_entities.json` equals the Python value recorded in testdata
  - Notes: score qualifiers are dropped (D-09, no critic scores). Port `test_score_statements_keep_their_qualifiers` as a test that score statements are removed, and note the deliberate divergence in the test name. Adding a property changes every hash and needs a `full` run (≈ 1 h).

- [ ] **P6.A.2.3** Port the Wikipedia connector for `en`, `pt` and `es`: batched summaries/extracts that reject disambiguation, missing pages and a mismatched `wikibase_item`, with concurrent requests within host limits.
  - Files/modules: `internal/ingest/wikipedia/wikipedia.go`, `internal/ingest/wikipedia/wikipedia_test.go`
  - Refs: SDD 06 §7.1 (CC BY-SA 4.0 attribution), §7.2, §7.3; legacy `connectors/wikipedia.py` `parse_extracts`, `targets`, `fetch`; tests `test_wikipedia_extracts`, `test_wikipedia_rejects_disambiguation_missing_and_other_entities`, `test_wikipedia_fetch_runs_requests_concurrently`
  - Depends on: P6.A.2.2
  - Model: Sonnet
  - Acceptance:
    - `go test ./internal/ingest/wikipedia/... -run 'TestWikipediaExtracts|TestWikipediaRejectsDisambiguationMissingAndOtherEntities|TestWikipediaFetchRunsRequestsConcurrently'` → `ok`
  - Notes: store the source URL and licence with each extract so DQ-06 can pass.

- [ ] **P6.A.2.4** Port the Open Library connector behind `ingest_openlibrary` and the A-05 decision: edition preferred over work, the work fills what the edition lacks, typed descriptions, and placeholder covers (`-1`) dropped.
  - Files/modules: `internal/ingest/openlibrary/openlibrary.go`, `internal/ingest/openlibrary/openlibrary_test.go`
  - Refs: SDD 06 §7.1, §7.3; A-05; `docs/legal/open-library.md`; legacy `connectors/openlibrary.py` `document_path`, `work_path`, `parse_doc`, `complete_ids`; tests `test_openlibrary_paths_prefer_editions`, `test_openlibrary_edition`, `test_openlibrary_work_fills_what_the_edition_lacks`, `test_openlibrary_reads_typed_descriptions`, `test_openlibrary_drops_placeholder_covers`
  - Depends on: P6.A.2.2, P6.A.1.1
  - Model: Sonnet
  - Acceptance:
    - `go test ./internal/ingest/openlibrary/... -run 'TestOpenlibrary'` → the five ported tests `ok`
    - `go test ./tests/ingest -run TestOpenLibraryOffWithoutLegalFile` → flag off → `skipped`, zero hits
  - Notes: if `docs/legal/open-library.md` says "do not ingest", stop after the tests. Keep the connector disabled, purge imported `openlibrary` records in a separate reviewed microtask, and tell the owner.

#### P6.A.3 Normaliser

- [ ] **P6.A.3.1** Port `BuildItem` core fields as a pure function: skip without a Wikidata record or English title, slug (ASCII, ≤ 80, `-<year>` then `-<qid>` on collision, never changed), release date (min P577; series P580 first; fallback P571) with `release_precision` and `release_status`, titles per locale, external ids and quantities.
  - Files/modules: `internal/ingest/normalize/build.go`, `internal/ingest/normalize/slug.go`, `internal/ingest/normalize/dates.go`, `internal/ingest/normalize/build_test.go`
  - Refs: SDD 06 §7.3, FR-06-2; SDD 01 FR-01-15; open issue 02-1; legacy `normalize.py` `slugify`, `slug_candidates`, `parse_time`, `release_date`, `first_value`, `quantity`, `external_ids`, `build_translations`, `build_item`; tests `test_slugify_is_ascii_lowercase`, `test_slug_candidates_add_year_then_qid`, `test_parse_time_precisions`, `test_movie_mapping`, `test_series_mapping`, `test_items_without_wikidata_or_english_title_are_skipped`
  - Depends on: P6.A.2.2
  - Model: Opus
  - Acceptance:
    - `go test ./internal/ingest/normalize/... -run 'TestSlugify|TestSlugCandidates|TestParseTimePrecisions|TestMovieMapping|TestSeriesMapping|TestItemsWithoutWikidataOrEnglishTitleAreSkipped'` → `ok`
    - `go test ./internal/ingest/normalize/... -run TestBuildItemOrderIndependent -count=50` → a property test shuffling records gives identical output (FR-06-2)
  - Notes: `release_status = announced` when the date is in the future. The TMDB/IGDB "unreleased" field names are [NOT VERIFIED]; wire them in P6.B.1.1/P6.B.1.2 with fixtures. `scores` and `_score_time` are not ported (D-09).

- [ ] **P6.A.3.2** Port the enrichment merge with provenance: Wikidata wins, other sources fill empty fields only, `{field: source}` provenance and licence from `data_sources`, synopsis only from Wikipedia or `synopsis`, TMDB/IGDB/Open Library text never in `item_translations`, and `community` fields skipped for 180 days, with a `system` suggestion on divergence.
  - Files/modules: `internal/ingest/normalize/enrich.go`, `internal/ingest/normalize/provenance.go`, `internal/ingest/normalize/enrich_test.go`, `tests/ingest/community_lock_test.go`
  - Refs: SDD 06 FR-06-3, FR-06-4, §7.3; SDD 04 §7.3; open issue 04-3; ADR 0006; legacy `normalize.py` `enrichment`, `build_meta`; tests `test_tmdb_fills_only_what_wikidata_lacks`, `test_igdb_cover_wins_over_steam`, `test_game_mapping_with_steam_enrichment`, `test_book_mapping_uses_open_library_pages_as_fallback`, `test_own_synopsis_only_fills_locales_without_wikipedia`, `test_commons_cover_attribution`
  - Depends on: P6.A.3.1, P4.B
  - Model: Opus
  - Acceptance:
    - `go test ./internal/ingest/normalize/... -run 'TestTmdbFillsOnlyWhatWikidataLacks|TestIgdbCoverWinsOverSteam|TestGameMappingWithSteamEnrichment|TestBookMappingUsesOpenLibraryPagesAsFallback|TestOwnSynopsisOnlyFillsLocalesWithoutWikipedia|TestCommonsCoverAttribution'` → `ok`
    - `go test ./tests/ingest -run TestCommunityFieldLock` → at day 179 the source is skipped and a `system` suggestion opens; at day 181 a divergence still opens a suggestion and never overwrites (fake clock)
  - Notes: enrichment payloads come from fixtures here; the connectors land in P6.B.1. The `system` suggestion uses the SDD 04 model from P4.A/P4.B. Do not write to the suggestions tables directly; call its service.

- [ ] **P6.A.3.3** Extract relations: entity credits by role property (§7.4 table, cast first 15), `item_alt_titles` from aliases, transliterations, original-language and former titles (≤ 20, source `wikidata`), and `item_links` `based_on` from P144 when both QIDs are catalogue items.
  - Files/modules: `internal/ingest/normalize/relations.go`, `internal/ingest/normalize/alttitles.go`, `internal/ingest/normalize/relations_test.go`, `internal/ingest/testdata/parity_based_on.json`
  - Refs: SDD 06 §7.4, FR-06-11; SDD 01 §5.2 (`item_alt_titles`, `item_links`, `item_entities`, role compatibility), DQ-08, DQ-14; open issues 07-6, 08-1, 06-9; D-30 (no engines or edition contributors)
  - Depends on: P6.A.3.1
  - Model: Opus
  - Acceptance:
    - `go test ./internal/ingest/normalize/... -run 'TestRelationsByRole|TestAltTitles|TestBasedOnLink'` → a manga-based anime and a novel-based film fixture produce `based_on`; alt titles capped at 20 with the right `kind` and `-Latn` lang
  - Notes: role properties beyond the legacy set are [NOT VERIFIED] until fixtures pin them (SDD 06 issue 9). Record each fixture's QID and property in testdata. Entity matching against existing rows is P6.B.3.2; here, emit credits by QID only.

- [ ] **P6.A.3.4** Build the `normalize` job: pending QIDs from `source_records_pending_idx`, one transaction per 200 QIDs, slug collision against the database, `catalog.ItemsChanged{IDs}` per committed batch, and `RecomputeItemScore` via `DispatchTx` when a rated item's date, precision or status changes.
  - Files/modules: `internal/ingest/normalize/job.go`, `internal/ingest/normalize/save.go`, `tests/ingest/normalize_job_test.go`
  - Refs: SDD 06 §6 "Events", §7.3, NFR-06-3; SDD 02 §6; <https://vukem.dev/docs/digging-deeper/events> "Queued listeners"; <https://vukem.dev/docs/digging-deeper/queue> "Jobs and transactions"; legacy `normalize.py` `choose_slug`, `known_items`, `save_item`, `pending_qids`, `normalize_batch`; tests `test_normalize_writes_item_translations_and_scores`, `test_slug_collision_appends_year`
  - Depends on: P6.A.3.2, P6.A.3.3
  - Model: Opus
  - Acceptance:
    - `go test ./tests/ingest -run 'TestNormalizeWritesItemTranslations|TestSlugCollisionAppendsYear|TestRecomputeDispatchedOnDateChange|TestItemsChangedPerBatch'` → `ok`
  - Notes: stream rows in batches of 200 and read only the needed JSON keys (NFR-06-3, ≤ 256 MB). The ported translations test drops its score assertions (D-09).

- [ ] **P6.A.3.5** Build the parity test: `testdata/parity/qids.txt` (200 QIDs, 50 per type, top/mid/tail, non-Latin titles, year-precision dates, P580 series, ISBN-10-only book, Steam-only game) with frozen `source_records`, a Python runner that emits canonical JSON from `build_item`, and Go `TestParityLegacy`.
  - Files/modules: `testdata/parity/qids.txt`, `testdata/parity/records/*.json`, `testdata/parity/python/*.json` (committed outputs), `scripts/parity/dump_python.py` (run against `legacy/webscrappings` with `uv run`), `tests/ingest/parity_test.go`
  - Refs: SDD 06 §7.9 "Parity", FR-06-10, §9; open issue 06-9; D-11
  - Depends on: P6.A.3.4
  - Model: Opus
  - Acceptance:
    - `go test ./tests/ingest -tags parity -run TestParityLegacy` → `0 semantic diffs / 200`
  - Notes: compare exactly the §7.9 field list. Synopsis is compared after trim, whitespace collapse and NFC. Allowed diffs: scores, precision/status, entity ids, `creators` split by role. Commit the Python outputs so CI needs no Python. A diff that reveals a legacy bug is fixed in Go only after it is written up in the test as an allowed diff with the reason. Closes open issue 06-9.

#### P6.A.4 Schedules and release

- [ ] **P6.A.4.1** Wire the schedules and chains for the open sources (`bootstrap` command, `fetch.wikidata` incremental at 00:30, weekly Sunday Wikidata + entities, Monday Wikipedia, Tuesday Open Library, each followed by normalise) with `SCHEDULE_TIMEZONE=America/Sao_Paulo` and `WithoutOverlapping()`.
  - Files/modules: `routes/schedule.go`, `internal/ingest/chains.go`, `cmd/app/ingest_bootstrap.go`, `tests/ingest/schedule_test.go`
  - Refs: SDD 06 §7.6; <https://vukem.dev/docs/digging-deeper/scheduler> "Frequencies", "Overlapping runs"; <https://vukem.dev/docs/digging-deeper/queue> "Inside serve"; ADR 0007; D-08; VF-018
  - Depends on: P6.A.3.5
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests/ingest -run 'TestScheduleTable|TestChainDispatchesNext'` → expressions match §7.6 (`DailyAt("00:30")`, `WeeklyOn(time.Sunday, "01:00")` …); a successful stage dispatches the next; a failed stage stops the chain
    - `QUEUE_CONNECTION=database vukem queue:work -queue ingest -stop-when-empty` after `bootstrap --limit 20` on a dev database → 80 items normalised, `job_runs` all `succeeded`
  - Notes: scheduler locks live in `app.Cache()`. Use `CACHE_STORE=database` in production so a restart does not double-run (scheduler page "Overlapping runs"). Registering custom commands is [NOT VERIFIED] in SDD 06 §7.9; confirm it on the vukem CLI docs or expose bootstrap as an admin run only (P6.B.4.3).

- [ ] **P6.A.4.2** Release `v1.9.0` with `ingest_wikidata` and `ingest_wikipedia` on in staging (and `ingest_openlibrary` only if A-05 allows), after 7 days of green daily runs.
  - Files/modules: `CHANGELOG.md`, staging/production env `FLAG_INGEST_*`, `docs/PROGRESS.md`
  - Refs: ADR 0013; SDD 06 §6 "Kill switches", §9
  - Depends on: P6.A.4.1
  - Model: Sonnet
  - Acceptance:
    - `curl -s https://staging.hexascore.app/up` → `v1.9.0`
    - `psql "$STAGING_DATABASE_URL" -c "select job, status, count(*) from job_runs where created_at > now()-interval '7 days' group by 1,2"` → no `failed` rows for `fetch.wikidata` or `normalize`
  - Notes: production flags are turned on by the owner. Asserts arrive in P6.B.4.1, so watch item counts by hand in this window.

### P6.B Ingestion extras

#### P6.B.1 Display-only enrichment

- [ ] **P6.B.1.1** Port the TMDB connector (IDs only via `/find/{imdb_id}`, movie and TV details, a 401 stops the run with `ErrPermanent`) and the `tmdb.expire` job (refresh when `checked_at` > 150 d; delete, re-normalise and drop the TMDB cover when > 180 d), with audience `site` only.
  - Files/modules: `internal/ingest/tmdb/tmdb.go`, `internal/ingest/tmdb/expire.go`, `internal/ingest/tmdb/tmdb_test.go`, `tests/ingest/tmdb_expiry_test.go`
  - Refs: SDD 06 §7.1 (TMDB row), §7.6 (`tmdb.expire`), FR-06-5, DQ-I-06; ADR 0006; D-03; legacy `connectors/tmdb.py` `parse_find`, `parse_details`, `targets`, `fetch`; tests `test_tmdb_movie_details`, `test_tmdb_series_details`, `test_tmdb_find_picks_the_result_of_the_item_type`, `test_tmdb_fetch_resolves_missing_ids_through_imdb`, `test_tmdb_rejected_token_stops_the_run`
  - Depends on: P6.A.4.2
  - Model: Sonnet
  - Acceptance:
    - `go test ./internal/ingest/tmdb/... -run 'TestTmdb'` → the five ported tests `ok`
    - `go test ./tests/ingest -run TestTmdbExpiry` → at 151 d the record is refreshed; at 181 d the record and cover are gone and the item is re-normalised (fake clock); `psql -c "select count(*) from source_records where source='tmdb' and checked_at < now()-interval '180 days'"` → `0`
  - Notes: TMDB is never commercial and never in `/api/v1`. The logo and the not-endorsed notice on item pages that use TMDB fields and on `/about` are checked in P6.B.4.1 (DQ-I-12) and in the Playwright spec of P6.B.4.3.

- [ ] **P6.B.1.2** Port the IGDB connector (Twitch app token, regex-filtered slugs, batched games query, pre-1970 dates) and the D-36 interim rule: `igdb.refresh` every 6 h for records with `checked_at` > 24 h, and records or covers older than 7 days are deleted and re-normalised.
  - Files/modules: `internal/ingest/igdb/igdb.go`, `internal/ingest/igdb/refresh.go`, `internal/ingest/igdb/igdb_test.go`, `tests/ingest/igdb_retention_test.go`
  - Refs: SDD 06 §7.1 (IGDB row, "IGDB interim rule"), §7.6 (`igdb.refresh`), DQ-I-13; D-36; open issue 06-2; legacy `connectors/igdb.py` `games_query`, `parse_games`, `access_token`; tests `test_igdb_games`, `test_igdb_release_dates_before_1970`, `test_igdb_fetch_authenticates_and_drops_unsafe_slugs`
  - Depends on: P6.A.4.2
  - Model: Opus
  - Acceptance:
    - `go test ./internal/ingest/igdb/... -run 'TestIgdb'` → the three ported tests `ok`
    - `go test ./tests/ingest -run TestIgdbRetention` → nothing IGDB-sourced older than 7 days survives a `igdb.refresh` run (fake clock)
  - Notes: **owner input needed:** the owner records `docs/legal/igdb.md` before the ingestion milestone (D-36). If the file allows longer retention, change D-36, then SDD 06, then this job, in that order (AGENTS.md). Budget: 3.3/s, cap 4/s, 4 in flight.

- [ ] **P6.B.1.3** Port the Steam `appdetails` and GOG catalogue connectors (header image, OS platforms, store link; GOG matching by slug, then title, then slug prefix), with images kept site-only.
  - Files/modules: `internal/ingest/steam/steam.go`, `internal/ingest/gog/gog.go`, `internal/ingest/steam/steam_test.go`, `internal/ingest/gog/gog_test.go`
  - Refs: SDD 06 §7.1 (Steam, GOG rows), §7.2; open issue 06-8; legacy `connectors/steam.py` `parse_appdetails`, `connectors/gog.py` `_norm`, `pick_product`, `parse_catalog`; tests `test_steam_appdetails`, `test_gog_picks_product_by_slug_prefix`, `test_gog_no_match`, `test_enrichment_skips_failed_targets`
  - Depends on: P6.A.4.2
  - Model: Sonnet
  - Acceptance:
    - `go test ./internal/ingest/steam/... ./internal/ingest/gog/... -run 'TestSteamAppdetails|TestGogPicksProductBySlugPrefix|TestGogNoMatch|TestEnrichmentSkipsFailedTargets'` → `ok`
  - Notes: Steam and GOG terms are [NOT VERIFIED] (open issue 06-8). Images stay out of `/api/v1` and the app until the architect records the terms in `docs/legal/steam-gog.md`; that reading is a prerequisite for any change to the app audience, not for this microtask. Store ids come from Wikidata and are used for links only.

#### P6.B.2 Covers and trailers

- [ ] **P6.B.2.1** Encode WebP with a pinned `cwebp` (Alpine `libwebp-tools`) run in a resource-limited child process, behind an `Encoder` used by the covers job, with a ±10 % size parity test against the Pillow output.
  - Files/modules: `internal/ingest/covers/webp.go`, `internal/ingest/covers/webp_test.go`, `internal/ingest/testdata/covers/*.jpg|*.png` with Pillow reference sizes, `Dockerfile` (pinned `libwebp-tools`)
  - Refs: SDD 06 §7.5, §11 issue 3; SDD 11 §6.3; open issue 06-3; VF-004; legacy `covers.py` `to_webp`; tests `test_to_webp_resizes_and_strips_metadata`, `test_to_webp_keeps_alpha_and_never_upscales`, `test_to_webp_applies_exif_orientation`, `test_to_webp_thumb_fits_list_size`
  - Depends on: P6.A.4.2
  - Model: Opus
  - Acceptance:
    - `go test ./internal/ingest/covers/... -run 'TestToWebp|TestWebpSizeParity'` → the four ported tests `ok`; each fixture's 600 px (q78) and 96×144 (q70) outputs are within ±10 % of the Pillow byte size
    - `go test ./tests/ingest -run TestSecurity_A05_ImageBomb` → a 25+ MP or > 15 MB input is rejected before decoding
  - Notes: `contrib/engine` is [NOT IN VUKEM DOCS]. <https://vukem.dev/docs/deployment/single-binary> does not cover it, and SDD 06/11 cite only the local `vukem/docs/subsystems.md` (not a reference per AGENTS.md). Read the package godoc in the pinned vukem module. If `engine.Limits{CPUTime, AddressSpace}` exists, use it with 30 s / 512 MiB. Otherwise use `exec.CommandContext` with a 30 s timeout and no inherited env, and add a `vukem-feedback.md` entry. VF-004 workaround if `cwebp` is unavailable: keep JPEG and note it. One encode at a time (NFR-06-2).

- [ ] **P6.B.2.2** Port the covers job: source order (Open Library → TMDB → IGDB → Steam → GOG → Commons P3383, never P18), download via `httpclient.Safe`, sniff, EXIF, the two variants, the immutable key `covers/<id>-<sha256(source)[:16]>.webp` and `-t.webp` with `Put` on `drivers/s3`, skip on the same hash, and `covers.gc` 30-day deferred delete.
  - Files/modules: `internal/ingest/covers/job.go`, `internal/ingest/covers/pick.go`, `internal/ingest/covers/gc.go`, `internal/ingest/covers/job_test.go`
  - Refs: SDD 06 §7.5, NFR-06-7; SDD 11 §5 (bucket prefixes), open issues 11-1, 11-2, 06-4; ADR 0014 (amended); <https://vukem.dev/docs/digging-deeper/storage> "Reading and writing", "What the S3 disk does"; legacy `covers.py` `pick_source_cover`, `commons_cover`, `blob_pathname`, `store_cover`, `targets`, `run`, `drop_loose_commons_covers`; tests `test_store_cover_*`, `test_commons_cover_only_from_film_poster`, `test_only_a_full_run_removes_loose_commons_covers`, `test_run_isolates_per_item_failures`, `test_run_backfills_thumb_for_legacy_cover`, `test_run_skips_item_with_both_sizes_current`, `test_run_processes_items_concurrently`
  - Depends on: P6.B.2.1, P6.B.1.1, P6.B.1.2, P6.B.1.3
  - Model: Opus
  - Acceptance:
    - `go test ./internal/ingest/covers/... -run 'TestStoreCover|TestCommonsCover|TestRun'` → the ported tests `ok` against an in-memory disk
    - `go test ./tests/ingest -run TestCoversGC` → a replaced object is deleted only after 30 days (fake clock)
  - Notes: no bucket versioning on Railway; immutable keys plus deferred delete replace it (ADR 0014 as amended). `test_run_without_token_keeps_source_url` and `test_upload_rejects_unexpected_blob_response` are Vercel Blob-specific and not ported (SDD 06 §9). Never add a 300 px or 2× variant (open issues 09-9, 10-3).

- [ ] **P6.B.2.3** Port the trailers job: candidates (P1651 preferred, then TMDB/IGDB), oEmbed verdicts (200 available; 400/401/403/404 unavailable; 429/5xx no verdict), at most 2, reusing known verdicts, behind flag `youtube`.
  - Files/modules: `internal/ingest/trailers/trailers.go`, `internal/ingest/trailers/trailers_test.go`
  - Refs: SDD 06 §7.5 "Trailers", §7.6; legacy `trailers.py` `candidates`, `is_available`, `check`, `targets`, `fetch`; tests `test_candidates_order_wikidata_first_and_drop_malformed_ids`, `test_check_keeps_two_available_and_stops_asking`, `test_check_reuses_known_verdicts`, `test_rate_limit_is_not_a_verdict`, `test_fetch_writes_one_record_per_item_and_skips_failures`, `test_normalize_shows_only_checked_trailers_and_at_most_two`
  - Depends on: P6.B.1.1, P6.B.1.2
  - Model: Sonnet
  - Acceptance:
    - `go test ./internal/ingest/trailers/... -run 'TestCandidates|TestCheck|TestRateLimitIsNotAVerdict|TestFetch|TestNormalizeShowsOnlyCheckedTrailers'` → `ok`

#### P6.B.3 Editions and entity resolution

- [ ] **P6.B.3.1** Match editions (ADR 0012): ISBN strip, ISBN-10 → 13 with check digits, sources P212/P957, Wikidata editions (P747 with P407/P123/P577) and Open Library editions after A-05, ≤ 50 per work (pt/en/es first), MARC language or `und`, the primary edition rule, and `conflict` plus review when an ISBN belongs to another work. Then run the edition backfill.
  - Files/modules: `internal/ingest/editions/isbn.go`, `internal/ingest/editions/match.go`, `internal/ingest/editions/match_test.go`, `internal/ingest/backfill/backfill.go`, `tests/ingest/editions_test.go`
  - Refs: SDD 06 §7.4 "Editions", §7.8 (backfill keyset, 500 per transaction), §9 (books ≥ 60 % covers); SDD 01 §5.2 (`editions`), DQ-02, DQ-10; ADR 0012; D-10; open issue 01-1
  - Depends on: P6.A.4.2
  - Model: Opus
  - Acceptance:
    - `go test ./internal/ingest/editions/... -run 'TestIsbn10To13|TestIsbn13Checksum|TestPrimaryEdition|TestIsbnOnAnotherWorkIsNotMoved'` → `ok`
    - staging after `backfill.editions`: `psql -c "select round(100.0*count(*) filter (where cover_key is not null)/count(*)) from items where type='book'"` → ≥ 60 (only if A-05 allows Open Library; otherwise record the number and the reason)
  - Notes: `market_country = BR` from publish country `bl` is [NOT VERIFIED]; pin it with an Open Library fixture or leave it null. Edition contributors (translator, illustrator) are out (D-30).

- [ ] **P6.B.3.2** Resolve credits to entities: (1) QID; (2) source external id (`tmdb_person`, `tmdb_company`, `igdb_company`, `openlibrary_author`); (3) normalised name (NFKD, ASCII fold, lowercase, punctuation to space) + kind + year ±1. One candidate links, ≥ 2 (or a QID under another kind) opens `entity_match_reviews`, and none creates a native entity.
  - Files/modules: `internal/ingest/entities/match.go`, `internal/ingest/entities/normalize_name.go`, `internal/ingest/entities/match_test.go`
  - Refs: SDD 06 §7.4 "Matching per credit"; SDD 01 FR-01-4, DQ-08, DQ-09; D-01
  - Depends on: P6.A.3.4
  - Model: Opus
  - Acceptance:
    - `go test ./internal/ingest/entities/... -run 'TestNormalizeName|TestMatchByQid|TestMatchByExternalId|TestAmbiguousOpensReview|TestNoMatchCreatesNative'` → `ok`; re-running a batch creates no duplicate reviews (unique key)
  - Notes: fallback credits keep `item_entities.source` so the API filters them (ADR 0006).

- [ ] **P6.B.3.3** Build the match review queue at `/admin/ingest/reviews` with the actions `link`, `create` and `dismiss`, gated by `runIngestion` with CSRF.
  - Files/modules: `app/admin/ingest/reviews.go`, `resources/js/Pages/Admin/Ingest/Reviews.tsx`, `tests/admin_ingest_reviews_test.go`
  - Refs: SDD 06 §6 (reviews row), DQ-I-11; VF-008 (custom actions are app routes)
  - Depends on: P6.B.3.2, P4.B.1.1
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests -run TestAdminIngestReviews` → `link` sets the entity and status `linked`; a user without `runIngestion` → 403; a missing CSRF token → 419/403 per the vukem default
  - Notes: reuse the app-rendered admin shell from P4.B.1.1.

#### P6.B.4 Quality, admin and metrics

- [ ] **P6.B.4.1** Build the `dq.asserts` job: SDD 01 DQ-01…14 and SDD 06 DQ-I-01…13 as SQL checks writing `dq_results` (≤ 20 sample ids), run after each normalise and daily at 05:00, with `catalog_counts` snapshots and the `allow_count_drop` override.
  - Files/modules: `internal/ingest/dq/asserts.go`, `internal/ingest/dq/sql/*.sql`, `tests/ingest/dq_test.go`, `tests/ingest/testdata/dq/*.sql` (one faulty fixture per code)
  - Refs: SDD 06 §7.7, FR-06-9, §9; SDD 01 §7.6; ADR 0006 (DQ-I-12 serialiser check); D-36 (DQ-I-13)
  - Depends on: P6.B.2.2, P6.B.3.1, P6.B.3.2
  - Model: Opus
  - Acceptance:
    - `go test ./tests/ingest -run TestDQ` → each code fails on its faulty fixture and passes on the clean seed
    - latest run: `psql -c "select code from dq_results where run_id=$1 and not passed and severity='error'"` → no rows
  - Notes: DQ-I-04 calls `Stat` on 1,000 covers a day and on all covers weekly; a missing object matches `storage.ErrNotFound` (storage page). If the SQL files push the diff past ~400 lines, split into SDD 01 codes and SDD 06 codes.

- [ ] **P6.B.4.2** Build the `dq.frontend_fields` detector: daily, 500 items per type (top 200 + 300 random) through the item read-model builder, per-type required fields, nulls per type/field/locale into `field_coverage`, and ≤ 200 deduplicated `system` "missing field" suggestions a day.
  - Files/modules: `internal/ingest/dq/frontend_fields.go`, `tests/ingest/frontend_fields_test.go`
  - Refs: SDD 06 §7.7 "Missing-field detector"; SDD 09 (item read model); SDD 04 (`system` suggestions)
  - Depends on: P6.B.4.1, P4.B
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests/ingest -run TestFrontendFields` → a seeded movie without a director yields one `field_coverage` miss and one suggestion; a second run the same day adds no duplicate; the cap stops at 200

- [ ] **P6.B.4.3** Build `/admin/ingest`: last run per job, failures, failing asserts, oldest TMDB age and the coverage trend, plus `POST /admin/ingest/runs` (trigger), `POST /admin/ingest/runs/{id}/cancel`, and read-only panel resources for `job_runs`, `ingest_failures` and `dq_results`.
  - Files/modules: `app/admin/ingest/status.go`, `app/admin/ingest/runs.go`, `resources/js/Pages/Admin/Ingest/Index.tsx`, `tests/admin_ingest_test.go`, `qa/e2e/admin-ingest.spec.ts`
  - Refs: SDD 06 §6 (admin table), FR-06-8; <https://vukem.dev/docs/digging-deeper/admin>; VF-008; open issue 06-6
  - Depends on: P6.B.4.1, P4.B.1.1
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests -run TestAdminIngest` → 409 `run-active`, 423 `source-disabled`, 422 for > 500 qids; a trigger returns 303 and creates a `queued` run with `triggered_by = admin:<id>`
    - `bun run e2e -- admin-ingest` → Playwright + axe green
  - Notes: the triggers are app routes because custom panel actions are [NOT IN VUKEM DOCS] (VF-008). Keep the abilities `viewIngestion` and `runIngestion` distinct.

- [ ] **P6.B.4.4** Implement rollback of a bad run (restore history rows with `superseded_by_run`, clear `normalized_at`, delete inserted records, re-normalise, delete created items unless user data references them), `ingest.retry_failures` at 03:30, and the retention prunes at 02:30.
  - Files/modules: `internal/ingest/rollback/rollback.go`, `internal/ingest/run/retry_failures.go`, `internal/ingest/run/prune.go`, `app/admin/ingest/rollback.go`, `tests/ingest/rollback_test.go`
  - Refs: SDD 06 §7.8 "Bad run", §5 (retention: history 30 d, `job_runs` 180 d, `dq_results`/`field_coverage` 365 d), §6 (`rollbackIngestion`, 409 `history-expired`); ADR 0014
  - Depends on: P6.B.4.3
  - Model: Opus
  - Acceptance:
    - `go test ./tests/ingest -run 'TestRollbackRestoresPreviousState|TestRollbackKeepsItemsWithUserData|TestRollbackHistoryExpired|TestPruneRetention'` → `ok`; the rollback is logged as a `rollback` run plus an audit entry
  - Notes: beyond 30 days the path is the dump restore runbook (SDD 11), not code. Do not build it here.

- [ ] **P6.B.4.5** Add the pressure gate (in-flight count and last-minute item-page p95 hold batches above 16 or 300 ms), the `ingest.*` metrics on the global OTel meter, alert A7 and the ingestion dashboard, then wire the remaining weekly and daily schedules (Wed steam, Thu gog, Fri tmdb, Sat igdb, `tmdb.expire`, `igdb.refresh`, covers and trailer chains).
  - Files/modules: `app/http/middleware/inflight.go`, `internal/ingest/pressure.go`, `internal/ingest/metrics.go`, `routes/schedule.go`, `ops/grafana/ingestion.json`, `ops/grafana/alerts/a7.yaml`, `qa/k6/ingest-pressure.js`
  - Refs: SDD 06 §7.5 "Pressure gate", §7.6, §10 "Observability", NFR-06-1, NFR-06-2; SDD 12 §6.2, §7.5 (A6, A7); <https://vukem.dev/docs/digging-deeper/queue> "Metrics"; VF-019; open issues 06-7, 12-1
  - Depends on: P6.B.4.4
  - Model: Sonnet
  - Acceptance:
    - `go test ./internal/ingest/... -run 'TestPressureGateWaits|TestMetricsEmitted'` → `ok`
    - `k6 run qa/k6/ingest-pressure.js` on staging during the Friday TMDB run → item-page `p(95)` ≤ 1.2 × idle (SDD 06 §9)
  - Notes: VF-019 workaround per SDD 12 §6.2: `otel.GetMeterProvider().Meter("hexascore")` after `Config.Global`. Metric names exactly as SDD 12 (no underscore variants). Heavy runs start 00:00–06:00 America/Sao_Paulo.

#### P6.B.5 Release

- [ ] **P6.B.5.1** Release `v1.10.0` with every source on in staging, then production, and start the 30-day green-jobs clock for the phase exit.
  - Files/modules: `CHANGELOG.md`, env `FLAG_INGEST_*`, `docs/PROGRESS.md`
  - Refs: ADR 0013; SDD 06 §9; ROADMAP P6 exit criteria
  - Depends on: P6.B.4.5, P6.B.2.3, P6.B.3.3, P6.B.4.2
  - Model: Sonnet
  - Acceptance:
    - `curl -s https://staging.hexascore.app/up` → `v1.10.0`
    - every SDD 06 §9 box ticked with its decisive output pasted in `docs/PROGRESS.md`
  - Notes: IGDB stays site-only under D-36 until `docs/legal/igdb.md` exists. If it is still missing at release, say so in the changelog.

### P6.C Trending

#### P6.C.1 Activity primitives

- [ ] **P6.C.1.1** Create `item_activity_daily` and `trending_snapshots`, and extend the SDD 01 item merge to sum the source item's activity rows into the target.
  - Files/modules: `database/migrations/<ts>_create_trending_tables.go`, `app/catalog/merge.go`, `tests/recommend_merge_test.go`
  - Refs: SDD 08 §5, §8 (merged item); SDD 01 §7.5; open issue 08-3 (no exclusions table)
  - Depends on: P2.G
  - Model: Sonnet
  - Acceptance:
    - `vukem migrate && vukem migrate:status` → `create_trending_tables` Ran
    - `go test ./tests -run TestMergeSumsActivity` → the target's counts equal the sum, the source rows are gone
  - Notes: exclusion is `items.visibility <> 'public'` (SDD 01), not a table.

- [ ] **P6.C.1.2** Count views and search clicks in memory: middleware on 200 item-page GETs and `/app/v1` item reads, dedup by `cache.Add` with `HMAC-SHA256(daily_salt, visitor|item)` until 00:00 UTC, exclusions (prefetch, partial reloads, bots, > 50 items per visitor per day), the `SearchResultClicked` listener, and a buffered counter flushed by `s.Func("trending:flush", …).EveryMinute()` and on shutdown.
  - Files/modules: `app/recommend/primitives/views.go`, `app/recommend/primitives/counter.go`, `app/recommend/primitives/flush.go`, `routes/schedule.go`, `tests/recommend_views_test.go`
  - Refs: SDD 08 FR-08-1, NFR-08-4, NFR-08-5, §7.1, §8 (flush failure, 100k keys); <https://vukem.dev/docs/digging-deeper/cache>; <https://vukem.dev/docs/digging-deeper/scheduler> "Frequencies"; <https://vukem.dev/docs/digging-deeper/events>; VF-021; open issues 08-10, 10-5
  - Depends on: P6.C.1.1, P1.E
  - Model: Opus
  - Acceptance:
    - `go test ./tests -run 'TestViewsDedupPerVisitorPerDay|TestAppReadsDedupPerToken|TestNoRawIPStored|TestFlushIdempotentOnFailure'` → 5 GETs give `views = 1`; 5 app reads with one token give 1; a failed flush merges into the next one
    - `go test ./tests -run TestCounterDropsOldestAbove100k` → `trending.counter.dropped` incremented
  - Notes: VF-021 workaround (app `sync.Map` counter with a ticker-free flush on the scheduler). The scheduler page now documents `EveryMinute()` and `HourlyAt(m)`; append that finding to VF-021 in `vukem-feedback.md`, without deleting the entry. Hook shutdown with `app.OnShutdown` as SDD 08 §6 cites. Only `ItemID` from the click event is used.

- [ ] **P6.C.1.3** Build the rollup that replaces the durable columns for UTC today and today−1 from `rating_events`, status/time-to-beat events, `suggestion_events` (`submitted`) and `list_items.added_at`, with one count per user per (item, day, primitive) at the maximum SDD 02 weight and the public-list rules.
  - Files/modules: `app/recommend/trending/rollup.go`, `app/recommend/trending/rollup_test.go`
  - Refs: SDD 08 §7.1 (weights, lists), FR-08-2; SDD 02 §6 (events), §7.1, §7.5 (age tiers, shadow weight); SDD 14 §5 (`list_items.added_at`); SDD 04 (`suggestion_events`)
  - Depends on: P6.C.1.1, P4.A
  - Model: Opus
  - Acceptance:
    - `go test -tags=integration ./app/recommend/... -run 'TestRollupIdempotent|TestRollupWeights|TestListRules'` → running twice gives identical rows; a banned user weighs 0; a private list or an account younger than 7 days does not count
  - Notes: `*_w` columns are never exposed. Burst flags land within 24 h, inside the replace window.

#### P6.C.2 Score and snapshots

- [x] **P6.C.2.1** Implement the trend score, eligibility and the prior fill as pure functions: `D(a) = 2^(−a/7)`, `Q_d`, `U_d`, `T_i`, `A_i`, eligibility `A ≥ 5 ∧ T ≥ 30 ∧ released ∧ public`, and `F_i`.
  - Files/modules: `app/recommend/trending/score.go`, `app/recommend/trending/score_test.go`
  - Refs: SDD 08 §7.2, §7.3 step 3, FR-08-3, §10 "Unit"
  - Depends on: P6.C.1.3
  - Model: Sonnet
  - Acceptance:
    - `go test ./app/recommend/... -run 'TestTrendScore|TestFallback'` → golden `118.97` and `72.8` (±0.01), `D(0,7,14) = 1, 0.5, 0.25`, thresholds at 4.99/5.0, `F_i` 6.89 and 5.30 for the §7.3 examples
  - Notes: weights and thresholds are untuned (open issue 08-7). Keep them as named constants, not configuration.

- [ ] **P6.C.2.2** Build `ComputeTrending` (`Cron("5 * * * *")`, `WithoutOverlapping()`): rollup, then the per-type snapshot in one transaction ordered `(T desc, popularity_prior desc, id)` with gapless ranks and prior fill, then `trending:v1:{type}` invalidation. Add `PruneTrending` at 04:45 (90 days), the queued `trending_rebuild` on `user.banned`, the `trending.*` metrics and alert A14.
  - Files/modules: `app/jobs/compute_trending.go`, `app/jobs/prune_trending.go`, `app/recommend/trending/snapshot.go`, `app/recommend/trending/rebuild.go`, `routes/schedule.go`, `ops/grafana/alerts/a14.yaml`, `app/recommend/trending/snapshot_test.go`
  - Refs: SDD 08 §6 "Jobs", "Metrics", §7.3, FR-08-4, FR-08-6, FR-08-7, NFR-08-2, §9, §10; SDD 12 §6.2, §7.5 (A14); <https://vukem.dev/docs/digging-deeper/scheduler> "Overlapping runs"; <https://vukem.dev/docs/digging-deeper/queue>
  - Depends on: P6.C.2.1, P6.C.1.2
  - Model: Opus
  - Acceptance:
    - `go test ./app/recommend/... -run TestSnapshotDeterminism -count=3` → `ok`
    - `go test -tags=integration ./app/recommend/... -run TestBurstDoesNotEnterTop10` → `ok`; a Go benchmark on seeded Postgres with 20k active items → ≤ 30 s
  - Notes: whole-day decay keeps reruns within an hour identical. On failure readers keep the last snapshot; never write a partial one. `HourlyAt(5)` is equivalent now that it is documented; either is fine, so keep the SDD's `Cron`.

#### P6.C.3 Surfaces

- [ ] **P6.C.3.1** Render the home trending sections: an eager SSR Inertia prop `trending` from `cache.NewMemory` (one query on a miss), the listing pattern, the "Popular" chip and title rule, and `?from=trending` links that the canonical URL drops.
  - Files/modules: `app/web/home.go`, `resources/js/Pages/Home.tsx`, `resources/js/components/TrendingSection.tsx`, `tests/home_trending_test.go`, `qa/e2e/home.spec.ts`, `qa/k6/home.js`
  - Refs: SDD 08 §6 "Inertia props", FR-08-5, FR-08-6, FR-08-11, NFR-08-1; SDD 09; D-02; D-07; open issue 08-8; <https://vukem.dev/docs/digging-deeper/cache>
  - Depends on: P6.C.2.2
  - Model: Sonnet
  - Acceptance:
    - `bun run e2e -- home` → Playwright + axe green in three locales; a type with no released items has no section
    - `k6 run qa/k6/home.js` → `p(95) < 200ms` at 50 rps with a warm cache (SDD 08 §9)
  - Notes: this replaces the legacy `SectionHighlights` on the home page. `forYou` arrives in P6.D.2.2.

- [ ] **P6.C.3.2** Close P3.A.1.4 and the trending part of P3.C.2.1: serve `listTrending` and the business-only `getItemActivity` on `/api/v1`, and `/app/v1/{type}/trending`, with `web.ETagOf`, `c.NotModified` and `Cache-Control: public, max-age=300`.
  - Files/modules: `app/api/v1/trending.go`, `app/appapi/catalog.go`, `app/recommend/trending/read.go`, `tests/api_trending_test.go`, `openapi/api-v1.json`, `openapi/app-v1.json`
  - Refs: SDD 08 §6 (contracts table), §9; SDD 05 §6.2, FR-05-17; P3.A.1.4; <https://vukem.dev/docs/rest-apis/overview> "Cross-cutting"
  - Depends on: P6.C.2.2, P3.A.1.3, P3.C.2.1
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests -run 'TestAPITrending|TestAPIActivityPlan'` → pro key on `/activity` → 403 `plan-required`; business key → 200, counted as 1 request
    - `curl -sI http://localhost:3000/api/v1/games/trending -H "Authorization: Bearer $KEY"` → `ETag`; with `If-None-Match` → 304; `vukem openapi:export | diff - openapi/api-v1.json` → empty after the regenerated document is committed
  - Notes: P3.A.1.4 specifies `?window=day|week` and the top 100, while SDD 08 snapshots hold ≤ 10 ranks over a 30-day window. **Architect input needed:** default to SDD 08 (10 ranks, no `window`) and fix SDD 05 §6.2 in the same change, or extend the snapshot. Tick P3.A.1.4 here. Regenerate `@hexascore/app-client` so the mobile home (P5.B.1.1) gets data.

#### P6.C.4 Release

- [ ] **P6.C.4.1** Release `v1.11.0` (API `API_VERSION` minor bump for `listTrending`/`getItemActivity`).
  - Files/modules: `CHANGELOG.md`, `.env` `API_VERSION`, docs changelog entry
  - Refs: ADR 0013; SDD 05 §6.5; SDD 08 §9
  - Depends on: P6.C.3.1, P6.C.3.2
  - Model: Sonnet
  - Acceptance:
    - `psql "$DATABASE_URL" -c "select item_type, count(*) from trending_snapshots where computed_at = date_trunc('hour', now()) group by 1"` (staging) → 4 rows of 10
    - `schemathesis run openapi/api-v1.json --url https://staging-api.hexascore.app -H "Authorization: Bearer $HX_TEST_KEY" --checks all` → 0 failures on `listTrending` and `getItemActivity`, and the responses contain no TMDB/IGDB fields
  - Notes: the "hype" list for unreleased items (open issue 08-13) is post-launch and not part of this release.

### P6.D Similar items and personal recommendations

#### P6.D.1 Similar items

- [ ] **P6.D.1.1** Create `item_similar` and `user_recommendations`, and implement feature extraction (`(role, entity)` pairs, cast `position < 5`, franchise parents at half weight), the `g_f` weights and weighted-Jaccard `sim` as pure functions.
  - Files/modules: `database/migrations/<ts>_create_similar_tables.go`, `app/recommend/similar/features.go`, `app/recommend/similar/sim.go`, `app/recommend/similar/sim_test.go`
  - Refs: SDD 08 §5, §7.4, §10 "Unit"; SDD 01 §5.2 (`item_entities`)
  - Depends on: P6.C.4.1, P6.B.3.2
  - Model: Opus
  - Acceptance:
    - `vukem migrate && vukem migrate:status` → `create_similar_tables` Ran
    - `go test ./app/recommend/... -run TestSimilarity` → the Zelda example gives 0.70; `sim` is symmetric and in [0, 1]
  - Notes: `users.id` is text (SDD 03 §5); `user_recommendations.user_id` follows it.

- [ ] **P6.D.1.2** Build `ComputeSimilar` (`DailyAt("02:00")`, `WithoutOverlapping()`): candidates sharing a feature with `n_f ≤ 5 %·N_t`, the keep rule, ordering by `round(sim,2)`, list `cooccur`, popularity and id, ≤ 20 neighbours, `Chunk(ctx, 500, …)` with a 50 ms pause, skipping unchanged `feature_hash`, and the site and API variants.
  - Files/modules: `app/jobs/compute_similar.go`, `app/recommend/similar/job.go`, `app/recommend/similar/job_test.go`, `app/recommend/similar/bench_test.go`
  - Refs: SDD 08 §7.4, FR-08-8, NFR-08-2, NFR-08-3; open issues 08-9, 08-11; <https://vukem.dev/docs/database/orm> "Querying"; <https://vukem.dev/docs/digging-deeper/scheduler>
  - Depends on: P6.D.1.1
  - Model: Opus
  - Acceptance:
    - `go test -tags=integration ./app/recommend/... -run 'TestSimilarJob|TestSimilarSkipsUnchanged|TestApiVariantUsesAllowlistedSources'` → `ok`
    - `go test ./app/recommend/similar -bench BenchmarkComputeSimilar -benchtime 1x` on a 500k-item seed → ≤ 60 min on 1 core, recorded in SDD 08 §11 issue 11
  - Notes: if the benchmark exceeds 60 min, switch to incremental by `feature_hash` in a follow-up microtask (open issue 08-11). Do not tune weights here.

- [ ] **P6.D.1.3** Serve `similar` (8, with `reason`) and `crossMedia` (≤ 6, `item_links` in either direction, hidden when empty) as eager props on the item page and on `/app/v1/items/{id}/related`, with `reco.impressions`/`reco.clicks` and `?from=similar|crossmedia`.
  - Files/modules: `app/web/item.go`, `resources/js/Pages/Item/Show.tsx`, `resources/js/components/SimilarItems.tsx`, `app/appapi/related.go`, `tests/item_related_test.go`, `openapi/app-v1.json`
  - Refs: SDD 08 §6 (contracts, Inertia props, metrics), FR-08-8, FR-08-11; SDD 01 §5.2 (`item_links`); SDD 10 §6.1 (Item row `/items/{id}/related`); SDD 12 §6.2
  - Depends on: P6.D.1.2
  - Model: Sonnet
  - Acceptance:
    - `go test ./tests -run 'TestItemRelated|TestAppRelated'` → 8 similar items with a reason; cross-media hidden when empty; 404 for an unknown id
    - `bun run e2e -- item` → Playwright + axe green
  - Notes: the mobile item screen gains these blocks in a store release, not OTA (SDD 10 §7.8). Open that app task after this one. Similar items are not in `/api/v1` (SDD 08 §2).

#### P6.D.2 "For you"

- [ ] **P6.D.2.1** Implement the user profile and scoring as pure functions: `μ_u`, `ρ_ui`, `τ_ui`, `p_u[f]`, `s(u,j)`, the two candidate sources, removal of rated, statused, unreleased or excluded items, the greedy re-ranker (≤ 20, ≤ 3 per franchise, ≤ 4 per primary genre), explanations, and the onboarding fallback below 3 positive items.
  - Files/modules: `app/recommend/foryou/profile.go`, `app/recommend/foryou/rank.go`, `app/recommend/foryou/explain.go`, `app/recommend/foryou/foryou_test.go`
  - Refs: SDD 08 §7.5, FR-08-9, §10 "Unit"; A-02
  - Depends on: P6.D.1.2
  - Model: Opus
  - Acceptance:
    - `go test ./app/recommend/... -run TestProfile` → `ρ` clamped to [−1, 1]; caps hold; each item has one explanation "Because you rated {title} {stars}"; < 3 positives → `basis:"onboarding"`

- [ ] **P6.D.2.2** Serve "For you": the `recommend_user {UserID, Type}` queue job (`Retries(2)`, `Timeout(30s)`), stale-while-refresh (row > 24 h or ≥ 5 newer `rating_events`), a synchronous compute when there is no row, the opt-out (D-29) that deletes rows and stops jobs, `/app/v1/me/for-you`, the home `forYou` `Defer` group and the `ForYou/Index` page.
  - Files/modules: `app/jobs/recommend_user.go`, `app/recommend/foryou/service.go`, `app/appapi/foryou.go`, `app/web/foryou.go`, `resources/js/Pages/ForYou/Index.tsx`, `tests/foryou_test.go`, `qa/k6/foryou.js`
  - Refs: SDD 08 §6 (contracts, jobs, privacy), §7.5, FR-08-10, NFR-08-6, §8 (opt-out while queued); SDD 03 FR-03-19; D-29; <https://vukem.dev/docs/frontend/inertia>; <https://vukem.dev/docs/digging-deeper/queue>
  - Depends on: P6.D.2.1
  - Model: Opus
  - Acceptance:
    - `go test ./tests -run 'TestForYouOptOut|TestForYouStaleWhileRefresh|TestRecommendJobExitsWhenOptedOut'` → opting out deletes `user_recommendations` and returns no list; a queued job for an opted-out user exits without writing
    - `k6 run qa/k6/foryou.js` → `p(95)` ≤ 400 ms on a miss, ≤ 30 ms on a hit
  - Notes: the privacy policy text and the D-21 legal review are owner items (D-29). Confirm the policy states personalised recommendations before the flag goes on in production. The mobile "For you" screen is not in SDD 10 §6.1; do not add it without an SDD change.

#### P6.D.3 Evaluation

- [ ] **P6.D.3.1** Write the offline evaluation protocol: temporal hold-out of each user's latest 20 % of ratings (min 2), overall ≥ 70 as positives, precision@10 and nDCG@10 with a 95 % bootstrap CI and coverage, the monthly CF trigger check (≥ 5,000 users with ≥ 10 counted ratings), and the report file.
  - Files/modules: `app/recommend/eval/eval.go`, `app/recommend/eval/eval_test.go`, `app/jobs/reco_eval_monthly.go`, `docs/reco-eval/README.md`
  - Refs: SDD 08 §7.6, §11 issue 7 (90-day review); A-02; VF-029 (export helpers only write to `http.ResponseWriter`)
  - Depends on: P6.D.2.2
  - Model: Opus
  - Acceptance:
    - `go test ./app/recommend/eval/... -run 'TestHoldOutSplit|TestPrecisionAt10|TestNDCGAt10|TestBootstrapCI'` → `ok` on a hand-computed fixture
    - `go test ./tests -run TestCFTriggerCheck` → below 5,000 qualifying users the job records "not triggered" and computes nothing else
  - Notes: the content baseline is recorded at launch + 90 days, which is an owner calendar item, not code. CF ships only through a new ADR with ≥ 10 % precision@10 lift, a CI excluding 0 and coverage ≥ 80 %. VF-029 workaround: write the CSV report with `encoding/csv` and reuse the formula guard rule, not the response helpers.

#### P6.D.4 Release

- [ ] **P6.D.4.1** Release `v1.12.0` and close Phase 6 once daily jobs have been green for 30 days and trending and similar are live.
  - Files/modules: `CHANGELOG.md`, `docs/PROGRESS.md`, `ROADMAP.md` status table
  - Refs: ADR 0013; SDD 08 §9; SDD 06 §9; ROADMAP P6 exit criteria
  - Depends on: P6.D.1.3, P6.D.3.1, P6.B.5.1
  - Model: Sonnet
  - Acceptance:
    - `curl -s https://hexascore.app/up` → `v1.12.0`
    - `psql "$DATABASE_URL" -c "select job, count(*) from job_runs where status='failed' and created_at > now()-interval '30 days' group by 1"` → no rows for the scheduled jobs
  - Notes: the owner pushes the tag. Every SDD 08 §9 box must be ticked with its decisive output.

## Phase 7: social layer (post-1.0)

Goal: turn the library into a social network for games, movies, series and books, with Skoob, Goodreads and Letterboxd as references. D-06 was amended on 2026-10-31 (owner, Q-A1 in `docs/orchestration/QUESTIONS.md`): feed, follows, reviews, comments and reactions enter right after the cutover (`v1.0.0`); this phase starts then. SDD 14 §1 already reserves the room: "the model leaves room for follows and feed (D-06) without tables for them".

Design principles for the whole phase: every social write is moderated by the P4.B queue (reports, hide, ban); private profiles never leak through the feed, counts or reactions; LGPD basis is consent for the feed (opt-in at onboarding, off by default for existing users) and legitimate interest for aggregates (D-29); nothing social enters the public API before a licence decision (A-07).

### P7.A Follows and activity feed

- [ ] **P7.A.1.1** Amend D-06, write SDD 14 §12 "Social layer" (follows and feed subsection) (tables `follows`, `activity_events`, privacy matrix, retention 180 days) and ADR 0019 (feed built from the write path, no fan-out until 10k followers).
  - Files/modules: `docs/brainstorming.md`, `docs/sdd/14-user-library.md`, `docs/adrs/0019-social-feed.md`
  - Refs: SDD 14 §1, §2; D-06; D-29; web-research §6 (Goodreads), §7 (Skoob); Letterboxd: DOC-research-letterboxd report
  - Depends on: P2.G.4.2
  - Model: Opus
  - Acceptance:
    - `bunx markdownlint-cli2@0.24.0 docs/sdd/14-user-library.md docs/adrs/0019-social-feed.md` → 0 errors; the open-issues table lists the new rows

- [ ] **P7.A.1.2** Build `follows`: `PUT|DELETE /u/{nickname}/follow`, follower and following counts on the profile, `TestSecurity_A01_PrivateProfileFollowRequiresApproval`.
  - Files/modules: `database/migrations/<ts>_create_follows_table.go`, `app/social/follows.go`, `app/controllers/follows_controller.go`, `routes/social.go`, `resources/js/pages/User/Profile.tsx` (follow button, counts), `tests/follows_test.go`
  - Refs: SDD 14 §12; SDD 13 A01, A06 (rate limit 60/h)
  - Depends on: P7.A.1.1
  - Model: Sonnet
  - Acceptance:
    - `go test -tags integration ./tests -run 'TestFollows|TestSecurity_A01_PrivateProfileFollow'` → green
    - Playwright `social.spec.ts` → follow, unfollow and counts in the three locales

- [ ] **P7.A.1.3** Emit activity events from the existing write paths (rating saved, status changed, list created or item added, review published) and build `/feed` (followed users, 20 per page, cursor) plus the "recent activity" tab on the public profile.
  - Files/modules: `app/social/activity.go` (listener on `scoring`, `library` events), `app/controllers/feed_controller.go`, `resources/js/pages/Feed/Index.tsx`, `resources/js/components/social/ActivityCard.tsx`
  - Refs: SDD 14 §12; SDD 08 §7.1 (primitives shared with trending, P6.C.1); SDD 09 §7.3 (cursor)
  - Depends on: P7.A.1.2, P6.C.1.1
  - Model: Opus
  - Acceptance:
    - `k6 run qa/k6/feed.js` → p95 ≤ 300 ms at 30 rps with 10k events
    - `go test -tags integration ./tests -run TestFeedPrivacy` → private profiles and hidden lists never appear in another user's feed

- [ ] **P7.A.1.4** Release `v1.13.0` (follows and feed).
  - Depends on: P7.A.1.3
  - Model: Sonnet
  - Acceptance: `gh release view v1.13.0` → non-empty body

### P7.B Reviews, comments and reactions

- [ ] **P7.B.1.1** Reviews: optional text (≤ 5,000 chars, Markdown subset, spoiler flag) attached to a rating; shown on the item page (sorted by helpfulness, then date) and on the profile.
  - Files/modules: `database/migrations/<ts>_create_reviews_table.go`, `app/social/reviews.go`, `app/controllers/reviews_controller.go`, `resources/js/components/item/Reviews.tsx`, `resources/js/components/item/ReviewForm.tsx`
  - Refs: SDD 14 §12; SDD 13 A05 (sanitised Markdown, no raw HTML); A-07 (licence of contributions); Letterboxd review model: DOC-research-letterboxd report
  - Depends on: P7.A.1.1
  - Model: Opus
  - Acceptance:
    - `go test -tags integration ./tests -run 'TestReviews|TestSecurity_A05_ReviewMarkdown'` → green; a `<script>` in a review renders as text
    - Playwright `reviews.spec.ts` → write, edit, delete, spoiler toggle; axe 0 serious

- [ ] **P7.B.1.2** Comments on reviews and lists (one level of threading, 60/h per user) and reactions (`like`, `helpful`) on reviews, comments and lists, with counts updated in the same transaction.
  - Files/modules: `database/migrations/<ts>_create_comments_and_reactions.go`, `app/social/comments.go`, `app/social/reactions.go`, controllers and routes, `resources/js/components/social/CommentThread.tsx`, `ReactionBar.tsx`
  - Refs: SDD 14 §12; SDD 13 A06 (rate limit); SDD 14 §7.5 (notifications: "someone commented")
  - Depends on: P7.B.1.1
  - Model: Sonnet
  - Acceptance:
    - `go test -tags integration ./tests -run 'TestComments|TestReactions'` → green, including the unique (user, target) reaction and the rate limit
    - `/me/notifications` lists comment and reaction notifications

- [ ] **P7.B.1.3** Moderation: `reports.target_type` gains `review` and `comment`; the admin queue hides, deletes or bans; the author is notified.
  - Files/modules: `app/requests/report.go`, `app/library/reports.go`, admin resources, `resources/js/components/ReportDialog.tsx`
  - Refs: SDD 14 FR-14-9; SDD 04 (moderation SLA); P4.B
  - Depends on: P7.B.1.2, P4.B.1.1
  - Model: Sonnet
  - Acceptance:
    - `go test -tags integration ./tests -run TestReportsReviewComment` → green; a hidden review disappears from the item page and the feed

- [ ] **P7.B.1.4** Release `v1.14.0` (reviews, comments, reactions).
  - Depends on: P7.B.1.3
  - Model: Sonnet
  - Acceptance: `gh release view v1.14.0` → non-empty body

### P7.C Social discovery

- [ ] **P7.C.1.1** "Rated by people you follow" on the item page and the home, and the friends' average next to the community score.
  - Files/modules: `app/catalog/item_page.go`, `app/controllers/home.go`, `resources/js/components/item/FriendsScore.tsx`
  - Refs: SDD 08 §6; SDD 14 §12
  - Depends on: P7.A.1.3
  - Model: Sonnet
  - Acceptance:
    - `go test -tags integration ./tests -run TestFriendsScore` → only followed, non-private users count; `Defer` group `personal` keeps the public HTML cacheable

- [ ] **P7.C.1.2** Share cards: Open Graph image per item, list and profile (Go `image/draw`, cached on the bucket), and a "share" action on those pages.
  - Files/modules: `app/seo/og_image.go`, `routes/web.go` (`/og/{kind}/{id}.png`), `resources/js/components/Seo.tsx`
  - Refs: SDD 09 §7.7; VF-004 (no WebP encoder: PNG)
  - Depends on: P2.H.1.1
  - Model: Sonnet
  - Acceptance:
    - `curl -sI localhost:3000/og/items/<id>.png` → `image/png`, `Cache-Control: public, max-age=86400`; the item page `<meta property="og:image">` points at it

- [ ] **P7.C.1.3** Release `v1.15.0` and close Phase 7 once the moderation SLA covers reviews and comments for 30 days.
  - Depends on: P7.C.1.2, P7.B.1.4
  - Model: Sonnet
  - Acceptance: `gh release view v1.15.0` → non-empty body; moderation median ≤ 24 h over 30 days
