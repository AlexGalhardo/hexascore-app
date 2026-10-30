# REV-P3.A.5.3

Branch: `agent/implementer/P3.A.5.3` in worktree `C:\dev\hexascore-app\wt\hexascore-p3a53`.

Report: `.orchestration/reports/P3.A.5.3.md`. Do not treat that report as a review. The implementer did not append a merge mark.

Acceptance, from that worktree, then again on local main after the merge:

```
$env:TEST_DATABASE_URL = "postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable"
go test -tags=integration ./tests -count=1 -run "TestAPIWebhookTest|TestPruneDeliveries|TestMetrics_WebhookOutcome"
```

Untagged `go test ./tests` is not acceptance. The `tests` package is `//go:build integration`.

Scope: POST `/api/v1/webhooks/{id}/test` returns 200 `{delivered,status}` (not 201). Prune `webhook_deliveries` older than 30 days at 04:40 UTC (`40 4 * * *`). Counter `api.webhook.deliveries` label `outcome`. No customer delivery-log route. No new dependency. Do not implement P3.A.6. Production sender stays https-only.

On approval, append `PRONTO_PARA_MERGE` and merge `--no-ff` into local `main` at `C:\dev\hexascore-app\hexascore-vukem`. Do not push. Re-run the command on that main checkout, then tick ROADMAP P3.A.5.3 only if it passes. Write `C:\Users\aleex\AppData\Local\Temp\grok-goal-bff5b625ad4e\implementer\acceptance\P3.A.5.3.log` as UTF-8 without BOM (PowerShell `Out-File` is UTF-16; use .NET `UTF8Encoding($false)`). The log must be the re-run on local main, and it must include the merge commit hash. Do not stage `.env.example`, heartbeats, `status.md`, `test-results/`, `framework-prs.md`, or `__pycache__`. Do not start another task. Do not edit `wt\hexascore-p5a`, `wt\hexascore-p2f`, `wt\hexascore-vf001`, or any other worktree. Do not tick P5.
