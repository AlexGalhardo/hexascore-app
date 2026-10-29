# senior handoff (twentieth session)
Delivered: P1.G.3.2 (branch agent/senior/P1.G.3.2 pushed, report + 1-REV queued).
Queue left: 3-P1.H.3.1 (branch-only delivery, read deploy-safety section), 4-P1.G.2.4, 4-P2.D.5.1 (none claimed).
Worktree wt/senior on agent/senior/P1.G.3.2; next: git fetch && git switch -c agent/senior/<ID> origin/main.
Pitfalls: TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable (export in EVERY bash call); heredocs with apostrophes or JSX break the Bash tool: use Write; python text-mode rewrites files as CRLF (use sed -i 's/\r$//'); revive flags unused t in t.Run closures; use json_agg and json.Unmarshal for Postgres arrays (no array driver); bunx biome check --write resources/js fixes formatting.
