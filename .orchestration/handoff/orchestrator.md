# orchestrator handoff (2026-10-31, end of the first five-role session)

Start the next session with `PROMPT-WORKERS.md` section 1.1 (model, parallel-workers yes, services, OpenPencil MCP registration, launch at least three implementers plus the specialist).

## State

- Releases: `v0.12.0` (PR #60, squash 219c873) and hotfix `v0.12.1` (PR #61, b6ddba4, cover prop on item pages) are on `origin/main` and deployed on Railway production. Local `main` contains both (merge e59c437) plus later merges; it is ahead of `origin/main` and must only reach GitHub through the release flow (`PROMPT-WORKERS.md` 8.1).
- Production: bucket `hexascore-production` wired into the `app` service via `S3_*` (owner-approved 2026-10-31); covers return 200 on home, listings and item pages.
- CI: the Playwright job is skipped unless the repo variable `RUN_E2E` is `true` (owner decision); e2e gates nothing.
- vukem: PRs #104 (OTLP), #105 (rate-limit burst), #106 (pgtest drop timeout), #107 (Inertia HTML error page) merged; #109 (CSR shell) and others in review by the specialist. No vukem tag since v0.34.0; app `go.mod` still on v0.34.0 (adoption rows after the owner asks for a vukem release).

## Workers at hand-off time (check heartbeats; relaunch with the role brief + handoff)

- `especialista`: vukem REV rows (AUDIT-3 PR #109, AUDIT-5, docs font size, Schemathesis with a real Docker run), post-merge review of `docs/design-system.md`.
- `senior-a`: community track (FIX-community-problem-codes, P4.A.3.2, P4.A.3.3, P4.A.3.1d, then P4.B).
- `senior-b`: QA-ci-local-script (`scripts/ci/local.sh`), then Android P5.A.3.2+.
- `senior-fw`: vukem `fw:` rows (FW-last-updated-check, AUDIT-2, AUDIT-4, P2.H.3.1 enum, P5.A.3.1 ApiProblem.code, P4.A.3.1b flags, covers ETag).
- `designer`: not launched yet. Needs the OpenPencil MCP (section 1.1 step 4) or the OpenPencil-hosted agent; brief `.orchestration/briefs/designer.md`; outputs in `docs/openpencil/v1-editorial`, `v2-noir`, `v3-paper`; references `docs/design-system.md`, `docs/prompts-concept-site-pages.md`, `legacy/docs/design-system.md`.

## Update at session end (2026-10-31, later)

- vukem merged: #109 (CSR shell), #110 (admin custom pages), #111 (SSR cache docs, VF withdrawn), #112 (docs +25% font), #113 (native `api:fuzz` Schemathesis, digest-pinned), #114 (last-updated squash check). Open, CI green, not yet reviewed: #115–#120 (app name default, hydrating hook, doc-only enum, ApiProblem.code, flags docs, S3 ETag test). Each needs origin/main merged in after every squash.
- App local `main`: P4.A.3.2, P4.A.3.3 (boxes unticked until QA-community-e2e), FIX-community-problem-codes, QA-ci-local-script (`scripts/ci/local.sh`), route inventory fix; full gate green after the batch.
- `wt/` cleaned: only active or unmerged worktrees remain; no loose files.
- Next launches (never idle, three implementers + specialist): senior-a → P4.A.3.1d, FIX-community-create-ux, QA-community-e2e, P4.B; senior-b → Android P5.A.3.2+, FIX-mobile-login-test-timeout, FIX-ssr-supervise-flake; senior-fw → whatever `fw:` rows appear after #115–#120 land (none queued now; take app rows if empty); designer → OpenPencil three versions once the MCP is registered; especialista → vukem #115–#120.

## Owner items open (`docs/orchestration/QUESTIONS.md`)

Q-A5 (Phase 7 feed LGPD basis, legacy follows), Q-A7 (resolved by the specialist pushing vukem PRs), Q-A8 (second TypeScript version for the TS SDK generator), Q-A9 (Letterboxd human read, low), Q-A10 (workspace link for `packages/app-client`), Q-A11 (`popularity` in bulk ID exports, blocks P3.A.4.2), Q-A12 (keep legacy English labels in pt-BR/es, low), Q-A13 (osv GO-2026-5932 in x/crypto turns CI `deps` red), Q-A14 (vukem and `drivers/s3` release for the covers ETag). Owner-only cleanup: none pending for `.env` copies once their worktrees are removed.

## Rules added this session (also in `PROMPT-WORKERS.md` / `AGENTS.md` / memory)

- Never idle: at least three implementers in parallel plus the specialist.
- Release gate: all CI jobs except e2e locally before a release PR; squashed release commit; the owner runs the production merge.
- Isolated QA stacks per worker; never `down` project `hexascore`; never copy `.env`.
- Implement first, long checks at the end (section 8.2): no waiting on CI or long suites per task; vukem senior-fw pushes its batch and opens the PRs, fixes GitHub CI; the specialist reviews and squash-merges.
