# Runbook: API deprecation

Owner documents: SDD 05 §6.1 (headers), §6.5 (versioning). Header shape: `https://vukem.dev/docs/rest-apis/overview` "Versioning" (`Deprecation: @<unix>`, `Sunset`, `Link`).

No `/api/v1` route is deprecated. Do not mount `api.Deprecate` on a live v1 operation until this runbook's checks are done. Breaking removals ship as `/api/v2` beside v1. Additive changes are a minor version. Fixes are a patch.

## Window

`app/api/v1.Deprecate` refuses a `Sunset` that is under 183 days after the `Deprecation` instant (`ErrSunsetTooSoon`). 183 days is the six-month floor. A test-only route covers the headers; do not add a deprecated v1 route just to exercise them.

## Before you mark a route

1. Pick `since` (the deprecation instant, UTC) and `sunset` at least 183 days later.
2. Wrap only that operation with `v1.Deprecate(since, sunset, link, handler)`. `link` is the migration doc, for example `/docs/v2`.
3. Add a changelog entry under `[Unreleased]` naming the operation, `since`, `sunset`, and the replacement. A `docs` commit is not that entry. Cut it with the release that announces the deprecation.
4. Mail the owners of keys that called the operation in the last 90 days. This step is manual. There is no scheduled mail job. Record the send date and the count of recipients next to the changelog entry. Do not include secrets or raw keys in the mail or in git.

## After sunset

Remove the v1 operation only once `Sunset` has passed and the changelog plus the owner mail are both recorded. Until then the route keeps answering, with the three headers on every response.
