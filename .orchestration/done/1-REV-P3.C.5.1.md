# REV-P3.C.5.1

Branch: `agent/implementer/P3.C.5.1` in worktree `C:\dev\hexascore-app\wt\hexascore-p3c51`.

Report: `.orchestration/reports/P3.C.5.1.md`.

Acceptance:

- Fresh `openapi:export -prefix /app/v1` matches `openapi/app-v1.json`.
- Fresh `codegen:client -prefix /app/v1` matches `packages/app-client/src/api.ts`. Do not hand-edit that file.
- `bun run --filter @hexascore/app-client typecheck test` from the worktree → exit 0. The test calls `createClient` with a fake `fetch` and expects `ApiError` on a non-2xx response.
- `qa/schemathesis/run.sh openapi/app-v1.json` → 0 failures. Do not pass `--checks all`: it re-enables the empty-204 schema disables. `415` is expected for a non-JSON body.
- `qa/k6/app-v1.js` via `grafana/k6:2.3.0` → `http_req_duration` `p(95)<300` on anonymous `GET /app/v1/search`.

Reject if the implementer wrote `PRONTO_PARA_MERGE`. Reject a served `/app/v1/openapi.json` in production. Reject a new dependency.

On approval, append `PRONTO_PARA_MERGE agent/implementer/P3.C.5.1` to the report and commit that mark on the branch. Do not merge. Do not tick ROADMAP. Do not push.
