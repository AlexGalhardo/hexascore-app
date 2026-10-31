# REV-P4.A.2.1

Branch: `agent/implementer/P4.A.2.1` in worktree `C:\dev\hexascore-app\wt\hexascore-p3c51`.

Report: `.orchestration/reports/P4.A.2.1.md`.

Acceptance, from that worktree, with `TEST_DATABASE_URL` pointing at the compose Postgres (`127.0.0.1:5433`, database `postgres`):

- `go test -tags=integration ./app/community/... -count=1 -run TestQuota` → 3 submits return 303, the 4th returns 429 with type ending in `/suggestion-quota` and `Retry-After` > 0, and a submit at the next 00:00 UTC returns 303.
- `-run TestWithdrawStillCounts` → a withdrawn submit still counts, so the 4th of that UTC day is 429.
- `-run 'TestEligibility|TestConcurrentSubmit'` → 403 `suggestion-not-eligible` with `eligibleAt` when the account is under 7 days (and for a temporary ban or block); 20 concurrent submits admit exactly 3.

Reject a quota stored in the vukem limiter. Reject a refused submit that changes status or sets `submitted_at`. Reject if the implementer wrote `PRONTO_PARA_MERGE`.

On approval, append `PRONTO_PARA_MERGE agent/implementer/P4.A.2.1` to the report and commit that mark on the branch. Do not merge. Do not tick ROADMAP. Do not push.
