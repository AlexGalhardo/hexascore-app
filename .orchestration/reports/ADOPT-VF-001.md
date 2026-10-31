# ADOPT-VF-001 passkeys (P1.C.7.1, P1.C.7.2, P2.F.2.2b)
status: concluida
branch: agent/senior/P1.C.7.1 @ 9986a87
files: app/accounts/{passkeys,accounts,deletion,mails}.go, app/accounts/lang/*/auth.json, app/controllers/{account/passkeys,profile/profile}.go, app/http/middleware/security_headers*.go, app/mail/auth.go, app/requests/passkey.go, database/migrations/20261031120000_create_passkeys_table.go (+registry), database/legacyimport/passkeys*.go, routes/{auth,social}.go, resources/js/{lib/webauthn.ts,pages/Auth/Login*.tsx,pages/Profile/Index.tsx}, resources/lang/*.json, resources/views/mail/passkey_changed.html, packages/i18n/src/flatten.ts, qa/e2e/specs/passkeys.spec.ts, tests/passkeys_test.go, docs/sdd/{03,11,13}
acceptance:
- `go test -tags integration ./tests -run 'TestSecurity_A07_Passkey|TestSecurity_A01_PasskeyRemoveOnlyOwn|TestPasskeyRegisterLimit|TestPasskeyLoginSkipsTOTP|TestPasskeyChangeAuditedAndMailed|TestPasskeyProfileProps'` -> 9 PASS
- `go test ./database/legacyimport/... ./app/accounts/... ./app/http/middleware/...` -> ok (Permissions-Policy unit test covers publickey-credentials-create=(self))
- `bun run typecheck` ok; vitest Auth + Profile + packages/i18n 6 passed (cancelled-prompt case included)
validation: build ok, go vet ok, golangci-lint 0 issues (accounts, controllers/account, legacyimport, tests), bun lint: 4 pre-existing warnings in qa/e2e specs (not touched), browser: not run (passkeys.spec.ts is CDP virtual authenticator, left for specialist QA container)
changelog: Added: passkey registration, sign-in and removal (vukem v0.34.0), profile section and login button in en/pt-BR/es, `passkey_changed` mail, legacy passkey import stage.
roadmap: tick P1.C.7.1, P1.C.7.2, P2.F.2.2b
risks / follow-ups:
- Cherry-picks 1921bbb, 2efc7ae, b438941 from agent/implementer/P1.C.7.1; only conflict was database/migrations/registry.go. Migration renamed 20261030143000 -> 20261031120000 so it sorts after every migration on main (a prod DB that already ran later ones would otherwise skip it).
- Overlap with FIX-account-purge-lists: both edit app/accounts/deletion.go (passkey DELETE in anonymiseTx, 5 lines). `git merge-tree` of the two branches is conflict-free.
- Not exercised in a real browser/authenticator here; RP ID in production must be hexascore.app for imported legacy credentials (SDD 03 7.8). COSE/base64 legacy encoding stays [NOT VERIFIED] until proven on a Neon copy.
- Needs independent auth review (ROADMAP note).

## Review (especialista, 2026-10-31)

verdict: CORRIGIDO_E_APROVADO
re-run: `go test -tags integration ./tests -run 'TestSecurity_A07_Passkey|TestSecurity_A01_PasskeyRemoveOnlyOwn|TestPasskeyRegisterLimit|TestPasskeyLoginSkipsTOTP|TestPasskeyChangeAuditedAndMailed|TestPasskeyProfileProps'` → 9 PASS (on the branch merged with main); `go test ./database/legacyimport/... ./app/accounts/... ./app/http/middleware/...` → ok; `bunx playwright test qa/e2e/specs/passkeys.spec.ts` on QA image `hexascore-app:ADOPT-VF-001` → passkey add, sign-in and removal pass in all 6 projects (4 first attempts flaky in the shared `completeOnboarding` helper, green on retry; see follow-up 2)
checked: SDD 03 FR-03-20..24 and §7.8, SDD 13 Permissions-Policy, ADR 0017; authz (registration/removal `required, verified`; removal scoped by `user_id`; another user's id is 404), ceremonies (single-use session challenge, 5 min TTL, foreign credential, counter regression), ban and anonymised checks after the signature, session regeneration through `Web.Login`, sign-in rate limited by `LoginIP`, no key material in props/events (sha256 of the credential id), SQL bound, i18n en/pt-BR/es parity, scope
overlap with FIX-account-purge-lists: `anonymiseTx` passkey DELETE merges cleanly beside the merged purge; no purge test covered it
fixes:
1. `tests/deletion_test.go`: the purge test now proves anonymisation deletes the user's passkeys and leaves a blocked (Stripe-failing) user's row (053c426)
2. `qa/e2e/specs/passkeys.spec.ts`: removes the passkey (ROADMAP acceptance "remove it"), exact password label (the show-password toggle matched too), runs on `localhost` because WebAuthn refuses an IP RP ID and CI's default base URL is 127.0.0.1 (053c426)
follow-ups (not blocking):
1. `NoteFreshGoogle` stores a bare timestamp; bind it to the user id that completes the sign-in (low: only a passwordless user already signed in can use it within 5 min)
2. `completeOnboarding` loses the nickname fill on first attempts under load (hydration race); FIX row proposed to the orchestrator
3. Commit headers of the three senior commits have no scope (`feat: ...`); the merge commit carries `feat(auth)`
gate on main after merge: see below
PRONTO_PARA_MERGE agent/especialista/ADOPT-VF-001-fix (merged 75aa17b)
