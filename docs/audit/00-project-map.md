# 00. Project map (architect audit, 2026-11-02)

Executive summary: HexaScore is a single Go process on the vukem framework (v0.37.0) that serves the web app, the paid API, the first-party mobile API, the queue workers and the cron scheduler, backed by one Postgres and one S3 bucket on Railway. The codebase is far more mature than the audit prompt assumed: ingestion connectors for Wikidata, Wikipedia, Open Library, TMDB, IGDB, Steam and GOG exist behind kill switches (all OFF in production), trending has a designed formula (SDD 08) but the landing page uses a different, simpler query, and Telegram alerting exists for errors but not as an operational digest. Everything below cites a file, a command or a tool result; items marked "not verified" were not checked.

Sources: three read-only exploration passes (hexascore-vukem, vukem, worktrees), `route:list` and `schedule:list` run inside the QA container built from `main` 5e92fcc7, Railway CLI/MCP read calls, curl against production.

## 1. Repositories and workspace

| Path | What | Git |
| --- | --- | --- |
| `hexascore-vukem/` | The product (this repo), Go module `github.com/AlexGalhardo/hexascore-app`, Go 1.27.3 | `main` = `origin/main` at 5e92fcc7 (2026-11-02); latest tag v0.38.0 |
| `vukem/` | The framework, developed in parallel, consumed only through pinned module versions (no `replace`, no `go.work`, `go.mod:6-9`) | `main` clean; latest tag v0.37.0 |
| `wt/` | 16 worktree directories from past agent sessions plus three `.esp-rm-*.log` files | see `docs/audit/CLEANUP.md` |
| `legacy/` | Not present on this machine (AGENTS.md names it as a read-only Next.js reference; history on branch `legacy`) | n/a |
| root files | `PROMPT.md`, `PROMPT-ARCHITECT.md`, `PROMPT-WORKERS.md`, `questions-to-dev.md`, `developer-manual-todos.md`, `cloudflare.md`, `hexascore-spike.bundle` (203 KB), `apply-button.png`, `.caveman.json` | not in any repo |

Uncommitted changes found in `hexascore-vukem` at audit start (left untouched): `D docs/review/2026-11-01-architect-brief.md`, `M tsconfig.json` (drops `vite/client` types), `?? resources/js/vite-env.d.ts`, `?? local.properties` (Android `cmake.dir`, not gitignored).

## 2. Runtime architecture

```text
                     ┌──────────────────────────── Railway project "hexascore" ────────────────────────────┐
  hexascore.app      │  service "app" (1 replica, us-east4)                                                 │
  www / api.* ──────►│  /app/app (Go, cmd/app/main.go)                                                      │
  (Railway edge,     │   ├─ HTTP: web (Inertia+React SSR), /api/v1 (paid), /app/v1 (mobile), /admin         │
   gzip, TLS)        │   ├─ queue worker: queues default,ingest, concurrency 3, ingest cap 2 (config/queue.go)│
                     │   ├─ scheduler: ~45 tasks (routes/schedule.go), zone America/Sao_Paulo for night jobs │
                     │   └─ node SSR child supervised by Go (SSR_SUPERVISE=true, Dockerfile:63)             │
                     │        │                                   │                                          │
                     │        ▼                                   ▼                                          │
                     │  Postgres 18 (10 GB volume, 5.6 GB used)   bucket hexascore-production (covers, OG,   │
                     │  jobs, cache, sessions, rate limits,       backups, exports)                          │
                     │  catalogue, social, billing                                                           │
                     └───────────────────────────────────────────────────────────────────────────────────────┘
  External: Wikidata/Wikipedia/Open Library/TMDB/IGDB/Steam/GOG/YouTube oEmbed (ingest, OFF in prod),
            Stripe (billing), Resend (mail), Google OAuth, Cloudflare Turnstile + DNS, Telegram Bot API,
            Umami + GA4, IndexNow + Bing Webmaster.
```

- One process does everything (D-08 in `docs/brainstorming.md:52`; `.env.example:64`). Splitting a worker service is "the first thing to do when the ceiling is raised".
- Dockerfile: bun 1.4.4 assets stage, golang 1.27.3 build stage, node 26.13.0-alpine runtime with `postgresql18-client` and `libwebp-tools` (`Dockerfile:5,26,40-50`). No `railway.toml`; Railway config lives in the dashboard: pre-deploy `/app/app migrate`, healthcheck `/up`, tracing enabled, 3 custom domains (MCP `get-service-config`, 2026-11-02).
- Discrepancy: SDD 11 §7 and AGENTS.md say the pre-deploy entry point is `hexascore:predeploy` (drops INVALID indexes, then migrates); production runs plain `migrate` (`docs/sdd/11-infra-deploy.md:188`, MCP result above).
- Environments: `production` and `staging` exist on Railway; `staging.hexascore.app` does not resolve in DNS (curl 2026-11-02: "Could not resolve host"), so the staging runbook's owner DNS step is still pending.

## 3. Code layout (hexascore-vukem)

| Dir | Purpose |
| --- | --- |
| `cmd/app/` | Only `main`; any argument is a vukem console command |
| `bootstrap/` | Wiring (`app.go`): driver, migrations, jobs, schedule, flags, Inertia SSR, extra middleware (Inflight, APIEdge, CanonicalHost, Version, Noindex), health checks, log redaction, Telegram |
| `app/` | 37 packages: accounts, admin, alerts, api, appapi, backup, billing, botverify, catalog, challenge, community, console, controllers, forms, http/middleware, i18n, jobs (32 handlers), library, mail, models, moderation, notifications, policies, privacy, recommend (primitives, trending, similar, foryou, eval), requests, scoring, seo, social, ssr, webhooks, wellknown |
| `internal/ingest/` | Connectors (wikidata, wikipedia, openlibrary, tmdb, igdb, steam, gog, trailers), `httpx` (own throttle/retry/breaker), `store` (source_records with SHA-256 hash), `normalize`, `covers` (cwebp), `dq`, `editions`, `entities`, `expire`, `rollback`, `run` (flags, continuation), `backfill` |
| `internal/search/` | Postgres full-text search, suggest sync, dirty queue |
| `routes/` | One file per area; order in `routes/registry.go:9-38`; scheduler in `routes/schedule.go` |
| `config/` | Typed env config: api, billing, catalog, mail, queue, seo |
| `database/` | 82 Go migrations + registry, seeders (DemoCatalog), factories, legacy import/export |
| `resources/` | React 19 + Inertia 3 SPA (`js/`, 38 pages), Tailwind 4, `lang/{en,es,pt-BR}.json` (1453 leaf keys each, parity test), `views/` (shell, errors, mail), `docs/` (rendered to JSON) |
| `apps/mobile/` | Expo SDK 58 Android app using `/app/v1` through `packages/app-client` |
| `packages/`, `sdks/` | `app-client`, `i18n`, `tokens`; public SDKs `ts/` and `python/` (unpublished) |
| `openapi/` | Golden documents `api-v1.json`, `app-v1.json` |
| `tests/` (183 files), `qa/` | Go HTTP/integration tests; Playwright (33 specs), k6, Lighthouse, Schemathesis, ZAP |
| `scripts/` | `ci/local.sh` (release gate), release, openapi-check, infra, ops, legacy |
| `docs/` | 23 ADRs, 14 SDDs + `open-issues.md`, 22 runbooks, legal, setups, design-system, orchestration (BOARD, QUESTIONS) |
| `.orchestration/` | Task board (683 lines), briefs, reports, handoffs (gitignored) |

## 4. HTTP surface

`route:list` inside the container: 408 lines; 223 GET, 108 POST, 24 DELETE, 16 PUT, 8 PATCH. By first path segment: `/admin` 129, `/app` 53, `/api` 35, `/profile` 16, `/items` 11, `/lists` 8, `/user` 7, `/me` 6.

| Area | Prefix / examples | Auth | File |
| --- | --- | --- | --- |
| Public pages | `/`, `/{games,movies,series,books}`, `/{section}/{slug}`, entity pages `/{section}/{kind}/{slug}`, `/search`, `/about`, `/terms`, `/privacy`, `/pricing`, `/docs`, `/u/{nickname}`, `/contact`, `/sitemap.xml`, `/robots.txt`, `/llms.txt`, `/covers/{key...}`, `/og/{kind}/{file}` | none; scrape guard, Turnstile step-up | `routes/web.go:246-347`, `routes/library.go`, `routes/reports.go` |
| Auth and account | `/login`, `/signup`, passkeys, `/auth/google/*`, `/user/two-factor*`, `/profile*`, `/billing*`, `/profile/export` | guest or session | `routes/auth.go`, `social.go`, `two_factor.go`, `profile.go`, `billing.go` |
| User activity | `PUT/DELETE /items/{id}/{rating,status,time-to-beat,review,edition}`, `/lists*`, follows, comments, reactions, `/feed`, `/me/notifications`, `/reports`, suggestions (flag) | session, verified, per-user limits | `routes/ratings.go`, `status.go`, `lists.go`, `follows.go`, `community.go` |
| Admin | `/admin/*` (vukem panel + users, scoring, reports, items, moderation, ingest runs/reviews, flags) | role admin, IP allowlist, TOTP, idle 30 min | `routes/admin.go:44-232` |
| Paid API | `api.hexascore.app/v1`: `/{type}`, `/{type}/{id}`, `/{type}/trending`, `/search`, `/lookup`, `/changes`, `/exports`, `/webhooks*`, `/me/usage` | bearer API key, quota | `routes/api_v1.go:114-185` |
| Mobile API | `/app/v1`: tokens (PoW), catalog, search, me, library, lists, ratings (idempotency middleware), notifications, client-errors/metrics | bearer app token | `routes/app_*.go` |
| Webhooks in | `POST /stripe/webhook` (before session/CSRF), `POST /csp-reports`, `POST /turnstile` | signature | `routes/billing.go:31-34` |
| Health | `/up`, `/up?checks=1` (db, ssr, storage, scheduler) | none | `bootstrap/app.go:229-261` |

Production headers observed (curl 2026-11-02): gzip, CSP with nonce + `strict-dynamic`, HSTS 2 years, COOP/CORP, `X-Frame-Options: DENY`, rate-limit headers (60/min), `Cache-Control: public, max-age=60, stale-while-revalidate=300`, `x-render-cache: HIT`; home HTML 266 KB raw / 44 KB gzip; hreflang en/es/pt-BR/x-default; unknown paths answer 301 to `/`.

## 5. Data model (Postgres, 82 migrations)

| Area | Tables (migration) |
| --- | --- |
| Users and auth | `users` (+ username, locale, role, banned_*, deletion_scheduled_at, profile_private, legacy_id, is_seed, stripe_customer_id), `remember_tokens`, `access_tokens`, `password_resets`, `social_accounts`, `passkeys`, `account_events`, `admin_audit_logs`, `sessions` |
| Catalogue | `data_sources`, `items` (uuid, type CHECK, slug, `wikidata_qid` UNIQUE, release_*, visibility, trailers/external_ids/provenance jsonb, cover keys), `item_translations` (locale CHECK en/pt-BR/es), `item_alt_titles`, `entities`, `entity_translations`, `item_entities`, `entity_links`, `item_links`, `entity_stats`, `editions` (ISBN-13 function), `catalog_redirects`, `item_genre_sources`, `source_records`, `source_record_history` |
| Offers | `affiliate_links` (item_id + store, url, active) only; no price or availability tables |
| Scoring | `ratings` (c1..c6, generated `overall`), `rating_events`, `rating_flags`, `item_status`, `time_to_beat`, `score_actions`, `score_bursts`, `score_formula_versions`, `score_priors` |
| Social | `follows`, `reviews`, `comments`, `reactions`, `activity_events`, `lists`, `list_items`, `user_notifications`, `reports` |
| Trending and reco | `item_activity_daily` (item × day counters, weighted), `trending_snapshots` (type, rank 1..10, basis), `item_similar`, `user_recommendations` |
| Billing and API | `subscriptions`, `api_keys`, `api_quota_counters`, `api_usage_daily`, `api_credit_balances`, `credit_ledger`, `pix_charges` (legacy), `api_changes`, `webhook_endpoints`, `webhook_cursor`, `webhook_deliveries` |
| Search | tsvector columns, trigram indexes, `search_dirty`, `search_suggest`, `search_events` |
| Community | `suggestions`, `suggestion_events`, `suggestion_reviews`, `catalog_audit`, `catalog_locks`, `catalog_merge_audit`, `contributor_stats`, `community_honeypot_hits` |
| Ingestion ops | `job_runs`, `ingest_failures` (parked after 5 failed runs), `dq_results`, `field_coverage`, `catalog_counts`, `entity_match_reviews`, `cover_gc`, `catalog_changed_done`, `indexnow_cursor`, `legacy_import_runs` |
| Framework | `vukem_jobs`, `vukem_failed_jobs`, `vukem_cache`, flags, admin |

Production volume: 98,841 items, 2.1 M ratings, 2.05 M statuses, 677 k list items, 40 k entities, 10,003 users of which 10,000 are flagged seeds (architect brief 2026-11-01; `questions-to-dev.md` items 13-16). Disk 5.55 GB of 10 GB (Railway metrics 7 d: max 6.3 GB).

## 6. Background work

- Queue: vukem `queue.Database` (`FOR UPDATE SKIP LOCKED`), queues `default,ingest`, concurrency 3, ingest cap 2 (`.env.example:66-67`, `config/queue.go:4`). Retries per handler (ingest stages `Retries(3)`, `Timeout(10m)`, `internal/ingest/run/continuation.go:264`); `queue.ErrPermanent` used 24 times; dead letter `vukem_failed_jobs` plus `ingest_failures`.
- Scheduler (`routes/schedule.go:28-112`): trending flush every minute and `ComputeTrending` hourly at :05; scores 04:00; search suggest sync every minute; privacy purges 03:00-03:15; prunes 04:30-04:50; webhooks and IndexNow fan-out every minute; exports 07:00; backups 06:00 daily and monthly off-site; ingest chains (wikidata incremental 00:30, one weekly full per source, `tmdb-expire`, `igdb-refresh` 6 h, covers daily/weekly, `covers-gc`, dq, `retry-failures`, `alerts` hourly). All `WithoutOverlapping()`.
- Ingestion kill switches: `ingest_enabled` plus `ingest_<source>` flags, default false (`internal/ingest/run/flags.go:10-37`). Production variable names from Railway contain no `FLAG_INGEST_*` and no TMDB/IGDB credentials: ingestion is OFF in production. Staging has `FLAG_INGEST_ENABLED` defined (value not read).
- Back-pressure gate pauses ingestion under HTTP load (`internal/ingest/pressure.go`).
- Outbound HTTP for ingest uses its own throttle/retry/breaker (`internal/ingest/httpx/`), deliberately not vukem's `httpclient` transports (`client.go:3-6`). Rates per host in `httpx/hosts.go` match SDD 06 §7.2.

## 7. Trending today (verified)

- Designed path (SDD 08 §7.2-7.3): `ComputeTrending` rolls `item_activity_daily` up and writes the 30-day top 10 per type into `trending_snapshots`; readers are `/api/v1/{type}/trending` and for-you (`app/recommend/trending/snapshot.go`, `app/controllers/publicapi/trending.go:60`).
- Landing page path: `app/controllers/home.go:22-27,240-264` ignores `trending_snapshots` and selects the 12 most-rated items per type in the last 30 days by `sum(ratings_new_w)`, falling back to `items.rating_count`; cached 5 min under `home:trending30:v1:{locale}`.
- Dead code: `trending.CacheKey` (`snapshot.go:251`) is forgotten by the hourly job but nothing reads it.
- Consequence: the home list is a "most rated" list, not the designed decayed trend score; with 10,000 seed users still weighting ratings it is biased toward legacy popularity. RFC-002 addresses this.

## 8. Images, i18n, SEO, security, observability

- Covers: S3 bucket via `vukem/drivers/s3`; keys `covers/{section}/{slug}[-t].webp` plus content-hash keys; `cwebp` child process 600×600 q78 and 96×144 q70 (`internal/ingest/covers/webp.go:38-61`); served through `/covers/{key...}` with long cache; `COVERS_BASE_URL` can point at a CDN (unset in production). Known shortcuts: GC race (`gc.go:25`), cwebp without address-space limit (`webp.go:60`), OG cards never garbage-collected (`app/seo/og_image.go:73`).
- i18n: en, pt-BR, es JSON catalogues with parity test; only allow-listed namespaces are sent to the browser (`routes/web.go:160-180`); URL prefixes `/pt-br`, `/es`.
- SEO: sitemap index with 4 item sitemaps, robots.txt allows AI crawlers and disallows `?cursor=`, hreflang, canonical, OG tags, IndexNow job, `llms.txt`.
- Security: argon2id with legacy scrypt compat, TOTP, passkeys, Google OAuth, origin CSRF + admin token fallback, rate limits in the database store, scrape guard + Turnstile + proof-of-work + honeypot (ADR 0016, 0022), log redaction (`app/logging/redact.go`), admin guard (IP allowlist, TOTP, idle/absolute timeouts). No secret found in committed files; `.env` gitignored.
- Observability: slog JSON, Railway logs/metrics/traces only (D-35, no `/metrics`), `/up?checks=1`, Telegram notifier for alert lines and errors with 30-min dedup and a daily 08:00 digest (`app/alerts/`), hourly `ingest:alerts`. No periodic operational digest of ingestion throughput (RFC-001 adds it).

## 9. Environment variables (names only)

Production service `app` defines 45 variables (Railway `get-service-config`): `APP_*`, CACHE_STORE, CLIENT_IP_HEADER, DATABASE_URL, GA_MEASUREMENT_ID, `UMAMI_*`, BING_WEBMASTER_API_KEY, INDEXNOW_KEY, LOG_FORMAT, MAIL_MAILER, OTEL_LOGS_EXPORTER, OTEL_METRICS_EXPORTER, RATE_LIMIT_STORE, `S3_*` (7), SCHEDULE_TIMEZONE, SESSION_DRIVER, `SOCIAL_GOOGLE_*`, `STRIPE_MODE` + `STRIPE_*_LIVE` and the legacy unsuffixed `STRIPE_KEY`/`STRIPE_PRICE_*`/`STRIPE_WEBHOOK_SECRET`, TELEGRAM_BOT_TOKEN, TELEGRAM_CHAT_ID, TRUSTED_PROXIES, TRUST_FORWARDED_PROTO, TURNSTILE_*, ANDROID_CERT_SHA256, ORCHESTRATOR_ACCESS_CHECK.

Missing in production and relevant to this audit: `FLAG_INGEST_*`, `THEMOVIEDATABASE_API_READ_ACCESS_TOKEN`, `IGDB_TWITCH_CLIENT_*`, `INGEST_CONTACT_URL`, `COVERS_BASE_URL`, `TELEGRAM_ERROR_CHAT_ID`, `ADMIN_IP_ALLOWLIST`, `SEO_NOINDEX` (correctly absent). Full catalogue of names: `.env.example` (169 non-empty lines). Read in code but absent from `.env.example`: `APP_MIN_SUPPORTED_VERSION`, `APP_VERSION`, `GOOGLE_JWKS_URL`, `TURNSTILE_VERIFY_URL`, `POSTGRES_URL`.

## 10. Coupling to vukem

- Pinned: `vukem v0.37.0`, `contrib/stripe v0.5.1`, `drivers/postgres v0.6.1`, `drivers/s3 v0.4.2`. Import counts: database 210, vukemtest 127, web 114, schema 93, pgtest 88, auth 82, errs 73, queue 65, inertia 41, router 36, orm 36, ratelimit 35, storage 34, i18n 29.
- Not used although available: `drivers/otel`, `drivers/redis`, `contrib/imaging`, `contrib/engine`, `notify`, `httpclient` resilience transports (app has its own), `cache.Buffer` is used for view counters.
- vukem gaps relevant to the catalogue worker (framework inventory 2026-11-02): no unique or idempotent jobs, queue `Exponential` backoff has no jitter (`vukem/queue/queue.go:134-143`), no WebP/AVIF encoder (`contrib/imaging` outputs JPEG/PNG), no image hashing, no Telegram `notify.Channel`, no Prometheus exporter. All open VF entries were closed on 2026-11-01 (`vukem-feedback.md` reads "(none open)").
- vukem unmerged local branches: `agent/implementer/railway-observability` (1 ahead, 41 behind), `fix/railway-otlp-traces` (1 ahead), `chore/nested-v0.37.0` (1 ahead, nested go.mod bump), `docs/readme-and-updates` (4 ahead, includes VF-077/078 fixes), `release/v0.37.0` (5 ahead, changelog edits). Open PR: #108 dependabot only.

## 11. Tooling, tests and CI

- Local: `docker compose up -d --wait` (Postgres 18.6 on 5433, RustFS 9000), `vukem serve`, `go test ./...`, `go test -tags integration ./...`, `bun run lint|typecheck|test`, `bun run e2e`, `scripts/ci/local.sh` (every CI job locally). Verified on this host: go1.27.3, bun 1.4.4, golangci-lint 2.15.0, Docker 29.9.2.
- Tests: 342 Go test files (190 integration-tagged, 88 using pgtest, ~969 test functions), 83 TS test files, 56 mobile Jest files, 33 Playwright specs, k6, Schemathesis, ZAP, Lighthouse budgets (LCP 2500 ms, script 170 KB).
- CI (`.github/workflows/ci.yml`): lint, test, sdk, migrations round trip, vuln, deps, setup-windows, build + image check + `/up` smoke, openapi (diff, Redocly, Schemathesis), lighthouse, mobile; e2e only when `RUN_E2E=true`. Other workflows: release (ghcr image + GitHub release), dast, scorecard, changelog-sync, projected-spend, sdk-release, mobile-release, mobile-gradle.
- Build-context defect found while building the QA image: `.dockerignore` does not exclude `**/.venv/`, `wt/`, `.ci-local/`, `.hypothesis/`, `.lighthouseci/`; a `uv` venv at `sdks/python/.venv` (Linux-style junction `lib64`, symlink `bin/python`) broke `docker compose build` twice on Windows until the venv was moved out of the tree.

## 12. Cost and capacity baseline (Railway, 7-day window ending 2026-11-02)

| Metric | app | Postgres |
| --- | --- | --- |
| CPU avg / max (vCPU) | 0.004 / 0.74 | 0.006 / 0.77 |
| Memory avg / max (GB) | 0.11 / 0.82 | 0.26 / 4.03 |
| Disk (GB) | n/a | 5.55 current, 6.31 max, 10 GB volume |
| Network TX avg (GB per sample) | 0.001 | n/a |

Workspace usage month-to-date: US$8.36; soft limit US$15; hard limit US$20 (`railway usage limit status`; D-27 as amended 2026-11-01). The Postgres memory peak (4 GB) coincides with the legacy import and the set-based entities stage, not with steady state.

## 13. Process and team state

- ROADMAP: 343 boxes, 43 open (`grep -c '^- \[ \]' ROADMAP.md`); almost all owner gates. Open code-only item: rest of P6.B.3.1 (editions).
- `.orchestration/queue/TASKS.md` still lists dozens of rows as `ready` that were merged days ago; the board is stale.
- Temporary flow Q-A21: direct pushes to `main` deploy production without a PR (`AGENTS.md`, Rules).
- Owner decisions governing this audit: D-08 (single service), D-27 (US$20 hard limit), D-35 (Railway-native observability), D-36 (IGDB interim rule), ADR 0003 (Postgres only), ADR 0006 (data licensing posture), ADR 0007 (ingestion in Go).

## 14. Gate for the next phases

The map is complete for the questions the later phases need. Known unknowns, recorded in `docs/OPEN-QUESTIONS.md`: Railway start command for a future worker service, staging DNS, whether ingestion should be enabled in production before the worker RFC lands.
