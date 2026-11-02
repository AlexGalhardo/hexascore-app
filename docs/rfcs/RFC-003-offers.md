# RFC-003: Where to watch, where to buy (offers and availability)

Status: proposed (architect audit, 2026-11-02). Owner gates: OQ-08 (affiliate accounts), OQ-09 (TMDB watch providers posture). Builds on SDD 01 (`affiliate_links`, `item_links`, flag `affiliates`), ADR 0006 (licensing), `docs/data/SOURCES.md`, RFC-001 (worker, provenance).

## 0. Executive summary

Today the only "where to get it" data is `affiliate_links(item_id, store, url, active)` filled by admins, plus store IDs from Wikidata in `item_links`. This RFC adds an `offers` model keyed by item × country × store × kind, fed by (1) TMDB watch providers (JustWatch data, display-only, both attributions) for films and series, (2) store IDs already in the catalogue (Steam, GOG, Epic, itch.io, Google Play, App Store) for games, and (3) Amazon links by ISBN for books, with affiliate tags only where a programme account exists, a visible disclosure, `rel="sponsored nofollow"`, click tracking without personal data, a revalidation schedule run by the catalogue worker, broken-link detection and a fallback card when nothing is known. Paid APIs (Watchmode) are optional later; prices are not displayed until a licensed price feed exists.

## 1. Scope

| Category | Kinds | Sources | Country scope |
| --- | --- | --- | --- |
| Films, series | `stream` (subscription), `rent`, `buy`, `free` (ads), `cinema` (future) | TMDB `/movie/{id}/watch/providers`, `/tv/{id}/watch/providers` (per country, with `link` to the JustWatch page); optional Watchmode | BR, US, PT, ES, MX, AR first (the product's three languages), configurable list |
| Games | `buy_digital` (Steam, GOG, Epic, itch.io, Nintendo, PlayStation, Xbox store pages), `mobile` (Google Play, App Store), `physical` (Amazon) | Wikidata store IDs (P1733 Steam, P2725 GOG, P6278 Epic, P7294 itch.io, P3418 Google Play, P3861 App Store, P8084 Nintendo eShop, …), Steam `appdetails` for OS and free-to-play flag | global for digital stores; Amazon per marketplace |
| Books | `buy` (Amazon per marketplace), `read` (Open Library / Internet Archive when a readable edition exists), later publisher and local stores | ISBN-13 from `editions`, Open Library `ocaid` | BR (amazon.com.br), US (amazon.com), ES, PT later |

Out of scope: price comparison, stock, deep links into apps, cinema showtimes, any scraping of JustWatch, Steam top sellers or retailer pages.

## 2. Data model

```sql
CREATE TABLE stores (
  code text PRIMARY KEY,              -- 'netflix', 'steam', 'amazon_br', 'google_play'…
  name text NOT NULL,
  kind text NOT NULL CHECK (kind IN ('streaming','digital_store','mobile_store','retailer','library')),
  url_template text,                  -- '{id}' or '{isbn13}' placeholders; NULL when the source gives full URLs
  affiliate_program text,             -- 'amazon_associates', 'steam_curator', NULL
  affiliate_param text,               -- 'tag' for Amazon; NULL when the programme uses codes elsewhere
  logo_key text,                      -- bucket key; provider logos from TMDB are display-only with attribution
  attribution text,
  active boolean NOT NULL DEFAULT true
);

CREATE TABLE offers (
  id uuid PRIMARY KEY,
  item_id uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  edition_id uuid REFERENCES editions(id) ON DELETE CASCADE,   -- books: one offer per edition
  country char(2) NOT NULL,
  store text NOT NULL REFERENCES stores(code),
  kind text NOT NULL CHECK (kind IN ('stream','rent','buy','free','buy_digital','mobile','physical','read')),
  url text NOT NULL,                  -- canonical URL without affiliate params
  external_id text,                   -- store id (steam app id, ASIN, isbn13)
  source text NOT NULL,               -- 'tmdb_providers', 'wikidata', 'editions', 'admin', 'watchmode'
  licence text NOT NULL,              -- from data_sources
  confidence smallint NOT NULL DEFAULT 80,
  first_seen timestamptz NOT NULL DEFAULT now(),
  last_verified_at timestamptz,
  verify_status text NOT NULL DEFAULT 'unverified' CHECK (verify_status IN ('unverified','ok','redirected','broken','gone')),
  verify_http_status int,
  expires_at timestamptz,             -- TMDB data: last_verified_at + 180 d (ADR 0006 cache rule)
  active boolean NOT NULL DEFAULT true,
  UNIQUE (item_id, edition_id, country, store, kind)
);
CREATE INDEX offers_item_country_idx ON offers (item_id, country) WHERE active;
CREATE INDEX offers_verify_idx ON offers (last_verified_at NULLS FIRST) WHERE active;

CREATE TABLE offer_clicks (           -- aggregate, no user id, no IP (LGPD)
  day date NOT NULL,
  offer_id uuid NOT NULL REFERENCES offers(id) ON DELETE CASCADE,
  country char(2) NOT NULL,
  clicks int NOT NULL DEFAULT 0,
  PRIMARY KEY (day, offer_id, country)
);
```

`affiliate_links` is migrated into `offers` with `source='admin'` and kept read-only for two releases (expand/contract). Affiliate tags are never stored in `url`; they are appended at render time from `stores.affiliate_param` and the environment (`AFFILIATE_AMAZON_BR_TAG`, `AFFILIATE_AMAZON_US_TAG`, …) so changing a tag is a config change and the stored data stays clean for the API.

## 3. Collection and revalidation (catalogue worker, RFC-001)

- **Films and series**: stage `offers.tmdb` in the TMDB chain: for items with a TMDB id, `GET /{type}/{id}/watch/providers`, one request per item, keep `flatrate`→`stream`, `rent`, `buy`, `free`/`ads`→`free` for the configured countries; `url` = the JustWatch `link` TMDB returns (per country); `expires_at = now + 180 d`; provider logos are not copied, the provider name is shown (TMDB logo rights are display-only and the JustWatch attribution is mandatory). Budget: 100 k items ≈ 100 k requests ≈ 1.4 h at 20/s; refresh every 30 days by `expires_at`/`last_verified_at` order, newest releases first (availability changes fastest in the first 90 days).
- **Games**: stage `offers.stores` derives offers from `item_links`/`external_ids` with `stores.url_template` (no network); Steam `appdetails` (already fetched) adds `is_free` and the platforms; Google Play and App Store IDs from Wikidata become `mobile` offers with the official badges.
- **Books**: stage `offers.books` builds Amazon search-by-ISBN URLs (`https://www.amazon.com.br/s?k={isbn13}`, or `/dp/{asin}` when an ASIN is known from Wikidata P5749) per marketplace, and `read` offers from Open Library `ocaid`.
- **Revalidation**: job `offers.verify` daily at 02:45 verifies up to 5,000 offers (HEAD, follow ≤ 3 redirects, 10 s timeout, identified User-Agent, 1 req/s per host; retailers that block HEAD get a GET with a 64 KB body cap): 200 → `ok`, 301/302 to another path → `redirected` (store the final URL as a candidate), 404/410 → `gone` (deactivated after 2 consecutive runs), 5xx/timeout → unchanged, retry next day; Amazon is not probed (bot blocking; links are validated by ISBN format only) and TMDB-sourced offers are refreshed from the API instead of probed.
- **Fallback**: when an item has no active offer for the visitor's country, the UI shows "No official source known in {country}" with the JustWatch search link for films/series (allowed: it is a plain link) and the Wikidata-derived official website when present.

## 4. Display, affiliates and compliance

- Section "Where to watch" / "Where to buy" on the item page, grouped by kind, country from the locale (`pt-BR`→BR, `es`→ES then MX/AR, `en`→US) with a country switch; each entry: store name, kind label, "verified {date}" tooltip, `rel="sponsored nofollow noopener"`, `target=_blank`.
- Affiliate disclosure: a visible line above the section in the three locales ("HexaScore may earn a commission from purchases through these links. As an Amazon Associate we earn from qualifying purchases." where Amazon links appear) plus a paragraph in `/terms` and `/privacy`; the exact Amazon sentence is mandatory in the Operating Agreement (research 2026-11-02).
- Programme rules encoded in `stores`: Amazon (no price display without the Creators API, links only on pages with content, no link cloaking, no e-mail links, tag per marketplace), Steam (store links are plain; the Curator programme is opt-in and not needed for links), Epic (Support-A-Creator code only after enrolment, disclosure required), itch.io (partner code), GOG (affiliate by e-mail enrolment). Until OQ-08 is answered every link is plain (no tag), which is compliant by construction.
- LGPD: click tracking increments `offer_clicks` by day and country only; no cookies, no user id, no IP; the outbound click goes through `GET /go/{offer_id}` (302, rate-limited per IP, no logging of the referrer beyond the request log already redacted). The privacy policy lists the aggregate counter.
- Paid API: offers are **not** exposed in `/api/v1` (TMDB/JustWatch terms); a future "official store links" field may expose only `wikidata`-sourced store URLs, which are CC0 identifiers.

## 5. Tests and acceptance

- Unit: URL templating per store, affiliate tag appended at render only, disclosure rendered when any Amazon offer is present, country resolution from locale, `rel` attributes.
- Integration (pgtest): migration of `affiliate_links` → `offers`; `offers.verify` state machine with a fake HTTP server (200/301/404/500/timeout); TMDB providers fixture → offers with `expires_at`; expiry job deactivates TMDB offers older than 180 days.
- E2E: item page shows the section, fallback text when empty, `/go/{id}` redirects and counts once per request.
- Security: `TestSecurity_A10_OfferRedirectOnlyKnownURLs` (no open redirect: `/go/{id}` only follows stored URLs), rate limit on `/go`.

## 6. Backlog

| ID | Task | Est. (senior-days) | Depends | Acceptance |
| --- | --- | --- | --- | --- |
| OF-01 | Migrations `stores`, `offers`, `offer_clicks`; seed of stores; `affiliate_links` migration with expand/contract plan in SDD 01 | 1.5 | — | migrate round trip in CI; existing admin links appear as offers |
| OF-02 | `offers.stores` and `offers.books` derivation stages (no network) in the worker | 1.5 | RFC-001 CW-01 | 100 % of items with a Steam/GOG/Play/App Store id get offers; books with ISBN get Amazon BR/US |
| OF-03 | `offers.tmdb` stage with per-country providers, 180-day expiry, JustWatch + TMDB attribution strings in `stores` | 2 | OQ-09 | fixture test; staging run on 1,000 films within budget |
| OF-04 | `offers.verify` job with the state machine and `verify_status` | 1.5 | OF-01 | fake-server tests; digest reports broken/gone counts |
| OF-05 | Item page section + country switch + fallback + disclosure (3 locales) + `/go/{id}` redirect with `offer_clicks` | 2.5 | OF-01 | e2e and Lighthouse unchanged; a11y of the switch |
| OF-06 (Intern) | Affiliate programme checklist per store in `docs/legal/affiliates.md` with quotes and dates; `/terms` and `/privacy` paragraphs | 1 | OQ-08 | document reviewed by the owner |
| OF-07 | Admin: offer list per item with activate/deactivate and manual add (reuses the item admin page) | 1 | OF-01 | admin action audited in `admin_audit_logs` |

Rollback: the section is behind the existing `affiliates` flag (SDD 01); offers data is additive; `affiliate_links` stays until the contract migration two releases later.
