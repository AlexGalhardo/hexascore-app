# IMPL-web-v1-docs-legal-errors Docs, legal, error and pricing pages in design v1

status: concluida
branch: agent/senior/IMPL-web-v1-docs-legal-errors @ b9a2482 (stacked on agent/senior/IMPL-web-v1-recent-changes @ 9ec0044)
files: resources/js/pages/Docs/Page{,.test}.tsx, resources/js/pages/Legal/Page{,.test}.tsx, resources/js/pages/Error{,.test}.tsx, resources/js/pages/Pricing/Index{,.test}.tsx, resources/lang/{en,pt-BR,es}.json (additive `billing.pricing.{eyebrow,popular,lifetime}`, `terms.{contents,also,otherTitle}`, `policy.{contents,also,otherTitle}`), routes/web.go (one i18n prefix), tests/content_pages_test.go
acceptance:
- `bunx vitest run resources/js/pages/Error.test.tsx resources/js/pages/Legal resources/js/pages/Docs resources/js/pages/Pricing` -> 6/6 (404 search form posts `q` to the localized `/search`, 500 retry button and no search; legal anchors, contents list, link to the other document; docs current page and prev/next; pricing featured plan, lifetime label, whole-unit prices, current currency)
- `bun run lint` -> 192 files, no fixes; `bun run typecheck` -> exit 0; `bun run test` -> 40 files / 143 tests passed (top of the stack)
- `go test -tags integration ./tests/ -run 'TestLegalPages$'` -> red first (`prop "messages.updatedLabel" missing`), green after the fix
- Browser (QA :3140, built from the more-like-this branch, which contains this one): `.orchestration/reports/shots/impl-web/{desktop,mobile}-{light,dark}-_{docs_authentication,terms,pricing,no_such_page}.png`. Compared with `web-*__docs__default`, `__legal__default`, `__error__404`, `__pricing__default`: docs sidebar with small-caps heads, active page on surface with a 2 px ink edge, API reference in primary, "On this page"; eyebrow, 64 px serif h1, muted lede, ruled h2 sections, dark code block, square bullets, prev/next links. Legal: serif h1, "Last updated" small caps + date, italic serif intro, numbered sections under 2 px rules, sticky Contents column with "Also" link (stacks above the text on mobile). 404: big serif status in primary, short rule, serif heading, primary "Back to home" plus a GET search field. Pricing: eyebrow, serif h1, USD/BRL/EUR segmented links (ink fill on the current one), three plans under 2 px rules, Pro on surface with "Most chosen", serif prices, small-caps labels, primary/secondary buttons, ruled coverage block.
validation: lint ok, typecheck ok, vitest ok, go integration test ok, browser ok (light/dark x 1440/390)
changelog: Changed: API docs, terms and privacy, error and pricing pages follow design v1. Fixed: the terms and privacy pages showed the raw "updatedLabel" key instead of "Last updated".
roadmap: no box
risks / follow-ups:
- Found while shooting (pre-existing on `main`): the i18n prefix list had `"updatedLabel."`, which matches no key (`updatedLabel` is a top-level string), so the legal pages printed "UPDATEDLABEL". Fixed to `"updatedLabel"` in `routes/web.go`. senior-b's PERF-i18n-page-namespaces branch (184f71f) still carries `"updatedLabel."` in its new base list: whoever merges second must keep the dot-less entry.
- Namespaces used: docs page `docs.` only; legal pages `terms.`/`policy.` plus base `updatedLabel`; pricing `billing.` only; error page `errors.`, `search.label`, `nav.search` (all base). The legal "Also" link is labelled from its own namespace (`terms.otherTitle` / `policy.otherTitle`) because each page ships only its own.
- 500: "Try again" reloads the page (frame), "Back to home" is secondary. Other 4xx keep the server message as heading.
- Pre-existing copy: "1 requests per second" on the Free plan (`billing.pricing.rate` has no plural form); the 404 search field has no magnifier icon (frame has one). Left for a copy/i18n pass.
- No new binary assets.

## Review (especialista, 2026-11-01)
verdict: CORRIGIDO_E_APROVADO (one batch, `agent/especialista/IMPL-web-batch3` from the more-like-this tip, which stacks P6.D.1.3, edition-entity, recent-changes, docs-legal-errors)
re-run: `go test ./app/recommend/related ./app/catalog` ok; `go test -tags integration ./tests -run 'TestItemRelated|TestAppRelated|Inventory|TestEntity|TestEdition|TestRecentChanges|TestContent|TestLegal|TestDocs|TestPricing|TestI18n|OpenAPI|TestError'` ok; `scripts/openapi-check.sh` ok and `codegen:client -prefix /app/v1` equals packages/app-client/src/api.ts; biome on the 17 touched TS files clean; `bun run typecheck` ok; vitest on the 9 touched test files 16 pass; golangci-lint `--timeout 15m -j 2` on touched Go packages 0 issues
checked: sdd 08 §6/§7.4 (similar <= 8 with reason, crossMedia <= 6 both directions, public items only, 404 on unknown/hidden, private cache), authz, sql bound, i18n parity (570 keys per locale), docs `dangerouslySetInnerHTML` unchanged (build-time escaped), scope, tests
fixes: 1. routes/web.go conflict with PERF-i18n-page-namespaces: kept main's base list (`terms.`/`policy.` ship only on /terms and /policy through PageNamespaces; dot-less `updatedLabel` stays) (39031103) 2. mobile catalogues recopied for the new `item.*` keys (f70d09a7)
not done: browser QA stack (host memory); in-process reco counters have no exporter (shortcut noted by the author)
gate on main after merge: go build ok; targeted only (section 8.2)
PRONTO_PARA_MERGE agent/especialista/IMPL-web-batch3
