# SDD 08: Recommendation

Status: reviewed (2026-10-28 consistency pass). Decisions: D-01, D-02, D-07, D-13, A-01, A-02, A-07. ADRs: 0003, 0005, 0006. Related SDDs: 01, 02, 03, 04, 05, 07, 09, 12.

## 1. Objective

This domain drives discovery from HexaScore's own data. It produces three things:

- a "Top 10 trending (30 days)" list per master category (D-02), served by the API and the snapshot;
- a "Top 3 trending (7 days)" strip per master category on the home page (owner decision 2026-10-31, §7.2 "Home window");
- "You may also like" and cross-media blocks on item pages;
- an explainable "For you" list for logged-in users (A-02).

Interest stays separate from quality (SDD 02), as TMDB does with `popularity` vs `vote_average` (`web-research.md` §2.3). Like IGDB PopScore, we store raw daily primitives and "define our own trend metrics" (§3.3, §9.3). That makes the primitives a paid-API product no film/TV/book competitor offers.

## 2. Scope

In v1:

- Primitives, trend score, hourly snapshots, cold start and exclusions via SDD 01 `items.visibility` (D-02, open issue 08-3).
- Same-type similar items, plus cross-media via `based_on`.
- Content-based "For you".
- Lists as a tie-break signal.
- Contracts.

Out of v1:

- Collaborative filtering, until the §7.6 trigger (A-02).
- A "hype" list.
- Similar items in `/api/v1` (Open issue 9: when added, the API variant uses allowlisted-source relations only).
- Historical snapshots via the API.

Legacy baseline: there is no trending and no similarity.

- The home page renders `SectionHighlights` per type (`legacy/apps/web/app/[locale]/page.tsx`): three 6-row lists sorted by `-score`, `-release_date` and `-rating_count`.
- "Popular" is a static sort on `items.popularity`, the Wikidata sitelinks (`legacy/apps/web/lib/catalog/repository.ts` `sortExpression`).
- Similarity appears only as pg_trgm title search.

IMDb MOVIEmeter counts only views and is "a public-interest signal rather than quality" (§1.1). For that reason views are damped here.

## 3. Functional requirements

- FR-08-1: One view per visitor (session id; for `/app/v1` item reads the access-token id; else IP grouped by `web.Services.IPv6Prefix`) per item per UTC day (open issue 10-5). Verify: integration test, 5 GETs → `views = 1`, same for 5 app reads with one token.
- FR-08-2: The rollup is idempotent. Verify: running it twice gives identical rows.
- FR-08-3: The trend score and eligibility follow §7.2. Verify: golden unit test, thresholds at 4.99/5.0.
- FR-08-4: Snapshots have ≤ 10 gapless ranks per type, ordered `(T desc, popularity_prior desc, id)`. Verify: determinism test.
- FR-08-5: Missing ranks are filled by §7.3 with `basis:"prior"`, which is labelled in the UI. Verify: integration test with 3 eligible items.
- FR-08-6: Unreleased, cancelled and non-`public` (`items.visibility` `hidden`/`takedown`, SDD 01 FR-01-14) items never appear in any list. Verify: integration test.
- FR-08-7: 60 ratings from accounts younger than 7 days do not move an item into the top 10. Verify: integration test (§10).
- FR-08-8: `item_similar` holds ≤ 20 neighbours per item and the page shows 8. `crossMedia` shows ≤ 6 items and is hidden when empty. Verify: integration tests.
- FR-08-9: "For you" shows ≤ 20 items the user has not rated or statused, with ≤ 3 per franchise, ≤ 4 per primary genre, and one explanation each. Verify: unit test.
- FR-08-10: Personalised recommendations are on by default; setting `users.personalization_opt_out` (SDD 03 FR-03-19) removes the profile and its rows and stops "For you" (D-29). Verify: integration test.
- FR-08-11: The legacy listing pattern applies (`legacy/docs/frontend.md` "Padrão de listagens"): one `CategoryLabel` group per type in `ITEM_TYPES` order, `CoverThumb` before titles, empty groups hidden. Verify: Playwright + axe.

## 4. Non-functional requirements

- NFR-08-1: Home server time p95 ≤ 200 ms at 50 rps with the cache warm. The trending payload lives in `cache.NewMemory` (vukem `https://vukem.dev/docs/digging-deeper/cache` "Cache"); a miss costs one query. Verify: k6.
- NFR-08-2: Compute time, both checked by a Go benchmark on seeded Postgres:
  - hourly trending: ≤ 30 s at 20k active items/day;
  - nightly similar: ≤ 60 min at 500k items on 1 core (D-08).
- NFR-08-3: Storage, checked by the monthly size report (SDD 11):
  - `item_activity_daily` ≤ 300 MB, 400 days, active rows only;
  - `item_similar` ≤ 250 MB;
  - snapshots ≈ 86k rows.
- NFR-08-4: A crash loses at most 60 s of views and clicks (ADR 0003 trade-off). Durable primitives lose nothing.
- NFR-08-5: No per-user view or click data is persisted. Dedup keys are `HMAC-SHA256(daily_salt, visitor|item)`, held in memory, with the salt rotated at 00:00 UTC. Verify: test that no raw IP is stored.
- NFR-08-6: "For you" p95 ≤ 400 ms on a miss and ≤ 30 ms on a hit. Verify: k6.

## 5. Data model

```sql
create table item_activity_daily (          -- UTC day
  item_id uuid not null references items(id) on delete cascade,
  day date not null, item_type text not null,
  views int not null default 0, search_clicks int not null default 0,   -- flush adds
  ratings_new int not null default 0, ratings_updated int not null default 0, -- rollup replaces
  status_want int not null default 0, status_in_progress int not null default 0,
  status_completed int not null default 0, status_dropped int not null default 0,
  list_adds int not null default 0, suggestions int not null default 0,
  -- *_w: weighted sums over distinct users, never exposed
  ratings_new_w real not null default 0, ratings_updated_w real not null default 0,
  want_w real not null default 0, in_progress_w real not null default 0,
  completed_w real not null default 0, dropped_w real not null default 0,
  list_adds_w real not null default 0, suggestions_w real not null default 0,
  primary key (item_id, day));
create index on item_activity_daily (item_type, day);

create table trending_snapshots (           -- 90 days
  computed_at timestamptz not null, item_type text not null,
  rank smallint not null check (rank between 1 and 10),
  item_id uuid not null references items(id) on delete cascade,
  trend_score real not null, basis text not null check (basis in ('trend','prior')),
  primary key (computed_at, item_type, rank), unique (computed_at, item_type, item_id));

-- No trending_exclusions table: exclusion is SDD 01 items.visibility <> 'public' (open issue 08-3).

create table item_similar (
  item_id uuid primary key references items(id) on delete cascade,
  similar_ids uuid[] not null, scores real[] not null, cooccur int[] not null,
  feature_hash bytea not null, computed_at timestamptz not null);

-- item_links (based_on, Wikidata P144) is owned by SDD 01 §5.2 and filled by SDD 06 FR-06-11 (08-1).

create table user_recommendations (         -- 24 h cache
  user_id bigint not null references users(id) on delete cascade, item_type text not null,  -- users.id is bigint (migration 20261031203200; SDD 03 §5 says text, wrong)
  item_ids uuid[] not null, reason_item_ids uuid[] not null,
  events_at_compute bigint not null, computed_at timestamptz not null,
  primary key (user_id, item_type));
```

This SDD needs two columns owned elsewhere:

- `list_items.added_at` (SDD 14 §5). Legacy has no timestamp (`legacy/apps/web/lib/db/schema/social.ts`); migrated rows take the list's `created_at`.
- `users.personalization_opt_out boolean default false` (SDD 03 §5).
- `items.visibility` (SDD 01 §5.2).

The SDD 01 item merge adds the source's activity rows into the target's and deletes the source's `item_similar` row.

## 6. Contracts

Errors follow RFC 9457 (vukem `https://vukem.dev/docs/basics/error-handling` "Errors"). Items pass through SDD 01's audience serialiser, which applies the ADR 0006 allowlist to `/api/v1`. `rank`, `trendScore`, `basis` and the primitives are HexaScore-owned (A-07).

| Method, path | Auth | Output | Errors |
| --- | --- | --- | --- |
| GET `/api/v1/{type}/trending` (`listTrending`) | API key | `{computedAt, data:[{rank, basis, trendScore, item}]}` | 400 |
| GET `/api/v1/{type}/{id}/activity?from=&to=` (`getItemActivity`) | business plan (SDD 05 FR-05-17); counts as 1 request | raw daily counts, ≤ 90 days within the last 365 | 403 `plan-required`, 404, 422 |
| GET `/app/v1/{type}/trending` | token `catalog:read` (A-01) | same + display fields | 400 |
| GET `/app/v1/items/{id}/related` | token `catalog:read` | `{similar, crossMedia}` | 404 |
| GET `/app/v1/me/for-you?type=` | token `catalog:read` | `{basis:"profile"\|"onboarding", data:[{item, reason}]}` | 400 |

Trending responses send `web.ETagOf(payload)` and `c.NotModified(etag, computedAt)` (vukem `https://vukem.dev/docs/rest-apis/overview` "Cross-cutting"), plus `Cache-Control: public, max-age=300`.

```json
{"computedAt": "2026-10-28T14:00:00Z",
 "data": [{"rank": 1, "basis": "trend", "trendScore": 118.97,
   "item": {"id": "0192f0a1-…", "type": "game", "slug": "hollow-knight-silksong",
     "title": "Hollow Knight: Silksong", "score": 8.6, "ratingCount": 41}}]}
```

Inertia props (vukem `https://vukem.dev/docs/frontend/inertia` "Inertia v3 (BETA)"):

- Home:
  - `trending` is eager and SSR-rendered (D-07: public SEO content is eager; `Defer` only for personalised or below-the-fold non-SEO props, open issue 08-8), shaped `{game|movie|series|book: {computedAt, items:[{rank, basis, id, slug, title, coverThumbUrl, score, ratingCount}]}}` with at most 3 items per type over the last 7 days (not the snapshot, §7.2 "Home window").
  - `forYou` is `inertia.Defer(fn).Group("personal")`, logged-in only.
  - The home section title is "Trending this week" and prior entries show a "Popular" chip. If every entry is prior, the title becomes "Popular {type}". The `/app/v1` and `/api/v1` trending lists keep the 30-day window and 10 ranks.
- Item page: `similar` (8, each with `reason` such as `{"kind":"franchise","entity":"Zelda"}`) and `crossMedia`, both eager.
- `ForYou/Index`: one `Defer` section per type.

Inputs:

- Durable sources:
  - `rating_events` and the status/time-to-beat activity events, by the names in SDD 02 §6;
  - `suggestion_events` rows with status `submitted` (SDD 04);
  - `list_items.added_at`.
- `SearchResultClicked{ItemID, QueryHash, Position, Locale, At}` from SDD 07 §7.8 (open issue 08-5); only `ItemID` feeds `search_clicks`.
- Queued `user.banned` (SDD 03), which dispatches `trending_rebuild`.

Jobs (vukem `https://vukem.dev/docs/digging-deeper/scheduler` "Scheduler" and "Queue", both BETA):

- `s.Func("trending:flush", flush).EveryMinute()` swaps the counter map and runs one `insert … on conflict do update set views = views + excluded.views`. It also runs on `app.OnShutdown` (`https://vukem.dev/docs/database/mongodb` "MongoDB").
- `s.Job(jobs.ComputeTrending{}).Cron("5 * * * *").WithoutOverlapping()` runs the rollup, then the snapshot, then cache invalidation.
- `s.Job(jobs.ComputeSimilar{}).DailyAt("02:00").WithoutOverlapping()` and `s.Job(jobs.PruneTrending{}).DailyAt("04:45")`.
- Queue job `recommend_user {UserID, Type}` with `queue.Retries(2)` and `queue.Timeout(30*time.Second)`. Queue job `trending_rebuild {ItemIDs, FromDay}`.

Metrics (names from the SDD 12 §6.2 catalogue, open issue 12-6):

- compute time: `trending.compute.duration`, `similar.compute.duration`, `recommend.compute.duration` (histograms, seconds);
- list quality: `trending.eligible_items`, `trending.prior_slots`, `trending.churn` (new entries per hour / 10), `trending.snapshot_age_seconds`, `similar.empty_ratio`;
- click-through (D-35, no app metrics): INFO log lines `event=reco_impression` (sampled 1 in 10, `sample_rate` attribute) and `event=reco_click` with `surface`, `type` and `item`, counted in Railway log search (`app/recommend/related/events.go`). Links carry `?from=trending|similar|crossmedia|foryou`, which the canonical URL drops.

Privacy (LGPD):

- Views and clicks exist only as daily aggregates.
- A user's own activity feeds only that user's recommendations and the aggregates.
- Legal basis (D-29): legitimate interest for aggregates and trending; personalised recommendations on by default with an opt-out in settings (`users.personalization_opt_out`); the privacy policy states it; legal review before cutover (D-21 gate).
- Opting out or deleting the account removes `user_recommendations`.

## 7. Behaviour and algorithms

### 7.1 Primitives

Middleware counts 200 responses to item-page GETs. `/app/v1` item detail reads count the same way, deduplicated per access-token id per day like sessions (open issue 10-5). Excluded: `Purpose: prefetch` requests, partial reloads, bot user agents (SDD 13), and anything past 50 distinct items per visitor per day.

Dedup uses `cache.Add` ("atomic exclusive claim") on `cache.NewMemory(300_000)` with a TTL until 00:00 UTC. A buffered counter with periodic flush is [NOT IN VUKEM DOCS] (`vukem-feedback.md` candidate).

The rollup replaces durable columns for UTC `today` and `today−1`. Each user counts once per (item, day, primitive) at their maximum weight. Weights reuse SDD 02 §7.1 and §7.5:

- ratings: `rating_events.weight × coalesce(rating_flags.shadow_weight, 1)`;
- status, list and suggestion events: `a_u·s_u`, with age tiers 0.25/0.5/0.75/1.0 and 0 if banned.

Lists count only when public, not hidden, and owned by an account at least 7 days old. Burst flags land within 24 h, inside the replace window.

### 7.2 Trend score

With `a = today − day ∈ {0..29}`:

```text
D(a) = 2^(−a/7)                                       half-life 7 days
Q_d  = 10 · ln(1 + views + 2·search_clicks)           anonymous, damped
U_d  = 6·ratings_new_w + 1·ratings_updated_w + 3·want_w + 4·in_progress_w
     + 5·completed_w + 1·dropped_w + 4·list_adds_w + 2·suggestions_w
T_i  = Σ_a D(a) · (Q_d + U_d)
A_i  = Σ_a (ratings_new_w + want_w + in_progress_w + completed_w + list_adds_w)
eligible ⇔ A_i ≥ 5 ∧ T_i ≥ 30 ∧ released ∧ not excluded
```

Golden example, item X:

| `a` | Activity | Q | U | Contribution |
| --- | --- | --- | --- | --- |
| 0 | views 40, `ratings_new_w` 3, `want_w` 2 | 37.14 | 24 | 61.14 |
| 7 | views 100, `completed_w` 4, `list_adds_w` 2 | 46.15 | 28 | ×0.5 = 37.08 |
| 14 | views 200, `ratings_new_w` 5 | 53.03 | 30 | ×0.25 = 20.76 |

Result: `T = 118.97`, `A = 16`, eligible.

Home window (owner decision 2026-10-31): the home shows the trending of the last 7 days, 3 entries per type. It reuses the formula above with `a ∈ {0..6}` and the same eligibility and prior-fill rules (§7.3), computed on demand from `item_activity_daily` and held 5 minutes in the cache (`home:trending7:v1:{locale}`); nothing is stored in `trending_snapshots`, whose rows stay the 30-day top 10 served by the APIs. A quiet week therefore falls back to `prior` entries, labelled "Popular".

Burst example, item Y on day 0: 300 views and 60 ratings from accounts younger than 7 days with no status. Each rating weighs 0.175 × 0.25 (burst shadow), so `ratings_new_w = 2.625`. That gives `T = 57.07 + 15.75 = 72.8` and `A = 2.625`: not eligible. Unweighted, the same activity would score `T = 417.1`.

### 7.3 Snapshot and cold start

Per type, in one transaction:

1. Compute T and A, joined to `items` (SDD 02 §7.4 release gate, `release_status = 'released'`) with `visibility = 'public'` (SDD 01).
2. Take the top 10 by `(T desc, popularity_prior desc, id)`.
3. Fill the remaining ranks (`prior`) by `F_i = ln(1 + popularity_prior) + 3·max(0, 1 − days_since_release/180)`. For example, 80 sitelinks released 30 days ago gives 6.89, and a classic with 200 sitelinks gives 5.30.
4. Upsert with `computed_at = date_trunc('hour', now())`.

Whole-day decay keeps reruns within an hour identical. After commit the job invalidates `trending:v1:{type}`. On failure, readers keep the last snapshot. A snapshot older than 3 h raises an alert.

### 7.4 Similar items

Features are `item_entities` pairs `(role, entity)`. Cast counts only for `position < 5`; franchise `part_of` parents count at half weight.

```text
g_f = ω_role · max(0.1, ln(N_t / (1 + n_f)))
ω: franchise/series 5; author/director/creator/developer 3; writer 2;
   publisher/production_company/network 1.5; cast, genre 1;
   platform, game_mode 0.5; country, original_language 0.3
sim(i,j) = Σ_{f∈F_i∩F_j} g_f / Σ_{f∈F_i∪F_j} g_f          weighted Jaccard
```

Candidates share a feature with `n_f ≤ 5%·N_t`. A candidate is kept if `sim ≥ 0.05` and it either shares a feature with `ω ≥ 1.5` or has `sim ≥ 0.15`.

Order: `round(sim, 2)`, then `cooccur` (distinct qualifying public lists of ≤ 100 items containing both), then popularity, then `id`.

Example: two Zelda games share franchise (16.0), developer (12.3) and genre (0.9), with a union weight of 41.5, so sim = 0.70.

The job uses `Chunk(ctx, 500, …)` (vukem `https://vukem.dev/docs/database/orm` "Query builder"), one transaction per batch and a 50 ms pause, and skips the write when nothing changed. `reason` is the top shared feature. `crossMedia` reads `item_links` in either direction. Two variants share the job (open issue 08-9): the site variant uses every relation; an API variant, if `/api/v1` ever serves similar items, uses only `item_entities`/`item_links` rows whose `source` is in `CATALOG_API_SOURCES` (same split as search, 07-7).

### 7.5 "For you"

```text
μ_u   = mean overall_u if N_u ≥ 5 else 60
ρ_ui  = clamp((overall_ui − μ_u)/50, −1, 1); unrated: completed +0.3, want +0.2, dropped −0.5
τ_ui  = 2^(−age_days/180)
p_u[f] = Σ_i ρ_ui · τ_ui · g_f · 1[f ∈ F_i]
s(u,j) = Σ_f p_u[f]·g_f·1[f ∈ F_j] / sqrt(Σ_{f∈F_j} g_f²) + 0.1·B_j/100   (B_j if N_j ≥ 5)
```

Candidates come from two sources: the `item_similar` rows of the user's top 30 items, and the top 200 items of their 20 strongest features. Items already rated or statused, unreleased or excluded are removed. A greedy re-ranker then applies the FR-08-9 caps; "primary genre" is the genre at `position = 0`.

The explanation names the top contributor: "Because you rated {title} {stars}". With fewer than 3 positive items, the user gets the onboarding message "Rate 3 {type} to get recommendations" plus trending.

A row older than 24 h, or one with ≥ 5 newer `rating_events`, is served stale while `recommend_user` refreshes it. A user with no row gets a synchronous compute.

### 7.6 Collaborative filtering gate

Trigger: ≥ 5,000 users with ≥ 10 counted ratings each, checked monthly.

Protocol: temporal hold-out of each user's latest 20% of ratings (min 2), with overall ≥ 70 as positives. CF is compared with §7.5 on precision@10 and nDCG@10, using the content baseline recorded at launch + 90 days.

CF ships only through a new ADR, and only with:

- ≥ 10% precision@10 lift;
- a 95% bootstrap CI that excludes 0;
- coverage ≥ 80% of the baseline.

## 8. Edge cases

- Merged item: activity is summed. A snapshot row for a deleted id cascades away and readers tolerate the gap.
- Release date moves into the future: the item leaves every list at the next run.
- Open SDD 02 burst: the item stays listable with shadow-weighted ratings.
- No released or active item in a type: the section is hidden.
- Deleted rating: it stops counting today, while past days keep it.
- Flush failure: the counts merge into the next flush. Above 100k keys the oldest are dropped and `trending.counter.dropped` (SDD 12 §6.2) is incremented.
- Opt-out while a job is queued: the job checks the flag and exits.

## 9. Acceptance criteria

- [ ] `go test ./app/recommend/... -run 'TestTrendScore|TestFallback|TestSimilarity|TestProfile'` exits 0, with golden values 118.97 and 72.8 (±0.01).
- [ ] `go test ./app/recommend/... -run TestSnapshotDeterminism -count=3` exits 0.
- [ ] `go test -tags=integration ./app/recommend/... -run TestBurstDoesNotEnterTop10` exits 0.
- [ ] Staging: `psql "$DATABASE_URL" -c "select item_type, count(*) from trending_snapshots where computed_at = date_trunc('hour', now()) group by 1"` returns 4 rows of 10.
- [ ] `curl -sI …/api/v1/game/trending -H "Authorization: Bearer $KEY"` returns an `ETag`, and the same request with `If-None-Match` returns 304. `…/activity` with a non-business key returns 403.
- [ ] `k6 run tests/k6/home.js`: `p(95) < 200ms` with the cache warm.
- [ ] Schemathesis passes on `listTrending` and `getItemActivity`, and the responses contain no TMDB/IGDB fields.

## 10. Test strategy

- Unit: `D(0, 7, 14) = 1, 0.5, 0.25`; the §7.2 examples; threshold edges; tie-breaks; `F_i`; `sim` symmetry and bounds; `ρ` clamping; caps; explanations.
- Determinism: run `ComputeTrending` twice within one fake-clock hour, with insert order shuffled, and expect identical rows.
- Integration (Postgres): 10 organic items (`A ≥ 8`). Then 60 accounts aged 2 days rate item Z, which opens the SDD 02 burst, plus 300 distinct-IP views. Z must stay out and ranks 1–10 must not change. Also covered: rollup idempotency, merge, ban rebuild, opt-out and dedup.
- Contract: Schemathesis (SDD 05).
- E2E: Playwright + axe.
- Load: k6 on the home page and the API.

## 11. Open issues

1. `item_links` (P144). Decided (08-1): in SDD 01 §5.2 and SDD 06 FR-06-11.
2. Lists owner. Decided (08-2): SDD 14.
3. Exclusions. Decided (08-3): SDD 01 `items.visibility`; no exclusions table.
4. `personalization_opt_out`. Decided (08-4): SDD 03 §5, privacy settings.
5. `SearchResultClicked`. Decided (08-5): defined in SDD 07 §7.8.
6. Activity endpoint plan. Decided (08-6): `business` only, 1 request (SDD 05 FR-05-17).
7. Weights, half-life and thresholds are untuned. Owner: architect. Closes at the 90-day post-launch review.
8. `trending` eager. Decided (08-8): correct reading of D-07; SDD 09 follows.
9. Similar items in the API. Decided (08-9): allowlisted-source relations only for the API variant (§7.4).
10. A buffered counter flush and the `HourlyAt` argument are [NOT IN VUKEM DOCS]. Owner: vukem. Closes when filed in `vukem-feedback.md`.
11. The similar-items compute at 500k items is unmeasured. Owner: architect. Closes by a benchmark; above 60 min, switch to incremental by `feature_hash`.
12. LGPD legal basis. Decided (D-29): legitimate interest + opt-out, on by default; legal review before cutover (D-21).
13. A "hype" list for unreleased items. Owner: owner. Closes by a product decision after v1.0.0.
