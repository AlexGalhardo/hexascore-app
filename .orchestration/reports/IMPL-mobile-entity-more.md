# IMPL-mobile-entity-more Per-group "next" control on EntityPage

status: concluida
branch: agent/senior/IMPL-mobile-entity-more @ c5d3655 (from main ddcfe2c; worktree `wt/impl-mobile-detail`)
files: `apps/mobile/src/ui/screens/EntityPage.tsx`, `apps/mobile/src/ui/screens/EntityPage.test.tsx`
acceptance:

- red: `bun run test --runInBand src/ui/screens/EntityPage.test.tsx` → "a group with more rows shows a Next control" failed (no button), 4 passed
- green: same command → 5 passed, 0 failed

validation: lint ok (`bun run --cwd apps/mobile lint`), typecheck ok (`tsc --noEmit`), tests ok (above); no device look
changelog: Added: mobile entity page shows a "Next" control per item group when more items exist.
roadmap: no box
risks / follow-ups:

- API: `EntityGroup.hasMore?: boolean`, `EntityPageProps.moreLabel?: string` and `onMore?: (groupKey) => void`. The control renders only when both `hasMore` and `onMore` are set, so the current wiring (P5.B.2.5, not on `main` yet) is unchanged until it passes them.
- Several groups can show "Next"; each button carries `accessibilityHint` = the group title so screen readers can tell them apart.
- Wiring (P5.B.2.5 author): `moreLabel={t(locale, "entity.next")}`, `hasMore` from `meta.next_cursor`, `onMore` calls `fetchNextPage` of that group's infinite query.

## Review (especialista, 2026-11-01)

verdict: APROVADO
re-run: batch tip with main 8b3b0e9: `bunx tsc --noEmit` ok, `bunx biome ci src app` ok, `bun run --cwd apps/mobile test -- --runInBand --forceExit` 55 suites / 352 tests ok
checked: presentational only (optional props), a11y hint names the group, tests
fixes: none
gate on main after merge: targeted only (section 8.2); full gate once at the end of the batch
PRONTO_PARA_MERGE agent/especialista/IMPL-mobile-batch
