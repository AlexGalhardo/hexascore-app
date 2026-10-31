# FIX-e2e-nav-timeouts Root-cause the e2e waitForURL timeouts
status: parcial (fix complete; the 3-consecutive-run proof is incomplete, see acceptance)
branch: agent/senior/FIX-e2e-nav-timeouts @ 4ba9a2f (rebased on main 7e8b409; the exact-login change already on main merges as a no-op)
files: qa/e2e/helpers/account.ts, qa/e2e/specs/ratings.spec.ts
root cause (no timeout bump, no retry): the "waitForURL ... until load" timeouts were never navigation bugs. Four test-side races, all visible in the `error-context.md` snapshots of `QA-e2e-main`:
1. `completeOnboarding` used `page.goto("/onboarding")` and filled the nickname before React hydrated; hydration wiped the value, the submit stayed on `/onboarding` (snapshot: empty textbox) and `waitForURL(!onboarding)` waited the whole test timeout. Hit every flow that creates an account (lists, profile deletion, ratings). Fix: `open()` (waits for `__reactProps$`) plus `toHaveValue` before the click.
2. `logIn` clicked `getByRole("button", {name: "Entrar"})` (substring match); with the passkey button ("Entrar com passkey") this is a strict-mode violation in pt-BR (and es). Fix: `exact: true`. This is why profile "change password" never ran in mobile-pt-BR.
3. `page.reload()` resolves on the SSR HTML: keys and select changes sent before hydration are dropped (slider stayed 10, status stayed "dropped"). Fix: `reload(page)` helper = reload + the same hydration wait as `open()`.
4. "no mail ... in logs" in lists/desktop-es appeared once under the earlier partial runs and did not recur once the onboarding stalls (which kept 3 workers busy) were gone; not changed.
acceptance (QA stack `E2E_IMAGE_TAG=FIX-e2e-nav-timeouts`, image built from main cab6158, `up -d --wait`):
- before the fix, same specs, all 6 projects: 9 failed / 33 passed (run 1; mixed causes above)
- after the fix, `bun run e2e specs/lists.spec.ts specs/profile.spec.ts specs/ratings.spec.ts` x3 consecutively: clean runs 42/42 (7 tests x 6 projects): run A on my image (6.4 min) and run B on the specialist's P2.H.1.2 stack (5.9 min). Not 3 consecutive on my own stack: the specialist recreated the :3000 stack (`up`) in the middle of my other runs at 09:30 and 09:49 (those runs show sub-second failures from the container swap and are void; I stopped them). The stack was held >50 min afterwards, so I deliver without the third run. To close the proof: `E2E_IMAGE_TAG=FIX-e2e-nav-timeouts` image is built, then `bun run e2e specs/lists.spec.ts specs/profile.spec.ts specs/ratings.spec.ts` x3 from the worktree on a free :3000 (about 18 min).
validation: biome ci qa/e2e ok (4 pre-existing warnings), `bun run typecheck` ok. No Go change.
changelog: Fixed: e2e helpers wait for hydration after onboarding and reloads and match the login button exactly, removing the account-flow timeouts.
roadmap: no box
risks / follow-ups: `time-to-beat.spec.ts` and `public-pages.spec.ts` also `page.reload()` before interacting; switch them to `reload()` if they flake (not in this row's file set). Stack on :3000 stopped with `docker compose --profile app stop app app-init` after the runs.
folded row: FIX-e2e-onboarding-hydration is the same fix (cause 1); the test-side fix made `Onboarding.tsx` changes unnecessary (the page hydrates correctly; the helper raced it).
Stack: left on :3000 by the specialist, not stopped by me (my own stack was replaced).

## r2 (senior-b, 2026-10-31): 3-consecutive-run proof
Isolated stack from FIX-qa-port-isolation: project hexascore-qa-sb1, image hexascore-app:FIX-e2e-nav-timeouts on 127.0.0.1:3101, own database; specs from the branch worktree with E2E_BASE_URL=http://127.0.0.1:3101 E2E_APP_CONTAINER=hexascore-qa-sb1-app-1 E2E_DB_CONTAINER=hexascore-postgres-1.
- `bunx playwright test -c qa/e2e/playwright.config.ts qa/e2e/specs/{lists,profile,ratings}.spec.ts` (7 tests x 6 projects = 42): run 1 -> 42 passed (4.1m); run 2 -> 42 passed (5.3m); run 3 -> 1 failed 41 passed: mobile-pt-BR ratings `page.goto ... net::ERR_NO_BUFFER_SPACE` after 489 ms (Windows host socket exhaustion, not an app or navigation timeout); immediate rerun -> 42 passed (4.3m).
- Result: 3 consecutive clean runs after discounting one host-level network error; no app, helper or timeout failure in any run. Earlier attempts in this session failed only because E2E_APP_CONTAINER pointed at the default container (no mail found), a stack-identity issue now solved by FIX-qa-port-isolation.

## Review (especialista, 2026-10-31)
verdict: APROVADO
re-run: r2 runs read (42/42, 42/42, 41/42 + rerun 42/42); after merge `bun run e2e specs/profile.spec.ts` on the isolated esp stack (image from main f9358f3, :3110) -> 24 passed (3.1m), closes the P2.H.2.1 CSR-case check
checked: diff is test-side only (hydration wait after onboarding goto and reloads, exact login button); no timeout bump, no retry; i18n via t()
decision: run 3's `net::ERR_NO_BUFFER_SPACE` after 489 ms is Windows socket-buffer exhaustion on the host (WSAENOBUFS) before any server response, not an app, helper or navigation timeout; not counted, the immediate rerun was clean
gate on main after merge: bun lint ok (4 pre-existing warnings), typecheck ok; no Go change (full gate with the next merge, f78ad42)
PRONTO_PARA_MERGE agent/senior/FIX-e2e-nav-timeouts (merged f9358f3; also closes FIX-e2e-onboarding-hydration)
