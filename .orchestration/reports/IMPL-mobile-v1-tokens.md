# IMPL-mobile-v1-tokens

Branch `agent/senior/IMPL-mobile-v1-tokens` @ 33ece5f (from local `main` 2f4cb21). Worker: impl-mobile.

## What

- `packages/tokens/tokens.json`: every colour now holds the v1 Editorial round-2 value from `docs/openpencil/v1-editorial/README.md` (palette table), converted hex -> OKLCH with values that round-trip to the exact design hex through `oklchToHex`. Names are additive: `bg`, `fg`, `muted`, `surface`, `border` (v1 `rule`), accents, scores, `star`, `danger` keep their names, so web classes keep compiling. New: `primary`, `primary-hover`, `on-primary`, `ledge-primary`, `ledge-secondary`, `secondary`, `field`, `field-bg`, `on-score` (the ink number on score fills, same in both themes, per ROUND-2 contrast table), `star-line`.
- `textPairs` gains `primary/bg`, `on-primary/primary`, `secondary/bg` and `on-score` on the four score fills; all pass 4.5:1 in light and dark.
- `tokens.css` / `tokens.ts` regenerated with `bun run --cwd packages/tokens build`.
- `apps/mobile/tailwind.config.js`: `darkMode: "class"` (NativeWind v4) so `colorScheme.set` can override the system.
- `apps/mobile/src/theme/mode.ts`: `readThemePreference`, `applyThemePreference`, `restoreThemePreference` (system / light / dark) over the existing `PreferenceStore` port (read-only use of `src/platform/preferences.ts`; persisted once senior-a's SQLite store lands). The toggle UI comes in the shell row.

## Validation

- `bunx vitest run packages/tokens`: 34 passed.
- `bunx biome ci packages/tokens`: clean.
- `bun run --cwd apps/mobile lint`: clean. `typecheck`: clean. `test`: 16 suites, 115 tests passed (incl. `src/theme/mode.test.ts`, and `contrast.test.ts` proving NativeWind hex == token hex).

## Notes for review

- Web impact: IMPL-web-v1 reads these tokens; the web picks up the v1 colours immediately (no class renames needed). P2.H.3.4 (`primary` token) is now satisfied at the token level; its class replacement in `resources/js` is still open.
- Fonts (Playfair Display + Inter) are not bundled: system serif/sans until the owner OKs the font files (two-family budget, OFL). Question to the orchestrator in the handoff.

Changelog line (for the specialist): `Changed: design tokens use the v1 Editorial palette (light and dark) and add primary, ledge, field and on-score tokens; the Android app can follow the system theme or a stored light/dark choice.`

## Review (especialista, 2026-10-31)

verdict: APROVADO
re-run: `bun run --cwd packages/tokens build` → no diff in tokens.css/tokens.ts; `bunx vitest run packages/tokens` → 34 passed (every textPair >= 4.5:1 in light and dark, incl. primary, on-primary, secondary, on-score on the four score fills); web `bun run typecheck` ok, `vite build` ok, `build:ssr` ok, `bun run lint` ok, `bun run test` 28 files / 111 passed; mobile lint, typecheck ok, test 16 suites / 115 passed
checked: design v1 round 2 palette (Q-D2), token names additive (web classes keep compiling), NativeWind `darkMode: "class"` with `colorScheme.set("system"|"light"|"dark")`, scope, tests
notes: P2.H.3.4 stays open (token exists; the `indigo-*` class replacement in resources/js is not done). Fonts not bundled.
gate on main after merge (7f4777b): tokens vitest 34 passed, web tsc ok, mobile tsc ok, mobile jest 129 passed
PRONTO_PARA_MERGE agent/senior/IMPL-mobile-v1-tokens
