# IMPL-mobile-v1-item-ui v1 item detail with rating (presentational)

status: concluida
branch: agent/senior/IMPL-mobile-v1-item-ui @ 851f0a3 (stacked on agent/senior/IMPL-mobile-v1-catalog-ui @ b40410b)
files: `apps/mobile/src/ui/screens/ItemDetail.tsx`, `apps/mobile/src/ui/screens/ItemDetail.test.tsx`
acceptance:
- `bun run --cwd apps/mobile test -- src/ui/screens/ItemDetail` → 6 tests passed (kicker, title header, byline, genre chips `min-h-11`, lg score badge + "verdict · N ratings"; Rate / status / add-to-list callbacks; category bars labelled "<category>, <score>" with "–" when null; rating panel with one adjustable RatingInput per category, increment → onRateCategory(key, 8), overall badge, Save; saving disables Save; no score, no synopsis note, more like this opens)
- `bun run --cwd apps/mobile test` (stack tip) → 24 suites, 154 tests passed
validation: lint ok (biome ci), typecheck ok, tests ok; no device run. Frames: `android__item__default.png`, `web-mobile__item__{logged-in,logged-out,no-score,no-cover,no-synopsis}.png`, `web-mobile__rating-widget__states.png`
changelog: Added: v1 item detail component (score block, category bars, per-category half-star rating, more like this) for the Android app.
roadmap: no box
risks / follow-ups:
- Category bar fill width is a native `style={{ width: "<score*10>%" }}` (data-driven; no utility class can express it). The CSP rule targets web style attributes; flag if the reviewer wants another approach.
- Out of scope (web-only sections in the frame): radar chart, details table, time to beat, trailers, suggest an edit. Status picker and add-to-list should open `@expo/ui` bottom sheets from `onStatus` / `onAddToList` in the item wiring row.
- `rating === null` hides the panel (logged out); wiring routes `onRate` to login or scrolls to the panel.

## Review (especialista, 2026-10-31)

verdict: APROVADO
re-run: batch branch agent/especialista/IMPL-mobile-batch (all 8 rows merged on main 5100631, conflicts only in navigation.test.tsx): `bun run --cwd apps/mobile catalogues` no diff, `lint` ok, `typecheck` ok, `test` 45 suites / 275 tests ok; `vitest packages/i18n` 5 ok; `go test ./tests/i18n` ok\nchecked: SDD 10, i18n parity (all labels through t(), nav namespace bundled), no hardcoded copy, no secrets, scope, tests\ngate on main after merge: mobile lint/typecheck/test ok (full local.sh at the end of the batch)\nPRONTO_PARA_MERGE agent/senior/IMPL-mobile-v1-item-ui
