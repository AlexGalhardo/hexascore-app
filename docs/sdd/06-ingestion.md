# SDD 06: Ingestion and data operations

Status: reviewed (2026-10-28 consistency pass). Decisions: D-03, D-08, D-09, D-10, D-11, A-05. ADRs: 0003, 0006, 0007, 0012, 0014. Related SDDs: 01, 02, 04, 05, 11, 12.

## 1. Objective

Keep the catalogue (legacy production: 99,960 items) complete, fresh and legally clean without a laptop. Go connectors fetch Wikidata (identity spine) and enrichment sources into `source_records`; a pure normaliser rebuilds items, editions, entities and translations with provenance; follow-up jobs produce covers and trailers; asserts fail loudly. Users get complete pages, API buyers only licence-compatible fields (D-03), scoring gets correct release dates (SDD 02). `legacy/webscrappings/` and its 77 tests are the specification (D-11).

## 2. Scope

In (v1): Wikidata, Wikipedia, Open Library (gated by A-05), TMDB, IGDB, Steam, GOG, YouTube oEmbed; normalisation; entities (D-01); editions (D-10); covers; trailers; jobs in one process (D-08, ADR 0007); asserts; backfills; rollback; legacy import and parity.

Out: IMDb datasets (ADR 0006); critic scores (D-09); HTML scraping; IGDB webhooks/dumps, TMDB changes/exports (issue 10: revisit when the weekly TMDB run exceeds 2 h or the IGDB sync exceeds 1 h); Railway config-as-code (`.railway/railway.ts`, owned by SDD 11); generating own synopses; a separate worker service (first step when the D-08 ceiling rises).

## 3. Functional requirements

- FR-06-1: Connectors store a trimmed, validated payload keyed `(source, external_id)` with a canonical SHA-256; an unchanged hash only updates `checked_at`. Verify: ports of `test_upsert_is_noop_when_payload_unchanged`, `test_canonical_hash_ignores_key_order`.
- FR-06-2: The normaliser is a pure function of all records of one QID, independent of order. Verify: property test shuffling records.
- FR-06-3: Wikidata wins; other sources fill empty fields only, with field provenance; TMDB/IGDB/Open Library text never reaches `item_translations`. Verify: ports of `test_tmdb_fills_only_what_wikidata_lacks`, `test_igdb_cover_wins_over_steam`.
- FR-06-4: Fields with provenance `community` (approved SDD 04 edits) are skipped by every connector for 180 days after the applying audit row; a diverging source value opens a `system` suggestion (SDD 04 §7.3, open issue 04-3) instead of writing; after 180 days a divergence still opens a suggestion rather than overwriting. Verify: integration test at day 179 and 181 with a fake clock.
- FR-06-5: No TMDB value is older than 180 days by `checked_at`. Verify: fake-clock test; DQ-I-06.
- FR-06-6: Jobs are idempotent and resume from `job_runs.cursor`. Verify: cancel at a random batch, restart, compare with an uninterrupted run.
- FR-06-7: A disabled source makes zero requests; runs end `skipped`. Verify: `httptest` hit counter.
- FR-06-8: Admins see runs, failures and asserts and trigger runs. Verify: Playwright on `/admin/ingest`.
- FR-06-9: Asserts (§7.7) run after each normalise and daily; `error` results alert. Verify: one faulty fixture per assert.
- FR-06-10: Zero semantic diffs between Python and Go on the 200-QID sample. Verify: `TestParityLegacy`.
- FR-06-11: The Wikidata connector fills `item_alt_titles` (SDD 01 §5.2) from aliases (`kind=alias`), romanised labels (`transliteration`, `lang` with `-Latn`), original-language labels (`original`) and former titles (`former`), source `wikidata`, ≤ 20 per item (open issue 07-6); and `item_links` `based_on` from P144 when both QIDs are catalogue items (open issue 08-1). Verify: parity fixture with a manga-based anime and a novel-based film.

## 4. Non-functional requirements

- NFR-06-1 Priority: during the weekly TMDB run, item-page p95 rises ≤ 20 %. Verify: k6 on staging.
- NFR-06-2 CPU: ≤ 2 ingestion jobs at once; one WebP encode at a time; 100 ms between batches; pressure gate (§7.5).
- NFR-06-3 Memory: ≤ 256 MB; rows streamed in batches of 200; selections read only needed JSON keys (legacy `covers.py` `targets`).
- NFR-06-4 Freshness: Wikidata edits live within 26 h; TMDB ≤ 7 days in steady state, cap 180.
- NFR-06-5 Politeness: §7.2 rates never exceeded. Verify: fake-clock port of `test_rate_limit_spaces_requests_per_host`.
- NFR-06-6 Security: outbound HTTP only through vukem `httpclient.Safe` (<https://vukem.dev/docs/security/overview> "A01"), since cover URLs come from third-party payloads; host allowlist per connector; JSON ≤ 20 MB, images ≤ 15 MB; credentials are `config.Secret`.
- NFR-06-7 Cost: covers ≈ 2.1 GB per 100k items (legacy measured 20.5 GB per 1M).

## 5. Data model

Catalogue tables: SDD 01 §5.2, plus `items.release_precision` (`day|month|year|unknown`) and `items.release_status` (`announced|released|cancelled`) decided in `open-issues.md` (02-1) and filled here. Ingestion tables (via `s.Raw`, <https://vukem.dev/docs/database/migrations> "Migrations"):

```sql
create table source_records (            -- legacy catalog.ts shape + checked_at, last_run_id
  source text not null,                   -- wikidata, wikidata:entity, wikipedia:en|pt|es, openlibrary, tmdb, igdb, steam, gog, youtube, synopsis
  external_id text not null,              -- always a Wikidata QID (legacy common/db.py Record)
  payload jsonb not null,
  hash text not null check (hash ~ '^[0-9a-f]{64}$'),
  fetched_at timestamptz not null default now(),   -- last payload change
  checked_at timestamptz not null default now(),   -- last successful fetch (TMDB expiry clock)
  normalized_at timestamptz,
  item_id uuid references items(id) on delete set null,
  last_run_id bigint,
  primary key (source, external_id)
);
create index source_records_pending_idx on source_records (source)
  where normalized_at is null or normalized_at < fetched_at;
create index source_records_item_idx on source_records (item_id);   -- missing in legacy

create table source_record_history (     -- previous version on each hash change, kept 30 days
  source text not null, external_id text not null,
  hash text not null, payload jsonb not null, fetched_at timestamptz not null,
  superseded_at timestamptz not null default now(), superseded_by_run bigint not null,
  primary key (source, external_id, superseded_at)
);

create table job_runs (
  id bigint generated always as identity primary key,
  job text not null,                      -- fetch.wikidata, normalize, covers, dq.asserts, ...
  source text,
  mode text not null check (mode in ('bootstrap','incremental','full','followup','manual','backfill','rollback')),
  status text not null check (status in ('queued','running','succeeded','partial','failed','skipped','cancelled')),
  params jsonb not null default '{}',     -- since, limit, types, qids, allow_count_drop, reason
  cursor jsonb,                           -- {"after":"Q1234"}
  attempt int not null default 0,
  seen int not null default 0, changed int not null default 0,
  failed int not null default 0, skipped int not null default 0,
  rate_limited int not null default 0, total_estimate int,
  triggered_by text not null,             -- schedule | admin:<user_id> | chain:<run_id>
  error text,
  started_at timestamptz, heartbeat_at timestamptz, finished_at timestamptz,
  created_at timestamptz not null default now()
);
create unique index job_runs_one_active on job_runs (job, coalesce(source,''))
  where status in ('queued','running');

create table ingest_failures (            -- per-target dead letter
  source text not null, external_id text not null,
  last_run_id bigint not null,
  error_class text not null,              -- http_4xx, http_5xx, timeout, invalid_payload, conflict
  error text not null, attempts int not null default 1,
  parked boolean not null default false,  -- after 5 consecutive failed runs
  first_seen timestamptz not null default now(), last_seen timestamptz not null default now(),
  primary key (source, external_id)
);

create table entity_match_reviews (
  id bigint generated always as identity primary key,
  item_id uuid not null references items(id) on delete cascade,
  role text not null, source text not null,
  raw_name text not null, raw_year smallint, raw_external_id text,
  candidates uuid[] not null,
  status text not null default 'open' check (status in ('open','linked','created','dismissed')),
  resolved_by uuid, resolved_at timestamptz,
  unique (item_id, role, source, raw_name)
);

create table dq_results (
  run_id bigint not null references job_runs(id),
  code text not null,                     -- DQ-I-xx and SDD 01 DQ-xx
  severity text not null check (severity in ('error','warn')),
  passed boolean not null,
  observed numeric, threshold numeric,
  sample jsonb not null default '[]',     -- <= 20 offending ids
  checked_at timestamptz not null default now(),
  primary key (run_id, code)
);
create table catalog_counts (day date, type text, items int not null, primary key (day, type));
create table field_coverage (
  run_id bigint not null references job_runs(id),
  type text not null, field text not null, locale text not null default '',
  sampled int not null, missing int not null,
  primary key (run_id, type, field, locale)
);
```

Invariants: history is written in the same statement as the update (CTE); `job_runs_one_active` blocks overlap regardless of scheduler locks. Retention: history 30 days, `job_runs` 180, `dq_results`/`field_coverage` 365. Legacy `source_records` copies 1:1 (`checked_at = fetched_at`).

## 6. Contracts

Jobs (<https://vukem.dev/docs/digging-deeper/queue> "Queue"): payload `{run_id}`, parameters in `job_runs.params`; queue `ingest`, `queue.Timeout(10*time.Minute)`, `queue.Retries(3)`; `QUEUE_NAMES=default,ingest` (priority with the database driver inside `serve` [NOT VERIFIED], issue 5). Chunked continuation: a handler processes ≤ 500 targets or 5 min, saves `cursor` and dispatches its successor with `q.DispatchTx` in that transaction. Permanent errors (TMDB 401, bad config) wrap `queue.ErrPermanent`; `queue.AttemptFrom(ctx).Last()` closes the run row ("Progress is not a queue feature").

Events (<https://vukem.dev/docs/digging-deeper/events> "Events"): `catalog.ItemsChanged{IDs}` per committed normalise batch (SDD 07 search, SDD 09 cache); SDD 02 `RecomputeItemScore{ItemID}` via `DispatchTx` when a rated item's release date, precision or status changes.

Admin: read-only panel resources for `job_runs`, `ingest_failures`, `dq_results`; actions are app routes, custom panel actions being [NOT IN VUKEM DOCS] (VF-008).

| Method | Path | Ability | Input | Result / RFC 9457 |
| --- | --- | --- | --- | --- |
| GET | `/admin/ingest` | `viewIngestion` | — | last run per job, failures, failing asserts, oldest TMDB age, coverage trend |
| POST | `/admin/ingest/runs` | `runIngestion` + CSRF | `{job, source?, mode, since?, limit?, qids?[≤500], allow_count_drop?, reason?}` | 303; 409 `run-active`; 423 `source-disabled`; 422 |
| POST | `/admin/ingest/runs/{id}/cancel` | `runIngestion` | — | 303 |
| POST | `/admin/ingest/runs/{id}/rollback` | `rollbackIngestion` | `{reason}` | 303; 409 `history-expired` |
| GET/POST | `/admin/ingest/reviews[/{id}]` | `runIngestion` | `{action: link\|create\|dismiss, entity_id?}` | match queue |

Kill switches (<https://vukem.dev/docs/digging-deeper/feature-flags> "Feature flags"): `ingest_enabled`, `ingest_<source>` (plus `covers`, `youtube`), code default `false` (fail closed), on in production via `FLAG_INGEST_<SOURCE>=true`; a stored `false` at `/admin/_flags` wins within the 30 s cache TTL.

## 7. Behaviour and algorithms

### 7.1 Sources and licences

| Source | Licence / terms | Contributes | Paid API | Attribution |
| --- | --- | --- | --- | --- |
| Wikidata | CC0 | QID, titles, dates, credits, external IDs, trailers P1651, poster P3383, sitelinks | yes | — |
| Wikipedia en/pt/es | CC BY-SA 4.0 | synopses; entity descriptions (≥ 3 items) | yes, licence + URL | "From Wikipedia, CC BY-SA 4.0" + link per synopsis |
| Commons (P3383) | per file | film poster fallback | per file | author, licence, URL |
| Open Library | no rights asserted by the Internet Archive (<https://openlibrary.org/developers/licensing>); legal read cleared by the owner (Q-A16, 2026-11-01); connector behind `ingest_openlibrary`, default off | pages, editions, covers | identifiers and titles only (`limitedSources`) until a licence note widens it | — |
| TMDB | no commercial use, ≤ 6-month cache | poster, fallback facts, trailers | never | logo + not-endorsed notice on item pages using TMDB fields and `/about` |
| IGDB | interim rule D-36: site-only, refreshed every 24 h, never cached > 7 days, until the owner records `docs/legal/igdb.md` | cover, fallback facts, trailers | never | "Game data by IGDB.com" in game page footers and `/about` |
| Steam `appdetails` | [NOT VERIFIED] | header image, OS platforms | store ID (Wikidata) | — |
| GOG catalog | [NOT VERIFIED] | store link, cover, OS platforms | store ID (Wikidata) | — |
| YouTube oEmbed | YouTube ToS | embeddability verdict | IDs (Wikidata) | player |

Steam/GOG images are site-only, store IDs are used for links only, and images stay out of the app until both terms are read (roadmap task, open issue 06-8); affiliate URLs come from store IDs, flag `affiliates` (SDD 01).

IGDB interim rule (D-36): the owner reads the Twitch Developer Services Agreement and records the decision in `docs/legal/igdb.md` before the ingestion milestone (owner task). Until then IGDB fields are audience `site` only, `igdb.refresh` re-fetches every IGDB record whose `checked_at` is older than 24 h, and any IGDB record or cover older than 7 days is deleted and the item re-normalised (DQ-I-13).

A-05 blocking task (before `ingest_openlibrary` is enabled): read `https://openlibrary.org/developers/licensing` and `https://openlibrary.org/dev/docs/api/covers`, record quotes and date in `docs/legal/open-library.md`. Rule: data CC0/public domain → ingest, licence `CC0-1.0`, API-eligible; attribution-only → ingest, API-eligible with `data_sources.attribution`; silent, non-commercial or no-redistribution → do not ingest, purge, drop from `CATALOG_API_SOURCES`. Covers are API-eligible only with an explicit grant. Until recorded, imported records serve the API identifiers and titles only (`open-issues.md` 01-2).

### 7.2 Rate limits, retries, breaker

Per-host throttle shared across goroutines (slot reservation, legacy `common/http.py` `_throttle`) is app code: vukem has no outbound limiter or breaker [NOT IN VUKEM DOCS] (issue 11).

| Host | Rate | In flight | Batch | Budget |
| --- | --- | --- | --- | --- |
| query.wikidata.org | 1/s | 1 | 1 query/type | ≤ 50/day |
| [www.wikidata.org](https://www.wikidata.org) | 5/s | 2 | 50 ids | ≤ 1,000/day; full ≈ 8,000 |
| {en,pt,es}.wikipedia.org | 5/s each | 4 | 20 titles | ≈ 15,000/week |
| commons / upload.wikimedia.org | 5/s / 2/s | 2 | 1 | ≤ 2,000/day |
| openlibrary.org / covers.openlibrary.org | 2.5/s / 10/s | 2 / 4 | 1 | ≈ 50,000/week |
| api.themoviedb.org | 20/s ("~40", respect 429) | 8 | 1 | ≈ 75,000/week |
| api.igdb.com | 3.3/s, cap 4/s | 4 (limit 8) | 200 slugs | ≈ 150/week |
| store.steampowered.com | 1 per 1.6 s | 1 | 1 | ≈ 15,000/week |
| catalog.gog.com | 2/s | 1 | 1 | ≈ 3,000/week |
| [www.youtube.com](https://www.youtube.com) oEmbed | 10/s | 4 | 1 | ≤ 20,000/week |
| image.tmdb.org / images.igdb.com | 20/s | 4 | 1 | covers |

User-Agent `HexaScoreBot/1.0 (+$INGEST_CONTACT_URL)`; timeout 60 s; retry transport errors and 429/500/502/503/504, 5 attempts, `2^n` s ≤ 60 s, `Retry-After` ≤ 300 s applied to the whole host; other 4xx final (legacy `test_http.py`). After 20 rate-limit hits a run halves the host rate. Breaker per host: opens after 10 consecutive failures or ≥ 50 % of the last 50; open 5 min doubling to 1 h; one half-open probe; ends the run `partial` with its cursor.

### 7.3 Connectors and normaliser

```go
type Connector interface {
    Source() string
    Targets(ctx context.Context, db database.Executor, p RunParams, after string) iter.Seq2[Target, error]
    Fetch(ctx context.Context, c *Client, batch []Target) ([]Record, []TargetError)
}
```

Ported from `legacy/webscrappings/connectors/*.py`. Every connector checks field provenance before writing and skips `community` fields per FR-06-4. Wikidata selects by `P31` classes and sitelinks (bootstrap 35k movies, 15k series, 25k games, 25k books; incremental `dateModified ≥ since`, ≥ 15 sitelinks) and stores a trimmed entity (aliases and P144 are kept in the trim for FR-06-11); Wikipedia rejects disambiguation and mismatched `wikibase_item`; Open Library prefers edition over work, drops cover `-1`; TMDB resolves IDs only via `/find/{imdb_id}`, 401 stops the run; IGDB matches regex-filtered slugs; GOG matches slug, title, then slug prefix. Trimming is load-bearing: a new property changes every hash and needs a `full` Wikidata run (≈ 1 h).

`BuildItem(qid, records)` ports `normalize.py`: skip without Wikidata record or English title; type fixed; slug from the English title (ASCII, ≤ 80), collision → `-<year>` → `-<qid>`, never changed; release date = min P577 (series P580 first, fallback P571), padded as legacy, plus `release_precision` from Wikidata precision 11/10/9 and `release_status = announced` when the date is future or TMDB/IGDB say unreleased [NOT VERIFIED field names]; titles en ← `en`/`mul`/enwiki, pt-BR ← `pt-br`/`pt`/`mul`, es ← `es`/`mul`; synopsis only from Wikipedia or `synopsis`; ≤ 2 trailers from `youtube.available`. One transaction per 200 QIDs; provenance `{field: source}`, licence from `data_sources`, freshness from `checked_at`.

### 7.4 Entities and editions

Role properties (new ones [NOT VERIFIED] until fixtures pin them, issue 9): director P57, screenwriter P58, cast P161 (first 15), creator P170, author P50, developer P178, publisher P123, production company P272, network P449, platform P400, genre P136, game mode P404, series P179, franchise P8345, country P495, language P364 (books P407). Kind per SDD 01's role table; referenced QIDs fetched as `wikidata:entity` records, 50 per request.

Matching per credit: (1) QID; (2) source external ID (`tmdb_person`, `tmdb_company`, `igdb_company`, `openlibrary_author`); (3) normalised name (NFKD, ASCII fold, lowercase, punctuation to space) + kind + year ±1 or unknown. One candidate → link; ≥ 2, or QID stored under another kind → `entity_match_reviews`; none → native entity (DQ-09). Fallback credits keep their `item_entities.source`, so the API filters them.

Editions (ADR 0012): strip to `0-9X`; 10 chars → mod-11 check, convert to `978` + 9 digits + new check digit; 13 → `isbn13_is_valid`; else dropped and counted. Sources: P212/P957 on the work, Wikidata editions (P747) with P407/P123/P577, Open Library editions after A-05; ≤ 50 per work, pt/en/es first. An ISBN on another work is never moved (`conflict` + review). Language from MARC (`por`, `eng`, `spa`), else `und` (DQ-10); `market_country = BR` from publish country `bl` [NOT VERIFIED]; primary edition = earliest with a cover in the original language.

### 7.5 Covers, trailers, pressure gate

Covers (port of `covers.py`): Open Library (books) → TMDB → IGDB → Steam → GOG → Commons P3383, never P18; better sources replace lesser ones. Download via `httpclient.Safe` (≤ 15 MB), sniff, 25 MP check, EXIF orientation, fit 600×600 q78 and 96×144 q70 (the only two variants, legacy parity: no 300 px or 2× thumb unless the LCP > 2.0 s or mobile scroll budgets fail, open issues 09-9, 10-3), never upscale, strip metadata, WebP, `Put` on `drivers/s3` (Railway bucket or S3-compatible; no bucket versioning, ADR 0014 as amended: immutable hashed keys plus 30-day deferred delete, open issue 11-1) at `covers/<item_or_edition_id>-<sha256(source)[:16]>.webp` and `…-t.webp`, cache 1 year; same hash → skip. `contrib/imaging` encodes only PNG/JPEG (vukem brainstorming FUN-08), so `cwebp` runs via `contrib/engine` with `Limits{CPUTime: 30 s, AddressSpace: 512 MiB}` (local `vukem/docs/subsystems.md` "External binaries" [SITE PAGE MISSING]), pinned in the image (issue 3). `covers.gc` deletes replaced objects after 30 days. Target: books ≥ 60 % covered (from 34 %).

Trailers (port of `trailers.py`): P1651 (preferred first), then TMDB/IGDB; oEmbed 200 = available, 400/401/403/404 = unavailable, 429/5xx = no verdict; stop at 2.

Pressure gate: app middleware counts in-flight requests; batches wait while > 16 or last-minute item-page p95 > 300 ms. Heavy runs start 00:00–06:00 America/Sao_Paulo.

### 7.6 Jobs and schedules

`SCHEDULE_TIMEZONE=America/Sao_Paulo` (as SDD 02); tasks `WithoutOverlapping()` (<https://vukem.dev/docs/digging-deeper/scheduler> "Scheduler"); a successful stage dispatches the next.

| Job | When | Chain / notes |
| --- | --- | --- |
| `bootstrap` | manual | Wikidata top-N → enrichers missing-only → normalise → covers → trailers → asserts |
| `fetch.wikidata` incremental | daily 00:30 | `since 36h` → enrichers for changed QIDs → normalise → covers → trailers missing-only → asserts |
| weekly full per source | 01:00: Sun wikidata + entities, Mon wikipedia, Tue openlibrary, Wed steam, Thu gog, Fri tmdb, Sat igdb | → normalise → covers; Sun full trailer re-check |
| `tmdb.expire` | daily 03:00 | `checked_at` > 150 d → refresh; > 180 d → delete record, re-normalise, drop TMDB cover |
| `igdb.refresh` | every 6 h | D-36 interim: refresh IGDB records with `checked_at` > 24 h; delete records/covers > 7 days, re-normalise |
| `ingest.retry_failures` | daily 03:30 | non-parked failures |
| score recompute | daily 04:00 | owned by SDD 02 |
| `dq.asserts` | after normalise; daily 05:00 | §7.7 |
| `dq.frontend_fields` | daily 05:30 | §7.7 |
| `covers.gc`, prunes | daily 02:30 | retention §5 |

Dead letter: vukem `failed_jobs` (`queue:failed`, `queue:retry`) for jobs; `ingest_failures` per target, parked after 5 failed runs. A run with `heartbeat_at` older than 15 min becomes `failed` and resumable.

### 7.7 Asserts and missing-field detector

SDD 01 DQ-01..12 run as defined there. Additions:

| Code | Assert | Threshold | Severity |
| --- | --- | --- | --- |
| DQ-I-01 | items per type vs last `catalog_counts` | drop ≤ 1 % unless `allow_count_drop` + reason | error |
| DQ-I-02 | `en` title (= DQ-01) | 0 | error |
| DQ-I-03 | ISBN-13 checksum, unique (= DQ-02) | 0 | error |
| DQ-I-04 | `cover_key` ⇒ object exists (`Stat`; 1,000/day, all weekly) | 0 | error |
| DQ-I-05 | synopsis ⇒ licence + source URL (= DQ-06) | 0 | error |
| DQ-I-06 | TMDB `checked_at` age | > 170 d warn, > 180 d error | error |
| DQ-I-07 | orphan entities (= DQ-03) | listed | warn |
| DQ-I-08 | slug regex, uniqueness (= DQ-04) | 0 | error |
| DQ-I-09 | release date 1870-01-01..today + 10 y (books from 1400) | 0 | error |
| DQ-I-10 | duplicates: title + type + year; ISBN twins; entity twins | report | warn |
| DQ-I-11 | match reviews open > 7 days | ≤ 500 | warn |
| DQ-I-12 | API serialiser on 100 items emits no `tmdb`/`igdb`/`steam`/`gog` | 0 | error |
| DQ-I-13 | IGDB records with `checked_at` older than 7 days (D-36) | 0 | error |

Missing-field detector: daily 500 items per type (top 200 + 300 random) rendered through the item page read-model builder (SDD 09). Required: game (title, date, cover, genre, developer, platform, en synopsis), movie (+ runtime, director), series (+ seasons, creator or network), book (+ author, edition). Nulls per type/field/locale feed `field_coverage` (admin trend) and ≤ 200 deduplicated `system` "missing field" entries/day in the SDD 04 queue.

### 7.8 Backfills and rollback

New field over 100k items: nullable expand migration → `backfill.<name>`, keyset over `items.id`, 500 per transaction, `WHERE col IS NULL`, 100 ms pause, progress `seen/total_estimate` → assert `warn` until 100 %, then `error` → `NOT NULL` next release (ADR 0014).

Bad run: restore history rows with `superseded_by_run = <id>`, clear `normalized_at`, delete records the run inserted; the normaliser rebuilds items; items the run created are deleted unless user data references them (then flagged); covers re-run (old objects live 30 days). Logged as a `rollback` run plus audit entry. Beyond 30 days: restore the latest dump into a scratch database (ADR 0014) and copy catalogue tables back in one transaction, user tables untouched (runbook: SDD 11).

### 7.9 Legacy import and parity

Command `hexascore:import-legacy` (registration [NOT VERIFIED]) on the legacy dump restored into schema `legacy_import`: seed `data_sources` → copy `source_records` → `items` (ids kept, `popularity` → `popularity_prior`) → `item_translations` (`synopsis_license` → `synopsis_licence`, source from URL host) → Go normaliser over all records (entities, editions, provenance; `meta` strings become native entities only without records) → `external_ids` keep link IDs, `isbn13` moves to editions → Blob covers copied under §7.5 keys → `external_scores` dropped → asserts. Idempotent, resumable.

Parity: `testdata/parity/qids.txt`, 200 QIDs (50 per type; top/mid/tail; non-Latin titles, year-precision dates, P580 series, ISBN-10-only book, Steam-only game) with frozen `source_records`. Python `build_item` and Go `BuildItem` emit canonical JSON; exact diff on titles, release date, slug, genres, creators, platforms, countries, runtime, pages, seasons, episodes, trailers, external IDs; synopsis after trim, whitespace collapse, NFC. Allowed diffs: scores, precision/status, entity ids, `creators` split by role.

## 8. Edge cases

- Wikidata entity deleted or redirected: item kept, flagged, merged via SDD 01 after review.
- QID changes class: type stays; `conflict` failure.
- TMDB token rejected: run `failed`, alert; other sources continue.
- Crash mid-batch: rollback; chunk reruns from cursor.
- Community-edited field diverges: source skipped, `system` suggestion.
- ISBN on two works: second link rejected, review.
- Rated item's date moves to the future: `announced`, recompute dispatched.

## 9. Acceptance criteria

- [ ] `go test ./internal/ingest/...` → green; one Go test per legacy test except score, Blob, keep-alive and revalidate ones; fixtures copied from `legacy/webscrappings/tests/fixtures/`.
- [ ] `TEST_DATABASE_URL=… go test ./tests/ingest -run TestResumeAfterCancel -count=20` → tables equal an uninterrupted run.
- [ ] `go test ./tests/ingest -tags parity -run TestParityLegacy` → `0 semantic diffs / 200`.
- [ ] `psql -c "select count(*) from source_records where source='tmdb' and checked_at < now()-interval '180 days'"` → `0`.
- [ ] `ingest_tmdb` off, run triggered → `skipped`, zero stub hits.
- [ ] Latest `dq.asserts`: `select code from dq_results where run_id=$1 and not passed and severity='error'` → no rows.
- [ ] Book covers ≥ 60 % on staging after the edition backfill (if A-05 passes).
- [ ] k6 item-page p95 during the Friday TMDB run ≤ 1.2 × idle.

## 10. Test strategy

- Unit: fixture-based parsers, normaliser, ISBN, name normalisation, throttle/breaker on a fake clock.
- Integration (Postgres, vukem `pgtest`): history, continuation, rollback, matching, expiry, kill switches, `httptest` stubs.
- Parity (§7.9) in CI for `internal/ingest`; security `TestSecurity_A01_IngestSSRF`, image bombs; k6 during ingestion; asserts nightly and in the monthly restore drill.
- Observability (SDD 12): `ingest.run.duration{job,source,status}`, `ingest.records{source,outcome}`, `ingest.http.rate_limited{host}`, `ingest.breaker.open{host}`, `ingest.dq.failed{code}`, `ingest.tmdb.max_age_days`, plus built-in `messaging.process.duration`, `http.client.request.duration` (<https://vukem.dev/docs/digging-deeper/queue> "Metrics"); app metrics API [NOT IN VUKEM DOCS] (issue 7). Alerts: `failed` run; no successful incremental in 26 h; `error` assert; TMDB age > 170 d; breaker open > 1 h; `failed_jobs` growth.

## 11. Open issues

1. A-05: Open Library licence unread; connector off until `docs/legal/open-library.md` records the decision (blocking task before the ingestion milestone, 01-2/06-1). Owner: architect. Closes with that file.
2. Twitch Developer Services Agreement unread. Decided (D-36): owner records `docs/legal/igdb.md` before the ingestion milestone; interim rule in §7.1. Owner: owner.
3. No WebP encoder in `contrib/imaging`; plan: pinned `cwebp` via `contrib/engine`. Owner: vukem. Closes with FUN-08 and a size parity test (±10 % vs Pillow).
4. Cover key. Decided (06-4): legacy `covers/<id>-<hash16>.webp` kept; SDD 01 §5.4 aligned.
5. Queue priority/concurrency of the database driver inside `serve` undocumented. Owner: vukem. Closes with a `vukem-feedback.md` answer.
6. Manual triggers need custom admin actions (VF-008). Owner: vukem. Closes when VF-008 ships or is rejected.
7. No documented API for app-defined OTel metrics. Owner: vukem. Closes with a docs entry.
8. Steam and GOG terms unread. Decided (06-8): images site-only, store IDs for links only; terms read before enabling images in the app (roadmap task). Owner: architect.
9. New Wikidata role properties and the 15-cast cap unverified. Owner: architect. Closes with recorded fixtures.
10. TMDB changes/exports and IGDB webhooks unused in v1. Decided (06-10): accepted; revisit when the weekly TMDB run exceeds 2 h or the IGDB sync exceeds 1 h.
11. Per-host throttle and breaker are app code; `httpclient` candidate in `vukem-feedback.md`. Owner: vukem. Closes when answered.
