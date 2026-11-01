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
