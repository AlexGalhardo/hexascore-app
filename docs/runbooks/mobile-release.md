# Mobile release (EAS)

How to build, submit and update the Expo app in `apps/mobile` with EAS (Expo Application Services). Design and policy live in SDD 10 §7.8 (pipeline, staged rollout) and D-34 (EAS free tier, Gradle-on-Actions fallback). This runbook is the operating procedure for people and agents.

Scope today: **Android only** (SDD 10 §1, iOS is later). Every `--platform` below is `android`; when iOS enters, the same steps run with `--platform ios` (or `all`) after its credentials exist.

| Piece | Where |
| --- | --- |
| App config, `version` | `apps/mobile/app.config.ts` |
| Build and submit profiles | `apps/mobile/eas.json` (`development`, `preview`, `e2e-test`, `production`) |
| `versionCode` | EAS remote source, `autoIncrement` on `production` (SDD 10 [E8]); never edit it by hand |
| CI | `.github/workflows/mobile-release.yml` (EAS build, owner-gated submit); `.github/workflows/mobile-gradle.yml` (Gradle fallback, manual) |
| CLI | `eas-cli` pinned `27.2.0`, run as `bunx eas-cli@27.2.0` (no global install needed) |

## Hard rules

- Never ask for, print, paste or commit passwords, tokens (`EXPO_TOKEN`), keystores or the Play service account key. Login is interactive and done by the owner (`bunx eas-cli@27.2.0 login`); CI reads `EXPO_TOKEN` from a GitHub secret; the Play key lives in EAS credentials (`eas credentials -p android`).
- Stop and ask the owner when: the git tree has uncommitted changes; a `production` version change is proposed; a build fails (show the diagnosis before any fix); before **every** `eas submit`; before any `eas update` on a production channel.
- `eas submit` sends to the Play **internal** track as a **draft** (`eas.json` `submit.production`). Promotion to closed or production tracks, and the staged rollout (20 % then 100 %, SDD 10 §7.8), is a separate owner decision in Play Console.
- Never delete builds, channels, branches or credentials without an explicit owner yes in chat.
- Agents prepare and run checks; the owner (or the owner's explicit "sim" for that exact build) triggers paid or outward-facing steps.

## Step 0: environment check

Run from `apps/mobile` and report the result before going further.

```sh
bunx eas-cli@27.2.0 --version
bunx eas-cli@27.2.0 whoami        # not logged in: ask the owner to run `bunx eas-cli@27.2.0 login`
bun install --frozen-lockfile     # from the repo root
bunx expo-doctor
git status --porcelain            # non-empty: stop and ask
```

## Step 1: version

1. Read `version` in `apps/mobile/app.config.ts`.
2. For `production`, propose the new SemVer to the owner and wait for the answer before editing.
3. Edit only `version`. `versionCode` is remote and auto-incremented by EAS.
4. Commit as `chore(mobile): bump app version to X.Y.Z`. Mobile releases are tagged `mobile-vX.Y.Z` (SDD 10 §7.8).

## Step 2: tests and checks

```sh
bun run lint
bun run typecheck
bun run test -- --runInBand --forceExit
```

Any failure: stop and show the error. Do not build.

## Step 3: build

```sh
bunx eas-cli@27.2.0 build --platform android --profile <profile> --non-interactive --no-wait
bunx eas-cli@27.2.0 build:list --platform android --limit 1 --json --non-interactive
```

- Default profile is `production` (AAB). `preview` gives an APK for owner testing.
- Follow the build with `build:view <id>` until it finishes. On failure read the logs (`build:view <id>` links them), name the cause, and show the diagnosis before trying a fix.
- Report the build URL, `version` and the remote `versionCode`.

## Step 4: store submission (owner "sim" per build)

1. List exactly what goes out: platform, profile, `version`, `versionCode`, build id.
2. Only after the owner answers "sim" for that build:

   ```sh
   bunx eas-cli@27.2.0 submit --platform android --id <build-id> --profile production --non-interactive
   ```

   Prefer `--id` over `--latest` so the build that was confirmed is the one sent.
3. The result is a draft on the internal track. Review and rollout stay with the owner.

## Step 5: OTA update (JS and assets only)

Not available yet: `expo-updates` is not installed and needs owner approval as a new dependency (SDD 10 §7.8 plans `runtimeVersion` policy `fingerprint`, one channel per profile). Until then every change ships as a store build.

Once it is installed:

1. A store build is mandatory for a fingerprint change, a new native dependency or config plugin, a new permission, a target API bump, a new screen or feature, a data-collection change or a min-version raise (SDD 10 §7.8). Check with `bunx expo-updates fingerprint:generate` against the last build.
2. Otherwise:

   ```sh
   bunx eas-cli@27.2.0 update --branch <branch> --message "<summary>" --non-interactive
   ```

3. Ask the owner before publishing to the `production` branch.

## Gradle fallback on GitHub Actions (D-34)

Used only when EAS Build is unavailable or stops being free (SDD 10 §7.8). `.github/workflows/mobile-gradle.yml` is `workflow_dispatch` only. It runs lint, typecheck, Jest and `bundle:check`, then `expo prebuild` and `./gradlew bundleRelease assembleRelease` on `ubuntu-24.04` with Temurin JDK 21, and uploads two artifacts kept for 7 days: `app-release.aab` (Play) and `app-release.apk` (sideload). It never submits to Play.

```sh
bunx eas-cli@27.2.0 build:version:get --platform android   # owner: last remote versionCode, use it + 1
gh workflow run mobile-gradle.yml -f ref=main -f version_code=<N>
gh run watch
gh run download <run-id> -n app-release.aab
unzip -l app-release.aab | rg BundleConfig.pb              # one match
```

- Signing: the upload key comes from four Actions secrets, `ANDROID_UPLOAD_KEYSTORE_BASE64`, `ANDROID_UPLOAD_KEYSTORE_PASSWORD`, `ANDROID_UPLOAD_KEY_ALIAS` and `ANDROID_UPLOAD_KEY_PASSWORD`. The job fails before checkout when any is empty, and fails after the build when the AAB or the APK is not signed with that key (SHA-256 compared with `keytool`/`apksigner`). The keystore is decoded to `$RUNNER_TEMP` and deleted at the end; only the steps that sign or verify see the secrets.
- `versionCode`: EAS keeps it remotely, so the run takes it as the `version_code` input (exported as `ANDROID_VERSION_CODE`, read by `app.config.ts`). It must be higher than every build already sent to Play. After a fallback upload, set the EAS remote value to the same number (`eas build:version:set`) so the next EAS build does not collide.
- The repository is public: artifacts can be downloaded by any signed-in GitHub user while they are kept. Treat a fallback AAB as published.
- The generated `android/` folder and the keystore are never committed.
- Windows-only workarounds from the local sideload build (ninja long paths, `NODE_PATH`) are not needed on Linux.

### Owner setup for the fallback (one time)

1. Download the EAS-managed upload keystore: `bunx eas-cli@27.2.0 credentials -p android` > production > Keystore > Download. Note the alias and both passwords it prints.
2. Store it in GitHub secrets without echoing it: `base64 -w0 <keystore.jks> | gh secret set ANDROID_UPLOAD_KEYSTORE_BASE64`, then `gh secret set ANDROID_UPLOAD_KEYSTORE_PASSWORD`, `ANDROID_UPLOAD_KEY_ALIAS` and `ANDROID_UPLOAD_KEY_PASSWORD` (each prompts for the value).
3. Delete the local keystore copy.

## Final report

Short summary: what ran, `version` and `versionCode`, build links, submission status, open items.

## One-time owner setup

1. Expo account and project: `bunx eas-cli@27.2.0 login`, then `bunx eas-cli@27.2.0 init` in `apps/mobile` (writes `extra.eas.projectId` into the app config; commit it).
2. Android credentials: `bunx eas-cli@27.2.0 credentials -p android` (upload keystore managed by EAS). Record the SHA-256 fingerprints for `ANDROID_CERT_SHA256` (SDD 10 §7.4).
3. Play Console app `app.hexascore` and a service account JSON key uploaded to EAS for `eas submit` (never committed).
4. GitHub: secret `EXPO_TOKEN` (an Expo robot or personal access token), and environment `play-store` with the owner as required reviewer; the submit job waits on it.
