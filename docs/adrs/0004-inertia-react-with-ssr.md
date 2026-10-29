# ADR 0004: Inertia v3 + React with mandatory server-side rendering for public pages

Status: accepted (2026-10-28). Decisions: brainstorming D-04, D-07. Architect recommendation was HTMX + `html/template`; the owner chose Inertia and accepted the SSR condition.

## Context

vukem offers three rendering paths (`https://vukem.dev/docs/frontend/htmx`): `html/template` or templ views with HTMX v4 (default starter, no Node in production), and an Inertia v3 adapter (BETA) with React, Vue and Svelte starters, Vite integration, and SSR through a Node process (`vukem ssr:start`, POST `/render` to `127.0.0.1:13714`, 3 s timeout) that **falls back to client-side rendering on any failure**.

HexaScore is a catalogue whose growth depends on search engines indexing item, entity and listing pages, and on Core Web Vitals (the owner's performance references are PageSpeed reports). The legacy app is React, and the owner wants to keep React knowledge and component patterns and get rich client interactivity (star widget, dialogs, infinite lists).

## Decision

Use Inertia v3 with React (TypeScript, Vite) for the whole web app, with these binding conditions:

1. Every public page (home, listings, item, edition, entity, search, user profile and lists, docs, legal) is rendered by the SSR process. The SSR process runs in the same container as the Go app (D-08), is started and supervised by the app, and is a readiness dependency: `/up?checks=1` reports it.
2. CSR fallback is treated as an incident: a counter and an alert (SDD 12) fire when `OnFailure` is invoked for a public route; the error budget is 0.1 % of public page renders per day.
3. Core Web Vitals budgets are measured on the SSR HTML (SDD 09): LCP ≤ 2.0 s on mobile emulation, INP ≤ 150 ms, CLS ≤ 0.05, total JS ≤ 150 kB gzip on item pages.
4. Node is used only for SSR and the Vite build. No application logic, no data access, no API calls from Node except the Inertia render.
5. Logged-in-only pages (profile settings, admin) may skip SSR via `inertia.DisableSSR` to save CPU.
6. Inertia-specific features from the vukem docs that reduce payload are mandatory where they apply: partial reloads, `Defer` for below-the-fold props, `Scroll`/`ScrollCursor` for infinite lists, `Once` for static shared props.

## Alternatives considered

- **HTMX + `html/template` (architect's recommendation).** Smallest JS, best CWV, no Node, matches vukem's default and its security tests (`TestSecurity_A02_StarterViewsRespectCSP`). Costs: React components not reused; interactive widgets as Vite islands. Rejected by the owner.
- **Hybrid: HTMX for public pages, Inertia for the logged-in app.** Best of both for SEO and interactivity, but two view models in one app and two component libraries. Offered in Round 2; rejected by the owner in favour of a single model.
- **Inertia without SSR.** Unacceptable for a catalogue: slower indexing, worse LCP. Rejected.

## Consequences

- Positive: one component model across pages; React skills reused; design tokens shared with the React Native app (ADR 0009) as plain TypeScript constants.
- Negative: a Node process in production; larger JS than HTMX; Inertia is BETA in vukem so adapter changes are expected (watch `vukem/CHANGELOG.md`); SSR CPU competes with the Go app in the single container (D-08), so render caching of public pages (ETag + short TTL in the Go layer) is required.
- Obligations: `vukem-feedback.md` entries for any Inertia adapter gap; SDD 09 defines the page/prop contracts and the CWV test (Lighthouse CI in the QA gate); SDD 12 defines the SSR fallback alert.
