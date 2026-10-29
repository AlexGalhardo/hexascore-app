# REV P1.G.2.4
task: P1.G.2.4; branch: agent/senior/P1.G.2.4; repo: app; report: C:\dev\hexascore-app\hexascore-vukem\.orchestration\reports\P1.G.2.4.md
re-run: go test -tags integration ./tests -run TestInertiaPages_Entity_Show; golangci-lint run --build-tags integration ./...; go test ./... ; go test -tags integration ./...; bun run lint && bun run typecheck && bun run test
rodada: 0
