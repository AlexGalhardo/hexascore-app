# ADR 0003: PostgreSQL as the only data store; Redis and MongoDB require their own ADR

Status: accepted (2026-10-28). Decisions: owner requirement 4; brainstorming D-08, D-16.

## Context

The legacy app uses Postgres (Neon) for everything except API quota counters, which live in Upstash Redis (sliding window per second, Lua `incrByIfExists`, 4–5 commands per API request). vukem supports Postgres (`drivers/postgres`, pgx), Redis (`drivers/redis`: cache, session, queue stores, shared rate limiter, locks, pub/sub) and MongoDB (`drivers/mongo`) as separate modules, and ships memory and database drivers for cache, sessions, queue and rate limiting (`https://vukem.dev/docs/security/rate-limiting`, `https://vukem.dev/docs/database/mongodb` "Support matrix").

At launch the app runs as one instance (D-08). Every use case that normally justifies Redis (shared rate limits, shared cache, distributed locks, pub/sub between instances) is single-instance here.

## Decision

PostgreSQL is the system of record and, at launch, also the backing store for sessions, queue, cache entries that must survive restarts, idempotency keys, rate-limit counters that must be durable (API quotas) and the trending primitives. In-process memory stores are used for hot, loss-tolerant data (per-second throttles, rendered fragments) via vukem's memory drivers.

Adding Redis requires an ADR that names the trigger (second app instance, measured Postgres contention, or a feature that needs pub/sub) and the alternative rejected. Adding MongoDB requires an ADR that names a document workload Postgres `jsonb` cannot serve; none is foreseen.

## Alternatives considered

- **Redis from day one for API quotas (as legacy).** Adds a service and cost under the US$30 ceiling for counters a single Postgres instance handles at launch volumes (free plan 100 lifetime requests, pro 50k/month). Rejected until scale demands it.
- **MongoDB for catalogue documents.** `jsonb` with GIN indexes already serves the legacy `meta`/`external_ids` columns; the rewrite reduces jsonb use by promoting entities to tables (SDD 01). Rejected.
- **SQLite (vukem's default).** Fine for development and tests but not for a 100k-item catalogue with concurrent writes and FTS/trigram needs. Rejected for production; allowed for unit tests where the SQL is dialect-neutral.

## Consequences

- Positive: one backup, one failure domain, simpler local setup, transactional consistency between ratings and aggregates.
- Negative: API quota counters in Postgres cost one write per request; the design must batch (per-minute flush from memory) so Postgres is not hit per request; a crash loses at most one flush interval of counters (acceptable, and biased in the customer's favour).
- Obligations: integration tests run against real Postgres (`https://vukem.dev/docs/testing/overview`); the search design (ADR 0011) and trending (SDD 08) must fit Postgres; the infra SDD sizes the Postgres service for FTS indexes.
