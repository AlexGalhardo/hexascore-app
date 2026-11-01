# IMPL-mobile-v1-deps v1 fonts, icons and @expo/ui in the mobile app

status: concluida
branch: agent/senior/IMPL-mobile-v1-deps @ 078aa9e (stacked on agent/senior/IMPL-mobile-v1-ui @ 3519387)
files: `apps/mobile/package.json`, `bun.lock`, `apps/mobile/app.config.ts`, `apps/mobile/assets/fonts/` (Inter 400/600/700 + Playfair Display 700 TTF, two OFL files), `apps/mobile/tailwind.config.js`, `apps/mobile/jest.config.ts`, `apps/mobile/src/test/expo-font.ts`, `apps/mobile/src/theme/index.ts`, `apps/mobile/src/ui/` (new `Icon`, `TabIcon`, `icons.test.tsx`), `docs/adrs/0019-mobile-runtime-dependencies.md` (amendment 1)
acceptance:
- `bunx expo install --check` → flags only pre-existing expo-router/react/typescript drift; none of expo-font 58.0.11, @expo/vector-icons 15.1.1, @expo/ui 58.0.26 (bundledNativeModules: ~58.0.11, ^15.0.2, ~58.0.25)
- `bunx expo config --type introspect` → expo-font plugin resolves both families, no error
- `bun run --cwd apps/mobile test` → 19 suites, 133 tests passed (new: plugin weights + files + licences, tailwind families = theme, star glyphs full/half/outline, tab icon outline/filled + tint, moon/sun icon)
validation: lint ok (biome ci), typecheck ok, tests ok; no device run (no emulator in this session); markdownlint ADR 0 issues
changelog: Added: v1 fonts (Inter, Playfair Display, OFL, embedded via expo-font) and Ionicons glyphs in the Android app; @expo/ui for native sheets and pickers (ADR 0019 amendment 1).
roadmap: no box
risks / follow-ups:
- Fonts load through the expo-font config plugin (build-time embed), not `useFonts` in `app/_layout.tsx`: no runtime wait and the locked root layout stays untouched. They need a dev build / EAS build; Expo Go shows system fonts.
- Playfair Display 700 TTF comes from Google Fonts (upstream repo ships variable fonts only); Inter from the rsms/inter 4.1 release. Web uses WOFF2 of the same families.
- `jest.config.ts` maps `expo-font` to a stub so vector-icons render without the asset registry (jest.setup.ts is locked by P5.A.4.1).
- `@expo/ui` installed, not used yet (screen rows).

## Review (especialista, 2026-10-31)

verdict: CORRIGIDO_E_APROVADO
re-run: `bun install --frozen-lockfile` ok; mobile lint ok, typecheck ok, test 19 suites / 133 passed; `bunx expo config --type introspect` ok; markdownlint ADR 0019 clean; `bun audit`: the four advisories (deepmerge, glob, minimatch, zod) all come through expo/jest, none through the three new packages
checked: Q-B3 approval, exact pins inside the SDK 58 bundled ranges, no pre-releases, ADR 0019 amendment 1, OFL licence files next to the four static TTFs (valid sfnt headers), two families (design rule), Icon decorative (hidden from TalkBack), scope, tests
fixes: 1. merged main into the branch and kept the half-star alignment fix from IMPL-mobile-v1-ui (2010593) with the new Ionicons star (faf09bc)
notes: fonts need a dev/EAS build (Expo Go shows system fonts); no device run
gate on main after merge (2e591a2): mobile tsc ok, jest 133 passed
PRONTO_PARA_MERGE agent/especialista/IMPL-mobile-v1-deps-fix
