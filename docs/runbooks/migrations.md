# Runbook: migrations

Owner documents: SDD 11 §7.2, SDD 07 §7.10, ADR 0014. Migrations are forward-only in production: expand, migrate, contract two releases later.

## Run migrations

`migrate:locked` is `migrate` under a session-level Postgres advisory lock (`pg_advisory_lock(hashtext('hexascore:migrate'))`). vukem's runner takes no lock itself (VF-026), so two overlapping runners would race; with the lock the second waits, then prints `Nothing to migrate.`. Both exit 0.

```sh
./bin/app migrate:locked
```

The lock lives on a dedicated connection that is closed after the run, so a crashed runner never leaves the lock held. PostgreSQL only; any other dialect is an error.

The Railway pre-deploy command is still `/app/app migrate`. Switching it to `/app/app migrate:locked` is a deploy configuration change for the owner.

## Concurrent indexes

`CREATE INDEX CONCURRENTLY` cannot run inside a migration transaction (VF-017). Put the statement in `database/concurrent/<name>.sql` (index name `[a-z0-9_]+`), keep a regular migration with `CREATE INDEX IF NOT EXISTS` so `migrate:status` records it, and run after `migrate`:

```sh
PSQL='psql "$DATABASE_URL"' scripts/concurrent-index.sh <name>
```

The script drops an `INVALID` leftover of a failed earlier build, builds the index outside a transaction and exits non-zero unless `indisvalid` is true. `WATCH_LOCKS=1` samples `pg_locks` during the build and reports whether an `AccessExclusiveLock` on the table was ever granted; more than 5 s of `ACCESS EXCLUSIVE` blocks the release.

## Failure handling

- A failed migration rolls back its own transaction; fix forward with a new migration.
- Stale lock: closing the session releases it. `SELECT pid, state FROM pg_stat_activity WHERE query LIKE '%pg_advisory_lock%'` finds the holder.
