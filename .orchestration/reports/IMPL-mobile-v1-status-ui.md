# IMPL-mobile-v1-status-ui v1 look for the status and time-to-beat sheets

status: concluida
branch: agent/senior/IMPL-mobile-v1-status-ui @ ecbc46d (stacked on agent/senior/P5.B.2.4 @ ed792be; worktree `wt/impl-mobile-item`)
files: `apps/mobile/src/status/StatusSheet.tsx`, `apps/mobile/src/status/TimeToBeatSheet.tsx`, `apps/mobile/src/status/status.test.tsx`
acceptance:

- red: `bun run test src/status/status.test.tsx --runInBand --forceExit -t "v1 look"` → 2 failed (no `header` title, no check mark on the checked status), 1 passed
- green: `bun run test src/status/status.test.tsx --runInBand --forceExit` → 14 passed, 0 failed

validation: lint ok (`bunx biome ci src/status`), typecheck ok (`tsc --noEmit`), tests ok (above); no device look
changelog: Changed: mobile status and time-to-beat sheets follow the v1 rating-sheet look; the chosen status shows a check mark.
roadmap: no box
risks / follow-ups:

- Orchestrator decision: no separate frame, so both sheets reuse the item screen's rating panel (`border-t-2 border-primary bg-surface p-4`, display-font H2 title with `accessibilityRole="header"`). The classes are exported as `SHEET` / `SHEET_TITLE` from `StatusSheet.tsx`.
- Status options are ledge buttons (`rounded-sm border border-b-2`, sink on press); the checked option keeps the inverted fill and adds a `checkmark` glyph (`testID="status-checked"`), so colour is not the only signal.
- Each radio now has an explicit `accessibilityLabel` (the status label): the Ionicons glyph renders as a `Text` and would otherwise join the computed name.
- testIDs `status-sheet`, `time-to-beat-sheet` and the radio roles are kept. Behaviour (queue writes, validation) unchanged.
- Merge after P5.B.2.4.

## Review (especialista, 2026-11-01)

verdict: APROVADO
reviewed: batch agent/especialista/mobile-batch (main + status-ui + account-ui + search-i18n)
checked: mobile tsc, biome, jest `bun run test -- --runInBand --forceExit` 55 suites / 362 tests; web vitest 45/165; `go test ./app/i18n`; resources/lang change additive (search.palette keys) in all three locales.


PRONTO_PARA_MERGE agent/especialista/mobile-batch
