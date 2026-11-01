# IMPL-mobile-entity-paging Next control on the mobile entity groups

status: concluida
branch: agent/senior/IMPL-mobile-entity-paging @ a2febfa (from local main 6440d0a, which holds P5.B.2.5 and IMPL-mobile-entity-more; worktree `wt/impl-mobile-detail`)
files: `apps/mobile/src/screens/Entity.tsx`, `apps/mobile/src/screens/entity.test.tsx`, `apps/mobile/src/query/keys.ts`
acceptance:

- red: `bun run test -- src/screens/entity.test.tsx --runInBand --forceExit` → 1 failed ("Next follows meta.next_cursor and appends the group's next page"), 3 passed
- green: same command → 4 passed

validation: lint ok (`bun run lint`, 193 files), typecheck ok (`bun run typecheck`), tests ok (above); no device look
changelog: Added: each group on the mobile entity screen has a Next control that loads its next page.
roadmap: no box
risks / follow-ups:

- `EntityView` keeps one `useQueries` call over `(type, cursor)` slots: page one per type plus one slot per cursor the user followed. Next appends the last page's `meta.next_cursor` for that type; the rows of all pages of a type are concatenated. TanStack has no `useInfiniteQueries`, so this avoids one hook component per group.
- `queryKeys.entityItems` takes an optional `cursor` (part of the key object; prefix invalidation still matches every page).
- Next hides while the next page loads (no `next_cursor` yet), so a double tap cannot fetch twice; it hides for good on an empty `next_cursor`.
- The followed cursors are scoped to `id|locale`, so another entity or a locale change restarts at page one.
- The "Showing N of M" count uses `meta.total` from page one and the number of loaded rows.

## Review (especialista, 2026-11-01)

verdict: APROVADO
re-run: in batch `agent/especialista/mobile-batch-2` (main 35cd2062 + the six mobile branches): `bun run lint` (biome ci, 195 files) ok, `bun run typecheck` ok, `bun run test -- --runInBand --forceExit` → 56 suites, 373 tests passed
checked: sdd 10, authz, secrets, sql (n/a), i18n (no resources/lang change), scope, tests
notes: Per-type cursor list scoped by id and locale; cursor passed back opaque (VF-005); Next hidden while the next page loads, so no double request.
gate on main after merge: mobile lint/typecheck/jest ok (merge 736875d4); no Go or web TS touched
PRONTO_PARA_MERGE agent/especialista/mobile-batch-2 (merged 736875d4)
