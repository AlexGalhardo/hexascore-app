# Backlog (architect audit, 2026-11-02)

Priorities: P0 blocks the public 1.0 · P1 serious · P2 medium · P3 debt. Estimates in senior-days (Sonnet-class worker) unless marked Intern. Every task needs: context, scope and out of scope, verifiable acceptance, test plan, risk and rollback (PROMPT-WORKERS.md ticket format). References point to the finding or RFC that holds the detail. Owner-gated rows wait for the answer in `docs/OPEN-QUESTIONS.md`.

## P0

| ID | Task | Est. | Depends | Owner | Acceptance | Source |
| --- | --- | --- | --- | --- | --- | --- |
| B-001 | Deploy the render-cache locale fix (commit 695872f5) to production and verify `/pt-br/games/<slug>` answers `lang="pt-BR"` from a cold and a warm cache | 0.5 | release gate | Senior | curl shows `lang="pt-BR"` with `x-render-cache: HIT`; `TestRenderCacheKeysStrippedLocalePrefix` in CI | W-01, T-07 |
| B-002 | Nightly score recompute: `queue.Timeout(2h)` + `Retries(1)` on `scoring.recompute_scores`, `recommend.compute_similar`, `recommend.eval_monthly`; resumable cursor in `RecomputeAll`; one-off production run after deploy (owner starts it) | 2 | — | Senior | `go test ./app/scoring ./app/jobs`; staging run logs `recompute scores items=<all>`; `queue:failed` empty the next morning | T-08 |

## P1

| ID | Task | Est. | Depends | Owner | Acceptance | Source |
| --- | --- | --- | --- | --- | --- | --- |
| B-010 | Explicit `Timeout`/`Retries` on every handler in `app/jobs/registry.go` (table in the ticket), idempotent retry for `fanout_webhooks` (delivery claim) and `cover_fetch` (hash check) | 2 | — | Senior | tests for a retried delivery (sent once) and a retried cover (stored once) | T-08, T-12 |
| B-011 | LGPD erasure gaps: purge or null `score_bursts`, `score_actions`, `time_to_beat`, `rating_flags`, `search_events` user links; FKs with `ON DELETE SET NULL` | 2 | — | Senior | `TestSecurity_A03_PurgeRemovesScoringRows`; SDD 03 §erasure table updated | T-05 |
| B-012 | Fix the e2e account helper (name field removed in 64f9b814), run `desktop-en` green locally, re-enable the CI e2e job for release candidates | 1.5 | D-42 (approved) | Senior | `bunx playwright test --project=desktop-en` green; CI job runs on `release/*` | W-16, T-22 |
| B-013 | Worker service on Railway (RFC-001 CW-01, CW-02) | 4 | D-40 (approved) | Senior | see RFC-001 | RFC-001 |
| B-014 | Pre-deploy command `hexascore:predeploy` on Railway (done 2026-11-02), staging DNS CNAME (done; TLS certificate still provisioning at 20:15 UTC) | 0 | — | Owner | `curl https://staging.hexascore.app/up` 200 | T-19, T-23 |

## P2

| ID | Task | Est. | Depends | Owner | Acceptance | Source |
| --- | --- | --- | --- | --- | --- | --- |
| B-020 | Mobile item page order: title and synopsis above the rating widget at < 768 px | 1 | — | Senior | 360 px screenshot shows H1 within the first 1.5 screens; Lighthouse a11y 100 kept | W-02 |
| B-021 | 404/410 pages instead of 301 to home for unknown slugs, users and lists; keep `catalog_redirects` for merges | 1 | D-41 (approved) | Senior | `curl /games/does-not-exist` → 404 with the HTML page; e2e `not-found.spec.ts` updated | W-03 |
| B-022 | `og:image` on item pages (cover 600 px or the OG card), `twitter:card=summary_large_image` | 1 | — | Senior | curl shows `og:image`; Playwright `seo.spec.ts` asserts it | W-04 |
| B-023 | Search page: render results server-side on the first request (drop the client shell for `/search`) or keep the shell but add `<noscript>` results link | 1 | — | Senior | first `curl /search?q=x` contains result links | W-05 |
| B-024 | Retention prunes: `contact_messages` 180 d, `admin_audit_logs` (decision), `suggestion_events` 1 y, `ingest_failures` 90 d, `backup_runs` 1 y; SDD 11 §5 table | 1.5 | — | Senior | prune jobs scheduled and tested; table in SDD 11 | T-17 |
| B-025 | Index `item_status (item_id, status)` (concurrent migration) | 0.5 | — | Senior | migration round trip; `EXPLAIN` on the item count query uses it | T-16 |
| B-026 | Dependency ignore allowlist with dates and review deadline (`bun audit --ignore`, `osv-scanner.toml`), CI fails when a date passes | 1 | — | Senior | CI red on an expired ignore in a test branch | T-21 |
| B-027 | IGDB token cache with expiry | 0.5 | — | Senior | one token request per run (test with fake server) | T-11 |
| B-028 | Trending v2 (RFC-002 TR-01..TR-04) | 6.5 | D-40 (seeds excluded) | Senior | see RFC-002 | RFC-002 |
| B-029 | Catalogue worker streams (RFC-001 CW-03..CW-11) | 22 | B-013 | Senior | see RFC-001 | RFC-001 |
| B-030 | Offers v1 (RFC-003 OF-01..OF-05, OF-07) | 10 | D-40, B-013; affiliate tags after enrolment | Senior | see RFC-003 | RFC-003 |
| B-031 | ADR 0025 storage measures on `source_records` (lz4, partitions, cold export) | 2 | — | Senior | size drops ≥ 30 % on a staging copy | ADR 0025 |
| B-032 | Multi-replica readiness note: quota meter atomic counter, scheduler `OnOneServer` plan (no code until ADR 0024 trigger) | 0.5 | — | Senior | `docs/runbooks/scaling.md` | T-10, ADR 0024 |
| B-033 | Cleanup remainder per `docs/audit/CLEANUP.md`: B5/B6 Temp worktrees (owner command), B8 review of the 14 kept branches (cherry-pick `P2.F-data-migration` S3 retry and `manual-access` CI gate if still relevant, delete the rest), D1 Docker prune | 1 | — | Senior | `git worktree list` shows only the main checkout; branches listed in CLEANUP.md E | CLEANUP |
| B-034 | Task board rewrite (`.orchestration/queue/TASKS.md`) to the new PROMPT-WORKERS protocol; archive merged rows | 0.5 | — | Intern | board has only open rows | T-25 |

## P3

| ID | Task | Est. | Depends | Owner | Acceptance | Source |
| --- | --- | --- | --- | --- | --- | --- |
| B-040 | Listing: empty page beyond the last page shows a message or redirects to the last page; canonical keeps filters | 0.5 | — | Senior | tests on `/games?…&page=99` | W-06, W-07 |
| B-041 | Translate genre labels on item pages in pt-BR/es (reuse `app/catalog/genres.go` lists) | 0.5 | — | Intern | `/pt-br/games/<slug>` shows "Mistério" | W-08 |
| B-042 | Home `<title>` with tagline in 3 locales; pricing plural "1 request per second"; listing count lowercase noun; search trigger accessible name | 0.5 | — | Intern | strings updated in the 3 catalogues; Lighthouse a11y keeps 100 | W-09..W-12 |
| B-043 | `/up?checks=1` storage probe against the bucket | 0.5 | — | Senior | check fails when the bucket is unreachable (test with fake endpoint) | T-15 |
| B-044 | Alert codes reconciliation with SDD 12 (emit or mark not implemented) | 1 | — | Senior | table in SDD 12 §6 matches grep | T-14 |
| B-045 | Remove legacy routes and `credit_ledger`/`pix_charges` after the retention window (ADR 0013) | 1 | 30-day window | Senior | route inventory test updated; migration drops tables | T-28, T-29 |
| B-046 | Documentation to English: remaining Portuguese files (`questions-to-dev.md`, `developer-manual-todos.md`, `cloudflare.md` at the workspace root; `docs/orchestration/QUESTIONS.md` Portuguese answers) | 2 | — | Intern (Haiku) | files rewritten in English, same facts, links intact | PROMPT-ARCHITECT §2 |
| B-047 | vukem feedback entries: unique/idempotent jobs, backoff jitter, Telegram `notify.Channel`, document that `StripPathPrefix` runs before app server middleware | 0.5 | — | Senior | entries in `vukem-feedback.md` with the next free VF number | RFC-001, T-07 |
| B-048 | `ADMIN_IP_ALLOWLIST` in production or documented exception | 0 | — | Owner | variable set | T-06 |
