# REV FIX-entity-name-source
task: FIX-entity-name-source; branch: agent/senior/FIX-entity-name-source; repo: app; report: C:\dev\hexascore-app\hexascore-vukem\.orchestration\reports\FIX-entity-name-source.md
re-run: go test -tags integration ./tests -run 'TestAPILicenceFilter|TestAPIItems|TestAPIEntities|TestAPIContract'; golangci-lint run --build-tags integration ./...; go test -tags integration ./...; bun run test
rodada: 0
