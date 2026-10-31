# Handoff: especialista (2026-10-31, session 10)

Reason: ten reviews in this session (owner limit about 10). STATUS: HANDOFF.

## Done this session (local `main`, no push, no tag)

| Row | Verdict | Merge |
| --- | --- | --- |
| REV-P5.A.3.3 | APROVADO (d5edc4e), P5.A.3.3 ticked (Maestro staging run pending) | 8faf53c |
| REV-P5.A.3.4 | APROVADO (252dee7); box open until the Google picker runs on a device | 480ea87 |
| REV-P5.A.3.5 | CORRIGIDO_E_APROVADO (in-flight rotate guard 5251e94); box open until FIX-mobile-root-wiring routes 426 | 9ed1f1d |
| REV-P6.C.1.2 | CORRIGIDO_E_APROVADO (flush upsert sargable 69279ca, golines b14a8fd), ticked | daf9173 |
| REV-FIX-moderator-2fa | CORRIGIDO_E_APROVADO (SDD 13 FR-13-5 names moderators) | 3e0cebb |
| REV-P4.B.1.2 | APROVADO, ticked | 723955b |
| REV-P4.B.1.3 | CORRIGIDO_E_APROVADO (uuid guards, bulk version by uuid c172a32), ticked | cf86ba3 |
| REV-DESIGN-round-1 | APROVADO (0fb8e59, docs only) | 2b84e92 |
| REV-P6.C.1.3 | APROVADO, ticked; new row FIX-trending-rollup-indexes (prio 2) | 8b84c6c |
| REV-P6.C.2.2 | APROVADO, ticked; bench 10.2 s / 30 s; a14.yaml plausible, uid to confirm | 27e0ab7 |

Rows added or extended: FIX-trending-rollup-indexes (new), FIX-mobile-root-wiring (note: translate the "Device name" label and the "Update required" title).

Gates on main after each merge (targeted): mobile lint/typecheck/test, `go test ./tests/i18n`, moderation and recommend integration tests, golangci with the integration tag on touched packages and `./tests/`: all green. No full gate this session: run one on main before the next release.

## Open

1. NEXT: REV-P3.A.1.4 (+P6.C.3.2) at f70b11c, stacked on P6.C.2.2 (now merged). After the merge (orchestrator): regenerate `packages/app-client/src/api.ts` with `codegen:client` from a built image as `ci.yml` does, `sdks/ts` (`bun run sdk:generate`) and `sdks/python` (pinned uv container); the author's Windows codegen truncated the file and was reverted. OpenAPI goldens changed: run `scripts/openapi-check.sh` and `bun run openapi:lint`.
2. Other REV rows in the queue: REV-P4.B.2.1 (6618eec) and anything new.
3. The orchestrator is cutting the v0.12.2 search hotfix from `origin/main` in `wt/orch-hotfix-v0122`: do not touch `release/*`.

## Pitfalls

- Recheck `git log main..<branch>` against the reviewed SHA before every merge.
- `golangci-lint` refuses to run while another worker's instance is running ("parallel golangci-lint is running"): retry in a loop with a 20 s sleep.
- `git worktree remove` can fail with "Permission denied" on a locked folder: the worktree is unregistered anyway; delete the branch and `rm -rf` the folder later.
- Senior worktrees are reused by the designer for new branches (`wt/designer-round-1` now holds `agent/designer/v1-editorial-web-catalog`): delete only the merged branch, not that worktree.
- Main checkout has other workers' uncommitted files (`tsconfig.json`, `resources/js/vite-env.d.ts`, senior handoffs): stage only your own files.
