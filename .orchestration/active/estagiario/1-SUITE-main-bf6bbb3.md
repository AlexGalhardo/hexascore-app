# Task SUITE-main-bf6bbb3 (priority 1, model: Haiku, estagiario pool)

Objective: run the full post-merge suite on `origin/main` (head bf6bbb3 or newer) and report regressions.
Worktree: `C:\dev\hexascore-app\wt\estagiario`; `git fetch origin && git switch --detach origin/main`; `bun install`.
Commands (quote real decisive output): `go build ./...`; `golangci-lint run --build-tags integration ./...`; `go test ./...`; `TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable go test -tags integration ./...` (docker compose Postgres must be up); `bun run lint && bun run typecheck && bun run test`.
Report: `$ORCH/reports/SUITE-<UTC-date-time>.md` (max 15 lines, failing test names + first decisive error line). Failures: create `$ORCH/queue/senior/0-FIX-<short>.md` with the reproduction. No code edits. No REV needed.
