# ADR 0027: Stay on Postgres; the SQLite migration is deferred

Status: accepted (2026-11-07, owner chose "implement the recommended, stable solutions" on `sqlite-problems.md`). Related: ADR 0003 (Postgres only), ADR 0011 (search), ADR 0014 (backup), ADR 0024 (Redis deferred), SDD 11 (deploy and sizing), D-27 (hard spend limit).

## Context

The owner asked whether SQLite could be the only main database to save money on Railway. The analysis in `sqlite-problems.md` (repository root of the workspace, dated 2026-11-07) found a rewrite, not a driver swap: 47 to 67 engineer days, search quality loss in pt and es (no stemming, no trigram similarity), the backup and restore code to redo, a single app instance and a short outage per deploy (a Railway volume mounts in one deployment at a time, [NOT VERIFIED]). The net saving was estimated at US$8 to 11 a month. Option (c), making the current Postgres cheaper, was estimated at 1 to 2 days for US$4 to 6 a month with no risk.

## Decision

- Production and staging data stay on Postgres 18. ADR 0003, ADR 0011, ADR 0014 and SDD 11 are unchanged.
- SQLite is deferred, not rejected. It is reconsidered only for a reason other than cost (operational simplicity, local development without Docker), and then only through the 3-day spike of `sqlite-problems.md` section 7, with numbers for FTS5 search and write concurrency before any commitment.
- Cost is reduced through [docs/runbooks/postgres-cost.md](../runbooks/postgres-cost.md): measure, right-size the pool, fold staging into the production instance (second database) or make it ephemeral, then size the volume. Every step is owner-run and reversible.
- Code side: the cold tier of `source_record_history` already runs daily (`source_records.export`, ADR 0025, `routes/schedule.go`), so no new job was needed. The connection pool defaults are Railway variables (`DB_MAX_OPEN`, `DB_MAX_IDLE`), not code, and stay as they are until production measurements justify lowering them.

## Consequences

- Monthly saving stays modest and unmeasured until the owner runs the runbook.
- The local development cost of Docker Postgres remains.
- A later SQLite decision must start a new ADR that supersedes this one together with the ADRs it would replace.
