# R3: Ingestion stuck

Alerts A6 (an ingestion job kind failed twice in a row) and A7 (failed run, no incremental in 26 h, failing assert). Design: SDD 06 §10, ADR 0007. Runs live in `job_runs`.

1. Find runs past their timeout (the job timeout is 10 minutes):

```sql
SELECT id, job, source, mode, status, attempt, started_at FROM job_runs
WHERE status = 'running' AND started_at < now() - interval '15 minutes';
```

1. Check `/admin/ingest` for failures, the breaker state and rate-limited counts, and the logs for `alert=A6`.
2. Mark the stuck run failed: `UPDATE job_runs SET status = 'failed' WHERE id = <id>;`.
3. Re-dispatch from `/admin/ingest`; the run resumes from its `cursor`.
4. If the upstream source is down, disable it in `/admin/ingest` and note it on the status page.

## Related situations

- **`/up?checks=1` shows `worker: fail`** (only when `WORKER_HEARTBEAT_CHECK` is on): the process that consumes the `ingest` queue stopped beating. Look at the worker service logs, redeploy it, and see [worker](worker.md). Until it is back, jobs wait in the queue; nothing is lost. To fall back, set the web's `QUEUE_NAMES` to `default,ingest`.
- **Items missing after a normalise run**: some were held because they duplicate another item. Open `/admin/ingest/conflicts`; Allow creates the item on the next run (without the contested external id), Dismiss keeps it out. Merging two existing items is done from the item page, never by the worker.
- **`alert=A18`** (LLM budget): the month's spend reached `LLM_MONTHLY_BUDGET_USD` and the normaliser stopped until the next UTC month. Nothing is broken. Raise the ceiling only after reading the provider's current prices; every value the model produced waits in `/admin/ingest/field-reviews`.
- **Offer links look wrong**: `offers.verify` marks a link `broken` on its first 404 or 410 and deactivates it on the second. The item's manage page lists every offer with its check status and lets you restore or replace one. TMDB watch providers are refreshed from the API every 30 days and expire after 180 (ADR 0006).
- **`cover_missing` in `ingest_failures`**: `refresh.existing` found a cover key whose object is not in the bucket. The covers job re-fetches it on its next run (`covers-daily` or `covers-weekly`); run it from `/admin/ingest` to fix it sooner.
