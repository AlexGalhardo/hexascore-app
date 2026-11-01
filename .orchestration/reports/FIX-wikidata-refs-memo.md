# FIX-wikidata-refs-memo
status: concluida
branch: agent/senior/FIX-wikidata-refs-memo (stacked on FIX-wikidata-selection-cache), worktree wt/senior-P6.A.3.5a
files: `internal/ingest/wikidata/entities.go`, `internal/ingest/wikidata/select.go`, `internal/ingest/wikidata/entities_test.go`
red: `TestFetchReemitsEntityRecordsWhenTheChunkIsRetried` failed with `retried chunk emitted 0 entity records, want 13`.
green: `go test ./internal/ingest/wikidata` ok; golangci-lint (-j 2) 0 issues.
design: the `refs` memo is created per `Fetch` call (one call per 200-target chunk, one transaction) and passed to `resolveRefs`; the `Connector.refs` field is gone. A retried chunk re-emits its entity records. Cost: an entity referenced by two chunks is fetched twice; the upsert hash check keeps the second a no-op write.
changelog: Fixed (Wikidata entity records no longer lost when a chunk rolls back).

## Review (especialista, 2026-11-01)

verdict: APROVADO
reviewed: stack tip agent/senior/P6.B.3.1 @ e113fa6b with main merged (registry conflict resolved in id order), batch branch agent/especialista/P6-batch
checked: go build, vet; `go test ./internal/ingest/...`; `go test -tags integration ./tests/ingest/...` (all), `./app/jobs ./app/community ./app/catalog ./database/...`, `./tests -run 'Route|Schedule|Command|Migrat|Inventory|Bootstrap|Jobs'`; golangci (--build-tags integration, ingest/tests/app/jobs/community/migrations) 0 issues; migration-policy test

Batch also classifies `/.well-known/assetlinks.json` in the route inventory (red on main since P5.B.5.1) and documents the ingest env vars in `.env.example`.

PRONTO_PARA_MERGE agent/especialista/P6-batch
