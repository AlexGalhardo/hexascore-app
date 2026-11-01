# IMPL-web-v1-edition-entity Edition and entity pages in design v1

status: concluida
branch: agent/senior/IMPL-web-v1-edition-entity @ 4592fb6 (stacked on agent/senior/IMPL-web-v1-library-lists @ e2525f3; local `main` @ 04ee217 merged in at 466921e, clean)
files: resources/js/pages/Catalog/Edition{,.test}.tsx, resources/js/pages/Entity/Show{,.test}.tsx, resources/lang/{en,pt-BR,es}.json (additive edition/entity keys), app/catalog/entity_page.go, tests/entity_show_test.go
acceptance:
- `bunx vitest run resources/js/pages/Catalog/Edition.test.tsx resources/js/pages/Entity` -> 4/4
- `bun run lint` -> 186 files, no fixes; `bun run typecheck` -> exit 0; `bun run test` -> 35 files / 135 tests passed (on the recent-changes branch, which contains this one)
- `go test -tags integration ./tests/ -run 'TestInertiaPages_Entity_Show$'` -> red first (`prop "items.data" = null, want []`), green after the fix; `golangci-lint run ./app/catalog/` -> 0 issues; `golangci-lint fmt --diff` clean
- Browser (QA :3140): `.orchestration/reports/shots/impl-web/{desktop,mobile}-{light,dark}-_books_ex_labore_aute_editions_01a19466_6b2e_7d32_b19f_6e20f46b02c9.png`, `..._companies_baker_works_llc.png`, `..._franchises_velit.png`. Compared with `web-{desktop,mobile}{,-dark}__edition__default.png` and `__entity__{default,empty}.png`: breadcrumb, accent eyebrow ("Edition · Book", "Company"), 64 px serif h1, ghost "Suggest an edit" flush with the text edge (fixed in 1f883be after the first shots), 2 px ruled fact table, role chips with counts, per-type sections under a 2 px rule with "n of n", average score line, rows with 2:3 cover, serif title, "year · n ratings" and the score badge.
validation: lint ok, typecheck ok, vitest ok, go integration test ok, browser ok (light/dark x 1440/390)
changelog: Changed: the book edition and entity (people, companies, franchises, genres, platforms...) pages follow design v1. Fixed: an entity with no linked titles showed a blank page.
roadmap: no box
risks / follow-ups:
- Found while shooting: an entity without linked public items rendered a blank page (`items.data` was JSON `null` because `EntityPage.Groups` was a nil slice; the page reads `.length`). Pre-existing on `main`; fixed at the source in 4592fb6 with a test. The `_franchises_velit` shots were taken before that fix and are blank; they are retaken with the next QA build (see the docs-legal-errors report).
- The right-hand fact column of the entity frame (founded, country, description with licence) shows only when the entity carries those fields; the DemoCatalog entities have none, so the shots show the single-column layout.
- No new binary assets.

## Review (especialista, 2026-11-01)
verdict: CORRIGIDO_E_APROVADO (one batch, `agent/especialista/IMPL-web-batch3` from the more-like-this tip, which stacks P6.D.1.3, edition-entity, recent-changes, docs-legal-errors)
re-run: `go test ./app/recommend/related ./app/catalog` ok; `go test -tags integration ./tests -run 'TestItemRelated|TestAppRelated|Inventory|TestEntity|TestEdition|TestRecentChanges|TestContent|TestLegal|TestDocs|TestPricing|TestI18n|OpenAPI|TestError'` ok; `scripts/openapi-check.sh` ok and `codegen:client -prefix /app/v1` equals packages/app-client/src/api.ts; biome on the 17 touched TS files clean; `bun run typecheck` ok; vitest on the 9 touched test files 16 pass; golangci-lint `--timeout 15m -j 2` on touched Go packages 0 issues
checked: sdd 08 §6/§7.4 (similar <= 8 with reason, crossMedia <= 6 both directions, public items only, 404 on unknown/hidden, private cache), authz, sql bound, i18n parity (570 keys per locale), docs `dangerouslySetInnerHTML` unchanged (build-time escaped), scope, tests
fixes: 1. routes/web.go conflict with PERF-i18n-page-namespaces: kept main's base list (`terms.`/`policy.` ship only on /terms and /policy through PageNamespaces; dot-less `updatedLabel` stays) (39031103) 2. mobile catalogues recopied for the new `item.*` keys (f70d09a7)
not done: browser QA stack (host memory); in-process reco counters have no exporter (shortcut noted by the author)
gate on main after merge: go build ok; targeted only (section 8.2)
PRONTO_PARA_MERGE agent/especialista/IMPL-web-batch3
