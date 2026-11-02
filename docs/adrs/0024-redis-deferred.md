# ADR 0024: Redis is deferred; Postgres and process memory stay the only stores until a named trigger fires

Status: accepted (2026-11-02, architect audit Phase 4). Supersedes nothing; refines ADR 0003 with measured data and explicit triggers. Related: D-08 (single service), D-27 (US$20 hard limit), SDD 06 §7.6 (ingestion jobs), SDD 08 §7.1 (view counters), RFC-001 (catalogue worker).

## Context

ADR 0003 made Postgres the only data store and asked for a dedicated ADR before Redis is added. The audit of 2026-11-02 measured the production baseline and inventoried what the framework offers, so the question can be answered with numbers instead of habit.

Measured on Railway, 7-day window ending 2026-11-02 (MCP `get-service-metrics`):

| Service | CPU avg / max (vCPU) | Memory avg / max | Disk |
| --- | --- | --- | --- |
| `app` | 0.004 / 0.74 | 0.11 GB / 0.82 GB | — |
| `Postgres` | 0.006 / 0.77 | 0.26 GB / 4.03 GB (import peak) | 5.55 GB of 10 GB |

Workspace spend month-to-date: US$8.36 against a US$20 hard limit (D-27 as amended 2026-11-01). Railway bills per vCPU and per GB of memory per month plus volumes and egress, so any always-on service costs money even when idle; a small Redis with 256 MB reserved sits around US$3 to US$5 per month (estimate from the Railway pricing page, to be re-checked by the data-sources research; "not verified" at the time of writing).

Candidate uses and what serves them today:

| Use case | Today | Postgres-only alternative | Redis gain |
| --- | --- | --- | --- |
| Public page cache | In-process render cache, 60 s TTL, keyed per locale after the 2026-11-02 fix (`app/http/middleware/render_cache.go`) plus Railway edge `Cache-Control` | Same; `vukem_cache` database store for anything that must survive a restart | None with one replica |
| Rate limiting | `RATE_LIMIT_STORE=cache` on the database store (fixed window) | Same | Shared token buckets across replicas: needed only with ≥ 2 replicas |
| Sessions | `SESSION_DRIVER=database` | Same | Latency ~1 ms vs ~0.3 ms, irrelevant at current traffic |
| Queue | `queue.Database` with `FOR UPDATE SKIP LOCKED`, queues `default,ingest` | Same; vukem's driver already reserves with `SKIP LOCKED` and has `failed_jobs` | Higher throughput (thousands of jobs/s); the catalogue worker needs ≤ 10 jobs/s at the 1 M-item target (RFC-001 §6) |
| View and click counters | `cache.Buffer` write-behind in memory, flushed every minute to `item_activity_daily` (SDD 08 §7.1) | Same | Durability of the last minute only |
| Distributed locks | `WithoutOverlapping()` on the scheduler (memory lock store, fails closed on `OnOneServer` without a shared store) | `cache.Lock` on the database store, or Postgres advisory locks | Needed only when the scheduler runs in two processes |
| API quota counters | In-process meter flushed every minute (`app/api/quota.go`, shortcut comment) | `UPDATE api_quota_counters … RETURNING` atomic increment | Needed only with ≥ 2 replicas |
| Pub/sub (SSE, broadcast) | Not used by the product | `LISTEN/NOTIFY` via vukem `broadcast` when needed | Needed only across replicas |

Every row where Redis "wins" has the same precondition: more than one app process sharing state. D-08 keeps one process; RFC-001 proposes a second Railway service for the catalogue worker, but that worker shares nothing with the web process except the Postgres queue tables, and the scheduler stays in the web process.

## Decision

Redis is **deferred**. The app keeps the database and memory drivers. Before adding Redis, the following must be true and written in a new ADR that supersedes this one:

1. A trigger fired:
   - the web service runs ≥ 2 replicas (needs shared rate limits, shared quota meter, shared scheduler lock), or
   - measured Postgres contention from the queue or counters: `pg_stat_statements` shows the `vukem_jobs` reserve query or the `item_activity_daily` upsert above 5 % of total time, or lock waits above 50 ms p95 for a week, or
   - a product feature needs cross-process pub/sub with sub-second fan-out (live feed, presence).
2. The cost fits: projected spend with Redis stays under US$18/month (the SDD 12 projected-spend alert threshold) for three consecutive months.
3. The scheduler is moved to `OnOneServer()` with the Redis lock store in the same change, so two processes cannot double-run nightly jobs.

Until then, the single-replica shortcuts stay documented in code (`shortcut:` comments) and listed in `docs/BACKLOG.md` under "multi-replica readiness".

## Alternatives considered

- **Add Redis now for the catalogue worker queue.** Rejected: the Postgres queue handles the worker's rate (RFC-001 estimates ≤ 10 jobs/s; `SKIP LOCKED` on a 5-vCPU-second budget sustains hundreds/s), and a second service plus Redis would push spend toward the cap before the worker proves its value.
- **Add Redis for the page cache.** Rejected: the render cache is per process and the edge caches public pages for 60 s; a shared cache only matters with replicas.
- **Use Postgres `UNLOGGED` tables for hot counters.** Kept as the first thing to try if `item_activity_daily` writes ever show up in `pg_stat_statements`.

## Consequences

- Positive: no new service, no new secret, no new failure mode; spend stays near US$9/month plus the worker.
- Negative: scaling to two web replicas is a project (quota meter, rate limits, scheduler lock), not a slider. The backlog carries it as P2 "multi-replica readiness" with the three code changes above.
- Reassessment: at each phase gate of RFC-001 (250 k, 500 k, 1 M items) and whenever Railway metrics show app CPU above 70 % for 7 days (cost runbook predictor).
- Rollback plan if Redis is ever adopted: all vukem drivers are swappable by `CACHE_STORE`, `RATE_LIMIT_STORE`, `SESSION_DRIVER`, `QUEUE_DRIVER` style variables; switching back to `database` loses only in-flight cache entries and rate-limit windows, never durable data, because the queue tables and counters remain in Postgres.
