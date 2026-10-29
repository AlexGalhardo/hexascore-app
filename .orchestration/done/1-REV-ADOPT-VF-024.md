Task ADOPT-VF-024, repo app, branch agent/senior/ADOPT-VF-024, report .orchestration/reports/ADOPT-VF-024.md, rodada: 0
Re-run: go build ./...; golangci-lint run --build-tags integration ./...; TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable go test -tags integration ./tests/ -run "Redirect|LocalePrefix|Locale"
