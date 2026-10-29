# REV-P1.F.4.1-followup-ui-purge
STATUS: CORRIGIDO_E_APROVADO (rodada 1). Reviewed 5c86d73 (in main). Final branch: agent/especialista/P1.F.4.1-followup-ui-purge-fix (4750ae7, from origin/main).
Fix: race Build vs Purge: Build wrote the ZIP then UPDATEd a row Purge had deleted (0 rows, no error) -> orphan ZIP with personal data that Expire never sees (LGPD, SDD 14 §8). Build now checks RowsAffected and deletes the file. New TestExportBuildAfterPurgeLeavesNoFile (Disk wrapper purges mid-build; fails without fix).
Checked ok: Purge deletes files (Disk.Delete idempotent on missing files, local+S3) then rows; job joins errors and keeps purging other users; leftover ready rows still swept by Expire; fetch-based button (same-origin POST, CSRF via Fetch Metadata), 429 -> cooldown state; i18n keys in en/pt-BR/es.
Re-run: bun lint/typecheck clean, `bun run test` 50 passed; `go test -tags integration ./...` all ok; golangci-lint (integration) app/library, app/jobs, tests 0 issues.
Low (no action): pt-BR/es strings not native-reviewed; Purge stops at the first failing row of a user (rest retried only via Expire for ready rows).
Update: superseded by FIX-export-purge (5e0cf29, in main); agent/especialista/P1.F.4.1-followup-ui-purge-fix must NOT be merged.
