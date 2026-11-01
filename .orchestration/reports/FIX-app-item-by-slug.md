# FIX-app-item-by-slug slug lookup on GET /app/v1/{type}/{id}
status: concluida
branches:
- Go: agent/senior/FIX-app-item-by-slug @ a02f47a (from main cf4683c)
- mobile: agent/senior/FIX-app-item-by-slug-mobile @ 5de0017, stacked on agent/senior/P5.B.2.3 (the item screen and `itemIdOf` exist only on that chain); merge after P5.B.2.3 and after the Go branch
files: app/controllers/publicapi/items.go, app/appapi/catalog.go, routes/app_catalog.go, openapi/app-v1.json, packages/app-client/src/api.ts, tests/app_catalog_test.go; apps/mobile/app/[type]/[slug].tsx, apps/mobile/src/navigation.test.tsx
acceptance:
- red (Go): `go test -tags integration -count=1 -run 'TestAppCatalog$' ./tests/` -> `app_catalog_test.go:88: status = 404, want 200` (GET /app/v1/games/hollow-knight)
- green (Go): `go test -tags integration -count=1 -run 'TestAppCatalog$|TestAPIItems' ./tests/` -> ok
- `scripts/openapi-check.sh --write` -> openapi/app-v1.json +1 line (getItem description); api-v1 unchanged
- `go run ./cmd/app codegen:client -prefix /app/v1 -o -` -> api.ts; `bun run --filter @hexascore/app-client typecheck test` -> exit 0
- `golangci-lint run --allow-parallel-runners -j 2 --build-tags integration ./app/appapi/... ./app/controllers/publicapi/... ./routes/... ./tests/...` -> only 2 findings in tests/api_edge_test.go:132/173 (gci, golines), a file this row does not touch (pre-existing on main)
- red (mobile): `bun run --cwd apps/mobile test -- src/navigation.test.tsx -t "bare slug"` -> 1 failed (itemIdOf("arrival", "") was null)
- green (mobile): `bun run --cwd apps/mobile test -- src/navigation.test.tsx` -> 20 passed; `typecheck` ok; `lint` (biome ci) 137 files, no fixes
validation:
- Same route, no new operation: `Controller.SlugRedirects` (set only in routes/app_catalog.go, `catalog.ResolveSlug`) lets getItem take a slug. Live slug -> `SELECT id FROM items WHERE type = $1 AND slug = $2` (bound), then the normal id path (policy, locale, ETag). Renamed slug -> 301 to `/app/v1/{type}/{id}`. Unknown -> 404. Wrong type -> 404. Paid API: SlugRedirects nil, a slug stays 404 (asserted).
- Documented on the app op only (`Op.Description`), so the api-v1 golden is untouched.
- Mobile: a bare slug is sent as the id; fetch follows the 301 and ItemView's existing merge effect moves the query to the uuid key and calls onMerged (router.setParams({id})).
- The regenerated client also picks up the vukem 0.35.0 ApiProblem shape (`code`, extension members): main's committed api.ts was stale against the current codegen, so CI "App client matches codegen" would have failed on main without it.
changelog: Fixed: the Android app opens web item links that carry only the slug.
roadmap: no box (unblocks P5.B.5.1)
risks / follow-ups:
- Three branches of this batch regenerate openapi/app-v1.json and api.ts (FIX-app-item-cover-urls, FIX-app-search-query-params, this one); the second and third merge need `scripts/openapi-check.sh --write` + codegen:client instead of a textual merge.
- tests/api_edge_test.go lint findings on main (gci/golines) belong to whoever last touched it.

## Review (especialista, 2026-11-01)
verdict: CORRIGIDO_E_APROVADO
re-run: `scripts/openapi-check.sh` -> ok (golden regenerated with `--write`, client with `codegen:client -prefix /app/v1`, 23d1177); `go test -tags integration -run 'TestAppCatalog$|TestCoversBaseJSONAbsolute|TestAPITrending|TestAPIItems|TestRouteInventory|TestOpenAPI|TestAppSearch|TestSearch|Suggest|TestAPIEdge' ./tests/` -> ok; golangci-lint (touched pkgs, integration tag) 0 issues; app-client typecheck/test ok; mobile lint/typecheck ok, mobile Jest 297 tests (3 suites timed out under load on the full run, green alone); root lint/typecheck/vitest 129 ok
checked: sdd 05 (paid API keeps no covers, slug lookup app-only), sdd 10 FR-10-7, authz (scopes unchanged), sql (bound params), i18n (catalogues recopied), scope, tests
fixes: 1. conflict with cover-urls resolved: the slug description stays on the getItem op, the handler is the covered wrapper; both test blocks kept 2. mobile branch merged (navigation test conflict: kept the Settings profile label and the empty-slug case) 3. golden + client regenerated (23d1177)
gate on main after merge: local.sh test-go, test-go-integration, test-ts, migrations, vuln, deps, setup-windows, mobile, build PASS on bbe689d/fcf63bf; lint fails only on commitlint of historic local commits; lighthouse fails on main before this batch (FIX-lighthouse-budgets-main)
PRONTO_PARA_MERGE agent/especialista/FIX-app-batch (merged bd95e7d)
