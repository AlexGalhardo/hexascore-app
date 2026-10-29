# senior-b handoff
Done this session (pushed, reviews queued in queue/especialista): P1.G.2.1 (agent/senior-b/P1.G.2.1), P1.C.2.2 (agent/senior-b/P1.C.2.2), P1.F.3.1 (agent/senior-b/P1.F.3.1), P1.C.5.2 (agent/senior-b/P1.C.5.2). Earlier merged: P1.D.3.2, P1.C.5.1, P1.E.2.2, P1.G.1.2, P1.C.4.1, P1.D.3.1b.
Worktree wt/especialista is on agent/senior-b/P1.C.5.2 (clean). Next: `git fetch && git switch -c agent/senior-b/<ID> origin/main`.
Queue (senior) at handoff: 2-P1.F.3.2 (needs F.3.1 + C.5.2 merged), 2-P1.G.2.2 (needs G.2.1 merged), 2-P1.G.3.1, 3-P1.E.3.2, 3-P1.F.2.1, 3-P1.H.1.1.
Open notes / human gates for orchestrator:
- SESSION_DRIVER=database must be set on Railway (profile sessions list; added to .env.example). Shared CACHE_STORE needed in prod for PoW counters/single-use (memory = per process).
- F.3.2 contact form: reuse middleware.Challenger.Require(name) and add a rule to middleware.ChallengeRules; honeypot pattern in account.Register. Protected PoW tests attach `X-Challenge` via tests/pow_test.go `solved(...)`.
- notifications daily purge (NFR-14-6) not scheduled; Home/Index has no covers (no cover URL builder yet); G.2.2 can reuse catalog.Highlights/ListParams.
- password_set_at is now set at sign-up/reset/profile change; 2FA settings routes 403 until set.
Pitfalls: python text-mode writes CRLF on Windows (use newline='\n') or run `golangci-lint fmt ./...`. Run gate: TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable go test -tags integration ./... ; bun install in a fresh worktree before bun lint/typecheck. `export` of TEST_DATABASE_URL does not persist across Bash calls. Validation redirects go to Referer: tests must set the Referer header (see tests/profile_test.go). Cookie sessions can't list sessions: tests use SESSION_DRIVER=database via vukemtest.Options.Env.
Heartbeat: heartbeat/senior-b.txt.
