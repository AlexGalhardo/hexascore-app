Task: P1.B.1.2 | repo: app | branch: agent/senior/P1.B.1.2 | report: C:\dev\hexascore-app\hexascore-vukem\.orchestration\reports\P1.B.1.2.md | rodada: 0
Re-run: `go test ./app/catalog/... -run TestRoleCompatibility`; `go test -tags=integration ./app/catalog/... -run 'TestEntityConstraints|TestEntityStats'` (TEST_DATABASE_URL); `golangci-lint run --build-tags integration ./...`.
