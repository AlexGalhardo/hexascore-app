# Orchestration board

Readable mirror of [`.orchestration/queue/TASKS.md`](../../.orchestration/queue/TASKS.md). Refreshed by the intern (row `DOC-board-mirror`) or the orchestrator.

Updated: 2026-10-31 from `TASKS.md` as read at 06:20 UTC (workers launched: `estagiario` running; `senior-a` on P2.H.3.2, `senior-b` on P2.H.1.1, `especialista` on AUDIT-social-sdd; app `v0.11.0` + audit fixes on local `main`, vukem `v0.34.0`)

## Workers

| Role | Worker | Current |
| --- | --- | --- |
| senior | `senior-a`, `senior-b` | `senior-a`: P2.H.3.2 (flaky test, CI red); `senior-b`: P2.H.1.1 (covers) |
| estagiario | `estagiario` | DOC-board-mirror (this refresh), then QA-gate-2026-11-01 |
| especialista | `especialista` | AUDIT-social-sdd |

## Queue

Mirror of the `Open` table in `TASKS.md`, sorted by prio. Lock set (`files`) is in the source table.

| prio | id | role | status | owner | note |
| --- | --- | --- | --- | --- | --- |
| 0 | P2.H.3.2 | senior | doing | senior-a | ROADMAP; CI on `main` is red because of this test |
| 1 | P2.H.1.1 | senior | doing | senior-b | ROADMAP; first visual task, the others depend on it |
| 1 | P2.H.2.1 | senior | todo | | ROADMAP |
| 1 | P2.H.3.1 | senior | todo | | ROADMAP; burst number waits for Q-A2, enum part can start |
| 1 | REBASE-P4.A.2.4 | senior | todo | | cherry-pick `86a031e` onto fresh branch; never merge the old branch |
| 1 | ADOPT-VF-001 | senior | todo | | owner said `adotar VF-001`; cherry-pick passkey commits, then REV |
| 1 | FIX-legacy-import-no-login | senior | todo | | cherry-pick `661114b`, `6f99082`; decide on `263a3a7`, then REV |
| 1 | DOC-board-mirror | estagiario | doing | estagiario | this file |
| 1 | QA-gate-2026-11-01 | estagiario | todo | | four gate commands on `main` plus `bun run e2e` |
| 2 | FIX-legacy-cover-retry | senior | todo | | cherry-pick `656f426`, `0a35a00`, then REV |
| 2 | FIX-export-purge-zip | senior | todo | | cherry-pick `4750ae7`, then REV |
| 2 | FW-triage-branches | senior | todo | | `$FW` branches: open PR or delete, one-line report |
| 2 | P2.H.1.2 | senior | todo | | ROADMAP; after P2.H.1.1 |
| 2 | P2.H.1.3 | senior | todo | | ROADMAP; after P2.H.1.1 |
| 2 | P2.H.1.4 | senior | todo | | ROADMAP; after P2.H.1.1 and P2.H.2.1 |
| 2 | P2.H.1.5 | senior | todo | | ROADMAP; after P2.H.1.4 |
| 2 | P2.H.2.2 | senior | todo | | ROADMAP; `challenge.go` locked by P2.H.3.2 until it merges |
| 2 | P2.H.2.3 | senior | todo | | ROADMAP |
| 2 | P2.H.3.3 | senior | todo | | ROADMAP |
| 2 | P2.H.3.4 | senior | todo | | ROADMAP; schedule when no other UI row is `doing` |
| 2 | P2.H.3.5 | senior | todo | | ROADMAP |
| 2 | DOC-i18n-legal | estagiario | todo | | Q-D-007 legal copy proposal; no catalogue edit before owner approval |
| 2 | VF-062-adopt | senior | todo | | adopt vukem VF-062 fix once owner releases v0.35.0 |
| 3 | P4.A.3.1 | senior | todo | | ROADMAP reserve; community submission UI |
| 3 | P2.H.3.6 | senior | todo | | release `v0.16.0`; after every P2.H row |
| 3 | QA-mobile-proto-review | estagiario | todo | | read-only review of `apps/mobile` against P5.A.1–P5.A.2 |
| 3 | AUDIT-social-sdd | especialista | doing | especialista | SDD 14 §8 outline for Phase 7; no decision taken |

Blocked (owner gate, see `developer-manual-todos.md` and [QUESTIONS.md](QUESTIONS.md)): CI-release-environment; P2.A.\*, P2.B.2.\*, P2.C.\*, P2.D.5.2, P2.D.6.1, P2.E.2.3, P2.E.3.2, P2.E.4.1, P2.F.1.\*, P2.F.5.\*, P2.F.6.1, P2.G.\* (Railway, DNS, Stripe live, real data, cutover, spend, secrets); P7.\* (`v1.0.0`, D-06).

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
