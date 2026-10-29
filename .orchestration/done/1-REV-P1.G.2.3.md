task: P1.G.2.3
branch: agent/senior/P1.G.2.3
repo: app
report: .orchestration/reports/P1.G.2.3.md
rodada: 1
acceptance: TEST_DATABASE_URL=... go test -tags integration ./tests -run TestInertiaPages_Catalog_Show; golangci-lint run --build-tags integration ./...; bun run lint && bun run typecheck && bun run test
