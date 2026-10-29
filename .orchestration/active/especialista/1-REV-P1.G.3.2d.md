# REV P1.G.3.2d
Task ID: P1.G.3.2d. Branch: agent/senior/P1.G.3.2d. Repo: app. Report: .orchestration/reports/P1.G.3.2d.md. rodada: 1
Re-run: go test -tags integration ./tests -run TestEditionPickerList; golangci-lint run --build-tags integration ./...; bun run lint && bun run typecheck && bun run test.
