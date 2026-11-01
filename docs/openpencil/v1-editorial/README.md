# v1 Editorial Press

Round 1 concept. Document: `hexascore-v1-editorial.fig` (OpenPencil). PNG exports: [frames/](frames/).

## Intent

A printed culture section: an almanac you read, not a dashboard you operate. Numbers are the protagonists (large serif scores, tabular counts), hierarchy comes from type and hairline rules rather than cards and shadows. Closest to the legacy "editorial data" principles.

## Palette

| Token | Light | Dark | Use |
| --- | --- | --- | --- |
| `bg` | `#FAF7F0` paper | `#14120F` | page |
| `surface` | `#F1ECE0` | `#1F1C17` | panels, inputs |
| `fg` | `#16130F` ink | `#F3EEE3` | text |
| `muted` | `#5C554A` | `#B5AC9C` | secondary text |
| `rule` | `#D9D2C3` | `#3A352C` | hairlines |
| `primary` | `#5B3FC4` brand purple (round 2) | `#A796FF` | links, primary button, focus ring; text on it `#FFFFFF` / `#14120F` |
| `primary-hover` / `ledge-primary` / `ledge-secondary` | `#4A30AE` / `#2E1F6B` / `#16130F` | `#B9ACFF` / `#5E4DB8` / `#000000` | button hover fill and the 2 px depth shadow |
| `danger` | `#C8321E` | `#F0715E` | errors only (red no longer means "act") |
| `field` / `field-bg` | `#8A8273` / `#FFFFFF` | `#7D7566` / `#1F1C17` | input outlines (3:1) and fills |
| `star` / `star-line` | `#E6AD00` / `#8A6400` | `#F2C230` / `#F2C230` | rating stars |
| `secondary` | `#1F4E5F` deep teal | `#7FB3C4` | secondary actions |
| `accent-game` / `movie` / `series` / `book` | `#5B3FC4` / `#C8321E` / `#1F6F8B` / `#2E7D4F` | `#A796FF` / `#F0715E` / `#5FB8D6` / `#6CC08A` | one accent per section |
| `score-high` / `mid` / `low` / `none` | `#2F9F3D` / `#E6AD00` / `#F05653` / `#9B9FA5` | same | badge fills, ink number on top |

## Type

- Display (headings only): Playfair Display (OFL), 700, tight.
- UI, body and every score: Inter (OFL) with `font-variant-numeric: tabular-nums` (round 2; OpenPencil cannot draw tabular figures, so frames put numbers in fixed-width right-aligned boxes).
- Two families only (budget: at most two self-hosted WOFF2 families).

## Frames (round 1)

Pages `00 Cover`, `01 Foundations`, `02 Components`, `10 Web desktop 1440`, `11 Web mobile 390`, `20 Android 390x844`. Exports:

- `web__cover__default.png`, `web__foundations__default.png`, `web__components__default.png`
- `web-desktop__<page>__default.png` and `web-mobile__<page>__default.png` for `home`, `catalog-index`, `item`, `search`, `library`, `login`, `pricing`
- `android__home__default.png`, `android__item__default.png`

## Frames (DESIGN-v1-editorial-web-catalog)

Public catalogue group, desktop 1440 (`10 Web desktop 1440`) and mobile 390 (`11 Web mobile 390`), 37 new frames plus the reworked home default:

- `item`: `logged-out` (World of Warcraft, every block: score, radar, synopsis with licence, details, time to beat, trailers behind consent), `logged-in` (six ScoreInputs with three set, overall score, status picker, add to list, your time to beat), `no-score` (The Witcher IV, "Not enough ratings" on badge and panel, empty radar), `no-cover` (Iliad, book categories, cover placeholder), `no-synopsis` (Diablo V, as in production). `item/default` from round 1 stays the logged-out Diablo V reference.
- `home`: `default` now matches production (no "Recently rated" block); `recent` shows the block.
- `edition/default`, `entity/default` and `entity/empty`, `search/empty-query`, `search/too-short`, `search/no-results` (with "did you mean"), `catalog-index/empty` (desktop and mobile) and `catalog-index/filters-open` (mobile), `recent-changes/default`, `docs/default` (Authentication page), `legal/default` (terms), `error/404` and `error/500`.

Sources: `src/v1-prelude.js` (tokens and primitives), `src/v1-item.js` (item page), `src/v1-pages.js` (other pages) and one `src/frames/*.jsx` per frame. Render with `python scripts/openpencil/op.py render <doc> <prelude> <files…>` after concatenating the three `.js` files; a frame is replaced in place when a top-level frame with the same name exists.

Content notes: Recent changes is specified (SDD 09 §6.2, SDD 04 FR-04-13) but not built, so its entries are illustrative; the Neuromancer edition is a sample (production has no editions yet); the "did you mean" line has no catalogue key yet; the empty entity uses the Ouya platform as a sample.

## Content

Titles, years, counts (24873 games), cover images, plans and prices come from production (read-only, 2026-10-31). Production has no ratings yet (every score is `null`), so the score numbers in these frames are illustrative samples on the real 0.5 to 10 scale; the "–" state is shown where the real data has no score. Library content is illustrative (logged-in page; production was not logged in to).

## Round 2

The owner chose this version. Changes, sources and the before/after list are in [ROUND-2.md](../ROUND-2.md). New pages in the document: `12 Web dark` (a dark copy of every web frame); `99 Assets` gained `Shell components (round 2)` (light and dark), `Buttons (round 2)` (light and dark) and 17 more covers.

Third-party mark exception: the Google "G" mark on the "Sign in with Google" / "Sign up with Google" buttons (login and signup) is the one allowed third-party logo (owner, round 2, item 15). It follows Google's sign-in branding guidelines (official four-colour G, light button `#FFFFFF` / `#747775` / `#1F1F1F`, dark button `#131314` / `#8E918F` / `#E3E3E3`, wording translated: pt-BR "Fazer login com o Google" / "Inscrever-se com o Google", es "Iniciar sesión con Google" / "Registrarse con Google"). Deviations to confirm at implementation: Inter instead of Roboto Medium (two-family budget) and a 44 px height with the v1 depth ledge.

## Open questions

- Resolved in round 2: primary is purple. It now equals the games accent (`#5B3FC4`); the accent appears only as small section text, so the overlap is accepted unless the owner wants a distinct games hue.
