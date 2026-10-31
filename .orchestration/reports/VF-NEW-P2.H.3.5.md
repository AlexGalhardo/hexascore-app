# VF-NEW-P2.H.3.5 SSR output cache with per-request CSP nonce
status: concluida (no framework gap; docs clarification)
branch: fw: fix/vf-ssr-cache-docs @ 8e406128 (not pushed, no PR; worktree wt/fw-VF-NEW-P2.H.3.5)
files: site/content/docs/frontend/inertia.mdx, site/content/last-updated.json
design note: vukem already ships exactly the requested feature since v0.33.0: inertia.SSRConfig{Cache, CacheTTL, CacheWhen} (https://vukem.dev/docs/frontend/inertia "Caching renders"). It caches the Node render (head+body) keyed by SHA-256 of the page object; the CSP nonce is added by the root view per request, so it is nonce-safe and bypasses via CacheWhen (e.g. anonymous only). No new API needed, so no QUESTIONS.md entry. The entry was filed against the wrong page (frontend/ssr). Residual cause of HexaScore misses: any per-request shared prop (CSRF token, flash, request id) changes the key. Docs now say so.
acceptance:
- docs-only; `task ci` -> EXIT=0
changelog: Changed (docs) - note that per-request shared props defeat the SSR render cache.
risks / follow-ups: HexaScore should adopt SSRConfig.Cache + CacheWhen and drop its RenderCache middleware (verify shared props are stable for anonymous visits), re-run the k6 TTFB budget.
vukem-feedback status line: VF-NEW-P2.H.3.5 -> withdrawn: already provided by inertia.SSRConfig.Cache (v0.33.0); docs clarified in branch fix/vf-ssr-cache-docs.
