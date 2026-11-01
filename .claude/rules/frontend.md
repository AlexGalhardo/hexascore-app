---
paths:
  - "resources/**"
  - "e2e/**"
---

# Frontend rules

- React 19 + TypeScript (strict) pages under `resources/js/pages`, rendered by Inertia v3 with SSR (`resources/js/ssr.tsx`). Props shapes live in `resources/js/types.ts` and mirror the Go controllers in `app/controllers`.
- Base UI (`@base-ui/react`, approved in D-22, enters with P1.G.1.1) for behaviour of accessible primitives (menu, dialog, select, switch, tabs, toast); styling is Tailwind v4 only, tokens in `@theme` in `resources/css/app.css`. No shadcn, no other component kit.
- Visual direction follows the HexaScore design notes in `docs/sdd/09-web.md`. No third-party logos, names or copied assets.
- Every string comes from `resources/lang/<locale>.json` (`en`, `pt-BR`, `es`) through `resources/js/lib/i18n.ts`; the three locales stay in sync (a test checks the key sets). Format money/dates with `Intl.*` and the active locale.
- Dark/light: `.dark` class on `<html>`, set before paint by the nonce'd inline script in `app.html`; toggle persists in `localStorage` (`theme`).
- Accessibility: visible labels, `aria-label` on icon-only buttons, visible focus, `prefers-reduced-motion` respected. No `dangerouslySetInnerHTML` with user content.
- Validate form input client-side for UX only; the server (vukem validation) is the source of truth.
- Biome (`bun run lint`) and `bun run typecheck` must pass. Load `frontend-design` before UI work and audit with `web-design-guidelines`.
- Data fetching: TanStack Query is not on the first load. Wrap every component that calls a TanStack hook with `withQueryClient` (`resources/js/lib/queryClient.tsx`); `queryClientGuard.test.ts` fails otherwise. SSR creates a fresh client per mount.
- Budgets (Q-A20): script ≤ 170 KB and LCP ≤ 2500 ms per audited page; item pages have the least headroom. Lazy-load dialogs, sheets and anything opened on interaction (`React.lazy`, mount on first open or on intent); check `vite build` chunk sizes before handing off.
- Shared building blocks: `catalog/ItemTile` (grid card: title and year 2/3, full-height score 1/3), `CategoryIcon` (before every category name), `DocumentPage` (reading-width pages: terms, privacy, how-score-works, about), the `iconButton` and button press/shadow tokens in `app.css`. Reuse them instead of new variants.
- Carousels and anything that moves on its own need a pause control, stop on hover and focus, no autoplay under `prefers-reduced-motion`, and hidden inactive slides (WCAG 2.2.2).
- Cached anonymous pages must render the same HTML for everyone: per-visit randomness (for example the landing's recommended grid) is picked on the client after hydration, from a cached pool, in a fixed-size layout.
- Owner exceptions: the item page disables mouse text selection (`select-none`, inputs stay selectable); help text removed from view stays for assistive tech (`sr-only` plus `aria-describedby`).
