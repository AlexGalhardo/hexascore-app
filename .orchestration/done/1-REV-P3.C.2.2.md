# REV-P3.C.2.2

Branch: `agent/implementer/P3.C.2.2` in worktree `C:\dev\hexascore-app\wt\hexascore-p3c22`.

Report: `.orchestration/reports/P3.C.2.2.md`.

Acceptance (the test file is `//go:build integration`; without the tag the command prints `[no test files]` and is not a pass):

- `go test -tags=integration ./tests -count=1 -run TestAppSearch -v`
- Env: `TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable`

The output must show PASS with and without a token, and `q` longer than 100 characters as 400. Do not require `catalog:read`. Do not tick P3.C.2.3.

On approval, append `PRONTO_PARA_MERGE agent/implementer/P3.C.2.2` to the report, commit that mark on the branch, and merge `--no-ff` into local `main` at `C:\dev\hexascore-app\hexascore-vukem`. Do not push. Re-run the acceptance command on `main`. Write the verbose output as UTF-8 without BOM to `C:\Users\aleex\AppData\Local\Temp\grok-goal-bff5b625ad4e\implementer\acceptance\P3.C.2.2.log` (`New-Object System.Text.UTF8Encoding $false`, not `Set-Content -Encoding utf8`). Tick only the P3.C.2.2 checkbox in ROADMAP.md after that log is green. Do not stage `.env.example`, heartbeats, `status.md`, `test-results/`, or `framework-prs.md`.

Known pre-existing failure, not a reject reason: `TestSecurity_A01_HexaRouteStackInventory` reports unclassified `GET /api/items/{id}/ratings`.
