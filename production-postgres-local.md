# Production catalogue in local PostgreSQL

Use a local snapshot for frontend testing. The local app runs page counters, sessions and background tasks; pointing it at the production database could write to production. The production service is `Postgres` in Railway project `hexascore` (`production`). It has no public TCP proxy.

`railway connect Postgres --environment production` opens an interactive `psql` shell. For a repeatable export, use `railway connect Postgres --tunnel-only`: it keeps a private SSH tunnel open so `pg_dump` can read through `127.0.0.1:15432`. The script below starts and closes that tunnel automatically. It never stores the production connection string.

## One-time prerequisites

1. Sign in with `railway login` in your own PowerShell terminal.
2. Install the PostgreSQL 18 client and make `pg_dump` available on `PATH`. Its major version must be 18, matching Railway Postgres. `psql` is optional.
3. Start Docker Desktop. The local `postgres` service uses port `127.0.0.1:5433` and the development-only credentials in `docker-compose.yml`.

## Export and restore

From the repository root in PowerShell:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\production-catalog-local.ps1 -Action Sync
```

`Sync` creates `.local-seeds/production-schema.dump`, `.local-seeds/production-catalog.dump` and a manifest, then replaces only the local `hexascore_prod_preview` database. The directory is explicitly ignored by Git. After a Docker reset, restore the saved files without contacting Railway:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\production-catalog-local.ps1 -Action Restore
```

To refresh the files without touching the local preview database, use `-Action Export`. The script reads the Railway database URL into process memory, opens an SSH tunnel, exports the schema and selected catalogue tables with `pg_dump`, then removes the temporary tunnel output. It stops if any export or restore command fails. The local seed contains catalogue items, translations, entities, editions, offers and aggregate popularity/activity; it excludes users, ratings, sessions, billing, API keys and other account data. Scores already stored on `items` remain available for display. It does not copy cover images from object storage; cover URLs still point to the site's cover service.

The restore builds a fresh local database from the production schema, loads the catalogue rows before adding foreign keys and indexes, and reports the restored item count. Repeating `Restore` replaces only `hexascore_prod_preview`; it leaves the regular `hexascore` demo database intact.

## Preview the site

Run the app against the local preview database in a separate terminal. Leave the existing demo server on port 3000 if you want to compare them.

```powershell
$env:DATABASE_URL = 'postgres://hexascore:hexascore@127.0.0.1:5433/hexascore_prod_preview?sslmode=disable'
$env:SCHEDULE_IN_SERVE = 'false'
$env:QUEUE_CONNECTION = 'sync'
$env:FLAG_INGEST_ENABLED = 'false'
$env:COVERS_BASE_URL = 'https://hexascore.app'
$env:PORT = '3001'
$env:APP_URL = 'http://localhost:3001'
go run ./cmd/app serve
```

Open <http://localhost:3001/>. The preview has public catalogue data and no production accounts. Stop the process with Ctrl+C. Do not put the production `DATABASE_URL` in `.env` or `.env.example`.

