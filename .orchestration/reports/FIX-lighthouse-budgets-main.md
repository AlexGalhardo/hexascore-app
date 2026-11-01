# FIX-lighthouse-budgets-main
status: concluida (TBT still above budget, see risks)
branch: agent/senior/PERF-web-lcp @ 4f94e119 (shared with PERF-web-lcp; local commit, not pushed)
files: resources/js/components/Nav.tsx, packages/tokens/{tokens.json,tokens.css,tokens.ts}
acceptance:
- Root cause of the original failure (perf 0.62, best-practices 0.75, LCP 7.8 s, script 375 KB, stylesheet 230 KB): the Kaspersky Internet Security injection on the Windows host. The 3 audits show `gc.kis.v2.scr.kaspersky-labs.com` main.js (212 KB), abn/main.css (217 KB) and a 448 KB XHR, plus CSP console errors from its inline style and the is-on-https audit. Kaspersky also strips Accept-Encoding, so the HTML arrived uncompressed (161 KB) on host runs.
- Not a code problem: static assets are gzipped (app css 55 -> 11 KB, app js 34 -> 11 KB), CSS is already small, and `*.test.tsx` pages do NOT end up in the page graph (home loads app, runtime, jsx-runtime, route, DialogTitle and 5 small chunks).
- Clean measurement: `lhci autorun` (0.15.1, lighthouse 12.6.1) from a Linux container (node:22.12.0-bookworm + chromium) against the QA stack on :3150 via socat so the canonical host `localhost:3150` resolves. Before the fixes, that run failed on: script size (162-168 KB vs 150 KB), home accessibility 0.96 (color-contrast), item performance 0.88, TBT.
- Fixes: (1) `SearchPalette` is now `lazy` and mounted on first open, which drops the 20 KB base-ui chunk (`DialogTitle`) from every page (script size now passes everywhere); (2) light `--hs-accent-book` oklch 0.529 -> 0.5 (contrast 4.27 -> about 4.8 on the #f1ece0 surface; `.text-xs text-accent-book` on home); tokens regenerated with `bun run build.ts`, `bun test packages/tokens` 34 pass.
- After: every assertion in `lighthouserc.cjs` passes except total-blocking-time (see risks). LCP 1.1-1.7 s, perf >= 0.9, a11y 1, SEO 1, best-practices >= 0.95, CLS, server-response-time, font count/size all pass.
validation: Nav.test.tsx 3 pass; tsc --noEmit clean; two full lhci runs after the fix, both only TBT failing.
changelog: not edited (owner rule); suggested line: Changed - the search palette loads on first open; light book accent darkened for contrast.
roadmap: no box
risks / follow-ups:
- TBT median per page was 170-470 ms across 4 lhci runs (budget 200 ms), with huge run-to-run spread (95 to 725 ms on the same page) on this shared 16-core host with several workers running. Observed main-thread tasks are 60-110 ms (hydration plus style/layout of a 125-160 KB DOM), so the code is borderline, not far off. I could not tell load noise from real regression on this host. Decide on CI's result: if the ubuntu runner also exceeds 200 ms, cut hydration cost (smaller home DOM / fewer sections above the fold, or relax the budget to 300 ms).
- Host Lighthouse runs on this Windows machine are invalid while Kaspersky injects; use `--blocked-url-patterns='*kaspersky-labs.com*'` (does not fix the stripped Accept-Encoding) or the container recipe above. `scripts/ci/local.sh --only lighthouse` will keep failing on this host for that reason; consider running it in the container.
- The `127.0.0.1` base used by local.sh gets a 301 to `localhost:<port>` (APP_URL canonical host); it adds one redirect to every measured run.
vukem-feedback status line: none

## Review (especialista, 2026-11-01)

verdict: APROVADO
checked: `GzipPublicPages` compresses only GET, no session cookie (suffix match, so `__Host-` too), `Cache-Control` equal to the render cache's public value, text/html, with `Vary: Accept-Encoding`; SDD 13 BREACH note matches the code; `go test -tags integration ./app/http/middleware/...`, `./tests -run 'Gzip|Security_A02|Inventory|RenderCache|CSP'`, golangci 0 issues; vitest, tsc, biome. Lighthouse not rerun on this host (Kaspersky injection, see FIX-lighthouse-budgets-main); TBT is decided by the ubuntu CI run.
Merge note: Nav.tsx lazy palette duplicated ADOPT-tanstack-query-web; one lazy import kept.

PRONTO_PARA_MERGE agent/especialista/web-batch
