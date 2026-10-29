STATUS: concluida
Branch: agent/senior/FIX-entity-name-source (app/catalog/{readmodel,api_items}.go, audience_test.go, tests/api_items_test.go, CHANGELOG.md)
Change: EntityRefIn.NameSource (translation name_source / entities.provenance default_name); Assemble omits the name (relation kept) when the api policy disallows/unknown source, lists `entities.name` in meta.omittedFields; EntityRef.name now omitempty. Site unchanged.
Tests: fixture allowed relation + IGDB-sourced name + unsourced name -> absent on /api/v1, present in site read model (TestAPILicenceFilter).
Acceptance: go test -tags integration ./tests -run 'TestAPILicenceFilter|TestAPIItems|TestAPIEntities|TestAPIContract' -> ok; golangci-lint run --build-tags integration ./... -> 0 issues; go test -tags integration ./... -> all ok; bun run test -> 53 passed.
Risks: new omittedFields token `entities.name` (API consumers); OpenAPI schema is derived from Go types (name now optional).
