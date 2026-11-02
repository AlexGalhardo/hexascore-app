# R6: Disk full

Alert A11 (Postgres volume over 80 %, Railway monitor). Never resize the volume without the owner's yes (`AGENTS.md`).

1. Find the large relations:

```sql
SELECT relname, pg_size_pretty(pg_total_relation_size(oid)) FROM pg_class
WHERE relkind = 'r' ORDER BY pg_total_relation_size(oid) DESC LIMIT 15;
```

1. Prune the high-churn event tables that have a retention rule (`search_events`, `rating_events`) with the documented retention job; do not delete other tables by hand.
2. Rebuild a bloated search index: `/app/app search:reindex` (`railway ssh`) rewrites every SDD 07 §5 index with `REINDEX INDEX CONCURRENTLY` and rebuilds the suggest rows.
3. Still over 80 %: take a budget decision ([cost](cost.md)); the volume resize is the owner's.

## Source record history partitions

`source_record_history` is partitioned by month (ADR 0025). Check what exists and how big each month is:
`SELECT c.relname, pg_size_pretty(pg_total_relation_size(c.oid)) FROM pg_inherits i JOIN pg_class c ON c.oid = i.inhrelid WHERE i.inhparent = 'source_record_history'::regclass ORDER BY 1;`.
The job `source_records.export` (daily 04:15, needs `ingest_enabled`) exports closed months to `exports/source_record_history/YYYY-MM.ndjson.gz` and drops the ones older than 30 days. To free space by hand, first confirm the object exists in the bucket, then `DROP TABLE source_record_history_YYYY_MM;` for that month (owner's yes in production). If the job reports "updated partition constraint for default partition would be violated", rows of that month are in `source_record_history_default`: move them with `INSERT INTO source_record_history SELECT ... ; DELETE ...` after detaching the default, or export and delete them.
