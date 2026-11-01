# IMPL-web-v1-item Item page in design v1

status: concluida (without "More like this": no `similar` prop yet, row P6.D.1.3 opened for a senior)
branch: agent/senior/IMPL-web-v1-item @ 3faf40c (stacked on agent/senior/IMPL-web-v1-catalog @ bec4e3e)
files: packages/tokens/{tokens.json,tokens.css,tokens.ts}, resources/css/app.css, resources/js/pages/Catalog/{Show.tsx,Show.test.tsx}, resources/js/components/item/{CategoryAverages,ScoreBlock}.tsx (new), item/{RatingPanel,StatusPicker,TimeToBeat,Trailers,AddToList}.tsx, rating/StarRating{,.test}.tsx, CoverThumb.tsx, search/SearchPalette.tsx (backdrop), resources/lang/{en,pt-BR,es}.json (additive), qa/e2e/specs/{item,ratings,social}.spec.ts
acceptance:
- `bunx vitest run resources/js/components/rating resources/js/pages/Catalog` → StarRating 4/4 (20 radios, arrows clamp, Home/End/Delete, hover preview, disabled), Show 2/2 (logged-out, logged-in)
- `bun run lint` → 181 files, no fixes; `bun run typecheck` → exit 0; `bun run test` → 31 files / 120 tests passed (before the Show test was added; 16 files / 42 tests rerun after)
- Browser (QA stack hexascore-qa-implweb, :3140, /games/ad-consectetur-reprehenderit), shots in `.orchestration/reports/shots/impl-web/`: `{desktop,mobile}-{light,dark}-_games_ad_consectetur_reprehenderit.png` (logged out) and `...-rate-in.png` (admin logged in, three categories rated by keyboard, focus ring on Sound): avatar, bell, stars with "7.5 / 10", overall badge, disabled Save until all six, status toggles, Add to list.
- Lighthouse 12.6.1 (`bunx lighthouse@12.6.1 http://localhost:3140/games/ad-consectetur-reprehenderit`): mobile perf 0.92, LCP 2269 ms, CLS 0.001, TBT 183; desktop perf 1.00, LCP 525 ms, CLS 0.002. Mobile LCP is the cookie-banner paragraph and equals FCP (render delay 1.7 s under simulated slow 4G), so the 2.0 s budget miss is first paint of the shell, not this page; see the home row report for the site-wide follow-up.
validation: lint ok, typecheck ok, vitest ok, browser ok (light/dark x 1440/390, logged in/out)
changelog: Changed: the item page follows design v1 (score block with its band and explanation, genre/platform chips, category averages radar, ten half stars per category as a keyboard radio group, status toggles, time to beat rows, trailers behind consent). Added: `hover` and `backdrop` colour tokens in packages/tokens.
roadmap: no box
risks / follow-ups:
- SDD 09 §7.4 says "half-star slider"; the owner-approved v1 frame (rating-widget/states) and this row's brief specify one radio group of 20 options per category. Implemented as the frame; the SDD sentence needs the specialist's one-line update at merge.
- The server still requires all six categories (`app/scoring/write.go`), so Save stays disabled until every category is set; the frame's "3 of 6" overall is shown as a preview with "Rate every category to save."
- "More like this" left out (ROADMAP P6.D.1.3 note: no empty placeholder). Board row P6.D.1.3 added for a senior; the block is a small impl-web follow-up after it.
- Tokens: `hover` = surface 96 % + fg 4 % in OKLab, precomputed per mode; `backdrop` = opaque ink used as `bg-backdrop/50`. app.css no longer derives either. Mobile gains the two names (additive).
- e2e specs updated for the radio group and the status toggles (not run: e2e is outside the gate).
- The "Recent changes" link (existing feature, not in the frame) stays as a ghost link next to "Suggest an edit".
- 3faf40c draws the 120 stars from one shared SVG symbol (`StarSymbol`, rendered once in RatingPanel): item HTML 152 KB -> 140 KB; rerun mobile LCP 2116-2416 ms (simulated runs vary by about 0.3 s).
- No new binary assets.

## Review (especialista, 2026-10-31)

verdict: APROVADO
re-run: batch branch agent/especialista/IMPL-web-batch (shell, search-palette, catalog, item, home merged on main 54105f9; lang conflicts were sibling additions, union-resolved and JSON-validated): `bun run lint` ok, `typecheck` ok, `test` 33 files / 129 ok; mobile catalogues recopied (catalog.filters is now plural, unused on mobile), mobile test 47 / 292 ok; `go build ./...` ok; `go test ./tests/i18n` ok\nchecked: i18n parity (3 locales, no hardcoded copy), no dangerouslySetInnerHTML or inline scripts, suggest fetch encodes params and drops stale answers, localStorage guarded, trailers on youtube-nocookie behind consent, scope, tests\nnot re-checked in a browser by the reviewer (authors' frames in the reports)\ngate on main after merge: bun lint/typecheck/test ok (full local.sh at the end of the batch)\nPRONTO_PARA_MERGE agent/senior/IMPL-web-v1-item
