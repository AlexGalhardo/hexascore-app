# Handoff especialista (2026-10-29)
STATUS: HANDOFF (heavy context after 5 tasks). Queue left: 1-REV-P1.G.4.2.md, 1-REV-P1.H.1.1.md (both untouched).
Done this session:
- REV-P1.G.2.2 APROVADO -> agent/senior-b/P1.G.2.2 (already merged)
- REV-P1.G.3.1 CORRIGIDO_E_APROVADO -> agent/especialista/P1.G.3.1-fix (searchId in partial reload `only`)
- P1.F.3.2-fix (self) -> agent/especialista/P1.F.3.2-fix (duplicate "contact" i18n block + duplicate-key vitest)
- P1.F.4.1 (self, Opus) -> agent/especialista/P1.F.4.1 (LGPD export); follow-ups queued in senior/2-P1.F.4.1-followup-*.md
- REV-P1.G.4.1 CORRIGIDO_E_APROVADO -> agent/especialista/P1.G.4.1-fix (main build broken on clean checkout: resources/public/.gitkeep; /\host doc link)
Pitfalls:
- main currently fails `go test`/`go build` on a clean checkout until P1.G.4.1-fix merges (embed all:public).
- hexascore-postgres-1 container gets removed by someone (compose down?); `docker compose up -d postgres` in $APP restores it.
- Don't use python text-mode writes on Windows; use Edit tool. Git-bash sed `a\` loses newlines.
- errs.NewKey params are i18n params, not problem members (VF-016/039) -> codes missing in reports/scoring problems (senior follow-up queued).
- Worktree C:\dev\hexascore-app\wt\especialista is on agent/especialista/P1.G.4.1-fix, clean. Scratch DB esp_schema (migrated) exists in local Postgres.
Skills: karpathy-guidelines loaded; caveman/golang-security not loaded (context economy).
