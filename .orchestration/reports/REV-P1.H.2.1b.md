# REV-P1.H.2.1b
VERDICT: CORRIGIDO_E_APROVADO. Final branch agent/especialista/P1.H.2.1b-fix (b6a02ac, from origin/main; 1b already merged).
Findings fixed (licence leak, SDD 01 §6.1 "entity relations: only rows whose source is allowed"):
1. listEntityItems, listItems genre/platform filters and listGenres counts matched item_entities of ANY source: a client could enumerate IGDB/TMDB relations getItem hides (fixture genre relation is igdb and was listed). Fix: Policy.RelationSources() (nil site / allowlist API), ListParams.RelationSources applied in itemPage, Facets(..., sources). Site unchanged (nil).
2. LoadAPIEntity labelled default_name without provenance as 'hexascore' (fail open, contradicts own report). Now omitted.
Tests: api_entities_test (genre entity items + games?genre= empty; studio name omitted), api_items_test (games/genres empty). Red without fix: 2 failures.
Acceptance (re-run): `go test -count=1 -tags integration ./tests -run 'TestAPIEntities|TestAPIContract|TestAPIItems|TestAPILicenceFilter|TestCatalog'` -> ok; `go test -tags integration ./app/catalog/... ./tests/...` -> all ok; `golangci-lint run --build-tags integration ./...` -> 0 issues.
listGenres per-type operationIds: OK (SDD 05 already amended by orchestrator).
Not fixed (note): addEntities shows coalesce(t.name, te.name, default_name) under the relation source, not the name source (pre-existing P1.H.2.1).
Merge: CHANGELOG [Unreleased] Fixed conflicts trivially with P1.H.2.1c-fix (keep both lines).
