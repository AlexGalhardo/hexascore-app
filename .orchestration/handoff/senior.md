# senior handoff (2026-10-29, seventeenth session)
Delivered: P1.F.4.1-followup-problem-codes, P1.F.4.1-followup-ui-purge (both pushed, reports + 1-REV queued).
Queue left: 3-P1.F.2.1, 3-P1.H.3.1, 4-P1.G.2.4, 4-P2.D.5.1.
Worktree wt/senior on agent/senior/P1.F.4.1-followup-ui-purge; next: git fetch && git switch -c agent/senior/<ID> origin/main.
Pitfalls: TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable; tests needing now()/interval use pgtest.Database; python on Windows writes CRLF (sed -i 's/\r$//'); golangci-lint fmt <file> for golines; routes import app middleware as appmiddleware; app/http/middleware.CodedProblem wraps handlers to emit problem codes (VF-016 workaround); biome via bunx biome check --write.
