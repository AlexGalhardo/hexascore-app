# RFC-001: Catalogue worker 24/7 (100 k → 1 M items)

Status: proposed (architect audit, 2026-11-02). Owner gates: OQ-03 (second Railway service), OQ-04 (Telegram ops chat), OQ-05 (LLM provider and ceiling). Builds on SDD 06 (ingestion), ADR 0006 (licensing), ADR 0007 (ingestion in Go), ADR 0024/0025 (no Redis, no MongoDB), `docs/data/SOURCES.md`.

## 0. Executive summary

The ingestion pipeline the prompt asks for already exists in code (SDD 06: connectors, normaliser, covers, trailers, data-quality asserts, kill switches, chains, dead letter) but is switched off in production and runs inside the web process. This RFC does not redesign it. It (1) moves the ingest queue to a second Railway service built from the same image, (2) adds the five things the current design lacks for a 1 M-item target (job idempotency keys, retry jitter, a per-field confidence and review queue, a 30-minute Telegram digest, an LLM budget with kill switch), (3) fixes the throughput budgets of SDD 06 that would take 60 weeks at the Wikipedia rate, and (4) gives a milestone plan with cost projections showing the US$20/month cap holds until roughly 500 k items with the storage measures of ADR 0025, and needs an owner decision (raise the cap, or stop growth at the quality bar) beyond that.

## 1. Goals and non-goals

Goals: keep every existing item fresh (metadata, covers, scores, availability), translate titles into pt-BR and es when a source has them (never invent), validate trailers, add items from official or open sources in popularity order, reach 1 M items with per-field provenance and measured accuracy, inside US$20/month, with a 30-minute operational digest on Telegram.

Non-goals: scraping sources whose terms forbid it (JustWatch, HowLongToBeat, Goodreads, IMDb datasets: `docs/data/SOURCES.md`), critic scores (ADR 0006), changing the paid-API data posture, a metrics backend (D-35 stays), Redis (ADR 0024).

## 2. What exists today (verified 2026-11-02)

| Capability | Where | State |
| --- | --- | --- |
| Connectors Wikidata, Wikipedia (3 languages), Open Library, TMDB, IGDB, Steam, GOG, YouTube oEmbed | `internal/ingest/<source>/` | implemented, flags default off (`internal/ingest/run/flags.go`) |
| Chains (stage N queues stage N+1), resumable runs with cursor and 15-min heartbeat, `partial` runs | `internal/ingest/chains.go`, `run/continuation.go`, table `job_runs` | implemented |
| Per-host throttle, retry with `Retry-After`, breaker | `internal/ingest/httpx/` (own code, not vukem's `httpclient`) | implemented; breaker state is in-process only |
| Dead letter | vukem `vukem_failed_jobs` + `ingest_failures` (parked after 5 failed runs) | implemented |
| Provenance per field | `items.provenance jsonb` ({field: source}), `source_records` (payload, hash, `fetched_at`), licence via `data_sources` | implemented; no confidence level, no review queue except `entity_match_reviews` |
| Covers | `cwebp` child process, two variants, slug keys, GC with 30-day delay | implemented; cwebp without address-space limit, GC race (shortcuts) |
| Data quality | `dq_results`, `field_coverage`, `catalog_counts`, `dq.asserts` daily | implemented |
| Pressure gate | in-process HTTP in-flight counter pauses batches | implemented, only meaningful in the web process |
| Alerts | `alert=A7` log lines hourly, Telegram notifier for page alerts and errors, daily 08:00 digest | implemented; no throughput digest |
| Schedules | `routes/schedule.go:64-106`: daily incremental, one weekly full per source, expire/refresh jobs, covers, dq | implemented |
| Queue | `queue.Database` (`SKIP LOCKED`), `QUEUE_NAMES=default,ingest`, concurrency 3, `QUEUE_LIMITS=ingest=2` | implemented in the web process |

Gaps against the prompt: no separate worker, no idempotency key on jobs (vukem has none), deterministic backoff without jitter (vukem `queue.Exponential`), no confidence score or low-confidence review queue for fields, no 30-minute digest, no LLM path and no LLM budget, no image hash dedupe, trailer check is "embeddable", not "is the right trailer", throughput budgets in SDD 06 §7.2 are far too low for 1 M (Wikipedia 15 k/week).

## 3. Architecture

```text
Railway project hexascore
┌─────────────── service app (web) ───────────────┐   ┌──────────── service worker (new) ────────────┐
│ /app/app (serve)                                 │   │ /app/app queue:work                           │
│ HTTP + SSR + scheduler + queue "default"         │   │ QUEUE_NAMES=ingest  QUEUE_CONCURRENCY=3       │
│ QUEUE_NAMES=default  QUEUE_LIMITS=ingest=0        │   │ FLAG_INGEST_*=true  GOMEMLIMIT=400MiB          │
│ scheduler dispatches chain starts to "ingest"    │   │ worker pool per source host (httpx throttle)  │
│ writes pressure signal every 10 s (table)        │   │ reads pressure signal before each batch        │
└───────────────┬──────────────────────────────────┘   └───────────────┬──────────────────────────────┘
                │ Postgres: vukem_jobs (SKIP LOCKED), job_runs, source_records, items…, ingest_host_state
                └───────────────────────────────┬──────────────────────┘
                                                ▼
                       bucket hexascore-production (covers/, og/), Telegram Bot API, source APIs
```

Decisions:

1. **Same image, second service.** `/app/app queue:work` with `QUEUE_NAMES=ingest`; the web service drops `ingest` from its queues (`QUEUE_LIMITS=ingest=0` or `QUEUE_NAMES=default`). No code fork, one deploy pipeline (a push to `main` deploys both). Amends D-08; cost in §6. Plan B (if the owner keeps one service): `QUEUE_LIMITS=ingest=1` and heavy chains only 00:00–06:00 America/Sao_Paulo, which caps growth near 1,500 items/day.
2. **Scheduler stays in the web process** (`SCHEDULE_IN_SERVE=true` there, `false` on the worker). Chain starts are already queue jobs, so the worker only needs the queue. This avoids `OnOneServer()` and a shared lock store (ADR 0024).
3. **Pressure signal through Postgres.** The web process writes `{inflight, p95_ms, at}` to `ingest_host_state(key='web')` every 10 s (one `UPSERT`); the worker's batch loop reads it (1 query per batch) and waits while `inflight > 16` or `p95_ms > 300` (SDD 06 §7.5 thresholds). Stale signal (> 60 s) means the web is down: the worker pauses writes to `items` but keeps fetching into `source_records`.
4. **Breaker and throttle state persisted.** `ingest_host_state(key=host)` holds `rate`, `open_until`, `last_429_at`, `calls_today`; the worker loads it at boot so a restart does not reset a backoff (today the state dies with the process).
5. **Concurrency.** Worker pool per host as today (`httpx/hosts.go`), process-wide `QUEUE_CONCURRENCY=3`, `GOMEMLIMIT=400MiB`, `context` deadline per stage (10 min as now), graceful shutdown through vukem `StopTimeout` with `MaxFreeInterruptions` so a redeploy re-queues the stage without burning an attempt. Memory target: < 256 MB RSS steady, < 400 MB during a Wikidata batch (50 entities/request, trimmed before storing).
6. **Database protection.** Writes in batches of 200 QIDs per transaction (SDD 06), `statement_timeout=30s` on the worker connection, `DB_MAX_OPEN=4` on the worker, `synchronous_commit=off` for `source_records` writes only, and nightly windows for the `full` runs. The web service keeps its own pool.

## 4. Durable queue, idempotency, retries

- **Backend:** vukem `queue.Database` on Postgres (`FOR UPDATE SKIP LOCKED`, `vukem/queue/database.go:328`). Measured need: at 5,000 items/day with ~8 stages and batches of 50 to 200, the queue sees < 2,000 jobs/day; Redis is not justified (ADR 0024).
- **Idempotency key (new, app-level until vukem has unique jobs):** table `ingest_job_keys(key text primary key, job_id bigint, state text, created_at, done_at)`. Key = `sha256(job name + canonical params + stage + cursor)`. `Dispatch` first inserts the key with `ON CONFLICT DO NOTHING`; a conflict with `state in ('queued','running')` skips the dispatch, `done` within 24 h skips too, else re-dispatches. Every stage handler is already idempotent on data (content hashes in `source_records`, `UPSERT` on items), so duplicates are only wasted work; the key removes the waste and the double-count in `job_runs`. Framework item: VF-NEW "unique jobs" (`queue.Unique(key, ttl)`) so this table can go away.
- **Retry policy:** attempts 5 per stage, backoff `base·2^n` with full jitter (`rand(0, min(cap, base·2^n))`), base 30 s, cap 30 min, computed in the handler from `queue.Attempt(ctx)` and re-dispatched with `Delay` (vukem's `Exponential` has no jitter; VF-NEW "jitter option"). Permanent errors (`4xx` other than 429, schema validation failures) use `queue.ErrPermanent` as today.
- **Dead letter:** unchanged (`vukem_failed_jobs` for job-level failures, `ingest_failures` per target parked after 5 runs). New: `ingest:failures --list --source X` console command and the digest counts (§8). Parked targets are retried on the weekly full run only.
- **Scheduling:** unchanged cron table of SDD 06 §7.6, plus `ingest:digest` every 30 min and `ingest:refresh-existing` (§5.A) daily 01:30.

## 5. Work streams

### A. Keep existing items fresh (priority: items with a cover, then by `popularity_prior`)

Daily `ingest:refresh-existing` selects items by `coalesce(last_verified_at, '1970') asc`, cover first, popularity second, in slices sized to the day's budget (§6), and runs per item: re-fetch source records whose `checked_at` is older than their source TTL (Wikidata 30 d, Wikipedia 60 d, TMDB 150 d, IGDB 24 h per D-36, Steam/GOG 30 d, Open Library 90 d), re-normalise if any hash changed, re-check cover availability (HEAD on the bucket key, re-download if the source changed), re-run the trailer check, refresh `popularity_prior` from sitelinks and source popularity, and set `last_verified_at`. New columns: `items.last_verified_at timestamptz`, `items.quality_score smallint` (§7).

Translations: `item_translations` for pt-BR and es come only from Wikidata labels (`pt-br`, `pt`, `es`, `mul`), Wikipedia titles or TMDB/IGDB localized titles where licensing allows display; when none exists the row stores the original title with `source='original'` and `confidence=100` so the UI shows the original name and the API marks it as untranslated. No machine translation.

Trailers: today the check is "oEmbed answers 200". Add: (1) the video must come from Wikidata P1651, TMDB `videos` (type Trailer, official=true) or IGDB `videos`; (2) oEmbed `title` must contain the item's title or alt title (normalised) or the word "trailer"/"teaser"; (3) `author_name` is compared with the publisher/developer entity names or a known studio channel list; failures lower confidence instead of deleting (review queue). Optional YouTube Data API `videos.list` (1 unit) for `embeddable`, `privacyStatus=public` and `madeForKids`, capped at 5,000 units/day.

### B. Add new items

Order of preference per category is fixed by `docs/data/SOURCES.md`: official API → open dump → polite scraping of pages whose terms allow it (none today) → LLM extraction from already-fetched text (never from screenshots of sites that forbid it).

| Category | Discovery (candidates) | Enrichment | Not used |
| --- | --- | --- | --- |
| Films, series | Wikidata (P31 film/TV series classes, ≥ N sitelinks), TMDB discover for recent releases (IDs via `/find` only) | Wikidata, Wikipedia, TMDB (display-only), Commons posters | IMDb datasets, JustWatch, RT |
| Games | Wikidata (P31 video game, P400 platform), IGDB popularity lists (display-only), Steam app list for store IDs | Wikidata, Wikipedia, IGDB, Steam, GOG | HowLongToBeat |
| Books | Wikidata (P31 literary work, P50 author), Open Library dumps (works with ≥ 1 edition and a cover) | Wikidata, Wikipedia, Open Library editions | Google Books, ISBNdb, Goodreads, Skoob |

Candidate selection is popularity-first: sitelink count, TMDB popularity, IGDB rating count, Open Library edition count; the threshold moves down milestone by milestone (§10) so each milestone adds the most-wanted items first. A candidate becomes an item only when the normaliser has an English title, a release date or year, a type, and at least one external ID, and it passes the entity-resolution step below; otherwise it stays a `source_record` with `item_id NULL`.

### C. Deduplication and entity resolution

Canonical keys, in precedence order: `wikidata_qid` (unique on `items`), then `external_ids` (`tmdb`, `imdb`, `igdb`, `steam`, `gog`, `openlibrary_work`, `isbn13` on editions), then a normalised `(type, title_norm, year ± 1)` candidate match (`hs_norm`, trigram similarity ≥ 0.92). Rules: one external ID maps to at most one item (unique partial indexes on `external_ids->>'tmdb'` etc., to add); a candidate that matches two items by different keys goes to `entity_match_reviews` (exists) with kind `item_conflict`; a merge is only ever done by the existing admin merge (`catalog_redirects`, `catalog_merge_audit`), never by the worker. Source precedence per field (SDD 06 §7.3 extended): `community` > Wikidata > Wikipedia > Open Library > TMDB > IGDB > Steam > GOG > LLM, with the 180-day community lock (FR-06-4).

### D. Provenance and confidence per field

`items.provenance` grows from `{field: source}` to `{field: {source, fetched_at, licence, confidence}}` (jsonb, no migration of the column; the normaliser writes the new shape and readers accept both). Confidence (0-100) = source base (Wikidata 90, Wikipedia 85, Open Library 80, TMDB 85, IGDB 85, Steam 75, GOG 70, LLM 50) − 10 if two sources disagree − 20 if the value failed a dq assert, + 10 if two independent sources agree. Values below 60 do not reach `items` directly: they go to `field_reviews(item_id, field, value, source, confidence, reason, created_at, resolved_at, resolution)` (new table, admin page under `/admin/ingest/reviews` which already lists entity reviews). The API serialiser already filters by source allowlist; it additionally drops fields with confidence < 60.

### E. Images

Storage stays the Railway bucket via `drivers/s3`, keys by slug plus content-hash keys (exists). Additions: perceptual hash (dHash, 64-bit, pure Go, ~30 lines) stored in `cover_hashes(sha256, dhash, width, height, bytes, source, licence, created_at)` so the same cover from two sources is stored once and near-duplicates are detected; rights per cover recorded from the source row (`data_sources.licence`, Commons per-file licence), covers from sources without display rights are never written; `cwebp` keeps 600×600 q78 and 96×144 q70 (legacy parity) and runs under `contrib/engine` limits (`CPUTime 30 s, AddressSpace 512 MiB`, SDD 06 §7.5, not applied today). AVIF is deferred: no pure-Go encoder, `avifenc` would add ~20 MB to the image and CPU per cover; WebP is enough at these sizes. CDN: `COVERS_BASE_URL` already supports an external origin; the cost runbook's trigger (cover egress > US$5/month) decides when to put Cloudflare in front of the bucket.

### F. LLM as last resort

Only for normalising text the worker already fetched legally (for example extracting release year, publisher and genre from a Wikipedia infobox that Wikidata lacks, or classifying a synopsis language). Never as a source of record, never on screenshots of sites whose terms forbid reuse. Every call: structured output validated against a JSON schema, logged to `llm_usage(ts, job, model, tokens_in, tokens_out, cost_usd)`, confidence 50, so it always lands in `field_reviews`. Budget: `LLM_MONTHLY_BUDGET_USD` (default 3), `FLAG_INGEST_LLM` (default off), the worker refuses calls when `sum(cost_usd) this month ≥ budget` and alerts once. Provider and model are OQ-05.

## 6. Capacity and cost model

Inputs: Railway pricing read 2026-11-02 (US$20 per vCPU-month, US$10 per GB-month, US$0.15 per GB-month volume, US$0.05 per GB egress; billed per second), production baseline (app 0.11 GB avg memory, Postgres 0.26 GB, volume 5.55 GB used of 10 GB, month-to-date US$8.36 incl. the Pro seat), vukem queue throughput (hundreds of jobs/s on `SKIP LOCKED`; need < 1 job/s).

Per-item footprint (estimate from the current database size of 5.55 GB with 98,841 items, 2.1 M ratings, 2.05 M statuses, 677 k list items and no `source_records` yet; the catalogue tables are roughly 40 % of that): items + translations + entities + links + search columns ≈ 10 KB; trimmed source records ≈ 7 × 3 KB = 21 KB (ADR 0025 measures: keep only the latest payload, lz4 compression, partitioned history) ≈ 12 KB after compression; covers 21 KB in the bucket (2.1 GB per 100 k, NFR-06-7).

| Milestone | Items | Postgres data (GB) | Volume to provision | Bucket (GB) | Worker | Projected monthly (US$) |
| --- | --- | --- | --- | --- | --- | --- |
| today | 100 k | 5.6 | 10 | ~2 | none | 8.4 (measured MTD, Pro seat included) |
| M1 | 250 k | ~9 | 15 | ~5 | 0.1 vCPU avg, 0.25 GB | ≈ 13 (worker 2.5 + volume 0.8 + bucket 0.5 + Postgres memory growth 1) |
| M2 | 500 k | ~15 | 25 | ~10 | same | ≈ 16 |
| M3 | 1 M | ~28 | 40 | ~21 | same, night windows | ≈ 20-22 (**over the cap without a decision**) |

Plan B levers, in order: (1) stop storing payloads for sources with stable IDs (keep hash + fetched_at only) → −40 % Postgres; (2) move `source_record_history` older than 90 days to the bucket as NDJSON; (3) covers thumbnails only for items below a popularity threshold; (4) nightly-only windows (CPU to ~0.05 vCPU); (5) pause new-item discovery at M2 until the owner raises the cap (OQ-03). The `projected-spend.yml` alert at US$18 and the US$20 hard limit (D-27) are the automatic guard; the worker also stops discovery when `railway usage` MTD × days-in-month / elapsed > 18 (read via the Railway GraphQL usage query if the CLI offers it; otherwise the owner sets `INGEST_SPEND_GUARD=1` manually).

LLM ceiling: US$3/month (≈ 3 M input tokens at Haiku-class pricing; pricing to be verified when OQ-05 is answered), a hard stop in code, kill switch flag.

Throughput budget (corrects SDD 06 §7.2 for 1 M): Wikidata entity API 50 ids/request at 5/s → 2,000 req/day = 100 k entities/day; Wikipedia TextExtracts 20 titles/request, 2/s per language serially → 50 k synopses/day/language; Open Library dumps for the bootstrap (no API bulk); TMDB `/find` ≤ 20/s with a 20 k/day cap (display-only data, 6-month expiry means re-fetch cost grows with the catalogue: at 1 M items the expiry job alone needs ~5.5 k/day); IGDB 4/s, 200 slugs per query. Needed rate per §10 is 2,500 to 3,500 items/day, so network is not the constraint; quality review and storage are.

## 7. Quality

- Automatic validation per field (existing `dq.asserts` extended): type, year range, ISBN check digits, URL schemes, language codes, image dimensions, duplicate external IDs, titles that equal a QID or a slug.
- `items.quality_score` (0-100): cover present 25, synopsis in en 15, pt-BR and es titles 10, release date 10, ≥ 2 credits 10, ≥ 1 trailer verified 10, provenance confidence mean ≥ 80 20. Listings and trending (RFC-002) use a floor (≥ 60) for visibility in curated sections.
- Sampling: weekly, 200 random items per category exported to `qa/reports/ingest-sample-<date>.ndjson`; the intern checks 50 against the source pages and records accuracy in `docs/qa/ingest-accuracy.md`; target ≥ 97 % field accuracy at each milestone gate.
- Reprocessing is safe by construction (hashes, upserts, idempotency keys); `ingest:rollback --run <id>` exists (`internal/ingest/rollback`) and is extended to field reviews.

## 8. Observability and Telegram

- Structured logs (slog JSON) with `run_id`, `source`, `stage`, `batch`, `items`, `duration_ms`, `rate_limited`, `breaker` fields; redaction stays (`app/logging/redact.go`).
- Counters per source persisted hourly in `catalog_counts` and `job_runs` (exists): created, updated, failed, skipped, rate-limited, parked; new: `ingest_cost_daily(day, llm_usd, requests_by_host jsonb, bytes_down, covers_bytes)`.
- Health: worker `heartbeat` row in `ingest_host_state(key='worker')` every minute; `/up?checks=1` on the web adds `worker` = fresh heartbeat (< 3 min); Railway healthcheck on the worker service is not possible (no HTTP), so the web check is the probe.
- **Telegram digest every 30 min** (`ingest:digest`, scheduler in web, reads tables only): items created / updated / failed in the window, backlog (`queue.Database.Pending('ingest')`, `job_runs` queued/running), errors by source (top 5 `error_class`), breaker states and hosts rate-limited, covers written, LLM spend MTD and projected month spend vs the US$18 alert, worker heartbeat age, data quality deltas (`field_reviews` open). Sent to the silent ops chat (OQ-04) through the existing `app/alerts` notifier; the message is skipped when nothing changed and the backlog is zero (quiet hours stay quiet), and always sent at 08:00 and 20:00.
- **Immediate alerts** (page chat): worker heartbeat older than 10 min while jobs are queued, `vukem_failed_jobs` grows by > 20 in 30 min, breaker open for > 2 h on Wikidata or Wikipedia, LLM budget reached, disk projection crosses 80 % of the volume, `A7` conditions (exist).
- Token: `TELEGRAM_BOT_TOKEN` from the environment only (already), ops chat id `TELEGRAM_OPS_CHAT_ID` new, falls back to `TELEGRAM_ERROR_CHAT_ID`.

## 9. Legal and rights (per source, enforced in code)

`data_sources` gains `display_allowed`, `api_allowed`, `cache_ttl_days`, `attribution` columns (today the allowlist is `CATALOG_API_SOURCES` env); the normaliser refuses to write a field whose source has `display_allowed=false`; covers are written only when the source's cover rights allow display (Open Library covers: display, not API; Steam/GOG: display only after the owner's recorded read, SDD 06 §7.1). Attribution strings are rendered from `data_sources.attribution` on item pages and `/about` (TMDB logo and notice, IGDB line, Wikipedia CC BY-SA per synopsis, Commons per file, JustWatch for providers in RFC-003).

## 10. Milestones

| Milestone | Target date | Net new items/day | Candidate threshold | Exit criteria |
| --- | --- | --- | --- | --- |
| M0 "worker live" | +3 weeks | 0 (refresh only) | — | worker service deployed, flags on in staging 7 green days then production (P6.A.4.2), digest on Telegram, `refresh-existing` covers 100 % of items with a cover in 14 days, accuracy sample ≥ 97 %, spend ≤ US$12 |
| M1 250 k | +3 months | ~2,500 | films/series ≥ 15 sitelinks or TMDB popularity top 100 k; games ≥ 8 sitelinks or IGDB ≥ 20 ratings; books ≥ 10 sitelinks or OL ≥ 3 editions with a cover | quality_score ≥ 60 on ≥ 90 % of new items, covers on ≥ 80 % (books ≥ 60 %), 0 duplicate external IDs, spend ≤ US$14 |
| M2 500 k | +6 months | ~2,800 | thresholds halved | same bars, `source_records` compressed and partitioned (ADR 0025), spend ≤ US$17 |
| M3 1 M | +11 months | ~3,300 | thresholds halved again, long tail only with a cover | same bars; requires the OQ-03 cap decision (projection US$20-22) |

Each gate is a release (`vX.Y.0`) with the sampling report attached.

## 11. Backlog (tasks with acceptance criteria; estimates in senior-days; owner = Senior unless stated)

| ID | Task | Est. | Depends | Acceptance |
| --- | --- | --- | --- | --- |
| CW-01 | Worker service on Railway: Dockerfile unchanged, `queue:work` command verified, variables documented in `docs/runbooks/worker.md`, `/up?checks=1` reports `worker` | 2 | OQ-03 | staging: `railway logs --service worker` shows jobs processed; web `QUEUE_LIMITS=ingest=0` |
| CW-02 | `ingest_host_state` table: pressure signal writer (web) and reader (worker); breaker and throttle persistence | 2 | CW-01 | `TestIngestPressure_PausesAboveThresholds`, `TestBreakerStateSurvivesRestart` |
| CW-03 | Idempotency keys (`ingest_job_keys`) + jittered retries in `run/continuation.go`; VF-NEW entries for unique jobs and jitter | 2 | — | duplicate dispatch within 24 h is a no-op (test); backoff values differ between two attempts (test) |
| CW-04 | `refresh-existing` job with source TTLs, `last_verified_at`, cover HEAD check, trailer verification v2 | 3 | CW-02 | 1,000 items refreshed in staging; trailer with a wrong title lands in `field_reviews` |
| CW-05 | Provenance v2 (`{source, fetched_at, licence, confidence}`), `field_reviews` table + admin page, serialiser confidence filter | 3 | — | API excludes a field with confidence 50 (test); admin resolves a review and the item updates |
| CW-06 | Entity resolution rules: unique partial indexes on external IDs, `(type,title_norm,year)` candidate match, `item_conflict` reviews | 2 | CW-05 | duplicate TMDB id insert fails; near-duplicate goes to review (tests with fixtures) |
| CW-07 | Cover dHash + `cover_hashes`, rights check from `data_sources`, `contrib/engine` limits for cwebp | 2 | — | same image from two sources stored once; cwebp over 512 MiB is killed (test with `FAKE_CWEBP`) |
| CW-08 | `ingest:digest` every 30 min + immediate alerts, `TELEGRAM_OPS_CHAT_ID`, `ingest_cost_daily` | 2 | CW-01 | message fixture matches; quiet window skipped; `alerts:test --digest` sends one |
| CW-09 | Discovery v2: popularity thresholds per milestone as config, Open Library dump loader (streamed, no API bulk), TMDB discover for recent releases | 4 | CW-06 | staging adds 5,000 items in one night within rate budgets; 0 duplicates |
| CW-10 | LLM normaliser behind `FLAG_INGEST_LLM` with schema validation, `llm_usage`, monthly ceiling | 2 | OQ-05, CW-05 | call refused at the ceiling (test); output lands in `field_reviews` |
| CW-11 | ADR 0025 storage measures: lz4 on `source_records.payload`, partition history by month, cold export job | 2 | — | `pg_total_relation_size` drops ≥ 30 % on a staging copy |
| CW-12 (Intern) | Accuracy sampling procedure + `docs/qa/ingest-accuracy.md`, weekly | 1 | CW-04 | first report with ≥ 50 checked items |
| CW-13 (Intern) | `docs/data/SOURCES.md` owner reads for the "not verified" rows | 0.5 | owner | rows updated with dated quotes |
| CW-14 | SDD 06 update (budgets §7.2, pressure signal, digest, provenance v2) and runbook `ingestion-stuck.md` | 1 | all | docs match code; CHANGELOG entry |

Risk and rollback per task: every change is behind the existing kill switches (`ingest_enabled`, per source) plus the new `FLAG_INGEST_LLM`; the worker service can be deleted and `QUEUE_LIMITS=ingest=2` restored on the web service in one deploy; data writes are reversible through `ingest:rollback --run`.
