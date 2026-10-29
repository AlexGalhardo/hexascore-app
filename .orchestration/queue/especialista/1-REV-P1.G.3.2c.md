# REV P1.G.3.2c
task: P1.G.3.2c; branch: agent/senior/P1.G.3.2c; repo: app; report: C:\dev\hexascore-app\hexascore-vukem\.orchestration\reports\P1.G.3.2c.md
re-run: go test -tags integration ./tests -run 'TestEditionPickerList|TestEditionNotOfWork'; golangci-lint run --build-tags integration ./...; go test -tags integration ./...; bun run lint && bun run typecheck && bun run test
rodada: 0
