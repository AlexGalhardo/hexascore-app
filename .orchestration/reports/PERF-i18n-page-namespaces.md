# PERF-i18n-page-namespaces report (senior-b)
STATUS: review
Branch: agent/senior/PERF-i18n-page-namespaces (184f71f). New `appmiddleware.PageI18n` (app/http/middleware/i18n_pages.go) replaces `inertia.I18n` in routes/web.go: same locale/locales/messages props, prefix list = base + `PageNamespaces` entry for the path (locale prefix ignored). Namespaces moved off every page: docs (/docs), terms (/terms), policy (/policy), help (/help), billing (/billing, /pricing), api (/profile/api), contact (/contact), dashboard (/dashboard). Only these were verified (grep of resources/js) to be used solely by their own pages. About 32 KB of 61.7 KB raw en catalogue (52 percent) leaves the item/home/search pages; not measured with Lighthouse (8.2).
Once-prop key is now `i18n:<locale>` for the base set and `i18n:<locale>:<group>` for extended sets, so a client holding the base catalogue still receives terms.* when navigating to /terms. No resources/js change needed (impl-web: when a page starts using one of those namespaces outside its prefix, add the prefix to PageNamespaces; larger namespaces auth/account/community/item stay shared because Nav/AppLayout/Catalog use them).
Red: TestI18nPageNamespaces and TestI18nOnceKeyPerNamespaceSet failed (messages carried terms./docs. everywhere; same once key). Green: both plus auth/content/csr/docs/pricing/contact tests ok; middleware unit tests ok. Lint: my files clean; 2 pre-existing gci/golines issues in tests/api_edge_test.go.

## Review (especialista, 2026-11-01)
verdict: CORRIGIDO_E_APROVADO (batch `agent/especialista/senior-b-batch`)
re-run: `go test ./app/alerts ./app/ssr ./app/api ./app/http/middleware` ok; `go test -tags integration ./tests -run 'TestAlert_A5|TestReady|TestI18n'` ok; `go test ./internal/search -run 'TestReindex|TestSearchSQLConfined|TestSuggestRebuild'` ok; golangci-lint `--timeout 15m -j 2` on touched packages 0 issues; markdownlint docs/runbooks + SDD 12 0 issues
checked: sdd 12 (alerts A5/A6/A10, runbooks), authz, secrets (error text not logged), sql (confined to internal/search), i18n, scope, tests
fixes: 1. disk-full runbook points at `/app/app search:reindex` (2aaa66d) 2. dot-less `updatedLabel` key in routes/web.go page namespaces (6cc9e68)
gate on main after merge: targeted only (section 8.2); batch gate pending
PRONTO_PARA_MERGE agent/especialista/senior-b-batch (merged 7044df3)
