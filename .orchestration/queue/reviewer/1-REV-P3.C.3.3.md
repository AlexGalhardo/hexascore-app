# REV-P3.C.3.3

Branch: `agent/implementer/P3.C.3.3` in worktree `C:\dev\hexascore-app\wt\hexascore-p3c33`.

Report: `.orchestration/reports/P3.C.3.3.md`.

Acceptance (the test file is `//go:build integration`; without the tag the command prints `[no test files]` and is not a pass):

- `go test -tags=integration ./tests -count=1 -run 'TestAppLibrary|TestAppNotifications|TestAppReports' -v`
- Env: `TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable`

The output must show PASS for all three tests. They must hit HTTP `/app/v1`. A private profile returns the nickname only (no stats, ratings or lists). 101 notification ids is 400. Reporting yourself is 403 `report-self`. Reject a second mount of `/me/library`, `/users/{nickname}`, `/me/notifications`, `POST /me/notifications/read` or `POST /reports`. Reject if the implementer wrote `PRONTO_PARA_MERGE`.

On approval, append `PRONTO_PARA_MERGE agent/implementer/P3.C.3.3` to the report, commit that mark on the branch, and merge `--no-ff` into local `main` at `C:\dev\hexascore-app\hexascore-vukem` with message `merge: P3.C.3.3 app library onto local main`. Do not push. Re-run the acceptance command on `main`. Write the verbose output as UTF-8 without BOM to `C:\Users\aleex\AppData\Local\Temp\grok-goal-bff5b625ad4e\implementer\acceptance\P3.C.3.3.log` (`New-Object System.Text.UTF8Encoding $false`, not `Set-Content -Encoding utf8`). The first four bytes must be `3D 3D 3D 20`. Tick only the P3.C.3.3 checkbox in ROADMAP.md after that log is green. Commit the tick as `docs: tick P3.C.3.3 after the main acceptance run`. Do not stage `.env.example`, heartbeats, `status.md`, `test-results/`, or `framework-prs.md`.
