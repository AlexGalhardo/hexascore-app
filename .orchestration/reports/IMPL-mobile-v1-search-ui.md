# IMPL-mobile-v1-search-ui v1 search sheet (presentational)

status: concluida
branch: agent/senior/IMPL-mobile-v1-search-ui @ fab4d05 (stacked on agent/senior/IMPL-mobile-v1-shell @ 6f6894c)
files: `apps/mobile/src/ui/screens/SearchSheet.tsx`, `apps/mobile/src/ui/screens/SearchSheet.test.tsx`
acceptance:
- `bun run --cwd apps/mobile test` → 21 suites, 140 tests passed (autofocus + cancel; recent tap/remove + trending with "–" score; loading; no results without see-all; results capped at 8 with year, type, score, row opens, see all)
validation: lint ok (biome ci), typecheck ok, tests ok; no device run. Frames: `web-mobile__search-palette__{empty,loading,no-results,results}.png`
changelog: Added: v1 full-screen search sheet component for the Android app.
roadmap: no box
risks / follow-ups:
- Pure props: strings come in as `labels` (functions for counts/queries), so wiring in `app/(tabs)/search.tsx` (locked by senior-a's stack) maps i18n keys + the search query hook into `state`. Keys needed: placeholder (`nav.search` exists), cancel, recent, remove recent, trending, loading, no results, "first N of M", see all; check `search.*` namespace first.
- Results/no-results/loading text is a polite live region (TalkBack announces the count). Each row is one button whose label joins title, year, type and score.
- The first test in the file can hit Jest's 5 s timeout on a cold transform cache; second run is clean.

## Review (especialista, 2026-10-31)

verdict: CORRIGIDO_E_APROVADO
re-run: mobile lint ok, tsc ok, `test -- src/ui` 25 passed
checked: round 2 items 5 and 16 (mobile), one button per row with a joined label, polite live regions, 44 dp targets, labels from callers, scope, tests
fixes: 1. the search TextInput had no accessible name (placeholder only); `accessibilityLabel={labels.placeholder}` plus a test assert (daf7194)
notes: the placeholder uses the RN default colour, not `muted`; check contrast in both themes when the screen is wired (shell-wire / search wiring row)
gate on main after merge (ef56eaf): mobile lint ok, tsc ok, jest 28 suites / 205 passed
PRONTO_PARA_MERGE agent/especialista/IMPL-mobile-v1-search-ui-fix
