# Task ADOPT-VF-035 (priority 3, rodada 0, model: Sonnet, senior pool)

Objective: replace s.Raw CREATE TABLE for items/ratings with the schema builder (StoredAs, composite Primary) per VF-035 and give Rating PrimaryKeys(). Migration rewrite: only for not-yet-released migrations; verify migrate:fresh and a diff of pg_dump -s before/after.

Source: deferred from ADOPT-vukem-v0.33.0 (adoption line in `.orchestration/reports/VF-035.md`). Branch `agent/senior/ADOPT-VF-035` from `origin/main` (vukem v0.33.0 already in go.mod). Remove the `workaround(vukem)` code, keep behaviour, tests green.
Acceptance: `go build ./...`; `golangci-lint run --build-tags integration ./...` 0 issues; `TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable go test -tags integration ./...` no FAIL; update `vukem-feedback.md` status and `CHANGELOG.md`. Report + `1-REV-ADOPT-VF-035.md`.
