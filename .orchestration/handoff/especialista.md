# Handoff especialista (2026-10-29, session 5)
STATUS: IDLE (queue empty). Worktree C:\dev\hexascore-app\wt\especialista on agent/especialista/P1.H.2.2-fix, clean.
Done this session:
- REV-P1.H.2.1b -> CORRIGIDO_E_APROVADO, agent/especialista/P1.H.2.1b-fix (relation filters/counts restricted to allowlisted item_entities.source via Policy.RelationSources; entity default_name without provenance omitted).
- REV-P1.H.2.1c -> CORRIGIDO_E_APROVADO, agent/especialista/P1.H.2.1c-fix (lookup public-only + per-source indexable query).
- REV-P1.G.3.2b -> APROVADO, agent/senior/P1.G.3.2b (choose-edition still bloqueada: needs editions source decision).
- REV-P1.H.2.2 -> CORRIGIDO_E_APROVADO, agent/especialista/P1.H.2.2-fix (stale docs.json).
Pitfalls (new):
- 1b-fix and 1c-fix both add a line to CHANGELOG [Unreleased] Fixed: trivial conflict.
- Editing resources/docs/*.md needs `bun run build:docs`, else vitest build-docs test fails.
- Pending note: addEntities shows entity names under the relation source, not the name source (pre-existing).
- TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable per Bash call.
