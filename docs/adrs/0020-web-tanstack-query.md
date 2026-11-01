# ADR 0020: TanStack Query for client-fetched data on the web

Status: accepted (2026-11-01). Decision: owner, in chat on 2026-11-01 ("se sim, aplique", asked whether react-query works with Inertia and React on the web). Row `ADOPT-tanstack-query-web`. SDD 09.

## Context

Inertia gives every page its data as server props, and those need no client cache. A few components fetch JSON by themselves after the page has loaded: the Ctrl+K palette and suggest box (`/search/suggest`), the user-ratings dialog (`/api/items/{id}/ratings`), the library edition picker (`/items/{id}/editions`) and the data-export button (`POST /profile/export`). Each one had its own `fetch` with hand-written abort, stale-answer, loading and error state. The Android app already uses `@tanstack/react-query` 5.105.0 (ADR 0019).

## Decision

- Add `@tanstack/react-query` **5.105.0** to the root `package.json`, pinned exactly. This is the same version as `apps/mobile`.
- **Scope.** Only data the client fetches by itself goes through `useQuery`, `useInfiniteQuery` or `useMutation`. Inertia page props, partial reloads, deferred props and Inertia form submits stay on Inertia. Proof-of-work (`lib/challenge.ts`) and the auth forms keep their current flow.
- **Client lifecycle** (`resources/js/lib/queryClient.tsx`; since FIX-lighthouse-script-size each data-fetching component is wrapped in `withQueryClient(Component)` instead of Inertia's `withApp`, so the query core loads only with those components):
  - In the browser, one `QueryClient` lives for the whole app and survives Inertia page visits.
  - SSR creates a fresh client for every render, so no cached data crosses requests. Nothing is prefetched on the server.
  - Defaults: `staleTime` 30 s, `retry` 1, `refetchOnWindowFocus` off. The suggest query sets `retry: false`, because the next keystroke is the retry.
- **CSRF.** Mutations send the same request the old `fetch` sent: a relative URL, `credentials: "same-origin"` and `Accept: application/json`. CSRF protection stays Fetch Metadata (`http.CrossOriginProtection`, SDD 13 §7.5). A test asserts the mutation's request options.

## Consequences

- The four components lose their custom abort and sequence code. Requests are keyed and cancelled through the query `signal`, and repeated suggest queries are served from the cache for 30 s.
- Bundle: the `QueryClient` core sits on the first-load path (about 7 KB gzip, because the provider wraps the app). To pay for it, the Ctrl+K palette and the query hooks it uses now load on its first open (`React.lazy` in `Nav.tsx`). First-load JS went from 134.4 KB to 121.8 KB gzip. The numbers are in `reports/ADOPT-tanstack-query-web.md`. Later the provider moved onto the five consumers (DataExport, RatingsDialog, EditionPicker, SuggestBox, SearchPalette), which took the query core off the first-load path entirely (`reports/FIX-lighthouse-script-size.md`).
- Tests render through `resources/js/lib/testing.tsx`, which wraps `render` in a fresh client with retries off.

## Alternatives considered

- **Keep hand-written `fetch` + `useEffect`.** No dependency, but every component repeats the abort, stale-answer and loading logic, and nothing is cached or deduplicated.
- **SWR.** Similar size, but it is a second data library next to the one the mobile app already uses, and its mutation and infinite-list APIs are weaker.
- **Inertia partial reloads / `useHttp`.** Inertia v3's `useHttp` handles one request but has no shared cache or deduplication. Partial reloads are visits that change page props, which is wrong for a palette or dialog fetch, and the 429 problem+json answer of the export would open Inertia's invalid-response modal.
