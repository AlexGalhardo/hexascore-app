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
| `primary` | `#C8321E` signal red | `#F0715E` | links, primary button |
| `secondary` | `#1F4E5F` deep teal | `#7FB3C4` | secondary actions |
| `accent-game` / `movie` / `series` / `book` | `#5B3FC4` / `#C8321E` / `#1F6F8B` / `#2E7D4F` | `#A796FF` / `#F0715E` / `#5FB8D6` / `#6CC08A` | one accent per section |
| `score-high` / `mid` / `low` / `none` | `#2F9F3D` / `#E6AD00` / `#F05653` / `#9B9FA5` | same | badge fills, ink number on top |

## Type

- Display and scores: Playfair Display (OFL), 700, tight.
- UI and body: Inter (OFL), tabular figures for counts and scores in tables.
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

## Open questions

- Dark mode is specified in Foundations only; the dark frames (`12 Web dark`) come in round 2 for the chosen version.
- Red is both the primary colour and the movie accent; on movie pages the primary button would switch to ink. Keep or pick another primary?
