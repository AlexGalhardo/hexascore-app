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
- self-hosted Prometheus/Grafana (§7.3);
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
| G-price | <https://grafana.com/pricing/> |

## 3. Functional requirements

- FR-12-1: Each request writes one JSON line to stdout. Fields: `request_id`, `trace_id`, `route`, `method`, `status`, `duration_ms`, and `user_hash` when authenticated. No email ever appears. Verify: `TestLog_RequestLineShape`.
- FR-12-2: Every route group (§6.2) emits `http.server.request.duration`. Verify: `TestMetrics_EveryRouteGroupEmitsLatency`.
- FR-12-3: The SSR `OnFailure` hook increments `hexa.ssr.fallbacks{http.route,public}`. On public routes it also logs at `telegram.LevelFatal` [V-inertia; VF-002]. Verify: integration test that stops Node.
- FR-12-4: Each alert in §7.5 has a rule plus a fixture test that fires it. Verify: `promtool test rules`.
- FR-12-5: `/up?checks=1` reports `db`, `ssr`, `storage` and `scheduler`. Verify: `TestReady_ChecksPresent`.
- FR-12-6: A digest of error fingerprints is sent to Telegram daily at 08:00 America/Sao_Paulo. Verify: fake-clock test.
- FR-12-7: `POST /app/v1/client-errors` (route in SDD 05 §6.4) takes opt-in mobile crash reports (SDD 10) with no PII. Each is logged as `client.error`, increments `client.errors`, is rate-limited to 10/min per IP, and is kept 7 days (open issue 10-2). Verify: integration test.
- FR-12-9: `POST /app/v1/client-metrics` (SDD 05 §6.4) takes the opt-in, anonymous web-vitals beacon from SDD 09 (LCP, INP, CLS, TTFB per route pattern and locale) into `client.web_vitals`; raw rows kept 7 days (open issue 09-11). Verify: integration test.
- FR-12-10: A "projected spend at day 20" alert (A17) fires when month-to-date Railway usage extrapolated to the full month exceeds US$27, so the owner can raise the US$30 hard limit before it stops the service (D-27). Verify: fixture test on the projection.
- FR-12-8: Logs never contain passwords, tokens, API keys, emails or raw search text. Verify: `TestSecurity_A09_HexaLogsRedacted`.

## 4. Non-functional requirements

- NFR-12-1 Cost: US$0/month. Grafana Cloud Free covers "10k active series", 50 GB of logs and 50 GB of traces, "14 days" retention, and "100k API test executions" [G-price]. Telegram is free.
- NFR-12-2 Cardinality: at most 6,000 active series, which is 60 % of the free tier. Labels never carry ids, slugs, user data or raw paths.
- NFR-12-3 Overhead: telemetry adds at most 3 % CPU and 40 MB RSS. Verify: k6 with tracing on and off.
- NFR-12-4 Log volume: at most 50 lines/s at 50 rps. Railway drops logs above 500 lines/s per replica [R-logs].
- NFR-12-5 Retention:
  - Railway keeps logs 7 days on Hobby and 30 on Pro [R-logs].
  - The LGPD maximum for logs that carry IPs is 30 days (NFR-03-7; L-sec says IP retention is the app's responsibility).
  - Grafana keeps data 14 days.
  - `error_digest_daily` keeps 90 days.
- NFR-12-6 Detection: app down within 2 min; SSR fallback within 5 min.

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
- Sampling: 2xx/3xx on `/up`, `/build/*`, `/search/suggest` and `/search/events` is kept at 1 in 100. Every 4xx, every 5xx and anything over 1 s is always kept. The sampling handler is app code [NOT IN VUKEM DOCS].
- Redaction:
  - `Authorization` and `Cookie` are never logged, and security events go out at WARN with stable names [V-sec].
  - Password and token redaction is documented only in [L-sec] (`TestSecurity_A09_LoggerNeverLogsCredentials`).

### 6.2 Metrics catalogue

Export uses `drivers/otel` OTLP/HTTP push. Built-ins: `http.server.request.duration`, `http.client.request.duration`, `db.client.operation.duration`, `messaging.process.duration` and `vukem.queue.job.failures` [V-otel "Metrics"].

App metrics work around VF-019. `Config.Global` "installs the providers ... as OpenTelemetry globals" [V-otel "Configuration"]. The app then calls `otel.GetMeterProvider().Meter("hexascore")` on the upstream API that `drivers/otel` already pulls in; vukem has no accessor of its own.

Route groups are derived from the `http.route` prefix at query time, with no extra label: `web.public`, `web.app`, `admin`, `api.v1`, `app.v1`, `search`, `health`.

| Name | Type | Labels | Owner |
| --- | --- | --- | --- |
| `http.server.request.duration` (built-in) | histogram | method, `http.route`, status | 12 |
| `hexa.http.inflight` | up-down counter | — | 12 |
| `hexa.ssr.render.duration` | histogram | `http.route` | 09 |
| `hexa.ssr.fallbacks` | counter | `http.route`, `public` | 12 |
| `db.client.operation.duration` (built-in) | histogram | operation, `error.type` | 12 |
| `hexa.db.pool.in_use`, `.max`, `.wait_count` | gauge | — | 12 (`db.SQL()` stats [V-db]; `DB_MAX_OPEN` default 10) |
| `hexa.db.slow_queries` (> 200 ms) | counter | operation | 12 (`database.QueryObserver` [V-dev]) |
| `hexa.pg.connections`, `hexa.pg.database.size_bytes` | gauge (SQL every 60 s) | — | 11 |
| `messaging.process.duration`, `vukem.queue.job.failures` (built-in) | histogram, counter | `vukem.job.kind` | 06 |
| `hexa.schedule.runs`, `hexa.scheduler.heartbeat_age_seconds` | counter, gauge | task, outcome | 12 |
| `ingest.run.duration`, `ingest.records`, `ingest.http.rate_limited`, `ingest.breaker.open`, `ingest.dq.failed`, `ingest.tmdb.max_age_days` | per SDD 06 §10 | job, source, host, code | 06 |
| `score.recomputes` | counter | path | 02 |
| `score.burst.trips`, `score.flags.created` | counter | reason | 02 |
| `score.dq.drift` | gauge | — | 02 |
| `suggest.queue.size`, `.age_p50_seconds`, `.age_p95_seconds`, `.oldest_seconds` | gauge | — | 04 |
| `suggest.decisions`, `suggest.sla.breaches` | counter | decision | 04 |
| `api.requests` | counter | plan, status class | 05 |
| `api.quota.denials`, `api.refunds_5xx`, `api.quota.flush_errors` | counter | plan | 05 |
| `api.webhook.deliveries` | counter | outcome | 05 |
| `search.duration` | histogram | kind | 07 |
| `search.performed`, `search.zero_results` | counter | locale | 07 |
| `trending.compute.duration`, `trending.churn`, `trending.snapshot_age_seconds` | histogram, gauge, gauge | type | 08 |
| `trending.eligible_items`, `trending.prior_slots`, `trending.counter.dropped` | gauge, gauge, counter | type | 08 |
| `similar.compute.duration`, `similar.empty_ratio`, `recommend.compute.duration` | histogram, gauge, histogram | type | 08 |
| `reco.impressions`, `reco.clicks` | counter | surface, type | 08 |
| `backup.last_success_age_seconds`, `backup.dump_size_bytes` | gauge | — | 11 (ADR 0014) |
| `biz.signups`, `biz.verified` | counter | — | 03 |
| `biz.ratings_24h`, `biz.api_keys_active` | gauge (SQL every 5 min) | — | 12 |
| `hexa.errors` | counter | `error.type` (50 values, then `other`) | 12 |
| `client.errors` | counter | `app_version` (10 values) | 10 |
| `client.web_vitals` | histogram | name (LCP/INP/CLS/TTFB), route group, locale | 09 |

These names are canonical: SDD 08 §6 uses them as written (open issue 12-6); no underscore variants exist.

No `/metrics` endpoint at launch (D-35): metrics go to Grafana Cloud over OTLP only. If one is added later it needs the OTel Prometheus exporter (a new dependency, owner approval), would be mounted as `r.Handle("GET /metrics", handler)` [V-route] and would require `Authorization: Bearer $METRICS_TOKEN` (`config.Secret`, constant-time compare, 404 when unset).

### 6.3 Traces

- `otel.New` is listed first in `WithProviders`. It traces:
  - HTTP;
  - SQL, without query text;
  - queue dispatch and attempts;
  - mail;
  - outbound calls (sources and the SSR POST) through `httpclient.Options.WrapTransport`.

  Spans never carry paths or query strings [V-otel "Outbound HTTP calls", "Security"].
- Sampler: `parentbased_traceidratio`. `OTEL_TRACES_SAMPLER_ARG` is 0.1 in production and 1.0 in staging.
- Destination: Grafana Cloud OTLP, the same place as metrics, via `OTEL_EXPORTER_OTLP_ENDPOINT` and `_HEADERS` (a `config.Secret`). The URL format is [NOT VERIFIED].
- Fallback: Railway's receiver. It accepts "traces only", keeps them for the log retention period, and has a limit of 1,000 spans per 10 s per replica. It needs `OTEL_METRICS_EXPORTER=none`, or `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT` pointed elsewhere [R-trace; V-otel].
- Correlation: responses carry `X-Request-ID` (vukem) and `X-Trace-ID` (app middleware). Making request id equal trace id is [NOT IN VUKEM DOCS] (issue 2).

### 6.4 Routes

| Method | Path | Auth | Notes |
| --- | --- | --- | --- |
| GET | `/up` | none | liveness; `vukem.health` [V-route] |
| GET | `/up?checks=1` | none | readiness; returns `ok`/`fail` per check, never error text [L-sec] |
| POST | `/app/v1/client-errors` | none (opt-in in the app) | body `{appVersion, versionCode, androidApi, screen, message, stack≤8KB}` → 202; over 16 KB → 413; 10/min per IP; no PII; 7 days |
| POST | `/app/v1/client-metrics` | none (opt-in cookie banner) | body `{name, value, page, locale}` → 202; over 4 KB → 413; 60/min per IP; no PII; 7 days |

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
  - increments `hexa.errors`;
  - buffers upserts into `error_digest_daily`.
- The 08:00 digest posts the top 20 fingerprints with one request id each, via `app.Telegram.SendSync` [V-tg].

### 7.3 Where data goes

| Option | Cost | Verdict |
| --- | --- | --- |
| Grafana Cloud Free (metrics, traces, alerting, synthetics) | US$0; 10k series, 14 days [G-price] | **Recommended** |
| Railway metrics and monitors | Monitors need the Pro plan [R-alerts]; container metrics only | Complement |
| Prometheus + Grafana inside the app container | CPU, RAM and a TSDB volume inside the US$30 container | **Rejected**: competes with HTTP/SSR (D-08) and dies with what it watches, so it cannot report "app down" |

### 7.4 Synthetic probe

Grafana Synthetic Monitoring runs three checks:

| Check | Interval | Probe | Executions/month |
| --- | --- | --- | --- |
| `/up` | 60 s | A | 43,200 |
| `/up` | 120 s | B | 21,600 |
| `/up?checks=1` | 300 s | A | 8,640 |

The total is 73,440, under the 100k free tier [G-price]. Railway manages TLS (ADR 0002). Certificate-expiry checks are [NOT VERIFIED].

### 7.5 Alerts

Channels:

- Grafana alerting sends to Telegram, with email as fallback. The Telegram contact point is [NOT VERIFIED].
- App alerts go through the vukem Telegram handler [V-tg] with `TELEGRAM_MIN_LEVEL=error` and `TELEGRAM_RATE_LIMIT=20`. `ChatIDs` sends `LevelFatal` to the page chat; the error chat is `Silent: true`.
- Railway notifies about deploys by email and in-app [R-alerts].

On-call is the owner. Quiet hours run 23:00–08:00 America/Sao_Paulo: only `page` rings. Ack targets: page within 15 min while awake, warn the same day.

| # | Alert | Condition | Window | Sev | Source |
| --- | --- | --- | --- | --- | --- |
| A1 | App down | `/up` fails 2 consecutive probes | 2 min | page | synthetic |
| A2 | Readiness | `/up?checks=1` ≠ 200 on 2 runs | 10 min | page | synthetic |
| A3 | 5xx rate | > 1 % with at least 20 requests | 5 min | page | Grafana |
| A4 | Latency p95 | `web.public` 500 ms; `api.v1` 300 ms (NFR-05-1); search 300 ms; suggest 50 ms; `app.v1` 300 ms | 10 min | warn | Grafana |
| A5 | SSR fallback | `increase(hexa.ssr.fallbacks{public="true"}) > 0` | 5 min | page | Grafana + app |
| A6 | Job failed twice | same kind fails 2 runs in a row (`queue.JobFailed` via `q.AddObserver` [V-queue]) | — | warn | app |
| A7 | Ingestion | SDD 06 §10 list: failed run; no incremental in 26 h; `error` assert; TMDB data older than 170 d; breaker open over 1 h | SDD 06 | warn | Grafana |
| A8 | Backup missing | age > 93,600 s (26 h), or no data | 15 min | page (not during quiet hours) | Grafana |
| A9 | Moderation | oldest open > 20 h (NFR-04-2), or p95 age > 24 h | 30 min | warn | Grafana |
| A10 | Quota flush | 5 consecutive flush errors (SDD 05 §7) | — | page | app |
| A11 | Disk | DB size / `PG_VOLUME_BYTES` > 0.8 | 30 min | warn | Grafana |
| A12 | Connections | pool in use / max > 0.8 (8 of 10), or server connections > 0.8 × `max_connections` | 5 min | page | Grafana |
| A13 | Score drift | `score.dq.drift > 0` (SDD 02) | nightly | warn | Grafana |
| A14 | Trending stale | snapshot older than 3 h (SDD 08) | 15 min | warn | Grafana |
| A15 | Cost | Railway usage ≥ US$24 (80 % of US$30), set as "Custom email alert" [R-cost]; hard limit US$30 (D-27) | cycle | warn | Railway |
| A16 | Telemetry gone | no HTTP metric data | 10 min | warn | Grafana |
| A17 | Projected spend | on day 20 of the billing cycle (and daily after), month-to-date usage × days in cycle / elapsed days > US$27 (D-27: hard limit US$30) | daily | warn | GitHub Actions cron reading Railway usage [R-cost] (programmatic access [NOT VERIFIED]; fallback: calendar reminder to the owner on day 20) |

### 7.6 Dashboards

Dashboards live in Grafana Cloud; their JSON is in `ops/grafana/dashboards/`.

| Dashboard | Panels |
| --- | --- |
| Service health | rate, 5xx %, p50/p95/p99 per group, in-flight, SSR p95/fallbacks, pool, slow queries, CPU/RAM, uptime |
| API customers | requests per plan, quota denials, refunds, webhook success, active keys |
| Ingestion | runs by source/outcome, duration, records, 429s per host, breakers, DQ failures, TMDB age |
| Moderation | queue size, oldest, p50/p95 age, decisions/day, SLA breaches |
| Scoring integrity | recomputes, burst trips, flags, drift, ratings/24 h, signups, verified |

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
3. Grafana series and GB.
4. Postgres size vs plan.
5. Bucket size (≈ 2.1 GB per 100k items, NFR-06-7).
6. Egress; cover egress > US$5/month triggers the CDN decision (SDD 11 open issue 2).
7. Client intake volume (`client_reports` rows per day) and that the 7-day prune ran.

Budget-increase predictors (the first step is splitting out the worker, D-08):

- CPU > 70 % for 7 days;
- Postgres > 60 % of plan;
- night job backlog > 1 h.

## 8. Edge cases

- Exporter down: data is dropped, serving continues, and A16 fires.
- Telegram down: stdout still has the logs; Grafana falls back to email.
- SSR fallback on a logged-in route: counted with `public="false"`, no page.
- Container down: in-app alerts are silent, but A1, A8 and A16 are evaluated externally.
- Label cap overflow: the value becomes `other`, with one WARN.

## 9. Acceptance criteria

- `go test ./... -run 'TestMetrics_|TestLog_|TestReady_|TestSecurity_A09_Hexa'` passes.
- `promtool test rules ops/alerts/tests.yaml`, in a pinned `prom/prometheus` image, reports `SUCCESS` for A3–A16.
- `curl -s https://staging.hexascore.app/up?checks=1` returns 200 with all four checks `ok`.
- After `pkill node`, the same call returns 503 with `ssr: fail`, and a Telegram page arrives within 5 min.
- `curl -sI https://staging.hexascore.app/` shows `X-Request-ID` and `X-Trace-ID`, and the trace id opens in Grafana.
- After the k6 run of `qa/k6/api-v1.js`, Grafana shows at most 6,000 active series.
- `railway logs` lines parse as JSON with `level`, `request_id` and `route`, and a scripted grep finds no email pattern.
- The Railway usage page shows the US$24 email alert.

## 10. Test strategy

- Unit: fingerprinting, label caps, sampling (5xx always kept).
- `TestMetrics_EveryRouteGroupEmitsLatency`: `Config.MetricReader` = `sdkmetric.NewManualReader` [V-otel]; hit one route per group with `vukemtest`; assert one histogram point per group.
- Integration (Postgres): SQL gauges on fixtures; digest flush and prune.
- Alert rules: `promtool` fixtures fire and resolve each rule. Grafana ruler import is [NOT VERIFIED].
- `TestSecurity_A09_HexaLogsRedacted`: a request carrying a password, a bearer token, an API key, an email and a search query leaves none of them in the captured JSON.
- Synthetic: the staging probe is a D-21 gate. Monthly A1 drill: stop the service for 3 min.
- Load: k6 for NFR-12-3.

## 11. Open issues

1. VF-019: there is no app meter accessor. Workaround: `Config.Global` plus the upstream API. Owner: vukem. Closes when `otel.Meter()` is documented.
2. Request-logger policy: hash the user id, truncate the IP, make the request id equal the trace id. All three are [NOT IN VUKEM DOCS]. Owner: vukem (new VF entry). Closes when it is filed and resolved.
3. `/metrics`. Decided (D-35): no endpoint at launch (§6.2).
4. The Railway plan (Hobby or Pro) sets log retention (7 or 30 days) and whether monitors exist. Owner: architect (SDD 11). Closes when SDD 11 records the plan.
5. Status page hosting. Decided (12-5): GitHub Pages (§7.7).
6. SDD 08 metric names. Decided (12-6): §6.2 names are canonical; SDD 08 aligned.
7. Metric Views to drop high-cardinality attributes are [NOT IN VUKEM DOCS]. Owner: vukem. Trigger: staging series above 6,000.
8. Three Grafana points are [NOT VERIFIED]: Telegram as a contact point, the OTLP URL, and cert-expiry checks. Owner: architect. Closes in the observability roadmap task.
9. Railway hard usage limit. Decided (D-27): US$30, e-mail alert US$24, plus the day-20 projection alert A17.
10. Site documentation gaps, all [SITE PAGE MISSING]. Owner: vukem. Closes when a logging/health page is published.
    - the health check API (`AddHealthCheckTimeout`, `ReadyHandler`, `ok`/`fail` output);
    - route name and user id in request logs;
    - `logging.NewDedup` and `WithLogger` semantics;
    - the A09/A10 test names.

    Separately, there is no documented metric hook for scheduled tasks, only the ERROR log, so `hexa.schedule.runs` is an app wrapper around `s.Func`/`s.Job`.
