# vukem feedback from HexaScore

Long-term log of bugs, problems, DX improvements and feedback found while building HexaScore on vukem (`github.com/AlexGalhardo/vukem`). One entry per finding; never delete, only update `status`. Entries are written for the vukem maintainers: reproducible, with the vukem version and the doc or file that was consulted.

Documentation reference (owner rule, 2026-10-28): every entry cites the public docs site <https://vukem.dev/docs> (page URL + heading). The local checkout `vukem/docs/*.md` is cited only when the site has no page for the fact, tagged `[SITE PAGE MISSING]`.

Severity: `critical` (blocks a roadmap task), `high` (workaround costs more than a day), `medium` (workaround under a day), `low` (DX or docs).
Status: `open`, `reported` (issue link), `fixed in vX.Y.Z`, `wontfix`, `withdrawn`.

| id | date | severity | status |
| --- | --- | --- | --- |
| VF-001 | 2026-10-28 | high | wontfix (post-1.0, FQ-001) |
| VF-002 | 2026-10-28 | medium | merged in main (PR #7), unreleased |
| VF-003 | 2026-10-28 | medium | open |
| VF-004 | 2026-10-28 | low | open |
| VF-005 | 2026-10-28 | medium | merged in main (PR #11), unreleased |
| VF-006 | 2026-10-28 | low | open |
| VF-007 | 2026-10-28 | low | open |
| VF-008 | 2026-10-28 | medium | open |
| VF-009 | 2026-10-28 | low | open |
| VF-010 | 2026-10-28 | low | open |
| VF-011 | 2026-10-28 | low | open |
| VF-012 | 2026-10-28 | low | open |
| VF-013 | 2026-10-28 | low | open |
| VF-014 | 2026-10-28 | low | open |
| VF-015 | 2026-10-28 | medium | merged in main (PR #12), unreleased |
| VF-016 | 2026-10-28 | low | open |
| VF-017 | 2026-10-28 | medium | in_progress (framework, vukem PR #13) |
| VF-018 | 2026-10-28 | medium | open |
| VF-019 | 2026-10-28 | medium | open |
| VF-020 | 2026-10-28 | low | open |
| VF-021 | 2026-10-28 | low | withdrawn (2026-10-29: `HourlyAt` documented) |
| VF-022 | 2026-10-28 | low | open |
| VF-023 | 2026-10-28 | low | open |
| VF-024 | 2026-10-28 | medium | merged in main (PR #10), unreleased |
| VF-025 | 2026-10-28 | medium | open |
| VF-026 | 2026-10-28 | medium | in_progress (framework, vukem PR #13) |
| VF-027 | 2026-10-28 | medium | merged in main (PR #9), unreleased |
| VF-028 | 2026-10-28 | low | open |
| VF-029 | 2026-10-28 | low | open |
| VF-030 | 2026-10-28 | medium | open |
| VF-031 | 2026-10-28 | medium | open |
| VF-032 | 2026-10-28 | medium | open |
| VF-033 | 2026-10-28 | low | open |
| VF-034 | 2026-10-28 | low | open |
| VF-035 | 2026-10-29 | medium | open |
| VF-036 | 2026-10-29 | low | open |
| VF-037 | 2026-10-29 | medium | open |
| VF-038 | 2026-10-29 | medium | open |
| VF-039 | 2026-10-29 | low | open |
| VF-040 | 2026-10-29 | medium | open |
| VF-041 | 2026-10-29 | low | open |
| VF-042 | 2026-10-29 | low | open |
| VF-043 | 2026-10-29 | low | open |
| VF-044 | 2026-10-29 | low | open |
| VF-045 | 2026-10-29 | low | open |

## VF-001: No passkey / WebAuthn support in `auth`

- **Date**: 2026-10-28. **Severity**: high. **Status**: wontfix (post-1.0, FQ-001): vukem keeps passkeys/WebAuthn on its post-1.0 list (docs/security.md); owner chose option (a) on 2026-10-29.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/security/two-factor` (Concepts, Guards, Hashing [heading not on page 2026-10-29], `install:auth`, Two-factor, Social login).
- **Context**: HexaScore's legacy app (Better Auth 1.7.7) offers passkeys. The rewrite drops them at launch (brainstorming D-06) because the framework has no WebAuthn registration/assertion flow, credential storage or recovery path.
- **Reproduction**: search `docs/` and `auth/` for `webauthn`, `passkey`, `FIDO`: no matches.
- **Workaround**: none in-framework; a third-party Go WebAuthn library would be a new dependency and a parallel credential model. Deferred feature.
- **Request**: a `install:auth --passkeys` option with credential table, registration/login endpoints, and `TestSecurity_*` coverage.

## VF-002: Inertia SSR falls back to client rendering silently by default

- **Date**: 2026-10-28. **Severity**: medium. **Status**: merged in main (PR #7, <https://github.com/AlexGalhardo/vukem/pull/7>), unreleased. Native API: `inertia.SSRConfig{Strict: true}` (503 + `Retry-After`, auto `inertia-ssr` readiness check) or per-request `inertia.RequireSSR(ctx)`.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/frontend/inertia` "Inertia v3 (BETA)" [heading not on page 2026-10-29] → SSR row: "falls back to client-side rendering on any failure (reported to `OnFailure`, default a warning log)".
- **Context**: for a catalogue that depends on search indexing (ADR 0004), a silent fallback hides an SEO incident. HexaScore needs per-route policy: public routes must fail readiness or page an alert when SSR fails; logged-in routes may fall back.
- **Reproduction**: stop the SSR process; request a public page; observe a 200 with client-rendered shell and only a warning log.
- **Workaround**: implement `OnFailure` to increment a metric and alert; add the SSR process to `AddHealthCheck` so `/up?checks=1` reports it.
- **Request**: a `SSRConfig.Strict` (or per-route `inertia.RequireSSR(ctx)`) that returns 503 with `Retry-After` instead of falling back, plus a built-in health check for the SSR endpoint.

## VF-003: No built-in bot/abuse challenge for auth and form routes

- **Date**: 2026-10-28. **Severity**: medium. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/security/rate-limiting` "Rate limiting", `https://vukem.dev/docs/security/rate-limiting` "Rate limiting", `https://vukem.dev/docs/security/overview`.
- **Context**: the legacy app used Vercel BotID on signup, login and contact. vukem offers GCRA rate limiting by IP/token and login throttling, but no challenge mechanism (proof-of-work, CAPTCHA adapter) when thresholds trip.
- **Reproduction**: n/a (feature gap).
- **Workaround**: rate limits plus honeypot fields plus an optional third-party challenge (Cloudflare Turnstile) wired by hand; specified in SDD 13.
- **Request**: a `middleware.Challenge` hook point with a reference adapter (Turnstile/hCaptcha) and a stdlib proof-of-work option.

## VF-004: `contrib/imaging` flagged as needing work in the framework's own review

- **Date**: 2026-10-28. **Severity**: low. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `vukem/vukem-feedback.md` (framework's own review, item FUN-08 / FB-10 on `contrib/imaging`).
- **Context**: HexaScore's cover pipeline (ADR 0007) needs WebP encode at 600 px and 96×144 thumbnails with deterministic output matching the legacy Python pipeline. The framework review marks imaging as confirmed-needs-work with large effort.
- **Reproduction**: pending; SDD 06 defines a fixture set and expected hashes to test against.
- **Workaround**: if WebP encoding is unavailable, store JPEG/PNG and serve with content negotiation, or run a one-off conversion job later.
- **Request**: document supported formats and determinism guarantees; WebP encode support.

## VF-005: No keyset (cursor) pagination helper in the ORM

- **Date**: 2026-10-28. **Severity**: medium. **Status**: merged in main (PR #11, <https://github.com/AlexGalhardo/vukem/pull/11>), unreleased. Native API: `orm.QueryOf[T](db).OrderBy("score", orm.Desc).CursorPaginate(ctx, cursor, n, orm.Desc)` pages by `(score DESC, id DESC)`; cursor values typed per column, unsigned by design (bound parameters, validated columns).
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/database/orm` "Querying"; `https://vukem.dev/docs/rest-apis/pagination` "Query conventions"; `https://vukem.dev/docs/frontend/inertia` Inertia `ScrollCursor(orm.CursorPage, name)`.
- **Context**: `ScrollCursor` accepts an `orm.CursorPage`, but the documented query builder shows no way to build a keyset page over a composite order (`score DESC, id DESC`) with an opaque signed cursor. HexaScore lists (SDD 01, 05, 07) need this for every listing and for the paid API.
- **Reproduction**: search `https://vukem.dev/docs/database/overview` for `Cursor`: only the Inertia prop is documented.
- **Workaround**: the app builds the predicate with `WhereRaw` and signs/encodes the cursor itself.
- **Request**: document `orm.CursorPage` construction and add a `Query.CursorAfter(cursor, orderColumns...)` helper with a signed cursor.

## VF-006: Postgres mapping of `t.JSON` and `t.Enum` undocumented

- **Date**: 2026-10-28. **Severity**: low. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/database/migrations` "Migrations".
- **Context**: SDD 01 needs `jsonb` columns with GIN indexes and enum-like columns. The migration builder's column types for these on Postgres are not documented, so the DDL sketches use `s.Raw`.
- **Reproduction**: n/a (docs gap).
- **Workaround**: `s.Raw` statements in migrations.
- **Request**: a table in `https://vukem.dev/docs/database/overview` with each builder type and its SQL per dialect, and `jsonb` + GIN index support in the builder.

## VF-007: No sitemap helper

- **Date**: 2026-10-28. **Severity**: low. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/basics/routing`, `https://vukem.dev/docs/digging-deeper/queue`, `https://vukem.dev/docs/rest-apis/overview`.
- **Context**: a catalogue with hundreds of thousands of URLs needs a sitemap index with gzip chunks of 50,000 URLs, `lastmod`, and `hreflang` alternates.
- **Workaround**: app handler streaming XML (SDD 01/09).
- **Request**: a `sitemap` package (index + chunks, streaming, `hreflang`) in core or contrib.

## VF-008: Admin panel has no custom row actions, bulk actions, assignment or diff view

- **Date**: 2026-10-28. **Severity**: medium. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/digging-deeper/admin`.
- **Context**: moderation of community suggestions (SDD 04) and anti-abuse review (SDD 02) need per-row actions ("approve", "reject with reason", "revert"), bulk actions, assignment to a moderator and a before/after diff. The documented admin panel generates CRUD from models and `z` schemas with an audit log.
- **Workaround**: a custom admin page in the app at `/admin/moderation` using the admin layout and auth.
- **Request**: `admin.Action` (row and bulk) with confirmation and audit; a diff renderer for JSON patch-like fields.

## VF-009: Generated columns and the ORM on insert/update

- **Date**: 2026-10-28. **Severity**: low. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/database/orm` "Defining models", "Querying".
- **Context**: `ratings.overall` is a Postgres generated column (SDD 02). It is unclear whether `orm.Create`/`Save` omit generated columns from INSERT/UPDATE.
- **Reproduction**: pending at implementation (test: create a model with a generated column on Postgres and call `orm.Create`).
- **Workaround**: raw SQL for inserts on that table.
- **Request**: a `generated` tag or documented behaviour.

## VF-010: Admin `BeforeSave` hook does not expose previous values

- **Date**: 2026-10-28. **Severity**: low. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/digging-deeper/admin`.
- **Context**: the catalogue audit log (SDD 04) needs before/after values for every admin edit; the hook as documented receives only the new values.
- **Workaround**: the hook re-reads the row before saving.
- **Request**: pass the previous record to `BeforeSave`/`AfterSave`.

## VF-011: No persistent in-app notifications

- **Date**: 2026-10-28. **Severity**: low. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/digging-deeper/mail` (Mail, Events, Broadcasting [heading not on page 2026-10-29]).
- **Context**: suggestion decisions (SDD 04) and moderation outcomes need a per-user inbox with read state; vukem has mail and real-time broadcasting but no stored notification model.
- **Workaround**: app-owned `user_notifications` table and page.
- **Request**: a `notify` package (database channel, mail channel, mark-as-read, unread count) like Laravel notifications.

## VF-012: No admin impersonation ("login as user")

- **Date**: 2026-10-28. **Severity**: low. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/security/authentication`, `https://vukem.dev/docs/digging-deeper/admin`.
- **Context**: the legacy app (Better Auth admin plugin) let admins impersonate a user to reproduce support issues, with `session.impersonated_by` recorded. vukem documents no equivalent. HexaScore v1 ships without it (SDD 03); admins use read-only views.
- **Workaround**: none safe in-app; read-only "view as" pages.
- **Request**: `auth.Impersonate(ctx, adminID, userID)` creating a flagged session with a time limit, an audit entry and a visible banner; `auth.StopImpersonating`.

## VF-013: Locale for mail sent from queued jobs

- **Date**: 2026-10-28. **Severity**: low. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/digging-deeper/queue` "Mail", "i18n", "Queue".
- **Context**: transactional mail (verification, reset, suggestion decisions) is often sent from a job without a request context; the docs describe locale resolution through the locale middleware and request context only.
- **Workaround**: the job payload carries the user's locale and the mailer is called with an explicit locale value.
- **Request**: document `i18n.WithLocale(ctx, locale)` (or equivalent) for use outside requests, and let mail templates take the locale from the context set that way.

## VF-014: Envelope and problem key casing is fixed to snake_case

- **Date**: 2026-10-28. **Severity**: low. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/rest-apis/overview` "Envelopes" [heading not on page 2026-10-29], "Errors".
- **Context**: HexaScore payload fields are camelCase (legacy API contract and SDKs); the documented envelope keys (`next_cursor`, `request_id`) are snake_case and no option to change them is documented, so responses mix conventions.
- **Workaround**: accept the mix and document it; or bypass the envelope helper and write the envelope by hand.
- **Request**: an envelope/problem key naming option (`api.KeyCase(api.CamelCase)`) or documented custom envelope types.

## VF-015: No host-based routing helper

- **Date**: 2026-10-28. **Severity**: medium. **Status**: merged in main (PR #12, https://github.com/AlexGalhardo/vukem/pull/12), unreleased. Native API: `r.Host("api.hexascore.app", func(g *router.Group) {...})` (ServeMux host patterns, port ignored); `r.URL` returns `//host/path` for host routes; `RouteInfo.Host`.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/basics/routing` "Routing", `https://vukem.dev/docs/rest-apis/overview` "Routing", "Versioning".
- **Context**: the paid API is served at `api.hexascore.app/v1/...` while the app also serves `hexascore.app`. The router documents path groups only. The app needs a `Host`-matched group (or a middleware that rewrites `/v1/*` to `/api/v1/*` when the host matches) plus per-host CORS and security headers.
- **Workaround**: app middleware inspecting `r.Host` and rewriting `r.URL.Path`, registered before the router; OpenAPI `servers` entry set by hand.
- **Request**: `router.Host("api.example.com", func(g *router.Group) {...})` with URL generation aware of the host.

## VF-016: Problem+json extension members undocumented

- **Date**: 2026-10-28. **Severity**: low. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/rest-apis/overview` "Errors", `https://vukem.dev/docs/basics/error-handling` errors section.
- **Context**: RFC 9457 allows extension members; HexaScore's API uses `code`, `plan`, `limit`, `retryAfter` on quota and plan problems (legacy `lib/api/problem.ts`). The docs do not show how to add members to `errs.HTTPError` output or to the problem catalogue entries used by OpenAPI.
- **Workaround**: a custom error type implementing the problem interface, if one exists, or a hand-written handler for API routes.
- **Request**: `errs.WithExtension(key, value)` and OpenAPI schema generation for documented extensions.

## VF-017: No non-transactional migrations (`CREATE INDEX CONCURRENTLY`)

- **Date**: 2026-10-28. **Severity**: medium. **Status**: in_progress (framework).
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/database/migrations` "Migrations", "Managing them with migrate".
- **Context**: Postgres refuses `CREATE INDEX CONCURRENTLY` inside a transaction. HexaScore's search indexes (SDD 07) and later index additions on the 100k-row `items` table and larger `ratings` table must be built without locking writes in production. The docs do not show a per-migration "no transaction" flag.
- **Workaround**: a one-off script run by hand, or building indexes before the table receives traffic.
- **Request**: `schema.Migration{Transactional: false}` (or a `//vukem:no-transaction` marker) with the runner executing such migrations outside a transaction and recording them like the others.

## VF-018: Queue priority and per-queue concurrency with the database driver inside `serve`

- **Date**: 2026-10-28. **Severity**: medium. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/digging-deeper/scheduler` "Queue", "Scheduler" [heading not on page 2026-10-29]; `https://vukem.dev/docs/cli/reference-runtime` `serve`.
- **Context**: HexaScore runs HTTP, scheduler and queue workers in one process under a US$30/month budget (brainstorming D-08). Ingestion jobs (SDD 06) must run at low concurrency (1–2) and lower priority than user-facing jobs (mail, recompute). The docs do not state whether named queues, priorities and per-queue worker counts apply to the database driver when workers run inside `serve`, nor how to cap CPU use of workers.
- **Reproduction**: n/a (docs gap).
- **Workaround**: a single worker pool with app-level semaphores per job family and a scheduler window at night.
- **Request**: document named queues, priority and `Workers(queue, n)` for the database driver in `serve`; expose a "pause queue" switch for operators.

## VF-019: No documented API for app-level OpenTelemetry metrics

- **Date**: 2026-10-28. **Severity**: medium. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/digging-deeper/opentelemetry` "OpenTelemetry", "Logging & observability" [heading not on page 2026-10-29]; `drivers/otel` description in `https://vukem.dev/docs/architecture/lifecycle`.
- **Context**: HexaScore needs custom counters and histograms (ingestion items/s, SSR fallbacks, suggestion queue age, quota flushes) exported with the framework's traces. The docs cover tracing setup and built-in instrumentation only.
- **Workaround**: the app wires `go.opentelemetry.io/otel/metric` itself and shares the exporter configuration.
- **Request**: `otel.Meter()` accessor and a documented pattern for app metrics, plus a Prometheus-compatible `/metrics` option for hosts without an OTLP collector.

## VF-020: `httpclient` lacks per-host throttling and circuit breaking

- **Date**: 2026-10-28. **Severity**: low. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/architecture/lifecycle` (`httpclient/` SSRF protection), `https://vukem.dev/docs/digging-deeper/queue`.
- **Context**: ingestion (SDD 06) must respect per-host rate limits (IGDB 4 rps / 8 in flight; TMDB 429 handling; Wikimedia etiquette) and stop hammering a failing host. The SSRF-safe client has no documented rate limiter, retry/backoff policy or circuit breaker.
- **Workaround**: app-level wrapper with `ratelimit.NewMemory` per host, exponential backoff on 429/5xx honouring `Retry-After`, and a simple breaker.
- **Request**: `httpclient.WithRateLimit(host, limit)`, `WithRetry(policy)` and `WithBreaker(...)` options.

## VF-021: No buffered counter primitive and no documented `HourlyAt` schedule helper

- **Date**: 2026-10-28. **Severity**: low. **Status**: withdrawn (2026-10-29: <https://vukem.dev/docs/digging-deeper/scheduler> "Frequencies" now lists `HourlyAt(7)`; verified during P0.F.6.1).
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/digging-deeper/scheduler` "Cache", "Scheduler" [heading not on page 2026-10-29], "Events".
- **Context**: trending primitives (SDD 08) and API quota metering (SDD 05) need an in-process counter that aggregates increments and flushes to Postgres once per minute (ADR 0003, no Redis). The scheduler docs show cron expressions; a typed `HourlyAt(minute)` helper used in the SDD sketch is not documented.
- **Workaround**: app-level `sync.Map` counter with a ticker flush; cron expressions for schedules.
- **Request**: `cache.Counter` with `Incr`/`FlushEvery(d, fn)` semantics, and typed schedule helpers (`EveryMinute`, `HourlyAt`, `DailyAt`) in the scheduler docs.

## VF-022: Request logger lacks privacy options (hashed user id, truncated IP, request id = trace id)

- **Date**: 2026-10-28. **Severity**: low. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/digging-deeper/opentelemetry`, `https://vukem.dev/docs/basics/middleware` (Logger, RequestID); local `vukem/docs/subsystems.md` "Logging & observability" `[SITE PAGE MISSING]` for the request-log fields.
- **Context**: LGPD-minded logging (SDD 12/13) needs the user id hashed, the client IP truncated (/24, /48) and the request id equal to the trace id so logs and traces join. None of these are documented options of the request logger.
- **Workaround**: a custom logging middleware replacing the built-in one.
- **Request**: `middleware.LoggerOptions{HashUserID, TruncateIP, RequestIDFromTrace}` documented on the site.

## VF-023: Site documentation gaps found while writing the observability spec

- **Date**: 2026-10-28. **Severity**: low. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: <https://vukem.dev/docs> (full page map) versus local `vukem/docs/*.md`.
- **Context**: facts that exist only in the repository docs, tagged `[SITE PAGE MISSING]` in the HexaScore SDDs: health-check API (`AddHealthCheckTimeout`, `ReadyHandler`, `ok`/`fail` output format), route name and user id fields in request logs, `logging.NewDedup` and `WithLogger` behaviour with a custom logger, the `TestSecurity_A09_*`/`A10_*` test names, tooling/release process, stability policy, performance baselines, and a metric hook for scheduled tasks.
- **Workaround**: cite the local file.
- **Request**: publish those sections on vukem.dev (an "Operations" page for health checks and logging; "Releases and stability" page).

## VF-024: `i18n.Middleware` has no URL-prefix locale source

- **Date**: 2026-10-28. **Severity**: medium. **Status**: merged in main (PR #10, <https://github.com/AlexGalhardo/vukem/pull/10>), unreleased. Native API: `i18n.MiddlewareConfig{PathPrefixes: map[string]string{"pt-br": "pt-BR"}}` + `i18n.SourcePath`; `StripPathPrefix` installed by `DefaultMiddlewareWith`; `i18n.PathPrefix(ctx)`, `i18n.LocalizePath(ctx, path)`.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/digging-deeper/localization`, `https://vukem.dev/docs/frontend/inertia-i18n`.
- **Context**: HexaScore keeps the legacy URL scheme (`/`, `/pt-br/...`, `/es/...`, brainstorming D-17). The documented locale sources are query, cookie and `Accept-Language`; a path-prefix source with prefix stripping and URL generation that re-adds the prefix is not documented.
- **Workaround**: app middleware that reads the first path segment, sets the locale in the context and rewrites the path; `route` helpers wrapped to add the prefix.
- **Request**: `i18n.SourcePath(prefixes map[string]string)` plus prefix-aware `URLFor`.

## VF-025: No cache hook for Inertia SSR output

- **Date**: 2026-10-28. **Severity**: medium. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/frontend/inertia` (SSR).
- **Context**: public catalogue pages are rendered for anonymous users thousands of times with identical props; the SSR call (Node, 3 s timeout) runs on every request. A documented hook to cache the rendered `head`/`body` by a key (component + props hash + locale) for a short TTL would cut CPU in the single container (D-08).
- **Workaround**: an app middleware caching the whole HTML response for anonymous requests, which must also handle the per-request CSP nonce (re-inject the nonce into cached markup).
- **Request**: `SSRConfig.Cache cache.Store` with TTL and a key function, nonce-safe.

## VF-026: Migration runner has no documented lock for concurrent starts

- **Date**: 2026-10-28. **Severity**: medium. **Status**: in_progress (framework).
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/database/migrations`, `https://vukem.dev/docs/deployment/docker-and-railway`.
- **Context**: zero-downtime deploys can start a new container while the old one still runs; if migrations run at boot, two runners can race. The docs do not mention a Postgres advisory lock or a `migrations` table row lock.
- **Workaround**: the app takes `pg_advisory_lock` before calling the runner, or migrations run as a separate pre-deploy command.
- **Request**: document the locking behaviour of the runner per dialect, or add it.

## VF-027: `serve` and `build` gaps for the SSR sidecar and release metadata

- **Date**: 2026-10-28. **Severity**: medium. **Status**: merged in main (PR #9, <https://github.com/AlexGalhardo/vukem/pull/9>), unreleased. Native API: `bin/app serve -with-ssr` (supervised SSR runner), `vukem build -ssr` and `-ldflags`, `vukem.WithVersion(v)` (`version` in `/up`).
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/cli/reference-runtime`, `https://vukem.dev/docs/frontend/inertia`, `https://vukem.dev/docs/deployment/single-binary`.
- **Context**: HexaScore runs the Node SSR process inside the same container as the Go app (D-08). The docs show `vukem ssr:start` as a separate command and `vukem build` without the SSR bundle; there is no documented way to have `serve` supervise the SSR process in production, to expose the build version at `/up`, or to pass extra `-ldflags` through `vukem build`.
- **Workaround**: a `build:ssr` step (`vite build --ssr`) in the pipeline; the app starts and restarts the Node process itself; version injected by a plain `go build -ldflags` step and reported by a custom health field.
- **Request**: `serve --with-ssr` (supervised child with restart and health), `vukem build --ssr`, `--ldflags` pass-through, and a `version` field in `/up`.

## VF-028: Root view data has no theme hook

- **Date**: 2026-10-28. **Severity**: low. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/frontend/inertia` (root view, `inertia.RootData`).
- **Context**: dark mode without a flash of the wrong theme needs `data-theme` on `<html>` from a cookie at render time; the root view receives only `inertia.RootData` plus view functions.
- **Workaround**: a custom view function reading the theme cookie.
- **Request**: `inertia.Config.RootData func(r *http.Request) map[string]any` to add arbitrary root-view values.

## VF-029: Export helpers only target `http.ResponseWriter`

- **Date**: 2026-10-28. **Severity**: low. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/digging-deeper/exports`.
- **Context**: the LGPD data export (SDD 14) builds a ZIP of CSV/JSON files in a background job and uploads it to storage behind a signed URL; the documented streaming CSV/JSON/NDJSON helpers (with the CSV formula guard) write to a response, not to an `io.Writer`.
- **Workaround**: re-implement the formula guard and encoders for the job, or call the helpers with an in-memory `ResponseRecorder`.
- **Request**: `export.CSV(w io.Writer, ...)` variants usable outside a request, sharing the same safety rules.

## VF-030: No admin hardening options (mandatory 2FA, shorter sessions, IP allowlist)

- **Date**: 2026-10-28. **Severity**: medium. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/digging-deeper/admin`, `https://vukem.dev/docs/security/two-factor`, `https://vukem.dev/docs/basics/sessions`.
- **Context**: SDD 13 requires admins to have 2FA enabled, a shorter absolute session lifetime than regular users, and an optional IP allowlist for `/admin`. The admin panel documents role gating and an audit log, but none of these three controls.
- **Workaround**: app middleware on the `/admin` group (checks `two_factor_enabled`, applies a separate session timeout, compares the client IP with a configured CIDR list).
- **Request**: `admin.Config{Require2FA, SessionTTL, AllowCIDRs}` with `TestSecurity_*` coverage.

## VF-031: No re-encryption command after `APP_KEY` rotation

- **Date**: 2026-10-28. **Severity**: medium. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/security/encryption`, `https://vukem.dev/docs/configuration`.
- **Context**: encrypted columns (2FA secrets, OAuth tokens) are keyed by `APP_KEY`; the encryption page describes key rotation for cookies/values but no CLI to re-encrypt stored columns under the new key, so a rotation leaves rows readable only with the previous key.
- **Workaround**: keep the previous key configured as a secondary decrypt key and run a one-off app job that reads and rewrites every encrypted column.
- **Request**: `vukem crypt:rotate --from <old> --to <new>` that walks registered encrypted columns in batches, plus docs for the previous-key grace window.

## VF-032: `install:auth --locales` rejects `es`

- **Date**: 2026-10-28. **Severity**: medium. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/cli/reference-install` ("`vukem install:auth`"), `https://vukem.dev/docs/digging-deeper/localization`.
- **Context**: `vukem install:auth -social google -2fa -locales en,pt-BR,es` fails with `usage error: no auth message catalog for locale "es" (one of en, pt-BR); translate the generated catalog instead`. HexaScore needs `en`, `pt-BR`, `es` (D-xx i18n).
- **Repro**: run the command above in a fresh `vukem new --inertia react` app; exit code 2 and nothing is generated. With `-locales en,pt-BR` it succeeds.
- **Workaround**: generate with `en,pt-BR`, then copy `app/accounts/lang/en/*.json` to `app/accounts/lang/es/` and translate (text-editor task).
- **Request**: ship an `es` auth catalog, or accept any locale and generate the `en` copy for it.

## VF-033: `vukemtest.New` always uses in-memory SQLite

- **Date**: 2026-10-28. **Severity**: low. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/testing/overview`.
- **Context**: scaffold tests generated by `install:auth` pass without any database server because `vukemtest.New` sets `DATABASE_URL` to a shared in-memory SQLite file; `TEST_DATABASE_URL` is only read by `pgtest.Database`. So "tests green" does not prove Postgres compatibility of the scaffold.
- **Workaround**: `url := pgtest.Database(t, ""); vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{"DATABASE_URL": url}})` (verified, `TestRegisterOnPostgres` passes).
- **Request**: document this pattern on the testing page, or add `vukemtest.Options.Postgres bool`.

## VF-034: OpenAPI cannot map a host-rewritten public path (`/v1`) to the `/api/v1` router group

- **Date**: 2026-10-28. **Severity**: low. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/rest-apis/openapi`, `https://vukem.dev/docs/cli/reference-runtime`, `https://vukem.dev/docs/digging-deeper/billing` ("Routes", `r.UseServer`).
- **Context**: `api.<host>/v1/*` is served from the `/api/v1` group through an app `r.UseServer` rewrite (VF-015). `openapi.Info.Servers` can list `https://api.<host>/v1`, but the document keeps router paths (`/api/v1/ping`), so the full URL becomes `/v1/api/v1/ping`. `Servers` is a static option (no config key; only the `APP_URL` default), so it needs a process env read in `bootstrap.New`.
- **Repro**: spike `agent/senior/P0.F.4.1`: `API_URL=<http://api.localhost:3001/v1> vukem openapi:export -prefix /api/v1 | jq -r '.servers[0].url, (.paths|keys[])'` gives `http://api.localhost:3001/v1` and `/api/v1/ping`.
- **Workaround**: post-process the export: `jq '.paths |= with_entries(.key |= sub("^/api/v1";""))'` gives paths `/ping`, matching the server URL.
- **Request**: `openapi.Info.StripPrefix` (or `PathPrefix` stripping) and an `OPENAPI_SERVERS` config key. Also note: in Git Bash `-prefix //api/v1` was passed through literally and failed (`path prefix matches no route "//api/v1"`); `MSYS_NO_PATHCONV=1 ... -prefix /api/v1` works.

## VF-035: Migration builder lacks uuid/composite primary keys, smallint and generated columns; ORM has a single-column key

- **Date**: 2026-10-29. **Severity**: medium. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/database/migrations` ("Creating migrations"), `https://vukem.dev/docs/database/orm` ("Defining models").
- **Context**: SDD 01 `items` (uuid primary key) and SDD 02 `ratings` (composite key `(user_id, item_id)`, `smallint` scores, stored generated `overall`) cannot be expressed with `schema.Table`: no `smallint`, no primary key on a `t.UUID` column, no composite primary key, no generated column. `t.JSON` is `jsonb` and `t.Enum` is `text` + CHECK on PostgreSQL (verified, spike `TestGeneratedOverallColumn`), which closes the doubt in VF-006 for those two types. The ORM filters `Update`/`Delete` by one key column (`Keyer.PrimaryKey`), so on a composite-key table `orm.Update`/`orm.Delete` would match every row of the first key.
- **Workaround**: `s.Raw` for whole `CREATE TABLE` statements (guard with `s.Dialect() == "postgres"`); ORM model `Rating` with `PrimaryKey() = "user_id"` used only for `orm.Create` and `orm.QueryOf[Rating](db).Where("user_id", u).Where("item_id", i).Update(...)`; never `orm.Update`/`orm.Delete` on it.
- **Request**: `t.SmallInteger`, `t.Primary(cols...)`/`Column.Primary()`, `Column.GeneratedAs(expr).Stored()`, and composite keys in `Keyer` (or a loud error when a model with a non-unique key calls `Update`/`Delete`).

## VF-036: ORM column naming splits acronyms (`ExternalIDs` becomes `external_i_ds`)

- **Date**: 2026-10-29. **Severity**: low. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/database/orm` ("Defining models").
- **Context**: `naming.Snake("ExternalIDs")` yields `external_i_ds`; `orm.Create` failed with `column "external_i_ds" of relation "items" does not exist`.
- **Workaround**: explicit `db:"external_ids"` tag on fields ending in an acronym (`ID`, `URL`, `IDs`).
- **Request**: treat trailing runs of capitals as one word (`ExternalIDs` becomes `external_ids`), or document the rule and the `db` tag on the models page.

Update to VF-009 (2026-10-29, P0.F.5.1): the ORM writes every mapped column, so a model that maps a generated column fails on `orm.Create` with `cannot insert a non-DEFAULT value into column "overall"` (SQLSTATE 428C9). Map the column with `db:"-"` (never written, never read) or read it through a separate read-only model. Insert and update through the ORM then succeed. A `generated` tag is still the request.

## VF-037: `install:billing` has no per-currency prices and its plan names are fixed in code

- **Date**: 2026-10-29. **Severity**: medium. **Status**: open.
- **vukem version**: v0.32.0 (`contrib/stripe` v0.4.1). **Docs consulted**: `https://vukem.dev/docs/cli/reference-install` ("`vukem install:billing`"), `https://vukem.dev/docs/digging-deeper/billing` ("Configuration", "Testing").
- **Context**: the scaffold writes plans `basic` and `pro` in `config/billing.go` (`plans()`), one `STRIPE_PRICE_<NAME>` id per plan and a `Checkout(ctx, user, plan)` without a currency. HexaScore sells `pro` and `business` in BRL, USD and EUR (SDD 05 §6.3, ADR 0015) using one Stripe price per plan with `currency_options`.
- **Workaround** (spike `agent/senior/P0.F.8.1`): rename the plans in `plans()` and in the tests; add `Plan.Amounts` (display only) and a `config.Currency` allowlist; pass `"currency"` in the `stripe.Params` of `CreateCheckoutSession` (`Params` is a free-form map, so no dependency change); the controller reads the `currency` form field and an unsupported value returns 422 before any Stripe call. Whether Stripe accepts `currency` with a `currency_options` price in a live sandbox is [NOT VERIFIED] here (only the fake-Stripe tests ran; test keys only).
- **Request**: document the `currency_options` pattern, accept a currency (or locale to currency map) in the scaffold's `Checkout`, and generate plans from config rather than code.
- Also: `POST /stripe/webhook` is mounted with `r.UseServer` and is not listed by `vukem route:list` (documented as such); the acceptance check for it is the webhook tests.

## VF-038: Starter Vite plugin breaks `vite build --ssr`; SSR entry pattern undocumented; `ssr:start` exits silently

- **Date**: 2026-10-29. **Severity**: medium. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/frontend/inertia` "Server-side rendering"; `https://vukem.dev/docs/cli/reference-runtime` "`vukem ssr:start` and `vukem ssr:check`"; `https://vukem.dev/docs/frontend/vite`.
- **Context**: ROADMAP P0.F.2.1, SDD 09 §6.1, ADR 0004.
- **Reproduction**: in an app from `vukem new -inertia react`, run `vite build --ssr resources/js/ssr.tsx`. The starter plugin `resources/vite/vukem.js` sets `build.outDir = resources/public/build` and `emptyOutDir: true` for every build, so the server bundle becomes publicly served and the client build is deleted. With an `export default createInertiaApp(...)` entry, `@inertiajs/vite` 3.9.1 does not inject `createServer`; `vukem ssr:start` then runs a bundle that exits at once and itself exits 0 with no output.
- **Workaround**: plugin returns `{ build: { outDir: "bootstrap/ssr", emptyOutDir: true } }` when `isSsrBuild`; SSR entry is a top-level `createInertiaApp({...})` statement; `build:ssr` script added by hand.
- **Request**: ship the `isSsrBuild` branch, a `resources/js/ssr.tsx` and a `build:ssr` script in the Inertia starters (or have `vukem build` produce the SSR bundle, see VF-027); document the entry shape on the Inertia page; make `ssr:start` log and exit non-zero when the bundle exits without listening.

## VF-039: No signed opaque-token helper and no problem+json extension members

- **Date**: 2026-10-29. **Severity**: low. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/rest-apis/pagination` ("Query conventions", "Envelopes"), `https://vukem.dev/docs/frontend/inertia` (Scroll), `https://vukem.dev/docs/digging-deeper/storage` (signed URLs).
- **Context**: SDD 01 §7.2 assumed `signed.Signer` can sign a cursor payload; it only signs URLs (`Sign(rawURL, ttl)`, `Verify(*http.Request)`), so the spike uses `crypto/hmac` with a key from `App.DeriveKey("catalog-cursor", 32)`. `web.Problem` (RFC 9457) has no extension members, so the SDD 05 problem code `invalid-cursor` needs a hand-written `application/problem+json` body. `orm.CursorPaginate` rejects `sort`, hence the app-owned keyset (VF-005).
- **Workaround**: `app/catalog/cursor.go` (HMAC-SHA256, constant-time compare, sort key and direction inside the signed payload) and `invalidCursor` in `app/controllers/items.go`.
- **Request**: `signed.Token` (sign and verify arbitrary payloads with key rotation), `Problem.Extensions`/`Code`, and a documented keyset cursor over `(sort, id)` in the ORM.

## VF-040: Migration runner has no lock hook and no non-transactional migrations (custom command needed)

- **Date**: 2026-10-29. **Severity**: medium. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/database/migrations` ("Managing them with migrate"), `https://vukem.dev/docs/deployment/docker-and-railway`, `https://vukem.dev/docs/cli/reference-runtime`.
- **Context**: confirms VF-026 and VF-017 on a real run. `app.Migrator().Up` takes no lock, so two `migrate` processes can race; `schema.Migration` has no hook for a lock and no non-transactional mode.
- **Workaround** (spike `agent/senior/P0.F.5.2`): custom command `migrate:locked` (`vukem.CommandFunc`) holding `pg_advisory_lock(hashtext('hexascore:migrate'))` on a dedicated `*sql.Conn` from `db.SQL()` around `Migrator().Up`, duplicating the output of `migrate` (`Migrated:` lines, `Nothing to migrate.`); and `scripts/concurrent-index.sh` running `database/concurrent/<name>.sql` through psql outside a transaction after dropping an INVALID leftover. Concurrent indexes are not recorded by `migrate:status`.
- **Request**: lock inside `Migrator.Up` (Postgres advisory, MySQL `GET_LOCK`, documented), `schema.Migration{NoTransaction: true}` recorded like the others, and an exported helper for the `migrate` command output so wrappers do not copy it.

## VF-041: Custom view functions are not request-bound; the page URL is the stripped URL under a path-prefix rewrite

- **Date**: 2026-10-29. **Severity**: low. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/frontend/inertia` (root view), `https://vukem.dev/docs/frontend/inertia-i18n`. The two behaviours below were confirmed in the v0.32.0 source (`view/funcs.go`, `inertia/inertia.go`), not found in the docs: `[NOT IN VUKEM DOCS]`.
- **Context**: refines VF-024 and VF-028 after proving them under SSR. (1) `WithViewFuncs` functions take no `context.Context`, so a view function cannot read a cookie (VF-028 workaround as written is impossible). (2) Inertia builds `Page.URL` from `r.URL.RequestURI()`, so a path-prefix middleware that strips `/pt-br` before routing (VF-024 workaround) makes every page object report `/` and breaks client history.
- **Workaround** (spike `agent/senior/P0.F.2.3`): the theme cookie is validated and shared as an Inertia prop `theme`, read by the root view as `.Page.Props.theme`; a route-level `RestoreURL` middleware puts the original prefixed URL back on the request after routing, before the Inertia middleware.
- **Request**: `inertia.Config.RootData func(*http.Request) map[string]any` (as in VF-028), and a first-class path-prefix locale source (VF-024) that keeps the original URL for `Page.URL` and `URLFor`.

## VF-042: Problem extension members, group-level error handling and bodyless 202 (confirms VF-014, VF-016)

- **Date**: 2026-10-29. **Severity**: low. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/rest-apis/openapi`, `https://vukem.dev/docs/cli/reference-runtime` ("`vukem openapi:export`"), `https://vukem.dev/docs/rest-apis/overview` ("Errors"). The three behaviours below were confirmed by running them and reading the v0.32.0 source (`web/errors.go`, `auth/middleware.go`, `api/typed.go`): `[NOT IN VUKEM DOCS]`.
- **Context**: (1) `web.Problem` has fixed members (`request_id` snake_case, no `code`, no extensions) and the OpenAPI `Problem` schema is derived from it, so a `code` member is invisible to generated clients (VF-014, VF-016 confirmed). (2) Errors returned by handlers reach only the router-wide `Services.Errors`; `auth.ProblemJSON`-style group middleware that swaps services does not see them, so per-group problem formatting is impossible. (3) `api.Op{Status: 202}` with an `Out` of `struct{}` writes the body `{}` while the OpenAPI document says "no body", so `AssertMatchesOpenAPI` fails.
- **Workaround** (spike `agent/senior/P0.F.7.1`): `problems.Install(router, prefixes...)` wraps `Services.Errors`, runs vukem's default handler into a buffer and adds `code` plus extension members to its JSON; `problems.Error` carries the code. The 202 route returns a one-field `Accepted` struct.
- **Request**: `errs.HTTPError` code and extension members rendered by `DefaultErrorHandler` and declared in the OpenAPI `Problem` schema; a documented per-group error handler; `Out = struct{}` with an explicit 2xx status writes no body.

## VF-043: Static assets and SSR HTML are not compressed unless the app adds `middleware.Gzip()`

- **Date**: 2026-10-29. **Severity**: low. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/basics/middleware`, `https://vukem.dev/docs/frontend/vite` (production build). `[NOT IN VUKEM DOCS]`: no page says that the starter serves `/build/assets/*` and Inertia HTML uncompressed (confirmed by Lighthouse: script transfer 568 kB uncompressed vs 116 kB gzip).
- **Context**: the performance budgets (NFR-09-5, LCP) are measured on transfer size. The scaffold ships no compression, so a container behind a proxy that does not compress fails the JS budget (354 kB raw `app-*.js`).
- **Workaround** (spike `agent/senior/P0.F.10.1`): `r.UseServer(middleware.Gzip())` before `StaticFS` in `routes.Web`. Brotli is not offered by `middleware.Gzip` (`HEAD` and `Range` are passed through).
- **Request**: starter template enables gzip (or documents it as a required deploy step), and a Brotli option for static assets.

## VF-044: Generated types mark every response property optional; untyped routes appear as `*/*` in the document

- **Date**: 2026-10-29. **Severity**: low. **Status**: open.
- **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/rest-apis/client-generation`, `https://vukem.dev/docs/cli/reference-runtime` ("`vukem codegen:client`", "`vukem codegen:routes`"). `[NOT IN VUKEM DOCS]`: nothing on `required` in output schemas or on plain handlers.
- **Context**: (1) the OpenAPI schemas generated from `Out` types have no `required` list, so `codegen:client` emits `Item = { id?: string; slug?: string; ... }` and the SDKs force null checks on fields the server always sends (inputs keep `required` from `validate` tags). (2) A route registered with a plain handler (`g.Get("/ping", ...)`) is documented with a `*/*` response and an empty schema; `openapi-python-client` 0.29.1 warns and omits the response.
- **Workaround**: register every public route through `api.Register` with typed output; accept optional properties in the first-party client.
- **Request**: mark non-pointer, non-`omitempty` output fields `required` in the generated schemas (or a `doc:"required"` tag), and document or reject plain handlers inside OpenAPI prefixes.

## VF-045: `storage.Handler` sets no Cache-Control or ETag, and S3 objects get no Last-Modified

- **Date**: 2026-10-29. **Severity**: low. **Status**: open.
- **vukem version**: v0.32.0 (`drivers/s3` v0.3.1). **Docs consulted**: `https://vukem.dev/docs/digging-deeper/storage` (Serving files; the page documents `X-Content-Type-Options`, the sandbox CSP and `Last-Modified` for seekable disks only), `https://vukem.dev/docs/digging-deeper/exports` (VF-029 confirmed: `export.CSV` needs an `http.ResponseWriter`; `export.SafeCell` is exported and is the smaller workaround).
- **Context**: covers are content-addressed (`covers/<id>-<hash16>.webp`) and must answer `Cache-Control: public, max-age=31536000, immutable`. `storage.Handler` over an S3 disk sends no `Cache-Control`, no `ETag` and (the S3 reader is not an `io.ReadSeeker`) no `Last-Modified`, so conditional requests never get a 304 [NOT IN VUKEM DOCS].
- **Workaround**: `app/covers/serve.go` validates the key with a regexp, delegates to `storage.Handler` and adds `Cache-Control` on 200 responses only (spike branch `agent/senior/P0.F.9.1`).
- **Request**: `storage.HandlerWith(d, signer, storage.Options{CacheControl: ...})` and `Last-Modified`/`ETag` from `s3.Disk.Stat` for non-seekable disks.
