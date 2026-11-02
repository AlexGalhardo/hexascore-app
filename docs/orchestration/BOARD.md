# Orchestration board

Readable mirror of `.orchestration/queue/TASKS.md` (local, gitignored). Refreshed by the intern (row `B-034`, daily) or the orchestrator. Priorities and acceptance live in [docs/BACKLOG.md](../BACKLOG.md); milestones in [docs/ROADMAP-V1.md](../ROADMAP-V1.md); tickets in `.orchestration/queue/tasks/<ID>.md`.

Updated: 2026-11-02 20:53 UTC from `TASKS.md` (after the B-022 merge at `9bc1865f`).

## Workers

| Role | Worker | Current |
| --- | --- | --- |
| orchestrator | main session | B-001 (merge the audit branch, gate, push, production check) |
| senior | `senior` | B-012 (e2e helper drift, `desktop-en`, CI on release candidates) |
| intern | `estagiario` | free (board empty for the intern role); B-020 blocked on OQ-15 |

## Queue

| prio | id | role | status | owner | note |
| --- | --- | --- | --- | --- | --- |
| 0 | B-001 | orchestrator | doing | orchestrator | merge the audit branch, gate, push, production check (W-01) |
| 1 | B-012 | senior | doing | senior | Playwright helper drift, `desktop-en` green, CI on release candidates (D-42) |
| 2 | B-024 | senior | todo | | retention prunes and `item_status` index (T-17, T-16) |
| 2 | B-026 | senior | todo | | dated dependency-ignore allowlist (T-21) |
| 2 | B-027 | senior | todo | | IGDB token cache (T-11) |
| 2 | B-020 | intern | blocked | intern | mobile item page order; OQ-15 conflicts with the owner order of 2026-11-02 |
| 2 | B-021 | senior | todo | | 404 and 410 pages (W-03) |
| 2 | B-023 | senior | todo | | search results in the first HTML (W-05) |
| 2 | B-040 | senior | todo | | listing past the last page and canonical filters (W-06, W-07) |
| 3 | B-033 | senior | todo | | review of the 14 kept pre-rewrite branches |
| 3 | B-043 | senior | todo | | reserve: `/up?checks=1` storage probe (T-15) |
| 3 | B-044 | senior | todo | | reserve: alert codes reconciliation (T-14) |
| 3 | B-047 | senior | todo | | reserve: VF-NEW entries for vukem-feedback |

## Done this cycle

- B-002 (nightly jobs with explicit timeouts), B-010 (Timeout and Retries on every handler), B-011 (LGPD erasure of scoring and search rows), B-022 (merged at `9bc1865f`).
- B-034 (this mirror), B-042 (copy and SEO details, partial: home title finished by the orchestrator), B-046 (documents to English, partial), B-041 (closed with no change: production already shows translated genres).
- QA-sources-verify (re-read of the not-verified source terms), QA-llm-pricing (Haiku 5.5 prices for the LLM ceiling), QA-attribution-research (TMDB and JustWatch attribution rules).
- Audit branch `audit/architect-2026-11-02` merged into local `main` (not pushed yet at the time of the mirror; see B-001).
- `.orchestration/done/` keeps the rows of earlier boards; the full list is `ls .orchestration/done`, not mirrored here.

Backlog reference: [docs/BACKLOG.md](../BACKLOG.md).
