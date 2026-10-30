# vukem feedback from HexaScore

Open queue only. VF-001 through VF-061 were fixed, withdrawn, or wontfix by vukem v0.34.0 and were removed from this file on 2026-10-30. Git history keeps those entries. Do not reuse those ids.

One entry per finding. Update `status` when it changes. Cite <https://vukem.dev/docs> (page URL + heading). The local `vukem/docs` tree is cited only when the site has no page, tagged `[SITE PAGE MISSING]`.

Severity: `critical` (blocks a roadmap task), `high` (workaround costs more than a day), `medium` (workaround under a day), `low` (DX or docs).
Status: `open`, `reported` (issue or PR link), `fixed in vX.Y.Z`, `wontfix`, `withdrawn`.

| id | date | severity | status |
| --- | --- | --- | --- |
| VF-062 | 2026-10-30 | low | open |

## VF-062: A typed POST that returns a body defaults to 201, including commands that do not create a resource

- **Date**: 2026-10-30. **Severity**: low. **Status**: open.
- **vukem version**: v0.34.0 (`f3080367`). **Docs consulted**: <https://vukem.dev/docs/rest-apis/typed-handlers> "The output". The status table already says: `Op.Status` if set, `204` when `Out` is `struct{}`, `201 Created` when the method is `POST`, otherwise `200`. `Op.Status` godoc in `api/typed.go` says the same. The OpenAPI builder in `openapi/openapi.go` uses the same switch, so the exported document follows the handler.
- **Context**: `POST /api/v1/webhooks/{id}/test` returns `{delivered, status}` and does not create a resource. The handler body was already correct. The HTTP status and the generated operation were `201` until the route set `Op.Status` to `200`. Acceptance failed with `status = 201, want 200` while the JSON said the receiver had returned 200.
- **Repro**: register `POST /widgets/{id}/ping` with `api.Op{Name: "widgets.ping"}` and a non-empty output struct, and leave `Op.Status` at 0. The response is `201` and `openapi:export` lists `"201"` as the success response. Set `Op.Status` to `200` and both become `200`.
- **Workaround**: set `Op.Status` on every POST that is a command, not a create. HexaScore does this for `testWebhook` in `routes/api_v1.go`.
- **Request**: do not change the default. Existing create routes rely on `201`. On <https://vukem.dev/docs/rest-apis/typed-handlers> "The output", add a command-POST example (`POST /webhooks/{id}/test`) that sets `Op.Status: http.StatusOK`, and state that `openapi` uses the same default, so a forgotten status is written into the golden document and not only into one response.
