# Orchestration board

Readable mirror of `.orchestration/queue/TASKS.md` (local, gitignored). Refreshed by the intern (row `B-034`) or the orchestrator. Priorities and acceptance live in [docs/BACKLOG.md](../BACKLOG.md); milestones in [docs/ROADMAP-V1.md](../ROADMAP-V1.md); tickets in `.orchestration/queue/tasks/<ID>.md`.

Updated: 2026-11-02 21:00 UTC from `TASKS.md` seeded by the architect audit (batch 1, milestone V1-A).

## Workers

| Role | Worker | Current |
| --- | --- | --- |
| orchestrator | main session | B-001 (merge the audit branch, gate, push, production check) |
| senior | `senior` | not launched yet; first row B-002 |
| intern | `estagiario` | not launched yet; first row B-034 |

## Queue

| prio | id | role | status | owner | note |
| --- | --- | --- | --- | --- | --- |
| 0 | B-001 | orchestrator | todo | | deploy the render-cache locale fix and verify production (W-01) |
| 0 | B-002 | senior | todo | | nightly jobs with explicit timeouts, resumable recompute (T-08) |
| 1 | B-010 | senior | todo | | Timeout/Retries on every handler, idempotent webhook and cover retries (after B-002) |
| 1 | B-011 | senior | todo | | LGPD erasure of scoring and search rows (T-05) |
| 1 | B-012 | senior | todo | | Playwright helper drift, `desktop-en` green, CI on release candidates (D-42) |
| 1 | B-034 | intern | todo | | this mirror, daily |
| 2 | B-024 | senior | todo | | retention prunes + `item_status` index (T-17, T-16) |
| 2 | B-026 | senior | todo | | dated dependency-ignore allowlist (T-21) |
| 2 | B-027 | senior | todo | | IGDB token cache (T-11) |
| 2 | B-042 | intern | todo | | copy and SEO details (W-09..W-12) |
| 2 | B-041 | intern | todo | | genre labels translated on item pages (W-08) |
| 3 | B-033 | senior | todo | | review of the 14 kept pre-rewrite branches |
| 3 | B-046 | intern | todo | | documents to English |

Reserve (tickets pending): B-043 storage health probe, B-044 alert codes, B-047 VF-NEW entries, QA-sources-verify.

## Done this cycle

- Audit branch `audit/architect-2026-11-02` (not merged yet): render cache locale fix, Docker context, audit docs, RFCs, ADRs 0024-0025, owner answers OQ-01..14.
