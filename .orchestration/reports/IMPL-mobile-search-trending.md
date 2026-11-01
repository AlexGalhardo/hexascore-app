# IMPL-mobile-search-trending Search trending from the app trending endpoints

status: concluida
branch: agent/senior/IMPL-mobile-search-trending @ ca3de58 (stacked on agent/senior/IMPL-mobile-search-i18n @ f0c1ed6; worktree `wt/impl-mobile-search`)
files: `apps/mobile/src/screens/Search.tsx`, `apps/mobile/src/screens/search.test.tsx`
acceptance:

- red: `bun run test -- src/screens/search.test.tsx --runInBand --forceExit` → 1 failed ("trending lists the top items across types by rank and a row opens the item"), 7 passed
- green: same command → 8 passed

validation: lint ok (`bun run lint`, biome ci, 164 files), typecheck ok (`bun run typecheck`), tests ok (above); no device look
changelog: Added: the mobile search screen shows the top trending items across games, movies, series and books before you type.
roadmap: no box
risks / follow-ups:

- `SearchBody` runs one `useQueries` over `/app/v1/{type}/trending` with `queryKeys.trending(type, locale)`, the same keys as the home screen, so the cached and persisted lists are reused (no extra request after home has loaded).
- Selection: all entries with an item, sorted by `rank` ascending (ties keep catalogue order games, movies, series, books), first 5. A failed or empty type is skipped; with nothing left the empty message from IMPL-mobile-search-i18n shows.
- `SearchReader` now also picks the four `list*Trending` methods; the route passes `appClient.read`, which has them. No backend change.
- Opening a trending row does not add an empty query to recent searches (only queries of 2+ characters are remembered).
- While the trending queries load, the empty message shows briefly (same as an empty list); no skeleton for that section.

## Review (especialista, 2026-11-01)

verdict: APROVADO
re-run: in batch `agent/especialista/mobile-batch-2` (main 35cd2062 + the six mobile branches): `bun run lint` (biome ci, 195 files) ok, `bun run typecheck` ok, `bun run test -- --runInBand --forceExit` → 56 suites, 373 tests passed
checked: sdd 10, authz, secrets, sql (n/a), i18n (no resources/lang change), scope, tests
notes: Reuses the home trending query keys (cache shared); top 5 by rank, failed types skipped; tapping a trending row does not store a recent query.
gate on main after merge: mobile lint/typecheck/jest ok (merge 736875d4); no Go or web TS touched
PRONTO_PARA_MERGE agent/especialista/mobile-batch-2 (merged 736875d4)
