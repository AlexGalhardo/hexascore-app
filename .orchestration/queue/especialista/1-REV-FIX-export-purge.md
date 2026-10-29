# Review FIX-export-purge
task: FIX-export-purge; branch: agent/senior/FIX-export-purge; repo: app; report: .orchestration/reports/FIX-export-purge.md
re-run: golangci-lint run --build-tags integration ./...; go test ./...; go test -tags integration ./... (TEST_DATABASE_URL set)
rodada: 1
