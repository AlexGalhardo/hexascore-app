---
name: hexascore-ratings
description: Work on Hexascore ratings and the denormalized item stats (lib/ratings.ts, actions/ratings.ts, recomputeItemStats, item merge, seed stats). Use when changing how ratings are written, deleted, merged or aggregated, or when touching items.rating_count / score / score_c1..c6.
---

# Hexascore ratings

## Write path

`actions/ratings.ts` (`saveRating`, `deleteRating`) → `lib/ratings.ts` (`writeRating`, `removeRating`):

1. zod first (`ratingInputSchema`: uuid + exactly 6 integers 0–100).
2. `asMember(await getSessionUser())` (`lib/member.ts`): session, not banned, email verified, nickname set. Identity never comes from the input.
3. `hitRateLimit(db, "rating:<userId>", 30, 60s)`.
4. One transaction: `lockItem` (`SELECT id FROM items WHERE id = $1 FOR UPDATE`) → upsert/delete the row in `ratings` → `recomputeItemStats(tx, itemId)`.
5. After commit: `updateTag("item:<id>")` (read-your-own-writes for the rater) and `revalidateTag("catalog", "max")` (listings refresh in the background).

## Invariants

- Stats are **recomputed from `ratings`** (`count`, `avg(overall)`, `avg(c1..c6)`, rounded half away from zero), never incremented. Any code that changes `ratings` for an item (merge, imports, purges) must call `recomputeItemStats` in the same transaction, after locking the item.
- `items.score*` store the raw rounded averages for **any** count (null only at 0 ratings). The threshold `MIN_RATINGS_FOR_SCORE` (5) is applied **only on read** (`lib/catalog/repository.ts` → `toSummary`, `visible`). Do not null scores on write.
- `ratings.overall` is a generated column (`round(sum/6.0)`): never insert it. Drizzle's `insert().select()` rejects tables with generated columns — `mergeItems` uses raw SQL for that copy.
- Deleted accounts are anonymized (`lib/account-purge.ts`), not deleted: their ratings keep counting. `user` row deletion would cascade ratings — never hard-delete users.
- Lock order when touching several items: by id ascending (`mergeItems` does `ORDER BY id FOR UPDATE`).
- Time to beat (`lib/time-to-beat.ts`) has no denormalized stats: medians are computed on read inside `getItemById` (cached under `item:<id>`), so writes only `updateTag(item:<id>)`. Reports exist only while the user's status is `completed`; `setItemStatus` deletes them in the same transaction and returns `"cleared"` so the action refreshes the tag.
- The JS `overall()` in `lib/categories.ts` mirrors the SQL rounding for the live preview; keep them identical.

## Tests

`lib/ratings.test.ts` (PGlite): schema limits, insert/edit/delete, unknown item, 10 concurrent writers, threshold 4→5 through `getItemById`, anonymized account still counted. `lib/social/social.test.ts` covers merge stats. PGlite serializes transactions, so the concurrency test proves correctness of the recompute, not lock contention; real Postgres relies on `FOR UPDATE`.

## Local E2E gotchas

- PGlite data dirs under OneDrive (or a killed `next dev`) get corrupted (`RuntimeError: Aborted()`): use `PGLITE_DIR` outside synced folders, or a throwaway Postgres container with `DATABASE_URL`.
- Changing the DB behind a running/previous dev server leaves stale `unstable_cache` entries in `apps/web/.next/dev/cache` — delete it after out-of-band DB edits.
- Server Actions are optimistic in the UI: wait for the button to re-enable before reloading in Playwright, or the request may be cut.
