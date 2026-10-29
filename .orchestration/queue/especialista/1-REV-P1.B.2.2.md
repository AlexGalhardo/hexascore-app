task: P1.B.2.2
branch: agent/senior-b/P1.B.2.2
repo: app
report: .orchestration/reports/P1.B.2.2.md
acceptance: go test ./app/catalog/... -run 'TestMerge|TestRedirect|TestVisibility|TestCursor'; go test -tags=integration ./app/catalog/... -run TestListingFilters; golangci-lint run --build-tags integration ./...
rodada: 1
