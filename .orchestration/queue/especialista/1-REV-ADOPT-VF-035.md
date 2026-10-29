Task ADOPT-VF-035, repo app, branch agent/senior/ADOPT-VF-035, report .orchestration/reports/ADOPT-VF-035.md, rodada: 0
Re-run: go test ./app/models/; golangci-lint run --build-tags integration ./...; TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable go test -tags integration ./tests/scoring/
Check: decision to leave released migrations untouched.
