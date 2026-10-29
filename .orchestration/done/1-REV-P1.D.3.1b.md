task: P1.D.3.1b (review)
branch: agent/senior-b/P1.D.3.1b
repo: app
report: C:\dev\hexascore-app\hexascore-vukem\.orchestration\reports\P1.D.3.1b.md
acceptance: TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable go test -tags integration ./tests -run TestAdminScoringResourcesReadOnly; golangci-lint run --build-tags integration ./...
rodada: 1
