# REV-FIX-export-purge
STATUS: APROVADO (rodada 1). Branch agent/senior/FIX-export-purge (5e0cf29, already in main via fd640f1).
Checked: Build skips when the running-UPDATE hits 0 rows; final ready-UPDATE guarded by status='running', unsettled -> deletes own file, no event, no mail; Purge sweeps Disk.List(<userID>) (Local skips missing dir, S3 empty prefix ok), tolerates ErrNotExist, skips rows whose file failed, joins errors, drops expired rows; PurgeAnonymised retries every anonymised user with leftover rows (closes the "kept for next run" gap I flagged in REV-P1.F.4.1-followup-ui-purge).
Re-run on origin/main: `go test ./...` ok; `go test -tags integration ./...` ok; `golangci-lint run --build-tags integration ./...` 0 issues.
Note: supersedes agent/especialista/P1.F.4.1-followup-ui-purge-fix (same Build race, narrower fix) -> do NOT merge that branch.
