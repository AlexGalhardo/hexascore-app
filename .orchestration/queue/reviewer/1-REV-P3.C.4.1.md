# REV-P3.C.4.1

Branch: `agent/implementer/P3.C.4.1` in worktree `C:\dev\hexascore-app\wt\hexascore-p3c41`.

Report: `.orchestration/reports/P3.C.4.1.md`.

Acceptance (the test file is `//go:build integration`; without the tag the command prints `[no test files]` and is not a pass):

- `go test -tags=integration ./tests -count=1 -run TestAppTelemetry -v`
- Env: `TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable`
- Run it from the worktree. `go test` from `C:\dev\hexascore-app` fails (no go.mod).

The output must show PASS. The requests must hit HTTP `/app/v1/client-errors` and `/app/v1/client-metrics`. Expect 204, unknown field 400, the 11th error 429, bodies over 16 KiB and 4 KiB 413, and prune at day 8 on the fake clock. Reject a stored user id, token id, e-mail or IP. Reject 202 or a typed POST that becomes 201. Reject if the implementer wrote `PRONTO_PARA_MERGE`.

`TestSecurity_A01_HexaRouteStackInventory` failing only on `GET /api/items/{id}/ratings` is pre-existing. Do not fix it in this review.

On approval, append `PRONTO_PARA_MERGE agent/implementer/P3.C.4.1` to the report and commit that mark on the branch. Do not merge. Do not tick ROADMAP. Do not push.
