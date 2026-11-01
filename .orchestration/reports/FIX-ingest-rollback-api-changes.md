# FIX-ingest-rollback-api-changes
status: concluida
branch: agent/senior/FIX-ingest-rollback-api-changes (worktree wt/senior-fw-rollback, from local main 065cb6d6)
files: `internal/ingest/rollback/rollback.go`, `internal/ingest/normalize/{job.go,changes.go}`, `tests/ingest/{rollback_test.go,normalize_job_test.go}`
red: tests first; `go test -tags integration ./tests/ingest -run 'TestRollbackWritesAPIChanges|TestNormalizeWritesAPIChanges'` failed (0 api_changes rows for 2 deleted/hidden items, 0 created, no updated row).
green: same command ok; `-run 'TestRollback|TestNormalize|TestRecompute|TestLocked|TestSlug|TestCancelled|TestBad'` ok; golangci-lint (--build-tags integration, normalize, rollback, tests/ingest) 0 issues.
design: rollback writes a `deleted` row for every item it deletes or hides (a hidden item leaves the public API the same way). Normalize snapshots the API-visible values of an existing item (slug, wikidata, releaseDate, runtimeMinutes, pageCount, externalIds, title and synopsis per locale) before and after `saveItem` and writes `created` for a new item or `updated` with the differing field names; an unchanged re-run writes nothing. Rows are written inside the item savepoint, so a failed item leaves none.
not done (outside the asked scope, still in the row text): a non-wikidata rollback does not requeue the Wikidata record or clear that source's cover; restored IGDB/TMDB payloads keep the bad run's `checked_at`; `entities` (credits) changes are not part of the normalize snapshot, so a credits-only change writes no row.
changelog: Fixed (ingestion and rollback now feed /changes).
roadmap: no box (follow-up of P6.B)

## Review (especialista, 2026-11-01)
verdict: CORRIGIDO_E_APROVADO
re-run: `go test -tags integration -p 1 ./tests/ingest -run 'TestRollback|TestNormalize|TestCovers|TestAlertA7|TestRecompute|TestLocked|TestSlug|TestCancelled|TestBad|TestResume|TestWiring'` → ok (32 s); `go test ./internal/ingest/covers ./internal/ingest/normalize` → ok; `golangci-lint run --timeout 15m -j 2 --build-tags integration ./tests/ingest/... ./internal/ingest/... ./app/jobs/...` → 0 issues; `golangci-lint fmt --diff` → clean
fixes (CORRIGIDO): 1. rows.Err() checks in tests/ingest/resume_test.go and wiring_test.go, the two rowserrcheck findings (8d1a251c)
gate on main after merge: go build ok; targeted tests and lint above (identical tree); full integration suite not run (host short on memory), left to the batch gate
PRONTO_PARA_MERGE agent/especialista/FIX-ingest-rollback-api-changes-fix (merge 9c366902)
checked: sdd 05 §7.3 and §8, sdd 01 FR-01-14, authz n/a, secrets n/a, sql bound, i18n n/a, scope ok, tests red→green
notes: `deleted` for a hidden item is right: FR-01-14 excludes non-public items from the paid API (404), so for a /changes consumer it is gone exactly like a deleted one; SDD 05 "hidden-only changes write nothing" means non-public *fields*, which `changes.Visible` already drops. Snapshot keys are all in the public allowlist. Rows ride the item savepoint, so a failed item writes none.
gaps (not blocking, follow-up row FIX-ingest-rollback-gaps): 1. non-wikidata rollback does not requeue Wikidata nor clear that source's cover; 2. restored IGDB/TMDB payloads keep the bad run's `checked_at`; 3. `entities` (credits) changes are not in the normalize snapshot; 4. normalize writes `updated` for an item that is already `hidden`/`takedown` (feed names an id the API answers 404 for), and nothing re-announces an item made public again.
