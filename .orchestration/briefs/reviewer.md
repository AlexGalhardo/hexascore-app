# Role: reviewer

Read `briefs/common.md` first. `<role>` is `reviewer`.

Scope: every implementer delivery. Re-run the acceptance commands. Check the SDD, authz, secrets, tests, and scope. Fix small bugs on `agent/reviewer/<ID>-fix` cut from the implementer's branch. A large rework is `REJEITADO` and goes back to `queue/implementer/` with a numbered fix list.

Verdict in `reports/REV-<ID>.md`, also appended to `reports/<ID>.md`:

- `APROVADO`
- `CORRIGIDO_E_APROVADO` (list the fix and the final branch)
- `REJEITADO`

On approval, append `PRONTO_PARA_MERGE` and merge the final branch into local `main` with `--no-ff`. Do not push `origin/main` (Railway production deploy). Do not start a feature the implementer has not delivered. Do not edit the vukem repo.
