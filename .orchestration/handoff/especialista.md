# Handoff: especialista (2026-11-01, session 17)

Reason: 15 rows reviewed and merged (past the ~12 review budget). STATUS: HANDOFF.

## Done this session (local `main`, no push, no tag)

| Rows | Verdict | Merge |
| --- | --- | --- |
| P5.B.4.2, P5.B.4.3, P5.B.2.4, P5.B.2.5, P5.B.4.1, P5.B.5.1 (senior-a, stack order) | 4.2/2.4/2.5/4.1 CORRIGIDO_E_APROVADO, 4.3/5.1 APROVADO; integration fixes 09abc257 (status/time-to-beat buttons in the restyled ItemDetail, `AppItemJSON` in Entity, `listLaneResult` takes `() => AppLocale`, entity route test wrapped in `offline()`) | 008d7739; ROADMAP P5.B.4.2, 2.4, 2.5 ticked (4.3, 4.1, 5.1 left open: Maestro/device acceptance) |
| IMPL-mobile-v1-search-wire, IMPL-mobile-entity-more | CORRIGIDO_E_APROVADO (navigation.test conflict) / APROVADO | cd7b53e8 |
| P3.B.1.1, 1.2, 1.3, 2.1, P3.B.3.1, FIX-problems-redirect-301, FIX-docs-next-cursor (senior-up) | APROVADO except P3.B.3.1 CORRIGIDO (Python webhook snippet skips malformed base64, 6eacf461; build-docs.test.ts conflict) | b463f045; ROADMAP P3.B.1.1-1.3 ticked (2.1 owner TestPyPI run, 3.1 staging/e2e left open) |

Checks: mobile tsc, biome ci, full mobile jest `--runInBand --forceExit` (55 suites, 352 tests) on the batch tips; `TestAppMe`, `TestAssetLinks`, `TestProblemCatalog`; golangci on appapi/wellknown/routes 0 issues; sdks/ts typecheck + test, `sdk:generate` no diff; `go build ./...` on main ok. Not run: Python SDK pytest (no `uv` on this host), full gate.

`.env.example`: the main checkout's uncommitted lines from other workers (RESEND/TELEGRAM/LEGACY/BLOB) were saved, P5.B.5.1 merged, and they were re-applied uncommitted. TASKS.md rows set to `ready` on disk, not committed (carries other agents' edits).

## Open, in order

1. senior-fw P6.A stack (REV-P6.A.1.3 ... P6.A.4.1, FIX-ingest-trim-titles, FIX-ingest-runner-wiring) with the session 14 verification list (1.3 continuation lock/guard, 2.1 resume skip, 1.4 breaker probe and timeout, 2.2/2.3 infra errors counted as target failures, 2.2 unbounded refs cache, 1.4 thin SSRF tests). Then P6.B: FIX-wikidata-selection-cache, refs-memo, override-seam, P6.B.1.1, incremental-since, P6.B.3.2 (check its migration id no longer collides with 20261101130000), P6.B.3.1.
2. impl-web ADOPT-tanstack-query-web is `review` (no REV row yet): pinned `@tanstack/react-query` 5.105.0 (owner approved) + ADR.
3. senior-a P7.A.1.1, P7.A.1.2 are `review` (flags `social`/`feed_opt_in` default off); P7.A.1.3 doing.
4. senior-up PERF-web-lcp / FIX-lighthouse-budgets-main still `doing`.
5. P4.B.5.1 REV row still `todo` (not in this session's queue).
6. At the end: `scripts/ci/local.sh` once on `main`, alone (next step v0.13.0 release).
7. REV-FIX-seed-flag: do not merge (hotfix v0.12.4, PR #64).

## Follow-ups noticed

- `+native-intent.tsx` still drops the deep-link locale (row FIX-mobile-deeplink-locale exists).
- Dependabot (`js` group covers `/` and `/sdks/ts`) will propose typescript 7 for sdks/ts again; hey-api 0.99.0 needs the TS 6 compiler API.
- EntityPage "Next" control (IMPL-mobile-entity-more) is not wired yet: P5.B.2.5 shows the first page only.

## Pitfalls

- Mobile catalogue conflicts: take either side, then `bun run --cwd apps/mobile catalogues`; `resources/lang/*.json` 3-way merge with a JSON deep merge (format is `json.dumps(indent=2, ensure_ascii=False)`).
- Worktree node_modules: junctions for `node_modules` and `apps/mobile/node_modules`; `cmd //c rmdir` them before `git worktree remove`.
- markdownlint-cli2 from the repo picks up `C:\dev\hexascore-app\questions-to-dev.md` issues; not ours.
