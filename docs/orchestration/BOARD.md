# Orchestration board

Readable mirror of `.orchestration/`. Updated by the orchestrator each cycle.

Updated: 2026-10-29 (resume in progress; app at `v0.8.0` on origin/main; local branch `agent/senior/P1.I.5.1-r2` has the gzip and challenge-limit fixes; vukem `v0.33.0` in `go.mod`; vukem PR #63 merged)

## Workers

| Role | Model | Worktree | Current |
| --- | --- | --- | --- |
| senior | Sonnet 5.5 | `wt/senior` | P1.I.5.1 (Lighthouse) in progress on `agent/senior/P1.I.5.1-r2`; queue: FIX-e2e-signup-load-r2, P1.I.4.1, P2.B.1.4, P2.B.1.5, P2.C.1.2, P2.C.2.2 |
| especialista | Opus 5.5 | `wt/especialista` | P2.F.4.1 in progress; queue: REV-P1.I.3.1, P2.G.1.1, P2.E.1.2, P2.C.3.3; blocked: APP-PASSKEYS-b, P2.F.2.2b |
| framework | Opus 5.5 | `wt/framework` (vukem) | idle; open PRs #63, #71, #77, #80-#82, #87-#91; recreate #66/#67 on main |
| estagiario | Haiku 5.5 | `wt/estagiario` | idle |
| orchestrator | Opus 5.5 | `hexascore-vukem` (`main`) | integration, vukem PR merges, releases |

## Phase status

| Group | State |
| --- | --- |
| P1.A to P1.G | done (`v0.1.0` to `v0.7.0`) |
| P1.H API parity | done (`v0.8.0`); billing wiring of the pricing page waits on Q-H31 |
| P1.I QA parity gate | I.1.1, I.2.1 to I.2.4 merged (I.2.x boxes stay open for passkeys and pricing specs); I.3.1 in review; I.5.1 in progress; I.4.1, I.6.1 rows, I.7.1 (`v0.9.0`) left |
| P2 | code-only items merged ahead of the owner gates: A.3.2, C.1.1, D.2.1 to D.2.3, D.3.1, D.5.1, D.5.3 (incident docs, local), E.1.1, E.2.1, E.2.2, E.3.1, F.2.1 to F.2.7, F.3.1; staging verification deferred |

## Owner gates

See [QUESTIONS.md](QUESTIONS.md): Stripe live values (Q-H31), Railway staging/buckets/DNS/secrets, Grafana and Telegram accounts, the vukem release for the features merged since v0.33.0.
