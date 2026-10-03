# Changelog

All notable changes to this project are documented here.
Format: [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/). Versioning: [SemVer](https://semver.org/).

## [Unreleased]

### Added

- Item page credits the cover (source, author, license and link) when the ingestion stored `meta.coverAttribution` (Wikimedia Commons covers under CC BY/BY-SA require it).
- Tests: legal documents must match paragraph-for-paragraph across locales, carry the brand from `APP_NAME` and no locale file may contain the codename; `localeFromRequest` (`i18n/locale.ts`).
- E2E specs for the 0.3 flows (`qa/e2e/specs/ratings.spec.ts`, `social.spec.ts`, `admin.spec.ts`), replacing the `todo-0.3` placeholders: 6-category rating, edit and the 5-rating score threshold; item status and curated lists; public vs private profile; follow → `/feed`; report → admin resolves; `/admin` 404 for non-admins. DB helpers `createItem`, `createMember`, `insertRating`, `grantAdmin`.
- API QA gates in `.github/workflows/qa.yml` (run in sequence against the Vercel preview, pinned Docker images, no npm dependency): k6 2.3.0 load test (`qa/k6`, p95 < 300 ms, 429 with `Retry-After` on a burst, Free key 402 after the quota and 400 on `limit=11`), Schemathesis 4.29.0 fuzzing of `openapi.json` (`qa/schemathesis`, 5xx + response conformance) and an OWASP ZAP 2.17.0 baseline + API active scan (`qa/zap`, fails on High). Secrets `QA_API_KEY` (paid plan) and optional `QA_FREE_API_KEY`; jobs skip with a notice without them.

### Changed

- Navbar: sticky, "HexaScore" wordmark (green "Hexa", blue "Score"), search fills the row up to the actions, only a "Log in" button (no "Sign up"), language menu last and always showing the current language, icon links (Games, Movies, Series, Books, API, Pricing) spread with `justify-between`; on mobile only the brand, a search button and a hamburger menu.
- Search autocomplete shows the cover, type, year and final score of the 8 best matches.
- Nickname is 4–16 lowercase letters only (normalized while typing, enforced by the Better Auth username validator); sign-up email is capped at 36 characters and password inputs at 32. Seed and E2E nicknames follow the rule.
- Home page has no visible hero text (an `sr-only` h1 remains); footer keeps only "© 2026 HexaScore", Terms, Privacy and Contact. Default `APP_NAME` is "HexaScore".
- `api.hexascore.app` serves only the API: `/` returns a JSON index (base URL, docs, OpenAPI) and any path outside `/v1` is an RFC 9457 404.
- Terms of use (en, pt-BR, es): data sources and licenses (Wikidata CC0, Wikipedia CC BY-SA 4.0, Open Library, Steam, GOG, covers incl. Wikimedia Commons), API quotas/page size/rate limits and 5xx not counted, Stripe monthly vs PIX 1/12 months with reminders, credits only for Pro/Business, refund of duplicate charges, store responsibility for affiliate purchases, contact for withdrawal/refund/privacy requests.
- Terms and privacy intros name the service through `{appName}`, replaced by `APP_NAME` in `lib/legal.ts`.
- Verification and password-reset emails are translated (`auth.emails` in the three locales), picking the locale from the `locale` cookie or the calling page's prefix.
- Default domain is now `hexascore.app` (API at `api.hexascore.app/v1`): OpenAPI servers/links, docs examples, CSP `connect-src`, SDK base URLs and READMEs. Production `NEXT_PUBLIC_APP_URL` and `BETTER_AUTH_URL` point at `https://hexascore.app`; `www.hexascore.app` redirects to the apex.
- Ingestion covers (`webscrappings/covers.py`): with `BLOB_READ_WRITE_TOKEN`, covers are resized to fit 600×600, re-encoded as metadata-free WebP (quality 78, method 6; ~19 KB average, about 60% smaller than the sources) and stored in Vercel Blob at a content-addressed `covers/<item_id>-<hash>.webp` with a one-year cache; unchanged sources skip encode and upload. Without the token the source URL is kept, as before.
- New Python dependency: `pillow==12.3.0`.

### Fixed

- Hardcoded English left in localized pages: the navbar sections `aria-label` and the screen-reader labels of the pricing plan facts.
- Account anonymization (`lib/account-purge.ts`) now revokes the user's API keys and cancels a live Stripe subscription before wiping the account; if the cancellation fails or Stripe is not configured, the account stays scheduled and the daily cron retries.
- E2E signup spec waits out Better Auth's per-IP sign-up limit (429) instead of failing when parallel workers create accounts.

### Security

- Forms that carry a password or 2FA code (`/login`, `/signup`, password, 2FA and reset forms) use `method="post"`, so a submit before hydration never puts the secret in the URL.

## [0.0.16] - 2026-10-03

### Added

- Strict Content Security Policy with a per-request nonce (`proxy.ts`, `lib/security-headers.ts`): scripts only by nonce + `strict-dynamic`, `frame-ancestors 'none'`, frames only from `youtube-nocookie.com`, Scalar allowances scoped to `/docs/reference` and `/profile/api`; HSTS, `nosniff`, `Referrer-Policy`, `Permissions-Policy` and COOP on every response; `/api/*` locked to `default-src 'none'`.
- Vercel Analytics 2.0.1 and Speed Insights 2.0.0 (cookieless) in the locale layout.
- Feature flags (`flags` 4.3.1): `pix-checkout` and `public-api-signup` kill switches, server-side, default on (`FLAG_*=off` or Vercel Toolbar overrides).
- Playwright + axe E2E suite in `qa/e2e` (desktop/mobile × light/dark, en/pt-br/es): public pages, headers/CSP, catalog filters and pagination, item page (trailer consent, users' scores modal), search + autocomplete keyboard, 404s, signup → verification → login, TOTP 2FA, change password, account deletion countdown, contact, pricing/checkout states, docs + Scalar, API key lifecycle. Workflow `.github/workflows/qa.yml` runs it against the Vercel preview on PRs `dev` → `main`.
- Skill `hexascore-qa`; app icon (`app/icon.svg`).

### Changed

- Seed: ratings are written with columnar `unnest` upserts (10k users / 100k items / 2.2M ratings in ~7 min on Docker Postgres), raters follow a heavy-tailed activity curve and have a per-user strictness bias; phase timings are logged.
- Every `[locale]` page is now dynamically rendered (required by the nonce).

### Fixed

- `/favicon.ico` (and any unknown first segment) reached catalog queries with locale `favicon.ico` and logged a ZodError on every page view; unknown locales now 404 before any query (`asLocale`).
- `?genre=%00` on catalog pages and `q=%00` on `/search`, the autocomplete and `/v1/search` answered 500.
- Accessibility (axe): light-theme section accents, the filter button, the low score band, error/destructive colors and satisfied password rules were below 4.5:1; docs code blocks were not keyboard-scrollable; the API key revoke dialog overflowed at 360 px.

## [0.0.15] - 2026-10-03

### Added

- Ratings: save/delete with verified email, nickname and not banned; one transaction locks the item row and recomputes `rating_count`/`score`/`score_c1..c6`; rating panel with six keyboard-accessible score inputs and live overall.
- Item status (want / in progress / completed / dropped) and "Add to list" on the item page.
- Public profiles `/u/[nickname]` (private profiles render a minimal page), follow/unfollow, `/feed`, curated lists and reports of lists/users.
- `/admin` (role `admin`, 404 otherwise): contact messages, reports moderation (hide list, ban), affiliate links, item merge and ingestion status; `bun run admin:grant <email>`.
- `/onboarding` nickname step for Google sign-ups; migration `0003_social_moderation`.

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

[Unreleased]: https://github.com/AlexGalhardo/claude-code-internet-api/compare/v0.0.16...HEAD
[0.0.16]: https://github.com/AlexGalhardo/claude-code-internet-api/compare/v0.0.15...v0.0.16
[0.0.15]: https://github.com/AlexGalhardo/claude-code-internet-api/compare/v0.0.14...v0.0.15
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
