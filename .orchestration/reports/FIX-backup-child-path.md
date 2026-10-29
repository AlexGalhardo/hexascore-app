# FIX-backup-child-path
STATUS: concluida
Branch: agent/especialista/FIX-backup-child-path (pushed, from origin/main). Commits e79b46c (tests + CHANGELOG), 40c1e24 (parity-checklist R05: guest /admin -> /login accepted).
Audit: the backup code starts 3 children: pg_dump (Run), pg_restore (Restore/Drill) both via pgEnv = PG* + PATH; pg_restore --list (Verify) = PATH only. Prune and CopyMonthly start none. The fix itself already landed with P2.E.3.1 (binPath); this task adds the proof. HOME/TMPDIR not needed (custom format on stdin/stdout, no temp files).
- `TestPgEnvIsMinimal` (unit; APP_KEY set in parent is absent from the child env, PATH present) PASS.
- In golang:1.27.3-alpine3.24 + postgresql18-client: TestBackupDrill (now also runs r.Verify -> pg_restore --list with env {PATH}) PASS, TestRestoreReappliesDeletions PASS.
- golangci --build-tags integration (app/backup, tests) "0 issues"; markdownlint 0.
PRONTO_PARA_MERGE (self, needs orchestrator sampling) agent/especialista/FIX-backup-child-path
