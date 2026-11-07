# E2E coverage

Routes come from `routes/*.go`. Auth pages (`/login`, `/signup`, `/2fa`) are out of scope here (under redesign).

| route                                                                               | covered by                                                     | gap                                                   |
| ----------------------------------------------------------------------------------- | -------------------------------------------------------------- | ----------------------------------------------------- |
| `/`                                                                                 | smoke, csp, headers, hydration, crawler                        | none                                                  |
| `/games`, `/movies`, `/series`, `/books`                                            | public-pages, ssr-nojs, catalog, crawler                       | none                                                  |
| `/games/{slug}` (item)                                                              | item, crawler (when linked)                                    | none                                                  |
| `/{section}/{first}/{second}[/{third}]` (sub listings)                              | none                                                           | no spec                                               |
| `/games/{slug}/changes`, `/books/{slug}/editions/{id}/changes`                      | none                                                           | no spec                                               |
| `/search`, `/search/suggest`, `/search/events`                                      | search, seo, public-pages                                      | `POST /search/events` not covered                     |
| `/api/search/suggest` (legacy)                                                      | none                                                           | no spec                                               |
| `/api/items/{id}/ratings`                                                           | api-contract                                                   | covered: 404 unknown/malformed id, 400 bad page       |
| `/go/{id}`                                                                          | api-contract                                                   | 404 for unknown id only; redirect target untested     |
| `/covers/{key}`, `/og/{kind}/{file}`                                                | none                                                           | no spec                                               |
| `/pricing`                                                                          | pricing, perf-budget                                           | none                                                  |
| `/docs`, `/docs/{slug}`                                                             | docs, public-pages, not-found                                  | none                                                  |
| `/about`, `/how-score-works`                                                        | crawler (when linked)                                          | no direct spec                                        |
| `/terms`, `/privacy`                                                                | public-pages, ssr-nojs                                         | none                                                  |
| `/contact` (GET, POST)                                                              | contact, public-pages                                          | none                                                  |
| `/challenge`                                                                        | none                                                           | no spec (proof-of-work endpoint)                      |
| `/csp-reports`                                                                      | csp (policy only)                                              | POST not covered                                      |
| `/turnstile`                                                                        | none                                                           | no spec                                               |
| `/robots.txt`, `/sitemap.xml`, `/sitemaps/{file}`                                   | seo                                                            | `/sitemaps/{file}` child not covered                  |
| `/.well-known/assetlinks.json`                                                      | none                                                           | no spec                                               |
| `/recover-password`                                                                 | public-pages                                                   | flow not covered (Senior)                             |
| `/profile`, `/profile/details                                                       | password                                                       | privacy                                               | notifications` | profile, account helpers | per-route PATCH/PUT validation not covered |
| `/profile/sessions/*`                                                               | none                                                           | no spec                                               |
| `/profile/api`, `/profile/api/keys*`                                                | api-keys                                                       | none                                                  |
| `/profile/export`, `/me/export`                                                     | none                                                           | no spec                                               |
| `/user/two-factor*`                                                                 | passkeys (partial)                                             | no direct spec                                        |
| `/u/{nickname}`, `/u/{nickname}/lists/{id}`                                         | profile, social                                                | none                                                  |
| `/library`, `/items/{id}/edition`                                                   | library                                                        | edition list/put/delete not covered                   |
| `/for-you`                                                                          | none                                                           | no spec                                               |
| `/lists*`                                                                           | lists, lists-add-item                                          | delete and reorder edge cases partial                 |
| `/suggest/new/{domain}`, `/suggest/{type}/{id}`, `/suggestions*`, `/me/suggestions` | community, community-create, community-draft, community-status | withdraw route partial                                |
| `/billing*`                                                                         | none                                                           | no spec (Stripe, needs test keys)                     |
| `/auth/{provider}/redirect                                                          | callback`, `/auth/google/confirm`                              | social (partial)                                      | no direct spec |
| `/me/notifications`                                                                 | notifications, social                                          | none                                                  |
| `/reports` (POST)                                                                   | reviews/community partial                                      | direct POST not covered                               |
| `/admin/*`                                                                          | admin, community (moderation)                                  | ban, scoring, offers, trending, merge not covered     |
| `/api/v1/openapi.json`                                                              | api-contract, docs                                             | none                                                  |
| `/api/v1/*` (key guard)                                                             | headers, not-found                                             | authenticated endpoints need a key (PoW-gated)        |
| `/app/v1/*` (bearer)                                                                | api-contract (401 only)                                        | authenticated happy path and 404/400/422 need a token |
| `/{path...}` unknown                                                                | not-found                                                      | none                                                  |
| CSRF on session POST                                                                | security-hardening (needs docker mail access)                  | positive control with the real token not asserted     |
| Security headers                                                                    | headers, csp, security-hardening                               | none                                                  |
| Crawler: every reachable public page and internal link                              | crawler                                                        | pages behind login are not crawled by design          |
| Performance budget (TTFB, load)                                                     | perf-budget                                                    | LCP and script size remain Lighthouse-only            |

Status: round 5 (auth, passkeys, newsletter, 6 projects): 60 passed, 6 failed (all mobile, see Found bugs). Earlier green run: crawler, api-contract, security-hardening, perf-budget and smoke, 84 passed.

## Found bugs

1. Footer newsletter input loses a typed value when filled right after load.
   - Route: `/` (footer `NewsletterForm`), reproduced on desktop-en.
   - Expected: the typed address stays in `input[name=email]` and is posted.
   - Actual: the value is empty immediately after `fill`; the POST body has `"email": ""`. It works after `networkidle`. Suspected hydration re-mount of the form; root cause not confirmed.
   - Spec: `newsletter.spec.ts` waits for `networkidle` as a workaround and marks it with a comment.

2. Profile security buttons are unclickable on mobile (360 px).
   - Route: `/profile` (account security section), mobile-en, mobile-pt-BR, mobile-es.
   - Expected: "Enable 2FA" and the passkey "add" button receive clicks.
   - Actual: Playwright reports the button intercepted by other sections (profile section nav `<ul>`, password section label, details section `<p>`). Visual confirmation pending.
   - Spec: `auth.spec.ts` (TOTP) and `passkeys.spec.ts` fail on mobile; tests are not weakened.
