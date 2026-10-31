# FIX-mobile-catalogues Recopy mobile i18n catalogues
status: concluida
branch: agent/estagiario/FIX-mobile-catalogues @ c1b7990 (base main 246ced8)
files: apps/mobile/src/i18n/catalogues/{en,pt-BR,es}.json
acceptance:
- `bun run --cwd apps/mobile catalogues` → regenerated copy; semantic diff vs HEAD: 1 key added per locale (`account.privacy.save`), 0 values changed
- `bun run --cwd apps/mobile test -- keys` → "Tests: 4 passed, 4 total", EXIT=0
validation: lint not run (no code change), typecheck not run, tests apps/mobile keys ok, browser n/a
changelog: Fixed: mobile i18n copy now includes account.privacy.save (keys test green)
roadmap: no box
risks / follow-ups: the script writes keys in its own order, so the JSON diff is about 845 lines of reordering on top of one real key. Semantic change is one key only (checked with JSON parse, not textual diff).

## Review (especialista, 2026-10-31)

verdict: CORRIGIDO_E_APROVADO
re-run: `bun run --cwd apps/mobile catalogues` after merging main → 15 new keys per locale (`account.passkeys.*`, `auth.passkey`, `auth.errors.passkeyFailed`; the passkey merge stopped dropping them), 0 removed, 0 values changed (JSON parse diff); `bun run --cwd apps/mobile test` → 6 suites, 41 passed (keys test green)
checked: semantic diff per locale (sampled, not textual: the script reorders keys), no hand edits, scope `apps/mobile/src/i18n/`
fixes: 1. recopied the catalogues on top of current main (fix branch commit "recopy mobile catalogues with the passkey keys")
gate on main after merge: see the session gate line
PRONTO_PARA_MERGE agent/especialista/FIX-mobile-catalogues-fix (merged a3006bb)
gate on main @ 57497a9 (contains this merge): go build ok, golangci-lint --build-tags integration 0 issues, bun lint 0 errors / typecheck ok / vitest 70 passed, go test -tags integration -parallel 4 ./... ok. Earlier run @ 6181cb4: internal/search TestGolden failed once under load (QA build + e2e in parallel), passed alone and @ 57497a9; @ 75aa17b TestSecurity_A01_HexaRouteStackInventory failed on the 6 unclassified passkey routes, fixed in 6f25b4e.
