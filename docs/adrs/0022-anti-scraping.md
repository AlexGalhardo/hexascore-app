# ADR 0022: server-side anti-scraping for the public pages

Status: accepted (2026-11-01). Decisions: Q-A18 (owner, 2026-11-01); ADR 0016; SDD 13 §7.1.1, SDD 12 §6.2; board row SEC-anti-scraping.

## Context

The catalogue pages (home, listings, item, entity, public lists) are public, server-rendered and cached, and every Inertia visit returns the same page JSON. A script can walk the sitemap and harvest the catalogue that the paid API (SDD 05) sells. The owner wants only search engines to index pages and AI crawlers to keep access (Q-A18 a), with no copy blocking in the UI (Q-A18 b). There is no CDN bot management: Cloudflare (P2) is still owner-gated, and vukem has no Turnstile hook (VF-003).

## Decision

A server-level middleware, `ScrapeGuard` (`app/http/middleware/scrape_guard.go`), runs in production in front of the view counter, gzip and `RenderCache`, so a cached page is limited too. It covers GET and HEAD page views and `X-Inertia` visits. It skips `/api`, `/app`, `/build`, `/covers`, `/og`, `/up`, `/docs`, `/.well-known`, `/sitemaps`, `/challenge`, `/search/suggest` and any path with a file extension. In order:

1. **Allowed crawlers** (`app/botverify`). A User-Agent naming an allowed crawler is checked by reverse DNS plus forward confirmation against the operator's documented domains:
   - Googlebot: `googlebot.com`, `google.com`, `gae.googleusercontent.com`. Google also lists `googleusercontent.com` for user-triggered fetchers, but `bc.googleusercontent.com` names any Google Cloud customer VM, so only the `gae.` subdomain is accepted.
   - Bingbot: `search.msn.com`. Applebot: `applebot.apple.com`. YandexBot: `yandex.ru`, `yandex.net`, `yandex.com`.

   Lookups share one 1.5 s timeout (`net.DefaultResolver`). The verdict is cached in the shared cache store: 24 h when settled, 1 min after a DNS failure. A verified crawler is exempt from the limits and gets the same response as a browser.
2. **Spoofed crawlers.** When DNS answers and the IP is not the named crawler's, the request gets 403 and the log line `alert=scrape reason=spoofed_crawler`. A DNS timeout or SERVFAIL only means "unverified": the client keeps the anonymous limits, so a resolver outage never blocks Googlebot.
3. **Operators without rDNS.** DuckDuckBot, GPTBot, ClaudeBot, CCBot and PerplexityBot publish IP lists, not rDNS, so they are matched by User-Agent and keep the anonymous limits. Google-Extended and Applebot-Extended are robots.txt tokens only. Fetching the IP lists on a schedule is a follow-up, if those crawlers ever need the exemption.
4. **Honeypot.** Every page carries `<a href="/data/export" hidden aria-hidden="true" tabindex="-1" rel="nofollow">` in `resources/views/app.html`. It is never rendered, focused or announced, and robots.txt disallows the path. An unverified client that requests it is denied pages from its IP key (per address for IPv4, per /64 for IPv6) for 24 h (`scrape:ban:<key>` in the cache store). The hit logs `alert=scrape reason=honeypot`. A verified crawler gets 404 and no ban.
5. **Tool User-Agents** get 403. These are an empty UA and any UA containing python-requests, python-urllib, aiohttp, httpx, curl, wget, Scrapy, Go-http-client, axios, node-fetch, HeadlessChrome, PhantomJS or libwww-perl.
6. **Page limits per IP** (`app/http/limits`, vukem limiter on the shared store, 429 with `Retry-After`):

   | Visitor | Per minute | Per hour |
   | --- | --- | --- |
   | Anonymous | 60 | 600 |
   | With a session cookie | 300 | 3000 |

   A reader opens a page every few seconds, so 60/min still allows a burst of tabs, and 600/h caps a steady crawl at one page every 6 s. The cookie is only checked for presence, because the guard runs before the session middleware. A forged cookie buys the larger budget and nothing else, still per IP. The store counts fixed windows; the minute and the hour windows together give the burst and the sustained rate.

Every 403 goes through the app's error handler (problem+json for JSON clients) and points to `https://hexascore.app/docs`. A ban store that cannot be read answers 503 (fail closed). `RenderCache` never replays `RateLimit-*` headers, and its key does not vary by UA.

Other parts of the decision:

- robots.txt names the allowed crawlers in the `*` group and adds `Disallow: /data/export`.
- The item, entity and public list pages send only the props they render. The Wikidata id, external ids, credits, affiliate links, provenance, formula version, list version and timestamps, the added dates and raw cover keys are gone.
- The Terms gain an anti-scraping clause in en, pt-BR and es.
- The QA stack (production image) sets `PAGE_LIMIT_SCALE=1000` so one host can run e2e and Lighthouse. Production leaves it unset. Any value that is not a positive integer refuses to boot.

## Consequences

- A determined scraper with residential proxies and a real browser is slowed down, not stopped. Cloudflare bot management and WAF (P2, owner-gated) is the stronger layer, and Turnstile step-up waits for the keys and for a vukem hook (VF-003). Both are owner items.
- Shared IPs (CGNAT, offices) share a budget, and a trap hit from one of them bans the whole IP key for 24 h. The 403 points to the API docs.
- Each page view costs up to three shared-store round trips (ban read, two window increments), and a first visit from a crawler IP costs one DNS check (at most 1.5 s, once per 24 h).
- `curl https://hexascore.app/` returns 403. Smoke checks use `/up`.
