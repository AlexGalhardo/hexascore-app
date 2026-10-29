# senior-b handoff
Done this session (pushed, reviews queued): P1.E.2.2 (golden set), P1.G.1.2 (CSP) + P1.G.1.2-fix (worker-src/frame-src), P1.C.4.1 (/app/v1/tokens), P1.D.3.1b (scoring admin resources). Earlier: P1.E.2.1 (merged), P1.E.1.1, P1.C.2.1b, P1.C.3.2, P1.B.2.2, P1.B.4.1, P1.C.3.1.
Worktree wt/especialista is on agent/senior-b/P1.G.1.2-fix (clean). Next: `git fetch && git switch -c agent/senior-b/<ID> origin/main`.
Queue (senior) at handoff: 2-P1.D.4.1, 2-P1.D.4.2, 2-P1.F.1.1, 2-P1.G.2.1, 3-P1.C.2.2, 3-P1.C.5.2 (check before claiming).
Open notes: P1.C.4.1 left /app/v1/me* and 428 challenge for later (P1.C.2.2, P1.F.5.1, P1.C.5). P1.E.2.2 found: `zleda` ties at theta 0.2 (SDD 07 11.10, extension decision needed).
Pitfalls: NEVER write in $APP. Use Write tool for Go files (bash heredocs with quotes break). Tests default to SQLite (scoring/search tables are Postgres only): Postgres tests need //go:build integration and pgtest.Database + DATABASE_URL env. golangci-lint fails if another lint runs in parallel: retry. CSPSources starts a new directive from script-src/default-src (copies nonce): write directives into cfg.CSP instead. vukem z.String rejects control chars; admin Filter/Search reject sql.Null* columns. TOTP step 0 is spent by setup confirm: use steps=1 in tests.
Heartbeat: heartbeat/senior-b.txt.
