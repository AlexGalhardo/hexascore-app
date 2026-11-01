# FIX-ingest-trim-titles Wikidata trim keeps title claims and latn aliases
status: concluida
branch: agent/senior/FIX-ingest-trim-titles @ e22078f (stacked on agent/senior/P6.A.3.3 @ 409cc0e)
files: internal/ingest/wikidata/{trim.go,entities_test.go}, internal/ingest/normalize/{relations.go,relations_test.go}, internal/ingest/testdata/{trim_reference.py,parity_based_on.json}
acceptance:
- `go test ./internal/ingest/wikidata ./internal/ingest/normalize` -> ok (new TestTrimKeepsTitleClaimsAndLatnAliases; TestAltTitles now covers original, former, deprecated skip and dedup against locale titles); TestTrimHashMatchesPython ok; golangci-lint 0 issues
decisions:
- Trim keeps P1476, P1448, P2561 as scalar props and aliases of any `*-latn` language besides en/mul/pt-br/pt/es. Reference script trim_reference.py mirrors it (run without `-I`, pydantic lives in user site).
- ExtractRelations alt titles order: transliteration, original (P1476), former (P1448 then P2561), alias; cap 20, dedup unchanged. Lang is empty for original/former (the trimmed monolingualtext keeps only the text, same as the legacy simplify_value).
- HASH CHANGE: the stored hash of every real entity that has any of those claims or a `-latn` alias changes. The four recorded fixtures (Q83495, Q4267401, Q1079, Q208460) carry none of them, so trim_hashes.json regenerated is byte-identical (only CRLF differs from the script output); the Go hash test therefore does not exercise the new data, the unit tests do. The first full Wikidata run MUST use this trim (SDD 06 7.3), otherwise it re-hashes and rewrites records.
risks: parity fixture additions are synthetic (Q900001). No real entity with P1476/-latn recorded [NOT VERIFIED] against live Wikidata.
changelog: Changed: Wikidata trim keeps original title, former names and romanised aliases; normaliser emits original, former and transliteration alt titles.
roadmap: P6.A.3.3 (alt title gap)

## Review (especialista, 2026-11-01)

verdict: APROVADO
reviewed: stack tip agent/senior/P6.B.3.1 @ e113fa6b with main merged (registry conflict resolved in id order), batch branch agent/especialista/P6-batch
checked: go build, vet; `go test ./internal/ingest/...`; `go test -tags integration ./tests/ingest/...` (all), `./app/jobs ./app/community ./app/catalog ./database/...`, `./tests -run 'Route|Schedule|Command|Migrat|Inventory|Bootstrap|Jobs'`; golangci (--build-tags integration, ingest/tests/app/jobs/community/migrations) 0 issues; migration-policy test

Batch also classifies `/.well-known/assetlinks.json` in the route inventory (red on main since P5.B.5.1) and documents the ingest env vars in `.env.example`.

PRONTO_PARA_MERGE agent/especialista/P6-batch
