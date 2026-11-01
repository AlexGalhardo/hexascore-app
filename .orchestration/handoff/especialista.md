# Handoff: especialista (2026-11-01, session 16)

Reason: 14 rows reviewed and merged (about 12 reviews); context heavy. STATUS: HANDOFF.

## Done this session (local `main`, no push, no tag)

| Rows | Verdict | Merge |
| --- | --- | --- |
| P2.C.3.3, P2.C.alerts-a5-test, P2.C.4.2, P2.C.search-reindex, P2.B.2.2, PERF-i18n-page-namespaces (senior-b batch) | CORRIGIDO_E_APROVADO (fixes 2aaa66d, 6cc9e68 from session 15) | 7044df3; ROADMAP P2.C.3.3, P2.C.4.2, P2.B.2.2 ticked |
| FIX-legacyimport-entities-setbased (prio 0) | CORRIGIDO_E_APROVADO: editions insert ordered by isbn13 (cross-worker deadlock), long-slug LIKE prefix trimmed, `TestFirstFreeSlug`, runbook section "Legacy import: entities stage" in docs/runbooks/migrations.md with `./app hexascore:import-legacy --stage entities --workers 4` (2f664c82) | merged after a3f1791 |
| FIX-search-hit-meta, FIX-library-item-score | CORRIGIDO_E_APROVADO: `SearchHit` is in `/app/v1`, golden + app client regenerated (84ebe78c) | merged before 2bea45d |
| P6.D.1.3, IMPL-web-v1-edition-entity, -recent-changes, -docs-legal-errors, -more-like-this | CORRIGIDO_E_APROVADO: routes/web.go conflict resolved to main's base list (terms./policy. only through PageNamespaces, dot-less `updatedLabel`), mobile catalogues recopied (f70d09a7) | merged before 5b1c252; ROADMAP P6.D.1.3 ticked |

Openapi on main after the merges: `scripts/openapi-check.sh` ok, `codegen:client -prefix /app/v1` equals `packages/app-client/src/api.ts`, redocly lint ok, app-client typecheck + test ok. Not run: the Docker-image part of the `openapi` job (Schemathesis).
Full gate on main was not run this session (section 8.2: targeted tests only per row). Run `scripts/ci/local.sh --only test-go,test-ts,mobile` then `--only test-go-integration` alone first.

TASKS.md rows were set to `ready` on disk but TASKS.md is not committed (it carries other agents' uncommitted edits).

## Open, in order

1. senior-up: P3.B.1.1 -> 1.2 -> 1.3 -> P3.B.2.1 (stacked; review the 2.1 tip as one batch), P3.B.3.1, FIX-problems-redirect-301, FIX-docs-next-cursor. P4.B.5.1 is still `doing`. PERF-web-lcp / FIX-lighthouse-budgets-main still `doing`.
2. senior-a: P5.B.4.2, P5.B.2.4, P5.B.2.5, later rows.
3. senior-fw P6.A stack; the session 14 verification list still applies (1.3 continuation lock/guard, 2.1 resume skip, 1.4 breaker probe and timeout, 2.2/2.3 infra errors counted as target failures, 2.2 unbounded refs cache, 1.4 thin SSRF tests).
4. REV-FIX-seed-flag: do not merge (hotfix v0.12.4, PR #64).

## Cleanup left

- Merged branches still checked out in old worktrees: `agent/especialista/P2.B.2.2-fix`, `P2.C.3.3-fix`, `P2.C.4.2-fix`, `IMPL-web-batch2`. Remove those worktrees and branches when idle.
- senior-b's worktree `wt/senior-FIX-legacyimport-entities-setbased` and branch are senior-b's to remove.
- `wt/senior-FIX-app-item-cover-urls` folder (Permission denied) and detached `Temp/tmp.*` worktrees from session 15.
- Untracked `resources/js/vite-env.d.ts` in the main checkout is not mine; left alone.

## Pitfalls (new)

- Worktrees have no node_modules: `cmd /c mklink /J <wt>\node_modules <main>\node_modules`, then `cmd //c rmdir <wt>\node_modules` before `git worktree remove` (rmdir removes only the junction).
- `git diff --no-index` cannot read `/tmp` paths on Git Bash; write temp files to the scratchpad and use `diff -q`.
- Mobile catalogues: `bun apps/mobile/scripts/copy-catalogues.ts` from the repo root after any `resources/lang` change in a merged batch.
