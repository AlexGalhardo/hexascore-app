task: P1.B.4.1
branch: agent/senior-b/P1.B.4.1
repo: app
report: .orchestration/reports/P1.B.4.1.md
acceptance: go test -tags=integration ./app/catalog/... -run TestDQ; golangci-lint run --build-tags integration ./...
rodada: 1
