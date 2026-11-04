# ADR 0027: shadcn/ui components and a Vercel-like web theme

Status: accepted for an experiment on a branch (2026-11-03). Owner request in chat; it extends ADR 0026 (admin panel only) to the public site.

## Context

The web UI used Tailwind v4 plus Base UI behaviour primitives and hand-written classes: ledge shadows, ink-inversion hovers and a warm cream palette. The owner wants to see whether shadcn/ui and a neutral, Vercel-inspired look (flat, hairline borders, soft shadows, tight type) feels better, tested on the staging environment before any decision for production.

## Decision

- Add shadcn/ui's dependencies at exact versions: `class-variance-authority` 0.7.1, `clsx` 2.1.1, `tailwind-merge` 3.7.0, `lucide-react` 1.67.0 and `@radix-ui/react-slot` 1.5.0. `components.json` records the shadcn setup (new-york, neutral, CSS variables).
- shadcn components live in `resources/js/components/ui/` and are ours to edit (`button`, `input`, `textarea`, `label`, `card`, `badge`, `separator`, `skeleton`). `cn()` is in `resources/js/lib/utils.ts`. More components are added one at a time as pages need them.
- `Btn`, `btnClass` and `Field` stay as the call sites pages use and now delegate to the shadcn button and input, so every page changes without a rewrite.
- The theme overrides the `--hs-*` tokens in `resources/css/app.css`, not in `packages/tokens`, so the Android app keeps its palette. Category accents, score bands and star colours are unchanged. The old `shadow-ledge-*` names remain as a hairline plus a soft drop.
- Base UI stays for the primitives already built on it (nav sheet, search palette dialogs). Radix is only used through `Slot`.

## Consequences

- Reverting is one revert of the experiment branch: no data, API or route changes.
- Two primitive libraries coexist (Base UI and Radix `Slot`); a later ADR can pick one.
- Contrast and focus visibility must be re-checked against the new greys (WCAG AA) before this reaches production.
