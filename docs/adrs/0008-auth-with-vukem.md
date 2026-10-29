# ADR 0008: Authentication with vukem `auth` + `social`; passkeys deferred

Status: accepted (2026-10-28). Decisions: brainstorming D-05, D-06, A-01, A-04.

## Context

Legacy auth is Better Auth 1.7.7: email + password with required verification, Google OAuth, TOTP with backup codes, passkeys, username plugin, admin plugin (roles, bans, impersonation), database-backed rate limiting, BotID checks on auth routes. Tables `user`, `session`, `account`, `verification`, `two_factor`, `passkey`, `rate_limit`.

vukem (`https://vukem.dev/docs/security/two-factor`) provides guards, a user provider, argon2id/bcrypt hashers, sessions, `install:auth` scaffold with `--2fa` (TOTP), social login (OAuth2/OIDC, Ally-style), access tokens with abilities for APIs (`https://vukem.dev/docs/security/api-tokens` "Authentication"), authorization gates and policies, and login rate limiting. Passkeys/WebAuthn are not in the docs. `auth` and `social` are marked ALPHA.

## Decision

Use vukem `auth` for sessions, password hashing (argon2id), email verification, password reset and TOTP 2FA; vukem `social` for Google; vukem access tokens with abilities for the mobile app and first-party write API (A-01). Roles and bans are columns on the user model enforced by gates; admin actions go through the vukem admin panel with its audit log. Passkeys are deferred (D-06) and filed in `vukem-feedback.md` as a feature request. Bot protection for signup/login replaces BotID with a server-side option specified in SDD 03/13 (rate limits plus a challenge only when thresholds trip).

Migration from Better Auth (A-04): users, OAuth account links and 2FA secrets are migrated; password hashes are migrated as-is with their algorithm tag and verified by a compatibility hasher on first login, then re-hashed with argon2id; all sessions are invalidated at cutover; passkey credentials are not migrated (feature deferred) and users are told.

## Alternatives considered

- **Better Auth as a separate Node service.** Keeps passkeys and the current flows; adds a second runtime, a second session model and a cross-service trust boundary. Rejected.
- **External IdP (Clerk, Auth0, Keycloak hosted).** Less code, but monthly cost per active user, vendor dependency for the most sensitive data, and a new dependency needing an ADR. Rejected.
- **Hand-rolled auth on `net/http`.** No reason to bypass vukem's tested defaults. Rejected.

## Consequences

- Positive: one runtime; security defaults covered by `TestSecurity_*`; tokens with abilities give the mobile app least privilege.
- Negative: ALPHA APIs may change before vukem v1; passkey users lose that factor until vukem supports it; impersonation may not exist in vukem (SDD 03 checks and designs the minimum if absent).
- Obligations: SDD 03 specifies tables, flows, migration and tests; `vukem-feedback.md` gets entries for passkeys, impersonation (if missing) and any migration-hasher gap.
