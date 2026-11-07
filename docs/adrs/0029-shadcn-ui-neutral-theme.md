# ADR 0029: shadcn/ui components and the neutral zinc web theme

Status: accepted (2026-11-07, owner decision). Ported from PR #69 (tag `archive/staging-2026-11-07`, where it was drafted as "ADR 0026"; that number is taken by the admin panel decision, so this is 0029). Amends D-22 ("No shadcn, no other component kit" in `.agents/rules/frontend.md`) for the public site. Related: ADR 0026 (admin panel on shadcn-style components), `apps/mobile/src/theme/neutral.json`.

## Context

The public web UI used Tailwind v4 plus Base UI primitives and hand-written classes: ledge shadows, ink-inversion hovers and a warm cream palette. The owner wants a neutral, Vercel-inspired look (flat, hairline borders, soft shadows, tight type), and the Android app was rebuilt to match it.

## Decision

- shadcn/ui components live in `resources/js/components/ui/` (`button`, `input`, `textarea`, `label`, `card`, `badge`, `separator`, `skeleton`); `cn()` is in `resources/js/lib/utils.ts`; `components.json` records the setup (new-york, neutral, CSS variables). `Btn`, `btnClass` and `Field` stay as the call sites pages use and delegate to them.
- The one new direct dependency is `@radix-ui/react-slot` 1.5.0 (exact). `class-variance-authority`, `clsx`, `tailwind-merge` and `lucide-react` were already present from ADR 0026.
- The palette overrides the `--hs-*` tokens in `resources/css/app.css` (neutral zinc: near-black primary, hairline borders, rounded-sm 4 / md 8 / lg 12). Category accents, score bands and star colours are unchanged. The old `shadow-ledge-*` names stay as a hairline plus a soft drop.
- Web and Android share the palette: `apps/mobile/src/theme/neutral.json` holds the same colours in hex and the same radii, and `packages/tokens/neutral.test.ts` fails when the two drift or when a text pair drops under WCAG AA (4.5:1, fields 3:1).
- Base UI stays for the primitives already built on it; Radix is only used through `Slot`.
- Pages redesigned since the PR (login, signup and 2FA column, footer, item tile) keep their structure; only class-level styling of the PR was applied to them.

## Consequences

- Two primitive libraries coexist (Base UI and Radix `Slot`); a later ADR can pick one.
- `packages/tokens` (the cream palette) is still what other consumers read; the web theme is the override. Moving the neutral palette into `packages/tokens` itself is a follow-up.
