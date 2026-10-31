# v3 Soft Paper

Round 1 concept. Document: `hexascore-v3-paper.fig` (OpenPencil). PNG exports: [frames/](frames/).

## Intent

Approachable and warm: a product for casual raters as much as for list-makers. Rounded cards, pastel section tints, sticker-like score chips, one friendly rounded family, generous spacing and a 17 px body.

## Palette

| Token | Light | Dark | Use |
| --- | --- | --- | --- |
| `bg` | `#FFF9F2` cream | `#1B1A24` | page |
| `surface` | `#FFFFFF` card | `#252433` | cards |
| `fg` | `#2B2A3A` ink | `#F4F1FA` | text |
| `muted` | `#6E6B80` | `#B4B0C6` | secondary text |
| `primary` | `#6B4CE6` violet (5.5:1 with white; `#7A5CFA` failed at 4.4:1) | `#A894FF` | buttons, links |
| `danger` | `#E5484D` coral (text `#C2282D`) | `#FF8A8E` | errors, destructive |
| tints game / movie / series / book | `#DDD3FF` lilac / `#FFD9C2` peach / `#CFE8FF` sky / `#CDEFE0` mint | 20 % of the dark accents | section backdrops |
| accent text game / movie / series / book | `#5B3FC4` / `#9A3B12` / `#1F5F9E` / `#1D6B45` | `#C3B6FF` / `#FFB48F` / `#93C9FF` / `#8ED9B5` | text on tints |
| `score-high` / `mid` / `low` / `none` | chip `#CDEFE0` + `#1D6B45` / `#FFE7A8` + `#7A5200` / `#FFD6D6` + `#A3262A` / `#ECEAF2` + `#5C5970` | same hues at 25 % on dark | sticker chips, always with the number |

## Type

- One family: Nunito (OFL), 400/700/800. Rounded 800 numerals for scores. Body 17 px, line height 1.55.

## Frames (round 1)

Same set as the other versions: `00 Cover`, `01 Foundations`, `02 Components`, desktop 1440 and mobile 390 for `home`, `catalog-index`, `item`, `search`, `library`, `login`, `pricing`, plus Android `home` and `item` at 390x844. Exports follow `<platform>__<page-id>__<state>.png`.

## Content

Real titles, years, counts, covers, plans and prices from production (read-only, 2026-10-31). Production has no ratings yet, so score numbers are illustrative samples on the real 0.5 to 10 scale. Library content is illustrative.

## Open questions

- Lowest information density of the three: the home shows fewer rows per section above the fold. Acceptable for a data product?
- Pastel tints fail as text colours; the accent text colours above are the text-safe pair. Confirm before tokens are cut.
