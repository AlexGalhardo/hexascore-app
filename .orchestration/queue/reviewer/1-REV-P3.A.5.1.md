# REV-P3.A.5.1

Branch: `agent/implementer/P3.A.5.1` in worktree `C:\dev\hexascore-app\wt\hexascore-p3a51`.

Report: `.orchestration/reports/P3.A.5.1.md`. Do not treat that report as a review. The implementer did not append a merge mark.

Acceptance, from the worktree, with `TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable`:

- `go test ./app/webhooks -count=1 -run TestPublicURL`
- `go test -tags=integration ./tests -count=1 -run 'TestAPIWebhooks|TestSecurity_A10_WebhookSSRF|TestSecurity_A02_WebhookSecretEncrypted'`

The ROADMAP line omits `-tags=integration`. The test files are `//go:build integration`, so an untagged run is not acceptance.

On approval, append `PRONTO_PARA_MERGE` and merge `--no-ff` into local `main` at `C:\dev\hexascore-app\hexascore-vukem`. Do not push. Re-run both commands on that main checkout, then tick ROADMAP P3.A.5.1 only if they pass. Write `C:\Users\aleex\AppData\Local\Temp\grok-goal-bff5b625ad4e\implementer\acceptance\P3.A.5.1.log` as UTF-8 without BOM (PowerShell `Out-File` is UTF-16; use .NET `UTF8Encoding($false)`). The log must be the re-run on local main, and it must include the commit hash. Do not stage `.env.example`, heartbeats, `status.md`, `test-results/`, or `framework-prs.md`. Do not start another task. Do not edit `wt\hexascore-p5a` or any other worktree.
