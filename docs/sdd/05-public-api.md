# SDD 05: Public API, billing and first-party app API

Status: reviewed (2026-10-28 consistency pass). Decisions: D-03, D-06, D-12, D-13, D-14, D-19, D-20, A-01. ADRs: 0003, 0006, 0009, 0015. Related SDDs: 01, 02, 03, 04, 06, 07, 08, 10, 11, 12.

## 1. Objective

Sell a fast, legally clean, self-serve read API over the catalogue and scores (D-14) and give the React Native app a separate token-scoped write surface (A-01). Developers get one-click keys, one score scale (0–10, D-13), per-field `source`/`licence` (D-03), stable cursors, problem+json, published limits with headers, a generated OpenAPI 3.1 document, TypeScript and Python SDKs, and three freshness tools (web-research 2.2, 3.2, 9.5): daily ID exports (TMDB), a 14-day changes feed (TMDB) and signed webhooks (IGDB). Revenue: Stripe subscriptions (ADR 0015).

## 2. Scope

In scope (v1): paid read API `/api/v1` on `api.hexascore.app` (D-19) with legacy parity plus trending, entities, editions, changes, exports, webhooks; keys, plans, throttles, Postgres metering (ADR 0003); Stripe via `install:billing` + `contrib/stripe` in BRL/USD/EUR (ADR 0015); `/app/v1` for D-12 features (ADR 0009); OpenAPI, Scalar, docs, SDKs, contract and load tests.

Out of scope: PIX/AbacatePay and credit packs (D-06, D-14; boundary in 7.6); per-locale OpenAPI documents (D-06); writes in `/api/v1`; TMDB/IGDB fields in `/api/v1` (D-03); billing, keys, admin in the app (D-12).

## 3. Functional requirements

- FR-05-1: `/api/v1` requires `Authorization: Bearer hx_live_<32 base62>` (legacy `lib/api/keys.ts`); missing/unknown/revoked → 401 with `WWW-Authenticate: Bearer`. Verify: integration test.
- FR-05-2: Keys are shown once, stored as SHA-256 hex, listed with `prefix` (14 chars), `name`, `lastUsedAt` (≤ 1 write/min), `revokedAt`; ≤ 10 live keys per user. Verify: integration test.
- FR-05-3: Limits equal `legacy/apps/web/lib/api/plans.ts` (kept at launch, reviewed 90 days after cutover, D-33): free 100 lifetime, `limit` ≤ 10, 1 rps; pro 50,000/month, ≤ 100, 10 rps; business 500,000/month, ≤ 100, 50 rps; UTC calendar months. Legacy `credit_ledger` balances are honoured as extra quota consumed before the plan quota, no refund flow (D-23, 7.2). Verify: table-driven test; credit-before-plan test.
- FR-05-4: Counting per `legacy/docs/api.md`: every authenticated request reaching a handler counts (id 404 and 304 included); 401, query 400, unknown route 404, 429, 5xx do not (5xx refunded). Verify: test per status.
- FR-05-5: Exhausted quota → 402 `quota-exhausted` with `plan`, `limit`; throttle → 429 with `Retry-After`. Verify: quota test.
- FR-05-6: Only fields whose source is in `CATALOG_API_SOURCES` are emitted (SDD 01 FR-01-9), with `provenance` and `meta.omittedFields`. Verify: contract test with TMDB/IGDB fixtures.
- FR-05-7: Score fields are exactly SDD 02 §6 (`score`, `scoreByCategory`, `ratingCount`, `confidence`, `formulaVersion`), 0–10, one decimal, `null` below 5 ratings. Verify: Schemathesis + golden test.
- FR-05-8: Detail and list responses carry a strong `ETag`; matching `If-None-Match` → 304. Verify: integration test.
- FR-05-9: `/v1/changes?since=` accepts ≤ 14 days back, else 400 `changes-window-exceeded`. Verify: integration test.
- FR-05-10: A daily ID export per type exists by 08:00 UTC, kept 90 days. Verify: job test, fake clock + S3 fake.
- FR-05-11: Pro (2 endpoints) and business (10) can register webhooks for `item.created|updated|deleted`; deliveries are signed and retried. Verify: delivery test.
- FR-05-12: Only Stripe webhooks set the effective plan (7.5). Verify: billing tests with `stripe.SignForTest`.
- FR-05-13: `/app/v1` accepts only `hx_app_` tokens with the route ability (SDD 03), is not plan-metered, and is rate-limited per token; an API key there → 401, an app token on `/api/v1` → 401. Verify: integration test.
- FR-05-14: Builds below `APP_MIN_SUPPORTED_VERSION` get 426 `app-upgrade-required`. Verify: integration test.
- FR-05-15: CI fails on stale OpenAPI documents, SDKs or app client. Verify: CI job.
- FR-05-16: `lookupItem` accepts third-party ids as input but returns only allowlisted fields and never echoes an external id whose source is restricted (TMDB, IGDB, Steam, GOG); Wikidata-sourced ids are returned (open issue 05-3, ADR 0006). Verify: contract test with a TMDB-id lookup asserting no `tmdb` key in the body.
- FR-05-17: `GET /api/v1/{type}/{id}/activity` (SDD 08) is business plan only (else 403 `plan-required`) and counts as 1 request (open issue 08-6). Verify: integration test per plan.
- FR-05-18: `POST /app/v1/client-errors` and `POST /app/v1/client-metrics` accept opt-in, anonymous reports (no PII), are rate-limited and retained 7 days (open issues 10-2, 09-11; SDD 12 owns processing). Verify: integration test (429 over limit, PII fields rejected, prune after 7 days with a fake clock).

## 4. Non-functional requirements

- NFR-05-1 Latency, one instance (D-08), 500k items, 50 rps for 10 min: list p95 ≤ 300 ms, detail p95 ≤ 150 ms, search p95 ≤ 300 ms, errors < 0.1%; auth + quota overhead p95 ≤ 5 ms (no DB write on the request path). k6 `qa/k6/api-v1.js`.
- NFR-05-2 Quota accuracy: a crash loses ≤ 60 s of counters, in the customer's favour (ADR 0003).
- NFR-05-3 Availability: Postgres down → 503 `Retry-After: 30` (fail closed, legacy parity); a revoked key is served ≤ 60 s.
- NFR-05-4 Security: keys never logged; webhook URLs SSRF-safe (`webhook.Sender` over `httpclient.Safe`, vukem `https://vukem.dev/docs/security/encryption` "Building blocks"); webhook secrets encrypted with `app.Crypt` (`https://vukem.dev/docs/digging-deeper/webhooks` "Encryption"); ZAP baseline.
- NFR-05-5 Contract: every route is an `api.Register` typed route with `Op.ValidateResponse: true` (`https://vukem.dev/docs/rest-apis/openapi` "OpenAPI").
- NFR-05-6 Cost: no Redis (ADR 0003); exports ≤ 50 MB/day gzipped.
- NFR-05-7 i18n: one English document (D-06); `locale=en|pt-BR|es` selects catalogue text.

## 5. Data model

Legacy source: `legacy/apps/web/lib/db/schema/billing.ts` (`api_keys`, `subscriptions`, `api_usage_daily`, `stripe_customers` migrate per A-04 (`processed_events` does not, see below); `credit_ledger` balances become `api_credit_balances` rows (D-23); `credit_ledger` and `pix_charges` are also copied verbatim for the second release). Id types follow the README: `users.id` is `text` (SDD 03 §5), item ids are `uuid`.

```sql
-- Own table, not vukem access_tokens: legacy keys are sha256(full key), not the
-- vukem <prefix><id>.<secret> shape, and A-04 migrates them unchanged.
create table api_keys (
  id uuid primary key,
  user_id text not null references users(id) on delete cascade,
  name text not null check (length(name) between 1 and 40),
  prefix text not null, hash char(64) not null unique,
  last_used_at timestamptz, revoked_at timestamptz,
  created_at timestamptz not null default now());
create index on api_keys(user_id) where revoked_at is null;

create table api_credit_balances (          -- D-23: legacy credits, consumed before the plan quota
  user_id text primary key references users(id) on delete cascade,
  remaining integer not null check (remaining >= 0),
  migrated_at timestamptz not null default now());

create table api_quota_counters (           -- period: 'life' | 'YYYYMM'
  user_id text not null references users(id) on delete cascade,
  period text not null, used integer not null check (used >= 0),
  primary key (user_id, period));
create table api_usage_daily (              -- kept 400 days
  user_id text not null references users(id) on delete cascade,
  day date not null, count integer not null, refunded integer not null default 0,
  primary key (user_id, day));

create table api_changes (                  -- kept 30 days
  id bigserial primary key, item_id uuid not null, item_type text not null,
  action text not null check (action in ('created','updated','deleted')),
  fields text[] not null default '{}', merged_into uuid,
  changed_at timestamptz not null default now());
create index on api_changes(changed_at, id);

create table webhook_endpoints (
  id uuid primary key,
  user_id text not null references users(id) on delete cascade,
  url text not null check (url like 'https://%'),
  events text[] not null, item_types text[] not null,
  secret_enc text not null, disabled_at timestamptz,
  created_at timestamptz not null default now());
-- Attempts: webhook.SQLDeliveryLog (https://vukem.dev/docs/digging-deeper/webhooks "Webhooks") in `webhook_deliveries`;
-- rows older than 30 days are pruned daily at 04:40 UTC by `s.Job(jobs.PruneWebhookDeliveries{})`. Fan-out position:
create table webhook_cursor (id smallint primary key check (id = 1), last_change_id bigint not null);

alter table subscriptions                   -- install:billing table (https://vukem.dev/docs/digging-deeper/billing "What it adds")
  add column plan text not null check (plan in ('pro','business')),
  add column provider text not null default 'stripe' check (provider in ('stripe','abacatepay')),
  add column state text not null check (state in ('active','past_due','canceled','expired'));
create unique index on subscriptions(user_id) where state in ('active','past_due');
```

Exports are listed from the bucket prefix (no table). There is no `processed_events` table (orchestrator decision 2026-10-29, P2.F.2.7): Stripe webhook dedup is vukem `contrib/stripe`'s cache-backed idempotency by event id, under two conditions: production uses a shared, persistent cache store (`CACHE_STORE=database`, SDD 11 §6.4) so dedup survives restarts and the cutover window, and every billing webhook handler stays idempotent by Stripe object id (it re-reads the subscription from Stripe and upserts by `stripe_subscription_id`). Proof: `TestSecurity_A08_StripeReplayIgnored` (`tests/billing_test.go`).

Legacy import (P2.F.2.7, `database/legacyimport/billing.go`): remaining credits per user = `greatest(sum(credit_ledger.delta), 0)`, the legacy `creditBalance` formula (`legacy/apps/web/lib/billing/subscriptions.ts`); debits still in the legacy Redis cache and not yet flushed to the ledger are forgiven, as legacy does when that key is lost. Subscriptions with plan `free` are not imported; the others keep `plan`, `provider` and `status` (as `status` and `state`), with `provider_ref` as `stripe_subscription_id`. `credit_ledger` and `pix_charges` are copied into tables of the legacy shape (migration `20261029200000_create_legacy_billing_tables`) that nothing in v1 writes. `processed_events` is not imported: the app has no such table, Stripe webhook ids are deduplicated by vukem `contrib/stripe` in the cache store, and a replayed legacy event fails the 5-minute signature window first (orchestrator decision 2026-10-29; conditions above).

## 6. Contracts

### 6.1 Host, stack, headers

`api.hexascore.app/v1/*` maps to the `/api/v1` group (legacy `apps/web/proxy.ts`). Host routing is [NOT IN VUKEM DOCS] (VF-015); v1 uses an `r.UseServer` middleware (shown in `https://vukem.dev/docs/digging-deeper/billing` "Routes") rewriting `/v1/` → `/api/v1/` when `Host` matches the API host; OpenAPI `servers` lists `https://api.hexascore.app` and the documented paths keep the `/api/v1` prefix, because vukem cannot strip a path prefix from the exported document (VF-050, 2026-10-29); both `/v1/...` and `/api/v1/...` answer on the API host. Route tables below write paths in their public form (`/v1/...` = `/api/v1/...` in the router).

Stack: `middleware.JSONOnly()`, CORS `*` without credentials, IP limit 120/min (`https://vukem.dev/docs/security/api-tokens` "Routing"), `auth.Required(apiKeyGuard)` (custom guard implementing `auth.APIGuard` + `auth.FailClosedChecker`, `https://vukem.dev/docs/security/rate-limiting` "Middleware", "JWT guard"), `NotBanned` (SDD 03), plan throttle `ratelimit.Middleware(limiters("api-<plan>", rps, time.Second), ratelimit.ByUser(apiKeyGuard))` (`https://vukem.dev/docs/security/rate-limiting` "Rate limiting"), quota meter, handler.

Headers: `X-Request-Id`; `RateLimit-Limit/Remaining/Reset` = per-second throttle (vukem-native, `https://vukem.dev/docs/rest-apis/client-generation` "Cross-cutting"); `X-Quota-Limit`, `X-Quota-Remaining`, `X-Quota-Reset` (seconds, `0` for lifetime); `ETag`; `Cache-Control: private, no-cache` (legacy `no-store` blocked 304); `Deprecation`/`Sunset`/`Link` via `api.Deprecate`. Casing: payload fields camelCase, vukem envelope keys (`next_cursor`, `request_id`) as vukem emits them (README convention, VF-014). Legacy `RateLimit-*` carried the quota; its SDKs were never published (`legacy/docs/api.md` "SDKs"), so nobody depends on it.

### 6.2 `/api/v1` routes (security `apiKey`)

| operationId | Method, path | Notes |
| --- | --- | --- |
| `listItems` | GET `/v1/{type}` | games/movies/series/books; `genre`, `platform` (games), `year` 1800–2100, `min_score`/`max_score` 0–10 step 0.5, `min_<category>` (A-06), legacy `sort`, `limit` 1..cap (default 10), `cursor`, `locale` |
| `getItem` | GET `/v1/{type}/{id}` | 301 for merged ids (SDD 01); non-`public` items 404 (SDD 01 FR-01-14) |
| `getItemActivity` | GET `/v1/{type}/{id}/activity` | business plan only (else 403 `plan-required`), counts as 1 request (SDD 08 §6, open issue 08-6) |
| `listGameGenres`, `listMovieGenres`, `listSeriesGenres`, `listBookGenres` | GET `/v1/{type}/genres` | `{data:[{slug,count}]}`; one operation per type because Go's ServeMux rejects `/{type}/genres` next to `/entities/{id}` (2026-10-29, P1.H.2.1b) |
| `listTrending` | GET `/v1/{type}/trending` | `window=day\|week` (SDD 08), top 100 |
| `listEditions` | GET `/v1/books/{id}/editions` | D-10 |
| `getEntity`, `listEntityItems` | GET `/v1/entities/{id}`, `/v1/entities/{id}/items` | SDD 01 |
| `getPerson`, `getOrganization`, `getFranchise`, `getPlatform` | GET `/v1/people/{id}`, `/organizations/{id}`, `/franchises/{id}`, `/platforms/{id}` | 404 if kind differs |
| `searchItems` | GET `/v1/search` | `q` 1–100 chars, `type`, `limit` ≤ 50 (SDD 07); ranks on allowlisted-source text only (07-7) |
| `lookupItem` | GET `/v1/lookup` | `source`, `id` (legacy sources); returns allowlisted fields only, never echoes restricted-source ids (FR-05-16) |
| `listChanges` | GET `/v1/changes` | `since` RFC 3339 required, `type`, `cursor`, `limit` ≤ 1000 |
| `listExports`, `getExport` | GET `/v1/exports`, `/v1/exports/{type}/{day}` | pro/business (else 403 `plan-required`); `getExport` → 302 to `disk.PresignGet(key, 5*time.Minute)` (`https://vukem.dev/docs/digging-deeper/exports` "S3") |
| `listWebhooks`, `createWebhook`, `deleteWebhook`, `testWebhook` | GET/POST `/v1/webhooks`, DELETE `/v1/webhooks/{id}`, POST `/v1/webhooks/{id}/test` | pro/business; secret returned once; not counted |
| `getUsage` | GET `/v1/me/usage` | legacy shape without credits; counts |

Cursors: legacy sorts need keyset on `(sort column, id)`; vukem `req.CursorPaginate` pages by primary key and rejects `sort` with a cursor (`https://vukem.dev/docs/rest-apis/pagination` "Query conventions"). Handlers sign their own opaque cursor (HMAC) and answer the vukem envelope `{data, meta:{next_cursor}, links:{next}}` (`https://vukem.dev/docs/rest-apis/pagination` "Envelopes"); tampered → 400 `invalid-cursor`. Unknown or empty params → 400 (legacy `lib/api/handler.ts`).

```json
{
  "id": "0192f3a4-7b1c-7e21-9c55-3f1e2d4a6b71", "type": "games", "slug": "hollow-knight", "title": "Hollow Knight",
  "locale": "en", "releaseDate": "2017-02-24", "externalIds": {"wikidata": "Q20984815"},
  "synopsis": {"text": "Hollow Knight is a 2017 Metroidvania...", "locale": "en",
    "licence": "CC-BY-SA-4.0", "sourceUrl": "https://en.wikipedia.org/wiki/Hollow_Knight"},
  "score": 8.7, "scoreByCategory": {"gameplay": 9.0, "fun": 8.8, "sound": 8.9,
    "technical": 8.2, "visual": 8.9, "storytelling": 7.9},
  "ratingCount": 412, "confidence": "high", "formulaVersion": 1,
  "provenance": {
    "title": {"source": "wikidata", "licence": "CC0-1.0"},
    "releaseDate": {"source": "wikidata", "licence": "CC0-1.0"},
    "synopsis": {"source": "wikipedia", "licence": "CC-BY-SA-4.0"},
    "score": {"source": "hexascore", "licence": "HEXASCORE-API-TERMS"}},
  "meta": {"omittedFields": ["cover", "runtimeMinutes", "trailers"]}
}
```

Coverage (published on `/docs/coverage` and stated on `/pricing` per D-33, ADR 0006; percentages measured by `qa/sql/api_coverage.sql` on staging, [NOT VERIFIED] until then):

| Field | Allowed sources | Expected gap |
| --- | --- | --- |
| id, type, slug, url, score fields, trending, timeToBeat | hexascore (A-07 API terms) | none |
| title, releaseDate, genres, people, organisations, franchises, platforms, countries, runtime, pages, seasons, episodes | wikidata (CC0), community (A-07) | runtime, platforms, episode counts patchy |
| synopsis | wikipedia (CC BY-SA 4.0 + `sourceUrl`) | items without an article |
| editions, book covers | openlibrary (A-05, [NOT VERIFIED]), wikidata | non-books n/a |
| covers for films, series, games; trailers | none in v1 | always omitted |
| externalIds | wikidata, isbn13, openlibrary | tmdb/igdb/steam omitted |

Problems (`https://vukem.dev/docs/rest-apis/overview` "Errors"; base `https://hexascore.app/problems/` via `APP_PROBLEM_BASE_URL`, each URL redirects to `/docs/errors#<code>`):

```json
{"type": "https://hexascore.app/problems/quota-exhausted", "title": "Quota exhausted",
 "status": 402, "detail": "Monthly quota is exhausted. Upgrade at https://hexascore.app/pricing.",
 "instance": "/api/v1/games", "request_id": "01JB7Q4V0Q8M3K2Y7T9H5N6C1D",
 "code": "quota-exhausted", "plan": "pro", "limit": 50000}
```

Codes: `unauthorized` 401, `invalid-query`/`invalid-cursor`/`changes-window-exceeded` 400, `quota-exhausted` 402, `plan-required`/`account-banned` 403, `not-found` 404, `rate-limited` 429, `internal-error` 500, `service-unavailable` 503. Extension members on `c.Problem` are [NOT IN VUKEM DOCS].

Webhook batch (≤ 100 events; Standard Webhooks headers `webhook-id`, `webhook-timestamp`, `webhook-signature` from `webhook.Sender`):

```json
{"type": "item.batch", "deliveredAt": "2026-11-02T10:15:00Z", "events": [
  {"id": "chg_1048576", "type": "item.updated", "itemType": "games", "itemId": "0192f3a4-7b1c-7e21-9c55-3f1e2d4a6b71",
   "fields": ["score", "ratingCount"], "changedAt": "2026-11-02T10:14:12Z"},
  {"id": "chg_1048577", "type": "item.deleted", "itemType": "movies", "itemId": "0192f3a4-...-5521",
   "mergedInto": "0192f3a4-...-5520", "changedAt": "2026-11-02T10:14:30Z"}]}
```

IGDB sends the shared secret in `X-Secret` (web-research 3.2); HexaScore signs with it instead (never on the wire, ±5 min replay window, verifiable with `webhook.Standard`). Payloads carry ids and field names only; clients refetch, so licence filtering stays in the serialiser.

### 6.3 Billing (install:billing, adapted)

| Route | Behaviour |
| --- | --- |
| `GET /billing` | Inertia page (SDD 09): plans, prices in locale currency, state |
| `POST /billing/checkout/{plan}` | pro/business, field `currency` brl/usd/eur; Checkout `mode=subscription`, one customer per user |
| `POST /billing/portal` | plan changes, cancellation |
| `POST /stripe/webhook` | `stripe.WebhookHandler` outside CSRF (`https://vukem.dev/docs/digging-deeper/billing` "Webhooks keep subscriptions in sync") |

Prices (minor units, `legacy/apps/web/lib/billing/catalog.ts`; kept at launch, reviewed 90 days after cutover, D-33): pro BRL 2900 / USD 600 / EUR 600; business BRL 14900 / USD 2900 / EUR 2700; default pt-BR→BRL, es→EUR, en→USD. Env `STRIPE_KEY` (`rk_`), `STRIPE_WEBHOOK_SECRET`, `STRIPE_PRICE_PRO`, `STRIPE_PRICE_BUSINESS`, each price with `currency_options` (legacy used `lookup_key`, `lib/billing/stripe.ts`). Passing `currency` is an edit of the scaffold code, not a documented option [NOT IN VUKEM DOCS].

Events (legacy `lib/billing/stripe-webhook.ts` + scaffold): `checkout.session.completed`, `checkout.session.async_payment_succeeded`, `customer.subscription.created|updated|deleted|paused|resumed`, `invoice.paid`, `invoice.payment_failed`.

### 6.4 `/app/v1` (A-01, ADR 0009)

Stack: `auth.Required(appTokenGuard)` (`hx_app_`, SDD 03), `NotBanned`, `auth.Verified`, `auth.Ability`; `ratelimit.ByToken` 120/min reads, 30/min writes, behind an IP limit. Read models use audience `site` (SDD 01 §6.1), so attributed TMDB/IGDB fields are allowed.

| Ability | Routes |
| --- | --- |
| none | `POST /app/v1/tokens`, `/tokens/google` (SDD 03 §6.2) |
| `catalog:read` | GET `/{type}`, `/{type}/{id}`, `/{type}/trending`, `/items/{id}/related`, `/me/for-you` (SDD 08), `/entities/{id}`, `/search`, `/items/{id}/me`, `/me`, `/me/library`, `/me/lists`, `/lists/{id}`, `/users/{nickname}`, `/me/notifications`, `/tokens` (SDD 14 §6.3 for library, profile and notifications) |
| `ratings:write` | PUT/DELETE `/items/{id}/rating` (SDD 02) |
| `status:write` | PUT/DELETE `/items/{id}/status`, `/items/{id}/time-to-beat` (SDD 02), `/items/{id}/edition` (SDD 14 §6.3) |
| `lists:write` | POST `/lists`, PATCH/DELETE `/lists/{id}`, PUT/DELETE `/lists/{id}/items/{itemId}`, PUT `/lists/{id}/order` |
| `profile:write` | PATCH `/me`, POST `/me/export` (202, creates the SDD 14 export job; was GET, open issue 14-2), POST/DELETE `/me/deletion`, POST `/me/notifications/read`, POST `/reports` (SDD 14), POST `/tokens/rotate`, DELETE `/tokens/{id}` |
| `suggestions:write` | reserved for community edits from the app (SDD 04, not in v1) |
| any valid app token, or none | POST `/client-errors`, POST `/client-metrics` (below) |

Writes accept `Idempotency-Key` (`api.Idempotency` on the database cache store, `https://vukem.dev/docs/rest-apis/idempotency` "Cross-cutting"). List limits copy `legacy/apps/web/lib/db/schema/social.ts` [NOT VERIFIED]. Tokens get all six README abilities (open issue 5). Route details, inputs and problems for library, profile, edition, notifications, reports and export are in SDD 14 §6.3.

Client telemetry intake (open issues 10-2, 09-11; processing and dashboards in SDD 12):

| Route | Input | Rules |
| --- | --- | --- |
| `POST /app/v1/client-errors` (`reportClientError`) | `{appVersion, versionCode, androidApi, screen, message, stack ≤ 8 KB}` (SDD 10 §7.7) | sent only after the user opts in (mobile settings, default off); no user id, token, e-mail or free-text user input; 10/min per IP; body > 16 KB → 413; 202; kept 7 days |
| `POST /app/v1/client-metrics` (`reportClientMetrics`) | `{name: LCP\|INP\|CLS\|TTFB, value, page (route pattern, not URL), locale}` (web-vitals beacon, SDD 09) | anonymous, opt-in via the cookie banner; 60/min per IP; body > 4 KB → 413; 202; kept 7 days |

Both accept unauthenticated calls (crashes can happen before login) behind the IP limit, CSRF-exempt because they carry no session effect; unknown fields → 400.

Versioning: the app sends `X-App-Version`; responses carry `X-App-Min-Version`; below `APP_MIN_SUPPORTED_VERSION` → 426 (logout exempt). Changes are additive; removal only once the minimum version excludes every build using the field and ≥ 90 days after the replacing build shipped; otherwise `/app/v2` side by side.

### 6.5 OpenAPI, reference, SDKs

| Document | Export | Consumers |
| --- | --- | --- |
| `openapi/api-v1.json` | `vukem openapi:export -prefix /api/v1` | Scalar; `@hexascore/sdk` (`@hey-api/openapi-ts` 0.99.0, `legacy/sdks/ts/openapi-ts.config.ts`); `hexascore` (openapi-python-client 0.29.1, `legacy/sdks/python/config.yml`) |
| `openapi/app-v1.json` | `vukem openapi:export -prefix /app/v1` | `vukem codegen:client -prefix /app/v1 -o packages/app-client/src/api.ts` → private `@hexascore/app-client` |

Serving: `vukem.ServeOpenAPIWith(app, r, "/api/v1/openapi.json", vukem.OpenAPIOptions{PathPrefix: "/api/v1", InProduction: true})`; app document not served in production; `apidocs.Serve(app, r, "/docs/reference", apidocs.Config{SpecURL: "/api/v1/openapi.json", InProduction: true})` (`https://vukem.dev/docs/rest-apis/openapi` "API reference UI"). Translated Scalar chrome (D-06) is [NOT IN VUKEM DOCS]; fallback is English Scalar inside a translated page shell. `operationId` = `Op.Name`, camelCase `<verb><Noun>` (`list|get|search|lookup|create|delete|test`), groups named `api.v1.` and `app.v1.` so each document reuses names without `ErrDuplicateOperationID` (`https://vukem.dev/docs/rest-apis/openapi` "OpenAPI"). Title `HexaScore API`, version `API_VERSION`, scheme `apiKey` declared via `vukem.WithOpenAPI`. SDK version = API version; npm publication follows `legacy/docs/sdk-publish-npm.md` (trusted publishing).

Docs (Inertia, English guides, translated chrome): quickstart, authentication, errors, pagination, rate limits, coverage, webhooks, exports, changes, changelog.

Versioning: API SemVer 1.0.0 at parity (D-20); additive = minor, fix = patch, breaking = `/api/v2` side by side, v1 marked with `api.Deprecate` ≥ 6 months before `Sunset` (legacy rule), announced in the changelog and by email to key owners.

## 7. Behaviour and algorithms

### 7.1 Key resolution

`sha256(key)` → in-process LRU (10k, TTL 60 s) of `{keyID, userID, plan, bannedAt}`; miss → one indexed join of `api_keys`, `users`, `subscriptions`. Revoke, ban and subscription sync evict in-process (single instance, D-08). `last_used_at` is written by the quota flush.

### 7.2 Quota metering (ADR 0003)

- Memory per user: `{period, base, delta}`; `base` loaded from `api_quota_counters` on first use.
- Admit if `base + delta < quota`, then `delta++`; 5xx refund `delta--`, `refunded++`.
- Flush every minute and on shutdown by an app runner in the serving process (`api.Meter`, registered in `routes/api_v1.go`; a scheduled job could run in another process without the in-memory deltas): per 500 users one transaction, `insert ... on conflict do update set used = used + excluded.used` into both tables; deltas cleared after commit.
- Crash loses ≤ 60 s of deltas: free requests, never overcharge (free worst case 60 extra).
- Rollover at 00:00 UTC on the 1st (`YYYYMM`), lifetime row `life`; plan change applies new limits immediately.
- Flush failure: keep deltas, retry; alert after 5 failures (SDD 12); fail closed (503) once 10 min are unflushed.
- Legacy credits (D-23): while `api_credit_balances.remaining > 0` a counted request decrements the balance (flushed like `delta`) instead of the plan counter; `X-Quota-Remaining` reports plan quota plus credits.

### 7.3 Changes, exports, webhooks

- `api_changes` is written in the cause's transaction: catalogue apply/revert (`catalog.changed`, SDD 04), ingestion upserts (SDD 06), merges (SDD 01), `ItemScoreChanged` (SDD 02) only when the one-decimal score, confidence or `ratingCount` changed. `fields` holds api-visible names only; hidden-only changes write nothing.
- Exports: `s.Job(jobs.ExportIDs{}).DailyAt("07:00")` UTC; gzip NDJSON `exports/{type}_ids_{YYYY_MM_DD}.json.gz`, lines `{"id","slug","wikidata","popularity","updatedAt"}`, streamed with `Query.Each` (`https://vukem.dev/docs/digging-deeper/exports` "Exports"); objects older than 90 days deleted.
- `webhook_deliveries` retention: 30 days, pruned daily (§5).
- Fan-out job every minute: read ≤ 10,000 rows after `webhook_cursor`, match endpoints by `events`/`item_types`, coalesce per item, send batches through `webhook.Sender` (8 retries from 30 s, 410 stops). Three exhausted deliveries in a row disable the endpoint and mail the owner; gaps are recovered from `/v1/changes`.

### 7.4 Request lifecycle

IP limit → key guard (401/503) → ban (403) → validation (400, free) → throttle (429) → quota (402) → handler → 2xx/304/404 counted, 5xx refunded. ETag via `web.ETagOf`, `c.NotModified(etag, updatedAt)`.

### 7.5 Plan state machine

| Stripe status | `state` | Effective plan |
| --- | --- | --- |
| `active`, `trialing` | active | plan while `current_period_end > now` |
| `past_due`, `unpaid`, `incomplete`, `paused` | past_due | free (legacy: no grant) |
| `canceled`, `incomplete_expired` | canceled | free |
| period ended, provider without renewal (PIX) | expired | free |

Each sync re-reads with `GetSubscription` and upserts by `stripe_subscription_id`; `cancel_at_period_end` stays active to the end. Falling to free applies free limits next request; the `life` counter continues. Duplicates are acknowledged by the vukem event claim (`CACHE_STORE=database`, `https://vukem.dev/docs/digging-deeper/billing`); the upsert is idempotent anyway. Deletion cancels subscriptions first (SDD 03).

Reminders: `s.Job(jobs.BillingReminders{}).DailyAt("12:00")` UTC (legacy `/api/cron/billing-reminders`) mails 7 days and 1 day before `current_period_end` when `cancel_at_period_end` is true; idempotency key `reminder:<user>:<end>:<days>` in the cache store (same persistent `database` store as webhook dedup, TTL past `current_period_end`); marks non-Stripe past-period rows `expired`. Card renewal and failure emails are Stripe's [NOT VERIFIED].

### 7.6 PIX boundary (second release)

```go
type Provider interface {
	Name() string // "stripe" | "abacatepay"
	StartCheckout(ctx context.Context, u UserID, o Offer) (Redirect, error)
}
type Offer struct{ Plan string; Months int; Currency string } // amounts from the catalogue only
```

Key resolution and quota read only `subscriptions(plan, state, current_period_end)`; AbacatePay adds a provider, webhook, `pix_charges` and credits without touching 7.1–7.4.

## 8. Edge cases

- Revoked key: next request 401 (in-process eviction; ≤ 60 s otherwise).
- Several keys of one user share throttle and quota (`ByUser`).
- `limit` above cap → 400 naming the cap; plan change mid-cursor keeps the cursor valid.
- Merged id → 301; feed emits `deleted` with `mergedInto`.
- `since` in the future → 400; exactly 14 days → allowed.
- All optional fields hidden → id, type, slug, score with `omittedFields`.
- Webhook URL resolving to a private IP → 422 at create, refused at delivery.
- Stripe event before the customer is linked → logged, acknowledged; checkout completion re-syncs.
- Crash during flush → rollback; at most one interval undercounted.

## 9. Acceptance criteria

- `go test ./app/api/... -run 'TestQuota|TestKeys|TestPlans'` → pass; free key requests 1–100 → 200, 101 → 402.
- `go test ./tests -run TestAPIContract` → every response passes `AssertMatchesOpenAPI()`.
- `go test ./tests -run TestAPILicenceFilter` → TMDB/IGDB fixture fields absent from `/api/v1`, present in `/app/v1`.
- `vukem openapi:export -prefix /api/v1 | diff - openapi/api-v1.json` → empty (same for `/app/v1`); Redocly lint (pinned) → 0 errors.
- `schemathesis run openapi/api-v1.json --url http://localhost:3000 -H "Authorization: Bearer $HX_TEST_KEY" --checks all` (pinned) → 0 failures.
- `k6 run qa/k6/api-v1.js` → list `p(95)<300`, detail `p(95)<150` at 50 rps, `http_req_failed<0.001`.
- `curl -si -H "Authorization: Bearer $K" -H "If-None-Match: $ETAG" https://staging-api.hexascore.app/v1/games/1` → 304, `X-Quota-Remaining` down by 1.
- `go test ./tests -run 'TestStripeWebhook|TestWebhookDelivery|TestAppMinVersion'` → duplicate processed once, forged 400, `past_due` → free; signed batch received, 410 disables; 426 below minimum.
- `cd sdks/ts && bun run sdk:generate && bun test`; `cd sdks/python && uv run pytest -q` → pass, plus smoke calls `listItems`, `getItem`, `getUsage` against staging.

## 10. Test strategy

- Unit: plan table, admit/refund/rollover, cursor signing, change coalescing, status mapping.
- Integration (Postgres, vukemtest): auth stack, counting per status, flush crash, changes window, exports with S3 fake, billing with the `STRIPE_API_BASE` fake (`https://vukem.dev/docs/digging-deeper/billing` "Tests"), webhooks with `httptest`.
- Contract: `AssertMatchesOpenAPI` in every route test; Schemathesis on both documents in CI; golden diff; Redocly.
- SDK smoke: generated TS and Python clients in CI.
- Load: k6 per NFR-05-1, with and without metering.
- Security: ZAP; `TestSecurity_*` for key hashing, revocation, webhook SSRF, abilities, cross-surface credentials.
- E2E (Playwright + axe, SDD 09): key shown once, checkout redirect, docs and reference.

## 11. Open issues

1. Host routing is [NOT IN VUKEM DOCS]. VF-015; decided: app middleware rewrites the path on the API host, OpenAPI `servers` = `https://api.hexascore.app/v1` (§6.1). Owner: vukem.
2. Licence of HexaScore and community data in the paid API. Decided: A-07.
3. `lookupItem` by third-party ids. Decided (05-3): ids accepted as input, allowlisted fields only, restricted-source ids never echoed (FR-05-16).
4. Legacy credit balances. Decided: honoured as extra quota before the plan quota, no refunds (D-23; 7.2).
5. Ability names. Decided: README convention; SDD 03 updated.
6. Casing mix: camelCase fields vs vukem snake_case envelope keys. VF-014; README convention accepted (§6.1). Owner: vukem.
7. Problem extension members and Scalar UI translation are [NOT IN VUKEM DOCS]. Owner: vukem. Closes with docs or `vukem-feedback.md` answers.
8. Keyset cursor on allowlisted sorts is [NOT IN VUKEM DOCS]; custom cursor used. Owner: vukem. Closes when `api.Query` supports it.
9. Problem code casing. Decided: kebab-case; SDD 03 updated.
10. Prices and limits. Decided: legacy table kept at launch, review 90 days after cutover, coverage stated on the pricing page (D-33).
11. Field coverage percentages unmeasured. Owner: architect. Closes after `qa/sql/api_coverage.sql` runs on the first staging copy.
