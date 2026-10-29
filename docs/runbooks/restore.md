# Restore runbook

Restores an encrypted `backup:run` dump (SDD 11 §7.3) into a new Postgres database and re-applies the account deletions made after the dump (SDD 13 §8, open issue 11). Design: SDD 11 §7.4 (monthly drill) and §7.5 (RTO steps). Timings of the game day land here with P2.E.3.2.

## What a dump consists of

- `backups/postgres/<kind>/YYYY/MM/DD/hexascore-<env>-<UTC stamp>.dump.enc`: `pg_dump -Fc` sealed in 4 MiB frames with `BACKUP_KEY`.
- `backups/manifests/postgres/<kind>/.../<same name>.dump.enc.json`: row count of every `public` table, counted in the snapshot `pg_dump` exported, sealed with the same key. Daily dumps have one; monthly copies and dumps older than P2.E.3.1 do not. `backup:restore` does not need it; `backup:drill` does.

## Restore (RTO step 3)

1. Create the new, empty database (Railway: new Postgres service in the same region).
2. Pick the dump: the newest key under `backups/postgres/daily/` (bucket listing, or `backup_runs` where `status = 'verified'`).
3. Run, on a container of the app image (it ships `pg_restore`):

   ```sh
   /app/app backup:restore backups/postgres/daily/<path>.dump.enc --target-url "$NEW_DATABASE_URL"
   ```

   - Environment: `S3_*` of the bucket holding the dump, `BACKUP_KEY` that sealed it, `DATABASE_URL`.
   - The dump streams through decryption into `pg_restore --no-owner --no-acl --exit-on-error`; nothing plaintext touches disk. One job only: `-j` needs a seekable file.
   - Then the deletion replay: users anonymised, scheduled for deletion or cancelled in `DATABASE_URL` (or `--replay-from <url>`) after the dump get the same state in the target, without mail. Output: `Replayed deletions: N anonymised, N scheduled, N cancelled`. Re-running it is a no-op.
   - The old database is gone (lost service): pass `--no-replay`, then re-apply by hand the deletions recorded after the dump (support mail, Stripe customer-deletion events). This is the only path where a deleted account can come back; log it in the incident.
4. Point `DATABASE_URL` at the new database and redeploy; the pre-deploy migrates (SDD 11 §7.5 step 4).
5. `hexascore:cutover-verify --counts-only`, `catalog:check`; `up`.

## Monthly drill

`backup:drill` runs on staging only, first Sunday 07:00 America/Sao_Paulo (scheduled when `BACKUP_SOURCE_S3_BUCKET` is set; staging also runs `APP_ENV=production`). It reads `BACKUP_SOURCE_<name>` first and falls back to `<name>`, so `BACKUP_SOURCE_S3_*` and `BACKUP_SOURCE_BACKUP_KEY` point it at production's bucket. Granting staging read access to the production bucket is the owner's call.

Steps: latest daily dump and its manifest → `DROP DATABASE IF EXISTS hexascore_drill WITH (FORCE)`, `CREATE DATABASE hexascore_drill` on the staging server → restore → row counts equal to the manifest → `catalog:check` asserts (SDD 01 §7.6 plus scoring; SDD 06 asserts join in P6) with 0 errors → drop → `backup_runs(kind='drill')` `verified`, or `failed` with the reason. Pass also needs `duration_s < 3600` (FR-11-6). The drill never restores over the staging database.

Manual run: `/app/app backup:drill` (without `--first-sunday` it runs any day). A mismatch prints one `count mismatch <table>: want N, got M` line per table.
