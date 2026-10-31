# FIX-export-purge-zip
status: concluida (no change needed)
branch: none (worktree and branch removed)
finding: cherry-pick of 4750ae7 conflicts with main, which already holds an equivalent, stricter fix in app/library/export.go (UPDATE ... AND status = 'running', `settled` flag, `deleteFile` when the row is gone) and tests TestExportBuildAfterPurgeLeavesNoFileAndNoMail / TestExportPurgeToleratesMissingFilesAndExpiredRows in tests/export_test.go.
changelog: none
roadmap: no box
risks: none
