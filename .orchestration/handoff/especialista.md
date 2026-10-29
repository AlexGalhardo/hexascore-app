# Handoff especialista (2026-10-29, session 9)
STATUS: HANDOFF after 3 deliveries. Worktree C:\dev\hexascore-app\wt\especialista on agent/especialista/P2.C.1.1, clean.
Queue left: 9-APP-PASSKEYS-b (blocked until a vukem release with VF-001).
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
