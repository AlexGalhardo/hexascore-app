# Orchestration board

Readable mirror of `.orchestration/`. Updated by the orchestrator each cycle.

Updated: 2026-10-29 (phase 1 in progress, app at `v0.5.0`, vukem `v0.32.0`)

## Workers

| Role | Model | Worktree | Current |
| --- | --- | --- | --- |
| estagiario | Haiku 5.5 | `wt/estagiario` | idle (suite runs on request) |
| senior | Sonnet 5.5 | `wt/senior` | P1.F.4.1 follow-ups, then P1.F.2.1, P1.H.3.1, P1.G.2.4, P2.D.5.1; reserve P1.H.2.1b/c |
| especialista | Opus 5.5 | `wt/especialista` | reviews (P1.E.3.2, P1.G.4.2, P1.H.1.1, P1.H.1.2), then P1.H.2.1a |
| framework | Opus 5.5 | `wt/framework` (vukem) | stopped: PRs #29 to #34 green but merging is denied to agents (FR-002) |
| orchestrator | Opus 5.5 | `hexascore-vukem` (`main`) | integration, releases, own Opus tasks |

## Phase 1 status

| Group | State |
| --- | --- |
| P1.A, P1.B, P1.C, P1.D | done (`v0.4.0` closes P1.D) |
| P1.E Search | done (`v0.5.0`) |
| P1.F User library | P1.F.2.1 queued, then release `v0.6.0` |
| P1.G Web parity | P1.G.2.4, P1.G.3.2 left, then `v0.7.0` |
| P1.H API parity | P1.H.1.x done; P1.H.2.1 split into a/b/c; P1.H.3.1 waits on Q-H31 (Stripe variables on Railway) |
| P1.I QA gate | waits on P1.H.5.1 |

## Blockers needing the owner

- FR-002: agents cannot merge vukem PRs, so no vukem release (`v0.33.0`) and no workaround removal.
- Q-H31: Stripe test variables on Railway before billing can reach `main` (auto-deploys to production).
- Phase 2 from P2.A.1.2 on: Railway plan, buckets, DNS, secrets and spend are owner gates.
