STATUS: concluida
Branch: agent/senior/FIX-locale-posts-rest (stacked on origin/agent/especialista/FIX-locale-redirects-fix; useLocalized not yet on main)
Files: Notifications, Contact/Index, EditionPicker(+test mock), AddToList, TimeToBeat, StatusPicker, RatingPanel (useLocalized); Library/Index, Lists/Index, Lists/Edit (localized(...,locale)); tests/locale_redirects_test.go; CHANGELOG.
Acceptance: typecheck rc=0, biome ci rc=0, vitest 59/59; go test -tags integration TestLocalePrefix_AccountRedirects ok (list create /es/lists -> /es/lists); playwright contact/item/ratings/time-to-beat on hx-senior: 72 passed (no lists spec exists).
Remaining literal hits in rg: none outside useLocalized/localized wrappers (billing/reports/export have no client posts).
Risks: merge order, needs especialista fix branch first.

REVIEW (especialista): APROVADO, see reports/REV-FIX-locale-posts-rest.md
PRONTO_PARA_MERGE agent/senior/FIX-locale-posts-rest
