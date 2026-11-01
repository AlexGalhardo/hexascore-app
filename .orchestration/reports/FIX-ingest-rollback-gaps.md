# FIX-ingest-rollback-gaps
status: concluida
branch: agent/senior/FIX-ingest-rollback-gaps (worktree wt/senior-fw-gaps, from local main fb1b61df)
files: `internal/ingest/rollback/rollback.go`, `internal/ingest/normalize/{changes.go,job.go,save.go}`, `internal/ingest/store/records.go`, `database/migrations/20261101220000_add_rollback_gap_fixes.go` (+ registry), `app/jobs/ingest.go`, `tests/ingest/{rollback_gaps,normalize_gaps}_test.go`, `app/jobs/ingest_pool_test.go`, SDD 06 and 11 notes
red: tests first; `go test -tags integration ./tests/ingest -run 'TestRollbackOfNonWikidata|TestRollbackRestore|TestNormalizeCreditsOnly|TestNormalizeNonPublic'` failed on all 5 (wikidata not requeued, cover kept, no `checked_at` column, no entities row, hidden item wrote a row, no re-announce).
green: same command ok; `-run 'TestRollback|TestNormalize|TestRecompute|TestLocked|TestSlug|TestCancelled|TestBad|TestCovers|TestWiring'` ok (36 s); `go test ./internal/ingest/... ./database/migrations ./app/jobs` ok; migration-policy test ok; golangci-lint (-j 2, integration tags, tests/ingest, internal/ingest, app/jobs, migrations) 0 issues.
design:
1. Rollback collects every restored and deleted record. For non-wikidata ones it sets the Wikidata record's `normalized_at` NULL and clears `cover_*` on items whose `cover_source` is that source (the `cover_gc` trigger releases the keys; the covers job re-picks).
2. Migration adds `source_record_history.checked_at`; `store.upsertSQL` archives it; rollback restores `coalesce(h.checked_at, h.fetched_at)` (rows archived before the migration fall back to `fetched_at`, the safe side for the 180-day TMDB clock).
3. Normalize wraps the credits save (`saveRelations`) with a before/after fingerprint of `item_entities` (role, entity, position, character) and writes `updated` with `entities` when it differs, only for an item that existed and is public (a new item already has `created`).
4. SDD 05 §7.3 and FR-01-14: a non-public item is 404 on the API, so normalize writes no row for it (no snapshot, no credits row). A trigger `items_visibility_changes` on `items.visibility` files `deleted` when an item leaves public and `created` when it comes back, whatever wrote the column (rollback, admin, SQL); hidden to takedown writes nothing. Rollback no longer writes the explicit `deleted` for a hidden item (the trigger does; a deleted item keeps the explicit row). `TestRollbackWritesAPIChanges` still asserts exactly one row each.
pool: vukem reads `DB_MAX_OPEN` (default 10, production 20). A step holds one connection and `covers.Reserver(db)` takes a second, so a pool of 1 would hang. `checkPool` in `newIngestRunner` fails the runner build for `DB_MAX_OPEN` 1 or invalid (0 = unlimited and unset pass); unit test `TestCheckPool`; SDD 11 note.
not done: the restored-record cover is cleared even if it would have been the same (one extra covers run); `item_entities` rows of vanished credits are still never deleted (earlier handoff).
changelog: Fixed (rollback of non-wikidata runs, restored checked_at, credits in /changes, non-public items in /changes). Not edited (rule).
roadmap: no box (follow-up of P6.B)

## Review (especialista, 2026-11-01)
verdict: CORRIGIDO_E_APROVADO
re-run: `go test -tags integration ./tests/ingest -run 'TestRollback|TestNormalize|TestRecompute|TestLocked|TestSlug|TestCancelled|TestBad|TestCovers|TestWiring'` → ok 37 s; `go test ./internal/ingest/... ./database/migrations ./app/jobs` → ok; `go test -tags integration ./tests -run 'TestAPIChanges|TestChanges|TestWebhook|TestMigration|TestCommunitySchema'` → ok; throwaway up/down/up test (trigger + `checked_at` present, gone, present again) → ok; `golangci-lint run --timeout 15m -j 2 --build-tags integration` on the touched packages → 0 issues.
checked: sdd 05 §7.3 and 06 §7.8, 11 pool, sql (bound params, `unnest` pair), scope, tests.
- Duplicates: only rollback writes `items.visibility` (admin `ItemsResource` is read-only, no moderation/takedown writer yet, merge and seed purge DELETE, INSERTs do not fire an `AFTER UPDATE OF` trigger). Rollback now writes the explicit `deleted` only for deleted items; hidden ones get the trigger row. No double row.
- Contract: trigger rows take `changed_at` default `now()` and the bigserial id like `changes.Record`, so the `(changed_at, id)` cursor order holds; re-publication is `created` (client re-fetches a whole item it was told was deleted); hidden↔takedown writes nothing.
- Cost: `AFTER UPDATE OF visibility ... WHEN (OLD.visibility IS DISTINCT FROM NEW.visibility)`, so ordinary `items` updates never call the function.
- Migration id 20261101220000 is unique across all local branches; last on main is 20261101210000.
- Cover clearing on restored records is right (the restored payload may carry a different cover); the cost is one extra covers pass. `coalesce(h.checked_at, h.fetched_at)` errs toward an earlier TMDB expiry, the safe side. `checkPool` rejects only 1 and garbage; 0/unset pass, matching database/sql.
fixes: 1. SDD 05 §7.3 lists the `items_visibility_changes` trigger among the `api_changes` writers (fix branch commit).
follow-ups (not blocking): scoring `ItemScoreChanged` and community apply still write `updated` for a non-public item; rollback sets a `takedown` item it created to `hidden` (pre-existing, run-created items are rarely takedown).
gate on main after merge: targeted only (host memory); full gate left to the batch gate.
PRONTO_PARA_MERGE agent/especialista/FIX-ingest-rollback-gaps-fix
