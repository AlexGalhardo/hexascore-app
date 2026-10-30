# ADR 0018: NativeWind styles the Android app

Status: accepted (2026-10-30). Amends the React Native styling sentence of ADR 0009 only. Decisions: owner approval of NativeWind on 2026-10-30; SDD 10 §7.1, §7.2; D-12.

## Context

ADR 0009 shares design tokens and i18n catalogues with the mobile app and refuses a shared UI package (React Native Web, Tamagui, or NativeWind as a universal component tree). The web app is Inertia + React on Tailwind CSS 4.3.3. The Android app is a separate Expo bundle.

The owner approved NativeWind on 2026-10-30 as the mobile styling layer, not as a component library. The current NativeWind Expo install guide (`https://www.nativewind.dev/docs/getting-started/installation`, accessed 2026-10-30) documents stable v4.2.7. NativeWind v5 was still a release candidate on that page, so it is not used.

Google Play has required target API 36 for new apps and updates since 2026-08-31 (SDD 10 [E14]). The Expo SDK reference (`https://docs.expo.dev/versions/latest/`, accessed 2026-10-30) lists:

| SDK | compileSdkVersion | targetSdkVersion |
| --- | --- | --- |
| 58.0.0 | 37 | 36 |
| 57.0.0 | 36 | 36 |
| 56.0.0 | 36 | 36 |

`expo@58.0.21` was the latest stable `expo` release on npm that day. SDK 58 still sets `targetSdkVersion` to 36. `compileSdkVersion` 37 is the compile surface, not the Play target. SDK 57 and 56 also target API 36; 58 is the latest stable that does.

## Decision

- The Android app (`apps/mobile`, package `app.hexascore`) is Expo SDK 58.0.21, managed (CNG). Native `android/` and `ios/` directories are not committed. Hermes is the only JavaScript engine SDK 58 ships. The app is Android-only: no `ios` config, and Android-only APIs stay behind `src/platform/`.
- Styling is NativeWind 4.2.7. Colours, spacing, type scale, radii, category accents and score bands come from `packages/tokens`. The app does not invent a second palette and does not restyle the web app. There is still no shared UI component package (ADR 0009).
- React is `19.3.1`, the version the workspace already pins. The SDK 58 bundle lists React `19.3.0`; `react-native@0.88.1` accepts `^19.3.0`. One React copy avoids a second renderer in the Bun workspace.
- Versions are exact. Packages this ADR adds, and why:

| Package | Version | Why |
| --- | --- | --- |
| `nativewind` | 4.2.7 | Owner-approved styling layer. Latest stable (v5 is an RC). |
| `tailwindcss` | 3.4.19 | Peer of NativeWind 4. Latest 3.4. The web app stays on Tailwind 4.3.3; this copy is only the mobile workspace. |
| `react-native-reanimated` | 4.7.1 | NativeWind install peer. Expo SDK 58 bundled version. |
| `react-native-worklets` | 0.13.0 | Peer of Reanimated 4.7.1 (`0.13.x`). The NativeWind install note for Expo 57+ says to install it with Reanimated 4. SDK 58 pins this version. |
| `react-native-safe-area-context` | 5.10.1 | NativeWind install peer. SDK 58 range `~5.10.1`. |
| `babel-preset-expo` | 58.0.22 | NativeWind install step that sets `jsxImportSource`. |
| `@babel/core` | 7.29.7 | Runtime for that preset. Babel 8 is current on npm but the preset's plugins are Babel 7. |
| `expo-router` | 58.0.28 | SDD 10 navigation. SDK 58 range `~58.0.28`. |
| `expo-linking` | 58.0.15 | Peer of `expo-router`. |
| `expo-constants` | 58.0.14 | Peer of `expo-router`. |
| `@expo/metro-runtime` | 58.0.15 | Peer of `expo-router`. |
| `react-native-screens` | 4.30.1 | Peer of `expo-router`. SDK 58 range `~4.30.0`. |
| `react-native-gesture-handler` | 3.3.0 | Peer of `expo-router`. SDK 58 range `~3.3.0`. |
| `jest-expo` | 58.0.13 | SDD 10 test runner. SDK 58 range `~58.0.13`. |
| `jest` | 30.5.2 | Peer of `jest-expo` (`^29` or `^30`). Latest stable 30. |
| `@testing-library/react-native` | 14.1.0 | SDD 10. Latest stable. |
| `react-test-renderer` | 19.3.1 | Renderer Jest uses with React 19.3.1. |
| `@types/react` | 19.3.0 | Matches the workspace. |
| `@types/jest` | 30.0.0 | Jest 30 globals for `tsc`. |
| `@types/node` | 26.6.6 | Types for `scripts/copy-catalogues.ts` (`node:fs`) only. |

`react-native-css-interop@0.2.7` is a dependency of `nativewind@4.2.7`, not a direct dependency. `prettier-plugin-tailwindcss` appears in the install command and is not used by any config step, so it is not installed.

No analytics package and no Expo template extras.

## Alternatives considered

- **React Native `StyleSheet` fed by the token objects (ADR 0009 as written).** No extra dependency. The owner chose NativeWind instead so mobile layout uses the same utility language as the token scale, without sharing components with the web.
- **NativeWind v5 RC.** Tailwind 4 and the v5 CSS engine. Rejected: pre-release.
- **Expo SDK 57.** Also targets API 36, and NativeWind 4.2.7 names SDK 57 explicitly. Rejected: SDK 58 is the newer stable release and still targets API 36.
- **A shared component package.** Rejected again. ADR 0009 stands.

## Consequences

- Positive: one token source; mobile screens can use utility classes; Play's API 36 target is the SDK default.
- Negative: the mobile workspace pins Tailwind 3 while the web app uses Tailwind 4. They must stay in different dependency trees. NativeWind 4.2.7's install note names SDK 57, not 58; the JS styling pipeline is the documented one, and a native rebuild is out of scope until a store build. Bun's isolated install does not hoist `babel-preset-expo`'s Babel plugins or `nativewind`'s `react-native-css-interop` to the app root. `apps/mobile/scripts/run-jest.cjs` puts those two package folders on `NODE_PATH` before Jest starts. That script is not a new dependency.
- Obligations: do not import web components into `apps/mobile`. Do not add a npm package the SDD does not name unless the owner accepts it here. Merge of the app shell waits on P3.C.5.2 (`@hexascore/app-client` is not the client this phase calls yet).
