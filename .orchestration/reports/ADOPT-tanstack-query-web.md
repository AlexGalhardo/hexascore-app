# ADOPT-tanstack-query-web TanStack Query for client-fetched web data

status: concluida
branch: agent/senior/ADOPT-tanstack-query-web @ 6d7fb2b (2 commits on `main` @ 2bea45d; worktree `wt/impl-web-react-query`)
files: `package.json`, `bun.lock` (`@tanstack/react-query` 5.105.0 exact, same as apps/mobile), `docs/adrs/0020-web-tanstack-query.md` + index, `docs/sdd/09-web.md` §6.1 (one line), `resources/js/lib/queryClient.tsx` (+test), `resources/js/lib/testing.tsx` (render inside a fresh client, retries off), `app.tsx`/`ssr.tsx` (`withApp: withQueryClient`), `components/search/useSuggest.ts` (+test), `components/item/RatingsDialog.tsx` (+test), `components/library/EditionPicker.tsx`, `components/DataExport.tsx`, `components/Nav.tsx` (lazy palette), 7 existing tests switched to `lib/testing` render
acceptance:
- red run (tests written first): `bunx vitest run <search, item, EditionPicker, DataExport, queryClient>` → 3 failed / 24 passed (suggest cache, ratings URL without the `r=` cache-buster, DataExport success test hit by a missing `cleanup`)
- green: `bun run test` → 37 files, 143 tests pass, then Nav lazy-palette test added (components+layouts: 17 files / 52 tests pass)
- `bun run lint` → 190 files clean; `bun run typecheck` → ok; `bunx vite build` ok; `bun run build:ssr` ok
- lifecycle: `withApp` keeps one module-level client in the browser (mounted once, survives visits); Inertia calls `withApp` per SSR render (`@inertiajs/react` dist, ssr branch) and gets `makeQueryClient()` each time; no server prefetch. `queryClient.test.tsx` asserts same client in browser, distinct per SSR call, and the defaults (staleTime 30 s, retry 1, refetchOnWindowFocus false)
- CSRF: there is no app-sent CSRF header; protection is Fetch Metadata (`CrossOriginProtection`, SDD 13 §7.5) and the vukem token fallback (`X-CSRF-Token`) is never exposed to pages. The export mutation keeps the old request (relative URL, `credentials: "same-origin"`, `Accept: application/json`); `DataExport.test.tsx` asserts those options (a guard test: it was green before the migration too)
- bundle (first-load JS = app entry + static imports, gzip -9 of the built files): `main` 134.41 KB → branch 121.76 KB (−12.65). The QueryClient core is about +7 KB eager (`query`, `mutation`, `removable` chunks + app delta); the Ctrl+K palette (Base UI dialog, useSuggest, useQuery bindings) now loads on first open via `React.lazy` in `Nav.tsx`
QA (`hexascore-qa-implweb` :3140, image rebuilt from the branch): `/`, `/games`, `/games/tempor-velit-quis`, `/search?q=zel` → 200 with SSR markup; served `app-*.js` imports `SearchPalette-*.js` lazily; Ctrl+K opens the palette with focus in the combobox, typing "temp" sends one `GET /search/suggest?q=temp` and lists 8 suggestions + see-all; ratings dialog loads `?page=1`, Next loads `?page=2`, close + reopen re-renders from cache with no request (screenshot `C:/dev/hexascore-app/.playwright-mcp/adopt-tanstack-ratings-dialog.png`). Console: only the Kaspersky CSP noise. Edition picker and data export (login required) are covered by vitest only
validation: lint ok, typecheck ok, tests ok (vitest), markdownlint ok on the ADR + index (the only reported issues are in the untracked questions-to-dev.md)
changelog: Changed: client-fetched web data (search suggestions, user-ratings dialog, edition picker, data export) uses TanStack Query 5.105.0 (ADR 0020); the Ctrl+K palette loads on first open
roadmap: no box
risks / follow-ups: (1) `resources/js/app.tsx` is in senior-up's FIX-lighthouse-budgets-main lock; my change there is one import + `withApp: withQueryClient` (trivial merge). (2) Behaviour deltas: ratings and editions get one silent retry (default retry 1) before the error shows; suggest keeps `retry: false`; a failed ratings page no longer keeps the previous page's rows under the error. (3) First Ctrl+K now waits for the palette chunk (local: imperceptible). (4) `lib/testing.tsx` gets a chunk in `vite build` only because test files are still in the page glob (senior-up's FIX row removes that). (5) Auth forms and `lib/challenge.ts` left on their flows (Inertia submits / proof-of-work), per the brief.

## Review (especialista, 2026-11-01)

verdict: APROVADO
checked: pinned `@tanstack/react-query` 5.105.0 (owner approved, ADR 0020 keeps number 0020; P7.A.1.1's social ADR is renumbered 0021); SSR gets a fresh client per render, browser one per app; cached queries are public data only (suggest, item ratings, editions); data export is a mutation, so no per-user data survives a logout in the shared browser client; app.tsx merges cleanly. `bun install --frozen-lockfile`, typecheck, biome ci, vitest 43 files / 160 tests, `vite build`, `build:ssr`.
Merge note: Nav.tsx lazy palette conflicted with PERF-web-lcp; kept this branch's single lazy import.
Batch fix 715afd11: `scripts/ci/sdk-job.test.ts` and `sdk-release.test.ts` import bun:test and broke `bun run test` on main since P3.B; excluded from Vitest and run with `bun test` in ci.yml and local.sh.

PRONTO_PARA_MERGE agent/especialista/web-batch
