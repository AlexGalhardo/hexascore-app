# 01. Web functional test report (2026-11-02)

Executive summary: the web app works end to end. A 353-page crawl of the local build found no 404 or 500; login, sign-up with e-mail verification, rating, status, lists, contact, API keys, locale switch, theme and logout all behave. One serious bug reached production: the HTML render cache ignores the URL locale prefix, so `/pt-br/...` and `/es/...` pages are served in whatever language was rendered first (W-01, P1). The rest are medium and cosmetic items, listed below with reproduction steps.

Environments: local QA stack built from `main` 5e92fcc7 (`hexascore-app:audit`, DemoCatalog seed, `http://localhost:3100`) for every write action; production `https://hexascore.app` (same commit, `x-hexascore-version: 5e92fcc`) read-only, with `curl` and a desktop Chrome. Tools: link crawler (fetch, same-origin BFS, 350-page cap), Chrome DevTools MCP (a11y snapshots, console, network, Lighthouse 12 mobile), Playwright suite (section 5). Console noise from the host antivirus (Kaspersky script injection, CSP inline-style reports) was excluded after confirming the served HTML has zero inline styles.

Severity: P0 blocking · P1 serious · P2 medium · P3 cosmetic.

## 1. Findings

| ID | Sev | Area | URL / repro | Expected | Observed | Evidence |
| --- | --- | --- | --- | --- | --- | --- |
| W-01 | P1 | i18n / cache | Production: `curl -A Mozilla https://hexascore.app/pt-br/games/burgertime`; also `/pt-br/games`, `/pt-br/` after another locale was requested | `<html lang="pt-BR">`, Portuguese copy | `<html lang="en">` on the item page, `lang="es"` + title "Juegos" on `/pt-br/games`; locally, after one `GET /pt-br/`, `GET /es/` and `GET /` answer Portuguese with `X-Render-Cache: HIT` | `app/http/middleware/render_cache.go:119-134` keys on `r.URL.Path`; vukem strips the locale prefix in an earlier server middleware (`vukem/http.go:314-316`, `i18n.StripPathPrefix`), so every locale of a path shares one key. Fix: add the resolved locale (or the recorded source prefix) to `pageKey`; extend `render_cache_test.go` "prefixed" cases to assert distinct entries per locale |
| W-02 | P2 | Mobile UX | `/pt-br/games/ullamco-magna` at 360×740 (any item page) | Title visible near the top | Order is cover, 4 action buttons, details, score card, full rating widget, then the H1 title and synopsis about 1,800 px down; only the breadcrumb names the item above the fold | full-page screenshot 2026-11-02; `resources/js/pages/Catalog/Show` column order on small screens |
| W-03 | P2 | SEO | `/games/does-not-exist`, `/u/nobody-here`, `/lists/<bad-uuid>` (local and production) | 404 page with status 404 (410 for removed items) | `301` to `/` (owner round 2026-11-01). Search engines treat it as a soft 404, the user loses context, and a typo in a shared link lands on the home page | curl: `301 loc=https://hexascore.app/`; `routes/web.go` catch-all `/{path...}` |
| W-04 | P2 | SEO / sharing | Any item page, production | `og:image` with the cover | Only `og:title`, `og:description`, `og:type`, `og:url`, `twitter:card=summary`; no `og:image`, so link previews have no cover although covers and an OG card generator (`/og/{kind}/{file}`, `app/seo/og_image.go`) exist | `curl https://hexascore.app/games/burgertime \| grep og:` |
| W-05 | P2 | Search | First request of a new query, e.g. `curl -A Mozilla 'https://hexascore.app/search?q=witcher'` | Results in the HTML | First response is a 70 KB client shell with no results and empty "{n} results" labels; the second request returns 117 KB with 55 results. No-JS clients and link previews see an empty page; browsers render it client-side after hydration | 4 sequential fetches 2026-11-02; vukem `inertia.ClientShell` (`vukem/inertia/inertia.go:91-103`) |
| W-06 | P3 | Listing | `/games?genre=horror&platform=xbox&sort=-score&per=80&page=2` | "No more results" or redirect to the last page | "4 Games found" header, empty grid, no message; `rel=prev` present, no `rel=next` | DevTools evaluation 2026-11-02 |
| W-07 | P3 | SEO | Same URL | Canonical keeps the filters or points to page 1 of the same filter set | `canonical=/games?page=2&per=80` drops `genre`, `platform` and `sort`, so page 2 of a filtered list canonicalises to page 2 of the unfiltered list | DevTools evaluation |
| W-08 | P3 | i18n | `/pt-br/games/ullamco-magna` | Translated genre labels | "Gêneros: Mystery, Romance" (English) while headings and dates are translated | DevTools evaluation; `app/catalog/genres.go` has translated lists, the item page shows source labels |
| W-09 | P3 | SEO | `/`, `/es/`, `/pt-br/` | `<title>` with a tagline ("HexaScore: community scores for games, movies, series and books") | `<title>HexaScore</title>` | curl |
| W-10 | P3 | Copy | `/pricing` | "1 request per second" | "1 requests per second" (Free and Hobby cards) | DevTools evaluation |
| W-11 | P3 | Copy | Listing count after a filter | "4 games found" | "4 Games found" (capitalised noun mid-sentence) | DevTools evaluation |
| W-12 | P3 | a11y | Header search trigger | Visible label contained in the accessible name | Visible "Buscar…" vs `aria-label="Abrir busca"` (Lighthouse `label-content-name-mismatch`, the only a11y failure) | Lighthouse report in the audit scratch folder, 2026-11-02 |
| W-13 | P3 | Performance | `/profile`, `/profile/api`, `/me/notifications` | Content on first paint | Private pages are client-rendered (by design, SDD 09 §6.1): blank main area for 1-3 s on this host and nothing without JavaScript | DevTools evaluation (`#app[data-page]` absent) |
| W-14 | P3 | SEO | `/search`, `/login`, `/signup`, `/contact` | canonical + description | No canonical and no description (all are `noindex` or disallowed in robots, so impact is minimal) | crawler report |
| W-15 | P2 | Infra | `https://staging.hexascore.app` | Staging reachable | DNS does not resolve (`curl: (6) Could not resolve host`); the Railway `staging` environment exists | curl 2026-11-02; `docs/runbooks/staging.md` owner DNS step |

## 2. Verified working (no finding)

- Crawl of the local build: 353 pages, 341 × 200, 12 × 302 to `/login` (auth-only pages), 0 × 4xx/5xx, 0 pages slower than 1.5 s; sections, entity filters, pagination (`?page=2`), locale prefixes, docs, legal pages, pricing currencies.
- Guest: home carousel (4 slides × 3, autoplay with pause), "more" grids, listing filters (sort, per-page 40/80/120, genre, platform, mode, year, min score), item page (details, synopsis, time to beat, category scores, rating widget disabled with login prompt), search page and suggest (`/search/suggest?q=ull`), language menu (EN/PT/ES, sets `locale` cookie and redirects to the prefixed URL), theme toggle, "More" menu, footer links, cookie banner.
- Auth: login with the demo admin → `/dashboard`; sign-up with nickname/e-mail/password validation hints → "Check your email" → verification link from the log mailer → login works for the new user; logout returns to `/` and `/dashboard` redirects to `/login`; `/admin` as a non-2FA admin redirects to `/profile#two-factor`; `/admin` as guest redirects to `/login`.
- Member actions: rating 6 categories (8.0) saved ("Rating saved."), status "Played" saved, list created ("Audit list", "Bot favourites"), item added to a list ("Added to the list."), library shows the rated item with status and score, API key created (shown once, stored hashed), contact form sent ("Thanks, your message was sent."), profile sections (name, password, 2FA, passkeys, sessions, privacy, notices, export, delete).
- Production headers: gzip (home 266 KB → 44 KB), CSP nonce + `strict-dynamic`, HSTS, COOP/CORP, `X-Frame-Options: DENY`, rate-limit headers, `Cache-Control: public, max-age=60, stale-while-revalidate=300`; robots.txt, sitemap index (4 item sitemaps: 24,873 games, 34,968 movies, 14,911 series, 24,089 books), hreflang en/es/pt-BR/x-default, JSON-LD on item pages, `/up` ok, paid API answers 401 problem+json without a key.
- Lighthouse 12 (mobile, local item page): Accessibility 100, SEO 100, Best Practices 73 (only HTTP on localhost and antivirus console errors).

## 3. Performance notes (production, curl, 2026-11-02)

| Page | Cold | Cached (`x-render-cache: HIT`) | HTML |
| --- | --- | --- | --- |
| `/` | 0.98 s | n/a (edge 60 s cache) | 266 KB raw, 44 KB gzip |
| `/games` | 1.11 s | — | 168 KB raw |
| item page (3 random movies) | 0.39-1.31 s | 0.40-1.74 s | 118-122 KB raw |
| `/search?q=…` | 1.28 s | — | 70 KB shell, 117 KB rendered |

The item HTML carries the full i18n catalogue slice and page JSON (~120 KB raw); the audit host measured TTFB only. Lighthouse performance was not run on this Windows host (invalid per AGENTS.md); CI budgets are LCP 2500 ms and script 170 KB.

## 4. Not tested

- Billing checkout and portal (Stripe placeholders locally, no real payment flows in production by rule).
- Google OAuth and passkeys (need a real provider or authenticator).
- Admin panel pages beyond the 2FA gate (demo admin has no TOTP; enrolling would be a one-off local step the Playwright suite covers in `admin.spec.ts`).
- Community suggestions and the social layer (behind `FLAG_COMMUNITY_SUBMISSIONS` and `FLAG_SOCIAL`, both off).
- Mobile app.

## 5. Playwright suite

Run 2026-11-02 against the local stack (`bunx playwright test -c qa/e2e/playwright.config.ts --workers=2`, 750 tests over 6 projects). Stopped after 66 tests and 25 minutes: every spec that creates an account failed in the helper, not in the app. `qa/e2e/helpers/account.ts:73` fills the sign-up "Name" field (`auth.fields.name`), which commit 64f9b814 (`feat(auth): sign up with the nickname only`) removed from `resources/js/pages/Auth/Signup.tsx`; the helper retries for 120 s per test and times out, so auth, api-keys, lists, profile and the other signed-in specs cannot run. The suite has not been executed in CI since the e2e job was gated on `RUN_E2E` (owner, 2026-10-31), which is how the drift went unnoticed.

| ID | Sev | Area | Repro | Expected | Observed | Evidence |
| --- | --- | --- | --- | --- | --- | --- |
| W-16 | P2 | QA | `bunx playwright test … qa/e2e/specs/auth.spec.ts --project=desktop-en` | green suite | `Test timeout of 120000ms exceeded` in `createAccount` for every signed-in spec | `qa/e2e/helpers/account.ts:71-84`; `git log -- resources/js/pages/Auth` |

Fix: drop the name step from the helper (and any spec asserting the name field), run the `desktop-en` project locally, then re-enable the CI job on release candidates (OQ-14). The manual pass in sections 1-2 covered the flows the suite could not.
