# Security audit, 2026-11-07

Branch `feat/todo-2026-11`, base commit `65f56203`. Method: the vendored `security-audit` skill (Cloudflare), full-audit artifacts, run by one auditor that worked through the hunting classes area by area. Records: [findings](security-audit-2026-11-07.findings.json) and [coverage ledger](security-audit-2026-11-07.coverage.json), both accepted by the skill's `validate-findings.cjs` and `validate-coverage-ledger.cjs` (run in the pinned Node image of the Dockerfile, because the validators need POSIX no-follow file opens that Windows Node lacks).

Honest limits: no parallel hunter or verifier agents ran, so Phase 5 "independent record verification" was a second pass by the same auditor, not a fresh agent. Every ledger check is `method: source` (the fixes carry unit tests, but no sandboxed artifact was promoted). Six units are `deferred`; the run is not a claim that the repository is exhaustively covered.

## Summary

| Severity | Confirmed | Fixed |
| --- | --- | --- |
| Critical | 0 | 0 |
| High | 0 | 0 |
| Medium | 0 | 0 |
| Low | 1 | 1 |
| Informational | 0 | 0 |

Plus one `needs_validation` record (owner decision) and two `rejected` candidates, one of which got a hardening commit anyway.

## Findings

### SA-01 (low, CWE-307): web 2FA challenge throttled per user and IP only

- Id: `hexascore:2fa-challenge-throttle-per-ip-only`. Fixed in `f78b356e`.
- Where: `routes/two_factor.go:62`, `app/controllers/account/two_factor_login.go:31-41` (`TwoFactorThrottleKey`), `app/controllers/account/two_factor.go:205` (`Challenge`).
- Impact: an actor who knows an account's password can guess the six-digit code from many addresses, because each address (IPv6 network) has its own five-per-minute bucket and a failed guess neither locks the account nor ends the pending login. The app API path already had a per-account bucket; the web path did not.
- Reproduction: sign in with the password to open a pending challenge, send five wrong codes and get 429, repeat from another address and get five more attempts. The regression test `TestSecurity_A07_TwoFactorAccountKeyIgnoresIP` asserts the new account bucket key is identical across addresses and empty without a pending login.
- Fix: new rule `two-factor.account` (10 per 15 minutes, keyed on the pending user id alone) added after the per-IP limiter on `POST /2fa`.

## Needs validation (owner decision)

### NV-01: Google ID-token exchange is not bound to the request, and the mobile flow is implicit over a custom scheme

- Id: `hexascore:google-id-token-not-bound-to-request`.
- Where: `app/controllers/appapi/tokens_controller.go:121,165`, `routes/app_api.go:19-33`, `apps/mobile/src/auth/googlePrompt.ts`.
- What the verifier checks. `auth/jwt.Verifier` as configured in `googleVerifiers`: RS256 only, signature from Google's JWKS, `iss` equals `https://accounts.google.com`, `aud` equals the Android client id or the web client id (one verifier per id), `exp` with 30 s leeway. It has no `nonce` option and no replay record, so any unexpired token minted for the client id is accepted. The web social path (`social.OIDCProvider.verify`) does check `iss`, `aud` (plus `azp` for several audiences), `exp`, `iat` and `nonce` (when non-empty; the web flow always passes the cookie nonce). The framework gap is recorded as HS-3 in `vukem-feedback.md`.
- Exposure: the app uses the implicit ID-token flow and returns to `hexascore://auth/google`. Another app that claims that scheme could receive the token. Whether Google issues tokens to this client type that way is not visible in source, so this stays unconfirmed.
- Decision needed: confirm in Google Cloud whether the sign-in works at all; if yes, move to the authorization code flow with PKCE (or the native Credential Manager) and a server-issued nonce. Adding a client nonce alone would not stop an interceptor, so it was not implemented. The CSPRNG nonce hardening below is independent of this.
- Status: fixed (owner decision 2026-11-07, see docs/brainstorming.md). The mobile app signs in with the authorization code flow, PKCE and `state` (`5e1d1643`); the server issues a single-use nonce (`POST /app/v1/auth/google/nonce`, hashed in `google_nonces`) and `/app/v1/tokens/google` mints a token only when the ID token carries an unused nonce it issued, consuming it in the same transaction (`a54ca583`, test `TestSecurity_A07_GoogleNonce`). Passkeys on mobile are out of scope.

## Hardening commits (not findings)

- `4772fd45` fix(mobile): the Google sign-in nonce is drawn from `crypto.getRandomValues`; with no CSPRNG the sign-in is skipped (no `Math.random` fallback). `expo-crypto` is only a transitive dependency, so no dependency was added. Owner decision: if Hermes on a target device exposes no `crypto.getRandomValues`, Google sign-in returns cancelled; approve `expo-crypto` as a direct dependency if that happens.
- `66386651` fix(uploads): `measureCover` refuses images over 25 MP (`TestSecurity_A04_CoverUploadRejectsDecompressionBombs`) and `covers.Fingerprint` returns `ErrTooLarge` before decoding (`TestSecurity_A04_FingerprintRejectsBombsBeforeDecoding`).

## Rejected candidates

- `hexascore:cover-pixel-cap-missing`: no lower-trust principal reaches a full decode of an oversized image; hardened above anyway.
- `hexascore:mobile-open-source-url`: `Linking.openURL(sourceUrl)` takes an ingest-written URL, not user input.

## What was reviewed and held

SQL construction and allow-listed sort/filter/search; the single `template.HTML`; SSRF guards on webhooks and cover fetch; redirects and locale handling; the render cache (anonymous-only, nonce swap); billing (framework-verified webhook that fails closed, subscription state re-fetched from Stripe, checkout and portal bound to the session user); API keys (crypto/rand, SHA-256 at rest, owner-scoped revoke, ban check) and app tokens (cap, conditional rotation); TOTP replay (locked step update), recovery-code consumption, password-gated disable and regenerate, session-bound 5 minute pending challenge; passkey challenge and RP binding; Google/social account linking (an unverified email never takes over an account, an unverified local account is reset first); admin routes (a scan of every `(c *web.Ctx) error` handler in `app/admin` found an authorization call in each; resource lists never select secret columns); owner filters on every user-owned UPDATE/DELETE; suggestion policy; export links (signed, 24 h, SafeCell CSV); og_image and SSR supervisor subprocesses (fixed names, own temp paths, timeouts); frontend XSS sinks; GitHub workflows and Dockerfile; mobile token storage (SecureStore, device-only), verified App Links and the deep-link allow-list; dependencies (`govulncheck`: 0 reachable; `bun audit`: four dev-tooling advisories already tracked in `osv-scanner.toml`).

## Deferred (coverage ledger)

Stripe cancellation and plan-change races, backup dump/restore argument handling, account-deletion erasure completeness, per-operation OpenAPI security versus handlers, ingest source parsers, and public API field exposure were not hunted or only sampled. A `deep` profile run should start there.

## Commits

| Commit | Change |
| --- | --- |
| `ba11d9b4` | first audit report (superseded by this one) |
| `4772fd45` | mobile nonce from the CSPRNG |
| `66386651` | image pixel caps |
| `f78b356e` | per-account 2FA challenge limiter (SA-01) |
