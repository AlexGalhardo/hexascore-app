# FIX-mobile-rating-a11y-now integer accessibility value on the rating input

status: concluida
branch: agent/senior/FIX-mobile-rating-a11y-now @ e0201e8 (from main fcf63bf)
files: `apps/mobile/src/ui/RatingInput.tsx`, `apps/mobile/src/ui/ui.test.tsx`
acceptance:
- red: `bun run --cwd apps/mobile test -- src/ui/ui.test.tsx -t "adjustable|clamps|6.5"` → 3 failed
- green: `bun run --cwd apps/mobile test -- src/ui/ui.test.tsx src/rating` → 28 passed (6.5 → `now: 13`, 8.5 → `now: 17`)
validation: `bunx biome check` on both files ok, `bun run typecheck` ok; no device run
changelog: Fixed: TalkBack reads the Android rating control's position correctly (integer half steps instead of a fractional value).
roadmap: no box
risks / follow-ups:
- `accessibilityValue` is now `{min: 1, max: 20, now: value * 2, text}`; `now` is omitted while there is no rating (it would sit below `min`), `text` ("6,5 / 10", "– / 10") is unchanged and is what TalkBack announces.
- `StarInput` (P5.B.2.2) wraps `RatingInput`, so the rate sheet gets the fix without changes.

## Review (especialista, 2026-11-01)
verdict: APROVADO
re-run: `scripts/openapi-check.sh` -> ok (golden regenerated with `--write`, client with `codegen:client -prefix /app/v1`, 23d1177); `go test -tags integration -run 'TestAppCatalog$|TestCoversBaseJSONAbsolute|TestAPITrending|TestAPIItems|TestRouteInventory|TestOpenAPI|TestAppSearch|TestSearch|Suggest|TestAPIEdge' ./tests/` -> ok; golangci-lint (touched pkgs, integration tag) 0 issues; app-client typecheck/test ok; mobile lint/typecheck ok, mobile Jest 297 tests (3 suites timed out under load on the full run, green alone); root lint/typecheck/vitest 129 ok
checked: sdd 05 (paid API keeps no covers, slug lookup app-only), sdd 10 FR-10-7, authz (scopes unchanged), sql (bound params), i18n (catalogues recopied), scope, tests

gate on main after merge: local.sh test-go, test-go-integration, test-ts, migrations, vuln, deps, setup-windows, mobile, build PASS on bbe689d/fcf63bf; lint fails only on commitlint of historic local commits; lighthouse fails on main before this batch (FIX-lighthouse-budgets-main)
PRONTO_PARA_MERGE agent/especialista/FIX-app-batch (merged bd95e7d)
