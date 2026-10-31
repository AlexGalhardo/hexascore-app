# FW-otlp-pr OTLP traces fix for Railway (vukem drivers/otel)
status: concluida
branch: fw: fix/railway-otlp-traces @ e3eca299 (not pushed, no PR)
files: drivers/otel/ (provider + test), site/content/docs/{deployment/docker-and-railway,deployment/logging,digging-deeper/opentelemetry}.mdx
acceptance:
- `task ci` (lint, vet, mod:check, snippets:check, test) in wt/senior-FW-railway-obs → all ok, no FAIL
- `task api:check` → exit 0 (no stable v1 tags, all modules skipped)
validation: lint ok, typecheck/vet ok, tests ok (all modules, -short)
changelog: Fixed - drivers/otel: OTLP trace export works on Railway (see commit)
roadmap: no box
risks / follow-ups: specialist pushes + opens PR. Docs pages changed: `site/content/last-updated.json` must be refreshed with `npm run last-updated` in `site/` (needs full-history clone) before the PR, else CI "Docs last updated" fails. golangci-lint refuses parallel runs across workers (task ci needs retry when another worker lints).
vukem-feedback status line: VF (Railway OTLP) -> fixed on branch fix/railway-otlp-traces, pending PR.

## Review (especialista, 2026-10-31)

verdict: CORRIGIDO_E_APROVADO
re-run: `task ci` → rc 0; `task api:check` → rc 0; `go test ./drivers/otel -run TestProvider` → ok; golangci-lint and vet on drivers/otel → 0 issues
fixes: 1. `site/content/last-updated.json` refreshed with `node scripts/last-updated.ts` on full history (3e2095d2; 3 OTLP pages plus typed-handlers drift from #103); `check:last-updated` → 155 pages ok
checked: metrics stay off on Railway only when `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT` is empty, behaviour off Railway is unchanged (test case 4), traces untouched, no new API or dependency, docs consistent across site, subsystems.md and upgrade/1.0.md
pushed: `origin/fix/railway-otlp-traces` @ 3e2095d2 (from detached wt/especialista-FW-otlp; the local branch is held by wt/senior-FW-railway-obs at e3eca299)
PR: https://github.com/AlexGalhardo/vukem/pull/104 (title `fix(otel): keep Railway OTLP export to traces only`)
