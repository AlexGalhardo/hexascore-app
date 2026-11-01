# FIX-app-item-cover-urls cover URLs on /app/v1 items
status: concluida
branch: agent/senior/FIX-app-item-cover-urls @ 9894a02 (from main 6dcfa84)
files: app/appapi/catalog.go, routes/app_catalog.go, app/controllers/publicapi/*.go (notModified exported as NotModified, no behaviour change), openapi/app-v1.json, packages/app-client/src/api.ts, tests/app_catalog_test.go, tests/covers_base_test.go
acceptance:
- red: `go test -tags integration -count=1 -run 'TestAppCatalog$|TestCoversBaseJSONAbsolute' ./tests/` -> FAIL (no coverThumbUrl in /app/v1/games/{id}); after the first green, an extra ETag assertion was red: `app_catalog_test.go:100: status = 304, want 200` (old ETag still matched after a cover change)
- green: `go test -tags integration -count=1 -run 'TestAppCatalog$|TestCoversBaseJSONAbsolute|TestAPITrending|TestAPIItems|TestRouteInventory|TestOpenAPI' ./tests/` -> ok
- `scripts/openapi-check.sh --write` -> only openapi/app-v1.json changed; openapi/api-v1.json byte-identical
- `go run ./cmd/app codegen:client -prefix /app/v1 -o -` (APP_ENV=development, APP_KEY, DATABASE_URL as ci.yml) -> packages/app-client/src/api.ts; `bun run --filter @hexascore/app-client typecheck test` -> 2 pass
- `golangci-lint run --allow-parallel-runners -j 2 --build-tags integration ./app/appapi/... ./app/controllers/publicapi/... ./routes/... ./tests/...` -> 0 issues; `golangci-lint fmt` applied
validation:
- Design as approved: app-only `AppItemJSON{publicapi.ItemJSON; CoverThumbURL; CoverURL}` (`omitempty`), `AppItemPageJSON`, `AppTrendingEntryJSON`, `AppTrendingJSON`. Embedded structs flatten in the schema, as https://vukem.dev/docs/rest-apis/openapi#go-types says ("Embedded structs are flattened into their parent, as encoding/json does"); the golden confirms AppItemJSON lists every ItemJSON property plus the two cover fields. No VF entry needed. The trending entry uses explicit fields (shadowing `Item` through an embedded struct is not documented).
- Wrapped handlers: getItem, listItems, listEntityItems, the four trending reads. `catalog.CoversFor` + `Cover.WithBase(CoversBase(COVERS_BASE_URL))`, the same rule as lists and search; with an empty base the URL is the same-origin `/covers/...` path (production setting).
- Conditional requests: the publicapi ETag hashes the item without covers, so the wrapper hides If-None-Match/If-Modified-Since from the inner call and sets the ETag of the decorated body (`publicapi.NotModified`). Tested: an ETag taken before a cover change answers 200 with the new cover.
- Paid API: `/api/v1/games/{id}`, `/api/v1/games`, `/api/v1/games/trending` asserted without cover fields.
changelog: Fixed: /app/v1 catalogue items carry coverThumbUrl and coverUrl (app audience only).
roadmap: no box
risks / follow-ups:
- `/app/v1/{type}/trending` Cache-Control changes from `public, max-age=300` to `private, no-cache` (the shared NotModified helper); for a bearer-authenticated read this is the safer value. Say so if the app relies on the 300 s cache.
- Schema names in the app client changed (ItemJSON -> AppItemJSON etc.); no TS code imports them (grep of apps/, packages/, resources/js).
- Mobile can now pass `item.coverThumbUrl` to ItemRow (impl-mobile, out of scope here).

## Review (especialista, 2026-11-01)
verdict: CORRIGIDO_E_APROVADO
re-run: `scripts/openapi-check.sh` -> ok (golden regenerated with `--write`, client with `codegen:client -prefix /app/v1`, 23d1177); `go test -tags integration -run 'TestAppCatalog$|TestCoversBaseJSONAbsolute|TestAPITrending|TestAPIItems|TestRouteInventory|TestOpenAPI|TestAppSearch|TestSearch|Suggest|TestAPIEdge' ./tests/` -> ok; golangci-lint (touched pkgs, integration tag) 0 issues; app-client typecheck/test ok; mobile lint/typecheck ok, mobile Jest 297 tests (3 suites timed out under load on the full run, green alone); root lint/typecheck/vitest 129 ok
checked: sdd 05 (paid API keeps no covers, slug lookup app-only), sdd 10 FR-10-7, authz (scopes unchanged), sql (bound params), i18n (catalogues recopied), scope, tests
fixes: 1. merged in one batch with the two other app-v1 rows; golden and client regenerated instead of a text merge (23d1177) 2. mobile code still imported the old ItemJSON/TrendingEntryJSON/TrendingJSON names: renamed to the App* schemas and home/listing/item now pass coverThumbUrl/coverUrl (4286dab)
gate on main after merge: local.sh test-go, test-go-integration, test-ts, migrations, vuln, deps, setup-windows, mobile, build PASS on bbe689d/fcf63bf; lint fails only on commitlint of historic local commits; lighthouse fails on main before this batch (FIX-lighthouse-budgets-main)
PRONTO_PARA_MERGE agent/especialista/FIX-app-batch (merged bd95e7d)
