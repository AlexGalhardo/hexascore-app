# vukem feedback from HexaScore

Open queue only. VF-001 through VF-061 were fixed, withdrawn, or wontfix by vukem v0.34.0 and were removed from this file on 2026-10-30. Git history keeps those entries. Do not reuse those ids.

One entry per finding. Update `status` when it changes. Cite <https://vukem.dev/docs> (page URL + heading). The local `vukem/docs` tree is cited only when the site has no page, tagged `[SITE PAGE MISSING]`.

Severity: `critical` (blocks a roadmap task), `high` (workaround costs more than a day), `medium` (workaround under a day), `low` (DX or docs).
Status: `open`, `reported` (issue or PR link), `fixed in vX.Y.Z`, `wontfix`, `withdrawn`.

| id | date | severity | status |
| --- | --- | --- | --- |
| VF-062 | 2026-10-30 | low | open |
| VF-NEW-AUDIT-1 | 2026-10-31 | medium | open |
| VF-NEW-AUDIT-2 | 2026-10-31 | low | open |
| VF-NEW-AUDIT-3 | 2026-10-31 | medium | open |
| VF-NEW-AUDIT-4 | 2026-10-31 | low | open |
| VF-NEW-AUDIT-5 | 2026-10-31 | medium | open |
| VF-NEW-AUDIT-6 | 2026-10-31 | medium | fixed on main, unreleased (PR #105) |
| VF-NEW-P2.H.3.1 | 2026-10-31 | low | open |
| VF-NEW-P5.A.3.1 | 2026-10-31 | low | open |
| VF-NEW-FIX-gate-timeouts | 2026-10-31 | medium | fixed on main, unreleased (PR #106) |
| VF-NEW-P4.A.3.1b | 2026-10-31 | low | open |
| VF-NEW-P2.H.3.5 | 2026-10-31 | medium | open |
| VF-NEW-FIX-covers-route | 2026-10-31 | low | open |
| VF-NEW-OWNER-schemathesis | 2026-10-31 | medium | open |

## VF-062: A typed POST that returns a body defaults to 201, including commands that do not create a resource

- **Date**: 2026-10-30. **Severity**: low. **Status**: open.
- **vukem version**: v0.34.0 (`f3080367`). **Docs consulted**: <https://vukem.dev/docs/rest-apis/typed-handlers> "The output". The status table already says: `Op.Status` if set, `204` when `Out` is `struct{}`, `201 Created` when the method is `POST`, otherwise `200`. `Op.Status` godoc in `api/typed.go` says the same. The OpenAPI builder in `openapi/openapi.go` uses the same switch, so the exported document follows the handler.
- **Context**: `POST /api/v1/webhooks/{id}/test` returns `{delivered, status}` and does not create a resource. The handler body was already correct. The HTTP status and the generated operation were `201` until the route set `Op.Status` to `200`. Acceptance failed with `status = 201, want 200` while the JSON said the receiver had returned 200.
- **Repro**: register `POST /widgets/{id}/ping` with `api.Op{Name: "widgets.ping"}` and a non-empty output struct, and leave `Op.Status` at 0. The response is `201` and `openapi:export` lists `"201"` as the success response. Set `Op.Status` to `200` and both become `200`.
- **Workaround**: set `Op.Status` on every POST that is a command, not a create. HexaScore does this for `testWebhook` in `routes/api_v1.go`.
- **Request**: do not change the default. Existing create routes rely on `201`. On <https://vukem.dev/docs/rest-apis/typed-handlers> "The output", add a command-POST example (`POST /webhooks/{id}/test`) that sets `Op.Status: http.StatusOK`, and state that `openapi` uses the same default, so a forgotten status is written into the golden document and not only into one response.

## VF-NEW-AUDIT-1: A 404 raised by a service renders the bare framework page in an Inertia app

- **Date**: 2026-10-31. **Severity**: medium. **Status**: open.
- **vukem version**: v0.34.0. **Docs consulted**: <https://vukem.dev/docs/frontend/inertia> "Rendering pages" (`inertia.RenderStatus`) and <https://vukem.dev/docs/basics/errors>. The error hook that renders the Inertia `Error` page covers Inertia visits; a plain `GET` that ends in `errs.New(404, ...)` from a service answers the framework HTML page without the application shell.
- **Context**: `/u/{nickname}` and `/u/{nickname}/lists/{id}` with an unknown nickname showed "404 Not Found" with no header, footer or locale, while the catch-all route rendered the shell. Found by the architecture audit of 2026-10-31.
- **Workaround**: `app/controllers/not_found.go` `pageError` maps `errs.Status(err) == 404` to `inertia.RenderStatus(c, 404, "Error", ...)` in every page controller that calls a service.
- **Request**: an `inertia.Config` option (for example `ErrorPage: "Error"`) so that any error reaching the error handler on an HTML request renders that component with `status`, `title` and `message` props, Inertia visit or not.

## VF-NEW-AUDIT-2: The default app name "Vukem" leaks into the brand, title and admin of a scaffolded app

- **Date**: 2026-10-31. **Severity**: low. **Status**: open.
- **vukem version**: v0.34.0. **Docs consulted**: <https://vukem.dev/docs/configuration> "APP_NAME" [NOT VERIFIED: page name assumed]. `app.Config.Name` is "Vukem" when `APP_NAME` is unset.
- **Context**: the QA compose stack, CI and the e2e suite ran without `APP_NAME`; every page title, the header, the footer and the admin ("Vukem admin") said Vukem for weeks and no test noticed.
- **Workaround**: `routes/web.go` shares `cmp.Or(app.Source.Get("APP_NAME"), "HexaScore")`; `docker-compose.yml` sets `APP_NAME`.
- **Request**: `vukem new <name>` writes the project name as the default of `Config.Name` (generated constant or `.env.example` plus a fallback to the module name), and the admin title reads `Config.Name`.

## VF-NEW-AUDIT-3: DisableSSR pages ship an empty root element

- **Date**: 2026-10-31. **Severity**: medium. **Status**: open.
- **vukem version**: v0.34.0. **Docs consulted**: <https://vukem.dev/docs/frontend/inertia> "Server-side rendering" (`DisableSSR`).
- **Context**: `/profile`, `/library`, `/lists`, `/me/notifications` are CSR by design (private data). The HTML has an empty `#app` and no `<title>` until hydration, so the first paint is a blank page (screenshots in the 2026-10-31 audit).
- **Workaround**: none yet; ROADMAP P2.H.2.1 adds an app-side skeleton.
- **Request**: let `DisableSSR` render the shell: a configurable placeholder template (or the `Error`-style component through SSR with a `loading` prop) plus the page `title` from props, so a CSR page is never blank.

## VF-NEW-AUDIT-4: The inert-until-hydrated loader gives no styling hook for the gap

- **Date**: 2026-10-31. **Severity**: low. **Status**: open.
- **vukem version**: v0.34.0. **Docs consulted**: <https://vukem.dev/docs/frontend/inertia> "The HTML loader" [NOT VERIFIED: heading assumed].
- **Context**: HexaScore sets `<html inert>` until `createInertiaApp` settles (its own fix for double submits). A click in that window does nothing and nothing tells the user; the audit reproduced a silent login click ~2 s after load.
- **Workaround**: none; P2.H.2.2 adds a pending state in the app.
- **Request**: the page loader toggles a `data-hydrating` attribute on `<html>` (set before the script, removed when the app mounts) so apps can style disabled buttons and show a pending hint with CSS only; document the pattern with the inert approach.

## VF-NEW-AUDIT-5: Custom admin pages do not inherit the admin navigation

- **Date**: 2026-10-31. **Severity**: medium. **Status**: open.
- **vukem version**: v0.34.0. **Docs consulted**: <https://vukem.dev/docs/digging-deeper/admin>.
- **Context**: `/admin/ingestion` is a hand-made page inside the admin layout; its navigation lists 7 entries while the resource pages list 11, and the title reads "HexaScore admin" against "Vukem admin" elsewhere.
- **Workaround**: none; P2.H.3.3 duplicates the nav in the app.
- **Request**: an `admin.Page` (or `Panel.AddPage`) registration that renders a custom handler inside the admin layout with the same navigation, title and guard as the resources.

## VF-NEW-AUDIT-6: Rate limiter has no burst allowance

- **Date**: 2026-10-31. **Severity**: medium. **Status**: fixed on main, unreleased (PR #105).
- **vukem version**: v0.34.0. **Docs consulted**: <https://vukem.dev/docs/security/rate-limiting> "Keys" [NOT VERIFIED whether a burst option exists; checked only the app usage in `app/http/limits`].
- **Context**: the public API free plan is 1 request/second; the first request leaves `ratelimit-remaining: 0` and the second within the same second is 429. Every sequential client script fails on its second call (audit 2026-10-31).
- **Workaround**: none; P2.H.3.1 proposes a burst in the app layer if the framework has none.
- **Request**: a token-bucket mode (`Limit{Rate: 1, Burst: 5}`) in `ratelimit`, with `RateLimit-Policy` headers reflecting it, or a documented way to compose two limits (per second and per minute) on one route.

## VF-NEW-P2.H.3.1: No way to document an enum on a typed-handler path parameter without enforcing it as a 422

- **Date**: 2026-10-31. **Severity**: low. **Status**: open.
- **vukem version**: v0.34.0. **Docs consulted**: <https://vukem.dev/docs/rest-apis/typed-handlers> "The input". The only tag that puts `enum` in the OpenAPI schema is `validate:"oneof=..."`, which also rejects the request with a 422 before the handler runs [NOT IN VUKEM DOCS: a doc-only enum].
- **Context**: `GET /api/v1/{type}` and `/{type}/{id}` must document `games|movies|series|books` for SDK users, but the API contract answers an unknown type with 400 `invalid-query` that the quota refunds (SDD 05 FR-05-4).
- **Workaround**: `validate:"oneof=..."` on `ListItemsInput.Type` and `GetItemInput.Type`, plus `publicapi.TypeGuard` (last middleware of the `/api/v1` group) that answers the 400 first. `// workaround(vukem): VF-NEW-P2.H.3.1`.
- **Request**: a documentation-only tag (for example `enum:"a,b"`) or an `api.Op` option to map validation errors to a chosen status.

## VF-NEW-P5.A.3.1: Generated ApiProblem type has no `code` field although problem responses carry one

- **Date**: 2026-10-31. **Severity**: low. **Status**: open.
- **vukem version**: v0.34.0 (per the P2.H.3.1 entry). **Docs consulted**: <https://vukem.dev/docs/rest-apis/overview> "Responses and errors": the page documents RFC 9457 problem details (`type`, `title`, `status`, `detail`, `instance`, `errors`) and no `code` member; the page has no section on the generated client, `codegen:client` or an `ApiProblem` type [NOT IN VUKEM DOCS: client codegen output shape].
- **Context**: `packages/app-client/src/api.ts` (generated, P5.A.3.1 branch) declares `ApiProblem` with `type`, `title`, `status`, `detail`, `instance`, `request_id`, `errors`, and no `code`. The HexaScore SDD 05 problem codes travel in that field, so `apps/mobile/src/api/problems.ts` reads `code` through an `unknown` cast.
- **Workaround**: `(error.problem as { code?: unknown } | undefined)?.code` in `apps/mobile/src/api/problems.ts`; the mobile tests cover unknown codes. `// workaround(vukem): VF-NEW-P5.A.3.1` to be added with the mobile task.
- **Request**: codegen types a `code` member on problem responses, or the framework documents the problem extension members it writes.

## VF-NEW-FIX-gate-timeouts: pgtest drop timeout is fixed at 30 s and fails parallel suites on slow disks

- **Date**: 2026-10-31. **Severity**: medium. **Status**: fixed on main, unreleased (PR #106).
- **vukem version**: v0.34.0. **Docs consulted**: <https://vukem.dev/docs/testing/database> [NOT VERIFIED whether the page documents the cleanup timeout; checked the behaviour from the error only].
- **Context**: `pgtest` drops each per-test database with `DROP DATABASE ... WITH (FORCE)` under a 30 s context. The drop forces a checkpoint; with about 130 parallel integration tests on Docker Desktop for Windows the fsync-bound checkpoint exceeds 30 s and the run fails with `pgtest: drop database ...: context deadline exceeded` although every assertion passed (`go test -tags integration -parallel 4 ./tests/`: FAIL after 717 s, three times).
- **Workaround**: dev/test Postgres in `docker-compose.yml` runs with `fsync=off`, `full_page_writes=off`, `synchronous_commit=off` (FIX-gate-timeouts, 717 s FAIL to 245 s ok). Production is untouched.
- **Request**: make the drop timeout configurable (`pgtest.Options{DropTimeout}` or an env var), or drop asynchronously after the test reports, so cleanup cannot fail a green test.

## VF-NEW-P4.A.3.1b: Flag name rules are undocumented and registering a flag silently mounts `/admin/_flags`

- **Date**: 2026-10-31. **Severity**: low. **Status**: open.
- **vukem version**: v0.34.0. **Docs consulted**: <https://vukem.dev/docs/digging-deeper/feature-flags> [NOT IN VUKEM DOCS: the allowed name characters and the admin panel side effect].
- **Context**: the SDD named flags `community.submissions` and `community.trusted_moderators`; vukem's flag name check (`nameRe`) accepts only `[a-z0-9_]`, found only by reading the source. Registering the first flag with `WithFlags` also mounts the `/admin/_flags` panel, which surfaced as new admin routes in the route-inventory security test.
- **Workaround**: flags renamed to `community_submissions` / `community_trusted_moderators` (env `FLAG_COMMUNITY_SUBMISSIONS`), SDD 04 §6 amended, route inventory updated.
- **Request**: document the name pattern and the env mapping on the feature-flags page, and either document the automatic `/admin/_flags` mount or make it opt-in.

## VF-NEW-P2.H.3.5: No page output cache that keeps the CSP nonce per request

- **Date**: 2026-10-31. **Severity**: medium. **Status**: open.
- **vukem version**: v0.34.0. **Docs consulted**: <https://vukem.dev/docs/frontend/ssr> [NOT IN VUKEM DOCS: an SSR output cache or a cache hook].
- **Context**: every anonymous first load runs SSR, so the page TTFB budget (k6 p75 ≤ 400 ms, SDD 09 §4) failed under load (470 to 1850 ms). Caching the rendered document in the app is unsafe by default because the body carries the per-request CSP nonce; the first draft replayed the cached CSP header and broke `TestSecurity_A05_NonceIsPerRequestAndInPage`.
- **Workaround**: app middleware `RenderCache` (P2.H.3.5): anonymous GETs cached in process for 60 s, key = path + query + theme/consent/locale cookies, the cached body's nonce swapped for the current request's nonce, CSP never replayed; p75 157 to 296 ms. `// workaround(vukem): VF-NEW-P2.H.3.5`.
- **Request**: an opt-in SSR output cache (TTL, vary keys, bypass on session cookie) that injects the request nonce at serve time, or a documented hook to cache the SSR HTML before nonce injection.

## VF-NEW-FIX-covers-route: S3 disk exposes no ETag, so `storage.HandlerWith` cannot answer 304

- **Date**: 2026-10-31. **Severity**: low. **Status**: open.
- **vukem version**: v0.34.0. **Docs consulted**: <https://vukem.dev/docs/digging-deeper/file-storage> [NOT VERIFIED whether the page documents conditional GET on the S3 driver].
- **Context**: the app serves covers from a private bucket through `storage.HandlerWith` with `Cache-Control: public, max-age=31536000, immutable` (FIX-covers-route). Against RustFS the response carries no `ETag` and `If-None-Match` never yields 304; the memory disk path was the only one tested by the author.
- **Workaround**: none needed while keys are content-hashed and immutable; browsers rely on the long max-age.
- **Request**: have the S3 driver surface the object ETag (and `Last-Modified`) through the disk stat so `HandlerWith` answers conditional requests with 304.

## VF-NEW-OWNER-schemathesis: Native Schemathesis support for property-based API testing

- **Date**: 2026-10-31. **Severity**: medium (feature request from the owner). **Status**: open.
- **vukem version**: v0.34.0. **Docs consulted**: <https://vukem.dev/docs/rest-apis/openapi> [NOT IN VUKEM DOCS: any contract or fuzz testing of the exported document]. Tool: <https://schemathesis.io/> (Schemathesis 4.29.0, run as a pinned Docker image).
- **Context**: HexaScore wires Schemathesis by hand in `qa/schemathesis/` (about 200 lines): `ci.sh` migrates a throwaway Postgres, inserts a business API key and an app token with raw SQL, starts the image and runs `run.sh` once per document (`/api/v1`, `/app/v1`); `run.sh` rewrites `127.0.0.1` to `host.docker.internal` for Docker; `schemathesis.toml` disables checks that a read-only API cannot satisfy and explains each; `ip_hook.py` rotates `X-Forwarded-For` so the per-IP rate limit does not turn the missing-auth check into 429s. Every vukem app with a typed REST API would repeat this.
- **Workaround**: the scripts above, run in CI (`OpenAPI` job) and locally.
- **Request**: a first-class integration, without a new Go dependency (Schemathesis stays an external pinned image):
  1. `vukem api:fuzz -prefix /api/v1` (name to be decided) that exports the document in-process, starts or targets the app, and runs Schemathesis with a pinned image tag/digest, the Docker host rewrite and a JUnit/JSON report directory.
  2. A test-credentials hook: a documented way to mint a scoped API key or token for the scan (for example a `fuzz` seeder callback or `--auth-header` from a command), instead of raw SQL.
  3. Sensible default config generated from the API metadata: read-only APIs skip stateful checks, `TRACE` handling documented, rate limits bypassed for the scan (trusted header or a test-only limiter key), max-examples and workers configurable.
  4. A `Taskfile`/CI snippet and a docs page under rest-apis, plus an example app wiring.
  5. HexaScore then replaces `qa/schemathesis/` with the native command (an `ADOPT-vukem-<version>` row).
