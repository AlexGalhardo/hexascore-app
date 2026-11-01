# IMPL-mobile-v1-home wire HomeSections into the home screen

status: concluida
branch: agent/senior/IMPL-mobile-v1-home @ 60ec6a9 (from main 5100631)
files: `apps/mobile/src/screens/Home.tsx`, `apps/mobile/src/screens/home.test.tsx`, `apps/mobile/app/(tabs)/index.tsx`, `apps/mobile/src/ui/screens/HomeSections.tsx` (rank label gets the item), `apps/mobile/scripts/copy-catalogues.ts` + `src/i18n/catalogues/*.json` (bundle `home.*`, 7 keys), `apps/mobile/src/components/CategorySection.tsx` (deleted)
acceptance:
- red: `bun run --cwd apps/mobile test -- src/screens/home.test.tsx` → 3 failed, 4 passed (before the wiring)
- `bun run --cwd apps/mobile test -- src/screens/home.test.tsx src/ui/screens/HomeSections src/navigation.test.tsx src/i18n` → 36 passed
validation: lint ok (biome ci), typecheck ok, tests ok; no device run
changelog: Changed: the Android home shows the v1 trending sections (three ranked cards and a grid of six per type, see-all to the section).
roadmap: no box
risks / follow-ups:
- P5.B.1.1 behaviour kept: ITEM_TYPES order, failed/empty type hidden, empty state, all-prior title "Popular {type}" (SDD 08 §6); a prior entry's rank reads "#n · Popular" instead of "#n · Trending this week".
- Data mapping: entries 1-3 are cards, 4-9 the grid; year from `releaseDate`; `coverUri` null until FIX-app-item-cover-urls merges (2:3 box reserved; one-line follow-up in `homeItem`).
- Labels reuse existing keys: see-all `catalog.landing.seeAll` ("All Games"), verdict `item.notEnough` / `item.communityScore`, count `catalog.ratingCount`, score `score.outOf` / `score.none`. No singular type key exists, so the card meta reads "2028 · Games".
- `ItemRow` stays (used by Listing and the item screen); only `CategorySection` was list-only.
- `FlatList` → `ScrollView`: at most 4 × 9 cells, no virtualisation needed.

## Review (especialista, 2026-11-01)
verdict: CORRIGIDO_E_APROVADO
re-run: `scripts/openapi-check.sh` -> ok (golden regenerated with `--write`, client with `codegen:client -prefix /app/v1`, 23d1177); `go test -tags integration -run 'TestAppCatalog$|TestCoversBaseJSONAbsolute|TestAPITrending|TestAPIItems|TestRouteInventory|TestOpenAPI|TestAppSearch|TestSearch|Suggest|TestAPIEdge' ./tests/` -> ok; golangci-lint (touched pkgs, integration tag) 0 issues; app-client typecheck/test ok; mobile lint/typecheck ok, mobile Jest 297 tests (3 suites timed out under load on the full run, green alone); root lint/typecheck/vitest 129 ok
checked: sdd 05 (paid API keeps no covers, slug lookup app-only), sdd 10 FR-10-7, authz (scopes unchanged), sql (bound params), i18n (catalogues recopied), scope, tests
fixes: 1. cover shortcut removed: coverUri = item.coverThumbUrl; App* type names (4286dab) 2. catalogues recopied (home.* keys from the web home)
gate on main after merge: local.sh test-go, test-go-integration, test-ts, migrations, vuln, deps, setup-windows, mobile, build PASS on bbe689d/fcf63bf; lint fails only on commitlint of historic local commits; lighthouse fails on main before this batch (FIX-lighthouse-budgets-main)
PRONTO_PARA_MERGE agent/especialista/FIX-app-batch (merged bd95e7d)
