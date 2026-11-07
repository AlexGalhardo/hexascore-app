# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **mobile:** `bun run android:dev` builds and runs the debug app on an Android emulator against a local dev backend (`EXPO_PUBLIC_API_BASE_URL` defaults to `http://10.0.2.2:3000`; debug builds already allow cleartext, release stays https-only). On Windows the native build runs in a short-path hoisted mirror (`C:\h`) because ninja cannot stat the 260+ character paths of bun's isolated `node_modules`. Runbook: `docs/runbooks/mobile-local-android.md`.

- **perf:** Two partial indexes on `items` for the default listing (migrations `20261107000000` and `20261107000100`): page one of a type sorted by score reads the index instead of sorting every item (42.8 ms to 0.12 ms on 22,500 games) and the total count is an index-only scan (39.2 ms to 1.3 ms).
- **docs:** `docs/runbooks/postgres-cost.md` (owner-run steps to measure, right-size the pool, fold staging into the production instance and size the volume) and ADR 0027 (stay on Postgres, SQLite deferred).
- **ui:** The footer newsletter form is live: it posts the email (with a honeypot) from the browser to vukem-email's `/subscribe/hexascore` (`NEWSLETTER_API_URL`, default `https://email.hexascore.app` in production and `http://localhost:3100` otherwise; `NEWSLETTER_PROJECT_SLUG`), whose origin joins `connect-src`. States: sending, check your inbox, invalid, rate limited, error.
- **auth:** Login and sign-up share one minimal centered layout (logo, title, Google and passkey buttons, divider, form, link to the other page); the two-factor step is its own `/2fa` page with six code boxes and a backup-code swap, and `/two-factor-challenge` redirects to it.
- **perf:** The landing page's highlights and upcoming blocks are cached per locale for 5 minutes and forgotten when a trending snapshot lands, so a visit no longer runs 12 catalogue queries.

- **ingest:** `wikidata-discover` chain (every 30 minutes, behind `FLAG_INGEST_WIKIDATA`): walks Wikidata below the top-N selection that capped the catalogue at 100,000 items, most popular first (sitelink count, floor 3) and one type at a time (films, series, games, books), stores the records of items not fetched yet and lets the normaliser create them. The position persists across runs. It needs no TMDB, Open Library or paid access, and Goodreads, Skoob, JustWatch and the IMDb datasets stay unused, as decided in `docs/data/SOURCES.md`
- **admin:** A progress bar sweeps across the top and the content dims while an admin page loads (after a 120 ms delay, so fast pages do not flash).

- **admin:** Every admin screen is now a React page in the new shell: ingestion and its runs, source records, the match, field and conflict review queues, the catalogue audit with revert, locks, item management (fields, affiliate links, offers, trending, merge), the moderation queue, detail, metrics and bulk result, and read-only list and detail screens for users, score bursts, rating flags, score actions, formula versions, reports, contact messages, items, job runs, ingest failures, data quality results, the audit log and the feature flags (`/admin/r/<slug>`, an allow-list of columns). Actions post to the same endpoints and show the server's message inline. Feature flags are changed from the flags screen and contact messages are marked handled from their detail page, both recorded in the audit log.

- **admin:** A new React admin dashboard at `/admin/overview` (Inertia, TanStack Query, shadcn-style components over Base UI) with a sidebar, light and dark themes, headline counts, worker heartbeat and recent ingestion runs. It sits behind the panel guard (2FA, CSRF, ability) and refreshes every 30 seconds. The framework's own panel is no longer mounted: `/admin`, `/admin/_flags`, `/admin/_audit` and `/admin/<resource>` redirect to the React screens (ADR 0026, VF-083).

- **catalog:** Upcoming releases within 90 days on the landing, 30-day activity ranking for released items, and a where-to-watch provider filter for movie and series listings
- **ui:** A grouped footer with a disabled newsletter preview and a catalogue demo seed with future releases for each category
- **catalog:** Where to watch or buy on every item page: streaming, rent, buy and free-with-ads providers from TMDB (display only, with the JustWatch credit), Steam, GOG, Epic, Google Play, App Store and Amazon links derived from the catalogue ids, a country switch that follows the locale, and `/go/{id}` as the only exit (rate limited, uncached, no user id or IP counted); affiliate tags are added at the redirect when `AFFILIATE_*_TAG` is set
- **admin:** The item manage page lists, deactivates, restores and adds offers, and pins or bans an item on the trending lists; `/admin/ingest/conflicts` holds items that duplicate another until an admin allows or dismisses them
- **catalog:** The normaliser no longer creates an item whose external id another item owns or whose English title and year match another item; unique indexes on the external ids back the rule
- **ingest:** `refresh.existing` verifies covers and stores a quality score; `llm.normalise` (off by default, monthly budget `LLM_MONTHLY_BUDGET_USD`) queues model-extracted facts for review; the worker heartbeat and `WORKER_HEARTBEAT_CHECK` prepare the separate worker service (`docs/runbooks/worker.md`)
- **release:** `changelog-sync.yml` checks that CHANGELOG.md and the GitHub releases match and that every released section is the one-commit-per-line git-cliff list (`scripts/release/check-sync.ts`)
- **config:** `TELEGRAM_OPS_CHAT_ID` declared in `.env.example` for the catalogue worker digest (read by the app once RFC-001 CW-08 lands)
- **docs:** Architect audit 2026-11-02: project map, web test report, technical review and cleanup ledger in `docs/audit/`; RFC-001 catalogue worker, RFC-002 trending, RFC-003 offers; ADR 0024 (Redis deferred) and ADR 0025 (MongoDB rejected); `docs/data/SOURCES.md`, `docs/ROADMAP-V1.md`, `docs/BACKLOG.md`, `docs/OPEN-QUESTIONS.md`; ROADMAP phase P8

### Fixed

- **auth:** A Google sign-in on an account with two-factor authentication goes to the `/2fa` code step (it went back to `/login`, which no longer opens a code dialog); both sign-in paths keep the locale prefix.
- **auth:** The web two-factor challenge (`POST /2fa`) now also stops guessing per account, 10 tries per 15 minutes whatever the address, so rotating IPs no longer widens the budget (security audit 2026-11-07).
- **uploads:** Community cover uploads and the covers fingerprint step refuse images over 25 megapixels before decoding them.
- **mobile:** The Google sign-in nonce comes from the platform CSPRNG; the sign-in is skipped when none exists instead of falling back to `Math.random`.
- **mobile:** Plural messages no longer crash the home screen on Android: Hermes has no `Intl.PluralRules`, so `translate` falls back to the en, es and pt-BR rules.

- **trending:** The landing carousel requires a cover: an item with a synopsis, both translated titles and a date reached the quality floor without one, took a slot and was then hidden, leaving a gap; pins need a cover too. "For you", similar items and cross-media links leave items without a cover out as well
- **ingest:** The Telegram digest counts right: `offers.tmdb` counts the items that got offers instead of every offer (it showed 447 thousand changed for 36 thousand seen), `refresh.existing` counts the items whose quality score moved instead of every checked item, and the pace and ETA use the last hour instead of a 15-minute window
- **ingest:** A book edition whose source gives an implausible publication year (0 or 9999) no longer fails the item's editions step: years outside 1400 to 2100 are stored as unknown
- **item:** An item whose trailers came from Wikidata (stored as a list of YouTube ids) no longer answers 500: the item page reads both that list and the older objects. Grand Theft Auto VI was the first to fail
- **catalog:** The canonical genres are permanent: migration `20261104000000_canonicalize_genres` repairs every item whose genres still show Wikidata labels or lost their canonical link (the same repair as `genres:canonicalize`, as SQL), and the normaliser now canonicalises each item it saves, so a new ingest no longer brings the source labels back and the command no longer needs to be run by hand
- **ingest:** An item whose alternative titles differ only in case or accents no longer fails its whole relations step (the insert now ignores any unique conflict, not only the exact title)
- **ingest:** The worker heartbeat counts as stale after 8 minutes instead of 3: it queues behind two running ingest chunks of up to 5 minutes, so a busy worker was reported stale
- **ratings:** Unreleased item pages explain when ratings open and disable guest and member rating controls; the existing server gate remains authoritative
- **ui:** The navbar bottom border uses the same light border token as the catalogue cards
- **ingest:** A cancelled refresh stops before opening a batch and between cover checks
- **trending:** Flagged seed accounts no longer count in the trending signals or in the collaborative-filtering gate (OQ-07)
- **a11y:** The Details links on the item page are 24 px touch targets again (Lighthouse target-size)
- **web:** The render cache keys pages by their locale prefix again: `/pt-br/...` and `/es/...` no longer serve the English (or Spanish) copy rendered first

### Changed

- **auth:** `/login`, `/signup`, `/2fa`, password recovery and reset render as standalone pages without the site navbar and footer, with a "‹ Home" link at the top left.
- **newsletter:** The footer form accepts addresses of 12 to 48 characters with a real domain, and says so before posting.
- Require vukem v0.39.1, which sends Resend addresses as `email` or `Name <email>` (production mail failed with "Invalid from" on v0.39.0).
- **auth:** Google sign-in in the Android app uses the authorization code flow with PKCE (S256) and `state` instead of the implicit ID-token flow. The app first calls the new `POST /app/v1/auth/google/nonce` (guest, rate limited); the ID token must carry that single-use, 5-minute nonce, which `POST /app/v1/tokens/google` checks after verifying the token and consumes when it mints the app token (migration `google_nonces`, hashed). `expo-crypto` 58.0.7 is a direct dependency of the app.

- **docs:** Agent guidance covers the shared framework feedback file (`HS-` ids), the `RUN_SLOW_CI` gate and `scripts/ci/local.sh`, the removed staging environment, integration and Windows test pitfalls, SSR private prefixes, JS-only forms, and the mobile local-dev notes.
- **ui:** The public site adopts shadcn/ui components and a neutral zinc theme (ADR 0029, ported from PR #69): flat buttons and inputs, hairline borders, rounded-md 8 px, blurred navbar with a shadow once scrolled, dark mode on the same greys. The palette and radii match the Android kit (`neutral.json`), checked by `packages/tokens/neutral.test.ts`. New dependency: `@radix-ui/react-slot` 1.5.0.
- **mobile:** The sign-in screen is rebuilt on the new kit (wordmark, labelled fields with Show/Hide, a separate two-factor page), screens use Card, ListItem, EmptyState, Skeleton and Avatar, toasts are in-app (`ToastHost`, announced to screen readers) instead of `ToastAndroid`, and detail screens keep clear of the status bar. Category accents stay as a 2 px rule on trending cards and the heading squares.

- **mobile:** The Android app wears the neutral zinc skin of the new web UI: grey palette in light and dark, 4/8/12 px radii, hairline borders instead of ledges, Inter as the only family (Playfair Display is no longer embedded). New shadcn-style kit in `apps/mobile/src/ui` (Button with default/secondary/outline/ghost/destructive and three sizes, Input, Label, TextField, Card, Badge, Separator, Skeleton, Avatar, SegmentedControl, Sheet, Toast, ListItem, EmptyState); `Btn` is replaced by `Button` and the language and theme pickers are segmented controls.

- **ingest:** The language-model gap filler (only for items still missing data after every source, budget US$10 a month spread at most 1/30 a day): a value written in the stored Wikipedia introduction is applied to the item at once, anything else goes to review; a new stage translates the English synopsis into pt-BR and es when they are missing (the most popular items first, never replacing a synopsis); and with `LLM_WEB_SEARCH=true` the most popular items with no stored text are looked up with web search, always into review. The client no longer sends `temperature`, which Claude Haiku 5.5 rejects, so no call could succeed before
- **ingest:** Discovery runs as four lanes, one per category (films, series, games, books): each lane walks and normalises only its type, so they run side by side, and with `INGEST_LANES=true` each goes to its own queue (`ingest-movie`, `ingest-series`, `ingest-game`, `ingest-book`); see the worker runbook for the order to switch them on
- **catalog:** The game mode filter of /games shows a few canonical modes named with a capital first letter in en, pt-BR and es (Single player, Multiplayer, Co-op, PvP, Online multiplayer, Split screen, MMO) instead of the Wikidata labels ("single-player video game", "multiplayer online game"): migration `20261104100000_canonicalize_game_modes` folds the existing links, and the normaliser does the same for every new item
- **item:** Where to watch shows plain text links instead of boxes, and each streaming service links to its own site (Netflix, Prime Video, Apple TV, Google Play, YouTube and the other known services) instead of TMDB's page about the title; a service with no known site is left out of the page and of the redirect
- **ingest:** The 30-minute "fill" schedules of the Wikidata, Open Library and TMDB full chains are removed: they re-read the same top 100,000 records and held the ingest slots; the weekly runs and the `wikidata-discover` chain remain
- **catalog:** The landing page and the four category pages hide items without a cover (`FLAG_HIDE_UNCOVERED`, on by default); an item without a score is still listed, and its score badge appears once it has one. The item page, search and the public API are unchanged
- **ingest:** The Telegram ingestion digest goes out every 15 minutes and shows new items by type with the pace and ETA to 1,000,000, the running runs with age, counts and a warning for a stalled one, records fetched per source, an example error, and a warning when no item was created in 45 minutes
- **home:** Release date badges on the landing carousel are larger, with bigger text, for easier reading.

- **catalog:** Grid and category cover images fill their card, and the score badge is slightly larger.

- **home:** Upcoming carousel cards show the release date as a bright orange badge with black text in the top-right corner.

- **catalog:** Item cards in the grids and category pages no longer show the title on the cover; the score stays in the top-right corner, and the title is kept for screen readers.

- **home:** Landing carousel covers fill their box (cropped to cover), and each card's footer shows only the release date; the title stays available to screen readers.

- **home:** Landing carousels show four cards per slide in a four-column grid, with the twelve upcoming cards spread over three slides.

- **item:** The where-to-watch and where-to-play block is hidden when an item has no offers, instead of showing an empty message.

- **item:** Where to watch (films and series) and where to play (games) show offers from every country at once, with no country selector, so a service listed in one country appears for every visitor.

- **ui:** Cover images are shown whole at every size (no cropping); the space left around them takes the card colour.

- **item:** Where to watch moved into the Details card, below the action buttons, and its labels are now served to the page (the `offers.` namespace was missing from the page translations).

- **home:** Landing carousels show 6 upcoming poster cards per row (12 in total); slides advance automatically every two seconds, pausing on hover or focus and when reduced motion is on; no arrow buttons; navigation is by arrow keys and swipe. Each carousel heading uses the trending label style, and each card shows its release date in the top-right corner.

- **item:** On unreleased items, logged-in users can only use "want" and "add to list"; the other status buttons are disabled and the time-to-beat block is hidden. Clearing an existing status stays allowed.

- **trending:** The landing carousels show upcoming announcements with their known release-date precision instead of scores and advance only when a visitor navigates; the released 30-day ranking fills the grids below them. The 7-day score v2 snapshots remain available to the API
- **ui:** Narrower site container, flatter and squarer catalogue cards, permanent navbar shadow, item verdict labels, shorter runtime display, and Amazon links below eligible book and game covers
- **ui:** Cards (landing rows, category grids, trending) have a soft grey edge instead of a category colour, and buttons, inputs and cards share one radius scale (`rounded-md` controls, `rounded-lg` cards, `rounded-sm` covers and badges); native controls follow the theme, text selection uses the brand colour and headings balance their line breaks
- **catalog:** Canonical genres: one translated list per category (en, pt-BR, es) replaces the Wikidata genre labels; `genres:canonicalize` relinks the items and keeps the source links hidden in `item_genre_sources`; the rules cover the most used labels of the production dry run (85% of links before, more now: LGBTQ+, educational, telenovela, tokusatsu, science fiction and fantasy games); listing filters show the translated names and the public `/{type}/genres` endpoints take `locale` and return `name`
- **api:** The API reference groups operations under games, movies, series, books and catalog, and every operation has a one-sentence description; the Python SDK modules follow the new tags
- **ui:** Category listings page 40, 80 or 120 items (default 40); the landing's "More" grid shows two cards per row at phone width
- **i18n:** Shorter /about intro in en, pt-BR and es
- **ui:** The item page drops the genre and platform buttons under the title (Details still links them); at phone width the cover fills the width and the rating column follows Details, before the title and synopsis
- **build:** `.dockerignore` keeps worktrees, tool state, venvs, nested env files and `node_modules` out of the Docker build context; `local.properties` is gitignored
- **agents:** `PROMPT-WORKERS.md` rewritten: orchestrator reviews and merges, senior and intern in standby, ticket and report formats, autonomy limits

### Removed

- **infra:** The `staging` environment (app, worker, Postgres and bucket; deleted in Railway by the owner on 2026-11-07, ADR 0028) and everything that pointed at it: the staging runbook, the staging-only scheduling of `backup:drill` (the command stays, run by hand), the staging-merge step of `postgres-cost.md`, and the staging cost rows of SDD 11. `SEO_NOINDEX` and the non-production `robots.txt` stay for previews and local production builds.

## [0.38.0] - 2026-11-02

### Added

- **ui:** Inter-only font, wider shell and category-bordered item cards ([639f533](https://github.com/AlexGalhardo/hexascore-app/commit/639f5332ef90f2aa15bd8214e545f1cf6e96463f))
- **ui:** Share contents sidebar layout across legal and about pages ([d0cb9e1](https://github.com/AlexGalhardo/hexascore-app/commit/d0cb9e18e884633131876868393f36e493626c0b))
- **catalog:** Page section listings by number and send genre pages to the filtered listing ([226feb2](https://github.com/AlexGalhardo/hexascore-app/commit/226feb2709dcb61b5fe5aff0c04c9787231c58ab))
- **catalog:** Rework item page columns and open time to beat to all signed-in users ([e5894c9](https://github.com/AlexGalhardo/hexascore-app/commit/e5894c99f02e95c936c746cfcfbb5fc202037005))
- **ui:** Square hover backgrounds, room under the navbar and a versioned favicon ([d155281](https://github.com/AlexGalhardo/hexascore-app/commit/d155281ca0efff96c8be7013647064373dc41b75))
- **ui:** Drop italics, fix the brand casing and refine menus, trending cards and document pages ([5bdcc93](https://github.com/AlexGalhardo/hexascore-app/commit/5bdcc9336a59ded3b17fde4ef80adf09a785d892))
- **ui:** Line up trending scores on the rank row ([3109abd](https://github.com/AlexGalhardo/hexascore-app/commit/3109abde1258ae772893eaab0012957bd4cf3b19))
- **ui:** Drop the score band and rating count from trending cards ([467220e](https://github.com/AlexGalhardo/hexascore-app/commit/467220ed62812c5a8c2c5711beeebdfc784c63b3))
- **ui:** Drop the section rules on the item page ([d8f51ef](https://github.com/AlexGalhardo/hexascore-app/commit/d8f51ef76fa85b068298654fd6e8e9e08ff60069))
- **ui:** Invert colours on navbar, menu and footer hover ([b10f2a7](https://github.com/AlexGalhardo/hexascore-app/commit/b10f2a721c12aff69fee64c2e2643847752388cc))
- **ui:** Copy buttons on docs code, three-column profile and bold landing sections ([dd34ee9](https://github.com/AlexGalhardo/hexascore-app/commit/dd34ee9ab38299fc8a55b09ba9e70b51f37bd747))
- **billing:** Hobby and pro plans with direct checkout from pricing ([12ef855](https://github.com/AlexGalhardo/hexascore-app/commit/12ef85569b586574e9204499a68a3681a52bffa5))
- **docs:** Translate the api documentation into pt-br and es ([c64c1bf](https://github.com/AlexGalhardo/hexascore-app/commit/c64c1bf5737158d93921e73341f4b5203fb09c8c))
- **catalog:** Add speedrun as a fourth time to beat style ([917ba55](https://github.com/AlexGalhardo/hexascore-app/commit/917ba55b2a9e268350318ddb3fe56f015a0ef713))
- **ui:** Add depth to item cards and paint the books border purple ([edfb755](https://github.com/AlexGalhardo/hexascore-app/commit/edfb755e15b0c082c2dd2e7611d2855f8f5c4139))
- **ui:** Raise carousel controls and the ratings dialog close button ([c0efb12](https://github.com/AlexGalhardo/hexascore-app/commit/c0efb1253b74b4ea227740edf5589816a4b49a3e))
- **ui:** Smaller section titles on the category listings ([0b5f5cc](https://github.com/AlexGalhardo/hexascore-app/commit/0b5f5cc5877e38bab7e508029850ab533ba25802))
- **billing:** Pick the stripe sandbox or live variable set with stripe_mode ([ca3bbd3](https://github.com/AlexGalhardo/hexascore-app/commit/ca3bbd36c9388cc3e2260e4740ce8b957f59920a))
- **ui:** Make trending cards one link and invert item cards on hover ([0a508ce](https://github.com/AlexGalhardo/hexascore-app/commit/0a508ce7c735a6d5658cba7473e7a02d82fe38c4))
- **auth:** Sign up with the nickname only ([64f9b81](https://github.com/AlexGalhardo/hexascore-app/commit/64f9b8140192d5089184a7f3656cbfd8895f2308))
- **auth:** Ask the two-factor code in a six-box dialog on the login page ([9e04afc](https://github.com/AlexGalhardo/hexascore-app/commit/9e04afc42aded3319962e264cedfd15941375f4e))
- **search:** Suggest eight items in the ctrl+k palette ([66d94fe](https://github.com/AlexGalhardo/hexascore-app/commit/66d94fe1bd3c69b3de44724c694beaba1ccf8013))
- **auth:** Centre the login and sign-up forms on their own ([3d1f142](https://github.com/AlexGalhardo/hexascore-app/commit/3d1f142b6b8b6bec5eb3d953509caed009e3c109))
- **catalog:** Turn every entity page into a filter of the section listings ([9fac666](https://github.com/AlexGalhardo/hexascore-app/commit/9fac66659e63001e34b7618dc980f00bba121aa5))
- **ui:** Faster landing carousels and a transparent cover letterbox ([595cd6b](https://github.com/AlexGalhardo/hexascore-app/commit/595cd6b33d836dbb141ebf1c9823b7076015459e))
- **api:** Readable operation titles in the api reference ([0ba60c7](https://github.com/AlexGalhardo/hexascore-app/commit/0ba60c7cb3e46d8f92515ebf96fa6fc43b7ef953))
- **ui:** Two-row trending cards with rank and score under the title ([a859669](https://github.com/AlexGalhardo/hexascore-app/commit/a859669dec3f8a96112b6364d8ef98d9d4357a41))
- **ui:** Hide the currency label on pricing ([b9359f2](https://github.com/AlexGalhardo/hexascore-app/commit/b9359f2a8bfc8b35fdecfbdac74ee9db57888d46))

### Changed

- **ui:** Rework the item page and drop more like this ([56b8714](https://github.com/AlexGalhardo/hexascore-app/commit/56b87143acdbb8a30914fdb49dfb72eb6bc891ab))

### Fixed

- **catalog:** Clamp listing pages and keep only the section title ([d5a9659](https://github.com/AlexGalhardo/hexascore-app/commit/d5a965937a94ac4165b012f4eb3fd4ce0f67b489))
- **api:** Align time to beat bounds and docs with open reports ([5141a87](https://github.com/AlexGalhardo/hexascore-app/commit/5141a87e13071754cb7ed1da686b4aa708ef6c48))
- **ui:** Generate the stronger hover token from tokens.json ([d8d6b5e](https://github.com/AlexGalhardo/hexascore-app/commit/d8d6b5e09f409be139628492da3ce1dae0b2e667))
- **i18n:** Keep the keys the android app still reads and resync its catalogues ([745dcb3](https://github.com/AlexGalhardo/hexascore-app/commit/745dcb3e6032375f8bc9820149bbf4117ceb28a1))
- **billing:** Read the checkout currency from json bodies ([42aeda7](https://github.com/AlexGalhardo/hexascore-app/commit/42aeda77eb55d3fd74fa423b26b70974c94fdfa9))
- **ui:** Make primary buttons' depth visible and guard the speedrun migration ([cc7c3c3](https://github.com/AlexGalhardo/hexascore-app/commit/cc7c3c3441b9f85a4081ad7b08105d80fcefea3a))
- **catalog:** Let browsers render the sitemap xml tree ([0521ab7](https://github.com/AlexGalhardo/hexascore-app/commit/0521ab7c34002e389958ae67c16d78f267c48de4))
- **billing:** Drop the renewal date of a canceled subscription ([d24ce7f](https://github.com/AlexGalhardo/hexascore-app/commit/d24ce7f6b87f8753ffa118ba9d4f8527bdc4a7d8))

## [0.37.0] - 2026-11-01

### Added

- **infra:** Alert on telegram at each railway spend threshold ([6990e02](https://github.com/AlexGalhardo/hexascore-app/commit/6990e027c8bb175d76750646ed5c64211f821dd6))
- **ui:** Move login into the More menu and simplify the landing sections ([11b307d](https://github.com/AlexGalhardo/hexascore-app/commit/11b307db31eca40200a46376834b970fb36c8fda))
- **ui:** Show only the site icon in the mobile navbar ([5ed7da5](https://github.com/AlexGalhardo/hexascore-app/commit/5ed7da5c02699d02728f0620035b87d800ec35a4))
- **ingest:** Finish edition matching with publishers, primary and open library data ([267466f](https://github.com/AlexGalhardo/hexascore-app/commit/267466fb14c7a89a0306f6c4d9f00c51244f9ce6))

### Changed

- **catalog:** Rekey covers 16 items at a time ([b3d2cf9](https://github.com/AlexGalhardo/hexascore-app/commit/b3d2cf96df1befd89a9bb64c28d9c7a35c96c853))

### Fixed

- **infra:** Alert when the projected railway spend passes us$18 ([707b092](https://github.com/AlexGalhardo/hexascore-app/commit/707b0921299880289abd3a0482b4cf4f6dbcd41f))
- **infra:** Send x-robots-tag only when seo_noindex is set ([648597a](https://github.com/AlexGalhardo/hexascore-app/commit/648597adb2cae236e6af1b3ad047fb8c753365c2))
- **mobile:** Keep category entity pages out of Android App Links ([9a1961c](https://github.com/AlexGalhardo/hexascore-app/commit/9a1961c16eb5b616912253b4f088a56d2ae53b4a))
- **catalog:** Reuse bucket connections in parallel cover rekeys ([cfd1e5b](https://github.com/AlexGalhardo/hexascore-app/commit/cfd1e5bfaf413e9cf4473bd4c7990bc93fa0e9c6))

## [0.36.0] - 2026-11-01

### Added

- **auth:** Take the nickname at sign-up and tighten sign-up bounds ([9a57ddf](https://github.com/AlexGalhardo/hexascore-app/commit/9a57ddfbdafbb03e63a1ff857a964eaaec0b40eb))
- **ui:** Narrow the contact form and take the sender from the session ([5cb9f74](https://github.com/AlexGalhardo/hexascore-app/commit/5cb9f74f7be7b79958f814c2313a7a6601583d0f))
- **ui:** Narrow the profile page into a centred column ([d99725b](https://github.com/AlexGalhardo/hexascore-app/commit/d99725bf226dba448530d8518bd22e1a93766d57))
- **auth:** Move the two-factor settings into the profile page ([c6f505e](https://github.com/AlexGalhardo/hexascore-app/commit/c6f505e3e4ae2651ae0ffbeb9579b977e9bb3214))
- **ui:** Head the rating card with "Your rating" and hide the help text ([5612503](https://github.com/AlexGalhardo/hexascore-app/commit/5612503ef8c02c15d79bc30076560bcb241d8d59))
- **ui:** Lay the item page out in three columns ([95666dc](https://github.com/AlexGalhardo/hexascore-app/commit/95666dcd105d8daac55f3aa4281e33892b82b7ab))
- **ui:** Show the category icon in the item page breadcrumb ([abb87c3](https://github.com/AlexGalhardo/hexascore-app/commit/abb87c329411d0dbef62fc72453dd61bf7dad9f3))
- **ui:** Move the score help to /how-score-works and the policy to /privacy ([2c6019f](https://github.com/AlexGalhardo/hexascore-app/commit/2c6019fd04dc39b502ce719f96d54db21d6c40d7))
- **catalog:** Page the section listings 20 at a time with buttons above and below ([f2e9801](https://github.com/AlexGalhardo/hexascore-app/commit/f2e980100bad1f01dd4cbf7c25dee6e417dee192))
- **catalog:** Add the filter sidebar and the 25/50/75/100 page size to the section listings ([bf9947a](https://github.com/AlexGalhardo/hexascore-app/commit/bf9947a53b8e7e4c30bb32488aa7c47c186ad174))
- **ui:** Redirect missing html pages to the localized home page ([2858377](https://github.com/AlexGalhardo/hexascore-app/commit/285837786875a5caa4f3878767764b3557cf13be))
- **catalog:** Move entity pages under their category ([85bb575](https://github.com/AlexGalhardo/hexascore-app/commit/85bb57511ba1f65354c7d89cd87b4fcbe3718b6d))
- **catalog:** Serve covers by slug at /covers/{section}/{slug}.webp ([ffb929d](https://github.com/AlexGalhardo/hexascore-app/commit/ffb929db20441dc73d81936a5d1a406df77685a8))
- **infra:** Send alerts to telegram page and silent chats with a daily error digest ([5f87434](https://github.com/AlexGalhardo/hexascore-app/commit/5f8743400c6fb855f9229bfca8d530bb96e5559b))
- **infra:** Keep staging out of search engines with SEO_NOINDEX ([a3ac210](https://github.com/AlexGalhardo/hexascore-app/commit/a3ac210a4ab0727d47db50cc041768f82ed9206f))

### Fixed

- **catalog:** Keep community covers out of covers:rekey and revalidate every slug cover URL ([26ea3be](https://github.com/AlexGalhardo/hexascore-app/commit/26ea3be2c3eeb38521537f7a826a0bbd48fa9396))

## [0.35.0] - 2026-11-01

### Added

- **api:** Measure per-field api coverage with the serialiser's source rules ([69ae04f](https://github.com/AlexGalhardo/hexascore-app/commit/69ae04f0a62fee4d196af44e5a21d391fb0790d3))
- **api:** Publish measured coverage on pricing and docs coverage pages ([c1617a6](https://github.com/AlexGalhardo/hexascore-app/commit/c1617a69ea2e43333e92251bdb78a0ee43d15d78))
- **infra:** Mail the owner a 7-day signed link to the monthly dump ([0a85251](https://github.com/AlexGalhardo/hexascore-app/commit/0a85251c0367b5f47d5ce6eca0cdb53b57d29ee9))
- **ingest:** Read language, year and isbn from wikidata edition records ([3a6c259](https://github.com/AlexGalhardo/hexascore-app/commit/3a6c259b611af8df1d0621281b3112d1c02dd5d5))
- **mobile:** Suggest while typing, grouped full search and click events in the search tab ([991bb2b](https://github.com/AlexGalhardo/hexascore-app/commit/991bb2bdb06d57ec91d106b52e680799af854622))
- **legal:** Redirect /legal/privacy to /policy in every locale ([d894e58](https://github.com/AlexGalhardo/hexascore-app/commit/d894e58b2e35083832869c994ea4ea54ddf8e078))
- **ui:** Add three app icon candidates and wire the first as navbar mark and favicon ([3477cc0](https://github.com/AlexGalhardo/hexascore-app/commit/3477cc0c737f29635cec21b8c1c26ea5d4f9405c))
- **ui:** Give every button a pressed state with depth and shadow ([27d8b98](https://github.com/AlexGalhardo/hexascore-app/commit/27d8b98ac008b62a4bd45965b802c619be8e79de))
- **ui:** Remove the recently rated strip from the landing ([2574435](https://github.com/AlexGalhardo/hexascore-app/commit/25744358a67d8597977a664f53485816fddfc9a3))
- **ui:** Show a category icon before every games, movies, tv and books label ([17ef96d](https://github.com/AlexGalhardo/hexascore-app/commit/17ef96daa07f7831506000ffd286dd28bc61e5f0))
- **ui:** Drop the user name from the account dropdown and add link icons ([daf826a](https://github.com/AlexGalhardo/hexascore-app/commit/daf826a6cea575c151a99e66c13041fad55c5d09))
- **ui:** Rank landing cards as trending and show the primary genre by the year ([93a5cd3](https://github.com/AlexGalhardo/hexascore-app/commit/93a5cd3897d2af29b1385bcf01e4693e70cec3f0))
- **ui:** Remove the trending definition line under the landing blocks ([54767b3](https://github.com/AlexGalhardo/hexascore-app/commit/54767b35ff5c030e15b8c67b6157aef5cafe0392))
- **ui:** Add a source code link to the footer ([a4ef718](https://github.com/AlexGalhardo/hexascore-app/commit/a4ef71813496e2fe97c80db63d63302d05ea2b4b))
- **ui:** Add the item tile card with a title, year and full-height score footer ([7d1456c](https://github.com/AlexGalhardo/hexascore-app/commit/7d1456c2bc0193c0bc24cfbebdba6be2d34010a6))
- **ui:** Centre a larger title count in the landing section header ([9ddc139](https://github.com/AlexGalhardo/hexascore-app/commit/9ddc13945f70bfd91811220ab9e283aee8736502))
- **ui:** Add the about page and its footer link ([82f3033](https://github.com/AlexGalhardo/hexascore-app/commit/82f3033878731b399ea16e5d4ad45245870a215a))
- **ui:** Remove the notifications bell and unread badge from the navbar ([e05fd18](https://github.com/AlexGalhardo/hexascore-app/commit/e05fd18497ee73217dfd212d865a86c7011cdf21))
- **ui:** Add the trending carousel and the random more grid to the landing blocks ([e5a5b26](https://github.com/AlexGalhardo/hexascore-app/commit/e5a5b268ba056a99d907579ef8da0dbc1161f337))
- **ui:** Make icon candidate 3 the active app icon and favicon ([0085e4d](https://github.com/AlexGalhardo/hexascore-app/commit/0085e4de7a7d81a08281d7f27cec82a106e90178))
- **mobile:** Use icon candidate 3 for the android app and the play listing ([6550802](https://github.com/AlexGalhardo/hexascore-app/commit/65508028d0309f6e47a83bdcfc9c5708147222f5))

### Fixed

- **repo:** Start the release range at the commit matching the highest tag ([bdbb25b](https://github.com/AlexGalhardo/hexascore-app/commit/bdbb25be4718eb983f89aa6e85c2ef9fb0c61270))
- **infra:** Never cache off-site dump downloads and cover the route in the security inventory ([599897b](https://github.com/AlexGalhardo/hexascore-app/commit/599897b99b1c77e3b5d609df06911f3e05ff8192))
- **infra:** Make the moderation k6 run measure the signed-in queue ([7f503f4](https://github.com/AlexGalhardo/hexascore-app/commit/7f503f408731e8094644442b9c9dfb833c01fb47))
- **ui:** Ship the reviews, comments and reactions strings on every page ([14533f4](https://github.com/AlexGalhardo/hexascore-app/commit/14533f4a27e2ca6a8247d1412d7c54e03111bf0c))
- **api:** Pass the measurement date in the api coverage usage line ([3da5330](https://github.com/AlexGalhardo/hexascore-app/commit/3da53301893e0e4fad0487edcf32ef076cce526b))
- **infra:** Redact link signatures and reset tokens from every log line ([5e20d4a](https://github.com/AlexGalhardo/hexascore-app/commit/5e20d4ae2e1112576b9feb440d92a82b660ca2a6))
- **ui:** Paint icon candidate 3 with the current ink token ([994764e](https://github.com/AlexGalhardo/hexascore-app/commit/994764ea1dcadd274d38bfe5ce53eeeef674a35e))
- **ui:** State the real 0.5 to 10 half-star rating range on the about page ([8657b51](https://github.com/AlexGalhardo/hexascore-app/commit/8657b51142dd399479d7b9003fd777d0f544d12a))

## [0.34.0] - 2026-11-01

### Added

- **seo:** Serve llms.txt and llms-full.txt with the vukem llmstxt package ([40bd3a7](https://github.com/AlexGalhardo/hexascore-app/commit/40bd3a723648d43e2761f31611dd5960b030066e))
- **seo:** Render vukem analytics and verification tags in the root view ([43daf18](https://github.com/AlexGalhardo/hexascore-app/commit/43daf181ef697ff7ed7623d460abddd339a420dc))
- **ui:** Gate GA4 on the cookie banner and add cookie settings ([d457e83](https://github.com/AlexGalhardo/hexascore-app/commit/d457e83eb71dc29ed5da433bff97a020b708bf8b))
- **seo:** Notify IndexNow of catalogue changes from the api_changes feed ([ab54754](https://github.com/AlexGalhardo/hexascore-app/commit/ab54754e530672ef7b8288ae04b4385397b883ba))
- **auth:** Protect sign-up, login and contact with Cloudflare Turnstile ([6bff251](https://github.com/AlexGalhardo/hexascore-app/commit/6bff251d4f7e85e1ae1c0a1416e46080e4d23317))
- **legal:** Keep locale on the delete-account help links and record Play data safety facts ([bae37e8](https://github.com/AlexGalhardo/hexascore-app/commit/bae37e828c675fbf3a2280503c402c101b1fd9d6))

### Changed

- **api:** Encode exports with export.EncodeNDJSON and EncodeCSV ([82930d8](https://github.com/AlexGalhardo/hexascore-app/commit/82930d8925dd3661e17a5e9d8c7f394dbacbdf28))
- **infra:** Add request-line route and user with web.AddLogAttrs ([f4b137f](https://github.com/AlexGalhardo/hexascore-app/commit/f4b137fe7380b39adad641bef6f43aa9986a03d2))
- **infra:** Set the support Reply-To on the vukem mailer ([46bdff7](https://github.com/AlexGalhardo/hexascore-app/commit/46bdff755ad99ecc9b8c8c10a81a699d97ff88da))
- **api:** Code API problems with the vukem ProblemDecorator ([b5b4786](https://github.com/AlexGalhardo/hexascore-app/commit/b5b47864457b9e1fb375f22c4e1371f4a5d32d0f))
- **ui:** Lazy-load the mobile menu sheet, query provider and report dialog ([9eeedbe](https://github.com/AlexGalhardo/hexascore-app/commit/9eeedbefd70aeab2794701991b068db96fa99247))

### Fixed

- **seo:** Satisfy lint in the IndexNow query, Turnstile header and route imports ([dc15d4b](https://github.com/AlexGalhardo/hexascore-app/commit/dc15d4ba46d1e25440db86474c02e7b7825fa6de))
- **infra:** Reach the production-mode QA app on its APP_URL host in CI ([5a03313](https://github.com/AlexGalhardo/hexascore-app/commit/5a03313fa254b33d3e93c799e9bafd64b4419043))
- **mobile:** Make the release JS bundle build from a clean bun install ([1a3c452](https://github.com/AlexGalhardo/hexascore-app/commit/1a3c45207f9277a72376c53efaab752cc60a0a72))
- **mobile:** Fail bundle:check when the bundle file is missing or empty ([7d2245b](https://github.com/AlexGalhardo/hexascore-app/commit/7d2245bb57c522cb5c674b5e0f55c3e9f7476c4c))
- **e2e:** Default Playwright base URL to localhost ([d108b87](https://github.com/AlexGalhardo/hexascore-app/commit/d108b87fd939f6b350813771dd4b76a5cce0e92f))
- **ci:** Submit mobile builds on tag pushes only, not on dispatches from a tag ref ([8e465aa](https://github.com/AlexGalhardo/hexascore-app/commit/8e465aaebe195f03ce0766603eddb45710a9d34a))

## [0.33.0] - 2026-11-01

### Added

- **mail:** Noreply sender, support Reply-To and contact inbox from config ([c40478c](https://github.com/AlexGalhardo/hexascore-app/commit/c40478c16b767dc8e9d46a5d9dd787e8ca39a7be))
- **security:** Guard public pages against scraping ([e80a214](https://github.com/AlexGalhardo/hexascore-app/commit/e80a2141a5e3ddf663765981a582329a53938366))
- **lists:** Open the add-to-list sheet from the mobile item screen ([f17bc11](https://github.com/AlexGalhardo/hexascore-app/commit/f17bc11f4f72f1c2005fb07536af44b674a896b1))
- **lists:** Restyle the mobile lists tab and add-to-list sheet to v1 ([e6ed5e2](https://github.com/AlexGalhardo/hexascore-app/commit/e6ed5e238c2f45ec5f1222bfa51b0e576d8e2625))
- **ui:** Page each mobile entity group with Next ([d648ce3](https://github.com/AlexGalhardo/hexascore-app/commit/d648ce3ba21b5af5f6215d026581585370be1c22))
- **mobile:** Feed search trending from the app trending endpoints ([2c03eb7](https://github.com/AlexGalhardo/hexascore-app/commit/2c03eb72a00066d2dc5f1c4e4ab4e7e49ea915e9))

### Changed

- **ingest:** Push the qids filter into the tmdb, igdb, trailers and openlibrary target SQL ([2a3e93c](https://github.com/AlexGalhardo/hexascore-app/commit/2a3e93c608c9cdacfa7f0940377724799c1df0da))

### Fixed

- **ci:** Fail fast when openapi or lighthouse run without the build job ([fd31cff](https://github.com/AlexGalhardo/hexascore-app/commit/fd31cff7499fa57cf0d9989ca8afe5ff0e71d277))
- **ci:** Keep main green after the P6.B and local.sh merges ([aaa60d0](https://github.com/AlexGalhardo/hexascore-app/commit/aaa60d08a2140522130b4b6934dd47e874cba33e))
- **mobile:** Recopy the bundled catalogues after the friends score keys ([cb5dc4c](https://github.com/AlexGalhardo/hexascore-app/commit/cb5dc4cbff5f51462acac30658411e43a6754105))
- **security:** Give the signed-in page budget only to a logged-in session ([d413a85](https://github.com/AlexGalhardo/hexascore-app/commit/d413a85826587bec62c81292ab829b27971ae0b0))
- **ui:** Apply the App Link locale to the mobile session ([8541a0a](https://github.com/AlexGalhardo/hexascore-app/commit/8541a0aed833ed42d2e283be078d498e3fb140e2))
- **ui:** Show loading while mobile search trending loads ([e83c5a0](https://github.com/AlexGalhardo/hexascore-app/commit/e83c5a0090ffb8487aa0d934198f1e17768b515d))
- **ingest:** Write api_changes rows from normalize and rollback ([e3ca34e](https://github.com/AlexGalhardo/hexascore-app/commit/e3ca34ed4f4d727a1a711fa946488961149e1870))
- **covers:** Reserve cover_gc keys outside the step transaction ([b3d6756](https://github.com/AlexGalhardo/hexascore-app/commit/b3d675696a568e57bfffec784a11c82fafe3547b))
- **ingest:** Keep running A7 checks after one fails ([94508af](https://github.com/AlexGalhardo/hexascore-app/commit/94508af155981acc9a99d96a788212b1d0862c2c))
- **ingest:** Close rollback and /changes gaps ([7476bd5](https://github.com/AlexGalhardo/hexascore-app/commit/7476bd574fdce619e0933f22d9fdc155b9f8a4d1))

## [0.32.0] - 2026-11-01

### Added

- **social:** Show the friends score on the item page and the home behind the social flag ([adf70a1](https://github.com/AlexGalhardo/hexascore-app/commit/adf70a1242c35a987441ae01c78e55c8385b31b0))
- **ingest:** Add the cwebp WebP encoder with resource limits and a fake-binary test ([af09956](https://github.com/AlexGalhardo/hexascore-app/commit/af09956c04b6d74876d935c5c92ad0e359884281))
- **ingest:** Add the youtube trailer check connector and weekly chain ([cd55529](https://github.com/AlexGalhardo/hexascore-app/commit/cd555295d4ae302c2b3eadc2c0a7fd543a399c0f))
- **ingest:** Add the dq.asserts job with DQ-01..14 and DQ-I-01..13 checks ([68e5a17](https://github.com/AlexGalhardo/hexascore-app/commit/68e5a174bc4e6d02075f6693403d314308a19395))
- **admin:** Add the ingestion match review queue ([762dcc1](https://github.com/AlexGalhardo/hexascore-app/commit/762dcc16f3ce9c5dc4e0b4981fa429492d9f8211))
- **ingest:** Add the dq.frontend_fields missing-field detector ([def289d](https://github.com/AlexGalhardo/hexascore-app/commit/def289dca0be5b1e281082f0fe80bc20082df86f))
- **admin:** Add the /admin/ingest status page, run trigger and cancel ([8b663bd](https://github.com/AlexGalhardo/hexascore-app/commit/8b663bd8ae74259ec0c4042bf95b9bf738c644a2))
- **ingest:** Honour the qids param of job_runs.params in fetch runs ([7900603](https://github.com/AlexGalhardo/hexascore-app/commit/7900603c2f336a9106abbbe03e00d5f76f43cffc))
- **ingest:** Add rollback, retry_failures, prunes, pressure gate and the A7 alert ([11ed9aa](https://github.com/AlexGalhardo/hexascore-app/commit/11ed9aa2b05d875c79788ecb40f3c2aceae8579a))
- **ingest:** Port the open library connector behind ingest_openlibrary ([d9f04ea](https://github.com/AlexGalhardo/hexascore-app/commit/d9f04ead231a609161f0d6e0ad04db764b57a664))
- **ingest:** Port the igdb connector and the d-36 igdb.refresh job ([7e259fa](https://github.com/AlexGalhardo/hexascore-app/commit/7e259faff2b9b8e404cee89ed32094ac97074dd6))
- **ingest:** Port the steam appdetails and gog catalogue connectors ([c825d69](https://github.com/AlexGalhardo/hexascore-app/commit/c825d69e81c0a8038bbf89b350747b5c32c9d8de))
- **ingest:** Port the covers job with deferred delete and wire dq stat to the bucket ([cd89352](https://github.com/AlexGalhardo/hexascore-app/commit/cd8935203d7b9aee064d9d426ebe66f6b4ca7b52))

### Changed

- **vukem:** Drop stale workarounds fixed by v0.35.0 ([6a0da08](https://github.com/AlexGalhardo/hexascore-app/commit/6a0da08c74e0050e131f15224bcf2cd762ad294f))
- **vukem:** Adopt native recovery hash and queue limits, retag stale VF ids ([6e6dc54](https://github.com/AlexGalhardo/hexascore-app/commit/6e6dc546bca67061f47c84f7af8b5571ae3c09eb))

### Fixed

- **api:** Answer an unknown export type with 400 invalid-query without TypeGuard ([61d9b03](https://github.com/AlexGalhardo/hexascore-app/commit/61d9b03f367a1c69e74cb8ea956be940a969630b))
- **social:** Gate review and comment reports and add an unhide outcome ([de6bdda](https://github.com/AlexGalhardo/hexascore-app/commit/de6bdda8b7c0bc74dbf3898378e0cd5c5f38cf03))
- **social:** Word the friends score summary without a plural ([677a2e9](https://github.com/AlexGalhardo/hexascore-app/commit/677a2e93d6c070f8c15da99e34d5c6a5d02aa737))
- **ingest:** Guard edition QIDs held by another work like ISBN conflicts ([cdd96a6](https://github.com/AlexGalhardo/hexascore-app/commit/cdd96a69854a5a3d5dcdc73332830310827b9279))
- **ingest:** Guard reviewed items on rollback and bound pressure samples ([e57e6d6](https://github.com/AlexGalhardo/hexascore-app/commit/e57e6d6a944700700ad378eeea1221662509c7c4))
- **ingest:** Give each outbound attempt its own deadline instead of a client timeout ([7cd6a27](https://github.com/AlexGalhardo/hexascore-app/commit/7cd6a27b971b511c030858d24a56c97d1954326d))

## [0.31.0] - 2026-11-01

### Added

- **ui:** Show a Feed nav link to signed-in users while the social flag is on ([4468273](https://github.com/AlexGalhardo/hexascore-app/commit/44682732ffaf465c9f2a06e130b63686c82739dc))
- **ui:** Restyle the follow bar, follow requests, feed and activity cards to design v1 ([ffa27a2](https://github.com/AlexGalhardo/hexascore-app/commit/ffa27a2d9e17799d30e6c9c41d38652b3a32762b))
- **social:** Add the feed opt-in, profile activity tab, retention and export behind the social flag ([5fd65e9](https://github.com/AlexGalhardo/hexascore-app/commit/5fd65e954df289b6819067b05a7d8c030202dbb2))
- **social:** Add reviews with a Markdown subset behind the social flag ([828101b](https://github.com/AlexGalhardo/hexascore-app/commit/828101b39f82ecb7a9de0a7500c1e28c979ef158))
- **social:** Add text-free share cards and a share action behind the social flag ([690cb37](https://github.com/AlexGalhardo/hexascore-app/commit/690cb37798816ca994c17cef253919aa25cd1b27))
- **social:** Add comments and reactions on reviews and lists behind the social flag ([949b648](https://github.com/AlexGalhardo/hexascore-app/commit/949b648733ff9530e696acba42ca83d08fbfcdbe))
- **social:** Moderate reported reviews and comments behind the social flag ([fe643b3](https://github.com/AlexGalhardo/hexascore-app/commit/fe643b30ffa8083447df615447880df60e7ea1a0))
- **web:** Replace hardcoded indigo classes with the primary token ([b0753a0](https://github.com/AlexGalhardo/hexascore-app/commit/b0753a0e20df59d4f2990b2dd3e4721ae3cbecd3))
- **api:** Serve listExports and getExport with a presigned redirect ([796228d](https://github.com/AlexGalhardo/hexascore-app/commit/796228d26cb41a09ff6444d8e014a2a0c5a05d07))

### Changed

- **recommend:** Log reco impressions and clicks as structured events ([189061e](https://github.com/AlexGalhardo/hexascore-app/commit/189061edfa10a0cc2a31a2c4da9102632caa6cab))

### Fixed

- **web:** Label other users' shelf scores neutrally ([b282225](https://github.com/AlexGalhardo/hexascore-app/commit/b282225630427de7b2b8949e89f8621eff6340d5))
- **ingest:** Keep the entity name predicate in internal/search and clear gate findings ([61498eb](https://github.com/AlexGalhardo/hexascore-app/commit/61498ebcf52a066fcda75fb43b6497fa9b3cd595))
- **ui:** Keep the lg navbar inside 1024 px with the feed link ([ac04007](https://github.com/AlexGalhardo/hexascore-app/commit/ac04007a3b394147e82a58f73c67d31c4f63090e))
- **ui:** Drop the account name beside the avatar when feed is inline ([beb11eb](https://github.com/AlexGalhardo/hexascore-app/commit/beb11eb77416b38fbf4a90ef9cb83d0b1d3eb9d0))
- **ui:** Collapse the mobile navbar into title, search field and hamburger sheet ([73582e3](https://github.com/AlexGalhardo/hexascore-app/commit/73582e3aa74d526d8058d2d55992c2c7e552341f))
- **ui:** Keep tab focus inside the mobile menu sheet ([2184d24](https://github.com/AlexGalhardo/hexascore-app/commit/2184d24ef7174f43c89af16b316a9aa0f13a66ab))
- **social:** Translate the target type in moderation notices ([573ac88](https://github.com/AlexGalhardo/hexascore-app/commit/573ac884b2638c8cb8df1c5b26b69864b4de5ae2))
- **api:** Keep popularity out of the daily id exports until the owner confirms it ([cefdf40](https://github.com/AlexGalhardo/hexascore-app/commit/cefdf404b05c69a1a227e065ff9a74b224b3e72c))

## [0.30.0] - 2026-11-01

### Added

- **ingest:** Add the ingestion chains, schedules and ingest:bootstrap command ([d9aa31e](https://github.com/AlexGalhardo/hexascore-app/commit/d9aa31e38e516f2d27b2c2d782ea29057d783e0e))
- **ingest:** Wire the ingest job registration and Runner steps into the app ([b24cbbd](https://github.com/AlexGalhardo/hexascore-app/commit/b24cbbdd2df0b2c7b89ee36afa16c40063b2aad6))
- **ingest:** Add the TMDB connector and the tmdb.expire job ([f9f0719](https://github.com/AlexGalhardo/hexascore-app/commit/f9f07199aa752a6ffa58f199333fc2c322531357))
- **ingest:** Resolve credits to entities and write item_entities ([5a38486](https://github.com/AlexGalhardo/hexascore-app/commit/5a384866130575ce61546f6638aa22b83d794b20))
- **ingest:** Add edition ISBN normalisation, selection and the ISBN conflict guard ([3567e47](https://github.com/AlexGalhardo/hexascore-app/commit/3567e47c8952f92e822e9e40bdc2f8e29269513f))
- **ingest:** Extract editions from wikidata and add backfill.editions ([c3e648c](https://github.com/AlexGalhardo/hexascore-app/commit/c3e648cda6b6cc309367e02a350953b5b57f753f))
- **ui:** Adopt tanstack query for client-fetched web data ([21c6ec3](https://github.com/AlexGalhardo/hexascore-app/commit/21c6ec3bc0185040f10a8a3f7d48317499f61fc8))
- **social:** Add follows with requests for private profiles behind the social flag ([0345aa9](https://github.com/AlexGalhardo/hexascore-app/commit/0345aa9e46a9ca7a4ffcf62f01bc7aa2886e1eb9))
- **social:** Record follow activity and add the /feed page behind the social flag ([dc8cd66](https://github.com/AlexGalhardo/hexascore-app/commit/dc8cd667b588c1dfd14691c9b0bfbf3c6fe9c9a9))
- **mobile:** V1 look for the status and time-to-beat sheets ([ab4bd65](https://github.com/AlexGalhardo/hexascore-app/commit/ab4bd651d77ff09292259dfc11135a36bd41c2a3))
- **mobile:** V1 look for account extras and an unread badge on the profile tab ([848e74f](https://github.com/AlexGalhardo/hexascore-app/commit/848e74fda8cc190e1808a287d88acf3c2bbb0c4c))
- **mobile:** Search i18n labels for type, recent removal and empty trending ([fb5a58f](https://github.com/AlexGalhardo/hexascore-app/commit/fb5a58f3f08c7fe230ad52adf8d80f48e9aaf280))

### Changed

- **ingest:** Replace the jobs.IngestOverride package variable with bootstrap.NewWithIngest ([f078a08](https://github.com/AlexGalhardo/hexascore-app/commit/f078a087805cdd121a57a794bd3cdc4959a66f1c))
- **ui:** Load the search palette on first open ([1b959e7](https://github.com/AlexGalhardo/hexascore-app/commit/1b959e7557a98a0e9925ccd549a22db882ee70d5))
- **web:** Gzip anonymous public catalogue html ([dd0de0c](https://github.com/AlexGalhardo/hexascore-app/commit/dd0de0cb5620f79af94866924cfb4654fb079bf0))
- **web:** Lazy-load the search palette and darken the light book accent for contrast ([b9e9422](https://github.com/AlexGalhardo/hexascore-app/commit/b9e94228aecb7a4fdd99ac2f718f27dbfd18ae16))

### Fixed

- **ingest:** Cache the Wikidata selection per run instead of per chunk ([965dc67](https://github.com/AlexGalhardo/hexascore-app/commit/965dc671b60688e463256bb241125f4d8f018375))
- **ingest:** Scope the Wikidata entity memo to one Fetch call ([a136c50](https://github.com/AlexGalhardo/hexascore-app/commit/a136c5023db9f7cdfe65268aa83063a6e857bb82))
- **ingest:** Persist the incremental since in the run params when a chain starts ([0c6d6e9](https://github.com/AlexGalhardo/hexascore-app/commit/0c6d6e9dd02251ef74ce94aadc8a392beaeca60a))
- **ingest:** Lock the run cursor, probe-only breaker close and outage errors abort the chunk ([44f5533](https://github.com/AlexGalhardo/hexascore-app/commit/44f5533685e9fbb01e3434d4e299b0e2d42a0417))
- **social:** Classify the follow and feed routes and format the flag list ([6aa724c](https://github.com/AlexGalhardo/hexascore-app/commit/6aa724c47ff287923c99ff4b69a56a8ea95a81d6))

## [0.29.0] - 2026-10-31

### Added

- **mobile:** Serve assetlinks.json and declare verified App Links ([2bc5446](https://github.com/AlexGalhardo/hexascore-app/commit/2bc544678b06f2eb7bdc0967609e946b9a81d938))
- **ui:** Wire the mobile search tab to the app search API ([97aebd3](https://github.com/AlexGalhardo/hexascore-app/commit/97aebd369ef46015039b7442233635d398251ecf))
- **ui:** Show search hit year and score on mobile ([77e0927](https://github.com/AlexGalhardo/hexascore-app/commit/77e0927829a0ed9a519a69e39ab9489fc218e840))
- **ui:** Add a per-group next control to the mobile entity page ([b2123da](https://github.com/AlexGalhardo/hexascore-app/commit/b2123da6560716df424af6743fa00ab3d46259e0))
- **docs:** Translate nav labels for the new guides ([375a1c1](https://github.com/AlexGalhardo/hexascore-app/commit/375a1c1187eaeaedf7c5ce7d20ed166790485037))
- **ingest:** Add run framework with continuation, dead letters and kill switches ([3f4038e](https://github.com/AlexGalhardo/hexascore-app/commit/3f4038ef23b83a47cb397e9a36fd8c5388662d64))
- **ingest:** Add outbound client with per-host throttle, retry and breaker ([cf14bf6](https://github.com/AlexGalhardo/hexascore-app/commit/cf14bf649af57b27b5fbc08f2407413c03550c06))
- **ingest:** Add connector interface and wikidata selection ([6761fc6](https://github.com/AlexGalhardo/hexascore-app/commit/6761fc673a17cd5d2b642b3d4f2551bbcc262014))
- **ingest:** Fetch and trim wikidata entities ([fd9caa3](https://github.com/AlexGalhardo/hexascore-app/commit/fd9caa32a338735766e5bde0d5cbed1959bb2263))
- **ingest:** Add wikipedia extracts connector ([d23df7c](https://github.com/AlexGalhardo/hexascore-app/commit/d23df7c3544cfecf3a3bf4247ef04017370d927b))
- **ingest:** Add item core mapping, slug and release dates ([9ab5233](https://github.com/AlexGalhardo/hexascore-app/commit/9ab5233cb676f012dc8c60427097dcd225591598))
- **ingest:** Extract credits, alt titles and based_on links ([0d9c2c8](https://github.com/AlexGalhardo/hexascore-app/commit/0d9c2c8a955d942787d05298ad6d86b21fbb5ada))
- **ingest:** Merge enrichment sources with provenance and community locks ([1279406](https://github.com/AlexGalhardo/hexascore-app/commit/12794068dfc3123661e29923cdfe393ed23fa205))
- **ingest:** Add the normalize job with slug collision, provenance and batch events ([d8187da](https://github.com/AlexGalhardo/hexascore-app/commit/d8187daff0f406084b928b799c8eeeb27c36c3b1))

### Fixed

- **mobile:** Integrate the P5.B batch with the v1 restyle and the locale store ([ce6ff62](https://github.com/AlexGalhardo/hexascore-app/commit/ce6ff62e8bdf49bd8f3cf0c122fcc67afacde7f7))
- **sdk:** Pin typescript 6.0.3 for openapi-ts and sync SDK version with the contract ([be6f8dc](https://github.com/AlexGalhardo/hexascore-app/commit/be6f8dc692373960c7cc022da507a51a5e87336f))
- **sdk:** Sync Python SDK version and README quota header with the contract ([4ddefa4](https://github.com/AlexGalhardo/hexascore-app/commit/4ddefa447b8c5f447d4dd46c5fae5f09c63bf4ca))
- **api:** Answer 301 on /problems/{code} ([1e63889](https://github.com/AlexGalhardo/hexascore-app/commit/1e63889d5895305fc1856f6e0c6434c49981ecf6))
- **ingest:** Keep title claims and latn aliases in the wikidata trim ([97e087f](https://github.com/AlexGalhardo/hexascore-app/commit/97e087ffc2aa72a0574494e992a1cc8e720977ef))
- **ingest:** Keep community fields the source lacks and run suggestions in the caller tx ([efad03f](https://github.com/AlexGalhardo/hexascore-app/commit/efad03f26967701fce84d9bb6bb57b52792f7ad0))
- **ingest:** Isolate bad records, keep cancelled and locked fields, retire only read record versions ([49f4b1f](https://github.com/AlexGalhardo/hexascore-app/commit/49f4b1f26354489c965e7d87a3bf0b86ccd3c980))

## [0.28.0] - 2026-11-01

### Added

- **search:** Add release year and community score to result hits ([5559b53](https://github.com/AlexGalhardo/hexascore-app/commit/5559b53084c027c703ca0be02a3ae0d01f678fc4))
- **library:** Add community score to library items, profile ratings and list entries ([191341f](https://github.com/AlexGalhardo/hexascore-app/commit/191341f486b447e4bf03468491a593b8b3bc1c4f))
- **catalog:** V1 editorial look for the edition and entity pages ([2eba3de](https://github.com/AlexGalhardo/hexascore-app/commit/2eba3de4f8cef0d7940060a6410d9b89483ef53e))
- **catalog:** V1 editorial table for the recent changes page ([499bb26](https://github.com/AlexGalhardo/hexascore-app/commit/499bb2624f320f7d32869cf0a223d6d89aee3cc2))
- **ui:** V1 docs, legal, error and pricing pages ([ca73cd0](https://github.com/AlexGalhardo/hexascore-app/commit/ca73cd061c262c85df163424e9133f3382a2ac20))
- **catalog:** Serve similar and cross-media blocks on the item page and /app/v1/items/{id}/related ([f1ea92f](https://github.com/AlexGalhardo/hexascore-app/commit/f1ea92fd53f126069bf6e835475f0dc1aa1c597c))
- **catalog:** More like this and cross-media strips on the item page ([49a2a49](https://github.com/AlexGalhardo/hexascore-app/commit/49a2a49af289d15e1a7e4a2c3226e6cdad6cd0a7))
- **mobile:** Build the profile tab with GET/PATCH /me, devices and consent toggles ([ea30c13](https://github.com/AlexGalhardo/hexascore-app/commit/ea30c130b68cdf0b715632709901c583a76fcb50))
- **mobile:** Add account export, deletion countdown, about and notifications ([be5cc55](https://github.com/AlexGalhardo/hexascore-app/commit/be5cc5532ecb3ed9824d4c54ecde0a2adb3e984d))
- **ratings:** Add the mobile status and time-to-beat sheets ([1214eb8](https://github.com/AlexGalhardo/hexascore-app/commit/1214eb8583bfd411a457b975c27c866e919e6abe))
- **catalog:** Wire the mobile entity screen ([a0759b4](https://github.com/AlexGalhardo/hexascore-app/commit/a0759b46055039a1437cc077507c08be02656e41))
- **lists:** Build the mobile lists tab, list detail and add-to-list sheet ([a4bcecb](https://github.com/AlexGalhardo/hexascore-app/commit/a4bcecb1f88a865906dcc79019065e0f563e32ed))

### Changed

- **import:** Make the legacy entities stage set-based with --workers ([5704e8a](https://github.com/AlexGalhardo/hexascore-app/commit/5704e8adf0b7714a75b771d0df4791d06255ade4))

### Fixed

- **i18n:** Match the updatedLabel key in the shared page namespaces ([53d584f](https://github.com/AlexGalhardo/hexascore-app/commit/53d584ffe3d3f68f561139a42addf16f4559dff8))
- **import:** Sort legacy edition inserts by isbn13 and document the entities one-off run ([fc25b6d](https://github.com/AlexGalhardo/hexascore-app/commit/fc25b6d54373f6fc1086827e1cf00699f56d1445))
- **openapi:** Regenerate the app golden and client for search hit year and score ([375abb2](https://github.com/AlexGalhardo/hexascore-app/commit/375abb28a44dec3d1a481e21a19cd592477211ae))
- **catalog:** No field label on merge rows of the recent changes table ([3e770a5](https://github.com/AlexGalhardo/hexascore-app/commit/3e770a528df296da5366d9681959aab042ccb3cb))
- **catalog:** Align the edition and entity ghost links with the text edge ([58971b7](https://github.com/AlexGalhardo/hexascore-app/commit/58971b79acdedff46636aa3601ebbbde1f849029))
- **catalog:** Align the edition and entity ghost links with the text edge ([54e3e99](https://github.com/AlexGalhardo/hexascore-app/commit/54e3e99bac6f8b8336d45cba511f4f38cd7955fe))
- **catalog:** Empty entity pages send an empty group list instead of null ([15f1d38](https://github.com/AlexGalhardo/hexascore-app/commit/15f1d3838f42a2cd96161cc2f09330bc24a79e2b))
- **ui:** Ship the updatedLabel string to the legal pages ([e0dbec0](https://github.com/AlexGalhardo/hexascore-app/commit/e0dbec0adff2f6bacff2c5a4624b7a6354d84bae))
- **ui:** Ship the updatedLabel string to the legal pages ([2bce82b](https://github.com/AlexGalhardo/hexascore-app/commit/2bce82b332e266c784523c7fcaa4d5282b882469))
- **i18n:** Recopy the mobile catalogues after the web v1 catalog and content pages ([445f1aa](https://github.com/AlexGalhardo/hexascore-app/commit/445f1aabe37d3bccd59e1021fe7e7c68649953b4))

## [0.27.0] - 2026-10-31

### Added

- **ui:** Wire the v1 home sections to trending data on mobile ([9a87294](https://github.com/AlexGalhardo/hexascore-app/commit/9a87294978a3a2583d4a259971fc2fe89a37e86a))
- **ui:** Wire the v1 catalogue page into the mobile section listing ([4f22bf7](https://github.com/AlexGalhardo/hexascore-app/commit/4f22bf7bc5e0feadfd84b4d65931e78d0061e8cb))
- **ui:** Wire the v1 item detail into the mobile item screen ([4000265](https://github.com/AlexGalhardo/hexascore-app/commit/40002656dc4595790e5e0e548389533a30a5a89c))
- **auth:** V1 login, signup and recovery pages with the Google button ([c51afb8](https://github.com/AlexGalhardo/hexascore-app/commit/c51afb8cc7cfa8a4d6172b2e1bf71ba04bb1f5ab))
- **ui:** V1 search page with grouped results, filters and no-results state ([1b41a15](https://github.com/AlexGalhardo/hexascore-app/commit/1b41a153011301a1205be9fd2bae9dba2799cc7d))
- **lists:** V1 library, list pages and shelf cards ([a6cc943](https://github.com/AlexGalhardo/hexascore-app/commit/a6cc9430878ff235f4129a647cb0eaaf2986f62f))
- **alerts:** Log alert=A5, A6 and A10 at error level ([0b57fca](https://github.com/AlexGalhardo/hexascore-app/commit/0b57fca8cc79ce2c911f5a63b5b60668b16dbc4f))
- **search:** Add search:reindex command ([9c9ddc0](https://github.com/AlexGalhardo/hexascore-app/commit/9c9ddc069ba6fbab501d46d81b08b1a3faf79a7e))

### Changed

- **web:** Send only the i18n namespaces each page uses ([0ae212d](https://github.com/AlexGalhardo/hexascore-app/commit/0ae212d5372dbbb5367d775f80c5ead1910cc818))

### Fixed

- **ui:** Larger covers on the home trending cards ([8a25b0e](https://github.com/AlexGalhardo/hexascore-app/commit/8a25b0eb3784fc5bdc1cdc1db1e8d9a62acc90b1))
- **api:** Send cover urls on app catalogue items ([436b338](https://github.com/AlexGalhardo/hexascore-app/commit/436b33804ca9370d03494a68ef08bcf8c77114a3))
- **api:** Declare the app search query parameters ([bcb0048](https://github.com/AlexGalhardo/hexascore-app/commit/bcb00480f7834dd1129181fc630083a64d00d5e0))
- **api:** Resolve item slugs on the app item read ([3dbe067](https://github.com/AlexGalhardo/hexascore-app/commit/3dbe0671380caf99977281cd81cb0db47a95b8f7))
- **mobile:** Open web app links that carry only the item slug ([8fb5703](https://github.com/AlexGalhardo/hexascore-app/commit/8fb570310463dac9ac0383d19a1d5d3bd985a5a6))
- **api:** Regenerate the app golden and client after the app batch merges ([a807949](https://github.com/AlexGalhardo/hexascore-app/commit/a807949d5e1ea9199cb00a9af2c31c20540fae9e))
- **ui:** Bundle the score namespace in the mobile catalogues ([9bfdcb0](https://github.com/AlexGalhardo/hexascore-app/commit/9bfdcb08b7847d90543028042f6cb44ba0dde410))
- **ui:** Send the rating accessibility value in integer half steps ([866ae1d](https://github.com/AlexGalhardo/hexascore-app/commit/866ae1db45594510a16eb713abecc36ab72864e8))
- **ui:** Keep only named genre entities as item chips ([1f5fd1a](https://github.com/AlexGalhardo/hexascore-app/commit/1f5fd1ac285d422186bc767cfd2ad604ad1a96fb))
- **mobile:** Follow the app client renames and show item covers ([c8e2946](https://github.com/AlexGalhardo/hexascore-app/commit/c8e2946d1ce0ce773d2f2a77c6f3c10614b3d8ce))
- **auth:** Mobile members layout without the card box and larger desktop heading ([fe7ed7f](https://github.com/AlexGalhardo/hexascore-app/commit/fe7ed7fbdf46ab82e00182e8c0a34c4502b0e9eb))
- **ui:** Plural section names and sans counts on the search page ([312793a](https://github.com/AlexGalhardo/hexascore-app/commit/312793a354a22b72a09742eec41b01769de81d61))
- **lists:** Regular weight in the library filter selects ([d09e194](https://github.com/AlexGalhardo/hexascore-app/commit/d09e194621eeb5e7a3e79385dcc8f7eeb24accd6))
- **i18n:** Recopy the mobile catalogues after the web v1 auth, search and library merges ([8d1b26f](https://github.com/AlexGalhardo/hexascore-app/commit/8d1b26ffe2d2bfa4e3a2a09252fca7f7de3130d0))
- **ssr:** Strip the locale prefix before the A5 public route check ([b22ea64](https://github.com/AlexGalhardo/hexascore-app/commit/b22ea6445fc88e6e79cba5c8a1feaa338448a6fc))

## [0.26.0] - 2026-10-31

### Added

- **mobile:** Add the home screen with trending sections per type ([6419293](https://github.com/AlexGalhardo/hexascore-app/commit/64192933231699cc0b4d7823b26d2b15bbc65efe))
- **mobile:** Add the section listing with sort, genre and cursor paging ([f90d499](https://github.com/AlexGalhardo/hexascore-app/commit/f90d499ce56a8a0299716152ef1e0c77567a5343))
- **mobile:** Add the item screen with score, entities, editions and own state ([bc99bb3](https://github.com/AlexGalhardo/hexascore-app/commit/bc99bb3b24479760d585291658ff1aa8d053b0b7))
- **ui:** Add v1 mobile catalogue page screen component ([c259f36](https://github.com/AlexGalhardo/hexascore-app/commit/c259f369f6b645d726fbb81fddb2cbc32e13daf6))
- **ui:** Add v1 mobile item detail screen component ([2e47fef](https://github.com/AlexGalhardo/hexascore-app/commit/2e47fef0356811535ab54eb9a891c70db990b4d6))
- **ui:** Add v1 mobile library screen component ([e131563](https://github.com/AlexGalhardo/hexascore-app/commit/e131563a25d197503105cd5ef4084d195f3c935e))
- **ui:** Add v1 mobile auth and settings screen components ([dcad7cd](https://github.com/AlexGalhardo/hexascore-app/commit/dcad7cd220578923ecf2158f91164f07b1560d4b))
- **ui:** Add v1 mobile state page and restyle blocked screens ([569d038](https://github.com/AlexGalhardo/hexascore-app/commit/569d0381caa17411a11637c21dfb68d85f42f3a1))
- **ui:** Add v1 mobile entity, list detail, notifications and editions components ([a868619](https://github.com/AlexGalhardo/hexascore-app/commit/a8686199e58dbcfec5b60ef8771543692dbe7df8))
- **ui:** Wire the v1 settings page into the mobile profile tab ([f9e24a0](https://github.com/AlexGalhardo/hexascore-app/commit/f9e24a0d43e9fcf690a2cfdbd09bf9fb05a52d21))
- **ui:** Wire the v1 mobile header and tab bar ([11dd2a0](https://github.com/AlexGalhardo/hexascore-app/commit/11dd2a06de8b90d30230a8a9b023a1d6848c53f8))
- **ratings:** Add the mobile rating data flow over the offline queue ([53347d3](https://github.com/AlexGalhardo/hexascore-app/commit/53347d31137b168a669cb8969fa6105b13f03f45))
- **ratings:** Add the mobile rate sheet and wire the offline queue ([847edba](https://github.com/AlexGalhardo/hexascore-app/commit/847edba3cbb1920316ce3712f83c1345df160193))
- **ui:** V1 editorial shell with fonts, button and fixed navbar ([b698e05](https://github.com/AlexGalhardo/hexascore-app/commit/b698e050989b36e234931044b6b656ddd5ff1f42))
- **ui:** Ctrl k search command palette in the navbar ([214eb5d](https://github.com/AlexGalhardo/hexascore-app/commit/214eb5d90cf1573085c15fb7feda06bd7662560a))
- **ui:** V1 catalogue listing with title alone, filter bar and score badges ([b22aea9](https://github.com/AlexGalhardo/hexascore-app/commit/b22aea99ded11296feb271d01865ff2ac332c238))
- **ui:** Add hover and backdrop colour tokens ([c0d0789](https://github.com/AlexGalhardo/hexascore-app/commit/c0d07893d80d2d75ae0f4118b1fbdd34ea329bac))
- **ui:** Item page in design v1 with the half-star rating radio group ([07f9d8f](https://github.com/AlexGalhardo/hexascore-app/commit/07f9d8fe77febc3e7f40f63b381fd00d081a1bce))
- **ui:** Home in design v1 with per-type trending cards and more titles ([b44427f](https://github.com/AlexGalhardo/hexascore-app/commit/b44427fc886f13e4313f59dff9a137ea9353b487))

### Changed

- **ui:** Draw rating stars from one shared svg symbol ([c6e368f](https://github.com/AlexGalhardo/hexascore-app/commit/c6e368f64dcbc321729a68f05cc98659664f47a1))

### Fixed

- **mobile:** Read the plural item type the app API returns ([661d7dd](https://github.com/AlexGalhardo/hexascore-app/commit/661d7dddcc6617a5e9ffc091a4a3f3722f8610d6))
- **ui:** Keep the navbar log in link on one line on phones ([c95ff02](https://github.com/AlexGalhardo/hexascore-app/commit/c95ff027d621f3ec394f174e9bf2ecfd338fffdb))
- **ui:** Match the palette rows, key hints and backdrop to the v1 frames ([5e0f2ac](https://github.com/AlexGalhardo/hexascore-app/commit/5e0f2acfe5d21cc312bb237c0470ee2b62202ccd))
- **ui:** Palette rows use the sans title with the year at the right ([2086e50](https://github.com/AlexGalhardo/hexascore-app/commit/2086e50e5da721d0ae814bfb8dc1443807ce2b84))
- **ui:** Fill the open filters toggle and drop the outline on the focused page title ([fb38b29](https://github.com/AlexGalhardo/hexascore-app/commit/fb38b2928be32501ced0ba4c3464c4cdb7834ad7))
- **ui:** Item page actions in two columns on phones and quieter time to beat hint ([cd9fba0](https://github.com/AlexGalhardo/hexascore-app/commit/cd9fba0c5196c0125f2cfbab5e6d6ef469d60ee2))

## [0.25.0] - 2026-10-31

### Added

- **mobile:** Persist read queries on expo sqlite ([0eb7974](https://github.com/AlexGalhardo/hexascore-app/commit/0eb7974e059cc930e6ee28604c1e5b10f3319b86))
- **mobile:** Add the offline mutation queue ([73b2928](https://github.com/AlexGalhardo/hexascore-app/commit/73b2928dfae45f015b0e0b13c9ef02d282984f51))
- **mobile:** Add the offline banner and opt-in crash reports ([4e0d264](https://github.com/AlexGalhardo/hexascore-app/commit/4e0d264f12d3983d14692df0c82e4adfe7f75b51))
- **mobile:** Add the banned and deletion-scheduled screens ([1b95f81](https://github.com/AlexGalhardo/hexascore-app/commit/1b95f814dc625e5cedad1cf416fdde99beb09711))
- **ui:** Add v1 mobile app header and tab bar theme ([8f52747](https://github.com/AlexGalhardo/hexascore-app/commit/8f52747762de7d2a2a18e79974275f862681c27e))
- **ui:** Add v1 mobile search sheet screen component ([313c624](https://github.com/AlexGalhardo/hexascore-app/commit/313c624c866f180e63d2d526776dfd6ce983f57d))
- **ui:** Add v1 mobile home sections screen component ([9c1da22](https://github.com/AlexGalhardo/hexascore-app/commit/9c1da222f3e3100fca66e344c131f62f4caf5624))
- **admin:** Add moderation metrics page ([c28e5cf](https://github.com/AlexGalhardo/hexascore-app/commit/c28e5cff9c7605007e6fc01b54e014b862f6fa11))
- **community:** Emit moderation metrics and run nightly asserts ([f85ffb2](https://github.com/AlexGalhardo/hexascore-app/commit/f85ffb240c176a7f24a055df6f1a0f1a24db2c0e))
- **community:** Create entities and editions from the suggest form ([523e6f4](https://github.com/AlexGalhardo/hexascore-app/commit/523e6f43121a44a8e30dec0993ee41f760d06246))
- **account:** Add the community email opt-out to the profile ([82d447a](https://github.com/AlexGalhardo/hexascore-app/commit/82d447abcce2124576511d9f97f623fc8e0d8f4f))

### Changed

- **BREAKING:** **infra:** Drop app metrics and Grafana artifacts, Railway native observability only ([4c1138f](https://github.com/AlexGalhardo/hexascore-app/commit/4c1138f68833e4f25b4013a7d5e9bc5b87fc75d6))

### Fixed

- **mobile:** Render nothing while the session restores ([e6da301](https://github.com/AlexGalhardo/hexascore-app/commit/e6da301a51b10cc6bbbbe897ed25e380be7e98a2))
- **ui:** Give the mobile search input an accessible name ([e0bc804](https://github.com/AlexGalhardo/hexascore-app/commit/e0bc80475390899e4dcded2eb864cb68fbad090d))
- **ui:** Keep the home rank label at AA contrast on the card surface ([c9c2cfc](https://github.com/AlexGalhardo/hexascore-app/commit/c9c2cfcf7ea2ec61bd7af1ca38a271ae24f71214))
- **test:** Use strconv for the for-you user id lint ([97e9b6b](https://github.com/AlexGalhardo/hexascore-app/commit/97e9b6bb0e0d27bb6b21f9e34cc0ebf4445a6fc6))
- **community:** Name the item target game, movie, series or book ([d7b457a](https://github.com/AlexGalhardo/hexascore-app/commit/d7b457ac852db49d04f475f03c73b9a7dc485938))
- **infra:** Keep the predeploy lock around the invalid-index drop and migrate ([69819e5](https://github.com/AlexGalhardo/hexascore-app/commit/69819e5a580b3357886e1aa444c7423cafc1c641))
- **api:** Serve only the API on the API host ([110a6e4](https://github.com/AlexGalhardo/hexascore-app/commit/110a6e4bc174a549ab4b4147da1dab91cd48566b))

## [0.24.0] - 2026-10-31

### Added

- **recommend:** Add offline evaluation protocol and monthly CF trigger check ([f490c40](https://github.com/AlexGalhardo/hexascore-app/commit/f490c40d49dd3c87ddce22570836184654d8a93f))
- **home:** Show the 7-day trending, 3 per type ([aa2ebcb](https://github.com/AlexGalhardo/hexascore-app/commit/aa2ebcb06a525784acb10cbcbefb92fa2870ea7f))
- **ingest:** Add ingestion tables, canonical hash and record store ([f3d005f](https://github.com/AlexGalhardo/hexascore-app/commit/f3d005f934255d04da6a71ba505324d1a7e4ead1))
- **ui:** Adopt the v1 editorial palette in tokens and mobile theme toggle ([866a286](https://github.com/AlexGalhardo/hexascore-app/commit/866a286b44c7d45fb1240e36e7d48bb16096ef84))
- **ui:** Add v1 mobile primitives with half-star rating input ([22e7991](https://github.com/AlexGalhardo/hexascore-app/commit/22e7991203dd9fde768dbe13cd7070490e02c5ca))
- **ui:** Embed v1 fonts and icon glyphs in the mobile app ([c3993b8](https://github.com/AlexGalhardo/hexascore-app/commit/c3993b8f6c7b0fa956a1abf2ab49ebf1c95103ff))

### Fixed

- **http:** Redirect non-canonical hosts to APP_URL with 301 ([0bc5e13](https://github.com/AlexGalhardo/hexascore-app/commit/0bc5e133863b041fd55a2b61ecc48ddfaa338b20))
- **http:** Satisfy lint on canonical host ([3e1b2a9](https://github.com/AlexGalhardo/hexascore-app/commit/3e1b2a93c48ba90cdb94ea53aab1c8d58879f479))
- **api:** Document the api host form (server /v1, paths without /api/v1) ([ec1c379](https://github.com/AlexGalhardo/hexascore-app/commit/ec1c379a515951b4c73f79a4a742b41dbadd0ffb))
- **ui:** Align the half-star overlay with its outline star ([be0021f](https://github.com/AlexGalhardo/hexascore-app/commit/be0021f78a58aba867feace5720cd9c397f65574))
- **mobile:** Wire the root layout to the session, rotation and update screen ([e7981f9](https://github.com/AlexGalhardo/hexascore-app/commit/e7981f9c88a17e55b1bff5e0e558679fee390575))

## [0.23.0] - 2026-10-31

### Added

- **ui:** Render home trending sections from the hourly snapshot ([2a7c07f](https://github.com/AlexGalhardo/hexascore-app/commit/2a7c07fb88fb82b5941731a6426458ea7122f5a6))
- **admin:** Revert catalogue audit rows single, cascade or force ([a6143a5](https://github.com/AlexGalhardo/hexascore-app/commit/a6143a5b99ecdff2763b562d7b022bf694593e63))
- **admin:** Write catalog_audit rows for admin catalogue field edits ([9ceeead](https://github.com/AlexGalhardo/hexascore-app/commit/9ceeead05201acecc3880dbfad033d9d9b081289))
- **admin:** Add catalogue locks pages and catalog_audit csv/ndjson export ([4da1203](https://github.com/AlexGalhardo/hexascore-app/commit/4da1203c3666e38eb1a2e60100a275e99cd9241e))
- **catalog:** Add public recent changes tab and profile contribution counts ([fef5fa2](https://github.com/AlexGalhardo/hexascore-app/commit/fef5fa2fc7651fc8c079fb3c2553675f841a48fb))
- **admin:** Apply entity and edition create suggestions ([db99a95](https://github.com/AlexGalhardo/hexascore-app/commit/db99a95c4e64324f76869e64518308b952b2222c))
- **community:** Notify authors of suggestion decisions in-app and by mail ([d1bc1f2](https://github.com/AlexGalhardo/hexascore-app/commit/d1bc1f2dca1dc138b82b6c969c0e2fde3c70861c))
- **community:** Withdraw and expire suggestions on ban, deletion and merge ([71a35ba](https://github.com/AlexGalhardo/hexascore-app/commit/71a35ba227cdcf9092bb70edb68b8beaee5828a0))
- **recommend:** Add item_similar tables and similarity primitives ([09d30e4](https://github.com/AlexGalhardo/hexascore-app/commit/09d30e4498b55a252df9ef7e2407dd90238beb90))
- **recommend:** Add ComputeSimilar job and nightly schedule ([0c72aaf](https://github.com/AlexGalhardo/hexascore-app/commit/0c72aaf5dbb5fdae8ed9ee2c787ddb3f6d1bea97))
- **recommend:** Add For you profile, scoring and re-ranker ([8f2440d](https://github.com/AlexGalhardo/hexascore-app/commit/8f2440d8b3801bdad958f85014ad510bba214563))
- **recommend:** Serve For you with stale-while-refresh, opt-out and page ([e855666](https://github.com/AlexGalhardo/hexascore-app/commit/e855666c7fcb101c3ec86ff6bb5b9c00845e40db))

### Changed

- **db:** Index the event-time columns read by the trending rollup ([bb7d802](https://github.com/AlexGalhardo/hexascore-app/commit/bb7d802667f997eb917afae43eb4246dc40c4493))
- **recommend:** Batch ComputeSimilar queries and skip candidate work for unchanged items ([485ce44](https://github.com/AlexGalhardo/hexascore-app/commit/485ce443f65ec15bb1f4fcb934582c1314ffd799))

### Fixed

- **ui:** Render no trending section when the home prop is absent ([4697c84](https://github.com/AlexGalhardo/hexascore-app/commit/4697c84a0f36bf07518c736db1aceb7e0ba81b36))
- Unblock the full gate on main ([ee50a90](https://github.com/AlexGalhardo/hexascore-app/commit/ee50a90d030eb0d92f7dc162776056d690fc72a9))
- **catalog:** Renumber create_similar_tables migration to 20261031203200 ([a115acc](https://github.com/AlexGalhardo/hexascore-app/commit/a115accce64d25b845d717dbc99c9789fd880bb9))

## [0.22.0] - 2026-10-31

### Added

- **trending:** Roll up durable activity columns for today and yesterday ([be30658](https://github.com/AlexGalhardo/hexascore-app/commit/be306588f87e540bddad6ef3e11dbb5522ec243b))
- **trending:** Hourly snapshot job, prune, ban rebuild, metrics and alert A14 ([aa09436](https://github.com/AlexGalhardo/hexascore-app/commit/aa09436952c421a4434e07c1bd93034226912b16))
- **api:** Serve listTrending and business-only getItemActivity ([bc6cf9a](https://github.com/AlexGalhardo/hexascore-app/commit/bc6cf9ac7b41886cb22045233e771606bb03f47e))
- **admin:** Apply approved community edits to the catalogue ([cc64a26](https://github.com/AlexGalhardo/hexascore-app/commit/cc64a2688504e47964d380eef5e9fa12e5aa0746))
- **catalog:** Run catalog.changed side effects once per audit id ([78f4dab](https://github.com/AlexGalhardo/hexascore-app/commit/78f4dab721d1704eed4d2a677074d16646ae671c))
- **admin:** Apply create and merge community suggestions ([464a373](https://github.com/AlexGalhardo/hexascore-app/commit/464a373fbd9f00d2c442e1cf6789dd2a8af206f7))
- **admin:** Split moderation into moderator review and admin apply ([ff7872a](https://github.com/AlexGalhardo/hexascore-app/commit/ff7872aa673f73ebd76723e9d41642be1e8f233a))

### Fixed

- **api:** Report activity lookup DB errors as 500, not 404 ([df6802a](https://github.com/AlexGalhardo/hexascore-app/commit/df6802aab5a56be3cc5a0ccb94da4028b6870621))
- **community:** Answer 409 when a create races another item on the same Q-id ([0718032](https://github.com/AlexGalhardo/hexascore-app/commit/07180327eb490c9da89196c7103759af546365ff))
- **api:** Serve community-sourced fields on the public API by default ([eb45c7b](https://github.com/AlexGalhardo/hexascore-app/commit/eb45c7b1ec5c2e611d62793e5f243f8e345d554d))

## [0.21.0] - 2026-10-31

### Added

- **mobile:** Add SecureStore vault and approved runtime dependencies ([19698f2](https://github.com/AlexGalhardo/hexascore-app/commit/19698f27fea290eaa8070eee80327648b0ba1c24))
- **admin:** Moderation abilities, trusted-moderators flag and page shell ([1cc1d7f](https://github.com/AlexGalhardo/hexascore-app/commit/1cc1d7ff9a062e70211859c6a5704ef428be93c6))
- **mobile:** Add login screen with 2FA and challenge branches ([6cfcb20](https://github.com/AlexGalhardo/hexascore-app/commit/6cfcb20880ef6452eda22ea7a0a264b268328603))
- **mobile:** Add Google sign-in flow with cancel and 2FA branches ([fdaf2cf](https://github.com/AlexGalhardo/hexascore-app/commit/fdaf2cfecec937623b4a6ea135aaf555da051c87))
- **mobile:** Rotate token on foreground, logout, update-required screen ([e81f1e5](https://github.com/AlexGalhardo/hexascore-app/commit/e81f1e548e1d0728d04401dd3c23a4a380505d86))
- **trending:** Count item views and search clicks in memory and flush every minute ([41d733f](https://github.com/AlexGalhardo/hexascore-app/commit/41d733f84da38a7f76a9439ba8bdeb49fdcddb59))
- **admin:** Moderation queue view with filters, badges, claim and release ([82ff856](https://github.com/AlexGalhardo/hexascore-app/commit/82ff856a9b96a93276d66e88eb4791fe68d660f8))
- **admin:** Moderation detail view, reject, clone and bulk actions ([9858853](https://github.com/AlexGalhardo/hexascore-app/commit/985885327c426b87f7764fc35fd6f1f8359fa5dd))

### Changed

- **trending:** Resolve flushed refs with index lookups instead of id::text ([8689410](https://github.com/AlexGalhardo/hexascore-app/commit/868941008a769a977b2592f37cf3ebb017edbbf3))

### Fixed

- **mobile:** Refresh the catalogue copies after the web string changes ([81da968](https://github.com/AlexGalhardo/hexascore-app/commit/81da96869b8a814cddebd43b02473902e04bc37f))
- **migrations:** Give the moderator role migration its own timestamp ([f7a5551](https://github.com/AlexGalhardo/hexascore-app/commit/f7a5551f44f9a59304b2a1021d55b3d1b40cf974))
- **community:** Keep the visitor locale on in-page links; e2e specs use the locale prefix ([919717c](https://github.com/AlexGalhardo/hexascore-app/commit/919717c6815a32cc88ab8cad16d1fc01baf2c6ef))
- **community:** Prefix form post targets with the visitor locale ([587c46b](https://github.com/AlexGalhardo/hexascore-app/commit/587c46bb29ca51a52fc6700f815a60502baf9e20))
- **search:** Keep /search under the timeout at production catalogue size ([562734a](https://github.com/AlexGalhardo/hexascore-app/commit/562734a3283896b9d0be51d7ce0b751a4ce742a5))
- **mobile:** Skip a foreground rotation while one is in flight ([e1b9dcb](https://github.com/AlexGalhardo/hexascore-app/commit/e1b9dcb0229e6ad53d017199752adcf7cad43c7c))
- **admin:** Require 2FA for moderators on /admin ([ee4828a](https://github.com/AlexGalhardo/hexascore-app/commit/ee4828a703a5af36dab5ad14eb60be94d90e5a97))
- **admin:** Reject non-uuid moderation ids before SQL and look up bulk versions by uuid ([b638307](https://github.com/AlexGalhardo/hexascore-app/commit/b6383072d8d7bac2fae9306abe3ea327dbc4e386))

## [0.20.0] - 2026-10-31

### Added

- **lists:** Add items from the list edit page with a suggest picker ([cd6bb23](https://github.com/AlexGalhardo/hexascore-app/commit/cd6bb23c23b4b7539cb0a70e1430df1fe53d2503))
- **ui:** Add community create and merge request page ([d44c617](https://github.com/AlexGalhardo/hexascore-app/commit/d44c617018a22a5a6fe30139800608c2dab1314d))
- **ui:** Add community suggestion status, my suggestions, submit and withdraw ([1d4854e](https://github.com/AlexGalhardo/hexascore-app/commit/1d4854e9c30dcc5a5cac4c698311f7af4c06d038))
- **community:** Edit release date and relation fields in SuggestEdit ([6a0cf8d](https://github.com/AlexGalhardo/hexascore-app/commit/6a0cf8d04cb6680e7d140287644546de0da14f98))
- **trending:** Add activity and snapshot tables and sum activity on item merge ([a4a0b18](https://github.com/AlexGalhardo/hexascore-app/commit/a4a0b18e89ef31a29ea4ec0e4dc55d05a376467d))
- **mobile:** Add token store and session state machine ([fbe8b70](https://github.com/AlexGalhardo/hexascore-app/commit/fbe8b70ea5308ad12582c419afd65c92942375d2))

### Changed

- **ui:** Cache anonymous catalogue html (render cache) ([0b128bb](https://github.com/AlexGalhardo/hexascore-app/commit/0b128bbb88d5eb2bf8ff77e5db15994aa5b775f2))
- **community:** Look up change names by uuid so the primary key index is used ([a2a35e6](https://github.com/AlexGalhardo/hexascore-app/commit/a2a35e65068a0839dace17c09540fd77dc8981cf))

### Fixed

- **ui:** Key the render cache by Accept-Language on unprefixed pages ([8f20b63](https://github.com/AlexGalhardo/hexascore-app/commit/8f20b638fcf40ff7bee742ae895a0006c8182c1f))
- **api:** Add community problem codes to the catalogue ([fef9d03](https://github.com/AlexGalhardo/hexascore-app/commit/fef9d03ed18784384ebae9e73f54339d2d2cc0db))
- **search:** Degrade gracefully when the search backend times out ([2e0aaf8](https://github.com/AlexGalhardo/hexascore-app/commit/2e0aaf85ca86f1dd70250d1430af981b5d0bc03e))
- **community:** Translate release date precision and relation roles in SuggestEdit ([65a4347](https://github.com/AlexGalhardo/hexascore-app/commit/65a4347fac9d942c31d1ac7eb54fbbbe4e126cb9))
- **auth:** Render an HTML 428 page for no-JS form posts past the PoW threshold ([268802e](https://github.com/AlexGalhardo/hexascore-app/commit/268802ee7e8cac31a5673a97df345730a5490c04))
- **community:** Search pickers for merge, create entry links, names in relation diff ([a20d68c](https://github.com/AlexGalhardo/hexascore-app/commit/a20d68c6ab9ba0f7aedd55acccc51d4b74969a85))
- **mobile:** Preload react-native components outside the test budget ([6fb52de](https://github.com/AlexGalhardo/hexascore-app/commit/6fb52de30906ac94433abfeafecde783354ed1f7))

## [0.19.0] - 2026-10-31

### Added

- **catalog:** Serve covers from the bucket at /covers/{key} ([b992a24](https://github.com/AlexGalhardo/hexascore-app/commit/b992a24cedb0ba57cbd715b67b7801c251c78375))
- **ui:** Add SuggestEdit page, entry points and community strings ([c65821e](https://github.com/AlexGalhardo/hexascore-app/commit/c65821eed6ed5ddbacdc5f78a1ade344effc66ec))
- **mobile:** Wrap the generated app client in src/api ([50a2cd8](https://github.com/AlexGalhardo/hexascore-app/commit/50a2cd85f079180bb02cd9085c299a0ef9a0ae22))
- **ui:** Shelf cards for library, list and profile pages ([f15c936](https://github.com/AlexGalhardo/hexascore-app/commit/f15c9365896502150e0b9cc294114cedf9d7e1fc))

### Fixed

- **ui:** Drop inline style attributes blocked by the CSP ([69d48b4](https://github.com/AlexGalhardo/hexascore-app/commit/69d48b4d82a719da40ddc81f169d6f240d7ad313))
- **infra:** Align the QA passkey origin with the stack port and a non-IP host ([b41d775](https://github.com/AlexGalhardo/hexascore-app/commit/b41d775eee40ac3b3483fa9aeb9efb5b8dd8e821))
- **api:** Regenerate the app client after the item type enum ([0604a93](https://github.com/AlexGalhardo/hexascore-app/commit/0604a935b6ef23208b21630cc4309f788f60a29d))
- **auth:** Post named fields and form token from the no-JS login and signup forms ([3e09a52](https://github.com/AlexGalhardo/hexascore-app/commit/3e09a524ffd0ab479afad386c253011bbe44a510))
- **i18n:** Recopy mobile catalogues after account.privacy.save ([33c4717](https://github.com/AlexGalhardo/hexascore-app/commit/33c47170d6e94ea3087cb015bfdda7749f453be1))
- **mobile:** Emit tab-indented catalogue copies ([5ce673d](https://github.com/AlexGalhardo/hexascore-app/commit/5ce673d5dda9cfb0834eaafbd51d9b2517fd12c1))
- **mobile:** Spell the catalogue indent as an escape ([2b58b80](https://github.com/AlexGalhardo/hexascore-app/commit/2b58b800bcd6625f4accc5db6a5c391dc9052d4a))

## [0.18.0] - 2026-10-31

### Added

- **ui:** Rebuild the item page layout with cover, score badge and SVG stars ([1119ad9](https://github.com/AlexGalhardo/hexascore-app/commit/1119ad9815e08a3fb34b6aed422f6c13cabb7bbf))
- **ui:** Show a pending state on auth submit buttons until hydration ([9d62e18](https://github.com/AlexGalhardo/hexascore-app/commit/9d62e18f4c70d246f979039a856ec434f4fcb139))
- **api:** Add daily ID exports job ([7d083b1](https://github.com/AlexGalhardo/hexascore-app/commit/7d083b1b9ad1de8cf67ea561ef78de3f5fa6fdb9))
- **ui:** Ship a title and busy skeleton in the HTML of CSR pages ([780bf1f](https://github.com/AlexGalhardo/hexascore-app/commit/780bf1f3b87976afff65cbbce73f71434da2a398))
- **community:** Add draft save service ([c46e448](https://github.com/AlexGalhardo/hexascore-app/commit/c46e448b7f09c6de3225a405033f95462c312fd7))
- **community:** Add suggest page and draft save routes ([47e1dd9](https://github.com/AlexGalhardo/hexascore-app/commit/47e1dd96de140c943356e985576e4223ecd33abb))

### Changed

- **catalog:** Serve cover base through a shared prop instead of package state ([2f8cbd7](https://github.com/AlexGalhardo/hexascore-app/commit/2f8cbd7b7c136d0bce216e76bf27f15133e6ff9b))
- **community:** Convert field change instead of copying fields ([b0441db](https://github.com/AlexGalhardo/hexascore-app/commit/b0441db62f0216061a3ae4ead7bc1bd389efca9e))

### Fixed

- **infra:** Retry legacy cover uploads the S3 client could not rewind ([18c5a12](https://github.com/AlexGalhardo/hexascore-app/commit/18c5a1259c2b4d05efd9b4ab770012c3f9de4606))
- **i18n:** Recopy mobile catalogues after account.privacy.save ([aa4434b](https://github.com/AlexGalhardo/hexascore-app/commit/aa4434b22ede53c3116998b389693c007c10600c))
- **i18n:** Recopy mobile catalogues with the passkey keys ([891d388](https://github.com/AlexGalhardo/hexascore-app/commit/891d3882636b81d22b1298f5ba9a8ec5bf20016a))
- **catalog:** Keep cover URLs absolute in JSON handlers ([ef6b520](https://github.com/AlexGalhardo/hexascore-app/commit/ef6b520cc618be968f8413acb748ebc3414bbe29))
- **catalog:** Keep cover URLs absolute in library and profile JSON ([e80bb55](https://github.com/AlexGalhardo/hexascore-app/commit/e80bb559b637fdf5e952dc08f8f02f3d264d5e2b))
- **api:** Regenerate the Python SDK from the current OpenAPI contract ([8acb1fb](https://github.com/AlexGalhardo/hexascore-app/commit/8acb1fbfba8840752fc63050aa1aec809e08886b))
- **community:** Answer a malformed draft id as not found ([20f6eb6](https://github.com/AlexGalhardo/hexascore-app/commit/20f6eb6939ce3b662136451e85a8a3dca6c1d1dc))
- **e2e:** Wait for hydration after onboarding and reloads, match the login button exactly ([a3f092c](https://github.com/AlexGalhardo/hexascore-app/commit/a3f092c71c9306636af8f94f998e88dc23a155b2))
- **community:** Take a draft update's target from the stored row ([7bca752](https://github.com/AlexGalhardo/hexascore-app/commit/7bca752df7ca111e543b2db897b6139e271eb183))

## [0.17.0] - 2026-10-30

### Added

- **ui:** Report dialog, dashboard links and the HexaScore name fallback ([02a3a83](https://github.com/AlexGalhardo/hexascore-app/commit/02a3a83f09b4eae8ccf900c130ccc5952540828b))
- **catalog:** Serve item covers in listings, search, library, lists and profiles ([aecc0c9](https://github.com/AlexGalhardo/hexascore-app/commit/aecc0c9718c25dd27318440204588077ac35875a))
- **api:** Document the type enum and allow a free-plan burst of 5 ([1dd7f78](https://github.com/AlexGalhardo/hexascore-app/commit/1dd7f78017dae84b234fdcec4ede7786293f233f))
- **ux:** Flash a status toast after rating, status, time-to-beat, list and edition changes ([8dfaec7](https://github.com/AlexGalhardo/hexascore-app/commit/8dfaec7890148f45915f853088de34d61c114da5))
- Opt in to legacy users with no login ([783ca18](https://github.com/AlexGalhardo/hexascore-app/commit/783ca182cd5c8fc72727ac47b454092dc9741bd3))
- Register and sign in with passkeys ([fcd61f9](https://github.com/AlexGalhardo/hexascore-app/commit/fcd61f90d58b2a61d02cf07f4b1d85ae7a78ce06))
- Add passkey profile and login UI ([13fa7b6](https://github.com/AlexGalhardo/hexascore-app/commit/13fa7b69237df5fdf9ff5675ddd28a295cc0e792))
- Import legacy passkeys ([e7f6209](https://github.com/AlexGalhardo/hexascore-app/commit/e7f6209cb84c57843e7b1e0f48d97d543044591f))
- **community:** Guard community covers and source URLs ([7c627bb](https://github.com/AlexGalhardo/hexascore-app/commit/7c627bb7ba39480cbb689f6dfdfa10f1ec4ceb22))
- **ui:** Add home hero and recently rated strip ([3509c9e](https://github.com/AlexGalhardo/hexascore-app/commit/3509c9e748c62143c40310cccfb306256a3d497a))

### Changed

- **community:** Inline honeypot migration sql and drop a dead branch ([0a3dee0](https://github.com/AlexGalhardo/hexascore-app/commit/0a3dee0f850c94d8fcc66b33310b38bfb93ca3ef))

### Fixed

- **ui:** Compact mobile header, labelled star rows and a cookie banner spacer ([1ba2e26](https://github.com/AlexGalhardo/hexascore-app/commit/1ba2e26c59aa57d0767d5ffed01eaf3d6c14686c))
- **admin:** Show the full panel navigation on the ingestion and item pages ([9beef54](https://github.com/AlexGalhardo/hexascore-app/commit/9beef54b79b03b4c506de1672937d96e0e07c82a))
- **repo:** Stop gate timeouts under load ([a757f5f](https://github.com/AlexGalhardo/hexascore-app/commit/a757f5f6740892c659d514f4b2a1f06afd3cd599))
- **accounts:** Purge lists, notifications and reports on account deletion ([1727eec](https://github.com/AlexGalhardo/hexascore-app/commit/1727eecaf18ab0337939be28d86f639537129469))
- Skip legacy runtimes outside the catalog check ([8ec8bb0](https://github.com/AlexGalhardo/hexascore-app/commit/8ec8bb075c2db4385bbc7925bcfdccaad8e1a48c))
- Retry transient legacy cover copies ([e51c3e3](https://github.com/AlexGalhardo/hexascore-app/commit/e51c3e3f6c3a8b961740b2387aef1997ba58a31e))

## [0.16.0] - 2026-10-31

### Added

- Record public catalogue changes inside the cause transaction ([9e2f538](https://github.com/AlexGalhardo/hexascore-app/commit/9e2f5380ca00173570fe5fe38bfe0156f9563a54))
- Replay post-cutover writes into the legacy shape ([9c28b80](https://github.com/AlexGalhardo/hexascore-app/commit/9c28b80a8e42f8de40380846a4d54e35105d1222))
- Serve the 14-day API changes feed ([65cdf20](https://github.com/AlexGalhardo/hexascore-app/commit/65cdf201a45468349e96c5ca195c75605600cd00))
- Show public ratings, trailers and the cookie banner ([d1ec087](https://github.com/AlexGalhardo/hexascore-app/commit/d1ec0876f5d82249a6be8028bdf71604c3780753))
- Build HexaScore on vukem v0.34.0 ([4cd0bdf](https://github.com/AlexGalhardo/hexascore-app/commit/4cd0bdfbacf963af1bc44124af661690f56e0f10))
- Add the community suggestion tables ([bff78ef](https://github.com/AlexGalhardo/hexascore-app/commit/bff78ef03fac48e1f2c806c7e9940b6c721df1d8))
- Validate community suggestion fields ([a46dd13](https://github.com/AlexGalhardo/hexascore-app/commit/a46dd139b7c299f55fff0448d7671120c87caad5))
- Add the community suggestion state machine ([75a7f78](https://github.com/AlexGalhardo/hexascore-app/commit/75a7f788dfcdf64b963c35d0bf1ce0da1b293693))
- Add webhook endpoint management ([da478e2](https://github.com/AlexGalhardo/hexascore-app/commit/da478e2bd4c53c9a0b445bb93842d885b1bf7fc0))
- Scaffold the Android Expo app ([57c0c8b](https://github.com/AlexGalhardo/hexascore-app/commit/57c0c8b72356f00443310ee52c7578a9034d0605))
- Add the SDD 10 Expo Router tree ([39de6c7](https://github.com/AlexGalhardo/hexascore-app/commit/39de6c77316937199380f981965159024466b210))
- Map shared tokens and catalogues into the Android app ([0d9439c](https://github.com/AlexGalhardo/hexascore-app/commit/0d9439c35d6515c56586de19d80b28bbaab3bd19))
- Fan out item changes to webhooks ([17eb58a](https://github.com/AlexGalhardo/hexascore-app/commit/17eb58a2418f091fbdc0d4dd32ac46a38f085c4e))
- Test webhooks, prune deliveries, and count outcomes ([24e34eb](https://github.com/AlexGalhardo/hexascore-app/commit/24e34ebba965343b5c15143f16ef67baa6a53bf0))
- Reject API sunsets shorter than 183 days ([d4dbbf1](https://github.com/AlexGalhardo/hexascore-app/commit/d4dbbf11eec67d15fc0648ea5fe10b4e252acdcf))

### Fixed

- Give each Schemathesis call its own client IP ([cf384ff](https://github.com/AlexGalhardo/hexascore-app/commit/cf384ff6a57e95a108cadfa26f813d4df8538ef9))
- Close the deferred page script so forms hydrate ([f9f13fb](https://github.com/AlexGalhardo/hexascore-app/commit/f9f13fb9522d8bb56be7880343f2c45f291ca490))
- Return the stored webhook created_at ([5c2bd4e](https://github.com/AlexGalhardo/hexascore-app/commit/5c2bd4ee77d235e5d2953ef208b5db69ee706942))
- Read the webhook attempt that was just sent ([e189fc8](https://github.com/AlexGalhardo/hexascore-app/commit/e189fc8284597660ca2c1eeacb61625ecf78d114))
- Copy workspace manifests before the image install ([da6cd40](https://github.com/AlexGalhardo/hexascore-app/commit/da6cd400091079355d4242ca5365855adb3953be))

## [0.15.0] - 2026-10-30

### Added

- **security:** Add route stack inventory and per-route body limits ([73e18e0](https://github.com/AlexGalhardo/hexascore-app/commit/73e18e0e498323105c505d2267ff6544842e1abb))
- **security:** Complete proof-of-work rules and sign-up fill time ([b5993ea](https://github.com/AlexGalhardo/hexascore-app/commit/b5993ea21881b9b4de877688a55ff678d6956746))
- **infra:** Add legacy finalise stage and hexascore:cutover-verify ([f85c8c8](https://github.com/AlexGalhardo/hexascore-app/commit/f85c8c8e397bd82c999501df6d3b0a2e6c177f2d))
- **recommend:** Add the trending score functions ([4b541ff](https://github.com/AlexGalhardo/hexascore-app/commit/4b541ff8ef7e4b5bbae2ba5038f0fc632ba8457f))
- **metrics:** Register the platform meter catalogue ([eb0a329](https://github.com/AlexGalhardo/hexascore-app/commit/eb0a32980eca1cdf24e9841aa5450773ca91d43d))
- **http:** Send the API CSP and accept CSP reports ([0966614](https://github.com/AlexGalhardo/hexascore-app/commit/09666144fb793fb2b244a1f88f609acb91930a9c))
- **privacy:** Retain session client data and refresh the policy ([c60c5ca](https://github.com/AlexGalhardo/hexascore-app/commit/c60c5caede6db54f8158d7bce5e2d96a36da1488))
- **ops:** Add the day-20 spend projection ([3a6eced](https://github.com/AlexGalhardo/hexascore-app/commit/3a6ecedcbeed0580a60171c184352bdead2edf16))
- Prove entity, item-cursor and edition API reads ([17f9f7e](https://github.com/AlexGalhardo/hexascore-app/commit/17f9f7ea21d400b6ee61c738f70e1aa7125dea35))

### Changed

- **http:** Keep the Lighthouse LCP median under 2000ms ([0fbe2b5](https://github.com/AlexGalhardo/hexascore-app/commit/0fbe2b5750ca6a9fbe2ba912a1608e54adbcc0c4))

### Fixed

- **security:** Compress only build assets and gate the challenge limit ([2ad8898](https://github.com/AlexGalhardo/hexascore-app/commit/2ad8898a05269a1c14dd43421a023d2db44c57fc))
- **http:** Compress build assets and paint listing facets once ([1337614](https://github.com/AlexGalhardo/hexascore-app/commit/133761474c0ee143b9bf33bf0239164c86401555))

## [0.14.0] - 2026-10-29

### Added

- **infra:** Add backup:restore with deletion replay and backup:drill ([1b9f23b](https://github.com/AlexGalhardo/hexascore-app/commit/1b9f23b695228f98506286d517c02ab6a158e397))
- **catalog:** Import legacy catalog stage ([e99f2f0](https://github.com/AlexGalhardo/hexascore-app/commit/e99f2f0e9fe16a6537516966fc3f97683e38fa49))
- **catalog:** Import legacy entities and editions ([14df30a](https://github.com/AlexGalhardo/hexascore-app/commit/14df30a9debf300f9e7acb3691e726347fe13034))
- **ratings:** Import legacy social and library rows ([c5f4f0f](https://github.com/AlexGalhardo/hexascore-app/commit/c5f4f0fbe5c2ea80cd0a176c5cc81e6fd768ff79))
- **web:** Add pricing page, coverage docs and public Scalar API reference ([3a538c7](https://github.com/AlexGalhardo/hexascore-app/commit/3a538c7c7c83dd948cb5d35581cc9977e22a72a5))
- **infra:** Add the legacy cover export script ([09cb1df](https://github.com/AlexGalhardo/hexascore-app/commit/09cb1df4a7160cf6f7849e2bf44a9d7f0d34965a))
- **security:** Centralise the rate-limit table ([fa2463d](https://github.com/AlexGalhardo/hexascore-app/commit/fa2463d6082e090aacb3c70435134b24270a5421))

### Changed

- **auth:** Make the proof-of-work solver allocation free ([0133ae6](https://github.com/AlexGalhardo/hexascore-app/commit/0133ae6edad75ab4a42baca9739e614e3d318deb))
- **infra:** Name the secure headers flag after what it controls ([d360398](https://github.com/AlexGalhardo/hexascore-app/commit/d36039868c63ad42745c4cd474b06fc1d1743c94))

### Fixed

- **auth:** Count the announced attempt in the challenge probe and drop bypassCSP from e2e ([447db34](https://github.com/AlexGalhardo/hexascore-app/commit/447db34f8c658a23f1876d16f7c6c567e217a0c3))
- **search:** Qualify hs_unaccent in hs_norm so dumps restore ([5251c61](https://github.com/AlexGalhardo/hexascore-app/commit/5251c61c2b6b9f504a111928ddd281d9e7d95df5))
- **ui:** Keep the locale prefix on account form posts ([59fe39f](https://github.com/AlexGalhardo/hexascore-app/commit/59fe39fdfdda9be929ccf9398e9678d098978cc4))
- **ui:** Keep the locale prefix on item, list, contact and notification posts ([84c73d4](https://github.com/AlexGalhardo/hexascore-app/commit/84c73d4dfa4585a3c7f5aba22d68fcabc8c8dad8))

## [0.13.0] - 2026-10-29

### Added

- **infra:** Add region latency script and partial ADR 0002 measurement ([7d6391e](https://github.com/AlexGalhardo/hexascore-app/commit/7d6391ed1011cecbe4773fa089a012f78c8de8d5))
- **openapi:** Lint exported documents with Redocly ([03cb2c3](https://github.com/AlexGalhardo/hexascore-app/commit/03cb2c38a65fa99920452e03c7932fd4730a6b9b))
- **infra:** Add hexascore:predeploy with INVALID index cleanup ([f087a09](https://github.com/AlexGalhardo/hexascore-app/commit/f087a099d8ca7bf7bf813b81aaf3c0aad9a6b3cb))
- **infra:** Add privacy-preserving request log handler ([b97a401](https://github.com/AlexGalhardo/hexascore-app/commit/b97a401298979e7f16bd6d742c6cbdc5180b5040))
- **infra:** Add backup:run with framed client-side encryption ([bc09f51](https://github.com/AlexGalhardo/hexascore-app/commit/bc09f51e9199abf87adef777d460202cf9661027))
- **admin:** Ban and unban from the user form ([cdd78f9](https://github.com/AlexGalhardo/hexascore-app/commit/cdd78f9c1acaacab1fdafbd1075ee77f1173206e))
- **admin:** Require 2fa, cap admin sessions and add ip allowlist ([e1bd008](https://github.com/AlexGalhardo/hexascore-app/commit/e1bd008ee0b766301599bcf8d99ce4ae3e80c8d0))
- **infra:** Add hexascore:import-legacy runner with checkpoints and pre-flight ([4ee0157](https://github.com/AlexGalhardo/hexascore-app/commit/4ee0157f61b9930d29103a1a961a319a09f53381))
- **auth:** Import legacy users with credential and google mapping ([5bcaf8a](https://github.com/AlexGalhardo/hexascore-app/commit/5bcaf8a60651c25e16b7f31580d0e180b82e0650))
- **auth:** Import legacy 2fa secrets over stdin ([7aa3966](https://github.com/AlexGalhardo/hexascore-app/commit/7aa39666889d19252a7e2a204ce008cd0f1da76a))
- **billing:** Import legacy api keys, subscriptions and credits ([a28416d](https://github.com/AlexGalhardo/hexascore-app/commit/a28416d6cf16e5e1b800eb25a3c6d731b1a6197f))
- **infra:** Verify, copy and prune database backups daily ([b6226f6](https://github.com/AlexGalhardo/hexascore-app/commit/b6226f610d0953d02be6b90d8330d3885eca03c2))

### Changed

- **i18n:** Adopt vukem locale path prefixes ([edf2d47](https://github.com/AlexGalhardo/hexascore-app/commit/edf2d474295491c094920d35be7ffde281a776c1))
- **models:** Declare Rating composite key with PrimaryKeys ([6c5bf99](https://github.com/AlexGalhardo/hexascore-app/commit/6c5bf99ce5c221ad8ac0511b556da1b2f4ff96f2))

### Fixed

- **infra:** Make region-latency executable and count HTTP errors as failures ([334efdf](https://github.com/AlexGalhardo/hexascore-app/commit/334efdf97e2289e16ea0fa1e0e7e7bb7b0f6f418))
- **api:** Keep CORS off /app/v1 while coding its problems ([888fa81](https://github.com/AlexGalhardo/hexascore-app/commit/888fa81c954f8122e680ec096515f41080aca381))
- **ui:** Keep already prefixed redirect targets as they are ([d81d2a7](https://github.com/AlexGalhardo/hexascore-app/commit/d81d2a7b497d9a653428aa73e87e506d5f9bc31c))
- **infra:** Sample suggest and events lines under the real locale prefixes ([e0d11d6](https://github.com/AlexGalhardo/hexascore-app/commit/e0d11d64eb828b04800ee17f580c10e1d7fec4a5))

## [0.12.0] - 2026-10-29

### Added

- **billing:** Install billing with plan state machine and subscription-backed API plans ([c345c24](https://github.com/AlexGalhardo/hexascore-app/commit/c345c2423307207f640afb5c765c08660affb488))
- **api:** Commit OpenAPI goldens with drift check and regenerate SDKs ([2b89939](https://github.com/AlexGalhardo/hexascore-app/commit/2b89939834f453d493802eda9dd7bc3e3b28baff))
- **library:** Page the edition picker list ([2d47242](https://github.com/AlexGalhardo/hexascore-app/commit/2d47242e96e32be672257bf073233f7584a60846))
- **infra:** Harden serve shutdown and add the four readiness checks ([ff35da6](https://github.com/AlexGalhardo/hexascore-app/commit/ff35da6e3c007c5fe95a72a2b492f1409191945e))

### Fixed

- **billing:** Cancel every Stripe-live subscription on account deletion ([12c6cb0](https://github.com/AlexGalhardo/hexascore-app/commit/12c6cb05647d0f56329da25f3fbff03c122a6ab6))
- **billing:** Cancel superseded subscriptions at Stripe and expire open checkouts ([b1cf4ec](https://github.com/AlexGalhardo/hexascore-app/commit/b1cf4ec24eb1425f70e08d4fd260448c0746461b))
- **catalog:** Show highlight and listing scores on the 0-10 scale ([a93ce1a](https://github.com/AlexGalhardo/hexascore-app/commit/a93ce1a27fae7a4a5342384b7479012cbaeb955e))
- **catalog:** Link countries and languages by code on the item page ([3d9e52f](https://github.com/AlexGalhardo/hexascore-app/commit/3d9e52fc4cdd2575c215cdf896669ddc84ecf650))
- **ui:** Label card scores out of 10 after the score scale fix ([61986cf](https://github.com/AlexGalhardo/hexascore-app/commit/61986cf8fba83b7bd20cc97e4f5c6b2a1efedf5e))
- **ui:** Load each edition page once on repeated clicks ([1d59e2c](https://github.com/AlexGalhardo/hexascore-app/commit/1d59e2c5da7f4c1f4663743e35156358f9bfbe92))
- **catalog:** Redirect lowercase country and language codes to canonical ([10fe7d7](https://github.com/AlexGalhardo/hexascore-app/commit/10fe7d783f9e10b3432cd90a35909fa1784045d5))
- **deps:** Override js-yaml to 4.3.2 in the TypeScript SDK ([9b3a0b7](https://github.com/AlexGalhardo/hexascore-app/commit/9b3a0b7cbbba85e9aef1405dad6ca9e049ec42f0))
- **qa:** Lint and typecheck the e2e suite, stop hiding seed failures ([0a03f69](https://github.com/AlexGalhardo/hexascore-app/commit/0a03f6963e67407e1970d5f35f5ef71181668511))
- **repo:** Make parity inventory runnable from the main checkout ([0db3585](https://github.com/AlexGalhardo/hexascore-app/commit/0db3585502c94cb15949d156e7c23235cb85071e))

## [0.11.0] - 2026-10-29

### Added

- **profile:** Data export button and purge of export files on account deletion ([bb62965](https://github.com/AlexGalhardo/hexascore-app/commit/bb62965f535ddd5c8c20aea341d3e682811de47d))
- **lists:** Add library view, edition choice and public profile ([aa6e5d1](https://github.com/AlexGalhardo/hexascore-app/commit/aa6e5d16950e326213b439b4bf19cc59e0087178))
- **ui:** Add profile, library and lists pages ([4774178](https://github.com/AlexGalhardo/hexascore-app/commit/4774178eaaf94e6f8f94311426fdceb361334546))
- **api:** Serve licence-filtered item reads on /api/v1 ([e35b5ff](https://github.com/AlexGalhardo/hexascore-app/commit/e35b5ffcdcbf3edcfa77325b5f44fc1d18779919))
- **api:** Serve entity read endpoints on /api/v1 ([ae28da0](https://github.com/AlexGalhardo/hexascore-app/commit/ae28da0d87e203d2d9a669bb1ccce586255aa347))
- **api:** Serve search and lookup endpoints on /api/v1 ([3890402](https://github.com/AlexGalhardo/hexascore-app/commit/38904027c1d88e8522c0b971df04111de63fc8e8))
- **library:** Filter counts, see-all links and clear-edition action ([546395a](https://github.com/AlexGalhardo/hexascore-app/commit/546395a764bfe27a3a0802e50f71b70c570998f0))
- **api:** Host routing, CORS, coded problems and problem catalogue ([8f98425](https://github.com/AlexGalhardo/hexascore-app/commit/8f9842569e791dd7be47765c820a0a28fca00e77))
- **library:** Add edition picker and session editions route ([75f8642](https://github.com/AlexGalhardo/hexascore-app/commit/75f864206795939b7c4eead84eb9e0b33e7d2abd))
- **catalog:** Add Entity/Show pages for catalogue entities ([9d42955](https://github.com/AlexGalhardo/hexascore-app/commit/9d4295506cd07373bd9be138329c95433355ea66))

### Changed

- **catalog:** Share score fields and listing query parsing ([eb08b2b](https://github.com/AlexGalhardo/hexascore-app/commit/eb08b2be7386a0ae08d53cbfed0b475f9bfc9fe5))

### Fixed

- **api:** Render service error codes as problem+json code members ([c402980](https://github.com/AlexGalhardo/hexascore-app/commit/c4029809491f637ee0489e2d502bb86f4507998e))
- **admin:** Answer 409 on merge conflicts as SDD 01 §7.5 requires ([c63586f](https://github.com/AlexGalhardo/hexascore-app/commit/c63586f0bd1fec76beef824928e6691956fd0213))
- **admin:** Audit the previous active flag of an affiliate link ([9ea99f7](https://github.com/AlexGalhardo/hexascore-app/commit/9ea99f727a2d8f8dda4e7bd97bf6ab04d8524946))
- **api:** Send Retry-After: 30 on the key guard's fail-closed 503 ([a4254e0](https://github.com/AlexGalhardo/hexascore-app/commit/a4254e086a686af46f1d7388b3d707eb14369014))
- **api:** Fail closed with Retry-After on quota and plan errors ([2c11948](https://github.com/AlexGalhardo/hexascore-app/commit/2c11948ba0ba920613a42890d9d1de56bb05f536))
- Close export purge retention gaps ([5b4528c](https://github.com/AlexGalhardo/hexascore-app/commit/5b4528ca009880660ab0f7f442450a52391c47bd))
- **lists:** Noindex banned owners' profiles and test edition moves on merge ([f6b5ba3](https://github.com/AlexGalhardo/hexascore-app/commit/f6b5ba3e010ccaf2059fa14c5f1563e61383d912))
- **lists:** Noindex banned owners' lists and fix library and list-edit page issues ([b06938e](https://github.com/AlexGalhardo/hexascore-app/commit/b06938ee6babbfc70f32c72a8ff2369fdb292c8a))
- **api:** Resolve lookup ids to public items with indexable queries ([d068cfd](https://github.com/AlexGalhardo/hexascore-app/commit/d068cfdeb7262deee53a64deda74cfb19e5e083e))
- **api:** Keep restricted-source relations out of api filters and counts ([b99bdb2](https://github.com/AlexGalhardo/hexascore-app/commit/b99bdb23ef3462e5947985e007ede97f5ebbe0b0))
- **docs:** Regenerate docs.json for the changes-window-exceeded section ([254aea1](https://github.com/AlexGalhardo/hexascore-app/commit/254aea10164e359b81d3175257cd45c94bd62574))
- **catalog:** Filter api entity names by their own source ([7070183](https://github.com/AlexGalhardo/hexascore-app/commit/7070183bc72e48ec1612c4afbcad9bff06bc1965))

## [0.10.0] - 2026-10-29

### Added

- **auth:** Profile sections for details, password, sessions and privacy ([89f04ec](https://github.com/AlexGalhardo/hexascore-app/commit/89f04ec2bb90f1276458419e89fdb5ce45952a66))
- **notifications:** User_notifications inbox with deduped Send, read routes and unread prop ([10a6293](https://github.com/AlexGalhardo/hexascore-app/commit/10a6293968851c67f42eee9c7085c963b2eb7245))
- **catalog:** Add item and edition pages with rating, status and time to beat ([7e55c4f](https://github.com/AlexGalhardo/hexascore-app/commit/7e55c4fa6be86d2756db3f02b826d2a520b9e0a5))
- **security:** Proof-of-work challenge and signup honeypot for auth routes ([e2794a8](https://github.com/AlexGalhardo/hexascore-app/commit/e2794a805ce2ff1aafe546ee0f730c254be2cc27))
- **reports:** Add reports, contact form and admin resolve route ([2d45338](https://github.com/AlexGalhardo/hexascore-app/commit/2d453387d5096efd41e1e430abb3bcf55d2d6bd1))
- **catalog:** Add Catalog/Index listings with filters, facets and infinite scroll ([08c2bf9](https://github.com/AlexGalhardo/hexascore-app/commit/08c2bf9ce9980976c0666e697c9ec78486c40500))
- **search:** Add Search/Index page and header suggest combobox ([599fd06](https://github.com/AlexGalhardo/hexascore-app/commit/599fd06acaf0bad9440e623815a7ff49967f4fed))
- **web:** Add docs, legal pages and robots.txt ([868c9d6](https://github.com/AlexGalhardo/hexascore-app/commit/868c9d67a4bf616587c5d2cebfabf40a62558b03))
- **api:** Add hx_live_ api keys, fail-closed key guard and key management page ([1171f9f](https://github.com/AlexGalhardo/hexascore-app/commit/1171f9f4f1ffb3bdd43cdc3a867636cf279eddd6))
- **admin:** Wire items, merge, affiliate links and ingestion pages ([c7a637d](https://github.com/AlexGalhardo/hexascore-app/commit/c7a637d5212c4fa957383c2fe0fa11daacfbce02))
- **library:** Add background LGPD data export with signed 24 h link ([076fc51](https://github.com/AlexGalhardo/hexascore-app/commit/076fc515a67d29775bef59010da9eb929ad3dfea))
- **api:** Add plans, /api/v1 stack and quota metering in postgres ([a3cf95f](https://github.com/AlexGalhardo/hexascore-app/commit/a3cf95fd3f823d7545a0a5f0c87b0c34fe2bca44))

### Changed

- **search:** Add k6 search scenario and SEARCH_CACHE=off bypass ([c37a2bb](https://github.com/AlexGalhardo/hexascore-app/commit/c37a2bbc11a0f19b338d9d7fd628ec08dbca9bbc))

### Fixed

- **repo:** Keep the highlights home route after merge ([ce890a7](https://github.com/AlexGalhardo/hexascore-app/commit/ce890a7b33480a8ee350cd4b4208e62ecae1c6ab))
- **auth:** Require the current password unless the account is social-only, drop remember tokens on session revoke ([6d682dd](https://github.com/AlexGalhardo/hexascore-app/commit/6d682ddd2c8b189ebf7bffaec892d65ac5c38e32))
- **auth:** Decode PoW challenges strictly so one challenge has one replay key ([ec098dc](https://github.com/AlexGalhardo/hexascore-app/commit/ec098dc2b6a3cefbc58e3f191104e8aca606b0b0))
- **search:** Reload searchId with paginated results so page clicks are recorded ([b06a050](https://github.com/AlexGalhardo/hexascore-app/commit/b06a050e06c3a3d53f8a6f4a72d614f80ed4b28f))
- **i18n:** Drop duplicate contact block that hid the contact page strings ([f266401](https://github.com/AlexGalhardo/hexascore-app/commit/f266401674006875cbe8117521f1285442add64b))
- **repo:** Keep resources/public tracked so the all:public embed builds on a clean checkout ([1c4b18e](https://github.com/AlexGalhardo/hexascore-app/commit/1c4b18e7346d48df4ee5463aed638ce94f9f224e))
- **web:** Refuse /\host doc links that browsers resolve to another host ([9576e95](https://github.com/AlexGalhardo/hexascore-app/commit/9576e9555696fb142e19e2a70eb5dad0674bbd8c))
- **api:** Keep flushing quota deltas visible to admission ([986b804](https://github.com/AlexGalhardo/hexascore-app/commit/986b8049117c6c36c4673a545c966aaa3777a7bd))

## [0.9.0] - 2026-10-29

### Added

- **search:** Add Searcher with tiered ranking, query parsing and audiences ([62d47bd](https://github.com/AlexGalhardo/hexascore-app/commit/62d47bdd2a3a3624ce835a85c27156bddc0361b8))
- **ui:** Add app shell, SEO head and Base UI dialog ([a6a0338](https://github.com/AlexGalhardo/hexascore-app/commit/a6a033874f824829b327815c63c3f12772960510))
- **search:** Add search routes, suggest, click events and purge ([c74940c](https://github.com/AlexGalhardo/hexascore-app/commit/c74940c68a7a05d1e7ed1855bbb4a1ef683bb21c))
- **ui:** Add nonce CSP with strict-dynamic and security header set ([a5c28ac](https://github.com/AlexGalhardo/hexascore-app/commit/a5c28ac8041b47d26685c6d3159cf81eade4f733))
- **auth:** Add /app/v1 token endpoints for password and Google ID token ([aefb107](https://github.com/AlexGalhardo/hexascore-app/commit/aefb1078d3f16275ec434828312839cb60da7e05))
- **admin:** Add read-only scoring resources ([cf2a085](https://github.com/AlexGalhardo/hexascore-app/commit/cf2a0854149ea27a57cf0d379f22769cd488efa8))
- **ratings:** Add item status and time-to-beat writes and GET /me ([b03da9f](https://github.com/AlexGalhardo/hexascore-app/commit/b03da9ffc1c2a0f36846d7397663660e58237763))
- **ratings:** Schedule score recompute, burst release and event prune, add ban/merge listeners and scoring DQ ([19d8ff4](https://github.com/AlexGalhardo/hexascore-app/commit/19d8ff4ee8714078c9220042f2757834fc1a51e5))
- **lists:** Add lists tables, ListPolicy and list create, update, delete ([e165fd6](https://github.com/AlexGalhardo/hexascore-app/commit/e165fd63427963b991258bcbe37ee4f0dd64ad32))
- **auth:** Localise account mails from the user's saved locale and add password and 2FA notices ([0715311](https://github.com/AlexGalhardo/hexascore-app/commit/0715311bd1c9d18a31bcdf8cadcf5da5208f1676))
- **lists:** Add list items, versioned reorder and merge move ([08ca217](https://github.com/AlexGalhardo/hexascore-app/commit/08ca2178340094dbe7b52d16094b09e55f70dc28))
- **ui:** Add half-star rating widget and /help/score page ([1e0446e](https://github.com/AlexGalhardo/hexascore-app/commit/1e0446e6f0dfe60bbd4320a49e42802f231853ad))
- **web:** Add the home page with per-type section highlights ([74db24f](https://github.com/AlexGalhardo/hexascore-app/commit/74db24f77cac2b5e70603146ed8840e31c37e6b1))

### Fixed

- **search:** Accept clicks only on returned results, widen PII redaction, key page cache by locale prefix ([89a5f68](https://github.com/AlexGalhardo/hexascore-app/commit/89a5f688c68216dcd523c22dd8cfd23afbe9bd12))
- **search:** Redact any numeral run and rate-limit click events per IP ([b135ce0](https://github.com/AlexGalhardo/hexascore-app/commit/b135ce0a49bc15eaefd86df0ec1c3d874ff33d4c))
- **auth:** Make app token rotation single-use and cap token issue per identifier ([2fbe578](https://github.com/AlexGalhardo/hexascore-app/commit/2fbe57850f1ebcb88843f09af67600394675dd82))
- **ui:** Keep frame-src and worker-src free of nonce and strict-dynamic ([d6c0b2d](https://github.com/AlexGalhardo/hexascore-app/commit/d6c0b2db37b5d2e6824a0f08849b553a99b9e2a7))
- **auth:** Cap app second-factor attempts per user and keep abilities on rotation ([06275cb](https://github.com/AlexGalhardo/hexascore-app/commit/06275cbf0b153e03699f33732e527d1158b3727b))
- **ratings:** Keep status writes to public items and freeze the score during a burst limit ([ef4f317](https://github.com/AlexGalhardo/hexascore-app/commit/ef4f317cf9cd3e6edb7c59a97359242f44cf1a59))
- **ratings:** Auto-release a burst only after its limit expired ([9e8e209](https://github.com/AlexGalhardo/hexascore-app/commit/9e8e2091285b866e89c90469c0019b1c83e3ede4))
- **repo:** Register the lists and status routes once after merge ([2eb53ff](https://github.com/AlexGalhardo/hexascore-app/commit/2eb53fffbb21ae56f50c526f8a4746cdf434df0f))

## [0.8.0] - 2026-10-29

### Added

- **catalog:** Serve sitemap index and chunked child sitemaps ([a22f0c9](https://github.com/AlexGalhardo/hexascore-app/commit/a22f0c908316f19fa3e23f8d5ddefcc794d8fc7f))
- **auth:** Add admin panel, users resource with ban and admin CLI ([3eb3139](https://github.com/AlexGalhardo/hexascore-app/commit/3eb3139113f9b829705a85489cd88b21e4c7034b))
- **auth:** Localise Auth pages and add nickname onboarding ([f52df54](https://github.com/AlexGalhardo/hexascore-app/commit/f52df5496f37b9774371838f690b03c9ff7b4527))
- **infra:** Expose the release version on every response ([7ae411b](https://github.com/AlexGalhardo/hexascore-app/commit/7ae411b203eedb0d87c7a3aabc171c8daf336f14))
- **search:** Add search extensions, hs_norm, text search configs, tsv trigger, trigram indexes and search_dirty ([c7c40ae](https://github.com/AlexGalhardo/hexascore-app/commit/c7c40ae7903e5f8121694fb09ae117b1f2798385))
- **ratings:** Add rating write path on web and /app/v1 ([43d1359](https://github.com/AlexGalhardo/hexascore-app/commit/43d135935e1a37470e629cf2e2edff0703beb8b4))
- **search:** Add search_suggest table and incremental sync runner ([fb5b3ea](https://github.com/AlexGalhardo/hexascore-app/commit/fb5b3eadf086838a7306b95c83e10fcd8a22d091))
- **auth:** Login by email or nickname, onboarding redirects and nickname gate ([b4a0169](https://github.com/AlexGalhardo/hexascore-app/commit/b4a01690c358ad9a0f588e55afa9eb34e140193a))
- **ratings:** Add burst detector, daily caps, item limiter and moderation routes ([df1d2ac](https://github.com/AlexGalhardo/hexascore-app/commit/df1d2ace153b6c6b88ce625ad876476bf97a15a3))
- **auth:** Schedule, cancel and purge account deletion with public help page ([7c39a1c](https://github.com/AlexGalhardo/hexascore-app/commit/7c39a1c2b3f082315e1fe51b3724d7ba8bde28a7))

### Fixed

- **admin:** Register all console commands; number VF-046; tick P1.C.3.1 ([a5af6c5](https://github.com/AlexGalhardo/hexascore-app/commit/a5af6c5a32235b07dde465a485f9401929f97430))
- **auth:** Key login throttles by the resolved account ([66fb3bb](https://github.com/AlexGalhardo/hexascore-app/commit/66fb3bba3972f5959aaefee1220675892e4a7422))
- **auth:** Cap login attempts per IP across identifiers instead of resolving accounts ([cb12575](https://github.com/AlexGalhardo/hexascore-app/commit/cb1257583fadc6eef0cde86b007f4912053700de))
- **accounts:** Register every job, defer rows.Close in the purge, align deletion test with identifier login ([93d632c](https://github.com/AlexGalhardo/hexascore-app/commit/93d632cf0113f0882de8ab5e270af4796fda4cb2))

## [0.7.0] - 2026-10-29

### Added

- **ui:** Add design tokens package with contrast tests ([01200f8](https://github.com/AlexGalhardo/hexascore-app/commit/01200f8064cc8e2c3be8174b05e00e799dfe5e14))
- **catalog:** Migrate entities and relations with hourly stats refresh ([963745a](https://github.com/AlexGalhardo/hexascore-app/commit/963745abe87b71a3f5e7012863dd997fda44d6b9))
- **auth:** Install vukem auth with Google and 2FA on the legacy paths ([43219cc](https://github.com/AlexGalhardo/hexascore-app/commit/43219cc40ced1c4f8ba85c61080717dbeeb5a229))
- **catalog:** Migrate editions, isbn13 function, affiliate links and redirects ([472cdbf](https://github.com/AlexGalhardo/hexascore-app/commit/472cdbf4ae224e346d54889b4d1b032051350c3d))
- **ui:** Add locale-prefix middleware, prefixed route helper and theme attribute ([ee4c530](https://github.com/AlexGalhardo/hexascore-app/commit/ee4c530495e67b1863059fdb6113a0bcf336f633))
- **catalog:** Add the audience-filtered item read model ([e2a42e8](https://github.com/AlexGalhardo/hexascore-app/commit/e2a42e8ad9f14d2403e95e1317c3a42eb0e93738))
- **auth:** Add Better Auth compatibility hasher and password rules ([1057e04](https://github.com/AlexGalhardo/hexascore-app/commit/1057e04ddd972d530547f7fc8c7a07cf7d0c29be))
- **catalog:** Add slug generation, catalog redirects and 301 middleware ([90303b9](https://github.com/AlexGalhardo/hexascore-app/commit/90303b981e933b864aa9929b5cd10b476afbd790))
- **catalog:** Add listing queries over a signed keyset cursor ([5abd2ca](https://github.com/AlexGalhardo/hexascore-app/commit/5abd2ca98a5f6b8e26da4182e4a0e05ae5da41b2))
- **ratings:** Migrate the scoring tables and item score columns ([75d27bb](https://github.com/AlexGalhardo/hexascore-app/commit/75d27bb9f38d2af780195e725d0d8729b34f0a76))
- **catalog:** Add DemoCatalog seeder and factories ([ffc78a7](https://github.com/AlexGalhardo/hexascore-app/commit/ffc78a781815cce8adea55a509c73fecbfd04f56))
- **catalog:** Add catalog:check command with DQ-01..DQ-14 asserts ([e0bd963](https://github.com/AlexGalhardo/hexascore-app/commit/e0bd9637a468cb2474d7451a5b0bb76f2f78b4df))
- **catalog:** Add item merge service and catalog:merge command ([a962955](https://github.com/AlexGalhardo/hexascore-app/commit/a962955c68002cb23a4daa1ccebc4efd73a43efb))
- **ratings:** Implement formula v1 in Go and the equivalent SQL recompute ([53ecf47](https://github.com/AlexGalhardo/hexascore-app/commit/53ecf47e9ba2b3303fb68896640acdb03c503c02))

### Fixed

- **catalog:** Use the vukem faker instead of math/rand in the demo seeder ([0bee60f](https://github.com/AlexGalhardo/hexascore-app/commit/0bee60f96748ad1852572a685fc24b0d2b05276f))
- **catalog:** Keep user data out of merge snapshots and carry anti-abuse state ([969a2c4](https://github.com/AlexGalhardo/hexascore-app/commit/969a2c4078695b553f0bd489295725691c2ce72a))

## [0.6.0] - 2026-10-28

### Added

- Scaffold the vukem Inertia React app on Postgres ([1bd7ce3](https://github.com/AlexGalhardo/hexascore-app/commit/1bd7ce3c08c06e53c34f2e6e1763f30de4be07bf))
- **ssr:** Enable Inertia SSR with Go-supervised Node, ssr health check and fallback counter ([8235e8b](https://github.com/AlexGalhardo/hexascore-app/commit/8235e8b117bac2ccd60d5b633cffef7e910f8ed9))
- **infra:** Add local compose stack and setup scripts ([a822f03](https://github.com/AlexGalhardo/hexascore-app/commit/a822f03100267499ae1c36c730109e11e0e4484a))
- **i18n:** Port legacy catalogues to vukem and share them as a once prop ([466fb81](https://github.com/AlexGalhardo/hexascore-app/commit/466fb8130764a485e3ccd5cc79c473a5eea75178))
- **infra:** Add advisory-locked migrate command and concurrent index script ([748aee4](https://github.com/AlexGalhardo/hexascore-app/commit/748aee45c5da2edd38f44c855a97cec977b4225d))
- **catalog:** Migrate data sources, items, translations and alt titles ([1852e7b](https://github.com/AlexGalhardo/hexascore-app/commit/1852e7b191268e76468d2cb54cbf033440215013))

### Fixed

- **i18n:** Drop copy for features removed by D-06 ([24ef66b](https://github.com/AlexGalhardo/hexascore-app/commit/24ef66bd7869b76e2ba83e54c4bc7055f41d9e3c))

## [0.5.0] - 2026-10-28

### Added

- **seed:** Admin dashboard data and valid UUIDs for seed rows ([2fe7ad1](https://github.com/AlexGalhardo/hexascore-app/commit/2fe7ad1151908f9ec445616d361c47f9796ec4c5))
- **ingest:** Show only embeddable trailers, at most two per item ([df5fb23](https://github.com/AlexGalhardo/hexascore-app/commit/df5fb23e03b8fb0ec259a26b95a5965fb192f8de))
- **sdk:** Publishable TypeScript SDK for npm and JSR ([11cae35](https://github.com/AlexGalhardo/hexascore-app/commit/11cae3569adb2e336b7a7ef4f869f561e6f4594f))
- **sdk:** Publishable Python SDK with hand-written project files ([368b1b9](https://github.com/AlexGalhardo/hexascore-app/commit/368b1b9eb19a89e76be45a2bd647eab14dba084c))

## [0.4.0] - 2026-10-10

### Added

- **web:** Cover thumbnails in catalog lists and fix table overflow ([a6f6940](https://github.com/AlexGalhardo/hexascore-app/commit/a6f6940afc7ec5b1f55143e3371459fe73146af5))
- **web:** Show only language codes in the locale menu ([76bc124](https://github.com/AlexGalhardo/hexascore-app/commit/76bc12417abaaf7fae93685af2973ef78454be11))
- **lists:** Cap lists at 10 per user and item type ([be35209](https://github.com/AlexGalhardo/hexascore-app/commit/be35209f89af4d67f839c11d2a8392d854e68fbd))
- **web:** Button-like link hover, navbar icons and profile nickname ([8054981](https://github.com/AlexGalhardo/hexascore-app/commit/8054981ab18b5d559ff7f0c51f7968efb059ed6b))
- **web:** Onboarding nickname matches sign-up rules ([c2605fc](https://github.com/AlexGalhardo/hexascore-app/commit/c2605fcc2ca5b1097aea2d6ee16153554764673f))
- **docs:** Localize the Scalar API reference UI ([1362d36](https://github.com/AlexGalhardo/hexascore-app/commit/1362d363620315a5d3427dfde628800524ccb3c0))
- **api:** Translate the OpenAPI documents to pt-BR and es ([e50e9da](https://github.com/AlexGalhardo/hexascore-app/commit/e50e9da63bec96e5544ed26e09b2dd98a1eee0d3))
- **web:** Cover, title and score together in the item header ([d5b1f6a](https://github.com/AlexGalhardo/hexascore-app/commit/d5b1f6a1f7cd168658319618c6d67a9e4d22b0fa))
- **covers:** Add 96x144 WebP list thumbnails ([de1709e](https://github.com/AlexGalhardo/hexascore-app/commit/de1709e34f56adcce1bb35f670012f4e904f9c71))
- **web:** Center the users' scores dialog and paginate it by page number ([264c956](https://github.com/AlexGalhardo/hexascore-app/commit/264c956a8ba18422e67452e2a5fcbcf488551878))
- **docs:** Translate the API guides to pt-BR and es ([f0636bd](https://github.com/AlexGalhardo/hexascore-app/commit/f0636bd761c4884c205ebab92e3ea10f4e255762))
- **seed:** Sandbox mode on existing items, purge, target guard and banner ([3c4f1ad](https://github.com/AlexGalhardo/hexascore-app/commit/3c4f1ad4a5b9d914af805dcab97181e8d38ffa8a))
- **web:** Item header score on the title row and metadata chips ([eacfc4c](https://github.com/AlexGalhardo/hexascore-app/commit/eacfc4c1886ed307b98c711de894f9d4f2676ea6))
- **web:** Catalog filters always visible, equal widths and result count ([2e560e6](https://github.com/AlexGalhardo/hexascore-app/commit/2e560e6ec36d535d16d91b4fa133ac7acb87ab18))
- **seed:** Default sandbox admin and admin docs ([fa1023e](https://github.com/AlexGalhardo/hexascore-app/commit/fa1023e4c5c3f27771058b7a81211e089141f270))
- **web:** Category-split listings with cover thumbnails and category icons ([fad358a](https://github.com/AlexGalhardo/hexascore-app/commit/fad358aee6bddb11849c20eec50548394ca933d5))
- **seed:** Allow a guarded production target for demo data ([d3cd6fe](https://github.com/AlexGalhardo/hexascore-app/commit/d3cd6fec7d2f7f47487059cddca6fc99639da003))
- **ingest:** Official covers and missing fields from TMDB and IGDB ([bbff3f3](https://github.com/AlexGalhardo/hexascore-app/commit/bbff3f33d54f736c9340589bcc053e1cb5a4fcdb))
- **web:** Show the item cover whole in its own proportion ([acc75cf](https://github.com/AlexGalhardo/hexascore-app/commit/acc75cf760dbfada361ddf6107de666c1203fdc4))
- **seo:** Serve the sitemap index at /sitemap.xml ([b6eb8ac](https://github.com/AlexGalhardo/hexascore-app/commit/b6eb8ace7cdf5a57f1370f6a4ea50e5ba3d6d611))

### Changed

- **ingest:** Fetch wikipedia summaries concurrently ([4d280b7](https://github.com/AlexGalhardo/hexascore-app/commit/4d280b772431894b0ada3eccd1556ea47c8eee85))
- **ingest:** Process covers concurrently with batched updates ([76190d8](https://github.com/AlexGalhardo/hexascore-app/commit/76190d8013be55175b069b31cfb8f8ff956b4f9d))
- **ingest:** Pipeline normalize batches ([44e44df](https://github.com/AlexGalhardo/hexascore-app/commit/44e44df01bccf5c8c2d247f6ea1ce4fa9a921e0e))

### Fixed

- **api:** Reject empty query parameters ([d652fee](https://github.com/AlexGalhardo/hexascore-app/commit/d652fee450cece438be7f9eb354806044a930956))
- **web:** Keep sticky content below an opaque navbar ([16e1a9d](https://github.com/AlexGalhardo/hexascore-app/commit/16e1a9defc72bc7b19bd4240d1335124ab0e4a84))
- **ingest:** Share wikipedia rate-limit backoff across workers ([60f3afb](https://github.com/AlexGalhardo/hexascore-app/commit/60f3afb6310aff51bda8ee4743003956d583decf))
- **web:** One centred layout for every modal ([4f60c42](https://github.com/AlexGalhardo/hexascore-app/commit/4f60c4214febd00a153696fb55d7d293f56a35ee))
- **ingest:** Batch Wikipedia extracts and pipeline source record writes ([1a1fe4e](https://github.com/AlexGalhardo/hexascore-app/commit/1a1fe4e4f32d29f9ee9e962fcbf6b3552adf6384))
- **ingest:** Keep-alive ping skips connections busy in a pipeline ([8ffb12c](https://github.com/AlexGalhardo/hexascore-app/commit/8ffb12ccfaadfaf69857cbd643db277f7bd3c2fd))
- **ingest:** Skip unexpected cover errors per item and weak-key the busy lock ([13d0dca](https://github.com/AlexGalhardo/hexascore-app/commit/13d0dca29b54dd8dd009f110007c11ec815c76e0))
- **ci:** Raise Node heap for the local build step ([6542bc1](https://github.com/AlexGalhardo/hexascore-app/commit/6542bc1b781cee02dc3d9923bafba4ef3fdb38c3))
- **ci:** Keep the local build independent of a pulled production env file ([df84e46](https://github.com/AlexGalhardo/hexascore-app/commit/df84e464ae4c11d52506cdaed7f684c0f893b8f5))

## [0.3.0] - 2026-10-03

### Added

- **security:** Strict per-request nonce CSP and security headers ([8a35c7a](https://github.com/AlexGalhardo/hexascore-app/commit/8a35c7aed3a5b1bc766138a901fa85df94d2af6f))
- **web:** Nonce the theme script, add Vercel Analytics and Speed Insights ([ea79e26](https://github.com/AlexGalhardo/hexascore-app/commit/ea79e26e0d7b5a09780b73799f7b743ef90fcfc0))
- **flags:** Pix-checkout and public-api-signup kill switches ([5a91c5b](https://github.com/AlexGalhardo/hexascore-app/commit/5a91c5b35bf22d71609eee432835394e6f090570))
- **legal:** Affiliate and license notices, i18n review ([2d91ea2](https://github.com/AlexGalhardo/hexascore-app/commit/2d91ea254f3ce09b04976ec17c715dccc79d27f5))
- **web:** Use hexascore.app as the default domain ([a6c654b](https://github.com/AlexGalhardo/hexascore-app/commit/a6c654b9250611e15ace1db41b71799ba325ac7c))
- **ingest:** Store covers as optimized webp in vercel blob ([6092491](https://github.com/AlexGalhardo/hexascore-app/commit/6092491ebdb640ee0c55379b9c49c7c3f9a379c4))
- **web:** New navbar, search previews, nickname rules and api host ([4e88a95](https://github.com/AlexGalhardo/hexascore-app/commit/4e88a952fdc939c3d160ec1dda090869523a9660))
- **seo:** Add robots.txt and sitemaps with hreflang ([6f5f224](https://github.com/AlexGalhardo/hexascore-app/commit/6f5f224f8117b22e3a6b17059de8e5962c5d56c2))
- **seo:** Add search console cli ([c61e339](https://github.com/AlexGalhardo/hexascore-app/commit/c61e3395f28bf2aff9876a75c105095e036725f1))
- **billing:** Add sandbox PIX payment button with paid-modal lock ([66d5e49](https://github.com/AlexGalhardo/hexascore-app/commit/66d5e499070010661a8c9f51ae747a14e36c8855))
- **web:** Crowdsourced time to beat for games ([8a07590](https://github.com/AlexGalhardo/hexascore-app/commit/8a07590882b4ab416bdf9e691805c16e5229b476))
- **web:** Use TanStack Query for client-side fetches ([7dde2f0](https://github.com/AlexGalhardo/hexascore-app/commit/7dde2f0e1becf26f0762cba607c192810d82ecc8))

### Changed

- **seed:** Columnar unnest upserts and heavy-tailed rater activity ([c73bc09](https://github.com/AlexGalhardo/hexascore-app/commit/c73bc09c0469766cafcccd94d624c4a316978872))

### Fixed

- **a11y:** Section accent contrast and keyboard-scrollable code blocks ([b91d427](https://github.com/AlexGalhardo/hexascore-app/commit/b91d4278c144d988e30c54f8a4c8ca79bd8d17f0))
- **i18n:** 404 unknown locales before querying, add the app icon ([7362487](https://github.com/AlexGalhardo/hexascore-app/commit/7362487de1720034c351fe3706e6008ca09f7831))
- **catalog:** Never send control characters to Postgres ([566c919](https://github.com/AlexGalhardo/hexascore-app/commit/566c919ebaf3eaf9155329e13748931ca19ec4b0))
- **a11y:** Score badge, danger text and password checklist contrast ([314531a](https://github.com/AlexGalhardo/hexascore-app/commit/314531af74d3756ec1a8de25fdc5c04eaaafd82b))
- **api:** Keep the revoke dialog inside a 360px viewport ([0870bcb](https://github.com/AlexGalhardo/hexascore-app/commit/0870bcb0da85e6ea6428e7cb5eb452de26a3912c))
- **account:** Revoke api keys and cancel stripe subscription on purge ([6783b28](https://github.com/AlexGalhardo/hexascore-app/commit/6783b280ed12a69e002ce0dace26409cb7031671))
- **auth:** Allow sign-up emails up to 48 characters ([46efacc](https://github.com/AlexGalhardo/hexascore-app/commit/46efacca63c5c29a95b5b3264045dccb97531647))
- **ingest:** Keep the database connection alive during long fetches ([1301826](https://github.com/AlexGalhardo/hexascore-app/commit/1301826fcf68eb59b0a8bc304bb36ecbd185d679))
- **security:** Allow the vercel toolbar frame on previews ([23cff62](https://github.com/AlexGalhardo/hexascore-app/commit/23cff62b601ba812ef8c0db11ca1e71a4250e673))

## [0.2.0] - 2026-10-03

### Added

- **auth:** Add account purge, contact, rate limit, botid and password rule helpers ([a758239](https://github.com/AlexGalhardo/hexascore-app/commit/a7582398faa6185ed8523c51493fbc27ddf63d3e))
- **auth:** Add account server actions, contact action, purge cron and BotID setup ([29fa9f7](https://github.com/AlexGalhardo/hexascore-app/commit/29fa9f7628d7d87f9df8ff42cf65ba7323555d99))
- **i18n:** Add auth, account, contact and legal messages in en, pt-BR and es ([9d0de22](https://github.com/AlexGalhardo/hexascore-app/commit/9d0de2262eafaa40d01343cdade89d7916bb979e))
- **auth:** Add login, signup, verification, recovery, profile, contact and legal pages ([d5c8b85](https://github.com/AlexGalhardo/hexascore-app/commit/d5c8b85246465c8a9dce70206823222564ee5589))
- **api:** Add public /v1 catalog endpoints with zod response contracts ([4279770](https://github.com/AlexGalhardo/hexascore-app/commit/42797703a13b93bd5cef099f81945363962949f3))
- **api:** Serve /v1 on the `api.<domain>` host via the proxy ([e8a74d5](https://github.com/AlexGalhardo/hexascore-app/commit/e8a74d596c366f2b62028ce8bcef2c802a611707))
- **api:** Generate openapi.json from the zod contracts and check it in CI ([3448392](https://github.com/AlexGalhardo/hexascore-app/commit/3448392054cc3e366e71349ea9e227f9b37d48fd))
- **ratings:** Transactional rating write path with item stats recomputation ([ee01620](https://github.com/AlexGalhardo/hexascore-app/commit/ee016208e5cb16eef30c192187474fef28d7686d))
- **social:** Profiles, follows, feed, item status, lists, reports and admin data access ([2e53b9c](https://github.com/AlexGalhardo/hexascore-app/commit/2e53b9c68dbbc6719986a1bb3e725d7483685345))
- **docs:** Add /docs guides and Scalar API reference ([5d6c4dd](https://github.com/AlexGalhardo/hexascore-app/commit/5d6c4ddb2c239e35a29950650545a6f4d7da2514))
- **api:** Add /profile/api with keys, usage chart and try-it console ([b343dda](https://github.com/AlexGalhardo/hexascore-app/commit/b343ddafbdfbce76d261bd8d9b1db05a63b14e52))
- **social:** Rating panel, item status, lists, public profiles, feed, reports and onboarding ([e56e49a](https://github.com/AlexGalhardo/hexascore-app/commit/e56e49a4170cb020ca6cff478b7037e9dbdbf5ba))
- **admin:** Moderation, contact, affiliates, item merge, ingestion and admin:grant ([ab7cdb9](https://github.com/AlexGalhardo/hexascore-app/commit/ab7cdb95523e52530507e0b427572d235410874e))
- **sdk:** Generate the @hexascore/sdk TypeScript client from openapi.json ([4578dda](https://github.com/AlexGalhardo/hexascore-app/commit/4578dda451a6a37065eca9bd31fa988450795aef))
- **sdk:** Generate the hexascore Python client and readable sort enums ([8d39a03](https://github.com/AlexGalhardo/hexascore-app/commit/8d39a03d1b423cab130bba4a8cd1e51ff0fa6b06))

### Changed

- Reuse the hashed constant-time cron check ([845f2b4](https://github.com/AlexGalhardo/hexascore-app/commit/845f2b451025962be0ec03e016d2f462d0b39896))

### Fixed

- **ingest:** Send Vercel automation bypass header on revalidation ([ea4dbcc](https://github.com/AlexGalhardo/hexascore-app/commit/ea4dbcc5af85181e0057df7f0c46d71d8dc5cf9f))
- **web:** Localize every page, externalize PGlite and keep build workers off the dev database ([b7e2318](https://github.com/AlexGalhardo/hexascore-app/commit/b7e23184f8934198cbb86be91686c0a58990262e))
- **email:** Pass template props through createElement so dev logs print the action link ([46436d6](https://github.com/AlexGalhardo/hexascore-app/commit/46436d6c9b44967d2ad8ebf4d59f652157c9465f))
- **auth:** Typed redirects and clear signup field errors on edit ([4593413](https://github.com/AlexGalhardo/hexascore-app/commit/459341350695b22dfa8488a92fcb8791eab5b1ec))
- **api:** Fail closed on store errors, load Scalar styles, disable Scalar AI ([3be8576](https://github.com/AlexGalhardo/hexascore-app/commit/3be8576da8e2e719a44dcf68be8309d7da23e4a3))
- **api:** Open the try-it reference only when a key is applied ([10e072e](https://github.com/AlexGalhardo/hexascore-app/commit/10e072e80275a8c62e8e032e8b2a1dc72993f4e8))
- **i18n:** Pluralize the API rate limit label ([8acbb68](https://github.com/AlexGalhardo/hexascore-app/commit/8acbb6894f339028835dbac752d117bc3eeddc30))
- **deps:** Override js-yaml to 4.3.2 to clear high advisories ([4bf25f6](https://github.com/AlexGalhardo/hexascore-app/commit/4bf25f6288fb012f86487c54d7f6c72cdd72e75f))

## [0.1.0] - 2026-10-03

### Added

- Scaffold Next.js 16 app with i18n, theme and layout ([ce8786c](https://github.com/AlexGalhardo/hexascore-app/commit/ce8786c8bed5a657a9213f75059dc17091e97031))
- **db:** Add Drizzle schema, migrations, auth config and zod env validation ([04c1c12](https://github.com/AlexGalhardo/hexascore-app/commit/04c1c12ee414327c22e7587952e461be3ad5b5cf))
- Add error boundaries, auth route, session helper and namespaced messages ([26970a7](https://github.com/AlexGalhardo/hexascore-app/commit/26970a764b8c2052b8627a793e63cb453f6a8211))
- **api:** Add API key management, quota with credits and RFC 9457 errors ([21b0043](https://github.com/AlexGalhardo/hexascore-app/commit/21b0043b3593ac2864df6ef7a6e11c8e5de63fd4))
- **infra:** Add local Docker stack with Postgres 18, Redis and Upstash REST proxy ([486d3e8](https://github.com/AlexGalhardo/hexascore-app/commit/486d3e8b725826d56a9850e5b4ecdab7dea15f1c))
- **setups:** Add idempotent local setup scripts for Docker and PGlite ([b399413](https://github.com/AlexGalhardo/hexascore-app/commit/b399413ab0cadd9ba7fc94df106e955e44f090af))
- **ingest:** Add uv project with http and db helpers ([617460b](https://github.com/AlexGalhardo/hexascore-app/commit/617460b350f989b88195a3b493df11a41c50c6b7))
- **ingest:** Add Wikidata, Wikipedia, Open Library, Steam and GOG connectors ([970927f](https://github.com/AlexGalhardo/hexascore-app/commit/970927f9013806520ea2f11b49f9d0b144fc6614))
- **ingest:** Add normalizer, covers, revalidation and bootstrap with snapshot export ([7887c69](https://github.com/AlexGalhardo/hexascore-app/commit/7887c692ea5d2f4c427e7fc8d1e185a49e4956c7))
- **db:** Type cover attribution in item meta ([225174e](https://github.com/AlexGalhardo/hexascore-app/commit/225174eba93065f279f27ae6ce145561283e7406))
- **catalog:** Add rating categories, catalog repository and dev seed ([c34ab33](https://github.com/AlexGalhardo/hexascore-app/commit/c34ab3318db85e74fe3208e71753fa975f6f4cc3))
- **billing:** Add price catalog, Stripe and AbacatePay webhook processing with idempotent credits ([0e6f8d1](https://github.com/AlexGalhardo/hexascore-app/commit/0e6f8d1efe92b9664ac297792012ed1555654104))
- **env:** Accept Vercel Upstash and AbacatePay sandbox variable names ([350b176](https://github.com/AlexGalhardo/hexascore-app/commit/350b1766947fa53b054fefd1b971879d14ec8f1c))
- **billing:** Add webhook routes, PIX status polling and daily renewal reminder cron ([6be5f92](https://github.com/AlexGalhardo/hexascore-app/commit/6be5f920c5ef502311c62b9b8b1d5805e9944834))
- **billing:** Add pricing and checkout pages, billing actions and Stripe setup script ([205bfb6](https://github.com/AlexGalhardo/hexascore-app/commit/205bfb6bc34b3434d57ff193bad9305e235951e1))
- **catalog:** Add catalog, item and search pages with score components ([ba9bdc8](https://github.com/AlexGalhardo/hexascore-app/commit/ba9bdc8b40da01a1242b207b4ae684e31259256a))

### Fixed

- **security:** Require auth secrets on every production build and stop logging email tokens ([b590253](https://github.com/AlexGalhardo/hexascore-app/commit/b590253841c8127bfc50c58d40c3c0dd38066024))
- **infra:** Bind local ports to loopback and password-protect Redis ([802978c](https://github.com/AlexGalhardo/hexascore-app/commit/802978c34381a84daaeb4ea13be9f858c2e9179d))
- **web:** Match localized routes in proxy and keep PGlite out of the server bundle ([898655b](https://github.com/AlexGalhardo/hexascore-app/commit/898655b310acd2680cf2c5b75406d2ef7e26ff96))
- **web:** Route locale-less paths through the i18n proxy and run PGlite under Next ([1e216b1](https://github.com/AlexGalhardo/hexascore-app/commit/1e216b1085f189903d2eb72c344e6d9b1a0fa5dd))

[Unreleased]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.38.0...HEAD
[0.38.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.37.0...v0.38.0
[0.37.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.36.0...v0.37.0
[0.36.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.35.0...v0.36.0
[0.35.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.34.0...v0.35.0
[0.34.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.33.0...v0.34.0
[0.33.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.32.0...v0.33.0
[0.32.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.31.0...v0.32.0
[0.31.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.30.0...v0.31.0
[0.30.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.29.0...v0.30.0
[0.29.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.28.0...v0.29.0
[0.28.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.27.0...v0.28.0
[0.27.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.26.0...v0.27.0
[0.26.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.25.0...v0.26.0
[0.25.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.24.0...v0.25.0
[0.24.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.23.0...v0.24.0
[0.23.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.22.0...v0.23.0
[0.22.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.21.0...v0.22.0
[0.21.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.20.0...v0.21.0
[0.20.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.19.0...v0.20.0
[0.19.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.18.0...v0.19.0
[0.18.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.17.0...v0.18.0
[0.17.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.16.0...v0.17.0
[0.16.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.15.0...v0.16.0
[0.15.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.14.0...v0.15.0
[0.14.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.13.0...v0.14.0
[0.13.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.12.0...v0.13.0
[0.12.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.11.0...v0.12.0
[0.11.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.10.0...v0.11.0
[0.10.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.9.0...v0.10.0
[0.9.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/AlexGalhardo/hexascore-app/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/AlexGalhardo/hexascore-app/releases/tag/v0.1.0
