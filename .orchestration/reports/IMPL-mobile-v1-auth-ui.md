# IMPL-mobile-v1-auth-ui v1 login, signup and settings (presentational)

status: concluida
branch: agent/senior/IMPL-mobile-v1-auth-ui @ 759e80f (stacked on agent/senior/IMPL-mobile-v1-library-ui @ 7df8b4e)
files: `apps/mobile/src/ui/GoogleButton.tsx`, `apps/mobile/src/ui/TextField.tsx`, `apps/mobile/src/ui/screens/AuthPage.tsx`, `apps/mobile/src/ui/screens/SettingsPage.tsx`, their `*.test.tsx`, `apps/mobile/assets/brand/google-g.svg`
acceptance:
- red run (test first): `bun run --cwd apps/mobile test -- src/ui/screens/AuthPage src/ui/screens/SettingsPage` → 2 suites failed, `Cannot find module '../TextField'` / `'./SettingsPage'` (0 tests)
- green: same command → 7 passed (kicker, header, lede, Google button with the G mark above the "or" divider, footer link; dark brand colours + disabled; form slot; password Show/Hide; field error as alert; settings account block, System/Light/Dark radios with checked state, crash-report slot under Privacy, Log out; signed-out Log in)
- `bun run --cwd apps/mobile lint` → biome ci clean (128 files); `typecheck` → clean
validation: lint ok, typecheck ok, targeted tests ok; no device run (the SVG mark renders through expo-image, untested on a device). Frames: `web-mobile__login__default.png`, `web-mobile__signup__default.png` (+ dark)
changelog: Added: v1 login/signup page, Google sign-in button and settings (theme mode, crash-report slot) components for the Android app.
roadmap: no box
risks / follow-ups:
- Google branding (owner item 15): official four-colour G (`assets/brand/google-g.svg`), Google's light `#FFFFFF/#747775/#1F1F1F` and dark `#131314/#8E918F/#E3E3E3` as arbitrary-value classes, the documented exception in `docs/openpencil/v1-editorial/README.md`; not tokens on purpose. Deviations from Google's guide, as the README flags: Inter instead of Roboto Medium and the 2 px v1 ledge. Wording keys ("Sign in with Google" / "Sign up with Google" and pt-BR/es) come in the wiring row.
- Signup: SDD 10 §1 sends signup to web `/signup`, so the mobile signup screen is `AuthPage` with Google plus a "continue on the web" action in the slot; no signup form on the device. Remember me and passkey from the web frame are web-only.
- Wiring (later row): restyle `src/auth/LoginForm` onto `AuthPage` + `TextField` (two-factor and device-name fields keep their logic), settings onto `applyThemePreference` and the telemetry consent switch.

## Review (especialista, 2026-10-31)

verdict: APROVADO
re-run: batch branch agent/especialista/IMPL-mobile-batch (all 8 rows merged on main 5100631, conflicts only in navigation.test.tsx): `bun run --cwd apps/mobile catalogues` no diff, `lint` ok, `typecheck` ok, `test` 45 suites / 275 tests ok; `vitest packages/i18n` 5 ok; `go test ./tests/i18n` ok\nchecked: SDD 10, i18n parity (all labels through t(), nav namespace bundled), no hardcoded copy, no secrets, scope, tests\ngate on main after merge: mobile lint/typecheck/test ok (full local.sh at the end of the batch)\nPRONTO_PARA_MERGE agent/senior/IMPL-mobile-v1-auth-ui
