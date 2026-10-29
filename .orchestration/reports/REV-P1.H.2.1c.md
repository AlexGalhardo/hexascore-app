# REV-P1.H.2.1c
VERDICT: CORRIGIDO_E_APROVADO. Final branch agent/especialista/P1.H.2.1c-fix (8369936, from origin/main; 1c already merged).
FR-05-16: OK (lookup body = getItem serializer; tmdb/igdb ids never echoed; test asserts). Search uses AudienceAPI (entity names only from allowlisted name_source, no default_name); kind validated by searcher -> 400.
Fixed in LookupItemID:
1. No visibility filter + ORDER BY id LIMIT 1: a hidden item sharing a third-party id shadowed the public one -> 404. Now `visibility = 'public'`. Regression test (red on old code: 404 want 200).
2. One WHERE with ORs incl. EXISTS(editions) defeats GIN/unique indexes (full items scan per lookup). Now one indexable branch per source (UNION ALL); isbn13 via `$2::char(13)` with length guard against truncation.
Acceptance (re-run): `go test -count=1 -tags integration ./tests -run 'TestAPISearch|TestAPILookup|TestAPIContract'` -> ok; `golangci-lint run --build-tags integration ./...` -> 0 issues; markdownlint CHANGELOG 0 issues.
Merge: CHANGELOG conflict with P1.H.2.1b-fix (keep both lines).
