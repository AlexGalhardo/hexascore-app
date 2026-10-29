Task ADOPT-VF-046, repo app, branch agent/senior/ADOPT-VF-046, report .orchestration/reports/ADOPT-VF-046.md, rodada: 0
Re-run: golangci-lint run --build-tags integration ./...; TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable go test -tags integration ./tests/ -run "Admin|A07"
Check: auth/security semantics of AfterSave ban path.
