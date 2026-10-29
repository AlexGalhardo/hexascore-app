# SDD 02: Scoring

Status: reviewed (2026-10-28 consistency pass). Decisions: D-00, D-02, D-13, D-15, A-04, A-06. ADRs: 0005. Related SDDs: 01, 03, 05, 08, 09, 10, 12, 13.

## 1. Objective

A score users can trust: one-click half-star input in six categories (D-00), an aggregate shown as stars 0–10 with one decimal, a count and a confidence level (D-13), and a published, versioned formula with published anti-manipulation rules (web-research 9.2). Score is "quality"; trending (D-02, SDD 08) is "interest", fed by primitives emitted here. Legacy baseline: `legacy/apps/web/lib/ratings.ts`, `legacy/apps/web/lib/db/schema/social.ts`, `legacy/apps/web/lib/categories.ts`, `legacy/.claude/skills/hexascore-ratings/SKILL.md`.

## 2. Scope

In scope (v1): `ratings`, `item_status`, `time_to_beat` and their write paths (web and first-party app API, A-01); formula v1, priors, confidence, denormalised `items` columns (owned here, referenced by SDD 01); anti-manipulation (release gate, burst limiter, weights, daily caps, shadow weight, notice, admin review per D-15); legacy migration (A-04); `rating_events` and events for SDD 08; category i18n; help page outline.

Out of scope: written reviews, critic scores (D-09), feed/follows (D-06), collaborative filtering (A-02), trending computation (SDD 08), API plans (SDD 05), bans (SDD 03; read here as `banned_at`).

## 3. Functional requirements

- FR-02-1: Six category values per rating, each 0.5–10.0 stars in 0.5 steps, stored as 5–100 in steps of 5; anything else → 422. Verified: `z` schema unit test; Postgres CHECK test.
- FR-02-2: No rating = no row (zero is not a value); one row per (user, item); re-submit updates and keeps `created_at`. Verified: integration test.
- FR-02-3: Ratings before the effective release date (7.4) → 409 `rating-not-released`. Verified: unit + route test.
- FR-02-4: Every rating, status, time-to-beat or flag write locks the item row and recomputes its aggregate in the same transaction. Verified: Postgres concurrency test (9).
- FR-02-5: The aggregate follows formula v1 (7.1); the API exposes only `score`, `scoreByCategory`, `ratingCount`, `confidence`, `formulaVersion`, on 0–10. Verified: golden unit test; Schemathesis (SDD 05).
- FR-02-6: `score`, `scoreByCategory`, `confidence` are `null` while `ratingCount < 5`. Verified: integration test 4→5.
- FR-02-7: A nightly job rebuilds priors, recomputes every rated item and logs drift (stored count ≠ recount). Verified: corrupt a row, run job.
- FR-02-8: Per-user daily caps (7.6) → 429 with `Retry-After`. Verified: fake-clock test.
- FR-02-9: The burst detector (7.5) runs on every insert and, when tripped, activates the limiter and the public notice. Verified: seeded-burst test.
- FR-02-10: Anti-abuse never deletes ratings: it sets a shadow weight in `rating_flags`; every action is reversible and logged in `score_actions`. Verified: flag → restore yields the original aggregate exactly.
- FR-02-11: Status is `want | in_progress | completed | dropped`; time-to-beat only for games and only while `completed`; leaving `completed` deletes it in the same transaction (parity). Verified: tests ported from legacy `lib/time-to-beat.ts`.
- FR-02-12: Each write appends one `rating_events` row in the transaction and dispatches a typed event after commit. Verified: `events.NewFake()` + row count.
- FR-02-13: Legacy migration rounds to the nearest multiple of 5 (clamped 5–100), one audit row per changed rating plus one summary row. Verified: fixture test (9).
- FR-02-14: Category labels in en, pt-BR, es; widget `role="slider"`, `aria-valuetext` "7.5 of 10 stars" localised. Verified: key-completeness test; Playwright + axe.

## 4. Non-functional requirements

- NFR-02-1: Rating write p95 ≤ 60 ms server time on an item with 10,000 ratings. Verified: k6 on staging (`qa/k6/rating-hot-item.js`). Feasibility measured in P1.D.2.2: the recompute statement alone is p95 28.6 ms on 5,000 ratings and 48.4 ms on 10,000 (local Postgres 18.6, `TestRecomputeLatencyHotItem`), so no covering index was added; the end-to-end write figure from staging is still to be recorded.
- NFR-02-2: Read path uses denormalised columns only. Verified: query-count test per request.
- NFR-02-3: Nightly recompute of 100,000 rated items ≤ 20 min, batches of 500 items, 100 ms pause (D-08 single process). Verified: timed staging run.
- NFR-02-4: No lost updates with 50 concurrent writers on one item; no deadlocks across rating, status and nightly paths. Verified: Postgres test, 100 runs.
- NFR-02-5: Flags, weights and detector data are admin-only; a user's individual ratings are public only via the profile (SDD 03); `rating_events` kept 180 days, pruned daily. Verified: authorisation and prune tests.
- NFR-02-6: Formula changes ship only as a new `score_formula_versions` row announced on the help page and API changelog (SDD 05) ≥ 14 days before activation. Verified: release checklist.
- NFR-02-7: Accessibility: star widget keyboard-operable (Left/Right ±0.5, Home 0.5, End 10, Delete clears), focus visible, 44×44 px target. Verified: Playwright + axe (SDD 09/10).

## 5. Data model

User id type follows SDD 03 (shown as `uuid`). Item columns are owned here; SDD 01 references them.

```sql
CREATE TABLE ratings (
  user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  item_id    uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  c1 smallint NOT NULL, c2 smallint NOT NULL, c3 smallint NOT NULL,
  c4 smallint NOT NULL, c5 smallint NOT NULL, c6 smallint NOT NULL,
  -- unrounded per-rating mean (D-13 keeps full precision; legacy rounded it)
  overall    numeric GENERATED ALWAYS AS ((c1+c2+c3+c4+c5+c6)::numeric / 6) STORED,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, item_id),
  CONSTRAINT ratings_half_star CHECK (
    c1 BETWEEN 5 AND 100 AND c1 % 5 = 0 AND c2 BETWEEN 5 AND 100 AND c2 % 5 = 0 AND
    c3 BETWEEN 5 AND 100 AND c3 % 5 = 0 AND c4 BETWEEN 5 AND 100 AND c4 % 5 = 0 AND
    c5 BETWEEN 5 AND 100 AND c5 % 5 = 0 AND c6 BETWEEN 5 AND 100 AND c6 % 5 = 0)
);
CREATE INDEX ratings_item_idx ON ratings (item_id, created_at);
CREATE INDEX ratings_user_idx ON ratings (user_id, created_at);

CREATE TYPE item_status_value AS ENUM ('want','in_progress','completed','dropped');
CREATE TABLE item_status (          -- legacy parity
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  item_id uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  status item_status_value NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, item_id));
CREATE INDEX item_status_user_idx ON item_status (user_id, status);

CREATE TABLE time_to_beat (         -- legacy parity: minutes, games only, only while completed
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  item_id uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  main_minutes int, extras_minutes int, completionist_minutes int,
  created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, item_id),
  CHECK ((main_minutes IS NULL OR main_minutes BETWEEN 30 AND 600000) AND
         (extras_minutes IS NULL OR extras_minutes BETWEEN 30 AND 600000) AND
         (completionist_minutes IS NULL OR completionist_minutes BETWEEN 30 AND 600000)),
  CHECK (coalesce(main_minutes, extras_minutes, completionist_minutes) IS NOT NULL));
CREATE INDEX time_to_beat_item_idx ON time_to_beat (item_id);

-- Create after score_bursts. Keyed by (user, item), not FK to ratings: survives delete + re-create, so deleting does not launder a flag.
CREATE TABLE rating_flags (
  user_id uuid NOT NULL, item_id uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  shadow_weight numeric(3,2) NOT NULL CHECK (shadow_weight BETWEEN 0 AND 1),
  reason text NOT NULL CHECK (reason IN ('burst_hold','admin_exclude','admin_downweight')),
  burst_id bigint REFERENCES score_bursts(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, item_id));

CREATE TABLE score_bursts (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  item_id uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  started_at timestamptz NOT NULL, limited_until timestamptz NOT NULL,
  trigger text NOT NULL CHECK (trigger IN ('volume','skew')),
  stats jsonb NOT NULL,                       -- window counts, baseline, means
  resolution text CHECK (resolution IN ('approved','excluded','auto_released','kept_held')),
  resolved_at timestamptz, resolved_by uuid);
CREATE INDEX score_bursts_open_idx ON score_bursts (item_id) WHERE resolution IS NULL;

CREATE TABLE score_actions (                  -- append-only audit of every anti-abuse/admin/migration action
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  at timestamptz NOT NULL DEFAULT now(), actor uuid,  -- NULL = system
  kind text NOT NULL,     -- burst_open, hold, release, exclude, restore, formula_activate, legacy_rounding
  item_id uuid, user_id uuid, burst_id bigint,
  before jsonb, after jsonb, note text);

CREATE TYPE rating_event_kind AS ENUM ('rating_created','rating_updated','rating_deleted',
  'status_changed','time_to_beat_reported','time_to_beat_cleared');
CREATE TABLE rating_events (                  -- source of truth for SDD 08 primitives; 180-day retention
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  at timestamptz NOT NULL DEFAULT now(),
  kind rating_event_kind NOT NULL,
  user_id uuid NOT NULL, item_id uuid NOT NULL, item_type item_type NOT NULL,
  overall_before numeric, overall_after numeric,
  status_before item_status_value, status_after item_status_value,
  weight numeric(4,3));                       -- weight at write time, for trending (D-02)
CREATE INDEX rating_events_item_idx ON rating_events (item_id, at);
CREATE INDEX rating_events_user_idx ON rating_events (user_id, at);

CREATE TABLE score_formula_versions (
  version smallint PRIMARY KEY,
  m_game smallint NOT NULL, m_movie smallint NOT NULL, m_series smallint NOT NULL, m_book smallint NOT NULL,
  params jsonb NOT NULL,                      -- every threshold in 7.x, so a version is self-describing
  activated_at timestamptz NOT NULL, changelog text NOT NULL);

CREATE TABLE score_priors (                   -- rebuilt nightly
  item_type item_type NOT NULL, formula_version smallint NOT NULL REFERENCES score_formula_versions,
  prior_c numeric[6] NOT NULL,                -- C_{t,k}, 0–100
  items_used int NOT NULL, computed_at timestamptz NOT NULL,
  PRIMARY KEY (item_type, formula_version));

CREATE TABLE rating_migration_audit (         -- A-04 one-off; kept 1 year, then dropped by migration
  user_id uuid NOT NULL, item_id uuid NOT NULL,
  old_c smallint[6] NOT NULL, new_c smallint[6] NOT NULL, migrated_at timestamptz NOT NULL,
  PRIMARY KEY (user_id, item_id));

-- Columns on items (owned by SDD 02)
ALTER TABLE items
  ADD rating_count int NOT NULL DEFAULT 0,          -- counted ratings (w > 0)
  ADD rating_weight numeric(12,3) NOT NULL DEFAULT 0,  -- W
  ADD score numeric(7,4),                            -- B, 0–100, full precision; NULL when no counted rating
  ADD score_c1 numeric(7,4), ADD score_c2 numeric(7,4), ADD score_c3 numeric(7,4),
  ADD score_c4 numeric(7,4), ADD score_c5 numeric(7,4), ADD score_c6 numeric(7,4),
  ADD score_se numeric(7,4),                         -- standard error of the weighted mean
  ADD score_confidence text CHECK (score_confidence IN ('low','medium','high')),
  ADD score_formula_version smallint,
  ADD score_computed_at timestamptz,
  ADD ratings_limited_until timestamptz,            -- burst limiter active while now() < this
  ADD unusual_activity_until timestamptz;           -- public notice while now() < this
```

Invariants: like legacy, stored values are the raw aggregate for any count; the `ratingCount < 5` gate applies on read (`legacy/apps/web/lib/categories.ts` `MIN_RATINGS_FOR_SCORE`). Deleted accounts are anonymised, not deleted (legacy `lib/account-purge.ts`), so their ratings keep counting.

Legacy migration (A-04): copy `ratings` after the synthetic users are purged; `new = LEAST(100, GREATEST(5, 5 * round(c / 5.0)))`. Integer inputs never tie (c/5 is never x.5), so no tie rule is needed; 0–2 become 5 (0.5 star). Every row with any changed category gets a `rating_migration_audit` row; one `score_actions` row `legacy_rounding` stores totals (rows read, rows changed, per-category histogram of deltas). `item_status` and `time_to_beat` copy unchanged. After the copy the nightly job runs once before cutover (D-21).

## 6. Contracts

Input is in stars on every surface; the server multiplies by 10. Errors are RFC 9457 problem+json (vukem `https://vukem.dev/docs/basics/validation` "Errors"; validation via `web.Parse`, `https://vukem.dev/docs/basics/validation` "Validation"), problem types under `https://hexascore.app/problems/`.

| Method, path | Auth | Input | Output | Errors |
| --- | --- | --- | --- | --- |
| `PUT /app/v1/items/{id}/rating` | token, ability `ratings:write` (A-01) | `{"scores":{"gameplay":7.5,...}}` all six keys of the item's type | 200 `{rating, item: {score,...}}` | 422 validation; 404; 409 `rating-not-released`; 429 `rating-daily-limit`, `item-ratings-limited` |
| `DELETE /app/v1/items/{id}/rating` | same | none | 204 | 404 |
| `GET /app/v1/items/{id}/me` | token | none | `{rating, status, timeToBeat}` | 404 |
| `PUT/DELETE /app/v1/items/{id}/status` | token, `status:write` | `{"status":"completed"}` | 200/204 | 422, 404 |
| `PUT/DELETE /app/v1/items/{id}/time-to-beat` | token, `status:write` | `{"mainHours":12.5,...}` 0.5 h steps, 0.5–10,000 h | 200/204 | 409 `status-not-completed`, 422 `not-a-game` |

Web (Inertia) uses the same handlers on session routes (SDD 09); validation failures return 303 with `props.errors`. Per-user rate limit: `ratelimit.Middleware(limiters("rating", 30, time.Minute), ratelimit.ByUser(guard))` (vukem `https://vukem.dev/docs/security/rate-limiting` "Rate limiting"), legacy parity with `hitRateLimit("rating:<userId>", 30, 60s)`.

Public read fields (SDD 05 schemas; `/api/v1` and Inertia props):

```json
{"score": 7.3, "scoreByCategory": {"gameplay": 7.6, "fun": 7.4, "sound": 7.0,
  "technical": 7.2, "visual": 7.5, "storytelling": 7.1},
 "ratingCount": 6, "confidence": "low", "formulaVersion": 1}
```

Site-only props: `unusualActivity`, `ratingsLimitedUntil`. `scoreByCategory` keys are the stable category keys (A-06), also API filter names (`min_gameplay`).

Events (vukem `https://vukem.dev/docs/digging-deeper/events` "Events"; dispatched from `tx.AfterCommit`, vukem `https://vukem.dev/docs/database/orm` "Query builder"): `RatingCreated`, `RatingUpdated`, `RatingDeleted`, `StatusChanged`, `TimeToBeatReported`, `ItemScoreChanged{ItemID, Before, After, FormulaVersion}`, `BurstOpened{ItemID, BurstID}`. SDD 08 consumes `rating_events` (and the `item_status`/`time_to_beat` activity events above) by these names as the source of its trending primitives; in-process events are lost on a crash after commit, so SDD 08 reads windows from the table, not the bus.

Subscriptions (open issue 02-8): this domain listens to SDD 03's `user.banned` and `user.unbanned` events and, per event, enqueues `RecomputeItemScore{ItemID}` (`q.DispatchTx`) for every item the user has rated, in batches of 500 item ids ordered ascending; the nightly `RecomputeScores` repairs any missed event.

Jobs (vukem `https://vukem.dev/docs/digging-deeper/scheduler` "Scheduler", "Queue"): `s.Job(jobs.RecomputeScores{}).DailyAt("04:00").WithoutOverlapping()` with `SCHEDULE_TIMEZONE=America/Sao_Paulo`; `s.Job(jobs.ReleaseBursts{}).Hourly()`; `s.Job(jobs.PruneRatingEvents{}).DailyAt("04:30")`. `RecomputeItemScore{ItemID}` is dispatched with `q.DispatchTx` after admin flag changes.

Admin (vukem `https://vukem.dev/docs/digging-deeper/admin`): read-only resources for `score_bursts`, `rating_flags`, `score_actions`, `score_formula_versions`; built-in audit log covers CRUD. Bulk "approve/exclude burst" and "restore rating" are custom POST routes under the admin gate (`/admin/scoring/bursts/{id}/approve|exclude`, `/admin/scoring/flags/{user}/{item}/restore`) because custom row actions on a resource are [NOT IN VUKEM DOCS] (`vukem-feedback.md` candidate: "admin resource custom actions").

## 7. Behaviour and algorithms

### 7.1 Formula v1

For item `i` of type `t`, category `k ∈ {1..6}`, counted ratings `u` (those with `w_u > 0`):

```text
w_u     = a_u · v_u · s_u · r_u
N       = |{u : w_u > 0}|                  (ratingCount)
W       = Σ w_u
S_k     = Σ w_u · c_{u,k}                  c in 5..100
B_k     = (S_k + m_t · C_{t,k}) / (W + m_t)
B       = (1/6) · Σ_k B_k                  = (Σ w_u·overall_u + m_t·C_t) / (W + m_t), C_t = mean_k C_{t,k}
stars   = round_half_away(B / 10, 1)       shown only if N ≥ 5
```

Weight factors:

| Factor | Rule (v1 params) |
| --- | --- |
| `a_u` account age at compute time | < 7 days 0.25; 7–29 days 0.5; 30–89 days 0.75; ≥ 90 days 1.0 |
| `v_u` verified activity | no status or `want` 0.7; `in_progress`/`dropped` 0.85; `completed` 1.0; games: `completed` without a time-to-beat report 0.9, with one 1.0 |
| `s_u` shadow | `rating_flags.shadow_weight`, default 1.0; 0 if the user has `banned_at` set (SDD 03) |
| `r_u` release | 0 if `ratings.created_at` < effective release (7.4), else 1 |

`a_u` grows with time, so the nightly job keeps age weights at most 24 h stale. Categories and overall share `W` and `m_t`, so `B` is exactly the mean of `B_k`.

Prior `C_{t,k}`: mean over items of type `t` with `N ≥ 5` of their weighted raw means `S_k / W`, item-level so blockbusters do not dominate; if fewer than 20 such items, `C_{t,k} = 60` (6.0 stars). Rebuilt nightly into `score_priors` before items are recomputed.

Default `m_t` (version 1): game 10, movie 10, series 8, book 8 [NOT VERIFIED against the migrated distribution]; seeded as version 1 by migration `20261029120001_seed_score_formula_v1` (P1.D.1.2). Reminder for P2.F: measure the median N per type on the first staging copy and, if needed, add version 2. Tuning, at most quarterly on production data: `m_t = clamp(median N over items of type t with N ≥ 1, 5, 50)`. Any change to `m_t`, weights, thresholds or the prior rule is a new `score_formula_versions` row (`params` + changelog), activated by migration ≥ 14 days after announcement (NFR-02-6) and followed by a full recompute; API `formulaVersion` is the active version.

Confidence (`SE` on the 0–100 scale, weighted standard deviation `σ = sqrt(Σ w_u (overall_u − S/W)² / W)`, `SE = σ / sqrt(W)`):

| Level | Rule |
| --- | --- |
| none (`null`) | N < 5 |
| low | N ≥ 5 and not medium |
| medium | N ≥ 15 and SE ≤ 4.0 |
| high | N ≥ 50 and SE ≤ 2.0 |

Worked example (game, `m = 10`, `C = 68`): six ratings with overall/weight 90/1.0, 85/1.0, 80/0.85, 95/0.7, 100/0.175 (account 3 days, no status: 0.25·0.7), 75/1.0. `W = 4.725`, `Σ w·overall = 402.0`, raw weighted mean 85.08, `B = (402 + 680) / 14.725 = 73.48`, stars **7.3**. Naive mean: 8.8. `σ = 7.35`, `SE = 3.38`, `N = 6` → `low`. This example is the golden unit test.

### 7.2 Write path

One `db.Tx` (vukem `https://vukem.dev/docs/database/orm` "Query builder", transactions and row locks):

1. `web.Parse` the input; reject non-0.5 multiples; resolve the item and lock it: `QueryOf[Item](tx).Where("id", id).LockForUpdate().First(ctx)` (404 if absent).
2. Release gate (7.4), user daily cap (7.6), item limiter (7.5).
3. Upsert/delete the row; append `rating_events`; on insert run the detector (7.5).
4. Recompute: one SQL statement over `ratings` joined to `users`, `item_status`, `time_to_beat`, `rating_flags`, computing `N, W, S_1..S_6, σ` with the weight rule, then `UPDATE items`. Never incremental.
5. `tx.AfterCommit`: dispatch events, invalidate the item cache (SDD 09/12).

Status and time-to-beat writes lock the item row, then the status row (legacy locked only the latter; status now changes weights), then recompute; so does the nightly job, per item. One lock order (item first; several items by id ascending, legacy `mergeItems`) prevents deadlocks. Item merge (SDD 01) moves ratings, statuses, reports and `rating_flags` (`ON CONFLICT DO NOTHING`, target wins) and recomputes the target in the same transaction.

Why transactional recompute: incremental counters cannot express time-varying weights and drift on any missed path; periodic-only breaks read-your-own-write. The row lock serialises writers per item (correct under concurrency) and the nightly full pass repairs any out-of-band edit (self-healing).

### 7.3 Edit and delete

Edit keeps `created_at`, sets `updated_at`, appends `rating_updated` (before/after overall). Delete removes the row and appends `rating_deleted`; `rating_flags` stays, so a held user who deletes and re-rates is still held. Edits are allowed in `limited` mode; they never clear a flag.

### 7.4 Release gate

Effective release date `E` from SDD 01 fields `release_date` and its precision: day → that date; month → last day of month; year → 31 December. Ratings allowed when `current_date` in America/Sao_Paulo `≥ E` (inclusive). Precision comes from `items.release_precision` (`unknown` with a date behaves as `day`). No release date: allowed unless `items.release_status` is `announced` or `cancelled` (SDD 01 §5.2, open issue 02-1), so old works with missing data stay rateable. Series: first episode date; books: earliest edition of the work (D-10); games: early-access date counts. Statuses are not gated (parity). If a date moves into the future, existing ratings get `r_u = 0` via recompute (kept, not deleted) and count again once it passes.

### 7.5 Burst detector and limiter

Runs in the insert transaction over `rating_created` events. Baseline `b` = mean daily inserts over days −30..−2, floor 1.

- Volume trigger: inserts in the last 24 h ≥ max(30, 8·b), or in the last 1 h ≥ max(15, 3·b).
- Skew trigger: ≥ 20 inserts in the last 24 h whose weighted mean overall differs from the item's score before the window by ≥ 25 points (2.5 stars), with N_before ≥ 10.

On trigger (no open burst): insert `score_bursts`, set `ratings_limited_until = now() + 48 h` and `unusual_activity_until = now() + 7 days`, write `score_actions(burst_open)`, and insert `rating_flags(shadow 0.25, burst_hold)` for every rating created in the window from accounts younger than 30 days or with no status. A new trigger while open extends both timestamps from the new trigger time.

Limiter while `now() < ratings_limited_until`: new ratings are accepted only from accounts ≥ 30 days old with a status other than `want`; others get 429 `item-ratings-limited`, `Retry-After` = seconds to `ratings_limited_until`, message key `scoring.limited`: "Ratings for this title are temporarily limited because of unusually high activity. You can rate it after {time}."
Public notice while `now() < unusual_activity_until`, key `scoring.unusual_notice`: "This title received unusual rating activity. Our published rules were applied and some ratings count less until reviewed. How scores work →" linking the help page.

Resolution: an admin approves (flags deleted) or excludes (shadow 0), target 24 h (D-15); unreviewed bursts older than 14 days are auto-resolved by `ReleaseBursts`: `auto_released` if the skew trigger no longer holds over the window, else `kept_held`. Every transition writes `score_actions` with the before/after aggregate.

### 7.6 Per-user caps

Counted from `rating_events` per UTC day (not from `ratings`, so delete + re-create still counts): new ratings 20/day for accounts < 7 days, 200/day otherwise; edits 100/day. Over cap: 429 `rating-daily-limit`, `Retry-After` until 00:00 UTC.

### 7.7 Time-to-beat

Legacy parity (`legacy/docs/data-model.md`): medians via `percentile_cont(0.5)` per style, computed on read and cached with the item; `null` below 3 reports. Input in hours, 0.5 h steps, stored in minutes.

## 8. Edge cases

- No ratings: all score columns `NULL`, `rating_count 0`, API `null`s.
- N = 5 from five new accounts: displayed, heavily shrunk to the prior, `low`; by design.
- Raters banned later: on `user.banned` this SDD queues a recompute per rated item; N can drop below 5 and the score hides again.
- Item type change or cross-type merge (SDD 01): rejected while ratings exist (category keys differ).
- Formula activated mid-run: the job reads the active version once per run; each item carries `score_formula_version`, so the API is correct per item.
- Account deleted: anonymised, ratings kept, age weight from the original `created_at`.
- `completed` → `dropped`: time-to-beat deleted, `v_u` 1.0 → 0.85, recomputed in the same transaction.
- New release, no baseline: `b = 1`, volume trigger `max(30, 8)`; a launch rush from established accounts with statuses passes the limiter.

## 9. Acceptance criteria

- `go test ./app/scoring/... -run TestFormulaGolden` → passes; the 7.1 example yields `B = 73.4805 ± 1e-4`, stars 7.3, confidence low.
- `go test ./app/scoring/... -run TestFormulaProperties` (stdlib `testing/quick`, 10,000 cases) → passes: raising any `c_{u,k}` never lowers `B_k` or `B`; adding a rating above `B` raises `B`, below lowers it; `B` lies between `C_t` and the raw weighted mean; raising `w_u` moves `B` toward `overall_u`; `B` in [5, 100].
- `TEST_DATABASE_URL=... go test ./tests/scoring -run TestConcurrentWriters -count=100` (vukem `pgtest.Database`, `https://vukem.dev/docs/testing/overview`) → 50 goroutines upserting/deleting on one item plus a nightly recompute; stored columns equal a from-scratch recompute every run; no deadlock errors.
- `go test ./tests/scoring -run TestThreshold` → 4 ratings: API `score: null`; 5th: non-null.
- `go test ./tests/scoring -run TestReleaseGate` → future date 409; year-precision date blocked until 31 Dec; null date allowed; announced item 409.
- `go test ./tests/scoring -run TestBurst` → seeded 40 inserts in 1 h on an item with `b = 2` opens one burst, holds new-account ratings at 0.25, returns 429 to a 3-day account, admin approve restores the aggregate to the value with all flags removed, exactly.
- `go test ./tests/scoring -run TestDailyCap` → 21st insert of a 2-day account returns 429 with `Retry-After`.
- `go test ./tests/migration -run TestLegacyRounding` → 0,1,2,3,12,13,97,98,100 map to 5,5,5,5,10,15,95,100,100; audit rows = changed rows.
- `go test ./tests/i18n -run TestCategoryKeys` → every `categories.<type>.<key>` exists in en, pt-BR, es.

## 10. Test strategy

- Unit + property: formula, weights, confidence, release rule, detector as pure functions with a fake clock; SQL recompute must match the Go function to 1e-6 on random fixtures.
- Integration (Postgres via `pgtest`): write paths, locks, nightly job, merge, ban recompute, retention prune.
- Contract: Schemathesis on SDD 05 schemas (`score` is `number|null`, `multipleOf 0.1`, 0–10).
- E2E: Playwright + axe on widget, notice and help page (SDD 09, 10).
- Load: k6 50 concurrent raters on one hot item (NFR-02-1).
- Data quality asserts (nightly): `rating_count` equals recount; no `score` outside [5,100]; drift count exported as a metric with an alert above 0 (SDD 12).

Help page outline (`/help/score`, localised per D-17): categories per type; half-star input; formula 7.1 with the worked example; current `m` and priors; minimum of 5; confidence; weights; release gate; burst limiter, notice and review/reversal; daily caps; version history.

i18n: labels in `resources/lang/<locale>.json` as `categories.game.gameplay` etc. (vukem `https://vukem.dev/docs/digging-deeper/localization` "i18n"), keys from `legacy/apps/web/lib/categories.ts` (A-06); widget `rating.valuetext` "{value} of 10 stars". How catalogues reach React components through Inertia is [NOT IN VUKEM DOCS]; SDD 09 owns it.

## 11. Open issues

1. Release-date precision and lifecycle field. Decided (02-1): SDD 01 `release_precision` and `release_status`; §7.4 uses them.
2. Default `m_t` and fallback prior 60 unverified; measure median rating count per type on migrated data (A-04). Owner: architect. Closes after the first staging data copy.
3. Recompute latency on 10k-rating items (NFR-02-1) unmeasured; if over budget, add a covering index. Owner: architect. Closes after the k6 run.
4. Admin custom row actions are [NOT IN VUKEM DOCS]; bursts/flags use custom routes. Owner: vukem. Closes when `vukem-feedback.md` entry is answered.
5. orm handling of generated columns (`ratings.overall`) on insert/update is [NOT IN VUKEM DOCS] (legacy Drizzle needed raw SQL). Owner: vukem. Closes with a test or doc entry.
6. Delivering i18n catalogues to Inertia/React. Decided (02-6): SDD 09 serves them as a shared `Once` prop, flattened once in `packages/i18n` for web and mobile (09-5).
7. Weight tiers and detector thresholds lack traffic data; review after 90 days of `rating_events`. Owner: owner + architect. Closes with formula v2 or a recorded "keep".
8. Ban/unban must trigger item recomputes. Decided (02-8): SDD 03 emits `user.banned`/`user.unbanned`; this SDD subscribes (§6). Daily caps (§7.6) and weight tiers (§7.1) are unchanged by the review pass.
