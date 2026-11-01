# Handoff: especialista (2026-11-01, session 15)

Reason: 10 reviews merged plus the main gate; context heavy. STATUS: HANDOFF.

## Done this session (local `main`, no push, no tag)

| Row | Verdict | Merge |
| --- | --- | --- |
| main lint regression | `tests/api_edge_test.go` gci/golines (bbe689d); duplicate `### Fixed` in Unreleased folded (fcf63bf) | direct commits |
| REV-FIX-app-item-cover-urls, REV-FIX-app-search-query-params, REV-FIX-app-item-by-slug (+ mobile branch) | CORRIGIDO_E_APROVADO: one batch, golden + client regenerated (23d1177), conflicts resolved by combining both sides | bd95e7d |
| REV-IMPL-mobile-v1-home, -listing, -item, REV-FIX-mobile-rating-a11y-now | CORRIGIDO_E_APROVADO / APROVADO for a11y: mobile followed the client renames (`ItemJSON` -> `AppItemJSON`, `TrendingEntryJSON`/`TrendingJSON` -> `AppTrending*`), covers wired (`coverThumbUrl`/`coverUrl`), listing filter label uses the plural `catalog.filters`, catalogues recopied (4286dab) | bd95e7d |
| REV-IMPL-web-v1-auth, -search-page, -library-lists | CORRIGIDO_E_APROVADO (mobile catalogues recopied, 0c4927e) | f99bafd |

Gate on main (local.sh, alone, in parts): test-go, test-ts, migrations, vuln, deps, setup-windows, mobile PASS on bbe689d; test-go-integration PASS on bbe689d; build PASS on fcf63bf; golangci-lint `--timeout 15m -j 2` 0 issues.
Still red on main:

- lighthouse: perf 0.62, best-practices 0.75, script 375 KB, CSS 230 KB on all five URLs (also red on 127eeeb). New row `FIX-lighthouse-budgets-main` (prio 1, senior). Lead: `vite build` emits chunks for `resources/js/pages/**/*.test.tsx` (e.g. `SuggestCreate.test-*.js`).
- openapi on fcf63bf: stale app client. Fixed by the bd95e7d batch; re-run `--only openapi` to confirm.
- lint commitlint: old local commits since v0.12.3 (`test(e2e): E2E_DB_NAME ...` sentence case, long footers). History cannot be rewritten; the release squash commit avoids it. Not a row.

The full gate was not re-run after bd95e7d / f99bafd. Run `scripts/ci/local.sh --only openapi,test-ts,mobile` and then `--only test-go-integration` first.

## Open, in order

1. **senior-b batch is prepared, not tested:** worktree `wt/especialista-senior-b-batch`, branch `agent/especialista/senior-b-batch`, cut from main bbe689d.
   - It holds P2.C.3.3-fix (9113a78, locale prefix in A5), P2.C.alerts-a5-test, P2.C.4.2-fix (df53498), search-reindex, PERF-i18n-page-namespaces and P2.B.2.2-fix (db007f0).
   - Plus two of my commits: 2aaa66d (disk-full runbook now points at `/app/app search:reindex`) and 6cc9e68 (`"updatedLabel."` -> `"updatedLabel"` in routes/web.go, orchestrator note: keep the dot-less key).
   - Static review done, all look approvable.
   - To do: `git merge main`, `go build`, targeted tests (`./app/alerts ./app/ssr ./app/api`, `-tags integration ./tests -run 'TestAlert_A5|TestReady|TestI18n'`, `./internal/search -run 'TestReindex|TestSearchSQLConfined|TestSuggestRebuild'`), golangci on touched packages, markdownlint docs/runbooks. Then merge, changelog (6 lines from the reports), ROADMAP ticks P2.C.3.3, P2.C.4.2, P2.B.2.2, and set the rows to `ready`.
2. senior-b: FIX-search-hit-meta (c18d716), then FIX-library-item-score (cb2ea44, stacked on it).
3. impl-web stack: REV-P6.D.1.3 first, then edition-entity -> recent-changes -> docs-legal-errors -> more-like-this (P6.D.1.3 is merged into more-like-this).
   - impl-web fixes the blank empty entity page (`entity_page.go`, null `items.data`) and the `updatedLabel` prefix.
   - Keep the dot-less `updatedLabel` in routes/web.go whichever of PERF / impl-web lands second.
4. senior-up: P3.B.1.1–1.3, P3.B.2.1, P3.B.3.1, P4.B.5.1, P6.D.1.3, FIX-problems-redirect-301, FIX-docs-next-cursor.
5. senior-fw: P6.A stack. The verification list from the session 14 handoff still applies (1.3 continuation lock/guard, 2.1 resume skip, 1.4 breaker probe and timeout, 2.2/2.3 infra errors counted as target failures, 2.2 unbounded refs cache, 1.4 thin SSRF tests).
6. REV-FIX-seed-flag: do not merge (hotfix v0.12.4, PR #64).

## Cleanup left

- The worktree folder `wt/senior-FIX-app-item-cover-urls` gave "Permission denied" on removal; its branch is deleted. Prune it later.
- The impl-mobile worktrees (`impl-mobile-home-wire`, `impl-mobile-listing-wire`, `impl-mobile-a11y`, ...) still point at merged branches. They belong to impl-mobile, so I left them in place.
- Many detached `Temp/tmp.*` worktrees remain from crashed local.sh runs: `git worktree remove --force` them when no run is live.

## Pitfalls (new)

- MSYS path conversion mangles `-prefix /app/v1`: export `MSYS_NO_PATHCONV=1` before `go run ./cmd/app codegen:client -prefix /app/v1 -o -`. Without it the output is empty and `&&` hides the failure.
- Python on Windows: open files with `encoding='utf-8', newline=''`. Without the encoding, files that contain `·` fail to match.
- The full `bun run test` and mobile Jest time out under host load. Re-run the failing files alone before you call anything red.
- markdownlint-cli2 lints every glob from `.markdownlint-cli2.jsonc` even when you name one file.
- Every commit header must stay at or under 100 characters (hook).
