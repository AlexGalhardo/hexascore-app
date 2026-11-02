# Data sources matrix

Status: architect audit 2026-11-02. Complements SDD 06 §7.1 (sources already connected) and ADR 0006 (licensing posture: the paid API exposes only Wikidata, Wikipedia, Open Library and HexaScore-own fields; TMDB and IGDB are display-only). Terms were read on 2026-11-02 by the research pass; rows marked "not verified" could not be read (blocked pages) and need an owner read before the source is enabled. Legal reads already cleared by the owner (Q-A16, 2026-11-01): IGDB/Twitch, Open Library, Steam, GOG.

Decision legend: **use** · **use with caveats** (conditions listed) · **do not use**.

## Films and series

| Source | Access | Licence / ToS | Rate limits and cost | Fields | Reliability | Attribution | Decision |
| --- | --- | --- | --- | --- | --- | --- | --- |
| TMDB | REST API with key (connector exists) | No commercial use without a written agreement; cache: the "≤ 6 months" figure is not stated on the terms page read 2026-11-02 (purge on termination only); no redistribution (themoviedb.org/api-terms-of-use, 2026-11-02) | ~40 req/s practical; free | poster, facts, trailers, external IDs, watch providers (`/watch/providers`, JustWatch data) | high | notice "…uses TMDB and the TMDB APIs but is not endorsed, certified, or otherwise approved by TMDB" + TMDB logo (terms §3, 2026-11-02); watch providers also need the JustWatch attribution (not stated, see Attribution rules) | **use with caveats**: display-only (ADR 0006), never in `/api/v1`; `tmdb.expire` keeps the 6-month rule; watch providers only with both attributions (RFC-003) |
| IMDb non-commercial datasets | TSV dumps | Personal, non-commercial only; commercial via AWS Data Exchange (listed US$150k/year, ADR 0006) | free | ratings, titles, crew | high | — | **do not use** (ADR 0006 §5) |
| Wikidata | API + weekly dumps (connector exists) | CC0 | 1 query/s SPARQL, 5/s entity API, maxlag, identified User-Agent; free | QID, titles, dates, credits, external IDs, posters P3383, trailers P1651, sitelinks | medium (coverage gaps on runtime, platforms) | none required | **use** (primary key for entity resolution) |
| Wikipedia en/pt/es | API (connector exists) | CC BY-SA 4.0; API etiquette (serial requests, User-Agent policy) | 5/s per language; free | synopses, entity descriptions | high | "From Wikipedia, CC BY-SA 4.0" + link per text | **use** (already in SDD 06) |
| JustWatch | no public or partner API found | Terms forbid scraping and commercial use (support.justwatch.com terms, 2026-11-02) | — | availability per country | — | — | **do not use** directly; its data arrives through TMDB watch providers under TMDB's terms |
| Watchmode | REST API with key | Paid tiers allow commercial use; Developer tier non-commercial with attribution; caching: paid plans keep data while subscribed; free plan must refresh or delete within 30 days, "should refresh or delete cached data within 30 days" (FAQ <https://api.watchmode.com/>, read 2026-11-02; the terms page <https://watchmode.com/tc> returned 404, so the T&C wording stays unverified) | Free 2,500 credits/month; Startup US$349/month (40k); Business US$599/month (100k) | streaming availability, deep links, prices per country | high (claimed) | required on the free tier | **use with caveats**: only if the owner funds it; outside the US$20 infra cap, so RFC-003 starts with TMDB providers and keeps Watchmode as the paid upgrade |
| Rotten Tomatoes / Metacritic | licensed products only (Fabric Origin) | no public API; licence required | ~US$60k/year (secondary source, not verified) | critic scores | — | link-back | **do not use** (ADR 0006 §4) |

### Attribution rules (TMDB and JustWatch, read 2026-11-02)

Each rule: source, date read, quote of at most 15 words. "Not stated" means the page read did not say it; nothing is inferred.

- TMDB notice text: "is not endorsed, certified, or otherwise approved by TMDB." (<https://www.themoviedb.org/api-terms-of-use>, §3, 2026-11-02). The notice names the product in brackets and starts "This [website, program, service, application, product] uses TMDB".
- TMDB logo required: "You must use the TMDB logo to identify Your use of TMDB" (same URL, §3, 2026-11-02). Logo is "less prominent" than the app's own mark: "must be less prominent than the logos or marks" (same URL, §3).
- Notice placement: "must place the following notice prominently" (same URL, §3). Minimum logo size: not stated on <https://www.themoviedb.org/api-terms-of-use> or <https://developer.themoviedb.org/docs/faq> (2026-11-02). The logos page <https://www.themoviedb.org/about/logos-attribution> was not read.
- Logo modification (FAQ, <https://developer.themoviedb.org/docs/faq>, 2026-11-02): "Our logo should not be modified in color, aspect ratio, flipped or rotated" (13 words).
- Link target (same FAQ, 2026-11-02): "please point your link to:" followed by <https://www.themoviedb.org>.
- Attribution condition (same FAQ, 2026-11-02): "as long as you attribute TMDB as the source of the data" (11 words), applies to non-commercial use.
- Commercial use: "Charging users a fee for Your Application" is listed as commercial use (same URL, §2, 2026-11-02); commercial use needs a written agreement.
- Caching limit: "cache ≤ 6 months" is not stated on <https://www.themoviedb.org/api-terms-of-use> (2026-11-02). The terms say on termination: "promptly delete or otherwise purge all TMDB Content, including any cached content" (same URL, §1.D).
- Watch-provider endpoint: the endpoint page <https://developer.themoviedb.org/reference/watch-providers-movie> returned 404 on 2026-11-02, so its attribution wording is not stated.
- JustWatch attribution for TMDB-sourced data: not stated on <https://www.justwatch.com/us/terms> (updated 5 May 2026, read 2026-11-02) or on <https://www.justwatch.com/> (2026-11-02). Logo rules for JustWatch: not stated on those pages.

## Games

| Source | Access | Licence / ToS | Rate limits and cost | Fields | Reliability | Attribution | Decision |
| --- | --- | --- | --- | --- | --- | --- | --- |
| IGDB (Twitch) | REST API with Twitch OAuth (connector exists) | Twitch Developer Services Agreement; commercial use through partnership; visible attribution; interim rule D-36 (site-only, refresh 24 h, never cached > 7 days); api-docs.igdb.com readable on 2026-11-02 (403 in the earlier pass): "The API is free for both non-commercial and commercial projects." (<https://api-docs.igdb.com/>, FAQ); rate limit 4 req/s re-read; the Twitch Developer Services Agreement text itself was not re-read | 4 req/s, 8 in flight; free | cover, facts, platforms, trailers, companies | high | "Game data by IGDB.com" on game pages and `/about` | **use with caveats**: display-only, D-36 refresh job, token cached per run (today it is re-fetched per call, `internal/ingest/igdb/igdb.go` ~328) |
| Steam store `appdetails` | undocumented store endpoint (connector exists); Steam Web API (key, 100k/day) does not cover store data | Steam Web API Terms of Use: data "as is", no implied endorsement; store endpoint has no published terms | ~200 req / 5 min community figure (not official); free | header image, OS platforms, price, store link | high | — | **use with caveats**: 1 request per 1.6 s as in SDD 06; images site-only; price only for the "where to buy" link, never stored as authoritative |
| GOG catalog | public catalog JSON (connector exists) | no published API terms; GOG User Agreement governs the site; affiliate programme by e-mail (`affiliate@gog.com`; page 404 on 2026-11-02) | 2/s; free | store link, cover, OS platforms | medium | — | **use with caveats**: links and platforms; images site-only until the owner records the read |
| HowLongToBeat | no official API; robots.txt blocked the research pass | not verified (terms URLs <https://www.howlongtobeat.com/terms> and <https://howlongtobeat.com/terms> returned 404 on 2026-11-02) | — | time to beat | medium | — | **do not use** until the owner reads the terms; HexaScore has its own `time_to_beat` from users |
| Epic Games Store, itch.io | Support-A-Creator codes (Epic), Partners affiliate codes (itch.io) | affiliate programme terms; Epic requires clear disclosure | free | store links | high | disclosure | **use with caveats**: links from Wikidata store IDs; codes only after enrolment (RFC-003) |
| Google Play / App Store | store page links + official badges | badge guidelines (no recolouring or scaling; one per layout) | free | store link | high | official badge | **use** |

## Books

| Source | Access | Licence / ToS | Rate limits and cost | Fields | Reliability | Attribution | Decision |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Open Library | API + monthly dumps + covers API (connector exists, flag `ingest_openlibrary`) | Internet Archive asserts no rights over the data (`docs/legal/open-library.md`, 2026-11-01); bulk via API is forbidden, use dumps; covers: no explicit grant, hot-linking preferred | 1 req/s unidentified, 3/s with User-Agent; free | works, editions, ISBNs, authors, covers | medium | — | **use**: dumps for the 1 M bootstrap, API for lookups; covers not API-eligible (open issue 01-2) |
| Google Books | API with key | Terms forbid charging users for an application that uses the data ("You may not charge users any fee", developers.google.com/books/terms, 2026-11-02); default quota 1,000 req/day (console, not verified: no quota figure on the "Using the API" page, <https://developers.google.com/books/docs/v1/using>, read 2026-11-02) | free | ISBN lookup, covers, descriptions | high | required display rules | **do not use** while HexaScore sells API plans; revisit only with a written agreement |
| ISBNdb | API with key | site returned 403; pricing from third parties US$14.99 to US$299.99/month, 1k to 50k req/day | paid | ISBN metadata, covers | medium | not verified | **do not use** for now (cost, unverified terms); Open Library + Wikidata cover the edition model (ADR 0012) |
| Goodreads / Skoob | Goodreads API retired (no new keys since 2020-12-08); Skoob has no API | scraping only, against ToS | — | ratings, shelves | — | — | **do not use** (product references only) |
| Wikidata / Wikipedia | as above | CC0 / CC BY-SA | — | works (P50 author, P212/P957 ISBN, P747 editions) | medium | as above | **use** |

## Cross-cutting

| Source | Access | Licence / ToS | Limits and cost | Decision |
| --- | --- | --- | --- | --- |
| YouTube oEmbed (connector exists) and Data API | oEmbed (no key); Data API 10,000 units/day, `search.list` costs 100 units | embed terms; do not cache unauthorised data > 30 days | free | **use with caveats**: oEmbed for the embed verdict (SDD 06 §7.5); Data API only for the "is this the official trailer" check in RFC-001 with a daily cap |
| Wikimedia Commons | per-file licence (P3383 posters) | per file | 5/s, 2/s uploads host | **use with caveats**: store author, licence and URL per file (SDD 06) |
| Amazon Associates (BR, US) | affiliate links; Creators API replaces PA-API 5 (deprecation notice read 2026-11-02) | Operating Agreement: disclosure text "As an Amazon Associate I earn from qualifying purchases."; prices only from the API, ≤ 24 h old with timestamp; sales threshold for API access not verified (BR portal did not resolve) | free | **use with caveats**: links with tag and visible disclosure (RFC-003); no price display until the Creators API is approved |
| LLM extraction (RFC-001 last resort) | provider API | provider terms; outputs validated against schema | hard monthly ceiling (OQ-05) | **use with caveats**: only for normalisation of fetched text, never as a data source of record |

## Re-read on 2026-11-02 (owner read needed before enabling)

Resolved by the re-read: IGDB API terms (FAQ readable; "free for both non-commercial and commercial projects", commercial use via the partner programme) and the Watchmode caching clause (FAQ: paid plans keep data while subscribed, free plan refreshes or deletes within 30 days).

Still not verified, with the reason in each row: ISBNdb pricing and terms (site 403), Rotten Tomatoes licensing price (secondary source only), HowLongToBeat terms (both URLs 404), IMDb via AWS Data Exchange (not re-read), Amazon Associates BR portal and the PA-API sales threshold (portal did not resolve), GOG affiliate terms (page 404), Google Books default quota (not on the "Using the API" page). Open Library: the licence page <https://openlibrary.org/terms> returned 404 on 2026-11-02; the owner clearance in Q-A16 (2026-11-01) stands, and no new quote was recorded.
