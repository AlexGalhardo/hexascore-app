# Cutover rollback (R1–R3)

Rollback is allowed until T+7 d (SDD 11 §7.6). After that, fix forward. Nothing here runs by itself, and nothing here touches Railway, Neon, DNS or Stripe.

## R1 — DNS back to Vercel

Owner gate. Follow `docs/runbooks/dns-legacy.txt` when that file is current. Abort if the legacy origin does not answer `GET /up` within the TTL (60 s at T-24 h).

## R2 — Neon writable again

Owner gate. On the legacy Neon database:

```sql
ALTER DATABASE <db> SET default_transaction_read_only = off;
```

Terminate leftover sessions only after the alter commits. Do not run R3 before this.

## R3 — replay writes made after the cutover

From a checkout of the app, against the local or staging app database (read-only is enough):

```sh
/app/app hexascore:export-since 2026-10-30T15:00:00Z \
  | TARGET_DATABASE_URL="$NEON_URL" LEGACY_SCHEMA=public scripts/legacy/replay.sh
```

The instant is the cutover, RFC3339. The command prints one upsert per line. `replay.sh` applies it with `psql` and refuses a schema name that is not a bare identifier.

What lands in the legacy shape:

- ratings, item status, time-to-beat and lists, upserted by natural key (`user_id` + `item_id`, list `id`, list item `list_id` + `item_id`)
- API keys created after the cutover, keyed by `hash` (the secret is not stored)
- accounts with no `legacy_id`, inserted as `hs:<numeric id>` plus a `verification` row `reset-password:<email>` / `required` (the argon2id hash is not copied; Better Auth cannot read it)

Ratings are copied as stored. A multiple of 5 stays that multiple. Rows last written before the cutover are left alone.

Re-running R3 is safe: every statement is an upsert.

## Accepted loss

Sessions, suggestions, moderation rows and flag changes have no legacy table. They are not in the export. Stripe catches up when the legacy webhook is back. The command does not send the reset mail; an operator mails the addresses in `verification`.

## Abort

If `replay.sh` exits non-zero, the transaction rolls back. Leave Neon as it was, keep DNS on R1, and do not retry until the printed statement is understood.
