# FIX-ingest-override-seam
status: concluida
branch: agent/senior/FIX-ingest-override-seam (stacked on FIX-wikidata-refs-memo), worktree wt/senior-P6.A.3.5a
files: `app/jobs/ingest.go`, `app/jobs/registry.go`, `bootstrap/app.go`, `tests/ingest/wiring_test.go`
red: wiring_test switched to `bootstrap.NewWithIngest(src, opts...)` first; `go vet -tags integration ./tests/ingest` failed with `undefined: bootstrap.NewWithIngest`.
green: `TestBootstrapThenQueueWorkNormalises`, `TestWikipediaChainRunsEveryEdition`, `TestOpenLibraryStageAnswersSkippedWhileUnlicensed`, `TestWeeklyWikidataRefreshesStoredEntities` pass (13.5 s, DB on 5433); golangci-lint (integration tag, -j 2) 0 issues.
design: no package state. `jobs.AllWith(app, IngestSources)` is the registry; `jobs.All` is `AllWith(app, IngestSources{})`. `bootstrap.New`/`NewWithSSRFailures` keep their signatures and call a private `build`; `bootstrap.NewWithIngest(src, opts...)` is the test constructor. Production code path is unchanged (zero value).
changelog: none (test seam only).

## Review (especialista, 2026-11-01)

verdict: APROVADO
reviewed: stack tip agent/senior/P6.B.3.1 @ e113fa6b with main merged (registry conflict resolved in id order), batch branch agent/especialista/P6-batch
checked: go build, vet; `go test ./internal/ingest/...`; `go test -tags integration ./tests/ingest/...` (all), `./app/jobs ./app/community ./app/catalog ./database/...`, `./tests -run 'Route|Schedule|Command|Migrat|Inventory|Bootstrap|Jobs'`; golangci (--build-tags integration, ingest/tests/app/jobs/community/migrations) 0 issues; migration-policy test

Batch also classifies `/.well-known/assetlinks.json` in the route inventory (red on main since P5.B.5.1) and documents the ingest env vars in `.env.example`.

PRONTO_PARA_MERGE agent/especialista/P6-batch
