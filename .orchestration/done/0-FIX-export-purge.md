# Task FIX-export-purge (priority 0, rodada 0, model: Sonnet, senior pool)

Objective: close three data-retention gaps (LGPD, SDD 03 deletion / SDD 14 export) found by the automated security review in the merged export code (`app/library/export.go`, `app/jobs/deletion_purge.go`):
1. TOCTOU: `Exports.Purge` hard-deletes `data_exports` rows that a running Build may still finish, so a ZIP of a purged user can be written and mailed afterwards. Fix: Build re-checks inside its final step (`UPDATE ... WHERE id = $1 AND status = 'running'` or `users.anonymized_at IS NULL`); on 0 rows it deletes the file it just wrote and skips the ready mail. Purge also sweeps the user's export directory, not only paths from rows.
2. Purge aborts on the first file error and fails on already-missing files: treat `fs.ErrNotExist` as success, skip/clear `path` for `expired` rows (Expire clears `path`), continue past per-row errors and `errors.Join` them.
3. Deletion purge marks users anonymised even if the export purge failed, and never retries. Fix: purge exports before anonymising (or in the same tx), or have the job also sweep `data_exports` of already-anonymised users; do not cache construction errors in `sync.OnceValues`.
Tests first: an integration test per gap (purge during a running build leaves no file and sends no mail; missing file is fine; failure is retried next run).
App work: branch `agent/<your-role>/FIX-export-purge` from `origin/main`. `TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable`. Acceptance: new tests green; `golangci-lint run --build-tags integration ./...` 0 issues; `go test ./...` and `go test -tags integration ./...`.
Report + `1-REV-FIX-export-purge.md`. Do not wait for CI.
