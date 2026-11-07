# ADR 0028: The staging environment is removed (2026-11-07)

Status: accepted (2026-11-07, owner decision). Amends SDD 11 (environments, cost, release gate), the `staging` runbook and the places that named staging as the verification target. Related: ADR 0027 (stay on Postgres), `docs/runbooks/postgres-cost.md`.

## Context

The Railway environment `staging` (created 2026-11-01) held `app`, `worker-staging`, `Postgres-cOb3` and the bucket `hexascore-staging`. It cost about US$7.40 a month in SDD 11's sheet and was the stated home of rehearsals, the restore drill, k6 and ZAP runs.

## Decision

- The owner deleted the staging environment in Railway on 2026-11-07 (app, worker, Postgres, bucket). Production was not touched.
- The project has one deployed environment, `production`. Verification that used staging runs on the local QA stack (`docker compose --profile app up -d --wait`, seeded with DemoCatalog, see AGENTS.md "Local QA") or, for restores, in a scratch database.
- Kept on purpose, because they protect more than staging did: `SEO_NOINDEX` and the non-production `robots.txt` (previews, local production builds), `MAIL_MAILER=log`, Turnstile test keys and the other local-safe defaults of `.env.example`.
- `backup:drill` stays as a command but is not scheduled anywhere: it was gated on `BACKUP_SOURCE_S3_BUCKET`, which only staging had. A new host for the monthly drill (FR-11-6) is an open owner decision.
- Living documents were updated to say this; dated records (CHANGELOG releases, `docs/audit/`, `docs/review/`, earlier ADRs, `docs/runbooks/worker.md` status lines of 2026-11-03) are left as written.

## Pending on the owner

- Remove the DNS record `staging.hexascore.app` (Cloudflare).
- The remote branch `staging` holds PR #69 (shadcn/ui work, never merged to `main`); decide whether to keep or delete it.

## Consequences

- Release rehearsals with a migration (SDD 11 §7.2) have no shared environment: run them against a restored copy of the latest dump in a local or scratch database.
- The monthly restore drill is manual until a host is chosen.
- Monthly infrastructure cost falls by the staging rows of SDD 11 §6.5.
