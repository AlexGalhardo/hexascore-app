# IMPL-web-v1-home Home in design v1

status: concluida
branch: agent/senior/IMPL-web-v1-home @ 2689965 (stacked on agent/senior/IMPL-web-v1-item @ 3faf40c, with `main` @ 7eff904 merged in for the FIX-trending-7d-home prop and migrations)
files: resources/js/pages/Home/Index.tsx, resources/js/components/home/TypeSection{,.test}.tsx (new), resources/js/components/catalog/SectionHighlights.tsx (now types and SECTION_PATHS only; the three ranked lists are gone), resources/js/components/CoverThumb.tsx (sizes `trend` 80x120/96x144 and `card` fluid 2:3), resources/js/components/ui/ScoreBadge.tsx (size `xs`), resources/lang/{en,pt-BR,es}.json (additive `home.*`)
acceptance:
- `bunx vitest run resources/js/components/home` → 3/3 (other titles skip trending and duplicates, max 6; three cards with rank, band or "n of 5 ratings", see-all link and title count; empty type renders nothing)
- `bun run lint` → 183 files, no fixes; `bun run typecheck` → exit 0; `bun run test` → 33 files / 125 tests passed
- Browser (QA :3140, `/`): `.orchestration/reports/shots/impl-web/{desktop,mobile}-{light,dark}-_.png` (full page), `...-_-scroll.png` (viewport, after the cover fix), `desktop-light-_-in.png` (logged in: avatar, bell). Each type: accent square, title, "75 titles", "See all games →", three trending cards (QA data has no ratings in the last 7 days, so the server falls back to `basis: prior` and the cards read "#1 popular"), "More games" with six 2:3 covers, year and score.
- Lighthouse 12.6.1 mobile (simulated, `http://localhost:3140`): home perf 0.97, LCP 1728 ms, CLS 0.001; listing `/games` perf 0.90, LCP 2325 ms (1536 ms on an earlier run), CLS 0.001; item perf 0.94, LCP 2416 ms (2269 earlier), CLS 0.002. Desktop: home perf 1.00, LCP 684 ms; item perf 1.00, LCP 525 ms.
validation: lint ok, typecheck ok, vitest ok, browser ok (light/dark x 1440/390, logged in/out)
changelog: Changed: the home page follows design v1: one block per item type with three cards trending over the last 7 days and six other titles, with a see-all link.
roadmap: no box
risks / follow-ups:
- CLS stays at or under 0.002 on every page, so metric-override font fallbacks were considered and not added (no shift to fix). Mobile LCP over 2.0 s on item and listing: the LCP element is the SSR cookie-banner paragraph and LCP is about FCP; the HTML is about 140 KB uncompressed, half of it i18n messages, and HTML is never gzipped (BREACH). PERF-i18n-page-namespaces opened for a senior. Font preload of the two WOFF2 files (fonts start only after the CSS) is the next lever for impl-web if the budget still fails after that.
- "Other titles" = most rated, then top rated, then newest, without the trending ones (no dedicated backend list; reuses the existing highlights props).
- The visible hero was removed per the frame; the page keeps a visually hidden `h1` (home.title) for SEO and headings order. Old keys `home.explore`, `home.trendingWeek`, `home.popularType`, `home.popular` and `catalog.landing.*` are now unused; left in place for the specialist to prune with the i18n parity test.
- The item branch's CoverThumb comment still says "lg 144x208"; this branch corrects it (kept off the item branch to avoid a conflict in the stack).
- No new binary assets.

## Review (especialista, 2026-10-31)

verdict: APROVADO
re-run: batch branch agent/especialista/IMPL-web-batch (shell, search-palette, catalog, item, home merged on main 54105f9; lang conflicts were sibling additions, union-resolved and JSON-validated): `bun run lint` ok, `typecheck` ok, `test` 33 files / 129 ok; mobile catalogues recopied (catalog.filters is now plural, unused on mobile), mobile test 47 / 292 ok; `go build ./...` ok; `go test ./tests/i18n` ok\nchecked: i18n parity (3 locales, no hardcoded copy), no dangerouslySetInnerHTML or inline scripts, suggest fetch encodes params and drops stale answers, localStorage guarded, trailers on youtube-nocookie behind consent, scope, tests\nnot re-checked in a browser by the reviewer (authors' frames in the reports)\ngate on main after merge: bun lint/typecheck/test ok (full local.sh at the end of the batch)\nPRONTO_PARA_MERGE agent/senior/IMPL-web-v1-home
