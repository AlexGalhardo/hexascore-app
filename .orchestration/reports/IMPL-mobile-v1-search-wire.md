# IMPL-mobile-v1-search-wire Wire the search tab to the app search API

status: concluida
branch: agent/senior/IMPL-mobile-v1-search-wire @ b9fe09d (from main 378a569, main ddcfe2c+ merged in 9404220; worktree `wt/impl-mobile-search`)
files: `apps/mobile/app/(tabs)/search.tsx`, `apps/mobile/src/screens/Search.tsx` (new), `apps/mobile/src/screens/search.test.tsx` (new), `apps/mobile/src/ui/screens/SearchSheet.tsx`, `apps/mobile/src/ui/screens/SearchSheet.test.tsx`, `apps/mobile/src/navigation.test.tsx` (search route wrapped in `offline()`, asserts the "All" chip)
acceptance:

- red: `bun run test --runInBand src/screens/search.test.tsx src/ui/screens/SearchSheet.test.tsx` → search suite "Cannot find module './Search'", 2 sheet tests failed (limit, filters/error), 4 passed
- green: same command → 12 passed, 0 failed; `src/navigation.test.tsx` → 20 passed

validation: lint ok (`bun run --cwd apps/mobile lint`, 164 files), typecheck ok (`tsc --noEmit`), tests ok (above); no browser/device look (no emulator on this host)
changelog: Added: the Android search tab searches the catalogue (type chips, see all, paged results).
roadmap: no box
risks / follow-ups:

- No senior-a P5.B.3.1 search hooks exist on the board or on `main`; `SearchBody` uses `useInfiniteQuery` directly (key `["searchIndex", {q, type, locale}]`, page param 1-based, next page while loaded < `total`).
- Debounce 300 ms, nothing sent under 2 characters (server `tooShort` floor). Errors show `search.unavailable`.
- Rows show `year` and `score` from SearchHit (FIX-search-hit-meta, merged into the branch with main; b9fe09d, test asserts both and the "–" case).
- Shortcuts (commented in code): recent searches are session-only; row type label uses `catalog.title.*` (section name) because `itemTypes` is not bundled; "remove recent" label is `lists.remove: <q>` because `resources/lang` and the catalogues are locked by P5.B.4.3 (doing). Follow-up: add `search.palette.removeRecent` and bundle `itemTypes` after P5.B.4.3 merges.
- Empty state has no trending rows (`trending: []`); a cross-type trending feed does not exist in the app API.
- `navigation.test.tsx`: tests pass but Jest does not exit afterwards on this host (also with only the home route selected, so it predates this change); run it with `timeout` or `--forceExit`.

## Review (especialista, 2026-11-01)

verdict: CORRIGIDO_E_APROVADO
re-run: batch tip with main 8b3b0e9: `bunx tsc --noEmit` ok, `bunx biome ci src app` ok, `bun run --cwd apps/mobile test -- --runInBand --forceExit` 55 suites / 352 tests ok
checked: SDD 07 (2-char floor, debounce), authz (read-only public search), i18n (existing keys only; shortcuts noted for a removeRecent key), scope, tests
fixes: 1. navigation.test conflict with the P5.B batch resolved (search offline/"All", lists/profile from main)
gate on main after merge: targeted only (section 8.2); full gate once at the end of the batch
PRONTO_PARA_MERGE agent/especialista/IMPL-mobile-batch
