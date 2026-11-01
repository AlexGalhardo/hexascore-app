# vukem feedback from HexaScore

Open queue only. Fixed, withdrawn and wontfix entries are removed; git history keeps them. Do not reuse removed ids.

Removed on 2026-10-31, all resolved on vukem `main` (23cde35a) but not in a release yet: VF-062 (#103), VF-NEW-AUDIT-1 (#107), VF-NEW-AUDIT-2 (#115), VF-NEW-AUDIT-3 (#109), VF-NEW-AUDIT-4 (#116), VF-NEW-AUDIT-5 (#110), VF-NEW-AUDIT-6 (#105), VF-NEW-P2.H.3.1 (#117), VF-NEW-P5.A.3.1 (#118), VF-NEW-FIX-gate-timeouts (#106), VF-NEW-P4.A.3.1b (#119), VF-NEW-P2.H.3.5 (withdrawn, #111), VF-NEW-FIX-covers-route (#120; adoption needs a `drivers/s3` release, Q-A14) and VF-NEW-OWNER-schemathesis (#113). HexaScore adopts them in an `ADOPT-vukem-<version>` row once the owner asks for a vukem release.

Adopted in vukem v0.35.0 (ADOPT-vukem-0.35.0), entries removed: all of the above plus VF-NEW-P4.B.2.5 (#121), VF-NEW-PGTEST-orphans (#122), VF-NEW-migrator-lock-snapshot (#123) and VF-NEW-OPENAPI-DOC-INFO (#124).

Numbered 2026-11-01 (especialista): VF-066..070 were VF-NEW-EXPORT-WRITER, PROBLEM-TYPE, HOST-ONLY, REQUEST-LINE-FIELDS and MAIL-REPLYTO; vukem PR #128 also names VF-063 as VF-NEW-ADMIN-GUARD-SCOPE and VF-064 as VF-NEW-BUFFERED-COUNTER. All seven shipped in vukem v0.36.0. ADOPT-vukem-0.36.0 adopted VF-066, VF-067, VF-069 and VF-070 (entries removed); VF-063, VF-064 and VF-068 stay open with the gap the v0.36.0 API leaves (see each entry). VF-065 is not covered on vukem `origin/main` (343667b1): no exported client-IP helper or crawler check.

One entry per finding. Update `status` when it changes. Cite <https://vukem.dev/docs> (page URL + heading). The local `vukem/docs` tree is cited only when the site has no page, tagged `[SITE PAGE MISSING]`.

Severity: `critical` (blocks a roadmap task), `high` (workaround costs more than a day), `medium` (workaround under a day), `low` (DX or docs).
Status: `open`, `reported` (issue or PR link), `fixed in vX.Y.Z`, `wontfix`, `withdrawn`.

| id | date | severity | status |
| --- | --- | --- | --- |
| VF-068 | 2026-11-01 | medium | open (v0.36.0 partial) |
| VF-063 | 2026-11-01 | low | open (v0.36.0 partial) |
| VF-064 | 2026-11-01 | low | open (v0.36.0 partial) |
| VF-065 | 2026-11-01 | low | open |

## VF-065: no client-IP helper for server middleware and no verified-crawler check

- **Date**: 2026-11-01. **Severity**: low. **Status**: open.
- **vukem version**: v0.35.0. **Docs consulted**: <https://vukem.dev/docs/security/rate-limiting> (`ratelimit.ByIP`, "Keys"). `[NOT IN VUKEM DOCS]`: how plain `func(http.Handler) http.Handler` middleware gets the raw client IP that honours `TRUSTED_PROXIES`; only `c.IP()` on `*web.Ctx` and the hashed or prefixed `ratelimit.ByIP` key exist.
- **Context**: the HexaScore page guard (ADR 0022) runs as `UseServer` middleware and needs the raw IP for reverse-DNS crawler verification (IPv6 included, so the `/64` key of `ByIP` does not fit).
- **Workaround**: `web.NewCtx(nil, r, web.ServicesFrom(r.Context())).IP()` in `app/http/middleware/scrape_guard.go`, marked `workaround(vukem): VF-065`.
- **Request**: export `web.ClientIP(r *http.Request) string` with the same trusted-proxy rules. Optionally ship a `security/crawlers` helper (rDNS plus forward-confirm with a timeout and a cache) for the Googlebot/Bingbot verification every public site needs.

## VF-068: a host group cannot be the only thing a host serves (provisional id VF-NEW-HOST-ONLY)

- **Date**: 2026-11-01. **Severity**: medium. **Status**: open (v0.36.0 partial; vukem PR #128).
- **vukem version**: v0.35.0. **Docs consulted**: <https://vukem.dev/docs/basics/routing> (heading "Host routes"): host groups match only their host, but "other hosts never reach these routes; they get the main routes or a `404`" says nothing about the reverse. `[NOT IN VUKEM DOCS]`: making a host answer only its own routes (404 problem+json for everything else, no web stack, no session cookie), and serving one handler set under two prefixes (`/api/v1` and the host's `/v1`).
- **Context**: `api.hexascore.app` must serve the API only, with `/v1/...` as the public path of the `/api/v1` group (SDD 05 section 6.1).
- **Repro**: register `r.Host("api.example.test", ...)` beside web routes: `GET /login` with `Host: api.example.test` still reaches the web stack and sets a session cookie.
- **Workaround**: `middleware.APIEdge` rewrites `/v1` to `/api/v1`, answers the index at `/` and a 404 problem for every other path on the API host.
- **Request**: a host-exclusive group (for example `r.Host(host, fn).Only()`) that answers a configurable 404 for unmatched paths on that host, and a way to mount a group under an alias prefix on one host.
- **v0.36.0 adoption (ADOPT-vukem-0.36.0)**: `(*router.Group).Only()` and the two-prefix recipe on <https://vukem.dev/docs/basics/routing> ("One set of routes under two prefixes") cover the API routes, but the HexaScore host contract (`TestAPIEdgeHostOnlyAPI`) also serves the framework routes `/up` (registered by vukem with an unexported handler) and `/api/v1/openapi.json` on the API host. `[NOT IN VUKEM DOCS]`: how an `Only()` host group serves the health route. `middleware.APIEdge` keeps the host gate and the `/v1` rewrite; its problem rewriting is gone (VF-067 adopted).
- **Remaining request**: export the health handler (or `Group.Health()`), or let `Only()` take framework routes that should stay reachable on the host.

## VF-063: the admin hardening options guard only the panel routes, answer 403 and cannot redirect to a 2FA setup

- **Date**: 2026-11-01. **Severity**: low. **Status**: open (v0.36.0 partial; vukem PR #128).
- **vukem version**: v0.35.0. **Docs consulted**: <https://vukem.dev/docs/digging-deeper/admin> (heading "Hardening"): `AllowCIDRs` answers 403 "before any login redirect"; `IdleTimeout` is "updated only by those routes, so using the rest of the app does n[ot refresh it]"; `panel.Mount` "is the only place that installs this stack", so routes beside the panel need `panel.Guard()`; `TwoFactor` is 403 when off.
- **Context**: HexaScore's `AdminGuard` (`app/http/middleware/admin_guard.go`, SDD 13 FR-13-5, `TestSecurity_A01_HexaAdminNeeds2FA`, `TestSecurity_A07_HexaAdminSessionExpiry`, `TestSecurity_A01_HexaAdminIPAllowlist`) must hide the panel from outside the allowlist with a 404 (not a 403 that confirms it exists), apply the 30 minute idle and 12 hour absolute caps to every request of an admin or moderator (also on public pages, so a signed-in moderator who browses elsewhere keeps one clock), and send an admin without TOTP to the 2FA setup page with a flash message.
- **Repro**: set `Config.AllowCIDRs` and request `/admin` from another IP: 403, not 404. Set `IdleTimeout`, browse public pages for longer than the timeout, open `/admin`: signed out, although the user was active. Set `TwoFactor` to a function returning false: 403, no redirect.
- **Workaround**: the app keeps its own router-wide `AdminGuard` and does not set the native options (the native checks would only duplicate it on the panel routes and change the answers the tests pin).
- **Request**: per-option behaviour switches on `admin.Config`, for example `DenyStatus int` (404 for `AllowCIDRs`), `TwoFactorRedirect string` (a path to redirect to instead of 403) and `IdleActivity func(*http.Request) bool` (or a documented `admin.TouchIdle(session, now)` for the app's own middleware), so an app-wide policy can use the native options instead of reimplementing them.
- **v0.36.0 adoption (ADOPT-vukem-0.36.0)**: not adopted. `DenyStatus`, `TwoFactorRedirect` and `TrackIdle` (<https://vukem.dev/docs/digging-deeper/admin>, "Hardening") still differ from SDD 13 FR-13-5 as pinned by the admin `TestSecurity_*` tests: `TrackIdle` only clocks sessions that already used the panel (a moderator who never opened it is not capped), `MaxSessionAge` counts from session creation, not from the first admin request, `TwoFactorRedirect` cannot flash the reason, and the options cover only panel routes (`Panel.Guard()`), not the app's other `/admin/*` pages. `AdminGuard` stays.
- **Remaining request**: an app-wide variant of the clock for every request of a user with an admin ability, and a hook (or flash key) on the 2FA redirect.

## VF-064: no buffered counter that batches hot-path increments into one periodic write

- **Date**: 2026-11-01. **Severity**: low. **Status**: open (v0.36.0 partial; vukem PR #128).
- **vukem version**: v0.35.0. **Docs consulted**: <https://vukem.dev/docs/digging-deeper/cache> (heading "Counters"): `Increment` is atomic, one store write per call, and the counter lives in the cache store; <https://vukem.dev/docs/digging-deeper/scheduler> (heading "Frequencies"): `EveryMinute()` and `HourlyAt(m)` exist, but no page describes accumulating counts in memory and flushing them on a schedule or at shutdown.
- **Context**: item page views and search clicks (SDD 08 NFR-08-4) must not cost one database write per request; the app adds them in memory and a scheduled task upserts one row per item every minute (`app/recommend/primitives/flush.go`, `routes/schedule.go`), with a flush at shutdown.
- **Repro**: `cache.Increment` on the database store runs one statement per view; on the memory store the count is not in a table the app can query or join, and a restart loses it.
- **Workaround**: an app-owned `sync.Mutex` counter map swapped by `Flush`; a failed flush puts the counts back for the next one.
- **Request**: a small `counter.Buffered[K]` (or `cache.Buffer`) with `Add(key, delta)`, `Swap() map[K]int64`, `Restore(map)` and a shutdown hook, so apps do not reimplement the swap-and-restore, plus a documented scheduler recipe.
- **v0.36.0 adoption (ADOPT-vukem-0.36.0)**: not adopted. `cache.Buffer[K]` (<https://vukem.dev/docs/digging-deeper/cache>, "Buffered counters") is unbounded and holds one `int64` per key; SDD 08 section 8 caps the trending buffer at `MaxKeys` (memory bound against unique-slug floods) with a drop alert and keeps views and clicks per key. `app/recommend/primitives.Counter` stays.
- **Remaining request**: an optional key cap with a drop callback on `cache.Buffer` (oldest keys dropped first).
