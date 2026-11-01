# FIX-mobile-release-bundle

Branch `fix/mobile-release-bundle` (from local `main` @ 0caee8d1). Worktree `C:\dev\hexascore-app\wt\FIX-mobile-release-bundle`.

## Result

`apps/mobile` now produces a release JS bundle from a clean `bun install --frozen-lockfile` with no `NODE_PATH` and no new dependency. Fresh worktree, no junctions: `bun run bundle:check` exit 0 ("Android Bundled ... 1849 modules", 29 assets).

## Causes and fixes

1. No `main`: added `"main": "expo-router/entry"` to `apps/mobile/package.json`.
2. `disableHierarchicalLookup = true` removed from `metro.config.js`. It was added in the scaffold commit 5d0c2872 with no recorded reason; `nodeModulesPaths` and `watchFolders` stay (monorepo root). Jest does not use Metro and passes.
3. Babel plugin, config only: `@babel/plugin-transform-react-jsx` is already installed transitively (via `babel-preset-expo`). `babel.config.js` now wraps the `nativewind/babel` preset and resolves its string plugin names (`@babel/plugin-transform-react-jsx`, `react-native-reanimated/plugin`) from the app dir and from `babel-preset-expo`. No dependency to declare, so no owner item and no `NODE_PATH` workaround is needed.
4. Fourth cause found while verifying: with hierarchical lookup on, Metro still could not resolve `react-native-css-interop/jsx-runtime` (injected into every app file by `jsxImportSource: "nativewind"`; the package is linked only next to nativewind). `metro.config.js` wraps the resolver to map that prefix to the package dir found from nativewind. `extraNodeModules` was tried first and did not work.

## CI check

`bun run --cwd apps/mobile bundle:check` (`expo export:embed --platform android --dev false --entry-file node_modules/expo-router/entry.js`, output in git-ignored `.bundle-check/`). Added as a step in the `mobile` job of `.github/workflows/ci.yml` and in `job_mobile` of `scripts/ci/local.sh`. Note: `expo export:embed` can exit 0 while printing a resolution error in some runs; the success run prints "Done writing bundle output". Not hardened further.

## Verification

- `bun run bundle:check`: pass (fresh worktree).
- `bun run lint`: pass. `bun run typecheck`: pass.
- `bun run test -- --runInBand --forceExit`: 56 suites, 373 tests pass.

## Not done

- Not run on Linux CI (host is Windows); the first CI run will confirm. `CI=1` was set locally.
- No EAS build run. CHANGELOG/ROADMAP/PROGRESS untouched per brief.
