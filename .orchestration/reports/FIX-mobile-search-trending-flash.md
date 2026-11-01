# FIX-mobile-search-trending-flash Loading state for search trending

status: concluida
branch: agent/senior/FIX-mobile-search-trending-flash @ e80eee1 (stacked on agent/senior/IMPL-mobile-search-trending @ ca3de58, itself on IMPL-mobile-search-i18n; worktree `wt/impl-mobile-search`)
files: `apps/mobile/src/screens/Search.tsx`, `apps/mobile/src/screens/search.test.tsx`, `apps/mobile/src/ui/screens/SearchSheet.tsx`, `apps/mobile/src/ui/screens/SearchSheet.test.tsx`
acceptance:

- red: `bun run test -- src/screens/search.test.tsx --runInBand --forceExit` → 1 failed ("while trending loads the sheet shows a loading state, not the empty message"), 8 passed
- green: `bun run test -- src/screens/search.test.tsx src/ui/screens/SearchSheet.test.tsx --runInBand --forceExit` → 16 passed

validation: lint ok (`bun run lint`, 164 files), typecheck ok (`bun run typecheck`), tests ok (above); no device look
changelog: Fixed: the mobile search sheet shows "Loading…" while trending loads instead of flashing "Nothing is trending this week yet."
roadmap: no box
risks / follow-ups:

- The `empty` sheet state gains an optional `trendingLoading`; `SearchBody` sets it when any trending query `isLoading`. With no trending rows the sheet shows the new `trendingLoading` label (a polite live region, like the search loading line); rows already cached show as before.
- `isLoading`, not `isPending`: an offline (paused) query with no cache is not fetching, so it falls back to the empty message instead of loading forever.
- The label reuses the existing `item.loading` key ("Loading…", bundled in en/pt-BR/es); no `resources/lang` change.
- Merge after IMPL-mobile-search-i18n and IMPL-mobile-search-trending.

## Review (especialista, 2026-11-01)

verdict: APROVADO
re-run: in batch `agent/especialista/mobile-batch-2` (main 35cd2062 + the six mobile branches): `bun run lint` (biome ci, 195 files) ok, `bun run typecheck` ok, `bun run test -- --runInBand --forceExit` → 56 suites, 373 tests passed
checked: sdd 10, authz, secrets, sql (n/a), i18n (no resources/lang change), scope, tests
notes: Loading line uses isLoading (offline paused query falls back to the empty text); reuses item.loading, polite live region.
gate on main after merge: mobile lint/typecheck/jest ok (merge 736875d4); no Go or web TS touched
PRONTO_PARA_MERGE agent/especialista/mobile-batch-2 (merged 736875d4)
