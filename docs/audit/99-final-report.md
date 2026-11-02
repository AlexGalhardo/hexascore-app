# 99. Closing report of the architect audit (2026-11-02)

Executive summary: HexaScore is a working, well-engineered product that is already in production; what keeps it from a stable public v1.0 is operational, not functional. In one session the audit mapped the system, tested the web app end to end, reviewed the code and infrastructure, fixed the one user-facing P1 that was live (wrong language served from the render cache), found a second P1 that explains stale scores (nightly job killed by a 1-minute timeout), wrote the three RFCs the owner asked for with a budget that fits US$20/month until ~500 k items, decided Redis (deferred) and MongoDB (rejected) with measured data, rewrote the team protocol, cleaned the worktree sprawl that was safe to clean, and left every irreversible decision to the owner in `docs/OPEN-QUESTIONS.md`.

All work is on branch `audit/architect-2026-11-02` (base `main` 5e92fcc7). Nothing was pushed or deployed.

## 1. What was done

| Phase | Output | Evidence |
| --- | --- | --- |
| 0 Reconnaissance | `docs/audit/00-project-map.md` | three read-only exploration passes, `route:list` (408 routes) and `schedule:list` from the built image, Railway CLI/MCP reads, curl |
| 1 Web test | `docs/audit/01-web-report.md` (W-01..W-16) | 353-page crawl, Chrome DevTools sessions (guest, admin, new user), Lighthouse mobile (a11y 100, SEO 100), production read-only checks, Playwright run (stopped: helper drift) |
| 2 Review and cleanup | `docs/audit/02-technical-review.md` (T-01..T-30), `docs/audit/CLEANUP.md` | two review passes with file:line evidence; cleanup executed for merged items only |
| 3 Docs and refactor | `docs/ROADMAP-V1.md`, `docs/BACKLOG.md`, `ROADMAP.md` phase P8, `AGENTS.md`, `CHANGELOG.md`, `docs/OPEN-QUESTIONS.md` | diffs on the branch |
| 4 Redis / MongoDB | `docs/adrs/0024-redis-deferred.md`, `docs/adrs/0025-mongodb-rejected.md` | Railway metrics 7 d, vukem driver inventory, cost model |
| 5 RFCs | `docs/rfcs/RFC-001-catalog-worker.md`, `RFC-002-trending.md`, `RFC-003-offers.md`, `docs/data/SOURCES.md` | licences researched 2026-11-02 with URLs; capacity model from production numbers |
| 8 Team | `PROMPT-WORKERS.md` (local, gitignored) | orchestrator reviews and merges; senior and intern in standby |

Code changes (2 commits, tests and lint green):

- `695872f5 fix(web): key the render cache by the stripped locale prefix` — verified in the rebuilt container: `/pt-br/`, `/es/`, `/` each serve their own language with `X-Render-Cache: HIT`.
- `b965c5d5 chore(build): keep tool state, worktrees and venvs out of the Docker context` — image rebuilt from the trimmed context.

## 2. What was removed

Per `docs/audit/CLEANUP.md`, after tag `backup/pre-cleanup-2026-11-02` and full-ref bundles in `backup-history-2026-11-02/`: 9 merged worktrees and their branches, 5 stale `node_modules`-only directories and 3 logs under `wt/` (now empty); 2 merged vukem branches; the dangling venv junction. Not removed (owner call or permission refused): 38 detached Temp worktrees (command in CLEANUP.md §E), 95 pre-rewrite local branches, 2 stashes, 3 uncommitted files, Docker images and volumes, `vercel-env/`.

## 3. Findings by severity

| Sev | Count | Headline items |
| --- | --- | --- |
| P0 | 0 | — |
| P1 | 3 | W-01/T-07 render cache locale (fixed, deploy pending); T-08 nightly recompute, similar and eval jobs cut at vukem's default 1-minute timeout (`app/jobs/recompute_scores.go:19`, `vukem/queue/queue.go:272,553`); W-16/T-22 e2e suite drifted (sign-up name field) and does not gate CI |
| P2 | 17 | LGPD erasure gaps (T-05), retention gaps (T-17), `item_status` index (T-16), quota meter single-replica (T-10), IGDB token (T-11), retry idempotency unproven (T-12), covers GC/cwebp limits (T-13), dependency ignores undated (T-21), staging DNS (T-23), pre-deploy mismatch (T-19), mobile item page order (W-02), 301 instead of 404 (W-03), no `og:image` (W-04), search client shell (W-05), Postgres volume headroom (T-18), dead trending cache key (T-09), Docker context (T-20, fixed) |
| P3 | 16 | copy, SEO titles, genre labels, a11y name mismatch, CSR private pages, alert codes, legacy routes and tables, board hygiene |

Verified good: security headers and CSP, auth stack, CSRF bypass list, SQL safety, pinned builds, CI without vacuous steps, backups and restore drills, kill switches everywhere.

## 4. Decisions taken

- ADR 0024: Redis deferred with three named triggers (≥ 2 replicas, measured Postgres contention, cross-process pub/sub) and a cost condition.
- ADR 0025: MongoDB rejected; raw payloads stay in Postgres `jsonb` with compression, partitioning and cold export.
- RFC-001: second Railway service `worker` on the same image (amends D-08; owner gate OQ-03); Postgres queue with app-level idempotency keys and jittered retries; provenance v2 with confidence and a review queue; 30-minute Telegram digest; LLM only as a normaliser under a US$3 ceiling; milestones 250 k / 500 k / 1 M with a cost projection that crosses US$20 at 1 M unless the owner raises the cap or applies the plan-B levers.
- RFC-002: home reads SDD 08 snapshots with a 3.5-day half-life, per-type z-score, entry/exit with 72-hour cooldown, diversity caps, quality floor, pin/ban, seed users excluded.
- RFC-003: `offers` model (item × country × store × kind), TMDB watch providers display-only with JustWatch attribution, store links from Wikidata IDs, Amazon by ISBN, affiliate tags only after enrolment, aggregate click counter without personal data.
- Team: orchestrator reviews and merges; senior (Sonnet) and intern (Haiku) in standby; ticket, report and review checklist formats in `PROMPT-WORKERS.md`.
- Audit deliverables written in English (owner choice), file names in English (`docs/audit/`, `docs/OPEN-QUESTIONS.md`, `docs/BACKLOG.md`).

## 5. Remaining risks

1. Scores in production are stale for most items until T-08 is fixed and a full recompute runs (owner starts it).
2. The locale fix is on a branch; until deployed, `/pt-br` and `/es` pages can still serve the wrong language.
3. The Playwright suite is broken by UI drift and off in CI; browser regressions are caught only by hand.
4. Account erasure leaves scoring and search rows linked to deleted users (LGPD exposure).
5. Postgres volume headroom (5.6 of 10 GB) is small for the worker's growth; resize is an owner action.
6. The US$20 cap does not reach 1 M items with full payload retention; the decision is explicit in RFC-001 §6.
7. Owner-only configuration gaps: pre-deploy command, staging DNS, admin IP allowlist, ingestion credentials.
8. Everything in `docs/data/SOURCES.md` marked "not verified" needs an owner read before a connector is enabled.

## 6. Next 10 steps

1. Owner: read `docs/OPEN-QUESTIONS.md` and answer OQ-01..OQ-14 (five minutes each; most are yes/no).
2. Orchestrator: run `scripts/ci/local.sh` on `audit/architect-2026-11-02`, merge to `main`, push (deploys the locale fix), verify `/pt-br/games/<slug>` in production (B-001).
3. Senior: B-002 and B-010 (job timeouts, resumable recompute); owner runs the one-off recompute.
4. Senior: B-012 (e2e helper) and re-enable the e2e job on release candidates.
5. Senior: B-011 (LGPD erasure) and B-024/B-025 (retention, index).
6. Owner: Railway pre-deploy command, staging CNAME, volume resize to 15 GB, `ADMIN_IP_ALLOWLIST`.
7. Orchestrator: execute the remaining cleanup the owner approves (CLEANUP.md B5-B8, C2, C3, D1-D4) and rewrite the task board from `docs/BACKLOG.md`.
8. Intern: B-046 (documents to English), B-041/B-042 copy items, `docs/data/SOURCES.md` verification rows.
9. Release `v0.39.0` (milestone V1-A) with the full local gate.
10. Start RFC-001 M0 (B-013): worker service in staging, digest on Telegram, 7 green days.

## 7. Verification run at the end of the audit

- `go test ./app/http/middleware/` ok; `golangci-lint run ./app/http/middleware/...` 0 issues.
- `docker compose --profile app build app-init` ok with the trimmed context; container healthy; locale check per language ok.
- `bunx markdownlint-cli2` on every new or edited Markdown file: 0 issues (pre-existing issues in the local `prompt.md` untouched).
- Not run: the full release gate (`scripts/ci/local.sh`) and `go test ./...`; they are the orchestrator's merge step (next step 2), not changed by this audit's two code commits beyond the middleware package.
