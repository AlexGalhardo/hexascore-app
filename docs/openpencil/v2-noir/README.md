# v2 Neon Noir

Round 1 concept. Document: `hexascore-v2-noir.fig` (OpenPencil). PNG exports: [frames/](frames/).

## Intent

A dark, cinematic catalogue for people who live in launchers and streaming apps: covers lead, scores glow in their band colour, sections are told apart by a neon accent. Dark is the primary theme; light mode is a cool grey variant of the same tokens.

## Palette

| Token | Dark (primary) | Light | Use |
| --- | --- | --- | --- |
| `bg` | `#0B0D12` void | `#F4F6FA` | page |
| `surface` | `#151923` | `#FFFFFF` | panels, cards |
| `raised` | `#1C2230` | `#EBEEF4` | inputs, hover |
| `border` | `#262C3B` | `#D5DAE3` | 1 px inner borders |
| `fg` | `#EDEFF5` | `#0F131B` | text |
| `muted` | `#9AA3B8` | `#545D70` | secondary text |
| `primary` | gradient `#8B7CFF` to `#5B8CFF`, text `#0B0D12` | `#5B47E0` | primary button, focus ring |
| `accent-game` / `movie` / `series` / `book` | `#8B7CFF` / `#FF5C6C` / `#22C7EE` / `#5BD38A` | `#5B47E0` / `#C81E3A` / `#00749A` / `#177A40` | one accent per section |
| `score-high` / `mid` / `low` / `none` | `#5BD38A` / `#FFC24B` / `#FF5C6C` / `#6B7385` | `#177A40` / `#8A5A00` / `#C81E3A` / `#6B7385` | ring and number colour (text-safe on `bg`) |

## Type

- Display, section labels and big scores: Oxanium (OFL), 600 to 700, uppercase labels with wide tracking.
- UI and body: Inter (OFL), tabular figures.

## Frames (round 1)

Same set as the other versions: `00 Cover`, `01 Foundations`, `02 Components`, desktop 1440 and mobile 390 for `home`, `catalog-index`, `item`, `search`, `library`, `login`, `pricing`, plus Android `home` and `item` at 390x844. Exports follow `<platform>__<page-id>__<state>.png`.

## Content

Real titles, years, counts, covers, plans and prices from production (read-only, 2026-10-31). Production has no ratings yet, so score numbers are illustrative samples on the real 0.5 to 10 scale. Library content is illustrative.

## Open questions

- Glow and gradients cost nothing in CSS, but the scanline/noise texture from the concept prompt was left out (image or SVG background, CSP and LCP risk). Bring it back?
- Light mode is secondary here; is a dark-first default acceptable for an SEO-heavy catalogue?
