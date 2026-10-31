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
