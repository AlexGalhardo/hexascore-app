# Role: orchestrator (main session)

Your contract is the whole of `PROMPT-WORKERS.md`; sections 2, 3, 4.6, 5 (last bullet) and 10 are yours.

Cycle:

1. Resume: read `handoff/orchestrator.md`, `docs/orchestration/QUESTIONS.md`, `C:\dev\hexascore-app\developer-manual-todos.md`; `git pull`; `docker compose up -d --wait`; run the gate once.
2. Fill `queue/TASKS.md`: at least 3 `todo` rows per role and 5 reserve rows, from ROADMAP microtasks without an open owner gate (P2.H first), `VF-*` entries by severity, debt from `reports/AUDIT-*.md`, docs sync. Every row has a `files` lock set that does not intersect another running row.
3. Launch the four workers in one message (`run_in_background: true`), each told to read `briefs/common.md` and `briefs/<role>.md`, and its `handoff/<role>.md` if any. Parallel workers need the owner's yes in each session.
4. Loop: watch heartbeats (section 4.6), answer worker questions in the row notes, set `merged` after the specialist's `PRONTO_PARA_MERGE` and a sampled diff, keep the queue full, write `docs/orchestration/QUESTIONS.md` entries for the owner, keep `docs/orchestration/BOARD.md` current (or delegate to the intern).
5. Releases on the owner's request: `scripts/release.sh vX.Y.Z` on a clean `main`, container smoke test, push the commit and the tag one at a time.
6. Pause: create `STOP`, wait for the four `STATUS: STOPPED`, write `handoff/orchestrator.md`, update `docs/PROGRESS.md`.

You may fix integration conflicts and run the gate on `main`; you do not take implementation rows while a worker is idle.
