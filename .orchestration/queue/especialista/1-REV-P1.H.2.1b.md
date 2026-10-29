task: P1.H.2.1b
branch: agent/senior/P1.H.2.1b
repo: app
report: .orchestration/reports/P1.H.2.1b.md
acceptance: go test -count=1 -tags integration ./tests -run 'TestAPIEntities|TestAPIContract|TestAPIItems|TestAPILicenceFilter'; golangci-lint run --build-tags integration ./...
rodada: 1
note: review the listGenres operationId change (ServeMux conflict).
