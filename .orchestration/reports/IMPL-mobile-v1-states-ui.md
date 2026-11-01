# IMPL-mobile-v1-states-ui v1 empty, error and blocked states

status: concluida
branch: agent/senior/IMPL-mobile-v1-states-ui @ 16b312c (stacked on agent/senior/IMPL-mobile-v1-auth-ui @ 759e80f)
files: `apps/mobile/src/ui/screens/StatePage.tsx`, `apps/mobile/src/ui/screens/StatePage.test.tsx`, `apps/mobile/src/account/Blocked.tsx`, `apps/mobile/app/update-required.tsx`
acceptance:
- red run (test first): `bun run --cwd apps/mobile test -- src/ui/screens/StatePage` → suite failed, `Cannot find module './StatePage'` (0 tests)
- green: same command → 4 passed (code + header title + body + ordered actions; no code → no number, live region polite; code hidden from accessibility so the title is read once; `role: "link"` action)
- `bun run --cwd apps/mobile test -- src/account src/ui/screens/StatePage src/navigation` → 26 passed (existing banned / deletion-scheduled tests unchanged and green)
- `bunx biome check` on the 4 files → clean; `bun run --cwd apps/mobile typecheck` → clean
validation: lint ok, typecheck ok, targeted tests ok; no device run. Frames: `web-mobile__error__404.png`, `web-mobile__error__500.png` (+ dark)
changelog: Changed: the Android banned, deletion-scheduled and update-required screens use the v1 state layout; Added: a shared v1 state page for empty and error states.
roadmap: no box
risks / follow-ups:
- `StatePage` is generic (optional display code, title, body lines, actions). The 404/500/offline/empty uses are wired by the screen rows that own them (errors.notFoundTitle/Body, errors.serverTitle/Body, errors.retry already exist in the catalogues).
- Token gaps: the frame draws the code at 120 px and the title at 32 px; the type scale stops at 36, so the code uses `text-36` and the title `text-30`. Token proposal (display size 120) for the designer; `docs/design-handoff.md` does not exist on this branch.
- Update required keeps its current copy: the store link label is the store URL and logout uses `account.sessions.revoke` (no `account.updateRequired.*` action keys exist); new keys would be an i18n row.

## Review (especialista, 2026-10-31)

verdict: APROVADO
re-run: batch branch agent/especialista/IMPL-mobile-batch (all 8 rows merged on main 5100631, conflicts only in navigation.test.tsx): `bun run --cwd apps/mobile catalogues` no diff, `lint` ok, `typecheck` ok, `test` 45 suites / 275 tests ok; `vitest packages/i18n` 5 ok; `go test ./tests/i18n` ok\nchecked: SDD 10, i18n parity (all labels through t(), nav namespace bundled), no hardcoded copy, no secrets, scope, tests\ngate on main after merge: mobile lint/typecheck/test ok (full local.sh at the end of the batch)\nPRONTO_PARA_MERGE agent/senior/IMPL-mobile-v1-states-ui
