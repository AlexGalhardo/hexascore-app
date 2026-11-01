# Staging

Railway environment `staging` of project `hexascore` (created 2026-11-01, owner request): checks frontend changes and backend scripts on a small fake dataset. It never holds production data or production secrets, never sends mail, never charges, and is never indexed or tracked.

## Topology

| Resource | Value |
| --- | --- |
| URL | <https://app-staging-9215.up.railway.app> (Railway domain; `staging.hexascore.app` is an owner DNS step, P2.A.2.2) |
| `app` | same service as production, source `AlexGalhardo/hexascore-app` branch `main`, healthcheck `/up`, pre-deploy `/app/app migrate`, region `us-east4-eqdc4a` |
| `Postgres-cOb3` | Railway Postgres 18 template, own volume, private only (no TCP proxy); the app reads `DATABASE_URL=${{Postgres-cOb3.DATABASE_URL}}` |
| bucket | `hexascore-staging`, region `iad`, private; credentials in the `S3_*` variables |

## Variables

Set by name, values never in git or chat:

- `APP_ENV=production` (SDD 11 §6.4: vukem serves the built assets only in production) and `APP_URL` = the staging URL.
- `APP_KEY` and `STRIPE_WEBHOOK_SECRET`: random, generated in the shell and piped with `--stdin`.
- `SEO_NOINDEX=true`: `robots.txt` disallows everything and every response carries `X-Robots-Tag: noindex, nofollow` (needs the commit that adds it on `main`).
- `MAIL_MAILER=log`: mail goes to the deploy logs, nothing is sent.
- Billing off: `STRIPE_KEY=rk_test_staging_billing_disabled` and placeholder prices satisfy the production boot check; any Stripe call fails with an auth error, so no payment can happen. Replace with real Stripe test keys (`rk_test_`/`sk_test_` only) to test checkout.
- Turnstile: Cloudflare's published test keys (always pass); `TURNSTILE_HOSTNAME` unset.
- `FLAG_INGEST_ENABLED=false`, `FLAG_SOCIAL=false`; every other flag unset (off).
- Copied from production because they are configuration, not secrets: `APP_NAME`, `CACHE_STORE`, `CLIENT_IP_HEADER`, `LOG_FORMAT`, `RATE_LIMIT_STORE`, `SCHEDULE_TIMEZONE`, `SESSION_DRIVER`, `TRUSTED_PROXIES`, `TRUST_FORWARDED_PROTO`, `OTEL_LOGS_EXPORTER`, `OTEL_METRICS_EXPORTER`, `S3_USE_PATH_STYLE`, `S3_INSECURE_HTTP`.
- Never set on staging: `GA_MEASUREMENT_ID`, `UMAMI_*`, `GOOGLE_SITE_VERIFICATION`, `BING_*`, `INDEXNOW_KEY`, `RESEND_API_KEY`, `SOCIAL_GOOGLE_*`, `LEGACY_*`, `BACKUP_SOURCE_*`, live Stripe keys, or any value read from production secrets.

## Deploy

A push to `main` deploys staging and production together (both follow `main`). To redeploy staging alone:

```bash
railway redeploy --project 51b50f39-b99d-4720-b177-e5b7f4f530d3 --service app --environment staging -y
railway deployment list --service app --environment staging --json   # wait for SUCCESS
curl -fsS "https://app-staging-9215.up.railway.app/up?checks=1"
```

## Seed

DemoCatalog (300 items, 75 per type, 52 users, 2000 ratings, entities, editions). Idempotent: a second run does nothing. Run from PowerShell (Git Bash rewrites `/app/app` into a Windows path):

```powershell
railway ssh --project 51b50f39-b99d-4720-b177-e5b7f4f530d3 --service app --environment staging -- /app/app db:seed --class DemoCatalog --force
railway ssh --project 51b50f39-b99d-4720-b177-e5b7f4f530d3 --service app --environment staging -- /app/app admin:revoke admin@example.test
railway ssh --project 51b50f39-b99d-4720-b177-e5b7f4f530d3 --service app --environment staging -- /app/app admin:revoke admin2@example.test
```

The demo password is public (in the repository), so the two demo admins lose the admin role after every seed. To moderate on staging, sign up with your own address and run `/app/app admin:grant <email>`. Demo users keep the public password: staging content is fake and disposable.

Reset: there is no wipe command for staging; to start over, delete the `Postgres-cOb3` volume data from the dashboard (owner) and redeploy, then seed again.

## Cost

Volume 50 GB allocated by the template (billed by use, cannot shrink). Stop the staging `app` between releases if usage passes US$24 (SDD 11 §6.5).
