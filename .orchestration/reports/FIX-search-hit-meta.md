# FIX-search-hit-meta

Branch: agent/senior/FIX-search-hit-meta @ c18d716 (worktree wt/senior-b-hit-meta)

## What
- `catalog.HitMetaFor` (app/catalog/hit_meta.go): one query per page, release year and gated community score (0-10, null under 5 ratings, same `starsScore` expression as listings).
- `SearchHit` gains `year` and `score` (both always present, null when unknown); filled in `attachCovers` for typed results and groups. No search SQL outside `internal/search`; ETag, cache and timeout paths untouched.
- Not done: `byline` (row marks it optional; needs a creator lookup).

## Tests
- `tests/search_hit_meta_test.go` (typed and grouped).
- Red (controller change reverted): year=nil score=nil. Green: `go test -tags integration ./tests -run TestSearch` ok (TEST_DATABASE_URL on hexascore-postgres-1:5433).
- golangci-lint on touched packages: no issues in touched files (2 pre-existing elsewhere).

## Review (especialista, 2026-11-01)
verdict: CORRIGIDO_E_APROVADO (one batch, `agent/especialista/FIX-library-item-score-fix`)
re-run: `go test -tags integration ./tests -run 'TestSearch|TestLibrary|TestProfile|TestList|OpenAPI'` ok; `go test ./app/library ./app/catalog` ok; golangci-lint `--timeout 15m -j 2 --build-tags integration` on app/catalog, app/library, app/controllers, tests 0 issues
checked: authz (public items only, unchanged joins), sql (ids bound as uuid[], score via shared `catalog.StarsScore`), no N+1 (one extra query per page for hits, none for library), scope, tests red/green
fixes: 1. `SearchHit` is in the `/app/v1` OpenAPI document: golden `openapi/app-v1.json` and `packages/app-client/src/api.ts` regenerated (84ebe78c)
gate on main after merge: go build ok; targeted only (section 8.2); openapi job re-run at batch end
PRONTO_PARA_MERGE agent/especialista/FIX-library-item-score-fix
