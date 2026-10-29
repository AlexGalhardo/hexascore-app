# REV P1.G.4.1
task: P1.G.4.1, branch: agent/senior/P1.G.4.1, repo: app, report: .orchestration/reports/P1.G.4.1.md, rodada: 0
Re-run: go test ./tests -run 'Robots|Docs|Legal|ErrorPage'; bun run lint && bun run typecheck && bun run test; golangci-lint run --build-tags integration ./...; check scripts/build-docs.ts escaping (XSS) and docs.json freshness (bun run build:docs gives no diff).
