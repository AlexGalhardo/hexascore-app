# Postgres cost reduction (owner-run)

**Owner-run.** Nothing here is automated or executed by an agent. Every step touches Railway (production) and needs the owner's explicit go. Decision behind it: [ADR 0027](../adrs/0027-stay-on-postgres.md) (stay on Postgres, SQLite deferred), which follows `sqlite-problems.md` option (c). Savings are the estimates of that document (the staging part of the US$4 to 6 a month in total is already realised by deleting staging), not measurements; re-check the Railway prices before starting. Statements marked [NOT VERIFIED] have not been checked against the Railway docs or a live project.

Staging is gone: the owner deleted the `staging` environment on 2026-11-07 ([ADR 0028](../adrs/0028-staging-removed.md)), which already removed its Postgres and the US$3.80 a month it cost. Folding staging into the production instance is no longer a step. The remaining levers are the pool, `shared_buffers` and the volume.

Rules for the whole runbook:

- Do one section at a time, in this order: measure, pool, volume.
- Take a fresh dump before any change that touches production data or the volume: `railway run --service app /app/app backup:run --kind daily`, then confirm the object in the bucket (see [restore](restore.md)).
- Every section ends with a rollback. Keep the previous value of every variable you change in your notes.
- Never print secrets. Variables are set by name, values are piped or referenced (`${{Postgres.PGPASSWORD}}`).

## 0. Measure first

Right-sizing needs numbers from production, not the plan figures of SDD 11.

1. Railway metrics of the `Postgres` service for 7 days: CPU, memory and disk (the ADR 0024 table is the previous snapshot: avg 0.006 vCPU, 0.26 GB, 5.55 GB of 10 GB used, peak 4.03 GB during an import).
2. Connections in use over a busy hour, from `railway connect Postgres`:

```sql
SELECT state, count(*) FROM pg_stat_activity WHERE datname = current_database() GROUP BY state ORDER BY 2 DESC;
SELECT max(numbackends) FROM pg_stat_database WHERE datname = current_database();
```

   Repeat it a few times at peak. `numbackends` is a point value; the alert A12 history in the logs (`alert=A12`) shows the real peaks.
3. Size of what can be removed or exported:

```sql
SELECT relname, pg_size_pretty(pg_total_relation_size(relid)) FROM pg_catalog.pg_statio_user_tables
ORDER BY pg_total_relation_size(relid) DESC LIMIT 15;
```

## 1. Source history (already scheduled)

`source_record_history` is exported to the bucket and its closed months are dropped after the retention window by the daily `source_records.export` job (ADR 0025, SDD 06 §7.5). It is scheduled in `routes/schedule.go` at 04:15 and needs `FLAG_INGEST_ENABLED=true` to do anything. Nothing to change in code. To confirm it runs in production:

1. Look for `source_records.export` in the Railway logs after 04:15 UTC-3.
2. List `exports/source_record_history/` in the bucket: one `YYYY-MM.ndjson.gz` object per closed month.
3. In SQL, `source_record_history` should hold only the current and recent partitions:

```sql
SELECT inhrelid::regclass FROM pg_inherits WHERE inhparent = 'source_record_history'::regclass;
```

If the job never ran, the pruning above does not happen and the volume keeps growing; fix that before shrinking anything.

## 2. Connection pool

The application default comes from vukem (`DB_MAX_OPEN` 10, `DB_MAX_IDLE` 5); production sets `DB_MAX_OPEN=20` and `DB_MAX_IDLE=10` (SDD 11 §5). The repository sets no other default, so the change is a Railway variable, not code. The floor is 2 (an ingest step plus the cover reservation; the ingest runner refuses lower).

1. Do section 0 first. Proceed only if the busiest observed `numbackends` from the app stays below 8.
2. Set `DB_MAX_OPEN=12` and `DB_MAX_IDLE=6` on `app` in the production environment, redeploy, watch A12 for 24 hours (the alert fires at 80 % of `DB_MAX_OPEN`, so it now fires at about 10 in use).
3. Fewer idle connections lower Postgres memory a little; it does not change the plan price by itself. The saving comes only if the `shared_buffers` change in step 4 lowers the memory or CPU you are billed for.
4. Optionally lower Postgres `shared_buffers` (SDD 11 §5 sets 256 MB, `max_connections=60`) if memory use stays under half of it: `ALTER SYSTEM SET shared_buffers = '128MB';` needs a restart of the Postgres service [NOT VERIFIED that Railway allows `ALTER SYSTEM`; ADR/SDD 11 §5 notes the same].

Rollback: set `DB_MAX_OPEN=20` and `DB_MAX_IDLE=10`, redeploy; `ALTER SYSTEM RESET shared_buffers;` and restart Postgres.

## 3. Volume size

Production data is about 3.6 GB planned (SDD 11 §5) against a 6 GB volume; ADR 0024 measured 5.55 GB used on a 10 GB disk at an import peak. Do this only after sections 0 and 1 show steady use well below the volume.

1. Compare `pg_database_size(current_database())` plus WAL to the volume. Keep at least 30 % headroom (alert A11 fires at 80 %).
2. Railway can grow a volume in place; shrinking is not supported [NOT VERIFIED]. A shrink therefore means a new Postgres service with a smaller volume and a restore: create it, `backup:restore` the latest dump into it ([restore](restore.md) steps 1 to 3), compare row counts of the largest tables, then switch `DATABASE_URL` on `app` and redeploy.
3. Keep the old service and volume for 7 days before deleting them.

Rollback: switch `DATABASE_URL` back to the old service. Writes made after the switch are lost on the old one, so do the switch in a quiet window and treat the old instance as read-only history.

## Order and expected effect

| Step | Estimated saving (US$ per month, from `sqlite-problems.md`) | Effort |
| --- | --- | --- |
| 3 volume and 1 history export | about 0.40 | 0.5 day |
| 2 pool, `shared_buffers`, CPU | about 1 to 2 | 0.5 day |

Record the real before and after figures of `railway usage` in this file when the owner has run a step.
