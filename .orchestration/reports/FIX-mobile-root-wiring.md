# FIX-mobile-root-wiring

Branch `agent/senior/FIX-mobile-root-wiring` @ 9e90725 (2 commits on `main` 127eeeb). Unblocks the P5.A.3.5 box (426 now routes to the screen).

## What changed

- `apps/mobile/src/providers/root.tsx` (new): `useRoot(deps)` restores the session, starts the foreground rotation, listens for the client's upgrade signal (and reads `upgradeRequired()` again in the effect so a 426 seen before mount is kept), cleans everything up on unmount; `routeGuards(state, upgrade)` (update wins; `active`/`deletionScheduled` -> app; `signedOut`/`banned` -> login); `LocaleProvider` / `useLocale`.
- `apps/mobile/app/_layout.tsx`: `Stack.Protected` groups (expo-router 58 `guard` prop) for `update-required`, `(auth)/login` and the app routes, wrapped in `LocaleProvider`. Returns `null` while the session is `loading` (every guard would be false).
- `apps/mobile/src/auth/runtime.ts`: device locale via `Intl.DateTimeFormat().resolvedOptions().locale` + `resolveLocale` (no localization package; Hermes ships Intl), passed to the client's `Accept-Language`; `onUpgradeRequired` fan-out to the root; `rootDeps` with rotation through `appClient.write` (POST `/tokens/rotate` carries an Idempotency-Key like every write).
- `LoginForm`: device name label `t(locale, "auth.fields.deviceName")`. `update-required.tsx`: title `account.updateRequired.title`, logout label in the active locale. `(auth)/login.tsx`: locale from `useLocale()`.
- New keys in `resources/lang/{en,pt-BR,es}.json`: `auth.fields.deviceName` (Device name / Nome do dispositivo / Nombre del dispositivo), `account.updateRequired.title` (Update required / Atualização necessária / Actualización necesaria); `bun run --cwd apps/mobile catalogues` recopied (427 keys per locale).

SDD: 10 §6.1 (update-required, login routes), §6.2 (rotate on foreground, 401/ban wipe -> login), FR-10-5 (update screen), FR-10-6 (locale match).

## Tests (written first; red run: 2 failed / 22 passed for LoginForm + navigation, root tests failed to import)

- `src/providers/root.test.tsx`: restore + rotation start/stop, 426 after mount, 426 before mount, guard table, locale provider.
- `LoginForm.test.tsx`: pt-BR device name label. `navigation.test.tsx`: update-required in es (title + logout); `Stack.Protected` added to the expo-router mock.
- `bun run --cwd apps/mobile lint`: `Checked 73 files ... No fixes applied.` `typecheck`: clean. `test`: 16 suites, 105 tests passed (before the last null-while-loading commit; navigation re-run after it: 19/19). `bunx vitest run packages/i18n/src/parity.test.ts`: 5 passed.

## Not done / limits

- Banned and deletion-scheduled screens do not exist yet: banned goes to login (token already wiped), deletion-scheduled stays in the app; their screens belong to a later row (SDD 10 §6.2).
- Locale override UI and persistence wait on P5.A.4.1 (Expo SQLite preferences); the device tag is the only input today.
- No device run (owner-gated). `Stack.Protected` behaviour checked against the expo-router 58 type declarations only, not on a device.

## Changelog line (for the specialist)

- Fixed: the Android root layout restores the session, rotates the token on foreground, and routes a 426 to the update-required screen; login and update screens use the device locale and catalogue strings.

## Review (especialista, 2026-10-31)

verdict: APROVADO
re-run: see gate below (merged as one ordered stack: root-wiring, P5.A.4.1, P5.A.4.2, P5.A.4.3, blocked-screens)
checked: SDD 10 §6.1, §6.2, FR-10-5, FR-10-6; guards (update wins, loading renders nothing), 426 before mount, rotation write carries an Idempotency-Key; i18n en/pt-BR/es parity; scope; tests
gate on main after the stack merge (3e10c99): `bun run --cwd apps/mobile catalogues` no diff (437 keys per locale); mobile lint ok, tsc ok, jest 26 suites / 198 passed; packages/i18n parity 5 passed; web vitest 28 files / 111 passed; `go test ./app/i18n/... ./tests/i18n/...` ok
merge: 21ab908
PRONTO_PARA_MERGE agent/senior/FIX-mobile-root-wiring
