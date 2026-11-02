---
paths:
  - "database/migrations/**"
  - "app/accounts/deletion.go"
  - "app/privacy/**"
---

# Database changes

- Expand → migrate → contract: a column or table is added in one release, backfilled by a job, and the old shape is dropped two releases later (SDD 11 §7). Never rename in place.
- One migration per logical change, file `YYYYMMDDHHMMSS_verb_noun.go` registered in `database/migrations/registry.go`; `Down` really reverses it (the CI `migrations` job runs migrate, rollback one step, migrate). Concurrent indexes use `NoTransaction: true` and `CREATE INDEX CONCURRENTLY IF NOT EXISTS`.
- Every new query path gets its index in the same change; say which query it serves in the migration comment. Partial indexes for status-filtered tables.
- Every new table has a row in the retention table of SDD 11 §5 and, when it keeps data longer than 30 days, a prune task in `routes/schedule.go` (audit 2026-11-02 T-17 lists the tables that were missing one).
- Every table with a `user_id` either has a foreign key with an `ON DELETE` policy or an explicit entry in `purgeUserData` (`app/accounts/deletion.go`), plus the `TestSecurity_A03_*` test that proves the row disappears or is anonymised on account erasure (LGPD; audit T-05).
- Foreign keys to `items`, `editions` and `entities` cascade; foreign keys to `users` set null for content that must survive the account (ratings aggregate, suggestions) and cascade for private data.
- `jsonb` columns carry a short comment with the shape they hold and the code that writes it; `CHECK` constraints for enums instead of free text.
- Never run a migration or a data command against production from an agent session; prepare the exact command for the owner (`railway ssh ... /app/app <command> --dry-run` first).
