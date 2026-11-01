# FIX-scoring-burst-clock TestReleaseBursts depends on the database clock

status: concluida
branch: agent/especialista/FIX-scoring-burst-clock @ 8c8964e
files: tests/scoring/maintenance_test.go
acceptance:
- `go test -tags integration -count=1 ./tests/scoring/` on main b500e2c → ok (70.7 s), exit 0
validation: golangci-lint (integration tags) tests/scoring 0 issues; test-only change
changelog: Fixed: TestReleaseBursts no longer fails when the Postgres container clock runs ahead of the host.
roadmap: no box
risks / follow-ups: the fixture now sets `limited_until` one minute in the past; production code is unchanged (ReleaseBursts compares with the host clock by design).

PRONTO_PARA_MERGE agent/especialista/FIX-scoring-burst-clock (self-fix from the f78ad42 gate; merged b500e2c)
