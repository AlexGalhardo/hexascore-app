# vukem feedback from HexaScore

Open queue only. Fixed, withdrawn and wontfix entries are removed; git history keeps them. Do not reuse removed ids.

Removed on 2026-10-31, all resolved on vukem `main` (23cde35a) but not in a release yet: VF-062 (#103), VF-NEW-AUDIT-1 (#107), VF-NEW-AUDIT-2 (#115), VF-NEW-AUDIT-3 (#109), VF-NEW-AUDIT-4 (#116), VF-NEW-AUDIT-5 (#110), VF-NEW-AUDIT-6 (#105), VF-NEW-P2.H.3.1 (#117), VF-NEW-P5.A.3.1 (#118), VF-NEW-FIX-gate-timeouts (#106), VF-NEW-P4.A.3.1b (#119), VF-NEW-P2.H.3.5 (withdrawn, #111), VF-NEW-FIX-covers-route (#120; adoption needs a `drivers/s3` release, Q-A14) and VF-NEW-OWNER-schemathesis (#113). HexaScore adopts them in an `ADOPT-vukem-<version>` row once the owner asks for a vukem release.

Adopted in vukem v0.35.0 (ADOPT-vukem-0.35.0), entries removed: all of the above plus VF-NEW-P4.B.2.5 (#121), VF-NEW-PGTEST-orphans (#122), VF-NEW-migrator-lock-snapshot (#123) and VF-NEW-OPENAPI-DOC-INFO (#124).

One entry per finding. Update `status` when it changes. Cite <https://vukem.dev/docs> (page URL + heading). The local `vukem/docs` tree is cited only when the site has no page, tagged `[SITE PAGE MISSING]`.

Severity: `critical` (blocks a roadmap task), `high` (workaround costs more than a day), `medium` (workaround under a day), `low` (DX or docs).
Status: `open`, `reported` (issue or PR link), `fixed in vX.Y.Z`, `wontfix`, `withdrawn`.

| id | date | severity | status |
| --- | --- | --- | --- |
| VF-NEW-EXPORT-WRITER | 2026-11-01 | medium | open |
| VF-NEW-PROBLEM-TYPE | 2026-11-01 | medium | open |
| VF-NEW-HOST-ONLY | 2026-11-01 | medium | open |
| VF-NEW-REQUEST-LINE-FIELDS | 2026-11-01 | low | open |
| VF-063 | 2026-11-01 | low | open |
| VF-064 | 2026-11-01 | low | open |
| VF-NEW-SEC-anti-scraping | 2026-11-01 | low | open |

## VF-NEW-SEC-anti-scraping: no client-IP helper for server middleware and no verified-crawler check

- **Date**: 2026-11-01. **Severity**: low. **Status**: open.
- **vukem version**: v0.35.0. **Docs consulted**: <https://vukem.dev/docs/security/rate-limiting> (`ratelimit.ByIP`, "Keys"). `[NOT IN VUKEM DOCS]`: how plain `func(http.Handler) http.Handler` middleware gets the raw client IP that honours `TRUSTED_PROXIES`; only `c.IP()` on `*web.Ctx` and the hashed or prefixed `ratelimit.ByIP` key exist.
- **Context**: the HexaScore page guard (ADR 0022) runs as `UseServer` middleware and needs the raw IP for reverse-DNS crawler verification (IPv6 included, so the `/64` key of `ByIP` does not fit).
- **Workaround**: `web.NewCtx(nil, r, web.ServicesFrom(r.Context())).IP()` in `app/http/middleware/scrape_guard.go`, marked `workaround(vukem): VF-NEW-SEC-anti-scraping`.
- **Request**: export `web.ClientIP(r *http.Request) string` with the same trusted-proxy rules. Optionally ship a `security/crawlers` helper (rDNS plus forward-confirm with a timeout and a cache) for the Googlebot/Bingbot verification every public site needs.

## VF-NEW-EXPORT-WRITER: export streams only to an HTTP response, so background jobs reimplement CSV/NDJSON

- **Date**: 2026-11-01. **Severity**: medium. **Status**: open.
- **vukem version**: v0.35.0. **Docs consulted**: <https://vukem.dev/docs/digging-deeper/exports> (table "`export.NDJSON(w, r, name, rows, opts...)`"). Every streaming helper (`CSV`, `WriteCSV`, `JSON`, `NDJSON`) takes `http.ResponseWriter` and `*http.Request`; none accepts an `io.Writer`. `[NOT IN VUKEM DOCS]`: how to produce the same output from a queue job.
- **Context**: HexaScore builds two exports outside a request: the account data export (`library.export`, zipped CSV/JSON written to the bucket) and the daily bulk ID exports for the paid API (`app/jobs/export_ids.go`, NDJSON written to S3, served later through a presigned URL). Both re-implement CSV formula guarding and NDJSON line encoding that `export` already has (`workaround(vukem): VF-029` in `app/jobs/export_ids.go`).
- **Repro**: in a job with no request, try to stream an ORM iterator to a `*os.File` or an S3 upload writer with `export.NDJSON`: there is no request and no `ResponseWriter`.
- **Workaround**: hand-written encoders in the app, duplicating the formula guard (CWE-1236) logic.
- **Request**: add writer-level helpers, for example `export.EncodeCSV(io.Writer, header, rows, record, opts...)` and `export.EncodeNDJSON(io.Writer, rows, opts...)`, sharing the same formula guard and flush options; keep the HTTP helpers as thin wrappers over them, and document the job use case on the exports page.

## VF-NEW-PROBLEM-TYPE: problems cannot carry a custom type, title or catalogue code from vukem-raised errors

- **Date**: 2026-11-01. **Severity**: medium. **Status**: open.
- **vukem version**: v0.35.0. **Docs consulted**: <https://vukem.dev/docs/basics/error-handling> (heading "Extension members"): `WithExtension` adds members, but "an extension never replaces a standard member (`type`, `title`, `status`, `detail`, ...)". `[NOT IN VUKEM DOCS]`: setting `type` or `title` on an `errs.HTTPError`, and a hook that decorates every problem, including the ones vukem raises itself (404, 405, typed-route 400s).
- **Context**: HexaScore's problem catalogue gives every problem a `code` and a `type` URL ending in it (SDD 05 section 6.1). The cooldown 429 of the data export also needs its own `title` (`app/controllers/exports_controller.go`).
- **Repro**: `errs.New(429, "x").WithExtension("type", "https://example.test/p/cooldown")` is ignored; `GET /api/v1/nope` answers `type: about:blank` with no `code`.
- **Workaround**: `middleware.APIEdge` rewrites every problem+json under `/api/v1` and `/app/v1` in a response writer wrapper; the export cooldown writes its body by hand.
- **Request**: add `HTTPError.WithType(uri)` and `WithTitle(title)`, and a `Services.ProblemDecorator func(*web.Problem)` (or equivalent) that runs for every problem, so the app can add `code` and `type` once.

## VF-NEW-HOST-ONLY: a host group cannot be the only thing a host serves

- **Date**: 2026-11-01. **Severity**: medium. **Status**: open.
- **vukem version**: v0.35.0. **Docs consulted**: <https://vukem.dev/docs/basics/routing> (heading "Host routes"): host groups match only their host, but "other hosts never reach these routes; they get the main routes or a `404`" says nothing about the reverse. `[NOT IN VUKEM DOCS]`: making a host answer only its own routes (404 problem+json for everything else, no web stack, no session cookie), and serving one handler set under two prefixes (`/api/v1` and the host's `/v1`).
- **Context**: `api.hexascore.app` must serve the API only, with `/v1/...` as the public path of the `/api/v1` group (SDD 05 section 6.1).
- **Repro**: register `r.Host("api.example.test", ...)` beside web routes: `GET /login` with `Host: api.example.test` still reaches the web stack and sets a session cookie.
- **Workaround**: `middleware.APIEdge` rewrites `/v1` to `/api/v1`, answers the index at `/` and a 404 problem for every other path on the API host.
- **Request**: a host-exclusive group (for example `r.Host(host, fn).Only()`) that answers a configurable 404 for unmatched paths on that host, and a way to mount a group under an alias prefix on one host.

## VF-NEW-REQUEST-LINE-FIELDS: the request line cannot carry the route or the user

- **Date**: 2026-11-01. **Severity**: low. **Status**: open.
- **vukem version**: v0.35.0. **Docs consulted**: <https://vukem.dev/docs/deployment/logging> (heading "The request line"): "The line has no user id and no route name: add them yourself ... with `c.Log()` or `slog.Default()`", which writes a second line, not fields on the request line. `[NOT IN VUKEM DOCS]`: a per-request attribute holder the request line reads after the handler ran.
- **Context**: SDD 12 wants one request line with route pattern, hashed user id, truncated IP and trace id, for dashboards and alerts.
- **Repro**: log in, `GET /library`: the `request` line has method, path, status, but no route pattern and no user.
- **Workaround**: `bootstrap/app.go` wraps `app.Log` with `logging.NewHandler` and `middleware.RequestLog` stores a `logging.Fields` holder in the context that inner middleware fill (`workaround(vukem)` comments in `app/http/middleware/request_log.go` and `bootstrap/app.go`).
- **Request**: `web.LogFields(ctx)` (or `c.LogAttr(k, v)`) whose attributes the request line includes, and a `Services.RequestLogAttrs func(*web.Ctx) []slog.Attr` hook run after the response.

## VF-063: the admin hardening options guard only the panel routes, answer 403 and cannot redirect to a 2FA setup

- **Date**: 2026-11-01. **Severity**: low. **Status**: open.
- **vukem version**: v0.35.0. **Docs consulted**: <https://vukem.dev/docs/digging-deeper/admin> (heading "Hardening"): `AllowCIDRs` answers 403 "before any login redirect"; `IdleTimeout` is "updated only by those routes, so using the rest of the app does n[ot refresh it]"; `panel.Mount` "is the only place that installs this stack", so routes beside the panel need `panel.Guard()`; `TwoFactor` is 403 when off.
- **Context**: HexaScore's `AdminGuard` (`app/http/middleware/admin_guard.go`, SDD 13 FR-13-5, `TestSecurity_A01_HexaAdminNeeds2FA`, `TestSecurity_A07_HexaAdminSessionExpiry`, `TestSecurity_A01_HexaAdminIPAllowlist`) must hide the panel from outside the allowlist with a 404 (not a 403 that confirms it exists), apply the 30 minute idle and 12 hour absolute caps to every request of an admin or moderator (also on public pages, so a signed-in moderator who browses elsewhere keeps one clock), and send an admin without TOTP to the 2FA setup page with a flash message.
- **Repro**: set `Config.AllowCIDRs` and request `/admin` from another IP: 403, not 404. Set `IdleTimeout`, browse public pages for longer than the timeout, open `/admin`: signed out, although the user was active. Set `TwoFactor` to a function returning false: 403, no redirect.
- **Workaround**: the app keeps its own router-wide `AdminGuard` and does not set the native options (the native checks would only duplicate it on the panel routes and change the answers the tests pin).
- **Request**: per-option behaviour switches on `admin.Config`, for example `DenyStatus int` (404 for `AllowCIDRs`), `TwoFactorRedirect string` (a path to redirect to instead of 403) and `IdleActivity func(*http.Request) bool` (or a documented `admin.TouchIdle(session, now)` for the app's own middleware), so an app-wide policy can use the native options instead of reimplementing them.

## VF-064: no buffered counter that batches hot-path increments into one periodic write

- **Date**: 2026-11-01. **Severity**: low. **Status**: open.
- **vukem version**: v0.35.0. **Docs consulted**: <https://vukem.dev/docs/digging-deeper/cache> (heading "Counters"): `Increment` is atomic, one store write per call, and the counter lives in the cache store; <https://vukem.dev/docs/digging-deeper/scheduler> (heading "Frequencies"): `EveryMinute()` and `HourlyAt(m)` exist, but no page describes accumulating counts in memory and flushing them on a schedule or at shutdown.
- **Context**: item page views and search clicks (SDD 08 NFR-08-4) must not cost one database write per request; the app adds them in memory and a scheduled task upserts one row per item every minute (`app/recommend/primitives/flush.go`, `routes/schedule.go`), with a flush at shutdown.
- **Repro**: `cache.Increment` on the database store runs one statement per view; on the memory store the count is not in a table the app can query or join, and a restart loses it.
- **Workaround**: an app-owned `sync.Mutex` counter map swapped by `Flush`; a failed flush puts the counts back for the next one.
- **Request**: a small `counter.Buffered[K]` (or `cache.Buffer`) with `Add(key, delta)`, `Swap() map[K]int64`, `Restore(map)` and a shutdown hook, so apps do not reimplement the swap-and-restore, plus a documented scheduler recipe.
