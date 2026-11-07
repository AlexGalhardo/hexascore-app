# Trending historical replay (TR-05)

Procedure for the historical replay of RFC-002 §6 ("Historical"), and the report it produces. Source: `docs/rfcs/RFC-002-trending.md`. The first run waits for TR-02 (`Snapshot` with the `7d` window and `basis`) to be merged and built locally. Until then every number below is **pending a run**.

## Scope and limits

- The replay compares the 12 items per day that v1 (most-rated) shows with what v2 would show, using the last 30 days of activity.
- It runs on a local scratch database only. Production is not read by this procedure. Exporting production data needs the owner's yes and is not part of this document.
- Seed users (`users.is_seed`) are excluded from every signal (OQ-07). The replay reports the seed share of the signal to confirm it is 0.
- No external API is called (TMDB, IGDB and the others are not used by the replay).
- Behaviour behind `FLAG_TRENDING_V2` (default off). The replay itself changes no production data.

## Inputs

| Input | Source | Check |
| --- | --- | --- |
| `item_activity_daily` (30 days) | table created by `database/migrations/20261031150000_create_trending_tables.go`; columns `item_id`, `day`, `item_type`, `views`, `search_clicks`, `ratings_new`, `ratings_updated`, `status_*`, `list_adds`, `suggestions` and their `*_w` weights | row count per day equals the number of active items that day |
| `trending_snapshots` | same migration; columns `computed_at`, `item_type`, `rank` (1 to 10), `item_id`, `trend_score`, `basis` (`trend` or `prior`) | v1 rows are kept; v2 rows are written with the replay's `computed_at` |
| v1 list | the most-rated query used by `app/controllers/home.go` (read the query there; do not copy it into this file) | 12 items per day, the same selection the home page makes |
| Seed flags | `users.is_seed` | every signal row joins back to non-seed users |

## Procedure

1. **Restore the window.** On a local scratch database, load the last 30 days of `item_activity_daily` from the export. Confirm `day` covers 30 consecutive dates and that no `item_id` is missing from `items`.
2. **Run v1 per day.** For each `day` in the window, record the v1 top 12 per `item_type`: item ids and rank.
3. **Run v2 per day.** Call the v2 snapshot job (TR-02) with the replay clock set to each day in turn, so that each day's `computed_at` matches that day. Record the 10 ranked items per type (`trending_snapshots`, `basis` column).
4. **Compare.** For each day and type, compute:
   - turnover: items in today's v2 list that were not in yesterday's v2 list (target 1 to 3 new items per type per day, RFC-002 §6);
   - overlap with v1 for the same day;
   - seed share: the share of the signal in `item_activity_daily` joined to `users.is_seed = true` (target 0, OQ-07);
   - fallback share: the number of slots filled with `basis = 'prior'` (fallback) per day.
5. **Write the report** below, with the day-level table in `docs/qa/trending-replay-<YYYY-MM-DD>.md` (new file per run; not committed until the numbers exist).

## Report template

| Metric | Target | Result (local) |
| --- | --- | --- |
| Turnover per type per day | 1 to 3 new items | pending a run |
| Seed share of the signal | 0 | pending a run |
| Fallback slots per day | recorded | pending a run |
| Overlap with v1 (per day, mean) | recorded | pending a run |
| Days with fewer than 10 v2 items | 0 | pending a run |

## Weekly metrics

Weekly metrics (RFC-002 §7: CTR per slot, conversion within 24 h, median days an item stays, turnover, diversity, fallback share, seed-free share, pin and ban actions) go to `docs/qa/trending-weekly.md` once there is data. This file defines only the replay.

## Open points

- The v1 query is not copied here on purpose: it lives in `app/controllers/home.go` and the replay must read it at run time.
- The export method for the local scratch database (a dump of the last 30 days) needs the owner's approval for any production read; until then the first run uses local data only.
