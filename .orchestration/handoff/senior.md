# senior handoff (nineteenth session)
Delivered: P1.F.2.1 (branch agent/senior/P1.F.2.1 pushed, report + 1-REV queued).
Queue left: 3-P1.H.3.1, 4-P1.G.2.4, 4-P2.D.5.1 (none claimed).
Worktree wt/senior on agent/senior/P1.F.2.1; next: git fetch && git switch -c agent/senior/<ID> origin/main.
Pitfalls: TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable (export in EVERY bash call); heredocs containing an apostrophe break the Bash tool: use Write; public routes need auth.Optional(guard) to see the viewer; users.username is lowercase (CHECK), display_username keeps case; editions.source needs a data_sources row (insert 'hexascore' in tests); golangci-lint revive rejects stuttering names (library.LibraryX).
