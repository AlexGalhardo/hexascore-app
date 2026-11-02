# RFC-002: Landing-page trending that enters and leaves on its own

Status: proposed (architect audit, 2026-11-02). Owner gates: OQ-07 (seed users). Builds on SDD 08 §7.1-7.3 (activity primitives, trend score, snapshots), the owner decision of 2026-10-31 (home shows 7-day trending, 3 per type) and the owner round of 2026-11-01 item 26 (12 per type, 30 days, carousel). Related: RFC-001 (quality score), T-09 in `docs/audit/02-technical-review.md`.

## 0. Executive summary

SDD 08 already defines a transparent, decayed trend score with anti-manipulation weights, and the hourly `ComputeTrending` job writes the 30-day top 10 per type into `trending_snapshots`. The landing page does not use any of it: `app/controllers/home.go:240-264` lists the 12 items with the most new ratings in 30 days, padded with the all-time most-rated items, cached 5 minutes. With 10,000 seed accounts still weighting ratings, that list is a popularity chart that barely moves. This RFC makes the home read a snapshot produced by the SDD 08 formula, extended with external popularity deltas and release recency, with explicit entry and exit rules, diversity and quality floors, a cooldown so the carousel does not churn every hour, pin/ban controls, and a test plan on synthetic and historical data.

## 1. Goals

Twelve items per type on the landing page that (a) reflect what is gaining attention this week on HexaScore and outside it, (b) rotate on their own, (c) cannot be gamed by a burst of fake accounts, (d) never show an item without a cover, synopsis and translations, (e) do not show four entries of one franchise, (f) can be pinned or banned by an admin, and (g) cost nothing on the request path (one indexed read, cached).

## 2. Signals

All signals are per item and per UTC day, stored in `item_activity_daily` (exists) plus two new columns and one new table.

| Signal | Source | Window | Weight w | Reliability r | Notes |
| --- | --- | --- | --- | --- | --- |
| `ratings_new_w` | HexaScore (`rating_events`, weighted by account age tier, burst shadow weight, banned = 0; SDD 02 §7.5) | 7 d, decay | 6 | 1.0 | seed users excluded (`users.is_seed`, OQ-07) |
| `ratings_updated_w` | same | 7 d | 1 | 1.0 | |
| `want_w`, `in_progress_w`, `completed_w`, `dropped_w` | `item_status` events | 7 d | 3 / 4 / 5 / 1 | 1.0 | |
| `list_adds_w` | public lists of accounts ≥ 7 days old | 7 d | 4 | 0.8 | |
| `reviews_w`, `comments_w`, `reactions_w` | social layer, when `FLAG_SOCIAL` is on | 7 d | 3 / 2 / 1 | 0.7 | new columns, 0 while the flag is off |
| `views`, `search_clicks` | anonymous, deduplicated per visitor and day, capped at 50 items/visitor/day (SDD 08 §7.1) | 7 d | 10·ln(1+views+2·clicks) | 0.5 | bots and prefetch excluded by the existing middleware |
| `external_delta` | new table `item_popularity_external(item_id, day, source, value)`: TMDB popularity, IGDB rating count and follows, Open Library edition/reader counts, Wikidata sitelinks; the signal is the 7-day relative change `ln(value_today / value_7d_ago)` clipped to [−1, 2] | 7 d | 8 | 0.6 | written by RFC-001's refresh job (display-only sources may feed a ranking signal; the value itself is never shown or exported) |
| `release_recency` | `items.release_date` | — | 3·max(0, 1 − days_since_release/90) | 1.0 | announced items with a date in the past 90 days; future dates count 0 |
| `sales_rank` | not available from a licensed source today | — | 0 | — | placeholder; Steam top sellers would be scraping |

Weighted contribution of a day `d` at age `a` (days ago): `D(a)=2^(−a/3.5)` (half-life 3.5 days for the 7-day home window; SDD 08 keeps 7 days for the 30-day API snapshot).

## 3. Score

```text
U_d = Σ_signals w·r·value_d                      (user and external signals of day d)
Q_d = 0.5 · 10·ln(1 + views_d + 2·clicks_d)       (anonymous, damped, reliability 0.5)
T_i = Σ_{a=0..6} D(a)·(U_d + Q_d) + release_recency_i
A_i = Σ_{a=0..6} (ratings_new_w + want_w + in_progress_w + completed_w + list_adds_w + reviews_w)   (distinct-account activity, unweighted by time)
```

Normalisation per category: `z_i = (T_i − median_type) / MAD_type` computed over items of the type with `A_i ≥ 1` on the day, so books and games compete only with their own kind and a slow category still produces a full carousel. Ranking key `(z desc, T desc, popularity_prior desc, id)`.

Worked example (golden test): a game with 40 views, 3 weighted new ratings and 2 wants today, 100 views and 4 completions 3 days ago, TMDB-like external delta +0.4 today: `U_0 = 6·3 + 3·2 + 8·0.6·0.4 = 25.9`, `Q_0 = 0.5·10·ln(41) = 18.6`, day 3 contribution `(5·4 + 0.5·10·ln(101))·2^(−3/3.5) = (20 + 23.1)·0.55 = 23.7`; `T = 68.2`, `A = 9`.

Burst example (manipulation): 300 views and 60 ratings from accounts younger than 7 days with no status on day 0: each rating weighs `0.175·0.25` (SDD 02 burst shadow), `ratings_new_w = 2.6`, `A = 2.6`, `T ≈ 15.6 + 28.5 = 44`; the item fails the entry rule below (A ≥ 5) and the z-score is modest; a `score_bursts` flag released by `ReleaseBursts` restores the weights only after review.

## 4. Entry, exit, curation

- **Eligible**: `visibility = 'public'`, `release_status = 'released'` (or `announced` with a date ≤ 90 days away), `quality_score ≥ 60` (RFC-001: cover, en synopsis, pt-BR and es titles, release date), not banned, not in `trending_exclusions`.
- **Enter**: `A_i ≥ 5` and `z_i ≥ 0.5` on the hourly run, or pinned.
- **Stay (cooldown)**: an item already in the carousel keeps its slot while `z_i ≥ 0` for up to 72 h after it last met the entry rule; this hysteresis stops hour-to-hour flicker. Rank changes inside the carousel are free.
- **Exit**: `z_i < 0` for 2 consecutive runs after the cooldown, or quality falls below the floor, or banned.
- **Diversity**: at most 2 items per franchise (`entities` role `franchise`/`series`) and at most 4 per primary genre in the 12; the next-ranked eligible item fills the slot. Ties by `popularity_prior`.
- **Fallback**: fewer than 12 entries → fill with `F_i = ln(1 + popularity_prior) + 3·max(0, 1 − days_since_release/180)` (SDD 08 §7.3 prior), labelled "Popular" in the UI (basis `prior`), as the home already labels quiet weeks.
- **Pin / ban**: table `trending_overrides(item_id, kind pinned|banned, until, reason, admin_id, created_at)`, admin page under `/admin/items` actions; pins take rank 1..n in pin order, bans are excluded everywhere (home, API snapshot, for-you).

## 5. Data and execution

- Snapshot table: reuse `trending_snapshots` with `window` (`7d` home, `30d` API), `rank` 1..12, `basis` (`trend`, `prior`, `pinned`), `score`, `z`, `entered_at`, `computed_at`. Unique `(window, item_type, rank, computed_at)`; readers take the latest `computed_at` per window and type.
- Job: `ComputeTrending` (hourly at :05, exists) computes both windows in one transaction per type, applies cooldown using the previous snapshot, writes the rows, then forgets `home:trending:v2:<locale>` (fixes T-09). Runtime target < 5 s at 1 M items: the query aggregates `item_activity_daily` for 7 days (≤ 7 rows per active item; a partial index `(item_type, day) WHERE day ≥ current_date − 7` keeps it small).
- Home: `app/controllers/home.go` reads the latest `7d` snapshot per type (one query, 48 rows) and caches 5 minutes; no computation on the request path. The listing payload gains `basis` and `rank` so the UI can show "#4 trending" (already) and "Popular".
- Retention: `PruneTrending` keeps 90 days of snapshots (exists at 04:45) so impact can be measured.

## 6. Tests

- Golden: the two worked examples above as table tests on `score.go`.
- Property: for any activity history, the score is monotone in each signal, bounded by the weights, and `z` is invariant to scaling the whole category's activity.
- Synthetic: a generator that emits 1,000 items × 30 days of activity with known "rising", "flat" and "burst" profiles; assert rising items enter within 24 h, bursts never enter, flat items exit within 72 h of losing signal, no franchise exceeds 2 slots, every slot meets the quality floor.
- Historical: replay the production `item_activity_daily` of the last 30 days (export to the staging database) and compare the 12 per day produced by v1 (most-rated) and v2; the report lists turnover per day (target 1-3 new items per type per day), seed-user share of the signal (target 0 after OQ-07), and how often the fallback fills slots.
- Performance: `go test -bench` of the snapshot query on a 1 M-item fixture (`HEXASCORE_BENCH=1`), budget 5 s; home TTFB unchanged (cached read).

## 7. Metrics to validate after launch

Snapshots retained 90 days plus `search_events`/activity give: carousel CTR per slot (clicks on `?from=trending` links / home views), rating conversion within 24 h of a trending click, median days an item stays, turnover per day, diversity (distinct franchises and genres per carousel), fallback share, seed-free share, and the number of pin/ban actions. Reported weekly by the intern in `docs/qa/trending-weekly.md`; thresholds to revisit weights after 30 days: CTR per slot below 0.5 % or turnover below 0.5 items/day for a category.

## 8. Backlog

| ID | Task | Est. (senior-days) | Acceptance |
| --- | --- | --- | --- |
| TR-01 | Score v2 (`score.go`): signals, decay 3.5 d, per-type z-score, release recency, seed exclusion; golden and property tests | 2 | tests above green; SDD 08 §7.2 updated |
| TR-02 | Entry/exit/cooldown/diversity in `Snapshot` with the `7d` window and `basis`; `trending_overrides` table and admin actions | 2 | synthetic suite green; admin pin shows at rank 1 within one run |
| TR-03 | `item_popularity_external` writer in RFC-001's refresh job (TMDB, IGDB, Open Library, Wikidata sitelinks) and the `external_delta` signal | 1.5 | delta computed on a fixture; 0 when a source is off |
| TR-04 | Home reads the snapshot (`home.go`), cache key v2, remove the dead `Forget`; UI label "Popular" for `prior` | 1 | home TTFB unchanged; `TestHome_TrendingFromSnapshot` |
| TR-05 (Intern) | Historical replay report on staging + weekly metrics doc | 1 | first report committed |
| TR-06 | SDD 08 and `/how-score-works` page section "How trending works" (transparent formula, pt-BR/es) | 1 | docs rendered in 3 locales |

Rollback: the home controller keeps the v1 query behind `FLAG_TRENDING_V2` (default on after TR-04 lands in staging); turning the flag off restores the most-rated list without a deploy.
