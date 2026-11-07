# Mobile: local Android debug run

Run the Expo app in `apps/mobile` as a **debug** build on an Android emulator, talking to the Go app on your own machine. For store and sideload builds see [mobile-release.md](mobile-release.md).

Tested on Windows 11 with JDK 21 (Temurin), Android Studio, SDK platform 36/37, NDK 27.1.12297006, an `android-33` `x86_64` Google APIs AVD (`omega-api33`), Docker Postgres 18.6 and `vukem serve` on port 3000. On macOS and Linux `android:dev` is plain `expo run:android --variant debug` (not tested here).

## What makes it work

- `EXPO_PUBLIC_API_BASE_URL` is the only runtime setting of the app. `bun run android:dev` defaults it to `http://10.0.2.2:3000`: inside the emulator `10.0.2.2` is the host's loopback, where `vukem serve` listens.
- The Go app only redirects foreign hosts (`CanonicalHost`) when `APP_ENV=production`, so a dev server (`APP_ENV=development`, the `.env.example` default) accepts the `Host: 10.0.2.2:3000` header with no extra setting.
- Android blocks cleartext `http` by default. The debug build type of the generated project already sets `usesCleartextTraffic="true"` (`android/app/src/debug/AndroidManifest.xml`, from the Expo template); release builds do not, so they stay https-only. Nothing to configure.
- `android/` is generated (CNG) and gitignored. `expo run:android` runs prebuild for you the first time.
- **Windows builds in a short-path mirror.** Ninja (the SDK's CMake, and also ninja 1.13) cannot stat the 260+ character object paths that bun's isolated `node_modules/.bun/...` produces (`ninja: error: Stat(...): Filename longer than 260 characters`; nothing in Gradle changes that). On Windows `android:dev` therefore copies only the native config (`package.json`, `app.config.ts`, `assets/`, `tsconfig.json`) to `C:\h` (override with `HEXASCORE_ANDROID_DIR`), runs `bun install --linker hoisted` there, prebuilds and builds the debug APK there with `expo run:android --no-bundler`, then starts Metro from the real checkout. JS always comes from the checkout, so editing TypeScript needs no sync; only a change to `package.json` or `app.config.ts` is picked up at the next `android:dev` run (it re-copies, and reinstalls when `package.json` changed; delete `C:\h\android` to force a new prebuild).

## One-time setup

1. Environment (PowerShell):

   ```powershell
   $env:ANDROID_HOME = "$env:LOCALAPPDATA\Android\Sdk"
   $env:PATH += ";$env:ANDROID_HOME\platform-tools;$env:ANDROID_HOME\emulator"
   ```

   Make `ANDROID_HOME` permanent in the Windows environment variables; `adb` is not on `PATH` by default. `java -version` must print 21.
2. From the repo root: `bun install --frozen-lockfile`.
3. Backend, per [setup-local-without-docker-windows.md](../setups/setup-local-without-docker-windows.md) or the Docker variant: Postgres on 5433, `.env` from `.env.example`, then:

   ```powershell
   vukem migrate
   vukem db:seed DemoCatalog
   vukem serve            # http://localhost:3000, leave it running
   ```

   The demo admin signs in as `admin@example.test` with password `Demo-password-1`. Create a normal member with the sign-up page of the local site (`/register`) if you want to test as a regular user.
4. Check the server from the host: `curl http://localhost:3000/up` returns 200.

## Every run

```powershell
emulator -avd omega-api33          # or any android-33 AVD; wait for the home screen
cd apps\mobile
bun run android:dev
```

`android:dev` is `expo run:android --variant debug` with the API URL set (and the Windows mirror above). The first build takes about 10 minutes plus the mirror install (Gradle, Kotlin, C++ for Reanimated and friends); later ones are incremental. It installs the APK, launches the app, runs `adb reverse tcp:8081 tcp:8081` and keeps Metro running in the terminal. Edit TypeScript and Fast Refresh reloads it. The first launch asks about crash reports; answer "Don't send". Sign in with the seeded `admin@example.test`.

Options:

- Another backend: `$env:EXPO_PUBLIC_API_BASE_URL = "http://10.0.2.2:3100"; bun run android:dev` (the variable is baked in at bundle time, restart Metro after changing it). A physical phone needs your LAN address instead of `10.0.2.2` and a dev server reachable on it.
- Extra Expo flags are forwarded: `bun run android:dev --device`.
- JS only after the first install: `bun run start` (the app is already installed; `adb reverse tcp:8081 tcp:8081` once per emulator boot).

Scripted checks without touching the screen:

```powershell
adb exec-out screencap -p > screen.png
adb logcat -s ReactNativeJS:V
```

## Google sign-in (debug APK only)

The button appears only when `EXPO_PUBLIC_GOOGLE_CLIENT_ID` is set at Metro start and at prebuild (it also registers the `com.googleusercontent.apps.<id>` redirect scheme in `app.config.ts`). The client must be a Google OAuth client for an installed app with the custom URI scheme enabled, and its id must be in the server's `GOOGLE_ANDROID_CLIENT_ID` (or `SOCIAL_GOOGLE_CLIENT_ID`). The flow is authorization code with PKCE and a server-issued nonce (`docs/brainstorming.md`, "Owner decisions 2026-11-07 (mobile)"). Unverified against the production client, so it is not part of release builds yet.

## Screenshots of every screen

Start Metro with `EXPO_PUBLIC_CAPTURE=1` (hides LogBox) and open screens with deep links instead of tapping: `adb shell am start -W -a android.intent.action.VIEW -d "hexascore://games/<slug>" app.hexascore`. Deep links reach item (`/{type}/{slug}`), `entities/{id}` and `lists/{id}` pages; the tabs, settings and notifications are reached by tapping (tab bar labels at y 2318 on a 1080x2400 emulator). Switch themes with `adb shell cmd uimode night yes|no` once Settings > Appearance is on System, and capture with `adb exec-out screencap -p > shot.png`. Restart Metro after adding files: its watcher can miss new files and then report `Unable to resolve module`.

## Troubleshooting

- **`Device emulator-5554 is in state offline`**: the emulator is still booting. Wait for `adb shell getprop sys.boot_completed` to print `1`.
- **`Filename longer than 260 characters` or `build.ninja still dirty`**: you are building outside the mirror (for example `expo run:android` by hand on Windows). Use `bun run android:dev`, or move the mirror closer to the drive root with `HEXASCORE_ANDROID_DIR`. The paths are ninja's relative object paths, so a shorter checkout path alone is not enough with bun's isolated layout.
- **Gradle sits idle for many minutes after `checkDebugDuplicateClasses` or similar**: Gradle 9 can deadlock on parallel artifact resolution. Stop it (`apps\mobile\android\gradlew.bat --stop`, kill the `expo` node process) and rerun with `$env:GRADLE_OPTS = "-Dorg.gradle.parallel=false"`.
- **Out of memory / the PC freezes**: close other heavy programs; the generated `android/gradle.properties` allows 2 GB for Gradle. For one ABI only Expo already passes the emulator's ABI. Add `--no-daemon --max-workers=2` through `$env:GRADLE_OPTS` if needed. The emulator itself needs about 2 GB.
- **`ANDROID_HOME` / `adb` not found**: set `ANDROID_HOME` as above and open a new terminal. `SDK location not found`: the generated `android/local.properties` is missing; rerun `bun run android:dev` or write `sdk.dir=C\:\\Users\\<you>\\AppData\\Local\\Android\\Sdk`.
- **App opens but shows network errors**: confirm `curl http://localhost:3000/up`; inside the emulator `adb shell curl -s http://10.0.2.2:3000/up` should return 200 (if `curl` is missing there, open `http://10.0.2.2:3000/up` in the emulator's Chrome). A 301 in the answer means the backend runs with `APP_ENV=production`; use development.
- **Metro "Unable to load script"**: `adb reverse tcp:8081 tcp:8081` and reload (`adb shell input keyevent 82`, Reload).
- **Stale data after switching backends**: the app caches queries in SQLite and tokens in SecureStore; clear with `adb shell pm clear app.hexascore`.
- **Expo touched tracked files** (`apps/mobile/expo-env.d.ts`, `apps/mobile/.gitignore`): Metro and prebuild rewrite them; revert with `git checkout apps/mobile/expo-env.d.ts apps/mobile/.gitignore` before committing.
- **Reset the native project**: delete `apps\mobile\android` and rerun; it is regenerated from `app.config.ts`.
