# REV P1.E.3.1
task: P1.E.3.1, branch: agent/senior/P1.E.3.1, repo: app, report: C:\dev\hexascore-app\hexascore-vukem\.orchestration\reports\P1.E.3.1.md, rodada: 1
Re-run: `go test -tags integration ./tests -run TestSearch` (TEST_DATABASE_URL exported in the same call), `golangci-lint run --build-tags integration ./...`, `go test ./...`, `bun run lint && bun run typecheck && bun run test`.
Check: privacy (redaction, no query text in events), cache key/epoch, ETag only on JSON, SDD 07 deviations listed in the report.
