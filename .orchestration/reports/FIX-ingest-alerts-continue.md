# FIX-ingest-alerts-continue
status: concluida
branch: agent/senior/FIX-ingest-rollback-api-changes (stacked, third commit)
files: `internal/ingest/alerts.go`, `tests/ingest/alerts_test.go`
red: tests first; `go test -tags integration ./tests/ingest -run TestAlertA7` failed in TestAlertA7ContinuesPastAFailingCheck (raised 0, tmdb_age never checked after failed_run errored).
green: same command ok (TestAlertA7Lines and the new test); golangci-lint on internal/ingest and tests/ingest: 0 issues in the touched files (two rowserrcheck findings in untouched resume_test.go and wiring_test.go already existed).
design: `CheckA7` logs state=unknown for a check that cannot run, keeps going and returns the raised count with the errors joined (`errors.Join`, each wrapped with the check name). The only caller (routes/schedule.go) already treats a non-nil error as the job result.
changelog: Fixed (one failing A7 check no longer skips the remaining checks).
roadmap: no box (follow-up of P6.B.4.5)

## Review (especialista, 2026-11-01)
verdict: CORRIGIDO_E_APROVADO (fix is the shared lint commit on the stacked branch)
re-run: `go test -tags integration -p 1 ./tests/ingest -run 'TestRollback|TestNormalize|TestCovers|TestAlertA7|TestRecompute|TestLocked|TestSlug|TestCancelled|TestBad|TestResume|TestWiring'` → ok (32 s); `go test ./internal/ingest/covers ./internal/ingest/normalize` → ok; `golangci-lint run --timeout 15m -j 2 --build-tags integration ./tests/ingest/... ./internal/ingest/... ./app/jobs/...` → 0 issues; `golangci-lint fmt --diff` → clean
fixes (CORRIGIDO): 1. rows.Err() checks in tests/ingest/resume_test.go and wiring_test.go, the two rowserrcheck findings (8d1a251c)
gate on main after merge: go build ok; targeted tests and lint above (identical tree); full integration suite not run (host short on memory), left to the batch gate
PRONTO_PARA_MERGE agent/especialista/FIX-ingest-rollback-api-changes-fix (merge 9c366902)
checked: sdd 12 §6.2 (A7 lines), scope ok, the only caller (routes/schedule.go) treats a non-nil error as the job result, tests red→green
