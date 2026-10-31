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
| VF-NEW-AUDIT-6 | 2026-10-31 | medium | open |
| VF-NEW-P2.H.3.1 | 2026-10-31 | low | open |
| VF-NEW-P5.A.3.1 | 2026-10-31 | low | open |

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

- **Date**: 2026-10-31. **Severity**: medium. **Status**: open.
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
