# Common brief

Your contract is `C:\dev\hexascore-app\hexascore-vukem\PROMPT-WORKERS.md`; read it whole before the first task, then `AGENTS.md` of the repository you edit. This card only repeats the parts you need every minute.

- Board: `C:\dev\hexascore-app\hexascore-vukem\.orchestration\queue\TASKS.md` (absolute path, never the copy inside your worktree). Claim = write your name in `owner` and `doing` in `status`, wait 3 s, re-read.
- Lock: the `files` column. Do not claim a row whose files intersect a `doing`, `review` or `rework` row.
- Worktree: `git -C C:\dev\hexascore-app\hexascore-vukem worktree add C:\dev\hexascore-app\wt\<role>-<ID> -b agent/<role>/<ID> main`. Work only there.
- Heartbeat: `heartbeat/<role>.txt`, UTC time + task id, every step. `STOP` present → finish the step, write `handoff/<role>.md`, stop.
- Delivery: `reports/<ID>.md` in the section 7.1 format, commit with `Refs: <ID>`, row → `review`, add the `REV-<ID>` row. Never push `main`.
- Shared files (`CHANGELOG.md`, `ROADMAP.md` boxes, `docs/PROGRESS.md`) are written by the specialist at merge and by the orchestrator; you put the lines in your report.
- Owner questions go to `docs/orchestration/QUESTIONS.md`; then take the next row.
- Idle: proactive routine of your role (section 5), then wait without blocking and re-read the board. Never finish the session for lack of work.
