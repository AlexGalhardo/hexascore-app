# REV P1.H.3.2 (rodada: 0)
Task ID: P1.H.3.2. Branch: agent/senior/P1.H.3.2. Repo: app. Report: reports/P1.H.3.2.md.
Re-run: bun run lint && bun run typecheck && bun run test; golangci-lint run --build-tags integration ./...; TEST_DATABASE_URL=... go test -tags integration ./tests -run 'Pricing|APIReference|DocsPages'; scripts/openapi-check.sh; e2e (isolated compose) `playwright test docs pricing` -> 36 passed; curl acceptance from the ROADMAP.
