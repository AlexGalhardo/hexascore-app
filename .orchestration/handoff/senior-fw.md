# senior-fw handoff (2026-10-31)
Delivered (review): FW-otlp-pr (e3eca299), VF-NEW-AUDIT-6 (fix/vf-ratelimit-burst bcbfbd04), VF-NEW-FIX-gate-timeouts (fix/vf-pgtest-drop-timeout d88f52b2), VF-NEW-AUDIT-1 (fix/vf-inertia-html-error-page 742b3e51). Reports in reports/.
Remaining todo (prio order): VF-NEW-AUDIT-3, AUDIT-5, P2.H.3.5 (design note first; QUESTIONS.md if API large), then AUDIT-2, AUDIT-4, P2.H.3.1, P5.A.3.1, P4.A.3.1b (docs only OK).
Recipe: worktree wt/fw-<ID> from $FW main; test first; `task ci` (redirect to wt/ci-<x>.log; retry if "parallel golangci-lint is running", other workers lint; use --allow-parallel-runners for ad hoc lint); after commit run `cd site && npm run last-updated` then `git commit --amend --no-edit` when docs pages changed; `task api:check`.
Pitfalls: Windows CRLF in some files (edit with newline='' in python); task ci takes ~12 min; worktrees stay until merged.
