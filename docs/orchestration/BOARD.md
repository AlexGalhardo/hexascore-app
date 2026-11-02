# Orchestration board

Readable mirror of `.orchestration/queue/TASKS.md` (local, gitignored). Refreshed by the intern (row `B-034`, daily) or the orchestrator. Priorities and acceptance live in [docs/BACKLOG.md](../BACKLOG.md); milestones in [docs/ROADMAP-V1.md](../ROADMAP-V1.md); tickets in `.orchestration/queue/tasks/<ID>.md`.

Updated: 2026-11-02 20:06 UTC from `TASKS.md` (batch 1, milestone V1-A, after the audit branch merge at `23a0c8e5`).

## Workers

| Role | Worker | Current |
| --- | --- | --- |
| orchestrator | main session | B-001 (merge the audit branch, gate, push, production check); the audit merge is already in local `main` |
| senior | `senior` | B-002 doing (`agent/senior/B-002`) |
| intern | `estagiario` | B-034 doing (`agent/intern/B-034`) |

## Queue

| prio | id | role | status | owner | note |
| --- | --- | --- | --- | --- | --- |
| 0 | B-001 | orchestrator | doing | orchestrator | merge the audit branch into `main`, full local gate, push, production check (W-01) |
| 0 | B-002 | senior | doing | senior | nightly jobs with explicit timeouts, resumable recompute (T-08) |
| 1 | B-010 | senior | todo | | Timeout/Retries on every handler, idempotent webhook and cover retries (after B-002) |
| 1 | B-011 | senior | todo | | LGPD erasure of scoring and search rows (T-05) |
| 1 | B-012 | senior | todo | | Playwright helper drift, `desktop-en` green, CI on release candidates (D-42) |
| 1 | B-034 | intern | doing | intern | this mirror, daily |
| 2 | B-024 | senior | todo | | retention prunes + `item_status` index (T-17, T-16) |
| 2 | B-026 | senior | todo | | dated dependency-ignore allowlist (T-21) |
| 2 | B-027 | senior | todo | | IGDB token cache (T-11) |
| 2 | B-042 | intern | todo | | copy and SEO details (W-09..W-12) |
| 2 | B-041 | intern | todo | | genre labels translated on item pages (W-08) |
| 3 | B-033 | senior | todo | | review of the 14 kept pre-rewrite branches |
| 3 | B-046 | intern | todo | | documents to English |
| 3 | B-043 | senior | todo | | reserve: `/up?checks=1` storage probe (T-15) |
| 3 | B-044 | senior | todo | | reserve: alert codes reconciliation (T-14) |
| 3 | B-047 | senior | todo | | reserve: VF-NEW entries for vukem-feedback |
| 3 | QA-sources-verify | intern | todo | | reserve: re-read the "not verified" rows of `docs/data/SOURCES.md` |

## Done this cycle

- Audit branch `audit/architect-2026-11-02` merged into local `main` at `23a0c8e5` (not pushed yet): render cache locale fix, Docker context, audit docs, RFCs, ADRs 0024-0025, owner answers OQ-01..14.
- `.orchestration/done/` holds 310 files modified since 2026-10-26 (rows of the previous board); the full list is `ls .orchestration/done`, not mirrored here.

Backlog reference: [docs/BACKLOG.md](../BACKLOG.md).
