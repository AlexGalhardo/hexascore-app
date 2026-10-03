# Changelog

All notable changes to this project are documented here.
Format: [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/). Versioning: [SemVer](https://semver.org/).

## [0.0.21](https://github.com/AlexGalhardo/claude-code-internet-api/compare/v0.0.20...v0.0.21) (2026-10-03)


### Added

* **ingest:** sandbox seed and faster wikipedia enrichment ([7942f4d](https://github.com/AlexGalhardo/claude-code-internet-api/commit/7942f4d1be5b66f6e78b6c0bed4d7a5d8bc275ec))
* **seed:** sandbox mode on existing items, purge, target guard and banner ([bc9b27f](https://github.com/AlexGalhardo/claude-code-internet-api/commit/bc9b27f7378498c2ea802c87f3abafcf7ad0a073))
* **seed:** sandbox mode on existing items, purge, target guard and banner ([ebf4ae2](https://github.com/AlexGalhardo/claude-code-internet-api/commit/ebf4ae2827434e0db4c595453fc91c0df8ecb1b7))


### Changed

* **ingest:** fetch wikipedia summaries concurrently ([3669be6](https://github.com/AlexGalhardo/claude-code-internet-api/commit/3669be686a26cb9beed62277c8a88548a8e7d9b8))
* **ingest:** fetch wikipedia summaries concurrently ([a704d22](https://github.com/AlexGalhardo/claude-code-internet-api/commit/a704d222301e26663ff50bc118811aee6a26b30b))

## [0.0.20](https://github.com/AlexGalhardo/claude-code-internet-api/compare/v0.0.19...v0.0.20) (2026-10-03)


### Added

* **docs:** translate the API guides to pt-BR and es ([99ab46d](https://github.com/AlexGalhardo/claude-code-internet-api/commit/99ab46d88b174caa8c6b9068989bb34f0bfa24d7))
* **web:** center the users' scores dialog and paginate it by page number ([3783518](https://github.com/AlexGalhardo/claude-code-internet-api/commit/37835183f70a680f3f94b2796aadfb0e19bde4aa))
* **web:** ratings dialog pagination and translated API guides ([731ff7d](https://github.com/AlexGalhardo/claude-code-internet-api/commit/731ff7d065d077abb6f1403f42518937a0c3dff8))

## [Unreleased]

### Changed

- Ingestion: Wikipedia summaries are fetched with 8 concurrent requests per run; `HttpClient`'s per-host throttle is thread-safe (slot reservation under a lock) and Wikipedia hosts allow 20 req/s. The bootstrap workflow gains a `skip_catalog` input to resume enrichment, covers and snapshot without re-running the Wikidata catalog job.

### Added

- Sandbox seed: `lib/db/seed.ts --existing-items` rates the catalog already in the database (sandbox Neon branch copied from production) without touching item rows except stats — keyset batches by id, popularity rank per type via `row_number()`, Zipf curve, per-item quality from the id hash, columnar `unnest` upsert, stats recomputed in SQL per batch — and creates statuses, time to beat (completed games only), follows and lists (≤ `LIST_LIMITS.perType` per type, items of the list's type) for seed users, idempotently. `--purge` deletes every seed user (id `seed-user-*` + `@example.test`, cascade) and recomputes all item stats.
- Seed target guard (`lib/db/seed-guard.ts`): always refuses `VERCEL_ENV=production`, refuses non-local databases unless `SEED_TARGET=sandbox`, and refuses databases with more than 50 verified non-seed users unless `--i-know-this-is-sandbox` is passed.
- `NEXT_PUBLIC_SANDBOX` env (`"true"`/`"false"`, default false): non-dismissable "sandbox — fake data" banner above the navbar (en/pt-BR/es, `role="status"`), `noindex, nofollow` metadata and `robots.txt` disallowing everything.

- API docs guides (`/docs`, authentication, errors, pagination, rate limits, changelog) are translated to pt-BR and es through the new `docs` namespace (`messages/<locale>/docs.json`), including per-page meta descriptions and the code-block copy button. Inline code and links use `t.rich`; code samples, HTTP headers, field/parameter names, error `code` values and URLs stay in English. E2E checks the Portuguese and Spanish quickstart headings (with axe).

### Changed

- Fixture seed recomputes item stats in keyset batches with raw SQL (only rows whose stats changed) instead of one catalog-wide `UPDATE`, and the seed's `hashtext` buckets use `abs(0::bigint + hashtext(…))` so an `int4` minimum can no longer overflow `abs`.
- "See users' scores" dialog is centred in the viewport (above the sticky navbar, full width minus margins on mobile) and paginated: 10 ratings per page, at most 100 pages (latest 1000), via `GET /api/items/[id]/ratings?page=1..100` → `{ data, page, page_count }` (count capped by a `LIMIT 1000` subquery, `OFFSET` ≤ 990); TanStack Query keeps the previous page while loading.
- Docs sidebar note for pt-BR/es now only says that field and parameter names stay in English; the docs `<article>` no longer forces `lang="en"`.

## [0.0.19](https://github.com/AlexGalhardo/claude-code-internet-api/compare/v0.0.18...v0.0.19) (2026-10-03)

### Added

* **covers:** add 96x144 WebP list thumbnails ([44852ab](https://github.com/AlexGalhardo/claude-code-internet-api/commit/44852ab810123755e05ce395ab1f3b3cdff8fc25))
* **covers:** add 96x144 WebP list thumbnails ([1802c6d](https://github.com/AlexGalhardo/claude-code-internet-api/commit/1802c6d02f4230441a8cc39294a0df1d4106b786))
* **web:** cover thumbnails, item header and sticky navbar fixes ([977fbe5](https://github.com/AlexGalhardo/claude-code-internet-api/commit/977fbe510b62832eb950ba7119319399dae7785a))
* **web:** cover, title and score together in the item header ([79a5300](https://github.com/AlexGalhardo/claude-code-internet-api/commit/79a5300bb28ec45d1401d30fda005de0af470880))
* **web:** item header cover and sticky navbar overlap fix ([83083bf](https://github.com/AlexGalhardo/claude-code-internet-api/commit/83083bf732cde5158c1d91decbe64c03d5e2b95e))

### Fixed

* **web:** keep sticky content below an opaque navbar ([64e16e1](https://github.com/AlexGalhardo/claude-code-internet-api/commit/64e16e10205fec3265e5b0d4c6cd3bb59c53f56e))

## [0.0.18](https://github.com/AlexGalhardo/claude-code-internet-api/compare/v0.0.17...v0.0.18) (2026-10-03)

### Added

* **api:** translate the OpenAPI documents to pt-BR and es ([a3aee48](https://github.com/AlexGalhardo/claude-code-internet-api/commit/a3aee48a8678f02c1d7588613a1fb39d7874e036))
* **api:** translate the OpenAPI documents to pt-BR and es ([832f453](https://github.com/AlexGalhardo/claude-code-internet-api/commit/832f4530c575791123cd19023abd24c221a63ed5))
* **docs:** localize the Scalar API reference UI ([d3a95a4](https://github.com/AlexGalhardo/claude-code-internet-api/commit/d3a95a4a796d814b2ef457c5ad9f4ab403212dbd))
* **docs:** localize the Scalar API reference UI ([cf97804](https://github.com/AlexGalhardo/claude-code-internet-api/commit/cf9780408682c2c2133a32ed69ddca867daac723))
* **lists:** cap lists at 10 per user and item type ([38a1d9d](https://github.com/AlexGalhardo/claude-code-internet-api/commit/38a1d9d13ab02c03ddb7fe1b4c2454b9c7ba7f75))
* **lists:** cap lists at 10 per user and item type ([6351916](https://github.com/AlexGalhardo/claude-code-internet-api/commit/63519169898a01f07e10d5ab510cca0c2670692e))
* **web:** button-like link hover, navbar icons and profile nickname ([d7d4f14](https://github.com/AlexGalhardo/claude-code-internet-api/commit/d7d4f14a7b5e320762924633887c7575ead09a60))
* **web:** button-like link hover, navbar icons and profile nickname ([efe1d1a](https://github.com/AlexGalhardo/claude-code-internet-api/commit/efe1d1ad815093ce11bacf38d7f8863d75634aee))
* **web:** catalog covers, link hover, navbar and onboarding polish ([a43bfa1](https://github.com/AlexGalhardo/claude-code-internet-api/commit/a43bfa15ebe981d95cd759250f282bcabc22af62))
* **web:** catalog covers, list limits, localized API docs and UI polish ([a43bfa1](https://github.com/AlexGalhardo/claude-code-internet-api/commit/a43bfa15ebe981d95cd759250f282bcabc22af62))
* **web:** cover thumbnails in catalog lists and fix table overflow ([af19adb](https://github.com/AlexGalhardo/claude-code-internet-api/commit/af19adb55daacb74ec7c4118d60a86bbef6e05ca))
* **web:** cover thumbnails in catalog lists and fix table overflow ([3501301](https://github.com/AlexGalhardo/claude-code-internet-api/commit/350130156bd596ca3a3db1542f39e11611028d2e))
* **web:** onboarding nickname matches sign-up rules ([f7157a1](https://github.com/AlexGalhardo/claude-code-internet-api/commit/f7157a118dfc2dd268a5f8f2c53e3790197c54c3))
* **web:** show only language codes in the locale menu ([b186fd3](https://github.com/AlexGalhardo/claude-code-internet-api/commit/b186fd3cde9921c8726966260aae44ae53396a0d))

### Fixed

* **api:** reject empty query parameters ([3413e2f](https://github.com/AlexGalhardo/claude-code-internet-api/commit/3413e2f8d4c39d2bc63b201e1aa2609e824d32f2))
* **api:** reject empty query parameters ([be7f4f0](https://github.com/AlexGalhardo/claude-code-internet-api/commit/be7f4f01e6755c59966ca84e5249f66a3da40568))

## [0.0.17](https://github.com/AlexGalhardo/claude-code-internet-api/compare/v0.0.16...v0.0.17) (2026-10-03)

### Added

* 1.0 launch prep (domain, navbar, legal, QA gates, covers webp, costs) ([df1046a](https://github.com/AlexGalhardo/claude-code-internet-api/commit/df1046abf58290736f06f7c31123283c21615359))
* **billing:** add sandbox PIX payment button with paid-modal lock ([60947a3](https://github.com/AlexGalhardo/claude-code-internet-api/commit/60947a34e511a29ac0a6ad7ce03ad94c23c49152))
* **ingest:** store covers as optimized webp in vercel blob ([6092491](https://github.com/AlexGalhardo/claude-code-internet-api/commit/6092491ebdb640ee0c55379b9c49c7c3f9a379c4))
* **legal:** affiliate and license notices, i18n review ([2d91ea2](https://github.com/AlexGalhardo/claude-code-internet-api/commit/2d91ea254f3ce09b04976ec17c715dccc79d27f5))
* **seo:** add robots.txt and sitemaps with hreflang ([7c3f78d](https://github.com/AlexGalhardo/claude-code-internet-api/commit/7c3f78d71596126db80f05fde0de788d2a208a1e))
* **seo:** add search console cli ([d46eadf](https://github.com/AlexGalhardo/claude-code-internet-api/commit/d46eadff08aebdcf5fcd71be7428d02283bf3c05))
* **web:** crowdsourced time to beat for games ([bc77d10](https://github.com/AlexGalhardo/claude-code-internet-api/commit/bc77d10c1044f938c7e688ce6831243fcb1d2beb))
* **web:** new navbar, search previews, nickname rules and api host ([f011932](https://github.com/AlexGalhardo/claude-code-internet-api/commit/f011932e936cf42b7755097dd38ebf81ad7cbb9d))
* **web:** use hexascore.app as the default domain ([a6c654b](https://github.com/AlexGalhardo/claude-code-internet-api/commit/a6c654b9250611e15ace1db41b71799ba325ac7c))
* **web:** use TanStack Query for client-side fetches ([9596fef](https://github.com/AlexGalhardo/claude-code-internet-api/commit/9596fef098b08cf2170979004d87fb1e798e6f17))
* **web:** use TanStack Query for client-side fetches ([a8d5c4b](https://github.com/AlexGalhardo/claude-code-internet-api/commit/a8d5c4b2c917e033c9fcff0f91a9874c182d1b9f))

### Fixed

* **account:** revoke api keys and cancel stripe subscription on purge ([6783b28](https://github.com/AlexGalhardo/claude-code-internet-api/commit/6783b280ed12a69e002ce0dace26409cb7031671))
* **auth:** allow sign-up emails up to 48 characters ([83978bc](https://github.com/AlexGalhardo/claude-code-internet-api/commit/83978bc9be30567dbeb353982232cf85de583b3d))
* **ingest:** keep the database connection alive during long fetches ([668ed2e](https://github.com/AlexGalhardo/claude-code-internet-api/commit/668ed2eab5f93181b66977c2ad5bd14dd965bb9b))
* **ingest:** keep the database connection alive during long fetches ([7fd737d](https://github.com/AlexGalhardo/claude-code-internet-api/commit/7fd737d3d801db9c9919b440080b0b0902ee829a))
* **security:** allow the vercel toolbar frame on previews ([0a7f84d](https://github.com/AlexGalhardo/claude-code-internet-api/commit/0a7f84d772f33de07f7fec0a90f37eeb8b0fddb9))
* **security:** allow the vercel toolbar frame on previews ([444891a](https://github.com/AlexGalhardo/claude-code-internet-api/commit/444891aa2050bffee08f80d7cbeb73c735aebc0b))

## Detailed notes for 0.0.17–0.0.19

### Added

- OpenAPI documents in Portuguese and Spanish: `public/openapi.pt-BR.json` and `public/openapi.es.json` next to `openapi.json` (en, unchanged byte for byte, still the SDK source). Every description/summary lives in `messages/<locale>/openapi.json` (typed `openApiText`/`apiDoc` in `lib/api/openapi-text.ts`); `bun run openapi` writes the three files and `openapi:check` verifies all of them. `lib/api/openapi.test.ts` asserts the localized documents are structurally identical to en and fully translated; the messages parity test covers the new file.
- `/docs/reference` loads the spec of the site language (and sets `lang` on it); the api. host root lists `openapi: { en, "pt-BR", es }` (was a single URL); `/docs/changelog` mentions the translated documents.

### Changed

- API reference (Scalar) UI follows the site language: `localization.locale` = `pt` for pt-BR and `es` for Spanish (built-in packs). Spec title is now "HexaScore API".

### Fixed

- Public API rejects empty query values (`?min_score=`, `?limit=`) with a 400 problem instead of coercing them to 0 (found by Schemathesis on the preview).

### Changed

- Every link looks like a small button: padded and rounded, inverted colours on hover (black/white in light, white/black in dark), no underline effect (`globals.css`; `[data-plain]` opts out the wordmark and Scalar).
- Navbar: "API Docs" label; the account button shows a user icon instead of the nickname initial; the account menu items carry icons.
- `/profile` shows `@nickname` at the right of the title; the sign-up hint says the nickname is unique and cannot be changed later. Google onboarding drops the avatar circle and shares the sign-up nickname rules and hint.

### Fixed

- Catalog tables use a fixed layout: long titles and genre lists truncate instead of widening the page past the navbar container (no horizontal scroll).

### Changed

- Curated lists have a category (`lists.type`, the `item_type` enum; migration `0005_list_type` backfills existing lists from their first item, `game` when empty). Each user can have at most 10 lists per category (`LIST_LIMITS.perType`), replacing the global cap of 50; a list only accepts items of its own category (`typeMismatch` error). The item page offers only lists of the item's type and shows the quota (e.g. "Game lists: 3/10"); the profile's new-list form requires a category; profile and list pages show the category badge.
- Catalog and search lists show a 2:3 cover thumbnail left of each title (`components/catalog/cover-thumb.tsx`, shared with the search autocomplete).
- Language menu shows codes only (EN / PT-BR / ES) in the trigger and the list; native names stay as accessible labels.

### Fixed

- Preview deployments allow the Vercel Toolbar iframe (`https://vercel.live`) in `frame-src`/`connect-src`; production CSP is unchanged. Ingestion keeps the Neon connection alive during long fetches.

### Added

- PIX sandbox payment for testing: with `ABACATEPAY_SANDBOX=true` (default off) the PIX QR modal shows a "PAY SANDBOX PIX" button (10 s cooldown) that calls AbacatePay `POST /v2/transparents/simulate-payment` through `simulatePixPaymentAction` (owner's pending charge only); the plan is still applied by the webhook/status poll.
- SEO: `robots.txt` (indexable only in production, private pages disallowed in every locale) and sitemaps with hreflang alternates: `/sitemap-index.xml` → `/sitemap/0.xml` (static pages) + `/sitemap/<n>.xml` (catalog items, 40k per file).
- `bun run seo:gsc` (apps/web): Search Console CLI over a service account (sites, sitemaps, submit, Search Analytics queries).
- Item page credits the cover (source, author, license and link) when the ingestion stored `meta.coverAttribution` (Wikimedia Commons covers under CC BY/BY-SA require it).
- Tests: legal documents must match paragraph-for-paragraph across locales, carry the brand from `APP_NAME` and no locale file may contain the codename; `localeFromRequest` (`i18n/locale.ts`).
- E2E specs for the 0.3 flows (`qa/e2e/specs/ratings.spec.ts`, `social.spec.ts`, `admin.spec.ts`), replacing the `todo-0.3` placeholders: 6-category rating, edit and the 5-rating score threshold; item status and curated lists; public vs private profile; follow → `/feed`; report → admin resolves; `/admin` 404 for non-admins. DB helpers `createItem`, `createMember`, `insertRating`, `grantAdmin`.
- API QA gates in `.github/workflows/qa.yml` (run in sequence against the Vercel preview, pinned Docker images, no npm dependency): k6 2.3.0 load test (`qa/k6`, p95 < 300 ms, 429 with `Retry-After` on a burst, Free key 402 after the quota and 400 on `limit=11`), Schemathesis 4.29.0 fuzzing of `openapi.json` (`qa/schemathesis`, 5xx + response conformance) and an OWASP ZAP 2.17.0 baseline + API active scan (`qa/zap`, fails on High). Secrets `QA_API_KEY` (paid plan) and optional `QA_FREE_API_KEY`; jobs skip with a notice without them.
- Docs: `docs/vercel-costs-monthly.md`, a monthly production cost estimate (Vercel Pro + Flat Rate CDN, Blob for 1M covers, Neon, Upstash, Resend, domain, GitHub Actions, Stripe/AbacatePay fees) with three traffic scenarios in USD/BRL, cost controls and how to upgrade Neon/Upstash through the Vercel Marketplace; prices checked 2026-10-03.
- Crowdsourced time to beat for games (replaces scraping HowLongToBeat): members whose status is "Played" can report hours (0.5–10000 h, 0.5 steps) for main story, main + extras and completionist; editable and deletable, removed when the status leaves "Played". The item page shows the median per style and the report count ("not enough reports" under 3). New table `time_to_beat` (migration `0004_time_to_beat`), actions `saveTimeToBeat`/`deleteTimeToBeat`, item merge moves the reports, dev seed adds reports for about a third of each game's raters, PGlite tests and an `@auth` E2E spec.
- Public API: `ItemDetail.time_to_beat` (`main`/`extras`/`completionist`: `{ median_hours, reports }`, `null` for non-games) — additive; `openapi.json` and both SDKs regenerated.

### Changed

- Client-side fetches use TanStack Query (`@tanstack/react-query` 5.104.1, new dependency) through a `QueryProvider` in the `[locale]` layout: search autocomplete (`useQuery`), the user scores modal (`useInfiniteQuery`) and PIX status polling (`useQuery` + `refetchInterval`) replace hand-rolled `useEffect` + `fetch`; zod validation and timeouts kept. Server Components remain the default data path.
- PIX checkout shows the QR code in a modal; once paid it cannot be closed for 10 seconds (visible, announced countdown).
- Navbar: sticky, "HexaScore" wordmark (green "Hexa", blue "Score"), search fills the row up to the actions, only a "Log in" button (no "Sign up"), language menu last and always showing the current language, icon links (Games, Movies, Series, Books, API, Pricing) spread with `justify-between`; on mobile only the brand, a search button and a hamburger menu.
- Search autocomplete shows the cover, type, year and final score of the 8 best matches.
- Nickname is 4–16 lowercase letters only (normalized while typing, enforced by the Better Auth username validator); sign-up email is capped at 48 characters and password inputs at 32. Seed and E2E nicknames follow the rule.
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
