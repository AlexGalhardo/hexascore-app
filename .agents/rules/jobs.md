---
paths:
  - "app/jobs/**"
  - "internal/ingest/**"
  - "routes/schedule.go"
  - "app/scheduler/**"
---

# Background jobs and scheduled tasks

Why this rule exists: on 2026-11-02 the audit found the nightly score recompute, the similarity job and the monthly evaluation registered with `queue.Handle` and no options, so vukem's default 1-minute timeout cancelled them every night (`docs/audit/02-technical-review.md` T-08).

- Every `queue.Handle` registration sets `queue.Timeout(...)` and `queue.Retries(...)` explicitly, sized to the real run (`vukem/queue` defaults are 1 minute and 3 retries and are never relied on). Say in the ticket why that timeout is enough.
- A job that can run longer than 15 minutes is chunked: it persists a cursor (`job_runs.cursor` or a small `job_cursors` row), does one chunk per execution and re-dispatches itself; a restart or a timeout then loses one chunk, not the night.
- Handlers are idempotent on retry: upserts keyed by a natural key, content hashes before uploads, delivery rows claimed with `UPDATE ... WHERE state = 'pending'` before a send, counters written with `ON CONFLICT ... DO UPDATE` only once per (key, period). Write the test that runs the handler twice.
- Honour `ctx`: check `ctx.Err()` between chunks and inside loops that touch the network; return `queue.ErrPermanent` for errors a retry cannot fix (4xx other than 429, schema failures).
- Retries that call external hosts use exponential backoff with full jitter computed from `queue.Attempt(ctx)` and re-dispatched with `queue.Delay` (vukem's `Exponential` has no jitter).
- Anything that calls an external service runs behind a feature flag or kill switch (`flags.Bool(...)`, default off in tests and production until the owner enables it) and goes through the per-host throttle and breaker of `internal/ingest/httpx`.
- Scheduled tasks use `WithoutOverlapping()`, name the time zone when the hour matters (`.In(saoPaulo)`), and stay clear of the 04:00-05:30 window reserved for scores, prunes and data-quality asserts unless they belong there (`routes/schedule.go` comments list the slots).
- A new table written by a job gets a retention window and a prune task in the same change (SDD 11 §5 lists them); user-linked rows also get an erasure entry in the account purge (`app/accounts/deletion.go`).
- Alerts are log lines `alert=A<n>` with the codes of SDD 12 §6; the Telegram notifier picks them up. New codes go into SDD 12 first.
- Production is never a test bed: new jobs run on staging with the DemoCatalog seed before the flag is turned on in production.
