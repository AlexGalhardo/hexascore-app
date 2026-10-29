# Task ADOPT-VF-040 (priority 3, rodada 0, model: Sonnet, senior pool)

Objective: remove migrate:locked (app/console/migrate_locked.go) and scripts/concurrent-index.sh in favour of the vukem v0.33.0 migration lock / concurrent index support documented in VF-040; migrations using CONCURRENTLY must still apply.

Source: deferred from ADOPT-vukem-v0.33.0 (adoption line in `.orchestration/reports/VF-040.md`). Branch `agent/senior/ADOPT-VF-040` from `origin/main` (vukem v0.33.0 already in go.mod). Remove the `workaround(vukem)` code, keep behaviour, tests green.
Acceptance: `go build ./...`; `golangci-lint run --build-tags integration ./...` 0 issues; `TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable go test -tags integration ./...` no FAIL; update `vukem-feedback.md` status and `CHANGELOG.md`. Report + `1-REV-ADOPT-VF-040.md`.

Note (especialista, P2.E.1.1): hexascore:predeploy now exists (agent/especialista/P2.E.1.1) and does not use migrate:locked; runbook already says the script is not used. Rebase on main after that merge.
