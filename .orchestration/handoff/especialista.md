# Handoff: especialista (2026-10-31, session 7)

Reason: the merge queue is done, six reviews are done, and context is heavy. STATUS: HANDOFF.

## Done this session

Merge queue (reviewed earlier, merged now; touched gate after each):

| Row | Merge |
| --- | --- |
| P4.A.3.1c (ticks P4.A.3.1) | 34efea6 |
| FIX-nojs-auth-post | dc99c4d |
| DOC-sdd-flag-names | d21e0db |
| FIX-scoring-burst-clock (own fix, report written) | b500e2c |
| P5.A.1.1 (catalogues auto-merged, regeneration gave no diff) | 25a7ee5 |
| P5.A.1.2 + FIX-mobile-nav-test-timeout; P5.A.2.1 has the same head 395bd7c | 1e3c970 |
| P5.A.3.1 | 774b3a8 |
| DOC-roadmap-p7-refs patch, items 1-10 | e3b8f0d |
| QA-e2e-speed | `blocked`: nothing delivered (no report, no REV row, 0 commits; uncommitted work in wt/senior-QA-e2e-speed) |

Full gate on main e3b8f0d: build, golangci (integration tags) 0 issues, biome, tsc, vitest 79, `go test -tags integration ./...` all ok.

Reviews:

| Row | Verdict | Result |
| --- | --- | --- |
| REV-P2.H.1.4 | APROVADO | merged c83174d |
| REV-P2.H.1.5 | APROVADO | merged b2c50c6 |
| REV-P2.H.3.5 | CORRIGIDO_E_APROVADO (Accept-Language in the render cache key on unprefixed pages, c731fc6) | merged 1915f77 |
| REV-VF-NEW-AUDIT-6 | APROVADO | vukem PR #105, squash 6b69a323 (33/33) |
| REV-VF-NEW-FIX-gate-timeouts | APROVADO | vukem PR #106, squash cb6a6f4a (33/33) |
| REV-VF-NEW-AUDIT-1 | APROVADO | vukem PR #107, squash 8fbe0f2a (33/33) |

`vukem-feedback.md` status for all three: `fixed on main, unreleased (PR #n)`. No vukem tag or release.

## Open

1. NEXT: REV-VF-NEW-AUDIT-3 (fix/vf-inertia-csr-shell, bf0036c7): the diff reads fine, but it conflicts with AUDIT-1's `renderRoot` extraction in `inertia/inertia.go`. AUDIT-1 is merged (#107), so merge origin/main into the branch and move the shell into `renderRoot` (error pages with ErrorFullPages then also get the shell, which is fine).
2. Then any new senior or `fw:` REV rows.
3. Proposed rows for the orchestrator:
   - vukem (prio 2): `check:last-updated` fails on `main` after every squash merge that touches docs: the squash commit is newer than the dates the branch wrote. Every docs PR has to merge origin/main and rerun `node scripts/last-updated.ts` (done on #105). Fix it in the checker (accept the entry when the page content equals the branch commit) or regenerate on main after the merge.
   - From session 6, still open: a search timeout answers 500 (FIX-search-timeout-500 exists); the S3 disk has no ETag (VF-NEW-FIX-covers-route exists).

## Pitfalls

- golangci-lint has a global lock. Another worker's lint makes vukem `task ci` exit 201 ("parallel golangci-lint is running"). Scratchpad `fwci.sh <wt dirs...>` retries on that error; for app lint, pass `--allow-parallel-runners`.
- Someone left `tsconfig.json` modified and `resources/js/vite-env.d.ts` untracked in the main checkout. That is not mine and I did not touch it. Run the gate in wt/especialista-gate, never in the main checkout.
- Other agents commit to main between my merges (vukem-feedback). Always `git add` named files.
- Worktrees especialista-{P2.H.3.5,P4.A.3.1c,FIX-nojs-auth-post,FIX-scoring-burst-clock} are unregistered, but their folders remain (Permission denied, a file lock). wt/especialista-P5.A and wt/fw-VF-NEW-AUDIT-6 were not removed. Delete them when nothing holds them.
- Stack esp is not running. DB hexascore_qa_esp still exists.
