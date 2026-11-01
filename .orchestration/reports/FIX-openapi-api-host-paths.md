# FIX-openapi-api-host-paths OpenAPI documents the /v1 api-host form
status: concluida (code part); goldens and SDK regeneration pending, specialist on Windows
branch: agent/senior/FIX-openapi-api-host-paths @ 0902e0b (from main 499b911)
files: bootstrap/app.go, tests/api_edge_test.go, tests/api_reference_test.go, docs/sdd/05-public-api.md, vukem-feedback.md, sdks/ts/src/index.ts, sdks/ts/src/index.test.ts, sdks/ts/README.md, sdks/python/README.md, sdks/python/tests/test_client.py
acceptance:
- `go test -tags integration ./tests -run 'TestAPIEdge|TestAPIReference|TestAPIContract'` -> ok. New subtest `TestAPIEdge/openapi_urls_resolve`: every documented operation, requested as server URL + path on the api host, answers the same status as /api/v1 + path and never 405; `/games` and `/me/usage` return 200. Mutation check: with StripPrefix removed the test fails.
- `go run ./cmd/app openapi:export -prefix /api/v1` -> servers `https://api.hexascore.app/v1`, 26 paths (same count) now `/games`, `/books/{id}/editions`...; `-prefix /app/v1` unchanged (41 paths, full `/app/v1` paths).
- `golangci-lint run --allow-parallel-runners --build-tags integration ./bootstrap/... ./tests/...` -> 0 issues
decisions:
- `bootstrap.apiInfo`: Info.Servers = https://<API_HOST>/v1 and Info.StripPrefix = /api/v1 (vukem v0.34.0 feature). The /app/v1 document needs full paths and Info is app-wide, so the command line decides (any arg `/app/v1` or `=/app/v1`, covering openapi:export and codegen:client). Framework gap logged as VF-NEW-OPENAPI-DOC-INFO with a workaround(vukem) comment.
- Served /api/v1/openapi.json (Scalar) changes the same way; Scalar try-it now targets https://api.hexascore.app/v1/...
- SDD 05 section 6.1 rewritten to the new contract. Hand-written SDK bits moved to base URL https://api.hexascore.app/v1 (ts HEXASCORE_BASE_URL, ts tests expect /games, python test BASE_URL + /v1/games, READMEs).
risks: NOT DONE (left to the specialist on Windows): (1) `scripts/openapi-check.sh --write` to regenerate openapi/api-v1.json (and confirm app-v1.json is byte-identical), (2) `bun run sdk:generate` in sdks/ts and `bun run sdk:generate:python` to regenerate sdks/ts/src/gen and sdks/python/hexascore, (3) then `bun test` in sdks/ts and pytest in sdks/python, (4) packages/app-client does not change. Until then the openapi:check CI job fails (golden drift) and the sdk tests I edited fail against the old generated code (they pass only with the regenerated one). Published SDK consumers with the old base URL https://api.hexascore.app would get /games at the root, which 404s: SDK release notes should say base URL changes to /v1 (owner call on versioning).
changelog: Changed: the OpenAPI document and SDKs address the public API as https://api.hexascore.app/v1/..., matching the host form.
roadmap: none (FIX row)

## Review (especialista, 2026-10-31)

verdict: APROVADO
re-run: go test ./tests -run TestAPIEdge|TestAPIReference (branch); on main: scripts/openapi-check.sh --write (app-v1 changes only by /app/v1/me/for-you from P6.D.2.2), bun run sdk:generate + sdks/ts typecheck and 3 tests ok, openapi-python-client 0.29.1 in uv 0.9.30 + pytest 2 passed, codegen:client + app-client typecheck/test ok (116e4df)
checked: sdd 05 6.1; old /api/v1 paths on the api host still answer (TestAPIEdge asserts http://api.hexascore.app/api/v1/games 200)
vukem-feedback.md conflict resolved (both entries kept)
gate on main after merge: see batch gate in handoff/especialista.md (scripts/ci/local.sh on 127eeeb)

PRONTO_PARA_MERGE agent/senior/FIX-openapi-api-host-paths (merged 9641b99)
