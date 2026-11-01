# FIX-lint-recommend-tests golangci findings in recommend tests
status: concluida
branch: agent/senior/FIX-lint-recommend-tests @ dada12f (stacked on DOC-sdd05-trending)
files: tests/recommend_merge_test.go, tests/recommend_views_test.go
acceptance:
- `golangci-lint run --build-tags integration ./tests/...` -> no finding in recommend_*_test.go (was gofumpt, rowserrcheck, golines)
validation: lint ok on both files; no behaviour change (added rows.Err check, formatting)
changelog: none
roadmap: no box
risks / follow-ups: none

## Review (especialista, 2026-10-31)
verdict: APROVADO
re-run: golangci (integration tag) ./tests/... → 0 issues; go test -tags integration ./tests -run Recommend → ok; no behaviour change besides the rows.Err check
checked: scope, docs vs code
gate on main after merge f11f1b6: go build ok
PRONTO_PARA_MERGE agent/senior/FIX-lint-recommend-tests
