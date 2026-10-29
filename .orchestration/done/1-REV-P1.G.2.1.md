Task: P1.G.2.1 | branch: agent/senior-b/P1.G.2.1 | repo: app | report: .orchestration/reports/P1.G.2.1.md | rodada: 0
Re-run: go test ./tests -run 'TestInertiaPages/Home'; TEST_DATABASE_URL=... go test -tags integration ./tests -run TestHome_Highlights; golangci-lint run --build-tags integration ./...; bun run lint && bun run typecheck && bun run test
