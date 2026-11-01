# FIX-mobile-item-add-to-list "Add to list" on the mobile item screen

status: concluida
branch: agent/senior/FIX-mobile-item-add-to-list @ 115851f (from local main 6440d0a, which holds P5.B.4.1 and IMPL-mobile-v1-item; worktree `wt/impl-mobile-search`)
files: `apps/mobile/src/screens/item/Item.tsx`, `apps/mobile/src/screens/item/item.test.tsx`, `apps/mobile/app/[type]/[slug].tsx`
acceptance:

- red: `bun run test -- src/screens/item/item.test.tsx --runInBand --forceExit` → 1 failed ("Add to list opens the add-to-list sheet with the item's own list type"), 10 passed
- green: `bun run test -- src/screens/item src/status --runInBand --forceExit` → 22 passed; `src/navigation.test.tsx` → 22 passed

validation: lint ok (`bun run lint`, 193 files), typecheck ok (`bun run typecheck`), tests ok (above); no device look
changelog: Added: the mobile item screen has an "Add to list" button that opens the add-to-list sheet for the item's type.
roadmap: no box
risks / follow-ups:

- Board check first: the row was `todo` with no owner, and senior-a had not wired it on main.
- `ItemView` takes an optional `lists?: ListsDeps`; the route passes `listsDeps` from `auth/runtime`. Without it there is no button (keeps `item.test.tsx` and `status.test.tsx` free of the runtime). The button is the existing `ItemDetail` `onAddToList` slot (label `lists.addToList`).
- The sheet opens in the same `Modal` as the status sheets (`sheet === "list"`), with `type = SINGULAR[type]` (cast to `ListType`; `SINGULAR` is typed `Record<CatalogType, string>`), and closes on `onDone` after an add or a create.
- Independent of IMPL-mobile-v1-lists-ui (both branch from main 6440d0a, no shared files); merge in either order.

## Review (especialista, 2026-11-01)

verdict: APROVADO
re-run: in batch `agent/especialista/mobile-batch-2` (main 35cd2062 + the six mobile branches): `bun run lint` (biome ci, 195 files) ok, `bun run typecheck` ok, `bun run test -- --runInBand --forceExit` → 56 suites, 373 tests passed
checked: sdd 10, authz, secrets, sql (n/a), i18n (no resources/lang change), scope, tests
notes: Button only when lists deps are passed; list type is SINGULAR of the route type; sheet reuses AddToListSheet (queued add).
gate on main after merge: mobile lint/typecheck/jest ok (merge 736875d4); no Go or web TS touched
PRONTO_PARA_MERGE agent/especialista/mobile-batch-2 (merged 736875d4)
