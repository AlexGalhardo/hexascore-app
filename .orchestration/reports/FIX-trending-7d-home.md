# FIX-trending-7d-home Home shows the 7-day trending, 3 per type
status: concluida
branch: agent/senior/FIX-trending-7d-home @ 3ee5afa (stacked on P6.D.3.1 @ 800bbd9)
files: docs/sdd/08-recommendation.md, app/recommend/trending/snapshot.go, app/controllers/home.go, resources/js/pages/Home/Index.tsx, resources/lang/{en,pt-BR,es}.json, tests/home_highlights_test.go, app/recommend/trending/snapshot_integration_test.go
acceptance:
- `go test -tags integration ./tests -run 'TestHome_'` -> ok (TestHome_Trending7d: 4 active items today, 1 with larger activity 20 days ago; home lists 3, ranks 1..3 gapless, old item absent, empty type has 0 items)
- `go test -tags integration ./app/recommend/trending` -> ok (TestWindowSevenDays: 7-day window excludes the 20-day-old burst, the 30-day window still ranks it; existing snapshot tests unchanged)
- `golangci-lint run` controllers and trending: 0 issues; biome clean; tsc started, see note
decisions:
- SDD 08 amended first (section 1, section 6 home props, new "Home window" paragraph in 7.2): home = last 7 days, 3 per type, same formula, eligibility and prior fill; the 30-day snapshot and both APIs are unchanged.
- No new table or column: `trending.Window` runs the snapshot candidate query with a 7-day window on demand; the home caches the result 5 minutes per locale (`home:trending7:v1:{locale}`). A quiet week shows `prior` entries labelled "Popular" (all prior: "Popular {type}").
- `trending` is an eager Home prop (SSR, SEO), computedAt = compute time.
- Strings: home.trendingWeek, home.popularType, home.popular in en, pt-BR, es.
risks: SDD 05 listTrending row was NOT edited: this branch is based before commit 5f290ce (main already has the 30-day text there), so the specialist should add the one-line note "the home alone shows 3 per type over 7 days (SDD 08 section 7.2 Home window)" at merge. Eligibility thresholds (T >= 30, A >= 5) are not scaled to 7 days, so a quiet week falls back to prior entries. Stacked on the P6.D.2.2 home.go changes.
changelog: Changed: the home page shows the trending of the last 7 days (3 per type); the 30-day API list is unchanged.
roadmap: none (FIX row)

## Review (especialista, 2026-10-31)

verdict: CORRIGIDO_E_APROVADO
re-run: on 860b48b: `go test -tags integration ./app/... ./database/... ./routes/... ./bootstrap/...` -> ok; `./tests -run TestHome|ForYou|Inventory|Recommend|Trending|Similar|Security_A01` -> ok after the inventory fix; tsc ok
checked: sdd 08 7.2 Home window; i18n en/pt-BR/es
fixes: 1. SDD 05 listTrending row now notes the home 7-day window (565f4db); 2. the superseded P6.C.3.1 home strip removed in the merge: TrendingSection(.test).tsx, tests/home_trending_test.go, home.trending.* strings; the old changelog line replaced. Note: app/jobs/compute_trending.go still forgets trending.CacheKey entries the home no longer reads (harmless, 5-minute TTL on home:trending7)
gate on main after merge: see batch gate in handoff/especialista.md (scripts/ci/local.sh on 127eeeb)

PRONTO_PARA_MERGE agent/senior/FIX-trending-7d-home (merged 565f4db)
