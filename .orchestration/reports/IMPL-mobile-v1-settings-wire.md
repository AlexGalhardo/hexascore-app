# IMPL-mobile-v1-settings-wire v1 settings page in the profile tab

status: concluida
branch: agent/senior/IMPL-mobile-v1-settings-wire @ 740e91e (stacked on agent/senior/IMPL-mobile-v1-detail-ui @ 05e50d0)
files: `apps/mobile/app/(tabs)/profile.tsx`, `apps/mobile/src/telemetry/CrashReportSwitch.tsx`, `apps/mobile/src/settingsWire.test.tsx`, `apps/mobile/src/ui/screens/SettingsPage.tsx`, `apps/mobile/src/navigation.test.tsx`, `resources/lang/{en,pt-BR,es}.json` (additive `account.settings.*`, 7 keys), `apps/mobile/src/i18n/catalogues/*` (recopied with `bun run catalogues`, 7 added lines each)
acceptance:
- red run (test first): `bun run --cwd apps/mobile test -- src/settingsWire` → suite failed, `Cannot find module './telemetry/CrashReportSwitch'` (0 tests)
- green: `bun run --cwd apps/mobile test -- src/settingsWire src/navigation src/ui/screens/SettingsPage src/i18n` → 33 passed (profile tab renders the settings page in pt-BR; theme radios apply and store the mode via `applyThemePreference(memoryPreferences, …)` and reflect `checked`; signed out "Log in" → `router.navigate("/login")`; consent switch off while unset, on when granted, writes both ways). One earlier run had a cold-cache 5 s timeout in `SettingsPage.test` (known pitfall); the rerun was green.
- `bun run --cwd apps/mobile lint` → biome ci clean (141 files); `bun run --cwd apps/mobile typecheck` → clean
validation: lint ok, typecheck ok, targeted tests ok; no device run
changelog: Added: the Android profile tab shows the v1 settings page (theme mode, crash-report consent switch, log in / log out).
roadmap: no box (P5.B.4.2 stays open)
risks / follow-ups:
- P5.B.4.2 (senior-a) owns `GET/PATCH /me`, devices and the opt-out toggle: it should reuse `CrashReportSwitch` for the consent toggle (do not build a second one) and fill `account.name`/`account.email`, which this row leaves null (`SettingsPage.account` fields are now nullable; the generated `meShow()` returns `unknown` and `/app/v1/me` 200 declares no schema). P5.B.4.2's planned `src/profile/Settings.tsx` can wrap or replace the route body; the route file is the shared touch point.
- Theme preference uses `memoryPreferences` like the shell-wire header toggle, so it does not survive a restart until a persistent `PreferenceStore` lands (same limit as IMPL-mobile-v1-shell-wire).
- Logout label reuses `account.sessions.revoke` ("Log out"), sign-in reuses `auth.login.title`, privacy heading `account.privacy.title`.
- The consent switch uses the RN `Switch` (no house switch component exists).

## Review (especialista, 2026-10-31)

verdict: APROVADO
re-run: batch branch agent/especialista/IMPL-mobile-batch (all 8 rows merged on main 5100631, conflicts only in navigation.test.tsx): `bun run --cwd apps/mobile catalogues` no diff, `lint` ok, `typecheck` ok, `test` 45 suites / 275 tests ok; `vitest packages/i18n` 5 ok; `go test ./tests/i18n` ok\nchecked: SDD 10, i18n parity (all labels through t(), nav namespace bundled), no hardcoded copy, no secrets, scope, tests\ngate on main after merge: mobile lint/typecheck/test ok (full local.sh at the end of the batch)\nPRONTO_PARA_MERGE agent/senior/IMPL-mobile-v1-settings-wire
