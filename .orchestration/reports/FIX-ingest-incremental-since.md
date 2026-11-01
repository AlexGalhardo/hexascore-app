# FIX-ingest-incremental-since
status: concluida
branch: agent/senior/FIX-ingest-incremental-since (stacked on P6.B.1.1), worktree wt/senior-P6.A.3.5a
files: `internal/ingest/chains.go`, `internal/ingest/fetch/fetch.go`, `internal/ingest/chains_test.go`
red: `go test ./internal/ingest -run Since` failed to build (`undefined: IncrementalWindow`; the behaviour test would also have failed: the params held no `since`).
green: `go test ./internal/ingest` ok (4 new tests: since persisted, explicit since kept, successor stage reuses it, full-mode stage gets none); golangci-lint (-j 2) 0 issues.
design: `Chain.spec` writes `since = now-36h` (UTC) into the params of an incremental stage when the params have none. `Advance` copies the params to the next stage, so the whole chain and any resume share one cursor. `ingest.IncrementalWindow` now lives in `ingest`; `fetch.IncrementalWindow` aliases it and stays as the fallback for runs not created by a chain.
not checked: integration tests (tag `integration`, need Postgres) were not run; the change touches no SQL.
changelog: Fixed (an incremental run no longer drifts its window across a worker restart).

## Review (especialista, 2026-11-01)

verdict: APROVADO
reviewed: stack tip agent/senior/P6.B.3.1 @ e113fa6b with main merged (registry conflict resolved in id order), batch branch agent/especialista/P6-batch
checked: go build, vet; `go test ./internal/ingest/...`; `go test -tags integration ./tests/ingest/...` (all), `./app/jobs ./app/community ./app/catalog ./database/...`, `./tests -run 'Route|Schedule|Command|Migrat|Inventory|Bootstrap|Jobs'`; golangci (--build-tags integration, ingest/tests/app/jobs/community/migrations) 0 issues; migration-policy test

Batch also classifies `/.well-known/assetlinks.json` in the route inventory (red on main since P5.B.5.1) and documents the ingest env vars in `.env.example`.

PRONTO_PARA_MERGE agent/especialista/P6-batch
