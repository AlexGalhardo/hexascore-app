# SDD 09: Web application

Status: reviewed (2026-10-28 consistency pass). Decisions: D-00, D-01, D-02, D-04, D-06, D-07, D-13, D-17, D-19, A-06. ADRs: 0004, 0005, 0009, 0016. Related SDDs: 01, 02, 03, 04, 05, 07, 08, 10, 12, 13.

## 1. Objective

hexascore.app (D-19) as one Inertia v3 + React 19 + TypeScript app served by Go: public pages server-rendered (D-07) and fast enough to win search traffic, one-click half-star rating (D-00), stars 0–10 with count and confidence (D-13), entity navigation (D-01), trending per type (D-02), three locales (D-17), tokens shared with Android (ADR 0009). Performance targets follow the owner's PageSpeed references (tools.vukem.dev mobile, marketplace.vukem.dev desktop).

## 2. Scope

In (v1): shell, SSR, routing/i18n, §6.2 pages, components, tokens, budgets, accessibility, SEO, CSP, notifications UI, build, tests. Out: feed, follows, PIX (D-06; passkeys are back in scope, ADR 0017); RTL; PWA; visual regression as a gate. Admin CRUD is vukem's admin panel (<https://vukem.dev/docs/digging-deeper/admin>).

## 3. Functional requirements

- FR-09-1: SSR routes (§6.2) return `<h1>`, primary content and meta without JavaScript. Verify: Playwright, `javaScriptEnabled: false`.
- FR-09-2: CSR routes call `inertia.DisableSSR(ctx)` (<https://vukem.dev/docs/frontend/inertia#server-side-rendering>) and render after hydration. Verify: `AssertInertia` + Playwright.
- FR-09-3: Locale from the URL prefix; unprefixed GETs 307 to the prefixed URL when cookie `locale`, else `Accept-Language`, selects `pt-BR`/`es`. Verify: Go integration, 9 cases.
- FR-09-4: Indexable pages emit a self canonical and `hreflang` `en`, `pt-BR`, `es`, `x-default` (→ en). Verify: `seo.spec.ts`.
- FR-09-5: The shared catalogue is a `Once` prop, never resent within a locale. Verify: Playwright inspecting Inertia JSON.
- FR-09-6: Star widget meets §7.4. Verify: Vitest + Playwright + axe.
- FR-09-7: Rating/status saves are optimistic for the user's own value, rolled back with a toast on failure; aggregates only from the server; suggestions never optimistic. Verify: Playwright with a forced 429.
- FR-09-8: `props.errors` shown per field; focus moves to the first invalid field. Verify: Playwright.
- FR-09-9: Unknown URLs and missing records (item, entity, document, hidden list, unknown profile) on a GET HTML page answer 301 to the home page of the locale (`/`, `/pt-br/`, `/es/`; owner decision 2026-11-01, item 14; `controllers.NotFound`). JSON requests, `/api`, `/app`, `/v1`, asset prefixes (`/build`, `/covers`, `/og`, `/sitemaps`, `/.well-known`), paths with a file extension and non-GET requests keep a real 404 (problem+json for the APIs, the `Error` component on Inertia visits); auth and validation errors are unchanged. Full-page 5xx keep their status with `resources/views/errors/500.html` and `503.html`, localised, with a link home (<https://vukem.dev/docs/basics/error-handling#custom-error-pages>). Verify: `TestMissingPage_RedirectsHome`, `TestServerError_PageLinksHome`, `not-found.spec.ts`.
- FR-09-10: Item JSON-LD per type; `aggregateRating` only when `score` is not null (SDD 02 FR-02-6). Verify: Go test on SSR HTML.
- FR-09-11: `/sitemap.xml` indexes chunks ≤ 40,000 URLs (legacy `SITEMAP_CHUNK`, `lib/seo.ts`) with `xhtml:link` alternates. Verify: XML integration test.
- FR-09-12: Navbar unread count and a page rendering `user_notifications` (SDD 04 §5) via i18n keys. Verify: Playwright.
- FR-09-13: Logout calls `inertia.ClearHistory`; logged-in pages use `inertia.EncryptHistory` (<https://vukem.dev/docs/frontend/inertia#history-encryption>). Verify: back-button test.
- FR-09-14: Zero CSP violations in E2E (`securitypolicyviolation` listener). Verify: Playwright fixture.

## 4. Non-functional requirements

| ID | Metric | Budget | Measured by |
| --- | --- | --- | --- |
| NFR-09-1 | LCP, mobile emulation (Moto G Power, slow 4G) | ≤ 2.5 s (was 2.0 s; raised per Q-A20) | Lighthouse CI, median of 3 runs |
| NFR-09-2 | INP | ≤ 150 ms field (p75); lab proxy TBT ≤ 200 ms (open issue 09-11) | Lighthouse CI TBT; Playwright Event Timing on star click; field: `web-vitals` beacon (LCP, INP, CLS, TTFB; opt-in, anonymous) to `POST /app/v1/client-metrics` (SDD 05 §6.4, SDD 12) |
| NFR-09-3 | CLS | ≤ 0.05 | Lighthouse CI |
| NFR-09-4 | TTFB p75 (SSR, staging) | ≤ 400 ms; ≤ 150 ms on props-cache hit | k6 + SDD 12 histogram |
| NFR-09-5 | JS on item page | ≤ 170 kB gzip total (was 150 kB; raised per Q-A20), ≤ 90 kB before interaction | Lighthouse `resource-summary` + `bun run size` |
| NFR-09-6 | CSS | ≤ 30 kB gzip | same |
| NFR-09-7 | Fonts | ≤ 3 WOFF2 files, ≤ 100 kB total, `font-display: swap`, 1 preload | Lighthouse |
| NFR-09-8 | Lighthouse scores (mobile) | Performance ≥ 90, Accessibility 100, SEO 100, Best Practices ≥ 95 | Lighthouse CI |
| NFR-09-9 | SSR render time p95 | ≤ 80 ms item page; SSR fallback ≤ 0.1 % of public renders/day (ADR 0004) | SDD 12 metrics |
| NFR-09-10 | Shared i18n prop | ≤ 12 kB gzip per locale | unit test on catalogue size |
| NFR-09-11 | Accessibility | WCAG 2.2 AA; axe 0 serious/critical | Playwright + axe |
| NFR-09-12 | SQL per page render | 0 N+1 alerts; item page ≤ 12 queries cold | vukem devtools `/_vukem/queries` (<https://vukem.dev/docs/digging-deeper/devtools#finding-an-n1-query>), dev checklist |
| NFR-09-13 | Viewport | every page usable at 360 px (legacy `docs/design-system.md` "Princípios") | Playwright mobile project |

Note (2026-11-01, Q-A20 in `docs/orchestration/QUESTIONS.md`): the owner raised NFR-09-1 from 2.0 s to 2.5 s (the Core Web Vitals "good" threshold) and NFR-09-5 from 150 kB to 170 kB after the release PR's Lighthouse check failed on GitHub runners (LCP medians 2.1 to 2.5 s). `lighthouserc.cjs` asserts the same values.

## 5. Data model

No tables owned here. Consumed: SDD 01 read models (`ItemSummary`, `ItemDetail`, `EntityDetail`, `Edition`), SDD 02 rating fields, SDD 14 `user_notifications` (open issue 14-1), SDD 08 trending table. Cookies (all `Secure`, `SameSite=Lax`, `Path=/`):

| Cookie | Value | Lifetime | HttpOnly |
| --- | --- | --- | --- |
| session (`__Host-` prefix, vukem default, <https://vukem.dev/docs/basics/sessions>) | encrypted | SDD 03 | yes |
| `locale` | `en`, `pt-BR`, `es` | 365 days (legacy `i18n/routing.ts`) | yes |
| `theme` | `light`, `dark`, `system` | 365 days | yes |

## 6. Contracts

### 6.1 Shell

- `vukem.WithInertia(inertia.Config{Version: func() string { return app.Vite.Version() + catalogHash8 }, SSR: &inertia.SSRConfig{Bundle: "bootstrap/ssr/ssr.js"}})` (<https://vukem.dev/docs/frontend/inertia#configuration>, `Version` is a `func() string`); mixing in the catalogue hash makes a translation-only deploy trigger the 409 full reload too (<https://vukem.dev/docs/frontend/inertia#asset-versioning>).
- Root view `resources/views/app.html`: `<html lang="{{locale}}" data-theme="{{theme}}">` where `theme` is an app view function reading the `theme` cookie (`light`/`dark`, empty for `system`), because `inertia.RootData` carries no theme value (VF-028, open issue 09-6), CSP nonce meta, `{{vite …}}`, `{{.Head}}`, `{{.Inertia}}` (<https://vukem.dev/docs/frontend/inertia#the-pieces-of-an-inertia-app>).
- `SSRConfig.Except`: `/admin`, `/profile`, `/me`, `/billing`, `/suggest`, `/suggestions`, `/library`, `/lists`. `OnFailure` receives an `inertia.SSRError` (`Type`, `Component`, `URL`) and increments `hexa.ssr.fallbacks{http.route, public}` (SDD 12 §6.2 name, alert A5) (<https://vukem.dev/docs/frontend/inertia#server-side-rendering>).
- Readiness: `vukem.WithHealthCheck("ssr", …)` wrapping `app.Inertia.SSRHealthy(ctx)` (same page), reported by `/up?checks=1` (<https://vukem.dev/docs/deployment/docker-and-railway>).
- Pages in `resources/js/pages/<Area>/<Name>.tsx`, lazy on the client, eager in `resources/js/ssr.tsx`.

Shared props (every page):

```ts
type Locale = "en" | "pt-BR" | "es";
type Shared = {
  auth: { user: AuthUser | null };                 // SDD 03 §6.1
  locale: Locale; locales: Locale[];               // from inertia.I18n(prefixes...)
  messages: Record<string, string | Record<string, string>>; // flat keys, once prop "i18n:<locale>"
  seo: { title: string; description: string; canonical: string;
         alternates: Record<Locale | "x-default", string>; robots: string;
         og: { image?: string; type: string }; jsonLd?: object[] };
  notifications?: { unread: number };              // logged in only
  flash: { status?: string; error?: string };      // page.flash via inertia.Flash
};
```

Prop types reuse the OpenAPI schemas: `vukem codegen:types` over the `/app/v1` document (ADR 0009) writes `resources/js/types/api.d.ts`; page-only props live in `resources/js/types/pages.ts`. Typed routes via `vukem codegen:routes` (<https://vukem.dev/docs/cli/reference-runtime#vukem-codegenroutes>; types: `#vukem-codegentypes`). The web makes no fetch calls except `/search/suggest` and `/search/events` (SDD 07 §6), the ratings dialog, the edition picker and the data export; no generated HTTP client ships to the browser. Those client-fetched calls go through TanStack Query (ADR 0020): one `QueryClient` per browser app, a fresh one per SSR render, and never for Inertia page props.

### 6.2 Page inventory

Paths shown without locale prefix; every public path also exists under `/pt-br` and `/es`. Legacy route tree: `legacy/apps/web/app/[locale]/`. "SSR" = rendered by Node; "CSR" = `DisableSSR`.

| Component | Path | SSR | Props (beyond shared) |
| --- | --- | --- | --- |
| `Home/Index` | `/` | SSR | `trending.{games,movies,series,books}: TrendingEntry[10]` all eager (public SEO content, open issue 08-8; `Defer` only for personalised or below-the-fold non-SEO props); `forYou` `Defer` group `personal` (logged in, SDD 08); `sections: {type, count}[]` |
| `Catalog/Index` | `/{section}` (`games`, `movies`, `series`, `books`) | SSR | `type`, `filters {genre, platform, year, minScore, sort, per, min_<category>}`, `facets` (`Optional`), `items` (`?per=` 25, 50, 75 or 100, any other value is 25; owner round 2026-11-01 items 12 and 20), `pagination {prev, next}` (signed cursors; previous/next buttons above and below the list, `rel=prev/next` head links that keep the query, canonical keeps `cursor` and a non-default `per` and drops the filters), `total`. Layout from `lg` up: sticky filter sidebar left, results right; below `lg` the filters collapse behind the "Filters" disclosure button |
| `Catalog/Show` | `/{section}/{slug}` | SSR | `item: ItemDetail`, `ratingSummary {score, scoreByCategory, ratingCount, confidence, formulaVersion, unusualActivity, ratingsLimitedUntil}`, `entities` by role (inline), `editions` (books: first 10 + `editionCount`), `timeToBeat` (games), `me {rating, status, timeToBeat}` (logged in; `Optional`), `similar` and `crossMedia` (eager, SDD 08 §6: internal links are SEO content), `recentChanges` (`Defer`, 5 entries, group `below`) |
| `Catalog/Edition` | `/books/{slug}/editions/{id}` | SSR | `work`, `edition: Edition` |
| `Catalog/RecentChanges` | `/{section}/{slug}/changes`, `/entities/{kind}/{slug}/changes` | SSR | `target`, `changes` (`Scroll`, 20) (SDD 04 FR-04-13) |
| `Entity/Show` | `/{section}/{people,companies,franchises,platforms,genres,game-modes}/{slug}`, `/{section}/{countries,languages}/{code}` (owner round 2026-11-01 item 19; changes page at `…/changes`). The old top-level `/{entity path}/{key}[/changes]` answer one 301 to the category named by `?type=`, else the one with the most public items of the entity (games when none). The section patterns are served by one dispatcher per segment count (`routes.sectionPages`), since ServeMux rejects `/games/{slug}/changes` next to `/games/platforms/{slug}`; the entity path words are reserved item slugs | SSR | `entity: EntityDetail`, `roles`, `items` (`Scroll`, the section's type only), `counts` (every type, for the "Also in" links), `avgScore`, `filters.type`; indexable from 3 public items of the section's type |
| `Search/Index` | `/search` | SSR, `noindex, follow` | per SDD 07 §6 |
| `User/Profile` | `/u/{nickname}` | SSR | per SDD 14 §6.2; private profile → 200 with the nickname only and `noindex` (legacy parity, open issue 14-5); unknown nickname → 301 home (FR-09-9) |
| `User/List` | `/u/{nickname}/lists/{id}` | SSR | `list`, `items`, `canEdit`, `seo{indexable}` (SDD 14 §6.2) |
| `Library/Index` | `/library` | CSR, `noindex` | SDD 14 §6.2 |
| `Lists/Index` | `/lists` | CSR, `noindex` | SDD 14 §6.2 |
| `Lists/Edit` | `/lists/{id}/edit` | CSR, `noindex` | SDD 14 §6.2 |
| `Auth/*` | SDD 03 §6.1 | SSR, `noindex` | SDD 03 |
| `Profile/Index` | `/profile` | CSR | sections per SDD 03 §6.1 |
| `Developer/ApiKeys` | `/profile/api` (legacy path) | CSR | `keys`, `plan`, `usage` (30 days), `newKey` (flash, shown once) |
| `Billing/Index` | `/billing` | CSR | SDD 05 §6.3 |
| `Pricing/Index` | `/pricing` | SSR | `plans`, `currency`, `coverage` |
| `Account/Notifications` | `/me/notifications` | CSR | `notifications` (`Scroll`, 20), SDD 14 §6.2 |
| `Community/*` | SDD 04 §6 | CSR | SDD 04 |
| (vukem views) | `/admin/moderation…` (SDD 04), `/admin/ingest` (SDD 06) | Go `html/template` | not Inertia: server-rendered views inside the vukem admin layout (<https://vukem.dev/docs/digging-deeper/admin>, <https://vukem.dev/docs/basics/views>), open issue 09-12 |
| `Docs/Page` | `/docs`, `/docs/{authentication,errors,pagination,rate-limits,coverage,webhooks,exports,changes,changelog}` | SSR | `slug`, `body` (pre-rendered HTML from Markdown at build, sanitised), `toc` |
| (vukem `apidocs`) | `/docs/reference` | Go | Scalar, SDD 05 §6 |
| `Help/Score` | `/how-score-works` (301 from `/help/score`) | SSR | SDD 02 help outline; same `DocumentPage` layout as `Legal/Page` |
| `Help/DeleteAccount` | `/help/delete-account` | SSR, public | how to delete the account in the app and on the web, what is kept (anonymised ratings) and the 30-day grace; link to `/profile#delete` (SDD 03 FR-03-15); URL registered in Google Play (SDD 10, open issue 10-7) |
| `Legal/Page` | `/terms`, `/privacy` (301 from `/policy` and `/legal/privacy`) | SSR | `slug`, `body`, `updatedAt` |
| `Contact/Index` | `/contact` | SSR | `topics`; POST with honeypot (SDD 03) |
| `Error` | catch-all `GET /{path...}` per prefix, and `inertia.RenderStatus` (<https://vukem.dev/docs/frontend/inertia#rendering-pages>) | SSR | `status`, `title`, `message` |

Web mutation routes (session guard, same handlers as `/app/v1`, SDD 02 §6): `PUT|DELETE /items/{id}/rating`, `PUT|DELETE /items/{id}/status`, `PUT|DELETE /items/{id}/time-to-beat`, lists, editions, reports, notifications and export per SDD 14 §6.2, `POST /locale` `{locale}`, `POST /theme` `{theme}`. All answer 303 back with flash; Inertia partial reload `only: ['ratingSummary','me']` after rating saves.

## 7. Behaviour and algorithms

### 7.1 Routing and i18n

- Routes are registered once per prefix with `r.Group(prefix, fn).Name(...)` (<https://vukem.dev/docs/basics/routing#groups>): `en.item.show`, `pt-br.item.show`, `es.item.show`; client helper `localized(name, params)`.
- Locale order: URL prefix (authoritative); on unprefixed GETs only, cookie `locale`, then `Accept-Language`; non-en → 307 to the prefixed URL with `Vary: Cookie, Accept-Language`. Crawlers get the canonical en page. `vukem.DefaultMiddlewareWith` gets `i18n.MiddlewareConfig{Order: {SourceCookie, SourceHeader}, Query: "-", SessionKey: "-", Cookie: "locale"}` (<https://vukem.dev/docs/frontend/inertia-i18n#configuration>); an app middleware after it applies the prefix with `i18n.WithLocale` (<https://vukem.dev/docs/digging-deeper/localization#translating-in-go-code>), before `app.Inertia.Middleware` and `inertia.I18n` (order rule in <https://vukem.dev/docs/frontend/inertia-i18n#common-mistakes>). A URL-prefix source is [NOT IN VUKEM DOCS].
- The switcher posts `/locale` and redirects to `seo.alternates[locale]`; slugs are identical across locales (D-17).
- Catalogues: `resources/lang/{en,pt-BR,es}.json` (vukem's one-file-per-locale rule), top-level keys = legacy namespaces (`legacy/apps/web/messages/en/`): `common, categories, catalog, item, ratings, search, errors, auth, account, lists, social, contact, billing, api, docs, legal, openapi`. A shared build step in `packages/i18n` (`bun run i18n:build`) flattens them once with the same fallback chain `inertia.I18n` applies (`pt-BR` → `pt` → `en`) and writes the flat form consumed by the web client tests and the mobile app (SDD 10), so both clients read identical keys (ADR 0009, open issue 09-5); CI diffs its output against the server's `messages` prop for the three locales.
- Delivery: `inertia.I18n("common.", "categories.", "catalog.", "item.", "ratings.", "search.", "errors.", "auth.", "lists.", "social.", "notifications.", "validation.")` shares `locale`, `locales` and `messages` (flat keys, fallback applied), `messages` being a once prop under `i18n:<locale>` (<https://vukem.dev/docs/frontend/inertia-i18n#step-by-step>). Prefixes are mandatory so server-only texts (mail, admin) never ship (<https://vukem.dev/docs/frontend/inertia-i18n#security>). Excluded namespaces `docs, legal, openapi, billing, api, account, contact` (legacy en: 41 of 62 kB raw) reach their pages as a `pageMessages` prop.
- Format: `{name}` placeholders and CLDR plural objects; legacy ICU plurals are converted once, nested `select` flattened. Client `translate()`/`useT()` per the guide (`Intl.PluralRules`, explicit `zero` wins); golden test Go vs TS for counts 0, 1, 2, 5, 21, 101.
- `Intl.NumberFormat`/`DateTimeFormat` per locale; scores one decimal (`7,8` in pt-BR); date-only values in `timeZone: "UTC"`; timestamps rendered UTC on SSR, re-formatted to the user's zone after mount.

### 7.2 Render caching and HTTP caching

- Public page props are cached with `cache.Remember` (<https://vukem.dev/docs/digging-deeper/cache#remember-compute-once-read-many-times>, stampede-safe) in `cache.NewMemory(5000)`, key `page:v1:{locale}:{component}:{path}?{sorted allowed query}`, TTL 60 s (home 300 s); `catalog.ItemsChanged` (SDD 06) and `ItemScoreChanged` (SDD 02) delete the item's and its entities' keys in 3 locales. Logged-in requests reuse them and add `me`, `notifications`.
- No documented hook caches SSR output [NOT IN VUKEM DOCS]; every first load runs SSR.
- Anonymous Inertia JSON visits: `web.ETagOf(pageObject)` + `c.NotModified` (<https://vukem.dev/docs/basics/responses#conditional-requests>), `private, no-cache`, `Vary: X-Inertia`. Full HTML loads of the public catalogue pages (home, listings, items, entities; `app/http/middleware.RenderCache`): `public, max-age=60, stale-while-revalidate=300` with `Vary: Cookie, Accept-Language` for anonymous 200 responses that set no cookie, so a cache keeps the page and the CSP header with its nonce together; `private, no-store` when the request has the session cookie or the response sets a cookie. The same anonymous 200 pages are also kept in process for 60 s (max 1000 entries, one render per key at a time, keyed by path, query, the `theme`, `consent` and `locale` cookies and, on unprefixed paths, `Accept-Language` (it picks the FR-09-3 redirect); session requests and responses that set a cookie bypass it) and replayed with the nonce swapped for the request's own CSP nonce, so SSR runs once per page per minute and the nonce stays per request (A05). Other HTML: `private, no-cache`. Never an ETag on HTML (a 304 would pair cached HTML's old nonce with a new CSP; decided 09-2). ETag only on JSON (Inertia visits, APIs) and assets; the TTFB budget is met by the props render cache plus SSR, not by 304s. Logged in: `private, no-store`.
- `/build/*` immutable (<https://vukem.dev/docs/frontend/vite#asset-caching>); covers by slug, `/covers/{section}/{slug}[-t].webp` (owner decision 2026-11-01, item 18): the route maps the slug to whatever key the row holds, slug keys get `public, max-age=86400` with the bucket ETag, content-hashed keys stay 1 year immutable, and an old `/covers/<item id>-<hash>.webp` answers 301 to the slug path (SDD 06); sitemaps and robots `public, max-age=3600` + ETag. Only those `/build` JS, CSS, SVG, font and source-map responses are gzip-encoded (`app/http/middleware.GzipStatic`). Anonymous public catalogue HTML (the pages `RenderCache` marks `public, max-age=60`, no session cookie in, no cookie out) is gzip-encoded by `GzipPublicPages` with `Vary: Accept-Encoding` (145 kB to about 25 kB, mobile LCP); all other HTML and any response that sets a cookie stay uncompressed (BREACH, SDD 13 §7.5).

### 7.3 Listings and pagination

Listing rule (legacy `docs/frontend.md` lines 60–62): mixed results grouped by type in `ITEM_TYPES` order under a `CategoryLabel`; `CoverThumb` left of every title; empty groups hidden. Sorted listings use handler-signed keyset cursors (SDD 05 §6), which `ScrollCursor(orm.CursorPage, …)` (primary-key paging) cannot express, so they use `inertia.Scroll(v, inertia.ScrollMeta{PageName: "cursor", PreviousPage, NextPage, CurrentPage})` filled by hand (<https://vukem.dev/docs/frontend/inertia#infinite-scroll>); opaque string cursors in those fields are [NOT VERIFIED] (the page shows page numbers). Loading: section listings (`Catalog/Index`) are paged, not scrolled (owner round 2026-11-01 item 12): plain `items` plus `pagination {prev, next}` cursors, rendered as previous/next links above and below the list (`?cursor=`, works without JS) and as `rel=prev/next` head links; entity pages keep the scroll prop. Cursor and filtered pages are `noindex, follow`.

### 7.4 Star rating widget

- One widget per category (six per type, `legacy/apps/web/lib/categories.ts`, A-06); 0.5–10 in 0.5 steps, unset = null; Save sends all six (SDD 02) and is enabled once all are set.
- Pointer: 10 stars × 2 half zones, hover preview, one click sets, drag updates. Touch: the row is one ≥ 44 px tall target plus 44×44 px − / + buttons for precision at 360 px.
- Keyboard (SDD 02 NFR-02-7): Left/Down −0.5, Right/Up +0.5 (no wrapping), Home 0.5, End 10, Delete/Backspace clears; one Tab stop per category; 2 px focus ring in `primary` around the group.
- ARIA (design v1): a radio group, not a slider. `role="radiogroup"` labelled by the category, 20 native `type="radio"` inputs (0.5 to 10) sharing one `name`, each labelled `rating.valuetext` "{value} of 10 stars"; no radio checked = unset ("Not rated"). Native radios keep the form usable without JavaScript and give screen readers the standard "n of 20" position.
- Filled `star`, empty `border`; the number is always visible; no animation under reduced motion.
- Disabled: logged out (login link), `rating-not-released`, `ratingsLimitedUntil`, 429 → notice keyed by problem `code`.
- `ScoreStars` (read-only): `7.8/10`, `ratingCount`, confidence chip; `null` → "Not enough ratings (n/5)".

### 7.5 Components and design tokens

Kept from legacy: `CategoryLabel`, `CoverThumb` (32×48 / 44×64, `srcSet` 96w/600w, lazy; item-page cover eager with `fetchpriority="high"`), `Field`/`Input`/`Notice`, `HexRadar`, `ScoreBar`, item table stacked below `md`. Dialogs are centred on both axes, full width minus margins on mobile, focus trapped, above navbar (`z-40`) and cookie banner (`z-50`) (legacy `docs/frontend.md` "Modais"), via one `dialogPopup(size)` helper enforced by a unit test. Primitives: Base UI. Every `Defer` group has a skeleton of its final size.

`packages/tokens`: TypeScript constants feeding Tailwind v4 `@theme` and RN `StyleSheet` (SDD 10); OKLCH for CSS, sRGB hex exported for RN by a committed script with a round-trip test.

| Token | Light | Dark | Use |
| --- | --- | --- | --- |
| `bg` / `fg` | `oklch(0.99 0.003 90)` / `oklch(0.2 0.01 260)` | `oklch(0.17 0.01 260)` / `oklch(0.94 0.005 90)` | surface, text |
| `muted`, `surface`, `border` | `0.5 0.01 260`, `0.96 0.005 90`, `0.88 0.005 260` | `0.7 0.01 260`, `0.22 0.01 260`, `0.32 0.01 260` | secondary |
| `accent.game/movie/series/book` | `0.5 0.2 290`, `0.52 0.19 25`, `0.5 0.15 220`, `0.5 0.13 150` | `0.72 0.16 290`, `0.72 0.16 25`, `0.72 0.13 220`, `0.72 0.12 150` | category accents |
| `score.low/mid/high/none` | `0.66 0.19 25`, `0.78 0.16 85`, `0.62 0.17 145`, `0.7 0.01 260` | same | bands < 5.0, 5.0–7.4, ≥ 7.5, null |
| `star` | = `score.mid` | same | filled star |
| `danger` | `0.5 0.19 25` | `0.72 0.17 25` | destructive |
| spacing | 4 px base: 0, 1, 2, 3, 4, 6, 8, 12, 16 steps | | |
| type scale | 12, 14, 16, 18, 20, 24, 30, 36 px; serif display, sans text, tabular mono for numbers | | |
| radii | 4, 8, 12 px, full | | |
| touch target | 44 px | | |
| icons | `game, movie, series, book, star, star-half, star-empty, search, user, bell, edit, external, close, check` | | inline SVG web, `react-native-svg` app |

Values from `legacy/apps/web/app/globals.css`; legacy 0–100 bands rescaled to stars. Dark mode: `prefers-color-scheme`, `theme` cookie override; text contrast ≥ 4.5:1 in both. Fonts: self-hosted WOFF2, Latin + Latin Extended, text face preloaded.

### 7.6 Accessibility and navigation

Skip link to `#main` first. On each Inertia `navigate` (not partial reloads) focus moves to `<h1 tabindex="-1">` and a polite live region announces the title. `lang` per locale, landmarks, visible focus, no motion under `prefers-reduced-motion`.

### 7.7 SEO

- Inertia `<Head>` from `seo`: title ≤ 60 chars, description ≤ 160, OG/Twitter card with the 600 px cover.
- JSON-LD data block (`<` escaped): `VideoGame`, `Movie`, `TVSeries`, `Book` (editions as `workExample` with `isbn`), `Person`, `Organization`; `aggregateRating {ratingValue, bestRating: 10, worstRating: 0.5, ratingCount}` only when `score` is not null.
- Sitemaps: index → `/sitemaps/static.xml`, `items-{type}-{n}.xml`, `entities-{kind}-{n}.xml`, app handler (SDD 01 open issue 8). An entity is listed once per category under that category (`/{section}/{entity path}/{key}`) and indexable there only with ≥ 3 public items of the category, else `noindex, follow`. Merged slugs 301.
- `robots.txt`: disallow `/admin`, `/profile`, `/me`, `/billing`, `/suggest`, `/suggestions`, `/library`, `/lists`, `/login`, `/signup`, `/*?cursor=`; `Sitemap:` line; search stays crawlable so its `noindex` is seen.
- `/llms.txt` and `/llms-full.txt` (vukem `llmstxt`, <https://vukem.dev/docs/digging-deeper/llms-txt>, `routes/llms.go`): what HexaScore is, the four catalogue sections, the API docs, pricing, terms and privacy; `llms-full.txt` streams the API docs Markdown (`resources/docs/*.md`), never item pages. Both are outside the scrape guard.
- Analytics and verification (vukem `analytics`, <https://vukem.dev/docs/digging-deeper/analytics>): `{{analytics}}` in `app.html` before the Vite tags renders Umami (cookieless, always on when `UMAMI_WEBSITE_ID` is set), GA4 in Consent Mode v2 (denied until the cookie banner's Accept, ADR 0023) and the Search Console and Bing verification metas. All off while their variables are unset.
- IndexNow (vukem `indexnow`, <https://vukem.dev/docs/digging-deeper/seo-integrations>): the `seo.indexnow_changes` job (every minute) reads `api_changes` after `indexnow_cursor` and the indexable entities updated since the last run, skips score-only updates and queues one `indexnow.Job` with the en, pt-br and es URLs. vukem serves the key at `/{INDEXNOW_KEY}.txt`. The Bing Webmaster API client is configured from `BING_WEBMASTER_API_KEY` (`App.Config.BingWebmaster`); no admin page uses it yet.

### 7.8 Security headers and CSP

vukem defaults (<https://vukem.dev/docs/security/security-headers#what-every-response-carries>) plus, configured in `config/http.go` through `SecureHeadersConfig.CSPSources`:

```text
default-src 'self'; script-src 'self' 'nonce-{n}' 'strict-dynamic'; style-src 'self' 'nonce-{n}';
img-src 'self' data: https://{covers-bucket-host}; font-src 'self'; connect-src 'self';
frame-src https://www.youtube-nocookie.com; worker-src 'self'; object-src 'none'; base-uri 'self';
form-action 'self' https://checkout.stripe.com https://billing.stripe.com https://accounts.google.com;
frame-ancestors 'none'; upgrade-insecure-requests
```

- `'strict-dynamic'` is one of the keywords `CSPSources` accepts (<https://vukem.dev/docs/security/security-headers#what-cspsources-refuses>).
- No `'unsafe-inline'` styles (vukem refuses it): SSR markup carries no `style=`; dynamic styles go through CSSOM after mount; CI `rg -n "style=\{\{" resources/js` against an allowlist.
- `/docs/reference`: `apidocs`' own nonce CSP (<https://vukem.dev/docs/rest-apis/openapi#api-reference-with-scalar>; nonce handling of Scalar styles: local `vukem/docs/security.md` A02 [SITE PAGE MISSING]) + `connect-src https://api.hexascore.app`. Turnstile (ADR 0016 amendment) adds `https://challenges.cloudflare.com` to `script-src`/`frame-src` only when both keys are set; the analytics providers add their origins the same way (vukem fills `CSPSources`; `SecureHeaders` appends to it, never replaces it).
- Extra: legacy `Permissions-Policy` list (`legacy/apps/web/lib/security-headers.ts`), `Cross-Origin-Resource-Policy: same-site`, HSTS without `preload` (vukem never adds it, <https://vukem.dev/docs/security/security-headers#lifetime>; legacy had it, preload listing dropped).
- CSRF: Fetch Metadata `http.CrossOriginProtection` + `SameSite=Lax` (<https://vukem.dev/docs/basics/csrf#how-a-request-is-checked>); token fallback on for browsers without Fetch Metadata (SDD 03, SDD 13 §6; consistency pass 2026-10-28); `/stripe/webhook` bypassed (SDD 05).

### 7.9 Notifications and flash

Bell badge from `notifications.unread` (partial index, caps at "99+"). Rows render `t("notifications." + kind, payload)` with links from payload ids, never HTML (SDD 14 §5, §7.5); opening posts `/read`. Flash `status`/`error` → toast in `role=status`/`role=alert`, 6 s, paused on hover/focus.

### 7.10 Build and dev workflow

- Bun installs (`bun.lock`); Node LTS runs SSR and Vite only. `vukem serve` runs the app plus `npm run dev` with HMR (<https://vukem.dev/docs/cli/reference-new#vukem-serve>).
- `vukem build` = `vite build` through the detected package manager (`bun x vite build`) + `go build` (<https://vukem.dev/docs/cli/reference-new#vukem-build>); the SSR bundle (`vite build --ssr resources/js/ssr.tsx --outDir bootstrap/ssr`, `ssr.noExternal: true`, no `node_modules` at runtime) is a separate `bun run build:ssr` [NOT IN VUKEM DOCS] (VF-027), a permanent step of SDD 11's build pipeline (open issue 09-7). Production: `vukem ssr:start` supervises it, restart delay 1 s (<https://vukem.dev/docs/cli/reference-runtime#vukem-ssrstart-and-vukem-ssrcheck>; SDD 11).
- Approved dependencies (D-22; exact pinned versions, no `^`, no pre-releases): Biome, Vitest, Testing Library, Playwright, axe-core, Lighthouse CI (dev-only) and Base UI (runtime). Any other npm dependency needs the owner's yes and an ADR.
- Lint/format: **Biome** (legacy choice; one binary, tabs by default per `.editorconfig`, a11y and hook rules built in); ESLint rejected for its plugin sprawl. Typecheck: `bun run typecheck` (`tsc`).

## 8. Edge cases

- Cookie `pt-BR` on a shared `/games/zelda` link: 307 to `/pt-br/games/zelda`; switcher choice en sets cookie `en`, no redirect.
- Missing key: the client shows the key itself, as the server does (<https://vukem.dev/docs/frontend/inertia-i18n#step-by-step>); a CI test diffs key sets of the three catalogues.
- Deploy mid-session: 409 → full reload; unsaved rating selections restored from `sessionStorage` (10 min).
- No cover: empty 2:3 frame with `width`/`height`; no CLS.
- Tampered cursor: 400 `invalid-cursor` on the `Error` page, link to page one.
- Private profile: 200 with nickname only, `noindex`, absent from sitemap and JSON-LD (14-5).
- Deferred group fails: `Defer(...).Rescue()`, retry button.

## 9. Acceptance criteria

- [ ] `curl -s $WEB/games/<slug> | grep -c '<h1'` → 1; same for `/pt-br/…`, `/es/…`.
- [ ] `curl -sI -H 'Cookie: locale=pt-BR' $WEB/games/<slug>` → 307, `Location: /pt-br/games/<slug>`.
- [ ] `curl -s $WEB/games/<slug> | grep -o 'hreflang="[^"]*"' | sort -u` → `en`, `es`, `pt-BR`, `x-default`.
- [ ] `curl -s $WEB/sitemap.xml` → index; every child ≤ 40,000 `<url>`; no entity with < 3 items.
- [ ] `curl -sI $WEB/` → CSP with `'nonce-` and `'strict-dynamic'`, no `unsafe-inline`; `X-Frame-Options: DENY`.
- [ ] `curl -s $WEB/up?checks=1` → `ssr: ok`; after killing the SSR process → `ssr: fail` and `hexa.ssr.fallbacks` > 0 (SDD 12 §6.2).
- [ ] `go test ./app/http/... -run TestInertiaPages` → every row of §6.2 asserts its component and props (anonymous and logged in).
- [ ] `bunx vitest run resources/js/components/rating` → passes keyboard, pointer, ARIA cases.
- [ ] `bunx playwright test` (qa/e2e) → all specs pass in 3 locales; axe 0 serious/critical; 0 CSP violations.
- [ ] `bunx lhci autorun` → assertions of §4 pass for home, `/games`, a game, a book, a person (mobile).
- [ ] `bunx biome ci resources/js packages && bun run typecheck` → 0 errors.
- [ ] `go test ./app/i18n/... -run TestPluralGolden` → Go and TS outputs identical.

## 10. Test strategy

- Unit (Vitest): star widget, `t` formatter, `CoverThumb`, dialog helper, token conversion.
- Go integration (Postgres): `AssertInertia` per page (<https://vukem.dev/docs/testing/overview#inertia-pages>), locale redirects, ETag/304, sitemap, robots, JSON-LD rules, cache invalidation.
- E2E (Playwright + axe), ports of `legacy/qa/e2e/specs/`: `admin`, `api-keys`, `auth`, `catalog`, `contact`, `docs`, `headers`, `item`, `not-found`, `pricing`, `profile`, `public-pages`, `ratings`, `search`, `social` (lists and reports; feed/follows deferred, D-06), `time-to-beat`; new: `i18n`, `seo`, `csp`, `notifications`, `suggestions`, `keyboard`, `ssr-nojs`. Mobile and desktop projects.
- Performance: Lighthouse CI in the QA gate (D-21); k6 TTFB. Security: ZAP baseline (0 High), `TestSecurity_*` (SDD 13). Visual regression optional, not a gate.

## 11. Open issues

1. No URL-prefix source in `i18n.Middleware` [NOT IN VUKEM DOCS]. VF-024; decided: app middleware resolves the locale from the prefix before vukem's middleware. Owner: vukem.
2. HTML ETag/304 vs CSP nonces. Decided (09-2): no ETag on HTML documents (§7.2).
3. No SSR output cache hook [NOT IN VUKEM DOCS]. Owner: vukem. Closes with `SSRConfig.Cache` or k6 proof of TTFB p75 ≤ 400 ms without it. Mitigated by `RenderCache` (§7.2), k6 `qa/k6/pages-ttfb.js` p75 ≈ 90 ms.
4. `ScrollMeta` is documented with page numbers; opaque string cursors [NOT VERIFIED] (SDD 07 open issue 4). Owner: vukem. Closes when `Catalog/Index` ships.
5. Flat catalogues for mobile. Decided (09-5): shared flatten step in `packages/i18n` (§7.1); SDD 10 consumes it.
6. `theme` for `data-theme` is [NOT IN VUKEM DOCS] (`inertia.RootData` has `.Inertia`, `.Head`, `.Page`, <https://vukem.dev/docs/frontend/inertia#the-pieces-of-an-inertia-app>). VF-028; decided: a view function reading the `theme` cookie (§6.1). Owner: vukem.
7. SSR bundle build not covered by `vukem build` [NOT IN VUKEM DOCS]. VF-027; decided: permanent `build:ssr` step in SDD 11. Owner: vukem.
8. Entity sitemap threshold. Decided (09-8): ≥ 3; SDD 01 FR-01-12 amended; revisit with Search Console data at launch + 60 days.
9. No 300 px cover. Decided (09-9): 600 px + 96×144 with `srcset`; a 300 px variant only if item-page LCP > 2.0 s.
10. New dependencies. Approved: D-22 (§7.10).
11. INP lab proxy. Decided (09-11): TBT ≤ 200 ms in the lab; field data via the web-vitals beacon to `/app/v1/client-metrics` (NFR-09-2).
12. `/admin` UI. Decided (09-12): moderation pages are vukem `html/template` views in the admin layout, not Inertia (§6.2); `Resource.Action` still requested (VF-008).
