# Handoff: especialista (2026-10-31, session 12)

Reason: 17 reviews this session (owner limit about 10). STATUS: HANDOFF.

## Done this session (local `main`, no push, no tag)

| Row | Verdict | Merge |
| --- | --- | --- |
| REV-P4.B.2.4 | APROVADO (revert/audit gate on revertCatalog/viewCatalogAudit, adminOnly) | b60b739 |
| REV-P4.B.2.5 | APROVADO (BeforeSave deviation accepted: dedicated locked route is safer than vukem resource saves) | c39da38 |
| REV-P4.B.2.6 | APROVADO (locks/export gate on lockCatalog/viewCatalogAudit) | ca5b800 |
| REV-P4.B.2.7 | APROVADO (skipped e2e/k6 accepted per 8.2) | 1ca36e6 |
| REV-P4.B.2.3b | CORRIGIDO_E_APROVADO (ON CONFLICT on item create restored in merge) | 56ad491 |
| REV-P4.B.3.1 | APROVADO (migration 203100 ok) | b1701c3, gofmt 7639f4f |
| REV-P4.B.3.2 | APROVADO | 5c7fe1a |
| REV-P6.D.1.1 | CORRIGIDO_E_APROVADO (migration renumbered 20261031203200) | via agent/especialista/P6.D.1.1-fix |
| REV-P6.D.1.2 + REV-PERF-compute-similar | APROVADO (@7673421; no new perf row, PERF-compute-similar-500k covers it) | b410f41 |
| REV-P6.D.2.1 | APROVADO | 265fafb |
| REV-P6.D.2.2 | CORRIGIDO_E_APROVADO (route inventory for /for-you and /app/v1/me/for-you) | 4de9592 |
| REV-P6.D.3.1 | APROVADO | d5d6cdd |
| REV-FIX-trending-7d-home | CORRIGIDO_E_APROVADO (SDD 05 note; superseded P6.C.3.1 home strip, TrendingSection and TestHome_Trending removed) | 565f4db |
| REV-FIX-www-redirect | APROVADO | 6da4602 |
| REV-FIX-openapi-api-host-paths | APROVADO; goldens, TS/Python SDKs, app client regenerated on main | 9641b99, 116e4df |
| REV-P6.A.1.2 | APROVADO (migration 20261101120000 ok) | 6fe9b99 |

ROADMAP ticked: P4.B.2.4-2.7, P4.B.3.1, P4.B.3.2, P6.D.1.1, P6.D.1.2, P6.D.2.1, P6.D.2.2, P6.D.3.1, P6.A.1.2. CHANGELOG lines added (2d042a5, 127eeeb); the old P6.C.3.1 "Top 10 trending (30 days)" home line was replaced by the 7-day line.

## Batch gate (scripts/ci/local.sh on 127eeeb)

Run 1 on 127eeeb: PASS migrations, vuln, deps, setup-windows, build, mobile. FAIL:
- test-go-integration: `TestCommunitySchema` (fixed rollback walk of 16 steps, too short now that newer migrations exist). Real regression, fixed in 5796933, passes locally.
- test-go: `TestReady_ChecksPresent` (storage/ssr "fail" in the clean checkout under load); passes locally.
- lint, test-ts (vitest workers timing out), openapi (Schemathesis connection refused, app container never came up), lighthouse (`@lhci/cli` does not install on this host, a known problem): all from host contention.

Run 2 on 5796933 (`--only lint,test-go,test-go-integration,test-ts,openapi`): test-ts PASS (28 files, 97 tests). Lint went OOM ("cannot allocate memory"), and the race tests went OOM too (ThreadSanitizer "failed to allocate" in api/changes, appapi, backup, metrics; internal/search build failed; `./tests` timed out at 10 minutes). openapi never ran: `--only` without `build` waits forever on the build job (pitfall). The host is out of memory, so the gate is NOT green yet. Rerun the full `scripts/ci/local.sh` on a quiet host (fewer live QA stacks: four `hexascore-qa-cilocal*` app containers are still up from earlier runs) before any release PR. Lighthouse needs `bunx lighthouse@12.6.1` by hand.

## Open

1. Next REV rows, in order: REV-IMPL-mobile-v1-tokens first (packages/tokens v1 palette shared with web: check web build and AA contrast in both themes; ahead of any IMPL-web-v1), then REV-IMPL-mobile-v1-ui (stacked); REV-P6.A.2.1 and REV-P6.A.2.2 (stacked on P6.A.1.4); the coming deps row for expo-font, @expo/vector-icons, @expo/ui (owner yes Q-B3): exact versions compatible with Expo SDK 58. REV-FIX-catalog-audit-guard-row-level is done by the orchestrator (hotfix v0.12.3), do not merge it. Also: REV-P4.B.4.1 (5fb0dae, NFR-04-5 decision), REV-P4.B.4.2 (7f87764, stacked), REV-FIX-community-create-kinds-ui (b23b0a1), REV-P6.A.1.3 (stacked on P6.A.1.2, now merged). Check migration ids against main (latest 20261101120000).
2. P6.D.1.3 is not delivered yet; when it lands, its goldens and codegen regeneration are the specialist's on main (`scripts/openapi-check.sh --write` with DATABASE_URL on 5433, `bun run sdk:generate`, Python in `ghcr.io/astral-sh/uv:0.9.30-python3.13-bookworm-slim` with openapi-python-client 0.29.1, `go run ./cmd/app codegen:client -prefix /app/v1 -o -`).
3. Low findings, no rows: a concurrent double revert of one audit row answers 500 (unique index) instead of 409; `app/jobs/compute_trending.go` still forgets `trending.CacheKey` entries the home no longer reads.
4. FIX-www-redirect: the QA compose stack sets APP_URL=http://localhost:PORT, so Lighthouse/e2e at 127.0.0.1 see one 301 first.
5. SDK consumers on the old base URL `https://api.hexascore.app` (no /v1) get 404 on root paths; release notes must say the base URL is now `/v1` (owner call on SDK versioning).

## Pitfalls

- `git -C hexascore-vukem worktree add wt/...` resolves the path inside the repo: always pass an absolute `C:/dev/hexascore-app/wt/...`.
- `git checkout --theirs` on a lang JSON drops every key main added since the branch base; resolve only the conflict hunks (`git checkout -m` recreates them; markers then read `ours`/`theirs`).
- Merging stacks onto a moved main: home.go, Home/Index.tsx, registry.go, route_inventory.txt and bootstrap/providers.go conflict often; run `-run Inventory` after any route addition.
- `scripts/ci/local.sh --only` must include `build` whenever `openapi` (or any image job) is listed, or it hangs.
- The leftover directory `wt/senior-P4.B.2.7` could not be deleted (Permission denied; a process holds a file). Git already dropped the worktree, so only `rm -rf` remains.
- TASKS.md is dirty with orchestrator edits: my rows are edited in place but TASKS.md is left for the orchestrator to commit.
