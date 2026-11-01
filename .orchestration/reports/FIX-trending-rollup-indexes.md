# FIX-trending-rollup-indexes Index the rollup event-time columns
status: concluida
branch: agent/senior/FIX-trending-rollup-indexes @ 33aa330
files: database/migrations/20261031190100_add_trending_rollup_indexes.go, database/migrations/registry.go, tests/trending_rollup_indexes_test.go
acceptance:
- `go test -tags integration ./tests -run TestTrendingRollupIndexes -v` -> ok; 100k rows per table over 365 days, one-day window EXPLAIN is a Bitmap Index Scan on rating_events_at_idx, item_status_updated_at_idx, list_items_added_at_idx, suggestion_events_created_at_idx (all four)
validation: golangci-lint (integration tag) 0 issues, gofmt ok; migrations ran through vukemtest (NoTransaction path)
changelog: Changed: indexes on rating_events.at, item_status.updated_at, list_items.added_at and suggestion_events.created_at so the hourly trending rollup no longer scans the full tables.
roadmap: no box
risks / follow-ups:
- Four single-statement migrations (20261031190100..190400), CREATE INDEX CONCURRENTLY IF NOT EXISTS with NoTransaction (vukem schema.Migration), per its one-statement rule. Timestamps are later than every migration on main (160000) and on other worktrees (173100); P6.D.1.1 uses 190000.
- Not run on the full 160k production-size seed; the test uses 100k rows per table with FKs bypassed (session_replication_role).
- Check docs/runbooks/migrations.md predeploy INVALID-index cleanup covers a failed CONCURRENTLY build (it should, per AGENTS.md hexascore:predeploy).

## Review (especialista, 2026-10-31)
verdict: APROVADO
re-run: `go test -tags integration ./tests -run 'TestTrendingRollupIndexes|Migrat'` → ok (61.5 s; EXPLAIN uses the four new indexes); on main after merge `-run 'Migrat|TestTrendingRollupIndexes|Moderation'` → ok; `bun test scripts/ci/migration-policy.test.ts` → 3 pass; golangci (integration tag) ./database/... → 0 issues
checked: one statement per NoTransaction migration, `CONCURRENTLY IF NOT EXISTS` with a concurrent Down, ids 190100..190400 sort after 190000 (suggestion_reviews) on main; failed builds are covered by `hexascore:predeploy` dropping INVALID indexes (runbooks/migrations.md); identifiers are constants; scope, tests
gate on main after merge 08dd302: go build ok
PRONTO_PARA_MERGE agent/senior/FIX-trending-rollup-indexes
