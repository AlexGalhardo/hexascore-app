# REV-P3.C.2.1

Branch: `agent/implementer/P3.C.2.1` in worktree `C:\dev\hexascore-app\wt\hexascore-p3c21`.

Report: `.orchestration/reports/P3.C.2.1.md`.

Acceptance (the test file is `//go:build integration`; without the tag the command prints `[no test files]` and is not a pass):

- `go test -tags=integration ./tests -count=1 -run 'TestAPILicenceFilter|TestAppCatalog' -v`
- Env: `TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable`

The output must show TMDB/IGDB attributed fields on `/app/v1` (synopsis text `TMDB overview`, genre `Metroidvania` or name `IGDB Published`). Trending must stay unmounted (404, not an empty list). Do not tick P3.A.1.4.

On approval, append `PRONTO_PARA_MERGE agent/implementer/P3.C.2.1` to the report, commit that mark on the branch, and merge `--no-ff` into local `main` at `C:\dev\hexascore-app\hexascore-vukem`. Do not push. Re-run the acceptance command on `main`. Write the verbose output as UTF-8 without BOM to `C:\Users\aleex\AppData\Local\Temp\grok-goal-bff5b625ad4e\implementer\acceptance\P3.C.2.1.log`. Tick only the P3.C.2.1 checkbox in ROADMAP.md after that log is green. Do not stage `.env.example`, heartbeats, `status.md`, `test-results/`, or `framework-prs.md`.
