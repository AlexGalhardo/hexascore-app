# FIX-wikidata-selection-cache
status: concluida
branch: agent/senior/FIX-wikidata-selection-cache (stacked on FIX-ingest-runner-wiring), worktree wt/senior-P6.A.3.5a
files: `internal/ingest/connector.go` (RunParams.RunID), `internal/ingest/fetch/fetch.go` (passes the run id), `internal/ingest/wikidata/select.go` (cache), `internal/ingest/wikidata/select_test.go`
red: `TestTargetsSelectsOncePerRunAndType` failed with `one run must issue one SPARQL query per type, got 12` (before the change every chunk reran the query).
green: `go test ./internal/ingest/...` ok; `golangci-lint --timeout 15m -j 2 --allow-parallel-runners ./internal/ingest/...` 0 issues.
design: the selected QIDs are cached in memory per (run id, type) inside the connector; RunID 0 disables the cache. A 35k movie bootstrap now issues one query per type per process. Bounded to 16 entries (reset on overflow).
risks / follow-ups:
- A restarted worker selects again; for incremental runs without `since` in params, `fetch.Step` recomputes `now-36h`, so the list can shift slightly across a restart. If the cursor QID dropped out of the new list, `Targets` yields nothing and the stage ends early. Persisting `since` in the run params at creation would close it (not done: outside this row).
changelog: Fixed (Wikidata bootstrap no longer reruns the SPARQL selection on every chunk).

## Review (especialista, 2026-11-01)

verdict: APROVADO
reviewed: stack tip agent/senior/P6.B.3.1 @ e113fa6b with main merged (registry conflict resolved in id order), batch branch agent/especialista/P6-batch
checked: go build, vet; `go test ./internal/ingest/...`; `go test -tags integration ./tests/ingest/...` (all), `./app/jobs ./app/community ./app/catalog ./database/...`, `./tests -run 'Route|Schedule|Command|Migrat|Inventory|Bootstrap|Jobs'`; golangci (--build-tags integration, ingest/tests/app/jobs/community/migrations) 0 issues; migration-policy test

Batch also classifies `/.well-known/assetlinks.json` in the route inventory (red on main since P5.B.5.1) and documents the ingest env vars in `.env.example`.

PRONTO_PARA_MERGE agent/especialista/P6-batch
