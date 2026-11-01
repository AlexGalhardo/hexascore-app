# IMPL-mobile-v1-listing wire CatalogPage into the section listing

status: concluida
branch: agent/senior/IMPL-mobile-v1-listing @ a89f72d (from main 54105f9)
files: `apps/mobile/src/screens/Listing.tsx`, `apps/mobile/src/screens/listing.test.tsx`, `apps/mobile/src/ui/screens/CatalogPage.tsx` (`total` nullable, `testID="listing"`), `resources/lang/*.json` + `apps/mobile/src/i18n/catalogues/*.json` (additive `catalog.page`), `apps/mobile/scripts/copy-catalogues.ts` (bundles `score`)
acceptance:
- red: `bun run --cwd apps/mobile test -- src/screens/listing.test.tsx` with HEAD `Listing.tsx`/`CatalogPage.tsx` → 4 failed, 1 passed
- green: `bun run --cwd apps/mobile test -- src/screens/listing.test.tsx src/ui/screens/CatalogPage.test.tsx` → 10 passed; `src/i18n/keys.test.ts` → 4 passed; `src/navigation.test.tsx` → 19 passed
validation: `bunx biome check src/screens src/ui/screens/CatalogPage.tsx src/i18n` ok, `bun run typecheck` ok; no device run
changelog: Changed: the Android section listing uses the v1 catalogue page (title, filter panel with sort and genre chips, Previous/Next pager, rows with year, rating count and score).
roadmap: no box
risks / follow-ups:
- P5.B.1.2 data flow kept: `useInfiniteQuery` on `queryKeys.listing`, opaque `next_cursor` passed back as received, sort/genre reset to page 1.
- The board note asked to keep `getItemLayout`; the v1 `CatalogPage` design is paged (Previous/Next) in a `ScrollView`, so there is no `FlatList` and no infinite scroll any more. One page is at most one API page (20 rows), so no virtualisation is needed. Page n is fetched from page n-1's cursor; Previous reads the cache.
- The cursor API sends no total: the "N items found" line is hidden (`total: null`), and the pager reads "Page n" while a next page exists, "Page n of n" on the last one (new key `catalog.page`, en/pt-BR/es).
- Genre chips show the slug (genres endpoint has no name yet). Rows have no byline and `coverUri` null until FIX-app-item-cover-urls merges (`coverThumbUrl` is not on `main` yet; one-line shortcut in `toRow`).
- Test client now mirrors the app `staleTime` (30 s) so the "Reset reads the cache" assertion matches runtime behaviour.
- a89f72d: `score.outOf`/`score.none` (row and home score labels) were not in `BUNDLED_NAMESPACES`, so the accessibility label read the raw key; `score` is now bundled (red: the new "Score 7.5 out of 10" assertion failed, green after `bun run catalogues`). IMPL-mobile-v1-home uses the same keys and gets the fix when both merge; home's branch predates `nav` in the namespace list, so recopy the catalogues after merging both.

## Review (especialista, 2026-11-01)
verdict: CORRIGIDO_E_APROVADO
re-run: `scripts/openapi-check.sh` -> ok (golden regenerated with `--write`, client with `codegen:client -prefix /app/v1`, 23d1177); `go test -tags integration -run 'TestAppCatalog$|TestCoversBaseJSONAbsolute|TestAPITrending|TestAPIItems|TestRouteInventory|TestOpenAPI|TestAppSearch|TestSearch|Suggest|TestAPIEdge' ./tests/` -> ok; golangci-lint (touched pkgs, integration tag) 0 issues; app-client typecheck/test ok; mobile lint/typecheck ok, mobile Jest 297 tests (3 suites timed out under load on the full run, green alone); root lint/typecheck/vitest 129 ok
checked: sdd 05 (paid API keeps no covers, slug lookup app-only), sdd 10 FR-10-7, authz (scopes unchanged), sql (bound params), i18n (catalogues recopied), scope, tests
fixes: 1. catalog.filters is a plural object on main since the web catalogue: the label now calls t(..., { count }) (2 listing tests were red after the merge) 2. coverUri = item.coverThumbUrl, byline shortcut kept (4286dab)
gate on main after merge: local.sh test-go, test-go-integration, test-ts, migrations, vuln, deps, setup-windows, mobile, build PASS on bbe689d/fcf63bf; lint fails only on commitlint of historic local commits; lighthouse fails on main before this batch (FIX-lighthouse-budgets-main)
PRONTO_PARA_MERGE agent/especialista/FIX-app-batch (merged bd95e7d)
