# Orchestration board

Readable mirror of `.orchestration/` (not versioned). Updated by the orchestrator each cycle.

Updated: 2026-10-28 (cycle 0, workers launching)

## Workers

| Role | Model | Worktree | Current task |
| --- | --- | --- | --- |
| estagiario | Haiku 5.5 | `wt/estagiario` | E-001 |
| senior | Sonnet 5.5 | `wt/senior` | P0.F.1.1 |
| especialista | Opus 5.5 | `wt/especialista` | P0.F.1.2 |
| framework | Opus 5.5 | `wt/framework` (vukem) | first open VF by severity (VF-001, high) |
| orchestrator | Opus 5.5 | `hexascore-vukem` (`dev`) | integration, queues |

## Phase 0 microtasks

| ID | State | Owner |
| --- | --- | --- |
| P0.F.1.1 | queued | senior |
| P0.F.1.2 | queued | especialista |
| P0.F.1.3 | waiting: P0.F.1.1, P0.F.1.2, Q-001 | — |
| P0.F.2.1 to P0.F.10.1 | waiting: P0.F.1.1 | — |
| P0.F.11.1 | waiting: all P0.F | orchestrator |

## Integrated

None yet.

## Notes

- Phase 0 code lives in the local-only spike repo `C:\dev\hexascore-app\hexascore-spike` (never pushed); only `docs/spike-report.md` and `vukem-feedback.md` land here.
- Framework releases are owner-gated (see `QUESTIONS.md`).
