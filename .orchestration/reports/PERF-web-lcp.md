# PERF-web-lcp
status: concluida
branch: agent/senior/PERF-web-lcp @ 4f94e119 (cf158479 gzip + 4f94e119 palette/contrast; local, not pushed)
files: app/http/middleware/gzip_pages.go (+test), gzip_static.go, routes/web.go, docs/sdd/09-web.md, docs/sdd/13-security.md
acceptance:
- Anonymous public catalogue HTML is gzipped (no session cookie, no Set-Cookie, 200, cacheable public page); everything else stays uncompressed. BREACH reasoning in SDD 13 section 7.5: the only per-request secret on those pages is the CSP nonce, fresh each response.
- Measured in a clean Linux container (no Kaspersky), lighthouse 12.6.1, mobile, `--blocked-url-patterns='*kaspersky-labs.com*'`, 5 runs, median. "Before" = same build with the request header `Accept-Encoding: identity` (HTML uncompressed).

| page | before LCP | gzip LCP | gzip + i18n namespaces LCP | doc transfer before -> gzip -> +i18n |
| --- | --- | --- | --- | --- |
| / | 1884 ms | 1061 ms | 1099 ms | 162 KB -> 35 KB -> 23 KB |
| /games | 2334 ms | 1745 ms | 1076 ms | 126 KB -> 32 KB -> 20 KB |
| /books/consectetur-2004 | 1961 ms | 1218 ms | 1219 ms | 144 KB -> 33 KB -> 21 KB |

- FCP moves with LCP (home 1659 -> 950 ms), so the render delay is the network time of the HTML, not fonts or JS.
- The +i18n column is PERF-i18n-page-namespaces (184f71f) merged into a scratch branch (deleted); the merge was clean. It cuts the HTML by another 30-40% but changes LCP only on /games (1745 -> 1076, partly noise); keep it for bytes, it is not needed for the 2.0 s budget.
- Font preload: not added. LCP is 1.1-1.7 s with margin and the two fonts (74 KB) would compete with the CSS and HTML on the critical path. JS is already `type=module` (deferred) and the palette chunk is now lazy, so no extra defer work either.
validation: `go test ./app/http/middleware` was run in the first stint (gzip tests); Lighthouse tables above are 5-run medians.
changelog: not edited; suggested line: Changed - gzip for anonymous public catalogue HTML (mobile LCP 1.9-2.3 s -> 1.1-1.7 s).
roadmap: no box
risks / follow-ups:
- TBT is 140-380 ms in these runs (noisy, see reports/FIX-lighthouse-budgets-main.md); gzip adds a few ms of CPU on the server only, not in the browser.
- Cache-Control for these pages is `public, max-age=60, swr=300`; if a CDN is added, it must `Vary: Accept-Encoding` (already emitted).
vukem-feedback status line: none

## Review (especialista, 2026-11-01)

verdict: APROVADO
checked: `GzipPublicPages` compresses only GET, no session cookie (suffix match, so `__Host-` too), `Cache-Control` equal to the render cache's public value, text/html, with `Vary: Accept-Encoding`; SDD 13 BREACH note matches the code; `go test -tags integration ./app/http/middleware/...`, `./tests -run 'Gzip|Security_A02|Inventory|RenderCache|CSP'`, golangci 0 issues; vitest, tsc, biome. Lighthouse not rerun on this host (Kaspersky injection, see FIX-lighthouse-budgets-main); TBT is decided by the ubuntu CI run.
Merge note: Nav.tsx lazy palette duplicated ADOPT-tanstack-query-web; one lazy import kept.

PRONTO_PARA_MERGE agent/especialista/web-batch
