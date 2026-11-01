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

## Review (especialista, 2026-11-01)
verdict: CORRIGIDO_E_APROVADO (fix is the shared lint commit on the stacked branch)
re-run: `go test -tags integration -p 1 ./tests/ingest -run 'TestRollback|TestNormalize|TestCovers|TestAlertA7|TestRecompute|TestLocked|TestSlug|TestCancelled|TestBad|TestResume|TestWiring'` → ok (32 s); `go test ./internal/ingest/covers ./internal/ingest/normalize` → ok; `golangci-lint run --timeout 15m -j 2 --build-tags integration ./tests/ingest/... ./internal/ingest/... ./app/jobs/...` → 0 issues; `golangci-lint fmt --diff` → clean
fixes (CORRIGIDO): 1. rows.Err() checks in tests/ingest/resume_test.go and wiring_test.go, the two rowserrcheck findings (8d1a251c)
gate on main after merge: go build ok; targeted tests and lint above (identical tree); full integration suite not run (host short on memory), left to the batch gate
PRONTO_PARA_MERGE agent/especialista/FIX-ingest-rollback-api-changes-fix (merge 9c366902)
checked: sdd 06 covers / ADR 0014 GC, authz n/a, secrets n/a, sql bound, scope ok, tests red→green
notes: writing the reservation outside the step transaction is the point of the fix and is safe: `Step` runs `j.Run` (Store, Reserve) before `apply`, so the step's own trigger inserts into `cover_gc` cannot hold a row lock the reserve waits on (no app-level self-deadlock); a committed cover's reservation is dropped by GC's first statement; the trigger's `DO UPDATE SET released_at` refreshes the grace when a reserved key is later released. Residual risk: each reserve needs a second pool connection while the step's tx is open (pool size must stay >= 2).
