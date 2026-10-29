STATUS: concluida
Branch: agent/senior/FIX-export-purge (pushed). Files: app/library/export.go, app/jobs/deletion_purge.go, tests/export_test.go.
1. Build: initial UPDATE 0 rows -> skip; final UPDATE ... AND status='running', 0 rows -> delete own file, no event, no mail.
2. Purge: sweeps Disk.List(<userID>), ErrNotExist = success, skips expired rows' path, continues past errors (errors.Join), rows with failed file kept.
3. New Exports.PurgeAnonymised (all anonymised users with leftover rows); job calls it every run; sync.OnceValues removed (no cached construction error).
Acceptance: golangci-lint run --build-tags integration ./... -> 0 issues; go test ./... ok; go test -tags integration ./... ok.
New tests: TestExportBuildAfterPurgeLeavesNoFileAndNoMail, TestExportPurgeToleratesMissingFilesAndExpiredRows, TestExportPurgeFailureIsRetriedForAnonymisedUsers.
Risks: Purge during a running Build with the file written after the List sweep is covered by Build's own delete (final UPDATE guard); file is briefly on disk.
