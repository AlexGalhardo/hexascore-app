# Google Play data safety facts (Android app)

Facts for the owner to copy into Play Console > App content > Data safety for `app.hexascore`. Source of truth: SDD 10 §7.8 "Data safety"; this file records how the code matches it. Re-check it whenever `apps/mobile/package.json` gains a dependency or the app sends a new field (that is a new answer, and a store release, not an OTA).

Checked against `apps/mobile` on 2026-11-01 (Expo SDK 58, React Native 0.88).

## Overview answers

| Question | Answer |
| --- | --- |
| Does the app collect or share any of the required user data types? | Yes (collects; does not share) |
| Is all user data encrypted in transit? | Yes: the app only calls `https://hexascore.app` (`EXPO_PUBLIC_API_BASE_URL`); Android blocks cleartext by default |
| Do you provide a way for users to request that their data be deleted? | Yes: in the app (Profile > Delete account, 10 s countdown, `POST /me/deletion`) and on the web at `https://hexascore.app/help/delete-account` (also `/pt-br/` and `/es/`), 30-day grace |
| Data shared with third parties | None. No analytics, ads, attribution or crash SDK is bundled; crash reports go to HexaScore's own server |

## Data types collected

All rows: collected, **not shared**, processed by HexaScore only, encrypted in transit, deletable.

| Play category > type | What exactly | Required or optional | Purposes | Where in code |
| --- | --- | --- | --- | --- |
| Personal info > Email address | sign-up and login email | Required for an account (browsing works without one) | App functionality, Account management | `src/auth/` |
| Personal info > Name | nickname shown on ratings and lists | Required for an account | App functionality, Account management | `src/profile/Settings.tsx` |
| App activity > App interactions | ratings, library statuses, list changes, notification read state | Optional (only when the user acts) | App functionality, Personalization (opt-out in Profile) | `src/rating/`, `src/status/`, `src/lists/`, `src/notifications/` |
| App activity > Other user-generated content | list names and descriptions | Optional | App functionality | `src/lists/` |
| App activity > In-app search history | queries sent to `/search` and `/search/suggest`; the server records `search.performed` and the app sends `search.clicked` through `POST /search/events`. The server keeps raw queries at most 30 days and stores clicks with a query hash | Optional | Analytics, App functionality | `src/screens/Search.tsx`, `src/search/events.ts`; server `search_events` (SDD 07 §7.8) |
| App info and performance > Crash logs | screen name, error message and stack (emails and tokens redacted), app version, versionCode, Android API level; no user id | Optional: sent only after the user grants crash-report consent | Analytics (stability) | `src/telemetry/crash.ts`, `src/telemetry/consent.ts` |

## Not collected

Location, contacts, calendar, photos and videos, audio, files, messages, health and fitness, financial info (billing stays on the web, D-12), web browsing history, and device or other IDs.

Notes for the form:

- The device **model name** (for example "Galaxy S20 FE") is sent as the default label of the login token so the user can recognise it in Profile > Devices. It is not an identifier and is not used for tracking; Play does not treat a model name as "Device or other IDs".
- Auth tokens are stored with `expo-secure-store` on the device; they are credentials, not a data type in the form.
- Google sign-in runs in the system browser (`expo-auth-session`); no Google SDK is bundled.

## Target API

Play requires target API 36 for new apps and updates since 2026-08-31 (SDD 10 [E14]). The Expo SDK 58 build targets 36: `aapt2 dump badging` on the release APK built on 2026-11-01 shows `targetSdkVersion:'36'`, `minSdkVersion:'24'` (`.orchestration/reports/APK-sideload-build.md`). Re-check on every Expo SDK upgrade.

## Owner items

- Fill the Play Console data safety form with the tables above.
- Add the Play app signing certificate SHA-256 (Play Console > App integrity) to the production `ANDROID_CERT_SHA256`, next to the upload key fingerprint.
- Set the account deletion URL in Play Console to `https://hexascore.app/help/delete-account`.
- Privacy policy URL: `https://hexascore.app/policy` (200 in production on 2026-11-01, also `/pt-br/policy` and `/es/policy`). SDD 10 §7.8 and ROADMAP P5.C.2.2 say `/legal/privacy`, which is a 404: the route is `/policy` (`routes/web.go`).
