# FIX-mobile-login-test-timeout Preload RN components outside the test budget
status: concluida
branch: agent/senior/FIX-mobile-login-test-timeout @ f686998
files: apps/mobile/jest.setup.ts
root cause: react-native exposes View/Text through lazy getters; the first render of a test file (the (auth)/login case) transformed and loaded them inside the 5 s test timer. Measured with --no-cache: first render 6551 ms (warm cache: 567 ms).
fix: jest.setup.ts touches Text and View at setup time (no test timer covers setup). Same measurement afterwards: first render 53 ms. No timeout raised.
acceptance:
- `bun run test --no-cache` (apps/mobile) -> 7 suites, 64 tests passed
validation: lint ok (biome ci), typecheck ok, tests mobile ok (cold cache)
changelog: Fixed: mobile navigation test no longer times out on a cold Jest cache.
roadmap: no box
risks / follow-ups: other components (Pressable, FlatList...) added by later screens may need the same touch if they hit cold-cache timing; extend the list in jest.setup.ts.

## Review (especialista, 2026-10-31)
verdict: APROVADO
re-run: `bun run test --no-cache` (apps/mobile) 7 suites / 64 tests; `bun run lint` (biome ci) clean
checked: root cause is react-native's lazy getters loading inside the first test's budget; preloading in jest.setup.ts, no timeout bump
PRONTO_PARA_MERGE agent/senior/FIX-mobile-login-test-timeout (merged 6d5708a)
