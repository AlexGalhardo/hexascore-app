# FIX-ingest-alerts-continue
status: concluida
branch: agent/senior/FIX-ingest-rollback-api-changes (stacked, third commit)
files: `internal/ingest/alerts.go`, `tests/ingest/alerts_test.go`
red: tests first; `go test -tags integration ./tests/ingest -run TestAlertA7` failed in TestAlertA7ContinuesPastAFailingCheck (raised 0, tmdb_age never checked after failed_run errored).
green: same command ok (TestAlertA7Lines and the new test); golangci-lint on internal/ingest and tests/ingest: 0 issues in the touched files (two rowserrcheck findings in untouched resume_test.go and wiring_test.go already existed).
design: `CheckA7` logs state=unknown for a check that cannot run, keeps going and returns the raised count with the errors joined (`errors.Join`, each wrapped with the check name). The only caller (routes/schedule.go) already treats a non-nil error as the job result.
changelog: Fixed (one failing A7 check no longer skips the remaining checks).
roadmap: no box (follow-up of P6.B.4.5)
