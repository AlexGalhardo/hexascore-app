# ADR 0014: Backups with RPO 24 h and RTO 1 h

Status: accepted (2026-10-28). Decisions: brainstorming D-08, D-18; owner requirement 7 (backups, rollbacks, safe migrations).

## Context

Legacy backups depended on Neon's branching and a documented local backup procedure (`legacy/docs/backup-local.md`). The rewrite runs one Postgres on Railway under a US$30/month ceiling. Catalogue data is re-derivable from sources over days; user data (accounts, ratings, lists, suggestions) is not.

## Decision

- **RPO 24 hours, RTO 1 hour** at launch. A scheduled job runs `pg_dump` (custom format, compressed) daily at a low-traffic hour to a Railway bucket or S3-compatible store, retention 30 daily snapshots plus 12 monthly; the dump is verified by size and a `pg_restore --list`; a monthly automated restore into a scratch database on staging proves the RTO and runs the data quality asserts.
- **Covers and uploads** live in object storage. Amendment 2026-10-28 (SDD 11 finding: Railway buckets offer no object versioning and no server-side encryption): cover keys are immutable and content-hashed (`covers/<id>-<hash16>.webp`), replaced covers are kept for 30 days before a deferred delete job removes them, and database dumps are encrypted client-side (vukem `crypt` or `age`) before upload. No separate backup of covers; they are re-derivable from sources. Amendment 2026-11-01 (owner decision, round item 18): item covers move to slug keys (`covers/{section}/{slug}.webp`, `-t.webp` thumbnail), so a cover replaced by the covers job is overwritten in place instead of kept 30 days; it stays re-derivable from `cover_source`. Keys a row stops pointing at (a rename, `covers:rekey`, a delete) still go through the 30-day deferred delete, and edition covers keep content-hashed keys.
- **Migrations** are forward-only, run by the app at boot in a transaction per migration with an advisory lock, never destructive in the same release as the code that stops using a column (expand → migrate → contract across two releases). Every migration ships with a tested `down` where the dialect allows and a documented manual rollback otherwise.
- **Restore runbook** in SDD 11: where the dumps are, how to restore to a new Railway Postgres, how to repoint the app, how to replay writes lost since the last dump (none expected; users are told the window).

## Alternatives considered

- **Continuous WAL archiving (pgBackRest, wal-g) for RPO ≤ 1 h.** Better RPO; more operations and storage; may exceed the ceiling. Deferred to the first budget increase.
- **Railway's native backups only.** Unverified retention and no restore drill. Rejected as the sole mechanism; may be enabled in addition.

## Consequences

- Positive: predictable cost; restores rehearsed, not assumed.
- Negative: up to 24 hours of user writes can be lost in a disaster; stated in the terms and on a status page.
- Obligations: SDD 11 specifies the job, storage, encryption of dumps (vukem `crypt` or server-side encryption of the bucket), alerting on a missed dump (SDD 12), and the drill checklist.
