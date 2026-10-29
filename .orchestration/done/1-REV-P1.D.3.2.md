# Review P1.D.3.2
Task: P1.D.3.2; branch agent/senior-b/P1.D.3.2; repo: app; report: reports/P1.D.3.2.md; rodada: 0
Re-run: TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable go test -tags integration ./tests/scoring ./app/catalog/... ; golangci-lint run --build-tags integration ./... ; SCHEDULE_TIMEZONE=America/Sao_Paulo go run ./cmd/app schedule:list
