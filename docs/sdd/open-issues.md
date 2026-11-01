# SDD open issues: triage log

Every SDD ends with an "Open issues" section. This file is the architect's triage of those items so the roadmap can pick them up. Status: `decided` (architect decision recorded here and the SDD is updated or will be in the review pass), `owner` (needs the owner; batched into the Phase 3 question round), `roadmap` (becomes a task with the trigger named), `vukem` (filed in `../../vukem-feedback.md`).

| SDD | # | Issue (short) | Status | Disposition |
| --- | --- | --- | --- | --- |
| 01 | 1 | Legacy ISBN-only books lack edition language/publisher | decided | Take language and publisher from Open Library edition data when present; otherwise language `und`, publisher null, data-quality flag DQ-10; never block the migration. |
| 01 | 2 | Open Library licence check (A-05) | decided | Read 2026-11-01 (P6.A.1.1, `docs/legal/open-library.md`): the Internet Archive asserts no rights over the data, so it is API-eligible under licence `OPEN-LIBRARY-NO-RIGHTS-ASSERTED`; covers have no explicit grant and stay out of the API. Identifiers and titles only (`limitedSources`) until the owner confirms the file. |
| 01 | 3 | Engines, edition contributors (translator, illustrator), episode-level data not in v1 | decided (owner) | Answered 2026-10-28, see brainstorming.md D-22 to D-36. Original note: asked in the Phase 3 question round; default if unanswered: out of v1.0.0. |
| 01 | 4 | Cursor pagination over (sort value, id) not in vukem docs | vukem | VF-005. App implements with `WhereRaw` and a signed cursor. |
| 01 | 5 | `t.JSON` / `t.Enum` mapping on Postgres undocumented | vukem | VF-006. DDL uses raw jsonb and text + CHECK. |
| 01 | 6 | TMDB/IGDB display in the mobile app | decided | The app is part of the same service as the site (display, not redistribution); attribution shown in the app's item screen and about screen. Owner still owes the Twitch agreement reading (ADR 0006). |
| 01 | 7 | Licence of HexaScore's own data | decided | A-07 in brainstorming.md. |
| 01 | 8 | Sitemap helper missing; thin entity pages | vukem + decided | VF-007 for the helper. Entities enter the sitemap only with ≥ 3 linked items; others are `noindex,follow` until they qualify. |
| 02 | 1 | Release-date precision and lifecycle field needed for the release gate | decided | SDD 01 gains `release_precision` (`day`, `month`, `year`, `unknown`) and `release_status` (`announced`, `released`, `cancelled`); applied in the review pass. |
| 02 | 2 | Default `m` per type and prior unverified | roadmap | Measure on the first staging copy; formula v1 constants frozen before cutover. |
| 02 | 3 | Recompute latency on 10k-rating items unmeasured | decided | P1.D.2.2 measured the recompute statement: p95 28.6 ms at 5,000 and 48.4 ms at 10,000 ratings, under the 60 ms bar, so no covering index. The k6 scenario `qa/k6/rating-hot-item.js` re-checks the end-to-end write on staging. |
| 02 | 4 | Admin custom row actions not in vukem docs | vukem | VF-008. Custom admin routes in the app. |
| 02 | 5 | ORM handling of generated column `ratings.overall` | vukem | VF-009. Test at implementation; fall back to raw SQL insert. |
| 02 | 6 | Delivering i18n catalogues to Inertia/React | decided | SDD 09 specifies: catalogues served as a shared Inertia prop per locale (`Once`), same JSON files as mobile (ADR 0009). |
| 02 | 7 | Weight tiers and burst thresholds have no traffic data | roadmap | Post-launch review at 90 days; formula v2 or recorded keep. |
| 02 | 8 | Ban/unban must trigger item recomputes | decided | SDD 03 emits `user.banned`/`user.unbanned` events; SDD 02 subscribes and enqueues recompute for the user's rated items. Checked in the review pass. |
| 04 | 1 | Can `source=community` values leave through the paid API? | decided | Yes, under A-07 and the API terms. |
| 04 | 2 | Fixed 3/day cap blocks higher quotas for trusted contributors | decided (owner) | Answered 2026-10-28, see brainstorming.md D-22 to D-36. Original note: asked in the Phase 3 question round; default: keep 3/day for everyone at launch (requirement 5), revisit with the trusted-moderator role. |
| 04 | 3 | Connectors must not overwrite `source=community` fields | decided | Connectors skip community-sourced fields for 180 days and open a `system` suggestion when the source value diverges; added to SDD 06 in the review pass. |
| 04 | 4 | Admin `BeforeSave` without previous values | vukem | VF-010. |
| 04 | 5 | Admin panel lacks custom/bulk actions, assignment, diff view | vukem | VF-008 (same entry). Moderation queue is an app page at `/admin/moderation`. |
| 04 | 6 | No persistent in-app notifications in vukem | vukem + decided | VF-011. App owns `user_notifications`; SDD 09 renders them. |
| 04 | 7 | Fixed-window limiter alignment to UTC day undocumented | decided | Daily quota counted in the database, not the limiter. |
| 04 | 8 | Automatic similarity check for copied synopses deferred | roadmap | Post-launch. |
| 04 | 9 | Merge is not reversible | decided | SDD 01 merge keeps a snapshot of both items (jsonb) in the audit log for manual restore; a reversible merge is post-launch. |
| 03 | 1 | Better Auth scrypt parameters and NFKC step unverified | roadmap | First auth task: run `bun install` in `legacy/`, read `node_modules/better-auth` password code, pin parameters in a Go test with a hash produced by `legacy/apps/web/scripts/seed-admin.ts` (or `admin-grant.ts`). |
| 03 | 2 | Better Auth 2FA secret encryption format unverified | decided | Export with a Bun script in `legacy/` that decrypts via Better Auth's own code and pipes plaintext secrets into the Go importer over stdin; never written to disk. |
| 03 | 3 | Google-only accounts must set a password before managing 2FA | decided (owner) | Answered 2026-10-28, see brainstorming.md D-22 to D-36. Original note: asked in the Phase 3 question round; default: keep vukem's rule (set a password first) and show the explanation in the UI. |
| 03 | 4 | Bot protection not in vukem | decided | ADR 0016; VF-003. |
| 03 | 5 | Passkeys back in scope (ADR 0017), wait for the vukem release with VF-001 | vukem | VF-001. |
| 03 | 6 | Impersonation not in vukem, not in v1 | vukem + decided | VF-012 to be filed; admins use read-only "view as" in the admin panel until then. |
| 03 | 7 | vukem session timeout env names unverified | roadmap | Verified when `install:auth` is run; recorded in the configuration table of SDD 11. |
| 03 | 8 | Locale for queued mail outside a request | vukem | VF-013 to be filed; workaround: the job payload carries the locale and the mailer is called with it explicitly. |
| 03 | 9 | Legacy password maximum of 32 characters is weak | decided (owner) | Answered 2026-10-28, see brainstorming.md D-22 to D-36. Original note: asked in the Phase 3 question round; architect recommends 128 (argon2id has no practical length limit). |
| 03 | 10 | `install:auth` column and page names unverified | roadmap | Resolved by running the scaffold in the first auth task; SDD updated after. |
| 05 | 1 | Host-based routing (`api.hexascore.app/v1` → `/api/v1`) not in vukem docs | vukem + decided | VF-015. App middleware rewrites the path when `Host` matches the API host; OpenAPI `servers` lists `https://api.hexascore.app/v1`. |
| 05 | 2 | Licence of HexaScore/community data in the API | decided | A-07. |
| 05 | 3 | `lookup` by TMDB/IGDB/Steam/IMDb IDs may be licence-sensitive | decided | Identifiers themselves are facts; lookup accepts them as input and returns only allowlisted fields (ADR 0006). Response never echoes third-party IDs that came from a restricted source; Wikidata-sourced IDs are fine. |
| 05 | 4 | Legacy `credit_ledger` balances cannot be spent in v1 | decided (owner) | Answered 2026-10-28, see brainstorming.md D-22 to D-36. Original note: asked in the Phase 3 question round; architect default: honour balances as extra quota until consumed (no refunds flow needed). |
| 05 | 5 | Ability names differ between SDD 03 and SDD 02/05 | decided | Convention in `README.md`; SDD 03 updated in the review pass. |
| 05 | 6 | camelCase payloads vs vukem snake_case envelope keys | vukem + decided | VF-014; convention in `README.md`. |
| 05 | 7 | Problem extension members and Scalar UI translation not in vukem docs | vukem | VF-016 (problem extensions); Scalar UI translation dropped: English reference UI with translated surrounding docs pages (D-06 spirit). |
| 05 | 8 | Keyset cursors on allowlisted sort columns | vukem | VF-005 (same). |
| 05 | 9 | Problem code casing differs across SDDs | decided | kebab-case convention; SDD 03 updated in the review pass. |
| 05 | 10 | Legacy prices and limits need owner review | decided (owner) | Answered 2026-10-28, see brainstorming.md D-22 to D-36. Original note: asked in the Phase 3 question round. |
| 05 | 11 | Per-field API coverage unmeasured | roadmap | `qa/sql/api_coverage.sql` runs on the first staging copy; numbers go to the pricing/docs pages. |
| 07 | 1 | Generated tsvector with `CASE` over `regconfig` may be non-immutable | decided | Use a trigger-maintained column from the start; no generated column. |
| 07 | 2 | `unaccent` before `portuguese_stem` may weaken matching | roadmap | Golden query set decides; alternative: index both accented and unaccented forms. |
| 07 | 3 | Migrations outside a transaction (`CREATE INDEX CONCURRENTLY`) not in vukem docs | vukem | VF-017. Workaround: build indexes before cutover on an empty-ish table; concurrent builds via a one-off script. |
| 07 | 4 | `inertia.Scroll` meta shape for offset pages | roadmap | Verify at implementation; fall back to plain props + `WhenVisible`. |
| 07 | 5 | Route prefix disagreement (`/v1` vs `/api/v1`) | decided | Convention in `README.md`; SDD 01 §6.2 updated in the review pass. |
| 07 | 6 | `item_alt_titles` table missing from SDD 01 | decided | Added to SDD 01 §5 in the review pass; SDD 06 fills it from Wikidata aliases. |
| 07 | 7 | API search matching on TMDB-sourced synopses | decided | Search for the paid API ranks only on allowlisted-source text (titles, alt titles, Wikipedia synopsis); site search may use all text. Two configurations of the same interface. |
| 07 | 8 | Suggest p95 at 2-character prefixes unverified | roadmap | k6 decides; fallback 3-character minimum. |
| 07 | 9 | Railway Postgres version and index sizing unverified | roadmap | SDD 11 records the Postgres version (≥ 16 target) and plan size; measured on the first staging copy. |
| 06 | 1 | Open Library licence unread (A-05) | decided | Read 2026-11-01 (same as 01-2): ingest and keep `openlibrary` in `CATALOG_API_SOURCES`; covers are not API-eligible; record in `docs/legal/open-library.md`. |
| 06 | 2 | Twitch Developer Services Agreement unread; IGDB retention unknown | decided (owner) | Answered 2026-10-28, see brainstorming.md D-22 to D-36. Original note: asked in the Phase 3 question round (owner reads and records in `docs/legal/igdb.md`); until then IGDB fields refreshed every 24 h and never cached longer than 7 days. |
| 06 | 3 | `contrib/imaging` lacks WebP encoding | vukem + decided | VF-004; plan: pinned `cwebp` via `contrib/engine`, parity test ±10 % size vs Pillow output. |
| 06 | 4 | Cover storage key format differs between SDD 01 and SDD 06 | decided | Keep the legacy key `covers/<id>-<hash16>.webp` (no migration of 80k objects); SDD 01 §5.4 updated in the review pass. |
| 06 | 5 | Queue priority/concurrency with the database driver inside `serve` undocumented | vukem | VF-018. |
| 06 | 6 | Manual job triggers need custom admin actions | vukem | VF-008 (same); app routes meanwhile. |
| 06 | 7 | No documented way to register app-level OpenTelemetry metrics | vukem | VF-019. |
| 06 | 8 | Steam `appdetails` and GOG terms unread | decided | Store images site-only; store IDs used for links only; terms read before enabling images in the app (roadmap task). |
| 06 | 9 | New Wikidata role properties and cast cap of 15 unchecked against fixtures | roadmap | Part of the connector parity test (200-QID sample). |
| 06 | 10 | TMDB changes/exports and IGDB webhooks unused in v1 | decided | Accepted; trigger to revisit: weekly TMDB run > 2 h or IGDB sync > 1 h. |
| 06 | 11 | Per-host throttle and circuit breaker are app code | vukem | VF-020. |
| 08 | 1 | `item_links` (`based_on`, Wikidata P144) belongs to SDD 01/06 | decided | Added to SDD 01 §5 and SDD 06 connector scope in the review pass. |
| 08 | 2 | Lists have no owning SDD | decided | New SDD 14 "user library" (lists, statuses view, profiles, reports, contact, notifications) commissioned. |
| 08 | 3 | `trending_exclusions` vs a general takedown flag | decided | SDD 01 gets `items.visibility` (`public`, `hidden`, `takedown`); trending and search exclude non-public; no separate exclusions table. |
| 08 | 4 | `users.personalization_opt_out` missing in SDD 03 | decided | Added to SDD 03 §5 in the review pass; exposed in profile privacy settings. |
| 08 | 5 | `SearchResultClicked` event contract pending SDD 07 | decided | SDD 07 §6 defines it (item id, query hash, position, locale); aligned in the review pass. |
| 08 | 6 | Business plan name and activity endpoint quota pending SDD 05 | decided | Plan `business`; `/api/v1/{type}/{id}/activity` counts as 1 request, business plan only. |
| 08 | 7 | Trend weights and thresholds untuned | roadmap | 90-day post-launch review (same as 02-7). |
| 08 | 8 | `trending` eager instead of `Defer` because of SSR | decided | Correct reading of D-07: public SEO content is eager; `Defer` only for personalised or below-the-fold non-SEO props. SDD 09 follows this rule. |
| 08 | 9 | Similar items in `/api/v1` and relation provenance | decided | Similarity computed from allowlisted-source relations only for the API variant; site variant may use all relations (same split as search, 07-7). |
| 08 | 10 | Buffered counter with per-minute flush and `HourlyAt` not in vukem docs | vukem | VF-021. |
| 08 | 11 | Nightly similar-items compute at scale unmeasured | roadmap | Measure on staging; incremental by `feature_hash` if > 60 min. |
| 08 | 12 | LGPD legal basis for personalisation | decided (owner) | Answered 2026-10-28, see brainstorming.md D-22 to D-36. Original note: asked in the Phase 3 question round; default: legitimate interest for aggregates, consent toggle for personalised recommendations (opt-out exposed). |
| 08 | 13 | "Hype" list for unreleased items | roadmap | Post-launch. |
| 10 | 1 | `/app/v1/me/notifications` missing from SDD 05 | decided | Owned by SDD 14 (notifications inbox); SDD 05 §6.4 cross-references it in the review pass. |
| 10 | 2 | `POST /app/v1/client-errors` crash intake missing | decided | Added to SDD 05 §6.4 and SDD 12 (opt-in, rate-limited, 7-day retention, no PII) in the review pass. |
| 10 | 3 | No density-specific cover widths in SDD 06 | decided | SDD 06 produces 600 px cover and 96×144 thumb only (legacy parity); the app requests the 600 px variant and scales; a 2× thumb (192×288) is added post-launch if scroll performance needs it. |
| 10 | 4 | Icon set and ICU message library depend on SDD 09 | decided | SDD 09 names them; SDD 10 follows. |
| 10 | 5 | App item reads count as views keyed by token id | decided | Yes; SDD 08 dedups per token id per day like sessions. |
| 10 | 6 | Persisted cache storage backend is a new dependency | decided (owner) | Answered 2026-10-28, see brainstorming.md D-22 to D-36. Original note: asked in the Phase 3 question round; architect default: Expo SQLite (first-party Expo module, no extra vendor). |
| 10 | 7 | Public `/help/delete-account` page required by Google Play | decided | Added to SDD 09 page inventory and SDD 03 (deletion flow link) in the review pass. |
| 10 | 8 | EAS Build cost outside the US$30 ceiling | decided (owner) | Answered 2026-10-28, see brainstorming.md D-22 to D-36. Original note: asked in the Phase 3 question round; architect default: EAS free tier if available, else local Android builds in CI (Gradle) with no EAS. |
| 10 | 9 | SDD 03 ability names and problem code casing | decided | Fixed in the review pass (conventions). |
| 10 | 10 | Source of the device model for `deviceName` | roadmap | Resolved at implementation with the Expo Device API; not a design question. |
| 12 | 1 | No app-level meter accessor | vukem | VF-019 narrowed: `Config.Global` + upstream OTel API works; request an accessor. |
| 12 | 2 | Request-logger policy (hashed user id, truncated IP, request id = trace id) | vukem | VF-022. |
| 12 | 3 | `/metrics` endpoint needs the OTel Prometheus exporter dependency | decided (owner) | Answered 2026-10-28, see brainstorming.md D-22 to D-36. Original note: asked in the Phase 3 question round; architect default: not at launch (Railway traces only; amended 2026-10-31: no app metrics at all). |
| 12 | 4 | Railway plan decides log retention and monitors | roadmap | SDD 11 records the plan; alerts use app metrics regardless. |
| 12 | 5 | Where to host the static status page | decided | GitHub Pages on a `status` repo or `status.hexascore.app` served by the same app as a static page; choose GitHub Pages so it survives an outage of the app. |
| 12 | 6 | Metric name alignment between SDD 08 and SDD 12 | decided | SDD 12 names win; SDD 08 updated in the review pass. |
| 12 | 7 | Metric views for high-cardinality attributes | vukem | VF-019 (same entry); trigger: > 6,000 series on staging. |
| 12 | 8 | Railway monitor options and webhook delivery unverified | roadmap | Verified when the Pro plan is on (D-35 amended 2026-10-31: Railway native only). |
| 12 | 9 | Railway hard usage limit value | decided (owner) | Answered 2026-10-28, see brainstorming.md D-22 to D-36. Original note: asked in the Phase 3 question round; architect default: hard limit US$45 (1.5 × ceiling), email alert at US$24. |
| 12 | 10 | vukem.dev documentation gaps (health-check API, logger options, scheduler metric hook) | vukem | VF-023. |
| 09 | 1 | URL-prefix locale source missing in `i18n.Middleware` | vukem + decided | VF-024. App middleware resolves the locale from the path prefix before vukem's middleware. |
| 09 | 2 | ETag/304 on full HTML clashes with the per-request CSP nonce | decided | No ETag on HTML documents; ETag only on JSON (Inertia visits, APIs) and assets. TTFB budget met by the Go-side render cache of props plus SSR, not by 304s. |
| 09 | 3 | No hook to cache SSR output | vukem + roadmap | VF-025. k6 decides whether a cache in front of the SSR call is needed (TTFB p75 ≤ 400 ms). |
| 09 | 4 | `ScrollMeta` with opaque cursors unverified | roadmap | Same as 07-4. |
| 09 | 5 | `inertia.I18n` flattens keys; mobile must flatten identically | decided | A shared build step flattens the catalogues once (`packages/i18n`), both web and mobile consume the flat form. SDD 10 updated in the review pass. |
| 09 | 6 | `theme` value for `data-theme` on the root view | vukem + decided | VF-028. Workaround: a view function reading the theme cookie. |
| 09 | 7 | `vukem build` does not cover the SSR bundle | vukem + decided | VF-027. Permanent `build:ssr` step in SDD 11's pipeline. |
| 09 | 8 | Entity sitemap threshold (≥ 3) vs SDD 01 (≥ 1) | decided | ≥ 3 wins (01-8); SDD 01 FR-01-12 amended in the review pass. |
| 09 | 9 | No 300 px cover variant | decided | Same as 10-3: start with 600 px + `srcset` sizes; add 300 px only if LCP > 2.0 s on item pages. |
| 09 | 10 | New dev dependencies (Biome, Vitest, Testing Library, Playwright, axe-core, Lighthouse CI, Base UI) | decided (owner) | Answered 2026-10-28, see brainstorming.md D-22 to D-36. Original note: asked in the Phase 3 question round (one yes/no for the set; all are dev-only except Base UI). |
| 09 | 11 | INP has no lab metric | decided | Lab: TBT ≤ 200 ms as proxy; field: web-vitals beacon to SDD 12 `/app/v1/client-metrics` (opt-in, anonymous). |
| 09 | 12 | `/admin` mixes vukem's template panel with Inertia pages | decided | Accepted: the vukem admin panel stays `html/template`; custom moderation pages use the same layout via vukem views, not Inertia (consistency inside `/admin`). SDD 04 updated in the review pass. |
| 11 | 1 | Railway buckets: no versioning, no server-side encryption | decided | ADR 0014 amended: immutable hashed cover keys + 30-day deferred delete replace versioning; dumps encrypted client-side (age or vukem `crypt`) before upload. |
| 11 | 2 | Buckets are private; app serves covers and pays egress | decided | App serves covers with `Cache-Control: public, max-age=31536000, immutable`; CDN decision when cover egress > US$5/month (trigger recorded in SDD 12 cost review). |
| 11 | 3 | No migration lock in vukem docs | vukem | VF-026. App-level advisory lock around the migration runner. |
| 11 | 4 | Non-transactional migrations | vukem | VF-017 (same). |
| 11 | 5 | SSR supervision in `serve`, version in `/up`, extra ldflags | vukem | VF-027. |
| 11 | 6 | Railway Postgres version, extensions, private hostname unverified | roadmap | Closes at provisioning (first infra task). |
| 11 | 7 | `QUEUE_NAMES` priority order with the database driver | vukem | VF-018 (same). |
| 11 | 8 | Railway IaC keys, config-as-code deprecation, CLI flags unverified | roadmap | Closes on the first staging deploy. |
| 11 | 9 | Legacy `credit_ledger` balances | decided (owner) | D-23. |
| 11 | 10 | Apex domain needs CNAME flattening/ALIAS at the DNS host | decided (owner) | Answered 2026-10-28, see brainstorming.md D-22 to D-36. Original note: asked in the Phase 3 question round (which DNS provider hosts hexascore.app). |
| 11 | 11 | Dumps and database on one provider | decided (owner) | Answered 2026-10-28, see brainstorming.md D-22 to D-36. Original note: asked in the Phase 3 question round; architect default: monthly dump downloaded by the owner until an off-site bucket fits the budget. |
| 11 | 12 | GitHub Actions billing availability | decided (owner) | Answered 2026-10-28, see brainstorming.md D-22 to D-36. Original note: asked in the Phase 3 question round; fallback: pre-push hook mirror as in legacy. |
| 11 | 13 | Cost sheet averages are estimates | roadmap | Re-baselined after 30 days of Railway metrics. |
| 14 | 1 | `user_notifications` DDL differs between SDD 04 and SDD 14 | decided | SDD 14 owns the table (UUID user id, `dedupe_key`); SDD 04 §5 replaced by a cross-reference in the review pass. |
| 14 | 2 | SDD 05 §6.4 lacks library, profile, edition, notifications and reports endpoints; export must be POST | decided | SDD 05 §6.4 updated in the review pass with the SDD 14 route table; `/app/v1/me/export` is `POST` (creates an export job). |
| 14 | 3 | No item visibility/takedown column in SDD 01 | decided | Same as 08-3: `items.visibility` added to SDD 01. |
| 14 | 4 | Export: synchronous JSON (SDD 03) vs background ZIP (SDD 14) | decided | Background ZIP with signed link wins (works for large libraries); SDD 03 FR-03-16/§6.4 updated in the review pass. |
| 14 | 5 | Private profile 404 (SDD 09) vs 200 + `noindex` (SDD 14, legacy) | decided | 200 with nickname only and `noindex` (legacy parity); SDD 09 updated, plus `/library`, `/lists`, `/lists/{id}/edit` pages added. |
| 14 | 6 | Export helpers write only to `http.ResponseWriter` | vukem | VF-029. |
| 14 | 7 | Deleting a user deletes public lists others may link to | decided (owner) | D-38: anonymise public lists, delete private ones. |
| 14 | 8 | Lawful basis for showing a user's activity to followers (Q-A5 a) | owner | Social layer behind flag `social` (default off); when on, activity reaches followers only after `users.feed_opt_in` (default false). Closes with a D-xx after legal review (D-21). |
| 14 | 9 | Legacy follows (Q-A5 b) | owner | Phase 7 starts with an empty graph; legacy rows stay in the cutover archive. A restore, if chosen, inserts `accepted` rows between existing, unbanned users. |
| 14 | 10 | Markdown subset for reviews (P7.B.1.1) | roadmap | No renderer in `go.mod` and none in vukem docs [NOT IN VUKEM DOCS]; default stdlib-only subset through `html/template`; a library needs the owner's yes and an ADR. |
| 14 | 11 | Letterboxd reference (Q-A9) | owner | Research report partial; SDD 14 §12 does not depend on it. |
| 13 | 1 | Data export conflict SDD 03 vs SDD 14 | decided | SDD 03 already follows SDD 14 (review pass); closed. |
| 13 | 2 | CSRF token fallback on (SDD 03/13) vs off (SDD 09) | decided | On; SDD 09 §7.8 line updated. |
| 13 | 3 | Lint tool Biome (SDD 09, D-22) vs ESLint (SDD 11) | decided | Biome; SDD 11 pipeline table updated. |
| 13 | 4 | Personalisation consent vs D-29 | decided | D-29 stands (legitimate interest + opt-out); confirmed at the D-21 legal review. |
| 13 | 5 | LGPD article numbers and ANPD deadline unverified | roadmap | Legal review task before cutover verifies article numbers and the breach-notification deadline against the official text. |
| 13 | 6 | DPO (encarregado) name/contact | decided (owner) | D-39: the owner, `privacy@hexascore.app`. |
| 13 | 7 | Admin mandatory 2FA, shorter session, IP allowlist not in vukem | vukem | VF-030; app middleware meanwhile. |
| 13 | 8 | No command to re-encrypt columns after `APP_KEY` rotation | vukem | VF-031; rotation runbook keeps the previous key readable until a one-off re-encrypt job finishes. |
| 13 | 9 | No antivirus for uploads | decided | Accepted with compensating controls (images only, sniffing, re-encode, quarantine until approval). |
| 13 | 10 | HSTS `preload` deferred | decided | Revisit 2027-06. |
| 13 | 11 | Restore must re-apply post-backup account deletions | decided | Note added to SDD 11 §7.5. |
| 13 | 12 | Turnstile in production needs an ADR 0016 update and a privacy line | decided | Recorded in ADR 0016 consequences already; privacy-policy line added when the flag is first enabled. |
