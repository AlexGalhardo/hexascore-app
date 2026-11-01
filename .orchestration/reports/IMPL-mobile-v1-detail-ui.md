# IMPL-mobile-v1-detail-ui v1 entity, list detail, notifications and editions (presentational)

status: concluida
branch: agent/senior/IMPL-mobile-v1-detail-ui @ 05e50d0 (stacked on agent/senior/IMPL-mobile-v1-states-ui @ 16b312c)
files: `apps/mobile/src/ui/ItemRow.tsx`, `apps/mobile/src/ui/screens/EntityPage.tsx`, `ListDetailPage.tsx`, `NotificationsPage.tsx`, `EditionsList.tsx` and one `*.test.tsx` each
acceptance:
- red run (test first): `bun run --cwd apps/mobile test -- src/ui/screens/EntityPage src/ui/screens/ListDetailPage src/ui/screens/NotificationsPage src/ui/screens/EditionsList` → 4 suites failed, `Cannot find module` (0 tests)
- green: same command → 10 passed
  - Entity: kind kicker, name header, role chips (`selected` state, onRole), facts, description + licence + source link, groups with header + "n of m" + note, empty groups hidden (SDD 10 §8), empty state with browse action
  - List detail: header, rows open, up/down reorder (first/last disabled), single remove; offline disables rename, delete and reorder and shows the offline note while remove stays enabled (queued, P5.B.4.1); non-owner sees no controls; empty text
  - Notifications: unread rows carry a visible "Unread" tag and it leads the accessible name (colour not the only signal), unread count, mark all (disabled with none unread), empty text
  - Editions: header + fact rows per edition; renders nothing without editions
- `bun run --cwd apps/mobile lint` → biome ci clean (139 files); `bun run --cwd apps/mobile typecheck` → clean
validation: lint ok, typecheck ok, targeted tests ok; no device run. Frames: `web-mobile__entity__default.png`, `web-mobile__entity__empty.png`, `web-mobile__edition__default.png` (+ dark)
changelog: Added: v1 entity, list detail, notifications and book editions components for the Android app.
roadmap: no box
risks / follow-ups:
- No public profile screen: SDD 10 §6.1 lists none for the app, so none was built. No edition route either: editions are a block for the book item screen (`/books/{id}/editions`), placed during item wiring.
- No v1 frames exist for list detail or notifications (account group designs are `DESIGN-v1-editorial-web-account`, todo); both follow the library/entity language (kicker, serif header, 2 px rule, hairline rows). Revisit when those frames land.
- `ItemRow` is new and shared by entity and list detail; `LibraryPage` keeps its own row (review-pending row, not touched). Folding it onto `ItemRow` is a small later cleanup.
- Reorder is up/down buttons, not drag (44 px targets, works with TalkBack); i18n keys come with the wiring rows (P5.B.2.5, P5.B.4.1, P5.B.4.3).
- `w-24` (96 px) for the fact label column vs 104 px in the frame (spacing scale).

## Review (especialista, 2026-10-31)

verdict: APROVADO
re-run: batch branch agent/especialista/IMPL-mobile-batch (all 8 rows merged on main 5100631, conflicts only in navigation.test.tsx): `bun run --cwd apps/mobile catalogues` no diff, `lint` ok, `typecheck` ok, `test` 45 suites / 275 tests ok; `vitest packages/i18n` 5 ok; `go test ./tests/i18n` ok\nchecked: SDD 10, i18n parity (all labels through t(), nav namespace bundled), no hardcoded copy, no secrets, scope, tests\ngate on main after merge: mobile lint/typecheck/test ok (full local.sh at the end of the batch)\nPRONTO_PARA_MERGE agent/senior/IMPL-mobile-v1-detail-ui
