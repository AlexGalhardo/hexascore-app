# Orchestration board

Readable mirror of [`.orchestration/queue/TASKS.md`](../../.orchestration/queue/TASKS.md). Refreshed by the intern (row `DOC-board-mirror`) or the orchestrator.

Updated: 2026-10-31 (architecture audit; no workers running; app `v0.11.0` + audit fixes on local `main`, vukem `v0.34.0`)

## Workers

| Role | Worker | Current |
| --- | --- | --- |
| senior | `senior-a`, `senior-b` | not launched; first rows P2.H.3.2 (flaky test, CI red) and P2.H.1.1 (covers) |
| estagiario | `estagiario` | not launched; first rows DOC-board-mirror, QA-gate-2026-11-01 |
| especialista | `especialista` | not launched; first row AUDIT-social-sdd when no review is pending |

## Phase status

| Group | State |
| --- | --- |
| P0, P1 | done (`v0.1.0` … `v0.9.0`) |
| P2.A–P2.G | code-only items merged; 39 boxes wait for owner gates (`developer-manual-todos.md`) |
| P2.H launch readiness | 14 boxes open, all code-only; target `v0.16.0` |
| P3.A, P3.C | built ahead of the cutover (`v0.11.0`); P3.A.1.4 trending waits for P6.C; P3.B publishing is owner-gated |
| P4.A | model, quotas and anti-abuse merged; submission UI (P4.A.3) open |
| P5 | Android shell only (P5.A.1.1 code on `main`) |
| P6 | not started (legal gates) |
| P7 social layer | defined; waits for the D-06 amendment (Q-A1) and `v1.0.0` |

## Owner gates

See [QUESTIONS.md](QUESTIONS.md): Q-H31 (Stripe values), Q-A1 (social layer / D-06), Q-A2 (API burst), Q-A3 (Docker cleanup), Q-A4 (unmerged branches), `adotar VF-001`, and the Railway/DNS/data/cutover items in `developer-manual-todos.md`.
