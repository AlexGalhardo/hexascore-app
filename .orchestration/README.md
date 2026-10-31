# Orchestration

State of the execution team defined in [`../PROMPT-WORKERS.md`](../PROMPT-WORKERS.md) (the contract). Roles: `orchestrator`, `senior` (two workers: `senior-a`, `senior-b`), `estagiario`, `especialista`.

| Path | What |
| --- | --- |
| `queue/TASKS.md` | the single task board: one row per task, claim by writing your name, lock by the `files` column |
| `queue/tasks/<ID>.md` | brief for tasks that are not a self-contained ROADMAP microtask |
| `briefs/<role>.md` | role card (summary of the contract for that role) |
| `reports/<ID>.md` | delivery report, review appended; `AUDIT-*.md` for proactive audits |
| `handoff/<role>.md` | state written at a context-limit stop; `orchestrator.md` is the resume point |
| `heartbeat/<role>.txt` | UTC timestamp + current task (not versioned) |
| `done/` | archived queue files of the previous three-role model (`implementer`, `reviewer`, `framework`), kept so old reports stay readable |
| `vf-map.md` | VF entry → ROADMAP microtask map (generated) |
| `STOP` | pause file; workers finish the current step and stop |

Old names in `done/` and `reports/`: `senior` and `estagiario` were later called `implementer`, `especialista` was `reviewer`; the 2026-10-31 team brings the original names back.
