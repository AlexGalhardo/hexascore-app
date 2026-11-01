# SDD 03: Users and authentication

Status: reviewed (2026-10-28 consistency pass). Decisions: D-05, D-06 (amended 2026-10-29), D-12, D-15, A-01, A-04. ADRs: 0008, 0017. Related SDDs: 05, 09, 10, 11, 13.

## 1. Objective

One account for web (session cookie) and Android (access token): email or Google sign-up, verified email, optional TOTP, public nickname, private profile, LGPD export and deletion. Built on vukem `install:auth --2fa --social google`, access tokens and the admin panel, with legacy Better Auth behaviour as the parity baseline; every real legacy account migrates without a forced reset.

## 2. Scope

In scope (v1): account model, email verification, password rules, reset, TOTP + recovery codes, Google, sessions, mobile access tokens (A-01, D-12), roles and bans (admins only, D-15), deletion with 30-day grace, LGPD export, transactional mail in en/pt-BR/es, bot protection, migration from Better Auth (A-04).

In scope since 2026-10-29 (D-06 amended, ADR 0017): passkeys on the web, built on the vukem `auth` passkey API of VF-001 (§7.8); not in the Android app in v1.

Out of scope: admin impersonation (not in v1, VF-012; no legacy UI used it; only the `session.impersonated_by` column exists, `legacy/apps/web/lib/db/schema/auth.ts`; admins use a read-only "view as" in the admin panel meanwhile), email change (not offered by legacy), sign-up inside the mobile app (D-12 lists login only; the app opens `/signup` in the browser), paid API keys (SDD 05).

## 3. Functional requirements

Each line names its test (Go unless stated).

- FR-03-1: Sign-up (email, password, display name, nickname) answers alike for new and existing emails (vukem "`install:auth` scaffold", Register). `TestSecurity_A07_SignupIsEnumerationSafe`.
- FR-03-2: Validation reuses `legacy/apps/web/lib/validation.ts`: password 8–32 chars (D-24, owner round 2026-11-01) with upper, lower, digit, non-alphanumeric; nickname `^[a-z]{4,16}$` after trim + lowercase; sign-up email 12–48, other emails ≤ 254; the sign-up form also takes the nickname (optional server-side, same rule and codes as onboarding; a taken nickname is checked before the email so the answer does not reveal accounts); name 1–100. Password rules apply to sign-up, reset, change, set; never to login. Table test on `app/requests/auth.go`.
- FR-03-3: Login takes email or nickname (contains `@` → email); unknown identifier and wrong password give the same 422. `TestSecurity_A07_LoginErrorsDoNotRevealUsers`, `TestLoginByNickname`.
- FR-03-4: Unverified users sign in and browse; every write, token, 2FA and export needs `auth.Verified`. Generated `TestUnverifiedUsersSeeTheNotice`.
- FR-03-5: Verification link signed, 1 h, bound to `email_hash`, needs the account's login; resend 3/min; verifying ends other sessions and tokens. `TestSecurity_A07_VerificationLinkRequiresLogin`.
- FR-03-6: Reset: same answer for unknown emails; SHA-256 token, 1 h, single use; success bumps `session_version`, deletes sessions and remember tokens, `Tokens.RevokeAll` (legacy `revokeSessionsOnPasswordReset: true`). `TestSecurity_A07_PasswordResetAnswersAlike`.
- FR-03-7: Password change needs the current password and ends other sessions and tokens; Google-only accounts may set one once. `TestSetPasswordOnlyWhenUnset`.
- FR-03-8: TOTP per vukem "Two-factor authentication" (<https://vukem.dev/docs/security/two-factor>); Google-only accounts (`password_set_at` NULL) must set a password before enabling or managing 2FA and the 2FA section explains why (D-25); 8 single-use recovery codes shown once, replay refused, challenge TTL 5 min, disable/regenerate need the password. `TestSecurity_A07_TwoFactorScaffoldDefaults`.
- FR-03-9: Google via `FindOrCreateFromSocial`; users without a nickname go to `/onboarding` before any write (legacy `lib/member.ts` `asMember`). `TestGoogleNewUserNeedsNickname`.
- FR-03-10: Nickname set once, never changed (legacy `actions/account.ts`); name, locale, privacy editable. `TestNicknameImmutable`.
- FR-03-11: Private profiles answer 200 with the nickname only and `noindex` to everyone but the owner on web and app API (legacy parity, open issue 14-5; page owned by SDD 14). Integration test.
- FR-03-12: Profile lists sessions (`Manager.UserSessions`) and tokens (`Tokens.List`), each revocable; "sign out everywhere" bumps `session_version` and revokes all tokens. `TestSecurity_A01_RevokeOnlyOwnSessions`.
- FR-03-13: Mobile tokens by password (+2FA) or Google ID token, named per device, expiring, rotatable, revocable (§7.4). §9.
- FR-03-14: Admins ban (reason, optional expiry) and unban in the panel; a ban ends sessions and tokens and, after commit, dispatches `user.banned {userId, at}` (unban and expiry clearing dispatch `user.unbanned {userId, at}`), consumed by SDD 02 to recompute the user's rated items (open issue 02-8); `admin` role only by CLI. `TestSecurity_A07_BanEndsSessionsAndTokens`, `TestBanDispatchesUserBanned` (`events.NewFake()`).
- FR-03-15: Deletion: 10 s countdown (legacy `DELETION_CONFIRM_SECONDS`), linked from the public `/help/delete-account` page required by Google Play (SDD 09, open issue 10-7), sign out everywhere, 30-day grace with cancel, then anonymisation keeping ratings. `TestDeletionPurgeAnonymisesKeepsRatings`.
- FR-03-16: LGPD export: a background ZIP job owned by SDD 14 §7.4 (24 h signed link, 7-day cooldown, open issue 14-4); this SDD defines the personal-record content (§6.4). `TestExportContainsAllPersonalTables` (SDD 14 `TestExportZip`).
- FR-03-19: `users.personalization_opt_out` is exposed in the profile privacy section; personalised recommendations are on by default and the toggle opts out (D-29, open issue 08-4); SDD 08 reads it. Integration test on `PATCH /profile/privacy`.
- FR-03-17: Mail queued and rendered in the user's locale (en, pt-BR, es). `TestMailTranslatedAllLocales`.
- FR-03-18: Legacy scrypt passwords log in and are rehashed to argon2id. `TestSecurity_A07_LegacyScryptVerifiesAndRehashes`.
- FR-03-20: A verified user adds a passkey from the profile (`/profile` section `passkeys`) after confirming the current password (accounts with `password_set_at` NULL confirm with a fresh Google sign-in [NOT VERIFIED: vukem confirm flow for passwordless accounts]); user verification required; name ≤ 64 chars (default "Passkey"; vukem v0.34.0 Credential has no AAGUID); at most 10 per user (409 `passkey-limit`). `TestPasskeyRegisterLimit`.
- FR-03-21: Sign in with a passkey from `/login` (discoverable credential, "Sign in with a passkey" button); the same ban and deletion checks as a password login; a passkey login satisfies 2FA (possession + user verification), as Better Auth did. `TestPasskeyLoginSkipsTOTP`.
- FR-03-22: A user lists and removes only their own passkeys; removal needs a verified session. `TestSecurity_A01_PasskeyRemoveOnlyOwn`.
- FR-03-23: Ceremonies refuse a credential of another user, a replayed or expired challenge (single use, 5 min, bound to the session), a wrong origin or RP ID, and a signature counter that goes backwards (unless both are 0). `TestSecurity_A07_PasskeyForeignCredential`, `TestSecurity_A07_PasskeyChallengeReplay`, `TestSecurity_A07_PasskeyCounterRegression`.
- FR-03-24: Adding or removing a passkey writes `account_events` (`passkey.added`, `passkey.removed`, credential id hash only) and mails `passkey_changed`; recovery for a lost passkey is password (+ TOTP) or e-mail reset, never a passkey-specific bypass. `TestPasskeyChangeAuditedAndMailed`.

## 4. Non-functional requirements

- NFR-03-1 Latency: login p95 ≤ 250 ms (argon2id dominates); `GET /app/v1/me` p95 ≤ 50 ms. k6 `qa/k6/auth.js`.
- NFR-03-2 Memory: ≤ 4 concurrent hashes (semaphore; 64 MiB each) on the single instance (D-08). k6 50 VUs on `/login`, RSS ≤ 512 MiB.
- NFR-03-3 Rate limits (store `RATE_LIMIT_STORE=cache` on `CACHE_STORE=database`, so a deploy does not reset buckets; memory refuses the shared mode, vukem `https://vukem.dev/docs/security/overview` A02):

| Flow | Limit | Source |
| --- | --- | --- |
| Login (web and token) | 5/min per identifier + IP; 20/10 min per account | vukem A07 |
| Sign-up | 10/10 min per IP; 5/h per address hash | vukem scaffold |
| Verification resend | 3/min | vukem scaffold |
| Forgot password | 5/h per address hash; 10/10 min per IP | HexaScore value; vukem only states "throttled by default" |
| 2FA challenge/confirm/disable/codes | 5/min per user + IP | vukem `--2fa` |
| Export, deletion schedule | export: 1 per 7 days (SDD 14 cooldown); deletion: 5/h per user | HexaScore value |
| Token rotate | 10/h per user | HexaScore value. A per-token bucket resets when rotate replaces the bearer, so the cap is the user. |

- NFR-03-4 Cookies: session cookie `__Host-` prefix, `HttpOnly`, `Secure`, `SameSite=Lax` (vukem A07 `TestSecurity_A07_SessionCookieFlags`); `SESSION_DRIVER=database` (needed for `RevokeOthers`; cookie sessions return `ErrRevokeUnsupported`); idle timeout 7 days, absolute 30 days, remember-me 30 days (`RememberFor` default). Env names for idle/absolute timeouts are [NOT VERIFIED] (vukem `https://vukem.dev/docs/basics/sessions` "Sessions" names the feature, not the variables).
- NFR-03-5 CSRF: `http.CrossOriginProtection` on all unsafe methods (vukem `https://vukem.dev/docs/basics/csrf` A01), token fallback enabled for old browsers; `/app/v1/*` is bearer-only, so the guard never accepts cookies there.
- NFR-03-6 Secrets at rest: passwords argon2id PHC; 2FA secret and recovery-code hashes `orm.Encrypted` with `APP_KEY` (`https://vukem.dev/docs/database/orm` "Encrypted columns"); tokens and reset tokens SHA-256. No secret column is shown in the admin panel (`https://vukem.dev/docs/digging-deeper/admin` "Security").
- NFR-03-7 Privacy: `SESSION_RECORD_CLIENT=true` (UA ≤ 256 bytes, IP truncated /24 or /48); request and audit logs keep IPs 30 days (vukem A09 marks retention as app responsibility).
- NFR-03-8 Accessibility: axe 0 serious/critical on auth/profile pages; 2FA input `autocomplete="one-time-code"`. Playwright + axe.
- NFR-03-9 Availability: store outages answer 503, sessions kept (`TestSecurity_A04_SessionKeptWhenUserLookupFails`).

## 5. Data model

The `install:auth` scaffold owns its migrations (user-owned code); HexaScore edits them before the first run. Column names below are HexaScore's; where the scaffold uses another name the scaffold is renamed, not the spec [NOT VERIFIED: exact scaffold column names].

```sql
CREATE TABLE users (
  id                     bigint PRIMARY KEY,          -- vukem scaffold identity (amended 2026-10-29)
  legacy_id              text UNIQUE,                 -- Better Auth id of migrated users; NULL for new rows
  email                  text NOT NULL,
  email_verified_at      timestamptz,
  password               text NOT NULL,               -- argon2id PHC | bcrypt | legacy scrypt "salt:key"
  password_set_at        timestamptz,                 -- NULL: random password (Google-only account)
  name                   text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 100),
  username               text CHECK (username ~ '^[a-z]{4,16}$'),
  display_username       text,
  image_url              text,                        -- from Google only; no uploads in v1
  locale                 text NOT NULL DEFAULT 'en' CHECK (locale IN ('en','pt-BR','es')),
  profile_private        boolean NOT NULL DEFAULT false,
  personalization_opt_out boolean NOT NULL DEFAULT false,  -- D-29: personalised recs on by default (08-4)
  role                   text NOT NULL DEFAULT 'user' CHECK (role IN ('user','admin')),
  banned_at              timestamptz,
  ban_reason             text CHECK (char_length(ban_reason) <= 500),
  ban_expires_at         timestamptz CHECK (ban_expires_at IS NULL OR banned_at IS NOT NULL),
  deletion_scheduled_at  timestamptz,
  anonymized_at          timestamptz,
  session_version        bigint NOT NULL DEFAULT 0,   -- auth.SessionVersioner
  two_factor_secret          text,                    -- orm.Encrypted[string]
  two_factor_recovery_codes  text,                    -- orm.Encrypted[[]string] of SHA-256
  two_factor_confirmed_at    timestamptz,
  two_factor_last_step       bigint,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX users_email_uq    ON users (lower(email));
CREATE UNIQUE INDEX users_username_uq ON users (username) WHERE username IS NOT NULL;
CREATE INDEX users_deletion_due_idx   ON users (deletion_scheduled_at) WHERE deletion_scheduled_at IS NOT NULL;

CREATE TABLE account_events (               -- user-facing security history and admin actions
  id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  user_id    text NOT NULL REFERENCES users(id),
  actor_id   text REFERENCES users(id),     -- NULL = the user or the system
  kind       text NOT NULL,                 -- password_changed, 2fa_enabled, 2fa_disabled, 2fa_reset,
                                            -- recovery_code_used, banned, unbanned, role_granted,
                                            -- deletion_scheduled, deletion_cancelled, anonymized, exported
  data       jsonb NOT NULL DEFAULT '{}',   -- ids and reasons only, never emails or IPs
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX account_events_user_idx ON account_events (user_id, created_at DESC);
```

vukem-owned tables reused unchanged: `sessions` (framework, `https://vukem.dev/docs/database/overview` "Tables the framework owns"), `remember_tokens`, `access_tokens` (with `revoked_at`), password reset tokens and `social_accounts` (unique provider + provider id) from `install:auth` / `--social`, `cache`, `jobs`, `failed_jobs`, admin audit log.

Invariants: `username` is stored lowercase (case folding is the identity rule; the legacy pattern already allows only `a-z`); emails compared by `lower()`; a banned user has `banned_at` set; `anonymized_at` set implies `email = 'deleted-' || id || '@invalid'`, `username IS NULL`, 2FA columns NULL. Retention: `account_events` 2 years, then deleted by a monthly job; revoked/expired tokens pruned after 90 days (`Tokens.Prune`).

Amendment (2026-10-29, P1.C.1.1): `users.id` stays the vukem scaffold's `bigint` identity instead of `text`. Rewriting every `install:auth` file and the SQLite-based scaffold tests for text ids bought nothing the `legacy_id` column does not give: the data migration (P2.F) stores the Better Auth id in `legacy_id` and maps every legacy `user_id` foreign key through it. `account_events.user_id` and `actor_id` are `bigint` accordingly. Scaffold facts recorded by P0.F.1.1: session env names `SESSION_LIFETIME`, `SESSION_IDLE_TIMEOUT`, `SESSION_DRIVER`, `SESSION_COOKIE`, `SESSION_SECURE`, `SESSION_RECORD_CLIENT` (open issue 03-7); pages are TypeScript `Auth/*.tsx` reusing the starter layout (03-10). Paths follow §6.1 (`/signup`, `/signup/check-email`, `/recover-password`, `/verify-email`, `/verify-email/confirm/{id}`, `/verify-email/resend`).

Migration mapping from `legacy/apps/web/lib/db/schema/auth.ts` (A-04; the 10k synthetic users are purged first):

| Legacy | Target | Rule |
| --- | --- | --- |
| `user.id` | `legacy_id` | new `bigint` id assigned; every legacy `user_id` FK is mapped through `legacy_id` |
| `user.name, email, image, created_at` | same columns | email lowercased; collision check below |
| `user.email_verified` | `email_verified_at` | `true` → `updated_at`; `false` → NULL |
| `user.username, display_username` | same | already lowercase and unique |
| `user.role` (comma list) | `role` | contains `admin` → `admin`, else `user` |
| `user.banned, ban_reason, ban_expires` | `banned_at` (= migration time), `ban_reason`, `ban_expires_at` | expired bans dropped |
| `user.profile_private, deletion_scheduled_at` | same | grace clock preserved |
| `account` with `provider_id='credential'` | `users.password`, `password_set_at` | hash copied verbatim |
| `account` with `provider_id='google'` | `social_accounts(provider='google', provider_id=account_id)` | OAuth access/refresh/id tokens dropped |
| No credential row | `password` = random argon2id, `password_set_at` NULL | |
| `two_factor` | `two_factor_*` | §7.6 |
| `session`, `verification`, `rate_limit` | dropped | cutover signs everyone out |
| `passkey` | `passkeys` (§7.8) | users-stage step: `credential_id`, `public_key`, `counter`, `transports`, `aaguid`, `device_type`/`backed_up` → backup flags, `name`, `created_at`; `user_handle` stays NULL until the first successful assertion pins it (Better Auth 1.7.7 does not store the handle); `algorithm` is COSE map key 3. Incomplete rows are skipped. If a real dump cannot be verified, ADR 0008's fallback still applies: passkey-only users mailed a reset link 14 days before cutover |

Pre-flight asserts (abort on non-zero): `SELECT lower(email) FROM "user" GROUP BY 1 HAVING count(*) > 1`; users with `two_factor_enabled` and no `two_factor` row; users with neither credential nor Google nor email (passkey-only accounts are counted and sent the reset notice).

Implementation (P2.F.2.2, `database/legacyimport/users.go`): batches keyed by the legacy user id upsert on `legacy_id`; `email_verified_at` takes the legacy `updated_at`, or `created_at` when that is NULL; `username` is lowercased on the way in; every user without a credential hash (Google-only and passkey-only) gets its own random argon2id; Google links are inserted with `ON CONFLICT (provider, provider_id) DO NOTHING`. Passkeys are imported by the users stage (`database/legacyimport/passkeys.go`, P2.F.2.2b), not a separate SDD 11 stage. Incomplete rows are skipped.

## 6. Contracts

### 6.1 Web routes (Inertia pages; localised under `/pt-br`, `/es` per D-17; legacy paths kept, scaffold paths renamed in the user-owned `routes/auth.go`)

| Method | Path | Auth | Page / effect |
| --- | --- | --- | --- |
| GET, POST | `/login` (name `login`) | guest | `Auth/Login`; POST `{identifier, password, remember, challenge?}` |
| GET, POST | `/signup` | guest | `Auth/Signup`; POST → `/signup/check-email` |
| GET | `/signup/check-email` | guest | `Auth/CheckEmail` |
| GET | `/verify-email` | session | `Auth/VerifyNotice`; POST `/verify-email/resend` |
| GET | `/verify-email/confirm?...signature` (name `verification.verify`) | session | verifies, redirect `/profile` |
| GET, POST | `/recover-password` | guest | `Auth/Recover` |
| GET, POST | `/reset-password/{token}` | guest | `Auth/Reset` |
| GET, POST | `/two-factor-challenge` | pending login | `Auth/TwoFactorChallenge` `{code}` or `{recoveryCode}` |
| GET, POST | `/onboarding` | session, verified, no nickname | `Auth/Onboarding` `{nickname}` |
| GET | `/auth/google/redirect`, `/auth/google/callback` | guest | vukem `social` |
| POST | `/logout` | session | |
| GET | `/profile` | session | `Profile/Index` sections: details, password, twoFactor, passkeys, sessions, privacy, export, delete |
| PATCH | `/profile/details` | verified | `{name, locale}` |
| PUT | `/profile/password` | verified | `{currentPassword?, newPassword}` (current required when `password_set_at` not NULL) |
| GET, POST | `/user/two-factor`, `/user/two-factor/confirm`, `/user/two-factor/recovery-codes`, `/user/two-factor/disable` | verified | vukem `--2fa` routes unchanged |
| POST | `/login/passkey/options`, `/login/passkey` | guest | discoverable sign-in; satisfies 2FA |
| POST | `/user/passkeys/options`, `/user/passkeys` | verified | registration options and verify |
| DELETE | `/profile/passkeys/{id}` | verified | remove one of the caller's passkeys |
| GET | `/auth/google/confirm` | verified | fresh Google sign-in when `password_set_at` is NULL |
| PATCH | `/profile/privacy` | verified | `{profilePrivate, personalizationOptOut}` |
| DELETE | `/profile/sessions/{handle}`; POST `/profile/sessions/revoke-all` | session | |
| POST | `/profile/export` | verified | creates an export job (SDD 14 §6, §7.4); ZIP link mailed, 24 h |
| POST, DELETE | `/profile/deletion` | session | schedule (`{confirm:true}`) / cancel |
| GET | `/help/delete-account` | none | public page (SDD 09) explaining in-app and web deletion, linking `/profile#delete` |

Shared Inertia prop `auth.user`: `{id, name, username, locale, role, emailVerified, twoFactorEnabled, passwordSet, profilePrivate, personalizationOptOut, deletionScheduledAt}`; never secrets (the `TwoFactor` model is separate, vukem `--2fa` Storage row).

### 6.2 Mobile token API (`/app/v1`, problem+json, OpenAPI tag `app-auth`)

| Method | Path | Auth | Input → output | Errors |
| --- | --- | --- | --- | --- |
| POST | `/app/v1/tokens` (`createAppToken`) | none | `{identifier, password, deviceName, code?, recoveryCode?, challenge?}` → 201 `{token, tokenId, expiresAt, abilities}` | 422 `invalid-credentials`, 422 `two-factor-required`, 403 `email-not-verified`, 403 `account-banned` `{reason, expiresAt}`, 428 `challenge-required`, 429 |
| POST | `/app/v1/tokens/google` (`createAppTokenFromGoogle`) | none | `{idToken, deviceName, code?, recoveryCode?}` → 201 as above | 401 `invalid-id-token`, 409 `email-unverified-at-provider`, rest as above |
| POST | `/app/v1/tokens/rotate` (`rotateAppToken`) | bearer | → 201 new token; current revoked | 401 |
| GET | `/app/v1/tokens` (`listAppTokens`) | bearer `catalog:read` | → `[{id, name, lastUsedAt, expiresAt, current}]` | |
| DELETE | `/app/v1/tokens/current`, `/app/v1/tokens/{id}` | bearer | 204 (`RevokeFor`, owner-checked) | 404 |
| GET, PATCH | `/app/v1/me` (`getMe`, `updateMe`) | bearer `catalog:read` / `profile:write` | `{name, locale, profilePrivate, personalizationOptOut, username?}` (username only when unset) | 422 `nickname-taken` |
| POST | `/app/v1/me/export` (`exportMe`) | bearer `profile:write` | → 202 `{exportId, status}` (SDD 14 §6.3) | 429 `export-cooldown` |
| POST, DELETE | `/app/v1/me/deletion` | bearer `profile:write` | schedule / cancel | |

Token format `hx_app_<id>.<secret>` (`Tokens.Prefix = "hx_app_"`, vukem "Token guard"); abilities per the README convention: `catalog:read`, `ratings:write`, `status:write`, `lists:write`, `profile:write`, `suggestions:write` (an app token gets all six; never `*`; no `app:read`/`app:write` umbrella, open issues 05-5, 10-9). Paid API keys (`hx_live_`, SDD 05) carry only `catalog:read` and are refused on `/app/v1`; app tokens are refused on `/api/v1`. Problem codes are kebab-case (05-9). Route stack: `auth.Required(apiGuard)`, `NotBanned`, `auth.Verified(apiGuard)`, `auth.Ability(...)`.

### 6.3 Admin and CLI

- Panel resource `users` (`admin.NewResource[models.User]`): list/search on `email`, `username`, filters `role`, `banned`; editable keys only `banReason`, `banExpiresAt`, `banned` (bool form field mapped in `BeforeSave`); `role`, `email`, 2FA and password columns `ReadOnly`/`Sensitive`; policy denies `create` and `delete` (deletion goes through §7.5). `BeforeSave` runs inside the transaction (`https://vukem.dev/docs/digging-deeper/admin` "Resources"): sets `banned_at`, bumps `session_version`, calls `Tokens.RevokeAll`, inserts `account_events`, and registers `tx.AfterCommit` to dispatch `user.banned`/`user.unbanned` (FR-03-14); the panel audit log records the diff. Ban and unban are also reachable as app routes `POST /admin/users/{id}/ban` and `POST /admin/users/{id}/unban` (admin gate, CSRF) for the moderation pages (SDD 04, SDD 14 report outcomes), sharing the same service function.
- Admin gate: `gate.Define(admin.AbilityAdmin, u.Role == "admin" && u.BannedAt == nil && u.EmailVerifiedAt != nil)`; replaces the scaffold's `ADMIN_EMAILS` list (user-owned `app/admin/admin.go`).
- CLI (`vukem make:command`): `./app admin:grant <email>`, `./app admin:revoke <email>`, `./app user:reset-2fa <email> --reason "<text>"` (manual identity check by the owner; clears 2FA, revokes everything, mails the user). Each writes `account_events` with `actor_id` NULL and `data.cli=true`.

### 6.4 Export payload

Content of the export (personal records): `profile{id,email,name,username,locale,profilePrivate,personalizationOptOut,createdAt}`, `socialAccounts[{provider}]`, ratings, statuses, time-to-beat, lists with items, edit suggestions, `apiKeys[{name,createdAt,revokedAt}]`, subscriptions, sessions, tokens, account events; secrets, hashes and other users' data excluded. Read in one repeatable-read transaction. Packaging, delivery and limits belong to SDD 14 §7.4 (background ZIP with JSON/CSV files, link signed for 24 h, one export per 7 days); `account_events` gets an `exported` row when the job finishes.

### 6.5 Mail (vukem `https://vukem.dev/docs/digging-deeper/mail` "Mail": `mail.View` with `html/template`, subject via `EnvelopeContext` + `i18n.T`, `SendLater` on `QUEUE_CONNECTION=database`)

Templates in `resources/views/mail/`: `verify_email`, `reset_password`, `account_exists` (scaffold), plus `password_changed`, `two_factor_changed` (enabled, disabled, recovery code used, reset by admin; carries "not you? reset your password"), `passkey_changed` (added, removed; same "not you?" line; FR-03-24), `deletion_scheduled` (date + cancel link to `/login`), `account_anonymized`, `passkeys_retired` (cutover notice). Catalogs: scaffold `en`, `pt-BR` plus app `resources/lang/es.json` (an app catalog key wins and can add a locale, `https://vukem.dev/docs/security/authentication` scaffold section). Locale = `users.locale`; jobs without a request build the context locale from that column [NOT VERIFIED: API to set the i18n locale on a non-request context].

## 7. Behaviour and algorithms

### 7.1 Login state machine

`identifier+password` → (challenge needed?) → hash check (dummy hash when unknown) → banned? → 2FA? → `Login` (session regenerated, version recorded). Ban and deletion checks run only after a correct password so they reveal nothing to guessers. Banned with `ban_expires_at < now()` counts as not banned; a daily job clears expired bans and writes `unbanned`.

### 7.2 Password hashing and legacy compatibility

Better Auth 1.7 `hashPassword` (used in `legacy/apps/web/lib/db/seed-admin.ts`) produces `hex(salt16):hex(scrypt(NFKC(password), salt, N=16384, r=16, p=1, dkLen=64))` [NOT VERIFIED: legacy `node_modules` is not installed; confirm against `better-auth/crypto` source of 1.7.7 before Phase 4]. HexaScore wraps `hash.Default()` in `CompatHasher`:

- `Verify`: `$argon2id$` or `$2a$/$2b$/$2y$` → delegate (vukem `Verify` reads both); `^[0-9a-f]{32}:[0-9a-f]{128}$` → scrypt via `golang.org/x/crypto/scrypt`, compare with `subtle.ConstantTimeCompare`; anything else → false.
- `NeedsRehash`: true for scrypt and bcrypt; the ORM provider then stores argon2id (vukem "Session guard": rehash when `NeedsRehash` and `PasswordUpdater`).
- If the verified source differs, only the regex and constants change.
- Sunset: on 2027-12-31 remaining scrypt hashes are replaced by random argon2id and the users mailed a reset link.

### 7.3 Bot protection (replaces Vercel BotID; [NOT IN VUKEM DOCS])

vukem offers rate limits only. ADR candidate "0015 bot protection": a stdlib proof-of-work challenge, triggered as ADR 0008 states ("only when thresholds trip"):

- Trigger: per IP, > 3 sign-ups/h, > 5 failed logins/10 min, > 3 resets/h, or a global sign-up rate > 30/h.
- `GET /challenge` → `{c, bits, exp}`; `c` = HMAC (`app.DeriveKey("pow", 32)`) over random 16 bytes + expiry (120 s) + route; client finds `n` with `SHA-256(c‖n)` having `bits` leading zero bits (18 default, ~0.3–1.5 s on a mid-range Android [NOT VERIFIED], 21 under attack).
- Server: verify HMAC, expiry, route, zero bits, single use via cache `Add` key `pow:<c>` (same pattern as `social` states); store error → 503.
- Honeypot field on `/signup` and `/contact`: filled → same success answer, nothing stored.
- Alternative recorded in the ADR: Cloudflare Turnstile (free, external script, CSP and privacy impact).

### 7.4 Mobile tokens

- Lifetime 60 days, no refresh token: the app calls `rotate` when < 15 days remain (create + revoke in one transaction); expiry or revocation → re-login. Max 10 live tokens per user; the 11th revokes the least recently used. `deviceName` 1–64 chars, control runes stripped, stored as token `name`.
- Google: the server verifies the ID token with `jwt.NewVerifier` over `jwt.RemoteJWKS` (`https://www.googleapis.com/oauth2/v3/certs`), issuer `https://accounts.google.com`, audience = Android + web client ids (`https://vukem.dev/docs/security/social-login` "JWT guard"), then `FindOrCreateFromSocial`. Not `s.UserFromToken`: another app's access token would pass a userinfo call.
- 2FA: `code`/`recoveryCode` in the same request (as vukem's `POST /api/v1/tokens`).
- Implementation (P1.C.4.1): one verifier per accepted client id (`GOOGLE_ANDROID_CLIENT_ID`, `SOCIAL_GOOGLE_CLIENT_ID`), no id configured means every ID token is refused; `GOOGLE_JWKS_URL` overrides the key URL for tests. Issue routes are limited to 10 per minute per IP. The solved proof of work travels in the `X-Challenge: <c>:<n>` header (428 `challenge-required` carries the challenge); the body `challenge` field is accepted and ignored (P1.C.5.2). Rotation creates the new token before revoking the old one (vukem `Tokens` has no transaction hook).

### 7.5 Deletion

Schedule: set `deletion_scheduled_at`, bump `session_version`, delete sessions and remember tokens, `Tokens.RevokeAll`, mail. Cancel: login in grace → banner → `DELETE /profile/deletion`. Purge (scheduler, daily 03:00 UTC, idempotent, batches of 100) ports `legacy/apps/web/lib/account-purge.ts`: cancel live Stripe subscriptions first (failure keeps the user scheduled); then in one transaction delete sessions, tokens and `social_accounts`, revoke API keys, set email `deleted-<id>@invalid`, name "Deleted user", username/image/2FA NULL, random password, `profile_private` true, `anonymized_at`. Ratings, statuses, time-to-beat and edit history keep `user_id`; lists, editions, notifications and exports follow SDD 14 §8 (public lists anonymised, private lists deleted, pending open issue 14-7).

### 7.6 2FA migration

Better Auth encrypts the TOTP secret and backup codes with `BETTER_AUTH_SECRET` [NOT VERIFIED: format]. A one-off Bun script in `legacy/` decrypts with Better Auth's own helpers and pipes NDJSON `{userId, secret, backupCodes[]}` into the Go importer's stdin (never on disk). The importer seals the secret (`orm.Encrypted`), SHA-256-hashes unused codes, sets `two_factor_confirmed_at` for verified rows; lockout counters are dropped.

Implementation (P2.F.2.3, verified against the `better-auth` 1.7.7 source): the plugin keys TOTP with the UTF-8 bytes of a 32-character secret and shows it base32 encoded in the `otpauth` URI, so the importer stores `base32(secret)` (vukem `totp` format); `symmetricDecrypt` is called with the plain `BETTER_AUTH_SECRET` string (no `secrets` rotation array in the legacy config); backup codes default to `storeBackupCodes: "encrypted"`, a JSON array of the unused codes, hashed the way vukem `totp.UseRecoveryCode` checks them (workaround for VF-060); a row with `verified = false` is imported pending (no `two_factor_confirmed_at`). The export is sorted by user id (`COLLATE "C"`), so a resumed import skips lines up to its checkpoint. `scripts/legacy/twofactor-export.ts` holds the run instructions.

### 7.7 Cutover

In the read-only window (D-21) sessions are not copied, so everyone signs in again; Google links survive via `(provider, provider_id)`; the Google client gets the new callback URI before DNS switches.

### 7.8 Passkeys

Built on `auth/webauthn` in vukem v0.34.0 (VF-001); no app-local WebAuthn verifier. Relying party: ID is the `APP_URL` hostname (`hexascore.app` in production, so imported credentials keep working; staging uses its own host and cannot use production credentials), name `APP_NAME`, origin `APP_URL` (`http` only when the hostname is `localhost`). Registration: options from the server (excluding the user's existing credentials, resident key required, user verification required, attestation `none`), browser ceremony, server verification, row stored. vukem v0.34.0 `Credential` has no AAGUID; the app does not parse attestation; `aaguid` stays nullable. The row also stores `user_handle` (opaque `NewUserHandle` bytes; NULL on import because Better Auth 1.7.7 does not persist the random handle; pinned from the assertion only after a successful signature) and `algorithm` (COSE alg). The challenge is session-bound, deleted before verify, and expires after 5 minutes. Sign-in: options without `allowCredentials` (discoverable), assertion verified against the stored public key and counter, then the §7.1 checks after the credential matched (ban, anonymised account refused, deletion grace still allowed), then `Login` with session regeneration. A passkey login satisfies 2FA. Table `passkeys`: id, `user_id` FK cascade, `credential_id` unique, `public_key`, `algorithm`, `sign_count`, `transports` (JSON), `aaguid` nullable, `user_handle` nullable, backup flags, `name`, `created_at`, `last_used_at`. Deletion purge deletes the rows inside account anonymisation, because the user row is kept and ON DELETE CASCADE does not run (§7.5). `Permissions-Policy` allows `publickey-credentials-get=(self)` and `publickey-credentials-create=(self)` (SDD 13). Legacy reference: `legacy/apps/web/lib/auth.ts` (`passkey({ rpName })`), `app/[locale]/profile/passkeys-section.tsx` (list, add, remove with label "Remove passkey {name}", "Added {date}"), `app/[locale]/login/login-form.tsx` (`signIn.passkey()`, cancelled prompt → "Passkey sign-in was cancelled or failed."), copy in `messages/*/account.json` `profile.passkeys` and `auth.json`; the legacy e2e suite has no passkey spec.

## 8. Edge cases

- Unverified login: allowed; writes and tokens 403 `email-not-verified`; web goes to `/verify-email`.
- Banned user with active tokens: revoked in the ban transaction; `NotBanned` middleware also rejects survivors (403 `account-banned`), including bans set by SQL (no `user.banned` event then; the nightly SDD 02 recompute repairs scores).
- Deletion cancelled during grace: data intact; tokens stay revoked, so each device signs in again. On mobile, `GET /app/v1/me` returns `deletionScheduledAt` and the app shows the cancel screen.
- Google account with an existing email: provider-verified email links; an unverified local account gets a random password and loses sessions/tokens (vukem `--social`); provider-unverified email → 409.
- Google-only user enabling 2FA: must set a password first (vukem needs it to disable; legacy had `allowPasswordless: true`; D-25); the 2FA section shows the explanation and a "set password" link.
- 2FA lockout (device and codes lost): no self-service bypass; owner verifies identity out of band, runs `user:reset-2fa`; user gets `two_factor_changed`.
- Nickname race at onboarding: unique violation → 422 `nickname-taken`.
- Passkey-only legacy user: keeps signing in with the imported passkey; only if the import proves incompatible (ADR 0017) is the user mailed `passkeys_retired` with a reset link before cutover.
- Last passkey removed by a passkey-only user: allowed; the profile warns first and the account recovers through e-mail reset (FR-03-24).

## 9. Acceptance criteria

- `go test ./tests/... -run 'TestSecurity_A0[147]_'` → pass, including vukem-generated `TestSecurity_A07_SignupIsEnumerationSafe`, `TestSecurity_A07_ProvingTheEmailEndsSquatterAccess`.
- `go test ./app/accounts -run TestSecurity_A07_LegacyScryptVerifiesAndRehashes` → Better Auth 1.7.7 fixture hash verifies, stored value then starts `$argon2id$`, wrong password fails.
- `go test ./tests -run 'TestSecurity_A07_Ban'` → requests after a ban answer 403 `account-banned`; one `user.banned` event dispatched.
- `go test ./tests -run TestSecurity_A01_AdminRoleNotEditableInPanel` → crafted `role=admin` POST changes nothing.
- `go test ./tests -run 'TestSecurity_A07_Pow'` → 428 without a solution past the threshold; reused solution refused.
- `go test ./tests -run TestSecurity_A07_GoogleIDTokenAudience` → token for another client id → 401.
- `go test -tags=integration ./database/legacyimport/...` (Postgres 18 + Better Auth fixture dump) → counts match, §5 asserts return 0 rows, migrated 2FA users pass the challenge.
- `go test ./jobs -run TestDeletionPurge` → ratings unchanged, no PII in `users`, Stripe failure keeps the user scheduled.
- `bunx playwright test qa/e2e/specs/auth.spec.ts qa/e2e/specs/profile.spec.ts` → signup, verify, TOTP login, reset, delete and cancel; axe clean in three locales.

## 10. Test strategy

- Unit: validation, `CompatHasher`, PoW verifier, rotation, export. Integration (Postgres): importer, purge, ban, indexes.
- Contract: Schemathesis on `app-auth`. E2E: Playwright + axe, §6.1 pages, three locales. Load: k6 20 rps, 5 min.
- Security: `TestSecurity_*` above, ZAP on auth pages (0 High).

## 11. Open issues

1. Better Auth 1.7.7 scrypt parameters and NFKC [NOT VERIFIED]. Owner: architect. Closes when a `seed-admin.ts` hash verifies in the Go test.
2. Better Auth 2FA encryption format [NOT VERIFIED]. Decided (03-2): a Bun script in `legacy/` decrypts with Better Auth's own code and pipes plaintext secrets into the Go importer over stdin, never to disk (§7.6). Closes when the import test passes on a staging copy.
3. Google-only accounts need a password to manage 2FA. Decided: D-25 (set a password first, UI explains; FR-03-8).
4. Bot protection [NOT IN VUKEM DOCS]; ADR 0016 (PoW vs Turnstile) pending. Owner: architect. Feedback entry: "built-in PoW challenge middleware for auth routes".
5. Passkeys: the API landed in vukem v0.34.0 (`auth/webauthn`, tag `v0.34.0`, not claimed merged beyond that tag). The app adopts it (P1.C.7, P2.F.2.2b). Original note: `vukem-feedback.md` VF-001. Owner: vukem for any further API change.
6. Impersonation [NOT IN VUKEM DOCS], not in v1. Minimal design: `POST /admin/_users/{id}/impersonate` storing `impersonator_id` in the session, 30-min cap, writes blocked, banner, `account_events`. Feedback entry: "`SessionGuard.Impersonate`/`StopImpersonating` with audit event" (VF-012). Owner: vukem. Not in v1 (open issue 03-6).
7. Session timeout env names and defaults [NOT VERIFIED]. Owner: architect. Closes when read in Phase 4.
8. Locale for queued mail outside a request [NOT VERIFIED]. Owner: vukem. Closes with a documented API.
9. Legacy password max 32 is weak. Decided: maximum 128 (D-24), reverted to 32 by the owner on 2026-11-01 (FR-03-2).
10. `install:auth` column and Inertia page names [NOT VERIFIED]. Owner: architect. Closes after running the scaffold.
