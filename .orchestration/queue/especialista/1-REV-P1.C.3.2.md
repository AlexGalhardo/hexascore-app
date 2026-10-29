task: P1.C.3.2
branch: agent/senior-b/P1.C.3.2
repo: app
report: C:\dev\hexascore-app\hexascore-vukem\.orchestration\reports\P1.C.3.2.md
acceptance: TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable go test -tags integration ./tests -run TestDeletion ; golangci-lint run --build-tags integration ./... ; bun run lint && bun run typecheck && bun run test
rodada: 1
