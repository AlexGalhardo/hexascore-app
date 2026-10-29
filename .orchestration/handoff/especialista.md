# Handoff especialista (2026-10-29, session 6)
STATUS: IDLE (queue + reserve empty). Worktree C:\dev\hexascore-app\wt\especialista on agent/especialista/P1.G.3.2d-fix, clean.
Done this session:
- REV-FIX-entity-name-source -> APROVADO (merged).
- REV-P1.G.2.4 -> CORRIGIDO_E_APROVADO, agent/especialista/P1.G.2.4-fix (item page linked countries/languages by slug; route matches code -> 404).
- REV-P1.G.3.2c -> APROVADO (merged); follow-up 2-P1.G.3.2d created, then reviewed below.
- REV-FIX-highlight-score-scale -> CORRIGIDO_E_APROVADO, agent/especialista/FIX-highlight-score-scale-fix (score.outOf label "out of 100" -> "out of 10").
- REV-P1.H.4.1 -> APROVADO (merged).
- REV-P2.D.5.1 -> CORRIGIDO_E_APROVADO, agent/especialista/P2.D.5.1-fix (sdks/ts/bun.lock in osv + dependabot); created queue/senior/2-P2.D.5.2 (js-yaml 4.2.0 high advisories via @hey-api 0.99.0; deps job red until fixed).
- REV-P1.G.3.2d -> CORRIGIDO_E_APROVADO, agent/especialista/P1.G.3.2d-fix (double-click duplicate page guard).
Pitfalls (new):
- G.2.4-fix and score-scale-fix both add CHANGELOG [Unreleased] ### Fixed: trivial conflict.
- Creating/editing any .env is denied by a deny rule -> `vukem serve` curl acceptance not feasible; running downloaded binaries (osv-scanner) denied by auto mode.
- `git show ref:.github/...` on Git Bash needs MSYS_NO_PATHCONV=1.
- EditionPicker.test.tsx needs explicit RTL cleanup (no vitest globals).
