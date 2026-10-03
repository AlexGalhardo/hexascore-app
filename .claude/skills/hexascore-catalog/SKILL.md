---
name: hexascore-catalog
description: Work on Hexascore catalog data access, caching and PGlite tests (lib/catalog, cached reads, revalidation tags, seed). Use when adding a catalog query, a page that reads items, a cache tag, or a repository test.
---

# Hexascore catalog

## Layers

1. `apps/web/lib/catalog/repository.ts` — pure functions `fn(db, params)` returning JSON-only DTOs (dates as ISO strings). No Next imports here: the public API (0.4) and tests reuse them.
2. `apps/web/lib/catalog/cached.ts` — the only module pages/route handlers import. Wraps the repository with `unstable_cache` + tags `catalog` and `item:<id>`.
3. Pages / route handlers.

## Caching in this Next 16 setup (gotchas)

- `"use cache"` / `cacheTag` / `cacheLife` need `cacheComponents: true` (see `node_modules/next/dist/docs/01-app/03-api-reference/01-directives/use-cache.md`). The app does **not** enable it, so use `unstable_cache` (docs: `02-guides/caching-without-cache-components.md`).
- `unstable_cache` serializes results as JSON: never return `Date`, `Map` or class instances. Errors are not cached; check them by `error.name` (e.g. `"InvalidCursorError"`), not only `instanceof`.
- Item-scoped entries: create the wrapper per call so the tag can include the id — see `forItem()` in `cached.ts`.
- `revalidateTag(tag, "max")` (two-arg form; the one-arg form is deprecated). Ingestion calls `POST /api/cron/revalidate` with `{"tags": ["catalog" | "item:<uuid>"]}` and `Authorization: Bearer $CRON_SECRET`.
- Pages that hit the DB without dynamic inputs (landing) call `await connection()` so `next build` never queries the database.

## PGlite

- Tests: `const { db, close } = await createTestDatabase()` from `lib/db/testing.ts` (in-memory, all migrations). Never import `lib/db/client.ts` in tests.
- `db.execute()` returns `{ rows }` on PGlite but an array on postgres-js — use the query builder so code works on both.
- Column names follow drizzle `casing: "snake_case"`: `scoreC1` → `score_c_1`. Reference columns through the schema, never hand-written names.
- `SELECT DISTINCT` + `ORDER BY <expr>` fails because drizzle re-parameterizes the expression; order by ordinal (`sql\`1\``).
- Keyset on timestamps: carry `created_at::text` in the cursor (µs precision), cast back with `::timestamptz`.
- Only one process can open `apps/web/.pglite`: stop `bun run dev` before `bun run db:seed`/`db:migrate`.

## Seed

`bun run db:seed` (defaults 50 users / 400 items). Flags `--users --items --batch` (run `bun lib/db/seed.ts ...` inside `apps/web`). Records are derived from `(type, index)` only, so raising `--items` never changes existing rows; upserts make reruns idempotent. Stats (`rating_count`, `score`, `score_c*`) are recomputed with one `UPDATE ... FROM` at the end.
