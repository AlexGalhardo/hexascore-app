---
name: hexascore-qa
description: Run Hexascore QA layers (Playwright + axe E2E in qa/e2e, security headers/CSP checks, scale seed) locally against next start and in CI (.github/workflows/qa.yml). Use before merging dev into main, when touching proxy.ts / security headers / CSP, or when adding a page, form or flow that needs E2E coverage.
---

# Hexascore QA

Reference: `docs/testing.md` → "E2E + a11y (0.6)". Layers that exist today: E2E + axe (`qa/e2e`), header/CSP unit tests (`apps/web/lib/security-headers.test.ts`). Schemathesis, ZAP, k6 and the exploratory QA agent are **not built yet** (0.6 gaps in `PLAN.md`).

## Local run (Docker Postgres, never PGlite with `next start`)

1. Own throwaway container on a free port (do not reuse `infra/` if another agent uses 5432): `docker run -d --name hexascore-qa-pg -e POSTGRES_USER=hexascore -e POSTGRES_PASSWORD=hexascore -e POSTGRES_DB=hexascore -p 127.0.0.1:55432:5432 postgres:18.6-alpine`.
2. In `apps/web` with `DATABASE_URL`, `BETTER_AUTH_SECRET`, `CRON_SECRET`, `BETTER_AUTH_URL`/`NEXT_PUBLIC_APP_URL=http://localhost:3100`: `bun run db:migrate`, seed (`bun lib/db/seed.ts --users 10000 --items 100000 --batch 1000` ≈ 7 min, or the default small seed), `bun run build`, `NODE_ENV=test bun --bun next start -p 3100` (in-memory API quota store).
3. `cd qa/e2e && E2E_BASE_URL=http://localhost:3100 E2E_DATABASE_URL=… E2E_BETTER_AUTH_SECRET=… bunx playwright test` (`--project=desktop-light`, `--last-failed`).

## Writing specs

- Import `test`, `expect`, `expectAccessible`, `LOCALES` from `helpers/fixtures.ts`: it presets the cookie banner choice (`test.use({ consent: null })` to see it) and fails the test on CSP violations / uncaught errors.
- Call `expectAccessible(page)` on every new page/state (serious/critical axe = fail). Exclude only third-party markup (Scalar).
- Account flows: tag the describe with `@auth`, skip without `hasAuthSecret()`, create users with `createVerifiedAccount()` and log in with `logIn()` (handles Better Auth's per-IP 429). Never add test-only routes or flags to the app.
- Email links: verification = JWT minted with the deployment's secret (`verificationUrl`), reset = `latestResetToken()` from the `verification` table.
- Never hard-delete users from tests (cascades ratings and skews stats); E2E accounts use `@e2e.example.test`.

## Gotchas

- Git Bash rewrites `/paths` in env vars and args: prefix commands with `MSYS_NO_PATHCONV=1`.
- After seeding or editing the DB behind a running server: `POST /api/cron/revalidate` `{"tags":["catalog"]}` with `Authorization: Bearer $CRON_SECRET` (the global setup also warms listings).
- CSP: pages are dynamic by design (nonce). New third-party script/frame/connect origins go in `lib/security-headers.ts` (scope route-specific ones like Scalar's) plus a test in `security-headers.test.ts`; `headers.spec.ts` checks the live response.
- `next build` output must show every `[locale]` route as `ƒ`; a static page would ship scripts without the nonce and break under `strict-dynamic`.

## CI

`.github/workflows/qa.yml` on PR `dev` → `main` and `workflow_dispatch` (`url` input): waits for the Vercel preview of the head SHA, runs the suite with `x-vercel-protection-bypass`. Secrets: `VERCEL_AUTOMATION_BYPASS_SECRET` (exists), `E2E_DATABASE_URL` + `E2E_BETTER_AUTH_SECRET` (needed for `@auth` specs). Advisory gate (branch protection off by owner decision) — do not merge with it red. Validate the workflow: `MSYS_NO_PATHCONV=1 docker run --rm -v "$(pwd -W):/repo" -w /repo rhysd/actionlint:1.7.12`.
