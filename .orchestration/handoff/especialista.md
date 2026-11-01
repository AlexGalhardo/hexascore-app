# Handoff: especialista (2026-11-01, session 19)

Reason: 15 reviews done (budget). STATUS: HANDOFF.

## Done this session (local `main`, no push, no tag)

| Rows | Verdict | Merge |
| --- | --- | --- |
| IMPL-web-feed-nav, IMPL-web-social-v1, FIX-nav-lg-width, FIX-nav-mobile-hamburger (one stack) | CORRIGIDO_E_APROVADO: main merged in (web.go imports; Nav.tsx keeps the hamburger sheet plus the lazy palette; tests from both sides); focus-trap selector covers every sheet control | cc9d6a13 |
| Gap audit class (a) | ticked P2.H.2.2, P6.B.1.1 (local acceptance re-run); `Status (2026-11-01)` lines on P2.F.2.2, P3.B.2.1, P3.B.3.1, P5.A.3.4, P5.B.4.1, P5.B.4.3, P5.B.5.1, P6.B.3.1 (owner gates) | 1b461017 |
| P7.A.1.3b, QA-social-e2e, P7.B.1.1, P7.C.1.2, P7.B.1.2, P7.B.1.3 (one stack) | CORRIGIDO_E_APROVADO: reviews migration renamed `20261101171000` (collided with main's `20261101170000`), merge conflicts, moderation notices translate `targetType`; ticked P7.B.1.2, P7.B.1.3, P7.C.1.2; P7.A.1.3 (k6) and P7.B.1.1 (Playwright) open with status lines | a9343e81 |
| P2.H.3.4 | APROVADO, ticked | f45e8a09 |
| FIX-reco-log-events | CORRIGIDO_E_APROVADO (conflict with P7 in item_controller) | d7a8c4e7 |
| P3.A.4.2 | CORRIGIDO_E_APROVADO: `popularity` dropped from the export files while Q-A11 is open (job, test, SDD 05, public doc + docs.json); Python SDK regenerated in `ghcr.io/astral-sh/uv:0.9.30-python3.13-bookworm-slim`; activity split accepted; ticked | b9d6a66a |
| P4.B.5.1 | APROVADO, box open (Playwright, k6, ZAP runs) | 8ac75c84 |
| CLEAN-stale-vukem-workarounds | CORRIGIDO_E_APROVADO: without TypeGuard `getExport` answered 422 on an unknown type, now 400; edge row for `/films/{id}/activity` is 404 (literal routes) | 614a9442 |

TASKS.md rows set to `ready` on disk, not committed (the file carries other agents' edits).

## Gate (half 2 on 3580956e)

deps PASS, mobile PASS (55 suites, 362 tests), setup-windows PASS (go test ok), build PASS when rerun alone (89 s; the first run under load failed the `/up` smoke with "Empty reply"). openapi: the parallel run had 20 Schemathesis "Network Error" (container unreachable under load); NOT re-verified: `scripts/ci/local.sh --only openapi` hangs forever in `need_build` because the build job is not selected (script bug: use `--only build,openapi`, or make `need_build` fail fast when build is not selected). Half 1 jobs were last green on caaa0217 except the rerun status. The full gate on current main is still to run (step 5).

## Open, in order

1. REV-FIX-P7.B.1.3-followups @ 2a7b4d63 (flag-gated `POST /reports` targets, three-locale test, `content_unhidden`; edits unreleased migration `20261101190000` in place; conflict with CONFIG-email-addresses in `routes/reports.go`). Old `agent/senior/P7.B.1.3` @ 9df20e90 is superseded: do not merge. Then REV-P7.C.1.1 @ 1e41ffc9.
2. senior-fw on `agent/senior/P6.B.3.3`: REV-P6.B.4.1 (its migration `20261101190000` collides with `20261101190000_moderate_reviews_and_comments` now on main: rename at merge), P6.B.3.3, 4.2, 4.3, FIX-ingest-qids-param @ 2d792c9f, P6.B.4.4 and 4.5 @ f785355b. Ruling for 4.5: tick with a note that the OTel metrics, dashboard and `a7.yaml` are replaced by `alert=A7` log lines per D-35; steam/gog/igdb schedules and "breaker open > 1 h" move to P6.B.1.2, P6.B.1.3, P6.B.2.2. Judge: rollback emits no `api_changes`/`ItemsChanged` for deleted items; `tmdb.expire` deletions are not restored; prune and retry are scheduler functions, not queue jobs. Then FIX-ingest-retry-after-timeout.
3. senior-b CONFIG-email-addresses (owner approved 2026-11-01: `noreply@` automatic mail, `support@` support inbox and Reply-To).
4. REV-FIX-seed-flag: do not merge (hotfix v0.12.4, PR #64).
5. Queue drained: full gate on `main`, alone, in halves; then the orchestrator prepares the v0.13.0 release PR.

## Pitfalls

- `.githooks/commit-msg` rejects a non-Conventional merge message, and `git merge -q main -m "merge main into X"` then leaves the merge uncommitted silently. Use `chore(<scope>): merge main into <branch>`.
- Delete the `node_modules` junction (`cmd //c rmdir <path>`) before `git worktree remove`; `bun run lint` is the right Biome call (`biome ci .` hits the mobile nested root).
- Mobile jest: `bun run test -- --runInBand --forceExit` in `apps/mobile` (main checkout; worktrees have no mobile node_modules).
- About 30 leftover detached `ci-local` temp worktrees were being removed by a background loop; run `git worktree prune` and check `git worktree list` next session. Old `wt/especialista-*` folders from earlier sessions are not registered worktrees.
- markdownlint from the repo also lints `C:\dev\hexascore-app\questions-to-dev.md` (not ours; it has MD022/MD032 errors).
