# SEC-anti-scraping Server-side anti-scraping of the public pages (Q-A18)
status: concluida
branch: agent/senior/SEC-anti-scraping @ 8b524131 (wt/senior-b-anti-scraping)
files: app/botverify/ (new), app/http/middleware/scrape_guard.go (+test), app/http/limits/limits.go, app/http/middleware/render_cache.go, routes/web.go, app/controllers/{robots,item,entity,lists_pages,legal}_controller.go, app/seo/seo.go, resources/views/app.html, resources/lang/{en,pt-BR,es}.json, docker-compose.yml, docs/adrs/0022-anti-scraping.md + README, docs/sdd/12-observability.md, docs/sdd/13-security.md, vukem-feedback.md, tests/{content_pages,item_show,entity_show,user_pages}_test.go, tests/i18n/terms_test.go
acceptance:
- red first: `go test ./app/botverify/ ./app/http/middleware/` → build failed (no impl); `go test ./tests/i18n/` → FAIL missing terms.sections.3.body.1; integration props → FAIL `prop "item.wikidataQid" present`, `"list.version" present`
- `go test ./app/botverify/ ./app/http/... ./app/seo/ ./tests/i18n/` → ok (TestSecurity_A06_ScrapeToolUADenied, _ScrapeAPIRoutesUnaffected, _ScrapeSpoofedGooglebot, _ScrapeVerifiedGooglebot, _ScrapePageRateLimit, _ScrapeHoneypotBan, TestSecurity_A10_ScrapeBanStoreDown503, TestPageLimitScale, fake resolver only)
- `go test ./tests/ -run 'TestRobots|TestCanonicalHost|Security|Inventory'` → ok (robots names bots, Disallow /data/export, Sitemap kept)
- `TEST_DATABASE_URL=… go test -tags integration ./tests/ -run 'TestInertiaPages_Catalog_Show|TestInertiaPages_Entity_Show|TestInertiaPages_UserLibraryLists|TestOG|Comments|Related|Ratings'` → ok
validation: golangci-lint (-j 2) touched pkgs 0 issues; markdownlint touched docs ok; no TS change; browser not checked (only a hidden <a> in app.html)
changelog: Added: server-side anti-scraping for public pages (verified search crawlers, tool UA deny, per-IP page limits, honeypot ban, robots allow list, slimmer page props, Terms clause) — ADR 0022
roadmap: no box
risks / follow-ups:
- Guard is production-only (QA image included); QA compose sets PAGE_LIMIT_SCALE=1000 for e2e/Lighthouse. Reviewer: run Lighthouse + e2e on a QA stack once.
- "Signed-in" = session cookie present (guard runs before sessions); forged cookie only buys the 300/min per-IP budget.
- AI crawlers and DuckDuckBot publish IP lists, not rDNS: UA-matched + anonymous limits (shortcut; IP-list fetch is a follow-up). Google accepts gae.googleusercontent.com only (bc. = customer VMs).
- Turnstile step-up skipped (no vukem hook VF-003, no keys) → owner item. VF-NEW-SEC-anti-scraping filed (raw client IP helper).
- ADR number 0022 taken as highest+1 at branch time; renumber if another branch merged a 0022 first.

## Review (especialista, 2026-11-01)

verdict: CORRIGIDO_E_APROVADO
re-run: `go test ./app/botverify/ ./app/http/... ./app/seo/ ./tests/i18n/ ./routes/` ok; `go test -tags integration ./tests -run 'TestRobots|TestCanonicalHost|Security|Inventory|TestInertiaPages|TestOG|Comments|Related|Ratings|TestItem|TestEntity|TestList|TestUserPages|TestFriendsScore|TestSSR'` ok; vitest Catalog, Entity, User pages 4 files, 8 tests; golangci-lint (`-j 2`) 0 issues; `golangci-lint fmt --diff` clean; markdownlint clean
QA stack (production image, own project `hexascore-qa-esp20`, `PAGE_LIMIT_SCALE=1000`): browser UA `/games` 200, curl UA 403, unverified Googlebot UA 403, `/up` with curl UA 200, robots.txt `Disallow: /data/export`, the honeypot renders as `<a href="/data/export" hidden aria-hidden="true" tabindex="-1" rel="nofollow">` (display none, out of the tab order and the accessibility tree; axe found nothing on it). e2e smoke (smoke, public-pages, item, seo; 4 projects): every failure is pre-existing on main and outside this change: `item.spec.ts:58` waits for a heading the design v1 item page removed, and four `/docs` pages fail axe `link-in-text-block` (`/docs` is exempt from the guard); new row QA-e2e-stale-item-docs. Lighthouse not measured: host runs are invalid while Kaspersky injects (FIX-lighthouse-budgets-main); the honeypot is a hidden, nofollow, robots-disallowed link, so neither the SEO nor the a11y audits see it.
checked: 1. vukem writes a session only when it is modified (session/session.go "loaded lazily and written only when modified"), but any client can keep or forge a `vukem_session` cookie, so "a session cookie is present" was not a login. 2. props: the item, entity and list pages, SSR and OG cards read none of the removed fields (`grep` of `resources/js` for wikidataQid, externalIds, affiliateLinks, provenance, releasePrecision, formulaVersion, list version and dates); `/app/v1` uses its own controllers. 3. ADR 0022 does not collide (main ends at 0021).
fixes:
1. the signed-in page budget (300/min, 3000/h) now needs a session that holds the web guard login: `ScrapeGuard.SignedIn` peeks at the session through `app.Sessions().Peek` and checks `auth_web`; `TestScrapeGuardSignedInKeyMatchesVukem` pins the key to `auth.SessionValues`, and the page-limit test checks a forged cookie stays on the anonymous budget (5fc31a86)
2. `VF-NEW-SEC-anti-scraping` numbered VF-065 (entry, index, code comment; the `auth_web` key read is marked with the same id)
PRONTO_PARA_MERGE agent/especialista/SEC-anti-scraping (merged 014825f2)
