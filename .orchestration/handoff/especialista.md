# Handoff: especialista (2026-10-31, session 9)

Reason: fifteen reviews in this session and the context is heavy. STATUS: HANDOFF.

## Done this session

vukem (squash merges, 33/33 green each; no tag, no release):

| Row | Verdict | PR | Squash |
| --- | --- | --- | --- |
| REV-VF-NEW-AUDIT-2 | APROVADO (app.go conflict with fuzzAuth resolved on the main merge) | #115 | 01ca804a |
| REV-VF-NEW-AUDIT-4 | APROVADO | #116 | c7b5efe4 |
| REV-VF-NEW-P2.H.3.1 | APROVADO | #117 | 7f9b9593 |
| REV-VF-NEW-P5.A.3.1 | APROVADO | #118 | 01023abd |
| REV-VF-NEW-P4.A.3.1b | APROVADO | #119 | 073fc0f7 |
| REV-VF-NEW-FIX-covers-route | CORRIGIDO_E_APROVADO (storage.mdx named the root version v0.33.0; now `drivers/s3` releases after v0.3.1, ca92aa48) | #120 | 23cde35a |

App (local `main`):

| Row | Verdict | Merge |
| --- | --- | --- |
| REV-FIX-search-timeout-500 | APROVADO | fe0f28d |
| REV-P4.A.3.1d | CORRIGIDO_E_APROVADO (precision and role labels translated) | 9bda991 |
| REV-FIX-nojs-428-page | CORRIGIDO_E_APROVADO (golines) | 4bf6057, 8772bf2 |
| REV-P6.C.1.1 | APROVADO, ticked | d85897a |
| REV-FIX-community-create-ux | CORRIGIDO_E_APROVADO (uuid lookup instead of `id::text IN`); r2 locale commits merged 7ebbc1a; P4.A.3.2 and P4.A.3.3 ticked (QA-community-e2e 12/12) | 7732fee, 7ebbc1a |
| REV-P4.B.1.1 | CORRIGIDO_E_APROVADO (migration moved to 20261031160000, it collided with P6.C.1.1); views in app/admin/views accepted; ticked | 311e39b |
| REV-P5.A.3.2 | APROVADO incl. late commit 7cc7586 (Q-B1 deps, ADR 0019); ticked | 2600b4a |
| REV-FIX-mobile-login-test-timeout | APROVADO | 6d5708a |
| REV-FIX-ssr-supervise-flake | APROVADO | b6c3fe8 |
| REV-FIX-search-prod-slow (prio 0) | CORRIGIDO_E_APROVADO (SDD 07 documents the 500-row trigram cap, lower-bound totals and the transaction-local `enable_seqscan = off`; golines); p95 345 ms on the 160k seed, above NFR-07-1 | 1c95323 |
| (own) FIX-mobile-catalogues-sync | mobile catalogue copies regenerated after the web string changes | 0237ecc |
| (own) FIX-gate-lint | 5 golangci issues in the new tests (golines, gci, gofumpt, rowserrcheck) | b344207 |

Gates on main: batch 1 (f9bcb8f) all green. Batch 2 (311e39b): build, bun, vitest 27/96 and `go test -tags integration ./...` green, golangci 5 issues fixed in FIX-gate-lint; re-check after b344207 + 7ebbc1a, then full gate on 1c95323 all green: golangci 0 issues, bun lint/typecheck/test green, `./tests -run 'Community|RouteStackInventory'` and `tests/i18n` ok.

## Open

1. NEXT: REV-P5.A.3.3, REV-P5.A.3.4, REV-P5.A.3.5 (senior), REV-DESIGN-round-1 (designer). All `todo`.
2. P4.B.1.1 is on local main; no release before FIX-moderator-2fa (orchestrator row, senior-a). Its worktree `wt/senior-FIX-moderator-2fa` was cut from 49529f0 (old P4.B.1.1, migration 20261031150000): the author must merge `main` first, or the moderator migration comes back under the colliding timestamp.
3. #120 adoption waits on Q-A14 (owner): the app pins `vukem/drivers/s3` v0.3.1 and the ETag fix needs a `drivers/s3` release, then an ADOPT bump. No app FIX row.

## Pitfalls

- A REV branch can gain commits after the review (P5.A.3.2, FIX-community-create-ux). Before `git merge`, run `git log main..<branch>` and compare with the reviewed SHA.
- Run golangci on `./tests/` (integration tag) for any row that adds tests; targeted package lint missed 5 issues.
- Web string changes in `resources/lang` break `apps/mobile/src/i18n/keys.test.ts` until `bun run --cwd apps/mobile catalogues` is committed.
- Leftover worktree folders after `git worktree remove` hold only bun junctions (no files). Delete the reparse points with `[System.IO.Directory]::Delete(path)` (does not follow), deepest first, then `rmdir /s /q`.
- vukem: review worktrees are detached (`wt/esp-fw-<n>`) because senior-fw has the branches checked out; push with `HEAD:<branch>`. Helper scripts: scratchpad `upd.sh` (merge origin/main, regenerate last-updated, push) and `ci.sh` (wait for checks). Never commit while a merge is half-resolved: on #116 a conflicted `last-updated.json` reached the remote and needed a follow-up commit.
