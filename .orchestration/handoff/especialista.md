# Handoff: especialista (2026-10-31, session 8)

Reason: twelve reviews and the worktree cleanup are done, and context is heavy. STATUS: HANDOFF.

## Done this session

vukem (squash merges, every PR 33/33 green; no tag, no release):

| Row | Verdict | PR | Squash |
| --- | --- | --- | --- |
| REV-VF-NEW-AUDIT-3 | CORRIGIDO_E_APROVADO (shell moved into `renderRoot` after #107; docs section moved out of the SSR bullet list) | #109 | 11761d20 |
| REV-VF-NEW-AUDIT-5 | CORRIGIDO_E_APROVADO (guard test renamed `TestSecurity_A01_...`) | #110 | 7fb2cec4 |
| REV-VF-NEW-P2.H.3.5 | APROVADO, VF withdrawn (`SSRConfig.Cache` exists) | #111 | 29d90fa0 |
| REV-FW-docs-font-size | APROVADO (no overflow at 390 and 1440 px, measured with Playwright) | #112 | 17a3f955 |
| REV-VF-NEW-OWNER-schemathesis | CORRIGIDO_E_APROVADO (digest pin, 0755/0644 mounts for the non-root image user, auth integration test; real Docker run ok) | #113 | 7ddace87 |
| REV-FW-last-updated-check | APROVADO | #114 | 66234fb4 |

App (local `main`, targeted checks only per section 8.2):

| Row | Verdict | Merge |
| --- | --- | --- |
| REV-DOC-design-system | post-merge APROVADO, facts sampled, no fixes | f9aef56 (orchestrator) |
| REV-FIX-community-problem-codes | APROVADO | afc0d43 |
| REV-P4.A.3.2 | APROVADO; ROADMAP box waits QA-community-e2e | 639005a |
| REV-P4.A.3.3 | APROVADO; ROADMAP box waits QA-community-e2e | 3314f11 |
| REV-QA-ci-local-script | CORRIGIDO_E_APROVADO (committed `.pyc` removed, `__pycache__/` ignored); AGENTS.md Commands row and Release gate pointer | 43dd3b5 |

Worktree cleanup (owner request):

| Worktree | Action | Ref |
| --- | --- | --- |
| estagiario-FIX-i18n-untranslated | net zero; rows `merged`; branch deleted | tag checkpoint/FIX-i18n-untranslated-2026-10-31 |
| especialista-ADOPT-VF-001 | HEAD is a main-into-branch merge; `git merge-tree main HEAD` = main tree; removed | tag checkpoint/ADOPT-VF-001-2026-10-31 |
| especialista-P2.H.1.2 | same, removed | tag checkpoint/P2.H.1.2-2026-10-31 |
| especialista-P2.H.1.3 | same, removed | tag checkpoint/P2.H.1.3-2026-10-31 |
| especialista-P2.H.2.2 | same, removed | tag checkpoint/P2.H.2.2-2026-10-31 |
| senior-P4.A.3.2, senior-P4.A.3.3, senior-QA-ci-local-script, senior-FIX-community-problem-codes | merged; worktrees and branches removed | see merges |
| esp-fw-* (6 review worktrees) | PRs merged; removed | see PRs |

## Open

1. NEXT: vukem PRs #115-#120 (REV rows 161-173 of TASKS.md): AUDIT-2 #115, AUDIT-4 #116, P2.H.3.1 #117, P5.A.3.1 #118, P4.A.3.1b #119 (one Windows job rerunning, failure `TestToolEnvDoesNotWaitForGrandchildren` is unrelated per senior-fw), FIX-covers-route #120. With #114 merged, a stale "Last updated" entry from a squash is tolerated; still, `main` requires the branch up to date (strict), so after each squash the next PR shows BEHIND: merge origin/main into it, push, wait for CI.
2. #120 reconciliation is written in `reports/VF-NEW-FIX-covers-route.md`: the app pins the separate module `vukem/drivers/s3` at v0.3.1, and no `drivers/s3/*` tag contains #29 (the ETag in `Open`). No app FIX row: the fix needs a `drivers/s3` release (owner) and then an ADOPT bump. Write that in `vukem-feedback.md` when #120 merges.
3. Full gate on main 6a0502c: build, golangci (integration), biome, tsc, vitest 26/91 green; `go test -tags integration ./...` failed only on `TestSecurity_A01_HexaRouteStackInventory` (five community routes from P4.A.3.2/3.3 missing from `tests/testdata/route_inventory.txt`). Fixed in b8195a6, merged 8942287; rerun of `./tests/...` on main is in the last section of `reports/P4.A.3.3.md`. Lesson: a targeted `-run Community` misses the route inventory; add `TestSecurity_A01_HexaRouteStackInventory` to the targeted set whenever routes change.
4. Follow-ups from QA-ci-local-script for the orchestrator: osv-scanner GO-2026-5932 (golang.org/x/crypto 0.58.0) turns CI `deps` red; `app/ssr` TestSuperviseRestartsUntilCancelled flakes under load.

## Pitfalls

- Owner rule 8.2: do not run vukem `task ci` locally. Review the diff, run the targeted tests, and rely on GitHub CI.
- The golangci-lint lock is global even with a private TMP. Two of my own runs blocked each other.
- A killed `fwci` bash leaves `task.exe` orphans that keep writing the same log. Do not trust a log that a killed run shared.
- Leftover worktree folders: use `cmd /c rmdir /s /q`, which does not follow links. PowerShell 5.1 `Remove-Item -Recurse` can follow junctions such as bun's `node_modules`.
- I did not touch the main checkout's uncommitted `tsconfig.json` or `resources/js/vite-env.d.ts`. `.orchestration/handoff/senior-*.md` are also modified there by other workers; commit named files only.
