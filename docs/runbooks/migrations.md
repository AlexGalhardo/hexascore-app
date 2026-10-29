# Runbook: migrations

Owner documents: SDD 11 §7.2, SDD 07 §7.10, ADR 0014. Migrations are forward-only in production: expand, migrate, contract two releases later.

## Run migrations

`migrate:locked` is `migrate` under a session-level Postgres advisory lock (`pg_advisory_lock(hashtext('hexascore:migrate'))`). vukem's runner takes no lock itself (VF-026), so two overlapping runners would race; with the lock the second waits, then prints `Nothing to migrate.`. Both exit 0.

```sh
./bin/app migrate:locked
```

The lock lives on a dedicated connection that is closed after the run, so a crashed runner never leaves the lock held. PostgreSQL only; any other dialect is an error.

## Pre-deploy

`hexascore:predeploy` is the single pre-deploy entry point (SDD 11 §7.2). It drops `INVALID` index leftovers of a failed concurrent build, then runs `migrate`, printing `Dropped INVALID index: <name>`, `Migrated: <id>` or `Nothing to migrate.`. Any error exits non-zero, so Railway does not proceed with the deployment and the previous one keeps serving.

```sh
/app/hexascore hexascore:predeploy
```

Since vukem v0.33.0 `migrate` itself holds an advisory lock for the whole run (VF-026): two overlapping pre-deploys serialise and the second prints `Nothing to migrate.`. The Railway pre-deploy command is still `/app/app migrate`; switching it to `/app/app hexascore:predeploy` is a deploy configuration change for the owner.

## Concurrent indexes

`CREATE INDEX CONCURRENTLY` cannot run inside a transaction. Since vukem v0.33.0 (VF-017) it is a regular migration with `NoTransaction: true`, one statement per migration, written to be re-run:

```go
var IndexItemsTitleTrgm = schema.Migration{
	ID:            "20261101120000_index_items_title_trgm",
	NoTransaction: true,
	Up: func(s *schema.Schema) error {
		return s.Raw("CREATE INDEX CONCURRENTLY IF NOT EXISTS items_title_trgm ON items USING gin (title gin_trgm_ops)")
	},
	Down: func(s *schema.Schema) error {
		return s.Raw("DROP INDEX CONCURRENTLY IF EXISTS items_title_trgm")
	},
}
```

A failed build (timeout, unique violation) leaves an `INVALID` index and the migration unrecorded; `IF NOT EXISTS` would skip that leftover on the retry. `hexascore:predeploy` therefore drops every `INVALID` index of the current schema that no backend is building (`pg_stat_progress_create_index`) before migrating; partitioned parent indexes are left alone. Ship a big index build alone in its release so it fits the pre-deploy timeout (SDD 11 §8). The P0.F.5.2 workaround (`database/concurrent/`, `scripts/concurrent-index.sh`) is not used for new indexes.

## Failure handling

- A failed migration rolls back its own transaction; fix forward with a new migration.
- Stale lock: closing the session releases it. `SELECT pid, state FROM pg_stat_activity WHERE query LIKE '%pg_advisory_lock%'` finds the holder.
