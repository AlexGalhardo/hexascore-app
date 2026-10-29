# senior handoff (2026-10-29, eighteenth session)
Delivered: FIX-export-purge (branch agent/senior/FIX-export-purge pushed, report + 1-REV queued).
Queue left: 3-P1.F.2.1, 3-P1.H.3.1, 4-P1.G.2.4, 4-P2.D.5.1 (none claimed).
Worktree wt/senior on agent/senior/FIX-export-purge; next: git fetch && git switch -c agent/senior/<ID> origin/main.
Pitfalls: TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable (export it in EVERY bash call); python open() defaults to cp1252 on Windows: use encoding utf-8 / binary (it corrupts the section sign); queue-dispatched jobs build their own Exports, so tests needing disk hooks must call svc.Build directly; golangci-lint fmt <file>; routes import app middleware as appmiddleware.
