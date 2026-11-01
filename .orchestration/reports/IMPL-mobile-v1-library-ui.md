# IMPL-mobile-v1-library-ui v1 library and lists screen (presentational)

status: concluida
branch: agent/senior/IMPL-mobile-v1-library-ui @ 7df8b4e (stacked on agent/senior/IMPL-mobile-v1-item-ui @ 5685d4c; stack rebased onto main 679005b)
files: `apps/mobile/src/ui/screens/LibraryPage.tsx`, `apps/mobile/src/ui/screens/LibraryPage.test.tsx`
acceptance:
- red run (test first): `bun run --cwd apps/mobile test -- src/ui/screens/LibraryPage` → suite failed, `Cannot find module './LibraryPage'` (0 tests)
- green: same command → 4 tests passed (kicker, header, "N titles · M rated", status caption + picker trigger, Filters; groups with header + count, rows with meta incl. "your score" and score badge, open callback; see-all link only when `total > rows.length`; empty state with Browse)
- `bun run --cwd apps/mobile lint` → biome ci clean (99 files); `bun run --cwd apps/mobile typecheck` → clean
validation: lint ok, typecheck ok, targeted test ok; no device run. Frames: `web-mobile__library__default.png`, `web-mobile-dark__library__default.png`
changelog: Added: v1 library screen component (status shelves, see all, empty state) for the Android app.
roadmap: no box
risks / follow-ups:
- Lists: no mobile lists frame exists; `LibraryGroup` is shape-neutral (key, title, total, first rows), so user lists render with the same component (one group per list). A dedicated list detail screen is out of scope until a frame exists.
- Status picker and Filters open `@expo/ui` sheets from `onStatus` / `onFilters` in the wiring row; i18n keys (`library.*`) are added there.

## Review (especialista, 2026-10-31)

verdict: APROVADO
re-run: batch branch agent/especialista/IMPL-mobile-batch (all 8 rows merged on main 5100631, conflicts only in navigation.test.tsx): `bun run --cwd apps/mobile catalogues` no diff, `lint` ok, `typecheck` ok, `test` 45 suites / 275 tests ok; `vitest packages/i18n` 5 ok; `go test ./tests/i18n` ok\nchecked: SDD 10, i18n parity (all labels through t(), nav namespace bundled), no hardcoded copy, no secrets, scope, tests\ngate on main after merge: mobile lint/typecheck/test ok (full local.sh at the end of the batch)\nPRONTO_PARA_MERGE agent/senior/IMPL-mobile-v1-library-ui
