# REV-P1.H.1.1 (priority 1, rodada 0)

Task: P1.H.1.1 (API keys, hx_live_ guard, /profile/api page). Implemented by the orchestrator directly on `main`, commit 5c0cb60 (already merged). Review `git show 5c0cb60`; deliver fixes as `agent/especialista/P1.H.1.1-fix` from `origin/main`.
Focus: security (golang-security skill): key generation, hash at rest, fail-closed guard (503 vs 401), cross-surface refusal, cap race (FOR UPDATE on users), newKey flash stored in the session until next request (SDD 09 says flash, shown once; judge if acceptable), no in-process LRU yet (shortcut comment, P1.H.1.2).
Acceptance to re-run: `go test ./app/api/... -run TestKeys`; `TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable go test -tags integration ./tests -run 'TestKeys|TestSecurity_A07_ApiKeyHashedAtRest|TestSecurity_A01_CrossSurfaceCredentials'`; `bun run lint && bun run typecheck`.
Report: `.orchestration/reports/REV-P1.H.1.1.md`; findings that need a change become a fix branch (you may fix directly).
