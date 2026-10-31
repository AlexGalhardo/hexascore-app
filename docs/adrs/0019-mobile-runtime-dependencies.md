# ADR 0019: Mobile runtime dependencies for auth, cache and media

Status: accepted (2026-10-31). Decision: owner approval Q-B1 (2026-10-31, `docs/orchestration/QUESTIONS.md`); SDD 10 §5, §6.2, §7; ROADMAP P5.A.3 to P5.B.

## Context

The P5.A.3.2 to P5.B rows need secure token storage, device naming, Google sign-in, a query cache with persistence, images and haptics. The owner approved this set only; any other package needs a new question.

## Decision

All versions are exact (no `^` or `~`). Expo packages were installed with `bunx expo install`, so they are the SDK 58 versions. The two TanStack packages are the latest stable releases.

| Package | Version | Why |
| --- | --- | --- |
| `expo-secure-store` | 58.0.2 | Token in Android Keystore (P5.A.3.2). |
| `expo-device` | 58.0.5 | `deviceName` on login (P5.A.3.3). |
| `expo-auth-session` | 58.0.16 | Google sign-in, first-party Expo library (P5.A.3.4). |
| `@tanstack/react-query` | 5.105.0 | Query cache with the legacy defaults (P5.A.4.1). |
| `@tanstack/react-query-persist-client` | 5.105.0 | `persistQueryClient` and persister contract; the storage is a custom SQLite persister. |
| `expo-sqlite` | 58.0.15 | Persisted cache store (10 MB, 24 h) and offline queue. |
| `expo-image` | 58.0.18 | `memory-disk` cover caching (P5.B.1.1). |
| `expo-haptics` | 58.0.7 | Selection haptic on the star widget (P5.B.2.2). |

The config plugins `expo-secure-store`, `expo-sqlite` and `expo-image` are added to `app.config.ts`.

## Alternatives considered

- **AsyncStorage persister.** Needs another package and has a 6 MB default cap on Android. Rejected for `expo-sqlite`.
- **Third-party Google sign-in library.** Native module, more config. Rejected for the first-party `expo-auth-session`.
