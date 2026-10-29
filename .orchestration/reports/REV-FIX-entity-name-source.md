REV FIX-entity-name-source: APROVADO (merged e1290f4). Branch: agent/senior/FIX-entity-name-source.
Checked: Assemble keeps the relation, drops name when api policy disallows NameSource (empty = fail closed); site policy Allows everything so site unchanged; name_source CASE mirrors api_entities.go; meta.omittedFields gets `entities.name`; resolves the pending note from REV-P1.H.2.1b.
Re-run on origin/main 2b194db: go test -tags integration ./tests -run 'TestAPILicenceFilter|TestAPIItems|TestAPIEntities|TestAPIContract' -> ok (no SKIP); golangci-lint --build-tags integration -> 0 issues; go test ./... and -tags integration ./... -> all ok; bun run test -> 57 passed.
Risk: EntityRef.name now optional in OpenAPI (documented by the author).
