# Task ADOPT-VF-046 (priority 3, rodada 0, model: Sonnet, senior pool)

Objective: move ban/unban into the admin panel form (Virtual("banned") + AfterSave + s.Tx.AfterCommit for user.banned) per VF-046; refactor accounts.Ban/Unban to take database.Executor. Touches auth/security semantics: keep the 'banned' guard tests.

Source: deferred from ADOPT-vukem-v0.33.0 (adoption line in `.orchestration/reports/VF-046.md`). Branch `agent/senior/ADOPT-VF-046` from `origin/main` (vukem v0.33.0 already in go.mod). Remove the `workaround(vukem)` code, keep behaviour, tests green.
Acceptance: `go build ./...`; `golangci-lint run --build-tags integration ./...` 0 issues; `TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable go test -tags integration ./...` no FAIL; update `vukem-feedback.md` status and `CHANGELOG.md`. Report + `1-REV-ADOPT-VF-046.md`.
