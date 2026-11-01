# FIX-library-item-score

Branch: agent/senior/FIX-library-item-score @ cb2ea44 (worktree wt/senior-b-libscore), stacked on FIX-search-hit-meta (c18d716) for the shared catalog.StarsScore.

## What
- `score` (0-10, null under 5 ratings, same expression as listings) on library.Item, ProfileRating and ListItemRef, selected in the existing queries (no extra round trips).
- Not done: rated-count for the "14 titles / 9 rated" header (row says "if cheap"; Counts is per status, a rated count needs a new query).

## Tests
- tests/library_score_test.go (library API, LoadProfile, Lists.Entries). Red: compile failure, Score field absent. Green: TestLibraryCommunityScore, TestLibraryGroups, TestProfile*, TestList* ok (TEST_DATABASE_URL on hexascore-postgres-1:5433).
- golangci-lint on app/library and tests: nothing in touched files (2 pre-existing elsewhere).

## Review (especialista, 2026-11-01)
verdict: CORRIGIDO_E_APROVADO (one batch, `agent/especialista/FIX-library-item-score-fix`)
re-run: `go test -tags integration ./tests -run 'TestSearch|TestLibrary|TestProfile|TestList|OpenAPI'` ok; `go test ./app/library ./app/catalog` ok; golangci-lint `--timeout 15m -j 2 --build-tags integration` on app/catalog, app/library, app/controllers, tests 0 issues
checked: authz (public items only, unchanged joins), sql (ids bound as uuid[], score via shared `catalog.StarsScore`), no N+1 (one extra query per page for hits, none for library), scope, tests red/green
fixes: 1. `SearchHit` is in the `/app/v1` OpenAPI document: golden `openapi/app-v1.json` and `packages/app-client/src/api.ts` regenerated (84ebe78c)
gate on main after merge: go build ok; targeted only (section 8.2); openapi job re-run at batch end
PRONTO_PARA_MERGE agent/especialista/FIX-library-item-score-fix
