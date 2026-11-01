# IMPL-mobile-v1-account-ui Profile tab unread badge and v1 look for account extras

status: concluida
branch: agent/senior/IMPL-mobile-v1-account-ui @ d650888 (stacked on agent/senior/P5.B.4.3 @ 5bf654e; worktree `wt/impl-mobile-auth`)
files: `apps/mobile/app/(tabs)/_layout.tsx`, `apps/mobile/src/profile/{About,AccountExtras,DeleteAccount}.tsx`, `apps/mobile/src/profile/account.test.tsx`, `apps/mobile/src/ui/tabBarTheme.ts`, `apps/mobile/src/ui/shell.test.tsx`, `apps/mobile/src/navigation.test.tsx` (smoke test wraps the tabs layout in a QueryClient)
acceptance:

- red: `bun run test src/profile/account.test.tsx src/ui/shell.test.tsx --runInBand --forceExit` → 5 failed (badge row, delete dialog header, About sections, tab badge style, `profileBadge`), 6 passed
- green: same command → 11 passed; `bun run test src/navigation.test.tsx src/profile src/notifications --runInBand --forceExit` → 37 passed (navigation alone 21 passed after the `offline()` wrap)

validation: lint ok (`bunx biome ci src/profile src/ui app/(tabs)/_layout.tsx src/navigation.test.tsx`), typecheck ok (`tsc --noEmit`), tests ok (above); no device look
changelog: Added: mobile Profile tab shows the unread notification count as a badge. Changed: mobile notifications row, deletion dialog and About screen follow the v1 look.
roadmap: no box
risks / follow-ups:

- Tab badge: `(tabs)/_layout.tsx` reads `session` and calls `useUnread(notificationsApi, signedIn)`; same query key as `AccountExtras`, so one request. `profileBadge(n)` gives none at 0, the count up to 99, then `99+`. Badge colours come from `tabBarTheme` (`primary` / `on-primary`, Inter 600).
- Notifications row: was a secondary `Btn` labelled "Notifications · N unread"; now a ruled row (display-font title + primary pill with the number). Accessible name is "Notifications, N unread" (or "Notifications"); no existing test queried the old name.
- Deletion dialog: `testID="delete-dialog"`, `border-t-2 border-danger bg-surface` panel, title as header. Button names unchanged (`Delete my account`, `Confirm (n)`, `Yes, delete my account`, `Keep my account`).
- About: version and the TMDB/IGDB attributions in two ruled sections (`about-sources`); testIDs `attribution-tmdb`/`attribution-igdb` kept. No new strings.
- Export: unchanged; it already uses `SettingsSection` and the secondary `Btn` (the v1 primitives).
- No design frame for these screens; the look follows the settings page sections and the rating panel. Merge after P5.B.4.2 and P5.B.4.3.

## Review (especialista, 2026-11-01)

verdict: CORRIGIDO_E_APROVADO
reviewed: batch agent/especialista/mobile-batch (main + status-ui + account-ui + search-i18n)
checked: mobile tsc, biome, jest `bun run test -- --runInBand --forceExit` 55 suites / 362 tests; web vitest 45/165; `go test ./app/i18n`; resources/lang change additive (search.palette keys) in all three locales.
Fix (batch commit): `src/tabsLayout.test.tsx` (added on main after this branch's base) rendered the layout without a QueryClient; wrapped it, since `useUnread` now runs in `(tabs)/_layout.tsx`.

PRONTO_PARA_MERGE agent/especialista/mobile-batch
