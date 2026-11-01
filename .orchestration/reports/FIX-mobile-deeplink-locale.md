# FIX-mobile-deeplink-locale App Link locale for the session only

status: concluida
branch: agent/senior/FIX-mobile-deeplink-locale @ 8909716 (from local main 6440d0a, which holds P5.B.4.2 and P5.B.5.1; worktree `wt/impl-mobile-detail`)
files: `apps/mobile/app/+native-intent.tsx`, `apps/mobile/src/i18n/locale.ts`, `apps/mobile/src/i18n/locale.test.ts`, `apps/mobile/src/nativeIntent.test.ts` (new)
acceptance:

- red: `bun run test -- src/i18n/locale.test.ts src/nativeIntent.test.ts --runInBand --forceExit` → 3 failed (`setSession` missing; `/pt-br` and `/es` links not applied), 4 passed
- green: `bun run test -- src/i18n/locale.test.ts src/nativeIntent.test.ts src/linking.test.ts --runInBand --forceExit` → 18 passed

validation: lint ok (`bun run lint`, 194 files), typecheck ok (`bun run typecheck`), tests ok (above); no device look (an App Link needs a signed build on a device)
changelog: Fixed: opening a `/pt-br` or `/es` App Link shows the mobile app in that language for the session, without changing the saved language.
roadmap: no box (P5.B.5.1 note)
risks / follow-ups:

- `LocaleStore` gains `setSession(locale)`: same notify path as `set`, without `writeLocaleOverride`. Subscribers (the root locale provider) re-render, so no restart.
- `redirectSystemPath` calls `appLocale.setSession` from `auth/runtime` when `resolveDeepLink` returns a locale. A link without a prefix keeps the current locale. `auth/runtime.ts` was not changed (on the row's lock set, not needed).
- The next cold start reads the stored override or the device language again, as the ROADMAP note asks.
- `nativeIntent.test.ts` mocks `auth/runtime` (only `appLocale.setSession`), so it does not load the real runtime.

## Review (especialista, 2026-11-01)

verdict: APROVADO
re-run: in batch `agent/especialista/mobile-batch-2` (main 35cd2062 + the six mobile branches): `bun run lint` (biome ci, 195 files) ok, `bun run typecheck` ok, `bun run test -- --runInBand --forceExit` → 56 suites, 373 tests passed
checked: sdd 10, authz, secrets, sql (n/a), i18n (no resources/lang change), scope, tests
notes: LocaleStore.setSession applies without writing the override; +native-intent calls it only when resolveDeepLink returns a locale.
gate on main after merge: mobile lint/typecheck/jest ok (merge 736875d4); no Go or web TS touched
PRONTO_PARA_MERGE agent/especialista/mobile-batch-2 (merged 736875d4)
