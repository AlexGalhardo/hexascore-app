# IMPL-mobile-v1-shell v1 header and tab bar (part 1: presentational)

status: parcial
branch: agent/senior/IMPL-mobile-v1-shell @ 6f6894c (stacked on agent/senior/IMPL-mobile-v1-deps @ 078aa9e)
files: `apps/mobile/src/ui/AppHeader.tsx`, `apps/mobile/src/ui/tabBarTheme.ts`, `apps/mobile/src/ui/shell.test.tsx` (TabIcon came with IMPL-mobile-v1-deps)
acceptance:
- `bun run --cwd apps/mobile test` → 20 suites, 136 tests passed (header: serif wordmark as `header` role, search button, theme toggle moon/sun, status-bar inset; tabBarTheme: primary active, muted inactive, bg + border hairline, Inter 600 12 px labels, light and dark)
validation: lint ok (biome ci), typecheck ok, tests ok; no device run
changelog: Added: v1 app header (serif wordmark, search, theme toggle) and tab bar theme for the Android app.
roadmap: no box
risks / follow-ups:
- Not wired: `apps/mobile/app/` is locked by senior-a's FIX-mobile-root-wiring stack, and the tab labels need the `nav` namespace bundled (`scripts/copy-catalogues.ts` BUNDLED_NAMESPACES + `src/i18n/catalogues/`, also locked). None of senior-a's branches touch `app/(tabs)/_layout.tsx`, so part 2 is small. New row IMPL-mobile-v1-shell-wire:
  - `(tabs)/_layout.tsx`: `screenOptions={{ ...tabBarTheme(scheme), header: () => <AppHeader .../> }}`, order Home, Search, Sections (`[type]`), Lists, Profile, `tabBarIcon: (p) => <TabIcon route="search" {...p} />`, titles from `nav.home`, `nav.search` (short: `nav.searchShort` minus ellipsis or a new `nav.searchTab`), `nav.sections`, `nav.lists`, `nav.profile`.
  - theme: `applyThemePreference` on toggle; `restoreThemePreference` in `src/providers/root.tsx`; search button routes to `/search`.
  - header labels: `nav.openSearch`; theme toggle needs two keys (to dark / to light) in en, pt-BR, es (only `nav.theme` "Toggle theme" exists).

## Review (especialista, 2026-10-31)

verdict: APROVADO (part 1, presentational; the wiring is row IMPL-mobile-v1-shell-wire)
re-run: mobile lint ok, tsc ok, jest ok
checked: tokens only (palette/fontFamily from packages/tokens), 44 dp targets, header role on the wordmark, labels from callers, scope, tests
gate on main after merge (6d5baca, with search-ui ef56eaf): mobile lint ok, tsc ok, jest 28 suites / 205 passed
PRONTO_PARA_MERGE agent/senior/IMPL-mobile-v1-shell
