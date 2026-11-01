# IMPL-mobile-v1-home-ui v1 home sections (presentational)

status: concluida
branch: agent/senior/IMPL-mobile-v1-home-ui @ 8cf2899 (stacked on agent/senior/IMPL-mobile-v1-search-ui @ fab4d05)
files: `apps/mobile/src/ui/screens/HomeSections.tsx`, `apps/mobile/src/ui/screens/HomeSections.test.tsx`
acceptance:
- `bun run --cwd apps/mobile test -- src/ui/screens` → 2 suites, 7 tests passed (section headers + see-all link per type; trending capped at 3 with rank, "year · type", verdict, rating count, "–" for no score, card opens with type; grid capped at 6 with title, year, score)
validation: lint ok (biome ci), typecheck ok, tests ok; no device run. Frames: `web-mobile__home__default.png`, `android__home__default.png`
changelog: Added: v1 home sections component (trending cards and grid per type) for the Android app.
roadmap: no box
risks / follow-ups:
- Pure props: `sections` (type, title, typeLabel, trending, grid) and `labels` (seeAll, rank, verdict, ratings, score) come from the wiring row (home wiring after P5.B.1.1). Accent per type uses literal class maps so NativeWind compiles them.
- Overlaps visually with P5.B.1.1's `src/components/CategorySection` + `ItemRow` (senior-a): wiring should pick one; this component follows the round-2 frame (cards + grid), the P5.B.1.1 one is a row list. Row IMPL-mobile-v1-home (restyle) may become "wire HomeSections" instead.
- Grid gutters use token spacing (`px-1`, `-mx-1`); covers keep the 2:3 reserved box.

## Review (especialista, 2026-10-31)

verdict: CORRIGIDO_E_APROVADO
re-run: mobile lint ok, tsc ok, `test -- src/ui/screens` 7 passed (first cold run hit the known 5 s SearchSheet timeout; warm run clean)
checked: round 2 home frames, literal accent class maps, 2:3 covers, header roles, one button per card with joined label, labels from callers, scope, tests
fixes: 1. contrast: the 12 px rank label was `accent-*` on `surface`; `accent-book` on `surface` is 4.28:1 in light (below 4.5:1), so the rank text is now `fg`, the accent stays on the card's top border and the section square (583568f)
notes: the token test covers accent pairs on `bg` only; any future accent text on `surface` must be checked (accent-movie 4.54, accent-book 4.28 light). Overlap with P5.B.1.1 CategorySection/ItemRow: the wiring row picks one.
gate on main after merge (e12f65c): mobile tsc ok, jest 29 suites / 208 passed
PRONTO_PARA_MERGE agent/especialista/IMPL-mobile-v1-home-ui-fix
