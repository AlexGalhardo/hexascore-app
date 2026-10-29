# ADR 0017: Passkeys return through vukem `auth` (VF-001)

Status: accepted (2026-10-29). Decisions: brainstorming D-06 (amended 2026-10-29), D-05, A-04. Partially supersedes ADR 0008 (the "passkeys deferred" and "passkey credentials are not migrated" parts only).

## Context

The legacy app (Better Auth 1.7.7, `@better-auth/passkey`) lets users add passkeys in the profile and sign in with them (`legacy/apps/web/lib/auth.ts` `passkey({ rpName: APP_NAME })`, `app/[locale]/profile/passkeys-section.tsx`, `app/[locale]/login/login-form.tsx`). ADR 0008 deferred passkeys because vukem `auth` had no WebAuthn support (`https://vukem.dev/docs/security/two-factor` lists none), and planned to mail passkey-only users a reset link before cutover.

On 2026-10-29 the owner reopened framework work on VF-001 (WebAuthn in vukem `auth`, a third-party WebAuthn dependency allowed if stdlib-only is not reasonable, exact pinned stable version) and decided that passkeys come back to HexaScore (`docs/orchestration/QUESTIONS.md`, "Owner decisions (2026-10-29, in chat)").

## Decision

- Passkeys are in scope for the first vukem release of HexaScore, built on the vukem `auth` passkey API delivered by VF-001 (`install:auth --passkeys` or the documented equivalent). No app-local WebAuthn verifier and no app-level WebAuthn dependency: if the vukem release is late, the feature waits, it is not hand-rolled.
- Flows and limits follow SDD 03 (FR-03-20 to FR-03-24): register from the profile, sign in from the login page, remove from the profile, at most 10 passkeys per user, every change audited and mailed; password, TOTP and e-mail reset stay as the recovery path.
- The relying party ID is the production host (`hexascore.app`, the same host the legacy app used), so credentials registered on the legacy app stay valid after cutover.
- Legacy `passkey` rows are imported by the users stage (P2.F.2.2) when the Better Auth storage format maps onto the vukem credential table (credential id, public key, counter, transports, AAGUID, backup flags). That mapping is [NOT VERIFIED] until it is proven on a Neon copy; if it is not compatible, ADR 0008's fallback (mail passkey-only users a reset link before the window) applies.
- The Android app keeps password, Google and tokens only in v1; native passkeys (Credential Manager) are a later decision.

## Alternatives considered

- **Keep passkeys deferred (ADR 0008).** Loses a factor legacy users rely on and forces passkey-only users through a reset. Rejected by the owner.
- **App-local WebAuthn (go-webauthn in the app).** Duplicates what VF-001 adds to the framework, and a new dependency in the app needs its own ADR. Rejected.

## Consequences

- Positive: legacy parity for sign-in methods; imported credentials avoid a forced reset for passkey users.
- Negative: the feature depends on an unreleased vukem version; the importer needs a format proof on real data.
- Obligations: SDD 03 specifies flows, data and tests; ROADMAP P1.C.7 holds the build, P1.I.2.1 adds the e2e coverage with a virtual authenticator, P2.F.2.2 imports the credentials; `Permissions-Policy` changes from `publickey-credentials-get=()` (`app/http/middleware/security_headers.go`) to `publickey-credentials-get=(self)` plus `publickey-credentials-create=(self)`, as legacy `lib/security-headers.ts` sends.
