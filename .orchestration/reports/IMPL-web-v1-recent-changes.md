# IMPL-web-v1-recent-changes Recent changes page in design v1

status: concluida
branch: agent/senior/IMPL-web-v1-recent-changes @ 9ec0044 (stacked on agent/senior/IMPL-web-v1-edition-entity @ 4592fb6)
files: resources/js/pages/Catalog/RecentChanges{,.test}.tsx, resources/lang/{en,pt-BR,es}.json (additive `community.recent.lede`, `overview`, `columns.{date,field,before,after,by}`)
acceptance:
- `bunx vitest run resources/js/pages/Catalog/RecentChanges.test.tsx` -> 2/2 (actors, values, paging, overview tab link; column headers, empty state without a table, no field label on merge rows)
- `bun run lint` -> 186 files, no fixes; `bun run typecheck` -> exit 0; `bun run test` -> 35 files / 135 tests passed
- Browser (QA :3140, four `catalog_audit` rows inserted by hand into `hexascore_qa_implweb` for item `tempor-velit-quis`): `.orchestration/reports/shots/impl-web/{desktop,mobile}-{light,dark}-_games_tempor_velit_quis_changes.png`. Compared with `web-{desktop,mobile}{,-dark}__recent-changes__default.png`: breadcrumb, serif h1, lede, Overview / Recent changes tabs with the 2 px active underline, table with small-caps column heads over a 2 px rule, date (Intl medium), field label from the community field catalogue, muted before, after, actor link in primary; on mobile each row stacks (field + date, before, "-> after", actor); Previous / Next as secondary buttons.
validation: lint ok, typecheck ok, vitest ok, browser ok (light/dark x 1440/390)
changelog: Changed: the recent changes page of items, editions and entities follows design v1 (table of edits with before and after).
roadmap: no box
risks / follow-ups:
- Not built (no data): the frame's cover thumbnail and "Game · 2029" eyebrow next to the h1 (the page props carry only `target.{type,id,title}`) and the Status column (Applied / Reverted; the audit entries carry no state). A revert shows as its own row. Open a data row if the owner wants them.
- The title keeps the existing copy "Recent changes: {title}" (frame: "{title} — recent changes"); copy change for the text-editor agent if wanted.
- Pagination stays Previous / Next links (frame: "Load older changes"); the server pages by `?page=`.
- No new binary assets.

## Review (especialista, 2026-11-01)
verdict: CORRIGIDO_E_APROVADO (one batch, `agent/especialista/IMPL-web-batch3` from the more-like-this tip, which stacks P6.D.1.3, edition-entity, recent-changes, docs-legal-errors)
re-run: `go test ./app/recommend/related ./app/catalog` ok; `go test -tags integration ./tests -run 'TestItemRelated|TestAppRelated|Inventory|TestEntity|TestEdition|TestRecentChanges|TestContent|TestLegal|TestDocs|TestPricing|TestI18n|OpenAPI|TestError'` ok; `scripts/openapi-check.sh` ok and `codegen:client -prefix /app/v1` equals packages/app-client/src/api.ts; biome on the 17 touched TS files clean; `bun run typecheck` ok; vitest on the 9 touched test files 16 pass; golangci-lint `--timeout 15m -j 2` on touched Go packages 0 issues
checked: sdd 08 §6/§7.4 (similar <= 8 with reason, crossMedia <= 6 both directions, public items only, 404 on unknown/hidden, private cache), authz, sql bound, i18n parity (570 keys per locale), docs `dangerouslySetInnerHTML` unchanged (build-time escaped), scope, tests
fixes: 1. routes/web.go conflict with PERF-i18n-page-namespaces: kept main's base list (`terms.`/`policy.` ship only on /terms and /policy through PageNamespaces; dot-less `updatedLabel` stays) (39031103) 2. mobile catalogues recopied for the new `item.*` keys (f70d09a7)
not done: browser QA stack (host memory); in-process reco counters have no exporter (shortcut noted by the author)
gate on main after merge: go build ok; targeted only (section 8.2)
PRONTO_PARA_MERGE agent/especialista/IMPL-web-batch3
