# Handoff: especialista (2026-11-01, session 23)

Reason: assigned work done. STATUS: HANDOFF.

## Done this session (local `main`, no push, no tag)

| Rows | Verdict | Merge |
| --- | --- | --- |
| FIX-ingest-rollback-gaps | CORRIGIDO_E_APROVADO: fix 058edf32 lists the `items_visibility_changes` trigger among the `api_changes` writers in SDD 05 §7.3 | b22161a2 |

Both rows are `ready`. Worktree `wt/senior-fw-gaps` and branches `agent/senior/FIX-ingest-rollback-gaps` and `agent/especialista/FIX-ingest-rollback-gaps-fix` are removed.

## Open

1. Follow-ups (not filed as rows): scoring `ItemScoreChanged` and community apply still write `updated` for non-public items; rollback turns a run-created `takedown` item into `hidden`.
2. Full integration suite not run on main after the merges (host short on memory); left to the batch gate.
3. Items 1-5 of the session-21 handoff (REV-FIX-seed-flag hotfix path, VF-065, provisional VF ids in app comments, no release PR, prio 3 rows) still stand.

## Pitfalls

- Git Bash mangles `branch:path` in `git show`: export `MSYS_NO_PATHCONV=1`.
- The commit-msg hook rejects git's default `Merge branch ...` header: pass `-m "chore(repo): merge ..."`.
- `TEST_DATABASE_URL` must be exported in the same Bash call as `go test` (each call is a fresh shell); `pgtest` fails, it does not skip.
- `markdownlint-cli2` on one file also lints its config globs; a stray `qa/e2e/test-results/**/error-context.md` reports MD047 (not ours).
- Previous pitfalls (`.env.example`/`tsconfig.json` local edits, golangci `fmt --diff`, `-j 2`, route inventory test, mobile catalogues) still apply.
