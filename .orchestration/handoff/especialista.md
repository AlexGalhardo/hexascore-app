# Handoff: especialista (2026-11-01, session 18)

Reason: past the ~12 review budget (32 rows in five batches). STATUS: HANDOFF.

## Done this session (local `main`, no push, no tag)

| Rows | Verdict | Merge |
| --- | --- | --- |
| senior-fw P6.A.1.3, 1.4, 2.1, 2.2, 2.3, 3.1, 3.3, FIX-ingest-trim-titles, 3.2 (rework), 3.4 (rework), 3.5a/b/c, 4.1, FIX-ingest-runner-wiring, FIX-wikidata-selection-cache, FIX-wikidata-refs-memo, FIX-ingest-override-seam, P6.B.1.1, FIX-ingest-incremental-since, P6.B.3.2, P6.B.3.1 | 1.3/1.4/2.1/2.2/2.3 CORRIGIDO_E_APROVADO (74c099a5: cursor/status/failed re-read under FOR UPDATE + guarded running update; only the breaker probe closes it; resume whose QID left the selection restarts; circuit open / cancel / host-not-allowed abort the chunk instead of parking; SSRF redirect test; ingest env lines), rest APROVADO; route inventory gets `/.well-known/assetlinks.json` (red on main since P5.B.5.1) | 7aaf5d7d; ROADMAP P6.A.1.3-P6.A.4.1 and P6.B.3.2 ticked (P6.B.1.1 waits P6.A.4.2, P6.B.3.1 partial) |
| ADOPT-tanstack-query-web, PERF-web-lcp, FIX-lighthouse-budgets-main | APROVADO; Nav.tsx single lazy palette kept; 715afd11 moves the bun:test SDK workflow tests out of Vitest (`bun run test` was red on main since P3.B) | f0ab6804 |
| senior-a P7.A.1.1, 1.2, 1.3 (part 1) | CORRIGIDO_E_APROVADO: social ADR renumbered 0021 (index, SDD 14, brainstorming, code, ROADMAP refs), open-issues rows 14-8..14-11, `FLAG_SOCIAL=false`, follow/feed routes in the inventory, golines | cdc7a77b; P7.A.1.1 and 1.2 ticked |
| IMPL-mobile-v1-status-ui, -account-ui, -search-i18n | APROVADO / CORRIGIDO (tabsLayout.test needs a QueryClient for `useUnread`) / CORRIGIDO (copy-catalogues keeps `entity` + `itemTypes`) | 1dcc6c3e |
| FIX-shelf-score-label | APROVADO | 9102d379 |
| Housekeeping | Dependabot: `/sdks/ts` in its own bun entry with a typescript major ignore (root and mobile are on TS 7) | 892b20f7 |

Checks per batch are in each report. TASKS.md rows set to `ready` on disk, not committed (file carries other agents' edits). `.env.example` uncommitted RESEND/TELEGRAM/LEGACY/BLOB lines from other workers were saved and re-applied after each merge.

## Open, in order

1. Gate: `scripts/ci/local.sh` on main in halves. Half 1 on 418ea358: migrations and vuln PASS; lint, test-go, test-go-integration, test-ts FAIL with real regressions, all fixed in caaa0217 (merge of agent/especialista/FIX-gate-v0.13): `TestSearchSQLConfined` (P6.B.3.2 used `hs_norm` outside internal/search; now `search.EntityNameMatch`), `TestSocialStage` (expected no `follows` table; P7.A.1.2 created it, now expects 0 rows), gosec/golines in `scripts/parity/capture/main.go`, lazy-palette `Nav.test.tsx` timing under load (5 s wait). The lint job also hit a golangci "Timeout exceeded" under host load. Rerun on caaa0217: test-go, test-go-integration, test-ts PASS; lint stopped again on the golangci 5 m config timeout under host load, so its steps were run by hand on caaa0217: `golangci-lint run --timeout 15m -j 2 --build-tags integration ./...` 0 issues, `fmt --diff` clean, `go vet` ok, `openapi-check.sh` rc 0, `openapi:lint` valid, Biome/tsc ok, markdownlint clean on tracked files (only untracked `qa/e2e/test-results/` and `questions-to-dev.md`), commitlint v0.12.3..HEAD fails only on 7e743592 (footer > 100; history rewrite needed, asked in questions-to-dev.md item 19; v0.12.0 used a squashed release commit). Half 2 on caaa0217: deps, setup-windows, build, mobile PASS; openapi FAIL only at Schemathesis (`curl: (52) Empty reply` on the app container, then connection refused: the known Docker Desktop limitation, RELEASE-v0.12.0-gate.md line 27; goldens, Redocly and app-client codegen steps passed). Lighthouse not run (Kaspersky on this host). Still owed before the release PR: Schemathesis (api-v1, app-v1) and k6 against a published port or on Linux, and Lighthouse in the `lhci-runner:local` container or on ubuntu CI (TBT decides there).
2. Orchestrator 2026-11-01, not started: (a) ROADMAP-gap-audit class (a) rows P2.F.2.2, P2.H.2.2, P3.B.2.1, P3.B.3.1, P5.A.3.4, P5.B.4.1, P5.B.4.3, P5.B.5.1, P6.B.1.1, P6.B.3.1, P7.A.1.3: tick only those whose acceptance is fully met on main, note owner-only gates inline (`.orchestration/reports/ROADMAP-gap-audit.md`); (b) REV-P2.H.3.4 (indigo -> primary tokens); (c) REV-P3.A.4.2 (`listExports`/`getExport`: `popularity` stays out of bulk exports until owner Q-A11; activity operation ids split into four literal routes, URLs unchanged: confirm, regenerate the Python SDK in the pinned uv container, CHANGELOG Changed line).
3. impl-web: REV-IMPL-web-feed-nav, REV-IMPL-web-social-v1, then REV-FIX-nav-lg-width (stacked).
4. senior-a P7.A.1.3b when it reaches review (flag default off while Q-A5 is open).
5. REV-P4.B.5.1 still `todo`.
6. New row FIX-ingest-retry-after-timeout (60 s client timeout wraps the Retry-After pause).
7. REV-FIX-seed-flag: do not merge (hotfix v0.12.4, PR #64).

## Pitfalls

- Mobile jest: run `bun run test -- --runInBand --forceExit` in `apps/mobile` (it goes through `scripts/run-jest.cjs`); plain `bunx jest` fails with a missing babel plugin.
- Do not `bun install` inside a worktree whose `node_modules` is a junction to main: it rewrites main's install.
- Worktrees left for their workers (clean, branches merged): senior-PERF-web-lcp, senior-a-P7.A.1.2/1.3, impl-mobile-item/auth/search, impl-web-search-palette (holds FIX-shelf-score-label).
- markdownlint from the repo also lints `C:\dev\hexascore-app\questions-to-dev.md` (not ours).
