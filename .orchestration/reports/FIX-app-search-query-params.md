# FIX-app-search-query-params query parameters on /app/v1/search and /search/suggest
status: concluida
branch: agent/senior/FIX-app-search-query-params @ 9e0d968 (from main 8a73a3c)
files: app/appapi/search.go, app/controllers/search_controller.go, openapi/app-v1.json, packages/app-client/src/api.ts, tests/app_search_openapi_test.go (new), tests/app_search_test.go
acceptance:
- red: `go test -count=1 -run TestAppSearchOpenAPIQuery ./tests/` -> FAIL (`/app/v1/search/suggest: query parameter "type" missing (got [])`, `200 has no JSON schema`, same for /app/v1/search)
- green: same command -> ok; `go test -tags integration -count=1 -run 'TestAppSearch|TestCoversBaseJSONAbsolute|TestSearch|TestSecurity.*Inventory|TestRouteInventory|Suggest' ./tests/` -> ok
- `scripts/openapi-check.sh --write` -> only openapi/app-v1.json changed
- `go run ./cmd/app codegen:client -prefix /app/v1 -o -` -> packages/app-client/src/api.ts: `searchIndex(args?: { query?: SearchIndexQuery }): Promise<SearchIndexResponse>`, same for searchSuggest; `bun run --filter @hexascore/app-client typecheck test` -> exit 0; `bun run --cwd apps/mobile typecheck` -> ok
- `golangci-lint run --allow-parallel-runners -j 2 --build-tags integration ./app/appapi/... ./app/controllers/ ./tests/...` -> 0 issues
validation:
- Plain routes only get a generic entry; GET input fields of a typed route become query parameters (https://vukem.dev/docs/rest-apis/openapi, "How operations are built"). So the two app reads are `api.Register` routes with `SearchInput{q,type,locale,page}` / `SuggestInput{q,type,locale}`; operationIds `search.index` / `search.suggest` kept (client method names unchanged).
- The inputs are string-typed and documentation only: the controller still parses the raw query (`searchParams`, `QueryInt("page", 1)`), so `page=x` stays a default (asserted 200), q > 100 runes stays 400, timeout stays 503 + Retry-After.
- `Search.AppIndex` became `AppSearch` returning `AppSearchJSON` (empty data is `[]`, not null). `Search.Suggest` (web `/search/suggest`) now wraps `SuggestRead`; body and ETag are the same JSON. The typed 304 uses the existing VF-049 workaround (`errSuggestNotModified`); asserted: anonymous app suggest with the ETag -> 304.
- Op.Security left empty: auth is optional on these routes, a documented 401 would be wrong.
changelog: Fixed: the app API documents the search query parameters, so the generated client can send q.
roadmap: no box
risks / follow-ups:
- FIX-app-item-cover-urls (9894a02) also regenerates openapi/app-v1.json and packages/app-client/src/api.ts. Both branched from main as asked; whichever merges second needs `scripts/openapi-check.sh --write` plus the codegen:client step instead of a textual merge of the generated files.
- IMPL-mobile-v1-search-wire can now call `searchSuggest({ query: { q } })`.

## Review (especialista, 2026-11-01)
verdict: CORRIGIDO_E_APROVADO
re-run: `scripts/openapi-check.sh` -> ok (golden regenerated with `--write`, client with `codegen:client -prefix /app/v1`, 23d1177); `go test -tags integration -run 'TestAppCatalog$|TestCoversBaseJSONAbsolute|TestAPITrending|TestAPIItems|TestRouteInventory|TestOpenAPI|TestAppSearch|TestSearch|Suggest|TestAPIEdge' ./tests/` -> ok; golangci-lint (touched pkgs, integration tag) 0 issues; app-client typecheck/test ok; mobile lint/typecheck ok, mobile Jest 297 tests (3 suites timed out under load on the full run, green alone); root lint/typecheck/vitest 129 ok
checked: sdd 05 (paid API keeps no covers, slug lookup app-only), sdd 10 FR-10-7, authz (scopes unchanged), sql (bound params), i18n (catalogues recopied), scope, tests
fixes: 1. golden + client regenerated after the cover-urls merge (23d1177)
gate on main after merge: local.sh test-go, test-go-integration, test-ts, migrations, vuln, deps, setup-windows, mobile, build PASS on bbe689d/fcf63bf; lint fails only on commitlint of historic local commits; lighthouse fails on main before this batch (FIX-lighthouse-budgets-main)
PRONTO_PARA_MERGE agent/especialista/FIX-app-batch (merged bd95e7d)
