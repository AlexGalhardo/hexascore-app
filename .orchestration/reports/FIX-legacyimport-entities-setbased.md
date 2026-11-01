# FIX-legacyimport-entities-setbased

Branch: agent/senior/FIX-legacyimport-entities-setbased (worktree `C:\dev\hexascore-app\wt\senior-FIX-legacyimport-entities-setbased`, from main 7facec7). Commit 1e755c59 (first 6a640ee, then a test fix).

## What
- `entities` stage is set-based: a batch of 1,000 items costs about 12 statements (was about 5,900). Read items, seed `legacy` source, look up entities by QID and by kind+name (2 statements), lock + slug read + one multi-row `INSERT ... ON CONFLICT DO NOTHING` for the missing ones, one `item_entities` insert (`ON CONFLICT DO NOTHING`), one `editions` insert (`ON CONFLICT (isbn13) DO NOTHING`), plus the checkpoint update.
- Same results: entity matched by QID (any kind), else native by kind + name; slugs assigned in the same order as `UniqueSlug` (base, `-2`, `-3`, redirects and reserved words honoured) via new `catalog.FirstFreeSlug`; same roles, positions, sources, editions.
- Same checkpoint: `legacy_import.items.id > checkpoint`, batch key = last item id. Production at 5,200 `running` resumes from there (the runner path is unchanged; the per-batch cache is gone, a batch re-reads).
- `--workers N` (default 1) on `hexascore:import-legacy`, `Runner.Workers`: stages with `Stage.KeysSQL` (entities, social, library) get their batch boundaries in one query; N goroutines each take one batch (disjoint id range) in its own transaction. The checkpoint advances only to the end of the longest prefix of committed batches (mutex-guarded), so a crash re-runs finished-out-of-order batches idempotently and never skips one. A short tail batch and the stage-ending empty batch (which runs social's `last` summary) stay sequential after the parallel part.
- Racing entity creation: before creating, a worker takes sorted `pg_advisory_xact_lock`s on the QID and on kind+slug base, re-reads, then inserts with `ON CONFLICT DO NOTHING`; the loser finds the winner's row after its commit.
- social and library batch by user-id range (`userBatches`), so they take the same option; they are several `INSERT ... SELECT` per batch, not one statement. `users`, `catalog` and `billing` are unchanged (no `KeysSQL`).

## Proof (integration, Postgres on 5433)
1. Oracle first: the old per-item code is kept as `database/legacyimport/entities_reference_test.go` (test-only). `TestEntitiesSetBasedMatchesReference` ran old vs old before the rewrite (harness check), then old vs new after: 400 synthetic items on top of the rehearsal fixture (colliding slugs Ação/Acao, QID vs native same name, 日本, blank names, valid/duplicate/invalid ISBNs), batch 37. Compares entities (kind, QID, name, slug), item_entities (item, role, position, source, entity) and editions row by row: equal, including slugs and the checkpoint row.
2. `TestEntitiesWorkersMatchSequential` (4 workers vs 1, slugs excluded since the `-2` suffix depends on commit order; also asserts no duplicated native entity): pass.
3. `TestEntitiesWorkersResume`: a batch killed mid parallel run; checkpoint never passes an unfinished item; resumed run equals an uninterrupted one: pass.
4. `TestSocialLibraryWorkersMatchSequential` (users, catalog, social, library; batch 1; 4 workers vs 1; ratings, audit, lists, list_items, statuses, time_to_beat, contact_messages, rounding summary, checkpoints): pass.
5. Existing `TestEntitiesStage`, runner, social, library tests pass.
6. Benchmark (`HEXASCORE_BENCH=1 go test -tags integration -run TestEntitiesThroughput -v`), 3,000 items, batch 1,000, local DB (no latency):

| Path | Time | Items/s | Statements | Per 1,000 items | At 150 ms/statement |
| --- | --- | --- | --- | --- | --- |
| per-item reference | 40.0 s | 75 | 17,620 | 5,873 | 44 min |
| set-based, 1 worker | 3.8 s | 793 | 36 | 12 | 5 s |
| set-based, 4 workers | 2.8 s | 1,069 | 36 | 12 | 5 s |

Projection for 98,841 items over a 150 ms link: about 100 batches x 12 statements x 0.15 s = about 3 minutes with 1 worker (the old path: about 40 h). Workers matter little locally (CPU/lock bound, 1.35x); over a link they overlap the round trips, so expect close to N times on the network part. Not measured over a real link.

## Owner action
Run the stage inside Railway, next to the database, as a one-off command (not from a laptop): `./app hexascore:import-legacy --stage entities --workers 4` on the `app` service environment (no `railway up`). Needs a DB pool of at least `--workers` connections (`DB_MAX_OPEN` or equivalent). Production sits at checkpoint 5,200, status `running`, no process attached: it resumes after that key. I never ran anything against production.

## Limits
- With workers > 1, which of two names that slugify alike gets the plain slug vs `-2` depends on commit order (workers=1 matches the old order exactly). A rare slug race (two different names whose slug candidates collide through a `-N` suffix) fails the batch with "rerun the stage"; the checkpoint keeps it safe.
- When two workers carry the same ISBN13 in different ranges, either work may win the edition (the old code: lowest item id); the worker equivalence tests therefore compare editions by ISBN13 and language only.
- Lint: `golangci-lint --timeout 15m -j 2 --allow-parallel-runners --build-tags integration` on database/legacyimport, app/catalog, app/console produced no findings after `golangci-lint fmt`.
- Not touched: CHANGELOG, ROADMAP, PROGRESS (docs/runbook for the `--workers` flag and Railway one-off are for the specialist/orchestrator).

## Review (especialista, 2026-11-01)
verdict: CORRIGIDO_E_APROVADO
re-run: `go test -tags integration -count=1 ./database/legacyimport/` ok (TestEntitiesSetBasedMatchesReference, TestEntitiesWorkersMatchSequential, TestEntitiesWorkersResume, TestSocialLibraryWorkersMatchSequential, TestEntitiesStage pass); `go test ./app/catalog` ok; golangci-lint `--timeout 15m -j 2 --build-tags integration` on database/legacyimport, app/catalog, app/console 0 issues
checked: 1. oracle compares old per-item code vs new row by row (entities incl. slugs, item_entities, editions, checkpoint) on a hard fixture: sound. 2. runParallel moves `last_key` only across the contiguous prefix of committed batches (mutex, `done[next] > 0`; full batches always report rows > 0), the update runs after each batch commit, so a crash re-runs out-of-order batches and never skips one; a cancelled update only lags the checkpoint. 3. advisory xact locks on kind+slug base and QID are sorted and taken in one statement, then re-lookup (READ COMMITTED sees the winner) and `ON CONFLICT DO NOTHING`; residual slug race fails the batch with "rerun the stage" (safe). 4. `FirstFreeSlug` matches UniqueSlug minus the year step; unit test added. sql bound, no secrets, scope ok.
fixes: 1. editions insert ordered by isbn13 (then item order, so the lowest item still wins): two workers sharing ISBNs across ranges could deadlock on the unique index in reverse orders. 2. long-base LIKE prefix trimmed of trailing dashes like `withSuffix` (only bit at 7-digit suffixes). 3. `TestFirstFreeSlug`. 4. runbook section in docs/runbooks/migrations.md with the one-off `./app hexascore:import-legacy --stage entities --workers 4` (2f664c82)
gate on main after merge: go build ok; targeted tests only (section 8.2)
PRONTO_PARA_MERGE agent/especialista/FIX-legacyimport-entities-setbased-fix
