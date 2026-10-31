# Orchestration board

Readable mirror of [`.orchestration/queue/TASKS.md`](../../.orchestration/queue/TASKS.md). Refreshed by the intern (row `DOC-board-mirror-2`) or the orchestrator.

Updated: 2026-10-31 07:25 UTC from `TASKS.md` as read at 07:25 UTC

## Workers

| Role | Worker | Current |
| --- | --- | --- |
| senior | `senior-a` | `FIX-gate-timeouts` (vitest and integration timeouts) |
| senior | `senior-b` | `P2.H.2.2` (auth field and layout copy) |
| estagiario | `estagiario` | `DOC-board-mirror-2` (this refresh); `QA-e2e-main` waits for port 3000 |
| especialista | `especialista` | `REV-P2.H.1.1` (covers review) |

## Queue

Mirror of the `Open` table in `TASKS.md`, sorted by prio. Lock set (`files`) is in the source table. Rows already `merged` are in its Done section and are not repeated here.

| prio | id | role | status | owner | note |
| --- | --- | --- | --- | --- | --- |
| 1 | P2.H.1.1 | senior | review | senior-b | ROADMAP; first visual task, the others depend on it; report `reports/P2.H.1.1.md` |
| 1 | REV-P2.H.1.1 | especialista | doing | especialista | review of P2.H.1.1 (covers), branch d98b5a6 |
| 1 | P2.H.2.1 | senior | todo | | ROADMAP |
| 1 | P2.H.3.1 | senior | review | senior-a | ROADMAP; burst number waits for Q-A2 |
| 1 | REV-P2.H.3.1 | especialista | todo | | review of reports/P2.H.3.1.md |
| 1 | REBASE-P4.A.2.4 | senior | todo | | cherry-pick `86a031e` onto a fresh branch; never merge the old branch |
| 1 | ADOPT-VF-001 | senior | todo | | owner said `adotar VF-001`; cherry-pick passkey commits, then REV |
| 1 | FIX-account-purge-lists | senior | todo | | `reports/AUDIT-2026-10-31.md` finding 1; sequence with ADOPT-VF-001 |
| 1 | FIX-legacy-import-no-login | senior | review | senior-a | cherry-pick `661114b`, `6f99082`; then REV |
| 1 | REV-FIX-legacy-import-no-login | especialista | todo | | review of reports/FIX-legacy-import-no-login.md |
| 1 | REV-FIX-legacy-cover-retry | especialista | todo | | review of reports/FIX-legacy-cover-retry.md (senior-a row is `review`) |
| 1 | QA-gate-2026-11-01 | estagiario | review | estagiario | parcial: build, lint, typecheck green; `go test ./tests` and `bun run test` hit load timeouts; e2e not run |
| 1 | REV-QA-gate-2026-11-01 | especialista | todo | | review of QA-gate-2026-11-01 (main @ 13a6e08) |
| 1 | FIX-i18n-untranslated | estagiario | review | estagiario | 4 pt-BR and es labels translated; report @ 6065fc7 |
| 1 | REV-FIX-i18n-untranslated | especialista | todo | | review of FIX-i18n-untranslated (6065fc7) |
| 1 | FIX-mobile-catalogues | estagiario | review | estagiario | `apps/mobile/src/i18n/`, branch @ c1b7990 |
| 1 | REV-FIX-mobile-catalogues | especialista | todo | | review of FIX-mobile-catalogues (c1b7990); `lists.quota.*` keys out of sync |
| 1 | FIX-gate-timeouts | senior | doing | senior-a | vitest worker timeouts and `go test -tags integration ./tests` 10 m default |
| 1 | FIX-sdk-ts-generator | senior | todo | | `bun run sdk:generate` crashes on typescript 7.0.2; after P2.H.3.1 merges |
| 1 | FIX-sdk-python-regen | senior | todo | | Python SDK regen needs `uv`; after P2.H.3.1 merges |
| 1 | QA-e2e-main | estagiario | todo | | `bun run e2e` on a container built from `main`; blocked while `hexascore-app-1` holds port 3000 |
| 2 | FIX-legacy-cover-retry | senior | review | senior-a | cherry-pick `656f426`, `0a35a00` |
| 2 | FW-triage-branches | senior | todo | | `$FW` branches: open PR or delete, one-line report |
| 2 | P2.H.1.2 | senior | todo | | ROADMAP; after P2.H.1.1 |
| 2 | P2.H.1.3 | senior | todo | | ROADMAP; after P2.H.1.1 |
| 2 | P2.H.1.4 | senior | todo | | ROADMAP; after P2.H.1.1 and P2.H.2.1 |
| 2 | P2.H.1.5 | senior | todo | | ROADMAP; after P2.H.1.4 |
| 2 | P2.H.2.2 | senior | doing | senior-b | ROADMAP; `challenge.go` unlocked after P2.H.3.2 merged |
| 2 | P2.H.2.3 | senior | review | senior-b | ROADMAP; report `reports/P2.H.2.3.md` @ 1487f66 |
| 1 | REV-P2.H.2.3 | especialista | todo | | review of P2.H.2.3 (flash toasts), branch 1487f66 |
| 2 | P2.H.3.3 | senior | review | senior-b | ROADMAP; report `reports/P2.H.3.3.md` @ cd29c95 |
| 1 | REV-P2.H.3.3 | especialista | todo | | review of P2.H.3.3 (admin nav), branch cd29c95 |
| 2 | P2.H.3.4 | senior | todo | | ROADMAP; schedule when no other UI row is `doing` |
| 2 | P2.H.3.5 | senior | todo | | ROADMAP |
| 2 | DOC-i18n-legal | estagiario | review | estagiario | Q-D-007 legal copy proposal; no catalogue edit before owner approval |
| 2 | REV-DOC-i18n-legal | especialista | todo | | review of the DOC-i18n-legal proposal; read `apps/mobile` and SDD 10 |
| 2 | QA-i18n-parity | estagiario | review | estagiario | key parity ok; 39 pt-BR and 28 es values equal to en, classified |
| 2 | REV-QA-i18n-parity | especialista | todo | | review of QA-i18n-parity (report only) |
| 2 | DOC-todo-inventory | estagiario | review | estagiario | TODO, FIXME and workaround markers inventory |
| 2 | REV-DOC-todo-inventory | especialista | todo | | review of DOC-todo-inventory (report only) |
| 2 | DOC-board-mirror-2 | estagiario | doing | estagiario | this refresh |
| 2 | VF-062-adopt | senior | todo | | adopt vukem VF-062 fix once owner releases v0.35.0 |
| 2 | QA-mobile-proto-review | estagiario | review | estagiario | read-only review of `apps/mobile` against P5.A.1 to P5.A.2 |
| 2 | REV-QA-mobile-proto-review | especialista | todo | | review of QA-mobile-proto-review (report only) |
| 2 | FIX-mobile-nav-test-timeout | senior | todo | | `apps/mobile/src/navigation.test.tsx` 5 s Jest timeout |
| 3 | DOC-pin-inventory | estagiario | review | estagiario | dependency range and Docker tag inventory |
| 3 | REV-DOC-pin-inventory | especialista | todo | | review of DOC-pin-inventory (report only) |
| 3 | P4.A.3.1 | senior | todo | | ROADMAP reserve; community submission UI |
| 3 | P4.A.3.2 | senior | todo | | ROADMAP; after P4.A.3.1 |
| 3 | P4.A.3.3 | senior | todo | | ROADMAP; after P4.A.3.2 |
| 3 | P2.H.3.6 | senior | todo | | release `v0.16.0`; after every P2.H row |
| 3 | P3.A.1.4 | senior | todo | | ROADMAP; after P2.H.3.1 and P2.H.1.1 merge |
| 3 | P3.A.2.2 | senior | todo | | `qa/sql/api_coverage.sql` |
| 3 | P3.A.4.1 | senior | todo | | `app/jobs/`, `app/api/exports*`; sequence with REBASE-P4.A.2.4 |
| 3 | P4.B.1.1 | senior | todo | | after REBASE-P4.A.2.4 and ADOPT-VF-001 |

Blocked (owner gate, see `developer-manual-todos.md` and [QUESTIONS.md](QUESTIONS.md)): CI-release-environment; P2.A.\*, P2.B.2.\*, P2.C.\*, P2.D.5.2, P2.D.6.1, P2.E.2.3, P2.E.3.2, P2.E.4.1, P2.F.1.\*, P2.F.5.\*, P2.F.6.1, P2.G.\* (Railway, DNS, Stripe live, real data, cutover, spend, secrets); P7.\* (`v1.0.0`, D-06).

## Phase status

| Group | State |
| --- | --- |
| P0, P1 | done (`v0.1.0` … `v0.9.0`) |
| P2.A–P2.G | code-only items merged; 39 boxes wait for owner gates (`developer-manual-todos.md`) |
| P2.H launch readiness | open, code-only; target `v0.16.0` |
| P3.A, P3.C | built ahead of the cutover (`v0.11.0`); P3.A.1.4 waits for P6.C; P3.B publishing is owner-gated |
| P4.A | model, quotas and anti-abuse merged; submission UI (P4.A.3) open |
| P5 | Android shell only (P5.A.1.1 code on `main`) |
| P6 | not started (legal gates) |
| P7 social layer | defined; waits for the D-06 amendment (Q-A1) and `v1.0.0` |

## Owner gates

See [QUESTIONS.md](QUESTIONS.md): Q-H31 (Stripe values), Q-A1 (social layer / D-06), Q-A2 (API burst), Q-A3 (Docker cleanup), Q-A4 (unmerged branches), `adotar VF-001`, and the Railway/DNS/data/cutover items in `developer-manual-todos.md`.
