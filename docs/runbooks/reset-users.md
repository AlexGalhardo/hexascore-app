# Reset users

Removes every account except two (the owner and one admin) and everything the removed accounts left on the site, in
one transaction. It is irreversible once committed, so the script is a dry run unless you ask for the commit.

Script: `scripts/ops/reset-users.sql`. Password hash helper: `scripts/ops/hashpw`.

## What it removes

- Every user except the two kept emails, and by cascade their ratings, reviews, comments, lists, follows, status,
  notifications, subscriptions, API keys, passkeys, sessions and exports.
- Their community suggestions with events and reviews (the catalogue audit rows stay, with the suggestion link cleared).
- Reports and contact messages, rating events, access and remember tokens, search events and password resets of
  removed accounts.
- Item scores that came only from removed ratings go back to unscored; the daily 04:00 recompute rebuilds the rest.

It keeps the catalogue, ingestion data, feature flags, the audit log and the kept accounts untouched. The admin
account is created when missing, or given the new password and the admin role when it exists.

## Before running

1. Back up: `railway ssh --service app -- vukem backup:run --kind daily`, and confirm the file exists (see
   [restore.md](restore.md)).
2. Stripe: the subscriptions of removed users are deleted here, not in Stripe. Cancel live customers and
   subscriptions in the Stripe dashboard first, or they keep being charged with no account behind them.
3. Create the hash on your machine, with the password in the environment so it never reaches the shell history:
   `PASSWORD='...' go run ./scripts/ops/hashpw`. Only the hash goes into the command below.
4. The admin panel needs two-factor authentication: sign in as the admin once and enable it before using `/admin`.

## Run

```sh
psql "$DATABASE_URL" -v owner_email=you@example.com -v admin_email=admin@example.com \
  -v admin_hash='<hash>' -f scripts/ops/reset-users.sql
```

Read the before and after counts: only the two kept users must remain, and the list at the end must show them.
When it is right, run the same command with `-v commit=1`. Use the public database URL from the Railway Postgres
service when running from your machine.

## After

Sign in with both accounts, open `/admin/overview`, and check the next scoring recompute (04:00 UTC) on `/admin/r/job-runs`.
