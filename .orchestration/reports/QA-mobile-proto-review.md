# QA-mobile-proto-review Android prototype vs P5.A.1–P5.A.2 acceptance
status: concluida (read-only review; no code changed)
branch: none (report only); main @ 13a6e08; prototype commits 5d0c287, 3d1439f, 873677b (`git log -- apps/mobile`)
files: .orchestration/reports/QA-mobile-proto-review.md
acceptance:
- `bun run --cwd apps/mobile test` → "Test Suites: 2 failed, 4 passed, 6 total", "Tests: 2 failed, 39 passed, 41 total", EXIT=1
  - FAIL `src/i18n/keys.test.ts` "committed catalogues are the build-time copy"
  - FAIL `src/navigation.test.tsx` "each SDD 10 route renders its placeholder": "Exceeded timeout of 5000 ms for a test"
- `grep -rn "fetch(" apps/mobile/src apps/mobile/app` → no hits outside tests; `apps/mobile/src/api/` → does not exist
- route files listed in ROADMAP P5.A.1.2 (`app/(auth)/login.tsx`, `app/(tabs)/{_layout,index,[type],search,lists,profile}.tsx`, `app/[type]/[slug].tsx`, `app/entities/[id].tsx`, `app/lists/[id].tsx`, `app/update-required.tsx`, `src/platform/{network,store}.ts`, `src/navigation.test.tsx`) → all present (`find apps/mobile/app apps/mobile/src -type f`)
validation: lint not run; typecheck not run; tests: mobile FAIL (2 of 41); browser n/a
changelog: none
roadmap: no box (P5.A boxes stay open per docs/PROGRESS.md "Execution: 2026-10-30 Android shell")
risks / follow-ups:

## Prototype vs acceptance

| Box | Acceptance (ROADMAP) | State in prototype |
| --- | --- | --- |
| P5.A.1.1 | typecheck + lint 0 errors; `fetch(` in `src/screens/` lint error; test asserts `global.HermesInternal` | Scaffold present (`package.json`: expo 58.0.21, expo-router 58.0.28, nativewind 4.2.7, react-native 0.88.1, jest-expo 58.0.13, @testing-library/react-native 14.1.0); `src/hermes.test.ts` present; `no-fetch-outside-api.grit` present. Not run here: lint/typecheck. Lint rule targets `src/api/`, which does not exist yet. |
| P5.A.1.2 | `test -- src/navigation` renders each §6.1 placeholder; `[type]` rejects non games/movies/series/books; typed href errors on unknown route | Every file in the Files list exists. Test runs and times out on the placeholder test (5 s). Route-rejection assertion not confirmed by this review. |
| P5.A.2.1 | `test -- src/theme src/i18n`: score bands, locale matcher (`pt-PT→pt-BR`, `es-MX→es`, `fr→en`), key parity; `test -- src/theme/contrast` ≥4.5:1 | `src/theme/{index,scoreBand}.ts`, `src/i18n/{locale,format,index}.ts`, `copy-catalogues.ts`, tests `scoreBand.test.ts`, `locale.test.ts`, `keys.test.ts`, `contrast.test.ts` present. `keys.test.ts` "committed catalogues are the build-time copy" FAILS: the copied catalogues in `src/i18n/catalogues/*.json` (407 flat keys each) differ from `resources/lang/*.json` in 12 keys per locale (`lists.quota.*` in my quick check). Probably the copy was not regenerated after the resources/lang edits; confirm with `bun run --cwd apps/mobile catalogues` then `git diff`, not run here. |
| P5.A.2.2 | (box exists in ROADMAP after P5.A.2.1; not reviewed here) | not reviewed |
| P5.A.3.1 | `src/api/` client, RFC 9457 mapping, 426/`X-App-Min-Version` | Not started: no `src/api/`, no generated client use (`packages/app-client` not referenced). |

## Other observations

- No `fetch(` anywhere outside tests, so the Biome/grit ban is not yet exercised by any code path.
- Placeholder screens: 10 files under `app/` import `Placeholder` (`src/ui/Placeholder.tsx`), consistent with "placeholder screens" in P5.A.1.2.
- Navigation test is slow (64.5 s suite time, cold transform); the 5 s per-test timeout is the failing point.
- Deliberately out of scope for this review: running lint/typecheck, fixing anything, editing catalogues.

## Suggested next rows (for the orchestrator, not created)

1. Senior, prio 2: regenerate the mobile catalogue copy and fix the navigation test timeout (jest timeout or lighter render) so `bun run --cwd apps/mobile test` is green before P5.A.2.2.
2. Senior, prio 2: P5.A.3.1 after P5.A.2.1 is green (ROADMAP "Depends on: P5.A.2.1"); its box text is in ROADMAP at P5.A.3.1.

## Review (especialista, 2026-10-31)
verdict: APROVADO (report only)
checked: both failures are confirmed by later deliveries: the catalogue drift (FIX-mobile-catalogues c1b7990, P5.A.1.1 copy script now emits tabs) and the 5 s navigation budget (P5.A.1.2 split per route, FIX-mobile-nav-test-timeout); the "src/api missing" gap is P5.A.3.1. Suggested rows 1 and 2 are already on the board as those rows.
