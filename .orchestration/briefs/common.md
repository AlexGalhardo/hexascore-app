# Common brief

You are one of three HexaScore agents. The runtime can be any coding agent. Read `AGENTS.md` in the repo you edit, then `.claude/skills/use-railway/SKILL.md` and <https://railway.com/agents/grok>. Do not assume a Claude Code tool, a skill loader, or a specific model. Do not run `railway up` unless the owner asked for a deploy.

## Paths

- App repo: `C:\dev\hexascore-app\hexascore-vukem`
- Framework repo: `C:\dev\hexascore-app\vukem` (framework agent only)
- Legacy app: `C:\dev\hexascore-app\legacy` (read-only)
- Orchestration: `C:\dev\hexascore-app\hexascore-vukem\.orchestration`
- Work in a git worktree `C:\dev\hexascore-app\wt\<role>`, not in a checkout another agent is using.

## Loop

1. Write the UTC time and the task id to `heartbeat/<role>.txt`.
2. If `STOP` exists, finish the current task, write `handoff/<role>.md`, and stop.
3. Claim the lowest numeric prefix in `queue/<role>/` by moving it to `active/<role>/`. If the move fails, take the next file.
4. If the queue is empty, write `IDLE` in the heartbeat and stop. Do not poll.
5. Do the brief. Acceptance lines are commands: run them and quote the real output.
6. Write `reports/<ID>.md` in at most 15 lines: status (`concluida`, `parcial`, `bloqueada`, or a review verdict), branch, files, commands and output, risks.
7. Implementer: add `queue/reviewer/1-REV-<ID>.md` with the branch, the report path, and the acceptance commands.
8. Reviewer `REJEITADO`: add `queue/implementer/0-<ID>-r<N+1>.md` with the fix list.
9. Reviewer approval: append `PRONTO_PARA_MERGE` and the final branch to `reports/<ID>.md`, then merge that branch into local `main` when the gate for that change is green. Do not `git push origin main`: that push deploys Railway production.
10. Move the task file to `done/` and return to step 1.

## Git

Branch `agent/<role>/<ID>` from `origin/main`. Conventional Commits, English, `Refs: <ID>`. Never force-push. Never edit `legacy`.
