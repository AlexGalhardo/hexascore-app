# FIX-lighthouse-script-size

status: concluida
branch: agent/impl-web/FIX-lighthouse-script-size @ cb15ab94 (from main 5adf1799; local commit, not pushed)
worktree: `C:\dev\hexascore-app\wt\FIX-lighthouse-script-size`
files: resources/js/components/{Nav.tsx, NavSheet.tsx (new), Nav.test.tsx, ReportDialog.tsx, ReportDialog.test.tsx (new), DataExport.tsx, item/RatingsDialog.tsx, library/EditionPicker.tsx, search/SuggestBox.tsx, search/SearchPalette.tsx}, resources/js/lib/{queryClient.tsx, queryClient.test.tsx}, resources/js/{app.tsx, ssr.tsx}, docs/adrs/0020-web-tanstack-query.md

## What changed

1. Mobile menu sheet: moved into `NavSheet.tsx` and loaded with `React.lazy`. It mounts on the first open, or earlier on hamburger `pointerdown` or `focus`, so the chunk is usually ready when the click arrives. The hamburger itself stays eager and unchanged: 44 px, `aria-expanded`, unread label and badge. The synchronous Tab trap stays in Nav. While the chunk loads, Tab is held on the hamburger and Escape cancels the open. After the sheet arrives, Tab and Shift+Tab wrap inside it, Esc returns focus to the hamburger (`finalFocus`), and `motion-reduce` classes are unchanged.
2. TanStack Query off the first-load path. `withApp: withQueryClient` was removed from `app.tsx` and `ssr.tsx`. `withQueryClient(Component)` is now a HOC on the five consumers (DataExport, RatingsDialog, EditionPicker, SuggestBox, SearchPalette):
   - In the browser it keeps one shared client for the whole app.
   - On SSR it creates a fresh client per mount.
   - An outer provider wins, so `lib/testing.tsx` still injects the test client.

   None of the five components is on the audited first load. ADR 0020 wiring text was amended; the decision itself is unchanged.
3. `ReportDialog` loads the Base UI `Dialog` on first open. Reviews on the item page import it eagerly, so after step 1 the 18.5 KB `DialogTitle` chunk was still on `/games/*` and `/books/*`.

Checked, no change needed:

- Turnstile is only imported by Login, Signup and Contact.
- The cookie banner is about 2.5 KB of source inside `app`, so it is not worth a lazy split.
- There is one React copy.
- No date or icon library is in the bundle: icons are inline SVG paths.

What remains in the shared vendor chunk is react-dom, `@inertiajs/core` and `@inertiajs/react`, plus `es-toolkit` and `laravel-precognition`. The last two are Inertia dependencies, so they are not trimmable without config changes.

## Lighthouse (clean Linux container)

Setup:

- Container: `lhci-runner:local` (node 22.12.0 bookworm, chromium, socat, `@lhci/cli` 0.15.1), using the repo `lighthouserc.cjs`. socat forwards `localhost:<port>` to the host QA stack.
- Runs: 3 per URL, median.
- Before: main c7b1594b, the especialista's gate stack `hexascore-qa-cilocal15212827912` on :3165.
- After: cb15ab94 on its own QA stack (:3171, db `hexascore_qa_lhsize`), now stopped and dropped.
- Script `lh.sh` is in the session scratchpad. It runs `lhci collect --settings.chromeFlags='--no-sandbox --headless=new'` and then `lhci assert`.

| URL | script before | script after | LCP before | LCP after |
| --- | --- | --- | --- | --- |
| / | 170.6 KB | 134.5 KB | 1131 ms | 1130 ms |
| /games | 173.9 KB | 137.7 KB | 1733 ms | 1360 ms |
| /games/adipiscing-consectetur | 184.1 KB | 146.1 KB | 1209 ms | 1208 ms |
| /books/consectetur-2004 | 184.1 KB | 146.1 KB | 1211 ms | 1209 ms |
| /people/ana-oliveira | 170.6 KB | 134.5 KB | 1660 ms | 1131 ms |

- Budget: 150 KB (153600 B). The margin after the change is 3.9 KB on the item pages and 12-15 KB elsewhere.
- `lhci assert`: before, exit 1 (script size); after, exit 0, every assertion passes. TBT after the change is 41-91 ms and performance is 0.99-1.
- Size model: wire size ≈ gzip-6 size (Go `GzipStatic`, default level) + about 1.5 KB of response headers per script request. Files under 1 KB are not compressed. Each extra small chunk therefore costs about 2 KB.

## Validation

| command | result |
| --- | --- |
| `bun run lint` | biome, 225 files, clean |
| `bun run typecheck` | clean |
| `bun run test` | 57 files, 214 tests pass |
| `vitest` on Nav, CookieBanner and Turnstile | 22 pass |
| `vitest` on ReportDialog and queryClient | 4 pass |

New tests:

- `Nav.test.tsx` "the lazy sheet is not in the first render and the first open still traps Tab". It sits first among the sheet tests so it covers the cold chunk. Tab presses before the sheet arrives keep focus on the hamburger, then 20 Tab and 20 Shift+Tab stay inside the sheet, then Esc returns focus to the hamburger. Mutation check: dropping the loading-time `preventDefault` makes it fail.
- `ReportDialog.test.tsx`: the dialog is absent before the click, opens on the first click, and Esc returns focus to the button.
- `queryClient.test.tsx`: the browser client is a singleton, SSR gets a fresh client, and the HOC uses the app client while an outer provider wins.

Existing sheet tests now `await findByRole("dialog")`.

## Other notes

- changelog: not edited (owner rule). Suggested line: "Changed - the mobile menu, report dialog and data-fetching code load on first use; first-load script down from 171-184 KB to 135-146 KB."
- roadmap: no box.

## Risks and follow-ups

- First open of the hamburger on a cold cache waits for one small chunk. The `pointerdown` and `focus` prefetch hides this on real taps; a keyboard Enter without prior focus is not possible.
- `withQueryClient` changed shape: the `withApp` hook became a HOC. Any new component using `useQuery` must be wrapped, or it throws "No QueryClient set". Tests do not catch a missing wrap, because `lib/testing.tsx` always provides a client.
- The item-page margin is 3.9 KB. The next eager import on Catalog/Show will eat it. Cheap future wins:
  - fewer tiny chunks, about 2 KB of headers each;
  - smaller security headers on `/build/*`.

vukem-feedback status line: none

## Review (especialista, 2026-11-01)

Verdict: CORRIGIDO_E_APROVADO. Merged into local `main` as a88d9350 (branch head a3f07a24).

- Fix a3f07a24: `resources/js/lib/queryClientGuard.test.ts` statically checks every non-test file under `resources/js` that calls a TanStack Query hook. The file must call `withQueryClient(`, or every file that imports it must be safe by the same rule (this covers `useSuggest.ts` through SuggestBox and SearchPalette). Mutation checks: unwrapping DataExport fails on `DataExport.tsx`; unwrapping SuggestBox fails on `useSuggest.ts`.
- SSR: `withQueryClient` calls `appQueryClient(true)` inside `useState` on the server, so each mount gets a fresh client and the module-level `browserClient` is never set on SSR. Nothing crosses requests.
- `scripts/ci/local.sh --only build` on a88d9350: PASS. It covers the Docker image, `check-image.sh` and the `/up` check with `"ssr":"ok"`.
- a11y: Nav, ReportDialog, queryClient and the lib tests pass (7 files, 24 tests). This includes the first-open lazy trap test and the ReportDialog test where Escape returns focus to the button.
- Lighthouse in `lhci-runner:local` against a QA stack of the merged image: `lhci assert` exit 0. Medians of 3 runs:

| URL | script | LCP | TBT | perf |
| --- | --- | --- | --- | --- |
| / | 134.8 KB | 1130 ms | 75 ms | 1 |
| /games | 138.1 KB | 1732 ms | 34 ms | 0.99 |
| /games/adipiscing-consectetur | 146.5 KB | 1210 ms | 65 ms | 1 |
| /books/consectetur-2004 | 146.5 KB | 1213 ms | 122 ms | 0.99 |
| /people/ana-oliveira | 134.8 KB | 1130 ms | 41 ms | 1 |

Note: `NavSheet.tsx` imports `Icon` and `iconButton` back from `Nav.tsx`. The cycle is harmless, because Nav is already loaded when the chunk loads. The item-page margin is 3.5 KB.
