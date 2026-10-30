# Runbook: migrations

Owner documents: SDD 11 §7.2, SDD 07 §7.10, ADR 0014. Migrations are forward-only in production: expand, migrate, contract two releases later.

## Run migrations

`migrate` takes a session-level advisory lock (vukem v0.33.0, VF-017/VF-026/VF-040), so two overlapping runners serialise: the second waits, then prints `Nothing to migrate.`. PostgreSQL, MySQL and SQLite are covered by the runner; a crashed runner releases the lock with its connection.

```sh
./bin/app migrate
```

## Pre-deploy

`hexascore:predeploy` is the single pre-deploy entry point (SDD 11 §7.2). It drops `INVALID` index leftovers of a failed concurrent build, then runs `migrate`, printing `Dropped INVALID index: <name>`, `Migrated: <id>` or `Nothing to migrate.`. Any error exits non-zero, so Railway does not proceed with the deployment and the previous one keeps serving.

```sh
/app/hexascore hexascore:predeploy
```

Since vukem v0.33.0 `migrate` itself holds an advisory lock for the whole run (VF-026): two overlapping pre-deploys serialise and the second prints `Nothing to migrate.`. The Railway pre-deploy command is still `/app/app migrate`; switching it to `/app/app hexascore:predeploy` is a deploy configuration change for the owner.

## Expand / contract policy

`scripts/ci/migration-policy.ts` reads migration source. It rejects, in the `Up` text only, `DROP COLUMN`, `DROP TABLE`, `RENAME`, `SET NOT NULL`, and a non-concurrent `CREATE INDEX` on `items`, `ratings`, or `item_translations`. `CREATE INDEX CONCURRENTLY` is allowed. A `contract: vX.Y.Z` marker waives the guard when that release is at least two minors older than the current app release (`v0.8.0` accepts `contract: v0.6.0`, not `v0.7.0`). `Down` is never checked: production does not run down migrations. CI runs `bun test scripts/ci/migration-policy.test.ts`. A PR that adds a migration also runs `vukem migrate`, `vukem migrate:rollback --step 1`, and `vukem migrate` on Postgres 18.6.

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
