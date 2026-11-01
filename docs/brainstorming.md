# Brainstorming log

Phase 2 deliverable (see [PROGRESS.md](PROGRESS.md)). Records every question, answer, alternative and decision between the owner and the architect. Decisions here are final for the SDDs and ADRs unless this file is updated first.

Format per round: question, options with pros/cons, architect recommendation, owner answer, decision.

## Decisions received before Phase 2

### D-00: Rating input is a star widget (owner, 2026-10-28)

- **Owner statement**: to make rating easier, use the star control with half-star steps up to 10 stars, set with a single click. Better than a 0–100 value with a horizontal slider.
- **Architect notes**:
  - Half-star steps give 20 levels per category (0.5 to 10.0). Store `stars × 10` so `c1..c6` stay in 0–100 in steps of 5; aggregation, the `overall` generated column and the public API contract are unchanged.
  - Existing ratings (0–100, any integer) are rounded to the nearest multiple of 5 in the migration; the rounding is recorded as a one-off audit event.
  - Zero stars is not a value: "no rating" is the absence of the row, as today.
  - Display of the aggregate (HexaScore 0–100) is a separate decision from the input control; open for Phase 2.
  - Accessibility: the widget must be keyboard-operable (arrow keys in 0.5 steps) and expose `role="slider"` with `aria-valuetext` such as "7.5 of 10 stars".
- **Decision**: accepted. Goes to the scoring SDD and the web/mobile design-token spec.

### D-01: Full secondary entities (owner, 2026-10-28)

- **Owner statement**: the site must have complete sub-categories such as authors, publishers, developers, brands, etc.
- **Architect notes**:
  - Today the legacy model keeps creators, platforms and genres as strings inside `items.meta` jsonb, so there are no pages, no stable URLs and no counts for them. The rewrite promotes them to first-class entities with their own table, slug, page and API endpoint.
  - Proposed entity set per domain (to confirm in Round 1): movies/series: directors, writers, cast, production companies, networks/streaming services, franchises/collections, genres, countries; games: developers, publishers, platforms, engines, franchises/series, genres, game modes; books: authors, publishers, series, genres, languages. "Brands" is modelled as a generic `organizations` entity with a `kind` (studio, publisher, network, platform holder, imprint) so one company can act in several roles across domains.
  - Source of identity: Wikidata QID when available (same key as items), otherwise a HexaScore-native ID; external IDs (TMDB person/company, IGDB company, Open Library author) kept in a jsonb map like items do today.
  - Each entity page lists its items by type, counts, average HexaScore where there are enough rated items, and is part of the sitemap for SEO.
- **Decision**: accepted in principle. Entity list and the "organization with roles" model are confirmed in Round 1; goes to the catalog SDD.

### D-02: "Top 10 trending in the last 30 days" per master category (owner, 2026-10-28)

- **Owner statement**: each of the four master categories (games, movies, series, books) shows a section with the top 10 trending items of the last 30 days; these are the main recommendations for users to read, play or watch.
- **Architect notes**:
  - Trending must be computed from HexaScore's own signals (ratings, status changes, list adds, item page views) over a rolling 30-day window, decayed so recent activity counts more, and gated by a minimum activity threshold so a handful of votes cannot push an item in. Wikidata sitelinks (today's `popularity`) are a static prior, not a trend.
  - Materialised per category once per hour (or on a schedule in vukem's scheduler), stored in a small table, served from cache; never computed on the request path.
  - Visible to anonymous users (SEO and first impression); the same data feeds the public API as `GET /v1/{type}/trending`.
  - Anti-manipulation: page views deduplicated per session/IP per day; rating bursts from new accounts weighted down (same signals as the score anti-abuse model).
- **Decision**: accepted. Goes to the recommendation SDD and the public API SDD.

## Round 1 answers (owner, 2026-10-28)

Answered through the interactive checklist in chat (the owner prefers clicking options to typing; `grill-me` is installed as a stub that calls a missing `grilling` skill, so the architect ran the interview with `AskUserQuestion`, recommended option first).

- **Q1 Licensing → "Wikidata-only in the API"** (D-03). TMDB and IGDB fields are display-only on the site and never leave through the public API; every field carries `source` and `licence`; TMDB-derived values are refreshed within the 6-month cache window; IGDB attribution visible in a static place. No commercial contracts pursued now. Architect note: this is the conservative and cheap option; it also means the paid API's film/TV/game coverage is Wikidata + Wikipedia + Open Library only, which must be stated on the pricing and docs pages so buyers are not surprised.
- **Q2 Rendering → "Inertia v3 + React"** (D-04). Architect disagreement recorded: for a catalogue whose growth depends on search traffic, server-rendered HTMX gives better Core Web Vitals, no Node in production and a smaller operational surface; Inertia is BETA in vukem. The owner chose Inertia for React reuse and richer interactivity. Conditions attached by the architect so the choice does not cost SEO: public pages (home, listings, item, entity, search, user profiles, docs) must be server-side rendered through vukem's Inertia SSR (`vukem ssr:start`, Node sidecar on Railway) with a health check and a hard rule that CSR fallback is a paging alert, not an accepted state; Core Web Vitals budgets in the web SDD apply to the SSR output; the Node runtime exists only for SSR and the Vite build, never for application logic. Round 2 confirms the SSR requirement.
- **Q3 Auth → "vukem auth + social"** (D-05). Passkeys deferred and filed in `vukem-feedback.md`; Better Auth scrypt hashes handled by a compatibility hasher on first login (Phase 3 verifies Better Auth's default hash format against `legacy/apps/web/lib/auth.ts`).
- **Q4 Parity cut → defer all four** (D-06): PIX/AbacatePay and credit packs, feed and follows, passkeys, per-locale OpenAPI documents. Everything else from the legacy app is in scope for the first vukem release. **Amended 2026-10-29 (owner, in chat):** passkeys are no longer deferred; they return on the vukem `auth` passkey API (VF-001) and legacy credentials are imported when compatible (ADR 0017). The other three deferrals stand. **Amended 2026-10-31 (owner, in chat, Q-A1):** feed and follows, plus reviews, comments and reactions, enter right after `v1.0.0` as ROADMAP Phase 7 (P7.A–P7.C); `v1.0.0` stays parity + P2.H launch readiness + cutover. PIX/credits and per-locale OpenAPI documents stay deferred. **Phase 7 rollout (2026-11-01, P7.A.1.1):** the social layer ships behind `SOCIAL_ENABLED` (default off) and a per-user feed opt-in (default off) until the owner answers Q-A5 (lawful basis, legacy follows); design in SDD 14 §12, ADR 0020.

## Round 2 answers (owner, 2026-10-28)

- **Q5 SSR → "SSR Node mandatory"** (D-07). Public pages are always server-rendered through vukem's Inertia SSR; CSR fallback raises an alert; Node exists only for SSR and the Vite build.
- **Q6 Railway → "Up to US$30/month"** (D-08). One Go service (HTTP + scheduler + queue workers in the same process) plus Postgres; the Node SSR process runs inside the same container, supervised by the Go process (`vukem ssr:start` behaviour) or as a sidecar command; no Redis, no MongoDB (each needs an ADR later); daily Postgres backups must fit the ceiling. Architect note: a single process means ingestion jobs compete with HTTP for CPU; the ingestion SDD must throttle jobs and keep them idempotent so a restart mid-run is safe. The split into a separate worker service is the first thing to do when the ceiling is raised.
- **Q7 Critic scores → "Drop at launch"** (D-09). `external_scores` leaves the model; critic data returns only under licence. Removes an open legal risk from the legacy app.
- **Q8 Books → "Work + editions"** (D-10). Item = work (Wikidata QID); editions carry ISBN-13, language, publisher, year, cover; ratings and statuses attach to the work; Brazilian editions highlighted.

## Round 3 answers (owner, 2026-10-28)

- **Q9 Ingestion → "Port to Go inside vukem"** (D-11). Connectors rewritten in Go on vukem's queue and scheduler, same process as HTTP under the US$30 ceiling; idempotent, throttled, resumable jobs; status in the admin panel. The Python package in `legacy/webscrappings/` and its 77 tests are the specification for the port (normalisation rules, hashing, cover sizes). Python stays only as a reference until parity is proven by comparing outputs on a sample.
- **Q10 Mobile v1 → "Browse + rate + status + lists"** (D-12). Login (email/password, Google, 2FA), search, item and entity pages, trending, star rating, status, time-to-beat, lists, profile. No billing, no API keys, no admin in the app.
- **Q11 Score display → "Stars 0–10 everywhere"** (D-13). The aggregate is shown as stars with one decimal (e.g. 7.8/10) on the site, in the app and in the API. Architect implementation note: storage stays 0–100 integers (half-star × 10 per category, aggregate kept at full precision) and the API exposes `score` as a 0–10 number with one decimal plus `ratingCount` and a `confidence` field; the 0–100 value is not exposed, so the public contract has one scale. Supersedes the "0–100 aggregate" note in D-00.
- **Q12 Monetisation → "Paid API via Stripe and AbacatePay"** (D-14, pending clarification). Site and app free, no ads, revenue from the developer API. The answer names AbacatePay while Round 1 (D-06) deferred PIX/AbacatePay to the second release; Round 4 asks which release carries it.

## Round 4 answers (owner, 2026-10-28)

- **Q13 PIX → "Second release"** (D-14 clarified). First release: Stripe only. AbacatePay PIX follows once vukem's billing model is stable in production.
- **Q14 Moderation → "Admins only at launch"** (D-15). Admin queue with a target SLA under 24 h; public status per suggestion; contributor accuracy tracked from day one so trusted moderators can be enabled later without a data migration.
- **Q15 Search → "Postgres FTS + pg_trgm"** (D-16). As today, plus `unaccent` in the pipeline and per-locale configuration. The catalog SDD keeps the search behind one interface so an engine swap is contained.
- **Q16 URLs and locales → "en without prefix, /pt-br and /es"** (D-17). Same scheme as the legacy app; existing URLs keep working; `hreflang` on every page; item slugs identical across locales.

## Round 5 answers (owner, 2026-10-28)

- **Q17 Backups → "RPO 24 h, RTO 1 h"** (D-18). Daily `pg_dump` to a Railway bucket or S3-compatible store, 30-day retention, restore tested monthly by script.
- **Q18 Brand → "HexaScore is final"** (D-19). `APP_NAME=HexaScore`, hexascore.app, api.hexascore.app, packages `@hexascore/sdk` and `hexascore`.
- **Q19 Versioning → "Restart at v0.1.0"** (D-20). Releases follow the vukem pattern (tags `vX.Y.Z`, generated changelog, GitHub release); one release per roadmap task group; parity with the legacy app = `v1.0.0`; the legacy app stays at 0.0.25 on branch `legacy`.
- **Q20 Cutover → "Big-bang after parity + QA"** (D-21). Staging on Railway with a data copy; gates: Playwright/axe, k6, Schemathesis, ZAP green plus a parity checklist; final data migration in a short read-only window on the legacy app; DNS switch; legacy kept 30 days for rollback.

## Architect decisions taken without a question (recorded for the SDDs; the owner may override)

- **A-01 Mobile auth**: the React Native app uses vukem access tokens (opaque, hashed, with abilities) obtained by email/password + 2FA or Google; no cookies. First-party write endpoints (`/app/v1/...`) are separate from the paid read API (`/api/v1/...`) and gated by abilities, not by API plans.
- **A-02 Recommendations**: content-based first (genres, creators, co-occurrence in lists and statuses, "similar items"); collaborative filtering only after a documented minimum of real users and ratings; trending per D-02.
- **A-03 Railway region**: the region closest to Brazilian users among those Railway offers at implementation time (US East as the default if no South American region exists); recorded in the infra ADR with the measured latency.
- **A-04 Legacy data migration**: users, ratings, statuses, time-to-beat, lists, API keys, subscriptions and catalogue data migrate from Neon; the 10k synthetic users and their ~2.1M ratings are purged before the copy; Better Auth tables are transformed into vukem's user/session/token tables.
- **A-05 Licence check for Open Library** and any other source added later happens in Phase 3 before the ingestion SDD is final; a source without a compatible licence is not ingested.
- **A-07 Licence of HexaScore's own data** (raised by SDD 01 open issue 7): contributors grant HexaScore a perpetual, irrevocable, non-exclusive licence on suggestions and reviews (same pattern as TMDB's terms, see web-research §2.4), stated in the terms of use; HexaScore-sourced fields (scores, counts, trending, community edits, HexaScore IDs and slugs) are served to API customers under the API terms: use within the customer's product, attribution "Data by HexaScore" with a link, no redistribution as a bulk dataset, no training of models without written permission. No open-data licence (CC) is applied to HexaScore data at launch; it can be opened later per field thanks to the provenance columns. Owner may override.
- **A-06 Category labels**: the six categories per type stay as in `legacy/apps/web/lib/categories.ts` unless Phase 3 finds a reason to change them; changing labels is a product decision and would reopen this file.

## Phase 3 decisions (owner, 2026-10-28, from SDD open issues)

Asked through the interactive checklist after the SDD drafts; they close every `owner` row in [sdd/open-issues.md](sdd/open-issues.md).

- **D-22 Dev and UI dependencies approved**: Biome, Vitest, Testing Library, Playwright, axe-core, Lighthouse CI, Base UI (web), Redocly CLI (OpenAPI lint, dev only; owner yes 2026-10-29, Q-REDOCLY); Expo SQLite for the persisted query cache (mobile). Exact versions pinned at implementation (no `^`, no pre-releases).
- **D-23 Legacy credit balances** are honoured as extra quota consumed before the plan quota; no refund flow; new credit packs only with PIX in the second release.
- **D-24 Password maximum 128 characters** (minimum as legacy).
- **D-25 Google-only accounts set a password before managing 2FA** (vukem rule kept; UI explains).
- **D-26 GitHub Actions is available** for `hexascore-app`; the full pipeline runs there; pre-push mirror kept as a convenience.
- **D-27 Railway spend**: e-mail alert at US$24, hard limit US$30. Architect note: a hard limit at the ceiling can take the site down at month end during a legitimate spike; SDD 12 adds a "projected spend" alert at day 20 so the owner can raise the limit in time.
- **D-28 DNS on Cloudflare**: the `hexascore.app` zone is hosted (or moved before cutover) on Cloudflare for CNAME flattening at the apex; proxy off at launch, available later as a CDN for covers.
- **D-29 LGPD basis for personalisation**: legitimate interest for aggregates and trending; personalised recommendations on by default with an opt-out in settings; privacy policy updated; legal review before cutover (D-21 gate).
- **D-30 Catalogue extras deferred**: game engines, edition contributors (translator, illustrator) and episode-level data come after v1.0.0; the relation model leaves room for them.
- **D-31 Suggestion quota**: 3 per day for every user at launch; a future trusted-moderator role gets its own quota by a new decision.
- **D-32 Off-site backups**: monthly dump downloaded by the owner through a signed link until an external bucket fits the budget; daily dumps stay on Railway.
- **D-33 API prices and limits**: legacy table kept at launch; review 90 days after cutover; pricing page states the API data coverage.
- **D-34 Android builds**: EAS Build on the free tier while it lasts; a Gradle pipeline on GitHub Actions is prepared as the fallback.
- **D-35 No Prometheus `/metrics` endpoint at launch**; observability uses only Railway native tools (service metrics, logs, OTLP traces to Railway, deploy and usage alerts). No Grafana, Grafana Cloud or promtool in this project (owner, amended 2026-10-31).
- **D-37 Tooling approved** (dev/CI only, never in the binary): git-cliff, commitlint, Task (Taskfile), markdownlint-cli2, Redocly CLI, and `golang.org/x/crypto/scrypt` for Better Auth hash verification at first login. Exact versions pinned at implementation.
- **D-38 Lists of deleted accounts**: public lists are anonymised (author shown as "deleted user") and stay reachable; private lists are deleted (SDD 14).
- **D-39 Data protection officer**: the owner is the LGPD "encarregado"; public contact `privacy@hexascore.app` (mailbox created on the domain); named in the privacy policy (SDD 13).
- **D-36 IGDB terms**: the owner reads the Twitch Developer Services Agreement and records the decision in `docs/legal/igdb.md` before the ingestion milestone; until then IGDB fields are site-only, refreshed every 24 h, never cached longer than 7 days.

## Closure

The owner confirmed the statement of intent below and declared the brainstorming closed on 2026-10-28 ("Sim, brainstorming encerrado"). Decisions D-00 to D-21 and A-01 to A-06 are final for Phase 3; reopening any of them requires editing this file first.

## Statement of intent (confirmed 2026-10-28)

- **Outcome**: HexaScore rebuilt on vukem (Go) with Inertia v3 + React and mandatory SSR, deployed on Railway under US$30/month, reaching parity with the Next.js app minus PIX, credits, feed/follows, passkeys and per-locale OpenAPI, then adding community edit suggestions, an Android app and own trending/recommendations.
- **User**: readers, players and viewers in Brazil and abroad (en, pt-BR, es) who want one place to track and rate games, films, series and books; developers who buy the read-only API.
- **Why now**: Vercel/Next coupling and cost, manual ingestion, legal exposure on third-party data, and a framework (vukem) the owner also develops.
- **Success**: parity QA gates green (Playwright/axe, k6, Schemathesis, ZAP), Core Web Vitals and API latency budgets met on SSR output, cutover with no data loss and rollback available for 30 days; v1.0.0 tagged.
- **Constraints**: Postgres only (Redis/Mongo need an ADR); single Go process plus Node SSR in one container; no third-party data in the paid API; star input with half steps; aggregate shown as stars 0–10 with one decimal; admins-only moderation at launch; Postgres search; Stripe only in v1.
- **Out of scope for v1**: PIX/AbacatePay, credit packs, feed/follows, passkeys, critic scores, consumer subscriptions, ads, Meilisearch, multi-instance scaling, collaborative-filtering recommendations.

## Round 1 (sent 2026-10-28)

Method note: the prompt asked for the `grill-me` skill; it is not installed (it delegates to a missing `grilling` skill), so `interview-me` is used with the prompt's round format (max 5 questions per round, each with the architect's guess and recommendation). Current hypothesis of intent: "a faster, cheaper-to-run, legally clean HexaScore on Go/vukem that reaches feature parity with the Next.js app, then adds community edits, mobile and better scores". Confidence ~60%: missing the licensing posture, the frontend rendering model, the auth path, the parity cut and the budget.

### Q1: Licensing posture for third-party data

- **Question**: Which posture do you want for TMDB, IGDB and critic scores (Metacritic, Rotten Tomatoes) before launch?
- **Options**:
  - (a) Pursue written agreements now (TMDB <sales@themoviedb.org> "Commercial API Use"; IGDB <partner@igdb.com>) and keep the fields; if an answer is negative or absent by launch, fall back to (b) by configuration.
  - (b) Wikidata + Wikipedia + Open Library only in the paid API; TMDB/IGDB fields displayed on the site but excluded from the API and refreshed within TMDB's 6-month cache window.
  - (c) Drop TMDB/IGDB entirely; Wikidata/Wikipedia for everything.
  - Critic scores: keep display-only values that come from Wikidata review-score statements (CC0) with attribution, or drop them at launch.
- **Architect guess/recommendation**: (a) with (b) as the automatic fallback, and critic scores dropped at launch. Every field carries `source` + `licence` so the fallback is a configuration change, not a migration.

### Q2: Rendering model inside vukem

- **Question**: HTMX + `html/template` (vukem default starter) or Inertia v3 + React for the web app?
- **Pros/cons**:
  - HTMX: server-rendered HTML, smallest JS, best Core Web Vitals and SEO for a catalogue, no Node in production, matches the PageSpeed references; cost: React components from `legacy/` are not reused, interactive widgets (star rating, dialogs) are written as small Vite/TypeScript islands.
  - Inertia + React: reuses React knowledge and some component code, richer client interactivity; cost: larger bundles, SEO needs the Node SSR server (vukem falls back to client rendering on SSR failure), two runtimes to operate, Inertia is BETA in vukem.
- **Architect guess/recommendation**: HTMX + `html/template` (or templ) for everything public and account pages; Vite TypeScript islands only where needed. Mobile does not share components either way; it shares the API contract and design tokens.

### Q3: Authentication

- **Question**: Replace Better Auth with vukem `auth` + `social` (both ALPHA) at parity minus passkeys, keep Better Auth as a separate Node service, or use an external identity provider?
- **Pros/cons**: vukem auth gives email+password, argon2id, 2FA via `install:auth --2fa`, social login (Google) and access tokens in one runtime; passkeys are not documented in vukem and would be a `vukem-feedback.md` request. Better Auth as a sidecar keeps passkeys but adds a Node service and a second session model. External IdP adds cost and a dependency on a vendor for the most sensitive data.
- **Architect guess/recommendation**: vukem auth + social; passkeys deferred and requested upstream; existing users migrated with password hashes re-hashed on first login (Better Auth uses scrypt by default; vukem supports argon2id and bcrypt, so a one-time compatibility hasher is needed and must be confirmed in Phase 3).

### Q4: Parity cut for the first vukem release

- **Question**: Which legacy features must be in the first production release of the rewrite, and which can wait?
- **Candidates to defer**: PIX via AbacatePay (behind a flag today), credit packs, feed, follows, reports, passkeys, impersonation, the three-locale OpenAPI documents (one English document plus translated reference UI instead).
- **Must keep (architect view)**: catalogue pages, search, ratings (star widget), status, time-to-beat, lists, auth with 2FA and Google, profile, public API v1 with keys and plans, Stripe subscriptions, admin (items, merge, reports, contact, ingestion status), sitemap/SEO, i18n en/pt-BR/es.
- **Architect guess/recommendation**: defer PIX, credits, feed, follows, passkeys and per-locale OpenAPI documents to a second release; keep everything else.

### Q5: Railway budget and data stores

- **Question**: What is the monthly infrastructure budget ceiling at launch (USD), and is a single region with one app instance plus one Postgres acceptable for the first months?
- **Context**: legacy estimate on Vercel+Neon was ~US$52/month at launch and ~US$160 in growth. On Railway the app, Postgres and a worker fit in one project; Redis is only needed when the app runs more than one instance (shared rate limits, cache, queue) and MongoDB has no identified use case, so both start absent (each later addition needs an ADR per requirement 4).
- **Architect guess/recommendation**: ceiling US$60/month at launch; single region (us-east or South America if Railway offers it, decided by latency to Brazilian users); one app instance, one worker for ingestion/scheduler, one Postgres with daily backups; Redis added only when scaling to two app instances.
