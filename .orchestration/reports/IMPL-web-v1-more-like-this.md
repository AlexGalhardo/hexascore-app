# IMPL-web-v1-more-like-this "More games like this" and cross-media strips on the item page

status: concluida
branch: agent/senior/IMPL-web-v1-more-like-this @ 96c7fab (stacked on agent/senior/IMPL-web-v1-docs-legal-errors @ b9a2482; agent/senior/P6.D.1.3 @ 1fbe834 merged in at cb97599, clean, because P6.D.1.3 is not on `main` yet)
files: resources/js/components/item/MoreLikeThis{,.test}.tsx (new), resources/js/pages/Catalog/Show.tsx (props `similar`, `crossMedia`, default `[]`), resources/lang/{en,pt-BR,es}.json (additive `item.moreLike.{game,movie,series,book}`, `item.similarReason`, `item.crossMedia`, `item.relation.{based_on,adapted_as}`)
acceptance:
- `bunx vitest run resources/js/components/item/MoreLikeThis.test.tsx resources/js/pages/Catalog` -> 7/7 (similar cards: title link `/<locale>/<section>/<slug>?from=similar`, reason "Shares {entity}", year, locale-formatted score badge; cross-media cards: "Movie · Adaptation" and `?from=crossmedia`; nothing rendered when both lists are empty)
- `bun run lint` -> 192 files, no fixes; `bun run typecheck` -> exit 0; `bun run test` -> 40 files / 143 tests passed
- `go test -tags integration ./tests/ -run 'TestItemRelated|TestAppRelated|TestLegalPages$|TestInertiaPages_Entity_Show$'` -> ok (P6.D.1.3 tests still green on the merged stack)
- Browser (QA :3140; the DemoCatalog has no neighbours for the shot item, so `item_similar` of game `tempor-velit-quis` was set by hand to eight public games and one `item_links` row (a movie based on it) was inserted): `.orchestration/reports/shots/impl-web/{desktop,mobile}-{light,dark}-_games_tempor_velit_quis.png`. Compared with `web-*__item__default.png`: "More games like this" under a 2 px rule with serif heading, a horizontal row of cards (2:3 cover, title, year + score badge, reason line), scrolling sideways with snap on mobile; "In other media" strip with medium and relation.
validation: lint ok, typecheck ok, vitest ok, browser ok (light/dark x 1440/390)
changelog: Added: "More games/movies/series/books like this" (up to 8, with the shared feature) and "In other media" strips on the item page.
roadmap: no box (the data side is P6.D.1.3)
risks / follow-ups:
- Merge order: P6.D.1.3 first (REV-P6.D.1.3 is still todo), then this branch. Without P6.D.1.3 the page still renders (props default to empty arrays, strips hidden).
- The reason shows only the shared entity ("Shares PlayStation"); `reason.kind` (franchise, genre, platform...) is not phrased per kind, to avoid one string per feature role. Per-kind copy is a follow-up if the owner wants it.
- Up to 8 cards in one scroll row (frame shows 5 on desktop): the rest is reached by horizontal scroll, keyboard reachable through the card links. Cover images are lazy (below the fold), so they do not compete with the item cover for LCP.
- No new binary assets.

## Review (especialista, 2026-11-01)
verdict: CORRIGIDO_E_APROVADO (one batch, `agent/especialista/IMPL-web-batch3` from the more-like-this tip, which stacks P6.D.1.3, edition-entity, recent-changes, docs-legal-errors)
re-run: `go test ./app/recommend/related ./app/catalog` ok; `go test -tags integration ./tests -run 'TestItemRelated|TestAppRelated|Inventory|TestEntity|TestEdition|TestRecentChanges|TestContent|TestLegal|TestDocs|TestPricing|TestI18n|OpenAPI|TestError'` ok; `scripts/openapi-check.sh` ok and `codegen:client -prefix /app/v1` equals packages/app-client/src/api.ts; biome on the 17 touched TS files clean; `bun run typecheck` ok; vitest on the 9 touched test files 16 pass; golangci-lint `--timeout 15m -j 2` on touched Go packages 0 issues
checked: sdd 08 §6/§7.4 (similar <= 8 with reason, crossMedia <= 6 both directions, public items only, 404 on unknown/hidden, private cache), authz, sql bound, i18n parity (570 keys per locale), docs `dangerouslySetInnerHTML` unchanged (build-time escaped), scope, tests
fixes: 1. routes/web.go conflict with PERF-i18n-page-namespaces: kept main's base list (`terms.`/`policy.` ship only on /terms and /policy through PageNamespaces; dot-less `updatedLabel` stays) (39031103) 2. mobile catalogues recopied for the new `item.*` keys (f70d09a7)
not done: browser QA stack (host memory); in-process reco counters have no exporter (shortcut noted by the author)
gate on main after merge: go build ok; targeted only (section 8.2)
PRONTO_PARA_MERGE agent/especialista/IMPL-web-batch3
