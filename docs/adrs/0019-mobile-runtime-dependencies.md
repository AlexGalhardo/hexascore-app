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

## Amendment 1 (2026-11-01): fonts, icons and native UI

Owner approval Q-B3 (2026-11-01, in chat), for the design v1 implementation (`IMPL-mobile-v1-*`). Versions from Expo's SDK 58 bundled native modules list (`expo-font ~58.0.11`, `@expo/vector-icons ^15.0.2`, `@expo/ui ~58.0.25`), latest stable release in that range, pinned exactly; `bunx expo install --check` reports none of them.

| Package | Version | Why |
| --- | --- | --- |
| `expo-font` | 58.0.11 | Config plugin embeds the v1 families at build time (no runtime load, no fallback flash): Inter 400/600/700 and Playfair Display 700. |
| `@expo/vector-icons` | 15.1.1 | Ionicons glyphs for the theme toggle (moon/sun), rating stars and tab icons. |
| `@expo/ui` | 58.0.26 | Native Jetpack Compose sheets, pickers and switches for the v1 screens. |

The font files are self-hosted in `apps/mobile/assets/fonts/` with their OFL licences: Inter 4.1 static TTFs from the `rsms/inter` release, Playfair Display 700 static TTF from Google Fonts (upstream `clauseggers/Playfair` ships only variable fonts). Android cannot load the web's WOFF2 files, so the two apps use the same families from different file formats.

## Alternatives considered

- **AsyncStorage persister.** Needs another package and has a 6 MB default cap on Android. Rejected for `expo-sqlite`.
- **Third-party Google sign-in library.** Native module, more config. Rejected for the first-party `expo-auth-session`.
