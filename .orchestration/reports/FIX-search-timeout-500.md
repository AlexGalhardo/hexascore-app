# FIX-search-timeout-500 search backend timeout no longer answers 500
status: concluida
branch: agent/senior/FIX-search-timeout-500 @ c12becc
files: app/controllers/search_controller.go, app/controllers/publicapi/search.go, app/controllers/publicapi/search_err_test.go, tests/search_timeout_test.go, resources/js/pages/Search/Index.tsx, resources/lang/{en,es,pt-BR}.json (search.unavailable)
acceptance:
- TEST_DATABASE_URL=<local ci pg> go test -tags integration ./tests -run 'TestSearchTimeoutDegrades|TestInertiaPages_Search|TestAppSearch|TestSearchSuggest' -> ok (the timeout test holds ACCESS EXCLUSIVE on item_translations so the 800 ms statement timeout fires for real)
- go test ./app/controllers/publicapi -run SearchErr -> ok
validation: go build ok, biome ci resources/js/pages/Search ok, lang JSON parses; golangci-lint and tsc not run (per 8.2)
design: HTML (/search): ErrTimeout becomes an empty page, prop `unavailable: true`, notice `search.unavailable`, `Retry-After: 5`, `Cache-Control: private, no-store`, not cached (Remember does not store errors), no search_performed event. JSON (/app/v1/search, /search/suggest via badQuery): 503 problem with `Retry-After: 5`. Public API searchErr already returned 503; it now also sends Retry-After.
changelog: Fixed: a search backend timeout degrades to an empty results page with a notice (HTML) or a 503 problem with Retry-After (JSON) instead of HTTP 500.
roadmap: no box
risks / follow-ups: tsc typecheck of the optional `unavailable` prop not run; Suggest keeps its existing empty-list behaviour inside the searcher (SDD 07 section 8).

## Review (especialista, 2026-10-31)
verdict: APROVADO
re-run: `go test -tags integration ./tests -run 'TestSearchTimeoutDegrades|TestInertiaPages_Search|TestAppSearch|TestSearchSuggest'` ok (13 s); `go test ./app/controllers/publicapi -run SearchErr` ok; `go test ./tests/i18n/...` ok (after the P4.A.3.1d merge, same lang files)
checked: SDD 07 §6/§8 (400 for invalid query unchanged, 503 + Retry-After for timeouts on JSON and public API), degraded HTML is `private, no-store` and records no search_performed event, Remember does not store the error, i18n parity (3 locales), notice uses role=status, scope limited to the row's files
gate on main after merge: see the batch gate line in reports/P4.A.3.1d.md
PRONTO_PARA_MERGE agent/senior/FIX-search-timeout-500 (merged fe0f28d)
