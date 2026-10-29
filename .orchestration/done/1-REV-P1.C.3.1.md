task: P1.C.3.1
branch: agent/senior-b/P1.C.3.1
repo: app
report: .orchestration/reports/P1.C.3.1.md
acceptance: go test ./tests -run 'TestSecurity_A07_Ban|TestSecurity_A01_AdminRoleNotEditableInPanel'; golangci-lint run --build-tags integration ./...
rodada: 1
