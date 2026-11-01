# IMPL-mobile-v1-item wire ItemDetail into the item screen

status: concluida
branch: agent/senior/IMPL-mobile-v1-item @ 05f95dc (stacked on IMPL-mobile-v1-listing a89f72d, main merged at 201f747)
files: `apps/mobile/src/screens/item/Item.tsx`, `apps/mobile/src/screens/item/item.test.tsx`, `apps/mobile/src/screens/item/Attribution.tsx` (padding/font only), `apps/mobile/src/screens/item/ItemHeader.tsx` (deleted), `apps/mobile/src/ui/screens/ItemDetail.tsx`
acceptance:
- red: `bun run --cwd apps/mobile test -- src/screens/item/item.test.tsx` with the new tests and the old screen → 4 failed, 5 passed
- green: `bun run --cwd apps/mobile test -- src/screens/item src/ui/screens/ItemDetail.test.tsx src/ui/screens/EditionsList.test.tsx src/rating src/navigation.test.tsx` → 53 passed; `src/i18n` → passed
validation: `bunx biome check src/screens src/ui scripts` ok, `bun run typecheck` ok; no device run
changelog: Changed: the Android item screen uses the v1 item page (kicker, score badge with verdict, genre chips, category bars, synopsis, people, editions, attribution); Rate opens the rating sheet.
roadmap: no box
risks / follow-ups:
- Rating: P5.B.2.2 and P5.B.2.3 are merged, so Rate opens the existing `RateSheet` modal (senior-a's `StarInput` + `useRateMutation`); `ItemDetail` gets `rating={null}`, so its inline panel is not used on mobile (no second input). The saved overall shows as a "Your overall score" badge row (`ownScore`), and the button reads "Update rating" once rated.
- `ItemDetail` changes (additive, existing tests unchanged): `testID="item"`, a `score-band-*` wrapper on the header badge, optional `statusLabel`/`onStatus`/`onAddToList` (hidden until library status and lists are wired, P5.B.4.x), optional `ownScore`, and a `children` slot at the end for entity groups, `EditionsList` and attribution.
- Data mapping: kicker `catalog.title.<type>` · year; `role === "genre"` entities become chips and leave the entity groups; synopsis from `synopsis.text`; categories from `scoreByCategory`. Confidence is no longer shown (the v1 design has none). Byline null; `coverUri` null until FIX-app-item-cover-urls merges (one-line shortcut in `detailOf`). Edition facts show the publisher only (no translated labels for year/ISBN yet). No "more like this" (no similar-items endpoint in `/app/v1`).
- `item.rateIt` came from `main` (b3dbc67), hence the `main` merge; the catalogues needed no recopy after it.

## Review (especialista, 2026-11-01)
verdict: CORRIGIDO_E_APROVADO
re-run: `scripts/openapi-check.sh` -> ok (golden regenerated with `--write`, client with `codegen:client -prefix /app/v1`, 23d1177); `go test -tags integration -run 'TestAppCatalog$|TestCoversBaseJSONAbsolute|TestAPITrending|TestAPIItems|TestRouteInventory|TestOpenAPI|TestAppSearch|TestSearch|Suggest|TestAPIEdge' ./tests/` -> ok; golangci-lint (touched pkgs, integration tag) 0 issues; app-client typecheck/test ok; mobile lint/typecheck ok, mobile Jest 297 tests (3 suites timed out under load on the full run, green alone); root lint/typecheck/vitest 129 ok
checked: sdd 05 (paid API keeps no covers, slug lookup app-only), sdd 10 FR-10-7, authz (scopes unchanged), sql (bound params), i18n (catalogues recopied), scope, tests
fixes: 1. coverUri = item.coverUrl; App* type names (4286dab)
gate on main after merge: local.sh test-go, test-go-integration, test-ts, migrations, vuln, deps, setup-windows, mobile, build PASS on bbe689d/fcf63bf; lint fails only on commitlint of historic local commits; lighthouse fails on main before this batch (FIX-lighthouse-budgets-main)
PRONTO_PARA_MERGE agent/especialista/FIX-app-batch (merged bd95e7d)
