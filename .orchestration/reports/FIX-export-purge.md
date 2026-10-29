STATUS: concluida
Branch: agent/senior/FIX-export-purge (pushed). Files: app/library/export.go, app/jobs/deletion_purge.go, tests/export_test.go.
1. Build: initial UPDATE 0 rows -> skip; final UPDATE ... AND status='running', 0 rows -> delete own file, no event, no mail.
2. Purge: sweeps Disk.List(<userID>), ErrNotExist = success, skips expired rows' path, continues past errors (errors.Join), rows with failed file kept.
3. New Exports.PurgeAnonymised (all anonymised users with leftover rows); job calls it every run; sync.OnceValues removed (no cached construction error).
Acceptance: golangci-lint run --build-tags integration ./... -> 0 issues; go test ./... ok; go test -tags integration ./... ok.
New tests: TestExportBuildAfterPurgeLeavesNoFileAndNoMail, TestExportPurgeToleratesMissingFilesAndExpiredRows, TestExportPurgeFailureIsRetriedForAnonymisedUsers.
Risks: Purge during a running Build with the file written after the List sweep is covered by Build's own delete (final UPDATE guard); file is briefly on disk.
# REV-FIX-export-purge
STATUS: APROVADO (rodada 1). Branch agent/senior/FIX-export-purge (5e0cf29, already in main via fd640f1).
Checked: Build skips when the running-UPDATE hits 0 rows; final ready-UPDATE guarded by status='running', unsettled -> deletes own file, no event, no mail; Purge sweeps Disk.List(<userID>) (Local skips missing dir, S3 empty prefix ok), tolerates ErrNotExist, skips rows whose file failed, joins errors, drops expired rows; PurgeAnonymised retries every anonymised user with leftover rows (closes the "kept for next run" gap I flagged in REV-P1.F.4.1-followup-ui-purge).
Re-run on origin/main: `go test ./...` ok; `go test -tags integration ./...` ok; `golangci-lint run --build-tags integration ./...` 0 issues.
Note: supersedes agent/especialista/P1.F.4.1-followup-ui-purge-fix (same Build race, narrower fix) -> do NOT merge that branch.
PRONTO_PARA_MERGE agent/senior/FIX-export-purge (already merged)
