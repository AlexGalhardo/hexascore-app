# FIX-covers-orphans-on-rollback
status: concluida
branch: agent/senior/FIX-ingest-rollback-api-changes (stacked, second commit)
files: `internal/ingest/covers/job.go`, `app/jobs/ingest.go`, `internal/ingest/covers/job_test.go`, `tests/ingest/covers_test.go`
red: tests first; `go vet ./internal/ingest/covers` failed to build (no `Job.Reserve`), so TestStoreReservesEachKeyBeforeItsPut, TestStoreWritesNothingWhenReserveFails and the integration TestCoversStepRollbackLeavesReservedKeys could not run.
green: `go test ./internal/ingest/covers` ok; `go test -tags integration ./tests/ingest -run TestCovers` ok; `go build ./app/...` ok; golangci-lint (--build-tags integration, covers, tests/ingest, app/jobs) 0 issues.
design: `Job.Reserve` (production: `covers.Reserver(db)`, an INSERT INTO cover_gc ... ON CONFLICT DO NOTHING on a handle outside the step's transaction) runs after encoding and before each Put. If the step's transaction rolls back, the key stays in cover_gc and `covers.GC` deletes the object after 30 days; a reserved key an item or edition points at is forgotten by GC's first statement, so a committed cover is never reaped. A reserve failure aborts that item before any write.
risks: a reserved key of a committed cover keeps a cover_gc row until the next GC pass forgets it; one extra small insert per written variant.
changelog: Fixed (orphaned cover objects after a rolled back covers step).
roadmap: no box (follow-up of P6.B.2.2)
