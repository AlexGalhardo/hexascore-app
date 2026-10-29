# REV P1.G.2.2
task: P1.G.2.2, branch agent/senior-b/P1.G.2.2, repo app, report reports/P1.G.2.2.md, rodada: 1
re-run: TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable go test -tags integration ./tests -run TestInertiaPages_Catalog_Index; golangci-lint run --build-tags integration ./...; bun run lint && bun run typecheck && bun run test
