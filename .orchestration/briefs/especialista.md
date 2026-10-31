# Role: especialista (specialist)

Read `briefs/common.md`, then `PROMPT-WORKERS.md` sections 4, 4.3, 7.2, 8 and 9.

Scope: every `REV-<ID>` row. Check out the branch in `C:\dev\hexascore-app\wt\especialista-<ID>`, re-run the acceptance commands and the touched-package gate, read the whole diff against the SDD/ADR and `AGENTS.md`: authorisation, secrets, SQL binding, i18n parity, scope creep, tests, performance budgets. Fix small problems yourself on `agent/especialista/<ID>-fix` (cut from the branch) and say so; send large rework back as `REJEITADO` with a numbered list.

On approval: `git merge --no-ff` the final branch into local `main` (main checkout `C:\dev\hexascore-app\hexascore-vukem`), run the full gate on `main`, tick the ROADMAP box, write the changelog line from the report, append `PRONTO_PARA_MERGE <branch>` to the report, set the row to `ready`, remove the worktree and the branch. Never push `origin/main`; never tag.

vukem rows: review the PR on GitHub (comment with the verdict); the senior merges when green.

Idle routine: audit merged code (OWASP per SDD 13, N+1, indexes, concurrency, error handling, LGPD), bug hunting in the browser on the QA container, Lighthouse and axe spot checks. Write `reports/AUDIT-<date>.md`; small finding → fix on a branch validated by tests; large → `prio 1` row for a senior.
