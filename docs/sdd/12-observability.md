# SDD 12: Observability and operations

Status: reviewed (2026-10-28 consistency pass). Decisions: D-07, D-08, D-18, D-21. ADRs: 0002, 0004, 0007, 0014. Related SDDs: 02, 04, 05, 06, 07, 08, 10, 11, 13.

## 1. Objective

One owner runs HexaScore on a single Railway container (D-08). This SDD makes sure they hear about SSR fallbacks (D-07), failed jobs, missed backups (D-18), quota drift and cost overruns before users or crawlers notice. The observability spend is US$0, inside the US$30 ceiling.

Legacy baseline:

- "logger Telegram para alertas. Sem Sentry por ora" (`legacy/PLAN.md` §2).
- Telegram variables declared but never wired (`legacy/apps/web/lib/env.ts` lines 48–56).
- `legacy/apps/web/instrumentation.ts` only reports disabled features.
- 4xx/5xx alerts "ainda manuais" (`legacy/docs/security.md` A09).

## 2. Scope and sources

In scope (v1):

- logs, metrics, traces and error aggregation;
- alerts, dashboards and health checks;
- runbooks, log privacy and cost monitoring;
- the mobile JS crash intake, which closes SDD 10 open issue 2.

Out of scope:

- a paid APM or hosted error tracker (each needs an ADR);
- Prometheus, Grafana (self-hosted or cloud) and any metrics pipeline outside Railway (D-35, §7.3);
- an automated status page;
- log shipping, since Railway "does not have a log drain setting" [R-logs].

Source keys, all accessed 2026-10-28. vukem facts cite <https://vukem.dev/docs>. A fact found only in `C:\dev\hexascore-app\vukem\docs\` is tagged [SITE PAGE MISSING].

| Key | URL |
| --- | --- |
| V-otel | <https://vukem.dev/docs/digging-deeper/opentelemetry> |
| V-tg | <https://vukem.dev/docs/digging-deeper/telegram-logging> |
| V-mw | <https://vukem.dev/docs/basics/middleware> ("Built-in middleware") |
| V-conf | <https://vukem.dev/docs/configuration> ("Core settings", "APP_KEY") |
| V-sec | <https://vukem.dev/docs/security/overview> |
| V-enc | <https://vukem.dev/docs/security/encryption> ("Key rotation", "After a leak") |
| V-db | <https://vukem.dev/docs/database/overview> |
| V-queue | <https://vukem.dev/docs/digging-deeper/queue> |
| V-sched | <https://vukem.dev/docs/digging-deeper/scheduler> |
| V-storage | <https://vukem.dev/docs/digging-deeper/storage> |
| V-dev | <https://vukem.dev/docs/digging-deeper/devtools> ("How it collects") |
| V-route | <https://vukem.dev/docs/basics/routing> ("Mounting plain handlers") |
| V-rt | <https://vukem.dev/docs/cli/reference-runtime> |
| V-inertia | <https://vukem.dev/docs/frontend/inertia> ("Server-side rendering") |
| V-rail | <https://vukem.dev/docs/deployment/docker-and-railway> |
| L-sub | `vukem/docs/subsystems.md` "Logging & observability" [SITE PAGE MISSING] |
| L-sec | `vukem/docs/security.md` A09/A10 [SITE PAGE MISSING] |
| R-logs | <https://docs.railway.com/observability/logs> |
| R-trace | <https://docs.railway.com/observability/tracing> |
| R-alerts | <https://docs.railway.com/guides/alerts-crashes-failed-deploys> |
| R-cost | <https://docs.railway.com/pricing/cost-control> |
| R-roll | <https://docs.railway.com/guides/roll-back-bad-deploy> |
| R-inc | <https://docs.railway.com/guides/debug-production-incident> |

## 3. Functional requirements

- FR-12-1: Each request writes one JSON line to stdout. Fields: `request_id`, `trace_id`, `route`, `method`, `status`, `duration_ms`, and `user_hash` when authenticated. No email ever appears. Verify: `TestLog_RequestLineShape`.
- FR-12-2: Traces cover every route group (HTTP, SQL, queue, outbound; §6.3). Verify: `vukemtest` request with an in-memory span exporter, one span per group.
- FR-12-3: The SSR `OnFailure` hook logs the fallback with `http.route` and `public`. On public routes it also logs at `telegram.LevelFatal` [V-inertia; VF-002]. Verify: integration test that stops Node.
- FR-12-4: Each app-sourced alert in §7.5 has an in-app check that logs `alert=<code>` and a fixture test that fires and clears it. Verify: the tests of §10.
- FR-12-5: `/up?checks=1` reports `db`, `ssr`, `storage` and `scheduler`. Verify: `TestReady_ChecksPresent`.
- FR-12-6: A digest of error fingerprints is sent to Telegram daily at 08:00 America/Sao_Paulo. Verify: fake-clock test.
- FR-12-7: `POST /app/v1/client-errors` (route in SDD 05 §6.4) takes opt-in mobile crash reports (SDD 10) with no PII. Each is logged as `client.error`, is rate-limited to 10/min per IP, and is kept 7 days (open issue 10-2). Verify: integration test.
- FR-12-9: `POST /app/v1/client-metrics` (SDD 05 §6.4) takes the opt-in, anonymous web-vitals beacon from SDD 09 (LCP, INP, CLS, TTFB per route pattern and locale) into `client_reports`; raw rows kept 7 days (open issue 09-11). Verify: integration test.
- FR-12-10: A "projected spend at day 20" alert (A17) fires when month-to-date Railway usage extrapolated to the full month exceeds US$27, so the owner can raise the US$30 hard limit before it stops the service (D-27). Verify: fixture test on the projection.
- FR-12-8: Logs never contain passwords, tokens, API keys, emails or raw search text. Verify: `TestSecurity_A09_HexaLogsRedacted`.

## 4. Non-functional requirements

- NFR-12-1 Cost: no spend beyond the Railway plan. Telegram is free.
- NFR-12-2 Cardinality: no metric series exist. Log fields and span attributes never carry ids, slugs, user data or raw paths.
- NFR-12-3 Overhead: telemetry adds at most 3 % CPU and 40 MB RSS. Verify: k6 with tracing on and off.
- NFR-12-4 Log volume: at most 50 lines/s at 50 rps. Railway drops logs above 500 lines/s per replica [R-logs].
- NFR-12-5 Retention:
  - Railway keeps logs 7 days on Hobby and 30 on Pro [R-logs].
  - The LGPD maximum for logs that carry IPs is 30 days (NFR-03-7; L-sec says IP retention is the app's responsibility).
  - `error_digest_daily` keeps 90 days.
- NFR-12-6 Detection: app down within 2 min (Railway healthcheck); SSR fallback within 5 min.

## 5. Data model

```sql
create table client_reports (                    -- client-errors and client-metrics intake; pruned at 7 days
  id bigint generated always as identity primary key,
  kind text not null check (kind in ('error','web_vital')),
  platform text not null check (platform in ('android','web')),
  app_version text, route text,                  -- route pattern or screen name, never a full URL
  name text, value real,                         -- web vitals only
  message text check (length(message) <= 1000), stack text check (length(stack) <= 8192),
  received_at timestamptz not null default now());  -- no user id, token id, IP or e-mail
create index client_reports_received_idx on client_reports (received_at);

create table error_digest_daily (               -- 60 s in-memory flush (VF-021 pattern)
  day date not null, fingerprint text not null,  -- sha256(route|error.type|event)[:16]
  route text not null, error_type text not null, count integer not null,
  first_request_id text, last_seen timestamptz not null,
  primary key (day, fingerprint));               -- pruned at 90 days, daily 04:50 UTC
```

Gauges read from tables owned by other SDDs:

- `backup_runs` (SDD 11)
- `edit_suggestions` (SDD 04)
- `ingest_runs` and `failed_jobs` (SDD 06)
- `api_quota_counters` (SDD 05)

## 6. Contracts

### 6.1 Logs

- Fields from vukem:
  - `Logger` writes `method`, `path`, `status`, `duration`, `bytes`, `request_id`, `peer` and `ip`; `RequestID` echoes `X-Request-ID` [V-mw].
  - Route name and user id are documented only in [L-sub].
  - `otel.LogHandler` adds `trace_id` and `span_id` [V-otel].
- Format and level: `LOG_FORMAT=json` in production [V-conf]. `LOG_LEVEL` is `info` in production and `debug` elsewhere.
- Handler chain: dedup, then JSON to stdout plus `telegram.NewHandler`, passed with `WithLogger`. The "used as given" rule and `logging.NewDedup` are documented only in [L-sub]. The site documents a 1m `LOG_DEDUP_WINDOW` [V-conf].
- Output goes to stdout because Railway maps stderr to `level.error` [R-logs].
- `user_hash` = hex(HMAC-SHA256(`app.DeriveKey("log", 32)`, user_id))[:16] [V-conf]. Two changes are [NOT IN VUKEM DOCS] (issue 2): hashing vukem's own user-id attribute and truncating `ip` to /24 or /48.
- Implementation (P2.C.1.1, VF-022 workaround): `app/logging` wraps `app.Log` before `routes.Web` installs vukem's default stack, so vukem's `Logger` writes through it while `LOG_FORMAT`, dedup and `otel.LogHandler` stay vukem's. On the `request` line it truncates `ip` and `peer`, replaces `duration` with `duration_ms`, and adds `route` (the matched `ServeMux` pattern, `unmatched` otherwise; router middleware `RequestRoute`) and `user_hash` (recorded by `NotBanned`, which resolves the user on every authenticated group) through a per-request holder installed by the server middleware `RequestLog`. `scripts/ops/check-logs.sh` checks a log stream.
- Sampling: 2xx/3xx on `/up`, `/build/*`, `/search/suggest` and `/search/events` is kept at 1 in 100. Every 4xx, every 5xx and anything over 1 s is always kept. The sampling handler is app code [NOT IN VUKEM DOCS].
- Redaction:
  - `Authorization` and `Cookie` are never logged, and security events go out at WARN with stable names [V-sec].
  - Password and token redaction is documented only in [L-sec] (`TestSecurity_A09_LoggerNeverLogsCredentials`).
  - App layer (`app/logging`, `TestSecurity_A09_HexaLogsRedacted`): values of keys such as `password`, `token`, `authorization`, `cookie`, `api_key`, `email`, `q`, `query` become `[REDACTED]`; emails, `Bearer`/`Basic` credentials, `hx_live_`/`hx_app_`/Stripe/Resend keys and sensitive query parameters are scrubbed from messages, strings and errors.

### 6.2 Signals

D-35 (amended 2026-10-31): observability uses only Railway's native tools. Railway's OTLP receiver accepts traces only, so the app exports no metrics: there is no `/metrics` endpoint, no OTel Prometheus exporter and no `hexa.*` meter instruments (removed in FIX-observability-railway-native; `OTEL_METRICS_EXPORTER=none` stays set). Adding any of them needs a new decision first.

The signals are:

| Signal | Where | Used for |
| --- | --- | --- |
| Service metrics (CPU, RAM, network, disk, volume) | Railway dashboard and monitors [R-alerts] | capacity, A11, A15 |
| JSON log lines (§6.1) | Railway log search | debugging, per-request facts, ingestion and job outcomes |
| Traces | Railway OTLP receiver (§6.3) | latency of a request, SQL, queue, outbound calls |
| `alert=` log lines | Railway log search; ERROR also reaches Telegram | the alerts of §7.5 that need app knowledge |

An in-app check is a job or middleware that evaluates its condition and writes one line with a stable key: `level=WARN|ERROR alert=A14 msg="trending snapshot stale" age_s=...`. Keys are the alert codes of §7.5, plus `trending_counter_dropped` (WARN, informational: the trending key counter hit its cap and dropped keys; no paging rule) and `community_asserts` (ERROR: a SDD 04 §10 nightly assert failed; the message carries counts and check names only); the line carries numbers, never ids, emails or raw paths. A check that cannot run logs `alert=<code> state=unknown` at ERROR (fail closed).

Business and queue figures (suggestion queue age, ratings per day, active keys, DB size) are SQL reads in the admin screens or in the check that needs them; they are not exported.

### 6.3 Traces

- `otel.New` is listed first in `WithProviders`. It traces:
  - HTTP;
  - SQL, without query text;
  - queue dispatch and attempts;
  - mail;
  - outbound calls (sources and the SSR POST) through `httpclient.Options.WrapTransport`.

  Spans never carry paths or query strings [V-otel "Outbound HTTP calls", "Security"].
- Sampler: `parentbased_traceidratio`. `OTEL_TRACES_SAMPLER_ARG` is 0.1 in production and 1.0 in staging.
- Destination: Railway's OTLP receiver. It accepts "traces only", keeps them for the log retention period, and has a limit of 1,000 spans per 10 s per replica. It needs `OTEL_METRICS_EXPORTER=none` [R-trace; V-otel].
- Correlation: responses carry `X-Request-ID` (vukem) and `X-Trace-ID` (app middleware). Making request id equal trace id is [NOT IN VUKEM DOCS] (issue 2).

### 6.4 Routes

| Method | Path | Auth | Notes |
| --- | --- | --- | --- |
| GET | `/up` | none | liveness; `vukem.health` [V-route] |
| GET | `/up?checks=1` | none | readiness; returns `ok`/`fail` per check, never error text [L-sec] |
| POST | `/app/v1/client-errors` | none (opt-in in the app) | body `{appVersion, versionCode, androidApi, screen, message, stack≤8KB}` → 204; over 16 KB → 413; 10/min per IP; no PII; 7 days |
| POST | `/app/v1/client-metrics` | none (opt-in cookie banner) | body `{name, value, page, locale}` → 204; over 4 KB → 413; 60/min per IP; no PII; 7 days |

## 7. Behaviour and algorithms

### 7.1 Health checks

Checks are registered with `app.AddHealthCheckTimeout` [L-sub], each with a 2 s timeout:

- `db`: added by `app.DB()`, "reported by `/up?checks=1`" [V-db].
- `ssr`: `GET 127.0.0.1:13714/health`, the same probe as `vukem ssr:check` [V-rt].
- `storage`: `storage.Check` (write, read back, delete) [V-storage], the vukem-recommended probe; an existence `HEAD` passes with credentials that cannot write. Local exports disk until the bucket of P2.A.2.1 exists.
- `scheduler`: heartbeat age under 180 s. The heartbeat is written by `s.Func("heartbeat").EveryMinute()` [V-sched], kept in memory and seeded at boot; the check exists only with `SCHEDULE_IN_SERVE=true`.

Railway polls the Healthcheck Path only on deploys and switches traffic after a 200 [R-roll]. vukem's guide uses `/up` from host `healthcheck.railway.app` [V-rail]. HexaScore uses `/up?checks=1`, so a deploy with broken SSR or DB never takes traffic. If query strings are rejected [NOT VERIFIED], mount `app.ReadyHandler()` at `/ready` [L-sub]. Runtime health comes from §7.4.

### 7.2 Errors

- No Sentry (legacy parity). A hosted tracker or native crash capture needs an ADR.
- vukem errors already arrive at ERROR level, so Telegram sees them:
  - `scheduled task failed` [V-sched];
  - `job failed` after the last retry [V-queue "Retries, backoff and timeouts"].

  Dedup collapses repeats.
- The app's error handler:
  - fingerprints each error from (`http.route`, `error.type`, `event`);
  - logs the fingerprint with the request id;
  - buffers upserts into `error_digest_daily`.
- The 08:00 digest posts the top 20 fingerprints with one request id each, via `app.Telegram.SendSync` [V-tg].

### 7.3 Where data goes

| Option | Cost | Verdict |
| --- | --- | --- |
| Railway metrics, logs, traces and alerts | included; monitors need the Pro plan [R-alerts] | **Chosen** (D-35) |
| Hosted metrics or alerting outside Railway (Grafana Cloud and similar) | free tiers, extra processor | **Rejected** by the owner (D-35, 2026-10-31) |
| Prometheus + Grafana inside the app container | CPU, RAM and a TSDB volume inside the US$30 container | **Rejected**: competes with HTTP/SSR (D-08) and dies with what it watches |

### 7.4 Liveness

Railway's healthcheck on `/up` (ADR 0002) restarts a failing container and emails the deploy failure [R-alerts]. `/up?checks=1` is read by hand and by the post-deploy smoke test. There is no external synthetic probe; certificate expiry is Railway's concern.

### 7.5 Alerts

Channels:

- Railway monitors and usage alerts notify by email and webhook [R-alerts].
- App checks write `alert=` log lines (§6.2). ERROR lines go through the vukem Telegram handler [V-tg] with `TELEGRAM_MIN_LEVEL=error` and `TELEGRAM_RATE_LIMIT=20`. `ChatIDs` sends `LevelFatal` to the page chat; the error chat is `Silent: true`.
- A saved Railway log search per `alert=` code is the way to review history.

On-call is the owner. Quiet hours run 23:00-08:00 America/Sao_Paulo: only `page` rings. Ack targets: page within 15 min while awake, warn the same day.

| # | Alert | Condition | Window | Sev | Source |
| --- | --- | --- | --- | --- | --- |
| A1 | App down | `/up` healthcheck fails and Railway restarts the service | 2 min | page | Railway healthcheck and deploy email |
| A2 | Readiness | `/up?checks=1` is not 200 after a deploy | per deploy | page | post-deploy smoke test |
| A3 | 5xx rate | > 1 % with at least 20 requests, from a rolling in-process counter in the request middleware | 5 min | page | app check |
| A4 | Latency | `request` lines over 1 s in more than 5 % of requests per route group | 10 min | warn | app check (traces for detail) |
| A5 | SSR fallback | the `OnFailure` hook fires on a public route (already logs at `telegram.LevelFatal`, FR-12-3) | immediate | page | app |
| A6 | Job failed twice | same kind fails 2 runs in a row (`queue.JobFailed` via `q.AddObserver` [V-queue]) | - | warn | app |
| A7 | Ingestion | SDD 06 §10 list: failed run; no incremental in 26 h; `error` assert; TMDB data older than 170 d; breaker open over 1 h | SDD 06 | warn | app check |
| A8 | Backup missing | newest backup older than 26 h, read from the backup bucket or `job_runs` | 15 min | page (not during quiet hours) | app check |
| A9 | Moderation | oldest open suggestion > 20 h (NFR-04-2) | 30 min | warn | app check |
| A10 | Quota flush | 5 consecutive flush errors (SDD 05 §7) | - | page | app |
| A11 | Disk | Postgres volume usage > 80 % | 30 min | warn | Railway monitor |
| A12 | Connections | pool in use / max > 0.8 (8 of 10), read from `db.SQL()` stats | 5 min | page | app check |
| A13 | Score drift | nightly drift check finds a mismatch (SDD 02) | nightly | warn | app check |
| A14 | Trending stale | newest snapshot older than 3 h (SDD 08) | 15 min | warn | app check |
| A15 | Cost | Railway usage >= US$24 (80 % of US$30), "Custom email alert" [R-cost]; hard limit US$30 (D-27) | cycle | warn | Railway |
| A16 | Logs gone | no `request` line for 10 min while the service is up | 10 min | warn | app check |
| A17 | Projected spend | on day 20 of the billing cycle (and daily after), month-to-date usage x days in cycle / elapsed days > US$27 (D-27: hard limit US$30) | daily | warn | GitHub Actions cron reading Railway usage [R-cost] (programmatic access [NOT VERIFIED]; fallback: calendar reminder to the owner on day 20) |

Checked 2026-10-30: no documented Railway usage API was found in the repo, so `projected-spend.yml` takes the fallback. It logs a day-20 reminder and, when `USAGE_USD` is set, evaluates the formula. It posts to Telegram only when that step's secret is present and the formula alerts or the reminder is due.

### 7.6 Views

No external dashboards. Service health is the Railway service-metrics page; the rest are saved Railway log searches per `alert=` code and the admin screens (moderation queue, API customers, ingestion runs from `job_runs`).

### 7.7 Status page

A static page in `ops/status/`, edited by hand during incidents, published on GitHub Pages from a `status` repository so it survives an outage of the app (open issue 12-5); `status.hexascore.app` is a Cloudflare CNAME to it (D-28). It states the 24 h RPO (ADR 0014).

### 7.8 Runbooks

R1 SSR crash loop:

1. Logs: `@level:error AND ssr`.
2. `vukem ssr:check` [V-rt].
3. Bad bundle: R7.
4. Node out of memory: `inertia.DisableSSR` on logged-in pages, then redeploy.
5. Fallbacks flat for 15 min.

R2 Postgres connections exhausted:

1. Group `pg_stat_activity` by `state`.
2. Terminate sessions idle in a transaction for over 60 s.
3. Pause ingestion (SDD 06).
4. Restart or lower `DB_MAX_OPEN`.
5. File a bug with the trace ids.

R3 Ingestion stuck:

1. Find the `ingest_runs` row `running` past its timeout.
2. Check the breaker and 429 panels.
3. Mark it `failed`.
4. Re-dispatch; it resumes by cursor (ADR 0007).
5. Disable the source if it is down.

R4 Quota drift:

1. Recount from `api_usage_daily` and compare with `api_quota_counters`.
2. Correct in the customer's favour (NFR-05-2).
3. Check flush errors.

R5 Suggestion backlog:

1. Open `/admin/moderation/metrics` and sort by age.
2. Bulk-reject spam (SDD 04).
3. Still over 24 h after 3 days: recruit moderators.

R6 Disk full:

1. Check `pg_total_relation_size`.
2. Prune `search_events`, `rating_events` and `error_digest_daily`.
3. `search:reindex` (SDD 07).
4. Still over 80 %: budget decision (§7.9).

R7 Rollback:

1. Use the Railway rollback on the previous good deployment. It restores the image and variables within image retention [R-inc].
2. Expand/contract migrations keep the old code valid (ADR 0014).
3. Check `/up?checks=1`.
4. Update the status page.

R8 Restore:

1. `vukem down -retry 60s`; `/up` keeps answering [V-rt].
2. Run the SDD 11 restore (RTO 1 h).
3. `catalog:check`.
4. `vukem up`.

R9 Leaked secret:

1. Revoke it at the provider.
2. Set the new value in Railway and redeploy.
3. `APP_KEY`: follow [V-enc]; never keep a leaked key in `APP_PREVIOUS_KEYS`.
4. Search the logs since the leak.
5. Record the incident (SDD 13).

### 7.9 Cost monitoring

Monthly checklist:

1. `railway usage limit status` [R-cost].
2. Invoice vs US$30.
3. Log volume and trace span count against Railway limits.
4. Postgres size vs plan.
5. Bucket size (≈ 2.1 GB per 100k items, NFR-06-7).
6. Egress; cover egress > US$5/month triggers the CDN decision (SDD 11 open issue 2).
7. Client intake volume (`client_reports` rows per day) and that the 7-day prune ran.

Budget-increase predictors (the first step is splitting out the worker, D-08):

- CPU > 70 % for 7 days;
- Postgres > 60 % of plan;
- night job backlog > 1 h.

## 8. Edge cases

- OTLP receiver down: traces are dropped, serving continues, logs are unaffected.
- Telegram down: stdout still has the `alert=` lines in Railway logs.
- SSR fallback on a logged-in route: logged with `public=false`, no page.
- Container down: in-app checks are silent; A1 comes from Railway's healthcheck.

## 9. Acceptance criteria

- `go test ./... -run 'TestLog_|TestReady_|TestSecurity_A09_Hexa'` passes.
- `curl -s https://staging.hexascore.app/up?checks=1` returns 200 with all four checks `ok`.
- After `pkill node`, the same call returns 503 with `ssr: fail`, and a Telegram page arrives within 5 min.
- `curl -sI https://staging.hexascore.app/` shows `X-Request-ID` and `X-Trace-ID`, and the trace id opens in Railway's trace view.
- `railway logs` lines parse as JSON with `level`, `request_id` and `route`, and a scripted grep finds no email pattern.
- Each in-app check writes its `alert=<code>` line on a fixture that breaches it, and nothing when it does not.
- The Railway usage page shows the US$24 email alert.

## 10. Test strategy

- Unit: fingerprinting, sampling (5xx always kept), the rolling 5xx counter.
- Integration (Postgres): in-app checks on fixtures (each fires and clears); digest flush and prune.
- `TestSecurity_A09_HexaLogsRedacted`: a request carrying a password, a bearer token, an API key, an email and a search query leaves none of them in the captured JSON.
- Monthly A1 drill: stop the service for 3 min and confirm Railway's restart and email.
- Load: k6 for NFR-12-3 (tracing on and off).

## 11. Open issues

1. VF-019: there is no app meter accessor. Moot since D-35 (amended 2026-10-31): the app has no meter instruments.
2. Request-logger policy: hash the user id, truncate the IP, make the request id equal the trace id. All three are [NOT IN VUKEM DOCS]. Owner: vukem (new VF entry). Closes when it is filed and resolved.
3. `/metrics`. Decided (D-35, amended 2026-10-31): no endpoint and no app metrics (§6.2).
4. The Railway plan (Hobby or Pro) sets log retention (7 or 30 days) and whether monitors exist. Owner: architect (SDD 11). Closes when SDD 11 records the plan.
5. Status page hosting. Decided (12-5): GitHub Pages (§7.7).
6. SDD 08 metric names. Decided (12-6): §6.2 names are canonical; SDD 08 aligned.
7. Metric Views to drop high-cardinality attributes are [NOT IN VUKEM DOCS]. Owner: vukem. Trigger: staging series above 6,000.
8. Railway monitor options (metrics offered, webhook delivery, plan) are [NOT VERIFIED] until the Pro plan is on. Owner: architect. Closes in the observability roadmap task.
9. Railway hard usage limit. Decided (D-27): US$30, e-mail alert US$24, plus the day-20 projection alert A17.
10. Site documentation gaps, all [SITE PAGE MISSING]. Owner: vukem. Closes when a logging/health page is published.
    - the health check API (`AddHealthCheckTimeout`, `ReadyHandler`, `ok`/`fail` output);
    - route name and user id in request logs;
    - `logging.NewDedup` and `WithLogger` semantics;
    - the A09/A10 test names.

    Separately, there is no documented metric hook for scheduled tasks, only the ERROR log; the `alert=` checks of §6.2 do not depend on one.
