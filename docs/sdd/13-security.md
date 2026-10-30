# SDD 13: Security, privacy and abuse

Status: reviewed (2026-10-28 consistency pass). Decisions: D-03, D-05, D-07, D-08, D-15, D-29, A-01, A-07. ADRs: 0004, 0006, 0008, 0014, 0016. Related SDDs: 02, 03, 04, 05, 06, 09, 10, 11, 12, 14.

Source keys (vukem site, fetched 2026-10-28; heading quoted at each use): [OV] `https://vukem.dev/docs/security/overview`, [HDR] `/docs/security/security-headers`, [BB] `/docs/security/building-blocks`, [RL] `/docs/security/rate-limiting`, [ENC] `/docs/security/encryption`, [SURL] `/docs/security/signed-urls`, [TOK] `/docs/security/api-tokens`, [AUTHN] `/docs/security/authentication`, [AUTHZ] `/docs/security/authorization`, [CSRF] `/docs/basics/csrf`, [SES] `/docs/basics/sessions`, [VAL] `/docs/basics/validation`, [UP] `/docs/digging-deeper/uploads`, [STO] `/docs/digging-deeper/storage`, [CFG] `/docs/configuration`, [DOCK] `/docs/deployment/docker-and-railway`, [INR] `/docs/frontend/inertia`, [ADM] `/docs/digging-deeper/admin`, [BILL] `/docs/digging-deeper/billing`. External (accessed 2026-10-28): [T10] OWASP Top 10:2025 `https://owasp.org/Top10/2025/`; [ASVS] OWASP ASVS 5.0.0 `https://owasp.org/www-project-application-security-verification-standard/`; [LGPD] Lei 13.709/2018 `https://www.planalto.gov.br/ccivil_03/_ato2015-2018/2018/lei/l13709.htm` (connection reset; re-verify, issue 5). Legacy baseline: `legacy/docs/security.md`.

## 1. Objective

Keep accounts, scores and the catalogue trustworthy on one Railway container (D-08) with no paid security vendor. vukem's secure defaults ([OV] "Defaults you get today") are the floor. This SDD adds HexaScore's threat model, headers, abuse limits, LGPD duties, admin hardening, payment and supply-chain rules, test cadence and incident response. Target: ASVS 5.0 level 2 for authentication, session, access control and validation [ASVS].

## 2. Scope

In (v1): web (including passkeys, ADR 0017), `/api/v1`, `/app/v1`, admin, webhooks, ingestion. Out: impersonation (VF-012), WAF/CDN (D-08), antivirus service (issue 9), mobile certificate pinning (SDD 10 §7.7 still holds: tokens carry no billing power), bug bounty.

## 3. Functional requirements

- FR-13-1: Every route belongs to one §6.1 stack, listed in `tests/testdata/route_inventory.txt`. Verify: `TestSecurity_A01_HexaRouteStackInventory` walks the router and fails on an unclassified route, on a route whose prefix or middleware contradicts its stack, and (with `TEST_DATABASE_URL`) on a stale line.
- FR-13-2: Every unsafe unauthenticated route has a §7.1 limit and answers 429 with `Retry-After`. Verify: table-driven `TestSecurity_A06_HexaLimits`.
- FR-13-3: The proof-of-work challenge (§7.2) triggers on the ADR 0016 thresholds and refuses a reused solution. Verify: `TestSecurity_A07_Pow*` (SDD 03).
- FR-13-4: Every body, query, path, webhook field and stored jsonb value passes a `z` schema before use. Verify: lint flags `c.FormValue` and `map[string]any` decoding in `app/`; fuzz test per boundary.
- FR-13-5: Admin routes need the `admin` role, confirmed 2FA and a session within §7.5 limits. Verify: `TestSecurity_A01_HexaAdminNeeds2FA`.
- FR-13-6: Data-subject rights work end to end (§7.6). Verify: SDD 03/14 integration tests.
- FR-13-7: Every incident follows §7.9 and gets a review within 5 working days. Verify: manual checklist.

## 4. Non-functional requirements

- NFR-13-1: Security middleware adds ≤ 2 ms p95. Verify: k6 delta between `/up` and a static page.
- NFR-13-2: Limiters fail closed (503) when the store errors ([RL] "fail closed"). Verify: vukem `TestSecurity_A06_RateLimitMiddlewareFailsClosed` plus Hexa wiring test.
- NFR-13-3: 0 High in ZAP (baseline + active) on staging before every release and the cutover (D-21).
- NFR-13-4: Logs with IPs ≤ 30 days. No raw emails, tokens or keys. User ids are logged as `HMAC-SHA256(app.DeriveKey("log", 32), id)[:16]` (SDD 12 FR-12-8).
- NFR-13-5: A called govulncheck finding or a `bun audit` high blocks merge. Fix SLA: critical 72 h, high 7 days, medium 30 days.
- NFR-13-6: Breach notification target: 72 h (§7.9).
- NFR-13-7: Security controls cost US$0/month (D-08).

## 5. Data model

No security-only tables. Owned elsewhere: `account_events` (SDD 03, 2 years), vukem `admin.AuditLog` (forever, [ADM] "The audit log"), `catalog_audit` (SDD 04), limiter and `pow:<c>` cache rows (pruned hourly by `cache:prune`, [RL]; the cache also holds `contrib/stripe`'s webhook event-id dedup, SDD 05 §5).

LGPD inventory (bases per [LGPD] Art. 7; D-29 for personalisation):

| Data | Where | Retention | Basis |
| --- | --- | --- | --- |
| Name, email, nickname, locale, argon2id hash, Google link | `users`, `account` | account life; 30-day grace, then anonymised | contract (Art. 7 V) |
| TOTP secret, recovery hashes (`orm.Encrypted`) | `two_factor` | until disabled/anonymised | contract |
| Session IP (/24, /48), UA ≤ 256 B | `session` | ≤ 30 days | legitimate interest (Art. 7 IX) |
| Security history | `account_events` | 2 years | legitimate interest |
| Ratings, statuses, lists, time-to-beat | SDD 02/14 | kept, unlinked after deletion | contract; legitimate interest |
| Suggestions, catalogue audit | SDD 04 | forever, actor nulled | legitimate interest + A-07 |
| Item views for recommendations | SDD 08 | per SDD 08 | legitimate interest + opt-out (D-29) |
| Contact messages | SDD 14 | 24 months | legitimate interest |
| Key hashes, usage, subscriptions | SDD 05 | billing 5 years | contract; legal obligation (Art. 7 II) |
| Request logs (IP) | Railway | ≤ 30 days | legitimate interest |
| Search logs (no IP) | SDD 07 | 30 days | legitimate interest |
| Web vitals, mobile crash reports | SDD 12, 10 | 7 days | consent (opt-in) |
| Export ZIPs | private bucket | 24 h | contract |
| Backups (encrypted, ADR 0014) | bucket | 30 daily + 12 monthly | legitimate interest; deletions replayed on restore |

Marketing: none.

## 6. Contracts

### 6.1 Stacks

| Stack | Routes | Auth | CSRF | Limit key |
| --- | --- | --- | --- | --- |
| web | `hexascore.app/*` | session | CrossOriginProtection + token fallback | IP, account |
| admin | `/admin/*` | session + `admin` + 2FA + short session | panel CSRF ([ADM] "Security") | account |
| api | `api.hexascore.app/v1/*` | API key header | none (no cookies) | IP, key/plan (SDD 05) |
| app | `/app/v1/*` | `hx_app_` bearer + ability | none (bearer only) | IP, token |
| webhook | `POST /stripe/webhook`, `POST /csp-reports` | signature / none | `CSRFBypass` | IP |

### 6.2 Headers

`SecureHeaders` comes from the default stacks and is configured, never added twice ([HDR] "Common mistakes").

| Header | Web host | API host (JSON) |
| --- | --- | --- |
| `Content-Security-Policy` | SDD 09 §7.8 policy (`'self' 'nonce-{n}' 'strict-dynamic'`, nonce styles, `object-src 'none'`, `frame-ancestors 'none'`) + `ReportURI` `/csp-reports` | `default-src 'none'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'` (custom `CSP`, legacy `API_CSP`) |
| `Strict-Transport-Security` | `max-age=63072000; includeSubDomains` | same |
| `X-Content-Type-Options` | `nosniff` | same |
| `X-Frame-Options` | `DENY` | same |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | `no-referrer` |
| `Cross-Origin-Opener-Policy` | `same-origin` | same |
| `Cross-Origin-Resource-Policy` | `same-site` | `cross-origin` |
| `Permissions-Policy` | `accelerometer=(), autoplay=(), camera=(), display-capture=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), midi=(), payment=(), usb=(), browsing-topics=(), publickey-credentials-get=()` | same; becomes `publickey-credentials-get=(self), publickey-credentials-create=(self)` when passkeys ship (ADR 0017) |
| CORS | none | `*`, no credentials (SDD 05) |

Notes:

- Defaults and field names: [HDR] "What every response carries", "Configuration".
- `'strict-dynamic'` is an allowed `CSPSources` keyword ([HDR] "What CSPSources refuses"). This closes SDD 09 issue 5, part 1.
- vukem never adds `preload` ([HDR] "Lifetime"). HexaScore does not preload in v1, because removal from the list takes months and it pins every future subdomain. Legacy had `preload` (issue 10).
- HSTS behind Railway uses `TRUSTED_PROXIES` + `TRUST_FORWARDED_PROTO`, or `HSTS_FORCE=true` ([DOCK] "HSTS without a trusted range"; SDD 11 picks).
- Scalar (`/docs/reference`) keeps `apidocs`' nonce CSP and `DENY` framing.
- Nonce pages are never publicly cached ([HDR] "Common mistakes").
- `/app/v1` sends no CORS headers; vukem's default is deny ([OV] "CORS denies by default").

### 6.3 Problem codes

`rate-limited` 429, `challenge-required` 428 (body carries the `/challenge` URL), `csrf-rejected` 403, `payload-too-large` 413, `service-unavailable` 503.

## 7. Behaviour and algorithms

### 7.0 Threat model

Assets: credentials and 2FA secrets; cookies, app tokens, API keys; rating integrity; catalogue integrity; subscriptions (cards stay at Stripe); personal data.

Actors: anonymous abusers, score manipulators, scrapers, malicious contributors, a compromised admin, compromised dependencies or sources.

| Boundary | STRIDE | Threat | Control | Owner |
| --- | --- | --- | --- | --- |
| Browser → web | S | credential stuffing | login limits, PoW, argon2id, 2FA | 03, 13 |
| Browser → web | T | CSRF, XSS | Fetch Metadata, nonce CSP, escaping, `z` | 09, 13 |
| Browser → web | I | session theft, back-button leak | `__Host-` HttpOnly cookie, `EncryptHistory` | 03, 13 |
| Browser → web | D | form spam | limits, honeypot, PoW | 13 |
| Mobile → `/app/v1` | S, E | stolen or over-scoped token | hashed 60-day tokens, abilities, revoke | 03, 10 |
| Client → `/api/v1` | I, D | bulk scraping, licence leak, quota exhaustion | quotas, provenance allowlist (ADR 0006), fail closed | 05 |
| Users → scores | T | sockpuppets | verified email, age weights, bursts, daily caps | 02 |
| Contributor → catalogue | T | vandalism, SSRF, image bombs | moderation, quota, `httpclient.Safe`, `upload` | 04 |
| Sources → ingestion | T | poisoned payloads | `z` before upsert, host allowlist, size caps | 06 |
| Stripe → webhook | S, R | forged or replayed events | signature, 5-min window, `contrib/stripe` event-id dedup in the persistent cache, handlers idempotent by Stripe object id | 05 |
| Admin → panel | E, R | compromised admin | 2FA, short session, audit, CLI-only role grant | 03, 13 |
| App ↔ Railway | I | leaked variables or logs | `config.Secret`, redaction, minimal team | 11, 12 |
| App ↔ bucket | I, T | public exports, traversal | private prefix, signed links, `os.Root` | 11, 14 |
| Build → image | T | dependency implant | `go.sum`, lockfile, digests, scanners | 11, 13 |

### 7.0.1 OWASP Top 10:2025 mapping [T10]

| Cat. | vukem default | HexaScore control | Test | Owner |
| --- | --- | --- | --- | --- |
| A01 | policies fail closed [AUTHZ]; `httpclient.Safe` [BB] "What is blocked"; `os.Root` [OV] A01 | owner filters, abilities, route inventory | `TestSecurity_A01_HexaRouteStackInventory` | 03, 04 |
| A02 | `SecureHeaders`, refusing `CSPSources` [HDR]; prod needs `APP_KEY` [OV] A02 | §6.2 values | `TestSecurity_A02_HexaHeadersWeb`, `_HexaHeadersAPI`, Playwright `headers.spec.ts` | 09, 13 |
| A03 | `go.sum`, govulncheck, Dependabot [OV] A03 | §7.7 | CI `vuln`, `deps` | 11 |
| A04 | HKDF sub-keys [ENC]; SHA-256 tokens [TOK] "Security" | rotation §7.4 | `TestSecurity_A04_HexaKeyRotationReadsOldColumns` | 03 |
| A05 | `z`, bound parameters [VAL] "Security" | search caps (SDD 07), diff escaping | `TestSecurity_A05_HexaSearchControlChars` | 04, 07 |
| A06 | body cap 10 MiB, `BodyLimit` [OV] A06; limiter fails closed [RL] | §7.1, PoW, server-side quotas | `TestSecurity_A06_HexaLimits` | 02, 05 |
| A07 | enumeration-safe flows [AUTHN]; `Regenerate` [SES] | admin 2FA, PoW | `TestSecurity_A07_*` (SDD 03) | 03 |
| A08 | signed billing webhook outside CSRF [BILL] "Security"; `webhook.Dedupe` [BB] | event-id dedup in the `database` cache store (shared, survives restarts) + handlers idempotent by Stripe object id (SDD 05 §5) | `TestSecurity_A08_StripeReplayIgnored` | 05 |
| A09 | `web.SecurityEvent`, `web.AuditEvent` [BB] "Audit events"; `config.Secret` [OV] | hashed ids, alerts (SDD 12) | `TestSecurity_A09_HexaLogsRedacted` | 12 |
| A10 | central error handler [OV] A10 | store down → 503 | `TestSecurity_A10_HexaPowStoreDown503` | 13 |

### 7.1 Rate limits

Store: `RATE_LIMIT_STORE=cache` on `CACHE_STORE=database` (SDD 03 NFR-03-3), which counts in fixed windows with an atomic `Increment` ([RL] "RATE_LIMIT_STORE"). Budgets survive deploys and crashes. The `memory` store (per-process GCRA) is for tests only: in production every restart would hand attackers a fresh budget.

The table lives in `app/http/limits` (one named rule per row; routes never build their own limiter, `TestSecurity_A06_HexaLimits` checks both). In production the app refuses to boot unless `RATE_LIMIT_STORE=cache` (the image defaults it with `CACHE_STORE=database`). Not wired yet: suggestions (P4.A routes), `client-errors`/`client-metrics` and `/csp-reports` (no routes yet), the `/app/v1` read/write split (one 60/min user budget today) and `/stripe/webhook` (mounted with `UseServer` by vukem billing, signature checked first).

Keys: IPv6 is bucketed by /64 [RL]. `c.IP()` trusts forwarding headers only from `TRUSTED_PROXIES` ([OV] A01). A 429 carries `Retry-After` and `RateLimit-Limit/Remaining/Reset` [RL].

| Route group | Limit | Key |
| --- | --- | --- |
| Login (web, token) | 5/min; 20/10 min | identifier + IP; account |
| Sign-up | 10/10 min; 5/h | IP; email hash |
| Forgot password | 10/10 min; 5/h | IP; email hash |
| 2FA | 5/min | user + IP |
| Contact | 5/h | user or IP |
| Suggestions | drafts 30/h, submit 10/min, covers 10/h; quota 3/day (SDD 04) | user, IP |
| Ratings | 30/min; daily caps SDD 02 §7.6 | user |
| Statuses, lists | 60/min | user |
| Reports | 10/h | user |
| Search page / suggest | 60/min / 300/min | IP |
| `/api/v1` | 120/min, then plan rps + quota | IP; key |
| `/app/v1` | reads 120/min, writes 30/min; 600/min | token; IP |
| `client-errors` / `client-metrics` | 10/min / 60/min | IP |
| `/csp-reports` | 30/min, body ≤ 16 KiB ([HDR] "Collecting violation reports") | IP |
| `/stripe/webhook` | 600/min, signature checked first | IP |
| Export | 1 per 7 days (SDD 14) | user |

### 7.2 Challenge and form integrity (ADR 0016)

- Trigger (SDD 03 §7.3), per IP: > 3 sign-ups/h, > 5 failed logins/10 min, > 3 resets/h. Global: > 30 sign-ups/h. A tripped flag holds 1 h (`challenge:<route>:<ip|all>:trip`, `TripHold`).
- Code: `app/challenge` (issue, verify, solve), `app/http/middleware/challenge.go` (thresholds, trip flags, fallback), `resources/js/challenge/worker.ts` (solver), `app/forms` (signed form render time). Tests: `TestSecurity_A07_Pow*`, `TestSecurity_A10_HexaPowStoreDown503`, `TestSecurity_A07_HexaHoneypotAndFillTime`; load: `qa/k6/login-pow.js`.
- `GET /challenge?route=signup` → `{c, bits, exp}`:
  - `c` = HMAC-SHA256 with `app.DeriveKey("pow", 32)` over 16 random bytes, the expiry and the route; TTL 120 s.
  - The client finds `n` such that `SHA-256(c‖n)` has `bits` leading zero bits (18; 21 under a global trip). It runs in a Web Worker (`worker-src 'self'`).
- Server check order: HMAC, expiry, route, zero bits, then single use via cache `Add("pow:"+c)`. A replay gets 428; a store error gets 503.
- Accessibility: no puzzle and no image. An `aria-live` region says "Checking your browser…". After 10 s the client may ask once for `bits-2` (1/h per IP): `GET /challenge?route=…&fallback=1`; the gate accepts solutions down to `bits-2`, since only that endpoint hands them out. Without JS, a localised notice offers "Continue with Google", which is never challenged.
- Honeypot `website` field, plus a minimum fill time of 3 s taken from the render time in the signed form token (`form_token`, `app/forms`; a missing or forged token counts as a violation). A violation gets the normal success answer, stores nothing and logs `security.honeypot`. Suggestions get a 422 instead (SDD 04).
- Turnstile: app setting `BOT_CHALLENGE=pow|turnstile` (HexaScore config, not vukem). It adds `https://challenges.cloudflare.com` to `script-src`/`frame-src` on challenged pages only. Using it in production needs an ADR 0016 update (issue 12). The framework hook is requested in VF-003.

### 7.3 Input handling

- `z` everywhere ([VAL] "Security": strings refuse invalid UTF-8 and control characters).
  - Webhooks: verify the signature first, then validate only the fields read.
  - Ingestion: validate before upsert (SDD 06).
  - jsonb: decode into typed structs, because rows written by older schema versions must not crash or inject.
- Body limits via `BodyLimit` ([OV] A06), implemented by `limits.Body` (`app/http/limits/body.go`) at server level (vukem's `MethodOverride` parses urlencoded forms before routing, VF-061) and at router level (replacing vukem's 10 MiB default); the Stripe webhook passes its limit to `stripe.WebhookConfig.MaxBody`. "Forms" means every web and admin route; "Other routes" means the `api` and `app` stacks. Verify: `TestSecurity_A06_HexaBodyLimits`.

| Route | Limit |
| --- | --- |
| Forms | 64 KiB |
| Suggestion JSON | 256 KiB |
| Cover upload | 5 MiB |
| Stripe webhook | 512 KiB |
| `client-errors` | 16 KiB |
| `client-metrics` | 4 KiB |
| Other routes | 1 MiB |
| Ingestion (SDD 06) | JSON 20 MB, images 15 MB |

- Covers ([UP] "How a file is checked"):
  - Type sniffed from the first 512 bytes; JPEG, PNG and WebP only, with matching extensions.
  - Random stored name. Over 5 MiB → 413.
  - `MaxPixels` 25 MP ([UP] "Image dimensions"); size from 300×450 to 4000×6000.
  - Re-encoded by the thumbnail step, which "removes all metadata (EXIF, GPS location)" ([UP] "Thumbnails"). Only the re-encoded file is kept.
- Antivirus: vukem offers ClamAV via `Rules.Scanner` ([UP] "Virus scanning"), but clamd is an extra service (D-08). Compensating controls: images only, sniffing, re-encoding, a private quarantine prefix until an admin approves, `nosniff` (issue 9).
- SSRF:
  - Suggestion source URLs: `https` only, stored and never fetched, rendered with `rel="nofollow ugc noopener"`.
  - Cover URLs, the YouTube oEmbed check (host `www.youtube.com`) and ingestion sources are fetched only through `httpclient.Safe` ([BB] "SSRF-safe HTTP client").
  - Tests: `TestSecurity_Community_A01_SourceURLSSRF`, `TestSecurity_A01_IngestSSRF`.
- Path traversal: storage uses `os.Root`, so symbolic links cannot escape [STO]. Object keys are generated, never user input.

### 7.4 Secrets and keys

- `APP_KEY`:
  - Created with `vukem key:generate` (`--show`, `--force`) [CFG].
  - Purpose keys are derived with HKDF-SHA256 via `app.DeriveKey` ([ENC] "One APP_KEY, many keys"). A missing key fails closed [ENC].
- Rotation (yearly, on staff change, after a leak), per [ENC] "Key rotation":
  1. Set the new `APP_KEY` and put the old one in `APP_PREVIOUS_KEYS`.
  2. Keep the overlap 31 days (remember-me lasts 30).
  3. Re-seal the `orm.Encrypted` columns that `NeedsRotation` reports. This job is app code, `[NOT IN VUKEM DOCS]` (issue 8).
  4. Drop the old key.
- After a leak: SDD 12 R9. The leaked key stays in `APP_PREVIOUS_KEYS` only while the re-seal runs under `vukem down`; then every session and remember token is invalidated.
- Other secrets (`config.Secret`, Railway variables, names in SDD 11): Postgres URL, Stripe restricted key ([BILL] "Security") and webhook secret, Google client secret, IGDB/TMDB credentials, Resend key, Telegram token, OTLP headers, bucket credentials, Turnstile secret.
  - Rotation: new value at the provider, set it in Railway, redeploy, check `/up?checks=1`, revoke the old value.
  - Never in git or images (`.dockerignore`, vukem `TestSecurity_A02_DockerIgnoresSecrets`).
- Export ZIPs are reached through 24 h signed URLs [SURL]. Tampering is rejected (vukem `TestSecurity_A01_SignedURLTampering`).

### 7.5 Sessions, CSRF, admin

- Session cookie: `__Host-`, `HttpOnly`, `Secure`, `SameSite=Lax`, database driver (SDD 03).
- `Regenerate()` on login, on 2FA confirmation and on every role change ([SES] "Login, logout and session fixation").
- Logout: `Invalidate()` + `inertia.ClearHistory(c)`, with `inertia.Config{EncryptHistory: true}` on globally ([INR] "History encryption").
- Sign out everywhere: bump `session_version` + `Tokens.RevokeAll` (SDD 03 FR-03-12).
- CSRF: `http.CrossOriginProtection` with `CSRFTokenFallback: true` [CSRF]. The fallback only applies to unsafe requests that send neither `Sec-Fetch-Site` nor `Origin` (issue 2). Bypass list: exactly `POST /stripe/webhook` and `POST /csp-reports`; the latter is also in `SessionBypass`.
- Admin:
  - Role granted by CLI only. TOTP mandatory: an admin without it is redirected to setup.
  - Session: idle 30 min, absolute 12 h. Optional `ADMIN_IP_ALLOWLIST` (CIDRs, checked with `c.IP()`). The 2FA gate, lifetimes and allowlist are HexaScore middleware, `[NOT IN VUKEM DOCS]` (issue 7): `AdminGuard` (`app/http/middleware/admin_guard.go`, P2.D.3.1), router-wide so it also covers the moderation routes mounted next to the panel; allowlist miss -> 404, expired admin session -> logout and `/login`, admin without confirmed TOTP on `/admin` -> `/user/two-factor`. vukem v0.33.0 `admin.Config{TwoFactor, MaxSessionAge, AllowCIDRs}` (VF-030) guards only the panel's own routes and has no idle timeout, so it is not used (VF-058). 2FA confirmation calls `Regenerate()`; a CLI role change ends the user's sessions (session version bump, remember-me tokens deleted).
  - Panel writes are audited in the same transaction ([ADM] "The audit log"). Other admin actions emit `web.AuditEvent` ([BB] "Audit events").
  - No impersonation (VF-012).

### 7.6 Data-subject rights ([LGPD] Art. 18; 15 days, Art. 19)

| Right | How |
| --- | --- |
| Access, portability | ZIP export (SDD 14 FR-14-12), signed 24 h link, ≤ 60 s; target always ≤ 15 days (issue 1) |
| Deletion | 30-day grace, then anonymisation (SDD 03 FR-03-15) |
| Rectification | profile; nickname via support |
| Objection | personalisation opt-out (D-29) |
| Consent withdrawal | cookie banner, mobile settings |
| Other | contact topic "Privacy", answered ≤ 15 days |

The privacy policy must list:

- the §5 inventory with bases and retention;
- cookies: session, 2FA, `locale`, `theme`, consent;
- opt-in analytics (SDD 12) and YouTube embeds after consent;
- processors: Railway, Grafana Cloud, Resend, Stripe, Google, Telegram; international transfers;
- IGDB/TMDB attribution and the A-07 licence;
- minimum age 13;
- the encarregado contact (issue 6).

Legacy text is `legacy/apps/web/messages/en/legal.json`; drop its Vercel, Neon, Upstash, passkey and AbacatePay mentions.

### 7.7 Payments and supply chain

- Payments: Stripe Checkout and Portal by redirect, so no card field exists on HexaScore pages (PCI DSS SAQ-A).
  - Access comes from webhooks only.
  - Signatures older than 5 minutes are refused ([BILL] "Security").
  - `contrib/stripe` deduplicates event ids in the persistent `database` cache store; handlers are idempotent by Stripe object id (SDD 05 §5).
- Go:
  - Pinned `go.sum`; checksum DB never disabled ([OV] A03).
  - `govulncheck` pinned as a `tool` directive (`vukem/docs/security.md`, `[SITE PAGE MISSING]`).
  - `osv-scanner` on `go.mod` and `bun.lock`, pinned (vukem uses it, `vukem/docs/maintenance.md`, `[SITE PAGE MISSING]`).
- JS: `bun install --frozen-lockfile`; exact versions, no `^`, no pre-releases (D-22); `bun audit --audit-level=high`; Biome (issue 3).
- Dependabot, chosen over Renovate because vukem and legacy already use it: weekly for `gomod`, `bun`, `docker` and `github-actions`. Actions are pinned by SHA; base images by `@sha256` (SDD 11). Optional syft SBOM on release.

### 7.8 Security testing cadence

| When | What |
| --- | --- |
| PR | `go test ./... -run TestSecurity_`, govulncheck, osv-scanner, `bun audit`, Playwright `headers`/`csp`, Schemathesis (both documents) |
| Weekly | ZAP baseline + active scan vs staging (`qa/zap/plan.yaml` port, `errorLevel: High`), govulncheck, Dependabot |
| Release | ZAP active 0 High; new routes checked against §6.1 (port of legacy `hexascore-secure-endpoint` skill) |
| Yearly | key rotation, admin and Railway member review, Stripe key scopes, retention jobs, policy current, restore drill with deletion replay, ASVS L2 walk |

New Hexa tests: §7.0.1 names plus `TestSecurity_A01_HexaAdminIPAllowlist`, `_A02_HexaCSPHasStrictDynamic`, `_A06_HexaBodyLimits`, `_A07_HexaAdminSessionExpiry`, `_A07_HexaHoneypotAndFillTime`, `_A08_HexaUploadReencodedNoExif`. vukem's suite stays on.

### 7.9 Incident response

| Severity | Examples | Response |
| --- | --- | --- |
| SEV1 | data breach, active exploitation, leaked `APP_KEY`/Stripe key | start ≤ 1 h, maintenance mode if needed, notify ≤ 72 h |
| SEV2 | exploitable flaw, no sign of use | fix ≤ 72 h |
| SEV3 | hardening gap, Medium ZAP finding | next release |

The owner (sole operator) decides severity and notification. Steps:

1. Contain: revoke, rotate (§7.4, SDD 12 R9), ban, `vukem down`.
2. Export Railway logs before they expire.
3. Assess the data and the people affected.
4. Notify ANPD and the affected users when there is relevant risk ([LGPD] Art. 48; the ANPD rule of 3 working days is `[NOT VERIFIED]`, issue 5).
5. Fix, then review.

User template (en/pt-BR/es): what happened, when, which data, what we did, what to do (reset password, revoke tokens), contact.

Review in `docs/incidents/YYYY-MM-DD-<slug>.md` within 5 working days: UTC timeline, impact, root cause, detection gap, actions with owners, SDD changes. The operator copy is `docs/runbooks/incident-response.md`; the empty review is `docs/incidents/TEMPLATE.md`; user text is `docs/runbooks/incident-notice.{en,pt-BR,es}.md`.

## 8. Edge cases

- Limiter store down → 503, never fail open.
- An IPv6 client rotating inside one /64 shares a bucket (vukem `TestSecurity_A06_RateLimitIPv6RotationSharesBucket`).
- A restore brings back anonymised accounts → the runbook re-applies newer `account_events` deletions (issue 11).
- A valid PoW solution while an account lockout is active → the lockout wins (same 422).
- A Stripe event arrives before the success redirect → access comes from the webhook; the success page grants nothing.
- An admin loses the TOTP device → recovery codes, or an owner CLI reset with an audit event.

## 9. Acceptance criteria

- [ ] `curl -sI https://staging.hexascore.app/` → CSP has `'nonce-` and `'strict-dynamic'`, no `unsafe-inline`; HSTS without `preload`; `Permissions-Policy` as §6.2.
- [ ] `curl -sI https://api.staging.hexascore.app/v1/items` → API CSP of §6.2, `access-control-allow-origin: *`, no credentials header.
- [ ] `curl -si -X POST https://staging.hexascore.app/contact -H 'Sec-Fetch-Site: cross-site'` → 403.
- [ ] Sixth `POST /login` in 1 min from one IP → 429 with `Retry-After` and `RateLimit-Remaining: 0`.
- [ ] `go test ./... -run 'TestSecurity_'` → PASS.
- [ ] `go tool govulncheck ./...`, `osv-scanner`, `bun audit --audit-level=high` → exit 0.
- [ ] `bash qa/zap/run.sh staging` → exit 0.
- [ ] `bunx playwright test headers csp` → pass, 0 CSP violations.
- [ ] A JPEG uploaded with GPS EXIF → `exiftool` on the stored file shows no EXIF.
- [ ] An export link with an altered `signature`, or used after 24 h → 403.
- [ ] Admin without 2FA → redirected to setup; idle 31 min → login.

## 10. Test strategy

- Unit: limiter keys, PoW verifier (fake clock, replay, wrong route), CSP builder, redaction.
- Integration (Postgres): limits survive a restart, Stripe replay, rotation over encrypted columns, deletion replay after restore.
- Contract: Schemathesis `--checks all` on `api-v1` and `app-v1`.
- E2E: Playwright headers, CSP, consent banner, admin 2FA; axe on the challenge notice.
- Load: k6 50 VUs on `/login` with PoW active (RSS ≤ 512 MiB).
- Security: ZAP, `TestSecurity_*`, dependency scans, yearly review.

## 11. Open issues

1. Export conflict: SDD 03 has a sync JSON export (1/h); SDD 14 has a ZIP (7 days). SDD 13 adopts SDD 14. Owner: architect. Closes when SDD 03 points to SDD 14.
2. CSRF token fallback: SDD 03 says on, SDD 09 §7.8 says off. SDD 13 picks on. Owner: architect. Closes when SDD 09 is edited.
3. Lint tool: SDD 09 says Biome, SDD 11's `lint` job says ESLint. Owner: architect. **Closed 2026-10-29:** SDD 11 §6.6 says Biome.
4. The brief asked for consent for personalisation, but D-29 binds legitimate interest + opt-out. Owner: owner. Closes with the D-21 legal review.
5. LGPD articles and the ANPD incident deadline were not fetched (planalto.gov.br reset the connection). Owner: owner/legal. Closes when they are cited verbatim.
6. Encarregado (DPO) identity, contact and small-agent exemption. Owner: owner. Closes when the policy names them.
7. Role-scoped 2FA, session lifetime and IP allowlist `[NOT IN VUKEM DOCS]`; file a VF entry. Owner: vukem. Closes when vukem ships them or `TestSecurity_A07_HexaAdminSessionExpiry` passes on app middleware.
8. No documented re-seal command after key rotation `[NOT IN VUKEM DOCS]`; VF candidate. Owner: vukem. Closes when documented or the app job's test passes.
9. No antivirus for uploads (D-08); accepted with §7.3 controls. Owner: owner. Revisit if uploads go beyond moderated images.
10. HSTS `preload` deferred. Owner: owner. Revisit 2027-06.
11. The restore runbook must re-apply deletions made after the dump. Owner: architect (SDD 11). Closes when the monthly drill covers it. Status 2026-10-29: `backup:restore` replays deletions, schedules and cancellations from the running database (`TestRestoreReappliesDeletions`, P2.E.3.1, runbook `docs/runbooks/restore.md`); stays open until the first staging drill.
12. Turnstile in production needs an ADR 0016 update and a policy line. Owner: owner. Closes on the ADR edit.
