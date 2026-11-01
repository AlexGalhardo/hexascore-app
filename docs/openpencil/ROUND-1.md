# Round 1: three visual directions

Status 2026-10-31: all three versions are delivered, 19 frames each (v1 Editorial Press, v2 Neon Noir, v3 Soft Paper).

Each version is one OpenPencil document with the pages `00 Cover`, `01 Foundations`, `02 Components`, `10 Web desktop 1440`, `11 Web mobile 390`, `20 Android 390x844` and `99 Assets` (real covers and shell components). PNGs are in each version's `frames/` folder, named `<platform>__<page-id>__<state>.png`.

## Side by side

| Frame | v1 Editorial Press | v2 Neon Noir | v3 Soft Paper |
| --- | --- | --- | --- |
| Cover | [cover](v1-editorial/frames/web__cover__default.png) | [cover](v2-noir/frames/web__cover__default.png) | [cover](v3-paper/frames/web__cover__default.png) |
| Foundations | [foundations](v1-editorial/frames/web__foundations__default.png) | [foundations](v2-noir/frames/web__foundations__default.png) | [foundations](v3-paper/frames/web__foundations__default.png) |
| Components | [components](v1-editorial/frames/web__components__default.png) | [components](v2-noir/frames/web__components__default.png) | [components](v3-paper/frames/web__components__default.png) |
| Home 1440 / 390 | [desktop](v1-editorial/frames/web-desktop__home__default.png) · [mobile](v1-editorial/frames/web-mobile__home__default.png) | [desktop](v2-noir/frames/web-desktop__home__default.png) · [mobile](v2-noir/frames/web-mobile__home__default.png) | [desktop](v3-paper/frames/web-desktop__home__default.png) · [mobile](v3-paper/frames/web-mobile__home__default.png) |
| `/games` 1440 / 390 | [desktop](v1-editorial/frames/web-desktop__catalog-index__default.png) · [mobile](v1-editorial/frames/web-mobile__catalog-index__default.png) | [desktop](v2-noir/frames/web-desktop__catalog-index__default.png) · [mobile](v2-noir/frames/web-mobile__catalog-index__default.png) | [desktop](v3-paper/frames/web-desktop__catalog-index__default.png) · [mobile](v3-paper/frames/web-mobile__catalog-index__default.png) |
| `/games/diablo-v` 1440 / 390 | [desktop](v1-editorial/frames/web-desktop__item__default.png) · [mobile](v1-editorial/frames/web-mobile__item__default.png) | [desktop](v2-noir/frames/web-desktop__item__default.png) · [mobile](v2-noir/frames/web-mobile__item__default.png) | [desktop](v3-paper/frames/web-desktop__item__default.png) · [mobile](v3-paper/frames/web-mobile__item__default.png) |
| Search 1440 / 390 | [desktop](v1-editorial/frames/web-desktop__search__default.png) · [mobile](v1-editorial/frames/web-mobile__search__default.png) | [desktop](v2-noir/frames/web-desktop__search__default.png) · [mobile](v2-noir/frames/web-mobile__search__default.png) | [desktop](v3-paper/frames/web-desktop__search__default.png) · [mobile](v3-paper/frames/web-mobile__search__default.png) |
| Library 1440 / 390 | [desktop](v1-editorial/frames/web-desktop__library__default.png) · [mobile](v1-editorial/frames/web-mobile__library__default.png) | [desktop](v2-noir/frames/web-desktop__library__default.png) · [mobile](v2-noir/frames/web-mobile__library__default.png) | [desktop](v3-paper/frames/web-desktop__library__default.png) · [mobile](v3-paper/frames/web-mobile__library__default.png) |
| Login 1440 / 390 | [desktop](v1-editorial/frames/web-desktop__login__default.png) · [mobile](v1-editorial/frames/web-mobile__login__default.png) | [desktop](v2-noir/frames/web-desktop__login__default.png) · [mobile](v2-noir/frames/web-mobile__login__default.png) | [desktop](v3-paper/frames/web-desktop__login__default.png) · [mobile](v3-paper/frames/web-mobile__login__default.png) |
| Pricing 1440 / 390 | [desktop](v1-editorial/frames/web-desktop__pricing__default.png) · [mobile](v1-editorial/frames/web-mobile__pricing__default.png) | [desktop](v2-noir/frames/web-desktop__pricing__default.png) · [mobile](v2-noir/frames/web-mobile__pricing__default.png) | [desktop](v3-paper/frames/web-desktop__pricing__default.png) · [mobile](v3-paper/frames/web-mobile__pricing__default.png) |
| Android home / item | [home](v1-editorial/frames/android__home__default.png) · [item](v1-editorial/frames/android__item__default.png) | [home](v2-noir/frames/android__home__default.png) · [item](v2-noir/frames/android__item__default.png) | [home](v3-paper/frames/android__home__default.png) · [item](v3-paper/frames/android__item__default.png) |
| Type | Playfair Display + Inter | Oxanium + Inter | Nunito |
| Default theme | light (paper) | dark (void) | light (cream) |
| Score mark | square band fill, ink number | ring in band colour with glow, number inside | pastel sticker chip, dark number |
| Density | highest (ranked tables, category columns) | medium (cover grids, panels) | lowest (tinted section panels, large cards) |

## v1 Editorial Press

Intent: a printed culture section, closest to the legacy "editorial data / almanac" principles. Strengths: the densest and most legible version; numbers lead (ranked lists, a `/games` table with all six category averages per row), hairline structure means almost no decoration to implement, and it maps cleanly to the current token set (light-first, 2 px radius, two font families). Risks: red is both primary and the movie accent (the primary button would need to switch to ink on movie pages); Playfair's old-style figures make the scores look uneven next to Inter (switch to lining figures or a different display face); it is the least "app-like" for the Android audience; covers are small, so weak cover art matters less but the page feels text-heavy.

## v2 Neon Noir

Intent: a dark, cinematic catalogue where covers and glowing score rings lead. Strengths: strongest visual identity and the most "product" feel on Android; cover-first grids make browsing fast; the score ring reads at any size and the band colour carries well on dark. Risks: dark-first is unusual for an SEO-heavy catalogue and needs a fully specified light twin (only tokens exist so far); gradients and glows add CSS weight and need care under the 30 kB CSS budget; placeholder covers (titles without art) look emptier in a cover grid than in a table; the uppercase condensed display face struggles with long pt-BR and es titles.

## v3 Soft Paper

Intent: approachable and warm, for casual raters as much as list-makers: cream page, white rounded cards, pastel section panels (lilac games, peach movies, sky series, mint books), sticker-like score chips and one family (Nunito, 17 px body). Strengths: the friendliest and most "app-like" of the three, closest to what the Android audience expects; the sticker chip and the tinted score panel on the item page make the band obvious without relying on colour (the number is always inside); one font family is the lightest load and the simplest to implement; pill buttons and 48 to 52 px controls meet the 44 px target everywhere. Risks: the lowest density (six cards per row on desktop, three-card horizontal scrollers on mobile), so `/games` shows a cover grid rather than the category columns of v1; pastel tints fail as text colours, so every tint needs its darker accent text token (see contrast below); large radii and tinted panels can drift into nested cards (the library side panel uses tinted rows inside a card, to fix in round 2 if v3 is chosen); Nunito 800 gets wide with long pt-BR and es titles, so cards must wrap to three lines.

## Contrast (WCAG 2.2, computed 2026-10-31)

Relative-luminance ratios computed from each README palette (4.5:1 for text, 3:1 for large text and UI).

| Pair | v1 Editorial | v2 Noir | v3 Paper |
| --- | --- | --- | --- |
| Light fg / bg | 17.3 | 17.2 | 13.4 |
| Light muted / bg | 6.9 | 6.1 | 4.93 (5.15 on surface) |
| Light primary button text | white on primary 5.34 | white on `#5B47E0` 6.1 | white on `#6B4CE6` 5.52 (old `#7A5CFA` 4.43, fixed) |
| Light primary text on bg | 4.99 | 5.67 | 5.28 |
| Light section accents on bg or tint | n/a (ink on fills) | movie 5.24, series 4.90, book 4.98 | on tint: game 5.02, movie 5.30 (old 4.08, fixed), series 5.22, book 5.25 |
| Light score marks | ink on fills: high 5.42, mid 9.12, low 5.43, none 6.96 | mid 5.48, none `#6B7385` 4.40 **fails** (4.76 on surface) | chips: high 5.25, mid 5.68, low 5.52, none 5.65 (old 4.32, fixed) |
| Dark fg / bg | 16.2 | 16.9 | 15.4 |
| Dark muted / bg | 8.3 | 7.69 (6.95 on surface) | 8.2 (7.2 on surface) |
| Dark primary | 6.4 | button text on gradient 5.95 to 6.15 | button text 6.8 |
| Dark accents | n/a | game 5.95, movie 6.48, series 9.69, book 10.3 | 6.3 to 6.7 |
| Dark score marks | n/a | high 9.31, mid 10.94, low 5.86, none `#6B7385` 4.09 on bg / 3.69 on surface **fails as text** | chips 5.9 to 6.8 |

Decorative hairlines (v2 border 1.39, v3 `line` 1.26) are not used as the only boundary of a control. v2 needs a lighter "none" colour (for example `#8A93A6`) before tokens are cut.

## OpenPencil gaps found in round 1

Recorded in `.claude/skills/openpencil-design/SKILL.md` ("Known gaps"): instance-bearing frames must be deleted before their main components; `open_file` on an open path merges stale nodes; text rendered while a Google font loads is stored with wrong glyphs; `set_font` needs the style name `Extra Bold`; `get_jsx` does not round-trip through `render`; the local MCP HTTP endpoint answers 503 after a few script sessions. Also: `render` takes one JSX expression (helpers go inside `{(() => { … })()}`, statements before the root do not run), and clipping is `overflow="hidden"` (`clip` is ignored).

## Notes for every version

- Real content from production (read-only, 2026-10-31): titles, years, counts, covers, plans and prices. Production has no ratings yet, so every score number is an illustrative sample on the real 0.5 to 10 scale; the "–" state is real. Library and the "rated by" names are illustrative.
- Covers: all 23 titles use their real catalogue cover (96x144 JPEG downscales of the production thumbnails), applied by `scripts/openpencil/set_image_fills.py`.
- Scores use the production scale (out of 10, bands below 5.0, 5.0 to 7.4, 7.5 and above), not the 0 to 100 scale in the legacy notes and the designer brief.
- Production `/search` returned HTTP 500 for `q=diablo` and `q=witcher` while collecting content; the search frames use plausible counts.

## Decisões para o owner (round 1)

- [ ] Versão: v1 Editorial, v2 Noir, v3 Paper, ou mistura (diga quais partes de cada uma).
- [ ] Paleta: manter a da versão escolhida? v1: trocar o vermelho primário (conflita com o acento de filmes)? v2: tema escuro como padrão ou claro como padrão?
- [ ] Tipografia: v1 com Playfair (algarismos old-style) ou outra serifada com algarismos alinhados; v2 Oxanium em caixa alta ou só nos rótulos.
- [ ] Densidade: tabela densa com as seis categorias (v1) ou grade de capas (v2) no `/games` e na home.
- [ ] Nota: anel (v2), bloco quadrado (v1) ou selo pastel (v3).
