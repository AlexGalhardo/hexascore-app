# Changelog

All notable changes to this project are documented here.
Format: [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/). Versioning: [SemVer](https://semver.org/).

## [0.0.15](https://github.com/AlexGalhardo/claude-code-internet-api/compare/v0.0.14...v0.0.15) (2026-10-03)


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

## [0.0.14] - 2026-10-03

### Fixed

- Override `js-yaml` to 4.3.2 to clear three high-severity advisories reported by `bun audit`.

## [0.0.13] - 2026-10-03

### Added

- Public API v1: list, detail, genres, search, lookup and usage endpoints with zod response contracts and `api.<domain>/v1` host rewrite.
- OpenAPI 3.1 generated from zod with a CI drift check; public docs and Scalar reference; `/profile/api` keys and usage dashboard.
- Generated TypeScript (`@hexascore/sdk`) and Python (`hexascore`) SDKs.

## [0.0.12] - 2026-10-03

### Added

- Login, signup, email verification, password recovery and reset pages.
- Profile with name, set/change password, 2FA with QR and backup codes, passkeys, private profile and 30-day account deletion with a 10-second confirmation and cancel banner; purge cron.
- Contact form with BotID and rate limit; privacy policy and terms in three languages.

## [0.0.11] - 2026-10-03

### Changed

- Single hashed constant-time check for cron authorization.

### Fixed

- Ingestion revalidation now sends the Vercel automation bypass header.

## [0.0.10] - 2026-10-03

### Added

- Catalog pages for games, movies, series and books with URL filters and cursor pagination.
- Item page with metascore, hexagonal radar, users' scores dialog, attributed external scores, click-to-load trailers and affiliate links.
- Search with fuzzy matching and keyboard-accessible autocomplete; tag-based cache revalidation endpoint; deterministic dev seed.

## [0.0.9] - 2026-10-02

### Added

- Billing: Stripe Checkout and Customer Portal (BRL/USD/EUR) and AbacatePay PIX (1 or 12 months, credits) with signed, idempotent webhooks.
- Pricing page, checkout page with live PIX status, credits ledger and renewal reminder cron.

### Fixed

- Proxy matcher that skipped localized routes and PGlite being bundled into the server build.

## [0.0.8] - 2026-10-02

### Added

- Environment aliases for the Vercel Upstash (`KV_REST_API_*`) and AbacatePay sandbox (`*_SANDBOX`) variable names.

### Changed

- Branch ruleset hardened: strict status checks, `audit` gate and no bypass actors.

## [0.0.7] - 2026-10-02

### Added

- Local Docker stack (Postgres 18.6, password-protected Redis, Upstash-compatible REST proxy, loopback-only ports) and a non-root `oven/bun:1.4.2` Dockerfile.
- Idempotent local setup scripts in `setups/` (Unix/Windows, Docker or PGlite).
- CI (lint, typecheck, test, build, audit), CodeQL, release-please, Dependabot and a ready-to-apply branch ruleset.
- Typed cover attribution in item metadata.

## [0.0.6] - 2026-10-02

### Added

- Python 3.14 + uv ingestion in `webscrappings/`: Wikidata, Wikipedia, Open Library, Steam and GOG connectors, normalizer, bootstrap with SQL snapshot export, covers, revalidation call and pytest suite.
- GitHub Actions workflows for bootstrap, nightly incremental and weekly ingestion, plus Python CI.

## [0.0.5] - 2026-10-02

### Added

- API core: hashed API keys, plan limits, Upstash-backed throttle and quota with prepaid credits, refunds on 5xx, RFC 9457 problem responses.
- API key create/revoke server actions and the daily usage flush cron.

## [0.0.4] - 2026-10-02

### Added

- Localized 404 and 500 pages plus a global error boundary.
- Better Auth route handler, `getSessionUser()` session helper and namespaced i18n message files.
- Production migrations during `vercel-build`; `fetchWithTimeout` helper for outbound calls.

## [0.0.3] - 2026-10-02

### Fixed

- Every production build (including Vercel previews) now requires its own auth and cron secrets.
- Development email logging no longer prints one-time tokens in production; password rules also apply to admin password endpoints.

## [0.0.2] - 2026-10-02

### Added

- Drizzle schema for catalog, ratings, social, billing and ingestion staging with initial migrations (Postgres extensions `pg_trgm`, `unaccent`).
- Database client using Neon/Postgres via `DATABASE_URL` or in-process PGlite for local development.
- Better Auth configuration (email + password, Google, username, 2FA, passkeys, admin) with password rules enforced server-side.
- Zod validation of all environment variables before build and server start, with explicit console errors.
- Pre-commit hook running Biome and `tsc --noEmit`; Vercel Functions on the Bun runtime.

## [0.0.1] - 2026-10-02

### Added

- Bun workspaces monorepo with `apps/web` on Next.js 16.3 running on the Bun runtime, TypeScript 7, Tailwind CSS 4 and Base UI.
- i18n with next-intl (EN default, `/pt-br`, `/es`), light/dark theme persisted in a cookie without flash, cookie consent banner, base layout with navbar and footer.
- Project plan (`PLAN.md`), agent guides (`AGENTS.md`, `CLAUDE.md`) and reference docs in `docs/`.

### Removed

- Files inherited from the previous project (old plan, docs, Prisma setup scripts, tests and workflows).

[Unreleased]: https://github.com/AlexGalhardo/claude-code-internet-api/compare/v0.0.14...HEAD
[0.0.14]: https://github.com/AlexGalhardo/claude-code-internet-api/compare/v0.0.13...v0.0.14
[0.0.13]: https://github.com/AlexGalhardo/claude-code-internet-api/compare/v0.0.12...v0.0.13
[0.0.12]: https://github.com/AlexGalhardo/claude-code-internet-api/compare/v0.0.11...v0.0.12
[0.0.11]: https://github.com/AlexGalhardo/claude-code-internet-api/compare/v0.0.10...v0.0.11
[0.0.10]: https://github.com/AlexGalhardo/claude-code-internet-api/compare/v0.0.9...v0.0.10
[0.0.9]: https://github.com/AlexGalhardo/claude-code-internet-api/compare/v0.0.8...v0.0.9
[0.0.8]: https://github.com/AlexGalhardo/claude-code-internet-api/compare/v0.0.7...v0.0.8
[0.0.7]: https://github.com/AlexGalhardo/claude-code-internet-api/compare/v0.0.6...v0.0.7
[0.0.6]: https://github.com/AlexGalhardo/claude-code-internet-api/compare/v0.0.5...v0.0.6
[0.0.5]: https://github.com/AlexGalhardo/claude-code-internet-api/compare/v0.0.4...v0.0.5
[0.0.4]: https://github.com/AlexGalhardo/claude-code-internet-api/compare/v0.0.3...v0.0.4
[0.0.3]: https://github.com/AlexGalhardo/claude-code-internet-api/compare/v0.0.2...v0.0.3
[0.0.2]: https://github.com/AlexGalhardo/claude-code-internet-api/compare/v0.0.1...v0.0.2
[0.0.1]: https://github.com/AlexGalhardo/claude-code-internet-api/releases/tag/v0.0.1
