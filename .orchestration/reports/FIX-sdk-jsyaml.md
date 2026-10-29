# FIX-sdk-jsyaml report
STATUS: parcial (js-yaml part bloqueada; redirect part concluida)
Branch: agent/senior/P2.D.5.2 (origin/main + especialista/P2.D.5.1-fix merged), pushed.
BLOCKED: `npm view @hey-api/openapi-ts version` -> 0.99.0 (already pinned, latest); its dep @hey-api/json-schema-ref-parser 1.4.4 (latest) pins js-yaml 4.2.0 exactly. No upgrade path. Options for orchestrator: (a) bun `overrides` js-yaml 4.3.2 in sdks/ts/package.json (dev-only codegen), (b) documented exception in docs/runbooks/dependencies.md. I did neither.
Done: /countries/{code} and /languages/{code} match code case-insensitively, 301 to canonical code (keeps locale prefix and query). Test in tests/entity_show_test.go: `go test -tags integration ./tests -run TestInertiaPages_Entity_Show` -> ok.
Checks: golangci-lint (integration tag, app+tests), bun run test pass.
