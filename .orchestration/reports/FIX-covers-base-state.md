# FIX-covers-base-state Remove package-level coversBase
status: concluida
branch: agent/senior/FIX-covers-base-state @ fe37067
files: app/catalog/covers.go, app/catalog/covers_test.go, app/http/middleware/share_props.go, routes/web.go, resources/js/components/CoverThumb.tsx, resources/js/types.ts, shell.test.tsx, ResultGroup.test.tsx, tests/home_test.go
design: Go cover URLs are always same-origin paths (`/covers/<key>`), no mutable state. COVERS_BASE_URL reaches the client as the shared Inertia prop `coversBase`; CoverThumb prepends it to src and srcSet. CSP img-src host logic unchanged.
acceptance:
- `go test ./app/catalog/ ./app/http/...` and `go test ./tests -run 'TestHome$'` (coversBase prop present) → ok
- `bun run typecheck`, `bun run test` → 19 files, 68 tests ok; golangci-lint 0 issues
changelog: Changed: cover origin (COVERS_BASE_URL) is delivered as a shared prop; no package-level state in app/catalog.
roadmap: no box
risks / follow-ups: the JSON of /search/suggest (coverThumbUrl) now carries a same-origin path; the web SuggestBox does not read it. No browser check run (port 3000 busy); only CoverThumb renders covers, and its test covers the prefix.

## Review (especialista, 2026-10-31)
verdict: REJEITADO
re-run: not run; the verdict rests on the API contract below
checked: web path ok (SSR and client both read the shared `coversBase` prop from the page object, so the server-rendered `<img>` gets the prefix too), no package state left in `app/catalog`; consumers of `catalog.CoverURL`/`NewCover`/`CoverFromThumbKey`: Inertia pages (fine), `/search/suggest` `coverThumbUrl` (JSON), and `GET /app/v1/lists/{id}` `items[].cover` (untyped `c.JSON`, built by `library.Lists.Entries`); `/api/v1` and both OpenAPI goldens carry no cover field; the mobile app reads neither field today
rework:
1. JSON responses must not change meaning with `COVERS_BASE_URL`: with the base set, `/app/v1/lists/{id}` `items[].cover.thumb/full` and `/search/suggest` `coverThumbUrl` become app-relative paths that a non-web client (the Android app, SDD 10) resolves against the API host, where no `/covers` route exists when covers live in a bucket. Either keep absolute URLs in those JSON handlers (pass the base into `library.Lists`/the suggest handler, no package state), or wait for the owner's answer to Q-A6 (`/covers` route on the app origin) and say so in the report and the CHANGELOG line.
2. Add a test with `COVERS_BASE_URL=https://bucket.example`: the `/app/v1/lists/{id}` item cover and the suggest `coverThumbUrl` are absolute; the Inertia `cover` stays a path and `coversBase` is the trimmed base.
3. Report: list every consumer of the cover fields (the two JSON endpoints above included) and state the API impact.

## r2 (senior-b, 2026-10-31)
branch: agent/senior/FIX-covers-base-state @ fb02136
fix: JSON handlers send absolute cover URLs again; no package state. `catalog.CoversBase`, `catalog.Absolute` and `(*Cover).WithBase` (copy, never mutates) take the base; `controllers.Search` and `appapi.Lists` hold it as a field read from `COVERS_BASE_URL` at construction (`routes/lists.go` passes it, `routes/web.go` reuses `CoversBase`).
consumers of the cover fields:
- Inertia pages (home, item, search, library, lists, profile): path + shared `coversBase` prop (unchanged, r1).
- `GET /app/v1/lists/{id}` `items[].item.cover`: absolute (fixed).
- `GET /search/suggest` and `GET /app/v1/search/suggest` `coverThumbUrl`: absolute (fixed; the web SuggestBox does not read it).
- `GET /app/v1/search` `data[].cover`: absolute (also missed in the review list; fixed).
- `/api/v1` and the OpenAPI goldens carry no cover field; mobile reads none today.
API impact: none; with an empty base the JSON stays a same-origin path as before r1.
acceptance:
- `go test -tags integration ./tests -run 'TestCoversBaseJSONAbsolute|TestAppLists|TestAppSearch|TestHome$|TestSearchSuggest|TestSearchPage'` -> ok (base `https://bucket.example`: all four JSON fields absolute, Inertia `cover` stays `/covers/...`, `coversBase` trimmed)
- `go test ./app/catalog/ ./app/appapi/` -> ok; `golangci-lint run ./app/... ./routes/...` -> 0 issues
changelog: unchanged from r1 (Changed: cover origin delivered as a shared prop; JSON endpoints keep absolute cover URLs).

## Review r2 (especialista, 2026-10-31)
verdict: CORRIGIDO_E_APROVADO
re-run: `go test -tags integration ./tests -run 'TestCoversBaseJSONAbsolute|TestLibrary|TestPublicProfile|TestAppLists|TestAppSearch|TestHome$|TestSearchSuggest|TestSearchPage'` -> ok (35.9 s); the new library assertion fails without the fix (`groups.0.items.0.cover.thumb = /covers/a-1-t.webp`)
checked: every `catalog.Cover` producer (`NewCover`, `CoverFromThumbKey`, `CoverURL`) traced to its handler; Inertia pages (home highlights and recent ratings, item, search, library, lists, profile) keep paths + `coversBase`; `ListSummary.coverThumbs` are raw keys (unchanged); no package state; private buckets (SDD 11 §274) so CSP host logic unchanged; authz/sql untouched
fixes: 1. two JSON consumers were still missed: `GET /app/v1/me/library` `groups[].items[].cover` and `GET /app/v1/users/{nickname}` `ratings[].cover` sent app paths with the base set; `controllers.Library` and `controllers.PublicProfile` take the base at construction like `appapi.Lists` (745feb6), asserted in `TestCoversBaseJSONAbsolute` 2. golines wrap in the test (2750d0e) 3. merge conflict with main's `lg` CoverThumb size: `src` keeps the lg/full choice and the base prefix (8bdfdf3)
gate on main @ 8bdfdf3: go build ok, golangci-lint (integration tags) 0 issues, bun lint ok (4 pre-existing warnings), typecheck ok, vitest 65 passed, go test integration all ok
PRONTO_PARA_MERGE agent/especialista/FIX-covers-base-state-fix (merged 8bdfdf3)
