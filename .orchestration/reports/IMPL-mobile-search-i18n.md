# IMPL-mobile-search-i18n Search i18n follow-ups (type label, recent removal, empty trending)

status: concluida
branch: agent/senior/IMPL-mobile-search-i18n @ f0c1ed6 (stacked on agent/senior/IMPL-mobile-v1-search-wire @ b9fe09d; worktree `wt/impl-mobile-search`)
files: `apps/mobile/scripts/copy-catalogues.ts`, `apps/mobile/src/i18n/catalogues/*.json`, `apps/mobile/src/screens/Search.tsx`, `apps/mobile/src/screens/search.test.tsx`, `apps/mobile/src/ui/screens/SearchSheet.tsx`, `apps/mobile/src/ui/screens/SearchSheet.test.tsx`, `resources/lang/{en,pt-BR,es}.json`
acceptance:

- red: `bun run test src/screens/search.test.tsx src/ui/screens/SearchSheet.test.tsx --runInBand --forceExit` → 4 failed (row type label "Games" instead of "Game", no "Remove “witcher” from recent searches" button, no trending empty state on the screen and in the sheet), 10 passed
- green: same command → 14 passed; with `src/i18n` → 21 passed

validation: lint ok (`bunx biome ci src scripts` in apps/mobile), typecheck ok (`tsc --noEmit`), tests ok (above); web `bunx vitest run resources/js/components/search/SearchPalette.test.tsx` 5 passed (shares `search.palette`); `bun run --cwd apps/mobile catalogues` → 576 keys per locale; no device look
changelog: Fixed: mobile search rows name the item type (Game, Movie…) instead of the section, the remove-recent button has its own label, and an empty trending list says so.
roadmap: no box
risks / follow-ups:

- `itemTypes` is now a bundled namespace (`BUNDLED_NAMESPACES`); the catalogues gained `itemTypes.{game,movie,series,book}` plus the two new keys.
- New strings (additive, all three locales): `search.palette.removeRecent` ("Remove “{q}” from recent searches") and `search.palette.trendingEmpty` ("Nothing is trending this week yet."). pt-BR and es written by me; a native check is welcome.
- `SearchLabels.trendingEmpty` is required; the trending heading now always shows in the empty state, with the message when there are no rows. The wiring still passes `trending: []` (no trending source on the search screen yet), so the message is what users see until that lands.
- Merge after IMPL-mobile-v1-search-wire. `resources/lang` overlaps P5.B.4.3 (`account.*`) and P7.A.1.3 (`social.feed.*`) only by file; keys are disjoint. Recopy catalogues after merging.

## Review (especialista, 2026-11-01)

verdict: CORRIGIDO_E_APROVADO
reviewed: batch agent/especialista/mobile-batch (main + status-ui + account-ui + search-i18n)
checked: mobile tsc, biome, jest `bun run test -- --runInBand --forceExit` 55 suites / 362 tests; web vitest 45/165; `go test ./app/i18n`; resources/lang change additive (search.palette keys) in all three locales.
Merge fix: `copy-catalogues.ts` BUNDLED_NAMESPACES keeps both `entity` (main) and `itemTypes`; catalogues recopied (640 keys per locale).

PRONTO_PARA_MERGE agent/especialista/mobile-batch
