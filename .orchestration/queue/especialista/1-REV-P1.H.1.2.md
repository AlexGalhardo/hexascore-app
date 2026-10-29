# REV-P1.H.1.2 (priority 1, rodada 0)

Task: P1.H.1.2 (plans, /api/v1 stack, quota meter in Postgres). Implemented by the orchestrator on `main`: commits 99bf4f6 and the follow-up fix "fix(api): keep flushing quota deltas visible to admission" (already merged). Review `git show 99bf4f6` plus that fix; deliver fixes as `agent/especialista/P1.H.1.2-fix` from `origin/main`.
Focus: concurrency of `api.Meter` (Admit/Refund/Flush/restore), fail-closed paths (503 on stale meter and plan errors), refund rules (400 and 5xx), credits before plan (D-23), month rollover, middleware order in `routes/api_v1.go` vs SDD 05 §7.4, the runner choice instead of a scheduled job (SDD 05 §7.2 amended).
Acceptance to re-run: `go test -race ./app/api/... -run 'TestQuota|TestKeys|TestPlans'`; `TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable go test -tags integration ./tests -run 'TestCreditBalanceConsumedFirst|TestRefundOn5xx|TestQuotaFlushCrash'`.
Report: `.orchestration/reports/REV-P1.H.1.2.md`.
