STATUS: concluida
Branch: agent/senior/FIX-locale-redirects (pushed). Files: resources/js/lib/useLocalized.ts (new), forms in Auth/{Login,Signup,Recover,Reset,Onboarding,TwoFactor,TwoFactorChallenge}, Developer/ApiKeys, Profile/Index, components/DeleteAccount (+test mock), tests/locale_prefix_test.go, qa/e2e (helpers/account.ts, mail.ts, profile/api-keys/auth specs), docker-compose.yml (E2E_IMAGE_TAG), CHANGELOG.
Root cause: server redirects already keep the prefix (localizedRedirects middleware), but the React forms posted to unprefixed URLs, so the request had no locale. Fix is client-side: post under the page locale. Login/Signup/Recover/Reset/Onboarding included (same defect).
Acceptance: go test -tags integration ./tests -run TestLocalePrefix_AccountRedirects ok (login->2FA challenge, password, api key, deletion under /pt-br and /es); go build ok; golangci-lint --build-tags integration 0 issues; bun lint/typecheck ok, bun test 59 passed.
E2E (hx-senior, image hexascore-app:e2e-senior): playwright auth profile api-keys run twice -> 78 passed, exit=0, both runs.
Notes: base lacked the mail.ts [REDACTED] fix from E2E-HARDEN (unmerged), so I included the identical 1-line change (merge-compatible). Post-mail-link steps in createVerifiedAccount use anyLocale (mailed links are unprefixed English).
Not covered: posts in items/lists/contact/notifications/library (same pattern, out of scope; follow-up candidate), OAuth callback redirect (no locale info).

REVIEW (especialista): CORRIGIDO_E_APROVADO, see reports/REV-FIX-locale-redirects.md
PRONTO_PARA_MERGE agent/especialista/FIX-locale-redirects-fix
