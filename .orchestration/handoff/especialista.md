# Handoff especialista (2026-10-29, end of session 9: context limit)
STATUS: IDLE, no task in progress. Worktree C:\dev\hexascore-app\wt\especialista (last on agent/especialista/P2.F.2.1, clean; next task: git fetch, switch -c from origin/main).
Queue for the next session, in order: 3-P2.F.2.2, 4-P2.F.2.3, 5-P2.F.2.7, 6-P2.E.2.2, 7-P2.E.3.1 (9-APP-PASSKEYS-b blocked until a vukem release with VF-001). Chain them without waiting; report each with PRONTO_PARA_MERGE (self, needs orchestrator sampling).
Delivered and merged this session: REV-P1.H.4.1b, P2.E.1.1 (hexascore:predeploy), P2.C.1.1 (privacy log handler), P2.E.2.1 (backup:run, drivers/s3 v0.3.1 accepted), REV ADOPT-VF-024 (fix), P1.I.2.1, ADOPT-VF-035/040/046, P2.D.3.1 (AdminGuard), P2.F.2.1 (legacyimport runner).
Context for the queue:
- Legacy import (P2.F.2.2/2.3/2.7): database/legacyimport. Stages() returns stages with nil Batch; fill Batch (BatchFunc: keyset after `after`, idempotent upserts, runner writes the checkpoint in the same tx). Fixture testdata/fixture.sql is synthetic (Better Auth column names from legacy apps/web/lib/db/schema/auth.ts); extend it per stage. Preflight "no_login_method" = no credential account with password, no google, no passkey.
- Backups (P2.E.2.2 verify/prune): app/backup FrameReader(r, enc, objectKey) decrypts; object key bound in AD; backup_runs table; Runner in dump.go. Host has no pg_dump/pg_restore: build GOOS=linux binary and run in postgres:18.6-alpine3.24 with --add-host host.docker.internal:host-gateway, APP_ENV=development; RustFS creds hexascore/hexascore-local-secret, bucket hexascore-local exists.
Pitfalls:
- vukem v0.33.0 already ships many VF fixes; check vukem-feedback.md status before building a workaround (VF-017/026/030/035 were fixed natively).
- Test admins need two_factor_confirmed_at (tests/admin_users_test.go makeUser does it) because of AdminGuard.
- e2e isolation: COMPOSE_PROJECT_NAME=hx-especialista + override file (postgres/rustfs ports !reset [], app 127.0.0.1:3010), E2E_BASE_URL, E2E_APP_CONTAINER=hx-especialista-app-1; `down -v` after.
- Python on Windows: use binary read/write; multi-occurrence replaces need unique context (assert count).
- golangci: nolint needs an explanation and its own line when the line is long (golines).
- Long suites: `timeout 590 env TEST_DATABASE_URL=... go test -tags integration ./...` > file; they may go to background; wait with an until-loop on the output file.
