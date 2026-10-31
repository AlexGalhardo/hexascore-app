# Handoff: especialista (2026-10-31, session 11)

Reason: ten reviews plus the batch full gate (owner limit about 10). STATUS: HANDOFF.

## Done this session (local `main`, no push, no tag)

| Row | Verdict | Merge |
| --- | --- | --- |
| REV-P3.A.1.4 (+P6.C.3.2) | CORRIGIDO_E_APROVADO (activity DB error no longer 404, 984e782); app client from a built image + sdks/ts + sdks/python (uv 0.9.30 container) regenerated 777b07a; inventory routes added later 4416ee1 | c9c0967 |
| REV-P4.B.2.1 | APROVADO (registry.go conflict with trending jobs: both kept) | ea71ed5 |
| REV-P4.B.2.2 | APROVADO | 057a96c |
| REV-P4.B.2.3 | CORRIGIDO_E_APROVADO (create racing an ingest on the same Q-id: 409, not 500; test d36d45f) | f697974 |
| REV-FIX-community-api-source | APROVADO | 88903ca~ |
| REV-FIX-moderation-two-step (Q-A15) | CORRIGIDO_E_APROVADO (A01 also covers moderator 404 on item merge and affiliate routes, fe17eeb) | 2e89655 |
| REV-DOC-sdd05-trending | APROVADO | see TASKS |
| REV-FIX-lint-recommend-tests | APROVADO | see TASKS |
| REV-P6.C.3.1 | CORRIGIDO_E_APROVADO (absent `trending` prop guard, ffddb58) | 3c87ec5 |
| REV-FIX-trending-rollup-indexes | APROVADO (registry conflict resolved) | 08dd302 |

ROADMAP ticked: P3.A.1.4, P6.C.3.2, P4.B.2.1, P4.B.2.2, P4.B.2.3, P6.C.3.1.

## Full gate on main (batch, after the merges)

- First run at 82c0091: golangci `./...` 0 issues; root `bun run lint && typecheck && test` ok; `apps/mobile` lint/typecheck/test ok; `go test -tags integration -parallel 4 -timeout 30m ./...` FAILED three ways, all fixed on `agent/especialista/FIX-gate-2026-10-31` (merged 499b911):
  1. `TestPredeploy_ConcurrentRunsSerialise` hung (30 min timeout): vukem `Migrator.locked` waits in a blocking `pg_advisory_lock` statement whose snapshot the holder's `CREATE INDEX CONCURRENTLY` (new 190100..190400 migrations) waits on, on another pool connection, so no deadlock is detected. Production predeploy would hang on two overlapping deploys. App workaround: `app/console/predeploy.go` polls a try-lock first (`workaround(vukem): VF-NEW-migrator-lock-snapshot`); new row VF-NEW-migrator-lock-snapshot for senior-fw.
  2. `TestSearchSQLConfined`: P4.B.2.2 wrote `search_dirty` in `app/jobs`; now `search.MarkDirty` / `search.IsDirty` in `internal/search/dirty.go`.
  3. `TestAppCatalog` still expected 404 on `/app/v1/games/trending`.
- Rerun at 499b911: `go test -tags integration -parallel 4 -timeout 30m ./...` → exit 0 (34 packages ok). Lint/bun/mobile not rerun (the fix is Go only; golangci on jobs, search, console, tests → 0 issues).

## Open

1. REV-P6.D.1.1 (524cce3) and REV-P6.D.1.2 (6febd4f): MIGRATION ID COLLISION, both use `20261031190000`, already taken by `create_suggestion_reviews`; renumber (190500+) at merge. Report notes users.id bigint vs SDD text.
2. REV-P4.B.2.4 (revert; check the route gates on `revertCatalog` and has a TestSecurity_A01), REV-P4.B.2.5 (deviates from BeforeSave), REV-P4.B.2.6: stacked, merge in order.
3. Release: P4.B.2.x is releasable now that FIX-moderation-two-step is on main. Before the release PR run the ci.yml jobs locally (release gate). Owner: if Railway pins `CATALOG_API_SOURCES`, append `community`.
4. ROADMAP P4.B rows that say "moderate" for approve should read `applySuggestion` (orchestrator).

## Pitfalls

- Never trust the targeted run alone for route additions: run `-run Inventory` (P3.A.1.4 missed `route_inventory.txt`).
- The orchestrator commits TASKS.md/CHANGELOG with whatever is in the working tree, mine included; TASKS.md is often dirty with their edits.
- `wt/especialista-gate` is a reusable detached worktree with root node_modules (mobile installs there too); `git checkout --detach <sha>` before a gate.
- `git worktree remove --force` on a worktree with node_modules takes minutes.
- golangci "parallel golangci-lint is running": retry loop with 20 s sleeps, or `--allow-parallel-runners`.
