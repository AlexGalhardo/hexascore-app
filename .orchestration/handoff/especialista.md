# Handoff especialista (2026-10-29, session 9)
STATUS: standby (owner rule: no HANDOFF). Last delivered: P2.E.2.1 (agent/especialista/P2.E.2.1). Earlier this session: Worktree C:\dev\hexascore-app\wt\especialista on agent/especialista/P2.C.1.1, clean.
Queue left: 4-P2.D.3.1, 5-P2.F.2.1, 9-APP-PASSKEYS-b (blocked). Wait for orchestrator message before claiming the next.
Done this session:
- REV-P1.H.4.1b -> APROVADO (agent/senior/P1.H.4.1b), no changes.
- P2.E.1.1 -> self PRONTO (agent/especialista/P2.E.1.1): hexascore:predeploy = drop INVALID indexes + native vukem m.Up (v0.33 lock + NoTransaction); VF-NEW-P2.E.1.1.
- P2.C.1.1 -> self PRONTO (agent/especialista/P2.C.1.1): app/logging handler wrapping app.Log + RequestLog/RequestRoute + NotBanned SetUser.
Pitfalls (new):
- vukem v0.33.0 already ships VF-017/VF-026 fixes: ROADMAP items still describing database/concurrent/ or hashtext lock are stale.
- vukem Logger captures app.Log when routes.Web calls DefaultMiddlewareWith; wrap app.Log in an earlier WithRoutes callback.
- Logger logs the outer request: inner context values are invisible; use a mutable holder put in ctx by an outer server middleware.
- Python on Windows text mode writes CRLF: use binary read/write for repo files.
- vukemtest: AppOptions vukem.WithStdout / WithLogOutput override the io.Discard defaults (appended later).
Merge overlap: P2.E.1.1 vs senior ADOPT-VF-040 (bootstrap/commands.go, runbook); CHANGELOG [Unreleased] on all branches.
- P2.E.2.1: app/backup frames + backup:run; added vukem drivers/s3 v0.3.1 (flagged). Local pg_dump: host has none; build GOOS=linux binary and run it in postgres:18.6-alpine3.24 with --add-host host.docker.internal:host-gateway, APP_ENV=development. RustFS creds hexascore/hexascore-local-secret; bucket hexascore-local exists.
- Session 9 cont.: REV-ADOPT-VF-024 CORRIGIDO (agent/especialista/ADOPT-VF-024-fix: double locale prefix on Back redirects + sampler prefix), REV-P1.I.2.1 APROVADO (+senior 2-E2E-HARDEN-P1.I.2.1), REV-ADOPT-VF-035 APROVADO, REV-ADOPT-VF-040 APROVADO.
- e2e isolation: COMPOSE_PROJECT_NAME=hx-especialista + override file (postgres/rustfs ports !reset [], app 127.0.0.1:3010), E2E_BASE_URL, E2E_APP_CONTAINER=hx-especialista-app-1; down -v after.
