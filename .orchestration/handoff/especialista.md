# Handoff: especialista (2026-10-31, session 4)

Reason: six reviews done; hand off early (predecessor reached 336k tokens). STATUS: HANDOFF.

## Done this session

| Row | Verdict | Merge on local `main` |
| --- | --- | --- |
| REV-FIX-covers-base-state r2 | CORRIGIDO_E_APROVADO (library + public profile JSON covers absolute 745feb6; golines 2750d0e; CoverThumb `lg` conflict) | 8bdfdf3 |
| REV-FIX-qa-port-isolation | CORRIGIDO_E_APROVADO (lighthouserc.cjs doc refs, AGENTS.md parallel-stack recipe); r2 (E2E_DB_NAME, community flag) | d99e757, 4c28b1e |
| REV-FIX-sdk-python-regen | APROVADO; P2.H.3.1 ticked | 0998e70 |
| REV-P3.A.4.1 | APROVADO (files not served; P3.A.4.2 waits Q-A11) | e6fcc24 |
| REV-P2.H.2.1 | CORRIGIDO_E_APROVADO (integration build tag, no-blank-frame e2e case, CSS conflict) | 720ec99 |
| REV-P4.A.3.1 (part 1) | CORRIGIDO_E_APROVADO (malformed draft id -> not found ad3b426); box stays open | 7bf0b7d |
| REV-DOC-vf-apiproblem-code | APROVADO (facts sampled) | 58b409a |

Gate on `main` @ 8bdfdf3 green (build, lint 0, bun lint/typecheck/vitest, go test integration). Gate @ 7bf0b7d: build, lint 0, bun lint/typecheck/vitest 73, go test integration green except `internal/search` TestAudienceIntegrationAPIOnlyAllowlisted "search timeout" under load (gate ran next to other stacks); the package passes alone (53 s). Not touched by these merges; if it recurs, open a FIX row. ROADMAP ticks: P2.H.3.1, P3.A.4.1, P2.H.2.1. CHANGELOG lines added for every merge.

## Next rows (in order)

1. REV-FIX-e2e-nav-timeouts: diff reviewed and fine (4ba9a2f); waits senior-b's r2 with the third clean run. Accept once the r2 section lands, merge, then rerun `profile.spec.ts` on a stack: P2.H.2.1's new CSR case failed in 2 of 6 projects only on the onboarding race this fix removes.
2. REV-P4.A.3.1b (2923481), then REV-P4.A.3.1c (4d346a9); merge in order. 3.1b must take the draft target from the stored row (see the P4.A.3.1 review note). FIX-community-problem-codes follows 3.1c.
3. Mobile stack: REV-P5.A.1.1 -> P5.A.1.2 (with FIX-mobile-nav-test-timeout) -> P5.A.2.1 -> P5.A.3.1.
4. Report-only rows: REV-DOC-i18n-legal, REV-QA-mobile-proto-review, REV-QA-i18n-parity, REV-QA-e2e-main, REV-DOC-roadmap-p7-refs (may apply the ROADMAP patch), REV-DOC-research-letterboxd (partial, Q-A9), FW-triage-branches (push blocked on Q-A7), REV-FIX-i18n-untranslated r2.

## Pitfalls

- Parallel QA stack (now on main, AGENTS.md "Local QA"): `createdb hexascore_qa_esp`, `E2E_IMAGE_TAG=<tag> QA_PORT=3110 QA_DATABASE_URL=postgres://hexascore:hexascore@host.docker.internal:5433/hexascore_qa_esp?sslmode=disable`, `docker compose --env-file C:/dev/hexascore-app/hexascore-vukem/.env -p hexascore-qa-esp --profile app build app-init` / `run --rm --no-deps app-init` / `up -d --wait --no-deps --no-build app`. e2e: `E2E_BASE_URL=http://127.0.0.1:3110 E2E_APP_CONTAINER=hexascore-qa-esp-app-1 E2E_DB_NAME=hexascore_qa_esp bun run e2e <spec>`. Run `bun install --frozen-lockfile` in a fresh worktree first; bare `bunx playwright` pulls `playwright@latest`. The esp stack is stopped; db `hexascore_qa_esp` still exists (drop or reuse).
- Gate script: scratchpad `gate.sh <sha>` uses `wt/especialista-gate` (detached). A stray untracked file there blocks checkout; remove it first.
- The board helper someone ran replaced shas in REV notes with `debd454`; I restored P3.A.4.1 (fce88fc) and P4.A.3.1 (296a3db). Re-check a REV note's sha against the report before trusting it.
- Vitest counts differ between the main checkout (21 files) and the clean gate worktree (17 files); the gate worktree is the reference.
- Worktrees to remove after `merged`: `wt/especialista-{FIX-covers-base-state,FIX-qa-port-isolation,P2.H.2.1,P4.A.3.1}` plus the session-3 list.
