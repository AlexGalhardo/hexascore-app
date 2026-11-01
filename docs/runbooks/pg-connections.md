# R2: Postgres connections exhausted

Alert A12 (pool in use over 80 %, 8 of 10 by default; production sets `DB_MAX_OPEN=20`, `DB_MAX_IDLE=10`, SDD 11 §5).

1. Group sessions by state (Railway Postgres shell or `railway connect`):

```sql
SELECT state, count(*) FROM pg_stat_activity WHERE datname = current_database() GROUP BY state ORDER BY 2 DESC;
```

1. Terminate sessions idle in a transaction for over 60 s:

```sql
SELECT pg_terminate_backend(pid) FROM pg_stat_activity
WHERE state = 'idle in transaction' AND now() - state_change > interval '60 seconds';
```

1. Pause ingestion: disable the source in `/admin/ingest` (SDD 06).
2. Restart the service, or lower `DB_MAX_OPEN` in Railway variables and redeploy.
3. File a bug with the trace ids of the slow requests (Railway traces).
