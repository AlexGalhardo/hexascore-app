REV P1.G.3.2d: CORRIGIDO_E_APROVADO. Final branch: agent/especialista/P1.G.3.2d-fix (on top of agent/senior/P1.G.3.2d).
Reviewed OK: limit+1 probe for next, OFFSET from ParseOffset (non-negative int, not raw input), item page inline list unchanged, i18n 3 locales.
Fixed: double click on "choose"/"more editions" fetched the same page twice and appended duplicate options (duplicate React keys) -> useRef in-flight guard; regression test in EditionPicker.test.tsx (red without fix: 1 failed, green with), plus RTL cleanup between tests.
Re-run: go test -tags integration ./tests -run 'TestEditionPickerList|TestEditionNotOfWork|TestInertiaPages_Catalog' -> all PASS; golangci-lint -> 0 issues; bun lint/typecheck clean, bun run test 58 passed.
