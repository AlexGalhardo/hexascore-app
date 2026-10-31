# REV-P3.C.3.2

Branch: `agent/implementer/P3.C.3.2` in worktree `C:\dev\hexascore-app\wt\hexascore-p3c32`.

Report: `.orchestration/reports/P3.C.3.2.md`.

Acceptance (the test file is `//go:build integration`; without the tag the command prints `[no test files]` and is not a pass):

- `go test -tags=integration ./tests -count=1 -run TestAppLists -v`
- Env: `TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable`

The output must show PASS. The test must hit HTTP `/app/v1`. Reject if GET `/app/v1/me/lists` or GET `/app/v1/lists/{id}` is missing, or if those reads require `lists:write` instead of `catalog:read`. Reject a second mount of the list writes or of PUT/DELETE `/app/v1/items/{id}/edition`. Edition stays on `status:write` (current SDD 05 §6.4 and SDD 14 §6.3). The ROADMAP note that defaults edition to `lists:write` is stale; do not move it. Reject if the implementer wrote `PRONTO_PARA_MERGE`.

On approval, append `PRONTO_PARA_MERGE agent/implementer/P3.C.3.2` to the report, commit that mark on the branch, and merge `--no-ff` into local `main` at `C:\dev\hexascore-app\hexascore-vukem` with message `merge: P3.C.3.2 app lists onto local main`. Do not push. Re-run the acceptance command on `main`. Write the verbose output as UTF-8 without BOM to `C:\Users\aleex\AppData\Local\Temp\grok-goal-bff5b625ad4e\implementer\acceptance\P3.C.3.2.log` (`New-Object System.Text.UTF8Encoding $false`, not `Set-Content -Encoding utf8`). The first four bytes must be `3D 3D 3D 20`. Tick only the P3.C.3.2 checkbox in ROADMAP.md after that log is green. Commit the tick as `docs: tick P3.C.3.2 after the main acceptance run`. Do not stage `.env.example`, heartbeats, `status.md`, `test-results/`, or `framework-prs.md`.
