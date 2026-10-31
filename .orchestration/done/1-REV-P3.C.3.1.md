# REV-P3.C.3.1

Branch: `agent/implementer/P3.C.3.1` in worktree `C:\dev\hexascore-app\wt\hexascore-p3c31`.

Report: `.orchestration/reports/P3.C.3.1.md`.

Acceptance (the test file is `//go:build integration`; without the tag the command prints `[no test files]` and is not a pass):

- `go test -tags=integration ./tests -count=1 -run TestAppMe -v`
- Env: `TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable`

The output must show PASS. The test must hit HTTP `/app/v1`: a taken nickname is 422 `nickname-taken`, and a second export inside 7 days is 429 `export-cooldown` with `Retry-After`. Reject a second mount of `POST /app/v1/me/export` (it already lives in `routes/exports.go`). Reject a handler that writes the nickname before returning 422.

On approval, append `PRONTO_PARA_MERGE agent/implementer/P3.C.3.1` to the report, commit that mark on the branch, and merge `--no-ff` into local `main` at `C:\dev\hexascore-app\hexascore-vukem`. Do not push. Re-run the acceptance command on `main`. Write the verbose output as UTF-8 without BOM to `C:\Users\aleex\AppData\Local\Temp\grok-goal-bff5b625ad4e\implementer\acceptance\P3.C.3.1.log` (`New-Object System.Text.UTF8Encoding $false`, not `Set-Content -Encoding utf8`). Tick only the P3.C.3.1 checkbox in ROADMAP.md after that log is green. Do not stage `.env.example`, heartbeats, `status.md`, `test-results/`, or `framework-prs.md`.
