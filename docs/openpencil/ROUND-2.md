# Round 2: v1 Editorial

The owner chose **v1 Editorial** (2026-10-31). v2 Noir and v3 Paper are frozen. Feedback: [OWNER-FEEDBACK-ROUND-2.md](OWNER-FEEDBACK-ROUND-2.md) (items 1 to 15). Document: `v1-editorial/hexascore-v1-editorial.fig`; exports in [v1-editorial/frames/](v1-editorial/frames/); every frame is rendered from the sources in [v1-editorial/src/](v1-editorial/src/).

## How the frames are built now

- One source per frame (`src/frames/<platform>_<page>_<state>.jsx`) over shared helpers: `v1-prelude.js` (tokens for both themes, text, score badge, bar, buttons, stars, pagination), `v1-shell.js` (navbar, footer, language menu, button sheet), `v1-item.js`, `v1-pages.js`, `v1-pages-r1.js` (the round-1 pages, rebuilt), `v1-new.js` (search suggestions, rating sheet).
- `src/build.sh [frames…]` writes `v1-all.js` (`DK = 0`) and `v1-all-dark.js` (`DK = 1`) and syntax-checks the frames with `bun build`. `src/gen_dark.py <page-id>` writes the dark copies (`web-desktop-dark/…`, `web-mobile-dark/…`) for page `12 Web dark`.
- Render: `python -I scripts/openpencil/op.py render <doc> src/v1-all.js src/frames/*.jsx` (and `v1-all-dark.js` with `src/frames-dark/*.jsx`). Export: `op.py export <doc> <out> <page> <names…>`.
- Container check (item 9): `op.py py <doc> scripts/openpencil/check_container.py` with `PAGES=0:7,<dark page>`.

## Before and after

| # | Item | Before (round 1) | After (round 2) |
| --- | --- | --- | --- |
| 1 | v3 ideas on the item page | kicker above the title, score line under it, no related titles | title, then meta line, then genre and platform chips (44 px pills), a score block ("High score · from 213 ratings · six categories, one formula", link to how the score works), actions; "More games like this" strip of five covers with year and score (four for books) |
| 2a | Score type | Playfair numbers | Inter 700 in fixed-width right-aligned boxes (`tabular-nums` in the build); Playfair only for headings |
| 2b | Category bars | value box could touch the bar | three fixed columns: label (120/104), track (240/160, hairline outline), value (40, right-aligned), 16 px gaps |
| 2c | Primary action | signal red `#C8321E` | brand purple `#5B3FC4` (dark `#A796FF`); red is now `danger` only |
| 3 | Dark theme | Foundations only | page `12 Web dark`: a dark copy of every web frame from the same sources (palette below) |
| 4 | Mobile footer | links in a wrapping row | links stacked and centred, each a 44 px row; copyright centred below |
| 5 | Navbar search | none | first drawn as a dropdown under the navbar field (8 rows: 2:3 cover, title and year, type, final score or "–"); **replaced by item 16**, its result rows reused inside the palette |
| 6 | Rating widget | slider per category | ten yellow stars per category, half stars (0.5 steps), value printed beside them ("8.5 / 10"), radio group of 20 per category, arrow keys ±0.5, Home/End, 44 px rows; `rating-widget/states` sheet (not rated, hover preview, keyboard focus, half, full, minimum, hit areas, keyboard table, colour) |
| 7 | Mobile categories | fixed 22 px gaps | `justify-content: space-between` across the full width |
| 8 | Theme toggle | half-circle icon | moon in light, sun in dark |
| 9 | Desktop container | some frames hand-placed | one container in the shell (`GX` = 144 px gutters, 1152 px content) used by navbar, footer and every page wrapper; checked by script on every 1440 frame, light and dark (list below) |
| 10 | Home | intro hero, then ranked lists | no hero: Games, Movies, TV Series, Books, each with a heading, a "see all" link, three large trending cards (rank, cover, title, year and type, score block) and a grid of six more titles with real covers |
| 11 | Category listing header | "Section · Games" eyebrow and a subtitle | the title alone, with the result count on the right |
| 12 | Buttons | flat; "Reset" was a text link | one `Btn` component (primary, secondary, ghost) with rest, hover, focus-visible, pressed and disabled, light and dark (`99 Assets` → `Buttons (round 2)`); 2 px ledge shadow at rest, pressed moves 1 px down with a 1 px ledge (reduced motion: no move, shadow only); "Reset" is a secondary button; every page button now comes from `Btn` |
| 13 | Listing pagination | "Load more" | numbered pagination above and below the results (previous, numbers with ellipses, next; current page filled), filters at the top; frames `catalog-index/default` (page 1), `page-622`, `last-page`; mobile shows previous / "Page N of M" / next |
| 14 | Language and navbar | EN PT ES links in the footer | language dropdown button in the navbar only (`header/desktop-language-open` shows the menu with the current language checked); footer has no language links; the navbar is fixed, `header/*-scrolled` adds a soft shadow under the rule |
| 15 | Google sign-in | none | "Sign in with Google" on login and "Sign up with Google" on the new signup frame, above the email form with an "or" divider; official G mark, Google light/dark button colours, pt-BR and es wording on the button sheet; same depth and states as item 12; noted as the one third-party mark in the v1 README |
| 16 | Search as a command palette | dropdown from item 5 | navbar trigger (icon, "Search…", "Ctrl K" hint; icon only on mobile) opens a centred modal over a dimmed backdrop (full-screen sheet on mobile) with the input focused; the item 5 rows, first 8; ↑/↓ highlight, Enter opens, Esc closes; focus trapped and returned to the trigger; frames `search-palette/empty` (recent searches and trending), `results` (with "See all results" to `/search`), `loading`, `no-results`, desktop and mobile, light and dark; the `header/*-search` components are gone |

## Dark palette and contrast (WCAG 2.2, computed)

| Pair | Light | Dark |
| --- | --- | --- |
| text on page | 17.31 | 16.16 |
| muted on page / surface | 6.88 / 6.24 | 8.32 / 7.56 |
| primary text on page | 6.65 | 7.52 |
| text on primary button | 7.11 (`#FFFFFF` on `#5B3FC4`) | 7.52 (`#14120F` on `#A796FF`) |
| field outline on field / page (UI, 3:1) | 3.80 / 3.55 | 3.73 / 4.10 |
| accents on page: games, movies, series, books | 6.65, 4.99, 5.30, 4.72 | 7.52, 6.44, 8.29, 8.50 |
| star outline on page (UI) | 5.03 (`#8A6400`) | 11.16 |
| ink number on score fills: high, mid, low, none | 5.42, 9.12, 5.43, 6.96 | same fills, same ink |

Light `books` accent on surface is 4.28: the accent is used on the page background only. Score fills keep ink numbers in both themes; in dark the fills sit on `#14120F` at 5.47 (high) or more.

## Frames

Every frame exists in light (`10 Web desktop 1440`, `11 Web mobile 390`) and dark (`12 Web dark`, platform `web-desktop-dark` / `web-mobile-dark`); PNGs `<platform>__<page>__<state>.png`.

- `catalog-index`: `default`, `empty`, `filters-open` (mobile only), `last-page`, `page-622`
- `docs`: `default`
- `edition`: `default`
- `entity`: `default`, `empty`
- `error`: `404`, `500`
- `home`: `default`, `recent`
- `item`: `default`, `logged-in`, `logged-out`, `no-cover`, `no-score`, `no-synopsis`
- `legal`: `default`
- `library`: `default`
- `login`: `default`
- `pricing`: `default`
- `rating-widget`: `states`
- `recent-changes`: `default`
- `search-palette`: `empty`, `loading`, `no-results`, `results`
- `search`: `default`, `empty-query`, `no-results`, `too-short`
- `signup`: `default`

Home frames changed by item 10: `web-desktop/home/default`, `web-desktop/home/recent`, `web-mobile/home/default`, `web-mobile/home/recent` and their `-dark` copies.

Item 9, frames corrected: the round-1 frames were hand-drawn copies with their own header and margins and are now rebuilt on the shared container: `home/default`, `home/recent`, `catalog-index/default`, `search/default`, `library/default`, `login/default`, `pricing/default`, `item/default` (desktop). `rating-widget/states` overflowed the right edge by 128 px during the round and was fixed. Every other 1440 frame already used the page wrapper; the check script now passes on all 33 light and 33 dark desktop frames.

`99 Assets`: `Shell components (round 2)` and `Shell components dark (round 2)` (navbar rest, scrolled, signed in, language menu open; footer; mobile navbar rest, scrolled, signed in; mobile footer), `Buttons (round 2)` and `Buttons dark (round 2)` (primary, secondary, ghost × rest, hover, focus-visible, pressed, disabled; Google button states and locales).

## Tool gaps found this round

Recorded in `.claude/skills/openpencil-design/SKILL.md` (known gaps): no tabular figures, CSS `shadow` string only, misleading render error positions (use `build.sh`), RPC timeouts on long frames, component mains deleted with their frame.

## Open points

- Trending (last 7 days) needs a rollup that production does not have; scores, counts and trending order on these frames are illustrative (production has no ratings yet).
- The games accent equals the new primary purple; accepted unless the owner wants a distinct games hue.
- Google button: Inter instead of Roboto Medium and 44 px with a ledge shadow; confirm against Google's guidelines at implementation.
- Android frames (`20 Android 390x844`) were not part of this round and still show round-1 colours.
