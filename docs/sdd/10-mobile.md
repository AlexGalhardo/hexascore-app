# SDD 10: Mobile app (Android first)

Status: reviewed (2026-10-28 consistency pass). Decisions: D-00, D-12, D-13, D-17, A-01. ADRs: 0005, 0008, 0009. Related SDDs: 02, 03, 05, 07, 08, 09, 12, 13. External sources [E1]–[E18] and vukem pages [V1]–[V4] (accessed 2026-10-28) are listed after §11.

## 1. Objective

Fastest way on Android to find a title, rate it with one tap and track it (D-12). A logic-free client over `/app/v1` (A-01) using the generated client, shared tokens and catalogues (ADR 0009), with the web's "editorial data" look (`legacy/docs/design-system.md` "Princípios"), portable to iOS.

## 2. Scope

In scope (v1, D-12):

- Login with email/password plus TOTP or a recovery code, and Google. Signup and recovery open web `/signup` and `/recover-password` in the browser (SDD 03 §2, §6.1).
- Browsing: home trending top 10 per type (D-02), listings, item page (stars, categories, book editions, entities, similar), entities, search with suggest.
- Account actions: rating, status, time-to-beat, list CRUD, own profile, notifications list, account deletion.

Out of scope: billing, API keys, admin (D-12); edit suggestions (later, SDD 04 `suggestions:write`); feed/follows (D-06); push; biometrics; offline list editing beyond single add/remove; iOS (later; Android-only APIs only behind `src/platform/`).

## 3. Functional requirements

- **FR-10-1:** Screens call only the §6.1 operations, and only via `@hexascore/app-client`. Verify: a lint rule bans `fetch(` outside `src/api/`; contract check.
- **FR-10-2:** The token comes from password (+2FA) or a Google ID token and lives only in SecureStore. Verify: Maestro `login.yaml`; Jest test that the token never enters the query cache.
- **FR-10-3:** The star widget sets 0.5–10.0 in 0.5 steps with one tap and fires a selection haptic. It is `adjustable` for TalkBack (§7.5). Verify: Jest + RNTL; TalkBack checklist.
- **FR-10-4:** Rating, status and time-to-beat writes are optimistic, queued offline, and the server wins on conflict (§7.3). Verify: Jest; Maestro airplane-mode flow.
- **FR-10-5:** A blocking update screen appears on a 426 `app-upgrade-required` or when `X-App-Min-Version` exceeds the running version. Logout still works there. Verify: Jest; staging with `APP_MIN_SUPPORTED_VERSION` raised.
- **FR-10-6:** The locale is the first supported entry of `getLocales()` [E10], matched by language, else `en`. An in-app override is saved via `PATCH /me {locale}`. Verify: Jest key-completeness.
- **FR-10-7:** Web URLs `/{,pt-br/,es/}{games|movies|series|books}/{slug}` and `/entities/{id}` open in-app via verified App Links. Verify: `adb shell pm get-app-links app.hexascore` → `verified`.
- **FR-10-8:** Profile → Delete account calls `POST /me/deletion` after a 10 s countdown (FR-03-15). The listing carries a web URL (§7.8). Verify: Maestro.
- **FR-10-9:** TMDB/IGDB attribution appears on the item and About screens (`open-issues.md` 01-6). Verify: Maestro.
- **FR-10-10:** Listings follow `legacy/docs/frontend.md` "Padrão de listagens" (per-type sections with icon label, cover thumb left of titles). Verify: RNTL.

## 4. Non-functional requirements

Reference device class: Android 12+, 4 GB RAM, 1080p, 2021–22 mid-range SoC (Snapdragon 6-series / Helio G8x; Galaxy A52, Moto G Power class); one unit kept for release checks. Budgets use a release build on it.

| ID | Budget | Measured by |
| --- | --- | --- |
| NFR-10-1 Cold start | p90 of 10 runs ≤ 2.5 s to first home frame (cached), ≤ 3.5 s (uncached) | `adb shell am start -W` `TotalTime` |
| NFR-10-2 Scroll | 60 fps target, ≤ 5 % janky frames on a 200-row fling | `dumpsys gfxinfo app.hexascore` |
| NFR-10-3 Item page, 4G cache miss | ≤ 1 s p75 | Maestro timing |
| NFR-10-4 Size | Hermes bundle ≤ 4 MB; Play download ≤ 25 MB | EAS artifact, Play Console |
| NFR-10-5 Memory | ≤ 250 MB PSS after 5 min | `dumpsys meminfo` |
| NFR-10-6 Query cache | ≤ 10 MB, 24 h max age | Jest |
| NFR-10-7 Image cache | ≤ 150 MB disk | release checklist |

Also: AA contrast (4.5:1) both themes, targets ≥ 44 × 44 dp, 200 % font scale without clipping; HTTPS only; no secrets in the bundle (`EXPO_PUBLIC_*` = API base URL and Google client id, both public).

## 5. Data model

There are no server tables. Device stores:

| Store | Content | Limit |
| --- | --- | --- |
| SecureStore (Keystore-encrypted [E1]) | `token`, `tokenId`, `expiresAt` | ≈ 100 B, under the ~2048 B ceiling [E1] warns of; wiped on logout/401/ban/deletion |
| Persisted query cache (`persistQueryClient` [E2]) on Expo SQLite (approved, D-22) | read queries only; no tokens, no problem bodies | 10 MB, `maxAge` 24 h, `buster` = app version + `formulaVersion` |
| Mutation queue | `{id, op, itemId, payload, idempotencyKey, createdAt, attempts}` | 50 entries, 72 h |
| Preferences | locale override, theme, crash consent | until uninstall |

## 6. Contracts

### 6.1 Screens, routes and endpoints

Routes are Expo Router files under `apps/mobile/app/`. Paths are relative to `/app/v1`. Abilities follow the cross-SDD convention, and app tokens carry all five (SDD 05 §6.4).

| Screen (route) | Operations | Ability |
| --- | --- | --- |
| Login `(auth)/login` | POST `/tokens`, `/tokens/google` | none |
| Home `(tabs)/index` | GET `/{type}/trending` ×4 | `catalog:read` |
| Section `(tabs)/[type]` | GET `/{type}?sort=&genre=&cursor=`, `/{type}/genres` | `catalog:read` |
| Item `[type]/[slug]` | GET `/{type}/{id}`, `/items/{id}/me`, `/items/{id}/related`, `/books/{id}/editions` | `catalog:read` |
| Rate sheet | PUT/DELETE `/items/{id}/rating` | `ratings:write` |
| Status / time-to-beat sheet | PUT/DELETE `/items/{id}/status`, `/items/{id}/time-to-beat` | `status:write` |
| Entity `entities/[id]` | GET `/entities/{id}`, `/entities/{id}/items?type=&role=` | `catalog:read` |
| Search `(tabs)/search` | GET `/search`, `/search/suggest`; POST `/search/events` | optional |
| Lists `(tabs)/lists`, `lists/[id]`, add-to-list sheet | GET `/me/lists`, `/lists/{id}`; POST `/lists`; PATCH/DELETE `/lists/{id}`; PUT/DELETE `/lists/{id}/items/{itemId}`; PUT `/lists/{id}/order` | `catalog:read`, `lists:write` |
| Profile `(tabs)/profile` | GET/PATCH `/me` (incl. `personalizationOptOut`, SDD 03); GET `/tokens`; DELETE `/tokens/{id}`; POST/DELETE `/me/deletion`; POST `/me/export` (202, ZIP link mailed, SDD 14 §7.4) | `catalog:read`, `profile:write` |
| Notifications | GET `/me/notifications`; POST `/me/notifications/read` (SDD 05 §6.4, SDD 14 §6.3) | `catalog:read`, `profile:write` |
| Update required | none; `market://details?id=app.hexascore` | none |
| Logout | DELETE `/tokens/current` | bearer |

Headers:

- `Authorization: Bearer hx_app_…`
- `X-App-Version: <semver>`
- `Accept-Language`
- `Idempotency-Key` (UUID v7) on writes ([V4] "Every answer": replay on same key; 409 while in flight; 422 if reused for a different body)

RFC 9457 `code` values (kebab-case) map to `errors.<code>`. An unknown code shows `errors.unexpected` plus `request_id`.

### 6.2 Token lifecycle (SDD 03 §7.4: 60-day life, no refresh token, rotate ≤ 10/h)

| Step | Trigger | Call | Client behaviour |
| --- | --- | --- | --- |
| Issue | Login | POST `/tokens` `{identifier, password, deviceName, code?, recoveryCode?, challenge?}` | `two-factor-required` → code field; 428 `challenge-required` → browser challenge; store result |
| Issue (Google) | Google sign-in | POST `/tokens/google` `{idToken, deviceName}` | same 2FA branch; 409 `email-unverified-at-provider` → message |
| Device name | Issue | — | default `HexaScore · <model>` (issue 10), editable, 1–64 chars |
| Rotate | Foreground with `expiresAt − now < 15 d`, max once per 24 h | POST `/tokens/rotate` | swap in SecureStore; retry next foreground on failure |
| 401 | Any call | — | wipe token and cache → login |
| 403 `account-banned` | Any call | — | wipe; show reason and end date |
| 403 `email-not-verified` | A write | — | banner linking to web verification |
| Deletion scheduled | `GET /me` returns `deletionScheduledAt` | — | cancel screen → web `/login` (SDD 03 §8) |
| Logout | User | DELETE `/tokens/current` | wipe locally even if the call fails |
| Revoke a device | Profile → Devices | DELETE `/tokens/{id}` | refetch |

Tokens never reach logs, the cache or crash reports.

## 7. Behaviour and algorithms

### 7.1 Stack decision

- **Expo managed (CNG) + EAS Build/Submit/Update.** Native projects are generated from app config [E3], so no `android/` to maintain; config plugins cover App Links; EAS gives signing, remote `versionCode` and OTA. Every module needed is in the Expo SDK (v58 current [E1]); no cited constraint forces bare RN. TypeScript strict.
- **Hermes** (RN default [E4]); CI asserts `global.HermesInternal`.
- **Expo Router** [E5]. It is file-based, has typed routes and automatic deep linking, and is built on React Navigation. Paths mirror the web URLs (D-17), so App Links need no mapping.
- **TanStack Query** with legacy defaults (`legacy/docs/frontend.md` line 50): `staleTime` 30 s, `retry` 1; refetch on foreground via `focusManager`.
- **FlatList** (no new dependency): fixed-height rows, `getItemLayout`, memoised `renderItem`, tuned `windowSize` [E6]. FlashList only via ADR if NFR-10-2 fails.
- Dependencies: latest stable at implementation, pinned exactly in `package.json`; Bun workspaces.

### 7.2 Shared assets (ADR 0009)

`packages/tokens` plain objects feed `StyleSheet.create` (no shared components): light/dark colours, spacing, radii, type scale (serif display, sans text, tabular numbers), category accents, icon names, and score bands (`legacy/docs/design-system.md` "Cor"), always shown with the number:

| Score (0–10) | Band |
| --- | --- |
| < 5.0 | `score-low` |
| 5.0–7.4 | `score-mid` |
| ≥ 7.5 | `score-high` |
| `null` | `score-none` |

i18n: the app consumes the flat catalogues produced once by the shared `packages/i18n` build step (same flattening and fallback chain as the server's `inertia.I18n`, SDD 09 §7.1, open issue 09-5), copied at build time. Bundled namespaces (as in `legacy/apps/web/messages/en/`): `common`, `auth`, `account`, `catalog`, `categories`, `item`, `lists`, `ratings`, `search`, `errors`. Excluded: `api`, `billing`, `docs`, `openapi`, `contact`, `legal` (web pages), `social` (D-06). ICU library as SDD 09 picks.

### 7.3 Offline, optimistic writes, conflicts

- **Reads:** the persisted cache renders instantly. An offline banner shows the network state via `src/platform/`.
- **Queued writes:** rating, status, time-to-beat and single list add/remove are optimistic, each with an `Idempotency-Key` set at creation; FIFO drain on reconnect/foreground, backoff 2 s, 10 s, 60 s; a newer write for the same `(itemId, op)` replaces the queued one and gets a fresh key (reusing it with a new body is a 422 [V4]); a 409 in-flight answer is retried, not rolled back.
- **Conflict rule:** the server wins.

| Response | Action |
| --- | --- |
| 2xx | replace the cache entry with the server body |
| 409 (`rating-not-released`, `status-not-completed`) or 422 | roll back to server state, toast |
| 429 `item-ratings-limited` | roll back, show `Retry-After` (SDD 02 §7) |
| 404 | drop |
| 5xx after 5 attempts, or older than 72 h | drop, toast |

- **Lists:** create, rename, delete and reorder are disabled offline.
- **Images:** `expo-image` with `cachePolicy="memory-disk"` [E7]. Rows use `coverThumbUrl` (96×144) at 44 × 64 dp, the item page `coverUrl` (600 px) scaled down. Only these two variants exist (SDD 06 §7.5, open issue 10-3); a 2× thumb (192×288) is added post-launch only if the scroll budget fails.

### 7.4 Versioning and forced update

- `version` is SemVer, tagged `mobile-vX.Y.Z` (D-20 pattern).
- `versionCode` uses the EAS remote source with `autoIncrement` on production [E8].
- The client checks `X-App-Min-Version` on every response. Raising `APP_MIN_SUPPORTED_VERSION` is the only kill switch (SDD 05 §6.4).

### 7.5 Star widget (mirrors SDD 09; ADR 0005)

- **Layout:** ten stars with left/right halves; a tap on star `n` sets `n − 0.5` or `n`; re-tapping the value is a no-op ("Remove rating" is separate); no drag in v1. Stars 32 dp + 6 dp hit slop (targets ≥ 44 dp); wraps below 360 dp.
- **Haptics:** `Haptics.selectionAsync()` per change [E9].
- **Accessibility [E11]:** one element, `accessibilityRole="adjustable"`, `accessibilityValue={{min: 1, max: 20, now: value*2, text: t('rating.valuetext', {value})}}` (doubled: `now` is an integer), increment/decrement ±0.5, label = localised category.
- **Rate sheet:** the type's six categories (A-06) in one PUT with all keys (SDD 02 §6), submit disabled until complete. Aggregate: `score` (one decimal), `ratingCount`, `confidence`; "–" when `null` (FR-02-6).

### 7.6 Deep links

`android.intentFilters` declare `autoVerify: true` for host `hexascore.app` and these prefixes: `/games`, `/movies`, `/series`, `/books`, `/entities`, `/lists`, `/pt-br`, `/es`.

The Go app serves `GET /.well-known/assetlinks.json` [E12] via `r.Get` ([V1] "Defining routes"):

- no auth, no redirect;
- `application/json`, `Cache-Control: public, max-age=3600`;
- built from `ANDROID_PACKAGE=app.hexascore` and `ANDROID_CERT_SHA256`, the Play signing and upload fingerprints from `eas credentials -p android`.

### 7.7 Security, analytics, crash reports

- **No certificate pinning at launch:** Railway rotates certificates (cadence [NOT VERIFIED], SDD 11) and a stale pin bricks installed builds; tokens are already scoped, expiring and revocable. SDD 13 revisits if tokens gain billing powers.
- Screenshots allowed (public data); biometrics later behind `src/platform/`.
- **No third-party analytics:** metrics from `/app/v1` traffic tagged by `X-App-Version` (SDD 12); client events only SDD 07 `search.performed`/`search.clicked`; views counted server-side (SDD 08 FR-08-1, issue 5).
- **Crash reports:** opt-in at first launch (LGPD, default off, revocable). JS handler + error boundary POST `{appVersion, versionCode, androidApi, screen, message, stack ≤ 8 KB}`, no user id or token, to `POST /app/v1/client-errors` (SDD 05 §6.4, SDD 12 §6.4; 204, kept 7 days). Native capture (Sentry or similar) needs an ADR.

### 7.8 Release engineering

- **Build service (D-34):** EAS Build on the free tier while it lasts; a Gradle pipeline on GitHub Actions (D-26) is kept ready as the fallback, building the AAB from `npx expo prebuild` output with the upload key from Actions secrets.
- **Pipeline:** profiles `development`, `preview` (APK), `e2e-test`, `production` (AAB). PR: lint, typecheck, Jest, contract, EAS Workflow build + Maestro [E13]. Tag `mobile-v*`: `eas build --profile production` + `eas submit` → internal testing → closed (≥ 7 days, crashes ≤ 1 % of sessions) → production staged 20 % → 100 % over ≥ 3 days.
- **Target API:** since 2026-08-31, new apps and updates target API 36 [E14]. The chosen Expo SDK must comply.
- **Account deletion [E15]:** directing users to web signup counts as in-app account creation, so deletion is offered in-app (FR-10-8) and on a web resource naming HexaScore that works without the app: `https://hexascore.app/help/delete-account` (SDD 09, issue 7), explaining the 30-day grace and the anonymised ratings kept (SDD 03 §7.5).
- **Data safety [E16]** (SDK traffic counts): collected, not shared, encrypted in transit, deletable: email and name; app interactions and user content (ratings, statuses, lists); in-app search history (30 days); crash logs (optional). Not collected: location, contacts, device IDs, financial data.
- **OTA:** `runtimeVersion` policy `fingerprint` [E17], one channel per profile, only for JS/asset, copy and catalogue fixes [E18]. A store release is mandatory for a fingerprint change, new permission, target API bump, new screen or feature, data-collection change, or a min-version raise.
- **Listing assets:** texts in three locales, 512 px icon, 1024 × 500 graphic, ≥ 4 screenshots per locale (Maestro `takeScreenshot`), privacy URL `/legal/privacy`.

## 8. Edge cases

| Case | Behaviour |
| --- | --- |
| Token expired in background | 401 → login; queue replays only for the same user, else discarded |
| Rotate response lost | old token revoked → 401 → re-login (accepted) |
| Rating queued before release | server 409 → rollback |
| Locale `pt-PT`, `es-MX`, `fr` | `pt-BR`, `es`, `en` |
| Merged item | follow 301, rewrite cache key |
| Empty editions or entity type group | hidden |
| Google sign-in cancelled | stay, no error |
| 426 with a pending queue | queue paused until update |

## 9. Acceptance criteria

- `bun run --cwd apps/mobile typecheck && bun run --cwd apps/mobile lint` → 0 errors.
- `bun run --cwd apps/mobile test` → pass, with coverage ≥ 80 % on `src/rating`, `src/queue` and `src/auth`.
- `vukem openapi:export -prefix /app/v1 | diff - openapi/app-v1.json` → empty ([V2]).
- `vukem codegen:client -prefix /app/v1 -o packages/app-client/src/api.ts && git diff --exit-code packages/app-client` → no diff, then typecheck green ([V3] "The command": runs in warmup mode, no secrets needed in CI).
- `maestro test .maestro/smoke/` (staging) → `login`, `rate-item`, `add-to-list`, `offline-rating` and `deep-link` pass.
- `curl -si https://staging.hexascore.app/.well-known/assetlinks.json` → 200 `application/json`, package `app.hexascore`, both fingerprints.
- `curl -si -H "X-App-Version: 0.0.1" -H "Authorization: Bearer $T" https://staging.hexascore.app/app/v1/me` → 426 `app-upgrade-required`.
- Reference device → TotalTime p90 ≤ 2500 ms (cached), ≤ 5 % janky frames, AAB download ≤ 25 MB.
- `grep -rE "hx_app_|sk_live|APP_KEY" apps/mobile/dist` → no match.

## 10. Test strategy

- **Unit:** Jest, `jest-expo` preset (Vitest on RN [NOT VERIFIED]): widget, per-locale formatters, locale matcher, queue (dedupe, backoff, rollback), rotation window, problem-code mapping.
- **Contract:** CI regenerates the client and typechecks; integration against staging fixtures (SDD 11).
- **E2E: Maestro over Detox.** EAS Workflows run Maestro natively [E13]; YAML flows need no native harness (Detox does, and CNG has no `android/`); flows carry over to iOS.
- **Accessibility:** role/label queries in screen tests; per-release TalkBack + 200 % font check; Accessibility Scanner on 8 screens, 0 findings.
- **Release checklist:** target API, data safety, assetlinks, deletion URL, min version, OTA rule, staged rollout.

## 11. Open issues

1. `/me/notifications`. Decided (10-1): owned by SDD 14 §6.3, listed in SDD 05 §6.4.
2. `POST /app/v1/client-errors`. Decided (10-2): in SDD 05 §6.4 and SDD 12 (opt-in, rate-limited, 7-day retention, no PII).
3. Density cover widths. Decided (10-3): 600 px + 96×144 only (§7.3).
4. RN icon set and ICU library. Decided (10-4): follow SDD 09 (§7.5 icons, §7.1 message format).
5. App views. Decided (10-5): counted per token id per day (SDD 08 FR-08-1).
6. Persister backend. Decided: Expo SQLite (D-22).
7. `/help/delete-account`. Decided (10-7): in SDD 09 §6.2 and SDD 03 FR-03-15.
8. EAS cost. Decided (D-34): free tier, Gradle-on-Actions fallback (§7.8).
9. SDD 03 abilities and codes. Fixed (10-9): README abilities, kebab-case codes.
10. `deviceName` source. Roadmap (10-10): resolved at implementation with the Expo Device API (model name), cited then.

References (accessed 2026-10-28):

- [E1] <https://docs.expo.dev/versions/latest/sdk/securestore/>
- [E2] <https://tanstack.com/query/latest/docs/framework/react/plugins/persistQueryClient>
- [E3] <https://docs.expo.dev/workflow/prebuild/>
- [E4] <https://reactnative.dev/docs/hermes>
- [E5] <https://docs.expo.dev/router/introduction/>
- [E6] <https://reactnative.dev/docs/optimizing-flatlist-configuration>
- [E7] <https://docs.expo.dev/versions/latest/sdk/image/>
- [E8] <https://docs.expo.dev/build-reference/app-versions/>
- [E9] <https://docs.expo.dev/versions/latest/sdk/haptics/>
- [E10] <https://docs.expo.dev/versions/latest/sdk/localization/>
- [E11] <https://reactnative.dev/docs/accessibility>
- [E12] <https://docs.expo.dev/linking/android-app-links/>
- [E13] <https://docs.expo.dev/eas/workflows/examples/e2e-tests/>
- [E14] <https://support.google.com/googleplay/android-developer/answer/11926878>
- [E15] <https://support.google.com/googleplay/android-developer/answer/13327111>
- [E16] <https://support.google.com/googleplay/android-developer/answer/10787469>
- [E17] <https://docs.expo.dev/eas-update/runtime-versions/>
- [E18] <https://docs.expo.dev/eas-update/introduction/>
- [V1] <https://vukem.dev/docs/basics/routing>
- [V2] <https://vukem.dev/docs/rest-apis/openapi>
- [V3] <https://vukem.dev/docs/rest-apis/client-generation>
- [V4] <https://vukem.dev/docs/rest-apis/idempotency>
