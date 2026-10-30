# REV-FIX-locale-redirects (rodada 1, especialista)
VERDICT: CORRIGIDO_E_APROVADO. Final branch agent/especialista/FIX-locale-redirects-fix (692366a, on top of agent/senior/FIX-locale-redirects).
Found and fixed:
1. TestLocalePrefix_AccountRedirects used pgtest in an untagged file: `go test ./tests` without TEST_DATABASE_URL -> FAIL "pgtest: no PostgreSQL server". Moved to tests/locale_redirects_test.go (//go:build integration).
2. Senior changed TestLocalePrefix_Routing to expect "/pt-br" for the cookie redirect; server answers "/pt-br/" (main passes, branch failed: `redirect to "/pt-br/", want "/pt-br"`). Reverted the assertion. Senior's claim of a green integration run did not hold for this test.
Re-run: go test ./tests -run TestLocalePrefix -> ok (no DB); TEST_DATABASE_URL=... go test -tags integration ./tests -run TestLocalePrefix -> Theme/Routing/AccountRedirects PASS; golangci-lint --build-tags integration ./tests/... 0 issues; bun lint 0, typecheck 0, test 0; isolated stack hx-especialista playwright auth profile api-keys -> 77 passed, 1 flaky (mobile-en profile change password, passed on retry), exit 0.
Review: useLocalized is minimal and correct; client posts now carry the page prefix; e2e specs assert the page locale instead of anyLocale. No security impact.
Merge note: P2.D.2.3 (agent/especialista/P2.D.2.3) also edits Auth/Login, Signup, Recover and qa/e2e/helpers/account.ts -> small textual conflicts; keep both changes.
Follow-up (not done, out of scope per senior): posts in items/lists/contact/notifications/library still unprefixed.
