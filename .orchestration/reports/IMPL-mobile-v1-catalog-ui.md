# IMPL-mobile-v1-catalog-ui v1 catalogue page (presentational)

status: concluida
branch: agent/senior/IMPL-mobile-v1-catalog-ui @ b40410b (stacked on agent/senior/IMPL-mobile-v1-home-ui @ 8cf2899)
files: `apps/mobile/src/ui/screens/CatalogPage.tsx`, `apps/mobile/src/ui/screens/CatalogPage.test.tsx`
acceptance:
- `bun run --cwd apps/mobile test -- src/ui/screens/CatalogPage` → 5 tests passed (title header, count, rows with "year · byline · N ratings" and score, row opens; pager top + bottom, Previous disabled on page 1, Next disabled on last page, page callbacks; filters toggle label with count, sort trigger, open panel replaces the list, `expanded` state, Apply/Reset; empty state with Reset and no pager)
validation: lint ok (biome ci), typecheck ok, tests ok; no device run. Frames: `web-mobile__catalog-index__{default,filters-open,empty,last-page}.png`
changelog: Added: v1 catalogue listing component (filters, sort, pager, rows) for the Android app.
roadmap: no box
risks / follow-ups:
- `filterPanel` is a slot: the wiring row (catalogue wiring, after P5.B.1.2) fills it with `@expo/ui` pickers for genre/platform/year/minimum score and opens the sort choice in an `@expo/ui` bottom sheet or picker from `onSort`. Keys needed: filters, close filters, sort, found, previous, next, page N of M, apply, reset, empty (check `catalog.*` first).
- Pager reuses `Btn` secondary (no chevrons, unlike the web frame) to keep one button primitive; rows render in a ScrollView (one API page, not virtualised).
- Cold Jest cache can time out the first test of the file; second run is clean.

## Review (especialista, 2026-10-31)

verdict: APROVADO
re-run: batch branch agent/especialista/IMPL-mobile-batch (all 8 rows merged on main 5100631, conflicts only in navigation.test.tsx): `bun run --cwd apps/mobile catalogues` no diff, `lint` ok, `typecheck` ok, `test` 45 suites / 275 tests ok; `vitest packages/i18n` 5 ok; `go test ./tests/i18n` ok\nchecked: SDD 10, i18n parity (all labels through t(), nav namespace bundled), no hardcoded copy, no secrets, scope, tests\ngate on main after merge: mobile lint/typecheck/test ok (full local.sh at the end of the batch)\nPRONTO_PARA_MERGE agent/senior/IMPL-mobile-v1-catalog-ui
