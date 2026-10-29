# Orchestration board

Readable mirror of `.orchestration/` (not versioned). Updated by the orchestrator each cycle.

Updated: 2026-10-29T01:05Z (cycle 1)

## Workers

| Role | Model | Worktree | Current task |
| --- | --- | --- | --- |
| estagiario | Haiku 5.5 | `wt/estagiario` | E-011 (relaunched after handoff) |
| senior | Sonnet 5.5 | `wt/senior`, `wt/spike-senior` | P0.F.4.1, then P0.F.5.1, P0.F.8.1 |
| especialista | Opus 5.5 | — | not running: launch denied twice by the auto-mode classifier; the orchestrator covers reviews |
| framework | Opus 5.5 | `wt/framework` (vukem) | VF-002 (PR #7); urgent FW-CI-MONGO queued |
| orchestrator | Opus 5.5 | `hexascore-vukem` (`dev`) | integration, reviews, Opus microtasks |

## Phase 0 microtasks

| ID | State | Owner |
| --- | --- | --- |
| P0.F.1.1 | integrated (works; VF-032, VF-033) | senior |
| P0.F.1.2 | integrated (works) | orchestrator |
| P0.F.1.3 | blocked: Q-001 (owner yes for `golang.org/x/crypto`) | — |
| P0.F.4.1 | in progress | senior |
| P0.F.5.1, P0.F.8.1 | queued | senior |
| P0.F.2.1, P0.F.3.1 | ready (Opus) | orchestrator |
| P0.F.2.2, P0.F.2.3, P0.F.5.2, P0.F.6.1, P0.F.7.x, P0.F.9.1, P0.F.10.1 | waiting on dependencies | — |
| P0.F.11.1 | waiting: all P0.F | orchestrator |

## Integrated in `dev`

E-004, E-009, E-010 (docs lint and citations), P0.F.1.1, P0.F.1.2.

## Blockers

- vukem CI job "Mongo driver (integration)" fails on every PR (floating `mongo:8` service); blocks PR #6 and framework merges. Task FW-CI-MONGO.
- Q-001 and FQ-001 in `QUESTIONS.md`.
