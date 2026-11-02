# 02. Technical review (2026-11-02)

Executive summary: the codebase is in good shape for its age: pinned dependencies, strong HTTP security defaults, trigger-maintained search, a durable Postgres queue, 969 Go test functions and a CI that cannot pass vacuously. The review found one P1 bug that was in production (render cache ignores the locale prefix, fixed in this audit), one P1 reliability defect (three long nightly jobs run under vukem's default 1-minute timeout, so the nightly score recompute cannot finish), a cluster of P2 data-retention and LGPD-erasure gaps, and operational mismatches between docs and the Railway dashboard. No secret was found in the repository or its history snapshot; no SQL injection path; CSRF bypasses are limited to signature- or rate-limit-protected endpoints.

Method: two read-only review passes over `app/`, `internal/`, `routes/`, `database/migrations/`, `bootstrap/`, `.github/workflows/`, `Dockerfile`, `.dockerignore`, plus targeted verification by the architect (code reads, `go test`, `golangci-lint`, curl against production, Railway CLI/MCP reads). Each finding cites file:line. "Not checked" at the end lists what this review did not cover; a production `queue:failed` read was denied by the tool permission classifier and is left to the owner.

Severity: P0 blocks launch or leaks data · P1 serious · P2 medium · P3 debt or cosmetic.

## 1. Findings

### Security (OWASP)

| ID | Sev | Finding | Evidence | Fix |
| --- | --- | --- | --- | --- |
| T-01 | ok | CSRF bypass list is limited to `POST /turnstile`, `/csp-reports`, `/stripe/webhook`, `/app/v1/client-errors`, `/app/v1/client-metrics` | `routes/web.go:76-82`; Stripe handler authenticates with `Stripe-Signature` before any write (`app/billing/billing.go:106,147`); telemetry routes carry per-IP rate limits (`routes/app_telemetry.go:32-36`); turnstile step-up can only grant the victim a clearance cookie (`routes/web.go:68-69`) | none |
| T-02 | ok | No SQL built from user input: the `fmt.Sprintf` in `app/catalog/listing.go:273-277` interpolates `expr`, `dir`, `spec.cast` from static sort specs, values go through placeholders; no `WhereRaw` or string-concatenated `ExecContext` in `app/`, `internal/`, `database/` | grep 2026-11-02 | none |
| T-03 | ok | Admin gating: role `admin` + verified + not banned, IP allowlist, mandatory TOTP, idle 30 min / absolute 12 h; non-admins get 404, guests 302 to login | `app/http/middleware/admin_guard.go:75-80`, local test `/admin` → `/profile#two-factor` | none |
| T-04 | ok | Secrets: none in committed files; `.env` gitignored; log redaction of e-mails, keys, bearer tokens, signed URLs | `app/logging/redact.go:45-120`; grep of token shapes hit only test fixtures | none |
| T-05 | P2 | Account erasure leaves user-linked rows: `score_bursts.user_id`, `score_actions.user_id`, `time_to_beat.user_id`, `search_events.user_id` have no FK to `users` and are not touched by `purgeUserData`; `rating_flags` not purged either | `database/migrations/20261029120000_create_scoring_tables.go:85,109,137`, `20261029140008:15`, `app/accounts/deletion.go:310-357` | add the tables to the purge (null or delete `user_id`), add FKs with `ON DELETE SET NULL`, test `TestSecurity_A03_PurgeRemovesScoringRows` |
| T-06 | P3 | Admin IP allowlist variable is absent in production (`ADMIN_IP_ALLOWLIST` not among the 45 names) | Railway `get-service-config` | set it, or document that TOTP alone is the second factor |

### Correctness and reliability

| ID | Sev | Finding | Evidence | Fix |
| --- | --- | --- | --- | --- |
| T-07 | P1 (fixed) | Render cache served one locale's HTML for every locale of a path (production `/pt-br/games/burgertime` → English). vukem's `i18n.StripPathPrefix` rewrites the path before the cache keys it | `app/http/middleware/render_cache.go:119-134`, `vukem/i18n/path.go:27-47`, `vukem/http.go:314-316`; web report W-01 | fixed in commit 695872f5 (`i18n.PathPrefix(ctx)` in the key, test `TestRenderCacheKeysStrippedLocalePrefix`); deploy and re-verify in production |
| T-08 | P1 | Nightly `scoring.recompute_scores`, daily `recommend.compute_similar` and monthly `recommend.eval_monthly` are registered with `queue.Handle` and no `Timeout`, so vukem bounds their context to the default 1 minute and cancels them (`context.WithTimeout`); `RecomputeAll` walks every item (`BatchSize` 500, one statement per item, ~30-48 ms measured in P1.D.2.2), which is ~50 minutes for 98,841 items. Each night the job is cut, retried 3 times and the same first ~2,000 items are recomputed; scores of the rest never refresh (the legacy import skipped the `finalise` stage and relied on this job, `questions-to-dev.md` item 14) | `app/jobs/recompute_scores.go:19-31`, `app/jobs/compute_similar.go:19`, `app/jobs/reco_eval_monthly.go:23`, `vukem/queue/queue.go:272,553-559`, `app/scoring/maintenance.go:47-66` | set `queue.Timeout(2*time.Hour)` and `queue.Retries(1)` on the three handlers (and `Timeout(30*time.Minute)` on `deletion_purge`, `cover_fetch`, `fanout_webhooks`, `check_sources`), make `RecomputeAll` resumable (persist the last id in `score_formula_versions` or a `job_cursors` row) and chunk it into 15-minute jobs; owner verifies with `railway ssh … /app/app queue:failed` and a one-off `scoring:recompute` run |
| T-09 | P2 | Hourly `ComputeTrending` forgets cache keys `trending:v1:<type>:<locale>` that nothing reads; the home page reads `home:trending30:v1:<locale>` and expires by TTL only | `app/jobs/compute_trending.go:40-46`, `app/recommend/trending/snapshot.go:251`, `app/controllers/home.go:70` | RFC-002 makes the home read `trending_snapshots`; until then delete the dead `Forget` |
| T-10 | P2 | API quota meter keeps unflushed deltas in process memory (1-minute flush, fail-closed after 10 minutes); with two replicas each process admits the full quota | `app/api/quota.go:17-27,88` (shortcut comment) | ADR 0024 keeps one replica; backlog item "multi-replica readiness": atomic `UPDATE … RETURNING` counter |
| T-11 | P2 | IGDB OAuth token fetched on every `Fetch` call | `internal/ingest/igdb/igdb.go:328,347` | cache on the connector with `expires_in`, refresh on 401 (RFC-001 CW-04) |
| T-12 | P2 | Retry side effects not proven for `fanout_webhooks` and `cover_fetch` (default 3 retries, no idempotency key per delivery) | `app/jobs/fanout_webhooks.go:20`, `app/jobs/cover_fetch.go:29` | add a delivery claim (`UPDATE webhook_deliveries SET state='sending' WHERE id=$1 AND state='pending'`) and a content-hash check before upload; tests for a retried job |
| T-13 | P2 | Covers GC race and `cwebp` without address-space limit (documented shortcuts) | `internal/ingest/covers/gc.go:25`, `internal/ingest/covers/webp.go:60` | RFC-001 CW-07 (`contrib/engine` limits, GC claim row) |
| T-14 | P3 | Alert codes A1, A2, A3, A6, A8, A12 of SDD 12 are referenced in `app/alerts` but no emitter outside the package was found (A5 SSR fallback, A7 ingestion, A10 emitted) | grep `alerts\.A<n>` 2026-11-02; needs reconciliation with SDD 12 §6 | map each code to an emitter or mark as "not implemented" in SDD 12 |
| T-15 | P3 | `/up?checks=1` "storage" probes the local disk, not the bucket | `bootstrap/app.go:239` (shortcut comment) | HEAD a known key on the bucket with a 2 s timeout |

### Database

| ID | Sev | Finding | Evidence | Fix |
| --- | --- | --- | --- | --- |
| T-16 | P2 | `item_status` has PK `(user_id,item_id)` and index `(user_id,status)` only; item-level counts scan | `database/migrations/20261029120000_create_scoring_tables.go:46,53` | `CREATE INDEX CONCURRENTLY item_status_item_status_idx ON item_status (item_id, status)` |
| T-17 | P2 | No retention: `contact_messages`, `admin_audit_logs`, `suggestion_events`, `ingest_failures` (only removed on successful retry), `backup_runs` | migrations `20261029140013:9`, `20261029130000:10`, `app/community/state.go:182`, `20261101120000:52`, `20261029180000:10`; prunes in `routes/schedule.go` cover rating events, api_changes, webhook deliveries, activity, exports, sessions, cache, search events only | add prune jobs with documented windows (contact 180 d, audit logs 2 y or permanent by decision, suggestion_events 1 y, ingest_failures 90 d, backup_runs 1 y) and list them in SDD 11 §5 |
| T-18 | P2 | Postgres volume at 5.55 GB of 10 GB with ingestion off; RFC-001 projects 9 GB at 250 k items | Railway metrics 7 d | resize to 15 GB before M1 (owner), ADR 0025 storage measures |
| T-19 | P3 | Pre-deploy mismatch: Railway runs `/app/app migrate`; docs and `hexascore:predeploy` expect the INVALID-index cleanup first | MCP `get-service-config`; `docs/sdd/11-infra-deploy.md:188`; AGENTS.md | OQ-01 |

### Build, CI and operations

| ID | Sev | Finding | Evidence | Fix |
| --- | --- | --- | --- | --- |
| T-20 | P2 (fixed) | `.dockerignore` sent `wt/`, `.ci-local/`, `.hypothesis/`, `.lighthouseci/`, nested `.env*`, venvs and `node_modules` to the builder; `COPY . .` in the build stage copies them into a layer; a `uv` venv broke the Windows build | `Dockerfile:31`; build logs 2026-11-02 | fixed in commit b965c5d5 |
| T-21 | P2 | `bun audit --ignore` lists four GHSA ids without date or reason; `osv-scanner.toml` ignores carry reasons but no expiry handling beyond Q-A13 | `.github/workflows/ci.yml:222`, `osv-scanner.toml:6-29` | move ignores to one dated allowlist with a review date; fail CI when the date passes |
| T-22 | P2 | Playwright e2e does not gate CI or releases (`RUN_E2E` unset) and has drifted: the account helper still fills the sign-up name field removed in 64f9b814, so every signed-in spec times out (web report W-16) | `.github/workflows/ci.yml:326-327`; `qa/e2e/helpers/account.ts:73` | fix the helper, then OQ-14: run the `desktop-en` project on release candidates (under 15 min) |
| T-23 | P2 | Staging DNS not configured; `staging` environment exists with its own variables | curl NXDOMAIN; `railway environment list`, variable names | OQ-02 |
| T-24 | P3 | Ingestion disabled in production without a documented date to enable; TMDB/IGDB credentials absent | Railway variable names | RFC-001 M0 |
| T-25 | P3 | Task board `.orchestration/queue/TASKS.md` lists dozens of merged rows as `ready` (683 lines) | file vs `git log` | rewrite with the new PROMPT-WORKERS board protocol |
| T-26 | ok | Dockerfile images pinned by tag and digest; CI has no `continue-on-error`, no "or true" shell suffix and no `if: false`; migrations round-trip job; govulncheck; osv; Schemathesis; Lighthouse | `Dockerfile:5,26,40`; `ci.yml` grep | none |

### Dead code and legacy

| ID | Sev | Finding | Evidence | Fix |
| --- | --- | --- | --- | --- |
| T-27 | P3 | `trending.CacheKey` and the hourly `Forget` are dead (see T-09) | as T-09 | remove with RFC-002 |
| T-28 | P3 | Legacy routes kept for the cutover: entity redirects, `/api/search/suggest`, `/profile/api/checkout` 301 | `routes/web.go:343-347`, `routes/billing.go:45` | keep until the 30-day retention window closes (ADR 0013), then remove with tests |
| T-29 | P3 | Legacy billing tables `credit_ledger`, `pix_charges` | `database/migrations/20261029200000` | drop two releases after cutover (expand/contract policy) if nothing reads them (not verified) |
| T-30 | P3 | Working tree carried a stale `tsconfig.json` edit, an untracked `vite-env.d.ts` and the deleted architect brief | `git status` | OQ-10 |

## 2. Verified good practices (keep)

- Security headers and CSP with nonce and `strict-dynamic`; scrape guard, Turnstile, proof-of-work, honeypot; argon2id with legacy scrypt upgrade; passkeys; log redaction; signed URLs for exports and backups; `TestSecurity_*` convention.
- Postgres-only architecture with trigger-maintained search, generated `ratings.overall`, `SKIP LOCKED` queue, advisory-locked migrations, concurrent index policy, encrypted backups with restore drills.
- Pinned dependencies everywhere (Go, bun, Docker digests, CI actions); release flow with git-cliff and sync checks.
- Feature flags and kill switches for every risky subsystem (ingest per source, community, social).

## 3. Not checked

Auth session internals and IDOR on every `/app/v1` route (spot-checked ownership on lists and ratings only), open redirects, header injection, JSON body limits per route, `migrate:rollback` of every migration, every job's idempotency, `_ =` swallowed errors (29 sites counted, not classified), `.env.example` names unused in code, unreferenced packages and scripts, `credit_ledger`/`pix_charges` readers, production `queue:failed` output (permission denied to the audit tool), mobile app code.

## 4. Changes made in this audit (branch `audit/architect-2026-11-02`)

| Commit | Change | Verification |
| --- | --- | --- |
| 695872f5 | `fix(web): key the render cache by the stripped locale prefix` | `go test ./app/http/middleware/` ok, `golangci-lint` 0 issues; new test fails without the fix (stack with `StripPathPrefix`) |
| b965c5d5 | `chore(build): keep tool state, worktrees and venvs out of the Docker context` + `local.properties` ignored | QA image rebuilt from the tree |

Remaining P1 (T-08) and the P2 retention/erasure items are backlog tasks with acceptance criteria in `docs/BACKLOG.md`, not opportunistic fixes, because they touch scoring semantics and LGPD behaviour that need a test plan and an owner read.
