# Changelog

All notable changes to this project are documented here.
Format: [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/). Versioning: [SemVer](https://semver.org/).

## 1.0.0 (2026-10-03)


### Added

* add error boundaries, auth route, session helper and namespaced messages ([26970a7](https://github.com/AlexGalhardo/claude-code-internet-api/commit/26970a764b8c2052b8627a793e63cb453f6a8211))
* **api:** add /profile/api with keys, usage chart and try-it console ([b343dda](https://github.com/AlexGalhardo/claude-code-internet-api/commit/b343ddafbdfbce76d261bd8d9b1db05a63b14e52))
* **api:** add API key management, quota with credits and RFC 9457 errors ([21b0043](https://github.com/AlexGalhardo/claude-code-internet-api/commit/21b0043b3593ac2864df6ef7a6e11c8e5de63fd4))
* **api:** add public /v1 catalog endpoints with zod response contracts ([4279770](https://github.com/AlexGalhardo/claude-code-internet-api/commit/42797703a13b93bd5cef099f81945363962949f3))
* **api:** generate openapi.json from the zod contracts and check it in CI ([3448392](https://github.com/AlexGalhardo/claude-code-internet-api/commit/3448392054cc3e366e71349ea9e227f9b37d48fd))
* **api:** serve /v1 on the api.&lt;domain&gt; host via the proxy ([e8a74d5](https://github.com/AlexGalhardo/claude-code-internet-api/commit/e8a74d596c366f2b62028ce8bcef2c802a611707))
* **auth:** add account purge, contact, rate limit, botid and password rule helpers ([a758239](https://github.com/AlexGalhardo/claude-code-internet-api/commit/a7582398faa6185ed8523c51493fbc27ddf63d3e))
* **auth:** add account server actions, contact action, purge cron and BotID setup ([29fa9f7](https://github.com/AlexGalhardo/claude-code-internet-api/commit/29fa9f7628d7d87f9df8ff42cf65ba7323555d99))
* **auth:** add login, signup, verification, recovery, profile, contact and legal pages ([d5c8b85](https://github.com/AlexGalhardo/claude-code-internet-api/commit/d5c8b85246465c8a9dce70206823222564ee5589))
* **billing:** add price catalog, Stripe and AbacatePay webhook processing with idempotent credits ([0e6f8d1](https://github.com/AlexGalhardo/claude-code-internet-api/commit/0e6f8d1efe92b9664ac297792012ed1555654104))
* **billing:** add pricing and checkout pages, billing actions and Stripe setup script ([205bfb6](https://github.com/AlexGalhardo/claude-code-internet-api/commit/205bfb6bc34b3434d57ff193bad9305e235951e1))
* **billing:** add webhook routes, PIX status polling and daily renewal reminder cron ([6be5f92](https://github.com/AlexGalhardo/claude-code-internet-api/commit/6be5f920c5ef502311c62b9b8b1d5805e9944834))
* **catalog:** add catalog, item and search pages with score components ([ba9bdc8](https://github.com/AlexGalhardo/claude-code-internet-api/commit/ba9bdc8b40da01a1242b207b4ae684e31259256a))
* **catalog:** add rating categories, catalog repository and dev seed ([c34ab33](https://github.com/AlexGalhardo/claude-code-internet-api/commit/c34ab3318db85e74fe3208e71753fa975f6f4cc3))
* **db:** add Drizzle schema, migrations, auth config and zod env validation ([04c1c12](https://github.com/AlexGalhardo/claude-code-internet-api/commit/04c1c12ee414327c22e7587952e461be3ad5b5cf))
* **db:** type cover attribution in item meta ([225174e](https://github.com/AlexGalhardo/claude-code-internet-api/commit/225174eba93065f279f27ae6ce145561283e7406))
* **docs:** add /docs guides and Scalar API reference ([5d6c4dd](https://github.com/AlexGalhardo/claude-code-internet-api/commit/5d6c4ddb2c239e35a29950650545a6f4d7da2514))
* **env:** accept Vercel Upstash and AbacatePay sandbox variable names ([350b176](https://github.com/AlexGalhardo/claude-code-internet-api/commit/350b1766947fa53b054fefd1b971879d14ec8f1c))
* **i18n:** add auth, account, contact and legal messages in en, pt-BR and es ([9d0de22](https://github.com/AlexGalhardo/claude-code-internet-api/commit/9d0de2262eafaa40d01343cdade89d7916bb979e))
* **infra:** add local Docker stack with Postgres 18, Redis and Upstash REST proxy ([486d3e8](https://github.com/AlexGalhardo/claude-code-internet-api/commit/486d3e8b725826d56a9850e5b4ecdab7dea15f1c))
* **ingest:** add normalizer, covers, revalidation and bootstrap with snapshot export ([7887c69](https://github.com/AlexGalhardo/claude-code-internet-api/commit/7887c692ea5d2f4c427e7fc8d1e185a49e4956c7))
* **ingest:** add uv project with http and db helpers ([617460b](https://github.com/AlexGalhardo/claude-code-internet-api/commit/617460b350f989b88195a3b493df11a41c50c6b7))
* **ingest:** add Wikidata, Wikipedia, Open Library, Steam and GOG connectors ([970927f](https://github.com/AlexGalhardo/claude-code-internet-api/commit/970927f9013806520ea2f11b49f9d0b144fc6614))
* scaffold Next.js 16 app with i18n, theme and layout ([ce8786c](https://github.com/AlexGalhardo/claude-code-internet-api/commit/ce8786c8bed5a657a9213f75059dc17091e97031))
* **sdk:** generate the @hexascore/sdk TypeScript client from openapi.json ([4578dda](https://github.com/AlexGalhardo/claude-code-internet-api/commit/4578dda451a6a37065eca9bd31fa988450795aef))
* **sdk:** generate the hexascore Python client and readable sort enums ([8d39a03](https://github.com/AlexGalhardo/claude-code-internet-api/commit/8d39a03d1b423cab130bba4a8cd1e51ff0fa6b06))
* **setups:** add idempotent local setup scripts for Docker and PGlite ([b399413](https://github.com/AlexGalhardo/claude-code-internet-api/commit/b399413ab0cadd9ba7fc94df106e955e44f090af))


### Fixed

* **api:** fail closed on store errors, load Scalar styles, disable Scalar AI ([3be8576](https://github.com/AlexGalhardo/claude-code-internet-api/commit/3be8576da8e2e719a44dcf68be8309d7da23e4a3))
* **api:** open the try-it reference only when a key is applied ([10e072e](https://github.com/AlexGalhardo/claude-code-internet-api/commit/10e072e80275a8c62e8e032e8b2a1dc72993f4e8))
* **auth:** typed redirects and clear signup field errors on edit ([4593413](https://github.com/AlexGalhardo/claude-code-internet-api/commit/459341350695b22dfa8488a92fcb8791eab5b1ec))
* **deps:** override js-yaml to 4.3.2 to clear high advisories ([4bf25f6](https://github.com/AlexGalhardo/claude-code-internet-api/commit/4bf25f6288fb012f86487c54d7f6c72cdd72e75f))
* **email:** pass template props through createElement so dev logs print the action link ([46436d6](https://github.com/AlexGalhardo/claude-code-internet-api/commit/46436d6c9b44967d2ad8ebf4d59f652157c9465f))
* **i18n:** pluralize the API rate limit label ([8acbb68](https://github.com/AlexGalhardo/claude-code-internet-api/commit/8acbb6894f339028835dbac752d117bc3eeddc30))
* **infra:** bind local ports to loopback and password-protect Redis ([802978c](https://github.com/AlexGalhardo/claude-code-internet-api/commit/802978c34381a84daaeb4ea13be9f858c2e9179d))
* **ingest:** send Vercel automation bypass header on revalidation ([ea4dbcc](https://github.com/AlexGalhardo/claude-code-internet-api/commit/ea4dbcc5af85181e0057df7f0c46d71d8dc5cf9f))
* **security:** require auth secrets on every production build and stop logging email tokens ([b590253](https://github.com/AlexGalhardo/claude-code-internet-api/commit/b590253841c8127bfc50c58d40c3c0dd38066024))
* **web:** localize every page, externalize PGlite and keep build workers off the dev database ([b7e2318](https://github.com/AlexGalhardo/claude-code-internet-api/commit/b7e23184f8934198cbb86be91686c0a58990262e))
* **web:** match localized routes in proxy and keep PGlite out of the server bundle ([898655b](https://github.com/AlexGalhardo/claude-code-internet-api/commit/898655b310acd2680cf2c5b75406d2ef7e26ff96))
* **web:** route locale-less paths through the i18n proxy and run PGlite under Next ([1e216b1](https://github.com/AlexGalhardo/claude-code-internet-api/commit/1e216b1085f189903d2eb72c344e6d9b1a0fa5dd))


### Changed

* reuse the hashed constant-time cron check ([845f2b4](https://github.com/AlexGalhardo/claude-code-internet-api/commit/845f2b451025962be0ec03e016d2f462d0b39896))

## [Unreleased]

### Added

- Public API v1 (0.4): `GET /api/v1/{games|movies|series|books}` (filters, `min_<category>`, sorts, cursor, plan-bound `limit`), `/{type}/{id}`, `/{type}/genres`, `/search`, `/lookup` (imdb/steam/igdb/tmdb/isbn13/openlibrary/wikidata) and `/me/usage`, all through `apiRoute` with zod response contracts (`lib/api/schemas.ts`); `api.<domain>/v1` served via the proxy.
- OpenAPI 3.1 generated from the zod schemas (`lib/api/openapi.ts`): `bun run openapi` writes `apps/web/public/openapi.json`, `bun run openapi:check` runs in the CI `lint` job.
- `/docs` (quickstart, authentication, errors, pagination, rate limits, changelog) and `/docs/reference` with Scalar (`@scalar/api-reference-react` 0.9.77); `api` i18n namespace; navbar "API" link.
- `/profile/api`: API keys (create once-shown key with copy, revoke with confirmation), plan/quota, today + 30-day usage SVG chart, in-browser "Try it".
- Generated SDKs: `sdks/ts` (`@hexascore/sdk`, `@hey-api/openapi-ts` 0.99.0, `bun run sdk:generate`) and `sdks/python` (`hexascore`, `openapi-python-client` 0.29.1 via uvx, `bun run sdk:generate:python`); not published yet.
- Catalog repository: `getItemIdByExternalId`, `listGenreCounts`, `wikidataQid`/`synopsisLocale` on `ItemDetail`; quota store `hget`.
- Route-level API tests (`lib/api/v1.test.ts`) and project skill `hexascore-api`.

### Changed

- `apiRoute` takes `{ params, query, response }`, validates path params (404), parses responses, maps bad cursors to 400, fails closed (503) when the rate-limit store errors and updates `api_keys.last_used_at` (at most once a minute).

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
