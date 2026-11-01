---
name: hexascore-mobile-release
description: Release the HexaScore Expo app (apps/mobile) with EAS - environment check, version bump, checks, eas build, store submission and OTA update with owner confirmation gates. Use when asked to build, release, submit or OTA-update the mobile app, or to cut a mobile-v* version.
---

# HexaScore mobile release (EAS)

Follow [docs/runbooks/mobile-release.md](../../../docs/runbooks/mobile-release.md) step by step; the rules in `.claude/rules/mobile-release.md` always apply. For EAS concepts load `expo-overview` and `eas-workflows`; prefer the Expo MCP tools when the session has them, with `bunx eas-cli@27.2.0` as the fallback.

## Flow

1. **Step 0, environment.** Run in `apps/mobile`: `bunx eas-cli@27.2.0 --version`, `whoami`, `bunx expo-doctor`, `git status --porcelain`. Show the diagnosis to the owner before Step 1. Not logged in: ask the owner to log in themselves. Dirty tree: stop and ask.
2. **Step 1, version.** Read `version` in `app.config.ts`. For `production`, propose the new version and wait for the owner. `versionCode` is EAS remote (`autoIncrement`).
3. **Step 2, checks.** `bun run lint`, `bun run typecheck`, `bun run test -- --runInBand --forceExit`. Failure: stop and show it.
4. **Step 3, build.** `bunx eas-cli@27.2.0 build --platform android --profile <profile> --non-interactive --no-wait` (default `production`), then poll `build:view <id>`. Failure: read logs, diagnose, show before fixing. Report build URL and versions.
5. **Step 4, submit.** List platform, profile, version, versionCode, build id. Run `submit --platform android --id <build-id> --profile production --non-interactive` only after the owner says "sim" for that build. It lands as a draft on the internal track.
6. **Step 5, OTA.** Off until `expo-updates` is approved. Then: fingerprint unchanged and JS/asset only, else a store build; ask before the production branch.
7. **Report.** What ran, versions, build links, submission status, open items.

## Never

- Ask for, print or store credentials or tokens.
- Submit, promote, publish to production, or delete builds, channels, branches or credentials without the owner's explicit yes for that action.
- Bump `versionCode` by hand or install `eas-cli` globally from an unpinned tag.
