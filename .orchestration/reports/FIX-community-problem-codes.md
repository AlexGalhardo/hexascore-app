# FIX-community-problem-codes community problem codes in the catalogue
status: concluida
branch: agent/senior/FIX-community-problem-codes @ a812b16
files: app/http/problems/catalog.go, catalog_test.go, resources/docs/errors.md, resources/docs/docs.json
acceptance:
- `go test -tags integration ./tests -run TestProblemCatalog` -> ok (every code redirects and has an errors.md section, incl. 9 new)
- `go test ./app/http/problems/...` -> ok (ForStatus never returns a Specific code)
validation: lint ok, typecheck n/a (no TS), tests problems+TestProblemCatalog ok, markdownlint ok; no problems.* i18n keys exist, none added
changelog: Fixed: community problem codes (suggestion-*, field-locked, stale-before, invalid-transition, community-submissions-closed, validation) now resolve at /problems/{code} and are documented.
roadmap: no box
risks / follow-ups: Problem.Specific keeps new 403/409/423/429/422 codes out of ForStatus defaults. Worktree has no node_modules, so the docs.json stale vitest was not run (build:docs regenerated it).

## Review (especialista, 2026-10-31)
verdict: APROVADO
re-run: `go test ./app/http/problems/...` ok; `go test -tags integration ./tests -run TestProblemCatalog` PASS; `golangci-lint run ./app/http/problems/...` 0 issues; after merge on main: `bunx vitest run` 24 files, 84 tests ok (docs.json not stale)
checked: every new code is raised in app/community or controllers/community with the catalogue status (403 closed/not-eligible, 409 duplicate/stale/transition, 423 locked, 429 quota/cooldown, 422 validation); `Specific` keeps them out of `ForStatus` (403 still plan-required, 429 still rate-limited); errors.md sections; no i18n keys involved; scope
gate on main after merge: targeted only (owner rule 8.2; full gate once per batch)
PRONTO_PARA_MERGE agent/senior/FIX-community-problem-codes (merged afc0d43)
