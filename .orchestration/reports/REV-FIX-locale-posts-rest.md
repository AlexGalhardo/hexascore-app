# REV-FIX-locale-posts-rest (rodada 1, especialista)
VERDICT: APROVADO. Branch agent/senior/FIX-locale-posts-rest (stacked on agent/especialista/FIX-locale-redirects-fix).
Re-run: bun typecheck 0, lint 0, test 0; TEST_DATABASE_URL=... go test -tags integration ./tests -run TestLocalePrefix -> ok 1.211s (incl. /es/lists -> /es/lists); `rg '\.(post|put|patch|delete)\(\s*[`"]/' resources/js` -> no hits.
Review: mechanical, correct; every write goes through useLocalized/localized. Not re-run: playwright (senior: 72 passed).
Merge note: overlaps agent/especialista/P1.I.2.4-fix on AddToList.tsx (var `loc` vs `l10n`), Lists/Index.tsx and Lists/Edit.tsx (identical edits) -> trivial conflicts, keep either. Order: FIX-locale-redirects-fix -> FIX-locale-posts-rest -> P1.I.2.4-fix (or P1.I.2.4-fix before posts-rest; both fine).
