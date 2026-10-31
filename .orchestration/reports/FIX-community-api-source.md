# FIX-community-api-source community provenance on the public API

status: concluida
branch: agent/senior/FIX-community-api-source @ 931210f (stacked on agent/senior/P4.B.2.3 @ f113594)
files: `config/catalog.go`, `app/catalog/write.go` (EnsureSource licence), `app/catalog/audience_test.go`, `tests/api_items_test.go`, `docs/sdd/01-catalog.md`
policy: included. SDD 01 §6.1 ("HexaScore-sourced fields (`hexascore` and `community` sources ...) reach API customers under the API terms (A-07)") and SDD 05 §5 field table ("wikidata (CC0), community (A-07)") say community edits are API-eligible. `DefaultCatalogAPISources` now ends with `community`; the `data_sources` row inserted on first apply carries licence `HEXASCORE-API-TERMS` (was `proprietary`, which contradicted A-07 in `provenance.*.licence`).
acceptance:

- `go test ./app/catalog/ -run TestVisibilityCommunityReachesAPI` → red before (`got title ""`), `ok` after; also asserts an allowlist without `community` hides the field
- `go test -tags integration ./tests -run 'TestAPICommunityProvenance|TestAPILicenceFilter|TestAPIContract'` → `ok` (7.7 s); without the config change `TestAPICommunityProvenance` fails with `JSON path "runtimeMinutes" not found`

validation: go build ok, `golangci-lint run --build-tags integration ./config/... ./app/catalog/... ./tests/...` 0 issues, markdownlint ok
changelog: Fixed: approved community edits are served on `/api/v1` (default `CATALOG_API_SOURCES` includes `community`, licence `HEXASCORE-API-TERMS`)
roadmap: no box
risks / follow-ups:

- `app/catalog/write.go` is also in the P4.B.2.1 lock set (my stacked chain); one-line change.
- Existing databases that already have a `community` row with licence `proprietary` keep it (`ON CONFLICT DO NOTHING`); no prod rows exist yet since P4.B is unreleased. If staging has one, `UPDATE data_sources SET licence='HEXASCORE-API-TERMS' WHERE code='community'`.
- Search: `search_tsv_api` indexes only `wikipedia` synopses (migration 20261029140003); `hexascore` and now `community` synopses are not API-searchable. Same as before for `hexascore`; a migration is needed if that should change (SDD 07 07-7).
- Production env: if Railway sets `CATALOG_API_SOURCES` explicitly, append `community` there (owner).

## Review (especialista, 2026-10-31)
verdict: APROVADO
re-run (main + branch): `go test ./app/catalog/ -run TestVisibilityCommunityReachesAPI` → ok; `go test -tags integration ./tests -run 'TestAPICommunityProvenance|TestAPILicenceFilter|TestAPIContract|TestAPIItems'` → ok; golangci (integration tag) config, catalog, tests → 0 issues; markdownlint SDD 01 → 0
checked: policy against SDD 01 §6.1 (community listed with hexascore under A-07) and SDD 05 §5; licence string matches the score licence constant; no env file pins CATALOG_API_SOURCES in the repo; scope, tests
gate on main after merge aa581af: go build ok
PRONTO_PARA_MERGE agent/senior/FIX-community-api-source
