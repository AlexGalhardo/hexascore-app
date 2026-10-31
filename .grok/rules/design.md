---
paths:
  - "docs/openpencil/**"
  - "docs/design-system.md"
  - "docs/design-handoff.md"
  - "packages/tokens/**"
  - "resources/js/**"
  - "resources/css/**"
  - "apps/mobile/app/**"
  - "apps/mobile/src/**"
---

# Design and UI rules

Which skill, when:

| Work | Load |
| --- | --- |
| Visual direction, typography, avoiding template looks | `frontend-design` |
| Critique, audit, polish, harden, typeset, layout, colorize (one command per pass) | `impeccable` (vendored 4.6.0; upstream is newer, update only with the owner's yes) |
| Review of UI code against interface guidelines | `web-design-guidelines` |
| Drawing in OpenPencil (frames, components, exports) | `openpencil-design` |
| Android screens | `expo-design-system`, `expo-native-ui` |

Project constraints that override any skill default:

- Tokens first: every colour, size, radius and font comes from `packages/tokens/` (web via `@theme`, mobile via NativeWind). A design value that is not a token is a token proposal in `docs/design-handoff.md`, not a one-off.
- CSP: no inline `style` attributes, no runtime CSS injection; anything a design needs must be expressible as utility classes.
- WCAG 2.2 AA in light and dark (4.5:1 text, 3:1 large text and UI), visible focus on every control, 44 px targets, colour never the only signal, a reduced-motion alternative for any motion.
- Scores: out of 10 with one decimal; bands below 5.0 low, 5.0 to 7.4 mid, 7.5 and above high, "–" with fewer than 5 ratings; the number is always printed next to the colour.
- Three locales (en, pt-BR, es): design with the longest string, no text in images, labels and menus wrap.
- Performance: at most two self-hosted WOFF2 families (OFL), `font-display: swap`; reserved boxes for covers (2:3) and filters (CLS ≤ 0.05, LCP ≤ 2.0 s).
- Real content only (titles, counts, covers from the catalogue), no lorem ipsum, no third-party logos.
- Impeccable's anti-pattern list applies (no nested cards, no eyebrow chips stacked on heroes, no side-stripe tabs, no glassmorphism by default); a version may break one deliberately only when its README says why.
- Implementation of a redesign starts only after the owner writes "aprovado" for a version and frame set (brief `.orchestration/briefs/designer.md`).
