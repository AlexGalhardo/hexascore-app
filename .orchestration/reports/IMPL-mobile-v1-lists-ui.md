# IMPL-mobile-v1-lists-ui v1 look for the lists tab and the add-to-list sheet

status: concluida
branch: agent/senior/IMPL-mobile-v1-lists-ui @ 691546c (from local main 6440d0a, which already holds P5.B.4.1 85226e5; worktree `wt/impl-mobile-search`)
files: `apps/mobile/src/lists/ListsTab.tsx`, `apps/mobile/src/lists/AddToListSheet.tsx`, `apps/mobile/src/lists/lists.test.tsx`
acceptance:

- red: `bun run test -- src/lists/lists.test.tsx --runInBand --forceExit` → 2 failed (row accessible name "Backlog, Games, 2 items, Public" missing, no `add-to-list-sheet` panel), 6 passed
- green: `bun run test -- src/lists --runInBand --forceExit` → 8 passed

validation: lint ok (`bun run lint`, 193 files), typecheck ok (`bun run typecheck`), tests ok (above); no device look
changelog: Changed: the mobile lists tab and the add-to-list sheet use the v1 look (ruled rows with type, count and visibility, chip type picker, sheet panel).
roadmap: no box
risks / follow-ups:

- Lists tab: ruled rows (2 px top rule) with title, a muted meta line "type · count · visibility", a chevron, and an explicit accessible name joining them; "New list" sits in the sheet panel look (`border-t-2 border-primary bg-surface`, display 24 title); the type picker is a chip row (`accessibilityState.selected`) instead of primary/secondary buttons.
- Add-to-list sheet: same panel (`testID="add-to-list-sheet"`), ruled rows with title, count and an add glyph, accessible name "title, count".
- Kept the names the tests and Maestro use: `Create list`, `Title`, rows as buttons; the row's visible title text is unchanged, so `.maestro/smoke/add-to-list.yaml` (`tapOn: "Maestro smoke"`) should still match, not run here (owner-gated).
- The panel classes duplicate `SHEET` from `src/status/StatusSheet.tsx` on the unmerged IMPL-mobile-v1-status-ui branch; once it merges, both can import it.
- No new strings, no `resources/lang` change.

## Review (especialista, 2026-11-01)

verdict: APROVADO
re-run: in batch `agent/especialista/mobile-batch-2` (main 35cd2062 + the six mobile branches): `bun run lint` (biome ci, 195 files) ok, `bun run typecheck` ok, `bun run test -- --runInBand --forceExit` → 56 suites, 373 tests passed
checked: sdd 10, authz, secrets, sql (n/a), i18n (no resources/lang change), scope, tests
notes: Styling only; rows keep role button with a full accessible label; type chips expose accessibilityState.selected; no new strings.
gate on main after merge: mobile lint/typecheck/jest ok (merge 736875d4); no Go or web TS touched
PRONTO_PARA_MERGE agent/especialista/mobile-batch-2 (merged 736875d4)
