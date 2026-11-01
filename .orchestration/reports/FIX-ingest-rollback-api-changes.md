# FIX-ingest-rollback-api-changes
status: concluida
branch: agent/senior/FIX-ingest-rollback-api-changes (worktree wt/senior-fw-rollback, from local main 065cb6d6)
files: `internal/ingest/rollback/rollback.go`, `internal/ingest/normalize/{job.go,changes.go}`, `tests/ingest/{rollback_test.go,normalize_job_test.go}`
red: tests first; `go test -tags integration ./tests/ingest -run 'TestRollbackWritesAPIChanges|TestNormalizeWritesAPIChanges'` failed (0 api_changes rows for 2 deleted/hidden items, 0 created, no updated row).
green: same command ok; `-run 'TestRollback|TestNormalize|TestRecompute|TestLocked|TestSlug|TestCancelled|TestBad'` ok; golangci-lint (--build-tags integration, normalize, rollback, tests/ingest) 0 issues.
design: rollback writes a `deleted` row for every item it deletes or hides (a hidden item leaves the public API the same way). Normalize snapshots the API-visible values of an existing item (slug, wikidata, releaseDate, runtimeMinutes, pageCount, externalIds, title and synopsis per locale) before and after `saveItem` and writes `created` for a new item or `updated` with the differing field names; an unchanged re-run writes nothing. Rows are written inside the item savepoint, so a failed item leaves none.
not done (outside the asked scope, still in the row text): a non-wikidata rollback does not requeue the Wikidata record or clear that source's cover; restored IGDB/TMDB payloads keep the bad run's `checked_at`; `entities` (credits) changes are not part of the normalize snapshot, so a credits-only change writes no row.
changelog: Fixed (ingestion and rollback now feed /changes).
roadmap: no box (follow-up of P6.B)
