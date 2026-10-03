# Changelog

All notable changes to this project are documented here.
Format: [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/). Versioning: [SemVer](https://semver.org/).

## [Unreleased]

### Added

- Billing (0.5): price catalog `lib/billing/catalog.ts` (Pro/Business/credits in BRL/USD/EUR, PIX 12 months for the price of 10, currency by locale).
- Stripe (`stripe@23.0.0`, API `2026-09-30.endive`): `bun run stripe:setup` creates Products/Prices by `lookup_key` and the Customer Portal config; `startStripeCheckout`/`openBillingPortal` actions; `POST /api/webhooks/stripe` (signature, idempotency via `processed_events`, subscriptions and credit packs).
- AbacatePay PIX (API v2): `createPixCharge` action, `GET /api/billing/pix/[id]` polling with provider re-check, `POST /api/webhooks/abacatepay` (URL secret + HMAC, idempotent, period stacking).
- Credits purchase writes `credit_ledger` once per `ref` and bumps the cached Redis balance (`addCachedCredits` in `lib/api/quota.ts`).
- `stripe_customers` table (migration `0002_stripe_customers`).
- Daily cron `/api/cron/billing-reminders`: PIX renewal emails 7d/1d before the end and expiration of ended PIX periods.
- Pages `/pricing` and `/profile/api/checkout` (Base UI radios, PIX QR + copy-paste + live status), `billing` i18n namespace (en, pt-BR, es), "payments not configured" fallbacks.
- Project skill `hexascore-billing` (local webhook testing, idempotency rules).

- `PLAN.md` with decisions and milestones; reference docs in `docs/` (architecture, data model, API, frontend, design system, payments, ingestion, security, testing, process).
- `AGENTS.md` and `CLAUDE.md` agent guides.
- Catalog ingestion in `webscrappings/` (Python 3.14 + uv): Wikidata, Wikipedia, Open Library, Steam and GOG connectors into `source_records`, normalizer to `items`/`item_translations`/`external_scores`, bootstrap with gzipped SQL snapshot export, covers (Vercel Blob optional), revalidation call, pytest suite with recorded fixtures.
- GitHub Actions: `ingest-bootstrap.yml` (manual), `ingest-incremental.yml` (nightly), `ingest-weekly.yml` (scores and synopses), `python-ci.yml` (ruff + pytest on PRs touching `webscrappings/**`).
- `infra/docker-compose.yaml`: Postgres 18.6, Redis 8.10.2 (password-protected, not published to the host) and an Upstash-compatible REST proxy (`serverless-redis-http` 0.0.10, port 8079) with healthchecks; ports bound to 127.0.0.1 and credentials documented as dev-only; optional `app` profile running the web app from `infra/Dockerfile` (`oven/bun:1.4.2`, multi-stage, non-root, migrates on start).
- Idempotent local setup scripts in `setups/` (Unix/Windows × Docker/PGlite): check Bun/Docker, install, create/fill `apps/web/.env.local` with generated secrets, start containers, migrate, seed when available.
- CI workflow (`lint`, `typecheck`, `test`, `build`, `audit`) on PRs and pushes to `dev`/`main`, with a shared `setup-bun` composite action and Bun install cache.
- CodeQL workflow (JavaScript/TypeScript + Python), gated on `CODE_SCANNING_ENABLED` for private repos.
- release-please (`simple`, manifest at `0.0.0`, Keep a Changelog sections) on pushes to `main`.
- Root `README.md` with quickstart, commands and docs index.
- `.github/rulesets/protect-main-dev.json`: ruleset for `main`/`dev` (PR required, checks `lint`/`typecheck`/`test`/`build`, no force push/deletion), ready to apply once the GitHub plan allows it.

### Changed

- Dependabot now covers `bun`, `uv` (`webscrappings/`), `github-actions`, `docker` and `docker-compose` (`infra/`), grouped minor/patch, targeting `dev`.
- `docs/process.md` and `docs/testing.md` list the implemented workflows and exact commands.

### Fixed

- `proxy.ts` matcher skipped every path longer than one character (`.*..*` instead of `.*\..*`), so localized routes 404ed.
- PGlite is kept out of the server bundle (`serverExternalPackages`) so its WASM/extension files load during build and `next start`.
- Catalog (0.1): rating categories per type (`lib/categories.ts`), pure catalog repository (`lib/catalog/`) with keyset pagination, filters, `pg_trgm` + `tsvector` search, facets, external scores, affiliate links and public ratings; tagged `unstable_cache` layer (`catalog`, `item:<id>`).
- Pages: landing highlights per section, `/games` `/movies` `/series` `/books` with URL filters and cursor pagination, item page (score badge, hexagonal radar, users' scores dialog, external scores, click-to-load trailers, details, where to buy, synopsis attribution, `RatingPanel` slot), `/search`, navbar autocomplete.
- Route handlers: `GET /api/items/[id]/ratings`, `GET /api/search/suggest`, `POST /api/cron/revalidate` (Bearer `CRON_SECRET`).
- `components/score/` (`ScoreBadge`, `HexRadar`, `ScoreBar`) and i18n namespaces `categories`, `catalog`, `item`, `search` (en, pt-BR, es).
- Deterministic dev seed `bun run db:seed` (`@faker-js/faker` 10.6.0): 50 users, 400 items, ratings; refuses `VERCEL_ENV=production`.
- PGlite test helper `createTestDatabase()` and repository tests.

### Fixed

- i18n proxy matcher only matched `/` (unescaped dot), so locale-less routes like `/games` returned 404.
- PGlite under Next: listed in `serverExternalPackages` (bundling broke its WASM path) and in-memory during `next build` (parallel workers could not share `.pglite`).
- Auth & account (0.2): `/login` (email or nickname, Google when configured, passkey, TOTP/backup-code 2FA step, resend verification), `/signup` (live nickname avatar, password checklist), `/verify-email`, `/recover-password`, `/reset-password`, `/profile` (name, set/change password, 2FA with QR + backup codes, passkeys, private profile, account deletion with 10-second confirm and 30-day grace) and a cancel-deletion banner after login. All in EN, pt-BR and ES.
- Server Actions in `apps/web/actions/` (`account.ts`, `contact.ts`) returning typed `{ ok } | { ok: false, error }` results.
- `/contact` (session identity overrides the form, DB rate limit per user/IP, optional email to `CONTACT_INBOX`), `/policy` and `/terms` in 3 languages.
- Daily Vercel cron `/api/cron/purge-accounts` (Bearer `CRON_SECRET`, constant-time) that anonymizes accounts 30 days after scheduled deletion, keeping their ratings.
- Vercel BotID on signup, login and contact (`botid` 1.5.11), bypassed outside Vercel. `qrcode` 1.5.4 renders the TOTP QR in the browser.
- Navbar user menu (Profile, API, Log out) and Login/Sign up links.

### Fixed

- `proxy.ts` matcher had an unescaped dot, so no page outside `/pt-br`/`/es` was localized (`/login` returned 404).
- PGlite is now a server external package and `next build` workers use an in-memory instance, so dev/build no longer crash or corrupt `.pglite`.
- Dev email log printed `link=(none)`: templates are now passed with `createElement`.

### Removed

- Files inherited from the previous project (RespondeAê): old plan, docs, setup scripts, tests, REST client samples, graphify output, `vercel.json`.
- Inherited Prisma-era `ci.yml` and `e2e.yml` workflows (replaced by the new CI; E2E returns in a later milestone).
