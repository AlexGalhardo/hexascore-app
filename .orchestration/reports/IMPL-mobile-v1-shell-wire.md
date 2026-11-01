# IMPL-mobile-v1-shell-wire v1 header and tab bar wiring

status: concluida
branch: agent/senior/IMPL-mobile-v1-shell-wire @ 5d7285c (from main 679005b)
files: `apps/mobile/app/(tabs)/_layout.tsx`, `apps/mobile/src/ui/TabIcon.tsx`, `apps/mobile/scripts/copy-catalogues.ts`, `apps/mobile/src/i18n/catalogues/*.json`, `resources/lang/{en,pt-BR,es}.json` (additive `nav.searchTab`, `nav.themeDark`, `nav.themeLight`), `apps/mobile/src/tabsLayout.test.tsx` (new), `apps/mobile/src/navigation.test.tsx` (router mock gets `useRouter`, safe-area mock)
acceptance:
- red run (test first): `bun run --cwd apps/mobile test -- src/tabsLayout` → 2 failed / 2 (order, titles, icons, header absent in the old layout)
- green: `bun run --cwd apps/mobile test -- src/tabsLayout src/navigation` → 2 suites, 21 passed; `src/i18n scripts` → 7 passed; `src/ui/shell src/ui/icons` → passed
- `bun run --cwd apps/mobile catalogues` → en/pt-BR/es 466 keys each (`nav` namespace now bundled)
- `bun run --cwd apps/mobile lint` → biome ci clean (116 files); `typecheck` → clean
validation: lint ok, typecheck ok, targeted tests ok; no device run
changelog: Added: Android app header (wordmark, search, theme toggle) and themed tab bar with translated titles.
roadmap: no box
risks / follow-ups:
- Theme preference persists through `memoryPreferences` (process-local, as `src/platform/preferences.ts` documents), so the choice resets on cold start until the preferences row moves to Expo SQLite; `restoreThemePreference` runs once in the tabs layout so the wiring is in place for that switch.
- `TabIcon.color` widened to `ColorValue` (the navigator's type); non-string platform colours fall back to the glyph default, which our hex tab theme never produces.
- `app/_layout.tsx` and `src/providers/root.tsx` untouched: the restore lives in the tabs layout to avoid a nativewind import in the root provider tests.

## Review (especialista, 2026-10-31)

verdict: APROVADO
re-run: batch branch agent/especialista/IMPL-mobile-batch (all 8 rows merged on main 5100631, conflicts only in navigation.test.tsx): `bun run --cwd apps/mobile catalogues` no diff, `lint` ok, `typecheck` ok, `test` 45 suites / 275 tests ok; `vitest packages/i18n` 5 ok; `go test ./tests/i18n` ok\nchecked: SDD 10, i18n parity (all labels through t(), nav namespace bundled), no hardcoded copy, no secrets, scope, tests\ngate on main after merge: mobile lint/typecheck/test ok (full local.sh at the end of the batch)\nPRONTO_PARA_MERGE agent/senior/IMPL-mobile-v1-shell-wire
