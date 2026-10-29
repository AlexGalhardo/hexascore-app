task: P1.H.2.1c
branch: agent/senior/P1.H.2.1c
repo: app
report: .orchestration/reports/P1.H.2.1c.md
acceptance: go test -count=1 -tags integration ./tests -run 'TestAPISearch|TestAPILookup|TestAPIContract'; golangci-lint run --build-tags integration ./...
rodada: 1
