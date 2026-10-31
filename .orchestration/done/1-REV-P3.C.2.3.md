# REV-P3.C.2.3

Branch: `agent/implementer/P3.C.2.3` in worktree `C:\dev\hexascore-app\wt\hexascore-p3c23`.

Report: `.orchestration/reports/P3.C.2.3.md`.

Acceptance (the test file is `//go:build integration`; without the tag the command prints `[no test files]` and is not a pass):

- `go test -tags=integration ./tests -count=1 -run TestAppScoring -v`
- Env: `TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable`

The output must show PASS. The test must hit HTTP `/app/v1`: a repeated rating PUT with the same `Idempotency-Key` stores one row, and a token without `ratings:write` is 403. Do not require a new `app/appapi/scoring.go` registrar: `routes/ratings.go` and `routes/status.go` already mount these routes, and a second mount panics. Reject a duplicate registration or a test that never calls the HTTP handlers.

On approval, append `PRONTO_PARA_MERGE agent/implementer/P3.C.2.3` to the report, commit that mark on the branch, and merge `--no-ff` into local `main` at `C:\dev\hexascore-app\hexascore-vukem`. Do not push. Re-run the acceptance command on `main`. Write the verbose output as UTF-8 without BOM to `C:\Users\aleex\AppData\Local\Temp\grok-goal-bff5b625ad4e\implementer\acceptance\P3.C.2.3.log` (`New-Object System.Text.UTF8Encoding $false`, not `Set-Content -Encoding utf8`). Tick only the P3.C.2.3 checkbox in ROADMAP.md after that log is green. Do not stage `.env.example`, heartbeats, `status.md`, `test-results/`, or `framework-prs.md`.
